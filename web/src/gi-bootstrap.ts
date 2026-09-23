// Stable, cache-busted shell entry; all executable app modules use content hashes.
// This avoids split panes importing a second unversioned instance of the app.
void import('./app.js').catch(() => {
    const host = document.getElementById('app');
    if (host) host.textContent = 'Unable to load Gi. Reload the page to retry.';
});
