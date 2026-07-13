package domain

import "strings"

// CanonicalizarRepo normaliza cualquier forma de referencia git a "host/owner/repo"
// (minúsculas, sin esquema, sin ".git", sin trailing slash) — RN-IDENT-1: la identidad y
// todo git-remote se canonicalizan ANTES de comparar o keyear, así "owner/repo" y
// "https://host/owner/repo.git" del mismo repo colapsan a una clave. Formas aceptadas:
// "owner/repo" (asume github.com), "https://host/owner/repo(.git)(/)",
// "git@host:owner/repo(.git)", "ssh://git@host/owner/repo". ok=false si no parsea a
// host/owner/repo — el caller conserva el crudo como dato visible (jamás lo descarta).
func CanonicalizarRepo(ref string) (canon string, ok bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", false
	}

	var rest string
	switch {
	case strings.HasPrefix(ref, "https://"):
		rest = strings.TrimPrefix(ref, "https://")
	case strings.HasPrefix(ref, "http://"):
		rest = strings.TrimPrefix(ref, "http://")
	case strings.HasPrefix(ref, "ssh://"):
		rest = strings.TrimPrefix(ref, "ssh://")
		if i := strings.Index(rest, "@"); i >= 0 {
			rest = rest[i+1:] // ssh://git@host/owner/repo — el usuario no es parte de la identidad.
		}
	case strings.HasPrefix(ref, "git@"):
		rest = strings.Replace(strings.TrimPrefix(ref, "git@"), ":", "/", 1)
	default:
		if strings.Contains(ref, "://") {
			return "", false // esquema desconocido: no inventamos un host.
		}
		rest = "github.com/" + ref // "owner/repo" corto: asume github.com.
	}

	rest = strings.TrimSuffix(rest, "/")
	rest = strings.TrimSuffix(rest, ".git")
	rest = strings.ToLower(rest)

	parts := strings.Split(rest, "/")
	if len(parts) < 3 {
		return "", false // sin host+owner+repo completos, no es una referencia resoluble.
	}
	if parts[0] == "" || !strings.Contains(parts[0], ".") {
		return "", false // host debe parecer un host (tiene un punto).
	}
	for _, p := range parts[1:] {
		if p == "" {
			return "", false
		}
	}
	return strings.Join(parts, "/"), true
}
