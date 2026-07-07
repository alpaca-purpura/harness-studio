package ports

import "context"

// This file declares the ports the T3 box conductor needs beyond AgentPort. The
// conductor owns the loop (12-Factor «own your control flow» / DAOP-A2) and reads the
// artifact `status` (document-as-cache) through ArtifactReader — it never scrapes the
// chat text (arch/boundaries/orquestacion-determinista-entre-cajas.md).

// ArtifactReader reads the document-as-cache `status:` of a box's output artifact. The
// artifact is the source of truth for done/blocked, not the conversation. A concrete
// adapter reads the frontmatter of the artifact file; a fake drives tests.
type ArtifactReader interface {
	// Status returns the artifact's `status:` value and whether the artifact exists yet.
	// dir is the arnés working directory the run is confined to (the conductor's cwd);
	// artifactRef resolves UNDER dir and must not escape it (boundary permisos-gui
	// `sesion-aislada-por-cwd` extended to reads: the conductor only looks inside its tree).
	Status(ctx context.Context, dir, artifactRef string) (status string, exists bool, err error)
}
