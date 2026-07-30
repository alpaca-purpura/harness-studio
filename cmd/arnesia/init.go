package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"

	doctrina "github.com/alpacapurpura/arnesia"
	"github.com/alpacapurpura/arnesia/internal/adapters/forja"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// errSemillaInsana hace exit≠0 cuando la salud no es `sana` — «no avanzamos si no está
// sana» (A-D4). El detalle ya salió por stdout; esto solo fija el código de salida.
var errSemillaInsana = errors.New("arnesia init: la semilla no está sana")

// runInit siembra el process-as-code `.arnesia/` en un proyecto (contrato
// semilla-arnesia.md; paquete 2026-07-30-arnesia-en-el-proyecto, A-T3):
//
//	arnesia init [--check] [--json] [dir]
//
// Sin --check: siembra Y chequea al final; con --check: solo el doctor. En ambos, la
// salud manda el exit code. La fuente es SIEMPRE la semilla embebida en el binario
// (A-D2); leer el arnes.yaml DEL PROYECTO es Fase 2 — TODO(--arnes-yaml), gap declarado
// en el contrato §5 y en la hoja capability forja/sembrar-semilla.yaml.
func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	check := fs.Bool("check", false, "solo chequea la salud de la semilla (doctor v0); no siembra")
	jsonOut := fs.Bool("json", false, "emite el resultado como JSON")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia init [--check] [--json] [dir]

Siembra el process-as-code .arnesia/ (terreno · product · wip · plantillas de proceso)
en el proyecto `+"`dir`"+` (default: el directorio actual). Idempotente: lo existente
JAMÁS se pisa (se reporta ya-existia). Exit ≠ 0 si la salud no queda sana.

  --check   solo el doctor: sana|ausente|incompleta + faltantes; no escribe nada.
  --json    salida JSON (informe + salud).
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := "."
	if fs.NArg() > 0 {
		dir = fs.Arg(0)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	semillaFS, err := iofs.Sub(doctrina.Semilla, "semilla")
	if err != nil {
		return fmt.Errorf("semilla embebida: %w", err)
	}
	svc := usecase.NewForjaService(forja.New(semillaFS))
	ctx := context.Background()

	if *check {
		salud, cerr := svc.Chequear(ctx, abs)
		if cerr != nil {
			return cerr
		}
		if perr := imprimirSalud(salud, *jsonOut); perr != nil {
			return perr
		}
		if salud.Estado != domain.SemillaSana {
			return errSemillaInsana
		}
		return nil
	}

	informe, serr := svc.Sembrar(ctx, abs)
	if serr != nil {
		return serr
	}
	salud, cerr := svc.Chequear(ctx, abs)
	if cerr != nil {
		return cerr
	}
	if *jsonOut {
		if perr := imprimirJSON(struct {
			Informe domain.InformeSemilla `json:"informe"`
			Salud   domain.SaludSemilla   `json:"salud"`
		}{informe, salud}); perr != nil {
			return perr
		}
	} else {
		fmt.Printf("semilla en %s\n", filepath.Join(abs, ".arnesia"))
		fmt.Printf("  creados: %d\n", len(informe.Creados))
		for _, r := range informe.Creados {
			fmt.Printf("    + %s\n", r)
		}
		fmt.Printf("  ya existían: %d\n", len(informe.YaExistian))
		for _, r := range informe.YaExistian {
			fmt.Printf("    = %s\n", r)
		}
		imprimirSaludTexto(salud)
	}
	if salud.Estado != domain.SemillaSana {
		return errSemillaInsana
	}
	return nil
}

// imprimirSalud emite la salud en JSON o texto según el flag.
func imprimirSalud(salud domain.SaludSemilla, jsonOut bool) error {
	if jsonOut {
		return imprimirJSON(salud)
	}
	imprimirSaludTexto(salud)
	return nil
}

// imprimirSaludTexto emite el veredicto del doctor en texto plano, con los faltantes
// LISTADOS (contrato §4: el veredicto lista, jamás resume en un adjetivo).
func imprimirSaludTexto(salud domain.SaludSemilla) {
	fmt.Printf("salud: %s\n", salud.Estado)
	if salud.Detalle != "" {
		fmt.Printf("  %s\n", salud.Detalle)
	}
	for _, f := range salud.Faltantes {
		fmt.Printf("  - falta %s\n", f)
	}
}
