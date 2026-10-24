import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
  esbuild: { jsx: "automatic" },
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["favicon.svg"],
      manifest: {
        name: "Smart Mess IIIT Pune",
        short_name: "Smart Mess",
        description:
          "Today's mess menus, crowd levels and meal ratings for IIIT Pune students.",
        lang: "en",
        display: "standalone",
        background_color: "#f8f8f7",
        theme_color: "#0072b2",
        start_url: "/",
        icons: [
          { src: "/pwa-192.png", sizes: "192x192", type: "image/png" },
          { src: "/pwa-512.png", sizes: "512x512", type: "image/png" },
          { src: "/pwa-512-maskable.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
        ],
      },
      workbox: {
        // App-shell caching: precache everything vite emits.
        globPatterns: ["**/*.{js,css,html,svg,png,ico}"],
        // The API is never served from cache (crowd levels must be current).
        navigateFallback: "/index.html",
        runtimeCaching: [],
      },
    }),
  ],
  server: {
    port: 5173,
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.{js,jsx}"],
  },
});
