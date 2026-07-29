# Decisiones — Persistencia: correctness + escalabilidad (2026-07-28)

> Cada decisión conversada se escribe EN EL MISMO TURNO (§10).
> Detonante: auditoría propia de la capa de persistencia (SQLite `index.db`/`telemetria.db` +
> JSON `portafolio.json`/`marketplaces.json`/`sessions.json`/`arneses.json`) a pedido del operador
> ("actuando como experto backend y en data, auditando... ¿es correcto? ¿debemos mejorar algo?"),
> seguida de "dime cuál sería tu plan para que todo esté arquitectónicamente correcto y me permita
> escalabilidad", y el `/goal` de implementar las fases 1 a 1 con prueba real contra el binario
> instalado.

## D1 · Backup + cuarentena para `portafolio.json` y `marketplaces.json` — ELEGIDA

`sessions.json` ya tiene respaldo (`respaldar()`) + cuarentena (`encuarentenar()`) en
`internal/adapters/store/migracion.go`. `portafolio.json` y `marketplaces.json` NO — pese a que la
doctrina llama a `portafolio.json` "identidad a largo plazo, la más irrecuperable". Una fila
top-level ilegible hoy se descarta en el siguiente write sin dejar rastro.

⇒ Clonar el mismo patrón (mismo nombre de archivo `.corrupto-<sello>` / `.v<N>-<sello>.bak`) hacia
los dos stores. No es diseño nuevo: es paridad con lo que ya existe y ya está probado en producción.

## D2 · Fase 1 revisada: flock cross-proceso, NO proxy HTTP al daemon — DECIDIDA (revisión de plan)

