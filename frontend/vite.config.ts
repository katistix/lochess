import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import { writeFileSync } from 'node:fs'

export default defineConfig({
  // The router plugin must come before react().
  plugins: [
    tanstackRouter({ target: 'react', autoCodeSplitting: true }),
    react(),
    tailwindcss(),
    {
      // Vite empties dist on every build; Go's //go:embed needs the folder
      // (and this tracked file) to exist, even on a fresh clone.
      name: 'keep-dist',
      apply: 'build',
      closeBundle: () => writeFileSync('dist/gitkeep', ''),
    },
  ],
})
