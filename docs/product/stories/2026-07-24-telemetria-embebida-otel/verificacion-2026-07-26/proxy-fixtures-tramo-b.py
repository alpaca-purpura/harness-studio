#!/usr/bin/env python3
"""Proxy READ-ONLY sobre el daemon real (4200) que sustituye /api/telemetria/* por fixtures.

- Solo reenvia GET al daemon real. Cualquier otro metodo devuelve 405 sin tocar el daemon.
- Escenario controlado por el archivo `escenario.txt` en el mismo directorio (hot-reload).
"""
import json
import os
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

UP = "http://127.0.0.1:4200"
AQUI = os.path.dirname(os.path.abspath(__file__))


def escenario():
    try:
        with open(os.path.join(AQUI, "escenario.txt")) as f:
            return f.read().strip()
    except FileNotFoundError:
        return "rico"


def cobertura(e, h, p, s):
    return {"esperados": None, "exacta": e, "por_hash": h, "por_proceso": p,
            "sin_dato": s, "no_llegaron": None}


CATALOGO = {"version": "2026-07-26", "rev": "b439a9a", "modelos": 694,
            "refrescado": "2026-07-26T10:00:00Z"}


def resumen(esc):
    base = {"desde": "2026-07-19T00:00:00Z", "hasta": "2026-07-26T00:00:00Z",
            "estimado": True, "costo_completo": True, "divergencia_pct": 0.0,
            "divergencia_sospechosa": False, "turnos": 61, "catalogo": CATALOGO}
    if esc == "vacio":
        return dict(base, costo_reportado_micros=None, costo_calculado_micros=None,
                    corridas=0, sesiones=0, cajas=0, escenario="s1", confianza="sin-dato",
                    cobertura=cobertura(0, 0, 0, 0))
    if esc == "fuera-de-ventana":
        # Hubo corridas, pero NO en esta ventana. El wire devuelve 0 corridas.
        return dict(base, costo_reportado_micros=None, costo_calculado_micros=None,
                    corridas=0, sesiones=0, cajas=0, escenario="s1", confianza="sin-dato",
                    cobertura=cobertura(0, 0, 0, 0))
    if esc == "s2":
        # Lo que el wire produce en s2-instrumentado: corrida_id es NULL => COUNT(DISTINCT
        # corrida_id) = 0, pero el costo, la cobertura (que cuenta TURNOS) y las cajas si tienen
        # dato. Ver consultas.go:95 vs consultas.go:143.
        return dict(base, costo_reportado_micros=1_920_000, costo_calculado_micros=1_900_000,
                    corridas=0, sesiones=12, cajas=4, escenario="s2-instrumentado",
                    confianza="exacta", cobertura=cobertura(58, 0, 0, 3))
    if esc == "parcial":
        return dict(base, costo_reportado_micros=1_920_000, costo_calculado_micros=1_920_000,
                    corridas=61, sesiones=12, cajas=4, escenario="s1", confianza="sin-dato",
                    cobertura=cobertura(20, 5, 3, 33))
    # rico
    return dict(base, costo_reportado_micros=1_920_000, costo_calculado_micros=1_900_000,
                corridas=61, sesiones=12, cajas=4, escenario="s1", confianza="por-proceso",
                cobertura=cobertura(44, 9, 5, 3))


PUNTO = {
    "id": "p-b1-hipaa", "detector": "b1-rewarm-por-ttl", "score_version": 1,
    "titulo": "La caja hipaa-check paga el re-warm del cache 3 de cada 4 veces",
    "lede": "USD 1,08 de los USD 1,92 de la ventana (56 %) se van en re-warm por TTL.",
    "caja_id": "hipaa-check", "caja_nombre": "hipaa-check",
    "gasto_micros": 1_080_000, "parte_del_total": 0.56,
    "contrafactual": "Con el TTL en 1 h, las mismas 14 corridas habrian costado USD 0,55 — se ahorran USD 0,53 en la ventana.",
    "diferencia_micros": 530_000,
    "umbral": "re-warms / corridas > 0,5",
    "calculo": "11 re-warms / 14 corridas = 0,79 > 0,5",
    "sesgo": "Asume que el cache de 1 h habria estado caliente: subestima el costo real si las sesiones se espacian mas.",
    "direccion_sesgo": "subestima",
    "fix": "Subi el TTL del cache a 1 h en la caja: cache_control: {\"type\":\"ephemeral\",\"ttl\":\"1h\"}",
    "fix_codigo": "cache_control: {\"type\":\"ephemeral\",\"ttl\":\"1h\"}",
    "confianza": "exacta", "corridas_usadas": 14, "corridas_totales": 14, "grave": False,
}

