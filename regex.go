package rubex

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lonigmo
#include <stdlib.h>
#include "chelper.h"
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"unicode/utf8"
	"unsafe"
)

const numMatchStartSize = 4
const numReadBufferStartSize = 256

var mutex sync.Mutex

type NamedGroupInfo map[string]int

var _ compliance = &Regexp{}

// Regexp is the representation of a compiled regular expression. A Regexp is
// safe for concurrent use by multiple goroutines.
type Regexp struct {
	pattern  string
	option   int
	encoding C.OnigEncoding

	regex     C.OnigRegex
	errorInfo *C.OnigErrorInfo
	errorBuf  *C.char

	numCaptures       int32
	namedGroupInfo    NamedGroupInfo
	hasMetacharacters bool
}

// NewRegexp creates and initializes a new Regexp with the given pattern and option.
func NewRegexp(pattern string, options int) (*Regexp, error) {
	return newRegExp(pattern, C.ONIG_ENCODING_UTF8, options)
}

// NewRegexpASCII is equivalent to NewRegexp, but with the encoding restricted to ASCII.
func NewRegexpASCII(pattern string, options int) (*Regexp, error) {
	return newRegExp(pattern, C.ONIG_ENCODING_ASCII, options)
}

func newRegExp(pattern string, encoding C.OnigEncoding, options int) (*Regexp, error) {
	re := &Regexp{
		pattern:  pattern,
		encoding: encoding,
		option:   options,
	}

	runtime.SetFinalizer(re, (*Regexp).Free)
	return re, initRegexp(re)
}

func initRegexp(re *Regexp) error {
	patternCharPtr := C.CString(re.pattern)
	defer C.free(unsafe.Pointer(patternCharPtr))

	mutex.Lock()
	defer mutex.Unlock()

	errorCode := C.NewOnigRegex(patternCharPtr, C.int(len(re.pattern)), C.int(re.option), &re.regex, &re.encoding, &re.errorInfo, &re.errorBuf)
	if errorCode != C.ONIG_NORMAL {
		return errors.New(C.GoString(re.errorBuf))
	}

	re.numCaptures = int32(C.onig_number_of_captures(re.regex)) + 1
	re.namedGroupInfo = re.getNamedGroupInfo()
	re.hasMetacharacters = QuoteMeta(re.pattern) != re.pattern

	return nil
}

// Compile parses a regular expression and returns, if successful, a Regexp
// object that can be used to match against text.
func Compile(str string) (*Regexp, error) {
	return NewRegexp(str, ONIG_OPTION_DEFAULT)
}

// MustCompile is like Compile but panics if the expression cannot be parsed.
// It simplifies safe initialization of global variables holding compiled
// regular expressions.
func MustCompile(str string) *Regexp {
	regexp, error := NewRegexp(str, ONIG_OPTION_DEFAULT)
	if error != nil {
		panic("regexp: compiling " + str + ": " + error.Error())
	}

	return regexp
}

func CompileWithOption(str string, option int) (*Regexp, error) {
	return NewRegexp(str, option)
}

func MustCompileWithOption(str string, option int) *Regexp {
	regexp, error := NewRegexp(str, option)
	if error != nil {
		panic("regexp: compiling " + str + ": " + error.Error())
	}

	return regexp
}

// MustCompileASCII is equivalent to MustCompile, but with the encoding restricted to ASCII.
func MustCompileASCII(str string) *Regexp {
	regexp, error := NewRegexpASCII(str, ONIG_OPTION_DEFAULT)
	if error != nil {
		panic("regexp: compiling " + str + ": " + error.Error())
	}

	return regexp
}

func (re *Regexp) Free() {
	mutex.Lock()
	if re.regex != nil {
		C.onig_free(re.regex)
		re.regex = nil
	}
	mutex.Unlock()
	if re.errorInfo != nil {
		C.free(unsafe.Pointer(re.errorInfo))
		re.errorInfo = nil
	}
	if re.errorBuf != nil {
		C.free(unsafe.Pointer(re.errorBuf))
		re.errorBuf = nil
	}
}

