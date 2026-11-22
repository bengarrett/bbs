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
//
//nolint:gochecknoglobals
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

	"github.com/bengarrett/bbs/internal/pattern"
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
	ErrBuff   = errors.New("bytes buffer cannot be nil")
	ErrReader = errors.New("reader cannot be nil")
)

//go:embed static/*
var static embed.FS

// Regular expressions to match BBS color codes.
const (
	CelerityRe  string = pattern.Celerity  // matches Celerity
	PCBoardRe   string = pattern.PCBoard   // matches PCBoard
	RenegadeRe  string = pattern.Renegade  // matches Renegade and Telegard pipes
	TelegardRe  string = pattern.Telegard  // matches Telegard graves
	WildcatRe   string = pattern.Wildcat   // matches Wildcat!
	WWIVHashRe  string = pattern.WWIVHash  // matches WWIV with hashes #
	WWIVHeartRe string = pattern.WWIVHeart // matches WWIV with hearts ♥
)

// Pre-compiled regular expressions for performance.
var (
	trimControlsRe = regexp.MustCompile(`@(CLS|CLS |PAUSE)@`)
	celerityRe     = pattern.CelerityRe
	pcboardRe      = pattern.PCBoardRe
	renegadeRe     = pattern.RenegadeRe
	telegardRe     = pattern.TelegardRe
	wildcatRe      = pattern.WildcatRe
	wwivHashRe     = pattern.WWIVHashRe
	wwivHeartRe    = pattern.WWIVHeartRe
)

// Clear is a PCBoard specific control to clear the screen that's occasionally found in ANSI text.
const (
	Clear string = "@CLS@"
)

const celerityCodes = "kbgcrmywdBGCRMYWS"

// CelerityANSI writes to buf the ANSI SGR equivalent of Celerity BBS color codes.
func CelerityANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Celerity, Columns, src...)
	if err := split.CelerityANSI(buf, wrapped); err != nil {
		return fmt.Errorf("celerity ansi: %w", err)
	}
	return nil
}

// CelerityHTML writes to buf the HTML equivalent of Celerity BBS color codes with
// matching CSS color classes.
func CelerityHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Celerity, Columns, src...)
	if err := split.CelerityHTML(buf, wrapped); err != nil {
		return fmt.Errorf("celerity html: %w", err)
	}
	return nil
}

// RenegadeANSI writes to buf the ANSI SGR equivalent of Renegade BBS color codes.
func RenegadeANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Renegade, Columns, src...)
	if err := split.VBarsANSI(buf, wrapped); err != nil {
		return fmt.Errorf("renegade ansi: %w", err)
	}
	return nil
}

// RenegadeHTML writes to buf the HTML equivalent of Renegade BBS color codes with
// matching CSS color classes.
func RenegadeHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Renegade, Columns, src...)
	if err := split.VBarsHTML(buf, wrapped); err != nil {
		return fmt.Errorf("renegade html: %w", err)
	}
	return nil
}

// WildcatANSI writes to buf the ANSI SGR equivalent of Wildcat! BBS color codes.
func WildcatANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Wildcat, Columns, src...)
	p := wildcatRe.ReplaceAll(wrapped, []byte(`@X$1$2`))
	if err := split.PCBoardANSI(buf, p); err != nil {
		return fmt.Errorf("wildcat ansi: %w", err)
	}
	return nil
}

// WildcatHTML writes to buf the HTML equivalent of Wildcat! BBS color codes with
// matching CSS color classes.
func WildcatHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Wildcat, Columns, src...)
	p := wildcatRe.ReplaceAll(wrapped, []byte(`@X$1$2`))
	if err := split.PCBoardHTML(buf, p); err != nil {
		return fmt.Errorf("wildcat html: %w", err)
	}
	return nil
}

