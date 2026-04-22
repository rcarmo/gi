// @ts-nocheck

import { zlibSync } from 'fflate';

import { loadRemoteDisplayWasmDecoder } from './panes/remote-display-decoder.js';
import { disposeSocketBoundaryBestEffort } from './panes/pane-runtime-safety.js';
import { WebSocketRemoteDisplayBoundary } from './panes/remote-display-socket.js';
import { VncRemoteDisplayProtocol } from './panes/remote-display-vnc.js';

const DEFAULT_TARGET = '192.168.1.10:5917';
const DEFAULT_PASSWORD = 'cd8a99cd';
const DEFAULT_ENCODING_SEQUENCE = '16,5,2,1,0,-223';

const encoder = new TextEncoder();

function bytes(...values) {
    return Uint8Array.from(values);
}

function parseBoolean(value, fallback = false) {
    if (value == null) return fallback;
    const text = String(value).trim().toLowerCase();
    if (!text) return fallback;
    return text === '1' || text === 'true' || text === 'yes' || text === 'on';
}

function nowIso() {
    return new Date().toISOString();
}

function parseEncodingCandidates(raw) {
    const base = typeof raw === 'string'
        ? raw
            .split(',')
            .map((value) => value.trim())
            .filter((value) => value.length > 0)
        : Array.isArray(raw)
            ? raw.map((value) => String(value).trim()).filter((value) => value.length > 0)
            : [];
    const seen = new Set();
    const values = [];
    for (const item of base) {
        const numeric = Number(item);
        if (!Number.isFinite(numeric)) continue;
        if (seen.has(numeric)) continue;
        seen.add(numeric);
        values.push(numeric);
    }
    if (values.length > 0) return values;
    return parseEncodingCandidates(DEFAULT_ENCODING_SEQUENCE);
}

function encodingListToString(values) {
    if (Array.isArray(values)) {
        return values.map((value) => String(value)).join(',');
    }
    if (typeof values === 'string') {
        const parsed = parseEncodingCandidates(values);
        return parsed.join(',');
    }
    return '';
}

function toDefaultWsUrl(targetId) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/vnc/ws?target=${encodeURIComponent(targetId)}`;
}

function toAbsoluteWsUrl(raw, targetId) {
    const template = String(raw || '').trim();
    if (!template) return toDefaultWsUrl(targetId);
    const resolved = template.replaceAll('{target}', encodeURIComponent(targetId));
    if (/^wss?:\/\//i.test(resolved)) return resolved;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    if (resolved.startsWith('//')) return `${protocol}${resolved}`;
    if (resolved.startsWith('/')) return `${protocol}//${window.location.host}${resolved}`;
    return `${protocol}//${window.location.host}/${resolved}`;
}

function buildServerInit({ width, height, name }) {
    const nameBytes = encoder.encode(name);
    const buffer = new ArrayBuffer(24 + nameBytes.length);
    const view = new DataView(buffer);
    view.setUint16(0, width, false);
    view.setUint16(2, height, false);
    view.setUint8(4, 32);
    view.setUint8(5, 24);
    view.setUint8(6, 0);
    view.setUint8(7, 1);
    view.setUint16(8, 255, false);
    view.setUint16(10, 255, false);
    view.setUint16(12, 255, false);
    view.setUint8(14, 16);
    view.setUint8(15, 8);
    view.setUint8(16, 0);
    view.setUint32(20, nameBytes.length, false);
    const payload = new Uint8Array(buffer);
    payload.set(nameBytes, 24);
    return payload;
}

function createSurface(width, height) {
    const safeWidth = Math.max(1, Math.floor(Number(width || 0)));
    const safeHeight = Math.max(1, Math.floor(Number(height || 0)));
    return {
        width: safeWidth,
        height: safeHeight,
        pixels: new Uint8ClampedArray(safeWidth * safeHeight * 4),
    };
}

function setPixel(surface, x, y, rgba) {
    if (!surface) return;
    if (x < 0 || y < 0 || x >= surface.width || y >= surface.height) return;
    const idx = (y * surface.width + x) * 4;
    surface.pixels[idx] = rgba[0] || 0;
    surface.pixels[idx + 1] = rgba[1] || 0;
    surface.pixels[idx + 2] = rgba[2] || 0;
    surface.pixels[idx + 3] = rgba[3] == null ? 255 : rgba[3];
}

