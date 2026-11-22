// Package main provides the bbs command-line interface.
//
//nolint:gochecknoglobals
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bengarrett/bbs"
	"golang.org/x/text/encoding/charmap"
)

const (
	filePerm = 0o600
	repoURL  = "https://github.com/bengarrett/bbs"
)

var (
	ErrInput      = errors.New("no input file specified")
	ErrOutput     = errors.New("output file already exists")
	ErrFmtUnknown = errors.New("unknown format (valid: ansi, html, png, strip)")
	ErrBBSUnknown = errors.New("unknown bbs type (valid: " +
		"auto, pcboard, renegade, celerity, wildcat, telegard, wwivhash, wwivheart)")
)

var (
	version string // populated by GoReleaser ldflags
	date    string // populated by GoReleaser ldflags
)

type flags struct {
	output string
	format string
	kind   string
	noOW   bool
	check  bool
	css    bool
	utf8   bool
	unlock bool
	quiet  bool
	ver    bool
}

func main() {
	if err := run(os.Stdin, os.Stdout, os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "Oops, %v\n", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, out io.Writer, args ...string) error {
	flagSet := flag.NewFlagSet("bbs", flag.ContinueOnError)
	flagSet.SetOutput(out)
	flagSet.Usage = func() {
		printUsage(out)
	}

	var opts flags
	setupFlags(flagSet, &opts)

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	if opts.ver {
		fmt.Fprintf(out, "bbs %s\n© 2026 Ben Garrett. All rights reserved. MIT license.\n%s\n",
			buildVersion(), repoURL)
		return nil
	}

	posArgs := flagSet.Args()
	if len(posArgs) == 0 {
		printUsage(out)
		return ErrInput
	}
	if wantsHelp(posArgs) {
		printUsage(out)
		return nil
	}

	input := posArgs[0]
	if opts.check {
		return executeCheck(stdin, out, input)
	}

	output := determineOutput(input, opts.output, opts.format, posArgs...)
	return execute(stdin, out, input, output, &opts)
}

func wantsHelp(args []string) bool {
	return slices.ContainsFunc(args, func(arg string) bool {
		switch arg {
		case "-h", "-help", "--help", "-?", "/?":
			return true
		default:
			return false
		}
	})
}

func setupFlags(flagSet *flag.FlagSet, opts *flags) {
	flagSet.StringVar(&opts.output, "o", "",
		"Destination filename (defaults to stdout, or <input>.png for png)")
	flagSet.StringVar(&opts.output, "output", "",
		"Destination filename (defaults to stdout, or <input>.png for png)")
	flagSet.StringVar(&opts.format, "f", "ansi",
		"Output format: ansi, html, png, strip")
	flagSet.StringVar(&opts.format, "format", "ansi",
		"Output format: ansi, html, png, strip")
	flagSet.StringVar(&opts.kind, "b", "auto",
		"BBS code format: auto, pcboard, renegade, celerity, wildcat, telegard, wwivhash, wwivheart")
	flagSet.StringVar(&opts.kind, "bbs", "auto",
		"BBS code format: auto, pcboard, renegade, celerity, wildcat, telegard, wwivhash, wwivheart")
	flagSet.BoolVar(&opts.css, "css", false,
		"Prepend embedded CSS stylesheet when format is html")
	flagSet.BoolVar(&opts.utf8, "utf8", false,
		"Decode CP437 input bytes to UTF-8 (automatic for html)")
	flagSet.BoolVar(&opts.unlock, "u", false,
		"Unlock 80-column line wrapping")
	flagSet.BoolVar(&opts.unlock, "unlock", false,
		"Unlock 80-column line wrapping")
	flagSet.BoolVar(&opts.noOW, "n", false,
		"No overwriting existing output files")
	flagSet.BoolVar(&opts.noOW, "no-overwrite", false,
		"No overwriting existing output files")
	flagSet.BoolVar(&opts.check, "c", false,
		"Only detect the BBS color code format without converting")
	flagSet.BoolVar(&opts.check, "check", false,
		"Only detect the BBS color code format without converting")
	flagSet.BoolVar(&opts.quiet, "q", false,
		"Quiet status output when writing to a file")
	flagSet.BoolVar(&opts.quiet, "quiet", false,
		"Quiet status output when writing to a file")
	flagSet.BoolVar(&opts.ver, "v", false,
		"Display version")
	flagSet.BoolVar(&opts.ver, "version", false,
		"Display version")
}

func determineOutput(input, flagOut, format string, posArgs ...string) string {
	if len(posArgs) > 1 && posArgs[1] != "" {
		return posArgs[1]
	}
	if flagOut != "" {
		return flagOut
	}
	if strings.EqualFold(strings.TrimSpace(format), "png") {
		if input == "-" {
			return "bbs.png"
		}
		ext := filepath.Ext(input)
		if ext != "" && ext != filepath.Base(input) {
			return strings.TrimSuffix(input, ext) + ".png"
		}
		return input + ".png"
	}
	return ""
}

func executeCheck(stdin io.Reader, out io.Writer, input string) error {
	src, label, err := readInput(stdin, input)
	if err != nil {
		return err
	}

	found := bbs.Find(bytes.NewReader(src))
	if !found.Valid() {
		fmt.Fprintf(out, "%s: Plain text (no BBS color codes found)\n", label)
		return nil
	}
	fmt.Fprintf(out, "%s: %s (%s)\n", label, found.Name(), found.String())
	return nil
}

func execute(stdin io.Reader, out io.Writer, input, output string, opts *flags) error {
	if opts.noOW && output != "" && output != "-" {
		_, err := os.Stat(filepath.Clean(output))
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no overwrite: %w", ErrOutput)
		}
	}

	forcedKind, autoDetect, err := parseBBSKind(opts.kind)
	if err != nil {
		return err
	}

	src, _, err := readInput(stdin, input)
	if err != nil {
		return err
	}

	start := time.Now()
	result, err := convertSource(src, forcedKind, autoDetect, opts)
	if err != nil {
		return err
	}
	elapsed := time.Since(start)

	if output == "" || output == "-" {
		if _, err := out.Write(result); err != nil {
			return fmt.Errorf("write stdout: %w", err)
		}
		return nil
	}

	if err := os.WriteFile(filepath.Clean(output), result, filePerm); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	if !opts.quiet {
		reportFinished(out, output, elapsed)
	}
	return nil
}