El plan original (propuesto al operador) decía "CLI habla HTTP al daemon si está corriendo". Al leer
el código real (`internal/adapters/transport/http/auth.go`, `router.go`) esto exige exponerle al CLI
un token capaz de escribir en `/api/portafolio/*` — y el único candidato existente,
`FichaDaemon.TokenIngesta`, está **deliberadamente acotado** (decisión A6: "nunca el token de la
API... eso concede conducir un agente con acceso al filesystem"). Ensancharlo, o mintear un tercer
token de escritura, es un cambio de superficie de seguridad que el operador no pidió y que no hace
falta para la garantía que realmente se busca.

La garantía real que hace falta es: **ninguna escritura se pierde cuando el daemon y un CLI
standalone mutan el mismo archivo a la vez.** Eso se resuelve con un lock de archivo cross-proceso
(flock) que envuelve el ciclo completo *recargar-desde-disco → mutar → guardar* — sin tocar el
perímetro HTTP ni su modelo de tokens.

⇒ Paquete nuevo `internal/adapters/filelock` (build tags unix/windows sobre `golang.org/x/sys`, ya
vendorizado transitivamente por `modernc.org/sqlite` — no se agrega ninguna dependencia nueva al
árbol de supply-chain). `Guard(path string, fn func() error) error` toma un lock exclusivo sobre
`<path>.lock` (sidecar, nunca el archivo de datos: el rename atómico no se toca) y corre `fn()`
adentro.

## D3 · Alcance del flock: 3 stores auto-contenidos SÍ, `sessions.json` NO (guard distinto) — DECIDIDA

Confirmado leyendo `cmd/arnesia/main.go`: `runPortafolio` (línea 778) construye su PROPIA instancia
de `portafolio.Store` con `newPortafolioService(nil)`, independiente de la que vive dentro de
`runServe` (línea 317) — el daemon corriendo y `arnesia portafolio agregar` en otra terminal son dos
cachés en memoria separadas escribiendo el mismo archivo. Mismo riesgo, confirmado, para
`arnesia sesiones recalibrar-llaves --aplicar` (`cmd/arnesia/sesiones.go`) contra `sessions.json`.
`marketplaces.json` hoy SOLO lo escribe el daemon (`newMarketplaceService` se llama una sola vez,
línea 324) — sin CLI propio, sin riesgo confirmado hoy, pero mismo patrón hermano ("calco exacto de
portafolio/store.go", comentario propio del archivo).

- `portafolio.Store`, `marketplace.Store`, `store.ArnesRegistry`: los tres cargan el archivo entero
  a un mapa en memoria, mutan, reescriben — mismo shape. Reciben el flock + un `load()` que se puede
  invocar de nuevo (resetea el mapa) DENTRO del lock, así la sección crítica siempre parte de lo que
  hay en disco en ESE instante, nunca de un caché que pudo quedar viejo.
- `store.Registry` (sesiones) es distinto A PROPÓSITO: `Load`/`Save` no cachean nada — el caché vive
  en el usecase del motor de chat vivo (`SessionService`), que es el componente más cargado de
  estado y el que más se tocó en HS-26/paneles de conversaciones. Tocar su contrato de caché para
  meterle un reload-under-lock es cirugía de alto riesgo sobre el camino más caliente del sistema,
  para cerrar un hueco que en la práctica solo abre un comando de mantenimiento explícito
  (`recalibrar-llaves`, no algo que un usuario corra sin querer).
  ⇒ En su lugar: `cmdSesiones` consulta la ficha del daemon (`descubrimiento.Ficha`, YA EXISTE,
  usada hoy por el hook de telemetría) ANTES de `--aplicar`; si el daemon está vivo, rechaza con un
  mensaje claro salvo que se pase `--force`. Cierra el riesgo real sin tocar el motor vivo.

## D4 · Fase 2: visibilidad de drift portafolio↔índice — DECIDIDA

`ObservarEnMapa` es HTTP-only por diseño (S1-D1: "el scan del Portafolio NUNCA registra en
`arneses.json`"), decisión correcta — pero hoy no hay señal de CUÁNTOS proyectos del Portafolio
nunca pasaron por ahí. Se agrega un chequeo ("N proyectos en portafolio sin cobertura en el índice")
al mismo mecanismo de honestidad que ya existe (`estado.sh` / drift-gate HS-21), no uno nuevo.

## D5 · Mecanismo de prueba promovido a skill, DESPUÉS de validarlo fase a fase — DECIDIDA

Por instrucción explícita del operador (`/goal`): antes de tocar código se armó y confirmó el
mecanismo de prueba (build sandboxed + HOME-copia del `~/.arnesia` real + verificación cruzada
HTTP/CLI + kill-por-ruta-exacta + concurrencia por re-exec de proceso real), y se usó sin cambios
en las 3 fases. Se promovió a `.claude/skills/verificando-binario-instalado/` (skill de APOYO,
`contract.caja: false` — no posee ninguna transición del proceso, es utilidad, per
`docs/architecture/knowledge/elements/skills.md` L2) recién DESPUÉS de la tercera corrida exitosa,
no antes: promover un mecanismo sin haberlo ejercitado sería documentar una intención, no un hecho.

Dos gotchas reales que el mecanismo atrapó en el camino, y que la skill documenta para no
repetirlos: (1) copiar `~/.arnesia` directo al HOME sandbox sin el subdirectorio `.arnesia/`
produce un store vacío silencioso, sin error; (2) un daemon sandbox de una corrida anterior que
queda vivo en el mismo puerto contesta curl con datos viejos sin señal de error — hay que matar
por ruta exacta y confirmar con `ps aux` ante la duda.

## D6 · Validación de cierre contra el binario REAL instalado (`make dev-sync`) — EJECUTADA

Tras las 3 fases en verde (`go test ./...` + fitness de arquitectura), se corrió `make dev-sync`
para reemplazar `~/.local/bin/arnesia` de verdad (el daemon no estaba corriendo — confirmado antes
de tocarlo). El primer arranque del binario nuevo contra el `~/.arnesia` REAL disparó, como es
esperable en cualquier primer boot de una versión nueva: la migración pendiente
`sessions.json→sesiones.json` (con su respaldo `.bak` automático, doctrina ya existente) + el
re-key CV-D18 (idempotente) + el aviso NUEVO de Fase 2 (3 de 4 entradas reales del Portafolio sin
observar en el Mapa — hallazgo real, no fabricado). Nada de esto es daño: es el mismo ciclo de
vida que el operador iba a disparar en su PRÓXIMO arranque normal de la app, solo que ocurrió un
rato antes, con las mismas garantías de respaldo. Se documenta acá para que quede trazable, no
porque haya algo que deshacer.

## Fuera de alcance (Fase 3, explícitamente NO ahora)

Migrar los JSON a SQL por "volumen": la cardinalidad real (arneses instalados por usuario) es
decenas, no millones — este es un daemon local mono-usuario, no un backend multi-tenant. Meter SQL
ahí sería complejidad sin beneficio hoy. Reconsiderar solo si aparece sync multi-máquina o
multi-usuario real en el roadmap.
