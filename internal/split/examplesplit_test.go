package split_test

import (
	"fmt"

	"github.com/bengarrett/bbs/internal/split"
)

func ExampleVBars() {
	b := []byte("|03Hello |07|19world")
	l := len(split.VBars(b))
	fmt.Printf("Color sequences: %d", l)
	// Output: Color sequences: 3
}

func ExampleCelerity() {
	b := []byte("|cHello |C|S|wworld")
	l := len(split.Celerity(b))
	fmt.Printf("Color sequences: %d", l)
	// Output: Color sequences: 4
}

func ExamplePCBoard() {
	s := []byte("@X03Hello world")
	l := len(split.PCBoard(s))
	fmt.Printf("Color sequences: %d", l)
	// Output: Color sequences: 1
}
