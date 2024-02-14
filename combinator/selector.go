package combinator

import (
	"go/ast"
	"go/types"

	"github.com/goghcrow/go-matcher"
)

func SelectorOf(m *Matcher, p Predicate[*ast.SelectorExpr]) ExprPattern {
	return matcher.MkPattern[ExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || sel == nil {
			return false
		}
		return p(ctx, sel)
	})
}

func SelectorObjectOf(m *Matcher, p Predicate[types.Object]) ExprPattern {
	return SelectorOf(m, func(ctx *MatchCtx, sel *ast.SelectorExpr) bool {
		obj := ctx.ObjectOf(sel.Sel)
		if obj == nil {
			return false
		}
		return p(ctx, obj)
	})
}

// SelectorPkgOf Assume X is ident
func SelectorPkgOf(m *Matcher, p Predicate[*types.Package]) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: IdentObjectOf(m, func(ctx *MatchCtx, obj types.Object) bool {
			pkg, ok := obj.(*types.PkgName)
			return ok && p(ctx, pkg.Imported())
		}),
	}
}

// // SelectorTypeOf Assume X is ident
// func SelectorTypeOf(m *Matcher, p Predicate[*types.TypeName]) *ast.SelectorExpr {
// 	return &ast.SelectorExpr{
// 		X: IdentObjectOf(m, func(ctx *MatchCtx, obj types.Object) bool {
// 			ty, ok := obj.(*types.TypeName)
// 			return ok && p(ctx, ty)
// 		}),
// 	}
// }

// SelectorTypeOf
// e.g., struct{x int}.x
// so can't use IdentObjectOf
func SelectorTypeOf(m *Matcher, p Predicate[types.Type]) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: TypeOf[ExprPattern](m, p),
	}
}

func SelectorStructOf(m *Matcher, p Predicate[*types.Struct]) *ast.SelectorExpr {
	return SelectorTypeOf(m, func(ctx *MatchCtx, ty types.Type) bool {
		assert(ty != nil, "invalid")
		st, ok := ty.Underlying().(*types.Struct)
		if !ok {
			return false
		}
		return p(ctx, st)
	})
}

func SelectorOfStructField(m *Matcher, pStruct Predicate[*types.Struct], pField Predicate[*types.Var]) ExprPattern {
	// must be struct field selector
	return AndEx[ExprPattern](m,
		SelectorStructOf(m, pStruct),
		SelectorObjectOf(m, func(ctx *MatchCtx, obj types.Object) bool {
			tv, ok := obj.(*types.Var)
			return ok && tv.IsField() && pField(ctx, tv)
		}),
		// matcher.MkPattern[ExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		// 	sel, _ := n.(*ast.SelectorExpr) // has confirmed
		// 	assert(sel != nil && m.Selections[sel] != nil, "invalid")
		// 	tv, ok := m.Selections[sel].Obj().(*types.Var)
		// 	return ok && tv.IsField() && pField(tv)
		// }),
	)
}
