package marketplace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// LectorLocal lee el catálogo del CHECKOUT que Claude Code ya mantiene: cero red, cero auth,
// funciona offline (AG-D9/AG-D10 punto 1). Es el camino barato y el primero que el usecase
// intenta. **Nunca escribe** en el checkout ajeno (enforcer del boundary
// `marketplace-referencia-es-solo-procedencia`, L2 punto 3).
type LectorLocal struct{}

var _ ports.CatalogoReader = (*LectorLocal)(nil)

// Fuente identifica la vía para EstadoLectura.Fuente (BR-3, trazabilidad).
func (l *LectorLocal) Fuente() string { return "local" }

// Leer resuelve `<InstallLocation>/.claude-plugin/marketplace.json` y, best-effort, el
// enriquecimiento opcional `<InstallLocation>/catalogo.json`. ctx no se usa (I/O local
// síncrono) — se acepta por simetría con el lector remoto.
func (l *LectorLocal) Leer(_ context.Context, m domain.MarketplaceConocido) (domain.Catalogo, error) {
	if m.InstallLocation == "" {
		return domain.Catalogo{}, fmt.Errorf("%w: el marketplace no tiene checkout local declarado", ErrNoEsMarketplace)
	}
	// EvalSymlinks primero (E-53): un symlink válido se sigue; uno ROTO o a un dir inexistente
	// no se sigue nunca — el motivo lo dice y el usecase cae al remoto (E-35).
	raiz, err := filepath.EvalSymlinks(m.InstallLocation)
	if err != nil {
		return domain.Catalogo{}, fmt.Errorf("%w: installLocation ya no existe en disco: %s", ErrNoEsMarketplace, m.InstallLocation)
	}
	if fi, serr := os.Stat(raiz); serr != nil || !fi.IsDir() {
		return domain.Catalogo{}, fmt.Errorf("%w: installLocation ya no existe en disco: %s", ErrNoEsMarketplace, m.InstallLocation)
	}

	cat, motivo, err := LeerCatalogoDeDir(raiz, m)
	if err != nil {
		return domain.Catalogo{}, err
	}
	cat.Lectura = domain.EstadoLectura{Tipo: domain.LecturaLeida, Entradas: len(cat.Entradas), Motivo: motivo, Fuente: l.Fuente()}
	return cat, nil
}

// LeerCatalogoDeDir es el mecanismo compartido: parsea el `marketplace.json` de raiz, aplica el
// enriquecimiento opcional y verifica las rutas relativas contra el disco. Exportada porque el
// usecase de `Traer` (camino A) necesita releer el catálogo del checkout sin pasar por el
// EstadoLectura del puerto. Devuelve (catálogo, motivo-adicional, error).
func LeerCatalogoDeDir(raiz string, m domain.MarketplaceConocido) (domain.Catalogo, string, error) {
	ruta := filepath.Join(raiz, ".claude-plugin", "marketplace.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta bajo el checkout que CC ya mantiene.
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			// E-36: sin permiso de lectura es «no puedo mirar», no «no es un marketplace».
			return domain.Catalogo{}, "", fmt.Errorf("marketplace: %s: %w", ruta, err)
		}
		return domain.Catalogo{}, "", fmt.Errorf("%w: %s: %w", ErrNoEsMarketplace, ruta, err)
	}

	cat, motivo, perr := parsearMarketplaceJSON(b)
	if perr != nil {
		return domain.Catalogo{}, "", perr
	}
	cat.Clase = m.Clase
	cat.Repo = m.Repo
	if m.Nombre != "" {
		cat.Marketplace = m.Nombre // el nombre del REGISTRO es la clave; el del archivo es un dato.
	}

	// Enriquecimiento OPCIONAL (`catalogo.json`, convención de prenter): ausente ⇒ degrada SIN
	// ruido; presente pero roto o de otro marketplace ⇒ aviso VISIBLE (E-72/E-73). Degradar sin
	// ruido ≠ ocultar un archivo roto que el operador puso a propósito.
	if eb, eerr := os.ReadFile(filepath.Join(raiz, "catalogo.json")); eerr == nil { //nolint:gosec // G304: ídem.
		if aviso := aplicarEnriquecimiento(&cat, eb); aviso != "" {
			motivo = strings.TrimSpace(motivo + " " + aviso)
		}
	}

	// Verificación de rutas SOLO en la lectura local (§5.2.6: ya estamos en disco, el `stat` es
	// barato). La fila se muestra IGUAL — el catálogo declara lo que declara; no somos nosotros
	// los que la borramos (E-34).
	for i := range cat.Entradas {
		src := cat.Entradas[i].Source
		if src.Tipo != domain.SourceRutaRelativa || src.Ruta == "" {
			continue
		}
		destino := filepath.Join(raiz, filepath.FromSlash(src.Ruta))
		if fi, serr := os.Stat(destino); serr != nil || !fi.IsDir() {
			cat.Entradas[i].Aviso = append(cat.Entradas[i].Aviso,
				fmt.Sprintf("el catálogo apunta a %s pero ese dir no existe en el checkout", src.Ruta))
		}
	}
	return cat, motivo, nil
}