func (re *Regexp) getNamedGroupInfo() NamedGroupInfo {
	numNamedGroups := int(C.onig_number_of_names(re.regex))
	// when any named capture exists, there is no numbered capture even if
	// there are unnamed captures.
	if numNamedGroups == 0 {
		return nil
	}

	namedGroupInfo := make(map[string]int)

	//try to get the names
	bufferSize := len(re.pattern) * 2
	nameBuffer := make([]byte, bufferSize)
	groupNumbers := make([]int32, numNamedGroups)
	bufferPtr := unsafe.Pointer(&nameBuffer[0])
	numbersPtr := unsafe.Pointer(&groupNumbers[0])

	length := int(C.GetCaptureNames(re.regex, bufferPtr, (C.int)(bufferSize), (*C.int)(numbersPtr)))
	if length == 0 {
		panic(fmt.Errorf("could not get the capture group names from %q", re.String()))
	}

	namesAsBytes := bytes.Split(nameBuffer[:length], ([]byte)(";"))
	if len(namesAsBytes) != numNamedGroups {
		panic(fmt.Errorf(
			"the number of named groups (%d) does not match the number names found (%d)",
			numNamedGroups, len(namesAsBytes),
		))
	}

	for i, nameAsBytes := range namesAsBytes {
		name := string(nameAsBytes)
		namedGroupInfo[name] = int(groupNumbers[i])
	}

	return namedGroupInfo
}

func (re *Regexp) find(b []byte, n int, offset int) []int {
	match := make([]int, re.numCaptures*2)

	if n == 0 {
		b = []byte{0}
	}

	bytesPtr := unsafe.Pointer(&b[0])

	// captures contains two pairs of ints, start and end, so we need list
	// twice the size of the capture groups.
	captures := make([]C.int, re.numCaptures*2)
	capturesPtr := unsafe.Pointer(&captures[0])

	var numCaptures int32
	numCapturesPtr := unsafe.Pointer(&numCaptures)

	pos := int(C.SearchOnigRegex(
		bytesPtr, C.int(n), C.int(offset), C.int(ONIG_OPTION_DEFAULT),
		re.regex, re.errorInfo, (*C.char)(nil), (*C.int)(capturesPtr), (*C.int)(numCapturesPtr),
	))

	if pos < 0 {
		return nil
	}

	if numCaptures <= 0 {
		panic("cannot have 0 captures when processing a match")
	}

	if re.numCaptures != numCaptures {
		panic(fmt.Errorf("expected %d captures but got %d", re.numCaptures, numCaptures))
	}

	for i := range captures {
		match[i] = int(captures[i])
	}

	return match
}

func getCapture(b []byte, beg int, end int) []byte {
	if beg < 0 || end < 0 {
		return nil
	}

	return b[beg:end:end]
}

func (re *Regexp) match(b []byte, n int, offset int) bool {
	if n == 0 {
		b = []byte{0}
	}

	bytesPtr := unsafe.Pointer(&b[0])
	pos := int(C.SearchOnigRegex(
		bytesPtr, C.int(n), C.int(offset), C.int(ONIG_OPTION_DEFAULT),
		re.regex, re.errorInfo, nil, nil, nil,
	))

	return pos >= 0
}

func (re *Regexp) findAll(b []byte, n int) [][]int {
	if n < 0 {
		n = len(b)
	}

	capture := make([][]int, 0, numMatchStartSize)
	var offset int
	for offset <= n {
		match := re.find(b, n, offset)
		if match == nil {
			break
		}

		capture = append(capture, match)

		// move offset to the ending index of the current match and prepare to
		// find the next non-overlapping match.
		offset = match[1]

		// if match[0] == match[1], it means the current match does not advance
		// the search. we need to exit the loop to avoid getting stuck here.
		if match[0] == match[1] {
			if offset < n && offset >= 0 {
				//there are more bytes, so move offset by a word
				_, width := utf8.DecodeRune(b[offset:])
				offset += width
			} else {
				//search is over, exit loop
				break
			}
		}
	}

	return capture
}

