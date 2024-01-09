package bbs_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/bengarrett/bbs"
)

func ExampleHTML() {
	var buf bytes.Buffer
	src := strings.NewReader("@X03Hello world")
	if _, err := bbs.HTML(&buf, src); err != nil {
		fmt.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleRenegadeHTML() {
	var buf bytes.Buffer
	src := []byte("|03Hello |07|19world")
	if err := bbs.RenegadeHTML(&buf, src); err != nil {
		log.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="P0 P3">Hello </i><i class="P0 P7"></i><i class="P19 P7">world</i>
}

func ExampleCelerityHTML() {
	var buf bytes.Buffer
	src := []byte("|cHello |C|S|wworld")
	if err := bbs.CelerityHTML(&buf, src); err != nil {
		fmt.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="PBk PFc">Hello </i><i class="PBk PFC"></i><i class="PBw PFC">world</i>
}

func ExamplePCBoardHTML() {
	var buf bytes.Buffer
	src := []byte("@X03Hello world")
	if err := bbs.PCBoardHTML(&buf, src); err != nil {
		log.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleIsCelerity() {
	b := []byte("|cHello |C|S|wworld")
	fmt.Printf("Is b Celerity BBS text? %v", bbs.IsCelerity(b))
	// Output: Is b Celerity BBS text? true
}

func ExampleIsPCBoard() {
	b := []byte("@X03Hello world")
	fmt.Printf("Is b PCBoard BBS text? %v", bbs.IsPCBoard(b))
	// Output: Is b PCBoard BBS text? true
}

func ExampleIsRenegade() {
	b := []byte("|03Hello |07|19world")
	fmt.Printf("Is b Renegade BBS text? %v", bbs.IsRenegade(b))
	// Output: Is b Renegade BBS text? true
}

func ExampleIsTelegard() {
	const grave = "\u0060" // godoc treats a grave character as a special control
	b := []byte(grave + "7Hello world")
	fmt.Printf("Is b Telegard BBS text? %v", bbs.IsTelegard(b))
	// Output: Is b Telegard BBS text? true
}

func ExampleIsWWIVHash() {
	b := []byte("|#7Hello world")
	fmt.Printf("Is b WVIV BBS # text? %v", bbs.IsWWIVHash(b))
	// Output: Is b WVIV BBS # text? true
}

func ExampleIsWWIVHeart() {
	b := []byte("\x037Hello world")
	fmt.Printf("Is b WWIV BBS ♥ text? %v", bbs.IsWWIVHeart(b))
	// Output: Is b WWIV BBS ♥ text? true
}

func ExampleIsWildcat() {
	b := []byte("@0F@Hello world")
	fmt.Printf("Is b Wildcat! BBS text? %v", bbs.IsWildcat(b))
	// Output: Is b Wildcat! BBS text? true
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
		fmt.Print("Reader is plain text")
	}
	// Output: Reader is plain text
}

func ExampleBBS_Bytes() {
	b := bbs.PCBoard.Bytes()
	fmt.Printf("%s %v", b, b)
	// Output: @X [64 88]
}

func ExampleBBS_CSS() {
	var dst bytes.Buffer
	if err := bbs.PCBoard.CSS(&dst); err != nil {
		fmt.Print(err)
	}
	// print the first 8 lines
	lines := strings.Split(dst.String(), "\n")
	for i := 0; i < 8; i++ {
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
	var buf bytes.Buffer
	src := []byte("@X03Hello world")
	if err := bbs.PCBoard.HTML(&buf, src); err != nil {
		log.Print(err)
	}
	fmt.Print(buf.String())
	// Output: <i class="PB0 PF3">Hello world</i>
}

func ExampleBBS_Name() {
	fmt.Print(bbs.PCBoard.Name())
	// Output: PCBoard
}

func ExampleBBS_Remove() {
	var buf bytes.Buffer
	src := []byte("@X03Hello @X07world")
	if err := bbs.PCBoard.Remove(&buf, src); err != nil {
		log.Print(err)
	}
	fmt.Printf("%q to %q", src, buf.String())
	// Output: "@X03Hello @X07world" to "Hello world"
}

func ExampleBBS_Remove_find() {
	var buf bytes.Buffer
	src := []byte("@X03Hello @X07world")
	r := bytes.NewReader(src)
	b := bbs.Find(r)
	if err := b.Remove(&buf, src); err != nil {
		log.Print(err)
	}
	fmt.Printf("%q to %q", src, buf.String())
	// Output: "@X03Hello @X07world" to "Hello world"
}

func ExampleBBS_String() {
	fmt.Print(bbs.PCBoard)
	// Output: PCBoard @X
}

func ExampleBBS_Valid() {
	r := strings.NewReader("Hello world")
	f := bbs.Find(r)
	fmt.Print("Is reader BBS text? ", f.Valid())
	// Output: Is reader BBS text? false
}
