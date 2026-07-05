// SEED declarado (fase 3 as-code); se completa al aterrizar web/src (fase 5).
// Workspace de Vitest con el proyecto `storybook` (story-as-test). enforced_by de
// arch/boundaries/fe-visual-fitness.md (`vitest.workspace.ts#storybook`) y ci.yml:visual-fitness.
// verificar versión exacta al instalar (Storybook 10: plugin en `@storybook/addon-vitest`;
// browser mode Playwright Chromium — jsdom no mide React Flow/CodeMirror).

import { storybookTest } from "@storybook/addon-vitest/vitest-plugin";
import { defineWorkspace } from "vitest/config";

export default defineWorkspace([
	{
		plugins: [storybookTest({ configDir: ".storybook" })],
		test: {
			name: "storybook",
			browser: {
				enabled: true,
				provider: "playwright",
				headless: true,
				instances: [{ browser: "chromium" }],
			},
		},
	},
]);
