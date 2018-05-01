package main

import (
	"flag"
	"fmt"
	"crypto/sha256"
	"strconv"
	"hash"
	"encoding/binary"
	"sort"
	"bytes"
	"time"
	"errors"
	"math/bits"
	"bufio"
	"os"
)

type HashPairs struct {
	hashSum    []byte
	inputSeeds []int
}

type hArrays []HashPairs

func hashNonce(digest hash.Hash, nonce int) hash.Hash {
	for i := 0; i < 8; i++ {
		buff := make([]byte, 8)
		// to stare a 8 * byte = 64 bit unsigned integer
		uinteger := uint64(nonce >> uint32(32*i))
		binary.LittleEndian.PutUint64(buff[0:], uinteger)
		digest.Write(buff)
	}
	return digest
}

func hashXi(digest hash.Hash, xi int) hash.Hash {
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(xi))
	digest.Write(buff)
	return digest
}

func hasCollision(hashI, hashJ []byte, i, l int) bool {
	start := ((i - 1) * l) / 8
	end := (i * l) / 8
	blockI := hashI[start:end]
	blockJ := hashJ[start:end]

	//x := binary.LittleEndian.Uint64(blockI[0:])
	//y := binary.LittleEndian.Uint64(blockJ[0:])
	//fmt.Printf("h: %v\n", strconv.FormatUint(x, 2))
	//fmt.Printf("h: %v\n", strconv.FormatUint(y, 2))

	if bytes.Equal(blockI, blockJ) {
		//fmt.Printf("%v and %v collide between %v : %v = %v, %v\n",
		//	hashI, hashJ, start, end, blockI, blockJ)
		return true
	} else {
		return false
	}
}

func distinctIndices(a, b []int) bool {
	hTable := make(map[string]int)
	for _, w := range (b) {
		hTable[strconv.Itoa(w)] = w
	}
	for _, v := range a {
		_, ok := hTable[strconv.Itoa(v)]
		if ok {
			return false
		}
	}
	return true
}

func countZeros(a []byte) int {
	x := binary.LittleEndian.Uint64(a[0:])
	count := bits.TrailingZeros64(x)
	//fmt.Printf("leadingZeros: %v\n", count)
	return count
}

func gbpBasic(digest hash.Hash, n, k int) [][]int {
	collisionLength := n / (k + 1)
	X := hArrays{}
	//fmt.Printf("Generating a list\n")

	// Generating a list (ordered) but needs to be changed to random
	for i := 0; i < int(Power(uint(2), uint(collisionLength+1))); i++ {
		// The value of digest is passed and a new value
		// is sent and stored in curr_digest,
		// original digest value does not change
		currDigest := hashXi(digest, i)
		pair := HashPairs{currDigest.Sum(nil), []int{i}}
		X = append(X, pair)
	}

	for i := 1; i < k; i++ {
		//fmt.Printf("Round : %d\n", i)
		//fmt.Printf("Sorting the list\n")
		sort.Sort(X)

		// initialise a new empty struct
		Xc := hArrays{}

		for len(X) > 0 {
			j := 1
			for j < len(X) {
				// checks if a block of bits collide,
				// if so, only then checks the next block.
				if ! hasCollision(X[len(X)-1].hashSum,
					X[len(X)-1-j].hashSum, i, collisionLength) {
					break
				}
				j++
			}
			// Store( Xi ^ Xj , (i, j)) in a table
			for l := 0; l < j-1; l++ {
				for m := l + 1; m < j; m++ {
					if distinctIndices(X[len(X)-1-l].inputSeeds,
						X[len(X)-1-m].inputSeeds) {
						var concat []int
						if X[len(X)-1-l].inputSeeds[0] <
							X[len(X)-1-m].inputSeeds[0] {
							concat = append(X[len(X)-1-l].inputSeeds,
								X[len(X)-1-m].inputSeeds...)
						} else {
							concat = append(X[len(X)-1-m].inputSeeds,
								X[len(X)-1-l].inputSeeds ...)
						}
						xored := SafeXORBytes(X[len(X)-1-l].hashSum,
							X[len(X)-1-m].hashSum)
						//fmt.Printf("Xored: %v\n", xored)
						Xc = append(Xc, HashPairs{xored,
							concat})
					}
				}
			}

			for j > 0 {
				//fmt.Printf("Length of X: %v\n", len(X))
				X = X[:len(X)-1]
				j = j - 1
			}
		}
		X = Xc
	}
	//fmt.Printf("Final Round\n")
	//fmt.Printf("Sorting List\n")
	sort.Sort(X)
	//fmt.Printf("%v\n", X)

	//for _, Xi := range X[len(X)-32:] {
	//	fmt.Printf("H(%v): %v\n", Xi.inputSeeds, Xi.hashSum)
	//}

	//fmt.Printf("Finding Collisions\n")
	sols := [][]int{}
	for i := 0; i < len(X)-1; i++ {
		res := SafeXORBytes(X[i].hashSum, X[i+1].hashSum)
		// res must be 64 zeros!
		//fmt.Printf("zeros: %v =? n: %v\n", countZeros(res), n)

		if countZeros(res) == n && (distinctIndices(X[i].inputSeeds,
			X[i+1].inputSeeds)) {
			//fmt.Printf("res: %#v\nuint: %v\n", res, binary.LittleEndian.Uint64(res))
			//fmt.Printf("zeros: %v =? n: %v\n", countZeros(res), n)
			//fmt.Printf("res: %x of len %v\n", res, len(res))
			//fmt.Printf("Found a solution\n")
			//fmt.Printf("%x\n", X[i].hashSum)
			//fmt.Printf("%x\n", X[i+1].hashSum)

			if X[i].inputSeeds[0] < X[i+1].inputSeeds[0] {
				//fmt.Printf("Hurray")
				s := [][]int{append(X[i].inputSeeds, X[i+1].inputSeeds ...)}
				sols = append(sols, s...)
			} else {
				//fmt.Printf("na na na")
				s := [][]int{append(X[i+1].inputSeeds, X[i].inputSeeds...)}
				sols = append(sols, s...)
			}
		}
	}
	return sols
}

