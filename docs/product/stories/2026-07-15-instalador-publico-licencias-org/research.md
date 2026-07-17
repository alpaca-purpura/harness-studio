---
story_id: 2026-07-15-instalador-publico-licencias-org
state: idea
created: 2026-07-15
researcher: chris + /pm
last_modified: 2026-07-15
research_iterations: 2
decision_pending: true
decision_options: ["refining", "parked", "dropped"]
---

# Research — Instalador público `curl | sh` + licencias por organización

## 1. Problema / Oportunidad

- **Dolor hoy:** `arnesia` no tiene ningún release publicado (`.goreleaser.yaml` configurado,
  cero tags en el repo, `gh release list` vacío). La única vía de instalación es build local
  (`scripts/bundle.sh`), que exige toolchain de desarrollador (go, pnpm, rust/cargo, deps
  webkit de Tauri). No escala a onboarding de más de una persona/organización.
- **Oportunidad (acotada):** un `curl | sh` público que baje el binario correcto por SO/arch, y
  un mecanismo para que **organizaciones internas de alpacapurpura** controlen quién de su
  equipo puede activar/usar la instalación, sin construir un producto de venta externa.
- **NO es** monetización de ArnesIA como producto — `vision.md:156` deja eso parqueado como
  «ficha futura». Confirmado con el operador (2026-07-15): alcance es **interno acotado**.

## 2. Opciones evaluadas — licensing/entitlement + distribución

> No son "competidores" de ArnesIA — son proveedores/patrones para resolver la pieza
> "org compra/tiene 1 key maestra → key maestra emite N sub-keys para usuarios", que el
> binario Go valida en el primer run.

| Opción | Jerarquía org→N-keys | Self-host | Precio a escala chica | Integración Go | Fuente |
|---|---|---|---|---|---|
| **Keygen.sh** | `Group` con `maxUsers`/`maxLicenses`/`maxMachines` — sin self-service admin listo, se arma capa fina propia | **Sí** — CE gratis / EE paga | CE: $0 software · EE: free tier 100 users, std ~$49-399/mes | **SDK Go puro, sin cgo** (`keygen-go`) | keygen.sh/docs, keygen-sh/keygen-api README (verificado en vivo 2 rondas) |
| **Cryptlex** | **Nativo** — entidad `Organization`, rol `organization-admin`, portal self-service, `AllowedUsers` como cap de seats | Solo tier Enterprise | $100-600/mes (org-admin exige Business+) | SDK Go oficial pero **envuelve cgo** — rompe `CGO_ENABLED=0` de `.goreleaser.yaml` | cryptlex.com/pricing + docs (verificado en vivo) |
| **LicenseSpring** | Nativo (Customer→License Manager→Users) | No, SaaS-only | $199-750/mes | SDK Go oficial | docs.licensespring.com |
| **10Duke** | Nativo, el más explícito (`Organization→Entitlement→seats`) | No, hosted single-tenant | desde $199/mes | Sin SDK Go, solo OIDC/REST | 10duke.com/pricing |
| **Lemon Squeezy** | No hay sub-keys reales (1 key/orden, `activation_limit` manual por SKU) | SaaS-only | 5% + $0.50/txn | REST plano, trivial | docs.lemonsqueezy.com |
| **Paddle / Stripe** | Ninguna — puro billing, jerarquía sería 100% custom encima | SaaS-only | 5%+$0.50 / 2.9%+$0.30+fee | SDK oficiales, pero sin concepto de key | paddle.com, docs.stripe.com |
| **DIY (Cloudflare Workers + D1/KV)** | Full custom, a medida exacta | Propio | ~$0-5/mes | Custom, ~10-14 días de build | estimación propia |

**Gap percibido en todas las opciones SaaS-only** (Lemon Squeezy/Paddle/Stripe/LicenseSpring/
10Duke): no dan la jerarquía org→N-keys gratis O no dan self-host — para "interno acotado" sin
presupuesto de licencia recurrente, esto las descarta de entrada.

## 3. Viabilidad técnica

### Stack actual relevante

- Módulos candidatos: `self-update` (CAP-60, lifecycle de actualización — el instalador es el
  paso previo que falta) · posible módulo nuevo `distribucion` (IL-D6, abierta).
- Capabilities afectadas: ninguna existe aún; nace `capabilities/{módulo}/activar-licencia.yaml`
  y similares al construir.
- Dependencias externas nuevas: Keygen CE self-hosted (Rails + Postgres + Redis) + SDK Go
  `keygen-go` + `keygen-sh/machineid` (fingerprint).

