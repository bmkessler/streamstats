package streamstats

// P2Histogram is a thread-safe, O(1) time and space data structure
// for estimating the evenly spaced histogram bins of a series of N data points based on the
// "The P2 Algorithm for Dynamic Computing Calculation of Quantiles and
// Histograms Without Storing Observations" by RAJ JAIN and IIMRICH CHLAMTAC
// Communications of the ACM Volume 28 Issue 10, Oct. 1985 Pages 1076-1085
type P2Histogram struct {
	b int       // the number of bins to be tracked
	n []uint64  // the actual counts for each marker
	q []float64 // the value of each marker, i.e. the estimated quantile
}

// NewP2Histogram intializes the data structure to track b bins
func NewP2Histogram(b int) P2Histogram {
	n := make([]uint64, b+1, b+1)
	q := make([]float64, b+1, b+1)
	for i := 0; i < b; i++ {
		n[i] = uint64(i) + 1
	}
	return P2Histogram{
		b: b,
		n: n,
		q: q,
	}
}

// Push updates the data structure with a given x value
func (h *P2Histogram) Push(x float64) {

	if h.n[h.b] < uint64(h.b)+1 {
		// Initialization:
		i := h.n[h.b] // the current count
		h.q[i] = x    // add the new element on the end
		// insertion sort the elements
		for i > 0 && h.q[i-1] > h.q[i] {
			t := h.q[i-1]
			h.q[i-1] = h.q[i]
			h.q[i] = t
			i--
		}
		h.n[h.b]++
	} else {
		// find which bin the new element lies in
		var k uint64
		if x < h.q[0] {
			h.q[0] = x // new minimum
			k = 0
		} else if h.q[h.b] < x {
			h.q[h.b] = x // new maximum
			k = uint64(h.b) - 2
		} else { // check which bin the measurement falls into
			for i := 1; i <= h.b; i++ {
				if x < h.q[i] {
					k = uint64(i - 1)
					break
				}
			}
		}
		// update the actual counts for the markers
		for i := k + 1; i < uint64(h.b)+1; i++ {
			h.n[i]++
		}
		// adjust heights of internal markers if necessary
		for i := 1; i < h.b; i++ {
			np := 1.0 + float64(i)*(float64(h.n[i])-1.0)/float64(h.b)
			d := np - float64(h.n[i]) // the difference from the target
			if (d >= 1.0 && h.n[i]+1 < h.n[i+1]) || (d <= -1.0 && h.n[i-1]+1 < h.n[i]) {
				// delta is always snapped to +/- 1
				if d >= 1.0 {
					d = 1.0
				} else {
					d = -1.0
				}
				// try using the piecewise polynomial degree 2 formula
				fNm := float64(h.n[i-1])
				fN := float64(h.n[i])
				fNp := float64(h.n[i+1])
				qp := h.q[i] + d*((fN-fNm+d)*(h.q[i+1]-h.q[i])/(fNp-fN)+(fNp-fN-d)*(h.q[i]-h.q[i-1])/(fN-fNm))/(fNp-fNm)
				if h.q[i-1] < qp && qp < h.q[i+1] {
					h.q[i] = qp
				} else { // use linear formula if degree 2 formula would result in out of order markers
					ip := i + int(d)
					h.q[i] += d * (h.q[ip] - h.q[i]) / (float64(h.n[ip]) - fN)
				}
				if d > 0 { // increment the counter for the bin after adjustments were made
					h.n[i]++
				} else {
					h.n[i]--
				}
			}
		}
	}
}

// N returns the number of observations seen so far
func (h *P2Histogram) N() uint64 {

	return h.n[h.b+1]
}

// CumulativeDensity represents the probability P of observing a value less than or equal to X
type CumulativeDensity struct {
	X float64
	P float64
}

// Histogram returns the histogram of observations seen so far
func (h *P2Histogram) Histogram() []CumulativeDensity {

	cdf := make([]CumulativeDensity, h.b+1, h.b+1)
	fN := float64(h.N())
	for i := 0; i < h.b+1; i++ {
		cdf[i] = CumulativeDensity{X: h.q[i], P: float64(h.n[i]) / fN}
	}
	return cdf
}

// Min returns the minimum of observations seen so far
func (h *P2Histogram) Min() float64 {

	return h.q[0]
}

// Max returns the maximum of observations seen so far
func (h *P2Histogram) Max() float64 {

	if h.n[h.b] < uint64(h.b)+1 && 0 < h.n[h.b] {
		return h.q[h.n[h.b]-1]
	}
	return h.q[h.b]
}

// Quantile returns the linear approximation to the given quantile based on the histogram data
func (h *P2Histogram) Quantile(p float64) float64 {

	var q float64
	if p <= 0.0 {
		q = h.Min()
	} else if 1.0 <= p {
		q = h.Max()
	}
	N := uint64(h.b)
	if h.n[h.b] < uint64(h.b)+1 && 0 < h.n[h.b] {
		N = h.n[h.b] // there aren't enough values yet
	}
	fN := float64(h.N())
	for i := uint64(0); i < N && float64(h.n[i])/fN < p; i++ {
		q = h.q[i] + p*fN*(h.q[i+1]-h.q[i])/(float64(h.n[i+1])-float64(h.n[i]))
	}
	return q
}

// CDF returns the linear approximation to the CDF at x based on the histogram data
func (h *P2Histogram) CDF(x float64) float64 {

	var q float64
	if x < h.Min() {
		q = 0.0
	} else if h.Max() <= x {
		q = 1.0
	}
	N := uint64(h.b)
	if h.n[h.b] < uint64(h.b)+1 && 0 < h.n[h.b] {
		N = h.n[h.b] // there aren't enough values yet
	}
	fN := float64(h.N())
	for i := uint64(0); i < N && h.q[i] < x; i++ {
		q = float64(h.n[i])/fN + x*(float64(h.n[i+1])-float64(h.n[i]))/(fN*(h.q[i+1]-h.q[i]))
	}
	return q
}
