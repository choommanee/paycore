/* ============================================================================
   PayCore Design System — paycore.js
   Small vanilla helper lib. No framework, no build step. Pairs with paycore.css.

   Include (defer recommended):
     <script src="/assets/paycore.js" defer></script>

   Exposes a single global:  window.PC   (also aliased as window.PayCore)

   ----------------------------------------------------------------------------
   API surface
   ----------------------------------------------------------------------------
   Brand / assets
     PC.logo(opts)            -> SVG string (wordmark | mark-only)
     PC.logoMark(size)        -> square SVG mark string (favicon-able)
     PC.faviconDataURL()      -> data: URL for <link rel=icon>
     PC.mount(sel|el)         -> inject logo into an element

   Config / networking  (base URL + key resolution)
     PC.config()              -> { base, key }
     PC.setKey(k) / PC.clearKey()
     PC.api(path, opts)       -> fetch wrapper, injects base + key + JSON, unwraps envelope
     PC.get/post/patch/del(path, body, opts)

   Money (satang <-> THB) — gateway money is int64 minor units (satang)
     PC.money(satang, opts)   -> "฿1,290.00"
     PC.satang(thb)           -> 129000
     PC.moneyFromDecimal(str) -> "฿1,290.00"  (accepts "1290.00" major-unit strings)

   UI helpers
     PC.badge(status, label?) -> status-badge HTML string
     PC.statusEl(status)      -> <span> element
     PC.toast(msg, opts)      -> transient toast
     PC.modal(opts)           -> open modal, returns { close }
     PC.confirm(opts)         -> Promise<boolean>
     PC.empty(opts)           -> empty-state HTML string
     PC.el(tag, attrs, kids)  -> tiny element factory
     PC.escape(str)           -> HTML-escape
     PC.fmtDate(v) / PC.timeAgo(v)

   Router (hash-based dashboard tabs)
     PC.router({ mount, routes, default })
   ========================================================================== */
