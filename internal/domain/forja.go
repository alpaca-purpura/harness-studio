package domain

// forja.go — el vocabulario de la siembra `.arnesia/` (process-as-code en el proyecto del
// usuario; contrato docs/architecture/contracts/semilla-arnesia.md, paquete
// 2026-07-30-arnesia-en-el-proyecto A-T3). La siembra es idempotente y ADITIVA: lo
// existente JAMÁS se pisa; la salud se juzga por PRESENCIA (sin hashes — la deriva de
// semilla es gap declarado Fase 2, contrato §3).

// InformeSemilla es el resultado de una siembra: qué archivos se crearon y cuáles ya
// existían (y por eso se respetaron byte a byte — contrato §4.1). Re-correr sobre una
// instalación sana ⇒ todo en YaExistian, cero escrituras.
type InformeSemilla struct {
	Creados    []string `json:"creados"`
	YaExistian []string `json:"ya_existian"`
}

// EstadoSemilla es el veredicto del doctor v0 (contrato §4.4): visible, jamás fabricado.
type EstadoSemilla string

const (
	// SemillaSana — `.arnesia/` presente con TODOS los archivos del árbol del contrato §2.
	SemillaSana EstadoSemilla = "sana"
	// SemillaAusente — no existe `.arnesia/` en el proyecto.
	SemillaAusente EstadoSemilla = "ausente"
	// SemillaIncompleta — existe pero faltan piezas (o el lock está malformado); el
	// veredicto LISTA los faltantes.
	SemillaIncompleta EstadoSemilla = "incompleta"
)

// SaludSemilla es el diagnóstico completo: estado + faltantes listados + detalle humano.
// «No avanzamos si no está sana» (A-D4): un estado ≠ sana hace exit≠0 en el CLI.
type SaludSemilla struct {
	Estado    EstadoSemilla `json:"estado"`
	Faltantes []string      `json:"faltantes,omitempty"`
	Detalle   string        `json:"detalle,omitempty"`
}
