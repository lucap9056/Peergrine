import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from "path";
import fs from "fs";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => [
            "ion-icon"
          ].includes(tag),
        }
      }
    })
  ],
  resolve: {
    alias: {
      "@Components": path.resolve(__dirname, "src/components"),
      "@API": path.resolve(__dirname, "src/api"),
      "@Src": path.resolve(__dirname, "src"),
      'vue': 'vue/dist/vue.esm-bundler.js',
    }
  },
  server: {
    host: "localhost",
    port: 3000,
    https: {
      key: fs.readFileSync(path.resolve(__dirname, 'ssl/localhost.key')),
      cert: fs.readFileSync(path.resolve(__dirname, 'ssl/localhost.crt'))
    },
    proxy: {
      "/api": {
        target: "http://localhost:80",
        ws: true,
        changeOrigin: true,
      }
    }
  }
})
