package rubex

import "io"

// FindIndex returns a two-element slice of integers defining the location of
// the leftmost match in b of the regular expression. The match itself is at
// b[loc[0]:loc[1]]. A return value of nil indicates no match.
func (re *Regexp) FindIndex(b []byte) []int {
	match := re.find(b, len(b), 0)
	if len(match) == 0 {
		return nil
	}

	return match[:2]
}

// Find returns a slice holding the text of the leftmost match in b of the
// regular expression. A return value of nil indicates no match.
func (re *Regexp) Find(b []byte) []byte {
	loc := re.FindIndex(b)
	if loc == nil {
		return nil
	}

	return getCapture(b, loc[0], loc[1])
}

// FindString returns a string holding the text of the leftmost match in s of
// the regular expression. If there is no match, the return value is an empty
// string, but it will also be empty if the regular expression successfully
// matches an empty string. Use FindStringIndex or FindStringSubmatch if it is
// necessary to distinguish these cases.
func (re *Regexp) FindString(s string) string {
	mb := re.Find([]byte(s))
	if mb == nil {
		return ""
	}

	return string(mb)
}

// FindStringIndex returns a two-element slice of integers defining the location
// of the leftmost match in s of the regular expression. The match itself is at
// s[loc[0]:loc[1]]. A return value of nil indicates no match.
func (re *Regexp) FindStringIndex(s string) []int {
	return re.FindIndex([]byte(s))
}

// FindAllIndex is the 'All' version of FindIndex; it returns a slice of all
// successive matches of the expression, as defined by the 'All' description in
// the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllIndex(b []byte, n int) [][]int {
	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}

	return matches
}

// FindAll is the 'All' version of Find; it returns a slice of all successive
// matches of the expression, as defined by the 'All' description in the package
// comment. A return value of nil indicates no match.
func (re *Regexp) FindAll(b []byte, n int) [][]byte {
	matches := re.FindAllIndex(b, n)
	if matches == nil {
		return nil
	}

	matchBytes := make([][]byte, 0, len(matches))
	for _, match := range matches {
		matchBytes = append(matchBytes, getCapture(b, match[0], match[1]))
	}

	return matchBytes
}

// FindAllString is the 'All' version of FindString; it returns a slice of all
// successive matches of the expression, as defined by the 'All' description in
// the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllString(s string, n int) []string {
	b := []byte(s)
	matches := re.FindAllIndex(b, n)
	if matches == nil {
		return nil
	}

	matchStrings := make([]string, 0, len(matches))
	for _, match := range matches {
		m := getCapture(b, match[0], match[1])
		if m == nil {
			matchStrings = append(matchStrings, "")
		} else {
			matchStrings = append(matchStrings, string(m))
		}
	}

	return matchStrings

}

// FindAllStringIndex is the 'All' version of FindStringIndex; it returns a
// slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllStringIndex(s string, n int) [][]int {
	return re.FindAllIndex([]byte(s), n)
}

// FindSubmatchIndex returns a slice holding the index pairs identifying the
// leftmost match of the regular expression in b and the matches, if any, of its
// subexpressions, as defined by the 'Submatch' and 'Index' descriptions in the
// package comment. A return value of nil indicates no match.
func (re *Regexp) FindSubmatchIndex(b []byte) []int {
	match := re.find(b, len(b), 0)
	if len(match) == 0 {
		return nil
	}

	return match
}

// FindSubmatch returns a slice of slices holding the text of the leftmost match
// of the regular expression in b and the matches, if any, of its subexpressions,
// as defined by the 'Submatch' descriptions in the package comment. A return
// value of nil indicates no match.
func (re *Regexp) FindSubmatch(b []byte) [][]byte {
	match := re.FindSubmatchIndex(b)
	if match == nil {
		return nil
	}

	length := len(match) / 2
	if length == 0 {
		return nil
	}

	results := make([][]byte, 0, length)
	for i := 0; i < length; i++ {
		results = append(results, getCapture(b, match[2*i], match[2*i+1]))
	}

	return results
}

