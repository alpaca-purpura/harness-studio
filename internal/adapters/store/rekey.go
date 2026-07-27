package store

import (
	"errors"

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
