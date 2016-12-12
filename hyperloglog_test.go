package streamstats

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func TestNewHyperLogLog(t *testing.T) {
	p := byte(5)
	m := uint64(1 << p)
	hll := NewHyperLogLog(p, fnv.New64())

	if uint64(len(hll.data)) != m {
		t.Errorf("Expected data to be length %d, got %d\n", m, len(hll.data))
	}
	expectedError := 1.04 / math.Sqrt(float64(m))
	if expectedError != hll.ExpectedError() {
		t.Errorf("Expected error to be %f, got %f\n", expectedError, hll.ExpectedError())
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