### Confirmado en vivo (2 rondas de WebFetch contra fuentes primarias, no memoria de training)

- **Costo CE:** README de `keygen-sh/keygen-api` dice explícito *"free (as in beer) to
  self-host for personal and commercial use"* — no es solo "uso interno", cubre licenciar
  ArnesIA. Licencia real del repo = Fair Core License 1.0 (FCL) con conversión automática a
  Apache 2.0 a los 2 años; la única restricción es no ofrecer el propio Keygen como servicio
  competidor — no aplica a nuestro caso (licenciamos ArnesIA, no revendemos Keygen).
- **EE-gated (pago):** request logs, event logs, environments, permisos granulares,
  import/export, **imágenes OCI/Docker**, SSO/SAML. `Groups` (la feature que necesitamos para
  jerarquía org→N-keys) **no aparece** en esa lista — indicio de que vive en CE, pero no hay
  tabla formal de features CE-vs-EE en la documentación pública.
- **Cross-platform cliente:** `keygen-go` es SDK Go puro (sin cgo) → cross-compila a los mismos
  targets que ya declara `.goreleaser.yaml` (`goos: [linux, darwin, windows]`, `goarch: [amd64,
  arm64]`, `CGO_ENABLED=0`). Fingerprint de máquina vía `keygen-sh/machineid`, confirmado
  cross-platform (usa el GUID nativo de cada SO).

### Gaps técnicos a resolver

- [ ] Confirmar que `Groups`/`maxLicenses` funciona en CE corriendo el server local (no solo
  leyendo docs) — bloquea `refining`.
- [ ] Si CE no trae imagen Docker/OCI oficial, deploy del server exige build-from-source (Rails
  + Postgres + Redis) — sumar a estimate de "runtime infra", no es bloqueante de diseño.
- [ ] Diseñar la "capa fina" de auto-gestión de sub-keys: Keygen no trae portal self-service
  (a diferencia de Cryptlex) — decidir si v1 es un endpoint/CLI propio o un script one-shot.
- [ ] Primer tag semver real: nunca se taggeó el repo. El `curl|sh` depende de que exista al
  menos 1 GitHub Release con binarios adjuntos (`goreleaser release`, no solo `--snapshot`).

### Riesgos técnicos

| Riesgo | Probabilidad | Impacto | Mitigación |
|---|---|---|---|
| `Groups` termina siendo EE, no CE | Media (no confirmado en código, solo por ausencia en la lista EE) | Alto — tira la recomendación completa | Levantar Keygen CE local YA, antes de comprometer diseño (gate de `refining`) |
| Docker/OCI EE-gated obliga build-from-source del server | Media | Bajo — solo effort de deploy, no de diseño | Documentar playbook de deploy manual si aplica |
| Pipeline de release nunca ejercitado end-to-end | Alta (nunca se corrió) | Medio | Primer tag de prueba (`v0.0.1-alpha`) antes de prometer el `curl\|sh` |

## 4. Costo estimado

### Build (one-time, orden de magnitud — sin desglosar por rol, no hay `/architect` forjado aún)

- Server Keygen CE self-hosted (deploy + hardening básico): ~2-3 días
- Capa fina org→N-sub-keys (endpoint/CLI, sin portal): ~1-2 días
- Integración `keygen-go` en `arnesia` (activación primer-run + validación offline periódica):
  ~2-3 días
- Primer pipeline de release real (tag → goreleaser → GH Release) + `install.sh` público
  hosteado: ~1-2 días

### Runtime (recurring)

| Componente | Costo |
|---|---|
| Server Keygen CE (VPS chico, Postgres+Redis+Rails) | ~$5-10/mes |
| Licencia Keygen | $0 (CE) |
| Hosting `install.sh` (Cloudflare Pages/Workers free tier) | $0 |

No aplica break-even — no es producto vendido, es costo de infraestructura interna.

## 5. Cómo lo haríamos

### Hipótesis driver

- **H1:** el binario se mantiene público (nada que gatear en la descarga) — la key se pide
  recién en el primer arranque (activación), que es el patrón estándar de la industria y el más
  barato de integrar.
- **H2:** reusamos el patrón conductor/provisioning ya existente (`internal/adapters/provision`)
  para materializar la activación en `~/.arnesia`, igual que hoy materializa doctrina+kit
  (CAP-43) — mismo lugar, no una superficie nueva de storage.

