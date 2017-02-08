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

// Compress produces a new LinearCouting with reduced size by 2^factor with reduced precision
// if new p < minLinearCountingP, p=minLinearCountingP , if factor=0 it just produces a copy
func (lc *LinearCounting) Compress(factor byte) *LinearCounting {
	var p byte
	if lc.p > factor {
		p = lc.p - factor
	}
	if p < minLinearCountingP {
		p = minLinearCountingP
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

	return newLC
}

// Union the estimate of two LinearCounting reducing the precision to the minimum of the two sets
// the function will return nil and an error if the hash functions mismatch
func (lc *LinearCounting) Union(lcB *LinearCounting) (*LinearCounting, error) {

	// check that both hash functions get the same result for "LinearCounting"
	lc.hash.Reset()
	lc.hash.Write([]byte("LinearCounting"))
	hash := lc.hash.Sum64()
	lcB.hash.Reset()
	lcB.hash.Write([]byte("LinearCounting"))
	hashB := lcB.hash.Sum64()
	if hash != hashB {
		return nil, fmt.Errorf("Hash functions are not identical, return %0x != %0x for \"LinearCounting\"", hash, hashB)
	}
	// determine if either precision needs to be reduced
	var combinedP byte
	var lc1, lc2, combinedLC *LinearCounting
	if lc.p < lcB.p {
		combinedP = lc.p
		factor := lcB.p - combinedP
		lc1 = lc
		lc2 = lcB.Compress(factor)
	} else if lcB.p < lc.p {
		combinedP = lcB.p
		factor := lc.p - combinedP
		lc1 = lc.Compress(factor)
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

// Intersect the estimate of two LinearCounting reducing the precision to the minimum of the two sets
// the function will return nil and an error if the hash functions mismatch
func (lc *LinearCounting) Intersect(lcB *LinearCounting) (*LinearCounting, error) {

	// check that both hash functions get the same result for "LinearCounting"
	lc.hash.Reset()
	lc.hash.Write([]byte("LinearCounting"))
	hash := lc.hash.Sum64()
	lcB.hash.Reset()
	lcB.hash.Write([]byte("LinearCounting"))
	hashB := lcB.hash.Sum64()
	if hash != hashB {
		return nil, fmt.Errorf("Hash functions are not identical, return %0x != %0x for \"LinearCounting\"", hash, hashB)
	}
	// determine if either precision needs to be reduced
	var combinedP byte
	var lc1, lc2, combinedLC *LinearCounting
	if lc.p < lcB.p {
		combinedP = lc.p
		factor := lcB.p - combinedP
		lc1 = lc
		lc2 = lcB.Compress(factor)
	} else if lcB.p < lc.p {
		combinedP = lcB.p
		factor := lc.p - combinedP
		lc1 = lc.Compress(factor)
		lc2 = lcB
	} else {
		combinedP = lc.p
		lc1 = lc
		lc2 = lcB
	}
	// for each bucket take the AND of the two LinearCounting
	combinedLC = NewLinearCounting(combinedP, lc.hash)
	for i := range combinedLC.bits {
		combinedLC.bits[i] = lc1.bits[i] & lc2.bits[i]
	}
	return combinedLC, nil
}

// Occupancy returns the ratio of filled buckets in the LinearCounting bitvector
func (lc LinearCounting) Occupancy() float64 {
	return float64(lc.bits.PopCount()) / float64(uint64(1<<lc.p))
}

// ExpectedError returns the expected error at the current filling in the LinearCounting
func (lc LinearCounting) ExpectedError() float64 {
	m := float64(uint64(1 << lc.p))
	loadFactor := lc.Occupancy()
	return 2 * math.Sqrt((math.Exp(loadFactor)-loadFactor-1)/m) / loadFactor
}

func (lc LinearCounting) String() string {
	N := lc.Distinct()
	delta := uint64(float64(N) * lc.ExpectedError())
	return fmt.Sprintf("LinearCounting N: %d +/- %d", N, delta)
}