function getPixel(surface, x, y) {
    if (!surface || x < 0 || y < 0 || x >= surface.width || y >= surface.height) return [0, 0, 0, 0];
    const idx = (y * surface.width + x) * 4;
    return [
        surface.pixels[idx],
        surface.pixels[idx + 1],
        surface.pixels[idx + 2],
        surface.pixels[idx + 3],
    ];
}

function applyRectsToSurface(surface, event) {
    if (!surface || event?.type !== 'framebuffer-update') return surface;
    let current = surface;

    // Pipeline mode: WASM has the full framebuffer
    if (event.framebuffer && event.framebuffer.length > 0) {
        for (const rect of event.rects || []) {
            if (rect.kind === 'resize') {
                current = createSurface(rect.width, rect.height);
            }
        }
        // Overwrite with the pipeline framebuffer
        const fb = event.framebuffer;
        if (fb.length === current.width * current.height * 4) {
            current.pixels.set(fb);
        }
        return current;
    }

    // Non-pipeline mode: per-rect RGBA blitting
    for (const rect of event.rects || []) {
        if (rect.kind === 'resize') {
            current = createSurface(rect.width, rect.height);
            continue;
        }
        if (rect.kind === 'rgba') {
            for (let row = 0; row < rect.height; row += 1) {
                for (let col = 0; col < rect.width; col += 1) {
                    const srcIdx = (row * rect.width + col) * 4;
                    setPixel(current, rect.x + col, rect.y + row, [
                        rect.rgba[srcIdx],
                        rect.rgba[srcIdx + 1],
                        rect.rgba[srcIdx + 2],
                        rect.rgba[srcIdx + 3],
                    ]);
                }
            }
            continue;
        }
        if (rect.kind === 'copy') {
            const copied = [];
            for (let row = 0; row < rect.height; row += 1) {
                for (let col = 0; col < rect.width; col += 1) {
                    copied.push(getPixel(current, rect.srcX + col, rect.srcY + row));
                }
            }
            let i = 0;
            for (let row = 0; row < rect.height; row += 1) {
                for (let col = 0; col < rect.width; col += 1) {
                    setPixel(current, rect.x + col, rect.y + row, copied[i++]);
                }
            }
        }
    }
    return current;
}

class HarnessUi {
    constructor() {
        this.statusEl = document.querySelector('[data-status]');
        this.metaEl = document.querySelector('[data-meta]');
        this.eventsEl = document.querySelector('[data-events]');
        this.liveSummaryEl = document.querySelector('[data-live-summary]');
        this.syntheticEl = document.querySelector('[data-synthetic]');
        this.canvas = document.querySelector('[data-display-canvas]');
        this.placeholderEl = document.querySelector('[data-placeholder]');
        this.targetEl = document.querySelector('[data-target]');
        this.passwordEl = document.querySelector('[data-password]');
        this.connectBtn = document.querySelector('[data-connect]');
        this.disconnectBtn = document.querySelector('[data-disconnect]');
        this.syntheticBtn = document.querySelector('[data-run-synthetic]');
        this.ctx = this.canvas?.getContext?.('2d', { alpha: false }) || null;
        if (this.ctx) this.ctx.imageSmoothingEnabled = false;
    }

    setStatus(text) {
        if (this.statusEl) this.statusEl.textContent = String(text || '');
    }

    setMeta(text) {
        if (this.metaEl) this.metaEl.textContent = String(text || '');
    }

    setLiveSummary(text) {
        if (this.liveSummaryEl) this.liveSummaryEl.textContent = String(text || '');
    }

    setSynthetic(text) {
        if (this.syntheticEl) this.syntheticEl.textContent = String(text || '');
    }

    appendEvent(line) {
        if (!this.eventsEl) return;
        const safe = String(line || '');
        const existing = this.eventsEl.textContent || '';
        const lines = `${existing}${existing ? '\n' : ''}${safe}`.split('\n');
        this.eventsEl.textContent = lines.slice(-400).join('\n');
        this.eventsEl.scrollTop = this.eventsEl.scrollHeight;
    }

