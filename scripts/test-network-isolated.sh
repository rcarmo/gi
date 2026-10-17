#!/usr/bin/env bash
# Opt-in Linux browser-test isolation from unrelated host interface churn.
# Namespace setup alone is privileged; the requested command runs as the caller.
set -euo pipefail
if [[ $# -eq 0 ]]; then echo 'Usage: test-network-isolated.sh command [args...]' >&2; exit 2; fi
if [[ $(uname -s) != Linux ]]; then echo 'Network namespace isolation requires Linux' >&2; exit 2; fi
uid=$(id -u); gid=$(id -g)
if [[ $uid == 0 ]]; then echo 'Invoke as the unprivileged test owner, not root' >&2; exit 2; fi
# Preserve only explicit non-secret test/tool settings. No credential environment
# or network access is needed by offline fixture runtimes.
vars=("PATH=$PATH" "HOME=$HOME")
for name in PROJECT_TMP_ROOT PROJECT_TMP_BASE TMPDIR TMP TEMP GI_TEST_RUN_ROOT GI_TEST_RUN_ID GI_PROFILED_SPECS FIXTURES_PROFILE GI_FIXTURE_BIN GI_TEST_URL GOTOOLCHAIN GOCACHE GOMODCACHE GOPATH BUN_RUNTIME_TRANSPILER_CACHE_PATH BUN_INSTALL_CACHE_DIR PLAYWRIGHT_BROWSERS_PATH XDG_CACHE_HOME PROFILING; do
  if [[ -v $name ]]; then vars+=("$name=${!name}"); fi
done
exec sudo -n /usr/bin/unshare --net -- /bin/sh -c '
  /usr/sbin/ip link set lo up || exit
  uid=$1; gid=$2; shift 2
  exec /usr/bin/setpriv --reuid="$uid" --regid="$gid" --init-groups /usr/bin/env -i "$@"
' sh "$uid" "$gid" "${vars[@]}" "$@"
