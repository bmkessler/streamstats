package streamstats

import (
	"fmt"
	"hash"
	"math"
)

const (
	minLinearCountingP = 6
	maxLinearCountingP = 24
)

// LinearCounting is a space efficient data structure for count distinct with hard upper bound
type LinearCounting struct {
	hash hash.Hash64 // a 64-bit hash function to map inputs to uniform buckets
	bits BitVector   // bitvector to hold the occupied buckets
	p    byte        // the number of buckets m = 2^p
}

// NewLinearCounting initializes a LinearCounting structure with size m=2^p and the given hash function
func NewLinearCounting(p byte, hash hash.Hash64) *LinearCounting {
	// p is bounded by 8 and 24 for practical implementations
	if p < minLinearCountingP {
		p = minLinearCountingP
	} else if p > maxLinearCountingP {
		p = maxLinearCountingP
	}
	m := uint64(1 << p)
	bits := NewBitVector(m)
	return &LinearCounting{p: p, hash: hash, bits: bits}
}

// Add adds an item to the multiset represented by the LinearCounting structure
func (lc *LinearCounting) Add(item []byte) {
	lc.hash.Reset()
	lc.hash.Write(item)
	hash := lc.hash.Sum64()
	bucket := hash >> (64 - lc.p) // top p bits are the bucket
	lc.bits.Set(bucket)
}

// Distinct returns the estimate of the number of distinct elements seen
// if the backing BitVector is full it returns m, the size of the BitVector
func (lc LinearCounting) Distinct() uint64 {
	m := uint64(1 << lc.p)
	zeroCount := m - lc.bits.PopCount()
	if zeroCount > 0 {
		return uint64(float64(m) * math.Log(float64(m)/float64(zeroCount)))
	}
	return (1 << lc.p)
}

// ReducePrecision produces a new HyperLogLog with reduced precision
// if p>hll.p it returns nil and an error, if p==hll.p it just produces a copy
func (lc *LinearCounting) ReducePrecision(p byte) (*LinearCounting, error) {

	if p > lc.p {
		return nil, fmt.Errorf("Precision %d is greater than the current LinearCounting precision %d", p, lc.p)
	} else if p < minLinearCountingP {
		return nil, fmt.Errorf("Precision %d is less than the minimum LinearCounting precision %d", p, minLinearCountingP)
	}
	newLC := NewLinearCounting(p, lc.hash)

	// copy the old BitVector to a new temporary one that can be folded
	bitsToFold := NewBitVector(uint64(1 << lc.p))
	for i := range lc.bits {
		bitsToFold[i] = lc.bits[i]
	}
	// "fold" the bit vector
	for i := lc.p; i > p; i-- {
		mFold := 1 << (i - 7) // half the current length in units of 64 bits
		for j := 0; j < mFold; j++ {
			bitsToFold[j] |= bitsToFold[j+mFold]
		}
	}
	// populate the folded vector into the new LinearCounting
	for i := range newLC.bits {
		newLC.bits[i] = bitsToFold[i]
	}

	return newLC, nil
}

// Combine the estimate of two LinearCounting reducing the precision to the minimum of the two sets
// the function will return nil and an error if the hash functions mismatch
func (lc *LinearCounting) Combine(lcB *LinearCounting) (*LinearCounting, error) {

	// check that both hash functions get the same result for "LinearCounting"
	lc.hash.Reset()
	lc.hash.Write([]byte("LinearCounting"))
	hash := lc.hash.Sum64()
	lcB.hash.Reset()
	lcB.hash.Write([]byte("LinearCounting"))
	hashB := lcB.hash.Sum64()
	if hash != hashB {
		return nil, fmt.Errorf("Hash functions are not identical, return %d != %d for \"LinearCounting\"", hash, hashB)
	}
	// determine if either precision needs to be reduced
	var combinedP byte
	var lc1, lc2, combinedLC *LinearCounting
	if lc.p < lcB.p {
		combinedP = lc.p
		lc1 = lc
		lc2, _ = lcB.ReducePrecision(lc.p)
	} else if lcB.p < lc.p {
		combinedP = lcB.p
		lc1, _ = lc.ReducePrecision(lcB.p)
		lc2 = lcB
	} else {
		combinedP = lc.p
		lc1 = lc
		lc2 = lcB
	}
	// for each bucket take the OR of the two LinearCounting
	combinedLC = NewLinearCounting(combinedP, lc.hash)
	for i := range combinedLC.bits {
		combinedLC.bits[i] = lc1.bits[i] | lc2.bits[i]
	}
	return combinedLC, nil
}
