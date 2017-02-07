package streamstats

import (
	"encoding/binary"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func TestLinearCountingPRNG(t *testing.T) {
	p := byte(13)
	lc := NewLinearCounting(p, fnv.New64())
	cardinality := uint64(1000)
	rand.Seed(42)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		lc.Add(b)
	}
	m := float64(uint64(1 << p))
	loadFactor := float64(cardinality) / m
	expectedError := 2 * math.Sqrt((math.Exp(loadFactor)-loadFactor-1)/m) / loadFactor
	actualError := math.Abs(float64(lc.Distinct())-float64(cardinality)) / float64(cardinality)
	if actualError > expectedError {
		t.Errorf("Expected cardinality %d, got %d\n", cardinality, lc.Distinct())
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}

	// Make a small LinearCounting and fill it completely
	p = byte(6)
	lc = NewLinearCounting(p, fnv.New64())
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		lc.Add(b)
	}
	if lc.Occupancy() != 1.0 {
		t.Errorf("Expected LinearCounting to be full, got occupancy %f", lc.Occupancy())
	}
	if lc.Distinct() != (1 << p) {
		t.Errorf("Expected LinearCounting to saturate at %d, got %d", (1 << p), lc.Distinct())
	}

	// test very small and very large linear counting are bounded
	lc = NewLinearCounting(4, fnv.New64())
	if lc.p != minLinearCountingP {
		t.Errorf("Expected minimum linear counting size %d, got %d", minLinearCountingP, lc.p)
	}
	lc = NewLinearCounting(26, fnv.New64())
	if lc.p != maxLinearCountingP {
		t.Errorf("Expected maximum linear counting size %d, got %d", maxLinearCountingP, lc.p)
	}
}

func TestLinearCountingVsHyperLogLog(t *testing.T) {
	// Expect to get exactly the same answer for the same algorithm
	p := byte(13)
	lc := NewLinearCounting(p, fnv.New64())
	hll := NewHyperLogLog(p, fnv.New64())
	cardinality := uint64(1234)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, i)
		lc.Add(b)
		hll.Add(b)
		if lc.Distinct() != hll.LinearCounting() {
			t.Errorf("%d: Expected LinearCounting %d and HyperLogLog %d to be equal", i, lc.Distinct(), hll.LinearCounting())
		}
	}
}

func TestLinearCountingReducePrecision(t *testing.T) {
	// Expect to get exactly the same answer after folding
	p := byte(8)
	lc := NewLinearCounting(p, fnv.New64())
	// interleave the bits
	for i := uint64(0); i < 64-4; i += 4 {
		lc.bits.Set(i)
		lc.bits.Set(64 + i + 1)
		lc.bits.Set(128 + i + 2)
		lc.bits.Set(196 + i + 3)
	}
	lcRed, err := lc.ReducePrecision(byte(6))
	if err != nil {
		t.Error(err)
	}
	if lc.bits.PopCount() != lcRed.bits.PopCount() {
		t.Errorf("PopCount: %d Reduced PopCount: %d", lc.bits.PopCount(), lcRed.bits.PopCount())
	}
	// collide the bits from each word
	lc = NewLinearCounting(p, fnv.New64())
	for i := uint64(0); i < (1 << p); i += 4 {
		lc.bits.Set(i)
	}
	lcRed, err = lc.ReducePrecision(byte(7))
	if err != nil {
		t.Error(err)
	}
	if lc.bits.PopCount() != lcRed.bits.PopCount()*2 {
		t.Errorf("PopCount: %d Reduced PopCount: %d", lc.bits.PopCount(), lcRed.bits.PopCount())
	}
	lcRed, err = lc.ReducePrecision(byte(6))
	if err != nil {
		t.Error(err)
	}
	if lc.bits.PopCount() != lcRed.bits.PopCount()*4 {
		t.Errorf("PopCount: %d Reduced PopCount: %d", lc.bits.PopCount(), lcRed.bits.PopCount())
	}
	// test reduce precision bounds
	lcRed, err = lc.ReducePrecision(minLinearCountingP - 1)
	if err == nil {
		t.Errorf("Expected error when trying to reduce below the minimum precision")
	}
	lcRed, err = lc.ReducePrecision(p + 1)
	if err == nil {
		t.Errorf("Expected error when trying to reduce above the starting precision")
	}
}

