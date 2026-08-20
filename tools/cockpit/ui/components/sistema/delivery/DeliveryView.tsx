'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { Rocket, Play, ShieldCheck, FolderOpen, RefreshCw } from 'lucide-react';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner } from '@/components/ui/Spinner';
import { Card } from '@/components/ui/Card';
import {
  ApiClientError,
  getDeliverySalud,
  getDeliveryPlantilla,
  crearDeliverySesion,
  getDeliverySesion,
  listStories,
  type StoryWithArchive,
} from '@/lib/api-client';
import { useSistema } from '@/components/providers/SistemaProvider';
import type {
  DeliverySalud,
  DeliveryPlantilla,
  DeliverySesion,
  ProcesoBinding,
} from '@/lib/types';

/**
 * Cockpit de delivery (F5 · DH-08 · SPEC specs/delivery-cockpit.md RN-41/RN-42).
 *
 * La tesis del norte v3 en una pantalla: el descriptor (capa 1) PARAMETRIZA la sesión
 * de agente (capa 3) con el contexto as-code (capa 2) inyectado. El prompt llega
 * PRECARGADO desde `/api/delivery/plantilla`; el humano ENCAUSA (edita/aprueba) y
 * lanza. CERO literales de estado en este archivo (RN-41): tramo, dueños, gates y
 * checklist vienen del dato. La sesión NO transiciona stories (RN-31) — el ciclo
 * sigue siendo de los skills; la salida la revisa y aplica un humano.
 */

function BindingChip({ b }: { b: ProcesoBinding }) {
  return (
    <span className="inline-flex items-baseline gap-1 rounded bg-[var(--color-panel)] border border-[var(--color-border)] px-1.5 py-0.5 text-[11px]">
      <code className="text-[var(--color-accent)]">{b.arnes ?? b.rol}</code>
      <span className="text-[var(--color-text-dim)]">· {b.rol}</span>
      {b.nota ? <span className="text-[var(--color-text-dim)]">({b.nota})</span> : null}
    </span>
  );
}

const ESTADO_SESION_CLS: Record<string, string> = {
  creada: 'bg-[#1e3a5f] text-[#93c5fd]',
  corriendo: 'bg-[#713f12] text-[#fbbf24] animate-pulse',
  terminada: 'bg-[#14532d] text-[#86efac]',
  error: 'bg-[#450a0a] text-[#fca5a5]',
};

const EVENTO_ICONO: Record<string, string> = {
  init: '◈',
  texto: '¶',
  herramienta: '⚙',
  resultado: '■',
  error: '✗',
};

