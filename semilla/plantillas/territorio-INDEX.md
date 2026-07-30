---
arnes: {{.ArnesID}}
sembrado: "{{.Fecha}}"
territorio: {{.Territorio}}
---

# {{.Label}} — {{.ProyectoNombre}}

> Territorio `{{.Territorio}}` (definición). Aquí viven las dimensiones que el proyecto
> define para este territorio. Cada hoja llega en Fase 2; hasta entonces el hueco queda
> VISIBLE como `pendiente`.

## Dimensiones

| Dimensión | Zachman | Estado |
|---|---|---|
{{range .Dimensiones}}| {{.ID}} | {{.Zachman}} | pendiente |
{{end}}
Guía: una dimensión se puebla con DECISIONES durables que todo el proyecto comparte
(cara-NORMA); lo que se produce para un paquete concreto va al WIP, no aquí.
