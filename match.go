package rubex

import (
	"bytes"
	"io"
)

// MatchString reports whether the string s contains any match of the regular expression re.
func MatchString(pattern string, s string) (matched bool, error error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(s), nil
}

// Match reports whether the byte slice b contains any match of the regular
// expression re.
func (re *Regexp) Match(b []byte) bool {
	return re.match(b, len(b), 0)
}

// MatchString reports whether the string s contains any match of the regular
// expression re.
func (re *Regexp) MatchString(s string) bool {
	return re.Match([]byte(s))
}

// MatchReader reports whether the text returned by the RuneReader contains any
// match of the regular expression re.
//
// In contrast with the standard library implementation, the reader it's fully
// loaded in memory.
func (re *Regexp) MatchReader(r io.RuneReader) bool {
	b, _ := readAll(r)
	return re.Match(b)
}

func readAll(r io.RuneReader) ([]byte, error) {
	var buf bytes.Buffer
	for {
		rune, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if _, err := buf.WriteRune(rune); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
