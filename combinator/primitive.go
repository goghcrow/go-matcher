package combinator

import (
	"go/ast"

	"github.com/goghcrow/go-matcher"
)

// Wildcard is a pattern that matches any node
func Wildcard[T Pattern](m *Matcher) T {
	return matcher.MkPattern[T](m, func(n ast.Node, ctx *MatchCtx) bool { return true })
}

// Nil literal represents wildcard[T] for convenient, so a special Nil pattern needed
func Nil[T Pattern](m *Matcher) T {
	return matcher.MkPattern[T](m, func(n ast.Node, ctx *MatchCtx) bool {
		return matcher.IsNilNode(n)
	})
}

// Bind match node to variable, so can be retrieved from env in callback's arg
func Bind[T Pattern](m *Matcher, variable string, ptn T) T {
	return And(m, ptn, matcher.MkVar[T](m, variable))
}

// Any subtree node matched pattern
func Any[T Pattern](m *Matcher, nodeOrPtn ast.Node) T {
	return matcher.MkPattern[T](m, func(root ast.Node, ctx *MatchCtx) bool {
		return ctx.Matched(nodeOrPtn, root)
	})
}
