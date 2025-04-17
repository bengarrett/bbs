// Package bbs is a Go module that interacts with legacy textfiles encoded with
// [Bulletin Board Systems] (BBS) color codes to reconstruct them into HTML documents.
//
// BBSes were popular in the 1980s and 1990s and allowed computer users to
// chat, message, and share files over the landline telephone network. The
// commercialization and ease of access to the Internet eventually replaced BBSes,
// as did the worldwide-web. These centralized systems, termed boards, used a text-based
// interface, and their owners often applied colorization, text themes, and art to
// differentiate themselves.
//
// While in the 1990s, [ANSI control codes] were in everyday use on the PC/MS-DOS,
// the standard comes from mainframe equipment. Home microcomputers often had
// difficulty interpreting it. So, BBS developers created their own, more straightforward
// methods to colorize and theme the text output to solve this.
//
// *Please note that many microcomputer, PC and MS-DOS based boards used ANSI control
// codes for colorizations that this library does not support.
//
// # PCBoard
//
// One of the most well-known applications for hosting a PC/MS-DOS BBS, PCBoard
// pioneered the file_id.diz file descriptor, as well as being endlessly expandable
// through software plugins known as PPEs. It developed the popular @X color code and
// @ control syntax.
//
// # Celerity
//
// Another PC/MS-DOS application that was very popular with the hacking, phreaking,
// and pirate communities in the early 1990s. It introduced a unique | pipe code
// syntax in late 1991 that revised the code syntax in version 2 of the software.
//
// # Renegade
//
// A PC/MS-DOS application that was a derivative of the source code of Telegard BBS.
// Surprisingly there was a new release of this software in 2021. Renegade had two
// methods to implement color, and this library uses the Pipe Bar Color Codes.
//
// # Telegard
//
// A PC/MS-DOS application became famous due to a source code leak or release by
// one of its authors back in an era when most developers were still highly
// secretive with their code. The source is incorporated into several other projects.
//
// # WVIV
//
// A mainstay in the PC/MS-DOS BBS scene of the 1980s and early 1990s, it became well
// known for releasing its source code to registered users. It allowed them to expand
// the code to incorporate additional software such as games or utilities and port it
// to other platforms. The source is now Open Source and is still updated.
// Confusingly WWIV has three methods of colorizing text, 10 Pipe colors, two-digit
// pipe colors, and its original Heart Codes.
//
// # Wildcat
//
// WILDCAT! was a popular, propriety PC/MS-DOS application from the late 1980s that
// later migrated to Windows. It was one of the few BBS applications that sold at
// retail in a physical box. It extensively used @ color codes throughout later
// revisions of its software.
//
// [Bulletin Board Systems]: https://spectrum.ieee.org/social-medias-dialup-ancestor-the-bulletin-board-system
// [ANSI control codes]: https://www.cse.psu.edu/~kxc104/class/cse472/09f/hw/hw7/vt100ansi.htm
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

// Generic text match errors.
// Errors returned can be tested against these errors using errors.Is.
var (
	ErrANSI = errors.New("ansi escape code found")
	ErrNone = errors.New("no bbs color code found")
)

// Syntax errors.
var (
	ErrBuff = errors.New("bytes buffer cannot be nil")
)

//go:embed static/*
var static embed.FS

// Regular expressions to match BBS color codes.
const (
	CelerityRe  string = `\|(k|b|g|c|r|m|y|w|d|B|G|C|R|M|Y|W|S)` // matches Celerity
	PCBoardRe   string = "(?i)@X([0-9A-F][0-9A-F])"              // matches PCBoard
	RenegadeRe  string = `\|(0[0-9]|1[1-9]|2[0-3])`              // matches Renegade
	TelegardRe  string = "(?i)`([0-9|A-F])([0-9|A-F])"           // matches Telegard
	WildcatRe   string = `(?i)@([0-9|A-F])([0-9|A-F])@`          // matches Wildcat!
	WWIVHashRe  string = `\|#(\d)`                               // matches WWIV with hashes #
	WWIVHeartRe string = `\x03(\d)`                              // matches WWIV with hearts ♥
)

// Clear is a PCBoard specific control to clear the screen that's occasionally found in ANSI text.
const (
	Clear string = "@CLS@"

	celerityCodes = "kbgcrmywdBGCRMYWS"
)

// CelerityHTML writes to buf the HTML equivalent of Celerity BBS color codes with
// matching CSS color classes.
func CelerityHTML(buf *bytes.Buffer, src ...byte) error {
	if err := split.CelerityHTML(buf, src); err != nil {
		return fmt.Errorf("celerity html: %w", err)
	}
	return nil
}

// RenegadeHTML writes to buf the HTML equivalent of Renegade BBS color codes with
// matching CSS color classes.
func RenegadeHTML(buf *bytes.Buffer, src ...byte) error {
	if err := split.VBarsHTML(buf, src); err != nil {
		return fmt.Errorf("renegade html: %w", err)
	}
	return nil
}

