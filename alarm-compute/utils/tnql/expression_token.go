package tnql

// ExpressionToken TODO
/*
	Represents a single parsed token.
*/
type ExpressionToken struct {
	Kind  TokenKind
	Value interface{}
}
