package combinator

import (
	"go/ast"
	"go/types"

	"github.com/goghcrow/go-matcher"
)

func TypeOf[T TypingPattern](m *Matcher, p Predicate[types.Type]) T {
	return matcher.MkPattern[T](m, func(n ast.Node, ctx *MatchCtx) bool {
		// typeof(n) = ast.Expr | *ast.Ident
		// n maybe nil, e.g., const x = 1
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		expr := n.(ast.Expr)
		if expr == nil { // ast.Expr(nil)
			return false
		}
		exprTy := ctx.TypeOf(expr)
		// assert(ty != nil, "type not found: "+m.ShowNode(expr))
		// if ty == nil { return false }
		return p(ctx, exprTy)
	})
}

func TypeConvertibleTo[T TypingPattern](m *Matcher, ty types.Type) T {
	return TypeOf[T](m, func(ctx *MatchCtx, t types.Type) bool {
		return types.ConvertibleTo(t, ty)
	})
}

func TypeAssignableTo[T TypingPattern](m *Matcher, ty types.Type) T {
	return TypeOf[T](m, func(ctx *MatchCtx, t types.Type) bool {
		return types.AssignableTo(t, ty)
	})
}

func TypeIdentical[T TypingPattern](m *Matcher, ty types.Type) T {
	return TypeOf[T](m, func(ctx *MatchCtx, t types.Type) bool {
		return types.Identical(t, ty)
	})
}

func TypeIdenticalIgnoreTags[T TypingPattern](m *Matcher, ty types.Type) T {
	return TypeOf[T](m, func(ctx *MatchCtx, t types.Type) bool {
		return types.IdenticalIgnoreTags(t, ty)
	})
}

func TypeImplements[T TypingPattern](m *Matcher, iface *types.Interface) T {
	return TypeOf[T](m, func(ctx *MatchCtx, t types.Type) bool {
		return types.Implements(t, iface)
	})
}
