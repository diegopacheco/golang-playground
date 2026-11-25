export function wordBeforeCaret(text, caret) {
  const before = text.slice(0, caret);
  const m = /([A-Za-z_][A-Za-z0-9_]*\.)?([A-Za-z_][A-Za-z0-9_]*)?$/.exec(before);
  const pkg = m[1] ?? "";
  const word = m[2] ?? "";
  return { query: pkg + word, replaceFrom: caret - word.length };
}

export function insertionFor(item) {
  if (item.insert) return { text: item.insert, caretBack: 0 };
  if (item.kind === "func" || (item.kind === "builtin" && /^(append|cap|clear|close|copy|delete|len|make|max|min|new|panic|print|println|recover)$/.test(item.label))) {
    return { text: item.label + "()", caretBack: 1 };
  }
  return { text: item.label, caretBack: 0 };
}

export function caretPosition(text, caret, tabSize = 4) {
  const lines = text.slice(0, caret).split("\n");
  const last = lines[lines.length - 1];
  let col = 0;
  for (const ch of last) col += ch === "\t" ? tabSize - (col % tabSize) : 1;
  return { line: lines.length - 1, col };
}

export function indentFor(text, caret) {
  const lineStart = text.lastIndexOf("\n", caret - 1) + 1;
  const line = text.slice(lineStart, caret);
  const indent = /^[\t ]*/.exec(line)[0];
  return /[{(\[:]\s*$/.test(line) ? indent + "\t" : indent;
}
