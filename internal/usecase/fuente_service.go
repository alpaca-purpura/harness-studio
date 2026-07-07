package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Sentinels del servicio de fuente (RF-93) — el transporte los mapea a HTTP sin
// conocer el porqué interno (dominio-independiente-de-transporte).
var (
	// ErrNodoNoEncontrado — el arnés existe pero no tiene ese nodo.
	ErrNodoNoEncontrado = errors.New("el nodo no existe en el grafo del arnés")
	// ErrFuenteNoDeclarada — el nodo no trae fuente_path (pendiente del reconocedor,
	// nomenclatura §3): estado honesto, jamás contenido inventado.
	ErrFuenteNoDeclarada = errors.New("el nodo no declara fuente_path — pendiente del reconocedor")
	// ErrArnesSinDirectorio — el arnés no está registrado a un directorio (S2), así que
	// no hay árbol al que confinar la lectura (p.ej. los fixtures embebidos del showcase).
	ErrArnesSinDirectorio = errors.New(
		"el arnés no tiene directorio registrado — carga la carpeta (PUT /api/arneses/{id}) para leer su fuente")
)

// FuenteService sirve el artefacto REAL de un nodo (tab Contenido del inspector,
// RF-93): la fuente ES la verdad del componente (autodocumentación como efecto).
// Lectura CONFINADA al directorio registrado del arnés — nunca un árbol compartido.
type FuenteService struct {
	index   ports.IndexPort
	workdir ports.WorkdirResolver
	files   ports.FuenteReader
}

// NewFuenteService returns a FuenteService over the index, the arnés→dir resolver and
// the confined reader.
func NewFuenteService(index ports.IndexPort, workdir ports.WorkdirResolver, files ports.FuenteReader) *FuenteService {
	return &FuenteService{index: index, workdir: workdir, files: files}
}

// Fuente returns the node's fuente_path and its raw bytes, read confined to the
// arnés's REGISTERED directory. Error taxonomy: unknown harness (index error) ·
// ErrNodoNoEncontrado · ErrFuenteNoDeclarada · ErrArnesSinDirectorio ·
// ports.ErrFueraDelArnes (traversal) · read errors (missing file said as-is).
func (s *FuenteService) Fuente(ctx context.Context, harnessID, nodeID string) (string, []byte, error) {
	g, err := s.index.Query(ctx, harnessID)
	if err != nil {
		return "", nil, err
	}
	b, ok := g.NodeByID(nodeID)
	if !ok {
		return "", nil, ErrNodoNoEncontrado
	}
	if b.FuentePath == "" {
		return "", nil, ErrFuenteNoDeclarada
	}
	dir, registered, err := s.workdir.Resolve(harnessID)
	if err != nil {
		return "", nil, fmt.Errorf("resolver directorio del arnés %s: %w", harnessID, err)
	}
	if !registered {
		// El resolver tiene un fallback por-arnés para las SESIONES; para leer fuente el
		// registro explícito es la única base honesta (el fallback está vacío).
		return "", nil, ErrArnesSinDirectorio
	}
	contenido, err := s.files.LeerConfinado(dir, b.FuentePath)
	if err != nil {
		return "", nil, err
	}
	return b.FuentePath, contenido, nil
}
