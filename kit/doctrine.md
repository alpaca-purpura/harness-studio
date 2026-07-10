# Doctrina ArnesIA — overlay del conductor (cuerpo ①, siempre presente)

Trabajas DENTRO de ArnesIA, la fábrica de arneses: el cwd de esta sesión es el repo de UN
arnés (cuerpo ③, el producto). Tu trabajo es crearlo, auditarlo o mejorarlo CONFORME a la
doctrina. Reglas que no se negocian:

## Anatomía (VISION A1–A7)

- Un arnés opera un proceso por **rol × proceso**: fases → cada fase tiene **cajas**; una
  caja = una skill-frente con contrato. Guardia (hooks) y Base (docs/architecture/knowledge/reglas) son
  bandas transversales, NO cajas.
- El trabajo lleva su estado por el **spine que el arnés declara** (`arnes.l0.json`); una
  caja posee UNA transición del spine (`estado: "a -> b"`). Jamás inventes estados fuera
  del spine declarado.

## Contrato de caja fusionado (METODOLOGIA §3)

Toda caja vive en `skills/<id>/SKILL.md` y su frontmatter ES el contrato — tres ejes:

1. **INTENCIÓN**: `why` · `capabilities[]` (cada una con `success` verificable) ·
   `constraints` · `non_goals`.
2. **CLASIFICACIÓN** (ortogonales): `clase` (enum de 10 primitivas CC-native) ·
   `arquetipo` (`pipeline`=automatización · `excepcion` · `abierto`=framed autonomy ·
   `no-arnesar`=NO es caja, se deja fuera con registro) · `perfil_harness` (T1/T2/T3;
   T3 = el conductor Go es dueño del loop). Precedencia: arquetipo > perfil.
3. **CABLEADO + ACEPTACIÓN**: `caja: true` · `fase` · `estado` · `necesita[]` (orígenes:
   usuario | base: | caja: | libreria: | maquinaria: | terceros: | marcas-dormidas:) ·
   `entrega[]` (UN `escritor_unico` por artefacto) · `ruta[]` · `gate` (HONESTO:
   `tipo: none` si no hay eval real; Gherkin ejecutable si `auto`) · `handoff` a humano.

## Reglas de honestidad (transversales)

- **Gris ≠ verde**: nada se inventa; un dato sin fuente se marca, un gate sin eval es
  `tipo: none`, un elemento no reconocido queda VISIBLE como `no-reconocido`.
- **Firewall CC-native** (§8.6): PROHIBIDO frontmatter que Claude Code ignora
  (`persistent_facts`, `activation_steps_prepend`, `customize`, `sanctum`). Si una clave
  no hace nada en CC, declararla es un hallazgo.
- **Aditivo sin pérdida**: mejoras nunca borran historia; supersede con nota, no
  reescribas lo firmado.
- **Nomenclatura** (`no reconocible = no existe`): skills en `skills/<id>/SKILL.md`
  (directorio, no archivo plano) · manifiesto `arnes.l0.json` en la raíz · el mapeo
  completo clase→ubicación es el estándar del reconocimiento (nomenclatura-arnes v1).

## Frontera de cuerpos (§9 — regla dura)

Esta maquinaria (el kit que te da estas skills) se inyecta por flags y vive FUERA del
repo del arnés: **jamás escribas archivos del kit, de ArnesIA o de esta doctrina dentro
del arnés** (② no contamina ③). En el arnés solo escribes lo que su nomenclatura define.

Referencia completa (read-only): el directorio `knowhow/` inyectado con `--add-dir`
contiene los nodos del estándar (checklists evaluables por elemento). Consúltalos al
crear o auditar cada elemento; el veredicto mecánico lo da `arnesia conformance`.
