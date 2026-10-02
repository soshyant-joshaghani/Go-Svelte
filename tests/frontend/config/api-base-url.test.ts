import { describe, expect, it } from 'vitest';
import { normalizeApiBaseUrl } from '../../../frontend/src/lib/config/api-url';

describe('normalizeApiBaseUrl', () => {
  it('defaults to /api/v1 without trailing slash', () => {
    expect(normalizeApiBaseUrl(undefined)).toBe('/api/v1');
  });

  it('strips trailing slash', () => {
    expect(normalizeApiBaseUrl('/api/v1/')).toBe('/api/v1');
  });
});
