package split_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/bengarrett/bbs/internal/split"
)

func Test_VBars(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{""}, 0},
		{"first", args{"|00"}, 1},
		{"last", args{"|23"}, 1},
		{"out of range", args{"|24"}, 0},
		{"incomplete", args{"|2"}, 0},
		{"multiples", args{"|01Hello|00 |10world"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(split.VBars([]byte(tt.args.s))); got != tt.want {
				fmt.Println(split.VBars([]byte(tt.args.s)))
				t.Errorf("VBars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Celerity(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{""}, 0},
		{"invalid", args{"|s"}, 0},
		{"incomplete", args{"|"}, 0},
		{"first", args{"|k"}, 1},
		{"last", args{"|W"}, 1},
		{"swap", args{"|S"}, 1},
		{"multiples", args{"|k|S|wHello world"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(split.Celerity([]byte(tt.args.s))); got != tt.want {
				fmt.Println(split.Celerity([]byte(tt.args.s)))
				t.Errorf("Celerity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_PCBoard(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{""}, 0},
		{"first", args{"@X00"}, 1},
		{"last", args{"@XFF"}, 1},
		{"out of range", args{"@XFG"}, 0},
		{"incomplete", args{"@X0"}, 0},
		{"multiples", args{"@X01Hello@X00 @X10world"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(split.PCBoard([]byte(tt.args.s))); got != tt.want {
				fmt.Println(split.PCBoard([]byte(tt.args.s)))
				t.Errorf("PCBoard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_CelerityHTML(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{}, "", false},
		{"string", args{"the quick brown fox"}, "the quick brown fox", false},
		{"prefix", args{"|kHello world"}, "<i class=\"PBk PFk\">Hello world</i>", false},
		{
			"background",
			args{"|S|bHello world"},
			"<i class=\"PBb PFw\">Hello world</i>", false,
		},
		{
			"multi",
			args{"|S|gHello|Rworld"},
			"<i class=\"PBg PFw\">Hello</i><i class=\"PBR PFw\">world</i>", false,
		},
		{
			"newline",
			args{"|S|gHello\n|Rworld"},
			"<i class=\"PBg PFw\">Hello\n</i><i class=\"PBR PFw\">world</i>", false,
		},
		{"false positive", args{"| Hello world |"}, "| Hello world |", false},
		{"double bar", args{"||pipes"}, "||pipes", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			err := split.CelerityHTML(&got, []byte(tt.args.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("CelerityHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("CelerityHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}
