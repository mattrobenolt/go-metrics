# metrics - lightweight package for exporting metrics in Prometheus format

Heavily inspired and based on [VictoriaMetrics/metrics](https://github.com/VictoriaMetrics/metrics).

```
goos: darwin
goarch: arm64
pkg: x
cpu: Apple M1 Max
                                 │  mattware   │                vmmetrics                │                prometheus                 │
                                 │   sec/op    │    sec/op     vs base                   │    sec/op      vs base                    │
IncWithLabelValues-10              27.54n ± 2%    95.75n ± 1%  +247.76% (p=0.000 n=10)       47.36n ± 0%    +72.02% (p=0.000 n=10)
UpdateVMRangeHistogram-10          15.91n ± 1%    16.98n ± 0%    +6.72% (p=0.000 n=10)
UpdatePromHistogram-10             8.127n ± 1%                                              10.035n ± 0%    +23.48% (p=0.000 n=10)
WriteMetricsCounters-10            704.2µ ± 5%   1033.3µ ± 1%   +46.72% (p=0.000 n=10)     10068.0µ ± 6%  +1329.60% (p=0.000 n=10)
WriteMetricsVMRangeHistograms-10   462.8µ ± 1%   2373.2µ ± 2%  +412.83% (p=0.000 n=10)
WriteMetricsPromHistograms-10      108.2µ ± 1%                                               437.4µ ± 3%   +304.36% (p=0.000 n=10)
geomean                            2.238µ         7.947µ       +129.88%                ¹     6.764µ        +232.88%                ¹
¹ benchmark set differs from baseline; geomeans may not be comparable

                                 │    mattware    │                   vmmetrics                    │                  prometheus                   │
                                 │      B/op      │      B/op       vs base                        │      B/op       vs base                       │
IncWithLabelValues-10                 0.00 ± 0%         48.00 ± 0%            ? (p=0.000 n=10)            0.00 ± 0%           ~ (p=1.000 n=10) ¹
UpdateVMRangeHistogram-10            0.000 ± 0%         0.000 ± 0%            ~ (p=1.000 n=10) ¹
UpdatePromHistogram-10               0.000 ± 0%                                                          0.000 ± 0%           ~ (p=1.000 n=10) ¹
WriteMetricsCounters-10            80.00Ki ± 0%     1260.89Ki ± 0%    +1476.09% (p=0.000 n=10)       2944.91Ki ± 0%   +3581.09% (p=0.000 n=10)
WriteMetricsVMRangeHistograms-10     896.0 ± 0%     2685034.5 ± 0%  +299569.03% (p=0.000 n=10)
WriteMetricsPromHistograms-10        896.0 ± 0%                                                       229918.0 ± 0%  +25560.49% (p=0.000 n=10)
geomean                                         ²                   ?                          ³ ²                     +885.85%                ³ ²
¹ all samples are equal
² summaries must be >0 to compute geomean
³ benchmark set differs from baseline; geomeans may not be comparable

                                 │   mattware   │                    vmmetrics                    │                   prometheus                    │
                                 │  allocs/op   │   allocs/op     vs base                         │   allocs/op     vs base                         │
IncWithLabelValues-10              0.000 ± 0%         1.000 ± 0%             ? (p=0.000 n=10)           0.000 ± 0%             ~ (p=1.000 n=10) ¹
UpdateVMRangeHistogram-10          0.000 ± 0%         0.000 ± 0%             ~ (p=1.000 n=10) ¹
UpdatePromHistogram-10             0.000 ± 0%                                                           0.000 ± 0%             ~ (p=1.000 n=10) ¹
WriteMetricsCounters-10            1.000 ± 0%     10018.000 ± 0%  +1001700.00% (p=0.000 n=10)       40115.000 ± 0%  +4011400.00% (p=0.000 n=10)
WriteMetricsVMRangeHistograms-10   1.000 ± 0%     41422.000 ± 0%  +4142100.00% (p=0.000 n=10)
WriteMetricsPromHistograms-10      1.000 ± 0%                                                        5434.000 ± 0%   +543300.00% (p=0.000 n=10)
geomean                                       ²                   ?                           ³ ²                     +12050.85%                ³ ²
¹ all samples are equal
² summaries must be >0 to compute geomean
³ benchmark set differs from baseline; geomeans may not be comparable

                                 │   mattware    │                vmmetrics                │               prometheus               │
                                 │      B/s      │      B/s       vs base                  │     B/s       vs base                  │
WriteMetricsCounters-10            689.13Mi ± 5%   478.91Mi ± 1%  -30.50% (p=0.000 n=10)     48.23Mi ± 6%  -93.00% (p=0.000 n=10)
WriteMetricsVMRangeHistograms-10   1069.8Mi ± 1%    208.6Mi ± 2%  -80.50% (p=0.000 n=10)
WriteMetricsPromHistograms-10       811.6Mi ± 1%                                             200.6Mi ± 3%  -75.29% (p=0.000 n=10)
geomean                             842.6Mi         316.0Mi       -63.19%                ¹   98.35Mi       -86.85%                ¹
¹ benchmark set differs from baseline; geomeans may not be comparable

```

```
goos: darwin
goarch: arm64
pkg: go.withmatt.com/metrics
cpu: Apple M1 Max
BenchmarkCounterGetOrCreate
BenchmarkCounterGetOrCreate/hot
BenchmarkCounterGetOrCreate/hot-10     	47642600	        24.43 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterGetOrCreate/cold
BenchmarkCounterGetOrCreate/cold-10    	 7289173	       165.7 ns/op	      96 B/op	       3 allocs/op
BenchmarkCounterGetOrCreate/verycold
BenchmarkCounterGetOrCreate/verycold-10         	  733803	      1549 ns/op	     456 B/op	      16 allocs/op
BenchmarkCounterInc
BenchmarkCounterInc-10                          	173853097	         6.895 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterIncParallel
BenchmarkCounterIncParallel-10                  	16201200	        73.84 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter
BenchmarkExpfmtWriter/name
BenchmarkExpfmtWriter/name-10                   	201310080	         5.973 ns/op	3013.66 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_tags
BenchmarkExpfmtWriter/name_with_tags-10         	26503352	        44.46 ns/op	 787.31 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_many_tags
BenchmarkExpfmtWriter/name_with_many_tags-10    	 6595731	       182.5 ns/op	 761.59 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/uint64
BenchmarkExpfmtWriter/uint64-10                 	21459403	        56.55 ns/op	 707.33 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/float64
BenchmarkExpfmtWriter/float64-10                	13296459	        89.02 ns/op	 471.83 MB/s	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdate
BenchmarkHistogramUpdate-10                     	100000000	        10.88 ns/op	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdateParallel
BenchmarkHistogramUpdateParallel-10             	12915625	        94.55 ns/op	       0 B/op	       0 allocs/op
BenchmarkWritePrometheus
BenchmarkWritePrometheus-10                     	  497890	      2476 ns/op	1143.52 MB/s	      64 B/op	       1 allocs/op
BenchmarkValidate
BenchmarkValidate/MustIdent
BenchmarkValidate/MustIdent-10                  	57545100	        20.93 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/validateIdent
BenchmarkValidate/validateIdent-10              	79759168	        15.38 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/MustValue
BenchmarkValidate/MustValue-10                  	73344400	        16.22 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	go.withmatt.com/metrics	19.263s
```
