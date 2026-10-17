import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const server = process.env.SERVER_URL || "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/config": server,
      "/contract": server,
      "/contracts": server,
      "/settings": server,
      "/forecast": server,
      "/indicators": server,
      "/monitors": server,
      "/monitor": server,
    },
  },
});
