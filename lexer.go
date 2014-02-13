package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

//TODO: support other encodings besides utf-8 (conversion before the lexer?)

// ItemType identifies the type of lex Items.
type ItemType int

// Item represents a token or text string returned from the scanner.
type Item struct {
	Type ItemType // The type of this Item.
	Pos  int      // The starting position, in bytes, of this item in the input string.
	Val  string   // The value of this Item.
}

const (
	ItemError             ItemType = iota // error occurred; value is text of error
	ItemEOF                               // end of the file
	ItemWhitespace                        // a run of spaces, tabs and newlines
	ItemSingleLineComment                 // A comment like --
	ItemMultiLineComment                  // A multiline comment like /* ... */
	ItemKeyword                           // SQL language keyword like SELECT, INSERT, etc.
	ItemIdentifier                        // alphanumeric identifier or complex identifier like `a.b` and `c`.*
	ItemOperator                          // operators like '=', '<>', etc.
	ItemLeftParen                         // '('
	ItemRightParen                        // ')'
	ItemComma                             // ','
	ItemDot                               // '.'
	ItemStetementEnd                      // ';'
	ItemNumber                            // simple number, including imaginary
	ItemString                            // quoted string (includes quotes)
)

const EOF = -1

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(*Lexer) StateFn

// ValidatorFn represents a function that is used to check whether a specific rune matches certain rules.
type ValidatorFn func(rune) bool

// Lexer holds the state of the scanner.
type Lexer struct {
	state             StateFn       // the next lexing function to enter
	input             io.RuneReader // the input source
	inputCurrentStart int           // start position of this item
	buffer            []rune        // a slice of runes that contains the currently lexed item
	bufferPos         int           // the current position in the buffer
	Items             chan Item     // channel of scanned Items
}

// next() returns the next rune in the input.
func (l *Lexer) next() rune {
	if l.bufferPos < len(l.buffer) {
		res := l.buffer[l.bufferPos]
		l.bufferPos++
		return res
	}

	r, _, err := l.input.ReadRune()
	if err == io.EOF {
		r = EOF
	} else if err != nil {
		panic(err)
	}

	l.buffer = append(l.buffer, r)
	l.bufferPos++
	return r
}

// peek() returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	if l.bufferPos < len(l.buffer) {
		return l.buffer[l.bufferPos]
	}

	r, _, err := l.input.ReadRune()
	if err == io.EOF {
		r = EOF
	} else if err != nil {
		panic(err)
	}

	l.buffer = append(l.buffer, r)
	return r
}

// peek() returns but does not consume the next few runes in the input.
func (l *Lexer) peekNext(length int) string {
	lenDiff := l.bufferPos + length - len(l.buffer)
	if lenDiff > 0 {
		for i := 0; i < lenDiff; i++ {
			r, _, err := l.input.ReadRune()
			if err == io.EOF {
				r = EOF
			} else if err != nil {
				panic(err)
			}

			l.buffer = append(l.buffer, r)
		}
	}

	return string(l.buffer[l.bufferPos : l.bufferPos+length])
}

// backup steps back one rune
func (l *Lexer) backup() {
	l.backupWith(1)
}

// backup steps back many runes
func (l *Lexer) backupWith(length int) {
	if l.bufferPos < length {
		panic(fmt.Errorf("lexer: trying to backup with %d when the buffer position is %d", length, l.bufferPos))
	}

	l.bufferPos -= length
}

// emit passes an Item back to the client.
func (l *Lexer) emit(t ItemType) {
	l.Items <- Item{t, l.inputCurrentStart, string(l.buffer[:l.bufferPos])}
	l.ignore()
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	itemByteLen := 0
	for i := 0; i < l.bufferPos; i++ {
		itemByteLen += utf8.RuneLen(l.buffer[i])
	}

	l.inputCurrentStart += itemByteLen
	l.buffer = l.buffer[l.bufferPos:] //TODO: check for memory leaks, maybe copy remaining items into a new slice?
	l.bufferPos = 0
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) int {
	r := l.next()
	if strings.IndexRune(valid, r) >= 0 {
		return 1
	}
	l.backup()
	return 0
}

// acceptWhile consumes runes while the specified condition is true
func (l *Lexer) acceptWhile(fn ValidatorFn) int {
	r := l.next()
	count := 0
	for fn(r) {
		r = l.next()
		count++
	}
	l.backup()
	return count
}

