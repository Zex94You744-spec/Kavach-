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
	"github.com/zex94you744-spec/kavach-cli/internal/ux"
)

const (
	vaultDir     = ".kavach"
	manifestFile = ".kavach/manifest.json"
)

type FileRecord struct {
	OriginalName string `json:"original"`
	SHA256       string `json:"sha256"`
	AddedAt      string `json:"added_at"`
}

type Manifest struct {
	Version string                `json:"version"`
	Files   map[string]FileRecord `json:"files"`
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
	ux.Verbose("Loading manifest: " + manifestFile)
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
	ux.Verbose("Saving updated manifest")
	return os.WriteFile(manifestFile, data, 0600)
}

func Init() {
	if _, err := os.Stat(vaultDir); err == nil {
		ux.Warning("Vault already initialized. Skipping.")
		return
	}
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		ux.Error(fmt.Sprintf("Failed to create vault: %v", err))
		return
	}
	ux.Success("Vault initialized at .kavach/")
	ux.Tip("Keep your passphrase safe. Loss = permanent data lock.")
}

func readPassphrase() []byte {
	ux.Verbose("Prompting for passphrase (hidden input)")
	fmt.Print("Enter passphrase: ")
	bytePass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		ux.Error("Failed to read passphrase")
		os.Exit(1)
	}
	return bytePass
}

func AddFile(filePath string) {
	ux.Verbose("Reading plaintext: " + filePath)
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		ux.Error(fmt.Sprintf("Read error: %v", err))
		return
	}
	defer secureWipe(plaintext)

	hash := calculateHash(plaintext)
	encName := filepath.Base(filePath) + ".enc"

	pass := readPassphrase()
	defer secureWipe(pass)

	ux.Verbose("Encrypting with age (Scrypt + ChaCha20-Poly1305)")
	encData, err := crypto.Encrypt(plaintext, pass)
	if err != nil {
		ux.Error(fmt.Sprintf("Encrypt error: %v", err))
		return
	}

	ux.Verbose("Writing encrypted file")
	if err := os.WriteFile(filepath.Join(vaultDir, encName), encData, 0600); err != nil {
		ux.Error(fmt.Sprintf("Write error: %v", err))
		return
	}

	manifest, err := loadManifest()
	if err != nil {
		ux.Error(fmt.Sprintf("Manifest error: %v", err))
		return
	}
	manifest.Files[encName] = FileRecord{
		OriginalName: filepath.Base(filePath),
		SHA256:       hash,
		AddedAt:      time.Now().Format(time.RFC3339),
	}
	if err := saveManifest(manifest); err != nil {
		ux.Error(fmt.Sprintf("Manifest save failed: %v", err))
		return
	}

	ux.Success(fmt.Sprintf("Encrypted & hashed: %s → %s/%s", filePath, vaultDir, encName))
	ux.Tip("Run 'kavach audit' anytime to verify file integrity.")
}

func PullFile(encName string) {
	encPath := filepath.Join(vaultDir, encName)
	ux.Verbose("Reading encrypted file: " + encPath)
	encData, err := os.ReadFile(encPath)
	if err != nil {
		ux.Error(fmt.Sprintf("Read error: %v", err))
		return
	}

	pass := readPassphrase()
	defer secureWipe(pass)

	ux.Verbose("Decrypting in memory")
	decData, err := crypto.Decrypt(encData, pass)
	if err != nil {
		ux.Error(fmt.Sprintf("Decryption failed: %v", err))
		return
	}

	manifest, err := loadManifest()
	if err != nil {
		ux.Error(fmt.Sprintf("Manifest read error: %v", err))
		secureWipe(decData)
		return
	}

	record, exists := manifest.Files[encName]
	if !exists {
		ux.Warning("File not found in manifest. Possible external addition.")
	} else if calculateHash(decData) != record.SHA256 {
		ux.Error("🚨 INTEGRITY FAILURE: File hash mismatch. Tampered or corrupted!")
		secureWipe(decData)
		return
	}

	outName := strings.TrimSuffix(encName, ".enc")
	ux.Verbose("Writing decrypted file to disk")
	if err := os.WriteFile(outName, decData, 0600); err != nil {
		ux.Error(fmt.Sprintf("Write error: %v", err))
		secureWipe(decData)
		return
	}

	secureWipe(decData)
	ux.Success(fmt.Sprintf("Decrypted & verified: %s → %s", encPath, outName))
}

func Audit() {
	manifest, err := loadManifest()
	if err != nil || len(manifest.Files) == 0 {
		ux.Info("No files in vault to audit.")
		return
	}

	pass := readPassphrase()
	defer secureWipe(pass)

	ux.Info("Starting integrity audit...")
	allOK := true

	for encName, record := range manifest.Files {
		encPath := filepath.Join(vaultDir, encName)
		ux.Verbose("Checking: " + encName)
		encData, err := os.ReadFile(encPath)
		if err != nil {
			ux.Warning(fmt.Sprintf("%s: Missing or unreadable", encName))
			allOK = false
			continue
		}

		decData, err := crypto.Decrypt(encData, pass)
		if err != nil {
			ux.Warning(fmt.Sprintf("%s: Decryption failed", encName))
			allOK = false
			continue
		}

		if calculateHash(decData) == record.SHA256 {
			ux.Success(fmt.Sprintf("%s: OK (%s)", encName, record.OriginalName))
		} else {
			ux.Error(fmt.Sprintf("%s: TAMPERED/CORRUPTED!", encName))
			allOK = false
		}
		secureWipe(decData)
	}

	if allOK {
		ux.Info("✨ All files verified. Vault integrity intact.")
	} else {
		ux.Warning("Audit failed. Restore from backup or re-encrypt originals.")
	}
}
