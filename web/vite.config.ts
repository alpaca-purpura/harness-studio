import { writeFileSync } from "node:fs";
import { fileURLToPath, URL } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, type Plugin } from "vite";

// dist/.gitkeep viaja COMMITEADO para que el go:embed de la SPA (embed_webdist.go,
// `all:web/dist`) compile en un checkout limpio (CI, clone fresco). emptyOutDir lo
// borra en cada build — este hook lo regenera para que git jamás vea la deleción.
const keepGitkeep = (): Plugin => ({
	name: "arnesia:keep-gitkeep",
	closeBundle() {
		writeFileSync(fileURLToPath(new URL("./dist/.gitkeep", import.meta.url)), "");
	},
});

// SPA de ArnesIA. Sirve a Tauri 2 en dev (devUrl :5173) y se compila a ../dist para
// go:embed / bundle. Alias @ = src (espeja tsconfig paths).
export default defineConfig({
	plugins: [react(), tailwindcss(), keepGitkeep()],
	resolve: {
		alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
	},
	// Tauri: puerto fijo, no limpiar el screen para ver errores del daemon.
	clearScreen: false,
	server: { port: 5173, strictPort: true },
	build: { outDir: "dist", target: "es2022", sourcemap: true },
});
