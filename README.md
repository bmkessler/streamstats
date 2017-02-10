# streamstats
[![GoDoc](https://godoc.org/github.com/bmkessler/streamstats?status.svg)](https://godoc.org/github.com/bmkessler/streamstats)
[![Build Status](https://travis-ci.org/bmkessler/streamstats.svg?branch=master)](https://travis-ci.org/bmkessler/streamstats)
[![Coverage Status](https://coveralls.io/repos/github/bmkessler/streamstats/badge.svg?branch=master)](https://coveralls.io/github/bmkessler/streamstats?branch=master)

Streaming stats data structures and algorithms in golang that are `O(1)` time and space in the number of elements processed.

## Moment-based Statistics
Single variable moments up to fourth order and first-order covariance use the methods of:

"Formulas for Robust, One-Pass Parallel Computation of Covariances and Arbitrary-Order Statistical Moments." 
Philippe P. Pébay,
Technical Report SAND2008-6212, Sandia National Laboratories, September 2008.

which extend the results of:

"Note on a method for calculating corrected sums of squares and products". 
 B. P. Welford (1962).
 Technometrics 4(3):419–420
(popularized by Donald Knuth in "The Art of Computer Programming") 

to arbitrary moments and combinations of arbitrary sized populations allowing parallel aggregation.
These moments are also extended to two dependent variables with a covariance `Sxy`

This also an includes exponentially-weighted moving average with damping factor, 0 < *lambda* < 1, 
using update formula `m = (1-lambda)*m + lambda*x`

## Order Statistics

Quantiles and Histograms are based on the P2-algorithm:

"The P2 algorithm for dynamic calculation of quantiles and histograms without storing observations."
Raj Jain and Imrich Chlamtac,
Communications of the ACM 
Volume 28 Issue 10, October 1985
Pages 1076-1085

## Count Distinct

Count distinct is provided by an implementation of the HyperLogLog data structure based on:

"Hyperloglog: The analysis of a near-optimal cardinality estimation algorithm"
Philippe Flajolet and Éric Fusy and Olivier Gandouet and et al.
in AOFA ’07: PROCEEDINGS OF THE 2007 INTERNATIONAL CONFERENCE ON ANALYSIS OF ALGORITHMS

This implementation includes some of the HyperLogLog++ enhancements such as the 64-bit hash function
which eliminates the large cardinality correction for hash collisions and an empirical bias correction for small cardinalities
The implementation is space in-efficient since bits are used to store the counts which could be at most 60 < 2^6

An additional LinearCounting implementation that is backed by a BitVector is available as well.  If the maximum possible 
cardinality is known, this structure uses only 12.5% of the memory as the HyperLogLog and runs much faster for both `Add` and `Distinct`.
However, the data structure saturates at the maximum value while HyperLogLog can count to virtually unlimited cardinalities.

## Set Membership

Approximate set membership is provided by a BloomFilter implementation based on:

"Space/time trade-offs in hash coding with allowable errors"
Burton H. Bloom
Communications of the ACM
Volume 13 Issue 7, July 1970
Pages 422-426

the size m of the filter is rounded up to the nearest power of two for speed of addition and membership check, 
which could result in a larger filter depending on the cardinality and false positive target you supply.
the k different hash functions are derived from top (`h1`) and bottom (`h2`) 32-bits of a 64-bit hash function using
`h[i] = h1 + i* h2 mod m for i in 0...m-1` based on

"Less hashing, same performance: Building a better Bloom filter"
Adam Kirsch, Michael Mitzenmacher 
Random Structures & Algorithms
Volume 33 Issue 2, September 2008
Pages 187-218 

## Benchmarks
```
Intel(R) Core(TM) i3-4010U CPU @ 1.70GHz
go version go1.7.3 linux/amd64
BenchmarkBloomFilterAdd-4              	20000000	        75.5 ns/op
BenchmarkBloomFilterCheck-4            	20000000	        70.0 ns/op
BenchmarkEWMAAdd-4                     	200000000	         8.28 ns/op
BenchmarkHyperLogLogP10Add-4           	30000000	        56.3 ns/op
BenchmarkHyperLogLogP10Distinct-4      	 1000000	      2178 ns/op
BenchmarkLinearCountingP10Add-4        	50000000	        36.1 ns/op
BenchmarkLinearCountingP10Distinct-4   	10000000	       175 ns/op
BenchmarkMomentStatsAdd-4              	100000000	        19.7 ns/op
BenchmarkP2Histogram8Add-4             	10000000	       165 ns/op
BenchmarkP2Histogram16Add-4            	 5000000	       349 ns/op
BenchmarkP2Histogram32Add-4            	 2000000	       702 ns/op
BenchmarkP2Histogram64Add-4            	 1000000	      1365 ns/op
BenchmarkP2Histogram128Add-4           	  500000	      2673 ns/op
BenchmarkP2QuantileAdd-4               	20000000	        66.4 ns/op
```
