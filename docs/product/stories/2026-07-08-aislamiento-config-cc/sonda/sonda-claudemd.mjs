// Sonda dirigida HS-17 RF-162: pregunta puntual por el contenido de std-spec (CLAUDE.md
// propio del cwd del arnés) para confirmar que --setting-sources project,local NO se lo
// lleva puesto. Mismo protocolo que sonda.mjs, pregunta distinta.
import { spawn } from 'node:child_process'
import { writeFileSync } from 'node:fs'

const [cwd, outFile, ...rest] = process.argv.slice(2)
const sepIdx = rest.indexOf('--')
const claudeArgs = rest.slice(sepIdx + 1)
const PROBE =
  'Citá textualmente, si lo tenés cargado en tu CLAUDE.md de contexto, qué campos debe ' +
  'tener todo spec.md según la regla de la banda Base "std-spec" de este arnés. Si NO lo ' +
  'tenés cargado, decilo explícitamente — no inventes.'

const child = spawn('claude', claudeArgs, { cwd, stdio: ['pipe', 'pipe', 'pipe'] })
let resultText = ''
let sawInit = false
let buf = ''
child.stdout.on('data', (d) => {
  buf += d.toString()
  let idx
  while ((idx = buf.indexOf('\n')) >= 0) {
    const line = buf.slice(0, idx)
    buf = buf.slice(idx + 1)
    if (!line.trim()) continue
    let f
    try {
      f = JSON.parse(line)
    } catch {
      continue
    }
    if (f.type === 'control_response' && !sawInit) {
      sawInit = true
      child.stdin.write(
        `${JSON.stringify({ type: 'user', message: { role: 'user', content: [{ type: 'text', text: PROBE }] } })}\n`,
      )
    }
    if (f.type === 'control_request' && f.request?.subtype === 'can_use_tool') {
      child.stdin.write(
        `${JSON.stringify({
          type: 'control_response',
          response: {
            subtype: 'success',
            request_id: f.request_id,
            response: { behavior: 'deny', message: 'sonda de solo lectura', toolUseID: f.request.tool_use_id },
          },
        })}\n`,
      )
    }
    if (f.type === 'assistant' && f.message?.content) {
      for (const b of f.message.content) if (b.type === 'text') resultText += b.text
    }
    if (f.type === 'result') child.stdin.end()
  }
})
child.stdin.write(
  `${JSON.stringify({ type: 'control_request', request_id: 'req_1_sonda', request: { subtype: 'initialize' } })}\n`,
)
setTimeout(() => child.kill('SIGKILL'), 60_000)
child.on('close', () => {
  writeFileSync(outFile, resultText)
  console.log(resultText)
})
