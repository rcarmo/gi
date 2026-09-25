// Add mobile dismissal controls only; source components stay byte-identical.
// Geometry lives in the host CSS and has a pinned external reference.
function replace(source,from,to){
 if(source.split(from).length!==2)throw new Error(`Picker geometry adapter anchor changed: ${from}`);
 return source.replace(from,to);
}
export function patchPickerGeometry(source){
 if(source.includes('class="gi-picker-close"'))throw new Error('Picker geometry adapter already applied');
 source=replace(source,'                            <div class="compose-model-popup-title">Select model</div>',`                            <button type="button" class="gi-picker-close" aria-label="Close model picker" onClick=\${() => { setShowModelPopup(false); requestAnimationFrame(() => { if (document.activeElement === document.body) modelHintRef.current?.focus(); }); }}>Close</button>
                            <div class="compose-model-popup-title">Select model</div>`);
 return replace(source,'                            <div class="compose-model-popup-title">Manage sessions & agents</div>',`                            <button type="button" class="gi-picker-close" aria-label="Close session picker" onClick=\${() => closeSessionPopup(true)}>Close</button>
                            <div class="compose-model-popup-title">Manage sessions & agents</div>`);
}
