package streamstats

import (
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	var maxItems, expectedM, expectedK uint64
	var targetFalsePositiveRate float64

	maxItems = 107
	targetFalsePositiveRate = 0.0101

	expectedM = 1024 // -maxItems * ln(targetFalsePositiveRate)/(ln(2)^2)
	expectedK = 7    // (m/maxItems) * ln(2)

	bf := NewBloomFilter(maxItems, targetFalsePositiveRate, fnv.New64())
	if bf.m != expectedM {
		t.Errorf("Expected m to be %d, got %d\n", expectedM, bf.m)
	}
	if bf.k != expectedK {
		t.Errorf("Expected k to be %d, got %d\n", expectedK, bf.k)
	}

	rand.Seed(42) // fill the BloomFilter to the expected number of items
	for i := uint64(0); i < maxItems; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		bf.Add(b)
	}
	rand.Seed(42) // reset and check that all of those elements are in the BloomFilter
	for i := uint64(0); i < maxItems; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		if bf.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter", i)
		}
	}
	// check 1000 items that weren't in the filter
	var falsePositives, samples uint64
	samples = 1000
	for i := uint64(0); i < samples; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		if bf.Check(b) {
			falsePositives++
		}
	}
	measuredFPR := float64(falsePositives) / float64(samples)
	if measuredFPR > targetFalsePositiveRate {
		t.Errorf("Measured false positive rate %f using %d samples exceeds target false positive rate %f", measuredFPR, samples, targetFalsePositiveRate)
	}
	expectedFalsePositiveRate := bf.FalsePositiveRate()
	if measuredFPR > expectedFalsePositiveRate {
		t.Errorf("Measured false positive rate %f using %d samples exceeds expected false positive rate %f", measuredFPR, samples, expectedFalsePositiveRate)
	}
	estimatedItems := bf.Distinct()
	loadFactor := bf.Occupancy()
	expectedError := 2 * math.Sqrt((math.Exp(loadFactor)-loadFactor-1)*float64(bf.k)/float64(bf.m)) / loadFactor
	actualError := math.Abs(float64(estimatedItems)-float64(maxItems)) / float64(maxItems)
	if actualError > expectedError {
		t.Errorf("Expected cardinality %d, got %d\n", maxItems, estimatedItems)
		t.Errorf("Expected error %f, got %f\n", expectedError, actualError)
	}
}

