/**
 * Helpers para construir paths relativos al workspace root desde
 * un absoluto, usado por componentes que pasan paths a /api/file.
 *
 * Genérico (F-4 · sin lista hardcodeada de sistemas): el caller pasa la `sistema`
 * (la conoce por contexto — `story.sistema` o `useSistema()`) y la heurística
 * ancla en `{sistema}/docs/...`. Para contextos no-sistema ancla en el primer
 * segmento `docs` / `tools` / `.claude`. No requiere roundtrip al servidor.
 */

const NON_SISTEMA_ROOTS = new Set(['docs', 'tools', '.claude']);

/**
 * Convierte un path absoluto del story (ej.
 * /home/.../workspace/{sistema}/docs/product/stories/X) en path
 * relativo al workspace root: {sistema}/docs/product/stories/X.
 *
 * Si no encuentra un segmento ancla, devuelve null (el caller
 * debería usar /api/open con abs path en su lugar).
 */
export function absToRel(absPath: string, sistema?: string | null): string | null {
  if (!absPath) return null;
  const segments = absPath.split('/').filter(Boolean);
  for (let i = 0; i < segments.length; i++) {
    const seg = segments[i];
    // Ancla primaria: `{sistema}/docs/...` (sistema conocida por el caller)
    if (sistema && seg === sistema && segments[i + 1] === 'docs') {
      return segments.slice(i).join('/');
    }
    // Ancla no-sistema: docs/ raíz (platform) · tools/ · .claude/
    if (NON_SISTEMA_ROOTS.has(seg)) {
      return segments.slice(i).join('/');
    }
  }
  return null;
}

/**
 * Devuelve la rel path al checkpoint.md de una story.
 */
export function storyArtifactRel(
  storyAbsPath: string,
  artifact: string,
  sistema?: string | null
): string | null {
  const rel = absToRel(storyAbsPath, sistema);
  if (!rel) return null;
  return `${rel}/${artifact}`;
}
