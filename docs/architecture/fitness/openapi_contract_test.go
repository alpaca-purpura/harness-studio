package fitness

// Enforcer de `docs/architecture/boundaries/ruta-servida-esta-declarada.md`.
//
// POR QUÉ EXISTE. Hasta hoy el único paso de contrato de CI (`ci.yml:73`, `openapi-gen-check`)
// estaba **inerte**: su condicional mira `web/src/shared/api/generated/**`, un directorio que no
// existe. Con nadie mirando, el router se fue a **54 rutas** mientras el contrato declaraba
// **39**. Las 15 de diferencia no son una decisión: son lo que pasa cuando un documento no tiene
// quien lo compare contra la realidad.
//
// QUÉ ENFORZA. La ecuación en las dos direcciones, más la disciplina de la exención:
//
//	TestRutaServidaEstaDeclarada     — servida ⇒ declarada, o exenta con razón (el ratchet)
//	TestRutaDeclaradaSeSirve         — declarada ⇒ servida (mata las rutas fantasma)
//	TestQueryParamEstaDeclarado      — un `?param` que el handler LEE está en `parameters:`
//	TestExencionDeContratoTieneRazon — ninguna exención con la razón vacía
//
// ⚠ SE ESCANEAN LOS DOS VERBOS DE REGISTRO. Mirar sólo `mux.HandleFunc(` deja fuera
// `GET /api/events`, que se registra con `mux.Handle(` (un `http.Handler`, no una func), y
// produce un falso «declarada sin servir». Verificado: es exactamente el caso de `router.go`.
//
// `gopkg.in/yaml.v3` ya es dependencia directa del módulo y esto es código de test: cero peso
// nuevo en el binario (`peso-del-binario-es-presupuesto` satisfecho).

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	rutaOpenAPI     = "docs/architecture/contracts/api/openapi.yaml"
	rutaSinDeclarar = "docs/architecture/contracts/api/_sin-declarar.yaml"
	dirTransporte   = "internal/adapters/transport/http"
)

// registroMux captura `mux.Handle("<patrón>"` y `mux.HandleFunc("<patrón>"`. El patrón de
// net/http 1.22 es `[MÉTODO ]/path`, que es exactamente la forma en que las guardamos.
var registroMux = regexp.MustCompile(`mux\.(?:Handle|HandleFunc)\(\s*"([^"]+)"`)

// rutasDelRouter devuelve lo que el daemon REALMENTE sirve, leído del registro del mux.
func rutasDelRouter(t *testing.T) map[string]bool {
	t.Helper()
	root := repoRoot()
	if root == "" {
		t.Skip("sin go.mod: nada que escanear")
	}
	dir := filepath.Join(root, dirTransporte)
	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", dirTransporte, err)
	}
	out := map[string]bool{}
	for _, e := range entradas {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		//nolint:gosec // G304: `dir` es una constante del repo y `e.Name()` viene de ReadDir
		// sobre ese mismo dir; no hay entrada del usuario en el camino. Es un test de fitness.
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("no se pudo leer %s: %v", e.Name(), err)
		}
		for _, m := range registroMux.FindAllStringSubmatch(string(b), -1) {
			out[strings.TrimSpace(m[1])] = true
		}
	}
	if len(out) == 0 {
		t.Fatalf("cero rutas encontradas en %s — el escáner se rompió, no el router", dirTransporte)
	}
	return out
}

// docOpenAPI es la porción del contrato que este enforcer mira.
type docOpenAPI struct {
	Servers []struct {
		URL string `yaml:"url"`
	} `yaml:"servers"`
	Paths map[string]map[string]struct {
		Parameters []struct {
			Name string `yaml:"name"`
			In   string `yaml:"in"`
			Ref  string `yaml:"$ref"`
		} `yaml:"parameters"`
	} `yaml:"paths"`
}

var verbosHTTP = map[string]bool{
	"get": true, "post": true, "put": true, "patch": true,
	"delete": true, "head": true, "options": true,
}

