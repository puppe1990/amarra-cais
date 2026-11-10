export function extractMainHTML(html) {
  const str = String(html ?? "");
  const open = extractMainOpen(str);
  if (!open) return null;
  const start = open.index + open[0].length;
  return sliceMatchingClose(str, start, open[1]);
}

export function extractMainTagName(html) {
  return extractMainOpen(String(html ?? ""))?.[1]?.toUpperCase() ?? null;
}

function sliceMatchingClose(str, start, tag) {
  const lower = str.toLowerCase();
  const name = tag.toLowerCase();
  const closeToken = `</${name}>`;
  let depth = 1;
  let i = start;
  while (i < str.length && depth > 0) {
    const nextOpen = findOpenTag(lower, i, name);
    const nextClose = lower.indexOf(closeToken, i);
    if (nextClose === -1) return null;
    if (nextOpen !== -1 && nextOpen < nextClose) {
      depth += 1;
      i = nextOpen + name.length + 1;
      continue;
    }
    depth -= 1;
    if (depth === 0) return str.slice(start, nextClose);
    i = nextClose + closeToken.length;
  }
  return null;
}

function extractMainOpen(str) {
  return str.match(/<([a-zA-Z][\w:-]*)(?=[^>]*\sid\s*=\s*["']amarra-main["'])[^>]*>/i);
}

function findOpenTag(lower, from, name) {
  const token = `<${name}`;
  let i = from;
  while (i < lower.length) {
    const j = lower.indexOf(token, i);
    if (j === -1) return -1;
    const after = lower[j + token.length];
    if (after === ">" || after === "/" || (after && /\s/.test(after))) return j;
    i = j + token.length;
  }
  return -1;
}
