import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    sourcemap: false,
    rollupOptions: {
      output: {
        // Mirror BookMyShow's chunking: runtime/vendor/app split.
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (/node_modules[\\/](react|react-dom|scheduler)[\\/]/.test(id)) return 'vendor-primary';
            return 'vendor-secondary';
          }
        },
      },
    },
  },
  ssr: {
    // Bundle styled-components into the server build so its ESM entry is used.
    noExternal: ['styled-components'],
  },
});
