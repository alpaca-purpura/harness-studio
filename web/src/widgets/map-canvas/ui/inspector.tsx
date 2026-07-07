import { useEffect, useState } from "react"
import {
  alwFor,
  type Box,
  type Clase,
  type ConformanceResult,
  type Graph,
  handleFor,
  isDelPuesto,
  KIND,
  SEC_TIP,
  selectHallazgosConformance,
  selectVieneDe,
  tipDe,
} from "@/entities/arnes"
import { Glyph } from "@/shared/canvas"
import { cn } from "@/shared/lib/cn"

// Inspector is the read-only node drawer of the Map (S3, RF-71/73 + paquete
// inspector-drawer RF-80..96). It renders the selected node's fused `contract` straight
// from the loaded graph, in three tabs (Resumen | Contenido | Corridas) with an expanded
// mode that covers the map area. Styles live in app/styles/inspector.css (verbatim port
// of mockup v6, scoped .arnesia-inspector). It is chrome the page composes over the
// canvas; it never fetches (fe-transporte-independiente: the page injects data/callbacks).
//
// Close semantics (decisión #5e): ✕ CLOSES the drawer entirely (onClose — the map stays
// without a node drawer); ⤡ collapses expanded→normal; Esc collapses only in expanded.
//
// Per-class view (inspector-por-clase.md, Tier A): a node WITHOUT a fused contract is
// framed by what it IS, pending per-class fields are named honestly, every node shows
// its Fuente (fuente_path · origen).

// CLASS_ROLE — doctrinal framing per clase for contract-less nodes. `pendiente` names the
// fields waiting on the nomenclatura §3 recognizers — said out loud, never invented.
const CLASS_ROLE: Record<Clase, { rol: string; pendiente?: string }> = {
  skill: {
    rol: "Skill de apoyo — sin contrato de caja.",
    pendiente: "description · allowed-tools",
  },
  subagent: { rol: "Subagente — apoyo de su caja.", pendiente: "description · tools · model" },
  command: { rol: "Comando de invocación manual.", pendiente: "description · argument-hint" },
  hook: {
    rol: "Hook de la Guardia — corre en un evento del ciclo.",
    pendiente: "evento · matcher · comando · timeout",
  },
  rule: { rol: "Regla de la Base — contexto que guía sin bloquear." },
  mcp: { rol: "Server MCP — capacidad externa.", pendiente: "transporte · command/url · tools" },
  plugin: {
    rol: "Plugin — contenedor del arnés (nodo raíz).",
    pendiente: "versión · marketplace · componentes",
  },
  settings: { rol: "Ajustes del arnés.", pendiente: "permisos · env · scope" },
  "output-style": { rol: "Output style — estilo de salida.", pendiente: "description" },
  statusline: { rol: "Statusline.", pendiente: "tipo · comando" },
  "no-reconocido": {
    rol: "El reconocedor no entendió este artefacto — revisa la celda canónica de su clase (nomenclatura §3).",
  },
}

// activacionRegla — the rule load axis as display text. Derived from the PROPOSAL alwFor
// fixture until the loader derives it from `paths:` (rules.md L2.6) — labeled as such.
function activacionRegla(alw: boolean | undefined): string {
  if (alw === true) return "siempre en contexto · PROPUESTA"
  if (alw === false) return "condicional (paths:) · PROPUESTA"
  return "desconocida"
}

// Section — título + «i» doctrinal (RF-85): el tooltip dice qué agrupa la sección y su
// eje del contrato fusionado. La «i» es un botón real (focusable nativo, aria-label
// legal) operable por teclado (RF-88); el texto viene del diccionario de la entity.
function Section({ title, children }: { title: string; children: React.ReactNode }) {
  const tip = SEC_TIP.get(title)
  return (
    <section className="sec">
      <h4>
        {title}
        {tip && (
          <button type="button" className="info" data-tip={tip} aria-label={`Qué agrupa ${title}`}>
            i
          </button>
        )}
      </h4>
      {children}
    </section>
  )
}

