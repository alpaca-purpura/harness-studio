import { storybookTest } from "@storybook/addon-vitest/vitest-plugin"
import { playwright } from "@vitest/browser-playwright"
import { defineConfig } from "vitest/config"

// Vitest 4 removed `defineWorkspace`/vitest.workspace.ts in favour of `test.projects`, and the
// browser provider is now a factory (`@vitest/browser-playwright`), not the string "playwright".
// The `storybook` project runs every *.stories.tsx as a browser test (story = test, enforced_by
// fe-visual-fitness.md · ci.yml:visual-fitness). `extends: "./vite.config.ts"` inherits the @
// alias + react + tailwind plugins so stories resolve and render with the real tokens. Chromium
// (Playwright) — jsdom can't measure layout for the canvas. Preview annotations (theme/a11y) are
// auto-applied by @storybook/addon-vitest 10.3+, so no setup file is needed.
export default defineConfig({
  test: {
    projects: [
      {
        extends: "./vite.config.ts",
        plugins: [storybookTest({ configDir: ".storybook" })],
        test: {
          name: "storybook",
          browser: {
            enabled: true,
            provider: playwright(),
            headless: true,
            instances: [{ browser: "chromium" }],
          },
        },
      },
    ],
  },
})
