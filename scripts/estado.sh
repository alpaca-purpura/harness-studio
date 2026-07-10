#!/usr/bin/env bash
# Genera (o VERIFICA con --check) el bloque de cifras vivas de docs/product/checkpoint.md desde el
# estado REAL del repo. RF-178 (reorg-docs, HS-18) + HS-20: las cifras NO se teclean — se generan.
# HS-21: --check cablea la generación a CI (drift-gate) — cifras stale rompen el merge.
#   - ruleset --todo / dogfood --arnes → `arnesia conformance`
#   - arch boundaries → conteo de docs/architecture/boundaries/*.md
#   - knowledge nodos·checks → grupo del ruleset por `elemento` de knowledge/elements/ (json)
#   - capabilities → distribución de `status:` del árbol docs/product/capabilities/
# Uso:
#   bash scripts/estado.sh            regenera el bloque in-place (default)
#   bash scripts/estado.sh --check    drift-gate de CI: compara CIFRAS (ignora la fecha de medición),
#                                     exit 1 si el checkpoint está stale; NO escribe el árbol.
set -euo pipefail
cd "$(dirname "$0")/.."

MODE=generate
if [ "${1:-}" = "--check" ]; then MODE=check; fi

TODO_LINE=$(go run ./cmd/arnesia conformance --todo 2>/dev/null | grep -E "checks ·" | head -1 | sed 's/^[[:space:]]*//')
ARNES_LINE=$(go run ./cmd/arnesia conformance --arnes dogfood/dev-full-cycle.graph.json 2>/dev/null | grep -E "checks ·" | head -1 | sed 's/^[[:space:]]*//')
if [ -z "${TODO_LINE:-}" ]; then
  echo "no se pudo obtener la cifra de conformance --todo" >&2
  exit 1
fi
STAMP=$(git log -1 --format=%cd --date=short 2>/dev/null || echo "s/f")

# Reescribe (o compara, si --check) el bloque <!--stats ... /stats--> del checkpoint.
# (python obtiene el json vía subprocess — el heredoc ya ocupa stdin del intérprete.)
python3 - "$TODO_LINE" "$ARNES_LINE" "$STAMP" "$MODE" <<'PY'
import re, sys, os, json, glob, subprocess, difflib
from collections import Counter

todo, arnes, stamp, mode = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
data = json.loads(subprocess.run(
    ['go', 'run', './cmd/arnesia', 'conformance', '--todo', '--json'],
    capture_output=True, text=True).stdout)

# arch boundaries = hojas de boundaries/
boundaries = len(glob.glob('docs/architecture/boundaries/*.md'))

# knowledge nodos·checks = checks del ruleset cuyo `elemento` es un nodo de knowledge/elements/
know = {os.path.splitext(f)[0] for f in os.listdir('docs/architecture/knowledge/elements') if f.endswith('.md')}
c = Counter(x['check']['elemento'] for x in data['results'])
kn_nodos = sum(1 for k in c if k in know)
kn_checks = sum(v for k, v in c.items() if k in know)

# capabilities: distribución de status (parse dep-free del frontmatter)
st = Counter()
for f in glob.glob('docs/product/capabilities/*/*.yaml'):
    for ln in open(f, encoding='utf-8'):
        s = ln.strip()
        if s.startswith('status:'):
            st[s[len('status:'):].split('#')[0].strip()] += 1
            break
total = sum(st.values())
caps = (f"{total} — {st['vivo']} vivo · {st['vivo·nc']} vivo·nc · {st['parcial']} parcial · "
        f"{st['stub']} stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)")

block = (
    "<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->\n"
    f"- **ruleset `--todo`:** `{todo}` (medido {stamp}, `go run ./cmd/arnesia conformance --todo`)\n"
    f"- **dogfood `--arnes`:** `{arnes}` (warn honesto `art-es-path`, el diente no se silencia) — medido {stamp}\n"
    f"- **arch/:** {boundaries} boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)\n"
    f"- **docs/architecture/knowledge/:** {kn_nodos} nodos · {kn_checks} checks\n"
    f"- **capabilities (SSoT):** {caps}\n"
    "<!--/stats-->"
)

p = 'docs/product/checkpoint.md'
t = open(p, encoding='utf-8').read()


def norm(s):
    # La fecha `medido <...>` es PROCEDENCIA (cuándo se midió), no una cifra. Se ignora en la
    # comparación del gate: si no, cualquier commit de otro día (que solo avanza el committer-date
    # de HEAD) rompería el merge sin que ninguna cifra haya cambiado. Ver decisiones.md D1.
    # Anclado a fecha ISO (o el fallback `s/f`) — NO `[^,\n]+` — para no enmascarar por accidente
    # una cifra futura que quede tras el token «medido» en una línea de stats (auditoría HS-21).
    return re.sub(r'medido (\d{4}-\d{2}-\d{2}|s/f)', 'medido <DATE>', s)


if mode == 'check':
    m = re.search(r'<!--stats:.*?<!--/stats-->', t, flags=re.S)
    if not m:
        print('DRIFT: no se encontró el bloque <!--stats ... /stats--> en checkpoint.md', file=sys.stderr)
        raise SystemExit(1)
    if norm(m.group(0)) == norm(block):
        print('estado.sh --check: cifras de docs/product/checkpoint.md en sync con el estado real ✓')
        raise SystemExit(0)
    diff = difflib.unified_diff(
        norm(m.group(0)).splitlines(), norm(block).splitlines(),
        fromfile='checkpoint.md (commiteado)', tofile='estado real (regenerado)', lineterm='')
    print('DRIFT: las cifras de docs/product/checkpoint.md están stale (cifras ≠ estado real).', file=sys.stderr)
    print('\n'.join(diff), file=sys.stderr)
    print('\n→ corré: bash scripts/estado.sh   y commiteá el checkpoint.', file=sys.stderr)
    raise SystemExit(1)

# mode == generate (default): reescribe el bloque in-place.
t2, n = re.subn(r'<!--stats:.*?<!--/stats-->', lambda _: block, t, count=1, flags=re.S)
if n != 1:
    print('no se encontró el bloque <!--stats ... /stats--> en checkpoint.md', file=sys.stderr)
    raise SystemExit(1)
open(p, 'w', encoding='utf-8').write(t2)
print('docs/product/checkpoint.md → bloque de cifras regenerado:')
print(block)
PY
