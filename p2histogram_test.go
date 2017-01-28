package streamstats

import "testing"

var initialP2H4 = P2Histogram{
	b: 4,
	n: []uint64{1, 2, 3, 4, 0},
	q: []float64{0, 0, 0, 0, 0},
}

var initialP2H10 = P2Histogram{
	b: 10,
	n: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 0},
	q: []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

func TestNewP2Histogram(t *testing.T) {
	// test 4 bins, same as median p=0.5 P2Histogram
	h := NewP2Histogram(4)
	if h.b != initialP2H4.b {
		t.Errorf("Expected b to be %v, got %v", initialP2H4.b, h.b)
	}
	if len(h.n) != h.b+1 {
		t.Errorf("Expected len(n)==%v, got len(n)=%v", h.b+1, len(h.n))
	}
	for j := 0; j < len(h.n); j++ {
		// check n
		if initialP2H4.n[j] != h.n[j] {
			t.Errorf("Expected n[%v]=%v, got n[%v]=%v", j, initialP2H4.n[j], j, h.n[j])
		}
		// check q
		if initialP2H4.q[j] != h.q[j] {
			t.Errorf("Expected q[%v]=%v, got q[%v]=%v", j, initialP2H4.q[j], j, h.q[j])
		}
	}

	// test 10 bins
	h = NewP2Histogram(10)
	if h.b != initialP2H10.b {
		t.Errorf("Expected b to be %v, got %v", initialP2H10.b, h.b)
	}
	if len(h.n) != h.b+1 {
		t.Errorf("Expected len(n)==%v, got len(n)=%v", h.b+1, len(h.n))
	}
	for j := 0; j < len(h.n); j++ {
		// check n
		if initialP2H10.n[j] != h.n[j] {
			t.Errorf("Expected n[%v]=%v, got n[%v]=%v", j, initialP2H10.n[j], j, h.n[j])
		}
		// check q
		if initialP2H10.q[j] != h.q[j] {
			t.Errorf("Expected q[%v]=%v, got q[%v]=%v", j, initialP2H10.q[j], j, h.q[j])
		}
	}

}

var histogramSmallNTestData = []float64{4, 6, 5, 7, 3, 1, 2}
var histogramSmallNExpectedData = []float64{1, 2, 3, 4, 5, 6, 7}

func TestP2HistorgramSmallN(t *testing.T) {
	h := NewP2Histogram(len(histogramSmallNTestData) - 1)
	for _, x := range histogramSmallNTestData {
		h.Push(x)
	}
	for i, x := range histogramSmallNExpectedData {
		if float64(h.n[i]) != x {
			t.Errorf("Expected n[%v]=%v, got n[%v]=%v", i, x, i, h.n[i])
		}
		if h.q[i] != x {
			t.Errorf("Expected q[%v]=%v, got q[%v]=%v", i, x, i, h.q[i])
		}
	}
}

