package split_test

import (
	"fmt"
	"testing"

	"github.com/bengarrett/bbs/internal/split"
)

func Test_Bars(t *testing.T) {
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
			if got := len(split.Bars(tt.args.s)); got != tt.want {
				fmt.Println(split.Bars(tt.args.s))
				t.Errorf("Bars() = %v, want %v", got, tt.want)
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
			if got := len(split.PCBoard(tt.args.s)); got != tt.want {
				fmt.Println(split.PCBoard(tt.args.s))
				t.Errorf("PCBoard() = %v, want %v", got, tt.want)
			}
		})
	}
}
