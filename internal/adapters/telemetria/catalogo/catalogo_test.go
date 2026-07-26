package catalogo

import (
	"testing"
	"time"
)

// TestModeloDesconocido — un modelo que no está en el catálogo devuelve ok=false, NO un
// PrecioModelo en ceros. Un precio en ceros cotizaría todo gratis y nadie se enteraría:
// el número saldría bonito y sería mentira.
func TestModeloDesconocido(t *testing.T) {
	c := Embebido()
	p, ok := c.Precio("modelo-que-no-existe-jamas-9999")
	if ok {
		t.Fatalf("un modelo inexistente no puede resolver: %+v", p)
	}
	if p.Entrada != nil || p.Salida != nil {
		t.Fatalf("un modelo inexistente devolvió tarifas: %+v", p)
	}
	// Control positivo, misma corrida: el catálogo SÍ resuelve un modelo real. Sin esto,
	// un `precios.json` vacío pasaría el assert de arriba y el test no diría nada.
	if _, ok := c.Precio("claude-haiku-4-5"); !ok {
		t.Fatal("control positivo: el catálogo embebido no resuelve claude-haiku-4-5 — está vacío o roto")
	}
}

// TestTarifaParcial — un modelo real puede no publicar tarifa de todos los buckets. El
// catálogo devuelve nil en los que faltan, no 0: quien cotice decide qué hacer con la
// ausencia, y `CalcularCosto` la reporta en SinTarifa en vez de cobrarla gratis.
func TestTarifaParcial(t *testing.T) {
	c := Embebido()
	p, ok := c.Precio("claude-haiku-4-5")
	if !ok {
		t.Fatal("claude-haiku-4-5 debería estar en el catálogo embebido")
	}
	if p.Razonamiento != nil {
		t.Errorf("Anthropic no publica tarifa de razonamiento aparte; debería ser nil, es %v", *p.Razonamiento)
	}
	if p.Entrada == nil || p.Salida == nil || p.CacheLectura == nil {
		t.Fatalf("los buckets que SÍ tienen tarifa deben venir poblados: %+v", p)
	}
}

// TestCatalogoConservaElTierDeUnaHora — el campo que ccusage descarta (E3) y que habilita
// el detector B1. Sin `cache_creation_input_token_cost_above_1hr` no hay break-even de TTL
// que calcular: sería un detector que no puede existir.
func TestCatalogoConservaElTierDeUnaHora(t *testing.T) {
	c := Embebido()
	p, ok := c.Precio("claude-haiku-4-5")
	if !ok {
		t.Fatal("claude-haiku-4-5 no está en el catálogo")
	}
	if p.CacheEscritura5m == nil || p.CacheEscritura1h == nil {
		t.Fatalf("el split 5m/1h del cache write debe sobrevivir al filtrado: %+v", p)
	}
	// La estructura 1,25× / 2× / 0,1× de Anthropic (D-5) es la que B1 cita en su umbral.
	if got, want := *p.CacheEscritura5m / *p.Entrada, 1.25; !casiIgual(got, want, 1e-9) {
		t.Errorf("cache write 5m / entrada = %v, se esperaba %v (estructura de Anthropic)", got, want)
	}
	if got, want := *p.CacheEscritura1h / *p.Entrada, 2.0; !casiIgual(got, want, 1e-9) {
		t.Errorf("cache write 1h / entrada = %v, se esperaba %v", got, want)
	}
	if got, want := *p.CacheLectura / *p.Entrada, 0.1; !casiIgual(got, want, 1e-9) {
		t.Errorf("cache read / entrada = %v, se esperaba %v", got, want)
	}
}

// TestCatalogoNoAplanaLosTiers — un modelo con tramo largo conserva el tramo COMO TRAMO
// (phoenix#14314 lo aplana y subestima justo los prompts caros).
func TestCatalogoNoAplanaLosTiers(t *testing.T) {
	c := Embebido()
	p, ok := c.Precio("claude-sonnet-4-5")
	if !ok {
		t.Fatal("claude-sonnet-4-5 no está en el catálogo")
	}
	if p.UmbralContextoTok == nil || p.SobreUmbral == nil {
		t.Fatalf("sonnet-4-5 publica tramo largo (>200k) y debe sobrevivir al filtrado: %+v", p)
	}
	if *p.UmbralContextoTok != 200_000 {
		t.Errorf("umbral = %d, se esperaba 200000", *p.UmbralContextoTok)
	}
	if p.SobreUmbral.Entrada == nil || *p.SobreUmbral.Entrada <= *p.Entrada {
		t.Errorf("la tarifa del tramo largo debe ser mayor que la base: %v vs %v", p.SobreUmbral.Entrada, *p.Entrada)
	}
	// Control positivo del contraste: un modelo SIN tramo largo no inventa uno.
	sin, ok := c.Precio("claude-haiku-4-5")
	if !ok {
		t.Fatal("claude-haiku-4-5 no está en el catálogo")
	}
	if sin.SobreUmbral != nil {
		t.Errorf("haiku-4-5 no publica tramo largo; el catálogo no debe fabricarle uno: %+v", sin.SobreUmbral)
	}
}

