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
		// IDEMPOTENCIA (paquete 2026-07-23-portafolio-agregar-marketplace): RN-IDENT-1 exige
		// «canonicalizar los DOS lados antes de comparar», y para que eso funcione la función
		// tiene que ser idempotente. Un ref YA canónico ("host/owner/repo", ≥3 segmentos con el
		// primero pareciendo un host) no vuelve a recibir el prefijo `github.com/`: antes
		// `CanonicalizarRepo("github.com/a/b")` daba `github.com/github.com/a/b`, así que
		// comparar un valor ya canónico contra un crudo canonicalizado NUNCA coincidía.
		if yaCanonico(ref) {
			rest = ref
		} else {
			rest = "github.com/" + ref // "owner/repo" corto: asume github.com.
		}
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

// yaCanonico reporta si ref ya tiene la forma "host/owner/repo" (sin esquema): ≥3 segmentos y
// el primero parece un host (contiene un punto y no está vacío). Sostiene la idempotencia de
// CanonicalizarRepo — ver el comentario en el `default` del switch.
func yaCanonico(ref string) bool {
	partes := strings.Split(strings.TrimSuffix(strings.TrimSuffix(ref, "/"), ".git"), "/")
	if len(partes) < 3 {
		return false
	}
	return partes[0] != "" && strings.Contains(partes[0], ".")
}
