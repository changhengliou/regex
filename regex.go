package main

import (
	"fmt"
)

// *+?()|
// \+ for escape
// e1 matches s and e2 matches t, then e1|e2 matches s or t, and e1e2 matches st.
// The metacharacters *, +, and ? are repetition operators:
// e1* matches a sequence of zero or more (possibly different) strings, each of which match e1; e1+ matches one or more;
// e1? matches zero or one.
// The sequence of execution is first alternation, then concatenation, and finally the repetition operators.
// A backreference like \1 or \2 matches the string matched by a previous parenthesized expression,
// and only that string: (cat|dog)\1 matches catcat and dogdog but not catdog nor dogcat.
// A deterministic finite automaton (DFA), because in any state, each possible input letter leads to at most one new state

type state struct {
	state int
	next  *state
	next2 *state
	last  int
}

type stack struct {
	val []interface{}
}

func (s *stack) push(val interface{}) {
	s.val = append(s.val, val)
}

func (s *stack) pop() interface{} {
	val := s.val[len(s.val)-1]
	s.val = s.val[:len(s.val)-1]
	return val
}

func (s *stack) top() interface{} {
	return s.val[len(s.val)-1]
}

func (s *stack) isEmpty() bool {
	return len(s.val) == 0
}

// 1. Collation-related bracket symbols [==] [::] [..]
// 2. Escaped characters \
// 3. Character set (bracket expression) []
// 4. Grouping ()
// 5. Single-character-ERE duplication * + ? {m,n}
// 6. Concatenation
// 7. Anchoring ^$
// 8. Alternation |
const (
	LPAREN = iota + 1
	RPAREN
	ALTERN
	CONCAT
	KLEENE
)

// (a|b)*a
func parse(str string) []rune {
	parenCount := 0
	isConcat := false
	arr := make([]rune, 0)
	for _, c := range str {
		switch c {
		case '(':
			parenCount += 1
			isConcat = false
			arr = append(arr, LPAREN)
			continue
		case ')':
			parenCount -= 1
			arr = append(arr, RPAREN)
		case '*':
			arr = append(arr, KLEENE)
		case '|':
			arr = append(arr, ALTERN)
			isConcat = false
			continue
		default:
			if isConcat {
				arr = append(arr, CONCAT)
			}
			arr = append(arr, c)
		}
		isConcat = true
		if parenCount < 0 {
			panic("Unbalanced regular expression")
		}
	}
	return arr
}

// The compilation of the . that follows pops the two b NFA fragment from the stack and pushes an NFA fragment for the
// concatenation bb.. Each NFA fragment is defined by its start state and its outgoing arrow
// https://www.boost.org/doc/libs/1_56_0/libs/regex/doc/html/boost_regex/syntax/basic_extended.html#boost_regex.syntax.basic_extended.operator_precedence
func infixToPostfix(infix []rune) []rune {
	postfix := make([]rune, 0)
	opStack := &stack{}
	for _, c := range infix {
		switch {
		case c == LPAREN:
			opStack.push(c)
		case c == RPAREN:
			for opStack.top().(int32) != LPAREN {
				postfix = append(postfix, opStack.pop().(rune))
			}
			opStack.pop()
		case c == ALTERN:
			fallthrough
		case c == CONCAT:
			fallthrough
		case c == KLEENE:
			for !opStack.isEmpty() && c <= opStack.top().(int32) && opStack.top() != LPAREN {
				postfix = append(postfix, opStack.pop().(int32))
			}
			opStack.push(c)
		default:
			postfix = append(postfix, c)
		}
	}
	for !opStack.isEmpty() {
		postfix = append(postfix, opStack.pop().(rune))
	}
	return postfix
}

func main() {
	fmt.Println(infixToPostfix(parse("(ab)*c")))    // "ab.c*"
	fmt.Println(infixToPostfix(parse("(a(b|d))*"))) // "abd|.*"
	fmt.Println(infixToPostfix(parse("a(bb)+c")))   // "abb..c.+"
}
