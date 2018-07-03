// Test GoAWK Lexer

package lexer_test

import (
	"strings"
	"testing"

	. "github.com/benhoyt/goawk/lexer"
)

func TestString(t *testing.T) {
	input :=
		"+ += && = : , -- / /= $ == >= > ++ { [ < ( <= ~ % %= * *= !~ ! != || ^ ^= ? } ] ) ; - -= " +
			"BEGIN break continue delete do else END exit for if in next print printf return while " +
			"atan2 cos exp gsub index int length log match rand sin split sprintf sqrt srand sub substr tolower toupper " +
			"x \"str\\n\" 1234\n " +
			"@ ."

	strs := make([]string, 0, LAST+1)
	seen := make([]bool, LAST+1)
	l := NewLexer([]byte(input))
	for {
		_, tok, _ := l.Scan()
		strs = append(strs, tok.String())
		seen[int(tok)] = true
		if tok == EOF {
			break
		}
	}
	output := strings.Join(strs, " ")

	expected :=
		"+ += && = : , -- / /= $ == >= > ++ { [ < ( <= ~ % %= * *= !~ ! != || ^ ^= ? } ] ) ; - -= " +
			"BEGIN break continue delete do else END exit for if in next print printf return while " +
			"atan2 cos exp gsub index int length log match rand sin split sprintf sqrt srand sub substr tolower toupper " +
			"<name> <string> <number> <newline> " +
			"<illegal> <illegal> <eof>"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}

	for i, s := range seen {
		if !s {
			t.Errorf("token %s (%d) not seen", Token(i), i)
		}
	}
}
