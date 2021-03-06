package streamstats

import (
	"math"
	"math/rand"
	"os"
	"testing"
)

// for benchmark results
const (
	b    = 14 // 14-bits for test data
	N    = 1 << b
	mask = N - 1
)

var result float64
var count32 uint32
var count uint64
var gaussianTestData = [N]float64{}
var exponentialTestData = [N]float64{}
var uniformTestData = [N]float64{}
var randomBytes = [N][]byte{}
var longRandomBytes = [N][]byte{}

func TestMain(m *testing.M) {
	rand.Seed(42)
	for i := 0; i < N; i++ {
		gaussianTestData[i] = gaussianRandomVariable(0, 1)
		exponentialTestData[i] = exponentialRandomVariable(1)
		uniformTestData[i] = uniformRandomVariable(0, 1)
		b := make([]byte, 8)
		rand.Read(b)
		randomBytes[i] = b
		d := make([]byte, 29)
		rand.Read(d)
		longRandomBytes[i] = d
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
