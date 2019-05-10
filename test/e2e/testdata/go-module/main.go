package main

import (
	"fmt"
	"golang.org/x/crypto/acme"
)

func main() {
	fmt.Println(acme.LetsEncryptURL)
}
