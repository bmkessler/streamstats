package streamstats

import (
	"hash/fnv"
	"math/rand"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	var maxItems, expectedM, expectedK uint64
	var fpr float64

	maxItems = 107
	fpr = 0.0101

	expectedM = 1024
	expectedK = 7

	bf := NewBloomFilter(maxItems, fpr, fnv.New64())
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
	var falsePositives, samples uint64 // check 1000 items that weren't in the filter
	samples = 1000
	for i := uint64(0); i < samples; i++ {
		b := make([]byte, 8)
		rand.Read(b)
		if bf.Check(b) {
			falsePositives++
		}
	}
	measuredFPR := float64(falsePositives) / float64(samples)
	if measuredFPR > fpr {
		t.Errorf("Measured false positive rate %f using %d samples exceeds target false positive rate %f", measuredFPR, samples, fpr)
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
