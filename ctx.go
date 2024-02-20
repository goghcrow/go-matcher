package matcher

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type (
	Package    = packages.Package
	PatternVar = string
	Binds      map[PatternVar]ast.Node

	MatchCtx struct {
		Pkg     *Package
		Stack   []ast.Node
		Names   []string // stack parent fileld name
		Binds   Binds
		Matcher *Matcher
	}
	MatchFun func(n ast.Node, ctx *MatchCtx) bool
)

func newMCtx(m *Matcher, pkg *Package, stack []ast.Node, names []string) *MatchCtx {
	return &MatchCtx{
		Matcher: m,
		Pkg:     pkg,
		Stack:   stack,
		Binds:   map[PatternVar]ast.Node{},
	}
}

func (c *MatchCtx) match(x, y ast.Node) bool            { return c.Matcher.match(x, y, c) }
func (c *MatchCtx) Match(ptn, node ast.Node, f Matched) { c.Matcher.Match(c.Pkg, ptn, node, f) }
func (c *MatchCtx) Matched(ptn, root ast.Node) bool     { return c.Matcher.Matched(c.Pkg, ptn, root) }

func (c *MatchCtx) ShowNode(n ast.Node) string {
	fset := c.Pkg.Fset
	if runningWithGoTest {
		return ShowNode(fset, n)
	}
	return ShowNode(fset, n) + "\nat " + fset.Position(n.Pos()).String()
}

func (c *MatchCtx) TypeInfo() *types.Info               { return c.Pkg.TypesInfo }
func (c *MatchCtx) ObjectOf(id *ast.Ident) types.Object { return c.TypeInfo().ObjectOf(id) }
func (c *MatchCtx) TypeOf(e ast.Expr) types.Type        { return c.TypeInfo().TypeOf(e) }
func (c *MatchCtx) Callee(call *ast.CallExpr) types.Object {
	return typeutil.Callee(c.TypeInfo(), call)
}

func (c *MatchCtx) UpdateType(e ast.Expr, t types.Type) {
	c.TypeInfo().Types[e] = types.TypeAndValue{Type: t}
}

func (c *MatchCtx) UpdateUses(idOrSel ast.Expr, obj types.Object) {
	info := c.TypeInfo()
	switch x := idOrSel.(type) {
	case *ast.Ident:
		info.Uses[x] = obj
	case *ast.SelectorExpr:
		info.Uses[x.Sel] = obj
	default:
		panic("unreached")
	}
}

func (c *MatchCtx) UpdateDefs(idOrSel ast.Expr, obj types.Object) {
	info := c.TypeInfo()
	switch x := idOrSel.(type) {
	case *ast.Ident:
		info.Defs[x] = obj
	case *ast.SelectorExpr:
		info.Defs[x.Sel] = obj
	default:
		panic("unreached")
	}
}

func (c *MatchCtx) CopyTypeInfo(new, old ast.Expr) {
	info := c.TypeInfo()
	//goland:noinspection GoReservedWordUsedAsName
	switch new := new.(type) {
	case *ast.Ident:
		orig := old.(*ast.Ident)
		if obj, ok := info.Defs[orig]; ok {
			info.Defs[new] = obj
		}
		if obj, ok := info.Uses[orig]; ok {
			info.Uses[new] = obj
		}

	case *ast.SelectorExpr:
		orig := old.(*ast.SelectorExpr)
		if sel, ok := info.Selections[orig]; ok {
			info.Selections[new] = sel
		}
	}

	if tv, ok := info.Types[old]; ok {
		info.Types[new] = tv
	}
}

func (c *MatchCtx) NewIdent(name string, t types.Type) *ast.Ident {
	ident := ast.NewIdent(name)
	c.UpdateType(ident, t)

	obj := types.NewVar(token.NoPos, c.Pkg.Types, name, t)
	c.UpdateUses(ident, obj)
	return ident
}
