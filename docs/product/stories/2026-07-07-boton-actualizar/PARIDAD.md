# PARIDAD — mockup v2 ↔ app (se llena fila por fila al implementar)

> Contrato de «lo que ves en el mockup ES lo que hace la app». Gate final: todas las
> filas ✅ + el daemon actualizándose A SÍ MISMO de verdad (binario instalado en
> ~/.local/bin, huella nueva tras el reinicio), consola limpia, screenshots revisados.
> **Verificación 2026-07-07/08:** el daemon instalado (`~/.local/bin/arnesia`, huella
> `ed223a4`) se actualizó A SÍ MISMO a `0a42644` desde el botón — evidencia completa en
> [`validacion.md`](./validacion.md); screenshots en el scratchpad de la sesión
> (`shots/01-idle-light.png` … `08-sin-repo.png` + `mockup-caso-01..05.png`).

| RF | Elemento (mockup v2) | Componente real | Story/test | Estado |
|---|---|---|---|---|
| RF-100 | vista Ajustes nace con la tarjeta (223-247) | pages/shell/ui/global-view.tsx (AjustesView) | click-through: rail ⚙ → vista real, breadcrumb `alpacapurpura / Ajustes`, línea muted de futuras | ✅ |
| RF-101 | identidad honesta: huella·ruta·repo (159-171) | update-card + GET /api/version | story Idle + live: `ed223a4 · 2026-07-08 · ~/.local/bin/arnesia · pill ok · repo` (shots 01) | ✅ |
| RF-102 | no-escribible: disabled + migración (207-216) | update-card | story NoEscribible + live (chmod 555): pill warn + warnbox `install -m755` + disabled con title + POST→503 (shot 07) | ✅ |
| RF-103 | sin repo: disabled + honesto (211) | update-card | story SinRepo + live (daemon sin --repo): «repo no configurado» + disabled + POST→503 (shot 08) | ✅ |
| RF-104 | 5 pasos, corte al fallo, binario intacto (141-157·195-203) | ports.SelfUpdater + adapter + usecase + POST /api/self-update | go test -race ✓ + live: éxito 5 veredictos reales (shot 04) · build roto TS2322 → critbox con stderr, 3 pasos «no corrido», binario intacto `0a42644` (shot 06) · ya-al-día sin reinstalar (shot 05) | ✅ |
| RF-105 | reinicio + polling huella nueva (184-191) | update-card + re-exec daemon | stories Reiniciando/Exito + live: re-exec in situ (MISMO pid 930295), polling reconectó y confirmó `0a42644`, okbox «daemon reiniciado, UI reconectada» | ✅ |
| RF-106 | seguridad: withAuth · cero params · 409 | transport + usecase | go test (params ignorados·409·503) ✓ + live: Host/Origin ajenos → 403 · 2º POST en vuelo → 409 · lock retenido tras «actualizado» | ✅ |
| RF-107 | GET /api/version (buildinfo VCS) | Go + client.ts | go test (wire) ✓ + live: `{huella, fecha, instalado_en, escribible, repo, sucio}` reales del binario instalado | ✅ |

Leyenda: ⬜ pendiente · 🔶 implementado sin verificar · ✅ verificado en paridad.

## Desviaciones registradas (se consultan, jamás se maquillan)

1. **Driver de validación = Playwright headless (Chromium), no el MCP de Chrome DevTools.**
   El browser MCP quedó tomado (el WM maximiza la ventana y rompe `Browser.setContentsSize`
   en el handshake) — MISMO bloqueo y mismo fallback documentado del click-through del
   mockup v2. Todo lo demás del guion (screenshots, consola, asserts) se cumplió igual.
2. **Consola durante el update real: 2 líneas, ambas explicadas.** (a) un `409 (Conflict)`
   — lo produjo el PROPIO test del doble-POST (RF-106), es el veredicto esperado; (b) un
   `net::ERR_INCOMPLETE_CHUNKED_ENCODING` — el stream SSE `/events` muere cuando el daemon
   se re-ejecuta (inherente al reinicio; la UI reconecta). Todos los demás escenarios:
   consola **0 mensajes**.
3. **Pill no-escribible dice «root · actualizar pide sudo»** (texto fijo del spec RF-101 /
   mockup:211) aunque en el test la causa fue un chmod, no root. El caso real del operador
   ES root (/usr/bin vía .deb); la warnbox acompaña con la verdad genérica («sin permiso
   de escritura»).
4. **Futuras de Ajustes = línea muted** (spec RF-100 lo ordena: «texto, jamás tarjetas
   vacías fingiendo»); el caso 00 del mockup las dibujaba como tarjetas punteadas — el
   spec firmado supersede al frame ilustrativo.
5. **Ruta mostrada = `os.Executable()` completo** (`/home/chalreme/.local/bin/arnesia`);
   el mockup abreviaba `~/…`. Dato real sin abreviar.
6. **Checklist NO se anima durante «actualizando»** — cementado en spec RF-104/design
   (la animación del mockup 02 era demo): el estado muestra botón bloqueado + nota, y los
   veredictos aparecen SOLO al llegar la respuesta.

## Firma

- [x] 🧑‍⚖️ **Gate humano** — FIRMADO 2026-07-09 por orden del operador («firma todos los gates
  humanos y procede con los capabilities»). Las **6 desviaciones** quedan **aceptadas**: son verdad
  cruda, no maquillaje — driver Playwright (mismo fallback documentado del MCP tomado) · 2 líneas de
  consola ambas EXPLICADAS (el 409 lo produce el propio test del doble-POST · el ERR_INCOMPLETE del
  SSE es inherente al re-exec) · textos fijos del spec (root/sudo, ruta completa `os.Executable()`,
  futuras como línea muted, checklist sin animar) que el spec firmado supersede al frame ilustrativo.
  Respaldo IRREFUTABLE: **el daemon instalado se actualizó A SÍ MISMO** `ed223a4→0a42644`
  (`~/.local/bin/arnesia`, huella nueva tras reinicio in-situ, mismo pid, polling reconectó) —
  evidencia en `validacion.md` + `go test -race` ✓. Re-confirmado esta sesión: `go build/vet/test`
  ok · `conformance --todo 247·42·0·205`. Self-update vive en `internal/adapters/selfupdate` (Go),
  no tocado por el reorg de `docs/`. `chris_verify.signoff → true`.
