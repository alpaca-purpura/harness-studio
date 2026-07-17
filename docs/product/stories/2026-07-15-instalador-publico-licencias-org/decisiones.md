# Decisiones — Instalador público + licencias por organización (2026-07-15)

> `tipo: idea/research` (aún no es spike formal ni story `refining`). Cada decisión conversada
> se escribe EN EL MISMO TURNO (§10). Nada firmado 🧑‍⚖️ todavía.

## IL-D0 · Alcance — DECIDIDA (con el operador, `AskUserQuestion`)

- **Interno acotado**, NO pivot de negocio: no es ArnesIA vendida a orgs externas.
- No requiere enmendar `vision.md:156` (*"Si ArnesIA misma se vuelve producto, será ficha
  futura — no es esta visión"*). El buyer/JTBD del negocio (empresas compran arneses, ArnesIA es
  medio de producción) **no cambia**.
- El caso real: alpacapurpura controla qué organizaciones/equipos internos (posibles clientes
  propios, no terceros que compran ArnesIA) pueden instalar y activar la app, con el operador
  emitiendo una key maestra por org y esa org emitiendo hasta N sub-keys para sus usuarios.

## IL-D1 · Punto de gate: activación en primer-run, no en la descarga — DECIDIDA

- El binario/instalador (`curl | sh`) queda **público**, sin key.
- La key se pide y valida en el **primer arranque** del binario (patrón estándar de la
  industria — Docker Desktop, JetBrains, etc.). Más barato de integrar que gatear la descarga
  misma, y no exige que `install.sh` conozca nada de auth.

## IL-D2 · Proveedor de licensing: Keygen.sh (self-hosted CE) — PROPUESTA, no firmada

- Comparación de 8 opciones (research de mercado + 2 rondas de verificación en vivo, ver
  `research.md` §2-3): Keygen es la única con SDK Go sin cgo (crítico — el build actual usa
  `CGO_ENABLED=0`, Cryptlex lo rompería), con validación offline lista (keys firmadas, cache +
  reverify periódico), self-host disponible, y sin costo recurrente en CE.
- Alternativa de respaldo si IL-D6 sale mal: Cryptlex Business ($600/mes) — trae org-admin
  self-service out-of-the-box, pero cuesta y complica el build `CGO_ENABLED=0` (Opción C en
  `research.md` §5).

## IL-D3 · Costo confirmado: Keygen CE = $0 software — HALLAZGO (verificado en vivo)

- README del repo `keygen-sh/keygen-api`: *"free (as in beer) to self-host for personal and
  commercial use"* — no es solo "uso interno", cubre licenciar un producto propio.
- Licencia real = Fair Core License 1.0 (Apache 2.0 automático a los 2 años). La restricción de
  "Competing Use" solo aplica a ofrecer Keygen mismo como servicio competidor — no aplica acá.
- Único costo real: infraestructura propia del server (VPS chico, ~$5-10/mes).

## IL-D4 · Cross-platform cliente confirmado — HALLAZGO (verificado en vivo)

- `keygen-go` (SDK) es Go puro sin cgo → cross-compila a los mismos targets que ya declara
  `.goreleaser.yaml` (`darwin`/`windows`/`linux` × `amd64`/`arm64`), sin tocar el toolchain.
- Fingerprint de máquina vía `keygen-sh/machineid`, confirmado cross-platform (GUID nativo por
  SO — Windows registry, macOS IOPlatformUUID).

## IL-D5 · Modelo org→N-keys: `Group` de Keygen + capa fina propia — PROPUESTA, no firmada

- Keygen no trae portal self-service para que un org-admin emita sus propias sub-keys (a
  diferencia de Cryptlex). El diseño v1 propuesto (Opción A de `research.md` §5) es minimalista:
  el operador emite sub-keys vía CLI/script propio contra la API de Keygen, acotado a pocas
  orgs — no un endpoint HTTP con auth propia para cada org-admin (eso sería Opción B, mayor
  superficie).

## IL-D6 · Módulo destino en el seam — ABIERTA (bloquea `refining`)

- Ningún `domain_module` existente calza limpio: `self-update` (CAP-60) es el vecino más
  cercano (mismo lifecycle: instalar→activar→actualizar) pero hoy solo cubre updates
  post-instalación. `provisioning` (CAP-43/44) materializa doctrina+kit, no identidad/licencia.
