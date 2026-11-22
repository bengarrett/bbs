package bbs

import (
	"bytes"
	"regexp"
	"unicode/utf8"
)

// Columns is the default terminal line width used to wrap BBS color textfiles.
const Columns = 80

const (
	tabStop  = 8
	asciiEOF = 0x1a
)

// Options configures optional BBS conversion behavior.
type Options struct {
	// Unlock disables the default 80-column line wrapping.
	Unlock bool
	// Columns overrides the default 80-column wrap width when positive and Unlock is false.
	Columns int
}

func (o Options) wrapWidth() int {
	if o.Unlock {
		return 0
	}
	if o.Columns > 0 {
		return o.Columns
	}
	return Columns
}

// Wrap wraps src at cols visible columns without counting or splitting BBS color
// control codes. If cols <= 0 or len(src) == 0, src is returned unchanged.
func Wrap(kind BBS, cols int, src ...byte) []byte {
	if cols <= 0 || len(src) == 0 {
		return src
	}

	expr := kind.pattern()
	isUTF8 := hasUTF8HighBit(src)
	lineBreak := detectLineBreak(src)

	var out bytes.Buffer
	out.Grow(len(src))

	col := 0
	for pos := 0; pos < len(src); {
		if expr != nil {
			if codeLen := matchPrefix(expr, src[pos:]); codeLen > 0 {
				out.Write(src[pos : pos+codeLen])
				pos += codeLen
				continue
			}
		}

		var step int
		step, col = writeWrappedByte(&out, src[pos:], lineBreak, cols, col, isUTF8)
		pos += step
	}

	return out.Bytes()
}

func writeWrappedByte(
	out *bytes.Buffer,
	tail []byte,
	lineBreak []byte,
	cols, col int,
	isUTF8 bool,
) (int, int) {
	cur := tail[0]
	switch cur {
	case '\r', '\n':
		out.WriteByte(cur)
		return 1, 0
	case asciiEOF:
		out.WriteByte(cur)
		return 1, col
	case '\t':
		if col >= cols {
			out.Write(lineBreak)
			col = 0
		}
		out.WriteByte(cur)
		return 1, min((col+tabStop)&^(tabStop-1), cols)
	default:
		step := 1
		if isUTF8 {
			_, step = utf8.DecodeRune(tail)
		}
		if col >= cols {
			out.Write(lineBreak)
			col = 0
		}
		out.Write(tail[:step])
		return step, col + 1
	}
}

func (b BBS) pattern() *regexp.Regexp {
	switch b {
	case Celerity:
		return celerityRe
	case PCBoard:
		return pcboardRe
	case Renegade:
		return renegadeRe
	case Telegard:
		return telegardRe
	case Wildcat:
		return wildcatRe
	case WWIVHash:
		return wwivHashRe
	case WWIVHeart:
		return wwivHeartRe
	case ANSI:
		return nil
	default:
		return nil
	}
}

func matchPrefix(expr *regexp.Regexp, tail []byte) int {
	loc := expr.FindIndex(tail)
	if loc != nil && loc[0] == 0 {
		return loc[1]
	}
	return 0
}

func hasUTF8HighBit(src []byte) bool {
	highBit := false
	for _, elem := range src {
		if elem >= utf8.RuneSelf {
			highBit = true
			break
		}
	}
	return highBit && utf8.Valid(src)
}

func detectLineBreak(src []byte) []byte {
	if bytes.Contains(src, []byte("\r\n")) {
		return []byte("\r\n")
	}
	return []byte("\n")
}
