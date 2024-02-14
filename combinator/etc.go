package combinator

import (
	"go/ast"
	"go/token"

	"github.com/goghcrow/go-matcher"
)

type (
	Matcher  = matcher.Matcher
	MatchFun = matcher.MatchFun
	MatchCtx = matcher.MatchCtx

	Pattern       = matcher.Pattern
	TypingPattern = matcher.TypingPattern
	SlicePattern  = matcher.SlicePattern
	ElemPattern   = matcher.ElemPattern

	NodePattern      = matcher.NodePattern
	StmtPattern      = matcher.StmtPattern
	RestStmtPattern  = matcher.RestStmtPattern
	ExprPattern      = matcher.ExprPattern
	RestExprPattern  = matcher.RestExprPattern
	DeclPattern      = matcher.DeclPattern
	SpecPattern      = matcher.SpecPattern
	IdentPattern     = matcher.IdentPattern
	FieldPattern     = matcher.FieldPattern
	FieldListPattern = matcher.FieldListPattern
	CallExprPattern  = matcher.CallExprPattern
	FuncTypePattern  = matcher.FuncTypePattern
	BlockStmtPattern = matcher.BlockStmtPattern
	BasicLitPattern  = matcher.BasicLitPattern
	TokenPattern     = matcher.TokenPattern
	StmtsPattern     = matcher.StmtsPattern
	ExprsPattern     = matcher.ExprsPattern
	SpecsPattern     = matcher.SpecsPattern
	IdentsPattern    = matcher.IdentsPattern
	FieldsPattern    = matcher.FieldsPattern

	FunNode    = matcher.FunNode
	StmtsNode  = matcher.StmtsNode
	ExprsNode  = matcher.ExprsNode
	SpecsNode  = matcher.SpecsNode
	IdentsNode = matcher.IdentsNode
	FieldsNode = matcher.FieldsNode
	TokenNode  = matcher.TokenNode
)

type (
	Predicate[T any] func(*MatchCtx, T) bool
)

func PtrOf(expr ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{
		Op: token.AND,
		X:  expr,
	}
}

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}
