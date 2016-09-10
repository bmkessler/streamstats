package streamstats

import "sync"

type EWMA struct {
	sync.RWMutex
	M      float64
	lambda float64
}

func NewEWMA(initialVal float64, lambda float64) EWMA {
	return EWMA{M: initialVal, lambda: lambda}
}

func (e *EWMA) Push(x float64) {
	e.Lock()
	defer e.Unlock()
	e.M = (1-e.lambda)*e.M + e.lambda*x
}

func (e *EWMA) Mean() float64 {
	e.RLock()
	defer e.RUnlock()
	return e.M
}
