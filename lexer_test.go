package lexer

import (
	"fmt"
	"testing"
)

func ExampleInsertLexing() {
	itemNames := map[ItemType]string{
		ItemError:          "error",
		ItemEOF:            "EOF",
		ItemKeyword:        "keyword",
		ItemOperator:       "operator",
		ItemIdentifier:     "identifier",
		ItemLeftParen:      "left_paren",
		ItemNumber:         "number",
		ItemRightParen:     "right_paren",
		ItemSpace:          "space",
		ItemString:         "string",
		ItemComment:        "comment",
		ItemStatementStart: "statement_start",
		ItemStetementEnd:   "statement_end",
	}

	ppItem = func(i Item) string {
		return fmt.Sprintf("%q('%q')", itemNames[i.t], t.val)
	}

	query := "SELECT * FROM `users` WHERE id = 15;"

	lexer := lex("testlexer", query)

	for {
		item, ok := <-lexer.Items
		if !ok {
			break
		}
		fmt.Println(ppItem(item))
	}

	// Output:
	// statement_start('')
	// keyword('SELECT')
	// identifier('*')
	// keyword('FROM')
	// identifier('`users`')
	// keyword('WHERE')
	// identifier('id')
	// operator('=')
	// number('15')
	// statement_start(';')
	// EOF('')
}
