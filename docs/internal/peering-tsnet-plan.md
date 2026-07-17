# Gi tsnet peering plan

Goal: provide a local-first peer discovery and private connectivity backend for future gi agents without exposing HTTP routes publicly by default.

## Implemented foundation

- `internal/peering` embeds `tailscale.com/tsnet` behind a disabled-by-default manager.
- Runtime config accepts a `peering` block from `.pi/settings.json`:

```json
{
  "peering": {
    "enabled": false,
    "hostname": "gi",
    "state_dir": ".gi/tsnet",
    "auth_key_env": "TS_AUTHKEY",
    "auth_key_keychain": "tailscale/authkey"
  }
}
```

- The `peering` built-in tool reports backend/status/configuration visibility.
- `auth_key_keychain` resolves through `internal/secrets.Resolver`; the default adapter uses Piclaw's injected environment-name convention (`tailscale/authkey` → `TAILSCALE_AUTHKEY`) without writing secrets to files.

## Next implementation steps

1. Add safe start/stop lifecycle hooks for `peering.enabled=true`.
2. Expose peer discovery (`status`, `peers`, `ping`) through the `peering` tool.
4. Route selected connectivity endpoints over tsnet listeners instead of public bind addresses.
5. Add Gherkin or integration tests with a fake/control tsnet setup if we need end-to-end peer assertions.

## Safety defaults

- Peering is disabled by default.
- No listeners are exposed by adding the package.
- Missing env/keychain auth produces explicit status/errors rather than falling back to unauthenticated external access.
