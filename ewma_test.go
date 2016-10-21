package streamstats

import (
	"testing"
	"time"
)

func TestEWMA(t *testing.T) {
	initialVal := 4.0
	secondVal := 8.0
	lambda := 0.5
	e := NewEWMA(initialVal, lambda)
	if e.Mean() != initialVal {
		t.Errorf("expected initial value %f, got %f", initialVal, e.Mean())
	}
	e.Push(secondVal)
	expectedVal := 6.0 // if lambda = 0.5 expect to get the average
	if e.Mean() != expectedVal {
		t.Errorf("expected value %f, got %f", expectedVal, e.Mean())
	}

	initialVal = 10.0
	secondVal = 20.0
	lambda = 0.9
	e = NewEWMA(initialVal, lambda)
	if e.Mean() != initialVal {
		t.Errorf("expected initial value %f, got %f", initialVal, e.Mean())
	}
	e.Push(secondVal)
	expectedVal = 19.0 // if lambda = 0.9 and small number comes first
	if e.Mean() != expectedVal {
		t.Errorf("expected value %f, got %f", expectedVal, e.Mean())
	}

	initialVal = 20.0
	secondVal = 10.0
	lambda = 0.9
	e = NewEWMA(initialVal, lambda)
	if e.Mean() != initialVal {
		t.Errorf("expected initial value %f, got %f", initialVal, e.Mean())
	}
	e.Push(secondVal)
	expectedVal = 11.0 // if lambda = 0.9 and small number comes second
	if e.Mean() != expectedVal {
		t.Errorf("expected value %f, got %f", expectedVal, e.Mean())
	}
}

func BenchmarkEWMAPush(b *testing.B) {
	e := NewEWMA(0.0, 0.5)
	for i := 0; i < b.N; i++ {
		e.Push(gaussianTestData[i%N])
	}
	result = e.Mean() // to avoid optimizing out the loop entirely
}

func BenchmarkEWMAPushReadContention(b *testing.B) {
	e := NewEWMA(0.0, 0.5)
	contentionInterval := time.Nanosecond * 1 // interval to contend
	go func() {
		ticker := time.NewTicker(contentionInterval)
		for _ = range ticker.C {
			result = e.Mean() // a contentious read
		}
	}()

	for i := 0; i < b.N; i++ {
		e.Push(gaussianTestData[i%N])
	}
	result = e.Mean() // to avoid optimizing out the loop entirely
}

func BenchmarkEWMAPushWriteContention(b *testing.B) {
	e := NewEWMA(0.0, 0.5)
	contentionInterval := time.Nanosecond * 1 // interval to contend
	go func() {
		ticker := time.NewTicker(contentionInterval)
		for t := range ticker.C {
			e.Push(gaussianTestData[t.Nanosecond()%N]) // a contentious write
		}
	}()
	for i := 0; i < b.N; i++ {
		e.Push(gaussianTestData[i%N])
	}
	result = e.Mean() // to avoid optimizing out the loop entirely
}
