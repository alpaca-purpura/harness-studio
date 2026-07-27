package main

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TestRutasDelRegistro — con `--sessions` explícito, esa ruta ES el registro vigente y no
// hay legado que migrar: el operador está apuntando a un archivo suyo a propósito (una
// copia, otro perfil, un test). Sin el flag, el par por defecto.
func TestRutasDelRegistro(t *testing.T) {
	vigente, legado := rutasDelRegistro("/tmp/mio/sesiones.json")
	if vigente != "/tmp/mio/sesiones.json" || legado != "" {
		t.Errorf("con --sessions explícito = (%q,%q), quiero la ruta pedida y ningún legado", vigente, legado)
	}

	vigente, legado = rutasDelRegistro("")
	if !strings.HasSuffix(vigente, filepath.Join(".arnesia", "sesiones.json")) {
		t.Errorf("registro vigente = %q", vigente)
	}
	if !strings.HasSuffix(legado, filepath.Join(".arnesia", "sessions.json")) {
		t.Errorf("registro legado = %q — es el que NO se toca y hace gratis volver atrás", legado)
	}
}

// TestArranqueMigraYLoguea (RF-306 CA-3) — el arranque migra el registro de la versión
// anterior y DICE las cuatro cosas: qué migró, dónde quedó el respaldo, qué reparó y qué
// llaves movió. Un arranque silencioso es un fallo del ticket, no una virtud.
func TestArranqueMigraYLoguea(t *testing.T) {
	dir := t.TempDir()
	legado := filepath.Join(dir, "sessions.json")
	viejo := `[
	  {"id":"s25123a2c","arnes":"vitalia","cwd":"/proj/vitalia","claude_session_id":"cc-1",
	   "ctx_pct":68,"rotacion_pendiente":true,
	   "conv":[{"rol":"user","text":"arreglá el hook de sellado"},{"rol":"assistant","text":"listo"}]},
	  {"id":"s0fec7798","arnes":"vitalia","conv":[]}
	]`
	if err := os.WriteFile(legado, []byte(viejo), 0o600); err != nil {
		t.Fatal(err)
	}
	vigente := filepath.Join(dir, "sesiones.json")

	clave := func(idPelado, _ string) (string, bool, string) {
		if strings.Contains(idPelado, "~") {
			return "", false, "ya-calificada"
		}
		if idPelado == "vitalia" {
			return "sin-home~vitalia~vitalia", true, "resuelta-por-id"
		}
		return "", false, "sin-candidata"
	}

	reg, inf, err := store.AbrirRegistro(vigente, legado, "2607262100", clave)
	if err != nil {
		t.Fatal(err)
	}
	if !inf.Migro || inf.DesdeVersion != 1 || inf.RespaldoEn == "" {
		t.Fatalf("informe = %+v, quiero migró desde la 1 con respaldo", inf)
	}
	if len(inf.Recalibradas) != 2 {
		t.Fatalf("recalibradas = %d, quiero una fila por sesión", len(inf.Recalibradas))
	}

	// El log dice lo que pasó, con las rutas.
	var buf bytes.Buffer
	previo := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	loguearInforme(inf, vigente)
	slog.SetDefault(previo)

	salida := buf.String()
	for _, quiero := range []string{"migrado a la forma nueva", inf.RespaldoEn, "llave recalibrada", "s25123a2c"} {
		if !strings.Contains(salida, quiero) {
			t.Errorf("el log de arranque no dice %q:\n%s", quiero, salida)
		}
	}

	// Y lo que se leyó es lo que había: la migración no perdió los turnos.
	sesiones, err := reg.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(sesiones) != 2 {
		t.Fatalf("sesiones = %d, quiero 2", len(sesiones))
	}
	c, ok := sesiones[0].Activa()
	if !ok {
		t.Fatal("la migrada quedó sin conversación activa")
	}
	if c.NumTurnos() != 2 || c.CtxPct != 68 || !c.RotacionPendiente {
		t.Errorf("la conversación migrada perdió estado: %+v", c)
	}
	if sesiones[0].Arnes != "sin-home~vitalia~vitalia" {
		t.Errorf("la llave no se recalibró: %q", sesiones[0].Arnes)
	}

	// El registro de la versión anterior sigue intacto: volver atrás es gratis.
	quedo, err := os.ReadFile(legado) //nolint:gosec // ruta del propio test.
	if err != nil || string(quedo) != viejo {
		t.Error("el registro legado cambió: el arranque no puede tocarlo")
	}
}

