# TON Wallet Address Finder

generate custom TON wallet addresses (V4R2 and V5R2) that end with a specific suffix.
### ⭐ Support the Project by giving a satr!

If you find this project helpful or interesting, please consider giving it a star! Your support is much appreciated.

## Features
- **Multi-threaded**: Utilizes multiple CPU cores to generate wallets in parallel.
- **Custom suffix**: Check if wallet addresses end with a specific string (case-sensitive or case-insensitive).
- **Supports Mainnet/Testnet**: Select the network where the wallets are generated.
- **Bounceable/non-bounceable**: Option to generate bounceable or non-bounceable addresses.
- **Real-time logging**: Logs the number of addresses processed every second.

## Installation

1. Ensure you have [Go installed](https://go.dev/doc/install).
2. Clone the repository and navigate to the project directory.
    ```bash
    git clone https://github.com/ariadata/tongen.git
    cd tongen
    ```

3. Build the project using the following command:

   ```bash
   # Linux
   CGO_ENABLED=0 go build -o tongen main.go

   # Windows
    go build -o tongen.exe main.go
    ```
You should now have an executable named tongen in your project directory.
### Download Pre-built Binaries [Click Here](https://github.com/ariadata/tongen/releases)
```bash
    # Linux 
    curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-linux -o tongen && chmod +x tongen

    # Windows
    curl -sSfL https://github.com/ariadata/tongen/releases/latest/download/tongen-windows.exe -o tongen.exe
```
### Usage

> `-suffix` (required): The desired suffix that the wallet address should end with.

> `-case-sensitive` (optional): Enable case-sensitive suffix matching. Defaults to false.

> `-bounce` (optional): Enable bounceable addresses. Defaults to false.

> `-threads` (optional): Number of parallel threads. Defaults to 0 (use all CPU cores).

> `-testnet` (optional): Use the testnet instead of the mainnet. Defaults to false.

> `-version` (optional): Wallet version 4 or 5 (V4R2 or V5R2). Defaults to 5 (V5R2).

## Examples:
```bash
# Generate a wallet-v4 non-bouncable address that ends with "_Neo" (case-sensitive) using all CPU cores on the mainnet
./tongen -suffix="_Xx" -case-sensitive=true -bounce=false -threads=0 -testnet=false -version=4

# Generate a wallet-v5 bouncable address that ends with "_Test" (not case-insensitive) using 4 threads on testnet 
./tongen -suffix="_Test" -case-sensitive=false -bounce=true -threads=4 -testnet=false -version=5

```

### Example Output:
```bash
2024/10/01 20:00:01 Using 8 threads
2024/10/01 20:00:02 Processed 65 addresses in the last second
2024/10/01 20:00:03 Processed 68 addresses in the last second
=== FOUND ===
Seed phrase: "apple banana cherry date elephant ..."
Wallet address: UQDxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

-----

## Contributing
Feel free to submit issues, fork the repository, and make contributions. Pull requests are welcome!

License
This project is licensed under the MIT License.
