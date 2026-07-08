package ports

import "context"

// Injection is the flag-package that carries the doctrine bodies ①+② into ONE spawn
// (investigación de inyección, FIRMADO HS-10; METODOLOGIA §9): the kit as a session-scoped
// local plugin, the doctrine overlay as an appended system prompt, and the knowhow
// checklists as a read-only reference dir. Everything lives OUTSIDE the arnés tree —
// the conductor loads it by flags, never by writing into ③.
type Injection struct {
	// PluginDirs are local plugin roots for `--plugin-dir` (② la maquinaria/kit).
	PluginDirs []string
	// SystemPromptFile is the doctrine overlay for `--append-system-prompt-file` (①).
	SystemPromptFile string
	// AddDirs are read-only reference roots for `--add-dir` (① los checklists knowhow).
	AddDirs []string
}

// InjectionProvisioner materializes the embedded doctrine+kit to the app-owned dir
// (~/.arnesia) — idempotent, invisible to the user — and returns the Injection every
// conductor spawn carries. Cuerpo ② nace del binario: actualizar la app ES actualizar
// la maquinaria que ven las sesiones.
type InjectionProvisioner interface {
	Provision(ctx context.Context) (Injection, error)
}
