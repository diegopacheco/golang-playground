async function request(path, options = {}) {
  const resp = await fetch(path, { headers: { "Content-Type": "application/json" }, ...options });
  const body = await resp.json();
  if (!resp.ok) throw new Error(body.error ?? "request failed");
  return body;
}

export const api = {
  run: (code) => request("/api/run", { method: "POST", body: JSON.stringify({ code }) }),
  format: (code) => request("/api/format", { method: "POST", body: JSON.stringify({ code }) }),
  complete: (prefix) => request("/api/complete?prefix=" + encodeURIComponent(prefix)).catch(() => []),
  docs: () => request("/api/docs"),
  doc: (id) => request("/api/docs/" + encodeURIComponent(id)),
  samples: () => request("/api/samples"),
};
