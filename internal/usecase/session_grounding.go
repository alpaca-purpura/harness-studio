package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// GroundingSource arma la tarjeta de identidad del arnés de una sesión (RF-189/RF-190,
// paquete mejorar-arnes-conversando): qué arnés es, qué copia edita esta sesión
// (canónico / instalación / árbol sin registrar), su deriva y su rol. El composition root
// la cierra sobre el PortafolioService; nil = sin tarjeta (spawn con doctrina compartida,
// comportamiento previo).
type GroundingSource func(ctx context.Context, arnesID, cwd string) string

// SetGrounding cablea la fuente de la tarjeta de identidad. Se llama una vez en el
// composition root, antes de servir.
func (s *SessionService) SetGrounding(g GroundingSource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grounding = g
}

// TarjetaIdentidad construye la tarjeta (markdown) desde las entradas del Portafolio —
// función pura, testeable sin IO. Tres formas:
//   - cwd == canónico de una entrada → sesión sobre LA copia editable;
//   - cwd == instalación de una entrada → sesión de REPARACIÓN (A4, MC-D8): la tarjeta
//     enseña el loop diagnóstico→in-situ→causa→backport y deja claro que la deriva visible
//     es legal — lo prohibido es derivar en silencio;
//   - sin match → tarjeta mínima honesta (árbol sin registrar en el Portafolio).
func TarjetaIdentidad(entradas []domain.EntradaPortafolio, arnesID, cwd, rol string) string {
	canon := canonicalPathPortafolio(cwd)
	rolLinea := rol
	if rolLinea == "" {
		// DD-1 (deuda D): sin rol NO es read-only — el canal de permisos va siempre
		// cableado y cada escritura pasa por la tarjeta del panel. La tarjeta lo dice
		// así; la versión anterior («sin rol la sesión no puede escribir») hacía que el
		// modelo se auto-negara ANTES de intentar el Write y el HITL jamás se ejercía.
		rolLinea = "(sin rol declarado en el sello — podés escribir: cada escritura pasa por la tarjeta de permiso del panel y la aprueba el operador; sellar el rol después da autoridad fina y es una reparación válida)"
	}

	var b strings.Builder
	b.WriteString("## Tarjeta de identidad de la sesión (generada por ArnesIA, datos del Portafolio)\n\n")
	fmt.Fprintf(&b, "- **Arnés de esta sesión:** `%s` — rol: %s\n", arnesID, rolLinea)

	for _, e := range entradas {
		if e.Canonico != nil && canonicalPathPortafolio(e.Canonico.Path) == canon {
			fmt.Fprintf(&b, "- **Copia:** CANÓNICO de la identidad `%s` — la única copia editable; lo que mejores acá es la base de todas las instalaciones.\n", e.Identidad.Clave())
			return b.String()
		}
		for _, inst := range e.Instalaciones {
			if canonicalPathPortafolio(inst.InstallPath) != canon {
				continue
			}
			fmt.Fprintf(&b, "- **Copia:** INSTALACIÓN de la identidad `%s` (proyecto `%s`).\n", e.Identidad.Clave(), inst.ProyectoPath)
			fmt.Fprintf(&b, "- **Deriva actual:** %s", inst.Deriva)
			if inst.DerivaDetalle != "" {
				fmt.Fprintf(&b, " (%s)", inst.DerivaDetalle)
			}
			b.WriteString("\n\n### Sesión de REPARACIÓN (ley A4)\n\n")
			b.WriteString("Estás editando una INSTALACIÓN — el banco de pruebas real del arnés. El loop:\n\n")
			b.WriteString("1. Diagnosticá POR QUÉ no funciona: ¿la instalación/mapeo al proyecto fue mala, o el arnés está mal diseñado de base?\n")
			b.WriteString("2. Causa instalación → corregí IN SITU (como reinstalar y recablear para ESTE proyecto).\n")
			b.WriteString("3. Causa base → el fix se LEVANTA al canónico del marketplace (backport) y LUEGO se prueba sobre esta instalación.\n")
			b.WriteString("4. Cerrá SIEMPRE tu respuesta declarando la causa diagnosticada (instalación vs base) — ArnesIA lo usa para el backport.\n\n")
			b.WriteString("La deriva de esta copia queda VISIBLE en el Portafolio tras cada cambio — eso es legal y esperado; lo prohibido es derivar en silencio.\n")
			return b.String()
		}
	}

	b.WriteString("- **Copia:** árbol NO registrado en el Portafolio (carpeta suelta) — sin canónico ni deriva que rastrear todavía; «Identificar»/sellar es el camino para incorporarlo.\n")
	return b.String()
}
