package otlp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// receptor.go es el `http.Handler` de `/v1/logs` y `/v1/metrics`.
//
// **El emisor JAMÁS se bloquea** (decisión A12). Un receptor que hace esperar al agente
// degradaría el trabajo que está midiendo, que es la peor forma posible de instrumentar. De
// ahí: cola acotada, descarte contado, y `partialSuccess` en la respuesta — que además es lo
// que la spec OTLP pide.
//
// Y **nunca un 5xx por un evento malo**: un 5xx hace reintentar al exportador y multiplica el
// daño. Un lote con forma inválida es 400 y se acabó.

// MaxBodyDefault es el tope de cuerpo (§11). El mayor payload observado en vivo son 24 KB:
// 4 MiB deja 170× de margen y aun así corta un cuerpo absurdo antes de leerlo a memoria.
const MaxBodyDefault int64 = 4 << 20

// Opciones parametriza el receptor.
type Opciones struct {
	MaxBody int64
	Reloj   func() time.Time
	// Perfil es el del runtime que emite. Se estampa en cada evento (A18).
	Perfil domain.PerfilRuntime
}

// Contadores son los descartes que el receptor cuenta. Se exponen para que el composition
// root los persista: un descarte que nadie cuenta es una señal perdida sin rastro.
type Contadores struct {
	Recibidos         atomic.Int64
	RechazadosFormato atomic.Int64
	RechazadosTamano  atomic.Int64
	ColaLlena         atomic.Int64
	FueraDeLista      atomic.Int64
	Temporalidad      atomic.Int64
}

// Receptor es el único lugar del árbol que habla el wire OTLP por HTTP.
type Receptor struct {
	sink  ports.TelemetriaSink
	opts  Opciones
	salud *Contadores
	// persistir sube los contadores al almacén. Puede ser nil en tests.
	persistir func(clave string, n int64)
}

// NewReceptor arma el receptor. `sink` es la única salida: este paquete no conoce el almacén.
func NewReceptor(sink ports.TelemetriaSink, o Opciones) *Receptor {
	if o.MaxBody <= 0 {
		o.MaxBody = MaxBodyDefault
	}
	if o.Reloj == nil {
		o.Reloj = time.Now
	}
	if o.Perfil.Runtime == "" {
		o.Perfil = PerfilClaudeCode()
	}
	return &Receptor{sink: sink, opts: o, salud: &Contadores{}}
}

// Salud expone los contadores vivos.
func (r *Receptor) Salud() *Contadores { return r.salud }

// AlContar registra el sumidero de contadores persistidos.
func (r *Receptor) AlContar(f func(clave string, n int64)) { r.persistir = f }

func (r *Receptor) contar(c *atomic.Int64, clave string, n int64) {
	c.Add(n)
	if r.persistir != nil {
		r.persistir(clave, n)
	}
}

// respuestaOTLP es el cuerpo que la spec OTLP espera. `partialSuccess` es cómo se le dice al
// emisor «recibí, pero tiré N» sin mentirle con un 200 liso ni castigarlo con un 5xx.
type respuestaOTLP struct {
	PartialSuccess *parcial `json:"partialSuccess,omitempty"`
}

type parcial struct {
	RejectedLogRecords int64  `json:"rejectedLogRecords,omitempty"`
	RejectedDataPoints int64  `json:"rejectedDataPoints,omitempty"`
	ErrorMessage       string `json:"errorMessage,omitempty"`
}