    ensureCanvas(width, height, reveal = false) {
        if (!this.canvas || !this.ctx) return;
        const safeWidth = Math.max(1, Math.floor(Number(width || 0)));
        const safeHeight = Math.max(1, Math.floor(Number(height || 0)));
        if (this.canvas.width !== safeWidth || this.canvas.height !== safeHeight) {
            this.canvas.width = safeWidth;
            this.canvas.height = safeHeight;
        }
        if (reveal) {
            this.canvas.style.display = 'block';
            if (this.placeholderEl) this.placeholderEl.style.display = 'none';
        }
    }

    drawRect(rect, fbWidth, fbHeight) {
        if (!this.ctx || !this.canvas) return;
        this.ensureCanvas(fbWidth, fbHeight, true);
        const img = new ImageData(rect.rgba, rect.width, rect.height);
        this.ctx.putImageData(img, rect.x, rect.y);
    }

    copyRect(rect, fbWidth, fbHeight) {
        if (!this.ctx || !this.canvas) return;
        this.ensureCanvas(fbWidth, fbHeight, true);
        const img = this.ctx.getImageData(rect.srcX, rect.srcY, rect.width, rect.height);
        this.ctx.putImageData(img, rect.x, rect.y);
    }
}

class VncHarness {
    constructor() {
        this.url = new URL(window.location.href);
        this.target = this.url.searchParams.get('target') || DEFAULT_TARGET;
        this.password = this.url.searchParams.get('password') || DEFAULT_PASSWORD;
        this.wsTemplate = this.url.searchParams.get('ws') || '';
        this.preferredEncodings = String(this.url.searchParams.get('encodings') || '').trim();
        this.autoConnect = parseBoolean(this.url.searchParams.get('autoconnect'), true);
        this.runSyntheticOnLoad = parseBoolean(this.url.searchParams.get('runSynthetic'), false);
        this.ui = new HarnessUi();
        this.protocol = null;
        this.socketBoundary = null;
        this.connectedAt = null;
        this.firstDisplayInitAt = null;
        this.firstFramebufferAt = null;
        this.lastError = null;
        this.currentEncoding = 'default';
        this.fallbackEncodings = parseEncodingCandidates(this.preferredEncodings);
        this.fallbackEnabled = false;
        this.fallbackTimer = null;
        this.fallbackArmed = false;
        this.syntheticResults = [];
        this.stats = {
            startedAt: nowIso(),
            bytesIn: 0,
            bytesOut: 0,
            frames: 0,
            rects: 0,
            clipboardEvents: 0,
            bellEvents: 0,
            protocolEvents: {},
            logs: [],
        };

        this.ui.targetEl.value = this.target;
        this.ui.passwordEl.value = this.password;
        this.ui.connectBtn?.addEventListener('click', () => {
            this.target = String(this.ui.targetEl?.value || '').trim() || DEFAULT_TARGET;
            this.password = String(this.ui.passwordEl?.value || '');
            void this.connect();
        });
        this.ui.disconnectBtn?.addEventListener('click', () => this.disconnect());
        this.ui.syntheticBtn?.addEventListener('click', () => {
            const result = this.runSyntheticSuite();
            this.renderSyntheticSummary(result);
        });

        this.log(`Harness booted. target=${this.target}${this.wsTemplate ? ` ws=${this.wsTemplate}` : ''}`);
        if (this.autoConnect) {
            void this.connect();
        }
        if (this.runSyntheticOnLoad) {
            const result = this.runSyntheticSuite();
            this.renderSyntheticSummary(result);
        }
        this.renderMeta();
    }

    log(message) {
        const line = `[${nowIso()}] ${message}`;
        this.stats.logs.push(line);
        if (this.stats.logs.length > 500) this.stats.logs.splice(0, this.stats.logs.length - 500);
        this.ui.appendEvent(line);
    }

    renderMeta(extra = '') {
        const protocolState = this.protocol?.state ? `state=${this.protocol.state}` : 'state=idle';
        const firstFrame = this.firstFramebufferAt ? `firstFrame=${this.firstFramebufferAt}` : 'firstFrame=pending';
        const summary = `target=${this.target} · ${protocolState} · encoding=${this.currentEncoding || 'default'} · bytesIn=${this.stats.bytesIn} · bytesOut=${this.stats.bytesOut} · frames=${this.stats.frames} · rects=${this.stats.rects} · ${firstFrame}${extra ? ` · ${extra}` : ''}`;
        this.ui.setMeta(summary);
    }

