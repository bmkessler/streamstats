package streamstats

import "sync"

type CovarStats struct {
	sync.RWMutex
	xStats MomentStats
	yStats MomentStats
	sXY    float64
}

func (c *CovarStats) Push(x, y float64) {
	c.Lock()
	defer c.Unlock()

	c.sXY += (c.xStats.Mean() - x) * (c.yStats.Mean() - y) * float64(c.xStats.n) / float64(c.xStats.n+1)
	c.xStats.Push(x)
	c.yStats.Push(y)
}

func (c *CovarStats) Slope() float64 {
	c.RLock()
	defer c.RUnlock()
	sXX := c.xStats.Variance() * float64(c.xStats.n-1.0)
	return c.sXY / sXX
}

func (c *CovarStats) Intercept() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Mean() - c.Slope()*c.xStats.Mean()
}

func (c *CovarStats) Correlation() float64 {
	c.RLock()
	defer c.RUnlock()
	t := c.xStats.StdDev() * c.yStats.StdDev()
	return c.sXY / (float64(c.xStats.n-1) * t)
}

func (c *CovarStats) N() {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.N()
}

func (c *CovarStats) XVariance() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Variance()
}

func (c *CovarStats) XStdDev() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.StdDev()
}

func (c *CovarStats) XSkewness() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Skewness()
}

func (c *CovarStats) XKurtosis() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.xStats.Kurtosis()
}

func (c *CovarStats) YVariance() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Variance()
}

func (c *CovarStats) YStdDev() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.StdDev()
}

func (c *CovarStats) YSkewness() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Skewness()
}

func (c *CovarStats) YKurtosis() float64 {
	c.RLock()
	defer c.RUnlock()
	return c.yStats.Kurtosis()
}

func (a *CovarStats) Combine(b *CovarStats) CovarStats {
	var combined CovarStats

	a.RLock()
	b.RLock()
	defer a.RUnlock()
	defer b.RUnlock()

	combined.xStats = a.xStats.Combine(b.xStats)
	combined.yStats = a.yStats.Combine(b.yStats)

	deltaX := b.xStats.Mean() - a.xStats.Mean()
	deltaY := b.yStats.Mean() - a.yStats.Mean()
	combined.sXY = a.sXY + b.sXY + float64(a.xStats.n*b.xStats.n)*deltaX*deltaY/float64(combined.xStats.n)

	return combined
}
