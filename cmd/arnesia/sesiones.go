package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/descubrimiento"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

// resolverDeLlaves arma el resolvedor de CV-D16 cerrado sobre el Portafolio real: para un
// id de arnés pelado devuelve su clave calificada `(home,id,scope)`, y cuando NO hay
// exactamente una candidata NO decide y dice por qué.
//
// Vive en `cmd` porque el store no puede importar el Portafolio: sólo depende del dominio y
// de los puertos. El grafo de dependencias no es una formalidad — es lo que impide que un
// adaptador de disco termine sabiendo de identidades de arneses.
//
// Las dos vías, y el orden importa:
//
//  1. por **cwd** — la fuerte: la entrada que realmente vive en ese directorio;
//  2. por **id** — el fallback, y no es teórico: de las cuatro sesiones a recalibrar del
//     operador, TRES nunca spawnearon y por lo tanto no tienen cwd. Sin esta vía se
//     quedarían peladas para siempre.
func resolverDeLlaves(entradas []domain.EntradaPortafolio) store.ClaveCalificada {
	// id pelado → claves calificadas candidatas, y clave → los cwd donde vive esa entrada.
	porID := map[string][]string{}
	cwdsDe := map[string][]string{}
	for _, e := range entradas {
		clave := e.Identidad.Clave()
		porID[e.Identidad.ID] = append(porID[e.Identidad.ID], clave)
		if e.Canonico != nil && e.Canonico.Path != "" {
			cwdsDe[clave] = append(cwdsDe[clave], e.Canonico.Path)
		}
		for _, inst := range e.Instalaciones {
			if inst.InstallPath != "" {
				cwdsDe[clave] = append(cwdsDe[clave], inst.InstallPath)
			}
		}
	}

	return func(idPelado, cwd string) (string, bool, string) {
		// Una llave ya calificada no se toca. Es lo que hace idempotente al re-key.
		if strings.Contains(idPelado, "~") {
			return "", false, "ya-calificada"
		}
		candidatas := porID[idPelado]

		if cwd != "" {
			var porCwd []string
			for _, c := range candidatas {
				for _, d := range cwdsDe[c] {
					if mismoDir(d, cwd) {
						porCwd = append(porCwd, c)
						break
					}
				}
			}
			if len(porCwd) == 1 {
				return porCwd[0], true, "resuelta-por-cwd"
			}
		}

		switch len(candidatas) {
		case 1:
			return candidatas[0], true, "resuelta-por-id"
		case 0:
			return "", false, "sin-candidata"
		default:
			// Elegir una de N al azar fusionaría dos identidades por coincidencia, que es
			// exactamente lo que la identidad calificada existe para impedir.
			return "", false, fmt.Sprintf("ambigua: %d", len(candidatas))
		}
	}
}

// daemonPareceVivo consulta la ficha de descubrimiento (`descubrimiento.Ficha`, la misma que
// ya usa `arnesia hook proceso` para encontrar al daemon) para decidir si `arnesia serve`
// parece estar corriendo. Una ficha huérfana (daemon muerto sin shutdown ordenado) da un
// falso positivo — que acá es el lado seguro del error: en el peor caso el operador corre
// con --force de más, nunca pierde datos por asumir que no había nadie escuchando.
func daemonPareceVivo() (bool, string) {
	f, err := descubrimiento.New("")
	if err != nil {
		return false, ""
	}
	ficha, err := f.Leer()
	if err != nil {
		return false, ""
	}
	return true, ficha.Endpoint
}

