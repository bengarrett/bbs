// Package bbs interacts with legacy textfiles encoded with
// Bulletin Board Systems (BBS) color codes to reconstruct them into HTML documents.
//
// BBSes were popular in the 1980s and 1990s, and allowed computer users to chat,
// message, and share files over the landline telephone network. The commercialization
// and ease of access to the Internet eventually replaced BBSes, as did the world-wide-web.
//
// These centralized systems, termed boards, used a text-based interface, and their
// owners often applied colorization, text themes, and art to differentiate themselves.
//
// While in the 1990s, ANSI control codes were in everyday use on the PC/MS-DOS, the
// standard comes from mainframe equipment. Home microcomputers often had difficulty
// interpreting it. So BBS developers created their own, more straightforward methods
// to colorize and theme the text output to solve this.
//
// *Please note, while many PC/MS-DOS boards used ANSI control codes for colorizations,
// this library does not support the standard.
//
// PCBoard
//
// One of the most well-known applications for hosting a PC/MS-DOS BBS, PCBoard
// pioneered the file_id.diz file descriptor, as well as being endlessly expandable
// through software plugins known as PPEs. It developed the popular @X color code and
// @ control syntax.
//
// Celerity
//
// Another PC/MS-DOS application that was very popular with the hacking, phreaking,
// and pirate communities in the early 1990s. It introduced a unique | pipe code
// syntax in late 1991 that revised the code syntax in version 2 of the software.
//
// Renegade
//
// A PC/MS-DOS application that was a derivative of the source code of Telegard BBS.
// Surprisingly there was a new release of this software in 2021. Renegade had two
// methods to implement color, and this library uses the Pipe Bar Color Codes.
//
// Telegard
//
// A PC/MS-DOS application became famous due to a source code leak or release by
// one of its authors back in an era when most developers were still highly
// secretive with their code. The source is incorporated into several other projects.
//
// WVIV
//
// A mainstay in the PC/MS-DOS BBS scene of the 1980s and early 1990s, it became well
// known for releasing its source code to registered users. It allowed them to expand
// the code to incorporate additional software such as games or utilities and port it
// to other platforms. The source is now Open Source and is still updated.
// Confusingly WWIV has three methods of colorizing text, 10 Pipe colors, two-digit
// pipe colors, and its original Heart Codes.
//
// Wildcat
//
// WILDCAT! was a popular, propriety PC/MS-DOS application from the late 1980s that
// later migrated to Windows. It was one of the few BBS applications that sold at
// retail in a physical box. It extensively used @ color codes throughout later
// revisions of its software.
//
package bbs

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/bengarrett/bbs/internal/split"
)

var (
	ErrColorCodes = errors.New("no bbs color codes found")
	ErrANSI       = errors.New("ansi escape code found")

	//go:embed static/*
	static embed.FS
)

// Bulletin Board System color code format.
// Other than for Find, the ANSI type is not supported by this library.
type BBS int

const (
	ANSI      BBS = iota // ANSI escape sequences.
	Celerity             // Celerity BBS pipe codes.
	PCBoard              // PCBoard BBS @ codes.
	Renegade             // Renegade BBS pipe codes.
	Telegard             // Telegard BBS grave accent codes.
	Wildcat              // Wildcat! BBS @ codes.
	WWIVHash             // WWIV BBS # codes.
	WWIVHeart            // WWIV BBS ♥ codes.
)

const (
	// ClearCmd is a PCBoard specific control to clear the screen that's occasionally found in ANSI text.
	ClearCmd string = "@CLS@"

	// CelerityMatch is a regular expression to match Celerity BBS color codes.
	CelerityMatch string = `\|(k|b|g|c|r|m|y|w|d|B|G|C|R|M|Y|W|S)`

	// PCBoardMatch is a case-insensitive, regular expression to match PCBoard BBS color codes.
	PCBoardMatch string = "(?i)@X([0-9A-F][0-9A-F])"

	// RenegadeMatch is a regular expression to match Renegade BBS color codes.
	RenegadeMatch string = `\|(0[0-9]|1[1-9]|2[0-3])`

	// TelegardMatch is a case-insensitive, regular expression to match Telegard BBS color codes.
	TelegardMatch string = "(?i)`([0-9|A-F])([0-9|A-F])"

	// WildcatMatch is a case-insensitive, regular expression to match Wildcat! BBS color codes.
	WildcatMatch string = `(?i)@([0-9|A-F])([0-9|A-F])@`

	// WWIVHashMatch is a regular expression to match WWIV BBS # color codes.
	WWIVHashMatch string = `\|#(\d)`

	// WWIVHeartMatch is a regular expression to match WWIV BBS ♥ color codes.
	WWIVHeartMatch string = `\x03(\d)`

	celerityCodes = "kbgcrmywdBGCRMYWS"
)

