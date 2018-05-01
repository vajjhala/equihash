package main

import "bytes"

func Power(a, n uint) uint {
	var i, result uint
	result = 1
	for i = 0; i < n; i++ {
		result *= a
	}
	return result
}

func (h hArrays) Less(i, j int) bool {
	switch bytes.Compare(h[i].hashSum, h[j].hashSum){
	case -1: return true
	default: return false
	}
}

func (h hArrays) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h hArrays) Len() int {
	return len(h)
}

func SafeXORBytes(a, b []byte) []byte {

	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	dest := make([]byte, n)
	for i := 0; i < n; i++ {
		dest[i] = a[i] ^ b[i]
	}
	return dest
}