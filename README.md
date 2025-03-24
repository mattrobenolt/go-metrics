# metrics - lightweight package for exporting metrics in Prometheus format

Heavily inspired and based on [VictoriaMetrics/metrics](https://github.com/VictoriaMetrics/metrics).

```
goos: darwin
goarch: arm64
pkg: x
cpu: Apple M1 Max
                          │   mattware    │               vmmetrics                │               prometheus                │
                          │    sec/op     │    sec/op      vs base                 │    sec/op      vs base                  │
IncWithLabelValues-10        27.55n ±  1%   223.75n ± 89%  +712.01% (p=0.000 n=10)   112.10n ±  9%   +306.82% (p=0.000 n=10)
UpdateHistogram-10           31.81n ± 40%    39.73n ± 24%         ~ (p=0.105 n=10)    21.21n ± 13%    -33.31% (p=0.001 n=10)
WriteMetricsCounters-10      1.855m ± 14%    2.473m ±  9%   +33.33% (p=0.000 n=10)   29.788m ± 16%  +1506.10% (p=0.000 n=10)
WriteMetricsHistograms-10   1022.7µ ± 30%   4472.6µ ± 31%  +337.33% (p=0.000 n=10)    590.8µ ± 56%    -42.23% (p=0.001 n=10)
geomean                      6.385µ          17.71µ        +177.32%                   14.30µ         +123.99%

                          │    mattware    │                  vmmetrics                   │                 prometheus                  │
                          │      B/op      │      B/op       vs base                      │      B/op       vs base                     │
IncWithLabelValues-10          0.00 ± 0%         48.00 ± 0%            ? (p=0.000 n=10)          0.00 ± 0%           ~ (p=1.000 n=10) ¹
UpdateHistogram-10            0.000 ± 0%         0.000 ± 0%            ~ (p=1.000 n=10) ¹       0.000 ± 0%           ~ (p=1.000 n=10) ¹
WriteMetricsCounters-10     80.00Ki ± 0%     1260.87Ki ± 0%    +1476.06% (p=0.000 n=10)     2944.97Ki ± 0%   +3581.17% (p=0.000 n=10)
WriteMetricsHistograms-10     896.0 ± 0%     2685028.0 ± 0%  +299568.30% (p=0.000 n=10)      229918.5 ± 0%  +25560.55% (p=0.000 n=10)
geomean                                  ²                   ?                          ²                     +885.85%                ²
¹ all samples are equal
² summaries must be >0 to compute geomean

                          │   mattware   │                   vmmetrics                   │                  prometheus                   │
                          │  allocs/op   │   allocs/op     vs base                       │   allocs/op     vs base                       │
IncWithLabelValues-10       0.000 ± 0%         1.000 ± 0%             ? (p=0.000 n=10)         0.000 ± 0%             ~ (p=1.000 n=10) ¹
UpdateHistogram-10          0.000 ± 0%         0.000 ± 0%             ~ (p=1.000 n=10) ¹       0.000 ± 0%             ~ (p=1.000 n=10) ¹
WriteMetricsCounters-10     1.000 ± 0%     10018.000 ± 0%  +1001700.00% (p=0.000 n=10)     40115.000 ± 0%  +4011400.00% (p=0.000 n=10)
WriteMetricsHistograms-10   1.000 ± 0%     41422.000 ± 0%  +4142100.00% (p=0.000 n=10)      5434.000 ± 0%   +543300.00% (p=0.000 n=10)
geomean                                ²                   ?                           ²                     +12050.85%                ²
¹ all samples are equal
² summaries must be >0 to compute geomean

                          │    mattware    │               vmmetrics                │              prometheus               │
                          │      B/s       │      B/s        vs base                │      B/s       vs base                │
WriteMetricsCounters-10     261.69Mi ± 16%   200.19Mi ±  8%  -23.50% (p=0.000 n=10)   16.29Mi ± 14%  -93.78% (p=0.000 n=10)
WriteMetricsHistograms-10    484.5Mi ± 23%    110.9Mi ± 44%  -77.11% (p=0.000 n=10)   151.5Mi ± 37%  -68.74% (p=0.000 n=10)
geomean                      356.1Mi          149.0Mi        -58.16%                  49.67Mi        -86.05%
```

```
goos: darwin
goarch: arm64
pkg: go.withmatt.com/metrics
cpu: Apple M1 Max
BenchmarkCounterGetOrCreate
BenchmarkCounterGetOrCreate/hot
BenchmarkCounterGetOrCreate/hot-10     	50434580	        23.49 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterGetOrCreate/cold
BenchmarkCounterGetOrCreate/cold-10    	 7522981	       158.2 ns/op	      96 B/op	       3 allocs/op
BenchmarkCounterInc
BenchmarkCounterInc-10                 	175348912	         6.844 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterIncParallel
BenchmarkCounterIncParallel-10         	14449213	        82.76 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter
BenchmarkExpfmtWriter/name
BenchmarkExpfmtWriter/name-10          	202410867	         5.919 ns/op	3040.89 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_tags
BenchmarkExpfmtWriter/name_with_tags-10         	27399110	        43.87 ns/op	 797.90 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_many_tags
BenchmarkExpfmtWriter/name_with_many_tags-10    	 6609460	       181.3 ns/op	 766.84 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/uint64
BenchmarkExpfmtWriter/uint64-10                 	21335118	        56.29 ns/op	 710.64 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/float64
BenchmarkExpfmtWriter/float64-10                	13731331	        87.52 ns/op	 479.88 MB/s	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdate
BenchmarkHistogramUpdate-10                     	100000000	        11.84 ns/op	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdateParallel
BenchmarkHistogramUpdateParallel-10             	 2518386	       479.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkWritePrometheus
BenchmarkWritePrometheus-10                     	  498320	      2390 ns/op	1184.32 MB/s	      64 B/op	       1 allocs/op
BenchmarkValidate
BenchmarkValidate/MustIdent
BenchmarkValidate/MustIdent-10                  	54489726	        22.05 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/validateIdent
BenchmarkValidate/validateIdent-10              	80555158	        14.93 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/MustValue
BenchmarkValidate/MustValue-10                  	76044422	        15.72 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	go.withmatt.com/metrics	18.787s
```
