package forja

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// nombreLock es la única pieza sin plantilla que produce el sembrador (junto a los
// .gitkeep): el baseline D8 mínimo del contrato §3.
const nombreLock = "semilla.lock.json"

// Adapter implementa ports.ForjaPort sobre una semilla fs.FS (arnes.yaml + plantillas/).
// cmd lo cablea con la semilla embebida (iofs.Sub(doctrina.Semilla, "semilla")).
type Adapter struct {
	fsys fs.FS
	// ahora inyectable para tests deterministas; default time.Now (la fecha de siembra
	// es un INPUT del render, D12 — mismos inputs ⇒ mismo árbol).
	ahora func() time.Time
}

// New construye el adapter sobre la semilla dada.
func New(fsys fs.FS) *Adapter { return &Adapter{fsys: fsys, ahora: time.Now} }

// pieza es un archivo del árbol del contrato §2 YA RENDERIZADO: ruta relativa a
// `.arnesia/` (slash) + contenido final. El plan completo se arma ANTES de escribir
// nada — si una plantilla no renderiza, la siembra falla sin tocar el disco.
type pieza struct {
	Ruta      string
	Contenido []byte
}

// datosPlantilla es el set de placeholders del contrato §6. Los campos de territorio
// (Territorio/Label/Dimensiones) solo se pueblan al renderizar `territorio-INDEX.md`.
type datosPlantilla struct {
	ProyectoNombre string
	ArnesID        string
	Fecha          string
	Territorio     string
	Label          string
	Dimensiones    []datosDimension
}

// datosDimension es una fila de `{{range .Dimensiones}}` (contrato §6: {ID, Zachman, Label}).
type datosDimension struct {
	ID      string
	Zachman string
	Label   string
}

// lockSemilla es el schema 0 de `semilla.lock.json` (contrato §3). `Archivos` lleva los
// sha256 de lo que ESTA siembra escribió — el baseline por archivo; la EVALUACIÓN de
// deriva contra él es gap declarado Fase 2 (§3): el lock se escribe, no se evalúa.
type lockSemilla struct {
	Schema         int               `json:"schema"`
	ArnesID        string            `json:"arnes_id"`
	VersionPineada int               `json:"version_pineada"`
	Fecha          string            `json:"fecha"`
	Archivos       map[string]string `json:"archivos"`
}

// Sembrar materializa el árbol del contrato §2 bajo dir/.arnesia/ (reglas §4):
// idempotente (lo existente JAMÁS se pisa — viaja en YaExistian), determinista (render
// completo antes de escribir), aditivo puro. Al FINAL escribe `semilla.lock.json` — si la
// siembra falla a medias, no hay lock que mienta «completa».
func (a *Adapter) Sembrar(dir string) (domain.InformeSemilla, error) {
	sem, err := ParseSemilla(a.fsys)
	if err != nil {
		return domain.InformeSemilla{}, err
	}
	datos := datosPlantilla{
		ProyectoNombre: filepath.Base(filepath.Clean(dir)),
		ArnesID:        sem.ArnesID,
		Fecha:          a.ahora().Format("2006-01-02"),
	}
	piezas, err := a.plan(sem, datos)
	if err != nil {
		return domain.InformeSemilla{}, err
	}

	var inf domain.InformeSemilla
	raiz := filepath.Join(dir, ".arnesia")
	hashes := map[string]string{} // ruta rel a .arnesia/ → sha256 de lo escrito AHORA
	for _, p := range piezas {
		destino := filepath.Join(raiz, filepath.FromSlash(p.Ruta))
		if _, serr := os.Lstat(destino); serr == nil {
			inf.YaExistian = append(inf.YaExistian, path.Join(".arnesia", p.Ruta))
			continue
		}
		if merr := os.MkdirAll(filepath.Dir(destino), 0o750); merr != nil {
			return inf, fmt.Errorf("forja: crear %s: %w", filepath.Dir(destino), merr)
		}
		if werr := os.WriteFile(destino, p.Contenido, 0o644); werr != nil { //nolint:gosec // G306: process-as-code del proyecto, versionable por el usuario, no secreto.
			return inf, fmt.Errorf("forja: escribir %s: %w", destino, werr)
		}
		sum := sha256.Sum256(p.Contenido)
		hashes[p.Ruta] = hex.EncodeToString(sum[:])
		inf.Creados = append(inf.Creados, path.Join(".arnesia", p.Ruta))
	}

	// El lock, AL FINAL, y con la misma regla que todo lo demás: uno existente jamás se
	// pisa (re-correr sobre una instalación sana ⇒ todo ya-existia, cero escrituras).
	rutaLock := filepath.Join(raiz, nombreLock)
	if _, serr := os.Lstat(rutaLock); serr == nil {
		inf.YaExistian = append(inf.YaExistian, path.Join(".arnesia", nombreLock))
		return inf, nil
	}
	lock := lockSemilla{
		Schema:         0,
		ArnesID:        sem.ArnesID,
		VersionPineada: sem.Schema,
		Fecha:          datos.Fecha,
		Archivos:       hashes,
	}
	b, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return inf, fmt.Errorf("forja: serializar %s: %w", nombreLock, err)
	}
	if err := os.WriteFile(rutaLock, append(b, '\n'), 0o644); err != nil { //nolint:gosec // G306: baseline público del proyecto, no secreto.
		return inf, fmt.Errorf("forja: escribir %s: %w", rutaLock, err)
	}
	inf.Creados = append(inf.Creados, path.Join(".arnesia", nombreLock))
	return inf, nil
}

