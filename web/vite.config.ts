import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import tailwindcss from "@tailwindcss/vite";
import { version as reactVersion } from "react";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  define: {
    __REACT_VERSION__: JSON.stringify(reactVersion),
  },
  resolve: {
    tsconfigPaths: true,
  },
  server: {
    port: 6565,
    open: true,
  },
});
