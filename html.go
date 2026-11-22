package bbs

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

const vgaWOFF2Path = "static/font/Ac437_IBM_VGA_8x16.woff2"

const standaloneLayoutCSS = `
html {
  color-scheme: dark;
}

body {
  margin: 0;
  padding: 16px;
  background-color: var(--black);
  color: var(--grey);
}

pre {
  margin: 0;
  font-family: "IBM VGA 8x16", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 16px;
  line-height: 16px;
  font-style: normal;
  font-weight: 400;
  white-space: pre;
}
`

const standaloneVBarsCSS = `
i.P0 { color: var(--black); }
i.P1 { color: var(--blue); }
i.P2 { color: var(--green); }
i.P3 { color: var(--cyan); }
i.P4 { color: var(--red); }
i.P5 { color: var(--magenta); }
i.P6 { color: var(--brown); }
i.P7 { color: var(--grey); }
i.P8 { color: var(--darkgrey); }
i.P9 { color: var(--lightblue); }
i.P10 { color: var(--lightgreen); }
i.P11 { color: var(--lightcyan); }
i.P12 { color: var(--lightred); }
i.P13 { color: var(--lightmagenta); }
i.P14 { color: var(--yellow); }
i.P15 { color: var(--white); }

i.P16 { background-color: var(--black); }
i.P17 { background-color: var(--blue); }
i.P18 { background-color: var(--green); }
i.P19 { background-color: var(--cyan); }
i.P20 { background-color: var(--red); }
i.P21 { background-color: var(--magenta); }
i.P22 { background-color: var(--brown); }
i.P23 { background-color: var(--grey); }
`

const standaloneCelerityCSS = `
i.PFk { color: var(--black); }
i.PFb { color: var(--blue); }
i.PFg { color: var(--green); }
i.PFc { color: var(--cyan); }
i.PFr { color: var(--red); }
i.PFm { color: var(--magenta); }
i.PFy { color: var(--brown); }
i.PFw { color: var(--grey); }
i.PFd { color: var(--darkgrey); }
i.PFB { color: var(--lightblue); }
i.PFG { color: var(--lightgreen); }
i.PFC { color: var(--lightcyan); }
i.PFR { color: var(--lightred); }
i.PFM { color: var(--lightmagenta); }
i.PFY { color: var(--yellow); }
i.PFW { color: var(--white); }

i.PBk { background-color: transparent; }
i.PBb { background-color: var(--blue); }
i.PBg { background-color: var(--green); }
i.PBc { background-color: var(--cyan); }
i.PBr { background-color: var(--red); }
i.PBm { background-color: var(--magenta); }
i.PBy { background-color: var(--brown); }
i.PBw { background-color: var(--grey); }
`

// Document writes to buf a self-contained HTML5 document for src with embedded CSS.
func (b BBS) Document(buf *bytes.Buffer, src []byte) error {
	return b.DocumentOptions(buf, src, Options{})
}

// DocumentOptions writes to buf a self-contained HTML5 document for src using opts.
func (b BBS) DocumentOptions(buf *bytes.Buffer, src []byte, opts Options) error {
	if buf == nil {
		return ErrBuff
	}

	var body bytes.Buffer
	if err := b.HTMLOptions(&body, src, opts); err != nil {
		return err
	}

	css, err := b.StandaloneCSS()
	if err != nil {
		return err
	}

	buf.WriteString("<!doctype html>\n<html lang=\"en\">\n  <head>\n")
	buf.WriteString("    <meta charset=\"UTF-8\" />\n")
	buf.WriteString("    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\" />\n")
	buf.WriteString("    <style>\n")
	buf.WriteString(css)
	buf.WriteString(standaloneLayoutCSS)
	buf.WriteString("    </style>\n  </head>\n  <body>\n<pre>")
	buf.Write(body.Bytes())
	buf.WriteString("</pre>\n  </body>\n</html>\n")
	return nil
}

// StandaloneCSS returns a self-contained CSS stylesheet for b, including the
// base64-encoded IBM VGA 8x16 webfont, color variables, blink animations, and
// format-specific color classes.
func (b BBS) StandaloneCSS() (string, error) {
	files := []string{
		"static/css/text_bbs.css",
		"static/css/text_blink.css",
	}
	base, _ := b.normalize(nil)
	if !base.Valid() || base == ANSI {
		base = PCBoard
	}
	if base == PCBoard {
		files = append(files, "static/css/text_pcboard.css")
	}

	var out strings.Builder
	fontBytes, err := static.ReadFile(vgaWOFF2Path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", vgaWOFF2Path, err)
	}
	out.WriteString("@font-face {\n  font-family: \"IBM VGA 8x16\";\n")
	out.WriteString("  font-style: normal;\n  font-weight: 400;\n")
	out.WriteString("  src: url(\"data:font/woff2;base64,")
	out.WriteString(base64.StdEncoding.EncodeToString(fontBytes))
	out.WriteString("\") format(\"woff2\");\n}\n\n")

	for _, name := range files {
		raw, err := static.ReadFile(name)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", name, err)
		}
		out.WriteString(stripCSSImports(string(raw)))
		out.WriteByte('\n')
	}
	switch base {
	case Renegade:
		out.WriteString(standaloneVBarsCSS)
	case Celerity:
		out.WriteString(standaloneCelerityCSS)
	case ANSI, PCBoard, Wildcat, Telegard, WWIVHash, WWIVHeart:
	}
	return out.String(), nil
}

func stripCSSImports(css string) string {
	lines := strings.Split(css, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "@import ") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.TrimSpace(strings.Join(kept, "\n")) + "\n"
}
