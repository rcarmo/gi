import { test, expect } from 'bun:test';
import { notifyModelSettlement, subscribeModelSettlement } from '../../../web/src/gi-model-invalidation.ts';

test('model settlement only notifies mounted listeners for its captured session', () => {
    const seen: string[] = [];
    const offA = subscribeModelSettlement('gi:a', () => seen.push('a'));
    const offB = subscribeModelSettlement('gi:b', () => seen.push('b'));
    notifyModelSettlement('gi:a'); expect(seen).toEqual(['a']);
    offA(); offA(); notifyModelSettlement('gi:a'); expect(seen).toEqual(['a']);
    notifyModelSettlement('gi:b'); expect(seen).toEqual(['a', 'b']); offB();
    const offNew = subscribeModelSettlement('gi:a', () => seen.push('new'));
    offA(); // A stale cleanup must not remove the newer session listener set.
    notifyModelSettlement('gi:a'); expect(seen).toEqual(['a', 'b', 'new']); offNew();
});

test('consumer exceptions and listener removal cannot change settlement or skip peers', () => {
    const seen: string[] = [];
    const offBad = subscribeModelSettlement('gi:a', () => { throw new Error('consumer failure'); });
    let offSelf: () => void;
    offSelf = subscribeModelSettlement('gi:a', () => { offSelf(); seen.push('self'); });
    const offGood = subscribeModelSettlement('gi:a', () => seen.push('good'));
    expect(() => notifyModelSettlement('gi:a')).not.toThrow();
    notifyModelSettlement('gi:a'); expect(seen).toEqual(['self', 'good', 'good']);
    offBad(); offGood();
});
