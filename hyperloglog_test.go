package streamstats

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func TestNewHyperLogLog(t *testing.T) {
	for p := byte(4); p <= byte(16); p++ {

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
}

func BenchmarkHyperLogLogAdd(b *testing.B) {
	p := byte(10)
	hll := NewHyperLogLog(p, fnv.New64())
	for i := 0; i < b.N; i++ {
		hll.Add(randomBytes[i%N])
	}
	count = hll.Distinct() // to avoid optimizing out the loop entirely
}
