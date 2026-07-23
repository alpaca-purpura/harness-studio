# Decisiones — Portafolio · Agregar de marketplace → clonar + mejorar

> `tipo: decisiones` · paquete `2026-07-23-portafolio-agregar-marketplace`. Una entrada por
> decisión conversada, EN EL MISMO TURNO (METODOLOGIA §10). Estado: PENDIENTE-RESOLVER →
> PROPUESTA → FIRMADA. Una entrada en PENDIENTE-RESOLVER es una nota capturada tal cual la dejó
> el operador — todavía NO es una decisión tomada, es el punto de partida de la próxima
> conversación.

## PENDIENTE-01 · Reconciliación proyecto-cargado ↔ marketplace de origen — ABIERTA (2026-07-23)

**Estado: PENDIENTE-RESOLVER.** Nota textual del operador, a retomar en conversación nueva ANTES
de arrancar el mockup del paquete.

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

**Sub-preguntas que la próxima conversación tiene que cerrar** (según quedaron planteadas, sin
resolver todavía):
1. ¿Dónde vive el registro de "marketplaces mapeados" dentro de ArnesIA — extiende el modelo de
   Portafolio (Slice 0) o es una entidad nueva?
2. Mecanismo de matching arnés-instalado ↔ marketplace: ¿por manifiesto (`plugin.json`/`id`), por
   hash (`deriva`), o ambos?
3. Caso "sin match": ¿el operador elige el marketplace destino manualmente siempre, o hay
   sugerencia automática?
4. El flujo de "reparar sin pasar código" — ¿es una extensión del outcome 5 (Reparar + Backport)
   ya en `BACKLOG.md`, o es un flujo nuevo?
