package streamstats

import "sync"

type P2Quantile struct {
	sync.RWMutex
	p float64
	n [5]uint64
	q [5]float64
}

func NewP2Quantile(p float64) P2Quantile {
	return P2Quantile{p: p, n: [5]uint64{1, 2, 3, 4, 0}}
}

func (p *P2Quantile) Push(x float64) {
	p.Lock()
	defer p.Unlock()

	if p.n[5] < 5 {
		// switch to a sorted insert
		p.q[p.n[5]] = x
		p.n[5] += 1
	} else {
		// TODO: implement the p2 algorithm
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
		return 0.0 // TODO: handle case of too few data points smoothly
	}
}

func (p *P2Quantile) UpperQuantile() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[3]
	} else {
		return 0.0 // TODO: handle case of too few data points smoothly
	}
}

func (p *P2Quantile) LowerQuantile() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[1]
	} else {
		return 0.0 // TODO: handle case of too few data points smoothly
	}
}

func (p *P2Quantile) Max() float64 {
	p.RLock()
	defer p.RUnlock()
	if p.n[5] >= 5 {
		return p.q[4]
	} else {
		return p.q[p.n[5]]
	}
}

func (p *P2Quantile) Min() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.q[0]
}
