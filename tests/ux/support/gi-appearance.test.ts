import { test, expect } from 'bun:test';
import { APPEARANCE_KEY, defaultAppearance, readAppearance, saveAppearance, validateAppearance } from '../../../web/src/gi-appearance-state';
const presets = ['default', 'monokai'];

test('appearance validates its version and preset and normalises only hex tint', () => {
  expect(validateAppearance({ version: 1, theme: 'default', tint: ' #AbC ' }, presets).tint).toBe('#aabbcc');
  expect(validateAppearance({ version: 1, theme: 'monokai', tint: '#123456' }, presets).tint).toBe('');
  for (const patch of [{ theme: 'fake' }, { version: 2 }, { tint: 'red' }, { tint: '#12' }, { tint: null }]) {
    expect(() => validateAppearance({ ...defaultAppearance, ...patch }, presets)).toThrow();
  }
});

test('one browser record persists before callers render; denied writes throw', () => {
  const writes: any[] = [];
  expect(saveAppearance({ setItem: (...args) => { writes.push(args); } }, defaultAppearance, presets)).toEqual(defaultAppearance);
  expect(writes).toEqual([[APPEARANCE_KEY, JSON.stringify(defaultAppearance)]]);
  expect(() => saveAppearance({ setItem: () => { throw new Error('quota denied'); } }, defaultAppearance, presets)).toThrow('quota denied');
});

test('unavailable or malformed appearance is absent, not a startup failure', () => {
  for (const raw of [null, 'bad', 'null', '{}', JSON.stringify({ version: 2, theme: 'default', tint: '' })]) {
    expect(readAppearance({ getItem: () => raw }, presets)).toBeNull();
  }
  expect(readAppearance({ getItem: () => { throw new Error('denied'); } }, presets)).toBeNull();
  expect(readAppearance({ getItem: () => JSON.stringify(defaultAppearance) }, presets)).toEqual(defaultAppearance);
});
