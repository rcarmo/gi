// Minimal live RFB 3.8 fixture. Only accepts the messages used by the viewer.
import net from 'node:net';
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const root = process.env.GI_TEST_RUN_ROOT;
const mode = process.env.GI_VNC_VIEWER_MODE || 'readonly';
if (!root || !process.env.GI_FIXTURE_BIN || !['readonly', 'interactive', 'empty'].includes(mode)) {
  throw Error('Use make test-vnc-viewer');
}
const state = join(root, 'vnc-viewer', mode);
const guard = Bun.spawn(['bash', '-c', 'source scripts/project-test-env.sh && gi_test_path "$1/workspace" && gi_test_path "$1/home"', 'vnc-path-guard', state], { stdout: 'inherit', stderr: 'inherit' });
if (await guard.exited) process.exit(1);
mkdirSync(join(state, 'workspace'), { recursive: true });
mkdirSync(join(state, 'home'), { recursive: true });
const counter = join(state, 'input-frames');
let inputFrames = 0;
writeFileSync(counter, '0');
const sockets = new Set();
const rfb = net.createServer(socket => {
  sockets.add(socket);
  socket.on('close', () => sockets.delete(socket));
  socket.on('error', () => socket.destroy());
  socket.write('RFB 003.008\n');
  let pending = Buffer.alloc(0), phase = 0;
  socket.on('data', data => {
    pending = Buffer.concat([pending, data]);
    if (pending.length > 65536) { socket.destroy(); return; }
    while (pending.length) {
      if (phase === 0) {
        if (pending.length < 12) return;
        pending = pending.subarray(12);
        socket.write(Buffer.from([1, 1])); // one security type: None
        phase = 1;
      } else if (phase === 1) {
        if (pending.length < 1) return;
        if (pending[0] !== 1) { socket.destroy(); return; }
        pending = pending.subarray(1);
        socket.write(Buffer.alloc(4)); // security result: OK
        phase = 2;
      } else if (phase === 2) {
        pending = pending.subarray(1); // shared ClientInit
        const init = Buffer.alloc(24);
        init.writeUInt16BE(2, 0); init.writeUInt16BE(2, 2);
        init[4] = 32; init[5] = 24; init[7] = 1;
        init.writeUInt16BE(255, 8); init.writeUInt16BE(255, 10); init.writeUInt16BE(255, 12);
        init[14] = 16; init[15] = 8;
        init.writeUInt32BE(7, 20);
        socket.write(Buffer.concat([init, Buffer.from('fixture')]));
        phase = 3;
      } else {
        let length;
        const type = pending[0];
        if (type === 0) length = 20; // SetPixelFormat (viewer uses 32-bit little endian)
        else if (type === 2) { if (pending.length < 4) return; length = 4 + 4 * pending.readUInt16BE(2); }
        else if (type === 3) length = 10;
        else if (type === 4) length = 8;
        else if (type === 5) length = 6;
        else if (type === 6) { if (pending.length < 8) return; length = 8 + pending.readUInt32BE(4); }
        else { socket.destroy(); return; }
        if (length > 65536) { socket.destroy(); return; }
        if (pending.length < length) return;
        pending = pending.subarray(length);
        if (type === 4 || type === 5 || type === 6) writeFileSync(counter, String(++inputFrames));
        if (type === 3) {
          const frame = Buffer.alloc(32); // update header + raw rectangle header + 4 pixels
          frame.writeUInt16BE(1, 2);
          frame.writeUInt16BE(2, 8); frame.writeUInt16BE(2, 10);
          for (let i = 16; i < 32; i += 4) { frame[i] = 40; frame[i + 1] = 120; frame[i + 2] = 200; }
          socket.write(frame);
        }
      }
    }
  });
});
await new Promise((resolve, reject) => { rfb.once('error', reject); rfb.listen(0, '127.0.0.1', resolve); });
// Allocate an available HTTP port; fixed ports are an optional debugging override.
const probe = net.createServer();
await new Promise((resolve, reject) => { probe.once('error', reject); probe.listen(0, '127.0.0.1', resolve); });
const port = Number(process.env.GI_VNC_TEST_PORT || probe.address().port);
await new Promise(resolve => probe.close(resolve));
let child, cli;
const stop = () => { cli?.kill('SIGTERM'); child?.kill('SIGTERM'); for (const socket of sockets) socket.destroy(); };
const terminate = () => { stop(); process.exit(143); };
process.once('SIGTERM', terminate); process.once('SIGINT', terminate);
try {
  child = Bun.spawn([process.env.GI_FIXTURE_BIN, '-web', '-bind', '127.0.0.1', '-port', String(port), '-db', join(state, 'state.db'), '-workspace', join(state, 'workspace')], {
    env: { ...process.env, HOME: join(state, 'home'), FIXTURE_MODEL_URL: 'http://127.0.0.1:19999/v1',
      GI_WEB_VNC_TARGETS: JSON.stringify(mode === 'empty' ? [] : [{ id: 'fixture', label: 'Live RFB fixture', host: '127.0.0.1', port: rfb.address().port, readOnly: mode === 'readonly' }]),
      GI_WEB_VNC_ALLOW_DIRECT: 'false' }, stdout: 'inherit', stderr: 'inherit',
  });
  let ready = false;
  for (let i = 0; i < 100; i++) {
    if (child.exitCode !== null) throw Error('Viewer fixture exited before readiness');
    try { if ((await fetch(`http://127.0.0.1:${port}/api/sessions`)).ok) { ready = true; break; } } catch {}
    await Bun.sleep(100);
  }
  if (!ready) throw Error('Viewer fixture did not start');
  cli = Bun.spawn(['bash', 'scripts/run-playwright.sh', 'test', '--config', 'playwright.vnc.config.ts'], {
    env: { ...process.env, GI_TEST_URL: `http://127.0.0.1:${port}`, GI_VNC_VIEWER_MODE: mode, GI_VNC_INPUT_COUNTER: counter },
    stdout: 'inherit', stderr: 'inherit',
  });
  process.exitCode = await cli.exited;
} finally {
  stop();
  if (cli) await cli.exited;
  if (child) await child.exited;
  await new Promise(resolve => rfb.close(resolve));
  process.removeListener('SIGTERM', terminate); process.removeListener('SIGINT', terminate);
}
