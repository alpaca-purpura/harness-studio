package domain

import "time"

// This file models role-derived permissions (arch/boundaries/permisos-derivan-del-rol.md,
// Salesforce + APM normative frame). A permission-set is NOT a fixed default: it is what
// the ROLE authorizes, so the SAME tool carries different decisions per role. Permissions
// are enforced outside the model's reasoning (hooks / control_request) and can be
// task-based / expiring (least temporal privilege), not perpetual. The authority itself
// is external (the future L1 organigrama, VISION §Linaje) — this is the mechanism seam.

// Decision is the outcome of a permission lookup for a tool.
type Decision string

const (
	// DecisionAllow — the tool runs without human approval.
	DecisionAllow Decision = "allow"
	// DecisionAsk — the tool requires human-in-the-loop approval (control_request).
	DecisionAsk Decision = "ask"
	// DecisionDeny — the tool is hard-denied (enforced in a hook, exit 2).
	DecisionDeny Decision = "deny"
)

// PermissionSet is the tool authority a role grants. Deny wins over Ask wins over Allow.
// TTL is how long a granted (ask→approved) permission stays valid before it must be
// re-approved — 0 means per-task (no persistence across tasks), never perpetual.
type PermissionSet struct {
	Rol   string        `json:"rol"`
	Allow []string      `json:"allow,omitempty"`
	Ask   []string      `json:"ask,omitempty"`
	Deny  []string      `json:"deny,omitempty"`
	TTL   time.Duration `json:"ttl,omitempty"`
}

// Decide resolves a tool to a decision for this role. Precedence: Deny > Ask > Allow;
// anything not listed defaults to Ask (deny-by-default posture: never silently allow).
func (ps PermissionSet) Decide(tool string) Decision {
	if contains(ps.Deny, tool) {
		return DecisionDeny
	}
	if contains(ps.Ask, tool) {
		return DecisionAsk
	}
	if contains(ps.Allow, tool) {
		return DecisionAllow
	}
	return DecisionAsk // deny-by-default: unlisted tools need approval.
}

// Grant is a task-based, EXPIRING permission (least temporal privilege). A grant is
// minted when an Ask tool is approved; it is valid only until Expira.
type Grant struct {
	Tool     string    `json:"tool"`
	Rol      string    `json:"rol"`
	Otorgado time.Time `json:"otorgado"`
	Expira   time.Time `json:"expira"`
}

// Vigente reports whether the grant is still valid at `now`. An expired grant is NOT
// valid — the tool must be re-approved (no perpetual access).
func (g Grant) Vigente(now time.Time) bool {
	return now.Before(g.Expira)
}

// NuevoGrant mints a grant for a tool under a role's permission-set, expiring after the
// set's TTL. A zero TTL yields an already-expired grant (per-task: valid only for the
// approving turn), enforcing that nothing is granted perpetually.
func (ps PermissionSet) NuevoGrant(tool string, now time.Time) Grant {
	return Grant{Tool: tool, Rol: ps.Rol, Otorgado: now, Expira: now.Add(ps.TTL)}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
