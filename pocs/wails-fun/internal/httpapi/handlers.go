package httpapi

import (
	"net/http"

	"github.com/diegopacheco/wails-fun/internal/employee"
	"github.com/diegopacheco/wails-fun/internal/invoice"
	"github.com/diegopacheco/wails-fun/internal/receipt"
	"github.com/diegopacheco/wails-fun/internal/sale"
)

type handlers struct {
	svc Services
	rs  responder
}

func (h handlers) listEmployees(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Employees.List(r.Context())
	h.rs.reply(w, r, http.StatusOK, list, err)
}

func (h handlers) createEmployee(w http.ResponseWriter, r *http.Request) {
	in, err := decode[employee.Input](r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	e, err := h.svc.Employees.Create(r.Context(), in)
	h.rs.reply(w, r, http.StatusCreated, e, err)
}

func (h handlers) getEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	e, err := h.svc.Employees.Get(r.Context(), id)
	h.rs.reply(w, r, http.StatusOK, e, err)
}

func (h handlers) updateEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	in, err := decode[employee.Input](r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	e, err := h.svc.Employees.Update(r.Context(), id, in)
	h.rs.reply(w, r, http.StatusOK, e, err)
}

func (h handlers) listSales(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Sales.List(r.Context())
	h.rs.reply(w, r, http.StatusOK, list, err)
}

func (h handlers) createSale(w http.ResponseWriter, r *http.Request) {
	in, err := decode[sale.Input](r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	s, err := h.svc.Sales.Create(r.Context(), in)
	h.rs.reply(w, r, http.StatusCreated, s, err)
}

func (h handlers) getSale(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	s, err := h.svc.Sales.Get(r.Context(), id)
	h.rs.reply(w, r, http.StatusOK, s, err)
}

func (h handlers) listReceipts(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Receipts.List(r.Context())
	h.rs.reply(w, r, http.StatusOK, list, err)
}

func (h handlers) issueReceipt(w http.ResponseWriter, r *http.Request) {
	in, err := decode[receipt.Input](r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	rc, err := h.svc.Receipts.Issue(r.Context(), in)
	h.rs.reply(w, r, http.StatusCreated, rc, err)
}

func (h handlers) listInvoices(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Invoices.List(r.Context())
	h.rs.reply(w, r, http.StatusOK, list, err)
}

func (h handlers) createInvoice(w http.ResponseWriter, r *http.Request) {
	in, err := decode[invoice.Input](r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	inv, err := h.svc.Invoices.Create(r.Context(), in)
	h.rs.reply(w, r, http.StatusCreated, inv, err)
}

func (h handlers) payInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	inv, err := h.svc.Invoices.Pay(r.Context(), id)
	h.rs.reply(w, r, http.StatusOK, inv, err)
}

func (h handlers) voidInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.rs.error(r.Context(), w, err)
		return
	}
	inv, err := h.svc.Invoices.Void(r.Context(), id)
	h.rs.reply(w, r, http.StatusOK, inv, err)
}