DETECTORES = [
    {"detector": "b1-rewarm-por-ttl", "nombre": "re-warm TTL", "aplica": True, "hallazgos": 1},
    {"detector": "p1-caja-que-consume-y-se-rechaza", "nombre": "rechazo en gate",
     "aplica": False, "motivo": "sin senal de gate en estas corridas", "hallazgos": 0},
    {"detector": "b3-cambio-de-modelo-invalida-cache", "nombre": "modelo cambiado",
     "aplica": True, "hallazgos": 0},
    {"detector": "b6-sesion-abandonada", "nombre": "sesion abandonada", "aplica": True, "hallazgos": 0},
    {"detector": "b2-costo-de-la-rotacion", "nombre": "rotacion cara", "aplica": True, "hallazgos": 0},
    {"detector": "b4-gasto-por-arnes-empresa-puesto", "nombre": "concentracion de gasto",
     "aplica": True, "hallazgos": 0},
]

NO_MEDIDOS = [
    {"detector": "b5", "nombre": "herramienta que devuelve de mas", "aplica": False, "hallazgos": 0},
    {"detector": "p2", "nombre": "reintentos del mismo turno", "aplica": False, "hallazgos": 0},
]


def cajas(esc):
    if esc in ("vacio", "fuera-de-ventana"):
        return {"cajas": []}
    return {"cajas": [{
        "caja_id": "hipaa-check", "nombre": "hipaa-check", "atribuible": True,
        "costo_micros": 1_080_000, "parte": 0.56, "confianza": "exacta", "corridas": 14,
        "marcas": [{"detector": "b1-rewarm-por-ttl", "nombre": "re-warm TTL", "grave": False}],
    }]}


def mejoras(esc):
    if esc in ("vacio", "fuera-de-ventana"):
        return {"puntos": [], "no_aplican": [], "no_medidos": NO_MEDIDOS}
    if esc == "parcial":
        return {"puntos": [], "no_aplican": [d for d in DETECTORES if not d["aplica"]],
                "no_medidos": NO_MEDIDOS}
    if esc == "muchos":
        pts=[]
        for i in range(4):
            q=dict(PUNTO); q["id"]=f"p{i}"; q["diferencia_micros"]=530_000-i*10_000
            pts.append(q)
        return {"puntos": pts, "no_aplican": [d for d in DETECTORES if not d["aplica"]],
                "no_medidos": NO_MEDIDOS}
    return {"puntos": [PUNTO], "no_aplican": [d for d in DETECTORES if not d["aplica"]],
            "no_medidos": NO_MEDIDOS}


def detalle(esc):
    if esc in ("vacio", "fuera-de-ventana"):
        # El daemon real responde 200 con un detalle vacio para una caja sin corridas.
        return {"turnos_totales": 0}
    return {"turnos_totales": 14,
            "tokens": {"entrada": 10, "salida": 39, "cache_lectura": 17536,
                       "cache_escritura_5m": 0, "cache_escritura_1h": 8257},
            "paridad": {"reportado_micros": 1_080_000, "calculado_micros": 1_260_000,
                        "divergencia_pct": 0.16, "completo": True,
                        "catalogo_version": "2026-07-26"},
            "detectores": DETECTORES}


