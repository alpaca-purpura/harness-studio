package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	telcatalogo "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

func i64(v int64) *int64 { return &v }

// handlerConTelemetria arma el router con un servicio real sobre un almacén TEMPORAL.
// **Jamás el HOME real del operador.**
func handlerConTelemetria(t *testing.T, sembrar func(*usecase.TelemetriaService, *telstore.Store)) http.Handler {
	t.Helper()
	st, err := telstore.New(filepath.Join(t.TempDir(), "telemetria.db"),
		telstore.Opciones{LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(),
		domain.PerfilRuntime{Runtime: "claude-code", Aritmetica: domain.AritmeticaDisjunta},
		time.Now)
	svc.SetRetencion(st, telstore.NewRollup(st, time.Hour), 90, 24)
	if sembrar != nil {
		sembrar(svc, st)
	}
	vacio := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	return NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		svc, nil, vacio, vacio, AuthConfigFor("127.0.0.1:4200", ""))
}

func get(t *testing.T, h http.Handler, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:4200"+ruta, nil)
	req.Host = "127.0.0.1:4200"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

// TestNoAplicaNoEsCeroEnElWireHTTP — un bucket que el runtime no reportó **no puede aparecer
// en 0** en el JSON que el FE consume. Es el contrato transversal de toda la superficie.
func TestNoAplicaNoEsCeroEnElWireHTTP(t *testing.T) {
	h := handlerConTelemetria(t, func(svc *usecase.TelemetriaService, st *telstore.Store) {
		ctx := context.Background()
		if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{{
			LlaveJoin:  domain.LlaveJoin{SesionID: "s-1", TurnoID: "t-1", ArnesID: "vitalia", CajaID: "paso-3"},
			Emisor:     domain.EmisorOTLP,
			Runtime:    "claude-code",
			TSRecibido: time.Now().UTC(),
			TipoEvento: domain.EventoAPIRequest,
			Escenario:  domain.EscenarioS2Instrumentado,
			Modelo:     "claude-haiku-4-5",
			// Razonamiento y los tramos de cache write quedan nil: el runtime no los da.
			Tokens:               domain.Tokens{Entrada: i64(10), Salida: i64(39)},
			CostoReportadoMicros: i64(18473),
		}}); err != nil {
			t.Fatal(err)
		}
		if err := st.Sincronizar(ctx); err != nil {
			t.Fatal(err)
		}
	})

	w := get(t, h, "/api/telemetria/arneses/vitalia/cajas/paso-3")
	if w.Code != http.StatusOK {
		t.Fatalf("código %d: %s", w.Code, w.Body.String())
	}
	cuerpo := w.Body.String()
	if strings.Contains(cuerpo, `"razonamiento"`) {
		t.Errorf("un bucket que nadie reportó NO viaja: %s", cuerpo)
	}
	if strings.Contains(cuerpo, `"razonamiento":0`) {
		t.Errorf("y muchísimo menos en 0: %s", cuerpo)
	}
	// ── control positivo: lo que SÍ se midió viaja ──
	if !strings.Contains(cuerpo, `"entrada":10`) {
		t.Fatalf("control positivo: el bucket medido tiene que viajar: %s", cuerpo)
	}
	if !strings.Contains(cuerpo, `"reportado_micros":18473`) {
		t.Fatalf("control positivo: el costo reportado tiene que viajar: %s", cuerpo)
	}
}

// TestCajasSinDatoNoSeOmiten — una caja sin dato viaja con `atribuible:false`, motivo no vacío
// y costo `null`. Omitirla obligaría al FE a inventar por qué falta; ponerle 0 diría que
// corrió gratis.
func TestCajasSinDatoNoSeOmiten(t *testing.T) {
	h := handlerConTelemetria(t, func(svc *usecase.TelemetriaService, st *telstore.Store) {
		ctx := context.Background()
		base := func(sesion, caja string, atribuido bool, micros *int64) domain.EventoTelemetria {
			e := domain.EventoTelemetria{
				LlaveJoin:            domain.LlaveJoin{SesionID: sesion, TurnoID: "t-1", CajaID: caja},
				Emisor:               domain.EmisorOTLP,
				Runtime:              "claude-code",
				TSRecibido:           time.Now().UTC(),
				TipoEvento:           domain.EventoAPIRequest,
				Escenario:            domain.EscenarioS2Instrumentado,
				CostoReportadoMicros: micros,
			}
			if atribuido {
				e.ArnesID = "vitalia"
			}
			return e
		}
		if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{
			base("s-1", "paso-3", true, i64(1000)),
			// Una caja del MISMO arnés con actividad registrada y SIN dinero: el estado
			// normal del modo degradado. La caja viaja igual, con su motivo propio.
			func() domain.EventoTelemetria {
				e := base("s-2", "paso-11", true, nil)
				e.Emisor = domain.EmisorHook
				e.TipoEvento = domain.EventoTurnoFin
				return e
			}(),
		}); err != nil {
			t.Fatal(err)
		}
		if err := st.Sincronizar(ctx); err != nil {
			t.Fatal(err)
		}
	})

	w := get(t, h, "/api/telemetria/arneses/vitalia/cajas")
	if w.Code != http.StatusOK {
		t.Fatalf("código %d: %s", w.Code, w.Body.String())
	}
	var cajas []domain.GastoCaja
	if err := json.Unmarshal(w.Body.Bytes(), &cajas); err != nil {
		t.Fatal(err)
	}
	if len(cajas) == 0 {
		t.Fatal("el desglose no puede venir vacío: el test no comparó nada")
	}
	var sinDato *domain.GastoCaja
	for i := range cajas {
		if !cajas[i].Atribuible {
			sinDato = &cajas[i]
		}
	}
	if sinDato == nil {
		t.Fatalf("las cajas sin dato NO se omiten: %+v", cajas)
	}
	if sinDato.Motivo == "" {
		t.Error("el motivo es obligatorio cuando no es atribuible")
	}
	if sinDato.CostoMicros != nil {
		t.Errorf("una caja sin dato lleva costo null, no %d", *sinDato.CostoMicros)
	}
	if sinDato.CajaID != "paso-11" {
		t.Errorf("la caja sin dinero es paso-11: %q", sinDato.CajaID)
	}
	// ── control positivo: la caja CON dato sí trae su cifra ──
	var conDato *domain.GastoCaja
	for i := range cajas {
		if cajas[i].Atribuible {
			conDato = &cajas[i]
		}
	}
	if conDato == nil || conDato.CostoMicros == nil || *conDato.CostoMicros != 1000 {
		t.Fatalf("control positivo: la caja atribuible trae su costo: %+v", conDato)
	}
	// Y en el JSON crudo: `"costo_micros":null`, jamás `0`.
	if !strings.Contains(w.Body.String(), `"costo_micros":null`) {
		t.Errorf("el null tiene que viajar EXPLÍCITO en el wire: %s", w.Body.String())
	}
}

