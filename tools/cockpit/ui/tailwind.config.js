/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './app/**/*.{ts,tsx}',
    './components/**/*.{ts,tsx}',
    './lib/**/*.{ts,tsx}',
    '../../cockpit/ui/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        // Tokens del mockup v0.5.2 (dark theme · azul violeta primario)
        bg: '#0b0e14',
        panel: '#11151d',
        panel2: '#161b26',
        border: '#232936',
        text: '#e6edf3',
        muted: '#7d8590',
        accent: '#7c3aed',
        'accent-hover': '#6d28d9',
        // Estados semánticos (extraídos del mockup)
        success: {
          DEFAULT: '#16a34a',
          fg: '#86efac',
          bg: '#14532d',
        },
        warning: {
          DEFAULT: '#d97706',
          fg: '#fcd34d',
          bg: '#713f12',
        },
        danger: {
          DEFAULT: '#dc2626',
          fg: '#fca5a5',
          bg: '#7f1d1d',
        },
        info: {
          DEFAULT: '#2563eb',
          fg: '#93c5fd',
          bg: '#1e3a8a',
        },
      },
      fontFamily: {
        sans: ['ui-sans-serif', 'system-ui', '-apple-system', 'Segoe UI', 'Roboto', 'sans-serif'],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
};
