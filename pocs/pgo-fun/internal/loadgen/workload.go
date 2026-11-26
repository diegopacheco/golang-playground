package loadgen

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/diegopacheco/golang-playground/pocs/pgo-fun/internal/order"
)

type Request struct {
	Method string
	Path   string
	Body   []byte
}

type Workload struct {
	Seeds    []Request
	Requests []Request
}

var (
	countries = []string{"US", "BR", "DE", "PT", "JP"}
	coupons   = []string{"", "", "SAVE10", "BULK"}
	tags      = []string{"gift", "fragile", "promo", "eco", "express"}
)

func NewWorkload(seed uint64, orders, batches, batchSize int) (Workload, error) {
	rng := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	var w Workload
	for i := range orders {
		o := randomOrder(rng, fmt.Sprintf("o-%d", i))
		body, err := json.Marshal(o)
		if err != nil {
			return Workload{}, fmt.Errorf("marshal order: %w", err)
		}
		w.Seeds = append(w.Seeds, Request{Method: http.MethodPost, Path: "/orders", Body: body})
	}
	var summaries []Request
	for b := range batches {
		batch := make([]order.Order, batchSize)
		for i := range batch {
			batch[i] = randomOrder(rng, fmt.Sprintf("b-%d-%d", b, i))
		}
		body, err := json.Marshal(batch)
		if err != nil {
			return Workload{}, fmt.Errorf("marshal batch: %w", err)
		}
		summaries = append(summaries, Request{Method: http.MethodPost, Path: "/orders/summary", Body: body})
	}
	for i, s := range w.Seeds {
		w.Requests = append(w.Requests, s, Request{Method: http.MethodGet, Path: fmt.Sprintf("/orders/o-%d", i)})
		if len(summaries) > 0 {
			w.Requests = append(w.Requests, summaries[i%len(summaries)])
		}
	}
	return w, nil
}

func randomOrder(rng *rand.Rand, id string) order.Order {
	o := order.Order{
		ID:       id,
		Customer: fmt.Sprintf("customer-%03d", rng.IntN(200)),
		Country:  countries[rng.IntN(len(countries))],
		Coupon:   coupons[rng.IntN(len(coupons))],
	}
	n := 1 + rng.IntN(12)
	for j := range n {
		o.Items = append(o.Items, order.Item{
			SKU:       fmt.Sprintf("sku-%04d", rng.IntN(500)),
			Name:      fmt.Sprintf("product number %d with a descriptive name", j),
			Qty:       1 + rng.IntN(5),
			UnitCents: int64(100 + rng.IntN(20000)),
			Tags:      []string{tags[rng.IntN(len(tags))], tags[rng.IntN(len(tags))]},
		})
	}
	return o
}
