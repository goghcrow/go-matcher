package matcher

import (
	"go/ast"
	"go/constant"
	"go/token"
	"reflect"

	"golang.org/x/tools/go/ast/astutil"
)

// WildcardIdent
// typed nil ast.Ident means wildcard ident
// e.g. x, y := ...
// &ast.AssignStmt{ Lhs: []ast.Expr{ WildcardIdent, WildcardIdent }, },
var WildcardIdent *ast.Ident

type (
	Cursor = astutil.Cursor
	// Matched
	// If pre-ordered, the callback can return bool to control whether to continue traversing the subtree,
	// But if pre-ordered, it may miss the case where the modified subtree satisfies the pattern.
	// Postorder also has type problems, and may need to be handled with a working-list
	Matched func(*Cursor, *MatchCtx)
	Matcher struct {
		*matchFuns
		MatchCallEllipsis bool
		UnparenExpr       bool
	}
)

func New() *Matcher {
	return &Matcher{matchFuns: &matchFuns{}}
}

func (m *Matcher) Match(inPkg *Package, pattern, node ast.Node, f Matched) {
	buildStack := m.mkStackBuilder(node)
	postOrder(node, func(c *astutil.Cursor) bool {
		n := c.Node()
		stack, names := buildStack(n)
		mctx := newMCtx(m, inPkg, stack, names)
		if m.match(pattern, n, mctx) {
			f(c, mctx)
		}
		return true
	})
}

// Matched when any subtree of rootNode matched pattern, return immediately
func (m *Matcher) Matched(inPkg *Package, pattern, rootNode ast.Node) (matched bool) {
	var abort = new(int)
	defer func() {
		if r := recover(); r != nil && r != abort {
			panic(r)
		}
	}()
	m.Match(inPkg, pattern, rootNode, func(*Cursor, *MatchCtx) {
		matched = true
		panic(abort)
	})
	return matched
}

type stackBuilder func(node ast.Node) ([]ast.Node, []string)

func (m *Matcher) mkStackBuilder(root ast.Node) stackBuilder {
	type node struct {
		node  ast.Node
		field string
	}
	parents := map[ast.Node]node{}
	postOrder(root, func(c *astutil.Cursor) bool {
		parents[c.Node()] = node{c.Parent(), c.Name()}
		return true
	})

	return func(node ast.Node) (stack []ast.Node, names []string) {
		for node != nil {
			stack = append(stack, node)
			parent := parents[node]
			node = parent.node
			names = append(names, parent.field)
		}
		return stack, names
	}
}

// X is pattern, Y is node.
// If pattern x is nil, it is equivalent to wildcard, and true is returned.
// When y is nil, first call matchFunc, because nil-case may need
// Finally, pattern is not nil, but y is nil, return false
func (m *Matcher) match(x, y ast.Node, ctx *MatchCtx) bool {
	isWildcard := IsNilNode(x)
	if isWildcard {
		return true
	}

	if matchFun := m.tryGetNodeMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}

	if x, ok := x.(ast.Stmt); ok {
		if y, ok := y.(ast.Stmt); ok {
			return m.matchStmt(x, y, ctx)
		}
	}
	if x, ok := x.(ast.Expr); ok {
		if y, ok := y.(ast.Expr); ok {
			return m.matchExpr(x, y, ctx)
		}
	}
	if x, ok := x.(ast.Spec); ok {
		if y, ok := y.(ast.Spec); ok {
			return m.matchSpec(x, y, ctx)
		}
	}
	if x, ok := x.(ast.Decl); ok {
		if y, ok := y.(ast.Decl); ok {
			return m.matchDecl(x, y, ctx)
		}
	}

	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}
	switch x := x.(type) {

	default:
		panic("unexpect Node: " + ctx.ShowNode(x))

	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Fields ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

	case *ast.Field:
		y := y.(*ast.Field)
		if matchFun := m.tryGetFieldMatchFun(x); matchFun != nil {
			return matchFun(y, ctx)
		}
		if y == nil {
			return false
		}
		return m.matchIdents(x.Names, y.Names, ctx) &&
			m.matchExpr(x.Type, y.Type, ctx) &&
			m.matchExpr(x.Tag, y.Tag, ctx)

	case *ast.FieldList:
		y := y.(*ast.FieldList)
		if matchFun := m.tryGetFieldListMatchFun(x); matchFun != nil {
			return matchFun(y, ctx)
		}
		if matchFun := m.tryGetFieldsMatchFun(x.List); matchFun != nil {
			if y == nil {
				return matchFun(nil, ctx)
			} else {
				return matchFun(FieldsNode(y.List), ctx)
			}
		}
		if y == nil {
			return false
		}
		if len(x.List) != len(y.List) {
			return false
		}
		for i := range x.List {
			if !m.match(x.List[i], y.List[i], ctx) {
				return false
			}
		}
		return true

	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Comments ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
	case *ast.Comment:
		return true

	case *ast.CommentGroup:
		return true

	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Files and packages ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
	case *ast.File:
		return true

	case *ast.Package:
		return true
	}

	return false
}

