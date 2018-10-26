package core

import (
	"github.com/project-flogo/core/data/expression"
)

// Expr is an expression with its original source
type Expr struct {
	source string
	expression.Expr
}

// NewExpr creates a new expression
func NewExpr(source string, expr expression.Expr) *Expr {
	return &Expr{
		source: source,
		Expr:   expr,
	}
}

// String gets the source of the expression
func (e *Expr) String() string {
	return e.source
}
