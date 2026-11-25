const KEYWORDS = new Set(["break","case","chan","const","continue","default","defer","else","fallthrough","for","func","go","goto","if","import","interface","map","package","range","return","select","struct","switch","type","var"]);
const TYPES = new Set(["bool","byte","error","float32","float64","int","int8","int16","int32","int64","rune","string","uint","uint8","uint16","uint32","uint64","uintptr","any","comparable","complex64","complex128"]);
const BUILTINS = new Set(["append","cap","clear","close","complex","copy","delete","imag","len","make","max","min","new","panic","print","println","real","recover"]);
const CONSTS = new Set(["true","false","nil","iota"]);

const RULES = [
  ["com", /^\/\/[^\n]*/],
  ["com", /^\/\*[\s\S]*?(\*\/|$)/],
  ["str", /^"(\\.|[^"\\\n])*"?/],
  ["str", /^`[^`]*`?/],
  ["str", /^'(\\.|[^'\\\n])*'?/],
  ["num", /^(0[xX][0-9a-fA-F_]+|0[bB][01_]+|0[oO][0-7_]+|\d[\d_]*(\.\d*)?([eE][+-]?\d+)?|\.\d+)i?/],
  ["chan", /^<-/],
  ["word", /^[A-Za-z_][A-Za-z0-9_]*/],
  ["op", /^(:=|==|!=|<=|>=|&&|\|\||\+\+|--|\.\.\.|[-+*/%&|^<>=!])/],
  ["ws", /^\s+/],
  ["pun", /^./],
];

export function escapeHTML(s) {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

export function tokenize(src) {
  const tokens = [];
  let rest = src;
  while (rest.length > 0) {
    for (const [kind, re] of RULES) {
      const m = re.exec(rest);
      if (!m || m[0] === "") continue;
      tokens.push({ kind: kind === "word" ? classifyWord(m[0], rest.slice(m[0].length), tokens) : kind, text: m[0] });
      rest = rest.slice(m[0].length);
      break;
    }
  }
  return tokens;
}

function classifyWord(word, after, tokens) {
  if (KEYWORDS.has(word)) return "kw";
  if (TYPES.has(word)) return "ty";
  if (CONSTS.has(word)) return "con";
  const prev = tokens.length > 0 ? tokens[tokens.length - 1].text : "";
  if (after.startsWith("(")) return BUILTINS.has(word) && prev !== "." ? "bi" : "fn";
  if (after.startsWith(".") && prev !== ".") return "pkg";
  if (/^[A-Z]/.test(word)) return "exp";
  return "id";
}

export function highlight(src) {
  return tokenize(src)
    .map((t) => (t.kind === "ws" || t.kind === "pun" || t.kind === "id" ? escapeHTML(t.text) : `<span class="t-${t.kind}">${escapeHTML(t.text)}</span>`))
    .join("");
}
