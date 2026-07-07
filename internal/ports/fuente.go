package ports

import "errors"

// ErrFueraDelArnes marks a fuente_path whose resolved location escapes the arnés's
// registered directory (S2, boundary superficie-local-confinada): the read is refused.
// Declared here so use case and transport can errors.Is it without importing the
// concrete adapter.
var ErrFueraDelArnes = errors.New("la fuente escapa del directorio del arnés")

// FuenteReader reads a node's real source artifact CONFINED to the arnés directory
// (RF-93, tab Contenido del inspector). The drawer shows the file exactly as the
// loader recognized it — autodocumentación como efecto; the reader never looks
// outside the tree the arnés is registered to.
type FuenteReader interface {
	// LeerConfinado reads ref — relative to dir, or already-absolute INSIDE dir (the
	// loader stamps absolute fuente_path when the arnés was loaded from an absolute
	// base) — after verifying the resolved path cannot escape dir. A ref that escapes
	// returns ErrFueraDelArnes; a missing file returns the os error untouched.
	LeerConfinado(dir, ref string) ([]byte, error)
}
