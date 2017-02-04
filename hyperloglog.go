package streamstats

import (
	"hash"
	"math"
)

import "fmt"

// HyperLogLog a data structure for computing count distinct on arbitrary sized data
type HyperLogLog struct {
	hash  hash.Hash64
	alpha float64
	p     byte
	data  []byte
}

const (
	minimumHyperLogLogP = 4
	maximumHyperLogLogP = 16
)

// NewHyperLogLog returns a new HyperLogLog data structure with 2^p buckets based on
// Hyperloglog: The analysis of a near-optimal cardinality estimation algorithm
// Philippe Flajolet and Éric Fusy and Olivier Gandouet and et al.
// IN AOFA ’07: PROCEEDINGS OF THE 2007 INTERNATIONAL CONFERENCE ON ANALYSIS OF ALGORITHMS
// This implementation does not include any of the HyperLogLog++ enhancments except for the 64-bit hash function
// which eliminates the large cardinality correction for hash collisions
// this is also space in-efficient since bytes are used to store the counts which could be at most 60 < 2^6
func NewHyperLogLog(p byte, hash hash.Hash64) *HyperLogLog {
	// p is bounded by 4 and 16 for practical implementations
	if p < minimumHyperLogLogP {
		p = minimumHyperLogLogP
	} else if p > maximumHyperLogLogP {
		p = maximumHyperLogLogP
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

	alpha := hll.alpha
	m := float64(uint64(1 << hll.p))
	C := alpha * m
	var sum, zeroCount float64
	for _, d := range hll.data {
		sum += inversePowersOfTwo[int(d)]
		if d == 0 {
			zeroCount++
		}
	}
	rawEstimate := alpha * m * m / sum
	t := (rawEstimate - C) / C
	if t < 1.0 && zeroCount > 0 {
		// Use the linear counting estimate at low values because it has less variance
		rawEstimate = m * math.Log(m/float64(zeroCount))
	} else if t < 12.0 {
		// apply an empirical bias correction to intermediate values
		rawEstimate = rawEstimate - C*(math.Exp(-t)+0.125*t*(t-0.82)*math.Exp(-1.85*t))
	}
	return uint64(rawEstimate)
}

// LinearCounting returns the linear counting estimated number of distinct items in the multiset
func (hll *HyperLogLog) LinearCounting() uint64 {

	m := float64(uint64(1 << hll.p))
	zeroCount := 0
	for _, d := range hll.data {
		if d == 0 {
			zeroCount++
		}
	}
	return uint64(m * math.Log(m/float64(zeroCount)))
}

// RawEstimate returns the raw estimated number of distinct items in the multiset
func (hll *HyperLogLog) RawEstimate() uint64 {

	m := float64(uint64(1 << hll.p))
	var sum float64
	for _, d := range hll.data {
		sum += math.Pow(2.0, -1.0*float64(d))
	}
	return uint64(hll.alpha * m * m / sum)
}

// BiasCorrected returns the bias corrected estimated number of distinct items in the multiset
func (hll *HyperLogLog) BiasCorrected() uint64 {

	alpha := hll.alpha
	m := float64(uint64(1 << hll.p))
	C := alpha * m

	var sum float64
	for _, d := range hll.data {
		sum += math.Pow(2.0, -1.0*float64(d))
	}
	rawEstimate := (alpha * m * m / sum)
	t := (rawEstimate - C) / C
	return uint64(rawEstimate - C*(math.Exp(-t)+0.125*t*(t-0.82)*math.Exp(-1.85*t)))
}

// ExpectedError returns the estimated error in the number of distinct items in the multiset
func (hll *HyperLogLog) ExpectedError() float64 {

	m := float64(uint64(1 << hll.p))
	return 1.04 / math.Sqrt(m)

}

// Reset zeros out the estimated number of distinct items in the multiset
func (hll *HyperLogLog) Reset() {

	for i := range hll.data {
		hll.data[i] = 0
	}
}

// ReducePrecision produces a new HyperLogLog with reduced precision
// if p>hll.p it returns nil and an error, if p==hll.p it just produces a copy
func (hll *HyperLogLog) ReducePrecision(p byte) (*HyperLogLog, error) {

	if p > hll.p {
		return nil, fmt.Errorf("Precision %d is greater than the current HyperLogLog precision %d", p, hll.p)
	} else if p < 4 {
		return nil, fmt.Errorf("Precision %d is less than the mimimum HyperLogLog precision %d", p, minimumHyperLogLogP)
	}
	newHLL := NewHyperLogLog(p, hll.hash)
	// populate new hll by taking max over the stride length
	newM := (1 << p)
	strideLength := (1 << (hll.p - p))
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

var inversePowersOfTwo = [...]float64{
	math.Pow(2.0, 0.0),
	math.Pow(2.0, -1.0),
	math.Pow(2.0, -2.0),
	math.Pow(2.0, -3.0),
	math.Pow(2.0, -4.0),
	math.Pow(2.0, -5.0),
	math.Pow(2.0, -6.0),
	math.Pow(2.0, -7.0),
	math.Pow(2.0, -8.0),
	math.Pow(2.0, -9.0),
	math.Pow(2.0, -10.0),
	math.Pow(2.0, -11.0),
	math.Pow(2.0, -12.0),
	math.Pow(2.0, -13.0),
	math.Pow(2.0, -14.0),
	math.Pow(2.0, -15.0),
	math.Pow(2.0, -16.0),
	math.Pow(2.0, -17.0),
	math.Pow(2.0, -18.0),
	math.Pow(2.0, -19.0),
	math.Pow(2.0, -20.0),
	math.Pow(2.0, -21.0),
	math.Pow(2.0, -22.0),
	math.Pow(2.0, -23.0),
	math.Pow(2.0, -24.0),
	math.Pow(2.0, -25.0),
	math.Pow(2.0, -26.0),
	math.Pow(2.0, -27.0),
	math.Pow(2.0, -28.0),
	math.Pow(2.0, -29.0),
	math.Pow(2.0, -30.0),
	math.Pow(2.0, -31.0),
	math.Pow(2.0, -32.0),
	math.Pow(2.0, -33.0),
	math.Pow(2.0, -34.0),
	math.Pow(2.0, -35.0),
	math.Pow(2.0, -36.0),
	math.Pow(2.0, -37.0),
	math.Pow(2.0, -38.0),
	math.Pow(2.0, -39.0),
	math.Pow(2.0, -40.0),
	math.Pow(2.0, -41.0),
	math.Pow(2.0, -42.0),
	math.Pow(2.0, -43.0),
	math.Pow(2.0, -44.0),
	math.Pow(2.0, -45.0),
	math.Pow(2.0, -46.0),
	math.Pow(2.0, -47.0),
	math.Pow(2.0, -48.0),
	math.Pow(2.0, -49.0),
	math.Pow(2.0, -50.0),
	math.Pow(2.0, -51.0),
	math.Pow(2.0, -52.0),
	math.Pow(2.0, -53.0),
	math.Pow(2.0, -54.0),
	math.Pow(2.0, -55.0),
	math.Pow(2.0, -56.0),
	math.Pow(2.0, -57.0),
	math.Pow(2.0, -58.0),
	math.Pow(2.0, -59.0),
	math.Pow(2.0, -60.0),
	math.Pow(2.0, -61.0),
	math.Pow(2.0, -62.0),
	math.Pow(2.0, -63.0),
}
