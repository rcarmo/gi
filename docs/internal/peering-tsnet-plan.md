# Gi tsnet remote-access plan

Updated: 2026-09-25. Status: scaffold; no tailnet HTTP listener or runtime startup wiring.

The requested split is tsnet for operator web/API access and Iroh for inter-instance
chat. Both must run in Gi's pure-Go binary without CGO, native Rust libraries or
sidecars. The [feature matrix](../feature-parity.md) tracks these separately.
The existing `peering` config/tool name belongs to the tsnet scaffold; it does not
provide Piclaw `remote_peer` pairing or messaging.

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

## Remote-access work

* Wire explicit startup, cancellation and joined shutdown into the application listener lifecycle. A configured manager alone must not report reachable service.
* Serve the existing web UI, API and SSE over a tailnet-only HTTPS listener, with persistent node identity and env/keychain enrolment references. The scaffold currently sets `Ephemeral: true`; persistent operation needs an explicit change and tests.
* Keep browser authentication and same-origin checks. Tailnet membership is not Gi owner authentication. Never expose an unenrolled instance automatically.
* Select a stable HTTPS hostname and RP ID before production passkey enrolment. Passkeys registered for localhost do not automatically work on a tailnet hostname.
* Test listener startup failure, reconnect, shutdown, secret redaction and remote cookie/TLS behaviour with isolated fixtures, then verify authorised tailnet access explicitly.

## Separate Iroh work

Gi has no Iroh implementation. `tmc/go-iroh` is a candidate pure-Go stack whose
wire compatibility must be tested against Piclaw's installed remote-peer add-on.
The target application protocol includes explicit identity-confirmed pairing,
receiver-owned permissions, signed bounded messages/files, one-hop addressing,
epoch-based revocation and durable retry/deduplication. Whether the first release
must interoperate directly with existing Piclaw instances is an open scope decision.
There is no transport substitution that can establish Iroh compatibility by itself.

## Safety defaults

- Peering is disabled by default.
- No listeners are exposed by adding the package; Funnel/public exposure is not a default.
- Missing env/keychain auth produces explicit status/errors rather than falling back to unauthenticated external access.
