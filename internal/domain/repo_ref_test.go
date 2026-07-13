package domain_test

import (
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

func TestCanonicalizarRepo(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		want   string
		wantOk bool
	}{
		{"slug corto asume github.com", "alpacapurpura/prenter-marketplace", "github.com/alpacapurpura/prenter-marketplace", true},
		{"https", "https://github.com/alpacapurpura/prenter-marketplace", "github.com/alpacapurpura/prenter-marketplace", true},
		{"https con .git", "https://github.com/alpacapurpura/prenter-marketplace.git", "github.com/alpacapurpura/prenter-marketplace", true},
		{"git@ scp-like", "git@github.com:alpacapurpura/prenter-marketplace.git", "github.com/alpacapurpura/prenter-marketplace", true},
		{"ssh://", "ssh://git@github.com/alpacapurpura/prenter-marketplace", "github.com/alpacapurpura/prenter-marketplace", true},
		{"trailing slash", "https://github.com/alpacapurpura/prenter-marketplace/", "github.com/alpacapurpura/prenter-marketplace", true},
		{"mayúsculas", "https://GitHub.com/AlpacaPurpura/Prenter-Marketplace", "github.com/alpacapurpura/prenter-marketplace", true},
		{"host no-github", "https://gitlab.com/owner/repo", "gitlab.com/owner/repo", true},
		{"path con 3+ segmentos (subgrupo)", "https://gitlab.com/grupo/subgrupo/repo", "gitlab.com/grupo/subgrupo/repo", true},
		{"no-parseable: esquema desconocido", "ftp://host/owner/repo", "", false},
		{"no-parseable: sin owner/repo", "solo-un-nombre", "", false},
		{"no-parseable: vacío", "", "", false},
		{"no-parseable: host sin punto", "https://localhost/owner/repo", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := domain.CanonicalizarRepo(c.in)
			if ok != c.wantOk {
				t.Fatalf("CanonicalizarRepo(%q) ok = %v, quiero %v (got=%q)", c.in, ok, c.wantOk, got)
			}
			if ok && got != c.want {
				t.Errorf("CanonicalizarRepo(%q) = %q, quiero %q", c.in, got, c.want)
			}
		})
	}
}