func (m *Matcher) matchSpec(x, y ast.Spec, ctx *MatchCtx) bool {
	isWildcard := IsNilNode(x)
	if isWildcard {
		return true
	}

	if matchFun := m.tryGetSpecMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}

	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}

	switch x := x.(type) {

	default:
		panic("unexpect Spec: " + ctx.ShowNode(x))

	case *ast.ImportSpec:
		y := y.(*ast.ImportSpec)
		if y == nil {
			return false
		}
		if !m.matchIdent(x.Name, y.Name, ctx) {
			return false
		}
		// import path must be string literal
		// xp, _ := strconv.Unquote(x.Path.Value)
		// yp, _ := strconv.Unquote(y.Path.Value)
		// return xp == yp
		return m.matchBasicLit(x.Path, y.Path, ctx)

	case *ast.ValueSpec:
		y := y.(*ast.ValueSpec)
		if y == nil {
			return false
		}
		return m.matchIdents(x.Names, y.Names, ctx) &&
			m.matchExpr(x.Type, y.Type, ctx) &&
			m.matchExprs(x.Values, y.Values, ctx)

	case *ast.TypeSpec:
		y := y.(*ast.TypeSpec)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Name, y.Name, ctx) &&
			m.match(x.TypeParams, y.TypeParams, ctx) &&
			m.matchExpr(x.Type, y.Type, ctx)
	}
}

func (m *Matcher) matchDecl(x, y ast.Decl, ctx *MatchCtx) bool {
	isWildcard := IsNilNode(x)
	if isWildcard {
		return true
	}

	if matchFun := m.tryGetDeclMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}

	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}

	switch x := x.(type) {
	default:
		panic("unexpect Decl: " + ctx.ShowNode(x))

	case *ast.BadDecl:
		panic("unexpect BadDecl: " + ctx.ShowNode(x))

	case *ast.GenDecl:
		y := y.(*ast.GenDecl)
		if y == nil {
			return false
		}
		return m.matchToken(x.Tok, y.Tok, ctx) &&
			m.matchSpecs(x.Specs, y.Specs, ctx)

	case *ast.FuncDecl:
		y := y.(*ast.FuncDecl)
		if y == nil {
			return false
		}
		return m.match(x.Recv, y.Recv, ctx) &&
			m.matchExpr(x.Name, y.Name, ctx) &&
			m.matchExpr(x.Type, y.Type, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx)
	}
}

