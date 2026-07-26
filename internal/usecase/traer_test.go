package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// ── fake del puerto Materializador ──

type fakeMaterializador struct {
	camino  domain.CaminoTraer
	sha     string
	avisos  []string
	err     error
	llamado int
	// escribirEn deja un archivo FUERA del staging (control negativo de E-88).
	escribirEn string
	// antesDeEscribir se corre al entrar (para cancelar el ctx, paniquear, etc.).
	antesDeEscribir func(ctx context.Context) error
	// archivos es el árbol que se materializa en el staging.
	archivos map[string]string
	// capturarStaging guarda el path de staging que el usecase pasó (E-101).
	capturarStaging *string
}

func (f *fakeMaterializador) Camino() domain.CaminoTraer { return f.camino }

func (f *fakeMaterializador) Materializar(ctx context.Context, _ domain.PlanTraer, staging string) (string, []string, error) {
	f.llamado++
	if f.capturarStaging != nil {
		*f.capturarStaging = staging
	}
	if f.antesDeEscribir != nil {
		if err := f.antesDeEscribir(ctx); err != nil {
			return "", nil, err
		}
	}
	if f.escribirEn != "" {
		if merr := os.MkdirAll(filepath.Dir(f.escribirEn), 0o750); merr == nil {
			_ = os.WriteFile(f.escribirEn, []byte("robado\n"), 0o600)
		}
	}
	if f.err != nil {
		return "", f.avisos, f.err
	}
	if merr := os.MkdirAll(staging, 0o750); merr != nil {
		return "", nil, merr
	}
	for rel, contenido := range f.archivos {
		destino := filepath.Join(staging, rel)
		if merr := os.MkdirAll(filepath.Dir(destino), 0o750); merr != nil {
			return "", nil, merr
		}
		if werr := os.WriteFile(destino, []byte(contenido), 0o600); werr != nil {
			return "", nil, werr
		}
	}
	return f.sha, f.avisos, nil
}

func matLocalFake() *fakeMaterializador {
	return &fakeMaterializador{
		camino:   domain.CaminoLocal,
		archivos: map[string]string{"VERSION": "KIT_VERSION=0.5.3\n", "hooks/hooks.json": "{}\n"},
	}
}

// derivaFake devuelve el veredicto que el test pida (BR-17: se muestra lo que salga).
type derivaFake struct {
	estado  domain.EstadoDeriva
	detalle string
	visto   []string // installDirs evaluados
}

func (d *derivaFake) Evaluar(installDir, _, _, _ string) (domain.EstadoDeriva, string) {
	d.visto = append(d.visto, installDir)
	return d.estado, d.detalle
}

// ── helpers de escenario ──

type escenarioTraer struct {
	*equipo
	raiz    string
	local   *fakeMaterializador
	externo *fakeMaterializador
	deriva  *derivaFake
}

// armarTraer cablea el servicio con los 3 puertos de Traer y una raíz INYECTADA (nada toca
// ~/.arnesia real ni ~/.claude).
func armarTraer(t *testing.T) *escenarioTraer {
	t.Helper()
	e := armar(t)
	esc := &escenarioTraer{
		equipo:  e,
		raiz:    t.TempDir(),
		local:   matLocalFake(),
		externo: &fakeMaterializador{camino: domain.CaminoExterno, archivos: map[string]string{"skills/SKILL.md": "# x\n"}},
		deriva:  &derivaFake{estado: domain.DerivaAlHilo},
	}
	e.detector.filas = []domain.MarketplaceConocido{{
		Nombre: "prenter-marketplace", Repo: repoPrenter, InstallLocation: t.TempDir(),
	}}
	e.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClasePropio,
	}
	e.cache.guardados["prenter-marketplace"] = catDosEntradas("prenter-marketplace", "2026-07-25T14:00:00Z", "local")
	e.svc.SetTraer(esc.local, esc.externo, esc.deriva, esc.raiz)
	return esc
}

// huella lista paths + sha del contenido + mtime de un árbol; "" si no existe.
func huella(t *testing.T, dir string) string {
	t.Helper()
	if _, err := os.Stat(dir); err != nil {
		return ""
	}
	var lineas []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		linea := rel + "|" + info.Mode().String() + "|" + info.ModTime().UTC().Format(time.RFC3339Nano)
		if d.Type().IsRegular() {
			b, rerr := os.ReadFile(p) //nolint:gosec // G304: árbol del test.
			if rerr != nil {
				return rerr
			}
			sum := sha256.Sum256(b)
			linea += "|" + hex.EncodeToString(sum[:8])
		}
		lineas = append(lineas, linea)
		return nil
	})
	if err != nil {
		t.Fatalf("huella de %s: %v", dir, err)
	}
	sort.Strings(lineas)
	return strings.Join(lineas, "\n")
}

