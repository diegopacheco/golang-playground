import { api } from "./api.js";
import { Editor } from "./editor.js";
import { highlight, escapeHTML } from "./highlight.js";
import { icon } from "./icons.js";
import { setupPanels } from "./panels.js";
import { TABS, SHORTCUTS, shortcutById, keysLabel, filterShortcuts } from "./shortcuts.js";

const DEFAULT_CODE = `package main

import (
	"fmt"
	"sync"
)

func main() {
	results := make(chan string)
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			results <- fmt.Sprintf("goroutine %d says hi", i)
		})
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for r := range results {
		fmt.Println(r)
	}
}
`;

const desktop = window.desktop ?? null;
const state = { docs: [], samples: [], currentDoc: null, currentSample: null, zoom: Number(localStorage.getItem("zoom") ?? 1) };
const $ = (sel) => document.querySelector(sel);

const editor = new Editor($("#editor"), { onRun: run });
editor.value = localStorage.getItem("code") ?? DEFAULT_CODE;

function toast(msg) {
  const t = $("#toast");
  t.textContent = msg;
  t.hidden = false;
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => (t.hidden = true), 2400);
}

function renderTabs() {
  $("#tabs").innerHTML = TABS.map((t) => `<button class="tab" role="tab" data-tab="${t.id}" data-sc="tab-${t.id}">${icon(t.icon)}<span>${t.title}</span></button>`).join("");
  $("#tabs").addEventListener("click", (e) => {
    const b = e.target.closest(".tab");
    if (b) showTab(b.dataset.tab);
  });
}

function showTab(id) {
  document.body.dataset.tab = id;
  document.querySelectorAll(".tab").forEach((b) => b.setAttribute("aria-selected", String(b.dataset.tab === id)));
  document.querySelectorAll(".view").forEach((v) => (v.hidden = v.dataset.view !== id));
  localStorage.setItem("tab", id);
  if (id === "playground") editor.focus();
}

async function run() {
  showTab("playground");
  const out = $("#output");
  const status = $("#run-status");
  status.className = "status running";
  status.textContent = "running";
  out.innerHTML = `<div class="muted">Building and running...</div>`;
  try {
    const res = await api.run(editor.value);
    editor.markErrors(res.stderr);
    status.className = "status " + (res.timedOut ? "timeout" : res.exitCode === 0 ? "ok" : "fail");
    status.textContent = res.timedOut ? "timed out" : `exit ${res.exitCode} · ${res.durationMs} ms`;
    out.innerHTML = (res.stdout ? `<pre class="stdout">${escapeHTML(res.stdout)}</pre>` : "") + (res.stderr ? `<pre class="stderr">${escapeHTML(res.stderr)}</pre>` : "") || `<div class="muted">Program exited with no output.</div>`;
  } catch (err) {
    status.className = "status fail";
    status.textContent = "error";
    out.innerHTML = `<pre class="stderr">${escapeHTML(err.message)}</pre>`;
  }
}

async function format() {
  try {
    const res = await api.format(editor.value);
    editor.value = res.code;
    toast("Formatted with gofmt");
  } catch (err) {
    toast("Format failed: " + err.message);
  }
}

function clearOutput() {
  $("#output").innerHTML = `<div class="muted">Press Run or ⌘ Enter to execute your program.</div>`;
  $("#run-status").className = "status";
  $("#run-status").textContent = "idle";
}

function openInPlayground(code, runNow) {
  editor.value = code;
  showTab("playground");
  if (runNow) run();
}

