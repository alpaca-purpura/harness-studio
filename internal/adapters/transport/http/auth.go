package httpapi

import (
	"crypto/subtle"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
)

// auth.go confines the daemon's local API surface (boundary superficie-local-confinada,
// HS-06). The daemon binds loopback, but any local process — including a web page in the
// user's browser — can still reach 127.0.0.1. Since a turn drives a Claude Code agent with
// filesystem access, an unauthenticated open surface is effectively remote-ish RCE. Three
// layered gates defend it:
//
//  1. Host gate  — the Host header must be an expected loopback host (anti DNS-rebinding).
//  2. Origin gate — a browser Origin must be on the allowlist; replaces CORS `*` with a
//     reflected, single-origin ACAO (anti cross-site CSRF).
//  3. Token gate — when a capability token is configured (the Tauri shell mints one and
//     passes it via env), every request must carry it. Disabled in dev (no token) where
//     gates 1+2 still apply.

// AuthConfig parameterizes the middleware. An empty Token disables the token gate (dev).
type AuthConfig struct {
	Token          string
	AllowedHosts   []string
	AllowedOrigins []string
	// TokenIngesta es el SEGUNDO token (decisión A6, paquete de telemetría). Abre **solo**
	// los tres endpoints de ingesta. Filtrarlo concede «podés escribirme telemetría basura»
	// —que la atribución y el rate limit ya acotan—, nunca «podés dirigir un agente con
	// acceso al filesystem», que es lo que el token de la API concede.
	TokenIngesta string
	// IngestaTokenObligatorio es la ESCOTILLA de A22. Por default `/v1/*` acepta sin token
	// bajo el Host gate loopback, porque Claude Code **no expande `${VAR}` en el bloque
	// `env`** (verificado, ANEXO H10.2) y meter el token literal en un archivo versionable
	// sería publicar un secreto. Encenderla exige token también ahí — y la consecuencia
	// honesta, que la UI dice, es que `s2-instrumentado` deja de reportar.
	IngestaTokenObligatorio bool
}

// AuthConfigFor builds an AuthConfig for a daemon listening on addr, deriving the loopback
// host + origin allowlists from its port. token may be empty (dev mode).
func AuthConfigFor(addr, token string) AuthConfig {
	return AuthConfigConIngesta(addr, token, "", false)
}

// AuthConfigConIngesta arma la config con el segundo token (A6) y la escotilla de A22.
func AuthConfigConIngesta(addr, token, tokenIngesta string, ingestaEstricta bool) AuthConfig {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		port = "4200"
	}
	return AuthConfig{
		Token:                   token,
		TokenIngesta:            tokenIngesta,
		IngestaTokenObligatorio: ingestaEstricta,
		AllowedHosts: []string{
			"127.0.0.1:" + port,
			"localhost:" + port,
			"[::1]:" + port,
		},
		AllowedOrigins: []string{
			// Tauri WebView origins (platform-dependent scheme).
			"tauri://localhost",
			"http://tauri.localhost",
			"https://tauri.localhost",
			// Vite dev server.
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			// The daemon's own origin (embedded UI / same-host tools).
			"http://localhost:" + port,
			"http://127.0.0.1:" + port,
		},
	}
}

var noTokenWarn sync.Once

