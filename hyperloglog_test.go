package streamstats

import (
	"encoding/binary"
	"fmt"
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
	N := hll.Distinct()
	expectedError := hll.ExpectedError()
	delta := uint64(float64(N) * expectedError)
	actualError := math.Abs(float64(N)-float64(cardinality)) / float64(cardinality)
	if actualError > expectedError {
		t.Errorf("Expected cardinality %d, got %d\n", cardinality, N)
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}
	expectedString := fmt.Sprintf("HyperLogLog N: %d +/- %d", N, delta)
	if hll.String() != expectedString {
		t.Errorf("Expected string %s, got %s\n", hll, expectedString)
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

func TestHyperLogLogCompress(t *testing.T) {
	p := byte(7)
	hll := NewHyperLogLog(p, fnv.New64())
	m := byte(1 << p)
	// populate the hll with consecutive integers in the bins
	for i := byte(0); i < m; i++ {
		hll.data[i] = i
	}

	// reduce the precision
	factor := byte(3)
	reducedHll := hll.Compress(factor)
	if reducedHll.p != p-factor {
		t.Errorf("Expected compressed HyperLogLog to have p=%d got %d", p-factor, reducedHll.p)
	}

	newM := byte(p >> factor)
	stride := factor
	for i := byte(0); i < newM; i++ {
		if reducedHll.data[i] != (i+1)*stride-1 {
			t.Errorf("Expected max over the bin %d got %d", i*stride, reducedHll.data[i])
		}
	}

	// check reduce past minimum is an error
	reducedHll = hll.Compress(p + 3)
	if reducedHll.p != minimumHyperLogLogP {
		t.Errorf("Expected minimum HyperLogLog compression %d got %d", minimumHyperLogLogP, reducedHll.p)
	}
}

func TestHyperLogLogUnion(t *testing.T) {
	// Expect to get exactly the same answer after combining
	p := byte(12)
	hllA := NewHyperLogLog(p, fnv.New64())
	hllB := NewHyperLogLog(p, fnv.New64())
	hllb := NewHyperLogLog(p-3, fnv.New64())
	hllUnion := NewHyperLogLog(p, fnv.New64())
	hllIntersect := NewHyperLogLog(p, fnv.New64())

	cardinality := uint64(500)
	rand.Seed(42)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllA.Add(b)     // count in A
		hllUnion.Add(b) // count in Union
	}
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllA.Add(b)         // count in A
		hllB.Add(b)         // count in B
		hllb.Add(b)         // count in b
		hllUnion.Add(b)     // count in Union
		hllIntersect.Add(b) // count in Intersect
	}
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		hllB.Add(b)     // count in B
		hllb.Add(b)     // count in b
		hllUnion.Add(b) // count in Union
	}
	hllC, err := hllA.Union(hllB) // A | B should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() != hllUnion.Distinct() {
		t.Errorf("Expected union %d to equal total %d", hllC.Distinct(), hllUnion.Distinct())
	}
	hllC, err = hllA.Intersect(hllB) // A & B should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() < hllIntersect.Distinct() {
		t.Errorf("Expected intersect %d to count at least as many as true intersect %d", hllC.Distinct(), hllIntersect.Distinct())
	}
	// test combining with a compression
	hllUnionb := hllUnion.Compress(3)
	hllIntersectb := hllIntersect.Compress(3)
	hllC, err = hllA.Union(hllb) // A | B should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() != hllUnionb.Distinct() {
		t.Errorf("Expected Union %d to equal total %d", hllC.Distinct(), hllUnionb.Distinct())
	}

	hllC, err = hllA.Intersect(hllb) // A & B should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() < hllIntersectb.Distinct() {
		t.Errorf("Expected Intersect %d to equal total %d", hllC.Distinct(), hllIntersectb.Distinct())
	}

	// union in the opposite order
	hllC, err = hllb.Union(hllA) // B | A should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() != hllUnionb.Distinct() {
		t.Errorf("Expected Union %d to equal total %d", hllC.Distinct(), hllUnionb.Distinct())
	}

	hllC, err = hllb.Intersect(hllA) // B & A should equal total
	if err != nil {
		t.Error(err)
	}
	if hllC.Distinct() < hllIntersectb.Distinct() {
		t.Errorf("Expected Intersect %d to exceed total %d", hllC.Distinct(), hllIntersectb.Distinct())
	}

	// Confirm that combining with different hash functions is an error
	hllB.hash = fnv.New64a()
	hllC, err = hllA.Union(hllB) // A + B should equal total
	if err == nil {
		t.Errorf("Expected different hash functions to error on Union")
	}
	hllC, err = hllA.Intersect(hllB) // A + B should equal total
	if err == nil {
		t.Errorf("Expected different hash functions to error on Intersect")
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
