# metrics - lightweight package for exporting metrics in Prometheus format

```
goos: darwin
goarch: arm64
pkg: go.withmatt.com/metrics
cpu: Apple M1 Max
BenchmarkCounterGetOrCreate
BenchmarkCounterGetOrCreate/hot
BenchmarkCounterGetOrCreate/hot-10     	49523253	        24.68 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterGetOrCreate/cold
BenchmarkCounterGetOrCreate/cold-10    	 7437841	       161.3 ns/op	      96 B/op	       3 allocs/op
BenchmarkCounterInc
BenchmarkCounterInc-10                 	172056392	         6.984 ns/op	       0 B/op	       0 allocs/op
BenchmarkCounterIncParallel
BenchmarkCounterIncParallel-10         	15404413	        75.01 ns/op	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter
BenchmarkExpfmtWriter/name
BenchmarkExpfmtWriter/name-10          	154412713	         7.785 ns/op	2312.27 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_tags
BenchmarkExpfmtWriter/name_with_tags-10         	27872998	        43.65 ns/op	 801.80 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/name_with_many_tags
BenchmarkExpfmtWriter/name_with_many_tags-10    	 6427512	       187.0 ns/op	 743.28 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/uint64
BenchmarkExpfmtWriter/uint64-10                 	21411715	        57.25 ns/op	 698.75 MB/s	       0 B/op	       0 allocs/op
BenchmarkExpfmtWriter/float64
BenchmarkExpfmtWriter/float64-10                	13505794	        90.42 ns/op	 464.48 MB/s	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdate
BenchmarkHistogramUpdate-10                     	100000000	        12.01 ns/op	       0 B/op	       0 allocs/op
BenchmarkHistogramUpdateParallel
BenchmarkHistogramUpdateParallel-10             	 2701826	       412.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkWritePrometheus
BenchmarkWritePrometheus-10                     	  500120	      2427 ns/op	1166.65 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate
BenchmarkValidate/MustIdent
BenchmarkValidate/MustIdent-10                  	56988512	        20.94 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/validateIdent
BenchmarkValidate/validateIdent-10              	79696489	        15.34 ns/op	       0 B/op	       0 allocs/op
BenchmarkValidate/MustValue
BenchmarkValidate/MustValue-10                  	75651510	        16.09 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	go.withmatt.com/metrics	18.585s
```
