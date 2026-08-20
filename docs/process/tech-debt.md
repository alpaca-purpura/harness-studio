# Deuda técnica — carril L3 del CIL (código e infraestructura del producto)

> **Qué entra acá:** deuda del PRODUCTO — código, runtime, arquitectura, peso del binario.
> La fricción del proceso y las herramientas va a
> [`harness-backlog.md`](./harness-backlog.md), carril L1.
>
> **De dónde sale:** estos ítems ya vivían en
> [`docs/product/BACKLOG.md`](../product/BACKLOG.md) § Deuda viva, en prosa. Acá se
> tabulan para que el cockpit los pueda contar y pintar; el BACKLOG sigue siendo donde se
> explica cada uno en detalle, y este archivo apunta ahí. **No se duplicó la explicación:
> se indexó.**
>
> **Severidad:** 🔴 bloquea-pronto · 🟡 friccion · 🔵 mejora.
> **Estado:** `reported` → `triaged` → `ratified` → `applied` → `verified` · `deferred`.

| id | fecha | sev | item | estado | ref |
|---|---|---|---|---|---|
| TD-1 | 2026-08-19 | 🔴 | El daemon está al 94,5 % del techo absoluto de peso (23,62 de 25 MB). `text/template`, linkeado por el forjador de la semilla, costó +3,28 MB medidos. El headroom es ~1,4 MB: la próxima dependencia pesada lo revienta | triaged | `docs/product/BACKLOG.md` § Deuda viva · `arch/§11` |
| TD-2 | 2026-07-26 | 🔴 | N-21 · La marca de rotación se PIERDE EN SILENCIO con dos vistas abiertas: el frame solo se appendea si la copia local del transcript mide exactamente `TurnoIdx`, y un turno mandado desde otro cliente nunca entra en esa copia. La longitud queda corta para siempre. Viola el boundary `sesion-viva-consistente` | triaged | `stories/2026-07-26-conversaciones-del-panel/verificacion-e2e/INFORME.md` |
| TD-3 | 2026-07-26 | 🟡 | N-22 · RF-348 CA-1 sin construir: no hay marca «⟳ hilo reiniciado · checkpoint» cuando `tryHealResume` respawnea fresh. El literal tiene 0 ocurrencias en el árbol y el heal es «silent on success», así que el `cc-id` cambia por debajo sin que el operador se entere | reported | `internal/usecase/session_service.go` |
| TD-4 | 2026-07-26 | 🟡 | N-23 · El re-key de CV-D16 no corre solo ni avisa tras migrar: 3 de 5 sesiones del operador siguen invisibles hasta correr `arnesia sesiones recalibrar-llaves --aplicar` a mano, y nada se lo dice | reported | `cmd/arnesia/main.go` · `migracion.go` |
| TD-5 | 2026-07-26 | 🔵 | N-18 · El dry-run de CV-D16 contesta «no hay sesiones» si corre antes del primer arranque: abre solo el registro v2, que todavía no existe, y un registro ausente devuelve vacío en vez de decir «arrancá el daemon primero» | reported | `cmd/arnesia/sesiones.go` |
| TD-6 | 2026-07-26 | 🔵 | N-19 · El `＋` bloqueado no ofrece la salida: la spec fija «esperá a que termine el turno (■ para interrumpir)» y el código dice «…el turno en vuelo». Una línea de copy, esperando decisión del operador | reported | `web/src/widgets/chat-dock/ui/conversacion-row.tsx` |
| TD-7 | 2026-08-13 | 🟡 | `cargo build` produce un binario que apunta al `devUrl` y falla al conectar en la ventana; el válido lo da `tauri build`. Evaluar declarar el feature `custom-protocol` para que un `cargo build --release` suelto no engañe | reported | `web/src-tauri/Cargo.toml` |
| TD-8 | 2026-08-13 | 🔵 | Self-update en Windows: hoy es degradación honesta (staged `arnesia-nuevo.exe` + aviso accionable). Falta el rename-trick NTFS que lo haría in-place | deferred | `stories/2026-08-13-compilacion-windows/decisiones.md` § CW-D1 |
| TD-9 | 2026-07-30 | 🟡 | La SPA tiene `127.0.0.1:4200` hardcodeado: en otro puerto la UI entera dice «Failed to fetch» | reported | `docs/product/BACKLOG.md` § Deuda viva |
