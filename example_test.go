package bbs_test

import (
	"bytes"
	"embed"
	"fmt"
	"log"

	"github.com/bengarrett/bbs"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

//go:embed static/*
var static embed.FS

func Example() {
	// print about the file
	file, err := static.Open("static/examples/hello.pcb")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	s, name, err := bbs.Fields(file)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("Found %d %s color controls.\n\n", len(s), name)

	// reopen the file
	file, err = static.Open("static/examples/hello.pcb")
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()

	// transform the MS-DOS legacy text to Unicode
	decoder := charmap.CodePage437.NewDecoder()
	reader := transform.NewReader(file, decoder)

	// create the HTML equivalent of BBS color codes
	var buf bytes.Buffer
	if _, err := bbs.HTML(&buf, reader); err != nil {
		log.Print(err)
		return
	}
	fmt.Print(buf.String())

	// Output: Found 11 PCBoard @X color controls.
	//
	// <i class="PB0 PFF">    </i><i class="PB7 PF0"> ┌─────────────┐ </i><i class="PB0 PF7">
	// </i><i class="PB0 PFF">    </i><i class="PB7 PF0"> │ Hello </i><i class="PBF PF0">world </i><i class="PB7 PF0">│ </i><i class="PB0 PF7">
	// </i><i class="PB0 PFF">    </i><i class="PB7 PF0"> └─────────────┘ </i><i class="PB0 PF7"></i>
}
