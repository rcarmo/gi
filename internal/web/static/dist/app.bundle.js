// web/src/vendor/preact-htm.js
var f;
var L;
var R0;
var x0;
var I;
var X0;
var D0;
var F0;
var r;
var m;
var v;
var J0;
var u;
var o;
var i;
var L0;
var y = {};
var k = [];
var y0 = /acit|ex(?:s|g|n|p|$)|rph|grid|ows|mnc|ntw|ine[ch]|zoo|^ord|itera/i;
var e = Array.isArray;
function U(O, E) {
  for (var Y in E)
    O[Y] = E[Y];
  return O;
}
function n(O) {
  O && O.parentNode && O.parentNode.removeChild(O);
}
function E0(O, E, Y) {
  var g, V, Z, Q = {};
  for (Z in E)
    Z == "key" ? g = E[Z] : Z == "ref" ? V = E[Z] : Q[Z] = E[Z];
  if (arguments.length > 2 && (Q.children = arguments.length > 3 ? f.call(arguments, 2) : Y), typeof O == "function" && O.defaultProps != null)
    for (Z in O.defaultProps)
      Q[Z] === undefined && (Q[Z] = O.defaultProps[Z]);
  return x(O, Q, g, V, null);
}
function x(O, E, Y, g, V) {
  var Z = { type: O, props: E, key: Y, ref: g, __k: null, __: null, __b: 0, __e: null, __c: null, constructor: undefined, __v: V == null ? ++R0 : V, __i: -1, __u: 0 };
  return V == null && L.vnode != null && L.vnode(Z), Z;
}
function a(O) {
  return O.children;
}
function _(O, E) {
  this.props = O, this.context = E;
}
function H(O, E) {
  if (E == null)
    return O.__ ? H(O.__, O.__i + 1) : null;
  for (var Y;E < O.__k.length; E++)
    if ((Y = O.__k[E]) != null && Y.__e != null)
      return Y.__e;
  return typeof O.type == "function" ? H(O) : null;
}
function k0(O) {
  if (O.__P && O.__d) {
    var E = O.__v, Y = E.__e, g = [], V = [], Z = U({}, E);
    Z.__v = E.__v + 1, L.vnode && L.vnode(Z), O0(O.__P, Z, E, O.__n, O.__P.namespaceURI, 32 & E.__u ? [Y] : null, g, Y == null ? H(E) : Y, !!(32 & E.__u), V), Z.__v = E.__v, Z.__.__k[Z.__i] = Z, C0(g, Z, V), E.__e = E.__ = null, Z.__e != Y && d0(Z);
  }
}
function d0(O) {
  if ((O = O.__) != null && O.__c != null)
    return O.__e = O.__c.base = null, O.__k.some(function(E) {
      if (E != null && E.__e != null)
        return O.__e = O.__c.base = E.__e;
    }), d0(O);
}
function t(O) {
  (!O.__d && (O.__d = true) && I.push(O) && !p.__r++ || X0 != L.debounceRendering) && ((X0 = L.debounceRendering) || D0)(p);
}
function p() {
  try {
    for (var O, E = 1;I.length; )
      I.length > E && I.sort(F0), O = I.shift(), E = I.length, k0(O);
  } finally {
    I.length = p.__r = 0;
  }
}
function j0(O, E, Y, g, V, Z, Q, X, b, W, R) {
  var G, K, D, q, N, C, F, J = g && g.__k || k, B = E.length;
  for (b = p0(Y, E, J, b, B), G = 0;G < B; G++)
    (D = Y.__k[G]) != null && (K = D.__i != -1 && J[D.__i] || y, D.__i = G, C = O0(O, D, K, V, Z, Q, X, b, W, R), q = D.__e, D.ref && K.ref != D.ref && (K.ref && Y0(K.ref, null, D), R.push(D.ref, D.__c || q, D)), N == null && q != null && (N = q), (F = !!(4 & D.__u)) || K.__k === D.__k ? (b = q0(D, b, O, F), F && K.__e && (K.__e = null)) : typeof D.type == "function" && C !== undefined ? b = C : q && (b = q.nextSibling), D.__u &= -7);
  return Y.__e = N, b;
}
function p0(O, E, Y, g, V) {
  var Z, Q, X, b, W, R = Y.length, G = R, K = 0;
  for (O.__k = Array(V), Z = 0;Z < V; Z++)
    (Q = E[Z]) != null && typeof Q != "boolean" && typeof Q != "function" ? (typeof Q == "string" || typeof Q == "number" || typeof Q == "bigint" || Q.constructor == String ? Q = O.__k[Z] = x(null, Q, null, null, null) : e(Q) ? Q = O.__k[Z] = x(a, { children: Q }, null, null, null) : Q.constructor === undefined && Q.__b > 0 ? Q = O.__k[Z] = x(Q.type, Q.props, Q.key, Q.ref ? Q.ref : null, Q.__v) : O.__k[Z] = Q, b = Z + K, Q.__ = O, Q.__b = O.__b + 1, X = null, (W = Q.__i = f0(Q, Y, b, G)) != -1 && (G--, (X = Y[W]) && (X.__u |= 2)), X == null || X.__v == null ? (W == -1 && (V > R ? K-- : V < R && K++), typeof Q.type != "function" && (Q.__u |= 4)) : W != b && (W == b - 1 ? K-- : W == b + 1 ? K++ : (W > b ? K-- : K++, Q.__u |= 4))) : O.__k[Z] = null;
  if (G)
    for (Z = 0;Z < R; Z++)
      (X = Y[Z]) != null && (2 & X.__u) == 0 && (X.__e == g && (g = H(X)), S0(X, X));
  return g;
}
function q0(O, E, Y, g) {
  var V, Z;
  if (typeof O.type == "function") {
    for (V = O.__k, Z = 0;V && Z < V.length; Z++)
      V[Z] && (V[Z].__ = O, E = q0(V[Z], E, Y, g));
    return E;
  }
  O.__e != E && (g && (E && O.type && !E.parentNode && (E = H(O)), Y.insertBefore(O.__e, E || null)), E = O.__e);
  do
    E = E && E.nextSibling;
  while (E != null && E.nodeType == 8);
  return E;
}
function f0(O, E, Y, g) {
  var V, Z, Q, X = O.key, b = O.type, W = E[Y], R = W != null && (2 & W.__u) == 0;
  if (W === null && X == null || R && X == W.key && b == W.type)
    return Y;
  if (g > (R ? 1 : 0)) {
    for (V = Y - 1, Z = Y + 1;V >= 0 || Z < E.length; )
      if ((W = E[Q = V >= 0 ? V-- : Z++]) != null && (2 & W.__u) == 0 && X == W.key && b == W.type)
        return Q;
  }
  return -1;
}
function b0(O, E, Y) {
  E[0] == "-" ? O.setProperty(E, Y == null ? "" : Y) : O[E] = Y == null ? "" : typeof Y != "number" || y0.test(E) ? Y : Y + "px";
}
function T(O, E, Y, g, V) {
  var Z, Q;
  E:
    if (E == "style")
      if (typeof Y == "string")
        O.style.cssText = Y;
      else {
        if (typeof g == "string" && (O.style.cssText = g = ""), g)
          for (E in g)
            Y && E in Y || b0(O.style, E, "");
        if (Y)
          for (E in Y)
            g && Y[E] == g[E] || b0(O.style, E, Y[E]);
      }
    else if (E[0] == "o" && E[1] == "n")
      Z = E != (E = E.replace(J0, "$1")), Q = E.toLowerCase(), E = Q in O || E == "onFocusOut" || E == "onFocusIn" ? Q.slice(2) : E.slice(2), O.l || (O.l = {}), O.l[E + Z] = Y, Y ? g ? Y[v] = g[v] : (Y[v] = u, O.addEventListener(E, Z ? i : o, Z)) : O.removeEventListener(E, Z ? i : o, Z);
    else {
      if (V == "http://www.w3.org/2000/svg")
        E = E.replace(/xlink(H|:h)/, "h").replace(/sName$/, "s");
      else if (E != "width" && E != "height" && E != "href" && E != "list" && E != "form" && E != "tabIndex" && E != "download" && E != "rowSpan" && E != "colSpan" && E != "role" && E != "popover" && E in O)
        try {
          O[E] = Y == null ? "" : Y;
          break E;
        } catch (X) {}
      typeof Y == "function" || (Y == null || Y === false && E[4] != "-" ? O.removeAttribute(E) : O.setAttribute(E, E == "popover" && Y == 1 ? "" : Y));
    }
}
function K0(O) {
  return function(E) {
    if (this.l) {
      var Y = this.l[E.type + O];
      if (E[m] == null)
        E[m] = u++;
      else if (E[m] < Y[v])
        return;
      return Y(L.event ? L.event(E) : E);
    }
  };
}
function O0(O, E, Y, g, V, Z, Q, X, b, W) {
  var R, G, K, D, q, N, C, F, J, B, P, w, W0, $, h, S = E.type;
  if (E.constructor !== undefined)
    return null;
  128 & Y.__u && (b = !!(32 & Y.__u), Z = [X = E.__e = Y.__e]), (R = L.__b) && R(E);
  E:
    if (typeof S == "function")
      try {
        if (F = E.props, J = S.prototype && S.prototype.render, B = (R = S.contextType) && g[R.__c], P = R ? B ? B.props.value : R.__ : g, Y.__c ? C = (G = E.__c = Y.__c).__ = G.__E : (J ? E.__c = G = new S(F, P) : (E.__c = G = new _(F, P), G.constructor = S, G.render = a0), B && B.sub(G), G.state || (G.state = {}), G.__n = g, K = G.__d = true, G.__h = [], G._sb = []), J && G.__s == null && (G.__s = G.state), J && S.getDerivedStateFromProps != null && (G.__s == G.state && (G.__s = U({}, G.__s)), U(G.__s, S.getDerivedStateFromProps(F, G.__s))), D = G.props, q = G.state, G.__v = E, K)
          J && S.getDerivedStateFromProps == null && G.componentWillMount != null && G.componentWillMount(), J && G.componentDidMount != null && G.__h.push(G.componentDidMount);
        else {
          if (J && S.getDerivedStateFromProps == null && F !== D && G.componentWillReceiveProps != null && G.componentWillReceiveProps(F, P), E.__v == Y.__v || !G.__e && G.shouldComponentUpdate != null && G.shouldComponentUpdate(F, G.__s, P) === false) {
            E.__v != Y.__v && (G.props = F, G.state = G.__s, G.__d = false), E.__e = Y.__e, E.__k = Y.__k, E.__k.some(function(z) {
              z && (z.__ = E);
            }), k.push.apply(G.__h, G._sb), G._sb = [], G.__h.length && Q.push(G);
            break E;
          }
          G.componentWillUpdate != null && G.componentWillUpdate(F, G.__s, P), J && G.componentDidUpdate != null && G.__h.push(function() {
            G.componentDidUpdate(D, q, N);
          });
        }
        if (G.context = P, G.props = F, G.__P = O, G.__e = false, w = L.__r, W0 = 0, J)
          G.state = G.__s, G.__d = false, w && w(E), R = G.render(G.props, G.state, G.context), k.push.apply(G.__h, G._sb), G._sb = [];
        else
          do
            G.__d = false, w && w(E), R = G.render(G.props, G.state, G.context), G.state = G.__s;
          while (G.__d && ++W0 < 25);
        G.state = G.__s, G.getChildContext != null && (g = U(U({}, g), G.getChildContext())), J && !K && G.getSnapshotBeforeUpdate != null && (N = G.getSnapshotBeforeUpdate(D, q)), $ = R != null && R.type === a && R.key == null ? N0(R.props.children) : R, X = j0(O, e($) ? $ : [$], E, Y, g, V, Z, Q, X, b, W), G.base = E.__e, E.__u &= -161, G.__h.length && Q.push(G), C && (G.__E = G.__ = null);
      } catch (z) {
        if (E.__v = null, b || Z != null)
          if (z.then) {
            for (E.__u |= b ? 160 : 128;X && X.nodeType == 8 && X.nextSibling; )
              X = X.nextSibling;
            Z[Z.indexOf(X)] = null, E.__e = X;
          } else {
            for (h = Z.length;h--; )
              n(Z[h]);
            l(E);
          }
        else
          E.__e = Y.__e, E.__k = Y.__k, z.then || l(E);
        L.__e(z, E, Y);
      }
    else
      Z == null && E.__v == Y.__v ? (E.__k = Y.__k, E.__e = Y.__e) : X = E.__e = e0(Y.__e, E, Y, g, V, Z, Q, b, W);
  return (R = L.diffed) && R(E), 128 & E.__u ? undefined : X;
}
function l(O) {
  O && (O.__c && (O.__c.__e = true), O.__k && O.__k.some(l));
}
function C0(O, E, Y) {
  for (var g = 0;g < Y.length; g++)
    Y0(Y[g], Y[++g], Y[++g]);
  L.__c && L.__c(E, O), O.some(function(V) {
    try {
      O = V.__h, V.__h = [], O.some(function(Z) {
        Z.call(V);
      });
    } catch (Z) {
      L.__e(Z, V.__v);
    }
  });
}
function N0(O) {
  return typeof O != "object" || O == null || O.__b > 0 ? O : e(O) ? O.map(N0) : U({}, O);
}
function e0(O, E, Y, g, V, Z, Q, X, b) {
  var W, R, G, K, D, q, N, C = Y.props || y, F = E.props, J = E.type;
  if (J == "svg" ? V = "http://www.w3.org/2000/svg" : J == "math" ? V = "http://www.w3.org/1998/Math/MathML" : V || (V = "http://www.w3.org/1999/xhtml"), Z != null) {
    for (W = 0;W < Z.length; W++)
      if ((D = Z[W]) && "setAttribute" in D == !!J && (J ? D.localName == J : D.nodeType == 3)) {
        O = D, Z[W] = null;
        break;
      }
  }
  if (O == null) {
    if (J == null)
      return document.createTextNode(F);
    O = document.createElementNS(V, J, F.is && F), X && (L.__m && L.__m(E, Z), X = false), Z = null;
  }
  if (J == null)
    C === F || X && O.data == F || (O.data = F);
  else {
    if (Z = Z && f.call(O.childNodes), !X && Z != null)
      for (C = {}, W = 0;W < O.attributes.length; W++)
        C[(D = O.attributes[W]).name] = D.value;
    for (W in C)
      D = C[W], W == "dangerouslySetInnerHTML" ? G = D : W == "children" || (W in F) || W == "value" && ("defaultValue" in F) || W == "checked" && ("defaultChecked" in F) || T(O, W, null, D, V);
    for (W in F)
      D = F[W], W == "children" ? K = D : W == "dangerouslySetInnerHTML" ? R = D : W == "value" ? q = D : W == "checked" ? N = D : X && typeof D != "function" || C[W] === D || T(O, W, D, C[W], V);
    if (R)
      X || G && (R.__html == G.__html || R.__html == O.innerHTML) || (O.innerHTML = R.__html), E.__k = [];
    else if (G && (O.innerHTML = ""), j0(E.type == "template" ? O.content : O, e(K) ? K : [K], E, Y, g, J == "foreignObject" ? "http://www.w3.org/1999/xhtml" : V, Z, Q, Z ? Z[0] : Y.__k && H(Y, 0), X, b), Z != null)
      for (W = Z.length;W--; )
        n(Z[W]);
    X || (W = "value", J == "progress" && q == null ? O.removeAttribute("value") : q != null && (q !== O[W] || J == "progress" && !q || J == "option" && q != C[W]) && T(O, W, q, C[W], V), W = "checked", N != null && N != O[W] && T(O, W, N, C[W], V));
  }
  return O;
}
function Y0(O, E, Y) {
  try {
    if (typeof O == "function") {
      var g = typeof O.__u == "function";
      g && O.__u(), g && E == null || (O.__u = O(E));
    } else
      O.current = E;
  } catch (V) {
    L.__e(V, Y);
  }
}
function S0(O, E, Y) {
  var g, V;
  if (L.unmount && L.unmount(O), (g = O.ref) && (g.current && g.current != O.__e || Y0(g, null, E)), (g = O.__c) != null) {
    if (g.componentWillUnmount)
      try {
        g.componentWillUnmount();
      } catch (Z) {
        L.__e(Z, E);
      }
    g.base = g.__P = null;
  }
  if (g = O.__k)
    for (V = 0;V < g.length; V++)
      g[V] && S0(g[V], E, Y || typeof O.type != "function");
  Y || n(O.__e), O.__c = O.__ = O.__e = undefined;
}
function a0(O, E, Y) {
  return this.constructor(O, Y);
}
function c0(O, E, Y) {
  var g, V, Z, Q;
  E == document && (E = document.documentElement), L.__ && L.__(O, E), V = (g = typeof Y == "function") ? null : Y && Y.__k || E.__k, Z = [], Q = [], O0(E, O = (!g && Y || E).__k = E0(a, null, [O]), V || y, y, E.namespaceURI, !g && Y ? [Y] : V ? null : E.firstChild ? f.call(E.childNodes) : null, Z, !g && Y ? Y : V ? V.__e : E.firstChild, g, Q), C0(Z, O, Q);
}
f = k.slice, L = { __e: function(O, E, Y, g) {
  for (var V, Z, Q;E = E.__; )
    if ((V = E.__c) && !V.__)
      try {
        if ((Z = V.constructor) && Z.getDerivedStateFromError != null && (V.setState(Z.getDerivedStateFromError(O)), Q = V.__d), V.componentDidCatch != null && (V.componentDidCatch(O, g || {}), Q = V.__d), Q)
          return V.__E = V;
      } catch (X) {
        O = X;
      }
  throw O;
} }, R0 = 0, x0 = function(O) {
  return O != null && O.constructor === undefined;
}, _.prototype.setState = function(O, E) {
  var Y;
  Y = this.__s != null && this.__s != this.state ? this.__s : this.__s = U({}, this.state), typeof O == "function" && (O = O(U({}, Y), this.props)), O && U(Y, O), O != null && this.__v && (E && this._sb.push(E), t(this));
}, _.prototype.forceUpdate = function(O) {
  this.__v && (this.__e = true, O && this.__h.push(O), t(this));
}, _.prototype.render = a, I = [], D0 = typeof Promise == "function" ? Promise.prototype.then.bind(Promise.resolve()) : setTimeout, F0 = function(O, E) {
  return O.__v.__b - E.__v.__b;
}, p.__r = 0, r = Math.random().toString(8), m = "__d" + r, v = "__a" + r, J0 = /(PointerCapture)$|Capture$/i, u = 0, o = K0(false), i = K0(true), L0 = 0;
var M;
var d;
var Z0;
var U0;
var A = 0;
var s0 = [];
var j = L;
var B0 = j.__b;
var I0 = j.__r;
var M0 = j.diffed;
var P0 = j.__c;
var z0 = j.unmount;
var H0 = j.__;
function s(O, E) {
  j.__h && j.__h(d, O, A || E), A = 0;
  var Y = d.__H || (d.__H = { __: [], __h: [] });
  return O >= Y.__.length && Y.__.push({}), Y.__[O];
}
function w0(O) {
  return A = 1, v0($0, O);
}
function v0(O, E, Y) {
  var g = s(M++, 2);
  if (g.t = O, !g.__c && (g.__ = [Y ? Y(E) : $0(undefined, E), function(X) {
    var b = g.__N ? g.__N[0] : g.__[0], W = g.t(b, X);
    b !== W && (g.__N = [W, g.__[1]], g.__c.setState({}));
  }], g.__c = d, !d.__f)) {
    var V = function(X, b, W) {
      if (!g.__c.__H)
        return true;
      var R = g.__c.__H.__.filter(function(K) {
        return K.__c;
      });
      if (R.every(function(K) {
        return !K.__N;
      }))
        return !Z || Z.call(this, X, b, W);
      var G = g.__c.props !== X;
      return R.some(function(K) {
        if (K.__N) {
          var D = K.__[0];
          K.__ = K.__N, K.__N = undefined, D !== K.__[0] && (G = true);
        }
      }), Z && Z.call(this, X, b, W) || G;
    };
    d.__f = true;
    var { shouldComponentUpdate: Z, componentWillUpdate: Q } = d;
    d.componentWillUpdate = function(X, b, W) {
      if (this.__e) {
        var R = Z;
        Z = undefined, V(X, b, W), Z = R;
      }
      Q && Q.call(this, X, b, W);
    }, d.shouldComponentUpdate = V;
  }
  return g.__N || g.__;
}
function r0(O, E) {
  var Y = s(M++, 3);
  !j.__s && Q0(Y.__H, E) && (Y.__ = O, Y.u = E, d.__H.__h.push(Y));
}
function _0(O, E) {
  var Y = s(M++, 4);
  !j.__s && Q0(Y.__H, E) && (Y.__ = O, Y.u = E, d.__h.push(Y));
}
function o0(O) {
  return A = 5, G0(function() {
    return { current: O };
  }, []);
}
function G0(O, E) {
  var Y = s(M++, 7);
  return Q0(Y.__H, E) && (Y.__ = O(), Y.__H = E, Y.__h = O), Y.__;
}
function t0(O, E) {
  return A = 8, G0(function() {
    return O;
  }, E);
}
function E1() {
  for (var O;O = s0.shift(); ) {
    var E = O.__H;
    if (O.__P && E)
      try {
        E.__h.some(c), E.__h.some(g0), E.__h = [];
      } catch (Y) {
        E.__h = [], j.__e(Y, O.__v);
      }
  }
}
j.__b = function(O) {
  d = null, B0 && B0(O);
}, j.__ = function(O, E) {
  O && E.__k && E.__k.__m && (O.__m = E.__k.__m), H0 && H0(O, E);
}, j.__r = function(O) {
  I0 && I0(O), M = 0;
  var E = (d = O.__c).__H;
  E && (Z0 === d ? (E.__h = [], d.__h = [], E.__.some(function(Y) {
    Y.__N && (Y.__ = Y.__N), Y.u = Y.__N = undefined;
  })) : (E.__h.some(c), E.__h.some(g0), E.__h = [], M = 0)), Z0 = d;
}, j.diffed = function(O) {
  M0 && M0(O);
  var E = O.__c;
  E && E.__H && (E.__H.__h.length && (s0.push(E) !== 1 && U0 === j.requestAnimationFrame || ((U0 = j.requestAnimationFrame) || O1)(E1)), E.__H.__.some(function(Y) {
    Y.u && (Y.__H = Y.u), Y.u = undefined;
  })), Z0 = d = null;
}, j.__c = function(O, E) {
  E.some(function(Y) {
    try {
      Y.__h.some(c), Y.__h = Y.__h.filter(function(g) {
        return !g.__ || g0(g);
      });
    } catch (g) {
      E.some(function(V) {
        V.__h && (V.__h = []);
      }), E = [], j.__e(g, Y.__v);
    }
  }), P0 && P0(O, E);
}, j.unmount = function(O) {
  z0 && z0(O);
  var E, Y = O.__c;
  Y && Y.__H && (Y.__H.__.some(function(g) {
    try {
      c(g);
    } catch (V) {
      E = V;
    }
  }), Y.__H = undefined, E && j.__e(E, Y.__v));
};
var A0 = typeof requestAnimationFrame == "function";
function O1(O) {
  var E, Y = function() {
    clearTimeout(g), A0 && cancelAnimationFrame(E), setTimeout(O);
  }, g = setTimeout(Y, 35);
  A0 && (E = requestAnimationFrame(Y));
}
function c(O) {
  var E = d, Y = O.__c;
  typeof Y == "function" && (O.__c = undefined, Y()), d = E;
}
function g0(O) {
  var E = d;
  O.__c = O.__(), d = E;
}
function Q0(O, E) {
  return !O || O.length !== E.length || E.some(function(Y, g) {
    return Y !== O[g];
  });
}
function $0(O, E) {
  return typeof E == "function" ? E(O) : E;
}
var m0 = function(O, E, Y, g) {
  var V;
  E[0] = 0;
  for (var Z = 1;Z < E.length; Z++) {
    var Q = E[Z++], X = E[Z] ? (E[0] |= Q ? 1 : 2, Y[E[Z++]]) : E[++Z];
    Q === 3 ? g[0] = X : Q === 4 ? g[1] = Object.assign(g[1] || {}, X) : Q === 5 ? (g[1] = g[1] || {})[E[++Z]] = X : Q === 6 ? g[1][E[++Z]] += X + "" : Q ? (V = O.apply(X, m0(O, X, Y, ["", null])), g.push(V), X[0] ? E[0] |= 2 : (E[Z - 2] = 0, E[Z] = V)) : g.push(X);
  }
  return g;
};
var T0 = new Map;
function V0(O) {
  var E = T0.get(this);
  return E || (E = new Map, T0.set(this, E)), (E = m0(this, E.get(O) || (E.set(O, E = function(Y) {
    for (var g, V, Z = 1, Q = "", X = "", b = [0], W = function(K) {
      Z === 1 && (K || (Q = Q.replace(/^\s*\n\s*|\s*\n\s*$/g, ""))) ? b.push(0, K, Q) : Z === 3 && (K || Q) ? (b.push(3, K, Q), Z = 2) : Z === 2 && Q === "..." && K ? b.push(4, K, 0) : Z === 2 && Q && !K ? b.push(5, 0, true, Q) : Z >= 5 && ((Q || !K && Z === 5) && (b.push(Z, 0, Q, V), Z = 6), K && (b.push(Z, K, 0, V), Z = 6)), Q = "";
    }, R = 0;R < Y.length; R++) {
      R && (Z === 1 && W(), W(R));
      for (var G = 0;G < Y[R].length; G++)
        g = Y[R][G], Z === 1 ? g === "<" ? (W(), b = [b], Z = 3) : Q += g : Z === 4 ? Q === "--" && g === ">" ? (Z = 1, Q = "") : Q = g + Q[0] : X ? g === X ? X = "" : Q += g : g === '"' || g === "'" ? X = g : g === ">" ? (W(), Z = 1) : Z && (g === "=" ? (Z = 5, V = Q, Q = "") : g === "/" && (Z < 5 || Y[R][G + 1] === ">") ? (W(), Z === 3 && (b = b[0]), Z = b, (b = b[0]).push(2, 0, Z), Z = 0) : g === " " || g === "\t" || g === `
` || g === "\r" ? (W(), Z = 2) : Q += g), Z === 3 && Q === "!--" && (Z = 4, b = b[0]);
    }
    return W(), b;
  }(O)), E), arguments, [])).length > 1 ? E : E[0];
}
var X1 = V0.bind(E0);

// web/src/utils/storage.ts
function getLocalStorageItem(key) {
  if (typeof window === "undefined" || !window.localStorage)
    return null;
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}
function setLocalStorageItem(key, value) {
  if (typeof window === "undefined" || !window.localStorage)
    return;
  try {
    window.localStorage.setItem(key, value);
  } catch {
    return;
  }
}
function getLocalStorageBoolean(key, defaultValue = false) {
  const raw = getLocalStorageItem(key);
  if (raw === null)
    return defaultValue;
  return raw === "true";
}
function getLocalStorageNumber(key, defaultValue = null) {
  const raw = getLocalStorageItem(key);
  if (raw === null)
    return defaultValue;
  const parsed = parseInt(raw, 10);
  return Number.isFinite(parsed) ? parsed : defaultValue;
}

// web/src/ui/timeline-utils.ts
var dedupePosts = (items) => {
  const seen = new Set;
  return (items || []).filter((post) => {
    if (!post || seen.has(post.id))
      return false;
    seen.add(post.id);
    return true;
  });
};

// web/src/ui/app-realtime-timeline.ts
function appendUniqueTimelinePost(posts, nextPost) {
  if (!Array.isArray(posts) || posts.length === 0) {
    return [nextPost];
  }
  if (posts.some((post) => post?.id === nextPost?.id)) {
    return posts;
  }
  return [...posts, nextPost];
}

// web/src/ui/use-agent-state.ts
function useAgentState() {
  const [agentStatus, setAgentStatus] = w0(null);
  const [agentDraft, setAgentDraft] = w0({ text: "", totalLines: 0 });
  const [agentPlan, setAgentPlan] = w0("");
  const [agentThought, setAgentThought] = w0({ text: "", totalLines: 0 });
  const [pendingRequest, setPendingRequest] = w0(null);
  const [currentTurnId, setCurrentTurnId] = w0(null);
  const [steerQueuedTurnId, setSteerQueuedTurnId] = w0(null);
  const lastAgentEventRef = o0(null);
  const lastSilenceNoticeRef = o0(0);
  const isAgentRunningRef = o0(false);
  const draftBufferRef = o0("");
  const thoughtBufferRef = o0("");
  const previewResyncPendingRef = o0(false);
  const previewResyncGenerationRef = o0(0);
  const pendingRequestRef = o0(null);
  const stalledPostIdRef = o0(null);
  const currentTurnIdRef = o0(null);
  const steerQueuedTurnIdRef = o0(null);
  const thoughtExpandedRef = o0(false);
  const draftExpandedRef = o0(false);
  return {
    agentStatus,
    setAgentStatus,
    agentDraft,
    setAgentDraft,
    agentPlan,
    setAgentPlan,
    agentThought,
    setAgentThought,
    pendingRequest,
    setPendingRequest,
    currentTurnId,
    setCurrentTurnId,
    steerQueuedTurnId,
    setSteerQueuedTurnId,
    lastAgentEventRef,
    lastSilenceNoticeRef,
    isAgentRunningRef,
    draftBufferRef,
    thoughtBufferRef,
    previewResyncPendingRef,
    previewResyncGenerationRef,
    pendingRequestRef,
    stalledPostIdRef,
    currentTurnIdRef,
    steerQueuedTurnIdRef,
    thoughtExpandedRef,
    draftExpandedRef
  };
}

// web/src/api.ts
var API_BASE = "";
async function request(url, options = {}) {
  const startedAt = typeof performance !== "undefined" && typeof performance.now === "function" ? performance.now() : Date.now();
  let response;
  try {
    response = await fetch(API_BASE + url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers || {}
      }
    });
  } catch (error) {
    recordAppPerfRequest({
      method: String(options.method || "GET").toUpperCase(),
      url,
      startedAt,
      durationMs: performance.now() - startedAt,
      ok: false,
      detail: { failedBeforeResponse: true }
    });
    throw error;
  }
  const durationMs = performance.now() - startedAt;
  recordAppPerfRequest({
    method: String(options.method || "GET").toUpperCase(),
    url,
    startedAt,
    durationMs,
    status: response.status,
    ok: response.ok,
    requestId: response.headers?.get?.("x-request-id") || null,
    serverTiming: response.headers?.get?.("Server-Timing") || null
  });
  if (!response.ok) {
    const err = await response.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(err.error || `HTTP ${response.status}`);
  }
  return response.json();
}
async function getTimeline(limit = 50, beforeId = null, chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    return { posts: [] };
  let url = `/api/sessions/${encodeURIComponent(sessionId)}/messages?limit=${limit}`;
  if (beforeId)
    url += `&before=${beforeId}`;
  const data = await request(url);
  const messages = data.messages || [];
  return {
    posts: messages.map((m2) => ({
      id: m2.id,
      chat_jid: chatJid,
      content: m2.content,
      timestamp: m2.created_at,
      sender: m2.role === "user" ? "user" : "agent",
      is_from_me: m2.role === "user",
      is_bot_message: m2.role === "assistant",
      data: {
        type: m2.role === "assistant" ? "agent_response" : "user_message",
        content: m2.content,
        thread_id: null,
        agent_id: m2.role === "assistant" ? "gi" : null,
        content_blocks: m2.payload?.content_blocks || null,
        content_meta: null,
        link_previews: null,
        kind: m2.payload?.kind || null,
        source: m2.payload?.source || null,
        clipped: m2.payload?.clipped || false
      }
    }))
  };
}
async function getSystemMetrics() {
  return request("/api/system-metrics").catch(() => null);
}
async function getAgentModels(_chatJid = null) {
  const data = await request("/api/runtime/config").catch(() => ({}));
  const models = (data.enabled_models || []).map((id) => ({
    id,
    provider: data.default_provider || "",
    label: id
  }));
  return { models, current: data.default_model || "" };
}
async function sendAgentMessage(agentId, content, _threadId = null, _mediaIds = [], mode = null, chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    throw new Error("No active session");
  const intent = mode === "steer" ? "steer" : mode === "queue" ? "queue" : "prompt";
  return request(`/api/sessions/${encodeURIComponent(sessionId)}/prompt`, {
    method: "POST",
    body: JSON.stringify({ prompt: content, intent })
  });
}
async function uploadMedia(_file, _chatJid = null) {
  return null;
}
async function getMediaInfo(mediaId) {
  return request(`/api/media/${mediaId}`).catch(() => null);
}
function getMediaUrl(mediaId) {
  return `/api/media/${mediaId}/raw`;
}
function getThumbnailUrl(mediaId) {
  return `/api/media/${mediaId}/thumbnail`;
}
async function submitAdaptiveCardAction(_payload) {
  return null;
}
async function getWorkspaceTree(_chatJid = null) {
  return request("/api/workspace/tree");
}
async function getWorkspaceFile(path, _chatJid = null) {
  return request(`/api/workspace/file?path=${encodeURIComponent(path)}`);
}
async function getWorkspaceIndexStatus(_chatJid = null) {
  return { status: "ready", indexed_at: null };
}
async function reindexWorkspace(_chatJid = null) {
  return null;
}
async function createWorkspaceFile(path, content, _chatJid = null) {
  return request("/api/workspace/file", { method: "POST", body: JSON.stringify({ path, content }) }).catch(() => null);
}
async function renameWorkspaceFile(_oldPath, _newPath, _chatJid = null) {
  return null;
}
async function moveWorkspaceEntry(_from, _to, _chatJid = null) {
  return null;
}
async function deleteWorkspaceFile(_path, _chatJid = null) {
  return null;
}
async function uploadWorkspaceFile(_path, _file, _chatJid = null) {
  return null;
}
async function setWorkspaceVisibility(_path, _hidden, _chatJid = null) {
  return null;
}
function getWorkspaceDownloadUrl(path) {
  return `/api/workspace/file?path=${encodeURIComponent(path)}`;
}
async function getWorkspaceBranch(_chatJid = null) {
  return null;
}
function getWorkspaceRawUrl(path, options = {}) {
  const q = new URLSearchParams({ path: String(path || "") });
  if (options?.download)
    q.set("download", "1");
  return `/api/workspace/raw?${q.toString()}`;
}
function getWorkspaceFileDownloadUrl(path) {
  return getWorkspaceRawUrl(path, { download: true });
}
async function recordAppPerfRequest(_payload) {}

class SSEClient {
  onEvent;
  onStatusChange;
  chatJid;
  eventSource;
  reconnectTimeout;
  reconnectDelay;
  status;
  connecting;
  staleMonitor;
  constructor(onEvent, onStatusChange, options = {}) {
    this.onEvent = onEvent;
    this.onStatusChange = onStatusChange;
    this.chatJid = typeof options?.chatJid === "string" && options.chatJid.trim() ? options.chatJid.trim() : null;
    this.eventSource = null;
    this.reconnectTimeout = null;
    this.reconnectDelay = 1000;
    this.status = "disconnected";
    this.connecting = false;
    this.staleMonitor = null;
  }
  connect() {
    if (this.connecting)
      return;
    if (this.eventSource && this.status === "connected")
      return;
    this.connecting = true;
    if (this.eventSource)
      this.eventSource.close();
    this.clearStaleMonitor();
    const query = this.chatJid ? `?chat_jid=${encodeURIComponent(this.chatJid)}` : "";
    this.eventSource = new EventSource(API_BASE + "/sse/stream" + query);
    const bindJsonEvent = (eventType) => {
      this.eventSource.addEventListener(eventType, (e2) => {
        this.resetStaleMonitor();
        try {
          const data = JSON.parse(e2.data);
          this.onEvent(eventType, data);
        } catch {}
      });
    };
    this.eventSource.addEventListener("connected", () => {
      this.connecting = false;
      this.reconnectDelay = 1000;
      this.setStatus("connected");
      this.resetStaleMonitor();
    });
    this.eventSource.addEventListener("heartbeat", () => {
      this.resetStaleMonitor();
    });
    bindJsonEvent("new_post");
    bindJsonEvent("new_reply");
    bindJsonEvent("agent_response");
    bindJsonEvent("interaction_updated");
    bindJsonEvent("interaction_deleted");
    bindJsonEvent("agent_status");
    bindJsonEvent("agent_steer_queued");
    bindJsonEvent("agent_followup_queued");
    bindJsonEvent("agent_followup_consumed");
    bindJsonEvent("agent_followup_removed");
    bindJsonEvent("workspace_update");
    bindJsonEvent("agent_draft");
    bindJsonEvent("agent_draft_delta");
    bindJsonEvent("agent_thought");
    bindJsonEvent("agent_thought_delta");
    bindJsonEvent("model_changed");
    bindJsonEvent("ui_theme");
    bindJsonEvent("ui_meters");
    this.eventSource.onerror = () => {
      this.connecting = false;
      this.setStatus("disconnected");
      this.scheduleReconnect();
    };
  }
  disconnect() {
    this.clearStaleMonitor();
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    this.setStatus("disconnected");
  }
  reconnectIfNeeded() {
    if (this.status !== "connected")
      this.connect();
  }
  forceReconnect() {
    this.disconnect();
    this.connect();
  }
  setStatus(status) {
    if (this.status === status)
      return;
    this.status = status;
    this.onStatusChange?.(status);
  }
  scheduleReconnect() {
    if (this.reconnectTimeout)
      return;
    this.reconnectTimeout = setTimeout(() => {
      this.reconnectTimeout = null;
      this.reconnectDelay = Math.min(this.reconnectDelay * 1.5, 30000);
      this.connect();
    }, this.reconnectDelay);
  }
  resetStaleMonitor() {
    this.clearStaleMonitor();
    this.staleMonitor = setTimeout(() => {
      this.setStatus("stale");
      this.forceReconnect();
    }, 60000);
  }
  clearStaleMonitor() {
    if (this.staleMonitor) {
      clearTimeout(this.staleMonitor);
      this.staleMonitor = null;
    }
  }
}
async function getMediaBlob(..._args) {
  return null;
}

// web/src/ui/app-helpers.ts
function readSilenceOverride(key, fallback) {
  try {
    if (typeof window === "undefined")
      return fallback;
    const overrides = window.__PICLAW_SILENCE || {};
    const directKey = `__PICLAW_SILENCE_${key.toUpperCase()}_MS`;
    const raw = overrides[key] ?? window[directKey];
    const value = Number(raw);
    return Number.isFinite(value) ? value : fallback;
  } catch {
    return fallback;
  }
}
var SILENCE_WARNING_MS = readSilenceOverride("warning", 30000);
var SILENCE_FINALIZE_MS = readSilenceOverride("finalize", 120000);
var SILENCE_REFRESH_MS = readSilenceOverride("refresh", 30000);
function isIOSDevice() {
  if (/iPad|iPhone/.test(navigator.userAgent)) {
    return true;
  }
  return navigator.platform === "MacIntel" && navigator.maxTouchPoints > 1;
}

// web/src/ui/use-sse-connection.ts
function bindSseWakeLifecycle({ sse, onWake }, runtime = {}) {
  const win = runtime.window ?? (typeof window !== "undefined" ? window : null);
  const doc = runtime.document ?? (typeof document !== "undefined" ? document : null);
  if (!win || !doc || !sse) {
    return () => {};
  }
  const reconnectAfterReturn = () => {
    if (typeof sse.forceReconnect === "function") {
      sse.forceReconnect();
      return;
    }
    sse.reconnectIfNeeded();
  };
  const shouldUseFocusReconnect = typeof runtime.useFocusReconnect === "boolean" ? runtime.useFocusReconnect : !isIOSDevice();
  let pendingWake = doc.visibilityState && doc.visibilityState !== "visible";
  const handleHiddenState = () => {
    if (doc.visibilityState && doc.visibilityState !== "visible") {
      pendingWake = true;
      return true;
    }
    return false;
  };
  const handleVisibleReturn = () => {
    if (handleHiddenState())
      return;
    if (pendingWake) {
      pendingWake = false;
      reconnectAfterReturn();
      onWake?.();
    }
  };
  const handleWindowFocus = () => {
    if (handleHiddenState())
      return;
    if (pendingWake) {
      handleVisibleReturn();
      return;
    }
    if (shouldUseFocusReconnect) {
      sse.reconnectIfNeeded();
    }
  };
  const handlePageShow = () => {
    handleVisibleReturn();
  };
  const handleVisibilityChange = () => {
    handleVisibleReturn();
  };
  win.addEventListener("focus", handleWindowFocus);
  win.addEventListener("pageshow", handlePageShow);
  doc.addEventListener("visibilitychange", handleVisibilityChange);
  return () => {
    win.removeEventListener("focus", handleWindowFocus);
    win.removeEventListener("pageshow", handlePageShow);
    doc.removeEventListener("visibilitychange", handleVisibilityChange);
  };
}
function useSseConnection({ handleSseEvent, handleConnectionStatusChange, loadPosts, onWake, chatJid }) {
  const sseEventRef = o0(handleSseEvent);
  sseEventRef.current = handleSseEvent;
  const statusChangeRef = o0(handleConnectionStatusChange);
  statusChangeRef.current = handleConnectionStatusChange;
  const loadPostsRef = o0(loadPosts);
  loadPostsRef.current = loadPosts;
  const onWakeRef = o0(onWake);
  onWakeRef.current = onWake;
  r0(() => {
    const sse = new SSEClient((type, data) => sseEventRef.current(type, data), (status) => statusChangeRef.current(status), { chatJid });
    sse.connect();
    const disposeWakeLifecycle = bindSseWakeLifecycle({
      sse,
      onWake: () => onWakeRef.current?.()
    });
    return () => {
      disposeWakeLifecycle();
      sse.disconnect();
    };
  }, [chatJid]);
}

// web/src/ui/theme.ts
var THEME_STORAGE_KEY = "piclaw_theme";
var TINT_STORAGE_KEY = "piclaw_tint";
var CHAT_THEMES_STORAGE_KEY = "piclaw_chat_themes";
var DEFAULT_LIGHT = {
  bgPrimary: "#ffffff",
  bgSecondary: "#f7f9fa",
  bgHover: "#e8ebed",
  textPrimary: "#0f1419",
  textSecondary: "#536471",
  borderColor: "#eff3f4",
  accent: "#1d9bf0",
  accentHover: "#1a8cd8",
  warning: "#f0b429",
  danger: "#f4212e",
  success: "#00ba7c"
};
var DEFAULT_DARK = {
  bgPrimary: "#000000",
  bgSecondary: "#16181c",
  bgHover: "#1d1f23",
  textPrimary: "#e7e9ea",
  textSecondary: "#71767b",
  borderColor: "#2f3336",
  accent: "#1d9bf0",
  accentHover: "#1a8cd8",
  warning: "#f0b429",
  danger: "#f4212e",
  success: "#00ba7c"
};
var THEME_PRESETS = {
  default: {
    label: "Default",
    mode: "auto",
    light: DEFAULT_LIGHT,
    dark: DEFAULT_DARK
  },
  tango: {
    label: "Tango",
    mode: "light",
    light: {
      bgPrimary: "#f6f5f4",
      bgSecondary: "#efedeb",
      bgHover: "#e5e3e1",
      textPrimary: "#2e3436",
      textSecondary: "#5c6466",
      borderColor: "#d3d7cf",
      accent: "#3465a4",
      accentHover: "#2c5890",
      danger: "#cc0000",
      success: "#4e9a06"
    }
  },
  xterm: {
    label: "XTerm",
    mode: "dark",
    dark: {
      bgPrimary: "#000000",
      bgSecondary: "#0a0a0a",
      bgHover: "#121212",
      textPrimary: "#d0d0d0",
      textSecondary: "#8a8a8a",
      borderColor: "#1f1f1f",
      accent: "#00a2ff",
      accentHover: "#0086d1",
      danger: "#ff5f5f",
      success: "#5fff87"
    }
  },
  monokai: {
    label: "Monokai",
    mode: "dark",
    dark: {
      bgPrimary: "#272822",
      bgSecondary: "#2f2f2f",
      bgHover: "#3a3a3a",
      textPrimary: "#f8f8f2",
      textSecondary: "#cfcfc2",
      borderColor: "#3e3d32",
      accent: "#f92672",
      accentHover: "#e81560",
      danger: "#f92672",
      success: "#a6e22e"
    }
  },
  "monokai-pro": {
    label: "Monokai Pro",
    mode: "dark",
    dark: {
      bgPrimary: "#2d2a2e",
      bgSecondary: "#363237",
      bgHover: "#403a40",
      textPrimary: "#fcfcfa",
      textSecondary: "#c1c0c0",
      borderColor: "#444046",
      accent: "#ff6188",
      accentHover: "#f74f7e",
      danger: "#ff4f5e",
      success: "#a9dc76"
    }
  },
  ristretto: {
    label: "Ristretto",
    mode: "dark",
    dark: {
      bgPrimary: "#2c2525",
      bgSecondary: "#362d2d",
      bgHover: "#403535",
      textPrimary: "#f4f1ef",
      textSecondary: "#cbbdb8",
      borderColor: "#4a3c3c",
      accent: "#ff9f43",
      accentHover: "#f28a2e",
      danger: "#ff5f56",
      success: "#a9dc76"
    }
  },
  dracula: {
    label: "Dracula",
    mode: "dark",
    dark: {
      bgPrimary: "#282a36",
      bgSecondary: "#303445",
      bgHover: "#3a3f52",
      textPrimary: "#f8f8f2",
      textSecondary: "#c5c8d6",
      borderColor: "#44475a",
      accent: "#bd93f9",
      accentHover: "#a87ded",
      danger: "#ff5555",
      success: "#50fa7b"
    }
  },
  catppuccin: {
    label: "Catppuccin",
    mode: "dark",
    dark: {
      bgPrimary: "#1e1e2e",
      bgSecondary: "#24273a",
      bgHover: "#2c2f41",
      textPrimary: "#cdd6f4",
      textSecondary: "#a6adc8",
      borderColor: "#313244",
      accent: "#89b4fa",
      accentHover: "#74a0f5",
      danger: "#f38ba8",
      success: "#a6e3a1"
    }
  },
  nord: {
    label: "Nord",
    mode: "dark",
    dark: {
      bgPrimary: "#2e3440",
      bgSecondary: "#3b4252",
      bgHover: "#434c5e",
      textPrimary: "#eceff4",
      textSecondary: "#d8dee9",
      borderColor: "#4c566a",
      accent: "#88c0d0",
      accentHover: "#78a9c0",
      danger: "#bf616a",
      success: "#a3be8c"
    }
  },
  gruvbox: {
    label: "Gruvbox",
    mode: "dark",
    dark: {
      bgPrimary: "#282828",
      bgSecondary: "#32302f",
      bgHover: "#3c3836",
      textPrimary: "#ebdbb2",
      textSecondary: "#bdae93",
      borderColor: "#3c3836",
      accent: "#d79921",
      accentHover: "#c28515",
      danger: "#fb4934",
      success: "#b8bb26"
    }
  },
  solarized: {
    label: "Solarized",
    mode: "auto",
    light: {
      bgPrimary: "#fdf6e3",
      bgSecondary: "#f5efdc",
      bgHover: "#eee8d5",
      textPrimary: "#586e75",
      textSecondary: "#657b83",
      borderColor: "#e0d8c6",
      accent: "#268bd2",
      accentHover: "#1f78b3",
      danger: "#dc322f",
      success: "#859900"
    },
    dark: {
      bgPrimary: "#002b36",
      bgSecondary: "#073642",
      bgHover: "#0b3c4a",
      textPrimary: "#eee8d5",
      textSecondary: "#93a1a1",
      borderColor: "#18424a",
      accent: "#268bd2",
      accentHover: "#1f78b3",
      danger: "#dc322f",
      success: "#859900"
    }
  },
  tokyo: {
    label: "Tokyo",
    mode: "dark",
    dark: {
      bgPrimary: "#1a1b26",
      bgSecondary: "#24283b",
      bgHover: "#2f3549",
      textPrimary: "#c0caf5",
      textSecondary: "#9aa5ce",
      borderColor: "#414868",
      accent: "#7aa2f7",
      accentHover: "#6b92e6",
      danger: "#f7768e",
      success: "#9ece6a"
    }
  },
  miasma: {
    label: "Miasma",
    mode: "dark",
    dark: {
      bgPrimary: "#1f1f23",
      bgSecondary: "#29292f",
      bgHover: "#33333a",
      textPrimary: "#e5e5e5",
      textSecondary: "#b4b4b4",
      borderColor: "#3d3d45",
      accent: "#c9739c",
      accentHover: "#b8618c",
      danger: "#e06c75",
      success: "#98c379"
    }
  },
  github: {
    label: "GitHub",
    mode: "auto",
    light: {
      bgPrimary: "#ffffff",
      bgSecondary: "#f6f8fa",
      bgHover: "#eaeef2",
      textPrimary: "#24292f",
      textSecondary: "#57606a",
      borderColor: "#d0d7de",
      accent: "#0969da",
      accentHover: "#0550ae",
      danger: "#cf222e",
      success: "#1a7f37"
    },
    dark: {
      bgPrimary: "#0d1117",
      bgSecondary: "#161b22",
      bgHover: "#21262d",
      textPrimary: "#c9d1d9",
      textSecondary: "#8b949e",
      borderColor: "#30363d",
      accent: "#2f81f7",
      accentHover: "#1f6feb",
      danger: "#f85149",
      success: "#3fb950"
    }
  },
  gotham: {
    label: "Gotham",
    mode: "dark",
    dark: {
      bgPrimary: "#0b0f14",
      bgSecondary: "#111720",
      bgHover: "#18212b",
      textPrimary: "#cbd6e2",
      textSecondary: "#9bb0c3",
      borderColor: "#1f2a37",
      accent: "#5ccfe6",
      accentHover: "#48b8ce",
      danger: "#d26937",
      success: "#2aa889"
    }
  }
};
var THEME_VAR_KEYS = [
  "--bg-primary",
  "--bg-secondary",
  "--bg-hover",
  "--text-primary",
  "--text-secondary",
  "--border-color",
  "--accent-color",
  "--accent-hover",
  "--accent-color-alpha",
  "--accent-contrast-text",
  "--accent-soft",
  "--accent-soft-strong",
  "--warning-color",
  "--danger-color",
  "--success-color",
  "--search-highlight-color"
];
var currentTheme = {
  theme: "default",
  tint: null
};
var currentMode = "light";
var mediaListenerAttached = false;
function normalizeThemeName(value) {
  const raw = String(value || "").trim().toLowerCase();
  if (!raw)
    return "default";
  if (raw === "solarized-dark" || raw === "solarized-light")
    return "solarized";
  if (raw === "github-dark" || raw === "github-light")
    return "github";
  if (raw === "tokyo-night")
    return "tokyo";
  return raw;
}
function parseHexColor(input) {
  if (!input)
    return null;
  const raw = String(input).trim();
  if (!raw)
    return null;
  const hex = raw.startsWith("#") ? raw.slice(1) : raw;
  if (!/^[0-9a-fA-F]{3}$/.test(hex) && !/^[0-9a-fA-F]{6}$/.test(hex))
    return null;
  const full = hex.length === 3 ? hex.split("").map((c2) => c2 + c2).join("") : hex;
  const int = parseInt(full, 16);
  return {
    r: int >> 16 & 255,
    g: int >> 8 & 255,
    b: int & 255,
    hex: `#${full.toLowerCase()}`
  };
}
function resolveComputedCssColor(el, fallbackColor) {
  try {
    if (document.body) {
      el.style.display = "none";
      document.body.appendChild(el);
      const computed = getComputedStyle(el).color || el.style.color;
      document.body.removeChild(el);
      return computed;
    }
  } catch {
    return fallbackColor;
  }
  return fallbackColor;
}
function parseCssColor(input) {
  if (!input || typeof document === "undefined")
    return null;
  const raw = String(input).trim();
  if (!raw)
    return null;
  const el = document.createElement("div");
  el.style.color = "";
  el.style.color = raw;
  if (!el.style.color)
    return null;
  const computed = resolveComputedCssColor(el, el.style.color);
  const match = computed.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/i);
  if (!match)
    return null;
  const r2 = parseInt(match[1], 10);
  const g = parseInt(match[2], 10);
  const b = parseInt(match[3], 10);
  if (![r2, g, b].every((v2) => Number.isFinite(v2)))
    return null;
  const hex = `#${[r2, g, b].map((v2) => v2.toString(16).padStart(2, "0")).join("")}`;
  return { r: r2, g, b, hex };
}
function parseColor(input) {
  return parseHexColor(input) || parseCssColor(input);
}
function mixColors(base, overlay, ratio) {
  const r2 = Math.round(base.r + (overlay.r - base.r) * ratio);
  const g = Math.round(base.g + (overlay.g - base.g) * ratio);
  const b = Math.round(base.b + (overlay.b - base.b) * ratio);
  return `rgb(${r2} ${g} ${b})`;
}
function rgbaColor(color, alpha) {
  return `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`;
}
function relativeLuminance(c2) {
  const rs = c2.r / 255, gs = c2.g / 255, bs = c2.b / 255;
  const r2 = rs <= 0.03928 ? rs / 12.92 : Math.pow((rs + 0.055) / 1.055, 2.4);
  const g = gs <= 0.03928 ? gs / 12.92 : Math.pow((gs + 0.055) / 1.055, 2.4);
  const b = bs <= 0.03928 ? bs / 12.92 : Math.pow((bs + 0.055) / 1.055, 2.4);
  return 0.2126 * r2 + 0.7152 * g + 0.0722 * b;
}
function contrastTextColor(bg) {
  return relativeLuminance(bg) > 0.4 ? "#000000" : "#ffffff";
}
function resolveSystemMode() {
  if (typeof window === "undefined")
    return "light";
  try {
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  } catch {
    return "light";
  }
}
function resolvePreset(themeName) {
  return THEME_PRESETS[themeName] || THEME_PRESETS.default;
}
function resolveModeForPreset(preset) {
  return preset.mode === "auto" ? resolveSystemMode() : preset.mode;
}
function resolvePalette(themeName, mode) {
  const preset = resolvePreset(themeName);
  if (mode === "dark" && preset.dark)
    return preset.dark;
  if (mode === "light" && preset.light)
    return preset.light;
  return preset.dark || preset.light || DEFAULT_LIGHT;
}
function tintPaletteColor(value, tint, ratio) {
  const base = parseColor(value);
  if (!base)
    return value;
  return mixColors(base, tint, ratio);
}
function buildTintedPalette(basePalette, tintHex, mode) {
  const tint = parseColor(tintHex);
  if (!tint)
    return basePalette;
  const contrastColor = mode === "dark" ? "#ffffff" : "#000000";
  const contrast = parseHexColor(contrastColor);
  return {
    ...basePalette,
    bgPrimary: tintPaletteColor(basePalette.bgPrimary, tint, 0.08),
    bgSecondary: tintPaletteColor(basePalette.bgSecondary, tint, 0.12),
    bgHover: tintPaletteColor(basePalette.bgHover, tint, 0.16),
    textPrimary: tintPaletteColor(basePalette.textPrimary, tint, mode === "dark" ? 0.08 : 0.06),
    textSecondary: tintPaletteColor(basePalette.textSecondary, tint, mode === "dark" ? 0.12 : 0.1),
    borderColor: tintPaletteColor(basePalette.borderColor, tint, 0.1),
    accent: tint.hex,
    accentHover: contrast ? mixColors(tint, contrast, 0.18) : tint.hex,
    warning: tintPaletteColor(basePalette.warning || DEFAULT_LIGHT.warning, tint, 0.14),
    danger: tintPaletteColor(basePalette.danger, tint, 0.16),
    success: tintPaletteColor(basePalette.success, tint, 0.16)
  };
}
function resolveWarningColor(palette, mode) {
  const explicit = parseColor(palette?.warning);
  if (explicit)
    return explicit.hex;
  const defaultWarning = parseColor(mode === "dark" ? DEFAULT_DARK.warning : DEFAULT_LIGHT.warning) || parseColor(DEFAULT_LIGHT.warning);
  const accent = parseColor(palette?.accent);
  if (defaultWarning && accent) {
    return mixColors(defaultWarning, accent, mode === "dark" ? 0.18 : 0.14);
  }
  return mode === "dark" ? DEFAULT_DARK.warning : DEFAULT_LIGHT.warning;
}
function applyCssVariables(palette, mode) {
  if (typeof document === "undefined")
    return;
  const root = document.documentElement;
  const accentColor = palette.accent;
  const accentHex = parseColor(accentColor);
  const searchHighlight = accentHex ? rgbaColor(accentHex, mode === "dark" ? 0.35 : 0.2) : palette.searchHighlight || palette.searchHighlightColor;
  const accentSoft = accentHex ? rgbaColor(accentHex, mode === "dark" ? 0.16 : 0.12) : "rgba(29, 155, 240, 0.12)";
  const accentSoftStrong = accentHex ? rgbaColor(accentHex, mode === "dark" ? 0.28 : 0.2) : "rgba(29, 155, 240, 0.2)";
  const accentContrastText = accentHex ? contrastTextColor(accentHex) : mode === "dark" ? "#000000" : "#ffffff";
  const accentColorAlpha = accentHex ? rgbaColor(accentHex, mode === "dark" ? 0.35 : 0.25) : "rgba(29, 155, 240, 0.25)";
  const warningColor = resolveWarningColor(palette, mode);
  const vars = {
    "--bg-primary": palette.bgPrimary,
    "--bg-secondary": palette.bgSecondary,
    "--bg-hover": palette.bgHover,
    "--text-primary": palette.textPrimary,
    "--text-secondary": palette.textSecondary,
    "--border-color": palette.borderColor,
    "--accent-color": accentColor,
    "--accent-hover": palette.accentHover || accentColor,
    "--accent-color-alpha": accentColorAlpha,
    "--accent-soft": accentSoft,
    "--accent-soft-strong": accentSoftStrong,
    "--accent-contrast-text": accentContrastText,
    "--warning-color": warningColor,
    "--danger-color": palette.danger || DEFAULT_LIGHT.danger,
    "--success-color": palette.success || DEFAULT_LIGHT.success,
    "--search-highlight-color": searchHighlight || "rgba(29, 155, 240, 0.2)"
  };
  Object.entries(vars).forEach(([key, value]) => {
    if (value)
      root.style.setProperty(key, value);
  });
}
function clearCssVariables() {
  if (typeof document === "undefined")
    return;
  const root = document.documentElement;
  THEME_VAR_KEYS.forEach((key) => root.style.removeProperty(key));
}
function ensureMetaTag(name, options = {}) {
  if (typeof document === "undefined")
    return null;
  const id = typeof options.id === "string" && options.id.trim() ? options.id.trim() : null;
  let tag = id ? document.getElementById(id) : document.querySelector(`meta[name="${name}"]`);
  if (!tag) {
    tag = document.createElement("meta");
    document.head.appendChild(tag);
  }
  tag.setAttribute("name", name);
  if (id)
    tag.setAttribute("id", id);
  return tag;
}
function resolveThemeColorForMode(mode) {
  const themeName = normalizeThemeName(currentTheme?.theme || "default");
  const tint = currentTheme?.tint ? String(currentTheme.tint).trim() : null;
  let palette = resolvePalette(themeName, mode);
  if (themeName === "default" && tint) {
    palette = buildTintedPalette(palette, tint, mode);
  }
  if (palette?.bgPrimary)
    return palette.bgPrimary;
  return mode === "dark" ? DEFAULT_DARK.bgPrimary : DEFAULT_LIGHT.bgPrimary;
}
function updateMetaColor(color, mode) {
  if (typeof document === "undefined")
    return;
  const themeMeta = ensureMetaTag("theme-color", { id: "dynamic-theme-color" });
  if (themeMeta && color) {
    themeMeta.removeAttribute("media");
    themeMeta.setAttribute("content", color);
  }
  const lightThemeMeta = ensureMetaTag("theme-color", { id: "theme-color-light" });
  if (lightThemeMeta) {
    lightThemeMeta.setAttribute("media", "(prefers-color-scheme: light)");
    lightThemeMeta.setAttribute("content", resolveThemeColorForMode("light"));
  }
  const darkThemeMeta = ensureMetaTag("theme-color", { id: "theme-color-dark" });
  if (darkThemeMeta) {
    darkThemeMeta.setAttribute("media", "(prefers-color-scheme: dark)");
    darkThemeMeta.setAttribute("content", resolveThemeColorForMode("dark"));
  }
  const tileMeta = ensureMetaTag("msapplication-TileColor");
  if (tileMeta && color)
    tileMeta.setAttribute("content", color);
  const navMeta = ensureMetaTag("msapplication-navbutton-color");
  if (navMeta && color)
    navMeta.setAttribute("content", color);
  const statusMeta = ensureMetaTag("apple-mobile-web-app-status-bar-style");
  if (statusMeta)
    statusMeta.setAttribute("content", mode === "dark" ? "black-translucent" : "default");
}
function emitThemeChange() {
  if (typeof window === "undefined")
    return;
  const detail = { ...currentTheme, mode: currentMode };
  window.dispatchEvent(new CustomEvent("piclaw-theme-change", { detail }));
}
function getChatThemeMap() {
  try {
    const raw = getLocalStorageItem(CHAT_THEMES_STORAGE_KEY);
    if (!raw)
      return {};
    const parsed = JSON.parse(raw);
    return typeof parsed === "object" && parsed !== null ? parsed : {};
  } catch {
    return {};
  }
}
function getChatTheme(chatJid) {
  if (!chatJid)
    return null;
  const map = getChatThemeMap();
  return map[chatJid] || null;
}
function resolveCurrentChatJid() {
  if (typeof window === "undefined")
    return "web:default";
  try {
    const params = new URL(window.location.href).searchParams;
    const raw = params.get("chat_jid");
    return raw && raw.trim() ? raw.trim() : "web:default";
  } catch {
    return "web:default";
  }
}
function applyThemeState(nextTheme, options = {}) {
  if (typeof window === "undefined" || typeof document === "undefined")
    return;
  const themeName = normalizeThemeName(nextTheme?.theme || "default");
  const tint = nextTheme?.tint ? String(nextTheme.tint).trim() : null;
  const preset = resolvePreset(themeName);
  const mode = resolveModeForPreset(preset);
  const paletteBase = resolvePalette(themeName, mode);
  currentTheme = { theme: themeName, tint };
  currentMode = mode;
  const root = document.documentElement;
  root.dataset.theme = mode;
  root.dataset.colorTheme = themeName;
  root.dataset.tint = tint ? String(tint) : "";
  root.style.colorScheme = mode;
  let palette = paletteBase;
  if (themeName === "default" && tint) {
    palette = buildTintedPalette(paletteBase, tint, mode);
  }
  if (themeName === "default" && !tint) {
    clearCssVariables();
  } else {
    applyCssVariables(palette, mode);
  }
  updateMetaColor(palette.bgPrimary, mode);
  emitThemeChange();
  if (options.persist !== false) {
    setLocalStorageItem(THEME_STORAGE_KEY, themeName);
    if (tint)
      setLocalStorageItem(TINT_STORAGE_KEY, tint);
    else
      setLocalStorageItem(TINT_STORAGE_KEY, "");
  }
}
function handleSystemThemeChange() {
  const preset = resolvePreset(currentTheme.theme);
  if (preset.mode !== "auto")
    return;
  applyThemeState(currentTheme, { persist: false });
}
function initTheme() {
  if (typeof window === "undefined")
    return () => {};
  const chatJid = resolveCurrentChatJid();
  const chatOverride = getChatTheme(chatJid);
  const storedTheme = chatOverride ? normalizeThemeName(chatOverride.theme || "default") : normalizeThemeName(getLocalStorageItem(THEME_STORAGE_KEY) || "default");
  const storedTint = chatOverride ? chatOverride.tint ? String(chatOverride.tint).trim() : null : (() => {
    const raw = getLocalStorageItem(TINT_STORAGE_KEY);
    return raw ? raw.trim() : null;
  })();
  applyThemeState({ theme: storedTheme, tint: storedTint }, { persist: false });
  if (window.matchMedia && !mediaListenerAttached) {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    if (media.addEventListener) {
      media.addEventListener("change", handleSystemThemeChange);
    } else if (media.addListener) {
      media.addListener(handleSystemThemeChange);
    }
    mediaListenerAttached = true;
    return () => {
      if (media.removeEventListener) {
        media.removeEventListener("change", handleSystemThemeChange);
      } else if (media.removeListener) {
        media.removeListener(handleSystemThemeChange);
      }
      mediaListenerAttached = false;
    };
  }
  return () => {};
}
function getThemeMode() {
  if (typeof document === "undefined")
    return "light";
  const attr = document.documentElement?.dataset?.theme;
  if (attr === "dark" || attr === "light")
    return attr;
  return resolveSystemMode();
}

// web/src/ui/status-duration.ts
function parseStatusStartedAt(status) {
  if (!status || typeof status !== "object")
    return null;
  const raw = status.started_at ?? status.startedAt;
  if (typeof raw !== "string" || !raw)
    return null;
  const value = Date.parse(raw);
  return Number.isFinite(value) ? value : null;
}
function parseStatusRetryAt(status) {
  if (!status || typeof status !== "object")
    return null;
  const raw = status.retry_at ?? status.retryAt;
  if (typeof raw !== "string" || !raw)
    return null;
  const value = Date.parse(raw);
  return Number.isFinite(value) ? value : null;
}
function parseStatusLastEventAt(status) {
  if (!status || typeof status !== "object")
    return null;
  const raw = status.last_event_at ?? status.lastEventAt ?? status.started_at ?? status.startedAt;
  if (typeof raw !== "string" || !raw)
    return null;
  const value = Date.parse(raw);
  return Number.isFinite(value) ? value : null;
}
function isCompactionStatus(status) {
  if (!status || typeof status !== "object")
    return false;
  const intentKey = status.intent_key ?? status.intentKey;
  return status.type === "intent" && intentKey === "compaction";
}
function resolveStatusPanelTitle(status) {
  if (!status || typeof status !== "object")
    return "";
  const rawTitle = status.title;
  if (typeof rawTitle === "string" && rawTitle.trim())
    return rawTitle.trim();
  const statusText = status.status;
  if (typeof statusText === "string" && statusText.trim())
    return statusText.trim();
  return isCompactionStatus(status) ? "Compacting context" : "Working...";
}
function formatElapsedDuration(elapsedMs) {
  const totalSeconds = Math.max(0, Math.floor(elapsedMs / 1000));
  const seconds = totalSeconds % 60;
  const minutes = Math.floor(totalSeconds / 60) % 60;
  const hours = Math.floor(totalSeconds / 3600);
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}
function getStatusElapsedLabel(status, nowMs = Date.now()) {
  const startedAtMs = parseStatusStartedAt(status);
  if (startedAtMs === null)
    return null;
  return formatElapsedDuration(Math.max(0, nowMs - startedAtMs));
}
function getStatusRetryCountdownLabel(status, nowMs = Date.now()) {
  const retryAtMs = parseStatusRetryAt(status);
  if (retryAtMs === null)
    return null;
  const remainingMs = retryAtMs - nowMs;
  if (remainingMs <= 0)
    return "retrying now";
  return `retry in ${formatElapsedDuration(remainingMs)}`;
}

// web/src/utils/code-highlighting.ts
import {
  classHighlighter,
  highlightTree,
  StreamLanguage,
  cssLanguage,
  goLanguage,
  htmlLanguage,
  javascriptLanguage,
  jsxLanguage,
  tsxLanguage,
  typescriptLanguage,
  jsonLanguage,
  markdownLanguage,
  pythonLanguage,
  StandardSQL,
  xmlLanguage,
  yamlLanguage,
  dockerFile,
  powerShell,
  ruby,
  rust,
  shell,
  swift,
  toml
} from "#editor-vendor/codemirror";
function escapeHtml(value) {
  return value.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#39;");
}
var LANGUAGE_LABEL_ALIASES = {
  js: "JavaScript",
  javascript: "JavaScript",
  ts: "TypeScript",
  typescript: "TypeScript",
  jsx: "JSX",
  tsx: "TSX",
  py: "Python",
  python: "Python",
  sh: "Shell",
  shell: "Shell",
  bash: "Bash",
  zsh: "Zsh",
  ps1: "PowerShell",
  powershell: "PowerShell",
  md: "Markdown",
  markdown: "Markdown",
  yml: "YAML",
  yaml: "YAML",
  json: "JSON",
  html: "HTML",
  css: "CSS",
  sql: "SQL",
  go: "Go",
  rust: "Rust",
  ruby: "Ruby",
  swift: "Swift",
  toml: "TOML",
  dockerfile: "Dockerfile"
};
var LEGACY_SHELL_PARSER = StreamLanguage.define(shell).parser;
var LEGACY_POWERSHELL_PARSER = StreamLanguage.define(powerShell).parser;
var LEGACY_DOCKERFILE_PARSER = StreamLanguage.define(dockerFile).parser;
var LEGACY_RUBY_PARSER = StreamLanguage.define(ruby).parser;
var LEGACY_RUST_PARSER = StreamLanguage.define(rust).parser;
var LEGACY_SWIFT_PARSER = StreamLanguage.define(swift).parser;
var LEGACY_TOML_PARSER = StreamLanguage.define(toml).parser;
function normalizeCodeLanguageLabel(lang) {
  const raw = String(lang || "").trim().toLowerCase();
  if (!raw)
    return "text";
  return LANGUAGE_LABEL_ALIASES[raw] || String(lang || "").trim();
}
function parserForCodeFenceLanguage(lang) {
  const raw = String(lang || "").trim().toLowerCase();
  switch (raw) {
    case "js":
    case "javascript":
      return javascriptLanguage.parser;
    case "ts":
    case "typescript":
      return typescriptLanguage.parser;
    case "jsx":
      return jsxLanguage.parser;
    case "tsx":
      return tsxLanguage.parser;
    case "py":
    case "python":
      return pythonLanguage.parser;
    case "json":
      return jsonLanguage.parser;
    case "css":
      return cssLanguage.parser;
    case "html":
      return htmlLanguage.parser;
    case "xml":
      return xmlLanguage.parser;
    case "yaml":
    case "yml":
      return yamlLanguage.parser;
    case "md":
    case "markdown":
      return markdownLanguage.parser;
    case "sql":
      return StandardSQL.language.parser;
    case "go":
      return goLanguage.parser;
    case "sh":
    case "bash":
    case "shell":
    case "zsh":
      return LEGACY_SHELL_PARSER;
    case "ps1":
    case "powershell":
      return LEGACY_POWERSHELL_PARSER;
    case "dockerfile":
      return LEGACY_DOCKERFILE_PARSER;
    case "rb":
    case "ruby":
      return LEGACY_RUBY_PARSER;
    case "rs":
    case "rust":
      return LEGACY_RUST_PARSER;
    case "swift":
      return LEGACY_SWIFT_PARSER;
    case "toml":
      return LEGACY_TOML_PARSER;
    default:
      return null;
  }
}
function highlightCodeToHtml(code, lang) {
  const parser = parserForCodeFenceLanguage(lang);
  if (!parser)
    return escapeHtml(code);
  const tokens = [];
  try {
    const tree = parser.parse(code);
    highlightTree(tree, classHighlighter, (from, to, cls) => {
      if (!cls || from >= to)
        return;
      tokens.push({ from, to, cls });
    });
  } catch {
    return escapeHtml(code);
  }
  if (!tokens.length)
    return escapeHtml(code);
  tokens.sort((a2, b) => a2.from - b.from || a2.to - b.to);
  let cursor = 0;
  let html = "";
  for (const token of tokens) {
    if (token.from > cursor)
      html += escapeHtml(code.slice(cursor, token.from));
    html += `<span class="${escapeHtml(token.cls)}">${escapeHtml(code.slice(token.from, token.to))}</span>`;
    cursor = Math.max(cursor, token.to);
  }
  if (cursor < code.length)
    html += escapeHtml(code.slice(cursor));
  return html;
}

// web/src/markdown.ts
var HASHTAG_REGEX = /#(\w+)/g;
var ALLOWED_HTML_TAGS = new Set([
  "strong",
  "em",
  "b",
  "i",
  "u",
  "s",
  "del",
  "ins",
  "sub",
  "sup",
  "mark",
  "small",
  "br",
  "p",
  "ul",
  "ol",
  "li",
  "blockquote",
  "ruby",
  "rt",
  "rp",
  "span",
  "input"
]);
var SAFE_TAGS = new Set([
  "a",
  "abbr",
  "blockquote",
  "br",
  "code",
  "del",
  "div",
  "em",
  "hr",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "i",
  "img",
  "input",
  "ins",
  "kbd",
  "li",
  "mark",
  "ol",
  "p",
  "pre",
  "ruby",
  "rt",
  "rp",
  "s",
  "small",
  "span",
  "strong",
  "sub",
  "sup",
  "table",
  "tbody",
  "td",
  "th",
  "thead",
  "tr",
  "u",
  "ul",
  "math",
  "semantics",
  "mrow",
  "mi",
  "mn",
  "mo",
  "mtext",
  "mspace",
  "msup",
  "msub",
  "msubsup",
  "mfrac",
  "msqrt",
  "mroot",
  "mtable",
  "mtr",
  "mtd",
  "annotation"
]);
var GLOBAL_ALLOWED_ATTRS = new Set([
  "class",
  "title",
  "role",
  "aria-hidden",
  "aria-label",
  "aria-expanded",
  "aria-live",
  "data-mermaid",
  "data-hashtag"
]);
var TAG_ALLOWED_ATTRS = {
  a: new Set(["href", "target", "rel"]),
  img: new Set(["src", "alt", "title"]),
  input: new Set(["type", "checked", "disabled"])
};
var SAFE_PROTOCOLS = new Set(["http:", "https:", "mailto:", ""]);
function isSanitizedHtmlAttributeAllowed(tagName, attrName) {
  const normalizedTag = String(tagName || "").toLowerCase();
  const normalizedAttr = String(attrName || "").toLowerCase();
  if (!normalizedAttr || normalizedAttr.startsWith("on"))
    return false;
  if (normalizedAttr.startsWith("data-") || normalizedAttr.startsWith("aria-")) {
    return true;
  }
  const allowedAttrs = TAG_ALLOWED_ATTRS[normalizedTag] || new Set;
  return allowedAttrs.has(normalizedAttr) || GLOBAL_ALLOWED_ATTRS.has(normalizedAttr);
}
function escapeHtmlAttr(value) {
  return String(value || "").replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/'/g, "&#39;");
}
function sanitizeUrl(url, options = {}) {
  if (!url)
    return null;
  const raw = String(url).trim();
  if (!raw)
    return null;
  if (raw.startsWith("#") || raw.startsWith("/"))
    return raw;
  if (raw.startsWith("data:")) {
    if (options.allowDataImage && /^data:image\//i.test(raw)) {
      return raw;
    }
    return null;
  }
  if (raw.startsWith("blob:"))
    return raw;
  try {
    const parsed = new URL(raw, typeof window !== "undefined" ? window.location.origin : "http://localhost");
    if (!SAFE_PROTOCOLS.has(parsed.protocol))
      return null;
    return parsed.href;
  } catch {
    return null;
  }
}
function sanitizeHtml(html, options = {}) {
  if (!html)
    return "";
  if (options?.sanitize === false)
    return html;
  const doc = new DOMParser().parseFromString(html, "text/html");
  const nodes = [];
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_ELEMENT);
  let node;
  while (node = walker.nextNode()) {
    nodes.push(node);
  }
  for (const el of nodes) {
    const tag = el.tagName.toLowerCase();
    if (!SAFE_TAGS.has(tag)) {
      const parent = el.parentNode;
      if (!parent)
        continue;
      while (el.firstChild) {
        parent.insertBefore(el.firstChild, el);
      }
      parent.removeChild(el);
      continue;
    }
    const allowedAttrs = TAG_ALLOWED_ATTRS[tag] || new Set;
    for (const attr of Array.from(el.attributes)) {
      const name = attr.name.toLowerCase();
      const value = attr.value;
      if (name.startsWith("on")) {
        el.removeAttribute(attr.name);
        continue;
      }
      if (isSanitizedHtmlAttributeAllowed(tag, name)) {
        if (name === "href") {
          const safe = sanitizeUrl(value);
          if (!safe) {
            el.removeAttribute(attr.name);
          } else {
            el.setAttribute(attr.name, safe);
            if (tag === "a") {
              if (!el.getAttribute("rel")) {
                el.setAttribute("rel", "noopener noreferrer");
              }
              if (/^https?:\/\//i.test(safe)) {
                el.setAttribute("target", "_blank");
              }
            }
          }
        } else if (name === "src") {
          const rewritten = tag === "img" && typeof options.rewriteImageSrc === "function" ? options.rewriteImageSrc(value) : value;
          const safe = sanitizeUrl(rewritten, { allowDataImage: tag === "img" });
          if (!safe) {
            el.removeAttribute(attr.name);
          } else {
            el.setAttribute(attr.name, safe);
          }
        }
        continue;
      }
      el.removeAttribute(attr.name);
    }
  }
  return doc.body.innerHTML;
}
function decodeEntities(text) {
  if (!text)
    return text;
  const safe = text.replace(/</g, "&lt;").replace(/>/g, "&gt;");
  const doc = new DOMParser().parseFromString(safe, "text/html");
  return doc.documentElement.textContent;
}
function decodeEntitiesDeep(text, maxDepth = 2) {
  if (!text)
    return text;
  let current = text;
  for (let i2 = 0;i2 < maxDepth; i2 += 1) {
    const next = decodeEntities(current);
    if (next === current)
      break;
    current = next;
  }
  return current;
}
function extractLeadingYamlFrontmatter(text) {
  if (!text)
    return { text: "", frontmatter: null };
  const normalized = text.replace(/^\uFEFF/, "").replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  if (!normalized.startsWith(`---
`)) {
    return { text: normalized, frontmatter: null };
  }
  const lines = normalized.split(`
`);
  let closingIndex = -1;
  for (let i2 = 1;i2 < lines.length; i2 += 1) {
    if (/^(---|\.\.\.)\s*$/.test(lines[i2])) {
      closingIndex = i2;
      break;
    }
  }
  if (closingIndex <= 0) {
    return { text: normalized, frontmatter: null };
  }
  const frontmatter = lines.slice(1, closingIndex).join(`
`);
  const body = lines.slice(closingIndex + 1).join(`
`).replace(/^\n+/, "");
  return { text: body, frontmatter };
}
function normalizeLeadingFrontmatter(text) {
  const { text: body, frontmatter } = extractLeadingYamlFrontmatter(text);
  if (frontmatter === null)
    return body;
  return [
    "<!--frontmatter-block-start-->",
    "```yaml",
    frontmatter,
    "```",
    "<!--frontmatter-block-end-->",
    body
  ].filter(Boolean).join(`

`);
}
function extractMermaidBlocks(text) {
  if (!text)
    return { text: "", blocks: [] };
  const normalized = text.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  const blocks = [];
  const output = [];
  let inMermaid = false;
  let current = [];
  for (const line of lines) {
    if (!inMermaid && line.trim().match(/^```mermaid\s*$/i)) {
      inMermaid = true;
      current = [];
      continue;
    }
    if (inMermaid && line.trim().match(/^```\s*$/)) {
      const idx = blocks.length;
      blocks.push(current.join(`
`));
      output.push(`@@MERMAID_BLOCK_${idx}@@`);
      inMermaid = false;
      current = [];
      continue;
    }
    if (inMermaid) {
      current.push(line);
    } else {
      output.push(line);
    }
  }
  if (inMermaid) {
    output.push("```mermaid");
    output.push(...current);
  }
  return { text: output.join(`
`), blocks };
}
function decodeMermaidBlock(text) {
  if (!text)
    return text;
  return decodeEntitiesDeep(text, 5);
}
function toBase64(value) {
  const bytes = new TextEncoder().encode(String(value || ""));
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary);
}
function fromBase64(value) {
  const binary = atob(String(value || ""));
  const bytes = new Uint8Array(binary.length);
  for (let i2 = 0;i2 < binary.length; i2 += 1) {
    bytes[i2] = binary.charCodeAt(i2);
  }
  return new TextDecoder().decode(bytes);
}
function injectMermaidBlocks(html, blocks) {
  if (!html || !blocks || blocks.length === 0)
    return html;
  return html.replace(/@@MERMAID_BLOCK_(\d+)@@/g, (match, idxStr) => {
    const idx = Number(idxStr);
    const raw = blocks[idx] ?? "";
    const decoded = decodeMermaidBlock(raw);
    const encoded = toBase64(decoded);
    return `<div class="mermaid-container" data-mermaid="${encoded}"><div class="mermaid-loading">Loading diagram...</div></div>`;
  });
}
function normalizeHtmlCodeTags(text) {
  if (!text)
    return text;
  return text.replace(/<code>([\s\S]*?)<\/code>/gi, (match, code) => {
    if (code.includes(`
`)) {
      return `
\`\`\`
${code}
\`\`\`
`;
    }
    return `\`${code}\``;
  });
}
function applySyntaxHighlighting(html) {
  if (!html)
    return html;
  const highlighted = html.replace(/<pre><code(?:\s+class="language-([A-Za-z0-9_+-]+)")?>([\s\S]*?)<\/code><\/pre>/g, (match, lang, code) => {
    const normalizedLanguage = String(lang || "").trim().toLowerCase();
    const decodedCode = decodeEntitiesDeep(code, 2);
    const languageClass = normalizedLanguage || "plaintext";
    const highlightedCode = highlightCodeToHtml(decodedCode, normalizedLanguage);
    return `<pre><code class="hljs language-${escapeHtmlAttr(languageClass)}">${highlightedCode}</code></pre>`;
  });
  return highlighted.replace(/<!--frontmatter-block-start-->\s*<pre>/g, '<pre class="frontmatter-block">').replace(/<\/pre>\s*<!--frontmatter-block-end-->/g, "</pre>");
}
var RESTORABLE_HTML_ATTRS = {
  span: new Set(["title", "class", "lang", "dir"]),
  input: new Set(["type", "checked", "disabled"])
};
function extractRestorableAttributes(tagName, rawAttrs) {
  const allowed = RESTORABLE_HTML_ATTRS[tagName];
  if (!allowed || !rawAttrs)
    return "";
  const attrs = [];
  const regex = /([a-zA-Z_:][\w:.-]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'`=<>]+)))?/g;
  let match;
  while (match = regex.exec(rawAttrs)) {
    const name = (match[1] || "").toLowerCase();
    if (!name || name.startsWith("on") || !allowed.has(name))
      continue;
    const rawValue = match[2] ?? match[3] ?? match[4] ?? "";
    attrs.push(` ${name}="${escapeHtmlAttr(rawValue)}"`);
  }
  return attrs.join("");
}
function restoreAllowedHtmlTags(text) {
  if (!text)
    return text;
  return text.replace(/&lt;((?:[^"'<>]|"[^"]*"|'[^']*')*?)(?:&gt;|>)/g, (match, content) => {
    const trimmed = content.trim();
    const isClosing = trimmed.startsWith("/");
    const rawTag = isClosing ? trimmed.slice(1).trim() : trimmed;
    const isSelfClosing = rawTag.endsWith("/");
    const tagContent = isSelfClosing ? rawTag.slice(0, -1).trim() : rawTag;
    const [tagToken = ""] = tagContent.split(/\s+/, 1);
    const tagName = tagToken.toLowerCase();
    if (!tagName || !ALLOWED_HTML_TAGS.has(tagName))
      return match;
    if (tagName === "br") {
      return isClosing ? "" : "<br>";
    }
    if (isClosing)
      return `</${tagName}>`;
    const attrSource = tagContent.slice(tagToken.length).trim();
    const attrs = extractRestorableAttributes(tagName, attrSource);
    return `<${tagName}${attrs}>`;
  });
}
function decodeCodeEntities(html) {
  if (!html)
    return html;
  const normalize = (value) => value.replace(/&amp;lt;/g, "&lt;").replace(/&amp;gt;/g, "&gt;").replace(/&amp;quot;/g, "&quot;").replace(/&amp;#39;/g, "&#39;").replace(/&amp;amp;/g, "&amp;");
  return html.replace(/<pre><code>([\s\S]*?)<\/code><\/pre>/g, (match, code) => `<pre><code>${normalize(code)}</code></pre>`).replace(/<code>([\s\S]*?)<\/code>/g, (match, code) => `<code>${normalize(code)}</code>`);
}
function decodeTextEntities(html) {
  if (!html)
    return html;
  const doc = new DOMParser().parseFromString(html, "text/html");
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
  const decode = (value) => value.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&amp;/g, "&");
  let node;
  while (node = walker.nextNode()) {
    if (!node.nodeValue)
      continue;
    const next = decode(node.nodeValue);
    if (next !== node.nodeValue) {
      node.nodeValue = next;
    }
  }
  return doc.body.innerHTML;
}
function renderMath(html_content) {
  if (!window.katex)
    return html_content;
  const decodeMath = (value) => decodeEntities(value).replace(/&gt;/g, ">").replace(/&lt;/g, "<").replace(/&amp;/g, "&").replace(/<br\s*\/?\s*>/gi, `
`);
  const stripCodeBlocks = (html) => {
    const blocks = [];
    let output = html.replace(/<pre\b[^>]*>\s*<code\b[^>]*>[\s\S]*?<\/code>\s*<\/pre>/gi, (match) => {
      const idx = blocks.length;
      blocks.push(match);
      return `@@CODE_BLOCK_${idx}@@`;
    });
    output = output.replace(/<code\b[^>]*>[\s\S]*?<\/code>/gi, (match) => {
      const idx = blocks.length;
      blocks.push(match);
      return `@@CODE_INLINE_${idx}@@`;
    });
    return { html: output, blocks };
  };
  const restoreCodeBlocks = (html, blocks) => {
    if (!blocks.length)
      return html;
    return html.replace(/@@CODE_(?:BLOCK|INLINE)_(\d+)@@/g, (_match, idxStr) => {
      const idx = Number(idxStr);
      return blocks[idx] ?? "";
    });
  };
  const stripped = stripCodeBlocks(html_content);
  let processed = stripped.html;
  processed = processed.replace(/(^|\n|<br\s*\/?\s*>|<p>|<\/p>)\s*\$\$([\s\S]+?)\$\$\s*(?=\n|<br\s*\/?\s*>|<\/p>|$)/gi, (match, leading, tex) => {
    try {
      const rendered = katex.renderToString(decodeMath(tex.trim()), { displayMode: true, throwOnError: false });
      return `${leading}${rendered}`;
    } catch (e2) {
      return `<span class="math-error" title="${escapeHtmlAttr(e2.message)}">${match}</span>`;
    }
  });
  return restoreCodeBlocks(processed, stripped.blocks);
}
function linkifyHashtagsInHtml(html_content) {
  if (!html_content)
    return html_content;
  const doc = new DOMParser().parseFromString(html_content, "text/html");
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
  const nodes = [];
  let node;
  while (node = walker.nextNode()) {
    nodes.push(node);
  }
  for (const textNode of nodes) {
    const value = textNode.nodeValue;
    if (!value)
      continue;
    HASHTAG_REGEX.lastIndex = 0;
    if (!HASHTAG_REGEX.test(value))
      continue;
    HASHTAG_REGEX.lastIndex = 0;
    const parent = textNode.parentElement;
    if (parent && (parent.closest("a") || parent.closest("code") || parent.closest("pre")))
      continue;
    const parts = value.split(HASHTAG_REGEX);
    if (parts.length <= 1)
      continue;
    const fragment = doc.createDocumentFragment();
    parts.forEach((part, idx) => {
      if (idx % 2 === 1) {
        const link = doc.createElement("a");
        link.setAttribute("href", "#");
        link.className = "hashtag";
        link.setAttribute("data-hashtag", part);
        link.textContent = `#${part}`;
        fragment.appendChild(link);
      } else {
        fragment.appendChild(doc.createTextNode(part));
      }
    });
    textNode.parentNode?.replaceChild(fragment, textNode);
  }
  return doc.body.innerHTML;
}
function normalizeMathFences(text) {
  if (!text)
    return text;
  const normalized = text.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  const output = [];
  let inMath = false;
  for (const line of lines) {
    if (!inMath && line.trim().match(/^```(?:math|katex|latex)\s*$/i)) {
      inMath = true;
      output.push("$$");
      continue;
    }
    if (inMath && line.trim().match(/^```\s*$/)) {
      inMath = false;
      output.push("$$");
      continue;
    }
    output.push(line);
  }
  return output.join(`
`);
}
function prepareMarkdownSource(text) {
  const normalizedFrontmatter = normalizeLeadingFrontmatter(text || "");
  const normalizedMath = normalizeMathFences(normalizedFrontmatter);
  const { text: stripped, blocks: mermaidBlocks } = extractMermaidBlocks(normalizedMath);
  const decoded = decodeEntitiesDeep(stripped, 2);
  const normalized = normalizeHtmlCodeTags(decoded);
  const escaped = normalized.replace(/</g, "&lt;");
  const safeHtml = restoreAllowedHtmlTags(escaped);
  return { safeHtml, mermaidBlocks };
}
function renderMarkdown(text, onHashtagClick, options = {}) {
  if (!text)
    return "";
  const { safeHtml, mermaidBlocks } = prepareMarkdownSource(text);
  let html_content = window.marked ? marked.parse(safeHtml, { headerIds: false, mangle: false }) : safeHtml.replace(/\n/g, "<br>");
  html_content = decodeCodeEntities(html_content);
  html_content = decodeTextEntities(html_content);
  html_content = applySyntaxHighlighting(html_content);
  html_content = renderMath(html_content);
  html_content = linkifyHashtagsInHtml(html_content);
  html_content = injectMermaidBlocks(html_content, mermaidBlocks);
  html_content = sanitizeHtml(html_content, options);
  return html_content;
}
function renderThinkingMarkdown(text) {
  if (!text)
    return "";
  const normalized = text.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const decoded = decodeEntitiesDeep(normalized, 2);
  const normalizedHtml = normalizeHtmlCodeTags(decoded);
  const escaped = normalizedHtml.replace(/</g, "&lt;").replace(/>/g, "&gt;");
  const safeHtml = restoreAllowedHtmlTags(escaped);
  let html_content = window.marked ? marked.parse(safeHtml) : safeHtml.replace(/\n/g, "<br>");
  html_content = decodeCodeEntities(html_content);
  html_content = decodeTextEntities(html_content);
  html_content = sanitizeHtml(html_content);
  return html_content;
}
function roundPolylineCorners(svgString, radius = 6) {
  return svgString.replace(/<polyline\b([^>]*)\bpoints="([^"]+)"([^>]*)\/?\s*>/g, (_match, before, pointsStr, after) => {
    const pts = pointsStr.trim().split(/\s+/).map((p2) => {
      const [x2, y2] = p2.split(",").map(Number);
      return { x: x2, y: y2 };
    });
    if (pts.length < 3) {
      return `<polyline${before}points="${pointsStr}"${after}/>`;
    }
    const parts = [`M ${pts[0].x},${pts[0].y}`];
    for (let i2 = 1;i2 < pts.length - 1; i2++) {
      const prev = pts[i2 - 1];
      const curr = pts[i2];
      const next = pts[i2 + 1];
      const dxIn = curr.x - prev.x;
      const dyIn = curr.y - prev.y;
      const dxOut = next.x - curr.x;
      const dyOut = next.y - curr.y;
      const lenIn = Math.sqrt(dxIn * dxIn + dyIn * dyIn);
      const lenOut = Math.sqrt(dxOut * dxOut + dyOut * dyOut);
      const r2 = Math.min(radius, lenIn / 2, lenOut / 2);
      if (r2 < 0.5) {
        parts.push(`L ${curr.x},${curr.y}`);
        continue;
      }
      const ax = curr.x - dxIn / lenIn * r2;
      const ay = curr.y - dyIn / lenIn * r2;
      const bx = curr.x + dxOut / lenOut * r2;
      const by = curr.y + dyOut / lenOut * r2;
      const cross = dxIn * dyOut - dyIn * dxOut;
      const sweep = cross > 0 ? 1 : 0;
      parts.push(`L ${ax},${ay}`);
      parts.push(`A ${r2},${r2} 0 0 ${sweep} ${bx},${by}`);
    }
    parts.push(`L ${pts[pts.length - 1].x},${pts[pts.length - 1].y}`);
    return `<path${before}d="${parts.join(" ")}"${after}/>`;
  });
}
async function renderMermaidDiagrams(container) {
  if (!window.beautifulMermaid)
    return;
  const { renderMermaid, THEMES } = window.beautifulMermaid;
  const isDark = getThemeMode() === "dark";
  const theme = isDark ? THEMES["tokyo-night"] : THEMES["github-light"];
  const pending = container.querySelectorAll(".mermaid-container[data-mermaid]");
  for (const el of pending) {
    try {
      const encoded = el.dataset.mermaid;
      const raw = fromBase64(encoded || "");
      const code = decodeEntitiesDeep(raw, 2);
      let svg = await renderMermaid(code, { ...theme, transparent: true });
      svg = roundPolylineCorners(svg);
      el.innerHTML = svg;
      el.removeAttribute("data-mermaid");
    } catch (e2) {
      console.error("Mermaid render error:", e2);
      const pre = document.createElement("pre");
      pre.className = "mermaid-error";
      pre.textContent = `Diagram error: ${e2.message}`;
      el.innerHTML = "";
      el.appendChild(pre);
      el.removeAttribute("data-mermaid");
    }
  }
}

// web/src/utils/format.ts
function formatTime(timestamp) {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime()))
    return timestamp;
  const now = new Date;
  const diffMs = now - date;
  const diffSec = diffMs / 1000;
  const dayMs = 24 * 60 * 60 * 1000;
  if (diffMs < dayMs) {
    if (diffSec < 60)
      return "just now";
    if (diffSec < 3600)
      return `${Math.floor(diffSec / 60)}m`;
    return `${Math.floor(diffSec / 3600)}h`;
  }
  if (diffMs < 5 * dayMs) {
    const weekday = date.toLocaleDateString(undefined, { weekday: "short" });
    const time = date.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" });
    return `${weekday} ${time}`;
  }
  const datePart = date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  const timePart = date.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" });
  return `${datePart} ${timePart}`;
}
function formatCount(value) {
  if (!Number.isFinite(value))
    return "0";
  return Math.round(value).toLocaleString();
}
function formatFileSize(bytes) {
  if (bytes < 1024)
    return bytes + " B";
  if (bytes < 1024 * 1024)
    return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}
function formatTimestamp(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime()))
    return value;
  return date.toLocaleString();
}

// web/src/ui/adaptive-card-submission.ts
function formatSubmissionValue(value) {
  if (value == null)
    return "";
  if (typeof value === "string")
    return value.trim();
  if (typeof value === "number")
    return String(value);
  if (typeof value === "boolean")
    return value ? "yes" : "no";
  if (Array.isArray(value)) {
    return value.map((item) => formatSubmissionValue(item)).filter(Boolean).join(", ");
  }
  if (typeof value === "object") {
    return Object.entries(value).filter(([key]) => !key.startsWith("__")).map(([key, inner]) => `${key}: ${formatSubmissionValue(inner)}`).filter((entry) => !entry.endsWith(": ")).join(", ");
  }
  return String(value).trim();
}
function getSubmissionFields(data) {
  if (!(typeof data === "object") || data == null || Array.isArray(data))
    return [];
  return Object.entries(data).filter(([key]) => !key.startsWith("__")).map(([key, value]) => ({ key, value: formatSubmissionValue(value) })).filter((entry) => entry.value);
}
function isAdaptiveCardSubmissionBlock(block) {
  if (!block || typeof block !== "object")
    return false;
  const candidate = block;
  return candidate.type === "adaptive_card_submission" && typeof candidate.card_id === "string" && typeof candidate.source_post_id === "number" && typeof candidate.submitted_at === "string";
}
function extractAdaptiveCardSubmissionBlocks(contentBlocks) {
  if (!Array.isArray(contentBlocks))
    return [];
  return contentBlocks.filter(isAdaptiveCardSubmissionBlock);
}
function buildAdaptiveCardSubmissionFallbackText(block) {
  const label = String(block.title || block.card_id || "card").trim() || "card";
  const data = block.data;
  if (data == null)
    return `Card submission: ${label}`;
  if (typeof data === "string" || typeof data === "number" || typeof data === "boolean") {
    const formatted = formatSubmissionValue(data);
    return formatted ? `Card submission: ${label} — ${formatted}` : `Card submission: ${label}`;
  }
  if (typeof data === "object") {
    const fields = getSubmissionFields(data);
    const entries = fields.map(({ key, value }) => `${key}: ${value}`);
    return entries.length > 0 ? `Card submission: ${label} — ${entries.join(", ")}` : `Card submission: ${label}`;
  }
  return `Card submission: ${label}`;
}
function describeAdaptiveCardSubmission(block) {
  const title = String(block.title || block.card_id || "Card submission").trim() || "Card submission";
  const allFields = getSubmissionFields(block.data);
  const summary = allFields.length > 0 ? allFields.slice(0, 2).map(({ key, value }) => `${key}: ${value}`).join(", ") : formatSubmissionValue(block.data) || null;
  const fieldCount = allFields.length;
  return {
    title,
    summary,
    fields: allFields,
    fieldCount,
    submittedAt: block.submitted_at
  };
}

// web/src/utils/post-copy-markdown.ts
function cleanString(value) {
  return typeof value === "string" ? value.trim() : "";
}
function joinSections(sections) {
  return sections.map((section) => String(section || "").trim()).filter(Boolean).join(`

`).replace(/\n{3,}/g, `

`).trim();
}
function buildStructuredBlocksMarkdown(blocks, mediaIds) {
  const sections = [];
  const attachmentLines = [];
  const imageLines = [];
  blocks.forEach((block, index) => {
    if (!block || typeof block !== "object")
      return;
    const type = cleanString(block.type);
    if (type === "text") {
      const text = cleanString(block.text) || cleanString(block.content);
      if (text)
        sections.push(text);
      return;
    }
    if (type === "resource_link") {
      const uri = cleanString(block.uri);
      const title = cleanString(block.title) || cleanString(block.name) || uri;
      if (uri && title) {
        sections.push(title === uri ? uri : `[${title}](${uri})`);
      }
      return;
    }
    if (type === "resource") {
      const title = cleanString(block.title) || cleanString(block.name) || cleanString(block.uri) || "Embedded resource";
      const text = cleanString(block.text);
      if (text) {
        sections.push(`### ${title}

\`\`\`
${text}
\`\`\``);
      } else {
        sections.push(`### ${title}`);
      }
      return;
    }
    if (type === "generated_widget") {
      const title = cleanString(block.title) || cleanString(block.name) || "Generated widget";
      const description = cleanString(block.description) || cleanString(block.subtitle);
      sections.push(joinSections([`### ${title}`, description]));
      return;
    }
    if (type === "adaptive_card" && cleanString(block.fallback_text)) {
      sections.push(cleanString(block.fallback_text));
      return;
    }
    if (type === "adaptive_card_submission") {
      const fallback = buildAdaptiveCardSubmissionFallbackText(block);
      if (cleanString(fallback))
        sections.push(cleanString(fallback));
      return;
    }
    if (type === "file") {
      const label = cleanString(block.name) || cleanString(block.filename) || cleanString(block.title) || `attachment:${mediaIds[index] ?? index + 1}`;
      attachmentLines.push(`- ${label}`);
      return;
    }
    if (type === "image" || !type) {
      const label = cleanString(block.name) || cleanString(block.filename) || cleanString(block.title) || `attachment:${mediaIds[index] ?? index + 1}`;
      imageLines.push(`- ${label}`);
    }
  });
  if (imageLines.length > 0)
    sections.push(`Images:
${imageLines.join(`
`)}`);
  if (attachmentLines.length > 0)
    sections.push(`Attachments:
${attachmentLines.join(`
`)}`);
  return joinSections(sections);
}
function buildPostMarkdownCopyPayload(post) {
  const data = post?.data || {};
  const rawContent = typeof data.content === "string" ? data.content.replace(/\r\n/g, `
`).replace(/\r/g, `
`).trimEnd() : "";
  if (rawContent.trim())
    return rawContent;
  const blocks = Array.isArray(data.content_blocks) ? data.content_blocks : [];
  const mediaIds = Array.isArray(data.media_ids) ? data.media_ids : [];
  return buildStructuredBlocksMarkdown(blocks, mediaIds);
}

// web/src/ui/agent-utils.ts
var DEFAULT_AGENT_NAME = "PiClaw";
var AGENT_AVATAR_URL = "/static/icon-192.png";
function getAvatarInfo(name, avatarUrl, isAgent = false) {
  const resolvedName = name || DEFAULT_AGENT_NAME;
  const letter = resolvedName.charAt(0).toUpperCase();
  const colors = [
    "#FF6B6B",
    "#4ECDC4",
    "#45B7D1",
    "#FFA07A",
    "#98D8C8",
    "#F7DC6F",
    "#BB8FCE",
    "#85C1E2",
    "#F8B195",
    "#6C5CE7",
    "#00B894",
    "#FDCB6E",
    "#E17055",
    "#74B9FF",
    "#A29BFE",
    "#FD79A8",
    "#00CEC9",
    "#FFEAA7",
    "#DFE6E9",
    "#FF7675",
    "#55EFC4",
    "#81ECEC",
    "#FAB1A0",
    "#74B9FF",
    "#A29BFE",
    "#FD79A8"
  ];
  const index = letter.charCodeAt(0) % colors.length;
  const color = colors[index];
  const normalized = resolvedName.trim().toLowerCase();
  const normalizedAvatar = typeof avatarUrl === "string" ? avatarUrl.trim() : "";
  const customImage = normalizedAvatar ? normalizedAvatar : null;
  const shouldUseDefaultImage = isAgent || normalized === DEFAULT_AGENT_NAME.toLowerCase() || normalized === "pi";
  const image = customImage || (shouldUseDefaultImage ? AGENT_AVATAR_URL : null);
  return { letter, color, image };
}
function getAgentName(agentId, agents) {
  if (!agentId)
    return DEFAULT_AGENT_NAME;
  const name = agents[agentId]?.name || agentId;
  return name ? name.charAt(0).toUpperCase() + name.slice(1) : DEFAULT_AGENT_NAME;
}
function getAgentAvatarUrl(agentId, agents) {
  if (!agentId)
    return null;
  const agent = agents[agentId] || {};
  return agent.avatar_url || agent.avatarUrl || agent.avatar || null;
}
function getTurnColor(turnId) {
  if (!turnId)
    return null;
  if (typeof document !== "undefined") {
    const root = document.documentElement;
    const themeName = root?.dataset?.colorTheme || "";
    const tint = root?.dataset?.tint || "";
    const accent = getComputedStyle(root).getPropertyValue("--accent-color")?.trim();
    if (accent && (tint || themeName && themeName !== "default")) {
      return accent;
    }
  }
  const palette = [
    "#4ECDC4",
    "#FF6B6B",
    "#45B7D1",
    "#BB8FCE",
    "#FDCB6E",
    "#00B894",
    "#74B9FF",
    "#FD79A8",
    "#81ECEC",
    "#FFA07A"
  ];
  const str = String(turnId);
  let hash = 0;
  for (let i2 = 0;i2 < str.length; i2 += 1) {
    hash = (hash * 31 + str.charCodeAt(i2)) % 2147483647;
  }
  const index = Math.abs(hash) % palette.length;
  return palette[index];
}

// web/src/ui/attachment-preview.ts
var TEXT_PREVIEW_TYPES = new Set([
  "application/json",
  "application/xml",
  "text/csv",
  "text/html",
  "text/markdown",
  "text/plain",
  "text/xml"
]);
var MARKDOWN_PREVIEW_TYPES = new Set([
  "text/markdown"
]);
var OFFICE_PREVIEW_TYPES = new Set([
  "application/msword",
  "application/rtf",
  "application/vnd.ms-excel",
  "application/vnd.ms-powerpoint",
  "application/vnd.oasis.opendocument.presentation",
  "application/vnd.oasis.opendocument.spreadsheet",
  "application/vnd.oasis.opendocument.text",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
]);
var DRAWIO_PREVIEW_TYPES = new Set([
  "application/vnd.jgraph.mxfile"
]);
function normalize(value) {
  return typeof value === "string" ? value.trim().toLowerCase() : "";
}
function isDrawioFilename(filename) {
  const name = normalize(filename);
  return !!name && (name.endsWith(".drawio") || name.endsWith(".drawio.xml") || name.endsWith(".drawio.svg") || name.endsWith(".drawio.png"));
}
function isPdfFilename(filename) {
  const name = normalize(filename);
  return !!name && name.endsWith(".pdf");
}
function isOfficeFilename(filename) {
  const name = normalize(filename);
  return !!name && (name.endsWith(".docx") || name.endsWith(".doc") || name.endsWith(".odt") || name.endsWith(".rtf") || name.endsWith(".xlsx") || name.endsWith(".xls") || name.endsWith(".ods") || name.endsWith(".pptx") || name.endsWith(".ppt") || name.endsWith(".odp"));
}
var ARCHIVE_PREVIEW_TYPES = new Set([
  "application/zip",
  "application/x-zip-compressed"
]);
function isArchiveFilename(filename) {
  const name = normalize(filename);
  return !!name && name.endsWith(".zip");
}
function isHtmlFilename(filename) {
  const name = normalize(filename);
  return !!name && (name.endsWith(".html") || name.endsWith(".htm"));
}
function isTextFilename(filename) {
  const name = normalize(filename);
  if (!name)
    return false;
  return name.endsWith(".sh") || name.endsWith(".bash") || name.endsWith(".zsh") || name.endsWith(".sb");
}
function getAttachmentPreviewKind(contentType, filename) {
  const normalized = normalize(contentType);
  if (isDrawioFilename(filename) || DRAWIO_PREVIEW_TYPES.has(normalized))
    return "drawio";
  if (isPdfFilename(filename) || normalized === "application/pdf")
    return "pdf";
  if (isOfficeFilename(filename) || OFFICE_PREVIEW_TYPES.has(normalized))
    return "office";
  if (isArchiveFilename(filename) || ARCHIVE_PREVIEW_TYPES.has(normalized))
    return "archive";
  if (isHtmlFilename(filename) || normalized === "text/html")
    return "html";
  if (isTextFilename(filename))
    return "text";
  if (!normalized)
    return "unsupported";
  if (normalized.startsWith("video/"))
    return "video";
  if (normalized.startsWith("image/"))
    return "image";
  if (TEXT_PREVIEW_TYPES.has(normalized) || normalized.startsWith("text/"))
    return "text";
  return "unsupported";
}
function isMarkdownAttachmentPreview(contentType) {
  const normalized = normalize(contentType);
  return MARKDOWN_PREVIEW_TYPES.has(normalized);
}
function getAttachmentPreviewLabel(kind) {
  switch (kind) {
    case "image":
      return "Image preview";
    case "video":
      return "Video player";
    case "pdf":
      return "PDF preview";
    case "office":
      return "Office viewer";
    case "drawio":
      return "Draw.io preview (read-only)";
    case "html":
      return "HTML preview";
    case "text":
      return "Text preview";
    case "archive":
      return "ZIP archive preview";
    default:
      return "Preview unavailable";
  }
}

// web/src/ui/adaptive-card-input-lock.ts
function setAdaptiveCardAttributeBestEffort(element, name, value) {
  try {
    element.setAttribute(name, value);
    return true;
  } catch (_error) {
    return false;
  }
}
function setAdaptiveCardBooleanPropertyBestEffort(element, property) {
  try {
    element[property] = true;
    return true;
  } catch (_error) {
    return false;
  }
}
function lockAdaptiveCardInputs(root) {
  root.classList.add("adaptive-card-readonly");
  for (const input of Array.from(root.querySelectorAll("input, textarea, select, button"))) {
    const element = input;
    setAdaptiveCardAttributeBestEffort(element, "aria-disabled", "true");
    setAdaptiveCardAttributeBestEffort(element, "tabindex", "-1");
    if ("disabled" in element) {
      setAdaptiveCardBooleanPropertyBestEffort(element, "disabled");
    }
    if ("readOnly" in element) {
      setAdaptiveCardBooleanPropertyBestEffort(element, "readOnly");
    }
  }
}

// web/src/ui/adaptive-card-host-config.ts
function parseHexColor2(input) {
  const raw = String(input || "").trim();
  const match = raw.match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/i);
  if (!match)
    return null;
  const hex = match[1].length === 3 ? match[1].split("").map((part) => `${part}${part}`).join("") : match[1];
  return {
    r: parseInt(hex.slice(0, 2), 16),
    g: parseInt(hex.slice(2, 4), 16),
    b: parseInt(hex.slice(4, 6), 16)
  };
}
function parseRgbColor(input) {
  const raw = String(input || "").trim();
  const match = raw.match(/^rgba?\((\d+)[,\s]+(\d+)[,\s]+(\d+)/i);
  if (!match)
    return null;
  const r2 = Number(match[1]);
  const g = Number(match[2]);
  const b = Number(match[3]);
  if (![r2, g, b].every((value) => Number.isFinite(value)))
    return null;
  return { r: r2, g, b };
}
function parseColor2(input) {
  return parseHexColor2(input) || parseRgbColor(input);
}
function relativeLuminance2(color) {
  const toLinear = (channel) => {
    const value = channel / 255;
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  };
  const r2 = toLinear(color.r);
  const g = toLinear(color.g);
  const b = toLinear(color.b);
  return 0.2126 * r2 + 0.7152 * g + 0.0722 * b;
}
function contrastRatio(a2, b) {
  const lighter = Math.max(relativeLuminance2(a2), relativeLuminance2(b));
  const darker = Math.min(relativeLuminance2(a2), relativeLuminance2(b));
  return (lighter + 0.05) / (darker + 0.05);
}
function pickHighestContrastColor(background, candidates, fallback = "#ffffff") {
  const backgroundColor = parseColor2(background);
  if (!backgroundColor)
    return fallback;
  let best = fallback;
  let bestScore = -1;
  for (const candidate of candidates) {
    const parsed = parseColor2(candidate);
    if (!parsed)
      continue;
    const score = contrastRatio(backgroundColor, parsed);
    if (score > bestScore) {
      best = candidate;
      bestScore = score;
    }
  }
  return best;
}
function getAdaptiveCardThemeValues() {
  const style = getComputedStyle(document.documentElement);
  const getAny = (names, fallback) => {
    for (const name of names) {
      const value = style.getPropertyValue(name).trim();
      if (value)
        return value;
    }
    return fallback;
  };
  const fg = getAny(["--text-primary", "--color-text"], "#0f1419");
  const fgMuted = getAny(["--text-secondary", "--color-text-muted"], "#536471");
  const bgPrimary = getAny(["--bg-primary", "--color-bg-primary"], "#ffffff");
  const bg = getAny(["--bg-secondary", "--color-bg-secondary"], "#f7f9fa");
  const bgEmphasis = getAny(["--bg-hover", "--bg-tertiary", "--color-bg-tertiary"], "#e8ebed");
  const accent = getAny(["--accent-color", "--color-accent"], "#1d9bf0");
  const good = getAny(["--success-color", "--color-success"], "#00ba7c");
  const warning = getAny(["--warning-color", "--color-warning", "--accent-color"], "#f0b429");
  const attention = getAny(["--danger-color", "--color-error"], "#f4212e");
  const border = getAny(["--border-color", "--color-border"], "#eff3f4");
  const fontFamily = getAny(["--font-family"], "system-ui, sans-serif");
  const buttonTextColor = pickHighestContrastColor(accent, [fg, bgPrimary], fg);
  return {
    fg,
    fgMuted,
    bgPrimary,
    bg,
    bgEmphasis,
    accent,
    good,
    warning,
    attention,
    border,
    fontFamily,
    buttonTextColor
  };
}
function buildHostConfig() {
  const {
    fg,
    fgMuted,
    bg,
    bgEmphasis,
    accent,
    good,
    warning,
    attention,
    border,
    fontFamily
  } = getAdaptiveCardThemeValues();
  return {
    fontFamily,
    containerStyles: {
      default: {
        backgroundColor: bg,
        foregroundColors: {
          default: { default: fg, subtle: fgMuted },
          accent: { default: accent, subtle: accent },
          good: { default: good, subtle: good },
          warning: { default: warning, subtle: warning },
          attention: { default: attention, subtle: attention }
        }
      },
      emphasis: {
        backgroundColor: bgEmphasis,
        foregroundColors: {
          default: { default: fg, subtle: fgMuted },
          accent: { default: accent, subtle: accent },
          good: { default: good, subtle: good },
          warning: { default: warning, subtle: warning },
          attention: { default: attention, subtle: attention }
        }
      }
    },
    actions: {
      actionsOrientation: "horizontal",
      actionAlignment: "left",
      buttonSpacing: 8,
      maxActions: 5,
      showCard: { actionMode: "inline" },
      spacing: "default"
    },
    adaptiveCard: {
      allowCustomStyle: false
    },
    spacing: {
      small: 4,
      default: 8,
      medium: 12,
      large: 16,
      extraLarge: 24,
      padding: 12
    },
    separator: {
      lineThickness: 1,
      lineColor: border
    },
    fontSizes: {
      small: 12,
      default: 14,
      medium: 16,
      large: 18,
      extraLarge: 22
    },
    fontWeights: {
      lighter: 300,
      default: 400,
      bolder: 600
    },
    imageSizes: {
      small: 40,
      medium: 80,
      large: 120
    },
    textBlock: {
      headingLevel: 2
    }
  };
}

// web/src/ui/adaptive-card-renderer.ts
var SUPPORTED_VERSIONS = new Set(["1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6"]);
var sdkLoaded = false;
var sdkLoadPromise = null;
var markdownProcessorConfigured = false;
function clearAdaptiveCardNotice(container) {
  container.querySelector(".adaptive-card-notice")?.remove();
}
function showAdaptiveCardNotice(container, message, tone = "error") {
  clearAdaptiveCardNotice(container);
  const notice = document.createElement("div");
  notice.className = `adaptive-card-notice adaptive-card-notice-${tone}`;
  notice.textContent = message;
  container.appendChild(notice);
}
function processAdaptiveCardMarkdown(text, renderer = (source) => renderMarkdown(source, null)) {
  const source = typeof text === "string" ? text : String(text ?? "");
  if (!source.trim()) {
    return { outputHtml: "", didProcess: false };
  }
  return {
    outputHtml: renderer(source),
    didProcess: true
  };
}
function createAdaptiveCardMarkdownProcessor(renderer = (source) => renderMarkdown(source, null)) {
  return (text, result) => {
    try {
      const processed = processAdaptiveCardMarkdown(text, renderer);
      result.outputHtml = processed.outputHtml;
      result.didProcess = processed.didProcess;
    } catch (error) {
      console.error("[adaptive-card] Failed to process markdown:", error);
      result.outputHtml = String(text ?? "");
      result.didProcess = false;
    }
  };
}
function ensureAdaptiveCardMarkdownProcessor(AC) {
  if (markdownProcessorConfigured || !AC?.AdaptiveCard)
    return;
  AC.AdaptiveCard.onProcessMarkdown = createAdaptiveCardMarkdownProcessor();
  markdownProcessorConfigured = true;
}
async function ensureSdk() {
  if (sdkLoaded)
    return;
  if (sdkLoadPromise)
    return sdkLoadPromise;
  sdkLoadPromise = new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = "/static/js/vendor/adaptivecards.min.js";
    script.onload = () => {
      sdkLoaded = true;
      resolve();
    };
    script.onerror = () => reject(new Error("Failed to load adaptivecards SDK"));
    document.head.appendChild(script);
  });
  return sdkLoadPromise;
}
function getAC() {
  return globalThis.AdaptiveCards;
}
function isAdaptiveCardBlock(block) {
  if (!block || typeof block !== "object")
    return false;
  const b = block;
  return b.type === "adaptive_card" && typeof b.card_id === "string" && typeof b.schema_version === "string" && typeof b.payload === "object" && b.payload !== null;
}
function isSupportedVersion(version) {
  return SUPPORTED_VERSIONS.has(version);
}
function extractCardBlocks(contentBlocks) {
  if (!Array.isArray(contentBlocks))
    return [];
  return contentBlocks.filter(isAdaptiveCardBlock);
}
function normalizeAdaptiveCardAction(action) {
  const type = (typeof action?.getJsonTypeName === "function" ? action.getJsonTypeName() : "") || action?.constructor?.name || "Unknown";
  const title = (typeof action?.title === "string" ? action.title : "") || "";
  const url = (typeof action?.url === "string" ? action.url : "") || undefined;
  const data = action?.data ?? undefined;
  return { type, title, data, url, raw: action };
}
function formatSubmissionValue2(value) {
  if (value == null)
    return "";
  if (typeof value === "string")
    return value.trim();
  if (typeof value === "number")
    return String(value);
  if (typeof value === "boolean")
    return value ? "yes" : "no";
  if (Array.isArray(value)) {
    return value.map((item) => formatSubmissionValue2(item)).filter(Boolean).join(", ");
  }
  if (typeof value === "object") {
    const pairs = Object.entries(value).map(([key, inner]) => `${key}: ${formatSubmissionValue2(inner)}`).filter((entry) => !entry.endsWith(": "));
    return pairs.join(", ");
  }
  return String(value).trim();
}
function coerceInputValue(type, value, definition) {
  if (value == null)
    return value;
  if (type === "Input.Toggle") {
    if (typeof value === "boolean") {
      if (value)
        return definition?.valueOn ?? "true";
      return definition?.valueOff ?? "false";
    }
    return typeof value === "string" ? value : String(value);
  }
  if (type === "Input.ChoiceSet") {
    if (Array.isArray(value))
      return value.join(",");
    return typeof value === "string" ? value : String(value);
  }
  if (Array.isArray(value))
    return value.join(", ");
  if (typeof value === "object")
    return formatSubmissionValue2(value);
  return typeof value === "string" ? value : String(value);
}
function hydrateAdaptiveCardPayloadWithSubmission(payload, submission) {
  if (!payload || typeof payload !== "object")
    return payload;
  if (!submission || typeof submission !== "object" || Array.isArray(submission))
    return payload;
  const values = submission;
  const visit = (node) => {
    if (Array.isArray(node))
      return node.map((item) => visit(item));
    if (!node || typeof node !== "object")
      return node;
    const record = node;
    const hydrated = { ...record };
    if (typeof hydrated.id === "string" && hydrated.id in values && String(hydrated.type || "").startsWith("Input.")) {
      hydrated.value = coerceInputValue(hydrated.type, values[hydrated.id], hydrated);
    }
    for (const [key, value] of Object.entries(hydrated)) {
      if (Array.isArray(value) || value && typeof value === "object") {
        hydrated[key] = visit(value);
      }
    }
    return hydrated;
  };
  return visit(payload);
}
function formatAdaptiveCardTimestamp(value) {
  if (typeof value !== "string" || !value.trim())
    return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime()))
    return "";
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit"
  }).format(date);
}
function describeAdaptiveCardState(block) {
  if (block.state === "active")
    return null;
  const label = block.state === "completed" ? "Submitted" : block.state === "cancelled" ? "Cancelled" : "Failed";
  const submission = block.last_submission && typeof block.last_submission === "object" ? block.last_submission : null;
  const title = submission && typeof submission.title === "string" ? submission.title.trim() : "";
  const when = formatAdaptiveCardTimestamp(block.completed_at || submission?.submitted_at);
  const detail = [title || null, when || null].filter(Boolean).join(" · ") || null;
  return { label, detail };
}
async function renderAdaptiveCard(container, block, options) {
  if (!isSupportedVersion(block.schema_version)) {
    console.warn(`[adaptive-card] Unsupported schema version ${block.schema_version} for card ${block.card_id}`);
    return false;
  }
  try {
    await ensureSdk();
  } catch (err) {
    console.error("[adaptive-card] Failed to load SDK:", err);
    return false;
  }
  try {
    const AC = getAC();
    ensureAdaptiveCardMarkdownProcessor(AC);
    const card = new AC.AdaptiveCard;
    const themeValues = getAdaptiveCardThemeValues();
    card.hostConfig = new AC.HostConfig(buildHostConfig());
    const submissionData = block.last_submission && typeof block.last_submission === "object" ? block.last_submission.data : undefined;
    const payload = block.state === "active" ? block.payload : hydrateAdaptiveCardPayloadWithSubmission(block.payload, submissionData);
    card.parse(payload);
    card.onExecuteAction = (action) => {
      const normalizedAction = normalizeAdaptiveCardAction(action);
      if (options?.onAction) {
        clearAdaptiveCardNotice(container);
        container.classList.add("adaptive-card-busy");
        Promise.resolve(options.onAction(normalizedAction)).catch((error) => {
          console.error("[adaptive-card] Action failed:", error);
          const message = error instanceof Error ? error.message : String(error || "Action failed.");
          showAdaptiveCardNotice(container, message || "Action failed.", "error");
        }).finally(() => {
          container.classList.remove("adaptive-card-busy");
        });
      } else {
        console.log("[adaptive-card] Action executed (not wired yet):", normalizedAction);
      }
    };
    const rendered = card.render();
    if (!rendered) {
      console.warn(`[adaptive-card] Card ${block.card_id} rendered to null`);
      return false;
    }
    container.classList.add("adaptive-card-container");
    container.style.setProperty("--adaptive-card-button-text-color", themeValues.buttonTextColor);
    const stateMeta = describeAdaptiveCardState(block);
    if (stateMeta) {
      container.classList.add("adaptive-card-finished");
      const banner = document.createElement("div");
      banner.className = `adaptive-card-status adaptive-card-status-${block.state}`;
      const label = document.createElement("span");
      label.className = "adaptive-card-status-label";
      label.textContent = stateMeta.label;
      banner.appendChild(label);
      if (stateMeta.detail) {
        const detail = document.createElement("span");
        detail.className = "adaptive-card-status-detail";
        detail.textContent = stateMeta.detail;
        banner.appendChild(detail);
      }
      container.appendChild(banner);
    }
    clearAdaptiveCardNotice(container);
    container.appendChild(rendered);
    if (stateMeta) {
      lockAdaptiveCardInputs(rendered);
    }
    return true;
  } catch (err) {
    console.error(`[adaptive-card] Failed to render card ${block.card_id}:`, err);
    return false;
  }
}

// web/src/ui/generated-widget.ts
function getArtifact(block) {
  const artifact = block?.artifact || {};
  const kind = artifact.kind || block?.kind || null;
  if (kind !== "html" && kind !== "svg" && kind !== "session_tree")
    return null;
  if (kind === "html") {
    const html = typeof artifact.html === "string" ? artifact.html : typeof block?.html === "string" ? block.html : "";
    return html ? { kind, html } : null;
  }
  if (kind === "svg") {
    const svg = typeof artifact.svg === "string" ? artifact.svg : typeof block?.svg === "string" ? block.svg : "";
    return svg ? { kind, svg } : null;
  }
  const tree = artifact.tree && typeof artifact.tree === "object" ? artifact.tree : block?.tree && typeof block.tree === "object" ? block.tree : null;
  return { kind, tree };
}
function readFiniteNumber(value) {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}
function readOptionalString(value) {
  return typeof value === "string" && value.trim() ? value.trim() : null;
}
function normalizeCapabilities(input, interactiveFallback = false) {
  const values = Array.isArray(input) ? input : interactiveFallback ? ["interactive"] : [];
  const normalized = values.filter((value) => typeof value === "string").map((value) => value.trim().toLowerCase()).filter(Boolean);
  return Array.from(new Set(normalized));
}
var GENERATED_WIDGET_WINDOW_NAME_PREFIX = "__PICLAW_WIDGET_HOST__:";
function escapeJsonForInlineScript(value) {
  return JSON.stringify(value).replace(/</g, "\\u003c").replace(/>/g, "\\u003e").replace(/&/g, "\\u0026").replace(/\u2028/g, "\\u2028").replace(/\u2029/g, "\\u2029");
}
function buildGeneratedWidgetPayload(block, post) {
  if (!block || block.type !== "generated_widget")
    return null;
  const artifact = getArtifact(block);
  if (!artifact)
    return null;
  return {
    title: block.title || block.name || "Generated widget",
    subtitle: typeof block.subtitle === "string" ? block.subtitle : "",
    description: block.description || block.subtitle || "",
    originPostId: Number.isFinite(post?.id) ? post.id : null,
    originChatJid: typeof post?.chat_jid === "string" ? post.chat_jid : null,
    widgetId: block.widget_id || block.id || null,
    artifact,
    capabilities: normalizeCapabilities(block.capabilities, block.interactive === true),
    source: "timeline",
    status: "final"
  };
}
function canRenderGeneratedWidget(block) {
  return buildGeneratedWidgetPayload(block, null) !== null;
}
function getGeneratedWidgetSessionKey(widget) {
  const toolCallId = readOptionalString(widget?.toolCallId) || readOptionalString(widget?.tool_call_id);
  if (toolCallId)
    return toolCallId;
  const widgetId = readOptionalString(widget?.widgetId) || readOptionalString(widget?.widget_id);
  if (widgetId)
    return widgetId;
  const originPostId = readFiniteNumber(widget?.originPostId) ?? readFiniteNumber(widget?.origin_post_id);
  if (originPostId !== null)
    return `post:${originPostId}`;
  return null;
}
function isInteractiveGeneratedWidget(widget) {
  const artifact = widget?.artifact || {};
  const kind = artifact.kind || widget?.kind || null;
  const capabilities = Array.isArray(widget?.capabilities) ? widget.capabilities : [];
  const interactiveCapability = capabilities.some((value) => typeof value === "string" && value.trim().toLowerCase() === "interactive");
  return kind === "html" && (widget?.source === "live" || interactiveCapability);
}
function getGeneratedWidgetIframeSandbox(widget) {
  return isInteractiveGeneratedWidget(widget) ? "allow-downloads allow-scripts allow-same-origin" : "allow-downloads";
}
function getGeneratedWidgetInitPayload(widget) {
  return {
    title: readOptionalString(widget?.title) || "Generated widget",
    widgetId: readOptionalString(widget?.widgetId) || readOptionalString(widget?.widget_id),
    toolCallId: readOptionalString(widget?.toolCallId) || readOptionalString(widget?.tool_call_id),
    turnId: readOptionalString(widget?.turnId) || readOptionalString(widget?.turn_id),
    capabilities: Array.isArray(widget?.capabilities) ? widget.capabilities : [],
    source: widget?.source === "live" ? "live" : "timeline",
    status: readOptionalString(widget?.status) || "final"
  };
}
function getGeneratedWidgetHostPayload(widget) {
  return {
    ...getGeneratedWidgetInitPayload(widget),
    subtitle: readOptionalString(widget?.subtitle) || "",
    description: readOptionalString(widget?.description) || "",
    error: readOptionalString(widget?.error) || null,
    width: readFiniteNumber(widget?.width),
    height: readFiniteNumber(widget?.height),
    runtimeState: widget?.runtimeState && typeof widget.runtimeState === "object" ? widget.runtimeState : null
  };
}
function getGeneratedWidgetHostWindowName(widget) {
  return `${GENERATED_WIDGET_WINDOW_NAME_PREFIX}${JSON.stringify(getGeneratedWidgetHostPayload(widget))}`;
}
function getGeneratedWidgetEmptyStateMessage(widget) {
  const status = readOptionalString(widget?.status);
  if (status === "loading" || status === "streaming") {
    return "Widget is loading…";
  }
  if (status === "error") {
    return readOptionalString(widget?.error) || "Widget failed to load.";
  }
  if ((widget?.artifact?.kind || widget?.kind) === "session_tree") {
    return "Session tree widget is unavailable.";
  }
  return "Widget artifact is missing or unsupported.";
}
function buildWidgetBootstrapScript(widget) {
  const meta = getGeneratedWidgetInitPayload(widget);
  const safeMeta = escapeJsonForInlineScript(meta);
  return `<script>
(function () {
  const meta = ${safeMeta};
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

  const windowNamePrefix = ${escapeJsonForInlineScript(GENERATED_WIDGET_WINDOW_NAME_PREFIX)};
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
</script>`;
}
function buildWidgetSrcDoc(widget) {
  const artifact = widget?.artifact || {};
  const kind = artifact.kind || widget?.kind || null;
  const rawHtml = typeof artifact.html === "string" ? artifact.html : typeof widget?.html === "string" ? widget.html : "";
  const rawSvg = typeof artifact.svg === "string" ? artifact.svg : typeof widget?.svg === "string" ? widget.svg : "";
  const title = typeof widget?.title === "string" && widget.title.trim() ? widget.title.trim() : "Generated widget";
  const content = kind === "svg" ? rawSvg : rawHtml;
  if (!content)
    return "";
  const interactive = isInteractiveGeneratedWidget(widget);
  const csp = [
    "default-src 'none'",
    "img-src data: blob: https: http:",
    "style-src 'unsafe-inline'",
    "font-src 'self' data: https: http:",
    "media-src data: blob: https: http:",
    "connect-src 'none'",
    "frame-src 'none'",
    interactive ? "script-src 'unsafe-inline' 'self'" : "script-src 'none'",
    "object-src 'none'",
    "base-uri 'none'",
    "form-action 'none'"
  ].join("; ");
  const body = kind === "svg" ? `<div class="widget-svg-shell">${content}</div>` : content;
  const bootstrap = interactive ? buildWidgetBootstrapScript(widget) : "";
  return `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta http-equiv="Content-Security-Policy" content="${csp}" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>${title.replace(/[<&>]/g, "")}</title>
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
${bootstrap}
</head>
<body>${body}</body>
</html>`;
}

// web/src/components/body-portal.ts
function BodyPortal({ children, className = "" }) {
  const [host, setHost] = w0(null);
  r0(() => {
    if (typeof document === "undefined")
      return;
    const nextHost = document.createElement("div");
    if (className)
      nextHost.className = className;
    document.body.appendChild(nextHost);
    setHost(nextHost);
    return () => {
      try {
        c0(null, nextHost);
      } finally {
        nextHost.remove();
        setHost((current) => current === nextHost ? null : current);
      }
    };
  }, [className]);
  _0(() => {
    if (!host)
      return;
    c0(children, host);
    return;
  }, [children, host]);
  return null;
}

// web/src/components/image-modal.ts
function ImageModal({ src, onClose }) {
  r0(() => {
    const handleEsc = (e2) => {
      if (e2.key === "Escape")
        onClose();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [onClose]);
  return X1`
        <${BodyPortal} className="image-modal-portal-root">
            <div class="image-modal" onClick=${onClose}>
                <img src=${src} alt="Full size" />
            </div>
        </${BodyPortal}>
    `;
}

// web/src/components/file-pill.ts
function FilePill({
  prefix = "file",
  label,
  title,
  onRemove,
  onClick,
  removeTitle = "Remove",
  icon = "file"
}) {
  const pillClass = `${prefix}-file-pill`;
  const nameClass = `${prefix}-file-name`;
  const removeClass = `${prefix}-file-remove`;
  const iconSvg = icon === "message" ? X1`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>` : X1`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
      </svg>`;
  return X1`
    <span class=${pillClass} title=${title || label} onClick=${onClick}>
      ${iconSvg}
      <span class=${nameClass}>${label}</span>
      ${onRemove && X1`
        <button
          class=${removeClass}
          onClick=${(event) => {
    event.preventDefault();
    event.stopPropagation();
    onRemove();
  }}
          title=${removeTitle}
          type="button"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      `}
    </span>
  `;
}

// web/src/components/post-runtime-safety.ts
async function writeClipboardTextBestEffort(clipboard, value) {
  try {
    await clipboard?.writeText?.(value);
    return true;
  } catch (_error) {
    return false;
  }
}
function writeClipboardDataViaExecCommand(documentLike, payload) {
  const text = typeof payload?.text === "string" ? payload.text : "";
  const html = typeof payload?.html === "string" ? payload.html : "";
  if (!documentLike || !text || typeof documentLike.createElement !== "function" || typeof documentLike.execCommand !== "function") {
    return false;
  }
  let host = null;
  let copyHandled = false;
  const onCopy = (event) => {
    const clipboardData = event?.clipboardData;
    if (!clipboardData || typeof clipboardData.setData !== "function")
      return;
    clipboardData.setData("text/plain", text);
    if (html)
      clipboardData.setData("text/html", html);
    if (typeof event.preventDefault === "function")
      event.preventDefault();
    copyHandled = true;
  };
  try {
    host = documentLike.createElement("textarea");
    host.value = text;
    if (typeof host.setAttribute === "function")
      host.setAttribute("readonly", "");
    if (host.style) {
      host.style.position = "fixed";
      host.style.opacity = "0";
      host.style.pointerEvents = "none";
    }
    documentLike.body?.appendChild?.(host);
    if (typeof host.select === "function")
      host.select();
    if (typeof host.setSelectionRange === "function")
      host.setSelectionRange(0, host.value.length);
    documentLike.addEventListener?.("copy", onCopy, true);
    const commandResult = documentLike.execCommand("copy");
    return Boolean(copyHandled || commandResult);
  } catch {
    return false;
  } finally {
    documentLike.removeEventListener?.("copy", onCopy, true);
    if (host) {
      documentLike.body?.removeChild?.(host);
    }
  }
}
function normalizeSelectionNode(node) {
  if (!node || typeof node !== "object")
    return null;
  const maybeNode = node;
  if (typeof maybeNode.nodeType === "number" && maybeNode.nodeType === 3) {
    return maybeNode.parentNode || null;
  }
  return maybeNode;
}
function copyPlainTextSelectionFromElement(event, options) {
  const clipboardData = event?.clipboardData;
  const root = options?.root;
  const selection = options?.selection;
  if (!clipboardData || typeof clipboardData.setData !== "function" || !root || !selection)
    return false;
  if (selection.isCollapsed)
    return false;
  let intersectsRoot = false;
  const rangeCount = Number(selection.rangeCount || 0);
  if (rangeCount > 0 && typeof selection.getRangeAt === "function") {
    try {
      const range = selection.getRangeAt(0);
      if (range && typeof range.intersectsNode === "function") {
        intersectsRoot = Boolean(range.intersectsNode(root));
      }
    } catch {
      intersectsRoot = false;
    }
  }
  if (!intersectsRoot && typeof root.contains === "function") {
    const anchorNode = normalizeSelectionNode(selection.anchorNode);
    const focusNode = normalizeSelectionNode(selection.focusNode);
    intersectsRoot = Boolean(anchorNode && root.contains(anchorNode) || focusNode && root.contains(focusNode));
  }
  if (!intersectsRoot)
    return false;
  const text = typeof selection.toString === "function" ? String(selection.toString() || "").replace(/\u00a0/g, " ") : "";
  if (!text)
    return false;
  clipboardData.setData("text/plain", text);
  event?.preventDefault?.();
  return true;
}
function readSessionStorageFlagBestEffort(storage, key) {
  try {
    return Boolean(storage?.getItem?.(key));
  } catch (_error) {
    return false;
  }
}
function writeSessionStorageFlagBestEffort(storage, key, value) {
  try {
    storage?.setItem?.(key, value);
    return true;
  } catch (_error) {
    return false;
  }
}
function resolveLinkPreviewSiteName(siteName, safeUrl) {
  const normalizedSiteName = typeof siteName === "string" && siteName.trim() ? siteName.trim() : null;
  if (normalizedSiteName)
    return normalizedSiteName;
  if (!safeUrl)
    return null;
  try {
    return new URL(safeUrl).hostname;
  } catch (_error) {
    return safeUrl;
  }
}

// web/src/components/post.ts
function FileAttachment({ mediaId, onPreview }) {
  const [info, setInfo] = w0(null);
  r0(() => {
    getMediaInfo(mediaId).then(setInfo).catch((error) => {
      console.warn("[post] Failed to load attachment metadata for file card:", mediaId, error);
    });
  }, [mediaId]);
  if (!info)
    return null;
  const filename = info.filename || "file";
  const size = info.metadata?.size;
  const sizeStr = size ? formatFileSize(size) : "";
  const previewKind = getAttachmentPreviewKind(info.content_type, info.filename);
  const previewLabel = previewKind === "unsupported" ? "Details" : "Preview";
  return X1`
        <div class="file-attachment" onClick=${(e2) => e2.stopPropagation()}>
            <a href=${getMediaUrl(mediaId)} download=${filename} class="file-attachment-main">
                <svg class="file-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                    <line x1="16" y1="13" x2="8" y2="13"/>
                    <line x1="16" y1="17" x2="8" y2="17"/>
                    <polyline points="10 9 9 9 8 9"/>
                </svg>
                <div class="file-info">
                    <span class="file-name">${filename}</span>
                    <span class="file-meta-row">
                        ${sizeStr && X1`<span class="file-size">${sizeStr}</span>`}
                        ${info.content_type && X1`<span class="file-size">${info.content_type}</span>`}
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
                onClick=${(e2) => {
    e2.preventDefault();
    e2.stopPropagation();
    onPreview?.({ mediaId, info });
  }}
            >
                ${previewLabel}
            </button>
        </div>
    `;
}
function extractRecoveryMarkerBlocks(contentBlocks) {
  if (!Array.isArray(contentBlocks))
    return [];
  return contentBlocks.filter((block) => block && typeof block === "object" && block.type === "recovery_marker" && block.recovered);
}
function extractTimeoutMarkerBlocks(contentBlocks) {
  if (!Array.isArray(contentBlocks))
    return [];
  return contentBlocks.filter((block) => block && typeof block === "object" && block.type === "timeout_marker" && (block.timed_out ?? true));
}
var RECOVERY_CLASSIFIER_LABELS = {
  context_recover: "context limit exceeded",
  rate_limit: "rate limit hit",
  api_error: "API error",
  timeout: "request timeout",
  overloaded: "service overloaded",
  connection: "connection error"
};
function formatRecoveryChipTooltip(marker) {
  const attempts = Number(marker?.attempts_used || 0);
  const classifier = String(marker?.classifier || "").trim();
  const reason = RECOVERY_CLASSIFIER_LABELS[classifier] || (classifier ? classifier.replace(/_/g, " ") : "");
  const parts = ["Recovered automatically"];
  if (attempts > 1)
    parts[0] = `Recovered after ${attempts} attempts`;
  if (reason)
    parts.push(reason);
  return parts.join(" — ");
}
function formatTimeoutChipTooltip(marker) {
  const action = typeof marker?.tool_action_summary === "string" ? marker.tool_action_summary.trim() : "";
  return action ? `Turn timed out — ${action}` : "Turn timed out before the model finished responding";
}
function AttachmentPill({ attachment, onPreview }) {
  const mediaId = Number(attachment?.id);
  const [info, setInfo] = w0(null);
  r0(() => {
    if (!Number.isFinite(mediaId))
      return;
    getMediaInfo(mediaId).then(setInfo).catch((error) => {
      console.warn("[post] Failed to load attachment metadata for attachment pill:", mediaId, error);
    });
    return;
  }, [mediaId]);
  const filename = info?.filename || attachment.label || `attachment-${attachment.id}`;
  const downloadHref = Number.isFinite(mediaId) ? getMediaUrl(mediaId) : null;
  const previewKind = getAttachmentPreviewKind(info?.content_type, info?.filename || attachment?.label);
  const previewLabel = previewKind === "unsupported" ? "Details" : "Preview";
  return X1`
        <span class="attachment-pill" title=${filename}>
            ${downloadHref ? X1`
                    <a href=${downloadHref} download=${filename} class="attachment-pill-main" onClick=${(e2) => e2.stopPropagation()}>
                        <${FilePill}
                            prefix="post"
                            label=${attachment.label}
                            title=${filename}
                        />
                    </a>
                ` : X1`
                    <${FilePill}
                        prefix="post"
                        label=${attachment.label}
                        title=${filename}
                    />
                `}
            ${Number.isFinite(mediaId) && info && X1`
                <button
                    class="attachment-pill-preview"
                    type="button"
                    title=${previewLabel}
                    onClick=${(e2) => {
    e2.preventDefault();
    e2.stopPropagation();
    onPreview?.({ mediaId, info });
  }}
                >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12z"/>
                        <circle cx="12" cy="12" r="3"/>
                    </svg>
                </button>
            `}
        </span>
    `;
}
function AnnotationsBadge({ annotations }) {
  if (!annotations)
    return null;
  const { audience, priority, lastModified } = annotations;
  const formattedLastModified = lastModified ? formatTimestamp(lastModified) : null;
  return X1`
        <div class="content-annotations">
            ${audience && audience.length > 0 && X1`
                <span class="content-annotation">Audience: ${audience.join(", ")}</span>
            `}
            ${typeof priority === "number" && X1`
                <span class="content-annotation">Priority: ${priority}</span>
            `}
            ${formattedLastModified && X1`
                <span class="content-annotation">Updated: ${formattedLastModified}</span>
            `}
        </div>
    `;
}
function ResourceLinkBlock({ block }) {
  const name = block.title || block.name || block.uri;
  const description = block.description;
  const sizeStr = block.size ? formatFileSize(block.size) : "";
  const mimeType = block.mime_type || "";
  const icon = getMimeIcon(mimeType);
  const safeUrl = sanitizeUrl(block.uri);
  return X1`
        <a
            href=${safeUrl || "#"}
            class="resource-link"
            target=${safeUrl ? "_blank" : undefined}
            rel=${safeUrl ? "noopener noreferrer" : undefined}
            onClick=${(e2) => e2.stopPropagation()}>
            <div class="resource-link-main">
                <div class="resource-link-header">
                    <span class="resource-link-icon-inline">${icon}</span>
                    <div class="resource-link-title">${name}</div>
                </div>
                ${description && X1`<div class="resource-link-description">${description}</div>`}
                <div class="resource-link-meta">
                    ${mimeType && X1`<span>${mimeType}</span>`}
                    ${sizeStr && X1`<span>${sizeStr}</span>`}
                </div>
            </div>
            <div class="resource-link-icon">↗</div>
        </a>
    `;
}
function ResourceBlock({ block }) {
  const [open, setOpen] = w0(false);
  const title = block.uri || "Embedded resource";
  const contentText = block.text || "";
  const hasBlob = Boolean(block.data);
  const mimeType = block.mime_type || "";
  return X1`
        <div class="resource-embed">
            <button class="resource-embed-toggle" onClick=${(e2) => {
    e2.preventDefault();
    e2.stopPropagation();
    setOpen(!open);
  }}>
                ${open ? "▼" : "▶"} ${title}
            </button>
            ${open && X1`
                ${contentText && X1`<pre class="resource-embed-content">${contentText}</pre>`}
                ${hasBlob && X1`
                    <div class="resource-embed-blob">
                        <span class="resource-embed-blob-label">Embedded blob</span>
                        ${mimeType && X1`<span class="resource-embed-blob-meta">${mimeType}</span>`}
                        <button class="resource-embed-blob-btn" onClick=${(e2) => {
    e2.preventDefault();
    e2.stopPropagation();
    const blob = new Blob([Uint8Array.from(atob(block.data), (c2) => c2.charCodeAt(0))], { type: mimeType || "application/octet-stream" });
    const url = URL.createObjectURL(blob);
    const a2 = document.createElement("a");
    a2.href = url;
    a2.download = title.split("/").pop() || "resource";
    a2.click();
    URL.revokeObjectURL(url);
  }}>Download</button>
                    </div>
                `}
            `}
        </div>
    `;
}
function GeneratedWidgetLaunch({ block, post, onOpenWidget }) {
  if (!block)
    return null;
  const payload = buildGeneratedWidgetPayload(block, post);
  const supportsRender = canRenderGeneratedWidget(block);
  const kind = payload?.artifact?.kind || block?.artifact?.kind || block?.kind || null;
  const title = payload?.title || block.title || block.name || "Generated widget";
  const description = payload?.description || block.description || block.subtitle || "";
  const openLabel = block.open_label || "Open widget";
  const autoOpened = o0(false);
  const launchWidget = (e2) => {
    if (e2) {
      e2.preventDefault();
      e2.stopPropagation();
    }
    if (!payload)
      return;
    onOpenWidget?.(payload);
  };
  r0(() => {
    if (!block?.auto_open || !payload || !supportsRender || autoOpened.current)
      return;
    const postTime = post?.timestamp ? new Date(post.timestamp).getTime() : 0;
    if (postTime && Date.now() - postTime > 1e4)
      return;
    const key = `widget_opened_${block.widget_id || post?.id || ""}`;
    if (readSessionStorageFlagBestEffort(sessionStorage, key))
      return;
    autoOpened.current = true;
    writeSessionStorageFlagBestEffort(sessionStorage, key, "1");
    onOpenWidget?.(payload);
  }, [block?.auto_open, payload, supportsRender]);
  return X1`
        <div class="generated-widget-launch" onClick=${(e2) => e2.stopPropagation()}>
            <div class="generated-widget-launch-header">
                <div class="generated-widget-launch-eyebrow">Generated widget${kind ? ` • ${String(kind).toUpperCase()}` : ""}</div>
                <div class="generated-widget-launch-title">${title}</div>
            </div>
            ${description && X1`<div class="generated-widget-launch-description">${description}</div>`}
            <div class="generated-widget-launch-actions">
                <button
                    class="generated-widget-launch-btn"
                    type="button"
                    disabled=${!supportsRender}
                    onClick=${launchWidget}
                    title=${supportsRender ? "Open widget in a floating pane" : "Unsupported widget artifact"}
                >
                    ${openLabel}
                </button>
                <span class="generated-widget-launch-note">
                    ${supportsRender ? "Opens in a dismissible floating pane." : "This widget artifact is missing or unsupported."}
                </span>
            </div>
        </div>
    `;
}
function getMimeIcon(mimeType) {
  if (!mimeType)
    return "\uD83D\uDCCE";
  if (mimeType.startsWith("image/"))
    return "\uD83D\uDDBC️";
  if (mimeType.startsWith("audio/"))
    return "\uD83C\uDFB5";
  if (mimeType.startsWith("video/"))
    return "\uD83C\uDFAC";
  if (mimeType.includes("pdf"))
    return "\uD83D\uDCC4";
  if (mimeType.includes("zip") || mimeType.includes("gzip"))
    return "\uD83D\uDDDC️";
  if (mimeType.startsWith("text/"))
    return "\uD83D\uDCC4";
  return "\uD83D\uDCCE";
}
function buildLinkPreviewBackgroundStyle(imageUrl) {
  const safeImage = sanitizeUrl(imageUrl, { allowDataImage: true });
  return safeImage ? { backgroundImage: `url("${safeImage}")` } : undefined;
}
function LinkPreview({ preview }) {
  const safeUrl = sanitizeUrl(preview.url);
  const bgStyle = buildLinkPreviewBackgroundStyle(preview.image);
  const siteName = resolveLinkPreviewSiteName(preview.site_name, safeUrl);
  return X1`
        <a
            href=${safeUrl || "#"}
            class="link-preview ${bgStyle ? "has-image" : ""}"
            target=${safeUrl ? "_blank" : undefined}
            rel=${safeUrl ? "noopener noreferrer" : undefined}
            onClick=${(e2) => e2.stopPropagation()}
            style=${bgStyle}>
            <div class="link-preview-overlay">
                <div class="link-preview-site">${siteName || ""}</div>
                <div class="link-preview-title">${preview.title}</div>
                ${preview.description && X1`
                    <div class="link-preview-description">${preview.description}</div>
                `}
            </div>
        </a>
    `;
}
function getDisplayContent(content, _linkPreviews) {
  return typeof content === "string" ? content : "";
}
var CODE_COPY_RESET_MS = 1800;
var COPY_ICON_SVG = `
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>`;
var COPY_SUCCESS_SVG = `
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path d="M20 6L9 17l-5-5"></path>
    </svg>`;
var COPY_ERROR_SVG = `
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <circle cx="12" cy="12" r="9"></circle>
        <path d="M9 9l6 6M15 9l-6 6"></path>
    </svg>`;
var CLIPBOARD_STYLE = `
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
</style>`;
async function copyTextToClipboard(text) {
  const value = typeof text === "string" ? text : "";
  if (!value)
    return false;
  if (writeClipboardDataViaExecCommand(document, { text: value })) {
    return true;
  }
  if (await writeClipboardTextBestEffort(navigator.clipboard, value)) {
    return true;
  }
  try {
    const textarea = document.createElement("textarea");
    textarea.value = value;
    textarea.setAttribute("readonly", "");
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    textarea.style.pointerEvents = "none";
    document.body.appendChild(textarea);
    textarea.select();
    textarea.setSelectionRange(0, textarea.value.length);
    const copied = document.execCommand("copy");
    document.body.removeChild(textarea);
    return copied;
  } catch {
    return false;
  }
}
async function copyMessageToClipboard(markdown) {
  const value = typeof markdown === "string" ? markdown : "";
  if (!value)
    return false;
  const bodyHtml = renderMarkdown(value, null);
  const htmlDoc = `<html><head>${CLIPBOARD_STYLE}</head><body>${bodyHtml}</body></html>`;
  if (writeClipboardDataViaExecCommand(document, { text: value, html: htmlDoc })) {
    return true;
  }
  if (navigator.clipboard?.write && typeof ClipboardItem !== "undefined") {
    try {
      const item = new ClipboardItem({
        "text/plain": new Blob([value], { type: "text/plain" }),
        "text/html": new Blob([htmlDoc], { type: "text/html" })
      });
      await navigator.clipboard.write([item]);
      return true;
    } catch (error) {
      console.warn("[post] Rich clipboard write failed, falling back to plain text copy.", error);
    }
  }
  return copyTextToClipboard(value);
}
function enhanceCodeBlocks(container) {
  if (!container)
    return () => {};
  const blocks = Array.from(container.querySelectorAll("pre")).filter((pre) => pre.querySelector("code"));
  if (blocks.length === 0)
    return () => {};
  const resetTimers = new Map;
  const cleanups = [];
  const handleDocumentCopy = (event) => {
    const selection = window.getSelection?.();
    if (!selection || selection.isCollapsed)
      return;
    for (const pre of blocks) {
      if (copyPlainTextSelectionFromElement(event, { root: pre, selection })) {
        return;
      }
    }
  };
  document.addEventListener("copy", handleDocumentCopy, true);
  cleanups.push(() => document.removeEventListener("copy", handleDocumentCopy, true));
  const setButtonState = (button, state) => {
    const nextState = state || "idle";
    button.dataset.copyState = nextState;
    if (nextState === "success") {
      button.innerHTML = COPY_SUCCESS_SVG;
      button.setAttribute("aria-label", "Copied");
      button.setAttribute("title", "Copied");
      button.classList.add("is-success");
      button.classList.remove("is-error");
    } else if (nextState === "error") {
      button.innerHTML = COPY_ERROR_SVG;
      button.setAttribute("aria-label", "Copy failed");
      button.setAttribute("title", "Copy failed");
      button.classList.add("is-error");
      button.classList.remove("is-success");
    } else {
      button.innerHTML = COPY_ICON_SVG;
      button.setAttribute("aria-label", "Copy code");
      button.setAttribute("title", "Copy code");
      button.classList.remove("is-success", "is-error");
    }
  };
  blocks.forEach((pre) => {
    const wrapper = document.createElement("div");
    wrapper.className = "post-code-block";
    pre.parentNode?.insertBefore(wrapper, pre);
    wrapper.appendChild(pre);
    const button = document.createElement("button");
    button.type = "button";
    button.className = "post-code-copy-btn";
    setButtonState(button, "idle");
    wrapper.appendChild(button);
    const handleCopyClick = async (event) => {
      event.preventDefault();
      event.stopPropagation();
      const code = pre.querySelector("code");
      const text = code?.textContent || "";
      const ok = await copyTextToClipboard(text);
      setButtonState(button, ok ? "success" : "error");
      const existingTimer = resetTimers.get(button);
      if (existingTimer)
        clearTimeout(existingTimer);
      const timer = setTimeout(() => {
        setButtonState(button, "idle");
        resetTimers.delete(button);
      }, CODE_COPY_RESET_MS);
      resetTimers.set(button, timer);
    };
    button.addEventListener("click", handleCopyClick);
    cleanups.push(() => {
      button.removeEventListener("click", handleCopyClick);
      const timer = resetTimers.get(button);
      if (timer)
        clearTimeout(timer);
      if (wrapper.parentNode) {
        wrapper.parentNode.insertBefore(pre, wrapper);
        wrapper.remove();
      }
    });
  });
  return () => {
    cleanups.forEach((cleanup) => cleanup());
  };
}
function extractFileRefs(content) {
  if (!content)
    return { content, fileRefs: [] };
  const normalized = content.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  let start = -1;
  for (let i2 = 0;i2 < lines.length; i2 += 1) {
    if (lines[i2].trim() === "Files:" && lines[i2 + 1] && /^\s*-\s+/.test(lines[i2 + 1])) {
      start = i2;
      break;
    }
  }
  if (start === -1)
    return { content, fileRefs: [] };
  const refs = [];
  let end = start + 1;
  for (;end < lines.length; end += 1) {
    const line = lines[end];
    if (/^\s*-\s+/.test(line)) {
      refs.push(line.replace(/^\s*-\s+/, "").trim());
    } else if (!line.trim()) {
      break;
    } else {
      break;
    }
  }
  if (refs.length === 0)
    return { content, fileRefs: [] };
  const before = lines.slice(0, start);
  const after = lines.slice(end);
  let cleaned = [...before, ...after].join(`
`);
  cleaned = cleaned.replace(/\n{3,}/g, `

`).trim();
  return { content: cleaned, fileRefs: refs };
}
function extractMessageRefs(content) {
  if (!content)
    return { content, messageRefs: [] };
  const normalized = content.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  let start = -1;
  for (let i2 = 0;i2 < lines.length; i2 += 1) {
    if (lines[i2].trim() === "Referenced messages:" && lines[i2 + 1] && /^\s*-\s+/.test(lines[i2 + 1])) {
      start = i2;
      break;
    }
  }
  if (start === -1)
    return { content, messageRefs: [] };
  const refs = [];
  let end = start + 1;
  for (;end < lines.length; end += 1) {
    const line = lines[end];
    if (/^\s*-\s+/.test(line)) {
      const val = line.replace(/^\s*-\s+/, "").trim();
      const match = val.match(/^message:(\S+)$/i);
      if (match)
        refs.push(match[1]);
    } else if (!line.trim()) {
      break;
    } else {
      break;
    }
  }
  if (refs.length === 0)
    return { content, messageRefs: [] };
  const before = lines.slice(0, start);
  const after = lines.slice(end);
  let cleaned = [...before, ...after].join(`
`);
  cleaned = cleaned.replace(/\n{3,}/g, `

`).trim();
  return { content: cleaned, messageRefs: refs };
}
function extractAttachmentRefs(content) {
  if (!content)
    return { content, attachments: [] };
  const normalized = content.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  let start = -1;
  for (let i2 = 0;i2 < lines.length; i2 += 1) {
    const header = lines[i2].trim();
    if ((header === "Images:" || header === "Attachments:") && lines[i2 + 1] && /^\s*-\s+/.test(lines[i2 + 1])) {
      start = i2;
      break;
    }
  }
  if (start === -1)
    return { content, attachments: [] };
  const refs = [];
  let end = start + 1;
  for (;end < lines.length; end += 1) {
    const line = lines[end];
    if (/^\s*-\s+/.test(line)) {
      const raw = line.replace(/^\s*-\s+/, "").trim();
      const match = raw.match(/^attachment:([^\s)]+)\s*(?:\((.+)\))?$/i) || raw.match(/^attachment:([^\s]+)\s+(.+)$/i);
      if (match) {
        const id = match[1];
        const label = (match[2] || "").trim() || id;
        refs.push({ id, label, raw });
      } else {
        refs.push({ id: null, label: raw, raw });
      }
    } else if (!line.trim()) {
      break;
    } else {
      break;
    }
  }
  if (refs.length === 0)
    return { content, attachments: [] };
  const before = lines.slice(0, start);
  const after = lines.slice(end);
  let cleaned = [...before, ...after].join(`
`);
  cleaned = cleaned.replace(/\n{3,}/g, `

`).trim();
  return { content: cleaned, attachments: refs };
}
function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
function highlightHtml(html, query) {
  if (!html || !query)
    return html;
  const terms = String(query).trim().split(/\s+/).filter(Boolean);
  if (terms.length === 0)
    return html;
  const escapedTerms = terms.map(escapeRegex).sort((a2, b) => b.length - a2.length);
  const pattern = new RegExp(`(${escapedTerms.join("|")})`, "gi");
  const matcher = new RegExp(`^(${escapedTerms.join("|")})$`, "i");
  const doc = new DOMParser().parseFromString(html, "text/html");
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
  const nodes = [];
  let node;
  while (node = walker.nextNode())
    nodes.push(node);
  for (const textNode of nodes) {
    const value = textNode.nodeValue;
    if (!value || !pattern.test(value)) {
      pattern.lastIndex = 0;
      continue;
    }
    pattern.lastIndex = 0;
    const parent = textNode.parentElement;
    if (parent && parent.closest("code, pre, script, style"))
      continue;
    const parts = value.split(pattern).filter((part) => part !== "");
    if (parts.length === 0)
      continue;
    const frag = doc.createDocumentFragment();
    for (const part of parts) {
      if (matcher.test(part)) {
        const mark = doc.createElement("mark");
        mark.className = "search-highlight-term";
        mark.textContent = part;
        frag.appendChild(mark);
      } else {
        frag.appendChild(doc.createTextNode(part));
      }
    }
    textNode.parentNode.replaceChild(frag, textNode);
  }
  return doc.body.innerHTML;
}
function Post({ post, onClick, onHashtagClick, onMessageRef, onScrollToMessage, agentName, agentAvatarUrl, userName, userAvatarUrl, userAvatarBackground, onDelete, isThreadReply, isThreadPrev, isThreadNext, isRemoving, highlightQuery, onFileRef, onOpenWidget, onOpenAttachmentPreview }) {
  const [zoomedImage, setZoomedImage] = w0(null);
  const [copyState, setCopyState] = w0("idle");
  const contentRef = o0(null);
  const copyResetTimerRef = o0(null);
  const data = post.data;
  const isAgent = data.type === "agent_response";
  const resolvedUserName = userName || "You";
  const displayName = isAgent ? agentName || DEFAULT_AGENT_NAME : resolvedUserName;
  const searchChatAgentName = typeof post.chat_agent_name === "string" ? post.chat_agent_name.trim() : "";
  const showSearchChatAgentTag = Boolean(isAgent && highlightQuery && searchChatAgentName && searchChatAgentName !== displayName);
  const avatarInfo = isAgent ? getAvatarInfo(agentName, agentAvatarUrl, true) : getAvatarInfo(resolvedUserName, userAvatarUrl);
  const normalizedUserBackground = typeof userAvatarBackground === "string" ? userAvatarBackground.trim().toLowerCase() : "";
  const clearUserBackground = !isAgent && avatarInfo.image && (normalizedUserBackground === "clear" || normalizedUserBackground === "transparent");
  const clearAgentBackground = isAgent && Boolean(avatarInfo.image);
  const avatarStyle = `background-color: ${clearUserBackground || clearAgentBackground ? "transparent" : avatarInfo.color}`;
  const contentMeta = data.content_meta;
  const isTruncated = Boolean(contentMeta?.truncated);
  const isPreview = Boolean(contentMeta?.preview);
  const isHardTruncated = isTruncated && !isPreview;
  const truncatedInfo = isTruncated ? {
    originalLength: Number.isFinite(contentMeta?.original_length) ? contentMeta.original_length : data.content ? data.content.length : 0,
    maxLength: Number.isFinite(contentMeta?.max_length) ? contentMeta.max_length : 0
  } : null;
  const blocks = data.content_blocks || [];
  const mediaIds = data.media_ids || [];
  let displayContent = getDisplayContent(data.content, data.link_previews);
  const { content: cleanedContent, fileRefs } = extractFileRefs(displayContent);
  const { content: cleanedWithMsgRefs, messageRefs } = extractMessageRefs(cleanedContent);
  const { content: cleanedWithAttachments, attachments } = extractAttachmentRefs(cleanedWithMsgRefs);
  displayContent = cleanedWithAttachments;
  const directCardBlocks = extractCardBlocks(blocks);
  const submissionBlocks = extractAdaptiveCardSubmissionBlocks(blocks);
  const recoveryMarkerBlocks = extractRecoveryMarkerBlocks(blocks);
  const recoveryMarker = recoveryMarkerBlocks[0] || null;
  const timeoutMarkerBlocks = extractTimeoutMarkerBlocks(blocks);
  const timeoutMarker = timeoutMarkerBlocks[0] || null;
  const singleCardFallback = directCardBlocks.length === 1 && typeof directCardBlocks[0]?.fallback_text === "string" ? directCardBlocks[0].fallback_text.trim() : "";
  const singleSubmissionFallback = submissionBlocks.length === 1 ? buildAdaptiveCardSubmissionFallbackText(submissionBlocks[0]).trim() : "";
  const hideRenderedFallback = Boolean(singleCardFallback) && displayContent?.trim() === singleCardFallback || Boolean(singleSubmissionFallback) && displayContent?.trim() === singleSubmissionFallback;
  const shouldRenderContent = Boolean(displayContent) && !isHardTruncated && !hideRenderedFallback;
  const highlightQueryText = typeof highlightQuery === "string" ? highlightQuery.trim() : "";
  const renderedHtml = G0(() => {
    if (!displayContent || hideRenderedFallback)
      return "";
    const baseHtml = renderMarkdown(displayContent, onHashtagClick);
    return highlightQueryText ? highlightHtml(baseHtml, highlightQueryText) : baseHtml;
  }, [displayContent, hideRenderedFallback, highlightQueryText]);
  const markdownCopyPayload = G0(() => buildPostMarkdownCopyPayload(post), [post]);
  const handleImageClick = (e2, mediaId) => {
    e2.stopPropagation();
    setZoomedImage(getMediaUrl(mediaId));
  };
  const handleAttachmentPreview = (attachment) => {
    onOpenAttachmentPreview?.(attachment);
  };
  const handleDeleteClick = (e2) => {
    e2.stopPropagation();
    onDelete?.(post);
  };
  const handleCopyMarkdownClick = async (e2) => {
    e2.stopPropagation();
    const ok = await copyMessageToClipboard(markdownCopyPayload);
    setCopyState(ok ? "success" : "error");
    if (copyResetTimerRef.current)
      clearTimeout(copyResetTimerRef.current);
    copyResetTimerRef.current = setTimeout(() => {
      copyResetTimerRef.current = null;
      setCopyState("idle");
    }, CODE_COPY_RESET_MS);
  };
  const resolveInlineAttachments = (content, attachments2) => {
    const usedIds2 = new Set;
    if (!content || attachments2.length === 0) {
      return { content, usedIds: usedIds2 };
    }
    const replaced = content.replace(/attachment:([^\s)"']+)/g, (match, rawRef, offset, source) => {
      const ref = rawRef.replace(/^\/+/, "");
      const byName = attachments2.find((entry2) => entry2.name && entry2.name.toLowerCase() === ref.toLowerCase() && !usedIds2.has(entry2.id));
      const entry = byName || attachments2.find((item) => !usedIds2.has(item.id));
      if (!entry)
        return match;
      usedIds2.add(entry.id);
      const prefix = source.slice(Math.max(0, offset - 2), offset);
      if (prefix === "](") {
        return `/media/${entry.id}`;
      }
      return entry.name || "attachment";
    });
    return { content: replaced, usedIds: usedIds2 };
  };
  const imageItems = [];
  const fileIds = [];
  const attachmentEntries = [];
  const resourceLinks = [];
  const resources = [];
  const generatedWidgets = [];
  const textAnnotations = [];
  let mediaIndex = 0;
  if (blocks.length > 0) {
    blocks.forEach((block) => {
      if (block?.type === "text" && block.annotations) {
        textAnnotations.push(block.annotations);
      }
      if (block?.type === "generated_widget") {
        generatedWidgets.push(block);
      } else if (block?.type === "resource_link") {
        resourceLinks.push(block);
      } else if (block?.type === "resource") {
        resources.push(block);
      } else if (block?.type === "file") {
        const id = mediaIds[mediaIndex++];
        if (id) {
          fileIds.push(id);
          attachmentEntries.push({ id, name: block?.name || block?.filename || block?.title });
        }
      } else if (block?.type === "image" || !block?.type) {
        const id = mediaIds[mediaIndex++];
        if (id) {
          const mimeType = typeof block?.mime_type === "string" ? block.mime_type : undefined;
          imageItems.push({ id, annotations: block?.annotations, mimeType });
          attachmentEntries.push({ id, name: block?.name || block?.filename || block?.title });
        }
      }
    });
  } else if (mediaIds.length > 0) {
    const treatAsFiles = attachments.length > 0;
    mediaIds.forEach((id, index) => {
      const ref = attachments[index] || null;
      attachmentEntries.push({ id, name: ref?.label || null });
      if (treatAsFiles) {
        fileIds.push(id);
      } else {
        imageItems.push({ id, annotations: null });
      }
    });
  }
  if (attachments.length > 0) {
    attachments.forEach((ref) => {
      if (!ref?.id)
        return;
      const match = attachmentEntries.find((entry) => String(entry.id) === String(ref.id));
      if (match && !match.name) {
        match.name = ref.label;
      }
    });
  }
  const { content: resolvedContent, usedIds } = resolveInlineAttachments(displayContent, attachmentEntries);
  displayContent = resolvedContent;
  const filteredImageItems = imageItems.filter(({ id }) => !usedIds.has(id));
  const filteredFileIds = fileIds.filter((id) => !usedIds.has(id));
  const attachmentPills = attachments.length > 0 ? attachments.map((ref, idx) => ({
    id: ref.id || `attachment-${idx + 1}`,
    label: ref.label || `attachment-${idx + 1}`
  })) : attachmentEntries.map((entry, idx) => ({
    id: entry.id,
    label: entry.name || `attachment-${idx + 1}`
  }));
  const cardBlocks = G0(() => extractCardBlocks(blocks), [blocks]);
  const cardSubmissionBlocks = G0(() => extractAdaptiveCardSubmissionBlocks(blocks), [blocks]);
  const cardBlocksKey = G0(() => {
    return cardBlocks.map((b) => `${b.card_id}:${b.state}`).join("|");
  }, [cardBlocks]);
  r0(() => {
    if (!contentRef.current)
      return;
    renderMermaidDiagrams(contentRef.current);
    return enhanceCodeBlocks(contentRef.current);
  }, [renderedHtml]);
  r0(() => () => {
    if (copyResetTimerRef.current)
      clearTimeout(copyResetTimerRef.current);
  }, []);
  const cardContainerRef = o0(null);
  r0(() => {
    if (!cardContainerRef.current || cardBlocks.length === 0)
      return;
    const container = cardContainerRef.current;
    container.innerHTML = "";
    for (const block of cardBlocks) {
      const cardEl = document.createElement("div");
      container.appendChild(cardEl);
      renderAdaptiveCard(cardEl, block, {
        onAction: async (action) => {
          if (action.type === "Action.OpenUrl") {
            const safeUrl = sanitizeUrl(action.url || "");
            if (!safeUrl)
              throw new Error("Invalid URL");
            window.open(safeUrl, "_blank", "noopener,noreferrer");
            return;
          }
          if (action.type === "Action.Submit") {
            await submitAdaptiveCardAction({
              post_id: post.id,
              thread_id: data.thread_id || post.id,
              chat_jid: post.chat_jid || null,
              card_id: block.card_id,
              action: {
                type: action.type,
                title: action.title || "",
                data: action.data
              }
            });
            return;
          }
          console.warn("[post] unsupported adaptive card action:", action.type, action);
        }
      }).catch((err) => {
        console.error("[post] adaptive card render error:", err);
        cardEl.textContent = block.fallback_text || "Card failed to render.";
      });
    }
  }, [cardBlocksKey, post.id]);
  return X1`
        <div id=${`post-${post.id}`} class="post ${isAgent ? "agent-post" : ""} ${isThreadReply ? "thread-reply" : ""} ${isThreadPrev ? "thread-prev" : ""} ${isThreadNext ? "thread-next" : ""} ${isRemoving ? "removing" : ""}" onClick=${onClick}>
            <div class="post-avatar ${isAgent ? "agent-avatar" : ""} ${avatarInfo.image ? "has-image" : ""}" style=${avatarStyle}>
                ${avatarInfo.image ? X1`<img src=${avatarInfo.image} alt=${displayName} />` : avatarInfo.letter}
            </div>
            <div class="post-body">
                <div class="post-actions">
                    <button
                        class=${`post-action-btn post-copy-btn${copyState === "success" ? " is-success" : copyState === "error" ? " is-error" : ""}`}
                        type="button"
                        title=${copyState === "success" ? "Copied" : copyState === "error" ? "Copy failed" : "Copy message"}
                        aria-label=${copyState === "success" ? "Copied" : copyState === "error" ? "Copy failed" : "Copy message"}
                        onClick=${handleCopyMarkdownClick}
                        disabled=${!markdownCopyPayload}
                    >
                        ${copyState === "success" ? X1`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M20 6L9 17l-5-5"></path></svg>` : copyState === "error" ? X1`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><circle cx="12" cy="12" r="9"></circle><path d="M9 9l6 6M15 9l-6 6"></path></svg>` : X1`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><rect x="9" y="9" width="10" height="10" rx="2"></rect><path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path></svg>`}
                    </button>
                    <button
                        class="post-action-btn post-delete-btn"
                        type="button"
                        title="Delete message"
                        aria-label="Delete message"
                        onClick=${handleDeleteClick}
                    >
                        <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                            <path d="M18 6L6 18M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                <div class="post-meta">
                    <span class="post-author">${displayName}</span>
                    ${showSearchChatAgentTag && X1`<span class="post-chat-agent-tag" title=${`Chat: ${searchChatAgentName}`}>@${searchChatAgentName}</span>`}
                    ${recoveryMarker && X1`
                        <span
                            class="post-recovery-chip"
                            title=${formatRecoveryChipTooltip(recoveryMarker)}
                        >
                            recovered
                        </span>
                    `}
                    ${timeoutMarker && X1`
                        <span
                            class="post-recovery-chip post-timeout-chip"
                            title=${formatTimeoutChipTooltip(timeoutMarker)}
                        >
                            timeout
                        </span>
                    `}
                    <a class="post-time" href=${`#msg-${post.id}`} onClick=${(e2) => {
    e2.preventDefault();
    e2.stopPropagation();
    if (onMessageRef)
      onMessageRef(post.id);
  }}>${formatTime(post.timestamp)}</a>
                </div>
                ${isHardTruncated && truncatedInfo && X1`
                    <div class="post-content truncated">
                        <div class="truncated-title">Message too large to display.</div>
                        <div class="truncated-meta">
                            Original length: ${formatCount(truncatedInfo.originalLength)} chars
                            ${truncatedInfo.maxLength ? X1` • Display limit: ${formatCount(truncatedInfo.maxLength)} chars` : ""}
                        </div>
                    </div>
                `}
                ${isPreview && truncatedInfo && X1`
                    <div class="post-content preview">
                        <div class="truncated-title">Preview truncated.</div>
                        <div class="truncated-meta">
                            Showing first ${formatCount(truncatedInfo.maxLength)} of ${formatCount(truncatedInfo.originalLength)} chars. Download full text below.
                        </div>
                    </div>
                `}
                ${(fileRefs.length > 0 || messageRefs.length > 0 || attachmentPills.length > 0) && X1`
                    <div class="post-file-refs">
                        ${messageRefs.map((id) => {
    const scrollToRef = (e2) => {
      e2.preventDefault();
      e2.stopPropagation();
      if (onScrollToMessage) {
        onScrollToMessage(id, post.chat_jid || null);
      } else {
        const el = document.getElementById("post-" + id);
        if (el) {
          el.scrollIntoView({ behavior: "smooth", block: "center" });
          el.classList.add("post-highlight");
          setTimeout(() => el.classList.remove("post-highlight"), 2000);
        }
      }
    };
    return X1`
                                <a href=${`#msg-${id}`} class="post-msg-pill-link" onClick=${scrollToRef}>
                                    <${FilePill}
                                        prefix="post"
                                        label=${"msg:" + id}
                                        title=${"Message " + id}
                                        icon="message"
                                        onClick=${scrollToRef}
                                    />
                                </a>
                            `;
  })}
                        ${fileRefs.map((ref) => {
    const label = ref.split("/").pop() || ref;
    return X1`
                                <${FilePill}
                                    prefix="post"
                                    label=${label}
                                    title=${ref}
                                    onClick=${() => onFileRef?.(ref)}
                                />
                            `;
  })}
                        ${attachmentPills.map((attachment) => X1`
                            <${AttachmentPill}
                                key=${attachment.id}
                                attachment=${attachment}
                                onPreview=${handleAttachmentPreview}
                            />
                        `)}
                    </div>
                `}
                ${shouldRenderContent && X1`
                    <div 
                        ref=${contentRef}
                        class="post-content"
                        dangerouslySetInnerHTML=${{ __html: renderedHtml }}
                        onClick=${(e2) => {
    if (e2.target.classList.contains("hashtag")) {
      e2.preventDefault();
      e2.stopPropagation();
      const tag = e2.target.dataset.hashtag;
      if (tag)
        onHashtagClick?.(tag);
    } else if (e2.target.tagName === "IMG") {
      e2.preventDefault();
      e2.stopPropagation();
      setZoomedImage(e2.target.src);
    }
  }}
                    />
                `}
                ${cardBlocks.length > 0 && X1`
                    <div ref=${cardContainerRef} class="post-adaptive-cards" />
                `}
                ${cardSubmissionBlocks.length > 0 && X1`
                    <div class="post-adaptive-card-submissions">
                        ${cardSubmissionBlocks.map((block, idx) => {
    const meta = describeAdaptiveCardSubmission(block);
    const submissionKey = `${block.card_id}-${idx}`;
    return X1`
                                <div key=${submissionKey} class="adaptive-card-submission-receipt">
                                    <div class="adaptive-card-submission-header">
                                        <span class="adaptive-card-submission-icon" aria-hidden="true">✓</span>
                                        <div class="adaptive-card-submission-title-wrap">
                                            <span class="adaptive-card-submission-title">Submitted</span>
                                            <span class="adaptive-card-submission-title-action">${meta.title}</span>
                                        </div>
                                    </div>
                                    ${meta.fields.length > 0 && X1`
                                        <div class="adaptive-card-submission-fields">
                                            ${meta.fields.map((field) => X1`
                                                <span class="adaptive-card-submission-field" title=${`${field.key}: ${field.value}`}>
                                                    <span class="adaptive-card-submission-field-key">${field.key}</span>
                                                    <span class="adaptive-card-submission-field-sep">:</span>
                                                    <span class="adaptive-card-submission-field-value">${field.value}</span>
                                                </span>
                                            `)}
                                        </div>
                                    `}
                                    <div class="adaptive-card-submission-meta">
                                        Submitted ${formatTimestamp(meta.submittedAt)}
                                    </div>
                                </div>
                            `;
  })}
                    </div>
                `}
                ${generatedWidgets.length > 0 && X1`
                    <div class="generated-widget-launches">
                        ${generatedWidgets.map((block, idx) => X1`
                            <${GeneratedWidgetLaunch}
                                key=${block.widget_id || block.id || `${post.id}-widget-${idx}`}
                                block=${block}
                                post=${post}
                                onOpenWidget=${onOpenWidget}
                            />
                        `)}
                    </div>
                `}
                ${textAnnotations.length > 0 && X1`
                    ${textAnnotations.map((annotations, idx) => X1`
                        <${AnnotationsBadge} key=${idx} annotations=${annotations} />
                    `)}
                `}
                ${filteredImageItems.length > 0 && X1`
                    <div class="media-preview">
                        ${filteredImageItems.map(({ id, mimeType }) => {
    const isSvg = typeof mimeType === "string" && mimeType.toLowerCase().startsWith("image/svg");
    const imageSrc = isSvg ? getMediaUrl(id) : getThumbnailUrl(id);
    return X1`
                                <img 
                                    key=${id} 
                                    src=${imageSrc} 
                                    alt="Media" 
                                    loading="lazy"
                                    decoding="async"
                                    onClick=${(e2) => handleImageClick(e2, id)}
                                />
                            `;
  })}
                    </div>
                `}
                ${filteredImageItems.length > 0 && X1`
                    ${filteredImageItems.map(({ annotations }, idx) => X1`
                        ${annotations && X1`<${AnnotationsBadge} key=${idx} annotations=${annotations} />`}
                    `)}
                `}
                ${filteredFileIds.length > 0 && X1`
                    <div class="file-attachments">
                        ${filteredFileIds.map((id) => X1`
                            <${FileAttachment} key=${id} mediaId=${id} onPreview=${handleAttachmentPreview} />
                        `)}
                    </div>
                `}
                ${resourceLinks.length > 0 && X1`
                    <div class="resource-links">
                        ${resourceLinks.map((block, idx) => X1`
                            <div key=${idx}>
                                <${ResourceLinkBlock} block=${block} />
                                <${AnnotationsBadge} annotations=${block.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${resources.length > 0 && X1`
                    <div class="resource-embeds">
                        ${resources.map((block, idx) => X1`
                            <div key=${idx}>
                                <${ResourceBlock} block=${block} />
                                <${AnnotationsBadge} annotations=${block.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${data.link_previews?.length > 0 && X1`
                    <div class="link-previews">
                        ${data.link_previews.map((preview, i2) => X1`
                            <${LinkPreview} key=${i2} preview=${preview} />
                        `)}
                    </div>
                `}
            </div>
        </div>
        ${zoomedImage && X1`<${ImageModal} src=${zoomedImage} onClose=${() => setZoomedImage(null)} />`}

    `;
}

// web/src/components/timeline.ts
function Timeline({ posts, hasMore, onLoadMore, onPostClick, onHashtagClick, onMessageRef, onScrollToMessage, onFileRef, onOpenWidget, onOpenAttachmentPreview, emptyMessage, timelineRef, agents, user, onDeletePost, reverse = true, removingPostIds, searchQuery }) {
  const [loadingMore, setLoadingMore] = w0(false);
  const sentinelRef = o0(null);
  const hasIntersectionObserver = typeof IntersectionObserver !== "undefined";
  const triggerLoadMore = t0(async () => {
    if (!onLoadMore || !hasMore || loadingMore)
      return;
    setLoadingMore(true);
    try {
      await onLoadMore({ preserveScroll: true, preserveMode: "top" });
    } finally {
      setLoadingMore(false);
    }
  }, [hasMore, loadingMore, onLoadMore]);
  const handleScroll = t0((e2) => {
    const { scrollTop, scrollHeight, clientHeight } = e2.target;
    const distanceFromTop = reverse ? scrollHeight - clientHeight - scrollTop : scrollTop;
    const prefetchThreshold = Math.max(300, clientHeight);
    if (distanceFromTop < prefetchThreshold) {
      triggerLoadMore();
    }
  }, [reverse, triggerLoadMore]);
  r0(() => {
    if (!hasIntersectionObserver)
      return;
    const sentinel2 = sentinelRef.current;
    const root = timelineRef?.current;
    if (!sentinel2 || !root)
      return;
    const prefetchThreshold = 300;
    const observer = new IntersectionObserver((entries) => {
      for (const entry of entries) {
        if (!entry.isIntersecting)
          continue;
        triggerLoadMore();
      }
    }, {
      root,
      rootMargin: `${prefetchThreshold}px 0px ${prefetchThreshold}px 0px`,
      threshold: 0
    });
    observer.observe(sentinel2);
    return () => observer.disconnect();
  }, [hasIntersectionObserver, hasMore, onLoadMore, timelineRef, triggerLoadMore]);
  const triggerLoadMoreRef = o0(triggerLoadMore);
  triggerLoadMoreRef.current = triggerLoadMore;
  r0(() => {
    if (hasIntersectionObserver)
      return;
    if (!timelineRef?.current)
      return;
    const { scrollTop, scrollHeight, clientHeight } = timelineRef.current;
    const distanceFromTop = reverse ? scrollHeight - clientHeight - scrollTop : scrollTop;
    const prefetchThreshold = Math.max(300, clientHeight);
    if (distanceFromTop < prefetchThreshold) {
      triggerLoadMoreRef.current?.();
    }
  }, [hasIntersectionObserver, posts, hasMore, reverse, timelineRef]);
  r0(() => {
    if (!timelineRef?.current)
      return;
    if (!hasMore || loadingMore)
      return;
    const { scrollTop, scrollHeight, clientHeight } = timelineRef.current;
    const distanceFromTop = reverse ? scrollHeight - clientHeight - scrollTop : scrollTop;
    const prefetchThreshold = Math.max(300, clientHeight);
    if (scrollHeight <= clientHeight + 1 || distanceFromTop < prefetchThreshold) {
      triggerLoadMoreRef.current?.();
    }
  }, [posts, hasMore, loadingMore, reverse, timelineRef]);
  if (!posts) {
    return X1`<div class="loading"><div class="spinner"></div></div>`;
  }
  if (posts.length === 0) {
    return X1`
            <div class="timeline" ref=${timelineRef}>
                <div class="timeline-content">
                    <div style="padding: var(--spacing-xl); text-align: center; color: var(--text-secondary)">
                        ${emptyMessage || "No messages yet. Start a conversation!"}
                    </div>
                </div>
            </div>
        `;
  }
  const displayPosts = posts.slice().sort((a2, b) => a2.id - b.id);
  const resolveThreadRootId = (post) => {
    const raw = post?.data?.thread_id;
    if (raw === null || raw === undefined || raw === "")
      return null;
    const threadId = Number(raw);
    return Number.isFinite(threadId) ? threadId : null;
  };
  const threadGroups = new Map;
  for (let i2 = 0;i2 < displayPosts.length; i2 += 1) {
    const post = displayPosts[i2];
    const postId = Number(post?.id);
    const threadRootId = resolveThreadRootId(post);
    if (threadRootId !== null) {
      const group = threadGroups.get(threadRootId) || { anchorIndex: -1, replyIndexes: [] };
      group.replyIndexes.push(i2);
      threadGroups.set(threadRootId, group);
    } else if (Number.isFinite(postId)) {
      const group = threadGroups.get(postId) || { anchorIndex: -1, replyIndexes: [] };
      group.anchorIndex = i2;
      threadGroups.set(postId, group);
    }
  }
  const threadSequences = new Map;
  for (const [threadId, group] of threadGroups.entries()) {
    const ordered = new Set;
    if (group.anchorIndex >= 0)
      ordered.add(group.anchorIndex);
    for (const index of group.replyIndexes)
      ordered.add(index);
    threadSequences.set(threadId, Array.from(ordered).sort((a2, b) => a2 - b));
  }
  const threadInfoByIndex = displayPosts.map((post, index) => {
    const threadRootId = resolveThreadRootId(post);
    if (threadRootId === null)
      return { hasThreadPrev: false, hasThreadNext: false };
    const sequence = threadSequences.get(threadRootId);
    if (!sequence || sequence.length === 0)
      return { hasThreadPrev: false, hasThreadNext: false };
    const position = sequence.indexOf(index);
    if (position < 0)
      return { hasThreadPrev: false, hasThreadNext: false };
    return {
      hasThreadPrev: position > 0,
      hasThreadNext: position < sequence.length - 1
    };
  });
  const sentinel = X1`<div class="timeline-sentinel" ref=${sentinelRef}></div>`;
  return X1`
        <div class="timeline ${reverse ? "reverse" : "normal"}" ref=${timelineRef} onScroll=${handleScroll}>
            <div class="timeline-content">
                ${reverse ? sentinel : null}
                ${displayPosts.map((post, index) => {
    const isThreadReply = Boolean(post.data?.thread_id && post.data.thread_id !== post.id);
    const isRemoving = removingPostIds?.has?.(post.id);
    const threadInfo = threadInfoByIndex[index] || {};
    return X1`
                    <${Post}
                        key=${post.id}
                        post=${post}
                        isThreadReply=${isThreadReply}
                        isThreadPrev=${threadInfo.hasThreadPrev}
                        isThreadNext=${threadInfo.hasThreadNext}
                        isRemoving=${isRemoving}
                        highlightQuery=${searchQuery}
                        agentName=${getAgentName(post.data?.agent_id, agents || {})}
                        agentAvatarUrl=${getAgentAvatarUrl(post.data?.agent_id, agents || {})}
                        userName=${user?.name || user?.user_name}
                        userAvatarUrl=${user?.avatar_url || user?.avatarUrl || user?.avatar}
                        userAvatarBackground=${user?.avatar_background || user?.avatarBackground}
                        onClick=${() => onPostClick?.(post)}
                        onHashtagClick=${onHashtagClick}
                        onMessageRef=${onMessageRef}
                        onScrollToMessage=${onScrollToMessage}
                        onFileRef=${onFileRef}
                        onOpenWidget=${onOpenWidget}
                        onDelete=${onDeletePost}
                        onOpenAttachmentPreview=${onOpenAttachmentPreview}
                    />
                `;
  })}
                ${reverse ? null : sentinel}
            </div>
        </div>
    `;
}

// web/src/ui/popup-typeahead.ts
var POPUP_TYPEAHEAD_RESET_MS = 700;
function normalize2(value) {
  return String(value || "").toLowerCase().replace(/^@/, "").replace(/\s+/g, " ").trim();
}
function labelMatchesQuery(label, query) {
  const normalizedLabel = normalize2(label);
  const normalizedQuery = normalize2(query);
  if (!normalizedQuery)
    return false;
  return normalizedLabel.startsWith(normalizedQuery) || normalizedLabel.includes(normalizedQuery);
}
function isPopupTypeaheadKey(event) {
  if (!event)
    return false;
  if (event.isComposing)
    return false;
  if (event.ctrlKey || event.metaKey || event.altKey)
    return false;
  return typeof event.key === "string" && event.key.length === 1 && /\S/.test(event.key);
}
function updatePopupTypeaheadBuffer(previous, key, now = Date.now(), resetMs = POPUP_TYPEAHEAD_RESET_MS) {
  const prior = previous && typeof previous === "object" ? previous : { value: "", updatedAt: 0 };
  const char = String(key || "").trim().toLowerCase();
  if (!char)
    return { value: "", updatedAt: now };
  const shouldReset = !prior.value || !Number.isFinite(prior.updatedAt) || now - prior.updatedAt > resetMs;
  return {
    value: shouldReset ? char : `${prior.value}${char}`,
    updatedAt: now
  };
}
function rotatedIndices(length, startIndex) {
  const size = Math.max(0, Number(length) || 0);
  if (size <= 0)
    return [];
  const start = Number.isInteger(startIndex) ? startIndex : 0;
  const normalizedStart = (start % size + size) % size;
  const out = [];
  for (let i2 = 0;i2 < size; i2 += 1) {
    out.push((normalizedStart + i2) % size);
  }
  return out;
}
function findPopupTypeaheadMatch(items, query, startIndex = 0, getLabel = (item) => item) {
  const normalizedQuery = normalize2(query);
  if (!normalizedQuery)
    return -1;
  const list = Array.isArray(items) ? items : [];
  const indices = rotatedIndices(list.length, startIndex);
  const labels = list.map((item) => normalize2(getLabel(item)));
  for (const idx of indices) {
    if (labels[idx].startsWith(normalizedQuery))
      return idx;
  }
  for (const idx of indices) {
    if (labels[idx].includes(normalizedQuery))
      return idx;
  }
  return -1;
}
function resolvePopupTypeaheadMatch(items, query, currentIndex = -1, getLabel = (item) => item) {
  const list = Array.isArray(items) ? items : [];
  if (currentIndex >= 0 && currentIndex < list.length) {
    const currentLabel = getLabel(list[currentIndex]);
    if (labelMatchesQuery(currentLabel, query)) {
      return currentIndex;
    }
  }
  return findPopupTypeaheadMatch(list, query, 0, getLabel);
}

// web/src/ui/agent-mentions.ts
function normalizeAgentName(value) {
  return String(value || "").trim().toLowerCase();
}
function parseMentionAutocompleteQuery(value) {
  const match = String(value || "").match(/^@([a-zA-Z0-9_-]*)$/);
  if (!match)
    return null;
  return normalizeAgentName(match[1] || "");
}
function dedupeAgents(agents) {
  const seen = new Set;
  const result = [];
  for (const agent of Array.isArray(agents) ? agents : []) {
    const handle = normalizeAgentName(agent?.agent_name);
    if (!handle || seen.has(handle))
      continue;
    seen.add(handle);
    result.push(agent);
  }
  return result;
}
function filterMentionAgents(agents, value, options = {}) {
  const prefix = parseMentionAutocompleteQuery(value);
  if (prefix == null)
    return [];
  const currentChatJid = typeof options?.currentChatJid === "string" ? options.currentChatJid : null;
  return dedupeAgents(agents).filter((agent) => {
    if (currentChatJid && agent?.chat_jid === currentChatJid)
      return false;
    const handle = normalizeAgentName(agent?.agent_name);
    return handle.startsWith(prefix);
  });
}
function buildMentionValue(agentName) {
  const handle = normalizeAgentName(agentName);
  return handle ? `@${handle} ` : "";
}

// web/src/ui/compose-session-switcher.ts
function shouldOpenSessionSwitcherFromBlankCompose(event, value, options = {}) {
  if (!event || event.isComposing)
    return false;
  if (options.searchMode)
    return false;
  if (!options.showSessionSwitcherButton)
    return false;
  if (event.ctrlKey || event.metaKey || event.altKey)
    return false;
  if (event.key !== "@")
    return false;
  return String(value || "") === "";
}

// web/src/ui/branch-lifecycle.ts
function normalizeHandle(value) {
  const normalized = normalizeHandleName(value);
  return normalized ? `@${normalized}` : "";
}
function normalizeHandleName(value) {
  return String(value || "").trim().toLowerCase().replace(/[^a-z0-9_-]+/g, "-").replace(/^-+|-+$/g, "").replace(/-{2,}/g, "-");
}
function formatCurrentBranchLabel(currentSessionAgent, currentChatJid) {
  const currentHandle = typeof currentSessionAgent?.agent_name === "string" && currentSessionAgent.agent_name.trim() ? normalizeHandle(currentSessionAgent.agent_name) : String(currentChatJid || "").trim();
  const currentId = typeof currentSessionAgent?.chat_jid === "string" && currentSessionAgent.chat_jid.trim() ? currentSessionAgent.chat_jid.trim() : String(currentChatJid || "").trim();
  return `${currentHandle} — ${currentId} • current branch`;
}
function getBranchLifecycleBadges(chat, options = {}) {
  const badges = [];
  const currentChatJid = typeof options.currentChatJid === "string" ? options.currentChatJid.trim() : "";
  const chatJid = typeof chat?.chat_jid === "string" ? chat.chat_jid.trim() : "";
  if (currentChatJid && chatJid === currentChatJid) {
    badges.push("current");
  }
  if (chat?.archived_at) {
    badges.push("archived");
  } else if (chat?.is_active) {
    badges.push("active");
  }
  return badges;
}
function formatBranchPickerLabel(chat, options = {}) {
  const handle = normalizeHandle(chat?.agent_name) || String(chat?.chat_jid || "").trim();
  const chatJid = typeof chat?.chat_jid === "string" && chat.chat_jid.trim() ? chat.chat_jid.trim() : "unknown-chat";
  const badges = getBranchLifecycleBadges(chat, options);
  return badges.length > 0 ? `${handle} — ${chatJid} • ${badges.join(" • ")}` : `${handle} — ${chatJid}`;
}

// web/src/ui/status-dot.js
function buildTurnDotClass({ steerQueued = false, pulsing = false } = {}) {
  const classes = ["turn-dot"];
  if (steerQueued)
    classes.push("turn-dot-queued");
  if (pulsing)
    classes.push("turn-dot-pulsing");
  return classes.join(" ");
}
function buildComposeStatusDotClass({ pulsing = false } = {}) {
  const classes = ["compose-inline-status-dot"];
  if (pulsing)
    classes.push("compose-inline-status-dot-pulsing");
  return classes.join(" ");
}
function resolveRunningStatusIndicator(status, { isLastActivity = false, pendingRequest = false } = {}) {
  if (pendingRequest)
    return "dot";
  if (isLastActivity)
    return "none";
  if (status?.type === "error")
    return "none";
  if (status?.type === "intent")
    return "dot";
  const type = typeof status?.type === "string" ? status.type : "";
  const hasToolMetadata = Boolean(typeof status?.tool_name === "string" && status.tool_name.trim() || status?.tool_args);
  if (hasToolMetadata)
    return "spinner";
  if (type === "tool_call" || type === "tool_status" || type === "thinking" || type === "waiting") {
    return "spinner";
  }
  return "dot";
}
function shouldShowRunningStatusDot(status, options = {}) {
  return resolveRunningStatusIndicator(status, options) === "dot";
}

// web/src/ui/connection-status.ts
var RECONNECTING_HINT_DELAY_MS = 350;
function formatConnectionStatusLabel(status) {
  return String(status || "Connecting").replace(/[-_]+/g, " ").replace(/^./, (match) => match.toUpperCase());
}
function resolveConnectionStatusPresentation(status, options = {}) {
  const normalizedStatus = typeof status === "string" && status.trim() ? status.trim() : "connecting";
  if (normalizedStatus === "connected") {
    return {
      show: false,
      statusClass: "connected",
      label: "Connected",
      title: "Connection: Connected"
    };
  }
  if (normalizedStatus !== "disconnected") {
    const label = formatConnectionStatusLabel(normalizedStatus);
    return {
      show: true,
      statusClass: normalizedStatus,
      label,
      title: `Connection: ${label}`
    };
  }
  const delayMs = Number.isFinite(Number(options?.delayMs)) ? Math.max(0, Number(options.delayMs)) : RECONNECTING_HINT_DELAY_MS;
  const nowMs = Number.isFinite(Number(options?.nowMs)) ? Number(options.nowMs) : Date.now();
  const disconnectedAtMs = Number.isFinite(Number(options?.disconnectedAtMs)) ? Number(options.disconnectedAtMs) : nowMs;
  const showReconnecting = nowMs - disconnectedAtMs >= delayMs;
  return showReconnecting ? {
    show: true,
    statusClass: "disconnected",
    label: "Reconnecting",
    title: "Reconnecting"
  } : {
    show: true,
    statusClass: "connecting",
    label: "Connecting",
    title: "Connecting"
  };
}
function useConnectionStatusPresentation(status, options = {}) {
  const delayMs = Number.isFinite(Number(options?.delayMs)) ? Math.max(0, Number(options.delayMs)) : RECONNECTING_HINT_DELAY_MS;
  const [disconnectedAtMs, setDisconnectedAtMs] = w0(null);
  const [displayNowMs, setDisplayNowMs] = w0(() => Date.now());
  r0(() => {
    if (status === "disconnected") {
      const startedAt = Date.now();
      setDisconnectedAtMs((previous) => previous ?? startedAt);
      setDisplayNowMs(startedAt);
      return;
    }
    setDisconnectedAtMs(null);
    setDisplayNowMs(Date.now());
  }, [status]);
  r0(() => {
    if (status !== "disconnected" || disconnectedAtMs === null)
      return;
    const remainingMs = delayMs - (Date.now() - disconnectedAtMs);
    if (remainingMs <= 0)
      return;
    const timeoutId = setTimeout(() => {
      setDisplayNowMs(Date.now());
    }, remainingMs);
    return () => clearTimeout(timeoutId);
  }, [status, disconnectedAtMs, delayMs]);
  return G0(() => resolveConnectionStatusPresentation(status, {
    delayMs,
    disconnectedAtMs,
    nowMs: displayNowMs
  }), [status, delayMs, disconnectedAtMs, displayNowMs]);
}

// web/src/components/compose-model-refresh.ts
async function refreshAgentModelStateBestEffort(getAgentModels2, chatJid, emitModelState) {
  if (typeof getAgentModels2 !== "function")
    return false;
  try {
    const latest = await getAgentModels2(chatJid);
    if (!latest)
      return false;
    emitModelState(latest);
    return true;
  } catch (_error) {
    return false;
  }
}

// web/src/components/compose-box.ts
var SLASH_COMMANDS = [
  { name: "/model", description: "Select model or list available models" },
  { name: "/cycle-model", description: "Cycle to the next available model" },
  { name: "/thinking", description: "Show or set thinking/effort level" },
  { name: "/effort", description: "Show or set thinking/effort level (alias for /thinking)" },
  { name: "/cycle-thinking", description: "Cycle thinking level" },
  { name: "/theme", description: "Set UI theme (no name to show available themes)" },
  { name: "/meters", description: "Toggle the top-right CPU/RAM HUD (/meters on|off|toggle)" },
  { name: "/tint", description: "Tint default light/dark UI (usage: /tint #hex or /tint off)" },
  { name: "/btw", description: "Open a side conversation panel without interrupting the main chat" },
  { name: "/state", description: "Show current session state" },
  { name: "/stats", description: "Show session token and cost stats" },
  { name: "/context", description: "Show context window usage" },
  { name: "/last", description: "Show last assistant response" },
  { name: "/compact", description: "Manually compact the session" },
  { name: "/auto-compact", description: "Toggle auto-compaction" },
  { name: "/auto-retry", description: "Toggle auto-retry" },
  { name: "/abort", description: "Abort the current response" },
  { name: "/abort-retry", description: "Abort retry backoff" },
  { name: "/abort-bash", description: "Abort running bash command" },
  { name: "/shell", description: "Run a shell command and return output" },
  { name: "/bash", description: "Run a shell command and add output to context" },
  { name: "/queue", description: "Queue a follow-up message (one-at-a-time)" },
  { name: "/queue-all", description: "Queue a follow-up message (batch all)" },
  { name: "/steer", description: "Steer the current response" },
  { name: "/steering-mode", description: "Set steering mode (all|one)" },
  { name: "/followup-mode", description: "Set follow-up mode (all|one)" },
  { name: "/session-name", description: "Set or show the session name" },
  { name: "/new-session", description: "Start a new session" },
  { name: "/switch-session", description: "Switch to a session file" },
  { name: "/session-rotate", description: "Rotate the current persisted session into an archived file" },
  { name: "/clone", description: "Duplicate the current active branch into a new session" },
  { name: "/fork", description: "Fork from a previous message" },
  { name: "/forks", description: "List forkable messages" },
  { name: "/tree", description: "List the session tree" },
  { name: "/label", description: "Set or clear a label on a tree entry" },
  { name: "/labels", description: "List labeled entries" },
  { name: "/agent-name", description: "Set or show the agent display name" },
  { name: "/agent-avatar", description: "Set or show the agent avatar URL" },
  { name: "/user-name", description: "Set or show your display name" },
  { name: "/user-avatar", description: "Set or show your avatar URL" },
  { name: "/user-github", description: "Set name/avatar from GitHub profile" },
  { name: "/export-html", description: "Export session to HTML" },
  { name: "/passkey", description: "Manage passkeys (enrol/list/delete)" },
  { name: "/totp", description: "Show a TOTP enrolment QR code" },
  { name: "/qr", description: "Generate a QR code for text or URL" },
  { name: "/search", description: "Search notes and skills in the workspace" },
  { name: "/dream", description: "Run Dream memory maintenance over recent days (default 7)" },
  { name: "/tasks", description: "List scheduled tasks" },
  { name: "/scheduled", description: "List scheduled tasks" },
  { name: "/restart", description: "Restart the agent and stop subprocesses" },
  { name: "/exit", description: "Exit the current piclaw process immediately (Supervisor will restart it)" },
  { name: "/login", description: "Login to an AI model provider (OAuth or API key)" },
  { name: "/logout", description: "Logout from an AI model provider" },
  { name: "/commands", description: "List available commands" },
  { name: "/skill:", description: "Run a workspace skill (e.g. /skill:visual-artifact-generator, /skill:web-search)" }
];
var COMPOSE_HISTORY_STORAGE_KEY = "piclaw_compose_history";
function resolveComposePrefillRequest(prefillRequest, lastHandledToken, searchMode = false) {
  if (searchMode)
    return { shouldApply: false, nextToken: lastHandledToken, text: "" };
  if (!prefillRequest || typeof prefillRequest !== "object") {
    return { shouldApply: false, nextToken: lastHandledToken, text: "" };
  }
  const token = typeof prefillRequest.token === "string" ? prefillRequest.token : "";
  const text = typeof prefillRequest.text === "string" ? prefillRequest.text : "";
  if (!token || token === lastHandledToken || !text.trim()) {
    return { shouldApply: false, nextToken: lastHandledToken, text: "" };
  }
  return { shouldApply: true, nextToken: token, text };
}
function getComposeHistoryStorageKey(chatJid = "web:default") {
  const normalized = typeof chatJid === "string" && chatJid.trim() ? chatJid.trim() : "web:default";
  if (normalized === "web:default")
    return COMPOSE_HISTORY_STORAGE_KEY;
  return `${COMPOSE_HISTORY_STORAGE_KEY}:${encodeURIComponent(normalized)}`;
}
function resolveUiOnlyCommandNotice(commandText, response) {
  const message = typeof response?.command?.message === "string" ? response.command.message.trim() : "";
  if (!response?.ui_only || !message)
    return null;
  const trimmed = typeof commandText === "string" ? commandText.trim() : "";
  if (!trimmed.startsWith("/"))
    return null;
  const parts = trimmed.split(/\s+/).filter(Boolean);
  const slashName = parts[0]?.toLowerCase() || "";
  const hasArgs = parts.length > 1;
  if (!hasArgs && (slashName === "/model" || slashName === "/thinking" || slashName === "/effort")) {
    return message;
  }
  return null;
}
function resolveComposeSubmitButtonState(isAgentActive, canSend, isCompacting = false) {
  if (isAgentActive && isCompacting) {
    return {
      mode: "compacting",
      className: "icon-btn send-btn abort-mode compacting-mode",
      title: "Compacting context — Stop response",
      ariaLabel: "Compacting context — Stop response",
      disabled: false
    };
  }
  if (isAgentActive) {
    return {
      mode: "abort",
      className: "icon-btn send-btn abort-mode",
      title: "Stop response",
      ariaLabel: "Stop response",
      disabled: false
    };
  }
  return {
    mode: "send",
    className: "icon-btn send-btn",
    title: "Send (Enter)",
    ariaLabel: "Send message",
    disabled: !canSend
  };
}
function isComposeSubmitAbortMode(mode) {
  return mode === "abort" || mode === "compacting";
}
function resolveComposeExtensionWorkingDisplay(workingState, frameIndex = 0) {
  const message = typeof workingState?.message === "string" && workingState.message.trim() ? workingState.message.trim() : null;
  const indicator = workingState?.indicator && typeof workingState.indicator === "object" ? workingState.indicator : null;
  if (!message && !indicator) {
    return {
      visible: false,
      title: "",
      indicatorText: null,
      animateDot: false
    };
  }
  if (indicator?.mode === "hidden") {
    return {
      visible: Boolean(message),
      title: message || "Working…",
      indicatorText: null,
      animateDot: false
    };
  }
  if (indicator?.mode === "custom" && Array.isArray(indicator.frames) && indicator.frames.length > 0) {
    const frames = indicator.frames;
    const safeIndex = Number.isFinite(frameIndex) && frameIndex >= 0 ? Math.floor(frameIndex) % frames.length : 0;
    return {
      visible: true,
      title: message || "Working…",
      indicatorText: frames[safeIndex],
      animateDot: false
    };
  }
  return {
    visible: true,
    title: message || "Working…",
    indicatorText: null,
    animateDot: true
  };
}
function ContextPie({ usage, onCompact }) {
  const pct = Math.min(100, Math.max(0, usage.percent || 0));
  const tokens = usage.tokens;
  const window2 = usage.contextWindow;
  const compactLabel = `Compact context`;
  const label = tokens != null ? `Context: ${formatK(tokens)} / ${formatK(window2)} tokens (${pct.toFixed(0)}%)` : `Context: ${pct.toFixed(0)}%`;
  const title = `${label} — ${compactLabel}`;
  const r2 = 9;
  const circ = 2 * Math.PI * r2;
  const filled = pct / 100 * circ;
  const color = pct > 90 ? "var(--context-red, #ef4444)" : pct > 75 ? "var(--context-amber, #f59e0b)" : "var(--context-green, #22c55e)";
  return X1`
        <button
            class="compose-context-pie icon-btn"
            type="button"
            title=${title}
            aria-label="Compact context"
            onClick=${(e2) => {
    e2.preventDefault();
    e2.stopPropagation();
    onCompact?.();
  }}
        >
            <svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="12" cy="12" r=${r2}
                    fill="none"
                    stroke="var(--context-track, rgba(128,128,128,0.2))"
                    stroke-width="2.5" />
                <circle cx="12" cy="12" r=${r2}
                    fill="none"
                    stroke=${color}
                    stroke-width="2.5"
                    stroke-dasharray=${`${filled} ${circ}`}
                    stroke-linecap="round"
                    transform="rotate(-90 12 12)" />
            </svg>
        </button>
    `;
}
function formatK(n2) {
  if (n2 == null)
    return "?";
  if (n2 >= 1e6)
    return (n2 / 1e6).toFixed(1) + "M";
  if (n2 >= 1000)
    return (n2 / 1000).toFixed(0) + "K";
  return String(n2);
}
function formatModelPickerContextWindow(contextWindow) {
  const value = Number(contextWindow);
  if (!Number.isFinite(value) || value <= 0)
    return "";
  return `${formatK(value)} ctx`;
}
function formatModelPickerDisplayLabel(label, contextWindow) {
  const primaryLabel = typeof label === "string" ? label.trim() : "";
  const contextLabel = formatModelPickerContextWindow(contextWindow);
  if (!primaryLabel)
    return contextLabel;
  if (!contextLabel)
    return primaryLabel;
  return `${primaryLabel} • ${contextLabel}`;
}
function normalizeModelPickerLabel(value, provider = "", id = "") {
  const explicit = typeof value === "string" ? value.trim() : "";
  if (explicit)
    return explicit;
  const normalizedProvider = typeof provider === "string" ? provider.trim() : "";
  const normalizedId = typeof id === "string" ? id.trim() : "";
  if (normalizedProvider && normalizedId)
    return `${normalizedProvider}/${normalizedId}`;
  return normalizedId || normalizedProvider || "";
}
function normalizeModelPickerOptions(payload) {
  const structured = Array.isArray(payload?.model_options) ? payload.model_options : null;
  const legacy = Array.isArray(payload?.models) ? payload.models : [];
  const rawItems = structured && structured.length > 0 ? structured : legacy;
  const options = [];
  for (const item of rawItems) {
    if (typeof item === "string") {
      const label2 = item.trim();
      if (!label2)
        continue;
      const slashIndex = label2.indexOf("/");
      const provider2 = slashIndex > 0 ? label2.slice(0, slashIndex).trim() : "";
      const id2 = slashIndex > 0 ? label2.slice(slashIndex + 1).trim() : label2;
      options.push({
        label: label2,
        provider: provider2,
        id: id2,
        name: null,
        contextWindow: null,
        reasoning: false
      });
      continue;
    }
    if (!item || typeof item !== "object")
      continue;
    const provider = typeof item.provider === "string" ? item.provider.trim() : "";
    const id = typeof item.id === "string" ? item.id.trim() : "";
    const label = normalizeModelPickerLabel(item.label, provider, id);
    if (!label)
      continue;
    const name = typeof item.name === "string" && item.name.trim() ? item.name.trim() : null;
    const contextWindow = Number(item.context_window ?? item.contextWindow);
    options.push({
      label,
      provider,
      id,
      name,
      contextWindow: Number.isFinite(contextWindow) && contextWindow > 0 ? contextWindow : null,
      reasoning: Boolean(item.reasoning)
    });
  }
  options.sort((a2, b) => a2.label.localeCompare(b.label, undefined, { sensitivity: "base" }));
  return options;
}
function getModelPickerOptionSearchLabel(option) {
  if (!option || typeof option !== "object")
    return "";
  return [
    option.label,
    option.provider,
    option.id,
    option.name,
    formatModelPickerContextWindow(option.contextWindow)
  ].filter(Boolean).join(" ");
}
function resolveComposeModelPickerState(activeModel, agentModelsPayload) {
  const modelLabel = typeof activeModel === "string" ? activeModel.trim() : "";
  if (modelLabel) {
    return {
      showPicker: true,
      label: modelLabel,
      hasAvailableModels: true
    };
  }
  const hasAvailableModels = normalizeModelPickerOptions(agentModelsPayload).length > 0;
  return {
    showPicker: hasAvailableModels,
    label: hasAvailableModels ? "Select model" : "",
    hasAvailableModels
  };
}
function unwrapQueuedTranscriptContent(value) {
  if (!value)
    return value;
  const normalized = value.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  if (!normalized.includes(" @ ") || !normalized.includes(`:
`))
    return value;
  const lines = normalized.split(`
`);
  const collected = [];
  let index = 0;
  let sawTranscript = false;
  while (index < lines.length) {
    const line = lines[index];
    const trimmed = line.trim();
    if (!trimmed) {
      index += 1;
      continue;
    }
    if (trimmed === "Messages:" || trimmed.startsWith("Channel:")) {
      sawTranscript = true;
      index += 1;
      continue;
    }
    if (/^[^\n]+\s@\s[^\n]+:$/.test(trimmed)) {
      sawTranscript = true;
      index += 1;
      const bodyLines = [];
      while (index < lines.length) {
        const current = lines[index];
        const currentTrimmed = current.trim();
        if (/^[^\n]+\s@\s[^\n]+:$/.test(currentTrimmed))
          break;
        if (currentTrimmed.startsWith("Channel:") || currentTrimmed === "Messages:")
          break;
        bodyLines.push(current.startsWith("  ") ? current.slice(2) : current);
        index += 1;
      }
      if (bodyLines.length > 0) {
        collected.push(bodyLines.join(`
`).trim());
      }
      continue;
    }
    return value;
  }
  return sawTranscript && collected.length > 0 ? collected.filter(Boolean).join(`

`) : value;
}
function extractQueuedFileRefs(value) {
  if (!value)
    return { content: value, fileRefs: [] };
  const normalized = value.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  let start = -1;
  for (let i2 = 0;i2 < lines.length; i2 += 1) {
    if (lines[i2].trim() === "Files:" && lines[i2 + 1] && /^\s*-\s+/.test(lines[i2 + 1])) {
      start = i2;
      break;
    }
  }
  if (start === -1)
    return { content: value, fileRefs: [] };
  const refs = [];
  let end = start + 1;
  for (;end < lines.length; end += 1) {
    const line = lines[end];
    if (/^\s*-\s+/.test(line)) {
      refs.push(line.replace(/^\s*-\s+/, "").trim());
    } else if (!line.trim()) {
      break;
    } else {
      break;
    }
  }
  if (refs.length === 0)
    return { content: value, fileRefs: [] };
  const before = lines.slice(0, start);
  const after = lines.slice(end);
  const cleaned = [...before, ...after].join(`
`).replace(/\n{3,}/g, `

`).trim();
  return { content: cleaned, fileRefs: refs };
}
function extractQueuedMessageRefs(value) {
  if (!value)
    return { content: value, messageRefs: [] };
  const normalized = value.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  let start = -1;
  for (let i2 = 0;i2 < lines.length; i2 += 1) {
    if (lines[i2].trim() === "Referenced messages:" && lines[i2 + 1] && /^\s*-\s+/.test(lines[i2 + 1])) {
      start = i2;
      break;
    }
  }
  if (start === -1)
    return { content: value, messageRefs: [] };
  const refs = [];
  let end = start + 1;
  for (;end < lines.length; end += 1) {
    const line = lines[end];
    if (/^\s*-\s+/.test(line)) {
      const match = line.replace(/^\s*-\s+/, "").trim().match(/^message:(\S+)$/i);
      if (match)
        refs.push(match[1]);
    } else if (!line.trim()) {
      break;
    } else {
      break;
    }
  }
  if (refs.length === 0)
    return { content: value, messageRefs: [] };
  const before = lines.slice(0, start);
  const after = lines.slice(end);
  const cleaned = [...before, ...after].join(`
`).replace(/\n{3,}/g, `

`).trim();
  return { content: cleaned, messageRefs: refs };
}
function extractQueuedAttachmentRefs(value) {
  if (!value)
    return { content: value, attachmentRefs: [] };
  const normalized = value.replace(/\r\n/g, `
`).replace(/\r/g, `
`);
  const lines = normalized.split(`
`);
  let start = -1;
  for (let i2 = 0;i2 < lines.length; i2 += 1) {
    if (lines[i2].trim() === "Attachments:" && lines[i2 + 1] && /^\s*-\s+/.test(lines[i2 + 1])) {
      start = i2;
      break;
    }
  }
  if (start === -1)
    return { content: value, attachmentRefs: [] };
  const refs = [];
  let end = start + 1;
  for (;end < lines.length; end += 1) {
    const line = lines[end];
    if (/^\s*-\s+/.test(line)) {
      const item = line.replace(/^\s*-\s+/, "").trim();
      const match = item.match(/^attachment:(\d+)(?:\s*\((.+)\))?$/i);
      if (match) {
        refs.push({
          id: match[1],
          label: (match[2] || "").trim() || `attachment:${match[1]}`,
          raw: item
        });
      }
    } else if (!line.trim()) {
      break;
    } else {
      break;
    }
  }
  if (refs.length === 0)
    return { content: value, attachmentRefs: [] };
  const before = lines.slice(0, start);
  const after = lines.slice(end);
  const cleaned = [...before, ...after].join(`
`).replace(/\n{3,}/g, `

`).trim();
  return { content: cleaned, attachmentRefs: refs };
}
function parseQueuedContent(value) {
  const unwrapped = unwrapQueuedTranscriptContent(value || "");
  const withFiles = extractQueuedFileRefs(unwrapped || "");
  const withMessages = extractQueuedMessageRefs(withFiles.content || "");
  const withAttachments = extractQueuedAttachmentRefs(withMessages.content || "");
  return {
    text: withAttachments.content || "",
    fileRefs: withFiles.fileRefs,
    messageRefs: withMessages.messageRefs,
    attachmentRefs: withAttachments.attachmentRefs
  };
}
function QueuedFollowupStack({
  items = [],
  onInjectQueuedFollowup,
  onRemoveQueuedFollowup,
  onMoveQueuedFollowup,
  onOpenFilePill
}) {
  if (!Array.isArray(items) || items.length === 0)
    return null;
  return X1`
        <div class="compose-queue-stack">
            ${items.map((item, index) => {
    const rowText = typeof item?.content === "string" ? item.content : "";
    const parsed = parseQueuedContent(rowText);
    if (!parsed.text.trim() && parsed.fileRefs.length === 0 && parsed.messageRefs.length === 0 && parsed.attachmentRefs.length === 0)
      return null;
    const canMoveUp = index > 0;
    const canMoveDown = index < items.length - 1;
    return X1`
                    <div class="compose-queue-stack-item" role="listitem">
                        <div class="compose-queue-stack-content" title=${rowText}>
                            ${parsed.text.trim() && X1`<div class="compose-queue-stack-text">${parsed.text}</div>`}
                            ${(parsed.messageRefs.length > 0 || parsed.fileRefs.length > 0 || parsed.attachmentRefs.length > 0) && X1`
                                <div class="compose-queue-stack-refs">
                                    ${parsed.messageRefs.map((id) => X1`
                                        <${FilePill}
                                            key=${"queue-msg-" + id}
                                            prefix="compose"
                                            label=${"msg:" + id}
                                            title=${"Message reference: " + id}
                                            icon="message"
                                        />
                                    `)}
                                    ${parsed.fileRefs.map((path) => {
      const label = path.split("/").pop() || path;
      return X1`
                                            <${FilePill}
                                                key=${"queue-file-" + path}
                                                prefix="compose"
                                                label=${label}
                                                title=${path}
                                                onClick=${() => onOpenFilePill?.(path)}
                                            />
                                        `;
    })}
                                    ${parsed.attachmentRefs.map((attachment) => X1`
                                        <${FilePill}
                                            key=${"queue-attachment-" + attachment.id}
                                            prefix="compose"
                                            label=${attachment.label}
                                            title=${attachment.raw}
                                        />
                                    `)}
                                </div>
                            `}
                        </div>
                        <div class="compose-queue-stack-actions" role="group" aria-label="Queued follow-up controls">
                            ${items.length > 1 && X1`
                                <button
                                    class="compose-queue-stack-move-btn"
                                    type="button"
                                    title="Move up"
                                    aria-label="Move up in queue"
                                    disabled=${!canMoveUp}
                                    onClick=${() => canMoveUp && onMoveQueuedFollowup?.(index, index - 1)}
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
                                    disabled=${!canMoveDown}
                                    onClick=${() => canMoveDown && onMoveQueuedFollowup?.(index, index + 1)}
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
                                onClick=${() => onInjectQueuedFollowup?.(item)}
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
                                onClick=${() => onRemoveQueuedFollowup?.(item)}
                            >
                                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                    <line x1="18" y1="6" x2="6" y2="18" />
                                    <line x1="6" y1="6" x2="18" y2="18" />
                                </svg>
                            </button>
                        </div>
                    </div>
                `;
  })}
        </div>
    `;
}
function ComposeBox({
  onPost,
  onFocus,
  searchMode,
  searchScope = "current",
  onSearch,
  onSearchScopeChange,
  onEnterSearch,
  onExitSearch,
  fileRefs = [],
  onRemoveFileRef,
  onClearFileRefs,
  messageRefs = [],
  onRemoveMessageRef,
  onClearMessageRefs,
  activeModel = null,
  agentModelsPayload = null,
  modelUsage = null,
  thinkingLevel = null,
  supportsThinking = false,
  contextUsage = null,
  onContextCompact,
  notificationsEnabled = false,
  notificationPermission = "default",
  onToggleNotifications,
  onModelChange,
  onModelStateChange,
  activeEditorPath = null,
  onAttachEditorFile,
  onOpenFilePill,
  followupQueueItems = [],
  onInjectQueuedFollowup,
  onRemoveQueuedFollowup,
  onMoveQueuedFollowup,
  onSubmitIntercept,
  onMessageResponse,
  onPopOutChat,
  isAgentActive = false,
  activeChatAgents = [],
  currentChatJid = "web:default",
  connectionStatus = "connected",
  onSetFileRefs,
  onSetMessageRefs,
  onSubmitError,
  onSwitchChat,
  onRenameSession,
  isRenameSessionInProgress = false,
  onCreateSession,
  onDeleteSession,
  onRestoreSession,
  showQueueStack = true,
  statusNotice = null,
  extensionWorkingState = null,
  prefillRequest = null
}) {
  const [content, setContent] = w0("");
  const [searchText, setSearchText] = w0("");
  const [mediaFiles, setMediaFiles] = w0([]);
  const [isDragActive, setIsDragActive] = w0(false);
  const [slashMatches, setSlashMatches] = w0([]);
  const [slashIndex, setSlashIndex] = w0(0);
  const [showSlash, setShowSlash] = w0(false);
  const dynamicCommandsRef = o0(null);
  const [mentionMatches, setMentionMatches] = w0([]);
  const [mentionIndex, setMentionIndex] = w0(0);
  const [showMention, setShowMention] = w0(false);
  const [switchingModel, setSwitchingModel] = w0(false);
  const [showModelPopup, setShowModelPopup] = w0(false);
  const [showSessionPopup, setShowSessionPopup] = w0(false);
  const [modelOptions, setModelOptions] = w0([]);
  const [modelPopupIndex, setModelPopupIndex] = w0(0);
  const [sessionPopupIndex, setSessionPopupIndex] = w0(0);
  const [loadingModels, setLoadingModels] = w0(false);
  const [footerWidth, setFooterWidth] = w0(0);
  const [submitError, setSubmitError] = w0(null);
  const [submitNotice, setSubmitNotice] = w0(null);
  const [statusNoticeNowMs, setStatusNoticeNowMs] = w0(() => Date.now());
  const [extensionWorkingFrameIndex, setExtensionWorkingFrameIndex] = w0(0);
  const textareaRef = o0(null);
  const slashRef = o0(null);
  const mentionRef = o0(null);
  const modelPopupRef = o0(null);
  const modelHintRef = o0(null);
  const sessionPopupRef = o0(null);
  const sessionTriggerRef = o0(null);
  const footerRef = o0(null);
  const popupTypeaheadRef = o0({ value: "", updatedAt: 0 });
  const dragCounterRef = o0(0);
  const renameSessionInProgressRef = o0(false);
  const historyMax = 200;
  const historyStorageKey = getComposeHistoryStorageKey(currentChatJid);
  const normaliseHistory = (items) => {
    const seen = new Set;
    const cleaned = [];
    for (const item of items || []) {
      if (typeof item !== "string")
        continue;
      const trimmed = item.trim();
      if (!trimmed || seen.has(trimmed))
        continue;
      seen.add(trimmed);
      cleaned.push(trimmed);
    }
    return cleaned;
  };
  const loadHistory = (storageKey = historyStorageKey) => {
    const raw = getLocalStorageItem(storageKey);
    if (!raw)
      return [];
    try {
      const parsed = JSON.parse(raw);
      if (!Array.isArray(parsed))
        return [];
      return normaliseHistory(parsed);
    } catch {
      return [];
    }
  };
  const saveHistory = (history, storageKey = historyStorageKey) => {
    setLocalStorageItem(storageKey, JSON.stringify(history));
  };
  const historyRef = o0(loadHistory(historyStorageKey));
  const historyIndexRef = o0(-1);
  const historyDraftRef = o0("");
  const lastPrefillTokenRef = o0("");
  r0(() => {
    historyRef.current = loadHistory(historyStorageKey);
    historyIndexRef.current = -1;
    historyDraftRef.current = "";
  }, [historyStorageKey]);
  r0(() => {
    let cancelled = false;
    const chatJid = currentChatJid || "web:default";
    fetch(`/agent/commands?chat_jid=${encodeURIComponent(chatJid)}`).then((r2) => r2.ok ? r2.json() : null).then((data) => {
      if (cancelled || !data?.commands)
        return;
      dynamicCommandsRef.current = data.commands.map((c2) => ({
        name: c2.name,
        description: c2.description || ""
      }));
    }).catch((e2) => {
      console.debug("[compose] failed to fetch dynamic commands", e2);
    });
    return () => {
      cancelled = true;
    };
  }, [currentChatJid]);
  r0(() => {
    const resolved = resolveComposePrefillRequest(prefillRequest, lastPrefillTokenRef.current, searchMode);
    if (!resolved.shouldApply)
      return;
    lastPrefillTokenRef.current = resolved.nextToken;
    setSubmitError(null);
    setContent(resolved.text);
    updateSlashAutocomplete(resolved.text);
    updateMentionAutocomplete(resolved.text);
    requestAnimationFrame(() => {
      resizeTextarea();
      const textarea = textareaRef.current;
      if (!textarea)
        return;
      try {
        textarea.focus({ preventScroll: true });
      } catch {
        textarea.focus();
      }
      const end = resolved.text.length;
      textarea.setSelectionRange?.(end, end);
    });
  }, [prefillRequest, searchMode]);
  const canSend = content.trim() || mediaFiles.length > 0 || fileRefs.length > 0 || messageRefs.length > 0;
  const canShareLocation = typeof window !== "undefined" && typeof navigator !== "undefined" && Boolean(navigator.geolocation) && Boolean(window.isSecureContext);
  const notificationsSupported = typeof window !== "undefined" && typeof Notification !== "undefined";
  const notificationsSecure = typeof window !== "undefined" ? Boolean(window.isSecureContext) : false;
  const notificationDenied = notificationPermission === "denied";
  const notificationsAvailable = notificationsSupported && notificationsSecure && !notificationDenied;
  const notificationActive = notificationPermission === "granted" && notificationsEnabled;
  const statusNoticeIsCompaction = isCompactionStatus(statusNotice);
  const statusNoticeTitle = resolveStatusPanelTitle(statusNotice);
  const statusNoticeDetail = typeof statusNotice?.detail === "string" && statusNotice.detail.trim() ? statusNotice.detail.trim() : "";
  const statusNoticeElapsedLabel = statusNoticeIsCompaction ? getStatusElapsedLabel(statusNotice, statusNoticeNowMs) : null;
  const extensionWorkingDisplay = resolveComposeExtensionWorkingDisplay(extensionWorkingState, extensionWorkingFrameIndex);
  const extensionWorkingIndicator = extensionWorkingState?.indicator && typeof extensionWorkingState.indicator === "object" ? extensionWorkingState.indicator : null;
  const notificationTitle = notificationActive ? "Disable notifications" : "Enable notifications";
  const hasAttachments = mediaFiles.length > 0 || fileRefs.length > 0 || messageRefs.length > 0;
  const connectionStatusPresentation = useConnectionStatusPresentation(connectionStatus);
  const connectionStatusLabel = connectionStatusPresentation.label;
  const connectionStatusTitle = connectionStatusPresentation.title;
  const submitButtonState = resolveComposeSubmitButtonState(isAgentActive, canSend, statusNoticeIsCompaction);
  const mentionAgents = (Array.isArray(activeChatAgents) ? activeChatAgents : []).filter((chat) => !chat?.archived_at);
  const currentSessionAgent = (() => {
    for (const chat of Array.isArray(activeChatAgents) ? activeChatAgents : []) {
      const chatJid = typeof chat?.chat_jid === "string" ? chat.chat_jid.trim() : "";
      if (chatJid && chatJid === currentChatJid)
        return chat;
    }
    return null;
  })();
  const isCurrentRootSession = Boolean(currentSessionAgent && currentSessionAgent.chat_jid === (currentSessionAgent.root_chat_jid || currentSessionAgent.chat_jid));
  const switchableChatAgents = G0(() => {
    const seen = new Set;
    const chats = [];
    for (const chat of Array.isArray(activeChatAgents) ? activeChatAgents : []) {
      const chatJid = typeof chat?.chat_jid === "string" ? chat.chat_jid.trim() : "";
      if (!chatJid || chatJid === currentChatJid || seen.has(chatJid))
        continue;
      const agentName = typeof chat?.agent_name === "string" ? chat.agent_name.trim() : "";
      if (!agentName)
        continue;
      seen.add(chatJid);
      chats.push(chat);
    }
    return chats;
  }, [activeChatAgents, currentChatJid]);
  const hasSwitchableChatAgents = switchableChatAgents.length > 0;
  const canSwitchSession = hasSwitchableChatAgents && typeof onSwitchChat === "function";
  const canRestoreSession = hasSwitchableChatAgents && typeof onRestoreSession === "function";
  const renameInProgress = Boolean(isRenameSessionInProgress || renameSessionInProgressRef.current);
  const canRenameSession = !searchMode && typeof onRenameSession === "function" && !renameInProgress;
  const canCreateSession = !searchMode && typeof onCreateSession === "function";
  const canDeleteSession = !searchMode && typeof onDeleteSession === "function" && !isCurrentRootSession;
  const showSessionSwitcherButton = !searchMode && (canSwitchSession || canRestoreSession || canRenameSession || canCreateSession || canDeleteSession);
  const modelPickerState = resolveComposeModelPickerState(activeModel, agentModelsPayload);
  const showModelPickerHint = modelPickerState.showPicker;
  const modelHintLabel = modelPickerState.label;
  const modelHintSuffix = supportsThinking && thinkingLevel ? ` (${thinkingLevel})` : "";
  const modelThinkingLabel = modelHintSuffix.trim() ? `${thinkingLevel}` : "";
  const modelUsageLabel = typeof modelUsage?.hint_short === "string" ? modelUsage.hint_short.trim() : "";
  const modelUsageSectionLabel = [
    modelThinkingLabel || null,
    modelUsageLabel || null
  ].filter(Boolean).join(" • ");
  const modelUsageTitleParts = [
    activeModel ? `Current model: ${modelHintLabel}${modelHintSuffix}` : null,
    modelUsage?.plan ? `Plan: ${modelUsage.plan}` : null,
    modelUsageLabel || null,
    modelUsage?.primary?.reset_description || null,
    modelUsage?.secondary?.reset_description || null
  ].filter(Boolean);
  const modelHintTitle = switchingModel ? "Switching model…" : modelUsageTitleParts.join(" • ") || (showModelPickerHint ? "Select a model (tap to open model picker)" : `Current model: ${modelHintLabel}${modelHintSuffix} (tap to open model picker)`);
  const showComposeMetaRow = !searchMode && (showModelPickerHint || contextUsage && contextUsage.percent != null);
  const emitModelState = (payload) => {
    if (!payload || typeof payload !== "object")
      return;
    const modelLabel = payload.model ?? payload.current;
    if (typeof onModelStateChange === "function") {
      onModelStateChange({
        model: modelLabel ?? null,
        thinking_level: payload.thinking_level ?? null,
        thinking_level_label: payload.thinking_level_label ?? null,
        supports_thinking: payload.supports_thinking,
        provider_usage: payload.provider_usage ?? null
      });
    }
    if (modelLabel && typeof onModelChange === "function") {
      onModelChange(modelLabel);
    }
  };
  const resizeTextarea = (target) => {
    const textarea = target || textareaRef.current;
    if (!textarea)
      return;
    textarea.style.height = "auto";
    textarea.style.height = `${textarea.scrollHeight}px`;
    textarea.style.overflowY = "hidden";
  };
  const updateSlashAutocomplete = (value) => {
    if (!value.startsWith("/") || value.includes(`
`)) {
      setShowSlash(false);
      setSlashMatches([]);
      return;
    }
    const prefix = value.toLowerCase().split(" ")[0];
    if (prefix.length < 1) {
      setShowSlash(false);
      setSlashMatches([]);
      return;
    }
    const commandList = dynamicCommandsRef.current || SLASH_COMMANDS;
    const matches = commandList.filter((cmd) => cmd.name.startsWith(prefix) || cmd.name.replace(/-/g, "").startsWith(prefix.replace(/-/g, "")));
    if (matches.length > 0 && !(matches.length === 1 && matches[0].name === prefix)) {
      setShowMention(false);
      setMentionMatches([]);
      setSlashMatches(matches);
      setSlashIndex(0);
      setShowSlash(true);
    } else {
      setShowSlash(false);
      setSlashMatches([]);
    }
  };
  const acceptSlashCommand = (cmd) => {
    const current = content;
    const spaceIdx = current.indexOf(" ");
    const args = spaceIdx >= 0 ? current.slice(spaceIdx) : "";
    const newVal = cmd.name + args;
    setContent(newVal);
    setShowSlash(false);
    setSlashMatches([]);
    requestAnimationFrame(() => {
      const textarea = textareaRef.current;
      if (!textarea)
        return;
      const len = newVal.length;
      textarea.selectionStart = len;
      textarea.selectionEnd = len;
      textarea.focus();
    });
  };
  const updateMentionAutocomplete = (value) => {
    if (parseMentionAutocompleteQuery(value) == null) {
      setShowMention(false);
      setMentionMatches([]);
      return;
    }
    const matches = filterMentionAgents(mentionAgents, value, { currentChatJid });
    if (matches.length > 0 && !(matches.length === 1 && buildMentionValue(matches[0].agent_name).trim().toLowerCase() === String(value || "").trim().toLowerCase())) {
      setShowSlash(false);
      setSlashMatches([]);
      setMentionMatches(matches);
      setMentionIndex(0);
      setShowMention(true);
    } else {
      setShowMention(false);
      setMentionMatches([]);
    }
  };
  const acceptMention = (agent) => {
    const newVal = buildMentionValue(agent?.agent_name);
    if (!newVal)
      return;
    setContent(newVal);
    setShowMention(false);
    setMentionMatches([]);
    requestAnimationFrame(() => {
      const textarea = textareaRef.current;
      if (!textarea)
        return;
      const len = newVal.length;
      textarea.selectionStart = len;
      textarea.selectionEnd = len;
      textarea.focus();
    });
  };
  const openSessionPopup = () => {
    if (searchMode || !canSwitchSession && !canRestoreSession && !canRenameSession && !canCreateSession && !canDeleteSession)
      return false;
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    setShowModelPopup(false);
    setShowSlash(false);
    setSlashMatches([]);
    setShowMention(false);
    setMentionMatches([]);
    setShowSessionPopup(true);
    return true;
  };
  const toggleSessionPopup = (event) => {
    event?.preventDefault?.();
    event?.stopPropagation?.();
    if (searchMode || !canSwitchSession && !canRestoreSession && !canRenameSession && !canCreateSession && !canDeleteSession)
      return;
    if (showSessionPopup) {
      popupTypeaheadRef.current = { value: "", updatedAt: 0 };
      setShowSessionPopup(false);
      return;
    }
    openSessionPopup();
  };
  const handleSessionSwitch = (chatJid) => {
    const nextChatJid = typeof chatJid === "string" ? chatJid.trim() : "";
    setShowSessionPopup(false);
    if (!nextChatJid || nextChatJid === currentChatJid) {
      requestAnimationFrame(() => textareaRef.current?.focus());
      return;
    }
    onSwitchChat?.(nextChatJid);
  };
  const handleRestoreSession = async (chatJid) => {
    const nextChatJid = typeof chatJid === "string" ? chatJid.trim() : "";
    setShowSessionPopup(false);
    if (!nextChatJid || typeof onRestoreSession !== "function") {
      requestAnimationFrame(() => textareaRef.current?.focus());
      return;
    }
    try {
      await onRestoreSession(nextChatJid);
    } catch (error) {
      console.warn("Failed to restore session:", error);
      requestAnimationFrame(() => textareaRef.current?.focus());
    }
  };
  const findFirstEnabledPopupIndex = (items) => {
    const list = Array.isArray(items) ? items : [];
    const index = list.findIndex((item) => !item?.disabled);
    return index >= 0 ? index : 0;
  };
  const sessionPopupEntries = G0(() => {
    const entries = [];
    for (const chat of switchableChatAgents) {
      const archived = Boolean(chat?.archived_at);
      const agentName = typeof chat?.agent_name === "string" ? chat.agent_name.trim() : "";
      const chatJid = typeof chat?.chat_jid === "string" ? chat.chat_jid.trim() : "";
      if (!agentName || !chatJid)
        continue;
      entries.push({
        type: "session",
        key: `session:${chatJid}`,
        label: `@${agentName} — ${chatJid}${chat?.is_active ? " active" : ""}${archived ? " archived" : ""}`,
        chat,
        disabled: archived ? !canRestoreSession : !canSwitchSession
      });
    }
    if (canCreateSession) {
      entries.push({ type: "action", key: "action:new", label: "New session", action: "new", disabled: false });
    }
    if (canRenameSession) {
      entries.push({ type: "action", key: "action:rename", label: "Rename current session", action: "rename", disabled: renameInProgress });
    }
    if (canDeleteSession) {
      entries.push({ type: "action", key: "action:delete", label: "Delete current session", action: "delete", disabled: false });
    }
    return entries;
  }, [switchableChatAgents, canRestoreSession, canSwitchSession, canCreateSession, canRenameSession, canDeleteSession, renameInProgress]);
  const handleRenameSession = async (event) => {
    if (event?.preventDefault)
      event.preventDefault();
    if (event?.stopPropagation)
      event.stopPropagation();
    if (typeof onRenameSession !== "function" || isRenameSessionInProgress || renameSessionInProgressRef.current)
      return;
    renameSessionInProgressRef.current = true;
    setShowSessionPopup(false);
    try {
      await onRenameSession();
    } catch (error) {
      console.warn("Failed to rename session:", error);
    } finally {
      renameSessionInProgressRef.current = false;
    }
    requestAnimationFrame(() => textareaRef.current?.focus());
  };
  const handleCreateSession = async () => {
    if (typeof onCreateSession !== "function")
      return;
    setShowSessionPopup(false);
    try {
      await onCreateSession();
    } catch (error) {
      console.warn("Failed to create session:", error);
    }
    requestAnimationFrame(() => textareaRef.current?.focus());
  };
  const handleDeleteSession = async () => {
    if (typeof onDeleteSession !== "function")
      return;
    setShowSessionPopup(false);
    try {
      await onDeleteSession(currentChatJid);
    } catch (error) {
      console.warn("Failed to delete session:", error);
    }
    requestAnimationFrame(() => textareaRef.current?.focus());
  };
  const updateValue = (value) => {
    if (searchMode) {
      setSearchText(value);
    } else {
      setContent(value);
      updateSlashAutocomplete(value);
      updateMentionAutocomplete(value);
    }
    requestAnimationFrame(() => resizeTextarea());
  };
  const appendToValue = (snippet) => {
    const current = searchMode ? searchText : content;
    const prefix = current && !current.endsWith(`
`) ? `
` : "";
    const next = `${current}${prefix}${snippet}`.trimStart();
    updateValue(next);
  };
  const extractCurrentModel = (response) => {
    const fromLabel = response?.command?.model_label;
    if (fromLabel)
      return fromLabel;
    const message = response?.command?.message;
    if (typeof message === "string") {
      const currentMatch = message.match(/•\s+([^\n]+?)\s+\(current\)/);
      if (currentMatch?.[1])
        return currentMatch[1].trim();
    }
    return null;
  };
  const runModelCommand = async (commandText) => {
    if (searchMode || switchingModel)
      return;
    setSubmitError(null);
    setSubmitNotice(null);
    setSwitchingModel(true);
    try {
      const response = await sendAgentMessage("default", commandText, null, [], null, currentChatJid);
      const nextModel = extractCurrentModel(response);
      emitModelState({
        model: nextModel ?? activeModel ?? null,
        thinking_level: response?.command?.thinking_level,
        thinking_level_label: response?.command?.thinking_level_label,
        supports_thinking: response?.command?.supports_thinking
      });
      await refreshAgentModelStateBestEffort(getAgentModels, currentChatJid, emitModelState);
      setSubmitNotice(resolveUiOnlyCommandNotice(commandText, response));
      onPost?.(response);
      return true;
    } catch (error) {
      console.error("Failed to switch model:", error);
      alert("Failed to switch model: " + error.message);
      return false;
    } finally {
      setSwitchingModel(false);
    }
  };
  const handleCycleModel = async () => {
    await runModelCommand("/cycle-model");
  };
  const handleSelectModel = async (modelOption) => {
    const modelLabel = typeof modelOption === "string" ? modelOption : typeof modelOption?.label === "string" ? modelOption.label : "";
    if (!modelLabel || switchingModel)
      return;
    const ok = await runModelCommand(`/model ${modelLabel}`);
    if (ok)
      setShowModelPopup(false);
  };
  const runSessionPopupEntry = (entry) => {
    if (!entry || entry.disabled)
      return;
    if (entry.type === "session") {
      const chat = entry.chat;
      if (chat?.archived_at) {
        handleRestoreSession(chat.chat_jid);
      } else {
        handleSessionSwitch(chat.chat_jid);
      }
      return;
    }
    if (entry.type === "action") {
      if (entry.action === "new") {
        handleCreateSession();
        return;
      }
      if (entry.action === "rename") {
        handleRenameSession();
        return;
      }
      if (entry.action === "delete") {
        handleDeleteSession();
      }
    }
  };
  const toggleModelPopup = (event) => {
    event.preventDefault();
    event.stopPropagation();
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    setShowSessionPopup(false);
    setShowModelPopup((prev) => !prev);
  };
  const handleContextCompact = async () => {
    if (searchMode)
      return;
    onContextCompact?.();
    await handleSubmit("/compact", null, {
      includeMedia: false,
      includeFileRefs: false,
      includeMessageRefs: false,
      clearAfterSubmit: false,
      recordHistory: false
    });
  };
  const resolveSubmitMode = (mode) => {
    if (mode === "queue" || mode === "steer" || mode === "auto") {
      return mode;
    }
    return isAgentActive ? "queue" : undefined;
  };
  const handleSubmit = async (overrideContent, submitMode, submitOptions = {}) => {
    const {
      includeMedia = true,
      includeFileRefs = true,
      includeMessageRefs = true,
      clearAfterSubmit = true,
      recordHistory = true
    } = submitOptions || {};
    const inferred = typeof overrideContent === "string" ? overrideContent : overrideContent && typeof overrideContent?.target?.value === "string" ? overrideContent.target.value : content;
    const currentContent = typeof inferred === "string" ? inferred : "";
    if (!currentContent.trim() && (includeMedia ? mediaFiles.length === 0 : true) && (includeFileRefs ? fileRefs.length === 0 : true) && (includeMessageRefs ? messageRefs.length === 0 : true))
      return;
    setShowSlash(false);
    setSlashMatches([]);
    setShowMention(false);
    setMentionMatches([]);
    setShowSessionPopup(false);
    setSubmitError(null);
    setSubmitNotice(null);
    const capturedMediaFiles = includeMedia ? [...mediaFiles] : [];
    const capturedFileRefs = includeFileRefs ? [...fileRefs] : [];
    const capturedMessageRefs = includeMessageRefs ? [...messageRefs] : [];
    const baseContent = currentContent.trim();
    if (recordHistory && baseContent) {
      const current = historyRef.current;
      const deduped = normaliseHistory(current.filter((item) => item !== baseContent));
      deduped.push(baseContent);
      if (deduped.length > historyMax) {
        deduped.splice(0, deduped.length - historyMax);
      }
      historyRef.current = deduped;
      saveHistory(deduped);
      historyIndexRef.current = -1;
      historyDraftRef.current = "";
    }
    const restoreDraft = () => {
      if (includeMedia)
        setMediaFiles([...capturedMediaFiles]);
      if (includeFileRefs)
        onSetFileRefs?.(capturedFileRefs);
      if (includeMessageRefs)
        onSetMessageRefs?.(capturedMessageRefs);
      setContent(baseContent);
      requestAnimationFrame(() => resizeTextarea());
    };
    if (clearAfterSubmit) {
      setContent("");
      setMediaFiles([]);
      onClearFileRefs?.();
      onClearMessageRefs?.();
    }
    (async () => {
      try {
        const intercepted = await onSubmitIntercept?.({
          content: baseContent,
          submitMode,
          fileRefs: capturedFileRefs,
          messageRefs: capturedMessageRefs,
          mediaFiles: capturedMediaFiles
        });
        if (intercepted) {
          onPost?.(intercepted);
          return;
        }
        const mediaIds = [];
        for (const file of capturedMediaFiles) {
          const result = await uploadMedia(file);
          mediaIds.push(result.id);
        }
        const fileBlock = capturedFileRefs.length ? `Files:
${capturedFileRefs.map((path) => `- ${path}`).join(`
`)}` : "";
        const messageRefBlock = capturedMessageRefs.length ? `Referenced messages:
${capturedMessageRefs.map((id) => `- message:${id}`).join(`
`)}` : "";
        const mediaBlock = mediaIds.length ? `Attachments:
${mediaIds.map((id, index) => {
          const file = capturedMediaFiles[index];
          const label = file?.name || `attachment-${index + 1}`;
          return `- attachment:${id} (${label})`;
        }).join(`
`)}` : "";
        const message = [baseContent, fileBlock, messageRefBlock, mediaBlock].filter(Boolean).join(`

`);
        const response = await sendAgentMessage("default", message, null, mediaIds, resolveSubmitMode(submitMode), currentChatJid);
        onMessageResponse?.(response);
        if (response?.command) {
          emitModelState({
            model: response.command.model_label ?? activeModel ?? null,
            thinking_level: response.command.thinking_level,
            thinking_level_label: response.command.thinking_level_label,
            supports_thinking: response.command.supports_thinking
          });
          await refreshAgentModelStateBestEffort(getAgentModels, currentChatJid, emitModelState);
        }
        setSubmitNotice(resolveUiOnlyCommandNotice(baseContent, response));
        onPost?.(response);
      } catch (error) {
        if (clearAfterSubmit) {
          restoreDraft();
        }
        const message = error?.message || "Failed to send message.";
        setSubmitError(message);
        onSubmitError?.(message);
        console.error("Failed to post:", error);
      }
    })();
  };
  const handleInjectQueuedFollowup = (queuedItem) => {
    onInjectQueuedFollowup?.(queuedItem);
  };
  const handlePopupKeyboardEvent = t0((e2) => {
    if (searchMode || !showModelPopup && !showSessionPopup || e2?.isComposing)
      return false;
    const consume = () => {
      e2.preventDefault?.();
      e2.stopPropagation?.();
    };
    const resetPopupTypeahead = () => {
      popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    };
    if (e2.key === "Escape") {
      consume();
      resetPopupTypeahead();
      if (showModelPopup)
        setShowModelPopup(false);
      if (showSessionPopup)
        setShowSessionPopup(false);
      return true;
    }
    if (showModelPopup) {
      if (e2.key === "ArrowDown") {
        consume();
        resetPopupTypeahead();
        if (modelOptions.length > 0)
          setModelPopupIndex((idx) => (idx + 1) % modelOptions.length);
        return true;
      }
      if (e2.key === "ArrowUp") {
        consume();
        resetPopupTypeahead();
        if (modelOptions.length > 0)
          setModelPopupIndex((idx) => (idx - 1 + modelOptions.length) % modelOptions.length);
        return true;
      }
      if ((e2.key === "Enter" || e2.key === "Tab") && modelOptions.length > 0) {
        consume();
        resetPopupTypeahead();
        handleSelectModel(modelOptions[Math.max(0, Math.min(modelPopupIndex, modelOptions.length - 1))]);
        return true;
      }
      if (isPopupTypeaheadKey(e2) && modelOptions.length > 0) {
        consume();
        const nextBuffer = updatePopupTypeaheadBuffer(popupTypeaheadRef.current, e2.key);
        popupTypeaheadRef.current = nextBuffer;
        const match = resolvePopupTypeaheadMatch(modelOptions, nextBuffer.value, modelPopupIndex, (item) => getModelPickerOptionSearchLabel(item));
        if (match >= 0)
          setModelPopupIndex(match);
        return true;
      }
    }
    if (showSessionPopup) {
      if (e2.key === "ArrowDown") {
        consume();
        resetPopupTypeahead();
        if (sessionPopupEntries.length > 0)
          setSessionPopupIndex((idx) => (idx + 1) % sessionPopupEntries.length);
        return true;
      }
      if (e2.key === "ArrowUp") {
        consume();
        resetPopupTypeahead();
        if (sessionPopupEntries.length > 0)
          setSessionPopupIndex((idx) => (idx - 1 + sessionPopupEntries.length) % sessionPopupEntries.length);
        return true;
      }
      if ((e2.key === "Enter" || e2.key === "Tab") && sessionPopupEntries.length > 0) {
        consume();
        resetPopupTypeahead();
        runSessionPopupEntry(sessionPopupEntries[Math.max(0, Math.min(sessionPopupIndex, sessionPopupEntries.length - 1))]);
        return true;
      }
      if (isPopupTypeaheadKey(e2) && sessionPopupEntries.length > 0) {
        consume();
        const nextBuffer = updatePopupTypeaheadBuffer(popupTypeaheadRef.current, e2.key);
        popupTypeaheadRef.current = nextBuffer;
        const match = resolvePopupTypeaheadMatch(sessionPopupEntries, nextBuffer.value, sessionPopupIndex, (item) => item.label);
        if (match >= 0)
          setSessionPopupIndex(match);
        return true;
      }
    }
    return false;
  }, [
    searchMode,
    showModelPopup,
    showSessionPopup,
    modelOptions,
    modelPopupIndex,
    sessionPopupEntries,
    sessionPopupIndex,
    handleSelectModel
  ]);
  const handleKeyDown = (e2) => {
    if (e2.isComposing)
      return;
    if (searchMode && e2.key === "Escape") {
      e2.preventDefault();
      setSearchText("");
      onExitSearch?.();
      return;
    }
    if (handlePopupKeyboardEvent(e2)) {
      return;
    }
    const currentValue = textareaRef.current?.value ?? (searchMode ? searchText : content);
    if (shouldOpenSessionSwitcherFromBlankCompose(e2, currentValue, {
      searchMode,
      showSessionSwitcherButton
    })) {
      e2.preventDefault();
      openSessionPopup();
      return;
    }
    if (showMention && mentionMatches.length > 0) {
      const mentionValue = textareaRef.current?.value ?? (searchMode ? searchText : content);
      if (!String(mentionValue || "").match(/^@([a-zA-Z0-9_-]*)$/)) {
        setShowMention(false);
        setMentionMatches([]);
      } else {
        if (e2.key === "ArrowDown") {
          e2.preventDefault();
          setMentionIndex((i2) => (i2 + 1) % mentionMatches.length);
          return;
        }
        if (e2.key === "ArrowUp") {
          e2.preventDefault();
          setMentionIndex((i2) => (i2 - 1 + mentionMatches.length) % mentionMatches.length);
          return;
        }
        if (e2.key === "Tab" || e2.key === "Enter") {
          e2.preventDefault();
          acceptMention(mentionMatches[mentionIndex]);
          return;
        }
        if (e2.key === "Escape") {
          e2.preventDefault();
          setShowMention(false);
          setMentionMatches([]);
          return;
        }
      }
    }
    if (showSlash && slashMatches.length > 0) {
      const slashValue = textareaRef.current?.value ?? (searchMode ? searchText : content);
      if (!String(slashValue || "").startsWith("/")) {
        setShowSlash(false);
        setSlashMatches([]);
      } else {
        if (e2.key === "ArrowDown") {
          e2.preventDefault();
          setSlashIndex((i2) => (i2 + 1) % slashMatches.length);
          return;
        }
        if (e2.key === "ArrowUp") {
          e2.preventDefault();
          setSlashIndex((i2) => (i2 - 1 + slashMatches.length) % slashMatches.length);
          return;
        }
        if (e2.key === "Tab") {
          e2.preventDefault();
          acceptSlashCommand(slashMatches[slashIndex]);
          return;
        }
        if (e2.key === "Enter" && !e2.shiftKey) {
          const hasArgs = currentValue.includes(" ");
          if (!hasArgs) {
            e2.preventDefault();
            const cmd = slashMatches[slashIndex];
            setShowSlash(false);
            setSlashMatches([]);
            handleSubmit(cmd.name);
            return;
          }
        }
        if (e2.key === "Escape") {
          e2.preventDefault();
          setShowSlash(false);
          setSlashMatches([]);
          return;
        }
      }
    }
    if (!searchMode && (e2.key === "ArrowUp" || e2.key === "ArrowDown") && !e2.metaKey && !e2.ctrlKey && !e2.altKey && !e2.shiftKey) {
      const textarea = textareaRef.current;
      if (!textarea)
        return;
      const value = textarea.value || "";
      const atStart = textarea.selectionStart === 0 && textarea.selectionEnd === 0;
      const atEnd = textarea.selectionStart === value.length && textarea.selectionEnd === value.length;
      if (e2.key === "ArrowUp" && atStart || e2.key === "ArrowDown" && atEnd) {
        const history = historyRef.current;
        if (!history.length)
          return;
        e2.preventDefault();
        let idx = historyIndexRef.current;
        if (e2.key === "ArrowUp") {
          if (idx === -1) {
            historyDraftRef.current = value;
            idx = history.length - 1;
          } else if (idx > 0) {
            idx -= 1;
          }
          historyIndexRef.current = idx;
          updateValue(history[idx] || "");
        } else {
          if (idx === -1)
            return;
          if (idx < history.length - 1) {
            idx += 1;
            historyIndexRef.current = idx;
            updateValue(history[idx] || "");
          } else {
            historyIndexRef.current = -1;
            updateValue(historyDraftRef.current || "");
            historyDraftRef.current = "";
          }
        }
        requestAnimationFrame(() => {
          const target = textareaRef.current;
          if (!target)
            return;
          const len = target.value.length;
          target.selectionStart = len;
          target.selectionEnd = len;
        });
        return;
      }
    }
    if (e2.key === "Enter" && !e2.shiftKey && (e2.ctrlKey || e2.metaKey)) {
      e2.preventDefault();
      if (searchMode) {
        if (currentValue.trim()) {
          onSearch?.(currentValue.trim(), searchScope);
        }
      } else {
        handleSubmit(currentValue, "steer");
      }
      return;
    }
    if (e2.key === "Enter" && !e2.shiftKey) {
      e2.preventDefault();
      if (searchMode) {
        if (currentValue.trim()) {
          onSearch?.(currentValue.trim(), searchScope);
        }
      } else {
        handleSubmit(currentValue);
      }
    }
  };
  const addMediaFiles = (files) => {
    const list = Array.from(files || []).filter((file) => file instanceof File && !String(file.name || "").startsWith(".DS_Store"));
    if (!list.length)
      return;
    setMediaFiles((current) => [...current, ...list]);
    setSubmitError(null);
  };
  const handleFileChange = (e2) => {
    addMediaFiles(e2.target.files);
    e2.target.value = "";
  };
  const handleDragEnter = (e2) => {
    if (searchMode)
      return;
    e2.preventDefault();
    e2.stopPropagation();
    dragCounterRef.current += 1;
    setIsDragActive(true);
  };
  const handleDragLeave = (e2) => {
    if (searchMode)
      return;
    e2.preventDefault();
    e2.stopPropagation();
    dragCounterRef.current = Math.max(0, dragCounterRef.current - 1);
    if (dragCounterRef.current === 0)
      setIsDragActive(false);
  };
  const handleDragOver = (e2) => {
    if (searchMode)
      return;
    e2.preventDefault();
    e2.stopPropagation();
    if (e2.dataTransfer)
      e2.dataTransfer.dropEffect = "copy";
    setIsDragActive(true);
  };
  const handleDrop = (e2) => {
    if (searchMode)
      return;
    e2.preventDefault();
    e2.stopPropagation();
    dragCounterRef.current = 0;
    setIsDragActive(false);
    addMediaFiles(e2.dataTransfer?.files || []);
  };
  const handlePaste = (e2) => {
    if (searchMode)
      return;
    const items = e2.clipboardData?.items;
    if (!items || !items.length)
      return;
    const files = [];
    for (const item of items) {
      if (item.kind !== "file")
        continue;
      const file = item.getAsFile?.();
      if (file)
        files.push(file);
    }
    if (files.length > 0) {
      e2.preventDefault();
      addMediaFiles(files);
    }
  };
  const removeMediaFile = (index) => {
    setMediaFiles((current) => current.filter((_2, idx) => idx !== index));
  };
  const clearAllAttachmentRefs = () => {
    setSubmitError(null);
    setMediaFiles([]);
    onClearFileRefs?.();
    onClearMessageRefs?.();
  };
  const handleLocation = () => {
    if (!navigator.geolocation) {
      alert("Geolocation is not available in this browser.");
      return;
    }
    navigator.geolocation.getCurrentPosition((pos) => {
      const { latitude, longitude, accuracy } = pos.coords;
      const coords = `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`;
      const accuracyLabel = Number.isFinite(accuracy) ? ` ±${Math.round(accuracy)}m` : "";
      const mapLink = `https://maps.google.com/?q=${latitude},${longitude}`;
      const snippet = `Location: ${coords}${accuracyLabel} ${mapLink}`;
      appendToValue(snippet);
    }, (err) => {
      const message = err?.message || "Unable to retrieve location.";
      alert(`Location error: ${message}`);
    }, { enableHighAccuracy: true, timeout: 1e4, maximumAge: 0 });
  };
  r0(() => {
    if (!showModelPopup)
      return;
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    setLoadingModels(true);
    getAgentModels(currentChatJid).then((payload) => {
      setModelOptions(normalizeModelPickerOptions(payload));
      emitModelState(payload);
    }).catch((error) => {
      console.warn("Failed to load model list:", error);
      setModelOptions([]);
    }).finally(() => {
      setLoadingModels(false);
    });
  }, [showModelPopup, activeModel]);
  r0(() => {
    if (searchMode) {
      setShowModelPopup(false);
      setShowSessionPopup(false);
      setShowSlash(false);
      setSlashMatches([]);
      setShowMention(false);
      setMentionMatches([]);
    }
  }, [searchMode]);
  r0(() => {
    if (showSessionPopup && !showSessionSwitcherButton) {
      setShowSessionPopup(false);
    }
  }, [showSessionPopup, showSessionSwitcherButton]);
  r0(() => {
    if (!showModelPopup)
      return;
    const activeIndex = modelOptions.findIndex((model) => model?.label === activeModel);
    setModelPopupIndex(activeIndex >= 0 ? activeIndex : 0);
  }, [showModelPopup, modelOptions, activeModel]);
  r0(() => {
    if (!showSessionPopup)
      return;
    setSessionPopupIndex(findFirstEnabledPopupIndex(sessionPopupEntries));
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
  }, [showSessionPopup, currentChatJid]);
  r0(() => {
    if (!showModelPopup)
      return;
    const onPointerDown = (event) => {
      const popup = modelPopupRef.current;
      const hint = modelHintRef.current;
      const target = event.target;
      if (popup && popup.contains(target))
        return;
      if (hint && hint.contains(target))
        return;
      setShowModelPopup(false);
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [showModelPopup]);
  r0(() => {
    if (!showSessionPopup)
      return;
    const onPointerDown = (event) => {
      const popup = sessionPopupRef.current;
      const trigger = sessionTriggerRef.current;
      const target = event.target;
      if (popup && popup.contains(target))
        return;
      if (trigger && trigger.contains(target))
        return;
      setShowSessionPopup(false);
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, [showSessionPopup]);
  r0(() => {
    if (searchMode || !showModelPopup && !showSessionPopup)
      return;
    const onKeyDown = (event) => {
      handlePopupKeyboardEvent(event);
    };
    document.addEventListener("keydown", onKeyDown, true);
    return () => document.removeEventListener("keydown", onKeyDown, true);
  }, [searchMode, showModelPopup, showSessionPopup, handlePopupKeyboardEvent]);
  r0(() => {
    if (!showModelPopup)
      return;
    const popup = modelPopupRef.current;
    popup?.focus?.();
    const active = popup?.querySelector?.(".compose-model-popup-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showModelPopup, modelPopupIndex, modelOptions]);
  r0(() => {
    if (!showSessionPopup)
      return;
    const popup = sessionPopupRef.current;
    popup?.focus?.();
    const active = popup?.querySelector?.(".compose-model-popup-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showSessionPopup, sessionPopupIndex, sessionPopupEntries.length]);
  r0(() => {
    if (!showMention || !mentionRef.current)
      return;
    const popup = mentionRef.current;
    const active = popup.querySelector?.(".slash-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showMention, mentionIndex, mentionMatches.length]);
  r0(() => {
    if (!showSlash || !slashRef.current)
      return;
    const popup = slashRef.current;
    const active = popup.querySelector?.(".slash-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showSlash, slashIndex, slashMatches.length]);
  r0(() => {
    const updateFooterWidth = () => {
      const width = footerRef.current?.clientWidth || 0;
      setFooterWidth((current) => current === width ? current : width);
    };
    updateFooterWidth();
    const footer = footerRef.current;
    let observerFrame = 0;
    const scheduleFooterResize = () => {
      if (observerFrame) {
        cancelAnimationFrame(observerFrame);
      }
      observerFrame = requestAnimationFrame(() => {
        observerFrame = 0;
        updateFooterWidth();
      });
    };
    let observer = null;
    if (footer && typeof ResizeObserver !== "undefined") {
      observer = new ResizeObserver(() => scheduleFooterResize());
      observer.observe(footer);
    }
    if (typeof window !== "undefined") {
      window.addEventListener("resize", scheduleFooterResize);
    }
    return () => {
      if (observerFrame) {
        cancelAnimationFrame(observerFrame);
      }
      observer?.disconnect?.();
      if (typeof window !== "undefined") {
        window.removeEventListener("resize", scheduleFooterResize);
      }
    };
  }, [searchMode, activeModel, currentSessionAgent?.agent_name, showSessionSwitcherButton, contextUsage?.percent]);
  const handleInput = (e2) => {
    const value = e2.target.value;
    setSubmitError(null);
    setSubmitNotice(null);
    if (showSessionPopup)
      setShowSessionPopup(false);
    resizeTextarea(e2.target);
    updateValue(value);
  };
  r0(() => {
    requestAnimationFrame(() => resizeTextarea());
  }, [content, searchText, searchMode]);
  r0(() => {
    if (!statusNoticeIsCompaction)
      return;
    setStatusNoticeNowMs(Date.now());
    const timer = setInterval(() => setStatusNoticeNowMs(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [statusNoticeIsCompaction, statusNotice?.started_at, statusNotice?.startedAt]);
  r0(() => {
    setExtensionWorkingFrameIndex(0);
    if (extensionWorkingIndicator?.mode !== "custom" || !Array.isArray(extensionWorkingIndicator.frames) || extensionWorkingIndicator.frames.length <= 1) {
      return;
    }
    const intervalMs = typeof extensionWorkingIndicator.intervalMs === "number" && Number.isFinite(extensionWorkingIndicator.intervalMs) && extensionWorkingIndicator.intervalMs > 0 ? extensionWorkingIndicator.intervalMs : 120;
    const timer = setInterval(() => {
      setExtensionWorkingFrameIndex((prev) => (prev + 1) % extensionWorkingIndicator.frames.length);
    }, intervalMs);
    return () => clearInterval(timer);
  }, [extensionWorkingIndicator]);
  r0(() => {
    if (searchMode)
      return;
    updateMentionAutocomplete(content);
  }, [mentionAgents, currentChatJid, content, searchMode]);
  return X1`
        <div class="compose-box">
            ${showQueueStack && !searchMode && X1`
                <${QueuedFollowupStack}
                    items=${followupQueueItems}
                    onInjectQueuedFollowup=${handleInjectQueuedFollowup}
                    onRemoveQueuedFollowup=${onRemoveQueuedFollowup}
                    onMoveQueuedFollowup=${onMoveQueuedFollowup}
                    onOpenFilePill=${onOpenFilePill}
                />
            `}
            ${extensionWorkingDisplay.visible && X1`
                <div class="compose-inline-status extension-working" role="status" aria-live="polite">
                    <div class="compose-inline-status-row">
                        ${extensionWorkingDisplay.indicatorText ? X1`<span class="compose-inline-status-glyph" aria-hidden="true">${extensionWorkingDisplay.indicatorText}</span>` : extensionWorkingDisplay.animateDot ? X1`<span class=${buildComposeStatusDotClass({ pulsing: true })} aria-hidden="true"></span>` : null}
                        <span class="compose-inline-status-title">${extensionWorkingDisplay.title}</span>
                    </div>
                </div>
            `}
            ${statusNotice && X1`
                <div
                    class=${`compose-inline-status${statusNoticeIsCompaction ? " compaction" : ""}`}
                    role="status"
                    aria-live="polite"
                    title=${statusNoticeDetail || ""}
                >
                    <div class="compose-inline-status-row">
                        <span class=${buildComposeStatusDotClass({ pulsing: statusNoticeIsCompaction })} aria-hidden="true"></span>
                        <span class="compose-inline-status-title">${statusNoticeTitle}</span>
                        ${statusNoticeElapsedLabel && X1`<span class="compose-inline-status-elapsed">${statusNoticeElapsedLabel}</span>`}
                    </div>
                    ${statusNoticeDetail && X1`<div class="compose-inline-status-detail">${statusNoticeDetail}</div>`}
                </div>
            `}
            ${submitNotice && X1`
                <div class="compose-inline-status compose-command-notice" role="status" aria-live="polite">
                    <div class="compose-inline-status-detail compose-command-notice-text">${submitNotice}</div>
                </div>
            `}
            <div
                class=${`compose-input-wrapper${isDragActive ? " drag-active" : ""}`}
                onDragEnter=${handleDragEnter}
                onDragOver=${handleDragOver}
                onDragLeave=${handleDragLeave}
                onDrop=${handleDrop}
            >
                <div class="compose-input-main">
                    ${hasAttachments && X1`
                        <div class="compose-file-refs">
                            ${messageRefs.map((id) => {
    return X1`
                                    <${FilePill}
                                        key=${"msg-" + id}
                                        prefix="compose"
                                        label=${"msg:" + id}
                                        title=${"Message reference: " + id}
                                        removeTitle="Remove reference"
                                        icon="message"
                                        onRemove=${() => onRemoveMessageRef?.(id)}
                                    />
                                `;
  })}
                            ${fileRefs.map((path) => {
    const label = path.split("/").pop() || path;
    return X1`
                                    <${FilePill}
                                        prefix="compose"
                                        label=${label}
                                        title=${path}
                                        onClick=${() => onOpenFilePill?.(path)}
                                        removeTitle="Remove file"
                                        onRemove=${() => onRemoveFileRef?.(path)}
                                    />
                                `;
  })}
                            ${mediaFiles.map((file, index) => {
    const label = file?.name || `attachment-${index + 1}`;
    return X1`
                                    <${FilePill}
                                        key=${label + index}
                                        prefix="compose"
                                        label=${label}
                                        title=${label}
                                        removeTitle="Remove attachment"
                                        onRemove=${() => removeMediaFile(index)}
                                    />
                                `;
  })}
                            <button
                                type="button"
                                class="compose-clear-attachments-btn"
                                onClick=${clearAllAttachmentRefs}
                                title="Clear all attachments and references"
                                aria-label="Clear all attachments and references"
                            >
                                Clear all
                            </button>
                        </div>
                    `}
                    ${!searchMode && typeof onPopOutChat === "function" && X1`
                        <button
                            type="button"
                            class="compose-popout-btn"
                            onClick=${() => onPopOutChat?.()}
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
                        ref=${textareaRef}
                        placeholder=${searchMode ? "Search (Enter to run)..." : "Message (Enter to send, Shift+Enter for newline)..."}
                        value=${searchMode ? searchText : content}
                        onInput=${handleInput}
                        onKeyDown=${handleKeyDown}
                        onPaste=${handlePaste}
                        onFocus=${onFocus}
                        onClick=${onFocus}
                        rows="1"
                    />
                    ${showMention && mentionMatches.length > 0 && X1`
                        <div class="slash-autocomplete" ref=${mentionRef}>
                            ${mentionMatches.map((agent, i2) => X1`
                                <div
                                    key=${agent.chat_jid || agent.agent_name}
                                    class=${`slash-item${i2 === mentionIndex ? " active" : ""}`}
                                    onMouseDown=${(e2) => {
    e2.preventDefault();
    acceptMention(agent);
  }}
                                    onMouseEnter=${() => setMentionIndex(i2)}
                                >
                                    <span class="slash-name">@${agent.agent_name}</span>
                                    <span class="slash-desc">${agent.chat_jid || "Active agent"}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${showSlash && slashMatches.length > 0 && X1`
                        <div class="slash-autocomplete" ref=${slashRef}>
                            ${slashMatches.map((cmd, i2) => X1`
                                <div
                                    key=${cmd.name}
                                    class=${`slash-item${i2 === slashIndex ? " active" : ""}`}
                                    onMouseDown=${(e2) => {
    e2.preventDefault();
    acceptSlashCommand(cmd);
  }}
                                    onMouseEnter=${() => setSlashIndex(i2)}
                                >
                                    <span class="slash-name">${cmd.name}</span>
                                    <span class="slash-desc">${cmd.description}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${showModelPopup && !searchMode && X1`
                        <div class="compose-model-popup" ref=${modelPopupRef} tabIndex="-1" onKeyDown=${handlePopupKeyboardEvent}>
                            <div class="compose-model-popup-title">Select model</div>
                            <div class="compose-model-popup-menu" role="menu" aria-label="Model picker">
                                ${loadingModels && X1`
                                    <div class="compose-model-popup-empty">Loading models…</div>
                                `}
                                ${!loadingModels && modelOptions.length === 0 && X1`
                                    <div class="compose-model-popup-empty">No models available.</div>
                                `}
                                ${!loadingModels && modelOptions.map((modelOption, index) => {
    const modelLabel = typeof modelOption?.label === "string" ? modelOption.label : "";
    const contextWindowLabel = formatModelPickerContextWindow(modelOption?.contextWindow);
    return X1`
                                        <button
                                            key=${modelLabel}
                                            type="button"
                                            role="menuitem"
                                            class=${`compose-model-popup-item compose-model-popup-model-item${modelPopupIndex === index ? " active" : ""}${activeModel === modelLabel ? " current-model" : ""}`}
                                            onClick=${() => {
      handleSelectModel(modelOption);
    }}
                                            disabled=${switchingModel}
                                            title=${[modelLabel, contextWindowLabel].filter(Boolean).join(" • ")}
                                        >
                                            <span class="compose-model-popup-model-label">${formatModelPickerDisplayLabel(modelLabel, modelOption?.contextWindow)}</span>
                                        </button>
                                    `;
  })}
                            </div>
                            <div class="compose-model-popup-actions">
                                <button
                                    type="button"
                                    class="compose-model-popup-btn"
                                    onClick=${() => {
    handleCycleModel();
  }}
                                    disabled=${switchingModel}
                                >
                                    Next model
                                </button>
                            </div>
                        </div>
                    `}
                    ${showSessionPopup && !searchMode && X1`
                        <div class="compose-model-popup" ref=${sessionPopupRef} tabIndex="-1" onKeyDown=${handlePopupKeyboardEvent}>
                            <div class="compose-model-popup-title">Manage sessions & agents</div>
                            <div class="compose-model-popup-menu" role="menu" aria-label="Sessions and agents">
                                ${X1`
                                    <div class="compose-model-popup-item current" role="note" aria-live="polite">
                                        ${(() => {
    return formatCurrentBranchLabel(currentSessionAgent, currentChatJid);
  })()}
                                    </div>
                                `}
                                ${!hasSwitchableChatAgents && X1`
                                    <div class="compose-model-popup-empty">No other sessions yet.</div>
                                `}
                                ${hasSwitchableChatAgents && switchableChatAgents.map((chat, listIndex) => {
    const archived = Boolean(chat.archived_at);
    const isRoot = chat.chat_jid === (chat.root_chat_jid || chat.chat_jid);
    const canPrune = !isRoot && !chat.is_active && !archived && typeof onDeleteSession === "function";
    const label = formatBranchPickerLabel(chat, { currentChatJid });
    return X1`
                                        <div key=${chat.chat_jid} class=${`compose-model-popup-item-row${archived ? " archived" : ""}`}>
                                            <button
                                                type="button"
                                                role="menuitem"
                                                class=${`compose-model-popup-item${archived ? " archived" : ""}${sessionPopupIndex === listIndex ? " active" : ""}`}
                                                onClick=${() => {
      if (archived) {
        handleRestoreSession(chat.chat_jid);
        return;
      }
      handleSessionSwitch(chat.chat_jid);
    }}
                                                disabled=${archived ? !canRestoreSession : !canSwitchSession}
                                                title=${archived ? `Restore archived ${`@${chat.agent_name}`}` : `Switch to ${`@${chat.agent_name}`}`}
                                            >
                                                ${label}
                                            </button>
                                            ${canPrune && X1`
                                                <button
                                                    type="button"
                                                    class="compose-model-popup-item-delete"
                                                    title="Delete this branch"
                                                    aria-label=${`Delete @${chat.agent_name}`}
                                                    onClick=${(e2) => {
      e2.stopPropagation();
      setShowSessionPopup(false);
      onDeleteSession(chat.chat_jid);
    }}
                                                >
                                                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                                        <line x1="18" y1="6" x2="6" y2="18" />
                                                        <line x1="6" y1="6" x2="18" y2="18" />
                                                    </svg>
                                                </button>
                                            `}
                                        </div>
                                    `;
  })}
                            </div>
                            ${(canCreateSession || canRenameSession || canDeleteSession) && X1`
                                <div class="compose-model-popup-actions">
                                    ${canCreateSession && X1`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn primary${sessionPopupEntries.findIndex((entry) => entry.key === "action:new") === sessionPopupIndex ? " active" : ""}`}
                                            onClick=${() => {
    handleCreateSession();
  }}
                                            title="Create a new agent/session branch from this chat"
                                        >
                                            New
                                        </button>
                                    `}
                                    ${canRenameSession && X1`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn${sessionPopupEntries.findIndex((entry) => entry.key === "action:rename") === sessionPopupIndex ? " active" : ""}`}
                                            onClick=${(e2) => {
    handleRenameSession(e2);
  }}
                                            title="Rename the current branch handle"
                                            disabled=${renameInProgress}
                                        >
                                            Rename current…
                                        </button>
                                    `}
                                    ${canDeleteSession && X1`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn danger${sessionPopupEntries.findIndex((entry) => entry.key === "action:delete") === sessionPopupIndex ? " active" : ""}`}
                                            onClick=${() => {
    handleDeleteSession();
  }}
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
                <div class="compose-footer" ref=${footerRef}>
                    ${showComposeMetaRow && X1`
                    <div class="compose-meta-row">
                        ${showModelPickerHint && X1`
                            <div class="compose-model-meta">
                                <button
                                    ref=${modelHintRef}
                                    type="button"
                                    class="compose-model-hint compose-model-hint-btn"
                                    title=${modelHintTitle}
                                    aria-label="Open model picker"
                                    onClick=${toggleModelPopup}
                                    disabled=${switchingModel}
                                >
                                    ${switchingModel ? "Switching…" : modelHintLabel}
                                </button>
                                <div class="compose-model-meta-subline">
                                    ${!switchingModel && modelUsageSectionLabel && X1`
                                        <span class="compose-model-usage-hint" title=${modelHintTitle}>
                                            ${modelUsageSectionLabel}
                                        </span>
                                    `}
                                </div>
                            </div>
                        `}
                        ${!searchMode && contextUsage && contextUsage.percent != null && X1`
                            <${ContextPie} usage=${contextUsage} onCompact=${handleContextCompact} />
                        `}
                    </div>
                    `}
                    <div class="compose-actions ${searchMode ? "search-mode" : ""}">
                    ${showSessionSwitcherButton && X1`
                        <div
                            ref=${sessionTriggerRef}
                            class="compose-session-trigger-group"
                        >
                            ${currentSessionAgent?.agent_name && X1`
                                <button
                                    type="button"
                                    class=${`compose-session-trigger compose-session-trigger-pill${showSessionPopup ? " active" : ""}`}
                                    onClick=${toggleSessionPopup}
                                    title=${currentSessionAgent?.chat_jid || currentChatJid}
                                    aria-label=${`Manage sessions for @${currentSessionAgent.agent_name}`}
                                    aria-expanded=${showSessionPopup ? "true" : "false"}
                                >
                                    <span class="compose-current-agent-label active">@${currentSessionAgent.agent_name}</span>
                                </button>
                            `}
                            <button
                                type="button"
                                class=${`compose-session-trigger compose-session-trigger-icon-btn${showSessionPopup ? " active" : ""}`}
                                onClick=${toggleSessionPopup}
                                title=${currentSessionAgent?.chat_jid || currentChatJid}
                                aria-label=${currentSessionAgent?.agent_name ? `Manage sessions for @${currentSessionAgent.agent_name}` : "Manage Sessions/Agents"}
                                aria-expanded=${showSessionPopup ? "true" : "false"}
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
                    ${searchMode && X1`
                        <label class="compose-search-scope-wrap" title="Search scope">
                            <span class="compose-search-scope-label">Scope</span>
                            <select
                                class="compose-search-scope-select"
                                value=${searchScope}
                                onChange=${(e2) => onSearchScopeChange?.(e2.currentTarget.value)}
                            >
                                <option value="current">Current</option>
                                <option value="root">Branch family</option>
                                <option value="all">All chats</option>
                            </select>
                        </label>
                    `}
                    <button
                        class="icon-btn search-toggle"
                        onClick=${searchMode ? onExitSearch : onEnterSearch}
                        title=${searchMode ? "Close search" : "Search"}
                    >
                        ${searchMode ? X1`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 6L6 18M6 6l12 12"/>
                            </svg>
                        ` : X1`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="11" cy="11" r="8"/>
                                <path d="M21 21l-4.35-4.35"/>
                            </svg>
                        `}
                    </button>
                    ${canShareLocation && !searchMode && X1`
                        <button
                            class="icon-btn location-btn"
                            onClick=${handleLocation}
                            title="Share location"
                            type="button"
                            disabled=${false}
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="10" />
                                <path d="M12 2a14 14 0 0 1 0 20a14 14 0 0 1 0-20" />
                                <path d="M2 12h20" />
                            </svg>
                        </button>
                    `}
                    ${notificationsAvailable && !searchMode && X1`
                        <button
                            class=${`icon-btn notification-btn${notificationActive ? " active" : ""}`}
                            onClick=${onToggleNotifications}
                            title=${notificationTitle}
                            type="button"
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
                                <path d="M13.73 21a2 2 0 0 1-3.46 0" />
                            </svg>
                        </button>
                    `}
                    ${!searchMode && X1`
                        ${activeEditorPath && onAttachEditorFile && X1`
                            <button
                                class="icon-btn attach-editor-btn"
                                onClick=${onAttachEditorFile}
                                title=${`Attach open file: ${activeEditorPath}`}
                                type="button"
                                disabled=${fileRefs.includes(activeEditorPath)}
                            >
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z"/><polyline points="13 2 13 9 20 9"/></svg>
                            </button>
                        `}
                        <label class="icon-btn" title="Attach file">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                            <input type="file" multiple hidden onChange=${handleFileChange} />
                        </label>
                    `}
                    ${(connectionStatus !== "connected" || !searchMode) && X1`
                        <div class="compose-send-stack">
                            ${connectionStatus !== "connected" && X1`
                                <span class="compose-connection-status connection-status ${connectionStatusPresentation.statusClass}" title=${connectionStatusTitle}>
                                    ${connectionStatusLabel}
                                </span>
                            `}
                            ${!searchMode && X1`
                                <button 
                                    class=${submitButtonState.className}
                                    type="button"
                                    onClick=${() => {
    if (isComposeSubmitAbortMode(submitButtonState.mode)) {
      handleSubmit("/abort", "steer");
      return;
    }
    handleSubmit();
  }}
                                    disabled=${submitButtonState.disabled}
                                    title=${submitButtonState.title}
                                    aria-label=${submitButtonState.ariaLabel}
                                >
                                    ${submitButtonState.mode === "compacting" ? X1`
                                            <span class="compose-submit-spinner" aria-hidden="true">
                                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                                                    <circle class="compose-submit-spinner-ring" cx="12" cy="12" r="10.5" stroke-width="2.25" stroke-linecap="round"></circle>
                                                    <rect class="compose-submit-spinner-stop" x="6" y="6" width="12" height="12" rx="0" fill="currentColor"></rect>
                                                </svg>
                                            </span>
                                        ` : submitButtonState.mode === "abort" ? X1`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><rect x="4" y="4" width="16" height="16" rx="2.5"/></svg>` : X1`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>`}
                                </button>
                            `}
                        </div>
                    `}
                </div>
            </div>
        </div>
        </div>
    `;
}

// web/src/ui/tool-git-context.ts
function readTrimmedString(...values) {
  for (const value of values) {
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return null;
}
function stripOuterQuotes(value) {
  if (value.startsWith('"') && value.endsWith('"') || value.startsWith("'") && value.endsWith("'")) {
    return value.slice(1, -1);
  }
  return value;
}
function extractShellCwdFromCommand(command) {
  if (typeof command !== "string" || !command.trim())
    return null;
  const match = command.match(/^\s*cd\s+(.+?)(?:\s*(?:&&|;|\n))/s);
  if (!match?.[1])
    return null;
  const candidate = stripOuterQuotes(match[1].trim());
  return candidate || null;
}
function extractToolContextPath(toolName, args) {
  const record = args && typeof args === "object" ? args : null;
  if (!record)
    return null;
  const cwd = readTrimmedString(record.cwd, record.working_directory, record.workingDirectory);
  if (cwd)
    return cwd;
  const explicitRepoContext = readTrimmedString(record.project_dir, record.projectDir, record.repo_path, record.repoPath);
  if (explicitRepoContext)
    return explicitRepoContext;
  const command = readTrimmedString(record.command);
  const commandCwd = extractShellCwdFromCommand(command);
  if (commandCwd)
    return commandCwd;
  if (Array.isArray(record.commands)) {
    for (const entry of record.commands) {
      const nestedCwd = extractShellCwdFromCommand(entry);
      if (nestedCwd)
        return nestedCwd;
    }
  }
  return null;
}

// web/src/components/status.ts
var COPY_ICON_SVG2 = X1`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>
`;
var GIT_BRANCH_ICON_SVG = X1`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <path d="M6 3v12"></path>
        <circle cx="18" cy="6" r="3"></circle>
        <circle cx="6" cy="18" r="3"></circle>
        <path d="M18 9a9 9 0 0 1-9 9"></path>
    </svg>
`;
var CLOCK_ICON_SVG = X1`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <circle cx="12" cy="12" r="9"></circle>
        <path d="M12 7v5l3 2"></path>
    </svg>
`;
var STATUS_TIME_HINT_THRESHOLD_MS = 1e4;
function normalizeStatusHints(value) {
  const source = Array.isArray(value) ? value : value && Array.isArray(value.status_hints) ? value.status_hints : [];
  return source.filter((hint) => hint && typeof hint === "object").map((hint, index) => ({
    key: typeof hint.key === "string" && hint.key.trim() ? hint.key.trim() : `hint-${index}`,
    iconSvg: typeof hint.icon_svg === "string" ? hint.icon_svg.trim() : "",
    label: typeof hint.label === "string" ? hint.label.trim() : "",
    title: typeof hint.title === "string" ? hint.title.trim() : ""
  })).filter((hint) => hint.iconSvg && hint.label);
}
function resolveAgentStatusEscapeCollapseKey(expandedPanels) {
  if (!(expandedPanels instanceof Set) || expandedPanels.size === 0)
    return null;
  const keys = Array.from(expandedPanels.values());
  for (let index = keys.length - 1;index >= 0; index -= 1) {
    const key = keys[index];
    if (key === "thought" || key === "draft")
      return key;
  }
  return null;
}
function orderAgentStatusHints(statusHints) {
  if (!Array.isArray(statusHints) || statusHints.length === 0)
    return [];
  const priority = new Map([
    ["ssh", 0]
  ]);
  return statusHints.map((hint, index) => ({ hint, index })).sort((left, right) => {
    const leftPriority = priority.get(left.hint?.key) ?? 100;
    const rightPriority = priority.get(right.hint?.key) ?? 100;
    if (leftPriority !== rightPriority)
      return leftPriority - rightPriority;
    return left.index - right.index;
  }).map((entry) => entry.hint);
}
function formatAgentStatusGitLabel(repoPath, branch) {
  const normalizedRepoPath = typeof repoPath === "string" ? repoPath.trim() : "";
  const normalizedBranch = typeof branch === "string" ? branch.trim() : "";
  const repoName = normalizedRepoPath ? normalizedRepoPath.split(/[\\/]+/).filter(Boolean).pop() || normalizedRepoPath : "";
  return [repoName, normalizedBranch].filter(Boolean).join(" • ");
}
function shouldTickStatusActivityAge(status) {
  if (!status || typeof status !== "object")
    return false;
  const type = typeof status.type === "string" ? status.type : "";
  const isLastActivity = Boolean(status.last_activity || status.lastActivity);
  const isToolStatus = type === "tool_call" || type === "tool_status" || Boolean(status.tool_name || status.tool_args);
  if (!isLastActivity && !isToolStatus)
    return false;
  return parseStatusLastEventAt(status) !== null;
}
function shouldTickIntentElapsed(status) {
  if (!status || typeof status !== "object")
    return false;
  return status.type === "intent" && parseStatusStartedAt(status) !== null;
}
function hasMetStatusTimeHintThreshold(timestampMs, nowMs = Date.now()) {
  if (!Number.isFinite(timestampMs))
    return false;
  return nowMs - timestampMs >= STATUS_TIME_HINT_THRESHOLD_MS;
}
function resolveStatusActivityAgeLabel(status, nowMs = Date.now()) {
  if (!shouldTickStatusActivityAge(status))
    return null;
  const lastEventAtMs = parseStatusLastEventAt(status);
  if (lastEventAtMs === null || !hasMetStatusTimeHintThreshold(lastEventAtMs, nowMs))
    return null;
  const ageLabel = formatElapsed(new Date(lastEventAtMs).toISOString(), nowMs);
  return ageLabel ? `${ageLabel} ago` : null;
}
function resolveIntentElapsedLabel(status, nowMs = Date.now()) {
  if (!shouldTickIntentElapsed(status))
    return null;
  const startedAtMs = parseStatusStartedAt(status);
  if (startedAtMs === null || !hasMetStatusTimeHintThreshold(startedAtMs, nowMs))
    return null;
  return getStatusElapsedLabel(status, nowMs);
}
function resolveAgentStatusContent(status, options = {}) {
  const isLastActivity = options?.isLastActivity ?? Boolean(status?.last_activity || status?.lastActivity);
  const title = status?.title;
  const statusText = status?.status;
  let content = "";
  if (status?.type === "plan") {
    content = title ? `Planning: ${title}` : "Planning...";
  } else if (status?.type === "tool_call") {
    content = title ? `Running: ${title}` : "Running tool...";
  } else if (status?.type === "tool_status") {
    content = title ? `${title}: ${statusText || "Working..."}` : statusText || "Working...";
  } else if (status?.type === "error") {
    content = title || "Agent error";
  } else {
    content = title || statusText || "Working...";
  }
  if (!isLastActivity)
    return content;
  if (content && content !== "Working...") {
    return `Recent activity: ${content}`;
  }
  return "Last activity";
}
function formatElapsed(isoString, nowMs = Date.now()) {
  if (!isoString)
    return null;
  const ms = nowMs - new Date(isoString).getTime();
  if (!Number.isFinite(ms) || ms < 0)
    return null;
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m2 = Math.floor(totalSec % 3600 / 60);
  const s2 = totalSec % 60;
  if (h > 0)
    return `${h}h ${m2}m`;
  if (m2 > 0)
    return `${m2}m ${s2}s`;
  return `${s2}s`;
}
function AgentStatus({ status, draft, plan, thought, pendingRequest, intent, extensionPanels = [], pendingPanelActions = new Set, onExtensionPanelAction, turnId, steerQueued, onPanelToggle, showCorePanels = true, showExtensionPanels = true }) {
  const THOUGHT_MAX_LINES = 8;
  const DRAFT_MAX_LINES = 8;
  const normalizePreview = (value) => {
    if (!value)
      return { text: "", totalLines: 0, fullText: "" };
    if (typeof value === "string") {
      const text2 = value;
      const totalLines2 = text2 ? text2.replace(/\r\n/g, `
`).split(`
`).length : 0;
      return { text: text2, totalLines: totalLines2, fullText: text2 };
    }
    const text = value.text || "";
    const fullText = value.fullText || value.full_text || text;
    const totalLines = Number.isFinite(value.totalLines) ? value.totalLines : fullText ? fullText.replace(/\r\n/g, `
`).split(`
`).length : 0;
    return { text, totalLines, fullText };
  };
  const PREVIEW_MAX_CHARS_PER_LINE = 160;
  const stripInternalTags = (value) => String(value || "").replace(/<\/?internal>/gi, "");
  const countSoftLines = (line) => {
    if (!line)
      return 1;
    return Math.max(1, Math.ceil(line.length / PREVIEW_MAX_CHARS_PER_LINE));
  };
  const truncateLines = (text, maxLines, totalLinesOverride) => {
    const value = (text || "").replace(/\r\n/g, `
`).replace(/\r/g, `
`);
    if (!value) {
      const totalLines2 = Number.isFinite(totalLinesOverride) ? totalLinesOverride : 0;
      return { text: "", omitted: 0, totalLines: totalLines2, visibleLines: 0 };
    }
    const lines = value.split(`
`);
    const clipped = lines.length > maxLines ? lines.slice(0, maxLines).join(`
`) : value;
    const totalLines = Number.isFinite(totalLinesOverride) ? totalLinesOverride : lines.reduce((acc, line) => acc + countSoftLines(line), 0);
    const visibleLines = clipped ? clipped.split(`
`).reduce((acc, line) => acc + countSoftLines(line), 0) : 0;
    const omitted = Math.max(totalLines - visibleLines, 0);
    return { text: clipped, omitted, totalLines, visibleLines };
  };
  const planInfo = normalizePreview(plan);
  const thoughtInfo = normalizePreview(thought);
  const draftInfo = normalizePreview(draft);
  const hasPlan = Boolean(planInfo.text) || planInfo.totalLines > 0;
  const hasThought = Boolean(thoughtInfo.text) || thoughtInfo.totalLines > 0;
  const hasDraft = Boolean(draftInfo.fullText?.trim() || draftInfo.text?.trim());
  const hasCorePanels = Boolean(status || hasDraft || hasPlan || hasThought || pendingRequest || intent);
  const hasExtensionPanels = Array.isArray(extensionPanels) && extensionPanels.length > 0;
  const [expandedPanels, setExpandedPanels] = w0(new Set);
  const [hoveredSeriesPoint, setHoveredSeriesPoint] = w0(null);
  const [nowMs, setNowMs] = w0(() => Date.now());
  const toggleExpand = (key) => setExpandedPanels((prev) => {
    const next = new Set(prev);
    const willExpand = !next.has(key);
    if (willExpand)
      next.add(key);
    else
      next.delete(key);
    if (typeof onPanelToggle === "function") {
      onPanelToggle(key, willExpand);
    }
    return next;
  });
  r0(() => {
    setExpandedPanels(new Set);
    setHoveredSeriesPoint(null);
  }, [turnId]);
  r0(() => {
    const hasExpandedTimestampPanel = Array.isArray(extensionPanels) && extensionPanels.some((p2) => expandedPanels.has(p2?.key) && (p2?.started_at || p2?.last_activity_at));
    if (!hasExpandedTimestampPanel)
      return;
    const interval = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(interval);
  }, [expandedPanels, extensionPanels]);
  const escapeCollapseKey = G0(() => resolveAgentStatusEscapeCollapseKey(expandedPanels), [expandedPanels]);
  r0(() => {
    if (!escapeCollapseKey || typeof document === "undefined")
      return;
    const handleKeyDown = (event) => {
      if (event?.defaultPrevented)
        return;
      if (event?.key !== "Escape")
        return;
      if (event?.altKey || event?.ctrlKey || event?.metaKey || event?.shiftKey)
        return;
      const target = event?.target;
      if (target instanceof Element) {
        if (target.closest?.('input, textarea, select, [contenteditable="true"]'))
          return;
        if (target.isContentEditable)
          return;
      }
      setExpandedPanels((prev) => {
        if (!(prev instanceof Set) || !prev.has(escapeCollapseKey))
          return prev;
        const next = new Set(prev);
        next.delete(escapeCollapseKey);
        return next;
      });
      if (typeof onPanelToggle === "function") {
        onPanelToggle(escapeCollapseKey, false);
      }
      event.preventDefault?.();
      event.stopPropagation?.();
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [escapeCollapseKey, onPanelToggle]);
  const statusIsCompaction = isCompactionStatus(status);
  const isLastActivity = Boolean(status?.last_activity || status?.lastActivity);
  const shouldTickActivityAge = G0(() => shouldTickStatusActivityAge(status), [status]);
  const shouldTickIntentAge = G0(() => shouldTickIntentElapsed(status), [status]);
  const toolContextPath = G0(() => extractToolContextPath(status?.tool_name, status?.tool_args), [status?.tool_name, status?.tool_args]);
  const [toolRepoContext, setToolRepoContext] = w0(null);
  r0(() => {
    const shouldTick = Boolean(shouldTickIntentAge || status?.retry_at || status?.retryAt || shouldTickActivityAge);
    if (!shouldTick)
      return;
    setNowMs(Date.now());
    const timer = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [shouldTickActivityAge, shouldTickIntentAge, status?.retry_at, status?.retryAt, status?.last_event_at, status?.lastEventAt, status?.started_at, status?.startedAt, status?.type, status?.tool_name, status?.tool_args]);
  r0(() => {
    const isToolStatus = status?.type === "tool_call" || status?.type === "tool_status";
    if (!isToolStatus || !toolContextPath) {
      setToolRepoContext(null);
      return;
    }
    let active = true;
    getWorkspaceBranch(toolContextPath).then((payload) => {
      if (!active)
        return;
      if (payload?.branch) {
        setToolRepoContext({
          branch: payload.branch,
          repoPath: payload.repo_path || null,
          path: toolContextPath
        });
      } else {
        setToolRepoContext(null);
      }
    }).catch(() => {
      if (active)
        setToolRepoContext(null);
    });
    return () => {
      active = false;
    };
  }, [status?.type, toolContextPath]);
  const activeTurn = status?.turn_id || turnId;
  const turnColor = getTurnColor(activeTurn);
  const dotClass = buildTurnDotClass({ steerQueued });
  const panelTitle = (label) => label;
  const showRunningStatusDot = shouldShowRunningStatusDot(status, { isLastActivity });
  const runningIndicatorMode = resolveRunningStatusIndicator(status, { isLastActivity });
  const pendingIndicatorMode = resolveRunningStatusIndicator(null, { pendingRequest: true });
  const resolveIntentColor = (kind) => kind === "warning" ? "#f59e0b" : kind === "error" ? "var(--danger-color)" : kind === "success" ? "var(--success-color)" : turnColor;
  const intentKind = intent?.kind || "info";
  const intentColor = resolveIntentColor(intentKind);
  const statusIntentColor = resolveIntentColor(status?.kind || (statusIsCompaction ? "warning" : "info"));
  const content = resolveAgentStatusContent(status, { isLastActivity });
  const statusActivityAgeLabel = resolveStatusActivityAgeLabel(status, nowMs);
  const toolRepoRepoPath = toolRepoContext?.repoPath || "";
  const toolRepoBranch = toolRepoContext?.branch || "";
  const toolRepoLabel = toolRepoContext ? formatAgentStatusGitLabel(toolRepoRepoPath, toolRepoBranch) : "";
  const statusHints = normalizeStatusHints(status?.status_hints || status?.statusHints);
  const orderedStatusHints = G0(() => orderAgentStatusHints(statusHints), [statusHints]);
  const leadingStatusHints = G0(() => orderedStatusHints.filter((hint) => hint?.key === "ssh"), [orderedStatusHints]);
  const trailingStatusHints = G0(() => orderedStatusHints.filter((hint) => hint?.key !== "ssh"), [orderedStatusHints]);
  if ((!showCorePanels || !hasCorePanels) && (!showExtensionPanels || !hasExtensionPanels))
    return null;
  const renderThinkingPanel = ({ panelTitle: panelTitle2, text, fullText, totalLines, maxLines, titleClass, panelKey }) => {
    const isExpanded = expandedPanels.has(panelKey);
    const rawSourceText = fullText || text || "";
    const sourceText = panelKey === "thought" || panelKey === "draft" ? stripInternalTags(rawSourceText) : rawSourceText;
    const isCollapsible = typeof maxLines === "number";
    const showClose = isExpanded && isCollapsible;
    const truncated = isCollapsible ? truncateLines(sourceText, maxLines, totalLines) : { text: sourceText || "", omitted: 0, totalLines: Number.isFinite(totalLines) ? totalLines : 0 };
    if (!sourceText && !(Number.isFinite(truncated.totalLines) && truncated.totalLines > 0))
      return null;
    const bodyClass = `agent-thinking-body${isCollapsible ? " agent-thinking-body-collapsible" : ""}`;
    const bodyStyle = isCollapsible ? `--agent-thinking-collapsed-lines: ${maxLines};` : "";
    return X1`
            <div
                class="agent-thinking"
                data-expanded=${isExpanded ? "true" : "false"}
                data-collapsible=${isCollapsible ? "true" : "false"}
                style=${turnColor ? `--turn-color: ${turnColor};` : ""}
            >
                <div class="agent-thinking-title ${titleClass || ""}">
                    ${turnColor && X1`<span class=${dotClass} aria-hidden="true"></span>`}
                    ${panelTitle2}
                    ${showClose && X1`
                        <button
                            class="agent-thinking-close"
                            aria-label=${`Close ${panelTitle2} panel`}
                            onClick=${() => toggleExpand(panelKey)}
                        >
                            ×
                        </button>
                    `}
                </div>
                <div
                    class=${bodyClass}
                    style=${bodyStyle}
                    dangerouslySetInnerHTML=${{ __html: renderThinkingMarkdown(sourceText) }}
                />
                ${!isExpanded && truncated.omitted > 0 && X1`
                    <button class="agent-thinking-truncation" onClick=${() => toggleExpand(panelKey)}>
                        ▸ ${truncated.omitted} more lines
                    </button>
                `}
                ${isExpanded && truncated.omitted > 0 && X1`
                    <button class="agent-thinking-truncation" onClick=${() => toggleExpand(panelKey)}>
                        ▴ show less
                    </button>
                `}
            </div>
        `;
  };
  const pendingTitle = pendingRequest?.tool_call?.title;
  const pendingMessage = pendingTitle ? `Awaiting approval: ${pendingTitle}` : "Awaiting approval";
  const statusIntentElapsedLabel = resolveIntentElapsedLabel(status, nowMs);
  const renderIntentPanel = (payload, color, elapsedLabel = null) => {
    const titleText = resolveStatusPanelTitle(payload);
    const retryCountdownLabel = getStatusRetryCountdownLabel(payload, nowMs);
    const metaLabel = [elapsedLabel, retryCountdownLabel].filter(Boolean).join(" · ");
    const pulsingDotClass = buildTurnDotClass({
      steerQueued,
      pulsing: isCompactionStatus(payload) || Boolean(retryCountdownLabel)
    });
    return X1`
            <div
                class="agent-thinking agent-thinking-intent"
                aria-live="polite"
                style=${color ? `--turn-color: ${color};` : ""}
                title=${payload?.detail || ""}
            >
                <div class="agent-thinking-title intent">
                    ${color && X1`<span class=${pulsingDotClass} aria-hidden="true"></span>`}
                    <span class="agent-thinking-title-text">${titleText}</span>
                    ${metaLabel && X1`<span class="agent-status-elapsed">${metaLabel}</span>`}
                </div>
                ${payload.detail && X1`<div class="agent-thinking-body">${payload.detail}</div>`}
            </div>
        `;
  };
  const projectSeriesPoint = (point, width, height, minValue, maxValue, minRun, maxRun, paddingX = 8, paddingY = 8) => {
    const range = Math.max(maxValue - minValue, 0.000000001);
    const innerWidth = Math.max(width - paddingX * 2, 1);
    const innerHeight = Math.max(height - paddingY * 2, 1);
    const runSpan = Math.max(maxRun - minRun, 1);
    const x2 = maxRun === minRun ? width / 2 : paddingX + (point.run - minRun) / runSpan * innerWidth;
    const y2 = paddingY + (innerHeight - (point.value - minValue) / range * innerHeight);
    return { x: x2, y: y2 };
  };
  const buildLinePath = (points, width, height, minValue, maxValue, minRun, maxRun, paddingX = 8, paddingY = 8) => {
    if (!Array.isArray(points) || points.length === 0)
      return "";
    return points.map((point, index) => {
      const { x: x2, y: y2 } = projectSeriesPoint(point, width, height, minValue, maxValue, minRun, maxRun, paddingX, paddingY);
      return `${index === 0 ? "M" : "L"} ${x2.toFixed(2)} ${y2.toFixed(2)}`;
    }).join(" ");
  };
  const formatMetricValue = (value, unit = "") => {
    if (!Number.isFinite(value))
      return "—";
    const rounded = Math.abs(value) >= 100 ? value.toFixed(0) : value.toFixed(2).replace(/\.0+$/, "").replace(/(\.\d*[1-9])0+$/, "$1");
    return `${rounded}${unit}`;
  };
  const SERIES_COLOR_ANCHORS = [
    "var(--accent-color)",
    "var(--success-color)",
    "var(--warning-color, #f59e0b)",
    "var(--danger-color)"
  ];
  const resolveSeriesColor = (index, total) => {
    const anchors = SERIES_COLOR_ANCHORS;
    if (!Array.isArray(anchors) || anchors.length === 0)
      return "var(--accent-color)";
    if (anchors.length === 1 || !Number.isFinite(total) || total <= 1)
      return anchors[0];
    const clampedIndex = Math.max(0, Math.min(Number.isFinite(index) ? index : 0, total - 1));
    const scaled = clampedIndex / Math.max(1, total - 1) * (anchors.length - 1);
    const leftIndex = Math.floor(scaled);
    const rightIndex = Math.min(anchors.length - 1, leftIndex + 1);
    const mixRatio = scaled - leftIndex;
    const left = anchors[leftIndex];
    const right = anchors[rightIndex];
    if (!right || leftIndex === rightIndex || mixRatio <= 0.001)
      return left;
    if (mixRatio >= 0.999)
      return right;
    const leftWeight = Math.round((1 - mixRatio) * 1000) / 10;
    const rightWeight = Math.round(mixRatio * 1000) / 10;
    return `color-mix(in oklab, ${left} ${leftWeight}%, ${right} ${rightWeight}%)`;
  };
  const renderCombinedSeriesChart = (seriesList, panelKey = "autoresearch") => {
    const prepared = Array.isArray(seriesList) ? seriesList.map((series) => ({
      ...series,
      points: Array.isArray(series?.points) ? series.points.filter((point) => Number.isFinite(point?.value) && Number.isFinite(point?.run)) : []
    })).filter((series) => series.points.length > 0) : [];
    const normalized = prepared.map((series, index) => ({
      ...series,
      color: resolveSeriesColor(index, prepared.length)
    }));
    if (normalized.length === 0)
      return null;
    const width = 320;
    const height = 120;
    const allPoints = normalized.flatMap((series) => series.points);
    const values = allPoints.map((point) => point.value);
    const runs = allPoints.map((point) => point.run);
    const minValue = Math.min(...values);
    const maxValue = Math.max(...values);
    const minRun = Math.min(...runs);
    const maxRun = Math.max(...runs);
    return X1`
            <div class="agent-series-chart agent-series-chart-combined">
                <div class="agent-series-chart-header">
                    <span class="agent-series-chart-title">Tracked variables</span>
                    <span class="agent-series-chart-value">${normalized.length} series</span>
                </div>
                <div class="agent-series-chart-plot">
                    <svg class="agent-series-chart-svg" viewBox=${`0 0 ${width} ${height}`} preserveAspectRatio="none" aria-hidden="true">
                        ${normalized.map((series) => {
      const seriesKey = series?.key || series?.label || "series";
      const lineHovered = hoveredSeriesPoint?.panelKey === panelKey && hoveredSeriesPoint?.seriesKey === seriesKey;
      return X1`
                                <g key=${seriesKey}>
                                    <path
                                        class=${`agent-series-chart-line${lineHovered ? " is-hovered" : ""}`}
                                        d=${buildLinePath(series.points, width, height, minValue, maxValue, minRun, maxRun)}
                                        style=${`--agent-series-color: ${series.color};`}
                                        onMouseEnter=${() => setHoveredSeriesPoint({ panelKey, seriesKey })}
                                        onMouseLeave=${() => setHoveredSeriesPoint((prev) => prev?.panelKey === panelKey && prev?.seriesKey === seriesKey ? null : prev)}
                                    ></path>
                                </g>
                            `;
    })}
                    </svg>
                    <div class="agent-series-chart-points-layer">
                        ${normalized.flatMap((series) => {
      const unit = typeof series?.unit === "string" ? series.unit : "";
      const seriesKey = series?.key || series?.label || "series";
      return series.points.map((point, pointIndex) => {
        const projected = projectSeriesPoint(point, width, height, minValue, maxValue, minRun, maxRun);
        return X1`
                                    <button
                                        key=${`${seriesKey}-point-${pointIndex}`}
                                        type="button"
                                        class="agent-series-chart-point-hit"
                                        style=${`--agent-series-color: ${series.color}; left:${projected.x / width * 100}%; top:${projected.y / height * 100}%;`}
                                        onMouseEnter=${() => setHoveredSeriesPoint({
          panelKey,
          seriesKey,
          run: point.run,
          value: point.value,
          unit
        })}
                                        onMouseLeave=${() => setHoveredSeriesPoint((prev) => prev?.panelKey === panelKey ? null : prev)}
                                        onFocus=${() => setHoveredSeriesPoint({
          panelKey,
          seriesKey,
          run: point.run,
          value: point.value,
          unit
        })}
                                        onBlur=${() => setHoveredSeriesPoint((prev) => prev?.panelKey === panelKey ? null : prev)}
                                        aria-label=${`${series?.label || "Series"} ${formatMetricValue(point.value, unit)} at run ${point.run}`}
                                    >
                                        <span class="agent-series-chart-point"></span>
                                    </button>
                                `;
      });
    })}
                    </div>
                </div>
                <div class="agent-series-legend">
                    ${normalized.map((series) => {
      const latest = series.points[series.points.length - 1]?.value;
      const unit = typeof series?.unit === "string" ? series.unit : "";
      const seriesKey = series?.key || series?.label || "series";
      const hovered = hoveredSeriesPoint?.panelKey === panelKey && hoveredSeriesPoint?.seriesKey === seriesKey ? hoveredSeriesPoint : null;
      const hoveredValue = hovered && Number.isFinite(hovered.value) ? hovered.value : latest;
      const hoveredUnit = hovered && typeof hovered.unit === "string" ? hovered.unit : unit;
      const hoveredRun = hovered && Number.isFinite(hovered.run) ? hovered.run : null;
      return X1`
                            <div key=${`${seriesKey}-legend`} class=${`agent-series-legend-item${hovered ? " is-hovered" : ""}`} style=${`--agent-series-color: ${series.color};`}>
                                <span class="agent-series-legend-swatch" style=${`--agent-series-color: ${series.color};`}></span>
                                <span class="agent-series-legend-label">${series?.label || "Series"}</span>
                                ${hoveredRun !== null && X1`<span class="agent-series-legend-run">run ${hoveredRun}</span>`}
                                <span class="agent-series-legend-value">${formatMetricValue(hoveredValue, hoveredUnit)}</span>
                            </div>
                        `;
    })}
                </div>
            </div>
        `;
  };
  const renderExtensionPanel = (panel) => {
    if (!panel)
      return null;
    const panelKey = typeof panel?.key === "string" ? panel.key : `panel-${Math.random()}`;
    const isExpanded = expandedPanels.has(panelKey);
    const titleText = panel?.title || "Extension status";
    const collapsedText = panel?.collapsed_text || "";
    const stateLabel = String(panel?.state || "").replace(/[-_]+/g, " ").replace(/^./, (match) => match.toUpperCase());
    const color = resolveIntentColor(panel?.state === "completed" ? "success" : panel?.state === "failed" ? "error" : panel?.state === "stopped" ? "warning" : "info");
    const panelDotClass = buildTurnDotClass({
      steerQueued,
      pulsing: panel?.state === "running"
    });
    const detailText = typeof panel?.detail_markdown === "string" ? panel.detail_markdown.trim() : "";
    const lastRunText = typeof panel?.last_run_text === "string" ? panel.last_run_text.trim() : "";
    const tmuxCommand = typeof panel?.tmux_command === "string" ? panel.tmux_command.trim() : "";
    const series = Array.isArray(panel?.series) ? panel.series : [];
    const actions = Array.isArray(panel?.actions) ? panel.actions : [];
    const experimentElapsed = formatElapsed(panel?.started_at);
    const elapsedSuffix = experimentElapsed ? ` · ${experimentElapsed}` : "";
    const displayCollapsed = collapsedText + elapsedSuffix;
    const hasDetailColumn = Boolean(detailText || tmuxCommand || experimentElapsed);
    const isExpandable = Boolean(detailText || series.length > 0 || tmuxCommand);
    const collapsedTooltip = [titleText, displayCollapsed].filter(Boolean).join(" — ");
    return X1`
            <div
                class="agent-thinking agent-thinking-intent agent-thinking-autoresearch"
                aria-live="polite"
                data-expanded=${isExpanded ? "true" : "false"}
                style=${color ? `--turn-color: ${color};` : ""}
                title=${!isExpanded ? collapsedTooltip || titleText : ""}
            >
                <div class="agent-thinking-header agent-thinking-header-inline">
                    <button
                        class="agent-thinking-title intent agent-thinking-title-clickable"
                        type="button"
                        onClick=${() => isExpandable ? toggleExpand(panelKey) : null}
                    >
                        ${color && X1`<span class=${panelDotClass} aria-hidden="true"></span>`}
                        <span class="agent-thinking-title-text">${titleText}</span>
                        ${displayCollapsed && X1`<span class="agent-thinking-title-meta">${displayCollapsed}</span>`}
                    </button>
                    ${(actions.length > 0 || isExpandable) && X1`
                        <div class="agent-thinking-tools-inline">
                            ${actions.length > 0 && X1`
                                <div class="agent-thinking-actions agent-thinking-actions-inline">
                                    ${actions.map((action) => {
      const pendingKey = `${panelKey}:${action?.key || ""}`;
      const pending = pendingPanelActions?.has?.(pendingKey);
      return X1`
                                            <button
                                                key=${pendingKey}
                                                class=${`agent-thinking-action-btn${action?.tone === "danger" ? " danger" : ""}`}
                                                onClick=${() => onExtensionPanelAction?.(panel, action)}
                                                disabled=${Boolean(pending)}
                                            >
                                                ${pending ? "Working…" : action?.label || "Run"}
                                            </button>
                                        `;
    })}
                                </div>
                            `}
                            ${isExpandable && X1`
                                <button
                                    class="agent-thinking-corner-toggle agent-thinking-corner-toggle-inline"
                                    type="button"
                                    aria-label=${`${isExpanded ? "Collapse" : "Expand"} ${titleText}`}
                                    title=${isExpanded ? "Collapse details" : "Expand details"}
                                    onClick=${() => toggleExpand(panelKey)}
                                >
                                    <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        ${isExpanded ? X1`<polyline points="4 6 8 10 12 6"></polyline>` : X1`<polyline points="4 10 8 6 12 10"></polyline>`}
                                    </svg>
                                </button>
                            `}
                        </div>
                    `}
                </div>
                ${isExpanded && X1`
                    <div class=${`agent-thinking-autoresearch-layout${hasDetailColumn ? "" : " chart-only"}`}>
                        ${hasDetailColumn && X1`
                            <div class="agent-thinking-autoresearch-meta-stack">
                                ${experimentElapsed && X1`
                                    <div class="agent-thinking-autoresearch-elapsed">
                                        <span title="Experiment duration">⏱ ${experimentElapsed}</span>
                                        ${panel?.last_activity_at && panel?.state === "running" && X1`<span title="Since last activity">⟳ ${formatElapsed(panel.last_activity_at)} ago</span>`}
                                    </div>
                                `}
                                ${detailText && X1`
                                    <div
                                        class="agent-thinking-body agent-thinking-autoresearch-detail"
                                        dangerouslySetInnerHTML=${{ __html: renderThinkingMarkdown(detailText) }}
                                    />
                                `}
                                ${tmuxCommand && X1`
                                    <div class="agent-series-chart-command">
                                        <div class="agent-series-chart-command-header">
                                            <span>Attach to session</span>
                                        </div>
                                        <div class="agent-series-chart-command-shell">
                                            <pre class="agent-series-chart-command-code">${tmuxCommand}</pre>
                                            <button
                                                type="button"
                                                class="agent-series-chart-command-copy"
                                                aria-label="Copy tmux command"
                                                title="Copy tmux command"
                                                onClick=${() => onExtensionPanelAction?.(panel, { key: "copy_tmux", action_type: "autoresearch.copy_tmux", label: "Copy tmux" })}
                                            >
                                                ${COPY_ICON_SVG2}
                                            </button>
                                        </div>
                                    </div>
                                `}
                            </div>
                        `}
                        ${series.length > 0 ? X1`
                                <div class="agent-series-chart-stack">
                                    ${renderCombinedSeriesChart(series, panelKey)}
                                    ${lastRunText && X1`<div class="agent-series-chart-note">${lastRunText}</div>`}
                                </div>
                            ` : X1`<div class="agent-thinking-body agent-thinking-autoresearch-summary">Variable history will appear after the first completed run.</div>`}
                    </div>
                `}
            </div>
        `;
  };
  return X1`
        <div class="agent-status-panel">
            ${showCorePanels && intent && renderIntentPanel(intent, intentColor)}
            ${showExtensionPanels && Array.isArray(extensionPanels) && extensionPanels.map((panel) => renderExtensionPanel(panel))}
            ${showCorePanels && status?.type === "intent" && renderIntentPanel(status, statusIntentColor, statusIntentElapsedLabel)}
            ${showCorePanels && pendingRequest && X1`
                <div class="agent-status agent-status-request" aria-live="polite" style=${turnColor ? `--turn-color: ${turnColor};` : ""}>
                    ${pendingIndicatorMode === "dot" && X1`<span class=${dotClass} aria-hidden="true"></span>`}
                    ${pendingIndicatorMode === "spinner" && X1`<div class="agent-status-spinner"></div>`}
                    <span class="agent-status-text">${pendingMessage}</span>
                </div>
            `}
            ${showCorePanels && hasPlan && renderThinkingPanel({
    panelTitle: panelTitle("Planning"),
    text: planInfo.text,
    fullText: planInfo.fullText,
    totalLines: planInfo.totalLines,
    panelKey: "plan"
  })}
            ${showCorePanels && hasThought && renderThinkingPanel({
    panelTitle: panelTitle("Thoughts"),
    text: thoughtInfo.text,
    fullText: thoughtInfo.fullText,
    totalLines: thoughtInfo.totalLines,
    maxLines: THOUGHT_MAX_LINES,
    titleClass: "thought",
    panelKey: "thought"
  })}
            ${showCorePanels && hasDraft && renderThinkingPanel({
    panelTitle: panelTitle("Draft"),
    text: draftInfo.text,
    fullText: draftInfo.fullText,
    totalLines: draftInfo.totalLines,
    maxLines: DRAFT_MAX_LINES,
    titleClass: "thought",
    panelKey: "draft"
  })}
            ${showCorePanels && status && status?.type !== "intent" && X1`
                <div class=${`agent-status${isLastActivity ? " agent-status-last-activity" : ""}${status?.type === "error" ? " agent-status-error" : ""}${toolRepoLabel || statusHints.length > 0 || statusActivityAgeLabel ? " agent-status-multiline" : ""}`} aria-live="polite" style=${turnColor ? `--turn-color: ${turnColor};` : ""}>
                    ${turnColor && showRunningStatusDot && X1`<span class=${dotClass} aria-hidden="true"></span>`}
                    ${status?.type === "error" ? X1`<span class="agent-status-error-icon" aria-hidden="true">⚠</span>` : runningIndicatorMode === "spinner" && X1`<div class="agent-status-spinner"></div>`}
                    <div class="agent-status-copy">
                        <span class="agent-status-text">${content}</span>
                        ${(toolRepoLabel || orderedStatusHints.length > 0 || statusActivityAgeLabel) && X1`
                            <span class="agent-status-meta-row">
                                ${leadingStatusHints.map((hint) => X1`
                                    <span key=${hint.key} class="agent-status-hint-row" title=${hint.title || hint.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{ __html: hint.iconSvg }}></span>
                                        <span class="agent-status-hint-label">${hint.label}</span>
                                    </span>
                                `)}
                                ${toolRepoLabel && X1`
                                    <span class="agent-status-git-row" title=${toolContextPath || toolRepoLabel}>
                                        <span class="agent-status-git-icon">${GIT_BRANCH_ICON_SVG}</span>
                                        <span class="agent-status-git-label">
                                            ${toolRepoRepoPath && X1`<span class="agent-status-git-part">${toolRepoRepoPath}</span>`}
                                            ${toolRepoRepoPath && toolRepoBranch && X1`<span class="agent-status-git-separator" aria-hidden="true">•</span>`}
                                            ${toolRepoBranch && X1`<span class="agent-status-git-part">${toolRepoBranch}</span>`}
                                        </span>
                                    </span>
                                `}
                                ${trailingStatusHints.map((hint) => X1`
                                    <span key=${hint.key} class="agent-status-hint-row" title=${hint.title || hint.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{ __html: hint.iconSvg }}></span>
                                        <span class="agent-status-hint-label">${hint.label}</span>
                                    </span>
                                `)}
                                ${statusActivityAgeLabel && X1`
                                    <span class="agent-status-hint-row agent-status-activity-row" title=${`${isLastActivity ? "Recent activity" : "Last event"} ${statusActivityAgeLabel}`}>
                                        <span class="agent-status-hint-icon">${CLOCK_ICON_SVG}</span>
                                        <span class="agent-status-hint-label">${statusActivityAgeLabel}</span>
                                    </span>
                                `}
                            </span>
                        `}
                    </div>
                </div>
            `}
        </div>
    `;
}

// web/src/panes/pane-registry.ts
class PaneRegistryImpl {
  extensions = new Map;
  register(ext) {
    this.extensions.set(ext.id, ext);
  }
  unregister(id) {
    this.extensions.delete(id);
  }
  resolve(context) {
    let best;
    let bestPriority = -Infinity;
    for (const ext of this.extensions.values()) {
      if (ext.placement !== "tabs")
        continue;
      if (!ext.canHandle)
        continue;
      try {
        const result = ext.canHandle(context);
        if (result === false || result === 0)
          continue;
        const priority = result === true ? 0 : typeof result === "number" ? result : 0;
        if (priority > bestPriority) {
          bestPriority = priority;
          best = ext;
        }
      } catch (err) {
        console.warn(`[PaneRegistry] canHandle() error for "${ext.id}":`, err);
      }
    }
    return best;
  }
  list() {
    return Array.from(this.extensions.values());
  }
  getDockPanes() {
    return Array.from(this.extensions.values()).filter((ext) => ext.placement === "dock");
  }
  getTabPanes() {
    return Array.from(this.extensions.values()).filter((ext) => ext.placement === "tabs");
  }
  get(id) {
    return this.extensions.get(id);
  }
  get size() {
    return this.extensions.size;
  }
}
var paneRegistry = new PaneRegistryImpl;
// web/src/panes/pane-runtime-safety.ts
function readRandomUuidBestEffort(runtime) {
  try {
    return typeof runtime?.crypto?.randomUUID === "function" ? runtime.crypto.randomUUID() : null;
  } catch (_error) {
    return null;
  }
}
function removeStorageItemBestEffort(storage, key) {
  try {
    storage?.removeItem?.(key);
    return true;
  } catch (_error) {
    return false;
  }
}

// web/src/panes/editor-popout-transfer.ts
var EDITOR_POPOUT_STATE_TTL_MS = 5 * 60 * 1000;
function consumePanePopoutTransferToken(paramName, runtime = globalThis) {
  const win = runtime?.window ?? runtime;
  if (!win?.location?.href)
    return null;
  try {
    const url = new URL(win.location.href);
    const token = url.searchParams.get(paramName)?.trim() || "";
    if (!token)
      return null;
    url.searchParams.delete(paramName);
    win.history?.replaceState?.(win.history.state, win.document?.title || "", url.toString());
    return token;
  } catch {
    return null;
  }
}

// web/src/panes/terminal-lifecycle-runtime.ts
function runBestEffort(run) {
  try {
    run();
    return true;
  } catch (_error) {
    return false;
  }
}
function detachTerminalHostListenersBestEffort(options) {
  const {
    ownerWindow,
    themeChangeListener,
    mediaQuery,
    mediaQueryListener,
    dockResizeListener,
    windowResizeListener,
    themeObserver,
    resizeObserver
  } = options;
  runBestEffort(() => {
    if (themeChangeListener) {
      ownerWindow?.removeEventListener?.("piclaw-theme-change", themeChangeListener);
    }
  });
  runBestEffort(() => {
    if (mediaQuery && mediaQueryListener) {
      if (mediaQuery.removeEventListener)
        mediaQuery.removeEventListener("change", mediaQueryListener);
      else if (mediaQuery.removeListener)
        mediaQuery.removeListener(mediaQueryListener);
    }
  });
  runBestEffort(() => {
    if (dockResizeListener) {
      ownerWindow?.removeEventListener?.("dock-resize", dockResizeListener);
    }
    if (windowResizeListener) {
      ownerWindow?.removeEventListener?.("resize", windowResizeListener);
    }
  });
  runBestEffort(() => {
    themeObserver?.disconnect?.();
  });
  runBestEffort(() => {
    resizeObserver?.disconnect?.();
  });
}
function resizeTerminalRuntimeBestEffort(options) {
  options.syncHostLayout();
  runBestEffort(() => {
    options.terminal?.renderer?.remeasureFont?.();
  });
  runBestEffort(() => {
    options.fitAddon?.fit?.();
  });
  runBestEffort(() => {
    options.terminal?.refresh?.();
  });
  options.syncHostLayout();
  options.sendResize();
}
function disposeTerminalRuntimeBestEffort(options) {
  const {
    resizeFrame = 0,
    cancelAnimationFrameFn = cancelAnimationFrame,
    socket,
    fitAddon,
    terminal,
    termEl
  } = options;
  if (resizeFrame) {
    runBestEffort(() => {
      cancelAnimationFrameFn(resizeFrame);
    });
  }
  runBestEffort(() => {
    socket?.close?.();
  });
  runBestEffort(() => {
    fitAddon?.dispose?.();
  });
  runBestEffort(() => {
    terminal?.dispose?.();
  });
  termEl?.remove?.();
  return 0;
}

// web/src/panes/terminal-theme-runtime.ts
function runBestEffort2(run) {
  try {
    run();
    return true;
  } catch (_error) {
    return false;
  }
}
function applyTerminalThemeBestEffort(options) {
  const { termEl, bodyEl, terminal, theme, themeChanged = false, socket, resize } = options;
  runBestEffort2(() => {
    if (termEl?.style) {
      termEl.style.backgroundColor = theme.background;
    }
    if (bodyEl?.style) {
      bodyEl.style.backgroundColor = theme.background;
    }
    const host = bodyEl?.querySelector?.(".terminal-live-host");
    if (host && typeof host === "object" && "style" in host) {
      host.style.backgroundColor = theme.background;
      host.style.color = theme.foreground;
    }
    const canvas = bodyEl?.querySelector?.("canvas");
    if (canvas && typeof canvas === "object" && "style" in canvas) {
      canvas.style.backgroundColor = theme.background;
      canvas.style.color = theme.foreground;
    }
  });
  runBestEffort2(() => {
    if (terminal?.options) {
      terminal.options.theme = theme;
    }
  });
  if (themeChanged) {
    runBestEffort2(() => {
      terminal?.reset?.();
    });
  }
  runBestEffort2(() => {
    terminal?.renderer?.setTheme?.(theme);
    terminal?.renderer?.clear?.();
  });
  runBestEffort2(() => {
    terminal?.loadFonts?.();
  });
  runBestEffort2(() => {
    terminal?.renderer?.remeasureFont?.();
  });
  runBestEffort2(() => {
    if (terminal?.wasmTerm && terminal?.renderer?.render) {
      terminal.renderer.render(terminal.wasmTerm, true, terminal.viewportY || 0, terminal);
      terminal.renderer.render(terminal.wasmTerm, false, terminal.viewportY || 0, terminal);
    }
  });
  runBestEffort2(() => {
    resize?.();
  });
  runBestEffort2(() => {
    if (themeChanged && socket?.readyState === 1) {
      socket.send?.(JSON.stringify({ type: "input", data: "\f" }));
    }
  });
  runBestEffort2(() => {
    terminal?.refresh?.();
  });
}

// web/src/panes/terminal-pane.ts
var GHOSTTY_WEB_MODULE = "/static/js/vendor/ghostty-web.js";
var GHOSTTY_WASM_MODULE = "/static/js/vendor/ghostty-vt.wasm";
var TERMINAL_FONT_FAMILY = 'FiraCode Nerd Font Mono, ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace';
var TERMINAL_FONT_LOAD_SPEC = '400 13px "FiraCode Nerd Font Mono"';
var TERMINAL_FONT_LOAD_SPEC_BOLD = '700 13px "FiraCode Nerd Font Mono"';
var TERMINAL_ANON_CLIENT_HEADER = "x-piclaw-terminal-client";
var TERMINAL_ANON_CLIENT_STORAGE_KEY = "piclaw_terminal_client";
var LIGHT_TERMINAL_PALETTE = {
  yellow: "#9a6700",
  magenta: "#8250df",
  cyan: "#1b7c83",
  brightBlack: "#57606a",
  brightRed: "#cf222e",
  brightGreen: "#1a7f37",
  brightYellow: "#bf8700",
  brightBlue: "#0550ae",
  brightMagenta: "#6f42c1",
  brightCyan: "#0a7b83"
};
var DARK_TERMINAL_PALETTE = {
  yellow: "#d29922",
  magenta: "#bc8cff",
  cyan: "#39c5cf",
  brightBlack: "#8b949e",
  brightRed: "#ff7b72",
  brightGreen: "#7ee787",
  brightYellow: "#e3b341",
  brightBlue: "#79c0ff",
  brightMagenta: "#d2a8ff",
  brightCyan: "#56d4dd"
};
var ghosttyInitPromise = null;
var terminalFontsReadyPromise = null;
function shouldRewriteGhosttyWasmRequest(url) {
  if (!url)
    return false;
  return url.startsWith("data:application/wasm") || /(^|\/)ghostty-vt\.wasm(?:[?#].*)?$/.test(url);
}
async function withGhosttyWasmFetchShim(run) {
  const originalFetch = globalThis.fetch?.bind(globalThis);
  if (!originalFetch) {
    return await run();
  }
  const wasmUrl = new URL(GHOSTTY_WASM_MODULE, window.location.origin).href;
  const patchedFetch = (input, init) => {
    const requestUrl = input instanceof Request ? input.url : input instanceof URL ? input.href : String(input);
    if (!shouldRewriteGhosttyWasmRequest(requestUrl)) {
      return originalFetch(input, init);
    }
    if (input instanceof Request) {
      return originalFetch(new Request(wasmUrl, input));
    }
    return originalFetch(wasmUrl, init);
  };
  globalThis.fetch = patchedFetch;
  try {
    return await run();
  } finally {
    globalThis.fetch = originalFetch;
  }
}
async function loadGhosttyWeb() {
  const moduleUrl = new URL(GHOSTTY_WEB_MODULE, window.location.origin).href;
  const mod = await import(moduleUrl);
  if (!ghosttyInitPromise) {
    ghosttyInitPromise = withGhosttyWasmFetchShim(() => Promise.resolve(mod.init?.())).catch((error) => {
      ghosttyInitPromise = null;
      throw error;
    });
  }
  await ghosttyInitPromise;
  return mod;
}
async function ensureTerminalFontsReady() {
  if (typeof document === "undefined" || !("fonts" in document) || !document.fonts) {
    return;
  }
  if (!terminalFontsReadyPromise) {
    terminalFontsReadyPromise = Promise.allSettled([
      document.fonts.load(TERMINAL_FONT_LOAD_SPEC),
      document.fonts.load(TERMINAL_FONT_LOAD_SPEC_BOLD),
      document.fonts.ready
    ]).then(() => {
      return;
    }).catch(() => {
      return;
    });
  }
  await terminalFontsReadyPromise;
}
function createTerminalClientToken(runtimeWindow = typeof window !== "undefined" ? window : null) {
  try {
    if (typeof runtimeWindow?.crypto?.randomUUID === "function") {
      return runtimeWindow.crypto.randomUUID();
    }
  } catch (error) {
    console.debug("[terminal-pane] Failed to generate crypto-backed terminal client token; falling back.", error);
  }
  return `terminal-client-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}
function getOrCreateAnonymousTerminalClientToken(runtimeWindow = typeof window !== "undefined" ? window : null) {
  if (!runtimeWindow)
    return null;
  try {
    const storage = runtimeWindow.localStorage;
    const existing = typeof storage?.getItem === "function" ? String(storage.getItem(TERMINAL_ANON_CLIENT_STORAGE_KEY) || "").trim() : "";
    if (existing)
      return existing;
    const created = createTerminalClientToken(runtimeWindow);
    storage?.setItem?.(TERMINAL_ANON_CLIENT_STORAGE_KEY, created);
    return created;
  } catch (_error) {
    return createTerminalClientToken(runtimeWindow);
  }
}
async function fetchTerminalSession(clientToken = getOrCreateAnonymousTerminalClientToken()) {
  const response = await fetch("/terminal/session", {
    method: "GET",
    credentials: "same-origin",
    headers: clientToken ? { [TERMINAL_ANON_CLIENT_HEADER]: clientToken } : undefined
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(body?.error || `HTTP ${response.status}`);
  }
  return body;
}
async function requestTerminalHandoff(clientToken = getOrCreateAnonymousTerminalClientToken()) {
  const response = await fetch("/terminal/handoff", {
    method: "POST",
    credentials: "same-origin",
    headers: clientToken ? { [TERMINAL_ANON_CLIENT_HEADER]: clientToken } : undefined
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(body?.error || `HTTP ${response.status}`);
  }
  return typeof body?.handoff?.token === "string" && body.handoff.token.trim() ? body.handoff.token.trim() : null;
}
function buildTerminalWebSocketUrl(path, handoffToken = null, clientToken = getOrCreateAnonymousTerminalClientToken()) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = new URL(`${protocol}//${window.location.host}${path}`);
  if (handoffToken) {
    url.searchParams.set("handoff", String(handoffToken));
  }
  if (clientToken) {
    url.searchParams.set("client", String(clientToken));
  }
  return url.toString();
}
function detectDarkTheme(runtimeWindow = typeof window !== "undefined" ? window : null, runtimeDocument = typeof document !== "undefined" ? document : null) {
  if (!runtimeWindow || !runtimeDocument)
    return false;
  const root = runtimeDocument.documentElement;
  const body = runtimeDocument.body;
  const rootTheme = root?.getAttribute?.("data-theme")?.toLowerCase?.() || "";
  if (rootTheme === "dark")
    return true;
  if (rootTheme === "light")
    return false;
  if (root?.classList?.contains("dark") || body?.classList?.contains("dark"))
    return true;
  if (root?.classList?.contains("light") || body?.classList?.contains("light"))
    return false;
  return Boolean(runtimeWindow.matchMedia?.("(prefers-color-scheme: dark)")?.matches);
}
function readThemeVar(name, fallback = "", runtimeDocument = typeof document !== "undefined" ? document : null) {
  if (!runtimeDocument)
    return fallback;
  const value = getComputedStyle(runtimeDocument.documentElement).getPropertyValue(name)?.trim();
  return value || fallback;
}
function parseThemeColor(input) {
  const raw = String(input || "").trim();
  if (!raw)
    return null;
  const hex = raw.startsWith("#") ? raw.slice(1) : raw;
  if (/^[0-9a-fA-F]{3}$/.test(hex) || /^[0-9a-fA-F]{6}$/.test(hex)) {
    const full = hex.length === 3 ? hex.split("").map((c2) => c2 + c2).join("") : hex;
    const int = parseInt(full, 16);
    return {
      r: int >> 16 & 255,
      g: int >> 8 & 255,
      b: int & 255
    };
  }
  const rgbMatch = raw.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)/i);
  if (rgbMatch) {
    return {
      r: parseInt(rgbMatch[1], 10),
      g: parseInt(rgbMatch[2], 10),
      b: parseInt(rgbMatch[3], 10)
    };
  }
  return null;
}
function relativeLuminance3(color) {
  const toLinear = (value) => {
    const s2 = value / 255;
    return s2 <= 0.03928 ? s2 / 12.92 : Math.pow((s2 + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * toLinear(color.r) + 0.7152 * toLinear(color.g) + 0.0722 * toLinear(color.b);
}
function contrastRatio2(a2, b) {
  const l1 = relativeLuminance3(a2);
  const l2 = relativeLuminance3(b);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
}
function getHighestContrastTextColor(background) {
  const bg = parseThemeColor(background);
  if (!bg)
    return "#ffffff";
  const white = { r: 255, g: 255, b: 255 };
  const black = { r: 0, g: 0, b: 0 };
  return contrastRatio2(bg, white) >= contrastRatio2(bg, black) ? "#ffffff" : "#000000";
}
function toHexColor(color) {
  const clamp = (value) => Math.max(0, Math.min(255, Math.round(value || 0)));
  return `#${[color.r, color.g, color.b].map((value) => clamp(value).toString(16).padStart(2, "0")).join("")}`;
}
function mixThemeColors(base, target, amount) {
  const ratio = Math.max(0, Math.min(1, Number.isFinite(amount) ? amount : 0));
  return {
    r: base.r + (target.r - base.r) * ratio,
    g: base.g + (target.g - base.g) * ratio,
    b: base.b + (target.b - base.b) * ratio
  };
}
function ensureTerminalColorContrast(background, color, minimumRatio = 4.5) {
  const bg = parseThemeColor(background);
  const fg = parseThemeColor(color);
  if (!bg || !fg)
    return color;
  if (contrastRatio2(bg, fg) >= minimumRatio)
    return toHexColor(fg);
  const targetColor = parseThemeColor(getHighestContrastTextColor(background));
  if (!targetColor)
    return toHexColor(fg);
  let best = targetColor;
  let bestAmount = 1;
  let low = 0;
  let high = 1;
  for (let index = 0;index < 14; index += 1) {
    const mid = (low + high) / 2;
    const mixed = mixThemeColors(fg, targetColor, mid);
    if (contrastRatio2(bg, mixed) >= minimumRatio) {
      best = mixed;
      bestAmount = mid;
      high = mid;
    } else {
      low = mid;
    }
  }
  let resolved = toHexColor(best);
  let resolvedColor = parseThemeColor(resolved);
  while (resolvedColor && contrastRatio2(bg, resolvedColor) < minimumRatio && bestAmount < 1) {
    bestAmount = Math.min(1, bestAmount + 0.01);
    resolved = toHexColor(mixThemeColors(fg, targetColor, bestAmount));
    resolvedColor = parseThemeColor(resolved);
  }
  return resolved;
}
function withAlpha(hexColor, alphaHex) {
  if (!hexColor || !hexColor.startsWith("#"))
    return hexColor;
  const value = hexColor.slice(1);
  if (value.length === 3) {
    return `#${value[0]}${value[0]}${value[1]}${value[1]}${value[2]}${value[2]}${alphaHex}`;
  }
  if (value.length === 6) {
    return `#${value}${alphaHex}`;
  }
  return hexColor;
}
function buildTerminalTheme(runtimeWindow = typeof window !== "undefined" ? window : null, runtimeDocument = typeof document !== "undefined" ? document : null) {
  const isDark = detectDarkTheme(runtimeWindow, runtimeDocument);
  const palette = isDark ? DARK_TERMINAL_PALETTE : LIGHT_TERMINAL_PALETTE;
  const background = readThemeVar("--bg-primary", isDark ? "#000000" : "#ffffff", runtimeDocument);
  const foreground = ensureTerminalColorContrast(background, getHighestContrastTextColor(background), 7);
  const accent = readThemeVar("--accent-color", "#1d9bf0", runtimeDocument);
  const danger = readThemeVar("--danger-color", isDark ? "#ff7b72" : "#cf222e", runtimeDocument);
  const success = readThemeVar("--success-color", isDark ? "#7ee787" : "#1a7f37", runtimeDocument);
  const hover = readThemeVar("--bg-hover", isDark ? "#1d1f23" : "#e8ebed", runtimeDocument);
  const selectionBackground = readThemeVar("--accent-soft-strong", withAlpha(accent, isDark ? "47" : "33"), runtimeDocument);
  return {
    background,
    foreground,
    cursor: ensureTerminalColorContrast(background, accent, 3),
    cursorAccent: background,
    selectionBackground,
    selectionForeground: foreground,
    black: ensureTerminalColorContrast(background, hover, 3),
    red: ensureTerminalColorContrast(background, danger, 4.5),
    green: ensureTerminalColorContrast(background, success, 4.5),
    yellow: ensureTerminalColorContrast(background, palette.yellow, 4.5),
    blue: ensureTerminalColorContrast(background, accent, 4.5),
    magenta: ensureTerminalColorContrast(background, palette.magenta, 4.5),
    cyan: ensureTerminalColorContrast(background, palette.cyan, 4.5),
    white: foreground,
    brightBlack: ensureTerminalColorContrast(background, palette.brightBlack, 3),
    brightRed: ensureTerminalColorContrast(background, palette.brightRed, 4.5),
    brightGreen: ensureTerminalColorContrast(background, palette.brightGreen, 4.5),
    brightYellow: ensureTerminalColorContrast(background, palette.brightYellow, 4.5),
    brightBlue: ensureTerminalColorContrast(background, palette.brightBlue, 4.5),
    brightMagenta: ensureTerminalColorContrast(background, palette.brightMagenta, 4.5),
    brightCyan: ensureTerminalColorContrast(background, palette.brightCyan, 4.5),
    brightWhite: foreground
  };
}
class TerminalPaneInstance {
  container;
  ownerDocument;
  ownerWindow;
  disposed = false;
  termEl;
  bodyEl;
  statusEl;
  terminal = null;
  fitAddon = null;
  socket = null;
  themeObserver = null;
  themeChangeListener = null;
  mediaQuery = null;
  mediaQueryListener = null;
  resizeObserver = null;
  dockResizeListener = null;
  windowResizeListener = null;
  resizeFrame = 0;
  resizeRetryTimers = new Set;
  lastAppliedThemeSignature = null;
  lastResizeSignature = null;
  pendingHandoffToken = null;
  standbyHandoffToken = null;
  standbyHandoffRequest = null;
  constructor(container, context) {
    this.container = container;
    this.ownerDocument = container.ownerDocument || document;
    this.ownerWindow = this.ownerDocument.defaultView || window;
    const transferHandoffToken = typeof context?.transferState?.handoffToken === "string" && context.transferState.handoffToken.trim() ? context.transferState.handoffToken.trim() : null;
    const popoutHandoffToken = consumePanePopoutTransferToken("terminal_handoff");
    this.pendingHandoffToken = transferHandoffToken || popoutHandoffToken || null;
    this.termEl = this.ownerDocument.createElement("div");
    this.termEl.className = "terminal-pane-content";
    this.termEl.setAttribute("tabindex", "0");
    this.statusEl = this.ownerDocument.createElement("span");
    this.statusEl.className = "terminal-pane-status";
    this.statusEl.textContent = "Loading terminal…";
    this.bodyEl = this.ownerDocument.createElement("div");
    this.bodyEl.className = "terminal-pane-body";
    this.bodyEl.style.display = "flex";
    this.bodyEl.style.flex = "1 1 auto";
    this.bodyEl.style.minHeight = "0";
    this.bodyEl.innerHTML = '<div class="terminal-placeholder">Bootstrapping ghostty-web…</div>';
    this.termEl.append(this.bodyEl);
    container.appendChild(this.termEl);
    this.bootstrapGhostty();
  }
  setStatus(message) {
    this.statusEl.textContent = message;
    this.termEl.dataset.connectionStatus = message;
    this.termEl.setAttribute("aria-label", `Terminal ${message}`);
  }
  getResizeSignature() {
    try {
      const containerRect = this.container?.getBoundingClientRect?.();
      const bodyRect = this.bodyEl?.getBoundingClientRect?.();
      const cWidth = Number.isFinite(containerRect?.width) ? containerRect.width : 0;
      const cHeight = Number.isFinite(containerRect?.height) ? containerRect.height : 0;
      const bWidth = Number.isFinite(bodyRect?.width) ? bodyRect.width : 0;
      const bHeight = Number.isFinite(bodyRect?.height) ? bodyRect.height : 0;
      return `${Math.round(cWidth)}x${Math.round(cHeight)}:${Math.round(bWidth)}x${Math.round(bHeight)}`;
    } catch {
      return "0x0:0x0";
    }
  }
  syncHostLayout() {
    const host = this.bodyEl.querySelector(".terminal-live-host");
    if (!(host instanceof HTMLElement))
      return;
    host.style.display = "flex";
    host.style.flex = "1 1 auto";
    host.style.width = "100%";
    host.style.height = "100%";
    host.style.minWidth = "0";
    host.style.minHeight = "0";
    host.style.overflow = "hidden";
    const primaryChild = host.firstElementChild;
    if (primaryChild instanceof HTMLElement) {
      primaryChild.style.width = "100%";
      primaryChild.style.height = "100%";
      primaryChild.style.maxWidth = "100%";
      primaryChild.style.minWidth = "0";
      primaryChild.style.minHeight = "0";
      primaryChild.style.flex = "1 1 auto";
      primaryChild.style.display = "block";
    }
    const canvas = host.querySelector("canvas");
    if (canvas instanceof HTMLElement) {
      canvas.style.display = "block";
      canvas.style.maxWidth = "none";
      canvas.style.maxHeight = "none";
    }
  }
  queueResizeRetries(delays = [32, 96, 180, 320, 520, 900]) {
    if (this.disposed || !this.ownerWindow)
      return;
    this.clearResizeRetries();
    for (const delay of delays) {
      const timer = this.ownerWindow.setTimeout(() => {
        this.resizeRetryTimers.delete(timer);
        if (this.disposed)
          return;
        this.scheduleResize(true);
      }, delay);
      this.resizeRetryTimers.add(timer);
    }
  }
  clearResizeRetries() {
    if (!this.ownerWindow || this.resizeRetryTimers.size === 0)
      return;
    for (const timer of Array.from(this.resizeRetryTimers)) {
      try {
        this.ownerWindow.clearTimeout(timer);
      } catch (error) {
        console.debug("[terminal-pane] Ignoring timeout clear failure during resize retry drain.", error, { timer });
      }
    }
    this.resizeRetryTimers.clear();
  }
  scheduleResize(force = false) {
    if (this.disposed)
      return;
    const signature = this.getResizeSignature();
    if (!force && this.lastResizeSignature === signature) {
      return;
    }
    if (this.resizeFrame) {
      cancelAnimationFrame(this.resizeFrame);
    }
    this.resizeFrame = requestAnimationFrame(() => {
      this.resizeFrame = 0;
      this.lastResizeSignature = this.getResizeSignature();
      this.resize();
    });
  }
  async bootstrapGhostty() {
    try {
      const mod = await loadGhosttyWeb();
      await ensureTerminalFontsReady();
      if (this.disposed)
        return;
      this.bodyEl.innerHTML = "";
      const terminalHost = this.ownerDocument.createElement("div");
      terminalHost.className = "terminal-live-host";
      terminalHost.style.display = "flex";
      terminalHost.style.flex = "1 1 auto";
      terminalHost.style.width = "100%";
      terminalHost.style.height = "100%";
      terminalHost.style.minWidth = "0";
      terminalHost.style.minHeight = "0";
      this.bodyEl.appendChild(terminalHost);
      const terminal = new mod.Terminal({
        cols: 120,
        rows: 30,
        cursorBlink: true,
        fontFamily: TERMINAL_FONT_FAMILY,
        fontSize: 13,
        theme: buildTerminalTheme(this.ownerWindow, this.ownerDocument)
      });
      let fitAddon = null;
      if (typeof mod.FitAddon === "function") {
        fitAddon = new mod.FitAddon;
        terminal.loadAddon?.(fitAddon);
      }
      await terminal.open(terminalHost);
      terminalHost.__terminal = terminal;
      this.syncHostLayout();
      terminal.loadFonts?.();
      fitAddon?.observeResize?.();
      this.terminal = terminal;
      this.fitAddon = fitAddon;
      this.installThemeSync();
      this.installResizeSync();
      this.scheduleResize(true);
      this.queueResizeRetries([32, 96, 180, 320]);
      await this.connectBackend();
    } catch (error) {
      console.error("[terminal-pane] Failed to bootstrap ghostty-web:", error);
      if (this.disposed)
        return;
      this.bodyEl.innerHTML = '<div class="terminal-placeholder">Terminal failed to load. Check vendored assets and backend wiring.</div>';
      this.setStatus("Load failed");
    }
  }
  applyTheme() {
    if (!this.terminal)
      return;
    const theme = buildTerminalTheme(this.ownerWindow, this.ownerDocument);
    const themeSignature = JSON.stringify(theme);
    const themeChanged = this.lastAppliedThemeSignature !== null && this.lastAppliedThemeSignature !== themeSignature;
    applyTerminalThemeBestEffort({
      termEl: this.termEl,
      bodyEl: this.bodyEl,
      terminal: this.terminal,
      theme,
      themeChanged,
      socket: this.socket,
      resize: () => this.resize()
    });
    this.lastAppliedThemeSignature = themeSignature;
  }
  installThemeSync() {
    if (!this.ownerWindow || !this.ownerDocument)
      return;
    const syncTheme = () => requestAnimationFrame(() => this.applyTheme());
    syncTheme();
    const onThemeChange = () => syncTheme();
    this.ownerWindow.addEventListener("piclaw-theme-change", onThemeChange);
    this.themeChangeListener = onThemeChange;
    const media = this.ownerWindow.matchMedia?.("(prefers-color-scheme: dark)");
    const onMediaChange = () => syncTheme();
    if (media?.addEventListener)
      media.addEventListener("change", onMediaChange);
    else if (media?.addListener)
      media.addListener(onMediaChange);
    this.mediaQuery = media;
    this.mediaQueryListener = onMediaChange;
    const observer = typeof MutationObserver !== "undefined" ? new MutationObserver(() => syncTheme()) : null;
    observer?.observe(this.ownerDocument.documentElement, {
      attributes: true,
      attributeFilter: ["class", "data-theme", "style"]
    });
    if (this.ownerDocument.body) {
      observer?.observe(this.ownerDocument.body, {
        attributes: true,
        attributeFilter: ["class", "data-theme"]
      });
    }
    this.themeObserver = observer;
  }
  installResizeSync() {
    if (!this.ownerWindow)
      return;
    const onDockResize = () => this.scheduleResize();
    const onWindowResize = () => this.scheduleResize();
    this.ownerWindow.addEventListener("dock-resize", onDockResize);
    this.ownerWindow.addEventListener("resize", onWindowResize);
    this.dockResizeListener = onDockResize;
    this.windowResizeListener = onWindowResize;
    if (typeof ResizeObserver !== "undefined") {
      const observer = new ResizeObserver(() => {
        if (this.disposed)
          return;
        this.scheduleResize();
      });
      observer.observe(this.container);
      observer.observe(this.termEl);
      observer.observe(this.bodyEl);
      this.resizeObserver = observer;
    }
  }
  consumeStandbyHandoffToken() {
    const token = this.standbyHandoffToken || null;
    this.standbyHandoffToken = null;
    return token;
  }
  async ensureStandbyHandoffToken(force = false) {
    if (this.disposed)
      return null;
    if (!force && this.standbyHandoffToken) {
      return this.standbyHandoffToken;
    }
    if (this.standbyHandoffRequest) {
      return await this.standbyHandoffRequest;
    }
    this.standbyHandoffRequest = requestTerminalHandoff().then((token) => {
      if (!token || this.disposed) {
        return null;
      }
      this.standbyHandoffToken = token;
      return token;
    }).catch((error) => {
      console.warn("[terminal-pane] Failed to prepare standby handoff token:", error);
      return null;
    }).finally(() => {
      this.standbyHandoffRequest = null;
    });
    return await this.standbyHandoffRequest;
  }
  async connectBackend() {
    const terminal = this.terminal;
    if (!terminal)
      return;
    try {
      const session = await fetchTerminalSession();
      if (this.disposed)
        return;
      if (!session?.enabled) {
        terminal.write?.(`Terminal backend unavailable: ${session?.error || "disabled"}\r
`);
        this.setStatus("Unavailable");
        return;
      }
      const handoffToken = this.pendingHandoffToken || null;
      const socket = new WebSocket(buildTerminalWebSocketUrl(session.ws_path || "/terminal/ws", handoffToken));
      this.socket = socket;
      this.setStatus(handoffToken ? "Transferring…" : "Connecting…");
      terminal.onData?.((data) => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ type: "input", data }));
        }
      });
      terminal.onResize?.(({ cols, rows }) => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(JSON.stringify({ type: "resize", cols, rows }));
        }
      });
      socket.addEventListener("open", () => {
        if (this.disposed)
          return;
        if (handoffToken && this.pendingHandoffToken === handoffToken) {
          this.pendingHandoffToken = null;
        }
        this.ensureStandbyHandoffToken(false);
        this.setStatus("Connected");
        this.scheduleResize(true);
        this.queueResizeRetries([24, 72, 160, 320]);
      });
      socket.addEventListener("message", (event) => {
        if (this.disposed)
          return;
        let payload = null;
        try {
          payload = JSON.parse(String(event.data));
        } catch {
          payload = { type: "output", data: String(event.data) };
        }
        if (payload?.type === "session") {
          const sessionId = typeof payload.session_id === "string" ? payload.session_id : null;
          terminal.__piclawSessionMeta = {
            sessionId,
            createdAt: typeof payload.created_at === "string" ? payload.created_at : null,
            processPid: typeof payload.process_pid === "number" ? payload.process_pid : null
          };
          if (!this.standbyHandoffToken) {
            this.ensureStandbyHandoffToken(false);
          }
          return;
        }
        if (payload?.type === "output" && typeof payload.data === "string") {
          terminal.write?.(payload.data);
          return;
        }
        if (payload?.type === "exit") {
          terminal.write?.(`\r
[terminal exited]\r
`);
          this.setStatus("Exited");
        }
      });
      socket.addEventListener("close", () => {
        if (this.disposed)
          return;
        this.setStatus("Disconnected");
      });
      socket.addEventListener("error", () => {
        if (this.disposed)
          return;
        this.setStatus("Connection error");
      });
    } catch (error) {
      terminal.write?.(`Terminal backend unavailable: ${error instanceof Error ? error.message : String(error)}\r
`);
      this.setStatus("Unavailable");
    }
  }
  sendResize() {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN || !this.fitAddon?.proposeDimensions) {
      return;
    }
    const dims = this.fitAddon.proposeDimensions();
    if (!dims)
      return;
    this.socket.send(JSON.stringify({ type: "resize", cols: dims.cols, rows: dims.rows }));
  }
  detachHostListeners() {
    detachTerminalHostListenersBestEffort({
      ownerWindow: this.ownerWindow,
      themeChangeListener: this.themeChangeListener,
      mediaQuery: this.mediaQuery,
      mediaQueryListener: this.mediaQueryListener,
      dockResizeListener: this.dockResizeListener,
      windowResizeListener: this.windowResizeListener,
      themeObserver: this.themeObserver,
      resizeObserver: this.resizeObserver
    });
    this.themeChangeListener = null;
    this.mediaQuery = null;
    this.mediaQueryListener = null;
    this.themeObserver = null;
    this.resizeObserver = null;
    this.dockResizeListener = null;
    this.windowResizeListener = null;
  }
  beforeDetachFromHost() {
    this.setStatus("Moving terminal…");
  }
  afterAttachToHost(context) {
    const transferHandoffToken = typeof context?.transferState?.handoffToken === "string" && context.transferState.handoffToken.trim() ? context.transferState.handoffToken.trim() : null;
    if (transferHandoffToken) {
      this.pendingHandoffToken = transferHandoffToken;
    }
    this.installThemeSync();
    this.installResizeSync();
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.setStatus("Connected");
    } else if (this.pendingHandoffToken) {
      this.setStatus("Transferring…");
    } else if (this.socket?.readyState === WebSocket.CONNECTING) {
      this.setStatus("Connecting…");
    }
    this.scheduleResize(true);
    this.queueResizeRetries([32, 96, 180, 320]);
    requestAnimationFrame(() => this.focus());
  }
  moveHost(_container) {
    return false;
  }
  exportHostTransferState() {
    const handoffToken = this.standbyHandoffToken || this.pendingHandoffToken || null;
    return handoffToken ? {
      kind: "terminal",
      live: false,
      handoffToken
    } : null;
  }
  async preparePopoutTransfer() {
    let handoffToken = this.consumeStandbyHandoffToken();
    if (!handoffToken) {
      await this.ensureStandbyHandoffToken(true);
      handoffToken = this.consumeStandbyHandoffToken();
    }
    if (!handoffToken)
      return null;
    this.pendingHandoffToken = handoffToken;
    return { terminal_handoff: handoffToken };
  }
  getContent() {
    return;
  }
  isDirty() {
    return false;
  }
  focus() {
    if (this.terminal?.focus) {
      this.terminal.focus();
      return;
    }
    this.termEl?.focus();
  }
  resize() {
    resizeTerminalRuntimeBestEffort({
      syncHostLayout: () => this.syncHostLayout(),
      terminal: this.terminal,
      fitAddon: this.fitAddon,
      sendResize: () => this.sendResize()
    });
  }
  dispose() {
    if (this.disposed)
      return;
    this.disposed = true;
    this.standbyHandoffToken = null;
    this.standbyHandoffRequest = null;
    this.clearResizeRetries();
    this.detachHostListeners();
    this.resizeFrame = disposeTerminalRuntimeBestEffort({
      resizeFrame: this.resizeFrame,
      socket: this.socket,
      fitAddon: this.fitAddon,
      terminal: this.terminal,
      termEl: this.termEl
    });
  }
}
// web/src/panes/remote-display-socket.ts
function defaultParseControlMessage(text) {
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}
function measureOutboundSize(value) {
  if (typeof value === "string") {
    return new TextEncoder().encode(value).byteLength;
  }
  if (value instanceof ArrayBuffer) {
    return value.byteLength;
  }
  if (ArrayBuffer.isView(value)) {
    return value.byteLength;
  }
  if (value instanceof Blob) {
    return value.size;
  }
  return 0;
}
function measureInboundSize(value) {
  if (typeof value === "string") {
    return value.length;
  }
  if (value instanceof ArrayBuffer) {
    return value.byteLength;
  }
  if (value instanceof Blob) {
    return value.size;
  }
  return Number(value?.size || 0);
}
function closeWebSocketBestEffort(socket) {
  try {
    socket?.close?.();
    return true;
  } catch (_error) {
    return false;
  }
}

class WebSocketRemoteDisplayBoundary {
  socket = null;
  disposed = false;
  options;
  bytesIn = 0;
  bytesOut = 0;
  pendingOutbound = [];
  constructor(options) {
    this.options = options;
  }
  connect() {
    if (this.disposed)
      return;
    closeWebSocketBestEffort(this.socket);
    const socket = new WebSocket(this.options.url);
    socket.binaryType = this.options.binaryType || "arraybuffer";
    socket.addEventListener("open", () => {
      if (this.disposed || this.socket !== socket)
        return;
      this.flushPendingOutbound(socket);
      this.options.onOpen?.();
    });
    socket.addEventListener("message", (event) => {
      if (this.disposed || this.socket !== socket)
        return;
      const size = measureInboundSize(event.data);
      this.bytesIn += size;
      this.emitMetrics();
      if (typeof event.data === "string") {
        const parser = this.options.parseControlMessage || defaultParseControlMessage;
        this.options.onMessage?.({
          kind: "control",
          raw: event.data,
          payload: parser(event.data)
        });
        return;
      }
      this.options.onMessage?.({
        kind: "binary",
        data: event.data,
        size
      });
    });
    socket.addEventListener("close", () => {
      if (this.socket === socket) {
        this.socket = null;
      }
      if (this.disposed)
        return;
      this.options.onClose?.();
    });
    socket.addEventListener("error", () => {
      if (this.disposed || this.socket !== socket)
        return;
      this.options.onError?.();
    });
    this.socket = socket;
  }
  send(data) {
    if (this.disposed)
      return;
    const socket = this.socket;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      this.pendingOutbound.push(data);
      return;
    }
    this.sendNow(socket, data);
  }
  sendControl(payload) {
    this.send(JSON.stringify(payload ?? {}));
  }
  getMetrics() {
    return {
      bytesIn: this.bytesIn,
      bytesOut: this.bytesOut
    };
  }
  dispose() {
    if (this.disposed)
      return;
    this.disposed = true;
    closeWebSocketBestEffort(this.socket);
    this.socket = null;
  }
  emitMetrics() {
    this.options.onMetrics?.(this.getMetrics());
  }
  sendNow(socket, data) {
    const size = measureOutboundSize(data);
    this.bytesOut += size;
    this.emitMetrics();
    socket.send(data);
  }
  flushPendingOutbound(socket) {
    if (this.pendingOutbound.length === 0)
      return;
    const pending = this.pendingOutbound.splice(0);
    for (let index = 0;index < pending.length; index += 1) {
      if (this.disposed || this.socket !== socket || socket.readyState !== WebSocket.OPEN) {
        this.pendingOutbound.unshift(...pending.slice(index));
        return;
      }
      this.sendNow(socket, pending[index]);
    }
  }
}

// node_modules/@assemblyscript/loader/index.js
var ID_OFFSET = -8;
var SIZE_OFFSET = -4;
var ARRAYBUFFER_ID = 1;
var STRING_ID = 2;
var ARRAYBUFFERVIEW = 1 << 0;
var ARRAY = 1 << 1;
var STATICARRAY = 1 << 2;
var VAL_ALIGN_OFFSET = 6;
var VAL_SIGNED = 1 << 11;
var VAL_FLOAT = 1 << 12;
var VAL_MANAGED = 1 << 14;
var ARRAYBUFFERVIEW_BUFFER_OFFSET = 0;
var ARRAYBUFFERVIEW_DATASTART_OFFSET = 4;
var ARRAYBUFFERVIEW_BYTELENGTH_OFFSET = 8;
var ARRAYBUFFERVIEW_SIZE = 12;
var ARRAY_LENGTH_OFFSET = 12;
var ARRAY_SIZE = 16;
var E_NO_EXPORT_TABLE = "Operation requires compiling with --exportTable";
var E_NO_EXPORT_RUNTIME = "Operation requires compiling with --exportRuntime";
var F_NO_EXPORT_RUNTIME = () => {
  throw Error(E_NO_EXPORT_RUNTIME);
};
var BIGINT = typeof BigUint64Array !== "undefined";
var THIS = Symbol();
var STRING_SMALLSIZE = 192;
var STRING_CHUNKSIZE = 1024;
var utf16 = new TextDecoder("utf-16le", { fatal: true });
Object.hasOwn = Object.hasOwn || function(obj, prop) {
  return Object.prototype.hasOwnProperty.call(obj, prop);
};
function getStringImpl(buffer, ptr) {
  let len = new Uint32Array(buffer)[ptr + SIZE_OFFSET >>> 2] >>> 1;
  const wtf16 = new Uint16Array(buffer, ptr, len);
  if (len <= STRING_SMALLSIZE)
    return String.fromCharCode(...wtf16);
  try {
    return utf16.decode(wtf16);
  } catch {
    let str = "", off = 0;
    while (len - off > STRING_CHUNKSIZE) {
      str += String.fromCharCode(...wtf16.subarray(off, off += STRING_CHUNKSIZE));
    }
    return str + String.fromCharCode(...wtf16.subarray(off));
  }
}
function preInstantiate(imports) {
  const extendedExports = {};
  function getString(memory, ptr) {
    if (!memory)
      return "<yet unknown>";
    return getStringImpl(memory.buffer, ptr);
  }
  const env = imports.env = imports.env || {};
  env.abort = env.abort || function abort(msg, file, line, colm) {
    const memory = extendedExports.memory || env.memory;
    throw Error(`abort: ${getString(memory, msg)} at ${getString(memory, file)}:${line}:${colm}`);
  };
  env.trace = env.trace || function trace(msg, n2, ...args) {
    const memory = extendedExports.memory || env.memory;
    console.log(`trace: ${getString(memory, msg)}${n2 ? " " : ""}${args.slice(0, n2).join(", ")}`);
  };
  env.seed = env.seed || Date.now;
  imports.Math = imports.Math || Math;
  imports.Date = imports.Date || Date;
  return extendedExports;
}
function postInstantiate(extendedExports, instance) {
  const exports = instance.exports;
  const memory = exports.memory;
  const table = exports.table;
  const __new = exports.__new || F_NO_EXPORT_RUNTIME;
  const __pin = exports.__pin || F_NO_EXPORT_RUNTIME;
  const __unpin = exports.__unpin || F_NO_EXPORT_RUNTIME;
  const __collect = exports.__collect || F_NO_EXPORT_RUNTIME;
  const __rtti_base = exports.__rtti_base;
  const getTypeinfoCount = __rtti_base ? (arr) => arr[__rtti_base >>> 2] : F_NO_EXPORT_RUNTIME;
  extendedExports.__new = __new;
  extendedExports.__pin = __pin;
  extendedExports.__unpin = __unpin;
  extendedExports.__collect = __collect;
  function getTypeinfo(id) {
    const U32 = new Uint32Array(memory.buffer);
    if ((id >>>= 0) >= getTypeinfoCount(U32))
      throw Error(`invalid id: ${id}`);
    return U32[(__rtti_base + 4 >>> 2) + id];
  }
  function getArrayInfo(id) {
    const info = getTypeinfo(id);
    if (!(info & (ARRAYBUFFERVIEW | ARRAY | STATICARRAY)))
      throw Error(`not an array: ${id}, flags=${info}`);
    return info;
  }
  function getValueAlign(info) {
    return 31 - Math.clz32(info >>> VAL_ALIGN_OFFSET & 31);
  }
  function __newString(str) {
    if (str == null)
      return 0;
    const length = str.length;
    const ptr = __new(length << 1, STRING_ID);
    const U16 = new Uint16Array(memory.buffer);
    for (let i2 = 0, p2 = ptr >>> 1;i2 < length; ++i2)
      U16[p2 + i2] = str.charCodeAt(i2);
    return ptr;
  }
  extendedExports.__newString = __newString;
  function __newArrayBuffer(buf) {
    if (buf == null)
      return 0;
    const bufview = new Uint8Array(buf);
    const ptr = __new(bufview.length, ARRAYBUFFER_ID);
    const U8 = new Uint8Array(memory.buffer);
    U8.set(bufview, ptr);
    return ptr;
  }
  extendedExports.__newArrayBuffer = __newArrayBuffer;
  function __getString(ptr) {
    if (!ptr)
      return null;
    const buffer = memory.buffer;
    const id = new Uint32Array(buffer)[ptr + ID_OFFSET >>> 2];
    if (id !== STRING_ID)
      throw Error(`not a string: ${ptr}`);
    return getStringImpl(buffer, ptr);
  }
  extendedExports.__getString = __getString;
  function getView(alignLog2, signed, float) {
    const buffer = memory.buffer;
    if (float) {
      switch (alignLog2) {
        case 2:
          return new Float32Array(buffer);
        case 3:
          return new Float64Array(buffer);
      }
    } else {
      switch (alignLog2) {
        case 0:
          return new (signed ? Int8Array : Uint8Array)(buffer);
        case 1:
          return new (signed ? Int16Array : Uint16Array)(buffer);
        case 2:
          return new (signed ? Int32Array : Uint32Array)(buffer);
        case 3:
          return new (signed ? BigInt64Array : BigUint64Array)(buffer);
      }
    }
    throw Error(`unsupported align: ${alignLog2}`);
  }
  function __newArray(id, valuesOrCapacity = 0) {
    const input = valuesOrCapacity;
    const info = getArrayInfo(id);
    const align = getValueAlign(info);
    const isArrayLike = typeof input !== "number";
    const length = isArrayLike ? input.length : input;
    const buf = __new(length << align, info & STATICARRAY ? id : ARRAYBUFFER_ID);
    let result;
    if (info & STATICARRAY) {
      result = buf;
    } else {
      __pin(buf);
      const arr = __new(info & ARRAY ? ARRAY_SIZE : ARRAYBUFFERVIEW_SIZE, id);
      __unpin(buf);
      const U32 = new Uint32Array(memory.buffer);
      U32[arr + ARRAYBUFFERVIEW_BUFFER_OFFSET >>> 2] = buf;
      U32[arr + ARRAYBUFFERVIEW_DATASTART_OFFSET >>> 2] = buf;
      U32[arr + ARRAYBUFFERVIEW_BYTELENGTH_OFFSET >>> 2] = length << align;
      if (info & ARRAY)
        U32[arr + ARRAY_LENGTH_OFFSET >>> 2] = length;
      result = arr;
    }
    if (isArrayLike) {
      const view = getView(align, info & VAL_SIGNED, info & VAL_FLOAT);
      const start = buf >>> align;
      if (info & VAL_MANAGED) {
        for (let i2 = 0;i2 < length; ++i2) {
          view[start + i2] = input[i2];
        }
      } else {
        view.set(input, start);
      }
    }
    return result;
  }
  extendedExports.__newArray = __newArray;
  function __getArrayView(arr) {
    const U32 = new Uint32Array(memory.buffer);
    const id = U32[arr + ID_OFFSET >>> 2];
    const info = getArrayInfo(id);
    const align = getValueAlign(info);
    let buf = info & STATICARRAY ? arr : U32[arr + ARRAYBUFFERVIEW_DATASTART_OFFSET >>> 2];
    const length = info & ARRAY ? U32[arr + ARRAY_LENGTH_OFFSET >>> 2] : U32[buf + SIZE_OFFSET >>> 2] >>> align;
    return getView(align, info & VAL_SIGNED, info & VAL_FLOAT).subarray(buf >>>= align, buf + length);
  }
  extendedExports.__getArrayView = __getArrayView;
  function __getArray(arr) {
    const input = __getArrayView(arr);
    const len = input.length;
    const out = new Array(len);
    for (let i2 = 0;i2 < len; i2++)
      out[i2] = input[i2];
    return out;
  }
  extendedExports.__getArray = __getArray;
  function __getArrayBuffer(ptr) {
    const buffer = memory.buffer;
    const length = new Uint32Array(buffer)[ptr + SIZE_OFFSET >>> 2];
    return buffer.slice(ptr, ptr + length);
  }
  extendedExports.__getArrayBuffer = __getArrayBuffer;
  function __getFunction(ptr) {
    if (!table)
      throw Error(E_NO_EXPORT_TABLE);
    const index = new Uint32Array(memory.buffer)[ptr >>> 2];
    return table.get(index);
  }
  extendedExports.__getFunction = __getFunction;
  function getTypedArray(Type, alignLog2, ptr) {
    return new Type(getTypedArrayView(Type, alignLog2, ptr));
  }
  function getTypedArrayView(Type, alignLog2, ptr) {
    const buffer = memory.buffer;
    const U32 = new Uint32Array(buffer);
    return new Type(buffer, U32[ptr + ARRAYBUFFERVIEW_DATASTART_OFFSET >>> 2], U32[ptr + ARRAYBUFFERVIEW_BYTELENGTH_OFFSET >>> 2] >>> alignLog2);
  }
  function attachTypedArrayFunctions(ctor, name, align) {
    extendedExports[`__get${name}`] = getTypedArray.bind(null, ctor, align);
    extendedExports[`__get${name}View`] = getTypedArrayView.bind(null, ctor, align);
  }
  [
    Int8Array,
    Uint8Array,
    Uint8ClampedArray,
    Int16Array,
    Uint16Array,
    Int32Array,
    Uint32Array,
    Float32Array,
    Float64Array
  ].forEach((ctor) => {
    attachTypedArrayFunctions(ctor, ctor.name, 31 - Math.clz32(ctor.BYTES_PER_ELEMENT));
  });
  if (BIGINT) {
    [BigUint64Array, BigInt64Array].forEach((ctor) => {
      attachTypedArrayFunctions(ctor, ctor.name.slice(3), 3);
    });
  }
  extendedExports.memory = extendedExports.memory || memory;
  extendedExports.table = extendedExports.table || table;
  return demangle(exports, extendedExports);
}
function isResponse(src) {
  return typeof Response !== "undefined" && src instanceof Response;
}
function isModule(src) {
  return src instanceof WebAssembly.Module;
}
async function instantiate(source, imports = {}) {
  if (isResponse(source = await source))
    return instantiateStreaming(source, imports);
  const module = isModule(source) ? source : await WebAssembly.compile(source);
  const extended = preInstantiate(imports);
  const instance = await WebAssembly.instantiate(module, imports);
  const exports = postInstantiate(extended, instance);
  return { module, instance, exports };
}
async function instantiateStreaming(source, imports = {}) {
  if (!WebAssembly.instantiateStreaming) {
    return instantiate(isResponse(source = await source) ? source.arrayBuffer() : source, imports);
  }
  const extended = preInstantiate(imports);
  const result = await WebAssembly.instantiateStreaming(source, imports);
  const exports = postInstantiate(extended, result.instance);
  return { ...result, exports };
}
function demangle(exports, extendedExports = {}) {
  const setArgumentsLength = exports["__argumentsLength"] ? (length) => {
    exports["__argumentsLength"].value = length;
  } : exports["__setArgumentsLength"] || exports["__setargc"] || (() => {});
  for (let internalName of Object.keys(exports)) {
    const elem = exports[internalName];
    let parts = internalName.split(".");
    let curr = extendedExports;
    while (parts.length > 1) {
      let part = parts.shift();
      if (!Object.hasOwn(curr, part))
        curr[part] = {};
      curr = curr[part];
    }
    let name = parts[0];
    let hash = name.indexOf("#");
    if (hash >= 0) {
      const className = name.substring(0, hash);
      const classElem = curr[className];
      if (typeof classElem === "undefined" || !classElem.prototype) {
        const ctor = function(...args) {
          return ctor.wrap(ctor.prototype.constructor(0, ...args));
        };
        ctor.prototype = {
          valueOf() {
            return this[THIS];
          }
        };
        ctor.wrap = function(thisValue) {
          return Object.create(ctor.prototype, { [THIS]: { value: thisValue, writable: false } });
        };
        if (classElem)
          Object.getOwnPropertyNames(classElem).forEach((name2) => Object.defineProperty(ctor, name2, Object.getOwnPropertyDescriptor(classElem, name2)));
        curr[className] = ctor;
      }
      name = name.substring(hash + 1);
      curr = curr[className].prototype;
      if (/^(get|set):/.test(name)) {
        if (!Object.hasOwn(curr, name = name.substring(4))) {
          let getter = exports[internalName.replace("set:", "get:")];
          let setter = exports[internalName.replace("get:", "set:")];
          Object.defineProperty(curr, name, {
            get() {
              return getter(this[THIS]);
            },
            set(value) {
              setter(this[THIS], value);
            },
            enumerable: true
          });
        }
      } else {
        if (name === "constructor") {
          (curr[name] = function(...args) {
            setArgumentsLength(args.length);
            return elem(...args);
          }).original = elem;
        } else {
          (curr[name] = function(...args) {
            setArgumentsLength(args.length);
            return elem(this[THIS], ...args);
          }).original = elem;
        }
      }
    } else {
      if (/^(get|set):/.test(name)) {
        if (!Object.hasOwn(curr, name = name.substring(4))) {
          Object.defineProperty(curr, name, {
            get: exports[internalName.replace("set:", "get:")],
            set: exports[internalName.replace("get:", "set:")],
            enumerable: true
          });
        }
      } else if (typeof elem === "function" && elem !== setArgumentsLength) {
        (curr[name] = (...args) => {
          setArgumentsLength(args.length);
          return elem(...args);
        }).original = elem;
      } else {
        curr[name] = elem;
      }
    }
  }
  return extendedExports;
}

// web/src/panes/remote-display-gc.ts
function collectAssemblyScriptGarbageBestEffort(runtime) {
  try {
    runtime?.__collect?.();
    return true;
  } catch (_error) {
    return false;
  }
}

// web/src/panes/remote-display-decoder.ts
var REMOTE_DISPLAY_DECODER_WASM_URL = "/static/js/vendor/remote-display-decoder.wasm";
var pipelinePromise = null;
function normalizeInput(bytes) {
  if (bytes instanceof ArrayBuffer)
    return bytes;
  if (bytes.byteOffset === 0 && bytes.byteLength === bytes.buffer.byteLength) {
    return bytes.buffer;
  }
  return bytes.slice().buffer;
}
async function loadRemoteDisplayWasmDecoder() {
  if (pipelinePromise)
    return pipelinePromise;
  pipelinePromise = (async () => {
    try {
      let callProcess = function(fnName, data, x2, y2, w, h, pf) {
        const input = normalizeInput(data);
        const ptr = ex.__pin(ex.__newArrayBuffer(input));
        try {
          return ex[fnName](ptr, x2, y2, w, h, pf.bitsPerPixel, pf.bigEndian ? 1 : 0, pf.trueColor ? 1 : 0, pf.redMax, pf.greenMax, pf.blueMax, pf.redShift, pf.greenShift, pf.blueShift);
        } finally {
          ex.__unpin(ptr);
          collectAssemblyScriptGarbageBestEffort(ex);
        }
      };
      const response = await fetch(REMOTE_DISPLAY_DECODER_WASM_URL, { credentials: "same-origin" });
      if (!response.ok)
        throw new Error(`HTTP ${response.status}`);
      const instantiated = typeof instantiateStreaming === "function" ? await instantiateStreaming(response, {}) : await instantiate(await response.arrayBuffer(), {});
      const ex = instantiated.exports;
      for (const fn of [
        "initFramebuffer",
        "getFramebufferPtr",
        "getFramebufferLen",
        "getFramebufferWidth",
        "getFramebufferHeight",
        "processRawRect",
        "processCopyRect",
        "processRreRect",
        "processHextileRect",
        "processZrleTileData",
        "decodeRawRectToRgba"
      ]) {
        if (typeof ex[fn] !== "function")
          throw new Error(`${fn} export is missing.`);
      }
      return {
        initFramebuffer(width, height) {
          ex.initFramebuffer(width, height);
        },
        getFramebuffer() {
          const ptr = ex.getFramebufferPtr();
          const len = ex.getFramebufferLen();
          return new Uint8ClampedArray(new Uint8Array(ex.memory.buffer, ptr, len).slice().buffer);
        },
        getFramebufferWidth() {
          return ex.getFramebufferWidth();
        },
        getFramebufferHeight() {
          return ex.getFramebufferHeight();
        },
        processRawRect(data, x2, y2, w, h, pf) {
          return callProcess("processRawRect", data, x2, y2, w, h, pf);
        },
        processCopyRect(dstX, dstY, w, h, srcX, srcY) {
          return ex.processCopyRect(dstX, dstY, w, h, srcX, srcY);
        },
        processRreRect(data, x2, y2, w, h, pf) {
          return callProcess("processRreRect", data, x2, y2, w, h, pf);
        },
        processHextileRect(data, x2, y2, w, h, pf) {
          return callProcess("processHextileRect", data, x2, y2, w, h, pf);
        },
        processZrleTileData(decompressed, x2, y2, w, h, pf) {
          return callProcess("processZrleTileData", decompressed, x2, y2, w, h, pf);
        },
        decodeRawRectToRgba(data, width, height, pf) {
          const input = normalizeInput(data);
          const inputPtr = ex.__pin(ex.__newArrayBuffer(input));
          try {
            const outputPtr = ex.__pin(ex.decodeRawRectToRgba(inputPtr, width, height, pf.bitsPerPixel, pf.bigEndian ? 1 : 0, pf.trueColor ? 1 : 0, pf.redMax, pf.greenMax, pf.blueMax, pf.redShift, pf.greenShift, pf.blueShift));
            try {
              return new Uint8ClampedArray(ex.__getArrayBuffer(outputPtr));
            } finally {
              ex.__unpin(outputPtr);
            }
          } finally {
            ex.__unpin(inputPtr);
            collectAssemblyScriptGarbageBestEffort(ex);
          }
        }
      };
    } catch (error) {
      console.warn("[remote-display] Failed to load WASM pipeline, using JS fallback.", error);
      return null;
    }
  })();
  return pipelinePromise;
}

// web/src/panes/vnc-input.ts
function clamp(value, min, max) {
  return Math.max(min, Math.min(max, value));
}
function encodeVncPointerEvent(buttonMask, x2, y2) {
  const buffer = new Uint8Array(6);
  const safeX = clamp(Math.floor(Number(x2 || 0)), 0, 65535);
  const safeY = clamp(Math.floor(Number(y2 || 0)), 0, 65535);
  buffer[0] = 5;
  buffer[1] = clamp(Math.floor(Number(buttonMask || 0)), 0, 255);
  buffer[2] = safeX >> 8 & 255;
  buffer[3] = safeX & 255;
  buffer[4] = safeY >> 8 & 255;
  buffer[5] = safeY & 255;
  return buffer;
}
function vncButtonMaskForPointerButton(button) {
  switch (Number(button)) {
    case 0:
      return 1;
    case 1:
      return 2;
    case 2:
      return 4;
    default:
      return 0;
  }
}
function resolveVncPointerPressMask(event) {
  const direct = vncButtonMaskForPointerButton(event?.button);
  if (direct)
    return direct;
  const pointerType = String(event?.pointerType || "").toLowerCase();
  if (pointerType === "touch" || pointerType === "pen") {
    return vncButtonMaskForPointerButton(0);
  }
  const buttons = Number(event?.buttons || 0);
  if (buttons & 1)
    return vncButtonMaskForPointerButton(0);
  if (buttons & 4)
    return vncButtonMaskForPointerButton(1);
  if (buttons & 2)
    return vncButtonMaskForPointerButton(2);
  return 0;
}
function shouldReleaseVncPointerContact(event) {
  const type = String(event?.type || "").toLowerCase();
  if (type === "pointerup" || type === "pointercancel" || type === "lostpointercapture") {
    return true;
  }
  const buttons = Number(event?.buttons);
  if (Number.isFinite(buttons) && buttons === 0 && type !== "pointerdown") {
    return true;
  }
  const pointerType = String(event?.pointerType || "").toLowerCase();
  const pressure = Number(event?.pressure);
  if (pointerType === "touch" || pointerType === "pen") {
    if ((type === "pointerleave" || type === "pointerout") && type !== "pointerdown") {
      return true;
    }
    if (Number.isFinite(pressure) && pressure <= 0 && type !== "pointerdown") {
      return true;
    }
  }
  return false;
}
function shouldReleaseVncTouchContact(event) {
  const type = String(event?.type || "").toLowerCase();
  if (type === "touchend" || type === "touchcancel") {
    return true;
  }
  if (type === "touchmove") {
    const activeTouches = Number(event?.touches?.length || 0);
    return activeTouches <= 0;
  }
  return false;
}
function shouldArmVncImplicitReleaseTimer(pointerType) {
  const normalized = String(pointerType || "").toLowerCase();
  return normalized !== "mouse";
}
function mapClientToFramebufferPoint(clientX, clientY, rect, framebufferWidth, framebufferHeight) {
  const width = Math.max(1, Math.floor(Number(framebufferWidth || 0)));
  const height = Math.max(1, Math.floor(Number(framebufferHeight || 0)));
  const rectWidth = Math.max(1, Number(rect?.width || 0));
  const rectHeight = Math.max(1, Number(rect?.height || 0));
  const relX = (Number(clientX || 0) - Number(rect?.left || 0)) / rectWidth;
  const relY = (Number(clientY || 0) - Number(rect?.top || 0)) / rectHeight;
  return {
    x: clamp(Math.floor(relX * width), 0, Math.max(0, width - 1)),
    y: clamp(Math.floor(relY * height), 0, Math.max(0, height - 1))
  };
}
function buildVncWheelPointerEvents(deltaY, x2, y2, baseMask = 0) {
  const wheelBit = Number(deltaY) < 0 ? 8 : 16;
  const pressedMask = clamp(Number(baseMask || 0) | wheelBit, 0, 255);
  return [
    encodeVncPointerEvent(pressedMask, x2, y2),
    encodeVncPointerEvent(Number(baseMask || 0), x2, y2)
  ];
}
function encodeVncKeyEvent(down, keysym) {
  const buffer = new Uint8Array(8);
  const safeKeysym = Math.max(0, Math.min(4294967295, Number(keysym || 0) >>> 0));
  buffer[0] = 4;
  buffer[1] = down ? 1 : 0;
  buffer[4] = safeKeysym >>> 24 & 255;
  buffer[5] = safeKeysym >>> 16 & 255;
  buffer[6] = safeKeysym >>> 8 & 255;
  buffer[7] = safeKeysym & 255;
  return buffer;
}
function normalizeVncPassword(value) {
  if (typeof value !== "string")
    return null;
  return value.length > 0 ? value : null;
}
function computeContainedRemoteDisplayScale(availableWidth, availableHeight, framebufferWidth, framebufferHeight) {
  const safeAvailableWidth = Math.max(1, Math.floor(Number(availableWidth || 0)));
  const safeAvailableHeight = Math.max(1, Math.floor(Number(availableHeight || 0)));
  const safeFramebufferWidth = Math.max(1, Math.floor(Number(framebufferWidth || 0)));
  const safeFramebufferHeight = Math.max(1, Math.floor(Number(framebufferHeight || 0)));
  const scale = Math.min(safeAvailableWidth / safeFramebufferWidth, safeAvailableHeight / safeFramebufferHeight);
  if (!Number.isFinite(scale) || scale <= 0)
    return 1;
  return Math.max(0.01, scale);
}
var KEYSYM_BY_KEY = {
  Backspace: 65288,
  Tab: 65289,
  Enter: 65293,
  Escape: 65307,
  Insert: 65379,
  Delete: 65535,
  Home: 65360,
  End: 65367,
  PageUp: 65365,
  PageDown: 65366,
  ArrowLeft: 65361,
  ArrowUp: 65362,
  ArrowRight: 65363,
  ArrowDown: 65364,
  Shift: 65505,
  ShiftLeft: 65505,
  ShiftRight: 65506,
  Control: 65507,
  ControlLeft: 65507,
  ControlRight: 65508,
  Alt: 65513,
  AltLeft: 65513,
  AltRight: 65514,
  Meta: 65515,
  MetaLeft: 65515,
  MetaRight: 65516,
  Super: 65515,
  Super_L: 65515,
  Super_R: 65516,
  CapsLock: 65509,
  NumLock: 65407,
  ScrollLock: 65300,
  Pause: 65299,
  PrintScreen: 65377,
  ContextMenu: 65383,
  Menu: 65383,
  " ": 32
};
for (let i2 = 1;i2 <= 12; i2 += 1) {
  KEYSYM_BY_KEY[`F${i2}`] = 65470 + (i2 - 1);
}
function resolveVncKeysymFromKeyboardEvent(event) {
  const candidates = [event?.key, event?.code];
  for (const candidate of candidates) {
    if (candidate && Object.prototype.hasOwnProperty.call(KEYSYM_BY_KEY, candidate)) {
      return KEYSYM_BY_KEY[candidate];
    }
  }
  const key = String(event?.key || "");
  const keyCodePoint = key ? key.codePointAt(0) : null;
  const keyUnitLength = keyCodePoint == null ? 0 : keyCodePoint > 65535 ? 2 : 1;
  if (keyCodePoint != null && key.length === keyUnitLength) {
    if (keyCodePoint <= 255)
      return keyCodePoint;
    return (16777216 | keyCodePoint) >>> 0;
  }
  return null;
}

// node_modules/fflate/esm/browser.js
var u8 = Uint8Array;
var u16 = Uint16Array;
var i32 = Int32Array;
var fleb = new u8([0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0, 0, 0, 0]);
var fdeb = new u8([0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13, 0, 0]);
var clim = new u8([16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15]);
var freb = function(eb, start) {
  var b = new u16(31);
  for (var i2 = 0;i2 < 31; ++i2) {
    b[i2] = start += 1 << eb[i2 - 1];
  }
  var r2 = new i32(b[30]);
  for (var i2 = 1;i2 < 30; ++i2) {
    for (var j2 = b[i2];j2 < b[i2 + 1]; ++j2) {
      r2[j2] = j2 - b[i2] << 5 | i2;
    }
  }
  return { b, r: r2 };
};
var _a = freb(fleb, 2);
var fl = _a.b;
var revfl = _a.r;
fl[28] = 258, revfl[258] = 28;
var _b = freb(fdeb, 0);
var fd = _b.b;
var revfd = _b.r;
var rev = new u16(32768);
for (i2 = 0;i2 < 32768; ++i2) {
  x2 = (i2 & 43690) >> 1 | (i2 & 21845) << 1;
  x2 = (x2 & 52428) >> 2 | (x2 & 13107) << 2;
  x2 = (x2 & 61680) >> 4 | (x2 & 3855) << 4;
  rev[i2] = ((x2 & 65280) >> 8 | (x2 & 255) << 8) >> 1;
}
var x2;
var i2;
var hMap = function(cd, mb, r2) {
  var s2 = cd.length;
  var i3 = 0;
  var l2 = new u16(mb);
  for (;i3 < s2; ++i3) {
    if (cd[i3])
      ++l2[cd[i3] - 1];
  }
  var le = new u16(mb);
  for (i3 = 1;i3 < mb; ++i3) {
    le[i3] = le[i3 - 1] + l2[i3 - 1] << 1;
  }
  var co;
  if (r2) {
    co = new u16(1 << mb);
    var rvb = 15 - mb;
    for (i3 = 0;i3 < s2; ++i3) {
      if (cd[i3]) {
        var sv = i3 << 4 | cd[i3];
        var r_1 = mb - cd[i3];
        var v2 = le[cd[i3] - 1]++ << r_1;
        for (var m2 = v2 | (1 << r_1) - 1;v2 <= m2; ++v2) {
          co[rev[v2] >> rvb] = sv;
        }
      }
    }
  } else {
    co = new u16(s2);
    for (i3 = 0;i3 < s2; ++i3) {
      if (cd[i3]) {
        co[i3] = rev[le[cd[i3] - 1]++] >> 15 - cd[i3];
      }
    }
  }
  return co;
};
var flt = new u8(288);
for (i2 = 0;i2 < 144; ++i2)
  flt[i2] = 8;
var i2;
for (i2 = 144;i2 < 256; ++i2)
  flt[i2] = 9;
var i2;
for (i2 = 256;i2 < 280; ++i2)
  flt[i2] = 7;
var i2;
for (i2 = 280;i2 < 288; ++i2)
  flt[i2] = 8;
var i2;
var fdt = new u8(32);
for (i2 = 0;i2 < 32; ++i2)
  fdt[i2] = 5;
var i2;
var flm = /* @__PURE__ */ hMap(flt, 9, 0);
var flrm = /* @__PURE__ */ hMap(flt, 9, 1);
var fdm = /* @__PURE__ */ hMap(fdt, 5, 0);
var fdrm = /* @__PURE__ */ hMap(fdt, 5, 1);
var max = function(a2) {
  var m2 = a2[0];
  for (var i3 = 1;i3 < a2.length; ++i3) {
    if (a2[i3] > m2)
      m2 = a2[i3];
  }
  return m2;
};
var bits = function(d2, p2, m2) {
  var o2 = p2 / 8 | 0;
  return (d2[o2] | d2[o2 + 1] << 8) >> (p2 & 7) & m2;
};
var bits16 = function(d2, p2) {
  var o2 = p2 / 8 | 0;
  return (d2[o2] | d2[o2 + 1] << 8 | d2[o2 + 2] << 16) >> (p2 & 7);
};
var shft = function(p2) {
  return (p2 + 7) / 8 | 0;
};
var slc = function(v2, s2, e2) {
  if (s2 == null || s2 < 0)
    s2 = 0;
  if (e2 == null || e2 > v2.length)
    e2 = v2.length;
  return new u8(v2.subarray(s2, e2));
};
var ec = [
  "unexpected EOF",
  "invalid block type",
  "invalid length/literal",
  "invalid distance",
  "stream finished",
  "no stream handler",
  ,
  "no callback",
  "invalid UTF-8 data",
  "extra field too long",
  "date not in range 1980-2099",
  "filename too long",
  "stream finishing",
  "invalid zip data"
];
var err = function(ind, msg, nt) {
  var e2 = new Error(msg || ec[ind]);
  e2.code = ind;
  if (Error.captureStackTrace)
    Error.captureStackTrace(e2, err);
  if (!nt)
    throw e2;
  return e2;
};
var inflt = function(dat, st, buf, dict) {
  var sl = dat.length, dl = dict ? dict.length : 0;
  if (!sl || st.f && !st.l)
    return buf || new u8(0);
  var noBuf = !buf;
  var resize = noBuf || st.i != 2;
  var noSt = st.i;
  if (noBuf)
    buf = new u8(sl * 3);
  var cbuf = function(l3) {
    var bl = buf.length;
    if (l3 > bl) {
      var nbuf = new u8(Math.max(bl * 2, l3));
      nbuf.set(buf);
      buf = nbuf;
    }
  };
  var final = st.f || 0, pos = st.p || 0, bt = st.b || 0, lm = st.l, dm = st.d, lbt = st.m, dbt = st.n;
  var tbts = sl * 8;
  do {
    if (!lm) {
      final = bits(dat, pos, 1);
      var type = bits(dat, pos + 1, 3);
      pos += 3;
      if (!type) {
        var s2 = shft(pos) + 4, l2 = dat[s2 - 4] | dat[s2 - 3] << 8, t2 = s2 + l2;
        if (t2 > sl) {
          if (noSt)
            err(0);
          break;
        }
        if (resize)
          cbuf(bt + l2);
        buf.set(dat.subarray(s2, t2), bt);
        st.b = bt += l2, st.p = pos = t2 * 8, st.f = final;
        continue;
      } else if (type == 1)
        lm = flrm, dm = fdrm, lbt = 9, dbt = 5;
      else if (type == 2) {
        var hLit = bits(dat, pos, 31) + 257, hcLen = bits(dat, pos + 10, 15) + 4;
        var tl = hLit + bits(dat, pos + 5, 31) + 1;
        pos += 14;
        var ldt = new u8(tl);
        var clt = new u8(19);
        for (var i3 = 0;i3 < hcLen; ++i3) {
          clt[clim[i3]] = bits(dat, pos + i3 * 3, 7);
        }
        pos += hcLen * 3;
        var clb = max(clt), clbmsk = (1 << clb) - 1;
        var clm = hMap(clt, clb, 1);
        for (var i3 = 0;i3 < tl; ) {
          var r2 = clm[bits(dat, pos, clbmsk)];
          pos += r2 & 15;
          var s2 = r2 >> 4;
          if (s2 < 16) {
            ldt[i3++] = s2;
          } else {
            var c2 = 0, n2 = 0;
            if (s2 == 16)
              n2 = 3 + bits(dat, pos, 3), pos += 2, c2 = ldt[i3 - 1];
            else if (s2 == 17)
              n2 = 3 + bits(dat, pos, 7), pos += 3;
            else if (s2 == 18)
              n2 = 11 + bits(dat, pos, 127), pos += 7;
            while (n2--)
              ldt[i3++] = c2;
          }
        }
        var lt = ldt.subarray(0, hLit), dt = ldt.subarray(hLit);
        lbt = max(lt);
        dbt = max(dt);
        lm = hMap(lt, lbt, 1);
        dm = hMap(dt, dbt, 1);
      } else
        err(1);
      if (pos > tbts) {
        if (noSt)
          err(0);
        break;
      }
    }
    if (resize)
      cbuf(bt + 131072);
    var lms = (1 << lbt) - 1, dms = (1 << dbt) - 1;
    var lpos = pos;
    for (;; lpos = pos) {
      var c2 = lm[bits16(dat, pos) & lms], sym = c2 >> 4;
      pos += c2 & 15;
      if (pos > tbts) {
        if (noSt)
          err(0);
        break;
      }
      if (!c2)
        err(2);
      if (sym < 256)
        buf[bt++] = sym;
      else if (sym == 256) {
        lpos = pos, lm = null;
        break;
      } else {
        var add = sym - 254;
        if (sym > 264) {
          var i3 = sym - 257, b = fleb[i3];
          add = bits(dat, pos, (1 << b) - 1) + fl[i3];
          pos += b;
        }
        var d2 = dm[bits16(dat, pos) & dms], dsym = d2 >> 4;
        if (!d2)
          err(3);
        pos += d2 & 15;
        var dt = fd[dsym];
        if (dsym > 3) {
          var b = fdeb[dsym];
          dt += bits16(dat, pos) & (1 << b) - 1, pos += b;
        }
        if (pos > tbts) {
          if (noSt)
            err(0);
          break;
        }
        if (resize)
          cbuf(bt + 131072);
        var end = bt + add;
        if (bt < dt) {
          var shift = dl - dt, dend = Math.min(dt, end);
          if (shift + bt < 0)
            err(3);
          for (;bt < dend; ++bt)
            buf[bt] = dict[shift + bt];
        }
        for (;bt < end; ++bt)
          buf[bt] = buf[bt - dt];
      }
    }
    st.l = lm, st.p = lpos, st.b = bt, st.f = final;
    if (lm)
      final = 1, st.m = lbt, st.d = dm, st.n = dbt;
  } while (!final);
  return bt != buf.length && noBuf ? slc(buf, 0, bt) : buf.subarray(0, bt);
};
var wbits = function(d2, p2, v2) {
  v2 <<= p2 & 7;
  var o2 = p2 / 8 | 0;
  d2[o2] |= v2;
  d2[o2 + 1] |= v2 >> 8;
};
var wbits16 = function(d2, p2, v2) {
  v2 <<= p2 & 7;
  var o2 = p2 / 8 | 0;
  d2[o2] |= v2;
  d2[o2 + 1] |= v2 >> 8;
  d2[o2 + 2] |= v2 >> 16;
};
var hTree = function(d2, mb) {
  var t2 = [];
  for (var i3 = 0;i3 < d2.length; ++i3) {
    if (d2[i3])
      t2.push({ s: i3, f: d2[i3] });
  }
  var s2 = t2.length;
  var t22 = t2.slice();
  if (!s2)
    return { t: et, l: 0 };
  if (s2 == 1) {
    var v2 = new u8(t2[0].s + 1);
    v2[t2[0].s] = 1;
    return { t: v2, l: 1 };
  }
  t2.sort(function(a2, b) {
    return a2.f - b.f;
  });
  t2.push({ s: -1, f: 25001 });
  var l2 = t2[0], r2 = t2[1], i0 = 0, i1 = 1, i22 = 2;
  t2[0] = { s: -1, f: l2.f + r2.f, l: l2, r: r2 };
  while (i1 != s2 - 1) {
    l2 = t2[t2[i0].f < t2[i22].f ? i0++ : i22++];
    r2 = t2[i0 != i1 && t2[i0].f < t2[i22].f ? i0++ : i22++];
    t2[i1++] = { s: -1, f: l2.f + r2.f, l: l2, r: r2 };
  }
  var maxSym = t22[0].s;
  for (var i3 = 1;i3 < s2; ++i3) {
    if (t22[i3].s > maxSym)
      maxSym = t22[i3].s;
  }
  var tr = new u16(maxSym + 1);
  var mbt = ln(t2[i1 - 1], tr, 0);
  if (mbt > mb) {
    var i3 = 0, dt = 0;
    var lft = mbt - mb, cst = 1 << lft;
    t22.sort(function(a2, b) {
      return tr[b.s] - tr[a2.s] || a2.f - b.f;
    });
    for (;i3 < s2; ++i3) {
      var i2_1 = t22[i3].s;
      if (tr[i2_1] > mb) {
        dt += cst - (1 << mbt - tr[i2_1]);
        tr[i2_1] = mb;
      } else
        break;
    }
    dt >>= lft;
    while (dt > 0) {
      var i2_2 = t22[i3].s;
      if (tr[i2_2] < mb)
        dt -= 1 << mb - tr[i2_2]++ - 1;
      else
        ++i3;
    }
    for (;i3 >= 0 && dt; --i3) {
      var i2_3 = t22[i3].s;
      if (tr[i2_3] == mb) {
        --tr[i2_3];
        ++dt;
      }
    }
    mbt = mb;
  }
  return { t: new u8(tr), l: mbt };
};
var ln = function(n2, l2, d2) {
  return n2.s == -1 ? Math.max(ln(n2.l, l2, d2 + 1), ln(n2.r, l2, d2 + 1)) : l2[n2.s] = d2;
};
var lc = function(c2) {
  var s2 = c2.length;
  while (s2 && !c2[--s2])
    ;
  var cl = new u16(++s2);
  var cli = 0, cln = c2[0], cls = 1;
  var w = function(v2) {
    cl[cli++] = v2;
  };
  for (var i3 = 1;i3 <= s2; ++i3) {
    if (c2[i3] == cln && i3 != s2)
      ++cls;
    else {
      if (!cln && cls > 2) {
        for (;cls > 138; cls -= 138)
          w(32754);
        if (cls > 2) {
          w(cls > 10 ? cls - 11 << 5 | 28690 : cls - 3 << 5 | 12305);
          cls = 0;
        }
      } else if (cls > 3) {
        w(cln), --cls;
        for (;cls > 6; cls -= 6)
          w(8304);
        if (cls > 2)
          w(cls - 3 << 5 | 8208), cls = 0;
      }
      while (cls--)
        w(cln);
      cls = 1;
      cln = c2[i3];
    }
  }
  return { c: cl.subarray(0, cli), n: s2 };
};
var clen = function(cf, cl) {
  var l2 = 0;
  for (var i3 = 0;i3 < cl.length; ++i3)
    l2 += cf[i3] * cl[i3];
  return l2;
};
var wfblk = function(out, pos, dat) {
  var s2 = dat.length;
  var o2 = shft(pos + 2);
  out[o2] = s2 & 255;
  out[o2 + 1] = s2 >> 8;
  out[o2 + 2] = out[o2] ^ 255;
  out[o2 + 3] = out[o2 + 1] ^ 255;
  for (var i3 = 0;i3 < s2; ++i3)
    out[o2 + i3 + 4] = dat[i3];
  return (o2 + 4 + s2) * 8;
};
var wblk = function(dat, out, final, syms, lf, df, eb, li, bs, bl, p2) {
  wbits(out, p2++, final);
  ++lf[256];
  var _a2 = hTree(lf, 15), dlt = _a2.t, mlb = _a2.l;
  var _b2 = hTree(df, 15), ddt = _b2.t, mdb = _b2.l;
  var _c = lc(dlt), lclt = _c.c, nlc = _c.n;
  var _d = lc(ddt), lcdt = _d.c, ndc = _d.n;
  var lcfreq = new u16(19);
  for (var i3 = 0;i3 < lclt.length; ++i3)
    ++lcfreq[lclt[i3] & 31];
  for (var i3 = 0;i3 < lcdt.length; ++i3)
    ++lcfreq[lcdt[i3] & 31];
  var _e = hTree(lcfreq, 7), lct = _e.t, mlcb = _e.l;
  var nlcc = 19;
  for (;nlcc > 4 && !lct[clim[nlcc - 1]]; --nlcc)
    ;
  var flen = bl + 5 << 3;
  var ftlen = clen(lf, flt) + clen(df, fdt) + eb;
  var dtlen = clen(lf, dlt) + clen(df, ddt) + eb + 14 + 3 * nlcc + clen(lcfreq, lct) + 2 * lcfreq[16] + 3 * lcfreq[17] + 7 * lcfreq[18];
  if (bs >= 0 && flen <= ftlen && flen <= dtlen)
    return wfblk(out, p2, dat.subarray(bs, bs + bl));
  var lm, ll, dm, dl;
  wbits(out, p2, 1 + (dtlen < ftlen)), p2 += 2;
  if (dtlen < ftlen) {
    lm = hMap(dlt, mlb, 0), ll = dlt, dm = hMap(ddt, mdb, 0), dl = ddt;
    var llm = hMap(lct, mlcb, 0);
    wbits(out, p2, nlc - 257);
    wbits(out, p2 + 5, ndc - 1);
    wbits(out, p2 + 10, nlcc - 4);
    p2 += 14;
    for (var i3 = 0;i3 < nlcc; ++i3)
      wbits(out, p2 + 3 * i3, lct[clim[i3]]);
    p2 += 3 * nlcc;
    var lcts = [lclt, lcdt];
    for (var it = 0;it < 2; ++it) {
      var clct = lcts[it];
      for (var i3 = 0;i3 < clct.length; ++i3) {
        var len = clct[i3] & 31;
        wbits(out, p2, llm[len]), p2 += lct[len];
        if (len > 15)
          wbits(out, p2, clct[i3] >> 5 & 127), p2 += clct[i3] >> 12;
      }
    }
  } else {
    lm = flm, ll = flt, dm = fdm, dl = fdt;
  }
  for (var i3 = 0;i3 < li; ++i3) {
    var sym = syms[i3];
    if (sym > 255) {
      var len = sym >> 18 & 31;
      wbits16(out, p2, lm[len + 257]), p2 += ll[len + 257];
      if (len > 7)
        wbits(out, p2, sym >> 23 & 31), p2 += fleb[len];
      var dst = sym & 31;
      wbits16(out, p2, dm[dst]), p2 += dl[dst];
      if (dst > 3)
        wbits16(out, p2, sym >> 5 & 8191), p2 += fdeb[dst];
    } else {
      wbits16(out, p2, lm[sym]), p2 += ll[sym];
    }
  }
  wbits16(out, p2, lm[256]);
  return p2 + ll[256];
};
var deo = /* @__PURE__ */ new i32([65540, 131080, 131088, 131104, 262176, 1048704, 1048832, 2114560, 2117632]);
var et = /* @__PURE__ */ new u8(0);
var dflt = function(dat, lvl, plvl, pre, post, st) {
  var s2 = st.z || dat.length;
  var o2 = new u8(pre + s2 + 5 * (1 + Math.ceil(s2 / 7000)) + post);
  var w = o2.subarray(pre, o2.length - post);
  var lst = st.l;
  var pos = (st.r || 0) & 7;
  if (lvl) {
    if (pos)
      w[0] = st.r >> 3;
    var opt = deo[lvl - 1];
    var n2 = opt >> 13, c2 = opt & 8191;
    var msk_1 = (1 << plvl) - 1;
    var prev = st.p || new u16(32768), head = st.h || new u16(msk_1 + 1);
    var bs1_1 = Math.ceil(plvl / 3), bs2_1 = 2 * bs1_1;
    var hsh = function(i4) {
      return (dat[i4] ^ dat[i4 + 1] << bs1_1 ^ dat[i4 + 2] << bs2_1) & msk_1;
    };
    var syms = new i32(25000);
    var lf = new u16(288), df = new u16(32);
    var lc_1 = 0, eb = 0, i3 = st.i || 0, li = 0, wi = st.w || 0, bs = 0;
    for (;i3 + 2 < s2; ++i3) {
      var hv = hsh(i3);
      var imod = i3 & 32767, pimod = head[hv];
      prev[imod] = pimod;
      head[hv] = imod;
      if (wi <= i3) {
        var rem = s2 - i3;
        if ((lc_1 > 7000 || li > 24576) && (rem > 423 || !lst)) {
          pos = wblk(dat, w, 0, syms, lf, df, eb, li, bs, i3 - bs, pos);
          li = lc_1 = eb = 0, bs = i3;
          for (var j2 = 0;j2 < 286; ++j2)
            lf[j2] = 0;
          for (var j2 = 0;j2 < 30; ++j2)
            df[j2] = 0;
        }
        var l2 = 2, d2 = 0, ch_1 = c2, dif = imod - pimod & 32767;
        if (rem > 2 && hv == hsh(i3 - dif)) {
          var maxn = Math.min(n2, rem) - 1;
          var maxd = Math.min(32767, i3);
          var ml = Math.min(258, rem);
          while (dif <= maxd && --ch_1 && imod != pimod) {
            if (dat[i3 + l2] == dat[i3 + l2 - dif]) {
              var nl = 0;
              for (;nl < ml && dat[i3 + nl] == dat[i3 + nl - dif]; ++nl)
                ;
              if (nl > l2) {
                l2 = nl, d2 = dif;
                if (nl > maxn)
                  break;
                var mmd = Math.min(dif, nl - 2);
                var md = 0;
                for (var j2 = 0;j2 < mmd; ++j2) {
                  var ti = i3 - dif + j2 & 32767;
                  var pti = prev[ti];
                  var cd = ti - pti & 32767;
                  if (cd > md)
                    md = cd, pimod = ti;
                }
              }
            }
            imod = pimod, pimod = prev[imod];
            dif += imod - pimod & 32767;
          }
        }
        if (d2) {
          syms[li++] = 268435456 | revfl[l2] << 18 | revfd[d2];
          var lin = revfl[l2] & 31, din = revfd[d2] & 31;
          eb += fleb[lin] + fdeb[din];
          ++lf[257 + lin];
          ++df[din];
          wi = i3 + l2;
          ++lc_1;
        } else {
          syms[li++] = dat[i3];
          ++lf[dat[i3]];
        }
      }
    }
    for (i3 = Math.max(i3, wi);i3 < s2; ++i3) {
      syms[li++] = dat[i3];
      ++lf[dat[i3]];
    }
    pos = wblk(dat, w, lst, syms, lf, df, eb, li, bs, i3 - bs, pos);
    if (!lst) {
      st.r = pos & 7 | w[pos / 8 | 0] << 3;
      pos -= 7;
      st.h = head, st.p = prev, st.i = i3, st.w = wi;
    }
  } else {
    for (var i3 = st.w || 0;i3 < s2 + lst; i3 += 65535) {
      var e2 = i3 + 65535;
      if (e2 >= s2) {
        w[pos / 8 | 0] = lst;
        e2 = s2;
      }
      pos = wfblk(w, pos + 1, dat.subarray(i3, e2));
    }
    st.i = s2;
  }
  return slc(o2, 0, pre + shft(pos) + post);
};
var adler = function() {
  var a2 = 1, b = 0;
  return {
    p: function(d2) {
      var n2 = a2, m2 = b;
      var l2 = d2.length | 0;
      for (var i3 = 0;i3 != l2; ) {
        var e2 = Math.min(i3 + 2655, l2);
        for (;i3 < e2; ++i3)
          m2 += n2 += d2[i3];
        n2 = (n2 & 65535) + 15 * (n2 >> 16), m2 = (m2 & 65535) + 15 * (m2 >> 16);
      }
      a2 = n2, b = m2;
    },
    d: function() {
      a2 %= 65521, b %= 65521;
      return (a2 & 255) << 24 | (a2 & 65280) << 8 | (b & 255) << 8 | b >> 8;
    }
  };
};
var dopt = function(dat, opt, pre, post, st) {
  if (!st) {
    st = { l: 1 };
    if (opt.dictionary) {
      var dict = opt.dictionary.subarray(-32768);
      var newDat = new u8(dict.length + dat.length);
      newDat.set(dict);
      newDat.set(dat, dict.length);
      dat = newDat;
      st.w = dict.length;
    }
  }
  return dflt(dat, opt.level == null ? 6 : opt.level, opt.mem == null ? st.l ? Math.ceil(Math.max(8, Math.min(13, Math.log(dat.length))) * 1.5) : 20 : 12 + opt.mem, pre, post, st);
};
var wbytes = function(d2, b, v2) {
  for (;v2; ++b)
    d2[b] = v2, v2 >>>= 8;
};
var zlh = function(c2, o2) {
  var lv = o2.level, fl2 = lv == 0 ? 0 : lv < 6 ? 1 : lv == 9 ? 3 : 2;
  c2[0] = 120, c2[1] = fl2 << 6 | (o2.dictionary && 32);
  c2[1] |= 31 - (c2[0] << 8 | c2[1]) % 31;
  if (o2.dictionary) {
    var h = adler();
    h.p(o2.dictionary);
    wbytes(c2, 2, h.d());
  }
};
var zls = function(d2, dict) {
  if ((d2[0] & 15) != 8 || d2[0] >> 4 > 7 || (d2[0] << 8 | d2[1]) % 31)
    err(6, "invalid zlib data");
  if ((d2[1] >> 5 & 1) == +!dict)
    err(6, "invalid zlib data: " + (d2[1] & 32 ? "need" : "unexpected") + " dictionary");
  return (d2[1] >> 3 & 4) + 2;
};
var Inflate = /* @__PURE__ */ function() {
  function Inflate2(opts, cb) {
    if (typeof opts == "function")
      cb = opts, opts = {};
    this.ondata = cb;
    var dict = opts && opts.dictionary && opts.dictionary.subarray(-32768);
    this.s = { i: 0, b: dict ? dict.length : 0 };
    this.o = new u8(32768);
    this.p = new u8(0);
    if (dict)
      this.o.set(dict);
  }
  Inflate2.prototype.e = function(c2) {
    if (!this.ondata)
      err(5);
    if (this.d)
      err(4);
    if (!this.p.length)
      this.p = c2;
    else if (c2.length) {
      var n2 = new u8(this.p.length + c2.length);
      n2.set(this.p), n2.set(c2, this.p.length), this.p = n2;
    }
  };
  Inflate2.prototype.c = function(final) {
    this.s.i = +(this.d = final || false);
    var bts = this.s.b;
    var dt = inflt(this.p, this.s, this.o);
    this.ondata(slc(dt, bts, this.s.b), this.d);
    this.o = slc(dt, this.s.b - 32768), this.s.b = this.o.length;
    this.p = slc(this.p, this.s.p / 8 | 0), this.s.p &= 7;
  };
  Inflate2.prototype.push = function(chunk, final) {
    this.e(chunk), this.c(final);
  };
  return Inflate2;
}();
function zlibSync(data, opts) {
  if (!opts)
    opts = {};
  var a2 = adler();
  a2.p(data);
  var d2 = dopt(data, opts, opts.dictionary ? 6 : 2, 4);
  return zlh(d2, opts), wbytes(d2, d2.length - 4, a2.d()), d2;
}
var Unzlib = /* @__PURE__ */ function() {
  function Unzlib2(opts, cb) {
    Inflate.call(this, opts, cb);
    this.v = opts && opts.dictionary ? 2 : 1;
  }
  Unzlib2.prototype.push = function(chunk, final) {
    Inflate.prototype.e.call(this, chunk);
    if (this.v) {
      if (this.p.length < 6 && !final)
        return;
      this.p = this.p.subarray(zls(this.p, this.v - 1)), this.v = 0;
    }
    if (final) {
      if (this.p.length < 4)
        err(6, "invalid zlib data");
      this.p = this.p.subarray(0, -4);
    }
    Inflate.prototype.c.call(this, final);
  };
  return Unzlib2;
}();
var td = typeof TextDecoder != "undefined" && /* @__PURE__ */ new TextDecoder;
var tds = 0;
try {
  td.decode(et, { stream: true });
  tds = 1;
} catch (e2) {}

// web/src/panes/vnc-auth.ts
var IP_TABLE = [
  58,
  50,
  42,
  34,
  26,
  18,
  10,
  2,
  60,
  52,
  44,
  36,
  28,
  20,
  12,
  4,
  62,
  54,
  46,
  38,
  30,
  22,
  14,
  6,
  64,
  56,
  48,
  40,
  32,
  24,
  16,
  8,
  57,
  49,
  41,
  33,
  25,
  17,
  9,
  1,
  59,
  51,
  43,
  35,
  27,
  19,
  11,
  3,
  61,
  53,
  45,
  37,
  29,
  21,
  13,
  5,
  63,
  55,
  47,
  39,
  31,
  23,
  15,
  7
];
var FP_TABLE = [
  40,
  8,
  48,
  16,
  56,
  24,
  64,
  32,
  39,
  7,
  47,
  15,
  55,
  23,
  63,
  31,
  38,
  6,
  46,
  14,
  54,
  22,
  62,
  30,
  37,
  5,
  45,
  13,
  53,
  21,
  61,
  29,
  36,
  4,
  44,
  12,
  52,
  20,
  60,
  28,
  35,
  3,
  43,
  11,
  51,
  19,
  59,
  27,
  34,
  2,
  42,
  10,
  50,
  18,
  58,
  26,
  33,
  1,
  41,
  9,
  49,
  17,
  57,
  25
];
var E_TABLE = [
  32,
  1,
  2,
  3,
  4,
  5,
  4,
  5,
  6,
  7,
  8,
  9,
  8,
  9,
  10,
  11,
  12,
  13,
  12,
  13,
  14,
  15,
  16,
  17,
  16,
  17,
  18,
  19,
  20,
  21,
  20,
  21,
  22,
  23,
  24,
  25,
  24,
  25,
  26,
  27,
  28,
  29,
  28,
  29,
  30,
  31,
  32,
  1
];
var P_TABLE = [
  16,
  7,
  20,
  21,
  29,
  12,
  28,
  17,
  1,
  15,
  23,
  26,
  5,
  18,
  31,
  10,
  2,
  8,
  24,
  14,
  32,
  27,
  3,
  9,
  19,
  13,
  30,
  6,
  22,
  11,
  4,
  25
];
var PC1_TABLE = [
  57,
  49,
  41,
  33,
  25,
  17,
  9,
  1,
  58,
  50,
  42,
  34,
  26,
  18,
  10,
  2,
  59,
  51,
  43,
  35,
  27,
  19,
  11,
  3,
  60,
  52,
  44,
  36,
  63,
  55,
  47,
  39,
  31,
  23,
  15,
  7,
  62,
  54,
  46,
  38,
  30,
  22,
  14,
  6,
  61,
  53,
  45,
  37,
  29,
  21,
  13,
  5,
  28,
  20,
  12,
  4
];
var PC2_TABLE = [
  14,
  17,
  11,
  24,
  1,
  5,
  3,
  28,
  15,
  6,
  21,
  10,
  23,
  19,
  12,
  4,
  26,
  8,
  16,
  7,
  27,
  20,
  13,
  2,
  41,
  52,
  31,
  37,
  47,
  55,
  30,
  40,
  51,
  45,
  33,
  48,
  44,
  49,
  39,
  56,
  34,
  53,
  46,
  42,
  50,
  36,
  29,
  32
];
var ROTATIONS = [1, 1, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 2, 2, 2, 1];
var S_BOXES = [
  [
    [14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7],
    [0, 15, 7, 4, 14, 2, 13, 1, 10, 6, 12, 11, 9, 5, 3, 8],
    [4, 1, 14, 8, 13, 6, 2, 11, 15, 12, 9, 7, 3, 10, 5, 0],
    [15, 12, 8, 2, 4, 9, 1, 7, 5, 11, 3, 14, 10, 0, 6, 13]
  ],
  [
    [15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10],
    [3, 13, 4, 7, 15, 2, 8, 14, 12, 0, 1, 10, 6, 9, 11, 5],
    [0, 14, 7, 11, 10, 4, 13, 1, 5, 8, 12, 6, 9, 3, 2, 15],
    [13, 8, 10, 1, 3, 15, 4, 2, 11, 6, 7, 12, 0, 5, 14, 9]
  ],
  [
    [10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8],
    [13, 7, 0, 9, 3, 4, 6, 10, 2, 8, 5, 14, 12, 11, 15, 1],
    [13, 6, 4, 9, 8, 15, 3, 0, 11, 1, 2, 12, 5, 10, 14, 7],
    [1, 10, 13, 0, 6, 9, 8, 7, 4, 15, 14, 3, 11, 5, 2, 12]
  ],
  [
    [7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15],
    [13, 8, 11, 5, 6, 15, 0, 3, 4, 7, 2, 12, 1, 10, 14, 9],
    [10, 6, 9, 0, 12, 11, 7, 13, 15, 1, 3, 14, 5, 2, 8, 4],
    [3, 15, 0, 6, 10, 1, 13, 8, 9, 4, 5, 11, 12, 7, 2, 14]
  ],
  [
    [2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9],
    [14, 11, 2, 12, 4, 7, 13, 1, 5, 0, 15, 10, 3, 9, 8, 6],
    [4, 2, 1, 11, 10, 13, 7, 8, 15, 9, 12, 5, 6, 3, 0, 14],
    [11, 8, 12, 7, 1, 14, 2, 13, 6, 15, 0, 9, 10, 4, 5, 3]
  ],
  [
    [12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11],
    [10, 15, 4, 2, 7, 12, 9, 5, 6, 1, 13, 14, 0, 11, 3, 8],
    [9, 14, 15, 5, 2, 8, 12, 3, 7, 0, 4, 10, 1, 13, 11, 6],
    [4, 3, 2, 12, 9, 5, 15, 10, 11, 14, 1, 7, 6, 0, 8, 13]
  ],
  [
    [4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1],
    [13, 0, 11, 7, 4, 9, 1, 10, 14, 3, 5, 12, 2, 15, 8, 6],
    [1, 4, 11, 13, 12, 3, 7, 14, 10, 15, 6, 8, 0, 5, 9, 2],
    [6, 11, 13, 8, 1, 4, 10, 7, 9, 5, 0, 15, 14, 2, 3, 12]
  ],
  [
    [13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7],
    [1, 15, 13, 8, 10, 3, 7, 4, 12, 5, 6, 11, 0, 14, 9, 2],
    [7, 11, 4, 1, 9, 12, 14, 2, 0, 6, 10, 13, 15, 3, 5, 8],
    [2, 1, 14, 7, 4, 10, 8, 13, 15, 12, 9, 0, 3, 5, 6, 11]
  ]
];
var REVERSED_BITS = new Uint8Array(256);
for (let value = 0;value < 256; value += 1) {
  let reversed = 0;
  for (let bit = 0;bit < 8; bit += 1) {
    reversed = reversed << 1 | value >> bit & 1;
  }
  REVERSED_BITS[value] = reversed;
}
function toUint8Array(bytes) {
  if (bytes instanceof Uint8Array)
    return bytes;
  if (bytes instanceof ArrayBuffer)
    return new Uint8Array(bytes);
  if (ArrayBuffer.isView(bytes))
    return new Uint8Array(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  return new Uint8Array(0);
}
function bytesToBigInt(bytes) {
  let value = 0n;
  const source = toUint8Array(bytes);
  for (const byte of source) {
    value = value << 8n | BigInt(byte);
  }
  return value;
}
function bigIntToBytes(value, length) {
  const out = new Uint8Array(length);
  let remaining = BigInt(value);
  for (let index = length - 1;index >= 0; index -= 1) {
    out[index] = Number(remaining & 0xffn);
    remaining >>= 8n;
  }
  return out;
}
function permuteBits(input, table, inputBitLength) {
  let output = 0n;
  for (const position of table) {
    const bit = BigInt(input) >> BigInt(inputBitLength - position) & 1n;
    output = output << 1n | bit;
  }
  return output;
}
function rotateLeft28(value, amount) {
  const width = 28n;
  const mask = (1n << width) - 1n;
  const shift = BigInt(amount % 28);
  return (value << shift | value >> width - shift) & mask;
}
function buildDesSubkeys(keyBytes) {
  const key56 = permuteBits(bytesToBigInt(keyBytes), PC1_TABLE, 64);
  let left = key56 >> 28n & 0x0fffffffn;
  let right = key56 & 0x0fffffffn;
  const subkeys = [];
  for (const rotation of ROTATIONS) {
    left = rotateLeft28(left, rotation);
    right = rotateLeft28(right, rotation);
    const combined = left << 28n | right;
    subkeys.push(permuteBits(combined, PC2_TABLE, 56));
  }
  return subkeys;
}
function applySBoxes(value48) {
  let output = 0n;
  for (let index = 0;index < 8; index += 1) {
    const shift = BigInt((7 - index) * 6);
    const chunk = Number(value48 >> shift & 0x3fn);
    const row = (chunk & 32) >> 4 | chunk & 1;
    const column = chunk >> 1 & 15;
    output = output << 4n | BigInt(S_BOXES[index][row][column]);
  }
  return output;
}
function desFeistel(right32, subkey48) {
  const expanded = permuteBits(right32, E_TABLE, 32) ^ BigInt(subkey48);
  const substituted = applySBoxes(expanded);
  return permuteBits(substituted, P_TABLE, 32);
}
function encryptDesBlock(blockBytes, keyBytes) {
  const subkeys = buildDesSubkeys(keyBytes);
  const initial = permuteBits(bytesToBigInt(blockBytes), IP_TABLE, 64);
  let left = initial >> 32n & 0xffffffffn;
  let right = initial & 0xffffffffn;
  for (const subkey of subkeys) {
    const nextLeft = right;
    const nextRight = (left ^ desFeistel(right, subkey)) & 0xffffffffn;
    left = nextLeft;
    right = nextRight;
  }
  const preoutput = right << 32n | left;
  return bigIntToBytes(permuteBits(preoutput, FP_TABLE, 64), 8);
}
function buildVncPasswordKey(password) {
  const text = String(password ?? "");
  const key = new Uint8Array(8);
  for (let index = 0;index < 8; index += 1) {
    const codeUnit = index < text.length ? text.charCodeAt(index) & 255 : 0;
    key[index] = REVERSED_BITS[codeUnit];
  }
  return key;
}
function buildVncPasswordAuthResponse(password, challenge) {
  const challengeBytes = toUint8Array(challenge);
  if (challengeBytes.byteLength !== 16) {
    throw new Error(`Invalid VNC auth challenge length ${challengeBytes.byteLength}; expected 16 bytes.`);
  }
  const key = buildVncPasswordKey(password);
  const response = new Uint8Array(16);
  response.set(encryptDesBlock(challengeBytes.slice(0, 8), key), 0);
  response.set(encryptDesBlock(challengeBytes.slice(8, 16), key), 8);
  return response;
}

// web/src/panes/remote-display-vnc.ts
var PROTOCOL = "vnc";
function toEncodingValue(value) {
  return Number(value);
}
function normalizeEncodings(encodings) {
  const raw = Array.isArray(encodings) ? encodings : typeof encodings === "string" ? encodings.split(",").map((item) => item.trim()).filter((item) => item.length > 0) : [];
  const values = [];
  const seen = new Set;
  for (const item of raw) {
    const value = toEncodingValue(item);
    if (!Number.isFinite(value))
      continue;
    const normalized = Number(value);
    if (!seen.has(normalized)) {
      values.push(normalized);
      seen.add(normalized);
    }
  }
  if (values.length > 0)
    return values;
  return [5, 2, 1, 0, -223];
}
function toUint8Array2(chunk) {
  if (chunk instanceof Uint8Array)
    return chunk;
  if (chunk instanceof ArrayBuffer)
    return new Uint8Array(chunk);
  if (ArrayBuffer.isView(chunk))
    return new Uint8Array(chunk.buffer, chunk.byteOffset, chunk.byteLength);
  return new Uint8Array(0);
}
function concatBytes(a2, b) {
  const left = toUint8Array2(a2);
  const right = toUint8Array2(b);
  if (!left.byteLength)
    return new Uint8Array(right);
  if (!right.byteLength)
    return new Uint8Array(left);
  const merged = new Uint8Array(left.byteLength + right.byteLength);
  merged.set(left, 0);
  merged.set(right, left.byteLength);
  return merged;
}
function concatByteChunks(chunks) {
  let total = 0;
  for (const chunk of chunks || [])
    total += chunk?.byteLength || 0;
  const merged = new Uint8Array(total);
  let offset = 0;
  for (const chunk of chunks || []) {
    const bytes = toUint8Array2(chunk);
    merged.set(bytes, offset);
    offset += bytes.byteLength;
  }
  return merged;
}
function createZrleInflater() {
  return (compressed) => {
    const payload = toUint8Array2(compressed);
    try {
      const chunks = [];
      const inflator = new Unzlib((chunk) => {
        chunks.push(new Uint8Array(chunk));
      });
      inflator.push(payload, true);
      if (inflator.err) {
        throw new Error(inflator.msg || "zlib decompression error");
      }
      return concatByteChunks(chunks);
    } catch (error) {
      try {
        const fallback = zlibSync(payload);
        return fallback instanceof Uint8Array ? fallback : new Uint8Array(fallback);
      } catch (fallbackError) {
        const message = fallbackError instanceof Error ? fallbackError.message : "unexpected EOF";
        throw new Error(`unexpected EOF: ${message}`);
      }
    }
  };
}
function asciiBytes(text) {
  return new TextEncoder().encode(String(text || ""));
}
function bytesToAscii(bytes) {
  return new TextDecoder().decode(toUint8Array2(bytes));
}
function parseVersionString(text) {
  const match = /^RFB (\d{3})\.(\d{3})\n$/.exec(String(text || ""));
  if (!match)
    return null;
  return {
    major: parseInt(match[1], 10),
    minor: parseInt(match[2], 10),
    text: match[0]
  };
}
function chooseClientVersion(serverVersion) {
  if (!serverVersion)
    return `RFB 003.008
`;
  if (serverVersion.major > 3 || serverVersion.minor >= 8)
    return `RFB 003.008
`;
  if (serverVersion.minor >= 7)
    return `RFB 003.007
`;
  return `RFB 003.003
`;
}
function parsePixelFormat(view, offset = 0) {
  return {
    bitsPerPixel: view.getUint8(offset),
    depth: view.getUint8(offset + 1),
    bigEndian: view.getUint8(offset + 2) === 1,
    trueColor: view.getUint8(offset + 3) === 1,
    redMax: view.getUint16(offset + 4, false),
    greenMax: view.getUint16(offset + 6, false),
    blueMax: view.getUint16(offset + 8, false),
    redShift: view.getUint8(offset + 10),
    greenShift: view.getUint8(offset + 11),
    blueShift: view.getUint8(offset + 12)
  };
}
function encodePixelFormat(format) {
  const buffer = new ArrayBuffer(20);
  const view = new DataView(buffer);
  view.setUint8(0, 0);
  view.setUint8(1, 0);
  view.setUint8(2, 0);
  view.setUint8(3, 0);
  view.setUint8(4, format.bitsPerPixel);
  view.setUint8(5, format.depth);
  view.setUint8(6, format.bigEndian ? 1 : 0);
  view.setUint8(7, format.trueColor ? 1 : 0);
  view.setUint16(8, format.redMax, false);
  view.setUint16(10, format.greenMax, false);
  view.setUint16(12, format.blueMax, false);
  view.setUint8(14, format.redShift);
  view.setUint8(15, format.greenShift);
  view.setUint8(16, format.blueShift);
  return new Uint8Array(buffer);
}
function buildSetEncodings(encodings) {
  const list = Array.isArray(encodings) ? encodings : [];
  const buffer = new ArrayBuffer(4 + list.length * 4);
  const view = new DataView(buffer);
  view.setUint8(0, 2);
  view.setUint8(1, 0);
  view.setUint16(2, list.length, false);
  let offset = 4;
  for (const encoding of list) {
    view.setInt32(offset, Number(encoding || 0), false);
    offset += 4;
  }
  return new Uint8Array(buffer);
}
function buildFramebufferUpdateRequest(incremental, width, height, x3 = 0, y2 = 0) {
  const buffer = new ArrayBuffer(10);
  const view = new DataView(buffer);
  view.setUint8(0, 3);
  view.setUint8(1, incremental ? 1 : 0);
  view.setUint16(2, x3, false);
  view.setUint16(4, y2, false);
  view.setUint16(6, Math.max(0, width || 0), false);
  view.setUint16(8, Math.max(0, height || 0), false);
  return new Uint8Array(buffer);
}
function scaleChannel(value, max2) {
  const numericMax = Number(max2 || 0);
  if (numericMax <= 0)
    return 0;
  if (numericMax === 255)
    return value & 255;
  return Math.max(0, Math.min(255, Math.round((value || 0) * 255 / numericMax)));
}
function readPixelValue(bytes, offset, bytesPerPixel, bigEndian) {
  if (bytesPerPixel === 1)
    return bytes[offset];
  if (bytesPerPixel === 2) {
    return bigEndian ? (bytes[offset] << 8 | bytes[offset + 1]) >>> 0 : (bytes[offset] | bytes[offset + 1] << 8) >>> 0;
  }
  if (bytesPerPixel === 3) {
    return bigEndian ? (bytes[offset] << 16 | bytes[offset + 1] << 8 | bytes[offset + 2]) >>> 0 : (bytes[offset] | bytes[offset + 1] << 8 | bytes[offset + 2] << 16) >>> 0;
  }
  if (bytesPerPixel === 4) {
    return bigEndian ? (bytes[offset] << 24 >>> 0 | bytes[offset + 1] << 16 | bytes[offset + 2] << 8 | bytes[offset + 3]) >>> 0 : (bytes[offset] | bytes[offset + 1] << 8 | bytes[offset + 2] << 16 | bytes[offset + 3] << 24 >>> 0) >>> 0;
  }
  return 0;
}
function decodeRawRectToRgba(bytes, width, height, pixelFormat) {
  const format = pixelFormat || DEFAULT_CLIENT_PIXEL_FORMAT;
  const src = toUint8Array2(bytes);
  const bytesPerPixel = Math.max(1, Math.floor(Number(format.bitsPerPixel || 0) / 8));
  const expected = Math.max(0, width || 0) * Math.max(0, height || 0) * bytesPerPixel;
  if (src.byteLength < expected) {
    throw new Error(`Incomplete raw rectangle payload: expected ${expected} byte(s), got ${src.byteLength}`);
  }
  if (!format.trueColor) {
    throw new Error("Indexed-colour VNC framebuffers are not supported yet.");
  }
  const rgba = new Uint8ClampedArray(Math.max(0, width || 0) * Math.max(0, height || 0) * 4);
  let srcOffset = 0;
  let dstOffset = 0;
  for (let i3 = 0;i3 < Math.max(0, width || 0) * Math.max(0, height || 0); i3 += 1) {
    const value = readPixelValue(src, srcOffset, bytesPerPixel, format.bigEndian);
    const red = scaleChannel(value >>> format.redShift & format.redMax, format.redMax);
    const green = scaleChannel(value >>> format.greenShift & format.greenMax, format.greenMax);
    const blue = scaleChannel(value >>> format.blueShift & format.blueMax, format.blueMax);
    rgba[dstOffset] = red;
    rgba[dstOffset + 1] = green;
    rgba[dstOffset + 2] = blue;
    rgba[dstOffset + 3] = 255;
    srcOffset += bytesPerPixel;
    dstOffset += 4;
  }
  return rgba;
}
function decodePixelToRgba(bytes, offset, pixelFormat) {
  const format = pixelFormat || DEFAULT_CLIENT_PIXEL_FORMAT;
  const bytesPerPixel = Math.max(1, Math.floor(Number(format.bitsPerPixel || 0) / 8));
  if (bytes.byteLength < offset + bytesPerPixel)
    return null;
  const value = readPixelValue(bytes, offset, bytesPerPixel, format.bigEndian);
  return {
    rgba: [
      scaleChannel(value >>> format.redShift & format.redMax, format.redMax),
      scaleChannel(value >>> format.greenShift & format.greenMax, format.greenMax),
      scaleChannel(value >>> format.blueShift & format.blueMax, format.blueMax),
      255
    ],
    bytesPerPixel
  };
}
function fillRgbaRect(surface, surfaceWidth, x3, y2, width, height, rgba) {
  if (!rgba)
    return;
  for (let row = 0;row < height; row += 1) {
    for (let col = 0;col < width; col += 1) {
      const dst = ((y2 + row) * surfaceWidth + (x3 + col)) * 4;
      surface[dst] = rgba[0];
      surface[dst + 1] = rgba[1];
      surface[dst + 2] = rgba[2];
      surface[dst + 3] = rgba[3];
    }
  }
}
function blitRgbaTile(surface, surfaceWidth, tileX, tileY, tileWidth, tileHeight, tileRgba) {
  for (let row = 0;row < tileHeight; row += 1) {
    const srcStart = row * tileWidth * 4;
    const dstStart = ((tileY + row) * surfaceWidth + tileX) * 4;
    surface.set(tileRgba.subarray(srcStart, srcStart + tileWidth * 4), dstStart);
  }
}
function parseZrleRunLength(bytes, offset) {
  let cursor = offset;
  let runLength = 1;
  while (true) {
    if (bytes.byteLength < cursor + 1)
      return null;
    const value = bytes[cursor++];
    runLength += value;
    if (value !== 255)
      break;
  }
  return { consumed: cursor - offset, runLength };
}
function parseZrleRect(bytes, offset, width, height, pixelFormat, decodeRawRect, inflateZrle) {
  const format = pixelFormat || DEFAULT_CLIENT_PIXEL_FORMAT;
  const bytesPerPixel = Math.max(1, Math.floor(Number(format.bitsPerPixel || 0) / 8));
  if (bytes.byteLength < offset + 4)
    return null;
  const compressedLength = new DataView(bytes.buffer, bytes.byteOffset + offset, 4).getUint32(0, false);
  if (bytes.byteLength < offset + 4 + compressedLength)
    return null;
  const compressed = bytes.slice(offset + 4, offset + 4 + compressedLength);
  let decoded;
  try {
    decoded = inflateZrle(compressed);
  } catch {
    return {
      consumed: 4 + compressedLength,
      skipped: true
    };
  }
  let cursor = 0;
  const rgba = new Uint8ClampedArray(Math.max(0, width || 0) * Math.max(0, height || 0) * 4);
  for (let tileY = 0;tileY < height; tileY += 64) {
    const tileHeight = Math.min(64, height - tileY);
    for (let tileX = 0;tileX < width; tileX += 64) {
      const tileWidth = Math.min(64, width - tileX);
      if (decoded.byteLength < cursor + 1)
        return null;
      const subencoding = decoded[cursor++];
      const paletteSize = subencoding & 127;
      const runLengthEncoded = (subencoding & 128) !== 0;
      if (!runLengthEncoded && paletteSize === 0) {
        const rawLength = tileWidth * tileHeight * bytesPerPixel;
        if (decoded.byteLength < cursor + rawLength)
          return null;
        const tileRgba = decodeRawRect(decoded.slice(cursor, cursor + rawLength), tileWidth, tileHeight, format);
        cursor += rawLength;
        blitRgbaTile(rgba, width, tileX, tileY, tileWidth, tileHeight, tileRgba);
        continue;
      }
      if (!runLengthEncoded && paletteSize === 1) {
        const background = decodePixelToRgba(decoded, cursor, format);
        if (!background)
          return null;
        cursor += background.bytesPerPixel;
        fillRgbaRect(rgba, width, tileX, tileY, tileWidth, tileHeight, background.rgba);
        continue;
      }
      if (!runLengthEncoded && paletteSize > 1 && paletteSize <= 16) {
        const palette = [];
        for (let i3 = 0;i3 < paletteSize; i3 += 1) {
          const color = decodePixelToRgba(decoded, cursor, format);
          if (!color)
            return null;
          cursor += color.bytesPerPixel;
          palette.push(color.rgba);
        }
        const bitsPerIndex = paletteSize <= 2 ? 1 : paletteSize <= 4 ? 2 : 4;
        const rowBytes = Math.ceil(tileWidth * bitsPerIndex / 8);
        const packedLength = rowBytes * tileHeight;
        if (decoded.byteLength < cursor + packedLength)
          return null;
        for (let row = 0;row < tileHeight; row += 1) {
          const rowStart = cursor + row * rowBytes;
          for (let col = 0;col < tileWidth; col += 1) {
            const bitIndex = col * bitsPerIndex;
            const byteIndex = rowStart + (bitIndex >> 3);
            const shift = 8 - bitsPerIndex - (bitIndex & 7);
            const paletteIndex = decoded[byteIndex] >> shift & (1 << bitsPerIndex) - 1;
            fillRgbaRect(rgba, width, tileX + col, tileY + row, 1, 1, palette[paletteIndex]);
          }
        }
        cursor += packedLength;
        continue;
      }
      if (runLengthEncoded && paletteSize === 0) {
        let px = 0;
        let py = 0;
        while (py < tileHeight) {
          const color = decodePixelToRgba(decoded, cursor, format);
          if (!color)
            return null;
          cursor += color.bytesPerPixel;
          const run = parseZrleRunLength(decoded, cursor);
          if (!run)
            return null;
          cursor += run.consumed;
          for (let i3 = 0;i3 < run.runLength; i3 += 1) {
            fillRgbaRect(rgba, width, tileX + px, tileY + py, 1, 1, color.rgba);
            px += 1;
            if (px >= tileWidth) {
              px = 0;
              py += 1;
              if (py >= tileHeight)
                break;
            }
          }
        }
        continue;
      }
      if (runLengthEncoded && paletteSize > 0) {
        const palette = [];
        for (let i3 = 0;i3 < paletteSize; i3 += 1) {
          const color = decodePixelToRgba(decoded, cursor, format);
          if (!color)
            return null;
          cursor += color.bytesPerPixel;
          palette.push(color.rgba);
        }
        let px = 0;
        let py = 0;
        while (py < tileHeight) {
          if (decoded.byteLength < cursor + 1)
            return null;
          const indexByte = decoded[cursor++];
          let paletteIndex = indexByte;
          let runLength = 1;
          if (indexByte & 128) {
            paletteIndex = indexByte & 127;
            const run = parseZrleRunLength(decoded, cursor);
            if (!run)
              return null;
            cursor += run.consumed;
            runLength = run.runLength;
          }
          const color = palette[paletteIndex];
          if (!color)
            return null;
          for (let i3 = 0;i3 < runLength; i3 += 1) {
            fillRgbaRect(rgba, width, tileX + px, tileY + py, 1, 1, color);
            px += 1;
            if (px >= tileWidth) {
              px = 0;
              py += 1;
              if (py >= tileHeight)
                break;
            }
          }
        }
        continue;
      }
      return {
        consumed: 4 + compressedLength,
        skipped: true
      };
    }
  }
  return {
    consumed: 4 + compressedLength,
    rgba,
    decompressed: decoded
  };
}
function parseRreRect(bytes, offset, width, height, pixelFormat) {
  const format = pixelFormat || DEFAULT_CLIENT_PIXEL_FORMAT;
  const bytesPerPixel = Math.max(1, Math.floor(Number(format.bitsPerPixel || 0) / 8));
  if (bytes.byteLength < offset + 4 + bytesPerPixel)
    return null;
  const view = new DataView(bytes.buffer, bytes.byteOffset + offset, bytes.byteLength - offset);
  const subrectCount = view.getUint32(0, false);
  const totalSize = 4 + bytesPerPixel + subrectCount * (bytesPerPixel + 8);
  if (bytes.byteLength < offset + totalSize)
    return null;
  let cursor = offset + 4;
  const background = decodePixelToRgba(bytes, cursor, format);
  if (!background)
    return null;
  cursor += background.bytesPerPixel;
  const rgba = new Uint8ClampedArray(Math.max(0, width || 0) * Math.max(0, height || 0) * 4);
  fillRgbaRect(rgba, width, 0, 0, width, height, background.rgba);
  for (let i3 = 0;i3 < subrectCount; i3 += 1) {
    const color = decodePixelToRgba(bytes, cursor, format);
    if (!color)
      return null;
    cursor += color.bytesPerPixel;
    if (bytes.byteLength < cursor + 8)
      return null;
    const rectView = new DataView(bytes.buffer, bytes.byteOffset + cursor, 8);
    const x3 = rectView.getUint16(0, false);
    const y2 = rectView.getUint16(2, false);
    const rectWidth = rectView.getUint16(4, false);
    const rectHeight = rectView.getUint16(6, false);
    cursor += 8;
    fillRgbaRect(rgba, width, x3, y2, rectWidth, rectHeight, color.rgba);
  }
  return {
    consumed: cursor - offset,
    rgba
  };
}
function parseHextileRect(bytes, offset, width, height, pixelFormat, decodeRawRect) {
  const format = pixelFormat || DEFAULT_CLIENT_PIXEL_FORMAT;
  const bytesPerPixel = Math.max(1, Math.floor(Number(format.bitsPerPixel || 0) / 8));
  const rgba = new Uint8ClampedArray(Math.max(0, width || 0) * Math.max(0, height || 0) * 4);
  let cursor = offset;
  let background = [0, 0, 0, 255];
  let foreground = [255, 255, 255, 255];
  for (let tileY = 0;tileY < height; tileY += 16) {
    const tileHeight = Math.min(16, height - tileY);
    for (let tileX = 0;tileX < width; tileX += 16) {
      const tileWidth = Math.min(16, width - tileX);
      if (bytes.byteLength < cursor + 1)
        return null;
      const subencoding = bytes[cursor++];
      if (subencoding & 1) {
        const rawLength = tileWidth * tileHeight * bytesPerPixel;
        if (bytes.byteLength < cursor + rawLength)
          return null;
        const tileRgba = decodeRawRect(bytes.slice(cursor, cursor + rawLength), tileWidth, tileHeight, format);
        cursor += rawLength;
        blitRgbaTile(rgba, width, tileX, tileY, tileWidth, tileHeight, tileRgba);
        continue;
      }
      if (subencoding & 2) {
        const decoded = decodePixelToRgba(bytes, cursor, format);
        if (!decoded)
          return null;
        background = decoded.rgba;
        cursor += decoded.bytesPerPixel;
      }
      fillRgbaRect(rgba, width, tileX, tileY, tileWidth, tileHeight, background);
      if (subencoding & 4) {
        const decoded = decodePixelToRgba(bytes, cursor, format);
        if (!decoded)
          return null;
        foreground = decoded.rgba;
        cursor += decoded.bytesPerPixel;
      }
      if (subencoding & 8) {
        if (bytes.byteLength < cursor + 1)
          return null;
        const subrectCount = bytes[cursor++];
        for (let i3 = 0;i3 < subrectCount; i3 += 1) {
          let color = foreground;
          if (subencoding & 16) {
            const decoded = decodePixelToRgba(bytes, cursor, format);
            if (!decoded)
              return null;
            color = decoded.rgba;
            cursor += decoded.bytesPerPixel;
          }
          if (bytes.byteLength < cursor + 2)
            return null;
          const xy = bytes[cursor++];
          const wh = bytes[cursor++];
          const subX = xy >> 4;
          const subY = xy & 15;
          const subWidth = (wh >> 4) + 1;
          const subHeight = (wh & 15) + 1;
          fillRgbaRect(rgba, width, tileX + subX, tileY + subY, subWidth, subHeight, color);
        }
      }
    }
  }
  return {
    consumed: cursor - offset,
    rgba
  };
}
var DEFAULT_CLIENT_PIXEL_FORMAT = {
  bitsPerPixel: 32,
  depth: 24,
  bigEndian: false,
  trueColor: true,
  redMax: 255,
  greenMax: 255,
  blueMax: 255,
  redShift: 16,
  greenShift: 8,
  blueShift: 0
};

class VncRemoteDisplayProtocol {
  protocol = PROTOCOL;
  constructor(options = {}) {
    this.shared = options.shared !== false;
    this.decodeRawRect = typeof options.decodeRawRect === "function" ? options.decodeRawRect : decodeRawRectToRgba;
    this.pipeline = options.pipeline || null;
    this.encodings = normalizeEncodings(options.encodings || null);
    this.state = "version";
    this.buffer = new Uint8Array(0);
    this.serverVersion = null;
    this.clientVersionText = null;
    this.framebufferWidth = 0;
    this.framebufferHeight = 0;
    this.serverName = "";
    this.serverPixelFormat = null;
    this.clientPixelFormat = { ...DEFAULT_CLIENT_PIXEL_FORMAT };
    this.password = typeof options.password === "string" && options.password.length > 0 ? options.password : null;
    this.inflateZrle = typeof options.inflateZrle === "function" ? options.inflateZrle : createZrleInflater();
  }
  receive(chunk) {
    if (chunk) {
      this.buffer = concatBytes(this.buffer, chunk);
    }
    const events = [];
    const outgoing = [];
    let progressed = true;
    while (progressed) {
      progressed = false;
      if (this.state === "version") {
        if (this.buffer.byteLength < 12)
          break;
        const bytes = this.consume(12);
        const text = bytesToAscii(bytes);
        const version = parseVersionString(text);
        if (!version) {
          throw new Error(`Unsupported RFB version banner: ${text || "<empty>"}`);
        }
        this.serverVersion = version;
        this.clientVersionText = chooseClientVersion(version);
        outgoing.push(asciiBytes(this.clientVersionText));
        events.push({ type: "protocol-version", protocol: PROTOCOL, server: version.text.trim(), client: this.clientVersionText.trim() });
        this.state = version.minor >= 7 ? "security-types" : "security-type-33";
        progressed = true;
        continue;
      }
      if (this.state === "security-types") {
        if (this.buffer.byteLength < 1)
          break;
        const count = this.buffer[0];
        if (count === 0) {
          if (this.buffer.byteLength < 5)
            break;
          const view = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
          const reasonLength = view.getUint32(1, false);
          if (this.buffer.byteLength < 5 + reasonLength)
            break;
          this.consume(1);
          const reason = bytesToAscii(this.consume(4 + reasonLength).slice(4));
          throw new Error(reason || "VNC server rejected the connection.");
        }
        if (this.buffer.byteLength < 1 + count)
          break;
        this.consume(1);
        const types = Array.from(this.consume(count));
        events.push({ type: "security-types", protocol: PROTOCOL, types });
        let selectedType = null;
        if (types.includes(2) && this.password !== null) {
          selectedType = 2;
        } else if (types.includes(1)) {
          selectedType = 1;
        } else if (types.includes(2)) {
          throw new Error("VNC password authentication is required. Enter a password and reconnect.");
        } else {
          throw new Error(`Unsupported VNC security types: ${types.join(", ") || "none"}. This viewer currently supports only "None" and password-based VNC auth.`);
        }
        outgoing.push(Uint8Array.of(selectedType));
        events.push({ type: "security-selected", protocol: PROTOCOL, securityType: selectedType, label: selectedType === 2 ? "VNC Authentication" : "None" });
        this.state = selectedType === 2 ? "security-challenge" : "security-result";
        progressed = true;
        continue;
      }
      if (this.state === "security-type-33") {
        if (this.buffer.byteLength < 4)
          break;
        const view = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
        const securityType = view.getUint32(0, false);
        this.consume(4);
        if (securityType === 0) {
          if (this.buffer.byteLength < 4)
            break;
          const reasonView = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
          const reasonLength = reasonView.getUint32(0, false);
          if (this.buffer.byteLength < 4 + reasonLength)
            break;
          const reason = bytesToAscii(this.consume(4 + reasonLength).slice(4));
          throw new Error(reason || "VNC server rejected the connection.");
        }
        if (securityType === 2) {
          if (this.password === null) {
            throw new Error("VNC password authentication is required. Enter a password and reconnect.");
          }
          events.push({ type: "security-selected", protocol: PROTOCOL, securityType: 2, label: "VNC Authentication" });
          this.state = "security-challenge";
          progressed = true;
          continue;
        }
        if (securityType !== 1) {
          throw new Error(`Unsupported VNC security type ${securityType}. This viewer currently supports only "None" and password-based VNC auth.`);
        }
        events.push({ type: "security-selected", protocol: PROTOCOL, securityType: 1, label: "None" });
        outgoing.push(Uint8Array.of(this.shared ? 1 : 0));
        this.state = "server-init";
        progressed = true;
        continue;
      }
      if (this.state === "security-challenge") {
        if (this.buffer.byteLength < 16)
          break;
        if (this.password === null) {
          throw new Error("VNC password authentication is required. Enter a password and reconnect.");
        }
        const challenge = this.consume(16);
        outgoing.push(buildVncPasswordAuthResponse(this.password, challenge));
        this.state = "security-result";
        progressed = true;
        continue;
      }
      if (this.state === "security-result") {
        if (this.buffer.byteLength < 4)
          break;
        const view = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
        const result = view.getUint32(0, false);
        this.consume(4);
        if (result !== 0) {
          if (this.buffer.byteLength >= 4) {
            const reasonLength = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength).getUint32(0, false);
            if (this.buffer.byteLength >= 4 + reasonLength) {
              const reason = bytesToAscii(this.consume(4 + reasonLength).slice(4));
              throw new Error(reason || "VNC authentication failed.");
            }
          }
          throw new Error("VNC authentication failed.");
        }
        events.push({ type: "security-result", protocol: PROTOCOL, ok: true });
        outgoing.push(Uint8Array.of(this.shared ? 1 : 0));
        this.state = "server-init";
        progressed = true;
        continue;
      }
      if (this.state === "server-init") {
        if (this.buffer.byteLength < 24)
          break;
        const view = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
        const width = view.getUint16(0, false);
        const height = view.getUint16(2, false);
        const pixelFormat = parsePixelFormat(view, 4);
        const nameLength = view.getUint32(20, false);
        if (this.buffer.byteLength < 24 + nameLength)
          break;
        const fixed = this.consume(24);
        const fixedView = new DataView(fixed.buffer, fixed.byteOffset, fixed.byteLength);
        this.framebufferWidth = fixedView.getUint16(0, false);
        this.framebufferHeight = fixedView.getUint16(2, false);
        this.serverPixelFormat = parsePixelFormat(fixedView, 4);
        this.serverName = bytesToAscii(this.consume(nameLength));
        this.state = "connected";
        if (this.pipeline) {
          this.pipeline.initFramebuffer(this.framebufferWidth, this.framebufferHeight);
        }
        outgoing.push(encodePixelFormat(this.clientPixelFormat));
        outgoing.push(buildSetEncodings(this.encodings));
        outgoing.push(buildFramebufferUpdateRequest(false, this.framebufferWidth, this.framebufferHeight));
        events.push({
          type: "display-init",
          protocol: PROTOCOL,
          width,
          height,
          name: this.serverName,
          pixelFormat
        });
        progressed = true;
        continue;
      }
      if (this.state === "connected") {
        if (this.buffer.byteLength < 1)
          break;
        const type = this.buffer[0];
        if (type === 0) {
          if (this.buffer.byteLength < 4)
            break;
          const headerView = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
          const numberOfRectangles = headerView.getUint16(2, false);
          let offset = 4;
          const rects = [];
          let incomplete = false;
          const usePipeline = !!this.pipeline;
          for (let i3 = 0;i3 < numberOfRectangles; i3 += 1) {
            if (this.buffer.byteLength < offset + 12) {
              incomplete = true;
              break;
            }
            const rectView = new DataView(this.buffer.buffer, this.buffer.byteOffset + offset, 12);
            const x3 = rectView.getUint16(0, false);
            const y2 = rectView.getUint16(2, false);
            const width = rectView.getUint16(4, false);
            const height = rectView.getUint16(6, false);
            const encoding = rectView.getInt32(8, false);
            offset += 12;
            if (encoding === 0) {
              const bytesPerPixel = Math.max(1, Math.floor(Number(this.clientPixelFormat.bitsPerPixel || 0) / 8));
              const dataLength = width * height * bytesPerPixel;
              if (this.buffer.byteLength < offset + dataLength) {
                incomplete = true;
                break;
              }
              const raw = this.buffer.slice(offset, offset + dataLength);
              offset += dataLength;
              if (usePipeline) {
                this.pipeline.processRawRect(raw, x3, y2, width, height, this.clientPixelFormat);
                rects.push({ kind: "pipeline", x: x3, y: y2, width, height });
              } else {
                rects.push({
                  kind: "rgba",
                  x: x3,
                  y: y2,
                  width,
                  height,
                  rgba: this.decodeRawRect(raw, width, height, this.clientPixelFormat)
                });
              }
              continue;
            }
            if (encoding === 2) {
              const rre = parseRreRect(this.buffer, offset, width, height, this.clientPixelFormat);
              if (!rre) {
                incomplete = true;
                break;
              }
              if (usePipeline) {
                const rreData = this.buffer.slice(offset, offset + rre.consumed);
                this.pipeline.processRreRect(rreData, x3, y2, width, height, this.clientPixelFormat);
                rects.push({ kind: "pipeline", x: x3, y: y2, width, height });
              } else {
                rects.push({ kind: "rgba", x: x3, y: y2, width, height, rgba: rre.rgba });
              }
              offset += rre.consumed;
              continue;
            }
            if (encoding === 1) {
              if (this.buffer.byteLength < offset + 4) {
                incomplete = true;
                break;
              }
              const copyView = new DataView(this.buffer.buffer, this.buffer.byteOffset + offset, 4);
              const srcX = copyView.getUint16(0, false);
              const srcY = copyView.getUint16(2, false);
              offset += 4;
              if (usePipeline) {
                this.pipeline.processCopyRect(x3, y2, width, height, srcX, srcY);
                rects.push({ kind: "pipeline", x: x3, y: y2, width, height });
              } else {
                rects.push({ kind: "copy", x: x3, y: y2, width, height, srcX, srcY });
              }
              continue;
            }
            if (encoding === 16) {
              const zrle = parseZrleRect(this.buffer, offset, width, height, this.clientPixelFormat, this.decodeRawRect, this.inflateZrle);
              if (!zrle) {
                incomplete = true;
                break;
              }
              offset += zrle.consumed;
              if (zrle.skipped)
                continue;
              if (usePipeline && zrle.decompressed) {
                this.pipeline.processZrleTileData(zrle.decompressed, x3, y2, width, height, this.clientPixelFormat);
                rects.push({ kind: "pipeline", x: x3, y: y2, width, height });
              } else {
                rects.push({ kind: "rgba", x: x3, y: y2, width, height, rgba: zrle.rgba });
              }
              continue;
            }
            if (encoding === 5) {
              const hextile = parseHextileRect(this.buffer, offset, width, height, this.clientPixelFormat, this.decodeRawRect);
              if (!hextile) {
                incomplete = true;
                break;
              }
              if (usePipeline) {
                const hextileData = this.buffer.slice(offset, offset + hextile.consumed);
                this.pipeline.processHextileRect(hextileData, x3, y2, width, height, this.clientPixelFormat);
                rects.push({ kind: "pipeline", x: x3, y: y2, width, height });
              } else {
                rects.push({ kind: "rgba", x: x3, y: y2, width, height, rgba: hextile.rgba });
              }
              offset += hextile.consumed;
              continue;
            }
            if (encoding === -223) {
              this.framebufferWidth = width;
              this.framebufferHeight = height;
              if (usePipeline) {
                this.pipeline.initFramebuffer(width, height);
              }
              rects.push({ kind: "resize", x: x3, y: y2, width, height });
              continue;
            }
            throw new Error(`Unsupported VNC rectangle encoding ${encoding}. This viewer currently supports ZRLE, Hextile, RRE, CopyRect, raw rectangles, and DesktopSize only.`);
          }
          if (incomplete)
            break;
          this.consume(offset);
          const event = {
            type: "framebuffer-update",
            protocol: PROTOCOL,
            width: this.framebufferWidth,
            height: this.framebufferHeight,
            rects
          };
          if (usePipeline) {
            event.framebuffer = this.pipeline.getFramebuffer();
          }
          events.push(event);
          outgoing.push(buildFramebufferUpdateRequest(true, this.framebufferWidth, this.framebufferHeight));
          progressed = true;
          continue;
        }
        if (type === 2) {
          this.consume(1);
          events.push({ type: "bell", protocol: PROTOCOL });
          progressed = true;
          continue;
        }
        if (type === 3) {
          if (this.buffer.byteLength < 8)
            break;
          const view = new DataView(this.buffer.buffer, this.buffer.byteOffset, this.buffer.byteLength);
          const length = view.getUint32(4, false);
          if (this.buffer.byteLength < 8 + length)
            break;
          this.consume(8);
          const text = bytesToAscii(this.consume(length));
          events.push({ type: "clipboard", protocol: PROTOCOL, text });
          progressed = true;
          continue;
        }
        throw new Error(`Unsupported VNC server message type ${type}.`);
      }
    }
    return { events, outgoing };
  }
  consume(length) {
    const chunk = this.buffer.slice(0, length);
    this.buffer = this.buffer.slice(length);
    return chunk;
  }
}

// web/src/panes/vnc-pane.ts
var VNC_TAB_PREFIX = "piclaw://vnc";
function buildVncTabPath(targetId) {
  const target = String(targetId || "").trim();
  return target ? `${VNC_TAB_PREFIX}/${encodeURIComponent(target)}` : VNC_TAB_PREFIX;
}
var VNC_POPOUT_SECRET_PREFIX = "piclaw:vnc-popout:";
var VNC_POPOUT_SECRET_TTL_MS = 60000;
function getVncPopoutStorage(runtime = globalThis) {
  try {
    return runtime?.localStorage ?? null;
  } catch {
    return null;
  }
}
function generateVncPopoutSecretToken(runtime = globalThis) {
  const uuid = readRandomUuidBestEffort(runtime);
  if (uuid) {
    return uuid;
  }
  return `vnc-popout-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}
function sweepExpiredVncPopoutSecrets(storage, nowMs = Date.now()) {
  if (!storage || typeof storage.key !== "function" || !Number.isFinite(storage.length))
    return;
  const keys = [];
  for (let i3 = 0;i3 < Number(storage.length || 0); i3 += 1) {
    const key = storage.key(i3);
    if (key && key.startsWith(VNC_POPOUT_SECRET_PREFIX))
      keys.push(key);
  }
  for (const key of keys) {
    try {
      const raw = storage.getItem(key);
      if (!raw) {
        storage.removeItem(key);
        continue;
      }
      const parsed = JSON.parse(raw);
      const expiresAt = Number(parsed?.expiresAt || 0);
      if (!Number.isFinite(expiresAt) || expiresAt <= nowMs) {
        storage.removeItem(key);
      }
    } catch {
      removeStorageItemBestEffort(storage, key);
    }
  }
}
function stashVncPopoutPassword(password, runtime = globalThis, nowMs = Date.now()) {
  const normalized = normalizeVncPassword(password);
  if (normalized === null)
    return null;
  const storage = getVncPopoutStorage(runtime);
  if (!storage)
    return null;
  sweepExpiredVncPopoutSecrets(storage, nowMs);
  const token = generateVncPopoutSecretToken(runtime);
  try {
    storage.setItem(`${VNC_POPOUT_SECRET_PREFIX}${token}`, JSON.stringify({
      password: normalized,
      expiresAt: nowMs + VNC_POPOUT_SECRET_TTL_MS
    }));
    return token;
  } catch {
    return null;
  }
}
function consumeVncPopoutPassword(token, runtime = globalThis, nowMs = Date.now()) {
  const normalizedToken = String(token || "").trim();
  if (!normalizedToken)
    return null;
  const storage = getVncPopoutStorage(runtime);
  if (!storage)
    return null;
  sweepExpiredVncPopoutSecrets(storage, nowMs);
  const key = `${VNC_POPOUT_SECRET_PREFIX}${normalizedToken}`;
  try {
    const raw = storage.getItem(key);
    storage.removeItem(key);
    if (!raw)
      return null;
    const parsed = JSON.parse(raw);
    const expiresAt = Number(parsed?.expiresAt || 0);
    if (!Number.isFinite(expiresAt) || expiresAt <= nowMs)
      return null;
    return normalizeVncPassword(parsed?.password);
  } catch {
    try {
      storage.removeItem(key);
    } catch {}
    return null;
  }
}
function createVncPopoutTransferPayload(targetId, password, runtime = globalThis) {
  const target = String(targetId || "").trim();
  if (!target)
    return null;
  const payload = {
    pane_path: buildVncTabPath(target)
  };
  const passwordToken = stashVncPopoutPassword(password, runtime);
  if (passwordToken) {
    payload.vnc_secret = passwordToken;
  }
  return payload;
}
function parseVncTargetFromPath(path) {
  const raw = String(path || "");
  if (raw === VNC_TAB_PREFIX)
    return null;
  if (!raw.startsWith(`${VNC_TAB_PREFIX}/`))
    return null;
  const suffix = raw.slice(`${VNC_TAB_PREFIX}/`.length).trim();
  if (!suffix)
    return null;
  try {
    return decodeURIComponent(suffix);
  } catch {
    return suffix;
  }
}
function esc(value) {
  return String(value || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}
async function fetchVncSession(targetId = null) {
  const url = targetId ? `/vnc/session?target=${encodeURIComponent(targetId)}` : "/vnc/session";
  const response = await fetch(url, { credentials: "same-origin" });
  const body = await response.json().catch(() => ({}));
  if (!response.ok)
    throw new Error(body?.error || `HTTP ${response.status}`);
  return body;
}
function isPanePopoutMode() {
  if (typeof window === "undefined")
    return false;
  try {
    const raw = new URLSearchParams(window.location.search).get("pane_popout");
    const normalized = String(raw || "").trim().toLowerCase();
    return normalized === "1" || normalized === "true" || normalized === "yes";
  } catch {
    return false;
  }
}
function buildVncWebSocketUrl(targetId, handoffToken = null) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = new URL(`${protocol}//${window.location.host}/vnc/ws`);
  url.searchParams.set("target", String(targetId || ""));
  if (handoffToken) {
    url.searchParams.set("handoff", String(handoffToken));
  }
  return url.toString();
}
function buildDirectVncTargetReference(host, port) {
  const rawHost = String(host || "").trim();
  const normalizedPort = Math.floor(Number(port || 0));
  if (!rawHost || !Number.isFinite(normalizedPort) || normalizedPort <= 0 || normalizedPort > 65535)
    return null;
  const normalizedHost = rawHost.includes(":") && !rawHost.startsWith("[") ? `[${rawHost}]` : rawHost;
  return `${normalizedHost}:${normalizedPort}`;
}
function getVncTargetsEmptyStateCopy(options = {}) {
  const enabled = Boolean(options?.enabled);
  const directConnectEnabled = Boolean(options?.directConnectEnabled);
  const targetCount = Array.isArray(options?.targets) ? options.targets.length : Number(options?.targetCount || 0);
  if (targetCount > 0) {
    return {
      title: "",
      body: ""
    };
  }
  if (directConnectEnabled) {
    return {
      title: "No saved VNC targets yet.",
      body: "Connect directly above."
    };
  }
  if (!enabled) {
    return {
      title: "VNC is not configured yet.",
      body: "No saved targets are available and direct connect is disabled on this host."
    };
  }
  return {
    title: "No saved VNC targets yet.",
    body: "This host has no configured VNC targets, and direct connect is disabled."
  };
}
function consumePanePopoutTransferToken2(paramName) {
  if (typeof window === "undefined")
    return null;
  try {
    const url = new URL(window.location.href);
    const token = url.searchParams.get(paramName)?.trim() || "";
    if (!token)
      return null;
    url.searchParams.delete(paramName);
    window.history?.replaceState?.(window.history.state, document.title, url.toString());
    return token;
  } catch {
    return null;
  }
}
function shouldRetryVncPopoutWithoutHandoff(options) {
  const handoffToken = String(options?.handoffToken || "").trim();
  if (!handoffToken)
    return false;
  return Number(options?.bytesIn || 0) <= 0 && !Boolean(options?.hasRenderedFrame) && Number(options?.reconnectAttempts || 0) <= 0;
}
function relocateVncPaneRoot(root, container) {
  if (!root || !container || typeof container.appendChild !== "function")
    return false;
  try {
    container.innerHTML = "";
  } catch {}
  container.appendChild(root);
  return true;
}

class VncPaneInstance {
  container;
  root;
  statusEl;
  bodyEl;
  metricsEl;
  targetSubtitleEl;
  socketBoundary = null;
  protocol = null;
  disposed = false;
  targetId = null;
  targetLabel = null;
  bytesIn = 0;
  bytesOut = 0;
  canvas = null;
  canvasCtx = null;
  displayPlaceholderEl = null;
  displayInfoEl = null;
  displayMetaEl = null;
  displayStageEl = null;
  chromeEl = null;
  sessionShellEl = null;
  resizeObserver = null;
  displayScale = null;
  readOnly = false;
  pointerButtonMask = 0;
  pointerInputAbortController = null;
  pressedKeysyms = new Map;
  passwordInputEl = null;
  authPassword = null;
  directHostInputEl = null;
  directPortInputEl = null;
  directPasswordInputEl = null;
  hasRenderedFrame = false;
  frameTimeoutId = null;
  reconnectTimerId = null;
  reconnectAttempts = 0;
  rawFallbackAttempted = false;
  protocolRecovering = false;
  pendingHandoffToken = null;
  constructor(container, context) {
    this.container = container;
    this.targetId = parseVncTargetFromPath(context?.path);
    this.targetLabel = this.targetId || null;
    this.pendingHandoffToken = consumePanePopoutTransferToken2("vnc_handoff");
    const passwordToken = consumePanePopoutTransferToken2("vnc_secret");
    const transferredPassword = consumeVncPopoutPassword(passwordToken);
    if (transferredPassword !== null) {
      this.authPassword = transferredPassword;
    }
    this.root = document.createElement("div");
    this.root.className = "vnc-pane-shell";
    this.root.style.cssText = "display:flex;flex-direction:column;width:100%;height:100%;background:var(--bg-primary);color:var(--text-primary);";
    this.targetSubtitleEl = null;
    this.statusEl = document.createElement("div");
    this.statusEl.style.cssText = "display:none;";
    this.statusEl.textContent = "";
    this.bodyEl = document.createElement("div");
    this.bodyEl.style.cssText = "flex:1;min-height:0;display:flex;align-items:stretch;justify-content:stretch;padding:12px;";
    this.metricsEl = document.createElement("div");
    this.metricsEl.style.cssText = "display:none;";
    this.updateMetrics();
    this.root.append(this.statusEl, this.bodyEl);
    this.container.appendChild(this.root);
    this.load();
  }
  setStatus(message) {
    this.statusEl.textContent = String(message || "");
  }
  setSessionChromeVisible(visible) {
    if (this.chromeEl) {
      this.chromeEl.style.display = visible ? "grid" : "none";
    }
    if (this.sessionShellEl?.style) {
      this.sessionShellEl.style.gridTemplateRows = visible ? "auto minmax(0,1fr)" : "1fr";
    }
    if (this.displayStageEl?.style) {
      this.displayStageEl.style.padding = visible ? "12px" : "0";
      this.displayStageEl.style.border = visible ? "1px solid var(--border-color)" : "none";
      this.displayStageEl.style.borderRadius = visible ? "16px" : "0";
      this.displayStageEl.style.background = visible ? "#0a0a0a" : "#000";
    }
    if (this.displayPlaceholderEl?.style) {
      this.displayPlaceholderEl.style.display = visible && !this.hasRenderedFrame ? "block" : "none";
    }
  }
  clearReconnectTimer() {
    if (this.reconnectTimerId) {
      clearTimeout(this.reconnectTimerId);
      this.reconnectTimerId = null;
    }
  }
  scheduleReconnect(delayOverrideMs = null) {
    if (this.disposed || !this.targetId)
      return;
    this.clearReconnectTimer();
    const computedDelayMs = Math.min(8000, 1500 + this.reconnectAttempts * 1000);
    const delayMs = Number.isFinite(delayOverrideMs) ? Math.max(0, Number(delayOverrideMs)) : computedDelayMs;
    this.reconnectAttempts += 1;
    this.reconnectTimerId = setTimeout(() => {
      this.reconnectTimerId = null;
      if (this.disposed || !this.targetId)
        return;
      this.connectSocket();
    }, delayMs);
  }
  updateMetrics() {
    this.metricsEl.textContent = `Transport bytes — in: ${this.bytesIn} / out: ${this.bytesOut}`;
  }
  applyMetrics(metrics) {
    this.bytesIn = Number(metrics?.bytesIn || 0);
    this.bytesOut = Number(metrics?.bytesOut || 0);
    this.updateMetrics();
  }
  openTargetTab(targetId, label) {
    this.targetId = String(targetId || "").trim() || null;
    this.targetLabel = String(label || targetId || "").trim() || this.targetId || "VNC";
    if (this.targetId) {
      this.renderTargetSession({
        direct_connect_enabled: true,
        target: {
          id: this.targetId,
          label: this.targetLabel,
          read_only: false,
          direct_connect: true
        }
      });
      this.setStatus("Connecting…");
      this.updateDisplayInfo("Connecting…");
      this.updateDisplayMeta("connecting");
    }
    this.load();
  }
  requestPanePopout(path, label) {
    this.container.dispatchEvent(new CustomEvent("pane:popout", {
      bubbles: true,
      detail: { path, label }
    }));
  }
  resetLiveSession() {
    this.clearReconnectTimer();
    this.reconnectAttempts = 0;
    this.protocol = null;
    try {
      this.socketBoundary?.dispose?.();
    } catch {}
    this.socketBoundary = null;
    try {
      this.resizeObserver?.disconnect?.();
    } catch {}
    this.resizeObserver = null;
    try {
      this.pointerInputAbortController?.abort?.();
    } catch {}
    this.pointerInputAbortController = null;
    this.canvas = null;
    this.canvasCtx = null;
    this.displayPlaceholderEl = null;
    this.displayInfoEl = null;
    this.displayMetaEl = null;
    this.displayStageEl = null;
    this.displayScale = null;
    this.passwordInputEl = null;
    this.directHostInputEl = null;
    this.directPortInputEl = null;
    this.directPasswordInputEl = null;
    this.hasRenderedFrame = false;
    this.rawFallbackAttempted = false;
    this.protocolRecovering = false;
    if (this.frameTimeoutId) {
      clearTimeout(this.frameTimeoutId);
      this.frameTimeoutId = null;
    }
    this.pressedKeysyms.clear();
  }
  renderTargets(payload) {
    this.resetLiveSession();
    const targets = Array.isArray(payload?.targets) ? payload.targets : [];
    const directConnectEnabled = Boolean(payload?.direct_connect_enabled);
    const emptyState = getVncTargetsEmptyStateCopy({
      enabled: payload?.enabled,
      directConnectEnabled,
      targets
    });
    this.bodyEl.innerHTML = `
            <div style="width:100%;height:100%;min-height:0;display:grid;align-content:start;justify-items:center;gap:16px;overflow:auto;padding:24px;box-sizing:border-box;">
                ${directConnectEnabled ? `
                    <div style="width:min(540px,100%);padding:16px 16px 18px;border:1px solid var(--border-color);border-radius:10px;background:transparent;display:grid;gap:12px;box-shadow:none;">
                        <div style="display:grid;gap:6px;">
                            <div style="font-size:18px;font-weight:700;">Connect to VNC</div>
                            <div style="font-size:12px;color:var(--text-secondary);">Enter a server target to start a direct session.</div>
                        </div>
                        <div style="display:grid;gap:10px;align-items:end;">
                            <label style="display:grid;gap:6px;min-width:0;">
                                <span style="font-size:12px;color:var(--text-secondary);">Server</span>
                                <input type="text" data-vnc-direct-host placeholder="server" spellcheck="false" style="width:100%;padding:10px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;color:inherit;" />
                            </label>
                            <label style="display:grid;gap:6px;min-width:0;">
                                <span style="font-size:12px;color:var(--text-secondary);">Port</span>
                                <input type="number" data-vnc-direct-port min="1" max="65535" step="1" placeholder="5900" style="width:100%;padding:10px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;color:inherit;" />
                            </label>
                            <label style="display:grid;gap:6px;min-width:0;">
                                <span style="font-size:12px;color:var(--text-secondary);">Password</span>
                                <input type="password" data-vnc-direct-password placeholder="Optional" autocomplete="current-password" style="width:100%;padding:10px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;color:inherit;" />
                            </label>
                            <button type="button" data-direct-open-tab="1" style="padding:10px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;cursor:pointer;color:inherit;min-height:40px;font-weight:500;">Connect</button>
                        </div>
                    </div>
                ` : ""}
                ${targets.length ? `
                    <div style="width:min(100%,900px);min-height:0;display:grid;gap:12px;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));align-content:start;">
                        ${targets.map((target) => `
                            <div style="text-align:left;padding:14px;border:1px solid var(--border-color);border-radius:10px;background:transparent;color:inherit;display:flex;flex-direction:column;gap:10px;">
                                <div>
                                    <div style="font-weight:600;margin-bottom:6px;">${esc(target.label || target.id)}</div>
                                    <div style="font:12px var(--font-family-mono, monospace);color:var(--text-secondary);">${esc(target.id)}</div>
                                    <div style="margin-top:8px;font-size:12px;color:var(--text-secondary);">${target.readOnly ? "Read-only target" : "Interactive target"}</div>
                                </div>
                                <div style="display:flex;flex-wrap:wrap;gap:8px;">
                                    <button type="button" data-target-open-tab="${esc(target.id)}" data-target-label="${esc(target.label || target.id)}" style="padding:8px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;cursor:pointer;color:inherit;">Connect</button>
                                </div>
                            </div>
                        `).join("")}
                    </div>
                ` : `
                    <div style="min-height:0;display:grid;place-items:center;justify-items:center;">
                        <div style="width:min(100%,540px);text-align:center;padding:24px 20px;border:1px dashed var(--border-color);border-radius:10px;background:transparent;font-size:13px;color:var(--text-secondary);line-height:1.5;display:grid;gap:6px;">
                            <div style="font-weight:600;color:var(--text-primary);">${esc(emptyState.title)}</div>
                            <div>${esc(emptyState.body)}</div>
                        </div>
                    </div>
                `}
            </div>
        `;
    this.directHostInputEl = this.bodyEl.querySelector("[data-vnc-direct-host]");
    this.directPortInputEl = this.bodyEl.querySelector("[data-vnc-direct-port]");
    this.directPasswordInputEl = this.bodyEl.querySelector("[data-vnc-direct-password]");
    const openDirectTarget = () => {
      const targetRef = buildDirectVncTargetReference(this.directHostInputEl?.value, this.directPortInputEl?.value);
      if (!targetRef) {
        return;
      }
      this.authPassword = normalizeVncPassword(this.directPasswordInputEl ? this.directPasswordInputEl.value : this.authPassword);
      this.openTargetTab(targetRef, targetRef);
    };
    this.directHostInputEl?.addEventListener("keydown", (event) => {
      if (event.key !== "Enter")
        return;
      event.preventDefault();
      openDirectTarget();
    });
    this.directPortInputEl?.addEventListener("keydown", (event) => {
      if (event.key !== "Enter")
        return;
      event.preventDefault();
      openDirectTarget();
    });
    this.directPasswordInputEl?.addEventListener("keydown", (event) => {
      if (event.key !== "Enter")
        return;
      event.preventDefault();
      openDirectTarget();
    });
    this.bodyEl.querySelector("[data-direct-open-tab]")?.addEventListener("click", () => openDirectTarget());
    for (const button of Array.from(this.bodyEl.querySelectorAll("[data-target-open-tab]"))) {
      button.addEventListener("click", () => {
        const targetId = button.getAttribute("data-target-open-tab");
        const label = button.getAttribute("data-target-label") || targetId || "VNC";
        if (!targetId)
          return;
        this.openTargetTab(targetId, label);
      });
    }
  }
  renderTargetSession(payload) {
    this.resetLiveSession();
    const target = payload?.target || {};
    const targetLabel = target?.label || this.targetId || "VNC target";
    const compactWindow = isPanePopoutMode();
    this.targetLabel = targetLabel;
    this.readOnly = Boolean(target.read_only);
    this.pointerButtonMask = 0;
    this.hasRenderedFrame = false;
    this.pressedKeysyms.clear();
    this.bodyEl.innerHTML = compactWindow ? `
                <div data-vnc-session-shell style="width:100%;height:100%;min-height:0;display:grid;grid-template-rows:auto minmax(0,1fr);gap:6px;">
                    <div data-vnc-session-chrome style="padding:6px 8px;border:1px solid var(--border-color);border-radius:8px;background:transparent;display:flex;flex-wrap:wrap;gap:8px;align-items:center;">
                        <div data-display-info style="min-width:0;flex:1 1 240px;font-size:12px;color:var(--text-secondary);line-height:1.3;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">Negotiating remote display…</div>
                        <input type="password" data-vnc-password placeholder="Password" autocomplete="current-password" style="width:150px;max-width:100%;padding:6px 8px;border:1px solid var(--border-color);border-radius:6px;background:transparent;color:inherit;" />
                        <button type="button" data-vnc-reconnect="1" style="padding:6px 10px;border:1px solid var(--border-color);border-radius:6px;background:transparent;cursor:pointer;color:inherit;">Reconnect</button>
                    </div>
                    <div data-display-stage style="min-height:0;height:100%;border:1px solid var(--border-color);border-radius:8px;background:#0a0a0a;display:flex;align-items:center;justify-content:center;padding:4px;position:relative;overflow:hidden;">
                        <canvas data-display-canvas tabindex="0" style="display:none;max-width:100%;max-height:100%;width:auto;height:auto;image-rendering:auto;box-shadow:none;border-radius:2px;background:#000;"></canvas>
                        <div data-display-placeholder style="max-width:520px;text-align:center;color:#d7d7d7;line-height:1.5;">
                            <div style="font-weight:600;font-size:14px;margin-bottom:6px;">${esc(targetLabel)}</div>
                            <div style="font-size:12px;color:#b7b7b7;">Waiting for the VNC/RFB handshake and first framebuffer update…</div>
                        </div>
                    </div>
                </div>
            ` : `
                <div data-vnc-session-shell style="width:100%;height:100%;min-height:0;display:grid;grid-template-rows:auto minmax(0,1fr);gap:12px;">
                    <div data-vnc-session-chrome style="padding:10px 12px;border:1px solid var(--border-color);border-radius:10px;background:transparent;display:grid;gap:10px;">
                        <div style="display:grid;gap:4px;min-width:0;">
                            <div style="font:12px var(--font-family-mono, monospace);color:var(--text-secondary);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${esc(target.id || this.targetId || "")} · ${target.read_only ? "read-only" : "interactive"} · websocket → TCP proxy</div>
                            <div data-display-info style="font-size:13px;color:var(--text-primary);line-height:1.4;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">Negotiating remote display…</div>
                            <div data-display-meta style="font:11px var(--font-family-mono, monospace);color:var(--text-secondary);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;"></div>
                        </div>
                        <div style="display:flex;flex-wrap:wrap;gap:8px;align-items:end;">
                            <label style="display:grid;gap:4px;min-width:160px;flex:1 1 180px;">
                                <span style="font-size:11px;color:var(--text-secondary);">VNC password</span>
                                <input type="password" data-vnc-password placeholder="Optional" autocomplete="current-password" style="width:100%;padding:8px 10px;border:1px solid var(--border-color);border-radius:8px;background:transparent;color:inherit;" />
                            </label>
                            <button type="button" data-vnc-reconnect="1" style="padding:8px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;cursor:pointer;color:inherit;">Reconnect</button>
                            <button type="button" data-open-target-picker="1" style="padding:8px 12px;border:1px solid var(--border-color);border-radius:8px;background:transparent;cursor:pointer;color:inherit;">Target</button>
                        </div>
                    </div>
                    <div data-display-stage style="min-height:0;height:100%;border:1px solid var(--border-color);border-radius:10px;background:#0a0a0a;display:flex;align-items:center;justify-content:center;padding:8px;position:relative;overflow:hidden;">
                        <canvas data-display-canvas tabindex="0" style="display:none;max-width:100%;max-height:100%;width:auto;height:auto;image-rendering:auto;box-shadow:none;border-radius:4px;background:#000;"></canvas>
                        <div data-display-placeholder style="max-width:520px;text-align:center;color:#d7d7d7;line-height:1.6;">
                            <div style="font-weight:700;font-size:18px;margin-bottom:8px;">${esc(targetLabel)}</div>
                            <div style="font-size:13px;color:#b7b7b7;">Waiting for the VNC/RFB handshake and first framebuffer update…</div>
                        </div>
                    </div>
                </div>
            `;
    this.sessionShellEl = this.bodyEl.querySelector("[data-vnc-session-shell]");
    this.chromeEl = this.bodyEl.querySelector("[data-vnc-session-chrome]");
    this.displayStageEl = this.bodyEl.querySelector("[data-display-stage]");
    this.canvas = this.bodyEl.querySelector("[data-display-canvas]");
    this.displayPlaceholderEl = this.bodyEl.querySelector("[data-display-placeholder]");
    this.displayInfoEl = this.bodyEl.querySelector("[data-display-info]");
    this.displayMetaEl = this.bodyEl.querySelector("[data-display-meta]");
    this.canvasCtx = this.canvas?.getContext?.("2d", { alpha: false }) || null;
    if (this.canvasCtx) {
      this.canvasCtx.imageSmoothingEnabled = true;
      this.canvasCtx.imageSmoothingQuality = "high";
    }
    this.updateDisplayInfo("Waiting for VNC protocol negotiation…");
    this.updateDisplayMeta();
    this.setSessionChromeVisible(true);
    this.attachDisplayResizeObserver();
    this.attachCanvasPointerHandlers();
    this.attachCanvasKeyboardHandlers();
    this.passwordInputEl = this.bodyEl.querySelector("[data-vnc-password]");
    if (this.passwordInputEl && this.authPassword !== null) {
      this.passwordInputEl.value = this.authPassword;
    }
    this.passwordInputEl?.addEventListener("input", () => {
      this.authPassword = normalizeVncPassword(this.passwordInputEl.value);
    });
    this.passwordInputEl?.addEventListener("keydown", (event) => {
      if (event.key !== "Enter")
        return;
      event.preventDefault();
      this.connectSocket();
    });
    const reconnectBtn = this.bodyEl.querySelector("[data-vnc-reconnect]");
    reconnectBtn?.addEventListener("click", () => {
      this.authPassword = normalizeVncPassword(this.passwordInputEl ? this.passwordInputEl.value : this.authPassword);
      this.connectSocket();
    });
    const pickerBtn = this.bodyEl.querySelector("[data-open-target-picker]");
    pickerBtn?.addEventListener("click", () => {
      this.openTargetTab("", "VNC");
    });
  }
  updateDisplayInfo(message) {
    if (this.displayInfoEl) {
      this.displayInfoEl.textContent = String(message || "");
    }
  }
  updateDisplayMeta(extra = "") {
    if (!this.displayMetaEl)
      return;
    const protocolState = this.protocol?.state ? `state=${this.protocol.state}` : "state=idle";
    const size = this.protocol?.framebufferWidth && this.protocol?.framebufferHeight ? `${this.protocol.framebufferWidth}×${this.protocol.framebufferHeight}` : "pending";
    const name = this.protocol?.serverName ? ` · name=${this.protocol.serverName}` : "";
    const scale = this.displayScale ? ` · scale=${Math.round(this.displayScale * 100)}%` : "";
    const suffix = extra ? ` · ${extra}` : "";
    this.displayMetaEl.textContent = `${protocolState} · framebuffer=${size}${name}${scale}${suffix}`;
  }
  ensureCanvasSize(width, height, options = {}) {
    if (!this.canvas || !this.canvasCtx || !width || !height)
      return;
    if (this.canvas.width !== width || this.canvas.height !== height) {
      this.canvas.width = width;
      this.canvas.height = height;
    }
    const reveal = options?.reveal === true;
    this.canvas.style.display = reveal || this.hasRenderedFrame ? "block" : "none";
    this.canvas.style.aspectRatio = `${width} / ${height}`;
    if (this.displayPlaceholderEl) {
      this.displayPlaceholderEl.style.display = reveal || this.hasRenderedFrame ? "none" : "";
    }
    this.updateCanvasScale();
  }
  attachDisplayResizeObserver() {
    if (!this.displayStageEl || typeof ResizeObserver === "undefined")
      return;
    try {
      this.resizeObserver?.disconnect?.();
    } catch {}
    this.resizeObserver = new ResizeObserver(() => {
      this.updateCanvasScale();
    });
    this.resizeObserver.observe(this.displayStageEl);
  }
  updateCanvasScale() {
    if (!this.canvas || !this.displayStageEl || !this.canvas.width || !this.canvas.height)
      return;
    requestAnimationFrame(() => {
      if (!this.canvas || !this.displayStageEl)
        return;
      const bounds = this.displayStageEl.getBoundingClientRect?.();
      const availableWidth = Math.max(1, Math.floor(bounds?.width || this.displayStageEl.clientWidth || 0) - 32);
      const availableHeight = Math.max(1, Math.floor(bounds?.height || this.displayStageEl.clientHeight || 0) - 32);
      if (!availableWidth || !availableHeight)
        return;
      const scale = computeContainedRemoteDisplayScale(availableWidth, availableHeight, this.canvas.width, this.canvas.height);
      this.displayScale = scale;
      this.canvas.style.width = `${Math.max(1, Math.round(this.canvas.width * scale))}px`;
      this.canvas.style.height = `${Math.max(1, Math.round(this.canvas.height * scale))}px`;
      this.updateDisplayMeta();
    });
  }
  getFramebufferPointFromEvent(event) {
    if (!this.canvas || !this.protocol?.framebufferWidth || !this.protocol?.framebufferHeight)
      return null;
    const rect = this.canvas.getBoundingClientRect?.();
    if (!rect || !rect.width || !rect.height)
      return null;
    return mapClientToFramebufferPoint(event.clientX, event.clientY, rect, this.protocol.framebufferWidth, this.protocol.framebufferHeight);
  }
  sendPointerEvent(buttonMask, x3, y2) {
    if (!this.socketBoundary || !this.protocol || this.protocol.state !== "connected")
      return;
    this.socketBoundary.send(encodeVncPointerEvent(buttonMask, x3, y2));
  }
  attachCanvasPointerHandlers() {
    if (!this.canvas || this.readOnly)
      return;
    this.canvas.style.cursor = "crosshair";
    this.canvas.style.touchAction = "none";
    try {
      this.pointerInputAbortController?.abort?.();
    } catch {}
    const abortController = new AbortController;
    this.pointerInputAbortController = abortController;
    const signal = abortController.signal;
    const ownerDocument = this.canvas.ownerDocument || document;
    const ownerWindow = ownerDocument.defaultView || window;
    const pressedMaskByPointer = new Map;
    const lastPointByPointer = new Map;
    const idleReleaseTimerByPointer = new Map;
    const resolvePoint = (event) => this.getFramebufferPointFromEvent(event) || lastPointByPointer.get(event?.pointerId) || { x: 0, y: 0 };
    const resolveTouchPoint = (touch) => {
      if (!touch || !this.canvas || !this.protocol?.framebufferWidth || !this.protocol?.framebufferHeight)
        return null;
      const rect = this.canvas.getBoundingClientRect?.();
      if (!rect || !rect.width || !rect.height)
        return null;
      return mapClientToFramebufferPoint(touch.clientX, touch.clientY, rect, this.protocol.framebufferWidth, this.protocol.framebufferHeight);
    };
    const clearIdleReleaseTimer = (pointerId) => {
      const handle = idleReleaseTimerByPointer.get(pointerId);
      if (handle) {
        ownerWindow.clearTimeout(handle);
        idleReleaseTimerByPointer.delete(pointerId);
      }
    };
    const armIdleReleaseTimer = (event, delayMs = 80) => {
      const pointerType = String(event?.pointerType || "").toLowerCase();
      if (!shouldArmVncImplicitReleaseTimer(pointerType))
        return;
      const pointerId = Number(event?.pointerId);
      if (!Number.isFinite(pointerId))
        return;
      clearIdleReleaseTimer(pointerId);
      const handle = ownerWindow.setTimeout(() => {
        idleReleaseTimerByPointer.delete(pointerId);
        if (!pressedMaskByPointer.has(pointerId) && !this.pointerButtonMask)
          return;
        releasePointer({
          pointerId,
          pointerType,
          type: "pointercancel",
          clientX: event?.clientX,
          clientY: event?.clientY
        }, { resetAll: true });
      }, delayMs);
      idleReleaseTimerByPointer.set(pointerId, handle);
    };
    const releaseAllPointers = (point = null) => {
      if (!pressedMaskByPointer.size && !this.pointerButtonMask)
        return;
      for (const pointerId of idleReleaseTimerByPointer.keys()) {
        clearIdleReleaseTimer(pointerId);
      }
      const fallbackPoint = point || lastPointByPointer.values().next().value || { x: 0, y: 0 };
      pressedMaskByPointer.clear();
      lastPointByPointer.clear();
      this.pointerButtonMask = 0;
      this.sendPointerEvent(0, fallbackPoint.x, fallbackPoint.y);
    };
    const releasePointer = (event, options = {}) => {
      if (options.resetAll) {
        const pointerId2 = Number(event?.pointerId);
        clearIdleReleaseTimer(pointerId2);
        releaseAllPointers(resolvePoint(event));
        try {
          this.canvas?.releasePointerCapture?.(pointerId2);
        } catch {}
        return;
      }
      const point = resolvePoint(event);
      const pointerId = Number(event?.pointerId);
      clearIdleReleaseTimer(pointerId);
      const hadPressedMask = pressedMaskByPointer.has(pointerId);
      const pressedBit = pressedMaskByPointer.get(pointerId) ?? resolveVncPointerPressMask(event);
      if (!hadPressedMask && !pressedBit && !this.pointerButtonMask)
        return;
      pressedMaskByPointer.delete(pointerId);
      lastPointByPointer.delete(pointerId);
      if (pressedBit) {
        this.pointerButtonMask &= ~pressedBit;
      } else if (!pressedMaskByPointer.size) {
        this.pointerButtonMask = 0;
      }
      this.sendPointerEvent(this.pointerButtonMask, point.x, point.y);
      try {
        this.canvas?.releasePointerCapture?.(pointerId);
      } catch {}
    };
    this.canvas.addEventListener("contextmenu", (event) => {
      event.preventDefault();
    }, { signal });
    this.canvas.addEventListener("pointermove", (event) => {
      const point = this.getFramebufferPointFromEvent(event);
      if (!point)
        return;
      lastPointByPointer.set(event.pointerId, point);
      if (pressedMaskByPointer.has(event.pointerId) && shouldReleaseVncPointerContact(event)) {
        releasePointer(event, { resetAll: true });
        return;
      }
      if (this.pointerButtonMask && !pressedMaskByPointer.has(event.pointerId) && shouldReleaseVncPointerContact(event)) {
        releaseAllPointers(point);
        return;
      }
      if (pressedMaskByPointer.has(event.pointerId)) {
        armIdleReleaseTimer(event);
      }
      this.sendPointerEvent(this.pointerButtonMask, point.x, point.y);
    }, { signal });
    this.canvas.addEventListener("pointerdown", (event) => {
      const point = this.getFramebufferPointFromEvent(event);
      if (!point)
        return;
      event.preventDefault();
      this.canvas?.focus?.();
      lastPointByPointer.set(event.pointerId, point);
      if (String(event?.pointerType || "").toLowerCase() === "mouse") {
        try {
          this.canvas?.setPointerCapture?.(event.pointerId);
        } catch {}
      }
      const bit = resolveVncPointerPressMask(event);
      if (!bit)
        return;
      pressedMaskByPointer.set(event.pointerId, (pressedMaskByPointer.get(event.pointerId) ?? 0) | bit);
      this.pointerButtonMask |= bit;
      armIdleReleaseTimer(event);
      this.sendPointerEvent(this.pointerButtonMask, point.x, point.y);
    }, { signal, passive: false });
    this.canvas.addEventListener("pointerup", (event) => {
      event.preventDefault();
      releasePointer(event);
    }, { signal, passive: false });
    this.canvas.addEventListener("pointercancel", (event) => {
      event.preventDefault();
      releasePointer(event, { resetAll: true });
    }, { signal, passive: false });
    this.canvas.addEventListener("pointerleave", (event) => {
      if (!pressedMaskByPointer.has(event.pointerId))
        return;
      if (!shouldReleaseVncPointerContact(event))
        return;
      releasePointer(event, { resetAll: true });
    }, { signal });
    this.canvas.addEventListener("pointerout", (event) => {
      if (!pressedMaskByPointer.has(event.pointerId))
        return;
      if (!shouldReleaseVncPointerContact(event))
        return;
      releasePointer(event, { resetAll: true });
    }, { signal });
    this.canvas.addEventListener("lostpointercapture", (event) => {
      releasePointer(event, { resetAll: true });
    }, { signal });
    ownerWindow.addEventListener("pointermove", (event) => {
      if (!pressedMaskByPointer.size && !this.pointerButtonMask || !shouldReleaseVncPointerContact(event))
        return;
      if (!pressedMaskByPointer.has(event.pointerId) && !this.pointerButtonMask)
        return;
      releasePointer(event, { resetAll: true });
    }, { signal });
    ownerWindow.addEventListener("pointerup", (event) => {
      if (!pressedMaskByPointer.has(event.pointerId) && !this.pointerButtonMask)
        return;
      event.preventDefault?.();
      releasePointer(event, { resetAll: !pressedMaskByPointer.has(event.pointerId) });
    }, { signal, passive: false });
    ownerWindow.addEventListener("pointercancel", (event) => {
      if (!pressedMaskByPointer.has(event.pointerId) && !this.pointerButtonMask)
        return;
      event.preventDefault?.();
      releasePointer(event, { resetAll: true });
    }, { signal, passive: false });
    const releaseFromTouchEvent = (event) => {
      if (!pressedMaskByPointer.size && !this.pointerButtonMask)
        return;
      if (!shouldReleaseVncTouchContact(event))
        return;
      const changedTouch = event?.changedTouches?.[0] || event?.touches?.[0] || null;
      const point = resolveTouchPoint(changedTouch) || lastPointByPointer.values().next().value || { x: 0, y: 0 };
      releaseAllPointers(point);
    };
    const releaseFromWindowPointerEvent = (event, options = {}) => {
      if (!pressedMaskByPointer.size && !this.pointerButtonMask)
        return;
      if (!shouldReleaseVncPointerContact(event))
        return;
      event?.preventDefault?.();
      releasePointer(event, {
        resetAll: options.resetAll === true || !pressedMaskByPointer.has(event?.pointerId)
      });
    };
    this.canvas.addEventListener("touchend", releaseFromTouchEvent, { signal, passive: true, capture: true });
    this.canvas.addEventListener("touchcancel", releaseFromTouchEvent, { signal, passive: true, capture: true });
    ownerDocument.addEventListener("touchend", releaseFromTouchEvent, { signal, passive: true, capture: true });
    ownerDocument.addEventListener("touchcancel", releaseFromTouchEvent, { signal, passive: true, capture: true });
    ownerWindow.addEventListener("touchend", releaseFromTouchEvent, { signal, passive: true, capture: true });
    ownerWindow.addEventListener("touchcancel", releaseFromTouchEvent, { signal, passive: true, capture: true });
    ownerDocument.addEventListener("pointerup", (event) => {
      releaseFromWindowPointerEvent(event);
    }, { signal, passive: false, capture: true });
    ownerDocument.addEventListener("pointercancel", (event) => {
      releaseFromWindowPointerEvent(event, { resetAll: true });
    }, { signal, passive: false, capture: true });
    ownerWindow.addEventListener("mouseup", () => {
      if (!pressedMaskByPointer.size && !this.pointerButtonMask)
        return;
      releaseAllPointers();
    }, { signal });
    ownerWindow.addEventListener("blur", () => {
      if (!pressedMaskByPointer.size && !this.pointerButtonMask)
        return;
      releaseAllPointers();
    }, { signal });
    ownerDocument.addEventListener("visibilitychange", () => {
      if (ownerDocument.visibilityState === "hidden") {
        releaseAllPointers();
      }
    }, { signal });
    this.canvas.addEventListener("wheel", (event) => {
      const point = this.getFramebufferPointFromEvent(event);
      if (!point)
        return;
      event.preventDefault();
      for (const payload of buildVncWheelPointerEvents(event.deltaY, point.x, point.y, this.pointerButtonMask)) {
        this.socketBoundary?.send?.(payload);
      }
    }, { signal, passive: false });
  }
  sendKeyEvent(down, keysym) {
    if (!this.socketBoundary || !this.protocol || this.protocol.state !== "connected")
      return;
    this.socketBoundary.send(encodeVncKeyEvent(down, keysym));
  }
  releasePressedKeys() {
    for (const keysym of this.pressedKeysyms.values()) {
      this.sendKeyEvent(false, keysym);
    }
    this.pressedKeysyms.clear();
  }
  attachCanvasKeyboardHandlers() {
    if (!this.canvas || this.readOnly)
      return;
    this.canvas.addEventListener("keydown", (event) => {
      const keysym = resolveVncKeysymFromKeyboardEvent(event);
      if (keysym == null)
        return;
      if (event.repeat && this.pressedKeysyms.has(event.code || event.key)) {
        event.preventDefault();
        return;
      }
      event.preventDefault();
      const keyId = event.code || event.key;
      this.pressedKeysyms.set(keyId, keysym);
      this.sendKeyEvent(true, keysym);
    });
    this.canvas.addEventListener("keyup", (event) => {
      const keyId = event.code || event.key;
      const keysym = this.pressedKeysyms.get(keyId) ?? resolveVncKeysymFromKeyboardEvent(event);
      if (keysym == null)
        return;
      event.preventDefault();
      this.pressedKeysyms.delete(keyId);
      this.sendKeyEvent(false, keysym);
    });
    this.canvas.addEventListener("blur", () => {
      this.releasePressedKeys();
    });
  }
  drawRgbaRect(rect) {
    if (!this.canvasCtx || !this.canvas)
      return;
    this.ensureCanvasSize(this.canvas.width || rect.width, this.canvas.height || rect.height, { reveal: true });
    const imageData = new ImageData(rect.rgba, rect.width, rect.height);
    this.canvasCtx.putImageData(imageData, rect.x, rect.y);
    this.hasRenderedFrame = true;
  }
  copyCanvasRect(rect) {
    if (!this.canvasCtx || !this.canvas)
      return;
    this.ensureCanvasSize(this.canvas.width || rect.width, this.canvas.height || rect.height, { reveal: true });
    const imageData = this.canvasCtx.getImageData(rect.srcX, rect.srcY, rect.width, rect.height);
    this.canvasCtx.putImageData(imageData, rect.x, rect.y);
    this.hasRenderedFrame = true;
  }
  scheduleRawFallbackTimeout() {
    if (this.frameTimeoutId) {
      clearTimeout(this.frameTimeoutId);
      this.frameTimeoutId = null;
    }
    if (this.rawFallbackAttempted || this.protocolRecovering)
      return;
    this.frameTimeoutId = setTimeout(() => {
      if (this.hasRenderedFrame || this.rawFallbackAttempted || this.protocolRecovering)
        return;
      if (this.protocol && this.socketBoundary) {
        this.rawFallbackAttempted = true;
        this.protocolRecovering = true;
        this.setStatus("No framebuffer update yet; retrying with RAW encoding.");
        this.updateDisplayInfo("No framebuffer update yet. Retrying with RAW encoding.");
        this.updateDisplayMeta("reconnect-encoding-fallback");
        this.connectWithEncodings("0");
      }
    }, 2200);
  }
  applyRemoteDisplayEvent(event) {
    if (!event)
      return;
    switch (event.type) {
      case "protocol-version":
        this.setStatus(`Negotiated ${event.protocol.toUpperCase()} ${event.server} → ${event.client}.`);
        this.updateDisplayInfo(`Negotiated ${event.server} → ${event.client}.`);
        this.updateDisplayMeta();
        return;
      case "security-types":
        this.setStatus(`Server offered security types: ${event.types.join(", ") || "none"}.`);
        this.updateDisplayInfo(`Security types: ${event.types.join(", ") || "none"}.`);
        this.updateDisplayMeta();
        return;
      case "security-selected":
        this.setStatus(`Using ${event.protocol.toUpperCase()} security type ${event.label}.`);
        this.updateDisplayInfo(`Security: ${event.label}.`);
        this.updateDisplayMeta();
        return;
      case "security-result":
        this.setStatus("Security negotiation complete. Waiting for server init…");
        this.updateDisplayInfo("Security negotiation complete. Waiting for server init…");
        this.updateDisplayMeta();
        return;
      case "display-init":
        this.ensureCanvasSize(event.width, event.height);
        this.setSessionChromeVisible(false);
        this.setStatus(`Connected to ${this.targetLabel || this.targetId || "target"} — waiting for first framebuffer update (${event.width}×${event.height}).`);
        this.updateDisplayInfo(`Connected to ${event.name || this.targetLabel || this.targetId || "remote display"}. Waiting for first framebuffer update…`);
        this.updateDisplayMeta("awaiting-frame");
        this.scheduleRawFallbackTimeout();
        return;
      case "framebuffer-update":
        if (this.frameTimeoutId) {
          clearTimeout(this.frameTimeoutId);
          this.frameTimeoutId = null;
        }
        let painted = false;
        const hasPipelineRects = (event.rects || []).some((r2) => r2.kind === "pipeline");
        if (event.framebuffer && event.framebuffer.length > 0 && event.width > 0 && event.height > 0 && hasPipelineRects) {
          this.ensureCanvasSize(event.width, event.height, { reveal: true });
          for (const rect of event.rects || []) {
            if (rect.kind === "resize") {
              this.ensureCanvasSize(rect.width, rect.height);
            }
          }
          const ctx = this.canvas?.getContext("2d", { alpha: false });
          if (ctx) {
            const img = new ImageData(new Uint8ClampedArray(event.framebuffer), event.width, event.height);
            ctx.putImageData(img, 0, 0);
            painted = true;
          }
        } else {
          for (const rect of event.rects || []) {
            if (rect.kind === "resize") {
              this.ensureCanvasSize(rect.width, rect.height);
              continue;
            }
            if (rect.kind === "copy") {
              this.ensureCanvasSize(event.width, event.height, { reveal: true });
              this.copyCanvasRect(rect);
              painted = true;
              continue;
            }
            if (rect.kind === "rgba") {
              this.ensureCanvasSize(event.width, event.height, { reveal: true });
              this.drawRgbaRect(rect);
              painted = true;
            }
          }
        }
        if (painted || this.hasRenderedFrame) {
          this.protocolRecovering = false;
          this.setStatus(`Rendering live framebuffer — ${event.width}×${event.height}.`);
          this.updateDisplayInfo(`Framebuffer update applied (${(event.rects || []).length} rect${(event.rects || []).length === 1 ? "" : "s"}).`);
          this.updateDisplayMeta();
        } else {
          this.setStatus(`Connected to ${this.targetLabel || this.targetId || "target"} — waiting for painted framebuffer data.`);
          this.updateDisplayInfo(`Framebuffer update received, but no paintable rects yet (${(event.rects || []).length} rect${(event.rects || []).length === 1 ? "" : "s"}).`);
          this.updateDisplayMeta("awaiting-frame");
          this.scheduleRawFallbackTimeout();
        }
        return;
      case "clipboard":
        this.setStatus("Remote clipboard updated.");
        this.updateDisplayInfo(`Clipboard text received (${event.text.length} chars).`);
        this.updateDisplayMeta();
        return;
      case "bell":
        this.setStatus("Remote display bell received.");
        this.updateDisplayInfo("Remote display bell received.");
        this.updateDisplayMeta();
        return;
    }
  }
  async handleSocketMessage(message) {
    if (message?.kind === "control") {
      const payload = message.payload;
      if (payload?.type === "vnc.error") {
        this.setStatus(`Proxy error: ${payload.error || "Unknown error"}`);
        this.updateDisplayInfo(`Proxy error: ${payload.error || "Unknown error"}`);
        this.updateDisplayMeta("proxy-error");
        return;
      }
      if (payload?.type === "vnc.connected") {
        const label = payload?.target?.label || this.targetLabel || this.targetId;
        this.setStatus(`Connected to ${label}. Waiting for VNC/RFB data…`);
        this.updateDisplayInfo(`Connected to ${label}. Waiting for VNC handshake…`);
        this.updateDisplayMeta();
        return;
      }
      if (payload?.type === "pong") {
        return;
      }
      return;
    }
    const protocol = this.protocol || (this.protocol = new VncRemoteDisplayProtocol);
    try {
      const chunk = message.data instanceof Blob ? await message.data.arrayBuffer() : message.data;
      const result = protocol.receive(chunk);
      for (const outgoing of result.outgoing || []) {
        this.socketBoundary?.send?.(outgoing);
      }
      for (const event of result.events || []) {
        this.applyRemoteDisplayEvent(event);
      }
    } catch (error) {
      const message2 = error?.message || "Unknown error";
      this.setSessionChromeVisible(true);
      this.setStatus(`Display protocol error: ${message2}`);
      this.updateDisplayInfo(`Display protocol error: ${message2}`);
      this.updateDisplayMeta("protocol-error");
      if (this.frameTimeoutId) {
        clearTimeout(this.frameTimeoutId);
        this.frameTimeoutId = null;
      }
      if (!this.rawFallbackAttempted && !this.protocolRecovering && /unexpected eof|zlib|decompress|protocol|buffer|undefined|not an object|reading '0'/i.test(message2)) {
        this.rawFallbackAttempted = true;
        this.protocolRecovering = true;
        this.connectWithEncodings("0");
      }
    }
  }
  async connectSocket(preferredEncodings = null) {
    if (!this.targetId || this.disposed)
      return;
    this.clearReconnectTimer();
    if (this.protocolRecovering && preferredEncodings == null) {
      this.protocolRecovering = false;
    }
    try {
      this.socketBoundary?.dispose?.();
    } catch {}
    if (preferredEncodings == null) {
      this.rawFallbackAttempted = false;
      this.protocolRecovering = false;
    }
    const handoffToken = this.pendingHandoffToken || null;
    const selectedEncodings = preferredEncodings == null ? null : String(preferredEncodings).trim();
    const wasmDecoder = await loadRemoteDisplayWasmDecoder();
    const protocolOptions = {};
    if (wasmDecoder) {
      protocolOptions.pipeline = wasmDecoder;
      protocolOptions.decodeRawRect = (bytes, width, height, pixelFormat) => wasmDecoder.decodeRawRectToRgba(bytes, width, height, pixelFormat);
    }
    const normalizedPassword = normalizeVncPassword(this.authPassword);
    if (normalizedPassword !== null) {
      protocolOptions.password = normalizedPassword;
    }
    if (selectedEncodings) {
      protocolOptions.encodings = selectedEncodings;
    }
    const preserveRenderedFrame = Boolean(this.canvas && this.hasRenderedFrame);
    this.protocol = new VncRemoteDisplayProtocol(protocolOptions);
    this.hasRenderedFrame = preserveRenderedFrame;
    this.frameTimeoutId = null;
    if (this.canvas) {
      this.canvas.style.display = preserveRenderedFrame ? "block" : "none";
    }
    if (this.displayPlaceholderEl) {
      this.displayPlaceholderEl.style.display = preserveRenderedFrame ? "none" : "";
    }
    this.socketBoundary = new WebSocketRemoteDisplayBoundary({
      url: buildVncWebSocketUrl(this.targetId, handoffToken),
      binaryType: "arraybuffer",
      onOpen: () => {
        if (handoffToken && this.pendingHandoffToken === handoffToken) {
          this.pendingHandoffToken = null;
        }
        this.reconnectAttempts = 0;
        this.setStatus(`Connected to proxy for ${this.targetId}. Waiting for VNC/RFB data…`);
        this.updateDisplayInfo("WebSocket proxy connected. Waiting for handshake…");
        this.updateDisplayMeta();
        this.socketBoundary?.sendControl?.({ type: "ping" });
      },
      onMetrics: (metrics) => {
        this.applyMetrics(metrics);
      },
      onMessage: (message) => {
        this.handleSocketMessage(message);
      },
      onClose: () => {
        this.setSessionChromeVisible(true);
        if (this.frameTimeoutId) {
          clearTimeout(this.frameTimeoutId);
          this.frameTimeoutId = null;
        }
        if (this.disposed)
          return;
        if (shouldRetryVncPopoutWithoutHandoff({
          handoffToken,
          bytesIn: this.bytesIn,
          hasRenderedFrame: this.hasRenderedFrame,
          reconnectAttempts: this.reconnectAttempts
        })) {
          this.pendingHandoffToken = null;
          this.setStatus("Transferred VNC session was not ready yet. Retrying…");
          this.updateDisplayInfo("Transferred VNC session was not ready yet. Retrying without handoff…");
          this.updateDisplayMeta("handoff-retrying");
          this.scheduleReconnect(150);
          return;
        }
        const shouldReconnect = this.bytesIn > 0 || this.hasRenderedFrame || this.reconnectAttempts > 0;
        if (shouldReconnect) {
          this.setStatus("Remote display connection lost. Reconnecting…");
          this.updateDisplayInfo("Remote display transport closed. Attempting to reconnect…");
          this.updateDisplayMeta("reconnecting");
          this.scheduleReconnect();
          return;
        }
        this.setStatus(this.bytesIn > 0 ? `Proxy closed after receiving ${this.bytesIn} byte(s).` : "Proxy closed.");
        this.updateDisplayInfo(this.bytesIn > 0 ? "Remote display transport closed after receiving data." : "Remote display transport closed.");
        this.updateDisplayMeta("closed");
      },
      onError: () => {
        this.setSessionChromeVisible(true);
        if (shouldRetryVncPopoutWithoutHandoff({
          handoffToken,
          bytesIn: this.bytesIn,
          hasRenderedFrame: this.hasRenderedFrame,
          reconnectAttempts: this.reconnectAttempts
        })) {
          this.pendingHandoffToken = null;
          this.setStatus("Transferred VNC session was not ready yet. Retrying…");
          this.updateDisplayInfo("Transferred VNC session was not ready yet. Retrying without handoff…");
          this.updateDisplayMeta("handoff-retrying");
          this.scheduleReconnect(150);
          return;
        }
        const shouldReconnect = this.bytesIn > 0 || this.hasRenderedFrame || this.reconnectAttempts > 0;
        if (shouldReconnect) {
          this.setStatus("WebSocket proxy connection failed. Reconnecting…");
          this.updateDisplayInfo("WebSocket proxy connection failed. Attempting to reconnect…");
          this.updateDisplayMeta("socket-reconnecting");
          this.scheduleReconnect();
          return;
        }
        this.setStatus("WebSocket proxy connection failed.");
        this.updateDisplayInfo("WebSocket proxy connection failed.");
        this.updateDisplayMeta("socket-error");
      }
    });
    this.socketBoundary.connect();
  }
  connectWithEncodings(encodings) {
    return this.connectSocket(encodings);
  }
  async load() {
    this.setStatus("");
    try {
      const payload = await fetchVncSession(this.targetId);
      if (!payload?.enabled) {
        this.renderTargets(payload);
        this.setStatus("");
        return;
      }
      if (!this.targetId) {
        this.renderTargets(payload);
        this.setStatus("");
        return;
      }
      this.renderTargetSession(payload);
      await this.connectSocket();
    } catch (error) {
      this.resetLiveSession();
      this.bodyEl.innerHTML = `
                <div style="max-width:620px;text-align:center;padding:28px;border:1px dashed var(--border-color);border-radius:14px;background:var(--bg-secondary);">
                    <div style="font-size:32px;margin-bottom:10px;">⚠️</div>
                    <div style="font-weight:600;margin-bottom:6px;">Failed to load VNC session</div>
                    <div style="color:var(--text-secondary);font-size:13px;line-height:1.5;">${esc(error?.message || "Unknown error")}</div>
                </div>
            `;
      this.setStatus(`Session load failed: ${error?.message || "Unknown error"}`);
    }
  }
  beforeDetachFromHost() {
    this.releasePressedKeys();
    this.setStatus("Moving VNC session…");
    this.updateDisplayInfo("Moving VNC session to a new window…");
    this.updateDisplayMeta("moving");
  }
  afterAttachToHost() {
    this.attachDisplayResizeObserver();
    this.updateCanvasScale();
    requestAnimationFrame(() => this.focus());
  }
  moveHost(container) {
    if (this.disposed || !this.root)
      return false;
    this.releasePressedKeys();
    this.container = container;
    if (!relocateVncPaneRoot(this.root, container)) {
      return false;
    }
    this.afterAttachToHost();
    return true;
  }
  async preparePopoutTransfer() {
    return createVncPopoutTransferPayload(this.targetId, this.authPassword);
  }
  getContent() {
    return;
  }
  isDirty() {
    return false;
  }
  focus() {
    this.canvas?.focus?.();
    this.root?.focus?.();
  }
  resize() {
    this.updateCanvasScale();
  }
  dispose() {
    if (this.disposed)
      return;
    this.disposed = true;
    this.resetLiveSession();
    this.root?.remove?.();
  }
}
// web/src/panes/office-viewer-pane.ts
var OFFICE_EXTENSIONS = new Set([
  ".docx",
  ".doc",
  ".odt",
  ".rtf",
  ".xlsx",
  ".xls",
  ".ods",
  ".csv",
  ".pptx",
  ".ppt",
  ".odp"
]);
// web/src/panes/mindmap-pane.ts
var VENDOR_CACHE_BUST = String(Date.now());
// web/src/panes/kanban-pane.ts
var VENDOR_CACHE_BUST2 = String(Date.now());
// web/src/panes/tab-store.ts
class TabStoreImpl {
  tabs = new Map;
  activeId = null;
  mruOrder = [];
  listeners = new Set;
  onChange(listener) {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }
  notify() {
    const tabs = this.getTabs();
    const activeId = this.activeId;
    for (const listener of this.listeners) {
      try {
        listener(tabs, activeId);
      } catch (err2) {
        console.warn("[tab-store] Change listener failed:", err2);
      }
    }
  }
  open(path, label) {
    let tab = this.tabs.get(path);
    if (!tab) {
      tab = {
        id: path,
        label: label || path.split("/").pop() || path,
        path,
        dirty: false,
        pinned: false
      };
      this.tabs.set(path, tab);
    }
    this.activate(path);
    return tab;
  }
  activate(id) {
    if (!this.tabs.has(id))
      return;
    this.activeId = id;
    this.mruOrder = [id, ...this.mruOrder.filter((x3) => x3 !== id)];
    this.notify();
  }
  close(id) {
    const tab = this.tabs.get(id);
    if (!tab)
      return false;
    this.tabs.delete(id);
    this.mruOrder = this.mruOrder.filter((x3) => x3 !== id);
    if (this.activeId === id) {
      this.activeId = this.mruOrder[0] || null;
    }
    this.notify();
    return true;
  }
  closeOthers(keepId) {
    for (const [id, tab] of this.tabs) {
      if (id !== keepId && !tab.pinned) {
        this.tabs.delete(id);
        this.mruOrder = this.mruOrder.filter((x3) => x3 !== id);
      }
    }
    if (this.activeId && !this.tabs.has(this.activeId)) {
      this.activeId = keepId;
    }
    this.notify();
  }
  closeAll() {
    for (const [id, tab] of this.tabs) {
      if (!tab.pinned) {
        this.tabs.delete(id);
        this.mruOrder = this.mruOrder.filter((x3) => x3 !== id);
      }
    }
    if (this.activeId && !this.tabs.has(this.activeId)) {
      this.activeId = this.mruOrder[0] || null;
    }
    this.notify();
  }
  setDirty(id, dirty) {
    const tab = this.tabs.get(id);
    if (!tab || tab.dirty === dirty)
      return;
    tab.dirty = dirty;
    this.notify();
  }
  togglePin(id) {
    const tab = this.tabs.get(id);
    if (!tab)
      return;
    tab.pinned = !tab.pinned;
    this.notify();
  }
  saveViewState(id, viewState) {
    const tab = this.tabs.get(id);
    if (tab)
      tab.viewState = viewState;
  }
  getViewState(id) {
    return this.tabs.get(id)?.viewState;
  }
  rename(oldId, newPath, newLabel) {
    const tab = this.tabs.get(oldId);
    if (!tab)
      return;
    this.tabs.delete(oldId);
    tab.id = newPath;
    tab.path = newPath;
    tab.label = newLabel || newPath.split("/").pop() || newPath;
    this.tabs.set(newPath, tab);
    this.mruOrder = this.mruOrder.map((x3) => x3 === oldId ? newPath : x3);
    if (this.activeId === oldId)
      this.activeId = newPath;
    this.notify();
  }
  getTabs() {
    return Array.from(this.tabs.values());
  }
  getActiveId() {
    return this.activeId;
  }
  getActive() {
    return this.activeId ? this.tabs.get(this.activeId) || null : null;
  }
  get(id) {
    return this.tabs.get(id);
  }
  get size() {
    return this.tabs.size;
  }
  hasUnsaved() {
    for (const tab of this.tabs.values()) {
      if (tab.dirty)
        return true;
    }
    return false;
  }
  getDirtyTabs() {
    return Array.from(this.tabs.values()).filter((t2) => t2.dirty);
  }
  nextTab() {
    const tabs = this.getTabs();
    if (tabs.length <= 1)
      return;
    const idx = tabs.findIndex((t2) => t2.id === this.activeId);
    const next = tabs[(idx + 1) % tabs.length];
    this.activate(next.id);
  }
  prevTab() {
    const tabs = this.getTabs();
    if (tabs.length <= 1)
      return;
    const idx = tabs.findIndex((t2) => t2.id === this.activeId);
    const prev = tabs[(idx - 1 + tabs.length) % tabs.length];
    this.activate(prev.id);
  }
  mruSwitch() {
    if (this.mruOrder.length > 1) {
      this.activate(this.mruOrder[1]);
    }
  }
}
var tabStore = new TabStoreImpl;
// web/src/components/input-focus-safety.ts
function focusAndSelectBestEffort(input) {
  try {
    input?.focus?.();
    input?.select?.();
    return true;
  } catch (_error) {
    return false;
  }
}

// web/src/ui/workspace-scale.ts
var WORKSPACE_SCALE_STORAGE_KEY = "workspaceExplorerScale";
var WORKSPACE_SCALE_PRESETS = ["compact", "default", "comfortable"];
var WORKSPACE_SCALE_SET = new Set(WORKSPACE_SCALE_PRESETS);
var WORKSPACE_SCALE_METRICS = {
  compact: { indentPx: 14 },
  default: { indentPx: 16 },
  comfortable: { indentPx: 18 }
};
function normalizeWorkspaceScale(value, fallback = "default") {
  if (typeof value !== "string")
    return fallback;
  const normalized = value.trim().toLowerCase();
  return WORKSPACE_SCALE_SET.has(normalized) ? normalized : fallback;
}
function readWorkspaceScaleEnvironment() {
  if (typeof window === "undefined") {
    return { width: 0, isTouch: false };
  }
  const width = Number(window.innerWidth) || 0;
  const coarsePointer = Boolean(window.matchMedia?.("(pointer: coarse)")?.matches);
  const noHover = Boolean(window.matchMedia?.("(hover: none)")?.matches);
  const touchPoints = Number(globalThis.navigator?.maxTouchPoints || 0) > 0;
  return {
    width,
    isTouch: coarsePointer || touchPoints && noHover
  };
}
function getResponsiveWorkspaceScale(env = {}) {
  const width = Math.max(0, Number(env.width) || 0);
  const isTouch = Boolean(env.isTouch);
  if (isTouch)
    return "comfortable";
  if (width > 0 && width < 1180)
    return "comfortable";
  return "default";
}
function clampWorkspaceScale(scale, env = {}) {
  if (Boolean(env.isTouch) && scale === "compact")
    return "default";
  return scale;
}
function resolveWorkspaceScale(options = {}) {
  const responsive = getResponsiveWorkspaceScale(options);
  const requested = options.stored ? normalizeWorkspaceScale(options.stored, responsive) : responsive;
  return clampWorkspaceScale(requested, options);
}
function getWorkspaceScaleMetrics(scale) {
  return WORKSPACE_SCALE_METRICS[normalizeWorkspaceScale(scale)];
}

// web/src/ui/workspace-auto-open.ts
var MAX_EDITABLE_PREVIEW_BYTES = 256 * 1024;
function isEditableWorkspacePreview(preview) {
  if (!preview || preview.kind !== "text")
    return false;
  const size = Number(preview.size);
  return !Number.isFinite(size) || size <= MAX_EDITABLE_PREVIEW_BYTES;
}
function hasSpecializedWorkspaceTab(path, resolvePane) {
  const normalized = String(path || "").trim();
  if (!normalized || normalized.endsWith("/"))
    return false;
  if (typeof resolvePane !== "function")
    return false;
  const resolved = resolvePane({ path: normalized, mode: "edit" });
  if (!resolved || typeof resolved !== "object")
    return false;
  return resolved.id !== "editor";
}
function shouldAutoOpenWorkspaceFile(path, preview, options = {}) {
  const resolvePane = options.resolvePane;
  if (hasSpecializedWorkspaceTab(path, resolvePane))
    return true;
  return isEditableWorkspacePreview(preview);
}

// web/src/components/workspace-explorer.ts
var REFRESH_INTERVAL_MS = 60000;
var isHiddenNode = (node) => {
  if (!node || !node.name)
    return false;
  if (node.path === ".")
    return false;
  return node.name.startsWith(".");
};
function hasOpenableWorkspaceTab(path) {
  const normalized = String(path || "").trim();
  if (!normalized || normalized.endsWith("/"))
    return false;
  return hasSpecializedWorkspaceTab(normalized, (context) => paneRegistry.resolve(context));
}
function flattenTree(node, expanded, showHidden, depth = 0, rows = []) {
  if (!showHidden && isHiddenNode(node))
    return rows;
  if (!node)
    return rows;
  rows.push({ node, depth });
  if (node.type === "dir" && node.children && expanded.has(node.path)) {
    for (const child of node.children)
      flattenTree(child, expanded, showHidden, depth + 1, rows);
  }
  return rows;
}
function treeSignature(node, expanded, showHidden) {
  if (!node)
    return "";
  const parts = [];
  const walk = (item) => {
    if (!showHidden && isHiddenNode(item))
      return;
    parts.push(item.type === "dir" ? `d:${item.path}` : `f:${item.path}`);
    if (item.children && expanded?.has(item.path)) {
      for (const child of item.children)
        walk(child);
    }
  };
  walk(node);
  return parts.join("|");
}
function mergeTree(prev, next) {
  if (!next)
    return null;
  if (!prev)
    return next;
  if (prev.path !== next.path || prev.type !== next.type)
    return next;
  const prevKids = Array.isArray(prev.children) ? prev.children : null;
  const nextKids = Array.isArray(next.children) ? next.children : null;
  if (!nextKids)
    return prev;
  const prevMap = prevKids ? new Map(prevKids.map((c2) => [c2?.path, c2])) : new Map;
  let changed = !prevKids || prevKids.length !== nextKids.length;
  const merged = nextKids.map((child) => {
    const m2 = mergeTree(prevMap.get(child.path), child);
    if (m2 !== prevMap.get(child.path))
      changed = true;
    return m2;
  });
  return changed ? { ...next, children: merged } : prev;
}
function replaceNodeAtPath(node, targetPath, nextNode) {
  if (!node)
    return node;
  if (node.path === targetPath)
    return mergeTree(node, nextNode);
  if (!Array.isArray(node.children))
    return node;
  let changed = false;
  const children = node.children.map((child) => {
    const updated = replaceNodeAtPath(child, targetPath, nextNode);
    if (updated !== child)
      changed = true;
    return updated;
  });
  return changed ? { ...node, children } : node;
}
var STARBURST_MAX_DEPTH = 4;
var STARBURST_MAX_CHILDREN = 14;
var STARBURST_FETCH_DEPTH = 8;
var STARBURST_CACHE_LIMIT = 16;
function computeSubtreeBytes(node) {
  if (!node)
    return 0;
  if (node.type === "file") {
    const size = Math.max(0, Number(node.size) || 0);
    node.__bytes = size;
    return size;
  }
  const children = Array.isArray(node.children) ? node.children : [];
  let total = 0;
  for (const child of children)
    total += computeSubtreeBytes(child);
  node.__bytes = total;
  return total;
}
function buildFolderSizeHierarchy(node, depth = 0) {
  const size = Math.max(0, Number(node?.__bytes ?? node?.size ?? 0));
  const out = {
    name: node?.name || node?.path || ".",
    path: node?.path || ".",
    size,
    children: []
  };
  if (!node || node.type !== "dir" || depth >= STARBURST_MAX_DEPTH)
    return out;
  const children = Array.isArray(node.children) ? node.children : [];
  const entries = [];
  for (const child of children) {
    const childSize = Math.max(0, Number(child?.__bytes ?? child?.size ?? 0));
    if (childSize <= 0)
      continue;
    if (child.type === "dir") {
      entries.push({ kind: "dir", node: child, size: childSize });
    } else {
      entries.push({ kind: "file", name: child.name, path: child.path, size: childSize });
    }
  }
  entries.sort((a2, b) => b.size - a2.size);
  let trimmed = entries;
  if (entries.length > STARBURST_MAX_CHILDREN) {
    const head = entries.slice(0, STARBURST_MAX_CHILDREN - 1);
    const tail = entries.slice(STARBURST_MAX_CHILDREN - 1);
    const tailSize = tail.reduce((acc, entry) => acc + entry.size, 0);
    head.push({
      kind: "other",
      name: `+${tail.length} more`,
      path: `${out.path}/[other]`,
      size: tailSize
    });
    trimmed = head;
  }
  out.children = trimmed.map((entry) => {
    if (entry.kind === "dir")
      return buildFolderSizeHierarchy(entry.node, depth + 1);
    return { name: entry.name, path: entry.path, size: entry.size, children: [] };
  });
  return out;
}
function detectDarkTheme2() {
  if (typeof window === "undefined" || typeof document === "undefined")
    return false;
  const root = document.documentElement;
  const body = document.body;
  const rootTheme = root?.getAttribute?.("data-theme")?.toLowerCase?.() || "";
  if (rootTheme === "dark")
    return true;
  if (rootTheme === "light")
    return false;
  if (root?.classList?.contains("dark") || body?.classList?.contains("dark"))
    return true;
  if (root?.classList?.contains("light") || body?.classList?.contains("light"))
    return false;
  return Boolean(window.matchMedia?.("(prefers-color-scheme: dark)")?.matches);
}
function segmentColorFromAngle(startAngle, depth, isDarkTheme) {
  const hue = ((startAngle + Math.PI / 2) * 180 / Math.PI + 360) % 360;
  const sat = isDarkTheme ? Math.max(30, 70 - depth * 10) : Math.max(34, 66 - depth * 8);
  const light = isDarkTheme ? Math.min(70, 45 + depth * 5) : Math.min(60, 42 + depth * 4);
  return `hsl(${hue.toFixed(1)} ${sat}% ${light}%)`;
}
function polar(cx, cy, radius, angle) {
  return {
    x: cx + radius * Math.cos(angle),
    y: cy + radius * Math.sin(angle)
  };
}
function describeDonutSegment(cx, cy, innerRadius, outerRadius, startAngle, endAngle) {
  const maxSweep = Math.PI * 2 - 0.0001;
  const clampedEnd = endAngle - startAngle > maxSweep ? startAngle + maxSweep : endAngle;
  const outerStart = polar(cx, cy, outerRadius, startAngle);
  const outerEnd = polar(cx, cy, outerRadius, clampedEnd);
  const innerEnd = polar(cx, cy, innerRadius, clampedEnd);
  const innerStart = polar(cx, cy, innerRadius, startAngle);
  const largeArc = clampedEnd - startAngle > Math.PI ? 1 : 0;
  return [
    `M ${outerStart.x.toFixed(3)} ${outerStart.y.toFixed(3)}`,
    `A ${outerRadius} ${outerRadius} 0 ${largeArc} 1 ${outerEnd.x.toFixed(3)} ${outerEnd.y.toFixed(3)}`,
    `L ${innerEnd.x.toFixed(3)} ${innerEnd.y.toFixed(3)}`,
    `A ${innerRadius} ${innerRadius} 0 ${largeArc} 0 ${innerStart.x.toFixed(3)} ${innerStart.y.toFixed(3)}`,
    "Z"
  ].join(" ");
}
var STARBURST_RINGS = {
  1: [26, 46],
  2: [48, 68],
  3: [70, 90],
  4: [92, 112]
};
function buildStarburstSegments(rootNode, baseSize, isDarkTheme) {
  const segments = [];
  const legend = [];
  const baseTotal = Math.max(0, Number(baseSize) || 0);
  const walk = (node, start, end, depth) => {
    const children = Array.isArray(node?.children) ? node.children : [];
    if (!children.length)
      return;
    const nodeSize = Math.max(0, Number(node.size) || 0);
    if (nodeSize <= 0)
      return;
    const span = end - start;
    let cursor = start;
    children.forEach((child, index) => {
      const childSize = Math.max(0, Number(child.size) || 0);
      if (childSize <= 0)
        return;
      const ratio = childSize / nodeSize;
      const childStart = cursor;
      const childEnd = index === children.length - 1 ? end : cursor + span * ratio;
      cursor = childEnd;
      if (childEnd - childStart < 0.003)
        return;
      const ring = STARBURST_RINGS[depth];
      if (ring) {
        const color = segmentColorFromAngle(childStart, depth, isDarkTheme);
        segments.push({
          key: child.path,
          path: child.path,
          label: child.name,
          size: childSize,
          color,
          depth,
          startAngle: childStart,
          endAngle: childEnd,
          innerRadius: ring[0],
          outerRadius: ring[1],
          d: describeDonutSegment(120, 120, ring[0], ring[1], childStart, childEnd)
        });
        if (depth === 1) {
          legend.push({
            key: child.path,
            name: child.name,
            size: childSize,
            pct: baseTotal > 0 ? childSize / baseTotal * 100 : 0,
            color
          });
        }
      }
      if (depth < STARBURST_MAX_DEPTH) {
        walk(child, childStart, childEnd, depth + 1);
      }
    });
  };
  walk(rootNode, -Math.PI / 2, Math.PI * 3 / 2, 1);
  return { segments, legend };
}
function findHierarchyNode(root, targetPath) {
  if (!root || !targetPath)
    return null;
  if (root.path === targetPath)
    return root;
  const children = Array.isArray(root.children) ? root.children : [];
  for (const child of children) {
    const found = findHierarchyNode(child, targetPath);
    if (found)
      return found;
  }
  return null;
}
function buildFallbackStarburst(label, pathBase, size, isDarkTheme) {
  if (!size || size <= 0)
    return { segments: [], legend: [] };
  const ring = STARBURST_RINGS[1];
  if (!ring)
    return { segments: [], legend: [] };
  const start = -Math.PI / 2;
  const end = Math.PI * 3 / 2;
  const color = segmentColorFromAngle(start, 1, isDarkTheme);
  const keyBase = pathBase || ".";
  const key = `${keyBase}/[files]`;
  return {
    segments: [
      {
        key,
        path: key,
        label,
        size,
        color,
        depth: 1,
        startAngle: start,
        endAngle: end,
        innerRadius: ring[0],
        outerRadius: ring[1],
        d: describeDonutSegment(120, 120, ring[0], ring[1], start, end)
      }
    ],
    legend: [
      {
        key,
        name: label,
        size,
        pct: 100,
        color
      }
    ]
  };
}
function createFolderStarburstPayload(root, truncated = false, isDarkTheme = false) {
  if (!root)
    return null;
  const totalSize = computeSubtreeBytes(root);
  const hierarchy = buildFolderSizeHierarchy(root, 0);
  const baseSize = hierarchy.size || totalSize;
  let { segments, legend } = buildStarburstSegments(hierarchy, baseSize, isDarkTheme);
  if (!segments.length && baseSize > 0) {
    const fallback = buildFallbackStarburst("[files]", hierarchy.path, baseSize, isDarkTheme);
    segments = fallback.segments;
    legend = fallback.legend;
  }
  return {
    root: hierarchy,
    totalSize: baseSize,
    segments,
    legend,
    truncated,
    isDarkTheme
  };
}
function FolderStarburstChart({ payload }) {
  if (!payload)
    return null;
  const [hovered, setHovered] = w0(null);
  const [zoomPath, setZoomPath] = w0(payload?.root?.path || ".");
  const [zoomStack, setZoomStack] = w0(() => [payload?.root?.path || "."]);
  const [isZooming, setIsZooming] = w0(false);
  r0(() => {
    const rootPath = payload?.root?.path || ".";
    setZoomPath(rootPath);
    setZoomStack([rootPath]);
    setHovered(null);
  }, [payload?.root?.path, payload?.totalSize]);
  r0(() => {
    if (!zoomPath)
      return;
    setIsZooming(true);
    const timer = setTimeout(() => setIsZooming(false), 180);
    return () => clearTimeout(timer);
  }, [zoomPath]);
  const zoomRoot = G0(() => {
    return findHierarchyNode(payload.root, zoomPath) || payload.root;
  }, [payload?.root, zoomPath]);
  const baseSize = zoomRoot?.size || payload.totalSize || 0;
  const { segments, legend } = G0(() => {
    const computed = buildStarburstSegments(zoomRoot, baseSize, payload.isDarkTheme);
    if (computed.segments.length > 0)
      return computed;
    if (baseSize <= 0)
      return computed;
    const label = zoomRoot?.children?.length ? "Total" : "[files]";
    return buildFallbackStarburst(label, zoomRoot?.path || payload?.root?.path || ".", baseSize, payload.isDarkTheme);
  }, [zoomRoot, baseSize, payload.isDarkTheme, payload?.root?.path]);
  const [animatedSegments, setAnimatedSegments] = w0(segments);
  const prevSegmentsRef = o0(new Map);
  const animFrameRef = o0(0);
  r0(() => {
    const prevMap = prevSegmentsRef.current;
    const nextMap = new Map(segments.map((segment) => [segment.key, segment]));
    const start = performance.now();
    const duration = 220;
    const animate = (now) => {
      const t2 = Math.min(1, (now - start) / duration);
      const eased = t2 * (2 - t2);
      const interpolated = segments.map((segment) => {
        const prev = prevMap.get(segment.key);
        const from = prev || {
          startAngle: segment.startAngle,
          endAngle: segment.startAngle,
          innerRadius: segment.innerRadius,
          outerRadius: segment.innerRadius
        };
        const lerp = (a2, b) => a2 + (b - a2) * eased;
        const startAngle = lerp(from.startAngle, segment.startAngle);
        const endAngle = lerp(from.endAngle, segment.endAngle);
        const innerRadius = lerp(from.innerRadius, segment.innerRadius);
        const outerRadius = lerp(from.outerRadius, segment.outerRadius);
        return {
          ...segment,
          d: describeDonutSegment(120, 120, innerRadius, outerRadius, startAngle, endAngle)
        };
      });
      setAnimatedSegments(interpolated);
      if (t2 < 1) {
        animFrameRef.current = requestAnimationFrame(animate);
      }
    };
    if (animFrameRef.current)
      cancelAnimationFrame(animFrameRef.current);
    animFrameRef.current = requestAnimationFrame(animate);
    prevSegmentsRef.current = nextMap;
    return () => {
      if (animFrameRef.current)
        cancelAnimationFrame(animFrameRef.current);
    };
  }, [segments]);
  const displaySegments = animatedSegments.length ? animatedSegments : segments;
  const totalLabel = baseSize > 0 ? formatFileSize(baseSize) : "0 B";
  const rawLabel = zoomRoot?.name || "";
  const labelBase = rawLabel && rawLabel !== "." ? rawLabel : "Total";
  const activeLabel = labelBase || "Total";
  const activeValue = totalLabel;
  const canZoomOut = zoomStack.length > 1;
  const handleSegmentClick = (segment) => {
    if (!segment?.path)
      return;
    const node = findHierarchyNode(payload.root, segment.path);
    if (!node || !Array.isArray(node.children) || node.children.length === 0)
      return;
    setZoomStack((prev) => [...prev, node.path]);
    setZoomPath(node.path);
    setHovered(null);
  };
  const handleZoomOut = () => {
    if (!canZoomOut)
      return;
    setZoomStack((prev) => {
      const next = prev.slice(0, -1);
      setZoomPath(next[next.length - 1] || payload?.root?.path || ".");
      return next;
    });
    setHovered(null);
  };
  return X1`
        <div class="workspace-folder-starburst">
            <svg viewBox="0 0 240 240" class=${`workspace-folder-starburst-svg${isZooming ? " is-zooming" : ""}`} role="img"
                aria-label=${`Folder sizes for ${zoomRoot?.path || payload?.root?.path || "."}`}
                data-segments=${displaySegments.length}
                data-base-size=${baseSize}>
                ${displaySegments.map((segment) => X1`
                    <path
                        key=${segment.key}
                        d=${segment.d}
                        fill=${segment.color}
                        stroke="var(--bg-primary)"
                        stroke-width="1"
                        class=${`workspace-folder-starburst-segment${hovered?.key === segment.key ? " is-hovered" : ""}`}
                        onMouseEnter=${() => setHovered(segment)}
                        onMouseLeave=${() => setHovered(null)}
                        onTouchStart=${() => setHovered(segment)}
                        onTouchEnd=${() => setHovered(null)}
                        onClick=${() => handleSegmentClick(segment)}
                    >
                        <title>${segment.label} — ${formatFileSize(segment.size)}</title>
                    </path>
                `)}
                <g
                    class=${`workspace-folder-starburst-center-hit${canZoomOut ? " is-drill" : ""}`}
                    onClick=${handleZoomOut}
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
                    <text x="120" y="114" text-anchor="middle" class="workspace-folder-starburst-total-label">${activeLabel}</text>
                    <text x="120" y="130" text-anchor="middle" class="workspace-folder-starburst-total-value">${activeValue}</text>
                </g>
            </svg>
            ${legend.length > 0 && X1`
                <div class="workspace-folder-starburst-legend">
                    ${legend.slice(0, 8).map((entry) => X1`
                        <div key=${entry.key} class="workspace-folder-starburst-legend-item">
                            <span class="workspace-folder-starburst-swatch" style=${`background:${entry.color}`}></span>
                            <span class="workspace-folder-starburst-name" title=${entry.name}>${entry.name}</span>
                            <span class="workspace-folder-starburst-size">${formatFileSize(entry.size)}</span>
                            <span class="workspace-folder-starburst-pct">${entry.pct.toFixed(1)}%</span>
                        </div>
                    `)}
                </div>
            `}
            ${payload.truncated && X1`
                <div class="workspace-folder-starburst-note">Preview is truncated by tree depth/entry limits.</div>
            `}
        </div>
    `;
}
function triggerWorkspaceDownload(url) {
  if (typeof document === "undefined" || !url)
    return;
  const link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", "");
  link.rel = "noopener";
  link.style.display = "none";
  document.body.appendChild(link);
  link.click();
  link.remove();
}
function describeWorkspaceIndexState(snapshot) {
  switch (snapshot?.state) {
    case "indexing":
      return "Indexing workspace…";
    case "ready":
      return "Workspace index ready";
    case "stale":
      return "Workspace index may be stale";
    case "failed":
      return "Workspace index failed";
    case "never_indexed":
      return "Workspace index not built yet";
    default:
      return "Checking workspace index…";
  }
}
function buildWorkspaceIndexTitle(snapshot) {
  if (!snapshot)
    return "Workspace index status";
  const lines = [describeWorkspaceIndexState(snapshot)];
  if (snapshot.last_indexed_at)
    lines.push(`Last indexed: ${snapshot.last_indexed_at}`);
  if (typeof snapshot.indexed_file_count === "number")
    lines.push(`Indexed files: ${snapshot.indexed_file_count}`);
  if (Array.isArray(snapshot.roots) && snapshot.roots.length)
    lines.push(`Roots: ${snapshot.roots.join(", ")}`);
  if (snapshot.last_error)
    lines.push(`Error: ${snapshot.last_error}`);
  return lines.join(`
`);
}
function getWorkspaceTouchEventTargetElement(event) {
  const target = event?.target;
  if (target && typeof target === "object")
    return target;
  return target?.parentElement || null;
}
function isWorkspaceTouchDragHandleTarget(targetEl) {
  return Boolean(targetEl?.closest?.(".workspace-node-icon, .workspace-label-text"));
}
function getWorkspaceTouchStartIntent(event, renamingPath = null) {
  const targetEl = getWorkspaceTouchEventTargetElement(event);
  const row = targetEl?.closest?.(".workspace-row");
  if (!row)
    return null;
  const type = row.dataset.type;
  const path = row.dataset.path;
  if (!path || path === ".")
    return null;
  if (renamingPath === path)
    return null;
  const touch = event?.touches?.[0];
  if (!touch)
    return null;
  return {
    type,
    path,
    dragPath: isWorkspaceTouchDragHandleTarget(targetEl) ? path : null,
    startX: touch.clientX,
    startY: touch.clientY
  };
}
function WorkspaceExplorer({
  onFileSelect,
  visible = true,
  active = undefined,
  onOpenEditor,
  onOpenTerminalTab,
  onOpenVncTab,
  onToggleTerminal,
  terminalVisible = false
}) {
  const [tree, setTree] = w0(null);
  const [expanded, setExpanded] = w0(new Set(["."]));
  const [selectedPath, setSelectedPath] = w0(null);
  const [renamingPath, setRenamingPath] = w0(null);
  const [renameValue, setRenameValue] = w0("");
  const [preview, setPreview] = w0(null);
  const [, setDownloadId] = w0(null);
  const [initialLoad, setInitialLoad] = w0(true);
  const [loadingPreview, setLoadingPreview] = w0(false);
  const [error, setError] = w0(null);
  const [showHidden, setShowHidden] = w0(() => getLocalStorageBoolean("workspaceShowHidden", false));
  const [dragActive, setDragActive] = w0(false);
  const [dragMode, setDragMode] = w0(null);
  const [dragGhost, setDragGhost] = w0(null);
  const [dropTarget, setDropTarget] = w0(null);
  const [uploading, setUploading] = w0(false);
  const [uploadProgress, setUploadProgress] = w0(null);
  const [folderChart, setFolderChart] = w0(null);
  const [workspaceIndexStatus, setWorkspaceIndexStatus] = w0(null);
  const [workspaceReindexing, setWorkspaceReindexing] = w0(false);
  const [isDarkTheme, setIsDarkTheme] = w0(() => detectDarkTheme2());
  const [explorerScale, setExplorerScale] = w0(() => resolveWorkspaceScale({
    stored: getLocalStorageItem(WORKSPACE_SCALE_STORAGE_KEY),
    ...readWorkspaceScaleEnvironment()
  }));
  const [headerMenuOpen, setHeaderMenuOpen] = w0(false);
  const expandedRef = o0(expanded);
  const lastSigRef = o0("");
  const pendingRootRef = o0(null);
  const rafRef = o0(0);
  const pendingSubtreeRef = o0(new Set);
  const loadTreeFnRef = o0(null);
  const loadWorkspaceIndexStatusRef = o0(null);
  const nodeMapRef = o0(new Map);
  const onFileSelectRef = o0(onFileSelect);
  const onOpenEditorRef = o0(onOpenEditor);
  const loadPreviewRef = o0(null);
  const loadSubtreeRef = o0(null);
  const sidebarRef = o0(null);
  const treeListRef = o0(null);
  const renameInputRef = o0(null);
  const uploadInputRef = o0(null);
  const uploadTargetRef = o0(".");
  const uploadProgressTimerRef = o0(0);
  const touchDragRef = o0({ path: null, dragging: false, startX: 0, startY: 0 });
  const mouseDragRef = o0({ path: null, dragging: false, startX: 0, startY: 0 });
  const dragExpandRef = o0({ path: null, timer: 0 });
  const suppressClickRef = o0(false);
  const previewHeightRef = o0(0);
  const folderChartCacheRef = o0(new Map);
  const folderChartPayloadRef = o0(null);
  const folderChartPathRef = o0(null);
  const previewPaneHostRef = o0(null);
  const previewPaneInstanceRef = o0(null);
  const headerMenuRef = o0(null);
  const headerMenuButtonRef = o0(null);
  const showHiddenRef = o0(showHidden);
  const visibleRef = o0(visible);
  const activeRef = o0(active ?? visible);
  const dragDepthRef = o0(0);
  const dropTargetRef = o0(dropTarget);
  const dragActiveRef = o0(dragActive);
  const dragModeRef = o0(dragMode);
  const dragGhostRef = o0(null);
  const dragGhostPosRef = o0({ x: 0, y: 0 });
  const dragGhostRafRef = o0(0);
  const moveEntryToTargetRef = o0(null);
  const selectedPathRef = o0(selectedPath);
  const renamingPathRef = o0(renamingPath);
  const pendingProgrammaticFileClickRef = o0(null);
  const previewRef = o0(preview);
  onFileSelectRef.current = onFileSelect;
  onOpenEditorRef.current = onOpenEditor;
  r0(() => {
    expandedRef.current = expanded;
  }, [expanded]);
  r0(() => {
    showHiddenRef.current = showHidden;
  }, [showHidden]);
  r0(() => {
    visibleRef.current = visible;
  }, [visible]);
  r0(() => {
    activeRef.current = active ?? visible;
  }, [active, visible]);
  r0(() => {
    dropTargetRef.current = dropTarget;
  }, [dropTarget]);
  const clearUploadProgressTimer = t0(() => {
    if (!uploadProgressTimerRef.current)
      return;
    clearTimeout(uploadProgressTimerRef.current);
    uploadProgressTimerRef.current = 0;
  }, []);
  r0(() => () => clearUploadProgressTimer(), [clearUploadProgressTimer]);
  r0(() => {
    if (typeof window === "undefined")
      return;
    const syncScale = () => {
      setExplorerScale(resolveWorkspaceScale({
        stored: getLocalStorageItem(WORKSPACE_SCALE_STORAGE_KEY),
        ...readWorkspaceScaleEnvironment()
      }));
    };
    syncScale();
    const onResize = () => syncScale();
    const onFocus = () => syncScale();
    const onStorage = (event) => {
      if (!event || event.key === null || event.key === WORKSPACE_SCALE_STORAGE_KEY)
        syncScale();
    };
    window.addEventListener("resize", onResize);
    window.addEventListener("focus", onFocus);
    window.addEventListener("storage", onStorage);
    const pointerMedia = window.matchMedia?.("(pointer: coarse)");
    const hoverMedia = window.matchMedia?.("(hover: none)");
    const addMediaListener = (media, handler) => {
      if (!media)
        return;
      if (media.addEventListener)
        media.addEventListener("change", handler);
      else if (media.addListener)
        media.addListener(handler);
    };
    const removeMediaListener = (media, handler) => {
      if (!media)
        return;
      if (media.removeEventListener)
        media.removeEventListener("change", handler);
      else if (media.removeListener)
        media.removeListener(handler);
    };
    addMediaListener(pointerMedia, onResize);
    addMediaListener(hoverMedia, onResize);
    return () => {
      window.removeEventListener("resize", onResize);
      window.removeEventListener("focus", onFocus);
      window.removeEventListener("storage", onStorage);
      removeMediaListener(pointerMedia, onResize);
      removeMediaListener(hoverMedia, onResize);
    };
  }, []);
  r0(() => {
    const handleReveal = (e2) => {
      const path = e2?.detail?.path;
      if (!path)
        return;
      const parts = path.split("/");
      const parents = [];
      for (let i3 = 1;i3 < parts.length; i3++) {
        parents.push(parts.slice(0, i3).join("/"));
      }
      if (parents.length) {
        setExpanded((prev) => {
          const next = new Set(prev);
          next.add(".");
          for (const p2 of parents)
            next.add(p2);
          return next;
        });
      }
      setSelectedPath(path);
      requestAnimationFrame(() => {
        const row = document.querySelector(`[data-path="${CSS.escape(path)}"]`);
        if (row)
          row.scrollIntoView({ block: "nearest", behavior: "smooth" });
      });
    };
    window.addEventListener("workspace-reveal-path", handleReveal);
    return () => window.removeEventListener("workspace-reveal-path", handleReveal);
  }, []);
  r0(() => {
    dragActiveRef.current = dragActive;
  }, [dragActive]);
  r0(() => {
    dragModeRef.current = dragMode;
  }, [dragMode]);
  r0(() => {
    selectedPathRef.current = selectedPath;
  }, [selectedPath]);
  r0(() => {
    renamingPathRef.current = renamingPath;
  }, [renamingPath]);
  r0(() => {
    previewRef.current = preview;
  }, [preview]);
  r0(() => {
    if (typeof window === "undefined" || typeof document === "undefined")
      return;
    const syncTheme = () => setIsDarkTheme(detectDarkTheme2());
    syncTheme();
    const media = window.matchMedia?.("(prefers-color-scheme: dark)");
    const onMediaChange = () => syncTheme();
    if (media?.addEventListener)
      media.addEventListener("change", onMediaChange);
    else if (media?.addListener)
      media.addListener(onMediaChange);
    const observer = typeof MutationObserver !== "undefined" ? new MutationObserver(() => syncTheme()) : null;
    observer?.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class", "data-theme"]
    });
    if (document.body) {
      observer?.observe(document.body, {
        attributes: true,
        attributeFilter: ["class", "data-theme"]
      });
    }
    return () => {
      if (media?.removeEventListener)
        media.removeEventListener("change", onMediaChange);
      else if (media?.removeListener)
        media.removeListener(onMediaChange);
      observer?.disconnect();
    };
  }, []);
  r0(() => {
    if (!renamingPath)
      return;
    const input = renameInputRef.current;
    if (!input)
      return;
    const timer = requestAnimationFrame(() => {
      focusAndSelectBestEffort(input);
    });
    return () => cancelAnimationFrame(timer);
  }, [renamingPath]);
  r0(() => {
    if (!headerMenuOpen)
      return;
    const handleDocPointer = (event) => {
      const target = event?.target;
      if (!(target instanceof Element))
        return;
      if (headerMenuRef.current?.contains(target))
        return;
      if (headerMenuButtonRef.current?.contains(target))
        return;
      setHeaderMenuOpen(false);
    };
    const handleEscape = (event) => {
      if (event?.key === "Escape") {
        setHeaderMenuOpen(false);
        headerMenuButtonRef.current?.focus?.();
      }
    };
    document.addEventListener("mousedown", handleDocPointer);
    document.addEventListener("touchstart", handleDocPointer, { passive: true });
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleDocPointer);
      document.removeEventListener("touchstart", handleDocPointer);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [headerMenuOpen]);
  const loadPreview = async (path, options = {}) => {
    const autoOpen = Boolean(options?.autoOpen);
    const normalized = String(path || "").trim();
    setLoadingPreview(true);
    setPreview(null);
    setDownloadId(null);
    try {
      const data = await getWorkspaceFile(normalized, 20000);
      if (autoOpen && normalized && shouldAutoOpenWorkspaceFile(normalized, data, {
        resolvePane: (context) => paneRegistry.resolve(context)
      })) {
        onOpenEditorRef.current?.(normalized, data);
        return data;
      }
      setPreview(data);
      return data;
    } catch (err2) {
      const failure = { error: err2.message || "Failed to load preview" };
      setPreview(failure);
      return failure;
    } finally {
      setLoadingPreview(false);
    }
  };
  loadPreviewRef.current = loadPreview;
  const loadWorkspaceIndexStatus = t0(async () => {
    try {
      const status = await getWorkspaceIndexStatus("all");
      setWorkspaceIndexStatus(status);
      return status;
    } catch (err2) {
      console.warn("[workspace-explorer] Failed to load workspace index status:", err2);
      return null;
    }
  }, []);
  loadWorkspaceIndexStatusRef.current = loadWorkspaceIndexStatus;
  const refreshWorkspaceIndexStatus = t0(() => {
    loadWorkspaceIndexStatusRef.current?.();
  }, []);
  const loadTree = async () => {
    if (!visibleRef.current)
      return;
    try {
      const data = await getWorkspaceTree("", 1, showHiddenRef.current);
      const sig = treeSignature(data.root, expandedRef.current, showHiddenRef.current);
      if (sig === lastSigRef.current) {
        setInitialLoad(false);
        return;
      }
      lastSigRef.current = sig;
      pendingRootRef.current = data.root;
      if (!rafRef.current) {
        rafRef.current = requestAnimationFrame(() => {
          rafRef.current = 0;
          setTree((prev) => mergeTree(prev, pendingRootRef.current));
          setInitialLoad(false);
        });
      }
    } catch (err2) {
      setError(err2.message || "Failed to load workspace");
      setInitialLoad(false);
    }
  };
  const loadSubtree = async (path) => {
    if (!path)
      return;
    if (pendingSubtreeRef.current.has(path))
      return;
    pendingSubtreeRef.current.add(path);
    try {
      const data = await getWorkspaceTree(path, 1, showHiddenRef.current);
      setTree((prev) => replaceNodeAtPath(prev, path, data.root));
    } catch (err2) {
      setError(err2.message || "Failed to load workspace");
    } finally {
      pendingSubtreeRef.current.delete(path);
    }
  };
  loadSubtreeRef.current = loadSubtree;
  const resolveDropTargetPath = t0(() => {
    const selected = selectedPath;
    if (!selected)
      return ".";
    const node = nodeMapRef.current?.get(selected);
    if (node && node.type === "dir")
      return node.path;
    if (selected === "." || !selected.includes("/"))
      return ".";
    const parts = selected.split("/");
    parts.pop();
    const parent = parts.join("/");
    return parent || ".";
  }, [selectedPath]);
  const resolveDropTargetFromElement = t0((element) => {
    const row = element?.closest?.(".workspace-row");
    if (!row)
      return null;
    const path = row.dataset.path;
    const type = row.dataset.type;
    if (!path)
      return null;
    if (type === "dir")
      return path;
    if (path.includes("/")) {
      const parts = path.split("/");
      parts.pop();
      return parts.join("/") || ".";
    }
    return ".";
  }, []);
  const resolveDropTargetFromEvent = t0((event) => {
    return resolveDropTargetFromElement(event?.target || null);
  }, [resolveDropTargetFromElement]);
  const updateDropTarget = t0((value) => {
    dropTargetRef.current = value;
    setDropTarget(value);
  }, []);
  const clearDragExpandTimer = t0(() => {
    const current = dragExpandRef.current;
    if (current?.timer)
      clearTimeout(current.timer);
    dragExpandRef.current = { path: null, timer: 0 };
  }, []);
  const scheduleDragExpand = t0((targetPath) => {
    if (!targetPath || targetPath === ".") {
      clearDragExpandTimer();
      return;
    }
    const node = nodeMapRef.current?.get(targetPath);
    if (!node || node.type !== "dir") {
      clearDragExpandTimer();
      return;
    }
    if (expandedRef.current?.has(targetPath)) {
      clearDragExpandTimer();
      return;
    }
    if (dragExpandRef.current?.path === targetPath)
      return;
    clearDragExpandTimer();
    const timer = setTimeout(() => {
      dragExpandRef.current = { path: null, timer: 0 };
      loadSubtreeRef.current?.(targetPath);
      setExpanded((prev) => {
        const next = new Set(prev);
        next.add(targetPath);
        return next;
      });
    }, 600);
    dragExpandRef.current = { path: targetPath, timer };
  }, [clearDragExpandTimer]);
  const updateDragGhostPosition = t0((x3, y2) => {
    dragGhostPosRef.current = { x: x3, y: y2 };
    if (dragGhostRafRef.current)
      return;
    dragGhostRafRef.current = requestAnimationFrame(() => {
      dragGhostRafRef.current = 0;
      const el = dragGhostRef.current;
      if (!el)
        return;
      const pos = dragGhostPosRef.current;
      el.style.transform = `translate(${pos.x + 12}px, ${pos.y + 12}px)`;
    });
  }, []);
  const startDragGhost = t0((path) => {
    if (!path)
      return;
    const node = nodeMapRef.current?.get(path);
    const label = (node?.name || path.split("/").pop() || path).trim();
    if (!label)
      return;
    setDragGhost({ path, label });
  }, []);
  const clearDragGhost = t0(() => {
    setDragGhost(null);
    if (dragGhostRafRef.current) {
      cancelAnimationFrame(dragGhostRafRef.current);
      dragGhostRafRef.current = 0;
    }
    if (dragGhostRef.current) {
      dragGhostRef.current.style.transform = "translate(-9999px, -9999px)";
    }
  }, []);
  const resolveCreateTargetPath = t0((path) => {
    if (!path)
      return ".";
    const node = nodeMapRef.current?.get(path);
    if (node && node.type === "dir")
      return node.path;
    if (path === "." || !path.includes("/"))
      return ".";
    const parts = path.split("/");
    parts.pop();
    const parent = parts.join("/");
    return parent || ".";
  }, []);
  const cancelRename = t0(() => {
    setRenamingPath(null);
    setRenameValue("");
  }, []);
  const beginRename = t0((path) => {
    if (!path)
      return;
    const node = nodeMapRef.current?.get(path);
    const base = (node?.name || path.split("/").pop() || path).trim();
    if (!base || path === ".")
      return;
    setRenamingPath(path);
    setRenameValue(base);
  }, []);
  const commitRename = t0(async () => {
    const targetPath = renamingPathRef.current;
    if (!targetPath)
      return;
    const nextName = (renameValue || "").trim();
    if (!nextName) {
      cancelRename();
      return;
    }
    const node = nodeMapRef.current?.get(targetPath);
    const currentName = (node?.name || targetPath.split("/").pop() || targetPath).trim();
    if (nextName === currentName) {
      cancelRename();
      return;
    }
    try {
      const result = await renameWorkspaceFile(targetPath, nextName);
      const nextPath = result?.path || targetPath;
      const parent = targetPath.includes("/") ? targetPath.split("/").slice(0, -1).join("/") || "." : ".";
      cancelRename();
      setError(null);
      window.dispatchEvent(new CustomEvent("workspace-file-renamed", {
        detail: { oldPath: targetPath, newPath: nextPath, type: node?.type || "file" }
      }));
      if (node?.type === "dir") {
        setExpanded((prev) => {
          const next = new Set;
          for (const entry of prev) {
            if (entry === targetPath) {
              next.add(nextPath);
            } else if (entry.startsWith(`${targetPath}/`)) {
              next.add(`${nextPath}${entry.slice(targetPath.length)}`);
            } else {
              next.add(entry);
            }
          }
          return next;
        });
      }
      setSelectedPath(nextPath);
      if (node?.type === "dir") {
        setPreview(null);
        setLoadingPreview(false);
        setDownloadId(null);
      } else {
        loadPreviewRef.current?.(nextPath);
      }
      loadSubtreeRef.current?.(parent);
      refreshWorkspaceIndexStatus();
    } catch (err2) {
      setError(err2?.message || "Failed to rename file");
    }
  }, [cancelRename, renameValue, refreshWorkspaceIndexStatus]);
  const createUntitledFile = t0(async (targetPath) => {
    const base = "untitled";
    const ext = ".md";
    const folder = targetPath || ".";
    for (let i3 = 0;i3 < 50; i3 += 1) {
      const suffix = i3 === 0 ? "" : `-${i3}`;
      const name = `${base}${suffix}${ext}`;
      try {
        const result = await createWorkspaceFile(folder, name, "");
        const nextPath = result?.path || (folder === "." ? name : `${folder}/${name}`);
        if (folder && folder !== ".") {
          setExpanded((prev) => new Set([...prev, folder]));
        }
        setSelectedPath(nextPath);
        setError(null);
        loadSubtreeRef.current?.(folder);
        loadPreviewRef.current?.(nextPath);
        refreshWorkspaceIndexStatus();
        return;
      } catch (err2) {
        if (err2?.status === 409 || err2?.code === "file_exists") {
          continue;
        }
        setError(err2?.message || "Failed to create file");
        return;
      }
    }
    setError("Failed to create file (untitled name already in use).");
  }, []);
  const handleCreateFileClick = t0((event) => {
    event?.stopPropagation?.();
    if (uploading)
      return;
    const target = resolveCreateTargetPath(selectedPathRef.current);
    createUntitledFile(target);
  }, [uploading, resolveCreateTargetPath, createUntitledFile]);
  r0(() => {
    if (typeof window === "undefined")
      return;
    const handler = (event) => {
      const updates = event?.detail?.updates || [];
      if (!Array.isArray(updates) || updates.length === 0)
        return;
      setTree((prev) => {
        let next = prev;
        for (const update of updates) {
          if (!update?.root)
            continue;
          if (!next || update.path === "." || !update.path) {
            next = update.root;
          } else {
            next = replaceNodeAtPath(next, update.path, update.root);
          }
        }
        if (next) {
          lastSigRef.current = treeSignature(next, expandedRef.current, showHiddenRef.current);
        }
        setInitialLoad(false);
        return next;
      });
      const selected = selectedPathRef.current;
      const shouldRefreshStarburst = Boolean(selected) && updates.some((update) => {
        const path = update?.path || "";
        if (!path || path === ".")
          return true;
        return selected === path || selected.startsWith(`${path}/`) || path.startsWith(`${selected}/`);
      });
      if (shouldRefreshStarburst) {
        folderChartCacheRef.current.clear();
      }
      refreshWorkspaceIndexStatus();
      if (!selected || !previewRef.current)
        return;
      const node = nodeMapRef.current?.get(selected);
      if (node && node.type === "dir")
        return;
      const shouldRefresh = updates.some((update) => {
        const path = update?.path || "";
        if (!path || path === ".")
          return true;
        return selected === path || selected.startsWith(`${path}/`);
      });
      if (shouldRefresh) {
        loadPreviewRef.current?.(selected);
      }
    };
    window.addEventListener("workspace-update", handler);
    return () => window.removeEventListener("workspace-update", handler);
  }, []);
  loadTreeFnRef.current = loadTree;
  const updateVisibility = o0(() => {
    if (typeof window === "undefined")
      return;
    const media = window.matchMedia("(min-width: 1024px) and (orientation: landscape)");
    const active2 = activeRef.current ?? visibleRef.current;
    const visible2 = document.visibilityState !== "hidden" && (active2 || media.matches && visibleRef.current);
    setWorkspaceVisibility(visible2, showHiddenRef.current).catch((error2) => {
      console.debug("[workspace-explorer] Workspace visibility ping failed.", error2, {
        visible: visible2,
        showHidden: showHiddenRef.current
      });
    });
  }).current;
  const debouncedVisibilityRef = o0(0);
  const scheduleVisibilityUpdate = o0(() => {
    if (debouncedVisibilityRef.current) {
      clearTimeout(debouncedVisibilityRef.current);
    }
    debouncedVisibilityRef.current = setTimeout(() => {
      debouncedVisibilityRef.current = 0;
      updateVisibility();
    }, 250);
  }).current;
  r0(() => {
    if (visibleRef.current) {
      loadTreeFnRef.current?.();
      loadWorkspaceIndexStatusRef.current?.();
    }
    scheduleVisibilityUpdate();
  }, [visible, active]);
  r0(() => {
    loadTreeFnRef.current();
    loadWorkspaceIndexStatusRef.current?.();
    updateVisibility();
    const timer = setInterval(() => {
      loadTreeFnRef.current();
      loadWorkspaceIndexStatusRef.current?.();
    }, REFRESH_INTERVAL_MS);
    const saved = getLocalStorageNumber("previewHeight", null);
    const h = Number.isFinite(saved) ? Math.min(Math.max(saved, 80), 600) : 280;
    previewHeightRef.current = h;
    if (sidebarRef.current) {
      sidebarRef.current.style.setProperty("--preview-height", `${h}px`);
    }
    const media = window.matchMedia("(min-width: 1024px) and (orientation: landscape)");
    const onVisibilityChange = () => scheduleVisibilityUpdate();
    if (media.addEventListener) {
      media.addEventListener("change", onVisibilityChange);
    } else if (media.addListener) {
      media.addListener(onVisibilityChange);
    }
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () => {
      clearInterval(timer);
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = 0;
      }
      if (media.removeEventListener) {
        media.removeEventListener("change", onVisibilityChange);
      } else if (media.removeListener) {
        media.removeListener(onVisibilityChange);
      }
      document.removeEventListener("visibilitychange", onVisibilityChange);
      if (debouncedVisibilityRef.current) {
        clearTimeout(debouncedVisibilityRef.current);
        debouncedVisibilityRef.current = 0;
      }
      setWorkspaceVisibility(false, showHiddenRef.current).catch((error2) => {
        console.debug("[workspace-explorer] Workspace visibility teardown ping failed.", error2, {
          showHidden: showHiddenRef.current
        });
      });
    };
  }, []);
  const rows = G0(() => flattenTree(tree, expanded, showHidden), [tree, expanded, showHidden]);
  const nodeMap = G0(() => new Map(rows.map((r2) => [r2.node.path, r2.node])), [rows]);
  const workspaceScaleMetrics = G0(() => getWorkspaceScaleMetrics(explorerScale), [explorerScale]);
  nodeMapRef.current = nodeMap;
  const selectedNode = selectedPath ? nodeMapRef.current.get(selectedPath) : null;
  const selectedIsDir = selectedNode?.type === "dir";
  r0(() => {
    if (!selectedPath || !selectedIsDir) {
      setFolderChart(null);
      folderChartPayloadRef.current = null;
      folderChartPathRef.current = null;
      return;
    }
    const fetchPath = selectedPath;
    const cacheKey = `${showHidden ? "hidden" : "visible"}:${selectedPath}`;
    const cache = folderChartCacheRef.current;
    const cached = cache.get(cacheKey);
    if (cached?.root) {
      cache.delete(cacheKey);
      cache.set(cacheKey, cached);
      const payload = createFolderStarburstPayload(cached.root, Boolean(cached.truncated), isDarkTheme);
      if (payload) {
        folderChartPayloadRef.current = payload;
        folderChartPathRef.current = selectedPath;
        setFolderChart({ loading: false, error: null, payload });
      }
      return;
    }
    const lastPayload = folderChartPayloadRef.current;
    const lastPath = folderChartPathRef.current;
    setFolderChart({ loading: true, error: null, payload: lastPath === selectedPath ? lastPayload : null });
    getWorkspaceTree(selectedPath, STARBURST_FETCH_DEPTH, showHidden).then((data) => {
      if (selectedPathRef.current !== fetchPath)
        return;
      const entry = { root: data?.root, truncated: Boolean(data?.truncated) };
      cache.delete(cacheKey);
      cache.set(cacheKey, entry);
      while (cache.size > STARBURST_CACHE_LIMIT) {
        const oldest = cache.keys().next().value;
        if (!oldest)
          break;
        cache.delete(oldest);
      }
      const payload = createFolderStarburstPayload(entry.root, entry.truncated, isDarkTheme);
      folderChartPayloadRef.current = payload;
      folderChartPathRef.current = selectedPath;
      setFolderChart({ loading: false, error: null, payload });
    }).catch((err2) => {
      if (selectedPathRef.current !== fetchPath)
        return;
      setFolderChart({ loading: false, error: err2?.message || "Failed to load folder size chart", payload: lastPath === selectedPath ? lastPayload : null });
    });
  }, [selectedPath, selectedIsDir, showHidden, isDarkTheme]);
  const canEdit = Boolean(preview && preview.kind === "text" && !selectedIsDir && (!preview.size || preview.size <= 256 * 1024));
  const editTitle = canEdit ? "Open in editor" : preview?.size > 256 * 1024 ? "File too large to edit" : "File is not editable";
  const selectedHasOpenableTab = Boolean(selectedPath && !selectedIsDir && hasOpenableWorkspaceTab(selectedPath));
  const selectedCanRename = Boolean(selectedPath && selectedPath !== ".");
  const selectedCanDelete = Boolean(selectedPath && !selectedIsDir);
  const selectedCanDownload = Boolean(selectedPath && !selectedIsDir);
  const selectedFolderDownloadUrl = selectedPath && selectedIsDir ? getWorkspaceDownloadUrl(selectedPath, showHidden) : null;
  const workspaceIndexLabel = describeWorkspaceIndexState(workspaceIndexStatus);
  const workspaceIndexTitle = buildWorkspaceIndexTitle(workspaceIndexStatus);
  const workspaceIndexState = workspaceIndexStatus?.state || "never_indexed";
  const showWorkspaceIndexIndicator = workspaceIndexState !== "ready";
  const closeHeaderMenu = t0(() => setHeaderMenuOpen(false), []);
  const runMenuAction = t0(async (fn) => {
    closeHeaderMenu();
    try {
      await fn?.();
    } catch (err2) {
      console.warn("[workspace-explorer] Header menu action failed:", err2);
    }
  }, [closeHeaderMenu]);
  const handleWorkspaceReindex = t0(async (event) => {
    event?.stopPropagation?.();
    setWorkspaceReindexing(true);
    setWorkspaceIndexStatus((prev) => ({
      scope: "all",
      last_indexed_at: prev?.last_indexed_at || null,
      last_error: null,
      indexed_file_count: prev?.indexed_file_count || 0,
      roots: prev?.roots || [],
      updated_at: prev?.updated_at || null,
      state: "indexing"
    }));
    try {
      const status = await reindexWorkspace("all");
      setWorkspaceIndexStatus(status);
      setError(null);
      lastSigRef.current = "";
      loadTreeFnRef.current?.();
    } catch (err2) {
      const message = err2?.message || "Failed to reindex workspace";
      setWorkspaceIndexStatus((prev) => ({
        scope: "all",
        last_indexed_at: prev?.last_indexed_at || null,
        last_error: message,
        indexed_file_count: prev?.indexed_file_count || 0,
        roots: prev?.roots || [],
        updated_at: prev?.updated_at || null,
        state: "failed"
      }));
      setError(message);
    } finally {
      setWorkspaceReindexing(false);
    }
  }, []);
  r0(() => {
    const container = previewPaneHostRef.current;
    if (previewPaneInstanceRef.current) {
      previewPaneInstanceRef.current.dispose();
      previewPaneInstanceRef.current = null;
    }
    if (!container)
      return;
    container.innerHTML = "";
    if (!selectedPath || selectedIsDir || !preview || preview.error)
      return;
    const context = {
      path: selectedPath,
      content: typeof preview.text === "string" ? preview.text : undefined,
      mtime: preview.mtime,
      size: preview.size,
      preview,
      mode: "view"
    };
    const extension = paneRegistry.resolve(context) || paneRegistry.get("workspace-preview-default");
    if (!extension)
      return;
    const instance = extension.mount(container, context);
    previewPaneInstanceRef.current = instance;
    return () => {
      if (previewPaneInstanceRef.current === instance) {
        instance.dispose();
        previewPaneInstanceRef.current = null;
      }
      container.innerHTML = "";
    };
  }, [selectedPath, selectedIsDir, preview]);
  const getEventTargetElement = (event) => {
    const target = event?.target;
    if (target instanceof Element)
      return target;
    return target?.parentElement || null;
  };
  const isRowDragHandleTarget = (targetEl) => {
    return Boolean(targetEl?.closest?.(".workspace-node-icon, .workspace-label-text"));
  };
  const isEditableKeyboardTarget = (targetEl) => {
    if (!targetEl)
      return false;
    if (targetEl.closest?.('input, textarea, [contenteditable="true"]'))
      return true;
    return Boolean(targetEl.isContentEditable);
  };
  const handleTreeDblClick = o0((e2) => {
    const targetEl = getEventTargetElement(e2);
    const rowEl = targetEl?.closest?.("[data-path]");
    if (!rowEl)
      return;
    const clickedPath = rowEl.dataset.path;
    if (!clickedPath || clickedPath === ".")
      return;
    const isActionClick = Boolean(targetEl?.closest?.("button")) || Boolean(targetEl?.closest?.("a")) || Boolean(targetEl?.closest?.("input"));
    const isCaretClick = Boolean(targetEl?.closest?.(".workspace-caret"));
    if (isActionClick || isCaretClick)
      return;
    if (renamingPathRef.current === clickedPath)
      return;
    beginRename(clickedPath);
  }).current;
  const handleTreeClick = o0((e2) => {
    if (suppressClickRef.current) {
      suppressClickRef.current = false;
      return;
    }
    const targetEl = getEventTargetElement(e2);
    const rowEl = targetEl?.closest?.("[data-path]");
    treeListRef.current?.focus?.();
    if (!rowEl)
      return;
    const clickedPath = rowEl.dataset.path;
    const clickedType = rowEl.dataset.type;
    const isCaretClick = Boolean(targetEl?.closest?.(".workspace-caret"));
    const isActionClick = Boolean(targetEl?.closest?.("button")) || Boolean(targetEl?.closest?.("a")) || Boolean(targetEl?.closest?.("input"));
    const isSelected = selectedPathRef.current === clickedPath;
    const renaming = renamingPathRef.current;
    if (renaming) {
      if (renaming === clickedPath)
        return;
      cancelRename();
    }
    if (clickedType === "dir") {
      pendingProgrammaticFileClickRef.current = null;
      setSelectedPath(clickedPath);
      setPreview(null);
      setDownloadId(null);
      setLoadingPreview(false);
      const wasExpanded = expandedRef.current.has(clickedPath);
      if (!wasExpanded)
        loadSubtreeRef.current?.(clickedPath);
      if (isSelected && !isCaretClick)
        return;
      setExpanded((prev) => {
        const next = new Set(prev);
        if (next.has(clickedPath))
          next.delete(clickedPath);
        else
          next.add(clickedPath);
        return next;
      });
    } else {
      pendingProgrammaticFileClickRef.current = null;
      setSelectedPath(clickedPath);
      const node = nodeMapRef.current.get(clickedPath);
      if (node)
        onFileSelectRef.current?.(node.path, node);
      if (!isActionClick && !isCaretClick) {
        loadPreviewRef.current?.(clickedPath);
      }
    }
  }).current;
  const handleRefreshClick = o0(() => {
    lastSigRef.current = "";
    loadTreeFnRef.current();
    loadWorkspaceIndexStatusRef.current?.();
    const openPaths = Array.from(expandedRef.current || []).filter((p2) => p2 && p2 !== ".");
    openPaths.forEach((p2) => loadSubtreeRef.current?.(p2));
  }).current;
  const clearSelection = o0(() => {
    pendingProgrammaticFileClickRef.current = null;
    setSelectedPath(null);
    setPreview(null);
    setDownloadId(null);
    setLoadingPreview(false);
  }).current;
  const handleToggleHidden = o0(() => {
    setShowHidden((prev) => {
      const next = !prev;
      if (typeof window !== "undefined") {
        setLocalStorageItem("workspaceShowHidden", String(next));
      }
      showHiddenRef.current = next;
      setWorkspaceVisibility(true, next).catch((error2) => {
        console.debug("[workspace-explorer] Workspace visibility refresh after toggling hidden files failed.", error2, {
          showHidden: next
        });
      });
      lastSigRef.current = "";
      loadTreeFnRef.current?.();
      const openPaths = Array.from(expandedRef.current || []).filter((p2) => p2 && p2 !== ".");
      openPaths.forEach((p2) => loadSubtreeRef.current?.(p2));
      return next;
    });
  }).current;
  const handleBackgroundClick = o0((e2) => {
    const targetEl = getEventTargetElement(e2);
    if (targetEl?.closest?.("[data-path]"))
      return;
    clearSelection();
  }).current;
  const deleteFileAtPath = t0(async (path) => {
    if (!path)
      return;
    const filename = path.split("/").pop() || path;
    const confirmed = window.confirm(`Delete "${filename}"? This cannot be undone.`);
    if (!confirmed)
      return;
    try {
      await deleteWorkspaceFile(path);
      const parent = path.includes("/") ? path.split("/").slice(0, -1).join("/") || "." : ".";
      if (selectedPathRef.current === path) {
        clearSelection();
      }
      loadSubtreeRef.current?.(parent);
      setError(null);
      refreshWorkspaceIndexStatus();
    } catch (err2) {
      setPreview((prev) => ({ ...prev || {}, error: err2.message || "Failed to delete file" }));
    }
  }, [clearSelection]);
  const scrollRowIntoView = t0((path) => {
    const container = treeListRef.current;
    if (!container || !path || typeof CSS === "undefined" || typeof CSS.escape !== "function")
      return;
    const el = container.querySelector(`[data-path="${CSS.escape(path)}"]`);
    el?.scrollIntoView?.({ block: "nearest" });
  }, []);
  const handleTreeKeyDown = t0((e2) => {
    const targetEl = getEventTargetElement(e2);
    if (renamingPathRef.current || isEditableKeyboardTarget(targetEl))
      return;
    const currentRows = rows;
    if (!currentRows || currentRows.length === 0)
      return;
    const currentIndex = selectedPath ? currentRows.findIndex((r2) => r2.node.path === selectedPath) : -1;
    if (e2.key === "ArrowDown") {
      e2.preventDefault();
      const next = Math.min(currentIndex + 1, currentRows.length - 1);
      const row = currentRows[next];
      if (!row)
        return;
      setSelectedPath(row.node.path);
      if (row.node.type !== "dir") {
        onFileSelectRef.current?.(row.node.path, row.node);
        loadPreviewRef.current?.(row.node.path);
      } else {
        setPreview(null);
        setLoadingPreview(false);
        setDownloadId(null);
      }
      scrollRowIntoView(row.node.path);
      return;
    }
    if (e2.key === "ArrowUp") {
      e2.preventDefault();
      const next = currentIndex <= 0 ? 0 : currentIndex - 1;
      const row = currentRows[next];
      if (!row)
        return;
      setSelectedPath(row.node.path);
      if (row.node.type !== "dir") {
        onFileSelectRef.current?.(row.node.path, row.node);
        loadPreviewRef.current?.(row.node.path);
      } else {
        setPreview(null);
        setLoadingPreview(false);
        setDownloadId(null);
      }
      scrollRowIntoView(row.node.path);
      return;
    }
    if (e2.key === "ArrowRight" && currentIndex >= 0) {
      const row = currentRows[currentIndex];
      if (row?.node?.type === "dir" && !expanded.has(row.node.path)) {
        e2.preventDefault();
        loadSubtreeRef.current?.(row.node.path);
        setExpanded((prev) => new Set([...prev, row.node.path]));
      }
      return;
    }
    if (e2.key === "ArrowLeft" && currentIndex >= 0) {
      const row = currentRows[currentIndex];
      if (row?.node?.type === "dir" && expanded.has(row.node.path)) {
        e2.preventDefault();
        setExpanded((prev) => {
          const next = new Set(prev);
          next.delete(row.node.path);
          return next;
        });
      }
      return;
    }
    if (e2.key === "Enter" && currentIndex >= 0) {
      e2.preventDefault();
      const row = currentRows[currentIndex];
      if (!row)
        return;
      const path = row.node.path;
      if (row.node.type === "dir") {
        const wasExpanded = expandedRef.current.has(path);
        if (!wasExpanded)
          loadSubtreeRef.current?.(path);
        setExpanded((prev) => {
          const next = new Set(prev);
          if (next.has(path))
            next.delete(path);
          else
            next.add(path);
          return next;
        });
        setPreview(null);
        setDownloadId(null);
        setLoadingPreview(false);
      } else {
        onFileSelectRef.current?.(path, row.node);
        loadPreviewRef.current?.(path);
      }
      return;
    }
    if ((e2.key === "Delete" || e2.key === "Backspace") && currentIndex >= 0) {
      const row = currentRows[currentIndex];
      if (!row || row.node.type === "dir")
        return;
      e2.preventDefault();
      deleteFileAtPath(row.node.path);
      return;
    }
    if (e2.key === "Escape") {
      e2.preventDefault();
      clearSelection();
    }
  }, [clearSelection, deleteFileAtPath, expanded, rows, scrollRowIntoView, selectedPath]);
  const handleRowTouchStart = t0((event) => {
    const intent = getWorkspaceTouchStartIntent(event, renamingPathRef.current);
    if (!intent)
      return;
    touchDragRef.current = {
      path: intent.dragPath,
      dragging: false,
      startX: intent.startX,
      startY: intent.startY
    };
  }, []);
  const handleRowTouchEnd = t0(() => {
    const dragState = touchDragRef.current;
    if (dragState?.dragging && dragState.path) {
      const target = dropTargetRef.current || resolveDropTargetPath();
      const mover = moveEntryToTargetRef.current;
      if (typeof mover === "function")
        mover(dragState.path, target);
    }
    touchDragRef.current = { path: null, dragging: false, startX: 0, startY: 0 };
    dragDepthRef.current = 0;
    setDragActive(false);
    setDragMode(null);
    updateDropTarget(null);
    clearDragExpandTimer();
    clearDragGhost();
  }, [resolveDropTargetPath, clearDragGhost, updateDropTarget, clearDragExpandTimer]);
  const handleRowTouchMove = t0((event) => {
    const dragState = touchDragRef.current;
    const touch = event?.touches?.[0];
    if (!touch || !dragState?.path)
      return;
    const dx = Math.abs(touch.clientX - dragState.startX);
    const dy = Math.abs(touch.clientY - dragState.startY);
    const moved = dx > 8 || dy > 8;
    if (!dragState.dragging && moved) {
      dragState.dragging = true;
      setDragActive(true);
      setDragMode("move");
      startDragGhost(dragState.path);
    }
    if (dragState.dragging) {
      event.preventDefault();
      updateDragGhostPosition(touch.clientX, touch.clientY);
      const el = document.elementFromPoint(touch.clientX, touch.clientY);
      const target = resolveDropTargetFromElement(el) || resolveDropTargetPath();
      if (dropTargetRef.current !== target)
        updateDropTarget(target);
      scheduleDragExpand(target);
    }
  }, [resolveDropTargetFromElement, resolveDropTargetPath, startDragGhost, updateDragGhostPosition, updateDropTarget, scheduleDragExpand]);
  const handlePreviewSplitterMouseDown = o0((e2) => {
    e2.preventDefault();
    const sidebar = sidebarRef.current;
    if (!sidebar)
      return;
    const startY = e2.clientY;
    const startH = previewHeightRef.current || 280;
    const splitter = e2.currentTarget;
    splitter.classList.add("dragging");
    document.body.style.cursor = "row-resize";
    document.body.style.userSelect = "none";
    let lastY = startY;
    const onMove = (me) => {
      lastY = me.clientY;
      const maxH = sidebar.clientHeight - 80;
      const h = Math.min(Math.max(startH - (me.clientY - startY), 80), maxH);
      sidebar.style.setProperty("--preview-height", `${h}px`);
      previewHeightRef.current = h;
    };
    const onUp = () => {
      const maxH = sidebar.clientHeight - 80;
      const h = Math.min(Math.max(startH - (lastY - startY), 80), maxH);
      previewHeightRef.current = h;
      splitter.classList.remove("dragging");
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      setLocalStorageItem("previewHeight", String(Math.round(h)));
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
    };
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  }).current;
  const handlePreviewSplitterTouchStart = o0((e2) => {
    e2.preventDefault();
    const sidebar = sidebarRef.current;
    if (!sidebar)
      return;
    const touch = e2.touches[0];
    if (!touch)
      return;
    const startY = touch.clientY;
    const startH = previewHeightRef.current || 280;
    const splitter = e2.currentTarget;
    splitter.classList.add("dragging");
    document.body.style.userSelect = "none";
    const onMove = (te) => {
      const t2 = te.touches[0];
      if (!t2)
        return;
      te.preventDefault();
      const maxH = sidebar.clientHeight - 80;
      const h = Math.min(Math.max(startH - (t2.clientY - startY), 80), maxH);
      sidebar.style.setProperty("--preview-height", `${h}px`);
      previewHeightRef.current = h;
    };
    const onUp = () => {
      splitter.classList.remove("dragging");
      document.body.style.userSelect = "";
      setLocalStorageItem("previewHeight", String(Math.round(previewHeightRef.current || startH)));
      document.removeEventListener("touchmove", onMove);
      document.removeEventListener("touchend", onUp);
      document.removeEventListener("touchcancel", onUp);
    };
    document.addEventListener("touchmove", onMove, { passive: false });
    document.addEventListener("touchend", onUp);
    document.addEventListener("touchcancel", onUp);
  }).current;
  const handleDownload = t0((path = selectedPath) => {
    if (!path)
      return;
    triggerWorkspaceDownload(getWorkspaceFileDownloadUrl(path));
  }, [selectedPath]);
  const handleDeleteFile = async () => {
    if (!selectedPath || selectedIsDir)
      return;
    await deleteFileAtPath(selectedPath);
  };
  const isFileDrag = (event) => {
    const types = Array.from(event?.dataTransfer?.types || []);
    return types.includes("Files");
  };
  const handleDragEnter = t0((event) => {
    if (!isFileDrag(event))
      return;
    event.preventDefault();
    dragDepthRef.current += 1;
    if (!dragActiveRef.current)
      setDragActive(true);
    setDragMode("upload");
    const target = resolveDropTargetFromEvent(event) || resolveDropTargetPath();
    updateDropTarget(target);
    scheduleDragExpand(target);
  }, [resolveDropTargetPath, resolveDropTargetFromEvent, updateDropTarget, scheduleDragExpand]);
  const handleDragOver = t0((event) => {
    if (!isFileDrag(event))
      return;
    event.preventDefault();
    if (event.dataTransfer)
      event.dataTransfer.dropEffect = "copy";
    if (!dragActiveRef.current)
      setDragActive(true);
    if (dragModeRef.current !== "upload") {
      setDragMode("upload");
    }
    const target = resolveDropTargetFromEvent(event) || resolveDropTargetPath();
    if (dropTargetRef.current !== target)
      updateDropTarget(target);
    scheduleDragExpand(target);
  }, [resolveDropTargetPath, resolveDropTargetFromEvent, updateDropTarget, scheduleDragExpand]);
  const handleDragLeave = t0((event) => {
    if (!isFileDrag(event))
      return;
    event.preventDefault();
    dragDepthRef.current = Math.max(0, dragDepthRef.current - 1);
    if (dragDepthRef.current === 0) {
      setDragActive(false);
      setDragMode(null);
      updateDropTarget(null);
      clearDragExpandTimer();
    }
  }, [updateDropTarget, clearDragExpandTimer]);
  const uploadFilesToTarget = t0(async (files, targetPath = ".") => {
    const list = Array.from(files || []);
    if (list.length === 0)
      return;
    const target = targetPath && targetPath !== "" ? targetPath : ".";
    const targetLabel = target !== "." ? target : "workspace root";
    clearUploadProgressTimer();
    setUploading(true);
    setUploadProgress({ current: 0, total: list.length, name: "", percent: 0, done: false, error: null });
    try {
      let lastResult = null;
      for (let i3 = 0;i3 < list.length; i3++) {
        const file = list[i3];
        const name = file?.name || `file ${i3 + 1}`;
        setUploadProgress((prev) => ({ ...prev, current: i3 + 1, name, percent: 0 }));
        const onProgress = (p2) => setUploadProgress((prev) => ({ ...prev, percent: p2.percent }));
        try {
          lastResult = await uploadWorkspaceFile(file, target, { onProgress });
        } catch (err2) {
          const status = err2?.status;
          const code = err2?.code;
          if (status === 409 || code === "file_exists") {
            const confirmOverwrite = window.confirm(`"${name}" already exists in ${targetLabel}. Overwrite?`);
            if (!confirmOverwrite)
              continue;
            lastResult = await uploadWorkspaceFile(file, target, { overwrite: true, onProgress });
          } else {
            throw err2;
          }
        }
      }
      if (lastResult?.path) {
        pendingProgrammaticFileClickRef.current = lastResult.path;
        setSelectedPath(lastResult.path);
        loadPreviewRef.current?.(lastResult.path);
      }
      loadSubtreeRef.current?.(target);
      refreshWorkspaceIndexStatus();
      setUploadProgress((prev) => ({ ...prev, done: true }));
      clearUploadProgressTimer();
      uploadProgressTimerRef.current = window.setTimeout(() => {
        uploadProgressTimerRef.current = 0;
        setUploadProgress(null);
      }, 1500);
    } catch (err2) {
      setError(err2.message || "Failed to upload file");
      setUploadProgress((prev) => prev ? { ...prev, error: err2.message || "Upload failed" } : null);
      clearUploadProgressTimer();
      uploadProgressTimerRef.current = window.setTimeout(() => {
        uploadProgressTimerRef.current = 0;
        setUploadProgress(null);
      }, 4000);
    } finally {
      setUploading(false);
    }
  }, [clearUploadProgressTimer]);
  const moveEntryToTarget = t0(async (sourcePath, targetPath) => {
    if (!sourcePath)
      return;
    const node = nodeMapRef.current?.get(sourcePath);
    if (!node)
      return;
    const targetDir = targetPath && targetPath !== "" ? targetPath : ".";
    const sourceParent = sourcePath.includes("/") ? sourcePath.split("/").slice(0, -1).join("/") || "." : ".";
    if (targetDir === sourceParent)
      return;
    try {
      const result = await moveWorkspaceEntry(sourcePath, targetDir);
      const nextPath = result?.path || sourcePath;
      if (node.type === "dir") {
        setExpanded((prev) => {
          const next = new Set;
          for (const entry of prev) {
            if (entry === sourcePath) {
              next.add(nextPath);
            } else if (entry.startsWith(`${sourcePath}/`)) {
              next.add(`${nextPath}${entry.slice(sourcePath.length)}`);
            } else {
              next.add(entry);
            }
          }
          return next;
        });
      }
      setSelectedPath(nextPath);
      if (node.type === "dir") {
        setPreview(null);
        setLoadingPreview(false);
        setDownloadId(null);
      } else {
        loadPreviewRef.current?.(nextPath);
      }
      loadSubtreeRef.current?.(sourceParent);
      loadSubtreeRef.current?.(targetDir);
      refreshWorkspaceIndexStatus();
    } catch (err2) {
      setError(err2?.message || "Failed to move entry");
    }
  }, []);
  moveEntryToTargetRef.current = moveEntryToTarget;
  const handleDrop = t0(async (event) => {
    if (!isFileDrag(event))
      return;
    event.preventDefault();
    dragDepthRef.current = 0;
    setDragActive(false);
    setDragMode(null);
    setDropTarget(null);
    clearDragExpandTimer();
    const files = Array.from(event?.dataTransfer?.files || []);
    if (files.length === 0)
      return;
    const target = dropTargetRef.current || resolveDropTargetFromEvent(event) || resolveDropTargetPath();
    await uploadFilesToTarget(files, target);
  }, [resolveDropTargetPath, resolveDropTargetFromEvent, uploadFilesToTarget]);
  const handleFolderUploadClick = t0((event) => {
    event?.stopPropagation?.();
    if (uploading)
      return;
    const target = event?.currentTarget?.dataset?.uploadTarget || ".";
    uploadTargetRef.current = target;
    uploadInputRef.current?.click();
  }, [uploading]);
  const handleUploadButtonClick = t0(() => {
    if (uploading)
      return;
    const selected = selectedPathRef.current;
    const selectedNode2 = selected ? nodeMapRef.current?.get(selected) : null;
    uploadTargetRef.current = selectedNode2?.type === "dir" ? selectedNode2.path : ".";
    uploadInputRef.current?.click();
  }, [uploading]);
  const handleMenuCreateFile = t0(() => {
    runMenuAction(() => handleCreateFileClick(null));
  }, [runMenuAction, handleCreateFileClick]);
  const handleMenuUploadFiles = t0(() => {
    runMenuAction(() => handleUploadButtonClick());
  }, [runMenuAction, handleUploadButtonClick]);
  const handleMenuRefresh = t0(() => {
    runMenuAction(() => handleRefreshClick());
  }, [runMenuAction, handleRefreshClick]);
  const handleMenuToggleHidden = t0(() => {
    runMenuAction(() => handleToggleHidden());
  }, [runMenuAction, handleToggleHidden]);
  const handleMenuOpenTab = t0(() => {
    if (!selectedPath || !selectedHasOpenableTab)
      return;
    runMenuAction(() => onOpenEditorRef.current?.(selectedPath, preview));
  }, [runMenuAction, selectedPath, selectedHasOpenableTab, preview]);
  const handleMenuOpenEditor = t0(() => {
    if (!selectedPath || !canEdit)
      return;
    runMenuAction(() => onOpenEditorRef.current?.(selectedPath, preview));
  }, [runMenuAction, selectedPath, canEdit, preview]);
  const handleMenuRename = t0(() => {
    if (!selectedPath || selectedPath === ".")
      return;
    runMenuAction(() => beginRename(selectedPath));
  }, [runMenuAction, selectedPath, beginRename]);
  const handleMenuDelete = t0(() => {
    if (!selectedPath || selectedIsDir)
      return;
    runMenuAction(() => handleDeleteFile());
  }, [runMenuAction, selectedPath, selectedIsDir, handleDeleteFile]);
  const handleMenuDownload = t0(() => {
    if (!selectedPath || selectedIsDir)
      return;
    runMenuAction(() => handleDownload());
  }, [runMenuAction, selectedPath, selectedIsDir, handleDownload]);
  const handleMenuDownloadFolder = t0(() => {
    if (!selectedFolderDownloadUrl)
      return;
    closeHeaderMenu();
    triggerWorkspaceDownload(selectedFolderDownloadUrl);
  }, [closeHeaderMenu, selectedFolderDownloadUrl]);
  const handleMenuOpenTerminalTab = t0(() => {
    closeHeaderMenu();
    onOpenTerminalTab?.();
  }, [closeHeaderMenu, onOpenTerminalTab]);
  const handleMenuOpenVncTab = t0(() => {
    closeHeaderMenu();
    onOpenVncTab?.();
  }, [closeHeaderMenu, onOpenVncTab]);
  const handleMenuToggleTerminal = t0(() => {
    closeHeaderMenu();
    onToggleTerminal?.();
  }, [closeHeaderMenu, onToggleTerminal]);
  const handleRowMouseDown = t0((event) => {
    if (!event || event.button !== 0)
      return;
    const rowEl = event.currentTarget;
    if (!rowEl || !rowEl.dataset)
      return;
    const path = rowEl.dataset.path;
    if (!path || path === ".")
      return;
    if (renamingPathRef.current === path)
      return;
    const targetEl = getEventTargetElement(event);
    if (targetEl?.closest?.("button, a, input, .workspace-caret"))
      return;
    if (!isRowDragHandleTarget(targetEl))
      return;
    event.preventDefault();
    mouseDragRef.current = {
      path,
      dragging: false,
      startX: event.clientX,
      startY: event.clientY
    };
    const onMove = (me) => {
      const dragState = mouseDragRef.current;
      if (!dragState?.path)
        return;
      const dx = Math.abs(me.clientX - dragState.startX);
      const dy = Math.abs(me.clientY - dragState.startY);
      const moved = dx > 4 || dy > 4;
      if (!dragState.dragging && moved) {
        dragState.dragging = true;
        suppressClickRef.current = true;
        setDragActive(true);
        setDragMode("move");
        startDragGhost(dragState.path);
        updateDragGhostPosition(me.clientX, me.clientY);
        document.body.style.userSelect = "none";
        document.body.style.cursor = "grabbing";
      }
      if (dragState.dragging) {
        me.preventDefault();
        updateDragGhostPosition(me.clientX, me.clientY);
        const el = document.elementFromPoint(me.clientX, me.clientY);
        const target = resolveDropTargetFromElement(el) || resolveDropTargetPath();
        if (dropTargetRef.current !== target)
          updateDropTarget(target);
        scheduleDragExpand(target);
      }
    };
    const onUp = () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      const dragState = mouseDragRef.current;
      if (dragState?.dragging && dragState.path) {
        const target = dropTargetRef.current || resolveDropTargetPath();
        const mover = moveEntryToTargetRef.current;
        if (typeof mover === "function")
          mover(dragState.path, target);
      }
      mouseDragRef.current = { path: null, dragging: false, startX: 0, startY: 0 };
      dragDepthRef.current = 0;
      setDragActive(false);
      setDragMode(null);
      updateDropTarget(null);
      clearDragExpandTimer();
      clearDragGhost();
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
      setTimeout(() => {
        suppressClickRef.current = false;
      }, 0);
    };
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  }, [resolveDropTargetFromElement, resolveDropTargetPath, startDragGhost, updateDragGhostPosition, clearDragGhost, updateDropTarget, scheduleDragExpand, clearDragExpandTimer]);
  const handleUploadInputChange = t0(async (event) => {
    const files = Array.from(event?.target?.files || []);
    if (files.length === 0)
      return;
    const target = uploadTargetRef.current || ".";
    await uploadFilesToTarget(files, target);
    uploadTargetRef.current = ".";
    if (event?.target)
      event.target.value = "";
  }, [uploadFilesToTarget]);
  return X1`
        <aside
            class=${`workspace-sidebar${dragActive ? " workspace-drop-active" : ""}`}
            data-workspace-scale=${explorerScale}
            ref=${sidebarRef}
            onDragEnter=${handleDragEnter}
            onDragOver=${handleDragOver}
            onDragLeave=${handleDragLeave}
            onDrop=${handleDrop}
        >
            <input type="file" multiple style="display:none" ref=${uploadInputRef} onChange=${handleUploadInputChange} />
            <div class="workspace-header">
                <div class="workspace-header-left">
                    <div class="workspace-menu-wrap">
                        <button
                            ref=${headerMenuButtonRef}
                            class=${`workspace-menu-button${headerMenuOpen ? " active" : ""}`}
                            onClick=${(e2) => {
    e2.stopPropagation();
    setHeaderMenuOpen((prev) => !prev);
  }}
                            title="Workspace actions"
                            aria-label="Workspace actions"
                            aria-haspopup="menu"
                            aria-expanded=${headerMenuOpen ? "true" : "false"}
                        >
                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                <line x1="4" y1="7" x2="20" y2="7" />
                                <line x1="4" y1="12" x2="20" y2="12" />
                                <line x1="4" y1="17" x2="20" y2="17" />
                            </svg>
                        </button>
                        ${headerMenuOpen && X1`
                            <div class="workspace-menu-dropdown" ref=${headerMenuRef} role="menu" aria-label="Workspace options">
                                <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuCreateFile} disabled=${uploading}>New file</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuUploadFiles} disabled=${uploading}>Upload files</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuRefresh}>Refresh tree</button>
                                <button class="workspace-menu-item" role="menuitem" onClick=${() => runMenuAction(() => handleWorkspaceReindex())} disabled=${workspaceReindexing}>
                                    ${workspaceReindexing ? "Reindexing workspace…" : "Reindex workspace"}
                                </button>
                                <button class=${`workspace-menu-item${showHidden ? " active" : ""}`} role="menuitem" onClick=${handleMenuToggleHidden}>
                                    ${showHidden ? "Hide hidden files" : "Show hidden files"}
                                </button>

                                ${(onOpenTerminalTab || onOpenVncTab || onToggleTerminal) && X1`<div class="workspace-menu-separator"></div>`}
                                ${onOpenTerminalTab && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenTerminalTab}>
                                        Open terminal in tab
                                    </button>
                                `}
                                ${onOpenVncTab && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenVncTab}>
                                        Open VNC in tab
                                    </button>
                                `}
                                ${onToggleTerminal && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuToggleTerminal}>
                                        ${terminalVisible ? "Hide terminal dock" : "Show terminal dock"}
                                    </button>
                                `}

                                ${selectedPath && X1`<div class="workspace-menu-separator"></div>`}
                                ${selectedHasOpenableTab && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenTab}>Open in tab</button>
                                `}
                                ${selectedPath && !selectedIsDir && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenEditor} disabled=${!canEdit}>Open in editor</button>
                                `}
                                ${selectedCanRename && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuRename}>Rename selected</button>
                                `}
                                ${selectedCanDownload && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuDownload}>Download selected file</button>
                                `}
                                ${selectedFolderDownloadUrl && X1`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuDownloadFolder}>Download selected folder (zip)</button>
                                `}
                                ${selectedCanDelete && X1`
                                    <button class="workspace-menu-item danger" role="menuitem" onClick=${handleMenuDelete}>Delete selected file</button>
                                `}
                            </div>
                        `}
                    </div>
                    <span>Workspace</span>
                </div>
                <div class="workspace-header-actions">
                    <button class="workspace-create" onClick=${handleCreateFileClick} title="New file" disabled=${uploading}>
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <line x1="12" y1="5" x2="12" y2="19" />
                            <line x1="5" y1="12" x2="19" y2="12" />
                        </svg>
                    </button>
                    <button class="workspace-refresh" onClick=${handleRefreshClick} title="Refresh tree">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"
                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <circle cx="12" cy="12" r="8.5" stroke-dasharray="42 12" stroke-dashoffset="6"
                                transform="rotate(75 12 12)" />
                            <polyline points="21 3 21 9 15 9" />
                        </svg>
                    </button>
                </div>
            </div>
            ${showWorkspaceIndexIndicator && X1`
                <div class="workspace-index-status-row">
                    <div class=${`workspace-index-status-chip state-${workspaceIndexState}`} title=${workspaceIndexTitle}>
                        <span class="workspace-index-status-dot" aria-hidden="true"></span>
                        <span>${workspaceIndexLabel}</span>
                    </div>
                </div>
            `}
            <div class="workspace-tree" onClick=${handleBackgroundClick}>
                ${uploadProgress && X1`
                    <div class="workspace-upload-strip">
                        <div class="workspace-upload-strip-text">
                            ${uploadProgress.error ? X1`<span class="workspace-upload-strip-error">${uploadProgress.error}</span>` : uploadProgress.done ? X1`<span>Done</span>` : X1`<span>${uploadProgress.total > 1 ? `Uploading ${uploadProgress.current}/${uploadProgress.total}: ${uploadProgress.name}` : `Uploading ${uploadProgress.name}`}${uploadProgress.percent > 0 ? ` (${uploadProgress.percent}%)` : "…"}</span>`}
                        </div>
                        ${!uploadProgress.done && !uploadProgress.error && X1`
                            <div class="workspace-upload-strip-bar">
                                <div class="workspace-upload-strip-fill" style=${`width:${uploadProgress.percent || 0}%`}></div>
                            </div>
                        `}
                    </div>
                `}
                ${initialLoad && X1`<div class="workspace-loading">Loading…</div>`}
                ${error && X1`<div class="workspace-error">${error}</div>`}
                ${tree && X1`
                    <div
                        class="workspace-tree-list"
                        ref=${treeListRef}
                        tabIndex="0"
                        onClick=${handleTreeClick}
                        onDblClick=${handleTreeDblClick}
                        onKeyDown=${handleTreeKeyDown}
                        onTouchStart=${handleRowTouchStart}
                        onTouchEnd=${handleRowTouchEnd}
                        onTouchMove=${handleRowTouchMove}
                        onTouchCancel=${handleRowTouchEnd}
                    >
                        ${rows.map(({ node, depth }) => {
    const isDir = node.type === "dir";
    const isSelected = node.path === selectedPath;
    const isRenaming = node.path === renamingPath;
    const isOpen = isDir && expanded.has(node.path);
    const isDropTarget = dropTarget && node.path === dropTarget;
    const childCount = Array.isArray(node.children) && node.children.length > 0 ? node.children.length : Number(node.child_count) || 0;
    return X1`
                                <div
                                    key=${node.path}
                                    class=${`workspace-row${isSelected ? " selected" : ""}${isDropTarget ? " drop-target" : ""}`}
                                    style=${{ paddingLeft: `${8 + depth * workspaceScaleMetrics.indentPx}px` }}
                                    data-path=${node.path}
                                    data-type=${node.type}
                                    onMouseDown=${handleRowMouseDown}
                                >
                                    <span class="workspace-caret" aria-hidden="true">
                                        ${isDir ? isOpen ? X1`<svg viewBox="0 0 12 12"><polygon points="1,2 11,2 6,11"/></svg>` : X1`<svg viewBox="0 0 12 12"><polygon points="2,1 11,6 2,11"/></svg>` : null}
                                    </span>
                                    <svg class=${`workspace-node-icon${isDir ? " folder" : ""}`}
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                        aria-hidden="true">
                                        ${isDir ? X1`<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>` : X1`<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>`}
                                    </svg>
                                    ${isRenaming ? X1`
                                            <input
                                                class="workspace-rename-input"
                                                ref=${renameInputRef}
                                                value=${renameValue}
                                                onInput=${(e2) => setRenameValue(e2?.target?.value || "")}
                                                onKeyDown=${(e2) => {
      e2.stopPropagation();
      if (e2.key === "Enter") {
        e2.preventDefault();
        commitRename();
      } else if (e2.key === "Escape") {
        e2.preventDefault();
        cancelRename();
      }
    }}
                                                onBlur=${cancelRename}
                                                onClick=${(e2) => e2.stopPropagation()}
                                            />
                                        ` : X1`<span class="workspace-label"><span class="workspace-label-text">${node.name}</span></span>`}
                                    ${isDir && !isOpen && childCount > 0 && X1`
                                        <span class="workspace-count">${childCount}</span>
                                    `}
                                    ${isDir && X1`
                                        <button
                                            class="workspace-folder-upload"
                                            data-upload-target=${node.path}
                                            title="Upload files to this folder"
                                            onClick=${handleFolderUploadClick}
                                            disabled=${uploading}
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
                            `;
  })}
                    </div>
                `}
            </div>
            ${selectedPath && X1`
                <div class="workspace-preview-splitter-h" onMouseDown=${handlePreviewSplitterMouseDown} onTouchStart=${handlePreviewSplitterTouchStart}></div>
                <div class="workspace-preview">
                    <div class="workspace-preview-header">
                        <span class="workspace-preview-title">${selectedPath}</span>
                        <div class="workspace-preview-actions">
                            <button class="workspace-create" onClick=${handleCreateFileClick} title="New file" disabled=${uploading}>
                                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                    <line x1="12" y1="5" x2="12" y2="19" />
                                    <line x1="5" y1="12" x2="19" y2="12" />
                                </svg>
                            </button>
                            ${!selectedIsDir && X1`
                                <button
                                    class="workspace-download workspace-edit"
                                    onClick=${() => canEdit && onOpenEditorRef.current?.(selectedPath, preview)}
                                    title=${editTitle}
                                    disabled=${!canEdit}
                                >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M12 20h9" />
                                        <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
                                    </svg>
                                </button>
                                <button
                                    class="workspace-download workspace-delete"
                                    onClick=${handleDeleteFile}
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
                            ${selectedIsDir ? X1`
                                    <button class="workspace-download" onClick=${handleUploadButtonClick}
                                        title="Upload files to this folder" disabled=${uploading}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 8 12 3 17 8"/>
                                            <line x1="12" y1="3" x2="12" y2="15"/>
                                        </svg>
                                    </button>
                                    <a class="workspace-download" href=${getWorkspaceDownloadUrl(selectedPath, showHidden)} download
                                        title="Download folder as zip" onClick=${(e2) => e2.stopPropagation()}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 10 12 15 17 10"/>
                                            <line x1="12" y1="15" x2="12" y2="3"/>
                                        </svg>
                                    </a>` : X1`<a class="workspace-download" href=${getWorkspaceFileDownloadUrl(selectedPath)} download
                                        title="Download" onClick=${(e2) => e2.stopPropagation()}>
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                        <polyline points="7 10 12 15 17 10"/>
                                        <line x1="12" y1="15" x2="12" y2="3"/>
                                    </svg>
                                </a>`}
                        </div>
                    </div>
                    ${loadingPreview && X1`<div class="workspace-loading">Loading preview…</div>`}
                    ${preview?.error && X1`<div class="workspace-error">${preview.error}</div>`}
                    ${selectedIsDir && X1`
                        <div class="workspace-preview-text">Folder selected — create file, upload files, or download as zip.</div>
                        ${folderChart?.loading && X1`<div class="workspace-loading">Loading folder size preview…</div>`}
                        ${folderChart?.error && X1`<div class="workspace-error">${folderChart.error}</div>`}
                        ${folderChart?.payload && folderChart.payload.segments?.length > 0 && X1`
                            <${FolderStarburstChart} payload=${folderChart.payload} />
                        `}
                        ${folderChart?.payload && (!folderChart.payload.segments || folderChart.payload.segments.length === 0) && X1`
                            <div class="workspace-preview-text">No file size data available for this folder yet.</div>
                        `}
                    `}
                    ${preview && !preview.error && !selectedIsDir && X1`
                        <div class="workspace-preview-body" ref=${previewPaneHostRef}></div>
                    `}
                </div>
            `}
            ${dragGhost && X1`
                <div class="workspace-drag-ghost" ref=${dragGhostRef}>${dragGhost.label}</div>
            `}
        </aside>
    `;
}

// web/src/ui/tab-source-editor.ts
var SOURCE_EDITABLE_PANE_IDS = new Set(["html-viewer", "kanban-editor", "mindmap-editor"]);
function resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane) {
  const normalized = String(path || "").trim();
  if (!normalized)
    return null;
  if (paneOverrideId)
    return paneOverrideId;
  if (typeof resolvePane !== "function")
    return null;
  const resolved = resolvePane({ path: normalized, mode: "edit" });
  return resolved?.id || null;
}
function canTabEditSource(path, paneOverrideId, resolvePane) {
  const paneId = resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane);
  return paneId != null && SOURCE_EDITABLE_PANE_IDS.has(paneId);
}
function getTabEditSourceLabel(path, paneOverrideId, resolvePane) {
  const paneId = resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane);
  return paneId === "html-viewer" ? "Edit" : "Edit Source";
}

// web/src/ui/tab-compare-saved.ts
function canTabCompareToSaved(path, paneOverrideId, resolvePane) {
  return resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane) === "editor";
}

// web/src/components/tab-strip.ts
var OFFICE_EXTENSIONS2 = /\.(docx?|xlsx?|pptx?|odt|ods|odp|rtf)$/i;
var CSV_EXTENSIONS = /\.(csv|tsv)$/i;
var PDF_EXTENSIONS = /\.pdf$/i;
var IMAGE_EXTENSIONS = /\.(png|jpe?g|gif|webp|bmp|ico|svg)$/i;
var DRAWIO_EXTENSIONS = /\.drawio(\.xml|\.svg|\.png)?$/i;
function getStandaloneTabUrl(path, { hasPopOutTab = false } = {}) {
  const normalizedPath = typeof path === "string" ? path.trim() : "";
  if (!normalizedPath)
    return null;
  if (OFFICE_EXTENSIONS2.test(normalizedPath)) {
    const rawUrl = "/workspace/raw?path=" + encodeURIComponent(normalizedPath);
    const name = normalizedPath.split("/").pop() || "document";
    return "/office-viewer/?url=" + encodeURIComponent(rawUrl) + "&name=" + encodeURIComponent(name);
  }
  if (CSV_EXTENSIONS.test(normalizedPath)) {
    return "/csv-viewer/?path=" + encodeURIComponent(normalizedPath);
  }
  if (PDF_EXTENSIONS.test(normalizedPath)) {
    return "/workspace/raw?path=" + encodeURIComponent(normalizedPath);
  }
  if (IMAGE_EXTENSIONS.test(normalizedPath) && !DRAWIO_EXTENSIONS.test(normalizedPath)) {
    return "/image-viewer/?path=" + encodeURIComponent(normalizedPath);
  }
  if (DRAWIO_EXTENSIONS.test(normalizedPath) && !hasPopOutTab) {
    return "/drawio/edit?path=" + encodeURIComponent(normalizedPath);
  }
  return null;
}
function TabStrip({ tabs, activeId, onActivate, onClose, onCloseOthers, onCloseAll, onTogglePin, onTogglePreview, onToggleDiff, onEditSource, previewTabs, diffTabs, paneOverrides, detachedTabs, onReattachTab, onToggleDock, dockVisible, onToggleZen, zenMode, onPopOutTab }) {
  const [contextMenu, setContextMenu] = w0(null);
  const stripRef = o0(null);
  r0(() => {
    if (!contextMenu)
      return;
    const dismiss = (e2) => {
      if (e2.type === "keydown" && e2.key !== "Escape")
        return;
      setContextMenu(null);
    };
    document.addEventListener("click", dismiss);
    document.addEventListener("keydown", dismiss);
    return () => {
      document.removeEventListener("click", dismiss);
      document.removeEventListener("keydown", dismiss);
    };
  }, [contextMenu]);
  r0(() => {
    const onKeyDown = (e2) => {
      if (e2.ctrlKey && e2.key === "Tab") {
        e2.preventDefault();
        if (!tabs.length)
          return;
        const idx = tabs.findIndex((t2) => t2.id === activeId);
        if (e2.shiftKey) {
          const prev = tabs[(idx - 1 + tabs.length) % tabs.length];
          onActivate?.(prev.id);
        } else {
          const next = tabs[(idx + 1) % tabs.length];
          onActivate?.(next.id);
        }
        return;
      }
      if ((e2.ctrlKey || e2.metaKey) && e2.key === "w") {
        const editorPane = document.querySelector(".editor-pane");
        if (editorPane && editorPane.contains(document.activeElement)) {
          e2.preventDefault();
          if (activeId)
            onClose?.(activeId);
        }
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [tabs, activeId, onActivate, onClose]);
  const handleTabMouseDown = t0((e2, id) => {
    if (e2.button === 1) {
      e2.preventDefault();
      onClose?.(id);
    }
  }, [onClose]);
  const handleTabClick = t0((e2, id) => {
    if (e2.defaultPrevented)
      return;
    if (e2.button === 0) {
      onActivate?.(id);
    }
  }, [onActivate]);
  const handleContextMenu = t0((e2, id) => {
    e2.preventDefault();
    setContextMenu({ id, x: e2.clientX, y: e2.clientY });
  }, []);
  const handleClosePointerDown = t0((e2) => {
    e2.preventDefault();
    e2.stopPropagation();
  }, []);
  const handleCloseClick = t0((e2, id) => {
    e2.preventDefault();
    e2.stopPropagation();
    onClose?.(id);
  }, [onClose]);
  r0(() => {
    if (!activeId || !stripRef.current)
      return;
    const activeEl = stripRef.current.querySelector(".tab-item.active");
    if (activeEl) {
      activeEl.scrollIntoView({ block: "nearest", inline: "nearest", behavior: "smooth" });
    }
  }, [activeId]);
  const getPaneOverride = t0((id) => {
    if (!(paneOverrides instanceof Map))
      return null;
    return paneOverrides.get(id) || null;
  }, [paneOverrides]);
  const contextMenuTab = G0(() => tabs.find((tab) => tab.id === contextMenu?.id) || null, [contextMenu?.id, tabs]);
  const contextMenuCanEditSource = G0(() => {
    const tabId = contextMenu?.id;
    if (!tabId)
      return false;
    return canTabEditSource(tabId, getPaneOverride(tabId), (context) => paneRegistry.resolve(context));
  }, [contextMenu?.id, getPaneOverride]);
  const contextMenuEditSourceLabel = G0(() => {
    const tabId = contextMenu?.id;
    if (!tabId)
      return "Edit Source";
    return getTabEditSourceLabel(tabId, getPaneOverride(tabId), (context) => paneRegistry.resolve(context));
  }, [contextMenu?.id, getPaneOverride]);
  const isContextMenuTabDetached = G0(() => {
    const tabId = contextMenu?.id;
    if (!tabId || !(detachedTabs instanceof Map))
      return false;
    return detachedTabs.has(tabId);
  }, [contextMenu?.id, detachedTabs]);
  const contextMenuDiffOpen = G0(() => {
    const tabId = contextMenu?.id;
    if (!tabId || !(diffTabs instanceof Set))
      return false;
    return diffTabs.has(tabId);
  }, [contextMenu?.id, diffTabs]);
  const contextMenuCanCompareToSaved = G0(() => {
    const tabId = contextMenu?.id;
    if (!tabId)
      return false;
    const tab = tabs.find((item) => item.id === tabId) || null;
    if (!tab)
      return false;
    const supportsCompare = canTabCompareToSaved(tabId, getPaneOverride(tabId), (context) => paneRegistry.resolve(context));
    return supportsCompare && Boolean(tab.dirty || contextMenuDiffOpen);
  }, [contextMenu?.id, contextMenuDiffOpen, getPaneOverride, tabs]);
  if (!tabs.length)
    return null;
  return X1`
        <div class="tab-strip" ref=${stripRef} role="tablist">
            ${tabs.map((tab) => X1`
                <div
                    key=${tab.id}
                    class=${`tab-item${tab.id === activeId ? " active" : ""}${tab.dirty ? " dirty" : ""}${tab.pinned ? " pinned" : ""}`}
                    role="tab"
                    aria-selected=${tab.id === activeId}
                    title=${tab.path}
                    onMouseDown=${(e2) => handleTabMouseDown(e2, tab.id)}
                    onClick=${(e2) => handleTabClick(e2, tab.id)}
                    onContextMenu=${(e2) => handleContextMenu(e2, tab.id)}
                >
                    ${tab.pinned && X1`
                        <span class="tab-pin-icon" aria-label="Pinned">
                            <svg viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
                                <path d="M4.456.734a1.75 1.75 0 0 1 2.826.504l.613 1.327a3.1 3.1 0 0 0 2.084 1.707l2.454.584c1.332.317 1.8 1.972.832 2.94L11.06 10l3.72 3.72a.75.75 0 1 1-1.06 1.06L10 11.06l-2.204 2.205c-.968.968-2.623.5-2.94-.832l-.584-2.454a3.1 3.1 0 0 0-1.707-2.084l-1.327-.613a1.75 1.75 0 0 1-.504-2.826z"/>
                            </svg>
                        </span>
                    `}
                    <span class="tab-label">${tab.label}</span>
                    ${detachedTabs instanceof Map && detachedTabs.has(tab.id) && X1`
                        <span class="tab-detached-badge" aria-label="Detached" title="Open in separate window">↗</span>
                    `}
                    <button
                        type="button"
                        class="tab-close"
                        onPointerDown=${handleClosePointerDown}
                        onMouseDown=${handleClosePointerDown}
                        onClick=${(e2) => handleCloseClick(e2, tab.id)}
                        title=${tab.dirty ? "Unsaved changes" : "Close"}
                        aria-label=${tab.dirty ? "Unsaved changes" : `Close ${tab.label}`}
                    >
                        ${tab.dirty ? X1`<span class="tab-dirty-dot" aria-hidden="true"></span>` : X1`<svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true" focusable="false" style=${{ pointerEvents: "none" }}>
                                <line x1="4" y1="4" x2="12" y2="12" style=${{ pointerEvents: "none" }}/>
                                <line x1="12" y1="4" x2="4" y2="12" style=${{ pointerEvents: "none" }}/>
                            </svg>`}
                    </button>
                </div>
            `)}
            ${onToggleDock && X1`
                <div class="tab-strip-spacer"></div>
                <button
                    class=${`tab-strip-dock-toggle${dockVisible ? " active" : ""}`}
                    onClick=${onToggleDock}
                    title=${`${dockVisible ? "Hide" : "Show"} terminal (Ctrl+\`)`}
                    aria-label=${`${dockVisible ? "Hide" : "Show"} terminal`}
                    aria-pressed=${dockVisible ? "true" : "false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="1.75" y="2.25" width="12.5" height="11.5" rx="2"/>
                        <polyline points="4.5 5.25 7 7.75 4.5 10.25"/>
                        <line x1="8.5" y1="10.25" x2="11.5" y2="10.25"/>
                    </svg>
                </button>
            `}
            ${onToggleZen && X1`
                <button
                    class=${`tab-strip-zen-toggle${zenMode ? " active" : ""}`}
                    onClick=${onToggleZen}
                    title=${`${zenMode ? "Exit" : "Enter"} zen mode (Ctrl+Shift+Z)`}
                    aria-label=${`${zenMode ? "Exit" : "Enter"} zen mode`}
                    aria-pressed=${zenMode ? "true" : "false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        ${zenMode ? X1`<polyline points="4 8 1.5 8 1.5 1.5 14.5 1.5 14.5 8 12 8"/><polyline points="4 8 1.5 8 1.5 14.5 14.5 14.5 14.5 8 12 8"/>` : X1`<polyline points="5.5 1.5 1.5 1.5 1.5 5.5"/><polyline points="10.5 1.5 14.5 1.5 14.5 5.5"/><polyline points="5.5 14.5 1.5 14.5 1.5 10.5"/><polyline points="10.5 14.5 14.5 14.5 14.5 10.5"/>`}
                    </svg>
                </button>
            `}
        </div>
        ${contextMenu && X1`
            <div class="tab-context-menu" style=${{ left: contextMenu.x + "px", top: contextMenu.y + "px" }}>
                <button onClick=${() => {
    onClose?.(contextMenu.id);
    setContextMenu(null);
  }}>Close</button>
                <button onClick=${() => {
    onCloseOthers?.(contextMenu.id);
    setContextMenu(null);
  }}>Close Others</button>
                <button onClick=${() => {
    onCloseAll?.();
    setContextMenu(null);
  }}>Close All</button>
                <hr />
                <button onClick=${() => {
    onTogglePin?.(contextMenu.id);
    setContextMenu(null);
  }}>
                    ${contextMenuTab?.pinned ? "Unpin" : "Pin"}
                </button>
                ${contextMenuCanEditSource && onEditSource && X1`
                    <button onClick=${() => {
    onEditSource(contextMenu.id);
    setContextMenu(null);
  }}>${contextMenuEditSourceLabel}</button>
                `}
                ${isContextMenuTabDetached && onReattachTab && X1`
                    <button onClick=${() => {
    onReattachTab(contextMenu.id);
    setContextMenu(null);
  }}>Reattach</button>
                `}
                ${onPopOutTab && !isContextMenuTabDetached && X1`
                    <button onClick=${() => {
    const tab = tabs.find((t2) => t2.id === contextMenu.id);
    onPopOutTab(contextMenu.id, tab?.label);
    setContextMenu(null);
  }}>Open in Window</button>
                `}
                ${contextMenuCanCompareToSaved && onToggleDiff && X1`
                    <hr />
                    <button onClick=${() => {
    onActivate?.(contextMenu.id);
    onToggleDiff(contextMenu.id);
    setContextMenu(null);
  }}>${contextMenuDiffOpen ? "Hide Diff" : "Compare to Saved"}</button>
                `}
                ${onTogglePreview && /\.(md|mdx|markdown)$/i.test(contextMenu.id) && X1`
                    <hr />
                    <button onClick=${() => {
    onTogglePreview(contextMenu.id);
    setContextMenu(null);
  }}>
                        ${previewTabs?.has(contextMenu.id) ? "Hide Preview" : "Preview"}
                    </button>
                `}
                ${(() => {
    const standaloneUrl = getStandaloneTabUrl(contextMenu.id, {
      hasPopOutTab: typeof onPopOutTab === "function"
    });
    if (!standaloneUrl)
      return null;
    return X1`
                        <hr />
                        <button onClick=${() => {
      window.open(standaloneUrl, "_blank", "noopener");
      setContextMenu(null);
    }}>Open in New Tab</button>
                    `;
  })()}
            </div>
        `}
    `;
}

// web/src/components/generated-widget-host-bridge.ts
function setIframeNameBestEffort(iframe, hostName) {
  try {
    if (iframe) {
      iframe.name = hostName;
    }
    return true;
  } catch (_error) {
    return false;
  }
}
function postIframeMessageBestEffort(iframe, message) {
  try {
    iframe?.contentWindow?.postMessage?.(message, "*");
    return true;
  } catch (_error) {
    return false;
  }
}

// web/src/components/session-tree-widget.ts
function buildTreeFromFlat(flatNodes) {
  const byId = new Map;
  const roots = [];
  for (const node of flatNodes || []) {
    byId.set(node.id, { ...node, children: [], depth: 0 });
  }
  for (const node of flatNodes || []) {
    const current = byId.get(node.id);
    if (!current)
      continue;
    const parent = node.parentId ? byId.get(node.parentId) : null;
    if (parent)
      parent.children.push(current);
    else
      roots.push(current);
  }
  const folded = new Set;
  for (const [, node] of byId) {
    if (node.role !== "assistant" || !node.toolName)
      continue;
    if (node.children.length !== 1)
      continue;
    const child = node.children[0];
    if (child.role !== "toolResult")
      continue;
    node.resultDetail = child.detail || null;
    node.resultLength = child.contentLength || 0;
    node.resultId = child.id;
    node.merged = true;
    node.children = child.children;
    for (const gc of node.children)
      gc.parentId = node.id;
    folded.add(child.id);
  }
  const filteredRoots = roots.filter((r2) => !folded.has(r2.id));
  const stack = [];
  for (let i3 = filteredRoots.length - 1;i3 >= 0; i3--) {
    filteredRoots[i3].depth = 0;
    stack.push(filteredRoots[i3]);
  }
  while (stack.length > 0) {
    const node = stack.pop();
    const isBranch = node.children.length > 1;
    for (let i3 = node.children.length - 1;i3 >= 0; i3--) {
      node.children[i3].depth = isBranch ? node.depth + 1 : node.depth;
      stack.push(node.children[i3]);
    }
  }
  return filteredRoots;
}
function flattenTree2(roots) {
  const result = [];
  const stack = [];
  for (let i3 = roots.length - 1;i3 >= 0; i3--)
    stack.push(roots[i3]);
  while (stack.length > 0) {
    const node = stack.pop();
    result.push(node);
    for (let i3 = node.children.length - 1;i3 >= 0; i3--)
      stack.push(node.children[i3]);
  }
  return result;
}
function formatSize(chars) {
  if (!chars || chars <= 0)
    return "";
  if (chars < 1000)
    return `${chars}`;
  if (chars < 1e6)
    return `${(chars / 1000).toFixed(1)}k`;
  return `${(chars / 1e6).toFixed(1)}M`;
}
function formatSizeLong(chars) {
  if (!chars || chars <= 0)
    return "";
  if (chars < 1000)
    return `${chars} chars`;
  if (chars < 1e6)
    return `${(chars / 1000).toFixed(1)}k chars`;
  return `${(chars / 1e6).toFixed(1)}M chars`;
}
function getRowDisplay(node) {
  const type = node.type;
  if (type === "model_change")
    return { tag: "model", tagClass: "system", text: `${node.preview?.replace("[model ", "").replace("]", "") || ""}` };
  if (type === "thinking_level_change")
    return { tag: "thinking", tagClass: "system", text: node.preview?.replace("[thinking ", "").replace("]", "") || "" };
  if (type === "compaction")
    return { tag: "compaction", tagClass: "system", text: node.preview?.replace("[compaction: ", "").replace("]", "") || "" };
  if (type === "label")
    return { tag: "label", tagClass: "system", text: node.preview?.replace("[label ", "").replace("]", "") || "" };
  if (type === "session_info")
    return { tag: "session", tagClass: "system", text: node.preview?.replace("[session name ", "").replace("]", "") || "" };
  if (type === "branch_summary")
    return { tag: "summary", tagClass: "system", text: node.preview || "" };
  if (type !== "message")
    return { tag: type || "?", tagClass: "system", text: node.preview || "" };
  const role = node.role || "message";
  if (node.merged && node.toolName) {
    const cmd = node.toolInput || "";
    const firstLine = cmd.split(`
`)[0];
    const truncCmd = firstLine.length > 120 ? firstLine.slice(0, 119) + "…" : firstLine;
    return { tag: node.toolName, tagClass: "tool", text: truncCmd || "" };
  }
  if (role === "assistant" && node.toolName) {
    const cmd = node.toolInput || "";
    const firstLine = cmd.split(`
`)[0];
    const truncCmd = firstLine.length > 120 ? firstLine.slice(0, 119) + "…" : firstLine;
    return { tag: node.toolName, tagClass: "tool", text: truncCmd || "…" };
  }
  if (role === "toolResult") {
    const out = node.detail || "";
    const firstLine = out.split(`
`)[0];
    const trunc = firstLine.length > 120 ? firstLine.slice(0, 119) + "…" : firstLine;
    return { tag: `→ ${node.toolName || "result"}`, tagClass: "result", text: trunc };
  }
  if (role === "user") {
    const raw = node.previewText || node.detail || node.preview || "";
    const cleaned = raw.replace(/^user:\s*"?/, "").replace(/"?\s*$/, "");
    const firstLine = cleaned.split(`
`)[0];
    const trunc = firstLine.length > 120 ? firstLine.slice(0, 119) + "…" : firstLine;
    return { tag: "user", tagClass: "user", text: trunc };
  }
  if (role === "assistant") {
    const raw = node.detail || node.preview || "";
    const cleaned = raw.replace(/^assistant:\s*"?/, "").replace(/"?\s*$/, "");
    const firstLine = cleaned.split(`
`)[0];
    const trunc = firstLine.length > 120 ? firstLine.slice(0, 119) + "…" : firstLine;
    return { tag: "assistant", tagClass: "assistant", text: trunc };
  }
  return { tag: role, tagClass: "other", text: node.preview || "" };
}
function buildTreeNavigationPayload(targetId, summarize = false) {
  const cleanTargetId = typeof targetId === "string" ? targetId.trim() : "";
  if (!cleanTargetId)
    return null;
  return {
    text: summarize ? `/tree ${cleanTargetId} --summarize` : `/tree ${cleanTargetId}`,
    navigateTargetId: cleanTargetId,
    summarize: Boolean(summarize)
  };
}
function parseTreeNavigationCommand(text) {
  const raw = typeof text === "string" ? text.trim() : "";
  if (!raw.startsWith("/tree"))
    return null;
  const parts = raw.split(/\s+/).filter(Boolean);
  if (parts[0] !== "/tree")
    return null;
  let targetId = null;
  let summarize = false;
  for (let i3 = 1;i3 < parts.length; i3++) {
    const part = parts[i3];
    if (part === "--summarize") {
      summarize = true;
      continue;
    }
    if (!targetId && !part.startsWith("--")) {
      targetId = part;
    }
  }
  return targetId ? { targetId, summarize } : null;
}
function resolveTreeSelectionId(rows, currentSelectedId, preferredId, leafId) {
  const list = Array.isArray(rows) ? rows : [];
  if (list.length === 0)
    return null;
  const has = (id) => typeof id === "string" && list.some((row) => row?.id === id);
  if (has(currentSelectedId))
    return currentSelectedId;
  if (has(preferredId))
    return preferredId;
  if (has(leafId))
    return leafId;
  const active = list.find((row) => row?.active);
  if (active?.id)
    return active.id;
  return list[0]?.id || null;
}
function describeSessionTreeHostUpdate(update) {
  if (!update || typeof update !== "object")
    return null;
  const type = typeof update.type === "string" ? update.type : "";
  const preview = typeof update.preview === "string" ? update.preview.trim() : "";
  const error = typeof update.error === "string" ? update.error.trim() : "";
  const parsed = parseTreeNavigationCommand(preview);
  const label = preview || "tree command";
  if (type === "submit_pending") {
    return { tone: "pending", text: parsed ? `Sending ${label}` : "Sending tree command…", refreshDelays: [] };
  }
  if (type === "submit_sent") {
    return {
      tone: "info",
      text: parsed ? `Running ${label}` : "Tree command sent.",
      refreshDelays: parsed ? [500, 1500, 3500, 8000] : []
    };
  }
  if (type === "submit_queued") {
    return {
      tone: "info",
      text: parsed ? `Queued ${label}` : "Tree command queued.",
      refreshDelays: parsed ? [1200, 3200, 7000, 12000] : []
    };
  }
  if (type === "submit_failed") {
    return { tone: "error", text: error ? `Tree command failed: ${error}` : "Tree command failed.", refreshDelays: [] };
  }
  if (type === "refresh_building") {
    return { tone: "pending", text: "Refreshing widget…", refreshDelays: [] };
  }
  if (type === "refresh_failed") {
    return { tone: "error", text: error ? `Refresh failed: ${error}` : "Refresh failed.", refreshDelays: [] };
  }
  if (type === "refresh_dashboard" || type === "refresh_ack") {
    return { tone: "success", text: "Widget refreshed.", refreshDelays: [] };
  }
  return null;
}
function SessionTreeWidget({ widget, onWidgetEvent }) {
  const initialTree = widget?.artifact?.tree && typeof widget.artifact.tree === "object" ? widget.artifact.tree : null;
  const chatJid = typeof widget?.originChatJid === "string" && widget.originChatJid.trim() ? widget.originChatJid.trim() : null;
  const runtimeState = widget?.runtimeState && typeof widget.runtimeState === "object" ? widget.runtimeState : null;
  const hostUpdate = runtimeState?.lastHostUpdate && typeof runtimeState.lastHostUpdate === "object" ? runtimeState.lastHostUpdate : null;
  const [state, setState] = w0(() => ({ loading: !initialTree, error: null, data: initialTree }));
  const [selectedId, setSelectedId] = w0(() => initialTree?.leafId || null);
  const [searchFilter, setSearchFilter] = w0("");
  const searchInputRef = o0(null);
  const activeRowRef = o0(null);
  const preferredSelectionRef = o0(initialTree?.leafId || null);
  const loadTreeRef = o0(null);
  const scheduledRefreshKeyRef = o0("");
  const loadTree = async () => {
    setState((current) => ({ ...current, loading: true, error: null }));
    try {
      const qs = chatJid ? `?chat_jid=${encodeURIComponent(chatJid)}` : "";
      const response = await fetch(`/agent/session-tree${qs}`, { method: "GET", credentials: "same-origin", headers: { Accept: "application/json" } });
      const payload = await response.json().catch(() => ({}));
      if (!response.ok)
        throw new Error(payload?.error || `HTTP ${response.status}`);
      setState({ loading: false, error: null, data: payload });
    } catch (error) {
      setState((current) => ({ loading: false, error: error?.message || "Failed to load tree.", data: current?.data || initialTree || null }));
    }
  };
  loadTreeRef.current = loadTree;
  r0(() => {
    loadTree();
  }, [chatJid]);
  const flatRows = G0(() => {
    const data = state.data;
    if (!data || !Array.isArray(data.nodes) || data.nodes.length === 0)
      return [];
    return flattenTree2(data.flat ? buildTreeFromFlat(data.nodes) : data.nodes);
  }, [state.data]);
  r0(() => {
    const nextSelectedId = resolveTreeSelectionId(flatRows, selectedId, preferredSelectionRef.current, state.data?.leafId || null);
    if (nextSelectedId !== selectedId) {
      setSelectedId(nextSelectedId);
    }
    if (preferredSelectionRef.current && state.data?.leafId === preferredSelectionRef.current) {
      preferredSelectionRef.current = null;
    }
  }, [flatRows, selectedId, state.data?.leafId]);
  const filteredRows = G0(() => {
    const q = (searchFilter || "").trim().toLowerCase();
    if (!q)
      return flatRows;
    return flatRows.filter((node) => {
      const fields = [
        node.preview,
        node.toolInput,
        node.toolInputFull,
        node.detail,
        node.toolName,
        node.role,
        node.id,
        node.resultDetail,
        node.type,
        node.label
      ];
      return fields.some((f2) => typeof f2 === "string" && f2.toLowerCase().includes(q));
    });
  }, [flatRows, searchFilter]);
  const selectedNode = G0(() => filteredRows.find((n2) => n2.id === selectedId) || null, [filteredRows, selectedId]);
  const hostUpdateSummary = G0(() => describeSessionTreeHostUpdate(hostUpdate), [
    hostUpdate?.type,
    hostUpdate?.preview,
    hostUpdate?.error,
    hostUpdate?.submittedAt
  ]);
  r0(() => {
    if (activeRowRef.current)
      activeRowRef.current.scrollIntoView({ block: "center", behavior: "auto" });
  }, [selectedId, state.data?.leafId, filteredRows.length]);
  r0(() => {
    const parsed = parseTreeNavigationCommand(hostUpdate?.preview);
    if (parsed?.targetId) {
      preferredSelectionRef.current = parsed.targetId;
    }
    const refreshDelays = hostUpdateSummary?.refreshDelays || [];
    if (!refreshDelays.length)
      return;
    const refreshKey = [chatJid || "", hostUpdate?.type || "", hostUpdate?.submittedAt || "", hostUpdate?.preview || ""].join("|");
    if (scheduledRefreshKeyRef.current === refreshKey)
      return;
    scheduledRefreshKeyRef.current = refreshKey;
    const timers = refreshDelays.map((delay) => setTimeout(() => loadTreeRef.current?.(), delay));
    return () => timers.forEach((timer) => clearTimeout(timer));
  }, [chatJid, hostUpdate?.type, hostUpdate?.submittedAt, hostUpdate?.preview, hostUpdateSummary?.refreshDelays]);
  const submitNavigation = (summarize = false) => {
    const targetId = selectedNode?.id;
    const payload = buildTreeNavigationPayload(targetId, summarize);
    if (!payload)
      return;
    preferredSelectionRef.current = payload.navigateTargetId;
    onWidgetEvent?.({ kind: "widget.submit", payload }, widget);
  };
  const hostUpdateTone = hostUpdateSummary?.tone || "info";
  return X1`
        <div class="session-tree-widget">
            <div class="session-tree-toolbar">
                <div class="session-tree-toolbar-left">
                    <button class="session-tree-btn" type="button" onClick=${() => loadTree()} disabled=${state.loading}>${state.loading ? "Loading…" : "Refresh"}</button>
                    <input ref=${searchInputRef}
                        class="st-search-input" type="text" placeholder="Filter\u2026"
                        value=${searchFilter}
                        onInput=${(e2) => setSearchFilter(e2.currentTarget.value)}
                        onKeyDown=${(e2) => {
    if (e2.key === "Escape") {
      setSearchFilter("");
      e2.currentTarget.blur();
    }
  }}
                    />
                    ${searchFilter && X1`<span class="session-tree-meta">${filteredRows.length} match${filteredRows.length !== 1 ? "es" : ""}</span>`}
                    ${state.error && X1`<span class="session-tree-error-inline">${state.error}</span>`}
                </div>
                <div class="session-tree-toolbar-right">
                    ${hostUpdateSummary?.text && X1`<span class=${`session-tree-host-update ${hostUpdateTone}`}>${hostUpdateSummary.text}</span>`}
                    ${state.data?.capped && X1`<span class="session-tree-meta">Showing ${state.data?.nodes?.length || 0} of ${state.data?.total || 0}</span>`}
                    ${chatJid && X1`<span class="session-tree-meta">${chatJid}</span>`}
                </div>
            </div>

            <div class="session-tree-content">
                <div class="session-tree-list" role="tree" aria-label="Session tree">
                    ${state.loading && filteredRows.length === 0 && !searchFilter && X1`<div class="session-tree-empty">Loading session tree\u2026</div>`}
                    ${!state.loading && filteredRows.length === 0 && !searchFilter && X1`<div class="session-tree-empty">Session tree is empty.</div>`}
                    ${!state.loading && filteredRows.length === 0 && searchFilter && X1`<div class="session-tree-empty">No entries match \u201c${searchFilter}\u201d</div>`}
                    ${filteredRows.map((node) => {
    const sel = selectedId === node.id;
    const rowClass = `st-row${node.active ? " active" : ""}${sel ? " selected" : ""}`;
    const hasBranch = (node.children || []).length > 1;
    const d2 = getRowDisplay(node);
    return X1`
                            <button key=${node.id} ref=${node.active || sel ? activeRowRef : null}
                                class=${rowClass} type="button" role="treeitem" aria-selected=${sel}
                                onClick=${() => setSelectedId(node.id)}>
                                <span class="st-indent" style=${`width:${(node.depth || 0) * 16 + 6}px`}></span>
                                <span class=${`st-dot${node.active ? " active" : hasBranch ? " branch" : ""}`}></span>
                                <span class=${`st-tag ${d2.tagClass}`}>${d2.tag}</span>
                                <span class="st-text">${d2.text}</span>
                                ${node.merged && node.resultLength > 0 && X1`<span class="st-size">${formatSize(node.resultLength)}</span>`}
                                ${!node.merged && node.contentLength > 3000 && X1`<span class="st-size">${formatSize(node.contentLength)}</span>`}
                                ${node.hasThinking && X1`<span class="st-badge thinking">\u{1F4AD}</span>`}
                                ${node.label && X1`<span class="st-label">${node.label}</span>`}
                                ${node.active && X1`<span class="st-active">\u25C0</span>`}
                            </button>
                        `;
  })}
                </div>

                <aside class="session-tree-sidebar">
                    ${selectedNode ? X1`
                        <div class="st-side-section">
                            <div class="st-side-label">Entry</div>
                            <div class="st-side-mono">${selectedNode.id}${selectedNode.resultId ? ` → ${selectedNode.resultId}` : ""}</div>
                        </div>
                        <div class="st-side-section">
                            <div class="st-side-label">Type</div>
                            <div class="st-side-value">${selectedNode.role || selectedNode.type || "entry"}${selectedNode.toolName ? ` → ${selectedNode.toolName}` : ""}${selectedNode.merged ? " (merged)" : ""}</div>
                        </div>
                        ${selectedNode.toolInputFull && X1`
                            <div class="st-side-section">
                                <div class="st-side-label">${selectedNode.toolName === "bash" ? "Command" : "Input"}</div>
                                <pre class="st-side-code">${selectedNode.toolInputFull}</pre>
                            </div>
                        `}
                        ${selectedNode.resultDetail && X1`
                            <div class="st-side-section">
                                <div class="st-side-label">Result${selectedNode.resultLength ? ` (${formatSizeLong(selectedNode.resultLength)})` : ""}</div>
                                <pre class="st-side-code">${selectedNode.resultDetail}</pre>
                            </div>
                        `}
                        ${selectedNode.detail && !selectedNode.toolInput && X1`
                            <div class="st-side-section">
                                <div class="st-side-label">${selectedNode.role === "toolResult" ? "Output" : "Content"}${selectedNode.contentLength ? ` (${formatSizeLong(selectedNode.contentLength)})` : ""}</div>
                                <pre class="st-side-code">${selectedNode.detail}</pre>
                            </div>
                        `}
                        ${selectedNode.rawDetail && X1`
                            <div class="st-side-section">
                                <div class="st-side-label">Raw prompt${selectedNode.rawContentLength ? ` (${formatSizeLong(selectedNode.rawContentLength)})` : ""}</div>
                                <pre class="st-side-code">${selectedNode.rawDetail}</pre>
                            </div>
                        `}
                        ${selectedNode.timestamp && X1`
                            <div class="st-side-section">
                                <div class="st-side-label">Time</div>
                                <div class="st-side-value">${new Date(selectedNode.timestamp).toLocaleString()}</div>
                            </div>
                        `}
                        ${(selectedNode.contentLength > 0 || selectedNode.hasThinking) && X1`
                            <div class="st-side-section">
                                <div class="st-side-label">Size</div>
                                <div class="st-side-badges">
                                    ${selectedNode.contentLength > 0 && X1`<span class="st-pill">${formatSizeLong(selectedNode.contentLength)} content</span>`}
                                    ${selectedNode.hasThinking && X1`<span class="st-pill thinking">${formatSizeLong(selectedNode.thinkingLength)} thinking</span>`}
                                    ${selectedNode.merged && selectedNode.resultLength > 0 && X1`<span class="st-pill">${formatSizeLong(selectedNode.resultLength)} result</span>`}
                                </div>
                            </div>
                        `}
                        <div class="st-side-actions">
                            <button class="session-tree-btn primary" type="button" onClick=${() => submitNavigation(false)}>Navigate here</button>
                            <button class="session-tree-btn" type="button" onClick=${() => submitNavigation(true)}>Navigate + summarize</button>
                        </div>
                    ` : X1`<div class="session-tree-empty side">Select an entry to inspect.</div>`}
                </aside>
            </div>
        </div>
    `;
}

// web/src/components/floating-widget-pane.ts
function FloatingWidgetPane({ widget, onClose, onWidgetEvent }) {
  const frameRef = o0(null);
  const frameLoadedRef = o0(false);
  const srcDoc = G0(() => buildWidgetSrcDoc(widget), [
    widget?.artifact?.kind,
    widget?.artifact?.html,
    widget?.artifact?.svg,
    widget?.widgetId,
    widget?.toolCallId,
    widget?.turnId,
    widget?.title
  ]);
  r0(() => {
    if (!widget)
      return;
    const handleEsc = (e2) => {
      if (e2.key === "Escape")
        onClose?.();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [widget, onClose]);
  r0(() => {
    frameLoadedRef.current = false;
  }, [srcDoc]);
  r0(() => {
    if (!widget)
      return;
    const iframe = frameRef.current;
    if (!iframe)
      return;
    const postToFrame = (type) => {
      const hostName = getGeneratedWidgetHostWindowName(widget);
      const payload = type === "widget.init" ? getGeneratedWidgetInitPayload(widget) : getGeneratedWidgetHostPayload(widget);
      setIframeNameBestEffort(iframe, hostName);
      postIframeMessageBestEffort(iframe, {
        __piclawGeneratedWidgetHost: true,
        type,
        widgetId: widget?.widgetId || null,
        toolCallId: widget?.toolCallId || null,
        turnId: widget?.turnId || null,
        payload
      });
    };
    const syncHostState = () => {
      postToFrame("widget.init");
      postToFrame("widget.update");
    };
    const handleLoad = () => {
      frameLoadedRef.current = true;
      syncHostState();
    };
    iframe.addEventListener("load", handleLoad);
    const retryDelays = [0, 40, 120, 300, 800];
    const retryTimers = retryDelays.map((delay) => setTimeout(syncHostState, delay));
    return () => {
      iframe.removeEventListener("load", handleLoad);
      retryTimers.forEach((timer) => clearTimeout(timer));
    };
  }, [srcDoc, widget?.widgetId, widget?.toolCallId, widget?.turnId]);
  r0(() => {
    if (!widget)
      return;
    const iframe = frameRef.current;
    if (!iframe?.contentWindow)
      return;
    const hostName = getGeneratedWidgetHostWindowName(widget);
    const payload = getGeneratedWidgetHostPayload(widget);
    setIframeNameBestEffort(iframe, hostName);
    postIframeMessageBestEffort(iframe, {
      __piclawGeneratedWidgetHost: true,
      type: "widget.update",
      widgetId: widget?.widgetId || null,
      toolCallId: widget?.toolCallId || null,
      turnId: widget?.turnId || null,
      payload
    });
    return;
  }, [
    widget?.widgetId,
    widget?.toolCallId,
    widget?.turnId,
    widget?.status,
    widget?.subtitle,
    widget?.description,
    widget?.error,
    widget?.width,
    widget?.height,
    widget?.runtimeState
  ]);
  r0(() => {
    if (!widget)
      return;
    const handleMessage = (event) => {
      const data = event?.data;
      if (!data || data.__piclawGeneratedWidget !== true)
        return;
      const iframe = frameRef.current;
      const currentKey = getGeneratedWidgetSessionKey(widget);
      const incomingKey = getGeneratedWidgetSessionKey({
        widgetId: data.widgetId,
        toolCallId: data.toolCallId
      });
      if (incomingKey && currentKey && incomingKey !== currentKey)
        return;
      if (!incomingKey && iframe?.contentWindow && event.source !== iframe.contentWindow)
        return;
      onWidgetEvent?.(data, widget);
    };
    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [widget, onWidgetEvent]);
  if (!widget)
    return null;
  const artifact = widget?.artifact || {};
  const kind = artifact.kind || widget?.kind || "html";
  const title = typeof widget?.title === "string" && widget.title.trim() ? widget.title.trim() : "Generated widget";
  const subtitle = typeof widget?.subtitle === "string" && widget.subtitle.trim() ? widget.subtitle.trim() : "";
  const source = widget?.source === "live" ? "live" : "timeline";
  const status = typeof widget?.status === "string" && widget.status.trim() ? widget.status.trim() : "final";
  const originLabel = source === "live" ? `Live widget • ${status.toUpperCase()}` : widget?.originPostId ? `Message #${widget.originPostId}` : "Timeline launch";
  const description = typeof widget?.description === "string" && widget.description.trim() ? widget.description.trim() : "";
  const emptyState = !srcDoc;
  const emptyMessage = getGeneratedWidgetEmptyStateMessage(widget);
  const sandbox = getGeneratedWidgetIframeSandbox(widget);
  return X1`
        <div class="floating-widget-backdrop" onClick=${() => onClose?.()}>
            <section
                class="floating-widget-pane"
                aria-label=${title}
                onClick=${(e2) => e2.stopPropagation()}
            >
                <div class="floating-widget-header">
                    <div class="floating-widget-heading">
                        <div class="floating-widget-eyebrow">${originLabel} • ${kind.toUpperCase()}</div>
                        <div class="floating-widget-title">${title}</div>
                        ${(subtitle || description) && X1`
                            <div class="floating-widget-subtitle">${subtitle || description}</div>
                        `}
                    </div>
                    <button
                        class="floating-widget-close"
                        type="button"
                        onClick=${() => onClose?.()}
                        title="Close widget"
                        aria-label="Close widget"
                    >
                        Close
                    </button>
                </div>
                <div class="floating-widget-body">
                    ${kind === "session_tree" ? X1`<${SessionTreeWidget} widget=${widget} onWidgetEvent=${onWidgetEvent} />` : emptyState ? X1`<div class="floating-widget-empty">${emptyMessage}</div>` : X1`
                                <iframe
                                    ref=${frameRef}
                                    class="floating-widget-frame"
                                    title=${title}
                                    name=${getGeneratedWidgetHostWindowName(widget)}
                                    sandbox=${sandbox}
                                    referrerpolicy="no-referrer"
                                    srcdoc=${srcDoc}
                                ></iframe>
                            `}
                </div>
            </section>
        </div>
    `;
}

// web/src/ui/zip-preview.ts
var EOCD_SIGNATURE = 101010256;
var CENTRAL_DIRECTORY_SIGNATURE = 33639248;
var ZIP64_EOCD_LOCATOR_SIGNATURE = 117853008;
var MAX_EOCD_SEARCH_BYTES = 22 + 65535;
var UTF8_FLAG = 2048;
var utf8Decoder = new TextDecoder("utf-8", { fatal: false });
function readUint16(bytes, offset) {
  return bytes[offset] | bytes[offset + 1] << 8;
}
function readUint32(bytes, offset) {
  return (bytes[offset] | bytes[offset + 1] << 8 | bytes[offset + 2] << 16 | bytes[offset + 3] << 24) >>> 0;
}
function decodeBytes(bytes, offset, length) {
  return utf8Decoder.decode(bytes.subarray(offset, offset + length));
}
function findEndOfCentralDirectory(bytes) {
  const start = Math.max(0, bytes.length - MAX_EOCD_SEARCH_BYTES);
  for (let offset = bytes.length - 22;offset >= start; offset -= 1) {
    if (readUint32(bytes, offset) === EOCD_SIGNATURE)
      return offset;
  }
  return -1;
}
function detectZip64(bytes, eocdOffset) {
  const start = Math.max(0, eocdOffset - 20);
  for (let offset = start;offset <= eocdOffset - 4; offset += 1) {
    if (readUint32(bytes, offset) === ZIP64_EOCD_LOCATOR_SIGNATURE)
      return true;
  }
  return false;
}
function inferDirectoryNames(entries) {
  const names = new Set;
  for (const entry of entries) {
    const normalized = entry.path.replace(/\/+/g, "/");
    if (!normalized)
      continue;
    if (entry.isDirectory) {
      names.add(normalized.endsWith("/") ? normalized.slice(0, -1) : normalized);
      continue;
    }
    const parts = normalized.split("/").filter(Boolean);
    if (parts.length <= 1)
      continue;
    let prefix = "";
    for (let index = 0;index < parts.length - 1; index += 1) {
      prefix = prefix ? `${prefix}/${parts[index]}` : parts[index];
      names.add(prefix);
    }
  }
  return names;
}
function normalizePath(value) {
  return String(value || "").replace(/\\/g, "/").trim();
}
function getCompressionMethodLabel(method) {
  switch (Number(method)) {
    case 0:
      return "Stored";
    case 8:
      return "Deflate";
    case 9:
      return "Deflate64";
    case 12:
      return "BZIP2";
    case 14:
      return "LZMA";
    case 93:
      return "Zstandard";
    default:
      return `Method ${method}`;
  }
}
function parseZipPreview(bytesLike) {
  const bytes = bytesLike instanceof Uint8Array ? bytesLike : bytesLike instanceof ArrayBuffer ? new Uint8Array(bytesLike) : new Uint8Array(bytesLike.buffer, bytesLike.byteOffset, bytesLike.byteLength);
  if (bytes.length < 22) {
    throw new Error("ZIP archive is too small to contain a valid directory.");
  }
  const eocdOffset = findEndOfCentralDirectory(bytes);
  if (eocdOffset < 0) {
    throw new Error("ZIP archive directory could not be found.");
  }
  if (detectZip64(bytes, eocdOffset)) {
    throw new Error("ZIP64 archives are not previewable yet.");
  }
  const totalEntries = readUint16(bytes, eocdOffset + 10);
  const centralDirectorySize = readUint32(bytes, eocdOffset + 12);
  const centralDirectoryOffset = readUint32(bytes, eocdOffset + 16);
  if (centralDirectoryOffset + centralDirectorySize > bytes.length) {
    throw new Error("ZIP archive directory is truncated.");
  }
  const entries = [];
  let cursor = centralDirectoryOffset;
  const endOffset = centralDirectoryOffset + centralDirectorySize;
  while (cursor < endOffset) {
    if (cursor + 46 > bytes.length) {
      throw new Error("ZIP archive directory entry is truncated.");
    }
    if (readUint32(bytes, cursor) !== CENTRAL_DIRECTORY_SIGNATURE) {
      throw new Error("ZIP archive directory contains an unexpected record.");
    }
    const generalPurposeFlags = readUint16(bytes, cursor + 8);
    const compressionMethod = readUint16(bytes, cursor + 10);
    const compressedSize = readUint32(bytes, cursor + 20);
    const uncompressedSize = readUint32(bytes, cursor + 24);
    const fileNameLength = readUint16(bytes, cursor + 28);
    const extraLength = readUint16(bytes, cursor + 30);
    const commentLength = readUint16(bytes, cursor + 32);
    const nameOffset = cursor + 46;
    const commentOffset = nameOffset + fileNameLength + extraLength;
    const nextOffset = commentOffset + commentLength;
    if (nextOffset > bytes.length) {
      throw new Error("ZIP archive directory entry metadata is truncated.");
    }
    const supportsUtf8 = (generalPurposeFlags & UTF8_FLAG) === UTF8_FLAG;
    const path = normalizePath(decodeBytes(bytes, nameOffset, fileNameLength));
    const comment = decodeBytes(bytes, commentOffset, commentLength);
    const isDirectory = path.endsWith("/");
    if (path) {
      entries.push({
        path,
        isDirectory,
        compressedSize,
        uncompressedSize,
        compressionMethod,
        comment: supportsUtf8 ? comment : comment
      });
    }
    cursor = nextOffset;
  }
  entries.sort((left, right) => {
    if (left.isDirectory !== right.isDirectory)
      return left.isDirectory ? -1 : 1;
    return left.path.localeCompare(right.path);
  });
  const files = entries.filter((entry) => !entry.isDirectory);
  const directories = inferDirectoryNames(entries);
  return {
    entries,
    summary: {
      fileCount: files.length,
      directoryCount: directories.size,
      totalEntries: entries.length,
      compressedBytes: files.reduce((sum, entry) => sum + entry.compressedSize, 0),
      uncompressedBytes: files.reduce((sum, entry) => sum + entry.uncompressedSize, 0)
    }
  };
}
function formatCompressionRatio(summary) {
  if (!summary)
    return null;
  if (summary.uncompressedBytes <= 0)
    return null;
  const saved = 1 - summary.compressedBytes / summary.uncompressedBytes;
  if (!Number.isFinite(saved))
    return null;
  return `${Math.round(saved * 100)}% smaller`;
}

// web/src/components/attachment-preview-modal.ts
var HTML_ATTACHMENT_PREVIEW_SANDBOX = "allow-scripts";
function isProbablyTextBytes(bytes) {
  if (!(bytes instanceof Uint8Array) || bytes.length === 0)
    return true;
  let suspicious = 0;
  const sample = bytes.subarray(0, Math.min(bytes.length, 4096));
  for (const byte of sample) {
    if (byte === 0)
      return false;
    const isControl = byte < 32 && byte !== 9 && byte !== 10 && byte !== 13 && byte !== 12;
    if (isControl)
      suspicious += 1;
  }
  return suspicious / sample.length < 0.02;
}
function shouldSniffTextAttachment(info, filename) {
  const normalizedType = String(info?.content_type || "").trim().toLowerCase();
  const normalizedName = String(filename || "").trim().toLowerCase();
  if (normalizedType.startsWith("text/") || normalizedType === "application/json" || normalizedType === "application/xml") {
    return false;
  }
  return normalizedType === "application/octet-stream" || normalizedName.endsWith(".sb") || normalizedName.endsWith(".sh");
}
function decodeTextBytes(bytes) {
  try {
    return new TextDecoder("utf-8", { fatal: false }).decode(bytes);
  } catch {
    return new TextDecoder().decode(bytes);
  }
}
function buildMetadata(info, languageLabel = null, archivePreview = null) {
  const size = info?.metadata?.size;
  const contentType = info?.content_type || "application/octet-stream";
  const archiveSummary = archivePreview?.summary || null;
  return [
    { label: "Type", value: contentType },
    { label: "Syntax", value: languageLabel },
    { label: "Entries", value: archiveSummary ? String(archiveSummary.totalEntries) : null },
    { label: "Files", value: archiveSummary ? String(archiveSummary.fileCount) : null },
    { label: "Folders", value: archiveSummary ? String(archiveSummary.directoryCount) : null },
    { label: "Compressed", value: archiveSummary ? formatFileSize(archiveSummary.compressedBytes) : null },
    { label: "Uncompressed", value: archiveSummary ? formatFileSize(archiveSummary.uncompressedBytes) : null },
    { label: "Savings", value: formatCompressionRatio(archiveSummary) },
    { label: "Size", value: typeof size === "number" ? formatFileSize(size) : null },
    { label: "Added", value: info?.created_at ? formatTimestamp(info.created_at) : null }
  ].filter((entry) => entry.value);
}
function previewLanguageFromAttachment(info, filename) {
  const normalizedType = String(info?.content_type || "").trim().toLowerCase();
  const normalizedName = String(filename || "").trim().toLowerCase();
  const basename = normalizedName.split("/").pop() || normalizedName;
  if (normalizedName.endsWith(".yaml") || normalizedName.endsWith(".yml") || normalizedType === "text/yaml" || normalizedType === "application/yaml") {
    return "yaml";
  }
  if (normalizedName.endsWith(".json") || normalizedName.endsWith(".jsonl") || normalizedType === "application/json") {
    return "json";
  }
  if (normalizedName.endsWith(".xml") || normalizedName.endsWith(".svg") || normalizedType === "application/xml" || normalizedType === "text/xml" || normalizedType === "image/svg+xml") {
    return "xml";
  }
  if (normalizedName.endsWith(".html") || normalizedName.endsWith(".htm") || normalizedType === "text/html") {
    return "html";
  }
  if (normalizedName.endsWith(".css") || normalizedType === "text/css") {
    return "css";
  }
  if (normalizedName.endsWith(".ts") || normalizedName.endsWith(".tsx") || normalizedType === "text/typescript") {
    return normalizedName.endsWith(".tsx") ? "tsx" : "ts";
  }
  if (normalizedName.endsWith(".js") || normalizedName.endsWith(".jsx") || normalizedType === "text/javascript" || normalizedType === "application/javascript") {
    return normalizedName.endsWith(".jsx") ? "jsx" : "js";
  }
  if (normalizedName.endsWith(".py") || normalizedType === "text/x-python" || normalizedType === "application/x-python-code") {
    return "python";
  }
  if (normalizedName.endsWith(".go") || normalizedType === "text/x-go") {
    return "go";
  }
  if (normalizedName.endsWith(".rb") || normalizedType === "text/x-ruby") {
    return "ruby";
  }
  if (normalizedName.endsWith(".rs") || normalizedType === "text/x-rustsrc") {
    return "rust";
  }
  if (normalizedName.endsWith(".ps1") || normalizedName.endsWith(".psm1") || normalizedName.endsWith(".psd1") || normalizedType === "text/x-powershell") {
    return "powershell";
  }
  if (basename === "dockerfile" || normalizedName.endsWith(".dockerfile")) {
    return "dockerfile";
  }
  if (normalizedName.endsWith(".md") || normalizedName.endsWith(".markdown") || normalizedType === "text/markdown") {
    return "markdown";
  }
  if (normalizedName.endsWith(".sh") || normalizedName.endsWith(".bash") || normalizedName.endsWith(".zsh") || basename === ".bashrc" || basename === ".bash_profile" || basename === ".profile" || basename === ".zshrc" || basename === ".zprofile" || basename === ".zshenv" || basename === ".env" || basename.startsWith(".env.") || normalizedType === "text/x-shellscript") {
    return "bash";
  }
  if (normalizedName.endsWith(".sql")) {
    return "sql";
  }
  if (normalizedName.endsWith(".toml") || normalizedName.endsWith(".ini") || normalizedName.endsWith(".cfg") || normalizedName.endsWith(".conf") || normalizedName.endsWith(".properties") || normalizedType === "text/toml") {
    return "toml";
  }
  return null;
}
function buildFrameUrl(mediaId, filename, previewKind) {
  const safeName = encodeURIComponent(filename || `attachment-${mediaId}`);
  const safeMediaId = encodeURIComponent(String(mediaId));
  if (previewKind === "pdf") {
    return `/pdf-viewer/?media=${safeMediaId}&name=${safeName}#media=${safeMediaId}&name=${safeName}`;
  }
  if (previewKind === "office") {
    const mediaUrl = getMediaUrl(mediaId);
    return `/office-viewer/?url=${encodeURIComponent(mediaUrl)}&name=${safeName}`;
  }
  if (previewKind === "drawio") {
    return `/drawio/edit.html?media=${safeMediaId}&name=${safeName}&readonly=1#media=${safeMediaId}&name=${safeName}&readonly=1`;
  }
  return null;
}
function AttachmentPreviewModal({ mediaId, info, onClose }) {
  const filename = info?.filename || `attachment-${mediaId}`;
  const previewKind = G0(() => getAttachmentPreviewKind(info?.content_type, filename), [info?.content_type, filename]);
  const previewLabel = getAttachmentPreviewLabel(previewKind);
  const isMarkdown = G0(() => isMarkdownAttachmentPreview(info?.content_type), [info?.content_type]);
  const [loading, setLoading] = w0(previewKind === "text" || previewKind === "html" || previewKind === "archive");
  const [textContent, setTextContent] = w0("");
  const [archivePreview, setArchivePreview] = w0(null);
  const [error, setError] = w0(null);
  const markdownContainerRef = o0(null);
  const previewLanguage = G0(() => previewLanguageFromAttachment(info, filename), [info, filename]);
  const previewLanguageLabel = G0(() => previewLanguage ? normalizeCodeLanguageLabel(previewLanguage) : null, [previewLanguage]);
  const metadata = G0(() => buildMetadata(info, !isMarkdown ? previewLanguageLabel : null, archivePreview), [info, isMarkdown, previewLanguageLabel, archivePreview]);
  const frameUrl = G0(() => buildFrameUrl(mediaId, filename, previewKind), [mediaId, filename, previewKind]);
  const renderedMarkdown = G0(() => {
    if (!isMarkdown || !textContent)
      return "";
    return renderMarkdown(textContent);
  }, [isMarkdown, textContent]);
  const highlightedText = G0(() => {
    if (isMarkdown || !textContent || !previewLanguage)
      return "";
    return highlightCodeToHtml(textContent, previewLanguage);
  }, [isMarkdown, textContent, previewLanguage]);
  r0(() => {
    const handleEsc = (e2) => {
      if (e2.key === "Escape")
        onClose();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [onClose]);
  r0(() => {
    if (!markdownContainerRef.current || !renderedMarkdown)
      return;
    renderMermaidDiagrams(markdownContainerRef.current);
    return;
  }, [renderedMarkdown]);
  r0(() => {
    let cancelled = false;
    async function loadPreview() {
      if (previewKind !== "text" && previewKind !== "html" && previewKind !== "archive") {
        setLoading(false);
        setError(null);
        setTextContent("");
        setArchivePreview(null);
        return;
      }
      setLoading(true);
      setError(null);
      setTextContent("");
      setArchivePreview(null);
      try {
        const blob = await getMediaBlob(mediaId);
        const bytes = new Uint8Array(await blob.arrayBuffer());
        if (previewKind === "text" || previewKind === "html") {
          if (previewKind === "text" && shouldSniffTextAttachment(info, filename) && !isProbablyTextBytes(bytes)) {
            throw new Error("Attachment does not appear to contain text content.");
          }
          const text = decodeTextBytes(bytes);
          if (!cancelled)
            setTextContent(text);
          return;
        }
        const parsed = parseZipPreview(bytes);
        if (!cancelled)
          setArchivePreview(parsed);
      } catch (loadError) {
        if (!cancelled) {
          const detail = loadError instanceof Error ? loadError.message : String(loadError || "Unknown error");
          setError(previewKind === "archive" ? `Failed to load ZIP preview. ${detail}` : `Failed to load text preview. ${detail}`);
        }
      } finally {
        if (!cancelled)
          setLoading(false);
      }
    }
    loadPreview();
    return () => {
      cancelled = true;
    };
  }, [mediaId, previewKind]);
  return X1`
        <${BodyPortal} className="attachment-preview-portal-root">
            <div class="image-modal attachment-preview-modal" onClick=${onClose}>
                <div class="attachment-preview-shell" onClick=${(e2) => {
    e2.stopPropagation();
  }}>
                    <div class="attachment-preview-header">
                        <div class="attachment-preview-heading">
                            <div class="attachment-preview-title">${filename}</div>
                            <div class="attachment-preview-subtitle">${previewLabel}</div>
                        </div>
                        <div class="attachment-preview-header-actions">
                            ${frameUrl && X1`
                                <a
                                    href=${frameUrl}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="attachment-preview-download"
                                    onClick=${(e2) => e2.stopPropagation()}
                                >
                                    Open in Tab
                                </a>
                            `}
                            <a
                                href=${getMediaUrl(mediaId)}
                                download=${filename}
                                class="attachment-preview-download"
                                onClick=${(e2) => e2.stopPropagation()}
                            >
                                Download
                            </a>
                            <button class="attachment-preview-close" type="button" onClick=${onClose}>Close</button>
                        </div>
                    </div>
                    <div class="attachment-preview-body">
                        ${loading && X1`<div class="attachment-preview-state">Loading preview…</div>`}
                        ${!loading && error && X1`<div class="attachment-preview-state">${error}</div>`}
                        ${!loading && !error && previewKind === "image" && X1`
                            <img class="attachment-preview-image" src=${getMediaUrl(mediaId)} alt=${filename} />
                        `}
                        ${!loading && !error && previewKind === "video" && X1`
                            <video class="attachment-preview-video" src=${getMediaUrl(mediaId)} controls autoplay style="max-width:100%;max-height:100%;" />
                        `}
                        ${!loading && !error && previewKind === "html" && X1`
                            <iframe class="attachment-preview-frame" srcdoc=${textContent || ""} sandbox=${HTML_ATTACHMENT_PREVIEW_SANDBOX} title=${filename}></iframe>
                        `}
                        ${!loading && !error && (previewKind === "pdf" || previewKind === "office" || previewKind === "drawio") && frameUrl && X1`
                            <iframe class="attachment-preview-frame" src=${frameUrl} title=${filename}></iframe>
                        `}
                        ${!loading && !error && previewKind === "drawio" && X1`
                            <div class="attachment-preview-readonly-note">Draw.io preview is read-only. Editing tools are disabled in this preview.</div>
                        `}
                        ${!loading && !error && previewKind === "archive" && archivePreview && X1`
                            <div class="attachment-preview-archive">
                                <div class="attachment-preview-archive-summary">
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Files</span>
                                        <strong class="attachment-preview-archive-card-value">${archivePreview.summary.fileCount}</strong>
                                    </div>
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Folders</span>
                                        <strong class="attachment-preview-archive-card-value">${archivePreview.summary.directoryCount}</strong>
                                    </div>
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Compressed</span>
                                        <strong class="attachment-preview-archive-card-value">${formatFileSize(archivePreview.summary.compressedBytes)}</strong>
                                    </div>
                                    <div class="attachment-preview-archive-card">
                                        <span class="attachment-preview-archive-card-label">Uncompressed</span>
                                        <strong class="attachment-preview-archive-card-value">${formatFileSize(archivePreview.summary.uncompressedBytes)}</strong>
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
                                            ${archivePreview.entries.map((entry) => X1`
                                                <tr key=${entry.path}>
                                                    <td class="attachment-preview-archive-name">${entry.path}</td>
                                                    <td>${entry.isDirectory ? "Folder" : "File"}</td>
                                                    <td>${entry.isDirectory ? "—" : getCompressionMethodLabel(entry.compressionMethod)}</td>
                                                    <td>${entry.isDirectory ? "—" : formatFileSize(entry.compressedSize)}</td>
                                                    <td>${entry.isDirectory ? "—" : formatFileSize(entry.uncompressedSize)}</td>
                                                </tr>
                                            `)}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        `}
                        ${!loading && !error && previewKind === "text" && isMarkdown && X1`
                            <div
                                ref=${markdownContainerRef}
                                class="attachment-preview-markdown post-content"
                                dangerouslySetInnerHTML=${{ __html: renderedMarkdown }}
                            />
                        `}
                        ${!loading && !error && previewKind === "text" && !isMarkdown && highlightedText && X1`
                            <pre class="attachment-preview-text attachment-preview-code"><code dangerouslySetInnerHTML=${{ __html: highlightedText }} /></pre>
                        `}
                        ${!loading && !error && previewKind === "text" && !isMarkdown && !highlightedText && X1`
                            <pre class="attachment-preview-text">${textContent}</pre>
                        `}
                        ${!loading && !error && previewKind === "unsupported" && X1`
                            <div class="attachment-preview-state">
                                Preview is not available for this file type yet. You can still download it directly.
                            </div>
                        `}
                    </div>
                    <div class="attachment-preview-meta">
                        ${metadata.map((entry) => X1`
                            <div class="attachment-preview-meta-item" key=${entry.label}>
                                <span class="attachment-preview-meta-label">${entry.label}</span>
                                <span class="attachment-preview-meta-value">${entry.value}</span>
                            </div>
                        `)}
                    </div>
                </div>
            </div>
        </${BodyPortal}>
    `;
}

// web/src/ui/meters.ts
var METERS_STORAGE_KEY = "piclaw_system_meters_enabled";
var METERS_COLLAPSED_STORAGE_KEY = "piclaw_system_meters_collapsed";
var METERS_EVENT_NAME = "piclaw-meters-change";
var METERS_COLLAPSED_EVENT_NAME = "piclaw-meters-collapsed-change";
function dispatchMetersCollapsedChange(collapsed) {
  if (typeof window === "undefined")
    return;
  window.dispatchEvent(new CustomEvent(METERS_COLLAPSED_EVENT_NAME, {
    detail: { collapsed: Boolean(collapsed) }
  }));
}
function readStoredMetersEnabled(defaultValue = false) {
  return getLocalStorageBoolean(METERS_STORAGE_KEY, defaultValue);
}
function readStoredMetersCollapsed(defaultValue = false) {
  return getLocalStorageBoolean(METERS_COLLAPSED_STORAGE_KEY, defaultValue);
}
function applyMetersCollapsed(collapsed, options = {}) {
  const persist = options.persist !== false;
  const next = Boolean(collapsed);
  if (persist) {
    setLocalStorageItem(METERS_COLLAPSED_STORAGE_KEY, next ? "true" : "false");
  }
  dispatchMetersCollapsedChange(next);
  return next;
}
function toggleMetersCollapsed() {
  const next = !readStoredMetersCollapsed(false);
  return applyMetersCollapsed(next);
}

// web/src/components/system-meters-hud.ts
function sanitizeSeries(input, maxPoints = 30) {
  const series = Array.isArray(input) ? input.map((value) => Number(value)).filter((value) => Number.isFinite(value)) : [];
  return series.length > maxPoints ? series.slice(series.length - maxPoints) : series;
}
function clampPercentSeries(input, maxPoints = 30) {
  return sanitizeSeries(input, maxPoints).map((value) => Math.max(0, Math.min(100, value)));
}
function buildSparklinePath(series, width = 56, height = 16, options = {}) {
  const points = sanitizeSeries(series);
  if (points.length === 0)
    return "";
  const minValue = Number.isFinite(options.min) ? Number(options.min) : Math.min(...points);
  const maxValue = Number.isFinite(options.max) ? Number(options.max) : Math.max(...points);
  if (!(maxValue > minValue)) {
    const y2 = (height / 2).toFixed(2);
    return `M 0 ${y2} L ${width} ${y2}`;
  }
  if (points.length === 1) {
    const normalized = (points[0] - minValue) / (maxValue - minValue);
    const y2 = (height - normalized * height).toFixed(2);
    return `M 0 ${y2} L ${width} ${y2}`;
  }
  return points.map((value, index) => {
    const x3 = index / (points.length - 1 || 1) * width;
    const normalized = (value - minValue) / (maxValue - minValue);
    const y2 = height - normalized * height;
    return `${index === 0 ? "M" : "L"} ${x3.toFixed(2)} ${y2.toFixed(2)}`;
  }).join(" ");
}
function formatPercent(value) {
  return `${Math.round(Number(value) || 0)}%`;
}
function formatBytesCompact(value) {
  const bytes = Number(value);
  if (!Number.isFinite(bytes) || bytes <= 0)
    return "0B";
  const units = ["B", "K", "M", "G", "T"];
  let unitIndex = 0;
  let scaled = bytes;
  while (scaled >= 1024 && unitIndex < units.length - 1) {
    scaled /= 1024;
    unitIndex += 1;
  }
  const digits = scaled >= 100 || unitIndex === 0 ? 0 : scaled >= 10 ? 0 : 1;
  return `${scaled.toFixed(digits)}${units[unitIndex]}`;
}
function buildCompactMetersSummary(metrics) {
  const parts = [
    `CPU ${formatPercent(metrics?.cpu_percent)}`,
    `RAM ${formatPercent(metrics?.ram_percent)}`
  ];
  if (Number.isFinite(Number(metrics?.swap_percent)) && Number(metrics?.swap_total_bytes) > 0) {
    parts.push(`SWP ${formatPercent(metrics?.swap_percent)}`);
  }
  return parts.join(" • ");
}
function resolveCurrentRssBytes(metrics) {
  return Number(metrics?.process_memory?.vm_rss_bytes) > 0 ? Number(metrics.process_memory.vm_rss_bytes) : Number(metrics?.process_memory?.rss_bytes) || 0;
}
function shouldShowRss(metrics) {
  return resolveCurrentRssBytes(metrics) > 0 && sanitizeSeries(metrics?.process_rss_series_bytes).length > 0;
}
function readIsNarrowLayout() {
  if (typeof window === "undefined" || typeof window.matchMedia !== "function")
    return false;
  return window.matchMedia("(max-width: 900px)").matches;
}
function SystemMetersHud({ mode = "overlay" }) {
  const [enabled, setEnabled] = w0(() => readStoredMetersEnabled(false));
  const [collapsed, setCollapsed] = w0(() => readStoredMetersCollapsed(false));
  const [isNarrowLayout, setIsNarrowLayout] = w0(() => readIsNarrowLayout());
  const [metrics, setMetrics] = w0({
    cpu_percent: 0,
    ram_percent: 0,
    swap_percent: null,
    cpu_series: [],
    ram_series: [],
    swap_series: [],
    process_rss_series_bytes: [],
    process_memory: {
      rss_bytes: 0,
      vm_rss_bytes: null
    },
    swap_total_bytes: 0,
    swap_used_bytes: 0,
    sample_interval_ms: 2000,
    platform: ""
  });
  const [loading, setLoading] = w0(false);
  r0(() => {
    const onMetersChange = (event) => {
      setEnabled(Boolean(event?.detail?.enabled));
    };
    const onMetersCollapsedChange = (event) => {
      setCollapsed(Boolean(event?.detail?.collapsed));
    };
    window.addEventListener(METERS_EVENT_NAME, onMetersChange);
    window.addEventListener(METERS_COLLAPSED_EVENT_NAME, onMetersCollapsedChange);
    return () => {
      window.removeEventListener(METERS_EVENT_NAME, onMetersChange);
      window.removeEventListener(METERS_COLLAPSED_EVENT_NAME, onMetersCollapsedChange);
    };
  }, []);
  r0(() => {
    if (typeof window === "undefined" || typeof window.matchMedia !== "function")
      return;
    const mediaQuery = window.matchMedia("(max-width: 900px)");
    const sync = () => setIsNarrowLayout(Boolean(mediaQuery.matches));
    sync();
    if (typeof mediaQuery.addEventListener === "function") {
      mediaQuery.addEventListener("change", sync);
      return () => mediaQuery.removeEventListener("change", sync);
    }
    mediaQuery.addListener(sync);
    return () => mediaQuery.removeListener(sync);
  }, []);
  const activeMode = "overlay";
  const isActiveInstance = mode === activeMode;
  r0(() => {
    if (!enabled || !isActiveInstance)
      return;
    let cancelled = false;
    let timer = 0;
    const refresh = async () => {
      setLoading((prev) => prev || metrics.cpu_series.length > 0 ? prev : true);
      try {
        const next = await getSystemMetrics();
        if (cancelled)
          return;
        setMetrics({
          cpu_percent: Number(next?.cpu_percent) || 0,
          ram_percent: Number(next?.ram_percent) || 0,
          swap_percent: Number.isFinite(Number(next?.swap_percent)) ? Number(next?.swap_percent) : null,
          cpu_series: clampPercentSeries(next?.cpu_series),
          ram_series: clampPercentSeries(next?.ram_series),
          swap_series: clampPercentSeries(next?.swap_series),
          process_rss_series_bytes: sanitizeSeries(next?.process_rss_series_bytes),
          process_memory: {
            rss_bytes: Number(next?.process_memory?.rss_bytes) || 0,
            vm_rss_bytes: Number.isFinite(Number(next?.process_memory?.vm_rss_bytes)) ? Number(next?.process_memory?.vm_rss_bytes) : null
          },
          swap_total_bytes: Number(next?.swap_total_bytes) || 0,
          swap_used_bytes: Number(next?.swap_used_bytes) || 0,
          sample_interval_ms: Number(next?.sample_interval_ms) || 2000,
          platform: String(next?.platform || "")
        });
      } catch {
        if (cancelled)
          return;
      } finally {
        if (!cancelled)
          setLoading(false);
      }
    };
    refresh();
    timer = window.setInterval(() => {
      if (document?.visibilityState === "hidden")
        return;
      refresh();
    }, Math.max(1000, Number(metrics.sample_interval_ms) || 2000));
    return () => {
      cancelled = true;
      if (timer)
        window.clearInterval(timer);
    };
  }, [enabled, isActiveInstance]);
  const cpuPath = G0(() => buildSparklinePath(metrics.cpu_series, 56, 16, { min: 0, max: 100 }), [metrics.cpu_series]);
  const ramPath = G0(() => buildSparklinePath(metrics.ram_series, 56, 16, { min: 0, max: 100 }), [metrics.ram_series]);
  const swapPath = G0(() => buildSparklinePath(metrics.swap_series, 56, 16, { min: 0, max: 100 }), [metrics.swap_series]);
  const rssPath = G0(() => buildSparklinePath(metrics.process_rss_series_bytes), [metrics.process_rss_series_bytes]);
  const showSwap = Number.isFinite(Number(metrics.swap_percent)) && metrics.swap_total_bytes > 0;
  const currentRssBytes = resolveCurrentRssBytes(metrics);
  const showRss = shouldShowRss(metrics);
  const compactSummary = G0(() => buildCompactMetersSummary(metrics), [metrics]);
  if (!enabled || !isActiveInstance)
    return null;
  const title = collapsed ? "Show system meters" : loading ? "Updating system meters… Click to collapse." : "System meters — click to collapse.";
  const handleToggleCollapsed = (event) => {
    event?.stopPropagation?.();
    toggleMetersCollapsed();
  };
  return X1`
        <div class=${`system-meters-hud system-meters-hud-${mode}${collapsed ? " is-collapsed" : ""}`} aria-live="polite">
            <button
                class="system-meters-card"
                type="button"
                title=${title}
                aria-label=${title}
                aria-expanded=${collapsed ? "false" : "true"}
                onClick=${handleToggleCollapsed}
            >
                ${collapsed ? X1`<span class="system-meters-collapse-tab" aria-hidden="true">◂</span>` : isNarrowLayout ? X1`<span class="system-meters-compact-summary">${compactSummary}</span>` : X1`
                            <div class="system-meters-row cpu">
                                <span class="system-meters-label">CPU</span>
                                <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                    <path d=${cpuPath}></path>
                                </svg>
                                <span class="system-meters-value">${formatPercent(metrics.cpu_percent)}</span>
                            </div>
                            <div class="system-meters-row ram">
                                <span class="system-meters-label">RAM</span>
                                <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                    <path d=${ramPath}></path>
                                </svg>
                                <span class="system-meters-value">${formatPercent(metrics.ram_percent)}</span>
                            </div>
                            ${showRss && X1`
                                <div class="system-meters-row rss">
                                    <span class="system-meters-label">RSS</span>
                                    <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                        <path d=${rssPath}></path>
                                    </svg>
                                    <span class="system-meters-value">${formatBytesCompact(currentRssBytes)}</span>
                                </div>
                            `}
                            ${showSwap && X1`
                                <div class="system-meters-row swap">
                                    <span class="system-meters-label">SWP</span>
                                    <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                        <path d=${swapPath}></path>
                                    </svg>
                                    <span class="system-meters-value">${formatPercent(metrics.swap_percent)}</span>
                                </div>
                            `}
                        `}
            </button>
        </div>
    `;
}

// web/src/app.ts
var DEFAULT_SESSION_TITLE = "default";
var SESSION_KEY = "gi_session_id";
var DEFAULT_AGENT_ID = "gi";
function sessionToChatJid(id) {
  return `gi:${id}`;
}
async function ensureDefaultSession() {
  const stored = getLocalStorageItem(SESSION_KEY);
  if (stored) {
    try {
      const r3 = await fetch(`/api/sessions/${encodeURIComponent(stored)}`);
      if (r3.ok)
        return stored;
    } catch {}
  }
  const r2 = await fetch("/api/sessions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title: DEFAULT_SESSION_TITLE })
  });
  if (!r2.ok)
    throw new Error("Failed to create default session");
  const s2 = await r2.json();
  setLocalStorageItem(SESSION_KEY, s2.id);
  return s2.id;
}
async function getRuntimeConfig() {
  const r2 = await fetch("/api/runtime/config");
  if (!r2.ok)
    return {};
  return r2.json();
}
function GiApp() {
  const [ready, setReady] = w0(false);
  const [sessionId, setSessionId] = w0(null);
  const [runtimeConfig, setRuntimeConfig] = w0({});
  const [agents, setAgents] = w0({});
  const [userProfile, setUserProfile] = w0(null);
  const [workspaceOpen, setWorkspaceOpen] = w0(false);
  const [tabs, setTabs] = w0([]);
  const [activeTabId, setActiveTabId] = w0(null);
  const editorOpen = tabs.length > 0;
  const [posts, setPosts] = w0([]);
  const [hasMore, setHasMore] = w0(false);
  const timelineRef = o0(null);
  const [fileRefs, setFileRefs] = w0([]);
  const [messageRefs, setMessageRefs] = w0([]);
  const [followupQueueItems, setFollowupQueueItems] = w0([]);
  const [floatingWidget, setFloatingWidget] = w0(null);
  const [attachmentPreview, setAttachmentPreview] = w0(null);
  const [contextUsage, setContextUsage] = w0(null);
  const [activeModel, setActiveModel] = w0("");
  const [agentModelsPayload, setAgentModelsPayload] = w0(null);
  const [activeThinkingLevel, setActiveThinkingLevel] = w0("");
  const [connectionStatus, setConnectionStatus] = w0("connected");
  const [isAgentTurnActive, setIsAgentTurnActive] = w0(false);
  const isAgentRunningRef = o0(false);
  const {
    agentStatus,
    setAgentStatus,
    agentDraft,
    setAgentDraft,
    agentPlan,
    setAgentPlan,
    agentThought,
    setAgentThought,
    pendingRequest,
    setPendingRequest,
    currentTurnId,
    setCurrentTurnId,
    steerQueuedTurnId,
    setSteerQueuedTurnId,
    lastAgentEventRef,
    draftBufferRef,
    thoughtBufferRef,
    pendingRequestRef,
    currentTurnIdRef,
    steerQueuedTurnIdRef,
    thoughtExpandedRef,
    draftExpandedRef
  } = useAgentState();
  const currentChatJid = G0(() => sessionId ? sessionToChatJid(sessionId) : "", [sessionId]);
  r0(() => {
    const cleanupTheme = initTheme();
    if (getLocalStorageItem("piclaw_system_meters_enabled") === null) {
      setLocalStorageItem("piclaw_system_meters_enabled", "true");
    }
    Promise.all([
      ensureDefaultSession(),
      getRuntimeConfig()
    ]).then(([sid, cfg]) => {
      setSessionId(sid);
      setRuntimeConfig(cfg);
      setUserProfile({ name: cfg.user_name, avatarUrl: cfg.user_avatar, avatarBackground: cfg.user_avatar_background });
      setAgents({
        [DEFAULT_AGENT_ID]: {
          id: DEFAULT_AGENT_ID,
          name: cfg.assistant_name || "Gi",
          avatar_url: cfg.assistant_avatar || null
        }
      });
      setActiveModel(cfg.default_model || "");
      setActiveThinkingLevel(cfg.default_thinking_level || "");
      setReady(true);
    }).catch((err2) => {
      console.error("[gi] Bootstrap failed:", err2);
    });
    return cleanupTheme;
  }, []);
  const loadPosts = t0(async (opts = {}) => {
    if (!sessionId)
      return;
    const chatJid = sessionToChatJid(sessionId);
    const data = await getTimeline(50, opts.beforeId || null, chatJid);
    const incoming = data.posts || [];
    if (opts.beforeId) {
      setPosts((prev) => dedupePosts([...incoming, ...prev]));
    } else {
      setPosts(dedupePosts(incoming));
    }
    setHasMore(incoming.length >= 50);
  }, [sessionId]);
  const scrollToBottom = t0(() => {
    const el = timelineRef.current;
    if (!el)
      return;
    el.scrollTop = el.scrollHeight;
  }, []);
  const handleSseEvent = t0((eventType, data) => {
    if (eventType === "new_post" || eventType === "agent_response") {
      if (data && data.id) {
        setPosts((prev) => appendUniqueTimelinePost(prev, data));
        scrollToBottom();
      }
    }
    if (eventType === "agent_status") {
      setAgentStatus(data);
      const active = data?.status === "running" || data?.status === "cancelling";
      setIsAgentTurnActive(active);
      isAgentRunningRef.current = active;
    }
    if (eventType === "agent_draft_delta") {
      const delta = data?.delta || "";
      if (delta) {
        draftBufferRef.current = (draftBufferRef.current || "") + delta;
        setAgentDraft({ text: draftBufferRef.current, totalLines: 0, fullText: draftBufferRef.current });
      }
    }
    if (eventType === "agent_thought_delta") {
      const delta = data?.delta || "";
      if (delta) {
        thoughtBufferRef.current = (thoughtBufferRef.current || "") + delta;
        setAgentThought({ text: thoughtBufferRef.current, totalLines: 0, fullText: thoughtBufferRef.current });
      }
    }
    if (eventType === "agent_response") {
      draftBufferRef.current = "";
      thoughtBufferRef.current = "";
      setAgentDraft(null);
      setAgentThought(null);
    }
  }, [scrollToBottom]);
  const handleConnectionStatusChange = t0((status) => {
    setConnectionStatus(status);
  }, []);
  useSseConnection({
    handleSseEvent,
    handleConnectionStatusChange,
    loadPosts,
    onWake: () => {
      loadPosts();
    },
    chatJid: currentChatJid
  });
  r0(() => {
    if (!ready || !sessionId)
      return;
    loadPosts();
    const id = setInterval(() => loadPosts(), 1e4);
    return () => clearInterval(id);
  }, [ready, sessionId]);
  const handlePost = t0(async (response) => {
    await loadPosts();
    scrollToBottom();
  }, [loadPosts, scrollToBottom]);
  const openEditor = t0((path) => {
    const existing = tabs.find((t2) => t2.id === path || t2.path === path);
    if (existing) {
      setActiveTabId(existing.id);
      return;
    }
    setTabs((prev) => [...prev, { id: path, path, label: path.split("/").pop() || path, dirty: false, pinned: false }]);
    setActiveTabId(path);
  }, [tabs]);
  const handleTabClose = t0((id) => {
    setTabs((prev) => {
      const next = prev.filter((t2) => t2.id !== id);
      if (activeTabId === id)
        setActiveTabId(next[next.length - 1]?.id || null);
      return next;
    });
  }, [activeTabId]);
  const appShellClass = [
    "app-shell",
    workspaceOpen ? "" : "workspace-collapsed",
    editorOpen ? "editor-open" : ""
  ].filter(Boolean).join(" ");
  if (!ready) {
    return X1`<div id="app"><div style="padding:20px;text-align:center;color:var(--text-secondary,#888)">Loading…</div></div>`;
  }
  return X1`
        <div class=${appShellClass}>
            <${SystemMetersHud} mode="overlay" />
            <${WorkspaceExplorer}
                onFileSelect=${(path) => setFileRefs((p2) => [...p2, path])}
                visible=${workspaceOpen}
                active=${workspaceOpen || editorOpen}
                onOpenEditor=${openEditor}
                onOpenTerminalTab=${() => {}}
                onOpenVncTab=${() => {}}
            />
            <button
                class=${`workspace-toggle-tab${workspaceOpen ? " open" : " closed"}`}
                onClick=${() => setWorkspaceOpen((v2) => !v2)}
                title=${workspaceOpen ? "Hide workspace" : "Show workspace"}
            >
                <svg class="workspace-toggle-tab-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="6 3 11 8 6 13" />
                </svg>
            </button>
            <div class="workspace-splitter"></div>
            ${editorOpen && X1`
                <div class="editor-pane-container">
                    <${TabStrip}
                        tabs=${tabs}
                        activeId=${activeTabId}
                        onActivate=${(id) => setActiveTabId(id)}
                        onClose=${handleTabClose}
                        onCloseOthers=${(id) => setTabs((p2) => p2.filter((t2) => t2.id === id))}
                        onCloseAll=${() => {
    setTabs([]);
    setActiveTabId(null);
  }}
                        onTogglePin=${() => {}}
                    />
                    <div class="editor-pane-host"></div>
                </div>
                <div class="editor-splitter"></div>
            `}
            <div class="container">
                <${Timeline}
                    posts=${posts}
                    hasMore=${hasMore}
                    onLoadMore=${({ preserveScroll }) => {
    const oldest = posts[0];
    if (oldest)
      loadPosts({ beforeId: oldest.id });
  }}
                    timelineRef=${timelineRef}
                    onHashtagClick=${() => {}}
                    onMessageRef=${() => {}}
                    onScrollToMessage=${() => {}}
                    onFileRef=${openEditor}
                    onPostClick=${undefined}
                    onDeletePost=${() => {}}
                    onOpenWidget=${(w) => setFloatingWidget(w)}
                    onOpenAttachmentPreview=${setAttachmentPreview}
                    emptyMessage="Send a message to get started."
                    agents=${agents}
                    user=${userProfile}
                    reverse=${true}
                    removingPostIds=${new Set}
                    searchQuery=""
                />
                <${AgentStatus}
                    status=${isCompactionStatus(agentStatus) ? null : agentStatus}
                    draft=${agentDraft}
                    plan=${agentPlan}
                    thought=${agentThought}
                    pendingRequest=${pendingRequest}
                    intent=${null}
                    turnId=${currentTurnId}
                    steerQueued=${Boolean(steerQueuedTurnId)}
                    onPanelToggle=${() => {}}
                    showExtensionPanels=${false}
                />
                <${FloatingWidgetPane}
                    widget=${floatingWidget}
                    onClose=${() => setFloatingWidget(null)}
                    onWidgetEvent=${() => {}}
                />
                ${attachmentPreview && X1`
                    <${AttachmentPreviewModal}
                        mediaId=${attachmentPreview.mediaId}
                        info=${attachmentPreview.info}
                        onClose=${() => setAttachmentPreview(null)}
                    />
                `}
                <${QueuedFollowupStack}
                    items=${followupQueueItems}
                    onInjectQueuedFollowup=${() => {}}
                    onRemoveQueuedFollowup=${() => {}}
                    onMoveQueuedFollowup=${() => {}}
                    onOpenFilePill=${openEditor}
                />
                <${ComposeBox}
                    currentChatJid=${currentChatJid}
                    isAgentActive=${isAgentTurnActive}
                    onPost=${handlePost}
                    onFocus=${() => {
    if (!isIOSDevice())
      scrollToBottom();
  }}
                    agents=${agents}
                    currentSessionAgent=${agents[DEFAULT_AGENT_ID] ? {
    ...agents[DEFAULT_AGENT_ID],
    chat_jid: currentChatJid,
    agent_name: agents[DEFAULT_AGENT_ID].name
  } : null}
                    agentStatus=${agentStatus}
                    agentDraft=${agentDraft}
                    contextUsage=${contextUsage}
                    fileRefs=${fileRefs}
                    messageRefs=${messageRefs}
                    onRemoveFileRef=${(p2) => setFileRefs((prev) => prev.filter((x3) => x3 !== p2))}
                    onClearFileRefs=${() => setFileRefs([])}
                    onRemoveMessageRef=${() => {}}
                    onClearMessageRefs=${() => setMessageRefs([])}
                    connectionStatus=${connectionStatus}
                    activeChatAgents=${[]}
                    currentChatBranches=${[]}
                    formatBranchPickerLabel=${(b) => b?.label || b?.chat_jid || ""}
                    handleBranchPickerChange=${() => {}}
                    searchOpen=${false}
                    onEnterSearch=${() => {}}
                    onExitSearch=${() => {}}
                    onSearch=${() => {}}
                    searchScope="current"
                    onSearchScopeChange=${() => {}}
                    activeModel=${activeModel}
                    agentModelsPayload=${agentModelsPayload}
                    activeModelUsage=${null}
                    activeThinkingLevel=${activeThinkingLevel}
                    supportsThinking=${false}
                    followupQueueCount=${followupQueueItems.length}
                    notificationsEnabled=${false}
                    notificationPermission="default"
                    onToggleNotifications=${() => {}}
                    onComposeSubmitError=${() => {}}
                    pendingRequestRef=${pendingRequestRef}
                    setPendingRequest=${setPendingRequest}
                />
            </div>
        </div>
    `;
}
c0(X1`<${GiApp} />`, document.getElementById("app"));

//# debugId=A920E3DA3E3419FD64756E2164756E21
//# sourceMappingURL=app.js.map
