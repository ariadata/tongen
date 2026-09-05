# TON Wallet Address Finder

Generate custom TON wallet addresses (V4R2 and V5R2) that end with a specific suffix.

### Support the Project by giving a star!

If you find this project helpful or interesting, please consider giving it a star! Your support is much appreciated.

## Features
- **Multi-threaded**: Utilizes multiple CPU cores to generate wallets in parallel.
- **Custom suffix**: Check if wallet addresses end with a specific string (case-sensitive or case-insensitive).
- **Supports Mainnet/Testnet**: Select the network where the wallets are generated.
- **Bounceable/non-bounceable**: Option to generate bounceable or non-bounceable addresses.
- **Real-time logging**: Logs the number of addresses processed every second.
- **Daemon mode**: Run as a background daemon process.
- **Output file**: Save found wallets to a file for later use.
- **Telegram notify** (optional): After saving to file, send the found result via a Telegram bot.

## Quick Start
### Linux
```bash
# Download and make executable (amd64; use tongen-linux-arm64 on ARM)
curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-linux-amd64 -o tongen && chmod +x tongen

# Generate V5R2 address ending with "_Cool" (case-sensitive)
./tongen -suffix="_Cool" -case-sensitive=true -version=5
```

### Windows
```powershell
# Download (PowerShell) — amd64; use tongen-windows-arm64.exe on ARM
Invoke-WebRequest -Uri "https://github.com/ariadata/tongen/releases/latest/download/tongen-windows-amd64.exe" -OutFile tongen.exe

# Or with curl:
# curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-windows-amd64.exe -o tongen.exe

# Generate V5R2 address ending with "_Cool" (case-sensitive)
.\tongen.exe -suffix="_Cool" -case-sensitive=true -version=5
```

## Installation

1. Ensure you have [Go 1.27+](https://go.dev/doc/install) installed.
2. Clone the repository and navigate to the project directory:
   ```bash
   git clone https://github.com/ariadata/tongen.git
   cd tongen
   ```

3. Build the project:

   ```bash
   # Linux / macOS / WSL (native binary)
   CGO_ENABLED=0 go build -o tongen .

   # Windows (native build on Windows)
   go build -o tongen.exe .

   # Cross-compile Windows .exe from Linux / WSL / macOS (amd64 or arm64)
   CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o tongen.exe .
   CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o tongen-arm64.exe .
   ```

   > **Important:** Building with `go build -o tongen.exe` on Linux/WSL produces a **Linux** binary that only has a `.exe` name. That file will not run on Windows (you may see "Unsupported 16-Bit Application"). Always set `GOOS=windows` and the correct `GOARCH` when cross-compiling.

### Download Pre-built Binaries

[Releases](https://github.com/ariadata/tongen/releases)

Release assets:

| Platform | Asset |
|----------|--------|
| Linux (amd64) | `tongen-linux-amd64` |
| Linux (arm64) | `tongen-linux-arm64` |
| Windows (amd64) | `tongen-windows-amd64.exe` |
| Windows (arm64) | `tongen-windows-arm64.exe` |

```bash
# Linux amd64
curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-linux-amd64 -o tongen && chmod +x tongen

# Linux arm64
curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-linux-arm64 -o tongen && chmod +x tongen

# Windows amd64
curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-windows-amd64.exe -o tongen.exe

# Windows arm64
curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-windows-arm64.exe -o tongen.exe
```

## Usage

| Flag | Required | Description |
|------|----------|-------------|
| `-suffix` | yes | Desired suffix that the wallet address should end with |
| `-case-sensitive` | no | Case-sensitive suffix matching (default: `false`) |
| `-bounce` | no | Bounceable addresses (default: `false`) |
| `-threads` | no | Parallel threads; `0` = all CPU cores (default: `0`) |
| `-testnet` | no | Use testnet instead of mainnet (default: `false`) |
| `-version` | no | Wallet version `4` or `5` (V4R2 / V5R2, default: `5`) |
| `-o`, `--output` | no | Output file path to save found wallets |
| `-tg-token` | no | Telegram bot token (optional; with `-tg-chat-id`) |
| `-tg-chat-id` | no | Telegram chat ID (optional; with `-tg-token`) |
| `-d`, `--daemon` | no | Run as a background daemon process |
| `stop`, `--stop` | — | Stop the running daemon process |

## Examples

```bash
# V4 non-bounceable address ending with "_Xx" (case-sensitive), all CPU cores, mainnet
./tongen -suffix="_Xx" -case-sensitive=true -bounce=false -threads=0 -testnet=false -version=4

# V5 bounceable address ending with "_Test" (case-insensitive), 4 threads, mainnet
./tongen -suffix="_Test" -case-sensitive=false -bounce=true -threads=4 -testnet=false -version=5

# Save found wallet to a file
./tongen -suffix="_Cool" -o results.txt

# Notify Telegram when found (optional; if -o is set, file is written before Telegram)
./tongen -suffix="_Cool" -o results.txt -tg-token="123:ABC" -tg-chat-id="123456789"
```

### Example Output

```text
2024/10/01 20:00:01 Using 8 threads
2024/10/01 20:00:02 Processed 65 addresses in the last second
2024/10/01 20:00:03 Processed 68 addresses in the last second
=== FOUND ===
Seed phrase: "apple banana cherry date elephant ..."
Wallet address: UQDxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

> **Daemon mode (`-d`)**
>
> Built-in background mode (writes PID/log under `/tmp`):
>
> ```bash
> # Start
> ./tongen -suffix="_Cool" -d -o results.txt
>
> # Stop
> ./tongen stop
> ```
>
> PID file: `/tmp/tongen.pid` · Log file: `/tmp/tongen.log`
-----
## Optional: run in background with `nohup`

If you prefer not to use `-d`, you can detach the process with `nohup`. Discard progress logs to `/dev/null` so a multi-day run does not grow a huge log file; the found result still goes to `-o` when you set it.

```bash
# Start (save PID for later stop)
nohup ./tongen -suffix="_Cool" -case-sensitive=true -o="tongen-out.txt" > /dev/null 2>&1 & echo $! > tongen.pid

# Stop
kill "$(cat tongen.pid)"
```

> Run `echo $! > tongen.pid` in the **same shell session**, immediately after starting with `&`. Otherwise `$!` is empty and stop will fail.
-----


## Contributing

Feel free to submit issues, fork the repository, and make contributions. Pull requests are welcome!

## License

This project is licensed under the MIT License.


