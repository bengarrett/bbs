package bbs_test

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/bengarrett/bbs"
)

func ExampleFields() {
	r := strings.NewReader("@X03Hello @XF0world")

	s, b, err := bbs.Fields(r)
	if err != nil {
		log.Print(err)
	}

	fmt.Printf("Found %d, %s sequences\n", len(s), b)
	for i, item := range s {
		fmt.Printf("Sequence %d: %q\n", i+1, item)
	}
	// Output: Found 2, PCBoard @X sequences
	// Sequence 1: "03Hello "
	// Sequence 2: "F0world"
	//
}

func ExampleFields_ansi() {
	const reset = "\x1b[0m" // an ANSI escape sequence to reset the terminal
	r := strings.NewReader(reset + "Hello world")

	s, b, err := bbs.Fields(r)
	if errors.Is(err, bbs.ErrANSI) {
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Printf("Found %d, %s sequences\n", len(s), b)
	// Output: error: ansi escape code found
}

func ExampleFields_none() {
	r := strings.NewReader("Hello world")

	s, b, err := bbs.Fields(r)
	if errors.Is(err, bbs.ErrNone) {
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Printf("Found %d, %s sequences\n", len(s), b)
	// Output: error: no bbs color code found
}
