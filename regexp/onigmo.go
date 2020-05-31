// +build onigmo

package regexp

import (
	onigmo "github.com/go-enry/go-onigmo"
)

type Regexp = *onigmo.Regexp

func MustCompile(str string) Regexp {
	return onigmo.MustCompile(str)
}

func Compile(str string) (Regexp, error) {
	return onigmo.Compile(str)
}

func QuoteMeta(s string) string {
	return onigmo.QuoteMeta(s)
}
