---
name: forjar-arnes
description: Crear un arnés NUEVO y completo desde cero — manifiesto arnes.l0.json (rol × proceso, fases, spine), Guardia, Base y las cajas de cada fase, conforme a la doctrina ArnesIA (usa cuando el usuario pida crear/armar/forjar un arnés, arrancar uno nuevo, o convertir un proceso suyo en arnés; para AGREGAR una caja a un arnés que ya existe, usa forjar-caja).
---

# Forjar un arnés

Creás un arnés completo en el cwd: el manifiesto, las bandas transversales y las cajas.
`forjar-caja` sirve una caja sobre un arnés que YA existe; esto construye el arnés que
esa skill después asume.

## Paso 1 — abrí el estándar (no es opcional)

El directorio `knowhow/` viaja inyectado a esta sesión con `--add-dir`, pero eso da
ACCESO, no lectura: si no lo abrís, los checks del estándar no existen para vos.

**Antes de proponer nada, leé:**

- `knowhow/harness-profile.md` — cómo ejecuta una caja (T1/T2/T3) y cómo se rutea.
- `knowhow/rules.md` — la banda Base: qué va always-on y su presupuesto de contexto.
- `knowhow/hooks.md` — la banda Guardia: eventos, matchers y bloqueo real.

**Declarás en tu respuesta qué nodos abriste, NOMBRANDO cada archivo**, con una línea de lo
que sacaste de cada uno. «Ya leí los tres nodos requeridos» NO es declarar: no dice cuáles
ni prueba nada. Escribí los nombres. Es parte del trabajo, no un trámite.

Después, al escribir cada elemento, abrí SU nodo (`knowhow/<clase>.md`): son 12 y cada uno
trae su checklist evaluable. El resumen de `doctrine.md` no los reemplaza.

## Paso 2 — el grill: rol × proceso

Un arnés opera UN proceso para UN rol (VISION A1). Sin estas respuestas no arranques:

1. **Rol** y **proceso** — «quién» y «qué ciclo de trabajo», en una línea cada uno.
2. **Fases** — los tramos del proceso, en orden.
3. **Spine** — los ESTADOS por los que pasa el trabajo y sus transiciones legales. Es el
   esqueleto: todo lo demás se cuelga de acá y ninguna caja podrá inventar un estado fuera.
4. **`reporta_a`** — a qué rol le rinde (o `null` si es raíz). Es campo REQUERIDO.
5. **Empresa** y, si va a publicarse, `marketplace` + `canal`.

Preguntá lo que falte. Es más barato preguntar el spine ahora que descubrir a la caja 4
que faltaba un estado.

## Paso 3 — el manifiesto `arnes.l0.json`

En la RAÍZ del arnés (nomenclatura v1: `no reconocible = no existe`). **Esta es la forma
exacta — no improvises la estructura:**

```json
{
  "id": "soporte-n1",
  "nombre": "Soporte técnico N1",
  "descripcion": "Atención de tickets de soporte de primer nivel.",
  "rol": "Soporte técnico N1",
  "proceso": "atención de tickets",
  "reporta_a": "Coordinador de soporte",
  "empresas": ["acme"],
  "fases": ["triage", "resolucion", "cierre"],
  "spine": {
    "inicial": "nuevo",
    "terminales": ["cerrado"],
    "estados": ["nuevo", "triangulado", "resuelto", "cerrado"],
    "categorias": {
      "nuevo": "propuesto",
      "triangulado": "en-progreso",
      "resuelto": "en-progreso",
      "cerrado": "completado"
    },
    "transiciones": [
      { "de": "nuevo", "a": "triangulado" },
      { "de": "triangulado", "a": "resuelto" },
      { "de": "resuelto", "a": "cerrado" }
    ]
  }
}
```

El schema es **`additionalProperties: false`**: una clave que no esté en el ejemplo lo
invalida. Los tres errores que se cometen solos:

- **`fases` es una lista de STRINGS**, no de objetos. Ídem `estados`, `terminales` y
  `empresas`. El nombre legible de una fase no va en el manifiesto.
