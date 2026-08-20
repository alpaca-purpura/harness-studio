/**
 * Cliente fetch tipado para la API interna del cockpit.
 *
 * Convenciones:
 *   - Todos los endpoints son same-origin (Next.js API routes).
 *   - Errores HTTP ≥400 lanzan `ApiClientError` con el body parseado.
 *   - Helpers GET/POST/PATCH/PUT/DELETE encapsulan JSON.
 */

import type {
  Story,
  Release,
  Capability,
  OperatorInput,
  SystemMap,
  Ledger,
  StoryState,
  CapChangeType,
  RefType,
  ComputedStatusReport,
  CodeIndexReport,
  BidirectionalValidationReport,
  ActiveSession,
  TorreResponse,
  ProcesoResponse,
  DeliverySalud,
  DeliveryPlantilla,
  DeliverySesion,
  DeliverySesionResumen,
} from '@/lib/types';
import type { HarnessItem } from '@/lib/harness-backlog';

export class ApiClientError extends Error {
  status: number;
  body: unknown;
  constructor(message: string, status: number, body: unknown) {
    super(message);
    this.status = status;
    this.body = body;
  }
}

async function request<T>(input: string, init?: RequestInit): Promise<T> {
  const res = await fetch(input, {
    ...init,
    headers: {
      'content-type': 'application/json',
      ...(init?.headers ?? {}),
    },
  });
  let body: unknown = null;
  try {
    body = await res.json();
  } catch {
    // body vacío · OK
  }
  if (!res.ok) {
    const msg =
      (body as { error?: string } | null)?.error ?? `HTTP ${res.status}`;
    throw new ApiClientError(msg, res.status, body);
  }
  return body as T;
}

// ────────────────────────────────────────────────────────────────────────────
// Stories
// ────────────────────────────────────────────────────────────────────────────

export interface StoryWithArchive extends Story {
  is_archived: boolean;
}

export async function listStories(sistema: string): Promise<StoryWithArchive[]> {
  const data = await request<{ stories: StoryWithArchive[] }>(
    `/api/stories?sistema=${encodeURIComponent(sistema)}`
  );
  return data.stories;
}

export async function getStory(id: string, sistema: string): Promise<Story> {
  const data = await request<{ story: Story }>(
    `/api/stories/${encodeURIComponent(id)}?sistema=${encodeURIComponent(sistema)}`
  );
  return data.story;
}

export async function updateStory(
  id: string,
  sistema: string,
  patch: Partial<Story>
): Promise<Story> {
  const data = await request<{ story: Story }>(
    `/api/stories/${encodeURIComponent(id)}?sistema=${encodeURIComponent(sistema)}`,
    { method: 'PATCH', body: JSON.stringify({ fields: patch }) }
  );
  return data.story;
}

// ────────────────────────────────────────────────────────────────────────────
// Active sessions (build-claims · ADR-009 single-hub worktree)
// ────────────────────────────────────────────────────────────────────────────

export interface SessionsResponse {
  sessions: ActiveSession[];
  by_story: Record<string, ActiveSession>;
}

/** Sesiones Claude/opencode vivas trabajando sobre el hub (lee `.session-locks/`). */
export async function listSessions(): Promise<SessionsResponse> {
  return request<SessionsResponse>('/api/sessions');
}

// ────────────────────────────────────────────────────────────────────────────
// Releases
// ────────────────────────────────────────────────────────────────────────────

export async function listReleases(sistema: string): Promise<Release[]> {
  const data = await request<{ releases: Release[] }>(
    `/api/releases?sistema=${encodeURIComponent(sistema)}`
  );
  return data.releases;
}

export interface CreateReleaseInput {
  release_id: string;
  sistema: string;
  name: string;
  description: string;
  target_date?: string | null;
  order?: number;
}

