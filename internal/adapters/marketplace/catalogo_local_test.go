package marketplace

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// mkt arma un MarketplaceConocido apuntando a un dir de fixture.
func mkt(nombre, dir string, clase domain.ClaseMarketplace) domain.MarketplaceConocido {
	abs, _ := filepath.Abs(dir)
	return domain.MarketplaceConocido{Nombre: nombre, Clase: clase, InstallLocation: abs}
}

// copiarFixture duplica un árbol de testdata en un dir temporal, para poder mutarlo (borrar un
// subdir, romper un archivo, cambiar permisos) sin tocar el fixture del repo.
func copiarFixture(t *testing.T) string {
	t.Helper()
	const origen = "testdata/prenter"
	destino := t.TempDir()
	err := filepath.WalkDir(origen, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(origen, p)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		b, rerr := os.ReadFile(p) //nolint:gosec // G304: fixture del propio repo.
		if rerr != nil {
			return rerr
		}
		return os.WriteFile(target, b, 0o600) //nolint:gosec // G703: target se arma con filepath.Join sobre t.TempDir().
	})
	if err != nil {
		t.Fatalf("copiar fixture %s: %v", origen, err)
	}
	return destino
}

func entradaPorNombre(t *testing.T, cat domain.Catalogo, nombre string) domain.EntradaCatalogo {
	t.Helper()
	for _, e := range cat.Entradas {
		if e.Nombre == nombre {
			return e
		}
	}
	t.Fatalf("entrada %q ausente en el catálogo (%d entradas)", nombre, len(cat.Entradas))
	return domain.EntradaCatalogo{}
}

// E-02 · catálogo propio legible OFFLINE. Insumo: copia literal del checkout real de prenter.
func TestLectorLocalPrenterDosEntradas(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("prenter-marketplace", "testdata/prenter", domain.ClasePropio))
	if err != nil {
		t.Fatalf("Leer: %v", err)
	}
	if cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("Lectura.Tipo = %q, want leido", cat.Lectura.Tipo)
	}
	if cat.Lectura.Fuente != "local" {
		t.Fatalf("Lectura.Fuente = %q, want local", cat.Lectura.Fuente)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2", len(cat.Entradas))
	}
	if cat.Entradas[0].Nombre != "harness" || cat.Entradas[1].Nombre != "harness-beta" {
		t.Fatalf("nombres = %q/%q, want harness/harness-beta", cat.Entradas[0].Nombre, cat.Entradas[1].Nombre)
	}
	if cat.OwnerNombre != "Prenter" || cat.OwnerEmail != "hola@alpacapurpura.lat" {
		t.Fatalf("owner = %q/%q, want Prenter/hola@alpacapurpura.lat", cat.OwnerNombre, cat.OwnerEmail)
	}
	// E-03/E-04 en el lector real: mismo source ⇒ canales; versión derivada de la ruta.
	if got := cat.Entradas[0].ComparteSourceCon; len(got) != 1 || got[0] != "harness-beta" {
		t.Fatalf("ComparteSourceCon = %v, want [harness-beta]", got)
	}
	if cat.Entradas[0].Version != "0.5.3" || cat.Entradas[0].VersionDe != domain.VersionDeSource {
		t.Fatalf("version = %q/%q, want 0.5.3/derivada-de-source", cat.Entradas[0].Version, cat.Entradas[0].VersionDe)
	}
	if cat.Clase != domain.ClasePropio {
		t.Fatalf("Clase = %q, want propio (la del registro)", cat.Clase)
	}
}

