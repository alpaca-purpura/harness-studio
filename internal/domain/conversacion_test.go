package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// Estos tests prueban las cuatro operaciones del agregado SIN un solo fake. Si alguno
// necesitara uno, la invariante habría quedado en el lugar equivocado.

var (
	t0 = time.Date(2026, 7, 26, 10, 0, 0, 0, time.UTC)
	t1 = time.Date(2026, 7, 26, 11, 0, 0, 0, time.UTC)
	t2 = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
)

// sesionCon arma una sesión con las conversaciones dadas, tal cual (sin normalizar):
// los tests de reparación necesitan poder construir estados inválidos a propósito.
func sesionCon(convs ...Conversacion) *Session {
	return &Session{ID: "s0badc0de", Frente: "frente de prueba", Conversaciones: convs}
}

func conv(id, titulo string, activa bool, creada, ultima time.Time, turnos int) Conversacion {
	c := Conversacion{
		ID:       id,
		Titulo:   titulo,
		Activa:   activa,
		CreadaEn: creada.UTC().Format(time.RFC3339),
		Conv:     make([]Turn, 0, turnos),
	}
	if !ultima.IsZero() {
		c.UltimaInteraccion = ultima.UTC().Format(time.RFC3339)
	}
	for range turnos {
		c.Conv = append(c.Conv, Turn{Rol: RolUser, Text: "turno"})
	}
	return c
}

// TestInvarianteUnaActivaTrasCadaTransicion — E-01, E-15, E-20.
// La invariante se verifica DESPUÉS de cada una de las cuatro operaciones, sobre estados
// de partida distintos. Es el test que hace que ninguna transición pueda dejar el
// agregado en un estado que no se puede expresar.
func TestInvarianteUnaActivaTrasCadaTransicion(t *testing.T) {
	casos := []struct {
		nombre     string
		partida    func() *Session
		transicion func(*Session) error
	}{
		{
			nombre:  "crear sobre una sesión con una activa",
			partida: func() *Session { return sesionCon(conv("cv1", "la primera", true, t0, t1, 3)) },
			transicion: func(s *Session) error {
				_, _ = s.CrearConversacion("cv2", t2)
				return nil
			},
		},
		{
			nombre: "activar una inactiva",
			partida: func() *Session {
				return sesionCon(
					conv("cv1", "la primera", true, t0, t1, 3),
					conv("cv2", "la segunda", false, t1, t2, 1),
				)
			},
			transicion: func(s *Session) error {
				_, err := s.ActivarConversacion("cv2")
				return err
			},
		},
		{
			nombre: "activar la que YA está activa (no-op, E-15)",
			partida: func() *Session {
				return sesionCon(
					conv("cv1", "la primera", true, t0, t1, 3),
					conv("cv2", "la segunda", false, t1, t2, 1),
				)
			},
			transicion: func(s *Session) error {
				desactivada, err := s.ActivarConversacion("cv1")
				if desactivada != "" {
					t.Errorf("activar la activa desactivó %q — tiene que ser no-op", desactivada)
				}
				return err
			},
		},
		{
			nombre:  "renombrar",
			partida: func() *Session { return sesionCon(conv("cv1", "la primera", true, t0, t1, 3)) },
			transicion: func(s *Session) error {
				return s.RenombrarConversacion("cv1", "el bug del índice")
			},
		},
		{
			nombre:  "normalizar una sesión sin conversaciones (E-01)",
			partida: func() *Session { return sesionCon() },
			transicion: func(s *Session) error {
				s.NormalizarConversaciones(t2)
				return nil
			},
		},
		{
			nombre: "normalizar una sesión con DOS activas",
			partida: func() *Session {
				return sesionCon(
					conv("cv1", "la primera", true, t0, t1, 3),
					conv("cv2", "la segunda", true, t1, t2, 1),
				)
			},
			transicion: func(s *Session) error {
				s.NormalizarConversaciones(t2)
				return nil
			},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s := c.partida()
			if err := c.transicion(s); err != nil {
				t.Fatalf("la transición devolvió error: %v", err)
			}
			if err := VerificarUnaActiva(*s); err != nil {
				t.Fatalf("la invariante quedó rota tras la transición: %v", err)
			}
			if _, ok := s.Activa(); !ok {
				t.Fatal("VerificarUnaActiva pasó pero Activa() no encuentra ninguna — los dos leen lo mismo y se contradicen")
			}
		})
	}
}

