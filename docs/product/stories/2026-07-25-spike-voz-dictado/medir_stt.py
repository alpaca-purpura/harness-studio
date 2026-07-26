#!/usr/bin/env python3
"""Banco de medición del Fork A2 (STT local) — el que produjo los números de `spike-spec.md` §1.8.

No toca el sistema: venv aislado, sin `sudo`, sin instalar nada global.

    uv venv stt-venv --python 3.12
    uv pip install --python stt-venv/bin/python faster-whisper piper-tts

    # audio de prueba: voz española REAL sintetizada (no silencio, no ruido — si no, el número miente:
    # el decoder produce pocos tokens sobre silencio y la medición sale optimista)
    ./stt-venv/bin/python -m piper.download_voices es_ES-davefx-medium
    ./stt-venv/bin/python -m piper -m es_ES-davefx-medium.onnx -f dictado.wav < dictado.txt

    ./stt-venv/bin/python medir_stt.py base dictado.wav

Resultado del 2026-07-25 (16 CPUs, int8, 47.4 s de audio): tiny 1.3 s · base 2.2 s · small 5.4 s.
⚠️ La primera corrida de cada modelo incluye la DESCARGA en `carga_modelo_s` — para el número honesto
de carga hay que correrlo dos veces y quedarse con el segundo.

Lo que este banco NO mide (ver §1.8): es `faster-whisper` (CTranslate2), **no `whisper.cpp`**; voz
**sintética**, no la del operador por su micrófono; entrada WAV, **no el `audio/mp4` que graba la app**.
"""

import os
import sys
import time

from faster_whisper import WhisperModel

model_size = sys.argv[1] if len(sys.argv) > 1 else "base"
audio = sys.argv[2] if len(sys.argv) > 2 else "dictado.wav"

t0 = time.time()
model = WhisperModel(model_size, device="cpu", compute_type="int8")
t_load = time.time() - t0

t1 = time.time()
segments, info = model.transcribe(audio, language="es", beam_size=5)
# `segments` es un generador perezoso: la transcripción corre recién acá, dentro del cronómetro.
text = " ".join(s.text.strip() for s in segments)
t_infer = time.time() - t1

print(f"MODEL={model_size}")
print(f"carga_modelo_s={t_load:.1f}")
print(f"inferencia_s={t_infer:.1f}")
print(f"audio_s={info.duration:.1f}")
print(f"factor_tiempo_real={info.duration / t_infer:.2f}x")
print(f"cpus={os.cpu_count()}")
print(f"TEXTO: {text}")
