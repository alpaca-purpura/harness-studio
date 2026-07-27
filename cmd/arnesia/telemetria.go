package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	telcatalogo "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// telemetria.go es `arnesia telemetria <subcomando>` (decisión A17), mismo patrón que
// `arnesia portafolio`: **reusa el MISMO caso de uso que el HTTP, cero lógica propia.**
//
// Para qué existe: es la superficie observable del módulo **sin FE**. Mientras el mockup de la
// capa Mejora no tenga su firma, esto es lo que demuestra que el backend entrega valor — y
// después sigue siendo la vía de verificación E2E.
//
// Dos implementaciones de la misma consulta divergen el día que alguien arregla una sola. Por
// eso acá no hay ni una línea de agregación: se llama al servicio y se imprime.

func runTelemetria(args []string) error {
	if len(args) == 0 {
		usoTelemetria()
		return fmt.Errorf("telemetria: falta el subcomando")
	}
	fs := flag.NewFlagSet("telemetria", flag.ExitOnError)
	arnes := fs.String("arnes", "", "acotar a un arnés")
	instalacion := fs.String("instalacion", "", "acotar a una instalación")
	desde := fs.String("desde", "", "inicio de la ventana (RFC3339)")
	hasta := fs.String("hasta", "", "fin de la ventana (RFC3339)")
	dias := fs.Int("retencion", usecase.RetencionDefaultDias,
		"días de retención al purgar (firmado: 90 — D26.3)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	q := ports.ConsultaTelemetria{ArnesID: *arnes, InstalacionID: *instalacion}
	if *desde != "" {
		t, err := time.Parse(time.RFC3339, *desde)
		if err != nil {
			return fmt.Errorf("telemetria: --desde se espera RFC3339: %w", err)
		}
		q.Desde = t
	}
	if *hasta != "" {
		t, err := time.Parse(time.RFC3339, *hasta)
		if err != nil {
			return fmt.Errorf("telemetria: --hasta se espera RFC3339: %w", err)
		}
		q.Hasta = t
	}

	// `catalogo` no necesita abrir el almacén: es el único subcomando que lee solo el
	// binario. Se atiende antes para no crear un `.db` cuando nadie lo pidió.
	if args[0] == "catalogo" {
		return imprimir(telcatalogo.Embebido().Version())
	}

	ctx := context.Background()
	svc, cerrar, err := servicioTelemetriaCLI(*dias)
	if err != nil {
		return err
	}
	defer cerrar()

	switch args[0] {
	case "resumen":
		r, err := svc.Resumen(ctx, q)
		if err != nil {
			return err
		}
		return imprimir(r)

	case "mejoras":
		r, err := svc.Mejoras(ctx, q)
		if err != nil {
			return err
		}
		return imprimir(r)

	case "salud":
		r, err := svc.Salud(ctx)
		if err != nil {
			return err
		}
		return imprimir(r)

	case "cajas":
		r, err := svc.PorCaja(ctx, q)
		if err != nil {
			return err
		}
		if r == nil {
			r = []domain.GastoCaja{}
		}
		return imprimir(r)

	case "portafolio":
		r, err := svc.Portafolio(ctx, q)
		if err != nil {
			return err
		}
		if r == nil {
			r = []domain.FilaPortafolio{}
		}
		return imprimir(r)

	case "purgar":
		n, err := svc.Purgar(ctx, ports.PurgaTelemetria{ArnesID: *arnes})
		if err != nil {
			return err
		}
		return imprimir(map[string]any{
			"borrados":       n,
			"arnes":          *arnes,
			"retencion_dias": svc.RetencionDias(),
		})

	default:
		usoTelemetria()
		return fmt.Errorf("telemetria: subcomando desconocido %q", args[0])
	}
}

// servicioTelemetriaCLI arma el MISMO servicio que el daemon, contra el MISMO almacén. Sin
// receptor ni ficha: el CLI lee, no recibe.
//
// El `Portafolio` y el índice no se cablean acá a propósito — abrirlos duplicaría el estado
// del daemon vivo. La consecuencia se declara: `arnesia telemetria portafolio` devuelve la
// lista vacía cuando se corre sin daemon, y el HTTP es el que la sirve completa.
func servicioTelemetriaCLI(dias int) (*usecase.TelemetriaService, func(), error) {
	st, err := telstore.New("", telstore.Opciones{RetencionDias: dias})
	if err != nil {
		return nil, nil, fmt.Errorf("telemetria: almacén: %w", err)
	}
	roll := telstore.NewRollup(st, time.Hour)
	reg := usecase.NuevoRegistroAtribucion(st.BuscarHash, nil, st.AprenderHash)
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), reg, nil,
		domain.DetectoresMVP(), perfilCLI(), time.Now)
	svc.SetRetencion(st, roll, dias, usecase.RollupDefaultMeses)
	cerrar := func() {
		roll.Detener()
		_ = st.Close()
	}
	return svc, cerrar, nil
}

// perfilCLI es el perfil del runtime para el costeo del CLI. Se declara acá y no se importa
// de `otlp` para no arrastrar el receptor a un binario que solo lee.
func perfilCLI() domain.PerfilRuntime {
	return domain.PerfilRuntime{
		Runtime:     "claude-code",
		Aritmetica:  domain.AritmeticaDisjunta,
		Acumulacion: domain.AcumulacionPorRequest,
	}
}

// imprimir emite JSON indentado a stdout. El CLI es una vía de INSPECCIÓN: imprime.
func imprimir(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func usoTelemetria() {
	fmt.Fprint(os.Stderr, `arnesia telemetria — inspección del módulo de telemetría, sin FE

uso: arnesia telemetria <subcomando> [flags]

subcomandos:
  resumen     total de la ventana, con su cobertura y su confianza
  cajas       gasto por caja, INCLUIDAS las que no tienen dato atribuible
  mejoras     puntos de mejora + los que no aplican (con motivo) + los que no medimos
  portafolio  una fila por arnés x instalación
  salud       contadores del receptor, retención, catálogo y estado del almacén
  purgar      aplica el TTL, o borra todo lo de un arnés con --arnes
  catalogo    version, revisión de origen, huella y número de modelos

flags:
  --arnes X        acotar a un arnés
  --instalacion X  acotar a una instalación
  --desde / --hasta  ventana en RFC3339
  --retencion N    días de retención al purgar (firmado: 90)
`)
}