// mismoDir compara dos rutas por su forma limpia. No resuelve symlinks a propósito: el
// comando puede correr con el daemon detenido y sobre un árbol montado distinto, y una
// comparación que falla es `sin-candidata` (honesto), no una decisión equivocada.
func mismoDir(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

// cmdSesiones — `arnesia sesiones recalibrar-llaves`, la herramienta de CV-D16 con sus
// tres modos:
//
//	arnesia sesiones recalibrar-llaves                       # dry-run: imprime y no escribe
//	arnesia sesiones recalibrar-llaves --aplicar             # re-key idempotente
//	arnesia sesiones recalibrar-llaves --revertir --desde X  # deshace SÓLO el re-key
//
// El dry-run es el default a propósito: es lo que convierte «nada se borra» en algo que el
// operador ve antes de que pase, y no en una afirmación que tiene que creer.
func cmdSesiones(args []string) error {
	// El subcomando se saca ANTES de parsear: el paquete `flag` corta en el primer
	// argumento posicional, así que `sesiones recalibrar-llaves --sesiones X` dejaría el
	// flag sin leer y el comando caería al default en silencio. Lo caza un smoke-test, no
	// la lectura del código.
	if len(args) < 1 {
		return errors.New("uso: arnesia sesiones recalibrar-llaves [flags]")
	}
	sub, args := args[0], args[1:]

	fs := flag.NewFlagSet("sesiones", flag.ContinueOnError)
	sesionesPath := fs.String("sesiones", "", "registro de sesiones (default ~/.arnesia/sesiones.json)")
	aplicar := fs.Bool("aplicar", false, "escribe los cambios; sin este flag es un dry-run")
	revertir := fs.Bool("revertir", false, "deshace el re-key usando --desde, sin tocar las conversaciones")
	desde := fs.String("desde", "", "respaldo del que leer las llaves originales (con --revertir)")
	forzar := fs.Bool("force", false, "escribe aunque el daemon esté corriendo (Fase 1, D3: sin esto, este comando standalone y el daemon vivo son dos cachés en memoria separadas escribiendo el mismo archivo)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `uso: arnesia sesiones recalibrar-llaves [flags]

Recalibra la llave de arnés de cada sesión a su clave calificada (home,id,scope).
Sin flags es un DRY-RUN: imprime la tabla y no escribe nada.

flags:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if sub != "recalibrar-llaves" {
		fs.Usage()
		return fmt.Errorf("arnesia sesiones: subcomando desconocido %q", sub)
	}

	ruta := *sesionesPath
	if ruta == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("arnesia sesiones: no se pudo resolver el home: %w", err)
		}
		ruta = filepath.Join(home, ".arnesia", "sesiones.json")
	}
	reg, err := store.NewRegistry(ruta)
	if err != nil {
		return err
	}

	// D3 (Fase 1): este comando abre su PROPIA instancia de store.Registry, separada de la
	// que vive dentro de `arnesia serve` — igual que el bug confirmado en portafolio.json.
	// A diferencia de portafolio/marketplace, el caché de sesiones vivo pertenece al motor
	// de chat (SessionService), no a este adaptador — tocarlo acá sería cirugía de alto
	// riesgo sobre el camino más caliente del sistema. En su lugar: si el daemon parece
	// vivo y esto va a ESCRIBIR, se rechaza salvo --force.
	if (*aplicar || *revertir) && !*forzar {
		if vivo, motivo := daemonPareceVivo(); vivo {
			return fmt.Errorf("arnesia sesiones: el daemon parece estar corriendo (%s) — "+
				"escribir acá y desde el daemon a la vez puede perder cambios. "+
				"Cerrá la app o corré con --force si estás seguro", motivo)
		}
	}

	if *revertir {
		if *desde == "" {
			return errors.New("arnesia sesiones recalibrar-llaves --revertir: falta --desde <respaldo>")
		}
		respaldo, lerr := store.LeerRespaldo(*desde)
		if lerr != nil {
			return fmt.Errorf("arnesia sesiones: respaldo %s: %w", *desde, lerr)
		}
		filas, rerr := reg.RevertirLlaves(respaldo)
		if rerr != nil {
			return rerr
		}
		imprimirRecalibraciones(filas, false)
		return nil
	}

	// nil: el CLI no observa en Mapa — no necesita el puerto del índice.
	svc, _, _, err := newPortafolioService(nil)
	if err != nil {
		return err
	}
	entradas, _, lerr := svc.Listar(context.Background())
	if lerr != nil {
		return fmt.Errorf("arnesia sesiones: no se pudo leer el Portafolio: %w", lerr)
	}
	filas, rerr := reg.Recalibrar(resolverDeLlaves(entradas), !*aplicar)
	if rerr != nil {
		return rerr
	}
	imprimirRecalibraciones(filas, !*aplicar)
	return nil
}

// imprimirRecalibraciones emite la tabla que el operador transcribe. Una fila por sesión,
// se haya movido o no: un silencio sería indistinguible de «no la miré».
func imprimirRecalibraciones(filas []store.Recalibracion, seco bool) {
	var b strings.Builder
	if seco {
		b.WriteString("DRY-RUN — no se escribió nada. Corré con --aplicar para hacerlo.\n")
	}
	if len(filas) == 0 {
		b.WriteString("no hay sesiones en el registro.\n")
		fmt.Print(b.String())
		return
	}
	fmt.Fprintf(&b, "%-12s  %-28s  %-28s  %s\n", "sesión", "antes", "después", "motivo")
	movidas := 0
	for _, f := range filas {
		marca := " "
		if f.Movio() {
			marca, movidas = "→", movidas+1
		}
		fmt.Fprintf(&b, "%-12s  %-28s %s%-28s  %s\n", f.SesionID, f.Antes, marca, f.Despues, f.Motivo)
	}
	fmt.Fprintf(&b, "\n%d de %d sesiones cambian de llave. Nada se borra y nada se fusiona.\n", movidas, len(filas))
	fmt.Print(b.String())
}
