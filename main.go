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

	"github.com/xssnick/tonutils-go/ton/wallet"
)

// Input parameters
type Config struct {
	Suffix        string
	CaseSensitive bool
	Bounce        bool
	Threads       int
	Testnet       bool
}

func main() {
	// Parse input parameters
	config := parseFlags()

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
	suffix := flag.String("suffix", "", "Desired contract address suffix (required)")
	caseSensitive := flag.Bool("case-sensitive", false, "Enable case-sensitive suffix matching (default: false)")
	bounce := flag.Bool("bounce", false, "Enable bounceable address (default: false)")
	threads := flag.Int("threads", 0, "Number of parallel threads (default: 0, meaning use all CPU cores)")
	testnet := flag.Bool("testnet", false, "Use testnet (default: false)")
	flag.Parse()

	if *suffix == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	return Config{
		Suffix:        *suffix,
		CaseSensitive: *caseSensitive,
		Bounce:        *bounce,
		Threads:       *threads,
		Testnet:       *testnet,
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

			// Create a V5R1Final wallet using the seed
			w, err := wallet.FromSeed(nil, seed, wallet.ConfigV5R1Final{
				NetworkGlobalID: getNetworkID(config.Testnet),
				Workchain:       0, // Base workchain
			})
			if err != nil {
				log.Printf("Failed to create wallet: %v", err)
				continue
			}

			// Get the wallet address
			addr := w.WalletAddress()
			// Get the address string (mainnet or testnet) and check bounceable flag
			addressStr := addr.Testnet(config.Testnet).Bounce(config.Bounce).String()

			// Case-sensitive or case-insensitive suffix comparison
			if config.CaseSensitive {
				if strings.HasSuffix(addressStr, config.Suffix) {
					printFoundWallet(seed, addressStr)
					once.Do(func() { close(stopChan) })
					return
				}
			} else {
				if strings.HasSuffix(strings.ToLower(addressStr), strings.ToLower(config.Suffix)) {
					printFoundWallet(seed, addressStr)
					once.Do(func() { close(stopChan) })
					return
				}
			}

			// Increment the counter
			atomic.AddUint64(counter, 1)
		}
	}
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

// getNetworkID returns the correct network ID for mainnet or testnet
func getNetworkID(isTestnet bool) int32 {
	if isTestnet {
		return -3 // Testnet Global ID
	}
	return -239 // Mainnet Global ID
}

// printFoundWallet prints the found seed and wallet address
func printFoundWallet(seed []string, address string) {
	fmt.Println("=== FOUND ===")
	fmt.Println("Seed phrase:", strings.Join(seed, " "))
	fmt.Println("Wallet address:", address)
}