    resetRunState(options = {}) {
        const opts = { resetStats: true, ...options };
        this.firstDisplayInitAt = null;
        this.firstFramebufferAt = null;
        this.connectedAt = null;
        this.lastError = null;
        if (opts.resetStats) {
            this.stats.bytesIn = 0;
            this.stats.bytesOut = 0;
            this.stats.frames = 0;
            this.stats.rects = 0;
            this.stats.clipboardEvents = 0;
            this.stats.bellEvents = 0;
            this.stats.protocolEvents = {};
        }
    }

    summarizeLive() {
        const lines = [
            `target=${this.target}`,
            `connectedAt=${this.connectedAt || 'n/a'}`,
            `firstDisplayInitAt=${this.firstDisplayInitAt || 'n/a'}`,
            `firstFramebufferAt=${this.firstFramebufferAt || 'n/a'}`,
            `bytesIn=${this.stats.bytesIn}`,
            `bytesOut=${this.stats.bytesOut}`,
            `frames=${this.stats.frames}`,
            `rects=${this.stats.rects}`,
            `lastError=${this.lastError || 'none'}`,
        ];
        this.ui.setLiveSummary(lines.join(' · '));
    }

    renderSyntheticSummary(result) {
        const summary = result?.summary || {};
        const lines = [
            `Synthetic encoders: ${summary.passed || 0}/${summary.total || 0} passed`,
            ...((result?.cases || []).map((item) => `${item.pass ? '✅' : '❌'} ${item.name}${item.detail ? ` — ${item.detail}` : ''}`)),
        ];
        this.ui.setSynthetic(lines.join('\n'));
    }

    async connect(preferredEncodings = null, options = {}) {
        const opts = {
            fallbackEnabled: true,
            resetRun: true,
            ...options,
        };
        this.disconnect();
        const target = String(this.target || '').trim();
        if (!target) {
            this.ui.setStatus('Missing target');
            this.renderMeta('missing-target');
            return;
        }

        const wsUrl = toAbsoluteWsUrl(this.wsTemplate, target);
        this.ui.setStatus(`Connecting to ${target}…`);
        this.log(`Connecting websocket to ${wsUrl}`);

        const wasmDecoder = await loadRemoteDisplayWasmDecoder();
        const protocolOptions = wasmDecoder
            ? {
                pipeline: wasmDecoder,
                decodeRawRect: (chunk, width, height, pixelFormat) => wasmDecoder.decodeRawRectToRgba(chunk, width, height, pixelFormat),
            }
            : {};
        const requested = preferredEncodings == null ? this.preferredEncodings : String(preferredEncodings);
        const explicitSequence = Array.isArray(opts.fallbackEncodings) ? opts.fallbackEncodings : null;
        const sequence = explicitSequence == null ? parseEncodingCandidates(requested) : explicitSequence;
        const activeEncoding = sequence[0] == null ? null : String(sequence[0]);
        if (this.password) protocolOptions.password = this.password;
        if (activeEncoding != null) {
            protocolOptions.encodings = activeEncoding;
        }

        this.currentEncoding = activeEncoding || 'default';
        this.fallbackEnabled = opts.fallbackEnabled !== false && sequence.length > 1;
        this.fallbackEncodings = sequence;
        this.fallbackArmed = this.fallbackEnabled;
        this.fallbackTimer = null;
        this.resetRunState({ resetStats: opts.resetRun });

        this.protocol = new VncRemoteDisplayProtocol(protocolOptions);

        this.socketBoundary = new WebSocketRemoteDisplayBoundary({
            url: wsUrl,
            binaryType: 'arraybuffer',
            onOpen: () => {
                this.connectedAt = nowIso();
                this.ui.setStatus('WebSocket proxy connected. Waiting for VNC handshake…');
                this.log('websocket.open');
                this.renderMeta();
                this.socketBoundary?.sendControl?.({ type: 'ping' });
            },
            onMetrics: (metrics) => {
                this.stats.bytesIn = Number(metrics?.bytesIn || 0);
                this.stats.bytesOut = Number(metrics?.bytesOut || 0);
                this.renderMeta();
            },
            onMessage: (message) => {
                void this.handleMessage(message);
            },
            onClose: () => {
                if (this.fallbackTimer) {
                    clearTimeout(this.fallbackTimer);
                    this.fallbackTimer = null;
                }
                this.fallbackEnabled = false;
                this.fallbackArmed = false;
                this.log('websocket.close');
                this.ui.setStatus('WebSocket closed.');
                this.renderMeta('closed');
                this.summarizeLive();
            },
            onError: () => {
                this.lastError = 'socket-error';
                if (this.fallbackTimer) {
                    clearTimeout(this.fallbackTimer);
                    this.fallbackTimer = null;
                }
                this.fallbackEnabled = false;
                this.fallbackArmed = false;
                this.log('websocket.error');
                this.ui.setStatus('WebSocket error.');
                this.renderMeta('socket-error');
                this.summarizeLive();
            },
        });
        this.socketBoundary.connect();
    }