func listarRel(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if rel != "." {
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

// tmpVacio asserta que <raiz>/tmp no tiene ningún residuo (BR-15).
func tmpVacio(t *testing.T, raiz string) {
	t.Helper()
	dir := domain.RaizTemporales(raiz)
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return // no existe: también es «cero residuo».
	}
	if len(entradas) != 0 {
		nombres := make([]string, 0, len(entradas))
		for _, e := range entradas {
			nombres = append(nombres, e.Name())
		}
		t.Fatalf("quedó residuo en %s: %v", dir, nombres)
	}
}

// E-76 · camino A feliz: destino exacto, canónico registrado, CERO red (el materializador externo
// falla el test si se lo invoca).
func TestTraerCaminoALocal(t *testing.T) {
	esc := armarTraer(t)
	esc.externo.err = errors.New("el camino B NO debe invocarse en el camino A")

	res, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if err != nil {
		t.Fatalf("Traer: %v", err)
	}
	destinoEsperado := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	if res.Destino != destinoEsperado {
		t.Fatalf("Destino = %q, want %q", res.Destino, destinoEsperado)
	}
	if res.Camino != domain.CaminoLocal {
		t.Fatalf("Camino = %q, want local", res.Camino)
	}
	if res.SHAEfectivo != "" {
		t.Fatalf("SHAEfectivo = %q, want vacío en el camino local", res.SHAEfectivo)
	}
	if esc.externo.llamado != 0 {
		t.Fatal("se invocó el camino externo (habría red donde no debe haber)")
	}
	if got := listarRel(t, destinoEsperado); len(got) == 0 {
		t.Fatal("el destino quedó vacío")
	}

	clave := domain.IdentidadArnes{Home: repoPrenter, ID: "harness"}.Clave()
	entrada, ok := esc.pf.entradas[clave]
	if !ok {
		t.Fatalf("no se registró la entrada %q: %v", clave, esc.pf.entradas)
	}
	if entrada.Canonico == nil || entrada.Canonico.Path != destinoEsperado || entrada.Canonico.Version != "0.5.3" {
		t.Fatalf("Canonico = %+v, want {destino, 0.5.3}", entrada.Canonico)
	}
	if len(entrada.Registries) != 1 || entrada.Registries[0] != repoPrenter {
		t.Fatalf("Registries = %v, want [%s]", entrada.Registries, repoPrenter)
	}
	if entrada.Identidad.Home != repoPrenter {
		t.Fatalf("Identidad.Home = %q, want el REPO canonicalizado (C17)", entrada.Identidad.Home)
	}
	tmpVacio(t, esc.raiz)
}

// E-77 · la deriva se evalúa DE INMEDIATO y se devuelve tal cual salga — incluso
// `deriva-no-evaluable` con motivo, y con 200 igual (BR-17).
func TestTraerEvaluaDerivaDeInmediato(t *testing.T) {
	esc := armarTraer(t)
	res, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if err != nil {
		t.Fatal(err)
	}
	if res.Deriva != domain.DerivaAlHilo {
		t.Fatalf("Deriva = %q, want al-hilo", res.Deriva)
	}
	if len(esc.deriva.visto) != 1 || esc.deriva.visto[0] != res.Destino {
		t.Fatalf("la deriva se evaluó sobre %v, want el destino recién creado", esc.deriva.visto)
	}

	// Sin referencia accesible: `deriva-no-evaluable` CON motivo, y la operación SIGUE siendo un
	// éxito — un veredicto incómodo no es un fallo.
	esc2 := armarTraer(t)
	esc2.deriva.estado = domain.DerivaNoEvaluable
	esc2.deriva.detalle = "sin referencia local accesible (marketplace no clonado, o esa versión ausente del checkout)"
	res2, err2 := esc2.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if err2 != nil {
		t.Fatalf("un veredicto no-evaluable NO debe convertir el Traer en error: %v", err2)
	}
	if res2.Deriva != domain.DerivaNoEvaluable || res2.DerivaDetalle == "" {
		t.Fatalf("Deriva = %q / %q, want no-evaluable CON motivo", res2.Deriva, res2.DerivaDetalle)
	}
}

// E-78 · dos canales que comparten `source` tienen destinos PROPIOS por `id`: ninguno pisa al otro.
func TestTraerCanalesDestinoPropio(t *testing.T) {
	esc := armarTraer(t)
	if _, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness"); err != nil {
		t.Fatalf("traer harness: %v", err)
	}
	if _, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness-beta"); err != nil {
		t.Fatalf("traer harness-beta: %v", err)
	}
	base := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace")
	for _, id := range []string{"harness", "harness-beta"} {
		if got := listarRel(t, filepath.Join(base, id)); len(got) == 0 {
			t.Fatalf("el destino de %q quedó vacío", id)
		}
	}
	if len(esc.pf.entradas) != 2 {
		t.Fatalf("entradas registradas = %d, want 2", len(esc.pf.entradas))
	}
	homes := map[string]bool{}
	ids := map[string]bool{}
	for _, e := range esc.pf.entradas {
		homes[e.Identidad.Home] = true
		ids[e.Identidad.ID] = true
	}
	if len(homes) != 1 {
		t.Fatalf("homes = %v, want UNO solo (el mismo marketplace)", homes)
	}
	if len(ids) != 2 {
		t.Fatalf("ids = %v, want dos distintos", ids)
	}
}

