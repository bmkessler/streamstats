package streamstats

// P2Quantile is a thread-safe, O(1) time and space data structure
// for estimating the p-quantile of a series of N data points based on the
// "The P2 Algorithm for Dynamic Computing Calculation of Quantiles and
// Histograms Without Storing Observations" by RAJ JAIN and IIMRICH CHLAMTAC
// Communications of the ACM Volume 28 Issue 10, Oct. 1985 Pages 1076-1085
type P2Quantile struct {
	p   float64    // the p-quantile to be tracked
	n   [5]uint64  // the actual counts for each marker
	np  [5]float64 // the target counts for each marker
	dnp [5]float64 // the updates to the target counts for each additional measurement
	q   [5]float64 // the value of each marker, i.e. the estimated quantile
}

// NewP2Quantile intializes the data structure to track the p-quantile
func NewP2Quantile(p float64) P2Quantile {
	return P2Quantile{
		p:   p,
		n:   [5]uint64{1, 2, 3, 4, 0},
		np:  [5]float64{1, 1 + 2*p, 1 + 4*p, 3 + 2*p, 5},
		dnp: [5]float64{0, p / 2, p, (1 + p) / 2, 1},
	}
}

// Push updates the data structure with a given x value
func (p *P2Quantile) Push(x float64) {

	if p.n[4] < 5 {
		// Initialization:
		i := p.n[4] // the current count
		p.q[i] = x  // add the new element on the end
		// insertion sort the elements
		for i > 0 && p.q[i-1] > p.q[i] {
			t := p.q[i-1]
			p.q[i-1] = p.q[i]
			p.q[i] = t
			i--
		}
		p.n[4]++
	} else {
		// find which bin the new element lies in
		var k uint64
		switch {
		case x < p.q[0]:
			p.q[0] = x // new minimum
			k = 0
		case x < p.q[1]:
			k = 0
		case x < p.q[2]:
			k = 1
		case x < p.q[3]:
			k = 2
		case x < p.q[4]:
			k = 3
		default:
			p.q[4] = x // new maximum
			k = 3
		}
		// update the actual counts for the markers
		for i := k + 1; i < 5; i++ {
			p.n[i]++
		}
		// update the goal counts for the markers
		for i := 0; i < 5; i++ {
			p.np[i] += p.dnp[i]
		}
		// adjust heights of internal markers if neccesary
		for i := 1; i < 4; i++ {
			d := p.np[i] - float64(p.n[i]) // the difference from the target
			if (d >= 1.0 && p.n[i]+1 < p.n[i+1]) || (d <= -1.0 && p.n[i-1]+1 < p.n[i]) {
				// delta is always snapped to +/- 1
				if d >= 1.0 {
					d = 1.0
				} else {
					d = -1.0
				}
				// try using the piecewise polynomial degree 2 formula
				fNm := float64(p.n[i-1])
				fN := float64(p.n[i])
				fNp := float64(p.n[i+1])
				qp := p.q[i] + d*((fN-fNm+d)*(p.q[i+1]-p.q[i])/(fNp-fN)+(fNp-fN-d)*(p.q[i]-p.q[i-1])/(fN-fNm))/(fNp-fNm)
				if p.q[i-1] < qp && qp < p.q[i+1] {
					p.q[i] = qp
				} else { // use linear formula if degree 2 formula would result in out of order markers
					ip := i + int(d)
					p.q[i] += d * (p.q[ip] - p.q[i]) / (float64(p.n[ip]) - fN)
				}
				if d > 0 { // increment the counter for the bin after adjustments were made
					p.n[i]++
				} else {
					p.n[i]--
				}
			}
		}
	}
}

// P returns the quantile being tracked
func (p *P2Quantile) P() float64 {
	return p.p
}

// N returns the number of observations seen so far
func (p *P2Quantile) N() uint64 {
	return p.n[4]
}

// Quantile returns the estimated value for the p-quantile
func (p *P2Quantile) Quantile() float64 {
	if p.n[4] < 5 && 0 < p.n[4] {
		if p.n[4]%2 == 0 {
			return (p.q[p.n[4]/2-1] + p.q[p.n[4]/2]) / 2 // average of values around median for even N
		}
		return p.q[p.n[4]/2] // the median value seen in the first 5 for odd N
	}
	return p.q[2] // the estimate of the median
}

// UpperQuantile returns the estimate for the upper quantile, (1+p/2)
func (p *P2Quantile) UpperQuantile() float64 {
	if p.n[4] < 5 && 0 < p.n[4] {
		return (p.Quantile() + p.Max()) / 2 // average the data if we don't have enough points yet
	}
	return p.q[3]
}

// LowerQuantile returns the estimate for the lower quantile, p/2
func (p *P2Quantile) LowerQuantile() float64 {
	if p.n[4] < 5 && 0 < p.n[4] {
		return (p.Min() + p.Quantile()) / 2 // average the data if we don't have enough points yet
	}
	return p.q[1]
}

// Max returns the exact maximum value seen so far
func (p *P2Quantile) Max() float64 {
	if p.n[4] < 5 && 0 < p.n[4] {
		return p.q[p.n[4]-1] // the highest for small counts
	}
	return p.q[4]
}

// Min returns the exact minimum value seen so far
func (p *P2Quantile) Min() float64 {
	return p.q[0]
}
