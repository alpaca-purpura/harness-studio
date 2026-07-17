# 00-story.md — Instalador público + licencias por organización

> Owner: `/pm` (a mano — `/po` no forjado en este repo). Lo que sabe el PM del story ANTES de
> refinar. NO es spec ejecutable (eso sería `spec-funcional.md`, cuando se promueva a
> `refining` — ver gates abiertos en `decisiones.md` IL-D6/IL-D7).

---
story_id: instalador-publico-licencias-org
type: service-story
module: self-update        # tentativo, ver IL-D6
capability: ninguna         # nace al construir
links:
  research: "./research.md"
  decisiones: "./decisiones.md"
  checkpoint: "./checkpoint.md"
---

## Job-To-Be-Done

**Como** operador de alpacapurpura
**Quiero** distribuir `arnesia` vía un instalador público (`curl | sh`) que active la
instalación con una key emitida por organización (la organización, a su vez, emite hasta N
sub-keys para sus propios usuarios)
**Para** controlar qué organizaciones/equipos internos pueden instalar y correr la app, sin
depender de builds locales manuales por máquina y sin construir un producto de venta externa.

## Por qué importa

Hoy `arnesia` no tiene ningún release publicado — `.goreleaser.yaml` está configurado pero
nunca se taggeó (cero tags, cero GitHub Releases). La única forma de instalar es
`scripts/bundle.sh`, que exige toolchain completo de desarrollador (go, pnpm, rust/cargo, deps
de Tauri) en cada máquina. Eso no escala más allá de 1 desarrollador construyendo para sí
mismo. En cuanto haya una segunda organización/equipo interno usando ArnesIA, hace falta un
instalador que no dependa de compilar localmente, y una forma de saber/controlar quién activó
qué instalación.

## Outcome esperado

- Un operador de una organización corre `curl -sSL <url> | sh` en Mac, Linux o Windows y recibe
  el binario correcto para su SO/arch (mismos targets que ya build-ea `.goreleaser.yaml`).
- El primer arranque de la app pide una key de activación. Sin key válida, la app no habilita
  su funcionalidad completa (mensaje honesto, no crash silencioso — coherente con la disciplina
  de honestidad §4 del repo).
- El operador de alpacapurpura (yo) genera una key maestra por organización. Esa organización
  puede emitir hasta N sub-keys para sus propios usuarios (mecanismo v1: CLI/script propio
  contra la API de Keygen — ver IL-D5, sin portal web en el primer corte).
- Una key inválida o revocada bloquea el arranque de esa instalación con mensaje claro.

## Antecedentes / Contexto

- `self-update-sin-sudo` (CAP-60) ya resuelve la actualización de una instalación **existente**
  — esta story es el paso previo que falta: la instalación inicial.
- `materializar-doctrina-kit-a-arnesia` (CAP-43) materializa doctrina+kit en `~/.arnesia` de
  forma idempotente — mismo patrón/lugar candidato para materializar el estado de activación.
- Restricción de negocio ya investigada (`hs-research-provisioning-auth-2026`, memoria): la
  prohibición de Anthropic sobre hosted-login/pooled-billing aplica a la **suscripción de
  Claude**, no a licenciar ArnesIA — no es la misma restricción, no aplica acá.
- `vision.md:151-157`: el buyer/JTBD del negocio (empresas compran arneses, ArnesIA es medio de
  producción) no cambia con esta story — confirmado con el operador (IL-D0).

## Out of scope (explícito)

- **Venta de ArnesIA a organizaciones externas** — eso es la «ficha futura» que `vision.md:156`
  deja parqueada explícitamente. Esta story NO es ese pivot.
- **Portal web de auto-gestión de keys** para org-admins — v1 es CLI/script operado por el
  operador (Opción A de `research.md` §5); un portal sería Opción B, story separada si hace
  falta.
- **Features EE de Keygen** (SSO/SAML, audit logs, environments, import/export) — no hay
  necesidad identificada de eso en alcance interno acotado.
- **Feature-gating granular por plan** — v1 es binario: instalación activa o no activa, sin
  niveles de plan/tier.

## Riesgos / Asunciones

- **Riesgo:** `Groups` de Keygen (la primitiva de jerarquía org→N-keys) puede terminar siendo
  EE, no CE — no hay tabla formal de features en la documentación pública, solo evidencia por
  ausencia. **Mitigación:** gate obligatorio antes de `refining` — levantar Keygen CE local y
  probarlo con código real (IL-D7).
- **Riesgo:** si CE no trae imagen Docker/OCI oficial, el deploy del server exige
  build-from-source (Rails + Postgres + Redis). **Mitigación:** no bloquea diseño, solo suma
  esfuerzo de deploy — documentar playbook si aplica.
- **Asunción:** el pipeline de release (`goreleaser`) nunca se ejercitó de punta a punta (nunca
  hubo un tag). Antes de prometer el `curl | sh` público, hace falta un tag de prueba real.
- **Asunción:** el módulo destino en el seam (`self-update` extendido vs. `distribucion` nuevo)
  es una decisión de arquitectura que este paquete deja abierta (IL-D6) — no se resuelve sola.

## Próximo paso

No hay `/po` forjado en este repo — el operador conduce `refining` a mano cuando decida
promover esta idea. Antes de ese salto: resolver IL-D7 (Keygen CE `Groups` en código real) e
IL-D6 (módulo destino). Recién ahí nace `spec-funcional.md` con el modelo cerrado + casuística.
