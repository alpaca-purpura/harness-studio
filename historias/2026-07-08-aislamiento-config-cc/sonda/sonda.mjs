// Sonda desechable para HS-17 (D6): spawnea `claude` con el argv exacto que se le pasa
// (mimetiza SpawnArgs de conductor.go, sin re-implementarlo — cada fase pasa su propio
// argv), envía el turno-sonda, y vuelca la respuesta a stdout + a un archivo.
// Uso: node sonda.mjs <cwd> <label> <archivo-salida> -- <argv claude...>
import { spawn } from 'node:child_process'
import { writeFileSync } from 'node:fs'

const [cwd, label, outFile, ...rest] = process.argv.slice(2)
const sepIdx = rest.indexOf('--')
if (sepIdx === -1) {
  console.error('uso: node sonda.mjs <cwd> <label> <outFile> -- <argv claude...>')
  process.exit(2)
}
const claudeArgs = rest.slice(sepIdx + 1)

const PROBE =
  'Lista exhaustivamente qué skills, subagentes, comandos y servidores MCP tenés ' +
  'disponibles ahora mismo, agrupados por origen si podés inferirlo (ej. kit propio del ' +
  'arnés vs plugins/marketplaces del operador vs MCP de cuenta). No ejecutes ninguna ' +
  'herramienta, no leas archivos — respondé solo desde lo que ya sabés de tu propio ' +
  'contexto/system prompt.'

const bin = process.env.CLAUDE_BIN || 'claude'
const child = spawn(bin, claudeArgs, { cwd, stdio: ['pipe', 'pipe', 'pipe'], env: process.env })

let stdout = ''
let stderr = ''
let resultText = ''
let sawInit = false
let ctrlSeq = 0

child.on('error', (e) => {
  console.error(`[${label}] spawn error: ${e.message}`)
})
child.stderr.on('data', (d) => {
  stderr += d.toString()
  if (process.env.SONDA_DEBUG) process.stderr.write(d)
})

function send(obj) {
  child.stdin.write(`${JSON.stringify(obj)}\n`)
}

function sendInitialize() {
  ctrlSeq += 1
  send({
    type: 'control_request',
    request_id: `req_${ctrlSeq}_sonda`,
    request: { subtype: 'initialize' },
  })
}

function sendTurn(text) {
  send({
    type: 'user',
    message: { role: 'user', content: [{ type: 'text', text }] },
  })
}

let buf = ''
child.stdout.on('data', (d) => {
  buf += d.toString()
  let idx
  // eslint-disable-next-line no-cond-assign
  while ((idx = buf.indexOf('\n')) >= 0) {
    const line = buf.slice(0, idx)
    buf = buf.slice(idx + 1)
    if (!line.trim()) continue
    stdout += `${line}\n`
    let f
    try {
      f = JSON.parse(line)
    } catch {
      continue
    }
    if (process.env.SONDA_DEBUG) console.error(`[${label}] frame: ${f.type}/${f.subtype ?? ''}`)
    // El ack de nuestro control_request:initialize (frame "control_response") es la
    // señal fiable de que el canal está armado — el frame "system/init" puede llegar
    // después de hooks SessionStart lentos y esperarlo bloquea el envío del turno.
    if (f.type === 'control_response' && !sawInit) {
      sawInit = true
      sendTurn(PROBE)
    }
    if (f.type === 'control_request' && f.request?.subtype === 'can_use_tool') {
      // Sonda de solo lectura: denegamos cualquier intento de herramienta.
      send({
        type: 'control_response',
        response: {
          subtype: 'success',
          request_id: f.request_id,
          response: {
            behavior: 'deny',
            message: 'sonda de solo lectura, no ejecutar herramientas',
            toolUseID: f.request.tool_use_id,
          },
        },
      })
    }
    if (f.type === 'assistant' && f.message?.content) {
      for (const b of f.message.content) {
        if (b.type === 'text') resultText += b.text
      }
    }
    if (f.type === 'result') {
      child.stdin.end()
    }
  }
})

sendInitialize()

const timeout = setTimeout(() => {
  console.error(`[${label}] TIMEOUT tras 90s — matando proceso`)
  child.kill('SIGKILL')
}, 90_000)

child.on('close', (code) => {
  clearTimeout(timeout)
  const report = [
    `# Sonda: ${label}`,
    '',
    `argv: claude ${claudeArgs.join(' ')}`,
    `cwd: ${cwd}`,
    `exit code: ${code}`,
    '',
    '## Respuesta del modelo',
    '',
    resultText || '(vacía)',
    '',
    '## stderr',
    '',
    '```',
    stderr.trim() || '(vacío)',
    '```',
  ].join('\n')
  writeFileSync(outFile, report)
  console.log(report)
  console.log(`\n[${label}] volcado en ${outFile}`)
})
