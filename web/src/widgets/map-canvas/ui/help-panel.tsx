import { KIND } from "@/entities/arnes"
import { Glyph } from "@/shared/canvas"

// HelpPanel is the floating legend (mockup:174-199, shot6): relations, classes (6 glyphs
// deduped by shape), caja, regions, and the PROPUESTA facets (activation/origen) declared
// explicitly as proposals vs the canonical «Base». Read-only reference, toggled by the «?» FAB.

// One glyph per distinct shape (mockup legendGlyphs dedupes by shape): skill·subagent·hook·
// rule·mcp·command.
const LEGEND = Object.values(KIND).filter(
  (k, i, all) => all.findIndex((x) => x.shape === k.shape) === i,
)

export function HelpPanel({ hidden }: { hidden: boolean }) {
  return (
    <div className="help-panel" hidden={hidden}>
      <h3>ArnesIA · Mapa — MVP (Hito 1)</h3>
      <p>Superficie read-only, capa Estructura. Derivado de los componentes de Storybook (SSOT).</p>

      <h4>Relaciones</h4>
      <div className="lg-row">
        <span className="lg-line" /> invoca (flujo, con flecha)
      </div>
      <div className="lg-row">
        <span className="lg-line escribe" /> escribe (punteado)
      </div>
      <div className="lg-row">
        <span className="lg-line lee" /> lee (punteado)
      </div>
      <p style={{ marginTop: 8 }}>
        El <b>flujo</b> (invoca) siempre se ve. <b>Pasa el cursor</b> por un nodo para revelar qué{" "}
        <b>lee/escribe</b> y resaltar sus vecinos.
      </p>

      <h4>Clases (forma + color)</h4>
      <div className="lg-glyphs">
        {LEGEND.map((k) => (
          <span key={k.label} className="lg-g">
            <Glyph color={k.color} char={k.char} shape={k.shape} />
            {k.label}
          </span>
        ))}
      </div>

      <h4>Caja de proceso</h4>
      <p>
        El skill-frente de una fase (<b>contract.caja</b>) va con borde grueso, tinte y badge
        «caja». Los demás skills son de apoyo.
      </p>

      <h4>Regiones</h4>
      <p>
        <b>Guardia</b> (hooks) · <b>Proceso</b> (carriles por fase) · <b>Base</b> (reglas, knowledge
        y soporte del arnés).
      </p>

      <h4>Activación (propuesta)</h4>
      <p>
        La honestidad no es la banda sino la activación: <b>siempre en contexto</b> (reglas
        always-on) · <b>carga condicional</b> (reglas <code>paths:</code>) · <b>bajo demanda</b>{" "}
        (expertos, meta, terceros) · <b>dormida</b> (0 corridas). Las reglas se pliegan porque
        suelen ser muchas.
      </p>

      <h4>Origen (propuesta)</h4>
      <p>
        Borde izquierdo <b>punteado</b> = nodo <b>del puesto</b> (llenado en onboarding con
        conocimiento de rol·proceso·empresa). Borde sólido = <b>estándar</b> (viene del kit,
        agnóstico). Badge <b>propuesto</b> = índice semántico opcional (segundo cerebro; jamás
        dependencia dura, la norma es knowledge as-code path-scoped).
      </p>

      <h4>Navegación</h4>
      <ul>
        <li>
          <b>⤢ Ajustar</b> encuadra todo el arnés (vista al abrir); <b>+/−</b> y <b>ctrl+rueda</b>{" "}
          acercan; arrastra para desplazar.
        </li>
        <li>Sustrato HTML+SVG (no React Flow) — RF va al Organigrama.</li>
        <li>Solo capa Estructura; Tokens/Desempeño/Proceso esperan telemetría.</li>
        <li>
          Región inferior = <b>Base</b> (canónico VISION A6); el eje «activación» y el
          default-vs-llenado son propuestas a ratificar.
        </li>
      </ul>
    </div>
  )
}