func (m *Matcher) matchStmt(x, y ast.Stmt, ctx *MatchCtx) bool {
	isWildcard := IsNilNode(x)
	if isWildcard {
		return true
	}

	if matchFun := m.tryGetStmtMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}

	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}

	switch x := x.(type) {

	default:
		panic("unexpect Stmt: " + ctx.ShowNode(x))

	case *ast.BadStmt:
		panic("unexpect BadStmt: " + ctx.ShowNode(x))

	case *ast.EmptyStmt:
		// no need, checked in reflect.TypeOf
		// y := y.(*ast.EmptyStmt)
		return true

	case *ast.DeclStmt:
		y := y.(*ast.DeclStmt)
		if y == nil {
			return false
		}
		return m.matchDecl(x.Decl, y.Decl, ctx)

	case *ast.LabeledStmt:
		y := y.(*ast.LabeledStmt)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Label, y.Label, ctx) &&
			m.matchStmt(x.Stmt, y.Stmt, ctx)

	case *ast.ExprStmt:
		y := y.(*ast.ExprStmt)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx)

	case *ast.SendStmt:
		y := y.(*ast.SendStmt)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Chan, y.Chan, ctx) &&
			m.matchExpr(x.Value, y.Value, ctx)

	case *ast.IncDecStmt:
		y := y.(*ast.IncDecStmt)
		if y == nil {
			return false
		}
		return m.matchToken(x.Tok, y.Tok, ctx) &&
			m.matchExpr(x.X, y.X, ctx)

	case *ast.AssignStmt:
		y := y.(*ast.AssignStmt)
		if y == nil {
			return false
		}
		return m.matchToken(x.Tok, y.Tok, ctx) &&
			m.matchExprs(x.Lhs, y.Lhs, ctx) &&
			m.matchExprs(x.Rhs, y.Rhs, ctx)

	case *ast.GoStmt:
		y := y.(*ast.GoStmt)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Call, y.Call, ctx)

	case *ast.DeferStmt:
		y := y.(*ast.DeferStmt)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Call, y.Call, ctx)

	case *ast.ReturnStmt:
		y := y.(*ast.ReturnStmt)
		if y == nil {
			return false
		}
		return m.matchExprs(x.Results, y.Results, ctx)

	case *ast.BranchStmt:
		y := y.(*ast.BranchStmt)
		if y == nil {
			return false
		}
		return m.matchToken(x.Tok, y.Tok, ctx) &&
			m.matchExpr(x.Label, y.Label, ctx)

	case *ast.BlockStmt:
		y := y.(*ast.BlockStmt)
		if matchFun := m.tryGetBlockStmtMatchFun(x); matchFun != nil {
			return matchFun(y, ctx)
		}
		if y == nil {
			return false
		}
		return m.matchStmts(x.List, y.List, ctx)

	case *ast.IfStmt:
		y := y.(*ast.IfStmt)
		if y == nil {
			return false
		}
		return m.matchStmt(x.Init, y.Init, ctx) &&
			m.matchExpr(x.Cond, y.Cond, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx) &&
			m.matchStmt(x.Else, y.Else, ctx)

	case *ast.CaseClause:
		y := y.(*ast.CaseClause)
		if y == nil {
			return false
		}
		return m.matchExprs(x.List, y.List, ctx) &&
			m.matchStmts(x.Body, y.Body, ctx)

	case *ast.SwitchStmt:
		y := y.(*ast.SwitchStmt)
		if y == nil {
			return false
		}
		return m.matchStmt(x.Init, y.Init, ctx) &&
			m.matchExpr(x.Tag, y.Tag, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx)

	case *ast.TypeSwitchStmt:
		y := y.(*ast.TypeSwitchStmt)
		if y == nil {
			return false
		}
		return m.matchStmt(x.Init, y.Init, ctx) &&
			m.matchStmt(x.Assign, y.Assign, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx)

	case *ast.CommClause:
		y := y.(*ast.CommClause)
		if y == nil {
			return false
		}
		return m.matchStmt(x.Comm, y.Comm, ctx) &&
			m.matchStmts(x.Body, y.Body, ctx)

	case *ast.SelectStmt:
		y := y.(*ast.SelectStmt)
		if y == nil {
			return false
		}
		return m.matchStmt(x.Body, y.Body, ctx)

	case *ast.ForStmt:
		y := y.(*ast.ForStmt)
		if y == nil {
			return false
		}
		return m.matchStmt(x.Init, y.Init, ctx) &&
			m.matchExpr(x.Cond, y.Cond, ctx) &&
			m.matchStmt(x.Post, y.Post, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx)

	case *ast.RangeStmt:
		y := y.(*ast.RangeStmt)
		if y == nil {
			return false
		}
		return m.matchToken(x.Tok, y.Tok, ctx) &&
			m.matchExpr(x.Key, y.Key, ctx) &&
			m.matchExpr(x.Value, y.Value, ctx) &&
			m.matchExpr(x.X, y.X, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx)
		return true
	}
}

