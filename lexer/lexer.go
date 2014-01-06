
package Lexer

import (
	"io"
	"fmt"
)

// ItemType identifies the type of lex Items.
type ItemType int

// Item represents a token or text string returned from the scanner.
type Item struct {
	typ ItemType // The type of this Item.
	val string   // The value of this Item.
}

const (
	ItemError        ItemType = iota // error occurred; value is text of error
	ItemEOF
	ItemIdentifier // alphanumeric identifier not starting with '.'
	ItemLeftParen  // '('
	ItemNumber     // simple number, including imaginary
	ItemRightParen // ')'
	ItemSpace      // run of spaces separating arguments
	ItemString     // quoted string (includes quotes)
	ItemComment    // comments
	ItemStatementStart // start of a statement like SELECT
	ItemStetementEnd // ';'
	// etc.
)

// StateFn represents the state of the scanner as a function that returns the next state.
type StateFn func(*Lexer) StateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	name       string    // the name of the input; used only for error reports
	//TODO: maybe use a chan of runes?
	input      Reader    // the input source
	state      StateFn   // the next lexing function to enter
	//TODO: some way to remember current position, start of Item and last read witdh
	Items      chan Item // channel of scanned Items
	parenDepth int       // nesting depth of ( ) exprs
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	//TODO: implement
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	//TODO: implement
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	//TODO: implement
}

// emit passes an Item back to the client.
func (l *Lexer) emit(t ItemType) {
	//TODO: implement
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	//TODO: implement
}

// accept consumes the next rune if it's from the valid set.
func (l *Lexer) accept(valid string) bool {
	//TODO: implement
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) acceptRun(valid string) {
	//TODO: implement
}

// nextItem returns the next Item from the input.
func (l *Lexer) nextItem() Item {
	//TODO: implement
}

// lex creates a new scanner for the input string.
func lex(name, input, left, right string) *Lexer {
	//TODO: implement
}

// run runs the state machine for the Lexer.
func (l *Lexer) run() {
	//TODO: implement
}


//TODO: different state functions that correspond to different ItemTypes like:
// func lexComment(l *Lexer) StateFn {...}
// func lexSpace(l *Lexer) StateFn {...}
// func lexIdentifier(l *Lexer) StateFn {...}
// etc.