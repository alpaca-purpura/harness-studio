package usecase

import (
	"context"
	"errors"
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

// SetArchivoCerradas cablea el registro de sesiones archivadas (RF-200/306): Close deja de
// borrar sin rastro — la sesión entera, con sus conversaciones completas, se appendea ahí.
// nil = comportamiento previo (borrado seco), degradación honesta.
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

// archivarLocked persiste la sesión ENTERA que se cierra (RF-200, RF-305/306), con sus
// conversaciones COMPLETAS: identidad, arnés, cwd, fecha de archivado, y por cada
// conversación su cadena de ClaudeSessionIDs, su transcript y su checkpoint.
//
// ⚠ LEY INVERTIDA en el paquete 2026-07-26-conversaciones-del-panel (CV-D8 + enmienda F-3,
// ledger HS-29). Hasta acá el archivado tiraba `Conv` y `Checkpoint` con el argumento de
// que «la JSONL nativa es la verdad». La JSONL sigue siendo la verdad DEL CONTENIDO — pero
// `Conv` es nuestra copia de presentación, y es lo único que sobrevive a un GC del corpus
// de Claude Code, lo único sobre lo que se puede buscar, y lo que se repinta al retomar.
// Tirarlo era perder lo que el operador vio.
//
// Best-effort: un archivo que falla no impide cerrar (warn, jamás bloquea). Caller holds s.mu.
func (s *SessionService) archivarLocked(meta domain.Session) {
	if s.cerradas == nil {
		return
	}
	cerrada := meta
	// La cadena y el transcript son de cada CONVERSACIÓN (CV-D3): se archiva la sesión
	// entera con las suyas. Copia propia del slice — meta comparte backing array con el
	// runtime vivo y archivar no puede mutarlo.
	cerrada.Conversaciones = append([]domain.Conversacion(nil), meta.Conversaciones...)
	for i := range cerrada.Conversaciones {
		c := &cerrada.Conversaciones[i]
		if c.ClaudeSessionID != "" {
			c.CadenaCC = append(append([]string(nil), c.CadenaCC...), c.ClaudeSessionID)
		}
	}
	// Sesiones nacidas antes del estampado de Cwd (o cerradas sin spawn en esta vida del
	// daemon): el resolver conoce el dir del arnés — backfill honesto del join del corpus.
	if cerrada.Cwd == "" && s.resolver != nil {
		if cwd, _, rerr := s.resolver.Resolve(cerrada.Arnes); rerr == nil {
			cerrada.Cwd = cwd
		}
	}
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
		return nil, errors.New("historial de cerradas no cableado")
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

// HistorialCerrada devuelve los turnos de una sesión archivada (RF-202).
//
// Desde que el archivado conserva el transcript (CV-D8), el camino normal es leerlo del
// propio registro: es lo que el operador vio, y no depende de que el corpus de Claude Code
// siga en disco.
//
// El lector JSONL queda como FALLBACK para una sesión archivada cuyo `Conv` viene vacío —
// hoy, una que se archivó sin ningún turno. (A-9: este comentario decía «para las sesiones
// archivadas ANTES de la migración». Ese caso NO puede llegar acá: esas sesiones viven en
// el `sesiones-cerradas.json` que este binario ya no lee, así que todo lo que se lista
// salió del registro nuevo, que siempre conserva `Conv`. El fallback sigue siendo útil,
// pero por otra razón que la que estaba escrita.)
//
// Y ahí sigue siendo best-effort honesto: las JSONL que ya no están se reportan en
// `faltantes`, jamás se inventa contenido.
func (s *SessionService) HistorialCerrada(ctx context.Context, id string) (turnos []domain.Turn, faltantes []string, err error) {
	s.mu.Lock()
	reader := s.historial
	s.mu.Unlock()

	cerradasTodas, err := s.Cerradas(ctx, "")
	if err != nil {
		return nil, nil, err
	}
	for _, c := range cerradasTodas {
		if c.ID != id {
			continue
		}
		// Camino normal: el transcript archivado.
		archivado := false
		for _, conv := range c.Conversaciones {
			if len(conv.Conv) > 0 {
				turnos = append(turnos, conv.Conv...)
				archivado = true
			}
		}
		if archivado {
			return turnos, nil, nil
		}
		// Fallback: archivada antes de la migración, sin transcript propio. Se cose desde
		// las JSONL nativas de la cadena de cada conversación; el cwd es de la SESIÓN,
		// porque sus conversaciones corren todas en el mismo directorio.
		if reader == nil {
			return nil, nil, errors.New("lector de historial no cableado")
		}
		for _, conv := range c.Conversaciones {
			for _, ccid := range conv.CadenaCC {
				t, terr := reader.Turnos(c.Cwd, ccid)
				if terr != nil {
					faltantes = append(faltantes, ccid)
					continue
				}
				turnos = append(turnos, t...)
			}
		}
		return turnos, faltantes, nil
	}
	return nil, nil, errNotFound(id)
}