func leerOpenAPI(t *testing.T) (doc docOpenAPI, prefijo string) {
	t.Helper()
	root := repoRoot()
	//nolint:gosec // G304: `rutaOpenAPI` es una constante de este archivo; cero entrada externa.
	b, err := os.ReadFile(filepath.Join(root, rutaOpenAPI))
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", rutaOpenAPI, err)
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("%s no parsea como YAML: %v", rutaOpenAPI, err)
	}
	// El `servers[0].url` trae el host + el prefijo real (`http://127.0.0.1:4200/api`): los
	// `paths:` del contrato son relativos a él, y el router registra la ruta completa.
	if len(doc.Servers) > 0 {
		if i := strings.Index(doc.Servers[0].URL, "/api"); i >= 0 {
			prefijo = doc.Servers[0].URL[i:]
		}
	}
	if prefijo == "" {
		t.Fatalf("%s: no se pudo derivar el prefijo /api de servers[0].url", rutaOpenAPI)
	}
	return doc, prefijo
}

// rutasDelContrato devuelve `MÉTODO /prefijo/path` por cada operación declarada.
func rutasDelContrato(t *testing.T) map[string]bool {
	t.Helper()
	doc, prefijo := leerOpenAPI(t)
	out := map[string]bool{}
	for p, ops := range doc.Paths {
		for verbo := range ops {
			if !verbosHTTP[strings.ToLower(verbo)] {
				continue
			}
			out[fmt.Sprintf("%s %s%s", strings.ToUpper(verbo), prefijo, p)] = true
		}
	}
	if len(out) == 0 {
		t.Fatalf("cero operaciones en %s — el parser se rompió, no el contrato", rutaOpenAPI)
	}
	return out
}

type exencion struct {
	Ruta  string `yaml:"ruta"`
	Razon string `yaml:"razon"`
}

type docExenciones struct {
	Exenciones []exencion `yaml:"exenciones"`
}

// exenciones devuelve ruta→razón de la allowlist. Una ruta sin razón NO cuenta como exenta:
// la razón es el precio de la excepción, y `TestExencionDeContratoTieneRazon` la cobra.
func exenciones(t *testing.T) map[string]string {
	t.Helper()
	root := repoRoot()
	//nolint:gosec // G304: `rutaSinDeclarar` es una constante de este archivo; cero entrada externa.
	b, err := os.ReadFile(filepath.Join(root, rutaSinDeclarar))
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", rutaSinDeclarar, err)
	}
	var doc docExenciones
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("%s no parsea como YAML: %v", rutaSinDeclarar, err)
	}
	out := map[string]string{}
	for _, e := range doc.Exenciones {
		out[strings.TrimSpace(e.Ruta)] = strings.TrimSpace(e.Razon)
	}
	return out
}

// TestRutaServidaEstaDeclarada — el ratchet. Una ruta que el daemon sirve está en el contrato
// o está en la allowlist con razón. No hay tercera opción.
func TestRutaServidaEstaDeclarada(t *testing.T) {
	servidas := rutasDelRouter(t)
	declaradas := rutasDelContrato(t)
	exentas := exenciones(t)

	var huerfanas []string
	for r := range servidas {
		if declaradas[r] || exentas[r] != "" {
			continue
		}
		huerfanas = append(huerfanas, r)
	}
	sort.Strings(huerfanas)
	if len(huerfanas) > 0 {
		t.Fatalf("%d ruta(s) servidas sin declarar en %s ni exentas en %s:\n  %s\n"+
			"Declarala en el contrato, o agregala a la allowlist CON su razón.",
			len(huerfanas), rutaOpenAPI, rutaSinDeclarar, strings.Join(huerfanas, "\n  "))
	}
}

// TestRutaDeclaradaSeSirve — la dirección inversa: una operación del contrato que nadie sirve
// es una promesa vacía. Un cliente que la lea escribe código contra un 404.
func TestRutaDeclaradaSeSirve(t *testing.T) {
	servidas := rutasDelRouter(t)
	declaradas := rutasDelContrato(t)

	var fantasma []string
	for r := range declaradas {
		if !servidas[r] {
			fantasma = append(fantasma, r)
		}
	}
	sort.Strings(fantasma)
	if len(fantasma) > 0 {
		t.Fatalf("%d ruta(s) declaradas que el router NO sirve (contrato fantasma):\n  %s",
			len(fantasma), strings.Join(fantasma, "\n  "))
	}
}

