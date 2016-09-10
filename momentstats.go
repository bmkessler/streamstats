package streamstats

import (
	"math"
	"sync"
)

type OrderStats struct {
	sync.RWMutex
	N  uint64
	M1 float64
	M2 float64
	M3 float64
	M4 float64
}

func (o *OrderStats) Push(x float64) {
	o.Lock()
	defer o.Unlock()
	o.N++
	fN := float64(o.N) // explicitly cast the number of observations to float64 for arithmetic operations
	delta := x - o.M1
	delta_n := delta / fN
	delta_n2 := delta_n * delta_n
	term1 := delta * delta_n * (fN - 1)
	o.M1 += delta_n
	o.M4 += term1*delta_n2*(fN*fN-3*o.N+3) + 6*delta_n2*o.M2 - 4*delta_n*o.M3
	o.M3 += term1*delta_n*(fN-2) - 3*delta_n*o.M2
	o.M2 += term1
}

func (o *OrderStats) N() uint64 {
	o.RLock()
	defer o.RUnlock()
	return o.N
}

func (o *OrderStats) Mean() float64 {
	o.RLock()
	defer o.RUnlock()
	return o.M1
}

func (o *OrderStats) Variance() float64 {
	o.RLock()
	defer o.RUnlock()
	return o.M2 / (float64(o.N) - 1.0)
}

func (o *OrderStats) StdDev() float64 {
	o.RLock()
	defer o.RUnlock()
	return math.Sqrt(o.Variance())
}

func (o *OrderStats) Skewness() float64 {
	o.RLock()
	defer o.RUnlock()
	return math.Sqrt(float64(o.N)) * o.M3 / math.Pow(o.M2, 1.5)
}

func (o *OrderStats) Kurtosis() float64 {
	o.RLock()
	defer o.RUnlock()
	return float64(o.N)*o.M4/(o.M2*o.M2) - 3.0
}

func (a *OrderStats) Combine(b *OrderStats) OrderStats {
	var combined OrderStats
	a.RLock()
	b.RLock()
	defer a.RUnlock()
	defer b.RUnlock()

	combined.N = a.N + b.N

	a_N := float64(a.N) // convert to floats for arithmetic operations
	b_N := float64(b.N)
	c_N := float64(combined.N)

	delta := b.M1 - a.M1
	delta2 := delta * delta
	delta3 := delta * delta2
	delta4 := delta2 * delta2

	combined.M1 = (a_N*a.M1 + b_N*b.M1) / c_N

	combined.M2 = a.M2 + b.M2 + delta2*a_N*b_N/c_N

	combined.M3 = a.M3 + b.M3 + delta3*a_N*b_N*(a_N-b_N)/(c_N*c_N)
	combined.M3 += 3.0 * delta * (a_N*b.M2 - b_N*a.M2) / c_N

	combined.M4 = a.M4 + b.M4 + delta4*a_N*b_N*(a_N*a_N-a_N*b_N+b_N*b_N)/(c_N*c_N*c_N)
	combined.M4 += 6.0*delta2*(a_N*a_N*b.M2+b_N*b_N*a.M2)/(c_N*c_N) + 4.0*delta*(a_N*b.M3-b_N*a.M3)/c_N

	return combined
}


