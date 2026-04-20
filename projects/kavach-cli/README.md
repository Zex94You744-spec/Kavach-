# 🔒 Kavach CLI - Zero-Knowledge Secure File Vault

[![Release](https://img.shields.io/github/v/release/Zex94You744-spec/Kavach-?label=version&color=blue)](https://github.com/Zex94You744-spec/Kavach-/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/Zex94You744-spec/Kavach-/releases)
[![CI/CD](https://github.com/Zex94You744-spec/Kavach-/actions/workflows/release.yml/badge.svg)](https://github.com/Zex94You744-spec/Kavach-/actions)

> 🔐 **Client-side encrypted vault with signed auto-updates**  
> 📱 **Beginner-friendly CLI + Menu UI**  
> 🛡️ **Zero-knowledge by design — server never sees your data**

---

## 🚀 Quick Start (Mobile/Termux Users)

### 1️⃣ Install (Pre-built Binary)
```bash
# Download latest Linux ARM64 binary (for mobile/proot)
curl -LO https://raw.githubusercontent.com/Zex94You744-spec/kavach-updates/main/kavach-linux-arm64
curl -LO https://raw.githubusercontent.com/Zex94You744-spec/kavach-updates/main/kavach-linux-arm64.sig

# Verify signature (optional but recommended)
# Requires: public.pem from repo
# openssl pkeyutl -verify -pubin -inkey public.pem -in kavach-linux-arm64 -signature kavach-linux-arm64.sig

# Make executable
chmod +x kavach-linux-arm64
mv kavach-linux-arm64 kavach

# Test
./kavach --version