/*

// dataPoints is the test data from Table 1 in the paper
var dataPoints = []float64{
	0.02,
	0.15,
	0.74,
	3.39,
	0.83,
	22.37,
	10.15,
	15.43,
	38.62,
	15.92,
	34.60,
	10.28,
	1.47,
	0.40,
	0.05,
	11.39,
	0.27,
	0.42,
	0.09,
	11.37,
}

type expectedP2 struct {
	n  [5]uint64
	np [5]float64
	q  [5]float64
}

var dataPointsExpected = []expectedP2{
	{n: [5]uint64{1, 2, 3, 4, 1}, np: [5]float64{1, 2, 3, 4, 5}, q: [5]float64{0.02, 0, 0, 0, 0}},
	{n: [5]uint64{1, 2, 3, 4, 2}, np: [5]float64{1, 2, 3, 4, 5}, q: [5]float64{0.02, 0.15, 0, 0, 0}},
	{n: [5]uint64{1, 2, 3, 4, 3}, np: [5]float64{1, 2, 3, 4, 5}, q: [5]float64{0.02, 0.15, 0.74, 0, 0}},
	{n: [5]uint64{1, 2, 3, 4, 4}, np: [5]float64{1, 2, 3, 4, 5}, q: [5]float64{0.02, 0.15, 0.74, 3.39, 0}},
	{n: [5]uint64{1, 2, 3, 4, 5}, np: [5]float64{1, 2, 3, 4, 5}, q: [5]float64{0.02, 0.15, 0.74, 0.83, 3.39}},
	{n: [5]uint64{1, 2, 3, 4, 6}, np: [5]float64{1, 2.25, 3.5, 4.75, 6}, q: [5]float64{0.02, 0.15, 0.74, 0.83, 22.37}},
	{n: [5]uint64{1, 2, 3, 5, 7}, np: [5]float64{1, 2.5, 4, 5.5, 7}, q: [5]float64{0.02, 0.15, 0.74, 4.465, 22.37}},
	{n: [5]uint64{1, 2, 4, 6, 8}, np: [5]float64{1, 2.75, 4.5, 6.25, 8}, q: [5]float64{0.02, 0.15, 2.18, 8.60, 22.37}},
	{n: [5]uint64{1, 3, 5, 7, 9}, np: [5]float64{1, 3, 5, 7, 9}, q: [5]float64{0.02, 0.87, 4.75, 15.52, 38.62}},
	{n: [5]uint64{1, 3, 5, 7, 10}, np: [5]float64{1, 3.25, 5.5, 7.75, 10}, q: [5]float64{0.02, 0.87, 4.75, 15.52, 38.62}},
	{n: [5]uint64{1, 3, 6, 8, 11}, np: [5]float64{1, 3.5, 6, 8.5, 11}, q: [5]float64{0.02, 0.87, 9.28, 21.58, 38.62}},
	{n: [5]uint64{1, 3, 6, 9, 12}, np: [5]float64{1, 3.75, 6.5, 9.25, 12}, q: [5]float64{0.02, 0.87, 9.28, 21.58, 38.62}},
	{n: [5]uint64{1, 4, 7, 10, 13}, np: [5]float64{1, 4, 7, 10, 13}, q: [5]float64{0.02, 2.14, 9.28, 21.58, 38.62}},
	{n: [5]uint64{1, 5, 8, 11, 14}, np: [5]float64{1, 4.25, 7.5, 10.75, 14}, q: [5]float64{0.02, 2.14, 9.28, 21.58, 38.62}},
	{n: [5]uint64{1, 5, 8, 12, 15}, np: [5]float64{1, 4.5, 8, 11.5, 15}, q: [5]float64{0.02, 0.74, 6.30, 21.58, 38.62}},
	{n: [5]uint64{1, 5, 8, 13, 16}, np: [5]float64{1, 4.75, 8.5, 12.25, 16}, q: [5]float64{0.02, 0.74, 6.30, 21.58, 38.62}},
	{n: [5]uint64{1, 5, 9, 13, 17}, np: [5]float64{1, 5, 9, 13, 17}, q: [5]float64{0.02, 0.59, 6.30, 17.22, 38.62}},
	{n: [5]uint64{1, 6, 10, 14, 18}, np: [5]float64{1, 5.25, 9.5, 13.75, 18}, q: [5]float64{0.02, 0.59, 6.30, 17.22, 38.62}},
	{n: [5]uint64{1, 6, 10, 15, 19}, np: [5]float64{1, 5.5, 10, 14.5, 19}, q: [5]float64{0.02, 0.50, 4.44, 17.22, 38.62}},
	{n: [5]uint64{1, 6, 10, 16, 20}, np: [5]float64{1, 5.75, 10.5, 15.25, 20}, q: [5]float64{0.02, 0.50, 4.44, 17.22, 38.62}},
}

func TestP2DataPoints(t *testing.T) {
	q := NewP2Quantile(0.5)
	for i, x := range dataPoints {
		q.Push(x)
		for j := 0; j < 5; j++ {
			// check n
			if dataPointsExpected[i].n[j] != q.n[j] {
				t.Errorf("Added data point[%v] %v, expected n[%v]=%v, got n[%v]=%v", i, x, j, dataPointsExpected[i].n[j], j, q.n[j])
			}
			// check np
			if dataPointsExpected[i].np[j] != q.np[j] {
				t.Errorf("Added data point[%v] %v, expected np[%v]=%v, got np[%v]=%v", i, x, j, dataPointsExpected[i].np[j], j, q.np[j])
			}
			// check q
			if math.Abs(dataPointsExpected[i].q[j]-q.q[j]) > 0.02 { // published table is only printed to 2 decimals and appears to use ceiling
				t.Errorf("Added data point[%v] %v, expected q[%v]=%v, got q[%v]=%v", i, x, j, dataPointsExpected[i].q[j], j, q.q[j])
			}
		}
	}
}

func TestP2GaussianDist(t *testing.T) {
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
		eps := 3.0 * stdev / math.Sqrt(float64(N)) // expected error rate <0.3%
		q := NewP2Quantile(0.5)                    // test the median
		for i := 0; i < N; i++ {                   // put in 10,000 random normal numbers
			q.Push(gaussianRandomVariable(mean, stdev))
		}
		z25 := 0.6745 // expected deviation at the 25%
		m := 4.0      // expect at least 4 std deviations for min and max
		p50 := mean
		p25 := mean + -1.0*z25*stdev
		p75 := mean + z25*stdev
		min := mean - m*stdev
		max := mean + m*stdev
		if q.N() != uint64(N) {
			t.Errorf("Expected N: %v, got %v", N, q.N())
		}
		if math.Abs(q.Quantile()-p50) > eps {
			t.Errorf("Expected Median == %v, got %v", p50, q.Quantile())
		}
		if q.Min() > min {
			t.Errorf("Expected Min < %v, got %v", min, q.Min())
		}
		if q.Max() < max {
			t.Errorf("Expected Max > %v, got %v", max, q.Max())
		}
		if math.Abs(q.UpperQuantile()-p75) > eps {
			t.Errorf("Expected UpperQuantile == %v, got %v", p75, q.UpperQuantile())
		}
		if math.Abs(q.LowerQuantile()-p25) > eps {
			t.Errorf("Expected LowerQuantile == %v, got %v", p25, q.LowerQuantile())
		}
	}
}

func TestP2ExponentialDist(t *testing.T) {
	rand.Seed(42) // for deterministic testing
	N := 100000
	eps := 0.03 // expect errors less than 3% for all quantiles
	lambdas := []float64{1.0, 2.0, 0.5}
	ps := []float64{0.10, 0.25, 0.50, 0.65, 0.95}
	for _, lambda := range lambdas {
		for _, p := range ps {
			q := NewP2Quantile(p)
			for i := 0; i < N; i++ {
				q.Push(exponentialRandomVariable(lambda))
			}
			if math.Abs((exponentialQuantile(p, lambda)-q.Quantile())/exponentialQuantile(p, lambda)) > eps {
				t.Errorf("Expected %v, got %v", exponentialQuantile(p, lambda), q.Quantile())
			}
		}
	}
}

func TestP2UniformDist(t *testing.T) {
	rand.Seed(42) // for deterministic testing
	N := 100000
	eps := 0.03 // expect errors less than 3% for all quantiles
	minMaxs := [][2]float64{
		[2]float64{0.0, 1.0},
		[2]float64{-10.0, 7.0},
		[2]float64{3.0, 4.0},
		[2]float64{210432.0, 921737123.0},
	}
	ps := []float64{0.10, 0.25, 0.50, 0.65, 0.95}
	for _, minMax := range minMaxs {
		min := minMax[0]
		max := minMax[1]
		for _, p := range ps {
			q := NewP2Quantile(p)
			for i := 0; i < N; i++ {
				q.Push(uniformRandomVariable(min, max))
			}
			if math.Abs((uniformQuantile(p, min, max)-q.Quantile())/uniformQuantile(p, min, max)) > eps {
				t.Errorf("P: %v, min: %v, max: %v, Expected %v, got %v", p, min, max, uniformQuantile(p, min, max), q.Quantile())
			}
		}
	}
}

*/

func BenchmarkP2Histogram8Push(b *testing.B) {
	q := NewP2Histogram(8)
	for i := 0; i < b.N; i++ {
		q.Push(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram16Push(b *testing.B) {
	q := NewP2Histogram(16)
	for i := 0; i < b.N; i++ {
		q.Push(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram32Push(b *testing.B) {
	q := NewP2Histogram(32)
	for i := 0; i < b.N; i++ {
		q.Push(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram64Push(b *testing.B) {
	q := NewP2Histogram(64)
	for i := 0; i < b.N; i++ {
		q.Push(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram128Push(b *testing.B) {
	q := NewP2Histogram(128)
	for i := 0; i < b.N; i++ {
		q.Push(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}
