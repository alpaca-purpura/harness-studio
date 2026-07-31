# Decisiones — definición canónica de arnés

> Registro same-turn (disciplina §10). Firmas vía AskUserQuestion 2026-07-30.

- **DEF-D1 — FIRMADA 🧑‍⚖️ 2026-07-30 · feature/bugfix/spike = TIPOS de paquete de
  trabajo con pipeline propio DENTRO del mismo arnés-dev, NO arneses separados.**
  Aplica D20 del modelo de terreno (multi-pipeline por tipo, firmada 2026-07-10) al
  caso dev concreto: mismo rol → mismo arnés; cada tipo declara
  `{jerarquia, estados, criterio_cierre}` — el spike cierra en aprendizaje
  documentado (jamás producción), el bugfix salta discovery, la feature corre el
  spine completo. Regla mnemotécnica ratificada: **¿cambia QUIÉN opera? → otro
  arnés; ¿cambia solo la RUTA? → otro pipeline del mismo arnés.**
- **DEF-D2 — ABIERTA (el operador pidió debatir, 2026-07-30).** Definición v1
  propuesta («arnés = rol × proceso · 1 arnés = 1 plugin · end-to-end = cadena de
  arneses») NO ratificada — el operador ve un caso a discutir antes de firmar.
  v1, casos duros y la v2 propuesta → [`debate-definicion.md`](./debate-definicion.md).
  Nada de la v1 se cementa hasta cerrar este debate.
- **DEF-D3 — DIRECTIVA «diseñar ahora» (2026-07-30), diseño propuesto SIN firma.**
  La cadena (proceso end-to-end que atraviesa varios arneses) hoy no tiene entidad
  declarada (solo `reporta_a`/`empresa` en META + Galaxia visual). El operador
  eligió diseñarla ya, no diferir. Propuesta: `proceso` como entidad de primer
  orden declarada en el **home** (marketplace), mapa fases→arnés-owner con
  `gate_salida` verificable por fase; caso empresa-chica = mapeo N:1 (un arnés
  posee todas las fases) → `debate-definicion.md` §5. Se firma junto con DEF-D2.
- **DEF-D4 — registro (hallazgo, no decisión): gentle-ai analizado 2026-07-30.**
  NO es un plugin CC ni contiene arneses según nuestra definición: es un
  configurador de ecosistema transversal (14 agentes, memoria Engram, SDD opcional,
  receipts RDD, skill registry), outcome-first, **sin modelo de rol ni de proceso
  de negocio**. No compite con el arnés — compite con el harness crudo. Ideas a
  evaluar: instalación modular con scope por workspace · «receipt» único que
  validan todas las compuertas (nosotros: PARIDAD + conformance, más fuerte pero
  menos ergonómico). Detalle → `debate-definicion.md` §6.
