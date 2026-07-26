package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// sttFake is a scriptable TranscriptionPort.
type sttFake struct {
	disp   ports.Disponibilidad
	texto  string
	err    error
	mime   string // captured
	audio  []byte // captured
	llamdo int
}

func (f *sttFake) Disponible(context.Context) ports.Disponibilidad { return f.disp }

func (f *sttFake) Transcribir(_ context.Context, audio []byte, mime string) (string, error) {
	f.llamdo++
	f.audio, f.mime = audio, mime
	return f.texto, f.err
}

// limpiezaFake is a scriptable LimpiezaPort that records the context it was handed.
type limpiezaFake struct {
	salida   string
	err      error
	crudo    string
	contexto string
	demora   time.Duration
}

func (f *limpiezaFake) Ordenar(ctx context.Context, crudo, contexto string) (string, error) {
	f.crudo, f.contexto = crudo, contexto
	if f.demora > 0 {
		select {
		case <-time.After(f.demora):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return f.salida, f.err
}

// sesionesFake is a one-session SessionLookup.
type sesionesFake struct{ sess domain.Session }

func (f sesionesFake) Get(id string) (domain.Session, bool) {
	if id != f.sess.ID {
		return domain.Session{}, false
	}
	return f.sess, true
}

func sesion(conv ...domain.Turn) sesionesFake {
	return sesionesFake{sess: domain.Session{ID: "s1", Arnes: "vitalia", Conv: conv}}
}

func disponible() ports.Disponibilidad {
	return ports.Disponibilidad{Disponible: true, Motor: "whisper-cli"}
}

// TestDictarDevuelveLimpioCuandoLaLimpiezaAnda — el camino feliz: transcribir + ordenar.
func TestDictarDevuelveLimpioCuandoLaLimpiezaAnda(t *testing.T) {
	stt := &sttFake{disp: disponible(), texto: "este eh el pip de la tarjeta esa"}
	lim := &limpiezaFake{salida: "Cambiá el color del pip de la PermissionCard."}
	svc := usecase.NewDictadoService(stt, lim, sesion())

	got, err := svc.Dictar(context.Background(), "s1", []byte("audio"), "audio/mp4")
	if err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	if got.Estado != usecase.DictadoLimpio {
		t.Errorf("estado = %q, quiero %q", got.Estado, usecase.DictadoLimpio)
	}
	if got.Texto != "Cambiá el color del pip de la PermissionCard." {
		t.Errorf("texto = %q", got.Texto)
	}
	if got.Motivo != "" {
		t.Errorf("un dictado limpio no lleva motivo, tiene %q", got.Motivo)
	}
	if got.Motor != "whisper-cli" {
		t.Errorf("motor = %q, quiero whisper-cli", got.Motor)
	}
	// Lo que se le pasó a la limpieza es el CRUDO, no el audio ni nada re-derivado.
	if lim.crudo != "este eh el pip de la tarjeta esa" {
		t.Errorf("la limpieza recibió %q", lim.crudo)
	}
	if stt.mime != "audio/mp4" {
		t.Errorf("el mime no llegó al motor: %q", stt.mime)
	}
}

// TestDictarCaeACrudoSiLaLimpiezaFalla — V-D4: el dictado NO se pierde por una falla del
// paso opcional; vuelve crudo y MARCADO.
func TestDictarCaeACrudoSiLaLimpiezaFalla(t *testing.T) {
	stt := &sttFake{disp: disponible(), texto: "el pip de la tarjeta esa ponele warning"}
	lim := &limpiezaFake{err: errors.New("claude: exit status 1")}
	svc := usecase.NewDictadoService(stt, lim, sesion())

	got, err := svc.Dictar(context.Background(), "s1", []byte("audio"), "audio/mp4")
	if err != nil {
		t.Fatalf("una falla de limpieza NO es un error del dictado: %v", err)
	}
	if got.Estado != usecase.DictadoCrudo {
		t.Errorf("estado = %q, quiero %q", got.Estado, usecase.DictadoCrudo)
	}
	if got.Texto != "el pip de la tarjeta esa ponele warning" {
		t.Errorf("se perdió el crudo: %q", got.Texto)
	}
	if got.Motivo == "" {
		t.Error("un crudo SIN motivo es un gap escondido: el FE no puede explicar nada")
	}
}

// TestDictarMarcaElCorteDeTiempoDistintoDeLaFalla — «tardó» y «falló» se arreglan distinto,
// así que el motivo tiene que distinguirlos (con lo primero, reintentar sirve).
func TestDictarMarcaElCorteDeTiempoDistintoDeLaFalla(t *testing.T) {
	stt := &sttFake{disp: disponible(), texto: "algo dictado"}
	lim := &limpiezaFake{demora: time.Hour}
	svc := usecase.NewDictadoService(stt, lim, sesion())
	svc.SetTopeLimpieza(20 * time.Millisecond)

	got, err := svc.Dictar(context.Background(), "s1", []byte("audio"), "audio/mp4")
	if err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	if got.Estado != usecase.DictadoCrudo {
		t.Fatalf("estado = %q, quiero crudo", got.Estado)
	}
	if !strings.Contains(got.Motivo, "tardó") {
		t.Errorf("motivo = %q, quiero que diga que se cortó por tiempo", got.Motivo)
	}
}

// TestDictarNoInventaTextoSiLaTranscripcionVuelveVacia — silencio ≠ falla, pero tampoco es
// texto: vuelve error para que el FE avise y NO toque el composer (RF-227).
func TestDictarNoInventaTextoSiLaTranscripcionVuelveVacia(t *testing.T) {
	stt := &sttFake{disp: disponible(), texto: "   \n  "}
	svc := usecase.NewDictadoService(stt, &limpiezaFake{salida: "no deberia llamarse"}, sesion())

	got, err := svc.Dictar(context.Background(), "s1", []byte("audio"), "audio/mp4")
	if !errors.Is(err, usecase.ErrDictadoVacio) {
		t.Fatalf("err = %v, quiero ErrDictadoVacio", err)
	}
	if got.Texto != "" {
		t.Errorf("no se devuelve texto cuando no se entendió nada, tiene %q", got.Texto)
	}
}

// TestDictarRechazaSesionDesconocida — sin sesión no hay contexto, y sin contexto la
// limpieza es justamente lo que el spike probó que NO sirve.
func TestDictarRechazaSesionDesconocida(t *testing.T) {
	svc := usecase.NewDictadoService(&sttFake{disp: disponible(), texto: "x"}, &limpiezaFake{}, sesion())
	if _, err := svc.Dictar(context.Background(), "no-existe", []byte("a"), "audio/mp4"); !errors.Is(err, usecase.ErrSesionDesconocida) {
		t.Fatalf("err = %v, quiero ErrSesionDesconocida", err)
	}
}

// TestDictarNoAceptaUnDictadoQueNoPodraTranscribir — RF-223: mejor decir que no ahora que
// fallar después de que el operador habló tres minutos.
func TestDictarNoAceptaUnDictadoQueNoPodraTranscribir(t *testing.T) {
	stt := &sttFake{disp: ports.Disponibilidad{Motivo: "no hay motor de transcripción instalado"}}
	svc := usecase.NewDictadoService(stt, &limpiezaFake{}, sesion())

	_, err := svc.Dictar(context.Background(), "s1", []byte("audio"), "audio/mp4")
	if err == nil {
		t.Fatal("sin motor, Dictar tiene que fallar")
	}
	if !strings.Contains(err.Error(), "no hay motor") {
		t.Errorf("el error no dice el motivo: %v", err)
	}
	if stt.llamdo != 0 {
		t.Error("no se llama al motor si ya sabíamos que no está")
	}
}

// TestContextoDeLimpiezaRecortaAUltimosTurnos — V-D2: 2-3 turnos, no la conversación entera.
func TestContextoDeLimpiezaRecortaAUltimosTurnos(t *testing.T) {
	conv := []domain.Turn{
		{Rol: domain.RolUser, Text: "turno-viejisimo"},
		{Rol: domain.RolAssistant, Text: "respuesta-vieja"},
		{Rol: domain.RolUser, Text: "turno-1"},
		{Rol: domain.RolAssistant, Text: "turno-2"},
		{Rol: domain.RolUser, Text: "turno-3"},
	}
	lim := &limpiezaFake{salida: "ok"}
	svc := usecase.NewDictadoService(&sttFake{disp: disponible(), texto: "algo"}, lim, sesion(conv...))

	if _, err := svc.Dictar(context.Background(), "s1", []byte("a"), "audio/mp4"); err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	for _, quiero := range []string{"turno-1", "turno-2", "turno-3"} {
		if !strings.Contains(lim.contexto, quiero) {
			t.Errorf("falta %q en el contexto:\n%s", quiero, lim.contexto)
		}
	}
	if strings.Contains(lim.contexto, "turno-viejisimo") {
		t.Error("se coló un turno más viejo que los últimos 3 — V-D2 dice contexto CHICO")
	}
	// El orden importa: el bloque se lee cronológico, no al revés.
	if strings.Index(lim.contexto, "turno-1") > strings.Index(lim.contexto, "turno-3") {
		t.Error("los turnos salieron en orden invertido")
	}
}

// TestContextoDeLimpiezaLlevaElGlosario — el glosario es lo que repara los errores de STT.
func TestContextoDeLimpiezaLlevaElGlosario(t *testing.T) {
	lim := &limpiezaFake{salida: "ok"}
	svc := usecase.NewDictadoService(&sttFake{disp: disponible(), texto: "algo"}, lim, sesion())

	if _, err := svc.Dictar(context.Background(), "s1", []byte("a"), "audio/mp4"); err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	// `daemon` es literalmente uno de los términos que el STT erró en la medición (§1.8).
	if !strings.Contains(lim.contexto, "daemon") {
		t.Errorf("el glosario no llegó a la limpieza:\n%s", lim.contexto)
	}
}

// TestContextoDeLimpiezaIgnoraLaActividad — los pasos `act` («Read foo.go») son migas de
// herramienta, no conversación: meten ruido en un bloque que vale por ser chico.
func TestContextoDeLimpiezaIgnoraLaActividad(t *testing.T) {
	conv := []domain.Turn{
		{Rol: domain.RolUser, Text: "el pedido de verdad"},
		{Rol: domain.RolAct, Text: "Read internal/domain/box.go"},
		{Rol: domain.RolAct, Text: "Grep permisos"},
		{Rol: domain.RolSys, Text: "skill_activated"},
	}
	lim := &limpiezaFake{salida: "ok"}
	svc := usecase.NewDictadoService(&sttFake{disp: disponible(), texto: "algo"}, lim, sesion(conv...))

	if _, err := svc.Dictar(context.Background(), "s1", []byte("a"), "audio/mp4"); err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	if strings.Contains(lim.contexto, "Read internal/domain/box.go") {
		t.Error("la actividad se coló al contexto")
	}
	if !strings.Contains(lim.contexto, "el pedido de verdad") {
		t.Error("se perdió el turno de conversación que sí importa")
	}
}

// TestContextoDeLimpiezaRecortaTurnosGigantes — un turno enorme diluye el glosario.
func TestContextoDeLimpiezaRecortaTurnosGigantes(t *testing.T) {
	enorme := strings.Repeat("x", 5000)
	lim := &limpiezaFake{salida: "ok"}
	svc := usecase.NewDictadoService(
		&sttFake{disp: disponible(), texto: "algo"}, lim,
		sesion(domain.Turn{Rol: domain.RolAssistant, Text: enorme}),
	)

	if _, err := svc.Dictar(context.Background(), "s1", []byte("a"), "audio/mp4"); err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	if strings.Contains(lim.contexto, enorme) {
		t.Errorf("el turno gigante entró entero (%d chars de contexto)", len(lim.contexto))
	}
	if !strings.Contains(lim.contexto, "…") {
		t.Error("el recorte no quedó marcado")
	}
}

// TestDisponibilidadPasaElMotivoDelMotor — el motivo viaja tal cual hasta el FE.
func TestDisponibilidadReportaMotivoSinMotor(t *testing.T) {
	stt := &sttFake{disp: ports.Disponibilidad{
		Motivo:   "no hay motor de transcripción instalado",
		Instalar: []string{"whisper-cli", "faster-whisper"},
	}}
	svc := usecase.NewDictadoService(stt, &limpiezaFake{}, sesion())

	got := svc.Disponibilidad(context.Background())
	if got.Disponible {
		t.Fatal("sin motor no está disponible")
	}
	if got.Motivo == "" || len(got.Instalar) == 0 {
		t.Errorf("la degradación tiene que decir motivo Y qué instalar: %+v", got)
	}
}

// TestSinLimpiadorDevuelveCrudoHonesto — degradación, no falla.
func TestSinLimpiadorDevuelveCrudoHonesto(t *testing.T) {
	svc := usecase.NewDictadoService(&sttFake{disp: disponible(), texto: "lo dictado"}, nil, sesion())

	got, err := svc.Dictar(context.Background(), "s1", []byte("a"), "audio/mp4")
	if err != nil {
		t.Fatalf("Dictar: %v", err)
	}
	if got.Estado != usecase.DictadoCrudo || got.Motivo == "" {
		t.Errorf("quiero crudo con motivo, tengo %+v", got)
	}
}