function decorateCode(root) {
  root.querySelectorAll("pre.code").forEach((pre) => {
    const src = pre.textContent;
    const lang = pre.dataset.lang;
    pre.querySelector("code").innerHTML = lang === "go" ? highlight(src) : escapeHTML(src);
    const bar = document.createElement("div");
    bar.className = "code-bar";
    const runnable = lang === "go" && src.includes("package main");
    bar.innerHTML = `<span class="lang">${escapeHTML(lang)}</span><button class="mini copy">${icon("copy", 14)}Copy</button>${runnable ? `<button class="mini run">${icon("play", 14)}Run in Playground</button>` : ""}`;
    bar.querySelector(".copy").addEventListener("click", async () => {
      await navigator.clipboard.writeText(src);
      toast("Copied to clipboard");
    });
    bar.querySelector(".run")?.addEventListener("click", () => openInPlayground(src, true));
    pre.before(bar);
  });
}

async function loadDocs() {
  state.docs = await api.docs();
  $("#doc-list").innerHTML = state.docs
    .map((d, i) => `<li><button class="list-item" data-doc="${d.id}"><span class="num">${String(i + 1).padStart(2, "0")}</span><span><strong>${escapeHTML(d.title)}</strong><small>${escapeHTML(d.summary)}</small></span></button></li>`)
    .join("");
  $("#doc-list").addEventListener("click", (e) => {
    const b = e.target.closest("[data-doc]");
    if (b) openDoc(b.dataset.doc);
  });
  await openDoc(localStorage.getItem("doc") ?? state.docs[0]?.id);
}

async function openDoc(id) {
  if (!id) return;
  const doc = await api.doc(id).catch(() => api.doc(state.docs[0].id));
  state.currentDoc = doc.id;
  localStorage.setItem("doc", doc.id);
  document.querySelectorAll("[data-doc]").forEach((b) => b.classList.toggle("active", b.dataset.doc === doc.id));
  $("#doc-title").textContent = doc.title;
  const body = $("#doc-body");
  body.innerHTML = doc.html;
  decorateCode(body);
  body.scrollTop = 0;
}

async function loadSamples() {
  state.samples = await api.samples();
  $("#sample-list").innerHTML = state.samples
    .map((s) => `<li><button class="list-item" data-sample="${s.id}">${icon("file", 16)}<span><strong>${escapeHTML(s.title)}</strong><small>${s.code.split("\n").length} lines</small></span></button></li>`)
    .join("");
  $("#sample-list").addEventListener("click", (e) => {
    const b = e.target.closest("[data-sample]");
    if (b) openSample(b.dataset.sample);
  });
  openSample(state.samples[0]?.id);
}

function openSample(id) {
  const s = state.samples.find((x) => x.id === id);
  if (!s) return;
  state.currentSample = s;
  document.querySelectorAll("[data-sample]").forEach((b) => b.classList.toggle("active", b.dataset.sample === id));
  $("#sample-title").textContent = s.title;
  const lines = s.code.split("\n").map((_, i) => `<span>${i + 1}</span>`).join("");
  $("#sample-code").innerHTML = `<div class="ro-gutter">${lines}</div><pre><code>${highlight(s.code)}</code></pre>`;
}

function setZoom(z) {
  state.zoom = Math.min(2, Math.max(0.6, Math.round(z * 10) / 10));
  localStorage.setItem("zoom", String(state.zoom));
  if (desktop) desktop.setZoom(state.zoom);
  else document.documentElement.style.zoom = String(state.zoom);
  toast(`Zoom ${Math.round(state.zoom * 100)}%`);
}

async function toggleFullscreen() {
  if (desktop) return desktop.toggleFullscreen();
  if (document.fullscreenElement) await document.exitFullscreen();
  else await document.documentElement.requestFullscreen();
}

async function screenshot() {
  if (!desktop) {
    toast("Screenshots are available in the desktop app");
    return;
  }
  const file = await desktop.screenshot();
  toast("Saved " + file);
}

const ACTIONS = {
  palette: () => openPalette(),
  help: () => openHelp(),
  run,
  format,
  clear: clearOutput,
  fullscreen: toggleFullscreen,
  zoomIn: () => setZoom(state.zoom + 0.1),
  zoomOut: () => setZoom(state.zoom - 0.1),
  zoomReset: () => setZoom(1),
  screenshot,
};

