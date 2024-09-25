package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const leadingZeros = "00000000"

func keccak256(data []byte) []byte {
	hash := crypto.Keccak256Hash(data)
	return hash.Bytes()
}

func create2Address(deployer common.Address, salt [32]byte, initCode []byte) common.Address {
	// Constant 0xff used in CREATE2 calculation
	var prefix = []byte{0xff}

	// Concatenate 0xff + deployer + salt + keccak256(init_code)
	data := append(prefix, deployer.Bytes()...)
	data = append(data, salt[:]...)
	data = append(data, keccak256(initCode)...)

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

func worker(ctx context.Context, deployer common.Address, initCode []byte, results chan<- common.Address, wg *sync.WaitGroup) {
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

			contractAddress := create2Address(deployer, salt, initCode)

			if strings.HasPrefix(contractAddress.Hex()[2:], leadingZeros) {
				log.Printf("Found a match with salt: %x\n", salt)
				results <- contractAddress
				return
			}
		}
	}
}

func main() {
	deployerAddress := common.HexToAddress("yourDeployerAddressHere")

	initCode := []byte("YourContractBytecodeHere")

	numWorkers := 20

	results := make(chan common.Address)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, deployerAddress, initCode, results, &wg)
	}

	for i := 0; i < 3; i++ {
		result := <-results
		fmt.Printf("Contract address with leading zeros: %s\n", result.Hex())
	}

	cancel();

	wg.Wait()
	close(results)
}