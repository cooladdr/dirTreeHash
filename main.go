package main

import (
	"fmt"

	"github.com/cooladdr/dirtreehash/dirHash"
)

func main() {
	fmt.Println("sha1 output in file[", dirHash.ComputingHash([]string{"."}), "]")
}
