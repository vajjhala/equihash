package main

import (
	"flag"
	"fmt"
	"crypto/sha256"
	"strconv"
	"os"
	"os/signal"
	"hash"
	"encoding/binary"
	"sort"
	"bytes"
	"math/bits"
	"time"
	"errors"
)

type HashPairs struct {
	hashSum   []byte
	inputSeeds []int
}

type hArrays []HashPairs

func hash_nonce(digest hash.Hash, nonce int) hash.Hash {
	for i := 0; i < 8; i++ {
		buff := make([]byte, 8)
		uinteger := uint64(nonce >> (32 * 1))
		binary.LittleEndian.PutUint64(buff, uinteger)
		digest.Write(buff)
	}
	return digest
}

func hash_xi(digest hash.Hash, xi int) hash.Hash {
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(xi))
	digest.Write(buff)
	return digest
}

func has_collision(hI, hJ []byte, i, l int) bool {
	start := (i - 1) * l / 8
	end := i * i * l / 8
	return bytes.Equal(hI[start:end], hJ[start:end])
}

func distinctIndices(a, b []int) bool {
	for _, v := range(a){
		for _, w :=range (b){
			if v == w {
				return false
			}
		}
	}
	return true
}

func countZeros(a []byte) int {
	x := binary.LittleEndian.Uint64(a[0:])
	return bits.LeadingZeros64(x)
}

func gbp_basic(digest hash.Hash, n, k int) [][]int {
	collisionLength := n / (k + 1)
	X := hArrays{}
	fmt.Printf("Generating a list\n")
	// Generating a list (ordered) but needs to be changed to random
	for i := 0; i < int(Power(uint(2), uint(collisionLength+1))); i++ {
		// The value of digest is passed and a new value
		// is sent and stored in curr_digest,
		// original digest value does not change
		curr_digest := hash_xi(digest, i)
		pair := HashPairs{curr_digest.Sum(nil), []int{i}}
		X = append(X, pair)
	}
	for i := 1; i < k; i++ {
		fmt.Printf("Round : %d\n", i)
		sort.Sort(X)
		fmt.Printf("Sorting the list\n")
		for _, Xi := range X[len(X)-32:] {
			fmt.Printf("H(%v): %v", Xi.inputSeeds, Xi.hashSum)
		}
		Xc := hArrays{}
		for len(X) > 0 {
			j := 1
			for j < len(X) {
				if ! has_collision(X[len(X)-1].hashSum,
					X[len(X)-1-j].hashSum, i, collisionLength) {
					break
				}
				j++
			}
			for l := 0; l < j-1; l++ {
				for m := l + 1; m < j; m++ {
					if distinctIndices(X[len(X)-1-l].inputSeeds,
						X[len(X)-1-m].inputSeeds){
						var concat []int
						if bytes.Compare(X[len(X)-1-l].hashSum,
							X[len(X)-1-m].hashSum) == -1 {
							concat = append(X[len(X)-1-l].inputSeeds,
								X[len(X)-1-m].inputSeeds...)
						}else {
							concat = append(X[len(X)-1-m].inputSeeds,
								X[len(X)-1-l].inputSeeds ...)
						}
						xored := SafeXORBytes(X[len(X)-1-l].hashSum,
							X[len(X)-1-m].hashSum)
						Xc = append(Xc, HashPairs{xored,
						concat})
					}
				}
			}

			for j > 0{
				X = X[:len(X)-1]
				j = j - 1
			}
		}
		X = Xc
	}
	fmt.Printf("Final Round\n")
	fmt.Printf("Sorting List\n")
	sort.Sort(X)
	for _, Xi := range X[len(X)-32:] {
		fmt.Printf("H(%v): %v", Xi.inputSeeds, Xi.hashSum)
	}
	fmt.Printf("Finding Collisions\n")
	sols := [][]int{{}}
	for i := 0; i < len(X)-1; i++{
		res := SafeXORBytes(X[i].hashSum, X[i+1].hashSum)
		if countZeros(res) == n && (distinctIndices(X[i].inputSeeds,
			X[i+1].inputSeeds)){
			fmt.Printf("%#v\n", X[i])
			fmt.Printf("%#v\n", X[i+1])
		}
		if X[i].hashSum[0] < X[i].hashSum[1]{
			s := [][]int{append(X[i].inputSeeds, X[i+1].inputSeeds ...)}
			sols = append(sols, s...)
		}else{
			s := [][]int{append(X[i+1].inputSeeds, X[i].inputSeeds...)}
			sols = append(sols, s...)
		}
	}
	return sols
}

func blockHash(prevHash []byte, nonce int, soln []int) []byte {
	digest := sha256.New()
	digest.Write(prevHash)
	digest = hash_nonce(digest, nonce)
	for _, v := range(soln) {
		digest = hash_xi(digest, v)
		}
	h := digest.Sum(nil)
	digest.Reset()
	digest.Write(h)
	return digest.Sum(nil)
}

func difficultyFilter(prevHash []byte, nonce int,
	soln []int, d int) bool {
	h := blockHash(prevHash, nonce, soln)
	count := countZeros(h)
	return (count >= d)
}

func validateParams(n, k int) (bool, error){
	if k >= n{
		return false, errors.New("n must be larger than k")
	}
	if n/(k+1) % 8 != 0{
		return false, errors.New("Parameters must satisfy n/(k+1) = 0 mod 8")
	}
	return true, nil
}

func mine(n, k, d *int) {
	// n, k, d are pointers to n, k, d
	// Could be confusing to name the variable and the pointer
	// the same but whatever...
	_, err := validateParams(*n, *k)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Miner starting\n")
	shaHash := sha256.New()
	prevHash := shaHash.Sum(nil)
	// prevHash = []byte
	nonce := 0
	x := []int{}
	fmt.Println(strconv.Itoa(*n))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	var elapsed time.Duration
	for {
		select {
		case <-c:
			{
				fmt.Println("Keyboard Interrupt!")
				return
			}
		default:
			{
				digest := sha256.New()
				digest.Write(prevHash)
				start := time.Now()
				for (nonce >> 161) == 0{
					digest = hash_nonce(digest, nonce)
					solns := gbp_basic(digest, *n, *k)
					for _, soln := range(solns){
						if difficultyFilter(prevHash, nonce, soln, *d){
							x = soln
							break
							}
						}
					if len(x) != 0 {break}
					nonce = nonce + 1
					digest.Reset()
					}
				t := time.Now()
				elapsed = t.Sub(start)
				if len(x) == 0 {
					fmt.Printf("Could not find any valid nonce")
					return
				}
				currHash := blockHash(prevHash, nonce, x)
				fmt.Printf("-------------------------\n")
				fmt.Printf("Mined Block\n")
				fmt.Printf("Nonce: %#v\n", nonce)
				fmt.Printf("Previous Hash %#v\n", prevHash)
				fmt.Printf("Current Hash %#v\n", currHash)
				fmt.Printf("Time to find nonce %#v\n", elapsed)
				prevHash = currHash
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
