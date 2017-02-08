package streamstats

import (
	"math"
	"math/rand"
	"testing"
)

func TestCovarStats(t *testing.T) {
	rand.Seed(42) // for deterministic testing
	N := 10000

	cv := NewCovarStats()
	xS := NewMomentStats()
	yS := NewMomentStats()
	x0 := 1.5
	xNoise := 1.0
	slope := 2.5
	intercept := -0.5
	yNoise := 0.25

	for i := 0; i < N; i++ {
		x := gaussianRandomVariable(x0, xNoise)
		y := slope*x + intercept + gaussianRandomVariable(0.0, yNoise)
		cv.Add(x, y)
		xS.Add(x)
		yS.Add(y)
	}
	if cv.N() != uint64(N) {
		t.Errorf("Expected N %d got %d", N, cv.N())
	}
	acceptableError := 0.01
	if math.Abs(cv.Slope()-slope) > acceptableError {
		t.Errorf("Expected Slope %f got %f", slope, cv.Slope())
	}
	if math.Abs(cv.Intercept()-intercept) > acceptableError {
		t.Errorf("Expected Intercept %f got %f", intercept, cv.Intercept())
	}
	if 1.0-cv.Correlation() > acceptableError {
		t.Errorf("Expected Correlation %f got %f", 1.0, cv.Correlation())
	}
	if cv.XMean() != xS.Mean() {
		t.Errorf("Expected XMean %f got %f", xS.Mean(), cv.XMean())
	}
	if cv.XVariance() != xS.Variance() {
		t.Errorf("Expected XVariance %f got %f", xS.Variance(), cv.XVariance())
	}
	if cv.XStdDev() != xS.StdDev() {
		t.Errorf("Expected XStdDev %f got %f", xS.StdDev(), cv.XStdDev())
	}
	if cv.XSkewness() != xS.Skewness() {
		t.Errorf("Expected XSkewness %f got %f", xS.Skewness(), cv.XSkewness())
	}
	if cv.XKurtosis() != xS.Kurtosis() {
		t.Errorf("Expected XKurtosis %f got %f", xS.Kurtosis(), cv.XKurtosis())
	}

	if cv.YMean() != yS.Mean() {
		t.Errorf("Expected YMean %f got %f", yS.Mean(), cv.YMean())
	}
	if cv.YVariance() != yS.Variance() {
		t.Errorf("Expected YVariance %f got %f", yS.Variance(), cv.YVariance())
	}
	if cv.YStdDev() != yS.StdDev() {
		t.Errorf("Expected YStdDev %f got %f", yS.StdDev(), cv.YStdDev())
	}
	if cv.YSkewness() != yS.Skewness() {
		t.Errorf("Expected YSkewness %f got %f", yS.Skewness(), cv.YSkewness())
	}
	if cv.YKurtosis() != yS.Kurtosis() {
		t.Errorf("Expected YKurtosis %f got %f", yS.Kurtosis(), cv.YKurtosis())
	}
}
