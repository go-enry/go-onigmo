package onigmo

import "testing"

// Code copied from https://github.com/golang/go/blob/go1.14/src/regexp/all_test.go#L79-L134

func matchTest(t *testing.T, test *FindTest) {
	re := compileTest(t, test.pat, "")
	if re == nil {
		return
	}
	m := re.MatchString(test.text)
	if m != (len(test.matches) > 0) {
		t.Errorf("MatchString failure on %s: %t should be %t", test, m, len(test.matches) > 0)
	}
	// now try bytes
	m = re.Match([]byte(test.text))
	if m != (len(test.matches) > 0) {
		t.Errorf("Match failure on %s: %t should be %t", test, m, len(test.matches) > 0)
	}
}

func TestMatch(t *testing.T) {
	for _, test := range findTests {
		matchTest(t, &test)
	}
}

func matchFunctionTest(t *testing.T, test *FindTest) {
	m, err := MatchString(test.pat, test.text)
	if err == nil {
		return
	}
	if m != (len(test.matches) > 0) {
		t.Errorf("Match failure on %s: %t should be %t", test, m, len(test.matches) > 0)
	}
}

func TestMatchFunction(t *testing.T) {
	for _, test := range findTests {
		matchFunctionTest(t, &test)
	}
}

func copyMatchTest(t *testing.T, test *FindTest) {
	re := compileTest(t, test.pat, "")
	if re == nil {
		return
	}
	m1 := re.MatchString(test.text)
	m2 := re.Copy().MatchString(test.text)
	if m1 != m2 {
		t.Errorf("Copied Regexp match failure on %s: original gave %t; copy gave %t; should be %t",
			test, m1, m2, len(test.matches) > 0)
	}
}

func TestCopyMatch(t *testing.T) {
	for _, test := range findTests {
		copyMatchTest(t, &test)
	}
}

// End copied code
