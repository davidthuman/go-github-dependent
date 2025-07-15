# Go GitHub Dependent

Go Package for querying the dependent repositories within GitHub.

## Benchmarking

```
go test -run=XXX -bench=.
```

Current benchmarks are

```
$ go test -run=XXX -bench=.
goos: darwin
goarch: arm64
pkg: github.com/davidthuman/go-github-dependent
cpu: Apple M2 Pro
BenchmarkGetDependents1-10             7         142916357 ns/op               142.9 ms/page
BenchmarkGetDependents2-10             6         180301438 ns/op                90.08 ms/page
BenchmarkGetDependents3-10             4         281642708 ns/op                93.83 ms/page
BenchmarkGetDependents10-10            1        2481623709 ns/op               248.1 ms/page
BenchmarkGetDependents20-10            1        4095069458 ns/op               204.8 ms/page
BenchmarkGetDependents40-10            1        9312551875 ns/op               232.8 ms/page
PASS
ok      github.com/davidthuman/go-github-dependent      19.303s
```

```
$ go test -run=XXX -bench=^BenchmarkGetDependentsProducerConsumer  
goos: darwin
goarch: arm64
pkg: github.com/davidthuman/go-github-dependent
cpu: Apple M2 Pro
BenchmarkGetDependentsProducerConsumer1-10             2         532909417 ns/op               532.5 ms/page
BenchmarkGetDependentsProducerConsumer2-10             5         230293042 ns/op               115.1 ms/page
BenchmarkGetDependentsProducerConsumer3-10             3         345448347 ns/op               115.1 ms/page
BenchmarkGetDependentsProducerConsumer10-10            1        2840526000 ns/op               284.0 ms/page
BenchmarkGetDependentsProducerConsumer20-10            1        4386396500 ns/op               219.3 ms/page
```