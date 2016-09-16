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
	deltaN := delta / fN
	deltaN2 := deltaN * deltaN
	term1 := delta * deltaN * (fN - 1)
	m.m1 += deltaN
	m.m4 += term1*deltaN2*(fN*fN-3*fN+3) + 6*deltaN2*m.m2 - 4*deltaN*m.m3
	m.m3 += term1*deltaN*(fN-2) - 3*deltaN*m.m2
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

	aN := float64(a.n) // convert to floats for arithmetic operations
	bN := float64(b.n)
	cN := float64(combined.n)

	delta := b.m1 - a.m1
	delta2 := delta * delta
	delta3 := delta * delta2
	delta4 := delta2 * delta2

	combined.m1 = (aN*a.m1 + bN*b.m1) / cN

	combined.m2 = a.m2 + b.m2 + delta2*aN*bN/cN

	combined.m3 = a.m3 + b.m3 + delta3*aN*bN*(aN-bN)/(cN*cN)
	combined.m3 += 3.0 * delta * (aN*b.m2 - bN*a.m2) / cN

	combined.m4 = a.m4 + b.m4 + delta4*aN*bN*(aN*aN-aN*bN+bN*bN)/(cN*cN*cN)
	combined.m4 += 6.0*delta2*(aN*aN*b.m2+bN*bN*a.m2)/(cN*cN) + 4.0*delta*(aN*b.m3-bN*a.m3)/cN

	return combined
}

func (m *MomentStats) String() string {
	m.RLock()
	defer m.RUnlock()
	return fmt.Sprintf("Mean: %f Variance: %f Skewness: %f Kurtosis: %f N: %d", m.Mean(), m.Variance(), m.Skewness(), m.Kurtosis(), m.N())
}
