package ports

import "context"

// PublishPort publishes a harness to a marketplace repo (the release train, KIT-06).
type PublishPort interface {
	// Publish pushes harnessID to target (a marketplace repo).
	Publish(ctx context.Context, harnessID, target string) error
}