// FindStringSubmatch returns a slice of strings holding the text of the
// leftmost match of the regular expression in s and the matches, if any, of its
// subexpressions, as defined by the 'Submatch' description in the package
// comment. A return value of nil indicates no match.
func (re *Regexp) FindStringSubmatch(s string) []string {
	b := []byte(s)
	match := re.FindSubmatchIndex(b)
	if match == nil {
		return nil
	}

	length := len(match) / 2
	if length == 0 {
		return nil
	}

	results := make([]string, 0, length)
	for i := 0; i < length; i++ {
		cap := getCapture(b, match[2*i], match[2*i+1])
		if cap == nil {
			results = append(results, "")
		} else {
			results = append(results, string(cap))
		}
	}

	return results
}

// FindStringSubmatchIndex returns a slice holding the index pairs identifying
// the leftmost match of the regular expression in s and the matches, if any, of
// its subexpressions, as defined by the 'Submatch' and 'Index' descriptions in
// the package comment. A return value of nil indicates no match.
func (re *Regexp) FindStringSubmatchIndex(s string) []int {
	return re.FindSubmatchIndex([]byte(s))
}

// FindAllSubmatchIndex is the 'All' version of FindSubmatchIndex; it returns a
// slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllSubmatchIndex(b []byte, n int) [][]int {
	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}

	return matches
}

// FindAllSubmatch is the 'All' version of FindSubmatch; it returns a slice of
// all successive matches of the expression, as defined by the 'All' description
// in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllSubmatch(b []byte, n int) [][][]byte {
	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}

	allCapturedBytes := make([][][]byte, 0, len(matches))
	for _, match := range matches {
		length := len(match) / 2
		capturedBytes := make([][]byte, 0, length)
		for i := 0; i < length; i++ {
			capturedBytes = append(capturedBytes, getCapture(b, match[2*i], match[2*i+1]))
		}

		allCapturedBytes = append(allCapturedBytes, capturedBytes)
	}

	return allCapturedBytes
}

// FindAllStringSubmatch is the 'All' version of FindStringSubmatch; it returns
// a slice of all successive matches of the expression, as defined by the 'All'
// description in the package comment. A return value of nil indicates no match.
func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string {
	b := []byte(s)

	matches := re.findAll(b, n)
	if len(matches) == 0 {
		return nil
	}

	allCapturedStrings := make([][]string, 0, len(matches))
	for _, match := range matches {
		length := len(match) / 2
		capturedStrings := make([]string, 0, length)
		for i := 0; i < length; i++ {
			cap := getCapture(b, match[2*i], match[2*i+1])
			if cap == nil {
				capturedStrings = append(capturedStrings, "")
			} else {
				capturedStrings = append(capturedStrings, string(cap))
			}
		}

		allCapturedStrings = append(allCapturedStrings, capturedStrings)
	}

	return allCapturedStrings
}

// FindAllStringSubmatchIndex is the 'All' version of FindStringSubmatchIndex;
// it returns a slice of all successive matches of the expression, as defined
// by the 'All' description in the package comment. A return value of nil
// indicates no match.
func (re *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int {
	return re.FindAllSubmatchIndex([]byte(s), n)
}

// FindReaderIndex returns a two-element slice of integers defining the location
// of the leftmost match of the regular expression in text read from the
// RuneReader. The match text was found in the input stream at byte offset
// loc[0] through loc[1]-1. A return value of nil indicates no match.
//
// In contrast with the standard library implementation, the reader it's fully
// loaded in memory.
func (re *Regexp) FindReaderIndex(r io.RuneReader) []int {
	b, _ := readAll(r)
	return re.FindIndex(b)
}

// FindReaderSubmatchIndex returns a slice holding the index pairs identifying
// the leftmost match of the regular expression of text read by the RuneReader,
// and the matches, if any, of its subexpressions, as defined by the 'Submatch'
// and 'Index' descriptions in the package comment. A return value of nil
// indicates no match.
//
// In contrast with the standard library implementation, the reader it's fully
// loaded in memory.
func (re *Regexp) FindReaderSubmatchIndex(r io.RuneReader) []int {
	b, _ := readAll(r)
	return re.FindSubmatchIndex(b)
}
