package streamstats

import (
	"math"
	"math/rand"
)

var result float64 // for benchmark results

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
