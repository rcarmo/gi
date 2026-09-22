import {test,expect} from 'bun:test';
import {recoverQueueDraft} from '../../../web/src/gi-queue-return';

test('queued media is loaded from its origin and deduplicated against serialized refs',async()=>{
 const paths:string[]=[];
 const draft=await recoverQueueDraft({id:'turn',chat_jid:'gi:A',metadata:{media:[{media_id:7,session_id:'A',filename:'a.txt',content_type:'text/plain'}]}},{text:'queued',fileRefs:['dir/','a'],messageRefs:['msg'],attachmentRefs:[{id:'7',label:'a.txt'}]},(async path=>{paths.push(String(path));return new Response('bytes',{headers:{'Content-Type':'text/plain'}});}) as any);
 expect(paths).toEqual(['/api/sessions/A/media/7']);expect(draft.fileRefs).toEqual(['dir/','a']);expect(draft.media).toHaveLength(1);expect(await draft.media[0].text()).toBe('bytes');
});
test('foreign media and failed download reject recovery before any deletion can run',async()=>{
 await expect(recoverQueueDraft({id:'t',chat_jid:'gi:A',metadata:{media:[{media_id:1,session_id:'B'}]}},{},(async()=>{throw Error('must not fetch');}) as any)).rejects.toThrow('another session');
 await expect(recoverQueueDraft({id:'t',chat_jid:'gi:A'},{attachmentRefs:[{id:'1'}]},async()=>new Response('',{status:404}) as any)).rejects.toThrow('HTTP 404');
});
