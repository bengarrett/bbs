//go:build js && wasm

// Package main provides the WebAssembly in-browser API for bbs.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"syscall/js"
	"unicode/utf8"

	"github.com/bengarrett/bbs"
	"golang.org/x/text/encoding/charmap"
)

var (
	errNoInput     = errors.New("input data is required")
	errInvalidType = errors.New("input must be a Uint8Array, ArrayBuffer, or string")
	errInvalidOpts = errors.New("options must be an object")
	errUnknownBBS  = errors.New("unknown bbs type (valid: " +
		"auto, pcboard, renegade, celerity, wildcat, telegard, wwivhash, wwivheart)")
)

//nolint:gochecknoglobals
var (
	version = "devel"
	date    = ""
)

type options struct {
	kind       bbs.BBS
	autoDetect bool
	utf8       bool
	css        bool
	unlock     bool
}

func main() {
	api := js.Global().Get("Object").New()
	api.Set("toANSI", js.FuncOf(toANSIPromise))
	api.Set("toANSISync", js.FuncOf(toANSISync))
	api.Set("toHTML", js.FuncOf(toHTMLPromise))
	api.Set("toHTMLSync", js.FuncOf(toHTMLSync))
	api.Set("toPNG", js.FuncOf(toPNGPromise))
	api.Set("toPNGSync", js.FuncOf(toPNGSync))
	api.Set("strip", js.FuncOf(stripPromise))
	api.Set("stripSync", js.FuncOf(stripSync))
	api.Set("css", js.FuncOf(cssSync))
	api.Set("detect", js.FuncOf(detect))
	api.Set("version", js.FuncOf(func(_ js.Value, _ []js.Value) any {
		if date != "" {
			return fmt.Sprintf("%s (%s)", version, date)
		}
		return version
	}))

	js.Global().Set("bbs", api)

	// Keep runtime active in browser/JS environment.
	select {}
}

func toANSIPromise(_ js.Value, args []js.Value) any {
	return runAsync(args, convertANSIValue)
}

func toANSISync(_ js.Value, args []js.Value) any {
	return runSync(args, convertANSIValue)
}

func toHTMLPromise(_ js.Value, args []js.Value) any {
	return runAsync(args, convertHTMLValue)
}

func toHTMLSync(_ js.Value, args []js.Value) any {
	return runSync(args, convertHTMLValue)
}

func toPNGPromise(_ js.Value, args []js.Value) any {
	return runAsync(args, convertPNGValue)
}

func toPNGSync(_ js.Value, args []js.Value) any {
	return runSync(args, convertPNGValue)
}

func stripPromise(_ js.Value, args []js.Value) any {
	return runAsync(args, convertStripValue)
}

func stripSync(_ js.Value, args []js.Value) any {
	return runSync(args, convertStripValue)
}

func cssSync(_ js.Value, args []js.Value) any {
	kind := bbs.PCBoard
	if len(args) > 0 && args[0].Type() == js.TypeString {
		if parsed, auto, err := parseBBSKind(args[0].String()); err == nil && !auto {
			kind = parsed
		}
	}
	css, err := kind.StandaloneCSS()
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}
	return css
}

func detect(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		panic(js.Global().Get("Error").New(errNoInput.Error()))
	}

	input, err := getInputBytes(args[0])
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	found := bbs.Find(bytes.NewReader(input))
	res := js.Global().Get("Object").New()
	res.Set("valid", found.Valid())
	res.Set("format", found.String())
	res.Set("name", found.Name())
	res.Set("isANSI", found == bbs.ANSI)
	return res
}

func runAsync(args []js.Value, fn func([]byte, options) (js.Value, error)) any {
	if len(args) == 0 {
		return promiseReject(errNoInput)
	}

	inputVal := args[0]
	var optsVal js.Value
	if len(args) > 1 {
		optsVal = args[1]
	}

	handler := js.FuncOf(func(_ js.Value, pArgs []js.Value) any {
		resolve := pArgs[0]
		reject := pArgs[1]

		go func() {
			input, err := getInputBytes(inputVal)
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}

			opts, err := parseOptions(optsVal)
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}

			val, err := fn(input, opts)
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}

			resolve.Invoke(val)
		}()

		return nil
	})

	return js.Global().Get("Promise").New(handler)
}

func runSync(args []js.Value, fn func([]byte, options) (js.Value, error)) any {
	if len(args) == 0 {
		panic(js.Global().Get("Error").New(errNoInput.Error()))
	}

	input, err := getInputBytes(args[0])
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	var optsVal js.Value
	if len(args) > 1 {
		optsVal = args[1]
	}
	opts, err := parseOptions(optsVal)
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}

	val, err := fn(input, opts)
	if err != nil {
		panic(js.Global().Get("Error").New(err.Error()))
	}
	return val
}

func convertANSIValue(input []byte, opts options) (js.Value, error) {
	decodeCP437 := opts.utf8 && !isUTF8WithHighBit(input)
	src, kind, err := prepareSource(input, opts, decodeCP437)
	if err != nil {
		return js.Undefined(), err
	}

	var buf bytes.Buffer
	if err := kind.ANSIOptions(&buf, src, bbs.Options{Unlock: opts.unlock}); err != nil {
		return js.Undefined(), fmt.Errorf("convert to ansi: %w", err)
	}
	if opts.utf8 {
		return js.ValueOf(buf.String()), nil
	}
	return toUint8Array(buf.Bytes()), nil
}

