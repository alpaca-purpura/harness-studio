import { Button, useAppStore } from "@/shared";

// PLACEHOLDER de primera versión (vacía). Reserva las 3 zonas del shell firmado (HS-03 it.13,
// mockups/arnesia-shell-A-galaxia.html): Command Rail izq. · lienzo casi-fullscreen · dock de
// conversación invocable (⌘K). El shell REAL se construye en la próxima sesión (widgets/features).
export function App() {
  const theme = useAppStore((s) => s.theme);
  const toggleTheme = useAppStore((s) => s.toggleTheme);

  return (
    <div className="flex h-full w-full">
      {/* Command Rail (zona reservada — vacía por ahora) */}
      <aside className="flex w-14 flex-none flex-col items-center gap-3 border-r border-sidebar-border bg-sidebar py-3">
        <div className="size-8 rounded-md bg-sidebar-primary" title="ArnesIA" />
      </aside>

      {/* Lienzo (zona reservada — vacía por ahora) */}
      <main className="relative flex flex-1 flex-col items-center justify-center gap-4 p-8">
        <h1 className="text-2xl font-semibold text-foreground">ArnesIA</h1>
        <p className="max-w-md text-center text-sm text-muted-foreground">
          Scaffold vacío listo. El shell (Command Rail · lienzo · dock de conversación) se construye
          en la próxima sesión.
        </p>
        <Button variant="outline" size="sm" onClick={toggleTheme}>
          Tema: {theme}
        </Button>
      </main>
    </div>
  );
}
