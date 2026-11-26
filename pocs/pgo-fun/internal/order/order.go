package order

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrInvalid       = errors.New("invalid order")
	ErrUnknownCoupon = errors.New("unknown coupon")
)

type Item struct {
	SKU       string   `json:"sku"`
	Name      string   `json:"name"`
	Qty       int      `json:"qty"`
	UnitCents int64    `json:"unit_cents"`
	Tags      []string `json:"tags,omitempty"`
}

type Order struct {
	ID       string `json:"id"`
	Customer string `json:"customer"`
	Country  string `json:"country"`
	Coupon   string `json:"coupon,omitempty"`
	Items    []Item `json:"items"`
}

type Line struct {
	SKU        string `json:"sku"`
	Qty        int    `json:"qty"`
	TotalCents int64  `json:"total_cents"`
}

type Priced struct {
	ID            string `json:"id"`
	Customer      string `json:"customer"`
	Country       string `json:"country"`
	Lines         []Line `json:"lines"`
	ItemCount     int    `json:"item_count"`
	SubtotalCents int64  `json:"subtotal_cents"`
	DiscountCents int64  `json:"discount_cents"`
	TaxCents      int64  `json:"tax_cents"`
	TotalCents    int64  `json:"total_cents"`
}

type CustomerSummary struct {
	Customer   string `json:"customer"`
	Orders     int    `json:"orders"`
	TotalCents int64  `json:"total_cents"`
	TopSKU     string `json:"top_sku"`
}

var taxBasisPoints = map[string]int64{"US": 700, "BR": 1700, "DE": 1900}

const defaultTaxBasisPoints = 1000

func Price(o Order) (Priced, error) {
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.Customer) == "" || len(o.Items) == 0 {
		return Priced{}, fmt.Errorf("%w: id, customer and items are required", ErrInvalid)
	}
	p := Priced{ID: o.ID, Customer: o.Customer, Country: strings.ToUpper(o.Country), Lines: make([]Line, 0, len(o.Items))}
	for _, it := range o.Items {
		if it.SKU == "" || it.Qty <= 0 || it.UnitCents < 0 {
			return Priced{}, fmt.Errorf("%w: item %q needs sku, qty > 0 and unit_cents >= 0", ErrInvalid, it.SKU)
		}
		total := int64(it.Qty) * it.UnitCents
		p.Lines = append(p.Lines, Line{SKU: it.SKU, Qty: it.Qty, TotalCents: total})
		p.ItemCount += it.Qty
		p.SubtotalCents += total
	}
	pct, err := discountPercent(o.Coupon, p.ItemCount)
	if err != nil {
		return Priced{}, err
	}
	p.DiscountCents = p.SubtotalCents * pct / 100
	bp, ok := taxBasisPoints[p.Country]
	if !ok {
		bp = defaultTaxBasisPoints
	}
	taxable := p.SubtotalCents - p.DiscountCents
	p.TaxCents = (taxable*bp + 5000) / 10000
	p.TotalCents = taxable + p.TaxCents
	return p, nil
}

func discountPercent(coupon string, items int) (int64, error) {
	switch coupon {
	case "":
		return 0, nil
	case "SAVE10":
		return 10, nil
	case "BULK":
		if items >= 10 {
			return 5, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%w: %q", ErrUnknownCoupon, coupon)
	}
}

func Summarize(orders []Order) ([]CustomerSummary, error) {
	type acc struct {
		summary CustomerSummary
		qty     map[string]int
	}
	byCustomer := make(map[string]*acc)
	for _, o := range orders {
		p, err := Price(o)
		if err != nil {
			return nil, fmt.Errorf("order %q: %w", o.ID, err)
		}
		a, ok := byCustomer[p.Customer]
		if !ok {
			a = &acc{summary: CustomerSummary{Customer: p.Customer}, qty: make(map[string]int)}
			byCustomer[p.Customer] = a
		}
		a.summary.Orders++
		a.summary.TotalCents += p.TotalCents
		for _, l := range p.Lines {
			a.qty[l.SKU] += l.Qty
		}
	}
	out := make([]CustomerSummary, 0, len(byCustomer))
	for _, a := range byCustomer {
		a.summary.TopSKU = topSKU(a.qty)
		out = append(out, a.summary)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalCents != out[j].TotalCents {
			return out[i].TotalCents > out[j].TotalCents
		}
		return out[i].Customer < out[j].Customer
	})
	return out, nil
}

func topSKU(qty map[string]int) string {
	best, bestQty := "", -1
	for sku, q := range qty {
		if q > bestQty || (q == bestQty && sku < best) {
			best, bestQty = sku, q
		}
	}
	return best
}
