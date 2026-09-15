# Security

This document is a summary of the security review performed on go-chat and the
hardening applied, plus the items that remain as conscious design decisions or
future work. It is not a formal audit.

## Threat model

go-chat is a peer-to-peer chat. Every participant is both a client and a
server: anyone who discovers your address (mDNS on the LAN, a shared multiaddr,
or a relay) can open a connection to you. Consequently the trust boundary is
**any connected peer**, including malicious or compromised ones. All behaviour
below aims to protect against:

- Impersonation and message forgery.
- Remote code execution via terminal escape-sequence injection.
- Resource exhaustion / DoS through floods of messages, streams, or tunnels.
- Extraction of identity key material or private conversation data.

## Hardening applied

| Area | Change |
| :-- | :-- |
| Terminal injection | New `internal/safe` package strips ANSI/OSC/C0 control sequences from incoming message content, display names, and channel/org names before storage (`internal/network/stream.go`) and again at every TUI render point. |
| Inbound flood control | Per-peer token-bucket rate limiting (20 msg/s, burst 60) on every inbound message (`internal/network/host.go`, `internal/network/stream.go`). |
| Sync amplification | Sync batches bounded (`maxSyncMessages = 20000`) in `handleSyncRequest`. |
| Stream DoS | Global cap on concurrent inbound streams (`maxConcurrentStreams = 256`) in `Node.handleStream`. |
| Line parsing bug | `readLine` now accumulates across buffer chunks; large (>4 KiB) protocol lines are no longer silently truncated. |
| Key file permissions | SQLite DB and its `-wal`/`-shm` sidecars are chmod'd `0600` after open (`internal/storage/storage.go`). Identity Ed25519 private key lives only in this DB. |
| Tunnel server hang | Data-back handoff uses a non-blocking channel send so a second client dial can no longer block the handler goroutine forever (`internal/tunnel/tunnel.go`). |
| Network timeouts | `/publicip` HTTP lookup now uses a 10 s timeout instead of the default no-timeout client. |

## Existing protections

- Transport encryption + peer authentication via libp2p Noise.
- E2E message encryption: X25519 key exchange, HKDF session key, AES-256-GCM.
- Per-message Ed25519 signatures binding type/sender/channel/content/timestamp;
  timestamps must fall within ±5 minutes.
- Sync is signature-gated: `handleMessage` rejects any message whose sender
  key cannot be resolved, so a connected peer cannot spoof another known peer's
  identity (signature over the sender ID prevents it).
- Private-channel and DM access checks on both sync and message ingestion.
- No telemetry, no external services, no cloud dependencies.

## Remaining advisories (design decisions / future work)

These are intentional trade-offs or roadmap items, not known exploitable bugs.

1. **mDNS trust-on-first-use.** Peers discovered on the LAN are connected and
   fully synced automatically. A hostile device on the same network can read
   public-channel history and org/channel metadata. Mitigating factors: private
   channels and DMs are excluded from sync state for non-members, and message
   content requires the E2E session key. Recommended future work: a
   "trust / block peer" prompt on first discovery (the `peers.trusted` /
   `peers.blocked` columns already exist in the schema).
2. **Session keys not separated by direction.** Both directions currently use
   one HKDF-derived AES key with random 12-byte nonces; nonce collisions across
   the combined stream are possible in principle. The clean fix (derive two
   keys, one per direction) changes the wire protocol and is left for a
   protocol-version bump.
3. **Tunnel server is unauthenticated.** The public tunnel relay assigns
   predictable ports (`20000 + id`) to any caller and does not authenticate
   clients. It is not part of the encryption path (tunneled libp2p traffic is
   Noise-protected), but a public tunnel server can be abused as a bandwidth
   relay and its port space is predictable. Consider adding a shared-secret
   token via config for publicly exposed servers.
4. **At-rest encryption.** Content is decrypted locally and stored in SQLite in
   plaintext (0600 perms). Database-level encryption (e.g. SQLCipher / a
   passphrase-derived key, `database.encrypt` placeholder is already documented)
   is a planned enhancement.
5. **Guessable IDs.** Org/channel/invite IDs are derived from wall-clock
   timestamps. IDs are not an authorization boundary (all state changes are
   signature-gated), but they are guessable; fine for collaborative tools,
   undesirable if a public tunnel server were to use IDs as secrets.
6. **PLAN.md diverges from implementation.** The roadmap advertises Argon2id +
   ChaCha20-Poly1305; the implemented stack is X25519 + AES-256-GCM + HKDF +
   Ed25519, which is sound. PLAN.md should be updated to match (see top-level
   PLAN.md).
7. **File transfer / sync / notifications** packages are still stubs
   (milestone-2). When file transfer lands, enforce the configured size caps
   and always scan hashes; whitelist paths for downloads.

## Reporting

Security-sensitive issues should be reported privately via GitHub
(e.g. opening an issue with minimal repro details) rather than posting
exploits in public channels.