// E-05 · enriquecimiento `catalogo.json` real: canales + 4 versiones + EstadoCanal por fila.
// Corrección del spec anotada en plan-pruebas: `0.5.1` es deprecada en `versiones[]`, NO en una
// fila del catálogo (ninguna fila apunta a 0.5.1).
func TestEnriquecimientoCatalogoJSON(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("prenter-marketplace", "testdata/prenter", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Canales["estable"] != "0.5.3" || cat.Canales["beta"] != "0.5.3" {
		t.Fatalf("Canales = %v, want {estable:0.5.3, beta:0.5.3}", cat.Canales)
	}
	if len(cat.Versiones) != 4 {
		t.Fatalf("len(Versiones) = %d, want 4", len(cat.Versiones))
	}
	var vio bool
	for _, v := range cat.Versiones {
		if v.Version == "0.5.1" && v.Estado == "deprecada" {
			vio = true
		}
	}
	if !vio {
		t.Fatalf("Versiones no contiene {0.5.1 deprecada}: %+v", cat.Versiones)
	}
	if got := entradaPorNombre(t, cat, "harness").EstadoCanal; got != "habilitada" {
		t.Fatalf("harness.EstadoCanal = %q, want habilitada", got)
	}
}

// E-06 · un marketplace SIN `catalogo.json` degrada SIN RUIDO: es convención de prenter, no del
// estándar. Insumo: la muestra real del catálogo oficial (25 filas, sin catalogo.json).
func TestSinCatalogoJSONDegradaSinRuido(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("claude-plugins-official", "testdata/oficial", domain.ClaseReferencia))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Canales != nil || cat.Versiones != nil {
		t.Fatalf("Canales/Versiones = %v/%v, want nil/nil", cat.Canales, cat.Versiones)
	}
	for _, e := range cat.Entradas {
		if e.EstadoCanal != "" {
			t.Fatalf("%s: EstadoCanal = %q, want vacío", e.Nombre, e.EstadoCanal)
		}
	}
	if cat.Lectura.Motivo != "" {
		t.Fatalf("Lectura.Motivo = %q, want vacío (degradado SIN ruido)", cat.Lectura.Motivo)
	}
	if cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("Lectura.Tipo = %q, want leido", cat.Lectura.Tipo)
	}
	if len(cat.Entradas) != 25 {
		t.Fatalf("len(Entradas) = %d, want 25 (la muestra del oficial)", len(cat.Entradas))
	}
}

// AG-D13 · las CUATRO formas reales de `source` se normalizan; ninguna se descarta.
func TestLasCuatroFormasDeSource(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("claude-plugins-official", "testdata/oficial", domain.ClaseReferencia))
	if err != nil {
		t.Fatal(err)
	}
	casos := map[string]domain.TipoSource{
		"agent-sdk-dev":                 domain.SourceRutaRelativa,
		"42crunch-api-security-testing": domain.SourceGitSubdir,
		"agentforce-adlc":               domain.SourceURL,
		"fullstory":                     domain.SourceGitHub,
	}
	for nombre, tipo := range casos {
		e := entradaPorNombre(t, cat, nombre)
		if e.Source.Tipo != tipo {
			t.Fatalf("%s: Source.Tipo = %q, want %q", nombre, e.Source.Tipo, tipo)
		}
		if e.Source.Crudo == "" {
			t.Fatalf("%s: Source.Crudo vacío (es la clave de agrupamiento de canales)", nombre)
		}
	}
	// El `sha` es el pin universal (220/220 formas objeto); el `commit` de la forma `github` se
	// conserva para poder citar los dos en el aviso de divergencia (C18/E-95).
	crunch := entradaPorNombre(t, cat, "42crunch-api-security-testing")
	if crunch.Source.SHA != "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad" || crunch.Source.Ref != "v1.5.5" {
		t.Fatalf("42crunch source = %+v, want sha 30287f5e… + ref v1.5.5", crunch.Source)
	}
	if crunch.Source.Ruta != "plugins/api-security-testing" {
		t.Fatalf("42crunch Ruta = %q, want plugins/api-security-testing", crunch.Source.Ruta)
	}
	fs := entradaPorNombre(t, cat, "fullstory")
	if fs.Source.SHA != "b20614e2d08d7a7c70775bb62b5af640f60b024b" || fs.Source.Commit != "1ec5865e7ab1449f9a0859d164c4b6a8c53b6e2f" {
		t.Fatalf("fullstory source = %+v, want los DOS hashes conservados", fs.Source)
	}
	// AG-D14 · el campo `version` estándar gana cuando está (14 filas reales lo traen).
	if e := entradaPorNombre(t, cat, "clangd-lsp"); e.Version != "1.0.0" || e.VersionDe != domain.VersionDeCampo {
		t.Fatalf("clangd-lsp version = %q/%q, want 1.0.0/campo-version", e.Version, e.VersionDe)
	}
	// E-38 (lado adapter) · `renames` real del oficial: convex-backend → convex.
	if e := entradaPorNombre(t, cat, "convex"); len(e.NombreAnterior) != 1 || e.NombreAnterior[0] != "convex-backend" {
		t.Fatalf("convex.NombreAnterior = %v, want [convex-backend]", e.NombreAnterior)
	}
}