// Field — k/v con tooltip doctrinal (RF-86): el nombre del campo se subraya punteado y
// su tooltip explica el campo Y el valor concreto (tipDe, diccionario de la entity).
// Campos fuera del diccionario no se subrayan (jamás un tooltip vacío).
function Field({ k, v }: { k: string; v?: string | undefined }) {
  if (!v) return null
  const tip = tipDe(k, v)
  return (
    <div className="field">
      {tip ? (
        // biome-ignore lint/a11y/noNoninteractiveTabindex: tooltip CSS-only operable por teclado (RF-88) — el foco dispara :focus-visible::after.
        <span className="k tip" tabIndex={0} data-tip={tip}>
          {k}
        </span>
      ) : (
        <span className="k">{k}</span>
      )}
      <span className="v">{v}</span>
    </div>
  )
}

type Tab = "resumen" | "contenido" | "corridas"

const TABS: readonly { id: Tab; label: string }[] = [
  { id: "resumen", label: "Resumen" },
  { id: "contenido", label: "Contenido" },
  { id: "corridas", label: "Corridas" },
]

export interface InspectorProps {
  // Sin box = estado vacío (RF-84): la línea de affordance, no un panel en blanco ni ausencia.
  box?: Box | undefined
  onClose: () => void
  // El grafo cargado: alimenta Viene de (edges inversos) y la navegabilidad de los chips.
  graph?: Graph | undefined
  // Navegación de chips (RF-90): el mismo mecanismo que click en nodo — lo inyecta la página.
  onSelect?: ((id: string) => void) | undefined
  // Resultados de GET …/conformance (los inyecta la página); undefined = no disponible (se dice).
  conformance?: readonly ConformanceResult[] | undefined
}

export function Inspector({ box, onClose, graph, onSelect, conformance }: InspectorProps) {
  const [expanded, setExpanded] = useState(false)
  const [tab, setTab] = useState<Tab>("resumen")

  // Changing node resets to Resumen (each drawer opens on its summary, as the mockup's
  // per-card default); losing the selection also drops the expanded overlay.
  const boxId = box?.id
  useEffect(() => {
    setTab("resumen")
    if (!boxId) setExpanded(false)
  }, [boxId])

  // Esc COLLAPSES (never closes) — only listening while expanded (decisión #5e).
  useEffect(() => {
    if (!expanded) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setExpanded(false)
    }
    document.addEventListener("keydown", onKey)
    return () => document.removeEventListener("keydown", onKey)
  }, [expanded])

  if (!box) {
    return (
      <aside aria-label="Inspector · sin selección" className="arnesia-inspector empty">
        <div className="dw-body">
          <div className="tabpane">
            <Section title="Inspector">
              <p>
                Clic en un nodo del mapa: identidad, clasificación doctrinal, contrato, fuente,
                corridas y hallazgos.
              </p>
            </Section>
          </div>
        </div>
      </aside>
    )
  }

  const k = KIND[box.clase]

  return (
    <aside
      aria-label={`Inspector · ${box.nombre}`}
      className={cn("arnesia-inspector", expanded && "expanded")}
    >
      <header className="dw-head">
        <Glyph color={k.color} char={k.char} shape={k.shape} />
        <div className="dw-id">
          <div className="dw-nombre">{box.nombre}</div>
          <div className="dw-handle">
            {k.label} · <span className="mono">{handleFor(box, alwFor(box.id))}</span>
          </div>
        </div>
        <button
          type="button"
          className="dw-expand"
          aria-pressed={expanded}
          aria-label={expanded ? "Colapsar al drawer normal" : "Ampliar inspector"}
          title={
            expanded
              ? "Colapsar: vuelve al drawer lateral normal"
              : "Ampliar: el drawer ocupa todo el espacio del mapa"
          }
          onClick={() => setExpanded((e) => !e)}
        >
          {expanded ? "⤡" : "⤢"}
        </button>
        <button
          type="button"
          className="dw-close"
          aria-label="Cerrar inspector"
          title="Cerrar: quita el drawer y deja el mapa (≠ colapsar ⤡)"
          onClick={onClose}
        >
          ✕
        </button>
      </header>

      {/* div (no nav): a11y noNoninteractiveElementToInteractiveRole — tablist ARIA puro. */}
      <div className="dw-tabs" role="tablist" aria-label="Vistas del nodo">
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            className="dw-tab"
            role="tab"
            id={`dw-tab-${t.id}`}
            aria-selected={tab === t.id}
            aria-controls={`dw-pane-${t.id}`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="dw-body">
        <div
          className="tabpane tabpane-resumen"
          role="tabpanel"
          id="dw-pane-resumen"
          aria-labelledby="dw-tab-resumen"
          hidden={tab !== "resumen"}
        >
          <Resumen box={box} graph={graph} onSelect={onSelect} conformance={conformance} />
        </div>
        <div
          className="tabpane"
          role="tabpanel"
          id="dw-pane-contenido"
          aria-labelledby="dw-tab-contenido"
          hidden={tab !== "contenido"}
        >
          <Contenido box={box} />
        </div>
        <div
          className="tabpane"
          role="tabpanel"
          id="dw-pane-corridas"
          aria-labelledby="dw-tab-corridas"
          hidden={tab !== "corridas"}
        >
          <Corridas box={box} />
        </div>
      </div>
    </aside>
  )
}

