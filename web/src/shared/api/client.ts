// REST client for the ArnesIA daemon. The base URL points at the Go daemon (:4200 by
// default); override with VITE_ARNESIA_API. In the Tauri shell the daemon runs as a
// sidecar on the same host, so the default is correct there too.
//
// HS-06: the daemon confines its API (boundary superficie-local-confinada). When the Tauri
// shell minted a capability token, every request carries it (Authorization: Bearer). The
// shell injects the token as `window.__ARNESIA_TOKEN__` via initialization_script (it runs
// on every document, including the daemon-served SPA — candidata #8: the window navigates
// to :4200); in a plain browser the global is absent and the daemon falls back to
// Host+Origin only.

import type {
  Dictado,
  DisponibilidadDictado,
  EventoDiagnostico,
  NewSession,
  Session,
  Turn,
} from "./types"

const BASE = import.meta.env.VITE_ARNESIA_API ?? "http://127.0.0.1:4200"

// authToken is the capability token attached to every request (undefined in dev).
let authToken: string | undefined

// ApiError carries the HTTP status so callers can branch (e.g. 409 = session busy).
//
// `body` es el cuerpo CRUDO de la respuesta de error. Existe porque varios 4xx del daemon llevan
// datos estructurados que la UI necesita sin parsear prosa: el 409 de `POST /api/marketplaces`
// trae `{"error":…,"nombre":…}` para poder ofrecer «ir a él» (BR-7/E-18) y el 409 de
// `POST …/traidos` trae `{"error":…,"destino":…}` para «abrir el canónico que ya tenés»
// (BR-14/E-79). Se guarda tal cual (nunca se parsea acá: `shared/api` es domain-free).
export class ApiError extends Error {
  status: number
  body: string
  constructor(status: number, message: string, body = "") {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.body = body
  }
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  // 202 (turn accepted) and 204 (deleted) carry no body; empty text → undefined.
  const text = await reqText(path, init)
  return (text ? JSON.parse(text) : undefined) as T
}

// reqText — same auth/error discipline as req, but returns the raw body (text/plain
// endpoints like the node fuente, RF-93).
async function reqText(path: string, init?: RequestInit): Promise<string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" }
  if (authToken) headers["Authorization"] = `Bearer ${authToken}`
  const res = await fetch(`${BASE}${path}`, { ...init, headers: { ...headers, ...init?.headers } })
  if (!res.ok) {
    const body = await res.text().catch(() => "")
    throw new ApiError(
      res.status,
      `arnesia ${init?.method ?? "GET"} ${path}: ${res.status} ${body}`,
      body,
    )
  }
  return res.text()
}

// reqRaw posts a binary body (the dictation blob). No fija Content-Type a mano: lo aporta el
// Blob, que es quien sabe en qué formato grabó el motor.
async function reqRaw<T>(path: string, body: Blob, signal?: AbortSignal): Promise<T> {
  const headers: Record<string, string> = {}
  if (authToken) headers["Authorization"] = `Bearer ${authToken}`
  const init: RequestInit = { method: "POST", body, headers }
  // `signal` va condicional: con `exactOptionalPropertyTypes`, pasar `undefined` explícito no
  // es lo mismo que omitir la clave.
  if (signal) init.signal = signal
  const res = await fetch(`${BASE}${path}`, init)
  const text = await res.text().catch(() => "")
  if (!res.ok) {
    throw new ApiError(res.status, `arnesia POST ${path}: ${res.status} ${text}`, text)
  }
  return (text ? JSON.parse(text) : undefined) as T
}

