package bbs_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/bengarrett/bbs"
)

const (
	ansiEsc = "\x1B\x5B"
)

func TestBBS_String(t *testing.T) {
	tests := []struct {
		name string
		b    bbs.BBS
		want string
	}{
		{"too small", -1, ""},
		{"too big", 111, ""},
		{"first", bbs.Celerity, "Celerity |"},
		{"last", bbs.WWIVHeart, "WWIV ♥"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.want {
				t.Errorf("BBS.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBS_Name(t *testing.T) {
	tests := []struct {
		name string
		b    bbs.BBS
		want string
	}{
		{"too small", -1, ""},
		{"too big", 111, ""},
		{"first", bbs.Celerity, "Celerity"},
		{"last", bbs.WWIVHeart, "WWIV ♥"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.Name(); got != tt.want {
				t.Errorf("BBS.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bbs.BBS
	}{
		{"empty", args{""}, -1},
		{"ansi", args{ansiEsc + "0;"}, bbs.ANSI},
		{"cls", args{"@CLS@"}, -1},
		{"pcb+ans", args{"@CLS@" + ansiEsc + "0;"}, bbs.ANSI},
		{"pcb+ans", args{"@CLS@Hello world\nThis is some text." + ansiEsc + "0;"}, bbs.ANSI},
		{"celerity", args{"Hello world\n|WThis is a newline."}, bbs.Celerity},
		{"renegade", args{"Hello world\n|09This is a newline."}, bbs.Renegade},
		{"pcboard", args{"Hello world\n@X01This is a newline."}, bbs.PCBoard},
		{"telegard", args{"Hello world\n`09This is a newline."}, bbs.Telegard},
		{"wildcat", args{"Hello world\n@01@This is a newline."}, bbs.Wildcat},
		{"wwiv ♥", args{"Hello world\n\x031This is a newline."}, bbs.WWIVHeart},
		{"pcboard with nulls", args{"hello\n\n@X01world"}, bbs.PCBoard},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.args.s)
			if got := bbs.Find(r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBBS_HTML(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		bbs     bbs.BBS
		args    args
		want    string
		wantErr bool
	}{
		{"empty", -1, args{}, "", true},
		{"plaintext", -1, args{"text"}, "", true},
		{"plaintext", bbs.ANSI, args{"\x27\x91text"}, "", true},
		{
			"celerity", bbs.Celerity,
			args{"|S|gHello|Rworld"},
			"<i class=\"PBg PFw\">Hello</i><i class=\"PBR PFw\">world</i>", false,
		},
		{
			"xss", bbs.Celerity,
			args{"|S|gABC<script>alert('xss');</script>D|REF"},
			"<i class=\"PBg PFw\">ABC&lt;script&gt;alert(&#39;xss&#39;);&lt;/script&gt;D</i><i class=\"PBR PFw\">EF</i>", false,
		},
	}
	for _, tt := range tests {
		got := bytes.Buffer{}
		err := tt.bbs.HTML(&got, []byte(tt.args.s))
		if (err != nil) != tt.wantErr {
			t.Errorf("BBS.HTML() %v error = %v, wantErr %v", tt.name, err, tt.wantErr)
			return
		}
		if got.String() != tt.want {
			t.Errorf("BBS.HTML() = %v, want %v", got, tt.want)
		}
	}
}

func Test_IsCelerity(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{[]byte{}}, false},
		{"ansi", args{[]byte(ansiEsc + "0;")}, false},
		{"false positive z", args{[]byte("Hello |Zworld")}, false},
		{"false positive s", args{[]byte("Hello |sworld")}, false},
		{"cel B", args{[]byte("Hello |Bworld")}, true},
		{"cel W", args{[]byte("Hello world\n|WThis is a newline.")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.IsCelerity(tt.args.b); got != tt.want {
				t.Errorf("IsCelerity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findRenegade(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{nil}, false},
		{"celerity", args{[]byte("Hello |Bworld")}, false},
		{"first", args{[]byte("|00")}, true},
		{"end", args{[]byte("|23")}, true},
		{"newline", args{[]byte("Hello world\n|15This is a newline.")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.IsRenegade(tt.args.b); got != tt.want {
				t.Errorf("IsRenegade() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsPCBoard(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{nil}, false},
		{"malformed", args{[]byte("@XHello world")}, false},
		{"incomplete", args{[]byte("@X0Hello world")}, false},
		{"out of range", args{[]byte("@X0GHello world")}, false},
		{"first", args{[]byte("@X00Hello world")}, true},
		{"end", args{[]byte("@XFFHello world")}, true},
		{"newline", args{[]byte("Hello world\n@X00This is a newline.")}, true},
		{"false pos", args{[]byte("PCBoard @X code")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.IsPCBoard(tt.args.b); got != tt.want {
				t.Errorf("IsPCBoard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsWildcat(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{nil}, false},
		{"malformed", args{[]byte("@Hello world")}, false},
		{"incomplete", args{[]byte("@0Hello world")}, false},
		{"out of range", args{[]byte("@0@GHello world")}, false},
		{"first", args{[]byte("@00@Hello world")}, true},
		{"end", args{[]byte("@FF@Hello world")}, true},
		{"newline", args{[]byte("Hello world\n@00@This is a newline.")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.IsWildcat(tt.args.b); got != tt.want {
				t.Errorf("IsWildcat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsWWIVHeart(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{nil}, false},
		{"malformed", args{[]byte("\x03Hello world")}, false},
		{"first", args{[]byte("\x030Hello world")}, true},
		{"last", args{[]byte("\x039Hello world")}, true},
		{"lots of numbers", args{[]byte("\x0398765 Hello world")}, true},
		{"newline", args{[]byte("Hello world\n\x031This is a newline.")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.IsWWIVHeart(tt.args.b); got != tt.want {
				t.Errorf("IsWWIVHeart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsWWIVHash(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{nil}, false},
		{"malformed |#", args{[]byte("|#Hello world")}, false},
		{"malformed |0", args{[]byte("|0Hello world")}, false},
		{"malformed #0", args{[]byte("#0Hello world")}, false},
		{"first", args{[]byte("|#0Hello world")}, true},
		{"last", args{[]byte("|#9Hello world")}, true},
		{"lots of numbers", args{[]byte("|#98765 Hello world")}, true},
		{"newline", args{[]byte("Hello world\n|#1This is a newline.")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.IsWWIVHash(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsWWIVHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_RenegadeHTML(t *testing.T) {
	type args struct {
		s string
	}
	const black, white, red = "0", "7", "20"
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"string", args{"hello world"}, "hello world", false},
		{"false pos", args{"hello|world"}, "hello|world", false},
		{"false pos double", args{"| hello world |"}, "| hello world |", false},
		{"prefix", args{"|" + black + white + "Hello world"}, "<i class=\"P0 P7\">Hello world</i>", false},
		{
			"multi",
			args{"|" + black + white + "White |" + red + "Red Background"},
			"<i class=\"P0 P7\">White </i><i class=\"P20 P7\">Red Background</i>", false,
		},
		{
			"newline",
			args{"|07White\n|20Red Background"},
			"<i class=\"P0 P7\">White\n</i><i class=\"P20 P7\">Red Background</i>", false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			err := bbs.RenegadeHTML(&got, []byte(tt.args.s)...)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenegadeHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("RenegadeHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_PCBoardHTML(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args

		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"string", args{"hello world"}, "hello world", false},
		{"prefix", args{"@X07Hello world"}, "<i class=\"PB0 PF7\">Hello world</i>", false},
		{"casing", args{"@xaBHello world"}, "<i class=\"PBA PFB\">Hello world</i>", false},
		{"multi", args{"@X07Hello @X11world"}, "<i class=\"PB0 PF7\">Hello </i><i class=\"PB1 PF1\">world</i>", false},
		{"newline", args{"@X07Hello\n@X11world"}, "<i class=\"PB0 PF7\">Hello\n</i><i class=\"PB1 PF1\">world</i>", false},
		{"false pos 0", args{"@X code for PCBoard"}, "@X code for PCBoard", false},
		{"false pos 1", args{"PCBoard @X code"}, "PCBoard @X code", false},
		{"false pos 2", args{"PCBoard @Xcode"}, "PCBoard @Xcode", false},
		{"false pos 3", args{"Does PCBoard @X code offer a red @X?"}, "Does PCBoard @X code offer a red @X?", false},
		{"combo", args{"@X07@Xcodes combo"}, "<i class=\"PB0 PF7\">@Xcodes combo</i>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			err := bbs.PCBoardHTML(&got, []byte(tt.args.s)...)
			if (err != nil) != tt.wantErr {
				t.Errorf("PCBoardHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("PCBoardHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TelegardHTML(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{""}, "", false},
		{"string", args{"hello world"}, "hello world", false},
		{"prefix", args{"`07Hello world"}, "<i class=\"PB0 PF7\">Hello world</i>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			err := bbs.TelegardHTML(&got, []byte(tt.args.s)...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TelegardHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("TelegardHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_WWIVHashHTML(t *testing.T) {
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
		{"string", args{"hello world"}, "hello world", false},
		{"prefix", args{"|#7Hello world"}, "<i class=\"P0 P7\">Hello world</i>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			err := bbs.WWIVHashHTML(&got, []byte(tt.args.s)...)
			if (err != nil) != tt.wantErr {
				t.Errorf("WWIVHashHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("WWIVHashHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_WWIVHeartHTML(t *testing.T) {
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
		{"string", args{"hello world"}, "hello world", false},
		{"prefix", args{"\x037Hello world"}, "<i class=\"P0 P7\">Hello world</i>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			err := bbs.WWIVHeartHTML(&got, []byte(tt.args.s)...)
			if (err != nil) != tt.wantErr {
				t.Errorf("WWIVHeartHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("WWIVHeartHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_WildcatHTML(t *testing.T) {
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
		{"string", args{"hello world"}, "hello world", false},
		{"prefix", args{"@0F@Hello world"}, "<i class=\"PB0 PFF\">Hello world</i>", false},
	}
	for _, tt := range tests {
		got := bytes.Buffer{}
		err := bbs.WildcatHTML(&got, []byte(tt.args.s)...)
		if (err != nil) != tt.wantErr {
			t.Errorf("WildcatHTML() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if got.String() != tt.want {
			t.Errorf("WildcatHTML() = %v, want %v", got, tt.want)
		}
	}
}

func TestBBS_Remove(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		b       bbs.BBS
		args    args
		want    string
		wantErr bool
	}{
		{"empty", bbs.PCBoard, args{[]byte("")}, "", false},
		{"invalid", -1, args{[]byte("")}, "", true},
		{"incorrect", bbs.WWIVHash, args{[]byte("@X07Hello world")}, "@X07Hello world", false},
		{"ansi", bbs.ANSI, args{[]byte("")}, "", true},
		{"celerity", bbs.Celerity, args{[]byte("Hello |Bworld")}, "Hello world", false},
		{"pcboard", bbs.PCBoard, args{[]byte("@X07Hello world")}, "Hello world", false},
		{"pcboard nl", bbs.PCBoard, args{[]byte("@X07Hello\n@X11world@X01")}, "Hello\nworld", false},
		{"pcboard false pos", bbs.PCBoard, args{[]byte("@X07PCBoard @X code")}, "PCBoard @X code", false},
		{"renegade", bbs.Renegade, args{[]byte("Hello |15world")}, "Hello world", false},
		{"telegard", bbs.Telegard, args{[]byte("`07Hello world")}, "Hello world", false},
		{"whash", bbs.WWIVHash, args{[]byte("|#7Hello world")}, "Hello world", false},
		{"wheart", bbs.WWIVHeart, args{[]byte("\x037Hello world")}, "Hello world", false},
		{"wildcat", bbs.Wildcat, args{[]byte("@0F@Hello world")}, "Hello world", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytes.Buffer{}
			if err := tt.b.Remove(&got, tt.args.src...); (err != nil) != tt.wantErr {
				t.Errorf("BBS.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.String() != tt.want {
				t.Errorf("BBS.Remove() dst = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestTrimControls(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"empty", args{[]byte("")}, nil},
		{"none", args{[]byte("Hello world.")}, []byte("Hello world.")},
		{"clear", args{[]byte("@CLS@Hello world.")}, []byte("Hello world.")},
		{"pause", args{[]byte("@PAUSE@Hello world.")}, []byte("Hello world.")},
		{"both", args{[]byte("@CLS@@PAUSE@Hello world.")}, []byte("Hello world.")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bbs.TrimControls(tt.args.b...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrimControls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCelerityHTMLNil(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if err := bbs.CelerityHTML(nil, []byte{}...); err == nil {
			t.Errorf("CelerityHTML() error = %v, wantErr %v", err, true)
		}
	})
}
