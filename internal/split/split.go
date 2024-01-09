package split

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

var ErrBuff = errors.New("bytes buffer cannot be nil")

// colorInt template data for integer based color codes.
type colorInt struct {
	Background int
	Foreground int
	Content    string
}

// colorStr template data for string based color codes.
type colorStr struct {
	Background string
	Foreground string
	Content    string
}

const (
	// CelerityRe is a regular expression to match Celerity BBS color codes.
	CelerityRe string = `\|(k|b|g|c|r|m|y|w|d|B|G|C|R|M|Y|W|S)`

	// PCBoardRe is a case-insensitive, regular expression to match PCBoard BBS color codes.
	PCBoardRe string = "(?i)@X([0-9A-F][0-9A-F])"

	// VBarsRe is a regular expression to match Renegade BBS color codes.
	VBarsRe string = `\|(0[0-9]|1[1-9]|2[0-3])`
)

// VBars slices a string into substrings separated by "|" vertical bar codes.
// The first two bytes of each substring will contain a colour value.
// Vertical bar codes are used by Renegade, WWIV hash and WWIV heart formats.
// An empty slice is returned when no valid bar code values exists.
func VBars(src []byte) []string {
	const sep rune = 65535
	m := regexp.MustCompile(VBarsRe)
	repl := fmt.Sprintf("%s$1", string(sep))
	res := m.ReplaceAll(src, []byte(repl))
	if !bytes.ContainsRune(res, sep) {
		return nil
	}

	spl := bytes.Split(res, []byte(string(sep)))
	app := []string{}
	for _, val := range spl {
		if len(val) == 0 {
			continue
		}
		app = append(app, string(val))
	}
	return app
}

// VBarsHTML parses the string for BBS color codes that use
// vertical bar prefixes to apply a HTML template.
func VBarsHTML(dst *bytes.Buffer, src []byte) error {
	if dst == nil {
		return ErrBuff
	}
	const idiomaticTpl = `<i class="P{{.Background}} P{{.Foreground}}">{{.Content}}</i>`
	tmpl, err := template.New("idomatic").Parse(idiomaticTpl)
	if err != nil {
		return err
	}

	d := colorInt{
		Foreground: -1,
		Background: -1,
		Content:    "",
	}
	bars := VBars(src)
	if len(bars) == 0 {
		_, err := dst.Write(src)
		return err
	}

	for _, color := range bars {
		n, err := strconv.Atoi(color[0:2])
		if err != nil {
			continue
		}
		if barForeground(n) {
			d.Foreground = n
		}
		if barBackground(n) {
			d.Background = n
		}
		d.Content = color[2:]
		if err := tmpl.Execute(dst, d); err != nil {
			return err
		}
	}
	return nil
}

func barBackground(n int) bool {
	const first, last = 16, 23
	if n < first {
		return false
	}
	if n > last {
		return false
	}
	return true
}

func barForeground(n int) bool {
	const first, last = 0, 15
	if n < first {
		return false
	}
	if n > last {
		return false
	}
	return true
}

// Celerity slices a string into substrings separated by "|" vertical bar codes.
// The first byte of each substring will contain a Celerity colour value,
// that are comprised of a single, alphabetic character.
// An empty slice is returned when no valid Celerity code values exists.
func Celerity(src []byte) []string {
	// The format uses the vertical bar "|" followed by a case sensitive single alphabetic character.
	const sep rune = 65535
	m := regexp.MustCompile(CelerityRe)
	repl := fmt.Sprintf("%s$1", string(sep))
	res := m.ReplaceAll(src, []byte(repl))
	if !bytes.ContainsRune(res, sep) {
		return []string{}
	}

	spl := bytes.Split(res, []byte(string(sep)))
	clean := []string{}
	for _, val := range spl {
		if len(val) == 0 {
			continue
		}
		clean = append(clean, string(val))
	}
	return clean
}

// CelerityHTML parses the string for the unique Celerity BBS color codes
// to apply a HTML template.
func CelerityHTML(dst *bytes.Buffer, src []byte) error {
	if dst == nil {
		return ErrBuff
	}
	const idiomaticTpl, swapCmd = `<i class="PB{{.Background}} PF{{.Foreground}}">{{.Content}}</i>`, "S"
	tmpl, err := template.New("idomatic").Parse(idiomaticTpl)
	if err != nil {
		return err
	}

	background := false
	d := colorStr{
		Foreground: "w",
		Background: "k",
		Content:    "",
	}

	bars := Celerity(src)
	if len(bars) == 0 {
		_, err := dst.Write(src)
		return err
	}
	for _, color := range bars {
		if color == swapCmd {
			background = !background
			continue
		}
		if !background {
			d.Foreground = string(color[0])
		}
		if background {
			d.Background = string(color[0])
		}
		d.Content = color[1:]
		if err := tmpl.Execute(dst, d); err != nil {
			return err
		}
	}
	return nil
}

// PCBoard slices a string into substrings separated by PCBoard @X codes.
// The first two bytes of each substring will contain background
// and foreground hex colour values.
// An empty slice is returned when no valid @X code values exists.
func PCBoard(src []byte) []string {
	const sep rune = 65535
	m := regexp.MustCompile(PCBoardRe)
	repl := fmt.Sprintf("%s$1", string(sep))
	res := m.ReplaceAll(src, []byte(repl))
	if !bytes.ContainsRune(res, sep) {
		return []string{}
	}

	spl := bytes.Split(res, []byte(string(sep)))
	clean := []string{}
	for _, val := range spl {
		if len(val) == 0 {
			continue
		}
		clean = append(clean, string(val))
	}
	return clean
}

// PCBoardHTML parses the string for the common PCBoard BBS color codes
// to apply a HTML template.
func PCBoardHTML(dst *bytes.Buffer, src []byte) error {
	if dst == nil {
		return ErrBuff
	}
	const idiomaticTpl = `<i class="PB{{.Background}} PF{{.Foreground}}">{{.Content}}</i>`
	tmpl, err := template.New("idomatic").Parse(idiomaticTpl)
	if err != nil {
		return err
	}

	d := colorStr{
		Foreground: "",
		Background: "",
		Content:    "",
	}
	xcodes := PCBoard(src)
	if len(xcodes) == 0 {
		_, err := dst.Write(src)
		return err
	}
	for _, color := range xcodes {
		d.Background = strings.ToUpper(string(color[0]))
		d.Foreground = strings.ToUpper(string(color[1]))
		d.Content = color[2:]
		if err := tmpl.Execute(dst, d); err != nil {
			return err
		}
	}
	return nil
}