// TestArranqueSinNovedadesNoLoguea — el Informe sólo trae lo que ocurrió, así que un
// arranque normal no ensucia el log con ruido que nadie va a leer.
func TestArranqueSinNovedadesNoLoguea(t *testing.T) {
	var buf bytes.Buffer
	previo := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	loguearInforme(store.Informe{}, "/tmp/sesiones.json")
	slog.SetDefault(previo)

	if buf.Len() != 0 {
		t.Errorf("un arranque sin novedades logueó:\n%s", buf.String())
	}
}

// TestResolverDeLlavesUsaElPortafolioReal — el resolvedor que el CLI cablea, contra
// entradas del Portafolio con la forma de verdad. Las tres decisiones posibles.
func TestResolverDeLlavesUsaElPortafolioReal(t *testing.T) {
	entradas := []domain.EntradaPortafolio{
		{
			Identidad: domain.IdentidadArnes{ID: "vitalia", Scope: "vitalia"},
			Canonico:  &domain.Canonico{Path: "/home/x/luana-vitalia/vitalia"},
		},
		{
			Identidad:     domain.IdentidadArnes{Home: "acme.dev", ID: "vitalia"},
			Instalaciones: []domain.Instalacion{{InstallPath: "/home/x/otro/proyecto"}},
		},
		{Identidad: domain.IdentidadArnes{ID: "solo-uno"}},
	}
	clave := resolverDeLlaves(entradas)

	t.Run("por cwd desambigua entre dos candidatas", func(t *testing.T) {
		got, ok, motivo := clave("vitalia", "/home/x/luana-vitalia/vitalia")
		if !ok || motivo != "resuelta-por-cwd" {
			t.Fatalf("= (%q,%v,%q), quiero resuelta por cwd", got, ok, motivo)
		}
		if got != "sin-home~vitalia~vitalia" {
			t.Errorf("clave = %q", got)
		}
	})

	t.Run("sin cwd y con dos candidatas no se decide", func(t *testing.T) {
		got, ok, motivo := clave("vitalia", "")
		if ok || !strings.HasPrefix(motivo, "ambigua") {
			t.Errorf("= (%q,%v,%q) — con dos candidatas y sin cwd NO se decide", got, ok, motivo)
		}
	})

	t.Run("una sola candidata resuelve por id", func(t *testing.T) {
		got, ok, motivo := clave("solo-uno", "")
		if !ok || motivo != "resuelta-por-id" || got == "" {
			t.Errorf("= (%q,%v,%q), quiero resuelta por id", got, ok, motivo)
		}
	})

	t.Run("una llave ya calificada no se toca", func(t *testing.T) {
		_, ok, motivo := clave("sin-home~vitalia~vitalia", "")
		if ok || motivo != "ya-calificada" {
			t.Errorf("motivo = %q, quiero ya-calificada — es lo que hace idempotente al re-key", motivo)
		}
	})

	t.Run("un id que el Portafolio no conoce se declara, no se inventa", func(t *testing.T) {
		_, ok, motivo := clave("fantasma", "/cualquier/lado")
		if ok || motivo != "sin-candidata" {
			t.Errorf("motivo = %q, quiero sin-candidata", motivo)
		}
	})
}

// TestArchivoDeArchivadasCambioDeNombre — el registro de sesiones archivadas dejó de
// llamarse «cerradas»: las conversaciones no se cierran, se desactivan y se retoman.
func TestArchivoDeArchivadasCambioDeNombre(t *testing.T) {
	if got := cerradasPathDefault("/tmp/x/sesiones.json"); got != "/tmp/x/sesiones-archivadas.json" {
		t.Errorf("= %q, quiero el archivo de archivadas junto al de vivas", got)
	}
}
