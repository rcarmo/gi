export async function runPasskeyAttemptBestEffort(attempt: () => Promise<boolean>): Promise<boolean> {
  try {
    return await attempt();
  } catch (_error) {
    return false;
  }
}

export async function readJsonBodyBestEffort<T>(
  response: { json: () => Promise<T> },
  fallback: T,
): Promise<T> {
  try {
    return await response.json();
  } catch (_error) {
    return fallback;
  }
}

export async function probePasskeyCapabilityBestEffort(
  probe: () => Promise<boolean>,
): Promise<boolean> {
  try {
    return await probe();
  } catch (_error) {
    return false;
  }
}
