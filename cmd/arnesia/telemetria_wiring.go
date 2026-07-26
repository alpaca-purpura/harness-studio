package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	telcatalogo "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/descubrimiento"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/forward"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/hooks"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/otlp"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// telemetria_wiring.go es el composition root del módulo de telemetría (§10.1).
//
// **Todo lo de acá degrada honesto.** Si el almacén no abre, el daemon SIGUE sirviendo
// `/healthz` y la API, y la salud dice que la telemetría no está disponible **con motivo**. Se
// pierde telemetría, jamás la sesión del usuario: la instrumentación no puede degradar al
// sistema instrumentado.

// cablesTelemetria son las dependencias que el módulo toma del resto del daemon.
type cablesTelemetria struct {
	retencionDias    int
	forwardEndpoint  string
	refrescoCatalogo bool
	idx              *usecase.MapService
	portafolio       *usecase.PortafolioService
	roleFor          func(ctx context.Context, arnesID string) string
}

// cablearTelemetria arma el módulo entero y devuelve el servicio, el handler OTLP y el cierre.
// Un fallo del almacén **no** devuelve error: devuelve `nil` y el daemon arranca igual.
func cablearTelemetria(ctx context.Context, c cablesTelemetria) (*usecase.TelemetriaService, http.Handler, func()) {
	store, err := telstore.New("", telstore.Opciones{RetencionDias: c.retencionDias})
	if err != nil {
		// El daemon SIGUE. Sin almacén no hay telemetría, pero sí hay producto.
		slog.Warn("telemetria: almacén no disponible — el daemon arranca sin telemetría", "err", err)
		return nil, nil, nil
	}
	if err := telcatalogo.ErrEmbebido(); err != nil {
		// Un catálogo roto degrada el COSTO CALCULADO, no el módulo: el costo reportado por
		// el runtime sigue llegando. Se avisa y se sigue.
		slog.Warn("telemetria: catálogo de precios ilegible — el costo calculado quedará sin dato", "err", err)
	}
	cat := telcatalogo.Embebido()
	if c.refrescoCatalogo {
		// A11: el refresco es egreso del daemon y va apagado por default. Cuando el operador
		// lo enciende, todavía no hay descarga cableada — y eso se DICE en vez de fingir que
		// se refrescó.
		slog.Warn("telemetria: --telemetria-catalogo-refresco encendido, pero la descarga no está construida; " +
			"se sigue con el catálogo embebido y la salud lo reporta como nunca refrescado")
	}

	rollup := telstore.NewRollup(store, 2*time.Second)
	fwd := forward.New(c.forwardEndpoint, nil)

	// El registro de atribución cruza el almacén (tabla de huellas aprendidas) con el
	// Portafolio (huella de ruta → instalación conocida).
	reg := usecase.NuevoRegistroAtribucion(
		store.BuscarHash,
		func(huella string) (string, string, bool) { return resolverHuella(ctx, c.portafolio, huella) },
		store.AprenderHash,
	)

	svc := usecase.NewTelemetriaService(store, cat, reg, fwd,
		domain.DetectoresMVP(), otlp.PerfilClaudeCode(), time.Now)
	svc.SetRetencion(store, rollup, c.retencionDias, usecase.RollupDefaultMeses)
	svc.SetPortafolio(c.roleFor, func(ictx context.Context) []usecase.FilaInstalacion {
		return instalacionesDelPortafolio(ictx, c.portafolio)
	})

	receptor := otlp.NewReceptor(svc, otlp.Opciones{MaxBody: otlp.MaxBodyDefault})
	// Los contadores del receptor se PERSISTEN: «cuántos descarté» es dato de honestidad y
	// tiene que sobrevivir a un reinicio, no ser un gauge en memoria que se pierde.
	receptor.AlContar(func(clave string, n int64) { store.SumarSalud(ctx, clave, n) })

	// Purga al boot y cada 6 h.
	go purgarPeriodicamente(ctx, svc)
	// Rollup al boot: el tablero de la primera consulta ya sale del agregado.
	if aerr := rollup.Actualizar(ctx); aerr != nil {
		slog.Warn("telemetria: rollup inicial no aplicado", "err", aerr)
	}

	cerrar := func() {
		rollup.Detener()
		if cerr := store.Close(); cerr != nil {
			slog.Warn("telemetria: cierre del almacén", "err", cerr)
		}
	}
	return svc, receptor, cerrar
}

// purgarPeriodicamente aplica el TTL al boot y cada 6 h. Un fallo se loguea y se reintenta a
// la vuelta: la retención no puede tumbar el daemon.
func purgarPeriodicamente(ctx context.Context, svc *usecase.TelemetriaService) {
	aplicar := func() {
		if n, err := svc.Purgar(ctx, portsPurgaVacia()); err != nil {
			slog.Warn("telemetria: purga por TTL no aplicada", "err", err)
		} else if n > 0 {
			slog.Info("telemetria: purga por TTL", "filas", n, "dias", svc.RetencionDias())
		}
	}
	aplicar()
	t := time.NewTicker(6 * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			aplicar()
		}
	}
}

