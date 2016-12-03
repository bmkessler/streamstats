package streamstats

// BoxPlot represents a BoxPlot with interquartile range and whiskers backed by a P2-Quantile tracking the median, P=0.5
type BoxPlot struct {
	P2Quantile
}

// NewBoxPlot returns a new BoxPlot
func NewBoxPlot() BoxPlot {
	return BoxPlot{NewP2Quantile(0.5)}
}

// Median returns the estimated median
func (bp BoxPlot) Median() float64 {
	return bp.Quantile()
}

// UpperQuartile returns the estimated upper quartile
func (bp BoxPlot) UpperQuartile() float64 {
	return bp.UpperQuantile()
}

// LowerQuartile returns the estimated lower quartile
func (bp BoxPlot) LowerQuartile() float64 {
	return bp.LowerQuantile()
}

// InterQuartileRange returns the estimated interquartile range
func (bp BoxPlot) InterQuartileRange() float64 {
	return bp.UpperQuantile() - bp.LowerQuantile()
}

// UpperWhisker returns the estimated upper whisker, Q3 + 1.5 * IQR
func (bp BoxPlot) UpperWhisker() float64 {
	return bp.UpperQuantile() + 1.5*bp.InterQuartileRange()
}

// LowerWhisker returns the estimated lower whisker, Q1 - 1.5 * IQR
func (bp BoxPlot) LowerWhisker() float64 {
	return bp.LowerQuantile() - 1.5*bp.InterQuartileRange()
}

// IsOutlier returns true if the data is outside the whiskers
func (bp BoxPlot) IsOutlier(x float64) bool {
	return x < bp.LowerWhisker() || x > bp.UpperWhisker()
}

// MidHinge returns the MidHinge of the data, average of upper and lower quantiles
func (bp BoxPlot) MidHinge() float64 {
	return (bp.UpperQuartile() + bp.LowerQuartile()) / 2.0
}

// MidRange returns the MidRange of the data, average of max and min
func (bp BoxPlot) MidRange() float64 {
	return (bp.Max() + bp.Min()) / 2.0
}

// TriMean returns the TriMean of the data, average of Median and MidHinge
func (bp BoxPlot) TriMean() float64 {
	return (bp.UpperQuartile() + 2.0*bp.Median() + bp.LowerQuartile()) / 4.0
}