// Chip — insumo/destino tipado y NAVEGABLE (RF-90): click en un chip cuyo destino existe
// en el grafo cargado selecciona ese nodo (mismo mecanismo que click en nodo, vía el
// onSelect que inyecta la página); destino ausente → chip inerte rotulado, jamás muerto
// en silencio.
function Chip({
  label,
  org,
  targetId,
  graph,
  onSelect,
}: {
  label: string
  org?: string | undefined
  targetId?: string | undefined
  graph?: Graph | undefined
  onSelect?: ((id: string) => void) | undefined
}) {
  const navegable =
    targetId !== undefined && onSelect !== undefined && graph?.nodos.some((n) => n.id === targetId)
  return (
    <button
      type="button"
      className="chip"
      disabled={!navegable}
      title={navegable ? `Ver ${targetId} en el mapa` : "nodo fuera del grafo cargado"}
      onClick={navegable ? () => onSelect(targetId) : undefined}
    >
      {label}
      {org ? <span className="org">{org}</span> : null}
    </button>
  )
}

// Hallazgos — SIEMPRE presente (RF-91), determinista del dato: gate:none (A4, crit) ·
// no-reconocido (D-c, warn) · checks rojos del endpoint de conformance filtrados por
// nodo. «Sin hallazgos abiertos» también informa; conformance ausente se DICE.
function Hallazgos({
  box,
  conformance,
}: {
  box: Box
  conformance?: readonly ConformanceResult[] | undefined
}) {
  const rojos = conformance ? selectHallazgosConformance(conformance, box.id) : []
  const gateNone = box.contract?.gate?.tipo === "none"
  const noReconocido = box.clase === "no-reconocido"
  const sinHallazgos = !gateNone && !noReconocido && rojos.length === 0
  return (
    <Section title="Hallazgos">
      {gateNone && (
        <p className="hallazgo crit">
          <b>gate:none</b> — caja sin eval formal (A4): el hueco es un hallazgo visible, no un
          blanco.
        </p>
      )}
      {noReconocido && (
        <p className="hallazgo warn">
          <b>no-reconocido</b> — el loader no entendió el artefacto (D-c): revisa su celda canónica.
        </p>
      )}
      {rojos.map((r) => (
        <p
          key={r.check.id}
          className={cn("hallazgo", r.check.severidad === "error" ? "crit" : "warn")}
        >
          <b>{r.check.id}</b> — {r.detalle ?? r.check.que ?? "check rojo de conformance"}
        </p>
      ))}
      {sinHallazgos && <p className="mut">Sin hallazgos abiertos.</p>}
      {conformance === undefined && (
        <p className="src-note">
          checks de conformance no disponibles — el daemon no respondió{" "}
          <span className="mono">GET …/conformance</span>.
        </p>
      )}
    </Section>
  )
}

