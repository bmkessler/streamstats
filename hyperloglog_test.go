package streamstats

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"testing"
)

func TestNewHyperLogLog(t *testing.T) {
	p := byte(5)
	m := uint64(1 << p)
	hll := NewHyperLogLog(p, fnv.New64())
	cardinality := uint64(10000)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(i))
		hll.Add(b)
	}
	expectedError := 1.04 / math.Sqrt(float64(m))
	actualError := math.Abs(float64(hll.Distinct())-float64(cardinality)) / float64(cardinality)
	if actualError > expectedError {
		t.Log(hll.data)
		t.Errorf("Expected cardinality %d, got %d\n", cardinality, hll.Distinct())
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}
}
