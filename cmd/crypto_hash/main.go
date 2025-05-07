package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	row := "| 1 | Taras | 5000"
	h := sha256.Sum256([]byte(row)) // 32 byte
	fmt.Printf("%x\n", h)

	
	
}
