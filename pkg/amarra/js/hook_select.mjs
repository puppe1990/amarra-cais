const STATE = "_amarraSelectSearch";
const SR_CLASS = "cais-select-search-native";
let listSeq = 0;

// #325: search-by-default for selects. Progressive enhancement keeps the real
// <select> as the source of truth (form submit, name/value/required, iOS) and
// layers a trigger + filtered listbox on top. Coarse pointers keep the native
// picker so phones do not lose the OS sheet. Opt out with
// data-amarra-select-search="false".
export function makeSelectSearch(deps = {}) {
  const isCoarse = deps.isCoarse ?? defaultIsCoarse;
  return {
    connect(el) {
      mount(el, isCoarse);
    },
    updated(el) {
      unmount(el);
      mount(el, isCoarse);
    },
    disconnect(el) {
      unmount(el);
    },
  };
}

export const select = makeSelectSearch();

function defaultIsCoarse(el) {
  const win = el?.ownerDocument?.defaultView;
  // Headless Chrome reports "(hover: none)"; require a real touch point too so
  // the enhancement still runs on desktop and only phones keep the OS picker.
  const coarse = win?.matchMedia?.("(pointer: coarse)")?.matches;
  return !!coarse && (win?.navigator?.maxTouchPoints ?? 0) > 0;
}

function mount(selectEl, isCoarse) {
  if (!isEnhanceable(selectEl, isCoarse)) return;
  const doc = selectEl.ownerDocument;
  const state = buildUI(selectEl, doc);
  bindEvents(selectEl, state);
  selectEl.classList.add(SR_CLASS);
  selectEl[STATE] = state;
  syncLabel(selectEl, state);
}

function isEnhanceable(selectEl, isCoarse) {
  if (!selectEl || selectEl.tagName !== "SELECT" || selectEl.multiple) return false;
  if (selectEl[STATE]) return false;
  if (!selectEl.parentNode || !selectEl.ownerDocument) return false;
  if (selectEl.getAttribute?.("data-amarra-select-search") === "false") return false;
  return !isCoarse(selectEl);
}

function buildUI(selectEl, doc) {
  const wrapper = doc.createElement("div");
  wrapper.classList.add("cais-select-search");

  const trigger = doc.createElement("button");
  trigger.setAttribute("type", "button");
  trigger.setAttribute("aria-haspopup", "listbox");
  trigger.setAttribute("aria-expanded", "false");
  trigger.classList.add("cais-select-search-trigger");

  const label = doc.createElement("span");
  label.classList.add("cais-select-search-label");
  const chevron = doc.createElement("span");
  chevron.classList.add("cais-select-search-chevron");
  chevron.setAttribute("aria-hidden", "true");
  chevron.textContent = "▾";
  trigger.appendChild(label);
  trigger.appendChild(chevron);

  const panel = doc.createElement("div");
  panel.classList.add("cais-select-search-panel");
  panel.hidden = true;

  const listId = `amarra-select-list-${++listSeq}`;
  const list = doc.createElement("ul");
  list.setAttribute("role", "listbox");
  list.setAttribute("id", listId);
  list.classList.add("cais-select-search-list");
  trigger.setAttribute("aria-controls", listId);

  const input = doc.createElement("input");
  input.setAttribute("type", "text");
  input.setAttribute("autocomplete", "off");
  input.setAttribute("aria-label", searchPlaceholder(selectEl, doc));
  input.setAttribute("placeholder", searchPlaceholder(selectEl, doc));
  input.classList.add("cais-select-search-input");

  const rows = Array.from(selectEl.options ?? []).map((option, index) => {
    const row = doc.createElement("li");
    row.setAttribute("role", "option");
    row.setAttribute("data-value", option.value ?? "");
    row.setAttribute("data-index", String(index));
    row.classList.add("cais-select-search-option");
    if (option.disabled) row.setAttribute("aria-disabled", "true");
    row.textContent = option.textContent ?? "";
    list.appendChild(row);
    return row;
  });

  panel.appendChild(input);
  panel.appendChild(list);
  wrapper.appendChild(trigger);
  wrapper.appendChild(panel);
  selectEl.parentNode.insertBefore(wrapper, selectEl.nextSibling);
  return { doc, wrapper, trigger, panel, input, list, rows, label };
}

