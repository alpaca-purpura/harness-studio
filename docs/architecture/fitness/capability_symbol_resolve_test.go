package fitness

// Resolución de símbolo para R1 `cap-ptr-resuelve` (BACKLOG «R1 a nivel símbolo», auditoría
// 2026-07-14, doctrina codigo-traza-a-capability v1.4): hasta acá el enforcer solo stat-eaba el
// ARCHIVO del puntero `file#Símbolo`, nunca el `#Símbolo` — un símbolo renombrado pasaba en
// silencio. Este archivo agrega el resolver real por lenguaje (Go via go/parser; TS/TSX/Rust vía
// regex de declaración top-level — no hay parser TS/Rust en la stdlib de Go, y un regex de
// declaración es proporcional a lo que R1 necesita: ¿existe ese identificador declarado?).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

// capPointerParts separa un token de puntero en (ruta, símbolo, ok). ok=false si la ruta no
// termina en una extensión de código fuente. símbolo="" si el token no lleva `#Símbolo`
// (formas legítimas sin símbolo: archivo completo de un módulo pequeño, o `paquete/`).
func capPointerParts(tok string) (file, symbol string, ok bool) {
	p := tok
	if i := strings.IndexAny(tok, "#:,"); i >= 0 {
		p = tok[:i]
		if tok[i] == '#' {
			symbol = strings.TrimSpace(tok[i+1:])
		}
	}
	p = strings.TrimSpace(p)
	for _, e := range capSourceExts {
		if strings.HasSuffix(p, e) {
			return p, symbol, true
		}
	}
	return "", "", false
}

// capSymbolResolves reporta si `symbol` está declarado en el archivo `absPath`. Despacha por
// extensión; devuelve (resuelve, razón-si-no).
func capSymbolResolves(absPath, symbol string) (bool, string) {
	switch {
	case strings.HasSuffix(absPath, ".go"):
		return goSymbolExists(absPath, symbol)
	case strings.HasSuffix(absPath, ".ts"), strings.HasSuffix(absPath, ".tsx"):
		return tsSymbolExists(absPath, symbol)
	case strings.HasSuffix(absPath, ".rs"):
		return regexSymbolExists(absPath, symbol, rsDeclRe)
	default:
		return true, "" // extensión no cubierta por R1 (no forma parte de capSourceExts)
	}
}

// goSymbolExists parsea el archivo Go real (go/parser, sin type-check — más rápido, alcanza
// para nombres declarados) y busca `symbol` entre: funcs top-level, métodos como `Tipo.Método`
// (receptor, incluso genérico `Tipo[T]`), types, consts, vars.
func goSymbolExists(absPath, symbol string) (bool, string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		return false, "no se pudo parsear Go: " + err.Error()
	}
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == symbol {
				return true, ""
			}
			if d.Recv != nil && len(d.Recv.List) > 0 {
				if tn := goRecvTypeName(d.Recv.List[0].Type); tn != "" && tn+"."+d.Name.Name == symbol {
					return true, ""
				}
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name == symbol {
						return true, ""
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name == symbol {
							return true, ""
						}
					}
				}
			}
		}
	}
	return false, "símbolo no declarado en el archivo"
}

// goRecvTypeName desenvuelve `*T`/`T`/`T[Param]` (genérico) hasta el Ident del tipo receptor.
func goRecvTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return goRecvTypeName(t.X)
	case *ast.IndexExpr:
		return goRecvTypeName(t.X)
	case *ast.IndexListExpr:
		return goRecvTypeName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// tsDeclRe: declaraciones top-level (function/class/interface/type/enum/const/let/var NAME).
var tsDeclRe = regexp.MustCompile(
	`(?m)^\s*(?:export\s+)?(?:default\s+)?(?:declare\s+)?(?:async\s+)?` +
		`(?:function\s*\*?|class|interface|type|enum|const|let|var)\s+([A-Za-z_$][\w$]*)`,
)

// tsMemberRe: miembro de interface/type/object-literal (`nombre: tipo`/`nombre: valor,`) o método
// shorthand (`nombre(args) {`) — cubre el patrón de store Zustand (`create<S>((set,get) => ({
// accion: (...) => {...} }))`) y las firmas de interface que documentan la misma forma.
var tsMemberRe = regexp.MustCompile(
	`(?m)^\s*(?:readonly\s+)?(?:public\s+|private\s+|protected\s+)?(?:async\s+)?(?:get\s+|set\s+)?([A-Za-z_$][\w$]*)\??\s*[:(]`,
)

// tsImportRe: especificadores de `import { A, B as C } from "..."` — el identificador queda
// genuinamente ligado en ese archivo aunque no esté "declarado" ahí (reexport/uso, patrón visto
// en punteros a `.test.ts` que ejercitan un símbolo importado, o componentes que reciben una
// constante de otro módulo como prop).
var tsImportRe = regexp.MustCompile(`(?m)^\s*import\s+(?:type\s+)?\{([^}]*)\}\s*from\s*['"][^'"]+['"]`)

// rsDeclRe: declaraciones top-level de Rust (fn/struct/enum/trait/const/static/mod/type).
var rsDeclRe = regexp.MustCompile(
	`(?m)^\s*(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?` +
		`(?:fn|struct|enum|trait|const|static(?:\s+mut)?|mod|type)\s+([A-Za-z_][\w]*)`,
)

// tsSymbolExists escanea declaraciones top-level, miembros y especificadores de import.
func tsSymbolExists(absPath, symbol string) (bool, string) {
	b, err := os.ReadFile(absPath) //nolint:gosec // rutas del propio árbol del repo
	if err != nil {
		return false, "no se pudo leer: " + err.Error()
	}
	content := string(b)
	for _, re := range [...]*regexp.Regexp{tsDeclRe, tsMemberRe} {
		for _, m := range re.FindAllStringSubmatch(content, -1) {
			if m[1] == symbol {
				return true, ""
			}
		}
	}
	for _, m := range tsImportRe.FindAllStringSubmatch(content, -1) {
		for _, spec := range strings.Split(m[1], ",") {
			for _, name := range strings.Fields(spec) { // "Foo as Bar" -> ambos nombres cuentan
				if name != "as" && name == symbol {
					return true, ""
				}
			}
		}
	}
	return false, "símbolo no declarado (ni top-level, ni miembro, ni import) en el archivo"
}

// regexSymbolExists escanea el archivo con `re` y compara cada nombre capturado contra `symbol`.
func regexSymbolExists(absPath, symbol string, re *regexp.Regexp) (bool, string) {
	b, err := os.ReadFile(absPath) //nolint:gosec // rutas del propio árbol del repo
	if err != nil {
		return false, "no se pudo leer: " + err.Error()
	}
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		if m[1] == symbol {
			return true, ""
		}
	}
	return false, "símbolo no declarado en el archivo (regex de declaración top-level)"
}
