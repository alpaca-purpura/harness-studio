import { useEffect } from "react"
import { ShellPage } from "@/pages/shell"
import { useSessions } from "@/shared"

// App boots the shell: it loads the session registry, opens the multiplexed Dock
// stream, and binds ⌘K to invoke the conversation. Everything else is composed by
// ShellPage (mockup it.14).
export function App() {
  const init = useSessions((s) => s.init)
  const toggleChat = useSessions((s) => s.toggleChat)

  useEffect(() => {
    void init()
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault()
        toggleChat()
      }
    }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [init, toggleChat])

  return <ShellPage />
}
