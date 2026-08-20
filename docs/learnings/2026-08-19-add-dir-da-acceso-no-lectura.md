---
title: "--add-dir da acceso, no lectura: 121 checks viajaban inertes a cada sesión"
date: 2026-08-19
type: doctrina
promotable: si
applied: applied
tags: [inyeccion, knowhow, skills, progressive-disclosure]
---

# `--add-dir` da acceso, no lectura

## Qué creíamos

Que inyectar el árbol del estándar (`docs/architecture/knowledge/elements/` → `knowhow/`)
con `--add-dir` en cada spawn hacía que las sesiones trabajaran conforme a él. La cadena
estaba bien construida: embed → `~/.arnesia/knowhow/` → flag en el spawn. Los 12 nodos y
sus 138 checks llegaban a cada sesión.

## Qué pasaba en realidad

`--add-dir` **agrega el directorio a las rutas permitidas**. No carga nada en contexto.
El árbol llegaba y se quedaba ahí, esperando que alguien lo abriera. Y nadie lo abría:
`knowhow/` tenía cuatro menciones en todo el kit, todas citando `skills.md`. Los otros once
nodos —`hooks`, `rules`, `subagents`, `commands`, `mcp`, `plugins`,
`settings-permissions`, `output-styles`, `statusline`, `headless-sdk`, `harness-profile`—
no aparecían en ninguna skill. **121 de los 138 checks nunca llegaban al modelo.**

Peor: en `forjar-caja` el nodo se citaba en el paso 8 (verificar), cuando el SKILL.md ya se
había escrito en el paso 4. El estándar se consultaba *después* de haber decidido.

## Por qué es fácil de no ver

Porque todo lo verificable estaba verde. El provisioning tenía test, los flags tenían test
de fitness, el archivo existía en disco. Lo que faltaba no era un mecanismo sino una
**instrucción**, y ningún test mecánico pregunta «¿alguien lee esto?».

## Qué cambió

La lectura del nodo pasó a ser el paso 1 de las tres skills del kit, con obligación de
**declarar qué nodos se abrieron**. `doctrine.md` lo dice como regla dura: *una respuesta
que no abrió el nodo de la clase que está creando no cumple la doctrina*.

## La regla que queda

> Si el material tiene que influir en la decisión, alguien tiene que abrirlo **antes** de
> decidir, y tiene que poder decir que lo abrió. Un archivo accesible no es un archivo
> leído; y «lo tuve disponible» no es evidencia de nada.

Vale igual para `--add-dir`, para un `references/` de una skill y para cualquier
progressive disclosure: el nivel 2 no se carga solo.