    disconnect() {
        if (this.fallbackTimer) {
            clearTimeout(this.fallbackTimer);
            this.fallbackTimer = null;
        }
        this.fallbackEnabled = false;
        this.fallbackArmed = false;
        disposeSocketBoundaryBestEffort(this.socketBoundary);
        this.socketBoundary = null;
    }

    async runEncodingProbe(encoding) {
        const candidates = parseEncodingCandidates(encoding);
        const activeEncoding = candidates[0] == null ? null : String(candidates[0]);
        await this.connect(activeEncoding, { fallbackEnabled: false, resetRun: true });
        return this.currentEncoding;
    }

    async runEncodingProbeWithWait(encoding, waitMs, options = {}) {
        const opts = { fallbackEnabled: true, ...options };
        const requested = parseEncodingCandidates(encoding);
        const activeEncoding = requested[0] == null ? null : String(requested[0]);
        const encodedRequest = encodingListToString(requested);
        await this.connect(encodedRequest, { fallbackEnabled: opts.fallbackEnabled, resetRun: true });
        const started = Date.now();
        while (Date.now() - started < waitMs) {
            if (this.firstFramebufferAt || this.lastError) break;
            await new Promise((resolvePromise) => window.setTimeout(resolvePromise, 350));
        }
        return {
            ...this.snapshot(),
            requestedEncoding: activeEncoding,
            probeSequence: requested,
        };
    }

    scheduleRawFallback() {
        this.scheduleFallback('first-frame timeout');
    }

    scheduleFallback(reason = 'first-frame timeout') {
        if (!this.fallbackEnabled || this.fallbackEncodings.length <= 1) return;
        if (this.fallbackTimer) {
            clearTimeout(this.fallbackTimer);
            this.fallbackTimer = null;
        }
        this.fallbackTimer = setTimeout(() => {
            if (this.firstFramebufferAt || !this.fallbackEnabled || this.fallbackEncodings.length <= 1) return;
            this.tryNextFallbackEncoding(reason);
        }, 2200);
    }

    tryNextFallbackEncoding(reason = 'protocol retry') {
        if (!this.fallbackEnabled || this.fallbackEncodings.length <= 1) return;
        const current = String(this.fallbackEncodings.shift());
        const next = String(this.fallbackEncodings[0]);
        if (!next) {
            this.fallbackEnabled = false;
            this.fallbackArmed = false;
            return;
        }
        const remaining = [...this.fallbackEncodings];
        this.log(`enc-switch attempt due ${reason}: ${current || 'default'} -> ${next}`);
        void this.connect(String(next), {
            fallbackEnabled: remaining.length > 1,
            resetRun: true,
            fallbackEncodings: remaining,
        });
    }

