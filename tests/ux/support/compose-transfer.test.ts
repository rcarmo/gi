import {test,expect} from 'bun:test';
import {createComposeTransfers} from '../../../web/src/gi-compose-transfer';

test('transport phases are independent per session and concurrent completion cannot clear peers',()=>{
 const s=createComposeTransfers();let changes=0;const unsubscribe=s.subscribe(()=>changes++);
 const a=s.begin('a','upload'),b=s.begin('b','upload'),peer=s.begin('a','upload');
 a.progress(5,10,true);peer.progress(4,0,false);
 expect(s.snapshot('a')).toEqual({uploads:2,sending:0,loaded:9,total:10,computable:false});
 peer.end();a.progress(10,10,true);expect(s.snapshot('a')).toEqual({uploads:1,sending:0,loaded:10,total:10,computable:true});
 a.end();const sending=s.begin('a','send');
 expect(s.snapshot('a').uploads).toBe(0);expect(s.snapshot('a').sending).toBe(1);expect(s.snapshot('b').uploads).toBe(1);
 a.progress(99,100,true);a.end();expect(s.snapshot('a').sending).toBe(1);
 const second=s.begin('a','send');sending.end();expect(s.snapshot('a').sending).toBe(1);second.end();b.end();
 expect(s.snapshot('a').sending).toBe(0);expect(s.snapshot('b').uploads).toBe(0);
 unsubscribe();const before=changes;s.begin('c','send').end();expect(changes).toBe(before);
});

test('unknown byte totals remain indeterminate and fresh reload has no work',()=>{
 const s=createComposeTransfers(),op=s.begin('a','upload');
 op.progress(NaN,Infinity,true);expect(s.snapshot('a')).toMatchObject({loaded:0,total:0,computable:false});
 op.progress(20,10,true);expect(s.snapshot('a')).toMatchObject({loaded:10,total:10,computable:true});
 expect(createComposeTransfers().snapshot('a')).toMatchObject({uploads:0,sending:0});
 op.end();expect(s.snapshot('a')).toMatchObject({uploads:0,sending:0});
});