- **Una transición tiene SOLO `de` y `a`.** Nada de `fase`, `caja` ni `nombre` adentro —
  el cableado caja↔transición vive en el contrato de la caja, no acá.
- **`marketplace` y `canal` son strings, y se OMITEN si el arnés no se publica.** Un
  `{"publicado": false}` no es «no publicar»: es una clave inválida.

Requeridos: `id`, `rol`, `proceso`, `reporta_a` (usá `null` si es raíz — omitirlo NO vale).
El `spine` exige `inicial` y `estados`. Las `categorias` salen de un enum fijo de 5:
`propuesto · en-progreso · completado · descartado · pausado`.

Reglas duras del spine:

- Todo estado de `transiciones` existe en `estados`.
- Los `terminales` tienen categoría `completado` o `descartado`.
- Cada transición que una caja va a poseer aparece acá ANTES de escribir la caja.

Al terminar, **releé el archivo que escribiste** y comprobá clave por clave contra el
ejemplo de arriba. Un manifiesto que no valida no es un arnés: es un directorio con un JSON.

## Paso 4 — las bandas transversales

Un arnés nace en **forma-plugin**, que es como se distribuye por marketplace. El layout,
completo:

```
arnes.l0.json            el manifiesto (paso 3)
.claude-plugin/plugin.json
skills/<id>/SKILL.md     una caja por transición
hooks/hooks.json         la Guardia
CLAUDE.md                la Base
```

**No escribas en `.claude/`.** Esa es la forma «arnés instalado», que es OTRA cosa (la
nomenclatura reconoce las dos, y si están ambas manda el plugin); además Claude Code
protege esa ruta y te va a pedir autorización a mitad de la forja. La Guardia va en
`hooks/hooks.json`, no en `.claude/settings.json`.

Guardia y Base son bandas, **NO cajas** (A1): no poseen transición ni gate.

- **Guardia** (`hooks/hooks.json`): incluí `telemetry-emit` — la telemetría de nacimiento
  es principio 9, nace con el arnés y no es opt-in. Toda regla dura de la Base que sea
  destructiva necesita su `PreToolUse` que la enforce; sin él la regla es solo prosa, y eso
  es un hallazgo. Bloquear es **exit 2** (exit 1 no bloquea nada).
- **Base** (`CLAUDE.md` del arnés): hechos always-on, bajo el techo de contexto que fija
  `knowhow/rules.md` (~200 líneas). Lo que sea procedimiento multi-paso NO va acá: va a una
  skill.

## Paso 5 — las cajas, fase por fase

Una caja por transición del spine. Para cada una invocá **`forjar-caja`**, que ya conoce el
contrato fusionado y sus checks. No dupliques su trabajo acá.

Antes de forjarlas, decidí cuáles NO son cajas: si un tramo es puro juicio sin
descomposición contratable, se registra como `no-arnesar` con su motivo y se deja fuera.
Un arnés honesto tiene huecos declarados; uno inflado tiene cajas que nadie corre.

## Paso 6 — verificar y declarar los grises

Corré `arnesia conformance --arnes <graph>` si hay grafo exportado. Cerrá informando:

- qué nodos del knowhow abriste (paso 1);
- qué checks quedan GRISES y por qué — un check sin evidencia se declara, no se pinta verde;
- qué quedó como `no-arnesar`.

## Prohibido

- Escribir el manifiesto sin haber leído `knowhow/harness-profile.md` y `rules.md`.
- Inventar estados, o forjar una caja cuya transición no está en el spine.
- Frontmatter fantasma (`persistent_facts` / `activation_steps_prepend` / `customize` /
  `sanctum`): Claude Code los ignora en silencio, así que declararlos es un hallazgo.
- Copiar el kit, esta doctrina o el knowhow DENTRO del arnés (② no contamina ③). La zona
  legal de escritura del proceso es `.arnesia/` del proyecto.
- Dar por firmado un gate humano. Si no hay eval ejecutable, `gate.tipo: none` con el
  motivo.