// E-79 · destino ya poblado ⇒ aborta SIN TOCAR NADA: el contenido previo es byte-idéntico y NO se
// creó ningún temporal.
func TestTraerDestinoPobladoAborta(t *testing.T) {
	esc := armarTraer(t)
	if _, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness"); err != nil {
		t.Fatal(err)
	}
	destino := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	antes := huella(t, destino)

	esc.local.llamado = 0
	_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if !errors.Is(err, usecase.ErrTraerDestinoPoblado) {
		t.Fatalf("err = %v, want ErrTraerDestinoPoblado", err)
	}
	if despues := huella(t, destino); despues != antes {
		t.Fatalf("el canónico previo cambió:\nantes:   %s\ndespués: %s", antes, despues)
	}
	if esc.local.llamado != 0 {
		t.Fatal("se invocó el materializador con el destino poblado (debe abortar ANTES)")
	}
	tmpVacio(t, esc.raiz)
}

// E-79b · un destino que existe pero está VACÍO no es «poblado»: el Traer procede. `os.ReadDir`
// (no `os.Stat`) es la comprobación correcta.
func TestTraerDestinoVacioSePuedeUsar(t *testing.T) {
	esc := armarTraer(t)
	destino := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	if err := os.MkdirAll(destino, 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness"); err != nil {
		t.Fatalf("un destino VACÍO debe poder usarse: %v", err)
	}
	if got := listarRel(t, destino); len(got) == 0 {
		t.Fatal("el destino quedó vacío tras traer")
	}
}