// HTMLCelerity writes to dst the HTML equivalent of Celerity BBS color codes with
// matching CSS color classes.
func HTMLCelerity(dst *bytes.Buffer, src []byte) error {
	return split.HTMLCelerity(dst, src)
}

// HTMLRenegade writes to dst the HTML equivalent of Renegade BBS color codes with
// matching CSS color classes.
func HTMLRenegade(dst *bytes.Buffer, src []byte) error {
	return split.HTMLBars(dst, src)
}

// HTMLPCBoard writes to dst the HTML equivalent of PCBoard BBS color codes with
// matching CSS color classes.
func HTMLPCBoard(dst *bytes.Buffer, src []byte) error {
	return split.HTMLPCBoard(dst, src)
}

// HTMLTelegard writes to dst the HTML equivalent of Telegard BBS color codes with
// matching CSS color classes.
func HTMLTelegard(dst *bytes.Buffer, src []byte) error {
	r := regexp.MustCompile(TelegardMatch)
	x := r.ReplaceAll(src, []byte(`@X$1$2`))
	return split.HTMLPCBoard(dst, x)
}

// HTMLWildcat writes to dst the HTML equivalent of Wildcat! BBS color codes with
// matching CSS color classes.
func HTMLWildcat(dst *bytes.Buffer, src []byte) error {
	r := regexp.MustCompile(WildcatMatch)
	x := r.ReplaceAll(src, []byte(`@X$1$2`))
	return split.HTMLPCBoard(dst, x)
}

// HTMLWildcat writes to dst the HTML equivalent of WWIV BBS # color codes with
// matching CSS color classes.
func HTMLWHash(dst *bytes.Buffer, src []byte) error {
	r := regexp.MustCompile(WWIVHashMatch)
	x := r.ReplaceAll(src, []byte(`|0$1`))
	return split.HTMLBars(dst, x)
}

// HTMLWildcat writes to dst the HTML equivalent of WWIV BBS ♥ color codes with
// matching CSS color classes.
func HTMLWHeart(dst *bytes.Buffer, src []byte) error {
	r := regexp.MustCompile(WWIVHeartMatch)
	x := r.ReplaceAll(src, []byte(`|0$1`))
	return split.HTMLBars(dst, x)
}

// HasCelerity reports if the bytes contains Celerity BBS color codes.
// The format uses the vertical bar "|" followed by a case sensitive single alphabetic character.
func HasCelerity(src []byte) bool {
	// celerityCodes contains all the character sequences for Celerity.
	for _, code := range []byte(celerityCodes) {
		if bytes.Contains(src, []byte{Celerity.Bytes()[0], code}) {
			return true
		}
	}
	return false
}

// HasPCBoard reports if the bytes contains PCBoard BBS color codes.
// The format uses an "@X" prefix with a background and foreground, 4-bit hexadecimal color value.
func HasPCBoard(src []byte) bool {
	const first, last = 0, 15
	const hexxed = "%X%X"
	for bg := first; bg <= last; bg++ {
		for fg := first; fg <= last; fg++ {
			subslice := []byte(fmt.Sprintf(hexxed, bg, fg))
			subslice = append(PCBoard.Bytes(), subslice...)
			if bytes.Contains(src, subslice) {
				return true
			}
		}
	}
	return false
}

