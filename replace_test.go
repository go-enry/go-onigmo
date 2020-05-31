// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package onigmo

import (
	"strings"
	"testing"
)

// Code copied from: https://github.com/golang/go/blob/go1.14/src/regexp/all_test.go#L136-L348

type ReplaceTest struct {
	pattern, replacement, input, output string
}

var replaceTests = []ReplaceTest{
	// Test empty input and/or replacement, with pattern that matches the empty string.
	{"", "", "", ""},
	{"", "x", "", "x"},
	{"", "", "abc", "abc"},
	{"", "x", "abc", "xaxbxcx"},

	// Test empty input and/or replacement, with pattern that does not match the empty string.
	{"b", "", "", ""},
	{"b", "x", "", ""},
	{"b", "", "abc", "ac"},
	{"b", "x", "abc", "axc"},
	{"y", "", "", ""},
	{"y", "x", "", ""},
	{"y", "", "abc", "abc"},
	{"y", "x", "abc", "abc"},

	// Multibyte characters -- verify that we don't try to match in the middle
	// of a character.
	{"[a-c]*", "x", "\u65e5", "x\u65e5x"},
	{"[^\u65e5]", "x", "abc\u65e5def", "xxx\u65e5xxx"},

	// Start and end of a string.
	{"^[a-c]*", "x", "abcdabc", "xdabc"},
	{"[a-c]*$", "x", "abcdabc", "abcdx"},
	{"^[a-c]*$", "x", "abcdabc", "abcdabc"},
	{"^[a-c]*", "x", "abc", "x"},
	{"[a-c]*$", "x", "abc", "x"},
	{"^[a-c]*$", "x", "abc", "x"},
	{"^[a-c]*", "x", "dabce", "xdabce"},
	{"[a-c]*$", "x", "dabce", "dabcex"},
	{"^[a-c]*$", "x", "dabce", "dabce"},
	{"^[a-c]*", "x", "", "x"},
	{"[a-c]*$", "x", "", "x"},
	{"^[a-c]*$", "x", "", "x"},

	{"^[a-c]+", "x", "abcdabc", "xdabc"},
	{"[a-c]+$", "x", "abcdabc", "abcdx"},
	{"^[a-c]+$", "x", "abcdabc", "abcdabc"},
	{"^[a-c]+", "x", "abc", "x"},
	{"[a-c]+$", "x", "abc", "x"},
	{"^[a-c]+$", "x", "abc", "x"},
	{"^[a-c]+", "x", "dabce", "dabce"},
	{"[a-c]+$", "x", "dabce", "dabce"},
	{"^[a-c]+$", "x", "dabce", "dabce"},
	{"^[a-c]+", "x", "", ""},
	{"[a-c]+$", "x", "", ""},
	{"^[a-c]+$", "x", "", ""},

	// Other cases.
	{"abc", "def", "abcdefg", "defdefg"},
	{"bc", "BC", "abcbcdcdedef", "aBCBCdcdedef"},
	{"abc", "", "abcdabc", "d"},
	{"x", "xXx", "xxxXxxx", "xXxxXxxXxXxXxxXxxXx"},
	{"abc", "d", "", ""},
	{"abc", "d", "abc", "d"},
	{".+", "x", "abc", "x"},
	{"[a-c]*", "x", "def", "xdxexfx"},
	{"[a-c]+", "x", "abcbcdcdedef", "xdxdedef"},
	{"[a-c]*", "x", "abcbcdcdedef", "xdxdxexdxexfx"},

	// Substitutions
	{"a+", "($0)", "banana", "b(a)n(a)n(a)"},
	{"a+", "(${0})", "banana", "b(a)n(a)n(a)"},
	{"a+", "(${0})$0", "banana", "b(a)an(a)an(a)a"},
	{"a+", "(${0})$0", "banana", "b(a)an(a)an(a)a"},
	{"hello, (.+)", "goodbye, ${1}", "hello, world", "goodbye, world"},
	{"hello, (.+)", "goodbye, $1x", "hello, world", "goodbye, "},
	{"hello, (.+)", "goodbye, ${1}x", "hello, world", "goodbye, worldx"},
	{"hello, (.+)", "<$0><$1><$2><$3>", "hello, world", "<hello, world><world><><>"},
	{"hello, (?P<noun>.+)", "goodbye, $noun!", "hello, world", "goodbye, world!"},
	{"hello, (?P<noun>.+)", "goodbye, ${noun}", "hello, world", "goodbye, world"},
	{"(?P<x>hi)|(?P<x>bye)", "$x$x$x", "hi", "hihihi"},
	{"(?P<x>hi)|(?P<x>bye)", "$x$x$x", "bye", "byebyebye"},
	{"(?P<x>hi)|(?P<x>bye)", "$xyz", "hi", ""},
	{"(?P<x>hi)|(?P<x>bye)", "${x}yz", "hi", "hiyz"},
	{"(?P<x>hi)|(?P<x>bye)", "hello $$x", "hi", "hello $x"},
	{"a+", "${oops", "aaa", "${oops"},
	{"a+", "$$", "aaa", "$"},
	{"a+", "$", "aaa", "$"},

	// Substitution when subexpression isn't found
	{"(x)?", "$1", "123", "123"},
	{"abc", "$1", "123", "123"},

	// Substitutions involving a (x){0}
	{"(a)(b){0}(c)", ".$1|$3.", "xacxacx", "x.a|c.x.a|c.x"},
	{"(a)(((b))){0}c", ".$1.", "xacxacx", "x.a.x.a.x"},
	{"((a(b){0}){3}){5}(h)", "y caramb$2", "say aaaaaaaaaaaaaaaah", "say ay caramba"},
	{"((a(b){0}){3}){5}h", "y caramb$2", "say aaaaaaaaaaaaaaaah", "say ay caramba"},
}