// plan renderiza TODAS las piezas del árbol §2 en orden estable: terreno (INDEX + los
// territorios de definición en orden de documento) → product → wip (+ .gitkeep) →
// proceso (spines en orden alfabético de tipo; los pasos en orden de spine).
func (a *Adapter) plan(sem ArnesSemilla, datos datosPlantilla) ([]pieza, error) {
	var out []pieza

	agregar := func(ruta, plantilla string, d datosPlantilla) error {
		contenido, err := a.render(plantilla, d)
		if err != nil {
			return err
		}
		out = append(out, pieza{Ruta: ruta, Contenido: contenido})
		return nil
	}

	if err := agregar("terreno/INDEX.md", "terreno-INDEX.md", datos); err != nil {
		return nil, err
	}
	for _, t := range sem.Territorios {
		if t.Naturaleza != "definicion" {
			continue // WIP es instancia (D15): vive en wip/, no bajo terreno/ (contrato §2).
		}
		d := datos
		d.Territorio = t.ID
		d.Label = t.Label
		for _, dim := range sem.Dimensiones {
			if dim.Territorio == t.ID {
				d.Dimensiones = append(d.Dimensiones, datosDimension{ID: dim.ID, Zachman: dim.Zachman, Label: dim.Label})
			}
		}
		if err := agregar(path.Join("terreno", t.ID, "INDEX.md"), "territorio-INDEX.md", d); err != nil {
			return nil, err
		}
	}
	if err := agregar("product/backlog.md", "product-backlog.md", datos); err != nil {
		return nil, err
	}
	if err := agregar("product/roadmap.md", "product-roadmap.md", datos); err != nil {
		return nil, err
	}
	if err := agregar("product/stories/INDEX.md", "product-stories-INDEX.md", datos); err != nil {
		return nil, err
	}
	if err := agregar("wip/INDEX.md", "wip-INDEX.md", datos); err != nil {
		return nil, err
	}
	// Los .gitkeep no tienen plantilla (contrato §5): carpetas vacías que deben
	// sobrevivir a git.
	out = append(out, pieza{Ruta: "wip/activo/.gitkeep"}, pieza{Ruta: "wip/done/.gitkeep"})

	tipos := make([]string, 0, len(sem.Spines))
	for tipo := range sem.Spines {
		tipos = append(tipos, tipo)
	}
	sort.Strings(tipos)
	for _, tipo := range tipos {
		for _, paso := range sem.Spines[tipo] {
			if paso.Plantilla == "" {
				return nil, fmt.Errorf("forja: spine %q paso %q sin plantilla — el contrato §5 la exige", tipo, paso.Paso)
			}
			if err := agregar(path.Clean(paso.Plantilla), paso.Plantilla, datos); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// render ejecuta la plantilla `plantillas/<nombre>` con el set de placeholders §6.
// missingkey=error: un placeholder fuera del contrato rompe la siembra ANTES de escribir,
// jamás renderiza "<no value>".
func (a *Adapter) render(nombre string, datos datosPlantilla) ([]byte, error) {
	src, err := fs.ReadFile(a.fsys, path.Join("plantillas", nombre))
	if err != nil {
		return nil, fmt.Errorf("forja: semilla sin plantilla %s: %w", nombre, err)
	}
	tpl, err := template.New(nombre).Option("missingkey=error").Parse(string(src))
	if err != nil {
		return nil, fmt.Errorf("forja: plantilla %s ilegible: %w", nombre, err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, datos); err != nil {
		return nil, fmt.Errorf("forja: renderizar %s: %w", nombre, err)
	}
	return buf.Bytes(), nil
}
