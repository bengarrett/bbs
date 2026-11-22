package bbs_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bengarrett/bbs"
)

func TestBBS_DocumentOptions(t *testing.T) {
	var doc bytes.Buffer
	err := bbs.PCBoard.DocumentOptions(&doc, []byte("@X1EHello\r\n@X4CWorld"), bbs.Options{})
	if err != nil {
		t.Fatalf("DocumentOptions() error = %v", err)
	}

	got := doc.String()
	for _, want := range []string{
		"<!doctype html>",
		`<meta charset="UTF-8"`,
		"<style>",
		`@font-face`,
		`data:font/woff2;base64,`,
		`font-family: "IBM VGA 8x16"`,
		"--black: rgb(0, 0, 0);",
		"--blue: rgb(0, 0, 170);",
		"--lightblue: rgb(85, 85, 255);",
		"@keyframes blinkingblack",
		"i.PF0",
		"i.PFB {\n  color: var(--lightcyan);\n}",
		"i.PFC {\n  color: var(--lightred);\n}",
		"<pre>",
		`<i class="PB1 PFE">Hello`,
		`</i><i class="PB4 PFC">World</i>`,
		"</pre>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("DocumentOptions output missing %q", want)
		}
	}
	if strings.Contains(got, "@import ") || strings.Contains(got, "i.PFB { color: var(--lightblue); }") {
		t.Errorf("DocumentOptions output contains invalid or conflicting rules")
	}
}
