#!/usr/bin/env bash
set -uo pipefail  # ❗ Removed -e so script continues on non-critical errors

# 📦 Kavach CLI Release Builder
# Usage: ./build-release.sh <version>
# Example: ./build-release.sh v0.2.0

# 🎯 Configuration
PROJECT="kavach"
VERSION="${1:-v0.2.0}"
OUT_DIR="releases/${VERSION}"
SIGN_TOOL="${SIGN_TOOL:-$HOME/sign-tool/sign.go}"
PRIVATE_KEY="${SIGN_KEY:-$HOME/projects/kavach-cli/private.pem}"

# 🌍 GitHub config (lowercase for raw URLs)
GH_USER="zex94you744-spec"
GH_REPO="kavach-updates"
RAW_BASE="https://raw.githubusercontent.com/${GH_USER}/${GH_REPO}/main"

# 📱 Target platforms (OS/ARCH)
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"    # Mobile/Proot target
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

echo "🚀 Building ${PROJECT} ${VERSION}..."
rm -rf releases
mkdir -p "$OUT_DIR"

# 🔨 Build, Sign & Package for each platform
for platform in "${PLATFORMS[@]}"; do
  OS="${platform%%/*}"
  ARCH="${platform##*/}"
  EXT=""
  [ "$OS" = "windows" ] && EXT=".exe"

  BIN="${PROJECT}-${OS}-${ARCH}${EXT}"
  echo "🔨 ${BIN}..."

  # Cross-compile (static, stripped, reproducible)
  CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "${OUT_DIR}/${BIN}" \
    cmd/kavach/main.go

  # 🔏 Ed25519 signature
  echo "  🔏 Signing..."
  if [ -f "$SIGN_TOOL" ]; then
    go run "$SIGN_TOOL" "$PRIVATE_KEY" "${OUT_DIR}/${BIN}" "${OUT_DIR}/${BIN}.sig" || echo "  ⚠️  Signing failed, continuing..."
  else
    echo "  ⚠️  Sign tool not found, skipping signature"
  fi

  # 📦 Package (tar.gz for Unix, zip for Windows)
  if [ "$OS" = "windows" ]; then
    if command -v zip >/dev/null 2>&1; then
      zip -j "${OUT_DIR}/${BIN%.exe}.zip" "${OUT_DIR}/${BIN}" "${OUT_DIR}/${BIN}.sig" >/dev/null 2>&1 && echo "  📦 ${BIN%.exe}.zip created"
    else
      echo "  ⚠️  zip not installed, skipping Windows package (install with: apt install zip)"
    fi
  else
    tar -czf "${OUT_DIR}/${BIN}.tar.gz" -C "$OUT_DIR" "${BIN}" "${BIN}.sig" >/dev/null 2>&1 && echo "  📦 ${BIN}.tar.gz created"
  fi

  echo "  ✅ ${BIN} done"
done

# 🔍 Generate SHA256SUMS (always run this)
echo "🔍 Generating SHA256SUMS..."
cd "$OUT_DIR"
sha256sum "${PROJECT}"-* > SHA256SUMS 2>/dev/null || sha256sum *.exe *.sig *.gz *.zip 2>/dev/null > SHA256SUMS || echo "  ⚠️  Could not generate SHA256SUMS"
cd - >/dev/null || true

# 📝 Generate latest.json manifest (auto SHA256 + correct URLs)
echo "📝 Generating latest.json manifest..."

# Calculate SHA256 for linux-arm64 (primary mobile target)
LINUX_ARM64_HASH=""
if [ -f "${OUT_DIR}/kavach-linux-arm64" ]; then
  LINUX_ARM64_HASH=$(sha256sum "${OUT_DIR}/kavach-linux-arm64" | awk '{print $1}')
elif [ -f "${OUT_DIR}/kavach-linux-amd64" ]; then
  LINUX_ARM64_HASH=$(sha256sum "${OUT_DIR}/kavach-linux-amd64" | awk '{print $1}')
fi

if [ -n "$LINUX_ARM64_HASH" ]; then
  cat > "${OUT_DIR}/latest.json" << MANIFEST
{
  "version": "${VERSION}",
  "binary_url": "${RAW_BASE}/kavach-linux-arm64",
  "signature_url": "${RAW_BASE}/kavach-linux-arm64.sig",
  "sha256": "${LINUX_ARM64_HASH}"
}
MANIFEST
  echo "  ✅ latest.json created with hash: ${LINUX_ARM64_HASH:0:16}..."
else
  echo "  ⚠️  Could not calculate hash, skipping latest.json"
fi

# 🔄 Auto-commit to updates repo if we're inside it
if [ -d ".git" ] && git remote get-url origin 2>/dev/null | grep -q "kavach-updates" 2>/dev/null; then
  git add latest.json SHA256SUMS 2>/dev/null || true
  git commit -m "release: ${VERSION} assets + manifest" 2>/dev/null || true
  echo "  📦 Changes staged for push (if in updates repo)"
fi

# 🎉 Summary
echo ""
echo "🎉 Release ready: ${OUT_DIR}/"
echo "📦 Contents:"
ls -lh "${OUT_DIR}/" 2>/dev/null | grep -E "kavach-|latest|SHA256" || echo "  (listing failed)"
echo ""
echo "💡 To upload to GitHub:"
echo "  cd kavach-updates"
echo "  cp ${OUT_DIR}/* ./"
echo "  git add ."
echo "  git commit -m 'release: ${VERSION}'"
echo "  git push origin main"
echo ""
echo "🌐 Raw URL for CLI:"
echo "  ${RAW_BASE}"
