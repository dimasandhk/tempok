<div align="center">
  <h1>⏳ Tempok</h1>
  <p><strong>A self-hosted, lightweight reverse proxy & tunneling tool with a time-to-live (TTL) twist.</strong></p>
</div>

---

## 📖 What is Tempok?

Tempok (Tempo + ngrok) is an open-source, self-hosted tunneling tool written in Go. It allows you to expose local environments to the public internet securely using a single remote Virtual Machine, just like ngrok.

**The Twist:** Unlike other tunneling tools, Tempok introduces strict **Time-to-Live (TTL) functionality**. Tunnels are time-bound by default, self-destructing when time runs out. This prevents stale, forgotten tunnels from becoming security vulnerabilities.

## ✨ Features

- 🛠 **Single Binary:** A "Swiss Army Knife" design. One binary acts as both the VM Server and the Local Client.
- ⏱ **Time-to-Live (TTL):** Tunnels expire automatically. You can `extend` them or `persist` them forever.
- 🔒 **Secure by Default:** Requires a `--secret` handshake to establish connections. Prevents your server from becoming an open proxy.
- 🚀 **Multiplexed:** Uses HashiCorp's Yamux to run thousands of logical HTTP streams over a single long-lived TCP connection.
- 🐳 **Docker Ready:** Tiny, scratch-based container images for ultra-fast deployment.
- 💻 **Cross-Platform:** Binaries available for Windows, macOS, and Linux.

---

## 📦 Installation

### Option 1: Download Pre-compiled Binary (Easiest)
Grab the latest release for your OS from the [Releases Tab](../../releases/latest). 
Extract the archive and place the `tempok` executable in your `$PATH`.

### Option 2: Run via Docker (For the Server)
Tempok provides a minimal Docker image, perfect for running the server on a VPS.
```bash
docker pull ghcr.io/dimasandhk/tempok:latest
docker run -d -p 9999:9999 -p 80:80 --name tempok-server ghcr.io/dimasandhk/tempok:latest server --secret "your-strong-secret" --control-port 9999 --public-port 80
```

### Option 3: Build from Source
```bash
git clone https://github.com/dimasandhk/tempok.git
cd tempok
go build -o tempok .
```

---

## 🚀 Usage

### 1. Start the Server (On your VPS/VM)
The server listens for incoming Tempok clients on the control port and internet traffic on the public port.
```bash
tempok server --secret "my-super-secret" --control-port 9999 --public-port 80
```

### 2. Start a Tunnel (On your Local Machine)
Expose your local `localhost:8080` to the internet through the Tempok server. By default, tunnels expire after 1 hour.
```bash
tempok tunnel 8080 --server my-vps-ip:9999 --secret "my-super-secret" --expires 30m
```

### 3. Manage Your Tunnels (The API)
Tempok includes a built-in CLI API to manage your active tunnels without restarting connections.

**List all active tunnels:**
```bash
tempok list --secret "my-super-secret" --server my-vps-ip:9999
```

**Extend a tunnel's lifespan:**
```bash
# Add 2 hours to a specific tunnel
tempok extend <tunnel_id> 2h --secret "my-super-secret"
```

**Make a tunnel permanent (Disable TTL):**
```bash
tempok persist <tunnel_id> --secret "my-super-secret"
```

---

## 🏗 Architecture
- **Language:** Go (Golang)
- **CLI Framework:** Cobra (`spf13/cobra`)
- **Networking:** Yamux (`hashicorp/yamux`) for stream multiplexing
- **State Management:** Thread-safe in-memory maps with graceful teardowns via Contexts.

## 🤝 Contributing
Contributions, issues, and feature requests are welcome! 

## 📝 License
This project is open-source and available under the MIT License.
