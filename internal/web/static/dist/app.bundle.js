var w4,x1,u3,z9,J2,T3,g3,m3,I5,y4,s2,h3,k5,T5,M5,W9,E4={},P4=[],G9=/acit|ex(?:s|g|n|p|$)|rph|grid|ows|mnc|ntw|ine[ch]|zoo|^ord|itera/i,x4=Array.isArray;function a0($,j){for(var q in j)$[q]=j[q];return $}function R5($){$&&$.parentNode&&$.parentNode.removeChild($)}function E5($,j,q){var B,Q,J,K={};for(J in j)J=="key"?B=j[J]:J=="ref"?Q=j[J]:K[J]=j[J];if(arguments.length>2&&(K.children=arguments.length>3?w4.call(arguments,2):q),typeof $=="function"&&$.defaultProps!=null)for(J in $.defaultProps)K[J]===void 0&&(K[J]=$.defaultProps[J]);return k4($,K,B,Q,null)}function k4($,j,q,B,Q){var J={type:$,props:j,key:q,ref:B,__k:null,__:null,__b:0,__e:null,__c:null,constructor:void 0,__v:Q==null?++u3:Q,__i:-1,__u:0};return Q==null&&x1.vnode!=null&&x1.vnode(J),J}function b4($){return $.children}function o2($,j){this.props=$,this.context=j}function w2($,j){if(j==null)return $.__?w2($.__,$.__i+1):null;for(var q;j<$.__k.length;j++)if((q=$.__k[j])!=null&&q.__e!=null)return q.__e;return typeof $.type=="function"?w2($):null}function N9($){if($.__P&&$.__d){var j=$.__v,q=j.__e,B=[],Q=[],J=a0({},j);J.__v=j.__v+1,x1.vnode&&x1.vnode(J),P5($.__P,J,j,$.__n,$.__P.namespaceURI,32&j.__u?[q]:null,B,q==null?w2(j):q,!!(32&j.__u),Q),J.__v=j.__v,J.__.__k[J.__i]=J,l3(B,J,Q),j.__e=j.__=null,J.__e!=q&&p3(J)}}function p3($){if(($=$.__)!=null&&$.__c!=null)return $.__e=$.__c.base=null,$.__k.some(function(j){if(j!=null&&j.__e!=null)return $.__e=$.__c.base=j.__e}),p3($)}function M3($){(!$.__d&&($.__d=!0)&&J2.push($)&&!f4.__r++||T3!=x1.debounceRendering)&&((T3=x1.debounceRendering)||g3)(f4)}function f4(){try{for(var $,j=1;J2.length;)J2.length>j&&J2.sort(m3),$=J2.shift(),j=J2.length,N9($)}finally{J2.length=f4.__r=0}}function c3($,j,q,B,Q,J,K,X,z,U,L){var N,V,O,g,h,p,d,P=B&&B.__k||P4,t=j.length;for(z=L9(q,j,P,z,t),N=0;N<t;N++)(O=q.__k[N])!=null&&(V=O.__i!=-1&&P[O.__i]||E4,O.__i=N,p=P5($,O,V,Q,J,K,X,z,U,L),g=O.__e,O.ref&&V.ref!=O.ref&&(V.ref&&f5(V.ref,null,O),L.push(O.ref,O.__c||g,O)),h==null&&g!=null&&(h=g),(d=!!(4&O.__u))||V.__k===O.__k?(z=d3(O,z,$,d),d&&V.__e&&(V.__e=null)):typeof O.type=="function"&&p!==void 0?z=p:g&&(z=g.nextSibling),O.__u&=-7);return q.__e=h,z}function L9($,j,q,B,Q){var J,K,X,z,U,L=q.length,N=L,V=0;for($.__k=Array(Q),J=0;J<Q;J++)(K=j[J])!=null&&typeof K!="boolean"&&typeof K!="function"?(typeof K=="string"||typeof K=="number"||typeof K=="bigint"||K.constructor==String?K=$.__k[J]=k4(null,K,null,null,null):x4(K)?K=$.__k[J]=k4(b4,{children:K},null,null,null):K.constructor===void 0&&K.__b>0?K=$.__k[J]=k4(K.type,K.props,K.key,K.ref?K.ref:null,K.__v):$.__k[J]=K,z=J+V,K.__=$,K.__b=$.__b+1,X=null,(U=K.__i=V9(K,q,z,N))!=-1&&(N--,(X=q[U])&&(X.__u|=2)),X==null||X.__v==null?(U==-1&&(Q>L?V--:Q<L&&V++),typeof K.type!="function"&&(K.__u|=4)):U!=z&&(U==z-1?V--:U==z+1?V++:(U>z?V--:V++,K.__u|=4))):$.__k[J]=null;if(N)for(J=0;J<L;J++)(X=q[J])!=null&&(2&X.__u)==0&&(X.__e==B&&(B=w2(X)),i3(X,X));return B}function d3($,j,q,B){var Q,J;if(typeof $.type=="function"){for(Q=$.__k,J=0;Q&&J<Q.length;J++)Q[J]&&(Q[J].__=$,j=d3(Q[J],j,q,B));return j}$.__e!=j&&(B&&(j&&$.type&&!j.parentNode&&(j=w2($)),q.insertBefore($.__e,j||null)),j=$.__e);do j=j&&j.nextSibling;while(j!=null&&j.nodeType==8);return j}function V9($,j,q,B){var Q,J,K,X=$.key,z=$.type,U=j[q],L=U!=null&&(2&U.__u)==0;if(U===null&&X==null||L&&X==U.key&&z==U.type)return q;if(B>(L?1:0)){for(Q=q-1,J=q+1;Q>=0||J<j.length;)if((U=j[K=Q>=0?Q--:J++])!=null&&(2&U.__u)==0&&X==U.key&&z==U.type)return K}return-1}function S3($,j,q){j[0]=="-"?$.setProperty(j,q==null?"":q):$[j]=q==null?"":typeof q!="number"||G9.test(j)?q:q+"px"}function S4($,j,q,B,Q){var J,K;$:if(j=="style")if(typeof q=="string")$.style.cssText=q;else{if(typeof B=="string"&&($.style.cssText=B=""),B)for(j in B)q&&j in q||S3($.style,j,"");if(q)for(j in q)B&&q[j]==B[j]||S3($.style,j,q[j])}else if(j[0]=="o"&&j[1]=="n")J=j!=(j=j.replace(h3,"$1")),K=j.toLowerCase(),j=K in $||j=="onFocusOut"||j=="onFocusIn"?K.slice(2):j.slice(2),$.l||($.l={}),$.l[j+J]=q,q?B?q[s2]=B[s2]:(q[s2]=k5,$.addEventListener(j,J?M5:T5,J)):$.removeEventListener(j,J?M5:T5,J);else{if(Q=="http://www.w3.org/2000/svg")j=j.replace(/xlink(H|:h)/,"h").replace(/sName$/,"s");else if(j!="width"&&j!="height"&&j!="href"&&j!="list"&&j!="form"&&j!="tabIndex"&&j!="download"&&j!="rowSpan"&&j!="colSpan"&&j!="role"&&j!="popover"&&j in $)try{$[j]=q==null?"":q;break $}catch(X){}typeof q=="function"||(q==null||q===!1&&j[4]!="-"?$.removeAttribute(j):$.setAttribute(j,j=="popover"&&q==1?"":q))}}function y3($){return function(j){if(this.l){var q=this.l[j.type+$];if(j[y4]==null)j[y4]=k5++;else if(j[y4]<q[s2])return;return q(x1.event?x1.event(j):j)}}}function P5($,j,q,B,Q,J,K,X,z,U){var L,N,V,O,g,h,p,d,P,t,v,e,W1,D1,N1,X1=j.type;if(j.constructor!==void 0)return null;128&q.__u&&(z=!!(32&q.__u),J=[X=j.__e=q.__e]),(L=x1.__b)&&L(j);$:if(typeof X1=="function")try{if(d=j.props,P=X1.prototype&&X1.prototype.render,t=(L=X1.contextType)&&B[L.__c],v=L?t?t.props.value:L.__:B,q.__c?p=(N=j.__c=q.__c).__=N.__E:(P?j.__c=N=new X1(d,v):(j.__c=N=new o2(d,v),N.constructor=X1,N.render=F9),t&&t.sub(N),N.state||(N.state={}),N.__n=B,V=N.__d=!0,N.__h=[],N._sb=[]),P&&N.__s==null&&(N.__s=N.state),P&&X1.getDerivedStateFromProps!=null&&(N.__s==N.state&&(N.__s=a0({},N.__s)),a0(N.__s,X1.getDerivedStateFromProps(d,N.__s))),O=N.props,g=N.state,N.__v=j,V)P&&X1.getDerivedStateFromProps==null&&N.componentWillMount!=null&&N.componentWillMount(),P&&N.componentDidMount!=null&&N.__h.push(N.componentDidMount);else{if(P&&X1.getDerivedStateFromProps==null&&d!==O&&N.componentWillReceiveProps!=null&&N.componentWillReceiveProps(d,v),j.__v==q.__v||!N.__e&&N.shouldComponentUpdate!=null&&N.shouldComponentUpdate(d,N.__s,v)===!1){j.__v!=q.__v&&(N.props=d,N.state=N.__s,N.__d=!1),j.__e=q.__e,j.__k=q.__k,j.__k.some(function(x){x&&(x.__=j)}),P4.push.apply(N.__h,N._sb),N._sb=[],N.__h.length&&K.push(N);break $}N.componentWillUpdate!=null&&N.componentWillUpdate(d,N.__s,v),P&&N.componentDidUpdate!=null&&N.__h.push(function(){N.componentDidUpdate(O,g,h)})}if(N.context=v,N.props=d,N.__P=$,N.__e=!1,e=x1.__r,W1=0,P)N.state=N.__s,N.__d=!1,e&&e(j),L=N.render(N.props,N.state,N.context),P4.push.apply(N.__h,N._sb),N._sb=[];else do N.__d=!1,e&&e(j),L=N.render(N.props,N.state,N.context),N.state=N.__s;while(N.__d&&++W1<25);N.state=N.__s,N.getChildContext!=null&&(B=a0(a0({},B),N.getChildContext())),P&&!V&&N.getSnapshotBeforeUpdate!=null&&(h=N.getSnapshotBeforeUpdate(O,g)),D1=L!=null&&L.type===b4&&L.key==null?r3(L.props.children):L,X=c3($,x4(D1)?D1:[D1],j,q,B,Q,J,K,X,z,U),N.base=j.__e,j.__u&=-161,N.__h.length&&K.push(N),p&&(N.__E=N.__=null)}catch(x){if(j.__v=null,z||J!=null)if(x.then){for(j.__u|=z?160:128;X&&X.nodeType==8&&X.nextSibling;)X=X.nextSibling;J[J.indexOf(X)]=null,j.__e=X}else{for(N1=J.length;N1--;)R5(J[N1]);S5(j)}else j.__e=q.__e,j.__k=q.__k,x.then||S5(j);x1.__e(x,j,q)}else J==null&&j.__v==q.__v?(j.__k=q.__k,j.__e=q.__e):X=j.__e=Z9(q.__e,j,q,B,Q,J,K,z,U);return(L=x1.diffed)&&L(j),128&j.__u?void 0:X}function S5($){$&&($.__c&&($.__c.__e=!0),$.__k&&$.__k.some(S5))}function l3($,j,q){for(var B=0;B<q.length;B++)f5(q[B],q[++B],q[++B]);x1.__c&&x1.__c(j,$),$.some(function(Q){try{$=Q.__h,Q.__h=[],$.some(function(J){J.call(Q)})}catch(J){x1.__e(J,Q.__v)}})}function r3($){return typeof $!="object"||$==null||$.__b>0?$:x4($)?$.map(r3):a0({},$)}function Z9($,j,q,B,Q,J,K,X,z){var U,L,N,V,O,g,h,p=q.props||E4,d=j.props,P=j.type;if(P=="svg"?Q="http://www.w3.org/2000/svg":P=="math"?Q="http://www.w3.org/1998/Math/MathML":Q||(Q="http://www.w3.org/1999/xhtml"),J!=null){for(U=0;U<J.length;U++)if((O=J[U])&&"setAttribute"in O==!!P&&(P?O.localName==P:O.nodeType==3)){$=O,J[U]=null;break}}if($==null){if(P==null)return document.createTextNode(d);$=document.createElementNS(Q,P,d.is&&d),X&&(x1.__m&&x1.__m(j,J),X=!1),J=null}if(P==null)p===d||X&&$.data==d||($.data=d);else{if(J=J&&w4.call($.childNodes),!X&&J!=null)for(p={},U=0;U<$.attributes.length;U++)p[(O=$.attributes[U]).name]=O.value;for(U in p)O=p[U],U=="dangerouslySetInnerHTML"?N=O:U=="children"||(U in d)||U=="value"&&("defaultValue"in d)||U=="checked"&&("defaultChecked"in d)||S4($,U,null,O,Q);for(U in d)O=d[U],U=="children"?V=O:U=="dangerouslySetInnerHTML"?L=O:U=="value"?g=O:U=="checked"?h=O:X&&typeof O!="function"||p[U]===O||S4($,U,O,p[U],Q);if(L)X||N&&(L.__html==N.__html||L.__html==$.innerHTML)||($.innerHTML=L.__html),j.__k=[];else if(N&&($.innerHTML=""),c3(j.type=="template"?$.content:$,x4(V)?V:[V],j,q,B,P=="foreignObject"?"http://www.w3.org/1999/xhtml":Q,J,K,J?J[0]:q.__k&&w2(q,0),X,z),J!=null)for(U=J.length;U--;)R5(J[U]);X||(U="value",P=="progress"&&g==null?$.removeAttribute("value"):g!=null&&(g!==$[U]||P=="progress"&&!g||P=="option"&&g!=p[U])&&S4($,U,g,p[U],Q),U="checked",h!=null&&h!=$[U]&&S4($,U,h,p[U],Q))}return $}function f5($,j,q){try{if(typeof $=="function"){var B=typeof $.__u=="function";B&&$.__u(),B&&j==null||($.__u=$(j))}else $.current=j}catch(Q){x1.__e(Q,q)}}function i3($,j,q){var B,Q;if(x1.unmount&&x1.unmount($),(B=$.ref)&&(B.current&&B.current!=$.__e||f5(B,null,j)),(B=$.__c)!=null){if(B.componentWillUnmount)try{B.componentWillUnmount()}catch(J){x1.__e(J,j)}B.base=B.__P=null}if(B=$.__k)for(Q=0;Q<B.length;Q++)B[Q]&&i3(B[Q],j,q||typeof $.type!="function");q||R5($.__e),$.__c=$.__=$.__e=void 0}function F9($,j,q){return this.constructor($,q)}function b2($,j,q){var B,Q,J,K;j==document&&(j=document.documentElement),x1.__&&x1.__($,j),Q=(B=typeof q=="function")?null:q&&q.__k||j.__k,J=[],K=[],P5(j,$=(!B&&q||j).__k=E5(b4,null,[$]),Q||E4,E4,j.namespaceURI,!B&&q?[q]:Q?null:j.firstChild?w4.call(j.childNodes):null,J,!B&&q?q:Q?Q.__e:j.firstChild,B,K),l3(J,$,K)}w4=P4.slice,x1={__e:function($,j,q,B){for(var Q,J,K;j=j.__;)if((Q=j.__c)&&!Q.__)try{if((J=Q.constructor)&&J.getDerivedStateFromError!=null&&(Q.setState(J.getDerivedStateFromError($)),K=Q.__d),Q.componentDidCatch!=null&&(Q.componentDidCatch($,B||{}),K=Q.__d),K)return Q.__E=Q}catch(X){$=X}throw $}},u3=0,z9=function($){return $!=null&&$.constructor===void 0},o2.prototype.setState=function($,j){var q;q=this.__s!=null&&this.__s!=this.state?this.__s:this.__s=a0({},this.state),typeof $=="function"&&($=$(a0({},q),this.props)),$&&a0(q,$),$!=null&&this.__v&&(j&&this._sb.push(j),M3(this))},o2.prototype.forceUpdate=function($){this.__v&&(this.__e=!0,$&&this.__h.push($),M3(this))},o2.prototype.render=b4,J2=[],g3=typeof Promise=="function"?Promise.prototype.then.bind(Promise.resolve()):setTimeout,m3=function($,j){return $.__v.__b-j.__v.__b},f4.__r=0,I5=Math.random().toString(8),y4="__d"+I5,s2="__a"+I5,h3=/(PointerCapture)$|Capture$/i,k5=0,T5=y3(!1),M5=y3(!0),W9=0;var x2,n1,O5,k3,a2=0,n3=[],i1=x1,R3=i1.__b,E3=i1.__r,P3=i1.diffed,f3=i1.__c,w3=i1.unmount,x3=i1.__;function v4($,j){i1.__h&&i1.__h(n1,$,a2||j),a2=0;var q=n1.__H||(n1.__H={__:[],__h:[]});return $>=q.__.length&&q.__.push({}),q.__[$]}function M($){return a2=1,s3(o3,$)}function s3($,j,q){var B=v4(x2++,2);if(B.t=$,!B.__c&&(B.__=[q?q(j):o3(void 0,j),function(X){var z=B.__N?B.__N[0]:B.__[0],U=B.t(z,X);z!==U&&(B.__N=[U,B.__[1]],B.__c.setState({}))}],B.__c=n1,!n1.__f)){var Q=function(X,z,U){if(!B.__c.__H)return!0;var L=B.__c.__H.__.filter(function(V){return V.__c});if(L.every(function(V){return!V.__N}))return!J||J.call(this,X,z,U);var N=B.__c.props!==X;return L.some(function(V){if(V.__N){var O=V.__[0];V.__=V.__N,V.__N=void 0,O!==V.__[0]&&(N=!0)}}),J&&J.call(this,X,z,U)||N};n1.__f=!0;var{shouldComponentUpdate:J,componentWillUpdate:K}=n1;n1.componentWillUpdate=function(X,z,U){if(this.__e){var L=J;J=void 0,Q(X,z,U),J=L}K&&K.call(this,X,z,U)},n1.shouldComponentUpdate=Q}return B.__N||B.__}function w($,j){var q=v4(x2++,3);!i1.__s&&x5(q.__H,j)&&(q.__=$,q.u=j,n1.__H.__h.push(q))}function w5($,j){var q=v4(x2++,4);!i1.__s&&x5(q.__H,j)&&(q.__=$,q.u=j,n1.__h.push(q))}function T($){return a2=5,H1(function(){return{current:$}},[])}function H1($,j){var q=v4(x2++,7);return x5(q.__H,j)&&(q.__=$(),q.__H=j,q.__h=$),q.__}function m($,j){return a2=8,H1(function(){return $},j)}function H9(){for(var $;$=n3.shift();){var j=$.__H;if($.__P&&j)try{j.__h.some(R4),j.__h.some(y5),j.__h=[]}catch(q){j.__h=[],i1.__e(q,$.__v)}}}i1.__b=function($){n1=null,R3&&R3($)},i1.__=function($,j){$&&j.__k&&j.__k.__m&&($.__m=j.__k.__m),x3&&x3($,j)},i1.__r=function($){E3&&E3($),x2=0;var j=(n1=$.__c).__H;j&&(O5===n1?(j.__h=[],n1.__h=[],j.__.some(function(q){q.__N&&(q.__=q.__N),q.u=q.__N=void 0})):(j.__h.some(R4),j.__h.some(y5),j.__h=[],x2=0)),O5=n1},i1.diffed=function($){P3&&P3($);var j=$.__c;j&&j.__H&&(j.__H.__h.length&&(n3.push(j)!==1&&k3===i1.requestAnimationFrame||((k3=i1.requestAnimationFrame)||Y9)(H9)),j.__H.__.some(function(q){q.u&&(q.__H=q.u),q.u=void 0})),O5=n1=null},i1.__c=function($,j){j.some(function(q){try{q.__h.some(R4),q.__h=q.__h.filter(function(B){return!B.__||y5(B)})}catch(B){j.some(function(Q){Q.__h&&(Q.__h=[])}),j=[],i1.__e(B,q.__v)}}),f3&&f3($,j)},i1.unmount=function($){w3&&w3($);var j,q=$.__c;q&&q.__H&&(q.__H.__.some(function(B){try{R4(B)}catch(Q){j=Q}}),q.__H=void 0,j&&i1.__e(j,q.__v))};var b3=typeof requestAnimationFrame=="function";function Y9($){var j,q=function(){clearTimeout(B),b3&&cancelAnimationFrame(j),setTimeout($)},B=setTimeout(q,35);b3&&(j=requestAnimationFrame(q))}function R4($){var j=n1,q=$.__c;typeof q=="function"&&($.__c=void 0,q()),n1=j}function y5($){var j=n1;$.__c=$.__(),n1=j}function x5($,j){return!$||$.length!==j.length||j.some(function(q,B){return q!==$[B]})}function o3($,j){return typeof j=="function"?j($):j}var a3=function($,j,q,B){var Q;j[0]=0;for(var J=1;J<j.length;J++){var K=j[J++],X=j[J]?(j[0]|=K?1:2,q[j[J++]]):j[++J];K===3?B[0]=X:K===4?B[1]=Object.assign(B[1]||{},X):K===5?(B[1]=B[1]||{})[j[++J]]=X:K===6?B[1][j[++J]]+=X+"":K?(Q=$.apply(X,a3($,X,q,["",null])),B.push(Q),X[0]?j[0]|=2:(j[J-2]=0,j[J]=Q)):B.push(X)}return B},v3=new Map;function A9($){var j=v3.get(this);return j||(j=new Map,v3.set(this,j)),(j=a3(this,j.get($)||(j.set($,j=function(q){for(var B,Q,J=1,K="",X="",z=[0],U=function(V){J===1&&(V||(K=K.replace(/^\s*\n\s*|\s*\n\s*$/g,"")))?z.push(0,V,K):J===3&&(V||K)?(z.push(3,V,K),J=2):J===2&&K==="..."&&V?z.push(4,V,0):J===2&&K&&!V?z.push(5,0,!0,K):J>=5&&((K||!V&&J===5)&&(z.push(J,0,K,Q),J=6),V&&(z.push(J,V,0,Q),J=6)),K=""},L=0;L<q.length;L++){L&&(J===1&&U(),U(L));for(var N=0;N<q[L].length;N++)B=q[L][N],J===1?B==="<"?(U(),z=[z],J=3):K+=B:J===4?K==="--"&&B===">"?(J=1,K=""):K=B+K[0]:X?B===X?X="":K+=B:B==='"'||B==="'"?X=B:B===">"?(U(),J=1):J&&(B==="="?(J=5,Q=K,K=""):B==="/"&&(J<5||q[L][N+1]===">")?(U(),J===3&&(z=z[0]),J=z,(z=z[0]).push(2,0,J),J=0):B===" "||B==="\t"||B===`
`||B==="\r"?(U(),J=2):K+=B),J===3&&K==="!--"&&(J=4,z=z[0])}return U(),z}($)),j),arguments,[])).length>1?j:j[0]}var G=A9.bind(E5);function k0($){if(typeof window>"u"||!window.localStorage)return null;try{return window.localStorage.getItem($)}catch{return null}}function E0($,j){if(typeof window>"u"||!window.localStorage)return;try{window.localStorage.setItem($,j)}catch{return}}function t3($,j=!1){let q=k0($);if(q===null)return j;return q==="true"}function e3($,j=null){let q=k0($);if(q===null)return j;let B=parseInt(q,10);return Number.isFinite(B)?B:j}var b5=($)=>{let j=new Set;return($||[]).filter((q)=>{if(!q||j.has(q.id))return!1;return j.add(q.id),!0})};function $6(){let[$,j]=M(null),[q,B]=M({text:"",totalLines:0}),[Q,J]=M(""),[K,X]=M({text:"",totalLines:0}),[z,U]=M(null),[L,N]=M(null),[V,O]=M(null),g=T(null),h=T(0),p=T(!1),d=T(""),P=T(""),t=T(!1),v=T(0),e=T(null),W1=T(null),D1=T(null),N1=T(null),X1=T(!1),x=T(!1);return{agentStatus:$,setAgentStatus:j,agentDraft:q,setAgentDraft:B,agentPlan:Q,setAgentPlan:J,agentThought:K,setAgentThought:X,pendingRequest:z,setPendingRequest:U,currentTurnId:L,setCurrentTurnId:N,steerQueuedTurnId:V,setSteerQueuedTurnId:O,lastAgentEventRef:g,lastSilenceNoticeRef:h,isAgentRunningRef:p,draftBufferRef:d,thoughtBufferRef:P,previewResyncPendingRef:t,previewResyncGenerationRef:v,pendingRequestRef:e,stalledPostIdRef:W1,currentTurnIdRef:D1,steerQueuedTurnIdRef:N1,thoughtExpandedRef:X1,draftExpandedRef:x}}var B6="piclaw_theme",u5="piclaw_tint",D9="piclaw_chat_themes",e0={bgPrimary:"#ffffff",bgSecondary:"#f7f9fa",bgHover:"#e8ebed",textPrimary:"#0f1419",textSecondary:"#536471",borderColor:"#eff3f4",accent:"#1d9bf0",accentHover:"#1a8cd8",warning:"#f0b429",danger:"#f4212e",success:"#00ba7c"},m4={bgPrimary:"#000000",bgSecondary:"#16181c",bgHover:"#1d1f23",textPrimary:"#e7e9ea",textSecondary:"#71767b",borderColor:"#2f3336",accent:"#1d9bf0",accentHover:"#1a8cd8",warning:"#f0b429",danger:"#f4212e",success:"#00ba7c"},j6={default:{label:"Default",mode:"auto",light:e0,dark:m4},tango:{label:"Tango",mode:"light",light:{bgPrimary:"#f6f5f4",bgSecondary:"#efedeb",bgHover:"#e5e3e1",textPrimary:"#2e3436",textSecondary:"#5c6466",borderColor:"#d3d7cf",accent:"#3465a4",accentHover:"#2c5890",danger:"#cc0000",success:"#4e9a06"}},xterm:{label:"XTerm",mode:"dark",dark:{bgPrimary:"#000000",bgSecondary:"#0a0a0a",bgHover:"#121212",textPrimary:"#d0d0d0",textSecondary:"#8a8a8a",borderColor:"#1f1f1f",accent:"#00a2ff",accentHover:"#0086d1",danger:"#ff5f5f",success:"#5fff87"}},monokai:{label:"Monokai",mode:"dark",dark:{bgPrimary:"#272822",bgSecondary:"#2f2f2f",bgHover:"#3a3a3a",textPrimary:"#f8f8f2",textSecondary:"#cfcfc2",borderColor:"#3e3d32",accent:"#f92672",accentHover:"#e81560",danger:"#f92672",success:"#a6e22e"}},"monokai-pro":{label:"Monokai Pro",mode:"dark",dark:{bgPrimary:"#2d2a2e",bgSecondary:"#363237",bgHover:"#403a40",textPrimary:"#fcfcfa",textSecondary:"#c1c0c0",borderColor:"#444046",accent:"#ff6188",accentHover:"#f74f7e",danger:"#ff4f5e",success:"#a9dc76"}},ristretto:{label:"Ristretto",mode:"dark",dark:{bgPrimary:"#2c2525",bgSecondary:"#362d2d",bgHover:"#403535",textPrimary:"#f4f1ef",textSecondary:"#cbbdb8",borderColor:"#4a3c3c",accent:"#ff9f43",accentHover:"#f28a2e",danger:"#ff5f56",success:"#a9dc76"}},dracula:{label:"Dracula",mode:"dark",dark:{bgPrimary:"#282a36",bgSecondary:"#303445",bgHover:"#3a3f52",textPrimary:"#f8f8f2",textSecondary:"#c5c8d6",borderColor:"#44475a",accent:"#bd93f9",accentHover:"#a87ded",danger:"#ff5555",success:"#50fa7b"}},catppuccin:{label:"Catppuccin",mode:"dark",dark:{bgPrimary:"#1e1e2e",bgSecondary:"#24273a",bgHover:"#2c2f41",textPrimary:"#cdd6f4",textSecondary:"#a6adc8",borderColor:"#313244",accent:"#89b4fa",accentHover:"#74a0f5",danger:"#f38ba8",success:"#a6e3a1"}},nord:{label:"Nord",mode:"dark",dark:{bgPrimary:"#2e3440",bgSecondary:"#3b4252",bgHover:"#434c5e",textPrimary:"#eceff4",textSecondary:"#d8dee9",borderColor:"#4c566a",accent:"#88c0d0",accentHover:"#78a9c0",danger:"#bf616a",success:"#a3be8c"}},gruvbox:{label:"Gruvbox",mode:"dark",dark:{bgPrimary:"#282828",bgSecondary:"#32302f",bgHover:"#3c3836",textPrimary:"#ebdbb2",textSecondary:"#bdae93",borderColor:"#3c3836",accent:"#d79921",accentHover:"#c28515",danger:"#fb4934",success:"#b8bb26"}},solarized:{label:"Solarized",mode:"auto",light:{bgPrimary:"#fdf6e3",bgSecondary:"#f5efdc",bgHover:"#eee8d5",textPrimary:"#586e75",textSecondary:"#657b83",borderColor:"#e0d8c6",accent:"#268bd2",accentHover:"#1f78b3",danger:"#dc322f",success:"#859900"},dark:{bgPrimary:"#002b36",bgSecondary:"#073642",bgHover:"#0b3c4a",textPrimary:"#eee8d5",textSecondary:"#93a1a1",borderColor:"#18424a",accent:"#268bd2",accentHover:"#1f78b3",danger:"#dc322f",success:"#859900"}},tokyo:{label:"Tokyo",mode:"dark",dark:{bgPrimary:"#1a1b26",bgSecondary:"#24283b",bgHover:"#2f3549",textPrimary:"#c0caf5",textSecondary:"#9aa5ce",borderColor:"#414868",accent:"#7aa2f7",accentHover:"#6b92e6",danger:"#f7768e",success:"#9ece6a"}},miasma:{label:"Miasma",mode:"dark",dark:{bgPrimary:"#1f1f23",bgSecondary:"#29292f",bgHover:"#33333a",textPrimary:"#e5e5e5",textSecondary:"#b4b4b4",borderColor:"#3d3d45",accent:"#c9739c",accentHover:"#b8618c",danger:"#e06c75",success:"#98c379"}},github:{label:"GitHub",mode:"auto",light:{bgPrimary:"#ffffff",bgSecondary:"#f6f8fa",bgHover:"#eaeef2",textPrimary:"#24292f",textSecondary:"#57606a",borderColor:"#d0d7de",accent:"#0969da",accentHover:"#0550ae",danger:"#cf222e",success:"#1a7f37"},dark:{bgPrimary:"#0d1117",bgSecondary:"#161b22",bgHover:"#21262d",textPrimary:"#c9d1d9",textSecondary:"#8b949e",borderColor:"#30363d",accent:"#2f81f7",accentHover:"#1f6feb",danger:"#f85149",success:"#3fb950"}},gotham:{label:"Gotham",mode:"dark",dark:{bgPrimary:"#0b0f14",bgSecondary:"#111720",bgHover:"#18212b",textPrimary:"#cbd6e2",textSecondary:"#9bb0c3",borderColor:"#1f2a37",accent:"#5ccfe6",accentHover:"#48b8ce",danger:"#d26937",success:"#2aa889"}}},C9=["--bg-primary","--bg-secondary","--bg-hover","--text-primary","--text-secondary","--border-color","--accent-color","--accent-hover","--accent-color-alpha","--accent-contrast-text","--accent-soft","--accent-soft-strong","--warning-color","--danger-color","--success-color","--search-highlight-color"],C2={theme:"default",tint:null},Q6="light",v5=!1;function h4($){let j=String($||"").trim().toLowerCase();if(!j)return"default";if(j==="solarized-dark"||j==="solarized-light")return"solarized";if(j==="github-dark"||j==="github-light")return"github";if(j==="tokyo-night")return"tokyo";return j}function J6($){if(!$)return null;let j=String($).trim();if(!j)return null;let q=j.startsWith("#")?j.slice(1):j;if(!/^[0-9a-fA-F]{3}$/.test(q)&&!/^[0-9a-fA-F]{6}$/.test(q))return null;let B=q.length===3?q.split("").map((J)=>J+J).join(""):q,Q=parseInt(B,16);return{r:Q>>16&255,g:Q>>8&255,b:Q&255,hex:`#${B.toLowerCase()}`}}function I9($,j){try{if(document.body){$.style.display="none",document.body.appendChild($);let q=getComputedStyle($).color||$.style.color;return document.body.removeChild($),q}}catch{return j}return j}function O9($){if(!$||typeof document>"u")return null;let j=String($).trim();if(!j)return null;let q=document.createElement("div");if(q.style.color="",q.style.color=j,!q.style.color)return null;let Q=I9(q,q.style.color).match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/i);if(!Q)return null;let J=parseInt(Q[1],10),K=parseInt(Q[2],10),X=parseInt(Q[3],10);if(![J,K,X].every((U)=>Number.isFinite(U)))return null;let z=`#${[J,K,X].map((U)=>U.toString(16).padStart(2,"0")).join("")}`;return{r:J,g:K,b:X,hex:z}}function D2($){return J6($)||O9($)}function g5($,j,q){let B=Math.round($.r+(j.r-$.r)*q),Q=Math.round($.g+(j.g-$.g)*q),J=Math.round($.b+(j.b-$.b)*q);return`rgb(${B} ${Q} ${J})`}function u4($,j){return`rgba(${$.r}, ${$.g}, ${$.b}, ${j})`}function T9($){let j=$.r/255,q=$.g/255,B=$.b/255,Q=j<=0.03928?j/12.92:Math.pow((j+0.055)/1.055,2.4),J=q<=0.03928?q/12.92:Math.pow((q+0.055)/1.055,2.4),K=B<=0.03928?B/12.92:Math.pow((B+0.055)/1.055,2.4);return 0.2126*Q+0.7152*J+0.0722*K}function M9($){return T9($)>0.4?"#000000":"#ffffff"}function K6(){if(typeof window>"u")return"light";try{return window.matchMedia&&window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light"}catch{return"light"}}function m5($){return j6[$]||j6.default}function S9($){return $.mode==="auto"?K6():$.mode}function X6($,j){let q=m5($);if(j==="dark"&&q.dark)return q.dark;if(j==="light"&&q.light)return q.light;return q.dark||q.light||e0}function t0($,j,q){let B=D2($);if(!B)return $;return g5(B,j,q)}function U6($,j,q){let B=D2(j);if(!B)return $;let J=J6(q==="dark"?"#ffffff":"#000000");return{...$,bgPrimary:t0($.bgPrimary,B,0.08),bgSecondary:t0($.bgSecondary,B,0.12),bgHover:t0($.bgHover,B,0.16),textPrimary:t0($.textPrimary,B,q==="dark"?0.08:0.06),textSecondary:t0($.textSecondary,B,q==="dark"?0.12:0.1),borderColor:t0($.borderColor,B,0.1),accent:B.hex,accentHover:J?g5(B,J,0.18):B.hex,warning:t0($.warning||e0.warning,B,0.14),danger:t0($.danger,B,0.16),success:t0($.success,B,0.16)}}function y9($,j){let q=D2($?.warning);if(q)return q.hex;let B=D2(j==="dark"?m4.warning:e0.warning)||D2(e0.warning),Q=D2($?.accent);if(B&&Q)return g5(B,Q,j==="dark"?0.18:0.14);return j==="dark"?m4.warning:e0.warning}function k9($,j){if(typeof document>"u")return;let q=document.documentElement,B=$.accent,Q=D2(B),J=Q?u4(Q,j==="dark"?0.35:0.2):$.searchHighlight||$.searchHighlightColor,K=Q?u4(Q,j==="dark"?0.16:0.12):"rgba(29, 155, 240, 0.12)",X=Q?u4(Q,j==="dark"?0.28:0.2):"rgba(29, 155, 240, 0.2)",z=Q?M9(Q):j==="dark"?"#000000":"#ffffff",U=Q?u4(Q,j==="dark"?0.35:0.25):"rgba(29, 155, 240, 0.25)",L=y9($,j),N={"--bg-primary":$.bgPrimary,"--bg-secondary":$.bgSecondary,"--bg-hover":$.bgHover,"--text-primary":$.textPrimary,"--text-secondary":$.textSecondary,"--border-color":$.borderColor,"--accent-color":B,"--accent-hover":$.accentHover||B,"--accent-color-alpha":U,"--accent-soft":K,"--accent-soft-strong":X,"--accent-contrast-text":z,"--warning-color":L,"--danger-color":$.danger||e0.danger,"--success-color":$.success||e0.success,"--search-highlight-color":J||"rgba(29, 155, 240, 0.2)"};Object.entries(N).forEach(([V,O])=>{if(O)q.style.setProperty(V,O)})}function R9(){if(typeof document>"u")return;let $=document.documentElement;C9.forEach((j)=>$.style.removeProperty(j))}function v2($,j={}){if(typeof document>"u")return null;let q=typeof j.id==="string"&&j.id.trim()?j.id.trim():null,B=q?document.getElementById(q):document.querySelector(`meta[name="${$}"]`);if(!B)B=document.createElement("meta"),document.head.appendChild(B);if(B.setAttribute("name",$),q)B.setAttribute("id",q);return B}function q6($){let j=h4(C2?.theme||"default"),q=C2?.tint?String(C2.tint).trim():null,B=X6(j,$);if(j==="default"&&q)B=U6(B,q,$);if(B?.bgPrimary)return B.bgPrimary;return $==="dark"?m4.bgPrimary:e0.bgPrimary}function E9($,j){if(typeof document>"u")return;let q=v2("theme-color",{id:"dynamic-theme-color"});if(q&&$)q.removeAttribute("media"),q.setAttribute("content",$);let B=v2("theme-color",{id:"theme-color-light"});if(B)B.setAttribute("media","(prefers-color-scheme: light)"),B.setAttribute("content",q6("light"));let Q=v2("theme-color",{id:"theme-color-dark"});if(Q)Q.setAttribute("media","(prefers-color-scheme: dark)"),Q.setAttribute("content",q6("dark"));let J=v2("msapplication-TileColor");if(J&&$)J.setAttribute("content",$);let K=v2("msapplication-navbutton-color");if(K&&$)K.setAttribute("content",$);let X=v2("apple-mobile-web-app-status-bar-style");if(X)X.setAttribute("content",j==="dark"?"black-translucent":"default")}function P9(){if(typeof window>"u")return;let $={...C2,mode:Q6};window.dispatchEvent(new CustomEvent("piclaw-theme-change",{detail:$}))}function f9(){try{let $=k0(D9);if(!$)return{};let j=JSON.parse($);return typeof j==="object"&&j!==null?j:{}}catch{return{}}}function w9($){if(!$)return null;return f9()[$]||null}function x9(){if(typeof window>"u")return"web:default";try{let j=new URL(window.location.href).searchParams.get("chat_jid");return j&&j.trim()?j.trim():"web:default"}catch{return"web:default"}}function _6($,j={}){if(typeof window>"u"||typeof document>"u")return;let q=h4($?.theme||"default"),B=$?.tint?String($.tint).trim():null,Q=m5(q),J=S9(Q),K=X6(q,J);C2={theme:q,tint:B},Q6=J;let X=document.documentElement;X.dataset.theme=J,X.dataset.colorTheme=q,X.dataset.tint=B?String(B):"",X.style.colorScheme=J;let z=K;if(q==="default"&&B)z=U6(K,B,J);if(q==="default"&&!B)R9();else k9(z,J);if(E9(z.bgPrimary,J),P9(),j.persist!==!1)if(E0(B6,q),B)E0(u5,B);else E0(u5,"")}function g4(){if(m5(C2.theme).mode!=="auto")return;_6(C2,{persist:!1})}function z6(){if(typeof window>"u")return()=>{};let $=x9(),j=w9($),q=j?h4(j.theme||"default"):h4(k0(B6)||"default"),B=j?j.tint?String(j.tint).trim():null:(()=>{let Q=k0(u5);return Q?Q.trim():null})();if(_6({theme:q,tint:B},{persist:!1}),window.matchMedia&&!v5){let Q=window.matchMedia("(prefers-color-scheme: dark)");if(Q.addEventListener)Q.addEventListener("change",g4);else if(Q.addListener)Q.addListener(g4);return v5=!0,()=>{if(Q.removeEventListener)Q.removeEventListener("change",g4);else if(Q.removeListener)Q.removeListener(g4);v5=!1}}return()=>{}}function W6(){if(typeof document>"u")return"light";let $=document.documentElement?.dataset?.theme;if($==="dark"||$==="light")return $;return K6()}var b9="";async function K2($,j={}){let q=typeof performance<"u"&&typeof performance.now==="function"?performance.now():Date.now(),B;try{B=await fetch(b9+$,{...j,headers:{"Content-Type":"application/json",...j.headers||{}}})}catch(J){throw G6({method:String(j.method||"GET").toUpperCase(),url:$,startedAt:q,durationMs:performance.now()-q,ok:!1,detail:{failedBeforeResponse:!0}}),J}let Q=performance.now()-q;if(G6({method:String(j.method||"GET").toUpperCase(),url:$,startedAt:q,durationMs:Q,status:B.status,ok:B.ok,requestId:B.headers?.get?.("x-request-id")||null,serverTiming:B.headers?.get?.("Server-Timing")||null}),!B.ok){let J=await B.json().catch(()=>({error:"Unknown error"}));throw Error(J.error||`HTTP ${B.status}`)}return B.json()}async function N6($=50,j=null,q=null){let B=q?.startsWith("gi:")?q.slice(3):null;if(!B)return{posts:[]};let Q=`/api/sessions/${encodeURIComponent(B)}/messages?limit=${$}`;if(j)Q+=`&before=${j}`;return{posts:((await K2(Q)).messages||[]).map((X)=>({id:X.id,chat_jid:q,content:X.content,timestamp:X.created_at,sender:X.role==="user"?"user":"agent",is_from_me:X.role==="user",is_bot_message:X.role==="assistant",data:{thread_id:null,agent_id:X.role==="assistant"?"gi":null,content_blocks:X.payload?.content_blocks||null,kind:X.payload?.kind||null,source:X.payload?.source||null,clipped:X.payload?.clipped||!1}}))}}async function L6($,j=null){let q=j?.startsWith("gi:")?j.slice(3):null;if(!q)return null;let Q=(await K2(`/api/sessions/${encodeURIComponent(q)}/turns`).catch(()=>({turns:[]}))).turns||[],J=Q.find((K)=>K.status==="running"||K.status==="cancelling")||Q.find((K)=>K.status==="queued");if(!J)return null;return{type:J.status==="running"?"tool_call":"intent",title:J.status==="cancelling"?"Cancelling…":J.status==="queued"?"Queued":J.prompt,status:J.status}}async function p4($=null){let j=await K2("/api/runtime/config").catch(()=>({}));return{models:(j.enabled_models||[]).map((B)=>({id:B,provider:j.default_provider||"",label:B})),current:j.default_model||""}}async function u2($,j,q=null,B=[],Q=null,J=null){let K=J?.startsWith("gi:")?J.slice(3):null;if(!K)throw Error("No active session");let X=Q==="steer"?"steer":Q==="queue"?"queue":"prompt";return K2(`/api/sessions/${encodeURIComponent(K)}/prompt`,{method:"POST",body:JSON.stringify({prompt:j,intent:X})})}async function V6($,j=null){return null}async function h5($){return K2(`/api/media/${$}`).catch(()=>null)}function t2($){return`/api/media/${$}/raw`}function Z6($){return`/api/media/${$}/thumbnail`}async function F6($){return null}async function c4($=null){return K2("/api/workspace/tree")}async function H6($,j=null){return K2(`/api/workspace/file?path=${encodeURIComponent($)}`)}async function Y6($=null){return{status:"ready",indexed_at:null}}async function A6($=null){return null}async function D6($,j,q=null){return K2("/api/workspace/file",{method:"POST",body:JSON.stringify({path:$,content:j})}).catch(()=>null)}async function C6($,j,q=null){return null}async function I6($,j,q=null){return null}async function O6($,j=null){return null}async function p5($,j,q=null){return null}async function d4($,j,q=null){return null}function c5($){return`/api/workspace/file?path=${encodeURIComponent($)}`}async function T6($=null){return null}function M6($,j={}){let q=new URLSearchParams({path:String($||"")});if(j?.download)q.set("download","1");return`/api/workspace/raw?${q.toString()}`}function d5($){return M6($,{download:!0})}async function G6($){}import{classHighlighter as v9,highlightTree as u9,StreamLanguage as O2,cssLanguage as g9,goLanguage as m9,htmlLanguage as h9,javascriptLanguage as p9,jsxLanguage as c9,tsxLanguage as d9,typescriptLanguage as l9,jsonLanguage as r9,markdownLanguage as i9,pythonLanguage as n9,StandardSQL as s9,xmlLanguage as o9,yamlLanguage as a9,dockerFile as t9,powerShell as e9,ruby as $7,rust as j7,shell as q7,swift as B7,toml as Q7}from"#editor-vendor/codemirror";function I2($){return $.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;").replace(/'/g,"&#39;")}var J7=O2.define(q7).parser,K7=O2.define(e9).parser,X7=O2.define(t9).parser,U7=O2.define($7).parser,_7=O2.define(j7).parser,z7=O2.define(B7).parser,W7=O2.define(Q7).parser;function G7($){switch(String($||"").trim().toLowerCase()){case"js":case"javascript":return p9.parser;case"ts":case"typescript":return l9.parser;case"jsx":return c9.parser;case"tsx":return d9.parser;case"py":case"python":return n9.parser;case"json":return r9.parser;case"css":return g9.parser;case"html":return h9.parser;case"xml":return o9.parser;case"yaml":case"yml":return a9.parser;case"md":case"markdown":return i9.parser;case"sql":return s9.language.parser;case"go":return m9.parser;case"sh":case"bash":case"shell":case"zsh":return J7;case"ps1":case"powershell":return K7;case"dockerfile":return X7;case"rb":case"ruby":return U7;case"rs":case"rust":return _7;case"swift":return z7;case"toml":return W7;default:return null}}function S6($,j){let q=G7(j);if(!q)return I2($);let B=[];try{let K=q.parse($);u9(K,v9,(X,z,U)=>{if(!U||X>=z)return;B.push({from:X,to:z,cls:U})})}catch{return I2($)}if(!B.length)return I2($);B.sort((K,X)=>K.from-X.from||K.to-X.to);let Q=0,J="";for(let K of B){if(K.from>Q)J+=I2($.slice(Q,K.from));J+=`<span class="${I2(K.cls)}">${I2($.slice(K.from,K.to))}</span>`,Q=Math.max(Q,K.to)}if(Q<$.length)J+=I2($.slice(Q));return J}var l4=/#(\w+)/g,N7=new Set(["strong","em","b","i","u","s","del","ins","sub","sup","mark","small","br","p","ul","ol","li","blockquote","ruby","rt","rp","span","input"]),L7=new Set(["a","abbr","blockquote","br","code","del","div","em","hr","h1","h2","h3","h4","h5","h6","i","img","input","ins","kbd","li","mark","ol","p","pre","ruby","rt","rp","s","small","span","strong","sub","sup","table","tbody","td","th","thead","tr","u","ul","math","semantics","mrow","mi","mn","mo","mtext","mspace","msup","msub","msubsup","mfrac","msqrt","mroot","mtable","mtr","mtd","annotation"]),V7=new Set(["class","title","role","aria-hidden","aria-label","aria-expanded","aria-live","data-mermaid","data-hashtag"]),y6={a:new Set(["href","target","rel"]),img:new Set(["src","alt","title"]),input:new Set(["type","checked","disabled"])},Z7=new Set(["http:","https:","mailto:",""]);function F7($,j){let q=String($||"").toLowerCase(),B=String(j||"").toLowerCase();if(!B||B.startsWith("on"))return!1;if(B.startsWith("data-")||B.startsWith("aria-"))return!0;return(y6[q]||new Set).has(B)||V7.has(B)}function l5($){return String($||"").replace(/&/g,"&amp;").replace(/"/g,"&quot;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/'/g,"&#39;")}function T2($,j={}){if(!$)return null;let q=String($).trim();if(!q)return null;if(q.startsWith("#")||q.startsWith("/"))return q;if(q.startsWith("data:")){if(j.allowDataImage&&/^data:image\//i.test(q))return q;return null}if(q.startsWith("blob:"))return q;try{let B=new URL(q,typeof window<"u"?window.location.origin:"http://localhost");if(!Z7.has(B.protocol))return null;return B.href}catch{return null}}function k6($,j={}){if(!$)return"";if(j?.sanitize===!1)return $;let q=new DOMParser().parseFromString($,"text/html"),B=[],Q=q.createTreeWalker(q.body,NodeFilter.SHOW_ELEMENT),J;while(J=Q.nextNode())B.push(J);for(let K of B){let X=K.tagName.toLowerCase();if(!L7.has(X)){let U=K.parentNode;if(!U)continue;while(K.firstChild)U.insertBefore(K.firstChild,K);U.removeChild(K);continue}let z=y6[X]||new Set;for(let U of Array.from(K.attributes)){let L=U.name.toLowerCase(),N=U.value;if(L.startsWith("on")){K.removeAttribute(U.name);continue}if(F7(X,L)){if(L==="href"){let V=T2(N);if(!V)K.removeAttribute(U.name);else if(K.setAttribute(U.name,V),X==="a"){if(!K.getAttribute("rel"))K.setAttribute("rel","noopener noreferrer");if(/^https?:\/\//i.test(V))K.setAttribute("target","_blank")}}else if(L==="src"){let V=X==="img"&&typeof j.rewriteImageSrc==="function"?j.rewriteImageSrc(N):N,O=T2(V,{allowDataImage:X==="img"});if(!O)K.removeAttribute(U.name);else K.setAttribute(U.name,O)}continue}K.removeAttribute(U.name)}}return q.body.innerHTML}function R6($){if(!$)return $;let j=$.replace(/</g,"&lt;").replace(/>/g,"&gt;");return new DOMParser().parseFromString(j,"text/html").documentElement.textContent}function e2($,j=2){if(!$)return $;let q=$;for(let B=0;B<j;B+=1){let Q=R6(q);if(Q===q)break;q=Q}return q}function H7($){if(!$)return{text:"",frontmatter:null};let j=$.replace(/^\uFEFF/,"").replace(/\r\n/g,`
`).replace(/\r/g,`
`);if(!j.startsWith(`---
`))return{text:j,frontmatter:null};let q=j.split(`
`),B=-1;for(let K=1;K<q.length;K+=1)if(/^(---|\.\.\.)\s*$/.test(q[K])){B=K;break}if(B<=0)return{text:j,frontmatter:null};let Q=q.slice(1,B).join(`
`);return{text:q.slice(B+1).join(`
`).replace(/^\n+/,""),frontmatter:Q}}function Y7($){let{text:j,frontmatter:q}=H7($);if(q===null)return j;return["<!--frontmatter-block-start-->","```yaml",q,"```","<!--frontmatter-block-end-->",j].filter(Boolean).join(`

`)}function A7($){if(!$)return{text:"",blocks:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=[],Q=[],J=!1,K=[];for(let X of q){if(!J&&X.trim().match(/^```mermaid\s*$/i)){J=!0,K=[];continue}if(J&&X.trim().match(/^```\s*$/)){let z=B.length;B.push(K.join(`
`)),Q.push(`@@MERMAID_BLOCK_${z}@@`),J=!1,K=[];continue}if(J)K.push(X);else Q.push(X)}if(J)Q.push("```mermaid"),Q.push(...K);return{text:Q.join(`
`),blocks:B}}function D7($){if(!$)return $;return e2($,5)}function C7($){let j=new TextEncoder().encode(String($||"")),q="";for(let B of j)q+=String.fromCharCode(B);return btoa(q)}function I7($){let j=atob(String($||"")),q=new Uint8Array(j.length);for(let B=0;B<j.length;B+=1)q[B]=j.charCodeAt(B);return new TextDecoder().decode(q)}function O7($,j){if(!$||!j||j.length===0)return $;return $.replace(/@@MERMAID_BLOCK_(\d+)@@/g,(q,B)=>{let Q=Number(B),J=j[Q]??"",K=D7(J);return`<div class="mermaid-container" data-mermaid="${C7(K)}"><div class="mermaid-loading">Loading diagram...</div></div>`})}function E6($){if(!$)return $;return $.replace(/<code>([\s\S]*?)<\/code>/gi,(j,q)=>{if(q.includes(`
`))return`
\`\`\`
${q}
\`\`\`
`;return`\`${q}\``})}function T7($){if(!$)return $;return $.replace(/<pre><code(?:\s+class="language-([A-Za-z0-9_+-]+)")?>([\s\S]*?)<\/code><\/pre>/g,(q,B,Q)=>{let J=String(B||"").trim().toLowerCase(),K=e2(Q,2),X=J||"plaintext",z=S6(K,J);return`<pre><code class="hljs language-${l5(X)}">${z}</code></pre>`}).replace(/<!--frontmatter-block-start-->\s*<pre>/g,'<pre class="frontmatter-block">').replace(/<\/pre>\s*<!--frontmatter-block-end-->/g,"</pre>")}var M7={span:new Set(["title","class","lang","dir"]),input:new Set(["type","checked","disabled"])};function S7($,j){let q=M7[$];if(!q||!j)return"";let B=[],Q=/([a-zA-Z_:][\w:.-]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'`=<>]+)))?/g,J;while(J=Q.exec(j)){let K=(J[1]||"").toLowerCase();if(!K||K.startsWith("on")||!q.has(K))continue;let X=J[2]??J[3]??J[4]??"";B.push(` ${K}="${l5(X)}"`)}return B.join("")}function P6($){if(!$)return $;return $.replace(/&lt;((?:[^"'<>]|"[^"]*"|'[^']*')*?)(?:&gt;|>)/g,(j,q)=>{let B=q.trim(),Q=B.startsWith("/"),J=Q?B.slice(1).trim():B,X=J.endsWith("/")?J.slice(0,-1).trim():J,[z=""]=X.split(/\s+/,1),U=z.toLowerCase();if(!U||!N7.has(U))return j;if(U==="br")return Q?"":"<br>";if(Q)return`</${U}>`;let L=X.slice(z.length).trim(),N=S7(U,L);return`<${U}${N}>`})}function f6($){if(!$)return $;let j=(q)=>q.replace(/&amp;lt;/g,"&lt;").replace(/&amp;gt;/g,"&gt;").replace(/&amp;quot;/g,"&quot;").replace(/&amp;#39;/g,"&#39;").replace(/&amp;amp;/g,"&amp;");return $.replace(/<pre><code>([\s\S]*?)<\/code><\/pre>/g,(q,B)=>`<pre><code>${j(B)}</code></pre>`).replace(/<code>([\s\S]*?)<\/code>/g,(q,B)=>`<code>${j(B)}</code>`)}function w6($){if(!$)return $;let j=new DOMParser().parseFromString($,"text/html"),q=j.createTreeWalker(j.body,NodeFilter.SHOW_TEXT),B=(J)=>J.replace(/&lt;/g,"<").replace(/&gt;/g,">").replace(/&quot;/g,'"').replace(/&#39;/g,"'").replace(/&amp;/g,"&"),Q;while(Q=q.nextNode()){if(!Q.nodeValue)continue;let J=B(Q.nodeValue);if(J!==Q.nodeValue)Q.nodeValue=J}return j.body.innerHTML}function y7($){if(!window.katex)return $;let j=(K)=>R6(K).replace(/&gt;/g,">").replace(/&lt;/g,"<").replace(/&amp;/g,"&").replace(/<br\s*\/?\s*>/gi,`
`),q=(K)=>{let X=[],z=K.replace(/<pre\b[^>]*>\s*<code\b[^>]*>[\s\S]*?<\/code>\s*<\/pre>/gi,(U)=>{let L=X.length;return X.push(U),`@@CODE_BLOCK_${L}@@`});return z=z.replace(/<code\b[^>]*>[\s\S]*?<\/code>/gi,(U)=>{let L=X.length;return X.push(U),`@@CODE_INLINE_${L}@@`}),{html:z,blocks:X}},B=(K,X)=>{if(!X.length)return K;return K.replace(/@@CODE_(?:BLOCK|INLINE)_(\d+)@@/g,(z,U)=>{let L=Number(U);return X[L]??""})},Q=q($),J=Q.html;return J=J.replace(/(^|\n|<br\s*\/?\s*>|<p>|<\/p>)\s*\$\$([\s\S]+?)\$\$\s*(?=\n|<br\s*\/?\s*>|<\/p>|$)/gi,(K,X,z)=>{try{let U=katex.renderToString(j(z.trim()),{displayMode:!0,throwOnError:!1});return`${X}${U}`}catch(U){return`<span class="math-error" title="${l5(U.message)}">${K}</span>`}}),B(J,Q.blocks)}function k7($){if(!$)return $;let j=new DOMParser().parseFromString($,"text/html"),q=j.createTreeWalker(j.body,NodeFilter.SHOW_TEXT),B=[],Q;while(Q=q.nextNode())B.push(Q);for(let J of B){let K=J.nodeValue;if(!K)continue;if(l4.lastIndex=0,!l4.test(K))continue;l4.lastIndex=0;let X=J.parentElement;if(X&&(X.closest("a")||X.closest("code")||X.closest("pre")))continue;let z=K.split(l4);if(z.length<=1)continue;let U=j.createDocumentFragment();z.forEach((L,N)=>{if(N%2===1){let V=j.createElement("a");V.setAttribute("href","#"),V.className="hashtag",V.setAttribute("data-hashtag",L),V.textContent=`#${L}`,U.appendChild(V)}else U.appendChild(j.createTextNode(L))}),J.parentNode?.replaceChild(U,J)}return j.body.innerHTML}function R7($){if(!$)return $;let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=[],Q=!1;for(let J of q){if(!Q&&J.trim().match(/^```(?:math|katex|latex)\s*$/i)){Q=!0,B.push("$$");continue}if(Q&&J.trim().match(/^```\s*$/)){Q=!1,B.push("$$");continue}B.push(J)}return B.join(`
`)}function E7($){let j=Y7($||""),q=R7(j),{text:B,blocks:Q}=A7(q),J=e2(B,2),X=E6(J).replace(/</g,"&lt;");return{safeHtml:P6(X),mermaidBlocks:Q}}function M2($,j,q={}){if(!$)return"";let{safeHtml:B,mermaidBlocks:Q}=E7($),J=window.marked?marked.parse(B,{headerIds:!1,mangle:!1}):B.replace(/\n/g,"<br>");return J=f6(J),J=w6(J),J=T7(J),J=y7(J),J=k7(J),J=O7(J,Q),J=k6(J,q),J}function r5($){if(!$)return"";let j=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`),q=e2(j,2),Q=E6(q).replace(/</g,"&lt;").replace(/>/g,"&gt;"),J=P6(Q),K=window.marked?marked.parse(J):J.replace(/\n/g,"<br>");return K=f6(K),K=w6(K),K=k6(K),K}function P7($,j=6){return $.replace(/<polyline\b([^>]*)\bpoints="([^"]+)"([^>]*)\/?\s*>/g,(q,B,Q,J)=>{let K=Q.trim().split(/\s+/).map((z)=>{let[U,L]=z.split(",").map(Number);return{x:U,y:L}});if(K.length<3)return`<polyline${B}points="${Q}"${J}/>`;let X=[`M ${K[0].x},${K[0].y}`];for(let z=1;z<K.length-1;z++){let U=K[z-1],L=K[z],N=K[z+1],V=L.x-U.x,O=L.y-U.y,g=N.x-L.x,h=N.y-L.y,p=Math.sqrt(V*V+O*O),d=Math.sqrt(g*g+h*h),P=Math.min(j,p/2,d/2);if(P<0.5){X.push(`L ${L.x},${L.y}`);continue}let t=L.x-V/p*P,v=L.y-O/p*P,e=L.x+g/d*P,W1=L.y+h/d*P,N1=V*h-O*g>0?1:0;X.push(`L ${t},${v}`),X.push(`A ${P},${P} 0 0 ${N1} ${e},${W1}`)}return X.push(`L ${K[K.length-1].x},${K[K.length-1].y}`),`<path${B}d="${X.join(" ")}"${J}/>`})}async function x6($){if(!window.beautifulMermaid)return;let{renderMermaid:j,THEMES:q}=window.beautifulMermaid,Q=W6()==="dark"?q["tokyo-night"]:q["github-light"],J=$.querySelectorAll(".mermaid-container[data-mermaid]");for(let K of J)try{let X=K.dataset.mermaid,z=I7(X||""),U=e2(z,2),L=await j(U,{...Q,transparent:!0});L=P7(L),K.innerHTML=L,K.removeAttribute("data-mermaid")}catch(X){console.error("Mermaid render error:",X);let z=document.createElement("pre");z.className="mermaid-error",z.textContent=`Diagram error: ${X.message}`,K.innerHTML="",K.appendChild(z),K.removeAttribute("data-mermaid")}}function b6($){let j=new Date($);if(Number.isNaN(j.getTime()))return $;let B=new Date-j,Q=B/1000,J=86400000;if(B<J){if(Q<60)return"just now";if(Q<3600)return`${Math.floor(Q/60)}m`;return`${Math.floor(Q/3600)}h`}if(B<5*J){let z=j.toLocaleDateString(void 0,{weekday:"short"}),U=j.toLocaleTimeString(void 0,{hour:"2-digit",minute:"2-digit"});return`${z} ${U}`}let K=j.toLocaleDateString(void 0,{month:"short",day:"numeric"}),X=j.toLocaleTimeString(void 0,{hour:"2-digit",minute:"2-digit"});return`${K} ${X}`}function $4($){if(!Number.isFinite($))return"0";return Math.round($).toLocaleString()}function X2($){if($<1024)return $+" B";if($<1048576)return($/1024).toFixed(1)+" KB";return($/1048576).toFixed(1)+" MB"}function r4($){let j=new Date($);if(Number.isNaN(j.getTime()))return $;return j.toLocaleString()}function j4($){if($==null)return"";if(typeof $==="string")return $.trim();if(typeof $==="number")return String($);if(typeof $==="boolean")return $?"yes":"no";if(Array.isArray($))return $.map((j)=>j4(j)).filter(Boolean).join(", ");if(typeof $==="object")return Object.entries($).filter(([j])=>!j.startsWith("__")).map(([j,q])=>`${j}: ${j4(q)}`).filter((j)=>!j.endsWith(": ")).join(", ");return String($).trim()}function v6($){if(typeof $!=="object"||$==null||Array.isArray($))return[];return Object.entries($).filter(([j])=>!j.startsWith("__")).map(([j,q])=>({key:j,value:j4(q)})).filter((j)=>j.value)}function f7($){if(!$||typeof $!=="object")return!1;let j=$;return j.type==="adaptive_card_submission"&&typeof j.card_id==="string"&&typeof j.source_post_id==="number"&&typeof j.submitted_at==="string"}function i5($){if(!Array.isArray($))return[];return $.filter(f7)}function i4($){let j=String($.title||$.card_id||"card").trim()||"card",q=$.data;if(q==null)return`Card submission: ${j}`;if(typeof q==="string"||typeof q==="number"||typeof q==="boolean"){let B=j4(q);return B?`Card submission: ${j} — ${B}`:`Card submission: ${j}`}if(typeof q==="object"){let Q=v6(q).map(({key:J,value:K})=>`${J}: ${K}`);return Q.length>0?`Card submission: ${j} — ${Q.join(", ")}`:`Card submission: ${j}`}return`Card submission: ${j}`}function u6($){let j=String($.title||$.card_id||"Card submission").trim()||"Card submission",q=v6($.data),B=q.length>0?q.slice(0,2).map(({key:J,value:K})=>`${J}: ${K}`).join(", "):j4($.data)||null,Q=q.length;return{title:j,summary:B,fields:q,fieldCount:Q,submittedAt:$.submitted_at}}function u1($){return typeof $==="string"?$.trim():""}function g6($){return $.map((j)=>String(j||"").trim()).filter(Boolean).join(`

