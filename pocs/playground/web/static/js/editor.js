import { highlight, escapeHTML } from "./highlight.js";
import { wordBeforeCaret, insertionFor, caretPosition, indentFor } from "./completion.js";
import { api } from "./api.js";

export class Editor {
  constructor(root, { onRun }) {
    this.root = root;
    this.onRun = onRun;
    this.items = [];
    this.active = 0;
    this.errorLines = new Set();
    this.requestId = 0;
    root.innerHTML = `
      <div class="gutter" aria-hidden="true"><div class="gutter-inner"></div></div>
      <div class="code-area">
        <pre class="hl" aria-hidden="true"><code></code></pre>
        <textarea spellcheck="false" autocapitalize="off" autocomplete="off" autocorrect="off" wrap="off" aria-label="Go code editor"></textarea>
      </div>
      <ul class="completion" role="listbox" hidden></ul>
      <span class="measure">M</span>`;
    this.ta = root.querySelector("textarea");
    this.code = root.querySelector("pre code");
    this.pre = root.querySelector("pre");
    this.gutter = root.querySelector(".gutter-inner");
    this.popup = root.querySelector(".completion");
    this.measure = root.querySelector(".measure");
    this.ta.addEventListener("input", () => this.onInput());
    this.ta.addEventListener("scroll", () => this.syncScroll());
    this.ta.addEventListener("keydown", (e) => this.onKey(e));
    this.ta.addEventListener("blur", () => setTimeout(() => this.close(), 150));
    this.ta.addEventListener("click", () => this.close());
    this.popup.addEventListener("mousedown", (e) => {
      const li = e.target.closest("li");
      if (!li) return;
      e.preventDefault();
      this.accept(Number(li.dataset.index));
    });
  }

  get value() {
    return this.ta.value;
  }

  set value(v) {
    this.ta.value = v;
    this.errorLines.clear();
    this.render();
    this.ta.scrollTop = 0;
    this.syncScroll();
  }

  focus() {
    this.ta.focus();
  }

  markErrors(stderr) {
    this.errorLines = new Set([...stderr.matchAll(/main\.go:(\d+)/g)].map((m) => Number(m[1])));
    this.renderGutter();
  }

  render() {
    this.code.innerHTML = highlight(this.ta.value) + "\n";
    this.renderGutter();
    localStorage.setItem("code", this.ta.value);
  }

  renderGutter() {
    const count = this.ta.value.split("\n").length;
    const caretLine = caretPosition(this.ta.value, this.ta.selectionStart).line + 1;
    let html = "";
    for (let i = 1; i <= count; i++) {
      const cls = this.errorLines.has(i) ? "ln err" : i === caretLine ? "ln cur" : "ln";
      html += `<div class="${cls}">${i}</div>`;
    }
    this.gutter.innerHTML = html;
  }

  syncScroll() {
    this.pre.scrollTop = this.ta.scrollTop;
    this.pre.scrollLeft = this.ta.scrollLeft;
    this.gutter.style.transform = `translateY(${-this.ta.scrollTop}px)`;
    this.close();
  }

  onInput() {
    this.render();
    const ch = this.ta.value[this.ta.selectionStart - 1] ?? "";
    if (/[A-Za-z0-9_.]/.test(ch)) this.suggest(false);
    else this.close();
  }

  insert(text) {
    if (!document.execCommand("insertText", false, text)) {
      this.ta.setRangeText(text, this.ta.selectionStart, this.ta.selectionEnd, "end");
      this.render();
    }
  }

