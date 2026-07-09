---
regla: superficie-local-confinada
version: 1.3
updated: 2026-07-08
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
  - url: https://code.claude.com/docs/en/mcp
    autoridad: oficial
    revisado: 2026-07-08
  - url: https://code.claude.com/docs/en/settings
    autoridad: oficial
    revisado: 2026-07-08
enforced_by:
  - fitness/arch_test.go:TestNoWildcardCORS
  - fitness/arch_test.go:TestLocalSurfaceConfined
  - fitness/arch_test.go:TestConfigSourceMCPAislado
  - fitness/arch_test.go:TestConfigSourceSettingSourcesExcludeUser
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
gates en orden; `/healthz` es exento de Host+token (liveness pura, sin datos ni efectos) — **pero
SIGUE reflejando CORS** cuando el `Origin` está en el allowlist (v1.2, HS-14): el WebView vive en
su propio origin (`tauri://localhost`), distinto de `127.0.0.1:<port>`, así que un `fetch()` sin
`Access-Control-Allow-Origin` en la respuesta es indistinguible de un daemon caído para el cliente,
aunque el daemon haya respondido 200 en el cable — regresión real, encontrada en producción por el
operador el mismo día del fix ②.

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
capability) y lo expone al WebView por `initialization_script` (`window.__ARNESIA_TOKEN__`, corre en
cada documento del WebView — **candidata #8 firmada HS-11**, reemplaza el `invoke('auth_token')` de
v1.0: ese comando exigiría abrir IPC a un origin remoto, ya que la SPA que se ve vive en el daemon,
no en un origin `tauri://`). El crédito fluye shell→daemon (env) y shell→WebView (script de init);
**el core sigue sin importar al shell**. En attach a un daemon preexistente (dev, o huérfano-con-
token) el shell no conoce el token → WebView tokenless bajo Host+Origin (la página de arranque
`conectando.html` sondea `/healthz` — el único endpoint exento de los 3 gates — para no quedar en
401 eterno esperando el resto de la API; **HS-14 fix ②**). El sidecar se mata al salir el shell
(evita un daemon huérfano con un token irreplicable).

**Eje config-source del spawn (v1.3, HS-17).** El confinamiento de arriba es sobre *quién
le habla al daemon*; este eje es sobre *qué carga a contexto la sesión que el daemon
spawnea* — mismo concern (superficie que ve una sesión ajena al operador), eje distinto
(config, no red). Diagnóstico HS-17: cada arnés spawneado heredaba `~/.claude/
settings.json` del OPERADOR (`enabledPlugins` personales — golang-*, stitch, caveman)
más el MCP de SU cuenta (conectores claude.ai), encima de su propio kit ①② — el mismo
costo de contexto sin enforcer que `knowledge/elements/mcp.md` L1.6 ya documentaba como
advisory (`mcp-toolsearch-off`/`mcp-unused`). `SpawnArgs`
(`internal/adapters/agent/claudecode/conductor.go`) ahora lo enforcea a nivel producto:
`--mcp-config <archivo agregado>`+`--strict-mcp-config` (siempre que `Injection` esté
poblada) cortan el MCP a SOLO lo que el `Provisioner` materializa en
`~/.arnesia/mcp.json`, y `--setting-sources project,local` (incondicional, sin campo en
`SpawnOpts` — nadie puede reintroducir `user`) corta `enabledPlugins`/hooks personales.
Ambos ortogonales a auth (`--bare` sigue descartado, D1 de HS-17). Verificado en vivo
contra el dogfood real (`historias/2026-07-08-aislamiento-config-cc/`): MCP de cuenta
6→0, skills ajenas ~50→0, kit propio y CLAUDE.md del arnés intactos.
`--safe-mode` se investigó como capa extra (RF-163) y se DESCARTÓ: mata las skills del
plugin propio (②) pese a `--plugin-dir` explícito, aunque el overlay de doctrina
(`--append-system-prompt-file`/`--add-dir`, ①) sí sobrevive — ver
`historias/2026-07-08-aislamiento-config-cc/decisiones.md` D4.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| host-gate | toda request (salvo /healthz) valida `Host` contra el allowlist loopback | error | banda Guardia «daemon alcanzable por rebinding DNS» | arch_test.go:TestLocalSurfaceConfined |
| origin-allowlist | un `Origin` de navegador debe estar en el allowlist; se refleja exacto, no `*` | error | «CORS abierto — web ajena conduce el agente» | arch_test.go:TestLocalSurfaceConfined |
| cero-wildcard-cors | ningún handler emite `Access-Control-Allow-Origin: *` | error | banda Guardia «ACAO wildcard en API que ejecuta agente» | arch_test.go:TestNoWildcardCORS |
| token-cuando-configurado | con token configurado, toda request lo exige (header o query), compare constante | error | «API sin token pese a estar configurado» | arch_test.go:TestLocalSurfaceConfined |
| sse-token-o-origin | el stream SSE acepta token por `?token=` (EventSource) y valida Origin | warn | «SSE sin confinamiento (headers imposibles)» | arch_test.go:TestLocalSurfaceConfined |
| shell-emite-token | el shell mint el token e inyecta por env al spawnear + `initialization_script`; nunca por args | info | «token en args (visible en ps) o ausente en prod» | web/src-tauri/src/lib.rs (revisión) |
| healthz-refleja-cors | `/healthz` refleja `Access-Control-Allow-Origin` para un Origin allowlisted pese a saltar Host+token | error | «WebView ve /healthz como daemon caído por CORS aunque responda 200» | arch_test.go:TestHealthzCORSReflected |
| mcp-config-siempre | todo spawn con `Injection` poblada agrega `--mcp-config`+`--strict-mcp-config` | error | «MCP de cuenta del operador colado en el arnés» | arch_test.go:TestConfigSourceMCPAislado |
| setting-sources-siempre | todo spawn agrega `--setting-sources project,local`, nunca `user` | error | «enabledPlugins/hooks personales del operador colados en el arnés» | arch_test.go:TestConfigSourceSettingSourcesExcludeUser |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-06) — hueco de seguridad que la auditoría del shell Tauri
  v1 destapó (ningún boundary cubría «quién le habla al daemon»). L1 = defense-in-depth local
  (Host+Origin+token). L2 = middleware `withAuth` en 3 gates, shell = raíz de confianza que mint e
  inyecta el token (coherente con core⊥shell), degradación dev bajo Host+Origin. **Nace `enforced`**
  (código + tests shippean juntos, a diferencia de los nodos fundacionales `proposed`). 6 checks.
