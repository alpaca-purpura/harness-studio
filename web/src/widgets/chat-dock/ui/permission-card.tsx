import { useState } from "react"
import type { PermissionAsk } from "@/shared/api"
import { cn } from "@/shared/lib/cn"

// PermissionCard (RF-113, mockup-chat.html #permCard): la tarjeta inline de un
// control_request — herramienta + input legible + 3 acciones. La decisión viaja al
// daemon (POST /permission) y la tarjeta se cierra con el frame permission_result;
// «una vez» acota el grant a 1 s, «esta sesión» usa el TTL del rol (solo estrechable).
// AskUserQuestion (RF-113 bugfix) NO es un permiso — es una pregunta del modelo; se
// pinta como tarjeta de opciones, no como el genérico permitir/denegar (jamás JSON crudo
// con botones que no aplican).
export function PermissionCard({
  ask,
  onResolve,
}: {
  ask: PermissionAsk
  onResolve: (
    decision: "allow" | "deny",
    opts?: { once?: boolean; answers?: Record<string, string> },
  ) => void
}) {
  if (ask.tool === "AskUserQuestion") {
    const questions = parseAskUserQuestion(ask.input)
    if (questions) {
      return (
        <AskUserQuestionCard
          questions={questions}
          onAnswer={(answers) => onResolve("allow", { answers })}
          onDeny={() => onResolve("deny")}
        />
      )
    }
  }

  return (
    <div className="rounded-lg border border-warn bg-warn-soft text-xs">
      <div className="flex items-center gap-1.5 px-2.5 pt-2 font-semibold">
        <span aria-hidden>◐</span> Claude quiere usar <b className="font-mono">{ask.tool}</b>
      </div>
      <div className="px-2.5 py-1.5">
        <ToolInputPreview tool={ask.tool} input={ask.input} />
      </div>
      <p className="px-2.5 text-[10px] leading-snug text-muted-foreground">
        «Esta sesión» acuña un grant con el TTL del rol del arnés — mientras viva,{" "}
        <span className="font-mono">{ask.tool}</span> no re-pregunta. «Una vez» expira al instante.
        El deny del rol gana siempre al click.
      </p>
      <div className="flex flex-wrap gap-1.5 p-2.5">
        <button
          type="button"
          onClick={() => onResolve("allow")}
          className="rounded-md bg-primary px-2.5 py-1 font-semibold text-primary-foreground hover:brightness-110"
        >
          Permitir esta sesión
        </button>
        <button
          type="button"
          onClick={() => onResolve("allow", { once: true })}
          className="rounded-md border border-border bg-card px-2.5 py-1 hover:bg-secondary"
        >
          Permitir una vez
        </button>
        <button
          type="button"
          onClick={() => onResolve("deny")}
          className="rounded-md border border-border bg-card px-2.5 py-1 hover:border-destructive hover:text-destructive"
        >
          Denegar
        </button>
      </div>
    </div>
  )
}

// —— AskUserQuestion (RF-113 bugfix) ————————————————————————————————————————————

interface AskOption {
  label: string
  description?: string | undefined
}

interface AskQuestion {
  question: string
  header?: string | undefined
  options: AskOption[]
  multiSelect: boolean
}

// parseAskUserQuestion narrows el input crudo al shape esperado
// (`{questions:[{question,options:[{label,description}],header?,multiSelect?}]}`, el
// mismo zod schema de la herramienta). Honesto: si el shape no calza (versión futura de
// CC con campos distintos), devuelve undefined y PermissionCard cae al JSON crudo — nunca
// inventa opciones que no vinieron en el input.
function parseAskUserQuestion(input: unknown): AskQuestion[] | undefined {
  const qs = asRecord(input)["questions"]
  if (!Array.isArray(qs) || qs.length === 0) return undefined
  const out: AskQuestion[] = []
  for (const q of qs) {
    const qr = asRecord(q)
    const question = str(qr["question"])
    const rawOpts = qr["options"]
    if (question === undefined || !Array.isArray(rawOpts) || rawOpts.length === 0) return undefined
    const options: AskOption[] = []
    for (const o of rawOpts) {
      const or_ = asRecord(o)
      const label = str(or_["label"])
      if (label === undefined) return undefined
      options.push({ label, description: str(or_["description"]) })
    }
    out.push({
      question,
      header: str(qr["header"]),
      options,
      multiSelect: qr["multiSelect"] === true,
    })
  }
  return out
}