// IsCelerity reports if the bytes contains Celerity BBS color codes.
// The format uses the vertical bar (|) followed by a case sensitive single alphabetic character.
func IsCelerity(src []byte) bool {
	// celerityCodes contains all the character sequences for Celerity.
	seq := Celerity.Bytes()
	if len(seq) < 1 {
		return false
	}

	for _, code := range []byte(celerityCodes) {
		if bytes.Contains(src, []byte{seq[0], code}) {
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
			subslice := fmt.Appendf(nil, hexxed, bg, fg)
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
	const leadingZero = "%02d"
	for i := first; i <= last; i++ {
		subslice := fmt.Appendf(nil, leadingZero, i)
		subslice = append(Renegade.Bytes(), subslice...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// IsTelegard reports if the bytes contains Telegard BBS color codes.
// The format uses the grave accent (`) followed by two hexadecimal digits (00 to FF).
func IsTelegard(src []byte) bool {
	return telegardRe.Match(src)
}

// IsWWIVHash reports if the bytes contains WWIV BBS hash color codes.
// The format uses a vertical bar (|) with the hash (#) characters
// as a prefix with a numeric value between 0 and 9.
func IsWWIVHash(src []byte) bool {
	const first, last = 0, 9
	for i := first; i <= last; i++ {
		subslice := append(WWIVHash.Bytes(), strconv.Itoa(i)...)
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
		subslice := append(WWIVHeart.Bytes(), strconv.Itoa(i)...)
		if bytes.Contains(src, subslice) {
			return true
		}
	}
	return false
}

// IsWildcat reports if the bytes contains Wildcat! BBS color codes.
// The format uses a background and foreground,
// 4-bit hexadecimal color value enclosed with two at-sign (@) characters.
func IsWildcat(src []byte) bool {
	const first, last = 0, 15
	const format = "%s%X%X%s"
	for bg := first; bg <= last; bg++ {
		for fg := first; fg <= last; fg++ {
			subslice := fmt.Appendf(nil, format, Wildcat.Bytes(), bg, fg, Wildcat.Bytes())
			if bytes.Contains(src, subslice) {
				return true
			}
		}
	}
	return false
}

// PCBoardANSI writes to buf the ANSI SGR equivalent of PCBoard BBS color codes.
func PCBoardANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(PCBoard, Columns, src...)
	if err := split.PCBoardANSI(buf, wrapped); err != nil {
		return fmt.Errorf("pcboard ansi: %w", err)
	}
	return nil
}

// PCBoardHTML writes to buf the HTML equivalent of PCBoard BBS color codes with
// matching CSS color classes.
func PCBoardHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(PCBoard, Columns, src...)
	if err := split.PCBoardHTML(buf, wrapped); err != nil {
		return fmt.Errorf("pcboard html: %w", err)
	}
	return nil
}

// TelegardANSI writes to buf the ANSI SGR equivalent of Telegard BBS color codes.
func TelegardANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Telegard, Columns, src...)
	p := telegardRe.ReplaceAll(wrapped, []byte(`@X$1$2`))
	if err := split.PCBoardANSI(buf, p); err != nil {
		return fmt.Errorf("telegard ansi: %w", err)
	}
	return nil
}

// TelegardHTML writes to buf the HTML equivalent of Telegard BBS color codes with
// matching CSS color classes.
func TelegardHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(Telegard, Columns, src...)
	p := telegardRe.ReplaceAll(wrapped, []byte(`@X$1$2`))
	if err := split.PCBoardHTML(buf, p); err != nil {
		return fmt.Errorf("telegard html: %w", err)
	}
	return nil
}

// TrimControls removes common PCBoard BBS controls prefixes from the bytes.
// It trims the @CLS@ prefix used to clear the screen and the @PAUSE@ prefix
// used to pause the display render.
func TrimControls(src ...byte) []byte {
	return trimControlsRe.ReplaceAll(src, []byte(""))
}

// WWIVHashANSI writes to buf the ANSI SGR equivalent of WWIV BBS hash (#) color codes.
func WWIVHashANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(WWIVHash, Columns, src...)
	p := wwivHashRe.ReplaceAll(wrapped, []byte(`|0$1`))
	if err := split.VBarsANSI(buf, p); err != nil {
		return fmt.Errorf("wwiv hash ansi: %w", err)
	}
	return nil
}

// WWIVHashHTML writes to buf the HTML equivalent of WWIV BBS hash (#) color codes with
// matching CSS color classes.
func WWIVHashHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(WWIVHash, Columns, src...)
	p := wwivHashRe.ReplaceAll(wrapped, []byte(`|0$1`))
	if err := split.VBarsHTML(buf, p); err != nil {
		return fmt.Errorf("wwiv hash html: %w", err)
	}
	return nil
}

// WWIVHeartANSI writes to buf the ANSI SGR equivalent of WWIV BBS heart (♥) color codes.
func WWIVHeartANSI(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(WWIVHeart, Columns, src...)
	p := wwivHeartRe.ReplaceAll(wrapped, []byte(`|0$1`))
	if err := split.VBarsANSI(buf, p); err != nil {
		return fmt.Errorf("wwiv heart ansi: %w", err)
	}
	return nil
}

