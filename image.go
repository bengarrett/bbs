package bbs

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

//go:embed font_vga_8x16.bin
var fontVGA8x16 []byte

const (
	glyphWidth     = 8
	glyphHeight    = 16
	glyphByteLen   = 16
	glyphBitHigh   = 0x80
	bytesPerRGBA   = 4
	rgbaR          = 0
	rgbaG          = 1
	rgbaB          = 2
	rgbaA          = 3
	defaultCellFG  = 7
	defaultCellBG  = 0
	colorNibbleMax = 15
	bgColorMask    = 7
	vbarBGStart    = 16
	hexDecimalBase = 10
	sauceRecordLen = 128
	sauceTagLen    = 5
	asciiNUL       = 0x00
)

type gridCell struct {
	char byte
	fg   byte
	bg   byte
}

// Image renders the first detected BBS color format from src onto an 8x16 VGA
// character grid using the default 80-column wrap width.
func Image(src io.Reader) (image.Image, BBS, error) {
	return ImageOptions(src, Options{})
}

// ImageOptions renders the first detected BBS color format from src using opts.
func ImageOptions(src io.Reader, opts Options) (image.Image, BBS, error) {
	var dummy bytes.Buffer
	raw, find, err := findAndRead(&dummy, src)
	if err != nil {
		return nil, find, err
	}
	img, err := find.ImageOptions(raw, opts)
	return img, find, err
}

// PNG writes a PNG-encoded VGA rendering of src to buf using the default
// 80-column wrap width.
func PNG(buf *bytes.Buffer, src io.Reader) (BBS, error) {
	return PNGOptions(buf, src, Options{})
}

// PNGOptions writes a PNG-encoded VGA rendering of src to buf using opts.
func PNGOptions(buf *bytes.Buffer, src io.Reader, opts Options) (BBS, error) {
	if buf == nil {
		return -1, ErrBuff
	}
	img, find, err := ImageOptions(src, opts)
	if err != nil {
		return find, err
	}
	if err := png.Encode(buf, img); err != nil {
		return find, fmt.Errorf("encode png: %w", err)
	}
	return find, nil
}

// Image renders src onto an 8x16 VGA character grid using the default 80-column wrap width.
func (b BBS) Image(src []byte) (image.Image, error) {
	return b.ImageOptions(src, Options{})
}

// ImageOptions renders src onto an 8x16 VGA character grid using opts.
func (b BBS) ImageOptions(src []byte, opts Options) (image.Image, error) {
	if b == ANSI {
		return nil, ErrANSI
	}
	if !b.Valid() {
		return nil, ErrNone
	}
	if len(src) == 0 {
		return nil, ErrNone
	}

	cleaned := TrimControls(stripTrailingSAUCE(src)...)
	cp437Bytes := encodeCP437IfUTF8(cleaned)
	wrapped := Wrap(b, opts.wrapWidth(), cp437Bytes...)

	grid := b.buildGrid(wrapped)
	return drawVGAGrid(grid), nil
}

// PNG writes a PNG-encoded VGA rendering of src to buf using the default 80-column wrap width.
func (b BBS) PNG(buf *bytes.Buffer, src []byte) error {
	return b.PNGOptions(buf, src, Options{})
}

