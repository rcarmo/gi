export function composeHeightBounds(width:number,height:number){
 const min=width>=1024?70:50;
 return {min,autoMax:Math.max(min,Math.min(Math.floor(height*.4),300)),manualMax:Math.max(min,Math.min(Math.floor(height*.5),520))};
}
export function clampComposeHeight(value:number,width:number,height:number){
 const {min,manualMax}=composeHeightBounds(width,height);
 return Math.max(min,Math.min(manualMax,Math.round(Number.isFinite(value)?value:min)));
}
export function readComposeHeight(value:string|null):number|null{
 if(value==null||!value.trim())return null;
 const height=Number(value);return Number.isFinite(height)&&height>0?height:null;
}

