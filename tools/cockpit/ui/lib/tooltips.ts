/**
 * Diccionario central de tooltips para el cockpit.
 *
 * Origen: refactor 2026-05-28 · fuse /functionality into /map + tooltips terms.
 * Doctrine: Spanish neutro, 1-2 líneas, evita jerga sin definir.
 *
 * Uso:
 *   import { Tooltip } from '@/components/ui/Tooltip';
 *   import { TOOLTIPS } from '@/lib/tooltips';
 *   <Tooltip content={TOOLTIPS.scenarios}>...</Tooltip>
 */

export const TOOLTIPS = {
  // ── Cap schema (visibles en CapDrawer + MapView + DriftView) ─────────────
  change_log:
    'Bitácora append-only de cambios sobre la capacidad. Cada entry liga a una story que la modificó.',
  functional_area:
    "Sub-categoría dentro del agente. Ejemplo: 'valeria.agenda' agrupa todo lo de agenda de Valeria.",
  agent_owner:
    'Agente dueño UI/UX de esta capacidad. Define dónde aparece en el shell post-login (las cajas/agentes del SYSTEM-MAP del producto).',
  user_visible:
    "Si es 'true', aparece en el mapa principal del producto. Si 'false', es infra cross-cutting (se muestra con toggle 'mostrar infra').",
  nature:
    'Tipo de capacidad: feature (terminada user-facing), scaffold (estructura técnica), extension-point (la consumen otras caps).',
  parent_cap:
    'Si esta cap fue derivada de otra (cap_change_type: derive), aquí está el slug del padre.',
  derives_capabilities:
    'Caps hijas que fueron creadas como variantes derivadas de esta.',
  superseded_by:
    'Esta cap fue reemplazada/mergeada en otra. Ver el slug indicado para la visión consolidada.',
  architecture_pattern:
    'ADR sistema-local que cementa el patrón técnico de esta cap (ej. un ADR de sub-tabs del shell).',
  hipaa_lite_overlay:
    'Esta cap toca PHI (Protected Health Information). Aplica dual filter tenant+clinic, audit log sync, encryption at-rest.',
  dev_preview:
    'Pointer a cómo navegar a esta capacidad en una sesión local de desarrollo: ruta, pasos para llegar, componente principal, endpoints, test E2E.',
  yaml_ledger:
    "Archivo YAML fuente de verdad de la capacidad. Click 'xed' para abrirlo en tu editor (VS Code o el configurado).",
  frontmatter:
    "Bloque YAML al inicio del archivo (entre los '---') con los metadatos de la capacidad.",
  created_in_story:
    'Story ID que originó esta capability (referencia a su directorio en docs/product/stories/ o docs/archive/).',

  // ── v3.2 bloques (visibles en CapDrawer secciones nuevas) ────────────────
  scenarios:
    'Capacidades user-facing en lenguaje BDD (Given/When/Then). Capturan qué se puede hacer desde el punto de vista del usuario, no del código.',
  edge_cases:
    'Casos borde narrativos: qué pasa fuera del happy path (sin permisos, sin datos, errores de red, etc.).',
  story_spec_ref:
    'Link al archivo 01-spec.md original de la story que introdujo este scenario (trazabilidad spec → cap).',
  e2e_test:
    'Path al spec de Playwright que verifica este scenario end-to-end.',
  access_entry_points:
    'Rutas por donde se accede a la capacidad: URLs de UI, endpoints de API, webhooks o eventos.',
  requires_role:
    'Roles permitidos para acceder. Si el usuario actual NO está en la lista, el sistema bloquea con 403.',
  requires_clinic_scope:
    'Esta entry_point exige scope de clínica activa (HIPAA dual filter: tenant_id + clinic_id en toda query).',
  forbidden_roles:
    'Roles explícitamente prohibidos. Generan audit log si se intenta acceder.',
  business_rules:
    'Invariantes del negocio + compliance (HIPAA, retention, dual filter, etc.). Cada regla tiene severity + paths a las reglas .claude/ que la enforcearon.',
  audit_trail:
    'Esta acción genera un row en la tabla audit_log (inmutable · retention 10 años · regulación LatAm).',
  enforcement:
    'Paths a archivos `.claude/rules/*.md` o ADRs que enforcearon esta regla (auditor + pre-commit hooks).',
  related_capabilities:
    'Grafo de relaciones cap↔cap: caps de las que esta depende, caps que esta habilita, caps similares, caps que esta marca como obsoletas.',
  severity:
    'Nivel de impacto de la regla: critical (rompe compliance) · high (rompe negocio) · medium (rompe UX) · low (recordatorio).',
  code_files:
    'Archivos del código que declaran pertenecer a esta cap mediante header `# cap:` (Python) o `// cap:` (TypeScript). Generado por scripts/generate_code_to_cap_index.py.',
  cap_header:
    'Header al inicio de un archivo de código que declara su capacidad principal. Formato: `# cap: <module>.<slug>` (Python) o `// cap: <module>.<slug>` (TypeScript).',

  // ── Verificación (visibles en DriftView + CapDrawer · MapView badges) ────
  verified_live:
    'Cap declarada live + todos los e2e_test de sus scenarios pasan + los paths de verificación existen en filesystem. La realidad del código matchea la declaración.',
  declared_live:
    "Cap declarada live con scenarios, pero sin evidencia de tests pasando (e2e_test ausentes o sin verificar). La declaración dice 'live' pero no podemos confirmarlo desde el código.",
  drift:
    'Mismatch entre lo declarado y la realidad: archivos no existen, código/tests/acceso contradicen lo declarado, o headers de código apuntan a otra cap.',
  stub:
    "Cap sin scenarios todavía. No describe nada aún · suele ser una cap planeada o un placeholder que migrará a 'live' cuando una story la implemente.",
  partial:
    'Cap con algunos scenarios verificados y otros sin verificar (story en flight).',
  wip:
    'Scenarios declarados pero ninguno verificado todavía · pertenece a una story que aún no se cierra.',
  orphan:
    'Archivo de código sin cap owner principal. Header `# cap: __orphan__`. Candidato a refactor: atar a una cap o eliminar.',
  shared:
    'Archivo cross-cap consumer (consumen múltiples capacidades). Header `# cap: __shared__`. Típico: utilidades, hooks compartidos, Shadcn UI primitives.',
  cross_check_3:
    'Para cada scenarios[].e2e_test declarado, el archivo debe existir y contener un test pattern de Playwright. Detecta tests rotos o nunca creados.',
  cross_check_4:
    'Para cada access.entry_points[].requires_role, los roles declarados deben coincidir con los decorators @require_phi_access del código backend. Detecta drift entre permisos documentados y permisos enforcearon en runtime.',
  hard_fail:
    'El validador encontró drift en cross_checks marcados como HARD (1 y 3). Pre-push hook bloquea el push a main/release.',
  soft_drift:
    'Hay drift pero solo en cross_checks 2 o 4 (SOFT). Pre-push no bloquea, queda como deuda visible en este tab.',
  clean:
    'Cero drift en los 4 cross-checks. La cap y el código están alineados.',
  bidirectional_validation:
    'Verificación cruzada entre lo que las caps declaran y lo que el código de verdad implementa. 4 cross-checks bidireccionales: 2 HARD (bloquean push) y 2 SOFT (advisory).',

  // ── Estados del ciclo: MIGRADO (F6/RN-51) — los tooltips de columna del board
  // salen de `estado.descripcion` del descriptor de proceso, ya no de este diccionario.
  wip_cap:
    'WIP cap: cuántas stories pueden estar simultáneamente en este estado por worktree. Excederlo es advisory para evitar context switching excesivo.',
  cap_change_type:
    'Tipo de cambio que la story aplica al cap target: new (cap nueva), fix (bug), extend (scenarios nuevos), derive (cap hija).',
  defer_audit:
    'Auditoría pausada explícitamente por el operador. Mientras true, /dev-team no auto-handoff, /pm-{sistema} pingea deuda.',
  spawned_by:
    'Quién creó la story: el operador, auto-pm, ux-agentico-spawn, etc.',
  repro_verified:
    'Para hot-fixes: si true, el bug fue reproducido localmente antes de spawn builder (rule R26 hotfix-repro-mandatory).',
  outcome_legacy:
    'DEPRECATED · entidad legacy pre-paradigm-v4. Reemplazada por Release. Migración orgánica.',
  release_concept:
    'Agrupación temporal de stories hacia un milestone (F0, F1, F2...). Las stories se mueven entre releases con drag&drop.',

  // ── Harness Backlog estados (BoardView core/transversal · qué significa + qué hacer) ──
  harness_reported:
    'Recién capturado con /harness-issue, sin clasificar. Acción: ninguna urgente; se evalúa en el próximo lote de triage, donde decides si entra al harness o se descarta.',
  harness_triaged:
    'Ya revisado y dimensionado (severidad, alcance, esfuerzo), pero sin tu aprobación. Acción: revísalo y decide si lo apruebas (pasa a ratified) o lo dejas/descartas.',
  harness_ratified:
    'Aprobado por ti para aplicar, pero todavía sin implementar. Acción: ninguna de tu parte; queda en cola para que Claude lo ejecute en un lote (apply-pipeline).',
  harness_applied:
    'Implementado y commiteado, pero el efecto no se verificó aparte todavía (típico en cambios de docs). Acción: si quieres cerrarlo, pide verificar el efecto (gate, re-lectura o live) para moverlo a verified.',
  harness_verified:
    'Aplicado y con efecto confirmado (gate, live o re-lectura independiente). Acción: ninguna; está cerrado. Es el estado terminal sano del item.',
  harness_deferred:
    'Decidido pero pospuesto a propósito (alto costo, bajo beneficio o requiere sesión dedicada). Acción: ninguna ahora; revísalo si cambian las prioridades o el contexto que lo bloqueaba.',
  harness_otro:
    'Estado fuera del lifecycle del harness (typo o estado nuevo sin registrar). Acción: abre el archivo y corrige cómo está escrito el campo estado de ese item.',

  // ── Cockpit-specific (visibles en headers / breadcrumbs) ─────────────────
  system_map:
    'Archivo YAML manual editado por el operador (tracked) que define la taxonomía completa del producto: agentes × N functional_areas + flows cross-agent + data ownership. El cockpit lo usa como esqueleto del Mapa Implementado.',
  product_health:
    'Distribución honesta de las capabilities por estado de salud (verificado · declarado · parcial · stub · drift). Lee el summary computado de _status-computed.json — refleja la realidad, no la aspiración.',
  mapa_implementado:
    'Vista del producto: caps cementadas agrupadas por agente y functional_area. Lee del SYSTEM-MAP como esqueleto + cruza con las caps reales.',
  backlog_board:
    'Tablero Kanban con stories en cada estado del paradigm v4 (idea → done). Drag&drop para transitions permitidas.',
  roadmap:
    'Vista de releases (F0, F1, F2...) con stories agrupadas por release. Drag&drop para mover stories entre releases.',
  drift_tab:
    'Caps que no están verified-live. Lista priorizada por severidad para que sepas qué arreglar.',
  arquitectura:
    'Vista global del SYSTEM-MAP: agentes + functional_areas + flows cross-agent + data ownership entities.',
  operator_input:
    'Conversación asíncrona operador↔Claude por story. Cada skill appendea verdict (APLICADO/DUDA/REFUTADO/PROPONE) al cierre de su turn.',
  infra_role:
    "Marca caps no user-facing (cross-cutting): observability, platform scaffolding, payment infra. Aparecen solo si activas el toggle 'infra'.",
  planned_status:
    'Functional area declarada en SYSTEM-MAP pero todavía sin caps shipped. Mostrada como placeholder con target_release estimado.',
  v3_2_badge:
    'Esta cap tiene scenarios narrados (v3.2): se puede leer qué hace en lenguaje BDD sin tocar el código.',
  refining_lane:
    'Worktree dedicado para refinamiento ({workspace}-{sistema}-refine). Permite a /pm trabajar specs mientras /dev-team construye otra story en el worktree principal.',
} as const;

export type TooltipKey = keyof typeof TOOLTIPS;
