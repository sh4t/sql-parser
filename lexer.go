package lexer

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

//TODO: support other encodings besides utf-8 (conversion before the lexer?)

// ItemType identifies the type of lex Items.
type ItemType int

// Item represents a token or text string returned from the scanner.
type Item struct {
	t   ItemType // The type of this Item.
	pos int      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this Item.
}

const (
	ItemError ItemType = iota // error occurred; value is text of error
	ItemEOF
	ItemKeyword        // SQL language keyword like SELECT, INSERT, etc.
	ItemOperator       // operators like '=', '<>', etc.
	ItemIdentifier     // alphanumeric identifier
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

const eof = -1

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(*Lexer) StateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	input       bufio.Reader // the input source
	state       StateFn      // the next lexing function to enter
	pos         int          // current position in the input
	start       int          // start position of this item
	currentItem []rune       // a slice of runes that contains the currently lexed item
	Items       chan Item    // channel of scanned Items
}

// next returns the next rune in the input.
func (l *Lexer) next() (r rune, err error) {
	var rsize int
	r, rsize, err = l.input.ReadRune()
	l.pos += rsize
	l.currentItem = append(l.currentItem, r)
	return r, err
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() (r rune, err error) {
	r, _, err = l.input.ReadRune()
	if err != nil {
		l.input.UnreadRune()
	}
	return r, err
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	cl := len(l.currentItem)
	if cl < 1 {
		panic(errors.New("lexer: trying to backup in an empty item"))
	}

	l.pos -= utf8.RuneLen(l.currentItem[cl-1])
	l.input.UnreadRune()
}

// emit passes an Item back to the client.
func (l *Lexer) emit(t ItemType) {
	l.Items <- Item{t, l.start, string(l.currentItem)}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
	l.currentItem = l.currentItem[0:0]
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	if r, err := l.next(); err == nil && strings.IndexRune(valid, r) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	r, err := l.next()
	for err == nil && strings.IndexRune(valid, r) >= 0 {
		r, err = l.next()
	}
	l.backup()
}

// nextItem returns the next Item from the input.
func (l *Lexer) nextItem() Item {
	return <-l.Items
}

// lex creates a new scanner for the input string.
func lex(input io.Reader) *Lexer {
	l := &Lexer{
		input:       bufio.NewReader(input),
		currentItem: make([]rune, 0, 10),
		items:       make(chan Item),
	}
	go l.run()
	return l
}

// run runs the state machine for the Lexer.
func (l *Lexer) run() {
	for state := startState; state != nil; {
		state = state(lexer)
	}
}

//TODO: different state functions that correspond to different ItemTypes like:
// func lexComment(l *Lexer) StateFn {...}
// func lexSpace(l *Lexer) StateFn {...}
// func lexIdentifier(l *Lexer) StateFn {...}
// etc.
