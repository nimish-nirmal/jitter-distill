package jitterdistill

import (
	"testing"
)

func BenchmarkGenerateToken256(b *testing.B) {
	pool := NewEntropyPool(DefaultConfig())
	defer func() { _ = pool.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.GenerateToken()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateToken512(b *testing.B) {
	cfg := DefaultConfig()
	pool := NewEntropyPool(cfg)
	defer func() { _ = pool.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.GenerateTokenWithStrength(512)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkARPredictorUpdate(b *testing.B) {
	p := NewARPredictor(8, 0.0001)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Update(float64(i))
	}
}

func BenchmarkHarvestJitter(b *testing.B) {
	pool := NewEntropyPool(DefaultConfig())
	defer func() { _ = pool.Close() }()
	buf := make([]float64, DefaultSampleCount)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.harvestJitter(buf)
	}
}

func BenchmarkCompressToken(b *testing.B) {
	token := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompressToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompressToken(b *testing.B) {
	token := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	compressed, _ := CompressToken(token)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecompressToken(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}
