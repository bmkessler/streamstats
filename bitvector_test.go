package streamstats

import "testing"

func TestNewBitVector(t *testing.T) {
	var L uint64
	for L = 1; L <= 1024; L++ {
		bits := NewBitVector(L)
		m := 1 + ((L - 1) >> 6)
		if uint64(len(bits)) != m {
			t.Errorf("Expected data to be length %d, got %d\n", m, len(bits))
		}
	}
}

func TestBitVectorSetGetClear(t *testing.T) {
	var L uint64 = 88
	bits := NewBitVector(L)
	var i uint64
	for i = 0; i < L; i++ {
		if bits.Get(i) != 0 {
			t.Errorf("Expected bit %d to be unset.", i)
		}
		bits.Set(i)
		if bits.Get(i) != 1 {
			t.Errorf("Expected bit %d to be set.", i)
		}
		bits.Clear(i)
		if bits.Get(i) != 0 {
			t.Errorf("Expected bit %d to be unset.", i)
		}
	}
}

func TestBitVectorPopCount(t *testing.T) {
	var L uint64 = 1234
	bits := NewBitVector(L)
	if bits.PopCount() != 0 {
		t.Errorf("Expected PopCount to be 0 got %d", bits.PopCount())
	}
	var i uint64
	for i = 0; i < L; i++ {
		bits.Set(i)
		if bits.PopCount() != i+1 {
			t.Errorf("Expected PopCount to be %d got %d", i+1, bits.PopCount())
		}
	}
}

func TestBitVectorString(t *testing.T) {
	var L uint64 = 88
	bits := NewBitVector(L)

	zeroBitsring := "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	oneBitstring := "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	eightyEightBitstring := "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000"
	patternBitstring := "10101000100000001000000000000000100000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000001"

	if bits.String() != zeroBitsring {
		t.Errorf("Expected bitstring:\n%s\nGot:\n%s", bits.String(), zeroBitsring)
	}
	bits.Set(0)
	if bits.String() != oneBitstring {
		t.Errorf("Expected bitstring:\n%s\nGot:\n%s", bits.String(), oneBitstring)
	}
	bits.Clear(0)
	if bits.String() != zeroBitsring {
		t.Errorf("Expected bitstring:\n%s\nGot:\n%s", bits.String(), zeroBitsring)
	}
	bits.Set(88)
	if bits.String() != eightyEightBitstring {
		t.Errorf("Expected bitstring:\n%s\nGot:\n%s", bits.String(), eightyEightBitstring)
	}
	bits.Clear(88)
	if bits.String() != zeroBitsring {
		t.Errorf("Expected bitstring:\n%s\nGot:\n%s", bits.String(), zeroBitsring)
	}
	bits.Set(0)
	bits.Set(2)
	bits.Set(4)
	bits.Set(8)
	bits.Set(16)
	bits.Set(32)
	bits.Set(64)
	bits.Set(127)
	if bits.String() != patternBitstring {
		t.Errorf("Expected bitstring:\n%s\nGot:\n%s", bits.String(), patternBitstring)
	}
}
