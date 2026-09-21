import { test, expect } from 'bun:test';
import { createSelectionScope } from '../../../web/src/gi-session-state.ts';

test('late requests cannot update another selection or an A-B-A revisit', () => {
    const scope = createSelectionScope();
    scope.select('main');
    const oldMain = scope.capture();
    expect(scope.isCurrent(oldMain)).toBe(true);
    scope.select('research');
    const research = scope.capture();
    expect(scope.isCurrent(oldMain)).toBe(false);
    scope.select('main');
    expect(scope.isCurrent(oldMain)).toBe(false);
    expect(scope.isCurrent(research)).toBe(false);
    const current = scope.capture();
    scope.select('main');
    expect(scope.isCurrent(current)).toBe(true);
});
