import { fileURLToPath, URL } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// SPA de ArnesIA. Sirve a Tauri 2 en dev (devUrl :5173) y se compila a ../dist para
// go:embed / bundle. Alias @ = src (espeja tsconfig paths).
export default defineConfig({
	plugins: [react(), tailwindcss()],
	resolve: {
		alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
	},
	// Tauri: puerto fijo, no limpiar el screen para ver errores del daemon.
	clearScreen: false,
	server: { port: 5173, strictPort: true },
	build: { outDir: "dist", target: "es2022", sourcemap: true },
});
