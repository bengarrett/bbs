package bbs_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bengarrett/bbs"
)

func ExampleCelerityHTML() {
	src := []byte("|cHello |C|S|wworld")

	var buf bytes.Buffer
	if err := bbs.CelerityHTML(&buf, src...); err != nil {
		fmt.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="PBk PFc">Hello </i><i class="PBk PFC"></i><i class="PBw PFC">world</i>
}

func ExampleIsCelerity() {
	src := []byte("|cHello |C|S|wworld")

	fmt.Print(bbs.IsCelerity(src))
	// Output: true
}

func ExampleIsPCBoard() {
	src := []byte("@X03Hello world")

	fmt.Print(bbs.IsPCBoard(src))
	// Output: true
}

func ExampleIsRenegade() {
	src := []byte("|03Hello |07|19world")

	fmt.Print(bbs.IsRenegade(src))
	// Output: true
}

func ExampleIsTelegard() {
	const grave = "\u0060" // godoc treats a grave character as a special control
	src := []byte(grave + "7Hello world")

	fmt.Print(bbs.IsTelegard(src))
	// Output: true
}

func ExampleIsWWIVHash() {
	src := []byte("|#7Hello world")

	fmt.Print(bbs.IsWWIVHash(src))
	// Output: true
}

func ExampleIsWWIVHeart() {
	src := []byte("\x037Hello world")

	fmt.Print(bbs.IsWWIVHeart(src))
	// Output: true
}

func ExampleIsWildcat() {
	src := []byte("@0F@Hello world")

	fmt.Print(bbs.IsWildcat(src))
	// Output: true
}

func ExampleHTML() {
	src := strings.NewReader("@X03Hello world")

	var buf bytes.Buffer
	r, err := bbs.HTML(&buf, src)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("<!-- %s code -->\n", r)
	fmt.Print(buf.String())
	// Output: <!-- PCBoard @X code -->
	// <i class="PB0 PF3">Hello world</i>
}

func ExampleRenegadeHTML() {
	src := []byte("|03Hello |07|19world")

	var buf bytes.Buffer
	if err := bbs.RenegadeHTML(&buf, src...); err != nil {
		fmt.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="P0 P3">Hello </i><i class="P0 P7"></i><i class="P19 P7">world</i>
}

func ExamplePCBoardHTML() {
	src := []byte("@X03Hello world")

	var buf bytes.Buffer
	if err := bbs.PCBoardHTML(&buf, src...); err != nil {
		fmt.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleTrimControls() {
	src := []byte("@CLS@@PAUSE@Hello world")

	fmt.Printf("%q trims to %s", src, string(bbs.TrimControls(src...)))
	// Output: "@CLS@@PAUSE@Hello world" trims to Hello world
}

func ExampleFind() {
	src := strings.NewReader("@X03Hello world")

	f := bbs.Find(src)
	fmt.Printf("Found %s text", f.Name())
	// Output: Found PCBoard text
}

func ExampleFind_none() {
	src := strings.NewReader("Hello world")

	f := bbs.Find(src)
	if !f.Valid() {
		fmt.Print("Found plain text")
		return
	}
	fmt.Printf("Found %s text", f.Name())
	// Output: Found plain text
}

func ExampleFind_ansi() {
	const reset = "\x1b[0m" // an ANSI escape sequence to reset the terminal
	src := strings.NewReader(reset + "Hello world")

	f := bbs.Find(src)
	if !f.Valid() {
		fmt.Print("Found plain text")
		return
	}
	fmt.Printf("Found %s text", f.Name())
	// Output: Found ANSI text
}

func ExampleBBS_Bytes() {
	b := bbs.PCBoard.Bytes()
	fmt.Printf("Code as bytes %v\n", b)
	fmt.Printf("Code as string %s", b)
	// Output: Code as bytes [64 88]
	// Code as string @X
}

func ExampleBBS_CSS() {
	var css bytes.Buffer
	if err := bbs.PCBoard.CSS(&css); err != nil {
		fmt.Print(err)
	}
	// print the first 8 lines of the css
	lines := strings.Split(css.String(), "\n")
	for i := range 8 {
		fmt.Println(lines[i])
	}
	// Output: @import url("text_bbs.css");
	// @import url("text_blink.css");
	//
	// /* PCBoard and WildCat! BBS colours */
	//
	// i.PF0 {
	//     color: var(--black);
	// }
}

func ExampleBBS_HTML() {
	src := []byte("@X03Hello @X04world@X00")

	var buf bytes.Buffer
	if err := bbs.PCBoard.HTML(&buf, src); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print(buf.String())
	// Output: <i class="PB0 PF3">Hello </i><i class="PB0 PF4">world</i><i class="PB0 PF0"></i>
}

func ExampleBBS_HTML_find() {
	src := []byte("@X03Hello @X04world@X00")

	result := bbs.Find(bytes.NewReader(src))

	var buf bytes.Buffer
	if err := result.HTML(&buf, src); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print(buf.String())
	// Output: <i class="PB0 PF3">Hello </i><i class="PB0 PF4">world</i><i class="PB0 PF0"></i>
}

func ExampleBBS_HTML_ansi() {
	const reset = "\x1b[0m" // an ANSI escape sequence to reset the terminal
	src := []byte(reset + "Hello world")

	result := bbs.Find(bytes.NewReader(src))

	var buf bytes.Buffer
	if err := result.HTML(&buf, src); err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Print(buf.String())
	// Output: error: ansi escape code found
}

func ExampleBBS_Name() {
	fmt.Print(bbs.PCBoard.Name())
	// Output: PCBoard
}

func ExampleBBS_Remove() {
	src := []byte("@X03Hello @X07world")

	var buf bytes.Buffer
	if err := bbs.PCBoard.Remove(&buf, src...); err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%q to %s", src, buf.String())
	// Output: "@X03Hello @X07world" to Hello world
}

func ExampleBBS_Remove_find() {
	src := []byte("@X03Hello @X07world")

	result := bbs.Find(bytes.NewReader(src))

	var buf bytes.Buffer
	if err := result.Remove(&buf, src...); err != nil {
		fmt.Print(err)
	}
	fmt.Printf("%q to %s", src, buf.String())
	// Output: "@X03Hello @X07world" to Hello world
}

func ExampleBBS_String() {
	fmt.Print(bbs.PCBoard)
	// Output: PCBoard @X
}

func ExampleBBS_Valid() {
	src := "@X03Hello @X07world"

	r := strings.NewReader(src)
	ok := bbs.Find(r).Valid()
	fmt.Print(ok)
	// Output: true
}

func ExampleBBS_Valid_false() {
	src := "Hello world"

	r := strings.NewReader(src)
	ok := bbs.Find(r).Valid()
	fmt.Print(ok)
	// Output: false
}

func ExampleBBS_Valid_ansi() {
	const reset = "\x1b[0m" // an ANSI escape sequence to reset the terminal
	src := reset + "Hello world"

	r := strings.NewReader(src)
	ok := bbs.Find(r).Valid()
	fmt.Print(ok)
	// Output: true
}
