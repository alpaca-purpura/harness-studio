// REST client for the ArnesIA daemon. The base URL points at the Go daemon (:4200 by
// default); override with VITE_ARNESIA_API. In the Tauri shell the daemon runs as a
// sidecar on the same host, so the default is correct there too.

import type { NewSession, Session } from "./types"

const BASE = import.meta.env.VITE_ARNESIA_API ?? "http://127.0.0.1:4200"

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  })
  if (!res.ok) {
    const body = await res.text().catch(() => "")
    throw new Error(`arnesia ${init?.method ?? "GET"} ${path}: ${res.status} ${body}`)
  }
  // 202 (turn accepted) and 204 (deleted) carry no body; empty text → undefined.
  const text = await res.text()
  return (text ? JSON.parse(text) : undefined) as T
}

export const api = {
  base: BASE,

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
}
