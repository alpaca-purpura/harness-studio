package main

import (
	"errors"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// regFija is a minimal ports.ArnesRegistry fake for ownerOf: Register/Resolve unused.
type regFija []ports.ArnesPath

func (r regFija) Resolve(string) (string, bool, error) {
	return "", false, errors.New("regFija: Resolve no implementado")
}

func (r regFija) Register(string, string) error {
	return errors.New("regFija: Register no implementado")
}

func (r regFija) List() []ports.ArnesPath { return r }

// TestOwnerOf cubre RF-210: qué entrada de ArnesRegistry es dueña de un WatchEvent.Path —
// el path exacto de un arnés, un archivo/dir bajo él, un vecino con prefijo similar (no
// debe confundirse), y un path huérfano (ningún arnés registrado lo contiene).
func TestOwnerOf(t *testing.T) {
	reg := regFija{
		{Arnes: "a", Path: "/arneses/a"},
		{Arnes: "a-2", Path: "/arneses/a-2"},
	}

	tests := []struct {
		name   string
		path   string
		wantID string
		wantOK bool
	}{
		{"path exacto del arnés", "/arneses/a", "a", true},
		{"archivo bajo el árbol del arnés", "/arneses/a/skills/x/SKILL.md", "a", true},
		{"vecino con prefijo similar no colisiona", "/arneses/a-2/CLAUDE.md", "a-2", true},
		{"path huérfano", "/otro/lugar/x.txt", "", false},
		{"el propio path raíz de un arnés (no confundido con el vecino 'a')", "/arneses/a-2", "a-2", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap, ok := ownerOf(reg, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("ownerOf(%q) ok = %v, want %v", tt.path, ok, tt.wantOK)
			}
			if ok && ap.Arnes != tt.wantID {
				t.Errorf("ownerOf(%q) = %q, want %q", tt.path, ap.Arnes, tt.wantID)
			}
		})
	}
}
