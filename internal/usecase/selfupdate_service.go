package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Sentinels del self-update (RF-104/106) — el transporte los mapea a HTTP (409/503)
// sin conocer el porqué interno (dominio-independiente-de-transporte).
var (
	// ErrActualizacionEnCurso — ya hay un self-update en vuelo, o el daemon quedó a
	// segundos del re-exec (el lock se retiene tras «actualizado», decisión #9): un
	// self-update a la vez, jamás dos builds pisándose el binario.
	ErrActualizacionEnCurso = errors.New("ya hay un self-update en curso (o el daemon está por reiniciar)")
	// ErrNoActualizable — el update no puede NI empezar: binario no escribible
	// (RF-102) o repo no configurado (RF-103). El motivo viaja envuelto.
	ErrNoActualizable = errors.New("el self-update no puede correr")
)

// Resultados terminales del reporte (RF-104).
const (
	ResultadoActualizado = "actualizado"
	ResultadoYaAlDia     = "ya-al-dia"
	ResultadoFallo       = "fallo"
)

// Estados de un paso del reporte. La UI pinta la checklist SOLO con estos veredictos
// reales — jamás pasos animados inventados (nota de honestidad del spec, RF-104).
const (
	EstadoOK        = "ok"
	EstadoFallo     = "fallo"
	EstadoNoCorrido = "no-corrido"
	// EstadoAgendado — solo el paso reiniciar tras «actualizado»: la respuesta sale
	// ANTES del re-exec (RF-105), así que su veredicto real es «agendado», no «ok».
	EstadoAgendado = "agendado"
)

// Nombres canónicos de los 5 pasos (los del mockup v2 firmado, PASOS 141-147).
const (
	PasoVerificar        = "verificar"
	PasoBuild            = "build"
	PasoVerificarBinario = "verificar binario"
	PasoInstalar         = "instalar"
	PasoReiniciar        = "reiniciar"
)

// PasoReporte es el veredicto REAL de un paso (RF-104): la respuesta lleva los cinco,
// con los posteriores a un fallo en «no-corrido».
type PasoReporte struct {
	Paso    string `json:"paso"`
	Estado  string `json:"estado"`
	Detalle string `json:"detalle,omitempty"`
}

// SelfUpdateReport es la respuesta de POST /api/self-update: el desenlace terminal
// (actualizado · ya-al-dia · fallo) + el veredicto de cada paso.
type SelfUpdateReport struct {
	Resultado   string        `json:"resultado"`
	HuellaNueva string        `json:"huella_nueva,omitempty"`
	Pasos       []PasoReporte `json:"pasos"`
}

// SelfUpdateService orquesta el self-update sin sudo (RF-104): corre los pasos del
// puerto EN ORDEN, corta al primer fallo, arma el reporte honesto y garantiza un
// update a la vez (RF-106). La regla «ya al día» (huella nueva == corriente → éxito
// sin reinstalar) vive aquí — es negocio, no mecánica del adapter.
type SelfUpdateService struct {
	updater   ports.SelfUpdater
	repoStore ports.RepoConfigStore // nil = sin persistencia (dev/tests); ver ConfigurarRepo.
	mu        sync.Mutex
}

// NewSelfUpdateService returns a SelfUpdateService over the SelfUpdater port. repoStore
// puede ser nil (dev/tests sin persistencia) — ConfigurarRepo entonces fija el repo en
// caliente sin sobrevivir a un reinicio.
func NewSelfUpdateService(updater ports.SelfUpdater, repoStore ports.RepoConfigStore) *SelfUpdateService {
	return &SelfUpdateService{updater: updater, repoStore: repoStore}
}

// Version reporta la identidad del binario corriendo (RF-107) — passthrough al puerto.
func (s *SelfUpdateService) Version() ports.VersionInfo { return s.updater.Version() }

// Reiniciar re-ejecuta el daemon (paso ⑤). El transporte lo agenda POST-respuesta
// (RF-105): jamás dentro de Actualizar, que debe responder primero.
func (s *SelfUpdateService) Reiniciar() error { return s.updater.Reiniciar() }

// ConfigurarRepo valida y fija el repo activo (bugfix fix-repo-self-update, RF-109):
// delega la validación al puerto (las MISMAS reglas de Verificar) y, solo si pasa,
// persiste vía repoStore. Un path inválido no persiste ni activa nada — el error del
// puerto viaja tal cual (el transporte lo mapea a 400).
func (s *SelfUpdateService) ConfigurarRepo(ctx context.Context, path string) (string, error) {
	detalle, err := s.updater.ConfigurarRepo(ctx, path)
	if err != nil {
		return "", err
	}
	if s.repoStore != nil {
		if serr := s.repoStore.Guardar(path); serr != nil {
			return "", fmt.Errorf("repo válido pero no se pudo persistir (sobrevive esta sesión, no un reinicio): %w", serr)
		}
	}
	return detalle, nil
}