// mintTokenIngesta genera el SEGUNDO token (A6). Es distinto del de la API a propósito:
// filtrarlo concede «escribime telemetría», nunca «conducí un agente».
func mintTokenIngesta() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		// Sin aleatoriedad no se fabrica un token débil: se deja vacío, y entonces el gate
		// cae al token de la API. Es menos cómodo y más honesto que un secreto predecible.
		slog.Warn("telemetria: no se pudo generar el token de ingesta", "err", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// publicarFicha escribe la ficha del daemon DESPUÉS de que el listener acepta (A8) y devuelve
// el retiro para el shutdown ordenado.
//
// Un directorio de config no escribible **no tumba el daemon**: se avisa y la salud lo dice.
// S1 y `s2-instrumentado` siguen funcionando — van por env y por settings, no por ficha.
func publicarFicha(ctx context.Context, ln net.Listener, tokenIngesta string) (string, func()) {
	f, err := descubrimiento.New("")
	if err != nil {
		slog.Warn("telemetria: ficha del daemon no publicada", "err", err)
		return "no publicada: " + err.Error(), nil
	}
	endpoint := "http://" + ln.Addr().String()
	d := domain.FichaDaemon{
		Version: 1, PID: os.Getpid(), Endpoint: endpoint,
		RutaProceso: "/api/telemetria/proceso", TokenIngesta: tokenIngesta,
		Desde: time.Now().UTC(),
	}
	d.OTLP.Logs = "/v1/logs"
	d.OTLP.Metrics = "/v1/metrics"
	if bin, berr := os.Executable(); berr == nil {
		d.Binario = bin
	}
	if perr := f.Publicar(ctx, d); perr != nil {
		slog.Warn("telemetria: ficha del daemon no publicada — el hook no va a encontrarlo", "err", perr)
		return "no publicada: " + perr.Error(), nil
	}
	slog.Info("telemetria: ficha del daemon publicada", "ruta", f.Ruta(), "endpoint", endpoint)
	return f.Ruta(), func() {
		if rerr := f.Retirar(); rerr != nil {
			slog.Warn("telemetria: ficha no retirada", "err", rerr)
		}
	}
}

// instalacionesDelPortafolio proyecta el Portafolio a lo mínimo que la tabla de telemetría
// necesita. Vive acá y no en el caso de uso para no atar `telemetria` al modelo del
// Portafolio: si mañana cambia, cambia esta función.
func instalacionesDelPortafolio(ctx context.Context, pf *usecase.PortafolioService) []usecase.FilaInstalacion {
	if pf == nil {
		return nil
	}
	entradas, _, err := pf.Listar(ctx)
	if err != nil {
		slog.Warn("telemetria: Portafolio no legible — la tabla por arnés queda vacía", "err", err)
		return nil
	}
	var out []usecase.FilaInstalacion
	for _, e := range entradas {
		clave := e.Identidad.Clave()
		if len(e.Instalaciones) == 0 {
			out = append(out, usecase.FilaInstalacion{
				ArnesID: e.Identidad.ID, Clave: clave, Nombre: e.Nombre, Empresas: e.Empresas,
			})
			continue
		}
		for _, inst := range e.Instalaciones {
			out = append(out, usecase.FilaInstalacion{
				ArnesID: e.Identidad.ID, InstalacionID: inst.ProyectoPath,
				Clave: clave, Nombre: e.Nombre, Empresas: e.Empresas,
			})
		}
	}
	return out
}

// resolverHuella mapea la huella de un `cwd` a una instalación conocida.
//
// El hook manda la HUELLA, no la ruta (A14), así que la comparación se hace hasheando cada
// ruta conocida del Portafolio con la misma función. La ruta del usuario nunca entra al
// almacén, ni siquiera acá.
func resolverHuella(ctx context.Context, pf *usecase.PortafolioService, huella string) (string, string, bool) {
	if pf == nil || huella == "" {
		return "", "", false
	}
	entradas, _, err := pf.Listar(ctx)
	if err != nil {
		return "", "", false
	}
	for _, e := range entradas {
		for _, inst := range e.Instalaciones {
			if hooks.HuellaCWD(inst.ProyectoPath) == huella {
				return e.Identidad.ID, inst.ProyectoPath, true
			}
		}
	}
	return "", "", false
}

// portsPurgaVacia es una purga por TTL (sin arnés): el caso normal del temporizador.
func portsPurgaVacia() ports.PurgaTelemetria { return ports.PurgaTelemetria{} }
