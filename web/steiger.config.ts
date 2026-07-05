// steiger — linter FSD-estructural (secundario; enforced_by de arch/boundaries/fe-topologia-fsd.md).
// Cubre lo que dependency-cruiser no da gratis: public-api por slice, naming/pluralización de capas,
// insignificant-slice. BETA (0.5.0, no extensible) → nunca el único gate; el CI-gate es dependency-cruiser.
//
// NOTA (honestidad): declarado; se activa con la SPA web/ (fase 5).
import { defineConfig } from "steiger";
import fsd from "@feature-sliced/steiger-plugin";

export default defineConfig([
  ...fsd.configs.recommended,
  {
    // El Organigrama exige cross-import entities via la Public API @x — no lo marques como error.
    // Acotado a la carpeta @x (NO a todo entities): así los cross-imports laterales DIRECTOS (no-@x)
    // siguen bajo la regla; antes el override `./src/entities/**` los silenciaba todos.
    // LIMITACIÓN (verificar al scaffold, fase 5): el enforcement @x-only real depende de dónde
    // reporta `fsd/no-cross-imports` (archivo consumidor vs. archivo @x); si reporta en el consumidor,
    // este scope no basta y el @x-only queda sin enforcer automático hoy — pendiente.
    files: ["./src/entities/*/@x/**"],
    rules: { "fsd/no-cross-imports": "off" },
  },
]);
