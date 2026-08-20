import json, subprocess, sys

ENDPOINTS = [
    "/api/stories?brand=client-acme",
    "/api/stories?brand=client-xyz",
    "/api/stories?brand=internal-platform",
    "/api/stories/acme-001?brand=client-acme",
    "/api/capabilities?brand=client-acme",
    "/api/capabilities/auth/auth-service?brand=client-acme",
    "/api/releases?brand=client-acme",
    "/api/releases?brand=client-xyz",
    "/api/learnings?brand=client-acme",
    "/api/system-map?brand=client-acme",
    "/api/value-stream?brand=client-acme",
    "/api/sessions",
    "/api/capabilities/status?brand=client-acme",
    "/api/capabilities/bidirectional?brand=client-acme",
    "/api/capabilities/code-index?brand=client-acme",
    "/api/capabilities/doctor?brand=client-acme",
    "/api/harness",
    "/api/cil",
]

def fetch(port, ep):
    out = subprocess.run(["curl", "-s", f"http://localhost:{port}{ep}"], capture_output=True, text=True)
    try:
        return json.loads(out.stdout)
    except Exception:
        return {"__unparseable__": out.stdout[:200]}

def diff(a, b, path=""):
    diffs = []
    if type(a) != type(b):
        # int/float equivalence
        if isinstance(a, (int, float)) and isinstance(b, (int, float)) and a == b:
            return diffs
        diffs.append(f"{path}: type {type(a).__name__} vs {type(b).__name__} ({a!r} vs {b!r})")
        return diffs
    if isinstance(a, dict):
        for k in sorted(set(a) | set(b)):
            if k not in a:
                diffs.append(f"{path}.{k}: missing in NEXT")
            elif k not in b:
                diffs.append(f"{path}.{k}: missing in GO")
            else:
                diffs.extend(diff(a[k], b[k], f"{path}.{k}"))
    elif isinstance(a, list):
        if len(a) != len(b):
            diffs.append(f"{path}: list len {len(a)} vs {len(b)}")
        for i, (x, y) in enumerate(zip(a, b)):
            diffs.extend(diff(x, y, f"{path}[{i}]"))
    else:
        if a != b:
            diffs.append(f"{path}: {a!r} vs {b!r}")
    return diffs

total_diffs = 0
for ep in ENDPOINTS:
    a = fetch(4000, ep)
    b = fetch(4010, ep)
    d = diff(a, b, ep)
    if d:
        total_diffs += len(d)
        print(f"✗ {ep}: {len(d)} diffs")
        for line in d[:6]:
            print(f"    {line}")
    else:
        print(f"✓ {ep}")
print(f"\n{total_diffs} total diffs")
sys.exit(1 if total_diffs else 0)
