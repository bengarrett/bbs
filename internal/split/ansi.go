package split

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	"github.com/bengarrett/bbs/internal/pattern"
)

const (
	defaultFG    = 7
	defaultBG    = 0
	maxVBarFG    = 15
	vbarBGOffset = 16
	colorMask    = 7
	brightCutoff = 8
	hexDecimal   = 10
	sgrFGBase    = 30
	sgrBGBase    = 40
	swapCmd      = 'S'
)

// CelerityANSI writes to buf the ANSI SGR equivalent of Celerity BBS color codes
// while preserving raw text and CP437 bytes.
func CelerityANSI(buf *bytes.Buffer, src []byte) error {
	if buf == nil {
		return ErrBuff
	}
	if len(src) == 0 {
		return nil
	}
	matches := pattern.CelerityRe.FindAllSubmatchIndex(src, -1)
	if len(matches) == 0 {
		return writeRaw(buf, src)
	}

	curFG, curBG := defaultFG, defaultBG
	last := 0
	for _, match := range matches {
		if _, err := buf.Write(src[last:match[0]]); err != nil {
			return fmt.Errorf("write buffer: %w", err)
		}
		cmd := src[match[2]]
		if cmd == swapCmd {
			curFG, curBG = curBG, curFG
			writeANSI(buf, curFG, curBG, false)
		} else if colorIdx, ok := celerityIndex(cmd); ok {
			curFG = colorIdx
			writeANSI(buf, curFG, curBG, false)
		}
		last = match[1]
	}
	return writeRaw(buf, src[last:])
}

// PCBoardANSI writes to buf the ANSI SGR equivalent of PCBoard @X color codes
// while preserving raw text and CP437 bytes.
func PCBoardANSI(buf *bytes.Buffer, src []byte) error {
	if buf == nil {
		return ErrBuff
	}
	if len(src) == 0 {
		return nil
	}
	matches := pattern.PCBoardRe.FindAllSubmatchIndex(src, -1)
	if len(matches) == 0 {
		return writeRaw(buf, src)
	}

	last := 0
	for _, match := range matches {
		if _, err := buf.Write(src[last:match[0]]); err != nil {
			return fmt.Errorf("write buffer: %w", err)
		}
		hexPair := src[match[2]:match[3]]
		cgaBG := hexNibble(hexPair[0])
		cgaFG := hexNibble(hexPair[1])
		blink := cgaBG >= brightCutoff
		writeANSI(buf, cgaFG, cgaBG, blink)
		last = match[1]
	}
	return writeRaw(buf, src[last:])
}

// VBarsANSI writes to buf the ANSI SGR equivalent of vertical-bar BBS color codes
// (Renegade and WWIV) while preserving raw text and CP437 bytes.
func VBarsANSI(buf *bytes.Buffer, src []byte) error {
	if buf == nil {
		return ErrBuff
	}
	return vBarsANSI(buf, pattern.RenegadeRe, src)
}

func vBarsANSI(buf *bytes.Buffer, expr *regexp.Regexp, src []byte) error {
	if len(src) == 0 {
		return nil
	}
	matches := expr.FindAllSubmatchIndex(src, -1)
	if len(matches) == 0 {
		return writeRaw(buf, src)
	}

	curFG, curBG := defaultFG, defaultBG
	last := 0
	for index, match := range matches {
		if _, err := buf.Write(src[last:match[0]]); err != nil {
			return fmt.Errorf("write buffer: %w", err)
		}
		code, err := strconv.Atoi(string(src[match[2]:match[3]]))
		if err != nil {
			return fmt.Errorf("bar #%d is not a bbs code: %w", index, err)
		}
		if code <= maxVBarFG {
			curFG = code
		} else {
			curBG = code - vbarBGOffset
		}
		writeANSI(buf, curFG, curBG, false)
		last = match[1]
	}
	return writeRaw(buf, src[last:])
}

func writeRaw(buf *bytes.Buffer, src []byte) error {
	if _, err := buf.Write(src); err != nil {
		return fmt.Errorf("write buffer: %w", err)
	}
	return nil
}

func celerityIndex(code byte) (int, bool) {
	// Matches Celerity CGA color order: k,b,g,c,r,m,y,w,d,B,G,C,R,M,Y,W.
	const palette = "kbgcrmywdBGCRMYW"
	idx := bytes.IndexByte([]byte(palette), code)
	if idx < 0 {
		return 0, false
	}
	return idx, true
}

func hexNibble(val byte) int {
	switch {
	case val >= '0' && val <= '9':
		return int(val - '0')
	case val >= 'A' && val <= 'F':
		return int(val-'A') + hexDecimal
	default:
		return 0
	}
}

// cgaToANSI maps a 3-bit IBM CGA color index (0..7) to the ANSI SGR color offset (0..7).
func cgaToANSI(cgaColor int) int {
	palette := [...]int{
		0, // 0: Black -> 30/40
		4, // 1: Blue -> 34/44
		2, // 2: Green -> 32/42
		6, // 3: Cyan -> 36/46
		1, // 4: Red -> 31/41
		5, // 5: Magenta -> 35/45
		3, // 6: Brown -> 33/43
		7, // 7: Light Gray -> 37/47
	}
	return palette[cgaColor&colorMask]
}

func writeANSI(buf *bytes.Buffer, cgaFG, cgaBG int, blink bool) {
	fgCode := sgrFGBase + cgaToANSI(cgaFG)
	bgCode := sgrBGBase + cgaToANSI(cgaBG)

	buf.WriteString("\x1b[0")
	if blink {
		buf.WriteString(";5")
	}
	if cgaFG >= brightCutoff {
		buf.WriteString(";1")
	}
	fmt.Fprintf(buf, ";%d;%dm", bgCode, fgCode)
}
