package streamstats

import (
	"fmt"
	"math"
	"sync"
)

type MomentStats struct {
	sync.RWMutex
	n  uint64
	m1 float64
	m2 float64
	m3 float64
	m4 float64
}

func NewMomentStats() MomentStats {
	return MomentStats{}
}

func (m *MomentStats) Push(x float64) {
	m.Lock()
	defer m.Unlock()
	m.n++
	fN := float64(m.n) // explicitly cast the number of observations to float64 for arithmetic operations
	delta := x - m.m1
	delta_n := delta / fN
	delta_n2 := delta_n * delta_n
	term1 := delta * delta_n * (fN - 1)
	m.m1 += delta_n
	m.m4 += term1*delta_n2*(fN*fN-3*m.n+3) + 6*delta_n2*m.m2 - 4*delta_n*m.m3
	m.m3 += term1*delta_n*(fN-2) - 3*delta_n*m.m2
	m.m2 += term1
}

func (m *MomentStats) N() uint64 {
	m.RLock()
	defer m.RUnlock()
	return m.n
}

func (m *MomentStats) Mean() float64 {
	m.RLock()
	defer m.RUnlock()
	return m.m1
}

func (m *MomentStats) Variance() float64 {
	m.RLock()
	defer m.RUnlock()
	if m.n > 1 {
		return m.m2 / (float64(m.n) - 1.0)
	} else {
		return 0.0
	}
}

func (m *MomentStats) StdDev() float64 {
	m.RLock()
	defer m.RUnlock()
	return math.Sqrt(m.Variance())
}

func (m *MomentStats) Skewness() float64 {
	m.RLock()
	defer m.RUnlock()
	if m.m2 > 0.0 {
		return math.Sqrt(float64(m.n)) * m.m3 / math.Pow(m.m2, 1.5)
	} else {
		return 0.0
	}
}

func (m *MomentStats) Kurtosis() float64 {
	m.RLock()
	defer m.RUnlock()
	if m.m2 > 0.0 {
		return float64(m.n)*m.m4/(m.m2*m.m2) - 3.0
	} else {
		return 0.0
	}
}

func (a *MomentStats) Combine(b *MomentStats) MomentStats {
	var combined MomentStats
	a.RLock()
	b.RLock()
	defer a.RUnlock()
	defer b.RUnlock()

	combined.n = a.n + b.n

	a_N := float64(a.n) // convert to floats for arithmetic operations
	b_N := float64(b.n)
	c_N := float64(combined.n)

	delta := b.m1 - a.m1
	delta2 := delta * delta
	delta3 := delta * delta2
	delta4 := delta2 * delta2

	combined.m1 = (a_N*a.m1 + b_N*b.m1) / c_N

	combined.m2 = a.m2 + b.m2 + delta2*a_N*b_N/c_N

	combined.m3 = a.m3 + b.m3 + delta3*a_N*b_N*(a_N-b_N)/(c_N*c_N)
	combined.m3 += 3.0 * delta * (a_N*b.m2 - b_N*a.m2) / c_N

	combined.m4 = a.m4 + b.m4 + delta4*a_N*b_N*(a_N*a_N-a_N*b_N+b_N*b_N)/(c_N*c_N*c_N)
	combined.m4 += 6.0*delta2*(a_N*a_N*b.m2+b_N*b_N*a.m2)/(c_N*c_N) + 4.0*delta*(a_N*b.m3-b_N*a.m3)/c_N

	return combined
}

func (m *MomentStats) String() string {
	m.RLock()
	defer m.RUnlock()
	return fmt.Sprintf("Mean: %f Variance: %f Skewness: %f Kurtosis: %f N: %d", m.Mean(), m.Variance(), m.Skewness(), m.Kurtosis(), m.N())
}