- 2026-07-08 · v1.1 · HS-14: reparado drift de doc — `shell-emite-token` describía el `invoke('auth_token')`
  de v1.0, pero el código pasó a `initialization_script` desde HS-11 #8 (candidata «la UI vive en el
  daemon») sin sincronizar este nodo. Sin cambio de checks (siguen 6); solo texto L2 + checklist
  alineados con el código real. Documentado también el fix ② del diagnóstico HS-14: `conectando.html`
  sondeaba `/api/version` (SÍ exige token) en vez de `/healthz` (el único exento) — un daemon
  huérfano-con-token dejaba el probe en 401 eterno pese a que el gate `/healthz` siempre lo hubiera
  dejado pasar; corregido al endpoint que este boundary ya declaraba exento.
- 2026-07-08 · v1.2 · HS-14, REGRESIÓN REAL detectada por el operador el mismo día (instaló el fix
  ② y siguió viendo «el daemon no responde»): el early-return de `/healthz` saltaba TODO el
  middleware, incluida la reflexión CORS — un `fetch()` cross-origin del WebView (`tauri://localhost`
  ≠ `127.0.0.1:4200`) veía la respuesta 200 sin `Access-Control-Allow-Origin` como error de red, no
  como éxito. `curl` (sin enforcement CORS) no lo detectaba — por eso pasó la verificación previa.
  Fix: `/healthz` sigue exento de Host+token, pero refleja CORS si el `Origin` está en el allowlist.
  Check nuevo `healthz-refleja-cors`: test de regresión REAL (httptest) en
  `internal/adapters/transport/http/auth_test.go:TestHealthzReflectsCORSForAllowlistedOrigin` (corre
  siempre en `go test ./...`) + proxy source-scan en `arch_test.go:TestHealthzCORSReflected`
  (scoped a la rama `/healthz`, no todo el archivo) para que el motor `arnesia conformance` lo
  corra igual que sus hermanos de este nodo → 7 checks.
- 2026-07-08 · v1.3 · HS-17: suma el eje config-source del spawn (distinto de red/auth, mismo
  concern de superficie-que-ve-una-sesión-ajena). Diagnóstico: cada arnés spawneado heredaba
  `~/.claude/settings.json` del operador (`enabledPlugins` personales) + MCP de su cuenta,
  encima de su propio kit ①②. Fix: `--mcp-config`+`--strict-mcp-config` (RF-161, siempre que
  `Injection` esté poblada) + `--setting-sources project,local` incondicional (RF-162, sin
  campo configurable — nadie reintroduce `user`), ambos en
  `internal/adapters/agent/claudecode/conductor.go:SpawnArgs`. `--bare` sigue descartado (D1);
  `--safe-mode` se investigó (RF-163) y se DESCARTÓ (mata el plugin propio ② pese a
  `--plugin-dir` explícito). Verificado en vivo contra el dogfood real, sin regresión de
  `conformance --arnes`. → 9 checks.
