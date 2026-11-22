package split_test

import (
	"bytes"
	"testing"

	"github.com/bengarrett/bbs/internal/split"
)

func Test_CelerityANSI(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"plain", "plain text", "plain text", false},
		{"prefix", "|kHello", "\x1b[0;40;30mHello", false},
		{"bright", "|BBlue", "\x1b[0;1;40;34mBlue", false},
		{"swap", "|bHello |Sworld", "\x1b[0;40;34mHello \x1b[0;44;30mworld", false},
		{"swap default", "|S|bHello", "\x1b[0;47;30m\x1b[0;47;34mHello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bytes.Buffer
			err := split.CelerityANSI(&got, []byte(tt.src))
			if (err != nil) != tt.wantErr {
				t.Fatalf("CelerityANSI() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.String() != tt.want {
				t.Errorf("CelerityANSI() = %q, want %q", got.String(), tt.want)
			}
		})
	}

	if err := split.CelerityANSI(nil, []byte("x")); err == nil {
		t.Error("CelerityANSI(nil) expected error")
	}
}

func Test_PCBoardANSI(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"empty", "", ""},
		{"plain", "Hello world", "Hello world"},
		{"blue bg yellow fg", "@X1EHello", "\x1b[0;1;44;33mHello"},
		{"red bg white fg", "@X4F!", "\x1b[0;1;41;37m!"},
		{"blink black bg white fg", "@X8FSparkle", "\x1b[0;5;1;40;37mSparkle"},
		{"cp437 preserved", "@X04\xdb\n", "\x1b[0;40;31m\xdb\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bytes.Buffer
			if err := split.PCBoardANSI(&got, []byte(tt.src)); err != nil {
				t.Fatalf("PCBoardANSI() error = %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("PCBoardANSI() = %q, want %q", got.String(), tt.want)
			}
		})
	}

	if err := split.PCBoardANSI(nil, []byte("x")); err == nil {
		t.Error("PCBoardANSI(nil) expected error")
	}
}

func Test_VBarsANSI(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"empty", "", ""},
		{"plain", "Hello", "Hello"},
		{"fg and bg", "|17|14Yellow|04Red", "\x1b[0;44;37m\x1b[0;1;44;33mYellow\x1b[0;44;31mRed"},
		{"out of range kept literal", "|24 literal", "|24 literal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bytes.Buffer
			if err := split.VBarsANSI(&got, []byte(tt.src)); err != nil {
				t.Fatalf("VBarsANSI() error = %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("VBarsANSI() = %q, want %q", got.String(), tt.want)
			}
		})
	}

	if err := split.VBarsANSI(nil, []byte("x")); err == nil {
		t.Error("VBarsANSI(nil) expected error")
	}
}
