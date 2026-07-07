// Package doctrina embeds the as-code doctrine trees — knowledge/ (12 nodos·138
// checks), arch/boundaries + arch/conventions (16·8 nodos·97 checks) and the L0
// contract schemas — INSIDE the arnesia binary: cuerpo ① de los 3 cuerpos (FIRMADO
// 2026-07-07, HS-10). Con esto `arnesia conformance` corre en la máquina de un cliente
// SIN el repo fuente y sin gastar un token de contexto LLM (la doctrina como DATA es
// del motor Go; a las sesiones CC solo viaja el know-how generativo, cuerpo ②).
//
// La directiva embed no puede alcanzar directorios padre, por eso este archivo vive en
// la raíz del módulo (es el único .go de la raíz; el paquete se importa como
// github.com/alpacapurpura/arnesia con nombre local `doctrina`).
package doctrina

import "embed"

// Files carries the check-bearing trees the RulesetPort parses as DATA (knowledge/
// completo + arch/boundaries + arch/conventions) and the JSON Schemas of the L0
// contract (arch/contracts/schema). El build de cada release arrastra la doctrina
// vigente — actualizar el binario ES actualizar el estándar (knowledge/CADENCE.md).
//
//go:embed knowledge arch/boundaries arch/conventions arch/contracts/schema
var Files embed.FS

// Kit carries the maquinaria (cuerpo ②): el plugin CC propio de ArnesIA — overlay
// doctrine.md + skills que encarnan la doctrina (forjar-caja, auditar-arnes). El
// provisioner lo materializa a ~/.arnesia/ y el conductor lo inyecta por flags al
// spawn, session-scoped: jamás se escribe en el árbol del arnés (② ↛ ③, METODOLOGIA §9).
//
// (all: — el layout de plugin CC arranca en `.claude-plugin/`, y go:embed excluye
// dot-dirs sin ese prefijo.)
//
//go:embed all:kit
var Kit embed.FS
