package streamstats

import (
	"math"
	"math/rand"
	"os"
	"testing"
)

// for benchmark results
const N = 10000

var result float64
var gaussianTestData = [N]float64{}
var exponentialTestData = [N]float64{}
var uniformTestData = [N]float64{}

func TestMain(m *testing.M) {
	for i := 0; i < N; i++ {
		gaussianTestData[i] = gaussianRandomVariable(0, 1)
		exponentialTestData[i] = exponentialRandomVariable(1)
		uniformTestData[i] = uniformRandomVariable(0, 1)
	}
	os.Exit(m.Run())
}

func gaussianRandomVariable(mean float64, stdev float64) float64 {
	return mean + stdev*rand.NormFloat64()
}

func exponentialRandomVariable(lambda float64) float64 {
	return rand.ExpFloat64() / lambda
}

func exponentialQuantile(p, lambda float64) float64 {
	return -1.0 * math.Log(1-p) / lambda
}

func uniformRandomVariable(min, max float64) float64 {
	return min + (max-min)*rand.Float64()
}

func uniformQuantile(p, min, max float64) float64 {
	return min + (max-min)*p
}

func cauchyQuantile(p, x0, gamma float64) float64 {
	return x0 + gamma*math.Tan(math.Pi*(p-0.5))
}

func cauchyRandomVariable(x0, gamma float64) float64 {
	return cauchyQuantile(rand.Float64(), x0, gamma)
}
