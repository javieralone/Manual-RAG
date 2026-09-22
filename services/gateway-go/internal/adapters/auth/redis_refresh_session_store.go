package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"api-go/internal/core/domain"
)

const refreshSessionKeyPrefix = "auth:refresh-session:"

var rotateRefreshSessionScript = redis.NewScript(`
local current = redis.call('GET', KEYS[1])
if not current then return 0 end
local decoded = cjson.decode(current)
if decoded.username ~= ARGV[1] then return 0 end
local created = redis.call('SET', KEYS[2], ARGV[2], 'PX', ARGV[3], 'NX')
if not created then return 0 end
redis.call('DEL', KEYS[1])
return 1
`)

type redisSessionValue struct {
	Username string `json:"username"`
}

type RedisRefreshSessionStore struct {
	client redis.UniversalClient
	now    func() time.Time
}

func NewRedisRefreshSessionStore(client redis.UniversalClient) *RedisRefreshSessionStore {
	return &RedisRefreshSessionStore{client: client, now: time.Now}
}

func (s *RedisRefreshSessionStore) Create(ctx context.Context, session domain.RefreshSession) error {
	value, ttl, err := sessionValue(session, s.now())
	if err != nil {
		return err
	}
	created, err := s.client.SetNX(ctx, refreshSessionKey(session.ID), value, ttl).Result()
	if err != nil {
		return err
	}
	if !created {
		return errors.New("refresh session already exists")
	}
	return nil
}

func (s *RedisRefreshSessionStore) Rotate(ctx context.Context, current domain.RefreshSession, replacement domain.RefreshSession) (bool, error) {
	value, ttl, err := sessionValue(replacement, s.now())
	if err != nil {
		return false, err
	}
	result, err := rotateRefreshSessionScript.Run(ctx, s.client, []string{refreshSessionKey(current.ID), refreshSessionKey(replacement.ID)}, current.Username, value, ttl.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (s *RedisRefreshSessionStore) Revoke(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, refreshSessionKey(sessionID)).Err()
}

func (s *RedisRefreshSessionStore) Check(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisRefreshSessionStore) Close() error {
	return s.client.Close()
}

func refreshSessionKey(sessionID string) string {
	return refreshSessionKeyPrefix + sessionID
}

func sessionValue(session domain.RefreshSession, now time.Time) (string, time.Duration, error) {
	ttl := session.ExpiresAt.Sub(now)
	if session.ID == "" || session.Username == "" || ttl <= 0 {
		return "", 0, errors.New("invalid refresh session")
	}
	value, err := json.Marshal(redisSessionValue{Username: session.Username})
	return string(value), ttl, err
}