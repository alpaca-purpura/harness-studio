package ports

// WorkdirResolver maps an arnés id to the absolute working directory its Claude Code
// conductor runs in. It is the port behind boundary permisos-gui-human-in-the-loop's
// rule «aislar sesiones por cwd/worktree»: a session is confined to its arnés's tree,
// never the daemon's shared cwd. HS-06 resolves this from an explicit arnés→path
// registry; a future worktree-per-session resolver can swap in without touching the
// domain (the whole point of the port).
type WorkdirResolver interface {
	// Resolve returns the absolute, sanitized working directory for arnesID.
	//
	// registered reports whether the path came from an explicit registration (true) or a
	// safe per-arnés fallback (false, e.g. <root>/<arnesID>). Either way the returned path
	// is a directory dedicated to this arnés — the resolver NEVER returns a shared global
	// cwd, and it rejects paths that escape containment or hit protected locations
	// (~/.claude, ~/.ssh, $HOME root, filesystem root).
	Resolve(arnesID string) (path string, registered bool, err error)
}

// ArnesRegistry is the write side of the arnés→path mapping the WorkdirResolver reads.
// The composition root wires one concrete adapter that satisfies both. Kept separate so
// the resolver stays read-only for the session hot path.
type ArnesRegistry interface {
	WorkdirResolver
	// Register records (or replaces) the working directory for arnesID after validating
	// it (absolute, existing directory, contained, not a protected path).
	Register(arnesID, path string) error
	// List returns every registered arnés→path entry.
	List() []ArnesPath
}

// ArnesPath is one registered arnés→path entry.
type ArnesPath struct {
	Arnes string `json:"arnes"`
	Path  string `json:"path"`
}
