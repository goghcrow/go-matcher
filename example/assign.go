package example

import (
	"go/ast"
	"go/token"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfDefine(m *Matcher) ast.Node {
	return &ast.AssignStmt{
		Tok: token.DEFINE,
	}
}

func PatternOfAssign(m *Matcher) ast.Node {
	return &ast.AssignStmt{
		Tok: matcher.MkPattern[TokenPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
			tok := n.(TokenNode)
			// token.XXX_ASSIGN
			return token.Token(tok) != token.DEFINE
		}),
	}
}
