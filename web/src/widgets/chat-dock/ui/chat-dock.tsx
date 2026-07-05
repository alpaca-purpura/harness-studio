import { useEffect, useRef, useState } from "react"
import type { Session, Turn } from "@/shared"
import { selectActive, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { Pip } from "@/shared/ui/indicators"

// ChatDock is the invoked, collapsible conversation (mockup it.14): the live Claude
// Code session of the active work-front. Sending a turn streams the reply here.
export function ChatDock() {
  const active = useSessions(selectActive)
  const streaming = useSessions((s) => (active ? s.streaming[active.id] : undefined))
  const closeChat = useSessions((s) => s.closeChat)

  if (!active) return null

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex flex-none items-center gap-2 border-b border-border px-3 py-2.5">
        <span className="flex min-w-0 items-center gap-1.5 text-xs font-semibold">
          <Pip status={active.status} />
          <span className="truncate">{active.frente}</span>
        </span>
        <button
          type="button"
          onClick={closeChat}
          title="Colapsar el dock (⌘K para reabrir)"
          className="ml-auto flex items-center gap-1.5 whitespace-nowrap rounded-md px-2 py-1 text-[11px] text-muted-foreground hover:bg-secondary hover:text-foreground"
        >
          ⟩ colapsar
        </button>
      </div>

      <SessionLine session={active} />
      <Messages session={active} streaming={streaming} />
      <Composer />
    </div>
  )
}

function SessionLine({ session: s }: { session: Session }) {
  return (
    <div className="flex flex-none flex-wrap items-center gap-1.5 border-b border-border bg-secondary px-3 py-1.5 font-mono text-[9.5px] text-muted-foreground">
      <span className="text-primary">
        ◍ {s.claude_session_id ? s.claude_session_id.slice(0, 8) : "sin sesión CC"}
      </span>
      <span>· {s.arnes}</span>
      {s.model && <span>· {s.model}</span>}
      <span>· ctx</span>
      <span className="h-[5px] w-11 overflow-hidden rounded-full border border-border bg-card">
        <span className="block h-full bg-primary" style={{ width: `${s.ctx_pct ?? 0}%` }} />
      </span>
      <span>{s.ctx_pct ?? 0}%</span>
    </div>
  )
}

function Messages({ session: s, streaming }: { session: Session; streaming: string | undefined }) {
  const endRef = useRef<HTMLDivElement>(null)
  const conv = s.conv ?? []
  const showLive = s.status === "streaming"

  // biome-ignore lint/correctness/useExhaustiveDependencies: these deps are intentional scroll triggers; the body only reads the ref.
  useEffect(() => {
    endRef.current?.scrollIntoView({ block: "end" })
  }, [conv.length, streaming, showLive])

  return (
    <div className="flex flex-1 flex-col gap-2 overflow-auto p-3">
      {conv.length === 0 && !showLive && (
        <p className="m-auto max-w-[85%] text-center text-[11px] text-muted-foreground">
          Pídele un cambio a <b className="text-foreground">{s.arnes}</b>. Esta conversación ES la
          sesión Claude Code del frente «{s.frente}».
        </p>
      )}
      {conv.map((t, i) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: conv is append-only and immutable, so the index is a stable identity.
        <Bubble key={i} turn={t} />
      ))}
      {showLive && (
        <div className="max-w-[92%] self-start rounded-[10px] rounded-bl-[3px] border border-border bg-secondary px-2.5 py-2 text-xs leading-relaxed">
          {streaming ? (
            <span className="whitespace-pre-wrap">{streaming}</span>
          ) : (
            <span className="typing">
              <i />
              <i />
              <i />
            </span>
          )}
        </div>
      )}
      <div ref={endRef} />
    </div>
  )
}

function Bubble({ turn: t }: { turn: Turn }) {
  if (t.rol === "sys") {
    return (
      <div className="self-center rounded-md border border-dashed border-input bg-secondary px-2 py-0.5 font-mono text-[9.5px] text-muted-foreground">
        {t.text}
      </div>
    )
  }
  return (
    <div
      className={cn(
        "max-w-[92%] whitespace-pre-wrap rounded-[10px] border px-2.5 py-2 text-xs leading-relaxed",
        t.rol === "user"
          ? "self-end rounded-br-[3px] border-primary bg-accent-soft"
          : "self-start rounded-bl-[3px] border-border bg-secondary",
      )}
    >
      {t.text}
    </div>
  )
}

function Composer() {
  const sendTurn = useSessions((s) => s.sendTurn)
  const active = useSessions(selectActive)
  const [value, setValue] = useState("")
  const busy = active?.status === "streaming"

  const submit = () => {
    const text = value.trim()
    if (!text || busy) return
    setValue("")
    void sendTurn(text)
  }

  return (
    <div className="flex flex-none items-end gap-2 border-t border-border p-3">
      <textarea
        value={value}
        rows={1}
        placeholder={busy ? "generando…" : `Pídele un cambio a ${active?.arnes ?? "…"}`}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault()
            submit()
          }
        }}
        className="max-h-32 min-h-[36px] flex-1 resize-none rounded-lg border border-border bg-secondary px-2.5 py-2 text-xs text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none"
      />
      <button
        type="button"
        onClick={submit}
        disabled={busy || !value.trim()}
        title="Enviar (Claude Code headless detrás)"
        className="grid size-9 flex-none place-items-center rounded-lg bg-primary text-sm text-primary-foreground disabled:opacity-40"
      >
        ↑
      </button>
    </div>
  )
}
