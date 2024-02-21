package matcher

import (
	"go/ast"
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
		Names:   names,
		Binds:   map[PatternVar]ast.Node{},
	}
}

func (c *MatchCtx) match(x, y ast.Node) bool            { return c.Matcher.match(x, y, c) }
func (c *MatchCtx) Match(ptn, node ast.Node, f Matched) { c.Matcher.Match(c.Pkg, ptn, node, f) }
func (c *MatchCtx) Matched(ptn, root ast.Node) bool     { return c.Matcher.Matched(c.Pkg, ptn, root) }

func (c *MatchCtx) TypeInfo() *types.Info                { return c.Pkg.TypesInfo }
func (c *MatchCtx) ObjectOf(id *ast.Ident) types.Object  { return c.TypeInfo().ObjectOf(id) }
func (c *MatchCtx) TypeOf(e ast.Expr) types.Type         { return c.TypeInfo().TypeOf(e) }
func (c *MatchCtx) Callee(cl *ast.CallExpr) types.Object { return typeutil.Callee(c.TypeInfo(), cl) }

func (c *MatchCtx) ShowPos(n ast.Node) string {
	return c.Pkg.Fset.Position(n.Pos()).String()
}
func (c *MatchCtx) ShowNode(n ast.Node) string {
	fset := c.Pkg.Fset
	if runningWithGoTest {
		return ShowNode(fset, n)
	}
	return ShowNode(fset, n) + "\nat " + fset.Position(n.Pos()).String()
}
