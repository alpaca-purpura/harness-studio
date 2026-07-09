#!/usr/bin/env bash
# Genera el bloque de cifras vivas de ESTADO.md desde `arnesia conformance --todo`.
# RF-178 (paquete reorg-docs, HS-18): las cifras NO se teclean — se generan. D2 firmada.
# Uso: bash scripts/estado.sh   (desde la raíz del repo)
set -euo pipefail
cd "$(dirname "$0")/.."

# Ruleset completo (elemento|--arnes|--todo → --todo corre todo el ruleset).
LINE=$(go run ./cmd/arnesia conformance --todo 2>/dev/null | grep -E "checks ·" | head -1 | sed 's/^[[:space:]]*//')
if [ -z "${LINE:-}" ]; then
  echo "no se pudo obtener la cifra de conformance" >&2
  exit 1
fi
STAMP=$(git log -1 --format=%cd --date=short 2>/dev/null || echo "s/f")

# Reescribe SOLO la línea del ruleset dentro del bloque <!--stats ... -->.
python3 - "$LINE" "$STAMP" <<'PY'
import re, sys
line, stamp = sys.argv[1], sys.argv[2]
p = 'ESTADO.md'
t = open(p, encoding='utf-8').read()
new = f'- **ruleset `--todo`:** `{line}` (medido {stamp}, `go run ./cmd/arnesia conformance --todo`)'
t2 = re.sub(r'- \*\*ruleset `--todo`:\*\*.*', new, t, count=1)
open(p, 'w', encoding='utf-8').write(t2)
print('ESTADO.md ruleset →', line)
PY
