import { useRef, useState } from "react"
import Markdown, { type Components } from "react-markdown"
import rehypeHighlight from "rehype-highlight"
import remarkGfm from "remark-gfm"

// Md renderiza el markdown del asistente (CH-D4/D4b): GFM + sintaxis coloreada con los
// tokens vigentes (chat.css, sin theme externo); inline-code que parece RUTA → chip
// clickeable que copia; bloque de código → rótulo de lenguaje + botón copiar.
export function Md({ children }: { children: string }) {
  return (
    <div className="md">
      <Markdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeHighlight]}
        components={componentes}
      >
        {children}
      </Markdown>
    </div>
  )
}

// esRuta: token sin espacios con «/» o extensión corta. Un falso positivo solo pinta un
// chip copiable — jamás rompe nada.
function esRuta(s: string): boolean {
  if (s.length > 120 || /\s/.test(s)) return false
  return s.includes("/") || /\.[a-z0-9]{1,5}$/i.test(s)
}

const Code: Components["code"] = ({ node: _node, className, children, ...rest }) => {
  const texto = typeof children === "string" ? children : null
  if (!className && texto && esRuta(texto)) return <FileChip ruta={texto} />
  return (
    <code className={className} {...rest}>
      {children}
    </code>
  )
}

function FileChip({ ruta }: { ruta: string }) {
  const [copiada, setCopiada] = useState(false)
  return (
    <button
      type="button"
      className="md-file"
      title={copiada ? "ruta copiada" : "copiar ruta"}
      onClick={() =>
        void navigator.clipboard?.writeText(ruta).then(() => {
          setCopiada(true)
          window.setTimeout(() => setCopiada(false), 1200)
        })
      }
    >
      <span className="flex-none" aria-hidden>
        {copiada ? "✓" : "⧉"}
      </span>
      {ruta}
    </button>
  )
}

const Pre: Components["pre"] = ({ node: _node, children, ...rest }) => {
  const hijo = Array.isArray(children) ? children[0] : children
  const cls = (hijo as { props?: { className?: string } } | null)?.props?.className ?? ""
  const lang = /language-([\w-]+)/.exec(cls)?.[1] ?? ""
  return <PreBlock lang={lang} pre={<pre {...rest}>{children}</pre>} />
}

function PreBlock({ lang, pre }: { lang: string; pre: React.ReactNode }) {
  const bodyRef = useRef<HTMLDivElement>(null)
  const [copiado, setCopiado] = useState(false)
  return (
    <div className="md-pre">
      <div className="md-pre-head">
        <span>{lang || "código"}</span>
        <button
          type="button"
          onClick={() => {
            const texto = bodyRef.current?.innerText ?? ""
            void navigator.clipboard?.writeText(texto).then(() => {
              setCopiado(true)
              window.setTimeout(() => setCopiado(false), 1200)
            })
          }}
        >
          {copiado ? "✓ copiado" : "copiar"}
        </button>
      </div>
      <div ref={bodyRef}>{pre}</div>
    </div>
  )
}

const componentes: Components = { code: Code, pre: Pre }