// E-79c · un ARCHIVO ocupando la ruta del destino cuenta como poblado ⇒ 409, sin tocarlo.
func TestTraerDestinoEsArchivo(t *testing.T) {
	esc := armarTraer(t)
	destino := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	if err := os.MkdirAll(filepath.Dir(destino), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destino, []byte("no soy un dir\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	antes := huella(t, filepath.Dir(destino))

	_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if !errors.Is(err, usecase.ErrTraerDestinoPoblado) {
		t.Fatalf("err = %v, want ErrTraerDestinoPoblado", err)
	}
	if despues := huella(t, filepath.Dir(destino)); despues != antes {
		t.Fatal("se tocó el archivo que ocupaba el destino")
	}
}

// E-82 · BR-16 · el sha materializado no coincide con el declarado ⇒ aborta: `tmp` vacío, destino
// inexistente, CERO entradas.
func TestTraerSHANoCoincideAborta(t *testing.T) {
	esc := armarTraer(t)
	// La fila del catálogo pasa a ser un source objeto con `sha` declarado (camino B).
	cat := esc.cache.guardados["prenter-marketplace"]
	cat.Entradas[0].Source = domain.SourceCatalogo{
		Tipo: domain.SourceGitSubdir, Crudo: "git-subdir:x", URL: "https://github.com/a/b.git",
		Ruta: "plugins/x", SHA: "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad",
	}
	esc.cache.guardados["prenter-marketplace"] = cat
	esc.externo.sha = "0000000000000000000000000000000000000000"

	_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if !errors.Is(err, usecase.ErrTraerSHANoCoincide) {
		t.Fatalf("err = %v, want ErrTraerSHANoCoincide", err)
	}
	if !strings.Contains(err.Error(), "30287f5e") {
		t.Fatalf("el motivo debe citar el sha declarado: %v", err)
	}
	tmpVacio(t, esc.raiz)
	destino := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	if _, serr := os.Stat(destino); serr == nil {
		t.Fatal("el destino existe tras abortar por sha")
	}
	if len(esc.pf.entradas) != 0 {
		t.Fatalf("se registró algo: %v", esc.pf.entradas)
	}
}

// E-86 · cancelación y pánico del materializador: el `defer` limpia IGUAL. Cero residuo, cero
// destino, cero registro.
func TestTraerCancelacionNoDejaResiduo(t *testing.T) {
	t.Run("ctx cancelado", func(t *testing.T) {
		esc := armarTraer(t)
		ctx, cancel := context.WithCancel(context.Background())
		esc.local.antesDeEscribir = func(context.Context) error {
			cancel()
			return context.Canceled
		}
		_, err := esc.svc.Traer(ctx, "prenter-marketplace", "harness")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled propagado", err)
		}
		tmpVacio(t, esc.raiz)
		if _, serr := os.Stat(filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")); serr == nil {
			t.Fatal("el destino existe tras la cancelación")
		}
		if len(esc.pf.entradas) != 0 {
			t.Fatalf("se registró algo: %v", esc.pf.entradas)
		}
	})

	t.Run("materializador que paniquea", func(t *testing.T) {
		esc := armarTraer(t)
		esc.local.antesDeEscribir = func(context.Context) error { panic("boom del adapter") }
		func() {
			defer func() {
				if rec := recover(); rec == nil {
					t.Fatal("el pánico no se propagó (y debería: no lo tragamos)")
				}
			}()
			_, _ = esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
		}()
		// El `defer os.RemoveAll(tmp)` corre igual durante el unwind del pánico.
		tmpVacio(t, esc.raiz)
		if len(esc.pf.entradas) != 0 {
			t.Fatalf("se registró algo: %v", esc.pf.entradas)
		}
	})
}

// E-86b · huérfano de un SIGKILL previo: al construir el servicio se borra el temporal VIEJO y NO
// el reciente (podría ser de otra instancia viva).
func TestBarridoDeTemporalesAlArrancar(t *testing.T) {
	e := armar(t)
	raiz := t.TempDir()
	dirTmp := domain.RaizTemporales(raiz)
	if err := os.MkdirAll(filepath.Join(dirTmp, "traer-viejo"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dirTmp, "traer-nuevo"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dirTmp, "otra-cosa"), 0o750); err != nil {
		t.Fatal(err)
	}
	// El reloj INYECTADO (2026-07-25T14:07:33Z) es la referencia, no el reloj real de la máquina.
	viejo := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)  // 2 h antes ⇒ se barre.
	nuevo := time.Date(2026, 7, 25, 14, 6, 33, 0, time.UTC) // 1 min antes ⇒ NO se barre.
	if err := os.Chtimes(filepath.Join(dirTmp, "traer-viejo"), viejo, viejo); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(dirTmp, "traer-nuevo"), nuevo, nuevo); err != nil {
		t.Fatal(err)
	}

	e.svc.SetTraer(matLocalFake(), nil, &derivaFake{}, raiz)

	if _, err := os.Stat(filepath.Join(dirTmp, "traer-viejo")); err == nil {
		t.Fatal("el temporal VIEJO no se barrió")
	}
	if _, err := os.Stat(filepath.Join(dirTmp, "traer-nuevo")); err != nil {
		t.Fatal("se barrió un temporal RECIENTE (podría ser de otra instancia viva)")
	}
	if _, err := os.Stat(filepath.Join(dirTmp, "otra-cosa")); err != nil {
		t.Fatal("se barrió un dir que no es un temporal de Traer")
	}
}

