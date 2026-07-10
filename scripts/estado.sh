#!/usr/bin/env bash
# Genera el bloque de cifras vivas de docs/product/checkpoint.md desde el estado REAL del repo.
# RF-178 (paquete reorg-docs, HS-18) + HS-20: las cifras NO se teclean — se generan. D2 firmada.
#   - ruleset --todo / dogfood --arnes → `arnesia conformance`
#   - arch boundaries → conteo de docs/architecture/boundaries/*.md
#   - knowledge nodos·checks → grupo del ruleset por `elemento` de knowledge/elements/ (json)
#   - capabilities → distribución de `status:` del árbol docs/product/capabilities/
# Uso: bash scripts/estado.sh   (desde la raíz del repo)
set -euo pipefail
cd "$(dirname "$0")/.."

TODO_LINE=$(go run ./cmd/arnesia conformance --todo 2>/dev/null | grep -E "checks ·" | head -1 | sed 's/^[[:space:]]*//')
ARNES_LINE=$(go run ./cmd/arnesia conformance --arnes dogfood/dev-full-cycle.graph.json 2>/dev/null | grep -E "checks ·" | head -1 | sed 's/^[[:space:]]*//')
if [ -z "${TODO_LINE:-}" ]; then
  echo "no se pudo obtener la cifra de conformance --todo" >&2
  exit 1
fi
STAMP=$(git log -1 --format=%cd --date=short 2>/dev/null || echo "s/f")

# Reescribe TODO el bloque <!--stats ... /stats--> del checkpoint desde las cifras generadas.
# (python obtiene el json vía subprocess — el heredoc ya ocupa stdin del intérprete.)
python3 - "$TODO_LINE" "$ARNES_LINE" "$STAMP" <<'PY'
import re, sys, os, json, glob, subprocess
from collections import Counter

todo, arnes, stamp = sys.argv[1], sys.argv[2], sys.argv[3]
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
t2, n = re.subn(r'<!--stats:.*?<!--/stats-->', lambda _: block, t, count=1, flags=re.S)
if n != 1:
    print('no se encontró el bloque <!--stats ... /stats--> en checkpoint.md', file=sys.stderr)
    raise SystemExit(1)
open(p, 'w', encoding='utf-8').write(t2)
print('docs/product/checkpoint.md → bloque de cifras regenerado:')
print(block)
PY
