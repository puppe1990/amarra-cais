const POPSTATE = "_amarraNavPopstate";

// #27: Drive morphs only #amarra-main, so persistent chrome outside main
// (sticky sidebars, locale toggles) keeps stale active-link classes after
// navigation. The nav hook re-derives the active link from location on
// connect, amarra:morphed (via updated), and popstate.
export function makeNav(opts = {}) {
  const getLocation = opts.location ?? (() => globalThis.location);
  const getWindow = opts.window ?? globalThis.window;

  function sync(el) {
    const loc = getLocation();
    if (!el || !loc?.pathname) return;
    const on = classes(el, "data-amarra-nav-on", opts.onClasses);
    const off = classes(el, "data-amarra-nav-off", opts.offClasses);
    for (const link of el.querySelectorAll?.("a[href]") ?? []) {
      if (isActive(link, loc)) {
        off.forEach((c) => link.classList?.remove(c));
        on.forEach((c) => link.classList?.add(c));
        link.setAttribute?.("aria-current", "page");
      } else {
        on.forEach((c) => link.classList?.remove(c));
        if (off.length) off.forEach((c) => link.classList?.add(c));
        link.removeAttribute?.("aria-current");
      }
    }
  }

  return {
    connect(el) {
      if (!el) return;
      sync(el);
      const onPop = () => sync(el);
      el[POPSTATE] = onPop;
      getWindow?.addEventListener?.("popstate", onPop);
    },
    updated(el) {
      sync(el);
    },
    disconnect(el) {
      const fn = el?.[POPSTATE];
      if (!fn) return;
      getWindow?.removeEventListener?.("popstate", fn);
      delete el[POPSTATE];
    },
  };
}

function classes(el, attr, fallback) {
  const raw = el.getAttribute?.(attr) || fallback;
  if (!raw) return [];
  return raw.split(/\s+/).filter(Boolean);
}

// Query string ignored: /settings and /settings?tab=keys are the same page.
function isActive(link, loc) {
  const href = link.getAttribute?.("href");
  if (!href) return false;
  try {
    return new URL(href, loc.href).pathname === loc.pathname;
  } catch {
    return false;
  }
}

export const nav = makeNav();
