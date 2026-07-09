# Sonda: RF-162 post-settings (--setting-sources project,local)

argv: claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose --setting-sources project,local --plugin-dir /home/chalreme/.arnesia/kit --append-system-prompt-file /home/chalreme/.arnesia/doctrine.md --add-dir /home/chalreme/.arnesia/knowhow --mcp-config /home/chalreme/.arnesia/mcp.json --strict-mcp-config
cwd: /home/chalreme/Proyectos/harness-studio/dogfood/dev-full-cycle
exit code: 0

## Respuesta del modelo

Desde mi contexto actual, sin ejecutar nada. Marco en gris lo que **no puedo determinar** solo con lo que tengo cargado (fiel a la regla gris≠verde).

## Subagentes (Agent types)
Todos son del harness base de Claude Code — no infiero ninguno del kit:
- `claude` — catch-all, default de FleetView (tools: `*`)
- `Explore` — búsqueda read-only, fan-out (sin Edit/Write)
- `general-purpose` — investigación multi-paso (tools: `*`)
- `Plan` — diseño de planes de implementación (read-only)
- `statusline-setup` — configura la statusline (solo Read/Edit)

## Skills / comandos
En Claude Code skills y slash-commands comparten mecanismo; los agrupo por origen inferido.

**Kit propio del arnés** (namespace `arnesia-kit:` — la maquinaria ② inyectada por flags):
- `arnesia-kit:auditar-arnes` — auditar el arnés del cwd contra la doctrina
- `arnesia-kit:forjar-caja` — crear una caja nueva (skill + contrato fusionado)

**Anthropic-shipped (skills «de conocimiento/producto»):**
- `deep-research` · `dataviz` · `claude-api`

**Built-in del harness (workflow/config de Claude Code):**
- `code-review` · `simplify` · `review` · `security-review` — revisión de código/PR
- `verify` · `run` — ejecutar/validar la app
- `init` — generar CLAUDE.md
- `update-config` · `keybindings-help` · `fewer-permission-prompts` — configuración del harness
- `loop` · `schedule` — recurrencia / cron

> ⚠️ **Gris:** no puedo separar con certeza «shipped por Anthropic» de «plugin del operador» para varias de estas — la lista de skills no trae su origen. La única señal dura de origen es el namespace (`arnesia-kit:`); el resto lo infiero por función, no por dato.

## Servidores MCP
**Ninguno visible en mi contexto ahora mismo.** No hay herramientas con namespace `mcp__<server>__<tool>` cargadas ni diferidas. No puedo afirmar que no exista ninguno configurado sin leer settings/`.mcp.json` (que pediste no tocar) — solo que **en mi superficie de herramientas actual no hay ninguno**.

## Herramientas diferidas (contexto, no son skills/MCP)
Por completitud — están declaradas pero con schema sin cargar, todas del harness base (no MCP): `CronCreate/Delete/List`, `TaskCreate/Get/List/Output/Stop/Update`, `Monitor`, `SendMessage`, `EnterWorktree`/`ExitWorktree`, `WebFetch`, `WebSearch`, `NotebookEdit`, `DesignSync`, `PushNotification`, `RemoteTrigger`.

---
**Límite honesto:** «agrupados por origen si podés inferir» — solo el namespace `arnesia-kit:` es evidencia dura de procedencia. La partición Anthropic-shipped vs built-in vs plugin-del-operador para el resto es inferencia por función, no un hecho verificado en mi contexto.

## stderr

```
(vacío)
```

## Sonda dirigida — CLAUDE.md propio del arnés (E5, riesgo D3)

Pregunta puntual (`sonda-claudemd.mjs`): pedí cita textual de la regla `std-spec` de
`dogfood/dev-full-cycle/CLAUDE.md`. Respuesta — coincide carácter a carácter con el
archivo real:

> Todo `spec.md` de este arnés cumple:
> - Frontmatter con `status:` — document-as-cache...
> - Un `why` de una línea...
> - Capacidades `CAP-NN`...
> - `non_goals` explícitos...
> - Solo requisitos confirmados por el usuario en el grill...

**Confirma E5:** `--setting-sources project,local` NO se lleva el CLAUDE.md propio del cwd
del arnés — discovery mecanismo aparte, como documentaba la investigación.

## Auditoría de ancestros (residual documentado en D3)

`find` manual: `~/.arnesia/CLAUDE.md` → ninguno · `$HOME/CLAUDE.md` → ninguno ·
`$HOME/Proyectos/CLAUDE.md` → ninguno · `$HOME/Proyectos/harness-studio/CLAUDE.md` (raíz
del repo, un ANCESTRO de `dogfood/dev-full-cycle/`) → **existe**.

Sonda dirigida confirmó que SÍ se carga: preguntado si tenía cargado un CLAUDE.md que
mencione "ArnesIA"/"fábrica de arneses"/"VISION.md", respondió `Sí. "# ArnesIA — la
fábrica de arneses (crear · mapear · observar · mejorar)"` — la primera línea literal
del CLAUDE.md raíz del repo.

**Hallazgo residual (esperado por D3, NO bloqueante):** el walk ascendente de CLAUDE.md
es un mecanismo aparte de `--setting-sources` (E5) y sigue activo — en ESTE entorno de
dev el dogfood vive anidado dentro del propio repo `harness-studio`, así que su
CLAUDE.md raíz (el documento completo del producto) se cuela en cualquier spawn del
dogfood. En producción real un arnés del operador normalmente NO vive anidado bajo un
árbol con CLAUDE.md ajeno — pero si algún día lo está, este residual reaparece. Solo
`--bare` lo apaga del todo (descartado, D1). Sin fix en este paquete — documentado tal
como preveía D3, no cambia el veredicto de la fase (settings-sources SÍ cumple su
objetivo: cortar `enabledPlugins`/hooks personales del OPERADOR).