// frontmatter.go — equivalente de gray-matter + fs-writer (write atómico).
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type parsedDoc struct {
	Frontmatter map[string]any
	Content     string
	Raw         string
}

// parseYAMLMap convierte YAML a map[string]any (yaml.v3 ya usa string keys).
func parseYAMLMap(raw []byte) (map[string]any, error) {
	var m map[string]any
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// parseFrontmatter replica gray-matter: bloque `---\n...\n---` al inicio,
// resto = content. Sin bloque → frontmatter vacío + todo es content.
func parseFrontmatter(raw string) (parsedDoc, error) {
	doc := parsedDoc{Frontmatter: map[string]any{}, Raw: raw, Content: raw}

	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") && normalized != "---" {
		return doc, nil
	}
	rest := normalized[4:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return doc, nil
	}
	fmBlock := rest[:end]
	after := rest[end+4:]
	// gray-matter consume el newline que sigue al cierre
	after = strings.TrimPrefix(after, "\n")

	fm, err := parseYAMLMap([]byte(fmBlock))
	if err != nil {
		doc.Content = normalized
		return doc, fmt.Errorf("frontmatter YAML inválido: %w", err)
	}
	if fm == nil {
		fm = map[string]any{}
	}
	doc.Frontmatter = fm
	doc.Content = after
	return doc, nil
}

func readMarkdownWithFrontmatter(absPath string) (parsedDoc, error) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return parsedDoc{}, err
	}
	return parseFrontmatter(string(raw))
}

// stringifyFrontmatter replica matter.stringify: `---\n{yaml}---\n{body}`.
// keyOrder fija el orden de serialización (gray-matter preserva inserción JS);
// keys no listadas van al final en alfabético.
func stringifyFrontmatter(frontmatter map[string]any, body string, keyOrder []string) (string, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	emitted := map[string]bool{}

	appendKV := func(k string) error {
		v, ok := frontmatter[k]
		if !ok || emitted[k] {
			return nil
		}
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}
		valNode := &yaml.Node{}
		if err := valNode.Encode(v); err != nil {
			return err
		}
		node.Content = append(node.Content, keyNode, valNode)
		emitted[k] = true
		return nil
	}

	for _, k := range keyOrder {
		if err := appendKV(k); err != nil {
			return "", err
		}
	}
	var rest []string
	for k := range frontmatter {
		if !emitted[k] {
			rest = append(rest, k)
		}
	}
	// estable: alfabético para las no listadas
	for i := 0; i < len(rest); i++ {
		for j := i + 1; j < len(rest); j++ {
			if rest[j] < rest[i] {
				rest[i], rest[j] = rest[j], rest[i]
			}
		}
	}
	for _, k := range rest {
		if err := appendKV(k); err != nil {
			return "", err
		}
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return "", err
	}
	_ = enc.Close()

	yamlStr := buf.String()
	if !strings.HasSuffix(yamlStr, "\n") {
		yamlStr += "\n"
	}
	if !strings.HasPrefix(body, "\n") && body != "" {
		body = "\n" + body
	}
	return "---\n" + yamlStr + "---" + body, nil
}

// frontmatterKeyOrder extrae el orden actual de keys del bloque frontmatter
// de un archivo existente (para reescrituras que preservan orden).
func frontmatterKeyOrder(raw string) []string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return nil
	}
	rest := normalized[4:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return nil
	}
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(rest[:end]), &node); err != nil {
		return nil
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil
	}
	m := node.Content[0]
	var keys []string
	for i := 0; i+1 < len(m.Content); i += 2 {
		keys = append(keys, m.Content[i].Value)
	}
	return keys
}

// writeFileAtomic: write a {path}.tmp + rename (paridad con fs-writer.ts).
func writeFileAtomic(absPath, content string) error {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return err
	}
	tmp := absPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, absPath)
}

func writeMarkdownWithFrontmatter(absPath string, frontmatter map[string]any, body string, keyOrder []string) error {
	serialized, err := stringifyFrontmatter(frontmatter, body, keyOrder)
	if err != nil {
		return err
	}
	return writeFileAtomic(absPath, serialized)
}

// coalesce devuelve el primer valor no-nil/no-ausente.
func fmGet(fm map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := fm[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func fmGetOr(fm map[string]any, fallback any, keys ...string) any {
	if v := fmGet(fm, keys...); v != nil {
		return v
	}
	return fallback
}
