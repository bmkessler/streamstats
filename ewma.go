package streamstats

import "sync"

type EWMA struct {
	sync.RWMutex
	m      float64
	lambda float64
}

func NewEWMA(initialVal float64, lambda float64) EWMA {
	return EWMA{m: initialVal, lambda: lambda}
}

func (e *EWMA) Push(x float64) {
	e.Lock()
	defer e.Unlock()
	e.m = (1-e.lambda)*e.m + e.lambda*x
}

func (e *EWMA) Mean() float64 {
	e.RLock()
	defer e.RUnlock()
	return e.m
}
