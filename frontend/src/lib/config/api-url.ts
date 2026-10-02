/** Pure helper — safe to import from Vitest without SvelteKit env. */
export function normalizeApiBaseUrl(value: string | undefined, fallback = '/api/v1'): string {
  return (value ?? fallback).replace(/\/$/, '');
}

export function isLocalDevApiUrl(configured: string): boolean {
  if (!configured.startsWith('http://') && !configured.startsWith('https://')) return false;
  try {
    const { hostname, port } = new URL(configured);
    return (hostname === 'localhost' && port === '8000') || hostname === 'api.localhost';
  } catch {
    return false;
  }
}

export function getApiBaseUrl(configured: string): string {
  if (typeof window === 'undefined') return configured;
  const base = normalizeApiBaseUrl(configured);
  if (!base.startsWith('http')) return base;
  if (isLocalDevApiUrl(base)) return '/api/v1';
  return base;
}

export function toSameOriginApiUrl(url: string): string {
  if (url.startsWith('http://localhost:8000')) {
    return url.slice('http://localhost:8000'.length);
  }
  if (url.startsWith('http://api.localhost')) {
    return url.slice('http://api.localhost'.length);
  }
  return url;
}
