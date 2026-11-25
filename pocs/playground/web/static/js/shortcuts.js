export const TABS = [
  { id: "playground", title: "Playground", icon: "play", key: "1" },
  { id: "docs", title: "Docs", icon: "book", key: "2" },
  { id: "samples", title: "Samples", icon: "code", key: "3" },
];

export const GROUPS = [
  { id: "nav", title: "Navigation", icon: "search", color: "#2563eb" },
  { id: "editor", title: "Editor", icon: "editor", color: "#059669" },
  { id: "window", title: "Window", icon: "window", color: "#7c3aed" },
  { id: "panels", title: "Panels", icon: "layers", color: "#ea580c" },
];

export const SHORTCUTS = [
  { id: "palette", group: "nav", keys: ["⌘", "K"], desc: "Search anything", icon: "search" },
  { id: "help", group: "nav", keys: ["⌘", "/"], desc: "Show all shortcuts", icon: "keyboard" },
  ...TABS.map((t) => ({ id: "tab-" + t.id, group: "nav", keys: ["⌘", t.key], desc: "Go to " + t.title, icon: t.icon })),
  { id: "run", group: "editor", keys: ["⌘", "Enter"], desc: "Run code", icon: "play" },
  { id: "format", group: "editor", keys: ["⇧", "⌘", "F"], desc: "Format with gofmt", icon: "format" },
  { id: "complete", group: "editor", keys: ["⌃", "Space"], desc: "Open auto-complete", icon: "code" },
  { id: "accept", group: "editor", keys: ["Tab"], desc: "Accept completion", icon: "code" },
  { id: "clear", group: "editor", keys: ["⌘", "L"], desc: "Clear output", icon: "trash" },
  { id: "copy", group: "editor", keys: ["⌘", "C"], desc: "Copy", icon: "copy" },
  { id: "cut", group: "editor", keys: ["⌘", "X"], desc: "Cut", icon: "copy" },
  { id: "paste", group: "editor", keys: ["⌘", "V"], desc: "Paste", icon: "copy" },
  { id: "fullscreen", group: "window", keys: ["⇧", "⌘", "Enter"], desc: "Toggle full screen", icon: "fullscreen" },
  { id: "zoomIn", group: "window", keys: ["⌘", "+"], desc: "Zoom in", icon: "zoomIn" },
  { id: "zoomOut", group: "window", keys: ["⌘", "-"], desc: "Zoom out", icon: "zoomOut" },
  { id: "zoomReset", group: "window", keys: ["⌘", "0"], desc: "Reset zoom", icon: "search" },
  { id: "screenshot", group: "window", keys: ["⌘", "P"], desc: "Capture screenshot", icon: "camera" },
  { id: "expand", group: "panels", keys: ["Click"], desc: "Expand or collapse a panel", icon: "expand" },
  { id: "collapse", group: "panels", keys: ["Esc"], desc: "Collapse expanded panel", icon: "collapse" },
  { id: "hide", group: "panels", keys: ["Click"], desc: "Hide panel to the side", icon: "hide" },
];

export function shortcutById(id) {
  return SHORTCUTS.find((s) => s.id === id);
}

export function keysLabel(s) {
  return s.keys.join(" ");
}

export function filterShortcuts(query) {
  const q = query.trim().toLowerCase();
  return GROUPS.map((g) => {
    const all = SHORTCUTS.filter((s) => s.group === g.id);
    if (q === "" || g.title.toLowerCase().includes(q)) return { group: g, items: all };
    return { group: g, items: all.filter((s) => s.desc.toLowerCase().includes(q) || keysLabel(s).toLowerCase().includes(q)) };
  }).filter((r) => r.items.length > 0);
}