func TestBloomFilterCombine(t *testing.T) {
	var maxItems uint64
	var targetFalsePositiveRate float64

	maxItems = 300
	targetFalsePositiveRate = 0.05

	bfA := NewBloomFilter(maxItems, targetFalsePositiveRate, fnv.New64())
	bfB := NewBloomFilter(maxItems, targetFalsePositiveRate, fnv.New64())
	bfC := NewBloomFilter(maxItems, targetFalsePositiveRate, fnv.New64()) // the Union
	bfD := NewBloomFilter(maxItems, targetFalsePositiveRate, fnv.New64()) // the Intersection

	rand.Seed(42)                             // fill the BloomFilters with the first third in A & C
	for i := uint64(0); i < maxItems/3; i++ { // fill the BloomFilters with the first third in A & C
		b := make([]byte, 8)
		rand.Read(b)
		bfA.Add(b)
		bfC.Add(b)
	}
	for i := uint64(0); i < maxItems/3; i++ { // fill the BloomFilters with the middle third in A, B, C & D
		b := make([]byte, 8)
		rand.Read(b)
		bfA.Add(b)
		bfB.Add(b)
		bfC.Add(b)
		bfD.Add(b)
	}
	for i := uint64(0); i < maxItems/3; i++ { // fill the BloomFilters with the last third in B & C
		b := make([]byte, 8)
		rand.Read(b)
		bfB.Add(b)
		bfC.Add(b)
	}
	bfUnion, err := bfA.Union(bfB)
	if err != nil {
		t.Errorf("BloomFilter Union failed: %s", err)
	}
	bfIntersect, err := bfA.Intersect(bfB)
	if err != nil {
		t.Errorf("BloomFilter Intersect failed: %s", err)
	}
	for i := range bfUnion.bits {
		if bfUnion.bits[i] != bfC.bits[i] {
			t.Errorf("Expected Union BloomFilter bits[%d] = %0x to equal %0x from C", i, bfUnion.bits[i], bfC.bits[i])
		}
	}

	var AFalsePositives, BFalsePositives, DFalsePositives, IntersectFalsePositives uint64 // false positive rates for each filter

	rand.Seed(42)                             // reset and check that all of those elements are in the BloomFilters
	for i := uint64(0); i < maxItems/3; i++ { // the initial third
		b := make([]byte, 8)
		rand.Read(b)
		if bfA.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter A", i)
		}
		if bfB.Check(b) != false {
			BFalsePositives++
		}
		if bfC.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter C", i)
		}
		if bfD.Check(b) != false {
			DFalsePositives++
		}
		if bfUnion.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the Union filter", i)
		}
		if bfIntersect.Check(b) != false {
			IntersectFalsePositives++
		}
	}
	for i := uint64(0); i < maxItems/3; i++ { // the middle third
		b := make([]byte, 8)
		rand.Read(b)
		if bfA.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter A", i)
		}
		if bfB.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter B", i)
		}
		if bfC.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter C", i)
		}
		if bfD.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter D", i)
		}
		if bfUnion.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the Union filter", i)
		}
		if bfIntersect.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the Intersect filter", i)
		}
	}
	for i := uint64(0); i < maxItems/3; i++ { // the final third
		b := make([]byte, 8)
		rand.Read(b)
		if bfA.Check(b) != false {
			AFalsePositives++
		}
		if bfB.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter B", i)
		}
		if bfC.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the filter C", i)
		}
		if bfD.Check(b) != false {
			DFalsePositives++
		}
		if bfUnion.Check(b) != true {
			t.Errorf("Expected element %d with seed 42 to be in the Union filter", i)
		}
		if bfIntersect.Check(b) != false {
			IntersectFalsePositives++
		}
	}
	var maxFalsePositives uint64
	if AFalsePositives > BFalsePositives {
		maxFalsePositives = AFalsePositives
	} else {
		maxFalsePositives = BFalsePositives
	}
	if IntersectFalsePositives > maxFalsePositives {
		t.Errorf("Intersect Filter had greater false positives %d than the individual filters A: %d B: %d", IntersectFalsePositives, AFalsePositives, BFalsePositives)
	}
	if IntersectFalsePositives < DFalsePositives {
		t.Errorf("Intersect Filter had lass false positives %d than the individual filter D: %d", IntersectFalsePositives, DFalsePositives)
	}
	// test different m, k and hash functions fail to Union and Intersect
	bfB.hash = fnv.New64a()
	bfC.m = 13
	bfD.k = 17
	if _, err = bfA.Union(bfB); err == nil {
		t.Errorf("Expected Union using two different hash functions to return error")
	}
	if _, err = bfA.Intersect(bfB); err == nil {
		t.Errorf("Expected Intersect using two different hash functions to return error")
	}
	if _, err = bfA.Union(bfC); err == nil {
		t.Errorf("Expected Union using two different m")
	}
	if _, err = bfA.Intersect(bfC); err == nil {
		t.Errorf("Expected Intersect using two different m")
	}
	if _, err = bfA.Union(bfD); err == nil {
		t.Errorf("Expected Union using two different k")
	}
	if _, err = bfA.Intersect(bfD); err == nil {
		t.Errorf("Expected Intersect using two different k")
	}
}

func TestNextPowerOfTwo(t *testing.T) {
	var testCases = []struct {
		in   uint64
		want uint64
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{6, 8},
		{7, 8},
		{8, 8},
		{(1 << 4) - 1, (1 << 4)},
		{(1 << 4), (1 << 4)},
		{(1 << 4) + 1, (1 << 5)},
		{(1 << 47) - 1, (1 << 47)},
		{(1 << 47), (1 << 47)},
		{(1 << 47) + 1, (1 << 48)},
	}
	for _, test := range testCases {
		got := nextPowerOfTwo(test.in)
		if got != test.want {
			t.Errorf("Expected %d to have the next power of two %d got %d", test.in, test.want, got)
		}
	}

}

func BenchmarkBloomFilterAdd(b *testing.B) {
	var maxItems uint64
	var fpr float64
	maxItems = 10000
	fpr = 0.03
	bf := NewBloomFilter(maxItems, fpr, fnv.New64())
	for i := 0; i < b.N; i++ {
		bf.Add(randomBytes[i&mask])
	}
	if bf.Check([]byte{}) {
		count = 5
	} // to avoid optimizing out the loop entirely
}

func BenchmarkBloomFilterCheck(b *testing.B) {
	var maxItems uint64
	var fpr float64
	maxItems = 10000
	fpr = 0.03
	bf := NewBloomFilter(maxItems, fpr, fnv.New64())
	for i := uint64(0); i < maxItems; i++ {
		bf.Add(randomBytes[i&mask])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Check(randomBytes[i&mask])
	}
	if bf.Check([]byte{}) {
		count = 5
	} // to avoid optimizing out the loop entirely
}
