package streamstats

import "strconv"

// BitVector represents an arbitrary length vector of bits backed by 64-bit words
// it is used as the data structure backing the Bloom Filter and Linear Counting implementations
type BitVector []uint64

// NewBitVector returns a new BitVector of length L
func NewBitVector(L uint64) BitVector {
	// length of backing slice is # of 64-bit words, lower 6 bits index index inside the word
	return BitVector(make([]uint64, 1+(L>>6), 1+(L>>6)))
}

// Set sets the bit at position N
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) Set(N uint64) {
	b[N>>6] |= 1 << (N & 63)
}

// Get returns the bit at position N as a uint64
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) Get(N uint64) uint64 {
	return (b[N>>6] >> (N & 63)) & 1
}

// Clear clears the bit at position N
// for N >= L this will access memory out of bounds of the backing array and panic
func (b BitVector) Clear(N uint64) {
	b[N>>6] = b[N>>6] &^ (1 << (N & 63))
}

// String outputs a string representation of the binary string with the first bit at the left
// note that any padding zeros are present on the right hand side
func (b BitVector) String() string {
	buff := make([]byte, 0, 64*len(b))
	for _, word := range b {
		bits := []byte(strconv.FormatUint(word, 2))
		for i := len(bits) - 1; i >= 0; i-- {
			buff = append(buff, bits[i]) // append the bits in reverse order
		}
		for j := len(bits); j < 64; j++ {
			buff = append(buff, '0') // add any leading zeros
		}
	}
	return string(buff)
}

// PopCount returns the nubmer of set bits in the bit vector
// the algorithm for PopCount on a single 64-bit word is from
// 1957 due to Donald B. Gillies and Jeffrey C. P. Miller
// and referenced by Donald Knuth
func (b BitVector) PopCount() uint64 {
	var total uint64
	for _, word := range b {
		word = word - ((word) >> 1 & 0x5555555555555555)
		word = (word & 0x3333333333333333) + ((word >> 2) & 0x3333333333333333)
		word = (word + (word >> 4)) & 0x0F0F0F0F0F0F0F0F
		word += (word >> 8)
		word += (word >> 16)
		word += (word >> 32)
		total += word & 255
	}
	return total
}
