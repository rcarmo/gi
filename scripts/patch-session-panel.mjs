// Adapt session presentation only. Mutation, search and focus ownership stay
// with the supplied component and existing host guards.
export function patchSessionPanel(source){
 let result=source;
 const replace=(from,to)=>{if(result.split(from).length!==2)throw Error(`Session panel anchor changed: ${from.slice(0,90)}`);result=result.replace(from,to);};
 replace(`<button type="button" class="gi-picker-close" aria-label="Close session picker" onClick=\${() => closeSessionPopup(true)}>Close</button>
                            <div class="compose-model-popup-title">Manage sessions & agents</div>`, `<div class="compose-session-popup-header">
                                <label class="compose-model-popup-title compose-session-search-heading" for="gi-session-search">Search sessions</label>
                                <button type="button" class="compose-session-popup-close" aria-label="Close session picker" onClick=\${() => closeSessionPopup(true)}>×</button>
                            </div>`);
 // Query changes must settle the initial highlight before a visible row can
 // receive another key. A passive effect can overwrite that first navigation.
 replace(`    useEffect(() => {
        if (!showSessionPopup) return;
        const preferred = resolveSessionPickerSearchInitialIndex`, `    useLayoutEffect(() => {
        if (!showSessionPopup) return;
        const preferred = resolveSessionPickerSearchInitialIndex`);
 replace('                                ref=${sessionSearchRef}\n                                type="search"','                                ref=${sessionSearchRef}\n                                id="gi-session-search"\n                                type="search"');
 replace('<div id="compose-session-results" class="compose-model-popup-menu"','<div id="compose-session-results" class="compose-model-popup-menu compose-session-popup-results"');
 replace('<div class="compose-session-section-label">${group.label}</div>','<div class="compose-session-section-heading" role="presentation">${group.label}</div>');
 const pin=`                                                \${!archived && chat.capabilities?.pin !== false && typeof onPinSession === 'function' && html\`
                                                    <button type="button" class="compose-model-popup-btn" disabled=\${Boolean(sessionMutationPending)}
                                                        aria-label=\${\`\${chat.pinned ? 'Unpin' : 'Pin'} @\${chat.agent_name}\`}
                                                        onClick=\${() => { void runSessionMutation(chat, 'pin', !chat.pinned); }}>\${chat.pinned ? 'Unpin' : 'Pin'}</button>
                                                \`}
`;
 replace(pin,'');
 const row='<div key=${chat.chat_jid} data-session-jid=${chat.chat_jid} class=${`compose-model-popup-item-row${archived ? \' archived\' : \'\'}`}>';
 replace(row,row+`
                                            \${!archived && chat.capabilities?.pin !== false && typeof onPinSession === 'function' ? html\`
                                                <button type="button" class=\${\`compose-session-row-pin\${chat.pinned ? ' pinned' : ''}\`} disabled=\${Boolean(sessionMutationPending)}
                                                    aria-label=\${\`\${chat.pinned ? 'Unpin' : 'Pin'} @\${chat.agent_name}\`} aria-pressed=\${chat.pinned ? 'true' : 'false'}
                                                    onClick=\${() => { void runSessionMutation(chat, 'pin', !chat.pinned); }}>\${chat.pinned ? '★' : '☆'}</button>
                                            \` : html\`<span class="compose-session-row-pin-spacer" aria-hidden="true"></span>\`}
`);
 replace('                                                aria-current=${chat.chat_jid === currentChatJid ? \'true\' : undefined}', '                                                aria-label=${label}\n                                                aria-current=${chat.chat_jid === currentChatJid ? \'true\' : undefined}');
 replace(`                                                \${label}
                                            </button>`,`                                                <span class="compose-session-row-content">
                                                    <span class="compose-session-row-main">
                                                        <span class="compose-session-row-label">\${normalizeHandle(chat.agent_name) || chat.chat_jid}</span>
                                                        <span class="compose-session-row-meta"><span class="compose-session-row-jid">\${chat.chat_jid}</span>\${(chat.model || chat.model_label) && html\`<span> · \${chat.model || chat.model_label}</span>\`}</span>
                                                    </span>
                                                    <span class="compose-session-row-pills">
                                                        \${chat.chat_jid === currentChatJid && html\`<span class="compose-session-status-pill current">current</span>\`}
                                                        \${archived ? html\`<span class="compose-session-status-pill archived">archived</span>\` : chat.is_active && html\`<span class="compose-session-status-pill active">active</span>\`}
                                                    </span>
                                                </span>
                                            </button>`);
 return `import { normalizeHandle } from '../ui/branch-lifecycle.js';\n${result}`;
}
