package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/descubrimiento"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/hooks"
)

// hook.go es `arnesia hook proceso` (decisión A7): **el hook que viaja dentro del arnés es el
// propio binario.** Cero runtime nuevo que shipear, cero intérprete que asumir (Python no
// está garantizado en Windows, Node tampoco, y un `.sh` no corre en `cmd.exe`), y **el mismo
// código Go** de proyección que el daemon usa y testea — dos implementaciones de la misma
// allowlist es la forma segura de que una se olvide de un campo.
//
// El contrato, que no admite matices:
//
//	entrada     el payload del hook por stdin, tal cual
//	stdout      VACÍA. Un hook que imprime inyecta texto al contexto del agente
//	exit code   SIEMPRE 0. Sin excepciones. Un UserPromptSubmit que sale ≠0 bloquea el turno
//	duración    tope duro de 250 ms; pasado eso abandona y sale 0
//	destino     la ficha del daemon → loopback. JAMÁS red externa
//	sin daemon  fail-open silencioso: no reintenta, no escribe a disco, no avisa
//	qué manda   el evento YA PROYECTADO. NUNCA su stdin
//
// ⚠️ El runtime garantiza que un comando de hook que **falta** no rompe nada (medido, H10.3).
// Que este binario, que **existe**, no FALLE sigue siendo obligación nuestra — y por eso
// `runHook` no tiene una sola rama que devuelva error.
//
// «Fail-open silencioso» es silencioso **hacia el usuario**, no invisible: el hueco aparece
// después como turnos que no reportaron y como «sin dato». Nunca como 0.

// topeHook es el presupuesto de pared del hook entero (§11).
const topeHook = 250 * time.Millisecond

// runHook despacha los subcomandos del hook. **Devuelve nil SIEMPRE**: el `main` traduce el
// error a `os.Exit(1)`, y un exit ≠0 en `UserPromptSubmit` bloquea el turno del usuario.
func runHook(args []string) error {
	if len(args) == 0 || args[0] != "proceso" {
		// Ni siquiera un subcomando equivocado justifica romper un turno. Se avisa por
		// stderr —que el runtime no inyecta al contexto— y se sale 0.
		fmt.Fprintln(os.Stderr, "arnesia hook: uso: arnesia hook proceso  (el payload va por stdin)")
		return nil
	}
	hookProceso(os.Stdin)
	return nil
}

// hookProceso es el cuerpo real, separado para poder testearlo con un stdin cualquiera.
// **No devuelve error a propósito**: no hay ninguna rama de la que se pueda salir ≠0.
func hookProceso(entrada io.Reader) {
	ctx, cancelar := context.WithTimeout(context.Background(), topeHook)
	defer cancelar()

	// Se lee acotado: un stdin gigante no puede consumir el presupuesto entero.
	crudo, err := io.ReadAll(io.LimitReader(entrada, 1<<20))
	if err != nil || len(crudo) == 0 {
		return // fail-open: sin payload no hay nada que proyectar.
	}

	// 🔴 La proyección va PRIMERO, antes de mirar siquiera si hay daemon. El payload trae el
	// prompt, la respuesta y la salida de las herramientas en claro; nada de eso puede
	// existir en memoria más allá de esta línea, ni salir a ningún lado.
	ev, perr := hooks.Proyectar(crudo, nil, time.Now())
	if perr != nil {
		return // payload inválido o evento fuera del alcance del MVP: se sale sin ruido.
	}

	f, ferr := descubrimiento.New("")
	if ferr != nil {
		return
	}
	// La ficha se relee en CADA invocación (B4): el daemon pudo reiniciarse en otro puerto
	// entre dos turnos, y una cacheada apuntaría al viejo.
	ficha, lerr := f.Leer()
	if lerr != nil {
		return // sin daemon publicado: fail-open silencioso.
	}

	cuerpo, merr := json.Marshal(ev)
	if merr != nil {
		return
	}
	destino := ficha.Endpoint + ficha.RutaProceso
	req, rerr := http.NewRequestWithContext(ctx, http.MethodPost, destino, bytes.NewReader(cuerpo))
	if rerr != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if ficha.TokenIngesta != "" {
		// El token viene de la ficha `0600`, leída en runtime. Nada versionable.
		req.Header.Set("X-Arnesia-Token", ficha.TokenIngesta)
	}
	// El Host gate exige un host loopback esperado: el endpoint de la ficha ya lo es.
	resp, herr := (&http.Client{Timeout: topeHook}).Do(req)
	if herr != nil {
		return // el daemon murió sin retirar la ficha: se falla rápido y se sale 0.
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
	_ = resp.Body.Close()
	// Un 4xx/5xx del daemon tampoco se reporta: el hook no reintenta y no avisa. El hueco
	// aparece después en la cobertura, que es donde tiene que aparecer.
}