// Actualizar corre el flujo completo (RF-104). Devuelve error SOLO cuando el flujo no
// pudo NI empezar (ErrActualizacionEnCurso → 409 · ErrNoActualizable → 503); un paso
// fallido es un desenlace normal: reporte con Resultado=fallo y error nil.
func (s *SelfUpdateService) Actualizar(ctx context.Context) (SelfUpdateReport, error) {
	if !s.mu.TryLock() {
		return SelfUpdateReport{}, ErrActualizacionEnCurso
	}
	// Tras «actualizado» el lock se RETIENE (decisión #9): el proceso es terminal
	// (re-exec en ≤1s) y un segundo POST en esa ventana merece un 409 honesto, no un
	// rebuild fantasma sobre un binario ya reemplazado.
	liberar := true
	defer func() {
		if liberar {
			s.mu.Unlock()
		}
	}()

	v := s.updater.Version()
	if !v.Escribible {
		return SelfUpdateReport{}, fmt.Errorf(
			"%w: el binario en %s no es escribible por este proceso — migra a ~/.local/bin (RF-102)",
			ErrNoActualizable, v.InstaladoEn)
	}
	if v.Repo == "" {
		return SelfUpdateReport{}, fmt.Errorf(
			"%w: repo no configurado — arranca el daemon con --repo o ARNESIA_REPO (RF-103)",
			ErrNoActualizable)
	}

	pasos := make([]PasoReporte, 0, 5)
	falla := func(paso string, detalle string, err error, restantes ...string) SelfUpdateReport {
		d := err.Error()
		if detalle != "" {
			d += "\n" + detalle
		}
		pasos = append(pasos, PasoReporte{Paso: paso, Estado: EstadoFallo, Detalle: d})
		for _, r := range restantes {
			pasos = append(pasos, PasoReporte{Paso: r, Estado: EstadoNoCorrido})
		}
		return SelfUpdateReport{Resultado: ResultadoFallo, Pasos: pasos}
	}

	detalle, err := s.updater.Verificar(ctx)
	if err != nil {
		return falla(PasoVerificar, detalle, err, PasoBuild, PasoVerificarBinario, PasoInstalar, PasoReiniciar), nil
	}
	pasos = append(pasos, PasoReporte{Paso: PasoVerificar, Estado: EstadoOK, Detalle: detalle})

	detalle, err = s.updater.Build(ctx)
	if err != nil {
		return falla(PasoBuild, detalle, err, PasoVerificarBinario, PasoInstalar, PasoReiniciar), nil
	}
	pasos = append(pasos, PasoReporte{Paso: PasoBuild, Estado: EstadoOK, Detalle: detalle})

	huella, detalle, err := s.updater.VerificarBinario(ctx)
	if err != nil {
		return falla(PasoVerificarBinario, detalle, err, PasoInstalar, PasoReiniciar), nil
	}
	pasos = append(pasos, PasoReporte{Paso: PasoVerificarBinario, Estado: EstadoOK, Detalle: detalle})

	if huella == v.Huella {
		pasos = append(pasos,
			PasoReporte{Paso: PasoInstalar, Estado: EstadoNoCorrido, Detalle: "misma huella que el binario corriendo — nada que instalar"},
			PasoReporte{Paso: PasoReiniciar, Estado: EstadoNoCorrido, Detalle: "sin reinicio"})
		return SelfUpdateReport{Resultado: ResultadoYaAlDia, HuellaNueva: huella, Pasos: pasos}, nil
	}

	detalle, err = s.updater.Instalar(ctx)
	if err != nil {
		return falla(PasoInstalar, detalle, err, PasoReiniciar), nil
	}
	pasos = append(pasos, PasoReporte{Paso: PasoInstalar, Estado: EstadoOK, Detalle: detalle})

	pasos = append(pasos, PasoReporte{
		Paso: PasoReiniciar, Estado: EstadoAgendado,
		Detalle: "re-exec en ≤1s — la UI reconecta por /api/version",
	})
	liberar = false
	return SelfUpdateReport{Resultado: ResultadoActualizado, HuellaNueva: huella, Pasos: pasos}, nil
}
