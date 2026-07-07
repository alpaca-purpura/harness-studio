package doctrina

import "embed"

// WebDist carries the built SPA (web/dist) so the daemon can serve the UI itself —
// «arnesia serve = API + UI embebida» deja de ser promesa (deuda HS-10: el usage
// mentía). El dist NO se versiona (solo su .gitkeep, para que el embed compile
// siempre): en dev la UI la sirve vite/Tauri y este FS va casi vacío; el bundle de
// release (scripts/bundle.sh) compila la SPA ANTES del daemon, así el binario
// instalado la lleva dentro. El daemon decide en runtime: sin index.html embebido →
// responde honesto que la UI no viaja en este build.
//
//go:embed all:web/dist
var WebDist embed.FS
