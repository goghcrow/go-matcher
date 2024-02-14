package combinator

import (
	"go/ast"

	"github.com/goghcrow/go-matcher"
)

type (
	NodeOrPtn = ast.Node

	Unary[T any]  func(T) T
	Binary[T any] func(T, T) T
)

// Not a must be Pattern, can't be node literal, means TryGetMatchFun(m, a) != nil
func Not[Ptn Pattern](m *Matcher, a Ptn) Ptn {
	return combine1[Ptn](m, a, func(a MatchFun) MatchFun {
		return func(n ast.Node, ctx *MatchCtx) bool {
			return !a(n, ctx)
		}
	})
}

func NotEx[T Pattern](m *Matcher, a NodeOrPtn) T {
	return combineEx1[T](m, a, func(a MatchFun) MatchFun {
		return func(n ast.Node, ctx *MatchCtx) bool {
			return !a(n, ctx)
		}
	})
}

// And lhs, rhs must be Pattern, can't be node literal, means TryGetMatchFun(m, l or r) != nil
func And[Ptn Pattern](m *Matcher, lhs, rhs Ptn) Ptn {
	return combine[Ptn](m, lhs, rhs, func(lhs, rhs MatchFun) MatchFun {
		return func(n ast.Node, ctx *MatchCtx) bool {
			return lhs(n, ctx) && rhs(n, ctx)
		}
	})
}

func AndEx[Ptn Pattern](m *Matcher, lhs, rhs NodeOrPtn) Ptn {
	return combineEx[Ptn](m, lhs, rhs, func(lhs, rhs MatchFun) MatchFun {
		return func(n ast.Node, ctx *MatchCtx) bool {
			return lhs(n, ctx) && rhs(n, ctx)
		}
	})
}

// Or lhs, rhs must be Pattern, can't be node literal, means TryGetMatchFun(m, l or r) != nil
func Or[Ptn Pattern](m *Matcher, lhs, rhs Ptn) Ptn {
	return combine[Ptn](m, lhs, rhs, func(lhs, rhs MatchFun) MatchFun {
		return func(n ast.Node, ctx *MatchCtx) bool {
			return lhs(n, ctx) || rhs(n, ctx)
		}
	})
}

func OrEx[Ptn Pattern](m *Matcher, lhs, rhs NodeOrPtn) Ptn {
	return combineEx[Ptn](m, lhs, rhs, func(lhs, rhs MatchFun) MatchFun {
		return func(n ast.Node, ctx *MatchCtx) bool {
			return lhs(n, ctx) || rhs(n, ctx)
		}
	})
}

func combine1[T Pattern](m *Matcher, a T, un Unary[MatchFun]) T {
	return matcher.MkPattern[T](m, un(
		matcher.MustGetMatchFun[T](m, a),
	))
}

func combine[T Pattern](m *Matcher, a, b T, bin Binary[MatchFun]) T {
	return matcher.MkPattern[T](m, bin(
		matcher.MustGetMatchFun[T](m, a),
		matcher.MustGetMatchFun[T](m, b),
	))
}

func combineEx1[T Pattern](m *Matcher, a NodeOrPtn, un Unary[MatchFun]) T {
	return matcher.MkPattern[T](m, un(
		matcher.TryGetOrMkMatchFun[T](m, a),
	))
}

func combineEx[T Pattern](m *Matcher, a, b ast.Node, bin Binary[MatchFun]) T {
	return matcher.MkPattern[T](m, bin(
		matcher.TryGetOrMkMatchFun[T](m, a),
		matcher.TryGetOrMkMatchFun[T](m, b),
	))
}
