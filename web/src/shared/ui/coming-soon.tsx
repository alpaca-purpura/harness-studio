import { cn } from "@/shared/lib/cn"

interface ComingSoonProps {
  glyph?: string
  title: string
  note?: string
  className?: string
}

// ComingSoon is the neutral placeholder for every view whose interior is not built yet
// (scope: shell + multisesión + conversación first). Same look everywhere so the shell
// reads as intentionally staged, not broken.
export function ComingSoon({ glyph = "◵", title, note, className }: ComingSoonProps) {
  return (
    <div
      className={cn(
        "flex h-full w-full flex-col items-center justify-center gap-3 p-10 text-center",
        className,
      )}
    >
      <div className="grid size-16 place-items-center rounded-2xl border border-border bg-card text-3xl text-muted-foreground">
        {glyph}
      </div>
      <div className="text-lg font-semibold text-foreground">{title}</div>
      <span className="rounded-full border border-dashed border-input px-3 py-1 font-mono text-xs text-muted-foreground">
        próximamente
      </span>
      {note ? <p className="max-w-md text-sm text-muted-foreground">{note}</p> : null}
    </div>
  )
}
