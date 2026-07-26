package forward

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// dialerContado cuenta las conexiones salientes. Es el instrumento del control positivo:
// «no salió nada» no significa nada si el dialer nunca se hubiera usado igual.
type dialerContado struct{ n atomic.Int64 }

func (d *dialerContado) cliente() *http.Client {
	base := &net.Dialer{Timeout: 2 * time.Second}
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, red, dir string) (net.Conn, error) {
				d.n.Add(1)
				return base.DialContext(ctx, red, dir)
			},
		},
	}
}

func evento(sesion string) domain.EventoTelemetria {
	return domain.EventoTelemetria{
		LlaveJoin:        domain.LlaveJoin{SesionID: sesion, TurnoID: "t-1", ArnesID: "vitalia"},
		Emisor:           domain.EmisorOTLP,
		Runtime:          "claude-code",
		AdaptadorVersion: "cc-otlp/1",
		TSRecibido:       time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC),
		TipoEvento:       domain.EventoAPIRequest,
		Escenario:        domain.EscenarioS1,
		Atribucion:       domain.ConfianzaExacta,
	}
}

// TestForwardApagadoPorDefault — sin flag, **cero conexiones salientes**. Con flag,
// exactamente una.
//
// ⚠ Control positivo (§4): el assert no es «no salió nada» — es «con el flag apagado 0, con
// el flag encendido 1», sobre el mismo dialer contado y un servidor de puerto EFÍMERO.
func TestForwardApagadoPorDefault(t *testing.T) {
	var recibidos atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recibidos.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	d := &dialerContado{}
	apagado := New("", d.cliente())
	if apagado.Activo() {
		t.Fatal("sin destino configurado, el forward está APAGADO — es el default")
	}
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		if err := apagado.Enviar(ctx, []domain.EventoTelemetria{evento("s-apagado")}); err != nil {
			t.Fatal(err)
		}
	}
	if n := d.n.Load(); n != 0 {
		t.Fatalf("apagado tiene que abrir 0 conexiones salientes, abrió %d", n)
	}

	// ── control positivo: encendido, sale exactamente 1 ──
	d2 := &dialerContado{}
	encendido := New(srv.URL, d2.cliente())
	if !encendido.Activo() {
		t.Fatal("con destino configurado, el forward está encendido")
	}
	if err := encendido.Enviar(ctx, []domain.EventoTelemetria{evento("s-encendido")}); err != nil {
		t.Fatal(err)
	}
	if n := d2.n.Load(); n != 1 {
		t.Fatalf("encendido tiene que abrir exactamente 1 conexión, abrió %d", n)
	}
	if recibidos.Load() != 1 {
		t.Fatalf("el destino tiene que haber recibido 1 lote, recibió %d", recibidos.Load())
	}
}

// TestForwardNoReenviaCrudo — lo que sale es el evento PROYECTADO. El email y los
// identificadores de cuenta que el runtime manda en cada punto no pueden aparecer.
//
// ⚠ Control positivo: `sesion_id` **sí** está en lo enviado. Sin eso, un forward que mandara
// un cuerpo vacío pasaría el assert de ausencia.
func TestForwardNoReenviaCrudo(t *testing.T) {
	const marcaPII = "PII-EMAIL-forward@ejemplo.test"
	var cuerpo atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		cuerpo.Store(string(b))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	f := New(srv.URL, nil)
	// El evento canónico ni siquiera TIENE dónde poner la PII: eso ya lo garantiza el
	// dominio. Lo que este test prueba es que el forward no la reintroduce por otra vía —
	// no serializa un payload crudo guardado al costado, por ejemplo.
	ev := evento("SESION-CONTROL-POSITIVO")
	ev.Motivo = "un motivo cualquiera"
	if err := f.Enviar(context.Background(), []domain.EventoTelemetria{ev}); err != nil {
		t.Fatal(err)
	}
	enviado, _ := cuerpo.Load().(string)
	if enviado == "" {
		t.Fatal("el destino no recibió nada: el test no comparó nada")
	}
	if strings.Contains(enviado, marcaPII) {
		t.Errorf("el forward reenvió identidad: %s", enviado)
	}
	for _, prohibido := range []string{"user.email", "account_uuid", "organization.id", "prompt"} {
		if strings.Contains(enviado, prohibido) {
			t.Errorf("el forward reenvió el campo prohibido %q: %s", prohibido, enviado)
		}
	}
	// ── control positivo ──
	if !strings.Contains(enviado, "SESION-CONTROL-POSITIVO") {
		t.Fatalf("control positivo: el sesion_id tiene que viajar; el assert de ausencia no prueba nada sin esto: %s", enviado)
	}
	// Y es JSON del evento canónico, no un blob opaco.
	var vuelta []domain.EventoTelemetria
	if err := json.Unmarshal([]byte(enviado), &vuelta); err != nil {
		t.Fatalf("lo enviado tiene que ser el evento canónico: %v", err)
	}
	if len(vuelta) != 1 || vuelta[0].SesionID != "SESION-CONTROL-POSITIVO" {
		t.Errorf("el evento no viajó entero: %+v", vuelta)
	}
}

// TestForwardSoloPorFlagDelOperador — el destino se fija al construir, desde la config del
// daemon. **No hay ninguna vía para cambiarlo en caliente**: ni un setter, ni un campo
// exportado, ni un endpoint. Un arnés no puede alcanzarlo.
func TestForwardSoloPorFlagDelOperador(t *testing.T) {
	f := New("", nil)
	if f.Activo() {
		t.Fatal("el default es apagado")
	}
	// La única forma de encenderlo es construir otro con destino — o sea, reiniciar el
	// daemon con el flag. Eso es exactamente el punto.
	if f.Destino() != "" {
		t.Errorf("un forward apagado no tiene destino: %q", f.Destino())
	}
	// Control positivo: construido CON destino, está encendido y lo dice.
	g := New("http://ejemplo.invalido/otlp", nil)
	if !g.Activo() || g.Destino() != "http://ejemplo.invalido/otlp" {
		t.Errorf("el destino del operador tiene que quedar visible: activo=%v destino=%q", g.Activo(), g.Destino())
	}
}

// TestForwardQueFallaNoRompeLaTelemetriaLocal — un destino caído devuelve error para que el
// caller lo loguee, pero el error no puede propagarse hasta tumbar la ingesta local.
func TestForwardQueFallaNoRompeLaTelemetriaLocal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	f := New(srv.URL, nil)
	if err := f.Enviar(context.Background(), []domain.EventoTelemetria{evento("s-1")}); err == nil {
		t.Fatal("un destino que responde error tiene que devolver error, no fingir éxito")
	}
}
