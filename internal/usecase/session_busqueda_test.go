package usecase_test

// Tests del buscador del panel (T17 · RF-321/322/324/341 · CV-D8 · BR-CV-8).
//
// El fixture `testdata/conv-90-turnos.json` es un transcript REAL —90 turnos, 12 095 B de
// payload, sacado del registro del operador— y no uno sintético. Un transcript inventado
// tendría el largo que uno quiera y las palabras que uno quiera: mediría el test, no el
// buscador. Este trae 75 pasos de actividad, 13 respuestas y 2 mensajes del operador, que es
// la proporción real de una conversación de trabajo y la que hace que buscar «en el texto»
// signifique algo.

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// transcriptReal carga el fixture. Falla ruidoso si no está: un buscador probado contra cero
// turnos pasa por vacuidad.
func transcriptReal(t *testing.T) []domain.Turn {
	t.Helper()
	b, err := os.ReadFile("testdata/conv-90-turnos.json")
	if err != nil {
		t.Fatalf("no se pudo leer el fixture real: %v", err)
	}
	var turnos []domain.Turn
	if err := json.Unmarshal(b, &turnos); err != nil {
		t.Fatalf("el fixture no parsea: %v", err)
	}
	if len(turnos) != 90 {
		t.Fatalf("el fixture tiene %d turnos, se esperaban 90 — ¿se reemplazó por uno sintético?", len(turnos))
	}
	return turnos
}

// svcConSesion arma un servicio cuyo registro ya trae una sesión con las conversaciones dadas.
func svcConSesion(t *testing.T, convs ...domain.Conversacion) *usecase.SessionService {
	t.Helper()
	st := &storeQueFalla{sesiones: []domain.Session{{
		ID: "s1", Arnes: "vitalia", Frente: "repro del bug de carga", Conversaciones: convs,
	}}}
	return svcConv(t, &stubAgent{}, st, stubPub{})
}

func conv(id, titulo string, activa bool, textos ...string) domain.Conversacion {
	c := domain.Conversacion{ID: id, Titulo: titulo, Activa: activa, Conv: []domain.Turn{}}
	for _, tx := range textos {
		c.Conv = append(c.Conv, domain.Turn{Rol: domain.RolUser, Text: tx})
	}
	return c
}

func ids(rs []usecase.ConversacionResumen) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}

// TestBuscaEnTituloYEnTexto (E-03 · RF-321 CA-1): el buscador entra al texto del transcript,
// no sólo al título. Era el pedido explícito del operador — «buscar por lo que dije, no por
// cómo se llama la conversación».
func TestBuscaEnTituloYEnTexto(t *testing.T) {
	svc := svcConSesion(t,
		conv("cv1", "el manifiesto vacío", true),
		conv("cv2", "otra cosa", false, "acá hablamos del manifiesto y de nada más"),
		conv("cv3", "tercera", false, "esto no menciona la palabra"),
	)
	got, total, err := svc.Conversaciones("s1", "manifiesto")
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Errorf("total = %d, quiero 3 (el TOTAL de la sesión, no el de coincidencias)", total)
	}
	if g := ids(got); len(g) != 2 || g[0] != "cv1" || g[1] != "cv2" {
		t.Errorf("coincidencias = %v, quiero [cv1 cv2] — una por título y otra por texto", g)
	}
}

// TestBusquedaInsensibleAAcentosYMayusculas (E-23 · RF-324): «VACIO» encuentra «vacío». Es
// superset del buscador que el rail ya tiene, que sólo baja a minúsculas: el nuevo no puede
// ser peor que el viejo.
func TestBusquedaInsensibleAAcentosYMayusculas(t *testing.T) {
	svc := svcConSesion(t,
		conv("cv1", "una", true, "el manifiesto está vacío"),
		conv("cv2", "MAÑANA seguimos", false),
		conv("cv3", "tres", false, "sin coincidencias acá"),
	)
	for _, caso := range []struct{ q, quiero string }{
		{"VACIO", "cv1"},
		{"vacío", "cv1"},
		{"ESTA", "cv1"},
		{"manana", "cv2"},
		{"MaÑaNa", "cv2"},
	} {
		got, _, err := svc.Conversaciones("s1", caso.q)
		if err != nil {
			t.Fatal(err)
		}
		if g := ids(got); len(g) != 1 || g[0] != caso.quiero {
			t.Errorf("buscar %q = %v, quiero [%s]", caso.q, g, caso.quiero)
		}
	}
}

