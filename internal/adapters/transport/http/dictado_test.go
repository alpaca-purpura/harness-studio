package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// sttWireFake implementa ports.TranscriptionPort para probar el WIRE: qué códigos salen y
// qué campos viajan, no la lógica del motor (esa vive en su propio paquete).
type sttWireFake struct {
	disp  ports.Disponibilidad
	texto string
	err   error
	mime  string
}

func (f *sttWireFake) Disponible(context.Context) ports.Disponibilidad { return f.disp }
func (f *sttWireFake) Transcribir(_ context.Context, _ []byte, mime string) (string, error) {
	f.mime = mime
	return f.texto, f.err
}

// limpiezaWireFake implementa ports.LimpiezaPort.
type limpiezaWireFake struct {
	salida string
	err    error
}

func (f *limpiezaWireFake) Ordenar(context.Context, string, string) (string, error) {
	return f.salida, f.err
}

// sesionesWireFake implementa usecase.SessionLookup con una sola sesión conocida.
type sesionesWireFake struct{}

func (sesionesWireFake) Get(id string) (domain.Session, bool) {
	if id != "s1" {
		return domain.Session{}, false
	}
	return domain.Session{ID: "s1", Arnes: "vitalia"}, true
}

// postDictadoReq arma el request tal como lo manda el FE: el audio crudo en el body.
func postDictadoReq(body []byte, mime string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/sessions/s1/dictado", bytes.NewReader(body))
	r.SetPathValue("id", "s1")
	if mime != "" {
		r.Header.Set("Content-Type", mime)
	}
	return r
}

func svcCon(stt ports.TranscriptionPort, lim ports.LimpiezaPort) *usecase.DictadoService {
	return usecase.NewDictadoService(stt, lim, sesionesWireFake{})
}

func disponibleWire() ports.Disponibilidad {
	return ports.Disponibilidad{Disponible: true, Motor: "whisper-cli"}
}

// TestPostDictadoDevuelveLimpioOCrudo — el contrato central de RF-222: la respuesta SIEMPRE
// dice en qué estado quedó el texto. Sin ese campo el FE tendría que adivinar, y pasar un
// crudo por limpio es exactamente el pass fabricado que la doctrina prohíbe.
func TestPostDictadoDevuelveLimpioOCrudo(t *testing.T) {
	casos := []struct {
		nombre   string
		lim      ports.LimpiezaPort
		estado   string
		conMotiv bool
	}{
		{"la limpieza anduvo", &limpiezaWireFake{salida: "Cambiá el pip a warn."}, "limpio", false},
		{"la limpieza falló", &limpiezaWireFake{err: errors.New("boom")}, "crudo", true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			stt := &sttWireFake{disp: disponibleWire(), texto: "eh el pip ese ponele warning"}
			w := httptest.NewRecorder()
			postDictado(svcCon(stt, c.lim))(w, postDictadoReq([]byte("audio"), "audio/mp4"))

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, quiero 200: %s", w.Code, w.Body)
			}
			var got usecase.Dictado
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("respuesta no-JSON: %v", err)
			}
			if string(got.Estado) != c.estado {
				t.Errorf("estado = %q, quiero %q", got.Estado, c.estado)
			}
			if got.Texto == "" {
				t.Error("el composer necesita texto: vino vacío")
			}
			if c.conMotiv && got.Motivo == "" {
				t.Error("un crudo sin motivo no le sirve al FE para explicar nada")
			}
		})
	}
}

// TestPostDictadoPasaElMimeAlMotor — el motor necesita saber en qué formato viene el audio;
// perder el Content-Type en el transporte rompería la transcodificación. Se prueba con un
// mime CON parámetros porque es el caso que un cliente real manda.
func TestPostDictadoPasaElMimeAlMotor(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), texto: "algo"}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, postDictadoReq([]byte("a"), "audio/mp4;codecs=mp4a.40.2"))

	if stt.mime != "audio/mp4;codecs=mp4a.40.2" {
		t.Errorf("el motor recibió mime %q", stt.mime)
	}
}

// TestPostDictadoAsumeWavSinContentType — RF-229: el FE arma el WAV él mismo porque el
// `MediaRecorder` de WebKitGTK entrega 0 bytes, así que `audio/wav` es el formato real que
// sube la app y el default honesto para un cliente que no lo declara.
func TestPostDictadoAsumeWavSinContentType(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), texto: "algo"}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, postDictadoReq([]byte("a"), ""))

	if stt.mime != "audio/wav" {
		t.Errorf("mime por defecto = %q, quiero audio/wav", stt.mime)
	}
}

