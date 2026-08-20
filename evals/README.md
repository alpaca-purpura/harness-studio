# Evals de las skills del kit

Un eval mide si una skill **cambia la conducta del modelo**. No reemplaza a los tests: los
tests prueban que el archivo existe y se materializa; el eval prueba que, con la skill
puesta, la sesión hace lo que la skill dice.

Es lo que pide el propio estándar —
[`knowledge/elements/skills.md`](../docs/architecture/knowledge/elements/skills.md) L1.6
(«autoría dirigida por evals… baseline sin la skill → ≥3 escenarios») y el check
`skill-has-evals`.

## Por qué viven acá y no en `kit/`

`kit/` entra al binario por `//go:embed all:kit`, y el daemon está al **94,5 % del techo de
peso**. Los casos y el runner no tienen por qué viajar a la máquina de cada usuario: son
herramientas de desarrollo. `evals/` queda fuera del embed y fuera del gate R2 de
capabilities.

## Correrlos

```bash
python evals/forjar-arnes/run.py            # todos los casos
python evals/forjar-arnes/run.py --caso crear-desde-cero
python evals/forjar-arnes/run.py --conservar # no borra los temps, para inspeccionar
```

Sale con código ≠0 si algún caso falla, así sirve en CI.

## Cómo funciona

El runner **replica la inyección** que hace
[`provisioner.go::materialize`](../internal/adapters/provision/provisioner.go) —
`kit/` + `knowhow/` + `doctrine.md` en un temp— y lanza `claude -p` con los mismos flags
que arma
[`SpawnArgs`](../internal/adapters/agent/claudecode/conductor.go). Después parsea el
stream-json y evalúa qué hizo la sesión.

No usa el daemon a propósito: el spawn real lleva `--permission-mode default
--permission-prompt-tool stdio`, o sea que **cada escritura espera a un humano en el
Dock**. Headless nadie aprueba y la forja se cuelga.

### Divergencias declaradas respecto del spawn real

| Qué | Spawn real | Eval | Por qué |
|---|---|---|---|
| Permisos | `default` + prompt-tool stdio (HITL) | `acceptEdits`, confinado al temp | Es la única forma de correr desatendido. Lo que se mide es la conducta del modelo, no el HITL |
| MCP | `--mcp-config` + `--strict-mcp-config` | sin MCP | El kit hoy no declara servidores propios; el archivo quedaría vacío igual |
| Origen de la inyección | `~/.arnesia/` (materializado por el daemon) | temp armado por el runner | El provisioning real es perezoso (se dispara al spawnear); el eval no depende de que haya corrido |

Que el árbol materializado sea idéntico al del provisioner ya lo fija
`internal/adapters/provision/provisioner_test.go`. El eval cubre la otra mitad: qué hace el
modelo cuando lo recibe.

## Los prompts entregan TODO cerrado, a propósito

`forjar-arnes` es una caja **T2** (multi-paso interactivo): su paso 2 es un grill, y la
doctrina le prohíbe inventar lo que falta. Una sesión `claude -p` es de **un solo turno**.

La primera versión del caso `crear-desde-cero` omitía `reporta_a` —campo requerido por el
schema— y la sesión se frenó a preguntarlo, con el spine ya propuesto sobre la mesa. Eso es
la conducta **correcta**: inventar un campo obligatorio sería el pass fabricado que el
principio 10 prohíbe. Pero el assert quedaba en rojo por un motivo que no era el que el caso
decía medir.

Por eso cada prompt entrega el rol, el proceso, `reporta_a`, la empresa, las fases y el
spine completo. **No es hacerle el trabajo**: lo que se mide sigue siendo si abre el
estándar, si lo declara, y si el manifiesto que escribe es válido y coherente con el spine
declarado. Lo que se elimina es la única variable que un turno único no puede resolver.

Si algún día el eval necesita medir el grill en sí, eso pide otro runner (multi-turno con
respuestas guionadas), no aflojar estos asserts.

## Costo

Cada corrida completa son 4 sesiones reales de Claude. No es gratis y no corre solo.
