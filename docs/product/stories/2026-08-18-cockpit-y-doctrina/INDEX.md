# Cockpit del proceso + la doctrina que se abre al forjar

> Paquete de trabajo (METODOLOGIA §10). Release: `v0.8-cockpit-y-doctrina`.
> **Retomar aquí:** ver § Estado, al final.

## Por qué existe este paquete

Dos necesidades que llegaron juntas y se sostienen entre sí.

**1 · No había cómo VER el proceso.** El repo tiene 157 capabilities YAML (SSoT funcional,
docs-as-code) y 41 paquetes de trabajo, y ninguna herramienta para mirarlos: el estado
vivía en la prosa de `docs/product/checkpoint.md`. El cockpit de `vitalia-app` ya resuelve
exactamente eso —filesystem-as-DB: lee una ruta y por cada subcarpeta parsea los
`.md`/`.yaml`— y el plan `PLAN-cockpit-y-windows-arnesia.md` § Parte P ya había decidido
vendorearlo. Acá se ejecuta.

**2 · La doctrina no se abría al forjar.** ArnesIA declara que el estándar para construir
cada elemento vive en `docs/architecture/knowledge/` (12 nodos · 138 checks) y lo inyecta a
cada sesión con `--add-dir ~/.arnesia/knowhow`. Pero `--add-dir` da **acceso** al
directorio, no **carga** su contenido: hace falta que alguien lo lea. Y ninguna skill del
kit lo leía como paso obligatorio — el árbol viajaba a cada sesión y quedaba inerte.

## Los hallazgos que originan el trabajo

| # | Hallazgo | Evidencia |
|---|---|---|
| H1 | `skills.md` es el ÚNICO nodo citado por el kit; los otros 11 no aparecen en ninguna skill | 4 menciones de `knowhow/` en todo `kit/` |
| H2 | El nodo se consulta al FINAL: en `forjar-caja` aparece en el paso 8 (verificar), cuando el SKILL.md ya se escribió en el paso 4 | `kit/skills/forjar-caja/SKILL.md:47` |
| H3 | **No existe skill para crear un arnés**: `forjar-caja` asume `arnes.l0.json` ya escrito («si la fase no existe, detente y pregunta») | `kit/skills/forjar-caja/SKILL.md:13-16` |
| H4 | 37 de 41 paquetes no tenían `checkpoint.md` → el board del cockpit los ignoraba | `find docs/product/stories -name checkpoint.md` = 4 |
| H5 | 3 capabilities son documentos YAML sin `---` de cierre; el reader del cockpit devolvía defaults **fabricados** (`capability_id: ""`, `status: "live"`) | `forja/{chequear,sembrar}-semilla.yaml` · `portafolio/identificar.yaml` |

H5 es el más grave de los cinco: no es una vista incompleta, es un dato inventado — justo
lo que el principio 10 prohíbe. `scripts/cap_doctor.py::_load` ya toleraba las dos formas,
así que el repo las daba por válidas mientras el cockpit las leía mal.

## Qué se construyó

- `tools/cockpit/` — copia vendored del cockpit, puerto 4300, con sus diffs tabulados en
  [`README-vendored.md`](../../../../tools/cockpit/README-vendored.md) y fijados por
  `tools/cockpit/go/vendored_arnesia_test.go`.
- `scripts/backfill_checkpoints.py` — siembra los checkpoints faltantes (dry-run por
  defecto; jamás fabrica una firma) y sincroniza la pertenencia a release.
- `docs/product/releases/` — el eje de release, que no existía.
- `kit/skills/forjar-arnes/` + reescritura de `forjar-caja`, `auditar-arnes` y
  `doctrine.md` para que el nodo del estándar se abra ANTES de escribir.
- `evals/forjar-arnes/` — el eval que mide si la skill cambia la conducta de la sesión,
  con su caso baseline sin el kit. Cierra el check `skill-has-evals`, que estaba en warn.
- Los **carriles del CIL**: `docs/process/harness-backlog.md` (L1),
  `docs/process/tech-debt.md` (L3) y `docs/learnings/` (L2) — las vistas Harness y
  Learnings del cockpit estaban apagadas porque no existía la fuente.

## Hojas del paquete

- [`decisiones.md`](./decisiones.md) — las decisiones conversadas, en el mismo turno.
- [`PARIDAD.md`](./PARIDAD.md) — RF ⇒ código ⇒ verificación corrida de verdad.

## Estado

Ver [`checkpoint.md`](./checkpoint.md).