// E-41 · `owner` con `url` en vez de `email` (caveman real). Ningún campo inventado.
func TestOwnerConURL(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("caveman", "testdata/caveman", domain.ClaseReferencia))
	if err != nil {
		t.Fatal(err)
	}
	if cat.OwnerNombre != "Julius Brussee" {
		t.Fatalf("OwnerNombre = %q", cat.OwnerNombre)
	}
	if cat.OwnerURL != "https://github.com/JuliusBrussee" {
		t.Fatalf("OwnerURL = %q", cat.OwnerURL)
	}
	if cat.OwnerEmail != "" {
		t.Fatalf("OwnerEmail = %q, want vacío (no se inventa)", cat.OwnerEmail)
	}
	// `source: "./"` existe de verdad (§13.1 hecho 8).
	if e := entradaPorNombre(t, cat, "caveman"); e.Source.Ruta != "./" {
		t.Fatalf("caveman source = %+v, want ruta ./", e.Source)
	}
}

// E-42 · `description` top-level gana a `metadata.description`; sin ninguna, "".
func TestDescripcionTopLevelGana(t *testing.T) {
	oficial, err := (&LectorLocal{}).Leer(context.Background(), mkt("claude-plugins-official", "testdata/oficial", domain.ClaseReferencia))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(oficial.Descripcion, "Directory of popular Claude Code extensions") {
		t.Fatalf("oficial.Descripcion = %q, want la top-level", oficial.Descripcion)
	}
	prenter, err := (&LectorLocal{}).Leer(context.Background(), mkt("prenter-marketplace", "testdata/prenter", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prenter.Descripcion, "Private Prenter marketplace") {
		t.Fatalf("prenter.Descripcion = %q, want la de metadata", prenter.Descripcion)
	}
	ambas, err := (&LectorLocal{}).Leer(context.Background(), mkt("description-ambas", "testdata/dañados/description-ambas", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if ambas.Descripcion != "la top-level gana" {
		t.Fatalf("Descripcion = %q, want la top-level", ambas.Descripcion)
	}
}

// E-31 · `marketplace.json` SIN la clave `plugins`: Entradas nil (⇒ null en el cable) + motivo.
// «Leí y el archivo está incompleto», no «no tiene arneses» (BR-4).
func TestSinClavePlugins(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("sin-plugins", "testdata/dañados/sin-plugins", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Entradas != nil {
		t.Fatalf("Entradas = %v, want nil (null en el cable)", cat.Entradas)
	}
	if cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("Lectura.Tipo = %q, want leido", cat.Lectura.Tipo)
	}
	if cat.Lectura.Motivo != "marketplace.json sin la clave plugins" {
		t.Fatalf("Lectura.Motivo = %q", cat.Lectura.Motivo)
	}
}

// E-32 · `plugins: []` de VERDAD: Entradas es un slice vacío no-nil (⇒ `[]`, no `null`), Entradas=0
// y motivo vacío. Es una afirmación evidenciada, no un «no sé».
func TestPluginsVacioEsAfirmacion(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("plugins-vacio", "testdata/dañados/plugins-vacio", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Entradas == nil {
		t.Fatal("Entradas = nil, want [] (el marketplace declara plugins: [] de verdad)")
	}
	if len(cat.Entradas) != 0 {
		t.Fatalf("len(Entradas) = %d, want 0", len(cat.Entradas))
	}
	if cat.Lectura.Tipo != domain.LecturaLeida || cat.Lectura.Entradas != 0 || cat.Lectura.Motivo != "" {
		t.Fatalf("Lectura = %+v, want {leido 0 ''}", cat.Lectura)
	}
}

// E-33 · JSON sintácticamente corrupto: error con el OFFSET visible, cero pánico.
func TestJSONCorruptoNoCrashea(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("json-corrupto", "testdata/dañados/json-corrupto", domain.ClasePropio))
	if err == nil {
		t.Fatal("Leer devolvió nil error con JSON corrupto")
	}
	if !errors.Is(err, ErrNoEsMarketplace) {
		t.Fatalf("err = %v, want ErrNoEsMarketplace", err)
	}
	if !strings.Contains(err.Error(), "offset") {
		t.Fatalf("el motivo debe traer el offset del parseo: %v", err)
	}
	if cat.Entradas != nil {
		t.Fatalf("Entradas = %v, want nil", cat.Entradas)
	}
}

// Un archivo sin `name` raíz no es un marketplace (§5.4).
func TestSinNameRaizNoEsMarketplace(t *testing.T) {
	_, err := (&LectorLocal{}).Leer(context.Background(), mkt("sin-name", "testdata/dañados/sin-name", domain.ClasePropio))
	if !errors.Is(err, ErrNoEsMarketplace) || !strings.Contains(err.Error(), "sin `name`") {
		t.Fatalf("err = %v, want ErrNoEsMarketplace con «sin `name`»", err)
	}
}

// E-34 · el `source` apunta a un dir que NO existe en el checkout: la fila se muestra IGUAL, con
// el problema del dato a la vista. La derivación de versión no depende de que el dir exista.
func TestSourceRutaAusenteAvisaSinBorrarFila(t *testing.T) {
	dir := copiarFixture(t)
	if err := os.RemoveAll(filepath.Join(dir, "plugins", "harness", "0.5.3")); err != nil {
		t.Fatal(err)
	}
	cat, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Clase: domain.ClasePropio, InstallLocation: dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	e := entradaPorNombre(t, cat, "harness")
	want := "el catálogo apunta a ./plugins/harness/0.5.3 pero ese dir no existe en el checkout"
	if len(e.Aviso) != 1 || e.Aviso[0] != want {
		t.Fatalf("Aviso = %v, want [%q]", e.Aviso, want)
	}
	if e.Version != "0.5.3" {
		t.Fatalf("Version = %q, want 0.5.3 (la derivación no depende del dir)", e.Version)
	}
	if cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("Lectura.Tipo = %q, want leido", cat.Lectura.Tipo)
	}
}

// E-35 (lado adapter) · `installLocation` que ya no existe en disco ⇒ ErrNoEsMarketplace con el
// motivo exacto; el usecase cae al remoto. NUNCA un `[]`.
func TestInstallLocationAusenteCaeARemoto(t *testing.T) {
	ausente := filepath.Join(t.TempDir(), "borrado")
	_, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{
		Nombre: "x", InstallLocation: ausente,
	})
	if !errors.Is(err, ErrNoEsMarketplace) {
		t.Fatalf("err = %v, want ErrNoEsMarketplace", err)
	}
	if !strings.Contains(err.Error(), "installLocation ya no existe en disco") {
		t.Fatalf("motivo = %v, want «installLocation ya no existe en disco»", err)
	}
	// Sin checkout declarado, tampoco se inventa nada.
	if _, err2 := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "x"}); !errors.Is(err2, ErrNoEsMarketplace) {
		t.Fatalf("err = %v, want ErrNoEsMarketplace", err2)
	}
}