// TestCrearDesactivaLaAnterior — E-06. Crear es UNA transición: nace activa con 0 turnos
// y ctx 0 %, y la anterior queda inactiva CON su Conv y su Checkpoint intactos (RF-305).
func TestCrearDesactivaLaAnterior(t *testing.T) {
	anterior := conv("cv1", "por qué el mapa sale vacío", true, t0, t1, 90)
	anterior.Checkpoint = "digest de la rotación"
	anterior.CadenaCC = []string{"cc-uno", "cc-dos"}
	anterior.CtxPct = 68
	s := sesionCon(anterior)

	nueva, desactivada := s.CrearConversacion("cv2", t2)

	if desactivada != "cv1" {
		t.Errorf("desactivada = %q, se esperaba %q — el vacío la tiene que nombrar", desactivada, "cv1")
	}
	if len(s.Conversaciones) != 2 {
		t.Fatalf("quedaron %d conversaciones, se esperaban 2", len(s.Conversaciones))
	}
	if !nueva.Activa {
		t.Error("la nueva no nació activa")
	}
	if nueva.Titulo != TituloConversacionNueva {
		t.Errorf("título de la nueva = %q, se esperaba %q (RF-301 CA-3)", nueva.Titulo, TituloConversacionNueva)
	}
	if nueva.NumTurnos() != 0 {
		t.Errorf("la nueva nació con %d turnos, se esperaban 0", nueva.NumTurnos())
	}
	if nueva.CtxPct != 0 {
		t.Errorf("la nueva nació con ctx %d %%, se esperaba 0", nueva.CtxPct)
	}
	if nueva.Conv == nil {
		t.Error("la nueva nació con Conv nil — [] y nil no significan lo mismo (no-aplica-no-es-cero)")
	}
	if nueva.CreadaEn != "2026-07-26T12:00:00Z" {
		t.Errorf("CreadaEn = %q, se esperaba el sello RFC3339 UTC de ahora", nueva.CreadaEn)
	}

	vieja := s.Conversaciones[0]
	if vieja.Activa {
		t.Error("la anterior siguió activa")
	}
	if vieja.NumTurnos() != 90 {
		t.Errorf("la anterior perdió turnos: %d de 90 — desactivar NO es archivar (RF-305)", vieja.NumTurnos())
	}
	if vieja.Checkpoint == "" {
		t.Error("la anterior perdió su Checkpoint al desactivarse (RF-305)")
	}
	if len(vieja.CadenaCC) != 2 {
		t.Errorf("la anterior perdió su CadenaCC: %d de 2", len(vieja.CadenaCC))
	}
}

// TestCrearSobreSesionVaciaNoNombraDesactivada — el "" del contrato es un dato: una
// sesión sin conversaciones no desactivó a nadie, y el vacío no tiene a quién nombrar.
func TestCrearSobreSesionVaciaNoNombraDesactivada(t *testing.T) {
	s := sesionCon()
	_, desactivada := s.CrearConversacion("cv1", t0)
	if desactivada != "" {
		t.Errorf("desactivada = %q, se esperaba vacío", desactivada)
	}
}

// TestActivarConversacionAjenaEs404 — E-11 / RF-343 CA-3, la mitad de dominio: una
// conversación cuelga de SU sesión (BR-CV-2). Jamás se busca globalmente, así que pedir
// una ajena no es «no autorizado»: es «acá no está».
func TestActivarConversacionAjenaEs404(t *testing.T) {
	s := sesionCon(
		conv("cv1", "la primera", true, t0, t1, 3),
		conv("cv2", "la segunda", false, t1, t2, 1),
	)
	antes := s.Conversaciones[0].ID

	desactivada, err := s.ActivarConversacion("cv-de-otra-sesion")

	if !errors.Is(err, ErrConvNoEncontrada) {
		t.Fatalf("err = %v, se esperaba ErrConvNoEncontrada", err)
	}
	if desactivada != "" {
		t.Errorf("desactivada = %q — un fallo no desactiva nada (RF-308 CA-2)", desactivada)
	}
	activa, ok := s.Activa()
	if !ok || activa.ID != antes {
		t.Error("la activa cambió pese al error — el estado anterior tiene que quedar intacto")
	}
	if !strings.Contains(err.Error(), "cv-de-otra-sesion") {
		t.Errorf("el error no nombra el id pedido: %q", err.Error())
	}
	if invErr := VerificarUnaActiva(*s); invErr != nil {
		t.Fatalf("la invariante quedó rota tras un error: %v", invErr)
	}
}

