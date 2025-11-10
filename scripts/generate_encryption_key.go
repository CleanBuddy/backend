//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/cleanbuddy/backend/internal/utils"
)

// Generate a random encryption key for AES-256
// Run with: go run scripts/generate_encryption_key.go
func main() {
	key, err := utils.GenerateEncryptionKey()
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	fmt.Println("Generated AES-256 Encryption Key (base64-encoded):")
	fmt.Println(key)
	fmt.Println()
	fmt.Println("Add this to your .env file:")
	fmt.Printf("ENCRYPTION_KEY=%s\n", key)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Store this key securely!")
	fmt.Println("   - Never commit it to version control")
	fmt.Println("   - Use GCP Secret Manager or equivalent in production")
	fmt.Println("   - If lost, encrypted data cannot be recovered")
}
