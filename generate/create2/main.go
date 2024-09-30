package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

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

func worker(ctx context.Context, deployer common.Address, initCodeHash []byte, results chan<- common.Address, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			salt, err := generateRandomSalt()
			if err != nil {
				log.Printf("Error generating salt: %v\n", err)
				continue
			}

			contractAddress := create2Address(deployer, salt, initCodeHash)

			if strings.HasPrefix(contractAddress.Hex()[2:], leadingZeros) {
				log.Printf("Found a match with salt: %x\n", salt)
				results <- contractAddress
				return
			}
		}
	}
}

func main() {
	deployerAddress := common.HexToAddress("0xcfeA57885743b5C71Da9B1BaA94F21572A6abccb")
	initCodeHash := HexToBytes("0x8774a50bcdbcd9f23899eaddc829f407273965470f504c5c85f0e51116802760")
	
	numWorkers := 20

	results := make(chan common.Address)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, deployerAddress, initCodeHash, results, &wg)
	}

	fmt.Println("Searching for contract addresses with leading zeros...")

	for i := 0; i < 3; i++ {
		result := <-results
		fmt.Printf("Contract address with leading zeros: %s\n", result.Hex())
	}

	cancel();

	wg.Wait()
	close(results)
}