export async function createRelease(input: CreateReleaseInput): Promise<Release> {
  const data = await request<{ release: Release }>('/api/releases', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return data.release;
}

export interface UpdateReleaseInput {
  name?: string;
  description?: string;
  target_date?: string | null;
  order?: number;
  stories?: string[];
}

/** Edita campos editables de un release (bloqueado server-side si está shipped). */
export async function updateRelease(
  releaseId: string,
  sistema: string,
  patch: UpdateReleaseInput
): Promise<Release> {
  const data = await request<{ release: Release }>(
    `/api/releases?id=${encodeURIComponent(releaseId)}&sistema=${encodeURIComponent(sistema)}`,
    { method: 'PUT', body: JSON.stringify(patch) }
  );
  return data.release;
}

// ────────────────────────────────────────────────────────────────────────────
// Capabilities
// ────────────────────────────────────────────────────────────────────────────

export async function listCapabilities(sistema: string): Promise<Capability[]> {
  const data = await request<{ capabilities: Capability[] }>(
    `/api/capabilities?sistema=${encodeURIComponent(sistema)}`
  );
  return data.capabilities;
}

export async function getCapability(
  module: string,
  cap: string,
  sistema: string
): Promise<Capability> {
  const data = await request<{ capability: Capability }>(
    `/api/capabilities/${encodeURIComponent(module)}/${encodeURIComponent(cap)}?sistema=${encodeURIComponent(sistema)}`
  );
  return data.capability;
}

// ────────────────────────────────────────────────────────────────────────────
// operator-input
// ────────────────────────────────────────────────────────────────────────────

export async function getOperatorInput(
  storyId: string,
  sistema: string
): Promise<OperatorInput> {
  const data = await request<{ operatorInput: OperatorInput }>(
    `/api/operator-input/${encodeURIComponent(storyId)}?sistema=${encodeURIComponent(sistema)}`
  );
  return data.operatorInput;
}

export async function appendOperatorInputNote(
  storyId: string,
  sistema: string,
  text: string
): Promise<OperatorInput> {
  const data = await request<{ operatorInput: OperatorInput }>(
    `/api/operator-input/${encodeURIComponent(storyId)}?sistema=${encodeURIComponent(sistema)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ section: 'notes', entry: { text } }),
    }
  );
  return data.operatorInput;
}

export async function appendOperatorInputRef(
  storyId: string,
  sistema: string,
  ref: { type: RefType; value: string; comment?: string }
): Promise<OperatorInput> {
  const data = await request<{ operatorInput: OperatorInput }>(
    `/api/operator-input/${encodeURIComponent(storyId)}?sistema=${encodeURIComponent(sistema)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ section: 'refs', entry: ref }),
    }
  );
  return data.operatorInput;
}

