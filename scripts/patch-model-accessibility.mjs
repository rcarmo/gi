// Pinned Classic composite-widget semantics without changing supplied sources.
export function patchModelAccessibility(source){
 let result=source;
 const replace=(from,to)=>{if(result.split(from).length!==2)throw Error(`Model accessibility anchor changed: ${from.slice(0,90)}`);result=result.replace(from,to);};
 replace('    const modelPopupRef = useRef(null);',`    const modelPopupRef = useRef(null);
    const modelPanelIdRef = useRef(null);
    if (!modelPanelIdRef.current) modelPanelIdRef.current = nextModelPanelId();`);
 replace('for="gi-model-search"','for=${modelPanelIdRef.current + "-search"}');
 replace('<input id="gi-model-search" type="search" class="compose-model-catalogue-search" aria-label="Search models" placeholder="Search models…"',`<input id=\${modelPanelIdRef.current + '-search'} type="search" role="combobox" class="compose-model-catalogue-search" aria-label="Search models" placeholder="Search models…"
                                        aria-autocomplete="list" aria-expanded="true" aria-controls=\${modelPanelIdRef.current + '-results'}
                                        aria-activedescendant=\${!loadingModels && visibleModels[modelPopupIndex] && !modelEntries[modelPopupIndex]?.disabled ? modelPanelOptionId(modelPanelIdRef.current, visibleModels[modelPopupIndex].label) : undefined}`);
 replace('<div class="compose-model-popup-menu compose-model-catalogue-results" role="menu" aria-label="Model picker">','<div id=${modelPanelIdRef.current + "-results"} class="compose-model-popup-menu compose-model-catalogue-results" role="listbox" tabIndex="-1" aria-label="Models" aria-busy=${loadingModels ? "true" : "false"}>');
 replace('                                            data-model-index=${index}','                                            data-model-index=${index}\n                                            id=${modelPanelOptionId(modelPanelIdRef.current, modelLabel)}');
 replace('                                            role="menuitem"\n                                            class=${`compose-model-catalogue-option',`                                            role="option"
                                            tabIndex="-1"
                                            aria-selected=\${current ? 'true' : 'false'}
                                            aria-disabled=\${switchingModel || blocked ? 'true' : 'false'}
                                            aria-describedby=\${blocked ? modelPanelOptionId(modelPanelIdRef.current, modelLabel) + '-blocked' : undefined}
                                            onMouseDown=\${event => event.preventDefault()}
                                            class=\${\`compose-model-catalogue-option`);
 replace('<span class="compose-model-catalogue-option-note">Context window is smaller than the latest measured request.</span>', '<span id=${modelPanelOptionId(modelPanelIdRef.current, modelLabel) + "-blocked"} class="compose-model-catalogue-option-note">Context window is smaller than the latest measured request.</span>');
 replace('                                    aria-label="Open model picker"',`                                    aria-label="Open model picker"
                                    aria-haspopup="listbox" aria-expanded=\${showModelPopup ? 'true' : 'false'}
                                    aria-controls=\${showModelPopup ? modelPanelIdRef.current + '-results' : undefined}`);
 replace(`            emitModelState(state);
            setShowModelPopup(false);`, `            const ownedModelFocus = modelPopupRef.current?.contains(document.activeElement);
            const modelOpener = modelHintRef.current;
            emitModelState(state);
            setShowModelPopup(false);
            if (ownedModelFocus) requestAnimationFrame(() => {
                if (mountedRef.current && modelOpener?.isConnected && document.activeElement === document.body && !settingsOwnsKeyboard()) modelOpener.focus({ preventScroll: true });
            });`);
 return `import { nextModelPanelId, modelPanelOptionId } from '../gi-model-accessibility.js';\n${result}`;
}