// HasRenegade reports if the bytes contains Renegade BBS color codes.
// The format uses the vertical bar "|" followed by a padded, numeric value between 00 and 23.
func HasRenegade(src []byte) bool {
	const first, last = 0, 23
	const leadingZero = "%01d"
	for i := first; i <= last; i++ {
		subslice := []byte(fmt.Sprintf(leadingZero, i))
		subslice = append(Renegade.Bytes(), subslice...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// HasTelegard reports if the bytes contains Telegard BBS color codes.
// The format uses the grave accent followed by a padded, numeric value between 00 and 23.
func HasTelegard(src []byte) bool {
	const first, last = 0, 23
	const leadingZero = "%01d"
	for i := first; i <= last; i++ {
		subslice := []byte(fmt.Sprintf(leadingZero, i))
		subslice = append(Telegard.Bytes(), subslice...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// HasWildcat reports if the bytes contains Wildcat! BBS color codes.
// The format uses an a background and foreground,
// 4-bit hexadecimal color value enclosed by two at "@" characters.
func HasWildcat(src []byte) bool {
	const first, last = 0, 15
	for bg := first; bg <= last; bg++ {
		for fg := first; fg <= last; fg++ {
			subslice := []byte(fmt.Sprintf("%s%X%X%s",
				Wildcat.Bytes(), bg, fg, Wildcat.Bytes()))
			if bytes.Contains(src, subslice) {
				return true
			}
		}
	}
	return false
}

// HasWHash reports if the bytes contains WWIV BBS # (hash or pound) color codes.
// The format uses a vertical bar "|" with the hash "#" characters
// as a prefix with a numeric value between 0 and 9.
func HasWHash(src []byte) bool {
	const first, last = 0, 9
	for i := first; i <= last; i++ {
		subslice := append(WWIVHash.Bytes(), []byte(strconv.Itoa(i))...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// HasWHeart reports if the bytes contains WWIV BBS ♥ (heart) color codes.
// The format uses the ETX character as a prefix with a numeric value between 0 and 9.
// In the standard MS-DOS, USA codepage (CP-437), the ETX (end-of-text)
// character is substituted with a heart character.
func HasWHeart(src []byte) bool {
	const first, last = 0, 9
	for i := first; i <= last; i++ {
		subslice := append(WWIVHeart.Bytes(), []byte(strconv.Itoa(i))...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// TrimControls removes common PCBoard BBS controls prefixes from the bytes.
func TrimControls(src []byte) []byte {
	r := regexp.MustCompile(`@(CLS|CLS |PAUSE)@`)
	return r.ReplaceAll(src, []byte(""))
}

// Fields splits the io.Reader around the first instance of one or more consecutive BBS color codes.
// An error is returned if no color codes are found or if ANSI control sequences are first found.
func Fields(src io.Reader) ([]string, BBS, error) {
	var r1 bytes.Buffer
	r2 := io.TeeReader(src, &r1)
	f := Find(r2)
	if !f.Valid() {
		return nil, -1, ErrColorCodes
	}
	b, err := io.ReadAll(&r1)
	if err != nil {
		return nil, -1, err
	}
	switch f {
	case ANSI:
		return nil, -1, ErrANSI
	case Celerity:
		return split.Celerity(b), f, nil
	case PCBoard, Telegard, Wildcat:
		return split.PCBoard(b), f, nil
	case Renegade, WWIVHash, WWIVHeart:
		return split.Bars(b), f, nil
	}
	return nil, -1, ErrColorCodes
}

// Find the format of any known BBS color code sequence within the reader.
// If no sequences are found -1 is returned.
func Find(src io.Reader) BBS {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		b := scanner.Bytes()
		ts := bytes.TrimSpace(b)
		if ts == nil {
			continue
		}
		const l = len(ClearCmd)
		if len(ts) > l {
			if bytes.Equal(ts[0:l], []byte(ClearCmd)) {
				b = ts[l:]
			}
		}
		switch {
		case bytes.Contains(b, ANSI.Bytes()):
			return ANSI
		case bytes.Contains(b, Celerity.Bytes()):
			if HasRenegade(b) {
				return Renegade
			}
			if HasCelerity(b) {
				return Celerity
			}
			return -1
		case HasPCBoard(b):
			return PCBoard
		case HasTelegard(b):
			return Telegard
		case HasWildcat(b):
			return Wildcat
		case HasWHash(b):
			return WWIVHash
		case HasWHeart(b):
			return WWIVHeart
		}
	}
	return -1
}

// HTML writes to dst the HTML equivalent of BBS color codes with matching CSS color classes.
// The first found color code format is used for the remainder of the Reader.
func HTML(dst *bytes.Buffer, src io.Reader) (BBS, error) {
	var r1 bytes.Buffer

	r2 := io.TeeReader(src, &r1)

	find := Find(r2)
	b, err := io.ReadAll(&r1)
	if err != nil {
		return -1, err
	}
	return find, find.HTML(dst, b)
}

// Bytes returns the BBS color toggle sequence as bytes.
func (b BBS) Bytes() []byte {
	const (
		etx               byte = 3  // CP437 ♥
		esc               byte = 27 // CP437 ←
		hash                   = byte('#')
		atSign                 = byte('@')
		grave                  = byte('`')
		leftSquareBracket      = byte('[')
		verticalBar            = byte('|')
		upperX                 = byte('X')
	)
	switch b {
	case ANSI:
		return []byte{esc, leftSquareBracket}
	case Celerity, Renegade:
		return []byte{verticalBar}
	case PCBoard:
		return []byte{atSign, upperX}
	case Telegard:
		return []byte{grave}
	case Wildcat:
		return []byte{atSign}
	case WWIVHash:
		return []byte{verticalBar, hash}
	case WWIVHeart:
		return []byte{etx}
	default:
		return nil
	}
}

// CSS writes to dst the Cascading Style Sheets classes needed by the HTML.
// The CSS relies on cascading variables.
// See https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties for details.
func (b BBS) CSS(dst *bytes.Buffer) error {
	r, err := static.ReadFile("static/css/text_pcboard.css")
	if err != nil {
		return err
	}
	dst.Write(r)
	return nil
}

// HTML writes to dst the HTML equivalent of BBS color codes with matching CSS color classes.
func (b BBS) HTML(dst *bytes.Buffer, src []byte) error {
	x := TrimControls(src)
	switch b {
	case ANSI:
		return ErrANSI
	case Celerity:
		return HTMLCelerity(dst, x)
	case PCBoard:
		return HTMLPCBoard(dst, x)
	case Renegade:
		return HTMLRenegade(dst, x)
	case Telegard:
		return HTMLTelegard(dst, x)
	case Wildcat:
		return HTMLWildcat(dst, x)
	case WWIVHash:
		return HTMLWHash(dst, x)
	case WWIVHeart:
		return HTMLWHeart(dst, x)
	default:
		return ErrColorCodes
	}
}

// Name returns the name of the BBS color format.
func (b BBS) Name() string {
	if !b.Valid() {
		return ""
	}
	return [...]string{
		"ANSI",
		"Celerity",
		"PCBoard",
		"Renegade",
		"Telegard",
		"Wildcat!",
		"WWIV #",
		"WWIV ♥",
	}[b]
}

// Remove the BBS color codes from src and write it to dst.
func (b BBS) Remove(dst *bytes.Buffer, src []byte) error {
	switch b {
	case ANSI:
		return ErrANSI
	case Celerity:
		return remove(dst, src, CelerityMatch)
	case PCBoard:
		return remove(dst, src, PCBoardMatch)
	case Renegade:
		return remove(dst, src, RenegadeMatch)
	case Telegard:
		return remove(dst, src, TelegardMatch)
	case Wildcat:
		return remove(dst, src, WildcatMatch)
	case WWIVHash:
		return remove(dst, src, WWIVHashMatch)
	case WWIVHeart:
		return remove(dst, src, WWIVHeartMatch)
	default:
		return ErrColorCodes
	}
}

func remove(dst *bytes.Buffer, src []byte, expr string) error {
	m := regexp.MustCompile(expr)
	res := m.ReplaceAll(src, []byte(""))
	_, err := dst.Write(res)
	return err
}

// String returns the BBS color format name and toggle sequence.
func (b BBS) String() string {
	if !b.Valid() {
		return ""
	}
	return [...]string{
		"ANSI ←[",
		"Celerity |",
		"PCBoard @X",
		"Renegade |",
		"Telegard `",
		"Wildcat! @@",
		"WWIV |#",
		"WWIV ♥",
	}[b]
}

// Valid reports whether the BBS type is valid.
func (b BBS) Valid() bool {
	switch b {
	case ANSI,
		Celerity,
		PCBoard,
		Renegade,
		Telegard,
		Wildcat,
		WWIVHash,
		WWIVHeart:
		return true
	default:
		return false
	}
}
