---
title: "Un reader con defaults defensivos fabrica datos cuando el parseo falla en silencio"
date: 2026-08-19
type: tooling
promotable: si
applied: applied
tags: [honestidad, parseo, capabilities, gris-no-es-verde]
---

# Defaults defensivos + parseo que falla callado = datos inventados

## El caso

El reader de capabilities del cockpit hace lo que parece prudente:

```go
doc, parseErr := parseFrontmatter(string(raw))
if parseErr != nil { return nil, parseErr }
cap := map[string]any{
    "capability_id": fmGetOr(fm, "", "capability_id"),
    "status":        fmGetOr(fm, "live", "status"),
    …
}
```

Falla si el YAML es inválido, y si falta un campo pone un default razonable. Correcto en
apariencia.

Pero `parseFrontmatter` busca un bloque `---` … `---`. Tres capabilities de este repo son
**documentos YAML** que abren `---` y no cierran. Para esa función eso no es un error: es
«no hay frontmatter». Devuelve un mapa vacío **sin error**.

Entonces los defaults se aplican sobre la nada y sale una capability que no existe:
`capability_id: ""`, `status: "live"` — un status que ni siquiera pertenece al enum de este
repo (`vivo · vivo·nc · parcial · stub`). La API servía tres capabilities fantasma y nadie
lo notaba, porque *parecían* datos.

## Por qué nadie lo vio antes

`scripts/cap_doctor.py` valida las mismas 157 capabilities y las daba por buenas: su
`_load()` es tolerante (`text = body[:end] if end != -1 else body`), o sea que sí las leía
enteras. Doctor y cockpit **discrepaban en qué era el archivo**, y el que mentía era el que
nadie estaba mirando.

## La regla que queda

> Un default es honesto cuando rellena un campo AUSENTE. Es deshonesto cuando rellena un
> parseo FALLIDO, porque convierte «no pude leer esto» en «esto dice X». Inventar es peor
> que fallar: un error se ve, un dato falso se propaga.

Concretamente: si un parser puede devolver «vacío» por dos razones distintas —está vacío de
verdad, o no supe leerlo— esas dos ramas no pueden terminar en el mismo default.

Y el corolario: **cuando dos herramientas leen el mismo archivo, tienen que coincidir en
qué archivo es.** Si el doctor acepta dos formas, el visor también.
