package streamstats

import "sync"

type P2Quantile struct {
	sync.RWMutex
	p   float64    // the p-quantile to be tracked
	n   [5]uint64  // the actual counts for each marker
	np  [5]float64 // the target counts for each marker
	dnp [5]float64 // the updates to the target counts for each additional measurement
	q   [5]float64 // the value of each marker, i.e. the estimated quantile
}

func NewP2Quantile(p float64) P2Quantile {
	return P2Quantile{
		p:   p,
		n:   [5]uint64{1, 2, 3, 4, 0},
		np:  [5]float64{1, 1 + 2*p, 2 + 4*p, 3 + 2*p, 5},
		dnp: [5]float64{0, p, p / 2, (1 + p) / 2, 1},
	}
}

func (p *P2Quantile) Push(x float64) {
	p.Lock()
	defer p.Unlock()

	if p.n[5] < 5 {
		// Initialization:
		i := p.n[5]
		p.q[i] = x // add the new element
		// insertion sort the first 5 elements
		for i > 0 && p.q[i-1] > p.q[i] {
			t := p.q[i-1]
			p.q[i-1] = p.q[i]
			p.q[i] = t
			i--
		}
		p.n[5] += 1
	} else {
		// TODO: implement the p2 algorithm
		var k uint64
		switch {
		case x < p.q[0]:
			p.q[0] = x
			k = 0
		case x < p.q[1]:
			k = 1
		case x < p.q[2]:
			k = 2
		case x < p.q[3]:
			k = 3
		case x < p.q[4]:
			k = 4
		default:
			p.q[4] = x
			k = 4
		}

		for i := k + 1; i < 4; i++ {
			p.n[i]++
		}

		for i := 0; i < 4; i++ {
			p.np[i] += p.dnp[i]
		}
	}
}

func (p *P2Quantile) P() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.p
}

func (p *P2Quantile) N() uint64 {
	p.RLock()
	defer p.RUnlock()
	return p.n[5]
}

func (p *P2Quantile) Quantile() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[2]
	} else {
		if p.n[5]%2 == 1 {
			return p.q[p.n[5]/2] // the median value
		} else {
			return (p.q[p.n[5]/2-1] + p.q[p.n[5]/2]) / 2 //average of values around median
		}
	}
}

func (p *P2Quantile) UpperQuantile() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[3]
	} else {
		return (p.Quantile() + p.Max()) / 2 // Average of the quantile and the max value
	}
}

func (p *P2Quantile) LowerQuantile() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[1]
	} else {
		return (p.Min() + p.Quantile()) / 2 // Average of the min value and the quantile
	}
}

func (p *P2Quantile) Max() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[4]
	} else {
		return p.q[p.n[5]-1] // the highest seen so far
	}

}

func (p *P2Quantile) Min() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.q[0]
}