// TestPortafolioPuestoNullSinRol — D20: sin `rol` declarado el puesto viaja **null**, no una
// cadena vacía. La UI dice «puesto sin declarar»; una cadena vacía se pintaría como un puesto
// que existe y se llama «».
func TestPortafolioPuestoNullSinRol(t *testing.T) {
	h := handlerConTelemetria(t, func(svc *usecase.TelemetriaService, st *telstore.Store) {
		svc.SetPortafolio(
			func(ctx context.Context, arnesID string) string { return "" }, // ningún rol
			func(ctx context.Context) []usecase.FilaInstalacion {
				return []usecase.FilaInstalacion{{ArnesID: "vitalia", InstalacionID: "i1", Clave: "k1"}}
			},
		)
	})
	w := get(t, h, "/api/telemetria/portafolio")
	if w.Code != http.StatusOK {
		t.Fatalf("código %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"puesto":null`) {
		t.Errorf("sin rol declarado, el puesto viaja null: %s", w.Body.String())
	}
	// Y el que nunca corrió lleva `costo_por_corrida: null`, no 0.
	if !strings.Contains(w.Body.String(), `"costo_por_corrida":null`) {
		t.Errorf("el que nunca corrió lleva null, no 0: %s", w.Body.String())
	}
}

// TestVentanaIlegibleNoSeIgnora — un rango con formato equivocado se rechaza. Ignorarlo
// devolvería un período distinto del pedido, y el usuario leería cifras de otro rango
// creyendo que son del suyo.
func TestVentanaIlegibleNoSeIgnora(t *testing.T) {
	h := handlerConTelemetria(t, nil)
	if w := get(t, h, "/api/telemetria/resumen?desde=ayer"); w.Code != http.StatusBadRequest {
		t.Errorf("un `desde` ilegible tiene que dar 400, dio %d", w.Code)
	}
	if w := get(t, h, "/api/telemetria/resumen?hasta=2026-13-45"); w.Code != http.StatusBadRequest {
		t.Errorf("un `hasta` ilegible tiene que dar 400, dio %d", w.Code)
	}
	// ── control positivo: una ventana bien formada entra ──
	if w := get(t, h, "/api/telemetria/resumen?desde=2026-07-01T00:00:00Z"); w.Code != http.StatusOK {
		t.Fatalf("control positivo: una ventana RFC3339 tiene que aceptarse: %d %s", w.Code, w.Body.String())
	}
}

// TestSaludRotulaElTTLComoPropuesto — J-6: el número no está firmado y el wire lo dice.
func TestElTTLViajaFirmadoYSinRotulo(t *testing.T) {
	h := handlerConTelemetria(t, nil)
	w := get(t, h, "/api/telemetria/salud")
	if w.Code != http.StatusOK {
		t.Fatalf("código %d", w.Code)
	}
	cuerpo := w.Body.String()
	// El número viaja, porque la UI lo lee de acá y **nunca** lo hardcodea (A-2).
	if !strings.Contains(cuerpo, `"retencion_dias":90`) {
		t.Errorf("el TTL firmado tiene que viajar en el wire: %s", cuerpo)
	}
	// Y viaja SIN rótulo: el número está firmado (D26.3). Un «(propuesto)» sobre algo decidido
	// entrena a ignorar los rótulos que sí importan — y este es el único lugar donde el FE
	// podía sacarlo, así que el candado va acá.
	if strings.Contains(cuerpo, "retencion_propuesta") {
		t.Errorf("el TTL está firmado: el wire no lo rotula como propuesto: %s", cuerpo)
	}
}
