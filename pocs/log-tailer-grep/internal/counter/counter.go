package counter

type Counters struct {
	Lines   int64
	Matched int64
	Hits    []int64
}

func New(patterns int) *Counters {
	return &Counters{Hits: make([]int64, patterns)}
}

func (c *Counters) Record(hits []int) {
	c.Lines++
	if len(hits) > 0 {
		c.Matched++
	}
	for _, h := range hits {
		c.Hits[h]++
	}
}