    applyProtocolEvent(event) {
        if (!event) return;
        const type = String(event.type || 'unknown');
        this.stats.protocolEvents[type] = Number(this.stats.protocolEvents[type] || 0) + 1;

        if (type === 'protocol-version') {
            this.ui.setStatus(`Protocol ${event.server} → ${event.client}`);
            this.log(`protocol-version server=${event.server} client=${event.client}`);
            this.renderMeta();
            return;
        }
        if (type === 'security-types') {
            this.ui.setStatus(`Security types: ${(event.types || []).join(', ') || 'none'}`);
            this.log(`security-types ${JSON.stringify(event.types || [])}`);
            this.renderMeta();
            return;
        }
        if (type === 'security-selected') {
            this.ui.setStatus(`Security selected: ${event.label}`);
            this.log(`security-selected ${event.label}`);
            this.renderMeta();
            return;
        }
        if (type === 'security-result') {
            this.ui.setStatus('Security negotiation complete.');
            this.log('security-result ok=true');
            this.renderMeta();
            return;
        }
        if (type === 'display-init') {
            if (!this.firstDisplayInitAt) this.firstDisplayInitAt = nowIso();
            this.ui.ensureCanvas(event.width, event.height, false);
            this.ui.setStatus(`Display init ${event.width}×${event.height}${event.name ? ` (${event.name})` : ''}`);
            this.log(`display-init width=${event.width} height=${event.height} name=${event.name || ''}`);
            this.renderMeta('awaiting-framebuffer');
            this.summarizeLive();
            this.scheduleRawFallback();
            return;
        }
        if (type === 'framebuffer-update') {
            if (this.fallbackTimer) {
                clearTimeout(this.fallbackTimer);
                this.fallbackTimer = null;
            }
            let painted = false;

            // Pipeline mode: WASM owns the full framebuffer
            if (event.framebuffer && event.framebuffer.length > 0) {
                for (const rect of event.rects || []) {
                    if (rect.kind === 'resize') {
                        this.ui.ensureCanvas(rect.width, rect.height, false);
                    }
                }
                // Paint full framebuffer to canvas
                this.ui.ensureCanvas(event.width, event.height, true);
                if (this.ui.ctx && event.width > 0 && event.height > 0) {
                    const img = new ImageData(new Uint8ClampedArray(event.framebuffer), event.width, event.height);
                    this.ui.ctx.putImageData(img, 0, 0);
                    painted = true;
                }
            } else {
                // Non-pipeline: per-rect rendering
                for (const rect of event.rects || []) {
                    if (rect.kind === 'resize') {
                        this.ui.ensureCanvas(rect.width, rect.height, false);
                    } else if (rect.kind === 'copy') {
                        this.ui.copyRect(rect, event.width, event.height);
                        painted = true;
                    } else if (rect.kind === 'rgba') {
                        this.ui.drawRect(rect, event.width, event.height);
                        painted = true;
                    }
                }
            }
            if (painted && !this.firstFramebufferAt) this.firstFramebufferAt = nowIso();
            this.stats.frames += 1;
            this.stats.rects += Array.isArray(event.rects) ? event.rects.length : 0;
            this.ui.setStatus(painted
                ? `Framebuffer update applied (${(event.rects || []).length} rects)`
                : `Framebuffer update with no paintable rects (${(event.rects || []).length} rects)`);
            this.log(`framebuffer-update rects=${(event.rects || []).length} painted=${painted ? 1 : 0} size=${event.width}x${event.height}`);
            this.renderMeta(painted ? 'rendering' : 'awaiting-framebuffer');
            if (!painted && !this.firstFramebufferAt) {
                this.scheduleRawFallback();
            }
            this.summarizeLive();
            return;
        }
        if (type === 'clipboard') {
            this.stats.clipboardEvents += 1;
            this.log(`clipboard len=${String(event.text || '').length}`);
            this.renderMeta();
            return;
        }
        if (type === 'bell') {
            this.stats.bellEvents += 1;
            this.log('bell');
            this.renderMeta();
            return;
        }
    }

