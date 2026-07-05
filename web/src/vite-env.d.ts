/// <reference types="vite/client" />

// Typed env so `import.meta.env.VITE_ARNESIA_API` is a known property (dot access) —
// avoids the index-signature bracket that would otherwise be needed under strictest.
interface ImportMetaEnv {
  readonly VITE_ARNESIA_API?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
