package streamstats

import (
	"fmt"
	"hash"
	"math"
)

// BloomFilter is a datastructure for approximate set membership
// with no false negatives and limited false positives
type BloomFilter struct {
	hash hash.Hash64 // the base hash function
	bits BitVector   // the underlying occupied buckets
	k    uint64      // number of hash functions to calculate for each item
	m    uint64      // size of the BloomFilter in bits
}

// NewBloomFilter returns a pointer to a new BloomFilter that has been sized in m
// with k hash functions to target the given false positive rate
// at the given number of items using the given hash function
func NewBloomFilter(Nitems uint64, FalsePositiveRate float64, hash hash.Hash64) *BloomFilter {
	var k, m, optM uint64
	optM = uint64(-float64(Nitems) * math.Log(FalsePositiveRate) / (math.Ln2 * math.Ln2))
	if optM > (1 << 32) {
		m = 1 << 32 // maximum use is 32 bits of the 64 bit hash function
	} else {
		m = nextPowerOfTwo(optM)
	}
	bits := NewBitVector(m)
	k = uint64(float64(m)*math.Ln2/float64(Nitems) + 0.5) // add 0.5 to round properly
	return &BloomFilter{hash: hash, bits: bits, k: k, m: m}
}

// Add puts an item in the set represented by the BloomFilter
func (bf *BloomFilter) Add(item []byte) {
	bf.hash.Reset()
	bf.hash.Write(item)
	hash := bf.hash.Sum64()
	h1 := hash & ((1 << 32) - 1) // take the bottom 32 bits as the first hash
	h2 := hash >> 32             // take the top 32 bits as the second hash
	bf.bits.Set(h1 & (bf.m - 1))
	for i := uint64(1); i < bf.k; i++ {
		h1 += h2 // generate the k hash functions as h_i = h1 + i * h2 mod m
		bf.bits.Set(h1 & (bf.m - 1))
	}
}

// Check returns false if an item in is definitely not in the set represented by the BloomFilter
func (bf BloomFilter) Check(item []byte) bool {
	bf.hash.Reset()
	bf.hash.Write(item)
	hash := bf.hash.Sum64()
	h1 := hash & ((1 << 32) - 1)       // take the bottom 32 bits as the first hash
	h2 := hash >> 32                   // take the top 32 bits as the second hash
	if bf.bits.Get(h1&(bf.m-1)) != 1 { // if any bit is not set the item is not in the set
		return false
	}
	for i := uint64(1); i < bf.k; i++ {
		h1 += h2 // generate the k hash functions as h_i = h1 + i * h2 mod m
		if bf.bits.Get(h1&(bf.m-1)) != 1 {
			return false
		}
	}
	return true // all hash functions check out
}

// Occupancy returns the ratio of filled buckets in the BloomFilter
func (bf BloomFilter) Occupancy() float64 {
	return float64(bf.bits.PopCount()) / float64(bf.m)
}

// FalsePositiveRate returns the expected false positive rate based on the ratio of filled buckets in the BloomFilter
func (bf BloomFilter) FalsePositiveRate() float64 {
	return math.Pow(float64(bf.bits.PopCount())/float64(bf.m), float64(bf.k))
}

// Distinct estimates the number of elements in the filter by using the LinearCounting estimate accounting for the
// k hash functions calculated for each element
func (bf BloomFilter) Distinct() uint64 {
	return uint64(-(float64(bf.m) / float64(bf.k)) * math.Log(1-bf.Occupancy()))
}

// Union combines two BloomFilters producing one that contains all of the elements in either BloomFilter
// the BloomFilters must be the same size m and k as well as use the same hash function
func (bf BloomFilter) Union(bfB *BloomFilter) (*BloomFilter, error) {

	if bf.m != bfB.m {
		return nil, fmt.Errorf("BloomFilters do not have equal size m1 = %d != %d = m2", bf.m, bfB.m)
	}
	if bf.k != bfB.k {
		return nil, fmt.Errorf("BloomFilters do not have equal nubmer of hash functions k1 = %d != %d = k2", bf.k, bfB.k)
	}

	// check that both hash functions get the same result for "BloomFilter"
	bf.hash.Reset()
	bf.hash.Write([]byte("BloomFilter"))
	hash := bf.hash.Sum64()
	bfB.hash.Reset()
	bfB.hash.Write([]byte("BloomFilter"))
	hashB := bfB.hash.Sum64()
	if hash != hashB {
		return nil, fmt.Errorf("Hash functions are not identical, return %0x != %0x for \"BloomFilter\"", hash, hashB)
	}

	bits := NewBitVector(bf.m)
	for i := range bits {
		bits[i] = bf.bits[i] | bfB.bits[i]
	}

	return &BloomFilter{hash: bf.hash, bits: bits, m: bf.m, k: bf.k}, nil
}

// Intersect combines two BloomFilters producing one that contains only of the elements in both BloomFilters
// the BloomFilters must be the same size m and k as well as use the same hash function
func (bf BloomFilter) Intersect(bfB *BloomFilter) (*BloomFilter, error) {

	if bf.m != bfB.m {
		return nil, fmt.Errorf("BloomFilters do not have equal size m1 = %d != %d = m2", bf.m, bfB.m)
	}
	if bf.k != bfB.k {
		return nil, fmt.Errorf("BloomFilters do not have equal nubmer of hash functions k1 = %d != %d = k2", bf.k, bfB.k)
	}

	// check that both hash functions get the same result for "BloomFilter"
	bf.hash.Reset()
	bf.hash.Write([]byte("BloomFilter"))
	hash := bf.hash.Sum64()
	bfB.hash.Reset()
	bfB.hash.Write([]byte("BloomFilter"))
	hashB := bfB.hash.Sum64()
	if hash != hashB {
		return nil, fmt.Errorf("Hash functions are not identical, return %d != %d for \"BloomFilter\"", hash, hashB)
	}

	bits := NewBitVector(bf.m)
	for i := range bits {
		bits[i] = bf.bits[i] & bfB.bits[i]
	}

	return &BloomFilter{hash: bf.hash, bits: bits, m: bf.m, k: bf.k}, nil
}

// nextPowerOfTwo returns the next greater power of two for a given input
func nextPowerOfTwo(x uint64) uint64 {
	if x == 0 {
		return 1
	}
	x--         // if we start on a power of two go down by one
	x |= x >> 1 // "fold" the bits over to get a string of all ones
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	return x + 1 // increment to get the next power of two
}
