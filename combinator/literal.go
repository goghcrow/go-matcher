package combinator

import (
	"go/ast"
	"go/constant"
	"go/token"
	"reflect"
	"strconv"

	"github.com/goghcrow/go-matcher"
)

// Notice: BasicLit is an atomic Pattern,
// &ast.BasicLit{ Kind: token.INT } can be used for matching INT literal
// because zero Value is ambiguous, wildcard or zero value?

// Notice: LitXXXOf returns ExprPattern, so the type of callback param is ast.Expr

func LitKindOf(m *Matcher, kind token.Token) ExprPattern {
	return matcher.MkPattern[ExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		// Notice: ExprPattern returns, so param n of callback is ast.Expr
		// n is BasicLit expr and not nil
		lit, _ := n.(*ast.BasicLit)
		if lit == nil {
			return false
		}
		return lit.Kind == kind
	})
}

func LitOf(m *Matcher, kind token.Token, p Predicate[constant.Value]) ExprPattern {
	return matcher.MkPattern[ExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		lit, _ := n.(*ast.BasicLit)
		if lit == nil {
			return false
		}
		if lit.Kind != kind {
			return false
		}
		val := constant.MakeFromLiteral(lit.Value, kind, 0)
		return p(ctx, val)
	})
}

func LitIntOf(m *Matcher, p Predicate[constant.Value]) ExprPattern { return LitOf(m, token.INT, p) }

func LitFloatOf(m *Matcher, p Predicate[constant.Value]) ExprPattern {
	return LitOf(m, token.FLOAT, p)
}

func LitCharOf(m *Matcher, p Predicate[constant.Value]) ExprPattern {
	return LitOf(m, token.CHAR, p)
}

func LitStringOf(m *Matcher, p Predicate[constant.Value]) ExprPattern {
	return LitOf(m, token.STRING, p)
}

func LitStringValOf(m *Matcher, p Predicate[string]) ExprPattern {
	return matcher.MkPattern[ExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		lit, _ := n.(*ast.BasicLit)
		if lit == nil {
			return false
		}
		if lit.Kind != token.STRING {
			return false
		}
		val, _ := strconv.Unquote(lit.Value)
		return p(ctx, val)
	})
}

func TagOf(m *Matcher, p Predicate[*reflect.StructTag]) BasicLitPattern {
	return matcher.MkPattern[BasicLitPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		tagLit, _ := n.(*ast.BasicLit)
		if tagLit == nil {
			// return false
			return p(ctx, nil)
		}

		assert(tagLit.Kind == token.STRING, "")
		tag, _ := strconv.Unquote(tagLit.Value)
		structTag := reflect.StructTag(tag)
		return p(ctx, &structTag)
	})
}
