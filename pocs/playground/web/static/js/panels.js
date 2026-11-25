import { icon } from "./icons.js";

const STORE = "hiddenPanels";

function hiddenList() {
  try {
    return JSON.parse(localStorage.getItem(STORE) ?? "[]");
  } catch {
    return [];
  }
}

function saveHidden(list) {
  localStorage.setItem(STORE, JSON.stringify(list));
}

export function setupPanels() {
  document.querySelectorAll(".panel").forEach((panel) => {
    const head = panel.querySelector(".panel-head");
    const tools = document.createElement("div");
    tools.className = "panel-tools";
    tools.innerHTML = `
      <button class="icon-btn expand-btn" data-tip="Expand panel" aria-label="Expand panel">${icon("expand", 16)}</button>
      <button class="icon-btn hide-btn" data-tip="Hide panel" aria-label="Hide panel">${icon("hide", 16)}</button>`;
    head.appendChild(tools);
    tools.querySelector(".expand-btn").addEventListener("click", () => toggleExpand(panel));
    tools.querySelector(".hide-btn").addEventListener("click", () => hidePanel(panel));
  });
  for (const entry of hiddenList()) {
    const panel = document.querySelector(`.panel[data-panel="${entry.id}"]`);
    if (panel) dock(panel, entry.side, false);
  }
  document.addEventListener("keydown", (e) => {
    if (e.key !== "Escape") return;
    const expanded = document.querySelector(".panel.expanded");
    if (expanded && !document.querySelector(".modal:not([hidden])")) toggleExpand(expanded);
  });
}

export function toggleExpand(panel) {
  const on = panel.classList.toggle("expanded");
  document.body.classList.toggle("has-expanded", on);
  const btn = panel.querySelector(".expand-btn");
  btn.innerHTML = icon(on ? "collapse" : "expand", 16);
  btn.dataset.tip = on ? "Collapse panel  Esc" : "Expand panel";
  btn.setAttribute("aria-label", btn.dataset.tip);
}

function hidePanel(panel) {
  if (panel.classList.contains("expanded")) toggleExpand(panel);
  const r = panel.getBoundingClientRect();
  const side = r.left + r.width / 2 < window.innerWidth / 2 ? "left" : "right";
  dock(panel, side, true);
}

function dock(panel, side, animate) {
  const id = panel.dataset.panel;
  const finish = () => {
    panel.classList.remove("sliding-left", "sliding-right");
    panel.classList.add("docked");
    addTab(panel, side);
  };
  if (animate) {
    panel.classList.add(side === "left" ? "sliding-left" : "sliding-right");
    setTimeout(finish, 200);
  } else {
    finish();
  }
  const list = hiddenList().filter((e) => e.id !== id);
  list.push({ id, side });
  saveHidden(list);
}

function addTab(panel, side) {
  const container = document.querySelector(`.dock-${side}`);
  const tab = document.createElement("button");
  tab.className = "dock-tab";
  tab.dataset.view = panel.closest(".view").dataset.view;
  tab.dataset.tip = "Restore " + panel.dataset.title;
  tab.setAttribute("aria-label", tab.dataset.tip);
  tab.innerHTML = `${icon(side === "left" ? "arrowRight" : "arrowLeft", 14)}<span>${panel.dataset.title}</span>`;
  tab.addEventListener("click", () => {
    tab.remove();
    panel.classList.remove("docked");
    saveHidden(hiddenList().filter((e) => e.id !== panel.dataset.panel));
  });
  container.appendChild(tab);
}
