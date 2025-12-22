package tnql

// OperatorSymbol TODO
/*
	Represents the valid symbols for operators.

*/
type OperatorSymbol int

const (
	// VALUE TODO
	VALUE OperatorSymbol = iota
	// LITERAL TODO
	LITERAL
	// NOOP TODO
	NOOP
	// EQ TODO
	EQ
	// NEQ TODO
	NEQ
	// GT TODO
	GT
	// LT TODO
	LT
	// GTE TODO
	GTE
	// LTE TODO
	LTE
	// REQ TODO
	REQ
	// NREQ TODO
	NREQ
	// IN TODO
	IN

	// AND TODO
	AND
	// OR TODO
	OR

	// PLUS TODO
	PLUS
	// MINUS TODO
	MINUS
	// BITWISE_AND TODO
	BITWISE_AND
	// BITWISE_OR TODO
	BITWISE_OR
	// BITWISE_XOR TODO
	BITWISE_XOR
	// BITWISE_LSHIFT TODO
	BITWISE_LSHIFT
	// BITWISE_RSHIFT TODO
	BITWISE_RSHIFT
	// MULTIPLY TODO
	MULTIPLY
	// DIVIDE TODO
	DIVIDE
	// MODULUS TODO
	MODULUS
	// EXPONENT TODO
	EXPONENT

	// NEGATE TODO
	NEGATE
	// INVERT TODO
	INVERT
	// BITWISE_NOT TODO
	BITWISE_NOT

	// TERNARY_TRUE TODO
	TERNARY_TRUE
	// TERNARY_FALSE TODO
	TERNARY_FALSE
	// COALESCE TODO
	COALESCE

	// FUNCTIONAL TODO
	FUNCTIONAL
	// ACCESS TODO
	ACCESS
	// SEPARATE TODO
	SEPARATE
)

type operatorPrecedence int

const (
	noopPrecedence operatorPrecedence = iota
	valuePrecedence
	functionalPrecedence
	prefixPrecedence
	exponentialPrecedence
	additivePrecedence
	bitwisePrecedence
	bitwiseShiftPrecedence
	multiplicativePrecedence
	comparatorPrecedence
	ternaryPrecedence
	logicalAndPrecedence
	logicalOrPrecedence
	separatePrecedence
)

func findOperatorPrecedenceForSymbol(symbol OperatorSymbol) operatorPrecedence {

	switch symbol {
	case NOOP:
		return noopPrecedence
	case VALUE:
		return valuePrecedence
	case EQ:
		fallthrough
	case NEQ:
		fallthrough
	case GT:
		fallthrough
	case LT:
		fallthrough
	case GTE:
		fallthrough
	case LTE:
		fallthrough
	case REQ:
		fallthrough
	case NREQ:
		fallthrough
	case IN:
		return comparatorPrecedence
	case AND:
		return logicalAndPrecedence
	case OR:
		return logicalOrPrecedence
	case BITWISE_AND:
		fallthrough
	case BITWISE_OR:
		fallthrough
	case BITWISE_XOR:
		return bitwisePrecedence
	case BITWISE_LSHIFT:
		fallthrough
	case BITWISE_RSHIFT:
		return bitwiseShiftPrecedence
	case PLUS:
		fallthrough
	case MINUS:
		return additivePrecedence
	case MULTIPLY:
		fallthrough
	case DIVIDE:
		fallthrough
	case MODULUS:
		return multiplicativePrecedence
	case EXPONENT:
		return exponentialPrecedence
	case BITWISE_NOT:
		fallthrough
	case NEGATE:
		fallthrough
	case INVERT:
		return prefixPrecedence
	case COALESCE:
		fallthrough
	case TERNARY_TRUE:
		fallthrough
	case TERNARY_FALSE:
		return ternaryPrecedence
	case ACCESS:
		fallthrough
	case FUNCTIONAL:
		return functionalPrecedence
	case SEPARATE:
		return separatePrecedence
	}

	return valuePrecedence
}

/*
Map of all valid comparators, and their string equivalents.
Used during parsing of expressions to determine if a symbol is, in fact, a comparator.
Also used during evaluation to determine exactly which comparator is being used.
*/
var comparatorSymbols = map[string]OperatorSymbol{
	"==": EQ,
	"!=": NEQ,
	">":  GT,
	">=": GTE,
	"<":  LT,
	"<=": LTE,
	"=~": REQ,
	"!~": NREQ,
	"in": IN,
}

