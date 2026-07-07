import { type CSSProperties, useState } from "react"
import { ArnesNode, alwFor, type Box } from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"
import { ActivationChip } from "./activation-chip"

// RulesSubband is the collapsible «Reglas» group inside the Base band (mockup:363-387,
// RF-40/41/42). Collapsed by default (a real arnés brings many rules — 46 in the real case).
// Expanded, it splits rules into two groups by the PROPOSAL `alw` axis: «siempre en contexto
// (CLAUDE.md)» (tone --crit, alw===true) and «carga condicional (paths:)» (tone --warn, else).
// The header shows total + «N siempre» / «N condicional» chips.

type GroupStyle = CSSProperties & { "--ac"?: string }

function RuleGroup({
  tag,
  tone,
  nodes,
  related,
  selectedId,
  onSelect,
}: {
  tag: string
  tone: string
  nodes: Box[]
  related?: ReadonlySet<string> | undefined
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
}) {
  if (nodes.length === 0) return null
  const style: GroupStyle = { "--ac": tone }
  return (
    <div className="rule-grp">
      <div className="rule-grp-hd" style={style}>
        {tag}
      </div>
      <div className="row">
        {nodes.map((b) => (
          <div key={b.id} className="node-w">
            <ArnesNode
              box={b}
              dim={related !== undefined && !related.has(b.id)}
              selected={b.id === selectedId}
              onSelect={onSelect}
            />
          </div>
        ))}
      </div>
    </div>
  )
}

interface RulesSubbandProps {
  rules: Box[]
  related?: ReadonlySet<string> | undefined
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function RulesSubband({ rules, related, selectedId, onSelect }: RulesSubbandProps) {
  const [collapsed, setCollapsed] = useState(true)
  const always = rules.filter((r) => alwFor(r.id) === true)
  const conditional = rules.filter((r) => alwFor(r.id) !== true)

  // Collapsing toggles the body's `display:none`, resizing `.content`; the ResizeObserver in
  // use-edge-paths then redraws the edges whose endpoints appeared/vanished (RF-45).
  const toggle = () => setCollapsed((c) => !c)

  return (
    <div className={cn("subband", collapsed && "collapsed")}>
      <button type="button" className="subband-hd" aria-expanded={!collapsed} onClick={toggle}>
        <span className="tw" aria-hidden>
          {collapsed ? "▸" : "▾"}
        </span>
        <h4>Reglas</h4>
        <span className="count">{rules.length}</span>
        <ActivationChip label={`${always.length} siempre`} tone="var(--crit)" />
        <ActivationChip label={`${conditional.length} condicional`} tone="var(--warn)" />
      </button>
      <div className="subband-body">
        <RuleGroup
          tag="siempre en contexto (CLAUDE.md)"
          tone="var(--crit)"
          nodes={always}
          related={related}
          selectedId={selectedId}
          onSelect={onSelect}
        />
        <RuleGroup
          tag="carga condicional (paths:)"
          tone="var(--warn)"
          nodes={conditional}
          related={related}
          selectedId={selectedId}
          onSelect={onSelect}
        />
      </div>
    </div>
  )
}
