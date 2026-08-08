import { defineConfig } from 'vitest/config'
import { fileURLToPath } from 'node:url'

export default defineConfig({
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'node',
    globals: true,
    setupFiles: ['./src/utils/uni.mock.ts'],
    include: ['src/**/*.spec.ts'],
    coverage: {
      include: ['src/store/**', 'src/utils/**', 'src/lib/**'],
      reporter: ['text', 'json', 'html'],
    },
  },
})