package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Una Conversacion es UN hilo de trabajo dentro de una sesión (CV-D3). La sesión es el
// frente de trabajo — el arnés, la empresa, el puesto, el cwd —; la conversación es el
// diálogo con el conductor: su id de Claude Code, su modelo, su uso de contexto y su
// registro liviano de turnos.
//
// La relación es 1—N con N ≥ 1 y exactamente una marcada Activa. Esa invariante vive acá,
// en operaciones puras sobre el agregado: no en el handler, no en el store, no en el FE.

// Los sentinelas del agregado.
var (
	// ErrConvNoEncontrada — el id pedido no pertenece a ESTA sesión. Jamás se busca
	// globalmente (BR-CV-2): una conversación cuelga de su sesión, no de un string suelto.
	ErrConvNoEncontrada = errors.New("domain: la conversación no pertenece a esta sesión")
	// ErrTituloVacio — el título recortado quedó vacío; el título anterior NO se toca
	// (RF-344 CA-1).
	ErrTituloVacio = errors.New("domain: el título no puede quedar vacío")
	// ErrSinActiva — la sesión no tiene conversación activa. Es la violación de la
	// invariante que VerificarUnaActiva reporta y que NormalizarConversaciones repara.
	ErrSinActiva = errors.New("domain: la sesión quedó sin conversación activa")
)

// TituloConversacionNueva es el título con el que nace toda conversación (RF-301 CA-3).
// NO hereda el "nuevo frente" de la sesión: son dos nombres distintos y ambos existen.
const TituloConversacionNueva = "nueva conversación"

// prefijoConvID distingue de un vistazo un id de conversación de uno de sesión ("s…")
// cuando ambos aparecen en la misma línea de log.
const prefijoConvID = "cv"

// Conversacion es un hilo dentro de una sesión.
type Conversacion struct {
	// ID es el id estable de la conversación ("cv" + 8 hex). Único dentro de su sesión.
	ID string `json:"id"`

	// Titulo se auto-deriva del primer mensaje del usuario y es editable (CV-D9).
	// La sesión conserva su propio Frente: son dos nombres y ambos siguen existiendo.
	Titulo string `json:"titulo"`

	// TituloEditado marca que el operador lo escribió a mano. Con la marca puesta,
	// ningún turno vuelve a derivar el título (RF-303 CA-1).
	TituloEditado bool `json:"titulo_editado,omitempty"`

	// Activa marca el hilo que tiene conductor. Exactamente una por sesión.
	// SIN omitempty a propósito: `false` es un dato, no una ausencia (no-aplica-no-es-cero).
	Activa bool `json:"activa"`

	// CreadaEn (RFC3339, UTC) es el desempate de orden para las de 0 turnos, que no
	// tienen fecha de interacción y a las que está prohibido inventarles una.
	CreadaEn string `json:"creada_en"`

	// UltimaInteraccion (RFC3339, UTC) se estampa en cada turno (RF-304). Vacío ⟺ 0 turnos:
	// el vacío ya significa «no hubo turno» y la UI lo pinta con su propio literal.
	UltimaInteraccion string `json:"ultima_interaccion,omitempty"`

	// ClaudeSessionID es el id de Claude Code de esta conversación: lo que va a --resume.
	ClaudeSessionID string `json:"claude_session_id,omitempty"`
	// Model es el modelo capturado del init de CADA proceso de esta conversación.
	Model string `json:"model,omitempty"`

	// CtxPct es el último uso conocido de la ventana de contexto (0–100).
	CtxPct int `json:"ctx_pct,omitempty"`
	// CtxHist es el histórico de CtxPct por turno; insumo del umbral de rotación.
	CtxHist []int `json:"ctx_hist,omitempty"`
	// RotacionPendiente: el uso cruzó el umbral y el próximo turno spawnea proceso fresco.
	RotacionPendiente bool `json:"rotacion_pendiente,omitempty"`

	// CadenaCC son los ClaudeSessionID previos de ESTA conversación (sus rotaciones,
	// CV-D10): la rotación es invisible, es la misma conversación.
	CadenaCC []string `json:"cadena_cc,omitempty"`
	// Checkpoint es el digest mecánico de la última rotación; viaja al proceso fresco.
	Checkpoint string `json:"checkpoint,omitempty"`

	// Conv es el registro liviano de turnos que la UI repinta al retomar.
	// SIN omitempty, deliberado: con omitempty, nil («no cargado») y [] («leí y está
	// vacío») se serializan igual, que es justo lo que no-aplica-no-es-cero prohíbe.
	Conv []Turn `json:"conv"`
}

