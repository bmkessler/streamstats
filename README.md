# streamstats
[![GoDoc](https://godoc.org/github.com/bmkessler/streamstats?status.svg)](https://godoc.org/github.com/bmkessler/streamstats)
Streaming stats data structures and algorithms in golang that are `O(1)` in the number of elements

## Moment-based Statistics
Single variable moments up to fourth order and first-order covariance use the methods of:

"Formulas for Robust, One-Pass Parallel Computation of Covariances and Arbitrary-Order Statistical Moments." 
Philippe P. Pébay,
Technical Report SAND2008-6212, Sandia National Laboratories, September 2008.

which extend the results of B. P. Welford (1962).
"Note on a method for calculating corrected sums of squares and products". Technometrics 4(3):419–420

(popularized by Donald Knuth in "The Art of Computer Programming") 
to arbitrary moments and combinations of arbitrary sized populations allowing parallel aggregation.

These moments are also extended to two dependent variables with a covariance `Sxy`

Also includes exponentially-weighted moving average with damping factor, 0 < *lambda* < 1, using update `m = (1-lambda)*m + lambda*x`

## Order Statistics

Quantiles and Histograms are based on the P2-algorithm:

"The P2 algorithm for dynamic calculation of quantiles and histograms without storing observations."
Raj Jain and Imrich Chlamtac,
Commun. ACM 28, 10 (October 1985), 1076-1085.

## Count Distinct

An implementation of the HyperLogLog data structure based on:

"Hyperloglog: The analysis of a near-optimal cardinality estimation algorithm"
Philippe Flajolet and Éric Fusy and Olivier Gandouet and et al.
in AOFA ’07: PROCEEDINGS OF THE 2007 INTERNATIONAL CONFERENCE ON ANALYSIS OF ALGORITHMS

This implementation includes some of the HyperLogLog++ enhancements such as the 64-bit hash function
which eliminates the large cardinality correction for hash collisions and an empirical bias correction for small cardinalities
this is also space in-efficient since bytes are used to store the counts which could be at most 60 < 2^6

## Benchmarks
```
Intel(R) Core(TM) i3-4010U CPU @ 1.70GHz
go version go1.7.3 linux/amd64
BenchmarkEWMAPush-4                 	200000000	         8.27 ns/op
BenchmarkHyperLogLogP10Add-4        	30000000	        56.0 ns/op
BenchmarkHyperLogLogP10Distinct-4   	 1000000	      2168 ns/op
BenchmarkMomentStatsPush-4          	100000000	        19.8 ns/op
BenchmarkP2Histogram8Push-4         	20000000	       110 ns/op
BenchmarkP2Histogram16Push-4        	 5000000	       257 ns/op
BenchmarkP2Histogram32Push-4        	 3000000	       521 ns/op
BenchmarkP2Histogram64Push-4        	 1000000	      1114 ns/op
BenchmarkP2Histogram128Push-4       	 1000000	      2178 ns/op
BenchmarkP2QuantilePush-4           	20000000	        70.3 ns/op
```
