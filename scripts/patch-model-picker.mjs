// Add the missing shared picker contract without modifying supplied sources.
function replace(source, from, to) {
    if (source.split(from).length !== 2) throw new Error(`Model picker adapter anchor changed: ${from}`);
    return source.replace(from, to);
}
export function patchModelPicker(source) {
    source = replace(source, '    const [loadingModels, setLoadingModels] = useState(false);', `    const [loadingModels, setLoadingModels] = useState(false);
    const [modelQuery, setModelQuery] = useState('');
    const visibleModels = useMemo(() => filterModelOptions(modelOptions, modelQuery, getModelPickerOptionSearchLabel), [modelOptions, modelQuery]);
    const modelEntries = useMemo(() => visibleModels.map(option => ({
        label: getModelPickerOptionSearchLabel(option),
        disabled: loadingModels || switchingModel || modelContextBlocked(option, contextUsage),
    })), [visibleModels, loadingModels, switchingModel, contextUsage]);`);
    source = replace(source, `        setShowModelPopup((prev) => !prev);`, `        setModelQuery('');
        setShowModelPopup((prev) => !prev);`);
    const oldKeys = `        if (showModelPopup) {
            if (e.key === 'ArrowDown') {
                consume();
                resetPopupTypeahead();
                if (modelOptions.length > 0) setModelPopupIndex((idx) => (idx + 1) % modelOptions.length);
                return true;
            }
            if (e.key === 'ArrowUp') {
                consume();
                resetPopupTypeahead();
                if (modelOptions.length > 0) setModelPopupIndex((idx) => (idx - 1 + modelOptions.length) % modelOptions.length);
                return true;
            }
            if (e.key === 'Tab' || (e.key === 'Enter' && e.target?.closest?.('button'))) return false;
            if (e.key === 'Enter' && modelOptions.length > 0) {
                consume();
                resetPopupTypeahead();
                void handleSelectModel(modelOptions[Math.max(0, Math.min(modelPopupIndex, modelOptions.length - 1))]);
                return true;
            }
            if (isPopupTypeaheadKey(e) && modelOptions.length > 0) {
                consume();
                const nextBuffer = updatePopupTypeaheadBuffer(popupTypeaheadRef.current, e.key);
                popupTypeaheadRef.current = nextBuffer;
                const match = resolvePopupTypeaheadMatch(modelOptions, nextBuffer.value, modelPopupIndex, (item) => getModelPickerOptionSearchLabel(item));
                if (match >= 0) setModelPopupIndex(match);
                return true;
            }
        }`;
    source = replace(source, oldKeys, `        if (showModelPopup && modelPopupRef.current?.contains(e.target)) {
            const action = modelPickerKey(e, modelEntries, modelPopupIndex, popupTypeaheadRef.current);
            if (action) {
                consume();
                popupTypeaheadRef.current = action.buffer;
                if (action.index >= 0) {
                    setModelPopupIndex(action.index);
                    if (action.focus) modelPopupRef.current?.querySelector('[data-model-index="' + action.index + '"]')?.focus({ preventScroll: true });
                    if (action.activate) void handleSelectModel(visibleModels[action.index]);
                }
                return true;
            }
        }`);
    source = replace(source, `        modelOptions,
        modelPopupIndex,`, `        visibleModels,
        modelEntries,
        modelPopupIndex,`);
    source = replace(source, `    useEffect(() => {
        if (!showModelPopup) return;
        const activeIndex = modelOptions.findIndex((model) => model?.label === activeModel);
        setModelPopupIndex(activeIndex >= 0 ? activeIndex : 0);
    }, [showModelPopup, modelOptions, activeModel]);`, `    useLayoutEffect(() => {
        if (!showModelPopup) return;
        const activeIndex = visibleModels.findIndex((model, index) => model?.label === activeModel && !modelEntries[index].disabled);
        setModelPopupIndex(activeIndex >= 0 ? activeIndex : modelEntries.findIndex(entry => !entry.disabled));
    }, [showModelPopup, visibleModels, activeModel, modelEntries]);`);
    source = replace(source, '                            <div class="compose-model-popup-title">Select model</div>', `                            <div class="compose-model-popup-title">Select model</div>
                            <input type="search" class="compose-session-search" aria-label="Search models" placeholder="Search models"
                                value=\${modelQuery} onInput=\${event => {
                                    popupTypeaheadRef.current = { value: '', updatedAt: 0 };
                                    setModelQuery(event.currentTarget.value);
                                }} />`);
    source = replace(source, '${!loadingModels && modelOptions.length === 0 && html`', '${!loadingModels && visibleModels.length === 0 && html`');
    source = replace(source, '<div class="compose-model-popup-empty">No models available.</div>', '<div class="compose-model-popup-empty">${modelQuery ? "No models match your search." : "No models available."}</div>');
    source = replace(source, '${!loadingModels && modelOptions.map((modelOption, index) => {', '${!loadingModels && visibleModels.map((modelOption, index) => {');
    source = replace(source, '                                            key=${modelLabel}', '                                            key=${modelLabel}\n                                            data-model-index=${index}');
    return `import { filterModelOptions, modelPickerKey } from '../gi-model-picker.js';\n${source}`;
}