(function (root) {
  'use strict';

  var PC = {};

  /* ---------------------------------------------------------------- DOM -- */
  function q(sel) { return typeof sel === 'string' ? document.querySelector(sel) : sel; }

  function escapeHTML(s) {
    if (s == null) return '';
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }
  PC.escape = escapeHTML;

  // el('div', {class:'x', onclick:fn, dataset:{id:1}}, ['text', childEl])
  PC.el = function (tag, attrs, kids) {
    var e = document.createElement(tag);
    if (attrs) Object.keys(attrs).forEach(function (k) {
      var v = attrs[k];
      if (v == null) return;
      if (k === 'class' || k === 'className') e.className = v;
      else if (k === 'html') e.innerHTML = v;
      else if (k === 'text') e.textContent = v;
      else if (k === 'dataset') Object.keys(v).forEach(function (d) { e.dataset[d] = v[d]; });
      else if (k.slice(0, 2) === 'on' && typeof v === 'function') e.addEventListener(k.slice(2).toLowerCase(), v);
      else e.setAttribute(k, v);
    });
    if (kids != null) (Array.isArray(kids) ? kids : [kids]).forEach(function (c) {
      if (c == null) return;
      e.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
    });
    return e;
  };

  /* -------------------------------------------------------------- brand -- */
  // PayCore mark: a rounded-square shield with a stylized "P" cut, in accent blue.
  // Self-contained inline SVG — no external assets, currentColor-friendly.
  var ACCENT = '#185fa5', ACCENT_D = '#0c447c', INK = '#0f1729';

  PC.logoMark = function (size) {
    size = size || 28;
    return (
      '<svg width="' + size + '" height="' + size + '" viewBox="0 0 32 32" fill="none" ' +
      'xmlns="http://www.w3.org/2000/svg" role="img" aria-label="PayCore">' +
        '<defs><linearGradient id="pcg" x1="0" y1="0" x2="32" y2="32" gradientUnits="userSpaceOnUse">' +
          '<stop stop-color="' + ACCENT + '"/><stop offset="1" stop-color="' + ACCENT_D + '"/>' +
        '</linearGradient></defs>' +
        '<rect x="1" y="1" width="30" height="30" rx="8" fill="url(#pcg)"/>' +
        // "P" glyph carved from the tile
        '<path d="M11 8.5h6.4c3 0 5.1 2 5.1 4.9 0 3-2.1 5-5.1 5H14v5.1h-3V8.5Zm3 3v3.8h3.1c1.2 0 2-.75 2-1.9 0-1.15-.8-1.9-2-1.9H14Z" fill="#fff"/>' +
        // small "core" dot
        '<circle cx="21.5" cy="21.5" r="2.2" fill="#fff" opacity=".92"/>' +
      '</svg>'
    );
  };

  // Full lockup: mark + "PayCore" wordmark. opts:{ size, mono, name }
  PC.logo = function (opts) {
    opts = opts || {};
    var size = opts.size || 26;
    var name = opts.name != null ? opts.name : 'PayCore';
    var wordColor = opts.mono ? 'currentColor' : INK;
    var accentColor = opts.mono ? 'currentColor' : ACCENT;
    if (opts.markOnly) return PC.logoMark(size);
    return (
      '<span class="pc-brand" style="display:inline-flex;align-items:center;gap:8px">' +
        PC.logoMark(size) +
        (name ? '<span class="pc-brand-name" style="font-weight:600;letter-spacing:-.01em;color:' + wordColor + '">' +
          'Pay<span style="color:' + accentColor + '">Core</span></span>' : '') +
      '</span>'
    );
  };

  PC.faviconDataURL = function () {
    return 'data:image/svg+xml,' + encodeURIComponent(PC.logoMark(32));
  };

  // Inject logo markup into an element. PC.mount('#brand', {size:28})
  PC.mount = function (sel, opts) {
    var e = q(sel);
    if (e) e.innerHTML = PC.logo(opts);
    return e;
  };

  // Auto-set favicon + fill [data-pc-logo] elements on load.
  function autobrand() {
    try {
      if (!document.querySelector('link[rel~="icon"]')) {
        var l = document.createElement('link');
        l.rel = 'icon'; l.href = PC.faviconDataURL();
        document.head.appendChild(l);
      }
    } catch (_) {}
    document.querySelectorAll('[data-pc-logo]').forEach(function (n) {
      var mo = n.getAttribute('data-pc-logo') === 'mark';
      n.innerHTML = PC.logo({ markOnly: mo, size: +n.getAttribute('data-size') || 26 });
    });
  }

  /* ------------------------------------------------------------- config -- */
  // Base URL: ?api= query param  >  window.PAYCORE_API  >  localStorage  >  same-origin.
  // The pages are served same-origin with the API (WEB_DIR), so the default is
  // location.origin — this "just works" both locally (localhost:8080) and on a
  // hosted deploy (e.g. Railway), with no hardcoded host. Override with ?api=.
  // Key (merchant sk_/admin key): ?key=  >  window.PAYCORE_KEY  >  localStorage.
  var LS_BASE = 'paycore.api', LS_KEY = 'paycore.key';
  var SAME_ORIGIN = (root.location && root.location.origin) || '';

  function ls(k) { try { return localStorage.getItem(k); } catch (_) { return null; } }
  function lsSet(k, v) { try { v == null ? localStorage.removeItem(k) : localStorage.setItem(k, v); } catch (_) {} }

  PC.config = function () {
    var qp = new URLSearchParams(location.search);
    var qBase = qp.get('api');
    if (qBase) lsSet(LS_BASE, qBase);            // persist for later loads
    var qKey = qp.get('key');
    if (qKey) lsSet(LS_KEY, qKey);
    var base = qBase || root.PAYCORE_API || ls(LS_BASE) || SAME_ORIGIN;
    var key = qKey || root.PAYCORE_KEY || ls(LS_KEY) || '';
    return { base: base.replace(/\/+$/, ''), key: key };
  };
  PC.setKey = function (k) { lsSet(LS_KEY, k || null); };
  PC.clearKey = function () { lsSet(LS_KEY, null); };
  PC.setBase = function (b) { lsSet(LS_BASE, b || null); };

  /* ----------------------------------------------------------- networking */
  // Unwrap the standard { success, code, data, message } API envelope.
  function unwrap(res, body) {
    if (!res.ok || (body && body.success === false)) {
      var msg = (body && (body.message || body.code)) || ('HTTP ' + res.status);
      var err = new Error(msg);
      err.status = res.status; err.body = body; throw err;
    }
    return body && Object.prototype.hasOwnProperty.call(body, 'data') ? body.data : body;
  }

  // PC.api('/v1/payments', { method, body, headers, key, raw, idempotency })
  PC.api = function (path, opts) {
    opts = opts || {};
    var cfg = PC.config();
    var url = /^https?:/.test(path) ? path : cfg.base + (path[0] === '/' ? path : '/' + path);
    var headers = Object.assign({ 'Accept': 'application/json' }, opts.headers || {});

    var body = opts.body;
    if (body != null && typeof body === 'object' && !(body instanceof FormData) &&
        !(body instanceof URLSearchParams) && typeof body !== 'string') {
      headers['Content-Type'] = headers['Content-Type'] || 'application/json';
      body = JSON.stringify(body);
    }

    // Inject the stored key. Admin/merchant APIs use X-API-Key (per PayCore APIKeyAuth).
    var key = opts.key != null ? opts.key : cfg.key;
    if (key && opts.auth !== false) headers['X-API-Key'] = key;
    if (opts.idempotency) headers['Idempotency-Key'] = opts.idempotency === true
      ? (crypto.randomUUID ? crypto.randomUUID() : String(Date.now()) + Math.random())
      : opts.idempotency;

    return fetch(url, {
      method: opts.method || (body ? 'POST' : 'GET'),
      headers: headers, body: body, signal: opts.signal,
    }).then(function (res) {
      if (opts.raw) return res;
      return res.json().catch(function () { return null; }).then(function (j) { return unwrap(res, j); });
    });
  };
  PC.get = function (p, o) { return PC.api(p, Object.assign({ method: 'GET' }, o)); };
  PC.post = function (p, b, o) { return PC.api(p, Object.assign({ method: 'POST', body: b }, o)); };
  PC.patch = function (p, b, o) { return PC.api(p, Object.assign({ method: 'PATCH', body: b }, o)); };
  PC.del = function (p, o) { return PC.api(p, Object.assign({ method: 'DELETE' }, o)); };

  /* --------------------------------------------------------------- money -- */
  // Gateway money is int64 minor units (satang). 129000 satang -> "฿1,290.00".
  PC.money = function (satang, opts) {
    opts = opts || {};
    var n = Number(satang || 0) / 100;
    var cur = opts.currency || 'THB';
    var sym = opts.symbol != null ? opts.symbol : (cur === 'THB' ? '฿' : '');
    var s = n.toLocaleString(opts.locale || 'th-TH', {
      minimumFractionDigits: opts.min != null ? opts.min : 2,
      maximumFractionDigits: opts.max != null ? opts.max : 2,
    });
    if (sym) return sym + s;
    return s + (opts.code === false ? '' : ' ' + cur);
  };
  // THB (major, number or string) -> integer satang. PC.satang("1290.00") -> 129000
  PC.satang = function (thb) { return Math.round(Number(thb || 0) * 100); };
  // Accept a decimal major-unit string (the checkout API's amount format) -> "฿1,290.00"
  PC.moneyFromDecimal = function (dec, opts) { return PC.money(PC.satang(dec), opts); };

  /* ------------------------------------------------------------- badges -- */
  var STATUS_LABELS = {
    authorized: 'Authorized', captured: 'Captured', paid: 'Paid', succeeded: 'Succeeded',
    settled: 'Settled', refunded: 'Refunded', partially_refunded: 'Partially refunded',
    failed: 'Failed', declined: 'Declined', voided: 'Voided', canceled: 'Canceled',
    cancelled: 'Cancelled', expired: 'Expired', pending: 'Pending', processing: 'Processing',
    requires_action: 'Requires action', disputed: 'Disputed', chargeback: 'Chargeback',
  };
  PC.statusLabel = function (s) {
    if (!s) return '';
    return STATUS_LABELS[s] || String(s).replace(/_/g, ' ').replace(/\b\w/g, function (c) { return c.toUpperCase(); });
  };
  // Return an HTML string for a status badge.
  PC.badge = function (status, label) {
    var key = String(status || '').toLowerCase();
    return '<span class="pc-badge" data-status="' + escapeHTML(key) + '">' +
      escapeHTML(label != null ? label : PC.statusLabel(key)) + '</span>';
  };
  // Return a live DOM element (handy for appendChild).
  PC.statusEl = function (status, label) {
    var t = document.createElement('template');
    t.innerHTML = PC.badge(status, label);
    return t.content.firstChild;
  };

  /* -------------------------------------------------------------- toasts -- */
  function toastLayer() {
    var l = document.querySelector('.pc-toasts');
    if (!l) { l = PC.el('div', { class: 'pc-toasts' }); document.body.appendChild(l); }
    return l;
  }
  var TOAST_ICONS = {
    success: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>',
    error: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="m15 9-6 6M9 9l6 6"/></svg>',
    warn: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.3 3.7 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.7a2 2 0 0 0-3.4 0Z"/><path d="M12 9v4M12 17h.01"/></svg>',
    info: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 16v-4M12 8h.01"/></svg>',
  };
  // PC.toast('Saved', { type:'success', title, duration })  OR  PC.toast({...})
  PC.toast = function (msg, opts) {
    if (typeof msg === 'object') { opts = msg; msg = opts.message; }
    opts = opts || {};
    var type = opts.type || 'info';
    var dur = opts.duration != null ? opts.duration : 4000;
    var t = PC.el('div', { class: 'pc-toast ' + type });
    t.innerHTML =
      '<span class="pc-toast-ico">' + (TOAST_ICONS[type] || TOAST_ICONS.info) + '</span>' +
      '<div class="pc-toast-body">' +
        (opts.title ? '<div class="pc-toast-title">' + escapeHTML(opts.title) + '</div>' : '') +
        (msg ? '<div class="pc-toast-msg">' + escapeHTML(msg) + '</div>' : '') +
      '</div>' +
      '<button class="pc-toast-close" aria-label="Dismiss">&times;</button>';
    var remove = function () {
      t.classList.add('leaving');
      t.addEventListener('animationend', function () { t.remove(); });
      setTimeout(function () { t.remove(); }, 240);
    };
    t.querySelector('.pc-toast-close').addEventListener('click', remove);
    toastLayer().appendChild(t);
    if (dur > 0) setTimeout(remove, dur);
    return { dismiss: remove, el: t };
  };

  /* --------------------------------------------------------------- modal -- */
  var X_ICON = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12"/></svg>';
  // PC.modal({ title, body:htmlOrNode, footer:htmlOrNode, size, onClose, closeOnBackdrop })
  PC.modal = function (opts) {
    opts = opts || {};
    var overlay = PC.el('div', { class: 'pc-overlay show' });
    var modal = PC.el('div', { class: 'pc-modal' + (opts.size ? ' ' + opts.size : '') });
    var head = opts.title !== false
      ? '<div class="pc-modal-head"><h3>' + escapeHTML(opts.title || '') + '</h3>' +
        '<button class="pc-modal-close" aria-label="Close">' + X_ICON + '</button></div>'
      : '';
    modal.innerHTML = head + '<div class="pc-modal-body"></div>' +
      (opts.footer ? '<div class="pc-modal-foot"></div>' : '');
    var bodyEl = modal.querySelector('.pc-modal-body');
    var footEl = modal.querySelector('.pc-modal-foot');
    put(bodyEl, opts.body); if (footEl) put(footEl, opts.footer);
    overlay.appendChild(modal);
    document.body.appendChild(overlay);

    function put(host, c) {
      if (c == null) return;
      if (typeof c === 'string') host.innerHTML = c; else host.appendChild(c);
    }
    function close() {
      overlay.remove();
      document.removeEventListener('keydown', onKey);
      if (opts.onClose) opts.onClose();
    }
    function onKey(e) { if (e.key === 'Escape') close(); }
    var closeBtn = modal.querySelector('.pc-modal-close');
    if (closeBtn) closeBtn.addEventListener('click', close);
    if (opts.closeOnBackdrop !== false) overlay.addEventListener('click', function (e) { if (e.target === overlay) close(); });
    document.addEventListener('keydown', onKey);
    return { close: close, el: modal, overlay: overlay, body: bodyEl, footer: footEl };
  };

  // PC.confirm({ title, message, confirmText, danger }) -> Promise<boolean>
  PC.confirm = function (opts) {
    opts = opts || {};
    return new Promise(function (resolve) {
      var done = function (v) { m.close(); resolve(v); };
      var foot = PC.el('div');
      var cancel = PC.el('button', { class: 'pc-btn pc-btn-secondary', text: opts.cancelText || 'Cancel', onclick: function () { done(false); } });
      var ok = PC.el('button', {
        class: 'pc-btn ' + (opts.danger ? 'pc-btn-danger' : 'pc-btn-primary'),
        text: opts.confirmText || 'Confirm', onclick: function () { done(true); },
      });
      foot.appendChild(cancel); foot.appendChild(ok);
      var m = PC.modal({
        title: opts.title || 'Are you sure?', size: 'sm',
        body: '<p class="pc-muted" style="margin:0">' + escapeHTML(opts.message || '') + '</p>',
        footer: foot, onClose: function () { resolve(false); },
      });
    });
  };

  /* --------------------------------------------------------- empty state -- */
  var BOX_ICON = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8 12 3 3 8v8l9 5 9-5V8Z"/><path d="M3 8l9 5 9-5M12 13v8"/></svg>';
  // PC.empty({ title, message, icon, actionHTML })
  PC.empty = function (opts) {
    opts = opts || {};
    return '<div class="pc-empty">' +
      '<div class="pc-empty-ico">' + (opts.icon || BOX_ICON) + '</div>' +
      '<h3>' + escapeHTML(opts.title || 'Nothing here yet') + '</h3>' +
      (opts.message ? '<p>' + escapeHTML(opts.message) + '</p>' : '') +
      (opts.actionHTML || '') + '</div>';
  };

  /* ---------------------------------------------------------------- dates -- */
  PC.fmtDate = function (v, opts) {
    if (!v) return '';
    var d = v instanceof Date ? v : new Date(v);
    if (isNaN(d)) return escapeHTML(String(v));
    return d.toLocaleString(opts && opts.locale || 'en-GB', opts && opts.opts || {
      year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit',
    });
  };
  PC.timeAgo = function (v) {
    var d = v instanceof Date ? v : new Date(v);
    if (isNaN(d)) return '';
    var s = Math.round((Date.now() - d.getTime()) / 1000);
    var units = [['y', 31536000], ['mo', 2592000], ['d', 86400], ['h', 3600], ['m', 60]];
    for (var i = 0; i < units.length; i++) {
      var n = Math.floor(Math.abs(s) / units[i][1]);
      if (n >= 1) return (s < 0 ? 'in ' : '') + n + units[i][0] + (s < 0 ? '' : ' ago');
    }
    return 'just now';
  };

  /* -------------------------------------------------------------- router -- */
  // Minimal hash router for dashboard tabs.
  //   PC.router({ mount:'#view', default:'overview', routes:{
  //     overview: function(ctx){ ctx.el.innerHTML = '...'; },
  //     payments: function(ctx){ ... },
  //   }})
  // Marks [data-route="name"] nav elements .active. ctx = { name, params, el }.
  PC.router = function (opts) {
    var mount = q(opts.mount);
    var routes = opts.routes || {};
    var def = opts.default || Object.keys(routes)[0];

    function parse() {
      var h = (location.hash || '').replace(/^#\/?/, '');
      var parts = h.split('?');
      var name = parts[0] || def;
      var params = {};
      if (parts[1]) new URLSearchParams(parts[1]).forEach(function (v, k) { params[k] = v; });
      // sub-path segments after name (e.g. #/payments/123)
      var segs = name.split('/');
      name = segs.shift();
      if (segs.length) params._path = segs;
      return { name: routes[name] ? name : def, params: params };
    }
    function markNav(name) {
      document.querySelectorAll('[data-route]').forEach(function (n) {
        n.classList.toggle('active', n.getAttribute('data-route') === name);
      });
    }
    function render() {
      var r = parse();
      markNav(r.name);
      var fn = routes[r.name];
      if (typeof fn === 'function') {
        try { fn({ name: r.name, params: r.params, el: mount, go: PC.router.go }); }
        catch (e) { console.error('[PC.router]', e); if (mount) mount.innerHTML = PC.empty({ title: 'Failed to render', message: e.message }); }
      }
    }
    PC.router.go = function (name, params) {
      var hash = '#/' + name;
      if (params) hash += '?' + new URLSearchParams(params).toString();
      if (location.hash === hash) render(); else location.hash = hash;
    };
    window.addEventListener('hashchange', render);
    render();
    return { go: PC.router.go, render: render };
  };

  /* --------------------------------------------------------- table build -- */
  // Convenience: build a table body from rows + column defs.
  // cols: [{ key, label, class, render:(row)=>htmlString }]
  PC.tableRows = function (rows, cols) {
    if (!rows || !rows.length) return '';
    return rows.map(function (row) {
      return '<tr>' + cols.map(function (c) {
        var val = c.render ? c.render(row) : escapeHTML(row[c.key] == null ? '' : row[c.key]);
        return '<td' + (c.class ? ' class="' + c.class + '"' : '') + '>' + val + '</td>';
      }).join('') + '</tr>';
    }).join('');
  };

  /* --------------------------------------------------------------- init -- */
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', autobrand);
  else autobrand();

  root.PC = root.PayCore = PC;
})(window);
