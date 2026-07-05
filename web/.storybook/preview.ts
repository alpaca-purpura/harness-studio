import type { Preview } from "@storybook/react-vite";
import "../src/app/styles/index.css";

// Toolbar de tema → setea data-theme en el <html> del preview (mismo mecanismo que el shell).
const preview: Preview = {
  parameters: {
    a11y: { test: "error" },
    layout: "centered",
  },
  globalTypes: {
    theme: {
      description: "Tema",
      defaultValue: "light",
      toolbar: {
        icon: "circlehollow",
        items: [
          { value: "light", title: "Light" },
          { value: "dark", title: "Dark" },
        ],
        dynamicTitle: true,
      },
    },
  },
  decorators: [
    (Story, context) => {
      document.documentElement.dataset.theme = context.globals.theme;
      return Story();
    },
  ],
};

export default preview;
