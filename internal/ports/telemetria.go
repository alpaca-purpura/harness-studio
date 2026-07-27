package ports

import (
	"context"
	"errors"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// telemetria.go declara los puertos del módulo de telemetría (arquitectura-modulo.md §2.5).
// Ninguna de las seis interfaces menciona OTLP, SQLite ni HTTP: los casos de uso no saben
// qué hay del otro lado (boundary dominio-independiente-de-transporte + .go-arch-lint.yml).

// TelemetriaSink es la ÚNICA puerta de escritura del módulo. Los cuatro emisores
// (receptor OTLP, adaptador de agente, hook, daemon) escriben por acá y por ningún otro lado.
//
// Ingerir NUNCA bloquea al llamador más allá de encolar: un receptor que hace esperar al
// agente degradaría el trabajo que mide (boundary telemetria-de-nacimiento, fail-open).
//
// Devuelve cuántos aceptó. **La diferencia con len(evs) son descartes CONTADOS, no
// silencio**: un sink que devolviera solo `error` no podría distinguir «guardé los 10» de
// «guardé 7 y tiré 3», y esa diferencia es exactamente lo que la cobertura muestra.
type TelemetriaSink interface {
	Ingerir(ctx context.Context, evs []domain.EventoTelemetria) (aceptados int, err error)
}

// ConsultaTelemetria acota una lectura. Desde/Hasta en cero = «toda la historia retenida».
// Limite en cero = el default del adaptador; jamás «sin límite».
type ConsultaTelemetria struct {
	ArnesID       string
	InstalacionID string
	CajaID        string
	SesionID      string
	Desde, Hasta  time.Time
	Limite        int
}

// PurgaTelemetria parametriza un borrado. `AntesDe` aplica la retención por TTL; `ArnesID`
// borra lo de un arnés (D15.3, el botón de la UI).
//
// **Desde/Hasta acotan el borrado por arnés a una ventana** (D26.5 · A-4). Antes no existían y
// el `DELETE` era siempre total, mientras la confirmación de la UI declaraba el conteo **de la
// ventana activa**: con la ventana en 7 días sobre dos años de historial, la confirmación
// subdeclaraba la destrucción. Es la única acción irreversible de la superficie, así que su
// alcance declarado tiene que ser su alcance real.
//
// Las dos en cero = todo el historial del arnés, que sigue siendo un caso legítimo y explícito.
type PurgaTelemetria struct {
	AntesDe time.Time
	ArnesID string
	Desde   time.Time
	Hasta   time.Time
}

// TelemetriaStore es el almacén. Extiende el sink porque el mismo adaptador escribe y lee,
// pero los casos de uso de ingesta piden SOLO el sink (interfaz mínima en el consumidor).
type TelemetriaStore interface {
	TelemetriaSink

	// Resumen agrega sobre el rollup horario. Devuelve punteros nil —no ceros— cuando no
	// hubo dato que agregar (boundary no-aplica-no-es-cero).
	Resumen(ctx context.Context, q ConsultaTelemetria) (domain.ResumenTelemetria, error)
	// PorCaja devuelve el gasto por caja, INCLUIDAS las cajas sin dato atribuible: la
	// ausencia viaja explícita con motivo, nunca omitida.
	PorCaja(ctx context.Context, q ConsultaTelemetria) ([]domain.GastoCaja, error)
	// Turnos hace el join dinero×proceso por (SesionID, TurnoID) — la llave verificada
	// (ANEXO H1). Es el drill-down: toca la tabla cruda, no el rollup.
	Turnos(ctx context.Context, q ConsultaTelemetria) ([]domain.TurnoUnido, error)
	// EsperarTurno registra que un turno OCURRIÓ, lo sepamos medir o no. Es el denominador
	// de la cobertura (§6.4); sin él, «cuánto medimos» se leería como «cuánto hubo».
	EsperarTurno(ctx context.Context, sesionID, turnoID, arnesID, cajaID string) error
	// Purgar aplica retención o borra un arnés; devuelve cuántas filas se fueron.
	// Recomputa el rollup de las horas afectadas dentro de la MISMA transacción: si no,
	// quedaría una cifra agregada huérfana alimentándose de filas que ya no existen.
	Purgar(ctx context.Context, p PurgaTelemetria) (int64, error)
	// Descartar / Recuperar / Descartados son las DECISIONES del operador sobre los puntos de
	// mejora (D26.4). Viven en el almacén y no en memoria porque un descarte que se pierde al
	// reiniciar no es un descarte: el punto vuelve solo y el operador vuelve a descartarlo.
	Descartar(ctx context.Context, arnesID, puntoID string, ahora time.Time) error
	Recuperar(ctx context.Context, arnesID, puntoID string) error
	Descartados(ctx context.Context, arnesID string) (map[string]bool, error)
	// Salud son los contadores del receptor + el estado del almacén. Persistidos, no en
	// memoria: «cuántos descarté» es dato de honestidad y tiene que sobrevivir al reinicio.
	Salud(ctx context.Context) (domain.SaludTelemetria, error)
	// Close libera los handles. A diferencia del índice, este .db NO es desechable:
	// cerrar bien importa.
	Close() error
}

// AtribucionRegistry es la tabla aprendida `plugin_id_hash → arnés` (INFORME §V3). Se
// aprende del spawn controlado: nosotros instalamos el arnés, así que podemos observar qué
// hash le corresponde. NO se asume determinismo entre máquinas (V7.2, sin verificar).
type AtribucionRegistry interface {
	PorHash(hash string) (arnesID, instalacionID string, ok bool)
	// PorCWD resuelve el cwd de un hook a una instalación conocida del Portafolio
	// (identidad (home,id)). La ruta se usa y se descarta: nunca se persiste cruda (A14).
	PorCWD(cwd string) (arnesID, instalacionID string, ok bool)
	// Aprender registra un par observado. `como` documenta de dónde salió, para poder
	// explicar después una atribución rara en vez de tener que adivinarla.
	Aprender(hash, arnesID, instalacionID, como string) error
}

// CatalogoPrecios cotiza un modelo. Un modelo desconocido devuelve ok=false — jamás un
// PrecioModelo en cero, que costearía todo gratis en silencio.
type CatalogoPrecios interface {
	Precio(modeloCanonico string) (domain.PrecioModelo, bool)
	// Canonizar aplica los alias de cloud (Bedrock/Vertex) al nombre crudo del runtime.
	// Un nombre que no matchea se devuelve TAL CUAL: no se inventa una canonización.
	Canonizar(modelo string) string
	Version() domain.VersionCatalogoPrecios
}

// DescubrimientoDaemon publica dónde está el daemon para que el hook lo encuentre sin que
// nadie le pase nada (D6.5). Publicar se llama DESPUÉS de que el listener acepta: una ficha
// que nombra un puerto muerto es una mentira que el hook cobra en timeouts.
type DescubrimientoDaemon interface {
	Publicar(ctx context.Context, f domain.FichaDaemon) error
	Leer() (domain.FichaDaemon, error)
	Retirar() error
}

// ForwardOTLP es el escape hatch del OPERADOR (C2/D13), apagado por default. Recibe el
// evento YA PROYECTADO — jamás el cuerpo OTLP crudo, que llevaría el email de quien corra
// el arnés (INFORME §V6). Un arnés no puede alcanzar ni configurar esto.
type ForwardOTLP interface {
	Enviar(ctx context.Context, evs []domain.EventoTelemetria) error
	Activo() bool
	Destino() string
}

var (
	// ErrColaLlena — el sink descartó por saturación. El caller lo CUENTA y sigue; jamás
	// bloquea ni reintenta: la instrumentación no degrada al sistema instrumentado.
	ErrColaLlena = errors.New("telemetria: cola llena")
	// ErrAlmacenNoDisponible — el .db no se pudo abrir/escribir (disco lleno, corrupción).
	// El daemon SIGUE: se pierde telemetría, no la sesión del usuario.
	ErrAlmacenNoDisponible = errors.New("telemetria: almacén no disponible")
	// ErrSinFicha — no hay ficha de daemon publicada. Es la señal de fail-open del hook.
	ErrSinFicha = errors.New("telemetria: no hay daemon publicado")
)
