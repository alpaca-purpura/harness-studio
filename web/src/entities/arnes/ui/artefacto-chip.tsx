import { cn } from "@/shared/lib/cn"
import type { ChipArtefacto } from "../model/artefactos"

// ArtefactoChip — el chip de documento derivado del contrato (franja-artefactos D1–D11,
// RF-141; mockup v2:122-139,490-508). Papel con esquina doblada (art-ico); relleno =
// opaco (D10); nombre en cursiva = etiqueta sin path (D3); tags = ↻vN (refina D9) ·
// plantilla (D4) · externo (D10) · sin consumidor (dead-end D8) · salida del proceso
// (terminal C15). data-node-id ancla el overlay de edges — el motor no cambia.

interface ArtefactoChipProps {
  chip: ChipArtefacto
  dim?: boolean | undefined
  // Click = ir a su caja productora (externos: al primer consumidor) — RF-145.
  onSelect?: ((id: string) => void) | undefined
}

export function ArtefactoChip({ chip, dim, onSelect }: ArtefactoChipProps) {
  const target = chip.productor ?? chip.consumidores[0]?.id
  const title =
    chip.art +
    (chip.path ? "" : " — etiqueta sin path (D3)") +
    (chip.opaco ? " — opaco (D10)" : "") +
    (chip.version > 1 ? ` — revisión v${chip.version} (refina, D9)` : "")
  return (
    <button
      type="button"
      className={cn(
        "artchip on",
        chip.externo && "externo",
        chip.dead && "dead",
        chip.final && "final",
        !chip.path && "etiqueta",
        chip.opaco && "opaco",
        dim && "dim",
      )}
      data-node-id={chip.id}
      title={title}
      onClick={() => target && onSelect?.(target)}
    >
      <span className="art-top">
        <span className="art-ico" />
        <span className="art-nm">{chip.art}</span>
      </span>
      <span className="art-src">
        {chip.productor ? `escribe: ${chip.productor}` : `entra de: ${chip.origen ?? ""}`}
      </span>
      {(chip.version > 1 || chip.plantilla || chip.externo || chip.dead || chip.final) && (
        <span className="art-tags">
          {chip.version > 1 && (
            <span className="art-tag" style={{ "--at": "var(--primary)" } as React.CSSProperties}>
              ↻ v{chip.version}
            </span>
          )}
          {chip.plantilla && (
            <span
              className="art-tag"
              style={{ "--at": "var(--c-knowledge)" } as React.CSSProperties}
            >
              plantilla
            </span>
          )}
          {chip.externo && (
            <span
              className="art-tag dashed"
              style={{ "--at": "var(--muted-foreground)" } as React.CSSProperties}
            >
              externo
            </span>
          )}
          {chip.dead && (
            <span className="art-tag" style={{ "--at": "var(--warn)" } as React.CSSProperties}>
              sin consumidor
            </span>
          )}
          {chip.final && (
            <span className="art-tag" style={{ "--at": "var(--ok)" } as React.CSSProperties}>
              salida del proceso
            </span>
          )}
        </span>
      )}
    </button>
  )
}
