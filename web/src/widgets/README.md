# widgets/ (FSD)

Bloques compuestos autónomos del shell: **Command Rail**, **lienzo del Mapa** (React Flow 12),
**dock de conversación** (⌘K), **Portafolio** (Organigrama ↔ Cuadrícula). Cada widget = una
carpeta con su Public API (`index.ts`).

Vacío por ahora. Regla FSD: importa `features/entities/shared`, nunca `pages/app` ni otro `widget`.
Regla taxonomía (canvas⊥chrome): un widget de canvas (Mapa) no importa chrome y viceversa.