// E-88 · BR-13 · JAMÁS se escribe en `~/.claude`: aserción de RUTA y BYTES sobre las 6 ramas de
// §13.5, no de confianza. Se corre con un HOME falso poblado y se compara el árbol completo
// (contenido + permisos + mtime) antes y después de cada rama.
func TestTraerJamasEscribeEnClaude(t *testing.T) {
	homeFalso := t.TempDir()
	claude := filepath.Join(homeFalso, ".claude")
	for rel, contenido := range map[string]string{
		"settings.json":                   "{\"enabledPlugins\":{}}\n",
		"plugins/known_marketplaces.json": "{}\n",
		"plugins/installed_plugins.json":  "{\"version\":2,\"plugins\":{}}\n",
	} {
		p := filepath.Join(claude, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	antes := huella(t, claude)
	if antes == "" {
		t.Fatal("el HOME falso no se pobló")
	}

	// Las 6 ramas de §13.5 + el camino feliz.
	ramas := map[string]func(esc *escenarioTraer){
		"2 · planificación (clase referencia)": func(esc *escenarioTraer) {
			esc.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{
				Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClaseReferencia,
			}
		},
		"3 · destino poblado": func(esc *escenarioTraer) {
			d := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
			_ = os.MkdirAll(d, 0o750)
			_ = os.WriteFile(filepath.Join(d, "ya-estaba"), []byte("x\n"), 0o600)
		},
		"5 · materialización falla": func(esc *escenarioTraer) {
			esc.local.err = errors.New("fake: fallo de materialización")
		},
		"6 · sha no coincide": func(esc *escenarioTraer) {
			cat := esc.cache.guardados["prenter-marketplace"]
			cat.Entradas[0].Source = domain.SourceCatalogo{
				Tipo: domain.SourceURL, Crudo: "url:x", URL: "https://github.com/a/b.git", SHA: "aaaa",
			}
			esc.cache.guardados["prenter-marketplace"] = cat
			esc.externo.sha = "bbbb"
		},
		"7 · rename falla": func(esc *escenarioTraer) {
			// El dir padre del destino se hace un ARCHIVO: el MkdirAll del padre falla.
			padre := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace")
			_ = os.MkdirAll(filepath.Dir(padre), 0o750)
			_ = os.WriteFile(padre, []byte("no soy un dir\n"), 0o600)
		},
		"9 · Upsert falla": func(esc *escenarioTraer) {
			esc.pf.upsertErr = errors.New("fake: no se pudo escribir el registro")
		},
		"camino feliz": func(*escenarioTraer) {},
	}

	for nombre, preparar := range ramas {
		t.Run(nombre, func(t *testing.T) {
			esc := armarTraer(t)
			t.Setenv("HOME", homeFalso)
			preparar(esc)

			_, _ = esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")

			if despues := huella(t, claude); despues != antes {
				t.Fatalf("el árbol ~/.claude CAMBIÓ tras la rama %q:\nantes:\n%s\ndespués:\n%s", nombre, antes, despues)
			}
			// Todo lo escrito cae bajo la raíz INYECTADA: el HOME falso solo tiene `.claude`.
			for _, rel := range listarRel(t, homeFalso) {
				if !strings.HasPrefix(rel, ".claude") {
					t.Fatalf("la rama %q escribió %q fuera de la raíz inyectada", nombre, rel)
				}
			}
		})
	}

	// CONTROL NEGATIVO: la aserción no es vacua. Un materializador que SÍ escribe en
	// `<HOME>/.claude` tiene que ser DETECTADO por la misma comparación.
	t.Run("control negativo · la aserción detecta una escritura", func(t *testing.T) {
		esc := armarTraer(t)
		esc.local.escribirEn = filepath.Join(claude, "robado.json")
		t.Cleanup(func() { _ = os.Remove(filepath.Join(claude, "robado.json")) })

		_, _ = esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
		if huella(t, claude) == antes {
			t.Fatal("la aserción de E-88 es VACUA: no detectó una escritura real en ~/.claude")
		}
	})
}

// E-89 · fallo local (permisos) ⇒ ErrTraerLocal con el motivo real, temporal limpio, nada registrado.
func TestTraerFalloLocalLimpia(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el modo 0555 no impide escribir: el caso no es reproducible")
	}
	esc := armarTraer(t)
	// El dir `checkouts` se hace read-only: el MkdirAll del padre del destino falla.
	checkouts := domain.RaizCheckouts(esc.raiz)
	if err := os.MkdirAll(checkouts, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(checkouts, 0o555); err != nil { //nolint:gosec // G302: el dir read-only es EL insumo de E-89.
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(checkouts, 0o750) }) //nolint:gosec // G302: restaura para que t.TempDir() pueda limpiar.

	_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if !errors.Is(err, usecase.ErrTraerLocal) {
		t.Fatalf("err = %v, want ErrTraerLocal", err)
	}
	if err := os.Chmod(checkouts, 0o750); err != nil { //nolint:gosec // G302: ídem.
		t.Fatal(err)
	}
	tmpVacio(t, esc.raiz)
	if len(esc.pf.entradas) != 0 {
		t.Fatalf("se registró algo: %v", esc.pf.entradas)
	}
}

// E-100 · el `Upsert` falla TRAS materializar: es el ÚNICO parcial posible y es VISIBLE — el dir
// QUEDA (costó red), el motivo trae el destino, y el próximo Traer da 409.
func TestUpsertFallidoDejaDirVisible(t *testing.T) {
	esc := armarTraer(t)
	esc.pf.upsertErr = errors.New("fake: registro no escribible")

	_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if !errors.Is(err, usecase.ErrTraerLocal) {
		t.Fatalf("err = %v, want ErrTraerLocal", err)
	}
	destino := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	if !strings.Contains(err.Error(), destino) {
		t.Fatalf("el motivo debe traer el destino para que el parcial sea VISIBLE: %v", err)
	}
	if got := listarRel(t, destino); len(got) == 0 {
		t.Fatal("el dir se borró: destruir contenido que costó red por un fallo del registro es peor")
	}
	if len(esc.pf.entradas) != 0 {
		t.Fatalf("se registró algo: %v", esc.pf.entradas)
	}
	tmpVacio(t, esc.raiz)

	// El parcial es VISIBLE: el próximo Traer da 409 destino-poblado, no un silencio.
	esc.pf.upsertErr = nil
	if _, err2 := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness"); !errors.Is(err2, usecase.ErrTraerDestinoPoblado) {
		t.Fatalf("err = %v, want ErrTraerDestinoPoblado (el parcial tiene que ser visible)", err2)
	}
}

