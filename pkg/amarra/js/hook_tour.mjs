const STATE = "_amarraTourState";
const START = "_amarraTourStart";
const PAD = 6;
const SHADOW = "0 0 0 9999px rgba(0, 0, 0, 0.72)";
const MIN_VISIBLE = 4;

// #322: every product reimplements the same spotlight walkthrough. This hook
// ships the engine (overlay, hole, balloon, keyboard) while the app only marks
// the steps and their copy with data-amarra-tour-* attributes.
// Markup: container[amarra-hook=tour] > [data-amarra-tour-start]
// + N x [data-amarra-tour-step] (title/text via data-amarra-tour-title/-text),
// walked in DOM order. Invisible steps (display:none, visibility:hidden, no
// layout box) are skipped so a responsive page does not tour hidden nodes.
export function makeTour(deps = {}) {
  const raf =
    deps.requestAnimationFrame ??
    ((fn) => {
      const doc = deps.document;
      const win = doc?.defaultView;
      if (win?.requestAnimationFrame) return win.requestAnimationFrame(fn);
      return fn();
    });

  const api = {
    connect(el) {
      if (!el?.querySelectorAll) return;
      const doc = el.ownerDocument ?? deps.document ?? null;
      if (!doc) return;
      const starts = [...(el.querySelectorAll("[data-amarra-tour-start]") ?? [])];
      const state = { el, doc, starts, tour: null };
      for (const btn of starts) {
        const fn = (ev) => {
          ev?.preventDefault?.();
          begin(state, btn, raf);
        };
        btn[START] = fn;
        btn.addEventListener?.("click", fn);
      }
      el[STATE] = state;
    },
    updated(el) {
      api.disconnect(el);
      api.connect(el);
    },
    disconnect(el) {
      const state = el?.[STATE];
      if (!state) return;
      for (const btn of state.starts ?? []) {
        const fn = btn?.[START];
        if (!fn) continue;
        btn.removeEventListener?.("click", fn);
        delete btn[START];
      }
      end(state);
      delete el[STATE];
    },
  };
  return api;
}

export const tour = makeTour();

function begin(state, opener, raf) {
  if (state.tour) return;
  const steps = collectSteps(state.el, state.doc);
  if (!steps.length) return;
  const t = {
    state,
    raf,
    steps,
    index: 0,
    ui: buildUI(state.doc),
    opener,
    queued: false,
  };
  t.onKey = (ev) => onKey(t, ev);
  t.onPlace = () => schedulePlace(t);
  bindButtons(t);
  state.doc.addEventListener?.("keydown", t.onKey, true);
  state.doc.defaultView?.addEventListener?.("scroll", t.onPlace, true);
  state.doc.defaultView?.addEventListener?.("resize", t.onPlace, true);
  state.tour = t;
  show(t, 0);
}

function end(state) {
  const t = state?.tour;
  if (!t) return;
  state.doc.removeEventListener?.("keydown", t.onKey);
  state.doc.defaultView?.removeEventListener?.("scroll", t.onPlace, true);
  state.doc.defaultView?.removeEventListener?.("resize", t.onPlace, true);
  for (const node of [t.ui.block, t.ui.hole, t.ui.tip]) {
    node?.parentNode?.removeChild?.(node);
  }
  state.tour = null;
  t.opener?.focus?.();
}

function bindButtons(t) {
  const { prevBtn, nextBtn, skipBtn } = t.ui;
  prevBtn.addEventListener?.("click", (ev) => {
    ev?.preventDefault?.();
    if (t.index > 0) show(t, t.index - 1);
  });
  nextBtn.addEventListener?.("click", (ev) => {
    ev?.preventDefault?.();
    if (t.index < t.steps.length - 1) show(t, t.index + 1);
    else end(t.state);
  });
  skipBtn.addEventListener?.("click", (ev) => {
    ev?.preventDefault?.();
    end(t.state);
  });
}

function show(t, index) {
  t.index = Math.max(0, Math.min(index, t.steps.length - 1));
  const step = t.steps[t.index];
  step.el.scrollIntoView?.({ block: "center", inline: "nearest" });
  t.ui.heading.textContent = step.copy.title;
  t.ui.body.textContent = step.copy.text;
  t.ui.count.textContent = `${t.index + 1} / ${t.steps.length}`;
  t.ui.prevBtn.disabled = t.index === 0;
  t.ui.nextBtn.textContent = t.index === t.steps.length - 1 ? "Concluir" : "Próximo →";
  place(t);
  schedulePlace(t);
}

function schedulePlace(t) {
  if (t.queued) return;
  t.queued = true;
  t.raf(() => {
    t.queued = false;
    place(t);
  });
}

function place(t) {
  const step = t.steps[t.index];
  if (!step) return;
  const rect = step.el.getBoundingClientRect?.();
  if (!rect) return;
  const { hole, tip } = t.ui;
  hole.style.top = `${rect.top - PAD}px`;
  hole.style.left = `${rect.left - PAD}px`;
  hole.style.width = `${rect.width + PAD * 2}px`;
  hole.style.height = `${rect.height + PAD * 2}px`;

  const tipRect = tip.getBoundingClientRect?.() ?? { width: 0, height: 0 };
  const win = t.state.doc.defaultView;
  const vw = win?.innerWidth ?? 0;
  const vh = win?.innerHeight ?? 0;
  const below = rect.bottom + PAD + 14;
  const above = rect.top - PAD - 14 - tipRect.height;
  let top = below + tipRect.height + 12 <= vh ? below : above;
  let left = rect.left + rect.width / 2 - tipRect.width / 2;
  left = Math.max(12, Math.min(left, vw - tipRect.width - 12));
  top = Math.max(12, Math.min(top, vh - tipRect.height - 12));
  tip.style.top = `${top}px`;
  tip.style.left = `${left}px`;
}

