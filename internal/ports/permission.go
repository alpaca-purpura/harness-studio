package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// PermissionPort resolves the permission-set an arnés runs with, parametrized by the ROLE
// that hydrates it (docs/architecture/boundaries/permisos-derivan-del-rol.md). The concrete adapter is
// the KitProvisioner; the authority behind the policy is external (the future L1
// organigrama), so this port is the seam ArnesIA consumes, not a consultancy of ours.
type PermissionPort interface {
	// ResolveForRole returns the permission-set the given role authorizes.
	ResolveForRole(ctx context.Context, rol string) (domain.PermissionSet, error)
}