// TestRenombrarVacioNoCambiaElTitulo — E-30. Descartar es dejar el título anterior, no
// guardar uno vacío y no inventar uno.
func TestRenombrarVacioNoCambiaElTitulo(t *testing.T) {
	for _, entrada := range []string{"", "   ", "\t\n "} {
		s := sesionCon(conv("cv1", "el bug del índice", true, t0, t1, 3))
		err := s.RenombrarConversacion("cv1", entrada)
		if !errors.Is(err, ErrTituloVacio) {
			t.Fatalf("entrada %q: err = %v, se esperaba ErrTituloVacio", entrada, err)
		}
		if got := s.Conversaciones[0].Titulo; got != "el bug del índice" {
			t.Errorf("entrada %q: el título cambió a %q", entrada, got)
		}
		if s.Conversaciones[0].TituloEditado {
			t.Errorf("entrada %q: un renombre rechazado marcó TituloEditado", entrada)
		}
	}

	// El camino feliz recorta y marca.
	s := sesionCon(conv("cv1", "nueva conversación", true, t0, t1, 3))
	if err := s.RenombrarConversacion("cv1", "  el bug del índice  "); err != nil {
		t.Fatalf("renombrar válido devolvió error: %v", err)
	}
	if got := s.Conversaciones[0].Titulo; got != "el bug del índice" {
		t.Errorf("título = %q, se esperaba recortado", got)
	}
	if !s.Conversaciones[0].TituloEditado {
		t.Error("un renombre exitoso no marcó TituloEditado (RF-303 CA-1)")
	}

	// Renombrar una ajena es 404, no 400.
	if err := s.RenombrarConversacion("cv-ajena", "cualquier cosa"); !errors.Is(err, ErrConvNoEncontrada) {
		t.Errorf("renombrar una ajena: err = %v, se esperaba ErrConvNoEncontrada", err)
	}
}

// TestNormalizarReparaYLoDice — E-01. Los tres modos de reparación, y el cuarto caso: el
// que no repara nada NO devuelve una línea de log (un arranque sin novedades no habla).
func TestNormalizarReparaYLoDice(t *testing.T) {
	t.Run("0 conversaciones crea una activa vacía", func(t *testing.T) {
		s := sesionCon()
		informe := s.NormalizarConversaciones(t2)
		if len(informe) != 1 {
			t.Fatalf("informe = %v, se esperaba exactamente 1 línea", informe)
		}
		if len(s.Conversaciones) != 1 || !s.Conversaciones[0].Activa {
			t.Fatalf("no quedó exactamente una activa: %+v", s.Conversaciones)
		}
		if s.Conversaciones[0].NumTurnos() != 0 {
			t.Error("la creada por reparación no nació vacía")
		}
		if !strings.Contains(informe[0], s.Conversaciones[0].ID) {
			t.Errorf("la línea no nombra el id creado: %q", informe[0])
		}
	})

	t.Run("ninguna activa activa la de última interacción más reciente", func(t *testing.T) {
		s := sesionCon(
			conv("cv-vieja", "la vieja", false, t0, t0, 5),
			conv("cv-nueva", "la nueva", false, t0, t2, 2),
			conv("cv-sin-turnos", "sin turnos", false, t1, time.Time{}, 0),
		)
		informe := s.NormalizarConversaciones(t2)
		if len(informe) != 1 {
			t.Fatalf("informe = %v, se esperaba 1 línea", informe)
		}
		activa, ok := s.Activa()
		if !ok || activa.ID != "cv-nueva" {
			t.Fatalf("se activó %v, se esperaba cv-nueva (última interacción más reciente)", activa)
		}
		if err := VerificarUnaActiva(*s); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ninguna activa y ninguna con turnos desempata por CreadaEn", func(t *testing.T) {
		s := sesionCon(
			conv("cv-a", "a", false, t0, time.Time{}, 0),
			conv("cv-b", "b", false, t2, time.Time{}, 0),
			conv("cv-c", "c", false, t1, time.Time{}, 0),
		)
		s.NormalizarConversaciones(t2)
		activa, _ := s.Activa()
		if activa.ID != "cv-b" {
			t.Errorf("se activó %q, se esperaba cv-b (CreadaEn más reciente)", activa.ID)
		}
	})

	t.Run("dos activas conserva una y lo dice", func(t *testing.T) {
		s := sesionCon(
			conv("cv1", "la vieja", true, t0, t0, 5),
			conv("cv2", "la nueva", true, t1, t2, 2),
			conv("cv3", "inactiva", false, t1, t1, 1),
		)
		informe := s.NormalizarConversaciones(t2)
		if len(informe) != 1 {
			t.Fatalf("informe = %v, se esperaba 1 línea", informe)
		}
		if !strings.Contains(informe[0], "cv2") {
			t.Errorf("la línea no nombra la conservada: %q", informe[0])
		}
		activa, _ := s.Activa()
		if activa.ID != "cv2" {
			t.Errorf("se conservó %q, se esperaba cv2", activa.ID)
		}
		if err := VerificarUnaActiva(*s); err != nil {
			t.Fatal(err)
		}
		if len(s.Conversaciones) != 3 {
			t.Errorf("normalizar borró conversaciones: quedaron %d de 3", len(s.Conversaciones))
		}
	})

	t.Run("una sola activa no repara ni habla", func(t *testing.T) {
		s := sesionCon(
			conv("cv1", "la primera", true, t0, t1, 3),
			conv("cv2", "la segunda", false, t1, t2, 1),
		)
		if informe := s.NormalizarConversaciones(t2); informe != nil {
			t.Errorf("informe = %v, se esperaba nil: no había nada que reparar", informe)
		}
	})
}

