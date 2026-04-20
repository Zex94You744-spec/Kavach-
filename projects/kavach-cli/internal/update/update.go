// internal/update/update.go
package update

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/zex94you744-spec/kavach-cli/internal/ux"
)

const (
	// 🔒 Update server URL (change to your production HTTPS URL)
        updateServerURL = "https://raw.githubusercontent.com/Zex94You744-spec/kavach-updates/main"
	// 🔑 Embed your Ed25519 PUBLIC KEY here (64 hex chars)
	embeddedPubKeyHex = "31236d5cc1248ff5d354071e8954b3900eb17b7cb840ba21aa8a18d05063a8cd"
	// 📦 Current CLI version
	currentVersion = "v0.3.0"
)

type UpdateInfo struct {
	Version      string `json:"version"`
	BinaryURL    string `json:"binary_url"`
	SignatureURL string `json:"signature_url"`
	SHA256       string `json:"sha256"`
}

func getBinaryName() string {
	return fmt.Sprintf("kavach-%s-%s", runtime.GOOS, runtime.GOARCH)
}

// CheckUpdate fetches manifest & checks version
func CheckUpdate() (bool, UpdateInfo, error) {
	resp, err := http.Get(updateServerURL + "/latest.json")
	if err != nil {
		return false, UpdateInfo{}, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, UpdateInfo{}, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return false, UpdateInfo{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if info.Version == currentVersion {
		return false, info, nil
	}
	return true, info, nil
}

// ApplyUpdate downloads, verifies, and atomically replaces binary
func ApplyUpdate() error {
	ux.Verbose("Checking for updates...")
	isNew, info, err := CheckUpdate()
	if err != nil {
		return fmt.Errorf("update check failed: %w", err)
	}
	if !isNew {
		ux.Success("Already on latest version: " + currentVersion)
		return nil
	}

	ux.Info(fmt.Sprintf("New version available: %s (current: %s)", info.Version, currentVersion))

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)
	tmpBin := filepath.Join(execDir, "kavach_update.tmp")
	tmpSig := tmpBin + ".sig"
	backupPath := execPath + ".bak"

	ux.Verbose("Downloading binary & signature...")
	if err := downloadFile(info.BinaryURL, tmpBin); err != nil {
		return fmt.Errorf("binary download failed: %w", err)
	}
	if err := downloadFile(info.SignatureURL, tmpSig); err != nil {
		os.Remove(tmpBin)
		return fmt.Errorf("signature download failed: %w", err)
	}

	ux.Verbose("Verifying SHA-256 integrity...")
	hash, err := calculateFileHash(tmpBin)
	if err != nil {
		cleanup(tmpBin, tmpSig)
		return err
	}
	if hash != info.SHA256 {
		cleanup(tmpBin, tmpSig)
		return fmt.Errorf("binary hash mismatch (expected %s, got %s)", info.SHA256, hash)
	}

	ux.Verbose("Verifying Ed25519 signature...")
	pubKey, _ := hex.DecodeString(embeddedPubKeyHex)
	binData, _ := os.ReadFile(tmpBin)
	sigData, _ := os.ReadFile(tmpSig)
	if !ed25519.Verify(pubKey, binData, sigData) {
		cleanup(tmpBin, tmpSig)
		ux.Error("🚨 Signature verification failed. Update rejected.")
		return fmt.Errorf("signature verification failed")
	}

	ux.Verbose("Applying atomic swap...")
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	if err := os.Rename(tmpBin, execPath); err != nil {
		ux.Warning("Update failed, rolling back safely...")
		os.Rename(backupPath, execPath)
		cleanup(tmpBin, tmpSig)
		return fmt.Errorf("update failed, rolled back safely: %w", err)
	}

	os.Chmod(execPath, 0755)
	cleanup(tmpSig, "")

	ux.Success("Update applied successfully. Restart to use new version.")
	ux.Info("Backup saved at: " + backupPath)
	return nil
}

// Helpers
func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func calculateFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func cleanup(files ...string) {
	for _, f := range files {
		if f != "" {
			os.Remove(f)
		}
	}
}
