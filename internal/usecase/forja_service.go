package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// ForjaService orquesta la siembra `.arnesia/` (contrato semilla-arnesia.md, A-T3): valida
// el dir destino con la MISMA política de path protegido del Portafolio y delega al puerto.
// No importa adapters — el FS embebido y el scaffolder concreto los cablea cmd
// (composition root, patrón arnesLoaderFunc).
type ForjaService struct {
	forja ports.ForjaPort
}

// NewForjaService construye el servicio sobre el puerto concreto que cmd cablea.
func NewForjaService(f ports.ForjaPort) *ForjaService { return &ForjaService{forja: f} }

// Sembrar valida dir (abs · existe · no-protegido) y siembra el árbol del contrato §2.
// Idempotente: lo existente jamás se pisa (informe Creados/YaExistian).
func (s *ForjaService) Sembrar(_ context.Context, dir string) (domain.InformeSemilla, error) {
	if err := validarDirForja(dir); err != nil {
		return domain.InformeSemilla{}, err
	}
	return s.forja.Sembrar(filepath.Clean(dir))
}

// Chequear valida dir con la misma política y corre el doctor v0 (§4): la salud se emite
// tal cual salga — el criterio de exit («no avanzamos si no está sana») vive en el caller.
func (s *ForjaService) Chequear(_ context.Context, dir string) (domain.SaludSemilla, error) {
	if err := validarDirForja(dir); err != nil {
		return domain.SaludSemilla{}, err
	}
	return s.forja.Chequear(filepath.Clean(dir))
}

// validarDirForja REUSA validarRootPortafolio (mismo paquete): abs · existe · dir ·
// no-protegido (~/.claude, ~/.arnesia, ~/.ssh, …) — una semilla jamás se siembra en una
// ubicación protegida. El prefijo «portafolio:» del error envuelto es la voz de la
// política compartida; el nuestro dice de qué operación vino.
func validarDirForja(dir string) error {
	if err := validarRootPortafolio(dir); err != nil {
		return fmt.Errorf("forja: dir destino inválido: %w", err)
	}
	return nil
}
