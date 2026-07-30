package ports

import "github.com/alpacapurpura/arnesia/internal/domain"

// ForjaPort es el puerto de la siembra `.arnesia/` (contrato semilla-arnesia.md, A-T3):
// satisfecho por internal/adapters/forja (scaffolder + doctor sobre la semilla embebida).
// El usecase valida el dir (política de path protegido) y delega acá; el wiring del FS
// embebido vive en cmd (composition root), jamás en el usecase.
type ForjaPort interface {
	// Sembrar materializa el árbol del contrato §2 bajo dir/.arnesia/ — idempotente:
	// archivo existente JAMÁS se pisa (viaja en YaExistian); al final escribe
	// `semilla.lock.json` (schema 0, §3) con los hashes de lo que ESTA siembra escribió.
	Sembrar(dir string) (domain.InformeSemilla, error)
	// Chequear juzga la salud de la instalación por PRESENCIA (doctor v0, §4):
	// sana | ausente | incompleta + faltantes. Un lock ilegible cuenta como malformado.
	Chequear(dir string) (domain.SaludSemilla, error)
}
