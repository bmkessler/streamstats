package streamstats

import (
	"math"
	"math/rand"
	"testing"
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
		eps := 3.0 * stdev / math.Sqrt(float64(N)) // expected error rate <0.3%
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
		if math.Abs(m.Variance()-stdev*stdev) > eps {
			t.Errorf("Expected Variance == %v, got %v", stdev*stdev, m.Variance())
		}
		if math.Abs(m.StdDev()-stdev) > eps {
			t.Errorf("Expected StdDev == %v, got %v", stdev, m.StdDev())
		}
		if math.Abs(m.Skewness()-skew) > eps {
			t.Errorf("Expected Skewness == %v, got %v", skew, m.Skewness())
		}
		if math.Abs(m.Kurtosis()-kurt) > eps {
			t.Errorf("Expected Kurtosis == %v, got %v", kurt, m.Kurtosis())
		}
	}
}
