import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import solidPlugin from "vite-plugin-solid";
import removeConsole from 'vite-plugin-remove-console';
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const packageJson = JSON.parse(
  readFileSync(join(__dirname, "package.json"), "utf-8")
);

export default defineConfig({
  root: ".",
  plugins: [tailwindcss(), solidPlugin(), removeConsole()],
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:8285",
        changeOrigin: true,
        ws: true, // Enable WebSocket proxying for /api routes
      },
    },
  },
  preview: {
    proxy: {
      "/api": {
        target: "http://localhost:8285",
        changeOrigin: true,
        ws: true, // Enable WebSocket proxying for /api routes in preview mode
      },
    },
  },
  build: {
    target: "esnext",
  },
  define: {
    PACKAGE_VERSION: JSON.stringify(packageJson.version),
  },
});
