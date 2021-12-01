package bbs_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bengarrett/bbs"
)

func ExampleHTML() {
	var out bytes.Buffer
	src := strings.NewReader("@X03Hello world")
	if _, err := bbs.HTML(&out, src); err != nil {
		fmt.Print(err)
	}
	fmt.Print(out.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleHTMLRenegade() {
	var out bytes.Buffer
	src := []byte("|03Hello |07|19world")
	if err := bbs.HTMLRenegade(&out, src); err != nil {
		log.Print(err)
	}
	fmt.Print(out.String())
	// Output: <i class="P0 P3">Hello </i><i class="P0 P7"></i><i class="P19 P7">world</i>
}

func ExampleHTMLCelerity() {
	var out bytes.Buffer
	src := []byte("|cHello |C|S|wworld")
	if err := bbs.HTMLCelerity(&out, src); err != nil {
		fmt.Print(err)
	}
	fmt.Print(out.String())
	// Output: <i class="PBk PFc">Hello </i><i class="PBk PFC"></i><i class="PBw PFC">world</i>
}

func ExampleHTMLPCBoard() {
	var out bytes.Buffer
	src := []byte("@X03Hello world")
	if err := bbs.HTMLPCBoard(&out, src); err != nil {
		log.Print(err)
	}
	fmt.Print(out.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleHasCelerity() {
	b := []byte("|cHello |C|S|wworld")
	fmt.Printf("Has b Celerity BBS text? %v", bbs.HasCelerity(b))
	// Output: Has b Celerity BBS text? true
}

func ExampleHasPCBoard() {
	b := []byte("@X03Hello world")
	fmt.Printf("Has b PCBoard BBS text? %v", bbs.HasPCBoard(b))
	// Output: Has b PCBoard BBS text? true
}

func ExampleHasRenegade() {
	b := []byte("|03Hello |07|19world")
	fmt.Printf("Has b Renegade BBS text? %v", bbs.HasRenegade(b))
	// Output: Has b Renegade BBS text? true
}

func ExampleHasTelegard() {
	const grave = "\u0060" // godoc treats a grave character as a special control
	b := []byte(grave + "7Hello world")
	fmt.Printf("Has b Telegard BBS text? %v", bbs.HasTelegard(b))
	// Output: Has b Telegard BBS text? true
}

func ExampleHasWHash() {
	b := []byte("|#7Hello world")
	fmt.Printf("Has b WVIV BBS # text? %v", bbs.HasWHash(b))
	// Output: Has b WVIV BBS # text? true
}
func ExampleHasWHeart() {
	b := []byte("\x037Hello world")
	fmt.Printf("Has b WWIV BBS ♥ text? %v", bbs.HasWHeart(b))
	// Output: Has b WWIV BBS ♥ text? true
}
func ExampleHasWildcat() {
	b := []byte("@0F@Hello world")
	fmt.Printf("Has b Wildcat! BBS text? %v", bbs.HasWildcat(b))
	// Output: Has b Wildcat! BBS text? true
}

func ExampleTrimControls() {
	b := []byte("@CLS@@PAUSE@Hello world")
	r := bbs.TrimControls(b)
	fmt.Print(string(r))
	// Output: Hello world
}

func ExampleFind() {
	r := strings.NewReader("@X03Hello world")
	f := bbs.Find(r)
	fmt.Printf("Reader is in a %s BBS format", f.Name())
	// Output: Reader is in a PCBoard BBS format
}

func ExampleFind_none() {
	r := strings.NewReader("Hello world")
	f := bbs.Find(r)
	if !f.Valid() {
		fmt.Print("reader is plain text")
	}
	// Output: reader is plain text
}

func ExampleBBS_Bytes() {
	b := bbs.PCBoard.Bytes()
	fmt.Printf("%s %v", b, b)
	// Output: @X [64 88]
}

func ExampleBBS_CSS() {
	var buf bytes.Buffer
	if err := bbs.PCBoard.CSS(&buf); err != nil {
		fmt.Print(err)
	}

	f, err := os.OpenFile("pcboard.css", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name()) // clean up

	if _, err := buf.WriteTo(f); err != nil {
		log.Fatal(err)
	}
}

func ExampleBBS_HTML() {
	var out bytes.Buffer
	src := []byte("@X03Hello world")
	if err := bbs.PCBoard.HTML(&out, src); err != nil {
		log.Print(err)
	}
	fmt.Print(out.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleBBS_Name() {
	fmt.Print(bbs.PCBoard.Name())
	// Output: PCBoard
}

func ExampleBBS_Remove() {
	var out bytes.Buffer
	src := []byte("@X03Hello @X07world")
	if err := bbs.PCBoard.Remove(&out, src); err != nil {
		log.Print(err)
	}
	fmt.Print(out.String())
	// Output: Hello world
}

func ExampleBBS_Remove_find() {
	var out bytes.Buffer
	src := []byte("@X03Hello @X07world")
	r := bytes.NewReader(src)
	b := bbs.Find(r)
	if err := b.Remove(&out, src); err != nil {
		log.Print(err)
	}
	fmt.Print(out.String())
	// Output: Hello world
}

func ExampleBBS_String() {
	fmt.Print(bbs.PCBoard)
	// Output: PCBoard @X
}

func ExampleBBS_Valid() {
	r := strings.NewReader("Hello world")
	f := bbs.Find(r)
	fmt.Print("reader is BBS text? ", f.Valid())
	// Output: reader is BBS text? false
}