    async handleMessage(message) {
        if (message?.kind === 'control') {
            const payload = message.payload;
            const type = payload?.type || 'control';
            if (type === 'vnc.error') {
                this.lastError = payload?.error || 'vnc.error';
                this.log(`control.vnc.error ${this.lastError}`);
                this.ui.setStatus(`Proxy error: ${this.lastError}`);
                this.renderMeta('proxy-error');
                this.summarizeLive();
                return;
            }
            if (type === 'vnc.connected') {
                this.log(`control.vnc.connected target=${payload?.target?.id || ''}`);
                this.ui.setStatus(`Connected to proxy for ${payload?.target?.label || this.target}.`);
                this.renderMeta();
                return;
            }
            if (type === 'pong') {
                this.log('control.pong');
                return;
            }
            this.log(`control.${type}`);
            return;
        }

        if (!this.protocol) return;
        try {
            const chunk = message.data instanceof Blob ? await message.data.arrayBuffer() : message.data;
            const result = this.protocol.receive(chunk);
            for (const outgoing of result.outgoing || []) {
                this.socketBoundary?.send?.(outgoing);
            }
            for (const event of result.events || []) {
                this.applyProtocolEvent(event);
            }
        } catch (error) {
            this.lastError = error?.message || String(error || 'Unknown protocol error');
            this.log(`protocol.error ${this.lastError}`);
            this.ui.setStatus(`Protocol error: ${this.lastError}`);
            this.renderMeta('protocol-error');
            this.summarizeLive();
            if (this.fallbackEnabled && this.fallbackEncodings.length > 1 && /unexpected eof|zlib|decompress|protocol|buffer/i.test(this.lastError || '')) {
                this.tryNextFallbackEncoding(`protocol-error@${this.currentEncoding}`);
            }
        }
    }

