import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
export default defineConfig({
  plugins: [vue({ template: { compilerOptions: { isCustomElement: (tag) => tag.startsWith("mdui-") } } })],
  server: { host: true, proxy: { "/api": "http://127.0.0.1:7212" } },
  build: { outDir: "../internal/web/assets", emptyOutDir: true },
});
