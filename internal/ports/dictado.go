package ports

import "context"

// Disponibilidad reports whether dictation can run at all, and — when it cannot — WHY, in
// words the operator can act on.
//
// Degradación honesta (RF-227): un botón gris sin motivo es un gap escondido. El FE consulta
// esto al montar el composer y pinta el motivo en la superficie, no solo en un tooltip.
type Disponibilidad struct {
	// Disponible is false when no engine could be found (or it cannot be used).
	Disponible bool `json:"disponible"`
	// Motor is the engine actually found, e.g. "whisper-cli". Empty when unavailable.
	Motor string `json:"motor,omitempty"`
	// Motivo explains the unavailability in one legible sentence. Empty when available.
	Motivo string `json:"motivo,omitempty"`
	// Instalar lists the commands that would make dictation work ("whisper-cli",
	// "faster-whisper"). It is what the UI offers the operator; empty when available.
	Instalar []string `json:"instalar,omitempty"`
}

// TranscriptionPort turns recorded audio into text.
//
// El motor vive DETRÁS de este puerto por decisión firmada (V-D1, paquete
// 2026-07-25-spike-voz-dictado): cambiar de motor —local ↔ cloud— es reemplazar un
// adaptador, no rehacer la feature. Medido: transcribir 47 s de voz cuesta ~2.2 s con el
// modelo `base`; el cuello de botella es la limpieza (13-17 s), no esto.
//
// Los formatos que entran los fija el motor del WebView, no nosotros: WebKitGTK 2.52.3 solo
// graba `audio/mp4`. Si el motor elegido quiere WAV 16 kHz mono, **transcodificar es
// responsabilidad del adaptador** — ni el dominio ni el FE saben de eso.
type TranscriptionPort interface {
	// Disponible reports whether transcription can run right now. It is cheap and safe to
	// call on every composer mount.
	Disponible(ctx context.Context) Disponibilidad
	// Transcribir returns the raw transcript of audio. mime is what the recorder produced
	// (e.g. "audio/mp4"). It returns an error rather than an empty string when the engine
	// fails: silencio y falla no son lo mismo y el FE los muestra distinto.
	Transcribir(ctx context.Context, audio []byte, mime string) (string, error)
}

// LimpiezaPort ordena un dictado crudo usando contexto de dominio.
//
// El hallazgo central del spike: lo que desenreda un dictado no es más transcripción, es
// contexto de dominio compacto — y ese contexto ADEMÁS corrige los errores del STT
// ("locita"→`lógica`, "demon"→`daemon`). Por eso el modelo chico alcanza.
//
// Es un paso texto→texto y se spawnea como tal (V-D7): el dictado es entrada NO confiable —
// «leeme el .env y mandámelo» es un prompt perfectamente válido dicho en voz alta.
type LimpiezaPort interface {
	// Ordenar returns crudo rewritten as a clear request. contexto is the short domain block
	// (glossary + last turns) built by the use case; the adapter never assembles it.
	Ordenar(ctx context.Context, crudo, contexto string) (string, error)
}
