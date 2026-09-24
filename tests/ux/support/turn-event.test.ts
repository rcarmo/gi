import {test,expect} from 'bun:test';
import {staleTerminalEvent} from '../../../web/src/gi-turn-event';
test('terminal frames only clear their own occurrence; legacy terminal frames refresh only',()=>{
 for(const type of ['agent_status','agent_response']){
  for(const turn_id of ['old',undefined])expect(staleTerminalEvent(type,{status:'idle',turn_id},'new')).toBe(true);
  expect(staleTerminalEvent(type,{status:'idle',turn_id:'new'},'new')).toBe(false);
  expect(staleTerminalEvent(type,{status:'idle',turn_id:'old'},null)).toBe(false);
 }
 for(const status of ['running','cancelling'])expect(staleTerminalEvent('agent_status',{status,turn_id:'new'},'old')).toBe(false);
 for(const type of ['agent_draft_delta','new_post','queue_changed'])expect(staleTerminalEvent(type,{turn_id:'old'},'new')).toBe(false);
});
