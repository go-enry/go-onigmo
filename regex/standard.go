// +build !oniguruma

package regex

import (
	"regexp"
)

type Regexp = *regexp.Regexp

func Compile(str string) (Regexp, error) {
	return regexp.Compile(str)
}

func MustCompile(str string) Regexp {
	return regexp.MustCompile(str)
}

func QuoteMeta(s string) string {
	return regexp.QuoteMeta(s)
}
