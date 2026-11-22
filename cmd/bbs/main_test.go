//nolint:cyclop
package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bengarrett/bbs"
)

func TestCLINoArgs(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := run(nil, &buf)
	if !errors.Is(err, ErrInput) {
		t.Fatalf("run() error = %v, want %v", err, ErrInput)
	}
	if !strings.Contains(buf.String(), "Usage:") {
		t.Errorf("expected Usage in output, got %q", buf.String())
	}
}

func TestCLIVersion(t *testing.T) {
	t.Parallel()
	for _, flag := range []string{"-version", "-v"} {
		var buf bytes.Buffer
		if err := run(nil, &buf, flag); err != nil {
			t.Fatalf("run(%s) error = %v", flag, err)
		}
		if !strings.Contains(buf.String(), "bbs ") {
			t.Errorf("run(%s) output = %q, want prefix 'bbs '", flag, buf.String())
		}
	}
}

func TestCLICheck(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	pcbFile := filepath.Join(tmpDir, "ad.pcb")
	if err := os.WriteFile(pcbFile, []byte("@CLS@@X1EHello"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := run(nil, &buf, "-check", pcbFile); err != nil {
		t.Fatalf("run(-check) error = %v", err)
	}
	if !strings.Contains(buf.String(), "PCBoard") {
		t.Errorf("expected PCBoard in output, got %q", buf.String())
	}

	plainFile := filepath.Join(tmpDir, "plain.txt")
	if err := os.WriteFile(plainFile, []byte("Just plain text"), 0o600); err != nil {
		t.Fatal(err)
	}
	var plainBuf bytes.Buffer
	if err := run(nil, &plainBuf, "-c", plainFile); err != nil {
		t.Fatalf("run(-c) error = %v", err)
	}
	if !strings.Contains(plainBuf.String(), "Plain text") {
		t.Errorf("expected Plain text in output, got %q", plainBuf.String())
	}
}

func TestCLIConvertFormatsAndIO(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	pcbFile := filepath.Join(tmpDir, "ad.pcb")
	if err := os.WriteFile(pcbFile, []byte("@CLS@@X1E\xdbHello@PAUSE@"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Default: ANSI to stdout (CP437 preserved)
	var ansiBuf bytes.Buffer
	if err := run(nil, &ansiBuf, pcbFile); err != nil {
		t.Fatalf("run(ansi stdout) error = %v", err)
	}
	if want := "\x1b[0;1;44;33m\xdbHello"; ansiBuf.String() != want {
		t.Errorf("ansi stdout = %q, want %q", ansiBuf.String(), want)
	}

	// CP437 -> UTF-8 flag
	var utf8Buf bytes.Buffer
	if err := run(nil, &utf8Buf, "-utf8", pcbFile); err != nil {
		t.Fatalf("run(-utf8) error = %v", err)
	}
	if want := "\x1b[0;1;44;33m█Hello"; utf8Buf.String() != want {
		t.Errorf("ansi utf8 = %q, want %q", utf8Buf.String(), want)
	}

	// HTML + CSS to explicit file
	htmlOut := filepath.Join(tmpDir, "ad.html")
	var statusBuf bytes.Buffer
	if err := run(nil, &statusBuf, "-f", "html", "-css", "-o", htmlOut, pcbFile); err != nil {
		t.Fatalf("run(html file) error = %v", err)
	}
	if !strings.Contains(statusBuf.String(), "Done!") {
		t.Errorf("expected Done! status, got %q", statusBuf.String())
	}
	htmlBytes, err := os.ReadFile(htmlOut)
	if err != nil {
		t.Fatal(err)
	}
	htmlStr := string(htmlBytes)
	if !strings.Contains(htmlStr, "<style>") || !strings.Contains(htmlStr, `<i class="PB1 PFE">█Hello</i>`) {
		t.Errorf("unexpected html output: %q", htmlStr)
	}

	// Already-UTF-8 input to -f html is not double-decoded
	var utf8HTMLBuf bytes.Buffer
	if err := run(strings.NewReader("@X1E█Hello"), &utf8HTMLBuf, "-f", "html", "-"); err != nil {
		t.Fatalf("run(utf8 html stdin) error = %v", err)
	}
	if want := `<i class="PB1 PFE">█Hello</i>`; utf8HTMLBuf.String() != want {
		t.Errorf("utf8 html stdin = %q, want %q", utf8HTMLBuf.String(), want)
	}

	// No-overwrite blocks replacing existing file
	var noOWBuf bytes.Buffer
	err = run(nil, &noOWBuf, "-n", pcbFile, htmlOut)
	if !errors.Is(err, ErrOutput) {
		t.Errorf("expected errOutput, got %v", err)
	}

	// Strip format via stdin
	var stripBuf bytes.Buffer
	if err := run(strings.NewReader("@CLS@@X0FHello@PAUSE@"), &stripBuf, "-f", "strip", "-"); err != nil {
		t.Fatalf("run(strip stdin) error = %v", err)
	}
	if stripBuf.String() != "Hello" {
		t.Errorf("strip stdin = %q, want %q", stripBuf.String(), "Hello")
	}

	// Default 80-column wrap vs -unlock
	longLine := "@X0F" + strings.Repeat("A", 80) + "B"
	var wrappedBuf bytes.Buffer
	if err := run(strings.NewReader(longLine), &wrappedBuf, "-f", "strip", "-"); err != nil {
		t.Fatalf("run(wrapped) error = %v", err)
	}
	if want := strings.Repeat("A", 80) + "\nB"; wrappedBuf.String() != want {
		t.Errorf("wrapped strip = %q, want %q", wrappedBuf.String(), want)
	}

	var unlockedBuf bytes.Buffer
	if err := run(strings.NewReader(longLine), &unlockedBuf, "-f", "strip", "-unlock", "-"); err != nil {
		t.Fatalf("run(-unlock) error = %v", err)
	}
	if want := strings.Repeat("A", 80) + "B"; unlockedBuf.String() != want {
		t.Errorf("unlocked strip = %q, want %q", unlockedBuf.String(), want)
	}

	// PNG output defaults to <input>.png when no output is given
	defaultPNGOut := filepath.Join(tmpDir, "ad.png")
	var pngStatus bytes.Buffer
	if err := run(nil, &pngStatus, "-f", "png", pcbFile); err != nil {
		t.Fatalf("run(png default file) error = %v", err)
	}
	if !strings.Contains(pngStatus.String(), "Done! Saved "+defaultPNGOut) {
		t.Errorf("expected Done! status for %s, got %q", defaultPNGOut, pngStatus.String())
	}
	pngBytes, err := os.ReadFile(defaultPNGOut)
	if err != nil {
		t.Fatal(err)
	}
	if len(pngBytes) < 8 || string(pngBytes[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Errorf("expected PNG signature, got %x", pngBytes[:min(len(pngBytes), 8)])
	}

	// PNG explicit stdout via -o -
	var pngStdout bytes.Buffer
	if err := run(nil, &pngStdout, "-f", "png", "-o", "-", pcbFile); err != nil {
		t.Fatalf("run(png stdout) error = %v", err)
	}
	if len(pngStdout.Bytes()) < 8 || string(pngStdout.Bytes()[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Errorf("expected PNG stdout signature, got %x", pngStdout.Bytes()[:min(len(pngStdout.Bytes()), 8)])
	}
}

func TestCLIErrors(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	ansFile := filepath.Join(tmpDir, "already.ans")
	if err := os.WriteFile(ansFile, []byte("\x1b[31mRed"), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := run(nil, &buf, ansFile)
	if !errors.Is(err, bbs.ErrANSI) {
		t.Errorf("expected bbs.ErrANSI, got %v", err)
	}

	err = run(nil, &buf, "-f", "bogus", ansFile)
	if !errors.Is(err, ErrFmtUnknown) {
		t.Errorf("expected errUnknownFormat, got %v", err)
	}

	err = run(nil, &buf, "-b", "bogus", ansFile)
	if !errors.Is(err, ErrBBSUnknown) {
		t.Errorf("expected errUnknownBBS, got %v", err)
	}
}
