# 🔒 Kavach CLI Security Policy

## 📦 Supported Versions
| Version | Supported          |
| ------- | ------------------ |
| `v0.2.x` | ✅ Yes             |
| `< v0.2.0` | ❌ No (auto-update required) |

## 🛡️ Security Architecture
`kavach-cli` is a **zero-knowledge, client-side encrypted vault**. Key design principles:
- 🔐 **Client-Side Only:** Encryption/decryption happens locally. Server/CDN never sees plaintext.
- 🧹 **Secure Memory Handling:** Passphrases & plaintext are explicitly wiped from RAM after use.
- 🔍 **Integrity Verification:** Every file is SHA-256 hashed pre-encryption. Decryption fails if hash mismatches.
- 🔄 **Supply-Chain Secure Updates:** Binaries are signed with Ed25519. Client verifies signature & SHA-256 before atomic swap.
- 🚫 **No Telemetry/Analytics:** Zero network calls except explicit `kavach update` fetch.

## 📋 Threat Model (STRIDE)
| Threat | Mitigation |
|--------|------------|
| **S**poofing (fake update server) | Ed25519 signature verification + embedded public key |
| **T**ampering (file corruption) | `age` AEAD tags + SHA-256 manifest verification |
| **R**epudiation (action denial) | Local audit logs + immutable `.kavach/` permissions (`0700`) |
| **I**nformation Disclosure | Zero-knowledge design, memory wipe, no plaintext logging |
| **D**enial of Service | Rate-limited update checks, local-only operations, no cloud dependencies |
| **E**levation of Privilege | Strict file permissions, no root required, sandboxed temp files |

## 🐛 Reporting a Vulnerability
We take security seriously. If you find a vulnerability:
1. **DO NOT** open a public issue.
2. Email: `security@yourdomain.com` (replace with your email)
3. Include:
   - Affected version & OS
   - Step-by-step reproduction
   - Impact assessment
   - Suggested fix (optional)
4. We will acknowledge within **48 hours** and aim to patch within **14 days**.

## 🔍 Responsible Disclosure Policy
- Researchers who report valid vulnerabilities will be credited in `CHANGELOG.md` & `SECURITY.md`.
- We do not offer monetary bounties (yet), but will publicly acknowledge ethical contributions.
- Testing must be done on **your own isolated environment**. Attacking others' systems is illegal.

## 🧪 Known Limitations
- `kavach-cli` is CLI-only. GUI/Desktop wrappers are third-party.
- Swap/pagefile memory is not locked (requires `mlock` + `CAP_IPC_LOCK`).
- Recovery via mnemonic phrase is not yet implemented (planned for v1.0).
- Always keep offline backups of `.kavach/` & your passphrase.

---
*Last Updated: 2026-04-19 | Maintainer: @zexxxxxxx*