var replaceLiteralTests = []ReplaceTest{
	// Substitutions
	{"a+", "($0)", "banana", "b($0)n($0)n($0)"},
	{"a+", "(${0})", "banana", "b(${0})n(${0})n(${0})"},
	{"a+", "(${0})$0", "banana", "b(${0})$0n(${0})$0n(${0})$0"},
	{"a+", "(${0})$0", "banana", "b(${0})$0n(${0})$0n(${0})$0"},
	{"hello, (.+)", "goodbye, ${1}", "hello, world", "goodbye, ${1}"},
	{"hello, (?P<noun>.+)", "goodbye, $noun!", "hello, world", "goodbye, $noun!"},
	{"hello, (?P<noun>.+)", "goodbye, ${noun}", "hello, world", "goodbye, ${noun}"},
	{"(?P<x>hi)|(?P<x>bye)", "$x$x$x", "hi", "$x$x$x"},
	{"(?P<x>hi)|(?P<x>bye)", "$x$x$x", "bye", "$x$x$x"},
	{"(?P<x>hi)|(?P<x>bye)", "$xyz", "hi", "$xyz"},
	{"(?P<x>hi)|(?P<x>bye)", "${x}yz", "hi", "${x}yz"},
	{"(?P<x>hi)|(?P<x>bye)", "hello $$x", "hi", "hello $$x"},
	{"a+", "${oops", "aaa", "${oops"},
	{"a+", "$$", "aaa", "$$"},
	{"a+", "$", "aaa", "$"},
}

type ReplaceFuncTest struct {
	pattern       string
	replacement   func(string) string
	input, output string
}

var replaceFuncTests = []ReplaceFuncTest{
	{"[a-c]", func(s string) string { return "x" + s + "y" }, "defabcdef", "defxayxbyxcydef"},
	{"[a-c]+", func(s string) string { return "x" + s + "y" }, "defabcdef", "defxabcydef"},
	{"[a-c]*", func(s string) string { return "x" + s + "y" }, "defabcdef", "xydxyexyfxabcydxyexyfxy"},
}

func TestReplaceAll(t *testing.T) {
	for _, tc := range replaceTests {
		re, err := Compile(tc.pattern)
		if err != nil {
			t.Errorf("Unexpected error compiling %q: %v", tc.pattern, err)
			continue
		}
		actual := re.ReplaceAllString(tc.input, tc.replacement)
		if actual != tc.output {
			t.Errorf("%q.ReplaceAllString(%q,%q) = %q; want %q",
				tc.pattern, tc.input, tc.replacement, actual, tc.output)
		}
		// now try bytes
		actual = string(re.ReplaceAll([]byte(tc.input), []byte(tc.replacement)))
		if actual != tc.output {
			t.Errorf("%q.ReplaceAll(%q,%q) = %q; want %q",
				tc.pattern, tc.input, tc.replacement, actual, tc.output)
		}
	}
}

func TestReplaceAllLiteral(t *testing.T) {
	// Run ReplaceAll tests that do not have $ expansions.
	for _, tc := range replaceTests {
		if strings.Contains(tc.replacement, "$") {
			continue
		}
		re, err := Compile(tc.pattern)
		if err != nil {
			t.Errorf("Unexpected error compiling %q: %v", tc.pattern, err)
			continue
		}
		actual := re.ReplaceAllLiteralString(tc.input, tc.replacement)
		if actual != tc.output {
			t.Errorf("%q.ReplaceAllLiteralString(%q,%q) = %q; want %q",
				tc.pattern, tc.input, tc.replacement, actual, tc.output)
		}
		// now try bytes
		actual = string(re.ReplaceAllLiteral([]byte(tc.input), []byte(tc.replacement)))
		if actual != tc.output {
			t.Errorf("%q.ReplaceAllLiteral(%q,%q) = %q; want %q",
				tc.pattern, tc.input, tc.replacement, actual, tc.output)
		}
	}

	// Run literal-specific tests.
	for _, tc := range replaceLiteralTests {
		re, err := Compile(tc.pattern)
		if err != nil {
			t.Errorf("Unexpected error compiling %q: %v", tc.pattern, err)
			continue
		}
		actual := re.ReplaceAllLiteralString(tc.input, tc.replacement)
		if actual != tc.output {
			t.Errorf("%q.ReplaceAllLiteralString(%q,%q) = %q; want %q",
				tc.pattern, tc.input, tc.replacement, actual, tc.output)
		}
		// now try bytes
		actual = string(re.ReplaceAllLiteral([]byte(tc.input), []byte(tc.replacement)))
		if actual != tc.output {
			t.Errorf("%q.ReplaceAllLiteral(%q,%q) = %q; want %q",
				tc.pattern, tc.input, tc.replacement, actual, tc.output)
		}
	}
}

func TestReplaceAllFunc(t *testing.T) {
	for _, tc := range replaceFuncTests {
		re, err := Compile(tc.pattern)
		if err != nil {
			t.Errorf("Unexpected error compiling %q: %v", tc.pattern, err)
			continue
		}
		actual := re.ReplaceAllStringFunc(tc.input, tc.replacement)
		if actual != tc.output {
			t.Errorf("%q.ReplaceFunc(%q,fn) = %q; want %q",
				tc.pattern, tc.input, actual, tc.output)
		}
		// now try bytes
		actual = string(re.ReplaceAllFunc([]byte(tc.input), func(s []byte) []byte { return []byte(tc.replacement(string(s))) }))
		if actual != tc.output {
			t.Errorf("%q.ReplaceFunc(%q,fn) = %q; want %q",
				tc.pattern, tc.input, actual, tc.output)
		}
	}
}

// End copied code