// acceptUntil consumes runes until the specified contidtion is met
func (l *Lexer) acceptUntil(fn ValidatorFn) int {
	r := l.next()
	count := 0
	for !fn(r) && r != EOF {
		r = l.next()
		count++
	}
	l.backup()
	return count
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) StateFn {
	l.Items <- Item{ItemError, l.inputCurrentStart, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next Item from the input.
func (l *Lexer) nextItem() Item {
	return <-l.Items
}

// lex creates a new scanner for the input string.
func Lex(input io.Reader) *Lexer {
	l := &Lexer{
		input:  bufio.NewReader(input),
		buffer: make([]rune, 0, 10),
		Items:  make(chan Item),
	}
	go l.run()
	return l
}

// run runs the state machine for the Lexer.
func (l *Lexer) run() {
	for state := lexWhitespace; state != nil; {
		state = state(l)
	}
	close(l.Items)
}

// isSpace reports whether r is a whitespace character (space or end of line).
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n' || r == EOF
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isOperator reports whether r is an operator.
func isOperator(r rune) bool {
	return r == '+' || r == '-' || r == '*' || r == '/' || r == '=' || r == '>' || r == '<' || r == '~' || r == '|' || r == '^' || r == '&' || r == '%'
}

func lexWhitespace(l *Lexer) StateFn {
	l.acceptWhile(isWhitespace)
	if l.bufferPos > 0 {
		l.emit(ItemWhitespace)
	}

	next := l.peek()
	nextTwo := l.peekNext(2)

	switch {
	case next == EOF:
		l.emit(ItemEOF)
		return nil

	case nextTwo == "--":
		return lexSingleLineComment

	case nextTwo == "/*":
		return lexMultiLineComment

	case next == '(':
		l.next()
		l.emit(ItemLeftParen)
		return lexWhitespace

	case next == ')':
		l.next()
		l.emit(ItemRightParen)
		return lexWhitespace

	case next == ',':
		l.next()
		l.emit(ItemComma)
		return lexWhitespace

	case next == ';':
		l.next()
		l.emit(ItemStetementEnd)
		return lexWhitespace

	case isOperator(next):
		return lexOperator

	case next == '"' || next == '\'':
		return lexString

	case ('0' <= next && next <= '9'):
		return lexNumber

	case isAlphaNumeric(next) || next == '`':
		return lexIdentifierOrKeyword

	default:
		l.errorf("don't know what to do with '%s'", nextTwo)
		return nil
	}
}

func lexSingleLineComment(l *Lexer) StateFn {
	l.acceptUntil(isEndOfLine)
	l.emit(ItemSingleLineComment)
	return lexWhitespace
}

func lexMultiLineComment(l *Lexer) StateFn {
	l.next()
	l.next()
	for {
		l.acceptUntil(func(r rune) bool { return r == '*' })
		if l.peekNext(2) == "*/" {
			l.next()
			l.next()
			l.emit(ItemMultiLineComment)
			return lexWhitespace
		}

		if l.peek() == EOF {
			l.errorf("reached EOF when looking for comment end")
			return nil
		}

		l.next()
	}
}

func lexOperator(l *Lexer) StateFn {
	l.acceptWhile(isOperator)
	l.emit(ItemOperator)
	return lexWhitespace
}

func lexNumber(l *Lexer) StateFn {
	count := 0
	count += l.acceptWhile(unicode.IsDigit)
	if l.accept(".") > 0 {
		count += 1 + l.acceptWhile(unicode.IsDigit)
	}
	if l.accept("eE") > 0 {
		count += 1 + l.accept("+-")
		count += l.acceptWhile(unicode.IsDigit)
	}

	if isAlphaNumeric(l.peek()) {
		// We were lexing an identifier all along - backup and pass the ball
		l.backupWith(count)
		return lexIdentifierOrKeyword
	}

	l.emit(ItemNumber)
	return lexWhitespace
}

func lexString(l *Lexer) StateFn {
	quote := l.next()

	for {
		n := l.next()

		if n == EOF {
			return l.errorf("unterminated quoted string")
		}
		if n == '\\' {
			//TODO: fix possible problems with NO_BACKSLASH_ESCAPES mode
			if l.peek() == EOF {
				return l.errorf("unterminated quoted string")
			}
			l.next()
		}

		if n == quote {
			if l.peek() == quote {
				l.next()
			} else {
				l.emit(ItemString)
				return lexWhitespace
			}
		}
	}

}

func lexIdentifierOrKeyword(l *Lexer) StateFn {
	for {
		s := l.next()

		if s == '`' {
			for {
				n := l.next()

				if n == EOF {
					return l.errorf("unterminated quoted string")
				} else if n == '`' {
					if l.peek() == '`' {
						l.next()
					} else {
						break
					}
				}
			}
			l.emit(ItemIdentifier)
		} else if isAlphaNumeric(s) {
			l.acceptWhile(isAlphaNumeric)

			//TODO: check whether token is a keyword or an identifier
			l.emit(ItemIdentifier)
		}

		l.acceptWhile(isWhitespace)
		if l.bufferPos > 0 {
			l.emit(ItemWhitespace)
		}

		if l.peek() != '.' {
			break
		}

		l.next()
		l.emit(ItemDot)
	}

	return lexWhitespace
}
