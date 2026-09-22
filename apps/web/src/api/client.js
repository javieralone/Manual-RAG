export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export const fetchWithSession = (url, options = {}) => fetch(url, { ...options, credentials: 'include' });

export function buildHeaders(token, extra = {}) {
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...extra,
  };
}

export async function handleApiResponse(response) {
  if (!response.ok) {
    let message = 'Error del servidor';
    try {
      const data = await response.json();
      if (data && data.error) {
        message = data.error;
      } else if (data && data.detail) {
        message = data.detail;
      }
    } catch {
      const text = await response.text();
      if (text) message = text;
    }

    throw new Error(message);
  }

  return response.json();
}