// TestQueryParamEstaDeclarado — un `?param` que el handler LEE es superficie: cambia la
// respuesta. Si no está en `parameters:`, el contrato describe una operación que no existe.
//
// ⚠ EL EMPAREJADO VA POR HANDLER, NO POR ARCHIVO. La primera versión de este enforcer
// cruzaba «rutas registradas en el archivo X» con «queries leídas en el archivo X» y pasaba
// **vacuamente**: en este árbol `router.go` registra TODAS las rutas y no lee ni un query,
// mientras `sessions.go` lee `?arnes=`/`?cerradas=` y no registra ninguna. Cero intersección,
// cero hallazgos, verde mentiroso. Se resuelve por AST: del registro se saca el identificador
// del constructor del handler (`listSessions` en `mux.HandleFunc("GET /api/sessions",
// listSessions(sessions))`) y se leen los queries del cuerpo de ESA función, closures incluidas
// — que es justo donde viven, porque los handlers de este árbol son closures devueltas.
//
// Alcance deliberado: sólo los parámetros de los handlers de rutas **declaradas**. Los de las
// exentas se declararán junto con su ruta; exigirlos ahora obligaría a documentar a medias un
// módulo que la propia allowlist dice que todavía se mueve.
func TestQueryParamEstaDeclarado(t *testing.T) {
	doc, prefijo := leerOpenAPI(t)

	// declarados: `MÉTODO /prefijo/path` → conjunto de nombres de query declarados. Un
	// `$ref` a components/parameters se cuenta como declarado: resolverlo entero es otro
	// enforcer, y acá lo que importa es que el autor lo haya nombrado.
	declarados := map[string]map[string]bool{}
	for p, ops := range doc.Paths {
		comunes := map[string]bool{}
		for verbo, op := range ops {
			if strings.EqualFold(verbo, "parameters") {
				for _, par := range op.Parameters {
					if par.In == "query" || par.Ref != "" {
						comunes[par.Name] = true
					}
				}
			}
		}
		for verbo, op := range ops {
			if !verbosHTTP[strings.ToLower(verbo)] {
				continue
			}
			set := map[string]bool{}
			for n := range comunes {
				set[n] = true
			}
			for _, par := range op.Parameters {
				if par.In == "query" || par.Ref != "" {
					set[par.Name] = true
				}
			}
			declarados[fmt.Sprintf("%s %s%s", strings.ToUpper(verbo), prefijo, p)] = set
		}
	}

	rutaHandler, queriesDe := handlersDelTransporte(t)
	var faltan []string
	for ruta, handler := range rutaHandler {
		set, esDeclarada := declarados[ruta]
		if !esDeclarada {
			continue // exenta o no declarada: lo cubre TestRutaServidaEstaDeclarada
		}
		for _, q := range queriesDe[handler] {
			if !set[q] {
				faltan = append(faltan, fmt.Sprintf("%s lee ?%s= (handler %s) y no está en parameters:", ruta, q, handler))
			}
		}
	}
	sort.Strings(faltan)
	faltan = dedup(faltan)
	if len(faltan) > 0 {
		t.Fatalf("%d parámetro(s) de query sin declarar en %s:\n  %s",
			len(faltan), rutaOpenAPI, strings.Join(faltan, "\n  "))
	}
}

