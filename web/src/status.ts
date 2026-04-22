let listener: ((text: string) => void) | null = null;

export function bindStatusListener(fn: (text: string) => void) {
  listener = fn;
}

export function recordStatus(text: string) {
  if (listener) listener(text);
}