func (m *Matcher) matchExpr(x, y ast.Expr, ctx *MatchCtx) bool {
	isWildcard := IsNilNode(x)
	if isWildcard {
		return true
	}

	if m.UnparenExpr {
		x = astutil.Unparen(x)
		y = astutil.Unparen(y)
	}

	if matchFun := m.tryGetExprMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}

	if reflect.TypeOf(x) != reflect.TypeOf(y) {
		return false
	}

	switch x := x.(type) {

	default:
		panic("unexpect Expr: " + ctx.ShowNode(x))

	case *ast.BadExpr:
		panic("unexpect BadExpr: " + ctx.ShowNode(x))

	case *ast.Ident:
		y := y.(*ast.Ident)
		if matchFun := m.tryGetIdentMatchFun(x); matchFun != nil {
			return matchFun(y, ctx)
		}
		if y == nil {
			return false
		}
		return x.Name == y.Name

	case *ast.Ellipsis:
		y := y.(*ast.Ellipsis)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Elt, y.Elt, ctx)

	case *ast.BasicLit:
		y := y.(*ast.BasicLit)
		return m.matchBasicLit(x, y, ctx)
	case *ast.FuncLit:
		y := y.(*ast.FuncLit)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Type, y.Type, ctx) &&
			m.matchStmt(x.Body, y.Body, ctx)

	case *ast.CompositeLit:
		y := y.(*ast.CompositeLit)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Type, y.Type, ctx) &&
			m.matchExprs(x.Elts, y.Elts, ctx)

	case *ast.ParenExpr:
		y := y.(*ast.ParenExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx)

	case *ast.SelectorExpr:
		y := y.(*ast.SelectorExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx) &&
			m.matchExpr(x.Sel, y.Sel, ctx)

	case *ast.IndexExpr:
		y := y.(*ast.IndexExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx) &&
			m.matchExpr(x.Index, y.Index, ctx)

	case *ast.IndexListExpr:
		y := y.(*ast.IndexListExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx) &&
			m.matchExprs(x.Indices, y.Indices, ctx)

	case *ast.SliceExpr:
		y := y.(*ast.SliceExpr)
		if y == nil {
			return false
		}
		return x.Slice3 == y.Slice3 &&
			m.matchExpr(x.X, y.X, ctx) &&
			m.matchExpr(x.Low, y.Low, ctx) &&
			m.matchExpr(x.High, y.High, ctx) &&
			m.matchExpr(x.Max, y.Max, ctx)

	case *ast.TypeAssertExpr:
		y := y.(*ast.TypeAssertExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx) &&
			m.matchExpr(x.Type, y.Type, ctx)

	case *ast.CallExpr:
		y := y.(*ast.CallExpr)
		if matchFun := m.tryGetCallExprMatchFun(x); matchFun != nil {
			return matchFun(y, ctx)
		}
		if y == nil {
			return false
		}
		if m.MatchCallEllipsis &&
			x.Ellipsis.IsValid() != y.Ellipsis.IsValid() {
			return false
		}
		return m.matchExpr(x.Fun, y.Fun, ctx) &&
			m.matchExprs(x.Args, y.Args, ctx)

	case *ast.StarExpr:
		y := y.(*ast.StarExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.X, y.X, ctx)

	case *ast.UnaryExpr:
		y := y.(*ast.UnaryExpr)
		if y == nil {
			return false
		}
		return m.matchToken(x.Op, y.Op, ctx) &&
			m.matchExpr(x.X, y.X, ctx)

	case *ast.BinaryExpr:
		y := y.(*ast.BinaryExpr)
		if y == nil {
			return false
		}
		return m.matchToken(x.Op, y.Op, ctx) &&
			m.matchExpr(x.X, y.X, ctx) &&
			m.matchExpr(x.Y, y.Y, ctx)

	case *ast.KeyValueExpr:
		y := y.(*ast.KeyValueExpr)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Key, y.Key, ctx) &&
			m.matchExpr(x.Value, y.Value, ctx)

	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Types Exprs ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

	case *ast.ArrayType:
		y := y.(*ast.ArrayType)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Len, y.Len, ctx) &&
			m.matchExpr(x.Elt, y.Elt, ctx)

	case *ast.StructType:
		y := y.(*ast.StructType)
		if y == nil {
			return false
		}
		return m.match(x.Fields, y.Fields, ctx)

	case *ast.FuncType:
		y := y.(*ast.FuncType)
		if matchFun := m.tryGetFuncTypeMatchFun(x); matchFun != nil {
			return matchFun(y, ctx)
		}
		if y == nil {
			return false
		}
		return m.match(x.TypeParams, y.TypeParams, ctx) &&
			m.match(x.Params, y.Params, ctx) &&
			m.match(x.Results, y.Results, ctx)

	case *ast.InterfaceType:
		y := y.(*ast.InterfaceType)
		if y == nil {
			return false
		}
		return m.match(x.Methods, y.Methods, ctx)

	case *ast.MapType:
		y := y.(*ast.MapType)
		if y == nil {
			return false
		}
		return m.matchExpr(x.Key, y.Key, ctx) &&
			m.matchExpr(x.Value, y.Value, ctx)

	case *ast.ChanType:
		y := y.(*ast.ChanType)
		if y == nil {
			return false
		}
		return x.Dir == y.Dir &&
			m.matchExpr(x.Value, y.Value, ctx)
	}
}

