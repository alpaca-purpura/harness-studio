# Decisiones del inspector/drawer — log por iteración

> Una entrada por decisión conversada sobre el mockup. Estado: PROPUESTA → FIRMADA.
> Regla: ninguna decisión entra al spec sin firma; ninguna llega al código sin spec.

| # | Fecha | Decisión | Porqué | Estado |
|---|-------|----------|--------|--------|
| 1 | 2026-07-07 | **Botón «ampliar» en el header del drawer**, al costado del de cerrar (⤢). Al click, el drawer ocupa TODO el espacio visual del mapa; mismo botón (⤡) o `Esc` restauran. En modo amplio el contenido reparte las secciones en columnas para aprovechar el ancho. | Operador: «visualizar con más detalles» — 340px queda corto para contratos ricos. | **FIRMADA** (operador, 2026-07-07 · v2) |
| 2 | 2026-07-07 | **Tooltips en dos niveles**: (a) cada título de sección (CLASIFICACIÓN, FUENTE, …) lleva una «i» al costado con tooltip que explica qué agrupa; (b) cada nombre de campo (banda, fase, canal, procedencia, …) se subraya PUNTEADO (affordance de hover) y su tooltip explica qué significa el campo Y qué significa el valor concreto que muestra. Aplica a todos los elementos. | Operador: el drawer debe autoexplicarse — doctrina legible sin salir del nodo (guía sin bloqueo). **Restricción del operador (mismo turno): el contenido de CADA tooltip debe ser consistente con nuestra doctrina** — fuentes: METODOLOGIA §3 (4 ejes del contrato) · §4 (honestidad, gris≠verde) · §8.1/§8.2 (arquetipo/perfil) · VISION A1–A7 · schema L0 (banda/canal/origen) · rules L2.6 (carga). El diccionario vive en el mockup (`SEC_TIP`/`DEF_CAMPO`/`DEF_VALOR`) y al implementar se cementa en la entity FE con las mismas fuentes. | PROPUESTA (v3 del mockup) |
