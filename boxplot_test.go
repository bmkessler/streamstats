package streamstats

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestBoxPlot(t *testing.T) {
	rand.Seed(42) // for deterministic testing
	N := 10000

	bp := NewBoxPlot()
	p := 0.5 // BoxPlots track the median
	q := NewP2Quantile(p)
	// add the same data to both and compare
	for i := 0; i < N; i++ {
		q.Push(exponentialTestData[i])
		bp.Push(exponentialTestData[i])
	}
	if bp.Median() != q.Quantile() {
		t.Errorf("Expected Median %v, got %v", q.Quantile(), bp.Median())
	}
	if bp.UpperQuartile() != q.UpperQuantile() {
		t.Errorf("Expected UpperQuartile %v, got %v", q.UpperQuantile(), bp.UpperQuartile())
	}
	if bp.LowerQuartile() != q.LowerQuantile() {
		t.Errorf("Expected LowerQuartile %v, got %v", q.LowerQuantile(), bp.LowerQuartile())
	}
	IQR := q.UpperQuantile() - q.LowerQuantile()
	if bp.InterQuartileRange() != IQR {
		t.Errorf("Expected InterQuartileRange %v, got %v", IQR, bp.InterQuartileRange())
	}
	upperWhisker := q.UpperQuantile() + 1.5*IQR
	if bp.UpperWhisker() != upperWhisker {
		t.Errorf("Expected upperWhisker %v, got %v", upperWhisker, bp.UpperWhisker())
	}
	lowerWhisker := q.LowerQuantile() - 1.5*IQR
	if bp.LowerWhisker() != lowerWhisker {
		t.Errorf("Expected InterQuartileRange %v, got %v", lowerWhisker, bp.LowerWhisker())
	}
	if bp.IsOutlier(q.Max()) == false {
		t.Errorf("Expected Max %v > UpperWhisker %v to be an outlier", q.Max(), bp.UpperWhisker())
	}
	if bp.IsOutlier(-1000) == false { // exponential distribution is skewed so min falls within the whisker
		t.Errorf("Expected Min %v < -1000 to be an outlier", q.Min())
	}
	midHinge := (q.UpperQuantile() + q.LowerQuantile()) / 2.0
	if bp.MidHinge() != midHinge {
		t.Errorf("Expected MidHinge %v, got %v", midHinge, bp.MidHinge())
	}
	midRange := (q.Max() + q.Min()) / 2.0
	if bp.MidRange() != midRange {
		t.Errorf("Expected MidRange %v, got %v", midRange, bp.MidRange())
	}
	triMean := (q.UpperQuantile() + 2.0*q.Quantile() + q.LowerQuantile()) / 4.0
	if bp.TriMean() != triMean {
		t.Errorf("Expected TriMean %v, got %v", triMean, bp.TriMean())
	}

	bp = NewBoxPlot()
	bp.Push(0.0)
	bp.Push(1.0)
	bp.Push(2.0)
	bp.Push(3.0)
	bp.Push(4.0)
	expectedString := fmt.Sprintf("Min: %0.3f LowerQuartile: %0.3f Median: %0.3f UpperQuartile: %0.3f Max: %0.3f N: %d", 0.0, 1.0, 2.0, 3.0, 4.0, 5)
	if expectedString != bp.String() {
		t.Errorf("Expected %s got %s", expectedString, bp)
	}
}