func (m *Matcher) matchIdent(x, y *ast.Ident, ctx *MatchCtx) bool {
	isWildcard := x == nil
	if isWildcard {
		return true
	}
	if matchFun := m.tryGetIdentMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}
	if y == nil {
		return false
	}
	return x.Name == y.Name
}

func (m *Matcher) matchBasicLit(x, y *ast.BasicLit, ctx *MatchCtx) bool {
	isWildcard := x == nil
	if isWildcard {
		return true
	}
	if matchFun := m.tryGetBasicLitMatchFun(x); matchFun != nil {
		return matchFun(y, ctx)
	}
	if y == nil {
		return false
	}
	// Notice: BasicLit is an atomic Pattern,
	// &ast.BasicLit{ Kind: token.INT } can be used for matching INT literal
	// because zero Value is ambiguous, wildcard or zero value?
	xVal := constant.MakeFromLiteral(x.Value, x.Kind, 0)
	yVal := constant.MakeFromLiteral(y.Value, y.Kind, 0)
	return constant.Compare(xVal, token.EQL, yVal)
}

func (m *Matcher) matchToken(x, y token.Token, ctx *MatchCtx) bool {
	// ast.RangeStmt.Tok is ILLEGAL if Key == nil
	// so, token.ILLEGAL can't be wildcard
	// isWildcard := x == token.ILLEGAL
	// if isWildcard { return true }
	if matchFun := m.tryGetTokenMatchFun(x); matchFun != nil {
		return matchFun(TokenNode(y), ctx)
	}
	return x == y
}

func (m *Matcher) matchStmts(xs, ys []ast.Stmt, ctx *MatchCtx) bool {
	isWildcard := xs == nil
	if isWildcard {
		return true
	}
	if matchFun := m.tryGetStmtsMatchFun(xs); matchFun != nil {
		return matchFun(StmtsNode(ys), ctx)
	}

	if len(xs) > 0 {
		if len(xs)-1 > len(ys) {
			return false
		}
		if matchFun := m.tryGetRestStmtMatchFun(xs[len(xs)-1]); matchFun != nil {
			// last with the rest pattern
			i := 0
			for ; i < len(xs)-1; i++ {
				if !m.matchStmt(xs[i], ys[i], ctx) {
					return false
				}
			}
			return matchFun(StmtsNode(ys[i:]), ctx)
		}
	}

	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if !m.matchStmt(xs[i], ys[i], ctx) {
			return false
		}
	}
	return true
}

func (m *Matcher) matchExprs(xs, ys []ast.Expr, ctx *MatchCtx) bool {
	// notice: nil is wildcard-pattern, but []ast.Expr{} exactly Matched empty ys
	isWildcard := xs == nil
	if isWildcard {
		return true
	}
	if matchFun := m.tryGetExprsMatchFun(xs); matchFun != nil {
		return matchFun(ExprsNode(ys), ctx)
	}

	if len(xs) > 0 {
		if len(xs)-1 > len(ys) {
			return false
		}
		if matchFun := m.tryGetRestExprMatchFun(xs[len(xs)-1]); matchFun != nil {
			// last with the rest pattern
			i := 0
			for ; i < len(xs)-1; i++ {
				if !m.matchExpr(xs[i], ys[i], ctx) {
					return false
				}
			}
			return matchFun(ExprsNode(ys[i:]), ctx)
		}
	}

	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if !m.matchExpr(xs[i], ys[i], ctx) {
			return false
		}
	}
	return true
}

func (m *Matcher) matchIdents(xs, ys []*ast.Ident, ctx *MatchCtx) bool {
	isWildcard := xs == nil
	if isWildcard {
		return true
	}
	if matchFun := m.tryGetIdentsMatchFun(xs); matchFun != nil {
		return matchFun(IdentsNode(ys), ctx)
	}
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if !m.matchIdent(xs[i], ys[i], ctx) {
			return false
		}
	}
	return true
}

func (m *Matcher) matchSpecs(xs, ys []ast.Spec, ctx *MatchCtx) bool {
	isWildcard := xs == nil
	if isWildcard {
		return true
	}
	if matchFun := m.tryGetSpecsMatchFun(xs); matchFun != nil {
		return matchFun(SpecsNode(ys), ctx)
	}
	if len(xs) != len(ys) {
		return false
	}
	for i := range xs {
		if !m.matchSpec(xs[i], ys[i], ctx) {
			return false
		}
	}
	return true
}
