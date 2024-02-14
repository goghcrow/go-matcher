package example

import (
	"go/ast"
	"go/types"
	"regexp"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfAppendWithNoValue(m *Matcher) ast.Node {
	return &ast.CallExpr{
		Fun: And(m,
			IdentNameOf(m, "append"),
			IsBuiltin(m),
		),
		Args: SliceLenEQ[ExprsPattern](m, 1),
	}
}

func PatternOfCallFunOrMethodWithSpecName(name string, m *Matcher) ast.Node {
	isSpecNameFun := IdentOf(m, func(ctx *MatchCtx, id *ast.Ident) bool {
		isFun := !ctx.TypeInfo().Types[id].IsType() // not type cast
		return isFun && id.Name == name
	})
	return &ast.CallExpr{
		Fun: Or(m,
			matcher.PatternOf[ExprPattern](m, isSpecNameFun),                         // cast id pattern to expr pattern
			matcher.PatternOf[ExprPattern](m, &ast.SelectorExpr{Sel: isSpecNameFun}), // cast to expr pattern
		),
	}
}

func PatternOfCallAtomicAdder(m *Matcher) ast.Node {
	adders := regexp.MustCompile("^(AddInt64|AddUintptr)$")
	return &ast.CallExpr{
		Fun: SelectorOfPkgPath(m, "sync/atomic", IdentNameMatch(m, adders)),
	}
}

func PatternOfAtomicSwapStructField(m *Matcher) ast.Node {
	// atomic.AddInt64(&structObject.field, *)
	return &ast.CallExpr{
		Fun: SelectorOfPkgPath(m, "sync/atomic", IdentNameMatch(m, regexp.MustCompile("^SwapInt64$"))),
		Args: []ast.Expr{
			PtrOf(SelectorOfStructField(m, func(_ *MatchCtx, t *types.Struct) bool {
				return true
			}, func(_ *MatchCtx, t *types.Var) bool {
				return true
			})),
			Wildcard[ExprPattern](m),
		},
	}
}

func PatternOfSecondArgIsCtx(m *Matcher, ctxIface *types.Interface) ast.Node {
	// ctxIface := l.MustLookup("context.Context").Type().Underlying().(*types.Interface)
	restWildcard := matcher.MkPattern[RestExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		// rest := n.(ExprsNode)
		return true
	})
	return &ast.CallExpr{
		Args: []ast.Expr{
			Wildcard[ExprPattern](m),
			TypeImplements[ExprPattern](m, ctxIface),
			restWildcard,
		},
	}
}

func PatternOfSecondStmtIsIf(m *Matcher) ast.Node {
	restWildcard := matcher.MkPattern[RestStmtPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		// rest := n.(StmtsNode)
		return true
	})
	return &ast.FuncDecl{
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				Wildcard[StmtPattern](m),
				&ast.IfStmt{},
				restWildcard,
			},
		},
	}
}

func PatternOfCallee(m *Matcher) ast.Node {
	return CalleeOf(m, func(_ *MatchCtx, f types.Object) bool {
		return f != nil
	})
}

func PatternOfBuiltinCallee(m *Matcher) ast.Node {
	return BuiltinCalleeOf(m, func(_ *MatchCtx, f *types.Builtin) bool {
		return f != nil
	})
}

func PatternOfVarCallee(m *Matcher) ast.Node {
	return VarCalleeOf(m, func(_ *MatchCtx, f *types.Var) bool {
		return f != nil
	})
}

func PatternOfFuncOrMethodCallee(m *Matcher) ast.Node {
	return FuncOrMethodCalleeOf(m, func(_ *MatchCtx, f *types.Func) bool {
		return f != nil
	})
}

func PatternOfFuncCallee(m *Matcher) ast.Node {
	return FuncCalleeOf(m, func(_ *MatchCtx, f *types.Func) bool {
		return f != nil
	})
}

func PatternOfMethodCallee(m *Matcher) ast.Node {
	return MethodCalleeOf(m, func(_ *MatchCtx, f *types.Func) bool {
		return f != nil
	})
}

func PatternOfStaticCallee(m *Matcher) ast.Node {
	return StaticCalleeOf(m, func(_ *MatchCtx, f *types.Func) bool {
		return f != nil
	})
}

func PatternOfIfaceCalleeOf(m *Matcher) ast.Node {
	return IfaceCalleeOf(m, func(_ *MatchCtx, f *types.Func) bool {
		return f != nil
	})
}
