// internal/vault/vault.go
package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/zex94you744-spec/kavach-cli/internal/crypto"
)

const (
	vaultDir     = ".kavach"
	manifestFile = ".kavach/manifest.json"
)

// FileRecord stores metadata for integrity verification
type FileRecord struct {
	OriginalName string `json:"original"`
	SHA256       string `json:"sha256"`
	AddedAt      string `json:"added_at"`
}

// Manifest holds all vault file records
type Manifest struct {
	Version string                `json:"version"`
	Files   map[string]FileRecord `json:"files"` // key: encrypted filename
}

func secureWipe(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
	runtime.KeepAlive(buf)
}

func calculateHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func loadManifest() (*Manifest, error) {
	data, err := os.ReadFile(manifestFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Manifest{Version: "1.0", Files: make(map[string]FileRecord)}, nil
		}
		return nil, err
	}
	var m Manifest
	return &m, json.Unmarshal(data, &m)
}

func saveManifest(m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestFile, data, 0600)
}

func Init() {
	if _, err := os.Stat(vaultDir); err == nil {
		fmt.Println("⚠️  Vault already initialized.")
		return
	}
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		fmt.Printf("❌ Failed to create vault: %v\n", err)
		return
	}
	fmt.Println("✅ Vault initialized at .kavach/")
	fmt.Println("🔑 Keep your passphrase safe. Loss = permanent data lock.")
	fmt.Println("💡 Tip: Use 12+ chars, mix case/numbers/symbols.")
}

func readPassphrase() []byte {
	fmt.Print("Enter passphrase: ")
	bytePass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Println("❌ Failed to read passphrase")
		os.Exit(1)
	}
	return bytePass
}

func AddFile(filePath string) {
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Read error: %v\n", err)
		return
	}
	defer secureWipe(plaintext)

	hash := calculateHash(plaintext)
	encName := filepath.Base(filePath) + ".enc"

	pass := readPassphrase()
	defer secureWipe(pass)

	encData, err := crypto.Encrypt(plaintext, pass)
	if err != nil {
		fmt.Printf("❌ Encrypt error: %v\n", err)
		return
	}

	// Save encrypted file
	if err := os.WriteFile(filepath.Join(vaultDir, encName), encData, 0600); err != nil {
		fmt.Printf("❌ Write error: %v\n", err)
		return
	}

	// Update manifest
	manifest, err := loadManifest()
	if err != nil {
		fmt.Printf("❌ Manifest error: %v\n", err)
		return
	}
	manifest.Files[encName] = FileRecord{
		OriginalName: filepath.Base(filePath),
		SHA256:       hash,
		AddedAt:      time.Now().Format(time.RFC3339),
	}
	if err := saveManifest(manifest); err != nil {
		fmt.Printf("❌ Failed to update manifest: %v\n", err)
		return
	}

	fmt.Printf("✅ Encrypted & hashed: %s → %s/%s\n", filePath, vaultDir, encName)
}

func PullFile(encName string) {
	encPath := filepath.Join(vaultDir, encName)
	encData, err := os.ReadFile(encPath)
	if err != nil {
		fmt.Printf("❌ Read error: %v\n", err)
		return
	}

	pass := readPassphrase()
	defer secureWipe(pass)

	decData, err := crypto.Decrypt(encData, pass)
	if err != nil {
		fmt.Printf("❌ Decryption failed: %v\n", err)
		return
	}

	// 🔍 Integrity verification BEFORE writing to disk
	manifest, err := loadManifest()
	if err != nil {
		fmt.Printf("❌ Manifest read error: %v\n", err)
		secureWipe(decData)
		return
	}

	record, exists := manifest.Files[encName]
	if !exists {
		fmt.Println("⚠️  File not found in manifest. Possible tampering.")
		secureWipe(decData)
		return
	}

	if calculateHash(decData) != record.SHA256 {
		fmt.Println("🚨 INTEGRITY FAILURE: File hash mismatch. Tampered or corrupted!")
		secureWipe(decData)
		return
	}

	outName := strings.TrimSuffix(encName, ".enc")
	if err := os.WriteFile(outName, decData, 0600); err != nil {
		fmt.Printf("❌ Write error: %v\n", err)
		secureWipe(decData)
		return
	}

	secureWipe(decData) // Wipe after successful write
	fmt.Printf("✅ Decrypted & verified: %s → %s\n", encPath, outName)
}

func Audit() {
	manifest, err := loadManifest()
	if err != nil || len(manifest.Files) == 0 {
		fmt.Println("📭 No files in vault to audit.")
		return
	}

	pass := readPassphrase()
	defer secureWipe(pass)

	fmt.Println("🔍 Starting integrity audit...")
	allOK := true

	for encName, record := range manifest.Files {
		encPath := filepath.Join(vaultDir, encName)
		encData, err := os.ReadFile(encPath)
		if err != nil {
			fmt.Printf("⚠️  %s: Missing or unreadable\n", encName)
			allOK = false
			continue
		}

		decData, err := crypto.Decrypt(encData, pass)
		if err != nil {
			fmt.Printf("🔒 %s: Decryption failed (wrong passphrase?)\n", encName)
			allOK = false
			continue
		}

		if calculateHash(decData) == record.SHA256 {
			fmt.Printf("✅ %s: OK (%s)\n", encName, record.OriginalName)
		} else {
			fmt.Printf("🚨 %s: TAMPERED/CORRUPTED!\n", encName)
			allOK = false
		}
		secureWipe(decData)
	}

	if allOK {
		fmt.Println("✨ All files verified. Vault integrity intact.")
	} else {
		fmt.Println("⚠️  Audit failed. Restore from backup or re-encrypt originals.")
	}
}