func readInput(stdin io.Reader, input string) ([]byte, string, error) {
	if input == "-" {
		const stin = "stdin"
		if stdin == nil {
			return nil, stin, io.ErrUnexpectedEOF
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, stin, fmt.Errorf("read stdin: %w", err)
		}
		return data, stin, nil
	}

	clean := filepath.Clean(input)
	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, filepath.Base(input), fmt.Errorf("open input: %w", err)
	}
	return data, filepath.Base(input), nil
}

func parseBBSKind(raw string) (bbs.BBS, bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto":
		return -1, true, nil
	case "pcboard", "pcb":
		return bbs.PCBoard, false, nil
	case "renegade", "rene":
		return bbs.Renegade, false, nil
	case "celerity":
		return bbs.Celerity, false, nil
	case "wildcat":
		return bbs.Wildcat, false, nil
	case "telegard":
		return bbs.Telegard, false, nil
	case "wwivhash", "wwiv#":
		return bbs.WWIVHash, false, nil
	case "wwivheart":
		return bbs.WWIVHeart, false, nil
	default:
		return -1, false, fmt.Errorf("%w: %q", ErrBBSUnknown, raw)
	}
}

func convertSource(src []byte, forcedKind bbs.BBS, autoDetect bool, opts *flags) ([]byte, error) {
	format := strings.ToLower(strings.TrimSpace(opts.format))
	if format == "" {
		format = "ansi"
	}

	prepared, kind, err := prepareSource(src, forcedKind, autoDetect, format, opts.utf8)
	if err != nil {
		return nil, err
	}

	bbsOpts := bbs.Options{
		Unlock: opts.unlock,
	}

	var buf bytes.Buffer
	switch format {
	case "ansi":
		err = kind.ANSIOptions(&buf, prepared, bbsOpts)
	case "html":
		err = convertHTML(&buf, kind, prepared, bbsOpts, opts.css)
	case "png":
		err = kind.PNGOptions(&buf, prepared, bbsOpts)
	case "strip", "plain":
		err = kind.RemoveOptions(&buf, bbsOpts, bbs.TrimControls(prepared...)...)
	default:
		return nil, fmt.Errorf("%w: %q", ErrFmtUnknown, opts.format)
	}
	if err != nil {
		return nil, fmt.Errorf("convert to %s: %w", format, err)
	}

	return buf.Bytes(), nil
}

