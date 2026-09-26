let sequence=0;
// DOM identity only, never persisted or used as a model/session identifier.
export function nextModelPanelId(){return `gi-model-panel-${++sequence}`;}
export function modelPanelOptionId(panel:string,label:string){return `${panel}-option-${encodeURIComponent(label)}`;}
