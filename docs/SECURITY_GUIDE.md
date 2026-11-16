# Security Guide & Threat Modeling
## Personal OS: AI Audio Motivation Engine

---

## 1. Security Philosophy & Perimeter Defense

Although Personal OS is designed as a local-first workstation system, its integration with external LLM APIs, webhooks, OS-level binaries (FFmpeg), and GPU microservices necessitates a strict **Zero-Trust Input Policy**:

- **Subprocess Isolation:** No unvalidated external string may enter shell interpreters.
- **Strict File Sanity:** Audio uploads and reference samples are treated as untrusted binary vectors.
- **Least Privilege Execution:** Microservices run under unprivileged system users with restricted filesystem read/write privileges.

---

## 2. Authorization & RBAC Matrix

For environments exposing the Personal OS Web Studio to local networks or VPNs, access is segmented into three role tiers:

| Resource / Endpoint | `GUEST` (Listener) | `OPERATOR` (Creator) | `ADMIN` (Root) |
| :--- | :---: | :---: | :---: |
| **Stream / Download Audio Tracks** | ✅ Yes | ✅ Yes | ✅ Yes |
| **View System Health & Status** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Trigger Audio Render Jobs** | ❌ Denied | ✅ Yes | ✅ Yes |
| **Upload Voice Reference Files** | ❌ Denied | ✅ Yes | ✅ Yes |
| **Edit System Prompts & Personas** | ❌ Denied | ❌ Denied | ✅ Yes |
| **Purge Storage Cache / Delete Tracks** | ❌ Denied | ❌ Denied | ✅ Yes |
| **Manage API Keys & Webhooks** | ❌ Denied | ❌ Denied | ✅ Yes |

---

## 3. Threat Model & OWASP Vulnerability Mitigations

### 3.1 Subprocess Injection via FFmpeg Arguments (Critical)
- **Threat Vector:** Malicious filenames or unescaped audio metadata passed into `ffmpeg` via shell concatenation can execute arbitrary operating system commands.
- **Vulnerability Example:** `ffmpeg -i "track.wav; rm -rf /" ...`
- **Mitigation:**
  1. Never use `sh -c` or `bash -c`. Always invoke parameterized binaries directly via Go's `os/exec.CommandContext(ctx, "ffmpeg", args...)`.
  2. Enforce strict character whitelisting on generated output names and UUIDs (`^[a-zA-Z0-9_-]+$`).

```go
// CORRECT SECURE IMPLEMENTATION
cmd := exec.CommandContext(ctx, "ffmpeg",
    "-y",
    "-i", sanitizedVocalPath,
    "-i", sanitizedAmbientPath,
    "-filter_complex", filterGraph,
    "-c:a", "libmp3lame",
    "-b:a", "320k",
    sanitizedOutputPath,
)
```

### 3.2 Path Traversal & Audio Upload Tampering (High)
- **Threat Vector:** Uploading crafted filenames like `../../etc/passwd` or embedding malicious polyglot scripts inside fake `.wav` files.
- **Mitigation:**
  1. **Magic Byte Verification:** Verify binary magic bytes rather than trusting `Content-Type` headers or file extensions.
     - WAV files must begin with `RIFF....WAVE`.
     - MP3 files must contain ID3 headers or sync frames (`0xFFFB`).
  2. **Filename Canonicalization:** Force-rename all uploaded reference samples to random UUIDv4 identifiers (`uuid.New().String() + ".wav"`). Never use original client-supplied filenames on disk.
  3. **Filesystem Confinement:** Restrict all file operations to `/storage` using `filepath.Clean()` and `strings.HasPrefix(targetPath, baseDir)`.

### 3.3 Server-Side Request Forgery (SSRF) on Webhooks (Medium)
- **Threat Vector:** Users configuring webhook URLs pointing to local internal services (e.g., `http://169.254.169.254` AWS metadata or `http://localhost:6379` Redis).
- **Mitigation:**
  1. Webhook dispatcher resolves IP addresses prior to dispatch and denies loopback (`127.0.0.1/8`), link-local (`169.254.0.0/16`), and private RFC 1918 subnets (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`).
  2. Sign all outgoing webhook payloads using HMAC-SHA256 (`X-Signature-SHA256`).

### 3.4 Denial of Service via VRAM Exhaustion (Medium)
- **Threat Vector:** Submitting requests with enormous text inputs causing GPU out-of-memory crashes and service failure.
- **Mitigation:**
  1. Enforce strict text length limits (maximum 10,000 characters per job).
  2. Rate limit job submissions via token bucket algorithm.
  3. Python worker drops requests when VRAM usage exceeds 90%.

---

## 4. Secrets Management & Operational Security

1. **Environment Variables:** All API keys (`DEEPSEEK_API_KEY`, `WEBHOOK_SECRET`, `AUTH_SECRET`) are loaded exclusively from `.env` or system environment variables.
2. **Zero-Leak Logging Policy:** Terminal log streaming (SSE) and server stdout automatically sanitize sensitive keys and authorization tokens before broadcast using regex scrubbers.
3. **Storage Security:** The `/storage` directory is mounted with `noexec` flags in production deployments to prevent arbitrary script execution from uploaded assets.

---

## 5. Cryptographic Verification & Audit Integrity

- Every completed master audio track computes an immutable **SHA-256 hash** upon write completion.
- The hash is stored in the database alongside track metadata.
- Verification checks run during startup to ensure audio libraries have not experienced bitrot or unauthorized modification.
