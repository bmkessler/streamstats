package streamstats

import (
	"hash"
)

// HyperLogLog a data structure for computing count distinct on arbitrary sized data
type HyperLogLog struct {
	hash hash.Hash64
	p    byte
	data []byte
}

// NewHyperLogLog returns a new HyperLogLog data structure with 2^p buckets
func NewHyperLogLog(hash hash.Hash64, p byte) *HyperLogLog {
	m := 1 << p
	return &HyperLogLog{
		hash: hash,
		p:    p,
		data: make([]byte, m, m),
	}
}
