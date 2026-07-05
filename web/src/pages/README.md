# pages/ (FSD)

**Composition-roots por hash-state** (app de escritorio SIN router, decisión HS-05). Cada `page`
compone `widgets` + `features` en respuesta a `useAppStore().view` (`#/mapa`, `#/portafolio`, …).
No hay `react-router`; la "ruta" es el fragmento de hash.

Vacío por ahora — se llena con el shell en la próxima sesión. Regla FSD: `pages` puede importar
`widgets/features/entities/shared`, nunca `app`.