export function DeliveryView() {
  const { sistema } = useSistema();

  const [salud, setSalud] = useState<DeliverySalud | null>(null);
  const [stories, setStories] = useState<StoryWithArchive[]>([]);
  const [storyId, setStoryId] = useState<string>('');
  const [contextoId, setContextoId] = useState<string>('');
  const [plantilla, setPlantilla] = useState<DeliveryPlantilla | null>(null);
  const [plantillaError, setPlantillaError] = useState<string | null>(null);
  const [prompt, setPrompt] = useState('');
  const [lanzando, setLanzando] = useState(false);
  const [sesion, setSesion] = useState<DeliverySesion | null>(null);
  const [error, setError] = useState<string | null>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // salud del sidecar + stories del sistema activo
  useEffect(() => {
    getDeliverySalud().then(setSalud).catch((e: unknown) =>
      setError(e instanceof Error ? e.message : String(e))
    );
    listStories(sistema)
      .then((s) => setStories(s.filter((st) => !st.is_archived)))
      .catch((e: unknown) => setError(e instanceof Error ? e.message : String(e)));
    setStoryId('');
    setPlantilla(null);
    setPlantillaError(null);
    setSesion(null);
  }, [sistema]);

  // Preselección por query param (?story= — RN-53, botón del story-drawer). Leída
  // una vez al montar (window, no useSearchParams: static export sin Suspense).
  useEffect(() => {
    const q = new URLSearchParams(window.location.search).get('story');
    if (q) setStoryId(q);
  }, []);

  // plantilla parametrizada (RN-39): el descriptor manda, el humano encausa
  useEffect(() => {
    if (!storyId) {
      setPlantilla(null);
      setPlantillaError(null);
      return;
    }
    getDeliveryPlantilla(sistema, storyId, contextoId || undefined)
      .then((pl) => {
        setPlantilla(pl);
        setPlantillaError(null);
        setPrompt(pl.prompt);
      })
      .catch((e: unknown) => {
        setPlantilla(null);
        setPlantillaError(
          e instanceof ApiClientError ? e.message : e instanceof Error ? e.message : String(e)
        );
      });
  }, [sistema, storyId, contextoId]);

  const detenerPoll = useCallback(() => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }, []);

  // eventos por polling ~1.5s mientras corre (RN-42)
  const seguirSesion = useCallback(
    (id: string) => {
      detenerPoll();
      const tick = () =>
        getDeliverySesion(id)
          .then((s) => {
            setSesion(s);
            if (s.estado === 'terminada' || s.estado === 'error') detenerPoll();
          })
          .catch(() => detenerPoll());
      tick();
      pollRef.current = setInterval(tick, 1500);
    },
    [detenerPoll]
  );

  useEffect(() => detenerPoll, [detenerPoll]);

  const lanzar = () => {
    if (!plantilla || !prompt.trim()) return;
    setLanzando(true);
    setError(null);
    crearDeliverySesion({ sistema, story: storyId, prompt, contexto: contextoId || undefined })
      .then(({ id }) => seguirSesion(id))
      .catch((e: unknown) => setError(e instanceof Error ? e.message : String(e)))
      .finally(() => setLanzando(false));
  };

  if (error && !salud) return <ErrorBanner message={error} />;
  if (!salud) return <Spinner />;

  const disponible = salud.disponible && salud.auth !== 'ninguna';

  return (
    <div className="p-4 space-y-4 max-w-5xl">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-sm font-semibold flex items-center gap-2">
            <Rocket size={16} className="text-[var(--color-accent)]" />
            Delivery — sesiones de agente
            <span className="text-[var(--color-text-dim)] font-normal">powered by Claude</span>
          </h1>
          <p className="text-[11px] text-[var(--color-text-dim)] mt-1 max-w-2xl">
            El proceso parametriza la sesión: estado → tramo siguiente → arnés dueño → gate.
            La plantilla propone, tú encausas y lanzas. La sesión trabaja en un workspace
            aislado — no transiciona estados ni toca el repo; su salida la revisas y aplicas tú.
          </p>
        </div>
        <span
          className={cn(
            'text-[10px] px-2 py-1 rounded border shrink-0',
            disponible
              ? 'border-[var(--color-border)] text-[var(--color-text-dim)]'
              : 'border-amber-700/50 bg-amber-900/30 text-amber-300'
          )}
        >
          {salud.disponible
            ? `sidecar arriba · auth: ${salud.auth}`
            : `delivery no disponible — ${salud.motivo ?? 'sidecar caído'}`}
        </span>
      </div>

      {error && <ErrorBanner message={error} />}

      {/* 1 · la story */}
      <Card>
        <div className="flex items-center gap-3 flex-wrap">
          <label className="text-[11px] text-[var(--color-text-dim)]">Story</label>
          <select
            value={storyId}
            onChange={(e) => {
              setStoryId(e.target.value);
              setContextoId('');
              setSesion(null);
            }}
            className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded px-2 py-1 text-xs min-w-64"
          >
            <option value="">— elegir story —</option>
            {stories.map((s) => (
              <option key={s.story_id} value={s.story_id}>
                {s.story_id} · {s.state}
              </option>
            ))}
          </select>
          {plantilla && (
            <>
              <label className="text-[11px] text-[var(--color-text-dim)] ml-2">Contexto as-code</label>
              <select
                value={contextoId || plantilla.contexto.id}
                onChange={(e) => setContextoId(e.target.value)}
                className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded px-2 py-1 text-xs"
              >
                {plantilla.contextos_disponibles.map((id) => (
                  <option key={id} value={id}>
                    {id}
                  </option>
                ))}
                {plantilla.contextos_disponibles.length === 0 && (
                  <option value="">sin arquitectura/visión declaradas</option>
                )}
              </select>
            </>
          )}
        </div>
        {plantillaError && (
          <p className="text-[11px] text-amber-300 mt-2">{plantillaError}</p>
        )}
      </Card>

      {/* 2 · el tramo que el descriptor dicta (RN-39 — nada de esto es hardcode) */}
      {plantilla && (
        <Card>
          <div className="flex items-center gap-2 flex-wrap text-xs">
            <span className="text-[var(--color-text-dim)]">Tramo:</span>
            <code>{plantilla.tramo.de}</code>
            <span className="text-[var(--color-text-dim)]">→</span>
            <code className="text-[var(--color-accent)]">{plantilla.tramo.a}</code>
            {plantilla.tramo.verbo && (
              <span
                className="rounded bg-[var(--color-panel)] border border-[var(--color-border)] px-1.5 py-0.5 text-[10px] text-[var(--color-text-dim)]"
                title="El evento que este tramo emite (CDEvents — RN-47)"
              >
                ⚡ <code>{plantilla.tramo.verbo}</code>
              </span>
            )}
            <span className="text-[var(--color-text-dim)] ml-3">Dueño del destino:</span>
            {(plantilla.tramo.duenos ?? []).map((b, i) => (
              <BindingChip key={i} b={b} />
            ))}
            {!plantilla.tramo.duenos?.length && (
              <span className="text-[var(--color-text-dim)]">(sin binding declarado)</span>
            )}
          </div>
          {(plantilla.gates ?? []).map((g) => (
            <div key={g.id} className="mt-2 text-[11px]">
              <span className="inline-flex items-center gap-1 text-[var(--color-text)]">
                <ShieldCheck size={12} className="text-[var(--color-accent)]" />
                Gate «{g.nombre}» (autoridad: {g.autoridad.rol})
              </span>
              <ul className="list-disc ml-6 mt-1 text-[var(--color-text-dim)]">
                {g.checklist.map((item, i) => (
                  <li key={i}>{item}</li>
                ))}
              </ul>
            </div>
          ))}
          <div className="mt-2 text-[11px] text-[var(--color-text-dim)]">
            Fuentes inyectadas:{' '}
            {plantilla.fuentes.map((f) => (
              <code key={f.ruta_abs} className="mr-2">
                {f.nombre}
              </code>
            ))}
          </div>
        </Card>
      )}

      {/* 3 · el prompt: precargado, editable — el humano encausa */}
      {plantilla && (
        <Card>
          <div className="flex items-center justify-between mb-2">
            <span className="text-[11px] text-[var(--color-text-dim)]">
              Prompt (precargado desde el descriptor — edítalo antes de lanzar)
            </span>
            <div className="flex gap-2">
              <button
                onClick={() => setPrompt(plantilla.prompt)}
                className="text-[11px] px-2 py-1 rounded border border-[var(--color-border)] text-[var(--color-text-dim)] hover:text-[var(--color-text)] inline-flex items-center gap-1"
              >
                <RefreshCw size={11} /> restaurar plantilla
              </button>
              <button
                onClick={lanzar}
                disabled={!disponible || lanzando || !prompt.trim()}
                className={cn(
                  'text-[11px] px-3 py-1 rounded inline-flex items-center gap-1 font-medium',
                  disponible && !lanzando
                    ? 'bg-[var(--color-accent)] text-black hover:opacity-90'
                    : 'bg-[var(--color-panel)] text-[var(--color-text-dim)] cursor-not-allowed'
                )}
              >
                <Play size={11} /> {lanzando ? 'lanzando…' : 'Lanzar sesión'}
              </button>
            </div>
          </div>
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            rows={14}
            spellCheck={false}
            className="w-full bg-[var(--color-panel)] border border-[var(--color-border)] rounded p-2 text-[11px] font-mono leading-relaxed resize-y"
          />
        </Card>
      )}

      {/* 4 · la sesión en vivo (polling RN-42) */}
      {sesion && (
        <Card>
          <div className="flex items-center gap-2 text-xs flex-wrap">
            <code>{sesion.id}</code>
            <span
              className={cn(
                'inline-block rounded px-1.5 py-0.5 text-[10px] font-medium',
                ESTADO_SESION_CLS[sesion.estado] ?? ''
              )}
            >
              {sesion.estado}
            </span>
            {sesion.meta.tramo && (
              <span className="text-[var(--color-text-dim)]">tramo {sesion.meta.tramo}</span>
            )}
            <span className="text-[var(--color-text-dim)] inline-flex items-center gap-1 ml-auto">
              <FolderOpen size={11} />
              <code className="text-[10px]">{sesion.workspace}</code>
            </span>
          </div>
          <div className="mt-2 max-h-72 overflow-y-auto space-y-1 font-mono text-[11px]">
            {sesion.eventos.map((ev, i) => (
              <div key={i} className="flex gap-2">
                <span className="text-[var(--color-text-dim)] shrink-0">
                  {EVENTO_ICONO[ev.tipo] ?? '·'} {ev.tipo}
                </span>
                <span className="text-[var(--color-text)] break-all">{ev.detalle}</span>
              </div>
            ))}
            {sesion.eventos.length === 0 && (
              <span className="text-[var(--color-text-dim)]">esperando eventos…</span>
            )}
          </div>
          {sesion.resultado && (
            <div className="mt-3 border-t border-[var(--color-border)] pt-2 text-[11px]">
              <p className="text-[var(--color-text)]">{sesion.resultado.resumen}</p>
              <p className="text-[var(--color-text-dim)] mt-1">
                salida/:{' '}
                {sesion.resultado.salida.length > 0 ? (
                  sesion.resultado.salida.map((f) => (
                    <code key={f} className="mr-2">
                      {f}
                    </code>
                  ))
                ) : (
                  <span>(vacía)</span>
                )}
              </p>
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
