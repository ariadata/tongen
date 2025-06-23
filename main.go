package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

// Input parameters
type Config struct {
	Version       int
	Suffix        string
	CaseSensitive bool
	Bounce        bool
	Threads       int
	Testnet       bool
	Output        string
	Daemon        bool
}

func main() {
	// Check for stop command first
	if len(os.Args) > 1 && (os.Args[1] == "stop" || os.Args[1] == "--stop") {
		stopDaemon()
		return
	}

	// Check for daemon flag before parsing all flags
	isDaemon := false
	for _, arg := range os.Args {
		if arg == "-d" || arg == "--daemon" {
			isDaemon = true
			break
		}
	}

	if isDaemon {
		// Initialize daemon context
		daemonContext := &daemon.Context{
			PidFileName: "/tmp/tongen.pid",
			PidFilePerm: 0644,
			LogFileName: "/tmp/tongen.log",
			LogFilePerm: 0640,
			WorkDir:     "./",
			Umask:       027,
			Args:        os.Args,
		}

		// Check if daemon is already running
		child, err := daemonContext.Reborn()
		if err != nil {
			log.Fatalf("Daemon error: %v", err)
		}

		if child != nil {
			// Parent process - exit
			log.Println("Daemon started successfully")
			log.Printf("PID file: %s", daemonContext.PidFileName)
			log.Printf("Log file: %s", daemonContext.LogFileName)
			return
		}

		// Child process - parse flags and run
		defer func() {
			if err := daemonContext.Release(); err != nil {
				log.Printf("Unable to release pid-file: %s", err.Error())
			}
		}()

		log.Println("Daemon started, running wallet generator...")
		config := parseFlags()
		runWalletGenerator(config)
		return
	}

	// Normal mode - parse flags and run
	config := parseFlags()
	runWalletGenerator(config)
}

// runWalletGenerator contains the main wallet generation logic
func runWalletGenerator(config Config) {
	// Determine the number of threads (default: use all CPU cores if threads=0)
	if config.Threads == 0 {
		config.Threads = runtime.NumCPU()
	}
	log.Printf("Using %d threads\n", config.Threads)

	// Channel to signal when a match is found
	stopChan := make(chan struct{})

	// Use sync.Once to ensure stopChan is closed only once
	var once sync.Once

	// Start tracking the number of processed wallets
	var counter uint64
	var wg sync.WaitGroup

	// Start logging progress every second
	go logProgress(&counter, stopChan)

	// Start wallet generation and processing
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processWallets(config, &counter, stopChan, &once)
		}()
	}

	// Wait for all threads to finish
	wg.Wait()
}

