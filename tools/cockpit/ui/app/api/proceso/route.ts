import { NextResponse } from 'next/server';
import { CATEGORIAS, TERMINALES, leerEspejo } from '@/lib/proceso-server';

/**
 * Paridad dev-mode de GET /api/proceso (RN-32 — el Go implementa esto en
 * handlers_proceso.go y es quien sirve en el binario; esta route solo existe para
 * `pnpm dev`). Lee el MISMO espejo gated que el Go embebe con go:embed
 * (lib/proceso-server.ts, compartida con la route de transition — F6/RN-51).
 */

export async function GET() {
  let doc;
  try {
    doc = await leerEspejo();
  } catch (e) {
    return NextResponse.json(
      { error: `espejo del descriptor ilegible: ${e instanceof Error ? e.message : e}` },
      { status: 500 }
    );
  }
  const estados = doc.estados.map((e) => ({
    ...e,
    terminal: TERMINALES.has(e.categoria),
  }));
  return NextResponse.json({
    descriptor: doc.descriptor,
    categorias: CATEGORIAS,
    estados,
    transiciones: doc.transiciones,
    gates: doc.gates,
    duenos: doc.duenos,
    parametros: doc.parametros,
    fuente: 'process/sdd-default.yaml (paridad dev — espejo gated)',
  });
}