// largoDeTitulo es a cuántos caracteres se recorta un nombre derivado de un mensaje.
const largoDeTitulo = 48

// RecorteDeTitulo normaliza un mensaje del usuario a un nombre corto: colapsa los blancos
// y recorta con puntos suspensivos. Devuelve "" cuando no queda nada — QUIÉN pone el
// default es del que llama, y no es el mismo: la sesión cae en "nuevo frente" y la
// conversación en "nueva conversación".
//
// Vive en el dominio porque la usan dos: el usecase al derivar el frente de una sesión y
// el migrador del registro al bautizar las conversaciones que nacen de un archivo viejo.
// El migrador no puede importar el usecase (el grafo de dependencias lo prohíbe), y tener
// dos copias de la regla sería tener dos reglas.
func RecorteDeTitulo(text string) string {
	text = strings.TrimSpace(strings.Join(strings.Fields(text), " "))
	if len(text) > largoDeTitulo {
		return text[:largoDeTitulo] + "…"
	}
	return text
}

// NumTurnos deriva la cuenta de turnos. NO existe un campo `Turnos` persistido: duplicar
// len(Conv) es un drift esperando ocurrir, y esto lo hace imposible de mentir.
func (c Conversacion) NumTurnos() int { return len(c.Conv) }

// NuevoConvID devuelve un id corto y resistente a colisiones para una conversación.
// Mismo generador que el de sesión, con otro prefijo.
func NuevoConvID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read no falla en las plataformas soportadas; degradá determinista.
		return prefijoConvID + "00000000"
	}
	return prefijoConvID + hex.EncodeToString(b[:])
}

