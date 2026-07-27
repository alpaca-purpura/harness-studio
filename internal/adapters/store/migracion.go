package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// migrador lleva el payload de sesiones de la versión N a la N+1. Falla ruidoso: un campo
// que no se supo llevar es un error, jamás un zero-value en silencio. `ahora` entra como
// argumento y no se lee del reloj adentro, para que la migración sea determinista y
// testeable — el único campo que no puede salir del archivo viejo es la fecha de creación.
type migrador func(payload json.RawMessage, ahora time.Time) (json.RawMessage, error)

// migradores es la cadena. Falta un eslabón ⇒ AbrirRegistro falla nombrando la versión que
// no supo migrar, en vez de leer a medias.
var migradores = map[int]migrador{1: deV1aV2}

// Informe es lo que el arranque LOGUEA. Nada de lo que pasa acá ocurre en silencio: si el
// registro se migró, se respaldó, se reparó o se puso en cuarentena, sale por el log con
// la ruta del archivo involucrado.
type Informe struct {
	Migro        bool     // corrió una migración en este arranque.
	DesdeVersion int      // 0 si no migró.
	RespaldoEn   string   // "" si no migró.
	Reparaciones []string // lo que NormalizarConversaciones arregló, una línea por arreglo.

	// Modos de fallo C y E (cuarentena y esquema futuro). Los puebla T10.
	Corrupto      bool
	CuarentenaEn  string
	EsquemaFuturo bool
}

// Hubo reporta si el informe tiene algo que decir. Un arranque sin novedades no loguea.
func (i Informe) Hubo() bool {
	return i.Migro || i.Corrupto || i.EsquemaFuturo || len(i.Reparaciones) > 0
}

// sesionV1 es la forma VIEJA del registro, congelada acá a propósito. Un migrador que
// decodifica con el tipo del dominio de hoy deja de ser un migrador: cuando el dominio
// cambie otra vez, leería el archivo viejo con la forma nueva y perdería lo que ya no
// existe. Esta struct no se toca nunca más.
type sesionV1 struct {
	ID                string        `json:"id"`
	Frente            string        `json:"frente"`
	Arnes             string        `json:"arnes"`
	Empresa           string        `json:"empresa"`
	Puesto            string        `json:"puesto"`
	Salud             domain.Salud  `json:"salud"`
	Status            string        `json:"status"`
	View              string        `json:"view"`
	Parked            string        `json:"parked"`
	Reparacion        bool          `json:"reparacion"`
	Cwd               string        `json:"cwd"`
	CerradaEn         string        `json:"cerrada_en"`
	ClaudeSessionID   string        `json:"claude_session_id"`
	Model             string        `json:"model"`
	CtxPct            int           `json:"ctx_pct"`
	CtxHist           []int         `json:"ctx_hist"`
	RotacionPendiente bool          `json:"rotacion_pendiente"`
	CadenaCC          []string      `json:"cadena_cc"`
	Checkpoint        string        `json:"checkpoint"`
	Conv              []domain.Turn `json:"conv"`
}

