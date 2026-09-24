import {expect,test} from 'bun:test';
import {uploadMedia} from '../../../web/src/api';
import {composeTransfers} from '../../../web/src/gi-compose-transfer';

test('native upload adapter aborts before encoding or during XHR, fences late completion and cleans listeners',async()=>{
 const original=globalThis.XMLHttpRequest;const calls:any[]=[];let throwOpen=false;
 class XHR {
  upload:any={};onabort:any;onerror:any;onload:any;status=201;responseText='{"media":{"id":42}}';
  open(){calls.push('open');if(throwOpen)throw new Error('open failed');}setRequestHeader(){}send(body:any){calls.push(body);}abort(){calls.push('abort');this.onabort?.();}
  constructor(){calls.push(this);}
 }
 globalThis.XMLHttpRequest=XHR as any;
 try{
  const early=new AbortController();early.abort();await expect(uploadMedia(new File(['x'],'x'),'gi:A',{signal:early.signal})).rejects.toMatchObject({name:'AbortError'});expect(calls).toHaveLength(0);
  const controller=new AbortController();let listeners=0;const add=controller.signal.addEventListener.bind(controller.signal),remove=controller.signal.removeEventListener.bind(controller.signal);
  controller.signal.addEventListener=((...args:any[])=>{listeners++;add(...args as [any,any]);}) as any;
  controller.signal.removeEventListener=((...args:any[])=>{listeners--;remove(...args as [any,any]);}) as any;
  const result=uploadMedia(new File(['exact bytes'],'x'),'gi:A',{signal:controller.signal});
  while(calls.length<3)await new Promise(r=>setTimeout(r,0));const xhr=calls[0];expect(composeTransfers.snapshot('A').uploads).toBe(1);controller.abort();
  await expect(result).rejects.toMatchObject({name:'AbortError'});expect(calls.filter(x=>x==='abort')).toHaveLength(1);expect(listeners).toBe(0);expect(composeTransfers.snapshot('A').uploads).toBe(0);expect(xhr.upload.onprogress).toBeNull();xhr.onload();xhr.onerror();expect(composeTransfers.snapshot('A').uploads).toBe(0);
  calls.length=0;const successController=new AbortController(),success=uploadMedia(new File(['y'],'y'),'gi:A',{signal:successController.signal});while(calls.length<3)await new Promise(r=>setTimeout(r,0));calls[0].onload();expect(await success).toEqual({id:42});successController.abort();expect(calls.includes('abort')).toBe(false);
  calls.length=0;throwOpen=true;const broken=new AbortController();await expect(uploadMedia(new File(['z'],'z'),'gi:A',{signal:broken.signal})).rejects.toThrow('open failed');broken.abort();expect(calls.includes('abort')).toBe(false);expect(composeTransfers.snapshot('A').uploads).toBe(0);
 }finally{globalThis.XMLHttpRequest=original;}
});
