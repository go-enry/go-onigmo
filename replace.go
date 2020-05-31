package rubex

// ReplaceAll returns a copy of src, replacing matches of the Regexp with the
// replacement text repl. Inside repl, $ signs are interpreted as in Expand, so
// for instance $1 represents the text of the first submatch.
func (re *Regexp) ReplaceAll(src, repl []byte) []byte {
	return re.replaceAll(src, repl, fillCapturedValues)
}

// ReplaceAllFunc returns a copy of src in which all matches of the Regexp have
// been replaced by the return value of function repl applied to the matched
// byte slice. The replacement returned by repl is substituted directly, without
// using Expand.
func (re *Regexp) ReplaceAllFunc(src []byte, repl func([]byte) []byte) []byte {
	return re.replaceAll(src, nil, func(_ []byte, matchBytes []byte, _ map[string][]byte) []byte {
		return repl(matchBytes)
	})
}

// ReplaceAllString returns a copy of src, replacing matches of the Regexp with
// the replacement string repl. Inside repl, $ signs are interpreted as in
// Expand, so for instance $1 represents the text of the first submatch.
func (re *Regexp) ReplaceAllString(src, repl string) string {
	return string(re.ReplaceAll([]byte(src), []byte(repl)))
}

// ReplaceAllStringFunc returns a copy of src in which all matches of the Regexp
// have been replaced by the return value of function repl applied to the
// matched substring. The replacement returned by repl is substituted directly,
// without using Expand.
func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string {
	return string(re.replaceAll([]byte(src), nil, func(_ []byte, matchBytes []byte, _ map[string][]byte) []byte {
		return []byte(repl(string(matchBytes)))
	}))
}

// ReplaceAllLiteralString returns a copy of src, replacing matches of the Regexp
// with the replacement string repl. The replacement repl is substituted directly,
// without using Expand.
func (re *Regexp) ReplaceAllLiteralString(src, repl string) string {
	return string(re.replaceAll([]byte(src), nil, func(dst []byte, _ []byte, _ map[string][]byte) []byte {
		return append(dst, repl...)
	}))
}

// ReplaceAllLiteral returns a copy of src, replacing matches of the Regexp
// with the replacement bytes repl. The replacement repl is substituted directly,
// without using Expand.
func (re *Regexp) ReplaceAllLiteral(src, repl []byte) []byte {
	return re.replaceAll(src, nil, func(dst []byte, _ []byte, _ map[string][]byte) []byte {
		return append(dst, repl...)
	})
}
