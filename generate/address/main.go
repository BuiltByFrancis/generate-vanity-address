package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

const leadingZeros = "00000000"

func worker(ctx context.Context, results chan<- *ecdsa.PrivateKey, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				log.Printf("Error generating private key: %v\n", err)
				continue
			}

			address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()[2:]

			if strings.HasPrefix(address, leadingZeros) {
				results <- privateKey
				return
			}
		}
	}
}

func main() {
	numWorkers := 20

	results := make(chan *ecdsa.PrivateKey)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, results, &wg)
	}

	fmt.Printf("Searching for Ethereum addresses which starts with 0x%s...\n", leadingZeros)

	for i := 0; i < 3; i++ {
		privateKey := <-results
		
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		privateKeyHex := fmt.Sprintf("%x", crypto.FromECDSA(privateKey))
		fmt.Printf("Found private key: %s\n", privateKeyHex)
		fmt.Printf("Corresponding address: %s\n", address.Hex())
	}

	cancel()

	wg.Wait()
	close(results)
}
