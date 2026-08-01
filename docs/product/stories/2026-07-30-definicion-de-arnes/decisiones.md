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
- **DEF-D2 — FIRMADA 🧑‍⚖️ 2026-08-01 · Definición canónica v3 + vocabulario.**
  Tras 2 rounds de debate (v1 rechazada-para-debate → v2 → aporte del operador en
  `chris-input.md` round 2 → v3):
  > **Arnés** = el paquete de know-how de **UN puesto** sobre **UN tipo de
  > terreno**: agrupa las **ACTIVIDADES** del puesto dentro de su tramo del
  > proceso (su macroproceso), cada actividad con su **procedimiento** (pipeline)
  > propio sobre una **base común**. 1 arnés = 1 plugin; **se instala POR
  > CARPETA** — cada instalación opera un terreno concreto. El proceso end-to-end
  > es entidad aparte (DEF-D3); el corte proceso→arneses y la agrupación
  > actividades→arnés se rigen por el criterio de corte. 2 roles → 2 arneses.

  **Vocabulario firmado:** la palabra «arnés» vive a nivel PAQUETE/plugin (como
  todo lo construido: sello, `(home,id)`, Portafolio, Mapa, `--arnes`); el nivel
  interno = **actividad** con su **procedimiento**. **Criterio de corte firmado:**
  ¿otro operador? → separá SIEMPRE (seguridad) · ¿otro terreno? → probablemente ·
  ¿solo otra actividad, mismo operador+terreno? → JAMÁS (pipeline del mismo
  arnés). Jerarquía completa y mapeo → `debate-definicion.md` §4b.
- **DEF-D3 — FIRMADA 🧑‍⚖️ 2026-08-01 · La cadena = `proceso` entidad de primer
  orden en el home.** `proceso/<id>.yaml` declarado en el marketplace de la
  familia (SSoT por ley anti-drift): lista de fases, cada una con arnés-owner +
  `gate_salida` VERIFICABLE (el gate de salida de una fase = contrato de entrada
  de la siguiente). META se afila: `proceso` string → referencia `proceso: <id>` +
  `fases: [...]`. Mapeo fases→arnés N:1-friendly (empresa chica: un arnés posee
  todas las fases; crecer = reasignar, no re-modelar). Galaxia renderiza la cadena
  desde el home; telemetría joinea por `proceso.id`. Visibilidad cross-rol vive
  acá (Mapa/Galaxia read-only), no en los plugins. Diseño → `debate-definicion.md`
  §5. **Quedan a la spec** (no bloquean la firma): ¿copia offline del tramo en
  cada `arnes.yaml`? · ¿versionado del proceso independiente del semver de cada
  arnés?
- **DEF-D4 — registro (hallazgo, no decisión): gentle-ai analizado 2026-07-30.**
  NO es un plugin CC ni contiene arneses según nuestra definición: es un
  configurador de ecosistema transversal (14 agentes, memoria Engram, SDD opcional,
  receipts RDD, skill registry), outcome-first, **sin modelo de rol ni de proceso
  de negocio**. No compite con el arnés — compite con el harness crudo. Ideas a
  evaluar: instalación modular con scope por workspace · «receipt» único que
  validan todas las compuertas (nosotros: PARIDAD + conformance, más fuerte pero
  menos ergonómico). Detalle → `debate-definicion.md` §6.
