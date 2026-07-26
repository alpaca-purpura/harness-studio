package httpapi

import (
	"errors"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// marketplace.go expone el plano Marketplaces + el catálogo + `↧ Traer canónico` (paquete
// 2026-07-23-portafolio-agregar-marketplace, design.md §7/§13.8). Cero regla de negocio acá: el
// status sale de `errors.Is` sobre los centinelas del usecase/dominio, jamás de parsear texto.
//
// Los tres códigos que este paquete distingue a propósito, porque el FE los usa distinto:
//   400 = «tu pedido está mal»   ·   502 = «miré y el estante mintió»   ·   503 = «no puedo mirar»

// conflictoMarketplaceBody — 409 de POST /api/marketplaces (BR-7): además del motivo, el nombre
// existente, para que la UI ofrezca navegar en vez de solo mostrar un error (E-18).
type conflictoMarketplaceBody struct {
	Error  string `json:"error"`
	Nombre string `json:"nombre"`
}

// conflictoTraerBody — 409 de POST …/traidos (BR-14): el destino viaja para que la UI pueda
// ofrecer «abrir el canónico que ya tenés».
type conflictoTraerBody struct {
	Error   string `json:"error"`
	Destino string `json:"destino"`
}

// listMarketplaces — GET /api/marketplaces: el plano (S2). 200 SIEMPRE que el servicio esté
// cableado, incluso con detector ilegible o registro corrupto (van en `aviso_detector` /
// `corruptas`) — nunca un 500, nunca una lista vacía muda (E-68/E-75).
func listMarketplaces(svc *usecase.MarketplaceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		out, err := svc.Listar(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// postValidarBody is the POST /api/marketplaces/validaciones payload.
type postValidarBody struct {
	URL string `json:"url"`
}

// postValidarMarketplace — POST /api/marketplaces/validaciones: prueba una url SIN persistir
// nada. **Solo un 200 pinta ✓** (BR-5/G3 hecho contrato de transporte).
func postValidarMarketplace(svc *usecase.MarketplaceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		var body postValidarBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		v, err := svc.Validar(r.Context(), body.URL)
		if err != nil {
			writeJSON(w, statusDeLectura(err), errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, v)
	}
}

// postRegistrarBody is the POST /api/marketplaces payload.
type postRegistrarBody struct {
	URL   string `json:"url"`
	Clase string `json:"clase"`
}

// postRegistrarMarketplace — POST /api/marketplaces: persiste el lado declarado. 409 con el
// nombre existente si ya está (BR-7: no duplica, NO PISA).
func postRegistrarMarketplace(svc *usecase.MarketplaceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		var body postRegistrarBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		fila, err := svc.Registrar(r.Context(), body.URL, domain.ClaseMarketplace(body.Clase))
		if err != nil {
			if errors.Is(err, usecase.ErrMarketplaceYaRegistrado) {
				writeJSON(w, http.StatusConflict, conflictoMarketplaceBody{Error: err.Error(), Nombre: fila.Nombre})
				return
			}
			writeJSON(w, statusDeLectura(err), errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, fila)
	}
}

// deleteOlvidarMarketplace — DELETE /api/marketplaces/{nombre}: quita el lado DECLARADO.
// `sigue_detectado` es honestidad: no podemos hacer que Claude Code deje de conocerlo (E-71).
func deleteOlvidarMarketplace(svc *usecase.MarketplaceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		olvidado, sigueDetectado, err := svc.Olvidar(r.Context(), r.PathValue("nombre"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		if !olvidado {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "nombre no registrado por el operador"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"olvidado": true, "sigue_detectado": sigueDetectado})
	}
}

// getCatalogoMarketplace — GET /api/marketplaces/{nombre}/catalogo: caché primero (la primera
// visita hace UNA lectura y la persiste: idempotente).
func getCatalogoMarketplace(svc *usecase.MarketplaceService) http.HandlerFunc {
	return catalogoHandler(svc, false)
}

// postLeerCatalogo — POST /api/marketplaces/{nombre}/lecturas: REFRESCO EXPLÍCITO (AG-D8
// decisión 3). Es lo que pegan `Refrescar`, `Reintentar` y `Leer catálogo`. Body vacío.
func postLeerCatalogo(svc *usecase.MarketplaceService) http.HandlerFunc {
	return catalogoHandler(svc, true)
}

// catalogoHandler es el handler compartido del GET y del POST de lecturas. 200 incluso si NO se
// pudo leer (`entradas: null` + `lectura.motivo`): **no hay 5xx por «no pude leer el catálogo»**
// — un 500 no le da al FE nada que renderizar y lo empujaría a mostrar una lista vacía, el pass
// fabricado que BR-4 mata.
func catalogoHandler(svc *usecase.MarketplaceService, refrescar bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		cat, err := svc.Catalogo(r.Context(), r.PathValue("nombre"), refrescar)
		if err != nil {
			if errors.Is(err, usecase.ErrMarketplaceNoConocido) {
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, cat)
	}
}

// postTraerBody is the POST /api/marketplaces/{nombre}/traidos payload. Se pide la FILA del
// catálogo, no el `source`: el cliente NO elige el mecanismo — lo decide `PlanificarTraer`.
type postTraerBody struct {
	Entrada string `json:"entrada"`
}

// postTraerCanonico — POST /api/marketplaces/{nombre}/traidos (AG-D17, §13.8). Un 200 puede traer
// `deriva: "en-deriva"` o `"deriva-no-evaluable"`: BR-17 dice que la deriva se muestra tal cual
// salga, y un veredicto incómodo NO es un fallo de la operación.
func postTraerCanonico(svc *usecase.MarketplaceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		var body postTraerBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		res, err := svc.Traer(r.Context(), r.PathValue("nombre"), body.Entrada)
		if err != nil {
			escribirErrorTraer(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}

// escribirErrorTraer mapea los centinelas de Traer a su status (tabla de §13.8). Cada código dice
// algo DISTINTO y el FE los usa distinto — por eso no hay un `default: 500` genérico para todo.
func escribirErrorTraer(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrMarketplaceNoConocido):
		writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
	case errors.Is(err, domain.ErrTraerClaseReferencia),
		errors.Is(err, domain.ErrTraerSourceNoMaterializable),
		errors.Is(err, domain.ErrTraerDestinoEscapa),
		errors.Is(err, usecase.ErrEntradaNoEnCatalogo):
		// Precondiciones del pedido: nada se intentó.
		writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
	case errors.Is(err, usecase.ErrTraerDestinoPoblado):
		writeJSON(w, http.StatusConflict, conflictoTraerBody{Error: err.Error(), Destino: destinoDelError(err)})
	case errors.Is(err, usecase.ErrTraerSinAuth), errors.Is(err, usecase.ErrTraerSinMaterializador),
		errors.Is(err, usecase.ErrSinViaDeLectura):
		// «No puedo mirar» NUNCA se pinta como «tu pedido está mal».
		writeJSON(w, http.StatusServiceUnavailable, errorBody{Error: err.Error()})
	case errors.Is(err, usecase.ErrTraerSHANoCoincide), errors.Is(err, usecase.ErrTraerRemotoNoTiene),
		errors.Is(err, usecase.ErrNoEsMarketplace):
		// «Miré, y lo que vino no es lo declarado»: el estante mintió, no el cliente.
		writeJSON(w, http.StatusBadGateway, errorBody{Error: err.Error()})
	default:
		// Fallo NUESTRO: disco, permisos, rename, Upsert.
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
	}
}

// destinoDelError extrae el destino del motivo del 409 (el usecase lo pone al final del texto).
// Si no se puede aislar, se devuelve "" — la UI degrada a mostrar solo el motivo, nunca inventa
// una ruta.
func destinoDelError(err error) string {
	msg := err.Error()
	const marca = "no se pisa el canónico: "
	if i := indiceDe(msg, marca); i >= 0 {
		return msg[i+len(marca):]
	}
	return ""
}

// indiceDe es strings.Index sin importar strings en este archivo (una sola llamada).
func indiceDe(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// statusDeLectura mapea los errores de LECTURA/registro a su status: 400 «tu url no sirve» vs
// 503 «no puedo mirar» (§7.2 — el FE los usa distinto y jamás insinúa que la url esté mal).
func statusDeLectura(err error) int {
	switch {
	case errors.Is(err, usecase.ErrSinViaDeLectura):
		return http.StatusServiceUnavailable
	case errors.Is(err, usecase.ErrMarketplaceNoConocido):
		return http.StatusNotFound
	default:
		// ErrURLNoCanonicalizable · ErrNoEsMarketplace · ErrClaseInvalida · JSON inválido.
		return http.StatusBadRequest
	}
}

// candidatosOrigenResponse es el wire de GET …/origen/candidatos. `actual` = el home ya declarado
// ("" si la identidad sigue provisional), para que el diálogo pueda decir «ya tiene origen».
type candidatosOrigenResponse struct {
	Candidatos []usecase.CandidatoOrigen `json:"candidatos"`
	Actual     string                    `json:"actual"`
}

// getCandidatosOrigen — GET /api/portafolio/arneses/{clave}/origen/candidatos (S7). El orden lo
// arma el backend por señales BLANDAS; NADA premarcado (BR-11).
func getCandidatosOrigen(svc *usecase.MarketplaceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc == nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "marketplaces: servicio no cableado"})
			return
		}
		candidatos, actual, err := svc.CandidatosDeOrigen(r.Context(), r.PathValue("clave"))
		if err != nil {
			if errors.Is(err, usecase.ErrObservarClaveNoEncontrada) {
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		if candidatos == nil {
			candidatos = []usecase.CandidatoOrigen{}
		}
		writeJSON(w, http.StatusOK, candidatosOrigenResponse{Candidatos: candidatos, Actual: actual})
	}
}

// postAsignarOrigenBody is the POST /api/portafolio/arneses/{clave}/origen payload.
// `sin_origen: true` es la opción «ninguno — dejarlo sin origen» (E-26): NO inventa un home.
type postAsignarOrigenBody struct {
	Home      string `json:"home,omitempty"`
	SinOrigen bool   `json:"sin_origen,omitempty"`
}

// postAsignarOrigen — POST /api/portafolio/arneses/{clave}/origen (S7). **No hay 2xx que clone o
// instale nada** (BR-11): el único efecto secundario es la re-evaluación de deriva, que es una
// lectura.
func postAsignarOrigen(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postAsignarOrigenBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		if body.Home == "" && !body.SinOrigen {
			writeJSON(w, http.StatusBadRequest, errorBody{
				Error: "portafolio: el body debe traer `home` o `sin_origen: true` (elegir «ninguno» es explícito, no un default)",
			})
			return
		}
		entrada, err := svc.AsignarOrigen(r.Context(), r.PathValue("clave"), body.Home)
		if err != nil {
			switch {
			case errors.Is(err, usecase.ErrObservarClaveNoEncontrada):
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrAsignarOrigenColisiona):
				writeJSON(w, http.StatusConflict, errorBody{Error: err.Error()})
			default:
				writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			}
			return
		}
		writeJSON(w, http.StatusOK, entradaWire{Clave: entrada.Identidad.Clave(), EntradaPortafolio: entrada})
	}
}
