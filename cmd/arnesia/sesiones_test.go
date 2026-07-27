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

// TestElArranqueCablearRecalibracionDeLlaves — CV-D18 / A-10, en el composition root.
//
// El defecto que cierra no era de lógica: era de CABLEADO. `RecalibrarAlArrancar` puede
// estar perfecta y con todos sus tests verdes, y si `runServe` no la llama, CV-D16 sigue
// sin construirse — que es exactamente lo que pasó (se pasaba `nil` y nadie se enteraba).
// Un test de unidad del store no puede ver eso; este mira el composition root.
//
// Se afirma sobre el FUENTE de `main.go` a propósito: `runServe` abre puertos, spawnea
// watchers y no vuelve, así que no se lo puede invocar en un test. El precedente del repo
// para esta clase de check es `docs/architecture/fitness/arch_test.go`, que ya afirma
// propiedades leyendo el árbol con go/parser.
func TestElArranqueCablearRecalibracionDeLlaves(t *testing.T) {
	b, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)

	if !strings.Contains(src, "RecalibrarAlArrancar") {
		t.Fatal("el arranque NO llama a RecalibrarAlArrancar: CV-D16 vuelve a estar firmada y sin construir (A-10)")
	}
	// Y le pasa el resolvedor REAL, no un nil ni un stub: sin Portafolio no hay con qué decidir.
	if !strings.Contains(src, "RecalibrarAlArrancar(resolverDeLlaves(entradas)") {
		t.Error("la recalibración del arranque no recibe el resolvedor construido desde el Portafolio")
	}
	// Y lo que hizo sale por el log: «nada en silencio» es la mitad de la decisión. Se busca
	// la LLAMADA con su argumento, no el nombre a secas: `loguearRecalibracion(` también
	// aparece en la línea que la DEFINE, así que buscar eso daba verde con la llamada
	// borrada — verificado por mutación, y fue este test el que estaba flojo.
	if !strings.Contains(src, "loguearRecalibracion(infRekey") {
		t.Error("el arranque recalibra y no loguea: CV-D18 exige decir cuántas movió y cuántas quedaron sin candidata")
	}
}

// TestLoguearRecalibracionDiceLasDosCosas — el informe del arranque distingue recalibradas
// de `sin-candidata`. La segunda es el caso que el E2E ya conoce (una de las 5 sesiones del
// operador) y que CV-D18 prohíbe callar.
func TestLoguearRecalibracionDiceLasDosCosas(t *testing.T) {
	previo := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previo) })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))

	loguearRecalibracion(store.InformeRecalibracion{
		Recalibradas: 2, SinCandidata: 1, YaCalificadas: 1, RespaldoEn: "/tmp/sessions.json.bak-x",
		Filas: []store.Recalibracion{
			{SesionID: "s1", Antes: "vitalia", Despues: "sin-home~vitalia~vitalia", Motivo: "resuelta-por-id"},
			{SesionID: "s2", Antes: "fantasma", Despues: "fantasma", Motivo: "sin-candidata"},
		},
	}, "/tmp/sessions.json")

	s := buf.String()
	for _, quiero := range []string{
		"recalibradas=2", "sin_candidata=1", "/tmp/sessions.json.bak-x",
		"s1", "sin-home~vitalia~vitalia", "s2",
		"no está en el Portafolio", "--revertir",
	} {
		if !strings.Contains(s, quiero) {
			t.Errorf("el log del arranque no dice %q.\nlog:\n%s", quiero, s)
		}
	}
}

// TestLoguearRecalibracionSinNovedadNoImprime — un arranque que no movió nada y no tiene
// ninguna sesión sin candidata no ensucia el log. Mismo criterio que `loguearInforme`.
func TestLoguearRecalibracionSinNovedadNoImprime(t *testing.T) {
	previo := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previo) })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	loguearRecalibracion(store.InformeRecalibracion{YaCalificadas: 5}, "/tmp/sessions.json")

	if s := strings.TrimSpace(buf.String()); s != "" {
		t.Errorf("un arranque sin novedad logueó igual:\n%s", s)
	}
}
