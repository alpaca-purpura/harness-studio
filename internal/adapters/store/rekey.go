package store

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// ClaveCalificada resuelve un id de arnés pelado a su clave calificada `(home,id,scope)`.
//
// `ok=false` cuando NO hay exactamente una candidata: entonces la llave se deja como está
// y el motivo viaja al informe. Jamás adivina — elegir una de dos candidatas al azar
// fusionaría dos identidades por coincidencia, que es exactamente lo que la identidad
// calificada existe para impedir.
//
// Se INYECTA porque `store` sólo puede depender del dominio y de los puertos: no importa
// el Portafolio. Lo cablea el composition root, que sí puede.
type ClaveCalificada func(idPelado, cwd string) (clave string, ok bool, motivo string)

// Recalibracion es el rastro auditable de una sesión: qué sesión, de qué llave a cuál, y
// cuando NO se movió, por qué. Se emite una por sesión, se haya movido o no — un silencio
// sería indistinguible de «no la miré».
type Recalibracion struct {
	SesionID string `json:"sesion_id"`
	Antes    string `json:"antes"`
	Despues  string `json:"despues"` // == Antes cuando no se pudo decidir.
	Motivo   string `json:"motivo"`
}

// Movio reporta si la llave efectivamente cambió.
func (r Recalibracion) Movio() bool { return r.Antes != r.Despues }

// reKey recalibra la llave de arnés de cada sesión y devuelve UNA fila por sesión. Sólo
// toca el campo `Arnes`: no mira las conversaciones, no borra nada y no fusiona nada.
//
// Es idempotente: una llave ya calificada la resuelve `clave` con ok=false y el motivo
// `ya-calificada`, así que una segunda corrida no mueve una sola fila.
func reKey(sesiones []domain.Session, clave ClaveCalificada) []Recalibracion {
	if clave == nil {
		return nil
	}
	out := make([]Recalibracion, 0, len(sesiones))
	for i := range sesiones {
		antes := sesiones[i].Arnes
		nueva, ok, motivo := clave(antes, sesiones[i].Cwd)
		r := Recalibracion{SesionID: sesiones[i].ID, Antes: antes, Despues: antes, Motivo: motivo}
		if ok && nueva != "" && nueva != antes {
			sesiones[i].Arnes = nueva
			r.Despues = nueva
		}
		out = append(out, r)
	}
	return out
}

// Recalibrar corre el re-key sobre el registro y lo persiste. Es el modo `--aplicar` del
// comando: existe para el caso en que el Portafolio no estuviera disponible al arrancar
// (todas `sin-candidata`), y se puede volver a correr sin miedo porque es idempotente.
//
// Con `soloReportar` no escribe nada — es el dry-run, y su tabla es lo que el operador lee
// ANTES de aplicar. Esa es la diferencia entre «nada se borra» como afirmación y como algo
// observable.
func (r *Registry) Recalibrar(clave ClaveCalificada, soloReportar bool) ([]Recalibracion, error) {
	if clave == nil {
		return nil, errors.New("store: recalibrar sin resolvedor de llaves: no hay con qué decidir")
	}
	sesiones, err := r.Load(nil) //nolint:staticcheck // Load ignora el ctx: lee un archivo local.
	if err != nil {
		return nil, err
	}
	filas := reKey(sesiones, clave)
	if soloReportar {
		return filas, nil
	}
	movio := false
	for _, f := range filas {
		if f.Movio() {
			movio = true
			break
		}
	}
	if !movio {
		return filas, nil // nada que escribir: idempotencia visible en el disco.
	}
	if err := r.Save(nil, sesiones); err != nil { //nolint:staticcheck // ídem.
		return filas, err
	}
	return filas, nil
}

// RevertirLlaves deshace SÓLO el re-key (procedimiento R2 de CV-D16): reescribe el campo
// `Arnes` de cada sesión con el valor que tenía en el respaldo, indexado por id de sesión,
// y NO toca nada más. Las conversaciones y los turnos que se generaron después de migrar
// se conservan — es lo que hace que «reversible» no sea un eufemismo de «restaurá el
// backup y perdé lo que hiciste desde entonces».
//
// Una sesión que no está en el respaldo (nació después) se deja como está: revertir no
// puede inventarle una llave anterior que nunca tuvo.
func (r *Registry) RevertirLlaves(respaldo []domain.Session) ([]Recalibracion, error) {
	previas := make(map[string]string, len(respaldo))
	for _, s := range respaldo {
		previas[s.ID] = s.Arnes
	}
	sesiones, err := r.Load(nil) //nolint:staticcheck // Load ignora el ctx: lee un archivo local.
	if err != nil {
		return nil, err
	}
	filas := make([]Recalibracion, 0, len(sesiones))
	movio := false
	for i := range sesiones {
		antes := sesiones[i].Arnes
		fila := Recalibracion{SesionID: sesiones[i].ID, Antes: antes, Despues: antes}
		previa, hay := previas[sesiones[i].ID]
		switch {
		case !hay:
			fila.Motivo = "no-estaba-en-el-respaldo"
		case previa == antes:
			fila.Motivo = "sin-cambios"
		default:
			sesiones[i].Arnes = previa
			fila.Despues = previa
			fila.Motivo = "revertida"
			movio = true
		}
		filas = append(filas, fila)
	}
	if !movio {
		return filas, nil
	}
	if err := r.Save(nil, sesiones); err != nil { //nolint:staticcheck // ídem.
		return filas, err
	}
	return filas, nil
}

