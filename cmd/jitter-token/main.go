package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	jd "github.com/nimish-nirmal/jitter-distill"
)

func main() {
	count := flag.Int("count", 3, "Number of tokens to generate")
	bits := flag.Int("bits", 256, "Token strength in bits (256 or 512)")
	workers := flag.Int("workers", 4, "Number of harvester goroutines")
	bench := flag.Bool("bench", false, "Run benchmark mode")
	benchDuration := flag.Duration("bench-duration", 10*time.Second, "Benchmark duration")
	sign := flag.String("sign", "", "Sign a token with private key (path to private.pem)")
	verify := flag.String("verify", "", "Verify a signed token (path to public.pem)")
	token := flag.String("token", "", "Token to verify (use with --verify)")
	genKeys := flag.Bool("gen-keys", false, "Generate RSA key pair")
	keyBits := flag.Int("key-bits", 2048, "Key size for generation")
	flag.Parse()

	if *genKeys {
		runGenerateKeys(*keyBits)
		return
	}

	if *sign != "" {
		runSign(*sign, *bits, *workers)
		return
	}

	if *verify != "" && *token != "" {
		runVerify(*verify, *token)
		return
	}

	if *bench {
		runBenchmark(*benchDuration, *bits)
		return
	}

	cfg := jd.DefaultConfig()
	cfg.WorkerCount = *workers

	pool := jd.NewEntropyPool(cfg)
	defer func() { _ = pool.Close() }()

	fmt.Printf("Jitter Distill Entropy Pool Initialized\n")
	fmt.Printf("   Workers: %d | Bits: %d\n\n", *workers, *bits)

	for i := 0; i < *count; i++ {
		tok, err := pool.GenerateTokenWithStrength(*bits)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Token #%d: %s\n", i+1, tok)
	}

	bytesGen, tokensGen, active := pool.Stats()
	fmt.Printf("\nStats: %d bytes | %d tokens | %d active workers\n", bytesGen, tokensGen, active)
}

func runGenerateKeys(bits int) {
	privatePEM, publicPEM, err := jd.GenerateKeyPair(bits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating keys: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("private.pem", privatePEM, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing private key: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile("public.pem", publicPEM, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d-bit RSA key pair:\n", bits)
	fmt.Printf("  Private key: private.pem (chmod 600)\n")
	fmt.Printf("  Public key:  public.pem (share freely)\n")
}

func runSign(privateKeyPath string, bits, workers int) {
	privateKeyPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading private key: %v\n", err)
		os.Exit(1)
	}

	cfg := jd.DefaultConfig()
	cfg.WorkerCount = workers
	pool := jd.NewEntropyPool(cfg)
	defer func() { _ = pool.Close() }()

	token, err := pool.GenerateTokenWithStrength(bits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	signed, err := jd.SignToken(token, privateKeyPEM)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Token: %s\n", signed.Token)
	fmt.Printf("Signature: %x\n", signed.Signature)
	fmt.Printf("\nTo verify (offline, no server):\n")
	fmt.Printf("  jitter-token --verify public.pem --token %s\n", signed.Token)
}

func runVerify(publicKeyPath, token string) {
	publicKeyPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading public key: %v\n", err)
		os.Exit(1)
	}

	// Parse signature from stdin or env var (for demo, we'll expect base64)
	fmt.Print("Enter signature (hex): ")
	var signatureHex string
	if _, err := fmt.Scanln(&signatureHex); err != nil {
		fmt.Fprintf(os.Stderr, "error reading signature: %v\n", err)
		os.Exit(1)
	}

	signature := make([]byte, len(signatureHex)/2)
	for i := 0; i < len(signature); i++ {
		if _, err := fmt.Sscanf(signatureHex[i*2:i*2+2], "%02x", &signature[i]); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing signature byte: %v\n", err)
			os.Exit(1)
		}
	}

	signed := &jd.SignedToken{
		Token:     token,
		Signature: signature,
	}

	valid := jd.VerifySignedToken(signed, publicKeyPEM)
	if valid {
		fmt.Println("\n✓ Token is VALID (verified offline, no server needed)")
	} else {
		fmt.Println("\n✗ Token is INVALID")
		os.Exit(1)
	}
}

func runBenchmark(duration time.Duration, bits int) {
	cfg := jd.DefaultConfig()
	pool := jd.NewEntropyPool(cfg)
	defer func() { _ = pool.Close() }()

	fmt.Printf("Running benchmark for %v...\n", duration)

	start := time.Now()
	count := 0
	for time.Since(start) < duration {
		_, err := pool.GenerateTokenWithStrength(bits)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		count++
	}

	elapsed := time.Since(start)
	fmt.Printf("Generated %d tokens in %v (%.2f tokens/sec)\n", count, elapsed, float64(count)/elapsed.Seconds())
}
