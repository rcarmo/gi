export const OOBE_PROVIDER_MISSING_DISMISSED_KEY = 'piclaw:oobe:provider-missing:dismissed';
export const OOBE_PROVIDER_READY_COMPLETED_KEY = 'piclaw:oobe:provider-ready:completed';

export type OobePanelKind = 'hidden' | 'provider-missing' | 'provider-ready';

export interface OobePanelState {
  kind: OobePanelKind;
  hasAvailableModels: boolean;
  availableModelCount: number;
  hasConfiguredModelHint: boolean;
}

export function countAvailableModels(payload: Record<string, unknown> | null | undefined): number {
  if (!payload || typeof payload !== 'object') return 0;

  const modelOptions = Array.isArray((payload as any).model_options)
    ? (payload as any).model_options.filter(Boolean)
    : [];
  if (modelOptions.length > 0) return modelOptions.length;

  const models = Array.isArray((payload as any).models)
    ? (payload as any).models.filter((value: unknown) => typeof value === 'string' && value.trim())
    : [];
  return models.length;
}

export function hasConfiguredModelHint(payload: Record<string, unknown> | null | undefined): boolean {
  if (!payload || typeof payload !== 'object') return false;
  const current = (payload as any).current;
  return typeof current === 'string' && current.trim().length > 0;
}

export function isProviderReadyCompletedForInstance(payload: Record<string, unknown> | null | undefined): boolean {
  if (!payload || typeof payload !== 'object') return false;
  const oobe = (payload as any).oobe;
  return Boolean(oobe && typeof oobe === 'object' && oobe.provider_ready_completed_instance === true);
}

export function resolveOobePanelState(options: {
  panePopoutMode?: boolean;
  modelsLoaded: boolean;
  modelPayload: Record<string, unknown> | null | undefined;
  providerMissingDismissed?: boolean;
  providerReadyCompleted?: boolean;
}): OobePanelState {
  const {
    panePopoutMode = false,
    modelsLoaded,
    modelPayload,
    providerMissingDismissed = false,
    providerReadyCompleted = false,
  } = options;

  const availableModelCount = countAvailableModels(modelPayload);
  const hasAvailableModels = availableModelCount > 0;
  const configuredModelHint = hasConfiguredModelHint(modelPayload);
  const providerReadyCompletedResolved = providerReadyCompleted || isProviderReadyCompletedForInstance(modelPayload);

  if (panePopoutMode || !modelsLoaded) {
    return {
      kind: 'hidden',
      hasAvailableModels,
      availableModelCount,
      hasConfiguredModelHint: configuredModelHint,
    };
  }

  if (!hasAvailableModels) {
    return {
      kind: configuredModelHint ? 'hidden' : (providerMissingDismissed ? 'hidden' : 'provider-missing'),
      hasAvailableModels,
      availableModelCount,
      hasConfiguredModelHint: configuredModelHint,
    };
  }

  return {
    kind: (providerReadyCompletedResolved || configuredModelHint) ? 'hidden' : 'provider-ready',
    hasAvailableModels,
    availableModelCount,
    hasConfiguredModelHint: configuredModelHint,
  };
}