// parseFlags handles command-line input parameters
func parseFlags() Config {
	// Create a new flag set to avoid conflicts
	fs := flag.NewFlagSet("tongen", flag.ExitOnError)

	version := fs.Int("version", 5, "Wallet version (4 or 5, default: 5)")
	suffix := fs.String("suffix", "", "Desired contract address suffix (required)")
	caseSensitive := fs.Bool("case-sensitive", false, "Enable case-sensitive suffix matching (default: false)")
	bounce := fs.Bool("bounce", false, "Enable bounceable address (default: false)")
	threads := fs.Int("threads", 0, "Number of parallel threads (default: 0, meaning use all CPU cores)")
	testnet := fs.Bool("testnet", false, "Use testnet (default: false)")
	output := fs.String("output", "", "Output file path to save results (use -o or --output)")
	fs.StringVar(output, "o", "", "Output file path to save results (short form)")

	// Filter out daemon flags from arguments
	var filteredArgs []string
	for _, arg := range os.Args {
		if arg == "-d" || arg == "--daemon" {
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}

	// Parse the filtered arguments
	fs.Parse(filteredArgs[1:]) // Skip the program name

	if *suffix == "" || (*version != 4 && *version != 5) {
		fs.PrintDefaults()
		os.Exit(1)
	}

	return Config{
		Version:       *version,
		Suffix:        *suffix,
		CaseSensitive: *caseSensitive,
		Bounce:        *bounce,
		Threads:       *threads,
		Testnet:       *testnet,
		Output:        *output,
		Daemon:        false, // Daemon flag is handled in main function
	}
}

// processWallets generates wallets, checks if the address matches the suffix, and stops on a match
func processWallets(config Config, counter *uint64, stopChan chan struct{}, once *sync.Once) {
	for {
		select {
		case <-stopChan:
			return
		default:
			// Generate the seed phrase
			seed := wallet.NewSeed()

			// Create a wallet based on the selected version (V4 or V5)
			var addressStr string
			var err error

			if config.Version == 5 {
				addressStr, err = generateV5Wallet(seed, config)
			} else {
				addressStr, err = generateV4Wallet(seed, config)
			}

			if err != nil {
				log.Printf("Failed to create wallet: %v", err)
				continue
			}

			// Case-sensitive or case-insensitive suffix comparison
			if config.CaseSensitive {
				if strings.HasSuffix(addressStr, config.Suffix) {
					printFoundWallet(seed, addressStr, config.Output)
					once.Do(func() { close(stopChan) })
					return
				}
			} else {
				if strings.HasSuffix(strings.ToLower(addressStr), strings.ToLower(config.Suffix)) {
					printFoundWallet(seed, addressStr, config.Output)
					once.Do(func() { close(stopChan) })
					return
				}
			}

			// Increment the counter
			atomic.AddUint64(counter, 1)
		}
	}
}

// generateV5Wallet creates a V5 wallet and returns the corresponding address
func generateV5Wallet(seed []string, config Config) (string, error) {
	// Create a V5R1Final wallet using the seed
	w, err := wallet.FromSeed(nil, seed, wallet.ConfigV5R1Final{
		NetworkGlobalID: getNetworkID(config.Testnet),
		Workchain:       0, // Base workchain
	})
	if err != nil {
		return "", err
	}

	// Get the wallet address
	addr := w.WalletAddress()
	addressStr := addr.Testnet(config.Testnet).Bounce(config.Bounce).String()
	return addressStr, nil
}

// generateV4Wallet creates a V4 wallet and returns the corresponding address
func generateV4Wallet(seed []string, config Config) (string, error) {
	// Create a V4R2 wallet using the seed
	w, err := wallet.FromSeed(nil, seed, wallet.V4R2)
	if err != nil {
		return "", err
	}

	// Get the wallet address
	addr := w.WalletAddress()
	addressStr := addr.Testnet(config.Testnet).Bounce(config.Bounce).String()
	return addressStr, nil
}

// logProgress logs how many wallets were processed in the last second
func logProgress(counter *uint64, stopChan chan struct{}) {
	var lastCount uint64
	for {
		select {
		case <-stopChan:
			return
		case <-time.After(1 * time.Second):
			currentCount := atomic.LoadUint64(counter)
			processedLastSecond := currentCount - lastCount
			lastCount = currentCount
			log.Printf("Processed %d addresses in the last second\n", processedLastSecond)
		}
	}
}

// getNetworkID returns the correct network ID for mainnet or testnet (only for V5)
func getNetworkID(isTestnet bool) int32 {
	if isTestnet {
		return -3 // Testnet Global ID
	}
	return -239 // Mainnet Global ID
}

// printFoundWallet prints the found seed and wallet address
func printFoundWallet(seed []string, address string, output string) {
	fmt.Println("=== FOUND ===")
	fmt.Println("Seed phrase:", strings.Join(seed, " "))
	fmt.Println("Wallet address:", address)

	if output != "" {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		content := fmt.Sprintf("=== FOUND %s ===\nSeed: %s\nAddress: %s\n\n", timestamp, strings.Join(seed, " "), address)
		err := writeToFile(output, content)
		if err != nil {
			log.Printf("Failed to write to file: %v", err)
		} else {
			fmt.Printf("Results saved to: %s\n", output)
		}
	}
}

func writeToFile(filename, content string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// stopDaemon stops the running daemon
func stopDaemon() {
	daemonContext := &daemon.Context{
		PidFileName: "/tmp/tongen.pid",
	}

	// Send signal to stop the daemon
	d, err := daemonContext.Search()
	if err != nil {
		log.Fatalf("Unable to search daemon: %s", err.Error())
	}

	if d == nil {
		log.Println("Daemon is not running")
		return
	}

	err = daemonContext.Release()
	if err != nil {
		log.Fatalf("Unable to release daemon: %s", err.Error())
	}

	log.Println("Daemon stopped successfully")
}
