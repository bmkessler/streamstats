package streamstats

type P2Quantile struct {
	P float64
	N [5]uint64
	Q [5]float64
}
