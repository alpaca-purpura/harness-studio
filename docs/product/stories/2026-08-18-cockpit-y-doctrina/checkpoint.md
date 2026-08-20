---
story_id: 2026-08-18-cockpit-y-doctrina
state: developed
release: v0.8-cockpit-y-doctrina
module: tooling
chris_verify:
  # El gate 🧑‍⚖️ de PARIDAD lo firma el operador tras ejercer el cockpit y una sesión de
  # forja. Hasta entonces: false, sin excepción (§4 — un pass fabricado es peor que un hueco).
  signoff: false
---

# Cockpit y doctrina — checkpoint

**Fase:** construido y verificado en vivo; falta el gate 🧑‍⚖️ del operador.

## Retomar aquí

1. Ejercer el cockpit: `powershell -File tools\cockpit\cockpit.ps1 run` → `http://localhost:4300`.
2. Ejercer `forjar-arnes` sobre un directorio vacío y confirmar que la sesión **declara qué
   nodos del knowhow abrió** antes de escribir el manifiesto. Esa declaración es la
   evidencia de que el árbol dejó de ser inerte.
3. Firmar (o rebotar) el gate en [`PARIDAD.md`](./PARIDAD.md).