function bindEvents(selectEl, state) {
  const { trigger, input, panel, wrapper, rows } = state;
  const doc = state.doc;

  state.onTriggerClick = (ev) => {
    ev?.preventDefault?.();
    if (panel.hidden) open(selectEl, state);
    else close(state);
  };
  state.onTriggerKey = (ev) => {
    if (ev?.key === "ArrowDown" || ev?.key === "Enter" || ev?.key === " ") {
      ev?.preventDefault?.();
      open(selectEl, state);
    }
  };
  state.onInput = () => filter(selectEl, state);
  state.onInputKey = (ev) => inputKey(selectEl, state, ev);
  state.onChange = () => syncLabel(selectEl, state);
  state.onDocClick = (ev) => {
    if (panel.hidden) return;
    if (wrapper.contains?.(ev?.target)) return;
    close(state);
  };
  state.rowClicks = rows.map((row) => {
    const fn = (ev) => {
      ev?.preventDefault?.();
      choose(selectEl, state, row);
    };
    row.addEventListener("click", fn);
    return fn;
  });

  trigger.addEventListener("click", state.onTriggerClick);
  trigger.addEventListener("keydown", state.onTriggerKey);
  input.addEventListener("input", state.onInput);
  input.addEventListener("keydown", state.onInputKey);
  selectEl.addEventListener("change", state.onChange);
  doc.addEventListener("click", state.onDocClick, true);
}

function open(selectEl, state) {
  state.panel.hidden = false;
  state.trigger.setAttribute("aria-expanded", "true");
  state.input.value = "";
  filter(selectEl, state);
  state.input.focus?.();
}

function close(state) {
  state.panel.hidden = true;
  state.trigger.setAttribute("aria-expanded", "false");
}

function choose(selectEl, state, row) {
  selectEl.value = row.getAttribute("data-value") ?? "";
  dispatchChange(selectEl);
  syncLabel(selectEl, state);
  close(state);
  state.trigger.focus?.();
}

function dispatchChange(selectEl) {
  if (typeof selectEl.dispatchEvent === "function" && typeof Event === "function") {
    selectEl.dispatchEvent(new Event("change", { bubbles: true }));
    return;
  }
  selectEl.fire?.("change", { type: "change" });
}

function syncLabel(selectEl, state) {
  const options = Array.from(selectEl.options ?? []);
  const selected = options.find((o) => (o.value ?? "") === (selectEl.value ?? "")) ?? options[0];
  state.label.textContent = selected?.textContent ?? "";
  state.rows.forEach((row, index) => {
    const on = (options[index]?.value ?? "") === (selectEl.value ?? "");
    row.setAttribute("aria-selected", on ? "true" : "false");
    row.classList.toggle("is-selected", on);
  });
}

function filter(selectEl, state) {
  const query = normalize(state.input.value);
  let firstVisible = null;
  state.rows.forEach((row) => {
    const hidden = query !== "" && !normalize(row.textContent).includes(query);
    row.classList.toggle("is-hidden", hidden);
    if (!hidden && firstVisible === null) firstVisible = row;
  });
  highlight(state, firstVisible);
}

function highlight(state, row) {
  state.rows.forEach((r) => r.classList.toggle("is-highlighted", r === row));
  state.highlighted = row ?? null;
  row?.scrollIntoView?.({ block: "nearest" });
}

function moveHighlight(state, delta) {
  const visible = state.rows.filter((r) => !r.classList.contains("is-hidden"));
  if (!visible.length) return;
  let current = visible.indexOf(state.highlighted);
  if (current < 0) current = delta > 0 ? -1 : visible.length;
  const next = Math.max(0, Math.min(visible.length - 1, current + delta));
  highlight(state, visible[next]);
}

function inputKey(selectEl, state, ev) {
  const key = ev?.key;
  if (key === "ArrowDown") {
    ev?.preventDefault?.();
    moveHighlight(state, 1);
  } else if (key === "ArrowUp") {
    ev?.preventDefault?.();
    moveHighlight(state, -1);
  } else if (key === "Enter") {
    ev?.preventDefault?.();
    if (state.highlighted) choose(selectEl, state, state.highlighted);
  } else if (key === "Escape") {
    ev?.preventDefault?.();
    close(state);
    state.trigger.focus?.();
  } else if (key === "Tab") {
    close(state);
  }
}

function unmount(selectEl) {
  const state = selectEl?.[STATE];
  if (!state) return;
  const { wrapper, trigger, input, rows } = state;
  trigger.removeEventListener?.("click", state.onTriggerClick);
  trigger.removeEventListener?.("keydown", state.onTriggerKey);
  input.removeEventListener?.("input", state.onInput);
  input.removeEventListener?.("keydown", state.onInputKey);
  selectEl.removeEventListener?.("change", state.onChange);
  state.doc?.removeEventListener?.("click", state.onDocClick, true);
  rows.forEach((row, index) => row.removeEventListener?.("click", state.rowClicks[index]));
  wrapper.parentNode?.removeChild?.(wrapper);
  selectEl.classList.remove(SR_CLASS);
  delete selectEl[STATE];
}

function searchPlaceholder(selectEl, doc) {
  const attr = selectEl.getAttribute?.("data-amarra-select-search-placeholder");
  if (attr) return attr;
  const lang = doc?.documentElement?.lang ?? "";
  return String(lang).toLowerCase().startsWith("pt") ? "Buscar…" : "Search…";
}

function normalize(text) {
  return String(text ?? "")
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim();
}