// TestVerificarUnaActivaDistingueLosTresEstadosRotos — el predicado no es un booleano
// disfrazado: dice CUÁL de las tres violaciones ocurrió, porque el log de arranque la cita.
func TestVerificarUnaActivaDistingueLosTresEstadosRotos(t *testing.T) {
	if err := VerificarUnaActiva(*sesionCon()); !errors.Is(err, ErrSinActiva) {
		t.Errorf("sin conversaciones: err = %v, se esperaba ErrSinActiva", err)
	}
	sinActiva := sesionCon(conv("cv1", "a", false, t0, t1, 1))
	if err := VerificarUnaActiva(*sinActiva); !errors.Is(err, ErrSinActiva) {
		t.Errorf("sin activa: err = %v, se esperaba ErrSinActiva", err)
	}
	dosActivas := sesionCon(conv("cv1", "a", true, t0, t1, 1), conv("cv2", "b", true, t1, t2, 1))
	err := VerificarUnaActiva(*dosActivas)
	if err == nil {
		t.Fatal("dos activas: no dio error")
	}
	if errors.Is(err, ErrSinActiva) {
		t.Error("dos activas se reportó como «sin activa» — son violaciones distintas y el log las tiene que distinguir")
	}
	if !strings.Contains(err.Error(), "2") {
		t.Errorf("el error no dice cuántas activas había: %q", err.Error())
	}
}

// TestTituloEditadoNoSeReDeriva — E-33. Cuando el operador y un turno escriben el mismo
// campo, gana el operador. Esta es la mitad de dominio de esa ley: el derivador vive en
// el usecase, la ley de precedencia vive acá.
func TestTituloEditadoNoSeReDeriva(t *testing.T) {
	s := sesionCon(conv("cv1", TituloConversacionNueva, true, t0, time.Time{}, 0))

	// Antes de que el operador toque nada, el primer turno sí deriva.
	if !s.Conversaciones[0].DerivarTitulo("no veo nada en el mapa, ¿podés repararlo?") {
		t.Fatal("el primer turno no derivó el título de una conversación recién nacida")
	}
	if got := s.Conversaciones[0].Titulo; got != "no veo nada en el mapa, ¿podés repararlo?" {
		t.Fatalf("título derivado = %q", got)
	}

	// El operador renombra.
	if err := s.RenombrarConversacion("cv1", "el bug del índice"); err != nil {
		t.Fatal(err)
	}

	// Y llegan 20 turnos más, cada uno con su derivación candidata.
	for i := range 20 {
		if s.Conversaciones[0].DerivarTitulo("otro texto cualquiera") {
			t.Fatalf("el turno %d re-derivó el título editado por el operador (RF-303 CA-1)", i)
		}
	}
	if got := s.Conversaciones[0].Titulo; got != "el bug del índice" {
		t.Errorf("título = %q, se esperaba el del operador", got)
	}

	// Un derivado vacío tampoco pisa un título sin editar.
	limpia := Conversacion{Titulo: TituloConversacionNueva}
	if limpia.DerivarTitulo("   ") {
		t.Error("un derivado vacío cambió el título")
	}
	if limpia.Titulo != TituloConversacionNueva {
		t.Errorf("título = %q tras un derivado vacío", limpia.Titulo)
	}
}
