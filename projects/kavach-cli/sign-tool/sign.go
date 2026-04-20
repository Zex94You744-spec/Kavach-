package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run sign.go <private.pem> <input> <output.sig>")
		os.Exit(1)
	}

	pemData, err := os.ReadFile(os.Args[1])
	if err != nil { fmt.Printf("❌ Read error: %v\n", err); os.Exit(1) }

	block, _ := pem.Decode(pemData)
	if block == nil { fmt.Println("❌ Failed to parse PEM"); os.Exit(1) }

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil { fmt.Printf("❌ Parse error: %v\n", err); os.Exit(1) }

	privKey, ok := key.(ed25519.PrivateKey)
	if !ok { fmt.Println("❌ Not Ed25519 key"); os.Exit(1) }

	data, err := os.ReadFile(os.Args[2])
	if err != nil { fmt.Printf("❌ Input error: %v\n", err); os.Exit(1) }

	sig := ed25519.Sign(privKey, data)
	if err := os.WriteFile(os.Args[3], sig, 0600); err != nil {
		fmt.Printf("❌ Write error: %v\n", err); os.Exit(1)
	}
	fmt.Printf("✅ Signature created: %s (%d bytes)\n", os.Args[3], len(sig))
}