// LeerRespaldo lee un archivo de respaldo (v1 desnudo o sobre v2) sin migrarlo ni tocarlo.
// Es el insumo de RevertirLlaves.
func LeerRespaldo(ruta string) ([]domain.Session, error) {
	reg := &Registry{path: ruta}
	return reg.Load(nil) //nolint:staticcheck // Load ignora el ctx: lee un archivo local.
}

// InformeRecalibracion es lo que el arranque LOGUEA del re-key (CV-D18). Lleva las filas
// crudas y los tres conteos que el operador necesita para saber qué pasó con SUS sesiones.
type InformeRecalibracion struct {
	Filas []Recalibracion

	// RespaldoEn es la copia previa. Vacío ⟺ no hubo nada que mover, así que no se
	// respaldó: respaldar de gusto llenaría el disco del operador de copias idénticas.
	RespaldoEn string

	// Recalibradas son las que efectivamente movieron su llave.
	Recalibradas int
	// SinCandidata son las que NO se pudieron resolver porque su arnés no está en el
	// Portafolio. Se cuenta aparte y se dice: es el caso que el E2E ya conoce (una de las
	// cinco sesiones del operador), y callarlo lo volvería un defecto mudo.
	SinCandidata int
	// YaCalificadas son las que ya tenían clave calificada. Es lo que hace visible la
	// idempotencia: en el segundo arranque, todas caen acá.
	YaCalificadas int
}

// Hubo reporta si hay algo que decir. Un arranque que no movió nada y no tiene ninguna
// sesión sin candidata no imprime: el silencio es correcto sólo cuando no hubo novedad.
func (i InformeRecalibracion) Hubo() bool { return i.Recalibradas > 0 || i.SinCandidata > 0 }

// RecalibrarAlArrancar es CV-D18: el daemon detecta las llaves a medias y las recalibra
// SOLO, al arrancar, con respaldo previo y sin silencio.
//
// Por qué esto muta datos del operador al arrancar, que es justo lo que la arquitectura del
// paquete quiso evitar: porque CV-D16 está FIRMADA y dice que el mismo paso que estrena el
// esquema versionado re-key las vivas, y el comando manual que la reemplazó dejaba 3 de las
// 5 sesiones del operador invisibles sin avisarle (N-23). Una decisión firmada que no corre
// no está construida. El operador eligió esta vía sabiendo el costo (CV-D18, 2026-07-27).
//
// El costo se paga con tres cosas que NO son opcionales:
//
//  1. **Primero la red, después el cambio.** Si el respaldo falla, no se recalibra nada. El
//     orden no es un detalle de implementación: es la decisión.
//  2. **Nada en silencio.** El informe distingue recalibradas de `sin-candidata`.
//  3. **Idempotente.** Sin nada que mover no se respalda ni se escribe, así que arrancar
//     dos veces no duplica respaldos ni vuelve a tocar el archivo.
//
// Y la reversión sigue existiendo: `RevertirLlaves` + el respaldo que esto deja.
func (r *Registry) RecalibrarAlArrancar(clave ClaveCalificada, sello string) (InformeRecalibracion, error) {
	var inf InformeRecalibracion
	if clave == nil {
		return inf, errors.New("store: recalibrar al arrancar sin resolvedor de llaves: no hay con qué decidir")
	}

	// PASO 1 · en seco. Se mira ANTES de tocar nada: hace falta saber si hay algo que
	// mover para decidir si corresponde respaldar.
	filas, err := r.Recalibrar(clave, true)
	if err != nil {
		return inf, err
	}
	inf.Filas = filas
	for _, f := range filas {
		switch {
		case f.Movio():
			inf.Recalibradas++
		case f.Motivo == "sin-candidata":
			inf.SinCandidata++
		case f.Motivo == "ya-calificada":
			inf.YaCalificadas++
		}
	}
	if inf.Recalibradas == 0 {
		return inf, nil // nada que mover ⇒ ni respaldo ni escritura.
	}

	// PASO 2 · la red. Si esto falla, se sale SIN recalibrar.
	respaldo, err := respaldarLlaves(r.path, sello)
	if err != nil {
		return inf, err
	}
	inf.RespaldoEn = respaldo

	// PASO 3 · recién ahora, el cambio.
	if _, err := r.Recalibrar(clave, false); err != nil {
		return inf, err
	}
	return inf, nil
}

// respaldarLlaves copia el registro vivo a `<ruta>.bak-<sello>` (el nombre que nombran
// CV-D16 y CV-D18). Si ese respaldo YA existe no se pisa: la primera copia es la que tiene
// el estado anterior de verdad, y sobrescribirla con el estado de una segunda corrida
// destruiría justamente aquello a lo que se quiere poder volver.
func respaldarLlaves(ruta, sello string) (string, error) {
	if sello == "" {
		sello = time.Now().UTC().Format("0601021504")
	}
	destino := fmt.Sprintf("%s.bak-%s", ruta, sello)
	// Sólo un ARCHIVO REGULAR cuenta como respaldo ya hecho. Un directorio con ese nombre
	// no es una copia de nada, y darlo por bueno dejaría recalibrar sin red — que es
	// exactamente lo que el orden «primero la red» existe para impedir.
	if fi, err := os.Stat(destino); err == nil && fi.Mode().IsRegular() {
		return destino, nil
	}
	b, err := os.ReadFile(ruta) //nolint:gosec // la ruta del propio registro.
	if err != nil {
		return "", fmt.Errorf("store: respaldo de llaves de %s: %w", ruta, err)
	}
	if err := os.WriteFile(destino, b, 0o600); err != nil { //nolint:gosec // ruta derivada de la del registro.
		return "", fmt.Errorf("store: respaldo de llaves en %s: %w", destino, err)
	}
	return destino, nil
}
