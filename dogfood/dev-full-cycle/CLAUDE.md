# std-spec — estándar de spec

Regla de la banda Base del arnés `dev-full-cycle`: el estándar que la caja `spec-writer`
consume al emitir `spec.md` (su input `base:std-spec`).

Todo `spec.md` de este arnés cumple:

- Frontmatter con `status:` — document-as-cache: el estado del trabajo vive en el artefacto,
  no en la conversación.
- Un `why` de una línea: la intención inmutable de la unidad de trabajo.
- Capacidades `CAP-NN`, cada una con criterio de éxito verificable — sin capability sin `success`.
- `non_goals` explícitos: lo que la unidad NO hace.
- Solo requisitos confirmados por el usuario en el grill — nada inventado.
