import type { PermissionAsk } from "@/shared/api"

// PermissionCard (RF-113, mockup-chat.html #permCard): la tarjeta inline de un
// control_request — herramienta + input legible + 3 acciones. La decisión viaja al
// daemon (POST /permission) y la tarjeta se cierra con el frame permission_result;
// «una vez» acota el grant a 1 s, «esta sesión» usa el TTL del rol (solo estrechable).
export function PermissionCard({
  ask,
  onResolve,
}: {
  ask: PermissionAsk
  onResolve: (decision: "allow" | "deny", once?: boolean) => void
}) {
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
          onClick={() => onResolve("allow", true)}
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
