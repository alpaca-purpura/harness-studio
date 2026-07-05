package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// SessionStore persists the registry of open sessions — the daemon is the source of
// truth for which work-fronts exist and where each one is parked (chosen HS-06:
// "daemon dueño + claude --resume"). The conversation content itself is NOT stored
// here; it lives in Claude Code's own JSONL and is rehydrated via --resume.
type SessionStore interface {
	// Load returns all persisted sessions (empty slice if none yet).
	Load(ctx context.Context) ([]domain.Session, error)
	// Save atomically replaces the persisted set with sessions.
	Save(ctx context.Context, sessions []domain.Session) error
}
