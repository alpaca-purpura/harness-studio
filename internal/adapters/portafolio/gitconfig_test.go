package portafolio

import (
	"os"
	"path/filepath"
	"testing"
)

func escribirCfg(t *testing.T, ruta, contenido string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestGitConfigParse cubre origin/upstream/worktree-gitdir/bare (S0-D8, sin go-git).
func TestGitConfigParse(t *testing.T) {
	t.Run("origin gana, upstream es secundario (C-N-2)", func(t *testing.T) {
		dir := t.TempDir()
		escribirCfg(t, filepath.Join(dir, ".git", "config"), `
[core]
	bare = false
[remote "origin"]
	url = https://github.com/usuario/fork.git
[remote "upstream"]
	url = https://github.com/canonico/original.git
`)
		gr, ok := ResolverRemotesGit(dir)
		if !ok {
			t.Fatal("quiero remotes resueltos")
		}
		if gr.Origin != "https://github.com/usuario/fork.git" {
			t.Errorf("Origin = %q", gr.Origin)
		}
		if gr.Upstream != "https://github.com/canonico/original.git" {
			t.Errorf("Upstream = %q", gr.Upstream)
		}
	})

	t.Run("worktree: .git archivo gitdir: resuelve al gitdir real (C-N-1)", func(t *testing.T) {
		principal := t.TempDir()
		escribirCfg(t, filepath.Join(principal, "config"), `
[remote "origin"]
	url = git@github.com:owner/repo.git
`)
		worktree := t.TempDir()
		escribirCfg(t, filepath.Join(worktree, ".git"), "gitdir: "+principal+"\n")

		gr, ok := ResolverRemotesGit(worktree)
		if !ok {
			t.Fatal("quiero resolver remotes vía worktree gitdir:")
		}
		if gr.Origin != "git@github.com:owner/repo.git" {
			t.Errorf("Origin = %q", gr.Origin)
		}
	})

	t.Run("bare repo detectado vía core.bare", func(t *testing.T) {
		dir := t.TempDir()
		escribirCfg(t, filepath.Join(dir, ".git", "config"), "[core]\n\tbare = true\n")
		gd, ok := gitDir(dir)
		if !ok {
			t.Fatal("quiero gitDir resuelto")
		}
		if !esBare(gd) {
			t.Error("quiero detectar bare=true")
		}
	})

	t.Run("no bare por defecto", func(t *testing.T) {
		dir := t.TempDir()
		escribirCfg(t, filepath.Join(dir, ".git", "config"), "[core]\n\tbare = false\n")
		gd, _ := gitDir(dir)
		if esBare(gd) {
			t.Error("no debe reportar bare cuando core.bare=false")
		}
	})

	t.Run("sin .git: no resuelve", func(t *testing.T) {
		if _, ok := ResolverRemotesGit(t.TempDir()); ok {
			t.Error("un dir sin .git no debe resolver remotes")
		}
	})
}
