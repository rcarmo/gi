import {test,expect} from 'bun:test';
import {buildTimelineQuickActionItems,normalizeTimelineQuickActionsSettingsData} from '../../../web/src/ui/timeline-quick-actions';

test('pinned Quick Actions groups/dedupe/filter obey native capability selections',()=>{
 const items=buildTimelineQuickActionItems({agents:[{chat_jid:'gi:a',agent_name:'Zulu'},{chat_jid:'gi:a',agent_name:'duplicate'},{chat_jid:'gi:b',agent_name:'archived',archived_at:'date'}],workspaceCommands:[{id:'toggle-workspace',label:'Show workspace',description:'Show workspace'},{id:'open-settings',label:'Settings',description:'not native'}],slashCommands:[{name:'/model',description:'model'},{name:'/model',description:'duplicate'},{name:'/skill:test',description:'unavailable'}],settings:{workspaceCommands:['toggle-workspace'],slashCommands:['/model']}});
 expect(items.map(i=>[i.kind,i.title])).toEqual([['agent','@Zulu'],['workspace','Show workspace'],['slash','/model']]);
 expect(buildTimelineQuickActionItems({agents:[{chat_jid:'gi:a',agent_name:'Zulu'}],query:'zUl'}).map(i=>i.title)).toEqual(['@Zulu']);
 expect(normalizeTimelineQuickActionsSettingsData({workspaceCommands:[],slashCommands:[]})).toEqual({workspaceCommands:[],slashCommands:[]});
});
