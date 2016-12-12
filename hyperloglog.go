package streamstats

import (
	"hash"
	"math"
	"sync"
)

import "fmt"

// HyperLogLog a data structure for computing count distinct on arbitrary sized data
type HyperLogLog struct {
	sync.RWMutex
	hash  hash.Hash64
	alpha float64
	p     byte
	data  []byte
}

// NewHyperLogLog returns a new HyperLogLog data structure with 2^p buckets based on
// Hyperloglog: The analysis of a near-optimal cardinality estimation algorithm
// Philippe Flajolet and Éric Fusy and Olivier Gandouet and et al.
// IN AOFA ’07: PROCEEDINGS OF THE 2007 INTERNATIONAL CONFERENCE ON ANALYSIS OF ALGORITHMS
// This implementation does not include any of the HyperLogLog++ enhancments except for the 64-bit hash function
// which eliminates the large cardinality correction for hash collisions
// this is also space in-efficient since bytes are used to store the counts which could be at most 60 < 2^6
func NewHyperLogLog(p byte, hash hash.Hash64) *HyperLogLog {
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
		hash:  hash,
		alpha: alpha,
		p:     p,
		data:  make([]byte, m, m),
	}
}

// Add adds an item to the multiset represented by the HyperLogLog
func (hll *HyperLogLog) Add(item []byte) {
	hll.Lock()
	defer hll.Unlock()

	hll.hash.Reset()
	hll.hash.Write(item)
	hash := hll.hash.Sum64()
	bucket := hash >> (64 - hll.p) // top p bits are the bucket
	trailingZeroCount := byte(1)   // the cardinality estimate based on number of zeros
	for k := 1; int(hash&uint64(1)) != 1 && k <= int((64-hll.p)); k++ {
		trailingZeroCount = byte(k) + 1
		hash = hash >> 1
	}
	// if the new estimate for the bucket is larger update it
	if trailingZeroCount > hll.data[bucket] {
		hll.data[bucket] = trailingZeroCount
	}
}

// Distinct returns the estimated number of distinct items in the multiset
func (hll *HyperLogLog) Distinct() uint64 {
	hll.RLock()
	defer hll.RUnlock()

	m := uint64(1 << hll.p)
	var sum float64
	for _, d := range hll.data {
		sum += math.Pow(2.0, -1.0*float64(d))
	}
	rawEstimate := hll.alpha * float64(m) * float64(m) / sum
	estimate := uint64(rawEstimate)
	if rawEstimate < 5.0*float64(m)/2 {
		zeroCount := 0
		for _, d := range hll.data {
			if d == 0 {
				zeroCount++
			}
		}
		if zeroCount > 0 { // Use the linear counting estimate
			estimate = uint64(float64(m) * math.Log(float64(m)/float64(zeroCount)))
		}
	}
	return estimate
}

// ExpectedError returns the estimated error in the number of distinct items in the multiset
func (hll *HyperLogLog) ExpectedError() float64 {
	hll.RLock()
	defer hll.RUnlock()

	m := uint64(1 << hll.p)
	return 1.04 / math.Sqrt(float64(m))

}

// Reset zeros out the estimated number of distinct items in the multiset
func (hll *HyperLogLog) Reset() {
	hll.Lock()
	defer hll.Unlock()

	for i := range hll.data {
		hll.data[i] = 0
	}
}

// ReducePrecision produces a new HyperLogLog with reduced precision
// if p>hll.p it returns nil and an error, if p==hll.p it just produces a copy
func (hll *HyperLogLog) ReducePrecision(p byte) (*HyperLogLog, error) {
	hll.RLock()
	defer hll.RUnlock()

	if p > hll.p {
		return nil, fmt.Errorf("Precision %d is greater than the current HyperLogLog precision %d", p, hll.p)
	}
	newHLL := NewHyperLogLog(p, hll.hash)
	// TODO populate new hll by taking max over the stride length
	m := (1 << hll.p)
	newM := (1 << p)
	strideLength := m - newM
	for i := 0; i < newM; i++ {
		for j := 0; j < strideLength; j++ {
			if newHLL.data[i] < hll.data[i*strideLength+j] {
				newHLL.data[i] = hll.data[i*strideLength+j]
			}
		}
	}
	return newHLL, nil
}

// Combine the estimate of two HyperLogLog reducing the precision to the minimum of the two sets
// the function will return nil and an error if the hash functions mismatch
func (hll *HyperLogLog) Combine(hllB *HyperLogLog) (*HyperLogLog, error) {
	hll.RLock()
	hllB.RLock()
	defer hll.RUnlock()
	defer hllB.RUnlock()

	// check that both hash functions get the same result for "HyperLogLog"
	hll.hash.Reset()
	hll.hash.Write([]byte("HyperLogLog"))
	hash := hll.hash.Sum64()
	hllB.hash.Reset()
	hllB.hash.Write([]byte("HyperLogLog"))
	hashB := hllB.hash.Sum64()
	if hash != hashB {
		return nil, fmt.Errorf("Hash functions are not identical, return %d != %d for \"HyperLogLog\"", hash, hashB)
	}
	// determine if either precision needs to be reduced
	var combinedP byte
	var hll1, hll2, combinedHLL *HyperLogLog
	if hll.p < hllB.p {
		combinedP = hll.p
		hll1 = hll
		hll2, _ = hllB.ReducePrecision(hll.p)
	} else if hllB.p < hll.p {
		combinedP = hllB.p
		hll1, _ = hll.ReducePrecision(hllB.p)
		hll2 = hllB
	} else {
		combinedP = hll.p
		hll1 = hll
		hll2 = hllB
	}
	// for each bucket take the max value from the two Hyperloglog
	combinedHLL = NewHyperLogLog(combinedP, hll.hash)
	for i := range combinedHLL.data {
		if hll1.data[i] > hll2.data[i] {
			combinedHLL.data[i] = hll1.data[i]
		} else {
			combinedHLL.data[i] = hll2.data[i]
		}
	}
	return combinedHLL, nil
}