// [Supporting negative values · Issue #93 · Knetic/govaluate](https://github.com/Knetic/govaluate/issues/93)
// fix A>-3, will be parse to ">-", should check comparatorSymbols at first token
// max len symbols in comparatorSymbols is 2 = tokenLen -1
var comparatorSymbolsRewind = map[string]struct{}{
	"==-": {},
	"!=-": {},
	">-":  {},
	">=-": {},
	"<-":  {},
	"<=-": {},
	"=~-": {},
	"!~-": {},
	"in-": {},
}

var logicalSymbols = map[string]OperatorSymbol{
	"&&": AND,
	"||": OR,
}

var bitwiseSymbols = map[string]OperatorSymbol{
	"^": BITWISE_XOR,
	"&": BITWISE_AND,
	"|": BITWISE_OR,
}

var bitwiseShiftSymbols = map[string]OperatorSymbol{
	">>": BITWISE_RSHIFT,
	"<<": BITWISE_LSHIFT,
}

var additiveSymbols = map[string]OperatorSymbol{
	"+": PLUS,
	"-": MINUS,
}

var multiplicativeSymbols = map[string]OperatorSymbol{
	"*": MULTIPLY,
	"/": DIVIDE,
	"%": MODULUS,
}

var exponentialSymbolsS = map[string]OperatorSymbol{
	"**": EXPONENT,
}

var prefixSymbols = map[string]OperatorSymbol{
	"-": NEGATE,
	"!": INVERT,
	"~": BITWISE_NOT,
}

var ternarySymbols = map[string]OperatorSymbol{
	"?":  TERNARY_TRUE,
	":":  TERNARY_FALSE,
	"??": COALESCE,
}

// this is defined separately from additiveSymbols et al because it's needed for parsing, not stage planning.
var modifierSymbols = map[string]OperatorSymbol{
	"+":  PLUS,
	"-":  MINUS,
	"*":  MULTIPLY,
	"/":  DIVIDE,
	"%":  MODULUS,
	"**": EXPONENT,
	"&":  BITWISE_AND,
	"|":  BITWISE_OR,
	"^":  BITWISE_XOR,
	">>": BITWISE_RSHIFT,
	"<<": BITWISE_LSHIFT,
}

var separatorSymbols = map[string]OperatorSymbol{
	",": SEPARATE,
}

// IsModifierType TODO
/*
	Returns true if this operator is contained by the given array of candidate symbols.
	False otherwise.
*/
func (this OperatorSymbol) IsModifierType(candidate []OperatorSymbol) bool {

	for _, symbolType := range candidate {
		if this == symbolType {
			return true
		}
	}

	return false
}

// String 用于打印
/*
	Generally used when formatting type check errors.
	We could store the stringified symbol somewhere else and not require a duplicated codeblock to translate
	OperatorSymbol to string, but that would require more memory, and another field somewhere.
	Adding operators is rare enough that we just stringify it here instead.
*/
func (this OperatorSymbol) String() string {

	switch this {
	case NOOP:
		return "NOOP"
	case VALUE:
		return "VALUE"
	case EQ:
		return "="
	case NEQ:
		return "!="
	case GT:
		return ">"
	case LT:
		return "<"
	case GTE:
		return ">="
	case LTE:
		return "<="
	case REQ:
		return "=~"
	case NREQ:
		return "!~"
	case AND:
		return "&&"
	case OR:
		return "||"
	case IN:
		return "in"
	case BITWISE_AND:
		return "&"
	case BITWISE_OR:
		return "|"
	case BITWISE_XOR:
		return "^"
	case BITWISE_LSHIFT:
		return "<<"
	case BITWISE_RSHIFT:
		return ">>"
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case MULTIPLY:
		return "*"
	case DIVIDE:
		return "/"
	case MODULUS:
		return "%"
	case EXPONENT:
		return "**"
	case NEGATE:
		return "-"
	case INVERT:
		return "!"
	case BITWISE_NOT:
		return "~"
	case TERNARY_TRUE:
		return "?"
	case TERNARY_FALSE:
		return ":"
	case COALESCE:
		return "??"
	}
	return ""
}
