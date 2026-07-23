package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// HistoryReader es el lector del corpus JSONL nativo (RF-201, B2/MC-D7) — satisfecho por
// history.Reader. La JSONL es la verdad; ArnesIA solo aporta el join (cwd + cadena).
type HistoryReader interface {
	Turnos(cwd, claudeSessionID string) ([]domain.Turn, error)
}

// SetArchivoCerradas cablea el registro de sesiones cerradas (RF-200): Close deja de
// borrar sin rastro — la metadata liviana (sin Conv) se appendea ahí. nil = comportamiento
// previo (borrado seco), degradación honesta.
func (s *SessionService) SetArchivoCerradas(store ports.SessionStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cerradas = store
}

// SetHistoryReader cablea el lector JSONL del historial (RF-201/202). nil = el endpoint
// de historial responde honesto que no hay lector.
func (s *SessionService) SetHistoryReader(r HistoryReader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.historial = r
}

// archivarLocked persiste la metadata liviana de una sesión que se cierra (RF-200):
// identidad, arnés, cwd, cadena de ClaudeSessionIDs (rotaciones + el vivo), cantidad de
// turnos y fecha de cierre — SIN Conv (B2: la JSONL nativa es la verdad del contenido).
// Best-effort: un archivo que falla no impide cerrar (warn, jamás bloquea). Caller holds s.mu.
func (s *SessionService) archivarLocked(meta domain.Session) {
	if s.cerradas == nil {
		return
	}
	cerrada := meta
	if cerrada.ClaudeSessionID != "" {
		cerrada.CadenaCC = append(cerrada.CadenaCC, cerrada.ClaudeSessionID)
	}
	cerrada.Turnos = len(cerrada.Conv)
	cerrada.Conv = nil
	cerrada.Checkpoint = ""
	cerrada.CerradaEn = time.Now().UTC().Format(time.RFC3339)
	previas, err := s.cerradas.Load(s.baseCtx)
	if err != nil {
		slog.Warn("session: archivo de cerradas ilegible — se archiva desde cero", "err", err)
		previas = nil
	}
	if err := s.cerradas.Save(s.baseCtx, append(previas, cerrada)); err != nil {
		slog.Warn("session: no se pudo archivar la sesión cerrada", "session", meta.ID, "err", err)
	}
}

// Cerradas lista la metadata de las sesiones cerradas (RF-202); arnesID no vacío filtra.
func (s *SessionService) Cerradas(ctx context.Context, arnesID string) ([]domain.Session, error) {
	s.mu.Lock()
	store := s.cerradas
	s.mu.Unlock()
	if store == nil {
		return nil, fmt.Errorf("historial de cerradas no cableado")
	}
	todas, err := store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("historial de cerradas: %w", err)
	}
	if arnesID == "" {
		return todas, nil
	}
	out := make([]domain.Session, 0, len(todas))
	for _, c := range todas {
		if c.Arnes == arnesID {
			out = append(out, c)
		}
	}
	return out, nil
}

// HistorialCerrada reconstruye la conversación de una sesión cerrada desde las JSONL
// nativas de su cadena (RF-202) — best-effort honesto: las JSONL que ya no están en disco
// se reportan en faltantes, jamás se inventa contenido.
func (s *SessionService) HistorialCerrada(ctx context.Context, id string) (turnos []domain.Turn, faltantes []string, err error) {
	s.mu.Lock()
	reader := s.historial
	s.mu.Unlock()
	if reader == nil {
		return nil, nil, fmt.Errorf("lector de historial no cableado")
	}
	cerradasTodas, err := s.Cerradas(ctx, "")
	if err != nil {
		return nil, nil, err
	}
	for _, c := range cerradasTodas {
		if c.ID != id {
			continue
		}
		for _, ccid := range c.CadenaCC {
			t, terr := reader.Turnos(c.Cwd, ccid)
			if terr != nil {
				faltantes = append(faltantes, ccid)
				continue
			}
			turnos = append(turnos, t...)
		}
		return turnos, faltantes, nil
	}
	return nil, nil, errNotFound(id)
}
