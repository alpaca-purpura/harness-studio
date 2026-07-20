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

import type { NewSession, Session } from "./types"

const BASE = import.meta.env.VITE_ARNESIA_API ?? "http://127.0.0.1:4200"

// authToken is the capability token attached to every request (undefined in dev).
let authToken: string | undefined

// ApiError carries the HTTP status so callers can branch (e.g. 409 = session busy).
export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
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
    )
  }
  return res.text()
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

  listSessions: () => req<Session[]>("/api/sessions"),

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
  // ttl_segundos ACOTA el TTL del grant del rol (1 = «permitir una vez»).
  resolvePermission: <T = unknown>(
    id: string,
    requestId: string,
    decision: "allow" | "deny",
    ttlSegundos?: number,
  ) =>
    req<T>(`/api/sessions/${id}/permission`, {
      method: "POST",
      body: JSON.stringify({
        request_id: requestId,
        decision,
        ...(ttlSegundos ? { ttl_segundos: ttlSegundos } : {}),
      }),
    }),

  // interrupt (RF-116) — Stop real: corta el turno en vuelo in-band; el cierre llega
  // como frame `result` por SSE.
  interrupt: (id: string) => req<void>(`/api/sessions/${id}/interrupt`, { method: "POST" }),

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
