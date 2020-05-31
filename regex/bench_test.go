package regex

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Code copied from https://github.com/golang/go/blob/go1.14/src/regexp/all_test.go#L569-L880

// Bitmap used by func special to check whether a character needs to be escaped.
var specialBytes [16]byte

// special reports whether byte b needs to be escaped by QuoteMeta.
func special(b byte) bool {
	return b < utf8.RuneSelf && specialBytes[b%16]&(1<<(b/16)) != 0
}

func init() {
	for _, b := range []byte(`\.+*?()|[]{}^$`) {
		specialBytes[b%16] |= 1 << (b / 16)
	}
}

func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	re := MustCompile("a+b+")
	wantSubs := "aaabb"
	s := []byte("acbb" + wantSubs + "dd")
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		subs := re.Find(s)
		if string(subs) != wantSubs {
			b.Fatalf("Find(%q) = %q; want %q", s, subs, wantSubs)
		}
	}
}

func BenchmarkFindAllNoMatches(b *testing.B) {
	re := MustCompile("a+b+")
	s := []byte("acddee")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		all := re.FindAll(s, -1)
		if all != nil {
			b.Fatalf("FindAll(%q) = %q; want nil", s, all)
		}
	}
}

func BenchmarkFindString(b *testing.B) {
	b.StopTimer()
	re := MustCompile("a+b+")
	wantSubs := "aaabb"
	s := "acbb" + wantSubs + "dd"
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		subs := re.FindString(s)
		if subs != wantSubs {
			b.Fatalf("FindString(%q) = %q; want %q", s, subs, wantSubs)
		}
	}
}

func BenchmarkFindSubmatch(b *testing.B) {
	b.StopTimer()
	re := MustCompile("a(a+b+)b")
	wantSubs := "aaabb"
	s := []byte("acbb" + wantSubs + "dd")
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		subs := re.FindSubmatch(s)
		if string(subs[0]) != wantSubs {
			b.Fatalf("FindSubmatch(%q)[0] = %q; want %q", s, subs[0], wantSubs)
		}
		if string(subs[1]) != "aab" {
			b.Fatalf("FindSubmatch(%q)[1] = %q; want %q", s, subs[1], "aab")
		}
	}
}

func BenchmarkFindStringSubmatch(b *testing.B) {
	b.StopTimer()
	re := MustCompile("a(a+b+)b")
	wantSubs := "aaabb"
	s := "acbb" + wantSubs + "dd"
	b.StartTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		subs := re.FindStringSubmatch(s)
		if subs[0] != wantSubs {
			b.Fatalf("FindStringSubmatch(%q)[0] = %q; want %q", s, subs[0], wantSubs)
		}
		if subs[1] != "aab" {
			b.Fatalf("FindStringSubmatch(%q)[1] = %q; want %q", s, subs[1], "aab")
		}
	}
}

func BenchmarkLiteral(b *testing.B) {
	x := strings.Repeat("x", 50) + "y"
	b.StopTimer()
	re := MustCompile("y")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.MatchString(x) {
			b.Fatalf("no match!")
		}
	}
}

func BenchmarkNotLiteral(b *testing.B) {
	x := strings.Repeat("x", 50) + "y"
	b.StopTimer()
	re := MustCompile(".y")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.MatchString(x) {
			b.Fatalf("no match!")
		}
	}
}

func BenchmarkMatchClass(b *testing.B) {
	b.StopTimer()
	x := strings.Repeat("xxxx", 20) + "w"
	re := MustCompile("[abcdw]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.MatchString(x) {
			b.Fatalf("no match!")
		}
	}
}

func BenchmarkMatchClass_InRange(b *testing.B) {
	b.StopTimer()
	// 'b' is between 'a' and 'c', so the charclass
	// range checking is no help here.
	x := strings.Repeat("bbbb", 20) + "c"
	re := MustCompile("[ac]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !re.MatchString(x) {
			b.Fatalf("no match!")
		}
	}
}

