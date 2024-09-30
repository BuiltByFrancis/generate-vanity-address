package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

const leadingZeros = "00000000"
const numWorkers = 20

func main() {
	var wg sync.WaitGroup
	guessChan := make(chan int)
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	totalGuesses := 0
	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, guessChan)
	}

	go func() {
		for {
			select {
			case guesses := <-guessChan:
				totalGuesses += guesses
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				averageGuessesPerSecond := float64(totalGuesses) / elapsed
				fmt.Printf("Running Average Guesses per second: %.2f\n", averageGuessesPerSecond)
			}
		}
	}()

	wg.Wait()
}

func worker(wg *sync.WaitGroup, guessChan chan int) {
	defer wg.Done()
	guesses := 0

	for {
		privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			log.Fatal(err)
		}
		
		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()[2:]
		if strings.HasPrefix(address, leadingZeros) {
			fmt.Printf("Found a match! Private key: %x\n", crypto.FromECDSA(privateKey))
		}

		guesses++
		if guesses%1000 == 0 {
			guessChan <- guesses
			guesses = 0
		}
	}
}
