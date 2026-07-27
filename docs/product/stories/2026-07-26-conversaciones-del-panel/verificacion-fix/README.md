# Verificación en pantalla — A-4 · A-5 · A-6

La auditoría declaró estos tres hallazgos como «derivados del código con certeza alta, **no
vistos en pantalla**» (§8.4: «no levanté el daemon ni el navegador»). Acá están vistos.

## Cómo se reprodujo

```bash
cd web && pnpm install --frozen-lockfile      # el symlink de node_modules NO sirve
npx storybook dev -p 6099 --no-open --quiet &
node docs/product/stories/2026-07-26-conversaciones-del-panel/verificacion-fix/verificar-teclado.mjs
```

Chromium headless (`web/node_modules/playwright`), `waitUntil: 'load'` — nunca `networkidle`:
el SPA mantiene el SSE abierto y `networkidle` no llega nunca.

## Resultado — **12/12**, en los dos temas

| # | qué se verificó | light | dark |
|---|---|---|---|
| A-4 | el foco aterriza solo en la **lista** tras el frame de carga (entrada por `▶`) | ✅ | ✅ |
| A-4 | `↓` mueve el cursor **sin ningún `focus()` de cortesía del test** | ✅ | ✅ |
| A-4 | el foco aterriza solo en el **buscador** (entrada por `🔍`) | ✅ | ✅ |
| A-4 | lo que se teclea **llega** al buscador sin tabular | ✅ | ✅ |
| A-5 | 20 `↓` sobre 50 conversaciones: `scrollTop` 710 → 1648 y el cursor queda **dentro** del viewport | ✅ | ✅ |
| A-6 | `Escape` desde la **lista** llega al contenedor y cierra | ✅ | ✅ |

## Dos precisiones honestas sobre el método

1. **Storybook corre el `play` de cada story antes de que el script toque nada.** Por eso los
   valores de partida no son los del montaje (el cursor ya está en la 2.ª fila, el `scrollTop`
   ya es 710). Las aserciones miden el **movimiento** desde donde esté, no un id fijo: medir un
   id fijo daba un falso rojo, y ese falso rojo fue mío, no del producto.
2. **El `value` del buscador no se actualiza al teclear** y eso es correcto: es un input
   controlado cuyo `onBusqueda` es un spy (`fn()`). Lo que se verifica es que la tecla
   **aterriza** en el input con el foco puesto, que es exactamente lo que A-4 rompía.

## Lo que estas capturas NO prueban

- **No hay color nuevo que medir.** Los tres arreglos son de comportamiento: ni una clase ni un
  token cambiaron, así que no hay par fg/bg nuevo que someter al gate de contraste.
- **No se corrió un lector de pantalla.** Igual que la auditoría (§8.3), el marcado ARIA se leyó
  pero no se escuchó.
- **No se levantó el daemon.** Esto es el componente en Storybook, que es el SSoT del UI del
  repo — no la app instalada. El gate de T6 (CI) sigue abierto.
