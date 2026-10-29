import path from 'node:path';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'node',
    include: ['tests/frontend/**/*.test.ts']
  },
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, '../../frontend/src/lib')
    }
  }
});