// PNGOptions writes a PNG-encoded VGA rendering of src to buf using opts.
func (b BBS) PNGOptions(buf *bytes.Buffer, src []byte, opts Options) error {
	if buf == nil {
		return ErrBuff
	}
	img, err := b.ImageOptions(src, opts)
	if err != nil {
		return err
	}
	if err := png.Encode(buf, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	return nil
}

func stripTrailingSAUCE(src []byte) []byte {
	trimmed := bytes.TrimRight(src, string([]byte{asciiEOF}))
	if len(trimmed) < sauceRecordLen {
		return trimmed
	}
	tail := trimmed[len(trimmed)-sauceRecordLen:]
	if !bytes.Equal(tail[:sauceTagLen], []byte("SAUCE")) {
		return trimmed
	}
	body := trimmed[:len(trimmed)-sauceRecordLen]
	return bytes.TrimRight(body, string([]byte{0x00, asciiEOF}))
}

func encodeCP437IfUTF8(src []byte) []byte {
	if !hasUTF8HighBit(src) {
		return src
	}
	enc := encoding.ReplaceUnsupported(charmap.CodePage437.NewEncoder())
	out, _, err := transform.Bytes(enc, src)
	if err != nil {
		return src
	}
	return out
}

func (b BBS) buildGrid(src []byte) [][]gridCell {
	kind, data := b.normalize(src)
	expr := kind.pattern()

	grid := [][]gridCell{{}}
	curFG := byte(defaultCellFG)
	curBG := byte(defaultCellBG)
	row := 0

	for pos := 0; pos < len(data); {
		if expr != nil {
			if codeLen := matchPrefix(expr, data[pos:]); codeLen > 0 {
				curFG, curBG = applyGridColor(kind, data[pos:pos+codeLen], curFG, curBG)
				pos += codeLen
				continue
			}
		}

		ch := data[pos]
		pos++

		switch ch {
		case '\r', asciiEOF:
			continue
		case '\n':
			grid = append(grid, []gridCell{})
			row++
		case '\t':
			spaces := tabStop - (len(grid[row]) % tabStop)
			for range spaces {
				grid[row] = append(grid[row], gridCell{char: ' ', fg: curFG, bg: curBG})
			}
		case asciiNUL:
			grid[row] = append(grid[row], gridCell{char: ' ', fg: curFG, bg: curBG})
		default:
			grid[row] = append(grid[row], gridCell{char: ch, fg: curFG, bg: curBG})
		}
	}

	if len(grid) > 1 && len(grid[len(grid)-1]) == 0 {
		grid = grid[:len(grid)-1]
	}
	return grid
}

func applyGridColor(kind BBS, token []byte, curFG, curBG byte) (byte, byte) {
	switch kind {
	case PCBoard:
		return applyPCBoardColor(token, curFG, curBG)
	case Renegade:
		return applyRenegadeColor(token, curFG, curBG)
	case Celerity:
		return applyCelerityColor(token, curFG, curBG)
	case ANSI, Wildcat, Telegard, WWIVHash, WWIVHeart:
		return curFG, curBG
	default:
		return curFG, curBG
	}
}

func applyPCBoardColor(token []byte, curFG, curBG byte) (byte, byte) {
	const minTokenLen = 4
	if len(token) < minTokenLen {
		return curFG, curBG
	}
	bg := parseHexChar(token[2]) & bgColorMask
	fg := parseHexChar(token[3]) & colorNibbleMax
	return fg, bg
}

func applyRenegadeColor(token []byte, curFG, curBG byte) (byte, byte) {
	const minTokenLen = 3
	if len(token) < minTokenLen {
		return curFG, curBG
	}
	if token[1] < '0' || token[1] > '9' || token[2] < '0' || token[2] > '9' {
		return curFG, curBG
	}
	code := (token[1]-'0')*hexDecimalBase + (token[2] - '0')
	if code < vbarBGStart {
		return code & colorNibbleMax, curBG
	}
	return curFG, (code - vbarBGStart) & bgColorMask
}

func applyCelerityColor(token []byte, curFG, curBG byte) (byte, byte) {
	const minTokenLen = 2
	if len(token) < minTokenLen {
		return curFG, curBG
	}
	if token[1] == 'S' {
		return curBG, curFG & bgColorMask
	}
	const celerityPalette = "kbgcrmywdBGCRMYW"
	for idx := byte(0); int(idx) < len(celerityPalette); idx++ {
		if celerityPalette[idx] == token[1] {
			return idx, curBG
		}
	}
	return curFG, curBG
}

func parseHexChar(val byte) byte {
	switch {
	case val >= '0' && val <= '9':
		return val - '0'
	case val >= 'A' && val <= 'F':
		return val - 'A' + hexDecimalBase
	default:
		return 0
	}
}

func drawVGAGrid(grid [][]gridCell) *image.NRGBA {
	maxCols := 1
	for _, line := range grid {
		if len(line) > maxCols {
			maxCols = len(line)
		}
	}
	rows := max(len(grid), 1)
	palette := vgaCGAPalette()
	img := image.NewNRGBA(image.Rect(0, 0, maxCols*glyphWidth, rows*glyphHeight))
	fillBackground(img, palette[defaultCellBG])

	for rowIdx, line := range grid {
		for colIdx, cell := range line {
			drawVGACell(img, colIdx, rowIdx, cell, palette)
		}
	}
	return img
}

func fillBackground(img *image.NRGBA, bg color.NRGBA) {
	pix := img.Pix
	for idx := 0; idx+rgbaA < len(pix); idx += bytesPerRGBA {
		pix[idx+rgbaR] = bg.R
		pix[idx+rgbaG] = bg.G
		pix[idx+rgbaB] = bg.B
		pix[idx+rgbaA] = bg.A
	}
}

func drawVGACell(img *image.NRGBA, col, row int, cell gridCell, palette [16]color.NRGBA) {
	fg := palette[cell.fg&colorNibbleMax]
	bg := palette[cell.bg&bgColorMask]

	glyphOffset := int(cell.char) * glyphByteLen
	if glyphOffset+glyphByteLen > len(fontVGA8x16) {
		return
	}
	glyph := fontVGA8x16[glyphOffset : glyphOffset+glyphByteLen]

	startX := col * glyphWidth
	startY := row * glyphHeight
	for gy := range glyphHeight {
		bits := glyph[gy]
		rowOff := (startY + gy) * img.Stride
		for gx := range glyphWidth {
			pixOff := rowOff + (startX+gx)*bytesPerRGBA
			pixel := bg
			if (bits & (glyphBitHigh >> gx)) != 0 {
				pixel = fg
			}
			img.Pix[pixOff+rgbaR] = pixel.R
			img.Pix[pixOff+rgbaG] = pixel.G
			img.Pix[pixOff+rgbaB] = pixel.B
			img.Pix[pixOff+rgbaA] = pixel.A
		}
	}
}

func vgaCGAPalette() [16]color.NRGBA {
	const (
		c00 = 0x00
		c55 = 0x55
		caa = 0xaa
		cff = 0xff
	)
	return [16]color.NRGBA{
		{R: c00, G: c00, B: c00, A: cff}, // 0: Black
		{R: c00, G: c00, B: caa, A: cff}, // 1: Blue
		{R: c00, G: caa, B: c00, A: cff}, // 2: Green
		{R: c00, G: caa, B: caa, A: cff}, // 3: Cyan
		{R: caa, G: c00, B: c00, A: cff}, // 4: Red
		{R: caa, G: c00, B: caa, A: cff}, // 5: Magenta
		{R: caa, G: c55, B: c00, A: cff}, // 6: Brown
		{R: caa, G: caa, B: caa, A: cff}, // 7: Light Gray
		{R: c55, G: c55, B: c55, A: cff}, // 8: Dark Gray
		{R: c55, G: c55, B: cff, A: cff}, // 9: Light Blue
		{R: c55, G: cff, B: c55, A: cff}, // A: Light Green
		{R: c55, G: cff, B: cff, A: cff}, // B: Light Cyan
		{R: cff, G: c55, B: c55, A: cff}, // C: Light Red
		{R: cff, G: c55, B: cff, A: cff}, // D: Light Magenta
		{R: cff, G: cff, B: c55, A: cff}, // E: Yellow
		{R: cff, G: cff, B: cff, A: cff}, // F: White
	}
}
