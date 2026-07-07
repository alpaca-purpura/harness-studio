import { type Box, handleFor, KIND } from "@/entities/arnes"
import { Glyph } from "@/shared/canvas"

// Inspector is the read-only node detail surface (S3, RF-71/73). It renders the selected node's
// fused `contract` (why · capabilities · arquetipo/perfil · necesita/entrega/ruta · gate · handoff)
// straight from the loaded graph — real, complete data (the dogfood/luana nodes carry it), so no
// getNode round-trip is needed in the MVP. Not in the signed shots (Hito 2), so it is chrome the
// page composes over the canvas; editing is Hito 3+.

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-1.5">
      <h4 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        {title}
      </h4>
      {children}
    </section>
  )
}

function Field({ k, v }: { k: string; v?: string | undefined }) {
  if (!v) return null
  return (
    <div className="flex gap-2 text-xs">
      <span className="shrink-0 text-muted-foreground">{k}</span>
      <span className="font-mono text-foreground">{v}</span>
    </div>
  )
}

export function Inspector({ box, onClose }: { box: Box; onClose: () => void }) {
  const k = KIND[box.clase]
  const c = box.contract

  return (
    <aside
      aria-label={`Inspector · ${box.nombre}`}
      className="absolute right-0 top-0 flex h-full w-[340px] flex-col overflow-auto border-l border-border bg-card"
    >
      <header className="flex items-start gap-2.5 border-b border-border p-4">
        <Glyph color={k.color} char={k.char} shape={k.shape} />
        <div className="min-w-0 flex-1">
          <div className="font-mono text-sm font-semibold text-foreground">{box.nombre}</div>
          <div className="text-xs text-muted-foreground">
            {k.label} · <span className="font-mono">{handleFor(box)}</span>
          </div>
        </div>
        <button
          type="button"
          aria-label="Cerrar inspector"
          onClick={onClose}
          className="grid size-6 shrink-0 place-items-center rounded-md border border-border text-muted-foreground hover:text-foreground"
        >
          ✕
        </button>
      </header>

      <div className="flex flex-col gap-4 p-4">
        <Section title="Clasificación">
          <Field k="banda" v={box.banda} />
          <Field k="fase" v={box.fase} />
          <Field k="estado" v={box.estado} />
          <Field k="canal" v={box.canal} />
          <Field k="procedencia" v={box.procedencia} />
          {c && (
            <>
              <Field k="arquetipo" v={c.arquetipo} />
              <Field k="perfil" v={c.perfil_harness} />
              <Field k="caja" v={c.caja ? "sí" : "no"} />
            </>
          )}
        </Section>

        {!c ? (
          <p className="text-xs italic text-muted-foreground">
            Sin contrato — este nodo no es una caja de proceso.
          </p>
        ) : (
          <>
            {c.why && (
              <Section title="Intención">
                <p className="text-xs text-foreground">{c.why}</p>
              </Section>
            )}

            {c.capabilities && c.capabilities.length > 0 && (
              <Section title="Capabilities">
                <ul className="flex flex-col gap-1.5">
                  {c.capabilities.map((cap) => (
                    <li key={cap.id} className="text-xs">
                      <span className="font-mono text-muted-foreground">{cap.id}</span> {cap.what}
                      <div className="text-muted-foreground">✓ {cap.success}</div>
                    </li>
                  ))}
                </ul>
              </Section>
            )}

            {c.necesita && c.necesita.length > 0 && (
              <Section title="Necesita">
                <ul className="flex flex-col gap-1 text-xs">
                  {c.necesita.map((inp) => (
                    <li key={`${inp.art}-${inp.de}`}>
                      {inp.art} <span className="text-muted-foreground">← {inp.de}</span>
                      {inp.requerido ? "" : " (opcional)"}
                    </li>
                  ))}
                </ul>
              </Section>
            )}

            {c.entrega && c.entrega.length > 0 && (
              <Section title="Entrega">
                <ul className="flex flex-col gap-1 text-xs">
                  {c.entrega.map((out) => (
                    <li key={out.art}>
                      {out.art}
                      {out.escritor_unico ? (
                        <span className="text-muted-foreground"> · escritor único</span>
                      ) : null}
                    </li>
                  ))}
                </ul>
              </Section>
            )}

            {c.ruta && c.ruta.length > 0 && (
              <Section title="Ruta">
                <ul className="flex flex-col gap-1 text-xs">
                  {c.ruta.map((r) => (
                    <li key={`${r.a}-${r.si ?? ""}`}>
                      → <span className="font-mono">{r.a}</span>
                      {r.si ? <span className="text-muted-foreground"> si {r.si}</span> : null}
                    </li>
                  ))}
                </ul>
              </Section>
            )}

            {c.gate && (
              <Section title="Gate">
                <Field k="tipo" v={c.gate.tipo} />
                {c.gate.detalle && <p className="text-xs text-foreground">{c.gate.detalle}</p>}
              </Section>
            )}

            {c.handoff && (
              <Section title="Handoff">
                <p className="text-xs text-foreground">
                  {c.handoff.cuando} <span className="text-muted-foreground">→ {c.handoff.a}</span>
                </p>
              </Section>
            )}
          </>
        )}
      </div>
    </aside>
  )
}
