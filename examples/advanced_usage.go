//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"time"

	jd "github.com/nimish-nirmal/jitter-distill"
)

func main() {
	fmt.Println("=== Advanced Jitter Distill Usage ===\n")

	// Custom configuration
	cfg := jd.DefaultConfig()
	cfg.WorkerCount = 8
	cfg.SampleCount = 1024
	cfg.ReseedPeriod = 15 * time.Second
	cfg.Salt = []byte("production-salt-2024")

	pool := jd.NewEntropyPool(cfg)
	defer pool.Close()

	// Generate multiple token types
	fmt.Println("Generating mixed token strengths:")
	for i := 0; i < 3; i++ {
		token256, _ := pool.GenerateToken()
		fmt.Printf("256-bit: %s\n", token256[:32]+"...")
	}

	for i := 0; i < 2; i++ {
		token512, _ := pool.GenerateTokenWithStrength(512)
		fmt.Printf("512-bit: %s\n", token512[:32]+"...")
	}

	// Monitor pool health
	bytes, tokens, workers := pool.Stats()
	fmt.Printf("\nPool Health:\n")
	fmt.Printf("  Bytes: %d\n", bytes)
	fmt.Printf("  Tokens: %d\n", tokens)
	fmt.Printf("  Workers: %d\n", workers)
}
