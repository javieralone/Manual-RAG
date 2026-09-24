import { useEffect, useMemo, useRef, useState } from 'react';
import { login as loginRequest, logout as logoutRequest, refreshToken as refreshTokenRequest } from './api/auth';
import { askNormalQuery, askStreamQuery } from './api/chat';
import { getStorageOptions, uploadDocument } from './api/ingestion';

const DEFAULT_COLLECTION = 'manuales_tecnicos';
const PDF_NAME_PATTERN = /^.+__parte-\d{3}\.pdf$/i;

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
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [storageOptions, setStorageOptions] = useState([]);
  const [uploadFile, setUploadFile] = useState(null);
  const [uploadBucket, setUploadBucket] = useState('manuals');
  const [uploadObjectKey, setUploadObjectKey] = useState('');
  const [uploadError, setUploadError] = useState('');
  const [uploadMessage, setUploadMessage] = useState('');
  const [uploading, setUploading] = useState(false);
  const streamCursorRef = useRef('');
  const activeRequestRef = useRef(null);
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
        next[next.length - 1] = {
          ...last,
          text: (last.text || '') + delta,
        };
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

  const handleLogout = async () => {
    try {
      await logoutRequest();
    } catch {
      // The local client state must still be cleared after a network failure.
    }
    persistToken('');
    setToken('');
    setIsAuthenticated(false);
    setMessages([]);
    setContextItems([]);
    streamCursorRef.current = '';
    setUsername('');
    setPassword('');
  };

  const handleRefreshToken = async () => {
    setError('');
    setLoading(true);

    try {
      const result = await refreshTokenRequest();
      const nextToken = result.access_token || result.token || '';

      if (!nextToken) {
        throw new Error('El servidor no devolvió un token de acceso.');
      }

      persistToken(nextToken);
      setToken(nextToken);
    } catch (err) {
      setError(err.message || 'No se pudo renovar la sesión');
    } finally {
      setLoading(false);
    }
  };

  const openUploadModal = async () => {
    setUploadError('');
    setUploadMessage('');
    setIsUploadOpen(true);
    try {
      const result = await getStorageOptions(token);
      setStorageOptions(result.buckets || []);
    } catch (err) {
      setUploadError(err.message || 'No se pudo cargar el almacenamiento.');
    }
  };

  const handleUploadFileChange = (event) => {
    const file = event.target.files?.[0] || null;
    setUploadFile(file);
    setUploadError('');
    setUploadMessage('');
    if (file && !PDF_NAME_PATTERN.test(file.name)) {
      setUploadError('El archivo debe llamarse <nombre_manual>__parte-xxx.pdf.');
      return;
    }
    if (file && !uploadObjectKey) {
      setUploadObjectKey(`generic_manuals/${file.name}`);
    }
  };

  const handleBucketChange = (event) => {
    setUploadBucket(event.target.value);
    setUploadObjectKey('');
  };

  const handleUpload = async (event) => {
    event.preventDefault();
    if (!uploadFile || !uploadBucket || !uploadObjectKey.trim()) return;
    if (!PDF_NAME_PATTERN.test(uploadFile.name)) {
      setUploadError('El archivo debe llamarse <nombre_manual>__parte-xxx.pdf.');
      return;
    }
    if (uploadObjectKey.split('/').pop() !== uploadFile.name) {
      setUploadError('El object_key debe terminar con el nombre original del PDF.');
      return;
    }

    setUploading(true);
    setUploadError('');
    setUploadMessage('Subiendo y encolando documento...');
    try {
      const result = await uploadDocument({
        token,
        file: uploadFile,
        bucket: uploadBucket,
        objectKey: uploadObjectKey.trim(),
      });
      setUploadMessage(`Job ${result.job_id} creado: ${result.status}.`);
      const refreshed = await getStorageOptions(token);
      setStorageOptions(refreshed.buckets || []);
      setUploadFile(null);
      event.target.reset();
    } catch (err) {
      setUploadError(err.message || 'No se pudo subir el documento.');
      setUploadMessage('');
    } finally {
      setUploading(false);
    }
  };

  const handleSend = async () => {
    if (!input.trim() || activeRequestRef.current) return;

    const question = input.trim();
    const controller = new AbortController();
    activeRequestRef.current = controller;
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
          signal: controller.signal,
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
          signal: controller.signal,
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
      if (err.name === 'AbortError') {
        setError('Consulta cancelada.');
        return;
      }
      setError(err.message || 'Error al consultar el sistema');
    } finally {
      if (activeRequestRef.current === controller) {
        activeRequestRef.current = null;
        setLoading(false);
      }
    }
  };

  const handleCancel = () => {
    activeRequestRef.current?.abort();
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
          <button
            className="ghost"
            onClick={handleRefreshToken}
            type="button"
            disabled={loading}
          >
            Renovar sesión
          </button>
          <button className="upload-trigger" onClick={openUploadModal} type="button">
            Subir PDF
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

            {loading && (
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
              disabled={loading}
            />
            {loading ? (
              <button className="ghost" onClick={handleCancel} type="button">
                Cancelar
              </button>
            ) : (
              <button onClick={handleSend} disabled={!input.trim()}>
                Enviar
              </button>
            )}
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

      {isUploadOpen && (
        <div className="modal-backdrop" role="presentation" onMouseDown={(event) => {
          if (event.target === event.currentTarget) setIsUploadOpen(false);
        }}>
          <section className="upload-modal" role="dialog" aria-modal="true" aria-labelledby="upload-title">
            <div className="modal-heading">
              <div>
                <p className="label">INGESTA</p>
                <h2 id="upload-title">Subir manual a MinIO</h2>
              </div>
              <button className="icon-button" type="button" onClick={() => setIsUploadOpen(false)} aria-label="Cerrar modal">
                ×
              </button>
            </div>

            <form className="upload-form" onSubmit={handleUpload}>
              <label>
                Bucket
                <select value={uploadBucket} onChange={handleBucketChange}>
                  {storageOptions.length === 0 && <option value="manuals">manuals</option>}
                  {storageOptions.map((bucket) => (
                    <option key={bucket.name} value={bucket.name}>{bucket.name}</option>
                  ))}
                </select>
              </label>

              <label>
                Object key existente
                <select
                  value={storageOptions.find((bucket) => bucket.name === uploadBucket)?.object_keys.includes(uploadObjectKey) ? uploadObjectKey : ''}
                  onChange={(event) => setUploadObjectKey(event.target.value)}
                >
                  <option value="">Seleccionar o escribir uno nuevo</option>
                  {(storageOptions.find((bucket) => bucket.name === uploadBucket)?.object_keys || []).map((key) => (
                    <option key={key} value={key}>{key}</option>
                  ))}
                </select>
              </label>

              <label>
                Object key
                <input
                  type="text"
                  value={uploadObjectKey}
                  onChange={(event) => setUploadObjectKey(event.target.value)}
                  placeholder="generic_manuals/manual__parte-001.pdf"
                  required
                />
              </label>

              <label>
                Archivo PDF
                <input type="file" accept="application/pdf,.pdf" onChange={handleUploadFileChange} required />
              </label>

              <p className="upload-hint">Formato requerido: &lt;nombre_manual&gt;__parte-xxx.pdf</p>
              {uploadError && <div className="error-box compact">{uploadError}</div>}
              {uploadMessage && <div className="success-box">{uploadMessage}</div>}

              <div className="modal-actions">
                <button className="ghost" type="button" onClick={() => setIsUploadOpen(false)}>Cerrar</button>
                <button type="submit" disabled={uploading || !uploadFile || !uploadObjectKey.trim()}>
                  {uploading ? 'Subiendo...' : 'Subir y encolar'}
                </button>
              </div>
            </form>
          </section>
        </div>
      )}
    </div>
  );
}

export default App;
