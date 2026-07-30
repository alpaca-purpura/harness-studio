#!/usr/bin/env python3
"""Inyección de telemetría por el CAMINO REAL: OTLP/JSON a POST /v1/logs del daemon sandbox.

Serie diseñada para disparar B4 (gasto concentrado: una caja > 1/3 del total):
  - caja `hipaa-check`            : 3 api_request · 42000+39000+41000 = 122000 micros
  - caja `shell-mockup-per-component`: 1 api_request · 9000 micros
  Total 131000 → hipaa-check = 93.1 % > 33 %  ⇒ B4 dispara con caja_id=hipaa-check.
  (B6 también disparará: la sesión gastó y no tiene señal de proceso.)

Shape copiado del golden file real logs-run1.json (Claude Code 2.1.220).
"""
import json
import time
import urllib.request

BASE = "http://127.0.0.1:4200"
ARNES = "sin-home~vitalia~vitalia"      # clave del ÍNDICE (la que usa el FE), no el arnes.id
INSTALACION = "home-local"
SESION = "e2ec1a2b-4f00-4c1a-9e2e-c1a2b3c4d5e6"

def attr(k, v):
    if isinstance(v, int):
        return {"key": k, "value": {"intValue": v}}
    return {"key": k, "value": {"stringValue": v}}

def api_request(ts_nano, prompt_id, caja, micros, inp, out, cr, cw, dur):
    return {
        "timeUnixNano": str(ts_nano),
        "observedTimeUnixNano": str(ts_nano),
        "body": {"stringValue": "claude_code.api_request"},
        "attributes": [
            attr("arnesia.arnes", ARNES),
            attr("arnesia.caja", caja),
            attr("arnesia.instalacion", INSTALACION),
            attr("session.id", SESION),
            attr("prompt.id", prompt_id),
            attr("event.name", "api_request"),
            attr("model", "claude-haiku-4-5-20251001"),
            attr("input_tokens", inp),
            attr("output_tokens", out),
            attr("cache_read_tokens", cr),
            attr("cache_creation_tokens", cw),
            attr("cost_usd_micros", micros),
            attr("duration_ms", dur),
            attr("speed", "normal"),
            attr("query_source", "sdk"),
            attr("terminal.type", "WarpTerminal"),
        ],
    }

now = time.time_ns()
m = 60 * 1_000_000_000  # un minuto en nanos
records = [
    api_request(now - 9 * m, "aaaa1111-0000-4000-8000-000000000001", "hipaa-check", 42000, 220, 900, 17000, 8000, 2100),
    api_request(now - 7 * m, "aaaa1111-0000-4000-8000-000000000002", "hipaa-check", 39000, 180, 850, 18500, 0, 1900),
    api_request(now - 5 * m, "aaaa1111-0000-4000-8000-000000000003", "hipaa-check", 41000, 210, 880, 18800, 0, 2000),
    api_request(now - 3 * m, "aaaa1111-0000-4000-8000-000000000004", "shell-mockup-per-component", 9000, 60, 200, 4000, 0, 800),
]

payload = {
    "resourceLogs": [
        {
            "resource": {
                "attributes": [
                    attr("arnesia.arnes", ARNES),
                    attr("arnesia.caja", "hipaa-check"),
                    attr("arnesia.instalacion", INSTALACION),
                    attr("host.arch", "amd64"),
                    attr("os.type", "linux"),
                    attr("service.name", "claude-code"),
                    attr("service.version", "2.1.220"),
                ],
                "droppedAttributesCount": 0,
            },
            "scopeLogs": [
                {
                    "scope": {"name": "com.anthropic.claude_code.events", "version": "2.1.220"},
                    "logRecords": records,
                }
            ],
        }
    ]
}

body = json.dumps(payload).encode()
req = urllib.request.Request(
    BASE + "/v1/logs", data=body, method="POST",
    headers={"Content-Type": "application/json"},
)
with urllib.request.urlopen(req, timeout=10) as resp:
    print("HTTP", resp.status, resp.read().decode())
