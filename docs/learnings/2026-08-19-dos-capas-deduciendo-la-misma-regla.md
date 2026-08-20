---
title: "Dos capas deduciendo la misma regla por su cuenta terminan discrepando"
date: 2026-08-19
type: arquitectura
promotable: si
applied: applied
tags: [contratos, frontend-backend, autoridad, cockpit]
---

# Dos capas deduciendo la misma regla

## El síntoma

Habilitamos las escrituras del board en el backend: `writableSistemas()`, el flag
`-allow-platform-writes`, los gates de `handlers_stories.go` pasando por ahí. Test verde.
`POST /api/transition` escribía en disco.

Y el board seguía bloqueado, con un candado y el cartel «CORE · solo lectura».

## La causa

El frontend tenía su propia deducción:

```tsx
if (isPlatform(sistema)) {
  toast('Platform es solo lectura · las transiciones las hace /pm.');
  return;
}
```

Dos capas respondiendo la misma pregunta —*¿se puede escribir acá?*— cada una con su
razonamiento. Mientras coincidieron, nadie notó que había dos. Al cambiar una, la otra
quedó mintiendo, y el usuario veía un candado **sin causa visible**: el servidor decía que
sí y la pantalla decía que no.

Lo mismo aparecía tres veces más: `RoadmapView`, `DriftView` y `MapView` repetían
`isPlatform(sistema)` para decidir si una vista aplicaba, en vez de preguntarle a
`viewAppliesTo()`.

## Qué cambió

`GET /api/sistemas` ahora devuelve `escribibles[]`. La UI **refleja** ese dato en vez de
deducirlo. Y los tres guards de vista pasaron a consultar la única función que sabe.

## La regla que queda

> Cuando dos capas necesitan la misma decisión, una la TOMA y la otra la RECIBE. La que la
> toma es la que tiene los datos para tomarla — acá el servidor, porque es quien tiene el
> flag. Duplicar el razonamiento no es redundancia defensiva: es garantizar que algún día
> discrepen, y que el usuario vea el desacuerdo antes que nosotros.

Señal de alarma concreta: el mismo predicado (`isPlatform(x)`, `esAdmin(u)`, `puedeEditar()`)
escrito de los dos lados de una API. Si aparece, uno de los dos sobra.
