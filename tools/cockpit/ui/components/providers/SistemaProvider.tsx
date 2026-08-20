'use client';

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { usePathname, useRouter } from 'next/navigation';
import {
  isTwoLevel,
  empresaOf,
  splitSistemaKey,
  sistemasFor,
  parseSistemaSearch,
  buildSistemaSearch,
  resolveSistemaIntent,
} from '@/lib/sistemas';

interface SistemaContextValue {
  /** Sistema activo = key navegable compuesta "empresa/sistema" (contrato hacia
   *  las ~20 vistas sistema-scoped; es el valor de ?sistema= y de resolveSistema). */
  sistema: string;
  setSistema: (key: string) => void;
  /** Keys navegables (planas) — contrato hacia las vistas sistema-scoped. */
  sistemas: string[];
  /** 'multi' si alguna key trae prefijo de empresa ("empresa/sistema"); 'single' si no.
   *  Stage 4 (CK-07): DevHub ya no consume el árbol rico de /api/portfolio (gaps/
   *  procedencia son negocio de Cockpit) — deriva esto de la lista plana. */
  mode: 'single' | 'multi';
  /** Empresa seleccionada en el selector. */
  empresa: string;
  setEmpresa: (slug: string) => void;
  /** Slug del sistema seleccionado en el dropdown dentro de la empresa. */
  sistemaSlug: string;
  setSistemaSlug: (slug: string) => void;
  /** Keys donde el servidor acepta escrituras (GET /api/sistemas → `escribibles`).
   *  La autoridad es el backend, que es quien tiene el flag `-allow-platform-writes`:
   *  la UI refleja este dato en vez de deducirlo, así las dos capas no discrepan. */
  escribibles: string[];
}

const SistemaContext = createContext<SistemaContextValue | null>(null);

const STORAGE_KEY = 'cockpit:sistema';
const FALLBACK_SISTEMA = 'main';