// deV1aV2 convierte cada sesión plana en una sesión con UNA conversación activa que se
// lleva todo el estado del diálogo. Es PURA: no hace IO, no conoce el Portafolio y no
// toca `Arnes` — la recalibración de llaves es un paso aparte, por eso mismo.
//
// Lo que hereda la conversación: el id de Claude Code, el modelo, el uso de contexto y su
// histórico, la marca de rotación, la cadena de rotaciones, el checkpoint y los turnos.
// Lo que NO se inventa: `ultima_interaccion` sale VACÍA aunque la conversación tenga 90
// turnos, porque esa fecha no existe en el archivo viejo y fabricarla sería mentir sobre
// cuándo se habló por última vez.
func deV1aV2(payload json.RawMessage, ahora time.Time) (json.RawMessage, error) {
	var viejas []sesionV1
	if err := json.Unmarshal(payload, &viejas); err != nil {
		return nil, fmt.Errorf("store: migración 1→2: no se pudo leer el registro viejo: %w", err)
	}
	nuevas := make([]domain.Session, 0, len(viejas))
	for _, v := range viejas {
		turnos := v.Conv
		if turnos == nil {
			// [] y null se serializan igual con omitempty, y no significan lo mismo: acá
			// se fija el vacío explícito para que «leí y no había turnos» sea legible.
			turnos = []domain.Turn{}
		}
		nuevas = append(nuevas, domain.Session{
			ID:         v.ID,
			Frente:     v.Frente,
			Arnes:      v.Arnes,
			Empresa:    v.Empresa,
			Puesto:     v.Puesto,
			Salud:      v.Salud,
			Status:     domain.SessionStatus(v.Status),
			View:       v.View,
			Parked:     v.Parked,
			Reparacion: v.Reparacion,
			Cwd:        v.Cwd,
			CerradaEn:  v.CerradaEn,
			Conversaciones: []domain.Conversacion{{
				ID:                domain.NuevoConvID(),
				Titulo:            tituloDeLaMigrada(turnos),
				Activa:            true,
				CreadaEn:          ahora.UTC().Format(time.RFC3339),
				UltimaInteraccion: "", // no existe en v1 — no se inventa (BR-CV-14).
				ClaudeSessionID:   v.ClaudeSessionID,
				Model:             v.Model,
				CtxPct:            v.CtxPct,
				CtxHist:           v.CtxHist,
				RotacionPendiente: v.RotacionPendiente,
				CadenaCC:          v.CadenaCC,
				Checkpoint:        v.Checkpoint,
				Conv:              turnos,
			}},
		})
	}
	out, err := json.Marshal(nuevas)
	if err != nil {
		return nil, fmt.Errorf("store: migración 1→2: no se pudo escribir el registro nuevo: %w", err)
	}
	return out, nil
}

// tituloDeLaMigrada bautiza la conversación con el primer mensaje del usuario, la misma
// regla con la que el daemon vivo nombra un frente. Sin ningún mensaje del usuario, el
// nombre por defecto de una conversación — que NO es el de una sesión.
func tituloDeLaMigrada(turnos []domain.Turn) string {
	for _, t := range turnos {
		if t.Rol != domain.RolUser {
			continue
		}
		if titulo := domain.RecorteDeTitulo(t.Text); titulo != "" {
			return titulo
		}
	}
	return domain.TituloConversacionNueva
}

// AbrirRegistro resuelve qué archivo manda, migra si hace falta y deja el Registry listo.
// Es lo único que el composition root llama; NewRegistry sigue siendo el constructor tonto
// y ports.SessionStore no cambia de firma, así que ningún caller se entera.
//
// rutaV2 es el registro nuevo; rutaLegado el viejo, que NO se toca, NO se borra y NO se
// renombra: mientras exista intacto, volver a un binario anterior es gratis.
func AbrirRegistro(rutaV2, rutaLegado, sello string) (*Registry, Informe, error) {
	var inf Informe
	reg := &Registry{path: rutaV2, sello: sello}

	origen, crudo, err := leerElQueMande(rutaV2, rutaLegado)
	if err != nil {
		return nil, inf, err
	}
	if crudo == nil {
		return reg, inf, nil // primer arranque: no hay nada que migrar.
	}

	version, err := detectarVersion(crudo)
	if err != nil {
		return nil, inf, fmt.Errorf("store: %s: %w", origen, err)
	}
	if version > EsquemaActual {
		return nil, inf, fmt.Errorf(
			"store: %s lo escribió un binario más nuevo (esquema %d, este binario entiende %d): "+
				"no se lee a medias ni se pisa", origen, version, EsquemaActual)
	}

	payload, err := payloadDe(crudo, version)
	if err != nil {
		return nil, inf, fmt.Errorf("store: %s: %w", origen, err)
	}

	ahora := time.Now().UTC()
	if version < EsquemaActual {
		respaldo, berr := respaldar(origen, version, sello)
		if berr != nil {
			return nil, inf, berr
		}
		inf.Migro, inf.DesdeVersion, inf.RespaldoEn = true, version, respaldo
		for v := version; v < EsquemaActual; v++ {
			paso, ok := migradores[v]
			if !ok {
				return nil, inf, fmt.Errorf("store: falta el migrador de la versión %d a la %d", v, v+1)
			}
			if payload, err = paso(payload, ahora); err != nil {
				return nil, inf, err
			}
		}
	}

	var sesiones []domain.Session
	if err := json.Unmarshal(payload, &sesiones); err != nil {
		return nil, inf, fmt.Errorf("store: %s: no se pudo leer el registro migrado: %w", origen, err)
	}
	for i := range sesiones {
		inf.Reparaciones = append(inf.Reparaciones, sesiones[i].NormalizarConversaciones(ahora)...)
	}

	// Sólo se escribe si algo cambió. Un arranque sin novedades deja el archivo igual byte
	// a byte: la idempotencia se ve en el disco, no se promete en un comentario.
	if inf.Migro || len(inf.Reparaciones) > 0 {
		if err := reg.Save(context.Background(), sesiones); err != nil {
			return nil, inf, err
		}
	}
	return reg, inf, nil
}

