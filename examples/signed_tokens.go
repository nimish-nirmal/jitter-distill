//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	jd "github.com/nimish-nirmal/jitter-distill"
)

func main() {
	fmt.Println("=== Signed Token Example: Server + Client Verification ===\n")

	// Server: Generate entropy pool and create tokens
	pool := jd.NewEntropyPool(jd.DefaultConfig())
	defer pool.Close()

	// Generate a token
	token, err := pool.GenerateToken()
	if err != nil {
		panic(err)
	}

	// Server: Generate RSA key pair (2048-bit)
	privateKeyPEM, publicKeyPEM, err := jd.GenerateKeyPair(2048)
	if err != nil {
		panic(err)
	}

	// Server: Sign token with private key
	signed, err := jd.SignToken(token, privateKeyPEM)
	if err != nil {
		panic(err)
	}

	// In real scenario, you'd send signed.Token + signed.Signature to client
	// For demo, we'll verify locally
	fmt.Printf("Token: %s\n", signed.Token)
	fmt.Printf("Signature: %x\n", signed.Signature[:32])

	// Client: Verify token WITHOUT server access (using public key only)
	valid := jd.VerifySignedToken(signed, publicKeyPEM)
	fmt.Printf("\nVerification (offline): %v\n", valid)

	// Try tampering with token
	tampered := &jd.SignedToken{
		Token:     "tampered_value",
		Signature: signed.Signature,
	}
	valid2 := jd.VerifySignedToken(tampered, publicKeyPEM)
	fmt.Printf("Tampered verification: %v (should be false)\n", valid2)

	// Save keys to files (real scenario)
	os.WriteFile("private.pem", privateKeyPEM, 0600)
	os.WriteFile("public.pem", publicKeyPEM, 0644)
	fmt.Println("\nKeys saved to private.pem and public.pem")
}