// selloUTC formatea un instante como la RFC3339 en UTC que el registro persiste.
func selloUTC(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// Activa devuelve la conversación activa de la sesión. El bool es el guardrail de un
// registro editado a mano: bajo la invariante siempre es true.
//
// ⚠ El puntero apunta al backing array de s.Conversaciones: no lo retengas a través de
// una transición que agregue conversaciones (CrearConversacion puede reubicar el slice).
func (s *Session) Activa() (*Conversacion, bool) {
	for i := range s.Conversaciones {
		if s.Conversaciones[i].Activa {
			return &s.Conversaciones[i], true
		}
	}
	return nil, false
}

// Instantanea devuelve una copia de la sesión segura para leer FUERA del candado que
// protege el registro vivo. Copia el slice de conversaciones: sin esto, el valor que
// devuelve un Get comparte el backing array con el runtime, y cada campo que el conductor
// mueve (RotacionPendiente, CtxPct, ClaudeSessionID, Titulo) es una escritura sobre la
// memoria que el lector ya tiene en la mano. Con la sesión plana el problema no existía —
// esos campos se copiaban por valor; al bajar a un slice, dejaron de copiarse.
//
// NO copia los slices DENTRO de cada conversación (Conv, CtxHist, CadenaCC) y es a
// propósito: sólo se les appendea, así que el lector ve un prefijo estable y jamás toca la
// posición que el escritor escribe. Copiarlos sería copiar el transcript entero en cada
// List() — el costo que este método existe para no pagar.
func (s Session) Instantanea() Session {
	s.Conversaciones = append([]Conversacion(nil), s.Conversaciones...)
	return s
}

// BuscarConversacion devuelve la conversación cid de ESTA sesión. Nunca busca fuera.
func (s *Session) BuscarConversacion(cid string) (*Conversacion, bool) {
	for i := range s.Conversaciones {
		if s.Conversaciones[i].ID == cid {
			return &s.Conversaciones[i], true
		}
	}
	return nil, false
}

// CrearConversacion desactiva la activa y agrega una nueva activa. Devuelve el id de la
// que quedó inactiva ("" si la sesión no tenía ninguna). Es UNA transición (BR-CV-4): no
// existe un estado intermedio con cero activas ni con dos.
//
// La desactivada conserva TODO — su Conv, su Checkpoint, su CadenaCC (RF-305). Desactivar
// no es archivar.
func (s *Session) CrearConversacion(id string, ahora time.Time) (nueva *Conversacion, desactivada string) {
	if id == "" {
		id = NuevoConvID()
	}
	for i := range s.Conversaciones {
		if s.Conversaciones[i].Activa {
			s.Conversaciones[i].Activa = false
			desactivada = s.Conversaciones[i].ID
		}
	}
	s.Conversaciones = append(s.Conversaciones, Conversacion{
		ID:       id,
		Titulo:   TituloConversacionNueva,
		Activa:   true,
		CreadaEn: selloUTC(ahora),
		Conv:     []Turn{},
	})
	return &s.Conversaciones[len(s.Conversaciones)-1], desactivada
}

// ActivarConversacion mueve la marca de activa a cid. Es no-op si cid ya es la activa
// (E-15): devuelve desactivada == "" y err == nil, sin tocar nada. Devuelve
// ErrConvNoEncontrada si cid no está en ESTA sesión (BR-CV-2).
func (s *Session) ActivarConversacion(cid string) (desactivada string, err error) {
	destino := -1
	for i := range s.Conversaciones {
		if s.Conversaciones[i].ID == cid {
			destino = i
			break
		}
	}
	if destino < 0 {
		return "", fmt.Errorf("%w: %q en la sesión %q", ErrConvNoEncontrada, cid, s.ID)
	}
	if s.Conversaciones[destino].Activa {
		return "", nil
	}
	for i := range s.Conversaciones {
		if s.Conversaciones[i].Activa {
			s.Conversaciones[i].Activa = false
			desactivada = s.Conversaciones[i].ID
		}
	}
	s.Conversaciones[destino].Activa = true
	return desactivada, nil
}

// RenombrarConversacion aplica el título recortado y marca TituloEditado. Con el título
// vacío devuelve ErrTituloVacio y el título NO cambia (RF-344 CA-1 / E-30).
func (s *Session) RenombrarConversacion(cid, titulo string) error {
	c, ok := s.BuscarConversacion(cid)
	if !ok {
		return fmt.Errorf("%w: %q en la sesión %q", ErrConvNoEncontrada, cid, s.ID)
	}
	limpio := strings.TrimSpace(titulo)
	if limpio == "" {
		return ErrTituloVacio
	}
	c.Titulo = limpio
	c.TituloEditado = true
	return nil
}

// DerivarTitulo aplica un título derivado automáticamente del primer turno. Es no-op si
// el operador ya editó el título (RF-303 CA-1) o si el derivado viene vacío. Devuelve si
// el título cambió.
//
// El derivador vive en el usecase (RF-303 CA-2 manda reusar el que ya existe): acá vive
// sólo la ley de quién gana cuando ambos escriben, y gana el operador.
func (c *Conversacion) DerivarTitulo(titulo string) bool {
	if c.TituloEditado {
		return false
	}
	limpio := strings.TrimSpace(titulo)
	if limpio == "" || limpio == c.Titulo {
		return false
	}
	c.Titulo = limpio
	return true
}

// VerificarUnaActiva es el predicado puro de la invariante: al menos una conversación y
// exactamente una activa. Lo corre el test de dominio después de CADA transición, el
// store después de cargar, y persistLocked antes de escribir.
func VerificarUnaActiva(s Session) error {
	if len(s.Conversaciones) == 0 {
		return fmt.Errorf("%w: la sesión %q no tiene conversaciones", ErrSinActiva, s.ID)
	}
	activas := 0
	for i := range s.Conversaciones {
		if s.Conversaciones[i].Activa {
			activas++
		}
	}
	switch {
	case activas == 0:
		return fmt.Errorf("%w: la sesión %q tiene %d conversaciones y ninguna activa", ErrSinActiva, s.ID, len(s.Conversaciones))
	case activas > 1:
		return fmt.Errorf("domain: la sesión %q tiene %d conversaciones activas y sólo puede tener una", s.ID, activas)
	default:
		return nil
	}
}

// NormalizarConversaciones REPARA la invariante y DEVUELVE qué reparó, una línea por
// arreglo, para que el arranque lo diga en el log. Nunca repara en silencio (E-01).
//
//   - 0 conversaciones ⇒ crea una activa vacía.
//   - ninguna activa   ⇒ activa la de UltimaInteraccion más reciente
//     (desempate: CreadaEn más reciente).
//   - ≥2 activas       ⇒ conserva esa misma y desactiva el resto.
//
// Devuelve nil cuando no hubo nada que reparar: un arranque sin novedades no loguea.
func (s *Session) NormalizarConversaciones(ahora time.Time) []string {
	if len(s.Conversaciones) == 0 {
		nueva, _ := s.CrearConversacion(NuevoConvID(), ahora)
		return []string{fmt.Sprintf(
			"sesión %q: no tenía conversaciones — se creó una activa vacía (%s)", s.ID, nueva.ID)}
	}

	var activas []int
	for i := range s.Conversaciones {
		if s.Conversaciones[i].Activa {
			activas = append(activas, i)
		}
	}

	switch {
	case len(activas) == 1:
		return nil

	case len(activas) == 0:
		todas := make([]int, len(s.Conversaciones))
		for i := range todas {
			todas[i] = i
		}
		gana := s.masReciente(todas)
		s.Conversaciones[gana].Activa = true
		return []string{fmt.Sprintf(
			"sesión %q: %d conversaciones y ninguna activa — se activó %q (%s)",
			s.ID, len(s.Conversaciones), s.Conversaciones[gana].Titulo, s.Conversaciones[gana].ID)}

	default:
		gana := s.masReciente(activas)
		desactivadas := 0
		for _, i := range activas {
			if i == gana {
				continue
			}
			s.Conversaciones[i].Activa = false
			desactivadas++
		}
		return []string{fmt.Sprintf(
			"sesión %q: había %d conversaciones activas — se conservó %q (%s) y se desactivaron las otras %d",
			s.ID, len(activas), s.Conversaciones[gana].Titulo, s.Conversaciones[gana].ID, desactivadas)}
	}
}

// masReciente elige, entre los índices dados, el de UltimaInteraccion más reciente; con
// empate (o sin interacción) gana el de CreadaEn más reciente; con empate también ahí,
// el último del registro. Los sellos son RFC3339 UTC de ancho fijo: comparar como texto
// es comparar como instante, y el vacío («sin turnos») queda por debajo de cualquier fecha.
func (s *Session) masReciente(idxs []int) int {
	gana := idxs[0]
	for _, i := range idxs[1:] {
		a, b := s.Conversaciones[i], s.Conversaciones[gana]
		switch {
		case a.UltimaInteraccion != b.UltimaInteraccion:
			if a.UltimaInteraccion > b.UltimaInteraccion {
				gana = i
			}
		case a.CreadaEn != b.CreadaEn:
			if a.CreadaEn > b.CreadaEn {
				gana = i
			}
		default:
			gana = i
		}
	}
	return gana
}