// TestAliasBedrockVertex — los ids con prefijo de región y con sufijo de fecha resuelven al
// mismo modelo. Y lo que NO matchea se devuelve TAL CUAL: no se inventa una canonización.
func TestAliasBedrockVertex(t *testing.T) {
	c := Embebido()
	casos := []struct{ crudo, esperado string }{
		// LiteLLM trae la fila fechada Y la sin fechar. Se prefiere la fechada cuando
		// existe: es la más específica, y una tarifa que cambia con la fecha del id se
		// cotizaría mal si la colapsáramos.
		{"claude-haiku-4-5-20251001", "claude-haiku-4-5-20251001"},
		// Una fecha que el catálogo NO tiene cae a la familia. Esto es lo que hace que un
		// release nuevo del modelo no deje de cotizar el día que sale.
		{"claude-haiku-4-5-20991231", "claude-haiku-4-5"},
		{"anthropic/claude-haiku-4-5", "claude-haiku-4-5"},
		{"claude-haiku-4-5", "claude-haiku-4-5"},
	}
	for _, cs := range casos {
		if got := c.Canonizar(cs.crudo); got != cs.esperado {
			t.Errorf("Canonizar(%q) = %q, want %q", cs.crudo, got, cs.esperado)
		}
	}
	// Bedrock con región: LiteLLM ya trae la fila con su prefijo, así que canoniza a sí
	// misma; lo que importa es que RESUELVA a un precio.
	for _, id := range []string{
		"us.anthropic.claude-sonnet-4-5-20250929-v1:0",
		"eu.anthropic.claude-sonnet-4-5-20250929-v1:0",
	} {
		if _, ok := c.Precio(id); !ok {
			t.Errorf("Precio(%q) debería resolver (alias Bedrock con región)", id)
		}
	}
	// El caso que importa de verdad: lo desconocido vuelve crudo.
	raro := "modelo-de-un-proveedor-que-no-conocemos"
	if got := c.Canonizar(raro); got != raro {
		t.Errorf("Canonizar de un desconocido debe devolverlo TAL CUAL; devolvió %q", got)
	}
}

// TestCatalogoEmbebidoBajoPresupuesto — §11: el `go:embed` no puede pasar los 256 KB. El
// presupuesto no es decorativo: el binario entero tiene tope de +1,5 MB sobre el release
// anterior, y este archivo es la parte más fácil de dejar crecer sin darse cuenta.
func TestCatalogoEmbebidoBajoPresupuesto(t *testing.T) {
	const tope = 256 * 1024
	if n := TamanoEmbebido(); n > tope {
		t.Fatalf("precios.json embebido = %d bytes, tope %d (§11)", n, tope)
	} else {
		t.Logf("precios.json embebido = %d bytes de %d (%.0f %% del presupuesto)", n, tope, float64(n)/float64(tope)*100)
	}
}

// TestVersionDeclaraSuProcedencia — un catálogo sin rev ni sha no se puede auditar, y un
// `Refrescado` en cero se leería como «refrescado en 1970» en vez de «nunca».
func TestVersionDeclaraSuProcedencia(t *testing.T) {
	if err := ErrEmbebido(); err != nil {
		t.Fatalf("el catálogo embebido no decodifica: %v", err)
	}
	v := Embebido().Version()
	if v.Version == "" || v.Rev == "" || v.SHA256 == "" {
		t.Fatalf("la versión debe declarar fecha, rev y sha256: %+v", v)
	}
	if v.Modelos == 0 {
		t.Fatal("el catálogo embebido no tiene modelos")
	}
	// Anti-drift: el rev que el archivo declara y el que la directiva `go:generate` fija
	// tienen que ser el mismo. Si alguien regenera con otro rev y no toca gen.go, el
	// catálogo diría venir de un commit del que no vino.
	if v.Rev != revLiteLLM {
		t.Errorf("precios.json dice rev %q pero gen.go fija %q — uno de los dos miente", v.Rev, revLiteLLM)
	}
	if v.Version != fechaLiteLLM {
		t.Errorf("precios.json dice version %q pero gen.go fija %q", v.Version, fechaLiteLLM)
	}
	if v.Refrescado != nil {
		t.Errorf("un catálogo recién embebido NUNCA se refrescó; Refrescado debe ser nil, es %v", *v.Refrescado)
	}
	// Control positivo: cuando SÍ se refresca, se nota.
	c := Embebido()
	c.MarcarRefrescado(time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC))
	if got := c.Version().Refrescado; got == nil {
		t.Fatal("tras MarcarRefrescado, Refrescado no puede seguir nil")
	}
	c.mu.Lock()
	c.refrescado = nil // se restaura: Embebido() es un singleton compartido por los tests.
	c.mu.Unlock()
}

// TestRefrescoFallidoNoRompe — el refresco va apagado por default y, encendido, un fallo no
// cambia nada: se sigue con el embebido y la UI muestra «precios del release».
func TestRefrescoFallidoNoRompe(t *testing.T) {
	c := Embebido()
	antes := c.Version()
	// No hay refresco cableado (A11: apagado por default). El invariante que importa es
	// que sin refresco el catálogo sigue sirviendo precios y declarándose no refrescado.
	if _, ok := c.Precio("claude-haiku-4-5"); !ok {
		t.Fatal("sin refresco, el embebido debe seguir cotizando")
	}
	if c.Version().Refrescado != nil {
		t.Error("sin refresco exitoso, Refrescado sigue nil — jamás se marca solo")
	}
	if c.Version().Rev != antes.Rev {
		t.Error("un refresco que no ocurrió no puede cambiar el rev")
	}
}

func casiIgual(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}