export const api = {
  base: BASE,

  // setToken records the capability token; token() exposes it for the SSE URL (EventSource
  // cannot set headers, so the token rides as a query param there).
  setToken: (t: string | undefined) => {
    authToken = t
  },
  token: () => authToken,

  // Map / Portfolio / Inspector reads (S1–S3). shared/api stays domain-free (it must not
  // import entities/*, an upward dependency) — these are generic so the page (composition-root)
  // parameterizes them with the domain type it owns: `api.getGraph<Graph>(id)`.
  getGraph: <T = unknown>(id: string) => req<T>(`/api/harnesses/${encodeURIComponent(id)}/graph`),

  listHarnesses: <T = unknown>() => req<T>("/api/harnesses"),

  getNode: <T = unknown>(id: string, nodeId: string) =>
    req<T>(`/api/harnesses/${encodeURIComponent(id)}/nodes/${encodeURIComponent(nodeId)}`),

  // getConformance (S9) — la auditoría del grafo indexado (ConformanceReport). El caller
  // (la página) filtra por nodo con el selector de la entity (RF-91).
  getConformance: <T = unknown>(id: string) =>
    req<T>(`/api/harnesses/${encodeURIComponent(id)}/conformance`),

  // getNodeFuente (RF-93) — el archivo REAL del nodo, servido text/plain por el daemon,
  // confinado al dir registrado del arnés (S2). 404 honesto cuando no hay fuente/dir.
  getNodeFuente: (id: string, nodeId: string) =>
    reqText(`/api/harnesses/${encodeURIComponent(id)}/nodes/${encodeURIComponent(nodeId)}/fuente`),

  // Portafolio (S1, Slice 1 FE — decisiones.md §2.5): listar/escanear/agregar/desvincular +
  // observar-en-Mapa del `~/.arnesia/portafolio.json`. Genéricos <T> (domain-free, mismo
  // patrón que getGraph) — la entity entities/portafolio parametriza con sus tipos.
  listPortafolio: <T = unknown>() => req<T>("/api/portafolio"),

  escanearProyecto: <T = unknown>(path: string, signal?: AbortSignal) =>
    req<T>("/api/portafolio/escaneos", {
      method: "POST",
      body: JSON.stringify({ path }),
      ...(signal ? { signal } : {}),
    }),

  agregarProyecto: <T = unknown>(path: string, elegidos: string[]) =>
    req<T>("/api/portafolio/proyectos", {
      method: "POST",
      body: JSON.stringify({ path, elegidos }),
    }),

  desvincularDelPortafolio: <T = unknown>(clave: string) =>
    req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}`, { method: "DELETE" }),

  observarEnMapa: <T = unknown>(clave: string, installPath: string) =>
    req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/mapa`, {
      method: "POST",
      body: JSON.stringify({ install_path: installPath }),
    }),

  // identificarArnes escribe el sello arnes.l0.json in-situ y re-keya la entrada (S1-D28).
  // id/nombre opcionales — el backend cae al basename de la carpeta si van vacíos.
  identificarArnes: <T = unknown>(clave: string, installPath: string, id: string, nombre: string) =>
    req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/identificar`, {
      method: "POST",
      body: JSON.stringify({
        install_path: installPath,
        ...(id ? { id } : {}),
        ...(nombre ? { nombre } : {}),
      }),
    }),

  // Marketplaces (paquete 2026-07-23-portafolio-agregar-marketplace, design.md §8.6): el plano
  // (S2), el catálogo con refresco explícito (S3/S4), registrar (S6), la reconciliación de origen
  // (S7) y `↧ Traer canónico` (S8). Genéricos <T> domain-free, mismo patrón que `listPortafolio`
  // — `shared/api` NO puede importar `entities/*` (sería una dependencia hacia arriba).
  //
  // `ApiError.status` ya existe ⇒ el FE distingue el 409 de `registrarMarketplace` (ya
  // registrado, BR-7), el 503 de `validarMarketplace` («no puedo mirar», ≠ «tu url está mal») y
  // el 409/502/503 de `traerCanonico` **por status, no parseando el `message`**.
  listMarketplaces: <T = unknown>() => req<T>("/api/marketplaces"),

  validarMarketplace: <T = unknown>(url: string, signal?: AbortSignal) =>
    req<T>("/api/marketplaces/validaciones", {
      method: "POST",
      body: JSON.stringify({ url }),
      ...(signal ? { signal } : {}),
    }),

  registrarMarketplace: <T = unknown>(url: string, clase: string) =>
    req<T>("/api/marketplaces", {
      method: "POST",
      body: JSON.stringify({ url, clase }),
    }),

  olvidarMarketplace: <T = unknown>(nombre: string) =>
    req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}`, { method: "DELETE" }),

  // GET = caché primero (la primera visita hace UNA lectura y la persiste: idempotente).
  catalogoDeMarketplace: <T = unknown>(nombre: string, signal?: AbortSignal) =>
    req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}/catalogo`, {
      ...(signal ? { signal } : {}),
    }),

  // POST …/lecturas = refresco EXPLÍCITO (AG-D8 decisión 3). Es lo que pegan `Refrescar`,
  // `Reintentar` y `Leer catálogo`. Body vacío; misma respuesta que el GET.
  leerCatalogo: <T = unknown>(nombre: string, signal?: AbortSignal) =>
    req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}/lecturas`, {
      method: "POST",
      ...(signal ? { signal } : {}),
    }),

  candidatosDeOrigen: <T = unknown>(clave: string) =>
    req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/origen/candidatos`),

  // `home === null` ⇒ `{sin_origen:true}`: «ninguno — dejarlo sin origen» es una ELECCIÓN que se
  // registra, no la ausencia de una (E-26). No clona ni instala nada (BR-11).
  asignarOrigen: <T = unknown>(clave: string, home: string | null) =>
    req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/origen`, {
      method: "POST",
      body: JSON.stringify(home === null ? { sin_origen: true } : { home }),
    }),

  // traerCanonico (S8, AG-D17) — materializa la fila del catálogo como canónico editable bajo
  // `~/.arnesia/checkouts/`. Se manda la ENTRADA, no el `source`: el cliente no elige el
  // mecanismo (lo decide `PlanificarTraer` en el dominio). Sin `signal` a propósito: cerrar el
  // catálogo NO aborta un POST de materialización — abortar a mitad es peor que esperar (§13.10).
  traerCanonico: <T = unknown>(nombre: string, entrada: string) =>
    req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}/traidos`, {
      method: "POST",
      body: JSON.stringify({ entrada }),
    }),

  listSessions: () => req<Session[]>("/api/sessions"),

  // Historial B2 (RF-202/203): conversaciones de un arnés (vivas + cerradas, metadata) y
  // los turnos de una cerrada reconstruidos desde la JSONL nativa (faltantes = honestas).
  conversacionesDeArnes: (arnes: string) =>
    req<{ sesiones: Session[]; cerradas?: Session[]; cerradas_error?: string }>(
      `/api/sessions?arnes=${encodeURIComponent(arnes)}&cerradas=1`,
    ),
  historialCerrada: (id: string) =>
    req<{ turnos: Turn[]; faltantes?: string[] }>(
      `/api/sessions/cerradas/${encodeURIComponent(id)}/historial`,
    ),

  createSession: (input: NewSession) =>
    req<Session>("/api/sessions", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  closeSession: (id: string) => req<void>(`/api/sessions/${id}`, { method: "DELETE" }),

  renameSession: (id: string, frente: string) =>
    req<Session>(`/api/sessions/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ frente }),
    }),

  setView: (id: string, view: string) =>
    req<Session>(`/api/sessions/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ view }),
    }),

  turn: (id: string, text: string) =>
    req<void>(`/api/sessions/${id}/turn`, {
      method: "POST",
      body: JSON.stringify({ text }),
    }),

  // resolvePermission (RF-113/RF-114) — la decisión humana sobre una tarjeta `permission`.
  // Sin `role`: el daemon usa la autoridad del ARNÉS de la sesión (decisión #6).
  // ttl_segundos ACOTA el TTL del grant del rol (1 = «permitir una vez»). answers
  // (RF-113 bugfix) solo lo manda la tarjeta de AskUserQuestion — question text → respuesta.
  resolvePermission: <T = unknown>(
    id: string,
    requestId: string,
    decision: "allow" | "deny",
    ttlSegundos?: number,
    answers?: Record<string, string>,
  ) =>
    req<T>(`/api/sessions/${id}/permission`, {
      method: "POST",
      body: JSON.stringify({
        request_id: requestId,
        decision,
        ...(ttlSegundos ? { ttl_segundos: ttlSegundos } : {}),
        ...(answers ? { answers } : {}),
      }),
    }),

  // interrupt (RF-116) — Stop real: corta el turno en vuelo in-band; el cierre llega
  // como frame `result` por SSE.
  interrupt: (id: string) => req<void>(`/api/sessions/${id}/interrupt`, { method: "POST" }),

  // dictado (RF-222) — sube la grabación y devuelve el texto para el composer. El cuerpo es
  // el audio CRUDO, no JSON ni base64: es un blob de una pieza que va derecho al motor, y
  // envolverlo solo agregaría una copia y una decodificación. El Content-Type lo pone el
  // Blob (`audio/mp4` — lo único que graba WebKitGTK).
  //
  // NO lleva timeout propio: el tope de cada etapa vive en el daemon (RF-228) y duplicarlo
  // acá solo desincronizaría los dos. El cancel del operador va por AbortSignal.
  dictado: (id: string, audio: Blob, signal?: AbortSignal) =>
    reqRaw<Dictado>(`/api/sessions/${id}/dictado`, audio, signal),

  // disponibilidadDictado (RF-223/RF-227) — el composer la consulta al montar para saber si
  // ofrece el botón, y con qué motivo lo deshabilita si no.
  disponibilidadDictado: () => req<DisponibilidadDictado>("/api/dictado/disponibilidad"),

  // diagnostico (RF-230) — manda un fallo del FE al log del daemon.
  //
  // Existe porque el FE corre en un WebView SIN devtools y su `console.error` no va a ningún
  // lado: el bug del `MediaRecorder` mudo (RF-229) fue invisible durante toda una versión
  // instalada por exactamente eso. El daemon es el único proceso de la app que escribe a
  // disco, así que el detalle se le manda a él.
  diagnostico: (ev: EventoDiagnostico) =>
    req<void>("/api/diagnostico", { method: "POST", body: JSON.stringify(ev) }),

  // getVersion (RF-107) — identidad honesta del binario del daemon; también es el
  // polling del reinicio del self-update (RF-105): responde ⇔ el daemon está vivo.
  getVersion: <T = unknown>() => req<T>("/api/version"),

  // selfUpdate (RF-104) — POST largo (compila el repo local): fetch no impone timeout
  // FE y aquí NO se agrega ninguno; la respuesta ES el reporte de pasos. Cero
  // parámetros en el request (RF-106): repo y destino los conoce SOLO el daemon.
  selfUpdate: <T = unknown>() => req<T>("/api/self-update", { method: "POST" }),

  // configurarRepo (RF-109, bugfix fix-repo-self-update) — ÚNICO endpoint que acepta
  // un path: fija+persiste el repo del self-update cuando el daemon arrancó sin
  // --repo/ARNESIA_REPO (instalación empaquetada). selfUpdate arriba sigue sin params.
  configurarRepo: <T = unknown>(path: string) =>
    req<T>("/api/self-update/repo", { method: "PUT", body: JSON.stringify({ path }) }),

  // registerArnes sets the working directory an arnés's sessions run claude in (S2).
  registerArnes: (id: string, path: string) =>
    req<{ arnes: string; path: string }>(`/api/arneses/${id}`, {
      method: "PUT",
      body: JSON.stringify({ path }),
    }),

  // listArneses — the registered arnés→path entries (S2). The page uses it to know,
  // WITHOUT a doomed 404 round-trip, whether a fuente read can even be confined.
  listArneses: () => req<{ arnes: string; path: string }[]>("/api/arneses"),

  // ── Telemetría · capa «Mejora» (paquete 2026-07-24, T29) ────────────────────────────────
  //
  // Las 8 rutas de `/api/telemetria/*` (router.go:42-49). Genéricas <T> y domain-free, mismo
  // patrón que `listPortafolio`: `shared/api` no puede importar `entities/*`
  // (`shared-no-upward`), así que la página parametriza con los tipos de
  // `entities/telemetria`.
  //
  // ⚠️ La ventana viaja como `desde`/`hasta` en RFC3339, **no** como un enum `7d|30d|todo`:
  // así lo lee `ventanaDeQuery` (telemetria.go:25), y un valor ilegible da 400 en vez de
  // devolver en silencio una ventana distinta de la pedida. La traducción
  // `Ventana → desde/hasta` vive en la página, que es la dueña del estado.
  telemetriaResumen: <T = unknown>(q?: VentanaQuery) =>
    req<T>(`/api/telemetria/resumen${qsVentana(q)}`),

  telemetriaSalud: <T = unknown>() => req<T>("/api/telemetria/salud"),

  telemetriaPortafolio: <T = unknown>(q?: VentanaQuery) =>
    req<T>(`/api/telemetria/portafolio${qsVentana(q)}`),

  telemetriaCajas: <T = unknown>(clave: string, q?: VentanaQuery) =>
    req<T>(`/api/telemetria/arneses/${encodeURIComponent(clave)}/cajas${qsVentana(q)}`),

  telemetriaDetalleCaja: <T = unknown>(clave: string, cajaId: string, q?: VentanaQuery) =>
    req<T>(
      `/api/telemetria/arneses/${encodeURIComponent(clave)}/cajas/${encodeURIComponent(cajaId)}${qsVentana(q)}`,
    ),

  telemetriaMejoras: <T = unknown>(clave: string, q?: VentanaQuery) =>
    req<T>(`/api/telemetria/arneses/${encodeURIComponent(clave)}/mejoras${qsVentana(q)}`),

  // El borrado del operador (RF-275). No es idempotente-silencioso: el diálogo espera su
  // respuesta antes de decir «listo», y un fallo deja el dato intacto y lo dice.
  //
  // **Acepta ventana** (D26.5 · A-4). Sin ella borra TODO el historial; con ella borra
  // exactamente lo que la confirmación declaró. Es la única llamada irreversible del cliente,
  // así que la ventana se pasa explícita — nunca por default silencioso.
  telemetriaBorrarArnes: <T = unknown>(clave: string, q?: VentanaQuery) =>
    req<T>(`/api/telemetria/arneses/${encodeURIComponent(clave)}${qsVentana(q)}`, {
      method: "DELETE",
    }),

  // D26.4 — el descarte de un punto de mejora y su vuelta atrás. Hasta acá el botón nacía
  // deshabilitado porque **no existía endpoint** (A-1): decirlo era honesto, pero no era la
  // afordancia. Son dos llamadas y no un toggle: el servidor guarda una decisión, no un
  // estado que el cliente pueda dar vuelta por su cuenta.
  telemetriaDescartarPunto: <T = unknown>(clave: string, puntoId: string) =>
    req<T>(
      `/api/telemetria/arneses/${encodeURIComponent(clave)}/mejoras/${encodeURIComponent(puntoId)}/descartar`,
      { method: "POST" },
    ),

  telemetriaRecuperarPunto: <T = unknown>(clave: string, puntoId: string) =>
    req<T>(
      `/api/telemetria/arneses/${encodeURIComponent(clave)}/mejoras/${encodeURIComponent(puntoId)}/descartar`,
      { method: "DELETE" },
    ),

  telemetriaProceso: <T = unknown>(evento: unknown) =>
    req<T>("/api/telemetria/proceso", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(evento),
    }),
}

/** El rango de la consulta, ya resuelto a instantes por la página. `undefined` = todo. */
export interface VentanaQuery {
  desde?: string | undefined
  hasta?: string | undefined
  arnes?: string | undefined
  instalacion?: string | undefined
  limite?: number | undefined
}

function qsVentana(q: VentanaQuery | undefined): string {
  if (!q) return ""
  const p = new URLSearchParams()
  if (q.desde) p.set("desde", q.desde)
  if (q.hasta) p.set("hasta", q.hasta)
  if (q.arnes) p.set("arnes", q.arnes)
  if (q.instalacion) p.set("instalacion", q.instalacion)
  if (q.limite !== undefined) p.set("limite", String(q.limite))
  const s = p.toString()
  return s === "" ? "" : `?${s}`
}

// fetchAuthToken reads the capability token the Tauri shell injected as a global via
// initialization_script. Returns undefined in a plain browser (no shell → no global) so
// the app still works under the daemon's Host+Origin gate. Kept async: callers await it
// and the token source may become async again (e.g. a refresh handshake).
export async function fetchAuthToken(): Promise<string | undefined> {
  if (typeof window === "undefined") return undefined
  // biome-ignore lint/style/useNamingConvention: global inyectado por el shell (initialization_script) — el dunder marca que NO es código de la SPA
  const t = (window as { __ARNESIA_TOKEN__?: unknown }).__ARNESIA_TOKEN__
  return typeof t === "string" && t !== "" ? t : undefined
}
