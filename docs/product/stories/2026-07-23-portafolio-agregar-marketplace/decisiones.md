# Decisiones — Portafolio · Agregar de marketplace → clonar + mejorar

> `tipo: decisiones` · paquete `2026-07-23-portafolio-agregar-marketplace`. Una entrada por
> decisión conversada, EN EL MISMO TURNO (METODOLOGIA §10). Estado: PENDIENTE-RESOLVER →
> PROPUESTA → FIRMADA. Una entrada en PENDIENTE-RESOLVER es una nota capturada tal cual la dejó
> el operador — todavía NO es una decisión tomada, es el punto de partida de la próxima
> conversación.

## PENDIENTE-01 · Reconciliación proyecto-cargado ↔ marketplace de origen — FIRMADA (2026-07-23)

**Estado: FIRMADA.** Resuelta en conversación nueva (Sonnet 5 propuso, operador confirmó las 4
sub-preguntas sin ajustes). Ya puede arrancar el mockup del paquete.

Cuando se agrega un **proyecto** a ArnesIA (no un arnés suelto — una carpeta con Claude Code
instalado), hay que resolver cómo se accionan los arneses que ya están instalados ahí dentro:

- Al cargar el proyecto se deberían VER los arneses instalados en él.
- Lo más probable: ese arnés instalado NO tiene todavía un match real contra ningún marketplace
  mapeado dentro de ArnesIA. **El arnés siempre debe poder decir de dónde proviene.**
- Para poder hacer ese match, ArnesIA necesita tener los **marketplaces mapeados** (registrados)
  internamente, no solo descubrir arneses sueltos.
- Cuando SÍ hay match: el sistema debe saber, con evidencia (código/mecanismo, no supuesto):
  de qué marketplace viene, qué arnés específico es, en qué versión está, si está desactualizado
  respecto al origen o no.
- Cuando NO hay match: debería poder **crearle un espacio** (mapeo) en un marketplace que el
  operador elija — no queda huérfano sin origen posible.
- **Para qué sirve esto (el objetivo de fondo):** que el flujo de "reparar" funcione end-to-end
  sin que el operador tenga que actualizar el proyecto a mano y pasarle código por git para que
  el usuario final haga pull. En vez de eso: el operador solo entrega **acceso al plugin +
  marketplace**; el usuario final lo instala, sigue el paso a paso indicado, y al terminar tiene
  todo funcionando según lo diseñado — sin intervención manual de código por parte del operador.

**Por qué queda ABIERTA y no se resuelve ahora:** toca el modelo de identidad/origen de Slice 0
(`domain.Arnes`, `(home,id)`, `deriva`) y potencialmente el registro de marketplaces como
entidad propia dentro de ArnesIA — necesita su propia conversación de diseño antes de tocar el
mockup de este paquete, no una resolución apurada en el medio de otra cosa.

**Las 4 sub-preguntas, resueltas:**

1. **Dónde vive el registro** — extiende Portafolio/Slice 0: nueva lista `marketplaces_conocidos`
   en el store del Portafolio (junto a `portafolio.json`). NO entidad top-level nueva. `Registries[]`
   (S0-D3) ya modela el lado "detectado por provenance"; esto agrega el lado "operador" (marketplaces
   que ArnesIA conoce/administra explícitamente).
2. **Matching** — ambos, en orden: manifiesto (`(home,id)` de `plugin.json`) primero (barato,
   determinista); si no hay `home`, cae a comparar `deriva` (hash) contra `home/plugins/<id>/<v>/`
   de los marketplaces conocidos. Reusa el mecanismo collect-all ya construido (S0-D14:
   lock>git-plugin>manifiesto) — no inventa uno nuevo.
3. **Sin match** — manual con sugerencia: se listan marketplaces conocidos ordenados por señales
   blandas (nombre/autor de `plugin.json`), el operador SIEMPRE confirma el destino explícitamente
   — mismo principio anti-drift que el resto del Portafolio (nunca auto-decide, ver S1-D2 colisión
   bare-id).
4. **Flujo "reparar sin pasar código"** — extiende el outcome 5 (Reparar + Backport) ya en
   `BACKLOG.md`, no un 6to outcome. Este PENDIENTE-01 es lo que le faltaba al item 2 (Agregar) para
   que el item 5 (Reparar) tenga con qué reconciliar — cierra ese bloqueo.

**Próximo paso:** arrancar el mockup (punto 1 del flujo del paquete, `INDEX.md`) con este modelo
como insumo.