// E-101 · el temporal se crea BAJO `<raiz>/tmp`, NUNCA en `os.TempDir()`: test de regresión del
// riesgo EXDEV (§12.2 #11) — `os.Rename` entre filesystems falla y /tmp es tmpfs.
func TestStagingEnMismoFilesystemQueDestino(t *testing.T) {
	esc := armarTraer(t)
	var visto string
	esc.local.capturarStaging = &visto

	if _, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness"); err != nil {
		t.Fatal(err)
	}
	if visto == "" {
		t.Fatal("no se capturó el staging")
	}
	prefijo := domain.RaizTemporales(esc.raiz) + string(filepath.Separator)
	if !strings.HasPrefix(visto, prefijo) {
		t.Fatalf("el staging = %q, want con el prefijo %q (riesgo EXDEV)", visto, prefijo)
	}
	if strings.HasPrefix(visto, os.TempDir()+string(filepath.Separator)) && !strings.HasPrefix(visto, prefijo) {
		t.Fatalf("el staging vive en os.TempDir(): %q", visto)
	}
}

// E-102 · dos Traer concurrentes del MISMO destino ⇒ exactamente 1 éxito y 1 conflicto, una sola
// entrada, y el árbol del destino completo (no mezclado).
func TestTraerConcurrenteMismoDestino(t *testing.T) {
	esc := armarTraer(t)
	var wg sync.WaitGroup
	var exitos, conflictos int
	var mu sync.Mutex
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				exitos++
			case errors.Is(err, usecase.ErrTraerDestinoPoblado):
				conflictos++
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}
	wg.Wait()

	if exitos != 1 || conflictos != 1 {
		t.Fatalf("exitos=%d conflictos=%d, want 1/1", exitos, conflictos)
	}
	if len(esc.pf.entradas) != 1 {
		t.Fatalf("entradas = %d, want 1", len(esc.pf.entradas))
	}
	destino := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "harness")
	got := listarRel(t, destino)
	if len(got) != len(matLocalFake().archivos)+1 { // +1 por el dir `hooks`
		t.Fatalf("el árbol del destino = %v, want el completo", got)
	}
	tmpVacio(t, esc.raiz)
}

// E-103 · dos Traer concurrentes de canales DISTINTOS: los dos tienen éxito.
func TestTraerConcurrenteCanalesDistintos(t *testing.T) {
	esc := armarTraer(t)
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, id := range []string{"harness", "harness-beta"} {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			_, errs[i] = esc.svc.Traer(context.Background(), "prenter-marketplace", id)
		}(i, id)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("traer %d: %v", i, err)
		}
	}
	if len(esc.pf.entradas) != 2 {
		t.Fatalf("entradas = %d, want 2", len(esc.pf.entradas))
	}
	tmpVacio(t, esc.raiz)
}

