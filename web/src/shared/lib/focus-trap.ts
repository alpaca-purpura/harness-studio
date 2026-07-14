import type { KeyboardEvent as ReactKeyboardEvent } from "react"

// Primitivo de a11y para dialogs modales propios (Portafolio T5 drawer + T6 wizard, G8).
//
// Por qué NO usamos `@base-ui-components/react/dialog` (YA es dependencia — S1-D18):
// `Dialog.Popup` exige un `<Dialog.Portal>` ancestro (tira si falta) y ese portal, por
// default, monta el contenido en `document.body` — sacándolo del subárbol que
// `within(canvasElement)` recorre en TODAS las stories de este repo (patrón fe-visual-fitness,
// ver portafolio-list.stories.tsx). Pasar un `container` custom exigiría colar una prop de
// "dónde portalar" a través del contrato puro `PortafolioDrawerProps` (que no la tiene y no
// debería tenerla — es un detalle de test, no del dominio). Este helper reimplementa SOLO lo
// que un dialog modal necesita (foco inicial + trap de Tab + Esc-cierra), sin portal, para que
// el contenido siga viviendo inline en el árbol — testeable con `within(canvasElement)` igual
// que el resto del Storybook.

const FOCUSABLE_SELECTOR = [
  "a[href]",
  "button:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  '[tabindex]:not([tabindex="-1"])',
].join(",")

/** Los elementos operables por teclado dentro de `container`, en orden del DOM. */
export function focusablesEn(container: HTMLElement): HTMLElement[] {
  return Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR))
}

// trapTabKeyDown — handler de `onKeyDown` (React, nativo del div `role="dialog"`, sin
// listeners globales que sobrevivan al unmount): Tab en el último focuseable vuelve al
// primero; Shift+Tab en el primero va al último (G8). Se recalculan los focuseables en CADA
// keydown — así una sección que aparece/desaparece dentro del dialog (ej. el confirm interno
// de Desvincular) extiende o achica el loop sin lógica aparte.
export function trapTabKeyDown(container: HTMLElement, event: ReactKeyboardEvent): void {
  if (event.key !== "Tab") return
  const items = focusablesEn(container)
  if (items.length === 0) return
  const first = items[0] as HTMLElement
  const last = items[items.length - 1] as HTMLElement
  const activo = document.activeElement

  if (event.shiftKey && activo === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && activo === last) {
    event.preventDefault()
    first.focus()
  }
}
