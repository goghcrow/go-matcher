package combinator

import (
	"go/ast"
	"go/types"
	"strings"
)

func FuncDeclOf(m *Matcher, p Predicate[*types.Func]) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: IdentObjectOf(m, func(ctx *MatchCtx, obj types.Object) bool {
			return p(ctx, obj.(*types.Func))
		}),
	}
}

func FuncFullNameOf(m *Matcher, name string) *ast.FuncDecl {
	funName := func(fun *types.Func) string {
		return strings.ReplaceAll(fun.FullName(), "command-line-arguments.", "")
	}
	return FuncDeclOf(m, func(ctx *MatchCtx, ft *types.Func) bool {
		return funName(ft) == name
	})
}

func InitFunc(m *Matcher) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: And(m,
			IdentNameOf(m, "init"),
			IdentSigOf(m, func(ctx *MatchCtx, sig *types.Signature) bool {
				return sig.Recv() == nil && sig.Params() == nil && sig.Results() == nil
			}),
		),
	}
}
