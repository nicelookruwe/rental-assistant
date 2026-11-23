import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// В разработке фронтенд живёт на :5173 и проксирует API на бэкенд :8080.
// Адрес бэкенда при необходимости меняется правкой apiTarget.
const apiTarget = 'http://localhost:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
