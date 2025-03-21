# metrics - lightweight package for exporting metrics in Prometheus format

```
goos: darwin
goarch: arm64
pkg: go.withmatt.com/metrics
cpu: Apple M1 Max
BenchmarkCounterGetOrCreate
BenchmarkCounterGetOrCreate/hot
BenchmarkCounterGetOrCreate/hot-10     	51054565	        23.24 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterGetOrCreate/cold
BenchmarkCounterGetOrCreate/cold-10    	 7587610	       157.5 ns/op	      96 B/op	       3 allocs/op
BenchmarkCounterInc
BenchmarkCounterInc-10                 	175391628	         6.839 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterIncParallel
BenchmarkCounterIncParallel-10         	14670920	        80.69 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter
BenchmarkExpfmtWriter/name
BenchmarkExpfmtWriter/name-10          	203007757	         5.907 ns/op	3047.24 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_tags
BenchmarkExpfmtWriter/name_with_tags-10         	28778386	        41.65 ns/op	 840.36 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_many_tags
BenchmarkExpfmtWriter/name_with_many_tags-10    	 6676158	       179.3 ns/op	 775.25 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/uint64
BenchmarkExpfmtWriter/uint64-10                 	21796863	        55.42 ns/op	 721.75 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/float64
BenchmarkExpfmtWriter/float64-10                	14011434	        85.72 ns/op	 489.94 MB/s	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdate
BenchmarkHistogramUpdate-10                     	100000000	        11.87 ns/op	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdateParallel
BenchmarkHistogramUpdateParallel-10             	 2637985	       448.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkWritePrometheus
BenchmarkWritePrometheus-10                     	  509984	      2352 ns/op	1203.67 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate
BenchmarkValidate/MustIdent
BenchmarkValidate/MustIdent-10                  	54330500	        22.02 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/validateIdent
BenchmarkValidate/validateIdent-10              	76251796	        14.98 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/MustValue
BenchmarkValidate/MustValue-10                  	75957976	        15.70 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	go.withmatt.com/metrics	18.646s
```
