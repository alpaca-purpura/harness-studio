// Package permission implements ports.PermissionPort: the KitProvisioner resolves an
// arnés's permission-set from the ROLE that hydrates it (permisos-derivan-del-rol.md).
// The policy here is a spike (HS-08) standing in for the external L1 authority; what
// matters is the SHAPE: the same tool yields different decisions per role, and grants
// expire (least temporal privilege). ArnesIA consumes the authority; it does not own it.
package permission

import (
	"context"
	"errors"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// KitProvisioner resolves permission-sets by role. It parametrizes the spawn: the
// resolved set becomes the conductor's --permission-mode/--settings at hydration.
type KitProvisioner struct {
	policy map[string]domain.PermissionSet
	base   domain.PermissionSet
}

var _ ports.PermissionPort = (*KitProvisioner)(nil)

// NewKitProvisioner returns a provisioner seeded with the spike policy. Real deployments
// will hydrate `policy` from the external L1 system; the base is the deny-by-default
// fallback for an unknown role.
func NewKitProvisioner() *KitProvisioner {
	return &KitProvisioner{
		base: domain.PermissionSet{
			Rol:   "*",
			Allow: []string{"Read", "Grep", "Glob"},
			Deny:  []string{"Bash", "Write", "Edit"},
			TTL:   0, // per-task: an unknown role gets nothing perpetual.
		},
		policy: map[string]domain.PermissionSet{
			// A full-cycle dev: broad authority, 15-minute grants.
			"backend-dev": {
				Rol:   "backend-dev",
				Allow: []string{"Read", "Grep", "Glob", "Edit", "Write"},
				Ask:   []string{"Bash", "WebFetch"},
				TTL:   15 * time.Minute,
			},
			// El rol REAL del dogfood (dev-full-cycle/arnes.l0.json META): misma autoridad
			// que backend-dev + los tools de plan/tareas (no tocan disco) pre-aprobados
			// para que el Dock no interrumpa por un TodoWrite. Sigue siendo spike: la
			// autoridad real vendrá del sistema L1 externo.
			"Ingeniería · Desarrollo full-cycle": {
				Rol:   "Ingeniería · Desarrollo full-cycle",
				Allow: []string{"Read", "Grep", "Glob", "Edit", "Write", "TodoWrite", "Task", "TaskCreate", "TaskUpdate", "TaskGet", "TaskList"},
				Ask:   []string{"Bash", "WebFetch", "WebSearch"},
				TTL:   15 * time.Minute,
			},
			// A reviewer: read-mostly, no mutation, short-lived approvals.
			"reviewer": {
				Rol:   "reviewer",
				Allow: []string{"Read", "Grep", "Glob"},
				Ask:   []string{"Bash"},
				Deny:  []string{"Write", "Edit"},
				TTL:   5 * time.Minute,
			},
		},
	}
}

// ResolveForRole returns the permission-set the role authorizes, or the deny-by-default
// base for an unknown role. It never errors on an unknown role — it degrades to minimum.
func (k *KitProvisioner) ResolveForRole(_ context.Context, rol string) (domain.PermissionSet, error) {
	if rol == "" {
		return domain.PermissionSet{}, errors.New("rol vacío: el permission-set deriva del rol (META de enganche)")
	}
	if ps, ok := k.policy[rol]; ok {
		return ps, nil
	}
	base := k.base
	base.Rol = rol
	return base, nil
}
