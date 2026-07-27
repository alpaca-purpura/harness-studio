package usecase_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// storeQueSerializa modela el costo REAL del disco: serializa el registro entero, que es
// lo que hace el store de verdad en cada turno. Un fake que sólo guarda el slice mediría
// el costo de un `append` y diría que todo es rápido.
type storeQueSerializa struct {
	sesiones []domain.Session
	bytes    int
}

func (s *storeQueSerializa) Load(context.Context) ([]domain.Session, error) {
	return append([]domain.Session{}, s.sesiones...), nil
}

func (s *storeQueSerializa) Save(_ context.Context, ss []domain.Session) error {
	b, err := json.Marshal(ss)
	if err != nil {
		return err
	}
	s.bytes = len(b)
	s.sesiones = append([]domain.Session{}, ss...)
	return nil
}

// conversacionDeReferencia arma una conversación del tamaño de la más larga que existe hoy
// en disco: 90 turnos, ~12 KB. La medición se hace contra el peso real, no contra uno
// cómodo.
func conversacionDeReferencia(id string, activa bool) domain.Conversacion {
	turnos := make([]domain.Turn, 0, 90)
	texto := strings.Repeat("el operador escribió algo de largo realista sobre el arnés. ", 2)
	for i := range 90 {
		rol := domain.RolAssistant
		if i%2 == 0 {
			rol = domain.RolUser
		}
		turnos = append(turnos, domain.Turn{Rol: rol, Text: texto})
	}
	return domain.Conversacion{
		ID: id, Titulo: "hilo de referencia", Activa: activa,
		CreadaEn: "2026-07-26T00:00:00Z", ClaudeSessionID: "cc-" + id,
		CtxPct: 42, Conv: turnos,
	}
}

// BenchmarkPersistLocked20Conversaciones (H-7) mide lo que cuesta persistir el registro con
// 20 conversaciones del tamaño de la más larga real (~240 KB), porque el registro se
// reescribe ENTERO en cada turno y ahora incluye N transcripts.
//
// Presupuesto: < 15 ms. Si lo supera, el corte es partir el registro en un archivo por
// sesión — carpeta que ya existe— y eso NO es un rediseño: es cambiar el path del store.
// Si no se mide, no se puede afirmar que alcanza, y eso sería el pass fabricado.
func BenchmarkPersistLocked20Conversaciones(b *testing.B) {
	st := &storeQueSerializa{}
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, st, stubPub{}, stubResolver{path: b.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		b.Fatal(err)
	}
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "el frente cargado"})
	if err != nil {
		b.Fatal(err)
	}

	// 20 conversaciones cargadas: una activa + 19 inactivas que siguen pesando en el JSON.
	cargada, _ := svc.Get(s.ID)
	cargada.Conversaciones = make([]domain.Conversacion, 0, 20)
	for i := range 20 {
		cargada.Conversaciones = append(cargada.Conversaciones,
			conversacionDeReferencia(domain.NuevoConvID(), i == 0))
	}
	if serr := st.Save(context.Background(), []domain.Session{cargada}); serr != nil {
		b.Fatal(serr)
	}
	svc2, err := usecase.NewSessionService(context.Background(), &stubAgent{}, st, stubPub{}, stubResolver{path: b.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		b.Fatal(err)
	}
	id := svc2.List()[0].ID

	b.ReportMetric(float64(st.bytes), "bytes_del_registro")
	b.ResetTimer()
	for range b.N {
		// SetView es la mutación más barata que persiste el registro entero: mide el costo
		// de persistir, no el de la operación.
		if _, verr := svc2.SetView(id, "Mapa"); verr != nil {
			b.Fatal(verr)
		}
	}
}
