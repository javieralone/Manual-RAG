import { useEffect, useMemo, useRef, useState } from 'react';
import { login as loginRequest } from './api/auth';
import { askNormalQuery, askStreamQuery } from './api/chat';

const DEFAULT_COLLECTION = 'manuales_tecnicos';

function App() {
  const [token, setToken] = useState(localStorage.getItem('manual-rag-token') || '');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(Boolean(token));
  const [selectedCollection, setSelectedCollection] = useState(DEFAULT_COLLECTION);
  const [queryMode, setQueryMode] = useState('normal');
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [contextItems, setContextItems] = useState([]);
  const streamCursorRef = useRef('');
  const bottomRef = useRef(null);

  const collectionOptions = useMemo(
    () => ['generic_manuals', 'manuales_tecnicos'],
    []
  );

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, loading, contextItems]);

  const persistToken = (nextToken) => {
    if (nextToken) {
      localStorage.setItem('manual-rag-token', nextToken);
    } else {
      localStorage.removeItem('manual-rag-token');
    }
  };

  const appendStreamChunk = (rawText) => {
    const currentText = streamCursorRef.current;
    if (!rawText) return;

    let delta = rawText;
    const maxOverlap = Math.min(currentText.length, rawText.length);

    for (let i = maxOverlap; i > 0; i--) {
      const suffix = currentText.slice(currentText.length - i);
      const prefix = rawText.slice(0, i);
      if (suffix === prefix) {
        delta = rawText.slice(i);
        break;
      }
    }

    if (!delta) return;
    streamCursorRef.current = currentText + delta;

    setMessages((prev) => {
      const next = [...prev];
      const last = next[next.length - 1];
      if (last && last.role === 'assistant') {
        last.text = (last.text || '') + delta;
      }
      return [...next];
    });
  };

  const handleLogin = async (event) => {
    event.preventDefault();
    setError('');
    setLoading(true);

    try {
      const result = await loginRequest(username, password);
      const nextToken = result.access_token || result.token || '';
      persistToken(nextToken);
      setToken(nextToken);
      setIsAuthenticated(Boolean(nextToken));
    } catch (err) {
      setError(err.message || 'Credenciales inválidas');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    persistToken('');
    setToken('');
    setIsAuthenticated(false);
    setMessages([]);
    setContextItems([]);
    streamCursorRef.current = '';
    setUsername('');
    setPassword('');
  };

  const handleSend = async () => {
    if (!input.trim()) return;

    const question = input.trim();
    setInput('');
    setError('');
    setLoading(true);
    setContextItems([]);
    streamCursorRef.current = '';

    const userMessage = { role: 'user', text: question };
    setMessages((prev) => [...prev, userMessage]);

    try {
      if (queryMode === 'stream') {
        const assistantMessage = { role: 'assistant', text: '' };
        setMessages((prev) => [...prev, assistantMessage]);

        await askStreamQuery({
          question,
          collection: selectedCollection,
          token,
          onToken: (text) => {
            appendStreamChunk(text || '');
          },
          onMetadata: (metadata) => {
            setContextItems(metadata?.context || []);
          },
          onComplete: () => {
            streamCursorRef.current = '';
          },
          onError: (message) => {
            setError(message);
          },
        });
      } else {
        const result = await askNormalQuery({
          question,
          collection: selectedCollection,
          token,
        });

        const assistantMessage = {
          role: 'assistant',
          text: result?.answer || 'Sin respuesta',
          context: result?.context || [],
        };

        setMessages((prev) => [...prev, assistantMessage]);
        setContextItems(result?.context || []);
      }
    } catch (err) {
      setError(err.message || 'Error al consultar el sistema');
    } finally {
      setLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="page-shell auth-shell">
        <form className="auth-card" onSubmit={handleLogin}>
          <div className="brand-lockup">
            <div className="brand-mark">M</div>
            <div>
              <p className="label">CONSULTA</p>
              <h1>Manual-RAG</h1>
            </div>
          </div>

          <label>
            Usuario
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
              autoComplete="username"
            />
          </label>

          <label>
            Contraseña
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              autoComplete="current-password"
            />
          </label>

          <button type="submit" disabled={loading}>
            {loading ? 'Conectando...' : 'Iniciar sesión'}
          </button>

          {error && <div className="error-box">{error}</div>}
        </form>
      </div>
    );
  }

  return (
    <div className="page-shell app-shell">
      <header className="topbar">
        <div className="brand-lockup compact">
          <div className="brand-mark small">M</div>
          <div>
            <p className="label">CONSULTA</p>
            <h2>Manual-RAG</h2>
          </div>
        </div>

        <div className="controls">
          <label>
            Colección
            <select
              value={selectedCollection}
              onChange={(e) => setSelectedCollection(e.target.value)}
            >
              {collectionOptions.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
            </select>
          </label>

          <label>
            Modo
            <select value={queryMode} onChange={(e) => setQueryMode(e.target.value)}>
              <option value="normal">Sin stream</option>
              <option value="stream">Con stream</option>
            </select>
          </label>

          <button className="ghost" onClick={handleLogout} type="button">
            Salir
          </button>
        </div>
      </header>

      <main className="chat-layout">
        <section className="chat-panel">
          <div className="messages">
            {messages.length === 0 && (
              <div className="empty-state">
                Haz tu primera pregunta sobre los manuales técnicos.
              </div>
            )}

            {messages.map((message, index) => (
              <div key={`${message.role}-${index}`} className={`message ${message.role}`}>
                <div className="bubble">
                  {message.text || '...'}
                </div>
              </div>
            ))}

            {loading && queryMode === 'stream' && (
              <div className="message assistant">
                <div className="bubble loading">Generando respuesta...</div>
              </div>
            )}

            <div ref={bottomRef} />
          </div>

          <div className="composer">
            <textarea
              rows="3"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Escribe tu pregunta..."
            />
            <button onClick={handleSend} disabled={loading || !input.trim()}>
              {loading ? 'Enviando...' : 'Enviar'}
            </button>
          </div>
        </section>

        <aside className="context-panel">
          <h3>Contexto recuperado</h3>
          <div className="context-list">
            {contextItems.length === 0 ? (
              <p className="no-context">No hay contexto aún.</p>
            ) : (
              contextItems.map((item, index) => (
                <div key={`${item.text}-${index}`} className="context-item">
                  <small>Score: {item.score?.toFixed?.(3) ?? 'n/a'}</small>
                  <p>{item.text}</p>
                </div>
              ))
            )}
          </div>

          {error && <div className="error-box compact">{error}</div>}
        </aside>
      </main>
    </div>
  );
}

export default App;
