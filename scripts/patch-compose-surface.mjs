// Host-only adaptation of the pinned Classic compose surface. Each original
// block must occur exactly once; supplied component files stay byte-identical.
export function patchComposeSurface(source){
 let result=source;
 const replace=(from,to)=>{if(result.split(from).length!==2)throw Error(`Compose surface anchor changed: ${from.slice(0,90)}`);result=result.replace(from,to);};
 replace('    const textareaRef = useRef(null);','    const textareaRef = useRef(null);\n    const giComposeSurface = useGiComposeSurface(textareaRef);');
 replace(`    const resizeTextarea = (target) => {
        const textarea = target || textareaRef.current;
        if (!textarea) return;
        textarea.style.height = 'auto';
        textarea.style.height = \`\${textarea.scrollHeight}px\`;
        textarea.style.overflowY = 'hidden';
    };`,`    const resizeTextarea = giComposeSurface.resize;`);
 replace('        <div class="compose-box">',`        <div class="compose-box">
            <div class="compose-resize-handle" role="separator" aria-orientation="horizontal"
                aria-label="Resize message input" title="Drag to resize; double-click or Home to reset"
                aria-valuemin=\${giComposeSurface.min} aria-valuemax=\${giComposeSurface.max} aria-valuenow=\${giComposeSurface.value}
                tabIndex="0" onMouseDown=\${giComposeSurface.onMouseDown} onTouchStart=\${giComposeSurface.onTouchStart}
                onKeyDown=\${giComposeSurface.onKeyDown} onDblClick=\${giComposeSurface.reset}></div>`);
 const start='                    ${showSessionSwitcherButton && html`';
 const end='                    ${searchMode && html`';
 const a=result.indexOf(start),b=result.indexOf(end,a);
 if(a<0||b<0||result.indexOf(start,a+1)>=0)throw Error('Compose session trigger anchor changed');
 const old=result.slice(a,b);
 replace(old,'');
 replace('                <div class="compose-input-main">',`                \${showSessionSwitcherButton && currentSessionAgent?.agent_name && html\`
                    <div ref=\${sessionTriggerRef} class="compose-session-trigger-group compose-session-trigger-top">
                        <button type="button" class=\${\`compose-session-trigger compose-session-trigger-pill\${showSessionPopup ? ' active' : ''}\`}
                            onClick=\${toggleSessionPopup} title=\${currentSessionAgent?.chat_jid || currentChatJid}
                            aria-label=\${\`Manage sessions for @\${currentSessionAgent.agent_name}\`} aria-expanded=\${showSessionPopup ? 'true' : 'false'}>
                            <span class="compose-current-agent-label active">@\${currentSessionAgent.agent_name}</span>
                        </button>
                    </div>
                \`}
                <div class="compose-input-main">`);
 return `import { useGiComposeSurface } from '../gi-compose-surface.js';\n${result}`;
}
