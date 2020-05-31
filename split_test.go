// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package rubex

import (
	"reflect"
	"strings"
	"testing"
)

// Code copied from https://github.com/golang/go/blob/go1.14/src/regexp/all_test.go#L464-L515

var splitTests = []struct {
	s   string
	r   string
	n   int
	out []string
}{
	{"foo:and:bar", ":", -1, []string{"foo", "and", "bar"}},
	{"foo:and:bar", ":", 1, []string{"foo:and:bar"}},
	{"foo:and:bar", ":", 2, []string{"foo", "and:bar"}},
	{"foo:and:bar", "foo", -1, []string{"", ":and:bar"}},
	{"foo:and:bar", "bar", -1, []string{"foo:and:", ""}},
	{"foo:and:bar", "baz", -1, []string{"foo:and:bar"}},
	{"baabaab", "a", -1, []string{"b", "", "b", "", "b"}},
	{"baabaab", "a*", -1, []string{"b", "b", "b"}},
	{"baabaab", "ba*", -1, []string{"", "", "", ""}},
	{"foobar", "f*b*", -1, []string{"", "o", "o", "a", "r"}},
	{"foobar", "f+.*b+", -1, []string{"", "ar"}},
	{"foobooboar", "o{2}", -1, []string{"f", "b", "boar"}},
	{"a,b,c,d,e,f", ",", 3, []string{"a", "b", "c,d,e,f"}},
	{"a,b,c,d,e,f", ",", 0, nil},
	{",", ",", -1, []string{"", ""}},
	{",,,", ",", -1, []string{"", "", "", ""}},
	{"", ",", -1, []string{""}},
	{"", ".*", -1, []string{""}},
	{"", ".+", -1, []string{""}},
	{"", "", -1, []string{}},
	{"foobar", "", -1, []string{"f", "o", "o", "b", "a", "r"}},
	{"abaabaccadaaae", "a*", 5, []string{"", "b", "b", "c", "cadaaae"}},
	{":x:y:z:", ":", -1, []string{"", "x", "y", "z", ""}},
}

func TestSplit(t *testing.T) {
	for i, test := range splitTests {
		re, err := Compile(test.r)
		if err != nil {
			t.Errorf("#%d: %q: compile error: %s", i, test.r, err.Error())
			continue
		}

		split := re.Split(test.s, test.n)
		if !reflect.DeepEqual(split, test.out) {
			t.Errorf("#%d: %q: got %q; want %q", i, test.r, split, test.out)
		}

		if QuoteMeta(test.r) == test.r {
			strsplit := strings.SplitN(test.s, test.r, test.n)
			if !reflect.DeepEqual(split, strsplit) {
				t.Errorf("#%d: Split(%q, %q, %d): regexp vs strings mismatch\nregexp=%q\nstrings=%q", i, test.s, test.r, test.n, split, strsplit)
			}
		}
	}
}

// End copied code
