package streamstats

import (
	"fmt"
	"math"
)

// MomentStats is a datastructure for computing the first four moments of a stream
type MomentStats struct {
	n  uint64
	m1 float64
	m2 float64
	m3 float64
	m4 float64
}

// NewMomentStats returns an empty MomentStats structure with no values
func NewMomentStats() MomentStats {
	return MomentStats{}
}

// Push updates the moment stats
func (m *MomentStats) Push(x float64) {
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

// N returns the observations stored so far
func (m *MomentStats) N() uint64 {
	return m.n
}

// Mean returns the mean of the observations seen so far
func (m *MomentStats) Mean() float64 {
	return m.m1
}

// Variance returns the variance of the observations seen so far
func (m *MomentStats) Variance() float64 {
	if m.n < 2 {
		return 0.0
	}
	return m.m2 / (float64(m.n) - 1.0)
}

// StdDev returns the standard deviation of the samples seen so far
func (m *MomentStats) StdDev() float64 {
	return math.Sqrt(m.Variance())
}

// Skewness returns the skewness of the samples seen so far
func (m *MomentStats) Skewness() float64 {
	if m.m2 <= 0.0 {
		return 0.0
	}
	return math.Sqrt(float64(m.n)) * m.m3 / math.Pow(m.m2, 1.5)
}

// Kurtosis returns the excess kurtosis of the samples seen so far
func (m *MomentStats) Kurtosis() float64 {
	if m.m2 <= 0.0 {
		return 0.0
	}
	return float64(m.n)*m.m4/(m.m2*m.m2) - 3.0
}

// Combine combines the stats from two MomentStats structures
func (m *MomentStats) Combine(b *MomentStats) MomentStats {
	var combined MomentStats

	combined.n = m.n + b.n

	mN := float64(m.n) // convert to floats for arithmetic operations
	bN := float64(b.n)
	cN := float64(combined.n)

	delta := b.m1 - m.m1
	delta2 := delta * delta
	delta3 := delta * delta2
	delta4 := delta2 * delta2

	combined.m1 = (mN*m.m1 + bN*b.m1) / cN

	combined.m2 = m.m2 + b.m2 + delta2*mN*bN/cN

	combined.m3 = m.m3 + b.m3 + delta3*mN*bN*(mN-bN)/(cN*cN)
	combined.m3 += 3.0 * delta * (mN*b.m2 - bN*m.m2) / cN

	combined.m4 = m.m4 + b.m4 + delta4*mN*bN*(mN*mN-mN*bN+bN*bN)/(cN*cN*cN)
	combined.m4 += 6.0*delta2*(mN*mN*b.m2+bN*bN*m.m2)/(cN*cN) + 4.0*delta*(mN*b.m3-bN*m.m3)/cN

	return combined
}

// String returns the standard string representation of the samples seen so far
func (m *MomentStats) String() string {
	return fmt.Sprintf("Mean: %f Variance: %f Skewness: %f Kurtosis: %f N: %d", m.Mean(), m.Variance(), m.Skewness(), m.Kurtosis(), m.N())
}