function onKey(t, ev) {
  const key = ev?.key;
  if (key === "ArrowRight" || key === "Enter") {
    ev?.preventDefault?.();
    if (t.index < t.steps.length - 1) show(t, t.index + 1);
    else end(t.state);
  } else if (key === "ArrowLeft") {
    ev?.preventDefault?.();
    if (t.index > 0) show(t, t.index - 1);
  } else if (key === "Escape") {
    ev?.preventDefault?.();
    end(t.state);
  }
}

function collectSteps(el, doc) {
  const found = [...(el.querySelectorAll?.("[data-amarra-tour-step]") ?? [])];
  return found
    .filter((node) => isVisible(node, doc))
    .map((node) => ({ el: node, copy: copyFor(node) }));
}

function isVisible(node, doc) {
  const rect = node?.getBoundingClientRect?.();
  if (!rect) return false;
  if (rect.width <= MIN_VISIBLE || rect.height <= MIN_VISIBLE) return false;
  const style = doc?.defaultView?.getComputedStyle?.(node);
  if (style && (style.visibility === "hidden" || style.display === "none")) return false;
  return true;
}

function copyFor(node) {
  return {
    title: node.getAttribute?.("data-amarra-tour-title") ?? "",
    text: node.getAttribute?.("data-amarra-tour-text") ?? "",
  };
}

function buildUI(doc) {
  const block = doc.createElement("div");
  block.setAttribute("data-amarra-tour-block", "");
  block.setAttribute("aria-hidden", "true");
  css(block, { position: "fixed", inset: "0", zIndex: "80" });

  const hole = doc.createElement("div");
  hole.setAttribute("data-amarra-tour-hole", "");
  css(hole, {
    position: "fixed",
    zIndex: "81",
    pointerEvents: "none",
    borderRadius: "0.5rem",
    boxShadow: SHADOW,
    outline: "2px solid currentColor",
    outlineOffset: "2px",
  });

  const tip = doc.createElement("div");
  tip.setAttribute("data-amarra-tour-tip", "");
  tip.setAttribute("role", "dialog");
  tip.setAttribute("aria-live", "polite");
  css(tip, {
    position: "fixed",
    zIndex: "82",
    width: "min(92vw, 23rem)",
    padding: "1rem",
    borderRadius: "0.75rem",
    background: "#0f172a",
    color: "#f8fafc",
    border: "1px solid rgba(148, 163, 184, 0.35)",
    boxShadow: "0 20px 40px -12px rgba(0, 0, 0, 0.5)",
    font: "14px/1.4 system-ui, sans-serif",
  });

  const header = doc.createElement("div");
  css(header, {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    gap: "0.5rem",
  });
  const count = doc.createElement("span");
  count.setAttribute("data-amarra-tour-count", "");
  css(count, { opacity: "0.7", fontSize: "0.75rem" });
  const skipBtn = tipButton(doc, "data-amarra-tour-skip", "pular", {});
  header.appendChild(count);
  header.appendChild(skipBtn);

  const heading = doc.createElement("h3");
  heading.setAttribute("data-amarra-tour-heading", "");
  css(heading, { margin: "0.5rem 0 0", fontSize: "1rem", fontWeight: "700" });

  const body = doc.createElement("p");
  body.setAttribute("data-amarra-tour-body", "");
  css(body, { margin: "0.25rem 0 0", opacity: "0.8" });

  const footer = doc.createElement("div");
  css(footer, {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    gap: "0.5rem",
    marginTop: "0.75rem",
  });
  const prevBtn = tipButton(doc, "data-amarra-tour-prev", "← Anterior", {});
  const nextBtn = tipButton(doc, "data-amarra-tour-next", "Próximo →", {
    background: "#38bdf8",
    color: "#0f172a",
    borderColor: "transparent",
  });
  footer.appendChild(prevBtn);
  footer.appendChild(nextBtn);

  tip.appendChild(header);
  tip.appendChild(heading);
  tip.appendChild(body);
  tip.appendChild(footer);

  doc.body?.appendChild(block);
  doc.body?.appendChild(hole);
  doc.body?.appendChild(tip);

  return { block, hole, tip, count, heading, body, prevBtn, nextBtn, skipBtn };
}

function tipButton(doc, attr, label, extra) {
  const btn = doc.createElement("button");
  btn.setAttribute("type", "button");
  btn.setAttribute(attr, "");
  btn.textContent = label;
  css(btn, {
    cursor: "pointer",
    padding: "0.375rem 0.75rem",
    borderRadius: "0.5rem",
    border: "1px solid rgba(148, 163, 184, 0.5)",
    background: "transparent",
    color: "inherit",
    font: "inherit",
    fontWeight: "600",
    ...extra,
  });
  return btn;
}

function css(node, props) {
  if (!node) return;
  if (!node.style) node.style = {};
  for (const [key, value] of Object.entries(props)) node.style[key] = value;
}
