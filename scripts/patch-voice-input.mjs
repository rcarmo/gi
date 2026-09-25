// Explicit browser capability only; no changes to supplied composer bytes.
export function patchVoiceInput(source){
 if(source.includes('useGiVoiceInput'))throw Error('Voice input adapter already applied');
 let result=source;
 const replace=(from,to)=>{if(result.split(from).length!==2)throw Error(`Voice input adapter anchor changed: ${from.slice(0,90)}`);result=result.replace(from,to);};
 replace('    const giComposeSurface = useGiComposeSurface(textareaRef);',`    const giComposeSurface = useGiComposeSurface(textareaRef);
    const giVoice = useGiVoiceInput(() => ({ owner: currentChatJid, content, searchMode, disabled: statusNotice?.type === 'compaction', setContent, resize: giComposeSurface.resize,
        onStart: () => { setShowModelPopup(false); setShowSessionPopup(false); setShowSlash(false); setShowMention(false); }
    }), textareaRef);`);
 replace('    const handleInput = (e) => {', '    const handleInput = (e) => {\n        giVoice.cancel();');
 replace('    const handleSubmit = async (overrideContent, submitMode, submitOptions = {}) => {', '    const handleSubmit = async (overrideContent, submitMode, submitOptions = {}) => {\n        giVoice.cancel();');
 // Add after key ownership guard so consumed/composing events remain untouched.
 replace('        if (declineComposeKey(e)) return;','        if (declineComposeKey(e)) return;\n        if (giVoice.beforeKey(e)) return;');
 replace('            ${showQueueStack && !searchMode && html`','            ${giVoice.status}\n            ${showQueueStack && !searchMode && html`');
 replace('                    ${notificationsAvailable && !searchMode && html`','                    ${giVoice.button}\n                    ${notificationsAvailable && !searchMode && html`');
 return `import { useGiVoiceInput } from '../gi-voice-input.js';\n${result}`;
}
