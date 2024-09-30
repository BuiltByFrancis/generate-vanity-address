package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const numWorkers = 20
const leadingZeros = "00000000"

func HexToBytes(str string) []byte {
	if len(str) >= 2 && str[:2] == "0x" {
		str = str[2:]
	}
	decoded, _ := hex.DecodeString(str)
	return decoded
}

func keccak256(data []byte) []byte {
	hash := crypto.Keccak256Hash(data)
	return hash.Bytes()
}

func create2Address(deployer common.Address, salt [32]byte, initCodeHash []byte) common.Address {
	// Constant 0xff used in CREATE2 calculation
	var prefix = []byte{0xff}

	// Concatenate 0xff + deployer + salt + keccak256(init_code)
	data := append(prefix, deployer.Bytes()...)
	data = append(data, salt[:]...)
	data = append(data, initCodeHash...)

	// Take the keccak256 of the concatenated data
	addressHash := keccak256(data)

	// Return the last 20 bytes (rightmost) as the contract address
	return common.BytesToAddress(addressHash[12:])
}

func generateRandomSalt() ([32]byte, error) {
	var salt [32]byte
	_, err := rand.Read(salt[:])
	return salt, err
}

func worker(wg *sync.WaitGroup, guessChan chan int, deployer common.Address, initCodeHash []byte) {
	defer wg.Done()
	guesses := 0

	for {
		salt, err := generateRandomSalt()
		if err != nil {
			log.Printf("Error generating salt: %v\n", err)
			continue
		}

		contractAddress := create2Address(deployer, salt, initCodeHash)

		if strings.HasPrefix(contractAddress.Hex()[2:], leadingZeros) {	
			log.Printf("Found a match with salt: %x\n", salt)
			return
		}

		guesses++
		if guesses%1000 == 0 {
			guessChan <- guesses
			guesses = 0
		}
	}
}

func main() {
	deployerAddress := common.HexToAddress("0xcfeA57885743b5C71Da9B1BaA94F21572A6abccb")
	initCodeHash := HexToBytes("0xb3c83d1e2dfcf5d534f086eecac706fe2eb58e03eb2101af5f79673ed78293f5")
	
	var wg sync.WaitGroup
	guessChan := make(chan int)
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	totalGuesses := 0
	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, guessChan, deployerAddress, initCodeHash)
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