// Botonera staged (RF-92): acciones del NODO al pie del Resumen — disabled + rotuladas
// con la fase que las cablea, jamás fingiendo funcionar.
function BotoneraStaged() {
  return (
    <footer className="dw-actions">
      <button
        type="button"
        className="act primary"
        disabled
        title="Se cablea en Fase 3/4 del Hito 2 — el backend (Fase E) ya vive"
      >
        Editar conversando
      </button>
      <button
        type="button"
        className="act"
        disabled
        title="Release train (KIT-06) — fuera de este paquete"
      >
        Evaluar A/B contra v anterior
      </button>
      <button
        type="button"
        className="act"
        disabled
        title="Release train (KIT-06) — fuera de este paquete"
      >
        Promover a estable
      </button>
      <button type="button" className="act" disabled title="Vista Historia — fuera de este paquete">
        Ver en Historia
      </button>
      <p className="act-note">
        staged — «Editar conversando» se cablea en Fase 3/4; el resto espera tren/Historia.
      </p>
    </footer>
  )
}

// Resumen — identidad doctrinal + contrato fusionado del nodo (tab por defecto).
function Resumen({
  box,
  graph,
  onSelect,
  conformance,
}: {
  box: Box
  graph?: Graph | undefined
  onSelect?: ((id: string) => void) | undefined
  conformance?: readonly ConformanceResult[] | undefined
}) {
  const c = box.contract
  const role = CLASS_ROLE[box.clase]
  // origen: the real L0 field wins; the proposals fixture stands in, labeled (spec §2.3).
  const origen = box.origen ?? (isDelPuesto(box.id) ? "del-puesto · PROPUESTA" : undefined)
  // Viene de (RF-89): edges inversos derivados del grafo; sin entradas → sección ausente.
  const vieneDe = graph ? selectVieneDe(graph, box.id) : []

  return (
    <>
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

      {(box.fuente_path || origen) && (
        <Section title="Fuente">
          <Field k="fuente" v={box.fuente_path} />
          <Field k="origen" v={origen} />
        </Section>
      )}

      {box.clase === "rule" && (
        <Section title="Activación">
          <Field k="carga" v={activacionRegla(alwFor(box.id))} />
        </Section>
      )}

      {!c ? (
        <Section title={box.clase === "no-reconocido" ? "Reconciliación" : "Rol"}>
          {/* Warn = fondo suave, texto normal — text-warn a 11px no pasa contraste AA. */}
          <p className={box.clase === "no-reconocido" ? "warnbox" : "mut"}>{role.rol}</p>
          {role.pendiente && (
            <p className="mut">
              pendiente del reconocedor: <span className="mono">{role.pendiente}</span>
            </p>
          )}
        </Section>
      ) : (
        <>
          {c.why && (
            <Section title="Intención">
              <p className="fg">{c.why}</p>
            </Section>
          )}

          {c.capabilities && c.capabilities.length > 0 && (
            <Section title="Capabilities">
              <ul className="caps">
                {c.capabilities.map((cap) => (
                  <li key={cap.id}>
                    <span className="mono mut">{cap.id}</span> {cap.what}
                    <div className="mut">✓ {cap.success}</div>
                  </li>
                ))}
              </ul>
            </Section>
          )}

          {c.necesita && c.necesita.length > 0 && (
            <Section title="Necesita">
              <div className="chips">
                {c.necesita.map((inp) => {
                  // Origen tipado del insumo: "caja:edit-caja" → tipo=caja, id=edit-caja;
                  // "usuario" (sin `:`) → solo tipo, chip inerte.
                  const idx = inp.de.indexOf(":")
                  const targetId = idx >= 0 ? inp.de.slice(idx + 1) : undefined
                  return (
                    <Chip
                      key={`${inp.art}-${inp.de}`}
                      label={inp.art + (inp.requerido === false ? " (opcional)" : "")}
                      org={`← ${idx >= 0 ? `${inp.de.slice(0, idx)}: ${targetId}` : inp.de}`}
                      targetId={targetId}
                      graph={graph}
                      onSelect={onSelect}
                    />
                  )
                })}
              </div>
            </Section>
          )}

          {c.entrega && c.entrega.length > 0 && (
            <Section title="Entrega">
              <ul>
                {c.entrega.map((out) => (
                  <li key={out.art}>
                    {out.art}
                    {out.escritor_unico ? <span className="mut"> · escritor único</span> : null}
                  </li>
                ))}
              </ul>
            </Section>
          )}

          {c.ruta && c.ruta.length > 0 && (
            <Section title="Ruta">
              <div className="chips">
                {c.ruta.map((r) => (
                  <span key={`${r.a}-${r.si ?? ""}`} className="chiprow">
                    → <Chip label={r.a} targetId={r.a} graph={graph} onSelect={onSelect} />
                    {r.si ? <span className="cond">si {r.si}</span> : null}
                  </span>
                ))}
              </div>
            </Section>
          )}

          {c.gate && (
            <Section title="Gate">
              <Field k="tipo" v={c.gate.tipo} />
              {c.gate.detalle && <p className="fg">{c.gate.detalle}</p>}
            </Section>
          )}

          {c.handoff && (
            <Section title="Handoff">
              <p className="fg">
                {c.handoff.cuando} <span className="mut">→ {c.handoff.a}</span>
              </p>
            </Section>
          )}
        </>
      )}

      {/* Viene de (RF-89): edges INVERSOS reales con su tipo; sin entradas = ausente. */}
      {vieneDe.length > 0 && (
        <Section title="Viene de">
          <div className="chips">
            {vieneDe.map((v) => (
              <span key={`${v.de}-${v.tipo}`} className="chiprow">
                <Chip label={v.de} org={v.tipo} targetId={v.de} graph={graph} onSelect={onSelect} />
              </span>
            ))}
          </div>
        </Section>
      )}

      <Hallazgos box={box} conformance={conformance} />

      <BotoneraStaged />
    </>
  )
}

