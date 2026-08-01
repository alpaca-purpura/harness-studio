# chris-input — libro mayor de pedidos del operador

## 2026-07-30 · pregunta fundacional (origen del paquete)

Pedido textual (resumido fiel): «¿Qué es un arnés? ¿Un "plugin" puede tener
múltiples arneses e instalarse todos? Quiero que tengamos claro qué es un arnés
por definición para nosotros.» Ejemplos del operador:
1. Arnés de desarrollo con proceso idea→producción: por seguridad y orden, una
   parte la ven los usuarios; cumplidas las especificaciones la revisa un
   developer que la lleva hasta PR; el pase a producción lo hace otra persona.
2. Un developer hace feature / bugfix / spike — cada una sigue estrictamente un
   proceso distinto (el spike nunca sale a producción). ¿Son arneses distintos?
3. Referencia externa: github.com/Gentleman-Programming/gentle-ai («un capi que
   hizo un plugin con múltiples arneses») — «quiero hacerlo mucho mejor».
4. Orden: primero definir qué es un arnés entre los dos; setear cómo se trabaja
   en DESARROLLO primero; luego generalizar a otros rubros. «ArnesIA es ese
   constructor, creador que mantiene el orden de todo.»

## 2026-08-01 · explicación del operador (round 2 del debate DEF-D2)

«v2 me parece bien pero quiero explicarte más el cómo funcionan las cosas»:
1. **El plugin es POR PROYECTO** — se instala a una carpeta estructurada con
   información (dev: código fuente).
2. Un developer en un proyecto hace **ACTIVIDADES** de su puesto: crear una
   funcionalidad (dado un SPEC) · bugfix · spike técnico · revisar un capability.
   Cada actividad tiene su **proceso/procedimiento** (flujo de pasos), «como un
   arnés propio», pero comparten muchos puntos porque todas son parte de un
   **MACROPROCESO** («Implementar software»).
3. Ejemplo CEO de startup chica: outbound inversores · outbound clientes clave ·
   publicidad pagada. Podría agruparlas en distintos plugins **o en uno solo,
   dependiendo del caso**.
4. Síntesis del operador: «el plugin es un macroproceso acotado a un puesto y que
   trabaja sobre una carpeta específica»; en empresas grandes con roles bien
   definidos «serían las actividades del puesto los "arneses"», e instalaría 1-2
   plugins si cumple 2 roles.

## 2026-08-01 · round 3 — orden: Mapa por actividades + spec al MVP

1. «Afinar cómo se muestra el mapa»: con múltiples procesos/actividades en un
   arnés, **que no se vea saturado** — ver el todo, **seleccionar una actividad y
   ver su proceso** para asegurar un entregable de calidad, pero **todo
   conectado**: hay aspectos que van más allá de un paso y afectan a cómo se hacen
   las cosas a nivel proyecto (p.ej. **arquitectura**).
2. «Abstraete y ponte en todos los escenarios complejos»: definir cómo vamos a
   **crear, visualizar y mantener** arneses con el nuevo esquema.
3. «Hazlo en un spec e intégralo al MVP ahora mismo.»
4. (mid-turn) Para probar: usar los arneses de
   `/home/chalreme/Proyectos/vitalia-app`. «Aparece como si estuvieran en un
   marketplace — **elimina eso**, haz como si no tuvieran marketplace.» Tomar ese
   material y **crear nuestro primer arnés** con toda la lógica nueva, colgarlo
   al marketplace; el plugin se llamará **`developer-vitalia`**.

## 2026-07-30 · respuestas al primer round (AskUserQuestion)

- Definición v1: **No — debatir más** (ve un caso a discutir antes de firmar).
- feature/bugfix/spike como pipelines internos: **Sí** → DEF-D1 FIRMADA.
- Cadena de arneses: **Diseñar ahora** (no diferir) → DEF-D3 directiva.
- Paquete de trabajo: **Sí, abrir** → este paquete.
