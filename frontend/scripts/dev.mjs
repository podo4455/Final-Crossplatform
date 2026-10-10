import { spawn } from 'node:child_process'

// The course notes use TMPDIR=/tmp on macOS. Apply it without a
// shell-specific command so the same npm script works on Windows.
const env = { ...process.env }
if (process.platform === 'darwin') env.TMPDIR = '/tmp'
const child = spawn(process.execPath, ['node_modules/nuxt/bin/nuxt.mjs', 'dev', ...process.argv.slice(2)], {
  stdio: 'inherit', env,
})
child.on('error', (error) => { console.error(error); process.exitCode = 1 })
child.on('exit', (code) => { process.exitCode = code ?? 1 })
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => child.kill(signal))
}
