function setAdaptiveCardAttributeBestEffort(element: HTMLElement, name: string, value: string): boolean {
  try {
    element.setAttribute(name, value);
    return true;
  } catch (_error) {
    return false;
  }
}

function setAdaptiveCardBooleanPropertyBestEffort(
  element: HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement | HTMLButtonElement,
  property: "disabled" | "readOnly",
): boolean {
  try {
    (element as any)[property] = true;
    return true;
  } catch (_error) {
    return false;
  }
}

export function lockAdaptiveCardInputs(root: HTMLElement): void {
  root.classList.add("adaptive-card-readonly");
  for (const input of Array.from(root.querySelectorAll("input, textarea, select, button"))) {
    const element = input as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement | HTMLButtonElement;
    setAdaptiveCardAttributeBestEffort(element, "aria-disabled", "true");
    setAdaptiveCardAttributeBestEffort(element, "tabindex", "-1");
    if ("disabled" in element) {
      setAdaptiveCardBooleanPropertyBestEffort(element, "disabled");
    }
    if ("readOnly" in element) {
      setAdaptiveCardBooleanPropertyBestEffort(element, "readOnly");
    }
  }
}
