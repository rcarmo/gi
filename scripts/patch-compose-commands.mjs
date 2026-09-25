// Bind the immutable composer to Gi's advertised commands and key ownership.
function replace(source,from,to){
 if(source.split(from).length!==2)throw new Error(`Compose command adapter anchor changed: ${from}`);
 return source.replace(from,to);
}
export function patchComposeCommands(source){
 source=replace(source,'    const dynamicCommandsRef = useRef(null);','    const dynamicCommandsRef = useRef([]);\n    const commandSearchModeRef = useRef(searchMode); commandSearchModeRef.current = searchMode;\n    const [commandCatalogueError, setCommandCatalogueError] = useState(false);');
 source=replace(source,`    // Fetch dynamic commands from the server for autocomplete
    useEffect(() => {
        let cancelled = false;
        const chatJid = currentChatJid || 'web:default';
        fetch(\`/agent/commands?chat_jid=\${encodeURIComponent(chatJid)}\`)
            .then(r => r.ok ? r.json() : null)
            .then(data => {
                if (cancelled || !data?.commands) return;
                dynamicCommandsRef.current = data.commands.map(c => ({
                    name: c.name,
                    description: c.description || '',
                }));
            })
            .catch((e) => {
                // keep hardcoded fallback — dynamic commands are optional
                console.debug("[compose] failed to fetch dynamic commands", e);
            });
        return () => { cancelled = true; };
    }, [currentChatJid]);`,`    // Gi never advertises the bundled Piclaw fallback catalogue.
    useEffect(() => {
        let cancelled = false;
        dynamicCommandsRef.current = [];
        setShowSlash(false); setSlashMatches([]);
        setCommandCatalogueError(false);
        getAgentCommands(currentChatJid).then(data => {
            if (cancelled) return;
            dynamicCommandsRef.current = normaliseComposeCommands(data);
            if (!settingsOwnsKeyboard() && !commandSearchModeRef.current) updateSlashAutocomplete(textareaRef.current?.value || '');
        }).catch(() => {
            if (cancelled) return;
            dynamicCommandsRef.current = [];
            setShowSlash(false); setSlashMatches([]);
            setCommandCatalogueError(true);
        });
        return () => { cancelled = true; };
    }, [currentChatJid]);`);
 source=replace(source,'const commandList = dynamicCommandsRef.current || SLASH_COMMANDS;','const commandList = dynamicCommandsRef.current;');
 source=replace(source,'            ${submitError && html`', '            ${commandCatalogueError && html`<div class="compose-submit-notice" role="status">Command suggestions unavailable. Reopen this chat or reload to retry.</div>`}\n            ${submitError && html`');
 source=replace(source,'    const handleKeyDown = (e) => {\n        if (e.isComposing) return;','    const handleKeyDown = (e) => {\n        if (declineComposeKey(e)) return;');
 source=replace(source,'        if (settingsOwnsKeyboard()) return false;\n        if (searchMode','        if (settingsOwnsKeyboard() || declineComposeKey(e)) return false;\n        if (searchMode');
 return "import { getAgentCommands } from '../api.js';\nimport { normaliseComposeCommands, declineComposeKey } from '../gi-compose-commands.js';\n"+source;
}
