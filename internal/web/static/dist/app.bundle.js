var m4,b1,e3,h$,J2,g3,$8,j8,R5,R4,j4,q8,m5,x5,v5,c$,v4={},b4=[],l$=/acit|ex(?:s|g|n|p|$)|rph|grid|ows|mnc|ntw|ine[ch]|zoo|^ord|itera/i,g4=Array.isArray;function $2($,j){for(var q in j)$[q]=j[q];return $}function g5($){$&&$.parentNode&&$.parentNode.removeChild($)}function p5($,j,q){var B,_,Q,U={};for(Q in j)Q=="key"?B=j[Q]:Q=="ref"?_=j[Q]:U[Q]=j[Q];if(arguments.length>2&&(U.children=arguments.length>3?m4.call(arguments,2):q),typeof $=="function"&&$.defaultProps!=null)for(Q in $.defaultProps)U[Q]===void 0&&(U[Q]=$.defaultProps[Q]);return w4($,U,B,_,null)}function w4($,j,q,B,_){var Q={type:$,props:j,key:q,ref:B,__k:null,__:null,__b:0,__e:null,__c:null,constructor:void 0,__v:_==null?++e3:_,__i:-1,__u:0};return _==null&&b1.vnode!=null&&b1.vnode(Q),Q}function p4($){return $.children}function q4($,j){this.props=$,this.context=j}function b2($,j){if(j==null)return $.__?b2($.__,$.__i+1):null;for(var q;j<$.__k.length;j++)if((q=$.__k[j])!=null&&q.__e!=null)return q.__e;return typeof $.type=="function"?b2($):null}function r$($){if($.__P&&$.__d){var j=$.__v,q=j.__e,B=[],_=[],Q=$2({},j);Q.__v=j.__v+1,b1.vnode&&b1.vnode(Q),h5($.__P,Q,j,$.__n,$.__P.namespaceURI,32&j.__u?[q]:null,B,q==null?b2(j):q,!!(32&j.__u),_),Q.__v=j.__v,Q.__.__k[Q.__i]=Q,U8(B,Q,_),j.__e=j.__=null,Q.__e!=q&&B8(Q)}}function B8($){if(($=$.__)!=null&&$.__c!=null)return $.__e=$.__c.base=null,$.__k.some(function(j){if(j!=null&&j.__e!=null)return $.__e=$.__c.base=j.__e}),B8($)}function p3($){(!$.__d&&($.__d=!0)&&J2.push($)&&!u4.__r++||g3!=b1.debounceRendering)&&((g3=b1.debounceRendering)||$8)(u4)}function u4(){try{for(var $,j=1;J2.length;)J2.length>j&&J2.sort(j8),$=J2.shift(),j=J2.length,r$($)}finally{J2.length=u4.__r=0}}function _8($,j,q,B,_,Q,U,X,J,W,V){var G,L,C,b,c,l,w,S=B&&B.__k||b4,m=j.length;for(J=d$(q,j,S,J,m),G=0;G<m;G++)(C=q.__k[G])!=null&&(L=C.__i!=-1&&S[C.__i]||v4,C.__i=G,l=h5($,C,L,_,Q,U,X,J,W,V),b=C.__e,C.ref&&L.ref!=C.ref&&(L.ref&&c5(L.ref,null,C),V.push(C.ref,C.__c||b,C)),c==null&&b!=null&&(c=b),(w=!!(4&C.__u))||L.__k===C.__k?(J=Q8(C,J,$,w),w&&L.__e&&(L.__e=null)):typeof C.type=="function"&&l!==void 0?J=l:b&&(J=b.nextSibling),C.__u&=-7);return q.__e=c,J}function d$($,j,q,B,_){var Q,U,X,J,W,V=q.length,G=V,L=0;for($.__k=Array(_),Q=0;Q<_;Q++)(U=j[Q])!=null&&typeof U!="boolean"&&typeof U!="function"?(typeof U=="string"||typeof U=="number"||typeof U=="bigint"||U.constructor==String?U=$.__k[Q]=w4(null,U,null,null,null):g4(U)?U=$.__k[Q]=w4(p4,{children:U},null,null,null):U.constructor===void 0&&U.__b>0?U=$.__k[Q]=w4(U.type,U.props,U.key,U.ref?U.ref:null,U.__v):$.__k[Q]=U,J=Q+L,U.__=$,U.__b=$.__b+1,X=null,(W=U.__i=i$(U,q,J,G))!=-1&&(G--,(X=q[W])&&(X.__u|=2)),X==null||X.__v==null?(W==-1&&(_>V?L--:_<V&&L++),typeof U.type!="function"&&(U.__u|=4)):W!=J&&(W==J-1?L--:W==J+1?L++:(W>J?L--:L++,U.__u|=4))):$.__k[Q]=null;if(G)for(Q=0;Q<V;Q++)(X=q[Q])!=null&&(2&X.__u)==0&&(X.__e==B&&(B=b2(X)),W8(X,X));return B}function Q8($,j,q,B){var _,Q;if(typeof $.type=="function"){for(_=$.__k,Q=0;_&&Q<_.length;Q++)_[Q]&&(_[Q].__=$,j=Q8(_[Q],j,q,B));return j}$.__e!=j&&(B&&(j&&$.type&&!j.parentNode&&(j=b2($)),q.insertBefore($.__e,j||null)),j=$.__e);do j=j&&j.nextSibling;while(j!=null&&j.nodeType==8);return j}function i$($,j,q,B){var _,Q,U,X=$.key,J=$.type,W=j[q],V=W!=null&&(2&W.__u)==0;if(W===null&&X==null||V&&X==W.key&&J==W.type)return q;if(B>(V?1:0)){for(_=q-1,Q=q+1;_>=0||Q<j.length;)if((W=j[U=_>=0?_--:Q++])!=null&&(2&W.__u)==0&&X==W.key&&J==W.type)return U}return-1}function h3($,j,q){j[0]=="-"?$.setProperty(j,q==null?"":q):$[j]=q==null?"":typeof q!="number"||l$.test(j)?q:q+"px"}function P4($,j,q,B,_){var Q,U;$:if(j=="style")if(typeof q=="string")$.style.cssText=q;else{if(typeof B=="string"&&($.style.cssText=B=""),B)for(j in B)q&&j in q||h3($.style,j,"");if(q)for(j in q)B&&q[j]==B[j]||h3($.style,j,q[j])}else if(j[0]=="o"&&j[1]=="n")Q=j!=(j=j.replace(q8,"$1")),U=j.toLowerCase(),j=U in $||j=="onFocusOut"||j=="onFocusIn"?U.slice(2):j.slice(2),$.l||($.l={}),$.l[j+Q]=q,q?B?q[j4]=B[j4]:(q[j4]=m5,$.addEventListener(j,Q?v5:x5,Q)):$.removeEventListener(j,Q?v5:x5,Q);else{if(_=="http://www.w3.org/2000/svg")j=j.replace(/xlink(H|:h)/,"h").replace(/sName$/,"s");else if(j!="width"&&j!="height"&&j!="href"&&j!="list"&&j!="form"&&j!="tabIndex"&&j!="download"&&j!="rowSpan"&&j!="colSpan"&&j!="role"&&j!="popover"&&j in $)try{$[j]=q==null?"":q;break $}catch(X){}typeof q=="function"||(q==null||q===!1&&j[4]!="-"?$.removeAttribute(j):$.setAttribute(j,j=="popover"&&q==1?"":q))}}function c3($){return function(j){if(this.l){var q=this.l[j.type+$];if(j[R4]==null)j[R4]=m5++;else if(j[R4]<q[j4])return;return q(b1.event?b1.event(j):j)}}}function h5($,j,q,B,_,Q,U,X,J,W){var V,G,L,C,b,c,l,w,S,m,O,d,g,U1,k,r=j.type;if(j.constructor!==void 0)return null;128&q.__u&&(J=!!(32&q.__u),Q=[X=j.__e=q.__e]),(V=b1.__b)&&V(j);$:if(typeof r=="function")try{if(w=j.props,S=r.prototype&&r.prototype.render,m=(V=r.contextType)&&B[V.__c],O=V?m?m.props.value:V.__:B,q.__c?l=(G=j.__c=q.__c).__=G.__E:(S?j.__c=G=new r(w,O):(j.__c=G=new q4(w,O),G.constructor=r,G.render=s$),m&&m.sub(G),G.state||(G.state={}),G.__n=B,L=G.__d=!0,G.__h=[],G._sb=[]),S&&G.__s==null&&(G.__s=G.state),S&&r.getDerivedStateFromProps!=null&&(G.__s==G.state&&(G.__s=$2({},G.__s)),$2(G.__s,r.getDerivedStateFromProps(w,G.__s))),C=G.props,b=G.state,G.__v=j,L)S&&r.getDerivedStateFromProps==null&&G.componentWillMount!=null&&G.componentWillMount(),S&&G.componentDidMount!=null&&G.__h.push(G.componentDidMount);else{if(S&&r.getDerivedStateFromProps==null&&w!==C&&G.componentWillReceiveProps!=null&&G.componentWillReceiveProps(w,O),j.__v==q.__v||!G.__e&&G.shouldComponentUpdate!=null&&G.shouldComponentUpdate(w,G.__s,O)===!1){j.__v!=q.__v&&(G.props=w,G.state=G.__s,G.__d=!1),j.__e=q.__e,j.__k=q.__k,j.__k.some(function(f){f&&(f.__=j)}),b4.push.apply(G.__h,G._sb),G._sb=[],G.__h.length&&U.push(G);break $}G.componentWillUpdate!=null&&G.componentWillUpdate(w,G.__s,O),S&&G.componentDidUpdate!=null&&G.__h.push(function(){G.componentDidUpdate(C,b,c)})}if(G.context=O,G.props=w,G.__P=$,G.__e=!1,d=b1.__r,g=0,S)G.state=G.__s,G.__d=!1,d&&d(j),V=G.render(G.props,G.state,G.context),b4.push.apply(G.__h,G._sb),G._sb=[];else do G.__d=!1,d&&d(j),V=G.render(G.props,G.state,G.context),G.state=G.__s;while(G.__d&&++g<25);G.state=G.__s,G.getChildContext!=null&&(B=$2($2({},B),G.getChildContext())),S&&!L&&G.getSnapshotBeforeUpdate!=null&&(c=G.getSnapshotBeforeUpdate(C,b)),U1=V!=null&&V.type===p4&&V.key==null?X8(V.props.children):V,X=_8($,g4(U1)?U1:[U1],j,q,B,_,Q,U,X,J,W),G.base=j.__e,j.__u&=-161,G.__h.length&&U.push(G),l&&(G.__E=G.__=null)}catch(f){if(j.__v=null,J||Q!=null)if(f.then){for(j.__u|=J?160:128;X&&X.nodeType==8&&X.nextSibling;)X=X.nextSibling;Q[Q.indexOf(X)]=null,j.__e=X}else{for(k=Q.length;k--;)g5(Q[k]);b5(j)}else j.__e=q.__e,j.__k=q.__k,f.then||b5(j);b1.__e(f,j,q)}else Q==null&&j.__v==q.__v?(j.__k=q.__k,j.__e=q.__e):X=j.__e=n$(q.__e,j,q,B,_,Q,U,J,W);return(V=b1.diffed)&&V(j),128&j.__u?void 0:X}function b5($){$&&($.__c&&($.__c.__e=!0),$.__k&&$.__k.some(b5))}function U8($,j,q){for(var B=0;B<q.length;B++)c5(q[B],q[++B],q[++B]);b1.__c&&b1.__c(j,$),$.some(function(_){try{$=_.__h,_.__h=[],$.some(function(Q){Q.call(_)})}catch(Q){b1.__e(Q,_.__v)}})}function X8($){return typeof $!="object"||$==null||$.__b>0?$:g4($)?$.map(X8):$2({},$)}function n$($,j,q,B,_,Q,U,X,J){var W,V,G,L,C,b,c,l=q.props||v4,w=j.props,S=j.type;if(S=="svg"?_="http://www.w3.org/2000/svg":S=="math"?_="http://www.w3.org/1998/Math/MathML":_||(_="http://www.w3.org/1999/xhtml"),Q!=null){for(W=0;W<Q.length;W++)if((C=Q[W])&&"setAttribute"in C==!!S&&(S?C.localName==S:C.nodeType==3)){$=C,Q[W]=null;break}}if($==null){if(S==null)return document.createTextNode(w);$=document.createElementNS(_,S,w.is&&w),X&&(b1.__m&&b1.__m(j,Q),X=!1),Q=null}if(S==null)l===w||X&&$.data==w||($.data=w);else{if(Q=Q&&m4.call($.childNodes),!X&&Q!=null)for(l={},W=0;W<$.attributes.length;W++)l[(C=$.attributes[W]).name]=C.value;for(W in l)C=l[W],W=="dangerouslySetInnerHTML"?G=C:W=="children"||(W in w)||W=="value"&&("defaultValue"in w)||W=="checked"&&("defaultChecked"in w)||P4($,W,null,C,_);for(W in w)C=w[W],W=="children"?L=C:W=="dangerouslySetInnerHTML"?V=C:W=="value"?b=C:W=="checked"?c=C:X&&typeof C!="function"||l[W]===C||P4($,W,C,l[W],_);if(V)X||G&&(V.__html==G.__html||V.__html==$.innerHTML)||($.innerHTML=V.__html),j.__k=[];else if(G&&($.innerHTML=""),_8(j.type=="template"?$.content:$,g4(L)?L:[L],j,q,B,S=="foreignObject"?"http://www.w3.org/1999/xhtml":_,Q,U,Q?Q[0]:q.__k&&b2(q,0),X,J),Q!=null)for(W=Q.length;W--;)g5(Q[W]);X||(W="value",S=="progress"&&b==null?$.removeAttribute("value"):b!=null&&(b!==$[W]||S=="progress"&&!b||S=="option"&&b!=l[W])&&P4($,W,b,l[W],_),W="checked",c!=null&&c!=$[W]&&P4($,W,c,l[W],_))}return $}function c5($,j,q){try{if(typeof $=="function"){var B=typeof $.__u=="function";B&&$.__u(),B&&j==null||($.__u=$(j))}else $.current=j}catch(_){b1.__e(_,q)}}function W8($,j,q){var B,_;if(b1.unmount&&b1.unmount($),(B=$.ref)&&(B.current&&B.current!=$.__e||c5(B,null,j)),(B=$.__c)!=null){if(B.componentWillUnmount)try{B.componentWillUnmount()}catch(Q){b1.__e(Q,j)}B.base=B.__P=null}if(B=$.__k)for(_=0;_<B.length;_++)B[_]&&W8(B[_],j,q||typeof $.type!="function");q||g5($.__e),$.__c=$.__=$.__e=void 0}function s$($,j,q){return this.constructor($,q)}function m2($,j,q){var B,_,Q,U;j==document&&(j=document.documentElement),b1.__&&b1.__($,j),_=(B=typeof q=="function")?null:q&&q.__k||j.__k,Q=[],U=[],h5(j,$=(!B&&q||j).__k=p5(p4,null,[$]),_||v4,v4,j.namespaceURI,!B&&q?[q]:_?null:j.firstChild?m4.call(j.childNodes):null,Q,!B&&q?q:_?_.__e:j.firstChild,B,U),U8(Q,$,U)}m4=b4.slice,b1={__e:function($,j,q,B){for(var _,Q,U;j=j.__;)if((_=j.__c)&&!_.__)try{if((Q=_.constructor)&&Q.getDerivedStateFromError!=null&&(_.setState(Q.getDerivedStateFromError($)),U=_.__d),_.componentDidCatch!=null&&(_.componentDidCatch($,B||{}),U=_.__d),U)return _.__E=_}catch(X){$=X}throw $}},e3=0,h$=function($){return $!=null&&$.constructor===void 0},q4.prototype.setState=function($,j){var q;q=this.__s!=null&&this.__s!=this.state?this.__s:this.__s=$2({},this.state),typeof $=="function"&&($=$($2({},q),this.props)),$&&$2(q,$),$!=null&&this.__v&&(j&&this._sb.push(j),p3(this))},q4.prototype.forceUpdate=function($){this.__v&&(this.__e=!0,$&&this.__h.push($),p3(this))},q4.prototype.render=p4,J2=[],$8=typeof Promise=="function"?Promise.prototype.then.bind(Promise.resolve()):setTimeout,j8=function($,j){return $.__v.__b-j.__v.__b},u4.__r=0,R5=Math.random().toString(8),R4="__d"+R5,j4="__a"+R5,q8=/(PointerCapture)$|Capture$/i,m5=0,x5=c3(!1),v5=c3(!0),c$=0;var u2,a1,w5,l3,B4=0,J8=[],o1=b1,r3=o1.__b,d3=o1.__r,i3=o1.diffed,n3=o1.__c,s3=o1.unmount,o3=o1.__;function h4($,j){o1.__h&&o1.__h(a1,$,B4||j),B4=0;var q=a1.__H||(a1.__H={__:[],__h:[]});return $>=q.__.length&&q.__.push({}),q.__[$]}function y($){return B4=1,K8(z8,$)}function K8($,j,q){var B=h4(u2++,2);if(B.t=$,!B.__c&&(B.__=[q?q(j):z8(void 0,j),function(X){var J=B.__N?B.__N[0]:B.__[0],W=B.t(J,X);J!==W&&(B.__N=[W,B.__[1]],B.__c.setState({}))}],B.__c=a1,!a1.__f)){var _=function(X,J,W){if(!B.__c.__H)return!0;var V=B.__c.__H.__.filter(function(L){return L.__c});if(V.every(function(L){return!L.__N}))return!Q||Q.call(this,X,J,W);var G=B.__c.props!==X;return V.some(function(L){if(L.__N){var C=L.__[0];L.__=L.__N,L.__N=void 0,C!==L.__[0]&&(G=!0)}}),Q&&Q.call(this,X,J,W)||G};a1.__f=!0;var{shouldComponentUpdate:Q,componentWillUpdate:U}=a1;a1.componentWillUpdate=function(X,J,W){if(this.__e){var V=Q;Q=void 0,_(X,J,W),Q=V}U&&U.call(this,X,J,W)},a1.shouldComponentUpdate=_}return B.__N||B.__}function E($,j){var q=h4(u2++,3);!o1.__s&&r5(q.__H,j)&&(q.__=$,q.u=j,a1.__H.__h.push(q))}function l5($,j){var q=h4(u2++,4);!o1.__s&&r5(q.__H,j)&&(q.__=$,q.u=j,a1.__h.push(q))}function M($){return B4=5,B1(function(){return{current:$}},[])}function B1($,j){var q=h4(u2++,7);return r5(q.__H,j)&&(q.__=$(),q.__H=j,q.__h=$),q.__}function i($,j){return B4=8,B1(function(){return $},j)}function o$(){for(var $;$=J8.shift();){var j=$.__H;if($.__P&&j)try{j.__h.some(x4),j.__h.some(u5),j.__h=[]}catch(q){j.__h=[],o1.__e(q,$.__v)}}}o1.__b=function($){a1=null,r3&&r3($)},o1.__=function($,j){$&&j.__k&&j.__k.__m&&($.__m=j.__k.__m),o3&&o3($,j)},o1.__r=function($){d3&&d3($),u2=0;var j=(a1=$.__c).__H;j&&(w5===a1?(j.__h=[],a1.__h=[],j.__.some(function(q){q.__N&&(q.__=q.__N),q.u=q.__N=void 0})):(j.__h.some(x4),j.__h.some(u5),j.__h=[],u2=0)),w5=a1},o1.diffed=function($){i3&&i3($);var j=$.__c;j&&j.__H&&(j.__H.__h.length&&(J8.push(j)!==1&&l3===o1.requestAnimationFrame||((l3=o1.requestAnimationFrame)||a$)(o$)),j.__H.__.some(function(q){q.u&&(q.__H=q.u),q.u=void 0})),w5=a1=null},o1.__c=function($,j){j.some(function(q){try{q.__h.some(x4),q.__h=q.__h.filter(function(B){return!B.__||u5(B)})}catch(B){j.some(function(_){_.__h&&(_.__h=[])}),j=[],o1.__e(B,q.__v)}}),n3&&n3($,j)},o1.unmount=function($){s3&&s3($);var j,q=$.__c;q&&q.__H&&(q.__H.__.some(function(B){try{x4(B)}catch(_){j=_}}),q.__H=void 0,j&&o1.__e(j,q.__v))};var a3=typeof requestAnimationFrame=="function";function a$($){var j,q=function(){clearTimeout(B),a3&&cancelAnimationFrame(j),setTimeout($)},B=setTimeout(q,35);a3&&(j=requestAnimationFrame(q))}function x4($){var j=a1,q=$.__c;typeof q=="function"&&($.__c=void 0,q()),a1=j}function u5($){var j=a1;$.__c=$.__(),a1=j}function r5($,j){return!$||$.length!==j.length||j.some(function(q,B){return q!==$[B]})}function z8($,j){return typeof j=="function"?j($):j}var G8=function($,j,q,B){var _;j[0]=0;for(var Q=1;Q<j.length;Q++){var U=j[Q++],X=j[Q]?(j[0]|=U?1:2,q[j[Q++]]):j[++Q];U===3?B[0]=X:U===4?B[1]=Object.assign(B[1]||{},X):U===5?(B[1]=B[1]||{})[j[++Q]]=X:U===6?B[1][j[++Q]]+=X+"":U?(_=$.apply(X,G8($,X,q,["",null])),B.push(_),X[0]?j[0]|=2:(j[Q-2]=0,j[Q]=_)):B.push(X)}return B},t3=new Map;function t$($){var j=t3.get(this);return j||(j=new Map,t3.set(this,j)),(j=G8(this,j.get($)||(j.set($,j=function(q){for(var B,_,Q=1,U="",X="",J=[0],W=function(L){Q===1&&(L||(U=U.replace(/^\s*\n\s*|\s*\n\s*$/g,"")))?J.push(0,L,U):Q===3&&(L||U)?(J.push(3,L,U),Q=2):Q===2&&U==="..."&&L?J.push(4,L,0):Q===2&&U&&!L?J.push(5,0,!0,U):Q>=5&&((U||!L&&Q===5)&&(J.push(Q,0,U,_),Q=6),L&&(J.push(Q,L,0,_),Q=6)),U=""},V=0;V<q.length;V++){V&&(Q===1&&W(),W(V));for(var G=0;G<q[V].length;G++)B=q[V][G],Q===1?B==="<"?(W(),J=[J],Q=3):U+=B:Q===4?U==="--"&&B===">"?(Q=1,U=""):U=B+U[0]:X?B===X?X="":U+=B:B==='"'||B==="'"?X=B:B===">"?(W(),Q=1):Q&&(B==="="?(Q=5,_=U,U=""):B==="/"&&(Q<5||q[V][G+1]===">")?(W(),Q===3&&(J=J[0]),Q=J,(J=J[0]).push(2,0,Q),Q=0):B===" "||B==="\t"||B===`
`||B==="\r"?(W(),Q=2):U+=B),Q===3&&U==="!--"&&(Q=4,J=J[0])}return W(),J}($)),j),arguments,[])).length>1?j:j[0]}var K=t$.bind(p5);function E0($){if(typeof window>"u"||!window.localStorage)return null;try{return window.localStorage.getItem($)}catch{return null}}function R0($,j){if(typeof window>"u"||!window.localStorage)return;try{window.localStorage.setItem($,j)}catch{return}}function Z8($,j=!1){let q=E0($);if(q===null)return j;return q==="true"}function V8($,j=null){let q=E0($);if(q===null)return j;let B=parseInt(q,10);return Number.isFinite(B)?B:j}var d5=($)=>{let j=new Set;return($||[]).filter((q)=>{if(!q||j.has(q.id))return!1;return j.add(q.id),!0})};function L8(){let[$,j]=y(null),[q,B]=y({text:"",totalLines:0}),[_,Q]=y(""),[U,X]=y({text:"",totalLines:0}),[J,W]=y(null),[V,G]=y(null),[L,C]=y(null),b=M(null),c=M(0),l=M(!1),w=M(""),S=M(""),m=M(!1),O=M(0),d=M(null),g=M(null),U1=M(null),k=M(null),r=M(!1),f=M(!1);return{agentStatus:$,setAgentStatus:j,agentDraft:q,setAgentDraft:B,agentPlan:_,setAgentPlan:Q,agentThought:U,setAgentThought:X,pendingRequest:J,setPendingRequest:W,currentTurnId:V,setCurrentTurnId:G,steerQueuedTurnId:L,setSteerQueuedTurnId:C,lastAgentEventRef:b,lastSilenceNoticeRef:c,isAgentRunningRef:l,draftBufferRef:w,thoughtBufferRef:S,previewResyncPendingRef:m,previewResyncGenerationRef:O,pendingRequestRef:d,stalledPostIdRef:g,currentTurnIdRef:U1,steerQueuedTurnIdRef:k,thoughtExpandedRef:r,draftExpandedRef:f}}var H8="piclaw_theme",n5="piclaw_tint",e$="piclaw_chat_themes",q2={bgPrimary:"#ffffff",bgSecondary:"#f7f9fa",bgHover:"#e8ebed",textPrimary:"#0f1419",textSecondary:"#536471",borderColor:"#eff3f4",accent:"#1d9bf0",accentHover:"#1a8cd8",warning:"#f0b429",danger:"#f4212e",success:"#00ba7c"},r4={bgPrimary:"#000000",bgSecondary:"#16181c",bgHover:"#1d1f23",textPrimary:"#e7e9ea",textSecondary:"#71767b",borderColor:"#2f3336",accent:"#1d9bf0",accentHover:"#1a8cd8",warning:"#f0b429",danger:"#f4212e",success:"#00ba7c"},N8={default:{label:"Default",mode:"auto",light:q2,dark:r4},tango:{label:"Tango",mode:"light",light:{bgPrimary:"#f6f5f4",bgSecondary:"#efedeb",bgHover:"#e5e3e1",textPrimary:"#2e3436",textSecondary:"#5c6466",borderColor:"#d3d7cf",accent:"#3465a4",accentHover:"#2c5890",danger:"#cc0000",success:"#4e9a06"}},xterm:{label:"XTerm",mode:"dark",dark:{bgPrimary:"#000000",bgSecondary:"#0a0a0a",bgHover:"#121212",textPrimary:"#d0d0d0",textSecondary:"#8a8a8a",borderColor:"#1f1f1f",accent:"#00a2ff",accentHover:"#0086d1",danger:"#ff5f5f",success:"#5fff87"}},monokai:{label:"Monokai",mode:"dark",dark:{bgPrimary:"#272822",bgSecondary:"#2f2f2f",bgHover:"#3a3a3a",textPrimary:"#f8f8f2",textSecondary:"#cfcfc2",borderColor:"#3e3d32",accent:"#f92672",accentHover:"#e81560",danger:"#f92672",success:"#a6e22e"}},"monokai-pro":{label:"Monokai Pro",mode:"dark",dark:{bgPrimary:"#2d2a2e",bgSecondary:"#363237",bgHover:"#403a40",textPrimary:"#fcfcfa",textSecondary:"#c1c0c0",borderColor:"#444046",accent:"#ff6188",accentHover:"#f74f7e",danger:"#ff4f5e",success:"#a9dc76"}},ristretto:{label:"Ristretto",mode:"dark",dark:{bgPrimary:"#2c2525",bgSecondary:"#362d2d",bgHover:"#403535",textPrimary:"#f4f1ef",textSecondary:"#cbbdb8",borderColor:"#4a3c3c",accent:"#ff9f43",accentHover:"#f28a2e",danger:"#ff5f56",success:"#a9dc76"}},dracula:{label:"Dracula",mode:"dark",dark:{bgPrimary:"#282a36",bgSecondary:"#303445",bgHover:"#3a3f52",textPrimary:"#f8f8f2",textSecondary:"#c5c8d6",borderColor:"#44475a",accent:"#bd93f9",accentHover:"#a87ded",danger:"#ff5555",success:"#50fa7b"}},catppuccin:{label:"Catppuccin",mode:"dark",dark:{bgPrimary:"#1e1e2e",bgSecondary:"#24273a",bgHover:"#2c2f41",textPrimary:"#cdd6f4",textSecondary:"#a6adc8",borderColor:"#313244",accent:"#89b4fa",accentHover:"#74a0f5",danger:"#f38ba8",success:"#a6e3a1"}},nord:{label:"Nord",mode:"dark",dark:{bgPrimary:"#2e3440",bgSecondary:"#3b4252",bgHover:"#434c5e",textPrimary:"#eceff4",textSecondary:"#d8dee9",borderColor:"#4c566a",accent:"#88c0d0",accentHover:"#78a9c0",danger:"#bf616a",success:"#a3be8c"}},gruvbox:{label:"Gruvbox",mode:"dark",dark:{bgPrimary:"#282828",bgSecondary:"#32302f",bgHover:"#3c3836",textPrimary:"#ebdbb2",textSecondary:"#bdae93",borderColor:"#3c3836",accent:"#d79921",accentHover:"#c28515",danger:"#fb4934",success:"#b8bb26"}},solarized:{label:"Solarized",mode:"auto",light:{bgPrimary:"#fdf6e3",bgSecondary:"#f5efdc",bgHover:"#eee8d5",textPrimary:"#586e75",textSecondary:"#657b83",borderColor:"#e0d8c6",accent:"#268bd2",accentHover:"#1f78b3",danger:"#dc322f",success:"#859900"},dark:{bgPrimary:"#002b36",bgSecondary:"#073642",bgHover:"#0b3c4a",textPrimary:"#eee8d5",textSecondary:"#93a1a1",borderColor:"#18424a",accent:"#268bd2",accentHover:"#1f78b3",danger:"#dc322f",success:"#859900"}},tokyo:{label:"Tokyo",mode:"dark",dark:{bgPrimary:"#1a1b26",bgSecondary:"#24283b",bgHover:"#2f3549",textPrimary:"#c0caf5",textSecondary:"#9aa5ce",borderColor:"#414868",accent:"#7aa2f7",accentHover:"#6b92e6",danger:"#f7768e",success:"#9ece6a"}},miasma:{label:"Miasma",mode:"dark",dark:{bgPrimary:"#1f1f23",bgSecondary:"#29292f",bgHover:"#33333a",textPrimary:"#e5e5e5",textSecondary:"#b4b4b4",borderColor:"#3d3d45",accent:"#c9739c",accentHover:"#b8618c",danger:"#e06c75",success:"#98c379"}},github:{label:"GitHub",mode:"auto",light:{bgPrimary:"#ffffff",bgSecondary:"#f6f8fa",bgHover:"#eaeef2",textPrimary:"#24292f",textSecondary:"#57606a",borderColor:"#d0d7de",accent:"#0969da",accentHover:"#0550ae",danger:"#cf222e",success:"#1a7f37"},dark:{bgPrimary:"#0d1117",bgSecondary:"#161b22",bgHover:"#21262d",textPrimary:"#c9d1d9",textSecondary:"#8b949e",borderColor:"#30363d",accent:"#2f81f7",accentHover:"#1f6feb",danger:"#f85149",success:"#3fb950"}},gotham:{label:"Gotham",mode:"dark",dark:{bgPrimary:"#0b0f14",bgSecondary:"#111720",bgHover:"#18212b",textPrimary:"#cbd6e2",textSecondary:"#9bb0c3",borderColor:"#1f2a37",accent:"#5ccfe6",accentHover:"#48b8ce",danger:"#d26937",success:"#2aa889"}}},$9=["--bg-primary","--bg-secondary","--bg-hover","--text-primary","--text-secondary","--border-color","--accent-color","--accent-hover","--accent-color-alpha","--accent-contrast-text","--accent-soft","--accent-soft-strong","--warning-color","--danger-color","--success-color","--search-highlight-color"],y2={theme:"default",tint:null},Y8="light",i5=!1;function d4($){let j=String($||"").trim().toLowerCase();if(!j)return"default";if(j==="solarized-dark"||j==="solarized-light")return"solarized";if(j==="github-dark"||j==="github-light")return"github";if(j==="tokyo-night")return"tokyo";return j}function A8($){if(!$)return null;let j=String($).trim();if(!j)return null;let q=j.startsWith("#")?j.slice(1):j;if(!/^[0-9a-fA-F]{3}$/.test(q)&&!/^[0-9a-fA-F]{6}$/.test(q))return null;let B=q.length===3?q.split("").map((Q)=>Q+Q).join(""):q,_=parseInt(B,16);return{r:_>>16&255,g:_>>8&255,b:_&255,hex:`#${B.toLowerCase()}`}}function j9($,j){try{if(document.body){$.style.display="none",document.body.appendChild($);let q=getComputedStyle($).color||$.style.color;return document.body.removeChild($),q}}catch{return j}return j}function q9($){if(!$||typeof document>"u")return null;let j=String($).trim();if(!j)return null;let q=document.createElement("div");if(q.style.color="",q.style.color=j,!q.style.color)return null;let _=j9(q,q.style.color).match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/i);if(!_)return null;let Q=parseInt(_[1],10),U=parseInt(_[2],10),X=parseInt(_[3],10);if(![Q,U,X].every((W)=>Number.isFinite(W)))return null;let J=`#${[Q,U,X].map((W)=>W.toString(16).padStart(2,"0")).join("")}`;return{r:Q,g:U,b:X,hex:J}}function M2($){return A8($)||q9($)}function s5($,j,q){let B=Math.round($.r+(j.r-$.r)*q),_=Math.round($.g+(j.g-$.g)*q),Q=Math.round($.b+(j.b-$.b)*q);return`rgb(${B} ${_} ${Q})`}function c4($,j){return`rgba(${$.r}, ${$.g}, ${$.b}, ${j})`}function B9($){let j=$.r/255,q=$.g/255,B=$.b/255,_=j<=0.03928?j/12.92:Math.pow((j+0.055)/1.055,2.4),Q=q<=0.03928?q/12.92:Math.pow((q+0.055)/1.055,2.4),U=B<=0.03928?B/12.92:Math.pow((B+0.055)/1.055,2.4);return 0.2126*_+0.7152*Q+0.0722*U}function _9($){return B9($)>0.4?"#000000":"#ffffff"}function D8(){if(typeof window>"u")return"light";try{return window.matchMedia&&window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light"}catch{return"light"}}function o5($){return N8[$]||N8.default}function Q9($){return $.mode==="auto"?D8():$.mode}function C8($,j){let q=o5($);if(j==="dark"&&q.dark)return q.dark;if(j==="light"&&q.light)return q.light;return q.dark||q.light||q2}function j2($,j,q){let B=M2($);if(!B)return $;return s5(B,j,q)}function I8($,j,q){let B=M2(j);if(!B)return $;let Q=A8(q==="dark"?"#ffffff":"#000000");return{...$,bgPrimary:j2($.bgPrimary,B,0.08),bgSecondary:j2($.bgSecondary,B,0.12),bgHover:j2($.bgHover,B,0.16),textPrimary:j2($.textPrimary,B,q==="dark"?0.08:0.06),textSecondary:j2($.textSecondary,B,q==="dark"?0.12:0.1),borderColor:j2($.borderColor,B,0.1),accent:B.hex,accentHover:Q?s5(B,Q,0.18):B.hex,warning:j2($.warning||q2.warning,B,0.14),danger:j2($.danger,B,0.16),success:j2($.success,B,0.16)}}function U9($,j){let q=M2($?.warning);if(q)return q.hex;let B=M2(j==="dark"?r4.warning:q2.warning)||M2(q2.warning),_=M2($?.accent);if(B&&_)return s5(B,_,j==="dark"?0.18:0.14);return j==="dark"?r4.warning:q2.warning}function X9($,j){if(typeof document>"u")return;let q=document.documentElement,B=$.accent,_=M2(B),Q=_?c4(_,j==="dark"?0.35:0.2):$.searchHighlight||$.searchHighlightColor,U=_?c4(_,j==="dark"?0.16:0.12):"rgba(29, 155, 240, 0.12)",X=_?c4(_,j==="dark"?0.28:0.2):"rgba(29, 155, 240, 0.2)",J=_?_9(_):j==="dark"?"#000000":"#ffffff",W=_?c4(_,j==="dark"?0.35:0.25):"rgba(29, 155, 240, 0.25)",V=U9($,j),G={"--bg-primary":$.bgPrimary,"--bg-secondary":$.bgSecondary,"--bg-hover":$.bgHover,"--text-primary":$.textPrimary,"--text-secondary":$.textSecondary,"--border-color":$.borderColor,"--accent-color":B,"--accent-hover":$.accentHover||B,"--accent-color-alpha":W,"--accent-soft":U,"--accent-soft-strong":X,"--accent-contrast-text":J,"--warning-color":V,"--danger-color":$.danger||q2.danger,"--success-color":$.success||q2.success,"--search-highlight-color":Q||"rgba(29, 155, 240, 0.2)"};Object.entries(G).forEach(([L,C])=>{if(C)q.style.setProperty(L,C)})}function W9(){if(typeof document>"u")return;let $=document.documentElement;$9.forEach((j)=>$.style.removeProperty(j))}function g2($,j={}){if(typeof document>"u")return null;let q=typeof j.id==="string"&&j.id.trim()?j.id.trim():null,B=q?document.getElementById(q):document.querySelector(`meta[name="${$}"]`);if(!B)B=document.createElement("meta"),document.head.appendChild(B);if(B.setAttribute("name",$),q)B.setAttribute("id",q);return B}function F8($){let j=d4(y2?.theme||"default"),q=y2?.tint?String(y2.tint).trim():null,B=C8(j,$);if(j==="default"&&q)B=I8(B,q,$);if(B?.bgPrimary)return B.bgPrimary;return $==="dark"?r4.bgPrimary:q2.bgPrimary}function J9($,j){if(typeof document>"u")return;let q=g2("theme-color",{id:"dynamic-theme-color"});if(q&&$)q.removeAttribute("media"),q.setAttribute("content",$);let B=g2("theme-color",{id:"theme-color-light"});if(B)B.setAttribute("media","(prefers-color-scheme: light)"),B.setAttribute("content",F8("light"));let _=g2("theme-color",{id:"theme-color-dark"});if(_)_.setAttribute("media","(prefers-color-scheme: dark)"),_.setAttribute("content",F8("dark"));let Q=g2("msapplication-TileColor");if(Q&&$)Q.setAttribute("content",$);let U=g2("msapplication-navbutton-color");if(U&&$)U.setAttribute("content",$);let X=g2("apple-mobile-web-app-status-bar-style");if(X)X.setAttribute("content",j==="dark"?"black-translucent":"default")}function K9(){if(typeof window>"u")return;let $={...y2,mode:Y8};window.dispatchEvent(new CustomEvent("piclaw-theme-change",{detail:$}))}function z9(){try{let $=E0(e$);if(!$)return{};let j=JSON.parse($);return typeof j==="object"&&j!==null?j:{}}catch{return{}}}function G9($){if(!$)return null;return z9()[$]||null}function Z9(){if(typeof window>"u")return"web:default";try{let j=new URL(window.location.href).searchParams.get("chat_jid");return j&&j.trim()?j.trim():"web:default"}catch{return"web:default"}}function O8($,j={}){if(typeof window>"u"||typeof document>"u")return;let q=d4($?.theme||"default"),B=$?.tint?String($.tint).trim():null,_=o5(q),Q=Q9(_),U=C8(q,Q);y2={theme:q,tint:B},Y8=Q;let X=document.documentElement;X.dataset.theme=Q,X.dataset.colorTheme=q,X.dataset.tint=B?String(B):"",X.style.colorScheme=Q;let J=U;if(q==="default"&&B)J=I8(U,B,Q);if(q==="default"&&!B)W9();else X9(J,Q);if(J9(J.bgPrimary,Q),K9(),j.persist!==!1)if(R0(H8,q),B)R0(n5,B);else R0(n5,"")}function l4(){if(o5(y2.theme).mode!=="auto")return;O8(y2,{persist:!1})}function T8(){if(typeof window>"u")return()=>{};let $=Z9(),j=G9($),q=j?d4(j.theme||"default"):d4(E0(H8)||"default"),B=j?j.tint?String(j.tint).trim():null:(()=>{let _=E0(n5);return _?_.trim():null})();if(O8({theme:q,tint:B},{persist:!1}),window.matchMedia&&!i5){let _=window.matchMedia("(prefers-color-scheme: dark)");if(_.addEventListener)_.addEventListener("change",l4);else if(_.addListener)_.addListener(l4);return i5=!0,()=>{if(_.removeEventListener)_.removeEventListener("change",l4);else if(_.removeListener)_.removeListener(l4);i5=!1}}return()=>{}}function M8(){if(typeof document>"u")return"light";let $=document.documentElement?.dataset?.theme;if($==="dark"||$==="light")return $;return D8()}function a5($,j){try{if(typeof window>"u")return j;let q=window.__PICLAW_SILENCE||{},B=`__PICLAW_SILENCE_${$.toUpperCase()}_MS`,_=q[$]??window[B],Q=Number(_);return Number.isFinite(Q)?Q:j}catch{return j}}var QB=a5("warning",30000),UB=a5("finalize",120000),XB=a5("refresh",30000);function y8(){if(/iPad|iPhone/.test(navigator.userAgent))return!0;return navigator.platform==="MacIntel"&&navigator.maxTouchPoints>1}function i4($){if(!$||typeof $!=="object")return null;let j=$.started_at??$.startedAt;if(typeof j!=="string"||!j)return null;let q=Date.parse(j);return Number.isFinite(q)?q:null}function V9($){if(!$||typeof $!=="object")return null;let j=$.retry_at??$.retryAt;if(typeof j!=="string"||!j)return null;let q=Date.parse(j);return Number.isFinite(q)?q:null}function t5($){if(!$||typeof $!=="object")return null;let j=$.last_event_at??$.lastEventAt??$.started_at??$.startedAt;if(typeof j!=="string"||!j)return null;let q=Date.parse(j);return Number.isFinite(q)?q:null}function K2($){if(!$||typeof $!=="object")return!1;let j=$.intent_key??$.intentKey;return $.type==="intent"&&j==="compaction"}function n4($){if(!$||typeof $!=="object")return"";let j=$.title;if(typeof j==="string"&&j.trim())return j.trim();let q=$.status;if(typeof q==="string"&&q.trim())return q.trim();return K2($)?"Compacting context":"Working..."}function S8($){let j=Math.max(0,Math.floor($/1000)),q=j%60,B=Math.floor(j/60)%60,_=Math.floor(j/3600);if(_>0)return`${_}:${String(B).padStart(2,"0")}:${String(q).padStart(2,"0")}`;return`${B}:${String(q).padStart(2,"0")}`}function s4($,j=Date.now()){let q=i4($);if(q===null)return null;return S8(Math.max(0,j-q))}function k8($,j=Date.now()){let q=V9($);if(q===null)return null;let B=q-j;if(B<=0)return"retrying now";return`retry in ${S8(B)}`}var L9="";async function z2($,j={}){let q=typeof performance<"u"&&typeof performance.now==="function"?performance.now():Date.now(),B;try{B=await fetch(L9+$,{...j,headers:{"Content-Type":"application/json",...j.headers||{}}})}catch(Q){throw E8({method:String(j.method||"GET").toUpperCase(),url:$,startedAt:q,durationMs:performance.now()-q,ok:!1,detail:{failedBeforeResponse:!0}}),Q}let _=performance.now()-q;if(E8({method:String(j.method||"GET").toUpperCase(),url:$,startedAt:q,durationMs:_,status:B.status,ok:B.ok,requestId:B.headers?.get?.("x-request-id")||null,serverTiming:B.headers?.get?.("Server-Timing")||null}),!B.ok){let Q=await B.json().catch(()=>({error:"Unknown error"}));throw Error(Q.error||`HTTP ${B.status}`)}return B.json()}async function f8($=50,j=null,q=null){let B=q?.startsWith("gi:")?q.slice(3):null;if(!B)return{posts:[]};let _=`/api/sessions/${encodeURIComponent(B)}/messages?limit=${$}`;if(j)_+=`&before=${j}`;return{posts:((await z2(_)).messages||[]).map((X)=>({id:X.id,chat_jid:q,content:X.content,timestamp:X.created_at,sender:X.role==="user"?"user":"agent",is_from_me:X.role==="user",is_bot_message:X.role==="assistant",data:{thread_id:null,agent_id:X.role==="assistant"?"gi":null,content_blocks:X.payload?.content_blocks||null,kind:X.payload?.kind||null,source:X.payload?.source||null,clipped:X.payload?.clipped||!1}}))}}async function P8($,j=null){let q=j?.startsWith("gi:")?j.slice(3):null;if(!q)return null;let _=(await z2(`/api/sessions/${encodeURIComponent(q)}/turns`).catch(()=>({turns:[]}))).turns||[],Q=_.find((U)=>U.status==="running"||U.status==="cancelling")||_.find((U)=>U.status==="queued");if(!Q)return null;return{type:Q.status==="running"?"tool_call":"intent",title:Q.status==="cancelling"?"Cancelling…":Q.status==="queued"?"Queued":Q.prompt,status:Q.status}}async function o4($=null){let j=await z2("/api/runtime/config").catch(()=>({}));return{models:(j.enabled_models||[]).map((B)=>({id:B,provider:j.default_provider||"",label:B})),current:j.default_model||""}}async function e5($,j,q=null,B=[],_=null,Q=null){let U=Q?.startsWith("gi:")?Q.slice(3):null;if(!U)throw Error("No active session");let X=_==="steer"?"steer":_==="queue"?"queue":"prompt";return z2(`/api/sessions/${encodeURIComponent(U)}/prompt`,{method:"POST",body:JSON.stringify({prompt:j,intent:X})})}async function R8($,j=null){return null}async function $3($){return z2(`/api/media/${$}`).catch(()=>null)}function i0($){return`/api/media/${$}/raw`}function w8($){return`/api/media/${$}/thumbnail`}async function x8($){return null}async function a4($=null){return z2("/api/workspace/tree")}async function v8($,j=null){return z2(`/api/workspace/file?path=${encodeURIComponent($)}`)}async function b8($=null){return{status:"ready",indexed_at:null}}async function u8($=null){return null}async function m8($,j,q=null){return z2("/api/workspace/file",{method:"POST",body:JSON.stringify({path:$,content:j})}).catch(()=>null)}async function g8($,j,q=null){return null}async function p8($,j,q=null){return null}async function h8($,j=null){return null}async function j3($,j,q=null){return null}async function t4($,j,q=null){return null}function q3($){return`/api/workspace/file?path=${encodeURIComponent($)}`}async function c8($=null){return null}function l8($,j={}){let q=new URLSearchParams({path:String($||"")});if(j?.download)q.set("download","1");return`/api/workspace/raw?${q.toString()}`}function B3($){return l8($,{download:!0})}async function E8($){}async function r8(...$){return null}import{classHighlighter as N9,highlightTree as F9,StreamLanguage as k2,cssLanguage as H9,goLanguage as Y9,htmlLanguage as A9,javascriptLanguage as D9,jsxLanguage as C9,tsxLanguage as I9,typescriptLanguage as O9,jsonLanguage as T9,markdownLanguage as M9,pythonLanguage as y9,StandardSQL as S9,xmlLanguage as k9,yamlLanguage as E9,dockerFile as f9,powerShell as P9,ruby as R9,rust as w9,shell as x9,swift as v9,toml as b9}from"#editor-vendor/codemirror";function S2($){return $.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;").replace(/'/g,"&#39;")}var u9={js:"JavaScript",javascript:"JavaScript",ts:"TypeScript",typescript:"TypeScript",jsx:"JSX",tsx:"TSX",py:"Python",python:"Python",sh:"Shell",shell:"Shell",bash:"Bash",zsh:"Zsh",ps1:"PowerShell",powershell:"PowerShell",md:"Markdown",markdown:"Markdown",yml:"YAML",yaml:"YAML",json:"JSON",html:"HTML",css:"CSS",sql:"SQL",go:"Go",rust:"Rust",ruby:"Ruby",swift:"Swift",toml:"TOML",dockerfile:"Dockerfile"},m9=k2.define(x9).parser,g9=k2.define(P9).parser,p9=k2.define(f9).parser,h9=k2.define(R9).parser,c9=k2.define(w9).parser,l9=k2.define(v9).parser,r9=k2.define(b9).parser;function d8($){let j=String($||"").trim().toLowerCase();if(!j)return"text";return u9[j]||String($||"").trim()}function d9($){switch(String($||"").trim().toLowerCase()){case"js":case"javascript":return D9.parser;case"ts":case"typescript":return O9.parser;case"jsx":return C9.parser;case"tsx":return I9.parser;case"py":case"python":return y9.parser;case"json":return T9.parser;case"css":return H9.parser;case"html":return A9.parser;case"xml":return k9.parser;case"yaml":case"yml":return E9.parser;case"md":case"markdown":return M9.parser;case"sql":return S9.language.parser;case"go":return Y9.parser;case"sh":case"bash":case"shell":case"zsh":return m9;case"ps1":case"powershell":return g9;case"dockerfile":return p9;case"rb":case"ruby":return h9;case"rs":case"rust":return c9;case"swift":return l9;case"toml":return r9;default:return null}}function e4($,j){let q=d9(j);if(!q)return S2($);let B=[];try{let U=q.parse($);F9(U,N9,(X,J,W)=>{if(!W||X>=J)return;B.push({from:X,to:J,cls:W})})}catch{return S2($)}if(!B.length)return S2($);B.sort((U,X)=>U.from-X.from||U.to-X.to);let _=0,Q="";for(let U of B){if(U.from>_)Q+=S2($.slice(_,U.from));Q+=`<span class="${S2(U.cls)}">${S2($.slice(U.from,U.to))}</span>`,_=Math.max(_,U.to)}if(_<$.length)Q+=S2($.slice(_));return Q}var $5=/#(\w+)/g,i9=new Set(["strong","em","b","i","u","s","del","ins","sub","sup","mark","small","br","p","ul","ol","li","blockquote","ruby","rt","rp","span","input"]),n9=new Set(["a","abbr","blockquote","br","code","del","div","em","hr","h1","h2","h3","h4","h5","h6","i","img","input","ins","kbd","li","mark","ol","p","pre","ruby","rt","rp","s","small","span","strong","sub","sup","table","tbody","td","th","thead","tr","u","ul","math","semantics","mrow","mi","mn","mo","mtext","mspace","msup","msub","msubsup","mfrac","msqrt","mroot","mtable","mtr","mtd","annotation"]),s9=new Set(["class","title","role","aria-hidden","aria-label","aria-expanded","aria-live","data-mermaid","data-hashtag"]),i8={a:new Set(["href","target","rel"]),img:new Set(["src","alt","title"]),input:new Set(["type","checked","disabled"])},o9=new Set(["http:","https:","mailto:",""]);function a9($,j){let q=String($||"").toLowerCase(),B=String(j||"").toLowerCase();if(!B||B.startsWith("on"))return!1;if(B.startsWith("data-")||B.startsWith("aria-"))return!0;return(i8[q]||new Set).has(B)||s9.has(B)}function _3($){return String($||"").replace(/&/g,"&amp;").replace(/"/g,"&quot;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/'/g,"&#39;")}function E2($,j={}){if(!$)return null;let q=String($).trim();if(!q)return null;if(q.startsWith("#")||q.startsWith("/"))return q;if(q.startsWith("data:")){if(j.allowDataImage&&/^data:image\//i.test(q))return q;return null}if(q.startsWith("blob:"))return q;try{let B=new URL(q,typeof window<"u"?window.location.origin:"http://localhost");if(!o9.has(B.protocol))return null;return B.href}catch{return null}}function n8($,j={}){if(!$)return"";if(j?.sanitize===!1)return $;let q=new DOMParser().parseFromString($,"text/html"),B=[],_=q.createTreeWalker(q.body,NodeFilter.SHOW_ELEMENT),Q;while(Q=_.nextNode())B.push(Q);for(let U of B){let X=U.tagName.toLowerCase();if(!n9.has(X)){let W=U.parentNode;if(!W)continue;while(U.firstChild)W.insertBefore(U.firstChild,U);W.removeChild(U);continue}let J=i8[X]||new Set;for(let W of Array.from(U.attributes)){let V=W.name.toLowerCase(),G=W.value;if(V.startsWith("on")){U.removeAttribute(W.name);continue}if(a9(X,V)){if(V==="href"){let L=E2(G);if(!L)U.removeAttribute(W.name);else if(U.setAttribute(W.name,L),X==="a"){if(!U.getAttribute("rel"))U.setAttribute("rel","noopener noreferrer");if(/^https?:\/\//i.test(L))U.setAttribute("target","_blank")}}else if(V==="src"){let L=X==="img"&&typeof j.rewriteImageSrc==="function"?j.rewriteImageSrc(G):G,C=E2(L,{allowDataImage:X==="img"});if(!C)U.removeAttribute(W.name);else U.setAttribute(W.name,C)}continue}U.removeAttribute(W.name)}}return q.body.innerHTML}function s8($){if(!$)return $;let j=$.replace(/</g,"&lt;").replace(/>/g,"&gt;");return new DOMParser().parseFromString(j,"text/html").documentElement.textContent}function _4($,j=2){if(!$)return $;let q=$;for(let B=0;B<j;B+=1){let _=s8(q);if(_===q)break;q=_}return q}function t9($){if(!$)return{text:"",frontmatter:null};let j=$.replace(/^\uFEFF/,"").replace(/\r\n/g,`
`).replace(/\r/g,`
`);if(!j.startsWith(`---
`))return{text:j,frontmatter:null};let q=j.split(`
`),B=-1;for(let U=1;U<q.length;U+=1)if(/^(---|\.\.\.)\s*$/.test(q[U])){B=U;break}if(B<=0)return{text:j,frontmatter:null};let _=q.slice(1,B).join(`
`);return{text:q.slice(B+1).join(`
`).replace(/^\n+/,""),frontmatter:_}}function e9($){let{text:j,frontmatter:q}=t9($);if(q===null)return j;return["<!--frontmatter-block-start-->","```yaml",q,"```","<!--frontmatter-block-end-->",j].filter(Boolean).join(`

`)}function $7($){if(!$)return{text:"",blocks:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=[],_=[],Q=!1,U=[];for(let X of q){if(!Q&&X.trim().match(/^```mermaid\s*$/i)){Q=!0,U=[];continue}if(Q&&X.trim().match(/^```\s*$/)){let J=B.length;B.push(U.join(`
`)),_.push(`@@MERMAID_BLOCK_${J}@@`),Q=!1,U=[];continue}if(Q)U.push(X);else _.push(X)}if(Q)_.push("```mermaid"),_.push(...U);return{text:_.join(`
`),blocks:B}}function j7($){if(!$)return $;return _4($,5)}function q7($){let j=new TextEncoder().encode(String($||"")),q="";for(let B of j)q+=String.fromCharCode(B);return btoa(q)}function B7($){let j=atob(String($||"")),q=new Uint8Array(j.length);for(let B=0;B<j.length;B+=1)q[B]=j.charCodeAt(B);return new TextDecoder().decode(q)}function _7($,j){if(!$||!j||j.length===0)return $;return $.replace(/@@MERMAID_BLOCK_(\d+)@@/g,(q,B)=>{let _=Number(B),Q=j[_]??"",U=j7(Q);return`<div class="mermaid-container" data-mermaid="${q7(U)}"><div class="mermaid-loading">Loading diagram...</div></div>`})}function o8($){if(!$)return $;return $.replace(/<code>([\s\S]*?)<\/code>/gi,(j,q)=>{if(q.includes(`
`))return`
\`\`\`
${q}
\`\`\`
`;return`\`${q}\``})}function Q7($){if(!$)return $;return $.replace(/<pre><code(?:\s+class="language-([A-Za-z0-9_+-]+)")?>([\s\S]*?)<\/code><\/pre>/g,(q,B,_)=>{let Q=String(B||"").trim().toLowerCase(),U=_4(_,2),X=Q||"plaintext",J=e4(U,Q);return`<pre><code class="hljs language-${_3(X)}">${J}</code></pre>`}).replace(/<!--frontmatter-block-start-->\s*<pre>/g,'<pre class="frontmatter-block">').replace(/<\/pre>\s*<!--frontmatter-block-end-->/g,"</pre>")}var U7={span:new Set(["title","class","lang","dir"]),input:new Set(["type","checked","disabled"])};function X7($,j){let q=U7[$];if(!q||!j)return"";let B=[],_=/([a-zA-Z_:][\w:.-]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'`=<>]+)))?/g,Q;while(Q=_.exec(j)){let U=(Q[1]||"").toLowerCase();if(!U||U.startsWith("on")||!q.has(U))continue;let X=Q[2]??Q[3]??Q[4]??"";B.push(` ${U}="${_3(X)}"`)}return B.join("")}function a8($){if(!$)return $;return $.replace(/&lt;((?:[^"'<>]|"[^"]*"|'[^']*')*?)(?:&gt;|>)/g,(j,q)=>{let B=q.trim(),_=B.startsWith("/"),Q=_?B.slice(1).trim():B,X=Q.endsWith("/")?Q.slice(0,-1).trim():Q,[J=""]=X.split(/\s+/,1),W=J.toLowerCase();if(!W||!i9.has(W))return j;if(W==="br")return _?"":"<br>";if(_)return`</${W}>`;let V=X.slice(J.length).trim(),G=X7(W,V);return`<${W}${G}>`})}function t8($){if(!$)return $;let j=(q)=>q.replace(/&amp;lt;/g,"&lt;").replace(/&amp;gt;/g,"&gt;").replace(/&amp;quot;/g,"&quot;").replace(/&amp;#39;/g,"&#39;").replace(/&amp;amp;/g,"&amp;");return $.replace(/<pre><code>([\s\S]*?)<\/code><\/pre>/g,(q,B)=>`<pre><code>${j(B)}</code></pre>`).replace(/<code>([\s\S]*?)<\/code>/g,(q,B)=>`<code>${j(B)}</code>`)}function e8($){if(!$)return $;let j=new DOMParser().parseFromString($,"text/html"),q=j.createTreeWalker(j.body,NodeFilter.SHOW_TEXT),B=(Q)=>Q.replace(/&lt;/g,"<").replace(/&gt;/g,">").replace(/&quot;/g,'"').replace(/&#39;/g,"'").replace(/&amp;/g,"&"),_;while(_=q.nextNode()){if(!_.nodeValue)continue;let Q=B(_.nodeValue);if(Q!==_.nodeValue)_.nodeValue=Q}return j.body.innerHTML}function W7($){if(!window.katex)return $;let j=(U)=>s8(U).replace(/&gt;/g,">").replace(/&lt;/g,"<").replace(/&amp;/g,"&").replace(/<br\s*\/?\s*>/gi,`
`),q=(U)=>{let X=[],J=U.replace(/<pre\b[^>]*>\s*<code\b[^>]*>[\s\S]*?<\/code>\s*<\/pre>/gi,(W)=>{let V=X.length;return X.push(W),`@@CODE_BLOCK_${V}@@`});return J=J.replace(/<code\b[^>]*>[\s\S]*?<\/code>/gi,(W)=>{let V=X.length;return X.push(W),`@@CODE_INLINE_${V}@@`}),{html:J,blocks:X}},B=(U,X)=>{if(!X.length)return U;return U.replace(/@@CODE_(?:BLOCK|INLINE)_(\d+)@@/g,(J,W)=>{let V=Number(W);return X[V]??""})},_=q($),Q=_.html;return Q=Q.replace(/(^|\n|<br\s*\/?\s*>|<p>|<\/p>)\s*\$\$([\s\S]+?)\$\$\s*(?=\n|<br\s*\/?\s*>|<\/p>|$)/gi,(U,X,J)=>{try{let W=katex.renderToString(j(J.trim()),{displayMode:!0,throwOnError:!1});return`${X}${W}`}catch(W){return`<span class="math-error" title="${_3(W.message)}">${U}</span>`}}),B(Q,_.blocks)}function J7($){if(!$)return $;let j=new DOMParser().parseFromString($,"text/html"),q=j.createTreeWalker(j.body,NodeFilter.SHOW_TEXT),B=[],_;while(_=q.nextNode())B.push(_);for(let Q of B){let U=Q.nodeValue;if(!U)continue;if($5.lastIndex=0,!$5.test(U))continue;$5.lastIndex=0;let X=Q.parentElement;if(X&&(X.closest("a")||X.closest("code")||X.closest("pre")))continue;let J=U.split($5);if(J.length<=1)continue;let W=j.createDocumentFragment();J.forEach((V,G)=>{if(G%2===1){let L=j.createElement("a");L.setAttribute("href","#"),L.className="hashtag",L.setAttribute("data-hashtag",V),L.textContent=`#${V}`,W.appendChild(L)}else W.appendChild(j.createTextNode(V))}),Q.parentNode?.replaceChild(W,Q)}return j.body.innerHTML}function K7($){if(!$)return $;let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=[],_=!1;for(let Q of q){if(!_&&Q.trim().match(/^```(?:math|katex|latex)\s*$/i)){_=!0,B.push("$$");continue}if(_&&Q.trim().match(/^```\s*$/)){_=!1,B.push("$$");continue}B.push(Q)}return B.join(`
`)}function z7($){let j=e9($||""),q=K7(j),{text:B,blocks:_}=$7(q),Q=_4(B,2),X=o8(Q).replace(/</g,"&lt;");return{safeHtml:a8(X),mermaidBlocks:_}}function B2($,j,q={}){if(!$)return"";let{safeHtml:B,mermaidBlocks:_}=z7($),Q=window.marked?marked.parse(B,{headerIds:!1,mangle:!1}):B.replace(/\n/g,"<br>");return Q=t8(Q),Q=e8(Q),Q=Q7(Q),Q=W7(Q),Q=J7(Q),Q=_7(Q,_),Q=n8(Q,q),Q}function Q3($){if(!$)return"";let j=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`),q=_4(j,2),_=o8(q).replace(/</g,"&lt;").replace(/>/g,"&gt;"),Q=a8(_),U=window.marked?marked.parse(Q):Q.replace(/\n/g,"<br>");return U=t8(U),U=e8(U),U=n8(U),U}function G7($,j=6){return $.replace(/<polyline\b([^>]*)\bpoints="([^"]+)"([^>]*)\/?\s*>/g,(q,B,_,Q)=>{let U=_.trim().split(/\s+/).map((J)=>{let[W,V]=J.split(",").map(Number);return{x:W,y:V}});if(U.length<3)return`<polyline${B}points="${_}"${Q}/>`;let X=[`M ${U[0].x},${U[0].y}`];for(let J=1;J<U.length-1;J++){let W=U[J-1],V=U[J],G=U[J+1],L=V.x-W.x,C=V.y-W.y,b=G.x-V.x,c=G.y-V.y,l=Math.sqrt(L*L+C*C),w=Math.sqrt(b*b+c*c),S=Math.min(j,l/2,w/2);if(S<0.5){X.push(`L ${V.x},${V.y}`);continue}let m=V.x-L/l*S,O=V.y-C/l*S,d=V.x+b/w*S,g=V.y+c/w*S,k=L*c-C*b>0?1:0;X.push(`L ${m},${O}`),X.push(`A ${S},${S} 0 0 ${k} ${d},${g}`)}return X.push(`L ${U[U.length-1].x},${U[U.length-1].y}`),`<path${B}d="${X.join(" ")}"${Q}/>`})}async function j5($){if(!window.beautifulMermaid)return;let{renderMermaid:j,THEMES:q}=window.beautifulMermaid,_=M8()==="dark"?q["tokyo-night"]:q["github-light"],Q=$.querySelectorAll(".mermaid-container[data-mermaid]");for(let U of Q)try{let X=U.dataset.mermaid,J=B7(X||""),W=_4(J,2),V=await j(W,{..._,transparent:!0});V=G7(V),U.innerHTML=V,U.removeAttribute("data-mermaid")}catch(X){console.error("Mermaid render error:",X);let J=document.createElement("pre");J.className="mermaid-error",J.textContent=`Diagram error: ${X.message}`,U.innerHTML="",U.appendChild(J),U.removeAttribute("data-mermaid")}}function $6($){let j=new Date($);if(Number.isNaN(j.getTime()))return $;let B=new Date-j,_=B/1000,Q=86400000;if(B<Q){if(_<60)return"just now";if(_<3600)return`${Math.floor(_/60)}m`;return`${Math.floor(_/3600)}h`}if(B<5*Q){let J=j.toLocaleDateString(void 0,{weekday:"short"}),W=j.toLocaleTimeString(void 0,{hour:"2-digit",minute:"2-digit"});return`${J} ${W}`}let U=j.toLocaleDateString(void 0,{month:"short",day:"numeric"}),X=j.toLocaleTimeString(void 0,{hour:"2-digit",minute:"2-digit"});return`${U} ${X}`}function Q4($){if(!Number.isFinite($))return"0";return Math.round($).toLocaleString()}function C0($){if($<1024)return $+" B";if($<1048576)return($/1024).toFixed(1)+" KB";return($/1048576).toFixed(1)+" MB"}function p2($){let j=new Date($);if(Number.isNaN(j.getTime()))return $;return j.toLocaleString()}function U4($){if($==null)return"";if(typeof $==="string")return $.trim();if(typeof $==="number")return String($);if(typeof $==="boolean")return $?"yes":"no";if(Array.isArray($))return $.map((j)=>U4(j)).filter(Boolean).join(", ");if(typeof $==="object")return Object.entries($).filter(([j])=>!j.startsWith("__")).map(([j,q])=>`${j}: ${U4(q)}`).filter((j)=>!j.endsWith(": ")).join(", ");return String($).trim()}function j6($){if(typeof $!=="object"||$==null||Array.isArray($))return[];return Object.entries($).filter(([j])=>!j.startsWith("__")).map(([j,q])=>({key:j,value:U4(q)})).filter((j)=>j.value)}function Z7($){if(!$||typeof $!=="object")return!1;let j=$;return j.type==="adaptive_card_submission"&&typeof j.card_id==="string"&&typeof j.source_post_id==="number"&&typeof j.submitted_at==="string"}function U3($){if(!Array.isArray($))return[];return $.filter(Z7)}function q5($){let j=String($.title||$.card_id||"card").trim()||"card",q=$.data;if(q==null)return`Card submission: ${j}`;if(typeof q==="string"||typeof q==="number"||typeof q==="boolean"){let B=U4(q);return B?`Card submission: ${j} — ${B}`:`Card submission: ${j}`}if(typeof q==="object"){let _=j6(q).map(({key:Q,value:U})=>`${Q}: ${U}`);return _.length>0?`Card submission: ${j} — ${_.join(", ")}`:`Card submission: ${j}`}return`Card submission: ${j}`}function q6($){let j=String($.title||$.card_id||"Card submission").trim()||"Card submission",q=j6($.data),B=q.length>0?q.slice(0,2).map(({key:Q,value:U})=>`${Q}: ${U}`).join(", "):U4($.data)||null,_=q.length;return{title:j,summary:B,fields:q,fieldCount:_,submittedAt:$.submitted_at}}function p1($){return typeof $==="string"?$.trim():""}function B6($){return $.map((j)=>String(j||"").trim()).filter(Boolean).join(`

`).replace(/\n{3,}/g,`

`).trim()}function V7($,j){let q=[],B=[],_=[];if($.forEach((Q,U)=>{if(!Q||typeof Q!=="object")return;let X=p1(Q.type);if(X==="text"){let J=p1(Q.text)||p1(Q.content);if(J)q.push(J);return}if(X==="resource_link"){let J=p1(Q.uri),W=p1(Q.title)||p1(Q.name)||J;if(J&&W)q.push(W===J?J:`[${W}](${J})`);return}if(X==="resource"){let J=p1(Q.title)||p1(Q.name)||p1(Q.uri)||"Embedded resource",W=p1(Q.text);if(W)q.push(`### ${J}

\`\`\`
${W}
\`\`\``);else q.push(`### ${J}`);return}if(X==="generated_widget"){let J=p1(Q.title)||p1(Q.name)||"Generated widget",W=p1(Q.description)||p1(Q.subtitle);q.push(B6([`### ${J}`,W]));return}if(X==="adaptive_card"&&p1(Q.fallback_text)){q.push(p1(Q.fallback_text));return}if(X==="adaptive_card_submission"){let J=q5(Q);if(p1(J))q.push(p1(J));return}if(X==="file"){let J=p1(Q.name)||p1(Q.filename)||p1(Q.title)||`attachment:${j[U]??U+1}`;B.push(`- ${J}`);return}if(X==="image"||!X){let J=p1(Q.name)||p1(Q.filename)||p1(Q.title)||`attachment:${j[U]??U+1}`;_.push(`- ${J}`)}}),_.length>0)q.push(`Images:
${_.join(`
`)}`);if(B.length>0)q.push(`Attachments:
${B.join(`
`)}`);return B6(q)}function _6($){let j=$?.data||{},q=typeof j.content==="string"?j.content.replace(/\r\n/g,`
`).replace(/\r/g,`
`).trimEnd():"";if(q.trim())return q;let B=Array.isArray(j.content_blocks)?j.content_blocks:[],_=Array.isArray(j.media_ids)?j.media_ids:[];return V7(B,_)}var Q6="PiClaw";function X3($,j,q=!1){let B=$||"PiClaw",_=B.charAt(0).toUpperCase(),Q=["#FF6B6B","#4ECDC4","#45B7D1","#FFA07A","#98D8C8","#F7DC6F","#BB8FCE","#85C1E2","#F8B195","#6C5CE7","#00B894","#FDCB6E","#E17055","#74B9FF","#A29BFE","#FD79A8","#00CEC9","#FFEAA7","#DFE6E9","#FF7675","#55EFC4","#81ECEC","#FAB1A0","#74B9FF","#A29BFE","#FD79A8"],U=_.charCodeAt(0)%Q.length,X=Q[U],J=B.trim().toLowerCase(),W=typeof j==="string"?j.trim():"",V=W?W:null,G=q||J==="PiClaw".toLowerCase()||J==="pi";return{letter:_,color:X,image:V||(G?"/static/icon-192.png":null)}}function U6($,j){if(!$)return"PiClaw";let q=j[$]?.name||$;return q?q.charAt(0).toUpperCase()+q.slice(1):"PiClaw"}function X6($,j){if(!$)return null;let q=j[$]||{};return q.avatar_url||q.avatarUrl||q.avatar||null}function W6($){if(!$)return null;if(typeof document<"u"){let Q=document.documentElement,U=Q?.dataset?.colorTheme||"",X=Q?.dataset?.tint||"",J=getComputedStyle(Q).getPropertyValue("--accent-color")?.trim();if(J&&(X||U&&U!=="default"))return J}let j=["#4ECDC4","#FF6B6B","#45B7D1","#BB8FCE","#FDCB6E","#00B894","#74B9FF","#FD79A8","#81ECEC","#FFA07A"],q=String($),B=0;for(let Q=0;Q<q.length;Q+=1)B=(B*31+q.charCodeAt(Q))%2147483647;let _=Math.abs(B)%j.length;return j[_]}var L7=new Set(["application/json","application/xml","text/csv","text/html","text/markdown","text/plain","text/xml"]),N7=new Set(["text/markdown"]),F7=new Set(["application/msword","application/rtf","application/vnd.ms-excel","application/vnd.ms-powerpoint","application/vnd.oasis.opendocument.presentation","application/vnd.oasis.opendocument.spreadsheet","application/vnd.oasis.opendocument.text","application/vnd.openxmlformats-officedocument.presentationml.presentation","application/vnd.openxmlformats-officedocument.spreadsheetml.sheet","application/vnd.openxmlformats-officedocument.wordprocessingml.document"]),H7=new Set(["application/vnd.jgraph.mxfile"]);function G2($){return typeof $==="string"?$.trim().toLowerCase():""}function Y7($){let j=G2($);return!!j&&(j.endsWith(".drawio")||j.endsWith(".drawio.xml")||j.endsWith(".drawio.svg")||j.endsWith(".drawio.png"))}function A7($){let j=G2($);return!!j&&j.endsWith(".pdf")}function D7($){let j=G2($);return!!j&&(j.endsWith(".docx")||j.endsWith(".doc")||j.endsWith(".odt")||j.endsWith(".rtf")||j.endsWith(".xlsx")||j.endsWith(".xls")||j.endsWith(".ods")||j.endsWith(".pptx")||j.endsWith(".ppt")||j.endsWith(".odp"))}var C7=new Set(["application/zip","application/x-zip-compressed"]);function I7($){let j=G2($);return!!j&&j.endsWith(".zip")}function O7($){let j=G2($);return!!j&&(j.endsWith(".html")||j.endsWith(".htm"))}function T7($){let j=G2($);if(!j)return!1;return j.endsWith(".sh")||j.endsWith(".bash")||j.endsWith(".zsh")||j.endsWith(".sb")}function X4($,j){let q=G2($);if(Y7(j)||H7.has(q))return"drawio";if(A7(j)||q==="application/pdf")return"pdf";if(D7(j)||F7.has(q))return"office";if(I7(j)||C7.has(q))return"archive";if(O7(j)||q==="text/html")return"html";if(T7(j))return"text";if(!q)return"unsupported";if(q.startsWith("video/"))return"video";if(q.startsWith("image/"))return"image";if(L7.has(q)||q.startsWith("text/"))return"text";return"unsupported"}function J6($){let j=G2($);return N7.has(j)}function K6($){switch($){case"image":return"Image preview";case"video":return"Video player";case"pdf":return"PDF preview";case"office":return"Office viewer";case"drawio":return"Draw.io preview (read-only)";case"html":return"HTML preview";case"text":return"Text preview";case"archive":return"ZIP archive preview";default:return"Preview unavailable"}}function z6($,j,q){try{return $.setAttribute(j,q),!0}catch(B){return!1}}function G6($,j){try{return $[j]=!0,!0}catch(q){return!1}}function Z6($){$.classList.add("adaptive-card-readonly");for(let j of Array.from($.querySelectorAll("input, textarea, select, button"))){let q=j;if(z6(q,"aria-disabled","true"),z6(q,"tabindex","-1"),"disabled"in q)G6(q,"disabled");if("readOnly"in q)G6(q,"readOnly")}}function M7($){let q=String($||"").trim().match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);if(!q)return null;let B=q[1].length===3?q[1].split("").map((_)=>`${_}${_}`).join(""):q[1];return{r:parseInt(B.slice(0,2),16),g:parseInt(B.slice(2,4),16),b:parseInt(B.slice(4,6),16)}}function y7($){let q=String($||"").trim().match(/^rgba?\((\d+)[,\s]+(\d+)[,\s]+(\d+)/i);if(!q)return null;let B=Number(q[1]),_=Number(q[2]),Q=Number(q[3]);if(![B,_,Q].every((U)=>Number.isFinite(U)))return null;return{r:B,g:_,b:Q}}function V6($){return M7($)||y7($)}function B5($){let j=(Q)=>{let U=Q/255;return U<=0.03928?U/12.92:((U+0.055)/1.055)**2.4},q=j($.r),B=j($.g),_=j($.b);return 0.2126*q+0.7152*B+0.0722*_}function S7($,j){let q=Math.max(B5($),B5(j)),B=Math.min(B5($),B5(j));return(q+0.05)/(B+0.05)}function k7($,j,q="#ffffff"){let B=V6($);if(!B)return q;let _=q,Q=-1;for(let U of j){let X=V6(U);if(!X)continue;let J=S7(B,X);if(J>Q)_=U,Q=J}return _}function W3(){let $=getComputedStyle(document.documentElement),j=(b,c)=>{for(let l of b){let w=$.getPropertyValue(l).trim();if(w)return w}return c},q=j(["--text-primary","--color-text"],"#0f1419"),B=j(["--text-secondary","--color-text-muted"],"#536471"),_=j(["--bg-primary","--color-bg-primary"],"#ffffff"),Q=j(["--bg-secondary","--color-bg-secondary"],"#f7f9fa"),U=j(["--bg-hover","--bg-tertiary","--color-bg-tertiary"],"#e8ebed"),X=j(["--accent-color","--color-accent"],"#1d9bf0"),J=j(["--success-color","--color-success"],"#00ba7c"),W=j(["--warning-color","--color-warning","--accent-color"],"#f0b429"),V=j(["--danger-color","--color-error"],"#f4212e"),G=j(["--border-color","--color-border"],"#eff3f4"),L=j(["--font-family"],"system-ui, sans-serif"),C=k7(X,[q,_],q);return{fg:q,fgMuted:B,bgPrimary:_,bg:Q,bgEmphasis:U,accent:X,good:J,warning:W,attention:V,border:G,fontFamily:L,buttonTextColor:C}}function L6(){let{fg:$,fgMuted:j,bg:q,bgEmphasis:B,accent:_,good:Q,warning:U,attention:X,border:J,fontFamily:W}=W3();return{fontFamily:W,containerStyles:{default:{backgroundColor:q,foregroundColors:{default:{default:$,subtle:j},accent:{default:_,subtle:_},good:{default:Q,subtle:Q},warning:{default:U,subtle:U},attention:{default:X,subtle:X}}},emphasis:{backgroundColor:B,foregroundColors:{default:{default:$,subtle:j},accent:{default:_,subtle:_},good:{default:Q,subtle:Q},warning:{default:U,subtle:U},attention:{default:X,subtle:X}}}},actions:{actionsOrientation:"horizontal",actionAlignment:"left",buttonSpacing:8,maxActions:5,showCard:{actionMode:"inline"},spacing:"default"},adaptiveCard:{allowCustomStyle:!1},spacing:{small:4,default:8,medium:12,large:16,extraLarge:24,padding:12},separator:{lineThickness:1,lineColor:J},fontSizes:{small:12,default:14,medium:16,large:18,extraLarge:22},fontWeights:{lighter:300,default:400,bolder:600},imageSizes:{small:40,medium:80,large:120},textBlock:{headingLevel:2}}}var E7=new Set(["1.0","1.1","1.2","1.3","1.4","1.5","1.6"]),N6=!1,_5=null,F6=!1;function J3($){$.querySelector(".adaptive-card-notice")?.remove()}function f7($,j,q="error"){J3($);let B=document.createElement("div");B.className=`adaptive-card-notice adaptive-card-notice-${q}`,B.textContent=j,$.appendChild(B)}function P7($,j=(q)=>B2(q,null)){let q=typeof $==="string"?$:String($??"");if(!q.trim())return{outputHtml:"",didProcess:!1};return{outputHtml:j(q),didProcess:!0}}function R7($=(j)=>B2(j,null)){return(j,q)=>{try{let B=P7(j,$);q.outputHtml=B.outputHtml,q.didProcess=B.didProcess}catch(B){console.error("[adaptive-card] Failed to process markdown:",B),q.outputHtml=String(j??""),q.didProcess=!1}}}function w7($){if(F6||!$?.AdaptiveCard)return;$.AdaptiveCard.onProcessMarkdown=R7(),F6=!0}async function x7(){if(N6)return;if(_5)return _5;return _5=new Promise(($,j)=>{let q=document.createElement("script");q.src="/static/js/vendor/adaptivecards.min.js",q.onload=()=>{N6=!0,$()},q.onerror=()=>j(Error("Failed to load adaptivecards SDK")),document.head.appendChild(q)}),_5}function v7(){return globalThis.AdaptiveCards}function b7($){if(!$||typeof $!=="object")return!1;let j=$;return j.type==="adaptive_card"&&typeof j.card_id==="string"&&typeof j.schema_version==="string"&&typeof j.payload==="object"&&j.payload!==null}function u7($){return E7.has($)}function z3($){if(!Array.isArray($))return[];return $.filter(b7)}function m7($){let j=(typeof $?.getJsonTypeName==="function"?$.getJsonTypeName():"")||$?.constructor?.name||"Unknown",q=(typeof $?.title==="string"?$.title:"")||"",B=(typeof $?.url==="string"?$.url:"")||void 0,_=$?.data??void 0;return{type:j,title:q,data:_,url:B,raw:$}}function K3($){if($==null)return"";if(typeof $==="string")return $.trim();if(typeof $==="number")return String($);if(typeof $==="boolean")return $?"yes":"no";if(Array.isArray($))return $.map((j)=>K3(j)).filter(Boolean).join(", ");if(typeof $==="object")return Object.entries($).map(([q,B])=>`${q}: ${K3(B)}`).filter((q)=>!q.endsWith(": ")).join(", ");return String($).trim()}function g7($,j,q){if(j==null)return j;if($==="Input.Toggle"){if(typeof j==="boolean"){if(j)return q?.valueOn??"true";return q?.valueOff??"false"}return typeof j==="string"?j:String(j)}if($==="Input.ChoiceSet"){if(Array.isArray(j))return j.join(",");return typeof j==="string"?j:String(j)}if(Array.isArray(j))return j.join(", ");if(typeof j==="object")return K3(j);return typeof j==="string"?j:String(j)}function p7($,j){if(!$||typeof $!=="object")return $;if(!j||typeof j!=="object"||Array.isArray(j))return $;let q=j,B=(_)=>{if(Array.isArray(_))return _.map((X)=>B(X));if(!_||typeof _!=="object")return _;let U={..._};if(typeof U.id==="string"&&U.id in q&&String(U.type||"").startsWith("Input."))U.value=g7(U.type,q[U.id],U);for(let[X,J]of Object.entries(U))if(Array.isArray(J)||J&&typeof J==="object")U[X]=B(J);return U};return B($)}function h7($){if(typeof $!=="string"||!$.trim())return"";let j=new Date($);if(Number.isNaN(j.getTime()))return"";return new Intl.DateTimeFormat(void 0,{month:"short",day:"numeric",hour:"numeric",minute:"2-digit"}).format(j)}function c7($){if($.state==="active")return null;let j=$.state==="completed"?"Submitted":$.state==="cancelled"?"Cancelled":"Failed",q=$.last_submission&&typeof $.last_submission==="object"?$.last_submission:null,B=q&&typeof q.title==="string"?q.title.trim():"",_=h7($.completed_at||q?.submitted_at),Q=[B||null,_||null].filter(Boolean).join(" · ")||null;return{label:j,detail:Q}}async function H6($,j,q){if(!u7(j.schema_version))return console.warn(`[adaptive-card] Unsupported schema version ${j.schema_version} for card ${j.card_id}`),!1;try{await x7()}catch(B){return console.error("[adaptive-card] Failed to load SDK:",B),!1}try{let B=v7();w7(B);let _=new B.AdaptiveCard,Q=W3();_.hostConfig=new B.HostConfig(L6());let U=j.last_submission&&typeof j.last_submission==="object"?j.last_submission.data:void 0,X=j.state==="active"?j.payload:p7(j.payload,U);_.parse(X),_.onExecuteAction=(V)=>{let G=m7(V);if(q?.onAction)J3($),$.classList.add("adaptive-card-busy"),Promise.resolve(q.onAction(G)).catch((L)=>{console.error("[adaptive-card] Action failed:",L);let C=L instanceof Error?L.message:String(L||"Action failed.");f7($,C||"Action failed.","error")}).finally(()=>{$.classList.remove("adaptive-card-busy")});else console.log("[adaptive-card] Action executed (not wired yet):",G)};let J=_.render();if(!J)return console.warn(`[adaptive-card] Card ${j.card_id} rendered to null`),!1;$.classList.add("adaptive-card-container"),$.style.setProperty("--adaptive-card-button-text-color",Q.buttonTextColor);let W=c7(j);if(W){$.classList.add("adaptive-card-finished");let V=document.createElement("div");V.className=`adaptive-card-status adaptive-card-status-${j.state}`;let G=document.createElement("span");if(G.className="adaptive-card-status-label",G.textContent=W.label,V.appendChild(G),W.detail){let L=document.createElement("span");L.className="adaptive-card-status-detail",L.textContent=W.detail,V.appendChild(L)}$.appendChild(V)}if(J3($),$.appendChild(J),W)Z6(J);return!0}catch(B){return console.error(`[adaptive-card] Failed to render card ${j.card_id}:`,B),!1}}function l7($){let j=$?.artifact||{},q=j.kind||$?.kind||null;if(q!=="html"&&q!=="svg"&&q!=="session_tree")return null;if(q==="html"){let _=typeof j.html==="string"?j.html:typeof $?.html==="string"?$.html:"";return _?{kind:q,html:_}:null}if(q==="svg"){let _=typeof j.svg==="string"?j.svg:typeof $?.svg==="string"?$.svg:"";return _?{kind:q,svg:_}:null}let B=j.tree&&typeof j.tree==="object"?j.tree:$?.tree&&typeof $.tree==="object"?$.tree:null;return{kind:q,tree:B}}function Q5($){return typeof $==="number"&&Number.isFinite($)?$:null}function N0($){return typeof $==="string"&&$.trim()?$.trim():null}function r7($,j=!1){let B=(Array.isArray($)?$:j?["interactive"]:[]).filter((_)=>typeof _==="string").map((_)=>_.trim().toLowerCase()).filter(Boolean);return Array.from(new Set(B))}var A6="__PICLAW_WIDGET_HOST__:";function Y6($){return JSON.stringify($).replace(/</g,"\\u003c").replace(/>/g,"\\u003e").replace(/&/g,"\\u0026").replace(/\u2028/g,"\\u2028").replace(/\u2029/g,"\\u2029")}function G3($,j){if(!$||$.type!=="generated_widget")return null;let q=l7($);if(!q)return null;return{title:$.title||$.name||"Generated widget",subtitle:typeof $.subtitle==="string"?$.subtitle:"",description:$.description||$.subtitle||"",originPostId:Number.isFinite(j?.id)?j.id:null,originChatJid:typeof j?.chat_jid==="string"?j.chat_jid:null,widgetId:$.widget_id||$.id||null,artifact:q,capabilities:r7($.capabilities,$.interactive===!0),source:"timeline",status:"final"}}function D6($){return G3($,null)!==null}function Z3($){let j=N0($?.toolCallId)||N0($?.tool_call_id);if(j)return j;let q=N0($?.widgetId)||N0($?.widget_id);if(q)return q;let B=Q5($?.originPostId)??Q5($?.origin_post_id);if(B!==null)return`post:${B}`;return null}function C6($){let q=($?.artifact||{}).kind||$?.kind||null,_=(Array.isArray($?.capabilities)?$.capabilities:[]).some((Q)=>typeof Q==="string"&&Q.trim().toLowerCase()==="interactive");return q==="html"&&($?.source==="live"||_)}function I6($){return C6($)?"allow-downloads allow-scripts allow-same-origin":"allow-downloads"}function U5($){return{title:N0($?.title)||"Generated widget",widgetId:N0($?.widgetId)||N0($?.widget_id),toolCallId:N0($?.toolCallId)||N0($?.tool_call_id),turnId:N0($?.turnId)||N0($?.turn_id),capabilities:Array.isArray($?.capabilities)?$.capabilities:[],source:$?.source==="live"?"live":"timeline",status:N0($?.status)||"final"}}function X5($){return{...U5($),subtitle:N0($?.subtitle)||"",description:N0($?.description)||"",error:N0($?.error)||null,width:Q5($?.width),height:Q5($?.height),runtimeState:$?.runtimeState&&typeof $.runtimeState==="object"?$.runtimeState:null}}function W5($){return`${A6}${JSON.stringify(X5($))}`}function O6($){let j=N0($?.status);if(j==="loading"||j==="streaming")return"Widget is loading…";if(j==="error")return N0($?.error)||"Widget failed to load.";if(($?.artifact?.kind||$?.kind)==="session_tree")return"Session tree widget is unavailable.";return"Widget artifact is missing or unsupported."}function d7($){let j=U5($);return`<script>
(function () {
  const meta = ${Y6(j)};
  function post(kind, payload) {
    try {
      window.parent.postMessage({
        __piclawGeneratedWidget: true,
        kind,
        widgetId: meta.widgetId || null,
        toolCallId: meta.toolCallId || null,
        turnId: meta.turnId || null,
        payload: payload || {}
      }, '*');
    } catch {
      /* expected: parent bridge may be unavailable while the iframe is unloading. */
    }
  }

  const windowNamePrefix = ${Y6(A6)};
  let lastWindowName = null;
  let pendingHostEnvelope = null;
  let pendingHostEnvelopeFrame = 0;
  let lastDispatchedEnvelopeKey = null;

  function getEnvelopeKey(data) {
    try {
      return JSON.stringify([
        data?.type || null,
        data?.widgetId || null,
        data?.toolCallId || null,
        data?.turnId || null,
        data?.payload || null,
      ]);
    } catch {
      return null;
    }
  }

  function flushHostEnvelope() {
    pendingHostEnvelopeFrame = 0;
    const data = pendingHostEnvelope;
    pendingHostEnvelope = null;
    if (!data) return;

    window.piclawWidget.lastHostMessage = data;
    const nextPayload = data.payload || null;
    if (data.type === 'widget.init') {
      const previous = window.piclawWidget.hostState && typeof window.piclawWidget.hostState === 'object'
        ? window.piclawWidget.hostState
        : null;
      if (nextPayload && typeof nextPayload === 'object') {
        window.piclawWidget.hostState = {
          ...(previous || {}),
          ...nextPayload,
          ...(Object.prototype.hasOwnProperty.call(nextPayload, 'runtimeState')
            ? {}
            : { runtimeState: previous?.runtimeState ?? null }),
        };
      } else {
        window.piclawWidget.hostState = previous || null;
      }
    } else if (data.type === 'widget.update' || data.type === 'widget.complete' || data.type === 'widget.error') {
      window.piclawWidget.hostState = nextPayload;
    }

    const effectivePayload = window.piclawWidget.hostState ?? nextPayload ?? null;
    const detail = (effectivePayload === data.payload)
      ? data
      : { ...data, payload: effectivePayload };
    const envelopeKey = getEnvelopeKey(detail);
    if (envelopeKey && envelopeKey === lastDispatchedEnvelopeKey) return;
    lastDispatchedEnvelopeKey = envelopeKey;
    window.dispatchEvent(new CustomEvent('piclaw:widget-message', { detail }));
  }

  function scheduleHostEnvelope(data) {
    if (!data) return;
    pendingHostEnvelope = data;
    if (pendingHostEnvelopeFrame) return;
    const schedule = typeof requestAnimationFrame === 'function'
      ? requestAnimationFrame
      : (cb) => setTimeout(cb, 0);
    pendingHostEnvelopeFrame = schedule(flushHostEnvelope);
  }

  function readWindowNameState() {
    try {
      const raw = window.name || '';
      if (!raw || raw === lastWindowName || !raw.startsWith(windowNamePrefix)) return;
      lastWindowName = raw;
      const payload = JSON.parse(raw.slice(windowNamePrefix.length));
      scheduleHostEnvelope({
        __piclawGeneratedWidgetHost: true,
        type: 'widget.update',
        widgetId: meta.widgetId || null,
        toolCallId: meta.toolCallId || null,
        turnId: meta.turnId || null,
        payload,
      });
    } catch {
      /* expected: host window.name payload can be absent or mid-update while polling. */
    }
  }

  window.piclawWidget = {
    meta,
    lastHostMessage: null,
    hostState: null,
    ready(payload) { post('widget.ready', payload); },
    close(payload) { post('widget.close', payload); },
    requestRefresh(payload) { post('widget.request_refresh', payload); },
    submit(payload) { post('widget.submit', payload); },
  };

  window.addEventListener('message', function (event) {
    const data = event && event.data;
    if (!data || data.__piclawGeneratedWidgetHost !== true) return;
    if ((data.widgetId || null) !== (meta.widgetId || null)) return;
    scheduleHostEnvelope(data);
  });

  function announceReady() {
    readWindowNameState();
    post('widget.ready', { title: document.title || meta.title || 'Generated widget' });
  }

  setInterval(readWindowNameState, 250);

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', announceReady, { once: true });
  } else {
    announceReady();
  }
})();
</script>`}function T6($){let j=$?.artifact||{},q=j.kind||$?.kind||null,B=typeof j.html==="string"?j.html:typeof $?.html==="string"?$.html:"",_=typeof j.svg==="string"?j.svg:typeof $?.svg==="string"?$.svg:"",Q=typeof $?.title==="string"&&$.title.trim()?$.title.trim():"Generated widget",U=q==="svg"?_:B;if(!U)return"";let X=C6($),J=["default-src 'none'","img-src data: blob: https: http:","style-src 'unsafe-inline'","font-src 'self' data: https: http:","media-src data: blob: https: http:","connect-src 'none'","frame-src 'none'",X?"script-src 'unsafe-inline' 'self'":"script-src 'none'","object-src 'none'","base-uri 'none'","form-action 'none'"].join("; "),W=q==="svg"?`<div class="widget-svg-shell">${U}</div>`:U,V=X?d7($):"";return`<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta http-equiv="Content-Security-Policy" content="${J}" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>${Q.replace(/[<&>]/g,"")}</title>
<style>
:root { color-scheme: dark light; }
html, body {
  margin: 0;
  padding: 0;
  min-height: 100%;
  background: #0f1117;
  color: #f5f7fb;
  font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}
body {
  box-sizing: border-box;
}
.widget-svg-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  box-sizing: border-box;
}
.widget-svg-shell svg {
  max-width: 100%;
  height: auto;
}
</style>
${V}
</head>
<body>${W}</body>
</html>`}function h2({children:$,className:j=""}){let[q,B]=y(null);return E(()=>{if(typeof document>"u")return;let _=document.createElement("div");if(j)_.className=j;return document.body.appendChild(_),B(_),()=>{try{m2(null,_)}finally{_.remove(),B((Q)=>Q===_?null:Q)}}},[j]),l5(()=>{if(!q)return;m2($,q);return},[$,q]),null}function M6({src:$,onClose:j}){return E(()=>{let q=(B)=>{if(B.key==="Escape")j()};return document.addEventListener("keydown",q),()=>document.removeEventListener("keydown",q)},[j]),K`
        <${h2} className="image-modal-portal-root">
            <div class="image-modal" onClick=${j}>
                <img src=${$} alt="Full size" />
            </div>
        </${h2}>
    `}function w0({prefix:$="file",label:j,title:q,onRemove:B,onClick:_,removeTitle:Q="Remove",icon:U="file"}){let X=`${$}-file-pill`,J=`${$}-file-name`,W=`${$}-file-remove`,V=U==="message"?K`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>`:K`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
      </svg>`;return K`
    <span class=${X} title=${q||j} onClick=${_}>
      ${V}
      <span class=${J}>${j}</span>
      ${B&&K`
        <button
          class=${W}
          onClick=${(G)=>{G.preventDefault(),G.stopPropagation(),B()}}
          title=${Q}
          type="button"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      `}
    </span>
  `}async function S6($,j){try{return await $?.writeText?.(j),!0}catch(q){return!1}}function V3($,j){let q=typeof j?.text==="string"?j.text:"",B=typeof j?.html==="string"?j.html:"";if(!$||!q||typeof $.createElement!=="function"||typeof $.execCommand!=="function")return!1;let _=null,Q=!1,U=(X)=>{let J=X?.clipboardData;if(!J||typeof J.setData!=="function")return;if(J.setData("text/plain",q),B)J.setData("text/html",B);if(typeof X.preventDefault==="function")X.preventDefault();Q=!0};try{if(_=$.createElement("textarea"),_.value=q,typeof _.setAttribute==="function")_.setAttribute("readonly","");if(_.style)_.style.position="fixed",_.style.opacity="0",_.style.pointerEvents="none";if($.body?.appendChild?.(_),typeof _.select==="function")_.select();if(typeof _.setSelectionRange==="function")_.setSelectionRange(0,_.value.length);$.addEventListener?.("copy",U,!0);let X=$.execCommand("copy");return Boolean(Q||X)}catch{return!1}finally{if($.removeEventListener?.("copy",U,!0),_)$.body?.removeChild?.(_)}}function y6($){if(!$||typeof $!=="object")return null;let j=$;if(typeof j.nodeType==="number"&&j.nodeType===3)return j.parentNode||null;return j}function k6($,j){let q=$?.clipboardData,B=j?.root,_=j?.selection;if(!q||typeof q.setData!=="function"||!B||!_)return!1;if(_.isCollapsed)return!1;let Q=!1;if(Number(_.rangeCount||0)>0&&typeof _.getRangeAt==="function")try{let J=_.getRangeAt(0);if(J&&typeof J.intersectsNode==="function")Q=Boolean(J.intersectsNode(B))}catch{Q=!1}if(!Q&&typeof B.contains==="function"){let J=y6(_.anchorNode),W=y6(_.focusNode);Q=Boolean(J&&B.contains(J)||W&&B.contains(W))}if(!Q)return!1;let X=typeof _.toString==="function"?String(_.toString()||"").replace(/\u00a0/g," "):"";if(!X)return!1;return q.setData("text/plain",X),$?.preventDefault?.(),!0}function E6($,j){try{return Boolean($?.getItem?.(j))}catch(q){return!1}}function f6($,j,q){try{return $?.setItem?.(j,q),!0}catch(B){return!1}}function P6($,j){let q=typeof $==="string"&&$.trim()?$.trim():null;if(q)return q;if(!j)return null;try{return new URL(j).hostname}catch(B){return j}}function i7({mediaId:$,onPreview:j}){let[q,B]=y(null);if(E(()=>{$3($).then(B).catch((W)=>{console.warn("[post] Failed to load attachment metadata for file card:",$,W)})},[$]),!q)return null;let _=q.filename||"file",Q=q.metadata?.size,U=Q?C0(Q):"",J=X4(q.content_type,q.filename)==="unsupported"?"Details":"Preview";return K`
        <div class="file-attachment" onClick=${(W)=>W.stopPropagation()}>
            <a href=${i0($)} download=${_} class="file-attachment-main">
                <svg class="file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                    <polyline points="10 9 9 9 8 9"/>
                </svg>
                <div class="file-info">
                    <span class="file-name">${_}</span>
                    <span class="file-meta-row">
                        ${U&&K`<span class="file-size">${U}</span>`}
                        ${q.content_type&&K`<span class="file-size">${q.content_type}</span>`}
                    </span>
                </div>
                <svg class="download-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="7 10 12 15 17 10"/>
                    <line x1="12" y1="15" x2="12" y2="3"/>
                </svg>
            </a>
            <button
                class="file-attachment-preview"
                type="button"
                onClick=${(W)=>{W.preventDefault(),W.stopPropagation(),j?.({mediaId:$,info:q})}}
            >
                ${J}
            </button>
        </div>
    `}function n7($){if(!Array.isArray($))return[];return $.filter((j)=>j&&typeof j==="object"&&j.type==="recovery_marker"&&j.recovered)}function s7($){if(!Array.isArray($))return[];return $.filter((j)=>j&&typeof j==="object"&&j.type==="timeout_marker"&&(j.timed_out??!0))}var o7={context_recover:"context limit exceeded",rate_limit:"rate limit hit",api_error:"API error",timeout:"request timeout",overloaded:"service overloaded",connection:"connection error"};function a7($){let j=Number($?.attempts_used||0),q=String($?.classifier||"").trim(),B=o7[q]||(q?q.replace(/_/g," "):""),_=["Recovered automatically"];if(j>1)_[0]=`Recovered after ${j} attempts`;if(B)_.push(B);return _.join(" — ")}function t7($){let j=typeof $?.tool_action_summary==="string"?$.tool_action_summary.trim():"";return j?`Turn timed out — ${j}`:"Turn timed out before the model finished responding"}function e7({attachment:$,onPreview:j}){let q=Number($?.id),[B,_]=y(null);E(()=>{if(!Number.isFinite(q))return;$3(q).then(_).catch((W)=>{console.warn("[post] Failed to load attachment metadata for attachment pill:",q,W)});return},[q]);let Q=B?.filename||$.label||`attachment-${$.id}`,U=Number.isFinite(q)?i0(q):null,J=X4(B?.content_type,B?.filename||$?.label)==="unsupported"?"Details":"Preview";return K`
        <span class="attachment-pill" title=${Q}>
            ${U?K`
                    <a href=${U} download=${Q} class="attachment-pill-main" onClick=${(W)=>W.stopPropagation()}>
                        <${w0}
                            prefix="post"
                            label=${$.label}
                            title=${Q}
                        />
                    </a>
                `:K`
                    <${w0}
                        prefix="post"
                        label=${$.label}
                        title=${Q}
                    />
                `}
            ${Number.isFinite(q)&&B&&K`
                <button
                    class="attachment-pill-preview"
                    type="button"
                    title=${J}
                    onClick=${(W)=>{W.preventDefault(),W.stopPropagation(),j?.({mediaId:q,info:B})}}
                >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z"/>
                        <circle cx="12" cy="12" r="3"/>
                    </svg>
                </button>
            `}
        </span>
    `}function J5({annotations:$}){if(!$)return null;let{audience:j,priority:q,lastModified:B}=$,_=B?p2(B):null;return K`
        <div class="content-annotations">
            ${j&&j.length>0&&K`
                <span class="content-annotation">Audience: ${j.join(", ")}</span>
            `}
            ${typeof q==="number"&&K`
                <span class="content-annotation">Priority: ${q}</span>
            `}
            ${_&&K`
                <span class="content-annotation">Updated: ${_}</span>
            `}
        </div>
    `}function $j({block:$}){let j=$.title||$.name||$.uri,q=$.description,B=$.size?C0($.size):"",_=$.mime_type||"",Q=Bj(_),U=E2($.uri);return K`
        <a
            href=${U||"#"}
            class="resource-link"
            target=${U?"_blank":void 0}
            rel=${U?"noopener noreferrer":void 0}
            onClick=${(X)=>X.stopPropagation()}>
            <div class="resource-link-main">
                <div class="resource-link-header">
                    <span class="resource-link-icon-inline">${Q}</span>
                    <div class="resource-link-title">${j}</div>
                </div>
                ${q&&K`<div class="resource-link-description">${q}</div>`}
                <div class="resource-link-meta">
                    ${_&&K`<span>${_}</span>`}
                    ${B&&K`<span>${B}</span>`}
                </div>
            </div>
            <div class="resource-link-icon">↗</div>
        </a>
    `}function jj({block:$}){let[j,q]=y(!1),B=$.uri||"Embedded resource",_=$.text||"",Q=Boolean($.data),U=$.mime_type||"";return K`
        <div class="resource-embed">
            <button class="resource-embed-toggle" onClick=${(X)=>{X.preventDefault(),X.stopPropagation(),q(!j)}}>
                ${j?"▼":"▶"} ${B}
            </button>
            ${j&&K`
                ${_&&K`<pre class="resource-embed-content">${_}</pre>`}
                ${Q&&K`
                    <div class="resource-embed-blob">
                        <span class="resource-embed-blob-label">Embedded blob</span>
                        ${U&&K`<span class="resource-embed-blob-meta">${U}</span>`}
                        <button class="resource-embed-blob-btn" onClick=${(X)=>{X.preventDefault(),X.stopPropagation();let J=new Blob([Uint8Array.from(atob($.data),(G)=>G.charCodeAt(0))],{type:U||"application/octet-stream"}),W=URL.createObjectURL(J),V=document.createElement("a");V.href=W,V.download=B.split("/").pop()||"resource",V.click(),URL.revokeObjectURL(W)}}>Download</button>
                    </div>
                `}
            `}
        </div>
    `}function qj({block:$,post:j,onOpenWidget:q}){if(!$)return null;let B=G3($,j),_=D6($),Q=B?.artifact?.kind||$?.artifact?.kind||$?.kind||null,U=B?.title||$.title||$.name||"Generated widget",X=B?.description||$.description||$.subtitle||"",J=$.open_label||"Open widget",W=M(!1),V=(G)=>{if(G)G.preventDefault(),G.stopPropagation();if(!B)return;q?.(B)};return E(()=>{if(!$?.auto_open||!B||!_||W.current)return;let G=j?.timestamp?new Date(j.timestamp).getTime():0;if(G&&Date.now()-G>1e4)return;let L=`widget_opened_${$.widget_id||j?.id||""}`;if(E6(sessionStorage,L))return;W.current=!0,f6(sessionStorage,L,"1"),q?.(B)},[$?.auto_open,B,_]),K`
        <div class="generated-widget-launch" onClick=${(G)=>G.stopPropagation()}>
            <div class="generated-widget-launch-header">
                <div class="generated-widget-launch-eyebrow">Generated widget${Q?` • ${String(Q).toUpperCase()}`:""}</div>
                <div class="generated-widget-launch-title">${U}</div>
            </div>
            ${X&&K`<div class="generated-widget-launch-description">${X}</div>`}
            <div class="generated-widget-launch-actions">
                <button
                    class="generated-widget-launch-btn"
                    type="button"
                    disabled=${!_}
                    onClick=${V}
                    title=${_?"Open widget in a floating pane":"Unsupported widget artifact"}
                >
                    ${J}
                </button>
                <span class="generated-widget-launch-note">
                    ${_?"Opens in a dismissible floating pane.":"This widget artifact is missing or unsupported."}
                </span>
            </div>
        </div>
    `}function Bj($){if(!$)return"\uD83D\uDCCE";if($.startsWith("image/"))return"\uD83D\uDDBC️";if($.startsWith("audio/"))return"\uD83C\uDFB5";if($.startsWith("video/"))return"\uD83C\uDFAC";if($.includes("pdf"))return"\uD83D\uDCC4";if($.includes("zip")||$.includes("gzip"))return"\uD83D\uDDDC️";if($.startsWith("text/"))return"\uD83D\uDCC4";return"\uD83D\uDCCE"}function _j($){let j=E2($,{allowDataImage:!0});return j?{backgroundImage:`url("${j}")`}:void 0}function Qj({preview:$}){let j=E2($.url),q=_j($.image),B=P6($.site_name,j);return K`
        <a
            href=${j||"#"}
            class="link-preview ${q?"has-image":""}"
            target=${j?"_blank":void 0}
            rel=${j?"noopener noreferrer":void 0}
            onClick=${(_)=>_.stopPropagation()}
            style=${q}>
            <div class="link-preview-overlay">
                <div class="link-preview-site">${B||""}</div>
                <div class="link-preview-title">${$.title}</div>
                ${$.description&&K`
                    <div class="link-preview-description">${$.description}</div>
                `}
            </div>
        </a>
    `}function Uj($,j){return typeof $==="string"?$:""}var R6=1800,Xj=`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>`,Wj=`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path d="M20 6L9 17l-5-5"></path>
    </svg>`,Jj=`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <circle cx="12" cy="12" r="9"></circle>
        <path d="M9 9l6 6M15 9l-6 6"></path>
    </svg>`,Kj=`
<style>
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    font-size: 14px;
    line-height: 1.6;
    color: #1a1a1a;
  }
  h1 { font-size: 1.6em; font-weight: 700; margin: 0.6em 0 0.4em; }
  h2 { font-size: 1.35em; font-weight: 700; margin: 0.6em 0 0.4em; }
  h3 { font-size: 1.15em; font-weight: 700; margin: 0.5em 0 0.3em; }
  h4, h5, h6 { font-size: 1em; font-weight: 700; margin: 0.5em 0 0.3em; }
  p { margin: 0.5em 0; }
  pre {
    background: #f6f8fa;
    border: 1px solid #d0d7de;
    border-radius: 6px;
    padding: 12px 16px;
    overflow-x: auto;
    margin: 0.5em 0;
  }
  code {
    font-family: "Fira Code", "Cascadia Code", Consolas, "Courier New", monospace;
    font-size: 0.9em;
  }
  pre code { background: none; padding: 0; border: none; }
  :not(pre) > code { background: #f0f2f5; padding: 2px 5px; border-radius: 3px; }
  blockquote { border-left: 3px solid #d0d7de; margin: 0.5em 0; padding-left: 12px; color: #57606a; }
  table { border-collapse: collapse; margin: 0.5em 0; }
  th, td { border: 1px solid #d0d7de; padding: 6px 12px; text-align: left; }
  th { background: #f6f8fa; font-weight: 600; }
  ul, ol { margin: 0.4em 0; padding-left: 1.8em; }
  li { margin: 0.15em 0; }
  a { color: #0969da; text-decoration: none; }
  hr { border: none; border-top: 1px solid #d0d7de; margin: 1em 0; }
  img { max-width: 100%; }
</style>`;async function w6($){let j=typeof $==="string"?$:"";if(!j)return!1;if(V3(document,{text:j}))return!0;if(await S6(navigator.clipboard,j))return!0;try{let q=document.createElement("textarea");q.value=j,q.setAttribute("readonly",""),q.style.position="fixed",q.style.opacity="0",q.style.pointerEvents="none",document.body.appendChild(q),q.select(),q.setSelectionRange(0,q.value.length);let B=document.execCommand("copy");return document.body.removeChild(q),B}catch{return!1}}async function zj($){let j=typeof $==="string"?$:"";if(!j)return!1;let q=B2(j,null),B=`<html><head>${Kj}</head><body>${q}</body></html>`;if(V3(document,{text:j,html:B}))return!0;if(navigator.clipboard?.write&&typeof ClipboardItem<"u")try{let _=new ClipboardItem({"text/plain":new Blob([j],{type:"text/plain"}),"text/html":new Blob([B],{type:"text/html"})});return await navigator.clipboard.write([_]),!0}catch(_){console.warn("[post] Rich clipboard write failed, falling back to plain text copy.",_)}return w6(j)}function Gj($){if(!$)return()=>{};let j=Array.from($.querySelectorAll("pre")).filter((U)=>U.querySelector("code"));if(j.length===0)return()=>{};let q=new Map,B=[],_=(U)=>{let X=window.getSelection?.();if(!X||X.isCollapsed)return;for(let J of j)if(k6(U,{root:J,selection:X}))return};document.addEventListener("copy",_,!0),B.push(()=>document.removeEventListener("copy",_,!0));let Q=(U,X)=>{let J=X||"idle";if(U.dataset.copyState=J,J==="success")U.innerHTML=Wj,U.setAttribute("aria-label","Copied"),U.setAttribute("title","Copied"),U.classList.add("is-success"),U.classList.remove("is-error");else if(J==="error")U.innerHTML=Jj,U.setAttribute("aria-label","Copy failed"),U.setAttribute("title","Copy failed"),U.classList.add("is-error"),U.classList.remove("is-success");else U.innerHTML=Xj,U.setAttribute("aria-label","Copy code"),U.setAttribute("title","Copy code"),U.classList.remove("is-success","is-error")};return j.forEach((U)=>{let X=document.createElement("div");X.className="post-code-block",U.parentNode?.insertBefore(X,U),X.appendChild(U);let J=document.createElement("button");J.type="button",J.className="post-code-copy-btn",Q(J,"idle"),X.appendChild(J);let W=async(V)=>{V.preventDefault(),V.stopPropagation();let L=U.querySelector("code")?.textContent||"",C=await w6(L);Q(J,C?"success":"error");let b=q.get(J);if(b)clearTimeout(b);let c=setTimeout(()=>{Q(J,"idle"),q.delete(J)},R6);q.set(J,c)};J.addEventListener("click",W),B.push(()=>{J.removeEventListener("click",W);let V=q.get(J);if(V)clearTimeout(V);if(X.parentNode)X.parentNode.insertBefore(U,X),X.remove()})}),()=>{B.forEach((U)=>U())}}function Zj($){if(!$)return{content:$,fileRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let W=0;W<q.length;W+=1)if(q[W].trim()==="Files:"&&q[W+1]&&/^\s*-\s+/.test(q[W+1])){B=W;break}if(B===-1)return{content:$,fileRefs:[]};let _=[],Q=B+1;for(;Q<q.length;Q+=1){let W=q[Q];if(/^\s*-\s+/.test(W))_.push(W.replace(/^\s*-\s+/,"").trim());else if(!W.trim())break;else break}if(_.length===0)return{content:$,fileRefs:[]};let U=q.slice(0,B),X=q.slice(Q),J=[...U,...X].join(`
`);return J=J.replace(/\n{3,}/g,`

`).trim(),{content:J,fileRefs:_}}function Vj($){if(!$)return{content:$,messageRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let W=0;W<q.length;W+=1)if(q[W].trim()==="Referenced messages:"&&q[W+1]&&/^\s*-\s+/.test(q[W+1])){B=W;break}if(B===-1)return{content:$,messageRefs:[]};let _=[],Q=B+1;for(;Q<q.length;Q+=1){let W=q[Q];if(/^\s*-\s+/.test(W)){let G=W.replace(/^\s*-\s+/,"").trim().match(/^message:(\S+)$/i);if(G)_.push(G[1])}else if(!W.trim())break;else break}if(_.length===0)return{content:$,messageRefs:[]};let U=q.slice(0,B),X=q.slice(Q),J=[...U,...X].join(`
`);return J=J.replace(/\n{3,}/g,`

`).trim(),{content:J,messageRefs:_}}function Lj($){if(!$)return{content:$,attachments:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let W=0;W<q.length;W+=1){let V=q[W].trim();if((V==="Images:"||V==="Attachments:")&&q[W+1]&&/^\s*-\s+/.test(q[W+1])){B=W;break}}if(B===-1)return{content:$,attachments:[]};let _=[],Q=B+1;for(;Q<q.length;Q+=1){let W=q[Q];if(/^\s*-\s+/.test(W)){let V=W.replace(/^\s*-\s+/,"").trim(),G=V.match(/^attachment:([^\s)]+)\s*(?:\((.+)\))?$/i)||V.match(/^attachment:([^\s]+)\s+(.+)$/i);if(G){let L=G[1],C=(G[2]||"").trim()||L;_.push({id:L,label:C,raw:V})}else _.push({id:null,label:V,raw:V})}else if(!W.trim())break;else break}if(_.length===0)return{content:$,attachments:[]};let U=q.slice(0,B),X=q.slice(Q),J=[...U,...X].join(`
`);return J=J.replace(/\n{3,}/g,`

`).trim(),{content:J,attachments:_}}function Nj($){return $.replace(/[.*+?^${}()|[\]\\]/g,"\\$&")}function Fj($,j){if(!$||!j)return $;let q=String(j).trim().split(/\s+/).filter(Boolean);if(q.length===0)return $;let B=q.map(Nj).sort((V,G)=>G.length-V.length),_=new RegExp(`(${B.join("|")})`,"gi"),Q=new RegExp(`^(${B.join("|")})$`,"i"),U=new DOMParser().parseFromString($,"text/html"),X=U.createTreeWalker(U.body,NodeFilter.SHOW_TEXT),J=[],W;while(W=X.nextNode())J.push(W);for(let V of J){let G=V.nodeValue;if(!G||!_.test(G)){_.lastIndex=0;continue}_.lastIndex=0;let L=V.parentElement;if(L&&L.closest("code, pre, script, style"))continue;let C=G.split(_).filter((c)=>c!=="");if(C.length===0)continue;let b=U.createDocumentFragment();for(let c of C)if(Q.test(c)){let l=U.createElement("mark");l.className="search-highlight-term",l.textContent=c,b.appendChild(l)}else b.appendChild(U.createTextNode(c));V.parentNode.replaceChild(b,V)}return U.body.innerHTML}function x6({post:$,onClick:j,onHashtagClick:q,onMessageRef:B,onScrollToMessage:_,agentName:Q,agentAvatarUrl:U,userName:X,userAvatarUrl:J,userAvatarBackground:W,onDelete:V,isThreadReply:G,isThreadPrev:L,isThreadNext:C,isRemoving:b,highlightQuery:c,onFileRef:l,onOpenWidget:w,onOpenAttachmentPreview:S}){let[m,O]=y(null),[d,g]=y("idle"),U1=M(null),k=M(null),r=$.data,f=r.type==="agent_response",_1=X||"You",Y1=f?Q||Q6:_1,k1=typeof $.chat_agent_name==="string"?$.chat_agent_name.trim():"",O1=Boolean(f&&c&&k1&&k1!==Y1),s=f?X3(Q,U,!0):X3(_1,J),t=typeof W==="string"?W.trim().toLowerCase():"",K1=!f&&s.image&&(t==="clear"||t==="transparent"),G1=f&&Boolean(s.image),I=`background-color: ${K1||G1?"transparent":s.color}`,a=r.content_meta,F1=Boolean(a?.truncated),X1=Boolean(a?.preview),T1=F1&&!X1,u1=F1?{originalLength:Number.isFinite(a?.original_length)?a.original_length:r.content?r.content.length:0,maxLength:Number.isFinite(a?.max_length)?a.max_length:0}:null,A1=r.content_blocks||[],t1=r.media_ids||[],M1=Uj(r.content,r.link_previews),{content:G0,fileRefs:l1}=Zj(M1),{content:I0,messageRefs:j0}=Vj(G0),{content:g0,attachments:F0}=Lj(I0);M1=g0;let q0=z3(A1),B0=U3(A1),S1=n7(A1)[0]||null,H0=s7(A1)[0]||null,_0=q0.length===1&&typeof q0[0]?.fallback_text==="string"?q0[0].fallback_text.trim():"",Q0=B0.length===1?q5(B0[0]).trim():"",h1=Boolean(_0)&&M1?.trim()===_0||Boolean(Q0)&&M1?.trim()===Q0,Y0=Boolean(M1)&&!T1&&!h1,E1=typeof c==="string"?c.trim():"",R1=B1(()=>{if(!M1||h1)return"";let A=B2(M1,q);return E1?Fj(A,E1):A},[M1,h1,E1]),w1=B1(()=>_6($),[$]),e1=(A,P)=>{A.stopPropagation(),O(i0(P))},f1=(A)=>{S?.(A)},A0=(A)=>{A.stopPropagation(),V?.($)},d1=async(A)=>{A.stopPropagation();let P=await zj(w1);if(g(P?"success":"error"),k.current)clearTimeout(k.current);k.current=setTimeout(()=>{k.current=null,g("idle")},R6)},f0=(A,P)=>{let Q1=new Set;if(!A||P.length===0)return{content:A,usedIds:Q1};return{content:A.replace(/attachment:([^\s)"']+)/g,(o,C1,v1,I1)=>{let y1=C1.replace(/^\/+/,""),D1=P.find((M0)=>M0.name&&M0.name.toLowerCase()===y1.toLowerCase()&&!Q1.has(M0.id))||P.find((M0)=>!Q1.has(M0.id));if(!D1)return o;if(Q1.add(D1.id),I1.slice(Math.max(0,v1-2),v1)==="](")return`/media/${D1.id}`;return D1.name||"attachment"}),usedIds:Q1}},x1=[],P1=[],r1=[],$0=[],Z0=[],m1=[],j1=[],H1=0;if(A1.length>0)A1.forEach((A)=>{if(A?.type==="text"&&A.annotations)j1.push(A.annotations);if(A?.type==="generated_widget")m1.push(A);else if(A?.type==="resource_link")$0.push(A);else if(A?.type==="resource")Z0.push(A);else if(A?.type==="file"){let P=t1[H1++];if(P)P1.push(P),r1.push({id:P,name:A?.name||A?.filename||A?.title})}else if(A?.type==="image"||!A?.type){let P=t1[H1++];if(P){let Q1=typeof A?.mime_type==="string"?A.mime_type:void 0;x1.push({id:P,annotations:A?.annotations,mimeType:Q1}),r1.push({id:P,name:A?.name||A?.filename||A?.title})}}});else if(t1.length>0){let A=F0.length>0;t1.forEach((P,Q1)=>{let z1=F0[Q1]||null;if(r1.push({id:P,name:z1?.label||null}),A)P1.push(P);else x1.push({id:P,annotations:null})})}if(F0.length>0)F0.forEach((A)=>{if(!A?.id)return;let P=r1.find((Q1)=>String(Q1.id)===String(A.id));if(P&&!P.name)P.name=A.label});let{content:H,usedIds:p}=f0(M1,r1);M1=H;let v=x1.filter(({id:A})=>!p.has(A)),e=P1.filter((A)=>!p.has(A)),$1=F0.length>0?F0.map((A,P)=>({id:A.id||`attachment-${P+1}`,label:A.label||`attachment-${P+1}`})):r1.map((A,P)=>({id:A.id,label:A.name||`attachment-${P+1}`})),V1=B1(()=>z3(A1),[A1]),L1=B1(()=>U3(A1),[A1]),Z1=B1(()=>{return V1.map((A)=>`${A.card_id}:${A.state}`).join("|")},[V1]);E(()=>{if(!U1.current)return;return j5(U1.current),Gj(U1.current)},[R1]),E(()=>()=>{if(k.current)clearTimeout(k.current)},[]);let N1=M(null);return E(()=>{if(!N1.current||V1.length===0)return;let A=N1.current;A.innerHTML="";for(let P of V1){let Q1=document.createElement("div");A.appendChild(Q1),H6(Q1,P,{onAction:async(z1)=>{if(z1.type==="Action.OpenUrl"){let o=E2(z1.url||"");if(!o)throw Error("Invalid URL");window.open(o,"_blank","noopener,noreferrer");return}if(z1.type==="Action.Submit"){await x8({post_id:$.id,thread_id:r.thread_id||$.id,chat_jid:$.chat_jid||null,card_id:P.card_id,action:{type:z1.type,title:z1.title||"",data:z1.data}});return}console.warn("[post] unsupported adaptive card action:",z1.type,z1)}}).catch((z1)=>{console.error("[post] adaptive card render error:",z1),Q1.textContent=P.fallback_text||"Card failed to render."})}},[Z1,$.id]),K`
        <div id=${`post-${$.id}`} class="post ${f?"agent-post":""} ${G?"thread-reply":""} ${L?"thread-prev":""} ${C?"thread-next":""} ${b?"removing":""}" onClick=${j}>
            <div class="post-avatar ${f?"agent-avatar":""} ${s.image?"has-image":""}" style=${I}>
                ${s.image?K`<img src=${s.image} alt=${Y1} />`:s.letter}
            </div>
            <div class="post-body">
                <div class="post-actions">
                    <button
                        class=${`post-action-btn post-copy-btn${d==="success"?" is-success":d==="error"?" is-error":""}`}
                        type="button"
                        title=${d==="success"?"Copied":d==="error"?"Copy failed":"Copy message"}
                        aria-label=${d==="success"?"Copied":d==="error"?"Copy failed":"Copy message"}
                        onClick=${d1}
                        disabled=${!w1}
                    >
                        ${d==="success"?K`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M20 6L9 17l-5-5"></path></svg>`:d==="error"?K`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><circle cx="12" cy="12" r="9"></circle><path d="M9 9l6 6M15 9l-6 6"></path></svg>`:K`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><rect x="9" y="9" width="10" height="10" rx="2"></rect><path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path></svg>`}
                    </button>
                    <button
                        class="post-action-btn post-delete-btn"
                        type="button"
                        title="Delete message"
                        aria-label="Delete message"
                        onClick=${A0}
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                            <path d="M18 6L6 18M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                <div class="post-meta">
                    <span class="post-author">${Y1}</span>
                    ${O1&&K`<span class="post-chat-agent-tag" title=${`Chat: ${k1}`}>@${k1}</span>`}
                    ${S1&&K`
                        <span
                            class="post-recovery-chip"
                            title=${a7(S1)}
                        >
                            recovered
                        </span>
                    `}
                    ${H0&&K`
                        <span
                            class="post-recovery-chip post-timeout-chip"
                            title=${t7(H0)}
                        >
                            timeout
                        </span>
                    `}
                    <a class="post-time" href=${`#msg-${$.id}`} onClick=${(A)=>{if(A.preventDefault(),A.stopPropagation(),B)B($.id)}}>${$6($.timestamp)}</a>
                </div>
                ${T1&&u1&&K`
                    <div class="post-content truncated">
                        <div class="truncated-title">Message too large to display.</div>
                        <div class="truncated-meta">
                            Original length: ${Q4(u1.originalLength)} chars
                            ${u1.maxLength?K` • Display limit: ${Q4(u1.maxLength)} chars`:""}
                        </div>
                    </div>
                `}
                ${X1&&u1&&K`
                    <div class="post-content preview">
                        <div class="truncated-title">Preview truncated.</div>
                        <div class="truncated-meta">
                            Showing first ${Q4(u1.maxLength)} of ${Q4(u1.originalLength)} chars. Download full text below.
                        </div>
                    </div>
                `}
                ${(l1.length>0||j0.length>0||$1.length>0)&&K`
                    <div class="post-file-refs">
                        ${j0.map((A)=>{let P=(Q1)=>{if(Q1.preventDefault(),Q1.stopPropagation(),_)_(A,$.chat_jid||null);else{let z1=document.getElementById("post-"+A);if(z1)z1.scrollIntoView({behavior:"smooth",block:"center"}),z1.classList.add("post-highlight"),setTimeout(()=>z1.classList.remove("post-highlight"),2000)}};return K`
                                <a href=${`#msg-${A}`} class="post-msg-pill-link" onClick=${P}>
                                    <${w0}
                                        prefix="post"
                                        label=${"msg:"+A}
                                        title=${"Message "+A}
                                        icon="message"
                                        onClick=${P}
                                    />
                                </a>
                            `})}
                        ${l1.map((A)=>{let P=A.split("/").pop()||A;return K`
                                <${w0}
                                    prefix="post"
                                    label=${P}
                                    title=${A}
                                    onClick=${()=>l?.(A)}
                                />
                            `})}
                        ${$1.map((A)=>K`
                            <${e7}
                                key=${A.id}
                                attachment=${A}
                                onPreview=${f1}
                            />
                        `)}
                    </div>
                `}
                ${Y0&&K`
                    <div 
                        ref=${U1}
                        class="post-content"
                        dangerouslySetInnerHTML=${{__html:R1}}
                        onClick=${(A)=>{if(A.target.classList.contains("hashtag")){A.preventDefault(),A.stopPropagation();let P=A.target.dataset.hashtag;if(P)q?.(P)}else if(A.target.tagName==="IMG")A.preventDefault(),A.stopPropagation(),O(A.target.src)}}
                    />
                `}
                ${V1.length>0&&K`
                    <div ref=${N1} class="post-adaptive-cards" />
                `}
                ${L1.length>0&&K`
                    <div class="post-adaptive-card-submissions">
                        ${L1.map((A,P)=>{let Q1=q6(A),z1=`${A.card_id}-${P}`;return K`
                                <div key=${z1} class="adaptive-card-submission-receipt">
                                    <div class="adaptive-card-submission-header">
                                        <span class="adaptive-card-submission-icon" aria-hidden="true">✓</span>
                                        <div class="adaptive-card-submission-title-wrap">
                                            <span class="adaptive-card-submission-title">Submitted</span>
                                            <span class="adaptive-card-submission-title-action">${Q1.title}</span>
                                        </div>
                                    </div>
                                    ${Q1.fields.length>0&&K`
                                        <div class="adaptive-card-submission-fields">
                                            ${Q1.fields.map((o)=>K`
                                                <span class="adaptive-card-submission-field" title=${`${o.key}: ${o.value}`}>
                                                    <span class="adaptive-card-submission-field-key">${o.key}</span>
                                                    <span class="adaptive-card-submission-field-sep">:</span>
                                                    <span class="adaptive-card-submission-field-value">${o.value}</span>
                                                </span>
                                            `)}
                                        </div>
                                    `}
                                    <div class="adaptive-card-submission-meta">
                                        Submitted ${p2(Q1.submittedAt)}
                                    </div>
                                </div>
                            `})}
                    </div>
                `}
                ${m1.length>0&&K`
                    <div class="generated-widget-launches">
                        ${m1.map((A,P)=>K`
                            <${qj}
                                key=${A.widget_id||A.id||`${$.id}-widget-${P}`}
                                block=${A}
                                post=${$}
                                onOpenWidget=${w}
                            />
                        `)}
                    </div>
                `}
                ${j1.length>0&&K`
                    ${j1.map((A,P)=>K`
                        <${J5} key=${P} annotations=${A} />
                    `)}
                `}
                ${v.length>0&&K`
                    <div class="media-preview">
                        ${v.map(({id:A,mimeType:P})=>{let z1=typeof P==="string"&&P.toLowerCase().startsWith("image/svg")?i0(A):w8(A);return K`
                                <img 
                                    key=${A} 
                                    src=${z1} 
                                    alt="Media" 
                                    loading="lazy"
                                    decoding="async"
                                    onClick=${(o)=>e1(o,A)}
                                />
                            `})}
                    </div>
                `}
                ${v.length>0&&K`
                    ${v.map(({annotations:A},P)=>K`
                        ${A&&K`<${J5} key=${P} annotations=${A} />`}
                    `)}
                `}
                ${e.length>0&&K`
                    <div class="file-attachments">
                        ${e.map((A)=>K`
                            <${i7} key=${A} mediaId=${A} onPreview=${f1} />
                        `)}
                    </div>
                `}
                ${$0.length>0&&K`
                    <div class="resource-links">
                        ${$0.map((A,P)=>K`
                            <div key=${P}>
                                <${$j} block=${A} />
                                <${J5} annotations=${A.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${Z0.length>0&&K`
                    <div class="resource-embeds">
                        ${Z0.map((A,P)=>K`
                            <div key=${P}>
                                <${jj} block=${A} />
                                <${J5} annotations=${A.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${r.link_previews?.length>0&&K`
                    <div class="link-previews">
                        ${r.link_previews.map((A,P)=>K`
                            <${Qj} key=${P} preview=${A} />
                        `)}
                    </div>
                `}
            </div>
        </div>
        ${m&&K`<${M6} src=${m} onClose=${()=>O(null)} />`}

    `}function v6({posts:$,hasMore:j,onLoadMore:q,onPostClick:B,onHashtagClick:_,onMessageRef:Q,onScrollToMessage:U,onFileRef:X,onOpenWidget:J,onOpenAttachmentPreview:W,emptyMessage:V,timelineRef:G,agents:L,user:C,onDeletePost:b,reverse:c=!0,removingPostIds:l,searchQuery:w}){let[S,m]=y(!1),O=M(null),d=typeof IntersectionObserver<"u",g=i(async()=>{if(!q||!j||S)return;m(!0);try{await q({preserveScroll:!0,preserveMode:"top"})}finally{m(!1)}},[j,S,q]),U1=i((s)=>{let{scrollTop:t,scrollHeight:K1,clientHeight:G1}=s.target,I=c?K1-G1-t:t,a=Math.max(300,G1);if(I<a)g()},[c,g]);E(()=>{if(!d)return;let s=O.current,t=G?.current;if(!s||!t)return;let K1=300,G1=new IntersectionObserver((I)=>{for(let a of I){if(!a.isIntersecting)continue;g()}},{root:t,rootMargin:`${K1}px 0px ${K1}px 0px`,threshold:0});return G1.observe(s),()=>G1.disconnect()},[d,j,q,G,g]);let k=M(g);if(k.current=g,E(()=>{if(d)return;if(!G?.current)return;let{scrollTop:s,scrollHeight:t,clientHeight:K1}=G.current,G1=c?t-K1-s:s,I=Math.max(300,K1);if(G1<I)k.current?.()},[d,$,j,c,G]),E(()=>{if(!G?.current)return;if(!j||S)return;let{scrollTop:s,scrollHeight:t,clientHeight:K1}=G.current,G1=c?t-K1-s:s,I=Math.max(300,K1);if(t<=K1+1||G1<I)k.current?.()},[$,j,S,c,G]),!$)return K`<div class="loading"><div class="spinner"></div></div>`;if($.length===0)return K`
            <div class="timeline" ref=${G}>
                <div class="timeline-content">
                    <div style="padding: var(--spacing-xl); text-align: center; color: var(--text-secondary)">
                        ${V||"No messages yet. Start a conversation!"}
                    </div>
                </div>
            </div>
        `;let r=$.slice().sort((s,t)=>s.id-t.id),f=(s)=>{let t=s?.data?.thread_id;if(t===null||t===void 0||t==="")return null;let K1=Number(t);return Number.isFinite(K1)?K1:null},_1=new Map;for(let s=0;s<r.length;s+=1){let t=r[s],K1=Number(t?.id),G1=f(t);if(G1!==null){let I=_1.get(G1)||{anchorIndex:-1,replyIndexes:[]};I.replyIndexes.push(s),_1.set(G1,I)}else if(Number.isFinite(K1)){let I=_1.get(K1)||{anchorIndex:-1,replyIndexes:[]};I.anchorIndex=s,_1.set(K1,I)}}let Y1=new Map;for(let[s,t]of _1.entries()){let K1=new Set;if(t.anchorIndex>=0)K1.add(t.anchorIndex);for(let G1 of t.replyIndexes)K1.add(G1);Y1.set(s,Array.from(K1).sort((G1,I)=>G1-I))}let k1=r.map((s,t)=>{let K1=f(s);if(K1===null)return{hasThreadPrev:!1,hasThreadNext:!1};let G1=Y1.get(K1);if(!G1||G1.length===0)return{hasThreadPrev:!1,hasThreadNext:!1};let I=G1.indexOf(t);if(I<0)return{hasThreadPrev:!1,hasThreadNext:!1};return{hasThreadPrev:I>0,hasThreadNext:I<G1.length-1}}),O1=K`<div class="timeline-sentinel" ref=${O}></div>`;return K`
        <div class="timeline ${c?"reverse":"normal"}" ref=${G} onScroll=${U1}>
            <div class="timeline-content">
                ${c?O1:null}
                ${r.map((s,t)=>{let K1=Boolean(s.data?.thread_id&&s.data.thread_id!==s.id),G1=l?.has?.(s.id),I=k1[t]||{};return K`
                    <${x6}
                        key=${s.id}
                        post=${s}
                        isThreadReply=${K1}
                        isThreadPrev=${I.hasThreadPrev}
                        isThreadNext=${I.hasThreadNext}
                        isRemoving=${G1}
                        highlightQuery=${w}
                        agentName=${U6(s.data?.agent_id,L||{})}
                        agentAvatarUrl=${X6(s.data?.agent_id,L||{})}
                        userName=${C?.name||C?.user_name}
                        userAvatarUrl=${C?.avatar_url||C?.avatarUrl||C?.avatar}
                        userAvatarBackground=${C?.avatar_background||C?.avatarBackground}
                        onClick=${()=>B?.(s)}
                        onHashtagClick=${_}
                        onMessageRef=${Q}
                        onScrollToMessage=${U}
                        onFileRef=${X}
                        onOpenWidget=${J}
                        onDelete=${b}
                        onOpenAttachmentPreview=${W}
                    />
                `})}
                ${c?null:O1}
            </div>
        </div>
    `}function K5($){return String($||"").toLowerCase().replace(/^@/,"").replace(/\s+/g," ").trim()}function Hj($,j){let q=K5($),B=K5(j);if(!B)return!1;return q.startsWith(B)||q.includes(B)}function L3($){if(!$)return!1;if($.isComposing)return!1;if($.ctrlKey||$.metaKey||$.altKey)return!1;return typeof $.key==="string"&&$.key.length===1&&/\S/.test($.key)}function N3($,j,q=Date.now(),B=700){let _=$&&typeof $==="object"?$:{value:"",updatedAt:0},Q=String(j||"").trim().toLowerCase();if(!Q)return{value:"",updatedAt:q};return{value:!_.value||!Number.isFinite(_.updatedAt)||q-_.updatedAt>B?Q:`${_.value}${Q}`,updatedAt:q}}function Yj($,j){let q=Math.max(0,Number($)||0);if(q<=0)return[];let _=((Number.isInteger(j)?j:0)%q+q)%q,Q=[];for(let U=0;U<q;U+=1)Q.push((_+U)%q);return Q}function Aj($,j,q=0,B=(_)=>_){let _=K5(j);if(!_)return-1;let Q=Array.isArray($)?$:[],U=Yj(Q.length,q),X=Q.map((J)=>K5(B(J)));for(let J of U)if(X[J].startsWith(_))return J;for(let J of U)if(X[J].includes(_))return J;return-1}function F3($,j,q=-1,B=(_)=>_){let _=Array.isArray($)?$:[];if(q>=0&&q<_.length){let Q=B(_[q]);if(Hj(Q,j))return q}return Aj(_,j,0,B)}function z5($){return String($||"").trim().toLowerCase()}function H3($){let j=String($||"").match(/^@([a-zA-Z0-9_-]*)$/);if(!j)return null;return z5(j[1]||"")}function Dj($){let j=new Set,q=[];for(let B of Array.isArray($)?$:[]){let _=z5(B?.agent_name);if(!_||j.has(_))continue;j.add(_),q.push(B)}return q}function b6($,j,q={}){let B=H3(j);if(B==null)return[];let _=typeof q?.currentChatJid==="string"?q.currentChatJid:null;return Dj($).filter((Q)=>{if(_&&Q?.chat_jid===_)return!1;return z5(Q?.agent_name).startsWith(B)})}function Y3($){let j=z5($);return j?`@${j} `:""}function u6($,j,q={}){if(!$||$.isComposing)return!1;if(q.searchMode)return!1;if(!q.showSessionSwitcherButton)return!1;if($.ctrlKey||$.metaKey||$.altKey)return!1;if($.key!=="@")return!1;return String(j||"")===""}function m6($){let j=Cj($);return j?`@${j}`:""}function Cj($){return String($||"").trim().toLowerCase().replace(/[^a-z0-9_-]+/g,"-").replace(/^-+|-+$/g,"").replace(/-{2,}/g,"-")}function g6($,j){let q=typeof $?.agent_name==="string"&&$.agent_name.trim()?m6($.agent_name):String(j||"").trim(),B=typeof $?.chat_jid==="string"&&$.chat_jid.trim()?$.chat_jid.trim():String(j||"").trim();return`${q} — ${B} • current branch`}function Ij($,j={}){let q=[],B=typeof j.currentChatJid==="string"?j.currentChatJid.trim():"",_=typeof $?.chat_jid==="string"?$.chat_jid.trim():"";if(B&&_===B)q.push("current");if($?.archived_at)q.push("archived");else if($?.is_active)q.push("active");return q}function p6($,j={}){let q=m6($?.agent_name)||String($?.chat_jid||"").trim(),B=typeof $?.chat_jid==="string"&&$.chat_jid.trim()?$.chat_jid.trim():"unknown-chat",_=Ij($,j);return _.length>0?`${q} — ${B} • ${_.join(" • ")}`:`${q} — ${B}`}function G5({steerQueued:$=!1,pulsing:j=!1}={}){let q=["turn-dot"];if($)q.push("turn-dot-queued");if(j)q.push("turn-dot-pulsing");return q.join(" ")}function A3({pulsing:$=!1}={}){let j=["compose-inline-status-dot"];if($)j.push("compose-inline-status-dot-pulsing");return j.join(" ")}function Z5($,{isLastActivity:j=!1,pendingRequest:q=!1}={}){if(q)return"dot";if(j)return"none";if($?.type==="error")return"none";if($?.type==="intent")return"dot";let B=typeof $?.type==="string"?$.type:"";if(Boolean(typeof $?.tool_name==="string"&&$.tool_name.trim()||$?.tool_args))return"spinner";if(B==="tool_call"||B==="tool_status"||B==="thinking"||B==="waiting")return"spinner";return"dot"}function h6($,j={}){return Z5($,j)==="dot"}var c6=350;function Oj($){return String($||"Connecting").replace(/[-_]+/g," ").replace(/^./,(j)=>j.toUpperCase())}function Tj($,j={}){let q=typeof $==="string"&&$.trim()?$.trim():"connecting";if(q==="connected")return{show:!1,statusClass:"connected",label:"Connected",title:"Connection: Connected"};if(q!=="disconnected"){let X=Oj(q);return{show:!0,statusClass:q,label:X,title:`Connection: ${X}`}}let B=Number.isFinite(Number(j?.delayMs))?Math.max(0,Number(j.delayMs)):c6,_=Number.isFinite(Number(j?.nowMs))?Number(j.nowMs):Date.now(),Q=Number.isFinite(Number(j?.disconnectedAtMs))?Number(j.disconnectedAtMs):_;return _-Q>=B?{show:!0,statusClass:"disconnected",label:"Reconnecting",title:"Reconnecting"}:{show:!0,statusClass:"connecting",label:"Connecting",title:"Connecting"}}function D3($,j={}){let q=Number.isFinite(Number(j?.delayMs))?Math.max(0,Number(j.delayMs)):c6,[B,_]=y(null),[Q,U]=y(()=>Date.now());return E(()=>{if($==="disconnected"){let X=Date.now();_((J)=>J??X),U(X);return}_(null),U(Date.now())},[$]),E(()=>{if($!=="disconnected"||B===null)return;let X=q-(Date.now()-B);if(X<=0)return;let J=setTimeout(()=>{U(Date.now())},X);return()=>clearTimeout(J)},[$,B,q]),B1(()=>Tj($,{delayMs:q,disconnectedAtMs:B,nowMs:Q}),[$,q,B,Q])}async function C3($,j,q){if(typeof $!=="function")return!1;try{let B=await $(j);if(!B)return!1;return q(B),!0}catch(B){return!1}}var Mj=[{name:"/model",description:"Select model or list available models"},{name:"/cycle-model",description:"Cycle to the next available model"},{name:"/thinking",description:"Show or set thinking/effort level"},{name:"/effort",description:"Show or set thinking/effort level (alias for /thinking)"},{name:"/cycle-thinking",description:"Cycle thinking level"},{name:"/theme",description:"Set UI theme (no name to show available themes)"},{name:"/meters",description:"Toggle the top-right CPU/RAM HUD (/meters on|off|toggle)"},{name:"/tint",description:"Tint default light/dark UI (usage: /tint #hex or /tint off)"},{name:"/btw",description:"Open a side conversation panel without interrupting the main chat"},{name:"/state",description:"Show current session state"},{name:"/stats",description:"Show session token and cost stats"},{name:"/context",description:"Show context window usage"},{name:"/last",description:"Show last assistant response"},{name:"/compact",description:"Manually compact the session"},{name:"/auto-compact",description:"Toggle auto-compaction"},{name:"/auto-retry",description:"Toggle auto-retry"},{name:"/abort",description:"Abort the current response"},{name:"/abort-retry",description:"Abort retry backoff"},{name:"/abort-bash",description:"Abort running bash command"},{name:"/shell",description:"Run a shell command and return output"},{name:"/bash",description:"Run a shell command and add output to context"},{name:"/queue",description:"Queue a follow-up message (one-at-a-time)"},{name:"/queue-all",description:"Queue a follow-up message (batch all)"},{name:"/steer",description:"Steer the current response"},{name:"/steering-mode",description:"Set steering mode (all|one)"},{name:"/followup-mode",description:"Set follow-up mode (all|one)"},{name:"/session-name",description:"Set or show the session name"},{name:"/new-session",description:"Start a new session"},{name:"/switch-session",description:"Switch to a session file"},{name:"/session-rotate",description:"Rotate the current persisted session into an archived file"},{name:"/clone",description:"Duplicate the current active branch into a new session"},{name:"/fork",description:"Fork from a previous message"},{name:"/forks",description:"List forkable messages"},{name:"/tree",description:"List the session tree"},{name:"/label",description:"Set or clear a label on a tree entry"},{name:"/labels",description:"List labeled entries"},{name:"/agent-name",description:"Set or show the agent display name"},{name:"/agent-avatar",description:"Set or show the agent avatar URL"},{name:"/user-name",description:"Set or show your display name"},{name:"/user-avatar",description:"Set or show your avatar URL"},{name:"/user-github",description:"Set name/avatar from GitHub profile"},{name:"/export-html",description:"Export session to HTML"},{name:"/passkey",description:"Manage passkeys (enrol/list/delete)"},{name:"/totp",description:"Show a TOTP enrolment QR code"},{name:"/qr",description:"Generate a QR code for text or URL"},{name:"/search",description:"Search notes and skills in the workspace"},{name:"/dream",description:"Run Dream memory maintenance over recent days (default 7)"},{name:"/tasks",description:"List scheduled tasks"},{name:"/scheduled",description:"List scheduled tasks"},{name:"/restart",description:"Restart the agent and stop subprocesses"},{name:"/exit",description:"Exit the current piclaw process immediately (Supervisor will restart it)"},{name:"/login",description:"Login to an AI model provider (OAuth or API key)"},{name:"/logout",description:"Logout from an AI model provider"},{name:"/commands",description:"List available commands"},{name:"/skill:",description:"Run a workspace skill (e.g. /skill:visual-artifact-generator, /skill:web-search)"}],l6="piclaw_compose_history";function yj($,j,q=!1){if(q)return{shouldApply:!1,nextToken:j,text:""};if(!$||typeof $!=="object")return{shouldApply:!1,nextToken:j,text:""};let B=typeof $.token==="string"?$.token:"",_=typeof $.text==="string"?$.text:"";if(!B||B===j||!_.trim())return{shouldApply:!1,nextToken:j,text:""};return{shouldApply:!0,nextToken:B,text:_}}function Sj($="web:default"){let j=typeof $==="string"&&$.trim()?$.trim():"web:default";if(j==="web:default")return l6;return`${l6}:${encodeURIComponent(j)}`}function r6($,j){let q=typeof j?.command?.message==="string"?j.command.message.trim():"";if(!j?.ui_only||!q)return null;let B=typeof $==="string"?$.trim():"";if(!B.startsWith("/"))return null;let _=B.split(/\s+/).filter(Boolean),Q=_[0]?.toLowerCase()||"";if(!(_.length>1)&&(Q==="/model"||Q==="/thinking"||Q==="/effort"))return q;return null}function kj($,j,q=!1){if($&&q)return{mode:"compacting",className:"icon-btn send-btn abort-mode compacting-mode",title:"Compacting context — Stop response",ariaLabel:"Compacting context — Stop response",disabled:!1};if($)return{mode:"abort",className:"icon-btn send-btn abort-mode",title:"Stop response",ariaLabel:"Stop response",disabled:!1};return{mode:"send",className:"icon-btn send-btn",title:"Send (Enter)",ariaLabel:"Send message",disabled:!j}}function Ej($){return $==="abort"||$==="compacting"}function fj($,j=0){let q=typeof $?.message==="string"&&$.message.trim()?$.message.trim():null,B=$?.indicator&&typeof $.indicator==="object"?$.indicator:null;if(!q&&!B)return{visible:!1,title:"",indicatorText:null,animateDot:!1};if(B?.mode==="hidden")return{visible:Boolean(q),title:q||"Working…",indicatorText:null,animateDot:!1};if(B?.mode==="custom"&&Array.isArray(B.frames)&&B.frames.length>0){let _=B.frames,Q=Number.isFinite(j)&&j>=0?Math.floor(j)%_.length:0;return{visible:!0,title:q||"Working…",indicatorText:_[Q],animateDot:!1}}return{visible:!0,title:q||"Working…",indicatorText:null,animateDot:!0}}function Pj({usage:$,onCompact:j}){let q=Math.min(100,Math.max(0,$.percent||0)),B=$.tokens,_=$.contextWindow,Q="Compact context",X=`${B!=null?`Context: ${I3(B)} / ${I3(_)} tokens (${q.toFixed(0)}%)`:`Context: ${q.toFixed(0)}%`} — ${"Compact context"}`,J=9,W=2*Math.PI*9,V=q/100*W,G=q>90?"var(--context-red, #ef4444)":q>75?"var(--context-amber, #f59e0b)":"var(--context-green, #22c55e)";return K`
        <button
            class="compose-context-pie icon-btn"
            type="button"
            title=${X}
            aria-label="Compact context"
            onClick=${(L)=>{L.preventDefault(),L.stopPropagation(),j?.()}}
        >
            <svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="12" cy="12" r=${9}
                    fill="none"
                    stroke="var(--context-track, rgba(128,128,128,0.2))"
                    stroke-width="2.5" />
                <circle cx="12" cy="12" r=${9}
                    fill="none"
                    stroke=${G}
                    stroke-width="2.5"
                    stroke-dasharray=${`${V} ${W}`}
                    stroke-linecap="round"
                    transform="rotate(-90 12 12)" />
            </svg>
        </button>
    `}function I3($){if($==null)return"?";if($>=1e6)return($/1e6).toFixed(1)+"M";if($>=1000)return($/1000).toFixed(0)+"K";return String($)}function O3($){let j=Number($);if(!Number.isFinite(j)||j<=0)return"";return`${I3(j)} ctx`}function Rj($,j){let q=typeof $==="string"?$.trim():"",B=O3(j);if(!q)return B;if(!B)return q;return`${q} • ${B}`}function wj($,j="",q=""){let B=typeof $==="string"?$.trim():"";if(B)return B;let _=typeof j==="string"?j.trim():"",Q=typeof q==="string"?q.trim():"";if(_&&Q)return`${_}/${Q}`;return Q||_||""}function d6($){let j=Array.isArray($?.model_options)?$.model_options:null,q=Array.isArray($?.models)?$.models:[],B=j&&j.length>0?j:q,_=[];for(let Q of B){if(typeof Q==="string"){let G=Q.trim();if(!G)continue;let L=G.indexOf("/"),C=L>0?G.slice(0,L).trim():"",b=L>0?G.slice(L+1).trim():G;_.push({label:G,provider:C,id:b,name:null,contextWindow:null,reasoning:!1});continue}if(!Q||typeof Q!=="object")continue;let U=typeof Q.provider==="string"?Q.provider.trim():"",X=typeof Q.id==="string"?Q.id.trim():"",J=wj(Q.label,U,X);if(!J)continue;let W=typeof Q.name==="string"&&Q.name.trim()?Q.name.trim():null,V=Number(Q.context_window??Q.contextWindow);_.push({label:J,provider:U,id:X,name:W,contextWindow:Number.isFinite(V)&&V>0?V:null,reasoning:Boolean(Q.reasoning)})}return _.sort((Q,U)=>Q.label.localeCompare(U.label,void 0,{sensitivity:"base"})),_}function xj($){if(!$||typeof $!=="object")return"";return[$.label,$.provider,$.id,$.name,O3($.contextWindow)].filter(Boolean).join(" ")}function vj($,j){let q=typeof $==="string"?$.trim():"";if(q)return{showPicker:!0,label:q,hasAvailableModels:!0};let B=d6(j).length>0;return{showPicker:B,label:B?"Select model":"",hasAvailableModels:B}}function bj($){if(!$)return $;let j=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`);if(!j.includes(" @ ")||!j.includes(`:
`))return $;let q=j.split(`
`),B=[],_=0,Q=!1;while(_<q.length){let X=q[_].trim();if(!X){_+=1;continue}if(X==="Messages:"||X.startsWith("Channel:")){Q=!0,_+=1;continue}if(/^[^\n]+\s@\s[^\n]+:$/.test(X)){Q=!0,_+=1;let J=[];while(_<q.length){let W=q[_],V=W.trim();if(/^[^\n]+\s@\s[^\n]+:$/.test(V))break;if(V.startsWith("Channel:")||V==="Messages:")break;J.push(W.startsWith("  ")?W.slice(2):W),_+=1}if(J.length>0)B.push(J.join(`
`).trim());continue}return $}return Q&&B.length>0?B.filter(Boolean).join(`

`):$}function uj($){if(!$)return{content:$,fileRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let W=0;W<q.length;W+=1)if(q[W].trim()==="Files:"&&q[W+1]&&/^\s*-\s+/.test(q[W+1])){B=W;break}if(B===-1)return{content:$,fileRefs:[]};let _=[],Q=B+1;for(;Q<q.length;Q+=1){let W=q[Q];if(/^\s*-\s+/.test(W))_.push(W.replace(/^\s*-\s+/,"").trim());else if(!W.trim())break;else break}if(_.length===0)return{content:$,fileRefs:[]};let U=q.slice(0,B),X=q.slice(Q);return{content:[...U,...X].join(`
`).replace(/\n{3,}/g,`

`).trim(),fileRefs:_}}function mj($){if(!$)return{content:$,messageRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let W=0;W<q.length;W+=1)if(q[W].trim()==="Referenced messages:"&&q[W+1]&&/^\s*-\s+/.test(q[W+1])){B=W;break}if(B===-1)return{content:$,messageRefs:[]};let _=[],Q=B+1;for(;Q<q.length;Q+=1){let W=q[Q];if(/^\s*-\s+/.test(W)){let V=W.replace(/^\s*-\s+/,"").trim().match(/^message:(\S+)$/i);if(V)_.push(V[1])}else if(!W.trim())break;else break}if(_.length===0)return{content:$,messageRefs:[]};let U=q.slice(0,B),X=q.slice(Q);return{content:[...U,...X].join(`
`).replace(/\n{3,}/g,`

`).trim(),messageRefs:_}}function gj($){if(!$)return{content:$,attachmentRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let W=0;W<q.length;W+=1)if(q[W].trim()==="Attachments:"&&q[W+1]&&/^\s*-\s+/.test(q[W+1])){B=W;break}if(B===-1)return{content:$,attachmentRefs:[]};let _=[],Q=B+1;for(;Q<q.length;Q+=1){let W=q[Q];if(/^\s*-\s+/.test(W)){let V=W.replace(/^\s*-\s+/,"").trim(),G=V.match(/^attachment:(\d+)(?:\s*\((.+)\))?$/i);if(G)_.push({id:G[1],label:(G[2]||"").trim()||`attachment:${G[1]}`,raw:V})}else if(!W.trim())break;else break}if(_.length===0)return{content:$,attachmentRefs:[]};let U=q.slice(0,B),X=q.slice(Q);return{content:[...U,...X].join(`
`).replace(/\n{3,}/g,`

`).trim(),attachmentRefs:_}}function pj($){let j=bj($||""),q=uj(j||""),B=mj(q.content||""),_=gj(B.content||"");return{text:_.content||"",fileRefs:q.fileRefs,messageRefs:B.messageRefs,attachmentRefs:_.attachmentRefs}}function T3({items:$=[],onInjectQueuedFollowup:j,onRemoveQueuedFollowup:q,onMoveQueuedFollowup:B,onOpenFilePill:_}){if(!Array.isArray($)||$.length===0)return null;return K`
        <div class="compose-queue-stack">
            ${$.map((Q,U)=>{let X=typeof Q?.content==="string"?Q.content:"",J=pj(X);if(!J.text.trim()&&J.fileRefs.length===0&&J.messageRefs.length===0&&J.attachmentRefs.length===0)return null;let W=U>0,V=U<$.length-1;return K`
                    <div class="compose-queue-stack-item" role="listitem">
                        <div class="compose-queue-stack-content" title=${X}>
                            ${J.text.trim()&&K`<div class="compose-queue-stack-text">${J.text}</div>`}
                            ${(J.messageRefs.length>0||J.fileRefs.length>0||J.attachmentRefs.length>0)&&K`
                                <div class="compose-queue-stack-refs">
                                    ${J.messageRefs.map((G)=>K`
                                        <${w0}
                                            key=${"queue-msg-"+G}
                                            prefix="compose"
                                            label=${"msg:"+G}
                                            title=${"Message reference: "+G}
                                            icon="message"
                                        />
                                    `)}
                                    ${J.fileRefs.map((G)=>{let L=G.split("/").pop()||G;return K`
                                            <${w0}
                                                key=${"queue-file-"+G}
                                                prefix="compose"
                                                label=${L}
                                                title=${G}
                                                onClick=${()=>_?.(G)}
                                            />
                                        `})}
                                    ${J.attachmentRefs.map((G)=>K`
                                        <${w0}
                                            key=${"queue-attachment-"+G.id}
                                            prefix="compose"
                                            label=${G.label}
                                            title=${G.raw}
                                        />
                                    `)}
                                </div>
                            `}
                        </div>
                        <div class="compose-queue-stack-actions" role="group" aria-label="Queued follow-up controls">
                            ${$.length>1&&K`
                                <button
                                    class="compose-queue-stack-move-btn"
                                    type="button"
                                    title="Move up"
                                    aria-label="Move up in queue"
                                    disabled=${!W}
                                    onClick=${()=>W&&B?.(U,U-1)}
                                >
                                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                        <polyline points="18 15 12 9 6 15"></polyline>
                                    </svg>
                                </button>
                                <button
                                    class="compose-queue-stack-move-btn"
                                    type="button"
                                    title="Move down"
                                    aria-label="Move down in queue"
                                    disabled=${!V}
                                    onClick=${()=>V&&B?.(U,U+1)}
                                >
                                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                        <polyline points="6 9 12 15 18 9"></polyline>
                                    </svg>
                                </button>
                            `}
                            <button
                                class="compose-queue-stack-steer-btn"
                                type="button"
                                title="Inject queued follow-up as steer"
                                aria-label="Inject queued follow-up as steer"
                                onClick=${()=>j?.(Q)}
                            >
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M4 20h12a2 2 0 0 0 2-2V8" />
                                    <polyline points="14 12 18 8 22 12" />
                                </svg>
                                <span>Steer</span>
                            </button>
                            <button
                                class="compose-queue-stack-close-btn"
                                type="button"
                                title="Cancel queued message"
                                aria-label="Cancel queued message"
                                onClick=${()=>q?.(Q)}
                            >
                                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                    <line x1="18" y1="6" x2="6" y2="18" />
                                    <line x1="6" y1="6" x2="18" y2="18" />
                                </svg>
                            </button>
                        </div>
                    </div>
                `})}
        </div>
    `}function i6({onPost:$,onFocus:j,searchMode:q,searchScope:B="current",onSearch:_,onSearchScopeChange:Q,onEnterSearch:U,onExitSearch:X,fileRefs:J=[],onRemoveFileRef:W,onClearFileRefs:V,messageRefs:G=[],onRemoveMessageRef:L,onClearMessageRefs:C,activeModel:b=null,agentModelsPayload:c=null,modelUsage:l=null,thinkingLevel:w=null,supportsThinking:S=!1,contextUsage:m=null,onContextCompact:O,notificationsEnabled:d=!1,notificationPermission:g="default",onToggleNotifications:U1,onModelChange:k,onModelStateChange:r,activeEditorPath:f=null,onAttachEditorFile:_1,onOpenFilePill:Y1,followupQueueItems:k1=[],onInjectQueuedFollowup:O1,onRemoveQueuedFollowup:s,onMoveQueuedFollowup:t,onSubmitIntercept:K1,onMessageResponse:G1,onPopOutChat:I,isAgentActive:a=!1,activeChatAgents:F1=[],currentChatJid:X1="web:default",connectionStatus:T1="connected",onSetFileRefs:u1,onSetMessageRefs:A1,onSubmitError:t1,onSwitchChat:M1,onRenameSession:G0,isRenameSessionInProgress:l1=!1,onCreateSession:I0,onDeleteSession:j0,onRestoreSession:g0,showQueueStack:F0=!0,statusNotice:q0=null,extensionWorkingState:B0=null,prefillRequest:x0=null}){let[S1,X0]=y(""),[H0,_0]=y(""),[Q0,h1]=y([]),[Y0,E1]=y(!1),[R1,w1]=y([]),[e1,f1]=y(0),[A0,d1]=y(!1),f0=M(null),[x1,P1]=y([]),[r1,$0]=y(0),[Z0,m1]=y(!1),[j1,H1]=y(!1),[H,p]=y(!1),[v,e]=y(!1),[$1,V1]=y([]),[L1,Z1]=y(0),[N1,A]=y(0),[P,Q1]=y(!1),[z1,o]=y(0),[C1,v1]=y(null),[I1,y1]=y(null),[W0,D1]=y(()=>Date.now()),[S0,M0]=y(0),c1=M(null),d2=M(null),i2=M(null),v0=M(null),W4=M(null),n2=M(null),D0=M(null),p0=M(null),V0=M({value:"",updatedAt:0}),U0=M(0),L0=M(!1),s0=200,b0=Sj(X1),Z2=(z)=>{let F=new Set,T=[];for(let h of z||[]){if(typeof h!=="string")continue;let J1=h.trim();if(!J1||F.has(J1))continue;F.add(J1),T.push(J1)}return T},V2=(z=b0)=>{let F=E0(z);if(!F)return[];try{let T=JSON.parse(F);if(!Array.isArray(T))return[];return Z2(T)}catch{return[]}},J4=(z,F=b0)=>{R0(F,JSON.stringify(z))},P0=M(V2(b0)),h0=M(-1),L2=M(""),s2=M("");E(()=>{P0.current=V2(b0),h0.current=-1,L2.current=""},[b0]),E(()=>{let z=!1;return fetch(`/agent/commands?chat_jid=${encodeURIComponent(X1||"web:default")}`).then((T)=>T.ok?T.json():null).then((T)=>{if(z||!T?.commands)return;f0.current=T.commands.map((h)=>({name:h.name,description:h.description||""}))}).catch((T)=>{console.debug("[compose] failed to fetch dynamic commands",T)}),()=>{z=!0}},[X1]),E(()=>{let z=yj(x0,s2.current,q);if(!z.shouldApply)return;s2.current=z.nextToken,v1(null),X0(z.text),Y4(z.text),t2(z.text),requestAnimationFrame(()=>{d0();let F=c1.current;if(!F)return;try{F.focus({preventScroll:!0})}catch{F.focus()}let T=z.text.length;F.setSelectionRange?.(T,T)})},[x0,q]);let P2=S1.trim()||Q0.length>0||J.length>0||G.length>0,K4=typeof window<"u"&&typeof navigator<"u"&&Boolean(navigator.geolocation)&&Boolean(window.isSecureContext),o0=typeof window<"u"&&typeof Notification<"u",z4=typeof window<"u"?Boolean(window.isSecureContext):!1,F5=o0&&z4&&g!=="denied",G4=g==="granted"&&d,N2=K2(q0),i1=n4(q0),c0=typeof q0?.detail==="string"&&q0.detail.trim()?q0.detail.trim():"",Z4=N2?s4(q0,W0):null,a0=fj(B0,S0),u0=B0?.indicator&&typeof B0.indicator==="object"?B0.indicator:null,H5=G4?"Disable notifications":"Enable notifications",Y5=Q0.length>0||J.length>0||G.length>0,_2=D3(T1),A5=_2.label,D5=_2.title,l0=kj(a,P2,N2),V4=(Array.isArray(F1)?F1:[]).filter((z)=>!z?.archived_at),g1=(()=>{for(let z of Array.isArray(F1)?F1:[]){let F=typeof z?.chat_jid==="string"?z.chat_jid.trim():"";if(F&&F===X1)return z}return null})(),n1=Boolean(g1&&g1.chat_jid===(g1.root_chat_jid||g1.chat_jid)),w2=B1(()=>{let z=new Set,F=[];for(let T of Array.isArray(F1)?F1:[]){let h=typeof T?.chat_jid==="string"?T.chat_jid.trim():"";if(!h||h===X1||z.has(h))continue;if(!(typeof T?.agent_name==="string"?T.agent_name.trim():""))continue;z.add(h),F.push(T)}return F},[F1,X1]),r0=w2.length>0,Q2=r0&&typeof M1==="function",U2=r0&&typeof g0==="function",x2=Boolean(l1||L0.current),t0=!q&&typeof G0==="function"&&!x2,m0=!q&&typeof I0==="function",k0=!q&&typeof j0==="function"&&!n1,X2=!q&&(Q2||U2||t0||m0||k0),L4=vj(b,c),F2=L4.showPicker,H2=L4.label,o2=S&&w?` (${w})`:"",C5=o2.trim()?`${w}`:"",a2=typeof l?.hint_short==="string"?l.hint_short.trim():"",N4=[C5||null,a2||null].filter(Boolean).join(" • "),I5=[b?`Current model: ${H2}${o2}`:null,l?.plan?`Plan: ${l.plan}`:null,a2||null,l?.primary?.reset_description||null,l?.secondary?.reset_description||null].filter(Boolean),F4=j1?"Switching model…":I5.join(" • ")||(F2?"Select a model (tap to open model picker)":`Current model: ${H2}${o2} (tap to open model picker)`),H4=!q&&(F2||m&&m.percent!=null),e0=(z)=>{if(!z||typeof z!=="object")return;let F=z.model??z.current;if(typeof r==="function")r({model:F??null,thinking_level:z.thinking_level??null,thinking_level_label:z.thinking_level_label??null,supports_thinking:z.supports_thinking,provider_usage:z.provider_usage??null});if(F&&typeof k==="function")k(F)},d0=(z)=>{let F=z||c1.current;if(!F)return;F.style.height="auto",F.style.height=`${F.scrollHeight}px`,F.style.overflowY="hidden"},Y4=(z)=>{if(!z.startsWith("/")||z.includes(`
`)){d1(!1),w1([]);return}let F=z.toLowerCase().split(" ")[0];if(F.length<1){d1(!1),w1([]);return}let h=(f0.current||Mj).filter((J1)=>J1.name.startsWith(F)||J1.name.replace(/-/g,"").startsWith(F.replace(/-/g,"")));if(h.length>0&&!(h.length===1&&h[0].name===F))m1(!1),P1([]),w1(h),f1(0),d1(!0);else d1(!1),w1([])},A4=(z)=>{let F=S1,T=F.indexOf(" "),h=T>=0?F.slice(T):"",J1=z.name+h;X0(J1),d1(!1),w1([]),requestAnimationFrame(()=>{let s1=c1.current;if(!s1)return;let z0=J1.length;s1.selectionStart=z0,s1.selectionEnd=z0,s1.focus()})},t2=(z)=>{if(H3(z)==null){m1(!1),P1([]);return}let F=b6(V4,z,{currentChatJid:X1});if(F.length>0&&!(F.length===1&&Y3(F[0].agent_name).trim().toLowerCase()===String(z||"").trim().toLowerCase()))d1(!1),w1([]),P1(F),$0(0),m1(!0);else m1(!1),P1([])},Y2=(z)=>{let F=Y3(z?.agent_name);if(!F)return;X0(F),m1(!1),P1([]),requestAnimationFrame(()=>{let T=c1.current;if(!T)return;let h=F.length;T.selectionStart=h,T.selectionEnd=h,T.focus()})},D4=()=>{if(q||!Q2&&!U2&&!t0&&!m0&&!k0)return!1;return V0.current={value:"",updatedAt:0},p(!1),d1(!1),w1([]),m1(!1),P1([]),e(!0),!0},C4=(z)=>{if(z?.preventDefault?.(),z?.stopPropagation?.(),q||!Q2&&!U2&&!t0&&!m0&&!k0)return;if(v){V0.current={value:"",updatedAt:0},e(!1);return}D4()},I4=(z)=>{let F=typeof z==="string"?z.trim():"";if(e(!1),!F||F===X1){requestAnimationFrame(()=>c1.current?.focus());return}M1?.(F)},v2=async(z)=>{let F=typeof z==="string"?z.trim():"";if(e(!1),!F||typeof g0!=="function"){requestAnimationFrame(()=>c1.current?.focus());return}try{await g0(F)}catch(T){console.warn("Failed to restore session:",T),requestAnimationFrame(()=>c1.current?.focus())}},O5=(z)=>{let T=(Array.isArray(z)?z:[]).findIndex((h)=>!h?.disabled);return T>=0?T:0},J0=B1(()=>{let z=[];for(let F of w2){let T=Boolean(F?.archived_at),h=typeof F?.agent_name==="string"?F.agent_name.trim():"",J1=typeof F?.chat_jid==="string"?F.chat_jid.trim():"";if(!h||!J1)continue;z.push({type:"session",key:`session:${J1}`,label:`@${h} — ${J1}${F?.is_active?" active":""}${T?" archived":""}`,chat:F,disabled:T?!U2:!Q2})}if(m0)z.push({type:"action",key:"action:new",label:"New session",action:"new",disabled:!1});if(t0)z.push({type:"action",key:"action:rename",label:"Rename current session",action:"rename",disabled:x2});if(k0)z.push({type:"action",key:"action:delete",label:"Delete current session",action:"delete",disabled:!1});return z},[w2,U2,Q2,m0,t0,k0,x2]),O4=async(z)=>{if(z?.preventDefault)z.preventDefault();if(z?.stopPropagation)z.stopPropagation();if(typeof G0!=="function"||l1||L0.current)return;L0.current=!0,e(!1);try{await G0()}catch(F){console.warn("Failed to rename session:",F)}finally{L0.current=!1}requestAnimationFrame(()=>c1.current?.focus())},T4=async()=>{if(typeof I0!=="function")return;e(!1);try{await I0()}catch(z){console.warn("Failed to create session:",z)}requestAnimationFrame(()=>c1.current?.focus())},M4=async()=>{if(typeof j0!=="function")return;e(!1);try{await j0(X1)}catch(z){console.warn("Failed to delete session:",z)}requestAnimationFrame(()=>c1.current?.focus())},A2=(z)=>{if(q)_0(z);else X0(z),Y4(z),t2(z);requestAnimationFrame(()=>d0())},T5=(z)=>{let F=q?H0:S1,T=F&&!F.endsWith(`
`)?`
`:"",h=`${F}${T}${z}`.trimStart();A2(h)},M5=(z)=>{let F=z?.command?.model_label;if(F)return F;let T=z?.command?.message;if(typeof T==="string"){let h=T.match(/•\s+([^\n]+?)\s+\(current\)/);if(h?.[1])return h[1].trim()}return null},y4=async(z)=>{if(q||j1)return;v1(null),y1(null),H1(!0);try{let F=await e5("default",z,null,[],null,X1),T=M5(F);return e0({model:T??b??null,thinking_level:F?.command?.thinking_level,thinking_level_label:F?.command?.thinking_level_label,supports_thinking:F?.command?.supports_thinking}),await C3(o4,X1,e0),y1(r6(z,F)),$?.(F),!0}catch(F){return console.error("Failed to switch model:",F),alert("Failed to switch model: "+F.message),!1}finally{H1(!1)}},y5=async()=>{await y4("/cycle-model")},e2=async(z)=>{let F=typeof z==="string"?z:typeof z?.label==="string"?z.label:"";if(!F||j1)return;if(await y4(`/model ${F}`))p(!1)},S5=(z)=>{if(!z||z.disabled)return;if(z.type==="session"){let F=z.chat;if(F?.archived_at)v2(F.chat_jid);else I4(F.chat_jid);return}if(z.type==="action"){if(z.action==="new"){T4();return}if(z.action==="rename"){O4();return}if(z.action==="delete")M4()}},k5=(z)=>{z.preventDefault(),z.stopPropagation(),V0.current={value:"",updatedAt:0},e(!1),p((F)=>!F)},E5=async()=>{if(q)return;O?.(),await Z("/compact",null,{includeMedia:!1,includeFileRefs:!1,includeMessageRefs:!1,clearAfterSubmit:!1,recordHistory:!1})},f5=(z)=>{if(z==="queue"||z==="steer"||z==="auto")return z;return a?"queue":void 0},Z=async(z,F,T={})=>{let{includeMedia:h=!0,includeFileRefs:J1=!0,includeMessageRefs:s1=!0,clearAfterSubmit:z0=!0,recordHistory:T0=!0}=T||{},C2=typeof z==="string"?z:z&&typeof z?.target?.value==="string"?z.target.value:S1,$4=typeof C2==="string"?C2:"";if(!$4.trim()&&(h?Q0.length===0:!0)&&(J1?J.length===0:!0)&&(s1?G.length===0:!0))return;d1(!1),w1([]),m1(!1),P1([]),e(!1),v1(null),y1(null);let S4=h?[...Q0]:[],k4=J1?[...J]:[],E4=s1?[...G]:[],I2=$4.trim();if(T0&&I2){let O2=P0.current,y0=Z2(O2.filter((P5)=>P5!==I2));if(y0.push(I2),y0.length>200)y0.splice(0,y0.length-200);P0.current=y0,J4(y0),h0.current=-1,L2.current=""}let b$=()=>{if(h)h1([...S4]);if(J1)u1?.(k4);if(s1)A1?.(E4);X0(I2),requestAnimationFrame(()=>d0())};if(z0)X0(""),h1([]),V?.(),C?.();(async()=>{try{let O2=await K1?.({content:I2,submitMode:F,fileRefs:k4,messageRefs:E4,mediaFiles:S4});if(O2){$?.(O2);return}let y0=[];for(let T2 of S4){let f4=await R8(T2);y0.push(f4.id)}let P5=k4.length?`Files:
${k4.map((T2)=>`- ${T2}`).join(`
`)}`:"",u$=E4.length?`Referenced messages:
${E4.map((T2)=>`- message:${T2}`).join(`
`)}`:"",m$=y0.length?`Attachments:
${y0.map((T2,f4)=>{let p$=S4[f4]?.name||`attachment-${f4+1}`;return`- attachment:${T2} (${p$})`}).join(`
`)}`:"",g$=[I2,P5,u$,m$].filter(Boolean).join(`

`),W2=await e5("default",g$,null,y0,f5(F),X1);if(G1?.(W2),W2?.command)e0({model:W2.command.model_label??b??null,thinking_level:W2.command.thinking_level,thinking_level_label:W2.command.thinking_level_label,supports_thinking:W2.command.supports_thinking}),await C3(o4,X1,e0);y1(r6(I2,W2)),$?.(W2)}catch(O2){if(z0)b$();let y0=O2?.message||"Failed to send message.";v1(y0),t1?.(y0),console.error("Failed to post:",O2)}})()},N=(z)=>{O1?.(z)},Y=i((z)=>{if(q||!H&&!v||z?.isComposing)return!1;let F=()=>{z.preventDefault?.(),z.stopPropagation?.()},T=()=>{V0.current={value:"",updatedAt:0}};if(z.key==="Escape"){if(F(),T(),H)p(!1);if(v)e(!1);return!0}if(H){if(z.key==="ArrowDown"){if(F(),T(),$1.length>0)Z1((h)=>(h+1)%$1.length);return!0}if(z.key==="ArrowUp"){if(F(),T(),$1.length>0)Z1((h)=>(h-1+$1.length)%$1.length);return!0}if((z.key==="Enter"||z.key==="Tab")&&$1.length>0)return F(),T(),e2($1[Math.max(0,Math.min(L1,$1.length-1))]),!0;if(L3(z)&&$1.length>0){F();let h=N3(V0.current,z.key);V0.current=h;let J1=F3($1,h.value,L1,(s1)=>xj(s1));if(J1>=0)Z1(J1);return!0}}if(v){if(z.key==="ArrowDown"){if(F(),T(),J0.length>0)A((h)=>(h+1)%J0.length);return!0}if(z.key==="ArrowUp"){if(F(),T(),J0.length>0)A((h)=>(h-1+J0.length)%J0.length);return!0}if((z.key==="Enter"||z.key==="Tab")&&J0.length>0)return F(),T(),S5(J0[Math.max(0,Math.min(N1,J0.length-1))]),!0;if(L3(z)&&J0.length>0){F();let h=N3(V0.current,z.key);V0.current=h;let J1=F3(J0,h.value,N1,(s1)=>s1.label);if(J1>=0)A(J1);return!0}}return!1},[q,H,v,$1,L1,J0,N1,e2]),D=(z)=>{if(z.isComposing)return;if(q&&z.key==="Escape"){z.preventDefault(),_0(""),X?.();return}if(Y(z))return;let F=c1.current?.value??(q?H0:S1);if(u6(z,F,{searchMode:q,showSessionSwitcherButton:X2})){z.preventDefault(),D4();return}if(Z0&&x1.length>0){let T=c1.current?.value??(q?H0:S1);if(!String(T||"").match(/^@([a-zA-Z0-9_-]*)$/))m1(!1),P1([]);else{if(z.key==="ArrowDown"){z.preventDefault(),$0((h)=>(h+1)%x1.length);return}if(z.key==="ArrowUp"){z.preventDefault(),$0((h)=>(h-1+x1.length)%x1.length);return}if(z.key==="Tab"||z.key==="Enter"){z.preventDefault(),Y2(x1[r1]);return}if(z.key==="Escape"){z.preventDefault(),m1(!1),P1([]);return}}}if(A0&&R1.length>0){let T=c1.current?.value??(q?H0:S1);if(!String(T||"").startsWith("/"))d1(!1),w1([]);else{if(z.key==="ArrowDown"){z.preventDefault(),f1((h)=>(h+1)%R1.length);return}if(z.key==="ArrowUp"){z.preventDefault(),f1((h)=>(h-1+R1.length)%R1.length);return}if(z.key==="Tab"){z.preventDefault(),A4(R1[e1]);return}if(z.key==="Enter"&&!z.shiftKey){if(!F.includes(" ")){z.preventDefault();let J1=R1[e1];d1(!1),w1([]),Z(J1.name);return}}if(z.key==="Escape"){z.preventDefault(),d1(!1),w1([]);return}}}if(!q&&(z.key==="ArrowUp"||z.key==="ArrowDown")&&!z.metaKey&&!z.ctrlKey&&!z.altKey&&!z.shiftKey){let T=c1.current;if(!T)return;let h=T.value||"",J1=T.selectionStart===0&&T.selectionEnd===0,s1=T.selectionStart===h.length&&T.selectionEnd===h.length;if(z.key==="ArrowUp"&&J1||z.key==="ArrowDown"&&s1){let z0=P0.current;if(!z0.length)return;z.preventDefault();let T0=h0.current;if(z.key==="ArrowUp"){if(T0===-1)L2.current=h,T0=z0.length-1;else if(T0>0)T0-=1;h0.current=T0,A2(z0[T0]||"")}else{if(T0===-1)return;if(T0<z0.length-1)T0+=1,h0.current=T0,A2(z0[T0]||"");else h0.current=-1,A2(L2.current||""),L2.current=""}requestAnimationFrame(()=>{let C2=c1.current;if(!C2)return;let $4=C2.value.length;C2.selectionStart=$4,C2.selectionEnd=$4});return}}if(z.key==="Enter"&&!z.shiftKey&&(z.ctrlKey||z.metaKey)){if(z.preventDefault(),q){if(F.trim())_?.(F.trim(),B)}else Z(F,"steer");return}if(z.key==="Enter"&&!z.shiftKey)if(z.preventDefault(),q){if(F.trim())_?.(F.trim(),B)}else Z(F)},x=(z)=>{let F=Array.from(z||[]).filter((T)=>T instanceof File&&!String(T.name||"").startsWith(".DS_Store"));if(!F.length)return;h1((T)=>[...T,...F]),v1(null)},R=(z)=>{x(z.target.files),z.target.value=""},n=(z)=>{if(q)return;z.preventDefault(),z.stopPropagation(),U0.current+=1,E1(!0)},q1=(z)=>{if(q)return;if(z.preventDefault(),z.stopPropagation(),U0.current=Math.max(0,U0.current-1),U0.current===0)E1(!1)},u=(z)=>{if(q)return;if(z.preventDefault(),z.stopPropagation(),z.dataTransfer)z.dataTransfer.dropEffect="copy";E1(!0)},W1=(z)=>{if(q)return;z.preventDefault(),z.stopPropagation(),U0.current=0,E1(!1),x(z.dataTransfer?.files||[])},K0=(z)=>{if(q)return;let F=z.clipboardData?.items;if(!F||!F.length)return;let T=[];for(let h of F){if(h.kind!=="file")continue;let J1=h.getAsFile?.();if(J1)T.push(J1)}if(T.length>0)z.preventDefault(),x(T)},O0=(z)=>{h1((F)=>F.filter((T,h)=>h!==z))},D2=()=>{v1(null),h1([]),V?.(),C?.()},m3=()=>{if(!navigator.geolocation){alert("Geolocation is not available in this browser.");return}navigator.geolocation.getCurrentPosition((z)=>{let{latitude:F,longitude:T,accuracy:h}=z.coords,J1=`${F.toFixed(5)}, ${T.toFixed(5)}`,s1=Number.isFinite(h)?` ±${Math.round(h)}m`:"",z0=`https://maps.google.com/?q=${F},${T}`,T0=`Location: ${J1}${s1} ${z0}`;T5(T0)},(z)=>{let F=z?.message||"Unable to retrieve location.";alert(`Location error: ${F}`)},{enableHighAccuracy:!0,timeout:1e4,maximumAge:0})};E(()=>{if(!H)return;V0.current={value:"",updatedAt:0},Q1(!0),o4(X1).then((z)=>{V1(d6(z)),e0(z)}).catch((z)=>{console.warn("Failed to load model list:",z),V1([])}).finally(()=>{Q1(!1)})},[H,b]),E(()=>{if(q)p(!1),e(!1),d1(!1),w1([]),m1(!1),P1([])},[q]),E(()=>{if(v&&!X2)e(!1)},[v,X2]),E(()=>{if(!H)return;let z=$1.findIndex((F)=>F?.label===b);Z1(z>=0?z:0)},[H,$1,b]),E(()=>{if(!v)return;A(O5(J0)),V0.current={value:"",updatedAt:0}},[v,X1]),E(()=>{if(!H)return;let z=(F)=>{let T=v0.current,h=W4.current,J1=F.target;if(T&&T.contains(J1))return;if(h&&h.contains(J1))return;p(!1)};return document.addEventListener("pointerdown",z),()=>document.removeEventListener("pointerdown",z)},[H]),E(()=>{if(!v)return;let z=(F)=>{let T=n2.current,h=D0.current,J1=F.target;if(T&&T.contains(J1))return;if(h&&h.contains(J1))return;e(!1)};return document.addEventListener("pointerdown",z),()=>document.removeEventListener("pointerdown",z)},[v]),E(()=>{if(q||!H&&!v)return;let z=(F)=>{Y(F)};return document.addEventListener("keydown",z,!0),()=>document.removeEventListener("keydown",z,!0)},[q,H,v,Y]),E(()=>{if(!H)return;let z=v0.current;z?.focus?.(),z?.querySelector?.(".compose-model-popup-item.active")?.scrollIntoView?.({block:"nearest"})},[H,L1,$1]),E(()=>{if(!v)return;let z=n2.current;z?.focus?.(),z?.querySelector?.(".compose-model-popup-item.active")?.scrollIntoView?.({block:"nearest"})},[v,N1,J0.length]),E(()=>{if(!Z0||!i2.current)return;i2.current.querySelector?.(".slash-item.active")?.scrollIntoView?.({block:"nearest"})},[Z0,r1,x1.length]),E(()=>{if(!A0||!d2.current)return;d2.current.querySelector?.(".slash-item.active")?.scrollIntoView?.({block:"nearest"})},[A0,e1,R1.length]),E(()=>{let z=()=>{let s1=p0.current?.clientWidth||0;o((z0)=>z0===s1?z0:s1)};z();let F=p0.current,T=0,h=()=>{if(T)cancelAnimationFrame(T);T=requestAnimationFrame(()=>{T=0,z()})},J1=null;if(F&&typeof ResizeObserver<"u")J1=new ResizeObserver(()=>h()),J1.observe(F);if(typeof window<"u")window.addEventListener("resize",h);return()=>{if(T)cancelAnimationFrame(T);if(J1?.disconnect?.(),typeof window<"u")window.removeEventListener("resize",h)}},[q,b,g1?.agent_name,X2,m?.percent]);let v$=(z)=>{let F=z.target.value;if(v1(null),y1(null),v)e(!1);d0(z.target),A2(F)};return E(()=>{requestAnimationFrame(()=>d0())},[S1,H0,q]),E(()=>{if(!N2)return;D1(Date.now());let z=setInterval(()=>D1(Date.now()),1000);return()=>clearInterval(z)},[N2,q0?.started_at,q0?.startedAt]),E(()=>{if(M0(0),u0?.mode!=="custom"||!Array.isArray(u0.frames)||u0.frames.length<=1)return;let z=typeof u0.intervalMs==="number"&&Number.isFinite(u0.intervalMs)&&u0.intervalMs>0?u0.intervalMs:120,F=setInterval(()=>{M0((T)=>(T+1)%u0.frames.length)},z);return()=>clearInterval(F)},[u0]),E(()=>{if(q)return;t2(S1)},[V4,X1,S1,q]),K`
        <div class="compose-box">
            ${F0&&!q&&K`
                <${T3}
                    items=${k1}
                    onInjectQueuedFollowup=${N}
                    onRemoveQueuedFollowup=${s}
                    onMoveQueuedFollowup=${t}
                    onOpenFilePill=${Y1}
                />
            `}
            ${a0.visible&&K`
                <div class="compose-inline-status extension-working" role="status" aria-live="polite">
                    <div class="compose-inline-status-row">
                        ${a0.indicatorText?K`<span class="compose-inline-status-glyph" aria-hidden="true">${a0.indicatorText}</span>`:a0.animateDot?K`<span class=${A3({pulsing:!0})} aria-hidden="true"></span>`:null}
                        <span class="compose-inline-status-title">${a0.title}</span>
                    </div>
                </div>
            `}
            ${q0&&K`
                <div
                    class=${`compose-inline-status${N2?" compaction":""}`}
                    role="status"
                    aria-live="polite"
                    title=${c0||""}
                >
                    <div class="compose-inline-status-row">
                        <span class=${A3({pulsing:N2})} aria-hidden="true"></span>
                        <span class="compose-inline-status-title">${i1}</span>
                        ${Z4&&K`<span class="compose-inline-status-elapsed">${Z4}</span>`}
                    </div>
                    ${c0&&K`<div class="compose-inline-status-detail">${c0}</div>`}
                </div>
            `}
            ${I1&&K`
                <div class="compose-inline-status compose-command-notice" role="status" aria-live="polite">
                    <div class="compose-inline-status-detail compose-command-notice-text">${I1}</div>
                </div>
            `}
            <div
                class=${`compose-input-wrapper${Y0?" drag-active":""}`}
                onDragEnter=${n}
                onDragOver=${u}
                onDragLeave=${q1}
                onDrop=${W1}
            >
                <div class="compose-input-main">
                    ${Y5&&K`
                        <div class="compose-file-refs">
                            ${G.map((z)=>{return K`
                                    <${w0}
                                        key=${"msg-"+z}
                                        prefix="compose"
                                        label=${"msg:"+z}
                                        title=${"Message reference: "+z}
                                        removeTitle="Remove reference"
                                        icon="message"
                                        onRemove=${()=>L?.(z)}
                                    />
                                `})}
                            ${J.map((z)=>{let F=z.split("/").pop()||z;return K`
                                    <${w0}
                                        prefix="compose"
                                        label=${F}
                                        title=${z}
                                        onClick=${()=>Y1?.(z)}
                                        removeTitle="Remove file"
                                        onRemove=${()=>W?.(z)}
                                    />
                                `})}
                            ${Q0.map((z,F)=>{let T=z?.name||`attachment-${F+1}`;return K`
                                    <${w0}
                                        key=${T+F}
                                        prefix="compose"
                                        label=${T}
                                        title=${T}
                                        removeTitle="Remove attachment"
                                        onRemove=${()=>O0(F)}
                                    />
                                `})}
                            <button
                                type="button"
                                class="compose-clear-attachments-btn"
                                onClick=${D2}
                                title="Clear all attachments and references"
                                aria-label="Clear all attachments and references"
                            >
                                Clear all
                            </button>
                        </div>
                    `}
                    ${!q&&typeof I==="function"&&K`
                        <button
                            type="button"
                            class="compose-popout-btn"
                            onClick=${()=>I?.()}
                            title="Open this chat in a new chat-only window"
                            aria-label="Open this chat in a new chat-only window"
                        >
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M14 5h5v5" />
                                <path d="M10 14 19 5" />
                                <path d="M19 14v5h-5" />
                                <path d="M5 10V5h5" opacity="0" />
                                <path d="M5 19h5" />
                                <path d="M5 19v-5" />
                            </svg>
                        </button>
                    `}
                    <textarea
                        ref=${c1}
                        placeholder=${q?"Search (Enter to run)...":"Message (Enter to send, Shift+Enter for newline)..."}
                        value=${q?H0:S1}
                        onInput=${v$}
                        onKeyDown=${D}
                        onPaste=${K0}
                        onFocus=${j}
                        onClick=${j}
                        rows="1"
                    />
                    ${Z0&&x1.length>0&&K`
                        <div class="slash-autocomplete" ref=${i2}>
                            ${x1.map((z,F)=>K`
                                <div
                                    key=${z.chat_jid||z.agent_name}
                                    class=${`slash-item${F===r1?" active":""}`}
                                    onMouseDown=${(T)=>{T.preventDefault(),Y2(z)}}
                                    onMouseEnter=${()=>$0(F)}
                                >
                                    <span class="slash-name">@${z.agent_name}</span>
                                    <span class="slash-desc">${z.chat_jid||"Active agent"}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${A0&&R1.length>0&&K`
                        <div class="slash-autocomplete" ref=${d2}>
                            ${R1.map((z,F)=>K`
                                <div
                                    key=${z.name}
                                    class=${`slash-item${F===e1?" active":""}`}
                                    onMouseDown=${(T)=>{T.preventDefault(),A4(z)}}
                                    onMouseEnter=${()=>f1(F)}
                                >
                                    <span class="slash-name">${z.name}</span>
                                    <span class="slash-desc">${z.description}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${H&&!q&&K`
                        <div class="compose-model-popup" ref=${v0} tabIndex="-1" onKeyDown=${Y}>
                            <div class="compose-model-popup-title">Select model</div>
                            <div class="compose-model-popup-menu" role="menu" aria-label="Model picker">
                                ${P&&K`
                                    <div class="compose-model-popup-empty">Loading models…</div>
                                `}
                                ${!P&&$1.length===0&&K`
                                    <div class="compose-model-popup-empty">No models available.</div>
                                `}
                                ${!P&&$1.map((z,F)=>{let T=typeof z?.label==="string"?z.label:"",h=O3(z?.contextWindow);return K`
                                        <button
                                            key=${T}
                                            type="button"
                                            role="menuitem"
                                            class=${`compose-model-popup-item compose-model-popup-model-item${L1===F?" active":""}${b===T?" current-model":""}`}
                                            onClick=${()=>{e2(z)}}
                                            disabled=${j1}
                                            title=${[T,h].filter(Boolean).join(" • ")}
                                        >
                                            <span class="compose-model-popup-model-label">${Rj(T,z?.contextWindow)}</span>
                                        </button>
                                    `})}
                            </div>
                            <div class="compose-model-popup-actions">
                                <button
                                    type="button"
                                    class="compose-model-popup-btn"
                                    onClick=${()=>{y5()}}
                                    disabled=${j1}
                                >
                                    Next model
                                </button>
                            </div>
                        </div>
                    `}
                    ${v&&!q&&K`
                        <div class="compose-model-popup" ref=${n2} tabIndex="-1" onKeyDown=${Y}>
                            <div class="compose-model-popup-title">Manage sessions & agents</div>
                            <div class="compose-model-popup-menu" role="menu" aria-label="Sessions and agents">
                                ${K`
                                    <div class="compose-model-popup-item current" role="note" aria-live="polite">
                                        ${(()=>{return g6(g1,X1)})()}
                                    </div>
                                `}
                                ${!r0&&K`
                                    <div class="compose-model-popup-empty">No other sessions yet.</div>
                                `}
                                ${r0&&w2.map((z,F)=>{let T=Boolean(z.archived_at),J1=z.chat_jid!==(z.root_chat_jid||z.chat_jid)&&!z.is_active&&!T&&typeof j0==="function",s1=p6(z,{currentChatJid:X1});return K`
                                        <div key=${z.chat_jid} class=${`compose-model-popup-item-row${T?" archived":""}`}>
                                            <button
                                                type="button"
                                                role="menuitem"
                                                class=${`compose-model-popup-item${T?" archived":""}${N1===F?" active":""}`}
                                                onClick=${()=>{if(T){v2(z.chat_jid);return}I4(z.chat_jid)}}
                                                disabled=${T?!U2:!Q2}
                                                title=${T?`Restore archived ${`@${z.agent_name}`}`:`Switch to ${`@${z.agent_name}`}`}
                                            >
                                                ${s1}
                                            </button>
                                            ${J1&&K`
                                                <button
                                                    type="button"
                                                    class="compose-model-popup-item-delete"
                                                    title="Delete this branch"
                                                    aria-label=${`Delete @${z.agent_name}`}
                                                    onClick=${(z0)=>{z0.stopPropagation(),e(!1),j0(z.chat_jid)}}
                                                >
                                                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                                        <line x1="18" y1="6" x2="6" y2="18" />
                                                        <line x1="6" y1="6" x2="18" y2="18" />
                                                    </svg>
                                                </button>
                                            `}
                                        </div>
                                    `})}
                            </div>
                            ${(m0||t0||k0)&&K`
                                <div class="compose-model-popup-actions">
                                    ${m0&&K`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn primary${J0.findIndex((z)=>z.key==="action:new")===N1?" active":""}`}
                                            onClick=${()=>{T4()}}
                                            title="Create a new agent/session branch from this chat"
                                        >
                                            New
                                        </button>
                                    `}
                                    ${t0&&K`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn${J0.findIndex((z)=>z.key==="action:rename")===N1?" active":""}`}
                                            onClick=${(z)=>{O4(z)}}
                                            title="Rename the current branch handle"
                                            disabled=${x2}
                                        >
                                            Rename current…
                                        </button>
                                    `}
                                    ${k0&&K`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn danger${J0.findIndex((z)=>z.key==="action:delete")===N1?" active":""}`}
                                            onClick=${()=>{M4()}}
                                            title="Delete (prune) current agent/session branch"
                                        >
                                            Delete current…
                                        </button>
                                    `}
                                </div>
                            `}
                        </div>
                    `}
                </div>
                <div class="compose-footer" ref=${p0}>
                    ${H4&&K`
                    <div class="compose-meta-row">
                        ${F2&&K`
                            <div class="compose-model-meta">
                                <button
                                    ref=${W4}
                                    type="button"
                                    class="compose-model-hint compose-model-hint-btn"
                                    title=${F4}
                                    aria-label="Open model picker"
                                    onClick=${k5}
                                    disabled=${j1}
                                >
                                    ${j1?"Switching…":H2}
                                </button>
                                <div class="compose-model-meta-subline">
                                    ${!j1&&N4&&K`
                                        <span class="compose-model-usage-hint" title=${F4}>
                                            ${N4}
                                        </span>
                                    `}
                                </div>
                            </div>
                        `}
                        ${!q&&m&&m.percent!=null&&K`
                            <${Pj} usage=${m} onCompact=${E5} />
                        `}
                    </div>
                    `}
                    <div class="compose-actions ${q?"search-mode":""}">
                    ${X2&&K`
                        <div
                            ref=${D0}
                            class="compose-session-trigger-group"
                        >
                            ${g1?.agent_name&&K`
                                <button
                                    type="button"
                                    class=${`compose-session-trigger compose-session-trigger-pill${v?" active":""}`}
                                    onClick=${C4}
                                    title=${g1?.chat_jid||X1}
                                    aria-label=${`Manage sessions for @${g1.agent_name}`}
                                    aria-expanded=${v?"true":"false"}
                                >
                                    <span class="compose-current-agent-label active">@${g1.agent_name}</span>
                                </button>
                            `}
                            <button
                                type="button"
                                class=${`compose-session-trigger compose-session-trigger-icon-btn${v?" active":""}`}
                                onClick=${C4}
                                title=${g1?.chat_jid||X1}
                                aria-label=${g1?.agent_name?`Manage sessions for @${g1.agent_name}`:"Manage Sessions/Agents"}
                                aria-expanded=${v?"true":"false"}
                            >
                                <span class="compose-session-trigger-icon" aria-hidden="true">
                                    <svg class="compose-mention-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" focusable="false">
                                        <circle cx="12" cy="12" r="4.25" />
                                        <path d="M16.25 7.75v5.4a2.1 2.1 0 0 0 4.2 0V12a8.45 8.45 0 1 0-4.2 7.33" />
                                    </svg>
                                </span>
                            </button>
                        </div>
                    `}
                    ${q&&K`
                        <label class="compose-search-scope-wrap" title="Search scope">
                            <span class="compose-search-scope-label">Scope</span>
                            <select
                                class="compose-search-scope-select"
                                value=${B}
                                onChange=${(z)=>Q?.(z.currentTarget.value)}
                            >
                                <option value="current">Current</option>
                                <option value="root">Branch family</option>
                                <option value="all">All chats</option>
                            </select>
                        </label>
                    `}
                    <button
                        class="icon-btn search-toggle"
                        onClick=${q?X:U}
                        title=${q?"Close search":"Search"}
                    >
                        ${q?K`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 6L6 18M6 6l12 12"/>
                            </svg>
                        `:K`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="11" cy="11" r="8"/>
                                <path d="M21 21l-4.35-4.35"/>
                            </svg>
                        `}
                    </button>
                    ${K4&&!q&&K`
                        <button
                            class="icon-btn location-btn"
                            onClick=${m3}
                            title="Share location"
                            type="button"
                            disabled=${!1}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="10" />
                                <path d="M12 2a14 14 0 0 1 0 20a14 14 0 0 1 0-20" />
                                <path d="M2 12h20" />
                            </svg>
                        </button>
                    `}
                    ${F5&&!q&&K`
                        <button
                            class=${`icon-btn notification-btn${G4?" active":""}`}
                            onClick=${U1}
                            title=${H5}
                            type="button"
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
                                <path d="M13.73 21a2 2 0 0 1-3.46 0" />
                            </svg>
                        </button>
                    `}
                    ${!q&&K`
                        ${f&&_1&&K`
                            <button
                                class="icon-btn attach-editor-btn"
                                onClick=${_1}
                                title=${`Attach open file: ${f}`}
                                type="button"
                                disabled=${J.includes(f)}
                            >
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
                            </button>
                        `}
                        <label class="icon-btn" title="Attach file">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                            <input type="file" multiple hidden onChange=${R} />
                        </label>
                    `}
                    ${(T1!=="connected"||!q)&&K`
                        <div class="compose-send-stack">
                            ${T1!=="connected"&&K`
                                <span class="compose-connection-status connection-status ${_2.statusClass}" title=${D5}>
                                    ${A5}
                                </span>
                            `}
                            ${!q&&K`
                                <button 
                                    class=${l0.className}
                                    type="button"
                                    onClick=${()=>{if(Ej(l0.mode)){Z("/abort","steer");return}Z()}}
                                    disabled=${l0.disabled}
                                    title=${l0.title}
                                    aria-label=${l0.ariaLabel}
                                >
                                    ${l0.mode==="compacting"?K`
                                            <span class="compose-submit-spinner" aria-hidden="true">
                                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                                                    <circle class="compose-submit-spinner-ring" cx="12" cy="12" r="10.5" stroke-width="2.25" stroke-linecap="round"></circle>
                                                    <rect class="compose-submit-spinner-stop" x="6" y="6" width="12" height="12" rx="0" fill="currentColor"></rect>
                                                </svg>
                                            </span>
                                        `:l0.mode==="abort"?K`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><rect x="4" y="4" width="16" height="16" rx="2.5"/></svg>`:K`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>`}
                                </button>
                            `}
                        </div>
                    `}
                </div>
            </div>
        </div>
        </div>
    `}function M3(...$){for(let j of $)if(typeof j==="string"&&j.trim())return j.trim();return null}function hj($){if($.startsWith('"')&&$.endsWith('"')||$.startsWith("'")&&$.endsWith("'"))return $.slice(1,-1);return $}function n6($){if(typeof $!=="string"||!$.trim())return null;let j=$.match(/^\s*cd\s+(.+?)(?:\s*(?:&&|;|\n))/s);if(!j?.[1])return null;return hj(j[1].trim())||null}function s6($,j){let q=j&&typeof j==="object"?j:null;if(!q)return null;let B=M3(q.cwd,q.working_directory,q.workingDirectory);if(B)return B;let _=M3(q.project_dir,q.projectDir,q.repo_path,q.repoPath);if(_)return _;let Q=M3(q.command),U=n6(Q);if(U)return U;if(Array.isArray(q.commands))for(let X of q.commands){let J=n6(X);if(J)return J}return null}var cj=K`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>
`,lj=K`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <path d="M6 3v12"></path>
        <circle cx="18" cy="6" r="3"></circle>
        <circle cx="6" cy="18" r="3"></circle>
        <path d="M18 9a9 9 0 0 1-9 9"></path>
    </svg>
`,rj=K`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <circle cx="12" cy="12" r="9"></circle>
        <path d="M12 7v5l3 2"></path>
    </svg>
`,dj=1e4;function ij($){return(Array.isArray($)?$:$&&Array.isArray($.status_hints)?$.status_hints:[]).filter((q)=>q&&typeof q==="object").map((q,B)=>({key:typeof q.key==="string"&&q.key.trim()?q.key.trim():`hint-${B}`,iconSvg:typeof q.icon_svg==="string"?q.icon_svg.trim():"",label:typeof q.label==="string"?q.label.trim():"",title:typeof q.title==="string"?q.title.trim():""})).filter((q)=>q.iconSvg&&q.label)}function nj($){if(!($ instanceof Set)||$.size===0)return null;let j=Array.from($.values());for(let q=j.length-1;q>=0;q-=1){let B=j[q];if(B==="thought"||B==="draft")return B}return null}function sj($){if(!Array.isArray($)||$.length===0)return[];let j=new Map([["ssh",0]]);return $.map((q,B)=>({hint:q,index:B})).sort((q,B)=>{let _=j.get(q.hint?.key)??100,Q=j.get(B.hint?.key)??100;if(_!==Q)return _-Q;return q.index-B.index}).map((q)=>q.hint)}function oj($,j){let q=typeof $==="string"?$.trim():"",B=typeof j==="string"?j.trim():"";return[q?q.split(/[\\/]+/).filter(Boolean).pop()||q:"",B].filter(Boolean).join(" • ")}function o6($){if(!$||typeof $!=="object")return!1;let j=typeof $.type==="string"?$.type:"",q=Boolean($.last_activity||$.lastActivity),B=j==="tool_call"||j==="tool_status"||Boolean($.tool_name||$.tool_args);if(!q&&!B)return!1;return t5($)!==null}function a6($){if(!$||typeof $!=="object")return!1;return $.type==="intent"&&i4($)!==null}function t6($,j=Date.now()){if(!Number.isFinite($))return!1;return j-$>=dj}function aj($,j=Date.now()){if(!o6($))return null;let q=t5($);if(q===null||!t6(q,j))return null;let B=y3(new Date(q).toISOString(),j);return B?`${B} ago`:null}function tj($,j=Date.now()){if(!a6($))return null;let q=i4($);if(q===null||!t6(q,j))return null;return s4($,j)}function ej($,j={}){let q=j?.isLastActivity??Boolean($?.last_activity||$?.lastActivity),B=$?.title,_=$?.status,Q="";if($?.type==="plan")Q=B?`Planning: ${B}`:"Planning...";else if($?.type==="tool_call")Q=B?`Running: ${B}`:"Running tool...";else if($?.type==="tool_status")Q=B?`${B}: ${_||"Working..."}`:_||"Working...";else if($?.type==="error")Q=B||"Agent error";else Q=B||_||"Working...";if(!q)return Q;if(Q&&Q!=="Working...")return`Recent activity: ${Q}`;return"Last activity"}function y3($,j=Date.now()){if(!$)return null;let q=j-new Date($).getTime();if(!Number.isFinite(q)||q<0)return null;let B=Math.floor(q/1000),_=Math.floor(B/3600),Q=Math.floor(B%3600/60),U=B%60;if(_>0)return`${_}h ${Q}m`;if(Q>0)return`${Q}m ${U}s`;return`${U}s`}function e6({status:$,draft:j,plan:q,thought:B,pendingRequest:_,intent:Q,extensionPanels:U=[],pendingPanelActions:X=new Set,onExtensionPanelAction:J,turnId:W,steerQueued:V,onPanelToggle:G,showCorePanels:L=!0,showExtensionPanels:C=!0}){let l=(H)=>{if(!H)return{text:"",totalLines:0,fullText:""};if(typeof H==="string"){let $1=H,V1=$1?$1.replace(/\r\n/g,`
`).split(`
`).length:0;return{text:$1,totalLines:V1,fullText:$1}}let p=H.text||"",v=H.fullText||H.full_text||p,e=Number.isFinite(H.totalLines)?H.totalLines:v?v.replace(/\r\n/g,`
`).split(`
`).length:0;return{text:p,totalLines:e,fullText:v}},w=160,S=(H)=>String(H||"").replace(/<\/?internal>/gi,""),m=(H)=>{if(!H)return 1;return Math.max(1,Math.ceil(H.length/160))},O=(H,p,v)=>{let e=(H||"").replace(/\r\n/g,`
`).replace(/\r/g,`
`);if(!e)return{text:"",omitted:0,totalLines:Number.isFinite(v)?v:0,visibleLines:0};let $1=e.split(`
`),V1=$1.length>p?$1.slice(0,p).join(`
`):e,L1=Number.isFinite(v)?v:$1.reduce((A,P)=>A+m(P),0),Z1=V1?V1.split(`
`).reduce((A,P)=>A+m(P),0):0,N1=Math.max(L1-Z1,0);return{text:V1,omitted:N1,totalLines:L1,visibleLines:Z1}},d=l(q),g=l(B),U1=l(j),k=Boolean(d.text)||d.totalLines>0,r=Boolean(g.text)||g.totalLines>0,f=Boolean(U1.fullText?.trim()||U1.text?.trim()),_1=Boolean($||f||k||r||_||Q),Y1=Array.isArray(U)&&U.length>0,[k1,O1]=y(new Set),[s,t]=y(null),[K1,G1]=y(()=>Date.now()),I=(H)=>O1((p)=>{let v=new Set(p),e=!v.has(H);if(e)v.add(H);else v.delete(H);if(typeof G==="function")G(H,e);return v});E(()=>{O1(new Set),t(null)},[W]),E(()=>{if(!(Array.isArray(U)&&U.some((v)=>k1.has(v?.key)&&(v?.started_at||v?.last_activity_at))))return;let p=setInterval(()=>G1(Date.now()),1000);return()=>clearInterval(p)},[k1,U]);let a=B1(()=>nj(k1),[k1]);E(()=>{if(!a||typeof document>"u")return;let H=(p)=>{if(p?.defaultPrevented)return;if(p?.key!=="Escape")return;if(p?.altKey||p?.ctrlKey||p?.metaKey||p?.shiftKey)return;let v=p?.target;if(v instanceof Element){if(v.closest?.('input, textarea, select, [contenteditable="true"]'))return;if(v.isContentEditable)return}if(O1((e)=>{if(!(e instanceof Set)||!e.has(a))return e;let $1=new Set(e);return $1.delete(a),$1}),typeof G==="function")G(a,!1);p.preventDefault?.(),p.stopPropagation?.()};return document.addEventListener("keydown",H),()=>document.removeEventListener("keydown",H)},[a,G]);let F1=K2($),X1=Boolean($?.last_activity||$?.lastActivity),T1=B1(()=>o6($),[$]),u1=B1(()=>a6($),[$]),A1=B1(()=>s6($?.tool_name,$?.tool_args),[$?.tool_name,$?.tool_args]),[t1,M1]=y(null);E(()=>{if(!Boolean(u1||$?.retry_at||$?.retryAt||T1))return;G1(Date.now());let p=setInterval(()=>G1(Date.now()),1000);return()=>clearInterval(p)},[T1,u1,$?.retry_at,$?.retryAt,$?.last_event_at,$?.lastEventAt,$?.started_at,$?.startedAt,$?.type,$?.tool_name,$?.tool_args]),E(()=>{if(!($?.type==="tool_call"||$?.type==="tool_status")||!A1){M1(null);return}let p=!0;return c8(A1).then((v)=>{if(!p)return;if(v?.branch)M1({branch:v.branch,repoPath:v.repo_path||null,path:A1});else M1(null)}).catch(()=>{if(p)M1(null)}),()=>{p=!1}},[$?.type,A1]);let G0=$?.turn_id||W,l1=W6(G0),I0=G5({steerQueued:V}),j0=(H)=>H,g0=h6($,{isLastActivity:X1}),F0=Z5($,{isLastActivity:X1}),q0=Z5(null,{pendingRequest:!0}),B0=(H)=>H==="warning"?"#f59e0b":H==="error"?"var(--danger-color)":H==="success"?"var(--success-color)":l1,x0=Q?.kind||"info",S1=B0(x0),X0=B0($?.kind||(F1?"warning":"info")),H0=ej($,{isLastActivity:X1}),_0=aj($,K1),Q0=t1?.repoPath||"",h1=t1?.branch||"",Y0=t1?oj(Q0,h1):"",E1=ij($?.status_hints||$?.statusHints),R1=B1(()=>sj(E1),[E1]),w1=B1(()=>R1.filter((H)=>H?.key==="ssh"),[R1]),e1=B1(()=>R1.filter((H)=>H?.key!=="ssh"),[R1]);if((!L||!_1)&&(!C||!Y1))return null;let f1=({panelTitle:H,text:p,fullText:v,totalLines:e,maxLines:$1,titleClass:V1,panelKey:L1})=>{let Z1=k1.has(L1),N1=v||p||"",A=L1==="thought"||L1==="draft"?S(N1):N1,P=typeof $1==="number",Q1=Z1&&P,z1=P?O(A,$1,e):{text:A||"",omitted:0,totalLines:Number.isFinite(e)?e:0};if(!A&&!(Number.isFinite(z1.totalLines)&&z1.totalLines>0))return null;let o=`agent-thinking-body${P?" agent-thinking-body-collapsible":""}`,C1=P?`--agent-thinking-collapsed-lines: ${$1};`:"";return K`
            <div
                class="agent-thinking"
                data-expanded=${Z1?"true":"false"}
                data-collapsible=${P?"true":"false"}
                style=${l1?`--turn-color: ${l1};`:""}
            >
                <div class="agent-thinking-title ${V1||""}">
                    ${l1&&K`<span class=${I0} aria-hidden="true"></span>`}
                    ${H}
                    ${Q1&&K`
                        <button
                            class="agent-thinking-close"
                            aria-label=${`Close ${H} panel`}
                            onClick=${()=>I(L1)}
                        >
                            ×
                        </button>
                    `}
                </div>
                <div
                    class=${o}
                    style=${C1}
                    dangerouslySetInnerHTML=${{__html:Q3(A)}}
                />
                ${!Z1&&z1.omitted>0&&K`
                    <button class="agent-thinking-truncation" onClick=${()=>I(L1)}>
                        ▸ ${z1.omitted} more lines
                    </button>
                `}
                ${Z1&&z1.omitted>0&&K`
                    <button class="agent-thinking-truncation" onClick=${()=>I(L1)}>
                        ▴ show less
                    </button>
                `}
            </div>
        `},A0=_?.tool_call?.title,d1=A0?`Awaiting approval: ${A0}`:"Awaiting approval",f0=tj($,K1),x1=(H,p,v=null)=>{let e=n4(H),$1=k8(H,K1),V1=[v,$1].filter(Boolean).join(" · "),L1=G5({steerQueued:V,pulsing:K2(H)||Boolean($1)});return K`
            <div
                class="agent-thinking agent-thinking-intent"
                aria-live="polite"
                style=${p?`--turn-color: ${p};`:""}
                title=${H?.detail||""}
            >
                <div class="agent-thinking-title intent">
                    ${p&&K`<span class=${L1} aria-hidden="true"></span>`}
                    <span class="agent-thinking-title-text">${e}</span>
                    ${V1&&K`<span class="agent-status-elapsed">${V1}</span>`}
                </div>
                ${H.detail&&K`<div class="agent-thinking-body">${H.detail}</div>`}
            </div>
        `},P1=(H,p,v,e,$1,V1,L1,Z1=8,N1=8)=>{let A=Math.max($1-e,0.000000001),P=Math.max(p-Z1*2,1),Q1=Math.max(v-N1*2,1),z1=Math.max(L1-V1,1),o=L1===V1?p/2:Z1+(H.run-V1)/z1*P,C1=N1+(Q1-(H.value-e)/A*Q1);return{x:o,y:C1}},r1=(H,p,v,e,$1,V1,L1,Z1=8,N1=8)=>{if(!Array.isArray(H)||H.length===0)return"";return H.map((A,P)=>{let{x:Q1,y:z1}=P1(A,p,v,e,$1,V1,L1,Z1,N1);return`${P===0?"M":"L"} ${Q1.toFixed(2)} ${z1.toFixed(2)}`}).join(" ")},$0=(H,p="")=>{if(!Number.isFinite(H))return"—";return`${Math.abs(H)>=100?H.toFixed(0):H.toFixed(2).replace(/\.0+$/,"").replace(/(\.\d*[1-9])0+$/,"$1")}${p}`},Z0=["var(--accent-color)","var(--success-color)","var(--warning-color, #f59e0b)","var(--danger-color)"],m1=(H,p)=>{let v=Z0;if(!Array.isArray(v)||v.length===0)return"var(--accent-color)";if(v.length===1||!Number.isFinite(p)||p<=1)return v[0];let $1=Math.max(0,Math.min(Number.isFinite(H)?H:0,p-1))/Math.max(1,p-1)*(v.length-1),V1=Math.floor($1),L1=Math.min(v.length-1,V1+1),Z1=$1-V1,N1=v[V1],A=v[L1];if(!A||V1===L1||Z1<=0.001)return N1;if(Z1>=0.999)return A;let P=Math.round((1-Z1)*1000)/10,Q1=Math.round(Z1*1000)/10;return`color-mix(in oklab, ${N1} ${P}%, ${A} ${Q1}%)`},j1=(H,p="autoresearch")=>{let v=Array.isArray(H)?H.map((o)=>({...o,points:Array.isArray(o?.points)?o.points.filter((C1)=>Number.isFinite(C1?.value)&&Number.isFinite(C1?.run)):[]})).filter((o)=>o.points.length>0):[],e=v.map((o,C1)=>({...o,color:m1(C1,v.length)}));if(e.length===0)return null;let $1=320,V1=120,L1=e.flatMap((o)=>o.points),Z1=L1.map((o)=>o.value),N1=L1.map((o)=>o.run),A=Math.min(...Z1),P=Math.max(...Z1),Q1=Math.min(...N1),z1=Math.max(...N1);return K`
            <div class="agent-series-chart agent-series-chart-combined">
                <div class="agent-series-chart-header">
                    <span class="agent-series-chart-title">Tracked variables</span>
                    <span class="agent-series-chart-value">${e.length} series</span>
                </div>
                <div class="agent-series-chart-plot">
                    <svg class="agent-series-chart-svg" viewBox=${`0 0 ${$1} ${V1}`} preserveAspectRatio="none" aria-hidden="true">
                        ${e.map((o)=>{let C1=o?.key||o?.label||"series",v1=s?.panelKey===p&&s?.seriesKey===C1;return K`
                                <g key=${C1}>
                                    <path
                                        class=${`agent-series-chart-line${v1?" is-hovered":""}`}
                                        d=${r1(o.points,$1,V1,A,P,Q1,z1)}
                                        style=${`--agent-series-color: ${o.color};`}
                                        onMouseEnter=${()=>t({panelKey:p,seriesKey:C1})}
                                        onMouseLeave=${()=>t((I1)=>I1?.panelKey===p&&I1?.seriesKey===C1?null:I1)}
                                    ></path>
                                </g>
                            `})}
                    </svg>
                    <div class="agent-series-chart-points-layer">
                        ${e.flatMap((o)=>{let C1=typeof o?.unit==="string"?o.unit:"",v1=o?.key||o?.label||"series";return o.points.map((I1,y1)=>{let W0=P1(I1,$1,V1,A,P,Q1,z1);return K`
                                    <button
                                        key=${`${v1}-point-${y1}`}
                                        type="button"
                                        class="agent-series-chart-point-hit"
                                        style=${`--agent-series-color: ${o.color}; left:${W0.x/$1*100}%; top:${W0.y/V1*100}%;`}
                                        onMouseEnter=${()=>t({panelKey:p,seriesKey:v1,run:I1.run,value:I1.value,unit:C1})}
                                        onMouseLeave=${()=>t((D1)=>D1?.panelKey===p?null:D1)}
                                        onFocus=${()=>t({panelKey:p,seriesKey:v1,run:I1.run,value:I1.value,unit:C1})}
                                        onBlur=${()=>t((D1)=>D1?.panelKey===p?null:D1)}
                                        aria-label=${`${o?.label||"Series"} ${$0(I1.value,C1)} at run ${I1.run}`}
                                    >
                                        <span class="agent-series-chart-point"></span>
                                    </button>
                                `})})}
                    </div>
                </div>
                <div class="agent-series-legend">
                    ${e.map((o)=>{let C1=o.points[o.points.length-1]?.value,v1=typeof o?.unit==="string"?o.unit:"",I1=o?.key||o?.label||"series",y1=s?.panelKey===p&&s?.seriesKey===I1?s:null,W0=y1&&Number.isFinite(y1.value)?y1.value:C1,D1=y1&&typeof y1.unit==="string"?y1.unit:v1,S0=y1&&Number.isFinite(y1.run)?y1.run:null;return K`
                            <div key=${`${I1}-legend`} class=${`agent-series-legend-item${y1?" is-hovered":""}`} style=${`--agent-series-color: ${o.color};`}>
                                <span class="agent-series-legend-swatch" style=${`--agent-series-color: ${o.color};`}></span>
                                <span class="agent-series-legend-label">${o?.label||"Series"}</span>
                                ${S0!==null&&K`<span class="agent-series-legend-run">run ${S0}</span>`}
                                <span class="agent-series-legend-value">${$0(W0,D1)}</span>
                            </div>
                        `})}
                </div>
            </div>
        `},H1=(H)=>{if(!H)return null;let p=typeof H?.key==="string"?H.key:`panel-${Math.random()}`,v=k1.has(p),e=H?.title||"Extension status",$1=H?.collapsed_text||"",V1=String(H?.state||"").replace(/[-_]+/g," ").replace(/^./,(D1)=>D1.toUpperCase()),L1=B0(H?.state==="completed"?"success":H?.state==="failed"?"error":H?.state==="stopped"?"warning":"info"),Z1=G5({steerQueued:V,pulsing:H?.state==="running"}),N1=typeof H?.detail_markdown==="string"?H.detail_markdown.trim():"",A=typeof H?.last_run_text==="string"?H.last_run_text.trim():"",P=typeof H?.tmux_command==="string"?H.tmux_command.trim():"",Q1=Array.isArray(H?.series)?H.series:[],z1=Array.isArray(H?.actions)?H.actions:[],o=y3(H?.started_at),C1=o?` · ${o}`:"",v1=$1+C1,I1=Boolean(N1||P||o),y1=Boolean(N1||Q1.length>0||P),W0=[e,v1].filter(Boolean).join(" — ");return K`
            <div
                class="agent-thinking agent-thinking-intent agent-thinking-autoresearch"
                aria-live="polite"
                data-expanded=${v?"true":"false"}
                style=${L1?`--turn-color: ${L1};`:""}
                title=${!v?W0||e:""}
            >
                <div class="agent-thinking-header agent-thinking-header-inline">
                    <button
                        class="agent-thinking-title intent agent-thinking-title-clickable"
                        type="button"
                        onClick=${()=>y1?I(p):null}
                    >
                        ${L1&&K`<span class=${Z1} aria-hidden="true"></span>`}
                        <span class="agent-thinking-title-text">${e}</span>
                        ${v1&&K`<span class="agent-thinking-title-meta">${v1}</span>`}
                    </button>
                    ${(z1.length>0||y1)&&K`
                        <div class="agent-thinking-tools-inline">
                            ${z1.length>0&&K`
                                <div class="agent-thinking-actions agent-thinking-actions-inline">
                                    ${z1.map((D1)=>{let S0=`${p}:${D1?.key||""}`,M0=X?.has?.(S0);return K`
                                            <button
                                                key=${S0}
                                                class=${`agent-thinking-action-btn${D1?.tone==="danger"?" danger":""}`}
                                                onClick=${()=>J?.(H,D1)}
                                                disabled=${Boolean(M0)}
                                            >
                                                ${M0?"Working…":D1?.label||"Run"}
                                            </button>
                                        `})}
                                </div>
                            `}
                            ${y1&&K`
                                <button
                                    class="agent-thinking-corner-toggle agent-thinking-corner-toggle-inline"
                                    type="button"
                                    aria-label=${`${v?"Collapse":"Expand"} ${e}`}
                                    title=${v?"Collapse details":"Expand details"}
                                    onClick=${()=>I(p)}
                                >
                                    <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        ${v?K`<polyline points="4 6 8 10 12 6"></polyline>`:K`<polyline points="4 10 8 6 12 10"></polyline>`}
                                    </svg>
                                </button>
                            `}
                        </div>
                    `}
                </div>
                ${v&&K`
                    <div class=${`agent-thinking-autoresearch-layout${I1?"":" chart-only"}`}>
                        ${I1&&K`
                            <div class="agent-thinking-autoresearch-meta-stack">
                                ${o&&K`
                                    <div class="agent-thinking-autoresearch-elapsed">
                                        <span title="Experiment duration">⏱ ${o}</span>
                                        ${H?.last_activity_at&&H?.state==="running"&&K`<span title="Since last activity">⟳ ${y3(H.last_activity_at)} ago</span>`}
                                    </div>
                                `}
                                ${N1&&K`
                                    <div
                                        class="agent-thinking-body agent-thinking-autoresearch-detail"
                                        dangerouslySetInnerHTML=${{__html:Q3(N1)}}
                                    />
                                `}
                                ${P&&K`
                                    <div class="agent-series-chart-command">
                                        <div class="agent-series-chart-command-header">
                                            <span>Attach to session</span>
                                        </div>
                                        <div class="agent-series-chart-command-shell">
                                            <pre class="agent-series-chart-command-code">${P}</pre>
                                            <button
                                                type="button"
                                                class="agent-series-chart-command-copy"
                                                aria-label="Copy tmux command"
                                                title="Copy tmux command"
                                                onClick=${()=>J?.(H,{key:"copy_tmux",action_type:"autoresearch.copy_tmux",label:"Copy tmux"})}
                                            >
                                                ${cj}
                                            </button>
                                        </div>
                                    </div>
                                `}
                            </div>
                        `}
                        ${Q1.length>0?K`
                                <div class="agent-series-chart-stack">
                                    ${j1(Q1,p)}
                                    ${A&&K`<div class="agent-series-chart-note">${A}</div>`}
                                </div>
                            `:K`<div class="agent-thinking-body agent-thinking-autoresearch-summary">Variable history will appear after the first completed run.</div>`}
                    </div>
                `}
            </div>
        `};return K`
        <div class="agent-status-panel">
            ${L&&Q&&x1(Q,S1)}
            ${C&&Array.isArray(U)&&U.map((H)=>H1(H))}
            ${L&&$?.type==="intent"&&x1($,X0,f0)}
            ${L&&_&&K`
                <div class="agent-status agent-status-request" aria-live="polite" style=${l1?`--turn-color: ${l1};`:""}>
                    ${q0==="dot"&&K`<span class=${I0} aria-hidden="true"></span>`}
                    ${q0==="spinner"&&K`<div class="agent-status-spinner"></div>`}
                    <span class="agent-status-text">${d1}</span>
                </div>
            `}
            ${L&&k&&f1({panelTitle:j0("Planning"),text:d.text,fullText:d.fullText,totalLines:d.totalLines,panelKey:"plan"})}
            ${L&&r&&f1({panelTitle:j0("Thoughts"),text:g.text,fullText:g.fullText,totalLines:g.totalLines,maxLines:8,titleClass:"thought",panelKey:"thought"})}
            ${L&&f&&f1({panelTitle:j0("Draft"),text:U1.text,fullText:U1.fullText,totalLines:U1.totalLines,maxLines:8,titleClass:"thought",panelKey:"draft"})}
            ${L&&$&&$?.type!=="intent"&&K`
                <div class=${`agent-status${X1?" agent-status-last-activity":""}${$?.type==="error"?" agent-status-error":""}${Y0||E1.length>0||_0?" agent-status-multiline":""}`} aria-live="polite" style=${l1?`--turn-color: ${l1};`:""}>
                    ${l1&&g0&&K`<span class=${I0} aria-hidden="true"></span>`}
                    ${$?.type==="error"?K`<span class="agent-status-error-icon" aria-hidden="true">⚠</span>`:F0==="spinner"&&K`<div class="agent-status-spinner"></div>`}
                    <div class="agent-status-copy">
                        <span class="agent-status-text">${H0}</span>
                        ${(Y0||R1.length>0||_0)&&K`
                            <span class="agent-status-meta-row">
                                ${w1.map((H)=>K`
                                    <span key=${H.key} class="agent-status-hint-row" title=${H.title||H.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{__html:H.iconSvg}}></span>
                                        <span class="agent-status-hint-label">${H.label}</span>
                                    </span>
                                `)}
                                ${Y0&&K`
                                    <span class="agent-status-git-row" title=${A1||Y0}>
                                        <span class="agent-status-git-icon">${lj}</span>
                                        <span class="agent-status-git-label">
                                            ${Q0&&K`<span class="agent-status-git-part">${Q0}</span>`}
                                            ${Q0&&h1&&K`<span class="agent-status-git-separator" aria-hidden="true">•</span>`}
                                            ${h1&&K`<span class="agent-status-git-part">${h1}</span>`}
                                        </span>
                                    </span>
                                `}
                                ${e1.map((H)=>K`
                                    <span key=${H.key} class="agent-status-hint-row" title=${H.title||H.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{__html:H.iconSvg}}></span>
                                        <span class="agent-status-hint-label">${H.label}</span>
                                    </span>
                                `)}
                                ${_0&&K`
                                    <span class="agent-status-hint-row agent-status-activity-row" title=${`${X1?"Recent activity":"Last event"} ${_0}`}>
                                        <span class="agent-status-hint-icon">${rj}</span>
                                        <span class="agent-status-hint-label">${_0}</span>
                                    </span>
                                `}
                            </span>
                        `}
                    </div>
                </div>
            `}
        </div>
    `}class $${extensions=new Map;register($){this.extensions.set($.id,$)}unregister($){this.extensions.delete($)}resolve($){let j,q=-1/0;for(let B of this.extensions.values()){if(B.placement!=="tabs")continue;if(!B.canHandle)continue;try{let _=B.canHandle($);if(_===!1||_===0)continue;let Q=_===!0?0:typeof _==="number"?_:0;if(Q>q)q=Q,j=B}catch(_){console.warn(`[PaneRegistry] canHandle() error for "${B.id}":`,_)}}return j}list(){return Array.from(this.extensions.values())}getDockPanes(){return Array.from(this.extensions.values()).filter(($)=>$.placement==="dock")}getTabPanes(){return Array.from(this.extensions.values()).filter(($)=>$.placement==="tabs")}get($){return this.extensions.get($)}get size(){return this.extensions.size}}var n0=new $$;var qQ=Symbol();var BQ=new TextDecoder("utf-16le",{fatal:!0});Object.hasOwn=Object.hasOwn||function($,j){return Object.prototype.hasOwnProperty.call($,j)};var qq={Backspace:65288,Tab:65289,Enter:65293,Escape:65307,Insert:65379,Delete:65535,Home:65360,End:65367,PageUp:65365,PageDown:65366,ArrowLeft:65361,ArrowUp:65362,ArrowRight:65363,ArrowDown:65364,Shift:65505,ShiftLeft:65505,ShiftRight:65506,Control:65507,ControlLeft:65507,ControlRight:65508,Alt:65513,AltLeft:65513,AltRight:65514,Meta:65515,MetaLeft:65515,MetaRight:65516,Super:65515,Super_L:65515,Super_R:65516,CapsLock:65509,NumLock:65407,ScrollLock:65300,Pause:65299,PrintScreen:65377,ContextMenu:65383,Menu:65383," ":32};for(let $=1;$<=12;$+=1)qq[`F${$}`]=65470+($-1);var Bq=new Uint8Array(256);for(let $=0;$<256;$+=1){let j=0;for(let q=0;q<8;q+=1)j=j<<1|$>>q&1;Bq[$]=j}var jU=String(Date.now());var KU=String(Date.now());class j${tabs=new Map;activeId=null;mruOrder=[];listeners=new Set;onChange($){return this.listeners.add($),()=>this.listeners.delete($)}notify(){let $=this.getTabs(),j=this.activeId;for(let q of this.listeners)try{q($,j)}catch(B){console.warn("[tab-store] Change listener failed:",B)}}open($,j){let q=this.tabs.get($);if(!q)q={id:$,label:j||$.split("/").pop()||$,path:$,dirty:!1,pinned:!1},this.tabs.set($,q);return this.activate($),q}activate($){if(!this.tabs.has($))return;this.activeId=$,this.mruOrder=[$,...this.mruOrder.filter((j)=>j!==$)],this.notify()}close($){if(!this.tabs.get($))return!1;if(this.tabs.delete($),this.mruOrder=this.mruOrder.filter((q)=>q!==$),this.activeId===$)this.activeId=this.mruOrder[0]||null;return this.notify(),!0}closeOthers($){for(let[j,q]of this.tabs)if(j!==$&&!q.pinned)this.tabs.delete(j),this.mruOrder=this.mruOrder.filter((B)=>B!==j);if(this.activeId&&!this.tabs.has(this.activeId))this.activeId=$;this.notify()}closeAll(){for(let[$,j]of this.tabs)if(!j.pinned)this.tabs.delete($),this.mruOrder=this.mruOrder.filter((q)=>q!==$);if(this.activeId&&!this.tabs.has(this.activeId))this.activeId=this.mruOrder[0]||null;this.notify()}setDirty($,j){let q=this.tabs.get($);if(!q||q.dirty===j)return;q.dirty=j,this.notify()}togglePin($){let j=this.tabs.get($);if(!j)return;j.pinned=!j.pinned,this.notify()}saveViewState($,j){let q=this.tabs.get($);if(q)q.viewState=j}getViewState($){return this.tabs.get($)?.viewState}rename($,j,q){let B=this.tabs.get($);if(!B)return;if(this.tabs.delete($),B.id=j,B.path=j,B.label=q||j.split("/").pop()||j,this.tabs.set(j,B),this.mruOrder=this.mruOrder.map((_)=>_===$?j:_),this.activeId===$)this.activeId=j;this.notify()}getTabs(){return Array.from(this.tabs.values())}getActiveId(){return this.activeId}getActive(){return this.activeId?this.tabs.get(this.activeId)||null:null}get($){return this.tabs.get($)}get size(){return this.tabs.size}hasUnsaved(){for(let $ of this.tabs.values())if($.dirty)return!0;return!1}getDirtyTabs(){return Array.from(this.tabs.values()).filter(($)=>$.dirty)}nextTab(){let $=this.getTabs();if($.length<=1)return;let j=$.findIndex((B)=>B.id===this.activeId),q=$[(j+1)%$.length];this.activate(q.id)}prevTab(){let $=this.getTabs();if($.length<=1)return;let j=$.findIndex((B)=>B.id===this.activeId),q=$[(j-1+$.length)%$.length];this.activate(q.id)}mruSwitch(){if(this.mruOrder.length>1)this.activate(this.mruOrder[1])}}var Qq=new j$;function q$($){try{return $?.focus?.(),$?.select?.(),!0}catch(j){return!1}}var V5="workspaceExplorerScale",Uq=["compact","default","comfortable"],Xq=new Set(Uq),Wq={compact:{indentPx:14},default:{indentPx:16},comfortable:{indentPx:18}};function B$($,j="default"){if(typeof $!=="string")return j;let q=$.trim().toLowerCase();return Xq.has(q)?q:j}function S3(){if(typeof window>"u")return{width:0,isTouch:!1};let $=Number(window.innerWidth)||0,j=Boolean(window.matchMedia?.("(pointer: coarse)")?.matches),q=Boolean(window.matchMedia?.("(hover: none)")?.matches),B=Number(globalThis.navigator?.maxTouchPoints||0)>0;return{width:$,isTouch:j||B&&q}}function Jq($={}){let j=Math.max(0,Number($.width)||0);if(Boolean($.isTouch))return"comfortable";if(j>0&&j<1180)return"comfortable";return"default"}function Kq($,j={}){if(Boolean(j.isTouch)&&$==="compact")return"default";return $}function k3($={}){let j=Jq($),q=$.stored?B$($.stored,j):j;return Kq(q,$)}function _$($){return Wq[B$($)]}function zq($){if(!$||$.kind!=="text")return!1;let j=Number($.size);return!Number.isFinite(j)||j<=262144}function E3($,j){let q=String($||"").trim();if(!q||q.endsWith("/"))return!1;if(typeof j!=="function")return!1;let B=j({path:q,mode:"edit"});if(!B||typeof B!=="object")return!1;return B.id!=="editor"}function Q$($,j,q={}){let B=q.resolvePane;if(E3($,B))return!0;return zq(j)}var Gq=60000,K$=($)=>{if(!$||!$.name)return!1;if($.path===".")return!1;return $.name.startsWith(".")};function Zq($){let j=String($||"").trim();if(!j||j.endsWith("/"))return!1;return E3(j,(q)=>n0.resolve(q))}function z$($,j,q,B=0,_=[]){if(!q&&K$($))return _;if(!$)return _;if(_.push({node:$,depth:B}),$.type==="dir"&&$.children&&j.has($.path))for(let Q of $.children)z$(Q,j,q,B+1,_);return _}function U$($,j,q){if(!$)return"";let B=[],_=(Q)=>{if(!q&&K$(Q))return;if(B.push(Q.type==="dir"?`d:${Q.path}`:`f:${Q.path}`),Q.children&&j?.has(Q.path))for(let U of Q.children)_(U)};return _($),B.join("|")}function w3($,j){if(!j)return null;if(!$)return j;if($.path!==j.path||$.type!==j.type)return j;let q=Array.isArray($.children)?$.children:null,B=Array.isArray(j.children)?j.children:null;if(!B)return $;let _=q?new Map(q.map((X)=>[X?.path,X])):new Map,Q=!q||q.length!==B.length,U=B.map((X)=>{let J=w3(_.get(X.path),X);if(J!==_.get(X.path))Q=!0;return J});return Q?{...j,children:U}:$}function P3($,j,q){if(!$)return $;if($.path===j)return w3($,q);if(!Array.isArray($.children))return $;let B=!1,_=$.children.map((Q)=>{let U=P3(Q,j,q);if(U!==Q)B=!0;return U});return B?{...$,children:_}:$}var G$=4,f3=14,Vq=8,Lq=16;function Z$($){if(!$)return 0;if($.type==="file"){let B=Math.max(0,Number($.size)||0);return $.__bytes=B,B}let j=Array.isArray($.children)?$.children:[],q=0;for(let B of j)q+=Z$(B);return $.__bytes=q,q}function V$($,j=0){let q=Math.max(0,Number($?.__bytes??$?.size??0)),B={name:$?.name||$?.path||".",path:$?.path||".",size:q,children:[]};if(!$||$.type!=="dir"||j>=G$)return B;let _=Array.isArray($.children)?$.children:[],Q=[];for(let X of _){let J=Math.max(0,Number(X?.__bytes??X?.size??0));if(J<=0)continue;if(X.type==="dir")Q.push({kind:"dir",node:X,size:J});else Q.push({kind:"file",name:X.name,path:X.path,size:J})}Q.sort((X,J)=>J.size-X.size);let U=Q;if(Q.length>f3){let X=Q.slice(0,f3-1),J=Q.slice(f3-1),W=J.reduce((V,G)=>V+G.size,0);X.push({kind:"other",name:`+${J.length} more`,path:`${B.path}/[other]`,size:W}),U=X}return B.children=U.map((X)=>{if(X.kind==="dir")return V$(X.node,j+1);return{name:X.name,path:X.path,size:X.size,children:[]}}),B}function X$(){if(typeof window>"u"||typeof document>"u")return!1;let{documentElement:$,body:j}=document,q=$?.getAttribute?.("data-theme")?.toLowerCase?.()||"";if(q==="dark")return!0;if(q==="light")return!1;if($?.classList?.contains("dark")||j?.classList?.contains("dark"))return!0;if($?.classList?.contains("light")||j?.classList?.contains("light"))return!1;return Boolean(window.matchMedia?.("(prefers-color-scheme: dark)")?.matches)}function L$($,j,q){let B=(($+Math.PI/2)*180/Math.PI+360)%360,_=q?Math.max(30,70-j*10):Math.max(34,66-j*8),Q=q?Math.min(70,45+j*5):Math.min(60,42+j*4);return`hsl(${B.toFixed(1)} ${_}% ${Q}%)`}function L5($,j,q,B){return{x:$+q*Math.cos(B),y:j+q*Math.sin(B)}}function x3($,j,q,B,_,Q){let U=Math.PI*2-0.0001,X=Q-_>U?_+U:Q,J=L5($,j,B,_),W=L5($,j,B,X),V=L5($,j,q,X),G=L5($,j,q,_),L=X-_>Math.PI?1:0;return[`M ${J.x.toFixed(3)} ${J.y.toFixed(3)}`,`A ${B} ${B} 0 ${L} 1 ${W.x.toFixed(3)} ${W.y.toFixed(3)}`,`L ${V.x.toFixed(3)} ${V.y.toFixed(3)}`,`A ${q} ${q} 0 ${L} 0 ${G.x.toFixed(3)} ${G.y.toFixed(3)}`,"Z"].join(" ")}var N$={1:[26,46],2:[48,68],3:[70,90],4:[92,112]};function F$($,j,q){let B=[],_=[],Q=Math.max(0,Number(j)||0),U=(X,J,W,V)=>{let G=Array.isArray(X?.children)?X.children:[];if(!G.length)return;let L=Math.max(0,Number(X.size)||0);if(L<=0)return;let C=W-J,b=J;G.forEach((c,l)=>{let w=Math.max(0,Number(c.size)||0);if(w<=0)return;let S=w/L,m=b,O=l===G.length-1?W:b+C*S;if(b=O,O-m<0.003)return;let d=N$[V];if(d){let g=L$(m,V,q);if(B.push({key:c.path,path:c.path,label:c.name,size:w,color:g,depth:V,startAngle:m,endAngle:O,innerRadius:d[0],outerRadius:d[1],d:x3(120,120,d[0],d[1],m,O)}),V===1)_.push({key:c.path,name:c.name,size:w,pct:Q>0?w/Q*100:0,color:g})}if(V<G$)U(c,m,O,V+1)})};return U($,-Math.PI/2,Math.PI*3/2,1),{segments:B,legend:_}}function R3($,j){if(!$||!j)return null;if($.path===j)return $;let q=Array.isArray($.children)?$.children:[];for(let B of q){let _=R3(B,j);if(_)return _}return null}function H$($,j,q,B){if(!q||q<=0)return{segments:[],legend:[]};let _=N$[1];if(!_)return{segments:[],legend:[]};let Q=-Math.PI/2,U=Math.PI*3/2,X=L$(Q,1,B),W=`${j||"."}/[files]`;return{segments:[{key:W,path:W,label:$,size:q,color:X,depth:1,startAngle:Q,endAngle:U,innerRadius:_[0],outerRadius:_[1],d:x3(120,120,_[0],_[1],Q,U)}],legend:[{key:W,name:$,size:q,pct:100,color:X}]}}function W$($,j=!1,q=!1){if(!$)return null;let B=Z$($),_=V$($,0),Q=_.size||B,{segments:U,legend:X}=F$(_,Q,q);if(!U.length&&Q>0){let J=H$("[files]",_.path,Q,q);U=J.segments,X=J.legend}return{root:_,totalSize:Q,segments:U,legend:X,truncated:j,isDarkTheme:q}}function Nq({payload:$}){if(!$)return null;let[j,q]=y(null),[B,_]=y($?.root?.path||"."),[Q,U]=y(()=>[$?.root?.path||"."]),[X,J]=y(!1);E(()=>{let f=$?.root?.path||".";_(f),U([f]),q(null)},[$?.root?.path,$?.totalSize]),E(()=>{if(!B)return;J(!0);let f=setTimeout(()=>J(!1),180);return()=>clearTimeout(f)},[B]);let W=B1(()=>{return R3($.root,B)||$.root},[$?.root,B]),V=W?.size||$.totalSize||0,{segments:G,legend:L}=B1(()=>{let f=F$(W,V,$.isDarkTheme);if(f.segments.length>0)return f;if(V<=0)return f;let _1=W?.children?.length?"Total":"[files]";return H$(_1,W?.path||$?.root?.path||".",V,$.isDarkTheme)},[W,V,$.isDarkTheme,$?.root?.path]),[C,b]=y(G),c=M(new Map),l=M(0);E(()=>{let f=c.current,_1=new Map(G.map((s)=>[s.key,s])),Y1=performance.now(),k1=220,O1=(s)=>{let t=Math.min(1,(s-Y1)/220),K1=t*(2-t),G1=G.map((I)=>{let F1=f.get(I.key)||{startAngle:I.startAngle,endAngle:I.startAngle,innerRadius:I.innerRadius,outerRadius:I.innerRadius},X1=(M1,G0)=>M1+(G0-M1)*K1,T1=X1(F1.startAngle,I.startAngle),u1=X1(F1.endAngle,I.endAngle),A1=X1(F1.innerRadius,I.innerRadius),t1=X1(F1.outerRadius,I.outerRadius);return{...I,d:x3(120,120,A1,t1,T1,u1)}});if(b(G1),t<1)l.current=requestAnimationFrame(O1)};if(l.current)cancelAnimationFrame(l.current);return l.current=requestAnimationFrame(O1),c.current=_1,()=>{if(l.current)cancelAnimationFrame(l.current)}},[G]);let w=C.length?C:G,S=V>0?C0(V):"0 B",m=W?.name||"",d=(m&&m!=="."?m:"Total")||"Total",g=S,U1=Q.length>1,k=(f)=>{if(!f?.path)return;let _1=R3($.root,f.path);if(!_1||!Array.isArray(_1.children)||_1.children.length===0)return;U((Y1)=>[...Y1,_1.path]),_(_1.path),q(null)},r=()=>{if(!U1)return;U((f)=>{let _1=f.slice(0,-1);return _(_1[_1.length-1]||$?.root?.path||"."),_1}),q(null)};return K`
        <div class="workspace-folder-starburst">
            <svg viewBox="0 0 240 240" class=${`workspace-folder-starburst-svg${X?" is-zooming":""}`} role="img"
                aria-label=${`Folder sizes for ${W?.path||$?.root?.path||"."}`}
                data-segments=${w.length}
                data-base-size=${V}>
                ${w.map((f)=>K`
                    <path
                        key=${f.key}
                        d=${f.d}
                        fill=${f.color}
                        stroke="var(--bg-primary)"
                        stroke-width="1"
                        class=${`workspace-folder-starburst-segment${j?.key===f.key?" is-hovered":""}`}
                        onMouseEnter=${()=>q(f)}
                        onMouseLeave=${()=>q(null)}
                        onTouchStart=${()=>q(f)}
                        onTouchEnd=${()=>q(null)}
                        onClick=${()=>k(f)}
                    >
                        <title>${f.label} — ${C0(f.size)}</title>
                    </path>
                `)}
                <g
                    class=${`workspace-folder-starburst-center-hit${U1?" is-drill":""}`}
                    onClick=${r}
                    role="button"
                    aria-label="Zoom out"
                >
                    <circle
                        cx="120"
                        cy="120"
                        r="24"
                        fill="var(--bg-secondary)"
                        stroke="var(--border-color)"
                        stroke-width="1"
                        class="workspace-folder-starburst-center"
                    />
                    <text x="120" y="114" text-anchor="middle" class="workspace-folder-starburst-total-label">${d}</text>
                    <text x="120" y="130" text-anchor="middle" class="workspace-folder-starburst-total-value">${g}</text>
                </g>
            </svg>
            ${L.length>0&&K`
                <div class="workspace-folder-starburst-legend">
                    ${L.slice(0,8).map((f)=>K`
                        <div key=${f.key} class="workspace-folder-starburst-legend-item">
                            <span class="workspace-folder-starburst-swatch" style=${`background:${f.color}`}></span>
                            <span class="workspace-folder-starburst-name" title=${f.name}>${f.name}</span>
                            <span class="workspace-folder-starburst-size">${C0(f.size)}</span>
                            <span class="workspace-folder-starburst-pct">${f.pct.toFixed(1)}%</span>
                        </div>
                    `)}
                </div>
            `}
            ${$.truncated&&K`
                <div class="workspace-folder-starburst-note">Preview is truncated by tree depth/entry limits.</div>
            `}
        </div>
    `}function J$($){if(typeof document>"u"||!$)return;let j=document.createElement("a");j.href=$,j.setAttribute("download",""),j.rel="noopener",j.style.display="none",document.body.appendChild(j),j.click(),j.remove()}function Y$($){switch($?.state){case"indexing":return"Indexing workspace…";case"ready":return"Workspace index ready";case"stale":return"Workspace index may be stale";case"failed":return"Workspace index failed";case"never_indexed":return"Workspace index not built yet";default:return"Checking workspace index…"}}function Fq($){if(!$)return"Workspace index status";let j=[Y$($)];if($.last_indexed_at)j.push(`Last indexed: ${$.last_indexed_at}`);if(typeof $.indexed_file_count==="number")j.push(`Indexed files: ${$.indexed_file_count}`);if(Array.isArray($.roots)&&$.roots.length)j.push(`Roots: ${$.roots.join(", ")}`);if($.last_error)j.push(`Error: ${$.last_error}`);return j.join(`
`)}function Hq($){let j=$?.target;if(j&&typeof j==="object")return j;return j?.parentElement||null}function Yq($){return Boolean($?.closest?.(".workspace-node-icon, .workspace-label-text"))}function Aq($,j=null){let q=Hq($),B=q?.closest?.(".workspace-row");if(!B)return null;let _=B.dataset.type,Q=B.dataset.path;if(!Q||Q===".")return null;if(j===Q)return null;let U=$?.touches?.[0];if(!U)return null;return{type:_,path:Q,dragPath:Yq(q)?Q:null,startX:U.clientX,startY:U.clientY}}function A$({onFileSelect:$,visible:j=!0,active:q=void 0,onOpenEditor:B,onOpenTerminalTab:_,onOpenVncTab:Q,onToggleTerminal:U,terminalVisible:X=!1}){let[J,W]=y(null),[V,G]=y(new Set(["."])),[L,C]=y(null),[b,c]=y(null),[l,w]=y(""),[S,m]=y(null),[,O]=y(null),[d,g]=y(!0),[U1,k]=y(!1),[r,f]=y(null),[_1,Y1]=y(()=>Z8("workspaceShowHidden",!1)),[k1,O1]=y(!1),[s,t]=y(null),[K1,G1]=y(null),[I,a]=y(null),[F1,X1]=y(!1),[T1,u1]=y(null),[A1,t1]=y(null),[M1,G0]=y(null),[l1,I0]=y(!1),[j0,g0]=y(()=>X$()),[F0,q0]=y(()=>k3({stored:E0(V5),...S3()})),[B0,x0]=y(!1),S1=M(V),X0=M(""),H0=M(null),_0=M(0),Q0=M(new Set),h1=M(null),Y0=M(null),E1=M(new Map),R1=M($),w1=M(B),e1=M(null),f1=M(null),A0=M(null),d1=M(null),f0=M(null),x1=M(null),P1=M("."),r1=M(0),$0=M({path:null,dragging:!1,startX:0,startY:0}),Z0=M({path:null,dragging:!1,startX:0,startY:0}),m1=M({path:null,timer:0}),j1=M(!1),H1=M(0),H=M(new Map),p=M(null),v=M(null),e=M(null),$1=M(null),V1=M(null),L1=M(null),Z1=M(_1),N1=M(j),A=M(q??j),P=M(0),Q1=M(I),z1=M(k1),o=M(s),C1=M(null),v1=M({x:0,y:0}),I1=M(0),y1=M(null),W0=M(L),D1=M(b),S0=M(null),M0=M(S);R1.current=$,w1.current=B,E(()=>{S1.current=V},[V]),E(()=>{Z1.current=_1},[_1]),E(()=>{N1.current=j},[j]),E(()=>{A.current=q??j},[q,j]),E(()=>{Q1.current=I},[I]);let c1=i(()=>{if(!r1.current)return;clearTimeout(r1.current),r1.current=0},[]);E(()=>()=>c1(),[c1]),E(()=>{if(typeof window>"u")return;let Z=()=>{q0(k3({stored:E0(V5),...S3()}))};Z();let N=()=>Z(),Y=()=>Z(),D=(u)=>{if(!u||u.key===null||u.key===V5)Z()};window.addEventListener("resize",N),window.addEventListener("focus",Y),window.addEventListener("storage",D);let x=window.matchMedia?.("(pointer: coarse)"),R=window.matchMedia?.("(hover: none)"),n=(u,W1)=>{if(!u)return;if(u.addEventListener)u.addEventListener("change",W1);else if(u.addListener)u.addListener(W1)},q1=(u,W1)=>{if(!u)return;if(u.removeEventListener)u.removeEventListener("change",W1);else if(u.removeListener)u.removeListener(W1)};return n(x,N),n(R,N),()=>{window.removeEventListener("resize",N),window.removeEventListener("focus",Y),window.removeEventListener("storage",D),q1(x,N),q1(R,N)}},[]),E(()=>{let Z=(N)=>{let Y=N?.detail?.path;if(!Y)return;let D=Y.split("/"),x=[];for(let R=1;R<D.length;R++)x.push(D.slice(0,R).join("/"));if(x.length)G((R)=>{let n=new Set(R);n.add(".");for(let q1 of x)n.add(q1);return n});C(Y),requestAnimationFrame(()=>{let R=document.querySelector(`[data-path="${CSS.escape(Y)}"]`);if(R)R.scrollIntoView({block:"nearest",behavior:"smooth"})})};return window.addEventListener("workspace-reveal-path",Z),()=>window.removeEventListener("workspace-reveal-path",Z)},[]),E(()=>{z1.current=k1},[k1]),E(()=>{o.current=s},[s]),E(()=>{W0.current=L},[L]),E(()=>{D1.current=b},[b]),E(()=>{M0.current=S},[S]),E(()=>{if(typeof window>"u"||typeof document>"u")return;let Z=()=>g0(X$());Z();let N=window.matchMedia?.("(prefers-color-scheme: dark)"),Y=()=>Z();if(N?.addEventListener)N.addEventListener("change",Y);else if(N?.addListener)N.addListener(Y);let D=typeof MutationObserver<"u"?new MutationObserver(()=>Z()):null;if(D?.observe(document.documentElement,{attributes:!0,attributeFilter:["class","data-theme"]}),document.body)D?.observe(document.body,{attributes:!0,attributeFilter:["class","data-theme"]});return()=>{if(N?.removeEventListener)N.removeEventListener("change",Y);else if(N?.removeListener)N.removeListener(Y);D?.disconnect()}},[]),E(()=>{if(!b)return;let Z=f0.current;if(!Z)return;let N=requestAnimationFrame(()=>{q$(Z)});return()=>cancelAnimationFrame(N)},[b]),E(()=>{if(!B0)return;let Z=(Y)=>{let D=Y?.target;if(!(D instanceof Element))return;if(V1.current?.contains(D))return;if(L1.current?.contains(D))return;x0(!1)},N=(Y)=>{if(Y?.key==="Escape")x0(!1),L1.current?.focus?.()};return document.addEventListener("mousedown",Z),document.addEventListener("touchstart",Z,{passive:!0}),document.addEventListener("keydown",N),()=>{document.removeEventListener("mousedown",Z),document.removeEventListener("touchstart",Z),document.removeEventListener("keydown",N)}},[B0]);let d2=async(Z,N={})=>{let Y=Boolean(N?.autoOpen),D=String(Z||"").trim();k(!0),m(null),O(null);try{let x=await v8(D,20000);if(Y&&D&&Q$(D,x,{resolvePane:(R)=>n0.resolve(R)}))return w1.current?.(D,x),x;return m(x),x}catch(x){let R={error:x.message||"Failed to load preview"};return m(R),R}finally{k(!1)}};e1.current=d2;let i2=i(async()=>{try{let Z=await b8("all");return G0(Z),Z}catch(Z){return console.warn("[workspace-explorer] Failed to load workspace index status:",Z),null}},[]);Y0.current=i2;let v0=i(()=>{Y0.current?.()},[]),W4=async()=>{if(!N1.current)return;try{let Z=await a4("",1,Z1.current),N=U$(Z.root,S1.current,Z1.current);if(N===X0.current){g(!1);return}if(X0.current=N,H0.current=Z.root,!_0.current)_0.current=requestAnimationFrame(()=>{_0.current=0,W((Y)=>w3(Y,H0.current)),g(!1)})}catch(Z){f(Z.message||"Failed to load workspace"),g(!1)}},n2=async(Z)=>{if(!Z)return;if(Q0.current.has(Z))return;Q0.current.add(Z);try{let N=await a4(Z,1,Z1.current);W((Y)=>P3(Y,Z,N.root))}catch(N){f(N.message||"Failed to load workspace")}finally{Q0.current.delete(Z)}};f1.current=n2;let D0=i(()=>{let Z=L;if(!Z)return".";let N=E1.current?.get(Z);if(N&&N.type==="dir")return N.path;if(Z==="."||!Z.includes("/"))return".";let Y=Z.split("/");return Y.pop(),Y.join("/")||"."},[L]),p0=i((Z)=>{let N=Z?.closest?.(".workspace-row");if(!N)return null;let Y=N.dataset.path,D=N.dataset.type;if(!Y)return null;if(D==="dir")return Y;if(Y.includes("/")){let x=Y.split("/");return x.pop(),x.join("/")||"."}return"."},[]),V0=i((Z)=>{return p0(Z?.target||null)},[p0]),U0=i((Z)=>{Q1.current=Z,a(Z)},[]),L0=i(()=>{let Z=m1.current;if(Z?.timer)clearTimeout(Z.timer);m1.current={path:null,timer:0}},[]),s0=i((Z)=>{if(!Z||Z==="."){L0();return}let N=E1.current?.get(Z);if(!N||N.type!=="dir"){L0();return}if(S1.current?.has(Z)){L0();return}if(m1.current?.path===Z)return;L0();let Y=setTimeout(()=>{m1.current={path:null,timer:0},f1.current?.(Z),G((D)=>{let x=new Set(D);return x.add(Z),x})},600);m1.current={path:Z,timer:Y}},[L0]),b0=i((Z,N)=>{if(v1.current={x:Z,y:N},I1.current)return;I1.current=requestAnimationFrame(()=>{I1.current=0;let Y=C1.current;if(!Y)return;let D=v1.current;Y.style.transform=`translate(${D.x+12}px, ${D.y+12}px)`})},[]),Z2=i((Z)=>{if(!Z)return;let Y=(E1.current?.get(Z)?.name||Z.split("/").pop()||Z).trim();if(!Y)return;G1({path:Z,label:Y})},[]),V2=i(()=>{if(G1(null),I1.current)cancelAnimationFrame(I1.current),I1.current=0;if(C1.current)C1.current.style.transform="translate(-9999px, -9999px)"},[]),J4=i((Z)=>{if(!Z)return".";let N=E1.current?.get(Z);if(N&&N.type==="dir")return N.path;if(Z==="."||!Z.includes("/"))return".";let Y=Z.split("/");return Y.pop(),Y.join("/")||"."},[]),P0=i(()=>{c(null),w("")},[]),h0=i((Z)=>{if(!Z)return;let Y=(E1.current?.get(Z)?.name||Z.split("/").pop()||Z).trim();if(!Y||Z===".")return;c(Z),w(Y)},[]),L2=i(async()=>{let Z=D1.current;if(!Z)return;let N=(l||"").trim();if(!N){P0();return}let Y=E1.current?.get(Z),D=(Y?.name||Z.split("/").pop()||Z).trim();if(N===D){P0();return}try{let R=(await g8(Z,N))?.path||Z,n=Z.includes("/")?Z.split("/").slice(0,-1).join("/")||".":".";if(P0(),f(null),window.dispatchEvent(new CustomEvent("workspace-file-renamed",{detail:{oldPath:Z,newPath:R,type:Y?.type||"file"}})),Y?.type==="dir")G((q1)=>{let u=new Set;for(let W1 of q1)if(W1===Z)u.add(R);else if(W1.startsWith(`${Z}/`))u.add(`${R}${W1.slice(Z.length)}`);else u.add(W1);return u});if(C(R),Y?.type==="dir")m(null),k(!1),O(null);else e1.current?.(R);f1.current?.(n),v0()}catch(x){f(x?.message||"Failed to rename file")}},[P0,l,v0]),s2=i(async(Z)=>{let D=Z||".";for(let x=0;x<50;x+=1){let n=`untitled${x===0?"":`-${x}`}.md`;try{let u=(await m8(D,n,""))?.path||(D==="."?n:`${D}/${n}`);if(D&&D!==".")G((W1)=>new Set([...W1,D]));C(u),f(null),f1.current?.(D),e1.current?.(u),v0();return}catch(q1){if(q1?.status===409||q1?.code==="file_exists")continue;f(q1?.message||"Failed to create file");return}}f("Failed to create file (untitled name already in use).")},[]),P2=i((Z)=>{if(Z?.stopPropagation?.(),F1)return;let N=J4(W0.current);s2(N)},[F1,J4,s2]);E(()=>{if(typeof window>"u")return;let Z=(N)=>{let Y=N?.detail?.updates||[];if(!Array.isArray(Y)||Y.length===0)return;W((q1)=>{let u=q1;for(let W1 of Y){if(!W1?.root)continue;if(!u||W1.path==="."||!W1.path)u=W1.root;else u=P3(u,W1.path,W1.root)}if(u)X0.current=U$(u,S1.current,Z1.current);return g(!1),u});let D=W0.current;if(Boolean(D)&&Y.some((q1)=>{let u=q1?.path||"";if(!u||u===".")return!0;return D===u||D.startsWith(`${u}/`)||u.startsWith(`${D}/`)}))H.current.clear();if(v0(),!D||!M0.current)return;let R=E1.current?.get(D);if(R&&R.type==="dir")return;if(Y.some((q1)=>{let u=q1?.path||"";if(!u||u===".")return!0;return D===u||D.startsWith(`${u}/`)}))e1.current?.(D)};return window.addEventListener("workspace-update",Z),()=>window.removeEventListener("workspace-update",Z)},[]),h1.current=W4;let K4=M(()=>{if(typeof window>"u")return;let Z=window.matchMedia("(min-width: 1024px) and (orientation: landscape)"),N=A.current??N1.current,Y=document.visibilityState!=="hidden"&&(N||Z.matches&&N1.current);t4(Y,Z1.current).catch((D)=>{console.debug("[workspace-explorer] Workspace visibility ping failed.",D,{visible:Y,showHidden:Z1.current})})}).current,o0=M(0),z4=M(()=>{if(o0.current)clearTimeout(o0.current);o0.current=setTimeout(()=>{o0.current=0,K4()},250)}).current;E(()=>{if(N1.current)h1.current?.(),Y0.current?.();z4()},[j,q]),E(()=>{h1.current(),Y0.current?.(),K4();let Z=setInterval(()=>{h1.current(),Y0.current?.()},Gq),N=V8("previewHeight",null),Y=Number.isFinite(N)?Math.min(Math.max(N,80),600):280;if(H1.current=Y,A0.current)A0.current.style.setProperty("--preview-height",`${Y}px`);let D=window.matchMedia("(min-width: 1024px) and (orientation: landscape)"),x=()=>z4();if(D.addEventListener)D.addEventListener("change",x);else if(D.addListener)D.addListener(x);return document.addEventListener("visibilitychange",x),()=>{if(clearInterval(Z),_0.current)cancelAnimationFrame(_0.current),_0.current=0;if(D.removeEventListener)D.removeEventListener("change",x);else if(D.removeListener)D.removeListener(x);if(document.removeEventListener("visibilitychange",x),o0.current)clearTimeout(o0.current),o0.current=0;t4(!1,Z1.current).catch((R)=>{console.debug("[workspace-explorer] Workspace visibility teardown ping failed.",R,{showHidden:Z1.current})})}},[]);let R2=B1(()=>z$(J,V,_1),[J,V,_1]),F5=B1(()=>new Map(R2.map((Z)=>[Z.node.path,Z.node])),[R2]),G4=B1(()=>_$(F0),[F0]);E1.current=F5;let i1=(L?E1.current.get(L):null)?.type==="dir";E(()=>{if(!L||!i1){t1(null),p.current=null,v.current=null;return}let Z=L,N=`${_1?"hidden":"visible"}:${L}`,Y=H.current,D=Y.get(N);if(D?.root){Y.delete(N),Y.set(N,D);let n=W$(D.root,Boolean(D.truncated),j0);if(n)p.current=n,v.current=L,t1({loading:!1,error:null,payload:n});return}let x=p.current,R=v.current;t1({loading:!0,error:null,payload:R===L?x:null}),a4(L,Vq,_1).then((n)=>{if(W0.current!==Z)return;let q1={root:n?.root,truncated:Boolean(n?.truncated)};Y.delete(N),Y.set(N,q1);while(Y.size>Lq){let W1=Y.keys().next().value;if(!W1)break;Y.delete(W1)}let u=W$(q1.root,q1.truncated,j0);p.current=u,v.current=L,t1({loading:!1,error:null,payload:u})}).catch((n)=>{if(W0.current!==Z)return;t1({loading:!1,error:n?.message||"Failed to load folder size chart",payload:R===L?x:null})})},[L,i1,_1,j0]);let c0=Boolean(S&&S.kind==="text"&&!i1&&(!S.size||S.size<=262144)),Z4=c0?"Open in editor":S?.size>262144?"File too large to edit":"File is not editable",a0=Boolean(L&&!i1&&Zq(L)),u0=Boolean(L&&L!=="."),H5=Boolean(L&&!i1),Y5=Boolean(L&&!i1),_2=L&&i1?q3(L,_1):null,A5=Y$(M1),D5=Fq(M1),l0=M1?.state||"never_indexed",V4=l0!=="ready",g1=i(()=>x0(!1),[]),n1=i(async(Z)=>{g1();try{await Z?.()}catch(N){console.warn("[workspace-explorer] Header menu action failed:",N)}},[g1]),w2=i(async(Z)=>{Z?.stopPropagation?.(),I0(!0),G0((N)=>({scope:"all",last_indexed_at:N?.last_indexed_at||null,last_error:null,indexed_file_count:N?.indexed_file_count||0,roots:N?.roots||[],updated_at:N?.updated_at||null,state:"indexing"}));try{let N=await u8("all");G0(N),f(null),X0.current="",h1.current?.()}catch(N){let Y=N?.message||"Failed to reindex workspace";G0((D)=>({scope:"all",last_indexed_at:D?.last_indexed_at||null,last_error:Y,indexed_file_count:D?.indexed_file_count||0,roots:D?.roots||[],updated_at:D?.updated_at||null,state:"failed"})),f(Y)}finally{I0(!1)}},[]);E(()=>{let Z=e.current;if($1.current)$1.current.dispose(),$1.current=null;if(!Z)return;if(Z.innerHTML="",!L||i1||!S||S.error)return;let N={path:L,content:typeof S.text==="string"?S.text:void 0,mtime:S.mtime,size:S.size,preview:S,mode:"view"},Y=n0.resolve(N)||n0.get("workspace-preview-default");if(!Y)return;let D=Y.mount(Z,N);return $1.current=D,()=>{if($1.current===D)D.dispose(),$1.current=null;Z.innerHTML=""}},[L,i1,S]);let r0=(Z)=>{let N=Z?.target;if(N instanceof Element)return N;return N?.parentElement||null},Q2=(Z)=>{return Boolean(Z?.closest?.(".workspace-node-icon, .workspace-label-text"))},U2=(Z)=>{if(!Z)return!1;if(Z.closest?.('input, textarea, [contenteditable="true"]'))return!0;return Boolean(Z.isContentEditable)},x2=M((Z)=>{let N=r0(Z),Y=N?.closest?.("[data-path]");if(!Y)return;let D=Y.dataset.path;if(!D||D===".")return;let x=Boolean(N?.closest?.("button"))||Boolean(N?.closest?.("a"))||Boolean(N?.closest?.("input")),R=Boolean(N?.closest?.(".workspace-caret"));if(x||R)return;if(D1.current===D)return;h0(D)}).current,t0=M((Z)=>{if(j1.current){j1.current=!1;return}let N=r0(Z),Y=N?.closest?.("[data-path]");if(d1.current?.focus?.(),!Y)return;let D=Y.dataset.path,x=Y.dataset.type,R=Boolean(N?.closest?.(".workspace-caret")),n=Boolean(N?.closest?.("button"))||Boolean(N?.closest?.("a"))||Boolean(N?.closest?.("input")),q1=W0.current===D,u=D1.current;if(u){if(u===D)return;P0()}if(x==="dir"){if(S0.current=null,C(D),m(null),O(null),k(!1),!S1.current.has(D))f1.current?.(D);if(q1&&!R)return;G((K0)=>{let O0=new Set(K0);if(O0.has(D))O0.delete(D);else O0.add(D);return O0})}else{S0.current=null,C(D);let W1=E1.current.get(D);if(W1)R1.current?.(W1.path,W1);if(!n&&!R)e1.current?.(D)}}).current,m0=M(()=>{X0.current="",h1.current(),Y0.current?.(),Array.from(S1.current||[]).filter((N)=>N&&N!==".").forEach((N)=>f1.current?.(N))}).current,k0=M(()=>{S0.current=null,C(null),m(null),O(null),k(!1)}).current,X2=M(()=>{Y1((Z)=>{let N=!Z;if(typeof window<"u")R0("workspaceShowHidden",String(N));return Z1.current=N,t4(!0,N).catch((D)=>{console.debug("[workspace-explorer] Workspace visibility refresh after toggling hidden files failed.",D,{showHidden:N})}),X0.current="",h1.current?.(),Array.from(S1.current||[]).filter((D)=>D&&D!==".").forEach((D)=>f1.current?.(D)),N})}).current,L4=M((Z)=>{if(r0(Z)?.closest?.("[data-path]"))return;k0()}).current,F2=i(async(Z)=>{if(!Z)return;let N=Z.split("/").pop()||Z;if(!window.confirm(`Delete "${N}"? This cannot be undone.`))return;try{await h8(Z);let D=Z.includes("/")?Z.split("/").slice(0,-1).join("/")||".":".";if(W0.current===Z)k0();f1.current?.(D),f(null),v0()}catch(D){m((x)=>({...x||{},error:D.message||"Failed to delete file"}))}},[k0]),H2=i((Z)=>{let N=d1.current;if(!N||!Z||typeof CSS>"u"||typeof CSS.escape!=="function")return;N.querySelector(`[data-path="${CSS.escape(Z)}"]`)?.scrollIntoView?.({block:"nearest"})},[]),o2=i((Z)=>{let N=r0(Z);if(D1.current||U2(N))return;let Y=R2;if(!Y||Y.length===0)return;let D=L?Y.findIndex((x)=>x.node.path===L):-1;if(Z.key==="ArrowDown"){Z.preventDefault();let x=Math.min(D+1,Y.length-1),R=Y[x];if(!R)return;if(C(R.node.path),R.node.type!=="dir")R1.current?.(R.node.path,R.node),e1.current?.(R.node.path);else m(null),k(!1),O(null);H2(R.node.path);return}if(Z.key==="ArrowUp"){Z.preventDefault();let x=D<=0?0:D-1,R=Y[x];if(!R)return;if(C(R.node.path),R.node.type!=="dir")R1.current?.(R.node.path,R.node),e1.current?.(R.node.path);else m(null),k(!1),O(null);H2(R.node.path);return}if(Z.key==="ArrowRight"&&D>=0){let x=Y[D];if(x?.node?.type==="dir"&&!V.has(x.node.path))Z.preventDefault(),f1.current?.(x.node.path),G((R)=>new Set([...R,x.node.path]));return}if(Z.key==="ArrowLeft"&&D>=0){let x=Y[D];if(x?.node?.type==="dir"&&V.has(x.node.path))Z.preventDefault(),G((R)=>{let n=new Set(R);return n.delete(x.node.path),n});return}if(Z.key==="Enter"&&D>=0){Z.preventDefault();let x=Y[D];if(!x)return;let R=x.node.path;if(x.node.type==="dir"){if(!S1.current.has(R))f1.current?.(R);G((q1)=>{let u=new Set(q1);if(u.has(R))u.delete(R);else u.add(R);return u}),m(null),O(null),k(!1)}else R1.current?.(R,x.node),e1.current?.(R);return}if((Z.key==="Delete"||Z.key==="Backspace")&&D>=0){let x=Y[D];if(!x||x.node.type==="dir")return;Z.preventDefault(),F2(x.node.path);return}if(Z.key==="Escape")Z.preventDefault(),k0()},[k0,F2,V,R2,H2,L]),C5=i((Z)=>{let N=Aq(Z,D1.current);if(!N)return;$0.current={path:N.dragPath,dragging:!1,startX:N.startX,startY:N.startY}},[]),a2=i(()=>{let Z=$0.current;if(Z?.dragging&&Z.path){let N=Q1.current||D0(),Y=y1.current;if(typeof Y==="function")Y(Z.path,N)}$0.current={path:null,dragging:!1,startX:0,startY:0},P.current=0,O1(!1),t(null),U0(null),L0(),V2()},[D0,V2,U0,L0]),N4=i((Z)=>{let N=$0.current,Y=Z?.touches?.[0];if(!Y||!N?.path)return;let D=Math.abs(Y.clientX-N.startX),x=Math.abs(Y.clientY-N.startY),R=D>8||x>8;if(!N.dragging&&R)N.dragging=!0,O1(!0),t("move"),Z2(N.path);if(N.dragging){Z.preventDefault(),b0(Y.clientX,Y.clientY);let n=document.elementFromPoint(Y.clientX,Y.clientY),q1=p0(n)||D0();if(Q1.current!==q1)U0(q1);s0(q1)}},[p0,D0,Z2,b0,U0,s0]),I5=M((Z)=>{Z.preventDefault();let N=A0.current;if(!N)return;let Y=Z.clientY,D=H1.current||280,x=Z.currentTarget;x.classList.add("dragging"),document.body.style.cursor="row-resize",document.body.style.userSelect="none";let R=Y,n=(u)=>{R=u.clientY;let W1=N.clientHeight-80,K0=Math.min(Math.max(D-(u.clientY-Y),80),W1);N.style.setProperty("--preview-height",`${K0}px`),H1.current=K0},q1=()=>{let u=N.clientHeight-80,W1=Math.min(Math.max(D-(R-Y),80),u);H1.current=W1,x.classList.remove("dragging"),document.body.style.cursor="",document.body.style.userSelect="",R0("previewHeight",String(Math.round(W1))),document.removeEventListener("mousemove",n),document.removeEventListener("mouseup",q1)};document.addEventListener("mousemove",n),document.addEventListener("mouseup",q1)}).current,F4=M((Z)=>{Z.preventDefault();let N=A0.current;if(!N)return;let Y=Z.touches[0];if(!Y)return;let D=Y.clientY,x=H1.current||280,R=Z.currentTarget;R.classList.add("dragging"),document.body.style.userSelect="none";let n=(u)=>{let W1=u.touches[0];if(!W1)return;u.preventDefault();let K0=N.clientHeight-80,O0=Math.min(Math.max(x-(W1.clientY-D),80),K0);N.style.setProperty("--preview-height",`${O0}px`),H1.current=O0},q1=()=>{R.classList.remove("dragging"),document.body.style.userSelect="",R0("previewHeight",String(Math.round(H1.current||x))),document.removeEventListener("touchmove",n),document.removeEventListener("touchend",q1),document.removeEventListener("touchcancel",q1)};document.addEventListener("touchmove",n,{passive:!1}),document.addEventListener("touchend",q1),document.addEventListener("touchcancel",q1)}).current,H4=i((Z=L)=>{if(!Z)return;J$(B3(Z))},[L]),e0=async()=>{if(!L||i1)return;await F2(L)},d0=(Z)=>{return Array.from(Z?.dataTransfer?.types||[]).includes("Files")},Y4=i((Z)=>{if(!d0(Z))return;if(Z.preventDefault(),P.current+=1,!z1.current)O1(!0);t("upload");let N=V0(Z)||D0();U0(N),s0(N)},[D0,V0,U0,s0]),A4=i((Z)=>{if(!d0(Z))return;if(Z.preventDefault(),Z.dataTransfer)Z.dataTransfer.dropEffect="copy";if(!z1.current)O1(!0);if(o.current!=="upload")t("upload");let N=V0(Z)||D0();if(Q1.current!==N)U0(N);s0(N)},[D0,V0,U0,s0]),t2=i((Z)=>{if(!d0(Z))return;if(Z.preventDefault(),P.current=Math.max(0,P.current-1),P.current===0)O1(!1),t(null),U0(null),L0()},[U0,L0]),Y2=i(async(Z,N=".")=>{let Y=Array.from(Z||[]);if(Y.length===0)return;let D=N&&N!==""?N:".",x=D!=="."?D:"workspace root";c1(),X1(!0),u1({current:0,total:Y.length,name:"",percent:0,done:!1,error:null});try{let R=null;for(let n=0;n<Y.length;n++){let q1=Y[n],u=q1?.name||`file ${n+1}`;u1((K0)=>({...K0,current:n+1,name:u,percent:0}));let W1=(K0)=>u1((O0)=>({...O0,percent:K0.percent}));try{R=await j3(q1,D,{onProgress:W1})}catch(K0){let O0=K0?.status,D2=K0?.code;if(O0===409||D2==="file_exists"){if(!window.confirm(`"${u}" already exists in ${x}. Overwrite?`))continue;R=await j3(q1,D,{overwrite:!0,onProgress:W1})}else throw K0}}if(R?.path)S0.current=R.path,C(R.path),e1.current?.(R.path);f1.current?.(D),v0(),u1((n)=>({...n,done:!0})),c1(),r1.current=window.setTimeout(()=>{r1.current=0,u1(null)},1500)}catch(R){f(R.message||"Failed to upload file"),u1((n)=>n?{...n,error:R.message||"Upload failed"}:null),c1(),r1.current=window.setTimeout(()=>{r1.current=0,u1(null)},4000)}finally{X1(!1)}},[c1]),D4=i(async(Z,N)=>{if(!Z)return;let Y=E1.current?.get(Z);if(!Y)return;let D=N&&N!==""?N:".",x=Z.includes("/")?Z.split("/").slice(0,-1).join("/")||".":".";if(D===x)return;try{let n=(await p8(Z,D))?.path||Z;if(Y.type==="dir")G((q1)=>{let u=new Set;for(let W1 of q1)if(W1===Z)u.add(n);else if(W1.startsWith(`${Z}/`))u.add(`${n}${W1.slice(Z.length)}`);else u.add(W1);return u});if(C(n),Y.type==="dir")m(null),k(!1),O(null);else e1.current?.(n);f1.current?.(x),f1.current?.(D),v0()}catch(R){f(R?.message||"Failed to move entry")}},[]);y1.current=D4;let C4=i(async(Z)=>{if(!d0(Z))return;Z.preventDefault(),P.current=0,O1(!1),t(null),a(null),L0();let N=Array.from(Z?.dataTransfer?.files||[]);if(N.length===0)return;let Y=Q1.current||V0(Z)||D0();await Y2(N,Y)},[D0,V0,Y2]),I4=i((Z)=>{if(Z?.stopPropagation?.(),F1)return;let N=Z?.currentTarget?.dataset?.uploadTarget||".";P1.current=N,x1.current?.click()},[F1]),v2=i(()=>{if(F1)return;let Z=W0.current,N=Z?E1.current?.get(Z):null;P1.current=N?.type==="dir"?N.path:".",x1.current?.click()},[F1]),O5=i(()=>{n1(()=>P2(null))},[n1,P2]),J0=i(()=>{n1(()=>v2())},[n1,v2]),O4=i(()=>{n1(()=>m0())},[n1,m0]),T4=i(()=>{n1(()=>X2())},[n1,X2]),M4=i(()=>{if(!L||!a0)return;n1(()=>w1.current?.(L,S))},[n1,L,a0,S]),A2=i(()=>{if(!L||!c0)return;n1(()=>w1.current?.(L,S))},[n1,L,c0,S]),T5=i(()=>{if(!L||L===".")return;n1(()=>h0(L))},[n1,L,h0]),M5=i(()=>{if(!L||i1)return;n1(()=>e0())},[n1,L,i1,e0]),y4=i(()=>{if(!L||i1)return;n1(()=>H4())},[n1,L,i1,H4]),y5=i(()=>{if(!_2)return;g1(),J$(_2)},[g1,_2]),e2=i(()=>{g1(),_?.()},[g1,_]),S5=i(()=>{g1(),Q?.()},[g1,Q]),k5=i(()=>{g1(),U?.()},[g1,U]),E5=i((Z)=>{if(!Z||Z.button!==0)return;let N=Z.currentTarget;if(!N||!N.dataset)return;let Y=N.dataset.path;if(!Y||Y===".")return;if(D1.current===Y)return;let D=r0(Z);if(D?.closest?.("button, a, input, .workspace-caret"))return;if(!Q2(D))return;Z.preventDefault(),Z0.current={path:Y,dragging:!1,startX:Z.clientX,startY:Z.clientY};let x=(n)=>{let q1=Z0.current;if(!q1?.path)return;let u=Math.abs(n.clientX-q1.startX),W1=Math.abs(n.clientY-q1.startY),K0=u>4||W1>4;if(!q1.dragging&&K0)q1.dragging=!0,j1.current=!0,O1(!0),t("move"),Z2(q1.path),b0(n.clientX,n.clientY),document.body.style.userSelect="none",document.body.style.cursor="grabbing";if(q1.dragging){n.preventDefault(),b0(n.clientX,n.clientY);let O0=document.elementFromPoint(n.clientX,n.clientY),D2=p0(O0)||D0();if(Q1.current!==D2)U0(D2);s0(D2)}},R=()=>{document.removeEventListener("mousemove",x),document.removeEventListener("mouseup",R);let n=Z0.current;if(n?.dragging&&n.path){let q1=Q1.current||D0(),u=y1.current;if(typeof u==="function")u(n.path,q1)}Z0.current={path:null,dragging:!1,startX:0,startY:0},P.current=0,O1(!1),t(null),U0(null),L0(),V2(),document.body.style.userSelect="",document.body.style.cursor="",setTimeout(()=>{j1.current=!1},0)};document.addEventListener("mousemove",x),document.addEventListener("mouseup",R)},[p0,D0,Z2,b0,V2,U0,s0,L0]),f5=i(async(Z)=>{let N=Array.from(Z?.target?.files||[]);if(N.length===0)return;let Y=P1.current||".";if(await Y2(N,Y),P1.current=".",Z?.target)Z.target.value=""},[Y2]);return K`
        <aside
            class=${`workspace-sidebar${k1?" workspace-drop-active":""}`}
            data-workspace-scale=${F0}
            ref=${A0}
            onDragEnter=${Y4}
            onDragOver=${A4}
            onDragLeave=${t2}
            onDrop=${C4}
        >
            <input type="file" multiple style="display:none" ref=${x1} onChange=${f5} />
            <div class="workspace-header">
                <div class="workspace-header-left">
                    <div class="workspace-menu-wrap">
                        <button
                            ref=${L1}
                            class=${`workspace-menu-button${B0?" active":""}`}
                            onClick=${(Z)=>{Z.stopPropagation(),x0((N)=>!N)}}
                            title="Workspace actions"
                            aria-label="Workspace actions"
                            aria-haspopup="menu"
                            aria-expanded=${B0?"true":"false"}
                        >
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <line x1="4" y1="7" x2="20" y2="7" />
                                <line x1="4" y1="12" x2="20" y2="12" />
                                <line x1="4" y1="17" x2="20" y2="17" />
                            </svg>
                        </button>
                        ${B0&&K`
                            <div class="workspace-menu-dropdown" ref=${V1} role="menu" aria-label="Workspace options">
                                <button class="workspace-menu-item" role="menuitem" onClick=${O5} disabled=${F1}>New file</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${J0} disabled=${F1}>Upload files</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${O4}>Refresh tree</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${()=>n1(()=>w2())} disabled=${l1}>
                                    ${l1?"Reindexing workspace…":"Reindex workspace"}
                                </button>
                                <button class=${`workspace-menu-item${_1?" active":""}`} role="menuitem" onClick=${T4}>
                                    ${_1?"Hide hidden files":"Show hidden files"}
                                </button>

                                ${(_||Q||U)&&K`<div class="workspace-menu-separator"></div>`}
                                ${_&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${e2}>
                                        Open terminal in tab
                                    </button>
                                `}
                                ${Q&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${S5}>
                                        Open VNC in tab
                                    </button>
                                `}
                                ${U&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${k5}>
                                        ${X?"Hide terminal dock":"Show terminal dock"}
                                    </button>
                                `}

                                ${L&&K`<div class="workspace-menu-separator"></div>`}
                                ${a0&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${M4}>Open in tab</button>
                                `}
                                ${L&&!i1&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${A2} disabled=${!c0}>Open in editor</button>
                                `}
                                ${u0&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${T5}>Rename selected</button>
                                `}
                                ${Y5&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${y4}>Download selected file</button>
                                `}
                                ${_2&&K`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${y5}>Download selected folder (zip)</button>
                                `}
                                ${H5&&K`
                                    <button class="workspace-menu-item danger" role="menuitem" onClick=${M5}>Delete selected file</button>
                                `}
                            </div>
                        `}
                    </div>
                    <span>Workspace</span>
                </div>
                <div class="workspace-header-actions">
                    <button class="workspace-create" onClick=${P2} title="New file" disabled=${F1}>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <line x1="12" y1="5" x2="12" y2="19" />
                            <line x1="5" y1="12" x2="19" y2="12" />
                        </svg>
                    </button>
                    <button class="workspace-refresh" onClick=${m0} title="Refresh tree">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <circle cx="12" cy="12" r="8.5" stroke-dasharray="42 12" stroke-dashoffset="6"
                                transform="rotate(75 12 12)" />
                            <polyline points="21 3 21 9 15 9" />
                        </svg>
                    </button>
                </div>
            </div>
            ${V4&&K`
                <div class="workspace-index-status-row">
                    <div class=${`workspace-index-status-chip state-${l0}`} title=${D5}>
                        <span class="workspace-index-status-dot" aria-hidden="true"></span>
                        <span>${A5}</span>
                    </div>
                </div>
            `}
            <div class="workspace-tree" onClick=${L4}>
                ${T1&&K`
                    <div class="workspace-upload-strip">
                        <div class="workspace-upload-strip-text">
                            ${T1.error?K`<span class="workspace-upload-strip-error">${T1.error}</span>`:T1.done?K`<span>Done</span>`:K`<span>${T1.total>1?`Uploading ${T1.current}/${T1.total}: ${T1.name}`:`Uploading ${T1.name}`}${T1.percent>0?` (${T1.percent}%)`:"…"}</span>`}
                        </div>
                        ${!T1.done&&!T1.error&&K`
                            <div class="workspace-upload-strip-bar">
                                <div class="workspace-upload-strip-fill" style=${`width:${T1.percent||0}%`}></div>
                            </div>
                        `}
                    </div>
                `}
                ${d&&K`<div class="workspace-loading">Loading…</div>`}
                ${r&&K`<div class="workspace-error">${r}</div>`}
                ${J&&K`
                    <div
                        class="workspace-tree-list"
                        ref=${d1}
                        tabIndex="0"
                        onClick=${t0}
                        onDblClick=${x2}
                        onKeyDown=${o2}
                        onTouchStart=${C5}
                        onTouchEnd=${a2}
                        onTouchMove=${N4}
                        onTouchCancel=${a2}
                    >
                        ${R2.map(({node:Z,depth:N})=>{let Y=Z.type==="dir",D=Z.path===L,x=Z.path===b,R=Y&&V.has(Z.path),n=I&&Z.path===I,q1=Array.isArray(Z.children)&&Z.children.length>0?Z.children.length:Number(Z.child_count)||0;return K`
                                <div
                                    key=${Z.path}
                                    class=${`workspace-row${D?" selected":""}${n?" drop-target":""}`}
                                    style=${{paddingLeft:`${8+N*G4.indentPx}px`}}
                                    data-path=${Z.path}
                                    data-type=${Z.type}
                                    onMouseDown=${E5}
                                >
                                    <span class="workspace-caret" aria-hidden="true">
                                        ${Y?R?K`<svg viewBox="0 0 12 12"><polygon points="1,2 11,2 6,11"/></svg>`:K`<svg viewBox="0 0 12 12"><polygon points="2,1 11,6 2,11"/></svg>`:null}
                                    </span>
                                    <svg class=${`workspace-node-icon${Y?" folder":""}`}
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                        aria-hidden="true">
                                        ${Y?K`<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>`:K`<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>`}
                                    </svg>
                                    ${x?K`
                                            <input
                                                class="workspace-rename-input"
                                                ref=${f0}
                                                value=${l}
                                                onInput=${(u)=>w(u?.target?.value||"")}
                                                onKeyDown=${(u)=>{if(u.stopPropagation(),u.key==="Enter")u.preventDefault(),L2();else if(u.key==="Escape")u.preventDefault(),P0()}}
                                                onBlur=${P0}
                                                onClick=${(u)=>u.stopPropagation()}
                                            />
                                        `:K`<span class="workspace-label"><span class="workspace-label-text">${Z.name}</span></span>`}
                                    ${Y&&!R&&q1>0&&K`
                                        <span class="workspace-count">${q1}</span>
                                    `}
                                    ${Y&&K`
                                        <button
                                            class="workspace-folder-upload"
                                            data-upload-target=${Z.path}
                                            title="Upload files to this folder"
                                            onClick=${I4}
                                            disabled=${F1}
                                        >
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                                <polyline points="7 8 12 3 17 8"/>
                                                <line x1="12" y1="3" x2="12" y2="15"/>
                                            </svg>
                                        </button>
                                    `}
                                </div>
                            `})}
                    </div>
                `}
            </div>
            ${L&&K`
                <div class="workspace-preview-splitter-h" onMouseDown=${I5} onTouchStart=${F4}></div>
                <div class="workspace-preview">
                    <div class="workspace-preview-header">
                        <span class="workspace-preview-title">${L}</span>
                        <div class="workspace-preview-actions">
                            <button class="workspace-create" onClick=${P2} title="New file" disabled=${F1}>
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                    <line x1="12" y1="5" x2="12" y2="19" />
                                    <line x1="5" y1="12" x2="19" y2="12" />
                                </svg>
                            </button>
                            ${!i1&&K`
                                <button
                                    class="workspace-download workspace-edit"
                                    onClick=${()=>c0&&w1.current?.(L,S)}
                                    title=${Z4}
                                    disabled=${!c0}
                                >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M12 20h9" />
                                        <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
                                    </svg>
                                </button>
                                <button
                                    class="workspace-download workspace-delete"
                                    onClick=${e0}
                                    title="Delete file"
                                >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <polyline points="3 6 5 6 21 6" />
                                        <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                                        <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
                                        <line x1="10" y1="11" x2="10" y2="17" />
                                        <line x1="14" y1="11" x2="14" y2="17" />
                                    </svg>
                                </button>
                            `}
                            ${i1?K`
                                    <button class="workspace-download" onClick=${v2}
                                        title="Upload files to this folder" disabled=${F1}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 8 12 3 17 8"/>
                                            <line x1="12" y1="3" x2="12" y2="15"/>
                                        </svg>
                                    </button>
                                    <a class="workspace-download" href=${q3(L,_1)} download
                                        title="Download folder as zip" onClick=${(Z)=>Z.stopPropagation()}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 10 12 15 17 10"/>
                                            <line x1="12" y1="15" x2="12" y2="3"/>
                                        </svg>
                                    </a>`:K`<a class="workspace-download" href=${B3(L)} download
                                        title="Download" onClick=${(Z)=>Z.stopPropagation()}>
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                        <polyline points="7 10 12 15 17 10"/>
                                        <line x1="12" y1="15" x2="12" y2="3"/>
                                    </svg>
                                </a>`}
                        </div>
                    </div>
                    ${U1&&K`<div class="workspace-loading">Loading preview…</div>`}
                    ${S?.error&&K`<div class="workspace-error">${S.error}</div>`}
                    ${i1&&K`
                        <div class="workspace-preview-text">Folder selected — create file, upload files, or download as zip.</div>
                        ${A1?.loading&&K`<div class="workspace-loading">Loading folder size preview…</div>`}
                        ${A1?.error&&K`<div class="workspace-error">${A1.error}</div>`}
                        ${A1?.payload&&A1.payload.segments?.length>0&&K`
                            <${Nq} payload=${A1.payload} />
                        `}
                        ${A1?.payload&&(!A1.payload.segments||A1.payload.segments.length===0)&&K`
                            <div class="workspace-preview-text">No file size data available for this folder yet.</div>
                        `}
                    `}
                    ${S&&!S.error&&!i1&&K`
                        <div class="workspace-preview-body" ref=${e}></div>
                    `}
                </div>
            `}
            ${K1&&K`
                <div class="workspace-drag-ghost" ref=${C1}>${K1.label}</div>
            `}
        </aside>
    `}var Dq=new Set(["html-viewer","kanban-editor","mindmap-editor"]);function N5($,j,q){let B=String($||"").trim();if(!B)return null;if(j)return j;if(typeof q!=="function")return null;return q({path:B,mode:"edit"})?.id||null}function D$($,j,q){let B=N5($,j,q);return B!=null&&Dq.has(B)}function C$($,j,q){return N5($,j,q)==="html-viewer"?"Edit":"Edit Source"}function I$($,j,q){return N5($,j,q)==="editor"}var Cq=/\.(docx?|xlsx?|pptx?|odt|ods|odp|rtf)$/i,Iq=/\.(csv|tsv)$/i,Oq=/\.pdf$/i,Tq=/\.(png|jpe?g|gif|webp|bmp|ico|svg)$/i,O$=/\.drawio(\.xml|\.svg|\.png)?$/i;function Mq($,{hasPopOutTab:j=!1}={}){let q=typeof $==="string"?$.trim():"";if(!q)return null;if(Cq.test(q)){let B="/workspace/raw?path="+encodeURIComponent(q),_=q.split("/").pop()||"document";return"/office-viewer/?url="+encodeURIComponent(B)+"&name="+encodeURIComponent(_)}if(Iq.test(q))return"/csv-viewer/?path="+encodeURIComponent(q);if(Oq.test(q))return"/workspace/raw?path="+encodeURIComponent(q);if(Tq.test(q)&&!O$.test(q))return"/image-viewer/?path="+encodeURIComponent(q);if(O$.test(q)&&!j)return"/drawio/edit?path="+encodeURIComponent(q);return null}function T$({tabs:$,activeId:j,onActivate:q,onClose:B,onCloseOthers:_,onCloseAll:Q,onTogglePin:U,onTogglePreview:X,onToggleDiff:J,onEditSource:W,previewTabs:V,diffTabs:G,paneOverrides:L,detachedTabs:C,onReattachTab:b,onToggleDock:c,dockVisible:l,onToggleZen:w,zenMode:S,onPopOutTab:m}){let[O,d]=y(null),g=M(null);E(()=>{if(!O)return;let I=(a)=>{if(a.type==="keydown"&&a.key!=="Escape")return;d(null)};return document.addEventListener("click",I),document.addEventListener("keydown",I),()=>{document.removeEventListener("click",I),document.removeEventListener("keydown",I)}},[O]),E(()=>{let I=(a)=>{if(a.ctrlKey&&a.key==="Tab"){if(a.preventDefault(),!$.length)return;let F1=$.findIndex((X1)=>X1.id===j);if(a.shiftKey){let X1=$[(F1-1+$.length)%$.length];q?.(X1.id)}else{let X1=$[(F1+1)%$.length];q?.(X1.id)}return}if((a.ctrlKey||a.metaKey)&&a.key==="w"){let F1=document.querySelector(".editor-pane");if(F1&&F1.contains(document.activeElement)){if(a.preventDefault(),j)B?.(j)}}};return document.addEventListener("keydown",I),()=>document.removeEventListener("keydown",I)},[$,j,q,B]);let U1=i((I,a)=>{if(I.button===1)I.preventDefault(),B?.(a)},[B]),k=i((I,a)=>{if(I.defaultPrevented)return;if(I.button===0)q?.(a)},[q]),r=i((I,a)=>{I.preventDefault(),d({id:a,x:I.clientX,y:I.clientY})},[]),f=i((I)=>{I.preventDefault(),I.stopPropagation()},[]),_1=i((I,a)=>{I.preventDefault(),I.stopPropagation(),B?.(a)},[B]);E(()=>{if(!j||!g.current)return;let I=g.current.querySelector(".tab-item.active");if(I)I.scrollIntoView({block:"nearest",inline:"nearest",behavior:"smooth"})},[j]);let Y1=i((I)=>{if(!(L instanceof Map))return null;return L.get(I)||null},[L]),k1=B1(()=>$.find((I)=>I.id===O?.id)||null,[O?.id,$]),O1=B1(()=>{let I=O?.id;if(!I)return!1;return D$(I,Y1(I),(a)=>n0.resolve(a))},[O?.id,Y1]),s=B1(()=>{let I=O?.id;if(!I)return"Edit Source";return C$(I,Y1(I),(a)=>n0.resolve(a))},[O?.id,Y1]),t=B1(()=>{let I=O?.id;if(!I||!(C instanceof Map))return!1;return C.has(I)},[O?.id,C]),K1=B1(()=>{let I=O?.id;if(!I||!(G instanceof Set))return!1;return G.has(I)},[O?.id,G]),G1=B1(()=>{let I=O?.id;if(!I)return!1;let a=$.find((X1)=>X1.id===I)||null;if(!a)return!1;return I$(I,Y1(I),(X1)=>n0.resolve(X1))&&Boolean(a.dirty||K1)},[O?.id,K1,Y1,$]);if(!$.length)return null;return K`
        <div class="tab-strip" ref=${g} role="tablist">
            ${$.map((I)=>K`
                <div
                    key=${I.id}
                    class=${`tab-item${I.id===j?" active":""}${I.dirty?" dirty":""}${I.pinned?" pinned":""}`}
                    role="tab"
                    aria-selected=${I.id===j}
                    title=${I.path}
                    onMouseDown=${(a)=>U1(a,I.id)}
                    onClick=${(a)=>k(a,I.id)}
                    onContextMenu=${(a)=>r(a,I.id)}
                >
                    ${I.pinned&&K`
                        <span class="tab-pin-icon" aria-label="Pinned">
                            <svg viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
                                <path d="M4.456.734a1.75 1.75 0 0 1 2.826.504l.613 1.327a3.1 3.1 0 0 0 2.084 1.707l2.454.584c1.332.317 1.8 1.972.832 2.94L11.06 10l3.72 3.72a.75.75 0 1 1-1.06 1.06L10 11.06l-2.204 2.205c-.968.968-2.623.5-2.94-.832l-.584-2.454a3.1 3.1 0 0 0-1.707-2.084l-1.327-.613a1.75 1.75 0 0 1-.504-2.826z"/>
                            </svg>
                        </span>
                    `}
                    <span class="tab-label">${I.label}</span>
                    ${C instanceof Map&&C.has(I.id)&&K`
                        <span class="tab-detached-badge" aria-label="Detached" title="Open in separate window">↗</span>
                    `}
                    <button
                        type="button"
                        class="tab-close"
                        onPointerDown=${f}
                        onMouseDown=${f}
                        onClick=${(a)=>_1(a,I.id)}
                        title=${I.dirty?"Unsaved changes":"Close"}
                        aria-label=${I.dirty?"Unsaved changes":`Close ${I.label}`}
                    >
                        ${I.dirty?K`<span class="tab-dirty-dot" aria-hidden="true"></span>`:K`<svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true" focusable="false" style=${{pointerEvents:"none"}}>
                                <line x1="4" y1="4" x2="12" y2="12" style=${{pointerEvents:"none"}}/>
                                <line x1="12" y1="4" x2="4" y2="12" style=${{pointerEvents:"none"}}/>
                            </svg>`}
                    </button>
                </div>
            `)}
            ${c&&K`
                <div class="tab-strip-spacer"></div>
                <button
                    class=${`tab-strip-dock-toggle${l?" active":""}`}
                    onClick=${c}
                    title=${`${l?"Hide":"Show"} terminal (Ctrl+\`)`}
                    aria-label=${`${l?"Hide":"Show"} terminal`}
                    aria-pressed=${l?"true":"false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="1.75" y="2.25" width="12.5" height="11.5" rx="2"/>
                        <polyline points="4.5 5.25 7 7.75 4.5 10.25"/>
                        <line x1="8.5" y1="10.25" x2="11.5" y2="10.25"/>
                    </svg>
                </button>
            `}
            ${w&&K`
                <button
                    class=${`tab-strip-zen-toggle${S?" active":""}`}
                    onClick=${w}
                    title=${`${S?"Exit":"Enter"} zen mode (Ctrl+Shift+Z)`}
                    aria-label=${`${S?"Exit":"Enter"} zen mode`}
                    aria-pressed=${S?"true":"false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        ${S?K`<polyline points="4 8 1.5 8 1.5 1.5 14.5 1.5 14.5 8 12 8"/><polyline points="4 8 1.5 8 1.5 14.5 14.5 14.5 14.5 8 12 8"/>`:K`<polyline points="5.5 1.5 1.5 1.5 1.5 5.5"/><polyline points="10.5 1.5 14.5 1.5 14.5 5.5"/><polyline points="5.5 14.5 1.5 14.5 1.5 10.5"/><polyline points="10.5 14.5 14.5 14.5 14.5 10.5"/>`}
                    </svg>
                </button>
            `}
        </div>
        ${O&&K`
            <div class="tab-context-menu" style=${{left:O.x+"px",top:O.y+"px"}}>
                <button onClick=${()=>{B?.(O.id),d(null)}}>Close</button>
                <button onClick=${()=>{_?.(O.id),d(null)}}>Close Others</button>
                <button onClick=${()=>{Q?.(),d(null)}}>Close All</button>
                <hr />
                <button onClick=${()=>{U?.(O.id),d(null)}}>
                    ${k1?.pinned?"Unpin":"Pin"}
                </button>
                ${O1&&W&&K`
                    <button onClick=${()=>{W(O.id),d(null)}}>${s}</button>
                `}
                ${t&&b&&K`
                    <button onClick=${()=>{b(O.id),d(null)}}>Reattach</button>
                `}
                ${m&&!t&&K`
                    <button onClick=${()=>{let I=$.find((a)=>a.id===O.id);m(O.id,I?.label),d(null)}}>Open in Window</button>
                `}
                ${G1&&J&&K`
                    <hr />
                    <button onClick=${()=>{q?.(O.id),J(O.id),d(null)}}>${K1?"Hide Diff":"Compare to Saved"}</button>
                `}
                ${X&&/\.(md|mdx|markdown)$/i.test(O.id)&&K`
                    <hr />
                    <button onClick=${()=>{X(O.id),d(null)}}>
                        ${V?.has(O.id)?"Hide Preview":"Preview"}
                    </button>
                `}
                ${(()=>{let I=Mq(O.id,{hasPopOutTab:typeof m==="function"});if(!I)return null;return K`
                        <hr />
                        <button onClick=${()=>{window.open(I,"_blank","noopener"),d(null)}}>Open in New Tab</button>
                    `})()}
            </div>
        `}
    `}function v3($,j){try{if($)$.name=j;return!0}catch(q){return!1}}function b3($,j){try{return $?.contentWindow?.postMessage?.(j,"*"),!0}catch(q){return!1}}function yq($){let j=new Map,q=[];for(let U of $||[])j.set(U.id,{...U,children:[],depth:0});for(let U of $||[]){let X=j.get(U.id);if(!X)continue;let J=U.parentId?j.get(U.parentId):null;if(J)J.children.push(X);else q.push(X)}let B=new Set;for(let[,U]of j){if(U.role!=="assistant"||!U.toolName)continue;if(U.children.length!==1)continue;let X=U.children[0];if(X.role!=="toolResult")continue;U.resultDetail=X.detail||null,U.resultLength=X.contentLength||0,U.resultId=X.id,U.merged=!0,U.children=X.children;for(let J of U.children)J.parentId=U.id;B.add(X.id)}let _=q.filter((U)=>!B.has(U.id)),Q=[];for(let U=_.length-1;U>=0;U--)_[U].depth=0,Q.push(_[U]);while(Q.length>0){let U=Q.pop(),X=U.children.length>1;for(let J=U.children.length-1;J>=0;J--)U.children[J].depth=X?U.depth+1:U.depth,Q.push(U.children[J])}return _}function Sq($){let j=[],q=[];for(let B=$.length-1;B>=0;B--)q.push($[B]);while(q.length>0){let B=q.pop();j.push(B);for(let _=B.children.length-1;_>=0;_--)q.push(B.children[_])}return j}function M$($){if(!$||$<=0)return"";if($<1000)return`${$}`;if($<1e6)return`${($/1000).toFixed(1)}k`;return`${($/1e6).toFixed(1)}M`}function c2($){if(!$||$<=0)return"";if($<1000)return`${$} chars`;if($<1e6)return`${($/1000).toFixed(1)}k chars`;return`${($/1e6).toFixed(1)}M chars`}function kq($){let j=$.type;if(j==="model_change")return{tag:"model",tagClass:"system",text:`${$.preview?.replace("[model ","").replace("]","")||""}`};if(j==="thinking_level_change")return{tag:"thinking",tagClass:"system",text:$.preview?.replace("[thinking ","").replace("]","")||""};if(j==="compaction")return{tag:"compaction",tagClass:"system",text:$.preview?.replace("[compaction: ","").replace("]","")||""};if(j==="label")return{tag:"label",tagClass:"system",text:$.preview?.replace("[label ","").replace("]","")||""};if(j==="session_info")return{tag:"session",tagClass:"system",text:$.preview?.replace("[session name ","").replace("]","")||""};if(j==="branch_summary")return{tag:"summary",tagClass:"system",text:$.preview||""};if(j!=="message")return{tag:j||"?",tagClass:"system",text:$.preview||""};let q=$.role||"message";if($.merged&&$.toolName){let _=($.toolInput||"").split(`
`)[0],Q=_.length>120?_.slice(0,119)+"…":_;return{tag:$.toolName,tagClass:"tool",text:Q||""}}if(q==="assistant"&&$.toolName){let _=($.toolInput||"").split(`
`)[0],Q=_.length>120?_.slice(0,119)+"…":_;return{tag:$.toolName,tagClass:"tool",text:Q||"…"}}if(q==="toolResult"){let _=($.detail||"").split(`
`)[0],Q=_.length>120?_.slice(0,119)+"…":_;return{tag:`→ ${$.toolName||"result"}`,tagClass:"result",text:Q}}if(q==="user"){let Q=($.previewText||$.detail||$.preview||"").replace(/^user:\s*"?/,"").replace(/"?\s*$/,"").split(`
`)[0];return{tag:"user",tagClass:"user",text:Q.length>120?Q.slice(0,119)+"…":Q}}if(q==="assistant"){let Q=($.detail||$.preview||"").replace(/^assistant:\s*"?/,"").replace(/"?\s*$/,"").split(`
`)[0];return{tag:"assistant",tagClass:"assistant",text:Q.length>120?Q.slice(0,119)+"…":Q}}return{tag:q,tagClass:"other",text:$.preview||""}}function Eq($,j=!1){let q=typeof $==="string"?$.trim():"";if(!q)return null;return{text:j?`/tree ${q} --summarize`:`/tree ${q}`,navigateTargetId:q,summarize:Boolean(j)}}function y$($){let j=typeof $==="string"?$.trim():"";if(!j.startsWith("/tree"))return null;let q=j.split(/\s+/).filter(Boolean);if(q[0]!=="/tree")return null;let B=null,_=!1;for(let Q=1;Q<q.length;Q++){let U=q[Q];if(U==="--summarize"){_=!0;continue}if(!B&&!U.startsWith("--"))B=U}return B?{targetId:B,summarize:_}:null}function fq($,j,q,B){let _=Array.isArray($)?$:[];if(_.length===0)return null;let Q=(X)=>typeof X==="string"&&_.some((J)=>J?.id===X);if(Q(j))return j;if(Q(q))return q;if(Q(B))return B;let U=_.find((X)=>X?.active);if(U?.id)return U.id;return _[0]?.id||null}function Pq($){if(!$||typeof $!=="object")return null;let j=typeof $.type==="string"?$.type:"",q=typeof $.preview==="string"?$.preview.trim():"",B=typeof $.error==="string"?$.error.trim():"",_=y$(q),Q=q||"tree command";if(j==="submit_pending")return{tone:"pending",text:_?`Sending ${Q}`:"Sending tree command…",refreshDelays:[]};if(j==="submit_sent")return{tone:"info",text:_?`Running ${Q}`:"Tree command sent.",refreshDelays:_?[500,1500,3500,8000]:[]};if(j==="submit_queued")return{tone:"info",text:_?`Queued ${Q}`:"Tree command queued.",refreshDelays:_?[1200,3200,7000,12000]:[]};if(j==="submit_failed")return{tone:"error",text:B?`Tree command failed: ${B}`:"Tree command failed.",refreshDelays:[]};if(j==="refresh_building")return{tone:"pending",text:"Refreshing widget…",refreshDelays:[]};if(j==="refresh_failed")return{tone:"error",text:B?`Refresh failed: ${B}`:"Refresh failed.",refreshDelays:[]};if(j==="refresh_dashboard"||j==="refresh_ack")return{tone:"success",text:"Widget refreshed.",refreshDelays:[]};return null}function S$({widget:$,onWidgetEvent:j}){let q=$?.artifact?.tree&&typeof $.artifact.tree==="object"?$.artifact.tree:null,B=typeof $?.originChatJid==="string"&&$.originChatJid.trim()?$.originChatJid.trim():null,_=$?.runtimeState&&typeof $.runtimeState==="object"?$.runtimeState:null,Q=_?.lastHostUpdate&&typeof _.lastHostUpdate==="object"?_.lastHostUpdate:null,[U,X]=y(()=>({loading:!q,error:null,data:q})),[J,W]=y(()=>q?.leafId||null),[V,G]=y(""),L=M(null),C=M(null),b=M(q?.leafId||null),c=M(null),l=M(""),w=async()=>{X((k)=>({...k,loading:!0,error:null}));try{let k=B?`?chat_jid=${encodeURIComponent(B)}`:"",r=await fetch(`/agent/session-tree${k}`,{method:"GET",credentials:"same-origin",headers:{Accept:"application/json"}}),f=await r.json().catch(()=>({}));if(!r.ok)throw Error(f?.error||`HTTP ${r.status}`);X({loading:!1,error:null,data:f})}catch(k){X((r)=>({loading:!1,error:k?.message||"Failed to load tree.",data:r?.data||q||null}))}};c.current=w,E(()=>{w()},[B]);let S=B1(()=>{let k=U.data;if(!k||!Array.isArray(k.nodes)||k.nodes.length===0)return[];return Sq(k.flat?yq(k.nodes):k.nodes)},[U.data]);E(()=>{let k=fq(S,J,b.current,U.data?.leafId||null);if(k!==J)W(k);if(b.current&&U.data?.leafId===b.current)b.current=null},[S,J,U.data?.leafId]);let m=B1(()=>{let k=(V||"").trim().toLowerCase();if(!k)return S;return S.filter((r)=>{return[r.preview,r.toolInput,r.toolInputFull,r.detail,r.toolName,r.role,r.id,r.resultDetail,r.type,r.label].some((_1)=>typeof _1==="string"&&_1.toLowerCase().includes(k))})},[S,V]),O=B1(()=>m.find((k)=>k.id===J)||null,[m,J]),d=B1(()=>Pq(Q),[Q?.type,Q?.preview,Q?.error,Q?.submittedAt]);E(()=>{if(C.current)C.current.scrollIntoView({block:"center",behavior:"auto"})},[J,U.data?.leafId,m.length]),E(()=>{let k=y$(Q?.preview);if(k?.targetId)b.current=k.targetId;let r=d?.refreshDelays||[];if(!r.length)return;let f=[B||"",Q?.type||"",Q?.submittedAt||"",Q?.preview||""].join("|");if(l.current===f)return;l.current=f;let _1=r.map((Y1)=>setTimeout(()=>c.current?.(),Y1));return()=>_1.forEach((Y1)=>clearTimeout(Y1))},[B,Q?.type,Q?.submittedAt,Q?.preview,d?.refreshDelays]);let g=(k=!1)=>{let r=O?.id,f=Eq(r,k);if(!f)return;b.current=f.navigateTargetId,j?.({kind:"widget.submit",payload:f},$)},U1=d?.tone||"info";return K`
        <div class="session-tree-widget">
            <div class="session-tree-toolbar">
                <div class="session-tree-toolbar-left">
                    <button class="session-tree-btn" type="button" onClick=${()=>w()} disabled=${U.loading}>${U.loading?"Loading…":"Refresh"}</button>
                    <input ref=${L}
                        class="st-search-input" type="text" placeholder="Filter\u2026"
                        value=${V}
                        onInput=${(k)=>G(k.currentTarget.value)}
                        onKeyDown=${(k)=>{if(k.key==="Escape")G(""),k.currentTarget.blur()}}
                    />
                    ${V&&K`<span class="session-tree-meta">${m.length} match${m.length!==1?"es":""}</span>`}
                    ${U.error&&K`<span class="session-tree-error-inline">${U.error}</span>`}
                </div>
                <div class="session-tree-toolbar-right">
                    ${d?.text&&K`<span class=${`session-tree-host-update ${U1}`}>${d.text}</span>`}
                    ${U.data?.capped&&K`<span class="session-tree-meta">Showing ${U.data?.nodes?.length||0} of ${U.data?.total||0}</span>`}
                    ${B&&K`<span class="session-tree-meta">${B}</span>`}
                </div>
            </div>

            <div class="session-tree-content">
                <div class="session-tree-list" role="tree" aria-label="Session tree">
                    ${U.loading&&m.length===0&&!V&&K`<div class="session-tree-empty">Loading session tree\u2026</div>`}
                    ${!U.loading&&m.length===0&&!V&&K`<div class="session-tree-empty">Session tree is empty.</div>`}
                    ${!U.loading&&m.length===0&&V&&K`<div class="session-tree-empty">No entries match \u201c${V}\u201d</div>`}
                    ${m.map((k)=>{let r=J===k.id,f=`st-row${k.active?" active":""}${r?" selected":""}`,_1=(k.children||[]).length>1,Y1=kq(k);return K`
                            <button key=${k.id} ref=${k.active||r?C:null}
                                class=${f} type="button" role="treeitem" aria-selected=${r}
                                onClick=${()=>W(k.id)}>
                                <span class="st-indent" style=${`width:${(k.depth||0)*16+6}px`}></span>
                                <span class=${`st-dot${k.active?" active":_1?" branch":""}`}></span>
                                <span class=${`st-tag ${Y1.tagClass}`}>${Y1.tag}</span>
                                <span class="st-text">${Y1.text}</span>
                                ${k.merged&&k.resultLength>0&&K`<span class="st-size">${M$(k.resultLength)}</span>`}
                                ${!k.merged&&k.contentLength>3000&&K`<span class="st-size">${M$(k.contentLength)}</span>`}
                                ${k.hasThinking&&K`<span class="st-badge thinking">\u{1F4AD}</span>`}
                                ${k.label&&K`<span class="st-label">${k.label}</span>`}
                                ${k.active&&K`<span class="st-active">\u25C0</span>`}
                            </button>
                        `})}
                </div>

                <aside class="session-tree-sidebar">
                    ${O?K`
                        <div class="st-side-section">
                            <div class="st-side-label">Entry</div>
                            <div class="st-side-mono">${O.id}${O.resultId?` → ${O.resultId}`:""}</div>
                        </div>
                        <div class="st-side-section">
                            <div class="st-side-label">Type</div>
                            <div class="st-side-value">${O.role||O.type||"entry"}${O.toolName?` → ${O.toolName}`:""}${O.merged?" (merged)":""}</div>
                        </div>
                        ${O.toolInputFull&&K`
                            <div class="st-side-section">
                                <div class="st-side-label">${O.toolName==="bash"?"Command":"Input"}</div>
                                <pre class="st-side-code">${O.toolInputFull}</pre>
                            </div>
                        `}
                        ${O.resultDetail&&K`
                            <div class="st-side-section">
                                <div class="st-side-label">Result${O.resultLength?` (${c2(O.resultLength)})`:""}</div>
                                <pre class="st-side-code">${O.resultDetail}</pre>
                            </div>
                        `}
                        ${O.detail&&!O.toolInput&&K`
                            <div class="st-side-section">
                                <div class="st-side-label">${O.role==="toolResult"?"Output":"Content"}${O.contentLength?` (${c2(O.contentLength)})`:""}</div>
                                <pre class="st-side-code">${O.detail}</pre>
                            </div>
                        `}
                        ${O.rawDetail&&K`
                            <div class="st-side-section">
                                <div class="st-side-label">Raw prompt${O.rawContentLength?` (${c2(O.rawContentLength)})`:""}</div>
                                <pre class="st-side-code">${O.rawDetail}</pre>
                            </div>
                        `}
                        ${O.timestamp&&K`
                            <div class="st-side-section">
                                <div class="st-side-label">Time</div>
                                <div class="st-side-value">${new Date(O.timestamp).toLocaleString()}</div>
                            </div>
                        `}
                        ${(O.contentLength>0||O.hasThinking)&&K`
                            <div class="st-side-section">
                                <div class="st-side-label">Size</div>
                                <div class="st-side-badges">
                                    ${O.contentLength>0&&K`<span class="st-pill">${c2(O.contentLength)} content</span>`}
                                    ${O.hasThinking&&K`<span class="st-pill thinking">${c2(O.thinkingLength)} thinking</span>`}
                                    ${O.merged&&O.resultLength>0&&K`<span class="st-pill">${c2(O.resultLength)} result</span>`}
                                </div>
                            </div>
                        `}
                        <div class="st-side-actions">
                            <button class="session-tree-btn primary" type="button" onClick=${()=>g(!1)}>Navigate here</button>
                            <button class="session-tree-btn" type="button" onClick=${()=>g(!0)}>Navigate + summarize</button>
                        </div>
                    `:K`<div class="session-tree-empty side">Select an entry to inspect.</div>`}
                </aside>
            </div>
        </div>
    `}function k$({widget:$,onClose:j,onWidgetEvent:q}){let B=M(null),_=M(!1),Q=B1(()=>T6($),[$?.artifact?.kind,$?.artifact?.html,$?.artifact?.svg,$?.widgetId,$?.toolCallId,$?.turnId,$?.title]);if(E(()=>{if(!$)return;let w=(S)=>{if(S.key==="Escape")j?.()};return document.addEventListener("keydown",w),()=>document.removeEventListener("keydown",w)},[$,j]),E(()=>{_.current=!1},[Q]),E(()=>{if(!$)return;let w=B.current;if(!w)return;let S=(U1)=>{let k=W5($),r=U1==="widget.init"?U5($):X5($);v3(w,k),b3(w,{__piclawGeneratedWidgetHost:!0,type:U1,widgetId:$?.widgetId||null,toolCallId:$?.toolCallId||null,turnId:$?.turnId||null,payload:r})},m=()=>{S("widget.init"),S("widget.update")},O=()=>{_.current=!0,m()};w.addEventListener("load",O);let g=[0,40,120,300,800].map((U1)=>setTimeout(m,U1));return()=>{w.removeEventListener("load",O),g.forEach((U1)=>clearTimeout(U1))}},[Q,$?.widgetId,$?.toolCallId,$?.turnId]),E(()=>{if(!$)return;let w=B.current;if(!w?.contentWindow)return;let S=W5($),m=X5($);v3(w,S),b3(w,{__piclawGeneratedWidgetHost:!0,type:"widget.update",widgetId:$?.widgetId||null,toolCallId:$?.toolCallId||null,turnId:$?.turnId||null,payload:m});return},[$?.widgetId,$?.toolCallId,$?.turnId,$?.status,$?.subtitle,$?.description,$?.error,$?.width,$?.height,$?.runtimeState]),E(()=>{if(!$)return;let w=(S)=>{let m=S?.data;if(!m||m.__piclawGeneratedWidget!==!0)return;let O=B.current,d=Z3($),g=Z3({widgetId:m.widgetId,toolCallId:m.toolCallId});if(g&&d&&g!==d)return;if(!g&&O?.contentWindow&&S.source!==O.contentWindow)return;q?.(m,$)};return window.addEventListener("message",w),()=>window.removeEventListener("message",w)},[$,q]),!$)return null;let X=($?.artifact||{}).kind||$?.kind||"html",J=typeof $?.title==="string"&&$.title.trim()?$.title.trim():"Generated widget",W=typeof $?.subtitle==="string"&&$.subtitle.trim()?$.subtitle.trim():"",V=$?.source==="live"?"live":"timeline",G=typeof $?.status==="string"&&$.status.trim()?$.status.trim():"final",L=V==="live"?`Live widget • ${G.toUpperCase()}`:$?.originPostId?`Message #${$.originPostId}`:"Timeline launch",C=typeof $?.description==="string"&&$.description.trim()?$.description.trim():"",b=!Q,c=O6($),l=I6($);return K`
        <div class="floating-widget-backdrop" onClick=${()=>j?.()}>
            <section
                class="floating-widget-pane"
                aria-label=${J}
                onClick=${(w)=>w.stopPropagation()}
            >
                <div class="floating-widget-header">
                    <div class="floating-widget-heading">
                        <div class="floating-widget-eyebrow">${L} • ${X.toUpperCase()}</div>
                        <div class="floating-widget-title">${J}</div>
                        ${(W||C)&&K`
                            <div class="floating-widget-subtitle">${W||C}</div>
                        `}
                    </div>
                    <button
                        class="floating-widget-close"
                        type="button"
                        onClick=${()=>j?.()}
                        title="Close widget"
                        aria-label="Close widget"
                    >
                        Close
                    </button>
                </div>
                <div class="floating-widget-body">
                    ${X==="session_tree"?K`<${S$} widget=${$} onWidgetEvent=${q} />`:b?K`<div class="floating-widget-empty">${c}</div>`:K`
                                <iframe
                                    ref=${B}
                                    class="floating-widget-frame"
                                    title=${J}
                                    name=${W5($)}
                                    sandbox=${l}
                                    referrerpolicy="no-referrer"
                                    srcdoc=${Q}
                                ></iframe>
                            `}
                </div>
            </section>
        </div>
    `}var Rq=new TextDecoder("utf-8",{fatal:!1});function l2($,j){return $[j]|$[j+1]<<8}function f2($,j){return($[j]|$[j+1]<<8|$[j+2]<<16|$[j+3]<<24)>>>0}function E$($,j,q){return Rq.decode($.subarray(j,j+q))}function wq($){let j=Math.max(0,$.length-65557);for(let q=$.length-22;q>=j;q-=1)if(f2($,q)===101010256)return q;return-1}function xq($,j){let q=Math.max(0,j-20);for(let B=q;B<=j-4;B+=1)if(f2($,B)===117853008)return!0;return!1}function vq($){let j=new Set;for(let q of $){let B=q.path.replace(/\/+/g,"/");if(!B)continue;if(q.isDirectory){j.add(B.endsWith("/")?B.slice(0,-1):B);continue}let _=B.split("/").filter(Boolean);if(_.length<=1)continue;let Q="";for(let U=0;U<_.length-1;U+=1)Q=Q?`${Q}/${_[U]}`:_[U],j.add(Q)}return j}function bq($){return String($||"").replace(/\\/g,"/").trim()}function f$($){switch(Number($)){case 0:return"Stored";case 8:return"Deflate";case 9:return"Deflate64";case 12:return"BZIP2";case 14:return"LZMA";case 93:return"Zstandard";default:return`Method ${$}`}}function P$($){let j=$ instanceof Uint8Array?$:$ instanceof ArrayBuffer?new Uint8Array($):new Uint8Array($.buffer,$.byteOffset,$.byteLength);if(j.length<22)throw Error("ZIP archive is too small to contain a valid directory.");let q=wq(j);if(q<0)throw Error("ZIP archive directory could not be found.");if(xq(j,q))throw Error("ZIP64 archives are not previewable yet.");let B=l2(j,q+10),_=f2(j,q+12),Q=f2(j,q+16);if(Q+_>j.length)throw Error("ZIP archive directory is truncated.");let U=[],X=Q,J=Q+_;while(X<J){if(X+46>j.length)throw Error("ZIP archive directory entry is truncated.");if(f2(j,X)!==33639248)throw Error("ZIP archive directory contains an unexpected record.");let G=l2(j,X+8),L=l2(j,X+10),C=f2(j,X+20),b=f2(j,X+24),c=l2(j,X+28),l=l2(j,X+30),w=l2(j,X+32),S=X+46,m=S+c+l,O=m+w;if(O>j.length)throw Error("ZIP archive directory entry metadata is truncated.");let d=(G&2048)===2048,g=bq(E$(j,S,c)),U1=E$(j,m,w),k=g.endsWith("/");if(g)U.push({path:g,isDirectory:k,compressedSize:C,uncompressedSize:b,compressionMethod:L,comment:d?U1:U1});X=O}U.sort((G,L)=>{if(G.isDirectory!==L.isDirectory)return G.isDirectory?-1:1;return G.path.localeCompare(L.path)});let W=U.filter((G)=>!G.isDirectory),V=vq(U);return{entries:U,summary:{fileCount:W.length,directoryCount:V.size,totalEntries:U.length,compressedBytes:W.reduce((G,L)=>G+L.compressedSize,0),uncompressedBytes:W.reduce((G,L)=>G+L.uncompressedSize,0)}}}function R$($){if(!$)return null;if($.uncompressedBytes<=0)return null;let j=1-$.compressedBytes/$.uncompressedBytes;if(!Number.isFinite(j))return null;return`${Math.round(j*100)}% smaller`}var uq="allow-scripts";function mq($){if(!($ instanceof Uint8Array)||$.length===0)return!0;let j=0,q=$.subarray(0,Math.min($.length,4096));for(let B of q){if(B===0)return!1;if(B<32&&B!==9&&B!==10&&B!==13&&B!==12)j+=1}return j/q.length<0.02}function gq($,j){let q=String($?.content_type||"").trim().toLowerCase(),B=String(j||"").trim().toLowerCase();if(q.startsWith("text/")||q==="application/json"||q==="application/xml")return!1;return q==="application/octet-stream"||B.endsWith(".sb")||B.endsWith(".sh")}function pq($){try{return new TextDecoder("utf-8",{fatal:!1}).decode($)}catch{return new TextDecoder().decode($)}}function hq($,j=null,q=null){let B=$?.metadata?.size,_=$?.content_type||"application/octet-stream",Q=q?.summary||null;return[{label:"Type",value:_},{label:"Syntax",value:j},{label:"Entries",value:Q?String(Q.totalEntries):null},{label:"Files",value:Q?String(Q.fileCount):null},{label:"Folders",value:Q?String(Q.directoryCount):null},{label:"Compressed",value:Q?C0(Q.compressedBytes):null},{label:"Uncompressed",value:Q?C0(Q.uncompressedBytes):null},{label:"Savings",value:R$(Q)},{label:"Size",value:typeof B==="number"?C0(B):null},{label:"Added",value:$?.created_at?p2($.created_at):null}].filter((U)=>U.value)}function cq($,j){let q=String($?.content_type||"").trim().toLowerCase(),B=String(j||"").trim().toLowerCase(),_=B.split("/").pop()||B;if(B.endsWith(".yaml")||B.endsWith(".yml")||q==="text/yaml"||q==="application/yaml")return"yaml";if(B.endsWith(".json")||B.endsWith(".jsonl")||q==="application/json")return"json";if(B.endsWith(".xml")||B.endsWith(".svg")||q==="application/xml"||q==="text/xml"||q==="image/svg+xml")return"xml";if(B.endsWith(".html")||B.endsWith(".htm")||q==="text/html")return"html";if(B.endsWith(".css")||q==="text/css")return"css";if(B.endsWith(".ts")||B.endsWith(".tsx")||q==="text/typescript")return B.endsWith(".tsx")?"tsx":"ts";if(B.endsWith(".js")||B.endsWith(".jsx")||q==="text/javascript"||q==="application/javascript")return B.endsWith(".jsx")?"jsx":"js";if(B.endsWith(".py")||q==="text/x-python"||q==="application/x-python-code")return"python";if(B.endsWith(".go")||q==="text/x-go")return"go";if(B.endsWith(".rb")||q==="text/x-ruby")return"ruby";if(B.endsWith(".rs")||q==="text/x-rustsrc")return"rust";if(B.endsWith(".ps1")||B.endsWith(".psm1")||B.endsWith(".psd1")||q==="text/x-powershell")return"powershell";if(_==="dockerfile"||B.endsWith(".dockerfile"))return"dockerfile";if(B.endsWith(".md")||B.endsWith(".markdown")||q==="text/markdown")return"markdown";if(B.endsWith(".sh")||B.endsWith(".bash")||B.endsWith(".zsh")||_===".bashrc"||_===".bash_profile"||_===".profile"||_===".zshrc"||_===".zprofile"||_===".zshenv"||_===".env"||_.startsWith(".env.")||q==="text/x-shellscript")return"bash";if(B.endsWith(".sql"))return"sql";if(B.endsWith(".toml")||B.endsWith(".ini")||B.endsWith(".cfg")||B.endsWith(".conf")||B.endsWith(".properties")||q==="text/toml")return"toml";return null}function lq($,j,q){let B=encodeURIComponent(j||`attachment-${$}`),_=encodeURIComponent(String($));if(q==="pdf")return`/pdf-viewer/?media=${_}&name=${B}#media=${_}&name=${B}`;if(q==="office"){let Q=i0($);return`/office-viewer/?url=${encodeURIComponent(Q)}&name=${B}`}if(q==="drawio")return`/drawio/edit.html?media=${_}&name=${B}&readonly=1#media=${_}&name=${B}&readonly=1`;return null}function w$({mediaId:$,info:j,onClose:q}){let B=j?.filename||`attachment-${$}`,_=B1(()=>X4(j?.content_type,B),[j?.content_type,B]),Q=K6(_),U=B1(()=>J6(j?.content_type),[j?.content_type]),[X,J]=y(_==="text"||_==="html"||_==="archive"),[W,V]=y(""),[G,L]=y(null),[C,b]=y(null),c=M(null),l=B1(()=>cq(j,B),[j,B]),w=B1(()=>l?d8(l):null,[l]),S=B1(()=>hq(j,!U?w:null,G),[j,U,w,G]),m=B1(()=>lq($,B,_),[$,B,_]),O=B1(()=>{if(!U||!W)return"";return B2(W)},[U,W]),d=B1(()=>{if(U||!W||!l)return"";return e4(W,l)},[U,W,l]);return E(()=>{let g=(U1)=>{if(U1.key==="Escape")q()};return document.addEventListener("keydown",g),()=>document.removeEventListener("keydown",g)},[q]),E(()=>{if(!c.current||!O)return;j5(c.current);return},[O]),E(()=>{let g=!1;async function U1(){if(_!=="text"&&_!=="html"&&_!=="archive"){J(!1),b(null),V(""),L(null);return}J(!0),b(null),V(""),L(null);try{let k=await r8($),r=new Uint8Array(await k.arrayBuffer());if(_==="text"||_==="html"){if(_==="text"&&gq(j,B)&&!mq(r))throw Error("Attachment does not appear to contain text content.");let _1=pq(r);if(!g)V(_1);return}let f=P$(r);if(!g)L(f)}catch(k){if(!g){let r=k instanceof Error?k.message:String(k||"Unknown error");b(_==="archive"?`Failed to load ZIP preview. ${r}`:`Failed to load text preview. ${r}`)}}finally{if(!g)J(!1)}}return U1(),()=>{g=!0}},[$,_]),K`
        <${h2} className="attachment-preview-portal-root">
            <div class="image-modal attachment-preview-modal" onClick=${q}>
                <div class="attachment-preview-shell" onClick=${(g)=>{g.stopPropagation()}}>
                    <div class="attachment-preview-header">
                        <div class="attachment-preview-heading">
                            <div class="attachment-preview-title">${B}</div>
                            <div class="attachment-preview-subtitle">${Q}</div>
                        </div>
                        <div class="attachment-preview-header-actions">
                            ${m&&K`
                                <a
                                    href=${m}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="attachment-preview-download"
                                    onClick=${(g)=>g.stopPropagation()}
                                >
                                    Open in Tab
                                </a>
                            `}
                            <a
                                href=${i0($)}
                                download=${B}
                                class="attachment-preview-download"
                                onClick=${(g)=>g.stopPropagation()}
                            >
                                Download
                            </a>
                            <button class="attachment-preview-close" type="button" onClick=${q}>Close</button>
                        </div>
                    </div>
                    <div class="attachment-preview-body">
                        ${X&&K`<div class="attachment-preview-state">Loading preview…</div>`}
                        ${!X&&C&&K`<div class="attachment-preview-state">${C}</div>`}
                        ${!X&&!C&&_==="image"&&K`
                            <img class="attachment-preview-image" src=${i0($)} alt=${B} />
                        `}
                        ${!X&&!C&&_==="video"&&K`
                            <video class="attachment-preview-video" src=${i0($)} controls autoplay style="max-width:100%;max-height:100%;" />
                        `}
                        ${!X&&!C&&_==="html"&&K`
                            <iframe class="attachment-preview-frame" srcdoc=${W||""} sandbox=${uq} title=${B}></iframe>
                        `}
                        ${!X&&!C&&(_==="pdf"||_==="office"||_==="drawio")&&m&&K`
                            <iframe class="attachment-preview-frame" src=${m} title=${B}></iframe>
                        `}
                        ${!X&&!C&&_==="drawio"&&K`
                            <div class="attachment-preview-readonly-note">Draw.io preview is read-only. Editing tools are disabled in this preview.</div>
                        `}
                        ${!X&&!C&&_==="archive"&&G&&K`
                            <div class="attachment-preview-archive">
                                <div class="attachment-preview-archive-summary">
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Files</span>
                                        <strong class="attachment-preview-archive-card-value">${G.summary.fileCount}</strong>
                                    </div>
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Folders</span>
                                        <strong class="attachment-preview-archive-card-value">${G.summary.directoryCount}</strong>
                                    </div>
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Compressed</span>
                                        <strong class="attachment-preview-archive-card-value">${C0(G.summary.compressedBytes)}</strong>
                                    </div>
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Uncompressed</span>
                                        <strong class="attachment-preview-archive-card-value">${C0(G.summary.uncompressedBytes)}</strong>
                                    </div>
                                </div>
                                <div class="attachment-preview-archive-table-wrap">
                                    <table class="attachment-preview-archive-table">
                                        <thead>
                                            <tr>
                                                <th>Name</th>
                                                <th>Type</th>
                                                <th>Method</th>
                                                <th>Compressed</th>
                                                <th>Size</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            ${G.entries.map((g)=>K`
                                                <tr key=${g.path}>
                                                    <td class="attachment-preview-archive-name">${g.path}</td>
                                                    <td>${g.isDirectory?"Folder":"File"}</td>
                                                    <td>${g.isDirectory?"—":f$(g.compressionMethod)}</td>
                                                    <td>${g.isDirectory?"—":C0(g.compressedSize)}</td>
                                                    <td>${g.isDirectory?"—":C0(g.uncompressedSize)}</td>
                                                </tr>
                                            `)}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        `}
                        ${!X&&!C&&_==="text"&&U&&K`
                            <div
                                ref=${c}
                                class="attachment-preview-markdown post-content"
                                dangerouslySetInnerHTML=${{__html:O}}
                            />
                        `}
                        ${!X&&!C&&_==="text"&&!U&&d&&K`
                            <pre class="attachment-preview-text attachment-preview-code"><code dangerouslySetInnerHTML=${{__html:d}} /></pre>
                        `}
                        ${!X&&!C&&_==="text"&&!U&&!d&&K`
                            <pre class="attachment-preview-text">${W}</pre>
                        `}
                        ${!X&&!C&&_==="unsupported"&&K`
                            <div class="attachment-preview-state">
                                Preview is not available for this file type yet. You can still download it directly.
                            </div>
                        `}
                    </div>
                    <div class="attachment-preview-meta">
                        ${S.map((g)=>K`
                            <div class="attachment-preview-meta-item" key=${g.label}>
                                <span class="attachment-preview-meta-label">${g.label}</span>
                                <span class="attachment-preview-meta-value">${g.value}</span>
                            </div>
                        `)}
                    </div>
                </div>
            </div>
        </${h2}>
    `}var rq="default",x$="gi_session_id",dq=1200,r2="gi";function u3($){return`gi:${$}`}async function iq(){let $=E0(x$);if($)try{if((await fetch(`/api/sessions/${encodeURIComponent($)}`)).ok)return $}catch{}let j=await fetch("/api/sessions",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({title:rq})});if(!j.ok)throw Error("Failed to create default session");let q=await j.json();return R0(x$,q.id),q.id}async function nq(){let $=await fetch("/api/runtime/config");if(!$.ok)return{};return $.json()}function sq(){let[$,j]=y(!1),[q,B]=y(null),[_,Q]=y({}),[U,X]=y({}),[J,W]=y(null),[V,G]=y(!1),[L,C]=y([]),[b,c]=y(null),l=L.length>0,[w,S]=y([]),[m,O]=y(!1),d=M(null),[g,U1]=y([]),[k,r]=y([]),[f,_1]=y([]),[Y1,k1]=y(null),[O1,s]=y(null),[t,K1]=y(null),[G1,I]=y(""),[a,F1]=y(null),[X1,T1]=y(""),[u1,A1]=y("connected"),[t1,M1]=y(!1),G0=M(!1),{agentStatus:l1,setAgentStatus:I0,agentDraft:j0,setAgentDraft:g0,agentPlan:F0,setAgentPlan:q0,agentThought:B0,setAgentThought:x0,pendingRequest:S1,setPendingRequest:X0,currentTurnId:H0,setCurrentTurnId:_0,steerQueuedTurnId:Q0,setSteerQueuedTurnId:h1,lastAgentEventRef:Y0,draftBufferRef:E1,thoughtBufferRef:R1,pendingRequestRef:w1,currentTurnIdRef:e1,steerQueuedTurnIdRef:f1,thoughtExpandedRef:A0,draftExpandedRef:d1}=L8(),f0=B1(()=>q?u3(q):"",[q]);E(()=>{let j1=T8();return Promise.all([iq(),nq()]).then(([H1,H])=>{B(H1),Q(H),W({name:H.user_name,avatarUrl:H.user_avatar}),X({[r2]:{id:r2,name:H.assistant_name||"Gi",avatar_url:H.assistant_avatar||null}}),I(H.default_model||""),T1(H.default_thinking_level||""),j(!0)}).catch((H1)=>{console.error("[gi] Bootstrap failed:",H1)}),j1},[]);let x1=i(async(j1={})=>{if(!q)return;let H1=u3(q),p=(await f8(50,j1.beforeId||null,H1)).posts||[];if(j1.beforeId)S((v)=>d5([...p,...v]));else S(d5(p));O(p.length>=50)},[q]),P1=i(()=>{let j1=d.current;if(!j1)return;j1.scrollTop=j1.scrollHeight},[]);E(()=>{if(!$||!q)return;x1();let j1=setInterval(async()=>{await x1();let H1=u3(q),H=await P8(r2,H1).catch(()=>null);if(H){I0(H);let p=H?.status==="running"||H?.status==="cancelling";M1(p),G0.current=p}else I0(null),M1(!1),G0.current=!1},dq);return()=>clearInterval(j1)},[$,q]);let r1=i(async(j1)=>{await x1(),P1()},[x1,P1]),$0=i((j1)=>{let H1=L.find((H)=>H.id===j1||H.path===j1);if(H1){c(H1.id);return}C((H)=>[...H,{id:j1,path:j1,label:j1.split("/").pop()||j1,dirty:!1,pinned:!1}]),c(j1)},[L]),Z0=i((j1)=>{C((H1)=>{let H=H1.filter((p)=>p.id!==j1);if(b===j1)c(H[H.length-1]?.id||null);return H})},[b]),m1=["app-shell",V?"":"workspace-collapsed",l?"editor-open":""].filter(Boolean).join(" ");if(!$)return K`<div id="app"><div style="padding:20px;text-align:center;color:var(--text-secondary,#888)">Loading…</div></div>`;return K`
        <div class=${m1}>
            <${A$}
                onFileSelect=${(j1)=>U1((H1)=>[...H1,j1])}
                visible=${V}
                active=${V||l}
                onOpenEditor=${$0}
                onOpenTerminalTab=${()=>{}}
                onOpenVncTab=${()=>{}}
            />
            <button
                class=${`workspace-toggle-tab${V?" open":" closed"}`}
                onClick=${()=>G((j1)=>!j1)}
                title=${V?"Hide workspace":"Show workspace"}
            >
                <svg class="workspace-toggle-tab-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="6 3 11 8 6 13" />
                </svg>
            </button>
            <div class="workspace-splitter"></div>
            ${l&&K`
                <div class="editor-pane-container">
                    <${T$}
                        tabs=${L}
                        activeId=${b}
                        onActivate=${(j1)=>c(j1)}
                        onClose=${Z0}
                        onCloseOthers=${(j1)=>C((H1)=>H1.filter((H)=>H.id===j1))}
                        onCloseAll=${()=>{C([]),c(null)}}
                        onTogglePin=${()=>{}}
                    />
                    <div class="editor-pane-host"></div>
                </div>
                <div class="editor-splitter"></div>
            `}
            <div class="container">
                <${v6}
                    posts=${w}
                    hasMore=${m}
                    onLoadMore=${({preserveScroll:j1})=>{let H1=w[0];if(H1)x1({beforeId:H1.id})}}
                    timelineRef=${d}
                    onHashtagClick=${()=>{}}
                    onMessageRef=${()=>{}}
                    onScrollToMessage=${()=>{}}
                    onFileRef=${$0}
                    onPostClick=${void 0}
                    onDeletePost=${()=>{}}
                    onOpenWidget=${(j1)=>k1(j1)}
                    onOpenAttachmentPreview=${s}
                    emptyMessage="Send a message to get started."
                    agents=${U}
                    user=${J}
                    reverse=${!0}
                    removingPostIds=${new Set}
                    searchQuery=""
                />
                <${e6}
                    status=${K2(l1)?null:l1}
                    draft=${j0}
                    plan=${F0}
                    thought=${B0}
                    pendingRequest=${S1}
                    intent=${null}
                    turnId=${H0}
                    steerQueued=${Boolean(Q0)}
                    onPanelToggle=${()=>{}}
                    showExtensionPanels=${!1}
                />
                <${k$}
                    widget=${Y1}
                    onClose=${()=>k1(null)}
                    onWidgetEvent=${()=>{}}
                />
                ${O1&&K`
                    <${w$}
                        mediaId=${O1.mediaId}
                        info=${O1.info}
                        onClose=${()=>s(null)}
                    />
                `}
                <${T3}
                    items=${f}
                    onInjectQueuedFollowup=${()=>{}}
                    onRemoveQueuedFollowup=${()=>{}}
                    onMoveQueuedFollowup=${()=>{}}
                    onOpenFilePill=${$0}
                />
                <${i6}
                    currentChatJid=${f0}
                    isAgentActive=${t1}
                    onPost=${r1}
                    onFocus=${()=>{if(!y8())P1()}}
                    agents=${U}
                    currentSessionAgent=${U[r2]?{...U[r2],chat_jid:f0,agent_name:U[r2].name}:null}
                    agentStatus=${l1}
                    agentDraft=${j0}
                    contextUsage=${t}
                    fileRefs=${g}
                    messageRefs=${k}
                    onRemoveFileRef=${(j1)=>U1((H1)=>H1.filter((H)=>H!==j1))}
                    onClearFileRefs=${()=>U1([])}
                    onRemoveMessageRef=${()=>{}}
                    onClearMessageRefs=${()=>r([])}
                    connectionStatus=${u1}
                    activeChatAgents=${[]}
                    currentChatBranches=${[]}
                    formatBranchPickerLabel=${(j1)=>j1?.label||j1?.chat_jid||""}
                    handleBranchPickerChange=${()=>{}}
                    searchOpen=${!1}
                    onEnterSearch=${()=>{}}
                    onExitSearch=${()=>{}}
                    onSearch=${()=>{}}
                    searchScope="current"
                    onSearchScopeChange=${()=>{}}
                    activeModel=${G1}
                    agentModelsPayload=${a}
                    activeModelUsage=${null}
                    activeThinkingLevel=${X1}
                    supportsThinking=${!1}
                    followupQueueCount=${f.length}
                    notificationsEnabled=${!1}
                    notificationPermission="default"
                    onToggleNotifications=${()=>{}}
                    onComposeSubmitError=${()=>{}}
                    pendingRequestRef=${w1}
                    setPendingRequest=${X0}
                />
            </div>
        </div>
    `}m2(K`<${sq} />`,document.getElementById("app"));

//# debugId=B30E117E7EB7E18A64756E2164756E21
//# sourceMappingURL=app.js.map
