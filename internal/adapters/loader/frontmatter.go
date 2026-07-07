package loader

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// frontmatter extrae y parsea el bloque YAML delimitado por los dos `---` iniciales de un
// archivo markdown (el formato oficial de SKILL.md). Devuelve error si el archivo no abre
// con `---`, si el bloque no cierra o si el YAML no parsea — el caller decide qué hacer con
// el error (para skills: nodo no-reconocido §4.5; para CLAUDE.md: probar la convención de
// heading). Usa gopkg.in/yaml.v3: los mapas anidados salen como map[string]any, directamente
// serializables a JSON para mapear al dominio.
func frontmatter(b []byte) (map[string]any, error) {
	s := strings.TrimPrefix(string(b), "\ufeff") // BOM fuera; el resto se respeta tal cual.
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return nil, errors.New("sin frontmatter YAML inicial (---)")
	}
	cuerpo := s[strings.Index(s, "\n")+1:]

	// El cierre es una línea que consiste EXACTAMENTE en `---` (tolerando \r de CRLF).
	lineas := strings.Split(cuerpo, "\n")
	cierre := -1
	for i, l := range lineas {
		if strings.TrimRight(l, "\r") == "---" {
			cierre = i
			break
		}
	}
	if cierre < 0 {
		return nil, errors.New("frontmatter sin línea de cierre ---")
	}

	var m map[string]any
	if err := yaml.Unmarshal([]byte(strings.Join(lineas[:cierre], "\n")), &m); err != nil {
		return nil, fmt.Errorf("frontmatter YAML inválido: %w", err)
	}
	if m == nil {
		return nil, errors.New("frontmatter vacío")
	}
	return m, nil
}