def portafolio(esc):
    if esc == "vacio":
        return {"filas": []}
    return {"filas": [
        # Fila con hallazgos pero SIN punto principal: el caso que D24.4 dice haber corregido.
        {"arnes_id": "vitalia", "clave": "sin-home~vitalia~vitalia", "nombre": "Vitalia",
         "instalacion_id": "h:ef84e6f2e14a0560", "puesto": "Ingenieria - Desarrollo full-cycle",
         "corridas": 61, "costo_micros": 1_920_000, "costo_por_corrida": 31_000,
         "confianza": "por-proceso", "puntos_de_mejora": 3},
        # Fila que SI corrio pero sin costo atribuible: corridas > 0, costo_por_corrida null.
        {"arnes_id": "harness", "clave": "github-com-alpacapurpura-prenter-marketplace~harness~",
         "nombre": "harness", "instalacion_id": "", "puesto": None,
         "corridas": 47, "costo_micros": None, "costo_por_corrida": None,
         "confianza": "sin-dato", "puntos_de_mejora": 0},
        # Fila limpia de verdad.
        {"arnes_id": "dev-full-cycle", "clave": "sin-home~dev-full-cycle~", "nombre": "dev",
         "instalacion_id": "i2", "puesto": None, "corridas": 9,
         "costo_micros": 90_000, "costo_por_corrida": 10_000,
         "confianza": "exacta", "puntos_de_mejora": 0},
    ]}


SALUD = {"retencion_dias": 400, "retencion_propuesta": False, "forward": True,
         "forward_destino": "https://otlp.datadoghq.com", "catalogo": CATALOGO,
         "ultima_recepcion": "2026-07-26T12:00:00Z", "almacen_disponible": False,
         "almacen_motivo": "disco lleno"}


SESIONES = [{
    "id": "s-audit", "frente": "Auditoria Tramo B", "arnes": "vitalia",
    "empresa": "vitalia", "puesto": "Ingenieria - Desarrollo full-cycle",
    "status": "idle", "view": "Mapa", "conv": [],
}]


class H(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, *a):
        pass

    def _cors(self):
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Headers", "*")
        self.send_header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

    def _json(self, obj, code=200):
        b = json.dumps(obj).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self._cors()
        self.send_header("Content-Length", str(len(b)))
        self.end_headers()
        self.wfile.write(b)

    def do_OPTIONS(self):
        self.send_response(204)
        self._cors()
        self.send_header("Content-Length", "0")
        self.end_headers()

    def do_GET(self):
        p = self.path.split("?")[0]
        esc = escenario()
        # Sesion SINTETICA: jamas se le pide al daemon real que cree ni liste sesiones.
        if p == "/api/sessions":
            return self._json(SESIONES)
        if p.startswith("/api/sessions/"):
            return self._json([])
        if p.startswith("/api/telemetria/"):
            if p == "/api/telemetria/resumen":
                con_ventana = "desde=" in self.path
                if esc == "ventana":
                    # Arnes con 340 corridas historicas y NINGUNA en los ultimos 7/30 dias.
                    if con_ventana:
                        return self._json(dict(resumen("vacio"), corridas=0))
                    return self._json(dict(resumen("rico"), corridas=340, sesiones=88))
                return self._json(resumen(esc))
            if p == "/api/telemetria/salud":
                return self._json(SALUD)
            if p == "/api/telemetria/portafolio":
                return self._json(portafolio(esc))
            if p.endswith("/mejoras"):
                return self._json(mejoras(esc))
            if p.endswith("/cajas"):
                return self._json(cajas(esc))
            if "/cajas/" in p:
                if esc == "detalle-roto":
                    return self._json({"error": "boom"}, 500)
                return self._json(detalle(esc))
            return self._json({}, 404)
        # Todo lo demas: GET al daemon real (solo lectura).
        try:
            with urllib.request.urlopen(UP + self.path, timeout=10) as r:
                body = r.read()
                self.send_response(r.status)
                for k in ("Content-Type",):
                    if r.headers.get(k):
                        self.send_header(k, r.headers[k])
                self._cors()
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)
        except Exception as e:
            b = str(e).encode()
            self.send_response(502)
            self._cors()
            self.send_header("Content-Length", str(len(b)))
            self.end_headers()
            self.wfile.write(b)

    def do_POST(self):
        self._json({"error": "proxy read-only"}, 405)

    def do_PUT(self):
        self._json({"error": "proxy read-only"}, 405)

    def do_DELETE(self):
        self._json({"error": "proxy read-only"}, 405)


ThreadingHTTPServer(("127.0.0.1", 4300), H).serve_forever()
