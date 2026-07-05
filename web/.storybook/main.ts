// SEED declarado (fase 3 as-code); se completa al aterrizar web/src (fase 5).
// Config mínima de Storybook 10 para Vite + React. enforced_by de arch/boundaries/fe-visual-fitness.md.
// El glob de stories apunta a `web/src/**` (aún inexistente); es correcto: declara el objetivo.
// verificar versión exacta al instalar (Storybook 10 = ESM-only, Node ≥ 20.16/22.19/24).
import type { StorybookConfig } from "@storybook/react-vite";

const config: StorybookConfig = {
  stories: ["../src/**/*.stories.@(ts|tsx)"],
  addons: ["@storybook/addon-a11y"],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
};

export default config;
