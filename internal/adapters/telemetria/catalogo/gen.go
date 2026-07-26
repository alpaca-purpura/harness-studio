package catalogo

// gen.go es la procedencia del catálogo, en un archivo con nombre evidente: quien vaya a
// actualizar precios entra por acá.
//
//	go generate ./internal/adapters/telemetria/catalogo/
//
// baja el catálogo del rev FIJADO de abajo, lo filtra y reescribe `precios.json`. Es un acto
// deliberado —con su diff a la vista— y NO un paso del build: el binario se compila offline y
// reproducible (D11, cero post-install).

//go:generate go run ./cmd/bajar-catalogo -rev b439a9a78864f52fba3d18d9bf8b09253e7181d7 -fecha 2026-07-26 -out precios.json

const (
	// revLiteLLM es el commit de BerriAI/litellm del que salió `precios.json`. Está acá y
	// no solo en la directiva de arriba para que un test pueda contrastarlo contra el
	// `rev` que el propio archivo declara: si alguien regenera con otro rev y se olvida de
	// actualizar la directiva, el drift se caza en CI en vez de vivir dos meses.
	revLiteLLM = "b439a9a78864f52fba3d18d9bf8b09253e7181d7"
	// fechaLiteLLM es la fecha de ese commit — la `version` que se muestra en pantalla.
	fechaLiteLLM = "2026-07-26"
)
