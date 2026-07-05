import StyleDictionary from "style-dictionary";

// Pipeline DTCG (tokens/base.tokens.json) → CSS themeable (:root + [data-theme=dark]) + TS.
// enforced_by de arch/boundaries/fe-tokens-contrato.md. Dark = $extensions.mode.dark.
// Correr: `npm run tokens:build`. Las salidas (src/app/styles/theme.css, src/shared/config/tokens.ts)
// se commitean para runnabilidad pre-install; se REGENERAN con este comando cuando los tokens cambian.
//
// NOTA: SD v5 no maneja "modos" nativamente; la lógica dark vive en el formato custom de abajo,
// que lee token.original.$extensions.mode.dark y resuelve referencias {…} contra los primitivos.

/** Nombre de CSS var (shadcn + convención mockup) a partir del path DTCG. */
function cssVar(path) {
	const [g0, g1, ...rest] = path;
	if (g0 === "color") {
		if (g1 === "primitive") return null; // crudos: no se emiten
		if (g1 === "semantic") return `--${rest.join("-")}`;
		if (g1 === "kind") return `--c-${rest.join("-")}`;
		if (g1 === "health") return `--${rest.join("-")}`;
		if (g1 === "chart") return `--chart-${rest.join("-")}`;
		if (g1 === "heat") return `--heat-${rest.join("-")}`;
		if (g1 === "sidebar")
			return rest[0] === "background"
				? "--sidebar"
				: `--sidebar-${rest.join("-")}`;
	}
	return `--${path.join("-")}`;
}

/** Resuelve un valor que puede ser referencia DTCG "{a.b.c}" contra el árbol de tokens. */
function resolve(value, byPath) {
	if (typeof value !== "string")
		return Array.isArray(value) ? value.join(", ") : value;
	const m = value.match(/^\{(.+)\}$/);
	if (!m) return value;
	const t = byPath.get(m[1]);
	return t ? resolve(t.$value ?? t.value, byPath) : value;
}

StyleDictionary.registerFormat({
	name: "css/theme-modes",
	format: ({ dictionary }) => {
		const byPath = new Map(
			dictionary.allTokens.map((t) => [t.path.join("."), t.original]),
		);
		const light = [];
		const dark = [];
		for (const t of dictionary.allTokens) {
			const name = cssVar(t.path);
			if (!name) continue;
			light.push(`  ${name}: ${resolve(t.original.$value, byPath)};`);
			const d = t.original.$extensions?.mode?.dark;
			if (d != null) dark.push(`  ${name}: ${resolve(d, byPath)};`);
		}
		return [
			"/* GENERADO por `npm run tokens:build` — NO editar a mano. Fuente: tokens/base.tokens.json */",
			":root {",
			...light,
			"}",
			"",
			':root[data-theme="dark"] {',
			...dark,
			"}",
			"",
		].join("\n");
	},
});

StyleDictionary.registerFormat({
	name: "ts/tokens-flat",
	format: ({ dictionary }) => {
		const byPath = new Map(
			dictionary.allTokens.map((t) => [t.path.join("."), t.original]),
		);
		const lines = dictionary.allTokens
			.map((t) => {
				const name = cssVar(t.path);
				if (!name) return null;
				return `  "${name.replace(/^--/, "")}": ${JSON.stringify(resolve(t.original.$value, byPath))},`;
			})
			.filter(Boolean);
		return [
			"// GENERADO por `npm run tokens:build` — NO editar a mano.",
			"export const tokens = {",
			...lines,
			"} as const;",
			"",
			"export type TokenName = keyof typeof tokens;",
			"",
		].join("\n");
	},
});

export default {
	source: ["tokens/*.tokens.json"],
	platforms: {
		css: {
			transformGroup: "css",
			buildPath: "src/app/styles/",
			files: [{ destination: "theme.css", format: "css/theme-modes" }],
		},
		ts: {
			transformGroup: "js",
			buildPath: "src/shared/config/",
			files: [{ destination: "tokens.ts", format: "ts/tokens-flat" }],
		},
	},
};
