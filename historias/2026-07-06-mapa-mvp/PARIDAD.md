# Checklist de paridad visual — Mapa MVP (Fase C, revisión §5)

> Round-trip: **cada elemento del mockup ⇒ está en `spec.md` (comportamiento) + `design.md` (visual) con
> refs**; **cada ref ⇒ existe y es exacta**. Tokens = var corta (`--c-*`/`--ok`…), no `--color-kind-*`.
> Screenshots (fixture Luana, 1680×1000, consola limpia): shot1 overview-light · shot2 overview-dark ·
> shot3 hover · shot4 reglas-colapsadas · shot5 reglas-expandidas · shot6 ayuda · shot7 dogfood
> (`mockups/.shots/mapa-*.png`).

| Elemento | Token(s) | mockup:línea | Componente Storybook | shot# | spec | design |
|----------|----------|--------------|----------------------|-------|------|--------|
| **Región Guardia** (tinte hook) | `--c-hook`,`--card`,`--border` | `:55,408` | `region`* / `band` compact | 1,2,7 | RF-11 | §3.1 |
| **Región Proceso** (tinte primary) | `--primary`,`--card` | `:56,417` | `region`* / `lane` | 1,2 | RF-12 | §3.1 |
| **Región Base** (tinte knowledge) | `--c-knowledge`,`--card` | `:57,440` | `region`* / `base-band`* | 1,2 | RF-13 | §3.1 |
| **region-hd** (h2 uppercase + rule) | `--text-xs`,`--muted-foreground`,`--border` | `:58-60` | `region`* | 1 | RF-14 | §3 |
| **Nodo base** (border-left 3px + glyph + nombre) | `--border`,`--secondary`,`--radius-md`,`--text-sm` | `:91,109-110` | `arnes-node.tsx` | 1 | RF-20 | §5 |
| **Handle** (`node-cmd`, derivado clase) | `--card`,`--border`,`--foreground` | `:112,327-334` | `arnes-node`* | 1 | RF-21 | §5.2 |
| **Glyph** 17px (6 formas + 10 clases) | `--c-*`,`--background` | `:116-120,203-214` | `glyph.tsx`+`kind.ts` | 1 | RF-20 | §4 |
| **Variante caja** (border-left 5px + tint + badge) | `--tc`(=`--c-*`),`--secondary` | `:104-105,343` | `arnes-node`* `caja` | 1,7 | RF-22 | §6.1 |
| **Transición spine** (`node-trans` ◇) | `--tc`,`--radius-full` | `:114-115,349` | `transition-tag`* | 1,7 | RF-23 | §5.1 |
| **Variante apoyo** (indent + hairline) | `--secondary`,`--border` | `:95-96,427-429` | `lane`* support | 1 | RF-24 | §6.3 |
| **Orden caja-first** | — | `:424` | `lane`* | 1 | RF-25 | §7 |
| **Variante compacta** (Guardia chip) | — | `:99-102,412` | `band`* compact | 1,7 | RF-26 | §6.4 |
| **Variante puesto** (border dashed) PROPUESTA | `--tc` | `:107,229` | `arnes-node`* `puesto` | 1 | RF-27 | §6.2 |
| **prop-badge** (propuesto) PROPUESTA | `--warn` | `:108,278,344` | `arnes-node`* `prop` | 1 | RF-28 | §6.2 |
| **Variante dim / hover** | `--tc` | `:92-93,490-496` | `arnes-node`* `dim` | 3 | RF-33 | §6.5 |
| **Edge invoca** (rojo, flecha) | `--crit` | `:240` | `edge-layer.tsx` | 1,3 | RF-30 | §9.1 |
| **Edge escribe** (verde, dash 3 3) | `--ok` | `:241` | `edge-layer`* | — | RF-30 | §9.1 |
| **Edge lee** (ámbar, dash 4 4, sin flecha) | `--warn` | `:242` | `edge-layer`* | 3 | RF-30 | §9.1 |
| **Spine-always + hover-reveal + foco** | — | `:465-467,484` | `use-edge-paths.ts` | 1,3 | RF-31/32 | §9.3 |
| **Bezier** (dx=max(30,\|Δx\|/2), ÷z) | — | `:475,478` | `use-edge-paths`* | — | RF-35 | §9.2 |
| **Lane** (232px, hd + count) | `--card`,`--radius-lg`,`--border`,`--secondary` | `:85-89` | `lane.tsx` | 1,7 | RF-12 | §7 |
| **Base subband Reglas** (colapsable) | `--secondary`,`--border`,`--radius-md` | `:71-77,363-386` | `rules-subband`* | 4,5 | RF-40/41 | §8.4 |
| **Split siempre/condicional** | `--crit`,`--warn` | `:378-379` | `rules-subband`* | 5 | RF-42 | §8.4 |
| **Knowledge & servicios** (leído) | `--c-knowledge` | `:389-392` | `base-band`* | 1 | RF-43 | §8.4 |
| **act-chip** (activación) PROPUESTA | `--ac`(crit/warn/mfg/knowledge) | `:68,231-238,354` | `activation-chip`* | 1 | RF-44 | §8.2-8.3 |
| **Banda dormida** (opacity .55) | — | `:63,223` | `band`* | — | RF-44 | §8.2 |
| **Controles zoom** (+/−/⤢ 36px) | `--card`,`--border`,`--radius-md` | `:124-126,163-167` | `use-viewport`* | 1 | RF-51 | §10 |
| **Toggle fixture / picker** | `--secondary`,`--foreground` | `:125-129,168-171` | (chrome) | 1,7 | RF-55 | §10 |
| **help-fab + panel** | `--card`,`--border`,`--radius-lg` | `:132-134,173-199` | (chrome) | 6 | RF-56 | §10 |
| **fit / overview-first** | — | `:453-459,517` | `use-viewport`* | 1,7 | RF-50 | §2 |
| **pan (drag) / ctrl-wheel** | — | `:505-508` | `use-viewport`* | — | RF-52 | §3.3 |
| **Theming light+dark** | todos | `:21-41` | `theme.css` | 1↔2 | RF (§11) | §11 |

`*` = componente **nuevo o a modificar** (delta en `architecture.md §6`). Cero magic-value de color;
literales de medida documentados en `design.md §12`.

## Estado deferred (declarado honesto, no verde falso)

| Ítem | Estado | Ref |
|------|--------|-----|
| Capas Tokens/Desempeño/Proceso | **staged** (telemetría JSONL ausente) | spec RF-60, design §1/§11 |
| Campos `origen`/`alw` | **PROPUESTA** (ausentes en schema+dominio) | spec §2.3, arch §5.1 |
| Campos `trans`/`prop` | derivados de `contract.estado`/`canal` (no campo nuevo) | arch §5.1 |
| 5 vars `--c-*` config | en source, **no regeneradas** a theme.css/tokens.ts | design §1.1, arch §6a |
| `dependency-cruiser` / `vitest.workspace` | **rotos** (no ejecutan story=test/canvas⊥chrome) | arch §1.5/§7 |
| `go-arch-lint` | binario+config **ausentes** | arch §1.5 |
| 5 boundaries FE | **`proposed`, ninguno `enforced`** | arch §1.1 |
| Seed backend | **3 nodos/1 edge keyed "demo"** (no 2, no dev-full-cycle) — corrige CLAUDE.md | arch §4.1 |
| Inspector + picker | **Hito 2** (getNode 501, listHarnesses 200 vacío) | spec §9.2, arch §4.2 |