func convertHTMLValue(input []byte, opts options) (js.Value, error) {
	// HTML requires UTF-8; decode CP437 automatically unless already non-ASCII UTF-8.
	decodeCP437 := !isUTF8WithHighBit(input)
	src, kind, err := prepareSource(input, opts, decodeCP437)
	if err != nil {
		return js.Undefined(), err
	}

	var buf bytes.Buffer
	bbsOpts := bbs.Options{Unlock: opts.unlock}
	if opts.css {
		if err := kind.DocumentOptions(&buf, src, bbsOpts); err != nil {
			return js.Undefined(), fmt.Errorf("convert to html document: %w", err)
		}
	} else if err := kind.HTMLOptions(&buf, src, bbsOpts); err != nil {
		return js.Undefined(), fmt.Errorf("convert to html: %w", err)
	}
	return js.ValueOf(buf.String()), nil
}

func convertPNGValue(input []byte, opts options) (js.Value, error) {
	src, kind, err := prepareSource(input, opts, false)
	if err != nil {
		return js.Undefined(), err
	}

	var buf bytes.Buffer
	if err := kind.PNGOptions(&buf, src, bbs.Options{Unlock: opts.unlock}); err != nil {
		return js.Undefined(), fmt.Errorf("convert to png: %w", err)
	}
	return toUint8Array(buf.Bytes()), nil
}

func convertStripValue(input []byte, opts options) (js.Value, error) {
	decodeCP437 := opts.utf8 && !isUTF8WithHighBit(input)
	src, kind, err := prepareSource(input, opts, decodeCP437)
	if err != nil {
		return js.Undefined(), err
	}

	var buf bytes.Buffer
	trimmed := bbs.TrimControls(src...)
	if err := kind.RemoveOptions(&buf, bbs.Options{Unlock: opts.unlock}, trimmed...); err != nil {
		return js.Undefined(), fmt.Errorf("strip codes: %w", err)
	}
	if opts.utf8 {
		return js.ValueOf(buf.String()), nil
	}
	return toUint8Array(buf.Bytes()), nil
}

func prepareSource(input []byte, opts options, decodeCP437 bool) ([]byte, bbs.BBS, error) {
	src := input
	if decodeCP437 {
		decoded, err := charmap.CodePage437.NewDecoder().Bytes(src)
		if err != nil {
			return nil, -1, fmt.Errorf("decode cp437: %w", err)
		}
		src = decoded
	}

	kind := opts.kind
	if opts.autoDetect {
		kind = bbs.Find(bytes.NewReader(src))
	}
	return src, kind, nil
}

func isUTF8WithHighBit(src []byte) bool {
	hasHighBit := false
	for _, elem := range src {
		if elem >= utf8.RuneSelf {
			hasHighBit = true
			break
		}
	}
	return hasHighBit && utf8.Valid(src)
}

func getInputBytes(val js.Value) ([]byte, error) {
	if val.IsUndefined() || val.IsNull() {
		return nil, errNoInput
	}

	if val.Type() == js.TypeString {
		return []byte(val.String()), nil
	}

	if val.InstanceOf(js.Global().Get("ArrayBuffer")) {
		val = js.Global().Get("Uint8Array").New(val)
	}

	if val.InstanceOf(js.Global().Get("Uint8Array")) {
		length := val.Get("length").Int()
		dst := make([]byte, length)
		js.CopyBytesToGo(dst, val)
		return dst, nil
	}

	return nil, errInvalidType
}

func parseOptions(optsVal js.Value) (options, error) {
	opts := options{
		kind:       -1,
		autoDetect: true,
	}
	if optsVal.IsUndefined() || optsVal.IsNull() {
		return opts, nil
	}
	if optsVal.Type() != js.TypeObject {
		return opts, errInvalidOpts
	}

	if bbsVal := optsVal.Get("bbs"); !bbsVal.IsUndefined() && !bbsVal.IsNull() {
		kind, auto, err := parseBBSKind(bbsVal.String())
		if err != nil {
			return opts, err
		}
		opts.kind = kind
		opts.autoDetect = auto
	}
	if utf8Val := optsVal.Get("utf8"); !utf8Val.IsUndefined() && !utf8Val.IsNull() {
		opts.utf8 = utf8Val.Bool()
	}
	if cssVal := optsVal.Get("css"); !cssVal.IsUndefined() && !cssVal.IsNull() {
		opts.css = cssVal.Bool()
	}
	if unlockVal := optsVal.Get("unlock"); !unlockVal.IsUndefined() && !unlockVal.IsNull() {
		opts.unlock = unlockVal.Bool()
	}

	return opts, nil
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
	case "wildcat", "wildcat!":
		return bbs.Wildcat, false, nil
	case "telegard":
		return bbs.Telegard, false, nil
	case "wwivhash", "wwiv#", "wwiv #":
		return bbs.WWIVHash, false, nil
	case "wwivheart", "wwiv ♥":
		return bbs.WWIVHeart, false, nil
	default:
		return -1, false, fmt.Errorf("%w: %q", errUnknownBBS, raw)
	}
}

func toUint8Array(data []byte) js.Value {
	arr := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(arr, data)
	return arr
}

func promiseReject(err error) js.Value {
	handler := js.FuncOf(func(_ js.Value, pArgs []js.Value) any {
		reject := pArgs[1]
		reject.Invoke(js.Global().Get("Error").New(err.Error()))
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}
