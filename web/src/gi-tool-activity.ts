// One authoritative occurrence in the existing status area; no output pane.
import { html, useEffect, useState } from './vendor/preact-htm.js';

export function toolElapsed(tool, now = Date.now()) {
    const start = Date.parse(tool?.started_at);
    const ms = tool?.state === 'running' ? now - start : tool?.duration_ms;
    if (typeof ms !== 'number' || !Number.isFinite(ms) || ms < 0) return '?';
    return `${Math.floor(ms / 1000)}s`;
}
export function ToolActivity({tool}) {
    const [now, setNow] = useState(Date.now);
    useEffect(() => {
        if (tool?.state !== 'running') return;
        setNow(Date.now());
        const timer = setInterval(() => setNow(Date.now()), 1000);
        return () => clearInterval(timer);
    }, [tool?.turn_id, tool?.start_seq, tool?.state]);
    if (!tool) return null;
    const label = {running:'Running',completed:'Completed',failed:'Failed',cancelled:'Cancelled',aborted:'Aborted',interrupted:'Interrupted'}[tool.state] || 'Unknown';
    const terminal = tool.state !== 'running';
    return html`<div class="agent-status-panel gi-tool-activity" data-tool-call-id=${tool.tool_call_id} data-tool-state=${tool.state} data-turn-id=${tool.turn_id}>
        <span class=${terminal ? 'gi-tool-glyph' : 'spinner'} aria-hidden="true">${terminal ? tool.state === 'completed' ? '✓' : '✕' : ''}</span>
        <div class="agent-status-text" aria-label=${`${tool.name}: ${label}`}>
            <span>${label}: ${tool.name}</span>
            ${tool.preview && html`<code class="agent-tool-argument" title=${tool.preview}>${tool.preview}</code>`}
        </div>
        <span class="gi-tool-elapsed" aria-label=${terminal ? 'Tool duration' : 'Tool elapsed'}>${toolElapsed(tool, now)}</span>
    </div>`;
}
