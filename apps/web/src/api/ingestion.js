import { API_BASE_URL, buildHeaders, handleApiResponse } from './client';

export async function getStorageOptions(token) {
  const response = await fetch(`${API_BASE_URL}/api/v1/ingestion/storage/options`, {
    headers: buildHeaders(token),
  });
  return handleApiResponse(response);
}

export async function uploadDocument({ token, file, bucket, objectKey }) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('bucket', bucket);
  formData.append('object_key', objectKey);

  const response = await fetch(`${API_BASE_URL}/api/v1/ingestion/upload`, {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: formData,
  });
  return handleApiResponse(response);
}