export async function appendOperatorInputConversation(
  storyId: string,
  sistema: string,
  text: string
): Promise<OperatorInput> {
  const data = await request<{ operatorInput: OperatorInput }>(
    `/api/operator-input/${encodeURIComponent(storyId)}?sistema=${encodeURIComponent(sistema)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ section: 'conversation', entry: { author: 'operador', text } }),
    }
  );
  return data.operatorInput;
}

// ────────────────────────────────────────────────────────────────────────────
// File contents
// ────────────────────────────────────────────────────────────────────────────

export async function getFile(relPath: string): Promise<string> {
  const data = await request<{ path: string; content: string }>(
    `/api/file?path=${encodeURIComponent(relPath)}`
  );
  return data.content;
}

export async function putFile(relPath: string, content: string): Promise<void> {
  await request('/api/file', {
    method: 'PUT',
    body: JSON.stringify({ path: relPath, content }),
  });
}

export async function openInEditor(
  absOrRelPath: string
): Promise<{ ok: true; editor?: string }> {
  return await request<{ ok: true; editor?: string }>('/api/open', {
    method: 'POST',
    body: JSON.stringify({ path: absOrRelPath }),
  });
}

// ────────────────────────────────────────────────────────────────────────────
// Harness backlog (read-only · transversal, no sistema-scoped)
// ────────────────────────────────────────────────────────────────────────────

export async function listHarnessItems(): Promise<{
  items: HarnessItem[];
  counts: Record<string, number>;
  source: string;
}> {
  return await request<{
    items: HarnessItem[];
    counts: Record<string, number>;
    source: string;
  }>('/api/harness');
}

// ────────────────────────────────────────────────────────────────────────────
// CIL board (read-only · transversal · 4 carriles del Continuous Improvement Loop)
// ────────────────────────────────────────────────────────────────────────────

import type { TechDebtItem } from '@/lib/tech-debt';

export interface CilBoard {
  l1: { items: HarnessItem[]; counts: Record<string, number>; open: number; source: string };
  l3: { items: TechDebtItem[]; counts: Record<string, number>; open: number; source: string };
  l2: { count: number; source: string; link: string };
  l4: { source: string; link: string; note: string };
}

export async function listCilBoard(): Promise<CilBoard> {
  return await request<CilBoard>('/api/cil');
}

// ────────────────────────────────────────────────────────────────────────────
// Transition
// ────────────────────────────────────────────────────────────────────────────

// ────────────────────────────────────────────────────────────────────────────
// Gherkin discovery — escenarios del spec + estado del gherkin-matrix
// ────────────────────────────────────────────────────────────────────────────

export interface GherkinScenario {
  id: string;
  title: string;
  gherkin: string;
  graders: string[];
  status: string;
  notes: string;
}

export async function getGherkinStatus(
  sistema: string,
  storyId: string
): Promise<{
  scenarios: GherkinScenario[];
  spec_exists: boolean;
  matrix_exists: boolean;
}> {
  return await request(
    `/api/gherkin-status?sistema=${encodeURIComponent(sistema)}&story=${encodeURIComponent(storyId)}`
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Nueva story desde cero (idea) — scaffold checkpoint + operator-input (R4)
// ────────────────────────────────────────────────────────────────────────────

export interface StoryNewInput {
  sistema: string;
  slug: string;
  goal: string;
  release?: string;
  type?: 'feature' | 'bugfix';
}

export async function postStoryNew(input: StoryNewInput): Promise<{
  storyId: string;
  checkpointPath: string;
  operatorInputPath: string;
}> {
  return await request('/api/story/new', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

// ────────────────────────────────────────────────────────────────────────────
// Nav config — tabs visibles/orden por workspace (cockpit.config.yaml)
// ────────────────────────────────────────────────────────────────────────────

export async function getNavConfig(sistema?: string): Promise<{
  nav: string[] | null;
  path: string;
  exists: boolean;
}> {
  // Per-board (I-51): el board activo decide de qué repo se lee el nav. En single-mode
  // (o sin board) el Go cae a getWorkspaceRoot() e ignora el param.
  const qs = sistema ? `?sistema=${encodeURIComponent(sistema)}` : '';
  return await request(`/api/nav-config${qs}`);
}

// ────────────────────────────────────────────────────────────────────────────
// Gate G — signoff del operador (proceso v5)
// ────────────────────────────────────────────────────────────────────────────

export interface OperatorVerifyInput {
  sistema: string;
  story_id: string;
  result: 'SATISFIED' | 'SATISFIED_WITH_FOLLOWUPS' | 'REJECTED';
  notes: string;
}

export async function postOperatorVerify(
  input: OperatorVerifyInput
): Promise<{ story: Story }> {
  return await request<{ story: Story }>('/api/operator-verify', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

// ────────────────────────────────────────────────────────────────────────────
// Cap doctor — gates canónicos G1-G9 (mismo SSoT que el pre-commit)
// ────────────────────────────────────────────────────────────────────────────

export interface CapDoctorGate {
  id: string;
  label: string;
  drift: number;
  total: number;
  details: unknown[];
}

export interface CapDoctorReport {
  sistema: string;
  healthy: boolean;
  total_drift: number;
  hard_enforced?: unknown;
  validated_at?: string | null;
  gates: CapDoctorGate[];
}

export async function getCapDoctor(
  sistema: string
): Promise<{ doctor: CapDoctorReport | null; hint?: string }> {
  return await request<{ doctor: CapDoctorReport | null; hint?: string }>(
    `/api/capabilities/doctor?sistema=${encodeURIComponent(sistema)}`
  );
}

/** Regen on-demand de los índices cap↔código (shellea los scripts canónicos). */
export async function postCapRegen(sistema: string): Promise<{
  ok: boolean;
  hint?: string;
  duration_ms?: number;
  output_tail?: string;
}> {
  return await request('/api/capabilities/regen', {
    method: 'POST',
    body: JSON.stringify({ sistema }),
  });
}

export async function postTransition(
  sistema: string,
  storyId: string,
  targetState: StoryState,
  reason?: string
): Promise<{ ok: true; newState: StoryState; verbo?: string }> {
  // `verbo` = el evento CDEvents que la transición emitió (F6/RN-47 — dato, sin log).
  return await request<{ ok: true; newState: StoryState; verbo?: string }>(
    '/api/transition',
    {
      method: 'POST',
      body: JSON.stringify({ sistema, storyId, targetState, reason }),
    }
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Story creation
// ────────────────────────────────────────────────────────────────────────────

export interface FromDoneInput {
  sistema: string;
  parentStoryId: string;
  newStorySlug: string;
  goal: string;
  release: string;
}

export async function postFromDone(input: FromDoneInput): Promise<{
  storyId: string;
  checkpointPath: string;
  operatorInputPath: string;
}> {
  return await request('/api/from-done', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export interface ExtendCapInput {
  sistema: string;
  parentCap: { module: string; slug: string };
  capChangeType: 'fix' | 'extend' | 'derive';
  newStorySlug: string;
  goal: string;
  release: string;
  derivedName?: string;
}

export async function postExtendCap(input: ExtendCapInput): Promise<{
  storyId: string;
  slug: string;
  checkpointPath: string;
  operatorInputPath: string;
}> {
  return await request('/api/extend-cap', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

// ────────────────────────────────────────────────────────────────────────────
// Merge release
// ────────────────────────────────────────────────────────────────────────────

export interface MergeReleasePlan {
  plan: Array<{
    op: 'archive_story' | 'update_release' | 'generate_release_notes';
    description: string;
    source?: string;
    target?: string;
  }>;
  executed: boolean;
  preview?: boolean;
  note?: string;
  release?: Release;
}

export async function postMergeRelease(
  sistema: string,
  releaseId: string,
  confirmFinal: boolean,
  opts?: { verified?: boolean; verificationNote?: string }
): Promise<MergeReleasePlan> {
  return await request('/api/merge-release', {
    method: 'POST',
    body: JSON.stringify({
      sistema,
      releaseId,
      confirmFinal,
      verified: opts?.verified ?? false,
      verificationNote: opts?.verificationNote,
    }),
  });
}

// ────────────────────────────────────────────────────────────────────────────
// Learnings
// ────────────────────────────────────────────────────────────────────────────

export interface LearningEntry {
  slug: string;
  path: string;
  /** Path relativo al workspace root (para /api/file + openInEditor). */
  rel_path: string;
  /** Dónde vive: sistema · transversal (docs/learnings) · tooling. */
  source: 'sistema' | 'transversal' | 'tooling';
  preview: string;
  title?: string;
  date?: string;
  type?: string;
  sistema?: string;
  promotable?: string;
  applied?: string;
  /** Ciclo de vida derivado: pending | applied | promoted | wont-apply | reference. */
  status: string;
  sistemas_affected?: string[];
  tags?: string[];
}

export type LearningApplied = 'pending' | 'applied' | 'promoted' | 'wont-apply';

/** PATCH del ciclo de vida — escribe `applied:` en el frontmatter del learning. */
export async function patchLearning(
  relPath: string,
  applied: LearningApplied
): Promise<{ ok: boolean }> {
  return await request('/api/learnings', {
    method: 'PATCH',
    body: JSON.stringify({ path: relPath, applied }),
  });
}

export async function listLearnings(sistema: string): Promise<LearningEntry[]> {
  const data = await request<{ learnings: LearningEntry[] }>(
    `/api/learnings?sistema=${encodeURIComponent(sistema)}`
  );
  return data.learnings;
}

// ────────────────────────────────────────────────────────────────────────────
// System Map
// ────────────────────────────────────────────────────────────────────────────

export async function getSystemMap(sistema: string): Promise<SystemMap> {
  const data = await request<{ system_map: SystemMap; path: string }>(
    `/api/system-map?sistema=${encodeURIComponent(sistema)}`
  );
  return { ...data.system_map, _path: data.path };
}

/**
 * Ledger de un sistema NO-SDD (lente "Evolución", I-45). `null` = el sistema no
 * trackea por ledger (no hay docs/product/ledger.yaml) → empty-state honesto.
 */
export async function getLedger(sistema: string): Promise<Ledger | null> {
  const data = await request<{ ledger: Ledger | null; path: string }>(
    `/api/ledger?sistema=${encodeURIComponent(sistema)}`
  );
  return data.ledger ? { ...data.ledger, _path: data.path } : null;
}

// ────────────────────────────────────────────────────────────────────────────
// Torre de Control (F2 · DH-04 · SPEC specs/torre-read-only.md)
// ────────────────────────────────────────────────────────────────────────────

/**
 * La torre completa (RN-15). `medir: 'gate_fabrica'` dispara la re-medición
 * explícita del gate de fábrica (RN-16 — respeta el TTL de 10 min; puede tardar:
 * los checks declarados son caros por diseño).
 */
export async function getTorre(medir?: 'gate_fabrica'): Promise<TorreResponse> {
  return request<TorreResponse>(`/api/torre${medir ? `?medir=${medir}` : ''}`);
}

// ────────────────────────────────────────────────────────────────────────────
// Descriptor de proceso (F4 · I-77 · RN-32)
// ────────────────────────────────────────────────────────────────────────────

export async function getProceso(): Promise<ProcesoResponse> {
  return request<ProcesoResponse>('/api/proceso');
}

// ────────────────────────────────────────────────────────────────────────────
// Cockpit de delivery (F5 · DH-08 · SPEC specs/delivery-cockpit.md)
// ────────────────────────────────────────────────────────────────────────────

/** Salud del sidecar. El 503 honesto (RN-40) se devuelve como dato, no como excepción. */
export async function getDeliverySalud(): Promise<DeliverySalud> {
  try {
    return await request<DeliverySalud>('/api/delivery/salud');
  } catch (e) {
    if (e instanceof ApiClientError && e.status === 503 && e.body) {
      return e.body as DeliverySalud;
    }
    throw e;
  }
}

export async function getDeliveryPlantilla(
  sistema: string,
  story: string,
  contexto?: string
): Promise<DeliveryPlantilla> {
  const params = new URLSearchParams({ sistema, story });
  if (contexto) params.set('contexto', contexto);
  return request<DeliveryPlantilla>(`/api/delivery/plantilla?${params}`);
}

export async function crearDeliverySesion(input: {
  sistema: string;
  story: string;
  prompt: string;
  contexto?: string;
}): Promise<{ id: string; workspace: string }> {
  return request(`/api/delivery/sesiones`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function getDeliverySesion(id: string): Promise<DeliverySesion> {
  return request<DeliverySesion>(`/api/delivery/sesiones/${encodeURIComponent(id)}`);
}

export async function listDeliverySesiones(): Promise<{ sesiones: DeliverySesionResumen[] }> {
  return request(`/api/delivery/sesiones`);
}

// ────────────────────────────────────────────────────────────────────────────
// Value-stream (slot del seam project.config.yaml · F-4)
// ────────────────────────────────────────────────────────────────────────────

import type { ValueStreamStage } from '@/lib/map-zones';

/** Etapas del value-stream del seam. `null` → slot vacío (usar fallback genérico). */
export async function getValueStreamStages(
  sistema: string
): Promise<ValueStreamStage[] | null> {
  try {
    const data = await request<{ stages: ValueStreamStage[] | null }>(
      `/api/value-stream?sistema=${encodeURIComponent(sistema)}`
    );
    return data.stages;
  } catch {
    return null; // sin seam → fallback genérico client-side
  }
}

// ────────────────────────────────────────────────────────────────────────────
// Capability computed status
// ────────────────────────────────────────────────────────────────────────────

export interface CapabilityStatusResponse {
  status: ComputedStatusReport | null;
  path: string;
  sistema: string;
  hint?: string;
}

export async function getCapabilityStatus(
  sistema: string
): Promise<CapabilityStatusResponse> {
  return await request<CapabilityStatusResponse>(
    `/api/capabilities/status?sistema=${encodeURIComponent(sistema)}`
  );
}

export interface CodeIndexResponse {
  index: CodeIndexReport | null;
  path: string;
  sistema: string;
  hint?: string;
}

export async function getCodeIndex(sistema: string): Promise<CodeIndexResponse> {
  return await request<CodeIndexResponse>(
    `/api/capabilities/code-index?sistema=${encodeURIComponent(sistema)}`
  );
}

export interface BidirectionalValidationResponse {
  validation: BidirectionalValidationReport | null;
  path: string;
  sistema: string;
  hint?: string;
}

export async function getBidirectionalValidation(
  sistema: string
): Promise<BidirectionalValidationResponse> {
  return await request<BidirectionalValidationResponse>(
    `/api/capabilities/bidirectional?sistema=${encodeURIComponent(sistema)}`
  );
}

// Re-export type for convenience
export type { CapChangeType };
