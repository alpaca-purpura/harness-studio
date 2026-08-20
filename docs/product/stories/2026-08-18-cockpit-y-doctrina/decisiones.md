# Decisiones — cockpit y doctrina (2026-08-18)

> Se escriben en el MISMO turno en que se conversan (METODOLOGIA §10). Cada una dice qué
> se decidió, por qué, y qué se descartó.

## CD-1 · El cockpit se vendorea, no se importa

**Decidido:** copia de `vitalia-app/tools/cockpit/` a `harness-studio/tools/cockpit/`,
manteniendo el module path upstream en `go/go.mod`.

**Por qué:** son dos repos independientes sin monorepo que los una; un import cruzado ataría
el build de ArnesIA a un repo que no controla su ciclo. El costo del vendoring es el drift,
y se mitiga con `README-vendored.md` (commit de origen + tabla de diffs) y con
`vendored_arnesia_test.go`, que falla si un merge borra una adaptación.

**Ratificado por el operador:** sí (plan aprobado). Hereda la decisión R-2 del plan previo.

## CD-2 · Las escrituras sobre `platform` son opt-in por flag

**Decidido:** `writableSistemas()` + flag `-allow-platform-writes` (o
`COCKPIT_ALLOW_PLATFORM_WRITES=1`). Sin el flag, el binario se comporta idéntico a upstream.

**Por qué:** upstream trata `platform` como pseudo-sistema de solo trazabilidad porque allá
la raíz es un agregado sobre sistemas reales. En harness-studio `docs/product/` VIVE en la
raíz: todo el contenido es `platform`, y con la regla upstream el board sería un museo (el
drag daba 400 en `handlers_stories.go:271`). Se hace opt-in y no un cambio de default para
que el drift contra upstream quede acotado a una función.

**Descartado:** crear un subdirectorio `platform/docs/product/` para simular un sistema
real — mover 157 capabilities y 41 paquetes para complacer a una herramienta es la cola
moviendo al perro.

## CD-3 · La UI refleja la escribibilidad, no la deduce

**Decidido:** `GET /api/sistemas` devuelve además `escribibles[]`; la UI lo consume desde
`SistemaProvider` y el board decide con eso.

**Por qué:** al habilitar el backend, el board SEGUÍA bloqueado — `BoardView.tsx` tenía
`isPlatform(sistema)` hardcodeado como «solo lectura». Dos capas deduciendo lo mismo por
su cuenta es una contradicción esperando ocurrir: el usuario veía un candado sin causa. La
autoridad es el servidor, que es quien tiene el flag.

**Efecto lateral bueno:** las tres vistas que repetían la deducción (`RoadmapView`,
`DriftView`, `MapView`) ahora consultan `viewAppliesTo`, la única fuente.

## CD-4 · El PATCH de capabilities se bloquea con 409 explicado

**Decidido:** si el workspace trae `scripts/capabilities_to_yaml.py`, el PATCH devuelve 409
diciendo que las caps son artefactos generados (R4).

**Por qué:** editarlas a mano escribe sobre algo que la próxima regeneración pisa sin
avisar, y su enum propio (`vivo · vivo·nc · parcial · stub`) ni siquiera es el de
`gates.go` (`live/beta/deprecated/sunset`). Sin esta guarda el PATCH caía en «sistema
desconocido: platform» — un 400 que MIENTE sobre la causa.

**Cómo se detecta:** por la presencia del generador en el workspace, no por una lista de
nombres de proyecto. Un workspace sin generador conserva el PATCH.

## CD-5 · El reader de capabilities no fabrica datos

**Decidido:** un `.yaml` que abre `---` y no cierra se parsea como documento YAML completo.

**Por qué:** 3 capabilities del repo tienen esa forma. `parseFrontmatter` devolvía mapa
vacío SIN error y los defensive defaults inventaban una capability inexistente
(`capability_id: ""` + `status: "live"`). **Inventar es peor que fallar** (§4: gris ≠
verde). `scripts/cap_doctor.py::_load` ya trataba ambas formas como válidas — ahora cockpit
y doctor coinciden en qué leen.

**Además:** los 3 archivos se normalizaron a la forma canónica que manda el template. Se
hicieron las dos cosas a propósito: el reader tolerante evita que el bug vuelva con el
próximo archivo mal cerrado; la normalización deja el repo consistente hoy.

## CD-6 · El backfill infiere estado, nunca firma

**Decidido:** `chris_verify.signoff: false` en los 37 checkpoints sembrados, sin excepción,
aunque el PARIDAD diga FIRMADO.

**Por qué:** la firma 🧑‍⚖️ la pone un humano. Deducirla de un grep es exactamente el «pass
fabricado» que la doctrina prohíbe. El `state: done` sí se infiere —es una lectura del
proceso, no una firma— y cada archivo lleva escrita la evidencia que lo sostiene.

**Ratificado por el operador:** sí, con la tabla del dry-run a la vista (reparto:
developed 15 · done 10 · developing 4 · refined 4 · idea 3 · parked 1).

## CD-7 · Nace `forjar-arnes`; el knowhow se abre ANTES de escribir

**Decidido:** skill nueva para crear un arnés completo, y en las tres skills del kit la
lectura del nodo del estándar pasa a ser el paso 1, con la obligación de DECLARAR qué nodos
se abrieron.