// TestPostDictadoRechazaAudioDemasiadoGrande — la cota funcional (3 min) vive en el FE y el
// endpoint no puede confiar en que se respetó.
func TestPostDictadoRechazaAudioDemasiadoGrande(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), texto: "algo"}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, postDictadoReq(bytes.Repeat([]byte("a"), maxAudio+1), "audio/mp4"))

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, quiero 413", w.Code)
	}
}

// TestPostDictadoRechazaCuerpoVacio — grabar 0 bytes no es un dictado.
func TestPostDictadoRechazaCuerpoVacio(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), texto: "algo"}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, postDictadoReq(nil, "audio/mp4"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, quiero 400", w.Code)
	}
}

// TestPostDictado404SiLaSesionNoExiste — sin sesión no hay contexto para la limpieza.
func TestPostDictado404SiLaSesionNoExiste(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), texto: "algo"}
	r := httptest.NewRequest(http.MethodPost, "/api/sessions/nope/dictado", strings.NewReader("audio"))
	r.SetPathValue("id", "nope")
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, quiero 404", w.Code)
	}
}

// TestPostDictadoNoSeEntendioNadaNoEs500 — el pipeline anduvo; el audio no tenía voz. Un 500
// haría que el FE muestre «error del servidor» por un silencio.
func TestPostDictadoNoSeEntendioNadaNoEs500(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), texto: "   "}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, postDictadoReq([]byte("audio"), "audio/mp4"))

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, quiero 422", w.Code)
	}
}

// TestPostDictadoRechazaMimeNoSoportado — el error del motor sube como 500 con su motivo;
// lo que NO puede pasar es que el FE reciba un 200 con texto inventado.
func TestPostDictadoRechazaMimeNoSoportado(t *testing.T) {
	stt := &sttWireFake{disp: disponibleWire(), err: errors.New("stt: formato de audio no soportado: \"audio/flac\"")}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{salida: "x"}))(w, postDictadoReq([]byte("audio"), "audio/flac"))

	if w.Code == http.StatusOK {
		t.Fatal("un mime que el motor no come no puede devolver 200")
	}
	if !strings.Contains(w.Body.String(), "no soportado") {
		t.Errorf("el error no dice el motivo: %s", w.Body)
	}
}

// TestGetDisponibilidadDiceElMotivo — RF-227: el FE pinta el motivo en la superficie, así
// que tiene que viajar por el wire, no quedarse en un log del daemon.
func TestGetDisponibilidadDiceElMotivo(t *testing.T) {
	stt := &sttWireFake{disp: ports.Disponibilidad{
		Motivo:   "no hay motor de transcripción instalado",
		Instalar: []string{"whisper-cli", "faster-whisper"},
	}}
	w := httptest.NewRecorder()
	getDisponibilidad(svcCon(stt, &limpiezaWireFake{}))(w, httptest.NewRequest(http.MethodGet, "/api/dictado/disponibilidad", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, quiero 200", w.Code)
	}
	var got ports.Disponibilidad
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("respuesta no-JSON: %v", err)
	}
	if got.Disponible {
		t.Error("sin motor no está disponible")
	}
	if got.Motivo == "" || len(got.Instalar) == 0 {
		t.Errorf("el wire perdió el motivo o el qué-instalar: %+v", got)
	}
}

// TestPostDictadoSinMotorEs503NoEs500 — «no puedo transcribir» es una capacidad ausente, no
// una falla interna. Un 500 mandaría al operador a buscar un bug donde falta un `apt install`.
func TestPostDictadoSinMotorEs503NoEs500(t *testing.T) {
	stt := &sttWireFake{disp: ports.Disponibilidad{
		Motivo:   "no hay motor de transcripción instalado",
		Instalar: []string{"whisper-cli"},
	}}
	w := httptest.NewRecorder()
	postDictado(svcCon(stt, &limpiezaWireFake{}))(w, postDictadoReq([]byte("audio"), "audio/mp4"))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, quiero 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), "no hay motor") {
		t.Errorf("el 503 no dice el motivo: %s", w.Body)
	}
}
