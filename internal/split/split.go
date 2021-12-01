package split

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

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
	// BarsMatch is a regular expression to match Renegade BBS color codes.
	BarsMatch string = `\|(0[0-9]|1[1-9]|2[0-3])`

	// CelerityMatch is a regular expression to match Celerity BBS color codes.
	CelerityMatch string = `\|(k|b|g|c|r|m|y|w|d|B|G|C|R|M|Y|W|S)`

	// PCBoardMatch is a case-insensitive, regular expression to match PCBoard BBS color codes.
	PCBoardMatch string = "(?i)@X([0-9A-F][0-9A-F])"
)

// Bars slices a string into substrings separated by "|" vertical bar codes.
// The first two bytes of each substring will contain a colour value.
// Vertical bar codes are used by Renegade, WWIV hash and WWIV heart formats.
// An empty slice is returned when no valid bar code values exists.
func Bars(src []byte) []string {
	const sep rune = 65535
	m := regexp.MustCompile(BarsMatch)
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

// HTMLBars parses the string for BBS color codes that use
// vertical bar prefixes to apply a HTML template.
func HTMLBars(dst *bytes.Buffer, src []byte) error {
	const idiomaticTpl = `<i class="P{{.Background}} P{{.Foreground}}">{{.Content}}</i>`
	tmpl, err := template.New("idomatic").Parse(idiomaticTpl)
	if err != nil {
		return err
	}

	d := colorInt{}
	bars := Bars(src)
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
	if n < 16 {
		return false
	}
	if n > 23 {
		return false
	}
	return true
}

func barForeground(n int) bool {
	if n < 0 {
		return false
	}
	if n > 15 {
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
	m := regexp.MustCompile(CelerityMatch)
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

// HTMLCelerity parses the string for the unique Celerity BBS color codes
// to apply a HTML template.
func HTMLCelerity(dst *bytes.Buffer, src []byte) error {
	const idiomaticTpl, swapCmd = `<i class="PB{{.Background}} PF{{.Foreground}}">{{.Content}}</i>`, "S"
	tmpl, err := template.New("idomatic").Parse(idiomaticTpl)
	if err != nil {
		return err
	}

	background := false
	d := colorStr{
		Foreground: "w",
		Background: "k",
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
	m := regexp.MustCompile(PCBoardMatch)
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

// HTMLPCBoard parses the string for the common PCBoard BBS color codes
// to apply a HTML template.
func HTMLPCBoard(dst *bytes.Buffer, src []byte) error {
	const idiomaticTpl = `<i class="PB{{.Background}} PF{{.Foreground}}">{{.Content}}</i>`
	tmpl, err := template.New("idomatic").Parse(idiomaticTpl)
	if err != nil {
		return err
	}

	d := colorStr{}
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