// TestQueryEnBlancoEsSinQuery (E-26): un `q` de puros espacios es «sin filtro», no «buscá el
// vacío». Devolver cero por un espacio accidental sería decirle al operador que su sesión no
// tiene conversaciones.
func TestQueryEnBlancoEsSinQuery(t *testing.T) {
	svc := svcConSesion(t,
		conv("cv1", "una", true),
		conv("cv2", "dos", false),
	)
	for _, q := range []string{"", "   ", "\t\n "} {
		got, total, err := svc.Conversaciones("s1", q)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 || total != 2 {
			t.Errorf("q=%q devolvió %d de %d, quiero las 2", q, len(got), total)
		}
		for _, r := range got {
			if r.Fragmento != "" {
				t.Errorf("q=%q trajo fragmento %q: sin búsqueda no hay nada que recortar", q, r.Fragmento)
			}
		}
	}
}

// TestFragmentoEsElPrimerMatch (E-29 · RF-322 CA-1): un fragmento por fila, el PRIMERO, con
// elipsis a los dos lados. El texto sale con sus acentos y mayúsculas originales —se busca
// sobre el plegado, se recorta sobre el crudo— y en texto plano: el resaltado es del FE.
func TestFragmentoEsElPrimerMatch(t *testing.T) {
	largo := strings.Repeat("relleno ", 20) + "el MANIFIESTO declara 0 elementos " + strings.Repeat("cola ", 20) + "manifiesto otra vez"
	svc := svcConSesion(t, conv("cv1", "una", true, "sin nada", largo))

	got, _, err := svc.Conversaciones("s1", "manifiesto")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("coincidencias = %d", len(got))
	}
	f := got[0].Fragmento
	if !strings.Contains(f, "MANIFIESTO") {
		t.Errorf("fragmento = %q: tiene que traer el texto ORIGINAL, con sus mayúsculas", f)
	}
	if !strings.HasPrefix(f, "…") || !strings.HasSuffix(f, "…") {
		t.Errorf("fragmento = %q: se recortó de los dos lados, tiene que decirlo con elipsis", f)
	}
	if strings.Count(f, "manifiesto")+strings.Count(f, "MANIFIESTO") != 1 {
		t.Errorf("fragmento = %q: es UNO, el primero — no la concatenación de todos", f)
	}
	if strings.Contains(f, "<mark") || strings.Contains(f, "<") {
		t.Errorf("fragmento = %q: el daemon manda texto plano, el resaltado es del FE", f)
	}
}

// TestCoincideSoloEnTituloNoTraeFragmento (E-27 · RF-322 CA-3): si coincidió el título, no
// hay nada que explicar. Un fragmento vacío ahí sería una línea en blanco con sangría.
func TestCoincideSoloEnTituloNoTraeFragmento(t *testing.T) {
	svc := svcConSesion(t, conv("cv1", "el bug del índice", true, "nada de esto menciona la palabra buscada"))
	got, _, err := svc.Conversaciones("s1", "índice")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("coincidencias = %d, quiero 1", len(got))
	}
	if got[0].Fragmento != "" {
		t.Errorf("fragmento = %q: la coincidencia fue en el título", got[0].Fragmento)
	}
}

