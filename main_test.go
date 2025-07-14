package main

import "testing"

func benchmarkGetDependents(i int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		GetDependents("inconshreveable", "mousetrap", QueryDependentsConfig{MaxPages: i})
	}
}

func BenchmarkGetDependents1(b *testing.B)  { benchmarkGetDependents(1, b) }
func BenchmarkGetDependents2(b *testing.B)  { benchmarkGetDependents(2, b) }
func BenchmarkGetDependents3(b *testing.B)  { benchmarkGetDependents(3, b) }
func BenchmarkGetDependents10(b *testing.B) { benchmarkGetDependents(10, b) }
func BenchmarkGetDependents20(b *testing.B) { benchmarkGetDependents(20, b) }
func BenchmarkGetDependents40(b *testing.B) { benchmarkGetDependents(40, b) }