// TestExencionDeContratoTieneRazon — la razón es lo que impide que la allowlist se vuelva un
// permiso permanente. Una entrada sin razón es una ruta escondida con un paso extra.
func TestExencionDeContratoTieneRazon(t *testing.T) {
	root := repoRoot()
	//nolint:gosec // G304: `rutaSinDeclarar` es una constante de este archivo; cero entrada externa.
	b, err := os.ReadFile(filepath.Join(root, rutaSinDeclarar))
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", rutaSinDeclarar, err)
	}
	var doc docExenciones
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("%s no parsea como YAML: %v", rutaSinDeclarar, err)
	}
	if len(doc.Exenciones) == 0 {
		t.Fatalf("%s sin entradas: o el parser se rompió, o el archivo perdió su contenido", rutaSinDeclarar)
	}
	var mudas []string
	for _, e := range doc.Exenciones {
		if strings.TrimSpace(e.Ruta) == "" {
			mudas = append(mudas, "(entrada sin `ruta`)")
			continue
		}
		if len(strings.TrimSpace(e.Razon)) < 20 {
			mudas = append(mudas, fmt.Sprintf("%q: razón vacía o de menos de 20 caracteres", e.Ruta))
		}
	}
	if len(mudas) > 0 {
		t.Fatalf("%d exención(es) sin razón escrita en %s:\n  %s",
			len(mudas), rutaSinDeclarar, strings.Join(mudas, "\n  "))
	}
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// handlersDelTransporte resuelve, por AST, las dos mitades que el emparejado necesita:
//
//	rutaHandler["GET /api/sessions"] = "listSessions"      (del registro en el mux)
//	queriesDe["listSessions"]        = ["arnes","cerradas"] (del cuerpo de esa función)
//
// El segundo argumento del registro es casi siempre una llamada al constructor del handler
// (`listSessions(sessions)`); cuando es un identificador pelado (`healthz`) o un valor
// (`events`), se toma ese nombre. Lo que no se puede resolver a un nombre no se empareja: es
// preferible no afirmar nada a afirmar de más.
func handlersDelTransporte(t *testing.T) (rutaHandler map[string]string, queriesDe map[string][]string) {
	t.Helper()
	root := repoRoot()
	dir := filepath.Join(root, dirTransporte)
	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", dirTransporte, err)
	}
	// `parser.ParseDir` está deprecado desde Go 1.25 (no mira build tags), así que se parsea
	// archivo por archivo — que además es lo único que este escáner necesita.
	fset := token.NewFileSet()
	var archivos []*ast.File
	for _, e := range entradas {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, n), nil, 0)
		if err != nil {
			t.Fatalf("no se pudo parsear %s: %v", n, err)
		}
		archivos = append(archivos, f)
	}

	rutaHandler = map[string]string{}
	queriesDe = map[string][]string{}
	{
		for _, f := range archivos {
			// (a) los registros del mux.
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) < 2 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || (sel.Sel.Name != "Handle" && sel.Sel.Name != "HandleFunc") {
					return true
				}
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				ruta := strings.Trim(lit.Value, `"`)
				if nombre := identDe(call.Args[1]); nombre != "" {
					rutaHandler[ruta] = nombre
				}
				return true
			})
			// (b) los queries que lee el cuerpo de cada función de nivel superior. Se recorre
			// el cuerpo ENTERO, así que las closures que la función devuelve —que es donde el
			// handler real vive— quedan incluidas.
			for _, d := range f.Decls {
				fn, ok := d.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				var qs []string
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok || len(call.Args) != 1 {
						return true
					}
					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok || sel.Sel.Name != "Get" {
						return true
					}
					// El receptor tiene que ser `<algo>.URL.Query()`.
					inner, ok := sel.X.(*ast.CallExpr)
					if !ok {
						return true
					}
					isel, ok := inner.Fun.(*ast.SelectorExpr)
					if !ok || isel.Sel.Name != "Query" {
						return true
					}
					usel, ok := isel.X.(*ast.SelectorExpr)
					if !ok || usel.Sel.Name != "URL" {
						return true
					}
					lit, ok := call.Args[0].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						return true
					}
					qs = append(qs, strings.Trim(lit.Value, `"`))
					return true
				})
				if len(qs) > 0 {
					sort.Strings(qs)
					queriesDe[fn.Name.Name] = dedup(qs)
				}
			}
		}
	}
	if len(rutaHandler) == 0 {
		t.Fatalf("cero handlers resueltos en %s — el AST-scan se rompió, no el router", dirTransporte)
	}
	return rutaHandler, queriesDe
}

// identDe saca el nombre del handler de la expresión del registro: `f(x)` → "f", `f` → "f".
func identDe(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.CallExpr:
		return identDe(v.Fun)
	case *ast.SelectorExpr:
		return v.Sel.Name
	}
	return ""
}
