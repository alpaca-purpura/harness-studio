package marketplace

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// SincronizadorGit realiza ports.CatalogoSync con `git pull --ff-only` sobre el
// InstallLocation del marketplace (DD-2/E-bis). Es un actor SEPARADO del lector a
// propósito: `LectorLocal` sigue sin escribir jamás (enforcer
// TestLectorLocalNoEscribeEnElCheckout intacto); la política de CUÁNDO sincronizar
// (refresco explícito + clase propio) vive en el usecase, acá solo el mecanismo.
//
// `--ff-only` a propósito: si el clone tiene historia divergente (alguien editó el
// checkout a mano), el pull FALLA con motivo visible en vez de mergear en silencio un
// árbol que no es de nadie. Un pull concurrente con el de CC lo arbitra git mismo
// (index.lock): el perdedor devuelve error y el motivo se muestra.
type SincronizadorGit struct {
	GitBin  string        // "" ⇒ "git"
	Timeout time.Duration // 0 ⇒ 30s
}

var _ ports.CatalogoSync = (*SincronizadorGit)(nil)

// Sincronizar hace pull ff-only del checkout y reporta si avanzó. El chequeo de clase es
// del usecase; acá solo se exige un checkout git existente.
func (s *SincronizadorGit) Sincronizar(ctx context.Context, m domain.MarketplaceConocido) (string, error) {
	if m.InstallLocation == "" {
		return "", fmt.Errorf("sincronizar %s: sin checkout local declarado", m.Nombre)
	}
	raiz, err := filepath.EvalSymlinks(m.InstallLocation)
	if err != nil {
		return "", fmt.Errorf("sincronizar %s: installLocation ya no existe en disco: %s", m.Nombre, m.InstallLocation)
	}
	if fi, serr := os.Stat(filepath.Join(raiz, ".git")); serr != nil || !fi.IsDir() {
		return "", fmt.Errorf("sincronizar %s: el checkout no es un repo git (%s)", m.Nombre, raiz)
	}

	timeout := s.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	antes := s.head(ctx, raiz)
	git := s.GitBin
	if git == "" {
		git = "git"
	}
	cmd := exec.CommandContext(ctx, git, "-C", raiz, "pull", "--ff-only", "--quiet") //nolint:gosec // G204: git fijo del binario/config, raiz = InstallLocation registrado, jamás input remoto.
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if rerr := cmd.Run(); rerr != nil {
		motivo := strings.TrimSpace(stderr.String())
		if motivo == "" {
			motivo = rerr.Error()
		}
		return "", fmt.Errorf("sincronizar %s: git pull --ff-only: %s", m.Nombre, motivo)
	}
	despues := s.head(ctx, raiz)
	if antes != "" && antes == despues {
		return "ya al día", nil
	}
	return "avanzó a " + corto(despues), nil
}

// head devuelve el sha de HEAD, "" si no se puede leer (best-effort: solo alimenta el detalle).
func (s *SincronizadorGit) head(ctx context.Context, raiz string) string {
	git := s.GitBin
	if git == "" {
		git = "git"
	}
	out, err := exec.CommandContext(ctx, git, "-C", raiz, "rev-parse", "HEAD").Output() //nolint:gosec // G204: mismo argumento que arriba.
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func corto(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	if sha == "" {
		return "HEAD nuevo"
	}
	return sha
}