// WWIVHeartHTML writes to buf the HTML equivalent of WWIV BBS heart (♥) color codes with
// matching CSS color classes.
func WWIVHeartHTML(buf *bytes.Buffer, src ...byte) error {
	wrapped := Wrap(WWIVHeart, Columns, src...)
	p := wwivHeartRe.ReplaceAll(wrapped, []byte(`|0$1`))
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
	Renegade             // Renegade pipe, despite the name these originate in Telegard.
	PCBoard              // PCBoard @ sign.
	Wildcat              // Wildcat! @ sign.
	Telegard             // Telegard grave accent.
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
		return nil, none, fmt.Errorf("read all: %w", err)
	}

	switch find {
	case ANSI:
		return nil, none, ErrANSI
	case Celerity:
		return split.Celerity(all), find, nil
	case PCBoard:
		return split.PCBoard(all), find, nil
	case Telegard:
		return split.PCBoard(telegardRe.ReplaceAll(all, []byte(`@X$1$2`))), find, nil
	case Wildcat:
		return split.PCBoard(wildcatRe.ReplaceAll(all, []byte(`@X$1$2`))), find, nil
	case Renegade:
		return split.VBars(all), find, nil
	case WWIVHash:
		return split.VBars(wwivHashRe.ReplaceAll(all, []byte(`|0$1`))), find, nil
	case WWIVHeart:
		return split.VBars(wwivHeartRe.ReplaceAll(all, []byte(`|0$1`))), find, nil
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
		case IsRenegade(slice):
			return Renegade
		case IsCelerity(slice):
			return Celerity
		case IsWWIVHash(slice):
			return WWIVHash
		case IsPCBoard(slice):
			return PCBoard
		case IsTelegard(slice):
			return Telegard
		case IsWildcat(slice):
			return Wildcat
		case IsWWIVHeart(slice):
			return WWIVHeart
		}
	}

	_ = scanner.Err() // any scan errors can be ignored
	return none
}

// ToANSI writes to buf the ANSI SGR equivalent of BBS color codes.
// The first found color code format is used for the remainder of the Reader.
func ToANSI(buf *bytes.Buffer, src io.Reader) (BBS, error) {
	return ToANSIOptions(buf, src, Options{})
}

// ToANSIOptions writes to buf the ANSI SGR equivalent of BBS color codes using opts.
func ToANSIOptions(buf *bytes.Buffer, src io.Reader, opts Options) (BBS, error) {
	p, find, err := findAndRead(buf, src)
	if err != nil {
		return find, err
	}
	return find, find.ANSIOptions(buf, p, opts)
}

// HTML writes to buf the HTML equivalent of BBS color codes with matching CSS color classes.
// The first found color code format is used for the remainder of the Reader.
func HTML(buf *bytes.Buffer, src io.Reader) (BBS, error) {
	return HTMLOptions(buf, src, Options{})
}

// HTMLOptions writes to buf the HTML equivalent of BBS color codes using opts.
func HTMLOptions(buf *bytes.Buffer, src io.Reader, opts Options) (BBS, error) {
	p, find, err := findAndRead(buf, src)
	if err != nil {
		return find, err
	}
	return find, find.HTMLOptions(buf, p, opts)
}

