package httpapi

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/diegopacheco/wails-fun/internal/employee"
	"github.com/diegopacheco/wails-fun/internal/invoice"
	"github.com/diegopacheco/wails-fun/internal/logx"
	"github.com/diegopacheco/wails-fun/internal/receipt"
	"github.com/diegopacheco/wails-fun/internal/sale"
)

type employees interface {
	Create(ctx context.Context, in employee.Input) (employee.Employee, error)
	Update(ctx context.Context, id int64, in employee.Input) (employee.Employee, error)
	Get(ctx context.Context, id int64) (employee.Employee, error)
	List(ctx context.Context) ([]employee.Employee, error)
}

type sales interface {
	Create(ctx context.Context, in sale.Input) (sale.Sale, error)
	Get(ctx context.Context, id int64) (sale.Sale, error)
	List(ctx context.Context) ([]sale.Sale, error)
}

type receipts interface {
	Issue(ctx context.Context, in receipt.Input) (receipt.Receipt, error)
	List(ctx context.Context) ([]receipt.Receipt, error)
}

type invoices interface {
	Create(ctx context.Context, in invoice.Input) (invoice.Invoice, error)
	Pay(ctx context.Context, id int64) (invoice.Invoice, error)
	Void(ctx context.Context, id int64) (invoice.Invoice, error)
	List(ctx context.Context) ([]invoice.Invoice, error)
}

type Services struct {
	Employees employees
	Sales     sales
	Receipts  receipts
	Invoices  invoices
}

func NewRouter(logger *slog.Logger, svc Services, assets fs.FS) http.Handler {
	h := handlers{svc: svc, rs: responder{logger: logger}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/employees", h.listEmployees)
	mux.HandleFunc("POST /api/employees", h.createEmployee)
	mux.HandleFunc("GET /api/employees/{id}", h.getEmployee)
	mux.HandleFunc("PUT /api/employees/{id}", h.updateEmployee)
	mux.HandleFunc("GET /api/sales", h.listSales)
	mux.HandleFunc("POST /api/sales", h.createSale)
	mux.HandleFunc("GET /api/sales/{id}", h.getSale)
	mux.HandleFunc("GET /api/receipts", h.listReceipts)
	mux.HandleFunc("POST /api/receipts", h.issueReceipt)
	mux.HandleFunc("GET /api/invoices", h.listInvoices)
	mux.HandleFunc("POST /api/invoices", h.createInvoice)
	mux.HandleFunc("POST /api/invoices/{id}/pay", h.payInvoice)
	mux.HandleFunc("POST /api/invoices/{id}/void", h.voidInvoice)
	mux.Handle("GET /", http.FileServerFS(assets))
	return logx.Middleware(logger, mux)
}
