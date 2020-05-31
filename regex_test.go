// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rubex

import (
	"runtime"
	"strings"
	"testing"
)

// Code copied from https://github.com/golang/go/blob/go1.14/src/regexp/all_test.go#L15-L77

var goodRe = []string{
	``,
	`.`,
	`^.$`,
	`a`,
	`a*`,
	`a+`,
	`a?`,
	`a|b`,
	`a*|b*`,
	`(a*|b)(c*|d)`,
	`[a-z]`,
	`[a-abc-c\-\]\[]`,
	`[a-z]+`,
	`[abc]`,
	`[^1234]`,
	`[^\n]`,
	`\!\\`,
}

type stringError struct {
	re  string
	err string
}

var badRe = []stringError{
	{`*`, "missing argument to repetition operator: `*`"},
	{`+`, "missing argument to repetition operator: `+`"},
	{`?`, "missing argument to repetition operator: `?`"},
	{`(abc`, "missing closing ): `(abc`"},
	{`abc)`, "unexpected ): `abc)`"},
	{`x[a-z`, "missing closing ]: `[a-z`"},
	{`[z-a]`, "invalid character class range: `z-a`"},
	{`abc\`, "trailing backslash at end of expression"},
	{`a**`, "invalid nested repetition operator: `**`"},
	{`a*+`, "invalid nested repetition operator: `*+`"},
	{`\x`, "invalid escape sequence: `\\x`"},
}

func compileTest(t *testing.T, expr string, error string) *Regexp {
	re, err := Compile(expr)
	if error == "" && err != nil {
		t.Error("compiling `", expr, "`; unexpected error: ", err.Error())
	}
	if error != "" && err == nil {
		t.Error("compiling `", expr, "`; missing error")
	} else if error != "" && !strings.Contains(err.Error(), error) {
		t.Error("compiling `", expr, "`; wrong error: ", err.Error(), "; want ", error)
	}
	return re
}

func TestGoodCompile(t *testing.T) {
	for i := 0; i < len(goodRe); i++ {
		compileTest(t, goodRe[i], "")
	}
}

func TestBadCompile(t *testing.T) {
	for i := 0; i < len(badRe); i++ {
		compileTest(t, badRe[i].re, badRe[i].err)
	}
}

// End copied code

func runParallel(testFunc func(chan bool), concurrency int) {
	runtime.GOMAXPROCS(4)
	done := make(chan bool, concurrency)
	for i := 0; i < concurrency; i++ {
		go testFunc(done)
	}
	for i := 0; i < concurrency; i++ {
		<-done
		<-done
	}
	runtime.GOMAXPROCS(1)
}

const numConcurrentRuns = 200

func TestCompile_Parallel(t *testing.T) {
	testFunc := func(done chan bool) {
		done <- false
		for i := 0; i < len(goodRe); i++ {
			compileTest(t, goodRe[i], "")
		}
		done <- true
	}
	runParallel(testFunc, numConcurrentRuns)
}

func TestCompile_WithOption(t *testing.T) {
	re := MustCompileWithOption("a$", ONIG_OPTION_IGNORECASE)
	if !re.MatchString("A") {
		t.Errorf("expect to match\n")
	}
	re = MustCompile("a$")
	if re.MatchString("A") {
		t.Errorf("expect to mismatch\n")
	}

}

type numSubexpCase struct {
	input    string
	expected int
}

var numSubexpCases = []numSubexpCase{
	{``, 0},
	{`.*`, 0},
	{`abba`, 0},
	{`ab(b)a`, 1},
	{`ab(.*)a`, 1},
	{`(.*)ab(.*)a`, 2},
	{`(.*)(ab)(.*)a`, 3},
	{`(.*)((a)b)(.*)a`, 4},
	{`(.*)(\(ab)(.*)a`, 3},
	{`(.*)(\(a\)b)(.*)a`, 3},
}

func TestNumSubexp(t *testing.T) {
	for _, c := range numSubexpCases {
		re := MustCompile(c.input)
		n := re.NumSubexp()
		if n != c.expected {
			t.Errorf("NumSubexp for %q returned %d, expected %d", c.input, n, c.expected)
		}
	}
}

func TestLonggest(t *testing.T) {
	re := MustCompile(`a(|b)`)

	find := re.FindString("ab")
	if find != "a" {
		t.Errorf("expected match; got %s: %s", find, "a")
	}

	re.Longest()

	find = re.FindString("ab")
	if find != "ab" {
		t.Errorf("expected match; got %s: %s", find, "ab")
	}
}
