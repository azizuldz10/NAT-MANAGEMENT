package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	fmt.Println("🔐 Generating RSA key pair untuk JWT signing...")

	// Generate 2048-bit RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("❌ Error generating private key: %v\n", err)
		os.Exit(1)
	}

	// Convert private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Convert public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		fmt.Printf("❌ Error marshaling public key: %v\n", err)
		os.Exit(1)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Write private key to file
	err = os.WriteFile("private.key", privateKeyPEM, 0600) // Read/write owner only
	if err != nil {
		fmt.Printf("❌ Error writing private key: %v\n", err)
		os.Exit(1)
	}

	// Write public key to file
	err = os.WriteFile("public.key", publicKeyPEM, 0644) // Read-only for others
	if err != nil {
		fmt.Printf("❌ Error writing public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ RSA key pair generated successfully!")
	fmt.Println("📁 private.key - Keep this secure, never share!")
	fmt.Println("📁 public.key - Can be shared for token verification")
	fmt.Println("")
	fmt.Println("🔒 Security Tips:")
	fmt.Println("   - Store private.key dengan permissions 600")
	fmt.Println("   - Backup private key securely")
	fmt.Println("   - Never commit private key ke git")
	fmt.Println("   - Rotate keys regularly (every 6-12 months)")
}
