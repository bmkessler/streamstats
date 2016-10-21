package streamstats

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestGaussianMomentStats(t *testing.T) {
	rand.Seed(42) // for deterministic testing
	N := 100000
	// mean/stdev pairs for testing
	testCases := [][2]float64{
		[2]float64{0.0, 1.0},    // standard normal distribution
		[2]float64{25.0, 1.0},   // shifted mean
		[2]float64{0.0, 15.0},   // higher variance
		[2]float64{-35.0, 12.5}, // shifted mean and higher variance
	}
	for _, testCase := range testCases {
		mean := testCase[0]
		stdev := testCase[1]
		skew := 0.0
		kurt := 0.0
		eps := 3.0 * stdev / math.Sqrt(float64(N)) // expected error rate <0.3% in the mean
		m := NewMomentStats()
		for i := 0; i < N; i++ { // put in 10,000 random normal numbers
			m.Push(gaussianRandomVariable(mean, stdev))
		}
		if m.N() != uint64(N) {
			t.Errorf("Expected N: %v, got %v", N, m.N())
		}
		if math.Abs(m.Mean()-mean) > eps {
			t.Errorf("Expected Mean == %v, got %v", mean, m.Mean())
		}
		if math.Abs(m.Variance()-stdev*stdev) > stdev*eps {
			t.Errorf("Expected Variance == %v, got %v", stdev*stdev, m.Variance())
		}
		if math.Abs(m.StdDev()-stdev) > eps {
			t.Errorf("Expected StdDev == %v, got %v", stdev, m.StdDev())
		}
		if math.Abs(m.Skewness()-skew) > 1.5*eps {
			t.Errorf("Expected Skewness == %v, got %v", skew, m.Skewness())
		}
		if math.Abs(m.Kurtosis()-kurt) > 2*eps {
			t.Errorf("Expected Kurtosis == %v, got %v", kurt, m.Kurtosis())
		}
	}
}

func BenchmarkMomentStatsPush(b *testing.B) {
	m := NewMomentStats()
	for i := 0; i < b.N; i++ {
		m.Push(gaussianTestData[i%N])
	}
	result = m.Mean() // to avoid optimizing out the loop entirely
}

func BenchmarkMomentStatsPushReadContention(b *testing.B) {
	m := NewMomentStats()
	contentionInterval := time.Nanosecond * 1 // interval to contend
	go func() {
		ticker := time.NewTicker(contentionInterval)
		for _ = range ticker.C {
			result = m.Mean() // a contentious read
		}
	}()
	for i := 0; i < b.N; i++ {
		m.Push(gaussianTestData[i%N])
	}
	result = m.Mean() // to avoid optimizing out the loop entirely
}

func BenchmarkMomentStatsPushWriteContention(b *testing.B) {
	m := NewMomentStats()
	contentionInterval := time.Nanosecond * 1 // interval to contend
	go func() {
		ticker := time.NewTicker(contentionInterval)
		for t := range ticker.C {
			m.Push(gaussianTestData[t.Nanosecond()%N]) // a contentious write
		}
	}()
	for i := 0; i < b.N; i++ {
		m.Push(gaussianTestData[i%N])
	}
	result = m.Mean() // to avoid optimizing out the loop entirely
}