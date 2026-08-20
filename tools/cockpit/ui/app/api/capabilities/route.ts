/**
 * GET /api/capabilities?sistema={sistema}  → { capabilities: Capability[] }
 *
 * Lee todos los YAMLs (markdown con frontmatter) en
 * `{sistema}/docs/product/capabilities/<module>/<slug>.yaml`.
 */

import { NextRequest, NextResponse } from 'next/server';
import path from 'node:path';
import { errorResponse } from '../_lib/responses';
import { readCapability } from '@/lib/cap-ledger';
import { globPaths } from '@/lib/fs-reader';
import { capabilitiesPath, getSistemas } from '@/lib/workspace';

export async function GET(req: NextRequest): Promise<NextResponse> {
  const sistema = req.nextUrl.searchParams.get('sistema');
  if (!sistema) return errorResponse('query param "sistema" requerido', 400);
  if (!getSistemas().includes(sistema)) {
    return errorResponse(`sistema desconocido: ${sistema}`, 400);
  }

  try {
    const baseDir = capabilitiesPath(sistema);
    // glob: capabilities/*/*.yaml
    const files = await globPaths('*/*.yaml', baseDir);

    const capabilities = await Promise.all(
      files.map(async (abs) => {
        try {
          return await readCapability(abs);
        } catch {
          return null;
        }
      })
    );

    const valid = capabilities.filter((c): c is NonNullable<typeof c> => c !== null);
    return NextResponse.json({ capabilities: valid });
  } catch (err) {
    return errorResponse('error agregando capabilities', 500, {
      detail: (err as Error).message,
    });
  }
}

// Helper interno usado por otros endpoints
export function capYamlPath(sistema: string, module: string, slug: string): string {
  return path.join(capabilitiesPath(sistema), module, `${slug}.yaml`);
}
