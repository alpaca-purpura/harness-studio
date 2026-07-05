---
regla: superficie-local-confinada
version: 1.0
updated: 2026-07-05
status: enforced
ledger: HS-06
sources:
  - url: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
    autoridad: experto
    revisado: 2026-07-05
  - url: https://v2.tauri.app/security/
    autoridad: oficial
    revisado: 2026-07-05
  - url: https://pkg.go.dev/crypto/subtle
    autoridad: oficial
    revisado: 2026-07-05
enforced_by:
  - fitness/arch_test.go:TestNoWildcardCORS
  - fitness/arch_test.go:TestLocalSurfaceConfined
severity: critical
---

# La superficie local del daemon está confinada y autenticada

## L1 · Principio (estándar de industria)

**Un daemon local que ejecuta un agente NO es una superficie de confianza por escuchar en
loopback.** Cualquier proceso de la máquina —incluida una página web abierta en el navegador del
usuario— alcanza `127.0.0.1`. Como un turno dirige a Claude Code con acceso al filesystem, una API
local abierta es efectivamente ejecución remota. La defensa es **defense-in-depth**, no una sola
barrera:

- **Host allowlist (anti DNS-rebinding).** Un sitio hostil puede apuntar un hostname a `127.0.0.1`;
  el navegador manda el `Host` del sitio, no `localhost`. Exigir que `Host` sea un loopback esperado
  corta el rebinding. *(experto: OWASP; principio DNS-rebinding)*
- **Origin allowlist en vez de CORS `*` (anti CSRF cross-site).** `Access-Control-Allow-Origin: *`
  deja que cualquier origen lea la respuesta; y aun sin leerla, un *simple request* ya ejecutó el
  efecto. Reflejar SOLO orígenes permitidos + `Vary: Origin` mata el CSRF de navegador. *(experto:
  OWASP CSRF)*
- **Token de capacidad (anti origin-spoofing local).** Un cliente NO-navegador local puede falsificar
  `Origin`. Un secreto que solo el shell de confianza conoce, comparado en tiempo constante, cierra
  esa vía. *(oficial: crypto/subtle; Tauri security — el shell firmado es la raíz de confianza)*

## L2 · Realización (este árbol Go+React)

El middleware `withAuth` (`internal/adapters/transport/http/auth.go`) envuelve TODO el mux en tres
gates en orden; `/healthz` es exento (liveness pura, sin datos ni efectos):

1. **Host gate (siempre).** `Host` ∈ {`127.0.0.1:<port>`, `localhost:<port>`, `[::1]:<port>`}
   derivados del `--addr`. ⇐ L1 rebinding.
2. **Origin gate (siempre, para navegadores).** Un `Origin` presente debe estar en el allowlist
   (WebView Tauri `tauri://localhost`/`http://tauri.localhost`, Vite dev `:5173`, el propio daemon);
   se **refleja ese origin exacto** con `Vary: Origin` — **jamás `*`**. El preflight OPTIONS se
   responde aquí, antes del token. ⇐ L1 CSRF.
3. **Token gate (si hay token configurado).** Con `ARNESIA_AUTH_TOKEN`/`--auth-token`, toda request
   trae el token: `Authorization: Bearer` o `X-Arnesia-Token` (fetch) o `?token=` (el `EventSource`
   SSE no puede poner headers). Compara con `crypto/subtle.ConstantTimeCompare`. Sin token (dev) el
   gate se salta y se loguea un warning; los gates 1+2 siguen. ⇐ L1 token.

**El shell es la raíz de confianza (coherente con [`core-no-importa-shell`](./core-no-importa-shell.md)).**
El shell Tauri (`web/src-tauri/src/lib.rs`) mint un token de 256 bits por lanzamiento, se lo pasa al
daemon **solo al spawnearlo** (env `ARNESIA_AUTH_TOKEN`, no por args → no toca el allowlist de la
capability) y lo expone al WebView por `invoke('auth_token')`. El crédito fluye shell→daemon (env) y
shell→WebView (comando); **el core sigue sin importar al shell**. En attach a un daemon preexistente
(dev) el shell no conoce el token → WebView tokenless bajo Host+Origin. El sidecar se mata al salir el
shell (evita un daemon huérfano con un token irreplicable).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| host-gate | toda request (salvo /healthz) valida `Host` contra el allowlist loopback | error | banda Guardia «daemon alcanzable por rebinding DNS» | arch_test.go:TestLocalSurfaceConfined |
| origin-allowlist | un `Origin` de navegador debe estar en el allowlist; se refleja exacto, no `*` | error | «CORS abierto — web ajena conduce el agente» | arch_test.go:TestLocalSurfaceConfined |
| cero-wildcard-cors | ningún handler emite `Access-Control-Allow-Origin: *` | error | banda Guardia «ACAO wildcard en API que ejecuta agente» | arch_test.go:TestNoWildcardCORS |
| token-cuando-configurado | con token configurado, toda request lo exige (header o query), compare constante | error | «API sin token pese a estar configurado» | arch_test.go:TestLocalSurfaceConfined |
| sse-token-o-origin | el stream SSE acepta token por `?token=` (EventSource) y valida Origin | warn | «SSE sin confinamiento (headers imposibles)» | arch_test.go:TestLocalSurfaceConfined |
| shell-emite-token | el shell mint el token e inyecta por env al spawnear + `invoke('auth_token')`; nunca por args | info | «token en args (visible en ps) o ausente en prod» | web/src-tauri/src/lib.rs (revisión) |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-06) — hueco de seguridad que la auditoría del shell Tauri
  v1 destapó (ningún boundary cubría «quién le habla al daemon»). L1 = defense-in-depth local
  (Host+Origin+token). L2 = middleware `withAuth` en 3 gates, shell = raíz de confianza que mint e
  inyecta el token (coherente con core⊥shell), degradación dev bajo Host+Origin. **Nace `enforced`**
  (código + tests shippean juntos, a diferencia de los nodos fundacionales `proposed`). 6 checks.
