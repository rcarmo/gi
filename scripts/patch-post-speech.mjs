// Frozen Piclaw 70d33bc93 speech control, applied without changing supplied Post.
function replace(source, from, to) {
    if (source.split(from).length !== 2) throw new Error('Post speech adapter anchor changed: ' + from);
    return source.replace(from, to);
}
export function patchPostSpeech(source) {
    if (source.includes('gi-post-speech.js')) throw new Error('Post speech adapter already applied');
    source = replace(source, "    const [copyState, setCopyState] = useState('idle');", `    const [copyState, setCopyState] = useState('idle');
    const speechOwner = useMemo(() => ({}), [post.id]);
    const [speechState, setSpeechState] = useState(speechPlayback.state);
    const speakableText = useMemo(() => buildSpeakablePostText(post), [post]);
    const speechSupported = speechPlayback.supported();
    const isSpeakingThisPost = speechState.speaking && speechState.owner === speechOwner;
    useEffect(() => speechPlayback.subscribe(setSpeechState), []);
    useEffect(() => () => speechPlayback.stop(speechOwner), [speechOwner]);
    useEffect(() => { speechPlayback.stop(speechOwner); }, [speechOwner, speakableText, post.data?.content]);
    const handleSpeakClick = (event) => {
        event.preventDefault(); event.stopPropagation();
        if (isSpeakingThisPost) speechPlayback.stop(speechOwner);
        else speechPlayback.speak(speechOwner, speakableText);
    };`);
    source = replace(source, '                <div class="post-actions">', `                <div class="post-actions">
                    \${isAgent && speechSupported && speakableText && html\`
                        <button type="button" class=\${'post-action-btn post-speak-btn' + (isSpeakingThisPost ? ' is-active' : '')}
                            onClick=\${handleSpeakClick}
                            title=\${isSpeakingThisPost ? 'Stop reading aloud' : 'Read aloud'}
                            aria-label=\${isSpeakingThisPost ? 'Stop reading aloud' : 'Read aloud'}
                            aria-pressed=\${isSpeakingThisPost}>
                            \${isSpeakingThisPost ? html\`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="7" y="7" width="10" height="10" rx="1"/></svg>\` : html\`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 5 6 9H3v6h3l5 4z"/><path d="M15.5 8.5a5 5 0 0 1 0 7"/><path d="M19 5a10 10 0 0 1 0 14"/></svg>\`}
                        </button>
                    \`}`);
    return `import { buildSpeakablePostText, speechPlayback } from '../gi-post-speech.js';\n${source}`;
}
