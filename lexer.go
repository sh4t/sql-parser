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
	ItemError ItemType = iota // error occurred; value is text of error
	ItemEOF
	ItemWhitespace
	ItemSingleLineComment
	ItemMultiLineComment
	ItemKeyword        // SQL language keyword like SELECT, INSERT, etc.
	ItemOperator       // operators like '=', '<>', etc.
	ItemStar           // *: identifier that matches every column in a table
	ItemIdentifier     // alphanumeric identifier or complex identifier like `a.b` and `c.*`
	ItemLeftParen      // '('
	ItemNumber         // simple number, including imaginary
	ItemRightParen     // ')'
	ItemSpace          // run of spaces separating arguments
	ItemString         // quoted string (includes quotes)
	ItemComment        // comments
	ItemStatementStart // start of a statement like SELECT
	ItemStetementEnd   // ';'
	// etc.
	//TODO: enumerate all item types
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

	r, _, _ := l.input.ReadRune()
	//TODO: handle EOF, panic on other errors
	l.buffer = append(l.buffer, r)
	l.bufferPos++
	return r
}

// peek() returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	if l.bufferPos < len(l.buffer) {
		return l.buffer[l.bufferPos]
	}

	r, _, _ := l.input.ReadRune()
	//TODO: handle EOF, panic on other errors

	l.buffer = append(l.buffer, r)
	return r
}

// peek() returns but does not consume the next few runes in the input.
func (l *Lexer) peekNext(length int) string {
	lenDiff := l.bufferPos + length - len(l.buffer)
	if lenDiff > 0 {
		for i := 0; i < lenDiff; i++ {
			r, _, _ := l.input.ReadRune()
			//TODO: handle EOF, panic on other errors
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
func (l *Lexer) accept(valid string) bool {
	r := l.next()
	if strings.IndexRune(valid, r) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptWhile consumes runes while the specified condition is true
func (l *Lexer) acceptWhile(fn ValidatorFn) {
	r := l.next()
	for fn(r) {
		r = l.next()
	}
	l.backup()
}

// acceptUntil consumes runes until the specified contidtion is met
func (l *Lexer) acceptUntil(fn ValidatorFn) {
	r := l.next()
	for !fn(r) {
		r = l.next()
	}
	l.backup()
}

// acceptUntil consumes runes until the specified string is met
func (l *Lexer) acceptUntilMatch(match string) {
	length := len(match)
	next := l.peekNext(length)
	for next != match {
		l.next()
		next = l.peekNext(length)
	}
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
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
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

	case next == '*':
		//TODO: determine if this is neccessary or should be classified as an identifier
		l.next()
		l.emit(ItemStar)
		return lexWhitespace

	case next == '(':
		l.next()
		l.emit(ItemLeftParen)
		return lexWhitespace

	case next == ')':
		l.next()
		l.emit(ItemRightParen)
		return lexWhitespace

	/*
		//TODO: finish different cases
		case next == '*':
			return lexStar
		case next == '`':
			return lexIdentifier
		case next == '"' || next == '\'':
			return lexString
		case next == '+' || next == '-' || ('0' <= next && next <= '9'):
			return lexNumber
	*/

	case isAlphaNumeric(next):
		return lexKeyWordOrIdentifier

	default:
		//TODO: enable panic :)
		//panic(fmt.Sprintf("don't know what to do with: %q", next))
		l.emit(ItemEOF)
		return nil
	}
}

func lexSingleLineComment(l *Lexer) StateFn {
	l.acceptUntil(isEndOfLine)
	l.emit(ItemSingleLineComment)
	return lexWhitespace
}

func lexMultiLineComment(l *Lexer) StateFn {
	l.acceptUntilMatch("*/")
	l.next()
	l.next()
	l.emit(ItemMultiLineComment)
	return lexWhitespace
}

func lexKeyWordOrIdentifier(l *Lexer) StateFn {
	l.acceptWhile(isAlphaNumeric)
	l.emit(ItemIdentifier)
	//TODO: determine whether this is a keyword
	return lexWhitespace
}
