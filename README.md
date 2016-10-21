# streamstats
Streaming stats `O(1)` data structures and algorithms for golang

## Moment-based Statistics
Single variable moments up to fourth order and first-order covariance use the methods of:
Philippe P. Pébay,
"Formulas for Robust, One-Pass Parallel Computation of Covariances and Arbitrary-Order Statistical Moments." 
Technical Report SAND2008-6212, Sandia National Laboratories, September 2008.

which extend the results of B. P. Welford (1962).
"Note on a method for calculating corrected sums of squares and products". Technometrics 4(3):419–420

(popularized by Donald Knuth in "The Art of Computer Programming") 
to arbitrary moments and combinations of arbitrary sized populations allowing parallel aggregation

## Order Statistics

Quantiles and Histograms are based on the P2-algorithm:

Raj Jain and Imrich Chlamtac,
"The P2 algorithm for dynamic calculation of quantiles and histograms without storing observations."
Commun. ACM 28, 10 (October 1985), 1076-1085.

## Exponentially Weighted Moving Average

Stores a damped average with damping factor, 0 < *lambda* < 1, using update `m = (1-lambda)*m + lambda*x`


## Benchmarks
```
2.8 GHz Intel Core i7
16 GB 1600 MHz DDR3
go version go1.7.3 darwin/amd64
BenchmarkMomentStatsPush-8                  	20000000	       106 ns/op
BenchmarkMomentStatsPushReadContention-8    	20000000	       116 ns/op
BenchmarkMomentStatsPushWriteContention-8   	10000000	       123 ns/op
BenchmarkP2Histogram8Push-8                 	10000000	       153 ns/op
BenchmarkP2Histogram16Push-8                	10000000	       216 ns/op
BenchmarkP2Histogram32Push-8                	 5000000	       352 ns/op
BenchmarkP2Histogram64Push-8                	 2000000	       629 ns/op
BenchmarkP2Histogram128Push-8               	 1000000	      1184 ns/op
BenchmarkP2QuantilePush-8                   	10000000	       145 ns/op
BenchmarkP2QuantilePushReadContention-8     	10000000	       153 ns/op
BenchmarkP2QuantilePushWriteContention-8    	10000000	       183 ns/op
```
