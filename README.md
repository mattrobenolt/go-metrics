# metrics - lightweight package for exporting metrics in Prometheus format

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
