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
}

// AuthConfigFor builds an AuthConfig for a daemon listening on addr, deriving the loopback
// host + origin allowlists from its port. token may be empty (dev mode).
func AuthConfigFor(addr, token string) AuthConfig {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		port = "4200"
	}
	return AuthConfig{
		Token: token,
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

		// 3. Token gate — capability. Skipped when no token is configured (dev).
		if cfg.Token != "" && !validToken(r, cfg.Token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
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