### Opciones de solución

#### Opción A — minimalista (recomendada para v1)
- Keygen CE self-hosted (sin Docker si hace falta, build-from-source) + `Group` por
  organización + script CLI propio (no portal web) para que el operador emita sub-keys.
  Activación en primer run del binario, validación offline con reverify periódico.
- Tradeoff: sin self-service para el admin de la org (lo emite el operador a mano vía CLI) —
  aceptable al alcance "interno acotado", pocas orgs.
- Estimado: ~6-10 días.

#### Opción B — completa
- Igual a A + endpoint HTTP propio para que cada org-admin emita sus propias sub-keys sin
  pedirle al operador (capa fina de verdad, con su propia auth).
- Tradeoff: más superficie a mantener (auth de org-admins, UI o CLI distribuido).
- Estimado: ~12-15 días.

#### Opción C — Cryptlex en vez de Keygen
- Si al levantar Keygen CE local `Groups` resulta EE-gated, pivotar a Cryptlex Business
  ($600/mes) que sí trae org-admin self-service out-of-the-box, aceptando el costo recurrente y
  resolviendo el problema de cgo con un build wrapper aparte (o CGO_ENABLED=1 solo para ese
  binario).
- Tradeoff: costo recurrente real + complica el build actual `CGO_ENABLED=0`.
- Estimado: ~4-6 días de integración + $600/mes.

## 6. Mockups / Prototipos

Ninguno aún — esta pieza es infra/backend (instalador + servidor de licencias), sin superficie
de UI nueva en v1 (Opción A). Si se promueve a `refining` y se elige Opción B, sí va a necesitar
mockup del flujo de activación (pantalla "pegá tu key" en el primer run de la app Tauri) — leer
`mockups/INDEX.md` antes de forkear cuando llegue ese momento.

## 7. Sources / referencias externas

- [Keygen pricing](https://keygen.sh/pricing/) — tiers CE/EE, trial 30 días EE.
- [keygen-sh/keygen-api LICENSE](https://github.com/keygen-sh/keygen-api/blob/master/LICENSE.md) — Fair Core License 1.0, conversión a Apache 2.0 a 2 años.
- [keygen-sh/keygen-api README](https://github.com/keygen-sh/keygen-api) — distinción explícita CE (free, personal+commercial) vs EE (features listadas, requiere key).
- [Keygen Groups API docs](https://keygen.sh/docs/api/groups/) — `maxUsers`/`maxLicenses`/`maxMachines`, sin gating CE/EE explícito en la página.
- [keygen-sh/keygen-go](https://github.com/keygen-sh/keygen-go) — SDK Go, fingerprint vía `keygen-sh/machineid`, archivos `keygen_windows.go`/`keygen_others.go`.
- [Cryptlex pricing](https://cryptlex.com/pricing) + docs de `Organizations` — org-admin self-service confirmado.
- Interno: `.goreleaser.yaml` (targets ya configurados) · `scripts/bundle.sh` (build local actual) · `docs/product/vision.md:151-157` (buyer/JTBD, restricción de alcance).

## 8. Decisión pendiente

- [ ] **Refinar** (state=idea → refining): una vez confirmado en código real que `Groups` vive
  en CE. Sin eso, cualquier spec se firma sobre una asunción no verificada.
- [ ] **Park:** si al validar CE resulta que `Groups` es EE y $600/mes de Cryptlex no se
  justifica para "interno acotado, pocas orgs" — reconsiderar si vale la pena vs. algo más
  simple (ver IL-D8 en `decisiones.md`).
- [ ] **Drop:** no evaluado — no hay razón identificada para descartar el problema en sí
  (la falta de instalador público es un dolor real, confirmado).

## 9. Bitácora

- 2026-07-14 — Conversación arranca preguntando "comando para instalar la última versión de
  arnesia" → se descubre que no hay ningún release publicado.
- 2026-07-15 — Exploración de alternativas de licensing (comparación 8 opciones, subagente
  research + 2 rondas de verificación WebFetch propia) → recomendación Keygen CE self-hosted.
  Gate de alcance con el operador: **interno acotado**, no pivot de negocio. Nace este paquete.
- 2026-07-15 — Makefile nuevo (`make installer`) para generar instaladores locales
  versionados en `instaladores/vX.Y.Z/`; convención de versión decidida en coherencia con el
  modelo de Keygen (semver plano sin prefijo `v`, canales futuros = vocabulario Keygen) — ver
  IL-D9 en `decisiones.md`.
