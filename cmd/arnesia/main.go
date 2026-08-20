// Command arnesia is the ArnesIA daemon. It is the composition root: the only place
// that wires concrete adapters (agent, index, watch, publish) to the use cases and the
// transport. Subcommands: serve | open | index | publish.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	iofs "io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	doctrina "github.com/alpacapurpura/arnesia"
	"github.com/alpacapurpura/arnesia/internal/adapters/agent/claudecode"
	"github.com/alpacapurpura/arnesia/internal/adapters/artifact"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/mechanism"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/ruleset"
	"github.com/alpacapurpura/arnesia/internal/adapters/forja"
	"github.com/alpacapurpura/arnesia/internal/adapters/history"
	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/adapters/loader"
	"github.com/alpacapurpura/arnesia/internal/adapters/logfile"
	"github.com/alpacapurpura/arnesia/internal/adapters/marketplace"
	"github.com/alpacapurpura/arnesia/internal/adapters/permission"
	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/adapters/provision"
	"github.com/alpacapurpura/arnesia/internal/adapters/publish"
	"github.com/alpacapurpura/arnesia/internal/adapters/selfupdate"
	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	stt "github.com/alpacapurpura/arnesia/internal/adapters/stt/local"
	"github.com/alpacapurpura/arnesia/internal/adapters/traer"
	httpapi "github.com/alpacapurpura/arnesia/internal/adapters/transport/http"
	"github.com/alpacapurpura/arnesia/internal/adapters/transport/sse"
	"github.com/alpacapurpura/arnesia/internal/adapters/watch"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch cmd := os.Args[1]; cmd {
	case "serve":
		err = runServe(os.Args[2:])
	case "open":
		err = runOpen(os.Args[2:])
	case "index":
		err = runIndex(os.Args[2:])
	case "init":
		err = runInit(os.Args[2:])
	case "publish":
		err = runPublish(os.Args[2:])
	case "conformance":
		err = runConformance(os.Args[2:])
	case "portafolio":
		err = runPortafolio(os.Args[2:])
	case "sesiones":
		err = cmdSesiones(os.Args[2:])
	case "telemetria":
		err = runTelemetria(os.Args[2:])
	case "hook":
		// `arnesia hook proceso` es la instrumentación, no una vía de inspección: su
		// contrato (stdout vacío, exit 0 SIEMPRE) es incompatible con un subcomando que
		// imprime, y por eso queda fuera de la familia `telemetria`.
		err = runHook(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "arnesia: unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}

	if err != nil {
		slog.Error("arnesia", "cmd", os.Args[1], "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `arnesia — la fábrica de arneses (daemon)

usage: arnesia <command> [flags]

commands:
  serve     watcher + index + HTTP/SSE API + UI embebida on :4200 (dev sin dist: solo API; el bundle la trae)
  open      open the UI (stub)
  index     load an arnés directory into a graph.l0 (nomenclatura-arnes.md)
  init      siembra el process-as-code .arnesia/ en un proyecto (contrato semilla-arnesia.md); --check = doctor
  publish   publica el canónico de una entrada del Portafolio en su marketplace-home (B2)
  conformance  run the ruleset against an element or an arnés (METODOLOGIA §6)
  portafolio   escanear/listar/agregar/desvincular arneses del Portafolio (Slice 0)
  sesiones     recalibrar-llaves: lleva la llave de arnés de cada sesión a su clave calificada (dry-run por default)
  telemetria   resumen/mejoras/salud/purgar/catalogo — verificación del módulo sin FE
  hook         hook de instrumentación: 'arnesia hook proceso' (stdin -> loopback, exit 0 siempre)
`)
}

// runServe wires the adapters + use cases + HTTP/SSE transport and listens on :4200.
// It is the only real subcommand in this skeleton.
func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:4200", "listen address")
	claudeBin := fs.String("claude", "claude", "path to the claude binary (the Dock conductor)")
	sessionsPath := fs.String("sessions", "", "session registry file (default ~/.arnesia/sessions.json)")
	arnesesPath := fs.String("arneses", "", "arnés→path registry file (default ~/.arnesia/arneses.json)")
	arnesRoot := fs.String("arnes-root", "", "root for unregistered-arnés fallback dirs (default ~/.arnesia/arneses)")
	indexPath := fs.String("index", "", "índice SQLite del Mapa (default ~/.arnesia/index.db)")
	maxTurns := fs.Int("max-turns", 40, "cap on the agent loop per turn (--max-turns); 0 disables the cap")
	repairCap := fs.Int("repair-cap", 3, "iteraciones máximas de reparación de una caja T3 (BoxConductor)")
	rotUmbral := fs.Int("rotacion-umbral", 40, "% de contexto que dispara la rotación invisible de la sesión (RF-195); 0 la apaga")
	authToken := fs.String("auth-token", os.Getenv("ARNESIA_AUTH_TOKEN"),
		"capability token required on the API (default $ARNESIA_AUTH_TOKEN; empty = Host+Origin only, dev)")
	repo := fs.String("repo", os.Getenv("ARNESIA_REPO"),
		"ruta del repo para el self-update (default $ARNESIA_REPO; vacío = botón Actualizar deshabilitado)")
	logPath := fs.String("log", os.Getenv("ARNESIA_LOG"),
		"archivo de log del daemon (default ~/.arnesia/logs/arnesia.log; '-' = solo stderr)")
	// ── Telemetría embebida (paquete 2026-07-24-telemetria-embebida-otel) ──
	// ⚠️ El 90 de la retención es un valor **PROPUESTO, no firmado** (J-6 · parada P2 del
	// plan): firmado en 90 días por D26.3. Sigue siendo configurable — firmar el default no
	// es clavarlo — y la UI lo lee de `GET /api/telemetria/salud`, nunca de una constante.
	telRetencion := fs.Int("telemetria-retencion", usecase.RetencionDefaultDias,
		"días de retención del detalle de telemetría (firmado: 90)")
	telForward := fs.String("telemetria-forward", os.Getenv("ARNESIA_TELEMETRIA_FORWARD"),
		"endpoint externo al que reenviar la telemetría YA PROYECTADA (vacío = apagado, que es el default)")
	telRefresco := fs.Bool("telemetria-catalogo-refresco", false,
		"refrescar el catálogo de precios por red (A11: apagado por default — es egreso del daemon)")
	telEstricta := fs.Bool("telemetria-ingesta-token-obligatorio", false,
		"exigir token también en /v1/* (A22; consecuencia honesta: s2-instrumentado deja de reportar)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Log a archivo (RF-230). Va PRIMERO: todo lo que sigue puede fallar, y hasta acá el
	// daemon solo escribía a stderr — que en la app instalada, lanzada desde el `.desktop`
	// del `.deb`, no lo lee nadie. Sin esto un incidente no deja rastro.
	if cerrar := activarLog(*logPath); cerrar != nil {
		defer cerrar()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Arnés→path registry: resolves each session's cwd so a conductor is confined to its
	// arnés's tree, never a shared global cwd (boundary permisos-gui `sesion-aislada-por-cwd`).
	// Se construye ANTES del índice: Rebuild (abajo) reconstruye desde ella (RF-207).
	arnesReg, err := store.NewArnesRegistry(*arnesesPath, *arnesRoot)
	if err != nil {
		return fmt.Errorf("arnes registry: %w", err)
	}

	// Outbound adapters (concrete — wired only here, the composition root).
	idx, err := index.New(*indexPath, arnesReg, loader.LoadArnes)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}
	defer func() { _ = idx.Close() }()
	// Rebuild reconstruye TODO el índice desde arnesReg (RF-207) — reemplaza el loop
	// que antes vivía acá suelto; el directorio de cada arnés registrado es la verdad,
	// el índice es desechable.
	if err = idx.Rebuild(ctx); err != nil {
		return fmt.Errorf("index rebuild: %w", err)
	}
	// loadArnesDir: la vía viva del loader real (HS-11) — un directorio registrado se
	// reconoce archivo-por-archivo (nomenclatura v1) y entra al índice. Falla honesto
	// (el registro queda; el índice no inventa). Usada por PUT /api/arneses/{id} (el
	// botón "Cargar" del FE), no por el boot (eso ya lo hace Rebuild arriba).
	loadArnesDir := func(id, path string) error {
		g, lerr := loader.LoadArnes(path)
		if lerr != nil {
			return fmt.Errorf("cargar arnés %s desde %s: %w", id, path, lerr)
		}
		// `id` es la llave que el caller ya eligió (PUT /api/arneses/{id}) — se indexa
		// bajo ESA, jamás re-derivada del propio g.Arnes.ID del grafo cargado (deuda
		// BACKLOG «re-key (home,id,scope)», 2026-07-23).
		return idx.Upsert(ctx, id, g)
	}
	agent := claudecode.New(resolveClaudeBin(*claudeBin)) // the Dock conductor.

	// El registro vivo se ABRE, no se construye: `AbrirRegistro` resuelve qué archivo
	// manda, migra la forma vieja con copia previa, recalibra las llaves a medias (CV-D16)
	// y repara la invariante de conversaciones. Nada de eso ocurre en silencio — todo lo
	// que hizo viaja en el Informe y sale por el log de arranque, con las rutas.
	//
	// El re-key NO corre acá: necesita el Portafolio, que se cablea más abajo. Se pasa nil
	// y la recalibración ocurre en su propio paso, apenas el Portafolio existe (CV-D18 ·
	// buscá `RecalibrarAlArrancar` en este archivo). Este `nil` ya NO significa «el
	// operador lo corre a mano»: eso fue una enmienda que un documento le hizo a una
	// decisión firmada, y el hallazgo A-10 la revirtió.
	sesionesV2, sessionsLegado := rutasDelRegistro(*sessionsPath)
	sessionStore, informe, err := store.AbrirRegistro(sesionesV2, sessionsLegado, selfupdate.Build, nil)
	if err != nil {
		return fmt.Errorf("session store: %w", err)
	}
	loguearInforme(informe, sesionesV2)
	avisarDelArchivoAbandonado(sesionesV2)

	// Watcher fsnotify (RF-210): observa el mismo árbol que Rebuild leyó — arnesReg —
	// para disparar reindex incremental cuando algo cambia en caliente.
	watcher := watch.New(arnesReg)
	events, err := watcher.Watch(ctx)
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	// Transport + services.
	broker := sse.NewBroker()
	mapSvc := usecase.NewMapService(idx)
	// Inyección de doctrina (HS-11, puente 2): materializa kit+knowhow embebidos a
	// ~/.arnesia (idempotente por huella) e inyecta --plugin-dir/--append-system-prompt-
	// file/--add-dir a cada spawn. El usuario no configura nada.
	injector, err := provision.New("", doctrina.Kit, doctrina.Files)
	if err != nil {
		return fmt.Errorf("provisioner: %w", err)
	}

	// Conformance del daemon (HS-11, puente 3): SIEMPRE el ruleset/schemas embebidos —
	// el binario instalado se comporta igual que el de dev; el scope `fabrica`
	// (arch-test/go-arch-lint, repoRoot="") difiere honesto, el scope `arnes` (schema +
	// spine + escritor único + firewall) da pass/fail real vía RunGraph.
	schemaFS, err := iofs.Sub(doctrina.Files, "docs/architecture/contracts/schema")
	if err != nil {
		return fmt.Errorf("schemas embebidos: %w", err)
	}
	// El SchemaSet embebido se comparte con el Portafolio (B1: Identificar valida el sello
	// contra graph.l0 antes de escribirlo) — una sola instancia, un solo caché de schemas.
	schemaSet := mechanism.NewSchemaSetFS(schemaFS)
	confSvc := usecase.NewConformanceService("", ruleset.NewFromFS(doctrina.Files),
		schemaSet, []ports.MechanismAdapter{
			mechanism.NLJudge{}, mechanism.StaticScan{}, mechanism.SchemaAdapter{},
		})

	// Permisos por rol (Fase E): el KitProvisioner resuelve el permission-set del rol
	// que hidrata (permisos-derivan-del-rol); las sesiones lo usan al responder un
	// control_request y el conductor T3 lo materializa en flags CC-native al spawn.
	perms := permission.NewKitProvisioner()

	// El rol de una sesión ES el del arnés que hidrata (graph.l0 META, decisión #6 del
	// paquete chat-cc-funcional): el índice lo conoce; el FE jamás elige autoridad.
	roleFor := func(rctx context.Context, arnesID string) string {
		g, gerr := mapSvc.Graph(rctx, arnesID)
		if gerr != nil || g.Arnes == nil {
			return ""
		}
		return g.Arnes.Rol
	}

	sessionSvc, err := usecase.NewSessionService(ctx, agent, sessionStore, brokerPublisher{broker}, arnesReg, *maxTurns, injector, perms, roleFor)
	if err != nil {
		return fmt.Errorf("session service: %w", err)
	}
	// CH-D6: el kit materializado es paquete CERRADO — el gate deniega por ruta cualquier
	// tool_use que lo toque, sin tarjeta.
	sessionSvc.ProtegerPaqueteCerrado(injector.BaseDir())

	// Fuente del nodo (RF-93, tab Contenido del inspector): lectura confinada al dir
	// registrado del arnés (S2) — el drawer muestra el archivo REAL, jamás reconstruye.
	fuenteSvc := usecase.NewFuenteService(idx, arnesReg, artifact.NewFuenteReader())
	// baseFor: el dir del arnés registrado resuelve los fuente_path relativos del
	// firewall; un arnés no registrado (fixtures embebidos) usa el repo si existe.
	confBase := func(id string) string {
		for _, ap := range arnesReg.List() {
			if ap.Arnes == id {
				return ap.Path
			}
		}
		if r, rerr := findRepoRoot(); rerr == nil {
			return r
		}
		return ""
	}

	// Conductor T3 (Fase E): el loop determinista de una caja, con el lector de
	// `status:` del artefacto (document-as-cache) confinado al árbol del arnés. El gate
	// post-run (deuda BACKLOG «run async», 2026-07-23) reusa confSvc/confBase — mismo
	// ConformancePort que GET /api/harnesses/{id}/conformance.
	conductor := usecase.NewBoxConductor(agent, artifact.NewReader(), *repairCap, *maxTurns)
	runSvc := usecase.NewRunService(idx, conductor, perms, arnesReg, injector, brokerPublisher{broker}, confSvc, confBase)

	// Self-update sin sudo (paquete boton-actualizar): el repo llega por flag/env —
	// JAMÁS del request (RF-106); sin repo la tarjeta lo dice y el botón queda disabled.
	// Bugfix fix-repo-self-update (RF-108): si no vino por flag/env, se intenta el
	// último configurado vía UI (persistido en ~/.arnesia/self-update.json) — así una
	// instalación empaquetada (Tauri/.deb, sin --repo en su sidecar) puede actualizar
	// tras configurarlo UNA vez desde Ajustes, sin volver a tocar CLI.
	repoStore, err := selfupdate.NewRepoStore("")
	if err != nil {
		return fmt.Errorf("self-update: repo store: %w", err)
	}
	repoInicial := *repo
	if repoInicial == "" {
		if persistido, perr := repoStore.Leer(); perr == nil {
			repoInicial = persistido
		} else {
			slog.Warn("self-update: repo store ilegible al boot — arranca sin repo", "err", perr)
		}
	}
	updater, err := selfupdate.New(repoInicial)
	if err != nil {
		return fmt.Errorf("self-update: %w", err)
	}
	updSvc := usecase.NewSelfUpdateService(updater, repoStore)

	// Portafolio de arneses (Slice 0, HS-22): store separado de arneses.json (A1) + walker
	// físico READ-ONLY + evaluador de deriva local + wrapper del loader real. Mismo wiring
	// que usa el subcomando `portafolio` — newPortafolioService lo factoriza para no
	// duplicarlo.
	portafolioSvc, pfStore, derivaEval, err := newPortafolioService(idx, schemaSet)
	if err != nil {
		return fmt.Errorf("portafolio service: %w", err)
	}
	avisarCoberturaPortafolioIndice(ctx, portafolioSvc, idx)
	// Plano Marketplaces + catálogo + `↧ Traer canónico` (paquete
	// 2026-07-23-portafolio-agregar-marketplace): comparte el store del Portafolio (el cruce
	// catálogo × portafolio lo necesita) y la MISMA instancia del evaluador de deriva.
	marketplaceSvc, err := newMarketplaceService(pfStore, derivaEval)
	if err != nil {
		return fmt.Errorf("marketplace service: %w", err)
	}
	// `▲ Publicar` (B2, paquete 2026-07-30-volverlo-de-arnesia-y-publicar): el publisher git +
	// el MISMO loader real y el MISMO motor de conformance que sirve GET …/conformance — el
	// gate del publish y el del Mapa son un solo veredicto. Credencial del operador (gh/PAT),
	// jamás de la app (BR-18).
	marketplaceSvc.SetPublicar(
		&publish.Publisher{GitBin: "git", GHBin: "gh", Token: os.Getenv("ARNESIA_GH_TOKEN")},
		arnesLoaderFunc(loader.LoadArnes), confSvc)

	// Forja: la siembra `.arnesia/` (paquete 2026-07-30-arnesia-en-el-proyecto, A-T3) —
	// mismo usecase que `arnesia init`, expuesto por HTTP para el FE futuro.
	semillaFS, err := iofs.Sub(doctrina.Semilla, "semilla")
	if err != nil {
		return fmt.Errorf("semilla embebida: %w", err)
	}
	forjaSvc := usecase.NewForjaService(forja.New(semillaFS))

	// CV-D18 (FIRMADA 2026-07-27) · el re-key de CV-D16 corre SOLO, acá, en el arranque.
	//
	// Va en este punto y no junto a `AbrirRegistro` por una razón dura: el resolvedor de
	// llaves necesita el Portafolio, que recién existe en esta línea. Antes se pasaba `nil`
	// a `AbrirRegistro` y `reKey` no corría nunca — CV-D16 estaba firmada y NO construida
	// (hallazgo A-10 de la auditoría, síntoma N-23: 3 de las 5 sesiones del operador
	// invisibles, sin que nada se lo avisara).
	//
	// Esto MUTA datos del operador al arrancar, que es justo lo que la arquitectura del
	// paquete quiso evitar, y el riesgo se acepta a ojos abiertos porque el operador lo
	// eligió sabiendo el costo. Lo que lo hace aceptable no es el argumento: son las tres
	// garantías que `RecalibrarAlArrancar` sostiene con tests —respaldo ANTES del cambio
	// (si falla el respaldo no se recalibra nada), informe que distingue recalibradas de
	// `sin-candidata`, e idempotencia— más la reversión, que sigue viva en
	// `arnesia sesiones recalibrar-llaves --revertir --desde <respaldo>`.
	//
	// Un fallo acá NO tumba el daemon: se dice y se sigue. Quedarse sin arrancar por no
	// poder recalibrar una llave sería peor que la llave a medias que se venía tolerando.
	if entradas, _, lerr := portafolioSvc.Listar(ctx); lerr != nil {
		slog.Warn("registro de sesiones: no se pudo leer el Portafolio para recalibrar las llaves — "+
			"quedan como están; corré `arnesia sesiones recalibrar-llaves --dry-run`", "err", lerr)
	} else if infRekey, rerr := sessionStore.RecalibrarAlArrancar(resolverDeLlaves(entradas), selfupdate.Build); rerr != nil {
		slog.Error("registro de sesiones: la recalibración de llaves NO se aplicó — el registro quedó intacto",
			"registro", sesionesV2, "err", rerr)
	} else {
		loguearRecalibracion(infRekey, sesionesV2)
		// El store ya escribió el disco, pero `sessionSvc` cargó el registro más arriba y
		// tiene las llaves viejas en memoria. Sin este puente el arranque que APLICA el
		// re-key sigue sirviendo lo de antes, y el operador estrena la función viendo el
		// bug que la función arregla (N-24, medido contra el binario).
		nuevas := map[string]string{}
		for _, f := range infRekey.Filas {
			if f.Movio() {
				nuevas[f.SesionID] = f.Despues
			}
		}
		if n := sessionSvc.AplicarRecalibracion(nuevas); n > 0 {
			slog.Info("registro de sesiones: llaves recalibradas aplicadas al registro vivo", "sesiones", n)
		}
	}
	// Tarjeta de identidad por sesión (RF-189): cada spawn sabe qué arnés es, qué copia
	// edita (canónico/instalación/suelto) y su rol — cerrada sobre el Portafolio real.
	sessionSvc.SetGrounding(func(gctx context.Context, arnesID, cwd string) string {
		entradas, _, lerr := portafolioSvc.Listar(gctx)
		if lerr != nil {
			entradas = nil
		}
		return usecase.TarjetaIdentidad(entradas, arnesID, cwd, roleFor(gctx, arnesID))
	})
	// Rotación de contexto invisible (RF-195): umbral configurable, default 40 %.
	sessionSvc.SetUmbralRotacion(*rotUmbral)
	// Historial B2 (RF-200/201): Close archiva metadata (no borra el rastro) y el lector
	// JSONL nativo reconstruye conversaciones cerradas. Fallos degradan honesto con warn.
	if cerradasStore, cerr := store.NewRegistry(cerradasPathDefault(*sessionsPath)); cerr != nil {
		slog.Warn("historial: registro de cerradas no disponible", "err", cerr)
	} else {
		sessionSvc.SetArchivoCerradas(cerradasStore)
	}
	if hreader, herr := history.New(""); herr != nil {
		slog.Warn("historial: lector JSONL no disponible", "err", herr)
	} else {
		sessionSvc.SetHistoryReader(hreader)
	}
	// Reindex-tras-turno (RF-184) + deriva honesta (RF-193): el Mapa refleja lo que el
	// chat edita y el Portafolio nunca finge `al-hilo` tras una edición.
	turnReindex := usecase.NewTurnReindexer(idx, loader.LoadArnes, brokerPublisher{broker})
	sessionSvc.SetReindexer(func(rctx context.Context, arnesID, cwd string) {
		turnReindex(rctx, arnesID, cwd)
		if _, derr := portafolioSvc.ReevaluarDeriva(rctx, cwd); derr != nil {
			slog.Warn("deriva tras turno", "arnes", arnesID, "err", derr)
		}
	})

	// Dictado por voz (paquete 2026-07-25-spike-voz-dictado). El motor de STT es el que el
	// operador tenga instalado (V-D1: adaptador por PATH) — si no hay ninguno, el servicio
	// reporta NO-disponible con motivo y el botón queda gris con la razón a la vista, en vez
	// de aceptar un dictado que después no se podría transcribir. La limpieza reusa el mismo
	// `claude` del Dock, pero spawneado ENDURECIDO (V-D7): 1 turno, tools negadas.
	dictadoSvc := usecase.NewDictadoService(
		stt.New(),
		claudecode.NewLimpiador(resolveClaudeBin(*claudeBin)),
		sessionSvc,
	)

	// ── Telemetría embebida: composition root del módulo (§10.1) ──
	// Todo lo de acá degrada honesto: si el almacén no abre, el daemon SIGUE sirviendo la
	// API y la salud dice que la telemetría no está disponible, con motivo. Se pierde
	// telemetría, nunca la sesión del usuario.
	tokenIngesta := mintTokenIngesta()
	telSvc, telHandler, cerrarTel := cablearTelemetria(ctx, cablesTelemetria{
		retencionDias:    *telRetencion,
		forwardEndpoint:  *telForward,
		refrescoCatalogo: *telRefresco,
		idx:              mapSvc,
		portafolio:       portafolioSvc,
		roleFor:          roleFor,
	})
	if cerrarTel != nil {
		defer cerrarTel()
	}

	// El listener se crea ANTES del handler y el Host gate se deriva de `ln.Addr()`, **no del
	// flag**. Con `--addr 127.0.0.1:0` —el puerto efímero que todo E2E de telemetría tiene que
	// usar para no chocar con un daemon viejo pegado a un puerto fijo— el flag dice `:0` y el
	// puerto real es otro: derivar la allowlist del flag dejaría al daemon respondiendo 403 a
	// su propio endpoint. (Encontrado corriendo el E2E, no razonando.)
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", *addr, err)
	}
	dirReal := ln.Addr().String()

	handler := httpapi.NewHandler(mapSvc, sessionSvc, runSvc, fuenteSvc, arnesReg, confSvc, confBase, loadArnesDir, updSvc, portafolioSvc, forjaSvc, marketplaceSvc, dictadoSvc, telSvc, telHandler, embeddedUI(), broker,
		httpapi.AuthConfigConIngesta(dirReal, *authToken, tokenIngesta, *telEstricta))

	// Filesystem changes drive incremental reindex + a map delta on the SSE bus
	// (RF-210, decisiones.md D5): SOLO el arnés dueño del path que cambió se recarga —
	// nunca idx.Rebuild(ctx) completo por evento, que recargaría TODOS los arneses
	// registrados por cada archivo tocado de uno solo. Reusa turnReindex (mismo patrón
	// que session_reindex.go ya prueba en producción para el chat).
	go func() {
		for ev := range events {
			ap, ok := ownerOf(arnesReg, ev.Path)
			if !ok {
				continue // evento huérfano: no cae bajo ningún arnés registrado.
			}
			turnReindex(ctx, ap.Arnes, ap.Path)
		}
	}()

	srv := &http.Server{
		Addr:              dirReal,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Listener EXPLÍCITO (A8, variante recomendada por §10.1): la ficha del daemon se
	// publica **después** de que el socket acepta, nunca antes. Una ficha que nombra un
	// puerto donde no escucha nadie es una mentira que el hook cobra en timeouts.
	fichaDaemon, retirarFicha := publicarFicha(ctx, ln, tokenIngesta)
	if retirarFicha != nil {
		defer retirarFicha()
	}
	if telSvc != nil {
		telSvc.SetDescubrimiento(fichaDaemon)
	}

	go func() {
		<-ctx.Done()
		// WithoutCancel: the parent ctx is already done (that is why we are here); the
		// shutdown needs its own deadline, inheriting values but not the cancellation.
		shutCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown", "err", err)
		}
	}()

	slog.Info("arnesia serve", "addr", ln.Addr().String())
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// activarLog manda el log del daemon a stderr **y** a un archivo rotativo (RF-230).
//
// Los dos destinos formatean distinto a propósito: el operador que corre `arnesia serve` en
// una terminal quiere texto, y el archivo —que es el que se lee después de un incidente—
// quiere JSON, porque los diagnósticos del FE traen un `detalle` anidado que el handler de
// texto aplasta a algo ilegible.
//
// Un archivo que no se puede abrir NO tumba el daemon: se avisa y queda solo stderr. Devuelve
// el cierre, o nil si no hay archivo que cerrar.
func activarLog(path string) func() {
	if path == "-" {
		return nil
	}
	w, err := logfile.Abrir(path, 0)
	if err != nil {
		slog.Warn("log: no pude abrir el archivo — el daemon queda solo con stderr", "err", err)
		return nil
	}
	slog.SetDefault(slog.New(dosDestinos{
		a: slog.NewTextHandler(os.Stderr, nil),
		b: slog.NewJSONHandler(w, nil),
	}))
	// Se nombra el archivo al arrancar: un log que nadie sabe dónde está no se lee.
	slog.Info("log", "archivo", w.Path())
	return func() { _ = w.Close() }
}

// dosDestinos reparte cada registro entre dos handlers.
//
// `slog` no trae fan-out y un `io.MultiWriter` no sirve acá: obligaría a un solo formato para
// los dos destinos, que es justo lo que no queremos.
type dosDestinos struct{ a, b slog.Handler }

func (d dosDestinos) Enabled(ctx context.Context, l slog.Level) bool {
	return d.a.Enabled(ctx, l) || d.b.Enabled(ctx, l)
}

func (d dosDestinos) Handle(ctx context.Context, r slog.Record) error {
	if d.a.Enabled(ctx, r.Level) {
		// Clone: cada handler consume los atributos del registro por su cuenta.
		_ = d.a.Handle(ctx, r.Clone())
	}
	if d.b.Enabled(ctx, r.Level) {
		return d.b.Handle(ctx, r)
	}
	return nil
}

func (d dosDestinos) WithAttrs(as []slog.Attr) slog.Handler {
	return dosDestinos{a: d.a.WithAttrs(as), b: d.b.WithAttrs(as)}
}

func (d dosDestinos) WithGroup(n string) slog.Handler {
	return dosDestinos{a: d.a.WithGroup(n), b: d.b.WithGroup(n)}
}

// brokerPublisher adapts the SSE broker (whose Publish returns the stored event) to
// the usecase.EventPublisher interface (which ignores it), keeping usecase free of any
// transport import.
type brokerPublisher struct{ b *sse.Broker }

func (p brokerPublisher) Publish(eventType string, data []byte) { p.b.Publish(eventType, data) }

// arnesLoaderFunc adapta una función libre (loader.LoadArnes) a ports.ArnesLoader — el
// usecase del Portafolio no puede importar el paquete loader (go-arch-lint), así que cmd
// cablea la función concreta detrás del puerto.
type arnesLoaderFunc func(dir string) (domain.Graph, error)

func (f arnesLoaderFunc) Load(dir string) (domain.Graph, error) { return f(dir) }

// ownerOf resuelve qué entrada de reg es dueña de path (RF-210): la entrada cuyo Path
// ES path, o del cual path cuelga (prefijo + separador, para no confundir
// "/a/arnes-2" con "/a/arnes"). Un path que no cae bajo ningún arnés registrado
// devuelve ok=false — el caller lo ignora, nunca reindexa a ciegas.
func ownerOf(reg ports.ArnesRegistry, path string) (ports.ArnesPath, bool) {
	for _, ap := range reg.List() {
		if ap.Path == path || strings.HasPrefix(path, ap.Path+string(filepath.Separator)) {
			return ap, true
		}
	}
	return ports.ArnesPath{}, false
}

// embeddedUI returns the SPA handler when this build carries web/dist (scripts/
// bundle.sh la compila antes del daemon), or nil for an honest dev build without UI.
// Fallback SPA: cualquier ruta sin archivo sirve index.html (la app navega por
// hash-state, sin router — pero los deep-links no deben 404).
func embeddedUI() http.Handler {
	ui, err := iofs.Sub(doctrina.WebDist, "web/dist")
	if err != nil {
		return nil
	}
	if _, err := iofs.Stat(ui, "index.html"); err != nil {
		return nil // build de dev: el dist embebido solo trae .gitkeep.
	}
	files := http.FileServerFS(ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" && p != "index.html" {
			if _, err := iofs.Stat(ui, p); err == nil {
				files.ServeHTTP(w, r)
				return
			}
		}
		// index.html (raíz, explícito o fallback SPA) va sin caché: tras un self-update
		// el WebView del shell y el browser deben recoger los assets nuevos (hasheados
		// por Vite) en la próxima carga, no cuando el heurístico de caché quiera.
		w.Header().Set("Cache-Control", "no-cache")
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		files.ServeHTTP(w, r2)
	})
}

// resolveClaudeBin hardens `claude` discovery for GUI launches: una app de escritorio
// (lanzada desde .desktop/acceso directo, no desde una shell) frecuentemente NO lleva
// ~/.local/bin en su PATH — el binario existe pero LookPath no lo ve y el spawn moriría
// silencioso hacia el Dock (hallazgo de la auditoría 2026-07-07). Un nombre pelado que
// PATH no resuelve se sondea en las rutas de instalación conocidas del OS; el error del
// spawn sigue siendo la autoridad final. En Windows, LookPath ya resuelve PATHEXT
// (claude.cmd/claude.exe) sobre el PATH heredado; los candidatos cubren instalaciones
// fuera de PATH (instalador nativo en ~/.local/bin, npm global en %APPDATA%\npm).
func resolveClaudeBin(bin string) string {
	if strings.ContainsRune(bin, os.PathSeparator) {
		return bin // ruta explícita del operador: se respeta tal cual.
	}
	if _, err := exec.LookPath(bin); err == nil {
		return bin
	}
	home, _ := os.UserHomeDir()
	dirs := []string{
		filepath.Join(home, ".local", "bin"), // instalador nativo de Claude Code (también en Windows)
		filepath.Join(home, ".claude", "local"),
	}
	sufijos := []string{""}
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			dirs = append(dirs, filepath.Join(appdata, "npm")) // shim de `npm install -g`
		}
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			dirs = append(dirs, filepath.Join(local, "Programs", "claude"))
		}
		sufijos = []string{".exe", ".cmd", ""}
	} else {
		dirs = append(dirs, "/usr/local/bin", "/opt/homebrew/bin")
	}
	for _, dir := range dirs {
		for _, suf := range sufijos {
			p := filepath.Join(dir, bin+suf)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				slog.Info("claude resuelto fuera de PATH (entorno GUI)", "path", p)
				return p
			}
		}
	}
	slog.Warn("claude no está en PATH ni en rutas conocidas — instala Claude Code o pasa --claude", "bin", bin)
	return bin
}

// runOpen opens the UI. Stub.
func runOpen(_ []string) error {
	fmt.Println("arnesia open: TODO(fase 5) — open the embedded UI / Tauri shell")
	return nil
}

// runIndex loads an arnés DIRECTORY into its graph.l0 via the nomenclatura loader
// (HS-11, puente 1: archivo-por-archivo → grafo, nomenclatura-arnes.md v1) and emits
// the graph. Sin argumento, es un smoke test: abre un índice SQLite descartable en un
// dir temporal y confirma que los grafos demo/dogfood embebidos decodifican — NO llama
// Rebuild (eso ahora reconstruye desde ArnesRegistry, RF-207; este subcomando no tiene
// uno que ofrecer, así que Rebuild solo vaciaría el índice de vuelta).
func runIndex(args []string) error {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	out := fs.String("o", "", "write the graph.l0 JSON to this file (default: stdout)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia index [-o out.graph.json] [<dir-del-arnés>]   (flags ANTES del dir — semántica flag de Go)

  <dir>   raíz de un arnés (plugin CC con .claude-plugin/, o proyecto con .claude/):
          se reconoce archivo-por-archivo según docs/architecture/contracts/nomenclatura-arnes.md
          y se emite su graph.l0 (fuente_path ESTAMPADOS; no-reconocido VISIBLE).
  sin dir: smoke test — confirma que los grafos demo/dogfood embebidos decodifican
           contra un índice SQLite descartable (no persiste, no toca ~/.arnesia).
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := fs.Arg(0)
	if dir == "" {
		tmp, terr := os.MkdirTemp("", "arnesia-index-smoke-")
		if terr != nil {
			return terr
		}
		defer func() { _ = os.RemoveAll(tmp) }()
		idx, ierr := index.New(filepath.Join(tmp, "index.db"), nil, nil)
		if ierr != nil {
			return ierr
		}
		defer func() { _ = idx.Close() }()
		fmt.Println("arnesia index: seed demo/dogfood decodifica OK (SQLite descartable, no persiste)")
		return nil
	}

	g, err := loader.LoadArnes(dir)
	if err != nil {
		return fmt.Errorf("index %s: %w", dir, err)
	}
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if *out != "" {
		if werr := os.WriteFile(*out, b, 0o600); werr != nil {
			return werr
		}
		fmt.Fprintf(os.Stderr, "arnesia index: %s → %s (%d nodos, %d edges)\n", dir, *out, len(g.Nodes), len(g.Edges))
		return nil
	}
	_, err = os.Stdout.Write(b)
	return err
}

// runPublish publica el canónico de una entrada del Portafolio en su marketplace-home (B2) —
// el MISMO usecase que POST /api/portafolio/arneses/{clave}/publicaciones, cero lógica propia
// acá (patrón runPortafolio). El gate de conformance corre con el ruleset embebido: el binario
// instalado publica con el mismo contrato que el daemon.
func runPublish(args []string) error {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia publish <clave>

  <clave>  la clave del arnés en el Portafolio (arnesia portafolio listar)

Publica el CANÓNICO en el marketplace de su home (clase propio): gate de conformance
verde → copia a plugins/<id>/<version>/ → marketplace.json + catalogo.json → commit →
push sin force → tag <id>/vX.Y.Z. La versión es la de plugin.json del canónico.
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	clave := fs.Arg(0)
	if clave == "" {
		fs.Usage()
		return errors.New("falta la clave del arnés")
	}

	schemas, err := schemasEmbebidos()
	if err != nil {
		return err
	}
	_, pfStore, derivaEval, err := newPortafolioService(nil, schemas)
	if err != nil {
		return err
	}
	svc, err := newMarketplaceService(pfStore, derivaEval)
	if err != nil {
		return err
	}
	confSvc := usecase.NewConformanceService("", ruleset.NewFromFS(doctrina.Files),
		schemas, []ports.MechanismAdapter{
			mechanism.NLJudge{}, mechanism.StaticScan{}, mechanism.SchemaAdapter{},
		})
	svc.SetPublicar(
		&publish.Publisher{GitBin: "git", GHBin: "gh", Token: os.Getenv("ARNESIA_GH_TOKEN")},
		arnesLoaderFunc(loader.LoadArnes), confSvc)

	res, err := svc.Publicar(context.Background(), clave)
	if err != nil {
		return err
	}
	fmt.Printf("publicado %s v%s → %s (commit %s)\n", res.ID, res.Version, res.Marketplace, res.Commit)
	if res.Tag != "" {
		fmt.Printf("tag %s\n", res.Tag)
	}
	for _, a := range res.Avisos {
		fmt.Printf("aviso: %s\n", a)
	}
	return nil
}

// newPortafolioService cablea el usecase del Portafolio con sus adapters por default
// (store en ~/.arnesia/portafolio.json, Scanner/Referencias con CCPluginsDir/MaxDepth
// default) — compartido entre `serve` y el subcomando `portafolio`, sin duplicar wiring.
// indice es el 5° puerto (S1-D1, Observar en Mapa): `serve` pasa el `idx` real que ya
// construyó; el subcomando CLI pasa nil — no necesita indexar (ObservarEnMapa con índice
// nil da el error honesto «requiere el daemon», jamás un nil-pointer panic). schemas es el
// 6° puerto (B1): `serve` pasa el SchemaSet embebido que ya construyó; los subcomandos CLI
// lo arman con schemasEmbebidos().
func newPortafolioService(indice ports.IndexPort, schemas ports.SchemaValidator) (*usecase.PortafolioService, ports.PortafolioStore, ports.DerivaEvaluator, error) {
	st, err := portafolio.NewStore("")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("portafolio store: %w", err)
	}
	// El evaluador de deriva se devuelve para REUSAR LA MISMA INSTANCIA en el servicio de
	// marketplaces (§12.2 riesgo 15): una segunda duplicaría la resolución de RutaReferencia y
	// podrían divergir.
	refs := &portafolio.Referencias{}
	return usecase.NewPortafolioService(st, &portafolio.Scanner{}, arnesLoaderFunc(loader.LoadArnes), refs, indice, schemas), st, refs, nil
}

// schemasEmbebidos construye el SchemaSet desde la doctrina embebida en el binario (mismo
// patrón que conformance.go): un subcomando CLI valida el sello contra el MISMO contrato
// que el daemon instalado, sin depender del repo en disco.
func schemasEmbebidos() (*mechanism.SchemaSet, error) {
	sub, err := iofs.Sub(doctrina.Files, "docs/architecture/contracts/schema")
	if err != nil {
		return nil, fmt.Errorf("schemas embebidos: %w", err)
	}
	return mechanism.NewSchemaSetFS(sub), nil
}

// newMarketplaceService cablea el plano Marketplaces (paquete
// 2026-07-23-portafolio-agregar-marketplace, design.md §3.9) a sus 7 puertos de lectura + los 3 de
// `↧ Traer canónico`. `remoto` se cablea SIEMPRE (el lector decide en runtime si tiene gh/PAT y
// degrada honesto) — un nil acá convertiría «no puedo leer» en «no existe la función», que es peor.
// El store del Portafolio y el evaluador de deriva se COMPARTEN con `portafolioSvc`.
func newMarketplaceService(pfStore ports.PortafolioStore, deriva ports.DerivaEvaluator) (*usecase.MarketplaceService, error) {
	st, err := marketplace.NewStore("")
	if err != nil {
		return nil, fmt.Errorf("marketplace store: %w", err)
	}
	cache, err := marketplace.NewCache("")
	if err != nil {
		return nil, fmt.Errorf("marketplace cache: %w", err)
	}
	token := os.Getenv("ARNESIA_GH_TOKEN")
	remoto := &marketplace.LectorRemoto{Token: token}
	svc := usecase.NewMarketplaceService(
		st, &marketplace.DetectorCC{}, &marketplace.LectorLocal{}, remoto, remoto, cache, pfStore,
	)
	// `↧ Traer canónico`: camino A sin red, camino B por `git` con `gh` como credential helper
	// (ArnesIA nunca ve el token, BR-18). raizArnesia "" ⇒ ~/.arnesia.
	svc.SetTraer(&traer.CopiadorLocal{}, &traer.ClonadorExterno{GHBin: "gh", Token: token}, deriva, "")
	// «↻ Refrescar» = pull ff-only del checkout de un marketplace PROPIO (DD-2/E-bis);
	// la política (refresco explícito + clase propio) vive en el usecase.
	svc.SetSync(&marketplace.SincronizadorGit{})
	return svc, nil
}

// runPortafolio es la vía de verificación E2E del Portafolio sin FE (S0-D9): reusa el
// MISMO usecase que el HTTP — cero lógica propia acá.
func runPortafolio(args []string) error {
	fs := flag.NewFlagSet("portafolio", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia portafolio <subcomando> [args]

  escanear <dir>            escanea dir (abs o relativo a cwd) y emite los candidatos crudos (JSON); NO persiste.
  listar                    emite las entradas del registro + las corruptas visibles (JSON).
  agregar <dir> <clave>...  re-escanea dir y persiste SOLO los candidatos con esas claves.
  desvincular <clave>       quita la entrada del registro; no borra nada de disco.
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return errors.New("arnesia portafolio: falta el subcomando")
	}

	// nil: el CLI no observa en Mapa (esa vía es HTTP-only, S1-D1) — no necesita el 5° puerto.
	// Los schemas embebidos sí van (B1): un `Identificar` futuro por CLI valida igual que el daemon.
	schemas, err := schemasEmbebidos()
	if err != nil {
		return err
	}
	svc, _, _, err := newPortafolioService(nil, schemas)
	if err != nil {
		return err
	}
	ctx := context.Background()

	switch sub, rest := fs.Arg(0), fs.Args()[1:]; sub {
	case "escanear":
		if len(rest) < 1 {
			return errors.New("uso: arnesia portafolio escanear <dir>")
		}
		dir, aerr := filepath.Abs(rest[0])
		if aerr != nil {
			return aerr
		}
		cands, serr := svc.Escanear(ctx, dir)
		if serr != nil {
			return serr
		}
		return imprimirJSON(cands)

	case "listar":
		sanas, corruptas, lerr := svc.Listar(ctx)
		if lerr != nil {
			return lerr
		}
		return imprimirJSON(struct {
			Entradas  []domain.EntradaPortafolio `json:"entradas"`
			Corruptas []domain.EntradaCorrupta   `json:"corruptas,omitempty"`
		}{sanas, corruptas})

	case "agregar":
		if len(rest) < 2 {
			return errors.New("uso: arnesia portafolio agregar <dir> <clave> [<clave>...]")
		}
		dir, aerr := filepath.Abs(rest[0])
		if aerr != nil {
			return aerr
		}
		persistidas, perr := svc.AgregarProyecto(ctx, dir, rest[1:])
		if perr != nil {
			return perr
		}
		return imprimirJSON(persistidas)

	case "desvincular":
		if len(rest) < 1 {
			return errors.New("uso: arnesia portafolio desvincular <clave>")
		}
		ok, derr := svc.Desvincular(ctx, rest[0])
		if derr != nil {
			return derr
		}
		return imprimirJSON(map[string]bool{"desvinculado": ok})

	default:
		fs.Usage()
		return fmt.Errorf("arnesia portafolio: subcomando desconocido %q", sub)
	}
}

// imprimirJSON emite v indentado a stdout — la forma común de las 4 salidas de
// `arnesia portafolio`.
func imprimirJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = os.Stdout.Write(b)
	return err
}

// rutasDelRegistro resuelve el par (registro vigente, registro de la versión anterior).
// El de la versión anterior NO se toca, NO se borra y NO se renombra: mientras exista
// intacto, volver a un binario anterior no necesita restaurar nada.
//
// Con `--sessions` explícito, esa ruta ES el registro vigente y no hay legado que migrar:
// el operador está apuntando a un archivo suyo a propósito (tests, copias, otro perfil).
func rutasDelRegistro(sessionsPath string) (vigente, legado string) {
	if sessionsPath != "" {
		return sessionsPath, ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "" // AbrirRegistro resolverá (y fallará) honesto por su cuenta.
	}
	dir := filepath.Join(home, ".arnesia")
	return filepath.Join(dir, "sesiones.json"), filepath.Join(dir, "sessions.json")
}

// loguearInforme cuenta lo que el arranque le hizo al registro. Un arranque sin novedades
// no imprime nada: el Informe sólo trae lo que efectivamente ocurrió.
// avisarDelArchivoAbandonado — A-9. `sesiones-cerradas.json` dejó de leerse cuando el
// archivado pasó a conservar el transcript completo (CV-D12 + CV-D8), y el operador lo borra
// a mano (CV-D6). Todo bien con eso; lo que NO estaba bien era el SILENCIO.
//
// Este paquete tiene un lema —«nada de esto ocurre en silencio»— y lo cumple para la
// migración, el respaldo, la cuarentena, el esquema futuro, las reparaciones y las
// recalibraciones. Abandonar un archivo de DATOS del operador era lo único que no decía
// nada. Y no aplica sólo a las 3 sesiones que él ya declaró no importantes: CUALQUIERA que
// haya archivado con el binario anterior, antes de actualizar, cayó ahí.
//
// Una línea, en el arranque, con la ruta y el número. No lo lee, no lo migra y no lo borra:
// sólo deja de fingir que no existe.
// avisarCoberturaPortafolioIndice cuenta cuántas entradas SANAS del Portafolio nunca pasaron
// por "Observar en Mapa" (Fase 2, D4: la ley anti-drift S1-D1 — "el scan del Portafolio NUNCA
// registra en arneses.json" — es correcta en intención, pero hasta acá no había señal de
// CUÁNTOS proyectos quedan así, indefinidamente). Silencioso si la cobertura es completa —
// mismo criterio que loguearRecalibracion: un arranque sin novedades no habla.
func avisarCoberturaPortafolioIndice(ctx context.Context, svc *usecase.PortafolioService, idx ports.IndexPort) {
	sanas, _, err := svc.Listar(ctx)
	if err != nil || len(sanas) == 0 {
		return // sin Portafolio o ilegible: nada que reportar acá (otros avisos ya lo cubren).
	}
	var sinIndice []string
	for _, e := range sanas {
		clave := e.Identidad.Clave()
		if _, qerr := idx.Query(ctx, clave); qerr != nil {
			sinIndice = append(sinIndice, clave)
		}
	}
	if len(sinIndice) == 0 {
		return
	}
	slog.Warn("portafolio: hay entradas sin observar en el Mapa — el índice no las conoce hasta que se aprieta \"Observar en Mapa\"",
		"sin_indice", len(sinIndice), "de_un_total", len(sanas), "claves", sinIndice)
}

func avisarDelArchivoAbandonado(sesionesPath string) {
	ruta := filepath.Join(filepath.Dir(sesionesPath), "sesiones-cerradas.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // ruta derivada de la del registro, no de la entrada del usuario.
	if err != nil {
		return // no existe (el caso normal) o no se puede leer: no es motivo para molestar.
	}
	var cerradas []struct {
		ID string `json:"id"`
	}
	n := -1 // -1 = está pero no se pudo contar; se dice igual, no se calla.
	if json.Unmarshal(b, &cerradas) == nil {
		n = len(cerradas)
	}
	slog.Warn("registro de sesiones: quedó un `sesiones-cerradas.json` del formato anterior — "+
		"este binario NO lo lee ni lo migra; lo que archives ahora va al registro nuevo. "+
		"Podés borrarlo a mano cuando quieras (CV-D6 · T32)",
		"archivo", ruta, "sesiones", n, "bytes", len(b))
}

// loguearRecalibracion cuenta lo que el re-key del arranque le hizo a las llaves (CV-D18).
//
// Dice las DOS cosas, y la segunda no es opcional: cuántas recalibró y cuántas quedaron
// `sin-candidata` porque su arnés no está en el Portafolio. Ese caso es real y conocido —el
// E2E lo vio en una de las cinco sesiones del operador— y callarlo lo volvería un defecto
// mudo: el operador vería una sesión que sigue sin aparecer y nada que se lo explique.
//
// Un arranque sin novedad no imprime nada, igual que `loguearInforme`.
func loguearRecalibracion(inf store.InformeRecalibracion, ruta string) {
	if !inf.Hubo() {
		return
	}
	slog.Info("registro de sesiones: llaves recalibradas al arrancar (CV-D18)",
		"recalibradas", inf.Recalibradas, "sin_candidata", inf.SinCandidata,
		"ya_calificadas", inf.YaCalificadas, "registro", ruta, "respaldo", inf.RespaldoEn)
	for _, r := range inf.Filas {
		switch {
		case r.Movio():
			slog.Info("registro de sesiones: llave recalibrada",
				"sesion", r.SesionID, "antes", r.Antes, "despues", r.Despues, "motivo", r.Motivo)
		case r.Motivo == "sin-candidata":
			slog.Warn("registro de sesiones: llave SIN recalibrar — su arnés no está en el Portafolio; "+
				"la sesión no va a aparecer al filtrar por arnés hasta que lo agregues",
				"sesion", r.SesionID, "llave", r.Antes)
		}
	}
	if inf.Recalibradas > 0 {
		slog.Info("registro de sesiones: para deshacer SÓLO el re-key, "+
			"`arnesia sesiones recalibrar-llaves --revertir --desde <respaldo>`", "respaldo", inf.RespaldoEn)
	}
}

func loguearInforme(inf store.Informe, ruta string) {
	if !inf.Hubo() {
		return
	}
	if inf.Migro {
		slog.Info("registro de sesiones: migrado a la forma nueva",
			"desde_version", inf.DesdeVersion, "hasta_version", store.EsquemaActual,
			"registro", ruta, "respaldo", inf.RespaldoEn)
	}
	if inf.Corrupto {
		slog.Error("registro de sesiones ILEGIBLE — se guardó entero y el registro arranca vacío",
			"cuarentena", inf.CuarentenaEn)
	}
	if inf.EsquemaFuturo {
		slog.Error("registro de sesiones escrito por un binario MÁS NUEVO — solo-lectura, no se toca un byte",
			"registro", ruta)
	}
	for _, arreglo := range inf.Reparaciones {
		slog.Warn("registro de sesiones: invariante reparada al cargar", "arreglo", arreglo)
	}
	for _, r := range inf.Recalibradas {
		if r.Movio() {
			slog.Info("registro de sesiones: llave recalibrada",
				"sesion", r.SesionID, "antes", r.Antes, "despues", r.Despues, "motivo", r.Motivo)
			continue
		}
		slog.Info("registro de sesiones: llave sin cambios",
			"sesion", r.SesionID, "llave", r.Antes, "motivo", r.Motivo)
	}
}

// cerradasPathDefault deriva el archivo del registro de sesiones ARCHIVADAS del de sesiones
// vivas: mismo dir, nombre propio (default ~/.arnesia/sesiones-archivadas.json).
//
// El nombre cambió con la ley (CV-D12 + CV-D8): «cerradas» describía un archivo terminal
// del que no se volvía, y ahora lo que se archiva es la sesión ENTERA con sus
// conversaciones completas. El `sesiones-cerradas.json` viejo NO se lee ni se migra — el
// operador lo borra a mano (CV-D6), porque su contenido es el que ya declaró que no le
// importa. Ese borrado es T32 y es del operador, no de este código.
func cerradasPathDefault(sessionsPath string) string {
	if sessionsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "" // NewRegistry("") resolverá (y fallará) honesto por su cuenta.
		}
		return filepath.Join(home, ".arnesia", "sesiones-archivadas.json")
	}
	return filepath.Join(filepath.Dir(sessionsPath), "sesiones-archivadas.json")
}
