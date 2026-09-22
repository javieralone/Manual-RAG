import { API_BASE_URL, buildHeaders, fetchWithSession, handleApiResponse } from './client';

export async function login(username, password) {
  const response = await fetchWithSession(`${API_BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify({ username, password }),
  });

  if (!response.ok) {
    throw new Error('Credenciales inválidas');
  }

  return handleApiResponse(response);
}

export async function refreshToken() {
  const response = await fetchWithSession(`${API_BASE_URL}/api/v1/auth/refresh`, {
    method: 'POST',
  headers: buildHeaders(),
  body: JSON.stringify({}),
  });

  return handleApiResponse(response);
}

export async function logout() {
  const response = await fetchWithSession(`${API_BASE_URL}/api/v1/auth/logout`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify({}),
  });

  if (!response.ok && response.status !== 204) {
    await handleApiResponse(response);
  }
}