  onKey(e) {
    const open = !this.popup.hidden;
    if (open && (e.key === "ArrowDown" || e.key === "ArrowUp")) {
      e.preventDefault();
      this.move(e.key === "ArrowDown" ? 1 : -1);
      return;
    }
    if (open && (e.key === "Enter" || e.key === "Tab") && !e.metaKey) {
      e.preventDefault();
      this.accept(this.active);
      return;
    }
    if (open && e.key === "Escape") {
      e.preventDefault();
      e.stopPropagation();
      this.close();
      return;
    }
    if (e.key === " " && e.ctrlKey) {
      e.preventDefault();
      this.suggest(true);
      return;
    }
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      this.onRun();
      return;
    }
    if (e.key === "Tab" && !e.metaKey) {
      e.preventDefault();
      this.insert("\t");
      return;
    }
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      const v = this.ta.value;
      const pos = this.ta.selectionStart;
      const indent = indentFor(v, pos);
      if (v[pos - 1] === "{" && v[pos] === "}") {
        this.insert("\n" + indent + "\n" + indent.slice(0, -1));
        const back = indent.length;
        this.ta.selectionStart = this.ta.selectionEnd = this.ta.selectionStart - back;
      } else {
        this.insert("\n" + indent);
      }
      return;
    }
    if (e.key === "{" || e.key === "(" || e.key === "[") {
      const pair = { "{": "}", "(": ")", "[": "]" }[e.key];
      if (this.ta.selectionStart === this.ta.selectionEnd) {
        e.preventDefault();
        this.insert(e.key + pair);
        this.ta.selectionStart = this.ta.selectionEnd = this.ta.selectionStart - 1;
      }
      return;
    }
    if ((e.key === "}" || e.key === ")" || e.key === "]") && this.ta.value[this.ta.selectionStart] === e.key) {
      e.preventDefault();
      this.ta.selectionStart = this.ta.selectionEnd = this.ta.selectionStart + 1;
    }
    requestAnimationFrame(() => this.renderGutter());
  }

  async suggest(force) {
    const { query } = wordBeforeCaret(this.ta.value, this.ta.selectionStart);
    if (!force && query.replace(/.*\./, "").length < 1 && !query.endsWith(".")) {
      this.close();
      return;
    }
    const id = ++this.requestId;
    const items = await api.complete(query);
    if (id !== this.requestId) return;
    const word = query.replace(/.*\./, "");
    this.items = items.filter((it) => !(it.label === word && items.length === 1));
    if (this.items.length === 0) {
      this.close();
      return;
    }
    this.active = 0;
    this.renderPopup();
  }

  renderPopup() {
    this.popup.innerHTML = this.items
      .map((it, i) => `<li role="option" data-index="${i}" class="${i === this.active ? "active" : ""}"><span class="kind k-${it.kind}">${it.kind[0]}</span><span class="label">${escapeHTML(it.label)}</span><span class="detail">${escapeHTML(it.detail)}</span></li>`)
      .join("");
    const { line, col } = caretPosition(this.ta.value, this.ta.selectionStart);
    const cw = this.measure.getBoundingClientRect().width;
    const lh = parseFloat(getComputedStyle(this.ta).lineHeight);
    const pad = parseFloat(getComputedStyle(this.ta).paddingLeft);
    const area = this.ta.getBoundingClientRect();
    const host = this.root.getBoundingClientRect();
    const x = area.left - host.left + pad + col * cw - this.ta.scrollLeft;
    const y = area.top - host.top + pad + (line + 1) * lh - this.ta.scrollTop + 2;
    this.popup.hidden = false;
    const maxX = host.width - this.popup.offsetWidth - 8;
    this.popup.style.left = Math.max(4, Math.min(x, maxX)) + "px";
    const below = y + this.popup.offsetHeight < host.height;
    this.popup.style.top = (below ? y : Math.max(4, y - lh - this.popup.offsetHeight - 4)) + "px";
  }

  move(delta) {
    this.active = (this.active + delta + this.items.length) % this.items.length;
    this.renderPopup();
    this.popup.querySelector(".active")?.scrollIntoView({ block: "nearest" });
  }

  accept(index) {
    const item = this.items[index];
    if (!item) return;
    const { replaceFrom } = wordBeforeCaret(this.ta.value, this.ta.selectionStart);
    this.ta.selectionStart = replaceFrom;
    const { text, caretBack } = insertionFor(item);
    this.insert(text);
    this.ta.selectionStart = this.ta.selectionEnd = this.ta.selectionStart - caretBack;
    this.close();
    this.render();
  }

  close() {
    this.popup.hidden = true;
    this.requestId++;
  }
}