// withAuth wraps next with the three-gate confinement. /healthz is exempt (pure liveness,
// no data, no side effects) so external probes work without a token.
func withAuth(cfg AuthConfig, next http.Handler) http.Handler {
	hosts := toSet(cfg.AllowedHosts)
	origins := toSet(cfg.AllowedOrigins)
	if cfg.Token == "" {
		noTokenWarn.Do(func() {
			slog.Warn("arnesia: no auth token set — API confined by Host+Origin only (dev). " +
				"Production sets ARNESIA_AUTH_TOKEN / --auth-token (the Tauri shell does this).")
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			// Sin Host/token gate (liveness pura), pero SIGUE necesitando el reflejo CORS: el
			// WebView vive en su propio origin (tauri://localhost / http://tauri.localhost),
			// distinto de 127.0.0.1:4200. Sin este header, un fetch() cross-origin del shell
			// ve /healthz como error de red pese a que el daemon respondió 200 — regresión
			// real encontrada en vivo (HS-14): el early-return original saltaba esto entero.
			if origin := r.Header.Get("Origin"); origin != "" && origins[origin] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
			}
			next.ServeHTTP(w, r)
			return
		}

		// 1. Host gate — anti DNS-rebinding.
		if !hosts[r.Host] {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}

		// 2. Origin gate — anti cross-site CSRF. Only browsers send Origin; a present one
		// must be allowlisted, and we reflect exactly it (never `*`).
		if origin := r.Header.Get("Origin"); origin != "" {
			if !origins[origin] {
				http.Error(w, "forbidden origin", http.StatusForbidden)
				return
			}
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Arnesia-Token, Last-Event-ID")
		}

		// Preflight is answered after the Origin gate (already CORS-decorated) and before
		// the token gate — browsers never send the token on a preflight.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 3. Token gate — capability. Skipped when no token is configured (dev). Aplica a
		// la API y al stream; la UI estática embebida (assets inertes en "/") queda tras
		// los gates 1+2 solamente — un browser debe poder CARGARLA antes de tener token
		// (el token viaja luego en cada llamada de la SPA a /api).
		//
		// Las rutas de INGESTA tienen su propio gate (A6/A22): el token acotado abre esas
		// tres y **ninguna más**, y `/v1/*` acepta sin token bajo el Host gate loopback
		// salvo que el operador encienda la escotilla.
		if isRutaIngesta(r.URL.Path) {
			if !validaIngesta(r, cfg) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		if cfg.Token != "" && isAPIPath(r.URL.Path) && !validToken(r, cfg.Token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAPIPath reports whether the path carries capability (API o stream) — lo que el
// token protege. Todo lo demás bajo el mux es la SPA embebida (estática).
//
// `/v1/` entra acá para que el receptor OTLP quede DENTRO del confinamiento (Host+Origin);
// su gate de token es el suyo propio, en `validaIngesta`.
func isAPIPath(p string) bool {
	return strings.HasPrefix(p, "/api") || p == "/events" || strings.HasPrefix(p, "/v1/")
}

// rutasIngesta son las TRES rutas que el token acotado abre, y ninguna más.
var rutasIngesta = map[string]bool{
	"/v1/logs":                true,
	"/v1/metrics":             true,
	"/api/telemetria/proceso": true,
}

// isRutaIngesta reporta si la ruta es de ingesta de telemetría.
func isRutaIngesta(p string) bool { return rutasIngesta[p] }

// validaIngesta aplica el gate de las rutas de ingesta.
//
// **A22, y conviene leerlo dos veces:** `/v1/logs` y `/v1/metrics` aceptan **sin token** bajo
// el Host gate loopback. No es un descuido — es que Claude Code **no expande `${VAR}` dentro
// del bloque `env`** (medido, ANEXO H10.2), así que en `s2-instrumentado` el token no puede
// llegar por indirección, y ponerlo literal metería un secreto vivo en un archivo versionable
// que iría a git y se publicaría con el paquete.
//
// Lo que se pierde está acotado por diseño previo: sin token, un proceso local del MISMO
// usuario puede inyectar ruido contable, y nada más. Y **el ruido no contamina ningún
// número**: entra sin atribución, no suma a ningún total, y es VISIBLE en la cobertura y en
// la salud. Quedan tres barreras: Host gate, tope de cuerpo y rate limit.
//
// `POST /api/telemetria/proceso` exige token SIEMPRE, con o sin escotilla: el hook lee la
// ficha `0600` en runtime, así que ahí sí puede llevarlo.
func validaIngesta(r *http.Request, cfg AuthConfig) bool {
	esOTLP := strings.HasPrefix(r.URL.Path, "/v1/")
	if esOTLP && !cfg.IngestaTokenObligatorio {
		return true // Host gate loopback ya se aplicó arriba.
	}
	if cfg.TokenIngesta == "" && cfg.Token == "" {
		return true // dev sin ningún token configurado.
	}
	// El token de INGESTA abre estas rutas; el de la API también, porque el shell ya tiene
	// permiso de todo y obligarlo a llevar dos tokens no agrega ninguna garantía.
	if cfg.TokenIngesta != "" && validToken(r, cfg.TokenIngesta) {
		return true
	}
	return cfg.Token != "" && validToken(r, cfg.Token)
}

// validToken reads the token from a header (Authorization: Bearer / X-Arnesia-Token) or,
// as a fallback for the SSE EventSource which cannot set headers, the `token` query param.
// The compare is constant-time.
func validToken(r *http.Request, want string) bool {
	var got string
	switch {
	case strings.HasPrefix(r.Header.Get("Authorization"), "Bearer "):
		got = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	case r.Header.Get("X-Arnesia-Token") != "":
		got = r.Header.Get("X-Arnesia-Token")
	default:
		got = r.URL.Query().Get("token")
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}