function openModal(id) {
  closeModals();
  const m = $(id);
  m.hidden = false;
  const input = m.querySelector("input");
  input.value = "";
  input.focus();
  return m;
}

function closeModals() {
  document.querySelectorAll(".modal").forEach((m) => (m.hidden = true));
}

function paletteEntries() {
  return [
    ...TABS.map((t) => ({ icon: t.icon, label: t.title, hint: "Tab · ⌘ " + t.key, go: () => showTab(t.id) })),
    ...state.docs.map((d) => ({ icon: "book", label: d.title, hint: "Doc", go: () => (showTab("docs"), openDoc(d.id)) })),
    ...state.samples.map((s) => ({ icon: "code", label: s.title, hint: "Sample", go: () => (showTab("samples"), openSample(s.id)) })),
    ...state.samples.map((s) => ({ icon: "play", label: "Run sample: " + s.title, hint: "Action", go: () => openInPlayground(s.code, true) })),
    ...["run", "format", "clear", "help", "fullscreen", "zoomIn", "zoomOut", "screenshot"].map((id) => {
      const s = shortcutById(id);
      return { icon: s.icon, label: s.desc, hint: keysLabel(s), go: ACTIONS[id] };
    }),
  ];
}

let paletteIndex = 0;
let paletteItems = [];

function openPalette() {
  const m = openModal("#palette");
  const input = m.querySelector("input");
  const render = () => {
    const q = input.value.trim().toLowerCase();
    paletteItems = paletteEntries().filter((e) => q === "" || e.label.toLowerCase().includes(q) || e.hint.toLowerCase().includes(q));
    paletteIndex = Math.min(paletteIndex, Math.max(0, paletteItems.length - 1));
    m.querySelector(".results").innerHTML =
      paletteItems.map((e, i) => `<li class="${i === paletteIndex ? "active" : ""}" data-i="${i}">${icon(e.icon)}<span>${escapeHTML(e.label)}</span><kbd>${escapeHTML(e.hint)}</kbd></li>`).join("") ||
      `<li class="empty">Nothing matches "${escapeHTML(input.value)}"</li>`;
    m.querySelector(".active")?.scrollIntoView({ block: "nearest" });
  };
  paletteIndex = 0;
  input.oninput = () => ((paletteIndex = 0), render());
  input.onkeydown = (e) => {
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      paletteIndex = (paletteIndex + (e.key === "ArrowDown" ? 1 : -1) + paletteItems.length) % Math.max(1, paletteItems.length);
      render();
    } else if (e.key === "Enter" && paletteItems[paletteIndex]) {
      closeModals();
      paletteItems[paletteIndex].go();
    }
  };
  m.querySelector(".results").onclick = (e) => {
    const li = e.target.closest("[data-i]");
    if (!li) return;
    closeModals();
    paletteItems[Number(li.dataset.i)].go();
  };
  render();
}

function openHelp() {
  const m = openModal("#help");
  const input = m.querySelector("input");
  const render = () => {
    const groups = filterShortcuts(input.value);
    const total = groups.reduce((n, g) => n + g.items.length, 0);
    m.querySelector(".count").textContent = `${total} of ${SHORTCUTS.length} shortcuts`;
    m.querySelector(".groups").innerHTML =
      groups
        .map(
          ({ group, items }) => `<section class="sc-group" style="--c:${group.color}"><h3>${icon(group.icon)}${group.title}</h3>${items
            .map((s) => `<div class="sc-row">${icon(s.icon, 16)}<span>${s.desc}</span><span class="keys">${s.keys.map((k) => `<kbd>${k}</kbd>`).join("")}</span></div>`)
            .join("")}</section>`,
        )
        .join("") || `<p class="empty">No shortcut matches "${escapeHTML(input.value)}"</p>`;
  };
  input.oninput = render;
  render();
}

