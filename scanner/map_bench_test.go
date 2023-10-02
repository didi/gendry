package scanner

import (
	"testing"
)

var p = 12
var a = person{"deen", 22, 1, &p}

// go test -run=BenchmarkMap -bench=BenchmarkMap -cpu=1,2,4,8 -benchtime=20000000x -benchmem
func BenchmarkMapWithCache(b *testing.B) {
	EnableMapNameCache(true)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = Map(&a, DefaultTagName)
		}
	})
}

func BenchmarkMapDisableCache(b *testing.B) {
	EnableMapNameCache(false)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = Map(&a, DefaultTagName)
		}
	})
}
