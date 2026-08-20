---
story_id: 2026-07-20-shell-topbar-selector-arnes
state: done
release: v0.7-mvp-instalable
module: shell
# Sembrado por scripts/backfill_checkpoints.py — el paquete es anterior a la disciplina
# de checkpoint (METODOLOGIA §10) y sin este archivo el board no lo ve.
# Estado inferido de: PARIDAD con gate 🧑‍⚖ firmado: «## Retomar aquí — GATE HUMANO FIRMADO 🧑‍⚖️»
# Corregilo a mano si la inferencia erró: este archivo es la fuente, no el script.
backfilled: true
chris_verify:
  # NUNCA se deduce una firma desde un grep (§4: gris ≠ verde). Si el gate 🧑‍⚖️ del
  # PARIDAD está firmado, ponelo en true a mano citando la línea que lo firma.
  signoff: false
---

# 2026-07-20-shell-topbar-selector-arnes

Checkpoint sembrado para dar visibilidad al paquete en el board. El contenido real del
paquete vive en su `INDEX.md` y en los artefactos hermanos.
