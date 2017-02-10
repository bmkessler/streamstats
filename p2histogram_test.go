package streamstats

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

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
	if uint64(len(h.n)) != h.b+1 {
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
	if uint64(len(h.n)) != h.b+1 {
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
	h := NewP2Histogram(uint64(len(histogramSmallNTestData) - 1))
	for _, x := range histogramSmallNTestData {
		h.Add(x)
	}
	for i, x := range histogramSmallNExpectedData {
		if float64(h.n[i]) != x {
			t.Errorf("Expected n[%v]=%v, got n[%v]=%v", i, x, i, h.n[i])
		}
		if h.q[i] != x {
			t.Errorf("Expected q[%v]=%v, got q[%v]=%v", i, x, i, h.q[i])
		}
	}
	min := 1.0
	max := 3.0
	Npoints := uint64(2)
	Nbins := uint64(20)
	q := NewP2Histogram(Nbins)
	q.Add(min)
	q.Add(max)
	expectedHistogram := []CumulativeDensity{{X: min, P: 1.0 / float64(Npoints)}, {X: max, P: 2.0 / float64(Npoints)}}
	if Npoints != q.N() {
		t.Errorf("Expected N: %v got %v", Npoints, q.N())
	}
	if min != q.Min() {
		t.Errorf("Expected Min: %v got %v", min, q.Min())
	}
	if max != q.Max() {
		t.Errorf("Expected Max: %v got %v", max, q.Max())
	}
	if reflect.DeepEqual(expectedHistogram, q.Histogram()) != true {
		t.Errorf("Expected Histogram: %v got %v", expectedHistogram, q.Histogram())
	}
	p := 0.75
	expectedQ := 2.0
	if p != q.CDF(expectedQ) {
		t.Logf("Expected p: %v got %v", p, q.Quantile(expectedQ))
	}
	if expectedQ != q.Quantile(p) {
		t.Logf("Expected x: %v got %v", expectedQ, q.Quantile(p))
	}
}

func TestP2DataPointsHistogram(t *testing.T) {
	q := NewP2Quantile(0.5)
	hist := NewP2Histogram(4)
	for i, x := range dataPoints {
		q.Add(x)
		hist.Add(x)
		for j := 0; j < 5; j++ {
			// check n
			if dataPointsExpected[i].n[j] != hist.n[j] {
				t.Errorf("Added data point[%v] %v, expected n[%v]=%v, got n[%v]=%v", i, x, j, dataPointsExpected[i].n[j], j, hist.n[j])
			}
			// check q
			if math.Abs(dataPointsExpected[i].q[j]-hist.q[j]) > 0.02 { // published table is only printed to 2 decimals and appears to use ceiling
				t.Errorf("Added data point[%v] %v, expected q[%v]=%v, got q[%v]=%v", i, x, j, dataPointsExpected[i].q[j], j, hist.q[j])
			}
		}
		if hist.N() != uint64(i+1) {
			t.Errorf("Expected the number of points to be %d got %d", i+1, hist.N())
		}
		if hist.Min() != q.Min() {
			t.Errorf("Expected Min to be %d got %d", q.Min(), hist.Min())
		}
		if hist.Max() != q.Max() {
			t.Errorf("Expected Max to be %d got %d", q.Max(), hist.Max())
		}
	}
}

func TestP2HistogramExponentialDist(t *testing.T) {
	rand.Seed(42) // for deterministic testing
	N := 100000
	Nbins := uint64(20)
	eps := 0.03 // expect errors less than 3% for all quantiles
	lambdas := []float64{1.0, 2.0, 0.5}
	ps := []float64{0.05, 0.10, 0.17, 0.25, 0.50, 0.65, 0.78, 0.95}
	for _, lambda := range lambdas {
		q := NewP2Histogram(Nbins)
		for i := 0; i < N; i++ {
			q.Add(exponentialRandomVariable(lambda))
		}
		for i, cd := range q.Histogram() {
			if i == 0 || uint64(i) == Nbins {
				// Don't expect max or min to match the given quantile for the distribution
				continue
			}
			Q := cd.X
			p := cd.P
			if q.Quantile(p) != Q {
				t.Errorf("For i: %d lambda: %v and p: %v Expected function Quantile %v, got %v", i, lambda, p, Q, q.Quantile(p))
			}
			if q.CDF(Q) != p {
				t.Errorf("For i: %d lambda: %v and Q: %v Expected function CDF %v, got %v", i, lambda, Q, p, q.CDF(Q))
			}
			if math.Abs((exponentialQuantile(p, lambda)-Q)/exponentialQuantile(p, lambda)) > eps {
				t.Errorf("For i: %d lambda: %v and p: %v Expected Histogram quantile %v, got %v", i, lambda, p, exponentialQuantile(p, lambda), Q)
			}
		}
		min := q.Min()
		max := q.Max()
		// test the bounds
		if q.Quantile(-1) != min {
			t.Errorf("Expected percentage <0 to return min: %v got %v", min, q.Quantile(-1))
		}
		if q.Quantile(2) != max {
			t.Errorf("Expected percentage >1 to return max: %v got %v", max, q.Quantile(2))
		}
		if q.CDF(min-1.0) != 0.0 {
			t.Errorf("Expected CDF of x < min to return 0.0, got %v", q.CDF(min-1.0))
		}
		if q.CDF(max+1.0) != 1.0 {
			t.Errorf("Expected  CDF of x > max to return 1.0, got %v", q.CDF(max+1.0))
		}
		for _, p := range ps {
			if math.Abs((exponentialQuantile(p, lambda)-q.Quantile(p))/exponentialQuantile(p, lambda)) > eps {
				t.Errorf("For lambda: %v and p: %v Expected Quantile %v, got %v", lambda, p, exponentialQuantile(p, lambda), q.Quantile(p))
			}
		}
	}
}

func BenchmarkP2Histogram8Add(b *testing.B) {
	q := NewP2Histogram(8)
	for i := 0; i < b.N; i++ {
		q.Add(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram16Add(b *testing.B) {
	q := NewP2Histogram(16)
	for i := 0; i < b.N; i++ {
		q.Add(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram32Add(b *testing.B) {
	q := NewP2Histogram(32)
	for i := 0; i < b.N; i++ {
		q.Add(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram64Add(b *testing.B) {
	q := NewP2Histogram(64)
	for i := 0; i < b.N; i++ {
		q.Add(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}

func BenchmarkP2Histogram128Add(b *testing.B) {
	q := NewP2Histogram(128)
	for i := 0; i < b.N; i++ {
		q.Add(gaussianTestData[i&mask])
	}
	result = q.Max() // to avoid optimizing out the loop entirely
}
