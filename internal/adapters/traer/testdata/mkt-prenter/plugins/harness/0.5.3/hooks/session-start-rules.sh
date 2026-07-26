#!/usr/bin/env bash
# core-harness · SessionStart rule injector (the always-on-rules workaround for plugin distribution).
# Plugins have NO always-on-rules channel (a plugin's CLAUDE.md/.claude/rules are NOT auto-loaded).
# So this hook emits the SLIM always-on rule core (≤10k chars · the additionalContext cap) at session
# start, the way `.claude/rules/` would. The FULL corpus is delivered separately by /harness:bootstrap
# (it copies core-harness/rules/* into the adopter's .claude/rules/ for native always-on load).
#
# Mechanism (verified date-aware 2026-06-09, official CC hooks docs):
#   SessionStart stdout / hookSpecificOutput.additionalContext is injected before the first prompt,
#   capped at 10,000 chars (overflow → saved to a file + a pointer, NOT the body) → keep the slim core lean.
#
# SCAFFOLD (W7 2026-06-09): always-on-core.manifest is EMPTY until the W7 rules-move curates it →
#   this hook is a safe NO-OP today (emits nothing). NOT wired into the project's settings.json (inert).
#
# Wired only when the harness plugin is installed (hooks/hooks.json → SessionStart).
set -euo pipefail

ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
MANIFEST="$ROOT/hooks/always-on-core.manifest"
RULES_DIR="$ROOT/rules"
CAP=9500   # headroom under the 10,000-char additionalContext cap

[ -f "$MANIFEST" ] || exit 0

# Collect the curated slim-core rule files (manifest lines; '#' comments + blanks ignored).
buf=""
while IFS= read -r line; do
  line="${line%%#*}"; line="$(echo "$line" | xargs || true)"
  [ -z "$line" ] && continue
  f="$RULES_DIR/$line"
  [ -f "$f" ] && buf+="$(cat "$f")"$'\n\n'
done < "$MANIFEST"

# Safe no-op when nothing is curated (scaffold state) or the corpus is absent.
[ -z "${buf// /}" ] && exit 0

# Truncate to the cap (leave a marker if truncated) + emit additionalContext JSON (python = safe escaping).
printf '%s' "$buf" | CAP="$CAP" python3 -c '
import json, os, sys
body = sys.stdin.read()
cap = int(os.environ["CAP"])
if len(body) > cap:
    body = body[:cap] + "\n\n[…slim core truncated at cap; full rules via /harness:bootstrap]"
print(json.dumps({"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": body}}))
'