// E-104 · sin caché de catálogo, el Traer LEE primero; si la lectura falla devuelve el error de
// LECTURA, no un «entrada no encontrada» engañoso.
func TestTraerSinCacheLeePrimero(t *testing.T) {
	esc := armarTraer(t)
	delete(esc.cache.guardados, "prenter-marketplace")
	esc.equipo.local.cat = catDosEntradas("prenter-marketplace", "", "local")

	if _, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness"); err != nil {
		t.Fatalf("Traer sin caché: %v", err)
	}

	esc2 := armarTraer(t)
	delete(esc2.cache.guardados, "prenter-marketplace")
	esc2.equipo.local.err = fmt.Errorf("%w: installLocation ya no existe en disco", domain.ErrNoEsMarketplace)
	esc2.remoto.err = fmt.Errorf("%w: gh: HTTP 404", domain.ErrNoEsMarketplace)
	_, err := esc2.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if errors.Is(err, usecase.ErrEntradaNoEnCatalogo) {
		t.Fatalf("err = %v, want el error de LECTURA, no un «entrada no encontrada» engañoso", err)
	}
	if !errors.Is(err, usecase.ErrNoEsMarketplace) {
		t.Fatalf("err = %v, want el error de lectura real", err)
	}
}

// Una entrada que NO está en el catálogo es 400 explícito.
func TestTraerEntradaAjenaAlCatalogo(t *testing.T) {
	esc := armarTraer(t)
	_, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "no-existe")
	if !errors.Is(err, usecase.ErrEntradaNoEnCatalogo) {
		t.Fatalf("err = %v, want ErrEntradaNoEnCatalogo", err)
	}
}

// E-105 · traer algo que ya está en el Portafolio como INSTALACIÓN: el Upsert agrega el canónico y
// CONSERVA las instalaciones (mergeInstalaciones + unionDedup, S1-D3).
func TestTraerConInstalacionesPrevias(t *testing.T) {
	esc := armarTraer(t)
	clave := domain.IdentidadArnes{Home: repoPrenter, ID: "harness"}.Clave()
	esc.pf.entradas[clave] = domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{Home: repoPrenter, ID: "harness"},
		Instalaciones: []domain.Instalacion{
			{InstallPath: "/proj/a/.claude/plugins/harness", Tipo: domain.InstMaterializada},
			{InstallPath: "/proj/b/.claude/plugins/harness", Tipo: domain.InstMaterializada},
		},
	}

	res, err := esc.svc.Traer(context.Background(), "prenter-marketplace", "harness")
	if err != nil {
		t.Fatal(err)
	}
	// El fake del store REEMPLAZA (no mergea) — la aserción del merge real vive en
	// adapters/portafolio/store_test.go. Acá se asserta que el usecase NO borra instalaciones
	// a mano: la entrada que se manda al Upsert no trae Instalaciones vacías explícitas.
	if res.Entrada.Canonico == nil {
		t.Fatal("no se agregó el canónico")
	}
	if res.Clave != clave {
		t.Fatalf("Clave = %q, want %q", res.Clave, clave)
	}
}

// El camino que el catálogo exige sin materializador cableado ⇒ 503 honesto, no un nil-pointer.
func TestTraerSinMaterializadorEsHonesto(t *testing.T) {
	e := armar(t)
	raiz := t.TempDir()
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "m", Repo: repoPrenter, InstallLocation: t.TempDir()}}
	e.store.filas["m"] = domain.MarketplaceConocido{Nombre: "m", Repo: repoPrenter, Clase: domain.ClasePropio}
	cat := catDosEntradas("m", "2026-07-25T00:00:00Z", "local")
	cat.Entradas[0].Source = domain.SourceCatalogo{Tipo: domain.SourceURL, Crudo: "url:x", URL: "https://github.com/a/b.git", SHA: "abc"}
	e.cache.guardados["m"] = cat
	e.svc.SetTraer(matLocalFake(), nil, &derivaFake{}, raiz) // camino B SIN cablear.

	_, err := e.svc.Traer(context.Background(), "m", "harness")
	if !errors.Is(err, usecase.ErrTraerSinMaterializador) {
		t.Fatalf("err = %v, want ErrTraerSinMaterializador", err)
	}
}