function setupKeys() {
  document.addEventListener(
    "keydown",
    (e) => {
      const mod = e.metaKey || e.ctrlKey;
      const anyModal = document.querySelector(".modal:not([hidden])");
      if (e.key === "Escape" && anyModal) {
        const input = anyModal.querySelector("input");
        e.preventDefault();
        e.stopPropagation();
        if (input.value !== "") {
          input.value = "";
          input.dispatchEvent(new Event("input"));
        } else closeModals();
        return;
      }
      if (!mod) return;
      const k = e.key.toLowerCase();
      const map = {
        k: "palette",
        "/": "help",
        l: "clear",
        p: "screenshot",
        "=": "zoomIn",
        "+": "zoomIn",
        "-": "zoomOut",
        0: "zoomReset",
      };
      let action = map[k];
      if (e.shiftKey && k === "f") action = "format";
      if (e.shiftKey && k === "enter") action = "fullscreen";
      if (!e.shiftKey && k === "enter" && e.target.tagName !== "TEXTAREA") action = "run";
      const tab = TABS.find((t) => t.key === e.key);
      if (tab && !e.shiftKey) {
        e.preventDefault();
        closeModals();
        showTab(tab.id);
        return;
      }
      if (action && ACTIONS[action]) {
        e.preventDefault();
        if (action !== "palette" && action !== "help") closeModals();
        ACTIONS[action]();
      }
    },
    true,
  );
  document.querySelectorAll(".modal").forEach((m) => m.addEventListener("mousedown", (e) => e.target === m && closeModals()));
  document.querySelectorAll("[data-action]").forEach((b) => b.addEventListener("click", () => ACTIONS[b.dataset.action]()));
}

function setupTooltips() {
  const tip = $("#tooltip");
  let timer;
  document.addEventListener("mouseover", (e) => {
    const el = e.target.closest("[data-sc],[data-tip]");
    clearTimeout(timer);
    tip.hidden = true;
    if (!el) return;
    timer = setTimeout(() => {
      const s = el.dataset.sc ? shortcutById(el.dataset.sc) : null;
      tip.innerHTML = s ? `${escapeHTML(s.desc)}<span class="keys">${s.keys.map((k) => `<kbd>${k}</kbd>`).join("")}</span>` : escapeHTML(el.dataset.tip);
      tip.hidden = false;
      const r = el.getBoundingClientRect();
      const x = Math.min(Math.max(6, r.left + r.width / 2 - tip.offsetWidth / 2), window.innerWidth - tip.offsetWidth - 6);
      const y = r.bottom + 8 + tip.offsetHeight > window.innerHeight ? r.top - tip.offsetHeight - 8 : r.bottom + 8;
      tip.style.left = x + "px";
      tip.style.top = y + "px";
    }, 400);
  });
  document.addEventListener("mousedown", () => {
    clearTimeout(timer);
    tip.hidden = true;
  });
}

function setupChrome() {
  document.querySelectorAll("[data-icon]").forEach((el) => (el.innerHTML = icon(el.dataset.icon, Number(el.dataset.size ?? 18)) + el.innerHTML));
  if (desktop) {
    document.body.classList.add("is-desktop");
    $(".titlebar").addEventListener("dblclick", (e) => {
      if (!e.target.closest("button")) desktop.toggleMaximize();
    });
    desktop.setZoom(state.zoom);
  } else if (state.zoom !== 1) {
    document.documentElement.style.zoom = String(state.zoom);
  }
  $("#sample-open").addEventListener("click", () => state.currentSample && openInPlayground(state.currentSample.code, false));
  $("#sample-run").addEventListener("click", () => state.currentSample && openInPlayground(state.currentSample.code, true));
}

renderTabs();
setupChrome();
setupPanels();
setupKeys();
setupTooltips();
clearOutput();
showTab(localStorage.getItem("tab") ?? "playground");
await Promise.all([loadDocs(), loadSamples()]);
