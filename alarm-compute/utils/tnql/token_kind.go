package tnql

// TokenKind TODO
/*
	Represents all valid types of tokens that a token can be.
*/
type TokenKind int

const (
	// UNKNOWN TODO
	UNKNOWN TokenKind = iota

	// PREFIX TODO
	PREFIX
	// NUMERIC TODO
	NUMERIC
	// BOOLEAN TODO
	BOOLEAN
	// STRING TODO
	STRING
	// PATTERN TODO
	PATTERN
	// TIME TODO
	TIME
	// VARIABLE TODO
	VARIABLE
	// FUNCTION TODO
	FUNCTION
	// SEPARATOR TODO
	SEPARATOR
	// ACCESSOR TODO
	ACCESSOR

	// COMPARATOR TODO
	COMPARATOR
	// LOGICALOP TODO
	LOGICALOP
	// MODIFIER TODO
	MODIFIER

	// CLAUSE TODO
	CLAUSE
	// CLAUSE_CLOSE TODO
	CLAUSE_CLOSE

	// TERNARY TODO
	TERNARY
)

// String 用于打印
/*
	GetTokenKindString returns a string that describes the given TokenKind.
	e.g., when passed the NUMERIC TokenKind, this returns the string "NUMERIC".
*/
func (kind TokenKind) String() string {

	switch kind {

	case PREFIX:
		return "PREFIX"
	case NUMERIC:
		return "NUMERIC"
	case BOOLEAN:
		return "BOOLEAN"
	case STRING:
		return "STRING"
	case PATTERN:
		return "PATTERN"
	case TIME:
		return "TIME"
	case VARIABLE:
		return "VARIABLE"
	case FUNCTION:
		return "FUNCTION"
	case SEPARATOR:
		return "SEPARATOR"
	case COMPARATOR:
		return "COMPARATOR"
	case LOGICALOP:
		return "LOGICALOP"
	case MODIFIER:
		return "MODIFIER"
	case CLAUSE:
		return "CLAUSE"
	case CLAUSE_CLOSE:
		return "CLAUSE_CLOSE"
	case TERNARY:
		return "TERNARY"
	case ACCESSOR:
		return "ACCESSOR"
	}

	return "UNKNOWN"
}
