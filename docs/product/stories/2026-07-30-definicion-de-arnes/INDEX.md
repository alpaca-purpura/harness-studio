# Definición canónica de arnés + cadena de proceso — paquete de doctrina

> Nace 2026-07-30 de la pregunta fundacional del operador: «¿qué es un arnés? ¿un
> plugin puede tener múltiples arneses? ¿feature/bugfix/spike son arneses distintos?»
> (transcripta en `chris-input.md`). Es un paquete de DOCTRINA: el entregable es la
> definición firmada + su bajada as-code, no código de app.

**Qué es:** cementar QUÉ ES UN ARNÉS por definición para ArnesIA — el corte
(rol × proceso vs end-to-end vs tipo-de-trabajo), la relación arnés↔plugin, y la
**cadena** (el proceso end-to-end que atraviesa varios arneses) como entidad
declarada. Primero dev como dogfood, luego generaliza a otros rubros (p7).

## Etapas §10

| Etapa | Estado |
|---|---|
| Mockup | n/a (doctrina; la superficie visual es la Galaxia, ya firmada HS-03) |
| Decisiones | ✅ **FIRMADAS 🧑‍⚖️ 2026-08-01** — DEF-D1 (2026-07-30) · DEF-D2 definición v3 + vocabulario · DEF-D3 cadena · DEF-D4 hallazgo gentle-ai |
| Spec | ⬜ ← **SIGUIENTE** |
| Implementación | ⬜ |
| PARIDAD | ⬜ |

## Retomar aquí

Decisiones CERRADAS (2 rounds de debate, `decisiones.md` + `debate-definicion.md`
§4b/§5). Sigue la **spec** de la bajada as-code:
1. **Hoja canónica de la definición** (candidato: `docs/architecture/contracts/`
   como `nomenclatura-arnes.md`, o glosario en `docs/product/`) — v3 + jerarquía
   (proceso→fase→arnés→actividad→caja) + criterio de corte + vocabulario.
2. **Schema `proceso/<id>.yaml`** (cadena, DEF-D3) resolviendo las 2 sub-preguntas
   diferidas: copia offline del tramo en cada `arnes.yaml` · versionado del
   proceso vs semver de arneses.
3. **Afilar META:** `proceso` string → referencia `proceso: <id>` + `fases: [...]`
   (aditivo; no romper loader/scanner ni `nomenclatura-arnes.md`).
4. **Dogfood:** declarar las actividades/tipos-de-paquete del arnés-dev (la
   semilla `.arnesia/proceso/{historia,spike}` del carril A del MVP ya materializa
   2; faltan bugfix · revisar-capability).
