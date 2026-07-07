// REST client for the ArnesIA daemon. The base URL points at the Go daemon (:4200 by
// default); override with VITE_ARNESIA_API. In the Tauri shell the daemon runs as a
// sidecar on the same host, so the default is correct there too.
//
// HS-06: the daemon confines its API (boundary superficie-local-confinada). When the Tauri
// shell minted a capability token, every request carries it (Authorization: Bearer). The
// token is fetched once on boot via `invoke('auth_token')` and set with setToken; in the dev
// browser (no Tauri) it stays undefined and the daemon falls back to Host+Origin only.

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

  // registerArnes sets the working directory an arnés's sessions run claude in (S2).
  registerArnes: (id: string, path: string) =>
    req<{ arnes: string; path: string }>(`/api/arneses/${id}`, {
      method: "PUT",
      body: JSON.stringify({ path }),
    }),
}

// fetchAuthToken asks the Tauri shell for the capability token. Returns undefined in the
// dev browser (no Tauri runtime) so the app still works under the daemon's Host+Origin gate.
export async function fetchAuthToken(): Promise<string | undefined> {
  if (typeof window === "undefined" || !("__TAURI_INTERNALS__" in window)) return undefined
  try {
    const { invoke } = await import("@tauri-apps/api/core")
    const t = await invoke<string>("auth_token")
    return t || undefined
  } catch {
    return undefined
  }
}
