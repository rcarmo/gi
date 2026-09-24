// web/src/gi-turn-event.ts
function staleTerminalEvent(type, data, currentTurn) {
  const terminal = type === "agent_response" || type === "agent_status" && !["running", "cancelling"].includes(data?.status);
  return Boolean(terminal && currentTurn && data?.turn_id !== currentTurn);
}

// web/src/vendor/preact-htm.js
var q;
var d;
var p_;
var V_;
var H;
var c_;
var h_;
var d_;
var K;
var R;
var F;
var m_;
var Z;
var Q;
var X;
var v_;
var V = {};
var B = [];
var B_ = /acit|ex(?:s|g|n|p|$)|rph|grid|ows|mnc|ntw|ine[ch]|zoo|^ord|itera/i;
var O = Array.isArray;
function C(e, _) {
  for (var t in _)
    e[t] = _[t];
  return e;
}
function __(e) {
  e && e.parentNode && e.parentNode.removeChild(e);
}
function e_(e, _, t) {
  var o, i, n, u = {};
  for (n in _)
    n == "key" ? o = _[n] : n == "ref" ? i = _[n] : u[n] = _[n];
  if (arguments.length > 2 && (u.children = arguments.length > 3 ? q.call(arguments, 2) : t), typeof e == "function" && e.defaultProps != null)
    for (n in e.defaultProps)
      u[n] === undefined && (u[n] = e.defaultProps[n]);
  return I(e, u, o, i, null);
}
function I(e, _, t, o, i) {
  var n = { type: e, props: _, key: t, ref: o, __k: null, __: null, __b: 0, __e: null, __c: null, constructor: undefined, __v: i == null ? ++p_ : i, __i: -1, __u: 0 };
  return i == null && d.vnode != null && d.vnode(n), n;
}
function z(e) {
  return e.children;
}
function A(e, _) {
  this.props = e, this.context = _;
}
function U(e, _) {
  if (_ == null)
    return e.__ ? U(e.__, e.__i + 1) : null;
  for (var t;_ < e.__k.length; _++)
    if ((t = e.__k[_]) != null && t.__e != null)
      return t.__e;
  return typeof e.type == "function" ? U(e) : null;
}
function $_(e) {
  if (e.__P && e.__d) {
    var _ = e.__v, t = _.__e, o = [], i = [], n = C({}, _);
    n.__v = _.__v + 1, d.vnode && d.vnode(n), t_(e.__P, n, _, e.__n, e.__P.namespaceURI, 32 & _.__u ? [t] : null, o, t == null ? U(_) : t, !!(32 & _.__u), i), n.__v = _.__v, n.__.__k[n.__i] = n, x_(o, n, i), _.__e = _.__ = null, n.__e != t && y_(n);
  }
}
function y_(e) {
  if ((e = e.__) != null && e.__c != null)
    return e.__e = e.__c.base = null, e.__k.some(function(_) {
      if (_ != null && _.__e != null)
        return e.__e = e.__c.base = _.__e;
    }), y_(e);
}
function Y(e) {
  (!e.__d && (e.__d = true) && H.push(e) && !j.__r++ || c_ != d.debounceRendering) && ((c_ = d.debounceRendering) || h_)(j);
}
function j() {
  try {
    for (var e, _ = 1;H.length; )
      H.length > _ && H.sort(d_), e = H.shift(), _ = H.length, $_(e);
  } finally {
    H.length = j.__r = 0;
  }
}
function g_(e, _, t, o, i, n, u, s, c, l, f) {
  var h, r, a, y, k, b, g = o && o.__k || B, p = _.length;
  for (c = j_(t, _, g, c, p), h = 0;h < p; h++)
    (a = t.__k[h]) != null && (r = a.__i != -1 && g[a.__i] || V, a.__i = h, b = t_(e, a, r, i, n, u, s, c, l, f), y = a.__e, a.ref && r.ref != a.ref && (r.ref && n_(r.ref, null, a), f.push(a.ref, a.__c || y, a)), k == null && y != null && (k = y), 4 & a.__u ? (c = b_(a, c, e), r.__e && (r.__e = null)) : typeof a.type == "function" && b !== undefined ? c = b : y && (c = y.nextSibling), a.__u &= -7);
  return t.__e = k, c;
}
function j_(e, _, t, o, i) {
  var n, u, s, c, l, f = t.length, h = f, r = 0;
  for (e.__k = Array(i), n = 0;n < i; n++)
    (u = _[n]) != null && typeof u != "boolean" && typeof u != "function" ? (typeof u == "string" || typeof u == "number" || typeof u == "bigint" || u.constructor == String ? u = e.__k[n] = I(null, u, null, null, null) : O(u) ? u = e.__k[n] = I(z, { children: u }, null, null, null) : u.constructor === undefined && u.__b > 0 ? u = e.__k[n] = I(u.type, u.props, u.key, u.ref ? u.ref : null, u.__v) : e.__k[n] = u, c = n + r, u.__ = e, u.__b = e.__b + 1, s = null, (l = u.__i = q_(u, t, c, h)) != -1 && (h--, (s = t[l]) && (s.__u |= 2)), s == null || s.__v == null ? (l == -1 && (i > f ? r-- : i < f && r++), typeof u.type != "function" && (u.__u |= 4)) : l != c && (l == c - 1 ? r-- : l == c + 1 ? r++ : (l > c ? r-- : r++, u.__u |= 4))) : e.__k[n] = null;
  if (h)
    for (n = 0;n < f; n++)
      (s = t[n]) != null && (2 & s.__u) == 0 && (s.__e == o && (o = U(s)), C_(s, s));
  return o;
}
function b_(e, _, t) {
  var o, i;
  if (typeof e.type == "function") {
    for (o = e.__k, i = 0;o && i < o.length; i++)
      o[i] && (o[i].__ = e, _ = b_(o[i], _, t));
    return _;
  }
  e.__e != _ && (_ && e.type && !_.parentNode && (_ = U(e)), _ = t.insertBefore(e.__e, _ || null));
  do
    _ = _ && _.nextSibling;
  while (_ != null && _.nodeType == 8);
  return _;
}
function q_(e, _, t, o) {
  var i, n, u, { key: s, type: c } = e, l = _[t], f = l != null && (2 & l.__u) == 0;
  if (l === null && s == null || f && s == l.key && c == l.type)
    return t;
  if (o > (f ? 1 : 0)) {
    for (i = t - 1, n = t + 1;i >= 0 || n < _.length; )
      if ((l = _[u = i >= 0 ? i-- : n++]) != null && (2 & l.__u) == 0 && s == l.key && c == l.type)
        return u;
  }
  return -1;
}
function f_(e, _, t) {
  _[0] == "-" ? e.setProperty(_, t == null ? "" : t) : e[_] = t == null ? "" : typeof t != "number" || B_.test(_) ? t : t + "px";
}
function L(e, _, t, o, i) {
  var n, u;
  _:
    if (_ == "style")
      if (typeof t == "string")
        e.style.cssText = t;
      else {
        if (typeof o == "string" && (e.style.cssText = o = ""), o)
          for (_ in o)
            t && _ in t || f_(e.style, _, "");
        if (t)
          for (_ in t)
            o && t[_] == o[_] || f_(e.style, _, t[_]);
      }
    else if (_[0] == "o" && _[1] == "n")
      n = _ != (_ = _.replace(m_, "$1")), u = _.toLowerCase(), _ = u in e || _ == "onFocusOut" || _ == "onFocusIn" ? u.slice(2) : _.slice(2), e.l || (e.l = {}), e.l[_ + n] = t, t ? o ? t[F] = o[F] : (t[F] = Z, e.addEventListener(_, n ? X : Q, n)) : e.removeEventListener(_, n ? X : Q, n);
    else {
      if (i == "http://www.w3.org/2000/svg")
        _ = _.replace(/xlink(H|:h)/, "h").replace(/sName$/, "s");
      else if (_ != "width" && _ != "height" && _ != "href" && _ != "list" && _ != "form" && _ != "tabIndex" && _ != "download" && _ != "rowSpan" && _ != "colSpan" && _ != "role" && _ != "popover" && _ in e)
        try {
          e[_] = t == null ? "" : t;
          break _;
        } catch (s) {}
      typeof t == "function" || (t == null || t === false && _[4] != "-" ? e.removeAttribute(_) : e.setAttribute(_, _ == "popover" && t == 1 ? "" : t));
    }
}
function a_(e) {
  return function(_) {
    if (this.l) {
      var t = this.l[_.type + e];
      if (_[R] == null)
        _[R] = Z++;
      else if (_[R] < t[F])
        return;
      return t(d.event ? d.event(_) : _);
    }
  };
}
function t_(e, _, t, o, i, n, u, s, c, l) {
  var f, h, r, a, y, k, b, g, p, x, T, S, M, s_, W, J, w = _.type;
  if (_.constructor !== undefined)
    return null;
  128 & t.__u && (c = !!(32 & t.__u), n = [s = _.__e = t.__e]), (f = d.__b) && f(_);
  _:
    if (typeof w == "function") {
      h = u.length;
      try {
        if (p = _.props, x = w.prototype && w.prototype.render, T = (f = w.contextType) && o[f.__c], S = f ? T ? T.props.value : f.__ : o, t.__c ? g = (r = _.__c = t.__c).__ = r.__E : (x ? _.__c = r = new w(p, S) : (_.__c = r = new A(p, S), r.constructor = w, r.render = z_), T && T.sub(r), r.state || (r.state = {}), r.__n = o, a = r.__d = true, r.__h = [], r._sb = []), x && r.__s == null && (r.__s = r.state), x && w.getDerivedStateFromProps != null && (r.__s == r.state && (r.__s = C({}, r.__s)), C(r.__s, w.getDerivedStateFromProps(p, r.__s))), y = r.props, k = r.state, r.__v = _, a)
          x && w.getDerivedStateFromProps == null && r.componentWillMount != null && r.componentWillMount(), x && r.componentDidMount != null && r.__h.push(r.componentDidMount);
        else {
          if (x && w.getDerivedStateFromProps == null && p !== y && r.componentWillReceiveProps != null && r.componentWillReceiveProps(p, S), _.__v == t.__v || !r.__e && r.shouldComponentUpdate != null && r.shouldComponentUpdate(p, r.__s, S) === false) {
            _.__v != t.__v && (r.props = p, r.state = r.__s, r.__d = false), _.__e = t.__e, _.__k = t.__k, _.__k.some(function(D) {
              D && (D.__ = _);
            }), B.push.apply(r.__h, r._sb), r._sb = [], r.__h.length && u.push(r), s = U(t);
            break _;
          }
          r.componentWillUpdate != null && r.componentWillUpdate(p, r.__s, S), x && r.componentDidUpdate != null && r.__h.push(function() {
            r.componentDidUpdate(y, k, b);
          });
        }
        if (r.context = S, r.props = p, r.__P = e, r.__e = false, M = d.__r, s_ = 0, x)
          r.state = r.__s, r.__d = false, M && M(_), f = r.render(r.props, r.state, r.context), B.push.apply(r.__h, r._sb), r._sb = [];
        else
          do
            r.__d = false, M && M(_), f = r.render(r.props, r.state, r.context), r.state = r.__s;
          while (r.__d && ++s_ < 25);
        r.state = r.__s, r.getChildContext != null && (o = C(C({}, o), r.getChildContext())), x && !a && r.getSnapshotBeforeUpdate != null && (b = r.getSnapshotBeforeUpdate(y, k)), W = f != null && f.type === z && f.key == null ? w_(f.props.children) : f, s = g_(e, O(W) ? W : [W], _, t, o, i, n, u, s, c, l), r.base = _.__e, _.__u &= -161, r.__h.length && u.push(r), g && (r.__E = r.__ = null);
      } catch (D) {
        if (u.length = h, _.__v = null, c || n != null) {
          if (D.then) {
            for (_.__u |= c ? 160 : 128;s && s.nodeType == 8 && s.nextSibling; )
              s = s.nextSibling;
            n != null && (n[n.indexOf(s)] = null), _.__e = s;
          } else if (n != null)
            for (J = n.length;J--; )
              __(n[J]);
        } else
          _.__e = t.__e;
        _.__k == null && (_.__k = t.__k || []), D.then || k_(_), d.__e(D, _, t);
      }
    } else
      n == null && _.__v == t.__v ? (_.__k = t.__k, _.__e = t.__e) : s = _.__e = O_(t.__e, _, t, o, i, n, u, c, l);
  return (f = d.diffed) && f(_), 128 & _.__u ? undefined : s;
}
function k_(e) {
  e && (e.__c && (e.__c.__e = true), e.__k && e.__k.some(k_));
}
function x_(e, _, t) {
  for (var o = 0;o < t.length; o++)
    n_(t[o], t[++o], t[++o]);
  d.__c && d.__c(_, e), e.some(function(i) {
    try {
      e = i.__h, i.__h = [], e.some(function(n) {
        n.call(i);
      });
    } catch (n) {
      d.__e(n, i.__v);
    }
  });
}
function w_(e) {
  return typeof e != "object" || e == null || e.__b > 0 ? e : O(e) ? e.map(w_) : e.constructor !== undefined ? null : C({}, e);
}
function O_(e, _, t, o, i, n, u, s, c) {
  var l, f, h, r, a, y, k, b = t.props || V, { props: g, type: p } = _;
  if (p == "svg" ? i = "http://www.w3.org/2000/svg" : p == "math" ? i = "http://www.w3.org/1998/Math/MathML" : i || (i = "http://www.w3.org/1999/xhtml"), n != null) {
    for (l = 0;l < n.length; l++)
      if ((a = n[l]) && "setAttribute" in a == !!p && (p ? a.localName == p : a.nodeType == 3)) {
        e = a, n[l] = null;
        break;
      }
  }
  if (e == null) {
    if (p == null)
      return document.createTextNode(g);
    e = document.createElementNS(i, p, g.is && g), s && (d.__m && d.__m(_, n), s = false), n = null;
  }
  if (p == null)
    b === g || s && e.data == g || (e.data = g);
  else {
    if (n = p == "textarea" && g.defaultValue != null ? null : n && q.call(e.childNodes), !s && n != null)
      for (b = {}, l = 0;l < e.attributes.length; l++)
        b[(a = e.attributes[l]).name] = a.value;
    for (l in b)
      a = b[l], l == "dangerouslySetInnerHTML" ? h = a : l == "children" || (l in g) || l == "value" && ("defaultValue" in g) || l == "checked" && ("defaultChecked" in g) || L(e, l, null, a, i);
    for (l in g)
      a = g[l], l == "children" ? r = a : l == "dangerouslySetInnerHTML" ? f = a : l == "value" ? y = a : l == "checked" ? k = a : s && typeof a != "function" || b[l] === a || L(e, l, a, b[l], i);
    if (f)
      s || h && (f.__html == h.__html || f.__html == e.innerHTML) || (e.innerHTML = f.__html), _.__k = [];
    else if (h && (e.innerHTML = ""), g_(_.type == "template" ? e.content : e, O(r) ? r : [r], _, t, o, p == "foreignObject" ? "http://www.w3.org/1999/xhtml" : i, n, u, n ? n[0] : t.__k && U(t, 0), s, c), n != null)
      for (l = n.length;l--; )
        __(n[l]);
    s && p != "textarea" || (l = "value", p == "progress" && y == null ? e.removeAttribute("value") : y != null && (y !== e[l] || p == "progress" && !y || p == "option" && y != b[l]) && L(e, l, y, b[l], i), l = "checked", k != null && k != e[l] && L(e, l, k, b[l], i));
  }
  return e;
}
function n_(e, _, t) {
  try {
    if (typeof e == "function") {
      var o = typeof e.__u == "function";
      o && e.__u(), o && _ == null || (e.__u = e(_));
    } else
      e.current = _;
  } catch (i) {
    d.__e(i, t);
  }
}
function C_(e, _, t) {
  var o, i;
  if (d.unmount && d.unmount(e), (o = e.ref) && (o.current && o.current != e.__e || n_(o, null, _)), (o = e.__c) != null) {
    if (o.componentWillUnmount)
      try {
        o.componentWillUnmount();
      } catch (n) {
        d.__e(n, _);
      }
    o.base = o.__P = o.__n = null;
  }
  if (o = e.__k)
    for (i = 0;i < o.length; i++)
      o[i] && C_(o[i], _, t || typeof e.type != "function");
  t || __(e.__e), e.__c = e.__ = e.__e = undefined;
}
function z_(e, _, t) {
  return this.constructor(e, t);
}
function G_(e, _, t) {
  var o, i, n, u;
  _ == document && (_ = document.documentElement), d.__ && d.__(e, _), i = (o = typeof t == "function") ? null : t && t.__k || _.__k, n = [], u = [], t_(_, e = (!o && t || _).__k = e_(z, null, [e]), i || V, V, _.namespaceURI, !o && t ? [t] : i ? null : _.firstChild ? q.call(_.childNodes) : null, n, !o && t ? t : i ? i.__e : _.firstChild, o, u), x_(n, e, u), e.props.children = null;
}
q = B.slice, d = { __e: function(e, _, t, o) {
  for (var i, n, u;_ = _.__; )
    if ((i = _.__c) && !i.__)
      try {
        if ((n = i.constructor) && n.getDerivedStateFromError != null && (i.setState(n.getDerivedStateFromError(e)), u = i.__d), i.componentDidCatch != null && (i.componentDidCatch(e, o || {}), u = i.__d), u)
          return i.__E = i;
      } catch (s) {
        e = s;
      }
  throw e;
} }, p_ = 0, V_ = function(e) {
  return e != null && e.constructor === undefined;
}, A.prototype.setState = function(e, _) {
  var t;
  t = this.__s != null && this.__s != this.state ? this.__s : this.__s = C({}, this.state), typeof e == "function" && (e = e(C({}, t), this.props)), e && C(t, e), e != null && this.__v && (_ && this._sb.push(_), Y(this));
}, A.prototype.forceUpdate = function(e) {
  this.__v && (this.__e = true, e && this.__h.push(e), Y(this));
}, A.prototype.render = z, H = [], h_ = typeof Promise == "function" ? Promise.prototype.then.bind(Promise.resolve()) : setTimeout, d_ = function(e, _) {
  return e.__v.__b - _.__v.__b;
}, j.__r = 0, K = Math.random().toString(8), R = "__d" + K, F = "__a" + K, m_ = /(PointerCapture)$|Capture$/i, Z = 0, Q = a_(false), X = a_(true), v_ = 0;
var P;
var m;
var o_;
var H_;
var E = 0;
var M_ = [];
var v = d;
var { __b: P_, __r: S_, diffed: U_, __c: D_, unmount: E_, __: N_ } = v;
function N(e, _) {
  v.__h && v.__h(m, e, E || _), E = 0;
  var t = m.__H || (m.__H = { __: [], __h: [] });
  return e >= t.__.length && t.__.push({}), t.__[e];
}
function F_(e) {
  return E = 1, A_(L_, e);
}
function A_(e, _, t) {
  var o = N(P++, 2);
  if (o.t = e, !o.__c && (o.__ = [t ? t(_) : L_(undefined, _), function(s) {
    var c = o.__N ? o.__N[0] : o.__[0], l = o.t(c, s);
    c !== l && (o.__N = [l, o.__[1]], o.__c.setState({}));
  }], o.__c = m, !m.__f)) {
    var i = function(s, c, l) {
      if (!o.__c.__H)
        return true;
      var f = false, h = o.__c.props !== s;
      if (o.__c.__H.__.some(function(a) {
        if (a.__N) {
          f = true;
          var y = a.__[0];
          a.__ = a.__N, a.__N = undefined, y !== a.__[0] && (h = true);
        }
      }), n) {
        var r = n.call(this, s, c, l);
        return f ? r || h : r;
      }
      return !f || h;
    };
    m.__f = true;
    var n = m.shouldComponentUpdate, u = m.componentWillUpdate;
    m.componentWillUpdate = function(s, c, l) {
      if (this.__e) {
        var f = n;
        n = undefined, i(s, c, l), n = f;
      }
      u && u.call(this, s, c, l);
    }, m.shouldComponentUpdate = i;
  }
  return o.__N || o.__;
}
function K_(e, _) {
  var t = N(P++, 3);
  !v.__s && i_(t.__H, _) && (t.__ = e, t.u = _, m.__H.__h.push(t));
}
function W_(e, _) {
  var t = N(P++, 4);
  !v.__s && i_(t.__H, _) && (t.__ = e, t.u = _, m.__h.push(t));
}
function Q_(e) {
  return E = 5, u_(function() {
    return { current: e };
  }, []);
}
function u_(e, _) {
  var t = N(P++, 7);
  return i_(t.__H, _) && (t.__ = e(), t.__H = _, t.__h = e), t.__;
}
function Y_(e, _) {
  return E = 8, u_(function() {
    return e;
  }, _);
}
function te() {
  for (var e;e = M_.shift(); ) {
    var _ = e.__H;
    if (e.__P && _)
      try {
        _.__h.some(G), _.__h.some(r_), _.__h = [];
      } catch (t) {
        _.__h = [], v.__e(t, e.__v);
      }
  }
}
v.__b = function(e) {
  m = null, P_ && P_(e);
}, v.__ = function(e, _) {
  e && _.__k && _.__k.__m && (e.__m = _.__k.__m), N_ && N_(e, _);
}, v.__r = function(e) {
  S_ && S_(e), P = 0;
  var _ = (m = e.__c).__H;
  _ && (o_ === m ? (_.__h = [], m.__h = [], _.__.some(function(t) {
    t.__N && (t.__ = t.__N), t.u = t.__N = undefined;
  })) : (_.__h.some(G), _.__h.some(r_), _.__h = [], P = 0)), o_ = m;
}, v.diffed = function(e) {
  U_ && U_(e);
  var _ = e.__c;
  _ && _.__H && (_.__H.__h.length && (M_.push(_) !== 1 && H_ === v.requestAnimationFrame || ((H_ = v.requestAnimationFrame) || ne)(te)), _.__H.__.some(function(t) {
    t.u && (t.__H = t.u, t.u = undefined);
  })), o_ = m = null;
}, v.__c = function(e, _) {
  _.some(function(t) {
    try {
      t.__h.some(G), t.__h = t.__h.filter(function(o) {
        return !o.__ || r_(o);
      });
    } catch (o) {
      _.some(function(i) {
        i.__h && (i.__h = []);
      }), _ = [], v.__e(o, t.__v);
    }
  }), D_ && D_(e, _);
}, v.unmount = function(e) {
  E_ && E_(e);
  var _, t = e.__c;
  t && t.__H && (t.__H.__.some(function(o) {
    try {
      G(o);
    } catch (i) {
      _ = i;
    }
  }), t.__H = undefined, _ && v.__e(_, t.__v));
};
var T_ = typeof requestAnimationFrame == "function";
function ne(e) {
  var _, t = function() {
    clearTimeout(o), T_ && cancelAnimationFrame(_), setTimeout(e);
  }, o = setTimeout(t, 35);
  T_ && (_ = requestAnimationFrame(t));
}
function G(e) {
  var _ = m, t = e.__c;
  typeof t == "function" && (e.__c = undefined, t()), m = _;
}
function r_(e) {
  var _ = m;
  e.__c = e.__(), m = _;
}
function i_(e, _) {
  return !e || e.length !== _.length || _.some(function(t, o) {
    return t !== e[o];
  });
}
function L_(e, _) {
  return typeof _ == "function" ? _(e) : _;
}
var I_ = function(e, _, t, o) {
  var i;
  _[0] = 0;
  for (var n = 1;n < _.length; n++) {
    var u = _[n++], s = _[n] ? (_[0] |= u ? 1 : 2, t[_[n++]]) : _[++n];
    u === 3 ? o[0] = s : u === 4 ? o[1] = Object.assign(o[1] || {}, s) : u === 5 ? (o[1] = o[1] || {})[_[++n]] = s : u === 6 ? o[1][_[++n]] += s + "" : u ? (i = e.apply(s, I_(e, s, t, ["", null])), o.push(i), s[0] ? _[0] |= 2 : (_[n - 2] = 0, _[n] = i)) : o.push(s);
  }
  return o;
};
var R_ = new Map;
function l_(e) {
  var _ = R_.get(this);
  return _ || (_ = new Map, R_.set(this, _)), (_ = I_(this, _.get(e) || (_.set(e, _ = function(t) {
    for (var o, i, n = 1, u = "", s = "", c = [0], l = function(r) {
      n === 1 && (r || (u = u.replace(/^\s*\n\s*|\s*\n\s*$/g, ""))) ? c.push(0, r, u) : n === 3 && (r || u) ? (c.push(3, r, u), n = 2) : n === 2 && u === "..." && r ? c.push(4, r, 0) : n === 2 && u && !r ? c.push(5, 0, true, u) : n >= 5 && ((u || !r && n === 5) && (c.push(n, 0, u, i), n = 6), r && (c.push(n, r, 0, i), n = 6)), u = "";
    }, f = 0;f < t.length; f++) {
      f && (n === 1 && l(), l(f));
      for (var h = 0;h < t[f].length; h++)
        o = t[f][h], n === 1 ? o === "<" ? (l(), c = [c], n = 3) : u += o : n === 4 ? u === "--" && o === ">" ? (n = 1, u = "") : u = o + u[0] : s ? o === s ? s = "" : u += o : o === '"' || o === "'" ? s = o : o === ">" ? (l(), n = 1) : n && (o === "=" ? (n = 5, i = u, u = "") : o === "/" && (n < 5 || t[f][h + 1] === ">") ? (l(), n === 3 && (c = c[0]), n = c, (c = c[0]).push(2, 0, n), n = 0) : o === " " || o === "\t" || o === `
` || o === "\r" ? (l(), n = 2) : u += o), n === 3 && u === "!--" && (n = 4, c = c[0]);
    }
    return l(), c;
  }(e)), _), arguments, [])).length > 1 ? _ : _[0];
}
var fe = l_.bind(e_);

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
function getLocalStorageJSON(key) {
  const raw = getLocalStorageItem(key);
  if (!raw)
    return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
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

// web/src/ui/use-agent-state.ts
function useAgentState() {
  const [agentStatus, setAgentStatus] = F_(null);
  const [agentDraft, setAgentDraft] = F_({ text: "", totalLines: 0 });
  const [agentPlan, setAgentPlan] = F_("");
  const [agentThought, setAgentThought] = F_({ text: "", totalLines: 0 });
  const [pendingRequest, setPendingRequest] = F_(null);
  const [currentTurnId, setCurrentTurnId] = F_(null);
  const [steerQueuedTurnId, setSteerQueuedTurnId] = F_(null);
  const lastAgentEventRef = Q_(null);
  const lastSilenceNoticeRef = Q_(0);
  const isAgentRunningRef = Q_(false);
  const draftBufferRef = Q_("");
  const thoughtBufferRef = Q_("");
  const previewResyncPendingRef = Q_(false);
  const previewResyncGenerationRef = Q_(0);
  const pendingRequestRef = Q_(null);
  const stalledPostIdRef = Q_(null);
  const currentTurnIdRef = Q_(null);
  const steerQueuedTurnIdRef = Q_(null);
  const thoughtExpandedRef = Q_(false);
  const draftExpandedRef = Q_(false);
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

// web/src/gi-model-invalidation.ts
var listeners = new Map;
function subscribeModelSettlement(chatJid, listener) {
  let scoped = listeners.get(chatJid);
  if (!scoped)
    listeners.set(chatJid, scoped = new Set);
  scoped.add(listener);
  return () => {
    scoped.delete(listener);
    if (!scoped.size && listeners.get(chatJid) === scoped)
      listeners.delete(chatJid);
  };
}
function notifyModelSettlement(chatJid) {
  for (const listener of [...listeners.get(chatJid) || []]) {
    try {
      listener();
    } catch {}
  }
}

// web/src/gi-session-state.ts
function sessionPickerAgents(sessions) {
  const byId = new Map(sessions.map((session) => [session.id, session]));
  return sessions.map((session) => {
    let root = session;
    const visited = new Set([root.id]);
    while (root.parent_session_id && byId.has(root.parent_session_id) && !visited.has(root.parent_session_id)) {
      root = byId.get(root.parent_session_id);
      visited.add(root.id);
    }
    return {
      chat_jid: `gi:${session.id}`,
      agent_name: (typeof session.title === "string" && session.title ? session.title.replace(/^@/, "") : session.scope?.agent_id) || session.id,
      agent_id: session.scope?.agent_id || "agent",
      branch_id: session.id,
      parent_branch_id: session.parent_session_id || null,
      parent_chat_jid: session.parent_session_id ? `gi:${session.parent_session_id}` : null,
      root_chat_jid: `gi:${root.id}`,
      model: session.state?.selected_model || session.state?.model || "",
      is_active: session.state?.status === "running" || session.state?.status === "queued" || Number(session.state?.queue_count || 0) > 0,
      archived_at: session.state?.archived_at || null,
      pinned: session.state?.pinned === true,
      capabilities: {
        rename: !session.state?.archived_at,
        pin: !session.state?.archived_at,
        archive: Boolean(session.parent_session_id) && !session.state?.archived_at,
        restore: Boolean(session.state?.archived_at)
      }
    };
  });
}
function createSelectionScope() {
  let sessionId = null;
  let generation = 0;
  return {
    select(next) {
      if (next !== sessionId) {
        sessionId = next;
        generation++;
      }
    },
    capture() {
      return { sessionId, generation };
    },
    isCurrent(captured) {
      return captured.sessionId === sessionId && captured.generation === generation;
    },
    current() {
      return sessionId;
    }
  };
}

// web/src/gi-message-media.ts
function projectMessageMedia(payload, sessionId) {
  const blocks = Array.isArray(payload?.content_blocks) ? payload.content_blocks : [];
  const refs = Array.isArray(payload?.media) ? payload.media.filter((ref) => Number.isSafeInteger(ref?.media_id) && ref.media_id > 0 && (!ref.session_id || ref.session_id === sessionId)) : [];
  if (!refs.length)
    return { media_ids: [], content_blocks: blocks.length ? blocks : null };
  const mediaBlocks = refs.map((ref) => ({
    type: /^image\/(png|jpeg|gif|webp|avif|bmp|svg\+xml)$/i.test(ref.content_type || "") ? "image" : "file",
    name: ref.filename || `attachment-${ref.media_id}`,
    mime_type: ref.content_type || "application/octet-stream"
  }));
  return {
    media_ids: refs.map((ref) => ref.media_id),
    content_blocks: [...blocks.filter((block) => block?.type !== "image" && block?.type !== "file"), ...mediaBlocks]
  };
}

// web/src/gi-compose-transfer.ts
function createComposeTransfers() {
  const sessions = new Map;
  const listeners = new Set;
  const emit = () => {
    for (const listener of listeners)
      listener();
  };
  return {
    subscribe(listener) {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    snapshot(session) {
      const value = { uploads: 0, sending: 0, loaded: 0, total: 0, computable: true };
      for (const op of sessions.get(session)?.values() || []) {
        if (op.phase === "send") {
          value.sending++;
          continue;
        }
        value.uploads++;
        value.loaded += op.loaded;
        value.total += op.total;
        value.computable &&= op.computable;
      }
      return value;
    },
    begin(session, phase) {
      const token = Symbol(phase), op = { phase, loaded: 0, total: 0, computable: false };
      let pending = sessions.get(session);
      if (!pending) {
        pending = new Map;
        sessions.set(session, pending);
      }
      pending.set(token, op);
      emit();
      let ended = false;
      return {
        progress(loaded, total, computable) {
          if (ended || phase !== "upload")
            return;
          op.computable = computable && Number.isFinite(total) && total > 0;
          op.total = op.computable ? total : 0;
          op.loaded = Number.isFinite(loaded) ? Math.max(0, op.computable ? Math.min(loaded, total) : loaded) : 0;
          emit();
        },
        end() {
          if (ended)
            return;
          ended = true;
          pending.delete(token);
          if (!pending.size)
            sessions.delete(session);
          emit();
        }
      };
    }
  };
}
var composeTransfers = createComposeTransfers();
function bindComposeSending(root, sending) {
  const owned = new Set;
  const paint = () => {
    const button = root.querySelector(".compose-send-stack .send-btn");
    if (!button)
      return;
    if (sending) {
      button.dataset.giSending = "true";
      button.setAttribute("aria-busy", "true");
      owned.add(button);
    } else {
      delete button.dataset.giSending;
      button.removeAttribute("aria-busy");
    }
  };
  paint();
  const observer = new MutationObserver(paint);
  observer.observe(root, { childList: true, subtree: true });
  return () => {
    observer.disconnect();
    for (const button of owned) {
      delete button.dataset.giSending;
      button.removeAttribute("aria-busy");
    }
  };
}

// web/src/gi-sse-client.ts
var API_BASE = "";

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
    const source = new EventSource(API_BASE + "/sse/stream" + query);
    this.eventSource = source;
    const current = () => this.eventSource === source;
    const bindJsonEvent = (eventType) => {
      source.addEventListener(eventType, (e) => {
        if (!current())
          return;
        this.resetStaleMonitor();
        try {
          const data = JSON.parse(e.data);
          this.onEvent(eventType, data);
        } catch {}
      });
    };
    source.addEventListener("connected", (event) => {
      if (!current())
        return;
      this.connecting = false;
      this.reconnectDelay = 1000;
      this.setStatus("connected");
      this.resetStaleMonitor();
      try {
        this.onEvent("connected", JSON.parse(event.data));
      } catch {}
    });
    source.addEventListener("heartbeat", () => {
      if (!current())
        return;
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
    bindJsonEvent("queue_changed");
    for (const event of ["compaction_started", "compaction_completed", "compaction_cancelled", "compaction_suppressed", "compaction_failed"])
      bindJsonEvent(event);
    bindJsonEvent("workspace_update");
    bindJsonEvent("agent_draft");
    bindJsonEvent("agent_draft_delta");
    bindJsonEvent("agent_thought");
    bindJsonEvent("agent_thought_delta");
    bindJsonEvent("routing_decision");
    bindJsonEvent("routing_incoming");
    bindJsonEvent("model_changed");
    bindJsonEvent("ui_theme");
    bindJsonEvent("ui_meters");
    source.onerror = () => {
      if (!current())
        return;
      this.eventSource = null;
      source.close();
      this.clearStaleMonitor();
      this.connecting = false;
      this.setStatus("disconnected");
      this.scheduleReconnect();
    };
  }
  disconnect() {
    this.connecting = false;
    this.clearStaleMonitor();
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    const source = this.eventSource;
    this.eventSource = null;
    source?.close();
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

// web/src/api.ts
var API_BASE2 = "";
async function request(url, options = {}) {
  const startedAt = typeof performance !== "undefined" && typeof performance.now === "function" ? performance.now() : Date.now();
  let response;
  try {
    response = await fetch(API_BASE2 + url, {
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
    throw Object.assign(new Error(err.error || `HTTP ${response.status}`), { status: response.status });
  }
  return response.json();
}
var quickActionsReady = false;
var quickActionsRevision = 0;
var isQuickActionsReady = () => quickActionsReady;
function resetQuickActionsReadiness() {
  quickActionsReady = false;
  ++quickActionsRevision;
}
async function getQuickActionsSettings() {
  const revision = ++quickActionsRevision;
  quickActionsReady = false;
  try {
    return { settings: await request("/api/quick-actions") };
  } catch {
    return { settings: { workspaceCommands: [], slashCommands: [] } };
  } finally {
    if (revision === quickActionsRevision)
      quickActionsReady = true;
  }
}
async function getAgentCommands(_chatJid = null) {
  return request("/api/quick-actions").then((data) => ({ commands: data.commands || [] }));
}
var DEFAULT_CHAT_JID = "web:default";
function sessionToChatJid(sessionId) {
  return sessionId ? `gi:${sessionId}` : DEFAULT_CHAT_JID;
}
async function getTimeline(limit = 50, beforeId = null, chatJid = null, after = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    return { posts: [] };
  let url = `/api/sessions/${encodeURIComponent(sessionId)}/messages?limit=${limit}`;
  if (beforeId)
    url += `&before=${encodeURIComponent(beforeId)}`;
  if (after)
    url += `&after=${encodeURIComponent(after)}`;
  const data = await request(url);
  const messages = data.messages || [];
  return {
    hasMore: data.has_more === true,
    before: data.before || null,
    after: data.after || null,
    posts: messages.map((m) => ({
      id: m.id,
      chat_jid: chatJid,
      content: m.content,
      timestamp: m.created_at,
      sender: m.role === "user" ? "user" : "agent",
      is_from_me: m.role === "user",
      is_bot_message: m.role === "assistant",
      data: {
        type: m.role === "assistant" ? "agent_response" : "user_message",
        content: m.content,
        thread_id: null,
        agent_id: m.payload?.agent_id || (m.role === "assistant" ? "agent" : null),
        ...projectMessageMedia(m.payload, sessionId),
        content_meta: null,
        link_previews: null,
        kind: m.payload?.kind || null,
        source: m.payload?.source || null,
        clipped: m.payload?.clipped || false
      }
    }))
  };
}
async function searchPosts(query, limit = 50, offset = 0, chatJid = null, scope = "current", _rootChatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    return { posts: [] };
  const params = new URLSearchParams({ q: query, scope, limit: String(limit), offset: String(offset) });
  const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/search?${params}`);
  const messages = data.messages || [];
  return { posts: messages.map((m) => ({
    id: m.id,
    chat_jid: sessionToChatJid(m.session_id),
    content: m.content,
    timestamp: m.created_at,
    sender: m.role === "user" ? "user" : "agent",
    is_from_me: m.role === "user",
    is_bot_message: m.role === "assistant",
    data: { type: m.role === "assistant" ? "agent_response" : "user_message", content: m.content, thread_id: null, agent_id: m.payload?.agent_id || (m.role === "assistant" ? "agent" : null), ...projectMessageMedia(m.payload, m.session_id) }
  })) };
}
async function getSystemMetrics() {
  return request("/api/system-metrics").catch(() => null);
}
async function getAgentStatus(agentId, chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    return null;
  const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/activity`);
  return {
    ...data,
    type: data.status === "running" ? "tool_call" : "intent",
    title: data.status === "cancelling" ? "Cancelling…" : data.status === "running" ? "Working…" : ""
  };
}
async function getSessionCompaction(chatJid) {
  if (!chatJid?.startsWith("gi:"))
    return { available: false };
  return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/compaction`);
}
async function compactSession(chatJid, token) {
  if (!chatJid?.startsWith("gi:") || !token)
    throw new Error("No compaction snapshot");
  return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/compaction`, { method: "POST", body: JSON.stringify({ token }) });
}
async function cancelSessionRun(chatJid, turnId) {
  if (!chatJid?.startsWith("gi:") || !turnId)
    throw new Error("No active run to stop");
  return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/activity`, { method: "POST", body: JSON.stringify({ turn_id: turnId }) });
}
async function getGiProviders() {
  return request("/api/settings/providers");
}
async function saveGiProviderKey(provider, revision, key) {
  return request("/api/settings/providers", { method: "PATCH", body: JSON.stringify({ provider, revision, key }) });
}
async function removeGiProviderKey(provider, revision) {
  return request("/api/settings/providers", { method: "DELETE", body: JSON.stringify({ provider, revision }) });
}
async function getGiCompactionPolicy() {
  return request("/api/settings/compaction");
}
async function saveGiCompactionPolicy(value) {
  return request("/api/settings/compaction", { method: "PATCH", body: JSON.stringify(value) });
}
async function getGiIdentity() {
  return request("/api/settings/identity");
}
async function saveGiIdentity(value) {
  return request("/api/settings/identity", { method: "PATCH", body: JSON.stringify(value) });
}
async function getGiSettingsSnapshot() {
  return request("/api/runtime/config");
}
async function getAgentModels(chatJid = null) {
  if (chatJid?.startsWith("gi:"))
    return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/model`);
  const data = await request("/api/runtime/config");
  const modelOptions = Array.isArray(data.model_options) ? data.model_options : [];
  const models = modelOptions.length > 0 ? modelOptions : (data.enabled_models || []).map((id) => ({
    id,
    provider: data.default_provider || "",
    label: id
  }));
  return {
    models,
    model_options: modelOptions,
    provider_options: Array.isArray(data.provider_options) ? data.provider_options : [],
    current: data.current || data.default_model || "",
    thinking_level: data.default_thinking_level || data.thinking_level || "",
    supports_thinking: Boolean(data.supports_thinking)
  };
}
async function selectAgentModel(chatJid, model) {
  if (!chatJid?.startsWith("gi:"))
    throw new Error("No model destination session");
  try {
    return await request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/model`, { method: "PATCH", body: JSON.stringify({ model }) });
  } finally {
    notifyModelSettlement(chatJid);
  }
}
async function getAgentQueueState(chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    return { items: [] };
  const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/queue`);
  return { activeTurnId: data.active_turn_id || null, items: (data.items || []).map((turn) => ({
    id: turn.id,
    text: turn.prompt,
    content: turn.prompt,
    chat_jid: chatJid,
    metadata: turn.metadata,
    created_at: turn.created_at,
    phase: turn.phase
  })) };
}
async function steerAgentQueueItem(itemId, chatJid, activeTurnId) {
  if (!chatJid?.startsWith("gi:") || !activeTurnId)
    throw new Error("Steer requires a matching active run");
  return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}/queue/${encodeURIComponent(itemId)}/steer`, {
    method: "POST",
    body: JSON.stringify({ active_turn_id: activeTurnId })
  });
}
async function removeAgentQueueItem(turnId, chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    throw new Error("No queue session");
  return request(`/api/sessions/${encodeURIComponent(sessionId)}/queue/${encodeURIComponent(turnId)}`, { method: "DELETE" });
}
async function getActiveChatAgents() {
  const data = await request("/api/sessions");
  const sessions = data.sessions || [];
  return {
    agents: sessionPickerAgents(sessions)
  };
}
async function getChatBranches(rootChatJid = null, _options = {}) {
  const data = await request("/api/sessions").catch(() => ({ sessions: [] }));
  let sessions = data.sessions || [];
  if (rootChatJid?.startsWith("gi:")) {
    const rootId = rootChatJid.slice(3);
    const byParent = new Map;
    for (const s of sessions) {
      const key = s.parent_session_id || "";
      const bucket = byParent.get(key) || [];
      bucket.push(s);
      byParent.set(key, bucket);
    }
    const wanted = new Set([rootId]);
    const queue = [rootId];
    while (queue.length > 0) {
      const current = queue.shift();
      const children = byParent.get(current) || [];
      for (const child of children) {
        if (!wanted.has(child.id)) {
          wanted.add(child.id);
          queue.push(child.id);
        }
      }
    }
    sessions = sessions.filter((s) => wanted.has(s.id));
  }
  const mapped = sessions.map((s) => ({
    chat_jid: sessionToChatJid(s.id),
    label: s.title || `@${s.scope?.agent_id || s.id}`,
    updated_at: s.updated_at,
    parent_chat_jid: s.parent_session_id ? sessionToChatJid(s.parent_session_id) : null,
    agent_id: s.scope?.agent_id || "agent"
  }));
  return { branches: mapped, chats: mapped };
}
async function forkChatBranch(sourceChatJid, options = {}) {
  const sessionId = sourceChatJid?.startsWith("gi:") ? sourceChatJid.slice(3) : null;
  if (!sessionId)
    throw new Error("No source session to fork");
  return request(`/api/sessions/${encodeURIComponent(sessionId)}/fork`, {
    method: "POST",
    body: JSON.stringify({ title: options?.title || null, agent_id: options?.agent_id || null })
  });
}
async function mutateChatSession(chatJid, mutation) {
  if (!chatJid?.startsWith("gi:") || !chatJid.slice(3))
    throw new Error("Invalid session identifier");
  return request(`/api/sessions/${encodeURIComponent(chatJid.slice(3))}`, {
    method: "PATCH",
    body: JSON.stringify(mutation)
  });
}
async function renameChatBranch(chatJid, options = {}) {
  return mutateChatSession(chatJid, { action: "rename", title: options.title });
}
async function pinChatSession(chatJid, pinned) {
  return mutateChatSession(chatJid, { action: "pin", pinned });
}
async function pruneChatBranch(chatJid) {
  return mutateChatSession(chatJid, { action: "archive" });
}
async function restoreChatBranch(chatJid, _options = {}) {
  return mutateChatSession(chatJid, { action: "restore" });
}
async function deletePost(postId, cascade = false, chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    throw new Error("No message destination session");
  return request(`/api/sessions/${encodeURIComponent(sessionId)}/messages/${encodeURIComponent(postId)}?cascade=${cascade}`, { method: "DELETE" });
}
async function sendAgentMessage(agentId, content, _threadId = null, _mediaIds = [], mode = null, chatJid = null, options = {}) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    throw new Error("No active session");
  const intent = mode === "steer" ? "steer" : mode === "queue" ? "queue" : "prompt";
  const targetAgentId = agentId && agentId !== "default" ? String(agentId).replace(/^@/, "") : null;
  const payload = {
    prompt: content,
    intent,
    target_agent_id: targetAgentId,
    media: _mediaIds.map((media_id) => ({ media_id, session_id: sessionId })),
    client_request_id: options?.client_request_id || undefined
  };
  if (options?.parent_turn_id) {
    payload.parent_turn_id = options.parent_turn_id;
  }
  const activity = composeTransfers.begin(sessionId, "send");
  try {
    return await request(`/api/sessions/${encodeURIComponent(sessionId)}/prompt`, {
      method: "POST",
      body: JSON.stringify(payload)
    });
  } finally {
    activity.end();
  }
}
async function uploadMedia(file, chatJid = null) {
  const sessionId = chatJid?.startsWith("gi:") ? chatJid.slice(3) : null;
  if (!sessionId)
    throw new Error("No attachment destination session");
  if (file.size > 10 * 1024 * 1024)
    throw new Error("Media exceeds 10 MiB limit");
  const activity = composeTransfers.begin(sessionId, "upload");
  try {
    const form = new FormData;
    form.append("file", file, file.name);
    const encoded = new Response(form);
    const contentType = encoded.headers.get("content-type");
    const body = await encoded.arrayBuffer();
    return await new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest;
      xhr.open("POST", `/api/sessions/${encodeURIComponent(sessionId)}/media`);
      xhr.setRequestHeader("Content-Type", contentType);
      xhr.upload.onprogress = (event) => activity.progress(event.loaded, event.total, event.lengthComputable);
      xhr.onerror = () => reject(new TypeError("Upload network request failed"));
      xhr.onabort = () => reject(new DOMException("Upload aborted", "AbortError"));
      xhr.onload = () => {
        let data = {};
        try {
          data = JSON.parse(xhr.responseText);
        } catch {}
        if (xhr.status < 200 || xhr.status >= 300) {
          reject(new Error(data.error || `Upload failed: HTTP ${xhr.status}`));
          return;
        }
        if (!data.media?.id) {
          reject(new Error("Upload returned no media identifier"));
          return;
        }
        resolve({ ...data.media, id: data.media.id });
      };
      xhr.send(body);
    });
  } finally {
    activity.end();
  }
}
async function getMediaInfo(mediaId) {
  return request(`/api/media/${mediaId}`).catch(() => null);
}
function getMediaUrl(mediaId) {
  return `/api/media/${mediaId}/raw`;
}
function getThumbnailUrl(mediaId) {
  return getMediaUrl(mediaId);
}
async function submitAdaptiveCardAction(_payload) {
  return null;
}
async function getWorkspaceTree(path = "", depth = 1, showHidden = false) {
  const query = new URLSearchParams({ path: path || ".", depth: String(depth), show_hidden: String(showHidden) });
  return { root: await request(`/api/workspace/tree?${query}`) };
}
async function getWorkspaceFile(path, maxBytes = 20000) {
  return request(`/api/workspace/file?path=${encodeURIComponent(path)}&max_bytes=${maxBytes}`);
}
async function getWorkspaceIndexStatus(scope = "all") {
  return request(`/api/workspace/index?scope=${encodeURIComponent(scope)}`);
}
async function reindexWorkspace(scope = "all") {
  return request(`/api/workspace/index?scope=${encodeURIComponent(scope)}`, { method: "POST" });
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
async function setWorkspaceVisibility(visible, showHidden) {
  return { visible, show_hidden: showHidden };
}
function getWorkspaceDownloadUrl(path) {
  return `/api/workspace/file?path=${encodeURIComponent(path)}`;
}
async function getWorkspaceBranch(_chatJid = null) {
  return null;
}
async function reorderAgentQueueItem(payload) {
  if (!payload.chatJid?.startsWith("gi:"))
    throw new Error("No queue session");
  return request(`/api/sessions/${encodeURIComponent(payload.chatJid.slice(3))}/queue`, { method: "PATCH", body: JSON.stringify({ expected: payload.expected, order: payload.order }) });
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
  const handlePageHide = () => {
    pendingWake = true;
    sse.disconnect?.();
  };
  const handleVisibilityChange = () => {
    handleVisibleReturn();
  };
  win.addEventListener("focus", handleWindowFocus);
  win.addEventListener("pageshow", handlePageShow);
  win.addEventListener("pagehide", handlePageHide);
  doc.addEventListener("visibilitychange", handleVisibilityChange);
  return () => {
    win.removeEventListener("focus", handleWindowFocus);
    win.removeEventListener("pageshow", handlePageShow);
    win.removeEventListener("pagehide", handlePageHide);
    doc.removeEventListener("visibilitychange", handleVisibilityChange);
  };
}
function useSseConnection({ handleSseEvent, handleConnectionStatusChange, loadPosts, onWake, chatJid, selectionKey = chatJid }) {
  const selectionRef = Q_(selectionKey);
  selectionRef.current = selectionKey;
  const sseEventRef = Q_(handleSseEvent);
  sseEventRef.current = handleSseEvent;
  const statusChangeRef = Q_(handleConnectionStatusChange);
  statusChangeRef.current = handleConnectionStatusChange;
  const loadPostsRef = Q_(loadPosts);
  loadPostsRef.current = loadPosts;
  const onWakeRef = Q_(onWake);
  onWakeRef.current = onWake;
  K_(() => {
    let active = true;
    const sse = new SSEClient((type, data) => {
      if (active && selectionRef.current === selectionKey)
        sseEventRef.current(type, data);
    }, (status) => {
      if (active && selectionRef.current === selectionKey)
        statusChangeRef.current(status);
    }, { chatJid });
    sse.connect();
    const disposeWakeLifecycle = bindSseWakeLifecycle({
      sse,
      onWake: () => onWakeRef.current?.()
    });
    return () => {
      active = false;
      disposeWakeLifecycle();
      sse.disconnect();
    };
  }, [chatJid, selectionKey]);
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
  const full = hex.length === 3 ? hex.split("").map((c) => c + c).join("") : hex;
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
  const r = parseInt(match[1], 10);
  const g = parseInt(match[2], 10);
  const b = parseInt(match[3], 10);
  if (![r, g, b].every((v) => Number.isFinite(v)))
    return null;
  const hex = `#${[r, g, b].map((v) => v.toString(16).padStart(2, "0")).join("")}`;
  return { r, g, b, hex };
}
function parseColor(input) {
  return parseHexColor(input) || parseCssColor(input);
}
function mixColors(base, overlay, ratio) {
  const r = Math.round(base.r + (overlay.r - base.r) * ratio);
  const g = Math.round(base.g + (overlay.g - base.g) * ratio);
  const b = Math.round(base.b + (overlay.b - base.b) * ratio);
  return `rgb(${r} ${g} ${b})`;
}
function rgbaColor(color, alpha) {
  return `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`;
}
function relativeLuminance(c) {
  const rs = c.r / 255, gs = c.g / 255, bs = c.b / 255;
  const r = rs <= 0.03928 ? rs / 12.92 : Math.pow((rs + 0.055) / 1.055, 2.4);
  const g = gs <= 0.03928 ? gs / 12.92 : Math.pow((gs + 0.055) / 1.055, 2.4);
  const b = bs <= 0.03928 ? bs / 12.92 : Math.pow((bs + 0.055) / 1.055, 2.4);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
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

// web/src/gi-appearance-state.ts
var APPEARANCE_KEY = "gi_browser_appearance_v1";
var defaultAppearance = { version: 1, theme: "default", tint: "" };
function validateAppearance(value, presets) {
  if (!value || value.version !== 1 || typeof value.theme !== "string" || !presets.includes(value.theme)) {
    throw new Error("Choose a supported theme preset.");
  }
  if (typeof value.tint !== "string")
    throw new Error("Tint must be a hex colour.");
  let tint = value.tint.trim().toLowerCase();
  if (tint && !/^#[0-9a-f]{3}([0-9a-f]{3})?$/.test(tint))
    throw new Error("Use #RGB or #RRGGBB for the tint, or leave it empty.");
  if (tint.length === 4)
    tint = "#" + [...tint.slice(1)].map((c) => c + c).join("");
  return { version: 1, theme: value.theme, tint: value.theme === "default" ? tint : "" };
}
function readAppearance(storage, presets) {
  try {
    const raw = storage.getItem(APPEARANCE_KEY);
    return raw ? validateAppearance(JSON.parse(raw), presets) : null;
  } catch {
    return null;
  }
}
function saveAppearance(storage, value, presets) {
  const next = validateAppearance(value, presets);
  storage.setItem(APPEARANCE_KEY, JSON.stringify(next));
  return next;
}

// web/src/gi-appearance.ts
var appearancePresets = Object.keys(THEME_PRESETS);
var changeEvent = "gi:appearance-changed";
function readStoredAppearance() {
  try {
    return readAppearance(window.localStorage, appearancePresets);
  } catch {
    return null;
  }
}
function currentAppearance() {
  const stored = readStoredAppearance();
  if (stored)
    return stored;
  const theme = document.documentElement.dataset.colorTheme || "default";
  return { version: 1, theme: appearancePresets.includes(theme) ? theme : "default", tint: document.documentElement.dataset.tint || "" };
}
function render(value) {
  applyThemeState(value, { persist: false });
  window.dispatchEvent(new CustomEvent(changeEvent, { detail: value }));
}
function persistAppearance(value) {
  const saved = saveAppearance(window.localStorage, value, appearancePresets);
  render(saved);
  return saved;
}
function subscribeAppearance(onChange) {
  const listener = (event) => onChange(event.detail);
  window.addEventListener(changeEvent, listener);
  return () => window.removeEventListener(changeEvent, listener);
}
function initGiAppearance() {
  const saved = readStoredAppearance();
  if (saved)
    render(saved);
  const storage = (event) => {
    if (event.storageArea !== window.localStorage || event.key !== APPEARANCE_KEY && event.key !== null)
      return;
    const next = readStoredAppearance();
    if (next)
      render(next);
    else if (event.newValue === null)
      render(defaultAppearance);
  };
  window.addEventListener("storage", storage);
  return () => window.removeEventListener("storage", storage);
}

// web/src/ui/chat-window.ts
function isStandaloneWebAppMode(runtime = {}) {
  const win = runtime.window ?? (typeof window !== "undefined" ? window : null);
  const nav = runtime.navigator ?? (typeof navigator !== "undefined" ? navigator : null);
  if (nav && nav.standalone === true) {
    return true;
  }
  if (!win || typeof win.matchMedia !== "function") {
    return false;
  }
  const queries = [
    "(display-mode: standalone)",
    "(display-mode: minimal-ui)",
    "(display-mode: fullscreen)"
  ];
  return queries.some((query) => {
    try {
      return Boolean(win.matchMedia(query)?.matches);
    } catch {
      return false;
    }
  });
}
function isMobileBrowserMode(runtime = {}) {
  const win = runtime.window ?? (typeof window !== "undefined" ? window : null);
  const nav = runtime.navigator ?? (typeof navigator !== "undefined" ? navigator : null);
  if (!win && !nav)
    return false;
  const userAgent = String(nav?.userAgent || "");
  const maxTouchPoints = Number(nav?.maxTouchPoints || 0);
  const mobileUa = /Android|webOS|iPhone|iPad|iPod|Mobile|Windows Phone/i.test(userAgent);
  const coarsePointer = (() => {
    if (!win || typeof win.matchMedia !== "function")
      return false;
    try {
      return Boolean(win.matchMedia("(pointer: coarse)")?.matches || win.matchMedia("(any-pointer: coarse)")?.matches);
    } catch {
      return false;
    }
  })();
  return Boolean(mobileUa || maxTouchPoints > 1 || coarsePointer);
}

// web/src/ui/pwa-display-scale.ts
var PWA_DISPLAY_SCALE_STORAGE_KEY = "piclawPwaDisplayScalePercent";
var PWA_DISPLAY_SCALE_EVENT = "piclaw:pwa-display-scale-changed";
var DEFAULT_PWA_DISPLAY_SCALE_PERCENT = 100;
var MIN_PWA_DISPLAY_SCALE_PERCENT = 20;
var MAX_PWA_DISPLAY_SCALE_PERCENT = 115;
var PWA_DISPLAY_SCALE_STEP_PERCENT = 5;
var DEFAULT_VIEWPORT_CONTENT = "width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover";
function getRuntimeWindow(runtime = typeof window !== "undefined" ? window : null) {
  return runtime && typeof runtime === "object" ? runtime : null;
}
function getRuntimeNavigator(runtime) {
  return runtime?.navigator || (typeof navigator !== "undefined" ? navigator : null);
}
function parseScalePercent(value) {
  if (typeof value === "number")
    return Number.isFinite(value) ? value : null;
  if (typeof value !== "string")
    return null;
  const trimmed = value.trim();
  if (!trimmed)
    return null;
  const numeric = trimmed.endsWith("%") ? trimmed.slice(0, -1) : trimmed;
  const parsed = Number(numeric);
  if (!Number.isFinite(parsed))
    return null;
  return parsed > 0 && parsed <= 2 ? parsed * 100 : parsed;
}
function normalizePwaDisplayScalePercent(value, fallback = DEFAULT_PWA_DISPLAY_SCALE_PERCENT) {
  const parsed = parseScalePercent(value);
  const base = parsed ?? parseScalePercent(fallback) ?? DEFAULT_PWA_DISPLAY_SCALE_PERCENT;
  return Math.min(MAX_PWA_DISPLAY_SCALE_PERCENT, Math.max(MIN_PWA_DISPLAY_SCALE_PERCENT, Math.round(base)));
}
function formatPwaDisplayScaleRatio(percent) {
  const ratio = normalizePwaDisplayScalePercent(percent) / 100;
  return ratio.toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
}
function readStoredPwaDisplayScalePercent(runtime = typeof window !== "undefined" ? window : null) {
  const runtimeWindow = getRuntimeWindow(runtime);
  try {
    return normalizePwaDisplayScalePercent(runtimeWindow?.localStorage?.getItem(PWA_DISPLAY_SCALE_STORAGE_KEY));
  } catch {
    return DEFAULT_PWA_DISPLAY_SCALE_PERCENT;
  }
}
function isMobileStandalonePwa(runtime = typeof window !== "undefined" ? window : null) {
  const runtimeWindow = getRuntimeWindow(runtime);
  const runtimeNavigator = getRuntimeNavigator(runtimeWindow);
  return isStandaloneWebAppMode({ window: runtimeWindow, navigator: runtimeNavigator }) && isMobileBrowserMode({ window: runtimeWindow, navigator: runtimeNavigator });
}
function buildPwaDisplayScaleViewportContent(percent, options = {}) {
  const normalized = normalizePwaDisplayScalePercent(percent);
  if (!options.applies || normalized === DEFAULT_PWA_DISPLAY_SCALE_PERCENT) {
    return DEFAULT_VIEWPORT_CONTENT;
  }
  const scale = formatPwaDisplayScaleRatio(normalized);
  return `width=device-width, initial-scale=${scale}, minimum-scale=${scale}, maximum-scale=${scale}, user-scalable=no, viewport-fit=cover`;
}
function applyPwaDisplayScale(runtime = typeof window !== "undefined" ? window : null) {
  const runtimeWindow = getRuntimeWindow(runtime);
  const percent = readStoredPwaDisplayScalePercent(runtimeWindow);
  const applied = isMobileStandalonePwa(runtimeWindow) && percent !== DEFAULT_PWA_DISPLAY_SCALE_PERCENT;
  const content = buildPwaDisplayScaleViewportContent(percent, { applies: applied });
  const viewport = runtimeWindow?.document?.querySelector?.('meta[name="viewport"]');
  viewport?.setAttribute?.("content", content);
  return { percent, applied, content };
}
function persistPwaDisplayScalePercent(percent, runtime = typeof window !== "undefined" ? window : null) {
  const runtimeWindow = getRuntimeWindow(runtime);
  const normalized = normalizePwaDisplayScalePercent(percent);
  try {
    runtimeWindow?.localStorage?.setItem(PWA_DISPLAY_SCALE_STORAGE_KEY, String(normalized));
  } catch (error) {
    console.warn("[pwa-display-scale] Unable to persist scale preference; applying in-memory value only.", error);
  }
  applyPwaDisplayScale(runtimeWindow);
  if (runtimeWindow?.dispatchEvent) {
    const event = typeof CustomEvent === "function" ? new CustomEvent(PWA_DISPLAY_SCALE_EVENT, { detail: { percent: normalized } }) : { type: PWA_DISPLAY_SCALE_EVENT, detail: { percent: normalized } };
    runtimeWindow.dispatchEvent(event);
  }
  return normalized;
}
function installPwaDisplayScaleSync(runtime = typeof window !== "undefined" ? window : null) {
  const runtimeWindow = getRuntimeWindow(runtime);
  if (!runtimeWindow?.addEventListener)
    return;
  const sync = () => applyPwaDisplayScale(runtimeWindow);
  sync();
  const onStorage = (event) => {
    if (!event || event.key === null || event.key === PWA_DISPLAY_SCALE_STORAGE_KEY)
      sync();
  };
  const onScaleChanged = () => sync();
  runtimeWindow.addEventListener("storage", onStorage);
  runtimeWindow.addEventListener("focus", sync);
  runtimeWindow.addEventListener(PWA_DISPLAY_SCALE_EVENT, onScaleChanged);
  const mediaQueries = ["standalone", "fullscreen", "minimal-ui"].map((mode) => {
    try {
      return runtimeWindow.matchMedia?.(`(display-mode: ${mode})`);
    } catch {
      return null;
    }
  }).filter(Boolean);
  for (const media of mediaQueries) {
    if (media.addEventListener)
      media.addEventListener("change", sync);
    else if (media.addListener)
      media.addListener(sync);
  }
  return () => {
    runtimeWindow.removeEventListener("storage", onStorage);
    runtimeWindow.removeEventListener("focus", sync);
    runtimeWindow.removeEventListener(PWA_DISPLAY_SCALE_EVENT, onScaleChanged);
    for (const media of mediaQueries) {
      if (media.removeEventListener)
        media.removeEventListener("change", sync);
      else if (media.removeListener)
        media.removeListener(sync);
    }
  };
}

// web/src/gi-display-scale.ts
function installGiDisplayScale(runtime = window) {
  const cleanupScale = installPwaDisplayScaleSync(runtime);
  const sync = () => runtime.document.documentElement.classList.toggle("gi-standalone-display", isStandaloneWebAppMode({ window: runtime, navigator: runtime.navigator }));
  const media = ["standalone", "fullscreen", "minimal-ui"].map((mode) => {
    try {
      return runtime.matchMedia?.(`(display-mode: ${mode})`);
    } catch {
      return null;
    }
  }).filter(Boolean);
  sync();
  runtime.addEventListener("focus", sync);
  for (const query of media) {
    if (query.addEventListener)
      query.addEventListener("change", sync);
    else
      query.addListener?.(sync);
  }
  return () => {
    cleanupScale?.();
    runtime.removeEventListener("focus", sync);
    for (const query of media) {
      if (query.removeEventListener)
        query.removeEventListener("change", sync);
      else
        query.removeListener?.(sync);
    }
    runtime.document.documentElement.classList.remove("gi-standalone-display");
  };
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
// web/src/panes/editor-popout-transfer.ts
var EDITOR_POPOUT_STATE_TTL_MS = 5 * 60 * 1000;
// node_modules/@assemblyscript/loader/index.js
var ARRAYBUFFERVIEW = 1 << 0;
var ARRAY = 1 << 1;
var STATICARRAY = 1 << 2;
var VAL_SIGNED = 1 << 11;
var VAL_FLOAT = 1 << 12;
var VAL_MANAGED = 1 << 14;
var THIS = Symbol();
var utf16 = new TextDecoder("utf-16le", { fatal: true });
Object.hasOwn = Object.hasOwn || function(obj, prop) {
  return Object.prototype.hasOwnProperty.call(obj, prop);
};

// web/src/panes/vnc-input.ts
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
for (let i = 1;i <= 12; i += 1) {
  KEYSYM_BY_KEY[`F${i}`] = 65470 + (i - 1);
}

// web/src/panes/vnc-auth.ts
var REVERSED_BITS = new Uint8Array(256);
for (let value = 0;value < 256; value += 1) {
  let reversed = 0;
  for (let bit = 0;bit < 8; bit += 1) {
    reversed = reversed << 1 | value >> bit & 1;
  }
  REVERSED_BITS[value] = reversed;
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
} from "/editor-vendor/codemirror.js";
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
  tokens.sort((a, b) => a.from - b.from || a.to - b.to);
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
  for (let i = 0;i < maxDepth; i += 1) {
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
  for (let i = 1;i < lines.length; i += 1) {
    if (/^(---|\.\.\.)\s*$/.test(lines[i])) {
      closingIndex = i;
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
  for (let i = 0;i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
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
    } catch (e) {
      return `<span class="math-error" title="${escapeHtmlAttr(e.message)}">${match}</span>`;
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
    const pts = pointsStr.trim().split(/\s+/).map((p) => {
      const [x, y] = p.split(",").map(Number);
      return { x, y };
    });
    if (pts.length < 3) {
      return `<polyline${before}points="${pointsStr}"${after}/>`;
    }
    const parts = [`M ${pts[0].x},${pts[0].y}`];
    for (let i = 1;i < pts.length - 1; i++) {
      const prev = pts[i - 1];
      const curr = pts[i];
      const next = pts[i + 1];
      const dxIn = curr.x - prev.x;
      const dyIn = curr.y - prev.y;
      const dxOut = next.x - curr.x;
      const dyOut = next.y - curr.y;
      const lenIn = Math.sqrt(dxIn * dxIn + dyIn * dyIn);
      const lenOut = Math.sqrt(dxOut * dxOut + dyOut * dyOut);
      const r = Math.min(radius, lenIn / 2, lenOut / 2);
      if (r < 0.5) {
        parts.push(`L ${curr.x},${curr.y}`);
        continue;
      }
      const ax = curr.x - dxIn / lenIn * r;
      const ay = curr.y - dyIn / lenIn * r;
      const bx = curr.x + dxOut / lenOut * r;
      const by = curr.y + dyOut / lenOut * r;
      const cross = dxIn * dyOut - dyIn * dxOut;
      const sweep = cross > 0 ? 1 : 0;
      parts.push(`L ${ax},${ay}`);
      parts.push(`A ${r},${r} 0 0 ${sweep} ${bx},${by}`);
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
    } catch (e) {
      console.error("Mermaid render error:", e);
      const pre = document.createElement("pre");
      pre.className = "mermaid-error";
      pre.textContent = `Diagram error: ${e.message}`;
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

// web/src/panes/workspace-preview-pane.ts
function escapeHtml2(value) {
  return String(value || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#39;");
}
function rewriteMarkdownImagePath(src, markdownPath) {
  const raw = String(src || "").trim();
  if (!raw)
    return raw;
  if (/^[a-zA-Z][a-zA-Z\d+.-]*:/.test(raw) || raw.startsWith("#") || raw.startsWith("data:") || raw.startsWith("blob:")) {
    return raw;
  }
  const match = raw.match(/^([^?#]*)(\?[^#]*)?(#.*)?$/);
  const relPath = match?.[1] || raw;
  const query = match?.[2] || "";
  const hash = match?.[3] || "";
  const baseDir = String(markdownPath || "").split("/").slice(0, -1).join("/");
  const isAbsolute = relPath.startsWith("/");
  const combined = isAbsolute ? relPath : `${baseDir ? `${baseDir}/` : ""}${relPath}`;
  const normalized = [];
  for (const segment of combined.split("/")) {
    if (!segment || segment === ".")
      continue;
    if (segment === "..") {
      if (normalized.length > 0)
        normalized.pop();
      continue;
    }
    normalized.push(segment);
  }
  const workspacePath = normalized.join("/");
  return `${getWorkspaceRawUrl(workspacePath)}${query}${hash}`;
}
function getPreview(context) {
  return context?.preview || null;
}
function fileExtensionFromPath(filePath) {
  const value = String(filePath || "");
  const lastSlash = Math.max(value.lastIndexOf("/"), value.lastIndexOf("\\"));
  const base = lastSlash >= 0 ? value.slice(lastSlash + 1) : value;
  const lastDot = base.lastIndexOf(".");
  if (lastDot <= 0 || lastDot === base.length - 1)
    return "none";
  return base.slice(lastDot + 1);
}
function previewKindLabel(preview) {
  if (!preview)
    return "unknown";
  if (preview.kind === "image")
    return "image";
  if (preview.kind === "text")
    return preview.content_type === "text/markdown" ? "markdown" : "text";
  if (preview.kind === "binary")
    return "binary";
  return String(preview.kind || "unknown");
}
function renderPreviewMetadata(context, preview) {
  const filePath = preview?.path || context?.path || "";
  const parts = [];
  if (preview?.content_type) {
    parts.push(`<span><strong>type:</strong> ${escapeHtml2(preview.content_type)}</span>`);
  }
  if (typeof preview?.size === "number") {
    parts.push(`<span><strong>size:</strong> ${escapeHtml2(formatFileSize(preview.size))}</span>`);
  }
  if (preview?.mtime) {
    parts.push(`<span><strong>modified:</strong> ${escapeHtml2(formatTimestamp(preview.mtime))}</span>`);
  }
  parts.push(`<span><strong>kind:</strong> ${escapeHtml2(previewKindLabel(preview))}</span>`);
  parts.push(`<span><strong>extension:</strong> ${escapeHtml2(fileExtensionFromPath(filePath))}</span>`);
  if (filePath) {
    parts.push(`<span><strong>path:</strong> ${escapeHtml2(filePath)}</span>`);
  }
  if (preview?.truncated) {
    parts.push("<span><strong>content:</strong> truncated</span>");
  }
  return `<div class="workspace-preview-meta workspace-preview-meta-inline">${parts.join("")}</div>`;
}
function renderWorkspacePreviewMarkup(context) {
  const preview = getPreview(context);
  if (!preview) {
    return '<div class="workspace-preview-text">No preview available.</div>';
  }
  const metadata = renderPreviewMetadata(context, preview);
  if (preview.kind === "image") {
    const src = preview.url || (preview.path ? getWorkspaceRawUrl(preview.path) : "");
    return `${metadata}
            <div class="workspace-preview-image">
                <img src="${escapeHtml2(src)}" alt="preview" />
            </div>
        `;
  }
  if (preview.kind === "text") {
    if (preview.content_type === "text/markdown") {
      const rendered = renderMarkdown(preview.text || "", null, {
        rewriteImageSrc: (src) => rewriteMarkdownImagePath(src, preview.path || context?.path)
      });
      return `${metadata}<div class="workspace-preview-text">${rendered}</div>`;
    }
    return `${metadata}<pre class="workspace-preview-text"><code>${escapeHtml2(preview.text || "")}</code></pre>`;
  }
  if (preview.kind === "binary") {
    return `${metadata}<div class="workspace-preview-text">Binary file — download to view.</div>`;
  }
  return `${metadata}<div class="workspace-preview-text">No preview available.</div>`;
}

class WorkspacePreviewInstance {
  constructor(container, context) {
    this.container = container;
    this.context = context;
    this.disposed = false;
    this.host = document.createElement("div");
    this.host.className = "workspace-preview-render-host";
    this.host.tabIndex = 0;
    this.container.appendChild(this.host);
    this.render();
  }
  render() {
    if (this.disposed)
      return;
    this.host.innerHTML = renderWorkspacePreviewMarkup(this.context);
  }
  getContent() {
    const preview = getPreview(this.context);
    return typeof preview?.text === "string" ? preview.text : undefined;
  }
  isDirty() {
    return false;
  }
  setContent(content, mtime) {
    const preview = getPreview(this.context);
    if (preview && preview.kind === "text") {
      preview.text = content;
      if (mtime !== undefined)
        preview.mtime = mtime;
    }
    this.context.content = content;
    if (mtime !== undefined)
      this.context.mtime = mtime;
    this.render();
  }
  focus() {
    this.host?.focus?.();
  }
  dispose() {
    if (this.disposed)
      return;
    this.disposed = true;
    this.host?.remove();
    this.container.innerHTML = "";
  }
}
var workspaceMarkdownPreviewPaneExtension = {
  id: "workspace-markdown-preview",
  label: "Workspace Markdown Preview",
  icon: "preview",
  capabilities: ["preview", "readonly"],
  placement: "tabs",
  canHandle(context) {
    const preview = getPreview(context);
    if (context?.mode !== "view")
      return false;
    if (!preview || preview.kind !== "text")
      return false;
    return preview.content_type === "text/markdown" ? 20 : false;
  },
  mount(container, context) {
    return new WorkspacePreviewInstance(container, context);
  }
};
var workspacePreviewPaneExtension = {
  id: "workspace-preview-default",
  label: "Workspace Preview",
  icon: "preview",
  capabilities: ["preview", "readonly"],
  placement: "tabs",
  canHandle(context) {
    if (context?.mode !== "view")
      return false;
    return getPreview(context) || context?.path ? 1 : false;
  },
  mount(container, context) {
    return new WorkspacePreviewInstance(container, context);
  }
};
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
      } catch (err) {
        console.warn("[tab-store] Change listener failed:", err);
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
    this.mruOrder = [id, ...this.mruOrder.filter((x) => x !== id)];
    this.notify();
  }
  close(id) {
    const tab = this.tabs.get(id);
    if (!tab)
      return false;
    this.tabs.delete(id);
    this.mruOrder = this.mruOrder.filter((x) => x !== id);
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
        this.mruOrder = this.mruOrder.filter((x) => x !== id);
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
        this.mruOrder = this.mruOrder.filter((x) => x !== id);
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
    this.mruOrder = this.mruOrder.map((x) => x === oldId ? newPath : x);
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
    return Array.from(this.tabs.values()).filter((t) => t.dirty);
  }
  nextTab() {
    const tabs = this.getTabs();
    if (tabs.length <= 1)
      return;
    const idx = tabs.findIndex((t) => t.id === this.activeId);
    const next = tabs[(idx + 1) % tabs.length];
    this.activate(next.id);
  }
  prevTab() {
    const tabs = this.getTabs();
    if (tabs.length <= 1)
      return;
    const idx = tabs.findIndex((t) => t.id === this.activeId);
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
  for (let i = 0;i < str.length; i += 1) {
    hash = (hash * 31 + str.charCodeAt(i)) % 2147483647;
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
  const r = Number(match[1]);
  const g = Number(match[2]);
  const b = Number(match[3]);
  if (![r, g, b].every((value) => Number.isFinite(value)))
    return null;
  return { r, g, b };
}
function parseColor2(input) {
  return parseHexColor2(input) || parseRgbColor(input);
}
function relativeLuminance2(color) {
  const toLinear = (channel) => {
    const value = channel / 255;
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  };
  const r = toLinear(color.r);
  const g = toLinear(color.g);
  const b = toLinear(color.b);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}
function contrastRatio(a, b) {
  const lighter = Math.max(relativeLuminance2(a), relativeLuminance2(b));
  const darker = Math.min(relativeLuminance2(a), relativeLuminance2(b));
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
  const [host, setHost] = F_(null);
  K_(() => {
    if (typeof document === "undefined")
      return;
    const nextHost = document.createElement("div");
    if (className)
      nextHost.className = className;
    document.body.appendChild(nextHost);
    setHost(nextHost);
    return () => {
      try {
        G_(null, nextHost);
      } finally {
        nextHost.remove();
        setHost((current) => current === nextHost ? null : current);
      }
    };
  }, [className]);
  W_(() => {
    if (!host)
      return;
    G_(children, host);
    return;
  }, [children, host]);
  return null;
}

// web/src/components/image-modal.ts
function ImageModal({ src, onClose }) {
  K_(() => {
    const handleEsc = (e) => {
      if (e.key === "Escape")
        onClose();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [onClose]);
  return fe`
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
  const iconSvg = icon === "message" ? fe`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>` : fe`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
      </svg>`;
  return fe`
    <span class=${pillClass} title=${title || label} onClick=${onClick}>
      ${iconSvg}
      <span class=${nameClass}>${label}</span>
      ${onRemove && fe`
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

// web/src/gi-clipboard-safety.ts
async function writeClipboardTextBestEffort(clipboard, value) {
  try {
    if (typeof clipboard?.writeText !== "function")
      return false;
    await clipboard.writeText(value);
    return true;
  } catch {
    return false;
  }
}

// web/src/components/post.ts
function FileAttachment({ mediaId, onPreview }) {
  const [info, setInfo] = F_(null);
  K_(() => {
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
  return fe`
        <div class="file-attachment" onClick=${(e) => e.stopPropagation()}>
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
                        ${sizeStr && fe`<span class="file-size">${sizeStr}</span>`}
                        ${info.content_type && fe`<span class="file-size">${info.content_type}</span>`}
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
                onClick=${(e) => {
    e.preventDefault();
    e.stopPropagation();
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
  const [info, setInfo] = F_(null);
  K_(() => {
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
  return fe`
        <span class="attachment-pill" title=${filename}>
            ${downloadHref ? fe`
                    <a href=${downloadHref} download=${filename} class="attachment-pill-main" onClick=${(e) => e.stopPropagation()}>
                        <${FilePill}
                            prefix="post"
                            label=${attachment.label}
                            title=${filename}
                        />
                    </a>
                ` : fe`
                    <${FilePill}
                        prefix="post"
                        label=${attachment.label}
                        title=${filename}
                    />
                `}
            ${Number.isFinite(mediaId) && info && fe`
                <button
                    class="attachment-pill-preview"
                    type="button"
                    title=${previewLabel}
                    onClick=${(e) => {
    e.preventDefault();
    e.stopPropagation();
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
  return fe`
        <div class="content-annotations">
            ${audience && audience.length > 0 && fe`
                <span class="content-annotation">Audience: ${audience.join(", ")}</span>
            `}
            ${typeof priority === "number" && fe`
                <span class="content-annotation">Priority: ${priority}</span>
            `}
            ${formattedLastModified && fe`
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
  return fe`
        <a
            href=${safeUrl || "#"}
            class="resource-link"
            target=${safeUrl ? "_blank" : undefined}
            rel=${safeUrl ? "noopener noreferrer" : undefined}
            onClick=${(e) => e.stopPropagation()}>
            <div class="resource-link-main">
                <div class="resource-link-header">
                    <span class="resource-link-icon-inline">${icon}</span>
                    <div class="resource-link-title">${name}</div>
                </div>
                ${description && fe`<div class="resource-link-description">${description}</div>`}
                <div class="resource-link-meta">
                    ${mimeType && fe`<span>${mimeType}</span>`}
                    ${sizeStr && fe`<span>${sizeStr}</span>`}
                </div>
            </div>
            <div class="resource-link-icon">↗</div>
        </a>
    `;
}
function ResourceBlock({ block }) {
  const [open, setOpen] = F_(false);
  const title = block.uri || "Embedded resource";
  const contentText = block.text || "";
  const hasBlob = Boolean(block.data);
  const mimeType = block.mime_type || "";
  return fe`
        <div class="resource-embed">
            <button class="resource-embed-toggle" onClick=${(e) => {
    e.preventDefault();
    e.stopPropagation();
    setOpen(!open);
  }}>
                ${open ? "▼" : "▶"} ${title}
            </button>
            ${open && fe`
                ${contentText && fe`<pre class="resource-embed-content">${contentText}</pre>`}
                ${hasBlob && fe`
                    <div class="resource-embed-blob">
                        <span class="resource-embed-blob-label">Embedded blob</span>
                        ${mimeType && fe`<span class="resource-embed-blob-meta">${mimeType}</span>`}
                        <button class="resource-embed-blob-btn" onClick=${(e) => {
    e.preventDefault();
    e.stopPropagation();
    const blob = new Blob([Uint8Array.from(atob(block.data), (c) => c.charCodeAt(0))], { type: mimeType || "application/octet-stream" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = title.split("/").pop() || "resource";
    a.click();
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
  const autoOpened = Q_(false);
  const launchWidget = (e) => {
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    if (!payload)
      return;
    onOpenWidget?.(payload);
  };
  K_(() => {
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
  return fe`
        <div class="generated-widget-launch" onClick=${(e) => e.stopPropagation()}>
            <div class="generated-widget-launch-header">
                <div class="generated-widget-launch-eyebrow">Generated widget${kind ? ` • ${String(kind).toUpperCase()}` : ""}</div>
                <div class="generated-widget-launch-title">${title}</div>
            </div>
            ${description && fe`<div class="generated-widget-launch-description">${description}</div>`}
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
  return fe`
        <a
            href=${safeUrl || "#"}
            class="link-preview ${bgStyle ? "has-image" : ""}"
            target=${safeUrl ? "_blank" : undefined}
            rel=${safeUrl ? "noopener noreferrer" : undefined}
            onClick=${(e) => e.stopPropagation()}
            style=${bgStyle}>
            <div class="link-preview-overlay">
                <div class="link-preview-site">${siteName || ""}</div>
                <div class="link-preview-title">${preview.title}</div>
                ${preview.description && fe`
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
  for (let i = 0;i < lines.length; i += 1) {
    if (lines[i].trim() === "Files:" && lines[i + 1] && /^\s*-\s+/.test(lines[i + 1])) {
      start = i;
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
  for (let i = 0;i < lines.length; i += 1) {
    if (lines[i].trim() === "Referenced messages:" && lines[i + 1] && /^\s*-\s+/.test(lines[i + 1])) {
      start = i;
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
  for (let i = 0;i < lines.length; i += 1) {
    const header = lines[i].trim();
    if ((header === "Images:" || header === "Attachments:") && lines[i + 1] && /^\s*-\s+/.test(lines[i + 1])) {
      start = i;
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
  const escapedTerms = terms.map(escapeRegex).sort((a, b) => b.length - a.length);
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
  const [zoomedImage, setZoomedImage] = F_(null);
  const [copyState, setCopyState] = F_("idle");
  const contentRef = Q_(null);
  const copyResetTimerRef = Q_(null);
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
  const renderedHtml = u_(() => {
    if (!displayContent || hideRenderedFallback)
      return "";
    const baseHtml = renderMarkdown(displayContent, onHashtagClick);
    return highlightQueryText ? highlightHtml(baseHtml, highlightQueryText) : baseHtml;
  }, [displayContent, hideRenderedFallback, highlightQueryText]);
  const markdownCopyPayload = u_(() => buildPostMarkdownCopyPayload(post), [post]);
  const handleImageClick = (e, mediaId) => {
    e.stopPropagation();
    setZoomedImage(getMediaUrl(mediaId));
  };
  const handleAttachmentPreview = (attachment) => {
    onOpenAttachmentPreview?.(attachment);
  };
  const handleDeleteClick = (e) => {
    e.stopPropagation();
    onDelete?.(post);
  };
  const handleCopyMarkdownClick = async (e) => {
    e.stopPropagation();
    const ok = await copyMessageToClipboard(markdownCopyPayload);
    setCopyState(ok ? "success" : "error");
    if (copyResetTimerRef.current)
      clearTimeout(copyResetTimerRef.current);
    copyResetTimerRef.current = setTimeout(() => {
      copyResetTimerRef.current = null;
      setCopyState("idle");
    }, CODE_COPY_RESET_MS);
  };
  const resolveInlineAttachments = (content, attachments) => {
    const usedIds = new Set;
    if (!content || attachments.length === 0) {
      return { content, usedIds };
    }
    const replaced = content.replace(/attachment:([^\s)"']+)/g, (match, rawRef, offset, source) => {
      const ref = rawRef.replace(/^\/+/, "");
      const byName = attachments.find((entry) => entry.name && entry.name.toLowerCase() === ref.toLowerCase() && !usedIds.has(entry.id));
      const entry = byName || attachments.find((item) => !usedIds.has(item.id));
      if (!entry)
        return match;
      usedIds.add(entry.id);
      const prefix = source.slice(Math.max(0, offset - 2), offset);
      if (prefix === "](") {
        return `/media/${entry.id}`;
      }
      return entry.name || "attachment";
    });
    return { content: replaced, usedIds };
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
  const cardBlocks = u_(() => extractCardBlocks(blocks), [blocks]);
  const cardSubmissionBlocks = u_(() => extractAdaptiveCardSubmissionBlocks(blocks), [blocks]);
  const cardBlocksKey = u_(() => {
    return cardBlocks.map((b) => `${b.card_id}:${b.state}`).join("|");
  }, [cardBlocks]);
  K_(() => {
    if (!contentRef.current)
      return;
    renderMermaidDiagrams(contentRef.current);
    return enhanceCodeBlocks(contentRef.current);
  }, [renderedHtml]);
  K_(() => () => {
    if (copyResetTimerRef.current)
      clearTimeout(copyResetTimerRef.current);
  }, []);
  const cardContainerRef = Q_(null);
  K_(() => {
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
  return fe`
        <div id=${`post-${post.id}`} class="post ${isAgent ? "agent-post" : ""} ${isThreadReply ? "thread-reply" : ""} ${isThreadPrev ? "thread-prev" : ""} ${isThreadNext ? "thread-next" : ""} ${isRemoving ? "removing" : ""}" onClick=${onClick}>
            <div class="post-avatar ${isAgent ? "agent-avatar" : ""} ${avatarInfo.image ? "has-image" : ""}" style=${avatarStyle}>
                ${avatarInfo.image ? fe`<img src=${avatarInfo.image} alt=${displayName} />` : avatarInfo.letter}
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
                        ${copyState === "success" ? fe`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><path d="M20 6L9 17l-5-5"></path></svg>` : copyState === "error" ? fe`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><circle cx="12" cy="12" r="9"></circle><path d="M9 9l6 6M15 9l-6 6"></path></svg>` : fe`<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false"><rect x="9" y="9" width="10" height="10" rx="2"></rect><path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path></svg>`}
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
                    ${showSearchChatAgentTag && fe`<span class="post-chat-agent-tag" title=${`Chat: ${searchChatAgentName}`}>@${searchChatAgentName}</span>`}
                    ${recoveryMarker && fe`
                        <span
                            class="post-recovery-chip"
                            title=${formatRecoveryChipTooltip(recoveryMarker)}
                        >
                            recovered
                        </span>
                    `}
                    ${timeoutMarker && fe`
                        <span
                            class="post-recovery-chip post-timeout-chip"
                            title=${formatTimeoutChipTooltip(timeoutMarker)}
                        >
                            timeout
                        </span>
                    `}
                    <a class="post-time" href=${`#msg-${post.id}`} onClick=${(e) => {
    e.preventDefault();
    e.stopPropagation();
    if (onMessageRef)
      onMessageRef(post.id);
  }}>${formatTime(post.timestamp)}</a>
                </div>
                ${isHardTruncated && truncatedInfo && fe`
                    <div class="post-content truncated">
                        <div class="truncated-title">Message too large to display.</div>
                        <div class="truncated-meta">
                            Original length: ${formatCount(truncatedInfo.originalLength)} chars
                            ${truncatedInfo.maxLength ? fe` • Display limit: ${formatCount(truncatedInfo.maxLength)} chars` : ""}
                        </div>
                    </div>
                `}
                ${isPreview && truncatedInfo && fe`
                    <div class="post-content preview">
                        <div class="truncated-title">Preview truncated.</div>
                        <div class="truncated-meta">
                            Showing first ${formatCount(truncatedInfo.maxLength)} of ${formatCount(truncatedInfo.originalLength)} chars. Download full text below.
                        </div>
                    </div>
                `}
                ${(fileRefs.length > 0 || messageRefs.length > 0 || attachmentPills.length > 0) && fe`
                    <div class="post-file-refs">
                        ${messageRefs.map((id) => {
    const scrollToRef = (e) => {
      e.preventDefault();
      e.stopPropagation();
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
    return fe`
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
    return fe`
                                <${FilePill}
                                    prefix="post"
                                    label=${label}
                                    title=${ref}
                                    onClick=${() => onFileRef?.(ref)}
                                />
                            `;
  })}
                        ${attachmentPills.map((attachment) => fe`
                            <${AttachmentPill}
                                key=${attachment.id}
                                attachment=${attachment}
                                onPreview=${handleAttachmentPreview}
                            />
                        `)}
                    </div>
                `}
                ${shouldRenderContent && fe`
                    <div 
                        ref=${contentRef}
                        class="post-content"
                        dangerouslySetInnerHTML=${{ __html: renderedHtml }}
                        onClick=${(e) => {
    if (e.target.classList.contains("hashtag")) {
      e.preventDefault();
      e.stopPropagation();
      const tag = e.target.dataset.hashtag;
      if (tag)
        onHashtagClick?.(tag);
    } else if (e.target.tagName === "IMG") {
      e.preventDefault();
      e.stopPropagation();
      setZoomedImage(e.target.src);
    }
  }}
                    />
                `}
                ${cardBlocks.length > 0 && fe`
                    <div ref=${cardContainerRef} class="post-adaptive-cards" />
                `}
                ${cardSubmissionBlocks.length > 0 && fe`
                    <div class="post-adaptive-card-submissions">
                        ${cardSubmissionBlocks.map((block, idx) => {
    const meta = describeAdaptiveCardSubmission(block);
    const submissionKey = `${block.card_id}-${idx}`;
    return fe`
                                <div key=${submissionKey} class="adaptive-card-submission-receipt">
                                    <div class="adaptive-card-submission-header">
                                        <span class="adaptive-card-submission-icon" aria-hidden="true">✓</span>
                                        <div class="adaptive-card-submission-title-wrap">
                                            <span class="adaptive-card-submission-title">Submitted</span>
                                            <span class="adaptive-card-submission-title-action">${meta.title}</span>
                                        </div>
                                    </div>
                                    ${meta.fields.length > 0 && fe`
                                        <div class="adaptive-card-submission-fields">
                                            ${meta.fields.map((field) => fe`
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
                ${generatedWidgets.length > 0 && fe`
                    <div class="generated-widget-launches">
                        ${generatedWidgets.map((block, idx) => fe`
                            <${GeneratedWidgetLaunch}
                                key=${block.widget_id || block.id || `${post.id}-widget-${idx}`}
                                block=${block}
                                post=${post}
                                onOpenWidget=${onOpenWidget}
                            />
                        `)}
                    </div>
                `}
                ${textAnnotations.length > 0 && fe`
                    ${textAnnotations.map((annotations, idx) => fe`
                        <${AnnotationsBadge} key=${idx} annotations=${annotations} />
                    `)}
                `}
                ${filteredImageItems.length > 0 && fe`
                    <div class="media-preview">
                        ${filteredImageItems.map(({ id, mimeType }) => {
    const isSvg = typeof mimeType === "string" && mimeType.toLowerCase().startsWith("image/svg");
    const imageSrc = isSvg ? getMediaUrl(id) : getThumbnailUrl(id);
    return fe`
                                <img 
                                    key=${id} 
                                    src=${imageSrc} 
                                    alt="Media" 
                                    loading="lazy"
                                    decoding="async"
                                    onClick=${(e) => handleImageClick(e, id)}
                                />
                            `;
  })}
                    </div>
                `}
                ${filteredImageItems.length > 0 && fe`
                    ${filteredImageItems.map(({ annotations }, idx) => fe`
                        ${annotations && fe`<${AnnotationsBadge} key=${idx} annotations=${annotations} />`}
                    `)}
                `}
                ${filteredFileIds.length > 0 && fe`
                    <div class="file-attachments">
                        ${filteredFileIds.map((id) => fe`
                            <${FileAttachment} key=${id} mediaId=${id} onPreview=${handleAttachmentPreview} />
                        `)}
                    </div>
                `}
                ${resourceLinks.length > 0 && fe`
                    <div class="resource-links">
                        ${resourceLinks.map((block, idx) => fe`
                            <div key=${idx}>
                                <${ResourceLinkBlock} block=${block} />
                                <${AnnotationsBadge} annotations=${block.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${resources.length > 0 && fe`
                    <div class="resource-embeds">
                        ${resources.map((block, idx) => fe`
                            <div key=${idx}>
                                <${ResourceBlock} block=${block} />
                                <${AnnotationsBadge} annotations=${block.annotations} />
                            </div>
                        `)}
                    </div>
                `}
                ${data.link_previews?.length > 0 && fe`
                    <div class="link-previews">
                        ${data.link_previews.map((preview, i) => fe`
                            <${LinkPreview} key=${i} preview=${preview} />
                        `)}
                    </div>
                `}
            </div>
        </div>
        ${zoomedImage && fe`<${ImageModal} src=${zoomedImage} onClose=${() => setZoomedImage(null)} />`}

    `;
}

// web/src/components/timeline.ts
function Timeline({ posts, hasMore, onLoadMore, onPostClick, onHashtagClick, onMessageRef, onScrollToMessage, onFileRef, onOpenWidget, onOpenAttachmentPreview, emptyMessage, timelineRef, agents, user, onDeletePost, reverse = true, removingPostIds, searchQuery }) {
  const [loadingMore, setLoadingMore] = F_(false);
  const sentinelRef = Q_(null);
  const hasIntersectionObserver = typeof IntersectionObserver !== "undefined";
  const triggerLoadMore = Y_(async () => {
    if (!onLoadMore || !hasMore || loadingMore)
      return;
    setLoadingMore(true);
    try {
      await onLoadMore({ preserveScroll: true, preserveMode: "top" });
    } finally {
      setLoadingMore(false);
    }
  }, [hasMore, loadingMore, onLoadMore]);
  const handleScroll = Y_((e) => {
    const { scrollTop, scrollHeight, clientHeight } = e.target;
    const distanceFromTop = reverse ? scrollHeight - clientHeight - scrollTop : scrollTop;
    const prefetchThreshold = Math.max(300, clientHeight);
    if (distanceFromTop < prefetchThreshold) {
      triggerLoadMore();
    }
  }, [reverse, triggerLoadMore]);
  K_(() => {
    if (!hasIntersectionObserver)
      return;
    const sentinel = sentinelRef.current;
    const root = timelineRef?.current;
    if (!sentinel || !root)
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
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasIntersectionObserver, hasMore, onLoadMore, timelineRef, triggerLoadMore]);
  const triggerLoadMoreRef = Q_(triggerLoadMore);
  triggerLoadMoreRef.current = triggerLoadMore;
  K_(() => {
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
  K_(() => {
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
    return fe`<div class="loading"><div class="spinner"></div></div>`;
  }
  if (posts.length === 0) {
    return fe`
            <div class="timeline" ref=${timelineRef}>
                <div class="timeline-content">
                    <div style="padding: var(--spacing-xl); text-align: center; color: var(--text-secondary)">
                        ${emptyMessage || "No messages yet. Start a conversation!"}
                    </div>
                </div>
            </div>
        `;
  }
  const displayPosts = posts.slice().sort((a, b) => a.id - b.id);
  const resolveThreadRootId = (post) => {
    const raw = post?.data?.thread_id;
    if (raw === null || raw === undefined || raw === "")
      return null;
    const threadId = Number(raw);
    return Number.isFinite(threadId) ? threadId : null;
  };
  const threadGroups = new Map;
  for (let i = 0;i < displayPosts.length; i += 1) {
    const post = displayPosts[i];
    const postId = Number(post?.id);
    const threadRootId = resolveThreadRootId(post);
    if (threadRootId !== null) {
      const group = threadGroups.get(threadRootId) || { anchorIndex: -1, replyIndexes: [] };
      group.replyIndexes.push(i);
      threadGroups.set(threadRootId, group);
    } else if (Number.isFinite(postId)) {
      const group = threadGroups.get(postId) || { anchorIndex: -1, replyIndexes: [] };
      group.anchorIndex = i;
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
    threadSequences.set(threadId, Array.from(ordered).sort((a, b) => a - b));
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
  const sentinel = fe`<div class="timeline-sentinel" ref=${sentinelRef}></div>`;
  return fe`
        <div class="timeline ${reverse ? "reverse" : "normal"}" ref=${timelineRef} onScroll=${handleScroll}>
            <div class="timeline-content">
                ${reverse ? sentinel : null}
                ${displayPosts.map((post, index) => {
    const isThreadReply = Boolean(post.data?.thread_id && post.data.thread_id !== post.id);
    const isRemoving = removingPostIds?.has?.(post.id);
    const threadInfo = threadInfoByIndex[index] || {};
    return fe`
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

// web/src/ui/compose-session-switcher.ts
var SECTION_LABELS = {
  current: "Current",
  pinned: "Pinned",
  active: "Active",
  tree: "This session tree",
  other: "Other sessions",
  archived: "Archived"
};
function clean(value) {
  return typeof value === "string" ? value.trim() : "";
}
function buildSessionPickerSearchDocument(chat) {
  const archived = Boolean(chat?.archived_at);
  const active = Boolean(chat?.is_active) && !archived;
  return [
    clean(chat?.agent_name) ? `@${clean(chat.agent_name)}` : "",
    clean(chat?.agent_name),
    clean(chat?.chat_jid),
    clean(chat?.root_chat_jid),
    clean(chat?.model),
    clean(chat?.model_label),
    clean(chat?.provider),
    archived ? "archived" : active ? "active" : "idle",
    clean(chat?.parent_branch_id),
    clean(chat?.branch_id)
  ].filter(Boolean).join(" ").toLocaleLowerCase();
}
function matchesSessionPickerSearch(chat, query) {
  const terms = clean(query).toLocaleLowerCase().split(/\s+/).filter(Boolean);
  if (terms.length === 0)
    return true;
  const document2 = buildSessionPickerSearchDocument(chat);
  return terms.every((term) => document2.includes(term));
}
function getSessionPickerHandleMatchRank(chat, query) {
  if (!matchesSessionPickerSearch(chat, query))
    return 4;
  const handle = clean(chat?.agent_name).toLocaleLowerCase();
  const terms = clean(query).toLocaleLowerCase().split(/\s+/).map((term) => term.replace(/^@/, "")).filter(Boolean);
  if (!handle || terms.length === 0)
    return 3;
  if (terms.some((term) => handle === term))
    return 0;
  if (terms.some((term) => handle.startsWith(term)))
    return 1;
  if (terms.some((term) => handle.includes(term)))
    return 2;
  return 3;
}
function resolveSessionPickerSearchInitialIndex(chats, query) {
  let bestIndex = 0;
  let bestRank = Number.POSITIVE_INFINITY;
  for (let index = 0;index < chats.length; index += 1) {
    const rank = getSessionPickerHandleMatchRank(chats[index], query);
    if (rank >= bestRank)
      continue;
    bestIndex = index;
    bestRank = rank;
    if (rank === 0)
      break;
  }
  return bestIndex;
}
function filterSessionPickerChats(chats, query) {
  const normalized = clean(query).toLocaleLowerCase();
  if (!normalized)
    return chats;
  const matches = new Set(chats.filter((chat) => matchesSessionPickerSearch(chat, normalized)).map((chat) => clean(chat.chat_jid)));
  const byBranchId = new Map(chats.map((chat) => [clean(chat.branch_id), chat]).filter(([id]) => Boolean(id)));
  for (const chat of chats) {
    if (!matches.has(clean(chat.chat_jid)))
      continue;
    let parentId = clean(chat.parent_branch_id);
    while (parentId) {
      const parent = byBranchId.get(parentId);
      if (!parent)
        break;
      matches.add(clean(parent.chat_jid));
      parentId = clean(parent.parent_branch_id);
    }
  }
  return chats.filter((chat) => matches.has(clean(chat.chat_jid)));
}
function groupSessionPickerChats(chats, currentChatJid, pinnedChatJids = []) {
  const current = clean(currentChatJid);
  const currentChat = chats.find((chat) => clean(chat.chat_jid) === current);
  const currentRoot = clean(currentChat?.root_chat_jid) || current;
  const pinned = new Set(Array.from(pinnedChatJids, clean).filter(Boolean));
  const buckets = new Map([
    ["current", []],
    ["pinned", []],
    ["active", []],
    ["tree", []],
    ["other", []],
    ["archived", []]
  ]);
  for (const chat of chats) {
    const jid = clean(chat.chat_jid);
    const archived = Boolean(chat.archived_at);
    const section = archived ? "archived" : jid === current ? "current" : pinned.has(jid) ? "pinned" : Boolean(chat.is_active) ? "active" : (clean(chat.root_chat_jid) || jid) === currentRoot ? "tree" : "other";
    buckets.get(section).push(chat);
  }
  return ["current", "pinned", "active", "tree", "other", "archived"].map((key) => ({ key, label: SECTION_LABELS[key], items: buckets.get(key) })).filter((section) => section.items.length > 0);
}
function moveSessionPickerIndex(current, length, key, pageSize = 8) {
  if (length <= 0)
    return 0;
  const index = Math.max(0, Math.min(current, length - 1));
  if (key === "Home")
    return 0;
  if (key === "End")
    return length - 1;
  if (key === "ArrowDown")
    return (index + 1) % length;
  if (key === "ArrowUp")
    return (index - 1 + length) % length;
  if (key === "PageDown")
    return Math.min(length - 1, index + pageSize);
  if (key === "PageUp")
    return Math.max(0, index - pageSize);
  return index;
}
function canUseComposeSessionSwitcher(options = {}) {
  if (options.searchMode)
    return false;
  return Boolean(options.showSessionSwitcherButton);
}
function shouldOpenSessionSwitcherFromBlankCompose(event, value, options = {}) {
  if (!event || event.isComposing)
    return false;
  if (!canUseComposeSessionSwitcher(options))
    return false;
  if (event.ctrlKey || event.metaKey || event.altKey)
    return false;
  if (event.key !== "@")
    return false;
  return String(value || "") === "";
}

// web/src/ui/popup-typeahead.ts
var POPUP_TYPEAHEAD_RESET_MS = 700;
function normalize2(value) {
  return String(value || "").toLowerCase().replace(/^@/, "").replace(/\s+/g, " ").trim();
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
  for (let i = 0;i < size; i += 1) {
    out.push((normalizedStart + i) % size);
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

// web/src/gi-session-typeahead.ts
function sessionTypeahead(event, entries, previous) {
  if (event.defaultPrevented || event.repeat || !isPopupTypeaheadKey(event) || event.target?.closest?.('input, textarea, select, [contenteditable="true"]'))
    return null;
  const buffer = updatePopupTypeaheadBuffer(previous, event.key);
  const enabled = entries.map((entry, index) => ({ entry, index })).filter((item) => !item.entry.disabled);
  const match = findPopupTypeaheadMatch(enabled, buffer.value, 0, (item) => item.entry.label);
  return { buffer, index: match < 0 ? -1 : enabled[match].index };
}

// web/src/gi-model-picker.ts
function filterModelOptions(options, query, label) {
  const terms = query.toLowerCase().trim().split(/\s+/).filter(Boolean);
  return options.filter((option) => terms.every((term) => label(option).toLowerCase().includes(term)));
}
function modelPickerKey(event, entries, current, previous) {
  if (event.defaultPrevented || event.isComposing || event.ctrlKey || event.metaKey || event.altKey)
    return null;
  const target = event.target;
  const editing = Boolean(target?.closest?.('input, textarea, select, [contenteditable="true"]'));
  const nativeButton = target?.closest?.("button");
  const focused = target?.closest?.("[data-model-index]")?.getAttribute("data-model-index");
  const index = focused == null ? current : Number(focused);
  const enabled = entries.map((entry, index) => ({ entry, index })).filter((item) => !item.entry.disabled);
  const buffer = { value: "", updatedAt: 0 };
  if (["ArrowDown", "ArrowUp", "PageDown", "PageUp"].includes(event.key) || !editing && ["Home", "End"].includes(event.key)) {
    const selected = enabled.findIndex((item) => item.index === index);
    return { index: enabled[moveSessionPickerIndex(selected, enabled.length, event.key)]?.index ?? -1, buffer, activate: false, focus: !editing };
  }
  if (event.key === "Enter") {
    if (nativeButton && !event.repeat)
      return null;
    return { index, buffer, activate: !event.repeat && Boolean(entries[index] && !entries[index].disabled), focus: false };
  }
  const typed = sessionTypeahead(event, entries, previous);
  return typed ? { ...typed, activate: false, focus: true } : null;
}

// web/src/gi-quick-actions.ts
function settingsOwnsKeyboard(doc = document) {
  return Boolean(doc.querySelector?.('.settings-dialog[aria-modal="true"]'));
}
function blocksQuickActions(event, ready) {
  const target = event.target;
  return !ready || event.defaultPrevented || event.repeat || Boolean(target?.closest?.('button, a, [role="button"], [role="menuitem"], .monaco-editor, .terminal-pane, .post-reply'));
}

// web/src/gi-drafts.ts
var emptyDraft = () => ({ text: "", media: [], fileRefs: [], messageRefs: [] });
var copy = (d) => ({ text: d.text, media: [...d.media], fileRefs: [...d.fileRefs], messageRefs: [...d.messageRefs] });
var key = (value) => typeof value === "object" ? JSON.stringify(value) : String(value);
var unique = (values, identity = key) => [...new Map(values.map((value) => [identity(value), value])).values()];
function mergeDrafts(captured, current) {
  const text = !captured.text || current.text === captured.text || current.text.startsWith(captured.text + `
`) ? current.text : [captured.text, current.text].filter(Boolean).join(`

`);
  return {
    text,
    media: unique([...captured.media, ...current.media], (f) => `${f.name}:${f.size}:${f.type}:${f.lastModified}`),
    fileRefs: unique([...captured.fileRefs, ...current.fileRefs]),
    messageRefs: unique([...captured.messageRefs, ...current.messageRefs])
  };
}
function indexedDraftStorage(factory = indexedDB) {
  const encodedFiles = new WeakMap;
  const encodeFile = (file) => {
    if (!encodedFiles.has(file))
      encodedFiles.set(file, file.arrayBuffer().then((bytes) => ({ name: file.name, type: file.type, lastModified: file.lastModified, bytes })));
    return encodedFiles.get(file);
  };
  const database = new Promise((resolve, reject) => {
    const request = factory.open("gi-session-drafts", 1);
    request.onupgradeneeded = () => {
      if (!request.result.objectStoreNames.contains("drafts"))
        request.result.createObjectStore("drafts", { keyPath: "sessionId" });
    };
    request.onsuccess = () => {
      request.result.onversionchange = () => request.result.close();
      resolve(request.result);
    };
    request.onerror = () => reject(request.error || new Error("Draft database unavailable"));
    request.onblocked = () => reject(new Error("Draft database upgrade blocked by another tab"));
  });
  return {
    async load() {
      const db = await database;
      return new Promise((resolve, reject) => {
        const tx = db.transaction("drafts", "readonly");
        const request = tx.objectStore("drafts").getAll();
        tx.oncomplete = () => {
          const decode = (draft) => ({ ...draft, media: draft.media.map((file) => file instanceof File ? file : new File([file.bytes], file.name, { type: file.type, lastModified: file.lastModified })) });
          resolve(request.result.map((row) => ({ ...row, draft: decode(row.draft), pending: row.pending.map((p) => ({ ...p, draft: decode(p.draft) })) })));
        };
        tx.onabort = tx.onerror = () => reject(tx.error || new Error("Could not load drafts"));
      });
    },
    async put(record) {
      const db = await database;
      const encode = async (draft) => ({ ...draft, media: await Promise.all(draft.media.map(encodeFile)) });
      const stored = { ...record, draft: await encode(record.draft), pending: await Promise.all(record.pending.map(async (pending) => ({ ...pending, draft: await encode(pending.draft) }))) };
      return new Promise((resolve, reject) => {
        const tx = db.transaction("drafts", "readwrite");
        tx.objectStore("drafts").put(stored);
        tx.oncomplete = () => resolve();
        tx.onabort = tx.onerror = () => reject(tx.error || new Error("Could not save draft"));
      });
    }
  };
}
function createDraftRepository(storage, onError = () => {}) {
  const records = new Map;
  let tail = Promise.resolve();
  const record = (id) => {
    if (!records.has(id))
      records.set(id, { sessionId: id, draft: emptyDraft(), pending: [] });
    return records.get(id);
  };
  const persist = (id) => {
    const source = record(id);
    const snapshot = { ...source, queueReturns: Object.fromEntries(Object.entries(source.queueReturns || {}).map(([id, entry]) => [id, { ...entry }])), draft: copy(source.draft), pending: source.pending.map((p) => ({ id: p.id, draft: copy(p.draft) })) };
    const write = tail.catch(() => {}).then(() => storage.put(snapshot));
    tail = write;
    write.catch((error) => onError(error));
    return write;
  };
  return {
    async load() {
      const rows = await storage.load();
      for (const row of rows)
        records.set(row.sessionId, row);
      for (const row of rows) {
        records.set(row.sessionId, row);
        if (row.pending.length) {
          for (const pending of [...row.pending].reverse())
            row.draft = mergeDrafts(pending.draft, row.draft);
          row.pending = [];
          row.error = "Recovered an unacknowledged send. Delivery is unknown; check the timeline before resending.";
          await persist(row.sessionId);
        }
      }
    },
    get(id) {
      return record(id).draft;
    },
    error(id) {
      return record(id).error || "";
    },
    update(id, patch) {
      const draft = record(id).draft;
      if (Object.entries(patch).every(([field, value]) => draft[field] === value))
        return;
      Object.assign(draft, patch);
      persist(id).catch(() => {});
    },
    begin(id, draft) {
      const token = crypto.randomUUID();
      const row = record(id);
      row.pending.push({ id: token, draft: copy(draft) });
      row.draft = emptyDraft();
      row.error = "";
      return { token, ready: persist(id) };
    },
    async accepted(id, token) {
      const row = record(id);
      row.pending = row.pending.filter((p) => p.id !== token);
      await persist(id);
    },
    failed(id, token, error) {
      const row = record(id);
      const pending = row.pending.find((p) => p.id === token);
      if (pending)
        row.draft = mergeDrafts(pending.draft, row.draft);
      row.pending = row.pending.filter((p) => p.id !== token);
      row.error = error;
      persist(id).catch(() => {});
      return copy(row.draft);
    },
    hasQueueReturn(id, queueId) {
      return Boolean(record(id).queueReturns?.[queueId]);
    },
    prepareQueueReturn(id, queueId, captured) {
      const row = record(id);
      row.queueReturns ||= {};
      if (!row.queueReturns[queueId]) {
        const current = row.draft;
        row.draft = mergeDrafts(captured, current);
        row.draft.text = [captured.text, current.text].filter(Boolean).join(`

`);
        row.queueReturns[queueId] = { state: "prepared", recoveredAt: Date.now() };
      }
      return { draft: copy(row.draft), ready: persist(id) };
    },
    queueReturnFailed(id, queueId, message) {
      const row = record(id);
      if (row.queueReturns?.[queueId]) {
        row.error = `Queue return incomplete: ${message}. Recovered content is retained; check whether the original turn ran before sending it again.`;
        persist(id).catch(() => {});
      }
    },
    async completeQueueReturn(id, queueId) {
      const entry = record(id).queueReturns?.[queueId];
      if (entry)
        entry.state = "removed";
      if (record(id).error?.startsWith("Queue return incomplete:"))
        record(id).error = "";
      await persist(id);
    },
    async flushStable() {
      let pending;
      do {
        pending = tail;
        await pending;
      } while (pending !== tail);
    },
    flush() {
      return tail;
    }
  };
}

// web/src/gi-context-usage.ts
var known = (value) => typeof value === "number" && Number.isFinite(value) && value >= 0;
function formatContextCount(value) {
  if (!known(value))
    return "?";
  if (value >= 1e6)
    return `${(value / 1e6).toFixed(1)}M`;
  if (value >= 1000)
    return `${(value / 1000).toFixed(0)}K`;
  return String(value);
}
function contextPresentation(usage, canCompact = false) {
  const percent = known(usage?.percent) ? usage.percent : null;
  const fill = percent == null ? 0 : Math.min(100, percent);
  const label = `Context: ${formatContextCount(usage?.tokens)} / ${formatContextCount(usage?.contextWindow)} tokens (${percent == null ? "?" : percent.toFixed(0)}%)`;
  const qualifier = usage?.source === "provider_request" ? " — latest measured provider request" : " — usage unavailable";
  return {
    fill,
    label,
    title: label + qualifier + (canCompact ? " — Compact context" : ""),
    color: percent == null ? "var(--text-secondary)" : percent > 90 ? "var(--context-red, #ef4444)" : percent > 75 ? "var(--context-amber, #f59e0b)" : "var(--context-green, #22c55e)"
  };
}
function modelContextBlocked(option, usage) {
  return known(usage?.tokens) && known(option?.contextWindow) && option.contextWindow > 0 && usage.tokens > option.contextWindow;
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

// web/src/ui/branch-lifecycle.ts
function normalizeHandle(value) {
  const normalized = normalizeHandleName(value);
  return normalized ? `@${normalized}` : "";
}
function normalizeHandleName(value) {
  return String(value || "").trim().toLowerCase().replace(/[^a-z0-9_-]+/g, "-").replace(/^-+|-+$/g, "").replace(/-{2,}/g, "-");
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
  const [disconnectedAtMs, setDisconnectedAtMs] = F_(null);
  const [displayNowMs, setDisplayNowMs] = F_(() => Date.now());
  K_(() => {
    if (status === "disconnected") {
      const startedAt = Date.now();
      setDisconnectedAtMs((previous) => previous ?? startedAt);
      setDisplayNowMs(startedAt);
      return;
    }
    setDisconnectedAtMs(null);
    setDisplayNowMs(Date.now());
  }, [status]);
  K_(() => {
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
  return u_(() => resolveConnectionStatusPresentation(status, {
    delayMs,
    disconnectedAtMs,
    nowMs: displayNowMs
  }), [status, delayMs, disconnectedAtMs, displayNowMs]);
}

// web/src/components/compose-model-refresh.ts
async function refreshAgentModelStateBestEffort(getAgentModels, chatJid, emitModelState) {
  if (typeof getAgentModels !== "function")
    return false;
  try {
    const latest = await getAgentModels(chatJid);
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
  { name: "/route-events", description: "Toggle routing event visibility in the timeline (/route-events on|off|toggle)" },
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
  const { fill: pct, label, title, color } = contextPresentation(usage, typeof onCompact === "function");
  const r = 9;
  const circ = 2 * Math.PI * r;
  const filled = pct / 100 * circ;
  return fe`
        <button
            class="compose-context-pie icon-btn"
            type="button"
            title=${title}
            aria-label=${label}
            disabled=${typeof onCompact !== "function"}
            onClick=${(e) => {
    e.preventDefault();
    e.stopPropagation();
    onCompact?.();
  }}
        >
            <svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="12" cy="12" r=${r}
                    fill="none"
                    stroke="var(--context-track, rgba(128,128,128,0.2))"
                    stroke-width="2.5" />
                <circle cx="12" cy="12" r=${r}
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
function formatK(n) {
  if (n == null)
    return "?";
  if (n >= 1e6)
    return (n / 1e6).toFixed(1) + "M";
  if (n >= 1000)
    return (n / 1000).toFixed(0) + "K";
  return String(n);
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
      const label = item.trim();
      if (!label)
        continue;
      const slashIndex = label.indexOf("/");
      const provider = slashIndex > 0 ? label.slice(0, slashIndex).trim() : "";
      const id = slashIndex > 0 ? label.slice(slashIndex + 1).trim() : label;
      options.push({
        label,
        provider,
        id,
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
  options.sort((a, b) => a.label.localeCompare(b.label, undefined, { sensitivity: "base" }));
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
  for (let i = 0;i < lines.length; i += 1) {
    if (lines[i].trim() === "Files:" && lines[i + 1] && /^\s*-\s+/.test(lines[i + 1])) {
      start = i;
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
  for (let i = 0;i < lines.length; i += 1) {
    if (lines[i].trim() === "Referenced messages:" && lines[i + 1] && /^\s*-\s+/.test(lines[i + 1])) {
      start = i;
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
  for (let i = 0;i < lines.length; i += 1) {
    if (lines[i].trim() === "Attachments:" && lines[i + 1] && /^\s*-\s+/.test(lines[i + 1])) {
      start = i;
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
  busy = false,
  onReturnQueuedFollowup,
  onInjectQueuedFollowup,
  onRemoveQueuedFollowup,
  onMoveQueuedFollowup,
  onOpenFilePill
}) {
  if (!Array.isArray(items) || items.length === 0)
    return null;
  return fe`
        <div class="compose-queue-stack">
            ${items.map((item, index) => {
    const rowText = typeof item?.content === "string" ? item.content : "";
    const parsed = parseQueuedContent(rowText);
    if (!parsed.text.trim() && parsed.fileRefs.length === 0 && parsed.messageRefs.length === 0 && parsed.attachmentRefs.length === 0)
      return null;
    const canMoveUp = index > 0;
    const canMoveDown = index < items.length - 1;
    return fe`
                    <div class="compose-queue-stack-item" role="listitem" data-queue-id=${item.id} aria-busy=${item.pending ? "true" : "false"}>
                        <div class="compose-queue-stack-content" title=${rowText}>
                            ${parsed.text.trim() && fe`<div class="compose-queue-stack-text">${parsed.text}</div>`}
                            ${(parsed.messageRefs.length > 0 || parsed.fileRefs.length > 0 || parsed.attachmentRefs.length > 0) && fe`
                                <div class="compose-queue-stack-refs">
                                    ${parsed.messageRefs.map((id) => fe`
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
      return fe`
                                            <${FilePill}
                                                key=${"queue-file-" + path}
                                                prefix="compose"
                                                label=${label}
                                                title=${path}
                                                onClick=${() => onOpenFilePill?.(path)}
                                            />
                                        `;
    })}
                                    ${parsed.attachmentRefs.map((attachment) => fe`
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
                            ${items.length > 1 && fe`
                                <button
                                    class="compose-queue-stack-move-btn"
                                    type="button"
                                    title="Move up"
                                    aria-label="Move up in queue"
                                    disabled=${busy || item.pending || items.some((entry) => entry.pending) || !canMoveUp}
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
                                    disabled=${busy || item.pending || items.some((entry) => entry.pending) || !canMoveDown}
                                    onClick=${() => canMoveDown && onMoveQueuedFollowup?.(index, index + 1)}
                                >
                                    <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                                        <polyline points="6 9 12 15 18 9"></polyline>
                                    </svg>
                                </button>
                            `}
                            ${typeof onReturnQueuedFollowup === "function" && fe`
                                <button type="button" class="compose-queue-stack-move-btn"
                                    title="Return to editor" aria-label="Return queued message to editor"
                                    disabled=${busy || item.pending} onClick=${() => onReturnQueuedFollowup(item)}>Return</button>
                            `}
                            ${typeof onInjectQueuedFollowup === "function" && fe`<button
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
                            </button>`}
                            <button
                                class="compose-queue-stack-close-btn"
                                type="button"
                                title="Cancel queued message"
                                aria-label="Cancel queued message"
                                disabled=${busy || item.pending}
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
  onModelMutationStart,
  onModelMutationEnd,
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
  onPinSession,
  onArchiveSession,
  showQueueStack = true,
  statusNotice = null,
  extensionWorkingState = null,
  prefillRequest = null,
  draftValue = "",
  draftMediaFiles = [],
  onContentChange,
  onDraftMediaChange,
  onCaptureDraft,
  onQueuedSubmissionStart,
  onQueuedSubmissionEnd,
  onDraftAccepted,
  onDraftFailed,
  onDraftStorageError,
  focusRestoredDraft = false
}) {
  const [content, setContent] = F_(draftValue);
  const mountedRef = Q_(true);
  W_(() => () => {
    mountedRef.current = false;
  }, []);
  const [searchText, setSearchText] = F_("");
  const [mediaFiles, setMediaFiles] = F_(draftMediaFiles);
  const [isDragActive, setIsDragActive] = F_(false);
  const [slashMatches, setSlashMatches] = F_([]);
  const [slashIndex, setSlashIndex] = F_(0);
  const [showSlash, setShowSlash] = F_(false);
  const dynamicCommandsRef = Q_(null);
  const [mentionMatches, setMentionMatches] = F_([]);
  const [mentionIndex, setMentionIndex] = F_(0);
  const [showMention, setShowMention] = F_(false);
  const [switchingModel, setSwitchingModel] = F_(false);
  const [showModelPopup, setShowModelPopup] = F_(false);
  const [showSessionPopup, setShowSessionPopup] = F_(false);
  const [modelOptions, setModelOptions] = F_([]);
  const [modelPopupIndex, setModelPopupIndex] = F_(0);
  const [modelQuery, setModelQuery] = F_("");
  const visibleModels = u_(() => filterModelOptions(modelOptions, modelQuery, getModelPickerOptionSearchLabel), [modelOptions, modelQuery]);
  const modelEntries = u_(() => visibleModels.map((option) => ({
    label: getModelPickerOptionSearchLabel(option),
    disabled: switchingModel || modelContextBlocked(option, contextUsage)
  })), [visibleModels, switchingModel, contextUsage]);
  const [sessionPopupIndex, setSessionPopupIndex] = F_(0);
  const [sessionPopupQuery, setSessionPopupQuery] = F_("");
  const [sessionMutationPending, setSessionMutationPending] = F_("");
  const [sessionMutationError, setSessionMutationError] = F_("");
  const [sessionMutationNotice, setSessionMutationNotice] = F_("");
  const [sessionEdit, setSessionEdit] = F_(null);
  const sessionMutationLock = Q_(false);
  const sessionPopupEpoch = Q_(0);
  const sessionEditRef = Q_(null);
  const [loadingModels, setLoadingModels] = F_(false);
  const [footerWidth, setFooterWidth] = F_(0);
  const [submitError, setSubmitError] = F_(null);
  const [submitNotice, setSubmitNotice] = F_(null);
  const [statusNoticeNowMs, setStatusNoticeNowMs] = F_(() => Date.now());
  const [extensionWorkingFrameIndex, setExtensionWorkingFrameIndex] = F_(0);
  const textareaRef = Q_(null);
  const slashRef = Q_(null);
  const mentionRef = Q_(null);
  const modelPopupRef = Q_(null);
  const modelHintRef = Q_(null);
  const sessionPopupRef = Q_(null);
  const sessionTriggerRef = Q_(null);
  const sessionSearchRef = Q_(null);
  const sessionReturnFocusRef = Q_(null);
  const footerRef = Q_(null);
  const popupTypeaheadRef = Q_({ value: "", updatedAt: 0 });
  const dragCounterRef = Q_(0);
  const renameSessionInProgressRef = Q_(false);
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
  const saveHistory = (history2, storageKey = historyStorageKey) => {
    setLocalStorageItem(storageKey, JSON.stringify(history2));
  };
  const historyRef = Q_(loadHistory(historyStorageKey));
  const historyIndexRef = Q_(-1);
  const historyDraftRef = Q_("");
  const lastPrefillTokenRef = Q_("");
  K_(() => {
    historyRef.current = loadHistory(historyStorageKey);
    historyIndexRef.current = -1;
    historyDraftRef.current = "";
  }, [historyStorageKey]);
  K_(() => {
    let cancelled = false;
    const chatJid = currentChatJid || "web:default";
    fetch(`/agent/commands?chat_jid=${encodeURIComponent(chatJid)}`).then((r) => r.ok ? r.json() : null).then((data) => {
      if (cancelled || !data?.commands)
        return;
      dynamicCommandsRef.current = data.commands.map((c) => ({
        name: c.name,
        description: c.description || ""
      }));
    }).catch((e) => {
      console.debug("[compose] failed to fetch dynamic commands", e);
    });
    return () => {
      cancelled = true;
    };
  }, [currentChatJid]);
  K_(() => {
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
  W_(() => {
    onContentChange?.(content);
  }, [content]);
  W_(() => {
    onDraftMediaChange?.(mediaFiles);
  }, [mediaFiles]);
  W_(() => {
    if (focusRestoredDraft)
      textareaRef.current?.focus();
  }, []);
  const latestDraftRef = Q_(null);
  latestDraftRef.current = { text: content, media: mediaFiles, fileRefs, messageRefs };
  const canSend = content.trim() || mediaFiles.length > 0 || fileRefs.length > 0 || messageRefs.length > 0;
  const canShareLocation = typeof window !== "undefined" && typeof navigator !== "undefined" && Boolean(window.isSecureContext) && typeof navigator.geolocation?.getCurrentPosition === "function";
  const notificationsSupported = typeof window !== "undefined" && typeof Notification !== "undefined";
  const notificationsSecure = typeof window !== "undefined" ? Boolean(window.isSecureContext) : false;
  const notificationDenied = notificationPermission === "denied";
  const notificationsAvailable = notificationsSupported && notificationsSecure && !notificationDenied && typeof onToggleNotifications === "function";
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
  const mentionAgents = (Array.isArray(activeChatAgents) ? activeChatAgents : []).filter((chat) => !chat?.archived_at).map((chat) => ({ ...chat, agent_name: chat.agent_id || chat.agent_name }));
  const currentSessionAgent = (() => {
    for (const chat of Array.isArray(activeChatAgents) ? activeChatAgents : []) {
      const chatJid = typeof chat?.chat_jid === "string" ? chat.chat_jid.trim() : "";
      if (chatJid && chatJid === currentChatJid)
        return chat;
    }
    return null;
  })();
  const isCurrentRootSession = Boolean(currentSessionAgent && currentSessionAgent.chat_jid === (currentSessionAgent.root_chat_jid || currentSessionAgent.chat_jid));
  const switchableChatAgents = u_(() => {
    const seen = new Set;
    const chats = [];
    for (const chat of Array.isArray(activeChatAgents) ? activeChatAgents : []) {
      const chatJid = typeof chat?.chat_jid === "string" ? chat.chat_jid.trim() : "";
      if (!chatJid || seen.has(chatJid))
        continue;
      const agentName = typeof chat?.agent_name === "string" ? chat.agent_name.trim() : "";
      if (!agentName)
        continue;
      seen.add(chatJid);
      chats.push(chat);
    }
    return chats;
  }, [activeChatAgents, currentChatJid]);
  const sessionPopupGroups = u_(() => {
    const matches = new Set(filterSessionPickerChats(switchableChatAgents, sessionPopupQuery).map((chat) => chat.chat_jid));
    return groupSessionPickerChats(switchableChatAgents, currentChatJid, switchableChatAgents.filter((chat) => chat.pinned).map((chat) => chat.chat_jid)).map((group) => ({ ...group, items: group.items.filter((chat) => matches.has(chat.chat_jid)) })).filter((group) => group.items.length > 0);
  }, [switchableChatAgents, currentChatJid, sessionPopupQuery]);
  const orderedSessionChats = u_(() => sessionPopupGroups.flatMap((group) => group.items), [sessionPopupGroups]);
  const hasSwitchableChatAgents = switchableChatAgents.some((chat) => chat.chat_jid !== currentChatJid);
  const canSwitchSession = hasSwitchableChatAgents && typeof onSwitchChat === "function";
  const canRestoreSession = hasSwitchableChatAgents && typeof onRestoreSession === "function";
  const renameInProgress = Boolean(isRenameSessionInProgress || renameSessionInProgressRef.current);
  const canRenameSession = !searchMode && typeof onRenameSession === "function" && !renameInProgress && currentSessionAgent?.capabilities?.rename !== false;
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
  const showComposeMetaRow = !searchMode && (showModelPickerHint || contextUsage);
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
        provider_usage: payload.provider_usage ?? null,
        context_usage: payload.context_usage
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
  const closeSessionPopup = (restoreFocus = false) => {
    setShowSessionPopup(false);
    if (restoreFocus)
      requestAnimationFrame(() => {
        const active = document.activeElement;
        if (active && active !== document.body && active.isConnected && !sessionPopupRef.current?.contains(active))
          return;
        const target = sessionReturnFocusRef.current;
        if (target?.isConnected)
          target.focus();
        else
          sessionTriggerRef.current?.querySelector("button")?.focus();
      });
  };
  const openSessionPopup = (trigger = null) => {
    if (searchMode || !canSwitchSession && !canRestoreSession && !canRenameSession && !canCreateSession && !canDeleteSession)
      return false;
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    setShowModelPopup(false);
    setShowSlash(false);
    setSlashMatches([]);
    setShowMention(false);
    setMentionMatches([]);
    sessionReturnFocusRef.current = trigger || sessionTriggerRef.current?.querySelector("button");
    setSessionPopupQuery("");
    setSessionMutationError("");
    setSessionMutationNotice("");
    setSessionEdit(null);
    setSessionPopupIndex(0);
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
    openSessionPopup(event?.currentTarget);
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
  const runSessionMutation = async (chat, action, value = undefined) => {
    if (!chat?.chat_jid || sessionMutationLock.current)
      return;
    const callback = { rename: onRenameSession, pin: onPinSession, archive: onArchiveSession, restore: onRestoreSession }[action];
    if (typeof callback !== "function")
      return;
    sessionMutationLock.current = true;
    const epoch = sessionPopupEpoch.current;
    setSessionMutationPending(`${action}:${chat.chat_jid}`);
    setSessionMutationError("");
    setSessionMutationNotice("");
    try {
      await callback(chat.chat_jid, value);
      if (epoch !== sessionPopupEpoch.current)
        return;
      setSessionEdit(null);
      setSessionMutationNotice(`${{ rename: "Renamed", pin: value ? "Pinned" : "Unpinned", archive: "Archived", restore: "Restored" }[action]} session.`);
      requestAnimationFrame(() => {
        if (epoch === sessionPopupEpoch.current)
          sessionSearchRef.current?.focus();
      });
    } catch (error) {
      if (epoch === sessionPopupEpoch.current)
        setSessionMutationError(error?.message || `Failed to ${action} session`);
    } finally {
      sessionMutationLock.current = false;
      setSessionMutationPending("");
    }
  };
  const handleRestoreSession = async (chatJid) => {
    const chat = switchableChatAgents.find((chat) => chat.chat_jid === chatJid);
    await runSessionMutation(chat, "restore");
  };
  const beginSessionEdit = (chat, action) => {
    if (sessionMutationLock.current)
      return;
    setSessionMutationError("");
    setSessionMutationNotice("");
    setSessionEdit({ chat, action, title: chat.agent_name || "" });
  };
  const findFirstEnabledPopupIndex = (items) => {
    const list = Array.isArray(items) ? items : [];
    const index = list.findIndex((item) => !item?.disabled);
    return index >= 0 ? index : 0;
  };
  const sessionPopupEntries = u_(() => {
    const entries = [];
    for (const chat of orderedSessionChats) {
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
    if (!sessionPopupQuery.trim() && canCreateSession) {
      entries.push({ type: "action", key: "action:new", label: "New session", action: "new", disabled: false });
    }
    if (!sessionPopupQuery.trim() && canRenameSession) {
      entries.push({ type: "action", key: "action:rename", label: "Rename current session", action: "rename", disabled: renameInProgress });
    }
    if (!sessionPopupQuery.trim() && canDeleteSession) {
      entries.push({ type: "action", key: "action:delete", label: "Delete current session", action: "delete", disabled: false });
    }
    return entries;
  }, [orderedSessionChats, sessionPopupQuery, canRestoreSession, canSwitchSession, canCreateSession, canRenameSession, canDeleteSession, renameInProgress]);
  const handleRenameSession = async (event) => {
    if (event?.preventDefault)
      event.preventDefault();
    if (event?.stopPropagation)
      event.stopPropagation();
    if (typeof onRenameSession !== "function" || isRenameSessionInProgress || renameSessionInProgressRef.current)
      return;
    beginSessionEdit(currentSessionAgent, "rename");
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
  const handleCycleModel = async () => {
    try {
      const available = normalizeModelPickerOptions(await getAgentModels(currentChatJid));
      if (!mountedRef.current || !available.length)
        return;
      const index = available.findIndex((option) => option.label === activeModel);
      await handleSelectModel(available[(index + 1) % available.length]);
    } catch (error) {
      if (mountedRef.current)
        setSubmitError(`Model catalogue failed: ${error.message}`);
    }
  };
  const modelMutationRef = Q_(false);
  const modelRevisionRef = Q_(0);
  const handleSelectModel = async (modelOption) => {
    const modelLabel = typeof modelOption === "string" ? modelOption : modelOption?.label;
    if (!modelLabel || modelMutationRef.current)
      return;
    if (modelContextBlocked(modelOption, contextUsage)) {
      setSubmitError("Model context window is smaller than the latest measured request.");
      return;
    }
    modelMutationRef.current = true;
    ++modelRevisionRef.current;
    const mutation = onModelMutationStart?.();
    setSwitchingModel(true);
    setSubmitError(null);
    try {
      const state = await selectAgentModel(currentChatJid, modelLabel);
      if (!mountedRef.current)
        return;
      emitModelState(state);
      setShowModelPopup(false);
    } catch (error) {
      if (mountedRef.current)
        setSubmitError(`Model selection failed: ${error.message}`);
    } finally {
      modelMutationRef.current = false;
      onModelMutationEnd?.(mutation);
      if (mountedRef.current)
        setSwitchingModel(false);
    }
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
    setModelQuery("");
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
    const capturedDraft = { text: baseContent, media: capturedMediaFiles, fileRefs: capturedFileRefs, messageRefs: capturedMessageRefs };
    const capturedChatJid = currentChatJid;
    const mode = resolveSubmitMode(submitMode);
    const capture = clearAfterSubmit ? onCaptureDraft?.(capturedDraft) : null;
    const queueToken = mode === "queue" ? capture?.token || crypto.randomUUID() : null;
    if (queueToken)
      onQueuedSubmissionStart?.(queueToken, baseContent || "[attachments]");
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
    const restoreDraft = (message) => {
      if (capture && onDraftFailed) {
        onDraftFailed(capture.token, message);
        return;
      }
      if (!mountedRef.current)
        return;
      const restored = mergeDrafts(capturedDraft, latestDraftRef.current);
      if (includeMedia)
        setMediaFiles(restored.media);
      if (includeFileRefs)
        onSetFileRefs?.(restored.fileRefs);
      if (includeMessageRefs)
        onSetMessageRefs?.(restored.messageRefs);
      setContent(restored.text);
      requestAnimationFrame(() => resizeTextarea());
    };
    const acknowledge = async () => {
      if (!capture)
        return;
      try {
        await onDraftAccepted?.(capture.token);
      } catch (error) {
        onDraftStorageError?.(error);
      }
    };
    if (clearAfterSubmit) {
      setContent("");
      setMediaFiles([]);
      onClearFileRefs?.();
      onClearMessageRefs?.();
    }
    let requestDispatched = false;
    let requestAcknowledged = false;
    (async () => {
      try {
        await capture?.ready;
        const intercepted = await onSubmitIntercept?.({
          content: baseContent,
          submitMode,
          fileRefs: capturedFileRefs,
          messageRefs: capturedMessageRefs,
          mediaFiles: capturedMediaFiles
        });
        if (intercepted) {
          requestAcknowledged = true;
          await acknowledge();
          onPost?.(intercepted);
          return;
        }
        const mediaIds = [];
        for (const file of capturedMediaFiles) {
          const result = await uploadMedia(file, capturedChatJid);
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
        requestDispatched = true;
        const response = await sendAgentMessage("default", message, null, mediaIds, mode, capturedChatJid, { client_request_id: queueToken });
        requestAcknowledged = true;
        await acknowledge();
        if (!mountedRef.current)
          return;
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
        const detail = error?.message || "Failed to send message.";
        if (requestAcknowledged) {
          if (mountedRef.current)
            setSubmitError(`Send acknowledged, but refresh failed: ${detail}`);
          return;
        }
        const uncertain = requestDispatched && ["TypeError", "AbortError"].includes(error?.name);
        const message = uncertain ? `Delivery is unknown; check the timeline before resending. ${detail}` : detail;
        if (clearAfterSubmit) {
          restoreDraft(message);
        }
        if (!mountedRef.current)
          return;
        if (!clearAfterSubmit || !onDraftFailed)
          setSubmitError(message);
        onSubmitError?.(message);
        console.error("Failed to post:", error);
      } finally {
        if (queueToken)
          onQueuedSubmissionEnd?.(queueToken);
      }
    })();
  };
  const handleInjectQueuedFollowup = (queuedItem) => {
    onInjectQueuedFollowup?.(queuedItem);
  };
  const handlePopupKeyboardEvent = Y_((e) => {
    if (settingsOwnsKeyboard())
      return false;
    if (searchMode || !showModelPopup && !showSessionPopup || e?.isComposing)
      return false;
    const consume = () => {
      e.preventDefault?.();
      e.stopPropagation?.();
    };
    const resetPopupTypeahead = () => {
      popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    };
    if (e.key === "Escape") {
      consume();
      if (showSessionPopup && sessionEdit) {
        setSessionEdit(null);
        sessionSearchRef.current?.focus();
        return true;
      }
      resetPopupTypeahead();
      if (showModelPopup) {
        setShowModelPopup(false);
        requestAnimationFrame(() => {
          if (document.activeElement === document.body)
            modelHintRef.current?.focus();
        });
      }
      if (showSessionPopup)
        closeSessionPopup(true);
      return true;
    }
    if (showModelPopup && modelPopupRef.current?.contains(e.target)) {
      const action = modelPickerKey(e, modelEntries, modelPopupIndex, popupTypeaheadRef.current);
      if (action) {
        consume();
        popupTypeaheadRef.current = action.buffer;
        if (action.index >= 0) {
          setModelPopupIndex(action.index);
          if (action.focus)
            modelPopupRef.current?.querySelector('[data-model-index="' + action.index + '"]')?.focus({ preventScroll: true });
          if (action.activate)
            handleSelectModel(visibleModels[action.index]);
        }
        return true;
      }
    }
    if (showSessionPopup && !sessionEdit && !sessionMutationPending && sessionPopupRef.current?.contains(e.target)) {
      const inSearch = e.target === sessionSearchRef.current;
      if (!inSearch) {
        const typed = sessionTypeahead(e, sessionPopupEntries, popupTypeaheadRef.current);
        if (typed) {
          consume();
          popupTypeaheadRef.current = typed.buffer;
          if (typed.index >= 0) {
            setSessionPopupIndex(typed.index);
            const entry = sessionPopupEntries[typed.index];
            const target = Array.from(sessionPopupRef.current.querySelectorAll("[data-session-entry-key]")).find((node) => node.dataset.sessionEntryKey === entry.key);
            target?.focus({ preventScroll: true });
          }
          return true;
        }
      }
      const navigation = ["ArrowDown", "ArrowUp", "PageDown", "PageUp"].includes(e.key) || !inSearch && ["Home", "End"].includes(e.key);
      if (navigation && !e.ctrlKey && !e.metaKey && !e.altKey) {
        consume();
        resetPopupTypeahead();
        const enabled = sessionPopupEntries.map((entry, index) => ({ entry, index })).filter(({ entry }) => !entry.disabled);
        const focusedKey = e.target?.closest?.("[data-session-entry-key]")?.dataset.sessionEntryKey;
        setSessionPopupIndex((current) => {
          const index = focusedKey ? sessionPopupEntries.findIndex((entry) => entry.key === focusedKey) : current;
          const selected = enabled.findIndex((item) => item.index === index);
          return enabled[moveSessionPickerIndex(selected, enabled.length, e.key)]?.index ?? 0;
        });
        sessionSearchRef.current?.focus();
        return true;
      }
      if (inSearch && e.key === "Enter") {
        consume();
        runSessionPopupEntry(sessionPopupEntries[sessionPopupIndex]);
        return true;
      }
    }
    return false;
  }, [
    searchMode,
    showModelPopup,
    showSessionPopup,
    visibleModels,
    modelEntries,
    modelPopupIndex,
    sessionPopupEntries,
    sessionPopupIndex,
    sessionEdit,
    sessionMutationPending,
    handleSelectModel
  ]);
  const handleKeyDown = (e) => {
    if (e.isComposing)
      return;
    if (searchMode && e.key === "Escape") {
      e.preventDefault();
      setSearchText("");
      onExitSearch?.();
      return;
    }
    if (handlePopupKeyboardEvent(e)) {
      return;
    }
    const currentValue = textareaRef.current?.value ?? (searchMode ? searchText : content);
    if (shouldOpenSessionSwitcherFromBlankCompose(e, currentValue, {
      searchMode,
      showSessionSwitcherButton
    })) {
      e.preventDefault();
      openSessionPopup();
      return;
    }
    if (showMention && mentionMatches.length > 0) {
      const mentionValue = textareaRef.current?.value ?? (searchMode ? searchText : content);
      if (!String(mentionValue || "").match(/^@([a-zA-Z0-9_-]*)$/)) {
        setShowMention(false);
        setMentionMatches([]);
      } else {
        if (e.key === "ArrowDown") {
          e.preventDefault();
          setMentionIndex((i) => (i + 1) % mentionMatches.length);
          return;
        }
        if (e.key === "ArrowUp") {
          e.preventDefault();
          setMentionIndex((i) => (i - 1 + mentionMatches.length) % mentionMatches.length);
          return;
        }
        if (e.key === "Tab" || e.key === "Enter") {
          e.preventDefault();
          acceptMention(mentionMatches[mentionIndex]);
          return;
        }
        if (e.key === "Escape") {
          e.preventDefault();
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
        if (e.key === "ArrowDown") {
          e.preventDefault();
          setSlashIndex((i) => (i + 1) % slashMatches.length);
          return;
        }
        if (e.key === "ArrowUp") {
          e.preventDefault();
          setSlashIndex((i) => (i - 1 + slashMatches.length) % slashMatches.length);
          return;
        }
        if (e.key === "Tab") {
          e.preventDefault();
          acceptSlashCommand(slashMatches[slashIndex]);
          return;
        }
        if (e.key === "Enter" && !e.shiftKey) {
          const hasArgs = currentValue.includes(" ");
          if (!hasArgs) {
            e.preventDefault();
            const cmd = slashMatches[slashIndex];
            setShowSlash(false);
            setSlashMatches([]);
            handleSubmit(cmd.name);
            return;
          }
        }
        if (e.key === "Escape") {
          e.preventDefault();
          setShowSlash(false);
          setSlashMatches([]);
          return;
        }
      }
    }
    if (!searchMode && (e.key === "ArrowUp" || e.key === "ArrowDown") && !e.metaKey && !e.ctrlKey && !e.altKey && !e.shiftKey) {
      const textarea = textareaRef.current;
      if (!textarea)
        return;
      const value = textarea.value || "";
      const atStart = textarea.selectionStart === 0 && textarea.selectionEnd === 0;
      const atEnd = textarea.selectionStart === value.length && textarea.selectionEnd === value.length;
      if (e.key === "ArrowUp" && atStart || e.key === "ArrowDown" && atEnd) {
        const history2 = historyRef.current;
        if (!history2.length)
          return;
        e.preventDefault();
        let idx = historyIndexRef.current;
        if (e.key === "ArrowUp") {
          if (idx === -1) {
            historyDraftRef.current = value;
            idx = history2.length - 1;
          } else if (idx > 0) {
            idx -= 1;
          }
          historyIndexRef.current = idx;
          updateValue(history2[idx] || "");
        } else {
          if (idx === -1)
            return;
          if (idx < history2.length - 1) {
            idx += 1;
            historyIndexRef.current = idx;
            updateValue(history2[idx] || "");
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
    if (e.key === "Enter" && !e.shiftKey && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      if (searchMode) {
        if (currentValue.trim()) {
          onSearch?.(currentValue.trim(), searchScope);
        }
      } else {
        handleSubmit(currentValue, "steer");
      }
      return;
    }
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
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
    if (list.some((file) => file.size > 10 * 1024 * 1024)) {
      setSubmitError("Media exceeds 10 MiB limit");
      return;
    }
    setMediaFiles((current) => [...current, ...list]);
    setSubmitError(null);
  };
  const handleFileChange = (e) => {
    addMediaFiles(e.target.files);
    e.target.value = "";
  };
  const handleDragEnter = (e) => {
    if (searchMode)
      return;
    e.preventDefault();
    e.stopPropagation();
    dragCounterRef.current += 1;
    setIsDragActive(true);
  };
  const handleDragLeave = (e) => {
    if (searchMode)
      return;
    e.preventDefault();
    e.stopPropagation();
    dragCounterRef.current = Math.max(0, dragCounterRef.current - 1);
    if (dragCounterRef.current === 0)
      setIsDragActive(false);
  };
  const handleDragOver = (e) => {
    if (searchMode)
      return;
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer)
      e.dataTransfer.dropEffect = "copy";
    setIsDragActive(true);
  };
  const handleDrop = (e) => {
    if (searchMode)
      return;
    e.preventDefault();
    e.stopPropagation();
    dragCounterRef.current = 0;
    setIsDragActive(false);
    addMediaFiles(e.dataTransfer?.files || []);
  };
  const handlePaste = (e) => {
    if (searchMode)
      return;
    const items = e.clipboardData?.items;
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
      e.preventDefault();
      addMediaFiles(files);
    }
  };
  const removeMediaFile = (index) => {
    setMediaFiles((current) => current.filter((_, idx) => idx !== index));
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
  K_(() => {
    if (!showModelPopup)
      return;
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    setLoadingModels(true);
    const revision = modelRevisionRef.current;
    let active = true;
    getAgentModels(currentChatJid).then((payload) => {
      if (!active || !mountedRef.current || revision !== modelRevisionRef.current)
        return;
      setModelOptions(normalizeModelPickerOptions(payload));
      emitModelState(payload);
    }).catch((error) => {
      if (!active || !mountedRef.current)
        return;
      setSubmitError(`Model catalogue failed: ${error.message}`);
      setModelOptions([]);
    }).finally(() => {
      if (active && mountedRef.current)
        setLoadingModels(false);
    });
    return () => {
      active = false;
    };
  }, [showModelPopup, activeModel]);
  K_(() => {
    if (searchMode) {
      setShowModelPopup(false);
      setShowSessionPopup(false);
      setShowSlash(false);
      setSlashMatches([]);
      setShowMention(false);
      setMentionMatches([]);
    }
  }, [searchMode]);
  K_(() => {
    if (showSessionPopup && !showSessionSwitcherButton) {
      setShowSessionPopup(false);
    }
  }, [showSessionPopup, showSessionSwitcherButton]);
  W_(() => {
    if (!showModelPopup)
      return;
    const activeIndex = visibleModels.findIndex((model, index) => model?.label === activeModel && !modelEntries[index].disabled);
    setModelPopupIndex(activeIndex >= 0 ? activeIndex : modelEntries.findIndex((entry) => !entry.disabled));
  }, [showModelPopup, visibleModels, activeModel, modelEntries]);
  K_(() => {
    if (!showSessionPopup)
      return;
    const preferred = resolveSessionPickerSearchInitialIndex(orderedSessionChats, sessionPopupQuery);
    setSessionPopupIndex(sessionPopupEntries[preferred]?.disabled ? findFirstEnabledPopupIndex(sessionPopupEntries) : preferred);
  }, [showSessionPopup, currentChatJid, sessionPopupQuery]);
  K_(() => {
    setSessionPopupIndex((index) => Math.max(0, Math.min(index, sessionPopupEntries.length - 1)));
  }, [sessionPopupEntries.length]);
  K_(() => {
    if (!showModelPopup)
      return;
    const onPointerDown = (event) => {
      if (settingsOwnsKeyboard())
        return;
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
  K_(() => {
    if (!showSessionPopup)
      return;
    const onPointerDown = (event) => {
      if (settingsOwnsKeyboard())
        return;
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
  W_(() => {
    if (searchMode || !showModelPopup && !showSessionPopup)
      return;
    const onKeyDown = (event) => {
      handlePopupKeyboardEvent(event);
    };
    document.addEventListener("keydown", onKeyDown, true);
    return () => document.removeEventListener("keydown", onKeyDown, true);
  }, [searchMode, showModelPopup, showSessionPopup, handlePopupKeyboardEvent]);
  W_(() => {
    if (showModelPopup)
      modelPopupRef.current?.focus();
  }, [showModelPopup]);
  K_(() => {
    if (!showModelPopup)
      return;
    const popup = modelPopupRef.current;
    const active = popup?.querySelector?.(".compose-model-popup-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showModelPopup, modelPopupIndex, modelOptions]);
  W_(() => {
    ++sessionPopupEpoch.current;
    if (showSessionPopup)
      sessionSearchRef.current?.focus();
    return () => {
      ++sessionPopupEpoch.current;
    };
  }, [showSessionPopup]);
  W_(() => {
    if (sessionEdit)
      sessionEditRef.current?.focus();
  }, [sessionEdit?.chat.chat_jid, sessionEdit?.action]);
  K_(() => {
    if (!showSessionPopup)
      return;
    const active = sessionPopupRef.current?.querySelector("[data-session-entry-key].active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showSessionPopup, sessionPopupIndex, sessionPopupEntries.length]);
  K_(() => {
    if (!showMention || !mentionRef.current)
      return;
    const popup = mentionRef.current;
    const active = popup.querySelector?.(".slash-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showMention, mentionIndex, mentionMatches.length]);
  K_(() => {
    if (!showSlash || !slashRef.current)
      return;
    const popup = slashRef.current;
    const active = popup.querySelector?.(".slash-item.active");
    active?.scrollIntoView?.({ block: "nearest" });
  }, [showSlash, slashIndex, slashMatches.length]);
  K_(() => {
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
  const handleInput = (e) => {
    const value = e.target.value;
    setSubmitError(null);
    setSubmitNotice(null);
    if (showSessionPopup)
      setShowSessionPopup(false);
    resizeTextarea(e.target);
    updateValue(value);
  };
  K_(() => {
    requestAnimationFrame(() => resizeTextarea());
  }, [content, searchText, searchMode]);
  K_(() => {
    if (!statusNoticeIsCompaction)
      return;
    setStatusNoticeNowMs(Date.now());
    const timer = setInterval(() => setStatusNoticeNowMs(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [statusNoticeIsCompaction, statusNotice?.started_at, statusNotice?.startedAt]);
  K_(() => {
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
  K_(() => {
    if (searchMode)
      return;
    updateMentionAutocomplete(content);
  }, [mentionAgents, currentChatJid, content, searchMode]);
  return fe`
        <div class="compose-box">
            ${showQueueStack && !searchMode && fe`
                <${QueuedFollowupStack}
                    items=${followupQueueItems}
                    onInjectQueuedFollowup=${handleInjectQueuedFollowup}
                    onRemoveQueuedFollowup=${onRemoveQueuedFollowup}
                    onMoveQueuedFollowup=${onMoveQueuedFollowup}
                    onOpenFilePill=${onOpenFilePill}
                />
            `}
            ${extensionWorkingDisplay.visible && fe`
                <div class="compose-inline-status extension-working" role="status" aria-live="polite">
                    <div class="compose-inline-status-row">
                        ${extensionWorkingDisplay.indicatorText ? fe`<span class="compose-inline-status-glyph" aria-hidden="true">${extensionWorkingDisplay.indicatorText}</span>` : extensionWorkingDisplay.animateDot ? fe`<span class=${buildComposeStatusDotClass({ pulsing: true })} aria-hidden="true"></span>` : null}
                        <span class="compose-inline-status-title">${extensionWorkingDisplay.title}</span>
                    </div>
                </div>
            `}
            ${statusNotice && fe`
                <div
                    class=${`compose-inline-status${statusNoticeIsCompaction ? " compaction" : ""}`}
                    role="status"
                    aria-live="polite"
                    title=${statusNoticeDetail || ""}
                >
                    <div class="compose-inline-status-row">
                        <span class=${buildComposeStatusDotClass({ pulsing: statusNoticeIsCompaction })} aria-hidden="true"></span>
                        <span class="compose-inline-status-title">${statusNoticeTitle}</span>
                        ${statusNoticeElapsedLabel && fe`<span class="compose-inline-status-elapsed">${statusNoticeElapsedLabel}</span>`}
                    </div>
                    ${statusNoticeDetail && fe`<div class="compose-inline-status-detail">${statusNoticeDetail}</div>`}
                </div>
            `}
            ${submitError && fe`<div class="compose-submit-error" role="alert">${submitError}</div>`}
            ${submitNotice && fe`
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
                    ${hasAttachments && fe`
                        <div class="compose-file-refs">
                            ${messageRefs.map((id) => {
    return fe`
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
    return fe`
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
    return fe`
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
                    ${!searchMode && typeof onPopOutChat === "function" && fe`
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
                    ${showMention && mentionMatches.length > 0 && fe`
                        <div class="slash-autocomplete" ref=${mentionRef}>
                            ${mentionMatches.map((agent, i) => fe`
                                <div
                                    key=${agent.chat_jid || agent.agent_name}
                                    class=${`slash-item${i === mentionIndex ? " active" : ""}`}
                                    onMouseDown=${(e) => {
    e.preventDefault();
    acceptMention(agent);
  }}
                                    onMouseEnter=${() => setMentionIndex(i)}
                                >
                                    <span class="slash-name">@${agent.agent_name}</span>
                                    <span class="slash-desc">${agent.chat_jid || "Active agent"}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${showSlash && slashMatches.length > 0 && fe`
                        <div class="slash-autocomplete" ref=${slashRef}>
                            ${slashMatches.map((cmd, i) => fe`
                                <div
                                    key=${cmd.name}
                                    class=${`slash-item${i === slashIndex ? " active" : ""}`}
                                    onMouseDown=${(e) => {
    e.preventDefault();
    acceptSlashCommand(cmd);
  }}
                                    onMouseEnter=${() => setSlashIndex(i)}
                                >
                                    <span class="slash-name">${cmd.name}</span>
                                    <span class="slash-desc">${cmd.description}</span>
                                </div>
                            `)}
                        </div>
                    `}
                    ${showModelPopup && !searchMode && fe`
                        <div class="compose-model-popup" ref=${modelPopupRef} tabIndex="-1" onKeyDown=${handlePopupKeyboardEvent}>
                            <div class="compose-model-popup-title">Select model</div>
                            <input type="search" class="compose-session-search" aria-label="Search models" placeholder="Search models"
                                value=${modelQuery} onInput=${(event) => {
    popupTypeaheadRef.current = { value: "", updatedAt: 0 };
    setModelQuery(event.currentTarget.value);
  }} />
                            <div class="compose-model-popup-menu" role="menu" aria-label="Model picker">
                                ${loadingModels && fe`
                                    <div class="compose-model-popup-empty">Loading models…</div>
                                `}
                                ${!loadingModels && visibleModels.length === 0 && fe`
                                    <div class="compose-model-popup-empty">${modelQuery ? "No models match your search." : "No models available."}</div>
                                `}
                                ${!loadingModels && visibleModels.map((modelOption, index) => {
    const modelLabel = typeof modelOption?.label === "string" ? modelOption.label : "";
    const contextWindowLabel = formatModelPickerContextWindow(modelOption?.contextWindow);
    const blocked = modelContextBlocked(modelOption, contextUsage);
    return fe`
                                        <button
                                            key=${modelLabel}
                                            data-model-index=${index}
                                            type="button"
                                            role="menuitem"
                                            class=${`compose-model-popup-item compose-model-popup-model-item${modelPopupIndex === index ? " active" : ""}${activeModel === modelLabel ? " current-model" : ""}`}
                                            onClick=${() => {
      handleSelectModel(modelOption);
    }}
                                            disabled=${switchingModel || blocked}
                                            title=${blocked ? `Blocked: ${modelLabel} context window is smaller than latest measured request` : [modelLabel, contextWindowLabel].filter(Boolean).join(" • ")}
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
                    ${showSessionPopup && !searchMode && fe`
                        <div class="compose-model-popup compose-session-popup" ref=${sessionPopupRef} tabIndex="-1" onKeyDown=${handlePopupKeyboardEvent}>
                            <div class="compose-model-popup-title">Manage sessions & agents</div>
                            ${sessionMutationError && fe`<div role="alert" class="compose-session-mutation-error">${sessionMutationError}</div>`}
                            ${sessionMutationNotice && fe`<div role="status" class="compose-session-mutation-notice">${sessionMutationNotice}</div>`}
                            ${sessionMutationPending && fe`<div role="status">Saving session…</div>`}
                            ${sessionEdit && fe`
                                <form class="compose-session-edit" onSubmit=${(event) => {
    event.preventDefault();
    runSessionMutation(sessionEdit.chat, sessionEdit.action, sessionEdit.title);
  }}>
                                    ${sessionEdit.action === "rename" ? fe`
                                        <label>Session name
                                            <input ref=${sessionEditRef} aria-label="Session name" value=${sessionEdit.title} maxLength="160"
                                                disabled=${Boolean(sessionMutationPending)}
                                                onInput=${(event) => setSessionEdit({ ...sessionEdit, title: event.currentTarget.value })} />
                                        </label>
                                    ` : fe`<p>Archive @${sessionEdit.chat.agent_name}? History and drafts are retained. Restore it from Archived.</p>`}
                                    <button ref=${sessionEdit.action === "archive" ? sessionEditRef : undefined} type="submit" class="compose-model-popup-btn"
                                        disabled=${Boolean(sessionMutationPending)}>${sessionEdit.action === "rename" ? "Save name" : "Confirm archive"}</button>
                                    <button type="button" class="compose-model-popup-btn" disabled=${Boolean(sessionMutationPending)} onClick=${() => {
    setSessionEdit(null);
    sessionSearchRef.current?.focus();
  }}>Cancel</button>
                                </form>
                            `}
                            <input
                                ref=${sessionSearchRef}
                                type="search"
                                class="compose-session-search"
                                aria-label="Search sessions"
                                aria-controls="compose-session-results"
                                placeholder="Handle, JID, state, or model"
                                value=${sessionPopupQuery}
                                disabled=${Boolean(sessionMutationPending)}
                                onInput=${(event) => setSessionPopupQuery(event.currentTarget.value)}
                            />
                            <div id="compose-session-results" class="compose-model-popup-menu" role="menu" aria-label="Sessions and agents">
                                ${orderedSessionChats.length === 0 && fe`
                                    <div class="compose-model-popup-empty" role="status">No sessions match your search.</div>
                                `}
                                ${sessionPopupGroups.map((group) => fe`
                                <div role="group" aria-label=${group.label}>
                                <div class="compose-session-section-label">${group.label}</div>
                                ${group.items.map((chat) => {
    const listIndex = sessionPopupEntries.findIndex((entry) => entry.key === `session:${chat.chat_jid}`);
    const archived = Boolean(chat.archived_at);
    const isRoot = chat.chat_jid === (chat.root_chat_jid || chat.chat_jid);
    const canPrune = !isRoot && !chat.is_active && !archived && typeof onDeleteSession === "function";
    const label = formatBranchPickerLabel(chat, { currentChatJid });
    return fe`
                                        <div key=${chat.chat_jid} data-session-jid=${chat.chat_jid} class=${`compose-model-popup-item-row${archived ? " archived" : ""}`}>
                                            <button
                                                type="button"
                                                role="menuitem"
                                                class=${`compose-model-popup-item${archived ? " archived" : ""}${sessionPopupIndex === listIndex ? " active" : ""}`}
                                                data-session-entry-key=${`session:${chat.chat_jid}`}
                                                aria-current=${chat.chat_jid === currentChatJid ? "true" : undefined}
                                                onClick=${() => {
      if (archived) {
        handleRestoreSession(chat.chat_jid);
        return;
      }
      handleSessionSwitch(chat.chat_jid);
    }}
                                                disabled=${Boolean(sessionMutationPending) || (archived ? !canRestoreSession || chat.capabilities?.restore === false : !canSwitchSession)}
                                                title=${archived ? `Restore archived ${`@${chat.agent_name}`}` : `Switch to ${`@${chat.agent_name}`}`}
                                            >
                                                ${label}
                                            </button>
                                            <div class="compose-session-row-actions">
                                                ${!archived && chat.capabilities?.pin !== false && typeof onPinSession === "function" && fe`
                                                    <button type="button" class="compose-model-popup-btn" disabled=${Boolean(sessionMutationPending)}
                                                        aria-label=${`${chat.pinned ? "Unpin" : "Pin"} @${chat.agent_name}`}
                                                        onClick=${() => {
      runSessionMutation(chat, "pin", !chat.pinned);
    }}>${chat.pinned ? "Unpin" : "Pin"}</button>
                                                `}
                                                ${!archived && chat.capabilities?.rename !== false && typeof onRenameSession === "function" && fe`
                                                    <button type="button" class="compose-model-popup-btn" disabled=${Boolean(sessionMutationPending)}
                                                        aria-label=${`Rename @${chat.agent_name}`} onClick=${() => beginSessionEdit(chat, "rename")}>Rename</button>
                                                `}
                                                ${!archived && !isRoot && !chat.is_active && chat.capabilities?.archive !== false && typeof onArchiveSession === "function" && fe`
                                                    <button type="button" class="compose-model-popup-btn" disabled=${Boolean(sessionMutationPending)}
                                                        aria-label=${`Archive @${chat.agent_name}`} onClick=${() => beginSessionEdit(chat, "archive")}>Archive</button>
                                                `}
                                                ${archived && chat.capabilities?.restore !== false && typeof onRestoreSession === "function" && fe`
                                                    <button type="button" class="compose-model-popup-btn" disabled=${Boolean(sessionMutationPending)}
                                                        aria-label=${`Restore @${chat.agent_name}`} onClick=${() => {
      runSessionMutation(chat, "restore");
    }}>Restore</button>
                                                `}
                                            </div>
                                            ${canPrune && fe`
                                                <button
                                                    type="button"
                                                    class="compose-model-popup-item-delete"
                                                    title="Delete this branch"
                                                    aria-label=${`Delete @${chat.agent_name}`}
                                                    onClick=${(e) => {
      e.stopPropagation();
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
                                `)}
                            </div>
                            ${!sessionPopupQuery.trim() && (canCreateSession || canRenameSession || canDeleteSession) && fe`
                                <div class="compose-model-popup-actions">
                                    ${canCreateSession && fe`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn primary${sessionPopupEntries.findIndex((entry) => entry.key === "action:new") === sessionPopupIndex ? " active" : ""}`}
                                            onClick=${() => {
    handleCreateSession();
  }}
                                            data-session-entry-key="action:new"
                                            title="Create a new agent/session branch from this chat"
                                        >
                                            New
                                        </button>
                                    `}
                                    ${canRenameSession && fe`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn${sessionPopupEntries.findIndex((entry) => entry.key === "action:rename") === sessionPopupIndex ? " active" : ""}`}
                                            onClick=${(e) => {
    handleRenameSession(e);
  }}
                                            data-session-entry-key="action:rename"
                                            title="Rename the current branch handle"
                                            disabled=${renameInProgress}
                                        >
                                            Rename current…
                                        </button>
                                    `}
                                    ${canDeleteSession && fe`
                                        <button
                                            type="button"
                                            class=${`compose-model-popup-btn danger${sessionPopupEntries.findIndex((entry) => entry.key === "action:delete") === sessionPopupIndex ? " active" : ""}`}
                                            onClick=${() => {
    handleDeleteSession();
  }}
                                            data-session-entry-key="action:delete"
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
                    ${showComposeMetaRow && fe`
                    <div class="compose-meta-row">
                        ${showModelPickerHint && fe`
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
                                    ${!switchingModel && modelUsageSectionLabel && fe`
                                        <span class="compose-model-usage-hint" title=${modelHintTitle}>
                                            ${modelUsageSectionLabel}
                                        </span>
                                    `}
                                </div>
                            </div>
                        `}
                        ${!searchMode && contextUsage && fe`
                            <${ContextPie} usage=${contextUsage} onCompact=${typeof onContextCompact === "function" ? handleContextCompact : undefined} />
                        `}
                    </div>
                    `}
                    <div class="compose-actions ${searchMode ? "search-mode" : ""}">
                    ${showSessionSwitcherButton && fe`
                        <div
                            ref=${sessionTriggerRef}
                            class="compose-session-trigger-group"
                        >
                            ${currentSessionAgent?.agent_name && fe`
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
                    ${searchMode && fe`
                        <label class="compose-search-scope-wrap" title="Search scope">
                            <span class="compose-search-scope-label">Scope</span>
                            <select
                                class="compose-search-scope-select"
                                value=${searchScope}
                                onChange=${(e) => onSearchScopeChange?.(e.currentTarget.value)}
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
                        ${searchMode ? fe`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M18 6L6 18M6 6l12 12"/>
                            </svg>
                        ` : fe`
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="11" cy="11" r="8"/>
                                <path d="M21 21l-4.35-4.35"/>
                            </svg>
                        `}
                    </button>
                    ${canShareLocation && !searchMode && fe`
                        <button
                            class="icon-btn location-btn"
                            onClick=${handleLocation}
                            title="Share location"
                            type="button"
                        >
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="10" />
                                <path d="M12 2a14 14 0 0 1 0 20a14 14 0 0 1 0-20" />
                                <path d="M2 12h20" />
                            </svg>
                        </button>
                    `}
                    ${notificationsAvailable && !searchMode && fe`
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
                    ${!searchMode && fe`
                        ${activeEditorPath && onAttachEditorFile && fe`
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
                    ${(connectionStatus !== "connected" || !searchMode) && fe`
                        <div class="compose-send-stack">
                            ${connectionStatus !== "connected" && fe`
                                <span class="compose-connection-status connection-status ${connectionStatusPresentation.statusClass}" title=${connectionStatusTitle}>
                                    ${connectionStatusLabel}
                                </span>
                            `}
                            ${!searchMode && fe`
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
                                    ${submitButtonState.mode === "compacting" ? fe`
                                            <span class="compose-submit-spinner" aria-hidden="true">
                                                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                                                    <circle class="compose-submit-spinner-ring" cx="12" cy="12" r="10.5" stroke-width="2.25" stroke-linecap="round"></circle>
                                                    <rect class="compose-submit-spinner-stop" x="6" y="6" width="12" height="12" rx="0" fill="currentColor"></rect>
                                                </svg>
                                            </span>
                                        ` : submitButtonState.mode === "abort" ? fe`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><rect x="4" y="4" width="16" height="16" rx="2.5"/></svg>` : fe`<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/></svg>`}
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

// web/src/gi-preview-overflow.ts
function usePreviewOverflow(draft, thought, expanded) {
  const [overflow, setOverflow] = F_({ draft: false, thought: false });
  const nodes = Q_({ draft: null, thought: null });
  const schedule = Q_(() => {});
  const refs = Q_(null);
  if (!refs.current)
    refs.current = Object.fromEntries(["draft", "thought"].map((key) => [key, (node) => {
      nodes.current[key] = node;
      schedule.current();
    }]));
  W_(() => {
    let frame = 0, live = true;
    const measure = () => {
      frame = 0;
      if (!live)
        return;
      setOverflow((previous) => {
        const next = { ...previous };
        for (const key of ["draft", "thought"]) {
          const node = nodes.current[key];
          if (!node?.isConnected) {
            next[key] = false;
            continue;
          }
          if (node.closest(".agent-thinking")?.getAttribute("data-expanded") === "true")
            continue;
          if (!node.getClientRects().length)
            continue;
          next[key] = node.scrollHeight > node.clientHeight + 1;
        }
        return next.draft === previous.draft && next.thought === previous.thought ? previous : next;
      });
    };
    const enqueue = () => {
      if (live && !frame)
        frame = requestAnimationFrame(measure);
    };
    const observer = typeof ResizeObserver === "function" ? new ResizeObserver(enqueue) : null;
    const layoutObserver = !observer && typeof MutationObserver === "function" ? new MutationObserver(enqueue) : null;
    const observe = () => {
      observer?.disconnect();
      layoutObserver?.disconnect();
      const ancestors = new Set;
      for (const node of Object.values(nodes.current)) {
        if (!node)
          continue;
        observer?.observe(node);
        for (const child of node.children)
          observer?.observe(child);
        if (layoutObserver) {
          for (let parent = node.parentElement;parent; parent = parent.parentElement) {
            if (!ancestors.has(parent)) {
              ancestors.add(parent);
              layoutObserver.observe(parent, { attributes: true, attributeFilter: ["class", "style"] });
            }
            if (parent.classList.contains("app-shell"))
              break;
          }
        }
      }
      enqueue();
    };
    schedule.current = observe;
    observe();
    window.addEventListener("resize", enqueue);
    if (layoutObserver)
      window.addEventListener("transitionend", enqueue);
    document.fonts?.addEventListener("loadingdone", enqueue);
    return () => {
      live = false;
      schedule.current = () => {};
      observer?.disconnect();
      layoutObserver?.disconnect();
      if (frame)
        cancelAnimationFrame(frame);
      window.removeEventListener("resize", enqueue);
      if (layoutObserver)
        window.removeEventListener("transitionend", enqueue);
      document.fonts?.removeEventListener("loadingdone", enqueue);
    };
  }, []);
  W_(() => {
    schedule.current();
  }, [draft, thought, expanded]);
  return { overflow, refs: refs.current };
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
var COPY_ICON_SVG2 = fe`
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <rect x="9" y="9" width="10" height="10" rx="2"></rect>
        <path d="M7 15H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h7a2 2 0 0 1 2 2v1"></path>
    </svg>
`;
var GIT_BRANCH_ICON_SVG = fe`
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">
        <path d="M6 3v12"></path>
        <circle cx="18" cy="6" r="3"></circle>
        <circle cx="6" cy="18" r="3"></circle>
        <path d="M18 9a9 9 0 0 1-9 9"></path>
    </svg>
`;
var CLOCK_ICON_SVG = fe`
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
  const m = Math.floor(totalSec % 3600 / 60);
  const s = totalSec % 60;
  if (h > 0)
    return `${h}h ${m}m`;
  if (m > 0)
    return `${m}m ${s}s`;
  return `${s}s`;
}
function AgentStatus({ status, draft, plan, thought, pendingRequest, intent, extensionPanels = [], pendingPanelActions = new Set, onExtensionPanelAction, turnId, steerQueued, onPanelToggle, showCorePanels = true, showExtensionPanels = true }) {
  const THOUGHT_MAX_LINES = 8;
  const DRAFT_MAX_LINES = 8;
  const normalizePreview = (value) => {
    if (!value)
      return { text: "", totalLines: 0, fullText: "" };
    if (typeof value === "string") {
      const text = value;
      const totalLines = text ? text.replace(/\r\n/g, `
`).split(`
`).length : 0;
      return { text, totalLines, fullText: text };
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
      const totalLines = Number.isFinite(totalLinesOverride) ? totalLinesOverride : 0;
      return { text: "", omitted: 0, totalLines, visibleLines: 0 };
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
  const [expandedPanels, setExpandedPanels] = F_(new Set);
  const previewOverflow = usePreviewOverflow(draft, thought, expandedPanels);
  const [hoveredSeriesPoint, setHoveredSeriesPoint] = F_(null);
  const [nowMs, setNowMs] = F_(() => Date.now());
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
  K_(() => {
    setExpandedPanels(new Set);
    setHoveredSeriesPoint(null);
  }, [turnId]);
  K_(() => {
    const hasExpandedTimestampPanel = Array.isArray(extensionPanels) && extensionPanels.some((p) => expandedPanels.has(p?.key) && (p?.started_at || p?.last_activity_at));
    if (!hasExpandedTimestampPanel)
      return;
    const interval = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(interval);
  }, [expandedPanels, extensionPanels]);
  const escapeCollapseKey = u_(() => resolveAgentStatusEscapeCollapseKey(expandedPanels), [expandedPanels]);
  K_(() => {
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
  const shouldTickActivityAge = u_(() => shouldTickStatusActivityAge(status), [status]);
  const shouldTickIntentAge = u_(() => shouldTickIntentElapsed(status), [status]);
  const toolContextPath = u_(() => extractToolContextPath(status?.tool_name, status?.tool_args), [status?.tool_name, status?.tool_args]);
  const [toolRepoContext, setToolRepoContext] = F_(null);
  K_(() => {
    const shouldTick = Boolean(shouldTickIntentAge || status?.retry_at || status?.retryAt || shouldTickActivityAge);
    if (!shouldTick)
      return;
    setNowMs(Date.now());
    const timer = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [shouldTickActivityAge, shouldTickIntentAge, status?.retry_at, status?.retryAt, status?.last_event_at, status?.lastEventAt, status?.started_at, status?.startedAt, status?.type, status?.tool_name, status?.tool_args]);
  K_(() => {
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
  const orderedStatusHints = u_(() => orderAgentStatusHints(statusHints), [statusHints]);
  const leadingStatusHints = u_(() => orderedStatusHints.filter((hint) => hint?.key === "ssh"), [orderedStatusHints]);
  const trailingStatusHints = u_(() => orderedStatusHints.filter((hint) => hint?.key !== "ssh"), [orderedStatusHints]);
  if ((!showCorePanels || !hasCorePanels) && (!showExtensionPanels || !hasExtensionPanels))
    return null;
  const renderThinkingPanel = ({ panelTitle, text, fullText, totalLines, maxLines, titleClass, panelKey }) => {
    const isExpanded = expandedPanels.has(panelKey);
    const rawSourceText = fullText || text || "";
    const sourceText = panelKey === "thought" || panelKey === "draft" ? stripInternalTags(rawSourceText) : rawSourceText;
    const isCollapsible = typeof maxLines === "number";
    const showClose = isExpanded && isCollapsible;
    const truncated = isCollapsible ? truncateLines(sourceText, maxLines, totalLines) : { text: sourceText || "", omitted: 0, totalLines: Number.isFinite(totalLines) ? totalLines : 0 };
    if (!sourceText && !(Number.isFinite(truncated.totalLines) && truncated.totalLines > 0))
      return null;
    const canDisclose = truncated.omitted > 0 || previewOverflow.overflow[panelKey] === true;
    const bodyClass = `agent-thinking-body${isCollapsible ? " agent-thinking-body-collapsible" : ""}`;
    const bodyStyle = isCollapsible ? `--agent-thinking-collapsed-lines: ${maxLines};` : "";
    return fe`
            <div
                class="agent-thinking"
                data-expanded=${isExpanded ? "true" : "false"}
                data-collapsible=${isCollapsible ? "true" : "false"}
                style=${turnColor ? `--turn-color: ${turnColor};` : ""}
            >
                <div class="agent-thinking-title ${titleClass || ""}">
                    ${turnColor && fe`<span class=${dotClass} aria-hidden="true"></span>`}
                    ${panelTitle}
                    ${showClose && fe`
                        <button
                            class="agent-thinking-close"
                            aria-label=${`Close ${panelTitle} panel`}
                            onClick=${() => toggleExpand(panelKey)}
                        >
                            ×
                        </button>
                    `}
                </div>
                <div
                    ref=${previewOverflow.refs[panelKey]}
                    class=${bodyClass}
                    style=${bodyStyle}
                    dangerouslySetInnerHTML=${{ __html: renderThinkingMarkdown(sourceText) }}
                />
                ${!isExpanded && canDisclose && fe`
                    <button class="agent-thinking-truncation" onClick=${() => toggleExpand(panelKey)}>
                        ${truncated.omitted > 0 ? `▸ ${truncated.omitted} more lines` : "▸ Show more"}
                    </button>
                `}
                ${isExpanded && canDisclose && fe`
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
    return fe`
            <div
                class="agent-thinking agent-thinking-intent"
                aria-live="polite"
                style=${color ? `--turn-color: ${color};` : ""}
                title=${payload?.detail || ""}
            >
                <div class="agent-thinking-title intent">
                    ${color && fe`<span class=${pulsingDotClass} aria-hidden="true"></span>`}
                    <span class="agent-thinking-title-text">${titleText}</span>
                    ${metaLabel && fe`<span class="agent-status-elapsed">${metaLabel}</span>`}
                </div>
                ${payload.detail && fe`<div class="agent-thinking-body">${payload.detail}</div>`}
            </div>
        `;
  };
  const projectSeriesPoint = (point, width, height, minValue, maxValue, minRun, maxRun, paddingX = 8, paddingY = 8) => {
    const range = Math.max(maxValue - minValue, 0.000000001);
    const innerWidth = Math.max(width - paddingX * 2, 1);
    const innerHeight = Math.max(height - paddingY * 2, 1);
    const runSpan = Math.max(maxRun - minRun, 1);
    const x = maxRun === minRun ? width / 2 : paddingX + (point.run - minRun) / runSpan * innerWidth;
    const y = paddingY + (innerHeight - (point.value - minValue) / range * innerHeight);
    return { x, y };
  };
  const buildLinePath = (points, width, height, minValue, maxValue, minRun, maxRun, paddingX = 8, paddingY = 8) => {
    if (!Array.isArray(points) || points.length === 0)
      return "";
    return points.map((point, index) => {
      const { x, y } = projectSeriesPoint(point, width, height, minValue, maxValue, minRun, maxRun, paddingX, paddingY);
      return `${index === 0 ? "M" : "L"} ${x.toFixed(2)} ${y.toFixed(2)}`;
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
    return fe`
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
      return fe`
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
        return fe`
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
      return fe`
                            <div key=${`${seriesKey}-legend`} class=${`agent-series-legend-item${hovered ? " is-hovered" : ""}`} style=${`--agent-series-color: ${series.color};`}>
                                <span class="agent-series-legend-swatch" style=${`--agent-series-color: ${series.color};`}></span>
                                <span class="agent-series-legend-label">${series?.label || "Series"}</span>
                                ${hoveredRun !== null && fe`<span class="agent-series-legend-run">run ${hoveredRun}</span>`}
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
    return fe`
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
                        ${color && fe`<span class=${panelDotClass} aria-hidden="true"></span>`}
                        <span class="agent-thinking-title-text">${titleText}</span>
                        ${displayCollapsed && fe`<span class="agent-thinking-title-meta">${displayCollapsed}</span>`}
                    </button>
                    ${(actions.length > 0 || isExpandable) && fe`
                        <div class="agent-thinking-tools-inline">
                            ${actions.length > 0 && fe`
                                <div class="agent-thinking-actions agent-thinking-actions-inline">
                                    ${actions.map((action) => {
      const pendingKey = `${panelKey}:${action?.key || ""}`;
      const pending = pendingPanelActions?.has?.(pendingKey);
      return fe`
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
                            ${isExpandable && fe`
                                <button
                                    class="agent-thinking-corner-toggle agent-thinking-corner-toggle-inline"
                                    type="button"
                                    aria-label=${`${isExpanded ? "Collapse" : "Expand"} ${titleText}`}
                                    title=${isExpanded ? "Collapse details" : "Expand details"}
                                    onClick=${() => toggleExpand(panelKey)}
                                >
                                    <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        ${isExpanded ? fe`<polyline points="4 6 8 10 12 6"></polyline>` : fe`<polyline points="4 10 8 6 12 10"></polyline>`}
                                    </svg>
                                </button>
                            `}
                        </div>
                    `}
                </div>
                ${isExpanded && fe`
                    <div class=${`agent-thinking-autoresearch-layout${hasDetailColumn ? "" : " chart-only"}`}>
                        ${hasDetailColumn && fe`
                            <div class="agent-thinking-autoresearch-meta-stack">
                                ${experimentElapsed && fe`
                                    <div class="agent-thinking-autoresearch-elapsed">
                                        <span title="Experiment duration">⏱ ${experimentElapsed}</span>
                                        ${panel?.last_activity_at && panel?.state === "running" && fe`<span title="Since last activity">⟳ ${formatElapsed(panel.last_activity_at)} ago</span>`}
                                    </div>
                                `}
                                ${detailText && fe`
                                    <div
                                        class="agent-thinking-body agent-thinking-autoresearch-detail"
                                        dangerouslySetInnerHTML=${{ __html: renderThinkingMarkdown(detailText) }}
                                    />
                                `}
                                ${tmuxCommand && fe`
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
                        ${series.length > 0 ? fe`
                                <div class="agent-series-chart-stack">
                                    ${renderCombinedSeriesChart(series, panelKey)}
                                    ${lastRunText && fe`<div class="agent-series-chart-note">${lastRunText}</div>`}
                                </div>
                            ` : fe`<div class="agent-thinking-body agent-thinking-autoresearch-summary">Variable history will appear after the first completed run.</div>`}
                    </div>
                `}
            </div>
        `;
  };
  return fe`
        <div class="agent-status-panel">
            ${showCorePanels && intent && renderIntentPanel(intent, intentColor)}
            ${showExtensionPanels && Array.isArray(extensionPanels) && extensionPanels.map((panel) => renderExtensionPanel(panel))}
            ${showCorePanels && status?.type === "intent" && renderIntentPanel(status, statusIntentColor, statusIntentElapsedLabel)}
            ${showCorePanels && pendingRequest && fe`
                <div class="agent-status agent-status-request" aria-live="polite" style=${turnColor ? `--turn-color: ${turnColor};` : ""}>
                    ${pendingIndicatorMode === "dot" && fe`<span class=${dotClass} aria-hidden="true"></span>`}
                    ${pendingIndicatorMode === "spinner" && fe`<div class="agent-status-spinner"></div>`}
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
            ${showCorePanels && status && status?.type !== "intent" && fe`
                <div class=${`agent-status${isLastActivity ? " agent-status-last-activity" : ""}${status?.type === "error" ? " agent-status-error" : ""}${toolRepoLabel || statusHints.length > 0 || statusActivityAgeLabel ? " agent-status-multiline" : ""}`} aria-live="polite" style=${turnColor ? `--turn-color: ${turnColor};` : ""}>
                    ${turnColor && showRunningStatusDot && fe`<span class=${dotClass} aria-hidden="true"></span>`}
                    ${status?.type === "error" ? fe`<span class="agent-status-error-icon" aria-hidden="true">⚠</span>` : runningIndicatorMode === "spinner" && fe`<div class="agent-status-spinner"></div>`}
                    <div class="agent-status-copy">
                        <span class="agent-status-text">${content}</span>
                        ${(toolRepoLabel || orderedStatusHints.length > 0 || statusActivityAgeLabel) && fe`
                            <span class="agent-status-meta-row">
                                ${leadingStatusHints.map((hint) => fe`
                                    <span key=${hint.key} class="agent-status-hint-row" title=${hint.title || hint.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{ __html: hint.iconSvg }}></span>
                                        <span class="agent-status-hint-label">${hint.label}</span>
                                    </span>
                                `)}
                                ${toolRepoLabel && fe`
                                    <span class="agent-status-git-row" title=${toolContextPath || toolRepoLabel}>
                                        <span class="agent-status-git-icon">${GIT_BRANCH_ICON_SVG}</span>
                                        <span class="agent-status-git-label">
                                            ${toolRepoRepoPath && fe`<span class="agent-status-git-part">${toolRepoRepoPath}</span>`}
                                            ${toolRepoRepoPath && toolRepoBranch && fe`<span class="agent-status-git-separator" aria-hidden="true">•</span>`}
                                            ${toolRepoBranch && fe`<span class="agent-status-git-part">${toolRepoBranch}</span>`}
                                        </span>
                                    </span>
                                `}
                                ${trailingStatusHints.map((hint) => fe`
                                    <span key=${hint.key} class="agent-status-hint-row" title=${hint.title || hint.label}>
                                        <span class="agent-status-hint-icon" dangerouslySetInnerHTML=${{ __html: hint.iconSvg }}></span>
                                        <span class="agent-status-hint-label">${hint.label}</span>
                                    </span>
                                `)}
                                ${statusActivityAgeLabel && fe`
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
  const prevMap = prevKids ? new Map(prevKids.map((c) => [c?.path, c])) : new Map;
  let changed = !prevKids || prevKids.length !== nextKids.length;
  const merged = nextKids.map((child) => {
    const m = mergeTree(prevMap.get(child.path), child);
    if (m !== prevMap.get(child.path))
      changed = true;
    return m;
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
  entries.sort((a, b) => b.size - a.size);
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
function detectDarkTheme() {
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
  const [hovered, setHovered] = F_(null);
  const [zoomPath, setZoomPath] = F_(payload?.root?.path || ".");
  const [zoomStack, setZoomStack] = F_(() => [payload?.root?.path || "."]);
  const [isZooming, setIsZooming] = F_(false);
  K_(() => {
    const rootPath = payload?.root?.path || ".";
    setZoomPath(rootPath);
    setZoomStack([rootPath]);
    setHovered(null);
  }, [payload?.root?.path, payload?.totalSize]);
  K_(() => {
    if (!zoomPath)
      return;
    setIsZooming(true);
    const timer = setTimeout(() => setIsZooming(false), 180);
    return () => clearTimeout(timer);
  }, [zoomPath]);
  const zoomRoot = u_(() => {
    return findHierarchyNode(payload.root, zoomPath) || payload.root;
  }, [payload?.root, zoomPath]);
  const baseSize = zoomRoot?.size || payload.totalSize || 0;
  const { segments, legend } = u_(() => {
    const computed = buildStarburstSegments(zoomRoot, baseSize, payload.isDarkTheme);
    if (computed.segments.length > 0)
      return computed;
    if (baseSize <= 0)
      return computed;
    const label = zoomRoot?.children?.length ? "Total" : "[files]";
    return buildFallbackStarburst(label, zoomRoot?.path || payload?.root?.path || ".", baseSize, payload.isDarkTheme);
  }, [zoomRoot, baseSize, payload.isDarkTheme, payload?.root?.path]);
  const [animatedSegments, setAnimatedSegments] = F_(segments);
  const prevSegmentsRef = Q_(new Map);
  const animFrameRef = Q_(0);
  K_(() => {
    const prevMap = prevSegmentsRef.current;
    const nextMap = new Map(segments.map((segment) => [segment.key, segment]));
    const start = performance.now();
    const duration = 220;
    const animate = (now) => {
      const t = Math.min(1, (now - start) / duration);
      const eased = t * (2 - t);
      const interpolated = segments.map((segment) => {
        const prev = prevMap.get(segment.key);
        const from = prev || {
          startAngle: segment.startAngle,
          endAngle: segment.startAngle,
          innerRadius: segment.innerRadius,
          outerRadius: segment.innerRadius
        };
        const lerp = (a, b) => a + (b - a) * eased;
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
      if (t < 1) {
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
  return fe`
        <div class="workspace-folder-starburst">
            <svg viewBox="0 0 240 240" class=${`workspace-folder-starburst-svg${isZooming ? " is-zooming" : ""}`} role="img"
                aria-label=${`Folder sizes for ${zoomRoot?.path || payload?.root?.path || "."}`}
                data-segments=${displaySegments.length}
                data-base-size=${baseSize}>
                ${displaySegments.map((segment) => fe`
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
            ${legend.length > 0 && fe`
                <div class="workspace-folder-starburst-legend">
                    ${legend.slice(0, 8).map((entry) => fe`
                        <div key=${entry.key} class="workspace-folder-starburst-legend-item">
                            <span class="workspace-folder-starburst-swatch" style=${`background:${entry.color}`}></span>
                            <span class="workspace-folder-starburst-name" title=${entry.name}>${entry.name}</span>
                            <span class="workspace-folder-starburst-size">${formatFileSize(entry.size)}</span>
                            <span class="workspace-folder-starburst-pct">${entry.pct.toFixed(1)}%</span>
                        </div>
                    `)}
                </div>
            `}
            ${payload.truncated && fe`
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
  const [tree, setTree] = F_(null);
  const [expanded, setExpanded] = F_(new Set(["."]));
  const [selectedPath, setSelectedPath] = F_(null);
  const [renamingPath, setRenamingPath] = F_(null);
  const [renameValue, setRenameValue] = F_("");
  const [preview, setPreview] = F_(null);
  const [, setDownloadId] = F_(null);
  const [initialLoad, setInitialLoad] = F_(true);
  const [loadingPreview, setLoadingPreview] = F_(false);
  const [error, setError] = F_(null);
  const [showHidden, setShowHidden] = F_(() => getLocalStorageBoolean("workspaceShowHidden", false));
  const [dragActive, setDragActive] = F_(false);
  const [dragMode, setDragMode] = F_(null);
  const [dragGhost, setDragGhost] = F_(null);
  const [dropTarget, setDropTarget] = F_(null);
  const [uploading, setUploading] = F_(false);
  const [uploadProgress, setUploadProgress] = F_(null);
  const [folderChart, setFolderChart] = F_(null);
  const [workspaceIndexStatus, setWorkspaceIndexStatus] = F_(null);
  const [workspaceReindexing, setWorkspaceReindexing] = F_(false);
  const [isDarkTheme, setIsDarkTheme] = F_(() => detectDarkTheme());
  const [explorerScale, setExplorerScale] = F_(() => resolveWorkspaceScale({
    stored: getLocalStorageItem(WORKSPACE_SCALE_STORAGE_KEY),
    ...readWorkspaceScaleEnvironment()
  }));
  const [headerMenuOpen, setHeaderMenuOpen] = F_(false);
  const expandedRef = Q_(expanded);
  const lastSigRef = Q_("");
  const pendingRootRef = Q_(null);
  const rafRef = Q_(0);
  const pendingSubtreeRef = Q_(new Set);
  const loadTreeFnRef = Q_(null);
  const loadWorkspaceIndexStatusRef = Q_(null);
  const nodeMapRef = Q_(new Map);
  const onFileSelectRef = Q_(onFileSelect);
  const onOpenEditorRef = Q_(onOpenEditor);
  const loadPreviewRef = Q_(null);
  const loadSubtreeRef = Q_(null);
  const sidebarRef = Q_(null);
  const treeListRef = Q_(null);
  const renameInputRef = Q_(null);
  const uploadInputRef = Q_(null);
  const uploadTargetRef = Q_(".");
  const uploadProgressTimerRef = Q_(0);
  const touchDragRef = Q_({ path: null, dragging: false, startX: 0, startY: 0 });
  const mouseDragRef = Q_({ path: null, dragging: false, startX: 0, startY: 0 });
  const dragExpandRef = Q_({ path: null, timer: 0 });
  const suppressClickRef = Q_(false);
  const previewHeightRef = Q_(0);
  const folderChartCacheRef = Q_(new Map);
  const folderChartPayloadRef = Q_(null);
  const folderChartPathRef = Q_(null);
  const previewPaneHostRef = Q_(null);
  const previewPaneInstanceRef = Q_(null);
  const headerMenuRef = Q_(null);
  const headerMenuButtonRef = Q_(null);
  const showHiddenRef = Q_(showHidden);
  const visibleRef = Q_(visible);
  const activeRef = Q_(active ?? visible);
  const dragDepthRef = Q_(0);
  const dropTargetRef = Q_(dropTarget);
  const dragActiveRef = Q_(dragActive);
  const dragModeRef = Q_(dragMode);
  const dragGhostRef = Q_(null);
  const dragGhostPosRef = Q_({ x: 0, y: 0 });
  const dragGhostRafRef = Q_(0);
  const moveEntryToTargetRef = Q_(null);
  const selectedPathRef = Q_(selectedPath);
  const renamingPathRef = Q_(renamingPath);
  const pendingProgrammaticFileClickRef = Q_(null);
  const previewRef = Q_(preview);
  onFileSelectRef.current = onFileSelect;
  onOpenEditorRef.current = onOpenEditor;
  K_(() => {
    expandedRef.current = expanded;
  }, [expanded]);
  K_(() => {
    showHiddenRef.current = showHidden;
  }, [showHidden]);
  K_(() => {
    visibleRef.current = visible;
  }, [visible]);
  K_(() => {
    activeRef.current = active ?? visible;
  }, [active, visible]);
  K_(() => {
    dropTargetRef.current = dropTarget;
  }, [dropTarget]);
  const clearUploadProgressTimer = Y_(() => {
    if (!uploadProgressTimerRef.current)
      return;
    clearTimeout(uploadProgressTimerRef.current);
    uploadProgressTimerRef.current = 0;
  }, []);
  K_(() => () => clearUploadProgressTimer(), [clearUploadProgressTimer]);
  K_(() => {
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
  K_(() => {
    const handleReveal = (e) => {
      const path = e?.detail?.path;
      if (!path)
        return;
      const parts = path.split("/");
      const parents = [];
      for (let i = 1;i < parts.length; i++) {
        parents.push(parts.slice(0, i).join("/"));
      }
      if (parents.length) {
        setExpanded((prev) => {
          const next = new Set(prev);
          next.add(".");
          for (const p of parents)
            next.add(p);
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
  K_(() => {
    dragActiveRef.current = dragActive;
  }, [dragActive]);
  K_(() => {
    dragModeRef.current = dragMode;
  }, [dragMode]);
  K_(() => {
    selectedPathRef.current = selectedPath;
  }, [selectedPath]);
  K_(() => {
    renamingPathRef.current = renamingPath;
  }, [renamingPath]);
  K_(() => {
    previewRef.current = preview;
  }, [preview]);
  K_(() => {
    if (typeof window === "undefined" || typeof document === "undefined")
      return;
    const syncTheme = () => setIsDarkTheme(detectDarkTheme());
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
  K_(() => {
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
  K_(() => {
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
    } catch (err) {
      const failure = { error: err.message || "Failed to load preview" };
      setPreview(failure);
      return failure;
    } finally {
      setLoadingPreview(false);
    }
  };
  loadPreviewRef.current = loadPreview;
  const loadWorkspaceIndexStatus = Y_(async () => {
    try {
      const status = await getWorkspaceIndexStatus("all");
      setWorkspaceIndexStatus(status);
      return status;
    } catch (err) {
      console.warn("[workspace-explorer] Failed to load workspace index status:", err);
      return null;
    }
  }, []);
  loadWorkspaceIndexStatusRef.current = loadWorkspaceIndexStatus;
  const refreshWorkspaceIndexStatus = Y_(() => {
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
    } catch (err) {
      setError(err.message || "Failed to load workspace");
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
    } catch (err) {
      setError(err.message || "Failed to load workspace");
    } finally {
      pendingSubtreeRef.current.delete(path);
    }
  };
  loadSubtreeRef.current = loadSubtree;
  const resolveDropTargetPath = Y_(() => {
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
  const resolveDropTargetFromElement = Y_((element) => {
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
  const resolveDropTargetFromEvent = Y_((event) => {
    return resolveDropTargetFromElement(event?.target || null);
  }, [resolveDropTargetFromElement]);
  const updateDropTarget = Y_((value) => {
    dropTargetRef.current = value;
    setDropTarget(value);
  }, []);
  const clearDragExpandTimer = Y_(() => {
    const current = dragExpandRef.current;
    if (current?.timer)
      clearTimeout(current.timer);
    dragExpandRef.current = { path: null, timer: 0 };
  }, []);
  const scheduleDragExpand = Y_((targetPath) => {
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
  const updateDragGhostPosition = Y_((x, y) => {
    dragGhostPosRef.current = { x, y };
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
  const startDragGhost = Y_((path) => {
    if (!path)
      return;
    const node = nodeMapRef.current?.get(path);
    const label = (node?.name || path.split("/").pop() || path).trim();
    if (!label)
      return;
    setDragGhost({ path, label });
  }, []);
  const clearDragGhost = Y_(() => {
    setDragGhost(null);
    if (dragGhostRafRef.current) {
      cancelAnimationFrame(dragGhostRafRef.current);
      dragGhostRafRef.current = 0;
    }
    if (dragGhostRef.current) {
      dragGhostRef.current.style.transform = "translate(-9999px, -9999px)";
    }
  }, []);
  const resolveCreateTargetPath = Y_((path) => {
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
  const cancelRename = Y_(() => {
    setRenamingPath(null);
    setRenameValue("");
  }, []);
  const beginRename = Y_((path) => {
    if (!path)
      return;
    const node = nodeMapRef.current?.get(path);
    const base = (node?.name || path.split("/").pop() || path).trim();
    if (!base || path === ".")
      return;
    setRenamingPath(path);
    setRenameValue(base);
  }, []);
  const commitRename = Y_(async () => {
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
    } catch (err) {
      setError(err?.message || "Failed to rename file");
    }
  }, [cancelRename, renameValue, refreshWorkspaceIndexStatus]);
  const createUntitledFile = Y_(async (targetPath) => {
    const base = "untitled";
    const ext = ".md";
    const folder = targetPath || ".";
    for (let i = 0;i < 50; i += 1) {
      const suffix = i === 0 ? "" : `-${i}`;
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
      } catch (err) {
        if (err?.status === 409 || err?.code === "file_exists") {
          continue;
        }
        setError(err?.message || "Failed to create file");
        return;
      }
    }
    setError("Failed to create file (untitled name already in use).");
  }, []);
  const handleCreateFileClick = Y_((event) => {
    event?.stopPropagation?.();
    if (uploading)
      return;
    const target = resolveCreateTargetPath(selectedPathRef.current);
    createUntitledFile(target);
  }, [uploading, resolveCreateTargetPath, createUntitledFile]);
  K_(() => {
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
  const updateVisibility = Q_(() => {
    if (typeof window === "undefined")
      return;
    const media = window.matchMedia("(min-width: 1024px) and (orientation: landscape)");
    const active = activeRef.current ?? visibleRef.current;
    const visible = document.visibilityState !== "hidden" && (active || media.matches && visibleRef.current);
    setWorkspaceVisibility(visible, showHiddenRef.current).catch((error) => {
      console.debug("[workspace-explorer] Workspace visibility ping failed.", error, {
        visible,
        showHidden: showHiddenRef.current
      });
    });
  }).current;
  const debouncedVisibilityRef = Q_(0);
  const scheduleVisibilityUpdate = Q_(() => {
    if (debouncedVisibilityRef.current) {
      clearTimeout(debouncedVisibilityRef.current);
    }
    debouncedVisibilityRef.current = setTimeout(() => {
      debouncedVisibilityRef.current = 0;
      updateVisibility();
    }, 250);
  }).current;
  K_(() => {
    if (visibleRef.current) {
      loadTreeFnRef.current?.();
      loadWorkspaceIndexStatusRef.current?.();
    }
    scheduleVisibilityUpdate();
  }, [visible, active]);
  K_(() => {
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
      setWorkspaceVisibility(false, showHiddenRef.current).catch((error) => {
        console.debug("[workspace-explorer] Workspace visibility teardown ping failed.", error, {
          showHidden: showHiddenRef.current
        });
      });
    };
  }, []);
  const rows = u_(() => flattenTree(tree, expanded, showHidden), [tree, expanded, showHidden]);
  const nodeMap = u_(() => new Map(rows.map((r) => [r.node.path, r.node])), [rows]);
  const workspaceScaleMetrics = u_(() => getWorkspaceScaleMetrics(explorerScale), [explorerScale]);
  nodeMapRef.current = nodeMap;
  const selectedNode = selectedPath ? nodeMapRef.current.get(selectedPath) : null;
  const selectedIsDir = selectedNode?.type === "dir";
  K_(() => {
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
    }).catch((err) => {
      if (selectedPathRef.current !== fetchPath)
        return;
      setFolderChart({ loading: false, error: err?.message || "Failed to load folder size chart", payload: lastPath === selectedPath ? lastPayload : null });
    });
  }, [selectedPath, selectedIsDir, showHidden, isDarkTheme]);
  const canEdit = Boolean(preview && preview.kind === "text" && !selectedIsDir && (!preview.size || preview.size <= 256 * 1024));
  const editTitle = canEdit ? "Open read-only tab" : preview?.size > 256 * 1024 ? "File too large for a read-only tab" : "File preview only";
  const selectedHasOpenableTab = Boolean(selectedPath && !selectedIsDir && hasOpenableWorkspaceTab(selectedPath));
  const selectedCanRename = Boolean(selectedPath && selectedPath !== ".");
  const selectedCanDelete = Boolean(selectedPath && !selectedIsDir);
  const selectedCanDownload = Boolean(selectedPath && !selectedIsDir);
  const selectedFolderDownloadUrl = selectedPath && selectedIsDir ? getWorkspaceDownloadUrl(selectedPath, showHidden) : null;
  const workspaceIndexLabel = describeWorkspaceIndexState(workspaceIndexStatus);
  const workspaceIndexTitle = buildWorkspaceIndexTitle(workspaceIndexStatus);
  const workspaceIndexState = workspaceIndexStatus?.state || "never_indexed";
  const showWorkspaceIndexIndicator = workspaceIndexState !== "ready";
  const closeHeaderMenu = Y_(() => setHeaderMenuOpen(false), []);
  const runMenuAction = Y_(async (fn) => {
    closeHeaderMenu();
    try {
      await fn?.();
    } catch (err) {
      console.warn("[workspace-explorer] Header menu action failed:", err);
    }
  }, [closeHeaderMenu]);
  const handleWorkspaceReindex = Y_(async (event) => {
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
    } catch (err) {
      const message = err?.message || "Failed to reindex workspace";
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
  K_(() => {
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
  const handleTreeDblClick = Q_((e) => {
    const targetEl = getEventTargetElement(e);
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
  const handleTreeClick = Q_((e) => {
    if (suppressClickRef.current) {
      suppressClickRef.current = false;
      return;
    }
    const targetEl = getEventTargetElement(e);
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
  const handleRefreshClick = Q_(() => {
    lastSigRef.current = "";
    loadTreeFnRef.current();
    loadWorkspaceIndexStatusRef.current?.();
    const openPaths = Array.from(expandedRef.current || []).filter((p) => p && p !== ".");
    openPaths.forEach((p) => loadSubtreeRef.current?.(p));
  }).current;
  const clearSelection = Q_(() => {
    pendingProgrammaticFileClickRef.current = null;
    setSelectedPath(null);
    setPreview(null);
    setDownloadId(null);
    setLoadingPreview(false);
  }).current;
  const handleToggleHidden = Q_(() => {
    setShowHidden((prev) => {
      const next = !prev;
      if (typeof window !== "undefined") {
        setLocalStorageItem("workspaceShowHidden", String(next));
      }
      showHiddenRef.current = next;
      setWorkspaceVisibility(true, next).catch((error) => {
        console.debug("[workspace-explorer] Workspace visibility refresh after toggling hidden files failed.", error, {
          showHidden: next
        });
      });
      lastSigRef.current = "";
      loadTreeFnRef.current?.();
      const openPaths = Array.from(expandedRef.current || []).filter((p) => p && p !== ".");
      openPaths.forEach((p) => loadSubtreeRef.current?.(p));
      return next;
    });
  }).current;
  const handleBackgroundClick = Q_((e) => {
    const targetEl = getEventTargetElement(e);
    if (targetEl?.closest?.("[data-path]"))
      return;
    clearSelection();
  }).current;
  const deleteFileAtPath = Y_(async (path) => {
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
    } catch (err) {
      setPreview((prev) => ({ ...prev || {}, error: err.message || "Failed to delete file" }));
    }
  }, [clearSelection]);
  const scrollRowIntoView = Y_((path) => {
    const container = treeListRef.current;
    if (!container || !path || typeof CSS === "undefined" || typeof CSS.escape !== "function")
      return;
    const el = container.querySelector(`[data-path="${CSS.escape(path)}"]`);
    el?.scrollIntoView?.({ block: "nearest" });
  }, []);
  const handleTreeKeyDown = Y_((e) => {
    const targetEl = getEventTargetElement(e);
    if (renamingPathRef.current || isEditableKeyboardTarget(targetEl))
      return;
    const currentRows = rows;
    if (!currentRows || currentRows.length === 0)
      return;
    const currentIndex = selectedPath ? currentRows.findIndex((r) => r.node.path === selectedPath) : -1;
    if (e.key === "ArrowDown") {
      e.preventDefault();
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
    if (e.key === "ArrowUp") {
      e.preventDefault();
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
    if (e.key === "ArrowRight" && currentIndex >= 0) {
      const row = currentRows[currentIndex];
      if (row?.node?.type === "dir" && !expanded.has(row.node.path)) {
        e.preventDefault();
        loadSubtreeRef.current?.(row.node.path);
        setExpanded((prev) => new Set([...prev, row.node.path]));
      }
      return;
    }
    if (e.key === "ArrowLeft" && currentIndex >= 0) {
      const row = currentRows[currentIndex];
      if (row?.node?.type === "dir" && expanded.has(row.node.path)) {
        e.preventDefault();
        setExpanded((prev) => {
          const next = new Set(prev);
          next.delete(row.node.path);
          return next;
        });
      }
      return;
    }
    if (e.key === "Enter" && currentIndex >= 0) {
      e.preventDefault();
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
    if ((e.key === "Delete" || e.key === "Backspace") && currentIndex >= 0) {
      const row = currentRows[currentIndex];
      if (!row || row.node.type === "dir")
        return;
      e.preventDefault();
      deleteFileAtPath(row.node.path);
      return;
    }
    if (e.key === "Escape") {
      e.preventDefault();
      clearSelection();
    }
  }, [clearSelection, deleteFileAtPath, expanded, rows, scrollRowIntoView, selectedPath]);
  const handleRowTouchStart = Y_((event) => {
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
  const handleRowTouchEnd = Y_(() => {
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
  const handleRowTouchMove = Y_((event) => {
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
  const handlePreviewSplitterMouseDown = Q_((e) => {
    e.preventDefault();
    const sidebar = sidebarRef.current;
    if (!sidebar)
      return;
    const startY = e.clientY;
    const startH = previewHeightRef.current || 280;
    const splitter = e.currentTarget;
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
  const handlePreviewSplitterTouchStart = Q_((e) => {
    e.preventDefault();
    const sidebar = sidebarRef.current;
    if (!sidebar)
      return;
    const touch = e.touches[0];
    if (!touch)
      return;
    const startY = touch.clientY;
    const startH = previewHeightRef.current || 280;
    const splitter = e.currentTarget;
    splitter.classList.add("dragging");
    document.body.style.userSelect = "none";
    const onMove = (te) => {
      const t = te.touches[0];
      if (!t)
        return;
      te.preventDefault();
      const maxH = sidebar.clientHeight - 80;
      const h = Math.min(Math.max(startH - (t.clientY - startY), 80), maxH);
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
  const handleDownload = Y_((path = selectedPath) => {
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
  const handleDragEnter = Y_((event) => {
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
  const handleDragOver = Y_((event) => {
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
  const handleDragLeave = Y_((event) => {
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
  const uploadFilesToTarget = Y_(async (files, targetPath = ".") => {
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
      for (let i = 0;i < list.length; i++) {
        const file = list[i];
        const name = file?.name || `file ${i + 1}`;
        setUploadProgress((prev) => ({ ...prev, current: i + 1, name, percent: 0 }));
        const onProgress = (p) => setUploadProgress((prev) => ({ ...prev, percent: p.percent }));
        try {
          lastResult = await uploadWorkspaceFile(file, target, { onProgress });
        } catch (err) {
          const status = err?.status;
          const code = err?.code;
          if (status === 409 || code === "file_exists") {
            const confirmOverwrite = window.confirm(`"${name}" already exists in ${targetLabel}. Overwrite?`);
            if (!confirmOverwrite)
              continue;
            lastResult = await uploadWorkspaceFile(file, target, { overwrite: true, onProgress });
          } else {
            throw err;
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
    } catch (err) {
      setError(err.message || "Failed to upload file");
      setUploadProgress((prev) => prev ? { ...prev, error: err.message || "Upload failed" } : null);
      clearUploadProgressTimer();
      uploadProgressTimerRef.current = window.setTimeout(() => {
        uploadProgressTimerRef.current = 0;
        setUploadProgress(null);
      }, 4000);
    } finally {
      setUploading(false);
    }
  }, [clearUploadProgressTimer]);
  const moveEntryToTarget = Y_(async (sourcePath, targetPath) => {
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
    } catch (err) {
      setError(err?.message || "Failed to move entry");
    }
  }, []);
  moveEntryToTargetRef.current = moveEntryToTarget;
  const handleDrop = Y_(async (event) => {
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
  const handleFolderUploadClick = Y_((event) => {
    event?.stopPropagation?.();
    if (uploading)
      return;
    const target = event?.currentTarget?.dataset?.uploadTarget || ".";
    uploadTargetRef.current = target;
    uploadInputRef.current?.click();
  }, [uploading]);
  const handleUploadButtonClick = Y_(() => {
    if (uploading)
      return;
    const selected = selectedPathRef.current;
    const selectedNode = selected ? nodeMapRef.current?.get(selected) : null;
    uploadTargetRef.current = selectedNode?.type === "dir" ? selectedNode.path : ".";
    uploadInputRef.current?.click();
  }, [uploading]);
  const handleMenuCreateFile = Y_(() => {
    runMenuAction(() => handleCreateFileClick(null));
  }, [runMenuAction, handleCreateFileClick]);
  const handleMenuUploadFiles = Y_(() => {
    runMenuAction(() => handleUploadButtonClick());
  }, [runMenuAction, handleUploadButtonClick]);
  const handleMenuRefresh = Y_(() => {
    runMenuAction(() => handleRefreshClick());
  }, [runMenuAction, handleRefreshClick]);
  const handleMenuToggleHidden = Y_(() => {
    runMenuAction(() => handleToggleHidden());
  }, [runMenuAction, handleToggleHidden]);
  const handleMenuOpenTab = Y_(() => {
    if (!selectedPath || !selectedHasOpenableTab)
      return;
    runMenuAction(() => onOpenEditorRef.current?.(selectedPath, preview));
  }, [runMenuAction, selectedPath, selectedHasOpenableTab, preview]);
  const handleMenuOpenEditor = Y_(() => {
    if (!selectedPath || !canEdit)
      return;
    runMenuAction(() => onOpenEditorRef.current?.(selectedPath, preview));
  }, [runMenuAction, selectedPath, canEdit, preview]);
  const handleMenuRename = Y_(() => {
    if (!selectedPath || selectedPath === ".")
      return;
    runMenuAction(() => beginRename(selectedPath));
  }, [runMenuAction, selectedPath, beginRename]);
  const handleMenuDelete = Y_(() => {
    if (!selectedPath || selectedIsDir)
      return;
    runMenuAction(() => handleDeleteFile());
  }, [runMenuAction, selectedPath, selectedIsDir, handleDeleteFile]);
  const handleMenuDownload = Y_(() => {
    if (!selectedPath || selectedIsDir)
      return;
    runMenuAction(() => handleDownload());
  }, [runMenuAction, selectedPath, selectedIsDir, handleDownload]);
  const handleMenuDownloadFolder = Y_(() => {
    if (!selectedFolderDownloadUrl)
      return;
    closeHeaderMenu();
    triggerWorkspaceDownload(selectedFolderDownloadUrl);
  }, [closeHeaderMenu, selectedFolderDownloadUrl]);
  const handleMenuOpenTerminalTab = Y_(() => {
    closeHeaderMenu();
    onOpenTerminalTab?.();
  }, [closeHeaderMenu, onOpenTerminalTab]);
  const handleMenuOpenVncTab = Y_(() => {
    closeHeaderMenu();
    onOpenVncTab?.();
  }, [closeHeaderMenu, onOpenVncTab]);
  const handleMenuToggleTerminal = Y_(() => {
    closeHeaderMenu();
    onToggleTerminal?.();
  }, [closeHeaderMenu, onToggleTerminal]);
  const handleRowMouseDown = Y_((event) => {
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
  const handleUploadInputChange = Y_(async (event) => {
    const files = Array.from(event?.target?.files || []);
    if (files.length === 0)
      return;
    const target = uploadTargetRef.current || ".";
    await uploadFilesToTarget(files, target);
    uploadTargetRef.current = ".";
    if (event?.target)
      event.target.value = "";
  }, [uploadFilesToTarget]);
  return fe`
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
                            onClick=${(e) => {
    e.stopPropagation();
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
                        ${headerMenuOpen && fe`
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

                                ${(onOpenTerminalTab || onOpenVncTab || onToggleTerminal) && fe`<div class="workspace-menu-separator"></div>`}
                                ${onOpenTerminalTab && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenTerminalTab}>
                                        Open terminal in tab
                                    </button>
                                `}
                                ${onOpenVncTab && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenVncTab}>
                                        Open VNC in tab
                                    </button>
                                `}
                                ${onToggleTerminal && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuToggleTerminal}>
                                        ${terminalVisible ? "Hide terminal dock" : "Show terminal dock"}
                                    </button>
                                `}

                                ${selectedPath && fe`<div class="workspace-menu-separator"></div>`}
                                ${selectedHasOpenableTab && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenTab}>Open in tab</button>
                                `}
                                ${selectedPath && !selectedIsDir && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuOpenEditor} disabled=${!canEdit}>Open read-only tab</button>
                                `}
                                ${selectedCanRename && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuRename}>Rename selected</button>
                                `}
                                ${selectedCanDownload && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuDownload}>Download selected file</button>
                                `}
                                ${selectedFolderDownloadUrl && fe`
                                    <button class="workspace-menu-item" role="menuitem" onClick=${handleMenuDownloadFolder}>Download selected folder (zip)</button>
                                `}
                                ${selectedCanDelete && fe`
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
            ${showWorkspaceIndexIndicator && fe`
                <div class="workspace-index-status-row">
                    <div class=${`workspace-index-status-chip state-${workspaceIndexState}`} title=${workspaceIndexTitle}>
                        <span class="workspace-index-status-dot" aria-hidden="true"></span>
                        <span>${workspaceIndexLabel}</span>
                    </div>
                </div>
            `}
            <div class="workspace-tree" onClick=${handleBackgroundClick}>
                ${uploadProgress && fe`
                    <div class="workspace-upload-strip">
                        <div class="workspace-upload-strip-text">
                            ${uploadProgress.error ? fe`<span class="workspace-upload-strip-error">${uploadProgress.error}</span>` : uploadProgress.done ? fe`<span>Done</span>` : fe`<span>${uploadProgress.total > 1 ? `Uploading ${uploadProgress.current}/${uploadProgress.total}: ${uploadProgress.name}` : `Uploading ${uploadProgress.name}`}${uploadProgress.percent > 0 ? ` (${uploadProgress.percent}%)` : "…"}</span>`}
                        </div>
                        ${!uploadProgress.done && !uploadProgress.error && fe`
                            <div class="workspace-upload-strip-bar">
                                <div class="workspace-upload-strip-fill" style=${`width:${uploadProgress.percent || 0}%`}></div>
                            </div>
                        `}
                    </div>
                `}
                ${initialLoad && fe`<div class="workspace-loading">Loading…</div>`}
                ${error && fe`<div class="workspace-error">${error}</div>`}
                ${tree && fe`
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
    return fe`
                                <div
                                    key=${node.path}
                                    class=${`workspace-row${isSelected ? " selected" : ""}${isDropTarget ? " drop-target" : ""}`}
                                    style=${{ paddingLeft: `${8 + depth * workspaceScaleMetrics.indentPx}px` }}
                                    data-path=${node.path}
                                    data-type=${node.type}
                                    onMouseDown=${handleRowMouseDown}
                                >
                                    <span class="workspace-caret" aria-hidden="true">
                                        ${isDir ? isOpen ? fe`<svg viewBox="0 0 12 12"><polygon points="1,2 11,2 6,11"/></svg>` : fe`<svg viewBox="0 0 12 12"><polygon points="2,1 11,6 2,11"/></svg>` : null}
                                    </span>
                                    <svg class=${`workspace-node-icon${isDir ? " folder" : ""}`}
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                        aria-hidden="true">
                                        ${isDir ? fe`<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>` : fe`<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>`}
                                    </svg>
                                    ${isRenaming ? fe`
                                            <input
                                                class="workspace-rename-input"
                                                ref=${renameInputRef}
                                                value=${renameValue}
                                                onInput=${(e) => setRenameValue(e?.target?.value || "")}
                                                onKeyDown=${(e) => {
      e.stopPropagation();
      if (e.key === "Enter") {
        e.preventDefault();
        commitRename();
      } else if (e.key === "Escape") {
        e.preventDefault();
        cancelRename();
      }
    }}
                                                onBlur=${cancelRename}
                                                onClick=${(e) => e.stopPropagation()}
                                            />
                                        ` : fe`<span class="workspace-label"><span class="workspace-label-text">${node.name}</span></span>`}
                                    ${isDir && !isOpen && childCount > 0 && fe`
                                        <span class="workspace-count">${childCount}</span>
                                    `}
                                    ${isDir && fe`
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
            ${selectedPath && fe`
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
                            ${!selectedIsDir && fe`
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
                            ${selectedIsDir ? fe`
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
                                        title="Download folder as zip" onClick=${(e) => e.stopPropagation()}>
                                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                            <polyline points="7 10 12 15 17 10"/>
                                            <line x1="12" y1="15" x2="12" y2="3"/>
                                        </svg>
                                    </a>` : fe`<a class="workspace-download" href=${getWorkspaceFileDownloadUrl(selectedPath)} download
                                        title="Download" onClick=${(e) => e.stopPropagation()}>
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                                        <polyline points="7 10 12 15 17 10"/>
                                        <line x1="12" y1="15" x2="12" y2="3"/>
                                    </svg>
                                </a>`}
                        </div>
                    </div>
                    ${loadingPreview && fe`<div class="workspace-loading">Loading preview…</div>`}
                    ${preview?.error && fe`<div class="workspace-error">${preview.error}</div>`}
                    ${selectedIsDir && fe`
                        <div class="workspace-preview-text">Folder selected — create file, upload files, or download as zip.</div>
                        ${folderChart?.loading && fe`<div class="workspace-loading">Loading folder size preview…</div>`}
                        ${folderChart?.error && fe`<div class="workspace-error">${folderChart.error}</div>`}
                        ${folderChart?.payload && folderChart.payload.segments?.length > 0 && fe`
                            <${FolderStarburstChart} payload=${folderChart.payload} />
                        `}
                        ${folderChart?.payload && (!folderChart.payload.segments || folderChart.payload.segments.length === 0) && fe`
                            <div class="workspace-preview-text">No file size data available for this folder yet.</div>
                        `}
                    `}
                    ${preview && !preview.error && !selectedIsDir && fe`
                        <div class="workspace-preview-body" ref=${previewPaneHostRef}></div>
                    `}
                </div>
            `}
            ${dragGhost && fe`
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
  const [contextMenu, setContextMenu] = F_(null);
  const stripRef = Q_(null);
  K_(() => {
    if (!contextMenu)
      return;
    const dismiss = (e) => {
      if (e.type === "keydown" && e.key !== "Escape")
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
  K_(() => {
    const onKeyDown = (e) => {
      if (e.ctrlKey && e.key === "Tab") {
        e.preventDefault();
        if (!tabs.length)
          return;
        const idx = tabs.findIndex((t) => t.id === activeId);
        if (e.shiftKey) {
          const prev = tabs[(idx - 1 + tabs.length) % tabs.length];
          onActivate?.(prev.id);
        } else {
          const next = tabs[(idx + 1) % tabs.length];
          onActivate?.(next.id);
        }
        return;
      }
      if ((e.ctrlKey || e.metaKey) && e.key === "w") {
        const editorPane = document.querySelector(".editor-pane");
        if (editorPane && editorPane.contains(document.activeElement)) {
          e.preventDefault();
          if (activeId)
            onClose?.(activeId);
        }
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [tabs, activeId, onActivate, onClose]);
  const handleTabMouseDown = Y_((e, id) => {
    if (e.button === 1) {
      e.preventDefault();
      onClose?.(id);
    }
  }, [onClose]);
  const handleTabClick = Y_((e, id) => {
    if (e.defaultPrevented)
      return;
    if (e.button === 0) {
      onActivate?.(id);
    }
  }, [onActivate]);
  const handleContextMenu = Y_((e, id) => {
    e.preventDefault();
    setContextMenu({ id, x: e.clientX, y: e.clientY });
  }, []);
  const handleClosePointerDown = Y_((e) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);
  const handleCloseClick = Y_((e, id) => {
    e.preventDefault();
    e.stopPropagation();
    onClose?.(id);
  }, [onClose]);
  K_(() => {
    if (!activeId || !stripRef.current)
      return;
    const activeEl = stripRef.current.querySelector(".tab-item.active");
    if (activeEl) {
      activeEl.scrollIntoView({ block: "nearest", inline: "nearest", behavior: "smooth" });
    }
  }, [activeId]);
  const getPaneOverride = Y_((id) => {
    if (!(paneOverrides instanceof Map))
      return null;
    return paneOverrides.get(id) || null;
  }, [paneOverrides]);
  const contextMenuTab = u_(() => tabs.find((tab) => tab.id === contextMenu?.id) || null, [contextMenu?.id, tabs]);
  const contextMenuCanEditSource = u_(() => {
    const tabId = contextMenu?.id;
    if (!tabId)
      return false;
    return canTabEditSource(tabId, getPaneOverride(tabId), (context) => paneRegistry.resolve(context));
  }, [contextMenu?.id, getPaneOverride]);
  const contextMenuEditSourceLabel = u_(() => {
    const tabId = contextMenu?.id;
    if (!tabId)
      return "Edit Source";
    return getTabEditSourceLabel(tabId, getPaneOverride(tabId), (context) => paneRegistry.resolve(context));
  }, [contextMenu?.id, getPaneOverride]);
  const isContextMenuTabDetached = u_(() => {
    const tabId = contextMenu?.id;
    if (!tabId || !(detachedTabs instanceof Map))
      return false;
    return detachedTabs.has(tabId);
  }, [contextMenu?.id, detachedTabs]);
  const contextMenuDiffOpen = u_(() => {
    const tabId = contextMenu?.id;
    if (!tabId || !(diffTabs instanceof Set))
      return false;
    return diffTabs.has(tabId);
  }, [contextMenu?.id, diffTabs]);
  const contextMenuCanCompareToSaved = u_(() => {
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
  return fe`
        <div class="tab-strip" ref=${stripRef} role="tablist">
            ${tabs.map((tab) => fe`
                <div
                    key=${tab.id}
                    class=${`tab-item${tab.id === activeId ? " active" : ""}${tab.dirty ? " dirty" : ""}${tab.pinned ? " pinned" : ""}`}
                    role="tab"
                    aria-selected=${tab.id === activeId}
                    title=${tab.path}
                    onMouseDown=${(e) => handleTabMouseDown(e, tab.id)}
                    onClick=${(e) => handleTabClick(e, tab.id)}
                    onContextMenu=${(e) => handleContextMenu(e, tab.id)}
                >
                    ${tab.pinned && fe`
                        <span class="tab-pin-icon" aria-label="Pinned">
                            <svg viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
                                <path d="M4.456.734a1.75 1.75 0 0 1 2.826.504l.613 1.327a3.1 3.1 0 0 0 2.084 1.707l2.454.584c1.332.317 1.8 1.972.832 2.94L11.06 10l3.72 3.72a.75.75 0 1 1-1.06 1.06L10 11.06l-2.204 2.205c-.968.968-2.623.5-2.94-.832l-.584-2.454a3.1 3.1 0 0 0-1.707-2.084l-1.327-.613a1.75 1.75 0 0 1-.504-2.826z"/>
                            </svg>
                        </span>
                    `}
                    <span class="tab-label">${tab.label}</span>
                    ${detachedTabs instanceof Map && detachedTabs.has(tab.id) && fe`
                        <span class="tab-detached-badge" aria-label="Detached" title="Open in separate window">↗</span>
                    `}
                    <button
                        type="button"
                        class="tab-close"
                        onPointerDown=${handleClosePointerDown}
                        onMouseDown=${handleClosePointerDown}
                        onClick=${(e) => handleCloseClick(e, tab.id)}
                        title=${tab.dirty ? "Unsaved changes" : "Close"}
                        aria-label=${tab.dirty ? "Unsaved changes" : `Close ${tab.label}`}
                    >
                        ${tab.dirty ? fe`<span class="tab-dirty-dot" aria-hidden="true"></span>` : fe`<svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" aria-hidden="true" focusable="false" style=${{ pointerEvents: "none" }}>
                                <line x1="4" y1="4" x2="12" y2="12" style=${{ pointerEvents: "none" }}/>
                                <line x1="12" y1="4" x2="4" y2="12" style=${{ pointerEvents: "none" }}/>
                            </svg>`}
                    </button>
                </div>
            `)}
            ${onToggleDock && fe`
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
            ${onToggleZen && fe`
                <button
                    class=${`tab-strip-zen-toggle${zenMode ? " active" : ""}`}
                    onClick=${onToggleZen}
                    title=${`${zenMode ? "Exit" : "Enter"} zen mode (Ctrl+Shift+Z)`}
                    aria-label=${`${zenMode ? "Exit" : "Enter"} zen mode`}
                    aria-pressed=${zenMode ? "true" : "false"}
                >
                    <svg viewBox="0 0 16 16" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        ${zenMode ? fe`<polyline points="4 8 1.5 8 1.5 1.5 14.5 1.5 14.5 8 12 8"/><polyline points="4 8 1.5 8 1.5 14.5 14.5 14.5 14.5 8 12 8"/>` : fe`<polyline points="5.5 1.5 1.5 1.5 1.5 5.5"/><polyline points="10.5 1.5 14.5 1.5 14.5 5.5"/><polyline points="5.5 14.5 1.5 14.5 1.5 10.5"/><polyline points="10.5 14.5 14.5 14.5 14.5 10.5"/>`}
                    </svg>
                </button>
            `}
        </div>
        ${contextMenu && fe`
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
                ${contextMenuCanEditSource && onEditSource && fe`
                    <button onClick=${() => {
    onEditSource(contextMenu.id);
    setContextMenu(null);
  }}>${contextMenuEditSourceLabel}</button>
                `}
                ${isContextMenuTabDetached && onReattachTab && fe`
                    <button onClick=${() => {
    onReattachTab(contextMenu.id);
    setContextMenu(null);
  }}>Reattach</button>
                `}
                ${onPopOutTab && !isContextMenuTabDetached && fe`
                    <button onClick=${() => {
    const tab = tabs.find((t) => t.id === contextMenu.id);
    onPopOutTab(contextMenu.id, tab?.label);
    setContextMenu(null);
  }}>Open in Window</button>
                `}
                ${contextMenuCanCompareToSaved && onToggleDiff && fe`
                    <hr />
                    <button onClick=${() => {
    onActivate?.(contextMenu.id);
    onToggleDiff(contextMenu.id);
    setContextMenu(null);
  }}>${contextMenuDiffOpen ? "Hide Diff" : "Compare to Saved"}</button>
                `}
                ${onTogglePreview && /\.(md|mdx|markdown)$/i.test(contextMenu.id) && fe`
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
    return fe`
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

// web/src/gi-workspace-tab-lifecycle.ts
function observeWorkspaceTab(container, resized) {
  const view = container.ownerDocument?.defaultView;
  if (!view)
    return () => {};
  if (view.ResizeObserver) {
    const observer = new view.ResizeObserver(resized);
    observer.observe(container);
    return () => observer.disconnect();
  }
  view.addEventListener("resize", resized);
  return () => view.removeEventListener("resize", resized);
}
function mountWorkspaceTab(container, path, changed, read, registry, options = {}) {
  let live = true, instance = null, unobserve = null;
  const stop = () => {
    if (!live)
      return;
    live = false;
    const mounted = instance, disconnect = unobserve;
    instance = null;
    unobserve = null;
    try {
      disconnect?.();
    } catch {}
    try {
      mounted?.dispose();
    } catch {}
    container.replaceChildren();
  };
  const fail = (error) => {
    if (!live)
      return;
    stop();
    changed({ loading: false, error: error instanceof Error ? error.message : "Preview failed." });
  };
  changed({ loading: true, error: "" });
  (async () => {
    try {
      const preview = await read(path, 20000);
      if (!live)
        return;
      const context = { path, mode: "view", preview };
      if (preview?.kind === "text" && typeof preview.text === "string")
        context.content = preview.text;
      if (typeof preview?.mtime === "string")
        context.mtime = preview.mtime;
      if (Number.isFinite(preview?.size) && preview.size >= 0)
        context.size = preview.size;
      const extension = registry.resolve(context);
      if (!extension)
        throw new Error("No read-only preview available.");
      if (extension.placement !== "tabs" || !extension.capabilities?.length || extension.capabilities.some((capability) => capability !== "readonly" && capability !== "preview")) {
        throw new Error("This pane requires capabilities outside Gi’s read-only preview host.");
      }
      instance = extension.mount(container, context);
      instance.onClose?.(() => {
        if (!live)
          return;
        stop();
        options.close?.();
      });
      if (!live)
        return;
      const resized = () => {
        if (!live)
          return;
        try {
          instance?.resize?.();
        } catch (error) {
          fail(error);
        }
      };
      const disconnect = (options.observeResize ?? observeWorkspaceTab)(container, resized);
      if (!live) {
        disconnect();
        return;
      }
      unobserve = disconnect;
      resized();
      if (live)
        changed({ loading: false, error: "" });
    } catch (error) {
      fail(error);
    }
  })();
  return stop;
}

// web/src/gi-workspace-tab.ts
function WorkspaceTab({ path, onClose }) {
  const host = Q_(null);
  const [state, setState] = F_({ loading: true, error: "" });
  const [attempt, setAttempt] = F_(0);
  const close = Q_(onClose);
  W_(() => {
    close.current = onClose;
  });
  W_(() => mountWorkspaceTab(host.current, path, setState, getWorkspaceFile, paneRegistry, { close: () => close.current?.() }), [path, attempt]);
  return fe`<section class="gi-workspace-tab editor-pane" role="region" aria-label=${`Read-only preview: ${path}`}>
        <div class="gi-workspace-tab-toolbar"><span>Read-only preview</span>
            <button onClick=${() => setAttempt((value) => value + 1)} disabled=${state.loading}>Refresh preview</button>
            <button onClick=${onClose}>Close preview</button></div>
        ${state.loading && fe`<p role="status">Loading preview…</p>`}
        ${state.error && fe`<p role="alert">${state.error} <button onClick=${() => setAttempt((value) => value + 1)}>Retry preview</button></p>`}
        <div class="workspace-preview-body gi-workspace-tab-body" ref=${host}></div>
    </section>`;
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
  const filteredRoots = roots.filter((r) => !folded.has(r.id));
  const stack = [];
  for (let i = filteredRoots.length - 1;i >= 0; i--) {
    filteredRoots[i].depth = 0;
    stack.push(filteredRoots[i]);
  }
  while (stack.length > 0) {
    const node = stack.pop();
    const isBranch = node.children.length > 1;
    for (let i = node.children.length - 1;i >= 0; i--) {
      node.children[i].depth = isBranch ? node.depth + 1 : node.depth;
      stack.push(node.children[i]);
    }
  }
  return filteredRoots;
}
function flattenTree2(roots) {
  const result = [];
  const stack = [];
  for (let i = roots.length - 1;i >= 0; i--)
    stack.push(roots[i]);
  while (stack.length > 0) {
    const node = stack.pop();
    result.push(node);
    for (let i = node.children.length - 1;i >= 0; i--)
      stack.push(node.children[i]);
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
  for (let i = 1;i < parts.length; i++) {
    const part = parts[i];
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
  const [state, setState] = F_(() => ({ loading: !initialTree, error: null, data: initialTree }));
  const [selectedId, setSelectedId] = F_(() => initialTree?.leafId || null);
  const [searchFilter, setSearchFilter] = F_("");
  const searchInputRef = Q_(null);
  const activeRowRef = Q_(null);
  const preferredSelectionRef = Q_(initialTree?.leafId || null);
  const loadTreeRef = Q_(null);
  const scheduledRefreshKeyRef = Q_("");
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
  K_(() => {
    loadTree();
  }, [chatJid]);
  const flatRows = u_(() => {
    const data = state.data;
    if (!data || !Array.isArray(data.nodes) || data.nodes.length === 0)
      return [];
    return flattenTree2(data.flat ? buildTreeFromFlat(data.nodes) : data.nodes);
  }, [state.data]);
  K_(() => {
    const nextSelectedId = resolveTreeSelectionId(flatRows, selectedId, preferredSelectionRef.current, state.data?.leafId || null);
    if (nextSelectedId !== selectedId) {
      setSelectedId(nextSelectedId);
    }
    if (preferredSelectionRef.current && state.data?.leafId === preferredSelectionRef.current) {
      preferredSelectionRef.current = null;
    }
  }, [flatRows, selectedId, state.data?.leafId]);
  const filteredRows = u_(() => {
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
      return fields.some((f) => typeof f === "string" && f.toLowerCase().includes(q));
    });
  }, [flatRows, searchFilter]);
  const selectedNode = u_(() => filteredRows.find((n) => n.id === selectedId) || null, [filteredRows, selectedId]);
  const hostUpdateSummary = u_(() => describeSessionTreeHostUpdate(hostUpdate), [
    hostUpdate?.type,
    hostUpdate?.preview,
    hostUpdate?.error,
    hostUpdate?.submittedAt
  ]);
  K_(() => {
    if (activeRowRef.current)
      activeRowRef.current.scrollIntoView({ block: "center", behavior: "auto" });
  }, [selectedId, state.data?.leafId, filteredRows.length]);
  K_(() => {
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
  return fe`
        <div class="session-tree-widget">
            <div class="session-tree-toolbar">
                <div class="session-tree-toolbar-left">
                    <button class="session-tree-btn" type="button" onClick=${() => loadTree()} disabled=${state.loading}>${state.loading ? "Loading…" : "Refresh"}</button>
                    <input ref=${searchInputRef}
                        class="st-search-input" type="text" placeholder="Filter\u2026"
                        value=${searchFilter}
                        onInput=${(e) => setSearchFilter(e.currentTarget.value)}
                        onKeyDown=${(e) => {
    if (e.key === "Escape") {
      setSearchFilter("");
      e.currentTarget.blur();
    }
  }}
                    />
                    ${searchFilter && fe`<span class="session-tree-meta">${filteredRows.length} match${filteredRows.length !== 1 ? "es" : ""}</span>`}
                    ${state.error && fe`<span class="session-tree-error-inline">${state.error}</span>`}
                </div>
                <div class="session-tree-toolbar-right">
                    ${hostUpdateSummary?.text && fe`<span class=${`session-tree-host-update ${hostUpdateTone}`}>${hostUpdateSummary.text}</span>`}
                    ${state.data?.capped && fe`<span class="session-tree-meta">Showing ${state.data?.nodes?.length || 0} of ${state.data?.total || 0}</span>`}
                    ${chatJid && fe`<span class="session-tree-meta">${chatJid}</span>`}
                </div>
            </div>

            <div class="session-tree-content">
                <div class="session-tree-list" role="tree" aria-label="Session tree">
                    ${state.loading && filteredRows.length === 0 && !searchFilter && fe`<div class="session-tree-empty">Loading session tree\u2026</div>`}
                    ${!state.loading && filteredRows.length === 0 && !searchFilter && fe`<div class="session-tree-empty">Session tree is empty.</div>`}
                    ${!state.loading && filteredRows.length === 0 && searchFilter && fe`<div class="session-tree-empty">No entries match \u201c${searchFilter}\u201d</div>`}
                    ${filteredRows.map((node) => {
    const sel = selectedId === node.id;
    const rowClass = `st-row${node.active ? " active" : ""}${sel ? " selected" : ""}`;
    const hasBranch = (node.children || []).length > 1;
    const d = getRowDisplay(node);
    return fe`
                            <button key=${node.id} ref=${node.active || sel ? activeRowRef : null}
                                class=${rowClass} type="button" role="treeitem" aria-selected=${sel}
                                onClick=${() => setSelectedId(node.id)}>
                                <span class="st-indent" style=${`width:${(node.depth || 0) * 16 + 6}px`}></span>
                                <span class=${`st-dot${node.active ? " active" : hasBranch ? " branch" : ""}`}></span>
                                <span class=${`st-tag ${d.tagClass}`}>${d.tag}</span>
                                <span class="st-text">${d.text}</span>
                                ${node.merged && node.resultLength > 0 && fe`<span class="st-size">${formatSize(node.resultLength)}</span>`}
                                ${!node.merged && node.contentLength > 3000 && fe`<span class="st-size">${formatSize(node.contentLength)}</span>`}
                                ${node.hasThinking && fe`<span class="st-badge thinking">\u{1F4AD}</span>`}
                                ${node.label && fe`<span class="st-label">${node.label}</span>`}
                                ${node.active && fe`<span class="st-active">\u25C0</span>`}
                            </button>
                        `;
  })}
                </div>

                <aside class="session-tree-sidebar">
                    ${selectedNode ? fe`
                        <div class="st-side-section">
                            <div class="st-side-label">Entry</div>
                            <div class="st-side-mono">${selectedNode.id}${selectedNode.resultId ? ` → ${selectedNode.resultId}` : ""}</div>
                        </div>
                        <div class="st-side-section">
                            <div class="st-side-label">Type</div>
                            <div class="st-side-value">${selectedNode.role || selectedNode.type || "entry"}${selectedNode.toolName ? ` → ${selectedNode.toolName}` : ""}${selectedNode.merged ? " (merged)" : ""}</div>
                        </div>
                        ${selectedNode.toolInputFull && fe`
                            <div class="st-side-section">
                                <div class="st-side-label">${selectedNode.toolName === "bash" ? "Command" : "Input"}</div>
                                <pre class="st-side-code">${selectedNode.toolInputFull}</pre>
                            </div>
                        `}
                        ${selectedNode.resultDetail && fe`
                            <div class="st-side-section">
                                <div class="st-side-label">Result${selectedNode.resultLength ? ` (${formatSizeLong(selectedNode.resultLength)})` : ""}</div>
                                <pre class="st-side-code">${selectedNode.resultDetail}</pre>
                            </div>
                        `}
                        ${selectedNode.detail && !selectedNode.toolInput && fe`
                            <div class="st-side-section">
                                <div class="st-side-label">${selectedNode.role === "toolResult" ? "Output" : "Content"}${selectedNode.contentLength ? ` (${formatSizeLong(selectedNode.contentLength)})` : ""}</div>
                                <pre class="st-side-code">${selectedNode.detail}</pre>
                            </div>
                        `}
                        ${selectedNode.rawDetail && fe`
                            <div class="st-side-section">
                                <div class="st-side-label">Raw prompt${selectedNode.rawContentLength ? ` (${formatSizeLong(selectedNode.rawContentLength)})` : ""}</div>
                                <pre class="st-side-code">${selectedNode.rawDetail}</pre>
                            </div>
                        `}
                        ${selectedNode.timestamp && fe`
                            <div class="st-side-section">
                                <div class="st-side-label">Time</div>
                                <div class="st-side-value">${new Date(selectedNode.timestamp).toLocaleString()}</div>
                            </div>
                        `}
                        ${(selectedNode.contentLength > 0 || selectedNode.hasThinking) && fe`
                            <div class="st-side-section">
                                <div class="st-side-label">Size</div>
                                <div class="st-side-badges">
                                    ${selectedNode.contentLength > 0 && fe`<span class="st-pill">${formatSizeLong(selectedNode.contentLength)} content</span>`}
                                    ${selectedNode.hasThinking && fe`<span class="st-pill thinking">${formatSizeLong(selectedNode.thinkingLength)} thinking</span>`}
                                    ${selectedNode.merged && selectedNode.resultLength > 0 && fe`<span class="st-pill">${formatSizeLong(selectedNode.resultLength)} result</span>`}
                                </div>
                            </div>
                        `}
                        <div class="st-side-actions">
                            <button class="session-tree-btn primary" type="button" onClick=${() => submitNavigation(false)}>Navigate here</button>
                            <button class="session-tree-btn" type="button" onClick=${() => submitNavigation(true)}>Navigate + summarize</button>
                        </div>
                    ` : fe`<div class="session-tree-empty side">Select an entry to inspect.</div>`}
                </aside>
            </div>
        </div>
    `;
}

// web/src/components/floating-widget-pane.ts
function FloatingWidgetPane({ widget, onClose, onWidgetEvent }) {
  const frameRef = Q_(null);
  const frameLoadedRef = Q_(false);
  const srcDoc = u_(() => buildWidgetSrcDoc(widget), [
    widget?.artifact?.kind,
    widget?.artifact?.html,
    widget?.artifact?.svg,
    widget?.widgetId,
    widget?.toolCallId,
    widget?.turnId,
    widget?.title
  ]);
  K_(() => {
    if (!widget)
      return;
    const handleEsc = (e) => {
      if (e.key === "Escape")
        onClose?.();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [widget, onClose]);
  K_(() => {
    frameLoadedRef.current = false;
  }, [srcDoc]);
  K_(() => {
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
  K_(() => {
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
  K_(() => {
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
  return fe`
        <div class="floating-widget-backdrop" onClick=${() => onClose?.()}>
            <section
                class="floating-widget-pane"
                aria-label=${title}
                onClick=${(e) => e.stopPropagation()}
            >
                <div class="floating-widget-header">
                    <div class="floating-widget-heading">
                        <div class="floating-widget-eyebrow">${originLabel} • ${kind.toUpperCase()}</div>
                        <div class="floating-widget-title">${title}</div>
                        ${(subtitle || description) && fe`
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
                    ${kind === "session_tree" ? fe`<${SessionTreeWidget} widget=${widget} onWidgetEvent=${onWidgetEvent} />` : emptyState ? fe`<div class="floating-widget-empty">${emptyMessage}</div>` : fe`
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
  const previewKind = u_(() => getAttachmentPreviewKind(info?.content_type, filename), [info?.content_type, filename]);
  const previewLabel = getAttachmentPreviewLabel(previewKind);
  const isMarkdown = u_(() => isMarkdownAttachmentPreview(info?.content_type), [info?.content_type]);
  const [loading, setLoading] = F_(previewKind === "text" || previewKind === "html" || previewKind === "archive");
  const [textContent, setTextContent] = F_("");
  const [archivePreview, setArchivePreview] = F_(null);
  const [error, setError] = F_(null);
  const markdownContainerRef = Q_(null);
  const previewLanguage = u_(() => previewLanguageFromAttachment(info, filename), [info, filename]);
  const previewLanguageLabel = u_(() => previewLanguage ? normalizeCodeLanguageLabel(previewLanguage) : null, [previewLanguage]);
  const metadata = u_(() => buildMetadata(info, !isMarkdown ? previewLanguageLabel : null, archivePreview), [info, isMarkdown, previewLanguageLabel, archivePreview]);
  const frameUrl = u_(() => buildFrameUrl(mediaId, filename, previewKind), [mediaId, filename, previewKind]);
  const renderedMarkdown = u_(() => {
    if (!isMarkdown || !textContent)
      return "";
    return renderMarkdown(textContent);
  }, [isMarkdown, textContent]);
  const highlightedText = u_(() => {
    if (isMarkdown || !textContent || !previewLanguage)
      return "";
    return highlightCodeToHtml(textContent, previewLanguage);
  }, [isMarkdown, textContent, previewLanguage]);
  K_(() => {
    const handleEsc = (e) => {
      if (e.key === "Escape")
        onClose();
    };
    document.addEventListener("keydown", handleEsc);
    return () => document.removeEventListener("keydown", handleEsc);
  }, [onClose]);
  K_(() => {
    if (!markdownContainerRef.current || !renderedMarkdown)
      return;
    renderMermaidDiagrams(markdownContainerRef.current);
    return;
  }, [renderedMarkdown]);
  K_(() => {
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
  return fe`
        <${BodyPortal} className="attachment-preview-portal-root">
            <div class="image-modal attachment-preview-modal" onClick=${onClose}>
                <div class="attachment-preview-shell" onClick=${(e) => {
    e.stopPropagation();
  }}>
                    <div class="attachment-preview-header">
                        <div class="attachment-preview-heading">
                            <div class="attachment-preview-title">${filename}</div>
                            <div class="attachment-preview-subtitle">${previewLabel}</div>
                        </div>
                        <div class="attachment-preview-header-actions">
                            ${frameUrl && fe`
                                <a
                                    href=${frameUrl}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="attachment-preview-download"
                                    onClick=${(e) => e.stopPropagation()}
                                >
                                    Open in Tab
                                </a>
                            `}
                            <a
                                href=${getMediaUrl(mediaId)}
                                download=${filename}
                                class="attachment-preview-download"
                                onClick=${(e) => e.stopPropagation()}
                            >
                                Download
                            </a>
                            <button class="attachment-preview-close" type="button" onClick=${onClose}>Close</button>
                        </div>
                    </div>
                    <div class="attachment-preview-body">
                        ${loading && fe`<div class="attachment-preview-state">Loading preview…</div>`}
                        ${!loading && error && fe`<div class="attachment-preview-state">${error}</div>`}
                        ${!loading && !error && previewKind === "image" && fe`
                            <img class="attachment-preview-image" src=${getMediaUrl(mediaId)} alt=${filename} />
                        `}
                        ${!loading && !error && previewKind === "video" && fe`
                            <video class="attachment-preview-video" src=${getMediaUrl(mediaId)} controls autoplay style="max-width:100%;max-height:100%;" />
                        `}
                        ${!loading && !error && previewKind === "html" && fe`
                            <iframe class="attachment-preview-frame" srcdoc=${textContent || ""} sandbox=${HTML_ATTACHMENT_PREVIEW_SANDBOX} title=${filename}></iframe>
                        `}
                        ${!loading && !error && (previewKind === "pdf" || previewKind === "office" || previewKind === "drawio") && frameUrl && fe`
                            <iframe class="attachment-preview-frame" src=${frameUrl} title=${filename}></iframe>
                        `}
                        ${!loading && !error && previewKind === "drawio" && fe`
                            <div class="attachment-preview-readonly-note">Draw.io preview is read-only. Editing tools are disabled in this preview.</div>
                        `}
                        ${!loading && !error && previewKind === "archive" && archivePreview && fe`
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
                                            ${archivePreview.entries.map((entry) => fe`
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
                        ${!loading && !error && previewKind === "text" && isMarkdown && fe`
                            <div
                                ref=${markdownContainerRef}
                                class="attachment-preview-markdown post-content"
                                dangerouslySetInnerHTML=${{ __html: renderedMarkdown }}
                            />
                        `}
                        ${!loading && !error && previewKind === "text" && !isMarkdown && highlightedText && fe`
                            <pre class="attachment-preview-text attachment-preview-code"><code dangerouslySetInnerHTML=${{ __html: highlightedText }} /></pre>
                        `}
                        ${!loading && !error && previewKind === "text" && !isMarkdown && !highlightedText && fe`
                            <pre class="attachment-preview-text">${textContent}</pre>
                        `}
                        ${!loading && !error && previewKind === "unsupported" && fe`
                            <div class="attachment-preview-state">
                                Preview is not available for this file type yet. You can still download it directly.
                            </div>
                        `}
                    </div>
                    <div class="attachment-preview-meta">
                        ${metadata.map((entry) => fe`
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
    const y = (height / 2).toFixed(2);
    return `M 0 ${y} L ${width} ${y}`;
  }
  if (points.length === 1) {
    const normalized = (points[0] - minValue) / (maxValue - minValue);
    const y = (height - normalized * height).toFixed(2);
    return `M 0 ${y} L ${width} ${y}`;
  }
  return points.map((value, index) => {
    const x = index / (points.length - 1 || 1) * width;
    const normalized = (value - minValue) / (maxValue - minValue);
    const y = height - normalized * height;
    return `${index === 0 ? "M" : "L"} ${x.toFixed(2)} ${y.toFixed(2)}`;
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
  const [enabled, setEnabled] = F_(() => readStoredMetersEnabled(false));
  const [collapsed, setCollapsed] = F_(() => readStoredMetersCollapsed(false));
  const [isNarrowLayout, setIsNarrowLayout] = F_(() => readIsNarrowLayout());
  const [metrics, setMetrics] = F_({
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
  const [loading, setLoading] = F_(false);
  K_(() => {
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
  K_(() => {
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
  K_(() => {
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
  const cpuPath = u_(() => buildSparklinePath(metrics.cpu_series, 56, 16, { min: 0, max: 100 }), [metrics.cpu_series]);
  const ramPath = u_(() => buildSparklinePath(metrics.ram_series, 56, 16, { min: 0, max: 100 }), [metrics.ram_series]);
  const swapPath = u_(() => buildSparklinePath(metrics.swap_series, 56, 16, { min: 0, max: 100 }), [metrics.swap_series]);
  const rssPath = u_(() => buildSparklinePath(metrics.process_rss_series_bytes), [metrics.process_rss_series_bytes]);
  const showSwap = Number.isFinite(Number(metrics.swap_percent)) && metrics.swap_total_bytes > 0;
  const currentRssBytes = resolveCurrentRssBytes(metrics);
  const showRss = shouldShowRss(metrics);
  const compactSummary = u_(() => buildCompactMetersSummary(metrics), [metrics]);
  if (!enabled || !isActiveInstance)
    return null;
  const title = collapsed ? "Show system meters" : loading ? "Updating system meters… Click to collapse." : "System meters — click to collapse.";
  const handleToggleCollapsed = (event) => {
    event?.stopPropagation?.();
    toggleMetersCollapsed();
  };
  return fe`
        <div class=${`system-meters-hud system-meters-hud-${mode}${collapsed ? " is-collapsed" : ""}`} aria-live="polite">
            <button
                class="system-meters-card"
                type="button"
                title=${title}
                aria-label=${title}
                aria-expanded=${collapsed ? "false" : "true"}
                onClick=${handleToggleCollapsed}
            >
                ${collapsed ? fe`<span class="system-meters-collapse-tab" aria-hidden="true">◂</span>` : isNarrowLayout ? fe`<span class="system-meters-compact-summary">${compactSummary}</span>` : fe`
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
                            ${showRss && fe`
                                <div class="system-meters-row rss">
                                    <span class="system-meters-label">RSS</span>
                                    <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                        <path d=${rssPath}></path>
                                    </svg>
                                    <span class="system-meters-value">${formatBytesCompact(currentRssBytes)}</span>
                                </div>
                            `}
                            ${showSwap && fe`
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

// web/src/gi-menu-dismissal.ts
function bindMenuDismissal(menu, trigger, close) {
  const doc = menu.ownerDocument;
  const inside = (event) => event.composedPath().some((node) => node === menu || node === trigger);
  const finish = () => {
    close();
    if (trigger?.isConnected && !trigger.hasAttribute("disabled"))
      trigger.focus({ preventScroll: true });
  };
  const outsideStart = (event) => {
    if (settingsOwnsKeyboard(doc) || inside(event))
      return;
    if (event.type === "mousedown")
      event.preventDefault();
    event.stopImmediatePropagation();
  };
  const click = (event) => {
    if (!event.isTrusted || settingsOwnsKeyboard(doc) || inside(event))
      return;
    event.preventDefault();
    event.stopImmediatePropagation();
    finish();
  };
  const key = (event) => {
    if (settingsOwnsKeyboard(doc) || event.key !== "Escape" || event.isComposing)
      return;
    event.preventDefault();
    event.stopImmediatePropagation();
    finish();
  };
  const events = ["pointerdown", "pointerup", "mousedown", "mouseup", "touchstart", "touchend"];
  for (const name of events)
    doc.addEventListener(name, outsideStart, true);
  doc.addEventListener("click", click, true);
  doc.addEventListener("keydown", key, true);
  return () => {
    for (const name of events)
      doc.removeEventListener(name, outsideStart, true);
    doc.removeEventListener("click", click, true);
    doc.removeEventListener("keydown", key, true);
  };
}

// web/src/ui/recent-files.ts
var RECENT_FILES_KEY = "piclaw_recent_files";
var MAX_RECENT_FILES = 5;
function isIgnoredPath(path) {
  const normalizedPath = path.trim();
  const legacyPath = normalizedPath.replace(/^\/+/, "");
  return legacyPath.startsWith("__terminal") || legacyPath.startsWith("__vnc") || normalizedPath.startsWith("piclaw://terminal") || normalizedPath.startsWith("piclaw://vnc");
}
function normalizeRecentFiles(value) {
  if (!Array.isArray(value))
    return [];
  const files = [];
  const seen = new Set;
  for (const entry of value) {
    if (typeof entry !== "string")
      continue;
    const path = entry.trim();
    if (!path || isIgnoredPath(path) || seen.has(path))
      continue;
    seen.add(path);
    files.push(path);
    if (files.length >= MAX_RECENT_FILES)
      break;
  }
  return files;
}
function getRecentFiles() {
  try {
    if (typeof localStorage === "undefined")
      return [];
    const raw = localStorage.getItem(RECENT_FILES_KEY);
    if (!raw)
      return [];
    return normalizeRecentFiles(JSON.parse(raw));
  } catch (error) {
    console.warn("[recent-files] Failed to read recent files from localStorage.", error);
    return [];
  }
}

// web/src/utils/i18n.ts
var DEFAULT_LOCALE = "en";
var SUPPORTED_LOCALES = ["en", "zh-CN", "ja"];
var LOCALE_LABELS = {
  en: "English",
  "zh-CN": "简体中文",
  ja: "日本語"
};
var LOCALE_STORAGE_KEY = "piclaw_locale";
var LOCALE_CHANGE_EVENT = "piclaw-locale-change";
var EN = {
  "compose.placeholder": "Message (Enter to send, Shift+Enter for newline)...",
  "compose.send": "Send",
  "compose.stop": "Stop",
  "compose.searchPlaceholder": "Search (Enter to run)...",
  "compose.clearAll": "Clear all",
  "compose.clearAllTitle": "Clear all attachments and references",
  "compose.scope": "Scope",
  "compose.searchScope": "Search scope",
  "compose.scopeCurrent": "Current",
  "compose.scopeBranchFamily": "Branch family",
  "compose.scopeAll": "All chats",
  "compose.filterImages": "Images",
  "compose.filterAttachments": "Attachments",
  "compose.search": "Search",
  "compose.closeSearch": "Close search",
  "compose.shareLocation": "Share location",
  "compose.attachFile": "Attach file",
  "compose.queueControls": "Queued follow-up controls",
  "compose.moveUp": "Move up",
  "compose.moveUpQueue": "Move up in queue",
  "compose.moveDown": "Move down",
  "compose.moveDownQueue": "Move down in queue",
  "compose.editInCompose": "Edit in compose",
  "compose.returnToEditor": "Return queued message to editor",
  "compose.injectSteer": "Inject queued follow-up as steer",
  "compose.steer": "Steer",
  "compose.cancelQueued": "Cancel queued message",
  "compose.resizeInput": "Resize message input",
  "compose.resizeInputHint": "Drag to resize message input",
  "compose.modelPicker": "Model picker",
  "compose.sessionsAndAgents": "Sessions and agents",
  "compose.openModelPicker": "Open model picker",
  "compose.newBranchTitle": "Create a new branch from this chat",
  "compose.newRootTitle": "Create a clean root session such as web:ops",
  "compose.renameSessionTitle": "Rename the current session",
  "compose.pruneSessionTitle": "Delete (prune) current agent/session branch",
  "compose.filterImagesTitle": "Only show messages with images",
  "compose.filterAttachmentsTitle": "Only show messages with attachments",
  "compose.selectModel": "Select model",
  "compose.loadingModels": "Loading models…",
  "compose.noModels": "No models available.",
  "compose.nextModel": "Next model",
  "compose.manageSessions": "Manage sessions & agents",
  "compose.noSessions": "No other sessions yet.",
  "compose.newBranch": "New branch",
  "compose.newRoot": "New root…",
  "compose.mergeCurrent": "Merge current w/ parent",
  "compose.renameCurrent": "Rename current…",
  "compose.deleteCurrent": "Delete current…",
  "compose.mergeInto": "Merge this branch into {target}",
  "compose.mergeBlocked": "This branch cannot be merged while active or while it has children",
  "workspace.title": "Workspace",
  "workspace.moveConfirm": 'Move {entry} "{name}" from {source} to {target}?',
  "workspace.root": "the workspace root",
  "workspace.file": "file",
  "workspace.folder": "folder",
  "workspace.newFile": "New file",
  "workspace.refresh": "Refresh",
  "workspace.actions": "Workspace actions",
  "workspace.uploadFiles": "Upload files",
  "workspace.reindexing": "Reindexing workspace…",
  "workspace.deleteFile": "Delete file",
  "workspace.download": "Download",
  "workspace.uploadToFolder": "Upload files to this folder",
  "workspace.addFolderHint": "Add folder hint to compose",
  "workspace.downloadZip": "Download folder as zip",
  "workspace.openInTab": "Open in tab",
  "workspace.openInEditor": "Open in editor",
  "workspace.renameSelected": "Rename selected",
  "workspace.downloadSelectedFile": "Download selected file",
  "workspace.downloadSelectedFolder": "Download selected folder (zip)",
  "workspace.deleteSelectedFile": "Delete selected file",
  "shell.settings": "Settings",
  "shell.newChat": "New chat",
  "shell.connecting": "Connecting…",
  "shell.connected": "Connected",
  "language.label": "Language",
  "settings.title": "Settings",
  "settings.close": "Close (Esc)",
  "settings.filter": "Filter…",
  "settings.loading": "Loading settings…",
  "settings.section.general": "General",
  "settings.section.sessions": "Sessions",
  "settings.section.recordings": "Recordings",
  "settings.section.compaction": "Compaction",
  "settings.section.budget": "Budget",
  "settings.section.keyboard": "Keyboard",
  "settings.section.workspace": "Workspace",
  "settings.section.environment": "Environment",
  "settings.section.providers": "Providers",
  "settings.section.models": "Models",
  "settings.section.theme": "Appearance",
  "settings.section.scheduled-tasks": "Scheduled Tasks",
  "settings.section.quick-actions": "Quick Actions",
  "settings.section.keychain": "Keychain",
  "settings.section.tools": "Tools",
  "settings.section.addons": "Add-ons",
  "settings.placeholder.recordings": "Filter recordings…",
  "settings.placeholder.keyboard": "Filter shortcuts…",
  "settings.placeholder.environment": "Filter environment…",
  "settings.placeholder.models": "Filter models…",
  "settings.placeholder.scheduled-tasks": "Filter scheduled tasks…",
  "settings.placeholder.quick-actions": "Filter quick actions…",
  "settings.placeholder.keychain": "Filter entries…",
  "settings.placeholder.tools": "Filter tools…",
  "settings.placeholder.addons": "Filter add-ons…",
  "preview.close": "Close",
  "preview.loading": "Loading preview…",
  "preview.files": "Files",
  "preview.folders": "Folders",
  "preview.compressed": "Compressed",
  "preview.uncompressed": "Uncompressed",
  "preview.name": "Name",
  "preview.type": "Type",
  "preview.method": "Method",
  "preview.size": "Size",
  "post.deleteMessage": "Delete message",
  "post.tooLarge": "Message too large to display.",
  "post.previewTruncated": "Preview truncated.",
  "post.submitted": "Submitted",
  "post.discard": "Discard",
  "post.save": "Save",
  "post.cancel": "Cancel",
  "post.addNote": "Add note",
  "post.addNotePlaceholder": "Add a note…",
  "post.restartNotice": "Restarting now — Reason: {reason}",
  "post.restartCompleted": "Restart completed.",
  "post.agentSelfResume": "Agent self-resume",
  "tab.close": "Close",
  "tab.closeOthers": "Close Others",
  "tab.closeAll": "Close All",
  "tab.reattach": "Reattach",
  "tab.openInWindow": "Open in Window",
  "tab.openInNewTab": "Open in New Tab",
  "tab.pinned": "Pinned",
  "tab.detached": "Detached",
  "tab.openSeparateWindow": "Open in separate window",
  "status.trackedVariables": "Tracked variables",
  "status.attachToSession": "Attach to session",
  "status.files": "Files",
  "status.proposedDiff": "Proposed diff",
  "status.copyTmux": "Copy tmux command",
  "status.experimentDuration": "Experiment duration",
  "status.sinceLastActivity": "Since last activity",
  "annotator.title": "Annotate image",
  "annotator.typeLabel": "Type label…",
  "annotator.undo": "Undo",
  "annotator.resetZoom": "Reset zoom",
  "tree.filter": "Filter…",
  "tree.sessionTree": "Session tree",
  "btw.label": "BTW side conversation",
  "btw.close": "Close BTW",
  "btw.thinking": "Thinking",
  "mdpreview.close": "Close preview",
  "mdpreview.unavailable": "Preview unavailable",
  "widget.close": "Close widget",
  "oobe.gettingStarted": "Getting started",
  "oobe.needsSetupTitle": "Instance needs setup",
  "oobe.configuredTitle": "Instance is configured",
  "oobe.needsSetupBody": "This instance is not yet configured. Open Settings and set up AI providers/models to start sending requests.",
  "oobe.configuredBody": "This instance looks configured. Review or update provider and model settings in Settings.",
  "oobe.openSettings": "Open Settings",
  "oobe.dismiss": "Dismiss",
  "oobe.done": "Done",
  "palette.placeholder": "Type to jump to an agent, workspace action, or slash command…",
  "palette.hideWorkspace": "Hide workspace",
  "palette.showWorkspace": "Show workspace",
  "palette.hideWorkspaceDesc": "Hide the workspace sidebar.",
  "palette.showWorkspaceDesc": "Show the workspace sidebar.",
  "palette.exitChatOnly": "Exit chat-only mode",
  "palette.chatOnly": "Chat-only mode",
  "palette.exitChatOnlyDesc": "Return to the split workspace layout.",
  "palette.chatOnlyDesc": "Switch to the chat-only layout.",
  "palette.groupAgents": "Agents",
  "palette.groupWorkspace": "Workspace",
  "palette.groupSlash": "Slash commands",
  "palette.hintMove": "Move",
  "palette.hintSelect": "Select",
  "palette.hintPopOut": "Pop out",
  "palette.hintClose": "Close",
  "settings.appliedNotice": "Settings applied. Changes take effect on the next turn.",
  "settings.sessions.lifecycle": "Session Lifecycle",
  "settings.sessions.autoRotate": "Auto-rotate sessions",
  "settings.sessions.maxSize": "Max session size (MB)",
  "settings.sessions.maxSizeAria": "max session size",
  "settings.sessions.agentBehaviour": "Agent Behaviour",
  "settings.sessions.toolBudget": "Tool use budget",
  "settings.sessions.toolBudgetAria": "tool use budget",
  "settings.sessions.toolBudgetHint": "max completed tool executions per turn",
  "settings.sessions.isolation": "Session isolation",
  "settings.sessions.isolationNone": "None — full cross-session visibility",
  "settings.sessions.isolationSummary": "Summary — tools visible, no arguments",
  "settings.sessions.isolationFull": "Full — sessions cannot see each other",
  "settings.editor.heading": "Editor",
  "settings.editor.vimMode": "Vim mode",
  "settings.editor.showWhitespace": "Show whitespace",
  "settings.editor.livePreview": "Markdown live preview",
  "settings.editor.fontSize": "Font size (px)",
  "settings.editor.fontSizeAria": "editor font size",
  "settings.editor.fontFamily": "Font family",
  "settings.editor.fontFamilyPlaceholder": "monospace (default)",
  "settings.editor.localOnlyHint": "This browser only. Editor changes are stored in local browser storage and take effect when you next open or reload a file tab.",
  "settings.appearance.syncing": "Syncing appearance…",
  "settings.appearance.default": "Default",
  "settings.appearance.autoLightDark": "auto (light/dark)",
  "settings.appearance.tint": "Tint:",
  "settings.appearance.clearTint": "Clear tint",
  "settings.appearance.none": "none",
  "settings.appearance.outputPadding": "Output padding",
  "settings.appearance.outputPaddingHint": "Extra space around messages and thinking panels.",
  "settings.keyboard.heading": "Keyboard",
  "settings.keyboard.hint1": "Customize app-wide shortcuts as comma-separated bindings. Changes apply immediately.",
  "settings.keyboard.hint1b": "is reserved for dismiss/abort and cannot be rebound.",
  "settings.keyboard.hint2mid": "and typing",
  "settings.keyboard.hint2end": "outside the compose box open this pane.",
  "settings.keyboard.resetAll": "Reset all to defaults",
  "settings.keyboard.defaultColon": "Default:",
  "settings.keyboard.save": "Save",
  "settings.keyboard.defaultBtn": "Default",
  "settings.keyboard.noMatch": "No shortcuts match this filter.",
  "settings.keyboard.invalidShortcut": "Invalid shortcut: {token}. Escape is reserved and cannot be rebound.",
  "settings.keyboard.saved": "Keyboard shortcuts saved.",
  "settings.keyboard.resetOne": "Keyboard shortcut reset to default.",
  "settings.keyboard.resetAllDone": "Keyboard shortcuts reset to defaults.",
  "settings.workspace.serverApplied": "Workspace settings applied. Server-side limits affect new workspace requests immediately.",
  "settings.workspace.browserApplied": "Browser workspace settings applied immediately in this tab.",
  "settings.workspace.access": "Access",
  "settings.workspace.enableTerminal": "Enable web terminal",
  "settings.workspace.allowVnc": "Allow direct VNC targets",
  "settings.workspace.accessHint": "Terminal access updates immediately. Direct VNC target policy applies to new VNC requests.",
  "settings.workspace.guardrails": "Server scan guardrails",
  "settings.workspace.maxDepth": "Max tree depth",
  "settings.workspace.maxDepthAria": "workspace tree max depth",
  "settings.workspace.maxDepthHintPre": "caps all",
  "settings.workspace.maxDepthHintPost": "requests",
  "settings.workspace.maxEntries": "Max entries per scan",
  "settings.workspace.maxEntriesAria": "workspace tree max entries",
  "settings.workspace.maxEntriesHint": "truncate oversized tree walks earlier",
  "settings.workspace.thisBrowser": "This browser",
  "settings.workspace.refreshInterval": "Refresh interval (seconds)",
  "settings.workspace.refreshIntervalAria": "workspace refresh interval",
  "settings.workspace.folderDepth": "Folder preview scan depth",
  "settings.workspace.folderDepthAria": "folder preview scan depth",
  "settings.workspace.folderDepthHintPre": "set to",
  "settings.workspace.folderDepthHintPost": "to disable folder size preview scans",
  "settings.workspace.footerHint": "Root and folder-expansion tree loads remain shallow; the folder size preview is the deepest workspace scan in the UI.",
  "settings.models.thinkingLevel": "Thinking level",
  "settings.models.noThinking": "Current model does not support thinking.",
  "settings.models.thinkingLevelLabel": "Thinking level:",
  "settings.models.loading": "Loading models…",
  "settings.models.summary": "Model and provider names may wrap in narrow panes to avoid clipping.",
  "settings.models.scopedOnly": "Scoped models only",
  "settings.models.scopedCheckboxPre": "Use Pi",
  "settings.models.scopedCheckboxPost": "for Piclaw model lists",
  "settings.models.scopedHintPre": "Filters this picker and the",
  "settings.models.scopedHintPost": "tool. TUI model selection remains unchanged.",
  "settings.models.colModel": "Model",
  "settings.models.colProvider": "Provider",
  "settings.models.colContext": "Context",
  "settings.models.colReasoning": "Reasoning",
  "settings.models.noMatch": 'No models match "{filter}"',
  "settings.tools.unavailable": "Tool data not available.",
  "settings.tools.search": "Search",
  "settings.tools.matchMode": "Match mode",
  "settings.tools.orMode": "Any keyword (OR) — results match at least one search term",
  "settings.tools.andMode": "All keywords (AND) — results must match every search term",
  "settings.tools.colEnabled": "Enabled",
  "settings.tools.colTool": "Tool",
  "settings.tools.colCompact": "Result compaction",
  "settings.tools.colKind": "Kind",
  "settings.tools.colSummary": "Summary",
  "settings.tools.colSource": "Source",
  "settings.tools.disableCompaction": "Disable tool-result compaction for this tool",
  "settings.tools.enableCompaction": "Enable tool-result compaction for this tool",
  "settings.tools.noMatch": 'No tools match "{filter}"',
  "settings.tools.footer": "Tool activation is managed by the agent runtime. Group checkboxes collapse/expand; the “Compact” column controls tool-result compaction eligibility.",
  "settings.environment.heading": "Environment",
  "settings.environment.introPre": "Showing non-keychain environment variables only. Overrides are stored in extension KV and applied to",
  "settings.environment.introPost": ", so subsequent tool calls inherit them.",
  "settings.environment.refresh": "Refresh",
  "settings.environment.addOverride": "Add override",
  "settings.environment.valuePlaceholder": "value",
  "settings.environment.save": "Save",
  "settings.environment.countLine": "{count} variables visible • {overrides} overrides active • {keychain} keychain-injected variables hidden",
  "settings.environment.overridden": "Overridden in KV",
  "settings.environment.inherited": "Inherited from process environment",
  "settings.environment.kindOverride": "override",
  "settings.environment.kindProcess": "process",
  "settings.environment.clear": "Clear",
  "settings.environment.noMatch": 'No environment variables match "{filter}".',
  "settings.environment.refreshedToast": "Environment refreshed.",
  "settings.environment.savedToast": "Saved environment override for {name}.",
  "settings.environment.clearedToast": "Cleared environment override for {name}.",
  "settings.quickActions.loading": "Loading…",
  "settings.quickActions.heading": "Timeline Quick Actions",
  "settings.quickActions.intro": "Choose which actions appear in the timeline typeahead. Agents are always pinned first, then workspace commands, then slash commands.",
  "settings.quickActions.enableAll": "Enable all",
  "settings.quickActions.saving": "Saving…",
  "settings.quickActions.saveApply": "Save & apply",
  "settings.quickActions.workspaceCommands": "Workspace commands",
  "settings.quickActions.noWorkspaceMatch": "No workspace commands match this filter.",
  "settings.quickActions.slashCommands": "Slash commands",
  "settings.quickActions.slashFallback": "slash command",
  "settings.quickActions.noSlashMatch": "No slash commands match this filter.",
  "settings.quickActions.savingToast": "Saving quick actions…",
  "settings.quickActions.savedToast": "Quick Actions saved.",
  "settings.providers.authApiKey": "API key",
  "settings.providers.authConfigured": "Configured",
  "settings.providers.heading": "Providers",
  "settings.providers.tagCustom": "Custom",
  "settings.providers.logout": "Logout",
  "settings.providers.reconfigure": "Reconfigure",
  "settings.providers.setUp": "Set up",
  "settings.providers.setupHint": "Sign-in flows open in the browser. In narrow panes the setup form stacks vertically to avoid clipping.",
  "settings.providers.starting": "Starting…",
  "settings.providers.signInOAuth": "Sign in with OAuth",
  "settings.providers.apiKeyLabel": "API Key",
  "settings.providers.apiKeyPlaceholder": "Enter API key",
  "settings.providers.save": "Save",
  "settings.providers.configuring": "Configuring…",
  "settings.providers.saveConfig": "Save configuration",
  "settings.providers.apiKeyEmpty": "API key cannot be empty.",
  "settings.providers.configuringToast": "Configuring {provider}…",
  "settings.providers.configured": "{provider} configured.",
  "settings.providers.startingOAuth": "Starting OAuth for {provider}…",
  "settings.providers.oauthOpened": "OAuth window opened. Complete the sign-in flow, then close this message.",
  "settings.providers.oauthStarted": "OAuth flow started for {provider}. Check the chat.",
  "settings.providers.loggingOut": "Logging out {provider}…",
  "settings.providers.loggedOut": "Logged out {provider}. Restart may be needed.",
  "settings.general.identity": "Identity",
  "settings.general.userLabel": "User",
  "settings.general.yourName": "Your name",
  "settings.general.agentLabel": "Agent",
  "settings.general.agentName": "Agent name",
  "settings.general.notifications": "Notifications",
  "settings.general.browserNotifications": "Browser notifications",
  "settings.general.notifSecureHint": "Use the \uD83D\uDD14 bell button in the compose bar to enable/disable notifications. Web Push requires HTTPS or localhost.",
  "settings.general.notifInsecureHint": "⚠ Not available — requires a secure context (HTTPS or localhost). Access via SSH tunnel or reverse proxy with TLS to enable.",
  "settings.general.display": "Display",
  "settings.general.systemMeters": "System meters",
  "settings.general.systemMetersHint": "CPU/memory/network meters in the status bar. This browser only.",
  "settings.general.instanceConfig": "Instance Configuration",
  "settings.general.composeUpload": "Compose upload (MB)",
  "settings.general.composeUploadAria": "compose upload limit",
  "settings.general.composeUploadHint": "chat/media attachments",
  "settings.general.workspaceUpload": "Workspace upload (MB)",
  "settings.general.workspaceUploadAria": "workspace upload limit",
  "settings.general.workspaceUploadHint": "defaults to 256 MB; chunked uploads allow up to 1 GB",
  "settings.general.agentRecovery": "Advanced · Agent recovery",
  "settings.general.automaticRecovery": "Automatic recovery",
  "settings.general.automaticRecoveryHint": "Retry recoverable failed turns automatically.",
  "settings.general.recoveryMaxAttempts": "Maximum attempts",
  "settings.general.recoveryMaxAttemptsAria": "automatic recovery maximum attempts",
  "settings.general.recoveryMaxAttemptsHint": "0 inherits the normal retry limit.",
  "settings.general.recoveryTotalBudget": "Total budget (ms)",
  "settings.general.recoveryTotalBudgetAria": "automatic recovery total budget in milliseconds",
  "settings.general.recoveryTotalBudgetHint": "0 derives one-third of the turn timeout, bounded to 6–60 minutes. Positive values are explicit caps.",
  "settings.general.recoveryEffectiveBudget": "Effective for the default turn timeout: {budget} ms.",
  "settings.general.authentication": "Authentication",
  "settings.general.widgetToken": "Widget bearer token",
  "settings.general.token": "Token",
  "settings.general.hideToken": "Hide token",
  "settings.general.revealToken": "Reveal token",
  "settings.general.copyToken": "Copy token",
  "settings.general.copied": "Copied",
  "settings.general.regenerating": "Regenerating…",
  "settings.general.regenerate": "Regenerate",
  "settings.general.tokenHintPre": "Read-only token for",
  "settings.general.tokenHintMid": "and",
  "settings.general.tokenHintPost": ". Use as",
  "settings.general.tokenHintEnd": ".",
  "settings.general.copyFailed": "Could not copy widget token. Select the token field and copy manually.",
  "settings.general.regenConfirm": "Regenerate the widget token? Existing macOS widgets using the old token will stop updating.",
  "settings.general.totpTitle": "TOTP setup QR",
  "settings.general.totpConfiguredHint": "Current web-login authenticator secret. Scan this QR to add another authenticator device.",
  "settings.general.totpUnconfiguredHint": "TOTP is not configured for this instance yet, so no setup QR is available.",
  "settings.general.issuer": "Issuer",
  "settings.general.label": "Label",
  "settings.general.secret": "Secret",
  "settings.general.avatarUpload": "Click to upload",
  "settings.developer.heading": "Developer",
  "settings.developer.devMode": "Developer mode",
  "settings.developer.localHint": "This browser only. Developer-mode toggles and add-on catalog overrides are stored in local browser storage.",
  "settings.developer.addonSources": "Add-on Sources",
  "settings.developer.catalogUrl": "Catalog URL",
  "settings.developer.catalogHint": "Primary add-on catalog URL. Leave empty to use the default",
  "settings.developer.additionalCatalogs": "Additional catalog URLs",
  "settings.developer.additionalHint": "Fetched in addition to the primary/default catalog. One URL per line.",
  "settings.developer.repoUrl": "Repo URL",
  "settings.developer.repoHintPre": "Override the git repo used for",
  "settings.developer.repoHintPost": "installs. Leave empty for default.",
  "settings.developer.debug": "Debug",
  "settings.developer.logSse": "Log SSE events",
  "settings.developer.logToolCalls": "Log tool calls",
  "settings.developer.debugHint": "Debug flags take effect on next page reload.",
  "settings.addons.installing": "Installing {slug}…",
  "settings.addons.removing": "Removing {slug}…",
  "settings.addons.installedToast": "Add-on installed.",
  "settings.addons.removedToast": "Add-on removed.",
  "settings.addons.restarting": "Restarting piclaw…",
  "settings.addons.restartComplete": "Restart complete — add-ons refreshed.",
  "settings.addons.restartTimeout": "Backend did not return in time. Reload the page manually.",
  "settings.addons.fetching": "Fetching add-ons…",
  "settings.addons.loadFailed": "Could not load add-ons.",
  "settings.addons.catalogFromPre": "Catalog from",
  "settings.addons.catalogMerged": "{count} catalog sources merged.",
  "settings.addons.installNote": "Package-first install via Bun; restart required after install/uninstall.",
  "settings.addons.failedFetchSingular": "Failed to fetch {count} catalog source:",
  "settings.addons.failedFetchPlural": "Failed to fetch {count} catalog sources:",
  "settings.addons.activeSources": "Active catalog sources ({count})",
  "settings.addons.windowsWarning": "Native Windows add-on installs are higher risk: Bun package installs, symlink cleanup, locked files, and restart timing can all be less predictable than in Linux/WSL. Prefer WSL or a container when possible.",
  "settings.addons.typeExtSkill": "extension + skill",
  "settings.addons.typeSkill": "skill",
  "settings.addons.typeExt": "extension",
  "settings.addons.update": "Update",
  "settings.addons.remove": "Remove",
  "settings.addons.install": "Install",
  "settings.addons.noMatch": 'No add-ons match "{filter}"',
  "settings.addons.restartNotice": "Extension changes are installed but inactive until piclaw restarts.",
  "settings.addons.restartNow": "Restart Now",
  "settings.recordings.modeFull": "full / trusted",
  "settings.recordings.modeMetadata": "metadata only",
  "settings.recordings.modeRedacted": "redacted",
  "settings.recordings.selectPrompt": "Select a recording to inspect, replay, export, or delete it.",
  "settings.recordings.playback": "Playback",
  "settings.recordings.refresh": "Refresh",
  "settings.recordings.delete": "Delete",
  "settings.recordings.status": "Status",
  "settings.recordings.mode": "Mode",
  "settings.recordings.chat": "Chat",
  "settings.recordings.started": "Started",
  "settings.recordings.ended": "Ended",
  "settings.recordings.events": "Events",
  "settings.recordings.redactions": "Redactions",
  "settings.recordings.exportJson": "Export JSON",
  "settings.recordings.exportJsonl": "Export JSONL",
  "settings.recordings.exportHtml": "Export standalone HTML",
  "settings.recordings.eventSummary": "Event summary",
  "settings.recordings.inspectHint": "Open or refresh details to inspect trace events.",
  "settings.recordings.firstEvents": "First events",
  "settings.recordings.heading": "Session Recording",
  "settings.recordings.intro": "Opt-in trace capture for deterministic playback and screen-recording exports. Playback never calls live agent or tool endpoints.",
  "settings.recordings.chatJid": "Chat JID",
  "settings.recordings.title": "Title",
  "settings.recordings.titlePlaceholder": "Demo recording",
  "settings.recordings.modeLabelField": "Mode",
  "settings.recordings.optRedacted": "Redacted",
  "settings.recordings.optMetadata": "Metadata only",
  "settings.recordings.optFull": "Full / trusted local",
  "settings.recordings.includeSnapshot": "Include timeline snapshot",
  "settings.recordings.extraKeys": "Extra redacted keys",
  "settings.recordings.extraPatterns": "Extra regex patterns",
  "settings.recordings.stopCurrent": "Stop current chat recording",
  "settings.recordings.start": "Start recording",
  "settings.recordings.redactionPreview": "Redaction preview",
  "settings.recordings.previewRedaction": "Preview redaction",
  "settings.recordings.loading": "Loading recordings…",
  "settings.recordings.noneYet": "No recordings yet.",
  "settings.recordings.noneYetHint": "Start a recording above, then use playback/export for deterministic screen capture.",
  "settings.recordings.listLabel": "Session recordings",
  "settings.recordings.eventsCount": "{count} events",
  "settings.recordings.noMatch": "No recordings match “{filter}”.",
  "settings.recordings.startedToast": "Recording started for {chat}.",
  "settings.recordings.startFailed": "Failed to start recording.",
  "settings.recordings.stoppedToast": "Recording stopped for {chat}.",
  "settings.recordings.stopFailed": "Failed to stop recording.",
  "settings.recordings.deleteConfirm": "Delete recording {id}?",
  "settings.recordings.deletedToast": "Recording deleted.",
  "settings.recordings.deleteFailed": "Failed to delete recording.",
  "settings.recordings.loadOneFailed": "Failed to load recording.",
  "settings.recordings.loadFailed": "Failed to load recordings.",
  "settings.recordings.previewFailed": "Preview failed.",
  "settings.keychain.loadFailed": "Failed to load keychain.",
  "settings.keychain.addFailed": "Failed to add entry.",
  "settings.keychain.deleteFailed": "Failed to delete entry.",
  "settings.keychain.saveNotesFailed": "Failed to save notes.",
  "settings.keychain.revealFailed": "Failed to reveal.",
  "settings.keychain.loading": "Loading keychain…",
  "settings.keychain.entryCountSingular": "{count} entry",
  "settings.keychain.entryCountPlural": "{count} entries",
  "settings.keychain.matchingFilter": ' matching "{filter}"',
  "settings.keychain.encryptedSuffix": ", encrypted at rest.",
  "settings.keychain.clickPrefix": "Click",
  "settings.keychain.revealSuffix": "to reveal.",
  "settings.keychain.cancel": "Cancel",
  "settings.keychain.addEntry": "+ Add entry",
  "settings.keychain.namePlaceholder": "Entry name (e.g. github/my-token)",
  "settings.keychain.secretPlaceholder": "Secret value",
  "settings.keychain.usernamePlaceholder": "Username (optional)",
  "settings.keychain.saving": "Saving…",
  "settings.keychain.save": "Save",
  "settings.keychain.userNotePlaceholder": "User note (visible in this UI only)",
  "settings.keychain.agentNotePlaceholder": "Agent note (safe to expose to agents)",
  "settings.keychain.noMatchFilter": "No entries match the filter.",
  "settings.keychain.noEntries": "No keychain entries.",
  "settings.keychain.hideSecret": "Hide secret",
  "settings.keychain.revealSecret": "Reveal secret",
  "settings.keychain.deleteQ": "Delete?",
  "settings.keychain.yes": "Yes",
  "settings.keychain.no": "No",
  "settings.keychain.deleteTitle": "Delete",
  "settings.keychain.userNote": "User note",
  "settings.keychain.agentNote": "Agent-readable note",
  "settings.keychain.userNoteHint": "Human/UI note only",
  "settings.keychain.agentNoteHint": "Safe guidance for agents",
  "settings.keychain.saveNotes": "Save notes",
  "settings.keychain.masterPassword": "Master password:",
  "settings.keychain.masterPasswordPlaceholder": "Enter keychain master password",
  "settings.keychain.unlock": "Unlock",
  "settings.keychain.totpCode": "TOTP code:",
  "settings.keychain.verify": "Verify",
  "settings.keychain.username": "Username",
  "settings.keychain.copyUsername": "Copy username",
  "settings.keychain.secret": "Secret",
  "settings.keychain.copySecret": "Copy secret",
  "settings.tasks.internalProtected": "internal/protected",
  "settings.tasks.noRunLogs": "No run logs recorded yet.",
  "settings.tasks.noSummary": "No summary",
  "settings.tasks.selectPrompt": "Select a task to inspect schedule, status, and run history.",
  "settings.tasks.pause": "Pause",
  "settings.tasks.resume": "Resume",
  "settings.tasks.delete": "Delete",
  "settings.tasks.status": "Status",
  "settings.tasks.kind": "Kind",
  "settings.tasks.schedule": "Schedule",
  "settings.tasks.nextRun": "Next run",
  "settings.tasks.lastRun": "Last run",
  "settings.tasks.lastResult": "Last result",
  "settings.tasks.chat": "Chat",
  "settings.tasks.model": "Model",
  "settings.tasks.cwd": "CWD",
  "settings.tasks.timeout": "Timeout",
  "settings.tasks.protection": "Protection",
  "settings.tasks.protectionHint": "Internal task actions require explicit confirmation.",
  "settings.tasks.command": "Command",
  "settings.tasks.prompt": "Prompt",
  "settings.tasks.recentRuns": "Recent runs",
  "settings.tasks.activeLabel": "Active",
  "settings.tasks.pausedLabel": "Paused",
  "settings.tasks.completedLabel": "Completed",
  "settings.tasks.allStatuses": "All statuses",
  "settings.tasks.filterChatPlaceholder": "Filter chat JID…",
  "settings.tasks.refresh": "Refresh",
  "settings.tasks.loading": "Loading scheduled tasks…",
  "settings.tasks.noneFound": "No scheduled tasks found.",
  "settings.tasks.noneFoundHint": "Tasks created with reminders, `/tasks`, or the scheduler tool will appear here.",
  "settings.tasks.listLabel": "Scheduled tasks",
  "settings.tasks.next": "Next",
  "settings.tasks.last": "Last",
  "settings.tasks.noMatch": "No tasks match “{filter}”.",
  "settings.tasks.confirmDelete": "Delete scheduled task {id}?",
  "settings.tasks.confirmPause": "Pause scheduled task {id}?",
  "settings.tasks.confirmResume": "Resume scheduled task {id}?",
  "settings.tasks.confirmProtected": "Task {id} is internal/protected. Continue with {action}?",
  "settings.tasks.deleting": "Deleting {id}…",
  "settings.tasks.pausing": "Pausing {id}…",
  "settings.tasks.resuming": "Resuming {id}…",
  "settings.tasks.deletedToast": "Scheduled task {id} deleted.",
  "settings.tasks.pausedToast": "Scheduled task {id} paused.",
  "settings.tasks.resumedToast": "Scheduled task {id} resumed.",
  "settings.tasks.actionFailed": "Failed to {action} task.",
  "settings.tasks.loadFailed": "Failed to load scheduled tasks.",
  "settings.compaction.appliedNotice": "Compaction settings applied. Existing turns keep their current timers; new turns use the updated values.",
  "settings.compaction.saving": "Saving compaction settings…",
  "settings.compaction.saveFailed": "Failed to save compaction settings.",
  "settings.compaction.saved": "Compaction settings saved.",
  "settings.compaction.clearing": "Clearing compaction suppression for {chat}…",
  "settings.compaction.clearFailed": "Failed to clear compaction suppression.",
  "settings.compaction.cleared": "Cleared compaction suppression for {chat}.",
  "settings.compaction.autoHeading": "Automatic compaction",
  "settings.compaction.enableAutomatic": "Enable automatic compaction",
  "settings.compaction.enableAutomaticHint": "Piclaw-managed pre-prompt/idle compaction. The upstream agent auto-compactor stays suppressed internally.",
  "settings.compaction.processingMethod": "Processing method",
  "settings.compaction.model": "Compaction model",
  "settings.compaction.modelHint": "Strict local smart-compaction model. If configured but unavailable, compaction stops and preserves the session instead of falling back.",
  "settings.compaction.modelPlaceholder": "provider/model (empty uses active model)",
  "settings.compaction.methodSelective": "Selective",
  "settings.compaction.methodSelectiveHint": "Extract high-value continuity excerpts, using complete progressive coverage whenever a bounded prompt cannot represent every discarded source event.",
  "settings.compaction.methodPipelined": "Pipelined",
  "settings.compaction.methodPipelinedHint": "Canonicalize and classify every discarded source event with an auditable coverage ledger before summarizing.",
  "settings.compaction.remoteNative": "Provider-native compaction",
  "settings.compaction.remoteNativeHint": "Opt-in for explicitly supported providers only ({providers}). Any failure falls back atomically to the selected local method.",
  "settings.compaction.remoteTimeout": "Provider-native timeout (sec)",
  "settings.compaction.remoteTimeoutAria": "provider-native compaction timeout",
  "settings.compaction.remoteTimeoutHint": "Deadline for the remote pre-pass before local fallback.",
  "settings.compaction.enableToolResult": "Enable tool-result compaction",
  "settings.compaction.enableToolResultHint": "When disabled, large tool results stay inline and are not externalized into searchable tool-output handles.",
  "settings.compaction.semanticSummaries": "Semantic summaries for compacted tool results",
  "settings.compaction.semanticSummariesHint": "When enabled, compacted outputs include a semantic summary generated with the active model (preview fallback on failure).",
  "settings.compaction.inputLimit": "Semantic summary input limit (chars)",
  "settings.compaction.inputLimitAria": "semantic summary input limit",
  "settings.compaction.inputLimitHint": "Maximum characters sampled from full tool output for semantic summarization.",
  "settings.compaction.maxTokens": "Semantic summary output max tokens",
  "settings.compaction.maxTokensAria": "semantic summary max tokens",
  "settings.compaction.maxTokensHint": "Upper bound for generated summary length.",
  "settings.compaction.summaryTimeout": "Semantic summary timeout (sec)",
  "settings.compaction.summaryTimeoutAria": "semantic summary timeout",
  "settings.compaction.summaryTimeoutHint": "Abort semantic summary generation after this timeout and fall back to preview compaction.",
  "settings.compaction.threshold": "Compaction threshold (%)",
  "settings.compaction.thresholdAria": "compaction threshold",
  "settings.compaction.thresholdHint": "auto-compact when context exceeds this % of window",
  "settings.compaction.timeout": "Compaction timeout (sec)",
  "settings.compaction.timeoutAria": "compaction timeout",
  "settings.compaction.timeoutHint": "Single wall-clock deadline for deterministic preparation, provider prefill/streaming, and settlement. Local provider requests inherit the remaining time.",
  "settings.compaction.backoffBase": "Failure backoff base (min)",
  "settings.compaction.backoffBaseAria": "compaction backoff base",
  "settings.compaction.backoffBaseHint": "First suppression window after a compaction failure.",
  "settings.compaction.backoffMax": "Failure backoff max (min)",
  "settings.compaction.backoffMaxAria": "compaction backoff max",
  "settings.compaction.backoffMaxHint": "Upper bound for exponential suppression after repeated failures.",
  "settings.compaction.decayFactor": "Backoff decay factor",
  "settings.compaction.decayFactorAria": "backoff decay factor",
  "settings.compaction.decayFactorHint": "% — halves backoff after each successful compaction",
  "settings.compaction.watchdogHeading": "Stall watchdog",
  "settings.compaction.enableWatchdog": "Enable watchdog",
  "settings.compaction.enableWatchdogHint": "Disabled by default. When enabled, a helper process terminates the runtime if an active phase stops heartbeating.",
  "settings.compaction.watchdogTimeout": "Watchdog timeout (sec)",
  "settings.compaction.watchdogTimeoutAria": "watchdog timeout",
  "settings.compaction.watchdogTimeoutHint": "How long an active phase can go without a heartbeat before the watchdog kills the runtime.",
  "settings.compaction.suppressionsHeading": "Active compaction suppressions",
  "settings.compaction.noBackoff": "No chats are currently under compaction backoff.",
  "settings.compaction.clear": "Clear",
  "settings.compaction.phasesHeading": "Live watchdog phases",
  "settings.compaction.noPhases": "No active tracked phases right now.",
  "menu.title": "Menu",
  "menu.showWorkspace": "Show workspace",
  "menu.hideWorkspace": "Hide workspace",
  "menu.openExplorer": "Open explorer",
  "menu.chatOnly": "Chat-only mode",
  "menu.exitChatOnly": "Exit chat-only mode",
  "menu.openTerminal": "Open terminal in tab",
  "menu.openVnc": "Open VNC in tab",
  "menu.newFile": "New file",
  "menu.openRecent": "Open Recent",
  "menu.refreshTree": "Refresh tree",
  "menu.reindex": "Reindex workspace",
  "menu.showHidden": "Show hidden files",
  "menu.hideHidden": "Hide hidden files",
  "menu.scale": "Scale",
  "menu.settings": "Settings"
};
var ZH_CN = {
  "compose.placeholder": "输入消息（回车发送，Shift+回车换行）...",
  "compose.send": "发送",
  "compose.stop": "停止",
  "compose.searchPlaceholder": "搜索（回车运行）...",
  "compose.clearAll": "清除全部",
  "compose.clearAllTitle": "清除所有附件和引用",
  "compose.scope": "范围",
  "compose.searchScope": "搜索范围",
  "compose.scopeCurrent": "当前",
  "compose.scopeBranchFamily": "分支系列",
  "compose.scopeAll": "所有聊天",
  "compose.filterImages": "图片",
  "compose.filterAttachments": "附件",
  "compose.search": "搜索",
  "compose.closeSearch": "关闭搜索",
  "compose.shareLocation": "分享位置",
  "compose.attachFile": "附加文件",
  "compose.queueControls": "排队后续消息控制",
  "compose.moveUp": "上移",
  "compose.moveUpQueue": "在队列中上移",
  "compose.moveDown": "下移",
  "compose.moveDownQueue": "在队列中下移",
  "compose.editInCompose": "在输入框中编辑",
  "compose.returnToEditor": "将排队消息返回编辑器",
  "compose.injectSteer": "作为引导插入排队的后续消息",
  "compose.steer": "引导",
  "compose.cancelQueued": "取消排队消息",
  "compose.resizeInput": "调整消息输入框大小",
  "compose.resizeInputHint": "拖动以调整消息输入框大小",
  "compose.modelPicker": "模型选择器",
  "compose.sessionsAndAgents": "会话与代理",
  "compose.openModelPicker": "打开模型选择器",
  "compose.newBranchTitle": "从此聊天创建新分支",
  "compose.newRootTitle": "创建一个干净的根会话，例如 web:ops",
  "compose.renameSessionTitle": "重命名当前会话",
  "compose.pruneSessionTitle": "删除（修剪）当前代理/会话分支",
  "compose.filterImagesTitle": "仅显示含图片的消息",
  "compose.filterAttachmentsTitle": "仅显示含附件的消息",
  "compose.selectModel": "选择模型",
  "compose.loadingModels": "正在加载模型…",
  "compose.noModels": "没有可用的模型。",
  "compose.nextModel": "下一个模型",
  "compose.manageSessions": "管理会话与代理",
  "compose.noSessions": "暂无其他会话。",
  "compose.newBranch": "新建分支",
  "compose.newRoot": "新建根会话…",
  "compose.mergeCurrent": "将当前合并到父级",
  "compose.renameCurrent": "重命名当前…",
  "compose.deleteCurrent": "删除当前…",
  "compose.mergeInto": "将此分支合并到 {target}",
  "compose.mergeBlocked": "当此分支处于活动状态或有子分支时无法合并",
  "workspace.title": "工作区",
  "workspace.moveConfirm": "将{entry}“{name}”从{source}移动到{target}？",
  "workspace.root": "工作区根目录",
  "workspace.file": "文件",
  "workspace.folder": "文件夹",
  "workspace.newFile": "新建文件",
  "workspace.refresh": "刷新",
  "workspace.actions": "工作区操作",
  "workspace.uploadFiles": "上传文件",
  "workspace.reindexing": "正在重建索引…",
  "workspace.deleteFile": "删除文件",
  "workspace.download": "下载",
  "workspace.uploadToFolder": "上传文件到此文件夹",
  "workspace.addFolderHint": "将文件夹提示添加到输入框",
  "workspace.downloadZip": "将文件夹下载为 zip",
  "workspace.openInTab": "在标签页打开",
  "workspace.openInEditor": "在编辑器打开",
  "workspace.renameSelected": "重命名所选",
  "workspace.downloadSelectedFile": "下载所选文件",
  "workspace.downloadSelectedFolder": "下载所选文件夹（zip）",
  "workspace.deleteSelectedFile": "删除所选文件",
  "shell.settings": "设置",
  "shell.newChat": "新建对话",
  "shell.connecting": "连接中…",
  "shell.connected": "已连接",
  "language.label": "语言",
  "settings.title": "设置",
  "settings.close": "关闭（Esc）",
  "settings.filter": "筛选…",
  "settings.loading": "加载设置中…",
  "settings.section.general": "常规",
  "settings.section.sessions": "会话",
  "settings.section.recordings": "录制",
  "settings.section.compaction": "压缩",
  "settings.section.budget": "预算",
  "settings.section.keyboard": "键盘",
  "settings.section.workspace": "工作区",
  "settings.section.environment": "环境",
  "settings.section.providers": "提供商",
  "settings.section.models": "模型",
  "settings.section.theme": "外观",
  "settings.section.scheduled-tasks": "计划任务",
  "settings.section.quick-actions": "快捷操作",
  "settings.section.keychain": "密钥串",
  "settings.section.tools": "工具",
  "settings.section.addons": "插件",
  "settings.placeholder.recordings": "筛选录制…",
  "settings.placeholder.keyboard": "筛选快捷键…",
  "settings.placeholder.environment": "筛选环境…",
  "settings.placeholder.models": "筛选模型…",
  "settings.placeholder.scheduled-tasks": "筛选计划任务…",
  "settings.placeholder.quick-actions": "筛选快捷操作…",
  "settings.placeholder.keychain": "筛选条目…",
  "settings.placeholder.tools": "筛选工具…",
  "settings.placeholder.addons": "筛选插件…",
  "preview.close": "关闭",
  "preview.loading": "正在加载预览…",
  "preview.files": "文件",
  "preview.folders": "文件夹",
  "preview.compressed": "压缩后",
  "preview.uncompressed": "未压缩",
  "preview.name": "名称",
  "preview.type": "类型",
  "preview.method": "方法",
  "preview.size": "大小",
  "post.deleteMessage": "删除消息",
  "post.tooLarge": "消息过大，无法显示。",
  "post.previewTruncated": "预览已截断。",
  "post.submitted": "已提交",
  "post.discard": "丢弃",
  "post.save": "保存",
  "post.cancel": "取消",
  "post.addNote": "添加备注",
  "post.addNotePlaceholder": "添加备注…",
  "post.restartNotice": "正在重启 — 原因：{reason}",
  "post.restartCompleted": "重启完成。",
  "post.agentSelfResume": "代理自行恢复",
  "tab.close": "关闭",
  "tab.closeOthers": "关闭其他",
  "tab.closeAll": "全部关闭",
  "tab.reattach": "重新附加",
  "tab.openInWindow": "在窗口中打开",
  "tab.openInNewTab": "在新标签页打开",
  "tab.pinned": "已固定",
  "tab.detached": "已分离",
  "tab.openSeparateWindow": "在独立窗口中打开",
  "status.trackedVariables": "跟踪的变量",
  "status.attachToSession": "附加到会话",
  "status.files": "文件",
  "status.proposedDiff": "建议的差异",
  "status.copyTmux": "复制 tmux 命令",
  "status.experimentDuration": "实验时长",
  "status.sinceLastActivity": "自上次活动以来",
  "annotator.title": "标注图片",
  "annotator.typeLabel": "输入标签…",
  "annotator.undo": "撤销",
  "annotator.resetZoom": "重置缩放",
  "tree.filter": "筛选…",
  "tree.sessionTree": "会话树",
  "btw.label": "BTW 附加对话",
  "btw.close": "关闭 BTW",
  "btw.thinking": "思考中",
  "mdpreview.close": "关闭预览",
  "mdpreview.unavailable": "预览不可用",
  "widget.close": "关闭小部件",
  "oobe.gettingStarted": "入门指南",
  "oobe.needsSetupTitle": "实例需要设置",
  "oobe.configuredTitle": "实例已配置",
  "oobe.needsSetupBody": "此实例尚未配置。请打开“设置”并设置 AI 提供商/模型以开始发送请求。",
  "oobe.configuredBody": "此实例看起来已配置。请在“设置”中查看或更新提供商和模型设置。",
  "oobe.openSettings": "打开设置",
  "oobe.dismiss": "忽略",
  "oobe.done": "完成",
  "palette.placeholder": "输入以跳转到代理、工作区操作或斜杠命令…",
  "palette.hideWorkspace": "隐藏工作区",
  "palette.showWorkspace": "显示工作区",
  "palette.hideWorkspaceDesc": "隐藏工作区侧边栏。",
  "palette.showWorkspaceDesc": "显示工作区侧边栏。",
  "palette.exitChatOnly": "退出仅聊天模式",
  "palette.chatOnly": "仅聊天模式",
  "palette.exitChatOnlyDesc": "返回分屏工作区布局。",
  "palette.chatOnlyDesc": "切换到仅聊天布局。",
  "palette.groupAgents": "代理",
  "palette.groupWorkspace": "工作区",
  "palette.groupSlash": "斜杠命令",
  "palette.hintMove": "移动",
  "palette.hintSelect": "选择",
  "palette.hintPopOut": "弹出",
  "palette.hintClose": "关闭",
  "settings.appliedNotice": "设置已应用。更改将在下一回合生效。",
  "settings.sessions.lifecycle": "会话生命周期",
  "settings.sessions.autoRotate": "自动轮换会话",
  "settings.sessions.maxSize": "最大会话大小（MB）",
  "settings.sessions.maxSizeAria": "最大会话大小",
  "settings.sessions.agentBehaviour": "代理行为",
  "settings.sessions.toolBudget": "工具使用预算",
  "settings.sessions.toolBudgetAria": "工具使用预算",
  "settings.sessions.toolBudgetHint": "每回合最大已完成工具执行次数",
  "settings.sessions.isolation": "会话隔离",
  "settings.sessions.isolationNone": "无 — 完全跨会话可见",
  "settings.sessions.isolationSummary": "摘要 — 工具可见，无参数",
  "settings.sessions.isolationFull": "完全 — 会话之间不可见",
  "settings.editor.heading": "编辑器",
  "settings.editor.vimMode": "Vim 模式",
  "settings.editor.showWhitespace": "显示空白字符",
  "settings.editor.livePreview": "Markdown 实时预览",
  "settings.editor.fontSize": "字号（px）",
  "settings.editor.fontSizeAria": "编辑器字号",
  "settings.editor.fontFamily": "字体",
  "settings.editor.fontFamilyPlaceholder": "monospace（默认）",
  "settings.editor.localOnlyHint": "仅限此浏览器。编辑器更改存储在本地浏览器中，并在下次打开或重新加载文件标签页时生效。",
  "settings.appearance.syncing": "正在同步外观…",
  "settings.appearance.default": "默认",
  "settings.appearance.autoLightDark": "自动（浅色/深色）",
  "settings.appearance.tint": "色调：",
  "settings.appearance.clearTint": "清除色调",
  "settings.appearance.none": "无",
  "settings.appearance.outputPadding": "输出内边距",
  "settings.appearance.outputPaddingHint": "消息和思考面板周围的额外空间。",
  "settings.keyboard.heading": "键盘",
  "settings.keyboard.hint1": "将应用级快捷键自定义为逗号分隔的绑定。更改立即生效。",
  "settings.keyboard.hint1b": "已保留用于关闭/中止，无法重新绑定。",
  "settings.keyboard.hint2mid": "以及键入",
  "settings.keyboard.hint2end": "（在输入框外）可打开此面板。",
  "settings.keyboard.resetAll": "全部重置为默认",
  "settings.keyboard.defaultColon": "默认：",
  "settings.keyboard.save": "保存",
  "settings.keyboard.defaultBtn": "默认",
  "settings.keyboard.noMatch": "没有匹配此筛选的快捷键。",
  "settings.keyboard.invalidShortcut": "无效快捷键：{token}。Escape 已保留，无法重新绑定。",
  "settings.keyboard.saved": "快捷键已保存。",
  "settings.keyboard.resetOne": "快捷键已重置为默认。",
  "settings.keyboard.resetAllDone": "快捷键已全部重置为默认。",
  "settings.workspace.serverApplied": "工作区设置已应用。服务器端限制立即影响新的工作区请求。",
  "settings.workspace.browserApplied": "浏览器工作区设置已在此标签页立即应用。",
  "settings.workspace.access": "访问",
  "settings.workspace.enableTerminal": "启用 Web 终端",
  "settings.workspace.allowVnc": "允许直接 VNC 目标",
  "settings.workspace.accessHint": "终端访问立即更新。直接 VNC 目标策略适用于新的 VNC 请求。",
  "settings.workspace.guardrails": "服务器扫描防护",
  "settings.workspace.maxDepth": "最大树深度",
  "settings.workspace.maxDepthAria": "工作区树最大深度",
  "settings.workspace.maxDepthHintPre": "限制所有",
  "settings.workspace.maxDepthHintPost": "请求",
  "settings.workspace.maxEntries": "每次扫描最大条目数",
  "settings.workspace.maxEntriesAria": "工作区树最大条目数",
  "settings.workspace.maxEntriesHint": "更早截断超大的树遍历",
  "settings.workspace.thisBrowser": "此浏览器",
  "settings.workspace.refreshInterval": "刷新间隔（秒）",
  "settings.workspace.refreshIntervalAria": "工作区刷新间隔",
  "settings.workspace.folderDepth": "文件夹预览扫描深度",
  "settings.workspace.folderDepthAria": "文件夹预览扫描深度",
  "settings.workspace.folderDepthHintPre": "设为",
  "settings.workspace.folderDepthHintPost": "以禁用文件夹大小预览扫描",
  "settings.workspace.footerHint": "根目录和文件夹展开的树加载保持较浅；文件夹大小预览是 UI 中最深的工作区扫描。",
  "settings.models.thinkingLevel": "思考级别",
  "settings.models.noThinking": "当前模型不支持思考。",
  "settings.models.thinkingLevelLabel": "思考级别：",
  "settings.models.loading": "正在加载模型…",
  "settings.models.summary": "在狭窄面板中，模型和提供商名称可能换行以避免裁切。",
  "settings.models.scopedOnly": "仅限范围内模型",
  "settings.models.scopedCheckboxPre": "使用 Pi 的",
  "settings.models.scopedCheckboxPost": "作为 Piclaw 模型列表",
  "settings.models.scopedHintPre": "筛选此选择器和",
  "settings.models.scopedHintPost": "工具。TUI 模型选择保持不变。",
  "settings.models.colModel": "模型",
  "settings.models.colProvider": "提供商",
  "settings.models.colContext": "上下文",
  "settings.models.colReasoning": "推理",
  "settings.models.noMatch": "没有匹配 “{filter}” 的模型",
  "settings.tools.unavailable": "工具数据不可用。",
  "settings.tools.search": "搜索",
  "settings.tools.matchMode": "匹配模式",
  "settings.tools.orMode": "任意关键词（OR）— 结果至少匹配一个搜索词",
  "settings.tools.andMode": "所有关键词（AND）— 结果必须匹配每个搜索词",
  "settings.tools.colEnabled": "已启用",
  "settings.tools.colTool": "工具",
  "settings.tools.colCompact": "结果压缩",
  "settings.tools.colKind": "类型",
  "settings.tools.colSummary": "摘要",
  "settings.tools.colSource": "来源",
  "settings.tools.disableCompaction": "为此工具禁用工具结果压缩",
  "settings.tools.enableCompaction": "为此工具启用工具结果压缩",
  "settings.tools.noMatch": "没有匹配 “{filter}” 的工具",
  "settings.tools.footer": "工具激活由代理运行时管理。组复选框可折叠/展开；“压缩”列控制工具结果压缩资格。",
  "settings.environment.heading": "环境",
  "settings.environment.introPre": "仅显示非 keychain 环境变量。覆盖项存储在扩展 KV 中并应用于",
  "settings.environment.introPost": "，因此后续工具调用会继承它们。",
  "settings.environment.refresh": "刷新",
  "settings.environment.addOverride": "添加覆盖",
  "settings.environment.valuePlaceholder": "值",
  "settings.environment.save": "保存",
  "settings.environment.countLine": "{count} 个变量可见 • {overrides} 个覆盖生效 • {keychain} 个 keychain 注入变量已隐藏",
  "settings.environment.overridden": "在 KV 中覆盖",
  "settings.environment.inherited": "继承自进程环境",
  "settings.environment.kindOverride": "覆盖",
  "settings.environment.kindProcess": "进程",
  "settings.environment.clear": "清除",
  "settings.environment.noMatch": "没有匹配 “{filter}” 的环境变量。",
  "settings.environment.refreshedToast": "环境已刷新。",
  "settings.environment.savedToast": "已保存 {name} 的环境覆盖。",
  "settings.environment.clearedToast": "已清除 {name} 的环境覆盖。",
  "settings.quickActions.loading": "加载中…",
  "settings.quickActions.heading": "时间线快捷操作",
  "settings.quickActions.intro": "选择哪些操作出现在时间线预输入中。代理始终优先固定，然后是工作区命令，再是斜杠命令。",
  "settings.quickActions.enableAll": "全部启用",
  "settings.quickActions.saving": "保存中…",
  "settings.quickActions.saveApply": "保存并应用",
  "settings.quickActions.workspaceCommands": "工作区命令",
  "settings.quickActions.noWorkspaceMatch": "没有匹配此筛选的工作区命令。",
  "settings.quickActions.slashCommands": "斜杠命令",
  "settings.quickActions.slashFallback": "斜杠命令",
  "settings.quickActions.noSlashMatch": "没有匹配此筛选的斜杠命令。",
  "settings.quickActions.savingToast": "正在保存快捷操作…",
  "settings.quickActions.savedToast": "快捷操作已保存。",
  "settings.providers.authApiKey": "API 密钥",
  "settings.providers.authConfigured": "已配置",
  "settings.providers.heading": "提供商",
  "settings.providers.tagCustom": "自定义",
  "settings.providers.logout": "注销",
  "settings.providers.reconfigure": "重新配置",
  "settings.providers.setUp": "设置",
  "settings.providers.setupHint": "登录流程在浏览器中打开。在狭窄面板中，设置表单会垂直堆叠以避免裁切。",
  "settings.providers.starting": "启动中…",
  "settings.providers.signInOAuth": "使用 OAuth 登录",
  "settings.providers.apiKeyLabel": "API 密钥",
  "settings.providers.apiKeyPlaceholder": "输入 API 密钥",
  "settings.providers.save": "保存",
  "settings.providers.configuring": "配置中…",
  "settings.providers.saveConfig": "保存配置",
  "settings.providers.apiKeyEmpty": "API 密钥不能为空。",
  "settings.providers.configuringToast": "正在配置 {provider}…",
  "settings.providers.configured": "{provider} 已配置。",
  "settings.providers.startingOAuth": "正在为 {provider} 启动 OAuth…",
  "settings.providers.oauthOpened": "OAuth 窗口已打开。完成登录流程，然后关闭此消息。",
  "settings.providers.oauthStarted": "已为 {provider} 启动 OAuth 流程。请查看聊天。",
  "settings.providers.loggingOut": "正在注销 {provider}…",
  "settings.providers.loggedOut": "已注销 {provider}。可能需要重启。",
  "settings.general.identity": "身份",
  "settings.general.userLabel": "用户",
  "settings.general.yourName": "你的名字",
  "settings.general.agentLabel": "代理",
  "settings.general.agentName": "代理名称",
  "settings.general.notifications": "通知",
  "settings.general.browserNotifications": "浏览器通知",
  "settings.general.notifSecureHint": "使用输入栏中的 \uD83D\uDD14 铃铛按钮来启用/禁用通知。Web Push 需要 HTTPS 或 localhost。",
  "settings.general.notifInsecureHint": "⚠ 不可用 — 需要安全上下文（HTTPS 或 localhost）。通过 SSH 隐道或带 TLS 的反向代理访问以启用。",
  "settings.general.display": "显示",
  "settings.general.systemMeters": "系统仪表",
  "settings.general.systemMetersHint": "状态栏中的 CPU/内存/网络仪表。仅限此浏览器。",
  "settings.general.instanceConfig": "实例配置",
  "settings.general.composeUpload": "撰写上传（MB）",
  "settings.general.composeUploadAria": "撰写上传限制",
  "settings.general.composeUploadHint": "聊天/媒体附件",
  "settings.general.workspaceUpload": "工作区上传（MB）",
  "settings.general.workspaceUploadAria": "工作区上传限制",
  "settings.general.workspaceUploadHint": "默认为 256 MB；分块上传最多允许 1 GB",
  "settings.general.agentRecovery": "高级 · 代理恢复",
  "settings.general.automaticRecovery": "自动恢复",
  "settings.general.automaticRecoveryHint": "自动重试可恢复的失败回合。",
  "settings.general.recoveryMaxAttempts": "最大尝试次数",
  "settings.general.recoveryMaxAttemptsAria": "自动恢复最大尝试次数",
  "settings.general.recoveryMaxAttemptsHint": "0 表示继承常规重试限制。",
  "settings.general.recoveryTotalBudget": "总预算（毫秒）",
  "settings.general.recoveryTotalBudgetAria": "自动恢复总预算（毫秒）",
  "settings.general.recoveryTotalBudgetHint": "0 表示取回合超时的三分之一，并限制在 6–60 分钟；正数为明确上限。",
  "settings.general.recoveryEffectiveBudget": "默认回合超时的有效预算：{budget} 毫秒。",
  "settings.general.authentication": "身份验证",
  "settings.general.widgetToken": "小部件 bearer 令牌",
  "settings.general.token": "令牌",
  "settings.general.hideToken": "隐藏令牌",
  "settings.general.revealToken": "显示令牌",
  "settings.general.copyToken": "复制令牌",
  "settings.general.copied": "已复制",
  "settings.general.regenerating": "正在重新生成…",
  "settings.general.regenerate": "重新生成",
  "settings.general.tokenHintPre": "只读令牌，用于",
  "settings.general.tokenHintMid": "和",
  "settings.general.tokenHintPost": "。用作",
  "settings.general.tokenHintEnd": "。",
  "settings.general.copyFailed": "无法复制小部件令牌。请选择令牌字段并手动复制。",
  "settings.general.regenConfirm": "重新生成小部件令牌？使用旧令牌的现有 macOS 小部件将停止更新。",
  "settings.general.totpTitle": "TOTP 设置二维码",
  "settings.general.totpConfiguredHint": "当前 Web 登录验证器密钥。扫描此二维码以添加另一个验证器设备。",
  "settings.general.totpUnconfiguredHint": "此实例尚未配置 TOTP，因此没有可用的设置二维码。",
  "settings.general.issuer": "颁发者",
  "settings.general.label": "标签",
  "settings.general.secret": "密钥",
  "settings.general.avatarUpload": "点击上传",
  "settings.developer.heading": "开发者",
  "settings.developer.devMode": "开发者模式",
  "settings.developer.localHint": "仅限此浏览器。开发者模式开关和插件目录覆盖存储在本地浏览器存储中。",
  "settings.developer.addonSources": "插件来源",
  "settings.developer.catalogUrl": "目录 URL",
  "settings.developer.catalogHint": "主插件目录 URL。留空以使用默认值",
  "settings.developer.additionalCatalogs": "其他目录 URL",
  "settings.developer.additionalHint": "在主/默认目录之外额外获取。每行一个 URL。",
  "settings.developer.repoUrl": "仓库 URL",
  "settings.developer.repoHintPre": "覆盖用于",
  "settings.developer.repoHintPost": "安装的 git 仓库。留空以使用默认值。",
  "settings.developer.debug": "调试",
  "settings.developer.logSse": "记录 SSE 事件",
  "settings.developer.logToolCalls": "记录工具调用",
  "settings.developer.debugHint": "调试标志在下次页面重新加载时生效。",
  "settings.addons.installing": "正在安装 {slug}…",
  "settings.addons.removing": "正在移除 {slug}…",
  "settings.addons.installedToast": "插件已安装。",
  "settings.addons.removedToast": "插件已移除。",
  "settings.addons.restarting": "正在重启 piclaw…",
  "settings.addons.restartComplete": "重启完成 — 插件已刷新。",
  "settings.addons.restartTimeout": "后端未能及时返回。请手动重新加载页面。",
  "settings.addons.fetching": "正在获取插件…",
  "settings.addons.loadFailed": "无法加载插件。",
  "settings.addons.catalogFromPre": "目录来自",
  "settings.addons.catalogMerged": "已合并 {count} 个目录来源。",
  "settings.addons.installNote": "通过 Bun 优先安装包；安装/卸载后需要重启。",
  "settings.addons.failedFetchSingular": "获取 {count} 个目录来源失败：",
  "settings.addons.failedFetchPlural": "获取 {count} 个目录来源失败：",
  "settings.addons.activeSources": "活动目录来源（{count}）",
  "settings.addons.windowsWarning": "原生 Windows 插件安装风险更高：Bun 包安装、符号链接清理、锁定文件和重启时机都可能不如 Linux/WSL 可预测。如果可能，请优先使用 WSL 或容器。",
  "settings.addons.typeExtSkill": "扩展 + 技能",
  "settings.addons.typeSkill": "技能",
  "settings.addons.typeExt": "扩展",
  "settings.addons.update": "更新",
  "settings.addons.remove": "移除",
  "settings.addons.install": "安装",
  "settings.addons.noMatch": "没有匹配 “{filter}” 的插件",
  "settings.addons.restartNotice": "扩展更改已安装，但在 piclaw 重启之前处于非活动状态。",
  "settings.addons.restartNow": "立即重启",
  "settings.recordings.modeFull": "完整 / 受信任",
  "settings.recordings.modeMetadata": "仅元数据",
  "settings.recordings.modeRedacted": "已脱敏",
  "settings.recordings.selectPrompt": "选择一个录制以检查、回放、导出或删除。",
  "settings.recordings.playback": "回放",
  "settings.recordings.refresh": "刷新",
  "settings.recordings.delete": "删除",
  "settings.recordings.status": "状态",
  "settings.recordings.mode": "模式",
  "settings.recordings.chat": "聊天",
  "settings.recordings.started": "开始",
  "settings.recordings.ended": "结束",
  "settings.recordings.events": "事件",
  "settings.recordings.redactions": "脱敏",
  "settings.recordings.exportJson": "导出 JSON",
  "settings.recordings.exportJsonl": "导出 JSONL",
  "settings.recordings.exportHtml": "导出独立 HTML",
  "settings.recordings.eventSummary": "事件摘要",
  "settings.recordings.inspectHint": "打开或刷新详情以检查跟踪事件。",
  "settings.recordings.firstEvents": "首批事件",
  "settings.recordings.heading": "会话录制",
  "settings.recordings.intro": "选择性加入的跟踪捕获，用于确定性回放和屏幕录制导出。回放绝不会调用实时代理或工具端点。",
  "settings.recordings.chatJid": "聊天 JID",
  "settings.recordings.title": "标题",
  "settings.recordings.titlePlaceholder": "演示录制",
  "settings.recordings.modeLabelField": "模式",
  "settings.recordings.optRedacted": "已脱敏",
  "settings.recordings.optMetadata": "仅元数据",
  "settings.recordings.optFull": "完整 / 受信任本地",
  "settings.recordings.includeSnapshot": "包含时间线快照",
  "settings.recordings.extraKeys": "额外脱敏键",
  "settings.recordings.extraPatterns": "额外正则模式",
  "settings.recordings.stopCurrent": "停止当前聊天录制",
  "settings.recordings.start": "开始录制",
  "settings.recordings.redactionPreview": "脱敏预览",
  "settings.recordings.previewRedaction": "预览脱敏",
  "settings.recordings.loading": "正在加载录制…",
  "settings.recordings.noneYet": "还没有录制。",
  "settings.recordings.noneYetHint": "在上方开始录制，然后使用回放/导出进行确定性屏幕捕获。",
  "settings.recordings.listLabel": "会话录制",
  "settings.recordings.eventsCount": "{count} 个事件",
  "settings.recordings.noMatch": "没有匹配 “{filter}” 的录制。",
  "settings.recordings.startedToast": "已为 {chat} 开始录制。",
  "settings.recordings.startFailed": "开始录制失败。",
  "settings.recordings.stoppedToast": "已为 {chat} 停止录制。",
  "settings.recordings.stopFailed": "停止录制失败。",
  "settings.recordings.deleteConfirm": "删除录制 {id}？",
  "settings.recordings.deletedToast": "录制已删除。",
  "settings.recordings.deleteFailed": "删除录制失败。",
  "settings.recordings.loadOneFailed": "加载录制失败。",
  "settings.recordings.loadFailed": "加载录制失败。",
  "settings.recordings.previewFailed": "预览失败。",
  "settings.keychain.loadFailed": "加载密钥链失败。",
  "settings.keychain.addFailed": "添加条目失败。",
  "settings.keychain.deleteFailed": "删除条目失败。",
  "settings.keychain.saveNotesFailed": "保存备注失败。",
  "settings.keychain.revealFailed": "显示失败。",
  "settings.keychain.loading": "正在加载密钥链…",
  "settings.keychain.entryCountSingular": "{count} 个条目",
  "settings.keychain.entryCountPlural": "{count} 个条目",
  "settings.keychain.matchingFilter": ' 匹配 "{filter}"',
  "settings.keychain.encryptedSuffix": "，静态加密。",
  "settings.keychain.clickPrefix": "点击",
  "settings.keychain.revealSuffix": "以显示。",
  "settings.keychain.cancel": "取消",
  "settings.keychain.addEntry": "+ 添加条目",
  "settings.keychain.namePlaceholder": "条目名称（例如 github/my-token）",
  "settings.keychain.secretPlaceholder": "密钥值",
  "settings.keychain.usernamePlaceholder": "用户名（可选）",
  "settings.keychain.saving": "正在保存…",
  "settings.keychain.save": "保存",
  "settings.keychain.userNotePlaceholder": "用户备注（仅在此界面可见）",
  "settings.keychain.agentNotePlaceholder": "代理备注（可安全暴露给代理）",
  "settings.keychain.noMatchFilter": "没有条目匹配筛选条件。",
  "settings.keychain.noEntries": "没有密钥链条目。",
  "settings.keychain.hideSecret": "隐藏密钥",
  "settings.keychain.revealSecret": "显示密钥",
  "settings.keychain.deleteQ": "删除？",
  "settings.keychain.yes": "是",
  "settings.keychain.no": "否",
  "settings.keychain.deleteTitle": "删除",
  "settings.keychain.userNote": "用户备注",
  "settings.keychain.agentNote": "代理可读备注",
  "settings.keychain.userNoteHint": "仅限人工/界面备注",
  "settings.keychain.agentNoteHint": "给代理的安全指引",
  "settings.keychain.saveNotes": "保存备注",
  "settings.keychain.masterPassword": "主密码：",
  "settings.keychain.masterPasswordPlaceholder": "输入密钥链主密码",
  "settings.keychain.unlock": "解锁",
  "settings.keychain.totpCode": "TOTP 代码：",
  "settings.keychain.verify": "验证",
  "settings.keychain.username": "用户名",
  "settings.keychain.copyUsername": "复制用户名",
  "settings.keychain.secret": "密钥",
  "settings.keychain.copySecret": "复制密钥",
  "settings.tasks.internalProtected": "内部/受保护",
  "settings.tasks.noRunLogs": "尚未记录运行日志。",
  "settings.tasks.noSummary": "无摘要",
  "settings.tasks.selectPrompt": "选择一个任务以查看计划、状态和运行历史。",
  "settings.tasks.pause": "暂停",
  "settings.tasks.resume": "恢复",
  "settings.tasks.delete": "删除",
  "settings.tasks.status": "状态",
  "settings.tasks.kind": "类型",
  "settings.tasks.schedule": "计划",
  "settings.tasks.nextRun": "下次运行",
  "settings.tasks.lastRun": "上次运行",
  "settings.tasks.lastResult": "上次结果",
  "settings.tasks.chat": "聊天",
  "settings.tasks.model": "模型",
  "settings.tasks.cwd": "工作目录",
  "settings.tasks.timeout": "超时",
  "settings.tasks.protection": "保护",
  "settings.tasks.protectionHint": "内部任务操作需要明确确认。",
  "settings.tasks.command": "命令",
  "settings.tasks.prompt": "提示",
  "settings.tasks.recentRuns": "最近运行",
  "settings.tasks.activeLabel": "活动",
  "settings.tasks.pausedLabel": "已暂停",
  "settings.tasks.completedLabel": "已完成",
  "settings.tasks.allStatuses": "所有状态",
  "settings.tasks.filterChatPlaceholder": "筛选聊天 JID…",
  "settings.tasks.refresh": "刷新",
  "settings.tasks.loading": "正在加载计划任务…",
  "settings.tasks.noneFound": "未找到计划任务。",
  "settings.tasks.noneFoundHint": "通过提醒、`/tasks` 或调度工具创建的任务将显示在此处。",
  "settings.tasks.listLabel": "计划任务",
  "settings.tasks.next": "下次",
  "settings.tasks.last": "上次",
  "settings.tasks.noMatch": "没有任务匹配 “{filter}”。",
  "settings.tasks.confirmDelete": "删除计划任务 {id}？",
  "settings.tasks.confirmPause": "暂停计划任务 {id}？",
  "settings.tasks.confirmResume": "恢复计划任务 {id}？",
  "settings.tasks.confirmProtected": "任务 {id} 是内部/受保护的。继续执行 {action}？",
  "settings.tasks.deleting": "正在删除 {id}…",
  "settings.tasks.pausing": "正在暂停 {id}…",
  "settings.tasks.resuming": "正在恢复 {id}…",
  "settings.tasks.deletedToast": "计划任务 {id} 已删除。",
  "settings.tasks.pausedToast": "计划任务 {id} 已暂停。",
  "settings.tasks.resumedToast": "计划任务 {id} 已恢复。",
  "settings.tasks.actionFailed": "执行 {action} 任务失败。",
  "settings.tasks.loadFailed": "加载计划任务失败。",
  "settings.compaction.appliedNotice": "压缩设置已应用。现有回合保留其当前计时器；新回合使用更新后的值。",
  "settings.compaction.saving": "正在保存压缩设置…",
  "settings.compaction.saveFailed": "保存压缩设置失败。",
  "settings.compaction.saved": "压缩设置已保存。",
  "settings.compaction.clearing": "正在清除 {chat} 的压缩抑制…",
  "settings.compaction.clearFailed": "清除压缩抑制失败。",
  "settings.compaction.cleared": "已清除 {chat} 的压缩抑制。",
  "settings.compaction.autoHeading": "自动压缩",
  "settings.compaction.enableAutomatic": "启用自动压缩",
  "settings.compaction.enableAutomaticHint": "由 Piclaw 管理的提示前/空闲压缩。上游代理自动压缩器会继续在内部保持禁用。",
  "settings.compaction.processingMethod": "处理方法",
  "settings.compaction.model": "压缩模型",
  "settings.compaction.modelHint": "用于本地智能压缩的严格模型。若已配置但不可用，压缩会停止并保留会话，不会回退。",
  "settings.compaction.modelPlaceholder": "提供商/模型（留空则使用当前模型）",
  "settings.compaction.methodSelective": "选择性",
  "settings.compaction.methodSelectiveHint": "提取高价值的连续性片段；当有界提示无法表示所有被丢弃的源事件时，使用完整的渐进式覆盖。",
  "settings.compaction.methodPipelined": "流水线",
  "settings.compaction.methodPipelinedHint": "在摘要前，对每个被丢弃的源事件进行规范化和分类，并生成可审计的覆盖账本。",
  "settings.compaction.remoteNative": "提供商原生压缩",
  "settings.compaction.remoteNativeHint": "仅对明确支持的提供商启用（{providers}）。任何失败都会自动回退到所选的本地方法。",
  "settings.compaction.remoteTimeout": "提供商原生超时（秒）",
  "settings.compaction.remoteTimeoutAria": "提供商原生压缩超时",
  "settings.compaction.remoteTimeoutHint": "远程预处理在回退到本地方法之前的截止时间。",
  "settings.compaction.enableToolResult": "启用工具结果压缩",
  "settings.compaction.enableToolResultHint": "禁用时，大型工具结果保持内联，不会外部化为可搜索的工具输出句柄。",
  "settings.compaction.semanticSummaries": "压缩工具结果的语义摘要",
  "settings.compaction.semanticSummariesHint": "启用时，压缩输出包含使用活动模型生成的语义摘要（失败时回退到预览）。",
  "settings.compaction.inputLimit": "语义摘要输入限制（字符）",
  "settings.compaction.inputLimitAria": "语义摘要输入限制",
  "settings.compaction.inputLimitHint": "用于语义摘要的完整工具输出采样的最大字符数。",
  "settings.compaction.maxTokens": "语义摘要输出最大令牌数",
  "settings.compaction.maxTokensAria": "语义摘要最大令牌数",
  "settings.compaction.maxTokensHint": "生成摘要长度的上限。",
  "settings.compaction.summaryTimeout": "语义摘要超时（秒）",
  "settings.compaction.summaryTimeoutAria": "语义摘要超时",
  "settings.compaction.summaryTimeoutHint": "在此超时后中止语义摘要生成并回退到预览压缩。",
  "settings.compaction.threshold": "压缩阈值（%）",
  "settings.compaction.thresholdAria": "压缩阈值",
  "settings.compaction.thresholdHint": "当上下文超过窗口的此百分比时自动压缩",
  "settings.compaction.timeout": "压缩超时（秒）",
  "settings.compaction.timeoutAria": "压缩超时",
  "settings.compaction.timeoutHint": "中止卡住的预提示/手动压缩，而不是永远挂起。",
  "settings.compaction.backoffBase": "失败退避基数（分钟）",
  "settings.compaction.backoffBaseAria": "压缩退避基数",
  "settings.compaction.backoffBaseHint": "压缩失败后的首个抑制窗口。",
  "settings.compaction.backoffMax": "失败退避最大值（分钟）",
  "settings.compaction.backoffMaxAria": "压缩退避最大值",
  "settings.compaction.backoffMaxHint": "重复失败后指数抑制的上限。",
  "settings.compaction.decayFactor": "退避衰减系数",
  "settings.compaction.decayFactorAria": "退避衰减系数",
  "settings.compaction.decayFactorHint": "% — 每次成功压缩后退避减半",
  "settings.compaction.watchdogHeading": "停滞监视器",
  "settings.compaction.enableWatchdog": "启用监视器",
  "settings.compaction.enableWatchdogHint": "默认禁用。启用时，如果活动阶段停止心跳，辅助进程将终止运行时。",
  "settings.compaction.watchdogTimeout": "监视器超时（秒）",
  "settings.compaction.watchdogTimeoutAria": "监视器超时",
  "settings.compaction.watchdogTimeoutHint": "活动阶段在监视器终止运行时之前可以无心跳持续多长时间。",
  "settings.compaction.suppressionsHeading": "活动压缩抑制",
  "settings.compaction.noBackoff": "当前没有聊天处于压缩退避状态。",
  "settings.compaction.clear": "清除",
  "settings.compaction.phasesHeading": "实时监视器阶段",
  "settings.compaction.noPhases": "目前没有活动的跟踪阶段。",
  "menu.title": "菜单",
  "menu.showWorkspace": "显示工作区",
  "menu.hideWorkspace": "隐藏工作区",
  "menu.openExplorer": "打开资源管理器",
  "menu.chatOnly": "仅聊天模式",
  "menu.exitChatOnly": "退出仅聊天模式",
  "menu.openTerminal": "在标签页中打开终端",
  "menu.openVnc": "在标签页中打开 VNC",
  "menu.newFile": "新建文件",
  "menu.openRecent": "打开最近文件",
  "menu.refreshTree": "刷新目录树",
  "menu.reindex": "重建工作区索引",
  "menu.showHidden": "显示隐藏文件",
  "menu.hideHidden": "隐藏隐藏文件",
  "menu.scale": "缩放",
  "menu.settings": "设置"
};
var JA = {
  "compose.placeholder": "メッセージ（Enterで送信、Shift+Enterで改行）...",
  "compose.send": "送信",
  "compose.stop": "停止",
  "compose.searchPlaceholder": "検索（Enterで実行）...",
  "compose.clearAll": "すべてクリア",
  "compose.clearAllTitle": "すべての添付と参照をクリア",
  "compose.scope": "範囲",
  "compose.searchScope": "検索範囲",
  "compose.scopeCurrent": "現在",
  "compose.scopeBranchFamily": "ブランチファミリー",
  "compose.scopeAll": "すべてのチャット",
  "compose.filterImages": "画像",
  "compose.filterAttachments": "添付",
  "compose.search": "検索",
  "compose.closeSearch": "検索を閉じる",
  "compose.shareLocation": "位置を共有",
  "compose.attachFile": "ファイルを添付",
  "compose.queueControls": "キュー済みフォローアップの操作",
  "compose.moveUp": "上に移動",
  "compose.moveUpQueue": "キュー内で上に移動",
  "compose.moveDown": "下に移動",
  "compose.moveDownQueue": "キュー内で下に移動",
  "compose.editInCompose": "入力欄で編集",
  "compose.returnToEditor": "キュー済みメッセージを入力欄に戻す",
  "compose.injectSteer": "キュー済みフォローアップをステアとして挿入",
  "compose.steer": "ステア",
  "compose.cancelQueued": "キュー済みメッセージをキャンセル",
  "compose.resizeInput": "メッセージ入力欄のサイズ変更",
  "compose.resizeInputHint": "ドラッグしてメッセージ入力欄のサイズを変更",
  "compose.modelPicker": "モデルピッカー",
  "compose.sessionsAndAgents": "セッションとエージェント",
  "compose.openModelPicker": "モデルピッカーを開く",
  "compose.newBranchTitle": "このチャットから新しいブランチを作成",
  "compose.newRootTitle": "web:ops のようなクリーンなルートセッションを作成",
  "compose.renameSessionTitle": "現在のセッションの名前を変更",
  "compose.pruneSessionTitle": "現在のエージェント/セッションブランチを削除（プルーン）",
  "compose.filterImagesTitle": "画像付きメッセージのみ表示",
  "compose.filterAttachmentsTitle": "添付付きメッセージのみ表示",
  "compose.selectModel": "モデルを選択",
  "compose.loadingModels": "モデルを読み込み中…",
  "compose.noModels": "利用可能なモデルがありません。",
  "compose.nextModel": "次のモデル",
  "compose.manageSessions": "セッションとエージェントを管理",
  "compose.noSessions": "他のセッションはまだありません。",
  "compose.newBranch": "新しいブランチ",
  "compose.newRoot": "新しいルート…",
  "compose.mergeCurrent": "現在を親にマージ",
  "compose.renameCurrent": "現在の名前を変更…",
  "compose.deleteCurrent": "現在を削除…",
  "compose.mergeInto": "このブランチを {target} にマージ",
  "compose.mergeBlocked": "このブランチはアクティブな間または子がある間はマージできません",
  "workspace.title": "ワークスペース",
  "workspace.moveConfirm": "{entry}「{name}」を{source}から{target}へ移動しますか？",
  "workspace.root": "ワークスペースのルート",
  "workspace.file": "ファイル",
  "workspace.folder": "フォルダー",
  "workspace.newFile": "新規ファイル",
  "workspace.refresh": "更新",
  "workspace.actions": "ワークスペース操作",
  "workspace.uploadFiles": "ファイルをアップロード",
  "workspace.reindexing": "ワークスペースを再インデックス中…",
  "workspace.deleteFile": "ファイルを削除",
  "workspace.download": "ダウンロード",
  "workspace.uploadToFolder": "このフォルダにファイルをアップロード",
  "workspace.addFolderHint": "フォルダのヒントを入力欄に追加",
  "workspace.downloadZip": "フォルダをzipでダウンロード",
  "workspace.openInTab": "タブで開く",
  "workspace.openInEditor": "エディタで開く",
  "workspace.renameSelected": "選択項目の名前を変更",
  "workspace.downloadSelectedFile": "選択したファイルをダウンロード",
  "workspace.downloadSelectedFolder": "選択したフォルダをダウンロード（zip）",
  "workspace.deleteSelectedFile": "選択したファイルを削除",
  "shell.settings": "設定",
  "shell.newChat": "新規チャット",
  "shell.connecting": "接続中…",
  "shell.connected": "接続済み",
  "language.label": "言語",
  "settings.title": "設定",
  "settings.close": "閉じる（Esc）",
  "settings.filter": "フィルター…",
  "settings.loading": "設定を読み込み中…",
  "settings.section.general": "一般",
  "settings.section.sessions": "セッション",
  "settings.section.recordings": "録画",
  "settings.section.compaction": "圧縮",
  "settings.section.budget": "予算",
  "settings.section.keyboard": "キーボード",
  "settings.section.workspace": "ワークスペース",
  "settings.section.environment": "環境",
  "settings.section.providers": "プロバイダー",
  "settings.section.models": "モデル",
  "settings.section.theme": "外観",
  "settings.section.scheduled-tasks": "スケジュールタスク",
  "settings.section.quick-actions": "クイックアクション",
  "settings.section.keychain": "キーチェーン",
  "settings.section.tools": "ツール",
  "settings.section.addons": "アドオン",
  "settings.placeholder.recordings": "録画をフィルター…",
  "settings.placeholder.keyboard": "ショートカットをフィルター…",
  "settings.placeholder.environment": "環境をフィルター…",
  "settings.placeholder.models": "モデルをフィルター…",
  "settings.placeholder.scheduled-tasks": "スケジュールタスクをフィルター…",
  "settings.placeholder.quick-actions": "クイックアクションをフィルター…",
  "settings.placeholder.keychain": "エントリをフィルター…",
  "settings.placeholder.tools": "ツールをフィルター…",
  "settings.placeholder.addons": "アドオンをフィルター…",
  "preview.close": "閉じる",
  "preview.loading": "プレビューを読み込み中…",
  "preview.files": "ファイル",
  "preview.folders": "フォルダ",
  "preview.compressed": "圧縮後",
  "preview.uncompressed": "非圧縮",
  "preview.name": "名前",
  "preview.type": "種類",
  "preview.method": "方式",
  "preview.size": "サイズ",
  "post.deleteMessage": "メッセージを削除",
  "post.tooLarge": "メッセージが大きすぎて表示できません。",
  "post.previewTruncated": "プレビューは切り詰められました。",
  "post.submitted": "送信済み",
  "post.discard": "破棄",
  "post.save": "保存",
  "post.cancel": "キャンセル",
  "post.addNote": "メモを追加",
  "post.addNotePlaceholder": "メモを追加…",
  "post.restartNotice": "再起動中 — 理由：{reason}",
  "post.restartCompleted": "再起動が完了しました。",
  "post.agentSelfResume": "エージェントの自己再開",
  "tab.close": "閉じる",
  "tab.closeOthers": "他を閉じる",
  "tab.closeAll": "すべて閉じる",
  "tab.reattach": "再アタッチ",
  "tab.openInWindow": "ウィンドウで開く",
  "tab.openInNewTab": "新しいタブで開く",
  "tab.pinned": "ピン留め済み",
  "tab.detached": "分離済み",
  "tab.openSeparateWindow": "別ウィンドウで開く",
  "status.trackedVariables": "追跡中の変数",
  "status.attachToSession": "セッションにアタッチ",
  "status.files": "ファイル",
  "status.proposedDiff": "提案された差分",
  "status.copyTmux": "tmuxコマンドをコピー",
  "status.experimentDuration": "実験の経過時間",
  "status.sinceLastActivity": "最後のアクティビティから",
  "annotator.title": "画像に注釈",
  "annotator.typeLabel": "ラベルを入力…",
  "annotator.undo": "元に戻す",
  "annotator.resetZoom": "ズームをリセット",
  "tree.filter": "フィルター…",
  "tree.sessionTree": "セッションツリー",
  "btw.label": "BTW サイド会話",
  "btw.close": "BTW を閉じる",
  "btw.thinking": "思考中",
  "mdpreview.close": "プレビューを閉じる",
  "mdpreview.unavailable": "プレビューを利用できません",
  "widget.close": "ウィジェットを閉じる",
  "oobe.gettingStarted": "はじめに",
  "oobe.needsSetupTitle": "インスタンスのセットアップが必要",
  "oobe.configuredTitle": "インスタンスは設定済み",
  "oobe.needsSetupBody": "このインスタンスはまだ設定されていません。設定を開き、AIプロバイダー/モデルを設定してリクエストの送信を開始してください。",
  "oobe.configuredBody": "このインスタンスは設定済みのようです。設定でプロバイダーとモデルの設定を確認または更新してください。",
  "oobe.openSettings": "設定を開く",
  "oobe.dismiss": "閉じる",
  "oobe.done": "完了",
  "palette.placeholder": "入力してエージェント、ワークスペース操作、またはスラッシュコマンドにジャンプ…",
  "palette.hideWorkspace": "ワークスペースを非表示",
  "palette.showWorkspace": "ワークスペースを表示",
  "palette.hideWorkspaceDesc": "ワークスペースサイドバーを非表示にします。",
  "palette.showWorkspaceDesc": "ワークスペースサイドバーを表示します。",
  "palette.exitChatOnly": "チャットのみモードを終了",
  "palette.chatOnly": "チャットのみモード",
  "palette.exitChatOnlyDesc": "分割ワークスペースレイアウトに戻ります。",
  "palette.chatOnlyDesc": "チャットのみのレイアウトに切り替えます。",
  "palette.groupAgents": "エージェント",
  "palette.groupWorkspace": "ワークスペース",
  "palette.groupSlash": "スラッシュコマンド",
  "palette.hintMove": "移動",
  "palette.hintSelect": "選択",
  "palette.hintPopOut": "ポップアウト",
  "palette.hintClose": "閉じる",
  "settings.appliedNotice": "設定を適用しました。変更は次のターンから有効になります。",
  "settings.sessions.lifecycle": "セッションのライフサイクル",
  "settings.sessions.autoRotate": "セッションを自動ローテーション",
  "settings.sessions.maxSize": "最大セッションサイズ（MB）",
  "settings.sessions.maxSizeAria": "最大セッションサイズ",
  "settings.sessions.agentBehaviour": "エージェントの動作",
  "settings.sessions.toolBudget": "ツール使用予算",
  "settings.sessions.toolBudgetAria": "ツール使用予算",
  "settings.sessions.toolBudgetHint": "1ターンあたりの完了済みツール実行回数の上限",
  "settings.sessions.isolation": "セッションの分離",
  "settings.sessions.isolationNone": "なし — セッション間で完全に可視",
  "settings.sessions.isolationSummary": "概要 — ツールは可視、引数は非表示",
  "settings.sessions.isolationFull": "完全 — セッション同士は互いに見えない",
  "settings.editor.heading": "エディター",
  "settings.editor.vimMode": "Vim モード",
  "settings.editor.showWhitespace": "空白文字を表示",
  "settings.editor.livePreview": "Markdown ライブプレビュー",
  "settings.editor.fontSize": "フォントサイズ（px）",
  "settings.editor.fontSizeAria": "エディターのフォントサイズ",
  "settings.editor.fontFamily": "フォントファミリー",
  "settings.editor.fontFamilyPlaceholder": "monospace（デフォルト）",
  "settings.editor.localOnlyHint": "このブラウザーのみ。エディターの変更はローカルブラウザーストレージに保存され、次にファイルタブを開くか再読み込みしたときに有効になります。",
  "settings.appearance.syncing": "外観を同期中…",
  "settings.appearance.default": "デフォルト",
  "settings.appearance.autoLightDark": "自動（ライト/ダーク）",
  "settings.appearance.tint": "色調：",
  "settings.appearance.clearTint": "色調をクリア",
  "settings.appearance.none": "なし",
  "settings.appearance.outputPadding": "出力の余白",
  "settings.appearance.outputPaddingHint": "メッセージと思考パネルの周囲に追加する余白です。",
  "settings.keyboard.heading": "キーボード",
  "settings.keyboard.hint1": "アプリ全体のショートカットをカンマ区切りのバインディングとしてカスタマイズします。変更はすぐに反映されます。",
  "settings.keyboard.hint1b": "は閉じる/中止用に予約されており、再割り当てできません。",
  "settings.keyboard.hint2mid": "と入力",
  "settings.keyboard.hint2end": "を入力欄の外で押すとこのペインが開きます。",
  "settings.keyboard.resetAll": "すべてデフォルトにリセット",
  "settings.keyboard.defaultColon": "デフォルト：",
  "settings.keyboard.save": "保存",
  "settings.keyboard.defaultBtn": "デフォルト",
  "settings.keyboard.noMatch": "このフィルターに一致するショートカットはありません。",
  "settings.keyboard.invalidShortcut": "無効なショートカット：{token}。Escape は予約されており、再割り当てできません。",
  "settings.keyboard.saved": "キーボードショートカットを保存しました。",
  "settings.keyboard.resetOne": "キーボードショートカットをデフォルトにリセットしました。",
  "settings.keyboard.resetAllDone": "キーボードショートカットをすべてデフォルトにリセットしました。",
  "settings.workspace.serverApplied": "ワークスペース設定を適用しました。サーバー側の制限は新しいワークスペースリクエストに直ちに反映されます。",
  "settings.workspace.browserApplied": "ブラウザーのワークスペース設定はこのタブで直ちに適用されました。",
  "settings.workspace.access": "アクセス",
  "settings.workspace.enableTerminal": "Web ターミナルを有効化",
  "settings.workspace.allowVnc": "直接 VNC ターゲットを許可",
  "settings.workspace.accessHint": "ターミナルアクセスは直ちに更新されます。直接 VNC ターゲットポリシーは新しい VNC リクエストに適用されます。",
  "settings.workspace.guardrails": "サーバースキャンのガードレール",
  "settings.workspace.maxDepth": "最大ツリー深度",
  "settings.workspace.maxDepthAria": "ワークスペースツリーの最大深度",
  "settings.workspace.maxDepthHintPre": "すべての",
  "settings.workspace.maxDepthHintPost": "リクエストを制限します",
  "settings.workspace.maxEntries": "スキャンあたりの最大エントリ数",
  "settings.workspace.maxEntriesAria": "ワークスペースツリーの最大エントリ数",
  "settings.workspace.maxEntriesHint": "大きすぎるツリー走査を早めに打ち切ります",
  "settings.workspace.thisBrowser": "このブラウザー",
  "settings.workspace.refreshInterval": "更新間隔（秒）",
  "settings.workspace.refreshIntervalAria": "ワークスペース更新間隔",
  "settings.workspace.folderDepth": "フォルダプレビューのスキャン深度",
  "settings.workspace.folderDepthAria": "フォルダプレビューのスキャン深度",
  "settings.workspace.folderDepthHintPre": "",
  "settings.workspace.folderDepthHintPost": "に設定するとフォルダサイズのプレビュースキャンを無効化します",
  "settings.workspace.footerHint": "ルートおよびフォルダ展開のツリー読み込みは浅いままです。フォルダサイズのプレビューは UI で最も深いワークスペーススキャンです。",
  "settings.models.thinkingLevel": "思考レベル",
  "settings.models.noThinking": "現在のモデルは思考をサポートしていません。",
  "settings.models.thinkingLevelLabel": "思考レベル：",
  "settings.models.loading": "モデルを読み込み中…",
  "settings.models.summary": "狭いペインでは、クリッピングを避けるためにモデル名とプロバイダー名が折り返される場合があります。",
  "settings.models.scopedOnly": "スコープ付きモデルのみ",
  "settings.models.scopedCheckboxPre": "Piclaw のモデル一覧に Pi の",
  "settings.models.scopedCheckboxPost": "を使用",
  "settings.models.scopedHintPre": "このピッカーと",
  "settings.models.scopedHintPost": "ツールをフィルタリングします。TUI のモデル選択は変更されません。",
  "settings.models.colModel": "モデル",
  "settings.models.colProvider": "プロバイダー",
  "settings.models.colContext": "コンテキスト",
  "settings.models.colReasoning": "推論",
  "settings.models.noMatch": "「{filter}」に一致するモデルはありません",
  "settings.tools.unavailable": "ツールデータを利用できません。",
  "settings.tools.search": "検索",
  "settings.tools.matchMode": "マッチモード",
  "settings.tools.orMode": "いずれかのキーワード（OR）— 少なくとも1つの検索語に一致",
  "settings.tools.andMode": "すべてのキーワード（AND）— すべての検索語に一致",
  "settings.tools.colEnabled": "有効",
  "settings.tools.colTool": "ツール",
  "settings.tools.colCompact": "結果圧縮",
  "settings.tools.colKind": "種類",
  "settings.tools.colSummary": "概要",
  "settings.tools.colSource": "ソース",
  "settings.tools.disableCompaction": "このツールのツール結果コンパクションを無効化",
  "settings.tools.enableCompaction": "このツールのツール結果コンパクションを有効化",
  "settings.tools.noMatch": "「{filter}」に一致するツールはありません",
  "settings.tools.footer": "ツールのアクティベーションはエージェントランタイムが管理します。グループのチェックボックスで折りたたみ/展開でき、「コンパクト」列はツール結果コンパクションの対象可否を制御します。",
  "settings.environment.heading": "環境",
  "settings.environment.introPre": "キーチェーン以外の環境変数のみを表示しています。オーバーライドは拡張機能の KV に保存され、",
  "settings.environment.introPost": "に適用されるため、以降のツール呼び出しに継承されます。",
  "settings.environment.refresh": "更新",
  "settings.environment.addOverride": "オーバーライドを追加",
  "settings.environment.valuePlaceholder": "値",
  "settings.environment.save": "保存",
  "settings.environment.countLine": "{count} 個の変数を表示 • {overrides} 個のオーバーライドが有効 • {keychain} 個のキーチェーン注入変数を非表示",
  "settings.environment.overridden": "KV でオーバーライド",
  "settings.environment.inherited": "プロセス環境から継承",
  "settings.environment.kindOverride": "オーバーライド",
  "settings.environment.kindProcess": "プロセス",
  "settings.environment.clear": "クリア",
  "settings.environment.noMatch": "「{filter}」に一致する環境変数はありません。",
  "settings.environment.refreshedToast": "環境を更新しました。",
  "settings.environment.savedToast": "{name} の環境オーバーライドを保存しました。",
  "settings.environment.clearedToast": "{name} の環境オーバーライドをクリアしました。",
  "settings.quickActions.loading": "読み込み中…",
  "settings.quickActions.heading": "タイムラインクイックアクション",
  "settings.quickActions.intro": "タイムラインのタイプアヘッドに表示するアクションを選択します。エージェントは常に最初に固定され、次にワークスペースコマンド、その次にスラッシュコマンドが表示されます。",
  "settings.quickActions.enableAll": "すべて有効化",
  "settings.quickActions.saving": "保存中…",
  "settings.quickActions.saveApply": "保存して適用",
  "settings.quickActions.workspaceCommands": "ワークスペースコマンド",
  "settings.quickActions.noWorkspaceMatch": "このフィルターに一致するワークスペースコマンドはありません。",
  "settings.quickActions.slashCommands": "スラッシュコマンド",
  "settings.quickActions.slashFallback": "スラッシュコマンド",
  "settings.quickActions.noSlashMatch": "このフィルターに一致するスラッシュコマンドはありません。",
  "settings.quickActions.savingToast": "クイックアクションを保存中…",
  "settings.quickActions.savedToast": "クイックアクションを保存しました。",
  "settings.providers.authApiKey": "API キー",
  "settings.providers.authConfigured": "設定済み",
  "settings.providers.heading": "プロバイダー",
  "settings.providers.tagCustom": "カスタム",
  "settings.providers.logout": "ログアウト",
  "settings.providers.reconfigure": "再設定",
  "settings.providers.setUp": "セットアップ",
  "settings.providers.setupHint": "サインインフローはブラウザーで開きます。狭いペインではセットアップフォームが縦に積み重なってクリッピングを防ぎます。",
  "settings.providers.starting": "開始中…",
  "settings.providers.signInOAuth": "OAuth でサインイン",
  "settings.providers.apiKeyLabel": "API キー",
  "settings.providers.apiKeyPlaceholder": "API キーを入力",
  "settings.providers.save": "保存",
  "settings.providers.configuring": "設定中…",
  "settings.providers.saveConfig": "設定を保存",
  "settings.providers.apiKeyEmpty": "API キーを空にすることはできません。",
  "settings.providers.configuringToast": "{provider} を設定中…",
  "settings.providers.configured": "{provider} を設定しました。",
  "settings.providers.startingOAuth": "{provider} の OAuth を開始中…",
  "settings.providers.oauthOpened": "OAuth ウィンドウを開きました。サインインフローを完了してから、このメッセージを閉じてください。",
  "settings.providers.oauthStarted": "{provider} の OAuth フローを開始しました。チャットを確認してください。",
  "settings.providers.loggingOut": "{provider} をログアウト中…",
  "settings.providers.loggedOut": "{provider} をログアウトしました。再起動が必要な場合があります。",
  "settings.general.identity": "アイデンティティ",
  "settings.general.userLabel": "ユーザー",
  "settings.general.yourName": "あなたの名前",
  "settings.general.agentLabel": "エージェント",
  "settings.general.agentName": "エージェント名",
  "settings.general.notifications": "通知",
  "settings.general.browserNotifications": "ブラウザ通知",
  "settings.general.notifSecureHint": "入力バーの \uD83D\uDD14 ベルボタンで通知を有効/無効にします。Web Push には HTTPS または localhost が必要です。",
  "settings.general.notifInsecureHint": "⚠ 利用不可 — セキュアコンテキスト（HTTPS または localhost）が必要です。SSH トンネルまたは TLS 付きリバースプロキシ経由でアクセスして有効化してください。",
  "settings.general.display": "表示",
  "settings.general.systemMeters": "システムメーター",
  "settings.general.systemMetersHint": "ステータスバーの CPU/メモリ/ネットワークメーター。このブラウザのみ。",
  "settings.general.instanceConfig": "インスタンス設定",
  "settings.general.composeUpload": "作成アップロード（MB）",
  "settings.general.composeUploadAria": "作成アップロード上限",
  "settings.general.composeUploadHint": "チャット/メディア添付",
  "settings.general.workspaceUpload": "ワークスペースアップロード（MB）",
  "settings.general.workspaceUploadAria": "ワークスペースアップロード上限",
  "settings.general.workspaceUploadHint": "デフォルトは 256 MB。チャンクアップロードは最大 1 GB まで許可",
  "settings.general.agentRecovery": "詳細 · エージェント復旧",
  "settings.general.automaticRecovery": "自動復旧",
  "settings.general.automaticRecoveryHint": "復旧可能な失敗ターンを自動的に再試行します。",
  "settings.general.recoveryMaxAttempts": "最大試行回数",
  "settings.general.recoveryMaxAttemptsAria": "自動復旧の最大試行回数",
  "settings.general.recoveryMaxAttemptsHint": "0 は通常の再試行上限を継承します。",
  "settings.general.recoveryTotalBudget": "合計予算（ミリ秒）",
  "settings.general.recoveryTotalBudgetAria": "自動復旧の合計予算（ミリ秒）",
  "settings.general.recoveryTotalBudgetHint": "0 はターンのタイムアウトの 3 分の 1（6〜60 分に制限）を使用し、正の値は明示的な上限です。",
  "settings.general.recoveryEffectiveBudget": "既定のターンタイムアウトでの有効予算：{budget} ミリ秒。",
  "settings.general.authentication": "認証",
  "settings.general.widgetToken": "ウィジェット bearer トークン",
  "settings.general.token": "トークン",
  "settings.general.hideToken": "トークンを隠す",
  "settings.general.revealToken": "トークンを表示",
  "settings.general.copyToken": "トークンをコピー",
  "settings.general.copied": "コピーしました",
  "settings.general.regenerating": "再生成中…",
  "settings.general.regenerate": "再生成",
  "settings.general.tokenHintPre": "次の読み取り専用トークン：",
  "settings.general.tokenHintMid": "および",
  "settings.general.tokenHintPost": "。次として使用：",
  "settings.general.tokenHintEnd": "。",
  "settings.general.copyFailed": "ウィジェットトークンをコピーできませんでした。トークンフィールドを選択して手動でコピーしてください。",
  "settings.general.regenConfirm": "ウィジェットトークンを再生成しますか？古いトークンを使用している既存の macOS ウィジェットは更新されなくなります。",
  "settings.general.totpTitle": "TOTP セットアップ QR",
  "settings.general.totpConfiguredHint": "現在の Web ログイン認証システムのシークレット。この QR をスキャンして別の認証デバイスを追加します。",
  "settings.general.totpUnconfiguredHint": "このインスタンスにはまだ TOTP が設定されていないため、セットアップ QR は利用できません。",
  "settings.general.issuer": "発行者",
  "settings.general.label": "ラベル",
  "settings.general.secret": "シークレット",
  "settings.general.avatarUpload": "クリックしてアップロード",
  "settings.developer.heading": "開発者",
  "settings.developer.devMode": "開発者モード",
  "settings.developer.localHint": "このブラウザのみ。開発者モードの切り替えとアドオンカタログのオーバーライドはローカルブラウザストレージに保存されます。",
  "settings.developer.addonSources": "アドオンソース",
  "settings.developer.catalogUrl": "カタログ URL",
  "settings.developer.catalogHint": "プライマリアドオンカタログ URL。空のままにするとデフォルトを使用します",
  "settings.developer.additionalCatalogs": "追加カタログ URL",
  "settings.developer.additionalHint": "プライマリ/デフォルトカタログに加えて取得されます。1 行に 1 つの URL。",
  "settings.developer.repoUrl": "リポジトリ URL",
  "settings.developer.repoHintPre": "git リポジトリを上書き（",
  "settings.developer.repoHintPost": "インストール用）。空のままでデフォルト。",
  "settings.developer.debug": "デバッグ",
  "settings.developer.logSse": "SSE イベントをログ記録",
  "settings.developer.logToolCalls": "ツール呼び出しをログ記録",
  "settings.developer.debugHint": "デバッグフラグは次回のページ再読み込み時に有効になります。",
  "settings.addons.installing": "{slug} をインストール中…",
  "settings.addons.removing": "{slug} を削除中…",
  "settings.addons.installedToast": "アドオンをインストールしました。",
  "settings.addons.removedToast": "アドオンを削除しました。",
  "settings.addons.restarting": "piclaw を再起動中…",
  "settings.addons.restartComplete": "再起動完了 — アドオンを更新しました。",
  "settings.addons.restartTimeout": "バックエンドが時間内に応答しませんでした。ページを手動で再読み込みしてください。",
  "settings.addons.fetching": "アドオンを取得中…",
  "settings.addons.loadFailed": "アドオンを読み込めませんでした。",
  "settings.addons.catalogFromPre": "カタログの取得元：",
  "settings.addons.catalogMerged": "{count} 個のカタログソースをマージしました。",
  "settings.addons.installNote": "Bun によるパッケージ優先インストール。インストール/アンインストール後に再起動が必要です。",
  "settings.addons.failedFetchSingular": "{count} 個のカタログソースの取得に失敗しました：",
  "settings.addons.failedFetchPlural": "{count} 個のカタログソースの取得に失敗しました：",
  "settings.addons.activeSources": "アクティブなカタログソース（{count}）",
  "settings.addons.windowsWarning": "ネイティブ Windows のアドオンインストールはリスクが高くなります：Bun パッケージのインストール、シンボリックリンクのクリーンアップ、ロックされたファイル、再起動のタイミングは、Linux/WSL よりも予測しにくい場合があります。可能であれば WSL またはコンテナを優先してください。",
  "settings.addons.typeExtSkill": "拡張機能 + スキル",
  "settings.addons.typeSkill": "スキル",
  "settings.addons.typeExt": "拡張機能",
  "settings.addons.update": "更新",
  "settings.addons.remove": "削除",
  "settings.addons.install": "インストール",
  "settings.addons.noMatch": "「{filter}」に一致するアドオンはありません",
  "settings.addons.restartNotice": "拡張機能の変更はインストールされましたが、piclaw が再起動するまで非アクティブです。",
  "settings.addons.restartNow": "今すぐ再起動",
  "settings.recordings.modeFull": "完全 / 信頼済み",
  "settings.recordings.modeMetadata": "メタデータのみ",
  "settings.recordings.modeRedacted": "編集済み",
  "settings.recordings.selectPrompt": "録画を選択して検査、再生、エクスポート、または削除します。",
  "settings.recordings.playback": "再生",
  "settings.recordings.refresh": "更新",
  "settings.recordings.delete": "削除",
  "settings.recordings.status": "ステータス",
  "settings.recordings.mode": "モード",
  "settings.recordings.chat": "チャット",
  "settings.recordings.started": "開始",
  "settings.recordings.ended": "終了",
  "settings.recordings.events": "イベント",
  "settings.recordings.redactions": "編集",
  "settings.recordings.exportJson": "JSON をエクスポート",
  "settings.recordings.exportJsonl": "JSONL をエクスポート",
  "settings.recordings.exportHtml": "スタンドアロン HTML をエクスポート",
  "settings.recordings.eventSummary": "イベント概要",
  "settings.recordings.inspectHint": "詳細を開くか更新してトレースイベントを検査します。",
  "settings.recordings.firstEvents": "最初のイベント",
  "settings.recordings.heading": "セッション録画",
  "settings.recordings.intro": "決定論的な再生と画面録画エクスポートのためのオプトイントレースキャプチャ。再生でライブエージェントやツールのエンドポイントを呼び出すことはありません。",
  "settings.recordings.chatJid": "チャット JID",
  "settings.recordings.title": "タイトル",
  "settings.recordings.titlePlaceholder": "デモ録画",
  "settings.recordings.modeLabelField": "モード",
  "settings.recordings.optRedacted": "編集済み",
  "settings.recordings.optMetadata": "メタデータのみ",
  "settings.recordings.optFull": "完全 / 信頼済みローカル",
  "settings.recordings.includeSnapshot": "タイムラインスナップショットを含める",
  "settings.recordings.extraKeys": "追加の編集キー",
  "settings.recordings.extraPatterns": "追加の正規表現パターン",
  "settings.recordings.stopCurrent": "現在のチャット録画を停止",
  "settings.recordings.start": "録画を開始",
  "settings.recordings.redactionPreview": "編集プレビュー",
  "settings.recordings.previewRedaction": "編集をプレビュー",
  "settings.recordings.loading": "録画を読み込み中…",
  "settings.recordings.noneYet": "まだ録画がありません。",
  "settings.recordings.noneYetHint": "上で録画を開始し、再生/エクスポートを使用して決定論的な画面キャプチャを行います。",
  "settings.recordings.listLabel": "セッション録画",
  "settings.recordings.eventsCount": "{count} 件のイベント",
  "settings.recordings.noMatch": "「{filter}」に一致する録画はありません。",
  "settings.recordings.startedToast": "{chat} の録画を開始しました。",
  "settings.recordings.startFailed": "録画の開始に失敗しました。",
  "settings.recordings.stoppedToast": "{chat} の録画を停止しました。",
  "settings.recordings.stopFailed": "録画の停止に失敗しました。",
  "settings.recordings.deleteConfirm": "録画 {id} を削除しますか？",
  "settings.recordings.deletedToast": "録画を削除しました。",
  "settings.recordings.deleteFailed": "録画の削除に失敗しました。",
  "settings.recordings.loadOneFailed": "録画の読み込みに失敗しました。",
  "settings.recordings.loadFailed": "録画の読み込みに失敗しました。",
  "settings.recordings.previewFailed": "プレビューに失敗しました。",
  "settings.keychain.loadFailed": "キーチェーンの読み込みに失敗しました。",
  "settings.keychain.addFailed": "エントリの追加に失敗しました。",
  "settings.keychain.deleteFailed": "エントリの削除に失敗しました。",
  "settings.keychain.saveNotesFailed": "メモの保存に失敗しました。",
  "settings.keychain.revealFailed": "表示に失敗しました。",
  "settings.keychain.loading": "キーチェーンを読み込み中…",
  "settings.keychain.entryCountSingular": "{count} 件のエントリ",
  "settings.keychain.entryCountPlural": "{count} 件のエントリ",
  "settings.keychain.matchingFilter": " 「{filter}」に一致",
  "settings.keychain.encryptedSuffix": "、保存時に暗号化。",
  "settings.keychain.clickPrefix": "クリック",
  "settings.keychain.revealSuffix": "で表示。",
  "settings.keychain.cancel": "キャンセル",
  "settings.keychain.addEntry": "+ エントリを追加",
  "settings.keychain.namePlaceholder": "エントリ名（例：github/my-token）",
  "settings.keychain.secretPlaceholder": "シークレット値",
  "settings.keychain.usernamePlaceholder": "ユーザー名（任意）",
  "settings.keychain.saving": "保存中…",
  "settings.keychain.save": "保存",
  "settings.keychain.userNotePlaceholder": "ユーザーメモ（この UI でのみ表示）",
  "settings.keychain.agentNotePlaceholder": "エージェントメモ（エージェントに公開しても安全）",
  "settings.keychain.noMatchFilter": "フィルターに一致するエントリはありません。",
  "settings.keychain.noEntries": "キーチェーンエントリがありません。",
  "settings.keychain.hideSecret": "シークレットを非表示",
  "settings.keychain.revealSecret": "シークレットを表示",
  "settings.keychain.deleteQ": "削除しますか？",
  "settings.keychain.yes": "はい",
  "settings.keychain.no": "いいえ",
  "settings.keychain.deleteTitle": "削除",
  "settings.keychain.userNote": "ユーザーメモ",
  "settings.keychain.agentNote": "エージェント読み取り可能メモ",
  "settings.keychain.userNoteHint": "人間/UI メモのみ",
  "settings.keychain.agentNoteHint": "エージェント向けの安全なガイダンス",
  "settings.keychain.saveNotes": "メモを保存",
  "settings.keychain.masterPassword": "マスターパスワード：",
  "settings.keychain.masterPasswordPlaceholder": "キーチェーンのマスターパスワードを入力",
  "settings.keychain.unlock": "ロック解除",
  "settings.keychain.totpCode": "TOTP コード：",
  "settings.keychain.verify": "検証",
  "settings.keychain.username": "ユーザー名",
  "settings.keychain.copyUsername": "ユーザー名をコピー",
  "settings.keychain.secret": "シークレット",
  "settings.keychain.copySecret": "シークレットをコピー",
  "settings.tasks.internalProtected": "内部/保護済み",
  "settings.tasks.noRunLogs": "まだ実行ログが記録されていません。",
  "settings.tasks.noSummary": "概要なし",
  "settings.tasks.selectPrompt": "タスクを選択してスケジュール、ステータス、実行履歴を確認します。",
  "settings.tasks.pause": "一時停止",
  "settings.tasks.resume": "再開",
  "settings.tasks.delete": "削除",
  "settings.tasks.status": "ステータス",
  "settings.tasks.kind": "種類",
  "settings.tasks.schedule": "スケジュール",
  "settings.tasks.nextRun": "次回実行",
  "settings.tasks.lastRun": "前回実行",
  "settings.tasks.lastResult": "前回の結果",
  "settings.tasks.chat": "チャット",
  "settings.tasks.model": "モデル",
  "settings.tasks.cwd": "作業ディレクトリ",
  "settings.tasks.timeout": "タイムアウト",
  "settings.tasks.protection": "保護",
  "settings.tasks.protectionHint": "内部タスクの操作には明示的な確認が必要です。",
  "settings.tasks.command": "コマンド",
  "settings.tasks.prompt": "プロンプト",
  "settings.tasks.recentRuns": "最近の実行",
  "settings.tasks.activeLabel": "アクティブ",
  "settings.tasks.pausedLabel": "一時停止",
  "settings.tasks.completedLabel": "完了",
  "settings.tasks.allStatuses": "すべてのステータス",
  "settings.tasks.filterChatPlaceholder": "チャット JID をフィルター…",
  "settings.tasks.refresh": "更新",
  "settings.tasks.loading": "スケジュールタスクを読み込み中…",
  "settings.tasks.noneFound": "スケジュールされたタスクが見つかりません。",
  "settings.tasks.noneFoundHint": "リマインダー、`/tasks`、またはスケジューラツールで作成されたタスクがここに表示されます。",
  "settings.tasks.listLabel": "スケジュールされたタスク",
  "settings.tasks.next": "次回",
  "settings.tasks.last": "前回",
  "settings.tasks.noMatch": "「{filter}」に一致するタスクはありません。",
  "settings.tasks.confirmDelete": "スケジュールタスク {id} を削除しますか？",
  "settings.tasks.confirmPause": "スケジュールタスク {id} を一時停止しますか？",
  "settings.tasks.confirmResume": "スケジュールタスク {id} を再開しますか？",
  "settings.tasks.confirmProtected": "タスク {id} は内部/保護済みです。{action} を続行しますか？",
  "settings.tasks.deleting": "{id} を削除中…",
  "settings.tasks.pausing": "{id} を一時停止中…",
  "settings.tasks.resuming": "{id} を再開中…",
  "settings.tasks.deletedToast": "スケジュールタスク {id} を削除しました。",
  "settings.tasks.pausedToast": "スケジュールタスク {id} を一時停止しました。",
  "settings.tasks.resumedToast": "スケジュールタスク {id} を再開しました。",
  "settings.tasks.actionFailed": "{action} タスクに失敗しました。",
  "settings.tasks.loadFailed": "スケジュールタスクの読み込みに失敗しました。",
  "settings.compaction.appliedNotice": "圧縮設定が適用されました。既存のターンは現在のタイマーを保持し、新しいターンは更新された値を使用します。",
  "settings.compaction.saving": "圧縮設定を保存中…",
  "settings.compaction.saveFailed": "圧縮設定の保存に失敗しました。",
  "settings.compaction.saved": "圧縮設定を保存しました。",
  "settings.compaction.clearing": "{chat} の圧縮抑制をクリア中…",
  "settings.compaction.clearFailed": "圧縮抑制のクリアに失敗しました。",
  "settings.compaction.cleared": "{chat} の圧縮抑制をクリアしました。",
  "settings.compaction.autoHeading": "自動圧縮",
  "settings.compaction.enableAutomatic": "自動圧縮を有効化",
  "settings.compaction.enableAutomaticHint": "Piclaw が管理するプロンプト前/アイドル時の圧縮です。上流エージェントの自動圧縮は内部的に抑制されたままです。",
  "settings.compaction.processingMethod": "処理方式",
  "settings.compaction.model": "圧縮モデル",
  "settings.compaction.modelHint": "ローカルスマート圧縮専用の厳密なモデルです。設定済みで利用できない場合は、フォールバックせずセッションを保持して停止します。",
  "settings.compaction.modelPlaceholder": "provider/model（空欄は現在のモデル）",
  "settings.compaction.methodSelective": "選択型",
  "settings.compaction.methodSelectiveHint": "重要な継続情報を抽出し、制限付きプロンプトですべての破棄対象イベントを表現できない場合は完全な段階的カバレッジを使用します。",
  "settings.compaction.methodPipelined": "パイプライン",
  "settings.compaction.methodPipelinedHint": "要約前に、破棄対象の各ソースイベントを正規化・分類し、監査可能なカバレッジ台帳を作成します。",
  "settings.compaction.remoteNative": "プロバイダー・ネイティブ圧縮",
  "settings.compaction.remoteNativeHint": "明示的に対応しているプロバイダー（{providers}）のみオプトインできます。失敗時は選択したローカル方式へアトミックにフォールバックします。",
  "settings.compaction.remoteTimeout": "プロバイダー・ネイティブのタイムアウト（秒）",
  "settings.compaction.remoteTimeoutAria": "プロバイダー・ネイティブ圧縮のタイムアウト",
  "settings.compaction.remoteTimeoutHint": "ローカル方式へフォールバックする前のリモート処理期限です。",
  "settings.compaction.enableToolResult": "ツール結果の圧縮を有効化",
  "settings.compaction.enableToolResultHint": "無効にすると、大きなツール結果はインラインのまま残り、検索可能なツール出力ハンドルに外部化されません。",
  "settings.compaction.semanticSummaries": "圧縮されたツール結果のセマンティック要約",
  "settings.compaction.semanticSummariesHint": "有効にすると、圧縮された出力にアクティブモデルで生成されたセマンティック要約が含まれます（失敗時はプレビューにフォールバック）。",
  "settings.compaction.inputLimit": "セマンティック要約の入力上限（文字）",
  "settings.compaction.inputLimitAria": "セマンティック要約の入力上限",
  "settings.compaction.inputLimitHint": "セマンティック要約のために完全なツール出力からサンプリングする最大文字数。",
  "settings.compaction.maxTokens": "セマンティック要約の出力最大トークン数",
  "settings.compaction.maxTokensAria": "セマンティック要約の最大トークン数",
  "settings.compaction.maxTokensHint": "生成される要約の長さの上限。",
  "settings.compaction.summaryTimeout": "セマンティック要約のタイムアウト（秒）",
  "settings.compaction.summaryTimeoutAria": "セマンティック要約のタイムアウト",
  "settings.compaction.summaryTimeoutHint": "このタイムアウト後にセマンティック要約の生成を中止し、プレビュー圧縮にフォールバックします。",
  "settings.compaction.threshold": "圧縮しきい値（%）",
  "settings.compaction.thresholdAria": "圧縮しきい値",
  "settings.compaction.thresholdHint": "コンテキストがウィンドウのこの％を超えたら自動圧縮",
  "settings.compaction.timeout": "圧縮タイムアウト（秒）",
  "settings.compaction.timeoutAria": "圧縮タイムアウト",
  "settings.compaction.timeoutHint": "スタックした事前プロンプト/手動圧縮を中止し、永久にハングしないようにします。",
  "settings.compaction.backoffBase": "失敗バックオフ基準（分）",
  "settings.compaction.backoffBaseAria": "圧縮バックオフ基準",
  "settings.compaction.backoffBaseHint": "圧縮失敗後の最初の抑制ウィンドウ。",
  "settings.compaction.backoffMax": "失敗バックオフ最大（分）",
  "settings.compaction.backoffMaxAria": "圧縮バックオフ最大",
  "settings.compaction.backoffMaxHint": "繰り返し失敗した後の指数的抑制の上限。",
  "settings.compaction.decayFactor": "バックオフ減衰係数",
  "settings.compaction.decayFactorAria": "バックオフ減衰係数",
  "settings.compaction.decayFactorHint": "% — 圧縮が成功するたびにバックオフを半減",
  "settings.compaction.watchdogHeading": "ストール監視",
  "settings.compaction.enableWatchdog": "監視を有効化",
  "settings.compaction.enableWatchdogHint": "デフォルトで無効。有効にすると、アクティブフェーズがハートビートを停止した場合、ヘルパープロセスがランタイムを終了します。",
  "settings.compaction.watchdogTimeout": "監視タイムアウト（秒）",
  "settings.compaction.watchdogTimeoutAria": "監視タイムアウト",
  "settings.compaction.watchdogTimeoutHint": "監視がランタイムを強制終了するまでに、アクティブフェーズがハートビートなしで継続できる時間。",
  "settings.compaction.suppressionsHeading": "アクティブな圧縮抑制",
  "settings.compaction.noBackoff": "現在、圧縮バックオフ中のチャットはありません。",
  "settings.compaction.clear": "クリア",
  "settings.compaction.phasesHeading": "ライブ監視フェーズ",
  "settings.compaction.noPhases": "現在、追跡中のアクティブなフェーズはありません。",
  "menu.title": "メニュー",
  "menu.showWorkspace": "ワークスペースを表示",
  "menu.hideWorkspace": "ワークスペースを非表示",
  "menu.openExplorer": "エクスプローラーを開く",
  "menu.chatOnly": "チャットのみモード",
  "menu.exitChatOnly": "チャットのみモードを終了",
  "menu.openTerminal": "ターミナルをタブで開く",
  "menu.openVnc": "VNC をタブで開く",
  "menu.newFile": "新規ファイル",
  "menu.openRecent": "最近のファイルを開く",
  "menu.refreshTree": "ツリーを更新",
  "menu.reindex": "ワークスペースを再インデックス",
  "menu.showHidden": "隠しファイルを表示",
  "menu.hideHidden": "隠しファイルを非表示",
  "menu.scale": "拡大縮小",
  "menu.settings": "設定"
};
var TRANSLATIONS = {
  en: EN,
  "zh-CN": ZH_CN,
  ja: JA
};
var currentLocale = DEFAULT_LOCALE;
var initialized = false;
function normalizeLocale(value) {
  const raw = String(value ?? "").trim().toLowerCase().replace(/_/g, "-");
  if (!raw)
    return DEFAULT_LOCALE;
  if (raw === "zh-cn" || raw === "zh" || raw === "zh-hans" || raw.startsWith("zh-hans")) {
    return "zh-CN";
  }
  if (raw === "ja" || raw.startsWith("ja-"))
    return "ja";
  if (raw === "en" || raw.startsWith("en-"))
    return "en";
  return DEFAULT_LOCALE;
}
function detectBrowserLocale() {
  if (typeof navigator === "undefined")
    return DEFAULT_LOCALE;
  const candidates = [
    ...Array.isArray(navigator.languages) ? navigator.languages : [],
    navigator.language
  ].filter((c) => typeof c === "string" && c.length > 0);
  for (const candidate of candidates) {
    const normalized = normalizeLocale(candidate);
    if (normalized !== DEFAULT_LOCALE)
      return normalized;
  }
  return DEFAULT_LOCALE;
}
function resolveInitialLocale() {
  const stored = getLocalStorageItem(LOCALE_STORAGE_KEY);
  if (stored)
    return normalizeLocale(stored);
  return detectBrowserLocale();
}
function emitLocaleChange(locale) {
  if (typeof window === "undefined")
    return;
  window.dispatchEvent(new CustomEvent(LOCALE_CHANGE_EVENT, { detail: { locale } }));
}
function getLocale() {
  if (!initialized)
    initLocale();
  return currentLocale;
}
function initLocale() {
  currentLocale = resolveInitialLocale();
  initialized = true;
  return currentLocale;
}
function setLocale(value, options = {}) {
  const next = normalizeLocale(value);
  initialized = true;
  if (next === currentLocale && options.persist === false)
    return currentLocale;
  currentLocale = next;
  if (options.persist !== false)
    setLocalStorageItem(LOCALE_STORAGE_KEY, next);
  emitLocaleChange(next);
  return currentLocale;
}
function interpolate(template, vars) {
  if (!vars)
    return template;
  return template.replace(/\{(\w+)\}/g, (match, name) => {
    const replacement = vars[name];
    return replacement === undefined || replacement === null ? match : String(replacement);
  });
}
function translate(key, vars, locale = getLocale()) {
  const fromLocale = TRANSLATIONS[locale]?.[key];
  const template = fromLocale ?? EN[key] ?? key;
  return interpolate(template, vars);
}
function t(key, vars) {
  return translate(key, vars);
}
function useLocale() {
  const [locale, setLocaleState] = F_(getLocale());
  K_(() => {
    if (typeof window === "undefined" || typeof window.addEventListener !== "function")
      return;
    const handler = (event) => {
      const detail = event.detail;
      const next = normalizeLocale(detail?.locale ?? getLocale());
      setLocaleState(next);
    };
    window.addEventListener(LOCALE_CHANGE_EVENT, handler);
    setLocaleState(getLocale());
    return () => window.removeEventListener(LOCALE_CHANGE_EVENT, handler);
  }, []);
  return [locale, (value) => setLocale(value)];
}
function useTranslation() {
  const [locale, setLocaleValue] = useLocale();
  return {
    locale,
    setLocale: setLocaleValue,
    t: (key, vars) => translate(key, vars, locale)
  };
}

// web/src/components/language-switcher.ts
function buildLanguageOptions(activeLocale) {
  return SUPPORTED_LOCALES.map((value) => ({
    value,
    label: LOCALE_LABELS[value],
    active: value === activeLocale
  }));
}
function LanguageSwitcher({
  variant = "inline",
  onChange
} = {}) {
  const { locale, setLocale, t } = useTranslation();
  const options = buildLanguageOptions(locale);
  const handleChange = (event) => {
    const next = event?.currentTarget?.value;
    setLocale(next);
    onChange?.(next);
  };
  return fe`
    <div class=${`language-switcher language-switcher-${variant}`} role="none">
      <label class="language-switcher-label" for="language-switcher-select">${t("language.label")}</label>
      <select
        id="language-switcher-select"
        class="language-switcher-select"
        value=${locale}
        aria-label=${t("language.label")}
        onClick=${(event) => event.stopPropagation()}
        onChange=${handleChange}
      >
        ${options.map((option) => fe`
          <option key=${option.value} value=${option.value}>${option.label}</option>
        `)}
      </select>
    </div>
  `;
}

// web/src/components/timeline-menu.ts
function TimelineMenu({
  workspaceOpen,
  toggleWorkspace,
  chatOnlyMode,
  openEditor,
  onOpenTerminalTab,
  onOpenVncTab
}) {
  const { t } = useTranslation();
  const [open, setOpen] = F_(false);
  const [pwaDisplayScalePercent, setPwaDisplayScalePercent] = F_(() => readStoredPwaDisplayScalePercent());
  const [pwaDisplayScaleDraft, setPwaDisplayScaleDraft] = F_(() => String(readStoredPwaDisplayScalePercent()));
  const [showHidden, setShowHidden] = F_(() => {
    try {
      return localStorage.getItem("workspaceShowHidden") === "true";
    } catch {
      return false;
    }
  });
  const [pos, setPos] = F_({ top: 8, left: 8 });
  const getSafeAreaTop = () => {
    if (typeof document === "undefined")
      return 0;
    const probe = document.createElement("div");
    probe.style.cssText = "position:fixed;top:0;left:0;width:0;height:env(safe-area-inset-top,0px);visibility:hidden;pointer-events:none";
    document.body.appendChild(probe);
    const h = probe.offsetHeight;
    probe.remove();
    return h;
  };
  const menuRef = Q_(null);
  const btnRef = Q_(null);
  const portalRef = Q_(null);
  K_(() => {
    if (typeof document === "undefined")
      return;
    const host = document.createElement("div");
    host.className = "timeline-menu-portal in-chat";
    document.body.appendChild(host);
    portalRef.current = host;
    return () => {
      host.remove();
      portalRef.current = null;
    };
  }, []);
  K_(() => {
    const update = () => {
      const safeTop = getSafeAreaTop();
      const topOffset = safeTop > 0 ? safeTop + 4 : 8;
      if (workspaceOpen) {
        const sidebar = document.querySelector(".workspace-sidebar");
        if (sidebar) {
          const r = sidebar.getBoundingClientRect();
          setPos({ top: r.top + topOffset, left: r.left + 8 });
        }
      } else {
        setPos({ top: topOffset, left: 8 });
      }
    };
    update();
    const observer = new ResizeObserver(update);
    const sidebar = document.querySelector(".workspace-sidebar");
    if (sidebar)
      observer.observe(sidebar);
    window.addEventListener("resize", update);
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", update);
    };
  }, [workspaceOpen]);
  K_(() => {
    if (portalRef.current)
      portalRef.current.className = `timeline-menu-portal ${workspaceOpen ? "in-workspace" : "in-chat"}`;
  }, [workspaceOpen]);
  K_(() => {
    if (!portalRef.current)
      return;
    const s = portalRef.current.style;
    s.top = `${pos.top}px`;
    s.left = `${pos.left}px`;
    s.right = "auto";
  }, [pos]);
  K_(() => {
    setOpen(false);
  }, [workspaceOpen]);
  K_(() => {
    const syncPwaDisplayScale = () => {
      const next = readStoredPwaDisplayScalePercent();
      setPwaDisplayScalePercent(next);
      setPwaDisplayScaleDraft(String(next));
    };
    const onStorage = (event) => {
      if (!event || event.key === null || event.key === PWA_DISPLAY_SCALE_STORAGE_KEY)
        syncPwaDisplayScale();
    };
    window.addEventListener("storage", onStorage);
    window.addEventListener("focus", syncPwaDisplayScale);
    window.addEventListener(PWA_DISPLAY_SCALE_EVENT, syncPwaDisplayScale);
    return () => {
      window.removeEventListener("storage", onStorage);
      window.removeEventListener("focus", syncPwaDisplayScale);
      window.removeEventListener(PWA_DISPLAY_SCALE_EVENT, syncPwaDisplayScale);
    };
  }, []);
  const handlePwaDisplayScaleInput = Y_((event) => {
    setPwaDisplayScaleDraft(String(event?.currentTarget?.value ?? ""));
  }, []);
  const commitPwaDisplayScale = Y_((value) => {
    const next = persistPwaDisplayScalePercent(value);
    setPwaDisplayScalePercent(next);
    setPwaDisplayScaleDraft(String(next));
  }, []);
  const handlePwaDisplayScaleCommit = Y_((event) => {
    commitPwaDisplayScale(event?.currentTarget?.value);
  }, [commitPwaDisplayScale]);
  const handlePwaDisplayScaleKeyDown = Y_((event) => {
    if (event?.key === "Enter") {
      commitPwaDisplayScale(event?.currentTarget?.value);
      event?.currentTarget?.blur?.();
    }
  }, [commitPwaDisplayScale]);
  const run = Y_((fn) => {
    setOpen(false);
    fn?.();
  }, []);
  const toggleChatOnly = Y_(() => {
    const url = new URL(window.location.href);
    if (chatOnlyMode) {
      url.searchParams.delete("chat_only");
    } else {
      url.searchParams.set("chat_only", "1");
    }
    window.location.href = url.toString();
  }, [chatOnlyMode]);
  const content = fe`
        <button ref=${btnRef} class=${`timeline-menu-btn${open ? " active" : ""}`} data-testid="hamburger"
            onClick=${() => setOpen((v) => !v)} title=${t("menu.title")} aria-label=${t("menu.title")}
            aria-haspopup="menu" aria-expanded=${open ? "true" : "false"}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <line x1="4" y1="7" x2="20" y2="7" />
                <line x1="4" y1="12" x2="20" y2="12" />
                <line x1="4" y1="17" x2="20" y2="17" />
            </svg>
        </button>
        ${open && fe`
            <div class="workspace-menu-dropdown timeline-menu-dropdown" ref=${menuRef} role="menu">
                <button class="workspace-menu-item" role="menuitem" onClick=${() => run(toggleWorkspace)}>
                    ${workspaceOpen ? t("menu.hideWorkspace") : t("menu.showWorkspace")}
                </button>
                ${!workspaceOpen && !chatOnlyMode && fe`
                    <button class="workspace-menu-item" role="menuitem" onClick=${() => run(() => {
    toggleWorkspace();
  })}>
                        ${t("menu.openExplorer")}
                    </button>
                `}
                <button class=${`workspace-menu-item${chatOnlyMode ? " active" : ""}`} role="menuitem" onClick=${() => run(toggleChatOnly)}>
                    ${chatOnlyMode ? t("menu.exitChatOnly") : t("menu.chatOnly")}
                </button>

                ${(onOpenTerminalTab || onOpenVncTab) && fe`<div class="workspace-menu-separator"></div>`}
                ${onOpenTerminalTab && fe`<button class="workspace-menu-item" role="menuitem" onClick=${() => run(onOpenTerminalTab)}>${t("menu.openTerminal")}</button>`}
                ${onOpenVncTab && fe`<button class="workspace-menu-item" role="menuitem" onClick=${() => run(onOpenVncTab)}>${t("menu.openVnc")}</button>`}

                <div class="workspace-menu-separator"></div>
                <button class="workspace-menu-item" role="menuitem" disabled=${!workspaceOpen} onClick=${() => run(() => window.dispatchEvent(new CustomEvent("piclaw:workspace-action", { detail: { action: "new-file" } })))}>${t("menu.newFile")}</button>
                ${(() => {
    const recent = getRecentFiles();
    if (recent.length === 0)
      return null;
    return fe`
                        <div class="workspace-menu-separator"></div>
                        <div class="workspace-menu-submenu-label">${t("menu.openRecent")}</div>
                        ${recent.map((path) => {
      const label = path.split("/").pop() || path;
      return fe`
                                <button class="workspace-menu-item workspace-menu-recent-item" role="menuitem" title=${path} onClick=${() => run(() => openEditor?.(path))}>${label}</button>
                            `;
    })}
                    `;
  })()}
                <div class="workspace-menu-separator"></div>
                <button class="workspace-menu-item" role="menuitem" disabled=${!workspaceOpen} onClick=${() => run(() => window.dispatchEvent(new CustomEvent("piclaw:workspace-action", { detail: { action: "refresh" } })))}>${t("menu.refreshTree")}</button>
                <button class="workspace-menu-item" role="menuitem" disabled=${!workspaceOpen} onClick=${() => run(() => window.dispatchEvent(new CustomEvent("piclaw:workspace-action", { detail: { action: "reindex" } })))}>${t("menu.reindex")}</button>
                <button class=${`workspace-menu-item${showHidden ? " active" : ""}`} role="menuitem" disabled=${!workspaceOpen} onClick=${() => run(() => {
    const next = !showHidden;
    setShowHidden(next);
    try {
      localStorage.setItem("workspaceShowHidden", String(next));
    } catch {
      setShowHidden(next);
    }
    window.dispatchEvent(new CustomEvent("piclaw:toggle-hidden-files", { detail: { showHidden: next } }));
  })}>
                    ${showHidden ? t("menu.hideHidden") : t("menu.showHidden")}
                </button>
                <div class="workspace-menu-scale-control" role="none">
                    <label for="timeline-pwa-display-scale">${t("menu.scale")}</label>
                    <div class="workspace-menu-scale-input-wrap">
                        <input
                            id="timeline-pwa-display-scale"
                            class="workspace-menu-scale-input"
                            type="number"
                            inputmode="numeric"
                            min=${MIN_PWA_DISPLAY_SCALE_PERCENT}
                            max=${MAX_PWA_DISPLAY_SCALE_PERCENT}
                            step=${PWA_DISPLAY_SCALE_STEP_PERCENT}
                            value=${pwaDisplayScaleDraft}
                            aria-label=${`PWA display scale percentage, currently ${pwaDisplayScalePercent}%`}
                            onClick=${(event) => event.stopPropagation()}
                            onInput=${handlePwaDisplayScaleInput}
                            onChange=${handlePwaDisplayScaleCommit}
                            onBlur=${handlePwaDisplayScaleCommit}
                            onKeyDown=${handlePwaDisplayScaleKeyDown}
                        />
                        <span aria-hidden="true">%</span>
                    </div>
                </div>
                <button class="workspace-menu-item" role="menuitem" onClick=${() => run(() => window.dispatchEvent(new CustomEvent("piclaw:open-settings")))}>${t("menu.settings")}</button>
                <div class="workspace-menu-separator"></div>
                <div class="workspace-menu-language" role="none">
                    <${LanguageSwitcher} variant="menu" />
                </div>
            </div>
        `}
    `;
  W_(() => {
    if (portalRef.current)
      G_(content, portalRef.current);
  });
  W_(() => {
    if (!open || !menuRef.current || !btnRef.current)
      return;
    return bindMenuDismissal(menuRef.current, btnRef.current, () => setOpen(false));
  }, [open]);
  return null;
}

// web/src/gi-quick-actions-focus.ts
function quickActionsOpener(doc = document) {
  const active = doc.activeElement;
  if (active && active !== doc.body && active !== doc.documentElement && !active.closest("[inert]"))
    return active;
  return doc.querySelector('.container[aria-label="Conversation"]');
}
function bindQuickActionsFocus(root, input, opener, close) {
  const view = root.ownerDocument.defaultView;
  let closed = false;
  const restore = () => {
    if (opener?.isConnected && !opener.closest("[inert]") && !opener.hasAttribute("disabled") && opener.getClientRects().length)
      opener.focus({ preventScroll: true });
  };
  const frame = view.requestAnimationFrame(() => {
    if (!closed && !settingsOwnsKeyboard(root.ownerDocument) && input.isConnected && !root.closest("[inert]"))
      input.focus({ preventScroll: true });
  });
  const dismiss = () => {
    if (closed || settingsOwnsKeyboard(root.ownerDocument))
      return;
    closed = true;
    view.cancelAnimationFrame(frame);
    close();
    restore();
  };
  const detach = bindMenuDismissal(root, null, dismiss);
  const button = root.querySelector(".gi-quick-actions-close");
  button?.addEventListener("click", dismiss);
  return () => {
    closed = true;
    view.cancelAnimationFrame(frame);
    detach();
    button?.removeEventListener("click", dismiss);
  };
}

// web/src/ui/keyboard-shortcuts.ts
var STORAGE_KEY = "piclaw_keyboard_shortcuts_v1";
var KEYBOARD_SHORTCUT_ACTIONS = [
  {
    id: "openHelp",
    label: "Open keyboard help",
    description: "Open Settings → Keyboard. Default: question mark and quote when focus is outside compose and other editable fields.",
    defaultBindings: ["?", '"']
  },
  {
    id: "openSettings",
    label: "Open settings",
    description: "Open the settings dialog.",
    defaultBindings: ["ctrl+,", "meta+,", "alt+,"]
  },
  {
    id: "previousChat",
    label: "Previous session",
    description: "Switch to the previous visible chat/session.",
    defaultBindings: ["["]
  },
  {
    id: "nextChat",
    label: "Next session",
    description: "Switch to the next visible chat/session.",
    defaultBindings: ["]"]
  },
  {
    id: "toggleDock",
    label: "Toggle dock",
    description: "Show or hide the bottom dock panes.",
    defaultBindings: ["ctrl+`"]
  },
  {
    id: "toggleZenMode",
    label: "Toggle zen mode",
    description: "Collapse surrounding chrome for a focused chat view.",
    defaultBindings: ["ctrl+shift+z", "meta+shift+z"]
  }
];
var ACTION_MAP = new Map(KEYBOARD_SHORTCUT_ACTIONS.map((action) => [action.id, action]));
var MODIFIER_ALIASES = {
  cmd: "meta",
  command: "meta",
  meta: "meta",
  super: "meta",
  ctrl: "ctrl",
  control: "ctrl",
  alt: "alt",
  option: "alt",
  shift: "shift"
};
var KEY_ALIASES = {
  esc: "escape",
  return: "enter",
  spacebar: "space"
};
var NAMED_KEYS = new Set([
  "tab",
  "enter",
  "space",
  "backspace",
  "delete",
  "insert",
  "clear",
  "home",
  "end",
  "pageup",
  "pagedown",
  "up",
  "down",
  "left",
  "right"
]);
function normalizeKeyToken(token) {
  const trimmed = String(token || "").trim().toLowerCase();
  if (!trimmed)
    return null;
  const aliased = KEY_ALIASES[trimmed] || trimmed;
  if (/^f(?:[1-9]|1[0-2])$/.test(aliased))
    return aliased;
  if (NAMED_KEYS.has(aliased))
    return aliased;
  if (aliased.length === 1)
    return aliased;
  if (/^[a-z0-9]+$/.test(aliased))
    return aliased;
  return null;
}
function normalizeShortcutBindingString(value) {
  const raw = String(value || "").trim();
  if (!raw)
    return null;
  const parts = raw.split("+").map((part) => part.trim()).filter(Boolean);
  if (!parts.length)
    return null;
  const parsed = {
    ctrl: false,
    meta: false,
    alt: false,
    shift: false,
    key: ""
  };
  for (const part of parts) {
    const normalizedPart = part.toLowerCase();
    const modifier = MODIFIER_ALIASES[normalizedPart];
    if (modifier) {
      parsed[modifier] = true;
      continue;
    }
    if (parsed.key)
      return null;
    const key = normalizeKeyToken(part);
    if (!key || key === "escape")
      return null;
    parsed.key = key;
  }
  if (!parsed.key)
    return null;
  const segments = [];
  if (parsed.ctrl)
    segments.push("ctrl");
  if (parsed.meta)
    segments.push("meta");
  if (parsed.alt)
    segments.push("alt");
  if (parsed.shift)
    segments.push("shift");
  segments.push(parsed.key);
  return segments.join("+");
}
function readStoredShortcutConfig() {
  const stored = getLocalStorageJSON(STORAGE_KEY);
  if (!stored || typeof stored !== "object")
    return {};
  const next = {};
  for (const action of KEYBOARD_SHORTCUT_ACTIONS) {
    const raw = stored[action.id];
    if (!Array.isArray(raw))
      continue;
    const normalized = raw.map((entry) => normalizeShortcutBindingString(String(entry || ""))).filter((entry) => Boolean(entry));
    next[action.id] = [...new Set(normalized)];
  }
  return next;
}
function getKeyboardShortcutAction(actionId) {
  return ACTION_MAP.get(actionId);
}
function getKeyboardShortcutBindings(actionId) {
  const stored = readStoredShortcutConfig()[actionId];
  if (Array.isArray(stored))
    return stored;
  return [...getKeyboardShortcutAction(actionId).defaultBindings];
}
function normalizeEventKey(key) {
  const raw = typeof key === "string" ? key : "";
  if (!raw)
    return "";
  if (raw.length === 1)
    return raw.toLowerCase();
  return normalizeKeyToken(raw) || raw.toLowerCase();
}
function parseNormalizedBinding(binding) {
  const normalized = normalizeShortcutBindingString(binding);
  if (!normalized)
    return null;
  const parsed = {
    ctrl: false,
    meta: false,
    alt: false,
    shift: false,
    key: ""
  };
  for (const part of normalized.split("+")) {
    if (part === "ctrl" || part === "meta" || part === "alt" || part === "shift") {
      parsed[part] = true;
      continue;
    }
    parsed.key = part;
  }
  return parsed.key ? parsed : null;
}
function matchesShortcutBinding(event, binding) {
  const parsed = parseNormalizedBinding(binding);
  if (!parsed)
    return false;
  const normalizedEventKey = normalizeEventKey(event?.key);
  if (normalizedEventKey !== parsed.key)
    return false;
  const symbolKeyAllowsImplicitShift = !parsed.shift && parsed.key.length === 1 && /[^a-z0-9]/i.test(parsed.key);
  return Boolean(event?.ctrlKey) === parsed.ctrl && Boolean(event?.metaKey) === parsed.meta && Boolean(event?.altKey) === parsed.alt && (symbolKeyAllowsImplicitShift || Boolean(event?.shiftKey) === parsed.shift);
}
function matchesKeyboardShortcutAction(event, actionId) {
  return getKeyboardShortcutBindings(actionId).some((binding) => matchesShortcutBinding(event, binding));
}

// web/src/ui/timeline-quick-actions.ts
var WORKSPACE_QUICK_ACTIONS_CATALOG = [
  {
    id: "toggle-workspace",
    label: "Toggle workspace",
    description: "Show or hide the workspace sidebar.",
    keywords: ["workspace", "sidebar", "explorer"]
  },
  {
    id: "open-explorer",
    label: "Open explorer",
    description: "Open the workspace explorer sidebar.",
    keywords: ["workspace", "explorer", "sidebar"]
  },
  {
    id: "toggle-chat-only",
    label: "Chat-only mode",
    description: "Toggle chat-only mode.",
    keywords: ["chat", "mode", "layout"]
  },
  {
    id: "open-terminal-tab",
    label: "Open terminal in tab",
    description: "Open the terminal pane in a workspace tab.",
    keywords: ["terminal", "shell", "tab"]
  },
  {
    id: "open-vnc-tab",
    label: "Open VNC in tab",
    description: "Open the VNC viewer in a workspace tab.",
    keywords: ["vnc", "remote", "desktop", "tab"]
  },
  {
    id: "open-settings",
    label: "Settings",
    description: "Open the settings dialog.",
    keywords: ["settings", "preferences", "config"]
  }
];
function normalizeToken(value) {
  return String(value || "").toLowerCase().replace(/^[@/]+/, "").replace(/\s+/g, " ").trim();
}
function hasTrimmedString(value) {
  return typeof value === "string" && value.trim().length > 0;
}
function matchesTimelineQuickActionQuery(query, ...parts) {
  const normalizedQuery = normalizeToken(query);
  if (!normalizedQuery)
    return true;
  const haystack = parts.map((part) => normalizeToken(part)).filter(Boolean);
  for (const value of haystack) {
    if (value.startsWith(normalizedQuery) || value.includes(normalizedQuery)) {
      return true;
    }
  }
  return false;
}
function dedupeStringList(values) {
  if (!Array.isArray(values))
    return null;
  const out = [];
  const seen = new Set;
  for (const value of values) {
    const normalized = String(value || "").trim();
    if (!normalized)
      continue;
    const key = normalized.toLowerCase();
    if (seen.has(key))
      continue;
    seen.add(key);
    out.push(normalized);
  }
  return out;
}
function normalizeTimelineQuickActionsSettingsData(data) {
  const source = data && typeof data === "object" ? data : {};
  return {
    workspaceCommands: dedupeStringList(source.workspaceCommands),
    slashCommands: dedupeStringList(source.slashCommands)
  };
}
function isEnabledBySelection(selection, value) {
  if (!Array.isArray(selection))
    return true;
  return selection.some((entry) => entry.toLowerCase() === value.toLowerCase());
}
function buildWorkspaceQuickActionItems(options) {
  const commands = Array.isArray(options?.commands) ? options.commands : [];
  const settings = normalizeTimelineQuickActionsSettingsData(options?.settings);
  const query = String(options?.query || "");
  return commands.filter((command) => isEnabledBySelection(settings.workspaceCommands, command.id)).filter((command) => matchesTimelineQuickActionQuery(query, command.label, command.description, ...command.keywords || [])).map((command) => ({
    key: `workspace:${command.id}`,
    kind: "workspace",
    title: command.label,
    subtitle: command.description,
    searchText: `${command.label} ${command.description} ${(command.keywords || []).join(" ")}`.trim(),
    visualHint: command.label.slice(0, 1).toUpperCase() || "W",
    categoryLabel: "Workspace",
    actionHint: "Run",
    commandId: command.id
  }));
}
function buildAgentQuickActionItems(options) {
  const agents = Array.isArray(options?.agents) ? options.agents : [];
  const query = String(options?.query || "");
  const seen = new Set;
  return agents.filter((agent) => {
    const chatJid = hasTrimmedString(agent?.chat_jid) ? agent.chat_jid.trim() : "";
    if (!chatJid || seen.has(chatJid))
      return false;
    if (agent?.archived_at)
      return false;
    seen.add(chatJid);
    return true;
  }).filter((agent) => matchesTimelineQuickActionQuery(query, `@${String(agent?.agent_name || "").trim()}`, agent?.session_name, agent?.chat_jid)).map((agent) => {
    const agentName = hasTrimmedString(agent?.agent_name) ? agent.agent_name.trim() : String(agent?.chat_jid || "").replace(/^[^:]+:/, "");
    const sessionName = hasTrimmedString(agent?.session_name) ? agent.session_name.trim() : "";
    const chatJid = String(agent?.chat_jid || "").trim();
    return {
      key: `agent:${chatJid}`,
      kind: "agent",
      title: `@${agentName}`,
      subtitle: sessionName || chatJid,
      searchText: `@${agentName} ${sessionName} ${chatJid}`.trim(),
      visualHint: agentName.slice(0, 1).toUpperCase() || "@",
      categoryLabel: "Agent",
      actionHint: "Open",
      chatJid
    };
  });
}
function buildSlashQuickActionItems(options) {
  const slashCommands = Array.isArray(options?.slashCommands) ? options.slashCommands : [];
  const settings = normalizeTimelineQuickActionsSettingsData(options?.settings);
  const query = String(options?.query || "");
  const seen = new Set;
  return slashCommands.filter((command) => {
    const name = hasTrimmedString(command?.name) ? command.name.trim() : "";
    if (!name || seen.has(name.toLowerCase()))
      return false;
    seen.add(name.toLowerCase());
    return isEnabledBySelection(settings.slashCommands, name);
  }).filter((command) => matchesTimelineQuickActionQuery(query, command?.name, command?.description, command?.source)).map((command) => {
    const name = String(command?.name || "").trim();
    const description = hasTrimmedString(command?.description) ? command.description.trim() : "slash command";
    const source = hasTrimmedString(command?.source) ? command.source.trim() : "";
    return {
      key: `slash:${name}`,
      kind: "slash",
      title: name,
      subtitle: description,
      searchText: `${name} ${description} ${String(command?.source || "")}`.trim(),
      visualHint: "/",
      categoryLabel: source || "Slash",
      actionHint: "Insert",
      commandName: name
    };
  });
}
function buildTimelineQuickActionItems(options) {
  return [
    ...buildAgentQuickActionItems({ agents: options?.agents, query: options?.query }),
    ...buildWorkspaceQuickActionItems({ commands: options?.workspaceCommands, settings: options?.settings, query: options?.query }),
    ...buildSlashQuickActionItems({ slashCommands: options?.slashCommands, settings: options?.settings, query: options?.query })
  ];
}

// web/src/components/timeline-quick-actions.ts
function isEditableTarget(target) {
  if (!target || typeof target !== "object")
    return false;
  if (target.isContentEditable)
    return true;
  if (typeof target.closest !== "function")
    return false;
  return Boolean(target.closest([
    "input",
    "textarea",
    "select",
    '[contenteditable="true"]',
    ".compose-box",
    ".compose-model-popup",
    ".compose-session-popup",
    ".settings-dialog",
    ".workspace-sidebar",
    ".workspace-explorer",
    ".editor-pane-container",
    ".dock-panel",
    ".timeline-menu-dropdown",
    ".rename-branch-overlay",
    ".agent-request-modal",
    ".attachment-preview-modal",
    ".vnc-pane-shell",
    ".kanban-plugin"
  ].join(", ")));
}
function isEligibleTimelineTarget(target) {
  if (!target || typeof target !== "object")
    return true;
  if (isEditableTarget(target))
    return false;
  const tagName = String(target.tagName || "").toUpperCase();
  if (tagName === "BODY" || tagName === "HTML")
    return true;
  if (typeof target.closest !== "function")
    return true;
  return Boolean(target.closest(".container, .timeline, .post, .post-body, .post-content, .agent-status-panel"));
}
function shouldOpenTimelineQuickActionsFromKeyEvent(event) {
  if (blocksQuickActions(event, isQuickActionsReady()))
    return false;
  if (!isPopupTypeaheadKey(event))
    return false;
  if (!isEligibleTimelineTarget(event?.target))
    return false;
  const matchesShortcut = KEYBOARD_SHORTCUT_ACTIONS.some((action) => matchesKeyboardShortcutAction(event, action.id));
  return !matchesShortcut;
}
function toggleChatOnlyMode(chatOnlyMode) {
  const url = new URL(window.location.href);
  if (chatOnlyMode) {
    url.searchParams.delete("chat_only");
  } else {
    url.searchParams.set("chat_only", "1");
  }
  window.location.href = url.toString();
}
function buildWorkspaceCommands(options) {
  const commands = [];
  const byId = new Map(WORKSPACE_QUICK_ACTIONS_CATALOG.map((entry) => [entry.id, entry]));
  const add = (id, overrides = {}) => {
    const base = byId.get(id);
    if (!base)
      return;
    commands.push({ ...base, ...overrides });
  };
  add("toggle-workspace", {
    label: options.workspaceOpen ? t("palette.hideWorkspace") : t("palette.showWorkspace"),
    description: options.workspaceOpen ? t("palette.hideWorkspaceDesc") : t("palette.showWorkspaceDesc")
  });
  if (!options.workspaceOpen && !options.chatOnlyMode) {
    add("open-explorer");
  }
  add("toggle-chat-only", {
    label: options.chatOnlyMode ? t("palette.exitChatOnly") : t("palette.chatOnly"),
    description: options.chatOnlyMode ? t("palette.exitChatOnlyDesc") : t("palette.chatOnlyDesc")
  });
  if (typeof options.onOpenTerminalTab === "function")
    add("open-terminal-tab");
  if (typeof options.onOpenVncTab === "function")
    add("open-vnc-tab");
  add("open-settings");
  return commands;
}
function sectionLabel(kind) {
  if (kind === "agent")
    return t("palette.groupAgents");
  if (kind === "workspace")
    return t("palette.groupWorkspace");
  return t("palette.groupSlash");
}
function renderQuickActionMedia(item) {
  if (item?.imageUrl) {
    return fe`<img class="timeline-quick-actions-item-avatar" src=${item.imageUrl} alt="" aria-hidden="true" />`;
  }
  return fe`<span class="timeline-quick-actions-item-placeholder" aria-hidden="true">${item?.visualHint || ""}</span>`;
}
function renderKeyboardHint(label, value) {
  return fe`
        <span class="timeline-quick-actions-keyhint">
            <kbd>${value}</kbd>
            <span>${label}</span>
        </span>
    `;
}
function openInNewTab(chatJid) {
  const url = new URL(window.location.href);
  url.searchParams.set("chat_jid", chatJid);
  url.searchParams.set("chat_only", "1");
  const a = document.createElement("a");
  a.href = url.toString();
  a.target = "_blank";
  a.rel = "noopener";
  a.style.display = "none";
  document.body.appendChild(a);
  a.click();
  a.remove();
}
function TimelineQuickActions({
  activeChatAgents = [],
  currentChatJid = "web:default",
  workspaceOpen = false,
  chatOnlyMode = false,
  onSwitchChat,
  onToggleWorkspace,
  onOpenTerminalTab,
  onOpenVncTab,
  onPrefillCompose
}) {
  const [open, setOpen] = F_(false);
  const [query, setQuery] = F_("");
  const [highlightIndex, setHighlightIndex] = F_(0);
  const [slashCommands, setSlashCommands] = F_([]);
  const [settings, setSettings] = F_({ workspaceCommands: null, slashCommands: null });
  const rootRef = Q_(null);
  const openerRef = Q_(null);
  const inputRef = Q_(null);
  const loadSettings = Y_(async () => {
    try {
      const payload = await getQuickActionsSettings();
      setSettings(normalizeTimelineQuickActionsSettingsData(payload?.settings));
    } catch {
      setSettings({ workspaceCommands: null, slashCommands: null });
    }
  }, []);
  K_(() => {
    loadSettings();
  }, [loadSettings]);
  K_(() => {
    let cancelled = false;
    getAgentCommands(currentChatJid).then((payload) => {
      if (cancelled)
        return;
      setSlashCommands(Array.isArray(payload?.commands) ? payload.commands : []);
    }).catch(() => {
      if (cancelled)
        return;
      setSlashCommands([]);
    });
    return () => {
      cancelled = true;
    };
  }, [currentChatJid]);
  const workspaceCommands = u_(() => buildWorkspaceCommands({
    workspaceOpen,
    chatOnlyMode,
    onOpenTerminalTab,
    onOpenVncTab
  }), [chatOnlyMode, onOpenTerminalTab, onOpenVncTab, workspaceOpen]);
  const items = u_(() => buildTimelineQuickActionItems({
    agents: activeChatAgents,
    workspaceCommands,
    slashCommands,
    settings,
    query
  }), [activeChatAgents, query, settings, slashCommands, workspaceCommands]);
  K_(() => {
    if (items.length === 0) {
      setHighlightIndex(-1);
      return;
    }
    if (!query.trim()) {
      setHighlightIndex(0);
      return;
    }
    const normalizedQuery = query.toLowerCase().replace(/^[@/]+/, "").trim();
    if (!normalizedQuery) {
      setHighlightIndex(0);
      return;
    }
    let bestIndex = 0;
    let bestScore = 0;
    for (let i = 0;i < items.length; i++) {
      const item = items[i];
      const title = (item.title || "").toLowerCase().replace(/^[@/]+/, "");
      if (title === normalizedQuery) {
        bestIndex = i;
        break;
      }
      let score = 0;
      if (title.startsWith(normalizedQuery)) {
        score = 3;
      } else if (title.includes(normalizedQuery)) {
        score = 2;
      } else if ((item.subtitle || "").toLowerCase().includes(normalizedQuery)) {
        score = 1;
      }
      if (score > bestScore) {
        bestScore = score;
        bestIndex = i;
      }
    }
    setHighlightIndex(bestIndex);
  }, [items, query]);
  W_(() => {
    if (!open || !rootRef.current || !inputRef.current)
      return;
    return bindQuickActionsFocus(rootRef.current, inputRef.current, openerRef.current, () => {
      setOpen(false);
      setQuery("");
    });
  }, [open]);
  W_(() => {
    const onKeyDown = (event) => {
      if (settingsOwnsKeyboard())
        return;
      if (rootRef.current?.contains(event.target) && event.target?.closest?.("button"))
        return;
      if (!open) {
        if (!shouldOpenTimelineQuickActionsFromKeyEvent(event))
          return;
        event.preventDefault();
        openerRef.current = quickActionsOpener();
        setQuery(String(event.key || ""));
        setHighlightIndex(0);
        setOpen(true);
        return;
      }
      if (event.key === "Escape")
        return;
      if (event.key === "ArrowDown") {
        event.preventDefault();
        setHighlightIndex((prev) => items.length > 0 ? (prev + 1 + items.length) % items.length : 0);
        return;
      }
      if (event.key === "ArrowUp") {
        event.preventDefault();
        setHighlightIndex((prev) => items.length > 0 ? (prev - 1 + items.length) % items.length : 0);
        return;
      }
      if (event.key === "Enter" && items[highlightIndex]) {
        event.preventDefault();
        const current = items[highlightIndex];
        const popOut = event.altKey;
        if (current) {
          if (current.kind === "agent" && current.chatJid) {
            if (popOut) {
              openInNewTab(current.chatJid);
            } else {
              onSwitchChat?.(current.chatJid);
            }
          } else if (current.kind === "workspace" && current.commandId) {
            if (current.commandId === "toggle-workspace" || current.commandId === "open-explorer")
              onToggleWorkspace?.();
            if (current.commandId === "toggle-chat-only")
              toggleChatOnlyMode(chatOnlyMode);
            if (current.commandId === "open-terminal-tab")
              onOpenTerminalTab?.();
            if (current.commandId === "open-vnc-tab")
              onOpenVncTab?.();
            if (current.commandId === "open-settings")
              window.dispatchEvent(new CustomEvent("piclaw:open-settings"));
          } else if (current.kind === "slash" && current.commandName) {
            onPrefillCompose?.(current.commandName);
          }
        }
        setOpen(false);
        setQuery("");
      }
    };
    window.addEventListener("keydown", onKeyDown, true);
    return () => {
      window.removeEventListener("keydown", onKeyDown, true);
    };
  }, [chatOnlyMode, highlightIndex, items, onOpenTerminalTab, onOpenVncTab, onPrefillCompose, onSwitchChat, onToggleWorkspace, open]);
  K_(() => {
    const handleSettingsSaved = (event) => {
      const nextSettings = normalizeTimelineQuickActionsSettingsData(event?.detail?.settings);
      if (event?.detail?.settings) {
        setSettings(nextSettings);
        return;
      }
      loadSettings();
    };
    window.addEventListener("focus", handleSettingsSaved);
    window.addEventListener("piclaw:quick-actions-settings-updated", handleSettingsSaved);
    return () => {
      window.removeEventListener("focus", handleSettingsSaved);
      window.removeEventListener("piclaw:quick-actions-settings-updated", handleSettingsSaved);
    };
  }, [loadSettings]);
  if (!open)
    return null;
  let lastKind = null;
  return fe`
        <div class="timeline-quick-actions-portal">
            <div class="timeline-quick-actions-overlay">
                <div class="timeline-quick-actions" ref=${rootRef}>
                    <div class="timeline-quick-actions-header">
                        <div class="timeline-quick-actions-search-row">
                            <input
                                ref=${inputRef}
                                class="timeline-quick-actions-input"
                                type="text"
                                value=${query}
                                placeholder=${t("palette.placeholder")}
                                onInput=${(event) => {
    setQuery(event.currentTarget?.value || "");
    setHighlightIndex(0);
  }}
                            />
                            <button type="button" class="gi-quick-actions-close" aria-label="Close quick actions" title="Close quick actions"><span aria-hidden="true">×</span></button>
                            <div class="timeline-quick-actions-hints" aria-hidden="true">
                                ${renderKeyboardHint(t("palette.hintMove"), "↑↓")}
                                ${renderKeyboardHint(t("palette.hintSelect"), "↵")}
                                ${renderKeyboardHint(t("palette.hintPopOut"), "Alt+↵")}
                                ${renderKeyboardHint(t("palette.hintClose"), "Esc")}
                            </div>
                        </div>
                    </div>
                    <div class="timeline-quick-actions-list">
                        ${items.length === 0 && fe`<div class="timeline-quick-actions-empty">No quick actions match.</div>`}
                        ${items.map((item, index) => {
    const showSection = item.kind !== lastKind;
    lastKind = item.kind;
    return fe`
                                ${showSection && fe`<div class="timeline-quick-actions-section">${sectionLabel(item.kind)}</div>`}
                                <button
                                    key=${item.key}
                                    type="button"
                                    class=${`timeline-quick-actions-item timeline-quick-actions-item-${item.kind}${index === highlightIndex ? " active" : ""}`}
                                    onMouseEnter=${null}
                                    onClick=${() => {
      if (item.kind === "agent" && item.chatJid)
        onSwitchChat?.(item.chatJid);
      if (item.kind === "workspace" && item.commandId === "toggle-workspace")
        onToggleWorkspace?.();
      if (item.kind === "workspace" && item.commandId === "open-explorer")
        onToggleWorkspace?.();
      if (item.kind === "workspace" && item.commandId === "toggle-chat-only")
        toggleChatOnlyMode(chatOnlyMode);
      if (item.kind === "workspace" && item.commandId === "open-terminal-tab")
        onOpenTerminalTab?.();
      if (item.kind === "workspace" && item.commandId === "open-vnc-tab")
        onOpenVncTab?.();
      if (item.kind === "workspace" && item.commandId === "open-settings")
        window.dispatchEvent(new CustomEvent("piclaw:open-settings"));
      if (item.kind === "slash" && item.commandName)
        onPrefillCompose?.(item.commandName);
      setOpen(false);
      setQuery("");
    }}
                                >
                                    <span class="timeline-quick-actions-item-media">
                                        ${renderQuickActionMedia(item)}
                                    </span>
                                    <span class="timeline-quick-actions-item-copy">
                                        <span class="timeline-quick-actions-item-title-row">
                                            <span class="timeline-quick-actions-item-title">${item.title}</span>
                                            ${item.actionHint ? fe`<span class="timeline-quick-actions-item-action-hint">${item.actionHint}</span>` : null}
                                        </span>
                                        <span class="timeline-quick-actions-item-subtitle">${item.subtitle}</span>
                                    </span>
                                    <span class="timeline-quick-actions-item-category">${item.categoryLabel || sectionLabel(item.kind)}</span>
                                </button>
                            `;
  })}
                    </div>
                </div>
            </div>
        </div>
    `;
}

// web/src/gi-settings-lazy.ts
var loaders = {
  models: () => import("./gi-settings-models-wq929588.js").then((module) => module.Models),
  appearance: () => import("./gi-settings-appearance-3nkkfng3.js").then((module) => module.Appearance),
  compaction: () => import("./gi-settings-compaction-8eztqbag.js").then((module) => module.GiSettingsCompaction),
  providers: () => import("./gi-settings-providers-0be8zmt2.js").then((module) => module.GiSettingsProviders)
};
var labels = { models: "Models", appearance: "Appearance", compaction: "Compaction", providers: "Providers" };
var components = new Map;
var pending = new Map;
function load(section) {
  if (components.has(section))
    return Promise.resolve(components.get(section));
  if (pending.has(section))
    return pending.get(section);
  const promise = loaders[section]().then((component) => {
    components.set(section, component);
    pending.delete(section);
    return component;
  }, (error) => {
    pending.delete(section);
    throw error;
  });
  pending.set(section, promise);
  return promise;
}
function LazySettingsPane({ section, chatJid, filter, onMutationStart, onMutationEnd, onApplied }) {
  const [component, setComponent] = F_(() => components.get(section) || null);
  const [error, setError] = F_(false);
  K_(() => {
    let live = true;
    if (!component)
      load(section).then((value) => {
        if (live)
          setComponent(() => value);
      }).catch(() => {
        if (live)
          setError(true);
      });
    return () => {
      live = false;
    };
  }, []);
  if (error)
    return fe`<div role="alert">Unable to load ${labels[section]}. Close Settings and try again. If the app was updated, save your work and reload the page.</div>`;
  if (!component)
    return fe`<div role="status" class="settings-loading-pane">Loading ${labels[section]} pane…</div>`;
  return fe`<${component} key=${chatJid} chatJid=${chatJid} filter=${filter} onMutationStart=${onMutationStart} onMutationEnd=${onMutationEnd} onApplied=${onApplied} />`;
}

// web/src/gi-settings.ts
var generalCache = null;
function Identity() {
  const [snapshot, setSnapshot] = F_(null);
  const [draft, setDraft] = F_({ assistant_name: "", user_name: "" });
  const [error, setError] = F_("");
  const [notice, setNotice] = F_("");
  const [busy, setBusy] = F_(false);
  const [attempt, setAttempt] = F_(0);
  const live = Q_(false);
  const saving = Q_(false);
  K_(() => {
    live.current = true;
    return () => {
      live.current = false;
    };
  }, []);
  K_(() => {
    let active = true;
    setError("");
    setNotice("");
    setSnapshot(null);
    getGiIdentity().then((value) => {
      if (active) {
        setSnapshot(value);
        setDraft({ assistant_name: value.saved.assistant_name, user_name: value.saved.user_name });
      }
    }).catch((err) => {
      if (active)
        setError(err.message);
    });
    return () => {
      active = false;
    };
  }, [attempt]);
  async function save() {
    if (!snapshot || saving.current)
      return;
    setError("");
    setNotice("");
    if ([draft.assistant_name, draft.user_name].some((name) => !name.trim() || [...name].length > 128 || /[\u0000-\u001f\u007f-\u009f]/.test(name))) {
      setError("Names must contain 1–128 characters without control characters.");
      return;
    }
    saving.current = true;
    setBusy(true);
    try {
      const value = await saveGiIdentity({ ...draft, revision: snapshot.saved.revision });
      if (live.current) {
        setSnapshot(value);
        setDraft({ assistant_name: value.saved.assistant_name, user_name: value.saved.user_name });
        setNotice(value.restart_required ? "Names saved. Restart Gi manually to activate them." : "Names saved. Active names already match.");
      }
    } catch (err) {
      if (live.current)
        setError(err.message);
    } finally {
      saving.current = false;
      if (live.current)
        setBusy(false);
    }
  }
  return fe`<section aria-label="Saved display names">
        <h3>Saved display names</h3>
        <p>Instance-wide · saved to .piclaw/config.json. Active names above stay unchanged until you restart Gi manually.</p>
        ${!snapshot && !error && fe`<p role="status">Loading saved names…</p>`}
        ${snapshot && fe`<label>Assistant display name<input aria-label="Assistant display name" type="text" value=${draft.assistant_name} disabled=${busy} onInput=${(e) => {
    setDraft((d) => ({ ...d, assistant_name: e.target.value }));
    setNotice("");
  }} /></label>
            <label>User display name<input aria-label="User display name" type="text" value=${draft.user_name} disabled=${busy} onInput=${(e) => {
    setDraft((d) => ({ ...d, user_name: e.target.value }));
    setNotice("");
  }} /></label>
            ${snapshot.restart_required && fe`<p data-testid="identity-restart-required">Restart required to activate the saved names.</p>`}
            <button disabled=${busy} onClick=${save}>${busy ? "Saving names…" : "Save names"}</button>`}
        <button disabled=${busy} onClick=${() => setAttempt((n) => n + 1)}>Reload saved names</button>
        ${error && fe`<p role="alert">${error}</p>`}
        ${notice && fe`<p role="status">${notice}</p>`}
    </section>`;
}
function General() {
  const [data, setData] = F_(generalCache);
  const [error, setError] = F_("");
  const [attempt, setAttempt] = F_(0);
  K_(() => {
    let live = true;
    setError("");
    getGiSettingsSnapshot().then((snapshot) => {
      if (live) {
        generalCache = snapshot;
        setData(snapshot);
      }
    }).catch((error) => {
      if (live)
        setError(error.message);
    });
    return () => {
      live = false;
    };
  }, [attempt]);
  return fe`<section aria-labelledby="gi-general-title">
        <h2 id="gi-general-title">General</h2>
        <p>Active instance settings · read-only</p>
        <p>Loaded at startup from <code>.piclaw/config.json</code> and <code>.pi/settings.json</code>. Edit the files and restart Gi to change these defaults.</p>
        ${error && fe`<div role="alert">${error} <button onClick=${() => setAttempt((x) => x + 1)}>Retry</button></div>`}
        ${!data && !error && fe`<p role="status">Loading settings…</p>`}
        ${data && fe`<dl class="gi-settings-values">
            <dt>Assistant</dt><dd>${data.assistant_name}</dd>
            <dt>User</dt><dd>${data.user_name}</dd>
            <dt>Workspace</dt><dd>${data.workspace_root}</dd>
            <dt>Default model</dt><dd>${data.current || data.default_model}</dd>
            <dt>Default thinking</dt><dd>${data.default_thinking_level || "Unknown"}</dd>
            <dt>Build</dt><dd>${data.version || "Unknown"}</dd>
        </dl>`}
        <${Identity} />
    </section>`;
}
function Dialog({ chatJid, onClose, onMutationStart, onMutationEnd, onApplied }) {
  const [section, setSection] = F_("general");
  const dialog = Q_(null);
  const filterRef = Q_(null);
  const [filter, setFilter] = F_("");
  const [busyScope, setBusyScope] = F_(null);
  const searchScope = u_(() => ({}), [section, chatJid]);
  const [layoutMode, setLayoutMode] = F_({ compact: false, narrow: false });
  W_(() => {
    const element = dialog.current;
    if (!element)
      return;
    const update = () => {
      const width = element.clientWidth || 0;
      setLayoutMode((previous) => {
        const next = { compact: width > 0 && width <= 860, narrow: width > 0 && width <= 720 };
        return previous.compact === next.compact && previous.narrow === next.narrow ? previous : next;
      });
    };
    update();
    if (typeof ResizeObserver === "function") {
      const observer = new ResizeObserver(update);
      observer.observe(element);
      return () => observer.disconnect();
    }
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, []);
  W_(() => {
    setFilter("");
    if (section === "models")
      filterRef.current?.focus();
  }, [section, chatJid]);
  W_(() => {
    const app = document.getElementById("app");
    const previousInert = app?.inert;
    if (app)
      app.inert = true;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    dialog.current?.querySelector(".settings-dialog-close")?.focus();
    const key = (event) => {
      if (event.isComposing)
        return;
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopImmediatePropagation();
        onClose();
        return;
      }
      if (event.key === "Tab") {
        const nodes = [...dialog.current?.querySelectorAll('button:not(:disabled), input:not(:disabled), select:not(:disabled), [tabindex="0"]') || []].filter((node) => node.getClientRects().length);
        const first = nodes[0], last = nodes.at(-1);
        if (event.shiftKey && (document.activeElement === first || !dialog.current?.contains(document.activeElement))) {
          event.preventDefault();
          last?.focus();
        } else if (!event.shiftKey && (document.activeElement === last || !dialog.current?.contains(document.activeElement))) {
          event.preventDefault();
          first?.focus();
        }
      }
      if (event.defaultPrevented)
        event.stopImmediatePropagation();
    };
    window.addEventListener("keydown", key, true);
    return () => {
      window.removeEventListener("keydown", key, true);
      if (app)
        app.inert = previousInert || false;
      document.body.style.overflow = previousOverflow;
    };
  }, []);
  return fe`<div class="settings-dialog-backdrop" onClick=${(e) => {
    if (e.target === e.currentTarget)
      onClose();
  }}>
        <div ref=${dialog} class=${`settings-dialog${layoutMode.compact ? " settings-dialog-compact" : ""}${layoutMode.narrow ? " settings-dialog-narrow" : ""}`} role="dialog" aria-modal="true" aria-labelledby="gi-settings-title" onKeyDown=${(e) => e.stopPropagation()}>
            <header class="settings-dialog-header"><span class="settings-dialog-title" id="gi-settings-title">Gi Settings</span>
                ${section === "models" && fe`<input ref=${filterRef} type="search" class="settings-header-filter" aria-label="Filter models" placeholder="Filter models…" value=${filter} disabled=${busyScope === searchScope} onInput=${(e) => setFilter(e.target.value)} />`}
                <button class="settings-dialog-close" aria-label="Close settings" onClick=${onClose}>✕</button></header>
            <div class="settings-dialog-body"><nav class="settings-nav" aria-label="Settings sections">
                ${["general", "models", "appearance", "compaction", "providers"].map((id) => fe`<button class=${`settings-nav-item ${section === id ? "active" : ""}`} aria-current=${section === id ? "page" : undefined} onClick=${() => setSection(id)}>${{ general: "General", models: "Models", appearance: "Appearance", compaction: "Compaction", providers: "Providers" }[id]}</button>`)}
            </nav><main class="settings-content">
                ${section === "general" ? fe`<${General} />` : fe`<${LazySettingsPane} key=${section} section=${section} chatJid=${chatJid} filter=${filter} onMutationStart=${() => {
    setBusyScope(searchScope);
    return onMutationStart();
  }} onMutationEnd=${(token) => {
    setBusyScope((previous) => previous === searchScope ? null : previous);
    onMutationEnd(token);
  }} onApplied=${onApplied} />`}
            </main></div>
        </div>
    </div>`;
}
function GiSettings({ chatJid, onMutationStart, onMutationEnd, onApplied }) {
  const [open, setOpen] = F_(false);
  const opener = Q_(null);
  const isOpen = Q_(false);
  const close = () => {
    isOpen.current = false;
    setOpen(false);
    requestAnimationFrame(() => {
      if (isOpen.current)
        return;
      const original = opener.current;
      const target = original?.isConnected && original !== document.body && !original.closest("[inert]") ? original : document.querySelector(".compose-box textarea");
      target?.focus({ preventScroll: true });
    });
  };
  K_(() => {
    const show = () => {
      if (!isOpen.current)
        opener.current = document.activeElement;
      isOpen.current = true;
      setOpen(true);
    };
    const shortcut = (event) => {
      if (event.key === "," && (event.ctrlKey || event.metaKey || event.altKey)) {
        event.preventDefault();
        event.stopImmediatePropagation();
        show();
      }
    };
    window.addEventListener("piclaw:open-settings", show);
    window.addEventListener("keydown", shortcut, true);
    return () => {
      window.removeEventListener("piclaw:open-settings", show);
      window.removeEventListener("keydown", shortcut, true);
    };
  }, []);
  return open && fe`<${BodyPortal} className="settings-portal"><${Dialog} chatJid=${chatJid} onClose=${close} onMutationStart=${onMutationStart} onMutationEnd=${onMutationEnd} onApplied=${onApplied} /><//>`;
}

// web/src/gi-message-deletion.ts
function createMessageDeletionState() {
  const removed = new Set;
  const pending = new Set;
  return {
    begin(id) {
      if (pending.has(id) || removed.has(id))
        return false;
      pending.add(id);
      return true;
    },
    finish(id, success) {
      pending.delete(id);
      if (success)
        removed.add(id);
    },
    filter(rows, animating = new Set) {
      return rows.filter((row) => !removed.has(row.id) || animating.has(row.id));
    }
  };
}

// web/src/ui/chat-swipe-navigation.ts
function hasClosest(target) {
  return Boolean(target && typeof target.closest === "function");
}
function createChatSwipeTouchState() {
  return {
    active: false,
    horizontalLocked: false,
    cancelled: false,
    startX: 0,
    startY: 0,
    lastX: 0,
    lastY: 0,
    startedAt: 0
  };
}
function createChatSwipeWheelState() {
  return {
    lastTriggeredAt: 0,
    accumX: 0
  };
}
function resetChatSwipeTouchState(state) {
  if (!state)
    return;
  state.active = false;
  state.horizontalLocked = false;
  state.cancelled = false;
  state.startX = 0;
  state.startY = 0;
  state.lastX = 0;
  state.lastY = 0;
  state.startedAt = 0;
}
var INTERACTIVE_SELECTOR = [
  "input",
  "textarea",
  "select",
  "button",
  "label",
  "a[href]",
  '[contenteditable="true"]',
  '[role="button"]',
  "[data-no-chat-swipe]",
  ".compose-box",
  ".compose-model-popup",
  ".compose-session-popup",
  ".workspace-explorer",
  ".editor-pane-container",
  ".dock-panel",
  ".terminal-pane-content",
  ".attachment-preview-modal",
  ".rename-branch-overlay",
  ".agent-request-modal",
  ".adaptive-card-container input",
  ".adaptive-card-container textarea",
  ".adaptive-card-container select",
  ".adaptive-card-container button"
].join(", ");
var SWIPE_PASSTHROUGH_ANCESTOR = [
  ".agent-thinking",
  ".agent-status-panel",
  ".agent-thinking-intent"
].join(", ");
function isEligibleChatSwipeTarget(target) {
  if (!target || !hasClosest(target))
    return false;
  const interactiveMatch = target.closest(INTERACTIVE_SELECTOR);
  if (!interactiveMatch)
    return true;
  return Boolean(interactiveMatch.closest(SWIPE_PASSTHROUGH_ANCESTOR));
}
function resolveSwipeableChatAgents(candidates, currentChatJid) {
  if (!Array.isArray(candidates))
    return currentChatJid ? [currentChatJid] : [];
  const seen = new Set;
  const candidateRows = candidates.filter((candidate) => Boolean(candidate && typeof candidate === "object")).filter((candidate) => {
    const chatJid = typeof candidate.chat_jid === "string" ? candidate.chat_jid.trim() : "";
    if (!chatJid || seen.has(chatJid))
      return false;
    if (candidate.archived_at)
      return false;
    seen.add(chatJid);
    return true;
  });
  candidateRows.sort((a, b) => {
    if (Boolean(a.is_active) !== Boolean(b.is_active))
      return a.is_active ? -1 : 1;
    return String(a.chat_jid).localeCompare(String(b.chat_jid));
  });
  const rows = candidateRows.map((candidate) => String(candidate.chat_jid).trim());
  if (currentChatJid && !seen.has(currentChatJid)) {
    rows.unshift(currentChatJid);
  }
  return rows;
}
function candidateName(candidates, chatJid) {
  const c = candidates.find((x) => x.chat_jid === chatJid);
  if (!c)
    return chatJid.replace(/^[^:]+:/, "");
  const raw = typeof c.agent_name === "string" ? c.agent_name.trim() : "";
  return raw || chatJid.replace(/^[^:]+:/, "");
}
function resolveSwipeNeighbours(options) {
  const rows = resolveSwipeableChatAgents(options.candidates, options.currentChatJid);
  if (rows.length <= 1)
    return { prev: null, next: null };
  const idx = rows.indexOf(options.currentChatJid);
  if (idx < 0)
    return { prev: null, next: null };
  const prevJid = rows[(idx - 1 + rows.length) % rows.length];
  const nextJid = rows[(idx + 1) % rows.length];
  return {
    prev: { chatJid: prevJid, name: candidateName(options.candidates, prevJid) },
    next: { chatJid: nextJid, name: candidateName(options.candidates, nextJid) }
  };
}
function shouldTriggerTouchChatSwipe(options) {
  const minDistancePx = Number.isFinite(options.minDistancePx) ? Number(options.minDistancePx) : 72;
  const axisRatio = Number.isFinite(options.axisRatio) ? Number(options.axisRatio) : 1.35;
  return Math.abs(options.dx) >= minDistancePx && Math.abs(options.dx) > Math.abs(options.dy) * axisRatio;
}
function ensureIndicatorElement(_container) {
  let indicator = document.querySelector(".chat-swipe-indicator");
  if (!indicator) {
    indicator = document.createElement("div");
    indicator.className = "chat-swipe-indicator";
    indicator.innerHTML = '<span class="chat-swipe-chevron"></span>' + '<span class="chat-swipe-name"></span>';
    document.body.appendChild(indicator);
  }
  return indicator;
}
function showIndicator(indicator, dx, _containerWidth, neighbours) {
  const absDx = Math.abs(dx);
  const progress = Math.min(absDx / 100, 1);
  const willTrigger = absDx >= 72;
  indicator.style.display = "flex";
  indicator.style.opacity = String(Math.min(progress * 2.5, 1));
  indicator.classList.toggle("chat-swipe-indicator--ready", willTrigger);
  const isNext = dx < 0;
  const target = isNext ? neighbours.next : neighbours.prev;
  const chevron = indicator.querySelector(".chat-swipe-chevron");
  if (chevron) {
    chevron.textContent = isNext ? "›" : "‹";
    chevron.style.order = isNext ? "2" : "0";
  }
  const nameEl = indicator.querySelector(".chat-swipe-name");
  if (nameEl)
    nameEl.textContent = target?.name ?? "";
}
function hideIndicator(indicator) {
  indicator.style.display = "none";
  indicator.style.opacity = "0";
}
function attachChatSwipeNavigation(options) {
  const {
    timelineRef,
    activeChatAgents,
    currentChatJid,
    onSwitch,
    isIOSDevice: checkIOS,
    isLikelySafari
  } = options;
  const el = timelineRef.current;
  if (!el)
    return () => {};
  const isIOS = checkIOS();
  const isSafari = typeof isLikelySafari === "function" ? isLikelySafari() : false;
  if (!isIOS && !isSafari)
    return () => {};
  const state = createChatSwipeTouchState();
  const wheelState = createChatSwipeWheelState();
  let indicator = null;
  let neighbours = { prev: null, next: null };
  let lastPointerDownWasPen = false;
  function refreshNeighbours() {
    neighbours = resolveSwipeNeighbours({
      candidates: activeChatAgents,
      currentChatJid
    });
  }
  refreshNeighbours();
  function getIndicator() {
    if (!indicator) {
      indicator = ensureIndicatorElement(el);
    }
    return indicator;
  }
  function doSwitch(direction) {
    const target = direction === "next" ? neighbours.next : neighbours.prev;
    if (target)
      onSwitch(target.chatJid);
  }
  function onPointerDown(event) {
    lastPointerDownWasPen = String(event.pointerType || "").toLowerCase() === "pen";
  }
  function onTouchStart(event) {
    resetChatSwipeTouchState(state);
    refreshNeighbours();
    if (!isIOS)
      return;
    if (event.touches.length !== 1)
      return;
    if (lastPointerDownWasPen)
      return;
    if (!isEligibleChatSwipeTarget(event.target))
      return;
    const touch = event.touches[0];
    state.active = true;
    state.startX = touch.clientX;
    state.startY = touch.clientY;
    state.lastX = touch.clientX;
    state.lastY = touch.clientY;
    state.startedAt = Date.now();
  }
  function onTouchMove(event) {
    if (!state.active || state.cancelled)
      return;
    const touch = event.touches[0];
    if (!touch)
      return;
    state.lastX = touch.clientX;
    state.lastY = touch.clientY;
    const dx = state.lastX - state.startX;
    const dy = state.lastY - state.startY;
    if (!state.horizontalLocked) {
      if (Math.abs(dy) > 16 && Math.abs(dy) >= Math.abs(dx)) {
        state.cancelled = true;
        hideIndicator(getIndicator());
        return;
      }
      if (Math.abs(dx) > 12 && Math.abs(dx) > Math.abs(dy) * 1.15) {
        state.horizontalLocked = true;
      }
    }
    if (state.horizontalLocked) {
      if (event.cancelable)
        event.preventDefault();
      showIndicator(getIndicator(), dx, el.clientWidth, neighbours);
    }
  }
  function onTouchEnd() {
    if (!state.active)
      return;
    const dx = state.lastX - state.startX;
    const dy = state.lastY - state.startY;
    const shouldNavigate = !state.cancelled && shouldTriggerTouchChatSwipe({ dx, dy });
    hideIndicator(getIndicator());
    resetChatSwipeTouchState(state);
    if (shouldNavigate) {
      doSwitch(dx < 0 ? "next" : "prev");
    }
  }
  function onTouchCancel() {
    hideIndicator(getIndicator());
    resetChatSwipeTouchState(state);
  }
  function onWheel(event) {
    if (isIOS)
      return;
    if (!isSafari)
      return;
    if (!isEligibleChatSwipeTarget(event.target))
      return;
    const deltaX = event.deltaX;
    const deltaY = event.deltaY;
    if (!Number.isFinite(deltaX) || Math.abs(deltaX) < 72)
      return;
    if (Math.abs(deltaX) <= Math.abs(deltaY) * 1.35)
      return;
    if (event.cancelable)
      event.preventDefault();
    const now = Date.now();
    if (now - wheelState.lastTriggeredAt < 450)
      return;
    wheelState.lastTriggeredAt = now;
    doSwitch(deltaX > 0 ? "next" : "prev");
  }
  el.addEventListener("pointerdown", onPointerDown, { passive: true });
  el.addEventListener("touchstart", onTouchStart, { passive: true });
  el.addEventListener("touchmove", onTouchMove, { passive: false });
  el.addEventListener("touchend", onTouchEnd, { passive: true });
  el.addEventListener("touchcancel", onTouchCancel, { passive: true });
  el.addEventListener("wheel", onWheel, { passive: false });
  return () => {
    el.removeEventListener("pointerdown", onPointerDown);
    el.removeEventListener("touchstart", onTouchStart);
    el.removeEventListener("touchmove", onTouchMove);
    el.removeEventListener("touchend", onTouchEnd);
    el.removeEventListener("touchcancel", onTouchCancel);
    el.removeEventListener("wheel", onWheel);
    if (indicator) {
      hideIndicator(indicator);
      indicator.remove();
      indicator = null;
    }
  };
}

// web/src/panes/pane-host-transfer.ts
var PANE_HOST_TRANSFER_TTL_MS = 5 * 60 * 1000;

// web/src/ui/app-pane-runtime-orchestration.ts
function isLikelySafariBrowser(runtimeNavigator = typeof navigator !== "undefined" ? navigator : null) {
  if (!runtimeNavigator)
    return false;
  const userAgent = String(runtimeNavigator.userAgent || "");
  const vendor = String(runtimeNavigator.vendor || "");
  const isAppleWebKit = /AppleWebKit/i.test(userAgent);
  const isSafariToken = /Safari/i.test(userAgent);
  const isExcluded = /Chrome|Chromium|CriOS|EdgiOS|EdgA|Edg\//i.test(userAgent);
  const isFirefoxiOS = /FxiOS/i.test(userAgent);
  return isAppleWebKit && (vendor.includes("Apple") || isSafariToken) && !isExcluded && !isFirefoxiOS;
}

// web/src/gi-queue-return.ts
async function recoverQueueDraft(item, parsed, fetcher = fetch) {
  if (!item?.chat_jid?.startsWith("gi:") || !item?.id)
    throw new Error("Missing queue origin");
  const session = item.chat_jid.slice(3);
  const attachments = new Map;
  for (const media of item.metadata?.media || []) {
    if (media.session_id && media.session_id !== session)
      throw new Error("Attachment belongs to another session");
    const id = media.media_id || String(media.id || "").replace(/^media:/, "");
    if (id)
      attachments.set(String(id), { ...media, id });
  }
  for (const ref of parsed.attachmentRefs || []) {
    if (!attachments.has(String(ref.id)))
      attachments.set(String(ref.id), ref);
  }
  const media = [];
  for (const ref of attachments.values()) {
    if (!/^\d+$/.test(String(ref.id)))
      throw new Error("Invalid queued attachment identifier");
    const response = await fetcher(`/api/sessions/${encodeURIComponent(session)}/media/${ref.id}`);
    if (!response.ok)
      throw new Error(`Cannot restore queued attachment: HTTP ${response.status}`);
    const blob = await response.blob();
    if (blob.size > 10 * 1024 * 1024)
      throw new Error("Queued attachment exceeds 10 MiB limit");
    media.push(new File([blob], ref.filename || ref.label || `attachment-${ref.id}`, {
      type: ref.content_type || blob.type,
      lastModified: Date.parse(ref.created_at || "") || 0
    }));
  }
  return { ...emptyDraft(), text: parsed.text || "", fileRefs: parsed.fileRefs || [], messageRefs: parsed.messageRefs || [], media };
}

// web/src/gi-compaction-state.ts
function createActivityRevision() {
  let revision = 0;
  return { capture: () => revision, invalidate: () => ++revision, accepts: (value) => value === revision };
}
function compactionNotice(activity, now = Date.now()) {
  const c = activity?.compaction;
  if (!c || c.turn_id !== activity.turn_id)
    return null;
  if (c.active && ["running", "cancelling"].includes(activity.status))
    return {
      type: "intent",
      intent_key: "compaction",
      title: activity.status === "cancelling" ? "Cancelling compaction" : "Compacting context",
      started_at: c.timestamp,
      turn_id: activity.turn_id,
      started_seq: c.seq
    };
  const age = now - Date.parse(c.timestamp);
  if (!Number.isFinite(age) || age < 0 || age > 1e4)
    return null;
  if (c.event_type === "compaction.suppressed")
    return { type: "notice", title: "Compaction temporarily suppressed", detail: c.detail || "Before-compact hook suppressed this attempt", turn_id: activity.turn_id };
  if (c.event_type === "compaction.failed")
    return { type: "notice", title: "Compaction failed", detail: c.detail || "Context retained", turn_id: activity.turn_id };
  return null;
}
function compactionElapsed(notice, now = Date.now()) {
  const elapsed = Math.max(0, Math.floor((now - Date.parse(notice?.started_at)) / 1000));
  return Number.isFinite(elapsed) ? `${Math.floor(elapsed / 60)}:${String(elapsed % 60).padStart(2, "0")}` : "0:00";
}

// web/src/gi-refresh-guards.ts
function createActivationRefreshGate() {
  let selection = null, connected = false, claimed = false;
  const select = (key) => {
    if (key !== selection) {
      selection = key;
      connected = false;
      claimed = false;
    }
  };
  const claim = () => {
    if (!connected || claimed)
      return false;
    claimed = true;
    return true;
  };
  return {
    select,
    activate(key) {
      select(key);
      return claim();
    },
    status(key, status) {
      select(key);
      if (status !== "connected") {
        connected = false;
        claimed = false;
        return false;
      }
      connected = true;
      return claim();
    },
    ready(key) {
      return key === selection && connected;
    }
  };
}
function createTimelineRevision() {
  let generation = 0;
  return { begin: () => ++generation, invalidate: () => ++generation, accepts: (value) => value === generation };
}
function createAssetVersionGuard(initial) {
  let baseline = initial?.trim() || "";
  const seen = new Set;
  return { observe(value) {
    if (typeof value !== "string" || !value.trim())
      return false;
    const version = value.trim();
    if (!baseline) {
      baseline = version;
      return false;
    }
    if (version === baseline || seen.has(version))
      return false;
    seen.add(version);
    return true;
  } };
}
function loadedAssetVersion(doc) {
  const script = doc.querySelector('script[src*="/dist/app.bundle.js"]');
  const src = script?.getAttribute("src");
  if (!src)
    return null;
  try {
    return new URL(src, doc.baseURI).searchParams.get("v");
  } catch {
    return null;
  }
}

// web/src/gi-search-state.ts
function createSearchView() {
  let view = { active: false, query: "", scope: "current", generation: 0 };
  return {
    capture: () => ({ ...view }),
    isCurrent: (value) => value.generation === view.generation,
    enter() {
      view = { ...view, active: true, generation: view.generation + 1 };
      return { ...view };
    },
    close() {
      view = { active: false, query: "", scope: "current", generation: view.generation + 1 };
      return { ...view };
    },
    query(query) {
      view = { ...view, query: query.trim(), generation: view.generation + 1 };
      return { ...view };
    },
    scope(scope) {
      if (!["current", "root", "all"].includes(scope))
        return { ...view };
      view = { ...view, scope, generation: view.generation + 1 };
      return { ...view };
    }
  };
}

// web/src/gi-message-pages.ts
function mergeMessagePages(current, incoming) {
  const rows = new Map(current.map((p) => [String(p.id), p]));
  for (const post of incoming)
    rows.set(String(post.id), post);
  const compare = (a, b) => String(a) < String(b) ? -1 : String(a) > String(b) ? 1 : 0;
  return [...rows.values()].sort((a, b) => compare(a.timestamp, b.timestamp) || compare(a.id, b.id));
}
function newMessageWindow() {
  return { before: null, after: null, hasMore: false, loaded: false };
}
function layoutTop(node) {
  const transform = getComputedStyle(node).transform;
  let y = 0;
  if (transform && transform !== "none") {
    try {
      y = new DOMMatrixReadOnly(transform).m42;
    } catch {}
  }
  return node.getBoundingClientRect().top - y;
}
function captureTimelineAnchor(root, previous) {
  if (!root)
    return null;
  if (previous?.root === root && previous.id && document.getElementById(previous.id))
    return previous;
  const box = root.getBoundingClientRect();
  const node = [...root.querySelectorAll(".post[id]")].find((el) => {
    const r = el.getBoundingClientRect();
    return r.bottom > box.top && r.top < box.bottom;
  });
  return { root, id: node?.id || "", top: node ? layoutTop(node) : 0, scroll: root.scrollTop, height: root.clientHeight };
}
function restoreTimelineAnchor(anchor) {
  if (!anchor?.id || !anchor.root.isConnected)
    return;
  const node = anchor.root.querySelector(`[id="${CSS.escape(anchor.id)}"]`);
  if (node) {
    anchor.root.scrollTop += layoutTop(node) - anchor.top;
    anchor.scroll = anchor.root.scrollTop;
  }
}

// web/src/gi-workspace-visibility.ts
function bindWorkspaceVisibility(sidebar) {
  let desired = null;
  let action = null;
  let scheduled = false;
  let disposed = false;
  const sync = () => {
    scheduled = false;
    if (disposed || desired === null && action === null)
      return;
    const menuButton = sidebar.querySelector(".workspace-menu-button");
    if (!menuButton)
      return;
    const buttons = Array.from(sidebar.querySelectorAll(".workspace-menu-dropdown .workspace-menu-item"));
    const toggle = buttons.find((button) => ["Show hidden files", "Hide hidden files"].includes(button.textContent?.trim() || ""));
    if (!toggle) {
      if (menuButton.getAttribute("aria-expanded") !== "true")
        menuButton.click();
      return;
    }
    if (action) {
      const label = action === "refresh" ? "Refresh tree" : "Reindex workspace";
      action = null;
      const target = buttons.find((button) => button.textContent?.trim() === label);
      if (target && !target.disabled)
        target.click();
      else
        menuButton.click();
      if (desired !== null)
        schedule();
      return;
    }
    const current = toggle.textContent?.trim() === "Hide hidden files";
    const next = desired;
    desired = null;
    if (current !== next)
      toggle.click();
    else
      menuButton.click();
  };
  const schedule = () => {
    if (!scheduled && !disposed) {
      scheduled = true;
      queueMicrotask(sync);
    }
  };
  const observer = new MutationObserver(schedule);
  observer.observe(sidebar, { subtree: true, childList: true });
  const onToggle = (event) => {
    const value = event.detail?.showHidden;
    if (typeof value !== "boolean")
      return;
    desired = value;
    schedule();
  };
  const onAction = (event) => {
    const value = event.detail?.action;
    if (value !== "refresh" && value !== "reindex")
      return;
    action = value;
    schedule();
  };
  window.addEventListener("piclaw:toggle-hidden-files", onToggle);
  window.addEventListener("piclaw:workspace-action", onAction);
  return () => {
    disposed = true;
    observer.disconnect();
    window.removeEventListener("piclaw:toggle-hidden-files", onToggle);
    window.removeEventListener("piclaw:workspace-action", onAction);
  };
}

// web/src/app.ts
paneRegistry.register(workspacePreviewPaneExtension);
paneRegistry.register(workspaceMarkdownPreviewPaneExtension);
var SESSION_KEY = "gi_session_id";
var DEFAULT_AGENT_ID = "web";
function RunBoundQueueStack({ steerEnabled, ...props }) {
  const root = Q_(null);
  W_(() => {
    const pending = new Set(props.items.filter((item) => item.pending).map((item) => String(item.id)));
    root.current?.querySelectorAll(".compose-queue-stack-steer-btn").forEach((button) => {
      button.disabled = !steerEnabled || props.busy || pending.has(button.closest("[data-queue-id]")?.dataset.queueId);
    });
  });
  return fe`<div ref=${root} style="display:contents"><${QueuedFollowupStack} ...${props} /></div>`;
}
function useWorkspaceFolderReference(visible, sessionId, fileRefs, attach) {
  const latest = Q_({ fileRefs, attach });
  latest.current = { fileRefs, attach };
  const syncRef = Q_(null);
  W_(() => {
    if (!visible)
      return;
    const sidebar = document.querySelector(".workspace-sidebar");
    const actions = sidebar?.querySelector(".workspace-header-actions");
    if (!sidebar || !actions)
      return;
    const button = document.createElement("button");
    button.type = "button";
    button.className = "menu-action-btn";
    button.textContent = "+ folder";
    button.setAttribute("aria-label", "Reference selected folder");
    const selected = () => sidebar.querySelector('.workspace-row.selected[data-type="dir"]')?.dataset.path || "";
    const sync = () => {
      const path = selected();
      button.hidden = !path;
      button.disabled = !path || latest.current.fileRefs.includes(path);
      button.title = path ? `Reference folder: ${path}` : "Reference selected folder";
    };
    const click = () => {
      const path = selected();
      if (path && !latest.current.fileRefs.includes(path))
        latest.current.attach(path);
    };
    button.addEventListener("click", click);
    actions.prepend(button);
    sync();
    syncRef.current = sync;
    const observer = new MutationObserver((records) => {
      if (records.some((record) => !button.contains(record.target)))
        sync();
    });
    observer.observe(sidebar, { subtree: true, childList: true, attributes: true, attributeFilter: ["class", "data-path", "data-type"] });
    return () => {
      observer.disconnect();
      button.removeEventListener("click", click);
      button.remove();
      syncRef.current = null;
    };
  }, [visible, sessionId]);
  W_(() => {
    syncRef.current?.();
  }, [fileRefs]);
}
function useContextTooltip(root, usage, notice, now, canStop, stop, compact) {
  W_(() => {
    const compose = root.current?.querySelector(".compose-box");
    if (!compose)
      return;
    const sync = () => {
      compose.querySelectorAll(".send-btn.abort-mode").forEach((button) => {
        if (button.disabled === canStop)
          button.disabled = !canStop;
      });
      compose.querySelectorAll(".compose-context-pie").forEach((button) => {
        const active = notice?.intent_key === "compaction";
        const normal = contextPresentation(usage, typeof compact === "function");
        const canCompact = typeof compact === "function" && !active;
        if (button.disabled === canCompact)
          button.disabled = !canCompact;
        const title = active ? `${notice.title} — ${compactionElapsed(notice, now)}` : normal.title + (canCompact ? "" : " — Context usage");
        const label = active ? `${notice.title} — ${normal.label}` : normal.label;
        if (button.getAttribute("title") !== title)
          button.setAttribute("title", title);
        if (button.getAttribute("aria-label") !== label)
          button.setAttribute("aria-label", label);
        if (button.getAttribute("data-tooltip") !== title)
          button.setAttribute("data-tooltip", title);
        if (button.classList.contains("is-compacting") !== active)
          button.classList.toggle("is-compacting", active);
        let elapsed = compose.querySelector(".gi-compaction-elapsed");
        if (active) {
          if (!elapsed) {
            elapsed = document.createElement("span");
            elapsed.className = "gi-compaction-elapsed";
            button.after(elapsed);
          }
          const text = compactionElapsed(notice, now);
          if (elapsed.textContent !== text)
            elapsed.textContent = text;
        } else
          elapsed?.remove();
      });
    };
    sync();
    const observer = new MutationObserver(sync);
    observer.observe(compose, { subtree: true, childList: true, attributes: true, attributeFilter: ["title", "disabled", "class"] });
    const onClick = (event) => {
      if (event.target.closest?.(".compose-context-pie")) {
        event.preventDefault();
        event.stopImmediatePropagation();
        compact?.();
        return;
      }
      if (event.target.closest?.(".send-btn.abort-mode")) {
        event.preventDefault();
        event.stopImmediatePropagation();
        stop();
      }
    };
    compose.addEventListener("click", onClick, true);
    return () => {
      observer.disconnect();
      compose.removeEventListener("click", onClick, true);
    };
  });
}
function sessionToChatJid2(id) {
  return `gi:${id}`;
}
async function ensureDefaultSession() {
  const linked = new URLSearchParams(location.search).get("chat_jid");
  if (linked?.startsWith("gi:")) {
    const id = linked.slice(3);
    const response = await fetch(`/api/sessions/${encodeURIComponent(id)}`);
    if (response.ok) {
      setLocalStorageItem(SESSION_KEY, id);
      return id;
    }
  }
  const stored = getLocalStorageItem(SESSION_KEY);
  if (stored) {
    try {
      const r = await fetch(`/api/sessions/${encodeURIComponent(stored)}`);
      if (r.ok)
        return stored;
    } catch {}
  }
  try {
    const existing = await fetch("/api/sessions");
    if (existing.ok) {
      const payload = await existing.json();
      const sessions = Array.isArray(payload?.sessions) ? payload.sessions : [];
      const matching = sessions.filter((session) => session?.scope?.agent_id === "web" && !session?.parent_session_id).sort((a, b) => String(b?.updated_at || "").localeCompare(String(a?.updated_at || "")));
      if (matching[0]?.id) {
        setLocalStorageItem(SESSION_KEY, matching[0].id);
        return matching[0].id;
      }
    }
  } catch {}
  const r = await fetch("/api/sessions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title: "@web", agent_id: "web" })
  });
  if (!r.ok)
    throw new Error("Failed to create default session");
  const s = await r.json();
  setLocalStorageItem(SESSION_KEY, s.id);
  return s.id;
}
async function getRuntimeConfig() {
  const r = await fetch("/api/runtime/config");
  if (!r.ok)
    return {};
  return r.json();
}
function GiApp() {
  const containerRef = Q_(null);
  const [ready, setReady] = F_(false);
  const [sessionId, setSessionId] = F_(null);
  const selection = Q_(createSelectionScope()).current;
  const [sessionError, setSessionError] = F_(null);
  const [draftStorageError, setDraftStorageError] = F_("");
  const [draftRestore, setDraftRestore] = F_(null);
  const [composePrefill, setComposePrefill] = F_(null);
  const draftsRef = Q_(null);
  if (!draftsRef.current)
    draftsRef.current = createDraftRepository(indexedDraftStorage(), (error) => setDraftStorageError(`Draft not saved: ${error.message}`));
  const drafts = draftsRef.current;
  const getDraft = (sid) => drafts.get(sid);
  const [runtimeConfig, setRuntimeConfig] = F_({});
  const [agents, setAgents] = F_({});
  const [userProfile, setUserProfile] = F_(null);
  const [workspaceOpen, setWorkspaceOpen] = F_(false);
  const [tabs, setTabs] = F_([]);
  const [activeTabId, setActiveTabId] = F_(null);
  const tabFocusEpoch = Q_(0);
  const editorOpen = tabs.length > 0;
  const [posts, setPosts] = F_([]);
  const [hasMore, setHasMore] = F_(false);
  const deletions = Q_(createMessageDeletionState()).current;
  const deletingAnimation = Q_(new Set);
  const [removingPostIds, setRemovingPostIds] = F_(new Set);
  const [deleteError, setDeleteError] = F_("");
  const messageWindow = Q_(newMessageWindow());
  const pageRequest = Q_(null);
  const pageRefreshPending = Q_(false);
  const scrollRestore = Q_(null);
  const readingAnchor = Q_(null);
  const searchView = Q_(createSearchView()).current;
  const [searchState, setSearchState] = F_(searchView.capture());
  const [searchError, setSearchError] = F_("");
  const timelineRevision = Q_(createTimelineRevision()).current;
  const versionGuard = Q_(createAssetVersionGuard(loadedAssetVersion(document))).current;
  const [newUIVersion, setNewUIVersion] = F_("");
  const timelineRef = Q_(null);
  const [fileRefs, setFileRefs] = F_([]);
  const [messageRefs, setMessageRefs] = F_([]);
  const [followupQueueItems, setFollowupQueueItems] = F_([]);
  const [queueError, setQueueError] = F_("");
  const [queueBusy, setQueueBusy] = F_(false);
  const queueMutation = Q_(null);
  const queueRevision = Q_(0);
  const [queueActiveTurnId, setQueueActiveTurnId] = F_(null);
  const modelRevision = Q_(0);
  const modelMutation = Q_(null);
  const connectionRevision = Q_(0);
  const streamDisconnected = Q_(true);
  const activationRefresh = Q_(createActivationRefreshGate()).current;
  const refreshAfterConnection = Q_(() => {});
  const refreshTimer = Q_(null);
  const [optimisticQueue, setOptimisticQueue] = F_([]);
  const [floatingWidget, setFloatingWidget] = F_(null);
  const [attachmentPreview, setAttachmentPreview] = F_(null);
  const [contextUsage, setContextUsage] = F_(null);
  const [activity, setActivity] = F_(null);
  const activityRevision = Q_(createActivityRevision()).current;
  const [activityNow, setActivityNow] = F_(Date.now());
  const [stopPending, setStopPending] = F_(false);
  const [stopError, setStopError] = F_("");
  const [activityFresh, setActivityFresh] = F_(false);
  const stopToken = Q_(null);
  const [compactState, setCompactState] = F_(null);
  const [compactPending, setCompactPending] = F_(false);
  const [compactError, setCompactError] = F_("");
  const compactToken = Q_(null);
  const manualCompact = activityFresh && activity?.status === "idle" && compactState?.available && !compactPending ? async () => {
    if (compactToken.current || streamDisconnected.current)
      return;
    const scope = selection.capture();
    const expected = compactState.token;
    const token = {};
    compactToken.current = token;
    setCompactPending(true);
    setCompactError("");
    try {
      await compactSession(sessionToChatJid2(scope.sessionId), expected);
    } catch (error) {
      if (selection.isCurrent(scope))
        setCompactError(`Compact failed: ${error.message}`);
    } finally {
      if (compactToken.current === token) {
        compactToken.current = null;
        setCompactPending(false);
      }
      if (selection.isCurrent(scope)) {
        activityRevision.invalidate();
        setActivityFresh(false);
        refreshAfterConnection.current();
      }
    }
  } : null;
  const notice = compactionNotice(activity, activityNow);
  K_(() => {
    if (!activity?.compaction)
      return;
    const timer = setInterval(() => setActivityNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [activity]);
  useContextTooltip(containerRef, contextUsage, notice, activityNow, activityFresh && !stopPending && !!activity?.turn_id && ["running", "cancelling"].includes(activity?.status), async () => {
    if (stopToken.current || !activityFresh || !activity?.turn_id || streamDisconnected.current)
      return;
    const scope = selection.capture();
    const run = activity.turn_id;
    const token = {};
    stopToken.current = token;
    setStopPending(true);
    setStopError("");
    try {
      await cancelSessionRun(sessionToChatJid2(scope.sessionId), run);
    } catch (error) {
      if (selection.isCurrent(scope) && error.status !== 409)
        setStopError(`Stop failed: ${error.message}`);
    } finally {
      if (stopToken.current === token) {
        stopToken.current = null;
        setStopPending(false);
      }
      if (selection.isCurrent(scope)) {
        activityRevision.invalidate();
        setActivityFresh(false);
        refreshAfterConnection.current();
      }
    }
  }, manualCompact);
  const [activeChatAgents, setActiveChatAgents] = F_([]);
  const sessionListRevision = Q_(0);
  const [currentChatBranches, setCurrentChatBranches] = F_([]);
  const [activeModel, setActiveModel] = F_("");
  const [agentModelsPayload, setAgentModelsPayload] = F_(null);
  const [activeThinkingLevel, setActiveThinkingLevel] = F_("");
  const [supportsThinking, setSupportsThinking] = F_(false);
  const [modelUsage, setModelUsage] = F_(null);
  const [connectionStatus, setConnectionStatus] = F_("connected");
  const [isAgentTurnActive, setIsAgentTurnActive] = F_(false);
  const isAgentRunningRef = Q_(false);
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
  const currentChatJid = u_(() => sessionId ? sessionToChatJid2(sessionId) : "", [sessionId]);
  const renderedSelection = selection.capture();
  K_(() => {
    const cleanupTheme = initTheme();
    const cleanupAppearance = initGiAppearance();
    const cleanupDisplayScale = installGiDisplayScale();
    if (getLocalStorageItem("piclaw_system_meters_enabled") === null) {
      setLocalStorageItem("piclaw_system_meters_enabled", "true");
    }
    Promise.all([
      ensureDefaultSession(),
      getRuntimeConfig(),
      drafts.load().catch((error) => setDraftStorageError(`Draft recovery unavailable: ${error.message}`))
    ]).then(([sid, cfg]) => {
      selection.select(sid);
      setSessionId(sid);
      setFileRefs(getDraft(sid).fileRefs);
      setMessageRefs(getDraft(sid).messageRefs);
      setRuntimeConfig(cfg);
      setUserProfile({ name: cfg.user_name, avatarUrl: cfg.user_avatar, avatarBackground: cfg.user_avatar_background });
      setAgents({
        [DEFAULT_AGENT_ID]: {
          id: DEFAULT_AGENT_ID,
          name: cfg.assistant_name || "Gi",
          avatar_url: cfg.assistant_avatar || null
        }
      });
      setActiveModel(cfg.current || cfg.default_model || "");
      setActiveThinkingLevel(cfg.default_thinking_level || "");
      setSupportsThinking(Boolean(cfg.supports_thinking));
      setAgentModelsPayload(cfg);
      setReady(true);
    }).catch((err) => {
      console.error("[gi] Bootstrap failed:", err);
    });
    return () => {
      cleanupTheme?.();
      cleanupAppearance();
      cleanupDisplayScale();
    };
  }, []);
  W_(() => {
    const pending = scrollRestore.current;
    scrollRestore.current = null;
    if (!pending || !selection.isCurrent(pending.scope) || !searchView.isCurrent(pending.view) || pending.connection !== connectionRevision.current)
      return;
    if (pending.bottom && timelineRef.current) {
      timelineRef.current.scrollTop = 0;
      readingAnchor.current = null;
    } else {
      restoreTimelineAnchor(pending.anchor);
      readingAnchor.current = pending.anchor;
    }
  }, [posts]);
  const loadPosts = Y_(async (opts = {}) => {
    if (!sessionId || !activationRefresh.ready(selection.capture().generation))
      return;
    const scope = selection.capture(), view = searchView.capture(), connection = connectionRevision.current;
    if (scope.sessionId !== sessionId || view.active)
      return;
    const older = opts.older === true;
    if (pageRequest.current?.connection === connection && pageRequest.current?.generation === scope.generation) {
      if (!older)
        pageRefreshPending.current = true;
      return pageRequest.current.promise;
    }
    if (older && (!messageWindow.current.loaded || !messageWindow.current.hasMore))
      return;
    const token = { connection, generation: scope.generation };
    pageRequest.current = token;
    const request = timelineRevision.begin();
    const valid = () => selection.isCurrent(scope) && searchView.isCurrent(view) && connection === connectionRevision.current && timelineRevision.accepts(request) && !streamDisconnected.current;
    token.promise = (async () => {
      try {
        const initial = !messageWindow.current.loaded || !messageWindow.current.after;
        let cursor = older ? messageWindow.current.before : messageWindow.current.after;
        do {
          const data = await getTimeline(50, older ? cursor : null, sessionToChatJid2(scope.sessionId), !older && !initial ? cursor : null);
          if (!valid())
            return;
          const root = timelineRef.current;
          scrollRestore.current = { scope, view, connection, anchor: captureTimelineAnchor(root, readingAnchor.current), bottom: !older && (initial || !root || Math.abs(root.scrollTop) < 80) };
          const incoming = data.posts || [];
          setPosts((prev) => deletions.filter(mergeMessagePages(prev, incoming), deletingAnimation.current));
          if (initial) {
            messageWindow.current = { loaded: true, before: data.before, after: data.after, hasMore: data.hasMore };
          } else if (older) {
            messageWindow.current.before = data.before || cursor;
            messageWindow.current.hasMore = data.hasMore;
          } else {
            messageWindow.current.after = data.after || cursor;
          }
          setHasMore(messageWindow.current.hasMore);
          if (older || initial || !data.hasMore || !data.after || data.after === cursor)
            break;
          cursor = data.after;
        } while (valid());
      } catch (error) {
        if (valid())
          setSessionError(`Timeline refresh failed: ${error.message}`);
      } finally {
        if (pageRequest.current === token) {
          pageRequest.current = null;
          if (pageRefreshPending.current) {
            pageRefreshPending.current = false;
            refreshAfterConnection.current();
          }
        }
      }
    })();
    return token.promise;
  }, [sessionId]);
  K_(() => {
    const root = timelineRef.current;
    if (!root || searchState.active)
      return;
    const onScroll = () => {
      const distance = root.scrollHeight - root.clientHeight + root.scrollTop;
      if (messageWindow.current.hasMore && distance < 200)
        loadPosts({ older: true });
    };
    const userScroll = () => {
      readingAnchor.current = null;
    };
    root.addEventListener("scroll", onScroll, { passive: true });
    root.addEventListener("wheel", userScroll, { passive: true });
    root.addEventListener("touchstart", userScroll, { passive: true });
    root.addEventListener("pointerdown", userScroll);
    root.addEventListener("keydown", userScroll);
    window.addEventListener("resize", userScroll);
    return () => {
      root.removeEventListener("scroll", onScroll);
      root.removeEventListener("wheel", userScroll);
      root.removeEventListener("touchstart", userScroll);
      root.removeEventListener("pointerdown", userScroll);
      root.removeEventListener("keydown", userScroll);
      window.removeEventListener("resize", userScroll);
    };
  }, [posts, searchState.active, loadPosts]);
  const runSearch = async (query, scopeValue) => {
    if (query !== undefined)
      setSearchState(searchView.query(query));
    if (scopeValue !== undefined)
      setSearchState(searchView.scope(scopeValue));
    const view = searchView.capture(), owner = selection.capture();
    if (!view.active || !owner.sessionId || !activationRefresh.ready(owner.generation))
      return;
    const request = timelineRevision.begin(), connection = connectionRevision.current;
    setSearchError("");
    if (!view.query) {
      setPosts([]);
      setHasMore(false);
      return;
    }
    try {
      const result = await searchPosts(view.query, 50, 0, sessionToChatJid2(owner.sessionId), view.scope);
      if (!selection.isCurrent(owner) || !searchView.isCurrent(view) || !timelineRevision.accepts(request) || connection !== connectionRevision.current || streamDisconnected.current)
        return;
      setPosts(deletions.filter(dedupePosts(result.posts || []), deletingAnimation.current));
      setHasMore(false);
    } catch (error) {
      if (selection.isCurrent(owner) && searchView.isCurrent(view) && timelineRevision.accepts(request) && connection === connectionRevision.current && !streamDisconnected.current)
        setSearchError(`Search failed: ${error.message}`);
    }
  };
  const enterSearch = () => {
    timelineRevision.invalidate();
    readingAnchor.current = null;
    pageRequest.current = null;
    pageRefreshPending.current = false;
    scrollRestore.current = null;
    setSearchState(searchView.enter());
    setPosts([]);
    setHasMore(false);
    setSearchError("");
  };
  const exitSearch = () => {
    timelineRevision.invalidate();
    setPosts([]);
    messageWindow.current = newMessageWindow();
    pageRequest.current = null;
    pageRefreshPending.current = false;
    setSearchState(searchView.close());
    setSearchError("");
    loadPosts();
  };
  const handleDeletePost = async (post) => {
    const id = post?.id, owner = selection.capture(), view = searchView.capture();
    const destination = post?.chat_jid;
    if (typeof id !== "string" || !destination?.startsWith("gi:") || !deletions.begin(id))
      return;
    setDeleteError("");
    try {
      await deletePost(id, false, destination);
      deletions.finish(id, true);
      const current = () => selection.isCurrent(owner) && searchView.isCurrent(view);
      if (!current())
        return;
      timelineRevision.invalidate();
      pageRequest.current = null;
      pageRefreshPending.current = false;
      deletingAnimation.current.add(id);
      setRemovingPostIds(new Set(deletingAnimation.current));
      await new Promise((resolve) => setTimeout(resolve, 220));
      deletingAnimation.current.delete(id);
      if (!current())
        return;
      const root = timelineRef.current;
      scrollRestore.current = { scope: owner, view, connection: connectionRevision.current, anchor: captureTimelineAnchor(root, readingAnchor.current), bottom: false };
      setRemovingPostIds(new Set(deletingAnimation.current));
      setPosts((prev) => deletions.filter(prev, deletingAnimation.current));
      refreshSelectedState();
    } catch (error) {
      deletions.finish(id, false);
      if (selection.isCurrent(owner) && searchView.isCurrent(view))
        setDeleteError(`Delete failed: ${error.message}`);
    }
  };
  const scrollToBottom = Y_(() => {
    const el = timelineRef.current;
    if (!el)
      return;
    if (Math.abs(el.scrollTop) < 80)
      el.scrollTop = 0;
  }, []);
  const refreshSessionLists = Y_(async (sid) => {
    if (!sid) {
      setActiveChatAgents([]);
      setCurrentChatBranches([]);
      return;
    }
    const scope = selection.capture();
    if (scope.sessionId !== sid)
      return;
    const chatJid = sessionToChatJid2(sid);
    const revision = ++sessionListRevision.current;
    const [agentsPayload, branchesPayload] = await Promise.all([
      getActiveChatAgents().catch(() => ({ agents: [] })),
      getChatBranches(chatJid).catch(() => ({ branches: [] }))
    ]);
    if (!selection.isCurrent(scope) || revision !== sessionListRevision.current)
      return;
    const agentsList = Array.isArray(agentsPayload?.agents) ? agentsPayload.agents : [];
    setAgents(Object.fromEntries(agentsList.map((entry) => [entry.agent_id, {
      id: entry.agent_id,
      name: entry.agent_name,
      avatar_url: null
    }])));
    const branchesList = Array.isArray(branchesPayload?.branches) ? branchesPayload.branches : Array.isArray(branchesPayload?.chats) ? branchesPayload.chats : [];
    setActiveChatAgents(agentsList);
    setCurrentChatBranches(branchesList);
  }, []);
  const handleSseEvent = Y_((eventType, data) => {
    if (!selection.current() || data?.chat_jid !== sessionToChatJid2(selection.current()))
      return;
    if (eventType === "connected" && versionGuard.observe(data?.app_asset_version))
      setNewUIVersion(data.app_asset_version);
    const staleTerminal = staleTerminalEvent(eventType, data, currentTurnIdRef.current);
    if (!staleTerminal && (eventType === "agent_status" || eventType.startsWith("compaction_") || ["queue_changed", "agent_response"].includes(eventType))) {
      activityRevision.invalidate();
      setActivityFresh(false);
    }
    if (eventType.startsWith("compaction_") || ["new_post", "agent_status", "agent_response", "queue_changed", "agent_followup_queued", "agent_followup_consumed", "agent_followup_removed"].includes(eventType)) {
      ++queueRevision.current;
      if (!refreshTimer.current)
        refreshTimer.current = setTimeout(() => {
          refreshTimer.current = null;
          refreshAfterConnection.current();
        }, 0);
    }
    if (eventType === "new_post" || eventType === "agent_response") {
      if (data?.id && data?.data && !searchView.capture().active) {
        const root = timelineRef.current;
        scrollRestore.current = { scope: selection.capture(), view: searchView.capture(), connection: connectionRevision.current, anchor: captureTimelineAnchor(root, readingAnchor.current), bottom: !root || Math.abs(root.scrollTop) < 80 };
        setPosts((prev) => mergeMessagePages(prev, [data]));
        scrollToBottom();
      }
    }
    if (staleTerminal)
      return;
    if (eventType === "agent_status") {
      setAgentStatus(data);
      const active = data?.status === "running" || data?.status === "cancelling";
      if (active && data.turn_id && currentTurnIdRef.current !== data.turn_id) {
        currentTurnIdRef.current = data.turn_id;
        setCurrentTurnId(data.turn_id);
        draftBufferRef.current = "";
        thoughtBufferRef.current = "";
        setAgentDraft(null);
        setAgentThought(null);
      }
      setIsAgentTurnActive(active);
      isAgentRunningRef.current = active;
    }
    if (eventType === "agent_draft_delta" || eventType === "agent_thought_delta") {
      if (data.turn_id && currentTurnIdRef.current && data.turn_id !== currentTurnIdRef.current)
        return;
      if (data.turn_id && !currentTurnIdRef.current) {
        currentTurnIdRef.current = data.turn_id;
        setCurrentTurnId(data.turn_id);
      }
    }
    if (eventType === "agent_draft_delta") {
      const delta = data?.delta || "";
      if (delta) {
        draftBufferRef.current = (draftBufferRef.current || "") + delta;
        setAgentDraft(draftBufferRef.current);
      }
    }
    if (eventType === "agent_thought_delta") {
      const delta = data?.delta || "";
      if (delta) {
        thoughtBufferRef.current = (thoughtBufferRef.current || "") + delta;
        setAgentThought(thoughtBufferRef.current);
      }
    }
    if (eventType === "agent_response") {
      currentTurnIdRef.current = null;
      setCurrentTurnId(null);
      draftBufferRef.current = "";
      thoughtBufferRef.current = "";
      setAgentDraft(null);
      setAgentThought(null);
    }
  }, [scrollToBottom]);
  const handleConnectionStatusChange = Y_((status) => {
    const shouldRefresh = activationRefresh.status(selection.capture().generation, status);
    ++connectionRevision.current;
    timelineRevision.invalidate();
    pageRequest.current = null;
    pageRefreshPending.current = false;
    scrollRestore.current = null;
    activityRevision.invalidate();
    setActivity(null);
    setActivityFresh(false);
    ++queueRevision.current;
    setQueueActiveTurnId(null);
    setConnectionStatus(status);
    streamDisconnected.current = status !== "connected";
    if (status !== "connected") {
      setAgentStatus(null);
      setAgentDraft(null);
      setAgentPlan(null);
      setAgentThought(null);
      setPendingRequest(null);
      setCurrentTurnId(null);
      setSteerQueuedTurnId(null);
      draftBufferRef.current = "";
      thoughtBufferRef.current = "";
      pendingRequestRef.current = null;
      currentTurnIdRef.current = null;
      steerQueuedTurnIdRef.current = null;
      setIsAgentTurnActive(false);
      isAgentRunningRef.current = false;
    } else if (shouldRefresh)
      refreshAfterConnection.current();
  }, []);
  useSseConnection({
    handleSseEvent: (type, data) => {
      if (selection.isCurrent(renderedSelection))
        handleSseEvent(type, data);
    },
    handleConnectionStatusChange: (status) => {
      if (selection.isCurrent(renderedSelection))
        handleConnectionStatusChange(status);
    },
    loadPosts,
    onWake: () => {
      if (selection.isCurrent(renderedSelection))
        refreshAfterConnection.current();
    },
    chatJid: currentChatJid,
    selectionKey: renderedSelection.generation
  });
  const refreshSelectedState = Y_(async () => {
    const scope = selection.capture();
    if (!sessionId || scope.sessionId !== sessionId || !activationRefresh.ready(scope.generation))
      return;
    const chat = sessionToChatJid2(sessionId);
    const revision = ++queueRevision.current;
    const connection = connectionRevision.current;
    const modelVersion = modelRevision.current;
    const activityVersion = activityRevision.capture();
    try {
      const [models, queue, status, compact] = await Promise.all([
        getAgentModels(chat),
        getAgentQueueState(chat),
        getAgentStatus("", chat),
        getSessionCompaction(chat)
      ]);
      if (!selection.isCurrent(scope) || connection !== connectionRevision.current || streamDisconnected.current)
        return;
      if (!activityRevision.accepts(activityVersion))
        return;
      setActivity(status);
      setCompactState(compact);
      setActivityFresh(true);
      setActivityNow(Date.now());
      if (modelVersion === modelRevision.current && !modelMutation.current) {
        setAgentModelsPayload(models);
        setActiveModel(models.current);
        setActiveThinkingLevel(models.thinking_level);
        setSupportsThinking(models.supports_thinking);
        setContextUsage(models.context_usage || null);
      }
      if (revision === queueRevision.current && !queueMutation.current) {
        setQueueActiveTurnId(queue.activeTurnId || null);
        setFollowupQueueItems(queue.items || []);
        const admitted = new Set((queue.items || []).map((item) => item.metadata?.client_request_id).filter(Boolean));
        setOptimisticQueue((items) => items.filter((item) => !admitted.has(item.id)));
      }
      setAgentStatus(status);
      const running = status?.status === "running" || status?.status === "cancelling";
      if (running && status.turn_id && currentTurnIdRef.current !== status.turn_id) {
        currentTurnIdRef.current = status.turn_id;
        setCurrentTurnId(status.turn_id);
        draftBufferRef.current = "";
        thoughtBufferRef.current = "";
        setAgentDraft(null);
        setAgentThought(null);
      }
      setIsAgentTurnActive(running);
      isAgentRunningRef.current = running;
      setSessionError(null);
    } catch (error) {
      if (selection.isCurrent(scope) && connection === connectionRevision.current && activityRevision.accepts(activityVersion) && !streamDisconnected.current)
        setSessionError(error.message || "Unable to refresh session");
    }
  }, [sessionId]);
  refreshAfterConnection.current = () => {
    if (!activationRefresh.ready(selection.capture().generation))
      return;
    if (searchView.capture().active)
      runSearch();
    else
      loadPosts();
    refreshSelectedState();
  };
  K_(() => () => {
    if (refreshTimer.current)
      clearTimeout(refreshTimer.current);
  }, []);
  K_(() => {
    if (!ready || !sessionId)
      return;
    if (activationRefresh.activate(selection.capture().generation))
      refreshAfterConnection.current();
    refreshSessionLists(sessionId);
    const id = setInterval(() => {
      refreshAfterConnection.current();
      refreshSessionLists(sessionId);
    }, 1e4);
    return () => clearInterval(id);
  }, [ready, sessionId, loadPosts, refreshSessionLists, refreshSelectedState]);
  const handlePost = Y_((_response) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    refreshAfterConnection.current();
    refreshSessionLists(sessionId);
  }, [refreshSessionLists, sessionId]);
  const handleSwitchChat = Y_((chatJid) => {
    const nextSessionId = typeof chatJid === "string" && chatJid.startsWith("gi:") ? chatJid.slice(3) : null;
    if (!nextSessionId || nextSessionId === sessionId)
      return;
    if (sessionId)
      drafts.update(sessionId, { fileRefs, messageRefs });
    selection.select(nextSessionId);
    resetQuickActionsReadiness();
    deletingAnimation.current.clear();
    setRemovingPostIds(new Set);
    setDeleteError("");
    setComposePrefill(null);
    const linkedURL = new URL(location.href);
    if (linkedURL.searchParams.has("chat_jid")) {
      linkedURL.searchParams.set("chat_jid", sessionToChatJid2(nextSessionId));
      history.replaceState(null, "", linkedURL);
    }
    activationRefresh.select(selection.capture().generation);
    streamDisconnected.current = true;
    setConnectionStatus("disconnected");
    setSearchState(searchView.close());
    setSearchError("");
    messageWindow.current = newMessageWindow();
    readingAnchor.current = null;
    pageRequest.current = null;
    pageRefreshPending.current = false;
    scrollRestore.current = null;
    timelineRevision.invalidate();
    stopToken.current = null;
    setStopPending(false);
    setStopError("");
    compactToken.current = null;
    setCompactPending(false);
    setCompactError("");
    setCompactState(null);
    activityRevision.invalidate();
    setActivity(null);
    setActivityFresh(false);
    setLocalStorageItem(SESSION_KEY, nextSessionId);
    setSessionId(nextSessionId);
    setPosts([]);
    setHasMore(false);
    setFollowupQueueItems([]);
    setQueueActiveTurnId(null);
    setCurrentChatBranches([]);
    queueMutation.current = null;
    ++queueRevision.current;
    setQueueBusy(false);
    setQueueError("");
    setOptimisticQueue([]);
    ++modelRevision.current;
    modelMutation.current = null;
    setFileRefs(getDraft(nextSessionId).fileRefs);
    setMessageRefs(getDraft(nextSessionId).messageRefs);
    setAgentStatus(null);
    setAgentDraft(null);
    setAgentThought(null);
    setAgentPlan(null);
    setPendingRequest(null);
    setCurrentTurnId(null);
    setSteerQueuedTurnId(null);
    draftBufferRef.current = "";
    thoughtBufferRef.current = "";
    currentTurnIdRef.current = null;
    steerQueuedTurnIdRef.current = null;
    setIsAgentTurnActive(false);
    isAgentRunningRef.current = false;
    setActiveModel("");
    setActiveThinkingLevel("");
    setSupportsThinking(false);
    setAgentModelsPayload(null);
    setContextUsage(null);
    setModelUsage(null);
    setSessionError(null);
  }, [sessionId, fileRefs, messageRefs]);
  W_(() => {
    if (!ready || !currentChatJid || !timelineRef.current)
      return;
    const timeline = timelineRef.current;
    const surface = containerRef.current;
    if (!surface || timeline.parentElement !== surface)
      return;
    const preserveSelection = (event) => {
      const target = event.target instanceof Element ? event.target : null;
      const eligibleSurface = target && (timeline.contains(target) || target.closest(".agent-status-panel")?.parentElement === surface);
      if (!eligibleSurface || window.getSelection()?.toString())
        event.stopImmediatePropagation();
    };
    surface.addEventListener("pointerdown", preserveSelection);
    surface.addEventListener("touchstart", preserveSelection);
    surface.addEventListener("wheel", preserveSelection);
    const detach = attachChatSwipeNavigation({
      timelineRef: { current: surface },
      activeChatAgents,
      currentChatJid,
      onSwitch: handleSwitchChat,
      isIOSDevice,
      isLikelySafari: isLikelySafariBrowser
    });
    return () => {
      detach();
      surface.removeEventListener("pointerdown", preserveSelection);
      surface.removeEventListener("touchstart", preserveSelection);
      surface.removeEventListener("wheel", preserveSelection);
    };
  }, [ready, currentChatJid, activeChatAgents, handleSwitchChat, posts.length === 0]);
  const handleCreateSession = Y_(async () => {
    if (!sessionId)
      return;
    const scope = selection.capture();
    try {
      const created = await forkChatBranch(sessionToChatJid2(sessionId));
      if (!selection.isCurrent(scope))
        return;
      if (!created?.branch?.chat_jid)
        throw new Error("Missing created chat identifier");
      handleSwitchChat(created.branch.chat_jid);
    } catch (error) {
      if (selection.isCurrent(scope))
        setSessionError(error.message || "Failed to create session");
    }
  }, [sessionId, handleSwitchChat]);
  const handleSessionMutation = async (chatJid, action, value) => {
    ++sessionListRevision.current;
    if (action === "rename")
      await renameChatBranch(chatJid, { title: value });
    else if (action === "pin")
      await pinChatSession(chatJid, value);
    else if (action === "archive")
      await pruneChatBranch(chatJid);
    else if (action === "restore")
      await restoreChatBranch(chatJid);
    else
      throw new Error("Unsupported session action");
    const revision = ++sessionListRevision.current;
    const data = await getActiveChatAgents();
    if (revision === sessionListRevision.current)
      setActiveChatAgents(data.agents || []);
  };
  const mutateQueue = async (action, itemOrIndex, toIndex) => {
    if (queueMutation.current)
      return;
    if (action === "steer" && (streamDisconnected.current || !isAgentTurnActive || !queueActiveTurnId || itemOrIndex.pending))
      return;
    const expectedActiveTurnId = queueActiveTurnId;
    const scope = selection.capture();
    if (!scope.sessionId)
      return;
    const token = {};
    queueMutation.current = token;
    ++queueRevision.current;
    setQueueBusy(true);
    setQueueError("");
    const before = [...followupQueueItems];
    const chat = sessionToChatJid2(scope.sessionId);
    try {
      if (action === "steer") {
        if (itemOrIndex.chat_jid !== chat)
          throw new Error("Queued item belongs to another session");
        await steerAgentQueueItem(itemOrIndex.id, chat, expectedActiveTurnId);
        if (selection.isCurrent(scope))
          setFollowupQueueItems((items) => items.filter((item) => item.id !== itemOrIndex.id));
      } else if (action === "return") {
        const item = itemOrIndex;
        if (item.chat_jid !== chat || item.pending)
          throw new Error("Queued item belongs to another session or has no durable ID");
        const recovered = drafts.hasQueueReturn(scope.sessionId, item.id) ? emptyDraft() : await recoverQueueDraft(item, parseQueuedContent(item.content));
        const prepared = drafts.prepareQueueReturn(scope.sessionId, item.id, recovered);
        if (selection.current() === scope.sessionId) {
          setFileRefs(prepared.draft.fileRefs);
          setMessageRefs(prepared.draft.messageRefs);
          setDraftRestore({ sessionId: scope.sessionId, ...prepared.draft, token: crypto.randomUUID() });
        }
        await prepared.ready;
        await drafts.flushStable();
        await removeAgentQueueItem(item.id, chat);
        await drafts.completeQueueReturn(scope.sessionId, item.id);
      } else if (action === "remove") {
        if (itemOrIndex.chat_jid !== chat)
          throw new Error("Queued item belongs to another session");
        setFollowupQueueItems(before.filter((item) => item.id !== itemOrIndex.id));
        await removeAgentQueueItem(itemOrIndex.id, chat);
      } else {
        const after = [...before];
        if (!after[itemOrIndex] || toIndex < 0 || toIndex >= after.length)
          throw new Error("Invalid queue position");
        const [moved] = after.splice(itemOrIndex, 1);
        after.splice(toIndex, 0, moved);
        setFollowupQueueItems(after);
        await reorderAgentQueueItem({ chatJid: chat, expected: before.map((item) => item.id), order: after.map((item) => item.id) });
      }
    } catch (error) {
      if (action === "return")
        drafts.queueReturnFailed(scope.sessionId, itemOrIndex.id, error.message);
      if (selection.isCurrent(scope)) {
        setFollowupQueueItems(before);
        setQueueError(`Queue action failed: ${error.message}`);
      }
    } finally {
      try {
        const fresh = await getAgentQueueState(chat);
        if (selection.isCurrent(scope)) {
          setFollowupQueueItems(fresh.items || []);
          if (!streamDisconnected.current)
            setQueueActiveTurnId(fresh.activeTurnId || null);
        }
      } catch (error) {
        if (selection.isCurrent(scope))
          setQueueError(`Queue refresh failed: ${error.message}`);
      }
      if (queueMutation.current === token) {
        ++queueRevision.current;
        queueMutation.current = null;
        setQueueBusy(false);
      }
    }
  };
  const openEditor = Y_((path) => {
    ++tabFocusEpoch.current;
    if (window.matchMedia("(max-width: 1023px), (orientation: portrait)").matches)
      setWorkspaceOpen(false);
    const existing = tabs.find((t) => t.id === path || t.path === path);
    if (existing) {
      setActiveTabId(existing.id);
      return;
    }
    setTabs((prev) => [...prev, { id: path, path, label: path.split("/").pop() || path, dirty: false, pinned: false }]);
    setActiveTabId(path);
  }, [tabs]);
  const handleTabClose = Y_((id) => {
    setTabs((prev) => {
      const next = prev.filter((t) => t.id !== id);
      if (activeTabId === id)
        setActiveTabId(next[next.length - 1]?.id || null);
      if (!next.length) {
        const epoch = ++tabFocusEpoch.current;
        requestAnimationFrame(() => {
          if (epoch !== tabFocusEpoch.current || document.activeElement !== document.body || document.querySelector('.settings-dialog[aria-modal="true"]'))
            return;
          document.querySelector(".compose-box textarea")?.focus({ preventScroll: true });
        });
      }
      return next;
    });
  }, [activeTabId]);
  const appShellClass = [
    "app-shell",
    workspaceOpen ? "" : "workspace-collapsed",
    editorOpen ? "editor-open" : ""
  ].filter(Boolean).join(" ");
  W_(() => {
    if (!ready || !workspaceOpen)
      return;
    const sidebar = document.querySelector(".workspace-sidebar");
    if (sidebar)
      return bindWorkspaceVisibility(sidebar);
  }, [ready, workspaceOpen]);
  useWorkspaceFolderReference(ready && workspaceOpen, sessionId, fileRefs, (path) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    const refs = [...new Set([...getDraft(sessionId).fileRefs, path])];
    drafts.update(sessionId, { fileRefs: refs });
    setFileRefs(refs);
  });
  if (!ready) {
    return fe`<div id="app"><div style="padding:20px;text-align:center;color:var(--text-secondary,#888)">Loading…</div></div>`;
  }
  return fe`
        <div class=${appShellClass}>
            <style>${`.app-shell .post-content:has(table) { overflow-x: auto; } .app-shell .post-content table { display: table; width: 100%; table-layout: auto; }`}</style>
            <${SystemMetersHud} mode="overlay" />
            <${GiSettings}
                chatJid=${currentChatJid}
                onMutationStart=${() => {
    const token = { scope: selection.capture() };
    ++modelRevision.current;
    modelMutation.current = token;
    return token;
  }}
                onMutationEnd=${(token) => {
    if (modelMutation.current === token) {
      ++modelRevision.current;
      modelMutation.current = null;
    }
  }}
                onApplied=${(state, token) => {
    if (!selection.isCurrent(token.scope) || modelMutation.current !== token)
      return;
    setAgentModelsPayload((previous) => ({ ...previous, ...state }));
    setActiveModel(state.current);
    setActiveThinkingLevel(state.thinking_level);
    setSupportsThinking(state.supports_thinking);
    setContextUsage(state.context_usage || null);
  }}
            />
            ${!searchState.active && fe`<${TimelineQuickActions}
                key=${sessionId}
                currentChatJid=${currentChatJid}
                activeChatAgents=${activeChatAgents}
                workspaceOpen=${workspaceOpen}
                chatOnlyMode=${false}
                onToggleWorkspace=${() => setWorkspaceOpen((v) => !v)}
                onSwitchChat=${handleSwitchChat}
                onPrefillCompose=${(command) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    setComposePrefill({ sessionId, token: crypto.randomUUID(), text: command.trim() + " " });
  }}
            />`}
            <${TimelineMenu}
                workspaceOpen=${workspaceOpen}
                toggleWorkspace=${() => setWorkspaceOpen((v) => !v)}
                chatOnlyMode=${false}
                openEditor=${openEditor}
            />
            <${WorkspaceExplorer}
                onFileSelect=${(path) => {
    const refs = [...new Set([...getDraft(sessionId).fileRefs, path])];
    drafts.update(sessionId, { fileRefs: refs });
    setFileRefs(refs);
  }}
                visible=${workspaceOpen}
                active=${workspaceOpen || editorOpen}
                onOpenEditor=${openEditor}
                onOpenTerminalTab=${() => {}}
                onOpenVncTab=${() => {}}
            />
            ${workspaceOpen && fe`<button
                class="workspace-drawer-backdrop"
                onClick=${() => setWorkspaceOpen(false)}
                aria-label="Hide workspace"
                title="Hide workspace"
            ></button>`}
            <button
                class=${`workspace-toggle-tab${workspaceOpen ? " open" : " closed"}`}
                onClick=${() => setWorkspaceOpen((v) => !v)}
                title=${workspaceOpen ? "Hide workspace" : "Show workspace"}
                aria-label=${workspaceOpen ? "Hide workspace" : "Show workspace"}
                aria-expanded=${workspaceOpen ? "true" : "false"}
            >
                <svg class="workspace-toggle-tab-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polyline points="6 3 11 8 6 13" />
                </svg>
            </button>
            <div class="workspace-splitter"></div>
            ${editorOpen && fe`
                <div class="editor-pane-container gi-readonly-tabs" onContextMenuCapture=${(e) => {
    if (e.target.closest(".tab-item")) {
      e.preventDefault();
      e.stopPropagation();
    }
  }}>
                    <${TabStrip}
                        tabs=${tabs}
                        activeId=${activeTabId}
                        onActivate=${(id) => setActiveTabId(id)}
                        onClose=${handleTabClose}
                        onCloseOthers=${(id) => setTabs((p) => p.filter((t) => t.id === id))}
                        onCloseAll=${() => {
    setTabs([]);
    setActiveTabId(null);
  }}
                        onTogglePin=${() => {}}
                    />
                    <div class="editor-pane-host">
                        ${activeTabId && fe`<${WorkspaceTab} key=${activeTabId} path=${activeTabId} onClose=${() => handleTabClose(activeTabId)} />`}
                    </div>
                </div>
                <div class="editor-splitter"></div>
            `}
            <div class="container" ref=${containerRef} tabIndex="0" role="region" aria-label="Conversation">
                <${Timeline}
                    posts=${posts}
                    hasMore=${false}
                    onLoadMore=${() => loadPosts({ older: true })}
                    timelineRef=${timelineRef}
                    onHashtagClick=${() => {}}
                    onMessageRef=${(id) => {
    const refs = [...new Set([...getDraft(sessionId).messageRefs, id])];
    drafts.update(sessionId, { messageRefs: refs });
    setMessageRefs(refs);
  }}
                    onScrollToMessage=${() => {}}
                    onFileRef=${openEditor}
                    onPostClick=${undefined}
                    onDeletePost=${handleDeletePost}
                    onOpenWidget=${(w) => setFloatingWidget(w)}
                    onOpenAttachmentPreview=${setAttachmentPreview}
                    emptyMessage=${searchState.active ? searchState.query ? "No matching messages." : "Enter a search query." : "Send a message to get started."}
                    agents=${agents}
                    user=${userProfile}
                    reverse=${true}
                    removingPostIds=${removingPostIds}
                    searchQuery=${searchState.active ? searchState.query : ""}
                />
                <${AgentStatus} key=${`${sessionId}:${currentTurnId || ""}`}
                    status=${isCompactionStatus(agentStatus) ? null : agentStatus}
                    draft=${agentDraft}
                    plan=${agentPlan}
                    thought=${agentThought}
                    pendingRequest=${pendingRequest}
                    intent=${null}
                    turnId=${currentTurnId}
                    steerQueued=${Boolean(steerQueuedTurnId)}
                    showExtensionPanels=${false}
                />
                <${FloatingWidgetPane}
                    widget=${floatingWidget}
                    onClose=${() => setFloatingWidget(null)}
                    onWidgetEvent=${() => {}}
                />
                ${attachmentPreview && fe`
                    <${AttachmentPreviewModal}
                        mediaId=${attachmentPreview.mediaId}
                        info=${attachmentPreview.info}
                        onClose=${() => setAttachmentPreview(null)}
                    />
                `}
                <${RunBoundQueueStack}
                    steerEnabled=${connectionStatus === "connected" && isAgentTurnActive && !!queueActiveTurnId}
                    onInjectQueuedFollowup=${(item) => mutateQueue("steer", item)}
                    items=${[...followupQueueItems, ...optimisticQueue.filter((item) => item.chat_jid === currentChatJid && !followupQueueItems.some((stored) => stored.id === item.id || stored.metadata?.client_request_id === item.id))]}
                    busy=${queueBusy}
                    onReturnQueuedFollowup=${(item) => mutateQueue("return", item)}
                    onRemoveQueuedFollowup=${(item) => mutateQueue("remove", item)}
                    onMoveQueuedFollowup=${(from, to) => mutateQueue("move", from, to)}
                    onOpenFilePill=${openEditor}
                />
                ${followupQueueItems.some((item) => item.phase === "steer_returned") && fe`<div role="alert">Steer was not consumed by its target run. The item remains queued and will not auto-send; return it to the editor, remove it, or Steer a new active run.</div>`}
                ${queueError && fe`<div role="alert">${queueError}</div>`}
                ${newUIVersion && fe`<div role="status" class="gi-version-warning">New UI available. Reload manually when ready; unsaved editor work may be lost.</div>`}
                ${sessionError && fe`<div role="alert">${sessionError}</div>`}
                ${deleteError && fe`<div role="alert">${deleteError}</div>`}
                ${searchError && fe`<div role="alert">${searchError}</div>`}
                ${searchState.active && fe`<div role="status">Search${searchState.query ? `: ${searchState.query}` : ""} · ${searchState.scope} · up to 50 results</div>`}
                ${stopError && fe`<div role="alert">${stopError}</div>`}
                ${compactError && fe`<div role="alert">${compactError}</div>`}
                ${draftStorageError && fe`<div role="alert">${draftStorageError}</div>`}
                ${drafts.error(sessionId) && fe`<div role="alert">${drafts.error(sessionId)}</div>`}
                <${ComposeTransfer} sessionId=${sessionId} hidden=${searchState.active} />
                <${ComposeBox}
                    statusNotice=${notice}
                    prefillRequest=${composePrefill?.sessionId === sessionId ? composePrefill : null}
                    showQueueStack=${false}
                    key=${`${sessionId}:${draftRestore?.sessionId === sessionId ? draftRestore.token : ""}`}
                    draftValue=${getDraft(sessionId).text}
                    draftMediaFiles=${getDraft(sessionId).media}
                    onContentChange=${(text) => {
    drafts.update(sessionId, { text });
    setComposePrefill(null);
  }}
                    onDraftMediaChange=${(media) => drafts.update(sessionId, { media })}
                    focusRestoredDraft=${draftRestore?.sessionId === sessionId}
                    onCaptureDraft=${(draft) => drafts.begin(sessionId, draft)}
                    onQueuedSubmissionStart=${(token, text) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    setOptimisticQueue((items) => [...items, { id: token, content: text, chat_jid: currentChatJid, pending: true }]);
  }}
                    onQueuedSubmissionEnd=${(token) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    setOptimisticQueue((items) => items.filter((item) => item.id !== token));
    refreshSelectedState();
  }}
                    onDraftAccepted=${(token) => drafts.accepted(sessionId, token)}
                    onDraftFailed=${(token, error) => {
    const draft = drafts.failed(sessionId, token, error);
    if (selection.current() === sessionId) {
      setFileRefs(draft.fileRefs);
      setMessageRefs(draft.messageRefs);
      setDraftRestore({ sessionId, ...draft, token: crypto.randomUUID() });
    }
  }}
                    onDraftStorageError=${(error) => setDraftStorageError(`Send acknowledged, but draft cleanup failed: ${error.message}. Reload recovery may contain already-delivered text.`)}
                    currentChatJid=${currentChatJid}
                    isAgentActive=${isAgentTurnActive}
                    onPost=${handlePost}
                    onFocus=${() => {
    if (!isIOSDevice())
      scrollToBottom();
  }}
                    onModelMutationStart=${() => {
    const token = {};
    ++modelRevision.current;
    modelMutation.current = token;
    return token;
  }}
                    onModelMutationEnd=${(token) => {
    if (modelMutation.current === token) {
      ++modelRevision.current;
      modelMutation.current = null;
    }
  }}
                    onModelChange=${(value) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    setActiveModel(value || "");
  }}
                    onModelStateChange=${(state) => {
    if (!selection.isCurrent(renderedSelection))
      return;
    if (state && typeof state === "object") {
      setAgentModelsPayload((prev) => ({ ...prev || {}, ...state || {} }));
      if (typeof state.model === "string")
        setActiveModel(state.model);
      if (typeof state.thinking_level_label === "string" && state.thinking_level_label.trim()) {
        setActiveThinkingLevel(state.thinking_level_label);
      } else if (typeof state.thinking_level === "string" && state.thinking_level.trim()) {
        setActiveThinkingLevel(state.thinking_level);
      }
      if (typeof state.supports_thinking === "boolean")
        setSupportsThinking(state.supports_thinking);
      if (state.provider_usage !== undefined)
        setModelUsage(state.provider_usage ?? null);
      if (state.context_usage !== undefined)
        setContextUsage(state.context_usage);
    }
  }}
                    agents=${agents}
                    currentSessionAgent=${activeChatAgents.find((entry) => entry?.chat_jid === currentChatJid) || null}
                    agentStatus=${agentStatus}
                    agentDraft=${agentDraft}
                    contextUsage=${contextUsage}
                    activeEditorPath=${activeTabId}
                    onAttachEditorFile=${() => {
    if (!activeTabId)
      return;
    const refs = [...new Set([...getDraft(sessionId).fileRefs, activeTabId])];
    drafts.update(sessionId, { fileRefs: refs });
    setFileRefs(refs);
  }}
                    fileRefs=${fileRefs}
                    messageRefs=${messageRefs}
                    onRemoveFileRef=${(p) => {
    const refs = getDraft(sessionId).fileRefs.filter((x) => x !== p);
    drafts.update(sessionId, { fileRefs: refs });
    setFileRefs(refs);
  }}
                    onClearFileRefs=${() => {
    drafts.update(sessionId, { fileRefs: [] });
    if (selection.current() === sessionId)
      setFileRefs([]);
  }}
                    onSetFileRefs=${(refs) => {
    drafts.update(sessionId, { fileRefs: refs });
    if (selection.current() === sessionId)
      setFileRefs(refs);
  }}
                    onRemoveMessageRef=${(id) => {
    const refs = getDraft(sessionId).messageRefs.filter((x) => x !== id);
    drafts.update(sessionId, { messageRefs: refs });
    setMessageRefs(refs);
  }}
                    onClearMessageRefs=${() => {
    drafts.update(sessionId, { messageRefs: [] });
    if (selection.current() === sessionId)
      setMessageRefs([]);
  }}
                    onSetMessageRefs=${(refs) => {
    drafts.update(sessionId, { messageRefs: refs });
    if (selection.current() === sessionId)
      setMessageRefs(refs);
  }}
                    connectionStatus=${connectionStatus}
                    activeChatAgents=${activeChatAgents}
                    currentChatBranches=${currentChatBranches}
                    onSwitchChat=${handleSwitchChat}
                    onCreateSession=${handleCreateSession}
                    onRenameSession=${(chatJid, title) => handleSessionMutation(chatJid, "rename", title)}
                    onPinSession=${(chatJid, pinned) => handleSessionMutation(chatJid, "pin", pinned)}
                    onArchiveSession=${(chatJid) => handleSessionMutation(chatJid, "archive")}
                    onRestoreSession=${(chatJid) => handleSessionMutation(chatJid, "restore")}
                    formatBranchPickerLabel=${(b) => b?.label || b?.chat_jid || ""}
                    handleBranchPickerChange=${() => {}}
                    searchMode=${searchState.active}
                    onEnterSearch=${enterSearch}
                    onExitSearch=${exitSearch}
                    onSearch=${(query) => runSearch(query)}
                    searchScope=${searchState.scope}
                    onSearchScopeChange=${(scope) => runSearch(undefined, scope)}
                    activeModel=${activeModel}
                    agentModelsPayload=${agentModelsPayload}
                    modelUsage=${modelUsage}
                    thinkingLevel=${activeThinkingLevel}
                    supportsThinking=${supportsThinking}
                    followupQueueCount=${followupQueueItems.length}
                    notificationsEnabled=${false}
                    notificationPermission="default"
                    onComposeSubmitError=${() => {}}
                    pendingRequestRef=${pendingRequestRef}
                    setPendingRequest=${setPendingRequest}
                />
            </div>
        </div>
    `;
}
function ComposeTransfer({ sessionId, hidden }) {
  const [, repaint] = F_(0);
  const ref = Q_(null);
  K_(() => composeTransfers.subscribe(() => repaint((n) => n + 1)), []);
  const state = composeTransfers.snapshot(sessionId);
  W_(() => {
    const root = ref.current?.parentElement;
    if (!root)
      return;
    return bindComposeSending(root, !hidden && state.sending > 0);
  }, [sessionId, hidden, state.sending]);
  const percent = state.computable && state.total > 0 ? Math.floor(state.loaded * 100 / state.total) : null;
  return fe`<div ref=${ref} class="gi-compose-transfer" hidden=${hidden || !state.uploads && !state.sending}>
        ${state.uploads > 0 && fe`<div class="gi-compose-upload" role="status" aria-live="polite">
            <span>Uploading ${state.uploads === 1 ? "attachment" : `${state.uploads} attachments`}${percent === null ? "…" : ` · ${percent}%${percent === 100 ? " · awaiting server" : ""}`}</span>
            <progress aria-label="Attachment upload progress" max="100" value=${percent === null ? undefined : percent}></progress>
        </div>`}
        ${state.sending > 0 && fe`<div class="gi-compose-sending" role="status" aria-live="polite">Sending${state.sending > 1 ? ` ${state.sending} messages` : " message"}…</div>`}
    </div>`;
}
G_(fe`<${GiApp} />`, document.getElementById("app"));
export {
  F_,
  K_,
  W_,
  Q_,
  fe,
  subscribeModelSettlement,
  getAgentStatus,
  getSessionCompaction,
  compactSession,
  cancelSessionRun,
  getGiProviders,
  saveGiProviderKey,
  removeGiProviderKey,
  getGiCompactionPolicy,
  saveGiCompactionPolicy,
  getAgentModels,
  selectAgentModel,
  defaultAppearance,
  appearancePresets,
  currentAppearance,
  persistAppearance,
  subscribeAppearance,
  modelContextBlocked,
  compactionNotice,
  compactionElapsed
};

//# debugId=692CF69AE67CEA9764756E2164756E21
//# sourceMappingURL=app-ktfdb5kf.js.map
