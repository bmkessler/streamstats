package streamstats

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestGaussianMomentStats(t *testing.T) {
	m := NewMomentStats()
	m.Add(1.0)
	if m.Variance() != 0.0 {
		t.Errorf("Expected zero Variance with only one point added got %f", m.Variance())
	}
	if m.Skewness() != 0.0 {
		t.Errorf("Expected zero Skewness with only one point added got %f", m.Skewness())
	}
	if m.Kurtosis() != 0.0 {
		t.Errorf("Expected zero Kurtosis with only one point added got %f", m.Kurtosis())
	}

	rand.Seed(42) // for deterministic testing
	N := 100000
	// mean/stdev pairs for testing
	testCases := [][2]float64{
		{0.0, 1.0},    // standard normal distribution
		{25.0, 1.0},   // shifted mean
		{0.0, 15.0},   // higher variance
		{-35.0, 12.5}, // shifted mean and higher variance
	}
	for _, testCase := range testCases {
		mean := testCase[0]
		stdev := testCase[1]
		skew := 0.0
		kurt := 0.0
		eps := 3.0 * stdev / math.Sqrt(float64(N)) // expected error rate <0.3% in the mean
		m = NewMomentStats()
		for i := 0; i < N; i++ { // put in 10,000 random normal numbers
			m.Add(gaussianRandomVariable(mean, stdev))
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
		expectedString := fmt.Sprintf("Mean: %0.3f Variance: %0.3f Skewness: %0.3f Kurtosis: %0.3f N: %d", m.Mean(), m.Variance(), m.Skewness(), m.Kurtosis(), m.N())
		if m.String() != expectedString {
			t.Errorf("Expected %s got %s", expectedString, m)
		}
	}
	// combine two measurements
	N = 1000
	meanA := 1.5
	meanB := -0.5
	meanC := (meanA + meanB) / 2.0
	stdevA := 2.0
	stdevB := 3.0
	stdevC := math.Sqrt(stdevA*stdevA + stdevB*stdevB)
	mA := NewMomentStats()
	mB := NewMomentStats()
	mTotal := NewMomentStats()
	for i := 0; i < N; i++ { // put in N random normal numbers
		x := meanA + stdevA*gaussianTestData[i]
		mA.Add(x)
		mTotal.Add(x)
	}
	for i := N; i < 2*N; i++ { // put in N random normal numbers
		x := meanB + stdevB*gaussianTestData[i]
		mB.Add(x)
		mTotal.Add(x)
	}
	mC := mA.Combine(mB)
	eps := 3.0 * stdevC / math.Sqrt(float64(N)) // expected error rate <0.3% in the mean
	if math.Abs(mC.Mean()-meanC) > eps {
		t.Errorf("Expected Combined Mean == %v, got %v", meanC, mC.Mean())
	}
	if math.Abs(mC.StdDev()-mTotal.StdDev()) > eps {
		t.Errorf("Expected Combined StdDev == %v, got %v", mTotal.StdDev(), mC.StdDev())
	}
}

func BenchmarkMomentStatsAdd(b *testing.B) {
	m := NewMomentStats()
	for i := 0; i < b.N; i++ {
		m.Add(gaussianTestData[i&mask])
	}
	result = m.Mean() // to avoid optimizing out the loop entirely
}
