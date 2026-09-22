import {test,expect} from 'bun:test';
import {contextPresentation,formatContextCount,modelContextBlocked} from '../../../web/src/gi-context-usage';

test('supplied context counts keep unknowns and compact K/M formatting',()=>{
 expect(formatContextCount(null)).toBe('?');expect(formatContextCount(undefined)).toBe('?');expect(formatContextCount(NaN)).toBe('?');expect(formatContextCount(0)).toBe('0');
 expect(formatContextCount(85000)).toBe('85K');expect(formatContextCount(1000000)).toBe('1.0M');
 const unknown=contextPresentation({tokens:null,contextWindow:128000,percent:null});
 expect(unknown.label).toBe('Context: ? / 128K tokens (?%)');expect(unknown.title).not.toContain('Compact context');
 const known=contextPresentation({tokens:85000,contextWindow:100000,percent:85,source:'provider_request'},true);
 expect(known.title).toContain('85K / 100K');expect(known.title).toContain('latest measured provider request');expect(known.title).toContain('Compact context');
});
test('supplied context colors and arc clamp retain percentage truth',()=>{
 for(const [percent,color] of [[0,'green'],[75,'green'],[75.01,'amber'],[90,'amber'],[90.01,'red'],[120,'red']] as const){
  const result=contextPresentation({percent});expect(result.color).toContain(color);expect(result.fill).toBe(Math.min(percent,100));
 }
 expect(contextPresentation({percent:120}).label).toContain('120%');
 expect(contextPresentation({percent:null}).color).toBe('var(--text-secondary)');
});
test('model fit blocks only measured overflow; unknown usage/capacity remains unknown',()=>{
 expect(modelContextBlocked({contextWindow:100},{tokens:101})).toBe(true);
 expect(modelContextBlocked({contextWindow:100},{tokens:100})).toBe(false);
 expect(modelContextBlocked({contextWindow:100},{tokens:null})).toBe(false);
 expect(modelContextBlocked({contextWindow:null},{tokens:1000})).toBe(false);
});
