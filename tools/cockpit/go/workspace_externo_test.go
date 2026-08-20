package main

import (
	"testing"
)

// Fase 3 · la tabla de reenvío systemWorkspaces resuelve la key compuesta
// `empresa/sistema` → (workspace externo, sistema local). La PROYECCIÓN del
// portfolio (gap vs navegable) es cockpit-owned y se testea en esa célula;
// acá se cubre el lado devhub del contrato: enableSystemWorkspaces + resolveSistema.
func TestSystemWorkspacesReenvio(t *testing.T) {
	ext := mkBoard(t, "products") // mimics ~/Proyectos/prenter-harness con board en products/
	p := registryProject{
		Name:   "prenter",
		Path:   t.TempDir(), // shell sin docs/product
		Active: false,       // inactivo: igual entra a la tabla (se arma de TODOS los proyectos)
		Directorio: directorioMeta{
			Kind: "factory",
			Sistemas: []sistemaMeta{
				{Slug: "prenter-harness", Procedencia: "propio", Workspace: ext},
			},
		},
	}
	enableMultiMode([]registryProject{p})
	enableSystemWorkspaces([]registryProject{p})
	defer func() { multiProjects = nil; systemWorkspaces = map[string]sysWorkspace{} }()

	root, local, err := resolveSistema("prenter/prenter-harness")
	if err != nil || root != ext || local != "products" {
		t.Errorf("resolveSistema reenvío mal: root=%q local=%q err=%v (want root=%q local=products)", root, local, err, ext)
	}
}

// defaultSistemaFromEnv: env vacío o "cross-sistema" → "" (sin preferencia); valor real → tal cual.
func TestDefaultSistemaFromEnv(t *testing.T) {
	cases := []struct{ env, want string }{
		{"", ""},
		{"cross-sistema", ""},
		{"vitalia/platform", "vitalia/platform"},
	}
	for _, c := range cases {
		t.Setenv("DEFAULT_SISTEMA", c.env)
		if got := defaultSistemaFromEnv(); got != c.want {
			t.Errorf("DEFAULT_SISTEMA=%q → %q, want %q", c.env, got, c.want)
		}
	}
}
