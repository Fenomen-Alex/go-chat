# go-chat

[![Release](https://github.com/Fenomen-Alex/go-chat/actions/workflows/release.yml/badge.svg)](https://github.com/Fenomen-Alex/go-chat/actions/workflows/release.yml)

**Terminal P2P Chat** — A serverless, encrypted, peer-to-peer messaging platform that runs entirely in your terminal.

```
                   _                    
                  | |                   
   ___   __ _  ___| |__   ___  ___  ___ 
  / __| / _` |/ __| '_ \ / _ \/ __|/ __|
 | (__ | (_| | (__| | | | (_) \__ \ (__ 
  \___| \__, |\___|_| |_|\___/|___/\___|
         __/ |                          
        |___/                           
```

## Features

- **No servers, no accounts** — Pure P2P. Every client is an equal peer.
- **End-to-end encrypted** — X25519 key exchange, ChaCha20-Poly1305 AEAD, HKDF session keys.
- **Local-first** — All data stored in SQLite. You own everything.
- **Terminal UI** — Built with Bubble Tea, Lip Gloss, and Bubbles.
- **libp2p networking** — TCP, QUIC, mDNS discovery, relay, NAT traversal, hole punching.
- **Color-coded senders** — Each peer gets a unique color in chat.
- **Organizations & Channels** — Create orgs, channels, manage members and permissions.
- **Private Channels** — Invite-only channels with lock icon markers.
- **Direct Messages** — Encrypted peer-to-peer conversations, shown in dedicated DM panel.
- **Automatic peer discovery** — mDNS for LAN, DHT for internet, bootstrap peers.
- **Ed25519 identities** — Generated locally on first launch. Fingerprint-based verification.
- **Offline-first** — Messages queued locally, synced when peers reconnect.
- **Saved connections** — Recent `/connect` targets saved for quick reconnection.
- **Cross-platform** — Linux amd64/arm64, macOS Intel/Apple Silicon, Windows amd64/arm64.

## Quick Start

### macOS

1. Go to [Releases](https://github.com/Fenomen-Alex/go-chat/releases)
2. Download `chat-darwin-amd64.zip` (Intel) or `chat-darwin-arm64.zip` (Apple Silicon)
3. Double-click the zip to extract — you'll see `chat.app`
4. **Right-click** `chat.app` → **Open** (first time only — macOS security)
5. The chat opens in Terminal — ready to use

> **Troubleshooting:** If you downloaded the standalone binary instead of the `.app` bundle, macOS may block it. Run these commands in Terminal:
> ```bash
> chmod +x chat-darwin-amd64
> xattr -d com.apple.quarantine chat-darwin-amd64
> ./chat-darwin-amd64
> ```

### Windows

1. Go to [Releases](https://github.com/Fenomen-Alex/go-chat/releases)
2. Download `chat-windows-amd64.exe.zip`
3. Extract and double-click `chat-windows-amd64.exe`

### Linux

1. Go to [Releases](https://github.com/Fenomen-Alex/go-chat/releases)
2. Download `chat-linux-amd64.tar.gz` or `chat-linux-arm64.tar.gz`
3. Extract and run `./chat` in terminal

### Terminal one-liner (any OS)

```bash
curl -sfL https://raw.githubusercontent.com/Fenomen-Alex/go-chat/main/install.sh | sh
```

### Build from source

#### Prerequisites

- Go 1.24+

```bash
go build -o chat ./cmd/chat
```

### Run

```bash
./chat
```

On first launch, go-chat automatically:
1. Generates an Ed25519 identity keypair
2. Derives your libp2p Peer ID
3. Creates a local SQLite database
4. Starts listening on a random TCP port
5. Discovers peers via mDNS on the local network

Type `/connect <address>` to connect to a peer. If your display name is still the default (`user-<hostname>`), you'll be prompted to set one on first connect.

## Configuration

Config is loaded from `~/.config/go-chat/config.yaml` (created automatically with defaults).

### Example config.yaml

```yaml
identity:
  display_name: ""             # defaults to user-<hostname>

network:
  port: 0                      # 0 = random port
  enable_relay: true
  enable_mdns: true
  enable_dht: true
  enable_quic: true
  enable_tcp: true
  bootstrap_peers: []          # multiaddrs for bootstrap
  relay_peers: []              # multiaddrs /ip4/.../tcp/.../p2p/...
  # relay_peers:
  #   - /ip4/1.2.3.4/tcp/4001/p2p/12D3KooW...

database:
  path: chat.db
  encrypt: false

downloads:
  path: downloads
  max_size: 104857600          # 100 MB

uploads:
  max_size: 52428800           # 50 MB

theme: dark
notifications:
  desktop: false
  bell: true
  mentions: true

logging:
  level: info                  # trace, debug, info, warn, error
  file: ""
  rotate: true

security:
  key_rotation_days: 30
  encrypt_database: false
```

Override with CLI flags:

```bash
./chat                           # Launch TUI
./chat --serve                   # Headless relay mode (no TUI)
./chat --serve --tunnel :1234    # Relay + TCP tunnel server
./chat --config /path/to/config.yaml
./chat --name "myname"           # Set display name (skips connect prompt)
./chat --version
```

> **Public relay:** Run `chat --serve` on any public server to create a relay that your peers can use without port forwarding.
> **TCP tunnel:** Add `--tunnel :1234` to also run a tunnel server — clients use `/tunnel <ip>:1234`.

## Usage

### TUI Layout

```
+-----------------------------+----------------------------------------+
| Channels                    | Chat                                   |
|                             |                                        |
|  # general                  |  12:30 alice: hello                    |
|  🔒 secret-project          |  12:31 bob: hi                         |
|                             |                                        |
+-----------------------------+                                        |
| Direct Messages             |                                        |
|  @ DM-alice                 |                                        |
+-----------------------------+                                        |
| Status: #general                        [INPUT] Peers: 2            |
+-----------------------------+----------------------------------------+
| Input / Command Line                                                   |
+-----------------------------------------------------------------------+
```

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Cycle: input mode → channels → DMs |
| `Up` / `Down` | Switch channels/DMs (nav mode) / scroll chat (input mode) |
| `Enter` | Send message / switch to input mode |
| `?` | Toggle help |
| `P` | Toggle peers list |
| `L` | Toggle logs overlay |
| `Ctrl+C` / `Ctrl+Q` | Quit |

### Commands

| Command | Description |
|---------|-------------|
| `/help` | Show help |
| `/myaddr` | Show your shareable multiaddress |
| `/connect <multiaddr>` | Connect to a peer directly |
| `/connect <index>` | Reconnect to a saved connection |
| `/connections` | List saved connections |
| `/relay <multiaddr>` | Connect via a relay peer (no port forwarding) |
| `/tunnel <addr>` | Create a TCP tunnel (requires tunnel server) |
| `/disconnect` | Disconnect all peers |
| `/peers` | List known peers |
| `/channel create <name>` | Create a public channel |
| `/channel private <name> [desc]` | Create a private channel (invite only) |
| `/invite <channel_id> <peer_id>` | Invite a peer to a private channel |
| `/invite accept <invite_id>` | Accept a channel invite |
| `/invites` | List pending invites |
| `/dm <peer_id>` | Open a direct message |
| `/name [name]` | Show or set your display name |
| `/profile` | Show your identity info (fingerprint) |
| `/publicip` | Look up your public IP (for port forwarding) |
| `/quit` | Exit |

### Connecting to Peers

#### LAN (automatic)

go-chat discovers peers on the same local network via mDNS automatically — no setup needed.

#### Across the Internet (manual)

To share your address with a remote peer, run `/myaddr` inside the chat. You'll see one or more lines like:

```
=== Peer ID: 12D3KooW... ===
  /ip4/192.168.1.42/tcp/43987/p2p/12D3KooW...
  /ip4/10.0.0.5/tcp/43987/p2p/12D3KooW...
---
Share: /connect /ip4/<public_ip>/tcp/43987/p2p/12D3KooW...
```

### Connect Across the Internet (no router config)

#### Option 1: Use a relay peer (recommended — zero config)

A **relay** is a publicly accessible go-chat instance that forwards traffic between peers. Neither side needs port forwarding.

**Step 1: Find a relay**

Either run your own on a public server (`chat --serve`) or use a public relay shared by the community.

**Step 2: Connect to the relay**

Inside go-chat, type:

```
/relay /ip4/<relay_ip>/tcp/<relay_port>/p2p/<relay_peer_id>
```

Or set `relay_peers` in `config.yaml` to auto-connect on every launch:

```yaml
network:
  relay_peers:
    - /ip4/<relay_ip>/tcp/<relay_port>/p2p/<relay_peer_id>
```

**Step 3: Share your relay address**

Run `/myaddr`. You'll see a relay address like:

```
/ip4/<relay_ip>/tcp/<relay_port>/p2p/<relay_peer_id>/p2p-circuit/p2p/<your_peer_id>
```

The remote peer connects to the same relay and you can communicate directly through it.

> **How it works:** libp2p's AutoRelay automatically detects when a direct connection isn't possible and routes through the relay. Both peers only need *outbound* connections — no router config required.

---

#### Option 2: Run your own relay on a public server

If you have access to any public server (free Oracle Cloud, AWS free tier, $5 VPS, etc.):

```bash
# On the server (no TUI needed)
chat --serve

# You'll see:
# Relay Peer ID: 12D3KooW...
#   /ip4/<server_ip>/tcp/<port>/p2p/12D3KooW...
```

Share the printed address with your peers. They connect with `/relay <address>`.

---

#### Option 3: Direct connection with port forwarding

If you can configure your router, run `/myaddr` and share the address. Forwards the port in your router:

1. Run `/myaddr` to see your port
2. In your router: forward that TCP port to your machine
3. Share your public IP + port + peer ID
4. Remote peer connects with `/connect <address>`

---

#### Option 4: TCP tunnel (no public IP or router config needed)

go-chat has a built-in TCP tunnel client. It works with any public bore-compatible server.

**Usage:**

1. Run a tunnel server on a public machine:
   ```bash
   chat --serve --tunnel :1234
   ```
2. On your machine, inside go-chat:
   ```
   /tunnel <server-ip>:1234
   ```
3. It prints a `/connect <address>` — share that with your peer
4. Peer connects:
   ```
   /connect /ip4/<tunnel_ip>/tcp/<port>/p2p/<peer_id>
   ```

**Run your own tunnel server** (on any public VPS, along with `--serve`):
```bash
chat --serve --tunnel :1234
```
Clients connect with:
```
/tunnel <server-ip>:1234
```

> **How it works:** The tunnel server assigns a public port and forwards TCP connections. Both sides connect *outbound* to the tunnel — no port forwarding needed.

---

#### Bootstrap

Add bootstrap peer multiaddrs to `network.bootstrap_peers` in config for always-on discovery.

## Identity

Your identity is automatically generated on first launch:

- **Display name** — Defaults to `user-<hostname>`, set during first `/connect`
- **Peer ID** — Derived from your Ed25519 public key (e.g., `12D3KooW...`)
- **Keypair** — Ed25519 for identity, X25519 for key exchange
- **Fingerprint** — First 16 bytes of SHA-256 of your public key

### View Identity

```
/profile
```

Example output:
```
Profile: user-myhost | PeerID: 12D3KooWAbCdEf... | Fingerprint: a1b2c3d4e5f6...
```

## Architecture

```
cmd/chat/            CLI entrypoint
internal/
  app/               Application coordinator
  config/            Configuration (YAML/JSON)
  logging/           Logging (levels, rotation)
  crypto/            X25519, Ed25519, ChaCha20-Poly1305, HKDF
  storage/           SQLite database layer
  network/           libp2p host, streams, mDNS
  protocol/          Message envelopes, encryption
  discovery/         Peer discovery (bootstrap, DHT)
  peer/              Peer management
  organization/      Organization CRUD
  channel/           Channel CRUD
  tui/               Bubble Tea terminal UI
  sync/              State synchronization
  file/              File transfer
  notification/      Desktop/bell notifications
  commands/          Command registry
```

## Storage

Local SQLite database (`chat.db`) with the following tables:

- `identities` — Your Ed25519 keypair and profile
- `peers` — Known peers, keys, trust status
- `organizations` — Org metadata and settings
- `channels` — Channel config and metadata
- `channel_members` — Private channel membership
- `memberships` — Peer-org role mappings
- `messages` — Chat history (encrypted content)
- `attachments` — File transfer metadata
- `invites` — Pending invitations
- `reactions` — Emoji reactions on messages
- `connections` — Saved `/connect` targets for quick reconnection
- `settings` — Key-value settings store
- `sessions` — Cryptographic session keys

## Security

- **Transport** — Encrypted via libp2p's Noise protocol
- **Messages** — Encrypted with ChaCha20-Poly1305 using session keys derived from X25519 + HKDF
- **Identity** — Ed25519 keypairs, verified via fingerprints
- **Forward secrecy** — Unique session keys with rotation
- **Replay protection** — Timestamps and nonce verification
- **No telemetry, no analytics, no cloud**

## Development

### Project Structure

```
cmd/chat/main.go          Entrypoint
internal/app/             Core application logic
internal/tui/             Bubble Tea model, styles, panels
internal/network/         libp2p host setup, stream handler
internal/crypto/          Cryptographic primitives
internal/storage/         SQLite schema, queries, migrations
internal/protocol/        Message encoding and encryption
internal/config/          Config parsing (YAML, JSON)
internal/logging/         Structured logging
internal/discovery/       Peer discovery (mDNS, DHT, bootstrap)
internal/organization/    Organization management
internal/channel/         Channel management
internal/peermgr/         Peer management
internal/sync/            State synchronization
internal/file/            File transfer
internal/notification/    Notifications
internal/commands/        Command registry
```

### Build

```bash
go build -o chat ./cmd/chat
```

### Cross-compile (all targets)

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/chat-linux-amd64 ./cmd/chat
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/chat-linux-arm64 ./cmd/chat

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/chat-darwin-amd64 ./cmd/chat
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/chat-darwin-arm64 ./cmd/chat

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/chat-windows-amd64.exe ./cmd/chat
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o build/chat-windows-arm64.exe ./cmd/chat
```

### Download Pre-built Binaries

Pre-built binaries for all platforms are available from the [Releases](https://github.com/Fenomen-Alex/go-chat/releases) page:

| Platform | Architecture | Binary |
|----------|-------------|--------|
| Linux | amd64 | `chat-linux-amd64` |
| Linux | arm64 | `chat-linux-arm64` |
| macOS | Intel | `chat-darwin-amd64` |
| macOS | Apple Silicon | `chat-darwin-arm64` |
| Windows | amd64 | `chat-windows-amd64.exe` |
| Windows | arm64 | `chat-windows-arm64.exe` |

Each release includes SHA-256 checksums for integrity verification.

## Automated Releases

Every version tag (`v*`) pushed to GitHub triggers an automated release via GitHub Actions:

1. **Lint** — `go vet` and formatting checks
2. **Test** — runs unit tests
3. **Cross-compile** — builds for all 6 targets (linux amd64/arm64, darwin amd64/arm64, windows amd64/arm64)
4. **Checksums** — generates SHA-256 checksums
5. **Archive** — creates `.tar.gz` (Linux/macOS) and `.zip` (Windows) archives
6. **Release** — publishes to GitHub Releases with binaries and checksums

### Create a Release

```bash
# Tag the release (force-push to override)
git tag -f v0.1.0
git push origin v0.1.0 -f
```

The workflow at `.github/workflows/release.yml` handles the rest.

## Milestones

| # | Area | Status |
|---|------|--------|
| 1 | Foundation — project structure, config, identity, SQLite, libp2p, TUI | Done |
| 2 | Chat — DMs, private channels, invites, sync, history | Done |
| 3 | Organizations — membership, roles, permissions, moderation | In Progress |
| 4 | Synchronization — multi-peer state, offline queue, conflict resolution | Planned |
| 5 | Advanced — threads, bookmarks, themes, command palette, encrypted backups | Planned |
| 6 | Polish — performance, security audit, cross-platform validation, releases | Planned |

## License

MIT
