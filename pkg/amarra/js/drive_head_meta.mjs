// #403: after a Drive morph, copy a closed set of document meta from the
// response <head>. Stylesheets, preloads, and arbitrary scripts stay on the
// layout — crawlers already got the first load.

export function applyDocumentMeta(doc, html) {
  const head = doc?.head;
  if (!head || typeof doc.createElement !== "function") return;
  if (!/<head\b/i.test(String(html ?? ""))) return;
  const next = parseDocumentMeta(html);
  syncSingle(doc, head, "link", isCanonical, next.canonical);
  syncKeyed(doc, head, "link", "hreflang", isAlternate, next.alternates, (el, tag) => {
    el.setAttribute("rel", "alternate");
    el.setAttribute("hreflang", tag.attrs.hreflang);
    el.setAttribute("href", tag.attrs.href ?? "");
  });
  syncSingle(doc, head, "meta", isDescription, next.description);
  syncKeyed(doc, head, "meta", "property", isOG, next.og, (el, tag) => {
    el.setAttribute("property", tag.attrs.property);
    el.setAttribute("content", tag.attrs.content ?? "");
  });
  syncKeyed(doc, head, "meta", "name", isTwitter, next.twitter, (el, tag) => {
    el.setAttribute("name", tag.attrs.name);
    el.setAttribute("content", tag.attrs.content ?? "");
  });
  replaceJSONLD(doc, head, next.jsonld);
}

function parseDocumentMeta(html) {
  const inner = extractHeadInner(html);
  const canonical = [];
  const alternates = [];
  const description = [];
  const og = [];
  const twitter = [];
  const jsonld = [];
  for (const tag of parseVoidTags(inner)) {
    if (tag.name === "link") {
      const rel = (tag.attrs.rel || "").toLowerCase();
      if (rel === "canonical") canonical.push(tag);
      else if (rel === "alternate" && tag.attrs.hreflang) alternates.push(tag);
    } else if (tag.name === "meta") {
      const name = tag.attrs.name || "";
      const prop = tag.attrs.property || "";
      if (name.toLowerCase() === "description") description.push(tag);
      else if (prop.toLowerCase().startsWith("og:")) og.push(tag);
      else if (name.toLowerCase().startsWith("twitter:")) twitter.push(tag);
    }
  }
  for (const tag of parseJSONLD(inner)) jsonld.push(tag);
  return {
    canonical: canonical[0] ?? null,
    alternates,
    description: description[0] ?? null,
    og,
    twitter,
    jsonld,
  };
}

function extractHeadInner(html) {
  const m = String(html ?? "").match(/<head\b[^>]*>([\s\S]*?)<\/head>/i);
  return m ? m[1] : "";
}

function parseVoidTags(inner) {
  const out = [];
  const re = /<(link|meta)\b([^>]*)\/?>/gi;
  let m;
  while ((m = re.exec(inner))) {
    out.push({ name: m[1].toLowerCase(), attrs: parseAttrs(m[2]) });
  }
  return out;
}

function parseJSONLD(inner) {
  const out = [];
  const re = /<script\b([^>]*)>([\s\S]*?)<\/script>/gi;
  let m;
  while ((m = re.exec(inner))) {
    const attrs = parseAttrs(m[1]);
    if ((attrs.type || "").toLowerCase() !== "application/ld+json") continue;
    out.push({ name: "script", attrs, text: m[2] });
  }
  return out;
}

function parseAttrs(raw) {
  const attrs = {};
  const re = /([^\s=]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+)))?/g;
  let m;
  while ((m = re.exec(raw))) {
    const name = m[1].toLowerCase();
    // Void tags in the layout are written `<meta … />`; the trailing slash
    // is not an attribute. Chromium throws InvalidCharacterError on "/".
    if (!/^[a-z_:][\w:.-]*$/.test(name)) continue;
    attrs[name] = m[2] ?? m[3] ?? m[4] ?? "";
  }
  return attrs;
}

function eachHead(head, tag, pred) {
  const nodes = [...(head.querySelectorAll?.(tag) ?? head.childNodes ?? [])];
  return nodes.filter((el) => {
    if ((el.tagName || "").toLowerCase() !== tag) return false;
    return !pred || pred(el);
  });
}

function attr(el, name) {
  return el.getAttribute?.(name) ?? "";
}

function isCanonical(el) {
  return attr(el, "rel").toLowerCase() === "canonical";
}

function isAlternate(el) {
  return attr(el, "rel").toLowerCase() === "alternate" && !!attr(el, "hreflang");
}

function isDescription(el) {
  return attr(el, "name").toLowerCase() === "description";
}

function isOG(el) {
  return attr(el, "property").toLowerCase().startsWith("og:");
}

function isTwitter(el) {
  return attr(el, "name").toLowerCase().startsWith("twitter:");
}

function syncSingle(doc, head, tagName, pred, next) {
  const existing = eachHead(head, tagName, pred);
  if (!next) {
    for (const el of existing) el.remove?.();
    return;
  }
  const el = existing[0] ?? createAndAppend(doc, head, tagName);
  for (const [k, v] of Object.entries(next.attrs)) el.setAttribute(k, v);
  for (const extra of existing.slice(1)) extra.remove?.();
}

function syncKeyed(doc, head, tagName, keyAttr, pred, nextList, write) {
  const existing = eachHead(head, tagName, pred);
  const nextBy = new Map(nextList.map((t) => [(t.attrs[keyAttr] || "").toLowerCase(), t]));
  const seen = new Set();
  for (const el of existing) {
    const key = attr(el, keyAttr).toLowerCase();
    const nxt = nextBy.get(key);
    if (!nxt) {
      el.remove?.();
      continue;
    }
    write(el, nxt);
    seen.add(key);
  }
  for (const [key, tag] of nextBy) {
    if (seen.has(key)) continue;
    write(createAndAppend(doc, head, tagName), tag);
  }
}

function replaceJSONLD(doc, head, next) {
  for (const el of eachHead(
    head,
    "script",
    (el) => attr(el, "type").toLowerCase() === "application/ld+json"
  )) {
    el.remove?.();
  }
  for (const tag of next) {
    const el = createAndAppend(doc, head, "script");
    el.setAttribute("type", "application/ld+json");
    el.textContent = tag.text ?? "";
  }
}

function createAndAppend(doc, head, tagName) {
  const el = doc.createElement(tagName);
  head.appendChild(el);
  return el;
}
