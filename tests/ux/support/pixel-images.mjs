import {PNG} from 'pngjs';

// Raw RGBA governs acceptance. Different dimensions are a failure, not a reason
// to scale/pad/crop either product to match the other.
export function comparePixels(a,b){
 if(a.width!==b.width||a.height!==b.height)return {dimensionsMatch:false,exactMatch:false,left:[a.width,a.height],right:[b.width,b.height]};
 const diff=new PNG({width:a.width,height:a.height}),overlay=new PNG({width:a.width,height:a.height});let changedPixels=0;
 for(let i=0;i<a.data.length;i+=4){const changed=[0,1,2,3].some(c=>a.data[i+c]!==b.data[i+c]);if(changed)changedPixels++;for(let c=0;c<3;c++){diff.data[i+c]=changed?(c===0||c===2?255:0):Math.round(a.data[i+c]*0.2);overlay.data[i+c]=Math.round((a.data[i+c]+b.data[i+c])/2);}diff.data[i+3]=255;overlay.data[i+3]=255;}
 return {dimensionsMatch:true,exactMatch:changedPixels===0,changedPixels,totalPixels:a.width*a.height,diff,overlay};
}

// Diagnostic regions retain both hosts' viewport coordinates. Never align or
// resize the elements. Full-frame comparison remains mandatory for acceptance.
export function unionRegion(rects,width,height){
 if(!rects.length||rects.some(r=>!r||r.width<=0||r.height<=0))throw Error('Missing pixel region');
 const x=Math.max(0,Math.floor(Math.min(...rects.map(r=>r.x))));
 const y=Math.max(0,Math.floor(Math.min(...rects.map(r=>r.y))));
 const right=Math.min(width,Math.ceil(Math.max(...rects.map(r=>r.x+r.width))));
 const bottom=Math.min(height,Math.ceil(Math.max(...rects.map(r=>r.y+r.height))));
 if(right<=x||bottom<=y)throw Error('Pixel region outside viewport');
 return {x,y,width:right-x,height:bottom-y};
}
export function extractRegion(image,region){
 const out=new PNG({width:region.width,height:region.height});
 PNG.bitblt(image,out,region.x,region.y,region.width,region.height,0,0);
 return out;
}
