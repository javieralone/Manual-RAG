import { API_BASE_URL, buildHeaders, handleApiResponse } from './client';

export async function login(username, password) {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify({ username, password }),
  });

  if (!response.ok) {
    throw new Error('Credenciales inválidas');
  }

  return handleApiResponse(response);
}

export async function refreshToken(refreshTokenValue) {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
    method: 'POST',
    headers: buildHeaders(),
    body: JSON.stringify({ refresh_token: refreshTokenValue }),
  });

  return handleApiResponse(response);
}