**Por qué:** `--add-dir` da acceso, no carga. Sin una instrucción que lo abra, los 138
checks viajaban a cada sesión sin llegar nunca al modelo. Y no había skill que cubriera
crear el manifiesto, el spine, las fases y la Guardia — `forjar-caja` empieza asumiendo que
ya existen.

**Sobre DA-8** («upstream al kit, jamás fork silencioso»): acá no hay fork — `kit/` ES el
kit propio de ArnesIA, así que esto es upstream directo.

**Ratificado por el operador:** sí (opción «skill nueva + cableado explícito»).

## CD-8 · Las tabs sin datos se ocultan

**Decidido:** `nav: [board, roadmap, proceso]`. Fuera `learnings` y `harness`.

**Por qué:** verificado endpoint por endpoint, no supuesto: `docs/learnings/` no existe en
este repo (los aprendizajes viven en `docs/product/ledger/HS-NN.md`) y
`docs/process/harness-backlog.md` tampoco (el backlog es `docs/product/BACKLOG.md`, con
otro formato). Una tab vacía no es neutral: hace parecer roto lo que simplemente no aplica.
Encenderlas exige crear el carril primero — trabajo que no está en el alcance de este
paquete.

## CD-9 · Release minor → v0.8.0

**Decidido:** `bump.sh minor`, no patch.

**Por qué:** el cockpit vendored y `forjar-arnes` son funcionalidad nueva, y
`docs/architecture/conventions/versionado.md` pide minor para eso. Además
`instaladores/v0.7.1/` ya existe y una generación publicada nunca se pisa.

**Ratificado por el operador:** sí.

## CD-10 · Los evals viven en `evals/`, NO en `kit/`

**Decidido:** `evals/forjar-arnes/` en la raíz del repo (casos + runner), con su README.

**Por qué:** `kit/` entra al binario por `//go:embed all:kit` y el daemon está al **94,5 %
del techo de peso**. Los casos y el runner son herramientas de desarrollo: no tienen por qué
viajar a la máquina de cada usuario dentro del ejecutable. `evals/` queda fuera del embed y
fuera del gate R2 (`cmd`/`internal`/`web/src`/`web/src-tauri/src`).

**Por qué eval y no verificación única:** el estándar propio lo pide —`knowledge/elements/
skills.md` L1.6 («construir evals ANTES de documentar… ≥3 escenarios») y el check
`skill-has-evals`, que estaba en warn. Una corrida a mano cierra el hueco de hoy; un eval lo
cierra también para el próximo cambio de la skill.

**Ratificado por el operador:** sí.

## CD-11 · El eval no pasa por el daemon, y lo declara

**Decidido:** el runner replica la inyección (`kit/` + `knowhow/` + `doctrine.md` en un
temp, espejo de `provisioner.go::materialize`) y lanza `claude -p` con los flags de
`SpawnArgs`, con dos divergencias tabuladas en `evals/README.md`: permisos `acceptEdits` en
vez de `default` + prompt-tool stdio, y sin MCP.

**Por qué:** el spawn real manda **cada escritura al Dock a esperar a un humano**. Headless
nadie aprueba y la forja se cuelga — no es una limitación del eval, es el diseño del HITL
funcionando. Lo que el eval mide es la conducta del modelo ante la inyección; que el árbol
materializado sea idéntico al del provisioner ya lo fija `provisioner_test.go`.

**Hallazgo durante la construcción:** la primera versión del detector buscaba el slug en
cualquier `tool_use` y daba falso positivo — un `Read` de `kit/skills/forjar-arnes/SKILL.md`
(que es justo lo que hace `forjar-caja` para orientarse) contaba como haber invocado la
skill. Ahora solo cuentan los `tool_use` de la herramienta `Skill`. Un assert laxo es peor
que no tenerlo: da verde por el motivo equivocado.

## CD-12 · Los carriles del CIL indexan la deuda, no la duplican

**Decidido:** `harness-backlog.md` (L1) y `tech-debt.md` (L3) tabulan ítems que YA vivían en
`docs/product/BACKLOG.md`, con una línea cada uno y un `ref` al lugar donde está explicado.

**Por qué:** el BACKLOG explica cada deuda en párrafos con evidencia; el cockpit necesita
filas parseables para contarlas y pintarlas. Copiar la explicación crearía dos fuentes que
se desincronizan al primer cambio. La tabla es un índice: el detalle sigue teniendo un solo
dueño.

**Los learnings sí son contenido nuevo** (carril L2): son tres aprendizajes de esta pasada
que no estaban escritos en ningún lado.

**Ratificado por el operador:** sí (opción «con la deuda real del repo»).

## Deuda registrada

| # | Qué | Dónde |
|---|---|---|
| D-1 | El re-serializado del frontmatter en cada transición borra los comentarios del checkpoint (yaml.v3 no los preserva). El marcador `backfilled: true` sobrevive; la línea de evidencia no. | `tools/cockpit/go/frontmatter.go::stringifyFrontmatter` |
| ~~D-2~~ | ~~Los carriles `learnings` y `harness` del CIL no existen en este repo (CD-8).~~ **CERRADA 2026-08-19**: se abrieron los tres carriles y se encendieron las tabs (CD-12). Queda registrada como `HB-2` en el propio carril. | `docs/process/harness-backlog.md` |
| D-3 | El primer render del board pide `?sistema=main` (el `FALLBACK_SISTEMA` del provider) y da tres 400 antes de resolver `platform`. Ruidoso en consola, sin efecto funcional. | `ui/components/providers/SistemaProvider.tsx:45` |
