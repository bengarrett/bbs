//go:build !(js && wasm)

// Package main provides a stub for non-WASM environments.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "bbs-web is a WebAssembly package designed to be compiled with GOOS=js GOARCH=wasm")
	os.Exit(1)
}
