import {test,expect} from 'bun:test';
import {projectMessageMedia} from '../../../web/src/gi-message-media.ts';

test('native message media keeps IDs, names and types paired without changing the payload',()=>{
 const payload={content_blocks:[{type:'text',text:'caption'},{type:'image',data:'stale'},{type:'generated_widget',widget_id:'w'}],media:[
  {media_id:12,session_id:'a',filename:'one.png',content_type:'image/png'},
  {media_id:13,session_id:'a',filename:'notes.txt',content_type:'text/plain'},
  {media_id:14,session_id:'a',filename:'vector.svg',content_type:'image/svg+xml'},
 ]};
 const before=JSON.stringify(payload),result=projectMessageMedia(payload,'a');
 expect(result.media_ids).toEqual([12,13,14]);
 expect(result.content_blocks).toEqual([{type:'text',text:'caption'},{type:'generated_widget',widget_id:'w'},
  {type:'image',name:'one.png',mime_type:'image/png'},{type:'file',name:'notes.txt',mime_type:'text/plain'},
  {type:'image',name:'vector.svg',mime_type:'image/svg+xml'}]);
 expect(JSON.stringify(payload)).toBe(before);
});
test('native media projection excludes foreign and invalid references',()=>{
 const result=projectMessageMedia({media:[null,{media_id:-1},{media_id:'4'},{media_id:Infinity},{media_id:4,session_id:'other'},{media_id:5,session_id:'a'}]},'a');
 expect(result.media_ids).toEqual([5]);expect(result.content_blocks).toEqual([{type:'file',name:'attachment-5',mime_type:'application/octet-stream'}]);
 expect(projectMessageMedia(null,'a')).toEqual({media_ids:[],content_blocks:null});
 expect(projectMessageMedia({content_blocks:[{type:'text',text:'hi'}]},'a').content_blocks).toEqual([{type:'text',text:'hi'}]);
});