func BenchmarkReplaceAll(b *testing.B) {
	x := "abcdefghijklmnopqrstuvwxyz"
	b.StopTimer()
	re := MustCompile("[cjrw]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.ReplaceAllString(x, "")
	}
}

func BenchmarkAnchoredLiteralShortNonMatch(b *testing.B) {
	b.StopTimer()
	x := []byte("abcdefghijklmnopqrstuvwxyz")
	re := MustCompile("^zbc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkAnchoredLiteralLongNonMatch(b *testing.B) {
	b.StopTimer()
	x := []byte("abcdefghijklmnopqrstuvwxyz")
	for i := 0; i < 15; i++ {
		x = append(x, x...)
	}
	re := MustCompile("^zbc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkAnchoredShortMatch(b *testing.B) {
	b.StopTimer()
	x := []byte("abcdefghijklmnopqrstuvwxyz")
	re := MustCompile("^.bc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkAnchoredLongMatch(b *testing.B) {
	b.StopTimer()
	x := []byte("abcdefghijklmnopqrstuvwxyz")
	for i := 0; i < 15; i++ {
		x = append(x, x...)
	}
	re := MustCompile("^.bc(d|e)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkOnePassShortA(b *testing.B) {
	b.StopTimer()
	x := []byte("abcddddddeeeededd")
	re := MustCompile("^.bc(d|e)*$")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkNotOnePassShortA(b *testing.B) {
	b.StopTimer()
	x := []byte("abcddddddeeeededd")
	re := MustCompile(".bc(d|e)*$")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkOnePassShortB(b *testing.B) {
	b.StopTimer()
	x := []byte("abcddddddeeeededd")
	re := MustCompile("^.bc(?:d|e)*$")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkNotOnePassShortB(b *testing.B) {
	b.StopTimer()
	x := []byte("abcddddddeeeededd")
	re := MustCompile(".bc(?:d|e)*$")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkOnePassLongPrefix(b *testing.B) {
	b.StopTimer()
	x := []byte("abcdefghijklmnopqrstuvwxyz")
	re := MustCompile("^abcdefghijklmnopqrstuvwxyz.*$")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkOnePassLongNotPrefix(b *testing.B) {
	b.StopTimer()
	x := []byte("abcdefghijklmnopqrstuvwxyz")
	re := MustCompile("^.bcdefghijklmnopqrstuvwxyz.*$")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		re.Match(x)
	}
}

func BenchmarkMatchParallelShared(b *testing.B) {
	x := []byte("this is a long line that contains foo bar baz")
	re := MustCompile("foo (ba+r)? baz")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			re.Match(x)
		}
	})
}

func BenchmarkMatchParallelCopied(b *testing.B) {
	x := []byte("this is a long line that contains foo bar baz")
	re := MustCompile("foo (ba+r)? baz")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		re := re.Copy()
		for pb.Next() {
			re.Match(x)
		}
	})
}

var sink string

func BenchmarkQuoteMetaAll(b *testing.B) {
	specials := make([]byte, 0)
	for i := byte(0); i < utf8.RuneSelf; i++ {
		if special(i) {
			specials = append(specials, i)
		}
	}
	s := string(specials)
	b.SetBytes(int64(len(s)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = QuoteMeta(s)
	}
}

func BenchmarkQuoteMetaNone(b *testing.B) {
	s := "abcdefghijklmnopqrstuvwxyz"
	b.SetBytes(int64(len(s)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = QuoteMeta(s)
	}
}

var compileBenchData = []struct{ name, re string }{
	{"Onepass", `^a.[l-nA-Cg-j]?e$`},
	{"Medium", `^((a|b|[d-z0-9])*(日){4,5}.)+$`},
	{"Hard", strings.Repeat(`((abc)*|`, 50) + strings.Repeat(`)`, 50)},
}

func BenchmarkCompile(b *testing.B) {
	for _, data := range compileBenchData {
		b.Run(data.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := Compile(data.re); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// End copied code
