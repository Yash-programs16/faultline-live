import { createServer } from 'node:http'
import { createReadStream } from 'node:fs'
import { fileURLToPath } from 'node:url'

const page = fileURLToPath(new URL('../preview/faultline-ui.html', import.meta.url))
const server = createServer((request, response) => {
  const path = new URL(request.url, 'http://127.0.0.1:5187').pathname
  if (path.startsWith('/api/')) {
    response.writeHead(503, { 'Content-Type': 'application/json', 'Cache-Control': 'no-store' })
    response.end(JSON.stringify({ error: 'UI preview only. The backend is not connected.' }))
    return
  }
  if (path === '/favicon.ico') { response.writeHead(204); response.end(); return }
  response.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store' })
  const stream = createReadStream(page)
  stream.on('error', () => response.destroy())
  stream.pipe(response)
})
server.on('error', error => {
  console.error(`Cannot start UI preview: ${error.message}`)
  process.exitCode = 1
})
server.listen(5187, '127.0.0.1', () => {
  console.log('Faultline Live UI preview: http://127.0.0.1:5187')
  console.log('Backend is not connected. No operational data is fabricated.')
})
