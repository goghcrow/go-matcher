package example

import (
	"go/ast"
	"go/types"
	"regexp"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfSelectorNameReg(name *regexp.Regexp, m *Matcher) *ast.CallExpr {
	// match *.XXX($args), and bind args to variable
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			// X: Wildcard[ExprPattern](m),
			Sel: IdentOf(m, func(ctx *MatchCtx, id *ast.Ident) bool {
				isFun := !ctx.TypeInfo().Types[id].IsType() // not type cast
				return isFun && name.MatchString(id.Name)
			}),
		},
		Args: matcher.MkVar[ExprsPattern](m, "args"),
	}
}

func PatternOfGetDBProxy(m *Matcher) *ast.CallExpr {
	fNames := regexp.MustCompile("^(GetDBProxy|GetDB)$")
	return PatternOfSelectorNameReg(fNames, m)
}

func SelectorOfPkgPath(m *Matcher, path string, sel IdentPattern) ExprPattern {
	return AndEx[ExprPattern](m,
		SelectorPkgOf(m, func(_ *MatchCtx, pkg *types.Package) bool {
			return pkg.Path() == path
		}),
		&ast.SelectorExpr{Sel: sel},
	)
}
