package streamstats

import (
	"hash/fnv"
	"math/rand"
	"testing"
)

func TestJenkins2_32_OAAT(t *testing.T) {

	j := NewJenkins2_32(golden32)
	jFull := NewJenkins2_32(golden32)

	rand.Seed(42)
	numberOfBytes := 15
	b := make([]byte, numberOfBytes)
	rand.Read(b)

	for i, x := range b {
		j.Write([]byte{x}) // write one byte into the hash at a time
		s32 := j.Sum32()
		jFull.Reset()
		jFull.Write(b[0 : i+1]) // write all the bytes up to the current byte into another hash
		sf32 := jFull.Sum32()
		if s32 != sf32 { // the two hashes should agree
			t.Errorf("Byte %v Expected OOAT hash %x to be same as full hash %x\n", i, s32, sf32)
		}
	}
}

func BenchmarkJenkins2_32_8bytes(b *testing.B) {
	j := NewJenkins2_32(golden32)
	for i := 0; i < b.N; i++ {
		j.Write(randomBytes[i%N])
		j.Sum32()
		j.Reset()
	}
	count = uint64(j.Sum32()) // to avoid optimizing out the loop entirely
}

func BenchmarkJenkins2_32_16bytes(b *testing.B) {
	j := NewJenkins2_32(golden32)
	for i := 0; i < b.N; i++ {
		j.Write(randomBytes[i%N])
		j.Write(randomBytes[(i+1)%N])
		j.Sum32()
		j.Reset()
	}
	count = uint64(j.Sum32()) // to avoid optimizing out the loop entirely
}

func BenchmarkJenkins2_32_24bytes(b *testing.B) {
	j := NewJenkins2_32(golden32)
	for i := 0; i < b.N; i++ {
		j.Write(randomBytes[i%N])
		j.Write(randomBytes[(i+1)%N])
		j.Write(randomBytes[(i+2)%N])
		j.Sum32()
		j.Reset()
	}
	count = uint64(j.Sum32()) // to avoid optimizing out the loop entirely
}

func BenchmarkFNV_32_8bytes(b *testing.B) {
	j := fnv.New32()
	for i := 0; i < b.N; i++ {
		j.Write(randomBytes[i%N])
		j.Sum32()
		j.Reset()
	}
	count = uint64(j.Sum32()) // to avoid optimizing out the loop entirely
}

func BenchmarkFNV_32_16bytes(b *testing.B) {
	j := fnv.New32()
	for i := 0; i < b.N; i++ {
		j.Write(randomBytes[i%N])
		j.Write(randomBytes[(i+1)%N])
		j.Sum32()
		j.Reset()
	}
	count = uint64(j.Sum32()) // to avoid optimizing out the loop entirely
}

func BenchmarkFNV_32_24bytes(b *testing.B) {
	j := fnv.New32()
	for i := 0; i < b.N; i++ {
		j.Write(randomBytes[i%N])
		j.Write(randomBytes[(i+1)%N])
		j.Write(randomBytes[(i+2)%N])
		j.Sum32()
		j.Reset()
	}
	count = uint64(j.Sum32()) // to avoid optimizing out the loop entirely
}
