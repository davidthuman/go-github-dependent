package main

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func benchmarkGetDependents(i int, b *testing.B) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	for b.Loop() {
		GetDependents("inconshreveable", "mousetrap", QueryDependentsConfig{MaxPages: i})
	}
	b.ReportMetric(float64(b.Elapsed().Milliseconds())/float64(i)/float64(b.N), "ms/page")
}

func BenchmarkGetDependents1(b *testing.B)  { benchmarkGetDependents(1, b) }
func BenchmarkGetDependents2(b *testing.B)  { benchmarkGetDependents(2, b) }
func BenchmarkGetDependents3(b *testing.B)  { benchmarkGetDependents(3, b) }
func BenchmarkGetDependents10(b *testing.B) { benchmarkGetDependents(10, b) }
func BenchmarkGetDependents20(b *testing.B) { benchmarkGetDependents(20, b) }
func BenchmarkGetDependents40(b *testing.B) { benchmarkGetDependents(40, b) }

func BenchmarkParseDependentsPage(b *testing.B) {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	page, err := os.ReadFile("./tests/github-dependents.html")
	if err != nil {
		return
	}
	doc, err := html.Parse(strings.NewReader(string(page)))
	if err != nil {
		return
	}
	for b.Loop() {
		parseDependentsPage(HtmlPage{Page: doc, Url: "https://github.com/inconshreveable/mousetrap/network/dependents?dependents_after=NDgwMjM1NTg5MzA"})
	}
}

func benchmarkGetDependentsProducerConsumer(i int, b *testing.B) {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	for b.Loop() {
		GetDependentsProducerConsumer("inconshreveable", "mousetrap", QueryDependentsConfig{MaxPages: i})
	}
	b.ReportMetric(float64(b.Elapsed().Milliseconds())/float64(i)/float64(b.N), "ms/page")
}

func BenchmarkGetDependentsProducerConsumer1(b *testing.B) {
	benchmarkGetDependentsProducerConsumer(1, b)
}
func BenchmarkGetDependentsProducerConsumer2(b *testing.B) {
	benchmarkGetDependentsProducerConsumer(2, b)
}
func BenchmarkGetDependentsProducerConsumer3(b *testing.B) {
	benchmarkGetDependentsProducerConsumer(3, b)
}
func BenchmarkGetDependentsProducerConsumer10(b *testing.B) {
	benchmarkGetDependentsProducerConsumer(10, b)
}
func BenchmarkGetDependentsProducerConsumer20(b *testing.B) {
	benchmarkGetDependentsProducerConsumer(20, b)
}
func BenchmarkGetDependentsProducerConsumer40(b *testing.B) {
	benchmarkGetDependentsProducerConsumer(40, b)
}
