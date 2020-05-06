// +build oniguruma

package regex

import (
	rubex "github.com/go-enry/go-oniguruma"
)

type Regexp = *rubex.Regexp

func MustCompile(str string) Regexp {
	return rubex.MustCompileASCII(str)
}

func Compile(str string) (Regexp, error) {
	return rubex.Compile(str)
}

func QuoteMeta(s string) string {
	return rubex.QuoteMeta(s)
}