// WildcatHTML writes to buf the HTML equivalent of Wildcat! BBS color codes with
// matching CSS color classes.
func WildcatHTML(buf *bytes.Buffer, src ...byte) error {
	re := regexp.MustCompile(WildcatRe)
	p := re.ReplaceAll(src, []byte(`@X$1$2`))
	if err := split.PCBoardHTML(buf, p); err != nil {
		return fmt.Errorf("wildcat html: %w", err)
	}
	return nil
}

// IsCelerity reports if the bytes contains Celerity BBS color codes.
// The format uses the vertical bar (|) followed by a case sensitive single alphabetic character.
func IsCelerity(src []byte) bool {
	// celerityCodes contains all the character sequences for Celerity.
	for _, code := range []byte(celerityCodes) {
		if bytes.Contains(src, []byte{Celerity.Bytes()[0], code}) {
			return true
		}
	}
	return false
}

// IsPCBoard reports if the bytes contains PCBoard BBS color codes.
// The format uses an at-sign x (@X) prefix with a background and foreground, 4-bit hexadecimal color value.
func IsPCBoard(src []byte) bool {
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

// IsRenegade reports if the bytes contains Renegade BBS color codes.
// The format uses the vertical bar (|) followed by a padded, numeric value between 00 and 23.
func IsRenegade(src []byte) bool {
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

// IsTelegard reports if the bytes contains Telegard BBS color codes.
// The format uses the grave accent (`) followed by a padded, numeric value between 00 and 23.
func IsTelegard(src []byte) bool {
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

// IsWWIVHash reports if the bytes contains WWIV BBS hash color codes.
// The format uses a vertical bar (|) with the hash (#) characters
// as a prefix with a numeric value between 0 and 9.
func IsWWIVHash(src []byte) bool {
	const first, last = 0, 9
	for i := first; i <= last; i++ {
		subslice := append(WWIVHash.Bytes(), []byte(strconv.Itoa(i))...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// IsWWIVHeart reports if the bytes contains WWIV BBS heart (♥) color codes.
// The format uses the ETX (end-of-text) character as a prefix with a numeric value between 0 and 9.
//
// In the MS-DOS era, the common North American [CP-437 codepage] substituted the ETX character with a heart symbol.
//
// [CP-437 codepage]: https://en.wikipedia.org/wiki/Code_page_437
func IsWWIVHeart(src []byte) bool {
	const first, last = 0, 9
	for i := first; i <= last; i++ {
		subslice := append(WWIVHeart.Bytes(), []byte(strconv.Itoa(i))...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// IsWildcat reports if the bytes contains Wildcat! BBS color codes.
// The format uses an a background and foreground,
// 4-bit hexadecimal color value enclosed with two at-sign (@) characters.
func IsWildcat(src []byte) bool {
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

// PCBoardHTML writes to buf the HTML equivalent of PCBoard BBS color codes with
// matching CSS color classes.
func PCBoardHTML(buf *bytes.Buffer, src ...byte) error {
	if err := split.PCBoardHTML(buf, src); err != nil {
		return fmt.Errorf("pcboard html: %w", err)
	}
	return nil
}

// TelegardHTML writes to buf the HTML equivalent of Telegard BBS color codes with
// matching CSS color classes.
func TelegardHTML(buf *bytes.Buffer, src ...byte) error {
	re := regexp.MustCompile(TelegardRe)
	p := re.ReplaceAll(src, []byte(`@X$1$2`))
	if err := split.PCBoardHTML(buf, p); err != nil {
		return fmt.Errorf("telegard html: %w", err)
	}
	return nil
}

// TrimControls removes common PCBoard BBS controls prefixes from the bytes.
// It trims the @CLS@ prefix used to clear the screen and the @PAUSE@ prefix
// used to pause the display render.
func TrimControls(src ...byte) []byte {
	re := regexp.MustCompile(`@(CLS|CLS |PAUSE)@`)
	return re.ReplaceAll(src, []byte(""))
}

// WWIVHashHTML writes to buf the HTML equivalent of WWIV BBS hash (#) color codes with
// matching CSS color classes.
func WWIVHashHTML(buf *bytes.Buffer, src ...byte) error {
	re := regexp.MustCompile(WWIVHashRe)
	p := re.ReplaceAll(src, []byte(`|0$1`))
	if err := split.VBarsHTML(buf, p); err != nil {
		return fmt.Errorf("wwiv hash html: %w", err)
	}
	return nil
}

// WWIVHeartHTML writes to buf the HTML equivalent of WWIV BBS heart (♥) color codes with
// matching CSS color classes.
func WWIVHeartHTML(buf *bytes.Buffer, src ...byte) error {
	re := regexp.MustCompile(WWIVHeartRe)
	p := re.ReplaceAll(src, []byte(`|0$1`))
	if err := split.VBarsHTML(buf, p); err != nil {
		return fmt.Errorf("wwiv heart html: %w", err)
	}
	return nil
}

// A BBS (Bulletin Board System) color code format,
// other than for [Find], the [ANSI] BBS is not supported by this library.
type BBS int

// BBS codes and sequences.
const (
	ANSI      BBS = iota // ANSI escape sequence.
	Celerity             // Celerity pipe.
	PCBoard              // PCBoard @ sign.
	Renegade             // Renegade pipe.
	Telegard             // Telegard grave accent.
	Wildcat              // Wildcat! @ sign.
	WWIVHash             // WWIV # symbol.
	WWIVHeart            // WWIV ♥ symbol.
)

const none = -1

// Fields splits the io.Reader around the first instance of one or more consecutive BBS color codes.
// An error is returned if no color codes are found or if ANSI control sequences are first found.
func Fields(src io.Reader) ([]string, BBS, error) {
	buf := bytes.Buffer{}
	r := io.TeeReader(src, &buf)
	find := Find(r)
	if !find.Valid() {
		return nil, none, ErrNone
	}
	all, err := io.ReadAll(&buf)
	if err != nil {
		return nil, -none, fmt.Errorf("read all: %w", err)
	}
	switch find {
	case ANSI:
		return nil, none, ErrANSI
	case Celerity:
		return split.Celerity(all), find, nil
	case PCBoard, Telegard, Wildcat:
		return split.PCBoard(all), find, nil
	case Renegade, WWIVHash, WWIVHeart:
		return split.VBars(all), find, nil
	}
	return nil, none, ErrNone
}

// Find the format of any known BBS color code sequence within the reader.
// If no sequences are found -1 is returned.
func Find(r io.Reader) BBS {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		slice := scanner.Bytes()
		trim := bytes.TrimSpace(slice)
		if trim == nil {
			continue
		}
		const l = len(Clear)
		if len(trim) > l {
			if bytes.Equal(trim[0:l], []byte(Clear)) {
				slice = trim[l:]
			}
		}
		switch {
		case bytes.Contains(slice, ANSI.Bytes()):
			return ANSI
		case bytes.Contains(slice, Celerity.Bytes()):
			if IsRenegade(slice) {
				return Renegade
			}
			if IsCelerity(slice) {
				return Celerity
			}
			return none
		case IsPCBoard(slice):
			return PCBoard
		case IsTelegard(slice):
			return Telegard
		case IsWildcat(slice):
			return Wildcat
		case IsWWIVHash(slice):
			return WWIVHash
		case IsWWIVHeart(slice):
			return WWIVHeart
		}
	}
	return none
}

// HTML writes to buf the HTML equivalent of BBS color codes with matching CSS color classes.
// The first found color code format is used for the remainder of the Reader.
func HTML(buf *bytes.Buffer, src io.Reader) (BBS, error) {
	if buf == nil {
		return -1, ErrBuff
	}
	w := bytes.Buffer{}
	r := io.TeeReader(src, &w)
	find := Find(r)
	p, err := io.ReadAll(&w)
	if err != nil {
		return -1, fmt.Errorf("read all: %w", err)
	}
	return find, find.HTML(buf, p)
}

// Bytes returns the BBS color toggle sequence.
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

// CSS writes to buf the Cascading Style Sheets classes needed by the HTML.
//
// The CSS results rely on [custom properties] which are not supported by legacy browsers.
//
// [custom properties]: https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties.
func (b BBS) CSS(buf *bytes.Buffer) error {
	if buf == nil {
		return ErrBuff
	}
	p, err := static.ReadFile("static/css/text_pcboard.css")
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	if _, err = buf.Write(p); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// HTML writes to buf the BBS color codes as CSS color classes within HTML <i> elements.
func (b BBS) HTML(buf *bytes.Buffer, src []byte) error {
	if buf == nil {
		return ErrBuff
	}
	trim := TrimControls(src...)
	switch b {
	case ANSI:
		return ErrANSI
	case Celerity:
		return CelerityHTML(buf, trim...)
	case PCBoard:
		return PCBoardHTML(buf, trim...)
	case Renegade:
		return RenegadeHTML(buf, trim...)
	case Telegard:
		return TelegardHTML(buf, trim...)
	case Wildcat:
		return WildcatHTML(buf, trim...)
	case WWIVHash:
		return WWIVHashHTML(buf, trim...)
	case WWIVHeart:
		return WWIVHeartHTML(buf, trim...)
	default:
		return ErrNone
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

// Remove the BBS color codes from src and write it to buf.
func (b BBS) Remove(buf *bytes.Buffer, src ...byte) error {
	if buf == nil {
		return ErrBuff
	}
	switch b {
	case ANSI:
		return ErrANSI
	case Celerity:
		return remove(buf, src, CelerityRe)
	case PCBoard:
		return remove(buf, src, PCBoardRe)
	case Renegade:
		return remove(buf, src, RenegadeRe)
	case Telegard:
		return remove(buf, src, TelegardRe)
	case Wildcat:
		return remove(buf, src, WildcatRe)
	case WWIVHash:
		return remove(buf, src, WWIVHashRe)
	case WWIVHeart:
		return remove(buf, src, WWIVHeartRe)
	}
	return ErrNone
}

func remove(buf *bytes.Buffer, src []byte, expr string) error {
	if buf == nil {
		return ErrBuff
	}
	re := regexp.MustCompile(expr)
	p := re.ReplaceAll(src, []byte(""))
	if _, err := buf.Write(p); err != nil {
		return fmt.Errorf("write buffer: %w", err)
	}
	return nil
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