    runSyntheticSuite() {
        const cases = [];

        const addCase = (name, fn) => {
            try {
                const detail = fn();
                cases.push({ name, pass: true, detail: detail || '' });
            } catch (error) {
                cases.push({ name, pass: false, detail: error?.message || String(error || 'Unknown error') });
            }
        };

        const expectPixel = (surface, x, y, rgba, label) => {
            const actual = getPixel(surface, x, y);
            const expected = Array.from(rgba);
            if (actual.length !== expected.length || actual.some((v, i) => v !== expected[i])) {
                throw new Error(`${label}: expected [${expected.join(', ')}], got [${actual.join(', ')}]`);
            }
        };

        const bootstrap = (width = 4, height = 4, options = {}) => {
            const protocol = new VncRemoteDisplayProtocol(options);
            protocol.receive(encoder.encode('RFB 003.008\n'));
            protocol.receive(bytes(1, 1));
            protocol.receive(bytes(0, 0, 0, 0));
            const init = protocol.receive(buildServerInit({ width, height, name: 'Harness synthetic' }));
            let surface = createSurface(width, height);
            for (const event of init.events || []) {
                if (event.type === 'display-init') {
                    surface = createSurface(event.width, event.height);
                }
            }
            return { protocol, surface };
        };

        addCase('raw', () => {
            const state = bootstrap(2, 2);
            const update = state.protocol.receive(bytes(
                0, 0, 0, 1,
                0, 0, 0, 0,
                0, 1, 0, 1,
                0, 0, 0, 0,
                0x00, 0x00, 0xff, 0x00,
            ));
            for (const event of update.events || []) state.surface = applyRectsToSurface(state.surface, event);
            expectPixel(state.surface, 0, 0, [255, 0, 0, 255], 'raw pixel');
            return '1x1 raw red pixel decoded';
        });

        addCase('rre', () => {
            const state = bootstrap(4, 4);
            const update = state.protocol.receive(bytes(
                0, 0, 0, 1,
                0, 0, 0, 0,
                0, 4, 0, 4,
                0, 0, 0, 2,
                0, 0, 0, 1,
                0xff, 0x00, 0x00, 0x00,
                0x00, 0x00, 0xff, 0x00,
                0, 1, 0, 1,
                0, 2, 0, 2,
            ));
            for (const event of update.events || []) state.surface = applyRectsToSurface(state.surface, event);
            expectPixel(state.surface, 0, 0, [0, 0, 255, 255], 'rre background');
            expectPixel(state.surface, 1, 1, [255, 0, 0, 255], 'rre subrect');
            return 'RRE background + subrect decoded';
        });

        addCase('copyrect', () => {
            const state = bootstrap(4, 2);
            const seed = state.protocol.receive(bytes(
                0, 0, 0, 1,
                0, 0, 0, 0,
                0, 2, 0, 2,
                0, 0, 0, 0,
                0x00, 0x00, 0xff, 0x00,
                0x00, 0xff, 0x00, 0x00,
                0xff, 0x00, 0x00, 0x00,
                0x00, 0x00, 0xff, 0x00,
            ));
            for (const event of seed.events || []) state.surface = applyRectsToSurface(state.surface, event);
            const copy = state.protocol.receive(bytes(
                0, 0, 0, 1,
                0, 2, 0, 0,
                0, 2, 0, 2,
                0, 0, 0, 1,
                0, 0, 0, 0,
            ));
            for (const event of copy.events || []) state.surface = applyRectsToSurface(state.surface, event);
            expectPixel(state.surface, 2, 0, [255, 0, 0, 255], 'copyrect pixel');
            expectPixel(state.surface, 3, 1, [255, 0, 0, 255], 'copyrect bottom-right');
            return 'CopyRect copied 2x2 block';
        });

        addCase('hextile', () => {
            const state = bootstrap(4, 4);
            const update = state.protocol.receive(bytes(
                0, 0, 0, 1,
                0, 0, 0, 0,
                0, 4, 0, 4,
                0, 0, 0, 5,
                0x0e,
                0xff, 0x00, 0x00, 0x00,
                0x00, 0x00, 0xff, 0x00,
                0x01,
                0x11,
                0x11,
            ));
            for (const event of update.events || []) state.surface = applyRectsToSurface(state.surface, event);
            expectPixel(state.surface, 0, 0, [0, 0, 255, 255], 'hextile background');
            expectPixel(state.surface, 1, 1, [255, 0, 0, 255], 'hextile subrect');
            return 'Hextile background + single subrect decoded';
        });

        addCase('zrle', () => {
            const state = bootstrap(2, 2);
            const zrleTile = bytes(
                0x01,
                0x00, 0x00, 0xff, 0x00,
            );
            const compressed = zlibSync(zrleTile);
            const length = compressed.length;
            const update = state.protocol.receive(Uint8Array.from([
                0, 0, 0, 1,
                0, 0, 0, 0,
                0, 2, 0, 2,
                0, 0, 0, 16,
                (length >>> 24) & 0xff,
                (length >>> 16) & 0xff,
                (length >>> 8) & 0xff,
                length & 0xff,
                ...compressed,
            ]));
            for (const event of update.events || []) state.surface = applyRectsToSurface(state.surface, event);
            expectPixel(state.surface, 0, 0, [255, 0, 0, 255], 'zrle pixel');
            return `ZRLE decoded compressed tile (${length} byte payload)`;
        });

        addCase('desktop-size', () => {
            const state = bootstrap(2, 2);
            const update = state.protocol.receive(bytes(
                0, 0, 0, 1,
                0, 0, 0, 0,
                0, 3, 0, 4,
                255, 255, 255, 33,
            ));
            for (const event of update.events || []) state.surface = applyRectsToSurface(state.surface, event);
            if (state.surface.width !== 3 || state.surface.height !== 4) {
                throw new Error(`desktop-size expected 3x4, got ${state.surface.width}x${state.surface.height}`);
            }
            return 'DesktopSize resized framebuffer to 3x4';
        });

        const passed = cases.filter((item) => item.pass).length;
        const report = {
            at: nowIso(),
            cases,
            summary: {
                total: cases.length,
                passed,
                failed: cases.length - passed,
            },
        };
        this.syntheticResults = report.cases;
        this.log(`synthetic-suite done ${report.summary.passed}/${report.summary.total} passed`);
        return report;
    }

    snapshot() {
        return {
            target: this.target,
            connectedAt: this.connectedAt,
            firstDisplayInitAt: this.firstDisplayInitAt,
            firstFramebufferAt: this.firstFramebufferAt,
            lastError: this.lastError,
            protocolState: this.protocol?.state || 'idle',
            currentEncoding: this.currentEncoding || 'default',
            fallbackQueue: this.fallbackEncodings?.slice?.(0) || [],
            stats: {
                ...this.stats,
            },
            syntheticResults: this.syntheticResults,
        };
    }
}

const harness = new VncHarness();
window.__VNC_HARNESS__ = {
    connect: (encodings, options = {}) => harness.connect(encodings, options),
    disconnect: () => harness.disconnect(),
    runSyntheticSuite: () => harness.runSyntheticSuite(),
    runEncodingProbeWithWait: (encoding, waitMs) => harness.runEncodingProbeWithWait(encoding, waitMs),
    snapshot: () => harness.snapshot(),
};
