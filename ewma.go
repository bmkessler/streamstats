package streamstats

// EWMA data structure for exponentially weighted moving average
type EWMA struct {
	m      float64
	lambda float64
}

// NewEWMA initializes an EWMA with weighting lambda and given initial value
func NewEWMA(initialValue float64, lambda float64) EWMA {
	return EWMA{
		m:      initialValue,
		lambda: lambda,
	}
}

// Push updates the average value with the stored weight
func (e *EWMA) Push(x float64) {
	e.m = (1-e.lambda)*e.m + e.lambda*x
}

// Mean returns the exponentially weighted average value
func (e *EWMA) Mean() float64 {
	return e.m
}
