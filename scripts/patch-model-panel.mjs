// Guarded host layout adaptation; this follows existing model/popup adapters.
export function patchModelPanel(source){
 let result=source;
 const replace=(from,to)=>{if(result.split(from).length!==2)throw Error(`Model panel anchor changed: ${from.slice(0,90)}`);result=result.replace(from,to);};
 replace('filterModelOptions(modelOptions, modelQuery, getModelPickerOptionSearchLabel)', 'modelPanelRows(filterModelOptions(modelOptions, modelQuery, getModelPickerOptionSearchLabel), activeModel)');
 replace('[modelOptions, modelQuery]);','[modelOptions, modelQuery, activeModel]);');
 replace('<div class="compose-model-popup" ref=${modelPopupRef}', '<div class="compose-model-popup compose-model-catalogue" ref=${modelPopupRef}');
 const start='                            <button type="button" class="gi-picker-close" aria-label="Close model picker"';
 const finish='                            <div class="compose-model-popup-menu" role="menu" aria-label="Model picker">';
 const a=result.indexOf(start),b=result.indexOf(finish,a);
 if(a<0||b<0||result.indexOf(start,a+1)>=0)throw Error('Model panel header anchor changed');
 replace(result.slice(a,b),`                            <div class="compose-model-catalogue-header">
                                <div class="gi-model-panel-heading"><label class="compose-model-catalogue-search-label" for="gi-model-search">Search models</label>
                                    <button type="button" class="gi-picker-close" aria-label="Close model picker" onClick=\${() => { setShowModelPopup(false); requestAnimationFrame(() => { if (document.activeElement === document.body) modelHintRef.current?.focus(); }); }}>Close</button>
                                </div>
                                <div class="compose-model-catalogue-search-row">
                                    <input id="gi-model-search" type="search" class="compose-model-catalogue-search" aria-label="Search models" placeholder="Search models…"
                                        value=\${modelQuery} onInput=\${event => { popupTypeaheadRef.current = { value: '', updatedAt: 0 }; setModelQuery(event.currentTarget.value); }} />
                                    \${modelQuery && html\`<button type="button" class="compose-model-catalogue-clear" aria-label="Clear model search" onClick=\${() => { setModelQuery(''); modelPopupRef.current?.querySelector('input')?.focus(); }}>×</button>\`}
                                </div>
                                <div class="compose-model-catalogue-summary" aria-live="polite"><span>\${visibleModels.length} \${visibleModels.length === 1 ? 'model' : 'models'}</span>\${loadingModels && html\`<span>Refreshing…</span>\`}</div>
                            </div>
`);
 replace(finish, '                            <div class="compose-model-popup-menu compose-model-catalogue-results" role="menu" aria-label="Model picker">');
 replace(`                                    return html\`
                                        <button
                                            key=\${modelLabel}`,`                                    const current = activeModel === modelLabel;
                                    const sectionStart = index === 0 || (visibleModels[index - 1]?.label === activeModel) !== current;
                                    const sectionCount = current ? 1 : visibleModels.filter(option => option.label !== activeModel).length;
                                    return html\`
                                        \${sectionStart && html\`<div class="compose-model-catalogue-section-heading" role="presentation"><span>\${current ? 'Current' : 'Available models'}</span><span>\${sectionCount}</span></div>\`}
                                        <button
                                            key=\${modelLabel}`);
 replace("class=${`compose-model-popup-item compose-model-popup-model-item${modelPopupIndex === index ? ' active' : ''}${activeModel === modelLabel ? ' current-model' : ''}`}","class=${`compose-model-catalogue-option compose-model-popup-model-item${modelPopupIndex === index ? ' active focused' : ''}${activeModel === modelLabel ? ' current-model selected' : ''}${blocked ? ' blocked' : ''}`}\n                                            aria-label=${formatModelPickerDisplayLabel(modelLabel, modelOption?.contextWindow)}");
 replace('<span class="compose-model-popup-model-label">${formatModelPickerDisplayLabel(modelLabel, modelOption?.contextWindow)}</span>',`<span class="gi-model-pin-unavailable" aria-hidden="true" title="Model pinning is not available"></span>
                                            <span class="compose-model-catalogue-option-content">
                                                <span class="compose-model-catalogue-option-primary"><span class="compose-model-catalogue-option-name">\${modelPanelName(modelOption)}</span></span>
                                                \${modelPanelName(modelOption) !== modelLabel && html\`<span class="compose-model-catalogue-option-key">\${modelLabel}</span>\`}
                                                <span class="compose-model-catalogue-option-badges">\${modelPanelContext(modelOption?.contextWindow) && html\`<span class="compose-model-catalogue-badge">\${modelPanelContext(modelOption.contextWindow)}</span>\`}\${modelOption?.reasoning && html\`<span class="compose-model-catalogue-badge">reasoning</span>\`}</span>
                                                \${blocked && html\`<span class="compose-model-catalogue-option-note">Context window is smaller than the latest measured request.</span>\`}
                                            </span>`);
 const actions=`                            <div class="compose-model-popup-actions">
                                <button
                                    type="button"
                                    class="compose-model-popup-btn"
                                    onClick=\${() => { void handleCycleModel(); }}
                                    disabled=\${switchingModel}
                                >
                                    Next model
                                </button>
                            </div>`;
 replace(actions,`                            <div class="compose-model-catalogue-footer">
                                <div class="compose-model-catalogue-footer-start">
                                    \${supportsThinking && thinkingLevel && html\`<label class="compose-model-catalogue-thinking" title="Thinking level is read-only in Gi"><span>Thinking</span><select aria-label="Thinking level (read-only)" disabled><option value=\${thinkingLevel}>\${thinkingLevel}</option></select></label>\`}
                                </div>
                                <button type="button" class="compose-model-popup-btn primary" disabled=\${switchingModel} onClick=\${() => { setShowModelPopup(false); openModelSettings(modelHintRef.current); }}>Open Models settings</button>
                            </div>`);
 replace("if (showModelPopup) modelPopupRef.current?.focus();","if (showModelPopup) modelPopupRef.current?.querySelector('input[type=search]')?.focus({ preventScroll: true });");
 replace("const active = popup?.querySelector?.('.compose-model-popup-item.active');", "const active = popup?.querySelector?.('.compose-model-catalogue-option.active');");
 return `import { modelPanelRows, modelPanelName, modelPanelContext, openModelSettings } from '../gi-model-panel.js';\n${result}`;
}