// AskUserQuestionCard pinta cada pregunta con sus opciones clicables + un campo «otra
// respuesta» (el propio schema dice que la opción "Other" la agrega el componente de
// permisos, no el modelo). El submit arma `answers: {pregunta: respuesta}` — question
// text → label(s) elegido(s) (multi-select: coma-separado) o el texto libre — que viaja
// como updatedInput del control_response allow (session_service.go
// askUserQuestionUpdatedInput).
function AskUserQuestionCard({
  questions,
  onAnswer,
  onDeny,
}: {
  questions: AskQuestion[]
  onAnswer: (answers: Record<string, string>) => void
  onDeny: () => void
}) {
  const [selected, setSelected] = useState<Record<string, Set<string>>>({})
  const [otro, setOtro] = useState<Record<string, string>>({})

  const toggle = (q: AskQuestion, label: string) => {
    setOtro((prev) => ({ ...prev, [q.question]: "" }))
    setSelected((prev) => {
      const cur = new Set(prev[q.question] ?? [])
      if (q.multiSelect) {
        if (cur.has(label)) cur.delete(label)
        else cur.add(label)
      } else {
        cur.clear()
        cur.add(label)
      }
      return { ...prev, [q.question]: cur }
    })
  }

  const respuestaDe = (q: AskQuestion) =>
    otro[q.question]?.trim() || [...(selected[q.question] ?? [])].join(", ")

  const listo = questions.every((q) => respuestaDe(q).length > 0)

  const enviar = () => {
    const answers: Record<string, string> = {}
    for (const q of questions) answers[q.question] = respuestaDe(q)
    onAnswer(answers)
  }

  return (
    <div className="rounded-lg border border-warn bg-warn-soft text-xs">
      <div className="flex items-center gap-1.5 px-2.5 pt-2 font-semibold">
        <span aria-hidden>◐</span> Claude necesita tu respuesta
      </div>
      <div className="flex flex-col gap-2.5 px-2.5 py-1.5">
        {questions.map((q) => (
          <div key={q.question} className="flex flex-col gap-1">
            {q.header && (
              <span className="w-fit rounded-full border border-border px-1.5 py-px text-[9px] uppercase text-muted-foreground">
                {q.header}
              </span>
            )}
            <p className="font-medium">{q.question}</p>
            <div className="flex flex-wrap gap-1.5">
              {q.options.map((o) => {
                const active = selected[q.question]?.has(o.label) ?? false
                return (
                  <button
                    key={o.label}
                    type="button"
                    title={o.description}
                    onClick={() => toggle(q, o.label)}
                    className={cn(
                      "rounded-md border px-2 py-1 text-left",
                      active
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border bg-card hover:bg-secondary",
                    )}
                  >
                    {o.label}
                  </button>
                )
              })}
            </div>
            <input
              value={otro[q.question] ?? ""}
              onChange={(e) => {
                const v = e.target.value
                setOtro((prev) => ({ ...prev, [q.question]: v }))
                if (v) setSelected((prev) => ({ ...prev, [q.question]: new Set() }))
              }}
              placeholder="otra respuesta…"
              className="rounded-md border border-border bg-card px-2 py-1 text-[10px] placeholder:text-muted-foreground focus:border-primary focus:outline-none"
            />
          </div>
        ))}
      </div>
      <div className="flex flex-wrap gap-1.5 p-2.5">
        <button
          type="button"
          disabled={!listo}
          onClick={enviar}
          className="rounded-md bg-primary px-2.5 py-1 font-semibold text-primary-foreground hover:brightness-110 disabled:opacity-40"
        >
          Enviar respuesta
        </button>
        <button
          type="button"
          onClick={onDeny}
          className="rounded-md border border-border bg-card px-2.5 py-1 hover:border-destructive hover:text-destructive"
        >
          Cancelar
        </button>
      </div>
    </div>
  )
}

// —— input legible por herramienta ————————————————————————————————————————————

interface EditInput {
  file_path?: string | undefined
  old_string?: string | undefined
  new_string?: string | undefined
}

interface MultiEditInput {
  file_path?: string | undefined
  edits?: { old_string?: string | undefined; new_string?: string | undefined }[] | undefined
}

// asRecord narrows the raw control_request input (unknown JSON) without trusting it.
function asRecord(v: unknown): Record<string, unknown> {
  return typeof v === "object" && v !== null ? (v as Record<string, unknown>) : {}
}