`).replace(/\n{3,}/g,`

`).trim()}function w7($,j){let q=[],B=[],Q=[];if($.forEach((J,K)=>{if(!J||typeof J!=="object")return;let X=u1(J.type);if(X==="text"){let z=u1(J.text)||u1(J.content);if(z)q.push(z);return}if(X==="resource_link"){let z=u1(J.uri),U=u1(J.title)||u1(J.name)||z;if(z&&U)q.push(U===z?z:`[${U}](${z})`);return}if(X==="resource"){let z=u1(J.title)||u1(J.name)||u1(J.uri)||"Embedded resource",U=u1(J.text);if(U)q.push(`### ${z}

\`\`\`
${U}
\`\`\``);else q.push(`### ${z}`);return}if(X==="generated_widget"){let z=u1(J.title)||u1(J.name)||"Generated widget",U=u1(J.description)||u1(J.subtitle);q.push(g6([`### ${z}`,U]));return}if(X==="adaptive_card"&&u1(J.fallback_text)){q.push(u1(J.fallback_text));return}if(X==="adaptive_card_submission"){let z=i4(J);if(u1(z))q.push(u1(z));return}if(X==="file"){let z=u1(J.name)||u1(J.filename)||u1(J.title)||`attachment:${j[K]??K+1}`;B.push(`- ${z}`);return}if(X==="image"||!X){let z=u1(J.name)||u1(J.filename)||u1(J.title)||`attachment:${j[K]??K+1}`;Q.push(`- ${z}`)}}),Q.length>0)q.push(`Images:
${Q.join(`
`)}`);if(B.length>0)q.push(`Attachments:
${B.join(`
`)}`);return g6(q)}function m6($){let j=$?.data||{},q=typeof j.content==="string"?j.content.replace(/\r\n/g,`
`).replace(/\r/g,`
`).trimEnd():"";if(q.trim())return q;let B=Array.isArray(j.content_blocks)?j.content_blocks:[],Q=Array.isArray(j.media_ids)?j.media_ids:[];return w7(B,Q)}var h6="PiClaw";function n5($,j,q=!1){let B=$||"PiClaw",Q=B.charAt(0).toUpperCase(),J=["#FF6B6B","#4ECDC4","#45B7D1","#FFA07A","#98D8C8","#F7DC6F","#BB8FCE","#85C1E2","#F8B195","#6C5CE7","#00B894","#FDCB6E","#E17055","#74B9FF","#A29BFE","#FD79A8","#00CEC9","#FFEAA7","#DFE6E9","#FF7675","#55EFC4","#81ECEC","#FAB1A0","#74B9FF","#A29BFE","#FD79A8"],K=Q.charCodeAt(0)%J.length,X=J[K],z=B.trim().toLowerCase(),U=typeof j==="string"?j.trim():"",L=U?U:null,N=q||z==="PiClaw".toLowerCase()||z==="pi";return{letter:Q,color:X,image:L||(N?"/static/icon-192.png":null)}}function p6($,j){if(!$)return"PiClaw";let q=j[$]?.name||$;return q?q.charAt(0).toUpperCase()+q.slice(1):"PiClaw"}function c6($,j){if(!$)return null;let q=j[$]||{};return q.avatar_url||q.avatarUrl||q.avatar||null}function d6($){if(!$)return null;if(typeof document<"u"){let J=document.documentElement,K=J?.dataset?.colorTheme||"",X=J?.dataset?.tint||"",z=getComputedStyle(J).getPropertyValue("--accent-color")?.trim();if(z&&(X||K&&K!=="default"))return z}let j=["#4ECDC4","#FF6B6B","#45B7D1","#BB8FCE","#FDCB6E","#00B894","#74B9FF","#FD79A8","#81ECEC","#FFA07A"],q=String($),B=0;for(let J=0;J<q.length;J+=1)B=(B*31+q.charCodeAt(J))%2147483647;let Q=Math.abs(B)%j.length;return j[Q]}var x7=new Set(["application/json","application/xml","text/csv","text/html","text/markdown","text/plain","text/xml"]);var b7=new Set(["application/msword","application/rtf","application/vnd.ms-excel","application/vnd.ms-powerpoint","application/vnd.oasis.opendocument.presentation","application/vnd.oasis.opendocument.spreadsheet","application/vnd.oasis.opendocument.text","application/vnd.openxmlformats-officedocument.presentationml.presentation","application/vnd.openxmlformats-officedocument.spreadsheetml.sheet","application/vnd.openxmlformats-officedocument.wordprocessingml.document"]),v7=new Set(["application/vnd.jgraph.mxfile"]);function S2($){return typeof $==="string"?$.trim().toLowerCase():""}function u7($){let j=S2($);return!!j&&(j.endsWith(".drawio")||j.endsWith(".drawio.xml")||j.endsWith(".drawio.svg")||j.endsWith(".drawio.png"))}function g7($){let j=S2($);return!!j&&j.endsWith(".pdf")}function m7($){let j=S2($);return!!j&&(j.endsWith(".docx")||j.endsWith(".doc")||j.endsWith(".odt")||j.endsWith(".rtf")||j.endsWith(".xlsx")||j.endsWith(".xls")||j.endsWith(".ods")||j.endsWith(".pptx")||j.endsWith(".ppt")||j.endsWith(".odp"))}var h7=new Set(["application/zip","application/x-zip-compressed"]);function p7($){let j=S2($);return!!j&&j.endsWith(".zip")}function c7($){let j=S2($);return!!j&&(j.endsWith(".html")||j.endsWith(".htm"))}function d7($){let j=S2($);if(!j)return!1;return j.endsWith(".sh")||j.endsWith(".bash")||j.endsWith(".zsh")||j.endsWith(".sb")}function s5($,j){let q=S2($);if(u7(j)||v7.has(q))return"drawio";if(g7(j)||q==="application/pdf")return"pdf";if(m7(j)||b7.has(q))return"office";if(p7(j)||h7.has(q))return"archive";if(c7(j)||q==="text/html")return"html";if(d7(j))return"text";if(!q)return"unsupported";if(q.startsWith("video/"))return"video";if(q.startsWith("image/"))return"image";if(x7.has(q)||q.startsWith("text/"))return"text";return"unsupported"}function l6($,j,q){try{return $.setAttribute(j,q),!0}catch(B){return!1}}function r6($,j){try{return $[j]=!0,!0}catch(q){return!1}}function i6($){$.classList.add("adaptive-card-readonly");for(let j of Array.from($.querySelectorAll("input, textarea, select, button"))){let q=j;if(l6(q,"aria-disabled","true"),l6(q,"tabindex","-1"),"disabled"in q)r6(q,"disabled");if("readOnly"in q)r6(q,"readOnly")}}function l7($){let q=String($||"").trim().match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);if(!q)return null;let B=q[1].length===3?q[1].split("").map((Q)=>`${Q}${Q}`).join(""):q[1];return{r:parseInt(B.slice(0,2),16),g:parseInt(B.slice(2,4),16),b:parseInt(B.slice(4,6),16)}}function r7($){let q=String($||"").trim().match(/^rgba?\((\d+)[,\s]+(\d+)[,\s]+(\d+)/i);if(!q)return null;let B=Number(q[1]),Q=Number(q[2]),J=Number(q[3]);if(![B,Q,J].every((K)=>Number.isFinite(K)))return null;return{r:B,g:Q,b:J}}function n6($){return l7($)||r7($)}function n4($){let j=(J)=>{let K=J/255;return K<=0.03928?K/12.92:((K+0.055)/1.055)**2.4},q=j($.r),B=j($.g),Q=j($.b);return 0.2126*q+0.7152*B+0.0722*Q}function i7($,j){let q=Math.max(n4($),n4(j)),B=Math.min(n4($),n4(j));return(q+0.05)/(B+0.05)}function n7($,j,q="#ffffff"){let B=n6($);if(!B)return q;let Q=q,J=-1;for(let K of j){let X=n6(K);if(!X)continue;let z=i7(B,X);if(z>J)Q=K,J=z}return Q}function o5(){let $=getComputedStyle(document.documentElement),j=(g,h)=>{for(let p of g){let d=$.getPropertyValue(p).trim();if(d)return d}return h},q=j(["--text-primary","--color-text"],"#0f1419"),B=j(["--text-secondary","--color-text-muted"],"#536471"),Q=j(["--bg-primary","--color-bg-primary"],"#ffffff"),J=j(["--bg-secondary","--color-bg-secondary"],"#f7f9fa"),K=j(["--bg-hover","--bg-tertiary","--color-bg-tertiary"],"#e8ebed"),X=j(["--accent-color","--color-accent"],"#1d9bf0"),z=j(["--success-color","--color-success"],"#00ba7c"),U=j(["--warning-color","--color-warning","--accent-color"],"#f0b429"),L=j(["--danger-color","--color-error"],"#f4212e"),N=j(["--border-color","--color-border"],"#eff3f4"),V=j(["--font-family"],"system-ui, sans-serif"),O=n7(X,[q,Q],q);return{fg:q,fgMuted:B,bgPrimary:Q,bg:J,bgEmphasis:K,accent:X,good:z,warning:U,attention:L,border:N,fontFamily:V,buttonTextColor:O}}function s6(){let{fg:$,fgMuted:j,bg:q,bgEmphasis:B,accent:Q,good:J,warning:K,attention:X,border:z,fontFamily:U}=o5();return{fontFamily:U,containerStyles:{default:{backgroundColor:q,foregroundColors:{default:{default:$,subtle:j},accent:{default:Q,subtle:Q},good:{default:J,subtle:J},warning:{default:K,subtle:K},attention:{default:X,subtle:X}}},emphasis:{backgroundColor:B,foregroundColors:{default:{default:$,subtle:j},accent:{default:Q,subtle:Q},good:{default:J,subtle:J},warning:{default:K,subtle:K},attention:{default:X,subtle:X}}}},actions:{actionsOrientation:"horizontal",actionAlignment:"left",buttonSpacing:8,maxActions:5,showCard:{actionMode:"inline"},spacing:"default"},adaptiveCard:{allowCustomStyle:!1},spacing:{small:4,default:8,medium:12,large:16,extraLarge:24,padding:12},separator:{lineThickness:1,lineColor:z},fontSizes:{small:12,default:14,medium:16,large:18,extraLarge:22},fontWeights:{lighter:300,default:400,bolder:600},imageSizes:{small:40,medium:80,large:120},textBlock:{headingLevel:2}}}var s7=new Set(["1.0","1.1","1.2","1.3","1.4","1.5","1.6"]),o6=!1,s4=null,a6=!1;function a5($){$.querySelector(".adaptive-card-notice")?.remove()}function o7($,j,q="error"){a5($);let B=document.createElement("div");B.className=`adaptive-card-notice adaptive-card-notice-${q}`,B.textContent=j,$.appendChild(B)}function a7($,j=(q)=>M2(q,null)){let q=typeof $==="string"?$:String($??"");if(!q.trim())return{outputHtml:"",didProcess:!1};return{outputHtml:j(q),didProcess:!0}}function t7($=(j)=>M2(j,null)){return(j,q)=>{try{let B=a7(j,$);q.outputHtml=B.outputHtml,q.didProcess=B.didProcess}catch(B){console.error("[adaptive-card] Failed to process markdown:",B),q.outputHtml=String(j??""),q.didProcess=!1}}}function e7($){if(a6||!$?.AdaptiveCard)return;$.AdaptiveCard.onProcessMarkdown=t7(),a6=!0}async function $$(){if(o6)return;if(s4)return s4;return s4=new Promise(($,j)=>{let q=document.createElement("script");q.src="/static/js/vendor/adaptivecards.min.js",q.onload=()=>{o6=!0,$()},q.onerror=()=>j(Error("Failed to load adaptivecards SDK")),document.head.appendChild(q)}),s4}function j$(){return globalThis.AdaptiveCards}function q$($){if(!$||typeof $!=="object")return!1;let j=$;return j.type==="adaptive_card"&&typeof j.card_id==="string"&&typeof j.schema_version==="string"&&typeof j.payload==="object"&&j.payload!==null}function B$($){return s7.has($)}function e5($){if(!Array.isArray($))return[];return $.filter(q$)}function Q$($){let j=(typeof $?.getJsonTypeName==="function"?$.getJsonTypeName():"")||$?.constructor?.name||"Unknown",q=(typeof $?.title==="string"?$.title:"")||"",B=(typeof $?.url==="string"?$.url:"")||void 0,Q=$?.data??void 0;return{type:j,title:q,data:Q,url:B,raw:$}}function t5($){if($==null)return"";if(typeof $==="string")return $.trim();if(typeof $==="number")return String($);if(typeof $==="boolean")return $?"yes":"no";if(Array.isArray($))return $.map((j)=>t5(j)).filter(Boolean).join(", ");if(typeof $==="object")return Object.entries($).map(([q,B])=>`${q}: ${t5(B)}`).filter((q)=>!q.endsWith(": ")).join(", ");return String($).trim()}function J$($,j,q){if(j==null)return j;if($==="Input.Toggle"){if(typeof j==="boolean"){if(j)return q?.valueOn??"true";return q?.valueOff??"false"}return typeof j==="string"?j:String(j)}if($==="Input.ChoiceSet"){if(Array.isArray(j))return j.join(",");return typeof j==="string"?j:String(j)}if(Array.isArray(j))return j.join(", ");if(typeof j==="object")return t5(j);return typeof j==="string"?j:String(j)}function K$($,j){if(!$||typeof $!=="object")return $;if(!j||typeof j!=="object"||Array.isArray(j))return $;let q=j,B=(Q)=>{if(Array.isArray(Q))return Q.map((X)=>B(X));if(!Q||typeof Q!=="object")return Q;let K={...Q};if(typeof K.id==="string"&&K.id in q&&String(K.type||"").startsWith("Input."))K.value=J$(K.type,q[K.id],K);for(let[X,z]of Object.entries(K))if(Array.isArray(z)||z&&typeof z==="object")K[X]=B(z);return K};return B($)}function X$($){if(typeof $!=="string"||!$.trim())return"";let j=new Date($);if(Number.isNaN(j.getTime()))return"";return new Intl.DateTimeFormat(void 0,{month:"short",day:"numeric",hour:"numeric",minute:"2-digit"}).format(j)}function U$($){if($.state==="active")return null;let j=$.state==="completed"?"Submitted":$.state==="cancelled"?"Cancelled":"Failed",q=$.last_submission&&typeof $.last_submission==="object"?$.last_submission:null,B=q&&typeof q.title==="string"?q.title.trim():"",Q=X$($.completed_at||q?.submitted_at),J=[B||null,Q||null].filter(Boolean).join(" · ")||null;return{label:j,detail:J}}async function t6($,j,q){if(!B$(j.schema_version))return console.warn(`[adaptive-card] Unsupported schema version ${j.schema_version} for card ${j.card_id}`),!1;try{await $$()}catch(B){return console.error("[adaptive-card] Failed to load SDK:",B),!1}try{let B=j$();e7(B);let Q=new B.AdaptiveCard,J=o5();Q.hostConfig=new B.HostConfig(s6());let K=j.last_submission&&typeof j.last_submission==="object"?j.last_submission.data:void 0,X=j.state==="active"?j.payload:K$(j.payload,K);Q.parse(X),Q.onExecuteAction=(L)=>{let N=Q$(L);if(q?.onAction)a5($),$.classList.add("adaptive-card-busy"),Promise.resolve(q.onAction(N)).catch((V)=>{console.error("[adaptive-card] Action failed:",V);let O=V instanceof Error?V.message:String(V||"Action failed.");o7($,O||"Action failed.","error")}).finally(()=>{$.classList.remove("adaptive-card-busy")});else console.log("[adaptive-card] Action executed (not wired yet):",N)};let z=Q.render();if(!z)return console.warn(`[adaptive-card] Card ${j.card_id} rendered to null`),!1;$.classList.add("adaptive-card-container"),$.style.setProperty("--adaptive-card-button-text-color",J.buttonTextColor);let U=U$(j);if(U){$.classList.add("adaptive-card-finished");let L=document.createElement("div");L.className=`adaptive-card-status adaptive-card-status-${j.state}`;let N=document.createElement("span");if(N.className="adaptive-card-status-label",N.textContent=U.label,L.appendChild(N),U.detail){let V=document.createElement("span");V.className="adaptive-card-status-detail",V.textContent=U.detail,L.appendChild(V)}$.appendChild(L)}if(a5($),$.appendChild(z),U)i6(z);return!0}catch(B){return console.error(`[adaptive-card] Failed to render card ${j.card_id}:`,B),!1}}function _$($){let j=$?.artifact||{},q=j.kind||$?.kind||null;if(q!=="html"&&q!=="svg"&&q!=="session_tree")return null;if(q==="html"){let Q=typeof j.html==="string"?j.html:typeof $?.html==="string"?$.html:"";return Q?{kind:q,html:Q}:null}if(q==="svg"){let Q=typeof j.svg==="string"?j.svg:typeof $?.svg==="string"?$.svg:"";return Q?{kind:q,svg:Q}:null}let B=j.tree&&typeof j.tree==="object"?j.tree:$?.tree&&typeof $.tree==="object"?$.tree:null;return{kind:q,tree:B}}function z$($,j=!1){let B=(Array.isArray($)?$:j?["interactive"]:[]).filter((Q)=>typeof Q==="string").map((Q)=>Q.trim().toLowerCase()).filter(Boolean);return Array.from(new Set(B))}function $3($,j){if(!$||$.type!=="generated_widget")return null;let q=_$($);if(!q)return null;return{title:$.title||$.name||"Generated widget",subtitle:typeof $.subtitle==="string"?$.subtitle:"",description:$.description||$.subtitle||"",originPostId:Number.isFinite(j?.id)?j.id:null,originChatJid:typeof j?.chat_jid==="string"?j.chat_jid:null,widgetId:$.widget_id||$.id||null,artifact:q,capabilities:z$($.capabilities,$.interactive===!0),source:"timeline",status:"final"}}function e6($){return $3($,null)!==null}function j3({children:$,className:j=""}){let[q,B]=M(null);return w(()=>{if(typeof document>"u")return;let Q=document.createElement("div");if(j)Q.className=j;return document.body.appendChild(Q),B(Q),()=>{try{b2(null,Q)}finally{Q.remove(),B((J)=>J===Q?null:J)}}},[j]),w5(()=>{if(!q)return;b2($,q);return},[$,q]),null}function $8({src:$,onClose:j}){return w(()=>{let q=(B)=>{if(B.key==="Escape")j()};return document.addEventListener("keydown",q),()=>document.removeEventListener("keydown",q)},[j]),G`
        <${j3} className="image-modal-portal-root">
            <div class="image-modal" onClick=${j}>
                <img src=${$} alt="Full size" />
            </div>
        </${j3}>
    `}function P0({prefix:$="file",label:j,title:q,onRemove:B,onClick:Q,removeTitle:J="Remove",icon:K="file"}){let X=`${$}-file-pill`,z=`${$}-file-name`,U=`${$}-file-remove`,L=K==="message"?G`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>`:G`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
      </svg>`;return G`
    <span class=${X} title=${q||j} onClick=${Q}>
      ${L}
      <span class=${z}>${j}</span>
      ${B&&G`
        <button
          class=${U}
          onClick=${(N)=>{N.preventDefault(),N.stopPropagation(),B()}}
          title=${J}
          type="button"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      `}
    </span>
  `}async function q8($,j){try{return await $?.writeText?.(j),!0}catch(q){return!1}}function q3($,j){let q=typeof j?.text==="string"?j.text:"",B=typeof j?.html==="string"?j.html:"";if(!$||!q||typeof $.createElement!=="function"||typeof $.execCommand!=="function")return!1;let Q=null,J=!1,K=(X)=>{let z=X?.clipboardData;if(!z||typeof z.setData!=="function")return;if(z.setData("text/plain",q),B)z.setData("text/html",B);if(typeof X.preventDefault==="function")X.preventDefault();J=!0};try{if(Q=$.createElement("textarea"),Q.value=q,typeof Q.setAttribute==="function")Q.setAttribute("readonly","");if(Q.style)Q.style.position="fixed",Q.style.opacity="0",Q.style.pointerEvents="none";if($.body?.appendChild?.(Q),typeof Q.select==="function")Q.select();if(typeof Q.setSelectionRange==="function")Q.setSelectionRange(0,Q.value.length);$.addEventListener?.("copy",K,!0);let X=$.execCommand("copy");return Boolean(J||X)}catch{return!1}finally{if($.removeEventListener?.("copy",K,!0),Q)$.body?.removeChild?.(Q)}}function j8($){if(!$||typeof $!=="object")return null;let j=$;if(typeof j.nodeType==="number"&&j.nodeType===3)return j.parentNode||null;return j}function B8($,j){let q=$?.clipboardData,B=j?.root,Q=j?.selection;if(!q||typeof q.setData!=="function"||!B||!Q)return!1;if(Q.isCollapsed)return!1;let J=!1;if(Number(Q.rangeCount||0)>0&&typeof Q.getRangeAt==="function")try{let z=Q.getRangeAt(0);if(z&&typeof z.intersectsNode==="function")J=Boolean(z.intersectsNode(B))}catch{J=!1}if(!J&&typeof B.contains==="function"){let z=j8(Q.anchorNode),U=j8(Q.focusNode);J=Boolean(z&&B.contains(z)||U&&B.contains(U))}if(!J)return!1;let X=typeof Q.toString==="function"?String(Q.toString()||"").replace(/\u00a0/g," "):"";if(!X)return!1;return q.setData("text/plain",X),$?.preventDefault?.(),!0}function Q8($,j){try{return Boolean($?.getItem?.(j))}catch(q){return!1}}function J8($,j,q){try{return $?.setItem?.(j,q),!0}catch(B){return!1}}function K8($,j){let q=typeof $==="string"&&$.trim()?$.trim():null;if(q)return q;if(!j)return null;try{return new URL(j).hostname}catch(B){return j}}function W$({mediaId:$,onPreview:j}){let[q,B]=M(null);if(w(()=>{h5($).then(B).catch((U)=>{console.warn("[post] Failed to load attachment metadata for file card:",$,U)})},[$]),!q)return null;let Q=q.filename||"file",J=q.metadata?.size,K=J?X2(J):"",z=s5(q.content_type,q.filename)==="unsupported"?"Details":"Preview";return G`
        <div class="file-attachment" onClick=${(U)=>U.stopPropagation()}>
            <a href=${t2($)} download=${Q} class="file-attachment-main">
                <svg class="file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                    <polyline points="10 9 9 9 8 9"/>
                </svg>
                <div class="file-info">
                    <span class="file-name">${Q}</span>
                    <span class="file-meta-row">
                        ${K&&G`<span class="file-size">${K}</span>`}
                        ${q.content_type&&G`<span class="file-size">${q.content_type}</span>`}
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
                onClick=${(U)=>{U.preventDefault(),U.stopPropagation(),j?.({mediaId:$,info:q})}}
            >
                ${z}
            </button>
        </div>
    `}function G$($){if(!Array.isArray($))return[];return $.filter((j)=>j&&typeof j==="object"&&j.type==="recovery_marker"&&j.recovered)}function N$($){if(!Array.isArray($))return[];return $.filter((j)=>j&&typeof j==="object"&&j.type==="timeout_marker"&&(j.timed_out??!0))}var L$={context_recover:"context limit exceeded",rate_limit:"rate limit hit",api_error:"API error",timeout:"request timeout",overloaded:"service overloaded",connection:"connection error"};function V$($){let j=Number($?.attempts_used||0),q=String($?.classifier||"").trim(),B=L$[q]||(q?q.replace(/_/g," "):""),Q=["Recovered automatically"];if(j>1)Q[0]=`Recovered after ${j} attempts`;if(B)Q.push(B);return Q.join(" — ")}function Z$($){let j=typeof $?.tool_action_summary==="string"?$.tool_action_summary.trim():"";return j?`Turn timed out — ${j}`:"Turn timed out before the model finished responding"}function F$({attachment:$,onPreview:j}){let q=Number($?.id),[B,Q]=M(null);w(()=>{if(!Number.isFinite(q))return;h5(q).then(Q).catch((U)=>{console.warn("[post] Failed to load attachment metadata for attachment pill:",q,U)});return},[q]);let J=B?.filename||$.label||`attachment-${$.id}`,K=Number.isFinite(q)?t2(q):null,z=s5(B?.content_type,B?.filename||$?.label)==="unsupported"?"Details":"Preview";return G`
        <span class="attachment-pill" title=${J}>
            ${K?G`
                    <a href=${K} download=${J} class="attachment-pill-main" onClick=${(U)=>U.stopPropagation()}>
                        <${P0}
                            prefix="post"
                            label=${$.label}
                            title=${J}
                        />
                    </a>
                `:G`
                    <${P0}
                        prefix="post"
                        label=${$.label}
                        title=${J}
                    />
                `}
            ${Number.isFinite(q)&&B&&G`
                <button
                    class="attachment-pill-preview"
                    type="button"
                    title=${z}
                    onClick=${(U)=>{U.preventDefault(),U.stopPropagation(),j?.({mediaId:q,info:B})}}
                >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z"/>
                        <circle cx="12" cy="12" r="3"/>
                    </svg>
                </button>
            `}
        </span>
    `}function o4({annotations:$}){if(!$)return null;let{audience:j,priority:q,lastModified:B}=$,Q=B?r4(B):null;return G`
        <div class="content-annotations">
            ${j&&j.length>0&&G`
                <span class="content-annotation">Audience: ${j.join(", ")}</span>
            `}
            ${typeof q==="number"&&G`
                <span class="content-annotation">Priority: ${q}</span>
            `}
            ${Q&&G`
                <span class="content-annotation">Updated: ${Q}</span>
            `}
        </div>
    `}function H$({block:$}){let j=$.title||$.name||$.uri,q=$.description,B=$.size?X2($.size):"",Q=$.mime_type||"",J=D$(Q),K=T2($.uri);return G`
        <a
            href=${K||"#"}
            class="resource-link"
            target=${K?"_blank":void 0}
            rel=${K?"noopener noreferrer":void 0}
            onClick=${(X)=>X.stopPropagation()}>
            <div class="resource-link-main">
                <div class="resource-link-header">
                    <span class="resource-link-icon-inline">${J}</span>
                    <div class="resource-link-title">${j}</div>
                </div>
                ${q&&G`<div class="resource-link-description">${q}</div>`}
                <div class="resource-link-meta">
                    ${Q&&G`<span>${Q}</span>`}
                    ${B&&G`<span>${B}</span>`}
                </div>
            </div>
            <div class="resource-link-icon">↗</div>
        </a>
    `}function Y$({block:$}){let[j,q]=M(!1),B=$.uri||"Embedded resource",Q=$.text||"",J=Boolean($.data),K=$.mime_type||"";return G`
        <div class="resource-embed">
            <button class="resource-embed-toggle" onClick=${(X)=>{X.preventDefault(),X.stopPropagation(),q(!j)}}>
                ${j?"▼":"▶"} ${B}
            </button>
            ${j&&G`
                ${Q&&G`<pre class="resource-embed-content">${Q}</pre>`}
                ${J&&G`
                    <div class="resource-embed-blob">
                        <span class="resource-embed-blob-label">Embedded blob</span>
                        ${K&&G`<span class="resource-embed-blob-meta">${K}</span>`}
                        <button class="resource-embed-blob-btn" onClick=${(X)=>{X.preventDefault(),X.stopPropagation();let z=new Blob([Uint8Array.from(atob($.data),(N)=>N.charCodeAt(0))],{type:K||"application/octet-stream"}),U=URL.createObjectURL(z),L=document.createElement("a");L.href=U,L.download=B.split("/").pop()||"resource",L.click(),URL.revokeObjectURL(U)}}>Download</button>
                    </div>
                `}
            `}
        </div>
    `}function A$({block:$,post:j,onOpenWidget:q}){if(!$)return null;let B=$3($,j),Q=e6($),J=B?.artifact?.kind||$?.artifact?.kind||$?.kind||null,K=B?.title||$.title||$.name||"Generated widget",X=B?.description||$.description||$.subtitle||"",z=$.open_label||"Open widget",U=T(!1),L=(N)=>{if(N)N.preventDefault(),N.stopPropagation();if(!B)return;q?.(B)};return w(()=>{if(!$?.auto_open||!B||!Q||U.current)return;let N=j?.timestamp?new Date(j.timestamp).getTime():0;if(N&&Date.now()-N>1e4)return;let V=`widget_opened_${$.widget_id||j?.id||""}`;if(Q8(sessionStorage,V))return;U.current=!0,J8(sessionStorage,V,"1"),q?.(B)},[$?.auto_open,B,Q]),G`
        <div class="generated-widget-launch" onClick=${(N)=>N.stopPropagation()}>
            <div class="generated-widget-launch-header">
                <div class="generated-widget-launch-eyebrow">Generated widget${J?` • ${String(J).toUpperCase()}`:""}</div>
                <div class="generated-widget-launch-title">${K}</div>
            </div>
            ${X&&G`<div class="generated-widget-launch-description">${X}</div>`}
            <div class="generated-widget-launch-actions">
                <button
                    class="generated-widget-launch-btn"
                    type="button"
                    disabled=${!Q}
                    onClick=${L}
                    title=${Q?"Open widget in a floating pane":"Unsupported widget artifact"}
                >
                    ${z}
                </button>
                <span class="generated-widget-launch-note">
                    ${Q?"Opens in a dismissible floating pane.":"This widget artifact is missing or unsupported."}
                </span>
            </div>
        </div>
    `}function D$($){if(!$)return"\uD83D\uDCCE";if($.startsWith("image/"))return"\uD83D\uDDBC️";if($.startsWith("audio/"))return"\uD83C\uDFB5";if($.startsWith("video/"))return"\uD83C\uDFAC";if($.includes("pdf"))return"\uD83D\uDCC4";if($.includes("zip")||$.includes("gzip"))return"\uD83D\uDDDC️";if($.startsWith("text/"))return"\uD83D\uDCC4";return"\uD83D\uDCCE"}function C$($){let j=T2($,{allowDataImage:!0});return j?{backgroundImage:`url("${j}")`}:void 0}function I$({preview:$}){let j=T2($.url),q=C$($.image),B=K8($.site_name,j);return G`
        <a
            href=${j||"#"}
            class="link-preview ${q?"has-image":""}"
            target=${j?"_blank":void 0}
            rel=${j?"noopener noreferrer":void 0}
            onClick=${(Q)=>Q.stopPropagation()}
            style=${q}>
            <div class="link-preview-overlay">
                <div class="link-preview-site">${B||""}</div>
                <div class="link-preview-title">${$.title}</div>
                ${$.description&&G`
                    <div class="link-preview-description">${$.description}</div>
                `}
            </div>
        </a>
    `}function O$($,j){return typeof $==="string"?$:""}var X8=1800,T$=`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>`,M$=`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path d="M20 6L9 17l-5-5"></path>
    </svg>`,S$=`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <circle cx="12" cy="12" r="9"></circle>
        <path d="M9 9l6 6M15 9l-6 6"></path>
    </svg>`,y$=`
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
</style>`;async function U8($){let j=typeof $==="string"?$:"";if(!j)return!1;if(q3(document,{text:j}))return!0;if(await q8(navigator.clipboard,j))return!0;try{let q=document.createElement("textarea");q.value=j,q.setAttribute("readonly",""),q.style.position="fixed",q.style.opacity="0",q.style.pointerEvents="none",document.body.appendChild(q),q.select(),q.setSelectionRange(0,q.value.length);let B=document.execCommand("copy");return document.body.removeChild(q),B}catch{return!1}}async function k$($){let j=typeof $==="string"?$:"";if(!j)return!1;let q=M2(j,null),B=`<html><head>${y$}</head><body>${q}</body></html>`;if(q3(document,{text:j,html:B}))return!0;if(navigator.clipboard?.write&&typeof ClipboardItem<"u")try{let Q=new ClipboardItem({"text/plain":new Blob([j],{type:"text/plain"}),"text/html":new Blob([B],{type:"text/html"})});return await navigator.clipboard.write([Q]),!0}catch(Q){console.warn("[post] Rich clipboard write failed, falling back to plain text copy.",Q)}return U8(j)}function R$($){if(!$)return()=>{};let j=Array.from($.querySelectorAll("pre")).filter((K)=>K.querySelector("code"));if(j.length===0)return()=>{};let q=new Map,B=[],Q=(K)=>{let X=window.getSelection?.();if(!X||X.isCollapsed)return;for(let z of j)if(B8(K,{root:z,selection:X}))return};document.addEventListener("copy",Q,!0),B.push(()=>document.removeEventListener("copy",Q,!0));let J=(K,X)=>{let z=X||"idle";if(K.dataset.copyState=z,z==="success")K.innerHTML=M$,K.setAttribute("aria-label","Copied"),K.setAttribute("title","Copied"),K.classList.add("is-success"),K.classList.remove("is-error");else if(z==="error")K.innerHTML=S$,K.setAttribute("aria-label","Copy failed"),K.setAttribute("title","Copy failed"),K.classList.add("is-error"),K.classList.remove("is-success");else K.innerHTML=T$,K.setAttribute("aria-label","Copy code"),K.setAttribute("title","Copy code"),K.classList.remove("is-success","is-error")};return j.forEach((K)=>{let X=document.createElement("div");X.className="post-code-block",K.parentNode?.insertBefore(X,K),X.appendChild(K);let z=document.createElement("button");z.type="button",z.className="post-code-copy-btn",J(z,"idle"),X.appendChild(z);let U=async(L)=>{L.preventDefault(),L.stopPropagation();let V=K.querySelector("code")?.textContent||"",O=await U8(V);J(z,O?"success":"error");let g=q.get(z);if(g)clearTimeout(g);let h=setTimeout(()=>{J(z,"idle"),q.delete(z)},X8);q.set(z,h)};z.addEventListener("click",U),B.push(()=>{z.removeEventListener("click",U);let L=q.get(z);if(L)clearTimeout(L);if(X.parentNode)X.parentNode.insertBefore(K,X),X.remove()})}),()=>{B.forEach((K)=>K())}}function E$($){if(!$)return{content:$,fileRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let U=0;U<q.length;U+=1)if(q[U].trim()==="Files:"&&q[U+1]&&/^\s*-\s+/.test(q[U+1])){B=U;break}if(B===-1)return{content:$,fileRefs:[]};let Q=[],J=B+1;for(;J<q.length;J+=1){let U=q[J];if(/^\s*-\s+/.test(U))Q.push(U.replace(/^\s*-\s+/,"").trim());else if(!U.trim())break;else break}if(Q.length===0)return{content:$,fileRefs:[]};let K=q.slice(0,B),X=q.slice(J),z=[...K,...X].join(`
`);return z=z.replace(/\n{3,}/g,`

`).trim(),{content:z,fileRefs:Q}}function P$($){if(!$)return{content:$,messageRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let U=0;U<q.length;U+=1)if(q[U].trim()==="Referenced messages:"&&q[U+1]&&/^\s*-\s+/.test(q[U+1])){B=U;break}if(B===-1)return{content:$,messageRefs:[]};let Q=[],J=B+1;for(;J<q.length;J+=1){let U=q[J];if(/^\s*-\s+/.test(U)){let N=U.replace(/^\s*-\s+/,"").trim().match(/^message:(\S+)$/i);if(N)Q.push(N[1])}else if(!U.trim())break;else break}if(Q.length===0)return{content:$,messageRefs:[]};let K=q.slice(0,B),X=q.slice(J),z=[...K,...X].join(`
`);return z=z.replace(/\n{3,}/g,`

`).trim(),{content:z,messageRefs:Q}}function f$($){if(!$)return{content:$,attachments:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let U=0;U<q.length;U+=1){let L=q[U].trim();if((L==="Images:"||L==="Attachments:")&&q[U+1]&&/^\s*-\s+/.test(q[U+1])){B=U;break}}if(B===-1)return{content:$,attachments:[]};let Q=[],J=B+1;for(;J<q.length;J+=1){let U=q[J];if(/^\s*-\s+/.test(U)){let L=U.replace(/^\s*-\s+/,"").trim(),N=L.match(/^attachment:([^\s)]+)\s*(?:\((.+)\))?$/i)||L.match(/^attachment:([^\s]+)\s+(.+)$/i);if(N){let V=N[1],O=(N[2]||"").trim()||V;Q.push({id:V,label:O,raw:L})}else Q.push({id:null,label:L,raw:L})}else if(!U.trim())break;else break}if(Q.length===0)return{content:$,attachments:[]};let K=q.slice(0,B),X=q.slice(J),z=[...K,...X].join(`
`);return z=z.replace(/\n{3,}/g,`

`).trim(),{content:z,attachments:Q}}function w$($){return $.replace(/[.*+?^${}()|[\]\\]/g,"\\$&")}function x$($,j){if(!$||!j)return $;let q=String(j).trim().split(/\s+/).filter(Boolean);if(q.length===0)return $;let B=q.map(w$).sort((L,N)=>N.length-L.length),Q=new RegExp(`(${B.join("|")})`,"gi"),J=new RegExp(`^(${B.join("|")})$`,"i"),K=new DOMParser().parseFromString($,"text/html"),X=K.createTreeWalker(K.body,NodeFilter.SHOW_TEXT),z=[],U;while(U=X.nextNode())z.push(U);for(let L of z){let N=L.nodeValue;if(!N||!Q.test(N)){Q.lastIndex=0;continue}Q.lastIndex=0;let V=L.parentElement;if(V&&V.closest("code, pre, script, style"))continue;let O=N.split(Q).filter((h)=>h!=="");if(O.length===0)continue;let g=K.createDocumentFragment();for(let h of O)if(J.test(h)){let p=K.createElement("mark");p.className="search-highlight-term",p.textContent=h,g.appendChild(p)}else g.appendChild(K.createTextNode(h));L.parentNode.replaceChild(g,L)}return K.body.innerHTML}function _8({post:$,onClick:j,onHashtagClick:q,onMessageRef:B,onScrollToMessage:Q,agentName:J,agentAvatarUrl:K,userName:X,userAvatarUrl:z,userAvatarBackground:U,onDelete:L,isThreadReply:N,isThreadPrev:V,isThreadNext:O,isRemoving:g,highlightQuery:h,onFileRef:p,onOpenWidget:d,onOpenAttachmentPreview:P}){let[t,v]=M(null),[e,W1]=M("idle"),D1=T(null),N1=T(null),X1=$.data,x=X1.type==="agent_response",U1=X||"You",E1=x?J||h6:U1,b1=typeof $.chat_agent_name==="string"?$.chat_agent_name.trim():"",P1=Boolean(x&&h&&b1&&b1!==E1),l=x?n5(J,K,!0):n5(U1,z),s=typeof U==="string"?U.trim().toLowerCase():"",Q1=!x&&l.image&&(s==="clear"||s==="transparent"),_1=x&&Boolean(l.image),D=`background-color: ${Q1||_1?"transparent":l.color}`,r=X1.content_meta,L1=Boolean(r?.truncated),j1=Boolean(r?.preview),M1=L1&&!j1,R1=L1?{originalLength:Number.isFinite(r?.original_length)?r.original_length:X1.content?X1.content.length:0,maxLength:Number.isFinite(r?.max_length)?r.max_length:0}:null,Y1=X1.content_blocks||[],p1=X1.media_ids||[],C1=O$(X1.content,X1.link_previews),{content:j0,fileRefs:s1}=E$(C1),{content:O0,messageRefs:K0}=P$(j0),{content:f0,attachments:G0}=f$(O0);C1=f0;let o1=e5(Y1),q0=i5(Y1),y1=G$(Y1)[0]||null,N0=N$(Y1)[0]||null,a1=o1.length===1&&typeof o1[0]?.fallback_text==="string"?o1[0].fallback_text.trim():"",t1=q0.length===1?i4(q0[0]).trim():"",a=Boolean(a1)&&C1?.trim()===a1||Boolean(t1)&&C1?.trim()===t1,z1=Boolean(C1)&&!M1&&!a,J1=typeof h==="string"?h.trim():"",A1=H1(()=>{if(!C1||a)return"";let Y=M2(C1,q);return J1?x$(Y,J1):Y},[C1,a,J1]),k1=H1(()=>m6($),[$]),c1=(Y,S)=>{Y.stopPropagation(),v(t2(S))},f1=(Y)=>{P?.(Y)},Y0=(Y)=>{Y.stopPropagation(),L?.($)},e1=async(Y)=>{Y.stopPropagation();let S=await k$(k1);if(W1(S?"success":"error"),N1.current)clearTimeout(N1.current);N1.current=setTimeout(()=>{N1.current=null,W1("idle")},X8)},l0=(Y,S)=>{let $1=new Set;if(!Y||S.length===0)return{content:Y,usedIds:$1};return{content:Y.replace(/attachment:([^\s)"']+)/g,(c,O1,w1,T1)=>{let S1=O1.replace(/^\/+/,""),I1=S.find((T0)=>T0.name&&T0.name.toLowerCase()===S1.toLowerCase()&&!$1.has(T0.id))||S.find((T0)=>!$1.has(T0.id));if(!I1)return c;if($1.add(I1.id),T1.slice(Math.max(0,w1-2),w1)==="](")return`/media/${I1.id}`;return I1.name||"attachment"}),usedIds:$1}},B0=[],m1=[],$0=[],Z0=[],A0=[],h1=[],Q0=[],D0=0;if(Y1.length>0)Y1.forEach((Y)=>{if(Y?.type==="text"&&Y.annotations)Q0.push(Y.annotations);if(Y?.type==="generated_widget")h1.push(Y);else if(Y?.type==="resource_link")Z0.push(Y);else if(Y?.type==="resource")A0.push(Y);else if(Y?.type==="file"){let S=p1[D0++];if(S)m1.push(S),$0.push({id:S,name:Y?.name||Y?.filename||Y?.title})}else if(Y?.type==="image"||!Y?.type){let S=p1[D0++];if(S){let $1=typeof Y?.mime_type==="string"?Y.mime_type:void 0;B0.push({id:S,annotations:Y?.annotations,mimeType:$1}),$0.push({id:S,name:Y?.name||Y?.filename||Y?.title})}}});else if(p1.length>0){let Y=G0.length>0;p1.forEach((S,$1)=>{let K1=G0[$1]||null;if($0.push({id:S,name:K1?.label||null}),Y)m1.push(S);else B0.push({id:S,annotations:null})})}if(G0.length>0)G0.forEach((Y)=>{if(!Y?.id)return;let S=$0.find(($1)=>String($1.id)===String(Y.id));if(S&&!S.name)S.name=Y.label});let{content:C,usedIds:b}=l0(C1,$0);C1=C;let E=B0.filter(({id:Y})=>!b.has(Y)),i=m1.filter((Y)=>!b.has(Y)),n=G0.length>0?G0.map((Y,S)=>({id:Y.id||`attachment-${S+1}`,label:Y.label||`attachment-${S+1}`})):$0.map((Y,S)=>({id:Y.id,label:Y.name||`attachment-${S+1}`})),V1=H1(()=>e5(Y1),[Y1]),Z1=H1(()=>i5(Y1),[Y1]),G1=H1(()=>{return V1.map((Y)=>`${Y.card_id}:${Y.state}`).join("|")},[V1]);w(()=>{if(!D1.current)return;return x6(D1.current),R$(D1.current)},[A1]),w(()=>()=>{if(N1.current)clearTimeout(N1.current)},[]);let F1=T(null);return w(()=>{if(!F1.current||V1.length===0)return;let Y=F1.current;Y.innerHTML="";for(let S of V1){let $1=document.createElement("div");Y.appendChild($1),t6($1,S,{onAction:async(K1)=>{if(K1.type==="Action.OpenUrl"){let c=T2(K1.url||"");if(!c)throw Error("Invalid URL");window.open(c,"_blank","noopener,noreferrer");return}if(K1.type==="Action.Submit"){await F6({post_id:$.id,thread_id:X1.thread_id||$.id,chat_jid:$.chat_jid||null,card_id:S.card_id,action:{type:K1.type,title:K1.title||"",data:K1.data}});return}console.warn("[post] unsupported adaptive card action:",K1.type,K1)}}).catch((K1)=>{console.error("[post] adaptive card render error:",K1),$1.textContent=S.fallback_text||"Card failed to render."})}},[G1,$.id]),G`
        <div id=${`post-${$.id}`} class="post ${x?"agent-post":""} ${N?"thread-reply":""} ${V?"thread-prev":""} ${O?"thread-next":""} ${g?"removing":""}" onClick=${j}>
            <div class="post-avatar ${x?"agent-avatar":""} ${l.image?"has-image":""}" style=${D}>
                ${l.image?G`<img src=${l.image} alt=${E1} />`:l.letter}
            </div>
            <div class="post-body">
                <div class="post-actions">
                    <button
                        class=${`post-action-btn post-copy-btn${e==="success"?" is-success":e==="error"?" is-error":""}`}
                        type="button"
                        title=${e==="success"?"Copied":e==="error"?"Copy failed":"Copy message"}
                        aria-label=${e==="success"?"Copied":e==="error"?"Copy failed":"Copy message"}
                        onClick=${e1}
                        disabled=${!k1}
                    >
                        ${e==="success"?G`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M20 6L9 17l-5-5"></path></svg>`:e==="error"?G`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><circle cx="12" cy="12" r="9"></circle><path d="M9 9l6 6M15 9l-6 6"></path></svg>`:G`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><rect x="9" y="9" width="10" height="10" rx="2"></rect><path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path></svg>`}
                    </button>
                    <button
                        class="post-action-btn post-delete-btn"
                        type="button"
                        title="Delete message"
                        aria-label="Delete message"
                        onClick=${Y0}
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                            <path d="M18 6L6 18M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                <div class="post-meta">
                    <span class="post-author">${E1}</span>
                    ${P1&&G`<span class="post-chat-agent-tag" title=${`Chat: ${b1}`}>@${b1}</span>`}
                    ${y1&&G`
                        <span
                            class="post-recovery-chip"
                            title=${V$(y1)}
                        >
                            recovered
                        </span>
                    `}
                    ${N0&&G`
                        <span
                            class="post-recovery-chip post-timeout-chip"
                            title=${Z$(N0)}
                        >
                            timeout
                        </span>
                    `}
                    <a class="post-time" href=${`#msg-${$.id}`} onClick=${(Y)=>{if(Y.preventDefault(),Y.stopPropagation(),B)B($.id)}}>${b6($.timestamp)}</a>
                </div>
                ${M1&&R1&&G`
                    <div class="post-content truncated">
                        <div class="truncated-title">Message too large to display.</div>
                        <div class="truncated-meta">
                            Original length: ${$4(R1.originalLength)} chars
                            ${R1.maxLength?G` • Display limit: ${$4(R1.maxLength)} chars`:""}
                        </div>
                    </div>
                `}
                ${j1&&R1&&G`
                    <div class="post-content preview">
                        <div class="truncated-title">Preview truncated.</div>
                        <div class="truncated-meta">
                            Showing first ${$4(R1.maxLength)} of ${$4(R1.originalLength)} chars. Download full text below.
                        </div>
                    </div>
                `}
                ${(s1.length>0||K0.length>0||n.length>0)&&G`
                    <div class="post-file-refs">
                        ${K0.map((Y)=>{let S=($1)=>{if($1.preventDefault(),$1.stopPropagation(),Q)Q(Y,$.chat_jid||null);else{let K1=document.getElementById("post-"+Y);if(K1)K1.scrollIntoView({behavior:"smooth",block:"center"}),K1.classList.add("post-highlight"),setTimeout(()=>K1.classList.remove("post-highlight"),2000)}};return G`
                                <a href=${`#msg-${Y}`} class="post-msg-pill-link" onClick=${S}>
                                    <${P0}
                                        prefix="post"
                                        label=${"msg:"+Y}
                                        title=${"Message "+Y}
                                        icon="message"
                                        onClick=${S}
                                    />
                                </a>
                            `})}
                        ${s1.map((Y)=>{let S=Y.split("/").pop()||Y;return G`
                                <${P0}
                                    prefix="post"
                                    label=${S}
                                    title=${Y}
                                    onClick=${()=>p?.(Y)}
                                />
                            `})}
                        ${n.map((Y)=>G`
                            <${F$}
                                key=${Y.id}
                                attachment=${Y}
                                onPreview=${f1}
                            />
                        `)}
                    </div>
                `}
                ${z1&&G`
                    <div 
                        ref=${D1}
                        class="post-content"
                        dangerouslySetInnerHTML=${{__html:A1}}
                        onClick=${(Y)=>{if(Y.target.classList.contains("hashtag")){Y.preventDefault(),Y.stopPropagation();let S=Y.target.dataset.hashtag;if(S)q?.(S)}else if(Y.target.tagName==="IMG")Y.preventDefault(),Y.stopPropagation(),v(Y.target.src)}}
                    />
                `}
                ${V1.length>0&&G`
                    <div ref=${F1} class="post-adaptive-cards" />
                `}
                ${Z1.length>0&&G`
                    <div class="post-adaptive-card-submissions">
                        ${Z1.map((Y,S)=>{let $1=u6(Y),K1=`${Y.card_id}-${S}`;return G`
                                <div key=${K1} class="adaptive-card-submission-receipt">
                                    <div class="adaptive-card-submission-header">
                                        <span class="adaptive-card-submission-icon" aria-hidden="true">✓</span>
                                        <div class="adaptive-card-submission-title-wrap">
                                            <span class="adaptive-card-submission-title">Submitted</span>
                                            <span class="adaptive-card-submission-title-action">${$1.title}</span>
                                        </div>
                                    </div>
                                    ${$1.fields.length>0&&G`
                                        <div class="adaptive-card-submission-fields">
                                            ${$1.fields.map((c)=>G`
                                                <span class="adaptive-card-submission-field" title=${`${c.key}: ${c.value}`}>
                                                    <span class="adaptive-card-submission-field-key">${c.key}</span>
                                                    <span class="adaptive-card-submission-field-sep">:</span>
                                                    <span class="adaptive-card-submission-field-value">${c.value}</span>
                                                </span>
                                            `)}
                                        </div>
                                    `}
                                    <div class="adaptive-card-submission-meta">
                                        Submitted ${r4($1.submittedAt)}
                                    </div>
                                </div>
                            `})}
                    </div>
                `}
                ${h1.length>0&&G`
                    <div class="generated-widget-launches">
                        ${h1.map((Y,S)=>G`
                            <${A$}
                                key=${Y.widget_id||Y.id||`${$.id}-widget-${S}`}
                                block=${Y}
                                post=${$}
                                onOpenWidget=${d}
                            />
                        `)}
                    </div>
                `}
                ${Q0.length>0&&G`
                    ${Q0.map((Y,S)=>G`
                        <${o4} key=${S} annotations=${Y} />
                    `)}
                `}
                ${E.length>0&&G`
                    <div class="media-preview">
                        ${E.map(({id:Y,mimeType:S})=>{let K1=typeof S==="string"&&S.toLowerCase().startsWith("image/svg")?t2(Y):Z6(Y);return G`
                                <img 
                                    key=${Y} 
                                    src=${K1} 
                                    alt="Media" 
                                    loading="lazy"
                                    decoding="async"
                                    onClick=${(c)=>c1(c,Y)}
                                />
                            `})}
                    </div>
                `}
                ${E.length>0&&G`
                    ${E.map(({annotations:Y},S)=>G`
                        ${Y&&G`<${o4} key=${S} annotations=${Y} />`}
                    `)}
                `}
                ${i.length>0&&G`
                    <div class="file-attachments">
                        ${i.map((Y)=>G`
                            <${W$} key=${Y} mediaId=${Y} onPreview=${f1} />
                        `)}
                    </div>
                `}
                ${Z0.length>0&&G`
                    <div class="resource-links">
                        ${Z0.map((Y,S)=>G`
                            <div key=${S}>
                                <${H$} block=${Y} />
                                <${o4} annotations=${Y.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${A0.length>0&&G`
                    <div class="resource-embeds">
                        ${A0.map((Y,S)=>G`
                            <div key=${S}>
                                <${Y$} block=${Y} />
                                <${o4} annotations=${Y.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${X1.link_previews?.length>0&&G`
                    <div class="link-previews">
                        ${X1.link_previews.map((Y,S)=>G`
                            <${I$} key=${S} preview=${Y} />
                        `)}
                    </div>
                `}
            </div>
        </div>
        ${t&&G`<${$8} src=${t} onClose=${()=>v(null)} />`}

    `}function z8({posts:$,hasMore:j,onLoadMore:q,onPostClick:B,onHashtagClick:Q,onMessageRef:J,onScrollToMessage:K,onFileRef:X,onOpenWidget:z,onOpenAttachmentPreview:U,emptyMessage:L,timelineRef:N,agents:V,user:O,onDeletePost:g,reverse:h=!0,removingPostIds:p,searchQuery:d}){let[P,t]=M(!1),v=T(null),e=typeof IntersectionObserver<"u",W1=m(async()=>{if(!q||!j||P)return;t(!0);try{await q({preserveScroll:!0,preserveMode:"top"})}finally{t(!1)}},[j,P,q]),D1=m((l)=>{let{scrollTop:s,scrollHeight:Q1,clientHeight:_1}=l.target,D=h?Q1-_1-s:s,r=Math.max(300,_1);if(D<r)W1()},[h,W1]);w(()=>{if(!e)return;let l=v.current,s=N?.current;if(!l||!s)return;let Q1=300,_1=new IntersectionObserver((D)=>{for(let r of D){if(!r.isIntersecting)continue;W1()}},{root:s,rootMargin:`${Q1}px 0px ${Q1}px 0px`,threshold:0});return _1.observe(l),()=>_1.disconnect()},[e,j,q,N,W1]);let N1=T(W1);if(N1.current=W1,w(()=>{if(e)return;if(!N?.current)return;let{scrollTop:l,scrollHeight:s,clientHeight:Q1}=N.current,_1=h?s-Q1-l:l,D=Math.max(300,Q1);if(_1<D)N1.current?.()},[e,$,j,h,N]),w(()=>{if(!N?.current)return;if(!j||P)return;let{scrollTop:l,scrollHeight:s,clientHeight:Q1}=N.current,_1=h?s-Q1-l:l,D=Math.max(300,Q1);if(s<=Q1+1||_1<D)N1.current?.()},[$,j,P,h,N]),!$)return G`<div class="loading"><div class="spinner"></div></div>`;if($.length===0)return G`
            <div class="timeline" ref=${N}>
                <div class="timeline-content">
                    <div style="padding: var(--spacing-xl); text-align: center; color: var(--text-secondary)">
                        ${L||"No messages yet. Start a conversation!"}
                    </div>
                </div>
            </div>
        `;let X1=$.slice().sort((l,s)=>l.id-s.id),x=(l)=>{let s=l?.data?.thread_id;if(s===null||s===void 0||s==="")return null;let Q1=Number(s);return Number.isFinite(Q1)?Q1:null},U1=new Map;for(let l=0;l<X1.length;l+=1){let s=X1[l],Q1=Number(s?.id),_1=x(s);if(_1!==null){let D=U1.get(_1)||{anchorIndex:-1,replyIndexes:[]};D.replyIndexes.push(l),U1.set(_1,D)}else if(Number.isFinite(Q1)){let D=U1.get(Q1)||{anchorIndex:-1,replyIndexes:[]};D.anchorIndex=l,U1.set(Q1,D)}}let E1=new Map;for(let[l,s]of U1.entries()){let Q1=new Set;if(s.anchorIndex>=0)Q1.add(s.anchorIndex);for(let _1 of s.replyIndexes)Q1.add(_1);E1.set(l,Array.from(Q1).sort((_1,D)=>_1-D))}let b1=X1.map((l,s)=>{let Q1=x(l);if(Q1===null)return{hasThreadPrev:!1,hasThreadNext:!1};let _1=E1.get(Q1);if(!_1||_1.length===0)return{hasThreadPrev:!1,hasThreadNext:!1};let D=_1.indexOf(s);if(D<0)return{hasThreadPrev:!1,hasThreadNext:!1};return{hasThreadPrev:D>0,hasThreadNext:D<_1.length-1}}),P1=G`<div class="timeline-sentinel" ref=${v}></div>`;return G`
        <div class="timeline ${h?"reverse":"normal"}" ref=${N} onScroll=${D1}>
            <div class="timeline-content">
                ${h?P1:null}
                ${X1.map((l,s)=>{let Q1=Boolean(l.data?.thread_id&&l.data.thread_id!==l.id),_1=p?.has?.(l.id),D=b1[s]||{};return G`
                    <${_8}
                        key=${l.id}
                        post=${l}
                        isThreadReply=${Q1}
                        isThreadPrev=${D.hasThreadPrev}
                        isThreadNext=${D.hasThreadNext}
                        isRemoving=${_1}
                        highlightQuery=${d}
                        agentName=${p6(l.data?.agent_id,V||{})}
                        agentAvatarUrl=${c6(l.data?.agent_id,V||{})}
                        userName=${O?.name||O?.user_name}
                        userAvatarUrl=${O?.avatar_url||O?.avatarUrl||O?.avatar}
                        userAvatarBackground=${O?.avatar_background||O?.avatarBackground}
                        onClick=${()=>B?.(l)}
                        onHashtagClick=${Q}
                        onMessageRef=${J}
                        onScrollToMessage=${K}
                        onFileRef=${X}
                        onOpenWidget=${z}
                        onDelete=${g}
                        onOpenAttachmentPreview=${U}
                    />
                `})}
                ${h?null:P1}
            </div>
        </div>
    `}function a4($){return String($||"").toLowerCase().replace(/^@/,"").replace(/\s+/g," ").trim()}function b$($,j){let q=a4($),B=a4(j);if(!B)return!1;return q.startsWith(B)||q.includes(B)}function B3($){if(!$)return!1;if($.isComposing)return!1;if($.ctrlKey||$.metaKey||$.altKey)return!1;return typeof $.key==="string"&&$.key.length===1&&/\S/.test($.key)}function Q3($,j,q=Date.now(),B=700){let Q=$&&typeof $==="object"?$:{value:"",updatedAt:0},J=String(j||"").trim().toLowerCase();if(!J)return{value:"",updatedAt:q};return{value:!Q.value||!Number.isFinite(Q.updatedAt)||q-Q.updatedAt>B?J:`${Q.value}${J}`,updatedAt:q}}function v$($,j){let q=Math.max(0,Number($)||0);if(q<=0)return[];let Q=((Number.isInteger(j)?j:0)%q+q)%q,J=[];for(let K=0;K<q;K+=1)J.push((Q+K)%q);return J}function u$($,j,q=0,B=(Q)=>Q){let Q=a4(j);if(!Q)return-1;let J=Array.isArray($)?$:[],K=v$(J.length,q),X=J.map((z)=>a4(B(z)));for(let z of K)if(X[z].startsWith(Q))return z;for(let z of K)if(X[z].includes(Q))return z;return-1}function J3($,j,q=-1,B=(Q)=>Q){let Q=Array.isArray($)?$:[];if(q>=0&&q<Q.length){let J=B(Q[q]);if(b$(J,j))return q}return u$(Q,j,0,B)}function t4($){return String($||"").trim().toLowerCase()}function K3($){let j=String($||"").match(/^@([a-zA-Z0-9_-]*)$/);if(!j)return null;return t4(j[1]||"")}function g$($){let j=new Set,q=[];for(let B of Array.isArray($)?$:[]){let Q=t4(B?.agent_name);if(!Q||j.has(Q))continue;j.add(Q),q.push(B)}return q}function W8($,j,q={}){let B=K3(j);if(B==null)return[];let Q=typeof q?.currentChatJid==="string"?q.currentChatJid:null;return g$($).filter((J)=>{if(Q&&J?.chat_jid===Q)return!1;return t4(J?.agent_name).startsWith(B)})}function X3($){let j=t4($);return j?`@${j} `:""}function G8($,j,q={}){if(!$||$.isComposing)return!1;if(q.searchMode)return!1;if(!q.showSessionSwitcherButton)return!1;if($.ctrlKey||$.metaKey||$.altKey)return!1;if($.key!=="@")return!1;return String(j||"")===""}function N8($){let j=m$($);return j?`@${j}`:""}function m$($){return String($||"").trim().toLowerCase().replace(/[^a-z0-9_-]+/g,"-").replace(/^-+|-+$/g,"").replace(/-{2,}/g,"-")}function L8($,j){let q=typeof $?.agent_name==="string"&&$.agent_name.trim()?N8($.agent_name):String(j||"").trim(),B=typeof $?.chat_jid==="string"&&$.chat_jid.trim()?$.chat_jid.trim():String(j||"").trim();return`${q} — ${B} • current branch`}function h$($,j={}){let q=[],B=typeof j.currentChatJid==="string"?j.currentChatJid.trim():"",Q=typeof $?.chat_jid==="string"?$.chat_jid.trim():"";if(B&&Q===B)q.push("current");if($?.archived_at)q.push("archived");else if($?.is_active)q.push("active");return q}function V8($,j={}){let q=N8($?.agent_name)||String($?.chat_jid||"").trim(),B=typeof $?.chat_jid==="string"&&$.chat_jid.trim()?$.chat_jid.trim():"unknown-chat",Q=h$($,j);return Q.length>0?`${q} — ${B} • ${Q.join(" • ")}`:`${q} — ${B}`}function e4({steerQueued:$=!1,pulsing:j=!1}={}){let q=["turn-dot"];if($)q.push("turn-dot-queued");if(j)q.push("turn-dot-pulsing");return q.join(" ")}function U3({pulsing:$=!1}={}){let j=["compose-inline-status-dot"];if($)j.push("compose-inline-status-dot-pulsing");return j.join(" ")}function $5($,{isLastActivity:j=!1,pendingRequest:q=!1}={}){if(q)return"dot";if(j)return"none";if($?.type==="error")return"none";if($?.type==="intent")return"dot";let B=typeof $?.type==="string"?$.type:"";if(Boolean(typeof $?.tool_name==="string"&&$.tool_name.trim()||$?.tool_args))return"spinner";if(B==="tool_call"||B==="tool_status"||B==="thinking"||B==="waiting")return"spinner";return"dot"}function Z8($,j={}){return $5($,j)==="dot"}function j5($){if(!$||typeof $!=="object")return null;let j=$.started_at??$.startedAt;if(typeof j!=="string"||!j)return null;let q=Date.parse(j);return Number.isFinite(q)?q:null}function p$($){if(!$||typeof $!=="object")return null;let j=$.retry_at??$.retryAt;if(typeof j!=="string"||!j)return null;let q=Date.parse(j);return Number.isFinite(q)?q:null}function _3($){if(!$||typeof $!=="object")return null;let j=$.last_event_at??$.lastEventAt??$.started_at??$.startedAt;if(typeof j!=="string"||!j)return null;let q=Date.parse(j);return Number.isFinite(q)?q:null}function g2($){if(!$||typeof $!=="object")return!1;let j=$.intent_key??$.intentKey;return $.type==="intent"&&j==="compaction"}function q5($){if(!$||typeof $!=="object")return"";let j=$.title;if(typeof j==="string"&&j.trim())return j.trim();let q=$.status;if(typeof q==="string"&&q.trim())return q.trim();return g2($)?"Compacting context":"Working..."}function F8($){let j=Math.max(0,Math.floor($/1000)),q=j%60,B=Math.floor(j/60)%60,Q=Math.floor(j/3600);if(Q>0)return`${Q}:${String(B).padStart(2,"0")}:${String(q).padStart(2,"0")}`;return`${B}:${String(q).padStart(2,"0")}`}function B5($,j=Date.now()){let q=j5($);if(q===null)return null;return F8(Math.max(0,j-q))}function H8($,j=Date.now()){let q=p$($);if(q===null)return null;let B=q-j;if(B<=0)return"retrying now";return`retry in ${F8(B)}`}var Y8=350;function c$($){return String($||"Connecting").replace(/[-_]+/g," ").replace(/^./,(j)=>j.toUpperCase())}function d$($,j={}){let q=typeof $==="string"&&$.trim()?$.trim():"connecting";if(q==="connected")return{show:!1,statusClass:"connected",label:"Connected",title:"Connection: Connected"};if(q!=="disconnected"){let X=c$(q);return{show:!0,statusClass:q,label:X,title:`Connection: ${X}`}}let B=Number.isFinite(Number(j?.delayMs))?Math.max(0,Number(j.delayMs)):Y8,Q=Number.isFinite(Number(j?.nowMs))?Number(j.nowMs):Date.now(),J=Number.isFinite(Number(j?.disconnectedAtMs))?Number(j.disconnectedAtMs):Q;return Q-J>=B?{show:!0,statusClass:"disconnected",label:"Reconnecting",title:"Reconnecting"}:{show:!0,statusClass:"connecting",label:"Connecting",title:"Connecting"}}function z3($,j={}){let q=Number.isFinite(Number(j?.delayMs))?Math.max(0,Number(j.delayMs)):Y8,[B,Q]=M(null),[J,K]=M(()=>Date.now());return w(()=>{if($==="disconnected"){let X=Date.now();Q((z)=>z??X),K(X);return}Q(null),K(Date.now())},[$]),w(()=>{if($!=="disconnected"||B===null)return;let X=q-(Date.now()-B);if(X<=0)return;let z=setTimeout(()=>{K(Date.now())},X);return()=>clearTimeout(z)},[$,B,q]),H1(()=>d$($,{delayMs:q,disconnectedAtMs:B,nowMs:J}),[$,q,B,J])}async function W3($,j,q){if(typeof $!=="function")return!1;try{let B=await $(j);if(!B)return!1;return q(B),!0}catch(B){return!1}}var l$=[{name:"/model",description:"Select model or list available models"},{name:"/cycle-model",description:"Cycle to the next available model"},{name:"/thinking",description:"Show or set thinking/effort level"},{name:"/effort",description:"Show or set thinking/effort level (alias for /thinking)"},{name:"/cycle-thinking",description:"Cycle thinking level"},{name:"/theme",description:"Set UI theme (no name to show available themes)"},{name:"/meters",description:"Toggle the top-right CPU/RAM HUD (/meters on|off|toggle)"},{name:"/tint",description:"Tint default light/dark UI (usage: /tint #hex or /tint off)"},{name:"/btw",description:"Open a side conversation panel without interrupting the main chat"},{name:"/state",description:"Show current session state"},{name:"/stats",description:"Show session token and cost stats"},{name:"/context",description:"Show context window usage"},{name:"/last",description:"Show last assistant response"},{name:"/compact",description:"Manually compact the session"},{name:"/auto-compact",description:"Toggle auto-compaction"},{name:"/auto-retry",description:"Toggle auto-retry"},{name:"/abort",description:"Abort the current response"},{name:"/abort-retry",description:"Abort retry backoff"},{name:"/abort-bash",description:"Abort running bash command"},{name:"/shell",description:"Run a shell command and return output"},{name:"/bash",description:"Run a shell command and add output to context"},{name:"/queue",description:"Queue a follow-up message (one-at-a-time)"},{name:"/queue-all",description:"Queue a follow-up message (batch all)"},{name:"/steer",description:"Steer the current response"},{name:"/steering-mode",description:"Set steering mode (all|one)"},{name:"/followup-mode",description:"Set follow-up mode (all|one)"},{name:"/session-name",description:"Set or show the session name"},{name:"/new-session",description:"Start a new session"},{name:"/switch-session",description:"Switch to a session file"},{name:"/session-rotate",description:"Rotate the current persisted session into an archived file"},{name:"/clone",description:"Duplicate the current active branch into a new session"},{name:"/fork",description:"Fork from a previous message"},{name:"/forks",description:"List forkable messages"},{name:"/tree",description:"List the session tree"},{name:"/label",description:"Set or clear a label on a tree entry"},{name:"/labels",description:"List labeled entries"},{name:"/agent-name",description:"Set or show the agent display name"},{name:"/agent-avatar",description:"Set or show the agent avatar URL"},{name:"/user-name",description:"Set or show your display name"},{name:"/user-avatar",description:"Set or show your avatar URL"},{name:"/user-github",description:"Set name/avatar from GitHub profile"},{name:"/export-html",description:"Export session to HTML"},{name:"/passkey",description:"Manage passkeys (enrol/list/delete)"},{name:"/totp",description:"Show a TOTP enrolment QR code"},{name:"/qr",description:"Generate a QR code for text or URL"},{name:"/search",description:"Search notes and skills in the workspace"},{name:"/dream",description:"Run Dream memory maintenance over recent days (default 7)"},{name:"/tasks",description:"List scheduled tasks"},{name:"/scheduled",description:"List scheduled tasks"},{name:"/restart",description:"Restart the agent and stop subprocesses"},{name:"/exit",description:"Exit the current piclaw process immediately (Supervisor will restart it)"},{name:"/login",description:"Login to an AI model provider (OAuth or API key)"},{name:"/logout",description:"Logout from an AI model provider"},{name:"/commands",description:"List available commands"},{name:"/skill:",description:"Run a workspace skill (e.g. /skill:visual-artifact-generator, /skill:web-search)"}],A8="piclaw_compose_history";function r$($,j,q=!1){if(q)return{shouldApply:!1,nextToken:j,text:""};if(!$||typeof $!=="object")return{shouldApply:!1,nextToken:j,text:""};let B=typeof $.token==="string"?$.token:"",Q=typeof $.text==="string"?$.text:"";if(!B||B===j||!Q.trim())return{shouldApply:!1,nextToken:j,text:""};return{shouldApply:!0,nextToken:B,text:Q}}function i$($="web:default"){let j=typeof $==="string"&&$.trim()?$.trim():"web:default";if(j==="web:default")return A8;return`${A8}:${encodeURIComponent(j)}`}function D8($,j){let q=typeof j?.command?.message==="string"?j.command.message.trim():"";if(!j?.ui_only||!q)return null;let B=typeof $==="string"?$.trim():"";if(!B.startsWith("/"))return null;let Q=B.split(/\s+/).filter(Boolean),J=Q[0]?.toLowerCase()||"";if(!(Q.length>1)&&(J==="/model"||J==="/thinking"||J==="/effort"))return q;return null}function n$($,j,q=!1){if($&&q)return{mode:"compacting",className:"icon-btn send-btn abort-mode compacting-mode",title:"Compacting context — Stop response",ariaLabel:"Compacting context — Stop response",disabled:!1};if($)return{mode:"abort",className:"icon-btn send-btn abort-mode",title:"Stop response",ariaLabel:"Stop response",disabled:!1};return{mode:"send",className:"icon-btn send-btn",title:"Send (Enter)",ariaLabel:"Send message",disabled:!j}}function s$($){return $==="abort"||$==="compacting"}function o$($,j=0){let q=typeof $?.message==="string"&&$.message.trim()?$.message.trim():null,B=$?.indicator&&typeof $.indicator==="object"?$.indicator:null;if(!q&&!B)return{visible:!1,title:"",indicatorText:null,animateDot:!1};if(B?.mode==="hidden")return{visible:Boolean(q),title:q||"Working…",indicatorText:null,animateDot:!1};if(B?.mode==="custom"&&Array.isArray(B.frames)&&B.frames.length>0){let Q=B.frames,J=Number.isFinite(j)&&j>=0?Math.floor(j)%Q.length:0;return{visible:!0,title:q||"Working…",indicatorText:Q[J],animateDot:!1}}return{visible:!0,title:q||"Working…",indicatorText:null,animateDot:!0}}function a$({usage:$,onCompact:j}){let q=Math.min(100,Math.max(0,$.percent||0)),B=$.tokens,Q=$.contextWindow,J="Compact context",X=`${B!=null?`Context: ${G3(B)} / ${G3(Q)} tokens (${q.toFixed(0)}%)`:`Context: ${q.toFixed(0)}%`} — ${"Compact context"}`,z=9,U=2*Math.PI*9,L=q/100*U,N=q>90?"var(--context-red, #ef4444)":q>75?"var(--context-amber, #f59e0b)":"var(--context-green, #22c55e)";return G`
        <button
            class="compose-context-pie icon-btn"
            type="button"
            title=${X}
            aria-label="Compact context"
            onClick=${(V)=>{V.preventDefault(),V.stopPropagation(),j?.()}}
        >
            <svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="12" cy="12" r=${9}
                    fill="none"
                    stroke="var(--context-track, rgba(128,128,128,0.2))"
                    stroke-width="2.5" />
                <circle cx="12" cy="12" r=${9}
                    fill="none"
                    stroke=${N}
                    stroke-width="2.5"
                    stroke-dasharray=${`${L} ${U}`}
                    stroke-linecap="round"
                    transform="rotate(-90 12 12)" />
            </svg>
        </button>
    `}function G3($){if($==null)return"?";if($>=1e6)return($/1e6).toFixed(1)+"M";if($>=1000)return($/1000).toFixed(0)+"K";return String($)}function N3($){let j=Number($);if(!Number.isFinite(j)||j<=0)return"";return`${G3(j)} ctx`}function t$($,j){let q=typeof $==="string"?$.trim():"",B=N3(j);if(!q)return B;if(!B)return q;return`${q} • ${B}`}function e$($,j="",q=""){let B=typeof $==="string"?$.trim():"";if(B)return B;let Q=typeof j==="string"?j.trim():"",J=typeof q==="string"?q.trim():"";if(Q&&J)return`${Q}/${J}`;return J||Q||""}function C8($){let j=Array.isArray($?.model_options)?$.model_options:null,q=Array.isArray($?.models)?$.models:[],B=j&&j.length>0?j:q,Q=[];for(let J of B){if(typeof J==="string"){let N=J.trim();if(!N)continue;let V=N.indexOf("/"),O=V>0?N.slice(0,V).trim():"",g=V>0?N.slice(V+1).trim():N;Q.push({label:N,provider:O,id:g,name:null,contextWindow:null,reasoning:!1});continue}if(!J||typeof J!=="object")continue;let K=typeof J.provider==="string"?J.provider.trim():"",X=typeof J.id==="string"?J.id.trim():"",z=e$(J.label,K,X);if(!z)continue;let U=typeof J.name==="string"&&J.name.trim()?J.name.trim():null,L=Number(J.context_window??J.contextWindow);Q.push({label:z,provider:K,id:X,name:U,contextWindow:Number.isFinite(L)&&L>0?L:null,reasoning:Boolean(J.reasoning)})}return Q.sort((J,K)=>J.label.localeCompare(K.label,void 0,{sensitivity:"base"})),Q}function $j($){if(!$||typeof $!=="object")return"";return[$.label,$.provider,$.id,$.name,N3($.contextWindow)].filter(Boolean).join(" ")}function jj($,j){let q=typeof $==="string"?$.trim():"";if(q)return{showPicker:!0,label:q,hasAvailableModels:!0};let B=C8(j).length>0;return{showPicker:B,label:B?"Select model":"",hasAvailableModels:B}}function qj($){if(!$)return $;let j=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`);if(!j.includes(" @ ")||!j.includes(`:
`))return $;let q=j.split(`
`),B=[],Q=0,J=!1;while(Q<q.length){let X=q[Q].trim();if(!X){Q+=1;continue}if(X==="Messages:"||X.startsWith("Channel:")){J=!0,Q+=1;continue}if(/^[^\n]+\s@\s[^\n]+:$/.test(X)){J=!0,Q+=1;let z=[];while(Q<q.length){let U=q[Q],L=U.trim();if(/^[^\n]+\s@\s[^\n]+:$/.test(L))break;if(L.startsWith("Channel:")||L==="Messages:")break;z.push(U.startsWith("  ")?U.slice(2):U),Q+=1}if(z.length>0)B.push(z.join(`
`).trim());continue}return $}return J&&B.length>0?B.filter(Boolean).join(`

`):$}function Bj($){if(!$)return{content:$,fileRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let U=0;U<q.length;U+=1)if(q[U].trim()==="Files:"&&q[U+1]&&/^\s*-\s+/.test(q[U+1])){B=U;break}if(B===-1)return{content:$,fileRefs:[]};let Q=[],J=B+1;for(;J<q.length;J+=1){let U=q[J];if(/^\s*-\s+/.test(U))Q.push(U.replace(/^\s*-\s+/,"").trim());else if(!U.trim())break;else break}if(Q.length===0)return{content:$,fileRefs:[]};let K=q.slice(0,B),X=q.slice(J);return{content:[...K,...X].join(`
`).replace(/\n{3,}/g,`

`).trim(),fileRefs:Q}}function Qj($){if(!$)return{content:$,messageRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let U=0;U<q.length;U+=1)if(q[U].trim()==="Referenced messages:"&&q[U+1]&&/^\s*-\s+/.test(q[U+1])){B=U;break}if(B===-1)return{content:$,messageRefs:[]};let Q=[],J=B+1;for(;J<q.length;J+=1){let U=q[J];if(/^\s*-\s+/.test(U)){let L=U.replace(/^\s*-\s+/,"").trim().match(/^message:(\S+)$/i);if(L)Q.push(L[1])}else if(!U.trim())break;else break}if(Q.length===0)return{content:$,messageRefs:[]};let K=q.slice(0,B),X=q.slice(J);return{content:[...K,...X].join(`
`).replace(/\n{3,}/g,`

`).trim(),messageRefs:Q}}function Jj($){if(!$)return{content:$,attachmentRefs:[]};let q=$.replace(/\r\n/g,`
`).replace(/\r/g,`
`).split(`
`),B=-1;for(let U=0;U<q.length;U+=1)if(q[U].trim()==="Attachments:"&&q[U+1]&&/^\s*-\s+/.test(q[U+1])){B=U;break}if(B===-1)return{content:$,attachmentRefs:[]};let Q=[],J=B+1;for(;J<q.length;J+=1){let U=q[J];if(/^\s*-\s+/.test(U)){let L=U.replace(/^\s*-\s+/,"").trim(),N=L.match(/^attachment:(\d+)(?:\s*\((.+)\))?$/i);if(N)Q.push({id:N[1],label:(N[2]||"").trim()||`attachment:${N[1]}`,raw:L})}else if(!U.trim())break;else break}if(Q.length===0)return{content:$,attachmentRefs:[]};let K=q.slice(0,B),X=q.slice(J);return{content:[...K,...X].join(`
`).replace(/\n{3,}/g,`

`).trim(),attachmentRefs:Q}}function Kj($){let j=qj($||""),q=Bj(j||""),B=Qj(q.content||""),Q=Jj(B.content||"");return{text:Q.content||"",fileRefs:q.fileRefs,messageRefs:B.messageRefs,attachmentRefs:Q.attachmentRefs}}function Xj({items:$=[],onInjectQueuedFollowup:j,onRemoveQueuedFollowup:q,onMoveQueuedFollowup:B,onOpenFilePill:Q}){if(!Array.isArray($)||$.length===0)return null;return G`
        <div class="compose-queue-stack">
            ${$.map((J,K)=>{let X=typeof J?.content==="string"?J.content:"",z=Kj(X);if(!z.text.trim()&&z.fileRefs.length===0&&z.messageRefs.length===0&&z.attachmentRefs.length===0)return null;let U=K>0,L=K<$.length-1;return G`
                    <div class="compose-queue-stack-item" role="listitem">
                        <div class="compose-queue-stack-content" title=${X}>
                            ${z.text.trim()&&G`<div class="compose-queue-stack-text">${z.text}</div>`}
                            ${(z.messageRefs.length>0||z.fileRefs.length>0||z.attachmentRefs.length>0)&&G`
                                <div class="compose-queue-stack-refs">
                                    ${z.messageRefs.map((N)=>G`
                                        <${P0}
                                            key=${"queue-msg-"+N}
                                            prefix="compose"
                                            label=${"msg:"+N}
                                            title=${"Message reference: "+N}
                                            icon="message"
                                        />
                                    `)}
                                    ${z.fileRefs.map((N)=>{let V=N.split("/").pop()||N;return G`
                                            <${P0}
                                                key=${"queue-file-"+N}
                                                prefix="compose"
                                                label=${V}
                                                title=${N}
                                                onClick=${()=>Q?.(N)}
                                            />
                                        `})}
                                    ${z.attachmentRefs.map((N)=>G`
                                        <${P0}
                                            key=${"queue-attachment-"+N.id}
                                            prefix="compose"
                                            label=${N.label}
                                            title=${N.raw}
                                        />
                                    `)}
                                </div>
                            `}
                        </div>
                        <div class="compose-queue-stack-actions" role="group" aria-label="Queued follow-up controls">
                            ${$.length>1&&G`
                                <button
                                    class="compose-queue-stack-move-btn"
                                    type="button"
                                    title="Move up"
                                    aria-label="Move up in queue"
                                    disabled=${!U}
                                    onClick=${()=>U&&B?.(K,K-1)}
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
                                    disabled=${!L}
                                    onClick=${()=>L&&B?.(K,K+1)}
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
                                onClick=${()=>j?.(J)}
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
                                onClick=${()=>q?.(J)}
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
    `}function I8({onPost:$,onFocus:j,searchMode:q,searchScope:B="current",onSearch:Q,onSearchScopeChange:J,onEnterSearch:K,onExitSearch:X,fileRefs:z=[],onRemoveFileRef:U,onClearFileRefs:L,messageRefs:N=[],onRemoveMessageRef:V,onClearMessageRefs:O,activeModel:g=null,agentModelsPayload:h=null,modelUsage:p=null,thinkingLevel:d=null,supportsThinking:P=!1,contextUsage:t=null,onContextCompact:v,notificationsEnabled:e=!1,notificationPermission:W1="default",onToggleNotifications:D1,onModelChange:N1,onModelStateChange:X1,activeEditorPath:x=null,onAttachEditorFile:U1,onOpenFilePill:E1,followupQueueItems:b1=[],onInjectQueuedFollowup:P1,onRemoveQueuedFollowup:l,onMoveQueuedFollowup:s,onSubmitIntercept:Q1,onMessageResponse:_1,onPopOutChat:D,isAgentActive:r=!1,activeChatAgents:L1=[],currentChatJid:j1="web:default",connectionStatus:M1="connected",onSetFileRefs:R1,onSetMessageRefs:Y1,onSubmitError:p1,onSwitchChat:C1,onRenameSession:j0,isRenameSessionInProgress:s1=!1,onCreateSession:O0,onDeleteSession:K0,onRestoreSession:f0,showQueueStack:G0=!0,statusNotice:o1=null,extensionWorkingState:q0=null,prefillRequest:H0=null}){let[y1,X0]=M(""),[N0,a1]=M(""),[t1,a]=M([]),[z1,J1]=M(!1),[A1,k1]=M([]),[c1,f1]=M(0),[Y0,e1]=M(!1),l0=T(null),[B0,m1]=M([]),[$0,Z0]=M(0),[A0,h1]=M(!1),[Q0,D0]=M(!1),[C,b]=M(!1),[E,i]=M(!1),[n,V1]=M([]),[Z1,G1]=M(0),[F1,Y]=M(0),[S,$1]=M(!1),[K1,c]=M(0),[O1,w1]=M(null),[T1,S1]=M(null),[U0,I1]=M(()=>Date.now()),[S0,T0]=M(0),g1=T(null),m2=T(null),h2=T(null),w0=T(null),B4=T(null),p2=T(null),F0=T(null),u0=T(null),L0=T({value:"",updatedAt:0}),J0=T(0),V0=T(!1),r0=200,x0=i$(j1),U2=(_)=>{let F=new Set,I=[];for(let f of _||[]){if(typeof f!=="string")continue;let B1=f.trim();if(!B1||F.has(B1))continue;F.add(B1),I.push(B1)}return I},_2=(_=x0)=>{let F=k0(_);if(!F)return[];try{let I=JSON.parse(F);if(!Array.isArray(I))return[];return U2(I)}catch{return[]}},Q4=(_,F=x0)=>{E0(F,JSON.stringify(_))},R0=T(_2(x0)),g0=T(-1),z2=T(""),c2=T("");w(()=>{R0.current=_2(x0),g0.current=-1,z2.current=""},[x0]),w(()=>{let _=!1;return fetch(`/agent/commands?chat_jid=${encodeURIComponent(j1||"web:default")}`).then((I)=>I.ok?I.json():null).then((I)=>{if(_||!I?.commands)return;l0.current=I.commands.map((f)=>({name:f.name,description:f.description||""}))}).catch((I)=>{console.debug("[compose] failed to fetch dynamic commands",I)}),()=>{_=!0}},[j1]),w(()=>{let _=r$(H0,c2.current,q);if(!_.shouldApply)return;c2.current=_.nextToken,w1(null),X0(_.text),L4(_.text),r2(_.text),requestAnimationFrame(()=>{c0();let F=g1.current;if(!F)return;try{F.focus({preventScroll:!0})}catch{F.focus()}let I=_.text.length;F.setSelectionRange?.(I,I)})},[H0,q]);let k2=y1.trim()||t1.length>0||z.length>0||N.length>0,J4=typeof window<"u"&&typeof navigator<"u"&&Boolean(navigator.geolocation)&&Boolean(window.isSecureContext),i0=typeof window<"u"&&typeof Notification<"u",K4=typeof window<"u"?Boolean(window.isSecureContext):!1,X5=i0&&K4&&W1!=="denied",X4=W1==="granted"&&e,W2=g2(o1),d1=q5(o1),m0=typeof o1?.detail==="string"&&o1.detail.trim()?o1.detail.trim():"",U4=W2?B5(o1,U0):null,n0=o$(q0,S0),b0=q0?.indicator&&typeof q0.indicator==="object"?q0.indicator:null,U5=X4?"Disable notifications":"Enable notifications",_5=t1.length>0||z.length>0||N.length>0,$2=z3(M1),z5=$2.label,W5=$2.title,h0=n$(r,k2,W2),_4=(Array.isArray(L1)?L1:[]).filter((_)=>!_?.archived_at),v1=(()=>{for(let _ of Array.isArray(L1)?L1:[]){let F=typeof _?.chat_jid==="string"?_.chat_jid.trim():"";if(F&&F===j1)return _}return null})(),l1=Boolean(v1&&v1.chat_jid===(v1.root_chat_jid||v1.chat_jid)),E2=H1(()=>{let _=new Set,F=[];for(let I of Array.isArray(L1)?L1:[]){let f=typeof I?.chat_jid==="string"?I.chat_jid.trim():"";if(!f||f===j1||_.has(f))continue;if(!(typeof I?.agent_name==="string"?I.agent_name.trim():""))continue;_.add(f),F.push(I)}return F},[L1,j1]),p0=E2.length>0,j2=p0&&typeof C1==="function",q2=p0&&typeof f0==="function",P2=Boolean(s1||V0.current),s0=!q&&typeof j0==="function"&&!P2,v0=!q&&typeof O0==="function",y0=!q&&typeof K0==="function"&&!l1,B2=!q&&(j2||q2||s0||v0||y0),z4=jj(g,h),G2=z4.showPicker,N2=z4.label,d2=P&&d?` (${d})`:"",G5=d2.trim()?`${d}`:"",l2=typeof p?.hint_short==="string"?p.hint_short.trim():"",W4=[G5||null,l2||null].filter(Boolean).join(" • "),N5=[g?`Current model: ${N2}${d2}`:null,p?.plan?`Plan: ${p.plan}`:null,l2||null,p?.primary?.reset_description||null,p?.secondary?.reset_description||null].filter(Boolean),G4=Q0?"Switching model…":N5.join(" • ")||(G2?"Select a model (tap to open model picker)":`Current model: ${N2}${d2} (tap to open model picker)`),N4=!q&&(G2||t&&t.percent!=null),o0=(_)=>{if(!_||typeof _!=="object")return;let F=_.model??_.current;if(typeof X1==="function")X1({model:F??null,thinking_level:_.thinking_level??null,thinking_level_label:_.thinking_level_label??null,supports_thinking:_.supports_thinking,provider_usage:_.provider_usage??null});if(F&&typeof N1==="function")N1(F)},c0=(_)=>{let F=_||g1.current;if(!F)return;F.style.height="auto",F.style.height=`${F.scrollHeight}px`,F.style.overflowY="hidden"},L4=(_)=>{if(!_.startsWith("/")||_.includes(`
`)){e1(!1),k1([]);return}let F=_.toLowerCase().split(" ")[0];if(F.length<1){e1(!1),k1([]);return}let f=(l0.current||l$).filter((B1)=>B1.name.startsWith(F)||B1.name.replace(/-/g,"").startsWith(F.replace(/-/g,"")));if(f.length>0&&!(f.length===1&&f[0].name===F))h1(!1),m1([]),k1(f),f1(0),e1(!0);else e1(!1),k1([])},V4=(_)=>{let F=y1,I=F.indexOf(" "),f=I>=0?F.slice(I):"",B1=_.name+f;X0(B1),e1(!1),k1([]),requestAnimationFrame(()=>{let r1=g1.current;if(!r1)return;let W0=B1.length;r1.selectionStart=W0,r1.selectionEnd=W0,r1.focus()})},r2=(_)=>{if(K3(_)==null){h1(!1),m1([]);return}let F=W8(_4,_,{currentChatJid:j1});if(F.length>0&&!(F.length===1&&X3(F[0].agent_name).trim().toLowerCase()===String(_||"").trim().toLowerCase()))e1(!1),k1([]),m1(F),Z0(0),h1(!0);else h1(!1),m1([])},L2=(_)=>{let F=X3(_?.agent_name);if(!F)return;X0(F),h1(!1),m1([]),requestAnimationFrame(()=>{let I=g1.current;if(!I)return;let f=F.length;I.selectionStart=f,I.selectionEnd=f,I.focus()})},Z4=()=>{if(q||!j2&&!q2&&!s0&&!v0&&!y0)return!1;return L0.current={value:"",updatedAt:0},b(!1),e1(!1),k1([]),h1(!1),m1([]),i(!0),!0},F4=(_)=>{if(_?.preventDefault?.(),_?.stopPropagation?.(),q||!j2&&!q2&&!s0&&!v0&&!y0)return;if(E){L0.current={value:"",updatedAt:0},i(!1);return}Z4()},H4=(_)=>{let F=typeof _==="string"?_.trim():"";if(i(!1),!F||F===j1){requestAnimationFrame(()=>g1.current?.focus());return}C1?.(F)},f2=async(_)=>{let F=typeof _==="string"?_.trim():"";if(i(!1),!F||typeof f0!=="function"){requestAnimationFrame(()=>g1.current?.focus());return}try{await f0(F)}catch(I){console.warn("Failed to restore session:",I),requestAnimationFrame(()=>g1.current?.focus())}},L5=(_)=>{let I=(Array.isArray(_)?_:[]).findIndex((f)=>!f?.disabled);return I>=0?I:0},_0=H1(()=>{let _=[];for(let F of E2){let I=Boolean(F?.archived_at),f=typeof F?.agent_name==="string"?F.agent_name.trim():"",B1=typeof F?.chat_jid==="string"?F.chat_jid.trim():"";if(!f||!B1)continue;_.push({type:"session",key:`session:${B1}`,label:`@${f} — ${B1}${F?.is_active?" active":""}${I?" archived":""}`,chat:F,disabled:I?!q2:!j2})}if(v0)_.push({type:"action",key:"action:new",label:"New session",action:"new",disabled:!1});if(s0)_.push({type:"action",key:"action:rename",label:"Rename current session",action:"rename",disabled:P2});if(y0)_.push({type:"action",key:"action:delete",label:"Delete current session",action:"delete",disabled:!1});return _},[E2,q2,j2,v0,s0,y0,P2]),Y4=async(_)=>{if(_?.preventDefault)_.preventDefault();if(_?.stopPropagation)_.stopPropagation();if(typeof j0!=="function"||s1||V0.current)return;V0.current=!0,i(!1);try{await j0()}catch(F){console.warn("Failed to rename session:",F)}finally{V0.current=!1}requestAnimationFrame(()=>g1.current?.focus())},A4=async()=>{if(typeof O0!=="function")return;i(!1);try{await O0()}catch(_){console.warn("Failed to create session:",_)}requestAnimationFrame(()=>g1.current?.focus())},D4=async()=>{if(typeof K0!=="function")return;i(!1);try{await K0(j1)}catch(_){console.warn("Failed to delete session:",_)}requestAnimationFrame(()=>g1.current?.focus())},V2=(_)=>{if(q)a1(_);else X0(_),L4(_),r2(_);requestAnimationFrame(()=>c0())},V5=(_)=>{let F=q?N0:y1,I=F&&!F.endsWith(`
`)?`
`:"",f=`${F}${I}${_}`.trimStart();V2(f)},Z5=(_)=>{let F=_?.command?.model_label;if(F)return F;let I=_?.command?.message;if(typeof I==="string"){let f=I.match(/•\s+([^\n]+?)\s+\(current\)/);if(f?.[1])return f[1].trim()}return null},C4=async(_)=>{if(q||Q0)return;w1(null),S1(null),D0(!0);try{let F=await u2("default",_,null,[],null,j1),I=Z5(F);return o0({model:I??g??null,thinking_level:F?.command?.thinking_level,thinking_level_label:F?.command?.thinking_level_label,supports_thinking:F?.command?.supports_thinking}),await W3(p4,j1,o0),S1(D8(_,F)),$?.(F),!0}catch(F){return console.error("Failed to switch model:",F),alert("Failed to switch model: "+F.message),!1}finally{D0(!1)}},F5=async()=>{await C4("/cycle-model")},i2=async(_)=>{let F=typeof _==="string"?_:typeof _?.label==="string"?_.label:"";if(!F||Q0)return;if(await C4(`/model ${F}`))b(!1)},H5=(_)=>{if(!_||_.disabled)return;if(_.type==="session"){let F=_.chat;if(F?.archived_at)f2(F.chat_jid);else H4(F.chat_jid);return}if(_.type==="action"){if(_.action==="new"){A4();return}if(_.action==="rename"){Y4();return}if(_.action==="delete")D4()}},Y5=(_)=>{_.preventDefault(),_.stopPropagation(),L0.current={value:"",updatedAt:0},i(!1),b((F)=>!F)},A5=async()=>{if(q)return;v?.(),await W("/compact",null,{includeMedia:!1,includeFileRefs:!1,includeMessageRefs:!1,clearAfterSubmit:!1,recordHistory:!1})},D5=(_)=>{if(_==="queue"||_==="steer"||_==="auto")return _;return r?"queue":void 0},W=async(_,F,I={})=>{let{includeMedia:f=!0,includeFileRefs:B1=!0,includeMessageRefs:r1=!0,clearAfterSubmit:W0=!0,recordHistory:I0=!0}=I||{},F2=typeof _==="string"?_:_&&typeof _?.target?.value==="string"?_.target.value:y1,n2=typeof F2==="string"?F2:"";if(!n2.trim()&&(f?t1.length===0:!0)&&(B1?z.length===0:!0)&&(r1?N.length===0:!0))return;e1(!1),k1([]),h1(!1),m1([]),i(!1),w1(null),S1(null);let I4=f?[...t1]:[],O4=B1?[...z]:[],T4=r1?[...N]:[],H2=n2.trim();if(I0&&H2){let Y2=R0.current,M0=U2(Y2.filter((C5)=>C5!==H2));if(M0.push(H2),M0.length>200)M0.splice(0,M0.length-200);R0.current=M0,Q4(M0),g0.current=-1,z2.current=""}let J9=()=>{if(f)a([...I4]);if(B1)R1?.(O4);if(r1)Y1?.(T4);X0(H2),requestAnimationFrame(()=>c0())};if(W0)X0(""),a([]),L?.(),O?.();(async()=>{try{let Y2=await Q1?.({content:H2,submitMode:F,fileRefs:O4,messageRefs:T4,mediaFiles:I4});if(Y2){$?.(Y2);return}let M0=[];for(let A2 of I4){let M4=await V6(A2);M0.push(M4.id)}let C5=O4.length?`Files:
${O4.map((A2)=>`- ${A2}`).join(`
`)}`:"",K9=T4.length?`Referenced messages:
${T4.map((A2)=>`- message:${A2}`).join(`
`)}`:"",X9=M0.length?`Attachments:
${M0.map((A2,M4)=>{let _9=I4[M4]?.name||`attachment-${M4+1}`;return`- attachment:${A2} (${_9})`}).join(`
`)}`:"",U9=[H2,C5,K9,X9].filter(Boolean).join(`

`),Q2=await u2("default",U9,null,M0,D5(F),j1);if(_1?.(Q2),Q2?.command)o0({model:Q2.command.model_label??g??null,thinking_level:Q2.command.thinking_level,thinking_level_label:Q2.command.thinking_level_label,supports_thinking:Q2.command.supports_thinking}),await W3(p4,j1,o0);S1(D8(H2,Q2)),$?.(Q2)}catch(Y2){if(W0)J9();let M0=Y2?.message||"Failed to send message.";w1(M0),p1?.(M0),console.error("Failed to post:",Y2)}})()},Z=(_)=>{P1?.(_)},H=m((_)=>{if(q||!C&&!E||_?.isComposing)return!1;let F=()=>{_.preventDefault?.(),_.stopPropagation?.()},I=()=>{L0.current={value:"",updatedAt:0}};if(_.key==="Escape"){if(F(),I(),C)b(!1);if(E)i(!1);return!0}if(C){if(_.key==="ArrowDown"){if(F(),I(),n.length>0)G1((f)=>(f+1)%n.length);return!0}if(_.key==="ArrowUp"){if(F(),I(),n.length>0)G1((f)=>(f-1+n.length)%n.length);return!0}if((_.key==="Enter"||_.key==="Tab")&&n.length>0)return F(),I(),i2(n[Math.max(0,Math.min(Z1,n.length-1))]),!0;if(B3(_)&&n.length>0){F();let f=Q3(L0.current,_.key);L0.current=f;let B1=J3(n,f.value,Z1,(r1)=>$j(r1));if(B1>=0)G1(B1);return!0}}if(E){if(_.key==="ArrowDown"){if(F(),I(),_0.length>0)Y((f)=>(f+1)%_0.length);return!0}if(_.key==="ArrowUp"){if(F(),I(),_0.length>0)Y((f)=>(f-1+_0.length)%_0.length);return!0}if((_.key==="Enter"||_.key==="Tab")&&_0.length>0)return F(),I(),H5(_0[Math.max(0,Math.min(F1,_0.length-1))]),!0;if(B3(_)&&_0.length>0){F();let f=Q3(L0.current,_.key);L0.current=f;let B1=J3(_0,f.value,F1,(r1)=>r1.label);if(B1>=0)Y(B1);return!0}}return!1},[q,C,E,n,Z1,_0,F1,i2]),A=(_)=>{if(_.isComposing)return;if(q&&_.key==="Escape"){_.preventDefault(),a1(""),X?.();return}if(H(_))return;let F=g1.current?.value??(q?N0:y1);if(G8(_,F,{searchMode:q,showSessionSwitcherButton:B2})){_.preventDefault(),Z4();return}if(A0&&B0.length>0){let I=g1.current?.value??(q?N0:y1);if(!String(I||"").match(/^@([a-zA-Z0-9_-]*)$/))h1(!1),m1([]);else{if(_.key==="ArrowDown"){_.preventDefault(),Z0((f)=>(f+1)%B0.length);return}if(_.key==="ArrowUp"){_.preventDefault(),Z0((f)=>(f-1+B0.length)%B0.length);return}if(_.key==="Tab"||_.key==="Enter"){_.preventDefault(),L2(B0[$0]);return}if(_.key==="Escape"){_.preventDefault(),h1(!1),m1([]);return}}}if(Y0&&A1.length>0){let I=g1.current?.value??(q?N0:y1);if(!String(I||"").startsWith("/"))e1(!1),k1([]);else{if(_.key==="ArrowDown"){_.preventDefault(),f1((f)=>(f+1)%A1.length);return}if(_.key==="ArrowUp"){_.preventDefault(),f1((f)=>(f-1+A1.length)%A1.length);return}if(_.key==="Tab"){_.preventDefault(),V4(A1[c1]);return}if(_.key==="Enter"&&!_.shiftKey){if(!F.includes(" ")){_.preventDefault();let B1=A1[c1];e1(!1),k1([]),W(B1.name);return}}if(_.key==="Escape"){_.preventDefault(),e1(!1),k1([]);return}}}if(!q&&(_.key==="ArrowUp"||_.key==="ArrowDown")&&!_.metaKey&&!_.ctrlKey&&!_.altKey&&!_.shiftKey){let I=g1.current;if(!I)return;let f=I.value||"",B1=I.selectionStart===0&&I.selectionEnd===0,r1=I.selectionStart===f.length&&I.selectionEnd===f.length;if(_.key==="ArrowUp"&&B1||_.key==="ArrowDown"&&r1){let W0=R0.current;if(!W0.length)return;_.preventDefault();let I0=g0.current;if(_.key==="ArrowUp"){if(I0===-1)z2.current=f,I0=W0.length-1;else if(I0>0)I0-=1;g0.current=I0,V2(W0[I0]||"")}else{if(I0===-1)return;if(I0<W0.length-1)I0+=1,g0.current=I0,V2(W0[I0]||"");else g0.current=-1,V2(z2.current||""),z2.current=""}requestAnimationFrame(()=>{let F2=g1.current;if(!F2)return;let n2=F2.value.length;F2.selectionStart=n2,F2.selectionEnd=n2});return}}if(_.key==="Enter"&&!_.shiftKey&&(_.ctrlKey||_.metaKey)){if(_.preventDefault(),q){if(F.trim())Q?.(F.trim(),B)}else W(F,"steer");return}if(_.key==="Enter"&&!_.shiftKey)if(_.preventDefault(),q){if(F.trim())Q?.(F.trim(),B)}else W(F)},k=(_)=>{let F=Array.from(_||[]).filter((I)=>I instanceof File&&!String(I.name||"").startsWith(".DS_Store"));if(!F.length)return;a((I)=>[...I,...F]),w1(null)},y=(_)=>{k(_.target.files),_.target.value=""},u=(_)=>{if(q)return;_.preventDefault(),_.stopPropagation(),J0.current+=1,J1(!0)},o=(_)=>{if(q)return;if(_.preventDefault(),_.stopPropagation(),J0.current=Math.max(0,J0.current-1),J0.current===0)J1(!1)},R=(_)=>{if(q)return;if(_.preventDefault(),_.stopPropagation(),_.dataTransfer)_.dataTransfer.dropEffect="copy";J1(!0)},q1=(_)=>{if(q)return;_.preventDefault(),_.stopPropagation(),J0.current=0,J1(!1),k(_.dataTransfer?.files||[])},z0=(_)=>{if(q)return;let F=_.clipboardData?.items;if(!F||!F.length)return;let I=[];for(let f of F){if(f.kind!=="file")continue;let B1=f.getAsFile?.();if(B1)I.push(B1)}if(I.length>0)_.preventDefault(),k(I)},C0=(_)=>{a((F)=>F.filter((I,f)=>f!==_))},Z2=()=>{w1(null),a([]),L?.(),O?.()},O3=()=>{if(!navigator.geolocation){alert("Geolocation is not available in this browser.");return}navigator.geolocation.getCurrentPosition((_)=>{let{latitude:F,longitude:I,accuracy:f}=_.coords,B1=`${F.toFixed(5)}, ${I.toFixed(5)}`,r1=Number.isFinite(f)?` ±${Math.round(f)}m`:"",W0=`https://maps.google.com/?q=${F},${I}`,I0=`Location: ${B1}${r1} ${W0}`;V5(I0)},(_)=>{let F=_?.message||"Unable to retrieve location.";alert(`Location error: ${F}`)},{enableHighAccuracy:!0,timeout:1e4,maximumAge:0})};w(()=>{if(!C)return;L0.current={value:"",updatedAt:0},$1(!0),p4(j1).then((_)=>{V1(C8(_)),o0(_)}).catch((_)=>{console.warn("Failed to load model list:",_),V1([])}).finally(()=>{$1(!1)})},[C,g]),w(()=>{if(q)b(!1),i(!1),e1(!1),k1([]),h1(!1),m1([])},[q]),w(()=>{if(E&&!B2)i(!1)},[E,B2]),w(()=>{if(!C)return;let _=n.findIndex((F)=>F?.label===g);G1(_>=0?_:0)},[C,n,g]),w(()=>{if(!E)return;Y(L5(_0)),L0.current={value:"",updatedAt:0}},[E,j1]),w(()=>{if(!C)return;let _=(F)=>{let I=w0.current,f=B4.current,B1=F.target;if(I&&I.contains(B1))return;if(f&&f.contains(B1))return;b(!1)};return document.addEventListener("pointerdown",_),()=>document.removeEventListener("pointerdown",_)},[C]),w(()=>{if(!E)return;let _=(F)=>{let I=p2.current,f=F0.current,B1=F.target;if(I&&I.contains(B1))return;if(f&&f.contains(B1))return;i(!1)};return document.addEventListener("pointerdown",_),()=>document.removeEventListener("pointerdown",_)},[E]),w(()=>{if(q||!C&&!E)return;let _=(F)=>{H(F)};return document.addEventListener("keydown",_,!0),()=>document.removeEventListener("keydown",_,!0)},[q,C,E,H]),w(()=>{if(!C)return;let _=w0.current;_?.focus?.(),_?.querySelector?.(".compose-model-popup-item.active")?.scrollIntoView?.({block:"nearest"})},[C,Z1,n]),w(()=>{if(!E)return;let _=p2.current;_?.focus?.(),_?.querySelector?.(".compose-model-popup-item.active")?.scrollIntoView?.({block:"nearest"})},[E,F1,_0.length]),w(()=>{if(!A0||!h2.current)return;h2.current.querySelector?.(".slash-item.active")?.scrollIntoView?.({block:"nearest"})},[A0,$0,B0.length]),w(()=>{if(!Y0||!m2.current)return;m2.current.querySelector?.(".slash-item.active")?.scrollIntoView?.({block:"nearest"})},[Y0,c1,A1.length]),w(()=>{let _=()=>{let r1=u0.current?.clientWidth||0;c((W0)=>W0===r1?W0:r1)};_();let F=u0.current,I=0,f=()=>{if(I)cancelAnimationFrame(I);I=requestAnimationFrame(()=>{I=0,_()})},B1=null;if(F&&typeof ResizeObserver<"u")B1=new ResizeObserver(()=>f()),B1.observe(F);if(typeof window<"u")window.addEventListener("resize",f);return()=>{if(I)cancelAnimationFrame(I);if(B1?.disconnect?.(),typeof window<"u")window.removeEventListener("resize",f)}},[q,g,v1?.agent_name,B2,t?.percent]);let Q9=(_)=>{let F=_.target.value;if(w1(null),S1(null),E)i(!1);c0(_.target),V2(F)};return w(()=>{requestAnimationFrame(()=>c0())},[y1,N0,q]),w(()=>{if(!W2)return;I1(Date.now());let _=setInterval(()=>I1(Date.now()),1000);return()=>clearInterval(_)},[W2,o1?.started_at,o1?.startedAt]),w(()=>{if(T0(0),b0?.mode!=="custom"||!Array.isArray(b0.frames)||b0.frames.length<=1)return;let _=typeof b0.intervalMs==="number"&&Number.isFinite(b0.intervalMs)&&b0.intervalMs>0?b0.intervalMs:120,F=setInterval(()=>{T0((I)=>(I+1)%b0.frames.length)},_);return()=>clearInterval(F)},[b0]),w(()=>{if(q)return;r2(y1)},[_4,j1,y1,q]),G`
        <div class="compose-box">
            ${G0&&!q&&G`
                <${Xj}
                    items=${b1}
                    onInjectQueuedFollowup=${Z}
                    onRemoveQueuedFollowup=${l}
                    onMoveQueuedFollowup=${s}
                    onOpenFilePill=${E1}
                />
            `}
            ${n0.visible&&G`
                <div class="compose-inline-status extension-working" role="status" aria-live="polite">
                    <div class="compose-inline-status-row">
                        ${n0.indicatorText?G`<span class="compose-inline-status-glyph" aria-hidden="true">${n0.indicatorText}</span>`:n0.animateDot?G`<span class=${U3({pulsing:!0})} aria-hidden="true"></span>`:null}
                        <span class="compose-inline-status-title">${n0.title}</span>
                    </div>
                </div>
            `}
            ${o1&&G`
                <div
                    class=${`compose-inline-status${W2?" compaction":""}`}
                    role="status"
                    aria-live="polite"
                    title=${m0||""}
                >
                    <div class="compose-inline-status-row">
                        <span class=${U3({pulsing:W2})} aria-hidden="true"></span>
                        <span class="compose-inline-status-title">${d1}</span>
                        ${U4&&G`<span class="compose-inline-status-elapsed">${U4}</span>`}
                    </div>
                    ${m0&&G`<div class="compose-inline-status-detail">${m0}</div>`}
                </div>
            `}
            ${T1&&G`
                <div class="compose-inline-status compose-command-notice" role="status" aria-live="polite">
                    <div class="compose-inline-status-detail compose-command-notice-text">${T1}</div>
                </div>
            `}
            <div
                class=${`compose-input-wrapper${z1?" drag-active":""}`}
                onDragEnter=${u}
                onDragOver=${R}
                onDragLeave=${o}
                onDrop=${q1}
            >
                <div class="compose-input-main">
                    ${_5&&G`
                        <div class="compose-file-refs">
                            ${N.map((_)=>{return G`
                                    <${P0}
                                        key=${"msg-"+_}
                                        prefix="compose"
                                        label=${"msg:"+_}
                                        title=${"Message reference: "+_}
                                        removeTitle="Remove reference"
                                        icon="message"
                                        onRemove=${()=>V?.(_)}
                                    />
                                `})}
                            ${z.map((_)=>{let F=_.split("/").pop()||_;return G`
                                    <${P0}
                                        prefix="compose"
                                        label=${F}
                                        title=${_}
                                        onClick=${()=>E1?.(_)}
                                        removeTitle="Remove file"
                                        onRemove=${()=>U?.(_)}
                                    />
                                `})}
                            ${t1.map((_,F)=>{let I=_?.name||`attachment-${F+1}`;return G`
                                    <${P0}
                                        key=${I+F}
                                        prefix="compose"
                                        label=${I}
                                        title=${I}
                                        removeTitle="Remove attachment"
                                        onRemove=${()=>C0(F)}
                                    />
                                `})}
                            <button
                                type="button"
                                class="compose-clear-attachments-btn"
                                onClick=${Z2}
                                title="Clear all attachments and references"
                                aria-label="Clear all attachments and references"
                            >
                                Clear all
                            </button>
                        </div>
                    `}
                    ${!q&&typeof D==="function"&&G`
                        <button
                            type="button"
                            class="compose-popout-btn"
                            onClick=${()=>D?.()}
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
                        ref=${g1}
                        placeholder=${q?"Search (Enter to run)...":"Message (Enter to send, Shift+Enter for newline)..."}
                        value=${q?N0:y1}
                        onInput=${Q9}
                        onKeyDown=${A}
                        onPaste=${z0}
                        onFocus=${j}
                        onClick=${j}
                        rows="1"
                    />
                    ${A0&&B0.length>0&&G`
                        <div class="slash-autocomplete" ref=${h2}>
                            ${B0.map((_,F)=>G`
                                <div
                                    key=${_.chat_jid||_.agent_name}
                                    class=${`slash-item${F===$0?" active":""}`}
                                    onMouseDown=${(I)=>{I.preventDefault(),L2(_)}}
                                    onMouseEnter=${()=>Z0(F)}
                                >
                                    <span class="slash-name">@${_.agent_name}</span>
                                    <span class="slash-desc">${_.chat_jid||"Active agent"}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${Y0&&A1.length>0&&G`
                        <div class="slash-autocomplete" ref=${m2}>
                            ${A1.map((_,F)=>G`
                                <div
                                    key=${_.name}
                                    class=${`slash-item${F===c1?" active":""}`}
                                    onMouseDown=${(I)=>{I.preventDefault(),V4(_)}}
                                    onMouseEnter=${()=>f1(F)}
                                >
                                    <span class="slash-name">${_.name}</span>
                                    <span class="slash-desc">${_.description}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${C&&!q&&G`
                        <div class="compose-model-popup" ref=${w0} tabIndex="-1" onKeyDown=${H}>
                            <div class="compose-model-popup-title">Select model</div>
                            <div class="compose-model-popup-menu" role="menu" aria-label="Model picker">
                                ${S&&G`
                                    <div class="compose-model-popup-empty">Loading models…</div>
                                `}
                                ${!S&&n.length===0&&G`
                                    <div class="compose-model-popup-empty">No models available.</div>
                                `}
                                ${!S&&n.map((_,F)=>{let I=typeof _?.label==="string"?_.label:"",f=N3(_?.contextWindow);return G`
                                        <button
                                            key=${I}
                                            type="button"
                                            role="menuitem"
                                            class=${`compose-model-popup-item compose-model-popup-model-item${Z1===F?" active":""}${g===I?" current-model":""}`}
                                            onClick=${()=>{i2(_)}}
                                            disabled=${Q0}
                                            title=${[I,f].filter(Boolean).join(" • ")}
                                        >
                                            <span class="compose-model-popup-model-label">${t$(I,_?.contextWindow)}</span>
                                        </button>
                                    `})}
                            </div>
                            <div class="compose-model-popup-actions">
                                <button
                                    type="button"
                                    class="compose-model-popup-btn"
                                    onClick=${()=>{F5()}}
                                    disabled=${Q0}
                                >
                                    Next model
                                </button>
                            </div>
                        </div>
                    `}
                    ${E&&!q&&G`
                        <div class="compose-model-popup" ref=${p2} tabIndex="-1" onKeyDown=${H}>
                            <div class="compose-model-popup-title">Manage sessions & agents</div>
                            <div class="compose-model-popup-menu" role="menu" aria-label="Sessions and agents">
                                ${G`
                                    <div class="compose-model-popup-item current" role="note" aria-live="polite">
                                        ${(()=>{return L8(v1,j1)})()}
                                    </div>
                                `}
                                ${!p0&&G`
                                    <div class="compose-model-popup-empty">No other sessions yet.</div>
                                `}
                                ${p0&&E2.map((_,F)=>{let I=Boolean(_.archived_at),B1=_.chat_jid!==(_.root_chat_jid||_.chat_jid)&&!_.is_active&&!I&&typeof K0==="function",r1=V8(_,{currentChatJid:j1});return G`
                                        <div key=${_.chat_jid} class=${`compose-model-popup-item-row${I?" archived":""}`}>
                                            <button
                                                type="button"
                                                role="menuitem"
                                                class=${`compose-model-popup-item${I?" archived":""}${F1===F?" active":""}`}
                                                onClick=${()=>{if(I){f2(_.chat_jid);return}H4(_.chat_jid)}}
                                                disabled=${I?!q2:!j2}
                                                title=${I?`Restore archived ${`@${_.agent_name}`}`:`Switch to ${`@${_.agent_name}`}`}
                                            >
                                                ${r1}
                                            </button>
                                            ${B1&&G`
                                                <button
                                                    type="button"
                                                    class="compose-model-popup-item-delete"
                                                    title="Delete this branch"
                                                    aria-label=${`Delete @${_.agent_name}`}
                                                    onClick=${(W0)=>{W0.stopPropagation(),i(!1),K0(_.chat_jid)}}
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
                            ${(v0||s0||y0)&&G`
                                <div class="compose-model-popup-actions">
                                    ${v0&&G`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn primary${_0.findIndex((_)=>_.key==="action:new")===F1?" active":""}`}
                                            onClick=${()=>{A4()}}
                                            title="Create a new agent/session branch from this chat"
                                        >
                                            New
                                        </button>
                                    `}
                                    ${s0&&G`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn${_0.findIndex((_)=>_.key==="action:rename")===F1?" active":""}`}
                                            onClick=${(_)=>{Y4(_)}}
                                            title="Rename the current branch handle"
                                            disabled=${P2}
                                        >
                                            Rename current…
                                        </button>
                                    `}
                                    ${y0&&G`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn danger${_0.findIndex((_)=>_.key==="action:delete")===F1?" active":""}`}
                                            onClick=${()=>{D4()}}
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
                <div class="compose-footer" ref=${u0}>
                    ${N4&&G`
                    <div class="compose-meta-row">
                        ${G2&&G`
                            <div class="compose-model-meta">
                                <button
                                    ref=${B4}
                                    type="button"
                                    class="compose-model-hint compose-model-hint-btn"
                                    title=${G4}
                                    aria-label="Open model picker"
                                    onClick=${Y5}
                                    disabled=${Q0}
                                >
                                    ${Q0?"Switching…":N2}
                                </button>
                                <div class="compose-model-meta-subline">
                                    ${!Q0&&W4&&G`
                                        <span class="compose-model-usage-hint" title=${G4}>
                                            ${W4}
                                        </span>
                                    `}
                                </div>
                            </div>
                        `}
                        ${!q&&t&&t.percent!=null&&G`
                            <${a$} usage=${t} onCompact=${A5} />
                        `}
                    </div>
                    `}
                    <div class="compose-actions ${q?"search-mode":""}">
                    ${B2&&G`
                        <div
                            ref=${F0}
                            class="compose-session-trigger-group"
                        >
                            ${v1?.agent_name&&G`
                                <button
                                    type="button"
                                    class=${`compose-session-trigger compose-session-trigger-pill${E?" active":""}`}
                                    onClick=${F4}
                                    title=${v1?.chat_jid||j1}
                                    aria-label=${`Manage sessions for @${v1.agent_name}`}
                                    aria-expanded=${E?"true":"false"}
                                >
                                    <span class="compose-current-agent-label active">@${v1.agent_name}</span>
                                </button>
                            `}
                            <button
                                type="button"
                                class=${`compose-session-trigger compose-session-trigger-icon-btn${E?" active":""}`}
                                onClick=${F4}
                                title=${v1?.chat_jid||j1}
                                aria-label=${v1?.agent_name?`Manage sessions for @${v1.agent_name}`:"Manage Sessions/Agents"}
                                aria-expanded=${E?"true":"false"}
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
                    ${q&&G`
                        <label class="compose-search-scope-wrap" title="Search scope">
                            <span class="compose-search-scope-label">Scope</span>
                            <select
                                class="compose-search-scope-select"
                                value=${B}
                                onChange=${(_)=>J?.(_.currentTarget.value)}
                            >
                                <option value="current">Current</option>
                                <option value="root">Branch family</option>
                                <option value="all">All chats</option>
                            </select>
                        </label>
                    `}
                    <button
                        class="icon-btn search-toggle"
                        onClick=${q?X:K}
                        title=${q?"Close search":"Search"}
                    >
                        ${q?G`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 6L6 18M6 6l12 12"/>
                            </svg>
                        `:G`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="11" cy="11" r="8"/>
                                <path d="M21 21l-4.35-4.35"/>
                            </svg>
                        `}
                    </button>
                    ${J4&&!q&&G`
                        <button
                            class="icon-btn location-btn"
                            onClick=${O3}
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
                    ${X5&&!q&&G`
                        <button
                            class=${`icon-btn notification-btn${X4?" active":""}`}
                            onClick=${D1}
                            title=${U5}
                            type="button"
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
                                <path d="M13.73 21a2 2 0 0 1-3.46 0" />
                            </svg>
                        </button>
                    `}
                    ${!q&&G`
                        ${x&&U1&&G`
                            <button
                                class="icon-btn attach-editor-btn"
                                onClick=${U1}
                                title=${`Attach open file: ${x}`}
                                type="button"
                                disabled=${z.includes(x)}
                            >
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
                            </button>
                        `}
                        <label class="icon-btn" title="Attach file">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                            <input type="file" multiple hidden onChange=${y} />
                        </label>
                    `}
                    ${(M1!=="connected"||!q)&&G`
                        <div class="compose-send-stack">
                            ${M1!=="connected"&&G`
                                <span class="compose-connection-status connection-status ${$2.statusClass}" title=${W5}>
                                    ${z5}
                                </span>
                            `}
                            ${!q&&G`
                                <button 
                                    class=${h0.className}
                                    type="button"
                                    onClick=${()=>{if(s$(h0.mode)){W("/abort","steer");return}W()}}
                                    disabled=${h0.disabled}
                                    title=${h0.title}
                                    aria-label=${h0.ariaLabel}
                                >
                                    ${h0.mode==="compacting"?G`
                                            <span class="compose-submit-spinner" aria-hidden="true">
                                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                                                    <circle class="compose-submit-spinner-ring" cx="12" cy="12" r="10.5" stroke-width="2.25" stroke-linecap="round"></circle>
                                                    <rect class="compose-submit-spinner-stop" x="6" y="6" width="12" height="12" rx="0" fill="currentColor"></rect>
                                                </svg>
                                            </span>
                                        `:h0.mode==="abort"?G`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><rect x="4" y="4" width="16" height="16" rx="2.5"/></svg>`:G`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>`}
                                </button>
                            `}
                        </div>
                    `}
                </div>
            </div>
        </div>
        </div>
    `}function L3(...$){for(let j of $)if(typeof j==="string"&&j.trim())return j.trim();return null}function Uj($){if($.startsWith('"')&&$.endsWith('"')||$.startsWith("'")&&$.endsWith("'"))return $.slice(1,-1);return $}function O8($){if(typeof $!=="string"||!$.trim())return null;let j=$.match(/^\s*cd\s+(.+?)(?:\s*(?:&&|;|\n))/s);if(!j?.[1])return null;return Uj(j[1].trim())||null}function T8($,j){let q=j&&typeof j==="object"?j:null;if(!q)return null;let B=L3(q.cwd,q.working_directory,q.workingDirectory);if(B)return B;let Q=L3(q.project_dir,q.projectDir,q.repo_path,q.repoPath);if(Q)return Q;let J=L3(q.command),K=O8(J);if(K)return K;if(Array.isArray(q.commands))for(let X of q.commands){let z=O8(X);if(z)return z}return null}var _j=G`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>
`,zj=G`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <path d="M6 3v12"></path>
        <circle cx="18" cy="6" r="3"></circle>
        <circle cx="6" cy="18" r="3"></circle>
        <path d="M18 9a9 9 0 0 1-9 9"></path>
    </svg>
`,Wj=G`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <circle cx="12" cy="12" r="9"></circle>
        <path d="M12 7v5l3 2"></path>
    </svg>
`,Gj=1e4;function Nj($){return(Array.isArray($)?$:$&&Array.isArray($.status_hints)?$.status_hints:[]).filter((q)=>q&&typeof q==="object").map((q,B)=>({key:typeof q.key==="string"&&q.key.trim()?q.key.trim():`hint-${B}`,iconSvg:typeof q.icon_svg==="string"?q.icon_svg.trim():"",label:typeof q.label==="string"?q.label.trim():"",title:typeof q.title==="string"?q.title.trim():""})).filter((q)=>q.iconSvg&&q.label)}function Lj($){if(!($ instanceof Set)||$.size===0)return null;let j=Array.from($.values());for(let q=j.length-1;q>=0;q-=1){let B=j[q];if(B==="thought"||B==="draft")return B}return null}function Vj($){if(!Array.isArray($)||$.length===0)return[];let j=new Map([["ssh",0]]);return $.map((q,B)=>({hint:q,index:B})).sort((q,B)=>{let Q=j.get(q.hint?.key)??100,J=j.get(B.hint?.key)??100;if(Q!==J)return Q-J;return q.index-B.index}).map((q)=>q.hint)}function Zj($,j){let q=typeof $==="string"?$.trim():"",B=typeof j==="string"?j.trim():"";return[q?q.split(/[\\/]+/).filter(Boolean).pop()||q:"",B].filter(Boolean).join(" • ")}function M8($){if(!$||typeof $!=="object")return!1;let j=typeof $.type==="string"?$.type:"",q=Boolean($.last_activity||$.lastActivity),B=j==="tool_call"||j==="tool_status"||Boolean($.tool_name||$.tool_args);if(!q&&!B)return!1;return _3($)!==null}function S8($){if(!$||typeof $!=="object")return!1;return $.type==="intent"&&j5($)!==null}function y8($,j=Date.now()){if(!Number.isFinite($))return!1;return j-$>=Gj}function Fj($,j=Date.now()){if(!M8($))return null;let q=_3($);if(q===null||!y8(q,j))return null;let B=V3(new Date(q).toISOString(),j);return B?`${B} ago`:null}function Hj($,j=Date.now()){if(!S8($))return null;let q=j5($);if(q===null||!y8(q,j))return null;return B5($,j)}function Yj($,j={}){let q=j?.isLastActivity??Boolean($?.last_activity||$?.lastActivity),B=$?.title,Q=$?.status,J="";if($?.type==="plan")J=B?`Planning: ${B}`:"Planning...";else if($?.type==="tool_call")J=B?`Running: ${B}`:"Running tool...";else if($?.type==="tool_status")J=B?`${B}: ${Q||"Working..."}`:Q||"Working...";else if($?.type==="error")J=B||"Agent error";else J=B||Q||"Working...";if(!q)return J;if(J&&J!=="Working...")return`Recent activity: ${J}`;return"Last activity"}function V3($,j=Date.now()){if(!$)return null;let q=j-new Date($).getTime();if(!Number.isFinite(q)||q<0)return null;let B=Math.floor(q/1000),Q=Math.floor(B/3600),J=Math.floor(B%3600/60),K=B%60;if(Q>0)return`${Q}h ${J}m`;if(J>0)return`${J}m ${K}s`;return`${K}s`}function k8({status:$,draft:j,plan:q,thought:B,pendingRequest:Q,intent:J,extensionPanels:K=[],pendingPanelActions:X=new Set,onExtensionPanelAction:z,turnId:U,steerQueued:L,onPanelToggle:N,showCorePanels:V=!0,showExtensionPanels:O=!0}){let p=(C)=>{if(!C)return{text:"",totalLines:0,fullText:""};if(typeof C==="string"){let n=C,V1=n?n.replace(/\r\n/g,`
`).split(`
`).length:0;return{text:n,totalLines:V1,fullText:n}}let b=C.text||"",E=C.fullText||C.full_text||b,i=Number.isFinite(C.totalLines)?C.totalLines:E?E.replace(/\r\n/g,`
`).split(`
`).length:0;return{text:b,totalLines:i,fullText:E}},d=160,P=(C)=>String(C||"").replace(/<\/?internal>/gi,""),t=(C)=>{if(!C)return 1;return Math.max(1,Math.ceil(C.length/160))},v=(C,b,E)=>{let i=(C||"").replace(/\r\n/g,`
`).replace(/\r/g,`
`);if(!i)return{text:"",omitted:0,totalLines:Number.isFinite(E)?E:0,visibleLines:0};let n=i.split(`
`),V1=n.length>b?n.slice(0,b).join(`
`):i,Z1=Number.isFinite(E)?E:n.reduce((Y,S)=>Y+t(S),0),G1=V1?V1.split(`
`).reduce((Y,S)=>Y+t(S),0):0,F1=Math.max(Z1-G1,0);return{text:V1,omitted:F1,totalLines:Z1,visibleLines:G1}},e=p(q),W1=p(B),D1=p(j),N1=Boolean(e.text)||e.totalLines>0,X1=Boolean(W1.text)||W1.totalLines>0,x=Boolean(D1.fullText?.trim()||D1.text?.trim()),U1=Boolean($||x||N1||X1||Q||J),E1=Array.isArray(K)&&K.length>0,[b1,P1]=M(new Set),[l,s]=M(null),[Q1,_1]=M(()=>Date.now()),D=(C)=>P1((b)=>{let E=new Set(b),i=!E.has(C);if(i)E.add(C);else E.delete(C);if(typeof N==="function")N(C,i);return E});w(()=>{P1(new Set),s(null)},[U]),w(()=>{if(!(Array.isArray(K)&&K.some((E)=>b1.has(E?.key)&&(E?.started_at||E?.last_activity_at))))return;let b=setInterval(()=>_1(Date.now()),1000);return()=>clearInterval(b)},[b1,K]);let r=H1(()=>Lj(b1),[b1]);w(()=>{if(!r||typeof document>"u")return;let C=(b)=>{if(b?.defaultPrevented)return;if(b?.key!=="Escape")return;if(b?.altKey||b?.ctrlKey||b?.metaKey||b?.shiftKey)return;let E=b?.target;if(E instanceof Element){if(E.closest?.('input, textarea, select, [contenteditable="true"]'))return;if(E.isContentEditable)return}if(P1((i)=>{if(!(i instanceof Set)||!i.has(r))return i;let n=new Set(i);return n.delete(r),n}),typeof N==="function")N(r,!1);b.preventDefault?.(),b.stopPropagation?.()};return document.addEventListener("keydown",C),()=>document.removeEventListener("keydown",C)},[r,N]);let L1=g2($),j1=Boolean($?.last_activity||$?.lastActivity),M1=H1(()=>M8($),[$]),R1=H1(()=>S8($),[$]),Y1=H1(()=>T8($?.tool_name,$?.tool_args),[$?.tool_name,$?.tool_args]),[p1,C1]=M(null);w(()=>{if(!Boolean(R1||$?.retry_at||$?.retryAt||M1))return;_1(Date.now());let b=setInterval(()=>_1(Date.now()),1000);return()=>clearInterval(b)},[M1,R1,$?.retry_at,$?.retryAt,$?.last_event_at,$?.lastEventAt,$?.started_at,$?.startedAt,$?.type,$?.tool_name,$?.tool_args]),w(()=>{if(!($?.type==="tool_call"||$?.type==="tool_status")||!Y1){C1(null);return}let b=!0;return T6(Y1).then((E)=>{if(!b)return;if(E?.branch)C1({branch:E.branch,repoPath:E.repo_path||null,path:Y1});else C1(null)}).catch(()=>{if(b)C1(null)}),()=>{b=!1}},[$?.type,Y1]);let j0=$?.turn_id||U,s1=d6(j0),O0=e4({steerQueued:L}),K0=(C)=>C,f0=Z8($,{isLastActivity:j1}),G0=$5($,{isLastActivity:j1}),o1=$5(null,{pendingRequest:!0}),q0=(C)=>C==="warning"?"#f59e0b":C==="error"?"var(--danger-color)":C==="success"?"var(--success-color)":s1,H0=J?.kind||"info",y1=q0(H0),X0=q0($?.kind||(L1?"warning":"info")),N0=Yj($,{isLastActivity:j1}),a1=Fj($,Q1),t1=p1?.repoPath||"",a=p1?.branch||"",z1=p1?Zj(t1,a):"",J1=Nj($?.status_hints||$?.statusHints),A1=H1(()=>Vj(J1),[J1]),k1=H1(()=>A1.filter((C)=>C?.key==="ssh"),[A1]),c1=H1(()=>A1.filter((C)=>C?.key!=="ssh"),[A1]);if((!V||!U1)&&(!O||!E1))return null;let f1=({panelTitle:C,text:b,fullText:E,totalLines:i,maxLines:n,titleClass:V1,panelKey:Z1})=>{let G1=b1.has(Z1),F1=E||b||"",Y=Z1==="thought"||Z1==="draft"?P(F1):F1,S=typeof n==="number",$1=G1&&S,K1=S?v(Y,n,i):{text:Y||"",omitted:0,totalLines:Number.isFinite(i)?i:0};if(!Y&&!(Number.isFinite(K1.totalLines)&&K1.totalLines>0))return null;let c=`agent-thinking-body${S?" agent-thinking-body-collapsible":""}`,O1=S?`--agent-thinking-collapsed-lines: ${n};`:"";return G`
            <div
                class="agent-thinking"
                data-expanded=${G1?"true":"false"}
                data-collapsible=${S?"true":"false"}
                style=${s1?`--turn-color: ${s1};`:""}
            >
                <div class="agent-thinking-title ${V1||""}">
                    ${s1&&G`<span class=${O0} aria-hidden="true"></span>`}
                    ${C}
                    ${$1&&G`
                        <button
                            class="agent-thinking-close"
                            aria-label=${`Close ${C} panel`}
                            onClick=${()=>D(Z1)}
                        >
                            ×
                        </button>
                    `}
                </div>
                <div
                    class=${c}
                    style=${O1}
                    dangerouslySetInnerHTML=${{__html:r5(Y)}}
                />
                ${!G1&&K1.omitted>0&&G`
                    <button class="agent-thinking-truncation" onClick=${()=>D(Z1)}>
                        ▸ ${K1.omitted} more lines
                    </button>
                `}
                ${G1&&K1.omitted>0&&G`
                    <button class="agent-thinking-truncation" onClick=${()=>D(Z1)}>
                        ▴ show less
                    </button>
                `}
            </div>
        `},Y0=Q?.tool_call?.title,e1=Y0?`Awaiting approval: ${Y0}`:"Awaiting approval",l0=Hj($,Q1),B0=(C,b,E=null)=>{let i=q5(C),n=H8(C,Q1),V1=[E,n].filter(Boolean).join(" · "),Z1=e4({steerQueued:L,pulsing:g2(C)||Boolean(n)});return G`
            <div
                class="agent-thinking agent-thinking-intent"
                aria-live="polite"
                style=${b?`--turn-color: ${b};`:""}
                title=${C?.detail||""}
            >
                <div class="agent-thinking-title intent">
                    ${b&&G`<span class=${Z1} aria-hidden="true"></span>`}
                    <span class="agent-thinking-title-text">${i}</span>
                    ${V1&&G`<span class="agent-status-elapsed">${V1}</span>`}
                </div>
                ${C.detail&&G`<div class="agent-thinking-body">${C.detail}</div>`}
            </div>
        `},m1=(C,b,E,i,n,V1,Z1,G1=8,F1=8)=>{let Y=Math.max(n-i,0.000000001),S=Math.max(b-G1*2,1),$1=Math.max(E-F1*2,1),K1=Math.max(Z1-V1,1),c=Z1===V1?b/2:G1+(C.run-V1)/K1*S,O1=F1+($1-(C.value-i)/Y*$1);return{x:c,y:O1}},$0=(C,b,E,i,n,V1,Z1,G1=8,F1=8)=>{if(!Array.isArray(C)||C.length===0)return"";return C.map((Y,S)=>{let{x:$1,y:K1}=m1(Y,b,E,i,n,V1,Z1,G1,F1);return`${S===0?"M":"L"} ${$1.toFixed(2)} ${K1.toFixed(2)}`}).join(" ")},Z0=(C,b="")=>{if(!Number.isFinite(C))return"—";return`${Math.abs(C)>=100?C.toFixed(0):C.toFixed(2).replace(/\.0+$/,"").replace(/(\.\d*[1-9])0+$/,"$1")}${b}`},A0=["var(--accent-color)","var(--success-color)","var(--warning-color, #f59e0b)","var(--danger-color)"],h1=(C,b)=>{let E=A0;if(!Array.isArray(E)||E.length===0)return"var(--accent-color)";if(E.length===1||!Number.isFinite(b)||b<=1)return E[0];let n=Math.max(0,Math.min(Number.isFinite(C)?C:0,b-1))/Math.max(1,b-1)*(E.length-1),V1=Math.floor(n),Z1=Math.min(E.length-1,V1+1),G1=n-V1,F1=E[V1],Y=E[Z1];if(!Y||V1===Z1||G1<=0.001)return F1;if(G1>=0.999)return Y;let S=Math.round((1-G1)*1000)/10,$1=Math.round(G1*1000)/10;return`color-mix(in oklab, ${F1} ${S}%, ${Y} ${$1}%)`},Q0=(C,b="autoresearch")=>{let E=Array.isArray(C)?C.map((c)=>({...c,points:Array.isArray(c?.points)?c.points.filter((O1)=>Number.isFinite(O1?.value)&&Number.isFinite(O1?.run)):[]})).filter((c)=>c.points.length>0):[],i=E.map((c,O1)=>({...c,color:h1(O1,E.length)}));if(i.length===0)return null;let n=320,V1=120,Z1=i.flatMap((c)=>c.points),G1=Z1.map((c)=>c.value),F1=Z1.map((c)=>c.run),Y=Math.min(...G1),S=Math.max(...G1),$1=Math.min(...F1),K1=Math.max(...F1);return G`
            <div class="agent-series-chart agent-series-chart-combined">
                <div class="agent-series-chart-header">
                    <span class="agent-series-chart-title">Tracked variables</span>
                    <span class="agent-series-chart-value">${i.length} series</span>
                </div>
                <div class="agent-series-chart-plot">
                    <svg class="agent-series-chart-svg" viewBox=${`0 0 ${n} ${V1}`} preserveAspectRatio="none" aria-hidden="true">
                        ${i.map((c)=>{let O1=c?.key||c?.label||"series",w1=l?.panelKey===b&&l?.seriesKey===O1;return G`
                                <g key=${O1}>
                                    <path
                                        class=${`agent-series-chart-line${w1?" is-hovered":""}`}
                                        d=${$0(c.points,n,V1,Y,S,$1,K1)}
                                        style=${`--agent-series-color: ${c.color};`}
                                        onMouseEnter=${()=>s({panelKey:b,seriesKey:O1})}
                                        onMouseLeave=${()=>s((T1)=>T1?.panelKey===b&&T1?.seriesKey===O1?null:T1)}
                                    ></path>
                                </g>
                            `})}
                    </svg>
                    <div class="agent-series-chart-points-layer">
                        ${i.flatMap((c)=>{let O1=typeof c?.unit==="string"?c.unit:"",w1=c?.key||c?.label||"series";return c.points.map((T1,S1)=>{let U0=m1(T1,n,V1,Y,S,$1,K1);return G`
                                    <button
                                        key=${`${w1}-point-${S1}`}
                                        type="button"
                                        class="agent-series-chart-point-hit"
                                        style=${`--agent-series-color: ${c.color}; left:${U0.x/n*100}%; top:${U0.y/V1*100}%;`}
                                        onMouseEnter=${()=>s({panelKey:b,seriesKey:w1,run:T1.run,value:T1.value,unit:O1})}
                                        onMouseLeave=${()=>s((I1)=>I1?.panelKey===b?null:I1)}
                                        onFocus=${()=>s({panelKey:b,seriesKey:w1,run:T1.run,value:T1.value,unit:O1})}
                                        onBlur=${()=>s((I1)=>I1?.panelKey===b?null:I1)}
                                        aria-label=${`${c?.label||"Series"} ${Z0(T1.value,O1)} at run ${T1.run}`}
                                    >
                                        <span class="agent-series-chart-point"></span>
                                    </button>
                                `})})}
                    </div>
                </div>
                <div class="agent-series-legend">
                    ${i.map((c)=>{let O1=c.points[c.points.length-1]?.value,w1=typeof c?.unit==="string"?c.unit:"",T1=c?.key||c?.label||"series",S1=l?.panelKey===b&&l?.seriesKey===T1?l:null,U0=S1&&Number.isFinite(S1.value)?S1.value:O1,I1=S1&&typeof S1.unit==="string"?S1.unit:w1,S0=S1&&Number.isFinite(S1.run)?S1.run:null;return G`
                            <div key=${`${T1}-legend`} class=${`agent-series-legend-item${S1?" is-hovered":""}`} style=${`--agent-series-color: ${c.color};`}>
                                <span class="agent-series-legend-swatch" style=${`--agent-series-color: ${c.color};`}></span>
                                <span class="agent-series-legend-label">${c?.label||"Series"}</span>
                                ${S0!==null&&G`<span class="agent-series-legend-run">run ${S0}</span>`}
                                <span class="agent-series-legend-value">${Z0(U0,I1)}</span>
                            </div>
                        `})}
                </div>
            </div>
        `},D0=(C)=>{if(!C)return null;let b=typeof C?.key==="string"?C.key:`panel-${Math.random()}`,E=b1.has(b),i=C?.title||"Extension status",n=C?.collapsed_text||"",V1=String(C?.state||"").replace(/[-_]+/g," ").replace(/^./,(I1)=>I1.toUpperCase()),Z1=q0(C?.state==="completed"?"success":C?.state==="failed"?"error":C?.state==="stopped"?"warning":"info"),G1=e4({steerQueued:L,pulsing:C?.state==="running"}),F1=typeof C?.detail_markdown==="string"?C.detail_markdown.trim():"",Y=typeof C?.last_run_text==="string"?C.last_run_text.trim():"",S=typeof C?.tmux_command==="string"?C.tmux_command.trim():"",$1=Array.isArray(C?.series)?C.series:[],K1=Array.isArray(C?.actions)?C.actions:[],c=V3(C?.started_at),O1=c?` · ${c}`:"",w1=n+O1,T1=Boolean(F1||S||c),S1=Boolean(F1||$1.length>0||S),U0=[i,w1].filter(Boolean).join(" — ");return G`
            <div
                class="agent-thinking agent-thinking-intent agent-thinking-autoresearch"
                aria-live="polite"
                data-expanded=${E?"true":"false"}
                style=${Z1?`--turn-color: ${Z1};`:""}
                title=${!E?U0||i:""}
            >
                <div class="agent-thinking-header agent-thinking-header-inline">
                    <button
                        class="agent-thinking-title intent agent-thinking-title-clickable"
                        type="button"
                        onClick=${()=>S1?D(b):null}
                    >
                        ${Z1&&G`<span class=${G1} aria-hidden="true"></span>`}
                        <span class="agent-thinking-title-text">${i}</span>
                        ${w1&&G`<span class="agent-thinking-title-meta">${w1}</span>`}
                    </button>
                    ${(K1.length>0||S1)&&G`
                        <div class="agent-thinking-tools-inline">
                            ${K1.length>0&&G`
                                <div class="agent-thinking-actions agent-thinking-actions-inline">
                                    ${K1.map((I1)=>{let S0=`${b}:${I1?.key||""}`,T0=X?.has?.(S0);return G`
                                            <button
                                                key=${S0}
                                                class=${`agent-thinking-action-btn${I1?.tone==="danger"?" danger":""}`}
                                                onClick=${()=>z?.(C,I1)}
                                                disabled=${Boolean(T0)}
                                            >
                                                ${T0?"Working…":I1?.label||"Run"}
                                            </button>
                                        `})}
                                </div>
                            `}
                            ${S1&&G`
                                <button
                                    class="agent-thinking-corner-toggle agent-thinking-corner-toggle-inline"
                                    type="button"
                                    aria-label=${`${E?"Collapse":"Expand"} ${i}`}
                                    title=${E?"Collapse details":"Expand details"}
                                    onClick=${()=>D(b)}
                                >
                                    <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        ${E?G`<polyline points="4 6 8 10 12 6"></polyline>`:G`<polyline points="4 10 8 6 12 10"></polyline>`}
                                    </svg>
                                </button>
                            `}
                        </div>
                    `}
                </div>
                ${E&&G`
                    <div class=${`agent-thinking-autoresearch-layout${T1?"":" chart-only"}`}>
                        ${T1&&G`
                            <div class="agent-thinking-autoresearch-meta-stack">
                                ${c&&G`
                                    <div class="agent-thinking-autoresearch-elapsed">
                                        <span title="Experiment duration">⏱ ${c}</span>
                                        ${C?.last_activity_at&&C?.state==="running"&&G`<span title="Since last activity">⟳ ${V3(C.last_activity_at)} ago</span>`}
                                    </div>
                                `}
                                ${F1&&G`
                                    <div
                                        class="agent-thinking-body agent-thinking-autoresearch-detail"
                                        dangerouslySetInnerHTML=${{__html:r5(F1)}}
                                    />
                                `}
                                ${S&&G`
                                    <div class="agent-series-chart-command">
                                        <div class="agent-series-chart-command-header">
                                            <span>Attach to session</span>
                                        </div>
                                        <div class="agent-series-chart-command-shell">
                                            <pre class="agent-series-chart-command-code">${S}</pre>
                                            <button
                                                type="button"
                                                class="agent-series-chart-command-copy"
                                                aria-label="Copy tmux command"
                                                title="Copy tmux command"
                                                onClick=${()=>z?.(C,{key:"copy_tmux",action_type:"autoresearch.copy_tmux",label:"Copy tmux"})}
                                            >
                                                ${_j}
                                            </button>
                                        </div>
                                    </div>
                                `}
                            </div>
                        `}
                        ${$1.length>0?G`
                                <div class="agent-series-chart-stack">
                                    ${Q0($1,b)}
                                    ${Y&&G`<div class="agent-series-chart-note">${Y}</div>`}
                                </div>
                            `:G`<div class="agent-thinking-body agent-thinking-autoresearch-summary">Variable history will appear after the first completed run.</div>`}
                    </div>
                `}
            </div>
        `};return G`
        <div class="agent-status-panel">
            ${V&&J&&B0(J,y1)}
            ${O&&Array.isArray(K)&&K.map((C)=>D0(C))}
            ${V&&$?.type==="intent"&&B0($,X0,l0)}
            ${V&&Q&&G`
                <div class="agent-status agent-status-request" aria-live="polite" style=${s1?`--turn-color: ${s1};`:""}>
                    ${o1==="dot"&&G`<span class=${O0} aria-hidden="true"></span>`}
                    ${o1==="spinner"&&G`<div class="agent-status-spinner"></div>`}
                    <span class="agent-status-text">${e1}</span>
                </div>
            `}
            ${V&&N1&&f1({panelTitle:K0("Planning"),text:e.text,fullText:e.fullText,totalLines:e.totalLines,panelKey:"plan"})}
            ${V&&X1&&f1({panelTitle:K0("Thoughts"),text:W1.text,fullText:W1.fullText,totalLines:W1.totalLines,maxLines:8,titleClass:"thought",panelKey:"thought"})}
            ${V&&x&&f1({panelTitle:K0("Draft"),text:D1.text,fullText:D1.fullText,totalLines:D1.totalLines,maxLines:8,titleClass:"thought",panelKey:"draft"})}
            ${V&&$&&$?.type!=="intent"&&G`
                <div class=${`agent-status${j1?" agent-status-last-activity":""}${$?.type==="error"?" agent-status-error":""}${z1||J1.length>0||a1?" agent-status-multiline":""}`} aria-live="polite" style=${s1?`--turn-color: ${s1};`:""}>
                    ${s1&&f0&&G`<span class=${O0} aria-hidden="true"></span>`}
                    ${$?.type==="error"?G`<span class="agent-status-error-icon" aria-hidden="true">⚠</span>`:G0==="spinner"&&G`<div class="agent-status-spinner"></div>`}
                    <div class="agent-status-copy">
                        <span class="agent-status-text">${N0}</span>
                        ${(z1||A1.length>0||a1)&&G`
                            <span class="agent-status-meta-row">
                                ${k1.map((C)=>G`
                                    <span key=${C.key} class="agent-status-hint-row" title=${C.title||C.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{__html:C.iconSvg}}></span>
                                        <span class="agent-status-hint-label">${C.label}</span>
                                    </span>
                                `)}
                                ${z1&&G`
                                    <span class="agent-status-git-row" title=${Y1||z1}>
                                        <span class="agent-status-git-icon">${zj}</span>
                                        <span class="agent-status-git-label">
                                            ${t1&&G`<span class="agent-status-git-part">${t1}</span>`}
                                            ${t1&&a&&G`<span class="agent-status-git-separator" aria-hidden="true">•</span>`}
                                            ${a&&G`<span class="agent-status-git-part">${a}</span>`}
                                        </span>
                                    </span>
                                `}
                                ${c1.map((C)=>G`
                                    <span key=${C.key} class="agent-status-hint-row" title=${C.title||C.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{__html:C.iconSvg}}></span>
                                        <span class="agent-status-hint-label">${C.label}</span>
                                    </span>
                                `)}
                                ${a1&&G`
                                    <span class="agent-status-hint-row agent-status-activity-row" title=${`${j1?"Recent activity":"Last event"} ${a1}`}>
                                        <span class="agent-status-hint-icon">${Wj}</span>
                                        <span class="agent-status-hint-label">${a1}</span>
                                    </span>
                                `}
                            </span>
                        `}
                    </div>
                </div>
            `}
        </div>
    `}class R8{extensions=new Map;register($){this.extensions.set($.id,$)}unregister($){this.extensions.delete($)}resolve($){let j,q=-1/0;for(let B of this.extensions.values()){if(B.placement!=="tabs")continue;if(!B.canHandle)continue;try{let Q=B.canHandle($);if(Q===!1||Q===0)continue;let J=Q===!0?0:typeof Q==="number"?Q:0;if(J>q)q=J,j=B}catch(Q){console.warn(`[PaneRegistry] canHandle() error for "${B.id}":`,Q)}}return j}list(){return Array.from(this.extensions.values())}getDockPanes(){return Array.from(this.extensions.values()).filter(($)=>$.placement==="dock")}getTabPanes(){return Array.from(this.extensions.values()).filter(($)=>$.placement==="tabs")}get($){return this.extensions.get($)}get size(){return this.extensions.size}}var d0=new R8;var eB=Symbol();var $Q=new TextDecoder("utf-16le",{fatal:!0});Object.hasOwn=Object.hasOwn||function($,j){return Object.prototype.hasOwnProperty.call($,j)};var Cj={Backspace:65288,Tab:65289,Enter:65293,Escape:65307,Insert:65379,Delete:65535,Home:65360,End:65367,PageUp:65365,PageDown:65366,ArrowLeft:65361,ArrowUp:65362,ArrowRight:65363,ArrowDown:65364,Shift:65505,ShiftLeft:65505,ShiftRight:65506,Control:65507,ControlLeft:65507,ControlRight:65508,Alt:65513,AltLeft:65513,AltRight:65514,Meta:65515,MetaLeft:65515,MetaRight:65516,Super:65515,Super_L:65515,Super_R:65516,CapsLock:65509,NumLock:65407,ScrollLock:65300,Pause:65299,PrintScreen:65377,ContextMenu:65383,Menu:65383," ":32};for(let $=1;$<=12;$+=1)Cj[`F${$}`]=65470+($-1);var Ij=new Uint8Array(256);for(let $=0;$<256;$+=1){let j=0;for(let q=0;q<8;q+=1)j=j<<1|$>>q&1;Ij[$]=j}var tQ=String(Date.now());var XJ=String(Date.now());class E8{tabs=new Map;activeId=null;mruOrder=[];listeners=new Set;onChange($){return this.listeners.add($),()=>this.listeners.delete($)}notify(){let $=this.getTabs(),j=this.activeId;for(let q of this.listeners)try{q($,j)}catch(B){console.warn("[tab-store] Change listener failed:",B)}}open($,j){let q=this.tabs.get($);if(!q)q={id:$,label:j||$.split("/").pop()||$,path:$,dirty:!1,pinned:!1},this.tabs.set($,q);return this.activate($),q}activate($){if(!this.tabs.has($))return;this.activeId=$,this.mruOrder=[$,...this.mruOrder.filter((j)=>j!==$)],this.notify()}close($){if(!this.tabs.get($))return!1;if(this.tabs.delete($),this.mruOrder=this.mruOrder.filter((q)=>q!==$),this.activeId===$)this.activeId=this.mruOrder[0]||null;return this.notify(),!0}closeOthers($){for(let[j,q]of this.tabs)if(j!==$&&!q.pinned)this.tabs.delete(j),this.mruOrder=this.mruOrder.filter((B)=>B!==j);if(this.activeId&&!this.tabs.has(this.activeId))this.activeId=$;this.notify()}closeAll(){for(let[$,j]of this.tabs)if(!j.pinned)this.tabs.delete($),this.mruOrder=this.mruOrder.filter((q)=>q!==$);if(this.activeId&&!this.tabs.has(this.activeId))this.activeId=this.mruOrder[0]||null;this.notify()}setDirty($,j){let q=this.tabs.get($);if(!q||q.dirty===j)return;q.dirty=j,this.notify()}togglePin($){let j=this.tabs.get($);if(!j)return;j.pinned=!j.pinned,this.notify()}saveViewState($,j){let q=this.tabs.get($);if(q)q.viewState=j}getViewState($){return this.tabs.get($)?.viewState}rename($,j,q){let B=this.tabs.get($);if(!B)return;if(this.tabs.delete($),B.id=j,B.path=j,B.label=q||j.split("/").pop()||j,this.tabs.set(j,B),this.mruOrder=this.mruOrder.map((Q)=>Q===$?j:Q),this.activeId===$)this.activeId=j;this.notify()}getTabs(){return Array.from(this.tabs.values())}getActiveId(){return this.activeId}getActive(){return this.activeId?this.tabs.get(this.activeId)||null:null}get($){return this.tabs.get($)}get size(){return this.tabs.size}hasUnsaved(){for(let $ of this.tabs.values())if($.dirty)return!0;return!1}getDirtyTabs(){return Array.from(this.tabs.values()).filter(($)=>$.dirty)}nextTab(){let $=this.getTabs();if($.length<=1)return;let j=$.findIndex((B)=>B.id===this.activeId),q=$[(j+1)%$.length];this.activate(q.id)}prevTab(){let $=this.getTabs();if($.length<=1)return;let j=$.findIndex((B)=>B.id===this.activeId),q=$[(j-1+$.length)%$.length];this.activate(q.id)}mruSwitch(){if(this.mruOrder.length>1)this.activate(this.mruOrder[1])}}var Tj=new E8;function P8($){try{return $?.focus?.(),$?.select?.(),!0}catch(j){return!1}}var Q5="workspaceExplorerScale",Mj=["compact","default","comfortable"],Sj=new Set(Mj),yj={compact:{indentPx:14},default:{indentPx:16},comfortable:{indentPx:18}};function f8($,j="default"){if(typeof $!=="string")return j;let q=$.trim().toLowerCase();return Sj.has(q)?q:j}function Z3(){if(typeof window>"u")return{width:0,isTouch:!1};let $=Number(window.innerWidth)||0,j=Boolean(window.matchMedia?.("(pointer: coarse)")?.matches),q=Boolean(window.matchMedia?.("(hover: none)")?.matches),B=Number(globalThis.navigator?.maxTouchPoints||0)>0;return{width:$,isTouch:j||B&&q}}function kj($={}){let j=Math.max(0,Number($.width)||0);if(Boolean($.isTouch))return"comfortable";if(j>0&&j<1180)return"comfortable";return"default"}function Rj($,j={}){if(Boolean(j.isTouch)&&$==="compact")return"default";return $}function F3($={}){let j=kj($),q=$.stored?f8($.stored,j):j;return Rj(q,$)}function w8($){return yj[f8($)]}function Ej($){if(!$||$.kind!=="text")return!1;let j=Number($.size);return!Number.isFinite(j)||j<=262144}function H3($,j){let q=String($||"").trim();if(!q||q.endsWith("/"))return!1;if(typeof j!=="function")return!1;let B=j({path:q,mode:"edit"});if(!B||typeof B!=="object")return!1;return B.id!=="editor"}function x8($,j,q={}){let B=q.resolvePane;if(H3($,B))return!0;return Ej(j)}var Pj=60000,m8=($)=>{if(!$||!$.name)return!1;if($.path===".")return!1;return $.name.startsWith(".")};function fj($){let j=String($||"").trim();if(!j||j.endsWith("/"))return!1;return H3(j,(q)=>d0.resolve(q))}function h8($,j,q,B=0,Q=[]){if(!q&&m8($))return Q;if(!$)return Q;if(Q.push({node:$,depth:B}),$.type==="dir"&&$.children&&j.has($.path))for(let J of $.children)h8(J,j,q,B+1,Q);return Q}function b8($,j,q){if(!$)return"";let B=[],Q=(J)=>{if(!q&&m8(J))return;if(B.push(J.type==="dir"?`d:${J.path}`:`f:${J.path}`),J.children&&j?.has(J.path))for(let K of J.children)Q(K)};return Q($),B.join("|")}function C3($,j){if(!j)return null;if(!$)return j;if($.path!==j.path||$.type!==j.type)return j;let q=Array.isArray($.children)?$.children:null,B=Array.isArray(j.children)?j.children:null;if(!B)return $;let Q=q?new Map(q.map((X)=>[X?.path,X])):new Map,J=!q||q.length!==B.length,K=B.map((X)=>{let z=C3(Q.get(X.path),X);if(z!==Q.get(X.path))J=!0;return z});return J?{...j,children:K}:$}function A3($,j,q){if(!$)return $;if($.path===j)return C3($,q);if(!Array.isArray($.children))return $;let B=!1,Q=$.children.map((J)=>{let K=A3(J,j,q);if(K!==J)B=!0;return K});return B?{...$,children:Q}:$}var p8=4,Y3=14,wj=8,xj=16;function c8($){if(!$)return 0;if($.type==="file"){let B=Math.max(0,Number($.size)||0);return $.__bytes=B,B}let j=Array.isArray($.children)?$.children:[],q=0;for(let B of j)q+=c8(B);return $.__bytes=q,q}function d8($,j=0){let q=Math.max(0,Number($?.__bytes??$?.size??0)),B={name:$?.name||$?.path||".",path:$?.path||".",size:q,children:[]};if(!$||$.type!=="dir"||j>=p8)return B;let Q=Array.isArray($.children)?$.children:[],J=[];for(let X of Q){let z=Math.max(0,Number(X?.__bytes??X?.size??0));if(z<=0)continue;if(X.type==="dir")J.push({kind:"dir",node:X,size:z});else J.push({kind:"file",name:X.name,path:X.path,size:z})}J.sort((X,z)=>z.size-X.size);let K=J;if(J.length>Y3){let X=J.slice(0,Y3-1),z=J.slice(Y3-1),U=z.reduce((L,N)=>L+N.size,0);X.push({kind:"other",name:`+${z.length} more`,path:`${B.path}/[other]`,size:U}),K=X}return B.children=K.map((X)=>{if(X.kind==="dir")return d8(X.node,j+1);return{name:X.name,path:X.path,size:X.size,children:[]}}),B}function v8(){if(typeof window>"u"||typeof document>"u")return!1;let{documentElement:$,body:j}=document,q=$?.getAttribute?.("data-theme")?.toLowerCase?.()||"";if(q==="dark")return!0;if(q==="light")return!1;if($?.classList?.contains("dark")||j?.classList?.contains("dark"))return!0;if($?.classList?.contains("light")||j?.classList?.contains("light"))return!1;return Boolean(window.matchMedia?.("(prefers-color-scheme: dark)")?.matches)}function l8($,j,q){let B=(($+Math.PI/2)*180/Math.PI+360)%360,Q=q?Math.max(30,70-j*10):Math.max(34,66-j*8),J=q?Math.min(70,45+j*5):Math.min(60,42+j*4);return`hsl(${B.toFixed(1)} ${Q}% ${J}%)`}function J5($,j,q,B){return{x:$+q*Math.cos(B),y:j+q*Math.sin(B)}}function I3($,j,q,B,Q,J){let K=Math.PI*2-0.0001,X=J-Q>K?Q+K:J,z=J5($,j,B,Q),U=J5($,j,B,X),L=J5($,j,q,X),N=J5($,j,q,Q),V=X-Q>Math.PI?1:0;return[`M ${z.x.toFixed(3)} ${z.y.toFixed(3)}`,`A ${B} ${B} 0 ${V} 1 ${U.x.toFixed(3)} ${U.y.toFixed(3)}`,`L ${L.x.toFixed(3)} ${L.y.toFixed(3)}`,`A ${q} ${q} 0 ${V} 0 ${N.x.toFixed(3)} ${N.y.toFixed(3)}`,"Z"].join(" ")}var r8={1:[26,46],2:[48,68],3:[70,90],4:[92,112]};function i8($,j,q){let B=[],Q=[],J=Math.max(0,Number(j)||0),K=(X,z,U,L)=>{let N=Array.isArray(X?.children)?X.children:[];if(!N.length)return;let V=Math.max(0,Number(X.size)||0);if(V<=0)return;let O=U-z,g=z;N.forEach((h,p)=>{let d=Math.max(0,Number(h.size)||0);if(d<=0)return;let P=d/V,t=g,v=p===N.length-1?U:g+O*P;if(g=v,v-t<0.003)return;let e=r8[L];if(e){let W1=l8(t,L,q);if(B.push({key:h.path,path:h.path,label:h.name,size:d,color:W1,depth:L,startAngle:t,endAngle:v,innerRadius:e[0],outerRadius:e[1],d:I3(120,120,e[0],e[1],t,v)}),L===1)Q.push({key:h.path,name:h.name,size:d,pct:J>0?d/J*100:0,color:W1})}if(L<p8)K(h,t,v,L+1)})};return K($,-Math.PI/2,Math.PI*3/2,1),{segments:B,legend:Q}}function D3($,j){if(!$||!j)return null;if($.path===j)return $;let q=Array.isArray($.children)?$.children:[];for(let B of q){let Q=D3(B,j);if(Q)return Q}return null}function n8($,j,q,B){if(!q||q<=0)return{segments:[],legend:[]};let Q=r8[1];if(!Q)return{segments:[],legend:[]};let J=-Math.PI/2,K=Math.PI*3/2,X=l8(J,1,B),U=`${j||"."}/[files]`;return{segments:[{key:U,path:U,label:$,size:q,color:X,depth:1,startAngle:J,endAngle:K,innerRadius:Q[0],outerRadius:Q[1],d:I3(120,120,Q[0],Q[1],J,K)}],legend:[{key:U,name:$,size:q,pct:100,color:X}]}}function u8($,j=!1,q=!1){if(!$)return null;let B=c8($),Q=d8($,0),J=Q.size||B,{segments:K,legend:X}=i8(Q,J,q);if(!K.length&&J>0){let z=n8("[files]",Q.path,J,q);K=z.segments,X=z.legend}return{root:Q,totalSize:J,segments:K,legend:X,truncated:j,isDarkTheme:q}}function bj({payload:$}){if(!$)return null;let[j,q]=M(null),[B,Q]=M($?.root?.path||"."),[J,K]=M(()=>[$?.root?.path||"."]),[X,z]=M(!1);w(()=>{let x=$?.root?.path||".";Q(x),K([x]),q(null)},[$?.root?.path,$?.totalSize]),w(()=>{if(!B)return;z(!0);let x=setTimeout(()=>z(!1),180);return()=>clearTimeout(x)},[B]);let U=H1(()=>{return D3($.root,B)||$.root},[$?.root,B]),L=U?.size||$.totalSize||0,{segments:N,legend:V}=H1(()=>{let x=i8(U,L,$.isDarkTheme);if(x.segments.length>0)return x;if(L<=0)return x;let U1=U?.children?.length?"Total":"[files]";return n8(U1,U?.path||$?.root?.path||".",L,$.isDarkTheme)},[U,L,$.isDarkTheme,$?.root?.path]),[O,g]=M(N),h=T(new Map),p=T(0);w(()=>{let x=h.current,U1=new Map(N.map((l)=>[l.key,l])),E1=performance.now(),b1=220,P1=(l)=>{let s=Math.min(1,(l-E1)/220),Q1=s*(2-s),_1=N.map((D)=>{let L1=x.get(D.key)||{startAngle:D.startAngle,endAngle:D.startAngle,innerRadius:D.innerRadius,outerRadius:D.innerRadius},j1=(C1,j0)=>C1+(j0-C1)*Q1,M1=j1(L1.startAngle,D.startAngle),R1=j1(L1.endAngle,D.endAngle),Y1=j1(L1.innerRadius,D.innerRadius),p1=j1(L1.outerRadius,D.outerRadius);return{...D,d:I3(120,120,Y1,p1,M1,R1)}});if(g(_1),s<1)p.current=requestAnimationFrame(P1)};if(p.current)cancelAnimationFrame(p.current);return p.current=requestAnimationFrame(P1),h.current=U1,()=>{if(p.current)cancelAnimationFrame(p.current)}},[N]);let d=O.length?O:N,P=L>0?X2(L):"0 B",t=U?.name||"",e=(t&&t!=="."?t:"Total")||"Total",W1=P,D1=J.length>1,N1=(x)=>{if(!x?.path)return;let U1=D3($.root,x.path);if(!U1||!Array.isArray(U1.children)||U1.children.length===0)return;K((E1)=>[...E1,U1.path]),Q(U1.path),q(null)},X1=()=>{if(!D1)return;K((x)=>{let U1=x.slice(0,-1);return Q(U1[U1.length-1]||$?.root?.path||"."),U1}),q(null)};return G`
        <div class="workspace-folder-starburst">
            <svg viewBox="0 0 240 240" class=${`workspace-folder-starburst-svg${X?" is-zooming":""}`} role="img"
                aria-label=${`Folder sizes for ${U?.path||$?.root?.path||"."}`}
                data-segments=${d.length}
                data-base-size=${L}>
                ${d.map((x)=>G`
                    <path
                        key=${x.key}
                        d=${x.d}
                        fill=${x.color}
                        stroke="var(--bg-primary)"
                        stroke-width="1"
                        class=${`workspace-folder-starburst-segment${j?.key===x.key?" is-hovered":""}`}
                        onMouseEnter=${()=>q(x)}
                        onMouseLeave=${()=>q(null)}
                        onTouchStart=${()=>q(x)}
                        onTouchEnd=${()=>q(null)}
                        onClick=${()=>N1(x)}
                    >
                        <title>${x.label} — ${X2(x.size)}</title>
                    </path>
                `)}
                <g
                    class=${`workspace-folder-starburst-center-hit${D1?" is-drill":""}`}
                    onClick=${X1}
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
                    <text x="120" y="114" text-anchor="middle" class="workspace-folder-starburst-total-label">${e}</text>
                    <text x="120" y="130" text-anchor="middle" class="workspace-folder-starburst-total-value">${W1}</text>
                </g>
            </svg>
            ${V.length>0&&G`
                <div class="workspace-folder-starburst-legend">
                    ${V.slice(0,8).map((x)=>G`
                        <div key=${x.key} class="workspace-folder-starburst-legend-item">
                            <span class="workspace-folder-starburst-swatch" style=${`background:${x.color}`}></span>
                            <span class="workspace-folder-starburst-name" title=${x.name}>${x.name}</span>
                            <span class="workspace-folder-starburst-size">${X2(x.size)}</span>
                            <span class="workspace-folder-starburst-pct">${x.pct.toFixed(1)}%</span>
                        </div>
                    `)}
                </div>
            `}
            ${$.truncated&&G`
                <div class="workspace-folder-starburst-note">Preview is truncated by tree depth/entry limits.</div>
            `}
        </div>
    `}function g8($){if(typeof document>"u"||!$)return;let j=document.createElement("a");j.href=$,j.setAttribute("download",""),j.rel="noopener",j.style.display="none",document.body.appendChild(j),j.click(),j.remove()}function s8($){switch($?.state){case"indexing":return"Indexing workspace…";case"ready":return"Workspace index ready";case"stale":return"Workspace index may be stale";case"failed":return"Workspace index failed";case"never_indexed":return"Workspace index not built yet";default:return"Checking workspace index…"}}function vj($){if(!$)return"Workspace index status";let j=[s8($)];if($.last_indexed_at)j.push(`Last indexed: ${$.last_indexed_at}`);if(typeof $.indexed_file_count==="number")j.push(`Indexed files: ${$.indexed_file_count}`);if(Array.isArray($.roots)&&$.roots.length)j.push(`Roots: ${$.roots.join(", ")}`);if($.last_error)j.push(`Error: ${$.last_error}`);return j.join(`
`)}function uj($){let j=$?.target;if(j&&typeof j==="object")return j;return j?.parentElement||null}function gj($){return Boolean($?.closest?.(".workspace-node-icon, .workspace-label-text"))}function mj($,j=null){let q=uj($),B=q?.closest?.(".workspace-row");if(!B)return null;let Q=B.dataset.type,J=B.dataset.path;if(!J||J===".")return null;if(j===J)return null;let K=$?.touches?.[0];if(!K)return null;return{type:Q,path:J,dragPath:gj(q)?J:null,startX:K.clientX,startY:K.clientY}}function o8({onFileSelect:$,visible:j=!0,active:q=void 0,onOpenEditor:B,onOpenTerminalTab:Q,onOpenVncTab:J,onToggleTerminal:K,terminalVisible:X=!1}){let[z,U]=M(null),[L,N]=M(new Set(["."])),[V,O]=M(null),[g,h]=M(null),[p,d]=M(""),[P,t]=M(null),[,v]=M(null),[e,W1]=M(!0),[D1,N1]=M(!1),[X1,x]=M(null),[U1,E1]=M(()=>t3("workspaceShowHidden",!1)),[b1,P1]=M(!1),[l,s]=M(null),[Q1,_1]=M(null),[D,r]=M(null),[L1,j1]=M(!1),[M1,R1]=M(null),[Y1,p1]=M(null),[C1,j0]=M(null),[s1,O0]=M(!1),[K0,f0]=M(()=>v8()),[G0,o1]=M(()=>F3({stored:k0(Q5),...Z3()})),[q0,H0]=M(!1),y1=T(L),X0=T(""),N0=T(null),a1=T(0),t1=T(new Set),a=T(null),z1=T(null),J1=T(new Map),A1=T($),k1=T(B),c1=T(null),f1=T(null),Y0=T(null),e1=T(null),l0=T(null),B0=T(null),m1=T("."),$0=T(0),Z0=T({path:null,dragging:!1,startX:0,startY:0}),A0=T({path:null,dragging:!1,startX:0,startY:0}),h1=T({path:null,timer:0}),Q0=T(!1),D0=T(0),C=T(new Map),b=T(null),E=T(null),i=T(null),n=T(null),V1=T(null),Z1=T(null),G1=T(U1),F1=T(j),Y=T(q??j),S=T(0),$1=T(D),K1=T(b1),c=T(l),O1=T(null),w1=T({x:0,y:0}),T1=T(0),S1=T(null),U0=T(V),I1=T(g),S0=T(null),T0=T(P);A1.current=$,k1.current=B,w(()=>{y1.current=L},[L]),w(()=>{G1.current=U1},[U1]),w(()=>{F1.current=j},[j]),w(()=>{Y.current=q??j},[q,j]),w(()=>{$1.current=D},[D]);let g1=m(()=>{if(!$0.current)return;clearTimeout($0.current),$0.current=0},[]);w(()=>()=>g1(),[g1]),w(()=>{if(typeof window>"u")return;let W=()=>{o1(F3({stored:k0(Q5),...Z3()}))};W();let Z=()=>W(),H=()=>W(),A=(R)=>{if(!R||R.key===null||R.key===Q5)W()};window.addEventListener("resize",Z),window.addEventListener("focus",H),window.addEventListener("storage",A);let k=window.matchMedia?.("(pointer: coarse)"),y=window.matchMedia?.("(hover: none)"),u=(R,q1)=>{if(!R)return;if(R.addEventListener)R.addEventListener("change",q1);else if(R.addListener)R.addListener(q1)},o=(R,q1)=>{if(!R)return;if(R.removeEventListener)R.removeEventListener("change",q1);else if(R.removeListener)R.removeListener(q1)};return u(k,Z),u(y,Z),()=>{window.removeEventListener("resize",Z),window.removeEventListener("focus",H),window.removeEventListener("storage",A),o(k,Z),o(y,Z)}},[]),w(()=>{let W=(Z)=>{let H=Z?.detail?.path;if(!H)return;let A=H.split("/"),k=[];for(let y=1;y<A.length;y++)k.push(A.slice(0,y).join("/"));if(k.length)N((y)=>{let u=new Set(y);u.add(".");for(let o of k)u.add(o);return u});O(H),requestAnimationFrame(()=>{let y=document.querySelector(`[data-path="${CSS.escape(H)}"]`);if(y)y.scrollIntoView({block:"nearest",behavior:"smooth"})})};return window.addEventListener("workspace-reveal-path",W),()=>window.removeEventListener("workspace-reveal-path",W)},[]),w(()=>{K1.current=b1},[b1]),w(()=>{c.current=l},[l]),w(()=>{U0.current=V},[V]),w(()=>{I1.current=g},[g]),w(()=>{T0.current=P},[P]),w(()=>{if(typeof window>"u"||typeof document>"u")return;let W=()=>f0(v8());W();let Z=window.matchMedia?.("(prefers-color-scheme: dark)"),H=()=>W();if(Z?.addEventListener)Z.addEventListener("change",H);else if(Z?.addListener)Z.addListener(H);let A=typeof MutationObserver<"u"?new MutationObserver(()=>W()):null;if(A?.observe(document.documentElement,{attributes:!0,attributeFilter:["class","data-theme"]}),document.body)A?.observe(document.body,{attributes:!0,attributeFilter:["class","data-theme"]});return()=>{if(Z?.removeEventListener)Z.removeEventListener("change",H);else if(Z?.removeListener)Z.removeListener(H);A?.disconnect()}},[]),w(()=>{if(!g)return;let W=l0.current;if(!W)return;let Z=requestAnimationFrame(()=>{P8(W)});return()=>cancelAnimationFrame(Z)},[g]),w(()=>{if(!q0)return;let W=(H)=>{let A=H?.target;if(!(A instanceof Element))return;if(V1.current?.contains(A))return;if(Z1.current?.contains(A))return;H0(!1)},Z=(H)=>{if(H?.key==="Escape")H0(!1),Z1.current?.focus?.()};return document.addEventListener("mousedown",W),document.addEventListener("touchstart",W,{passive:!0}),document.addEventListener("keydown",Z),()=>{document.removeEventListener("mousedown",W),document.removeEventListener("touchstart",W),document.removeEventListener("keydown",Z)}},[q0]);let m2=async(W,Z={})=>{let H=Boolean(Z?.autoOpen),A=String(W||"").trim();N1(!0),t(null),v(null);try{let k=await H6(A,20000);if(H&&A&&x8(A,k,{resolvePane:(y)=>d0.resolve(y)}))return k1.current?.(A,k),k;return t(k),k}catch(k){let y={error:k.message||"Failed to load preview"};return t(y),y}finally{N1(!1)}};c1.current=m2;let h2=m(async()=>{try{let W=await Y6("all");return j0(W),W}catch(W){return console.warn("[workspace-explorer] Failed to load workspace index status:",W),null}},[]);z1.current=h2;let w0=m(()=>{z1.current?.()},[]),B4=async()=>{if(!F1.current)return;try{let W=await c4("",1,G1.current),Z=b8(W.root,y1.current,G1.current);if(Z===X0.current){W1(!1);return}if(X0.current=Z,N0.current=W.root,!a1.current)a1.current=requestAnimationFrame(()=>{a1.current=0,U((H)=>C3(H,N0.current)),W1(!1)})}catch(W){x(W.message||"Failed to load workspace"),W1(!1)}},p2=async(W)=>{if(!W)return;if(t1.current.has(W))return;t1.current.add(W);try{let Z=await c4(W,1,G1.current);U((H)=>A3(H,W,Z.root))}catch(Z){x(Z.message||"Failed to load workspace")}finally{t1.current.delete(W)}};f1.current=p2;let F0=m(()=>{let W=V;if(!W)return".";let Z=J1.current?.get(W);if(Z&&Z.type==="dir")return Z.path;if(W==="."||!W.includes("/"))return".";let H=W.split("/");return H.pop(),H.join("/")||"."},[V]),u0=m((W)=>{let Z=W?.closest?.(".workspace-row");if(!Z)return null;let H=Z.dataset.path,A=Z.dataset.type;if(!H)return null;if(A==="dir")return H;if(H.includes("/")){let k=H.split("/");return k.pop(),k.join("/")||"."}return"."},[]),L0=m((W)=>{return u0(W?.target||null)},[u0]),J0=m((W)=>{$1.current=W,r(W)},[]),V0=m(()=>{let W=h1.current;if(W?.timer)clearTimeout(W.timer);h1.current={path:null,timer:0}},[]),r0=m((W)=>{if(!W||W==="."){V0();return}let Z=J1.current?.get(W);if(!Z||Z.type!=="dir"){V0();return}if(y1.current?.has(W)){V0();return}if(h1.current?.path===W)return;V0();let H=setTimeout(()=>{h1.current={path:null,timer:0},f1.current?.(W),N((A)=>{let k=new Set(A);return k.add(W),k})},600);h1.current={path:W,timer:H}},[V0]),x0=m((W,Z)=>{if(w1.current={x:W,y:Z},T1.current)return;T1.current=requestAnimationFrame(()=>{T1.current=0;let H=O1.current;if(!H)return;let A=w1.current;H.style.transform=`translate(${A.x+12}px, ${A.y+12}px)`})},[]),U2=m((W)=>{if(!W)return;let H=(J1.current?.get(W)?.name||W.split("/").pop()||W).trim();if(!H)return;_1({path:W,label:H})},[]),_2=m(()=>{if(_1(null),T1.current)cancelAnimationFrame(T1.current),T1.current=0;if(O1.current)O1.current.style.transform="translate(-9999px, -9999px)"},[]),Q4=m((W)=>{if(!W)return".";let Z=J1.current?.get(W);if(Z&&Z.type==="dir")return Z.path;if(W==="."||!W.includes("/"))return".";let H=W.split("/");return H.pop(),H.join("/")||"."},[]),R0=m(()=>{h(null),d("")},[]),g0=m((W)=>{if(!W)return;let H=(J1.current?.get(W)?.name||W.split("/").pop()||W).trim();if(!H||W===".")return;h(W),d(H)},[]),z2=m(async()=>{let W=I1.current;if(!W)return;let Z=(p||"").trim();if(!Z){R0();return}let H=J1.current?.get(W),A=(H?.name||W.split("/").pop()||W).trim();if(Z===A){R0();return}try{let y=(await C6(W,Z))?.path||W,u=W.includes("/")?W.split("/").slice(0,-1).join("/")||".":".";if(R0(),x(null),window.dispatchEvent(new CustomEvent("workspace-file-renamed",{detail:{oldPath:W,newPath:y,type:H?.type||"file"}})),H?.type==="dir")N((o)=>{let R=new Set;for(let q1 of o)if(q1===W)R.add(y);else if(q1.startsWith(`${W}/`))R.add(`${y}${q1.slice(W.length)}`);else R.add(q1);return R});if(O(y),H?.type==="dir")t(null),N1(!1),v(null);else c1.current?.(y);f1.current?.(u),w0()}catch(k){x(k?.message||"Failed to rename file")}},[R0,p,w0]),c2=m(async(W)=>{let A=W||".";for(let k=0;k<50;k+=1){let u=`untitled${k===0?"":`-${k}`}.md`;try{let R=(await D6(A,u,""))?.path||(A==="."?u:`${A}/${u}`);if(A&&A!==".")N((q1)=>new Set([...q1,A]));O(R),x(null),f1.current?.(A),c1.current?.(R),w0();return}catch(o){if(o?.status===409||o?.code==="file_exists")continue;x(o?.message||"Failed to create file");return}}x("Failed to create file (untitled name already in use).")},[]),k2=m((W)=>{if(W?.stopPropagation?.(),L1)return;let Z=Q4(U0.current);c2(Z)},[L1,Q4,c2]);w(()=>{if(typeof window>"u")return;let W=(Z)=>{let H=Z?.detail?.updates||[];if(!Array.isArray(H)||H.length===0)return;U((o)=>{let R=o;for(let q1 of H){if(!q1?.root)continue;if(!R||q1.path==="."||!q1.path)R=q1.root;else R=A3(R,q1.path,q1.root)}if(R)X0.current=b8(R,y1.current,G1.current);return W1(!1),R});let A=U0.current;if(Boolean(A)&&H.some((o)=>{let R=o?.path||"";if(!R||R===".")return!0;return A===R||A.startsWith(`${R}/`)||R.startsWith(`${A}/`)}))C.current.clear();if(w0(),!A||!T0.current)return;let y=J1.current?.get(A);if(y&&y.type==="dir")return;if(H.some((o)=>{let R=o?.path||"";if(!R||R===".")return!0;return A===R||A.startsWith(`${R}/`)}))c1.current?.(A)};return window.addEventListener("workspace-update",W),()=>window.removeEventListener("workspace-update",W)},[]),a.current=B4;let J4=T(()=>{if(typeof window>"u")return;let W=window.matchMedia("(min-width: 1024px) and (orientation: landscape)"),Z=Y.current??F1.current,H=document.visibilityState!=="hidden"&&(Z||W.matches&&F1.current);d4(H,G1.current).catch((A)=>{console.debug("[workspace-explorer] Workspace visibility ping failed.",A,{visible:H,showHidden:G1.current})})}).current,i0=T(0),K4=T(()=>{if(i0.current)clearTimeout(i0.current);i0.current=setTimeout(()=>{i0.current=0,J4()},250)}).current;w(()=>{if(F1.current)a.current?.(),z1.current?.();K4()},[j,q]),w(()=>{a.current(),z1.current?.(),J4();let W=setInterval(()=>{a.current(),z1.current?.()},Pj),Z=e3("previewHeight",null),H=Number.isFinite(Z)?Math.min(Math.max(Z,80),600):280;if(D0.current=H,Y0.current)Y0.current.style.setProperty("--preview-height",`${H}px`);let A=window.matchMedia("(min-width: 1024px) and (orientation: landscape)"),k=()=>K4();if(A.addEventListener)A.addEventListener("change",k);else if(A.addListener)A.addListener(k);return document.addEventListener("visibilitychange",k),()=>{if(clearInterval(W),a1.current)cancelAnimationFrame(a1.current),a1.current=0;if(A.removeEventListener)A.removeEventListener("change",k);else if(A.removeListener)A.removeListener(k);if(document.removeEventListener("visibilitychange",k),i0.current)clearTimeout(i0.current),i0.current=0;d4(!1,G1.current).catch((y)=>{console.debug("[workspace-explorer] Workspace visibility teardown ping failed.",y,{showHidden:G1.current})})}},[]);let R2=H1(()=>h8(z,L,U1),[z,L,U1]),X5=H1(()=>new Map(R2.map((W)=>[W.node.path,W.node])),[R2]),X4=H1(()=>w8(G0),[G0]);J1.current=X5;let d1=(V?J1.current.get(V):null)?.type==="dir";w(()=>{if(!V||!d1){p1(null),b.current=null,E.current=null;return}let W=V,Z=`${U1?"hidden":"visible"}:${V}`,H=C.current,A=H.get(Z);if(A?.root){H.delete(Z),H.set(Z,A);let u=u8(A.root,Boolean(A.truncated),K0);if(u)b.current=u,E.current=V,p1({loading:!1,error:null,payload:u});return}let k=b.current,y=E.current;p1({loading:!0,error:null,payload:y===V?k:null}),c4(V,wj,U1).then((u)=>{if(U0.current!==W)return;let o={root:u?.root,truncated:Boolean(u?.truncated)};H.delete(Z),H.set(Z,o);while(H.size>xj){let q1=H.keys().next().value;if(!q1)break;H.delete(q1)}let R=u8(o.root,o.truncated,K0);b.current=R,E.current=V,p1({loading:!1,error:null,payload:R})}).catch((u)=>{if(U0.current!==W)return;p1({loading:!1,error:u?.message||"Failed to load folder size chart",payload:y===V?k:null})})},[V,d1,U1,K0]);let m0=Boolean(P&&P.kind==="text"&&!d1&&(!P.size||P.size<=262144)),U4=m0?"Open in editor":P?.size>262144?"File too large to edit":"File is not editable",n0=Boolean(V&&!d1&&fj(V)),b0=Boolean(V&&V!=="."),U5=Boolean(V&&!d1),_5=Boolean(V&&!d1),$2=V&&d1?c5(V,U1):null,z5=s8(C1),W5=vj(C1),h0=C1?.state||"never_indexed",_4=h0!=="ready",v1=m(()=>H0(!1),[]),l1=m(async(W)=>{v1();try{await W?.()}catch(Z){console.warn("[workspace-explorer] Header menu action failed:",Z)}},[v1]),E2=m(async(W)=>{W?.stopPropagation?.(),O0(!0),j0((Z)=>({scope:"all",last_indexed_at:Z?.last_indexed_at||null,last_error:null,indexed_file_count:Z?.indexed_file_count||0,roots:Z?.roots||[],updated_at:Z?.updated_at||null,state:"indexing"}));try{let Z=await A6("all");j0(Z),x(null),X0.current="",a.current?.()}catch(Z){let H=Z?.message||"Failed to reindex workspace";j0((A)=>({scope:"all",last_indexed_at:A?.last_indexed_at||null,last_error:H,indexed_file_count:A?.indexed_file_count||0,roots:A?.roots||[],updated_at:A?.updated_at||null,state:"failed"})),x(H)}finally{O0(!1)}},[]);w(()=>{let W=i.current;if(n.current)n.current.dispose(),n.current=null;if(!W)return;if(W.innerHTML="",!V||d1||!P||P.error)return;let Z={path:V,content:typeof P.text==="string"?P.text:void 0,mtime:P.mtime,size:P.size,preview:P,mode:"view"},H=d0.resolve(Z)||d0.get("workspace-preview-default");if(!H)return;let A=H.mount(W,Z);return n.current=A,()=>{if(n.current===A)A.dispose(),n.current=null;W.innerHTML=""}},[V,d1,P]);let p0=(W)=>{let Z=W?.target;if(Z instanceof Element)return Z;return Z?.parentElement||null},j2=(W)=>{return Boolean(W?.closest?.(".workspace-node-icon, .workspace-label-text"))},q2=(W)=>{if(!W)return!1;if(W.closest?.('input, textarea, [contenteditable="true"]'))return!0;return Boolean(W.isContentEditable)},P2=T((W)=>{let Z=p0(W),H=Z?.closest?.("[data-path]");if(!H)return;let A=H.dataset.path;if(!A||A===".")return;let k=Boolean(Z?.closest?.("button"))||Boolean(Z?.closest?.("a"))||Boolean(Z?.closest?.("input")),y=Boolean(Z?.closest?.(".workspace-caret"));if(k||y)return;if(I1.current===A)return;g0(A)}).current,s0=T((W)=>{if(Q0.current){Q0.current=!1;return}let Z=p0(W),H=Z?.closest?.("[data-path]");if(e1.current?.focus?.(),!H)return;let A=H.dataset.path,k=H.dataset.type,y=Boolean(Z?.closest?.(".workspace-caret")),u=Boolean(Z?.closest?.("button"))||Boolean(Z?.closest?.("a"))||Boolean(Z?.closest?.("input")),o=U0.current===A,R=I1.current;if(R){if(R===A)return;R0()}if(k==="dir"){if(S0.current=null,O(A),t(null),v(null),N1(!1),!y1.current.has(A))f1.current?.(A);if(o&&!y)return;N((z0)=>{let C0=new Set(z0);if(C0.has(A))C0.delete(A);else C0.add(A);return C0})}else{S0.current=null,O(A);let q1=J1.current.get(A);if(q1)A1.current?.(q1.path,q1);if(!u&&!y)c1.current?.(A)}}).current,v0=T(()=>{X0.current="",a.current(),z1.current?.(),Array.from(y1.current||[]).filter((Z)=>Z&&Z!==".").forEach((Z)=>f1.current?.(Z))}).current,y0=T(()=>{S0.current=null,O(null),t(null),v(null),N1(!1)}).current,B2=T(()=>{E1((W)=>{let Z=!W;if(typeof window<"u")E0("workspaceShowHidden",String(Z));return G1.current=Z,d4(!0,Z).catch((A)=>{console.debug("[workspace-explorer] Workspace visibility refresh after toggling hidden files failed.",A,{showHidden:Z})}),X0.current="",a.current?.(),Array.from(y1.current||[]).filter((A)=>A&&A!==".").forEach((A)=>f1.current?.(A)),Z})}).current,z4=T((W)=>{if(p0(W)?.closest?.("[data-path]"))return;y0()}).current,G2=m(async(W)=>{if(!W)return;let Z=W.split("/").pop()||W;if(!window.confirm(`Delete "${Z}"? This cannot be undone.`))return;try{await O6(W);let A=W.includes("/")?W.split("/").slice(0,-1).join("/")||".":".";if(U0.current===W)y0();f1.current?.(A),x(null),w0()}catch(A){t((k)=>({...k||{},error:A.message||"Failed to delete file"}))}},[y0]),N2=m((W)=>{let Z=e1.current;if(!Z||!W||typeof CSS>"u"||typeof CSS.escape!=="function")return;Z.querySelector(`[data-path="${CSS.escape(W)}"]`)?.scrollIntoView?.({block:"nearest"})},[]),d2=m((W)=>{let Z=p0(W);if(I1.current||q2(Z))return;let H=R2;if(!H||H.length===0)return;let A=V?H.findIndex((k)=>k.node.path===V):-1;if(W.key==="ArrowDown"){W.preventDefault();let k=Math.min(A+1,H.length-1),y=H[k];if(!y)return;if(O(y.node.path),y.node.type!=="dir")A1.current?.(y.node.path,y.node),c1.current?.(y.node.path);else t(null),N1(!1),v(null);N2(y.node.path);return}if(W.key==="ArrowUp"){W.preventDefault();let k=A<=0?0:A-1,y=H[k];if(!y)return;if(O(y.node.path),y.node.type!=="dir")A1.current?.(y.node.path,y.node),c1.current?.(y.node.path);else t(null),N1(!1),v(null);N2(y.node.path);return}if(W.key==="ArrowRight"&&A>=0){let k=H[A];if(k?.node?.type==="dir"&&!L.has(k.node.path))W.preventDefault(),f1.current?.(k.node.path),N((y)=>new Set([...y,k.node.path]));return}if(W.key==="ArrowLeft"&&A>=0){let k=H[A];if(k?.node?.type==="dir"&&L.has(k.node.path))W.preventDefault(),N((y)=>{let u=new Set(y);return u.delete(k.node.path),u});return}if(W.key==="Enter"&&A>=0){W.preventDefault();let k=H[A];if(!k)return;let y=k.node.path;if(k.node.type==="dir"){if(!y1.current.has(y))f1.current?.(y);N((o)=>{let R=new Set(o);if(R.has(y))R.delete(y);else R.add(y);return R}),t(null),v(null),N1(!1)}else A1.current?.(y,k.node),c1.current?.(y);return}if((W.key==="Delete"||W.key==="Backspace")&&A>=0){let k=H[A];if(!k||k.node.type==="dir")return;W.preventDefault(),G2(k.node.path);return}if(W.key==="Escape")W.preventDefault(),y0()},[y0,G2,L,R2,N2,V]),G5=m((W)=>{let Z=mj(W,I1.current);if(!Z)return;Z0.current={path:Z.dragPath,dragging:!1,startX:Z.startX,startY:Z.startY}},[]),l2=m(()=>{let W=Z0.current;if(W?.dragging&&W.path){let Z=$1.current||F0(),H=S1.current;if(typeof H==="function")H(W.path,Z)}Z0.current={path:null,dragging:!1,startX:0,startY:0},S.current=0,P1(!1),s(null),J0(null),V0(),_2()},[F0,_2,J0,V0]),W4=m((W)=>{let Z=Z0.current,H=W?.touches?.[0];if(!H||!Z?.path)return;let A=Math.abs(H.clientX-Z.startX),k=Math.abs(H.clientY-Z.startY),y=A>8||k>8;if(!Z.dragging&&y)Z.dragging=!0,P1(!0),s("move"),U2(Z.path);if(Z.dragging){W.preventDefault(),x0(H.clientX,H.clientY);let u=document.elementFromPoint(H.clientX,H.clientY),o=u0(u)||F0();if($1.current!==o)J0(o);r0(o)}},[u0,F0,U2,x0,J0,r0]),N5=T((W)=>{W.preventDefault();let Z=Y0.current;if(!Z)return;let H=W.clientY,A=D0.current||280,k=W.currentTarget;k.classList.add("dragging"),document.body.style.cursor="row-resize",document.body.style.userSelect="none";let y=H,u=(R)=>{y=R.clientY;let q1=Z.clientHeight-80,z0=Math.min(Math.max(A-(R.clientY-H),80),q1);Z.style.setProperty("--preview-height",`${z0}px`),D0.current=z0},o=()=>{let R=Z.clientHeight-80,q1=Math.min(Math.max(A-(y-H),80),R);D0.current=q1,k.classList.remove("dragging"),document.body.style.cursor="",document.body.style.userSelect="",E0("previewHeight",String(Math.round(q1))),document.removeEventListener("mousemove",u),document.removeEventListener("mouseup",o)};document.addEventListener("mousemove",u),document.addEventListener("mouseup",o)}).current,G4=T((W)=>{W.preventDefault();let Z=Y0.current;if(!Z)return;let H=W.touches[0];if(!H)return;let A=H.clientY,k=D0.current||280,y=W.currentTarget;y.classList.add("dragging"),document.body.style.userSelect="none";let u=(R)=>{let q1=R.touches[0];if(!q1)return;R.preventDefault();let z0=Z.clientHeight-80,C0=Math.min(Math.max(k-(q1.clientY-A),80),z0);Z.style.setProperty("--preview-height",`${C0}px`),D0.current=C0},o=()=>{y.classList.remove("dragging"),document.body.style.userSelect="",E0("previewHeight",String(Math.round(D0.current||k))),document.removeEventListener("touchmove",u),document.removeEventListener("touchend",o),document.removeEventListener("touchcancel",o)};document.addEventListener("touchmove",u,{passive:!1}),document.addEventListener("touchend",o),document.addEventListener("touchcancel",o)}).current,N4=m((W=V)=>{if(!W)return;g8(d5(W))},[V]),o0=async()=>{if(!V||d1)return;await G2(V)},c0=(W)=>{return Array.from(W?.dataTransfer?.types||[]).includes("Files")},L4=m((W)=>{if(!c0(W))return;if(W.preventDefault(),S.current+=1,!K1.current)P1(!0);s("upload");let Z=L0(W)||F0();J0(Z),r0(Z)},[F0,L0,J0,r0]),V4=m((W)=>{if(!c0(W))return;if(W.preventDefault(),W.dataTransfer)W.dataTransfer.dropEffect="copy";if(!K1.current)P1(!0);if(c.current!=="upload")s("upload");let Z=L0(W)||F0();if($1.current!==Z)J0(Z);r0(Z)},[F0,L0,J0,r0]),r2=m((W)=>{if(!c0(W))return;if(W.preventDefault(),S.current=Math.max(0,S.current-1),S.current===0)P1(!1),s(null),J0(null),V0()},[J0,V0]),L2=m(async(W,Z=".")=>{let H=Array.from(W||[]);if(H.length===0)return;let A=Z&&Z!==""?Z:".",k=A!=="."?A:"workspace root";g1(),j1(!0),R1({current:0,total:H.length,name:"",percent:0,done:!1,error:null});try{let y=null;for(let u=0;u<H.length;u++){let o=H[u],R=o?.name||`file ${u+1}`;R1((z0)=>({...z0,current:u+1,name:R,percent:0}));let q1=(z0)=>R1((C0)=>({...C0,percent:z0.percent}));try{y=await p5(o,A,{onProgress:q1})}catch(z0){let C0=z0?.status,Z2=z0?.code;if(C0===409||Z2==="file_exists"){if(!window.confirm(`"${R}" already exists in ${k}. Overwrite?`))continue;y=await p5(o,A,{overwrite:!0,onProgress:q1})}else throw z0}}if(y?.path)S0.current=y.path,O(y.path),c1.current?.(y.path);f1.current?.(A),w0(),R1((u)=>({...u,done:!0})),g1(),$0.current=window.setTimeout(()=>{$0.current=0,R1(null)},1500)}catch(y){x(y.message||"Failed to upload file"),R1((u)=>u?{...u,error:y.message||"Upload failed"}:null),g1(),$0.current=window.setTimeout(()=>{$0.current=0,R1(null)},4000)}finally{j1(!1)}},[g1]),Z4=m(async(W,Z)=>{if(!W)return;let H=J1.current?.get(W);if(!H)return;let A=Z&&Z!==""?Z:".",k=W.includes("/")?W.split("/").slice(0,-1).join("/")||".":".";if(A===k)return;try{let u=(await I6(W,A))?.path||W;if(H.type==="dir")N((o)=>{let R=new Set;for(let q1 of o)if(q1===W)R.add(u);else if(q1.startsWith(`${W}/`))R.add(`${u}${q1.slice(W.length)}`);else R.add(q1);return R});if(O(u),H.type==="dir")t(null),N1(!1),v(null);else c1.current?.(u);f1.current?.(k),f1.current?.(A),w0()}catch(y){x(y?.message||"Failed to move entry")}},[]);S1.current=Z4;let F4=m(async(W)=>{if(!c0(W))return;W.preventDefault(),S.current=0,P1(!1),s(null),r(null),V0();let Z=Array.from(W?.dataTransfer?.files||[]);if(Z.length===0)return;let H=$1.current||L0(W)||F0();await L2(Z,H)},[F0,L0,L2]),H4=m((W)=>{if(W?.stopPropagation?.(),L1)return;let Z=W?.currentTarget?.dataset?.uploadTarget||".";m1.current=Z,B0.current?.click()},[L1]),f2=m(()=>{if(L1)return;let W=U0.current,Z=W?J1.current?.get(W):null;m1.current=Z?.type==="dir"?Z.path:".",B0.current?.click()},[L1]),L5=m(()=>{l1(()=>k2(null))},[l1,k2]),_0=m(()=>{l1(()=>f2())},[l1,f2]),Y4=m(()=>{l1(()=>v0())},[l1,v0]),A4=m(()=>{l1(()=>B2())},[l1,B2]),D4=m(()=>{if(!V||!n0)return;l1(()=>k1.current?.(V,P))},[l1,V,n0,P]),V2=m(()=>{if(!V||!m0)return;l1(()=>k1.current?.(V,P))},[l1,V,m0,P]),V5=m(()=>{if(!V||V===".")return;l1(()=>g0(V))},[l1,V,g0]),Z5=m(()=>{if(!V||d1)return;l1(()=>o0())},[l1,V,d1,o0]),C4=m(()=>{if(!V||d1)return;l1(()=>N4())},[l1,V,d1,N4]),F5=m(()=>{if(!$2)return;v1(),g8($2)},[v1,$2]),i2=m(()=>{v1(),Q?.()},[v1,Q]),H5=m(()=>{v1(),J?.()},[v1,J]),Y5=m(()=>{v1(),K?.()},[v1,K]),A5=m((W)=>{if(!W||W.button!==0)return;let Z=W.currentTarget;if(!Z||!Z.dataset)return;let H=Z.dataset.path;if(!H||H===".")return;if(I1.current===H)return;let A=p0(W);if(A?.closest?.("button, a, input, .workspace-caret"))return;if(!j2(A))return;W.preventDefault(),A0.current={path:H,dragging:!1,startX:W.clientX,startY:W.clientY};let k=(u)=>{let o=A0.current;if(!o?.path)return;let R=Math.abs(u.clientX-o.startX),q1=Math.abs(u.clientY-o.startY),z0=R>4||q1>4;if(!o.dragging&&z0)o.dragging=!0,Q0.current=!0,P1(!0),s("move"),U2(o.path),x0(u.clientX,u.clientY),document.body.style.userSelect="none",document.body.style.cursor="grabbing";if(o.dragging){u.preventDefault(),x0(u.clientX,u.clientY);let C0=document.elementFromPoint(u.clientX,u.clientY),Z2=u0(C0)||F0();if($1.current!==Z2)J0(Z2);r0(Z2)}},y=()=>{document.removeEventListener("mousemove",k),document.removeEventListener("mouseup",y);let u=A0.current;if(u?.dragging&&u.path){let o=$1.current||F0(),R=S1.current;if(typeof R==="function")R(u.path,o)}A0.current={path:null,dragging:!1,startX:0,startY:0},S.current=0,P1(!1),s(null),J0(null),V0(),_2(),document.body.style.userSelect="",document.body.style.cursor="",setTimeout(()=>{Q0.current=!1},0)};document.addEventListener("mousemove",k),document.addEventListener("mouseup",y)},[u0,F0,U2,x0,_2,J0,r0,V0]),D5=m(async(W)=>{let Z=Array.from(W?.target?.files||[]);if(Z.length===0)return;let H=m1.current||".";if(await L2(Z,H),m1.current=".",W?.target)W.target.value=""},[L2]);return G`
        <aside
            class=${`workspace-sidebar${b1?" workspace-drop-active":""}`}
            data-workspace-scale=${G0}
            ref=${Y0}
            onDragEnter=${L4}
            onDragOver=${V4}
            onDragLeave=${r2}
            onDrop=${F4}
        >
            <input type="file" multiple style="display:none" ref=${B0} onChange=${D5} />
            <div class="workspace-header">
                <div class="workspace-header-left">
                    <div class="workspace-menu-wrap">
                        <button
                            ref=${Z1}
                            class=${`workspace-menu-button${q0?" active":""}`}
                            onClick=${(W)=>{W.stopPropagation(),H0((Z)=>!Z)}}
                            title="Workspace actions"
                            aria-label="Workspace actions"
                            aria-haspopup="menu"
                            aria-expanded=${q0?"true":"false"}
                        >
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <line x1="4" y1="7" x2="20" y2="7" />
                                <line x1="4" y1="12" x2="20" y2="12" />
                                <line x1="4" y1="17" x2="20" y2="17" />
                            </svg>
                        </button>
                        ${q0&&G`
                            <div class="workspace-menu-dropdown" ref=${V1} role="menu" aria-label="Workspace options">
                                <button class="workspace-menu-item" role="menuitem" onClick=${L5} disabled=${L1}>New file</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${_0} disabled=${L1}>Upload files</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${Y4}>Refresh tree</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${()=>l1(()=>E2())} disabled=${s1}>
                                    ${s1?"Reindexing workspace…":"Reindex workspace"}
                                </button>
                                <button class=${`workspace-menu-item${U1?" active":""}`} role="menuitem" onClick=${A4}>
                                    ${U1?"Hide hidden files":"Show hidden files"}
                                </button>

                                ${(Q||J||K)&&G`<div class="workspace-menu-separator"></div>`}
                                ${Q&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${i2}>
                                        Open terminal in tab
                                    </button>
                                `}
                                ${J&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${H5}>
                                        Open VNC in tab
                                    </button>
                                `}
                                ${K&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${Y5}>
                                        ${X?"Hide terminal dock":"Show terminal dock"}
                                    </button>
                                `}

                                ${V&&G`<div class="workspace-menu-separator"></div>`}
                                ${n0&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${D4}>Open in tab</button>
                                `}
                                ${V&&!d1&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${V2} disabled=${!m0}>Open in editor</button>
                                `}
                                ${b0&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${V5}>Rename selected</button>
                                `}
                                ${_5&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${C4}>Download selected file</button>
                                `}
                                ${$2&&G`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${F5}>Download selected folder (zip)</button>
                                `}
                                ${U5&&G`
                                    <button class="workspace-menu-item danger" role="menuitem" onClick=${Z5}>Delete selected file</button>
                                `}
                            </div>
                        `}
                    </div>
                    <span>Workspace</span>
                </div>
                <div class="workspace-header-actions">
                    <button class="workspace-create" onClick=${k2} title="New file" disabled=${L1}>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <line x1="12" y1="5" x2="12" y2="19" />
                            <line x1="5" y1="12" x2="19" y2="12" />
                        </svg>
                    </button>
                    <button class="workspace-refresh" onClick=${v0} title="Refresh tree">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <circle cx="12" cy="12" r="8.5" stroke-dasharray="42 12" stroke-dashoffset="6"
                                transform="rotate(75 12 12)" />
                            <polyline points="21 3 21 9 15 9" />
                        </svg>
                    </button>
                </div>
            </div>
            ${_4&&G`
                <div class="workspace-index-status-row">
                    <div class=${`workspace-index-status-chip state-${h0}`} title=${W5}>
                        <span class="workspace-index-status-dot" aria-hidden="true"></span>
                        <span>${z5}</span>
                    </div>
                </div>
            `}
            <div class="workspace-tree" onClick=${z4}>
                ${M1&&G`
                    <div class="workspace-upload-strip">
                        <div class="workspace-upload-strip-text">
                            ${M1.error?G`<span class="workspace-upload-strip-error">${M1.error}</span>`:M1.done?G`<span>Done</span>`:G`<span>${M1.total>1?`Uploading ${M1.current}/${M1.total}: ${M1.name}`:`Uploading ${M1.name}`}${M1.percent>0?` (${M1.percent}%)`:"…"}</span>`}
                        </div>
                        ${!M1.done&&!M1.error&&G`
                            <div class="workspace-upload-strip-bar">
                                <div class="workspace-upload-strip-fill" style=${`width:${M1.percent||0}%`}></div>
                            </div>
                        `}
                    </div>
                `}
                ${e&&G`<div class="workspace-loading">Loading…</div>`}
                ${X1&&G`<div class="workspace-error">${X1}</div>`}
                ${z&&G`
                    <div
                        class="workspace-tree-list"
                        ref=${e1}
                        tabIndex="0"
                        onClick=${s0}
                        onDblClick=${P2}
                        onKeyDown=${d2}
                        onTouchStart=${G5}
                        onTouchEnd=${l2}
                        onTouchMove=${W4}
                        onTouchCancel=${l2}
                    >
                        ${R2.map(({node:W,depth:Z})=>{let H=W.type==="dir",A=W.path===V,k=W.path===g,y=H&&L.has(W.path),u=D&&W.path===D,o=Array.isArray(W.children)&&W.children.length>0?W.children.length:Number(W.child_count)||0;return G`
                                <div
                                    key=${W.path}
                                    class=${`workspace-row${A?" selected":""}${u?" drop-target":""}`}
                                    style=${{paddingLeft:`${8+Z*X4.indentPx}px`}}
                                    data-path=${W.path}
                                    data-type=${W.type}
                                    onMouseDown=${A5}
                                >
                                    <span class="workspace-caret" aria-hidden="true">
                                        ${H?y?G`<svg viewBox="0 0 12 12"><polygon points="1,2 11,2 6,11"/></svg>`:G`<svg viewBox="0 0 12 12"><polygon points="2,1 11,6 2,11"/></svg>`:null}
                                    </span>
                                    <svg class=${`workspace-node-icon${H?" folder":""}`}
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                        aria-hidden="true">
                                        ${H?G`<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>`:G`<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>`}
                                    </svg>
                                    ${k?G`
                                            <input
                                                class="workspace-rename-input"
                                                ref=${l0}
                                                value=${p}
                                                onInput=${(R)=>d(R?.target?.value||"")}
                                                onKeyDown=${(R)=>{if(R.stopPropagation(),R.key==="Enter")R.preventDefault(),z2();else if(R.key==="Escape")R.preventDefault(),R0()}}
                                                onBlur=${R0}
                                                onClick=${(R)=>R.stopPropagation()}
                                            />
                                        `:G`<span class="workspace-label"><span class="workspace-label-text">${W.name}</span></span>`}
                                    ${H&&!y&&o>0&&G`
                                        <span class="workspace-count">${o}</span>
                                    `}
                                    ${H&&G`
                                        <button
                                            class="workspace-folder-upload"
                                            data-upload-target=${W.path}
                                            title="Upload files to this folder"
                                            onClick=${H4}
                                            disabled=${L1}
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
            ${V&&G`
                <div class="workspace-preview-splitter-h" onMouseDown=${N5} onTouchStart=${G4}></div>
                <div class="workspace-preview">
                    <div class="workspace-preview-header">
                        <span class="workspace-preview-title">${V}</span>
                        <div class="workspace-preview-actions">
                            <button class="workspace-create" onClick=${k2} title="New file" disabled=${L1}>
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                    <line x1="12" y1="5" x2="12" y2="19" />
                                    <line x1="5" y1="12" x2="19" y2="12" />
                                </svg>
                            </button>
                            ${!d1&&G`
                                <button
                                    class="workspace-download workspace-edit"
                                    onClick=${()=>m0&&k1.current?.(V,P)}
                                    title=${U4}
                                    disabled=${!m0}
                                >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M12 20h9" />
                                        <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
                                    </svg>
                                </button>
                                <button
                                    class="workspace-download workspace-delete"
                                    onClick=${o0}
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
                            ${d1?G`
                                    <button class="workspace-download" onClick=${f2}
                                        title="Upload files to this folder" disabled=${L1}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 8 12 3 17 8"/>
                                            <line x1="12" y1="3" x2="12" y2="15"/>
                                        </svg>
                                    </button>
                                    <a class="workspace-download" href=${c5(V,U1)} download
                                        title="Download folder as zip" onClick=${(W)=>W.stopPropagation()}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 10 12 15 17 10"/>
                                            <line x1="12" y1="15" x2="12" y2="3"/>
                                        </svg>
                                    </a>`:G`<a class="workspace-download" href=${d5(V)} download
                                        title="Download" onClick=${(W)=>W.stopPropagation()}>
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                        <polyline points="7 10 12 15 17 10"/>
                                        <line x1="12" y1="15" x2="12" y2="3"/>
                                    </svg>
                                </a>`}
                        </div>
                    </div>
                    ${D1&&G`<div class="workspace-loading">Loading preview…</div>`}
                    ${P?.error&&G`<div class="workspace-error">${P.error}</div>`}
                    ${d1&&G`
                        <div class="workspace-preview-text">Folder selected — create file, upload files, or download as zip.</div>
                        ${Y1?.loading&&G`<div class="workspace-loading">Loading folder size preview…</div>`}
                        ${Y1?.error&&G`<div class="workspace-error">${Y1.error}</div>`}
                        ${Y1?.payload&&Y1.payload.segments?.length>0&&G`
                            <${bj} payload=${Y1.payload} />
                        `}
                        ${Y1?.payload&&(!Y1.payload.segments||Y1.payload.segments.length===0)&&G`
                            <div class="workspace-preview-text">No file size data available for this folder yet.</div>
                        `}
                    `}
                    ${P&&!P.error&&!d1&&G`
                        <div class="workspace-preview-body" ref=${i}></div>
                    `}
                </div>
            `}
            ${Q1&&G`
                <div class="workspace-drag-ghost" ref=${O1}>${Q1.label}</div>
            `}
        </aside>
    `}var hj=new Set(["html-viewer","kanban-editor","mindmap-editor"]);function K5($,j,q){let B=String($||"").trim();if(!B)return null;if(j)return j;if(typeof q!=="function")return null;return q({path:B,mode:"edit"})?.id||null}function a8($,j,q){let B=K5($,j,q);return B!=null&&hj.has(B)}function t8($,j,q){return K5($,j,q)==="html-viewer"?"Edit":"Edit Source"}function e8($,j,q){return K5($,j,q)==="editor"}var pj=/\.(docx?|xlsx?|pptx?|odt|ods|odp|rtf)$/i,cj=/\.(csv|tsv)$/i,dj=/\.pdf$/i,lj=/\.(png|jpe?g|gif|webp|bmp|ico|svg)$/i,$9=/\.drawio(\.xml|\.svg|\.png)?$/i;function rj($,{hasPopOutTab:j=!1}={}){let q=typeof $==="string"?$.trim():"";if(!q)return null;if(pj.test(q)){let B="/workspace/raw?path="+encodeURIComponent(q),Q=q.split("/").pop()||"document";return"/office-viewer/?url="+encodeURIComponent(B)+"&name="+encodeURIComponent(Q)}if(cj.test(q))return"/csv-viewer/?path="+encodeURIComponent(q);if(dj.test(q))return"/workspace/raw?path="+encodeURIComponent(q);if(lj.test(q)&&!$9.test(q))return"/image-viewer/?path="+encodeURIComponent(q);if($9.test(q)&&!j)return"/drawio/edit?path="+encodeURIComponent(q);return null}function j9({tabs:$,activeId:j,onActivate:q,onClose:B,onCloseOthers:Q,onCloseAll:J,onTogglePin:K,onTogglePreview:X,onToggleDiff:z,onEditSource:U,previewTabs:L,diffTabs:N,paneOverrides:V,detachedTabs:O,onReattachTab:g,onToggleDock:h,dockVisible:p,onToggleZen:d,zenMode:P,onPopOutTab:t}){let[v,e]=M(null),W1=T(null);w(()=>{if(!v)return;let D=(r)=>{if(r.type==="keydown"&&r.key!=="Escape")return;e(null)};return document.addEventListener("click",D),document.addEventListener("keydown",D),()=>{document.removeEventListener("click",D),document.removeEventListener("keydown",D)}},[v]),w(()=>{let D=(r)=>{if(r.ctrlKey&&r.key==="Tab"){if(r.preventDefault(),!$.length)return;let L1=$.findIndex((j1)=>j1.id===j);if(r.shiftKey){let j1=$[(L1-1+$.length)%$.length];q?.(j1.id)}else{let j1=$[(L1+1)%$.length];q?.(j1.id)}return}if((r.ctrlKey||r.metaKey)&&r.key==="w"){let L1=document.querySelector(".editor-pane");if(L1&&L1.contains(document.activeElement)){if(r.preventDefault(),j)B?.(j)}}};return document.addEventListener("keydown",D),()=>document.removeEventListener("keydown",D)},[$,j,q,B]);let D1=m((D,r)=>{if(D.button===1)D.preventDefault(),B?.(r)},[B]),N1=m((D,r)=>{if(D.defaultPrevented)return;if(D.button===0)q?.(r)},[q]),X1=m((D,r)=>{D.preventDefault(),e({id:r,x:D.clientX,y:D.clientY})},[]),x=m((D)=>{D.preventDefault(),D.stopPropagation()},[]),U1=m((D,r)=>{D.preventDefault(),D.stopPropagation(),B?.(r)},[B]);w(()=>{if(!j||!W1.current)return;let D=W1.current.querySelector(".tab-item.active");if(D)D.scrollIntoView({block:"nearest",inline:"nearest",behavior:"smooth"})},[j]);let E1=m((D)=>{if(!(V instanceof Map))return null;return V.get(D)||null},[V]),b1=H1(()=>$.find((D)=>D.id===v?.id)||null,[v?.id,$]),P1=H1(()=>{let D=v?.id;if(!D)return!1;return a8(D,E1(D),(r)=>d0.resolve(r))},[v?.id,E1]),l=H1(()=>{let D=v?.id;if(!D)return"Edit Source";return t8(D,E1(D),(r)=>d0.resolve(r))},[v?.id,E1]),s=H1(()=>{let D=v?.id;if(!D||!(O instanceof Map))return!1;return O.has(D)},[v?.id,O]),Q1=H1(()=>{let D=v?.id;if(!D||!(N instanceof Set))return!1;return N.has(D)},[v?.id,N]),_1=H1(()=>{let D=v?.id;if(!D)return!1;let r=$.find((j1)=>j1.id===D)||null;if(!r)return!1;return e8(D,E1(D),(j1)=>d0.resolve(j1))&&Boolean(r.dirty||Q1)},[v?.id,Q1,E1,$]);if(!$.length)return null;return G`
        <div class="tab-strip" ref=${W1} role="tablist">
            ${$.map((D)=>G`
                <div
                    key=${D.id}
                    class=${`tab-item${D.id===j?" active":""}${D.dirty?" dirty":""}${D.pinned?" pinned":""}`}
                    role="tab"
                    aria-selected=${D.id===j}
                    title=${D.path}
                    onMouseDown=${(r)=>D1(r,D.id)}
                    onClick=${(r)=>N1(r,D.id)}
                    onContextMenu=${(r)=>X1(r,D.id)}
                >
                    ${D.pinned&&G`
                        <span class="tab-pin-icon" aria-label="Pinned">
                            <svg viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
                                <path d="M4.456.734a1.75 1.75 0 0 1 2.826.504l.613 1.327a3.1 3.1 0 0 0 2.084 1.707l2.454.584c1.332.317 1.8 1.972.832 2.94L11.06 10l3.72 3.72a.75.75 0 1 1-1.06 1.06L10 11.06l-2.204 2.205c-.968.968-2.623.5-2.94-.832l-.584-2.454a3.1 3.1 0 0 0-1.707-2.084l-1.327-.613a1.75 1.75 0 0 1-.504-2.826z"/>
                            </svg>
                        </span>
                    `}
                    <span class="tab-label">${D.label}</span>
                    ${O instanceof Map&&O.has(D.id)&&G`
                        <span class="tab-detached-badge" aria-label="Detached" title="Open in separate window">↗</span>
                    `}
                    <button
                        type="button"
                        class="tab-close"
                        onPointerDown=${x}
                        onMouseDown=${x}
                        onClick=${(r)=>U1(r,D.id)}
                        title=${D.dirty?"Unsaved changes":"Close"}
                        aria-label=${D.dirty?"Unsaved changes":`Close ${D.label}`}
                    >
                        ${D.dirty?G`<span class="tab-dirty-dot" aria-hidden="true"></span>`:G`<svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true" focusable="false" style=${{pointerEvents:"none"}}>
                                <line x1="4" y1="4" x2="12" y2="12" style=${{pointerEvents:"none"}}/>
                                <line x1="12" y1="4" x2="4" y2="12" style=${{pointerEvents:"none"}}/>
                            </svg>`}
                    </button>
                </div>
            `)}
            ${h&&G`
                <div class="tab-strip-spacer"></div>
                <button
                    class=${`tab-strip-dock-toggle${p?" active":""}`}
                    onClick=${h}
                    title=${`${p?"Hide":"Show"} terminal (Ctrl+\`)`}
                    aria-label=${`${p?"Hide":"Show"} terminal`}
                    aria-pressed=${p?"true":"false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="1.75" y="2.25" width="12.5" height="11.5" rx="2"/>
                        <polyline points="4.5 5.25 7 7.75 4.5 10.25"/>
                        <line x1="8.5" y1="10.25" x2="11.5" y2="10.25"/>
                    </svg>
                </button>
            `}
            ${d&&G`
                <button
                    class=${`tab-strip-zen-toggle${P?" active":""}`}
                    onClick=${d}
                    title=${`${P?"Exit":"Enter"} zen mode (Ctrl+Shift+Z)`}
                    aria-label=${`${P?"Exit":"Enter"} zen mode`}
                    aria-pressed=${P?"true":"false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        ${P?G`<polyline points="4 8 1.5 8 1.5 1.5 14.5 1.5 14.5 8 12 8"/><polyline points="4 8 1.5 8 1.5 14.5 14.5 14.5 14.5 8 12 8"/>`:G`<polyline points="5.5 1.5 1.5 1.5 1.5 5.5"/><polyline points="10.5 1.5 14.5 1.5 14.5 5.5"/><polyline points="5.5 14.5 1.5 14.5 1.5 10.5"/><polyline points="10.5 14.5 14.5 14.5 14.5 10.5"/>`}
                    </svg>
                </button>
            `}
        </div>
        ${v&&G`
            <div class="tab-context-menu" style=${{left:v.x+"px",top:v.y+"px"}}>
                <button onClick=${()=>{B?.(v.id),e(null)}}>Close</button>
                <button onClick=${()=>{Q?.(v.id),e(null)}}>Close Others</button>
                <button onClick=${()=>{J?.(),e(null)}}>Close All</button>
                <hr />
                <button onClick=${()=>{K?.(v.id),e(null)}}>
                    ${b1?.pinned?"Unpin":"Pin"}
                </button>
                ${P1&&U&&G`
                    <button onClick=${()=>{U(v.id),e(null)}}>${l}</button>
                `}
                ${s&&g&&G`
                    <button onClick=${()=>{g(v.id),e(null)}}>Reattach</button>
                `}
                ${t&&!s&&G`
                    <button onClick=${()=>{let D=$.find((r)=>r.id===v.id);t(v.id,D?.label),e(null)}}>Open in Window</button>
                `}
                ${_1&&z&&G`
                    <hr />
                    <button onClick=${()=>{q?.(v.id),z(v.id),e(null)}}>${Q1?"Hide Diff":"Compare to Saved"}</button>
                `}
                ${X&&/\.(md|mdx|markdown)$/i.test(v.id)&&G`
                    <hr />
                    <button onClick=${()=>{X(v.id),e(null)}}>
                        ${L?.has(v.id)?"Hide Preview":"Preview"}
                    </button>
                `}
                ${(()=>{let D=rj(v.id,{hasPopOutTab:typeof t==="function"});if(!D)return null;return G`
                        <hr />
                        <button onClick=${()=>{window.open(D,"_blank","noopener"),e(null)}}>Open in New Tab</button>
                    `})()}
            </div>
        `}
    `}var y2="gi",q9="gi_current_session_id",ij=1200;function q4($){return $?`gi:${$}`:"web:default"}async function B9(){let $=await fetch("/api/sessions");if(!$.ok)return[];return(await $.json()).sessions||[]}async function nj($){let j=await fetch("/api/sessions",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({title:$})});if(!j.ok)throw Error("Failed to create session");return j.json()}async function sj(){let $=await fetch("/api/runtime/config");if(!$.ok)return{};return $.json()}function oj({sessions:$,currentSessionId:j,onSelect:q,onCreate:B}){let[Q,J]=M("");return G`
        <div class="gi-session-sidebar">
            <div class="workspace-header">
                <span class="workspace-header-left">Sessions</span>
                <div class="workspace-header-actions">
                    <button
                        class="workspace-create"
                        title="New session"
                        onClick=${async()=>{let K=await B(Q||"Session");J(""),q(K.id)}}
                    >
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
                        </svg>
                    </button>
                </div>
            </div>
            <div class="workspace-tree">
                ${$.map((K)=>G`
                    <button
                        key=${K.id}
                        class=${`workspace-row${j===K.id?" workspace-row-active":""}`}
                        style="padding-left:12px"
                        onClick=${()=>q(K.id)}
                    >
                        <span class="workspace-row-icon">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                            </svg>
                        </span>
                        <span class="workspace-row-label">${K.title||K.id}</span>
                    </button>
                `)}
            </div>
        </div>
    `}function aj(){let[$,j]=M({}),[q,B]=M([]),[Q,J]=M(()=>k0(q9)||null),K=H1(()=>q4(Q),[Q]),{agentStatus:X,setAgentStatus:z,agentDraft:U,setAgentDraft:L,agentPlan:N,setAgentPlan:V,agentThought:O,setAgentThought:g,pendingRequest:h,setPendingRequest:p,currentTurnId:d,setCurrentTurnId:P,steerQueuedTurnId:t,setSteerQueuedTurnId:v,lastAgentEventRef:e,lastSilenceNoticeRef:W1,isAgentRunningRef:D1,draftBufferRef:N1,thoughtBufferRef:X1,previewResyncPendingRef:x,previewResyncGenerationRef:U1,pendingRequestRef:E1,stalledPostIdRef:b1,currentTurnIdRef:P1,steerQueuedTurnIdRef:l,thoughtExpandedRef:s,draftExpandedRef:Q1}=$6(),[_1,D]=M([]),[r,L1]=M(!1),[j1,M1]=M(!1),[R1,Y1]=M([]),[p1,C1]=M(null),[j0,s1]=M({}),[O0,K0]=M(null),[f0,G0]=M(!1);w(()=>{let a=z6();return sj().then((z1)=>{j(z1),K0({name:z1.user_name,avatar_url:z1.user_avatar}),s1({[y2]:{id:y2,name:z1.assistant_name||"Gi",avatar_url:z1.assistant_avatar||null}})}),B9().then(B),a},[]),w(()=>{if(Q)E0(q9,Q)},[Q]);async function o1(a){J(a),D([]),await H0(a)}async function q0(a){let z1=await nj(a),J1=await B9();return B(J1),z1}async function H0(a=Q,z1=null){if(!a)return;let J1=q4(a),k1=(await N6(50,z1,J1)).posts||[];if(z1)D((c1)=>b5([...k1,...c1]));else D(b5(k1));L1(k1.length>=50)}w(()=>{if(!Q)return;let a=setInterval(async()=>{await H0(Q);let z1=q4(Q),J1=await L6(y2,z1).catch(()=>null);if(J1){z(J1);let A1=J1.status==="running"||J1.status==="cancelling";G0(A1),D1.current=A1}else z(null),G0(!1),D1.current=!1},ij);return H0(Q),()=>clearInterval(a)},[Q]);async function y1(a,z1={}){if(!Q)return;let J1=q4(Q);await u2(y2,a,null,[],z1.mode||"auto",J1)}async function X0(){if(!Q)return;let a=q4(Q);await u2(y2,"/abort",null,[],"steer",a).catch(()=>null)}function N0(a){let z1=R1.find((A1)=>A1.path===a);if(z1){C1(z1.id);return}let J1=`tab-${Date.now()}`;Y1((A1)=>[...A1,{id:J1,path:a,label:a.split("/").pop()||a}]),C1(J1)}function a1(a){Y1((z1)=>{let J1=z1.filter((A1)=>A1.id!==a);if(p1===a)C1(J1[J1.length-1]?.id||null);return J1})}let t1=Boolean(Q);return G`
        <div class="app-shell ${j1?"":"workspace-collapsed"}">
            <${o8}
                currentChatJid=${K}
                onOpenFile=${N0}
            />
            <div class="workspace-splitter"></div>
            <div class="container">
                <${oj}
                    sessions=${q}
                    currentSessionId=${Q}
                    onSelect=${o1}
                    onCreate=${q0}
                />
                <div class="chat-window">
                    <div class="chat-window-header">
                        <div class="chat-window-header-main">
                            <div class="chat-window-header-title">
                                ${$.assistant_name||"Gi"}
                            </div>
                            <div class="chat-window-header-subtitle">
                                ${Q?q.find((a)=>a.id===Q)?.title||Q:"No session — create or select one"}
                            </div>
                        </div>
                        <div class="chat-window-header-actions">
                            <button
                                class="chat-window-header-button"
                                title="Toggle workspace"
                                onClick=${()=>M1((a)=>!a)}
                            >
                                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                                    <polyline points="9 22 9 12 15 12 15 22"/>
                                </svg>
                            </button>
                        </div>
                    </div>
                    ${R1.length>0?G`
                        <${j9}
                            tabs=${R1}
                            activeId=${p1}
                            onActivate=${(a)=>C1(a)}
                            onClose=${a1}
                            onCloseOthers=${(a)=>Y1((z1)=>z1.filter((J1)=>J1.id===a))}
                            onCloseAll=${()=>Y1([])}
                            onTogglePin=${()=>{}}
                        />
                    `:null}
                    <${k8}
                        status=${X}
                        draft=${U}
                        plan=${N}
                        thought=${O}
                        pendingRequest=${h}
                        turnId=${d}
                        steerQueued=${Boolean(t)}
                    />
                    <${z8}
                        posts=${_1}
                        hasMore=${r}
                        onLoadMore=${async({preserveScroll:a})=>{let z1=_1[0];if(z1)await H0(Q,z1.id)}}
                        onPostClick=${()=>{}}
                        onHashtagClick=${()=>{}}
                        onMessageRef=${()=>{}}
                        onScrollToMessage=${()=>{}}
                        onFileRef=${N0}
                        onOpenWidget=${()=>{}}
                        onOpenAttachmentPreview=${()=>{}}
                        emptyMessage=${t1?"No messages yet.":"Create or select a session to start."}
                        agents=${j0}
                        user=${O0}
                        onDeletePost=${()=>{}}
                        reverse=${!0}
                    />
                    <${I8}
                        currentChatJid=${K}
                        isAgentActive=${f0}
                        onSend=${(a,z1)=>y1(a,z1)}
                        onAbort=${X0}
                        agents=${j0}
                        currentSessionAgent=${j0[y2]?{...j0[y2],chat_jid:K}:null}
                        agentStatus=${X}
                        agentDraft=${U}
                        contextUsage=${null}
                        disabled=${!t1}
                    />
                </div>
            </div>
        </div>
    `}b2(G`<${aj} />`,document.getElementById("app"));

//# debugId=410E268FB52E446664756E2164756E21
//# sourceMappingURL=app.js.map
