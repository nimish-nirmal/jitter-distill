//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"time"

	jd "github.com/nimish-nirmal/jitter-distill"
)

func main() {
	fmt.Println("=== Jitter Distill - Basic Usage Example ===\n")

	// Create a pool with default settings
	pool := jd.NewEntropyPool(jd.DefaultConfig())
	defer pool.Close()

	// Wait a moment for workers to start
	time.Sleep(500 * time.Millisecond)

	// Generate multiple tokens
	fmt.Println("Generating 256-bit tokens:")
	for i := 1; i <= 3; i++ {
		token, err := pool.GenerateToken()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("%d. %s\n", i, token)
	}

	// Generate a 512-bit token
	fmt.Println("\nGenerating 512-bit token:")
	strongToken, err := pool.GenerateTokenWithStrength(512)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Strong: %s\n", strongToken)
	}

	// Check pool statistics
	bytes, tokens, workers := pool.Stats()
	fmt.Printf("\n=== Pool Statistics ===\n")
	fmt.Printf("Bytes generated: %d\n", bytes)
	fmt.Printf("Tokens generated: %d\n", tokens)
	fmt.Printf("Active workers: %d\n", workers)
}