// E-36 · permisos: `marketplace.json` en modo 0000 ⇒ sin-acceso con `permission denied` y la
// ruta, jamás «no es un marketplace».
func TestSinPermisoDeLectura(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el chmod 0000 no impide leer: el caso no es reproducible")
	}
	dir := copiarFixture(t)
	ruta := filepath.Join(dir, ".claude-plugin", "marketplace.json")
	if err := os.Chmod(ruta, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ruta, 0o600) })

	cat, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "prenter-marketplace", InstallLocation: dir})
	if err == nil {
		t.Fatal("Leer devolvió nil error con el archivo en modo 0000")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Fatalf("err = %v, want os.ErrPermission (sin-acceso, no url-no-resuelve)", err)
	}
	if !strings.Contains(err.Error(), ruta) {
		t.Fatalf("el motivo debe traer la ruta: %v", err)
	}
	if cat.Entradas != nil {
		t.Fatalf("Entradas = %v, want nil", cat.Entradas)
	}
}

// E-37 · dos filas con el MISMO `name`: las dos se conservan, las dos con aviso. El sistema no
// elige una ni dedupea el error ajeno.
func TestNombreDuplicadoAmbasVisibles(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("nombre-duplicado", "testdata/dañados/nombre-duplicado", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 3 {
		t.Fatalf("len(Entradas) = %d, want 3 (las dos duplicadas + la sana)", len(cat.Entradas))
	}
	dup := 0
	for _, e := range cat.Entradas {
		if e.Nombre != "harness" {
			continue
		}
		dup++
		if !e.TieneAvisoDuplicado() {
			t.Fatalf("la fila duplicada %q no trae el aviso: %v", e.Nombre, e.Aviso)
		}
		// El dominio la fuerza a `no-comparable` (fila 4 de la tabla de verdad).
		s := domain.CalcularSituacion(e, []domain.CoincidenciaPortafolio{{
			Entrada: domain.EntradaPortafolio{
				Identidad: domain.IdentidadArnes{Home: "github.com/a/b", ID: "harness"},
				Canonico:  &domain.Canonico{Path: "/x", Version: "0.5.3"},
			},
			Via: domain.ViaHomeDeclarado,
		}})
		if s.Tipo != domain.SituacionNoComparable {
			t.Fatalf("una fila duplicada debe dar no-comparable, got %q", s.Tipo)
		}
	}
	if dup != 2 {
		t.Fatalf("filas «harness» = %d, want 2", dup)
	}
}

// E-39 (lado adapter) · forma de `source` no reconocida (`"source": 42`): visible, no descartada.
func TestSourceDesconocidoEsVisible(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("source-raro", "testdata/dañados/source-raro", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 1 {
		t.Fatalf("len(Entradas) = %d, want 1 (la fila NO se descarta)", len(cat.Entradas))
	}
	e := cat.Entradas[0]
	if e.Source.Tipo != domain.SourceDesconocido || e.Source.Crudo != "42" {
		t.Fatalf("Source = %+v, want {desconocido 42}", e.Source)
	}
	if len(e.Aviso) == 0 || !strings.Contains(e.Aviso[0], "42") {
		t.Fatalf("Aviso = %v, want el crudo visible", e.Aviso)
	}
}

// E-40 · fila sin `name`: se descarta (sin nombre no hay nada instalable) pero el descarte es
// VISIBLE en Truncado y en el motivo.
func TestFilaSinNombreSeDescartaVisible(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("fila-sin-nombre", "testdata/dañados/fila-sin-nombre", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 1 || cat.Entradas[0].Nombre != "sano" {
		t.Fatalf("Entradas = %+v, want solo la sana", cat.Entradas)
	}
	if cat.Truncado != 1 {
		t.Fatalf("Truncado = %d, want 1", cat.Truncado)
	}
	if !strings.Contains(cat.Lectura.Motivo, "1 fila(s) sin name") {
		t.Fatalf("Lectura.Motivo = %q, want «1 fila(s) sin name…»", cat.Lectura.Motivo)
	}
}

// AG-D14 · `version: null` explícito y `version: ""` son AUSENTE, no una versión fabricada.
func TestVersionNullOVaciaEsAusente(t *testing.T) {
	cat, err := (&LectorLocal{}).Leer(context.Background(), mkt("version-null-explicito", "testdata/dañados/version-null-explicito", domain.ClasePropio))
	if err != nil {
		t.Fatal(err)
	}
	for _, nombre := range []string{"con-null", "con-vacio"} {
		e := entradaPorNombre(t, cat, nombre)
		if e.Version != "" || e.VersionDe != domain.VersionAusente {
			t.Fatalf("%s: version = %q/%q, want ausente", nombre, e.Version, e.VersionDe)
		}
	}
	if e := entradaPorNombre(t, cat, "con-semver"); e.Version != "2.1.0" || e.VersionDe != domain.VersionDeCampo {
		t.Fatalf("con-semver version = %q/%q, want 2.1.0/campo-version", e.Version, e.VersionDe)
	}
}

// E-53 · symlink en `installLocation`: válido ⇒ lee normal; roto ⇒ igual que E-35.
func TestInstallLocationSymlink(t *testing.T) {
	base := t.TempDir()
	fixture, _ := filepath.Abs("testdata/prenter")
	link := filepath.Join(base, "link-ok")
	if err := os.Symlink(fixture, link); err != nil {
		t.Skipf("symlinks no disponibles: %v", err)
	}
	cat, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "prenter-marketplace", InstallLocation: link})
	if err != nil {
		t.Fatalf("symlink válido: %v", err)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2", len(cat.Entradas))
	}

	roto := filepath.Join(base, "link-roto")
	if serr := os.Symlink(filepath.Join(base, "no-existe"), roto); serr != nil {
		t.Fatal(serr)
	}
	_, err = (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "x", InstallLocation: roto})
	if !errors.Is(err, ErrNoEsMarketplace) {
		t.Fatalf("symlink roto: err = %v, want ErrNoEsMarketplace", err)
	}
}

