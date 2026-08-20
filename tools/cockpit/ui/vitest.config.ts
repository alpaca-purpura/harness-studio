import { defineConfig } from 'vitest/config';
import path from 'node:path';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['lib/**/*.{test,spec}.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['lib/**/*.ts'],
      exclude: ['lib/**/*.{test,spec}.ts', 'lib/__tests__/**'],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname),
      '@/lib': path.resolve(__dirname, 'lib'),
    },
  },
});