// leerElQueMande devuelve el contenido del registro que manda y de dónde salió. El v2
// gana siempre: es el camino de todos los arranques después del primero. Un crudo nil
// significa que no hay ningún registro todavía.
func leerElQueMande(rutaV2, rutaLegado string) (origen string, crudo []byte, err error) {
	b, err := os.ReadFile(rutaV2) //nolint:gosec // ruta del propio daemon, resuelta en el composition root.
	switch {
	case err == nil:
		return rutaV2, b, nil
	case !os.IsNotExist(err):
		return rutaV2, nil, fmt.Errorf("store: read %s: %w", rutaV2, err)
	}
	if rutaLegado == "" {
		return rutaV2, nil, nil
	}
	b, err = os.ReadFile(rutaLegado) //nolint:gosec // ídem.
	switch {
	case err == nil:
		return rutaLegado, b, nil
	case os.IsNotExist(err):
		return rutaV2, nil, nil
	default:
		return rutaLegado, nil, fmt.Errorf("store: read %s: %w", rutaLegado, err)
	}
}

// payloadDe extrae el array de sesiones: en v1 el archivo ENTERO lo es; de v2 en adelante
// sale del campo `sesiones` del sobre.
func payloadDe(crudo []byte, version int) (json.RawMessage, error) {
	if version == 1 {
		return crudo, nil
	}
	var s sobre
	if err := json.Unmarshal(crudo, &s); err != nil {
		return nil, fmt.Errorf("no se pudo abrir el sobre: %w", err)
	}
	if len(s.Sesiones) == 0 {
		return json.RawMessage("[]"), nil
	}
	return s.Sesiones, nil
}

// respaldar copia el archivo que se va a migrar, ANTES de tocar nada. El nombre lleva el
// sello de build para que dos migraciones no se pisen el respaldo.
func respaldar(ruta string, version int, sello string) (string, error) {
	if sello == "" {
		sello = time.Now().UTC().Format("0601021504")
	}
	destino := fmt.Sprintf("%s.v%d-%s.bak", ruta, version, sello)
	b, err := os.ReadFile(ruta) //nolint:gosec // la misma ruta que se acaba de leer.
	if err != nil {
		return "", fmt.Errorf("store: respaldo de %s: %w", ruta, err)
	}
	// El destino es la ruta que se acaba de leer + un sufijo propio: no viene de afuera.
	if err := os.WriteFile(destino, b, 0o600); err != nil { //nolint:gosec // ruta derivada de la del registro.
		return "", fmt.Errorf("store: respaldo en %s: %w", destino, err)
	}
	return destino, nil
}