// E-72 · `catalogo.json` presente pero CORRUPTO: las entradas salen completas, el
// enriquecimiento no se aplica, y el aviso queda VISIBLE. Degradar sin ruido ≠ ocultar un
// archivo roto que el operador puso a propósito.
func TestCatalogoJSONCorruptoAvisaSinRomper(t *testing.T) {
	dir := copiarFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "catalogo.json"), []byte(`{"marketplace": "prenter-mark`), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "prenter-marketplace", InstallLocation: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2 (el catálogo principal se leyó bien)", len(cat.Entradas))
	}
	if cat.Canales != nil || cat.Versiones != nil {
		t.Fatalf("Canales/Versiones = %v/%v, want nil/nil", cat.Canales, cat.Versiones)
	}
	if !strings.Contains(cat.Lectura.Motivo, "catalogo.json ignorado") {
		t.Fatalf("Lectura.Motivo = %q, want el sufijo «(catalogo.json ignorado: …)»", cat.Lectura.Motivo)
	}
}

// E-73 · `catalogo.json` de OTRO marketplace: se ignora con el mismo aviso. El enriquecimiento
// no se aplica a ciegas.
func TestCatalogoJSONDeOtroMarketplaceSeIgnora(t *testing.T) {
	dir := copiarFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "catalogo.json"), []byte(`{"marketplace":"otro","canales":{"estable":"9.9.9"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "prenter-marketplace", InstallLocation: dir})
	if err != nil {
		t.Fatal(err)
	}
	if cat.Canales != nil {
		t.Fatalf("Canales = %v, want nil (no se aplica a ciegas)", cat.Canales)
	}
	if !strings.Contains(cat.Lectura.Motivo, "catalogo.json ignorado") {
		t.Fatalf("Lectura.Motivo = %q, want el aviso visible", cat.Lectura.Motivo)
	}
}

// El lector local NUNCA escribe en el checkout ajeno (boundary
// `marketplace-referencia-es-solo-procedencia`, L2 punto 3): aserción de bytes y mtimes.
func TestLectorLocalNoEscribeEnElCheckout(t *testing.T) {
	dir := copiarFixture(t)
	antes := huellaDeArbol(t, dir)
	if _, err := (&LectorLocal{}).Leer(context.Background(), domain.MarketplaceConocido{Nombre: "prenter-marketplace", InstallLocation: dir}); err != nil {
		t.Fatal(err)
	}
	if despues := huellaDeArbol(t, dir); despues != antes {
		t.Fatalf("el lector local modificó el checkout ajeno:\nantes:   %s\ndespués: %s", antes, despues)
	}
}

// huellaDeArbol lista paths + tamaño + mtime de dir, para asertar que un árbol NO cambió.
func huellaDeArbol(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		fi, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		b.WriteString(rel)
		b.WriteString("|")
		if !d.IsDir() {
			b.WriteString(strings.Join([]string{
				strconvItoa(int(fi.Size())), fi.ModTime().UTC().Format("2006-01-02T15:04:05.000000000Z"),
			}, ","))
		}
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatalf("huella de %s: %v", dir, err)
	}
	return b.String()
}

func strconvItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}
