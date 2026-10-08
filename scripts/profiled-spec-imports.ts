import { join } from 'node:path';

// Relocate module references only. Leave shared test assertions unchanged.
export function relocateSpecImports(source: string, suite: string, playwright: string, fixtures: string): string {
  const location = (path: string) => path === '@playwright/test' ? playwright
    : path === '../fixtures' ? fixtures
    : path.startsWith('../') ? join(suite, path.slice(3))
    : path.startsWith('./') ? join(suite, 'specs', path.slice(2)) : null;
  return source.replace(/\bfrom\s+(['"])([^'"]+)\1/g, (original, _quote, path) => {
    const target = location(path);
    return target ? `from ${JSON.stringify(target)}` : original;
  }).replace(/\bimport\s*\(\s*(['"])([^'"]+)\1\s*\)/g, (original, _quote, path) => {
    const target = location(path);
    return target ? `import(${JSON.stringify(target)})` : original;
  });
}