// ServeHTTP implementa el contrato de §4.1, punto por punto.
func (r *Receptor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// `recover()` es la RED DE SEGURIDAD, no la estrategia: el decodificador no debe entrar
	// por acá y el fuzz existe para eso. Pero un pánico no puede tumbar el daemon entero.
	defer func() {
		if p := recover(); p != nil {
			slog.Error("telemetria: pánico en el receptor — el daemon sigue vivo", "panic", p)
			http.Error(w, "internal", http.StatusInternalServerError)
		}
	}()

	// 1 · método y ruta.
	if req.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "solo POST", http.StatusMethodNotAllowed)
		return
	}
	esLogs := strings.HasSuffix(req.URL.Path, "/v1/logs")
	esMetricas := strings.HasSuffix(req.URL.Path, "/v1/metrics")
	if !esLogs && !esMetricas {
		http.Error(w, "ruta desconocida", http.StatusNotFound)
		return
	}

	// 2 · Content-Type. Un `application/x-protobuf` responde **415 nombrando la causa Y el
	//     arreglo** — jamás un 200 que finge haber guardado, que es la forma de perder
	//     telemetría sin que nadie se entere.
	ct := req.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		r.contar(&r.salud.RechazadosFormato, "rechazados_formato", 1)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		_, _ = w.Write([]byte("este receptor habla OTLP/JSON; poné OTEL_EXPORTER_OTLP_PROTOCOL=http/json\n" +
			"(recibido Content-Type: " + ct + ")\n"))
		return
	}

	// 3 · tope de cuerpo: se corta ANTES de leerlo entero a memoria.
	req.Body = http.MaxBytesReader(w, req.Body, r.opts.MaxBody)
	cuerpo, err := leerTodo(req)
	if err != nil {
		r.contar(&r.salud.RechazadosTamano, "rechazados_tamano", 1)
		http.Error(w, "cuerpo demasiado grande", http.StatusRequestEntityTooLarge)
		return
	}

	// 4 · decodificación. Un error de forma es 400 y **el lote entero se rechaza**.
	var evs []domain.EventoTelemetria
	var descartes Descartes
	var noSoportados int64
	ahora := r.opts.Reloj().UTC()
	if esLogs {
		regs, derr := DecodificarLogs(cuerpo)
		if derr != nil {
			r.contar(&r.salud.RechazadosFormato, "rechazados_formato", 1)
			// Se loguean los primeros 200 bytes **sin el cuerpo**: el cuerpo puede traer
			// texto de la conversación.
			slog.Warn("telemetria: lote inválido rechazado entero", "bytes", len(cuerpo), "err", derr)
			http.Error(w, domain.ErrPayloadInvalido.Error(), http.StatusBadRequest)
			return
		}
		for _, reg := range regs {
			ev, d, merr := MapearLogRecord(reg, r.opts.Perfil, ahora)
			descartes.FueraDeLista += d.FueraDeLista
			descartes.Identidad += d.Identidad
			if merr != nil {
				continue // ignorado a propósito (default-deny de eventos) o sin sesión.
			}
			evs = append(evs, ev)
		}
	} else {
		puntos, derr := DecodificarMetricas(cuerpo)
		if derr != nil {
			r.contar(&r.salud.RechazadosFormato, "rechazados_formato", 1)
			http.Error(w, domain.ErrPayloadInvalido.Error(), http.StatusBadRequest)
			return
		}
		for _, p := range puntos {
			if !p.TemporalidadSoportada {
				noSoportados++
			}
			ev, d, merr := MapearPuntoMetrica(p, r.opts.Perfil, ahora)
			descartes.FueraDeLista += d.FueraDeLista
			descartes.Identidad += d.Identidad
			if merr != nil {
				continue
			}
			evs = append(evs, ev)
		}
	}
	r.contar(&r.salud.Recibidos, "recibidos", int64(len(evs)))
	if descartes.FueraDeLista > 0 {
		r.contar(&r.salud.FueraDeLista, "atributos_fuera_de_lista", int64(descartes.FueraDeLista))
	}
	if noSoportados > 0 {
		r.contar(&r.salud.Temporalidad, "temporalidad_no_soportada", noSoportados)
	}

	// 5 · encolado NO bloqueante. La cola llena descarta, lo cuenta, y **lo dice**.
	rechazados := int64(0)
	if len(evs) > 0 {
		aceptados, ierr := r.sink.Ingerir(req.Context(), evs)
		if ierr != nil || aceptados < len(evs) {
			rechazados = int64(len(evs) - aceptados)
			if rechazados < 0 {
				rechazados = 0
			}
			r.contar(&r.salud.ColaLlena, "descartados_cola_llena", rechazados)
		}
	}

	// 6 · respuesta. Siempre 200 con `partialSuccess` cuando hubo descartes: el emisor se
	//     entera y no reintenta, que es lo que un 5xx provocaría.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := respuestaOTLP{}
	if rechazados > 0 {
		p := &parcial{ErrorMessage: "cola llena"}
		if esLogs {
			p.RejectedLogRecords = rechazados
		} else {
			p.RejectedDataPoints = rechazados
		}
		resp.PartialSuccess = p
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// leerTodo lee el cuerpo acotado. Se separa para que el tope se aplique en un solo lugar.
func leerTodo(req *http.Request) ([]byte, error) {
	defer func() { _ = req.Body.Close() }()
	buf := make([]byte, 0, 64<<10)
	tmp := make([]byte, 32<<10)
	for {
		n, err := req.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				return buf, nil
			}
			// `http.MaxBytesReader` devuelve su propio error al pasarse: se propaga.
			if strings.Contains(err.Error(), "too large") {
				return nil, err
			}
			return buf, nil
		}
	}
}

// sinkNulo descarta todo. Existe para que el composition root pueda montar el receptor aunque
// el almacén no esté disponible: se pierde telemetría, no la sesión del usuario.
type sinkNulo struct{}

func (sinkNulo) Ingerir(context.Context, []domain.EventoTelemetria) (int, error) {
	return 0, ports.ErrAlmacenNoDisponible
}

// SinkNulo devuelve un sumidero que no guarda nada y lo dice. Nunca se usa para fingir éxito.
func SinkNulo() ports.TelemetriaSink { return sinkNulo{} }