func prepareSource(
	src []byte,
	forcedKind bbs.BBS,
	autoDetect bool,
	format string,
	decodeUTF8 bool,
) ([]byte, bbs.BBS, error) {
	if decodeUTF8 || (format == "html" && !isUTF8WithHighBit(src)) {
		decoded, err := charmap.CodePage437.NewDecoder().Bytes(src)
		if err != nil {
			return nil, -1, fmt.Errorf("decode cp437: %w", err)
		}
		src = decoded
	}

	kind := forcedKind
	if autoDetect {
		kind = bbs.Find(bytes.NewReader(src))
	}
	return src, kind, nil
}

func convertHTML(buf *bytes.Buffer, kind bbs.BBS, src []byte, opts bbs.Options, withCSS bool) error {
	if withCSS {
		if err := kind.DocumentOptions(buf, src, opts); err != nil {
			return fmt.Errorf("write html document: %w", err)
		}
		return nil
	}
	if err := kind.HTMLOptions(buf, src, opts); err != nil {
		return fmt.Errorf("write html: %w", err)
	}
	return nil
}

func isUTF8WithHighBit(src []byte) bool {
	hasHighBit := false
	for _, b := range src {
		if b >= utf8.RuneSelf {
			hasHighBit = true
			break
		}
	}
	return hasHighBit && utf8.Valid(src)
}

func buildVersion() string {
	ver := resolveVersion()
	if buildDate := resolveDate(); buildDate != "" {
		return fmt.Sprintf("%s (%s)", ver, buildDate)
	}
	return ver
}

func resolveVersion() string {
	if version != "" {
		return version
	}
	const shaLen = 7
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			if len(setting.Value) > shaLen {
				return setting.Value[:shaLen]
			}
			return setting.Value
		}
	}
	return "devel"
}

func resolveDate() string {
	raw := date
	if raw == "" {
		raw = vcsTime()
	}
	if raw == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err == nil {
		return parsed.UTC().Format("2006-JAN-02")
	}
	return raw
}

func vcsTime() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.time" {
			return setting.Value
		}
	}
	return ""
}

func reportFinished(out io.Writer, output string, elapsed time.Duration) {
	stat, err := os.Stat(filepath.Clean(output))
	if err == nil {
		fmt.Fprintf(out, "Done! Saved %s (%s) in %v\n",
			output, humanSize(stat.Size()), elapsed.Round(time.Millisecond))
		return
	}
	fmt.Fprintf(out, "Done in %v\n", elapsed.Round(time.Millisecond))
}

func humanSize(size int64) string {
	const kb = 1024
	const mb = 1024 * 1024

	switch {
	case size >= mb:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(mb))
	case size >= kb:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(kb))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "BBS: Converts legacy BBS color-coded textfiles to ANSI, HTML, PNG, or plain text.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  bbs [flags] <input|-> [output]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  -f, -format <name>   Output format: ansi, html, png, strip (default \"ansi\")")
	fmt.Fprintln(out, "  -b, -bbs <name>      Force BBS format instead of auto-detecting (default \"auto\")")
	fmt.Fprintln(out, "  -o, -output <file>   Destination filename (defaults to stdout, or <input>.png for png)")
	fmt.Fprintln(out, "  -n, -no-overwrite    Do not overwrite existing output files")
	fmt.Fprintln(out, "  -css                 Include embedded CSS stylesheet with -format html")
	fmt.Fprintln(out, "  -utf8                Decode CP437 input bytes to UTF-8 (always on for html)")
	fmt.Fprintln(out, "  -u, -unlock          Unlock 80-column line wrapping")
	fmt.Fprintln(out, "  -c, -check           Only detect the BBS color code format without converting")
	fmt.Fprintln(out, "  -q, -quiet           Silence status output when writing to a file")
	fmt.Fprintln(out, "  -v, -version         Display version")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  bbs display.pcb")
	fmt.Fprintln(out, "  bbs display.pcb output.ans")
	fmt.Fprintln(out, "  bbs -f html -css display.pcb display.html")
	fmt.Fprintln(out, "  bbs -f png display.pcb")
	fmt.Fprintln(out, "  bbs -f strip display.pcb")
	fmt.Fprintln(out, "  bbs -check display.pcb")
}
