//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	jd "github.com/nimish-nirmal/jitter-distill"
)

func main() {
	fmt.Println("=== Offline Token Verification Example ===")

	// Step 1: Generate key pair
	fmt.Println("1. Generating RSA-2048 key pair...")
	privatePEM, publicPEM, err := jd.GenerateKeyPair(2048)
	if err != nil {
		panic(err)
	}
	os.WriteFile("server_private.pem", privatePEM, 0600)
	os.WriteFile("client_public.pem", publicPEM, 0644)
	fmt.Println("   Keys saved")

	// Step 2: Server generates and signs token
	fmt.Println("\n2. Server: Generating token and signing...")
	pool := jd.NewEntropyPool(jd.DefaultConfig())
	defer pool.Close()

	token, _ := pool.GenerateToken()
	signed, _ := jd.SignToken(token, privatePEM)
	fmt.Printf("   Token: %s...", token[:32])
	fmt.Printf("   Signature: %x...", signed.Signature[:32])

	// Step 3: Client verifies OFFLINE
	fmt.Println("\n3. Client: Verifying offline with public key...")
	clientPublicPEM, _ := os.ReadFile("client_public.pem")
	valid := jd.VerifySignedToken(signed, clientPublicPEM)
	fmt.Printf("   Verification: %v\n", valid)

	// Step 4: Test tampering
	fmt.Println("\n4. Testing tampering detection...")
	tampered := &jd.SignedToken{
		Token:     "fake",
		Signature: signed.Signature,
	}
	valid2 := jd.VerifySignedToken(tampered, clientPublicPEM)
	fmt.Printf("   Tampered: %v (should be false)\n", valid2)

	fmt.Println("\n=== Summary ===")
	fmt.Println("Server: private.pem")
	fmt.Println("Client: public.pem (no server needed!)")
}
