package streamstats

// Jenkins2 hash from http://burtleburtle.net/bob/hash/evahash.html
// This is not optimized for speed and is 6-7x slower than the built-in FNV hash

// Jenkins2_32 32-bit version, hashes 12-bytes at a time
type Jenkins2_32 struct {
	key      uint32 // the key that the hash was initialized with
	a        uint32 // a, b, c are the state of the hash
	b        uint32
	c        uint32
	numBytes uint32 // the total number of bytes hashed so far, used in finalization
	buffer   []byte // a buffer to hold un-hashed values will always be less than 12-bytes
}

// the golden ratio 32-bits, an arbitrary initialization constant
const golden32 = 0x9e3779b9

// NewJenkins2_32 returns a new Jenkins2 32-bit hash structure with the given key
func NewJenkins2_32(key uint32) *Jenkins2_32 {
	return &Jenkins2_32{key: key, a: golden32, b: golden32, c: key}
}

// Reset zeroes the hash back struct back to its initial state to allow new bytes hashed
func (h *Jenkins2_32) Reset() {
	h.a = golden32
	h.b = golden32
	h.c = h.key
	h.numBytes = 0
	h.buffer = []byte{}
}

// Size returns 4 for the 4-byte (32-bit) output
func (h Jenkins2_32) Size() int {
	return 4 // the standard use returns a 32-bit hash
}

// BlockSize returns 12 sicne the hash operates on 12-byte blocks
func (h Jenkins2_32) BlockSize() int {
	return 12 // the hash operates on 12-byte blocks
}

// Sum returns 4-bytes of hash for bs (32-bit) without affecting the state
func (h Jenkins2_32) Sum(bs []byte) []byte {
	// copy the old struct to a new struct
	// since the requirement is the function doesn't affect the underlying state
	nj := Jenkins2_32{
		key:      h.key,
		a:        h.a,
		b:        h.b,
		c:        h.c,
		numBytes: h.numBytes,
		buffer:   h.buffer,
	}
	nj.Write(bs)
	nj.finalize()
	var out []byte
	for i := 0; i < 32; i += 8 {
		out = append(out, byte(nj.c))
	}
	return out
}

// Sum32 returns the hash as a 32-bit uint without affecting the state
func (h Jenkins2_32) Sum32() uint32 {
	// The expected semantics are unclear here, so not affecting the underlying state
	nj := Jenkins2_32{
		key:      h.key,
		a:        h.a,
		b:        h.b,
		c:        h.c,
		numBytes: h.numBytes,
		buffer:   h.buffer,
	}
	nj.finalize()
	return nj.c
}

// mix mixes the internal state of the hash using fast bit-wise operations
func (h *Jenkins2_32) mix() {
	h.a = h.a - h.b
	h.a = h.a - h.c
	h.a = h.a ^ (h.c >> 13)
	h.b = h.b - h.c
	h.b = h.b - h.a
	h.b = h.b ^ (h.a << 8)
	h.c = h.c - h.a
	h.c = h.c - h.b
	h.c = h.c ^ (h.b >> 13)
	h.a = h.a - h.b
	h.a = h.a - h.c
	h.a = h.a ^ (h.c >> 12)
	h.b = h.b - h.c
	h.b = h.b - h.a
	h.b = h.b ^ (h.a << 16)
	h.c = h.c - h.a
	h.c = h.c - h.b
	h.c = h.c ^ (h.b >> 5)
	h.a = h.a - h.b
	h.a = h.a - h.c
	h.a = h.a ^ (h.c >> 3)
	h.b = h.b - h.c
	h.b = h.b - h.a
	h.b = h.b ^ (h.a << 10)
	h.c = h.c - h.a
	h.c = h.c - h.b
	h.c = h.c ^ (h.b >> 15)
}

// hash12Bytes adds 12-bytes to the hash and mixes them
func (h *Jenkins2_32) hash12Bytes(k []byte) {
	h.a = h.a + (uint32(k[0]) + (uint32(k[1]) << 8) + (uint32(k[2]) << 16) + (uint32(k[3]) << 24))
	h.b = h.b + (uint32(k[4]) + (uint32(k[5]) << 8) + (uint32(k[6]) << 16) + (uint32(k[7]) << 24))
	h.c = h.c + (uint32(k[8]) + (uint32(k[9]) << 8) + (uint32(k[10]) << 16) + (uint32(k[11]) << 24))
	h.mix()
	h.numBytes += 12
}

// finalize handles the remaining bits that aren't a multiple of 12 and includes the overall length in the hash
func (h *Jenkins2_32) finalize() {
	h.c = h.c + h.numBytes + uint32(len(h.buffer))
	switch len(h.buffer) { // all the case statements fall through

	case 11:
		h.c = h.c + (uint32(h.buffer[10]) << 24)
		fallthrough
	case 10:
		h.c = h.c + (uint32(h.buffer[9]) << 16)
		fallthrough
	case 9:
		h.c = h.c + (uint32(h.buffer[8]) << 8)
		fallthrough
		// the first byte of c is reserved for the length
	case 8:
		h.b = h.b + (uint32(h.buffer[7]) << 24)
		fallthrough
	case 7:
		h.b = h.b + (uint32(h.buffer[6]) << 16)
		fallthrough
	case 6:
		h.b = h.b + (uint32(h.buffer[5]) << 8)
		fallthrough
	case 5:
		h.b = h.b + uint32(h.buffer[4])
		fallthrough
	case 4:
		h.a = h.a + (uint32(h.buffer[3]) << 24)
		fallthrough
	case 3:
		h.a = h.a + (uint32(h.buffer[2]) << 16)
		fallthrough
	case 2:
		h.a = h.a + (uint32(h.buffer[1]) << 8)
		fallthrough
	case 1:
		h.a = h.a + uint32(h.buffer[0])
		fallthrough
	default:
		// case 0: nothing left to add
	}
	h.mix()
}

// Write adds 12-byte chunks to the hash and stores the remainder in a bufer
func (h *Jenkins2_32) Write(p []byte) (n int, err error) {
	bytesToWrite := append(h.buffer, p...)
	chunks := len(bytesToWrite) / 12
	i := 0
	for i < chunks {
		h.hash12Bytes(bytesToWrite[i : i+12])
		i += 12
	}
	h.buffer = bytesToWrite[i:len(bytesToWrite)]
	return len(p), nil
}

/* TODO: port to 64-bits from Bob Jenkins
The modifications needed to hash() are straightforward. It should put 24-byte blocks into 3 8-byte registers and return an 8-byte result. The 64-bit golden ratio is 0x9e3779b97f4a7c13LL.

#define mix64(a,b,c) \
{ \
  a=a-b;  a=a-c;  a=a^(c>>43); \
  b=b-c;  b=b-a;  b=b^(a<<9); \
  c=c-a;  c=c-b;  c=c^(b>>8); \
  a=a-b;  a=a-c;  a=a^(c>>38); \
  b=b-c;  b=b-a;  b=b^(a<<23); \
  c=c-a;  c=c-b;  c=c^(b>>5); \
  a=a-b;  a=a-c;  a=a^(c>>35); \
  b=b-c;  b=b-a;  b=b^(a<<49); \
  c=c-a;  c=c-b;  c=c^(b>>11); \
  a=a-b;  a=a-c;  a=a^(c>>12); \
  b=b-c;  b=b-a;  b=b^(a<<18); \
  c=c-a;  c=c-b;  c=c^(b>>22); \
}
*/
