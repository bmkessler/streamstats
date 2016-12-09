package streamstats

import "hash/fnv"

// a 64-bit hash of a byte slice using the built-in FNV hash function
func hash64(data []byte) uint64 {
	hasher := fnv.New64()
	hasher.Write(data)
	return hasher.Sum64()
}

// HyperLogLog a data structure for computing count distinct on arbitrary sized data
type HyperLogLog struct {
	alpha float64
	p     byte
	data  []byte
}

// NewHyperLogLog returns a new HyperLogLog data structure with 2^p buckets based on
// Hyperloglog: The analysis of a near-optimal cardinality estimation algorithm
// Philippe Flajolet and Éric Fusy and Olivier Gandouet and et al.
// IN AOFA ’07: PROCEEDINGS OF THE 2007 INTERNATIONAL CONFERENCE ON ANALYSIS OF ALGORITHMS
func NewHyperLogLog(p byte) *HyperLogLog {
	// p is bounded by 4 and 16 for practical implementations
	if p < 4 {
		p = 4
	} else if p > 16 {
		p = 16
	}
	m := 1 << p
	var alpha float64 // the normalization constant dependent on m
	switch {
	case m == 16:
		alpha = 0.673
	case m == 32:
		alpha = 0.697
	case m == 64:
		alpha = 0.709
	default:
		alpha = 0.7213 / (1 + 1.079/float64(m))
	}
	return &HyperLogLog{
		alpha: alpha,
		p:     p,
		data:  make([]byte, m, m),
	}
}

// Add adds an item to the multiset represented by the HyperLogLog
func (hll *HyperLogLog) Add(item []byte) {
	hash := hash64(item)
	bucket := hash >> (64 - hll.p) // top p bits are the bucket
	w := hash                      // the cardinality estimate based on number of zeros
	trailingZeroCount := byte(1)
	for k := 1; k < 8 && int(w&uint64(1)) != 1; k++ {
		trailingZeroCount = byte(k)
		w = w >> 1
	}
	// if the new estimate for the bucket is larger update it
	if trailingZeroCount > hll.data[bucket] {
		hll.data[bucket] = trailingZeroCount
	}
}

// Distinct returns the estimated number of distinct items in the multiset
func (hll HyperLogLog) Distinct() uint64 {

}
