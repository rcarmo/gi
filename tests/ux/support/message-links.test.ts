import {test,expect} from 'bun:test';
import {remoteLinkUrl,projectResourceLinks,projectLinkPreviews} from '../../../web/src/gi-message-links.ts';
import {projectMessageMedia} from '../../../web/src/gi-message-media.ts';
test('remote metadata only accepts explicit credential-free HTTP(S) URLs',()=>{
 for(const bad of ['javascript:alert(1)','data:text/html,test','file:///tmp/file','//example.com','/local','https://name:secret@example.com','https://example.com\\evil','https://example.com/\n','https://example.com/'+ 'x'.repeat(2048),null,{}])expect(remoteLinkUrl(bad)).toBeNull();
 for(const url of ['https://example.invalid/a?q=β#part','http://example.invalid'])expect(remoteLinkUrl(url)).toBe(new URL(url).href);
});
test('resource projection bounds metadata and leaves unrelated blocks and media pairing intact',()=>{
 const text={type:'text',text:'retained'},payload={content_blocks:[text,{type:'resource_link',uri:'javascript:bad'},...Array.from({length:12},(_,i)=>({type:'resource_link',uri:`https://example.invalid/${i}`,title:'x'.repeat(400),description:'d'.repeat(2000),size:10,extra:'discard'}))],media:[{media_id:7,filename:'kept.png',content_type:'image/png'}]};
 const before=JSON.stringify(payload),projected=projectMessageMedia(payload,'main');
 expect(projected.content_blocks.filter(b=>b.type==='resource_link')).toHaveLength(8);expect(projected.content_blocks[0]).toBe(text);expect(projected.content_blocks[1].title).toHaveLength(200);expect(projected.content_blocks[1].description).toHaveLength(1000);expect(projected.content_blocks[1].extra).toBeUndefined();expect(projected.media_ids).toEqual([7]);expect(projected.content_blocks.at(-1).type).toBe('image');expect(JSON.stringify(payload)).toBe(before);
});
test('text previews never project remote images, fabricated sites, scripts or unbounded lists',()=>{
 const entry={url:'https://preview.example.invalid/path',title:'safe',image:'https://tracker.invalid/x',site_name:'fake'};
 const payload={link_previews:[entry,entry,{url:'javascript:bad'},...Array.from({length:20},(_,i)=>({url:`https://preview.example.invalid/${i}`}))]};
 const previews=projectLinkPreviews(payload);expect(previews).toHaveLength(8);expect(previews[0]).toEqual({url:entry.url,title:'safe',description:'',site_name:'preview.example.invalid'});expect(projectLinkPreviews(null)).toBeNull();expect(projectLinkPreviews({link_previews:[null]})).toBeNull();
});
