package main

import (
	"flag"
	"fmt"
	"crypto/sha256"
	"strconv"
	"os"
	"os/signal"
	"math/bits"
	"hash"
)

var DEBUG bool
//
//func generalised_birthday(digest, n, k int) []int {
//	collisoin_length := n/(k+1)
//	if DEBUG {}
//}


func hash_nonce(digest hash.Hash, nonce int) hash.Hash {
	return digest
}

func gbp_basic(digest hash.Hash, n, k int) []int {
	return []int{}
}

func mine(n, k, d *int) {
	// n, k, d are pointers to n, k, d
	// Could be confusing to name the variable and the pointer
	// the same but whatever...
	fmt.Printf("Miner starting\n")
	shaHash := sha256.New()
	initialHash := shaHash.Sum(nil)
	fmt.Println(strconv.Itoa(*n))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for  {
		select{
			case <- c: {
				fmt.Println("Keyboard Interrupt!")
				return
			}
			default:{
				digest := sha256.New()
				digest.Write(initialHash)
				//nonce := 0
				}
			}
		}
	}
}

func main() {
	n := flag.Int("n", 64,
		"number of bits for each number")
	k := flag.Int("k", 7,
		"number of hashes XORed to zero")
	d := flag.Int("d", 0,
		"number of leading zeros")
	flag.Parse()
	mine(n, k, d)
}