func findAndRead(buf *bytes.Buffer, src io.Reader) ([]byte, BBS, error) {
	if buf == nil {
		return nil, -1, ErrBuff
	}
	if src == nil {
		return nil, -1, ErrReader
	}

	w := bytes.Buffer{}
	r := io.TeeReader(src, &w)
	find := Find(r)

	p, err := io.ReadAll(&w)
	if err != nil {
		return nil, -1, fmt.Errorf("read all: %w", err)
	}
	return p, find, nil
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

	const name = "static/css/text_pcboard.css"
	p, err := static.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if _, err = buf.Write(p); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// ANSI writes to buf the BBS color codes as ANSI SGR escape sequences.
func (b BBS) ANSI(buf *bytes.Buffer, src []byte) error {
	return b.ANSIOptions(buf, src, Options{})
}

// ANSIOptions writes to buf the BBS color codes as ANSI SGR escape sequences using opts.
func (b BBS) ANSIOptions(buf *bytes.Buffer, src []byte, opts Options) error {
	return b.render(buf, src, opts, rendererSet{
		celerity: split.CelerityANSI,
		pcboard:  split.PCBoardANSI,
		vbars:    split.VBarsANSI,
	})
}

// HTML writes to buf the BBS color codes as CSS color classes within HTML <i> elements.
func (b BBS) HTML(buf *bytes.Buffer, src []byte) error {
	return b.HTMLOptions(buf, src, Options{})
}

// HTMLOptions writes to buf the BBS color codes as CSS color classes using opts.
func (b BBS) HTMLOptions(buf *bytes.Buffer, src []byte, opts Options) error {
	return b.render(buf, src, opts, rendererSet{
		celerity: split.CelerityHTML,
		pcboard:  split.PCBoardHTML,
		vbars:    split.VBarsHTML,
	})
}

// Name returns the name of the BBS color format.
func (b BBS) Name() string {
	if !b.Valid() {
		return ""
	}

	return [...]string{
		ANSI:      "ANSI",
		Celerity:  "Celerity",
		Renegade:  "Renegade",
		PCBoard:   "PCBoard",
		Wildcat:   "Wildcat!",
		Telegard:  "Telegard",
		WWIVHash:  "WWIV #",
		WWIVHeart: "WWIV ♥",
	}[b]
}

// Remove the BBS color codes from src and write it to buf.
func (b BBS) Remove(buf *bytes.Buffer, src ...byte) error {
	return b.RemoveOptions(buf, Options{}, src...)
}

// RemoveOptions removes the BBS color codes from src and writes it to buf using opts.
func (b BBS) RemoveOptions(buf *bytes.Buffer, opts Options, src ...byte) error {
	if buf == nil {
		return ErrBuff
	}

	wrapped := Wrap(b, opts.wrapWidth(), src...)
	switch b {
	case ANSI:
		return ErrANSI
	case Celerity:
		return remove(buf, wrapped, celerityRe)
	case PCBoard:
		return remove(buf, wrapped, pcboardRe)
	case Renegade:
		return remove(buf, wrapped, renegadeRe)
	case Telegard:
		return remove(buf, wrapped, telegardRe)
	case Wildcat:
		return remove(buf, wrapped, wildcatRe)
	case WWIVHash:
		return remove(buf, wrapped, wwivHashRe)
	case WWIVHeart:
		return remove(buf, wrapped, wwivHeartRe)
	default:
		return ErrNone
	}
}

func remove(buf *bytes.Buffer, src []byte, expr *regexp.Regexp) error {
	if buf == nil {
		return ErrBuff
	}

	p := expr.ReplaceAll(src, []byte(""))
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
		ANSI:      "ANSI ←[",
		Celerity:  "Celerity |",
		Renegade:  "Renegade |",
		PCBoard:   "PCBoard @X",
		Wildcat:   "Wildcat! @@",
		Telegard:  "Telegard `",
		WWIVHash:  "WWIV |#",
		WWIVHeart: "WWIV ♥",
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

type rendererSet struct {
	celerity func(*bytes.Buffer, []byte) error
	pcboard  func(*bytes.Buffer, []byte) error
	vbars    func(*bytes.Buffer, []byte) error
}

func (b BBS) normalize(src []byte) (BBS, []byte) {
	switch b {
	case Telegard:
		return PCBoard, telegardRe.ReplaceAll(src, []byte(`@X$1$2`))
	case Wildcat:
		return PCBoard, wildcatRe.ReplaceAll(src, []byte(`@X$1$2`))
	case WWIVHash:
		return Renegade, wwivHashRe.ReplaceAll(src, []byte(`|0$1`))
	case WWIVHeart:
		return Renegade, wwivHeartRe.ReplaceAll(src, []byte(`|0$1`))
	case ANSI, Celerity, Renegade, PCBoard:
		return b, src
	default:
		return b, src
	}
}

func (b BBS) render(
	buf *bytes.Buffer,
	src []byte,
	opts Options,
	set rendererSet,
) error {
	if buf == nil {
		return ErrBuff
	}
	if b == ANSI {
		return ErrANSI
	}
	if !b.Valid() {
		return ErrNone
	}

	trim := TrimControls(src...)
	wrapped := Wrap(b, opts.wrapWidth(), trim...)
	base, data := b.normalize(wrapped)
	switch base {
	case Celerity:
		return set.celerity(buf, data)
	case PCBoard:
		return set.pcboard(buf, data)
	case Renegade:
		return set.vbars(buf, data)
	case ANSI, Wildcat, Telegard, WWIVHash, WWIVHeart:
		return ErrNone
	default:
		return ErrNone
	}
}