- Opciones: (a) extender `self-update` para que cubra también instalación+activación inicial,
  (b) crear módulo nuevo `distribucion` en `project.config.yaml:domain_modules` — decisión de
  arquitectura, requiere `/architect` (no forjado; el operador la toma a mano si se promueve a
  `refining`).
- `checkpoint.md` de este paquete quedó con `module: self-update` como placeholder tentativo,
  NO firmado.

## IL-D7 · Confirmar `Groups` vive en CE, no EE — ABIERTA (bloquea `refining`, la más dura)

- La documentación pública no tiene tabla formal CE-vs-EE. La única evidencia es que `Groups`
  no aparece en la lista explícita de features EE (request/event logs, environments, permisos
  granulares, import/export, OCI/Docker, SSO/SAML) — evidencia por ausencia, no confirmación.
- **Gate obligatorio antes de `refining`:** levantar Keygen CE local y probar `Groups` con
  `maxLicenses` de verdad. Si resulta EE-gated, cae la recomendación completa de IL-D2 y hay que
  reevaluar Opción C (Cryptlex) del `research.md`.

## IL-D8 · ¿Vale la pena si `Groups` es EE? — ABIERTA (contingente a IL-D7)

- Si Cryptlex ($600/mes) es la única vía con org-admin self-service nativo, para "interno
  acotado, pocas orgs" ese costo recurrente puede no justificarse frente a una alternativa aún
  más simple (ej. Keygen CE sin `Groups` — el operador arma la jerarquía a mano con
  `metadata`/`policies` planos por org, sin la primitiva nativa). Decidir recién si IL-D7 sale
  negativo — no especular ahora.

## IL-D9 · Convención de versión para instaladores locales — FIRMADA 🧑‍⚖️ 2026-07-15 (HS-24)

- No bloquea `refining` (es tooling de build local, anterior a `curl | sh` público) pero se
  decide ahora para no rehacerla después: `make installer` (Makefile nuevo en la raíz) bumpea
  el patch (Z de X.Y.Z) en `web/src-tauri/Cargo.toml` (fuente de verdad), sincroniza
  `tauri.conf.json` + `web/package.json`, y copia el resultado de `scripts/bundle.sh` a
  `instaladores/vX.Y.Z/`.
- Verificado en vivo (2 subagentes de research, fuentes primarias): `Release.version` de
  Keygen exige **semver plano sin prefijo `v`** (keygen.sh/docs/api/releases/), y para
  prereleases el `channel` (stable/rc/beta/alpha/dev — 5 valores) **debe matchear el tag de
  prerelease de la versión**. `.goreleaser.yaml` acepta el prefijo `v` en el tag git pero lo
  strippea igual (goreleaser.com/limitations/semver/).
- Decisión: los manifiestos (`Cargo.toml`/`tauri.conf.json`/`package.json`) guardan semver
  plano `X.Y.Z` (compatible 1:1 con Keygen, sin transformación futura); la carpeta de salida
  local usa prefijo `v` (mismo estilo que los tags de goreleaser). No se agrega sufijo de
  canal (`-dev.N`) todavía — recién cuando exista un canal real que loguear en Keygen, y en
  ese momento el vocabulario de canal ya coincide 1:1 con los 5 valores de Keygen (evita
  divergencia terminológica al cablear IL-D5).
- Práctica confirmada en vivo (Tauri v2.tauri.app/reference/config/): si `tauri.conf.json`
  omite `version`, Tauri cae a `Cargo.toml` — se mantiene explícito en los 3 archivos igual
  porque la doc oficial lo recomienda así, y el Makefile los sincroniza para que no haya
  drift (ya existía: `tauri.conf.json` 0.2.0 vs `package.json` 0.1.0 antes de este cambio).
- El Makefile NO commitea ni taggea — el bump queda en el working tree, deliberado y
  revisable por el operador antes de commitear (patrón `goreleaser --snapshot` / builds
  locales de Kubernetes: no mutar/publicar automático fuera de una acción consciente).

## Próximo paso

IL-D9 firmada 🧑‍⚖️ 2026-07-15 (HS-24) — tooling de build local, no bloqueaba `refining` y ya
está vivo (`make installer`, enforcer `TestVersionManifestsInSync`). El resto de la lista sigue
sin firmar. Antes de `refining`: resolver IL-D7 (levantar Keygen CE local, probar `Groups`) e
IL-D6 (con `/architect` o el operador a mano). Recién ahí se escribe `spec-funcional.md` con el
modelo cerrado.
