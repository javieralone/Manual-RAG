#!/bin/sh
set -eu

certificate_dir=/etc/letsencrypt/live/manual-rag
certificate_file="$certificate_dir/fullchain.pem"
private_key_file="$certificate_dir/privkey.pem"

if [ ! -f "$certificate_file" ] || [ ! -f "$private_key_file" ]; then
    mkdir -p "$certificate_dir"
    openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
        -keyout "$private_key_file" \
        -out "$certificate_file" \
        -subj "/CN=localhost"
fi

(
    while true; do
        sleep 300
        nginx -s reload || true
    done
) &

exec nginx -g 'daemon off;'