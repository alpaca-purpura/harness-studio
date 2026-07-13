package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

func TestResolverOrigen(t *testing.T) {
	t.Run("lock gana sobre manifiesto para registry (F1)", func(t *testing.T) {
		o := domain.ResolverOrigen([]domain.EslabonOrigen{
			{Fuente: "manifiesto", Campo: "home", Valor: "github.com/alpacapurpura/prenter-marketplace"},
			{Fuente: "lock-devstudio", Campo: "registry", Valor: "github.com/alpacapurpura/prenter-marketplace"},
		})
		if o.Registry != "github.com/alpacapurpura/prenter-marketplace" {
			t.Errorf("registry = %q, quiero el del lock", o.Registry)
		}
	})

	t.Run("cc-plugins aporta version", func(t *testing.T) {
		o := domain.ResolverOrigen([]domain.EslabonOrigen{
			{Fuente: "cc-plugins", Campo: "version", Valor: "1.2.3"},
		})
		if o.Version != "1.2.3" {
			t.Errorf("version = %q, quiero 1.2.3", o.Version)
		}
	})

	t.Run("conflicto lock-vs-cc puebla discrepancias", func(t *testing.T) {
		o := domain.ResolverOrigen([]domain.EslabonOrigen{
			{Fuente: "lock-devstudio", Campo: "registry", Valor: "github.com/a/b"},
			{Fuente: "cc-plugins", Campo: "registry", Valor: "github.com/c/d"},
		})
		if len(o.Discrepancias) == 0 {
			t.Fatal("quiero al menos una discrepancia por el conflicto lock vs cc-plugins")
		}
		if o.Registry == "" {
			t.Error("aun en conflicto, debe elegir un ganador (rango, no silencio)")
		}
	})

	t.Run("manifiesto-home vs lock-registry puebla discrepancia informativa", func(t *testing.T) {
		o := domain.ResolverOrigen([]domain.EslabonOrigen{
			{Fuente: "manifiesto", Campo: "home", Valor: "github.com/alpacapurpura/prenter-marketplace"},
			{Fuente: "lock-devstudio", Campo: "registry", Valor: "github.com/mirror/prenter-marketplace"},
		})
		if len(o.Discrepancias) == 0 {
			t.Fatal("quiero una discrepancia visible cuando home≠registry (nunca elección silenciosa)")
		}
		if o.Registry != "github.com/mirror/prenter-marketplace" {
			t.Errorf("registry sigue siendo el de mayor autoridad (lock), got %q", o.Registry)
		}
	})

	t.Run("solo git-proyecto: registry sigue vacío (BR-5/C-OR-4)", func(t *testing.T) {
		o := domain.ResolverOrigen([]domain.EslabonOrigen{
			{Fuente: "git-proyecto", Campo: "proyecto-remote", Valor: "github.com/usuario/su-proyecto"},
		})
		if o.Registry != "" {
			t.Errorf("registry = %q, el remote del PROYECTO nunca es el registry del arnés (BR-5)", o.Registry)
		}
	})

	t.Run("nada resuelve: todo vacío honesto (C-OR-5)", func(t *testing.T) {
		o := domain.ResolverOrigen(nil)
		if o.Registry != "" || o.Version != "" || len(o.Discrepancias) != 0 {
			t.Errorf("esperaba Origen{} vacío honesto, got %+v", o)
		}
	})
}

func TestResolverIdentidad(t *testing.T) {
	t.Run("con marketplace resuelve (home,id)", func(t *testing.T) {
		a := &domain.Arnes{ID: "harness", Marketplace: "alpacapurpura/prenter-marketplace"}
		id, aviso := domain.ResolverIdentidad(a, "", "", "")
		if id.Provisional() {
			t.Fatal("con marketplace resoluble, la identidad NO debe ser provisional")
		}
		if id.Home != "github.com/alpacapurpura/prenter-marketplace" || id.ID != "harness" {
			t.Errorf("identidad = %+v", id)
		}
		if aviso != "" {
			t.Errorf("aviso inesperado: %q", aviso)
		}
	})

	t.Run("sin marketplace: provisional con scope (RN-IDENT-2)", func(t *testing.T) {
		id, _ := domain.ResolverIdentidad(nil, "mi-arnes", "proyectos/x", "")
		if !id.Provisional() {
			t.Fatal("sin home, la identidad debe ser provisional")
		}
		if id.ID != "mi-arnes" || id.Scope != "proyectos/x" {
			t.Errorf("identidad provisional = %+v", id)
		}
	})

	t.Run("scope remoto manda sobre scope local", func(t *testing.T) {
		id, _ := domain.ResolverIdentidad(nil, "x", "local/scope", "github.com/usuario/proyecto")
		if id.Scope != "github.com/usuario/proyecto" {
			t.Errorf("scope = %q, quiero el remoto", id.Scope)
		}
	})

	t.Run("precedencia id RN-IDENT-3 + aviso en discrepancia", func(t *testing.T) {
		a := &domain.Arnes{ID: "id-del-manifiesto"}
		id, aviso := domain.ResolverIdentidad(a, "id-del-plugin-json", "", "")
		if id.ID != "id-del-manifiesto" {
			t.Errorf("id = %q, arnes.l0.id debe ganar", id.ID)
		}
		if aviso == "" {
			t.Error("quiero un aviso visible ante la discrepancia de ids, no elección silenciosa")
		}
	})

	t.Run("Clave estable y sin '/'", func(t *testing.T) {
		id := domain.IdentidadArnes{Home: "github.com/alpacapurpura/prenter-marketplace", ID: "harness"}
		clave := id.Clave()
		if strings.Contains(clave, "/") {
			t.Errorf("Clave() = %q contiene '/'", clave)
		}
		if clave != id.Clave() {
			t.Error("Clave() no es determinista")
		}
	})

	t.Run("Clave de identidad provisional no colisiona con vacío", func(t *testing.T) {
		id := domain.IdentidadArnes{ID: "mi-arnes", Scope: "proyectos/x"}
		clave := id.Clave()
		if !strings.HasPrefix(clave, "sin-home~") {
			t.Errorf("Clave() = %q, quiero prefijo sin-home~ para identidad provisional", clave)
		}
	})
}

func TestArnesEmpresasTolerante(t *testing.T) {
	t.Run("unmarshal legacy escalar empresa", func(t *testing.T) {
		var a domain.Arnes
		if err := json.Unmarshal([]byte(`{"id":"x","empresa":"alpacapurpura"}`), &a); err != nil {
			t.Fatal(err)
		}
		if len(a.Empresas) != 1 || a.Empresas[0] != "alpacapurpura" {
			t.Errorf("Empresas = %v, quiero [alpacapurpura]", a.Empresas)
		}
	})

	t.Run("unmarshal forma nueva empresas[]", func(t *testing.T) {
		var a domain.Arnes
		if err := json.Unmarshal([]byte(`{"id":"x","empresas":["a","b"]}`), &a); err != nil {
			t.Fatal(err)
		}
		if len(a.Empresas) != 2 || a.Empresas[0] != "a" || a.Empresas[1] != "b" {
			t.Errorf("Empresas = %v, quiero [a b]", a.Empresas)
		}
	})

	t.Run("marshal emite solo empresas", func(t *testing.T) {
		a := domain.Arnes{ID: "x", Empresas: []string{"alpacapurpura"}}
		b, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"empresas":["alpacapurpura"]`) {
			t.Fatalf("marshal no emitió empresas[]: %s", b)
		}
		if strings.Contains(string(b), `"empresa":`) {
			t.Errorf("marshal emitió el escalar legacy, quiero SOLO empresas: %s", b)
		}
	})
}
