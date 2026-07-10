//go:build ignore

// Desechable HS-17: fuerza una Provision() real contra ~/.arnesia con el código de
// producción (no una copia a mano) para que la sonda post-MCP tenga un mcp.json real
// que apuntar con --mcp-config. `//go:build ignore` lo saca de `go build ./...`/lint.
package main

import (
	"context"
	"fmt"
	"os"

	doctrina "github.com/alpacapurpura/arnesia"
	"github.com/alpacapurpura/arnesia/internal/adapters/provision"
)

func main() {
	p, err := provision.New("", doctrina.Kit, doctrina.Files)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	inj, err := p.Provision(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("PluginDirs=%v\nSystemPromptFile=%s\nAddDirs=%v\nMCPConfigFile=%s\n",
		inj.PluginDirs, inj.SystemPromptFile, inj.AddDirs, inj.MCPConfigFile)
}
