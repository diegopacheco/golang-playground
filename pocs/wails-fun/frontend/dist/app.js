const state = { employees: [], sales: [], receipts: [], invoices: [] };

const $ = (sel) => document.querySelector(sel);

const money = (cents) =>
  (cents / 100).toLocaleString("en-US", { style: "currency", currency: "USD" });

const day = (iso) => new Date(iso).toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric" });

const esc = (s) => String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);

async function api(path, options = {}) {
  const res = await fetch(`/api${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

function toast(message, isError = false) {
  const el = $("#toast");
  el.textContent = message;
  el.className = isError ? "toast error" : "toast";
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => el.classList.add("hidden"), 2600);
}

async function attempt(action, success) {
  try {
    await action();
    if (success) toast(success);
    await refresh();
  } catch (err) {
    toast(err.message, true);
  }
}

function rows(body, items, render, cols) {
  $(body).innerHTML = items.length
    ? items.map(render).join("")
    : `<tr><td class="empty" colspan="${cols}">Nothing here yet</td></tr>`;
}

function renderStats() {
  const revenue = state.sales.reduce((sum, s) => sum + s.totalCents, 0);
  const open = state.invoices.filter((i) => i.status === "OPEN").reduce((sum, i) => sum + i.amountCents, 0);
  $("#stats").innerHTML = [
    ["Revenue", money(revenue)],
    ["Open invoices", money(open)],
    ["Employees", state.employees.length],
  ].map(([k, v]) => `<div class="stat"><span>${k}</span><strong>${v}</strong></div>`).join("");
}

function renderEmployees() {
  rows("#employees-body", state.employees, (e) => `
    <tr>
      <td>${e.id}</td><td>${esc(e.name)}</td><td>${esc(e.email)}</td><td>${esc(e.role)}</td><td>${day(e.createdAt)}</td>
      <td class="btns"><button class="small ghost" data-edit="${e.id}">Edit</button></td>
    </tr>`, 6);
  const select = $("#sale-form [name=employeeId]");
  const current = select.value;
  select.innerHTML = state.employees.map((e) => `<option value="${e.id}">${esc(e.name)}</option>`).join("");
  if (current) select.value = current;
}

function renderSales() {
  const hasReceipt = new Set(state.receipts.map((r) => r.saleId));
  const hasInvoice = new Set(state.invoices.map((i) => i.saleId));
  rows("#sales-body", state.sales, (s) => `
    <tr>
      <td>${s.id}</td><td>${day(s.createdAt)}</td><td>${esc(s.customer)}</td><td>${esc(s.employeeName)}</td>
      <td class="tag">${s.items.map((i) => `${i.quantity}x ${esc(i.description)}`).join(", ")}</td>
      <td class="num">${money(s.totalCents)}</td>
      <td class="btns">
        <button class="small ghost" data-receipt="${s.id}" ${hasReceipt.has(s.id) ? "disabled" : ""}>Receipt</button>
        <button class="small ghost" data-invoice="${s.id}" ${hasInvoice.has(s.id) ? "disabled" : ""}>Invoice</button>
      </td>
    </tr>`, 7);
}

function renderReceipts() {
  rows("#receipts-body", state.receipts, (r) => `
    <tr>
      <td><strong>${r.number}</strong></td><td>#${r.saleId}</td><td>${esc(r.customer)}</td>
      <td>${r.method}</td><td>${day(r.issuedAt)}</td><td class="num">${money(r.amountCents)}</td>
    </tr>`, 6);
}

function renderInvoices() {
  rows("#invoices-body", state.invoices, (i) => `
    <tr>
      <td><strong>${i.number}</strong></td><td>#${i.saleId}</td><td>${esc(i.customer)}</td>
      <td>${day(i.issuedAt)}</td><td>${day(i.dueAt)}</td>
      <td><span class="badge ${i.status}">${i.status}</span> ${i.overdue ? '<span class="badge OVERDUE">OVERDUE</span>' : ""}</td>
      <td class="num">${money(i.amountCents)}</td>
      <td class="btns">${i.status === "OPEN"
        ? `<button class="small" data-pay="${i.id}">Pay</button><button class="small danger" data-void="${i.id}">Void</button>`
        : ""}</td>
    </tr>`, 8);
}

async function refresh() {
  const [employees, sales, receipts, invoices] = await Promise.all([
    api("/employees"), api("/sales"), api("/receipts"), api("/invoices"),
  ]);
  Object.assign(state, { employees, sales, receipts, invoices });
  renderEmployees();
  renderSales();
  renderReceipts();
  renderInvoices();
  renderStats();
}

function addItemRow() {
  const row = document.createElement("div");
  row.className = "item-row";
  row.innerHTML = `
    <label>Description<input name="description" required></label>
    <label>Qty<input name="quantity" type="number" min="1" value="1" required></label>
    <label>Unit price ($)<input name="price" type="number" min="0" step="0.01" value="0" required></label>
    <div class="sub">$0.00</div>
    <button type="button" class="ghost" title="Remove">x</button>`;
  row.querySelector("button").onclick = () => {
    if ($("#items").children.length > 1) row.remove();
    updateTotal();
  };
  $("#items").appendChild(row);
}

function readItems() {
  return [...$("#items").children].map((row) => ({
    description: row.querySelector("[name=description]").value,
    quantity: parseInt(row.querySelector("[name=quantity]").value, 10) || 0,
    unitPriceCents: Math.round((parseFloat(row.querySelector("[name=price]").value) || 0) * 100),
  }));
}

function updateTotal() {
  const items = readItems();
  [...$("#items").children].forEach((row, i) => {
    row.querySelector(".sub").textContent = money(items[i].quantity * items[i].unitPriceCents);
  });
  $("#sale-total").textContent = money(items.reduce((sum, i) => sum + i.quantity * i.unitPriceCents, 0));
}

function resetEmployeeForm() {
  const form = $("#employee-form");
  form.reset();
  form.id.value = "";
  $("#employee-submit").textContent = "Add employee";
  $("#employee-cancel").classList.add("hidden");
}

function openDocDialog(kind, saleId) {
  const form = $("#doc-form");
  form.kind.value = kind;
  form.saleId.value = saleId;
  $("#doc-title").textContent = `${kind === "receipt" ? "Receipt" : "Invoice"} for sale #${saleId}`;
  $("#method-field").classList.toggle("hidden", kind !== "receipt");
  $("#due-field").classList.toggle("hidden", kind !== "invoice");
  $("#doc-dialog").showModal();
}

function bind() {
  document.querySelectorAll(".tab").forEach((tab) => {
    tab.onclick = () => {
      document.querySelectorAll(".tab").forEach((t) => t.classList.toggle("active", t === tab));
      document.querySelectorAll(".panel").forEach((p) => p.classList.toggle("hidden", p.id !== tab.dataset.tab));
    };
  });

  $("#add-item").onclick = () => { addItemRow(); updateTotal(); };
  $("#items").oninput = updateTotal;

  $("#sale-form").onsubmit = (ev) => {
    ev.preventDefault();
    const form = ev.target;
    attempt(async () => {
      await api("/sales", { method: "POST", body: {
        employeeId: parseInt(form.employeeId.value, 10) || 0,
        customer: form.customer.value,
        items: readItems(),
      } });
      form.customer.value = "";
      $("#items").innerHTML = "";
      addItemRow();
      updateTotal();
    }, "Sale saved");
  };

  $("#employee-form").onsubmit = (ev) => {
    ev.preventDefault();
    const form = ev.target;
    const body = { name: form.name.value, email: form.email.value, role: form.role.value };
    const id = form.id.value;
    attempt(async () => {
      await api(id ? `/employees/${id}` : "/employees", { method: id ? "PUT" : "POST", body });
      resetEmployeeForm();
    }, id ? "Employee updated" : "Employee added");
  };
  $("#employee-cancel").onclick = resetEmployeeForm;

  $("#employees-body").onclick = (ev) => {
    const id = ev.target.dataset.edit;
    if (!id) return;
    const e = state.employees.find((x) => String(x.id) === id);
    const form = $("#employee-form");
    form.id.value = e.id;
    form.name.value = e.name;
    form.email.value = e.email;
    form.role.value = e.role;
    $("#employee-submit").textContent = "Save changes";
    $("#employee-cancel").classList.remove("hidden");
  };

  $("#sales-body").onclick = (ev) => {
    const { receipt, invoice } = ev.target.dataset;
    if (receipt) openDocDialog("receipt", receipt);
    if (invoice) openDocDialog("invoice", invoice);
  };

  $("#doc-cancel").onclick = () => $("#doc-dialog").close();
  $("#doc-form").onsubmit = (ev) => {
    ev.preventDefault();
    const form = ev.target;
    const saleId = parseInt(form.saleId.value, 10);
    const isReceipt = form.kind.value === "receipt";
    $("#doc-dialog").close();
    attempt(() => isReceipt
      ? api("/receipts", { method: "POST", body: { saleId, method: form.method.value } })
      : api("/invoices", { method: "POST", body: { saleId, dueDays: parseInt(form.dueDays.value, 10) || 0 } }),
    isReceipt ? "Receipt issued" : "Invoice issued");
  };

  $("#invoices-body").onclick = (ev) => {
    const { pay, void: voidId } = ev.target.dataset;
    if (pay) attempt(() => api(`/invoices/${pay}/pay`, { method: "POST" }), "Invoice paid");
    if (voidId) attempt(() => api(`/invoices/${voidId}/void`, { method: "POST" }), "Invoice voided");
  };
}

bind();
addItemRow();
attempt(() => Promise.resolve());
