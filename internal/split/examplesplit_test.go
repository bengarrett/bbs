package split_test

import (
	"fmt"

	"github.com/bengarrett/bbs/internal/split"
)

func ExampleBars() {
	s := "|03Hello |07|19world"
	l := len(split.Bars(s))
	fmt.Printf("Color sequences: %d", l)
	// Output: Color sequences: 3
}

func ExampleCelerity() {
	s := "|cHello |C|S|wworld"
	l := len(split.Celerity(s))
	fmt.Printf("Color sequences: %d", l)
	// Output: Color sequences: 4
}

func ExamplePCBoard() {
	s := "@X03Hello world"
	l := len(split.PCBoard(s))
	fmt.Printf("Color sequences: %d", l)
	// Output: Color sequences: 1
}