func checkIfZero(h []byte) bool {
	x := binary.LittleEndian.Uint64(h[0:])
	count := bits.TrailingZeros64(x)
	//fmt.Printf("leadingZeros: %v\n", count)
	return count == 8*10
}

func blockHash(prevHash []byte, nonce int, soln []int, data []byte) []byte {
	digest := sha256.New()
	digest.Write(prevHash)
	digest.Write(data)
	digest = hashNonce(digest, nonce)
	for _, v := range soln {
		digest = hashXi(digest, v)
	}
	return digest.Sum(nil)
}

func difficultyFilter(prevHash []byte, nonce int,
	soln []int, data []byte, d int) bool {
	h := blockHash(prevHash, nonce, soln, data)
	//hBin := strconv.FormatUint(binary.LittleEndian.Uint64(h), 2)
	count := countZeros(h)
	//fmt.Printf("h: %v \nlen: %v \nzeros: %v\n",
	//	hBin, len(hBin), count)
	return count >= d
}

func validateParams(n, k int) (bool, error) {
	if k >= n {
		return false, errors.New("n must be larger than k")
	}
	check := n / (k + 1) % 8
	if check != 0 {
		return false, errors.New("parameters must satisfy n/(k+1) = 0 mod 8")
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
	for {
		nonce := 0
		x := []int{}
		solFlag := false
		var elapsed time.Duration
		shaDigest := sha256.New()
		// digest sha256 as [ byte byte .. x32]
		shaDigest.Write(prevHash)
		start := time.Now()

		// Take user input of data in the block
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Block data: ")
		scanner.Scan()
		inputData := scanner.Text()
		//fmt.Printf("Data in []byte: %v\n", []byte(inputData))

		dataBytes := []byte(inputData)

		for (nonce >> 161) == 0 {

			digest := hashNonce(shaDigest, nonce)
			solns := gbpBasic(digest, *n, *k)
			//fmt.Printf("Solutions %v, len: %v\n", solns, len(solns))
			for i, soln := range solns {
				if difficultyFilter(prevHash, nonce, soln, dataBytes, *d) {
					fmt.Printf("Sol%v satisfies difficulty: %v\n", i, soln)
					x = soln
					solFlag = true
					break
				}
			}
			if solFlag {
				break
			}
			nonce = nonce + 1
		}
		if !solFlag {
			fmt.Printf("Could not find valid nonce")
			return
		}
		t := time.Now()
		elapsed = t.Sub(start)
		currHash := blockHash(prevHash, nonce, x, dataBytes)
		fmt.Printf("-------------------------\n")
		fmt.Printf("Mined Block\n")
		fmt.Printf("Nonce: %#v\n", nonce)
		fmt.Printf("Previous Hash: %x\n", prevHash)
		fmt.Printf("Current Hash: %x\n", currHash)
		fmt.Printf("Time to find nonce: %v\n", elapsed)
		fmt.Printf("-------------------------\n")
		prevHash = currHash
	}
}

func main() {
	n := flag.Int("n", 64,
		"number of bits for each number")
	k := flag.Int("k", 7,
		"number of hashes XORed to zero")
	d := flag.Int("d", 2,
		"number of leading zeros")
	flag.Parse()
	fmt.Printf("n: %v, k: %v, d: %v\n", *n, *k, *d)
	mine(n, k, d)
}
