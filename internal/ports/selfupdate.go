package ports

import "context"

// VersionInfo es la identidad honesta del binario corriendo (RF-107): huella VCS del
// buildinfo, dónde está instalado y si el proceso puede reemplazarlo. La UI jamás la
// inventa — la lee de GET /api/version.
type VersionInfo struct {
	// Huella es vcs.revision[:7] del buildinfo, con sufijo "+sucio" si vcs.modified
	// (decisión #7 del paquete boton-actualizar); un binario sin VCS info dice "dev".
	Huella string
	// Fecha es la fecha del commit (vcs.time, solo día); vacía sin VCS info.
	Fecha string
	// InstaladoEn es la ruta REAL del ejecutable, capturada al construir el adapter
	// (tras el rename /proc/self/exe reporta «(deleted)» — decisión #8).
	InstaladoEn string
	// Repo es la ruta del repo configurada al daemon (flag --repo / env ARNESIA_REPO);
	// vacía = sin repo, el self-update no puede correr (RF-103).
	Repo string
	// Escribible dice si el proceso puede escribir el DIRECTORIO del ejecutable (el
	// rename atómico exige write en el dir, no en el archivo) — RF-102.
	Escribible bool
	// Sucio refleja vcs.modified del build corriente (aditivo al RF-107, decisión #7).
	Sucio bool
}

// SelfUpdater es el puerto del self-update sin sudo (RF-104): cada paso devuelve el
// detalle humano y su error; el ORDEN y el corte al primer fallo los orquesta el
// usecase, igual que la regla de negocio «ya al día» (comparar huellas es negocio,
// no mecánica del adapter).
type SelfUpdater interface {
	// Version reporta la identidad del binario corriendo (RF-107).
	Version() VersionInfo
	// Verificar valida el terreno: repo configurado y del árbol esperado, toolchain
	// presente (go · pnpm) — paso ①.
	Verificar(ctx context.Context) (detalle string, err error)
	// Build compila el árbol local (scripts/bundle.sh --daemon-only, cwd=repo,
	// cancelable) — paso ②. El detalle de un fallo lleva la cola del stderr.
	Build(ctx context.Context) (detalle string, err error)
	// VerificarBinario inspecciona el binario recién compilado SIN ejecutarlo
	// (existe · ejecutable · buildinfo legible) y devuelve su huella — paso ③.
	VerificarBinario(ctx context.Context) (huella, detalle string, err error)
	// Instalar reemplaza el ejecutable corriente: write-tmp EN SU MISMO directorio →
	// rename atómico; un fallo deja el binario instalado INTACTO — paso ④.
	Instalar(ctx context.Context) (detalle string, err error)
	// Reiniciar re-ejecuta el daemon (mismo path, mismos args) — paso ⑤; el transport
	// lo agenda POST-respuesta.
	Reiniciar() error
	// ConfigurarRepo valida path con las MISMAS reglas que Verificar (repo existe · es
	// el módulo esperado · trae scripts/bundle.sh · toolchain en PATH) y, solo si pasa,
	// lo fija como repo activo EN CALIENTE (bugfix fix-repo-self-update, RF-109). No
	// persiste — la persistencia vive en RepoConfigStore, responsabilidad del usecase.
	ConfigurarRepo(ctx context.Context, path string) (detalle string, err error)
}

// RepoConfigStore persiste el path del repo configurado vía UI cuando el daemon
// arrancó sin --repo/ARNESIA_REPO (bugfix fix-repo-self-update, RF-108). Vive aparte
// del puerto SelfUpdater porque es persistencia, no mecánica de update — mismo patrón
// que ports.PortafolioStore.
type RepoConfigStore interface {
	// Leer devuelve el path persistido, o "" si nunca se configuró nada.
	Leer() (path string, err error)
	// Guardar persiste path (escritura atómica); reemplaza cualquier valor previo.
	Guardar(path string) error
}