function str(v: unknown): string | undefined {
  return typeof v === "string" ? v : undefined
}

// ToolInputPreview pinta el input crudo del ask como lo haría el mockup: Edit/MultiEdit
// = diff −/+ · Write = contenido · Bash = comando · resto = JSON. Nunca inventa: si el
// shape no es el esperado, cae al JSON crudo (honesto).
export function ToolInputPreview({ tool, input }: { tool: string; input: unknown }) {
  const rec = asRecord(input)

  if (tool === "Edit") {
    const e: EditInput = {
      file_path: str(rec["file_path"]),
      old_string: str(rec["old_string"]),
      new_string: str(rec["new_string"]),
    }
    if (e.old_string !== undefined || e.new_string !== undefined) {
      return (
        <div className="overflow-hidden rounded-md border border-border bg-card">
          <FilePathRow path={e.file_path} />
          <DiffLines minus={e.old_string} plus={e.new_string} />
        </div>
      )
    }
  }

  if (tool === "MultiEdit") {
    const edits = Array.isArray(rec["edits"])
      ? (rec["edits"] as MultiEditInput["edits"])
      : undefined
    if (edits?.length) {
      return (
        <div className="overflow-hidden rounded-md border border-border bg-card">
          <FilePathRow path={str(rec["file_path"])} note={`${edits.length} ediciones`} />
          {edits.map((e, i) => (
            <DiffLines
              // biome-ignore lint/suspicious/noArrayIndexKey: el input del ask es inmutable — el índice es identidad estable.
              key={i}
              minus={str(asRecord(e)["old_string"])}
              plus={str(asRecord(e)["new_string"])}
            />
          ))}
        </div>
      )
    }
  }

  if (tool === "Write") {
    const content = str(rec["content"])
    if (content !== undefined) {
      return (
        <div className="overflow-hidden rounded-md border border-border bg-card">
          <FilePathRow path={str(rec["file_path"])} note="archivo completo" />
          <pre className="max-h-40 overflow-auto whitespace-pre-wrap bg-ok-soft px-2 py-1 font-mono text-[10px]">
            {content}
          </pre>
        </div>
      )
    }
  }

  if (tool === "Bash") {
    const cmd = str(rec["command"])
    if (cmd !== undefined) {
      return (
        <pre className="max-h-24 overflow-auto whitespace-pre-wrap rounded-md border border-border bg-card px-2 py-1.5 font-mono text-[10px]">
          <span className="text-ok">$ </span>
          {cmd}
        </pre>
      )
    }
  }

  return (
    <pre className="max-h-32 overflow-auto whitespace-pre-wrap rounded-md border border-border bg-card px-2 py-1.5 font-mono text-[10px] text-muted-foreground">
      {JSON.stringify(input ?? {}, null, 2)}
    </pre>
  )
}

function FilePathRow({ path, note }: { path: string | undefined; note?: string }) {
  return (
    <div className="flex items-center gap-1.5 border-b border-border px-2 py-1 font-mono text-[10px] text-muted-foreground">
      <span aria-hidden>✏️</span>
      <span className="truncate">{path ?? "(sin ruta)"}</span>
      {note && <span className="ml-auto flex-none">{note}</span>}
    </div>
  )
}

// DiffLines — diff mínimo de líneas cambiadas (patrón transcript: sin syntax highlight). Es el
// patrón correcto para aceptar/rechazar en bloque (investigacion.md del paquete original ya
// concluía esto); un editor CodeMirror merge solo aportaría algo si el producto pidiera editar
// el diff chunk-a-chunk ANTES de aprobar — no pedido hoy, BACKLOG.md lo separa como sub-ítem
// propio en vez de bloquear detrás de "fase de presentación".
function DiffLines({ minus, plus }: { minus: string | undefined; plus: string | undefined }) {
  return (
    <div className="max-h-40 overflow-auto font-mono text-[10px] leading-snug">
      {minus?.split("\n").map((l, i) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: contenido inmutable del ask.
        <div key={`m${i}`} className="whitespace-pre-wrap bg-crit-soft px-2">
          <span className="select-none text-crit">− </span>
          {l}
        </div>
      ))}
      {plus?.split("\n").map((l, i) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: contenido inmutable del ask.
        <div key={`p${i}`} className="whitespace-pre-wrap bg-ok-soft px-2">
          <span className="select-none text-ok">+ </span>
          {l}
        </div>
      ))}
    </div>
  )
}
