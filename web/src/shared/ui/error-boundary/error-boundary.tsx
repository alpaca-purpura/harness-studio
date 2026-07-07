import { Component, type ReactNode } from "react"
import { Button } from "../button"

// ErrorBoundary is the app's ONLY render-error net (reconciliación honesta,
// arch/contracts/nomenclatura-arnes.md §4.5: a broken render must be VISIBLE — never a white
// screen, never silence). Class component because React only exposes render-error capture via
// getDerivedStateFromError. The fallback is a chrome-styled panel on semantic tokens (zero
// magic values); «Reintentar» clears the caught error and re-renders the children. React
// itself logs the caught error (onCaughtError default), so the console keeps the full stack.

interface ErrorBoundaryProps {
  children: ReactNode
  // Context shown in the panel (e.g. the id of the arnés whose render failed).
  label?: string | undefined
}

interface ErrorBoundaryState {
  error: Error | null
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  override state: ErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error }
  }

  private readonly reset = (): void => {
    this.setState({ error: null })
  }

  override render(): ReactNode {
    const { error } = this.state
    if (error === null) return this.props.children
    return (
      <div
        role="alert"
        className="flex h-full w-full flex-col items-center justify-center p-10 text-center"
      >
        <div className="flex max-w-md flex-col items-center gap-3 rounded-lg border border-destructive bg-card p-6">
          <div className="text-lg font-semibold text-card-foreground">
            El lienzo no pudo renderizar este arnés
          </div>
          {this.props.label ? (
            <span className="text-sm text-muted-foreground">{this.props.label}</span>
          ) : null}
          <code className="max-w-full overflow-x-auto rounded-md bg-muted px-3 py-2 text-left font-mono text-xs text-muted-foreground">
            {error.message}
          </code>
          <Button variant="outline" size="sm" onClick={this.reset}>
            Reintentar
          </Button>
        </div>
      </div>
    )
  }
}