// TestBusquedaSobreTranscriptRealDe90Turnos (E-25): el fixture es el transcript real. Se
// verifica que el buscador entre a los tres roles que una conversación de trabajo produce —el
// mensaje del operador, la respuesta del conductor y los pasos de actividad— porque cada uno
// es un lugar distinto donde el operador puede recordar haber visto algo.
func TestBusquedaSobreTranscriptRealDe90Turnos(t *testing.T) {
	turnos := transcriptReal(t)
	svc := svcConSesion(t,
		domain.Conversacion{ID: "cv-real", Titulo: "repro del bug de carga", Activa: true, Conv: turnos},
		conv("cv-vacia", "sin turnos", false),
	)

	// Una palabra del MENSAJE del operador (rol user).
	got, total, err := svc.Conversaciones("s1", "mapa")
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Errorf("total = %d, quiero 2", total)
	}
	if g := ids(got); len(g) != 1 || g[0] != "cv-real" {
		t.Fatalf("buscar «mapa» = %v, quiero [cv-real]", g)
	}
	if got[0].Fragmento == "" {
		t.Error("una coincidencia en el texto tiene que traer su fragmento")
	}
	if got[0].Turnos != 90 {
		t.Errorf("turnos = %d, quiero 90 — se derivan del transcript, no de un campo aparte", got[0].Turnos)
	}
	// El resumen NO declara los turnos: no es una omisión de serialización, es que el
	// tipo no los tiene. Se verifica sobre el JSON, que es lo que efectivamente viaja.
	crudo, err := json.Marshal(got[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(crudo), `"conv"`) {
		t.Errorf("el resumen serializó el transcript: %s", crudo)
	}

	// Una palabra de un PASO DE ACTIVIDAD (rol act): 75 de los 90 turnos son eso.
	got, _, err = svc.Conversaciones("s1", "git rev-parse")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !strings.Contains(got[0].Fragmento, "git rev-parse") {
		t.Errorf("buscar en los pasos de actividad = %+v", got)
	}

	// Y algo que NO está: cero coincidencias, no «todas» ni un error.
	got, total, err = svc.Conversaciones("s1", "telemetría")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("buscar algo ausente = %v, quiero ninguna", ids(got))
	}
	if total != 2 {
		t.Errorf("total = %d con cero coincidencias, quiero 2: es el denominador del «N de M»", total)
	}
}

// TestTotalEsElTotalNoElDeCoincidencias (E-22 · RF-318 CA-2): el rótulo «N de M coinciden»
// necesita las dos cifras, y M es el total de la sesión. Devolver el total de coincidencias
// haría que el rótulo dijera siempre «N de N».
func TestTotalEsElTotalNoElDeCoincidencias(t *testing.T) {
	svc := svcConSesion(t,
		conv("cv1", "una", true, "manifiesto"),
		conv("cv2", "dos", false, "manifiesto"),
		conv("cv3", "tres", false),
		conv("cv4", "cuatro", false),
	)
	got, total, err := svc.Conversaciones("s1", "manifiesto")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || total != 4 {
		t.Errorf("«%d de %d coinciden», quiero «2 de 4»", len(got), total)
	}
}

// TestConversacionesDeSesionInexistenteEsError (E-35 · RF-340 CA-3 · BR-CV-10): jamás una
// lista vacía por una sesión que no existe. Vacío y ausente son cosas distintas, y confundir
// las dos deja al operador mirando un panel que dice «no hay nada» cuando el problema es otro.
func TestConversacionesDeSesionInexistenteEsError(t *testing.T) {
	svc := svcConSesion(t, conv("cv1", "una", true))
	if _, _, err := svc.Conversaciones("s-que-no-existe", ""); err == nil {
		t.Fatal("una sesión inexistente tiene que ser un error (→ 404), no una lista vacía")
	}
}

// TestLaListaEsSoloDeEstaSesion (CV-D4 · RF-318 CA-3): la lista jamás trae una conversación
// de otra sesión. Se fija en el dominio, no en la UI: ahí sólo se podría esconder.
func TestLaListaEsSoloDeEstaSesion(t *testing.T) {
	st := &storeQueFalla{sesiones: []domain.Session{
		{ID: "s1", Arnes: "vitalia", Conversaciones: []domain.Conversacion{conv("cv-a", "de s1", true, "manifiesto")}},
		{ID: "s2", Arnes: "vitalia", Conversaciones: []domain.Conversacion{conv("cv-b", "de s2", true, "manifiesto")}},
	}}
	svc := svcConv(t, &stubAgent{}, st, stubPub{})

	got, total, err := svc.Conversaciones("s1", "manifiesto")
	if err != nil {
		t.Fatal(err)
	}
	if g := ids(got); len(g) != 1 || g[0] != "cv-a" || total != 1 {
		t.Errorf("la lista de s1 = %v (total %d): se coló una conversación de otra sesión", g, total)
	}
}