export function SistemaProvider({
  children,
  sistemas: sistemasProp,
  defaultSistema,
}: {
  children: ReactNode;
  sistemas: string[];
  /** Sistema inicial preferido (de `DEFAULT_SISTEMA` env · cockpit-up.sh por-worktree). */
  defaultSistema?: string;
}) {
  // Static export (binario Go): el layout no puede leer el workspace en build-time
  // y pasa sistemas=[] — pedimos la lista plana a GET /api/sistemas (devhub-owned,
  // ver handlers_misc.go). Stage 4 (CK-07): ya no se pide /api/portfolio — ese
  // árbol rico (gaps/procedencia/servicios_compartidos) es negocio de Cockpit.
  const [sistemas, setSistemas] = useState<string[]>(sistemasProp);
  const [remoteDefault, setRemoteDefault] = useState<string | undefined>(undefined);
  // Vacío hasta que responda el server: ante la duda NO se ofrece escribir (el peor
  // fallo sería habilitar un drag que después el backend rechaza con 400).
  const [escribibles, setEscribibles] = useState<string[]>([]);

  // El fetch NO se saltea cuando la lista ya vino por props: `escribibles` solo lo sabe
  // el server (es quien tiene el flag) y en modo dev/SSR la prop no lo trae — sin esta
  // llamada el board quedaría en solo-lectura ahí. Lo que sí se condiciona es PISAR
  // `sistemas`: si vinieron por props, mandan ellas.
  useEffect(() => {
    let alive = true;
    fetch('/api/sistemas')
      .then((r) => (r.ok ? r.json() : null))
      .then(
        (
          data: { sistemas?: string[]; escribibles?: string[]; default_sistema?: string } | null
        ) => {
          if (!alive || !data) return;
          if (Array.isArray(data.escribibles)) setEscribibles(data.escribibles);
          if (sistemasProp.length > 0) return; // dev/SSR: ya vienen por getSelectableSistemas()
          if (!Array.isArray(data.sistemas)) return;
          setSistemas(data.sistemas);
          if (typeof data.default_sistema === 'string' && data.default_sistema) {
            setRemoteDefault(data.default_sistema);
          }
        }
      )
      .catch(() => {});
    return () => {
      alive = false;
    };
  }, [sistemasProp]);

  const mode: 'single' | 'multi' = isTwoLevel(sistemas) ? 'multi' : 'single';

  // Sistema activo inicial: env DEFAULT_SISTEMA (si es una key real presente) → primero detectado → fallback.
  const preferred = defaultSistema ?? remoteDefault;
  const initialSistema =
    preferred && sistemas.includes(preferred)
      ? preferred
      : sistemas[0] ?? FALLBACK_SISTEMA;

  const [sistema, setSistemaState] = useState<string>(() => {
    if (typeof window === 'undefined') return initialSistema;
    return window.localStorage.getItem(STORAGE_KEY) ?? initialSistema;
  });

  // Empresa + slug seleccionados. Arrancan derivados de la key activa; se
  // re-sincronizan cuando la key activa cambia (ej. al elegir un sistema navegable).
  const [empresa, setEmpresaState] = useState<string>(() => empresaOf(sistema));
  const [sistemaSlug, setSistemaSlugState] = useState<string>(
    () => splitSistemaKey(sistema).sistema
  );
  // Deep-link (I-50): el hydration desde la URL aplica el triple (key, empresa,
  // slug) de una — `applyingIntentRef` salta la derivación key→slug UNA vez
  // (solo cuando el hydration movió la key). `hydratedRef` marca que ya leímos
  // la URL (gate del write-effect).
  const applyingIntentRef = useRef(false);
  const hydratedRef = useRef(false);
  useEffect(() => {
    if (applyingIntentRef.current) {
      applyingIntentRef.current = false;
      return;
    }
    setEmpresaState(empresaOf(sistema));
    setSistemaSlugState(splitSistemaKey(sistema).sistema);
  }, [sistema]);

  useEffect(() => {
    // Verifica que la key activa persistida sigue disponible
    if (sistemas.length > 0 && !sistemas.includes(sistema)) {
      setSistemaState(
        preferred && sistemas.includes(preferred) ? preferred : sistemas[0]
      );
    }
  }, [sistema, sistemas, preferred]);

  // ── Deep-link por URL (I-50) ───────────────────────────────────────────────
  // SSoT de navegación = la barra (?empresa=&sistema=). Solo multi: en single/dev
  // no hay empresa en la URL. La forma de la query es independiente del API Go.
  const router = useRouter();
  const pathname = usePathname();

  // LEER (one-shot, al llegar la lista): la URL GANA sobre localStorage; sin
  // params/ inválida → se conserva el seed de localStorage. Leemos
  // window.location.search (no useSearchParams: en output:'export' exigiría un
  // boundary <Suspense> alrededor de todo el shell).
  useEffect(() => {
    if (hydratedRef.current || sistemas.length === 0) return; // espera a tener datos
    hydratedRef.current = true; // a partir de acá el write-effect puede escribir
    if (!isTwoLevel(sistemas)) return;
    const sel = resolveSistemaIntent(sistemas, parseSistemaSearch(window.location.search));
    if (!sel) return; // sin deep-link → localStorage manda
    if (sel.key && sel.key !== sistema) {
      applyingIntentRef.current = true; // salta la derivación key→slug una vez
      setSistema(sel.key); // persiste localStorage también
    }
    setEmpresaState(sel.empresa);
    setSistemaSlugState(sel.sistemaSlug);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sistemas]);

  // ESCRIBIR (sync bidireccional): cada cambio de empresa/sistema reescribe la
  // URL con router.replace (sin ensuciar el historial). Tras cambiar de tab
  // (Link sin query) el pathname cambia → re-agregamos la query. Idempotente.
  useEffect(() => {
    if (!hydratedRef.current || mode !== 'multi') return;
    const search = buildSistemaSearch(empresa, sistemaSlug);
    if (typeof window !== 'undefined' && search === window.location.search) return;
    router.replace(`${pathname}${search}`, { scroll: false });
  }, [empresa, sistemaSlug, pathname, mode, router]);

  const setSistema = (key: string) => {
    setSistemaState(key);
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, key);
    }
  };

  // Elegir empresa: selecciona su primer sistema y salta a él.
  const setEmpresa = (slug: string) => {
    setEmpresaState(slug);
    const first = sistemasFor(sistemas, slug)[0];
    setSistemaSlugState(first?.sistema ?? '');
    if (first) setSistema(first.key);
  };

  // Elegir sistema dentro de la empresa: siempre navegable (la lista plana solo
  // contiene sistemas reales) → salta directo.
  const setSistemaSlug = (slug: string) => {
    setSistemaSlugState(slug);
    const s = sistemasFor(sistemas, empresa).find((e) => e.sistema === slug);
    if (s) setSistema(s.key);
  };

  const value = useMemo(
    () => ({
      sistema, setSistema, sistemas, mode, empresa, setEmpresa, sistemaSlug, setSistemaSlug,
      escribibles,
    }),
    [sistema, sistemas, mode, empresa, sistemaSlug, escribibles]
  );

  return <SistemaContext.Provider value={value}>{children}</SistemaContext.Provider>;
}

export function useSistema(): SistemaContextValue {
  const ctx = useContext(SistemaContext);
  if (!ctx) throw new Error('useSistema fuera de SistemaProvider');
  return ctx;
}
