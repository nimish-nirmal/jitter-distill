package jitterdistill

import (
	"math"
	"strings"
	"sync"
	"testing"
)

func TestNewARPredictor(t *testing.T) {
	p := NewARPredictor(8, 0.0001)
	if p == nil {
		t.Fatal("expected non-nil predictor")
	}
	if p.order != 8 {
		t.Errorf("expected order 8, got %d", p.order)
	}
}

func TestARPredictorUpdate(t *testing.T) {
	p := NewARPredictor(4, 0.001)
	residual := p.Update(100.0)
	if math.IsNaN(residual) || math.IsInf(residual, 0) {
		t.Errorf("unexpected residual: %v", residual)
	}
}

func TestARPredictorReset(t *testing.T) {
	p := NewARPredictor(4, 0.001)
	p.Update(100.0)
	p.Update(200.0)
	p.Reset()
	for _, w := range p.weights {
		if w != 0.0 {
			t.Errorf("expected zero weight after reset, got %f", w)
		}
	}
}

func TestNewEntropyPool(t *testing.T) {
	cfg := DefaultConfig()
	pool := NewEntropyPool(cfg)
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	defer func() { _ = pool.Close() }()
}

func TestEntropyPoolGenerateToken(t *testing.T) {
	pool := NewEntropyPool(DefaultConfig())
	defer func() { _ = pool.Close() }()
	token, err := pool.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("expected 64-char hex token, got %d", len(token))
	}
}

func TestEntropyPoolGenerateToken512(t *testing.T) {
	pool := NewEntropyPool(DefaultConfig())
	defer func() { _ = pool.Close() }()
	token, err := pool.GenerateTokenWithStrength(512)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 128 {
		t.Errorf("expected 128-char hex token, got %d", len(token))
	}
}

func TestEntropyPoolStats(t *testing.T) {
	pool := NewEntropyPool(DefaultConfig())
	defer func() { _ = pool.Close() }()
	_, _, _ = pool.Stats()
}

func TestCompressDecompress(t *testing.T) {
	original := "abcd1234"
	compressed, err := CompressToken(original)
	if err != nil {
		t.Fatalf("compress failed: %v", err)
	}
	decompressed, err := DecompressToken(compressed)
	if err != nil {
		t.Fatalf("decompress failed: %v", err)
	}
	if decompressed != original {
		t.Errorf("expected %s, got %s", original, decompressed)
	}
}

func TestBitDiff(t *testing.T) {
	a := []byte{0b00000000, 0b11111111}
	b := []byte{0b11111111, 0b00000000}
	diff := BitDiff(a, b)
	if diff != 16 {
		t.Errorf("expected 16, got %d", diff)
	}
}

func TestVerifyTokenFormat(t *testing.T) {
	if !VerifyTokenFormat("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Error("expected valid 64-char hex")
	}
	if !VerifyTokenFormat("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Error("expected valid 128-char hex")
	}
	if VerifyTokenFormat("short") {
		t.Error("expected invalid short token")
	}
}

func TestEntropyEstimation(t *testing.T) {
	history := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	bits := EstimateEntropyBits(history)
	if bits < 0 || bits > 256 {
		t.Errorf("unexpected entropy bits: %f", bits)
	}
}

func TestConcurrentTokenGeneration(t *testing.T) {
	pool := NewEntropyPool(DefaultConfig())
	defer func() { _ = pool.Close() }()
	const goroutines = 10
	const perGoroutine = 5
	errCh := make(chan error, goroutines*perGoroutine)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				_, err := pool.GenerateToken()
				if err != nil {
					errCh <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("concurrent generation error: %v", err)
	}
}

func TestGenerateKeyPair(t *testing.T) {
	privatePEM, publicPEM, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	if len(privatePEM) == 0 {
		t.Error("private key is empty")
	}
	if len(publicPEM) == 0 {
		t.Error("public key is empty")
	}
	if !strings.Contains(string(privatePEM), "RSA PRIVATE KEY") {
		t.Error("private key missing PEM header")
	}
	if !strings.Contains(string(publicPEM), "PUBLIC KEY") {
		t.Error("public key missing PEM header")
	}
}

func TestSignAndVerifyToken(t *testing.T) {
	privatePEM, publicPEM, err := GenerateKeyPair(2048)
	if err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	token := "test_token_256bit_hex_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	signed, err := SignToken(token, privatePEM)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if signed.Token != token {
		t.Errorf("token mismatch: got %s, want %s", signed.Token, token)
	}
	if len(signed.Signature) == 0 {
		t.Error("signature is empty")
	}

	// Verify with correct public key
	if !VerifySignedToken(signed, publicPEM) {
		t.Error("failed to verify valid token")
	}

	// Tamper with token
	tampered := &SignedToken{
		Token:     "tampered",
		Signature: signed.Signature,
	}
	if VerifySignedToken(tampered, publicPEM) {
		t.Error("verified tampered token (should fail)")
	}

	// Tamper with signature
	tampered2 := &SignedToken{
		Token:     token,
		Signature: []byte("invalid"),
	}
	if VerifySignedToken(tampered2, publicPEM) {
		t.Error("verified tampered signature (should fail)")
	}
}
