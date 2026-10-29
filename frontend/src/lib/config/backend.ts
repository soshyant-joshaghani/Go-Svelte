import { env } from '$env/dynamic/public';
import { getApiBaseUrl, normalizeApiBaseUrl } from './api-url';

const configuredApiBaseUrl = normalizeApiBaseUrl(env.PUBLIC_API_BASE_URL);

export const API_BASE_URL = configuredApiBaseUrl;

export function apiBaseUrl(): string {
  return getApiBaseUrl(configuredApiBaseUrl);
}