func TestLinearCountingCombine(t *testing.T) {
	// Expect to get exactly the same answer after combining
	p := byte(12)
	lcA := NewLinearCounting(p, fnv.New64())
	lcB := NewLinearCounting(p, fnv.New64())
	lcTotal := NewLinearCounting(p, fnv.New64())

	cardinality := uint64(500)
	rand.Seed(42)
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		lcA.Add(b)     // count in A
		lcTotal.Add(b) // count in Total
	}
	for i := uint64(0); i < cardinality; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		lcB.Add(b)     // count in B
		lcTotal.Add(b) // count in Total
	}
	lcC, err := lcA.Combine(lcB) // A + B should equal Total
	if err != nil {
		t.Error(err)
	}
	if lcC.Distinct() != lcTotal.Distinct() {
		t.Errorf("Expected combined %d to equal total %d", lcC.Distinct(), lcTotal.Distinct())
	}

	// test combine with a reduction
	lcA = NewLinearCounting(p, fnv.New64())
	lcB = NewLinearCounting(p-3, fnv.New64())
	lcTotal = NewLinearCounting(p-3, fnv.New64())

	cardinality = uint64(500)
	rand.Seed(42)
	for i := uint64(0); i < cardinality/2; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		lcA.Add(b)     // count in A
		lcTotal.Add(b) // count in Total
	}
	for i := uint64(0); i < cardinality/2; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		lcB.Add(b)     // count in B
		lcTotal.Add(b) // count in Total
	}
	lcC, err = lcA.Combine(lcB) // A + B should equal Total
	if err != nil {
		t.Error(err)
	}
	// test combine in opposite order
	lcD, err := lcB.Combine(lcA) // B + A should equal Total
	if err != nil {
		t.Error(err)
	}
	if lcC.Distinct() != lcD.Distinct() {
		t.Errorf("Expected combined %d to equal reverse %d", lcC.Distinct(), lcD.Distinct())
	}
	// check error is still within expectation
	m := float64(uint64(1 << (p - 3))) // reduced p
	loadFactor := float64(cardinality) / m
	expectedError := 2 * math.Sqrt((math.Exp(loadFactor)-loadFactor-1)/m) / loadFactor
	actualError := math.Abs(float64(lcC.Distinct())-float64(cardinality)) / float64(cardinality)
	if actualError > expectedError {
		t.Errorf("Expected cardinality %d, got %d\n", cardinality, lcC.Distinct())
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}

	// check two different hash functions fail
	lcA = NewLinearCounting(p, fnv.New64())
	lcB = NewLinearCounting(p, fnv.New64a())
	// A + B should equal Total
	if _, err = lcA.Combine(lcB); err == nil {
		t.Errorf("Expected using two different hash functions to return error")
	}
}

func BenchmarkLinearCountingP10Add(b *testing.B) {
	p := byte(10)
	lc := NewLinearCounting(p, fnv.New64())
	for i := 0; i < b.N; i++ {
		lc.Add(randomBytes[i&mask])
	}
	count = lc.Distinct() // to avoid optimizing out the loop entirely
}

func BenchmarkLinearCountingP10Distinct(b *testing.B) {
	p := byte(10)
	lc := NewLinearCounting(p, fnv.New64())
	for i := 0; i < 5*(1<<p); i++ {
		lc.Add(randomBytes[i&mask])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Distinct()
	}
	count = lc.Distinct() // to avoid optimizing out the loop entirely
}
