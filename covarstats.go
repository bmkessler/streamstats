package streamstats

import "sync"

// CovarStats is a data structure for computing stats on two related variables x,y from a stream
type CovarStats struct {
	sync.RWMutex
	xStats MomentStats
	yStats MomentStats
	sXY    float64
}

// Push adds a sample of the two variables to the CovarStats data structure
func (c *CovarStats) Push(x, y float64) {
	c.Lock()
	defer c.Unlock()

	c.sXY += (c.xStats.Mean() - x) * (c.yStats.Mean() - y) * float64(c.xStats.n) / float64(c.xStats.n+1)
	c.xStats.Push(x)
	c.yStats.Push(y)
}

// Slope returns the slope of the correlation between x and y samples seen so far
func (c *CovarStats) Slope() float64 {
	c.RLock()
	defer c.RUnlock()
	sXX := c.xStats.Variance() * float64(c.xStats.n-1.0)
	return c.sXY / sXX
}

// Intercept returns the intercept of the correlation between x and y samples seen so far
func (c *CovarStats) Intercept() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Mean() - c.Slope()*c.xStats.Mean()
}

// Correlation returns the Pearson product-moment correlation coefficient of the x and y samples seen so far
func (c *CovarStats) Correlation() float64 {
	c.RLock()
	defer c.RUnlock()
	t := c.xStats.StdDev() * c.yStats.StdDev()
	return c.sXY / (float64(c.xStats.n-1) * t)
}

// N returns the number of samples seen so far
func (c *CovarStats) N() uint64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.N()
}

// XMean returns the mean of the x values seen so far
func (c *CovarStats) XMean() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Mean()
}

// XVariance returns the variance of the x values seen so far
func (c *CovarStats) XVariance() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Variance()
}

// XStdDev returns the standard deviation of the x values seen so far
func (c *CovarStats) XStdDev() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.StdDev()
}

// XSkewness returns the skewness of the x values seen so far
func (c *CovarStats) XSkewness() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Skewness()
}

// XKurtosis returns the kurtorsis of the x values seen so far
func (c *CovarStats) XKurtosis() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Kurtosis()
}

// YMean returns the mean of the y values seen so far
func (c *CovarStats) YMean() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Mean()
}

// YVariance returns the variance of the y values seen so far
func (c *CovarStats) YVariance() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Variance()
}

// YStdDev returns the standard deviation of the y values seen so far
func (c *CovarStats) YStdDev() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.StdDev()
}

// YSkewness returns the skewness of the y values seen so far
func (c *CovarStats) YSkewness() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Skewness()
}

// YKurtosis returns the kurtosis of the y values seen so far
func (c *CovarStats) YKurtosis() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Kurtosis()
}

// Combine returns the combination of two CovarStats datastructures
func (c *CovarStats) Combine(b *CovarStats) CovarStats {
	var combined CovarStats

	c.RLock()
	b.RLock()
	defer c.RUnlock()
	defer b.RUnlock()

	combined.xStats = c.xStats.Combine(&b.xStats)
	combined.yStats = c.yStats.Combine(&b.yStats)

	deltaX := b.xStats.Mean() - c.xStats.Mean()
	deltaY := b.yStats.Mean() - c.yStats.Mean()
	combined.sXY = c.sXY + b.sXY + float64(c.xStats.n*b.xStats.n)*deltaX*deltaY/float64(combined.xStats.n)

	return combined
}
