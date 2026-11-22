package bbs_test

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/bengarrett/bbs"
)

func assertPixelRGB(t *testing.T, img image.Image, x, y int, wantR, wantG, wantB uint32) {
	t.Helper()
	r, g, b, _ := img.At(x, y).RGBA()
	if r>>8 != wantR || g>>8 != wantG || b>>8 != wantB {
		t.Errorf("pixel(%d,%d) = (%d,%d,%d), want (%d,%d,%d)", x, y, r>>8, g>>8, b>>8, wantR, wantG, wantB)
	}
}

func Test_PNGColorsAndDimensions(t *testing.T) {
	var buf bytes.Buffer
	// Row 0: @X1E + CP437 0xDB (solid block) -> Yellow (255,255,85)
	// Row 1: @X4C + space -> Red bg (170,0,0)
	src := "@CLS@@X1E\xdb\r\n@X4C @PAUSE@"
	kind, err := bbs.PNG(&buf, strings.NewReader(src))
	if err != nil {
		t.Fatalf("PNG() error = %v", err)
	}
	if kind != bbs.PCBoard {
		t.Fatalf("PNG() kind = %v, want PCBoard", kind)
	}

	img, err := png.Decode(&buf)
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}
	if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 32 {
		t.Fatalf("PNG bounds = %dx%d, want 8x32", img.Bounds().Dx(), img.Bounds().Dy())
	}

	assertPixelRGB(t, img, 0, 0, 0xff, 0xff, 0x55)
	assertPixelRGB(t, img, 0, 16, 0xaa, 0x00, 0x00)
}

func Test_PNGWrapAndUnlock(t *testing.T) {
	long := "@X0F" + strings.Repeat("X", 85)
	var locked bytes.Buffer
	if err := bbs.PCBoard.PNG(&locked, []byte(long)); err != nil {
		t.Fatalf("PCBoard.PNG error = %v", err)
	}
	imgLocked, err := png.Decode(&locked)
	if err != nil {
		t.Fatal(err)
	}
	if imgLocked.Bounds().Dx() != 80*8 || imgLocked.Bounds().Dy() != 2*16 {
		t.Errorf("locked PNG = %dx%d, want 640x32", imgLocked.Bounds().Dx(), imgLocked.Bounds().Dy())
	}

	var unlocked bytes.Buffer
	if err := bbs.PCBoard.PNGOptions(&unlocked, []byte(long), bbs.Options{Unlock: true}); err != nil {
		t.Fatalf("PCBoard.PNGOptions(unlock) error = %v", err)
	}
	imgUnlocked, err := png.Decode(&unlocked)
	if err != nil {
		t.Fatal(err)
	}
	if imgUnlocked.Bounds().Dx() != 85*8 || imgUnlocked.Bounds().Dy() != 16 {
		t.Errorf("unlocked PNG = %dx%d, want 680x16", imgUnlocked.Bounds().Dx(), imgUnlocked.Bounds().Dy())
	}
}

func Test_PNGRenegadeCelerityAndUTF8(t *testing.T) {
	var rene bytes.Buffer
	if err := bbs.Renegade.PNG(&rene, []byte("|17|14Hi")); err != nil || rene.Len() == 0 {
		t.Errorf("Renegade.PNG error = %v, len = %d", err, rene.Len())
	}
	var cel bytes.Buffer
	if err := bbs.Celerity.PNG(&cel, []byte("|b|SHi")); err != nil || cel.Len() == 0 {
		t.Errorf("Celerity.PNG error = %v, len = %d", err, cel.Len())
	}
	var u8 bytes.Buffer
	if err := bbs.PCBoard.PNG(&u8, []byte("@X1E█")); err != nil {
		t.Fatalf("utf8 PCBoard.PNG error = %v", err)
	}
	img, err := png.Decode(&u8)
	if err != nil {
		t.Fatal(err)
	}
	assertPixelRGB(t, img, 0, 0, 0xff, 0xff, 0x55)
}

func Test_PNGLargeReaderNotTruncatedAt4KB(t *testing.T) {
	// bufio.Scanner buffers 4096 bytes at a time; ensure io.Reader helpers
	// read the entire stream past 4096 bytes when Find matches on line 1.
	line := "@X0F" + strings.Repeat("A", 80) + "\r\n"
	src := strings.Repeat(line, 60) // > 5000 bytes, 60 rows
	var buf bytes.Buffer
	if _, err := bbs.PNG(&buf, strings.NewReader(src)); err != nil {
		t.Fatalf("PNG(large reader) error = %v", err)
	}
	img, err := png.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dy() != 60*16 {
		t.Errorf("PNG height = %d, want %d (60 rows)", img.Bounds().Dy(), 60*16)
	}
}
