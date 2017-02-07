package streamstats

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func TestNewHyperLogLog(t *testing.T) {
	for p := byte(minimumHyperLogLogP); p <= maximumHyperLogLogP; p++ {

		hll := NewHyperLogLog(p, fnv.New64())
		m := uint64(1 << p)
		if uint64(len(hll.data)) != m {
			t.Errorf("Expected data to be length %d, got %d\n", m, len(hll.data))
		}
		expectedError := 1.04 / math.Sqrt(float64(m))
		if expectedError != hll.ExpectedError() {
			t.Errorf("Expected error to be %f, got %f\n", expectedError, hll.ExpectedError())
		}
	}
	hll := NewHyperLogLog(minimumHyperLogLogP-1, fnv.New64())
	if hll.p != minimumHyperLogLogP {
		t.Errorf("Expected minimum p of %d, got %d", minimumHyperLogLogP, hll.p)
	}
	hll = NewHyperLogLog(maximumHyperLogLogP+1, fnv.New64())
	if hll.p != maximumHyperLogLogP {
		t.Errorf("Expected maximum p of %d, got %d", maximumHyperLogLogP, hll.p)
	}
}

func TestHyperLogLogDistinctInts(t *testing.T) {
	p := byte(5)
	hll := NewHyperLogLog(p, fnv.New64())
	cardinality := uint64(1000000)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, i)
		hll.Add(b)
	}
	expectedError := hll.ExpectedError()
	actualError := math.Abs(float64(hll.Distinct())-float64(cardinality)) / float64(cardinality)
	if actualError > expectedError {
		t.Errorf("Expected cardinality %d, got %d\n", cardinality, hll.Distinct())
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}
	hll.Reset()
	for _, val := range hll.data {
		if val != 0 {
			t.Errorf("Expected reset to zero the data, got %0x", val)
		}
	}
}

func TestHyperLogLogDistinctPRNG(t *testing.T) {
	p := byte(5)
	hll := NewHyperLogLog(p, fnv.New64())
	cardinality := uint64(1000000)
	rand.Seed(42)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hll.Add(b)
	}
	expectedError := hll.ExpectedError()
	actualError := math.Abs(float64(hll.Distinct())-float64(cardinality)) / float64(cardinality)
	if actualError > expectedError {
		t.Errorf("Expected cardinality %d, got %d\n", cardinality, hll.Distinct())
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}
}

func TestHyperLogLogEstimates(t *testing.T) {
	p := byte(5)
	m := uint64(1 << p)
	hll := NewHyperLogLog(p, fnv.New64())
	rand.Seed(42)
	for i := uint64(0); i < uint64(p); i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hll.Add(b)
	}
	// in the low regime should use linear counting
	if hll.Distinct() != hll.LinearCounting() {
		t.Errorf("Expected HyperLogLog to use LinearCounting for small values")
	}
	for i := uint64(0); i < 2*m; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hll.Add(b)
	}
	// in the middle regime should use bias correction
	if hll.Distinct() != hll.BiasCorrected() {
		t.Errorf("Expected HyperLogLog to use BiasCorrected for middle values")
	}
	for i := uint64(0); i < 8*m; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hll.Add(b)
	}
	// in the high regime should use raw estimate
	if hll.Distinct() != hll.RawEstimate() {
		t.Errorf("Expected HyperLogLog to use RawEstimate for high values")
	}
}

func TestHyperLogLogDistinctReducePrecision(t *testing.T) {
	p := byte(7)
	hll := NewHyperLogLog(p, fnv.New64())
	m := byte(1 << p)
	// populate the hll with consecutive integers in the bins
	for i := byte(0); i < m; i++ {
		hll.data[i] = i
	}

	_, err := hll.ReducePrecision(p + 1)
	if err == nil {
		t.Errorf("Expected error when reduce precision attempted to reduce precision to a higher value than initial")
	}
	// reduce the precision
	newP := byte(4)
	reducedHll, err := hll.ReducePrecision(newP)
	if err != nil {
		t.Error(err)
	}
	newM := byte(1 << newP)
	stride := byte(1 << (p - newP))
	for i := byte(0); i < newM; i++ {
		if reducedHll.data[i] != (i+1)*stride-1 {
			t.Errorf("Expected max over the bin %d got %d", i*stride, reducedHll.data[i])
		}
	}

	// check reduce past minimum is an error
	reducedHll, err = hll.ReducePrecision(minimumHyperLogLogP - 1)
	if err == nil {
		t.Errorf("Expected error when reducing the precision below the minimum")
	}
}

func TestHyperLogLogCombine(t *testing.T) {
	// Expect to get exactly the same answer after combining
	p := byte(12)
	hllA := NewHyperLogLog(p, fnv.New64())
	hllB := NewHyperLogLog(p, fnv.New64())
	hllTotal := NewHyperLogLog(p, fnv.New64())

	cardinality := uint64(500)
	rand.Seed(42)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllA.Add(b)     // count in A
		hllTotal.Add(b) // count in Total
	}
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllB.Add(b)     // count in B
		hllTotal.Add(b) // count in Total
	}
	hllC, err := hllA.Combine(hllB) // A + B should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() != hllTotal.Distinct() {
		t.Errorf("Expected combined %d to equal total %d", hllC.Distinct(), hllTotal.Distinct())
	}

	// test combine with a reduction
	hllA = NewHyperLogLog(p, fnv.New64())
	hllB = NewHyperLogLog(p-3, fnv.New64())
	hllTotal = NewHyperLogLog(p-3, fnv.New64())

	rand.Seed(42)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllA.Add(b)     // count in A
		hllTotal.Add(b) // count in Total
	}
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllB.Add(b)     // count in B
		hllTotal.Add(b) // count in Total
	}
	hllC, err = hllA.Combine(hllB) // A + B should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() != hllTotal.Distinct() {
		t.Errorf("Expected combined %d to equal total %d", hllC.Distinct(), hllTotal.Distinct())
	}
	//combine in the opposite order
	hllC, err = hllB.Combine(hllA) // B + A should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() != hllTotal.Distinct() {
		t.Errorf("Expected combined %d to equal total %d", hllC.Distinct(), hllTotal.Distinct())
	}
	// Confirm that combining with different hash functions is an error
	hllB.hash = fnv.New64a()
	hllC, err = hllA.Combine(hllB) // A + B should equal total
	if err == nil {
		t.Errorf("Expected different hash functions to error on Combine")
	}

}

func BenchmarkHyperLogLogP10Add(b *testing.B) {
	p := byte(10)
	hll := NewHyperLogLog(p, fnv.New64())
	for i := 0; i < b.N; i++ {
		hll.Add(randomBytes[i&mask])
	}
	count = hll.Distinct() // to avoid optimizing out the loop entirely
}

func BenchmarkHyperLogLogP10Distinct(b *testing.B) {
	p := byte(10)
	hll := NewHyperLogLog(p, fnv.New64())
	for i := 0; i < 5*(1<<p); i++ {
		hll.Add(randomBytes[i&mask])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hll.Distinct()
	}
	count = hll.Distinct() // to avoid optimizing out the loop entirely
}
