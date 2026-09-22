import { API_BASE_URL, buildHeaders } from './client';

export async function askNormalQuery({ question, collection, token, signal }) {
  const response = await fetch(`${API_BASE_URL}/api/v1/query`, {
    method: 'POST',
    headers: buildHeaders(token),
    body: JSON.stringify({ question, collection }),
    signal,
  });

  if (!response.ok) {
    let message = 'Error al consultar';
    try {
      const data = await response.json();
      message = data.error || data.detail || message;
    } catch {
      message = await response.text();
    }
    throw new Error(message);
  }

  return response.json();
}

export async function askStreamQuery({
  question,
  collection,
  token,
  signal,
  onToken,
  onMetadata,
  onComplete,
  onError,
}) {
  const response = await fetch(`${API_BASE_URL}/api/v1/query/stream`, {
    method: 'POST',
    headers: buildHeaders(token),
    body: JSON.stringify({ question, collection }),
    signal,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || 'Error al iniciar el stream');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const parts = buffer.split('\n\n');
    buffer = parts.pop() || '';

    for (const part of parts) {
      const eventLine = part.split('\n').find((line) => line.startsWith('event: '));
      const dataLine = part.split('\n').find((line) => line.startsWith('data: '));

      if (!eventLine || !dataLine) continue;

      const event = eventLine.replace('event: ', '').trim();
      const payload = dataLine.replace('data: ', '').trim();

      if (!payload) continue;

      try {
        const json = JSON.parse(payload);

        if (event === 'metadata') {
          onMetadata?.(json);
        }

        if (event === 'token') {
          onToken?.(json.text || '');
        }

        if (event === 'complete') {
          onComplete?.(json);
        }

        if (event === 'error') {
          onError?.(json.error || 'Error en la respuesta en streaming');
        }
      } catch {
        onError?.('Error interpretando el stream');
      }
    }
  }
}
