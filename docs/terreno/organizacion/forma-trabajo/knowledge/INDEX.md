# knowledge · forma-trabajo (enforcement co-locado · D5)

> vig: activo · el conocimiento que se LEE/EJECUTA para hacer cumplir la dimensión `forma-trabajo`.
> Traza **bidireccional** (dato↔enforcement, estilo `codigo-traza-a-capability`): cada regla de la NORMA
> tiene aquí su enforcer, y cada enforcer apunta de vuelta a la NORMA.

**Regla de ubicación (D5):** las rules/skills **operativas del harness** (gates) viven en el kit
(`.claude/` · `lefthook.yml` · CI) — este INDEX las **referencia**, no las copia. Las de CONOCIMIENTO
DEL PROYECTO específicas de esta dimensión se co-locarían aquí como hojas atómicas (hoy no hace falta ninguna).

## Enforcers (qué garantiza cada regla de la NORMA)

| regla (NORMA) | enforcer | dónde vive |
|---|---|---|
| code-style Go | `go vet` / `gofmt` / go-arch-lint | CI (`docs/architecture/conventions/ci.md`) |
| code-style TS/React | Biome (format+lint) | `docs/architecture/conventions/editor.md` · config del kit |
| commits Conventional | hook `commit-msg` (lefthook) | [`lefthook.yml`](../../../../../lefthook.yml) |
| pre-commit / pre-push | hooks lefthook (`estado-cifras`, `capabilities`) | [`lefthook.yml`](../../../../../lefthook.yml) |
| PARIDAD verificada | gate humano 🧑‍⚖️ + `pnpm run verify` | proceso (caja paridad) |
| cifras GENERADAS (N2) | `scripts/estado.sh --check` (drift-gate CI + auto-cura pre-commit) | `scripts/estado.sh` · HS-21 |
| capability por cambio (R1-R4) | `scripts/cap_doctor.py` + boundary `codigo-traza-a-capability` | `docs/architecture/boundaries/` |

> Sin hoja atómica propia por ahora: los enforcers son operativos del kit. Si esta dimensión generara
> una rule/skill de conocimiento NUEVA (no operativa), se agregaría aquí como `<slug>.md` (schema D9).