// Contenido — la fuente del componente (RF-93..95). Sin conexión al daemon (esta
// versión aún no recibe loadFuente) el estado se DICE — jamás contenido inventado.
function Contenido({ box }: { box: Box }) {
  return (
    <>
      <span className="chip-versiona">versiona con el arnés</span>
      <Section title="Fuente del componente">
        <Field k="fuente" v={box.fuente_path} />
        {box.fuente_path ? (
          <p className="mut">
            El daemon sirve el archivo real confinado al dir del arnés (S2) — lectura no disponible
            en esta vista.
          </p>
        ) : (
          <p className="mut">
            Sin <span className="mono">fuente_path</span> — pendiente del reconocedor (nomenclatura
            §3): el loader aún no estampa la celda canónica de esta clase.
          </p>
        )}
      </Section>
    </>
  )
}

// Corridas — estado honesto: el indexer JSONL aún no existe (Hito 3). Para cajas se
// dice qué listará primero (runs D2 + la sesión CC viva del frente).
function Corridas({ box }: { box: Box }) {
  const esCaja = box.contract?.caja === true
  return (
    <>
      <Section title="Corridas donde actuó">
        <p className="mut">Sin corridas indexadas — llegan con el indexer JSONL (Hito 3).</p>
        {esCaja && (
          <p className="src-note">
            al implementar listan primero: las corridas de caja de{" "}
            <span className="mono">POST …/boxes/{box.id}/run</span> (D2, ya vivo) y la sesión CC
            viva del frente.
          </p>
        )}
      </Section>
      <div className="dw-actions">
        <button
          type="button"
          className="act"
          disabled
          title="Vista Corridas del arnés — Hito 3 (detalle: Conversación · Árbol · Waterfall · replay en el mapa)"
        >
          Ver todas las corridas del arnés
        </button>
      </div>
    </>
  )
}