// NumSubexp returns the number of parenthesized subexpressions in this Regexp.
func (re *Regexp) NumSubexp() int {
	return (int)(C.onig_number_of_captures(re.regex))
}

func fillCapturedValues(repl []byte, _ []byte, capturedBytes map[string][]byte) []byte {
	replLen := len(repl)
	newRepl := make([]byte, 0, replLen*3)
	groupName := make([]byte, 0, replLen)

	var inGroupNameMode, inEscapeMode bool
	for index := 0; index < replLen; index++ {
		ch := repl[index]
		if inGroupNameMode && ch == byte('<') {
		} else if inGroupNameMode && ch == byte('>') {
			inGroupNameMode = false
			capBytes := capturedBytes[string(groupName)]
			newRepl = append(newRepl, capBytes...)
			groupName = groupName[:0] //reset the name
		} else if inGroupNameMode {
			groupName = append(groupName, ch)
		} else if inEscapeMode && ch <= byte('9') && byte('1') <= ch {
			capNumStr := string(ch)
			capBytes := capturedBytes[capNumStr]
			newRepl = append(newRepl, capBytes...)
		} else if inEscapeMode && ch == byte('k') && (index+1) < replLen && repl[index+1] == byte('<') {
			inGroupNameMode = true
			inEscapeMode = false
			index++ //bypass the next char '<'
		} else if inEscapeMode {
			newRepl = append(newRepl, '\\')
			newRepl = append(newRepl, ch)
		} else if ch != '\\' {
			newRepl = append(newRepl, ch)
		}
		if ch == byte('\\') || inEscapeMode {
			inEscapeMode = !inEscapeMode
		}
	}

	return newRepl
}

func (re *Regexp) replaceAll(src, repl []byte, replFunc func([]byte, []byte, map[string][]byte) []byte) []byte {
	srcLen := len(src)
	matches := re.findAll(src, srcLen)
	if len(matches) == 0 {
		return src
	}

	dest := make([]byte, 0, srcLen)
	for i, match := range matches {
		length := len(match) / 2
		capturedBytes := make(map[string][]byte)

		if re.namedGroupInfo == nil {
			for j := 0; j < length; j++ {
				capturedBytes[strconv.Itoa(j)] = getCapture(src, match[2*j], match[2*j+1])
			}
		} else {
			for name, j := range re.namedGroupInfo {
				capturedBytes[name] = getCapture(src, match[2*j], match[2*j+1])
			}
		}

		matchBytes := getCapture(src, match[0], match[1])
		newRepl := replFunc(repl, matchBytes, capturedBytes)
		prevEnd := 0
		if i > 0 {
			prevMatch := matches[i-1][:2]
			prevEnd = prevMatch[1]
		}

		if match[0] > prevEnd && prevEnd >= 0 && match[0] <= srcLen {
			dest = append(dest, src[prevEnd:match[0]]...)
		}

		dest = append(dest, newRepl...)
	}

	lastEnd := matches[len(matches)-1][1]
	if lastEnd < srcLen && lastEnd >= 0 {
		dest = append(dest, src[lastEnd:]...)
	}

	return dest
}

func (re *Regexp) String() string {
	return re.pattern
}

// LiteralPrefix it's not implemented. Panics when it called.
func (re *Regexp) LiteralPrefix() (prefix string, complete bool) {
	panic("function not implemented")
	return "", false
}

// Copy returns a new Regexp object copied from re.
func (re *Regexp) Copy() *Regexp {
	copy, _ := newRegExp(re.pattern, re.encoding, re.option)
	return copy
}

// Longest makes future searches prefer the leftmost-longest match.
// That is, when matching against text, the regexp returns a match that
// begins as early as possible in the input (leftmost), and among those
// it chooses a match that is as long as possible.
// This method modifies the Regexp and may not be called concurrently
// with any other methods.
func (re *Regexp) Longest() {
	re.option = re.option | ONIG_OPTION_FIND_LONGEST
	initRegexp(re)
}

func (re *Regexp) SubexpNames() []string {
	return nil
}
