package matcher

import (
	"go/ast"
	"go/token"
)

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Pattern ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
// Public API

// Notice myself: How to add new Pattern
// 0. Declare XXXPattern type, add to Pattern interface
// 1. Add MkXXXPattern && MkPattern
// 2. Add TryGetXXXMatchFun & TryGetMatchFun
// 3. Add MkXXXPatternVar & MkVar
// 4. Hook Matcher, call TryGetXXXMatchFun

// Notice: BasicLit is an atomic Pattern, doesn't support expanding match
// e.g., matching BasicLit.Kind
// And to judge equivalence of literal by constant.Compare, not just Kind

type (
	Pattern interface {
		NodePattern |
			StmtPattern | RestStmtPattern |
			ExprPattern | RestExprPattern |
			DeclPattern | SpecPattern |
			IdentPattern | FieldPattern | FieldListPattern |
			CallExprPattern | FuncTypePattern | BlockStmtPattern | TokenPattern | BasicLitPattern |
			SlicePattern
	}
	TypingPattern interface {
		IdentPattern | ExprPattern
	}
	SlicePattern interface {
		StmtsPattern | ExprsPattern | SpecsPattern | IdentsPattern | FieldsPattern
	}
	ElemPattern interface {
		StmtPattern | ExprPattern | SpecPattern | IdentPattern | FieldPattern
	}

	NodePattern      = MatchFun
	StmtPattern      = *ast.BadStmt
	RestStmtPattern  = *ast.EmptyStmt
	ExprPattern      = *ast.BadExpr
	RestExprPattern  = *ast.Ellipsis
	DeclPattern      = *ast.BadDecl
	SpecPattern      = *ast.ImportSpec
	IdentPattern     = *ast.Ident
	FieldPattern     = *ast.Field
	FieldListPattern = *ast.FieldList
	CallExprPattern  = *ast.CallExpr
	FuncTypePattern  = *ast.FuncType
	BlockStmtPattern = *ast.BlockStmt
	BasicLitPattern  = *ast.BasicLit // for matching Field.Tag, Import.Path
	TokenPattern     = token.Token   // for matching token type
	StmtsPattern     = []ast.Stmt
	ExprsPattern     = []ast.Expr
	SpecsPattern     = []ast.Spec
	IdentsPattern    = []*ast.Ident // for matching Field.Name, etc.
	FieldsPattern    = []*ast.Field
	// ChanDirPattern = ast.ChanDir // no need, just two values, use Or to match
	// StringPattern = string // maybe for expanding match Ident, Import.Path
)

func IsPattern[T Pattern](m *Matcher, n any) bool {
	return TryGetMatchFun[T](m, n) != nil
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Factory ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// MkVar make variable for binding matched Node
func MkVar[T Pattern](m *Matcher, name string) T {
	return MkPattern[T](m, func(n ast.Node, ctx *MatchCtx) bool {
		ctx.Binds[name] = n
		return true
	})
}

// PatternOf make pattern from ast.Node
func PatternOf[T Pattern](m *Matcher, ptn ast.Node) T {
	assert(!IsPseudoNode(ptn), "invalid pattern")
	return MkPattern[T](m, func(n ast.Node, ctx *MatchCtx) bool {
		return ctx.match(ptn, n)
	})
}

// MkPattern make pattern from MatchFun
func MkPattern[T Pattern](m *Matcher, f MatchFun) T {
	var zero T
	switch any(zero).(type) {
	case NodePattern:
		return any(m.mkNodePattern(f)).(T)
	case StmtPattern:
		return any(m.mkStmtPattern(f)).(T)
	case RestStmtPattern:
		return any(m.mkRestStmtPattern(f)).(T)
	case ExprPattern:
		return any(m.mkExprPattern(f)).(T)
	case RestExprPattern:
		return any(m.mkRestExprPattern(f)).(T)
	case DeclPattern:
		return any(m.mkDeclPattern(f)).(T)
	case SpecPattern:
		return any(m.mkSpecPattern(f)).(T)
	case IdentPattern:
		return any(m.mkIdentPattern(f)).(T)
	case FieldPattern:
		return any(m.mkFieldPattern(f)).(T)
	case FieldListPattern:
		return any(m.mkFieldListPattern(f)).(T)
	case CallExprPattern:
		return any(m.mkCallExprPattern(f)).(T)
	case FuncTypePattern:
		return any(m.mkFuncTypePattern(f)).(T)
	case BlockStmtPattern:
		return any(m.mkBlockStmtPattern(f)).(T)
	case TokenPattern:
		return any(m.mkTokenPattern(f)).(T)
	case BasicLitPattern:
		return any(m.mkBasicLitPattern(f)).(T)
	case StmtsPattern:
		return any(m.mkStmtsPattern(f)).(T)
	case ExprsPattern:
		return any(m.mkExprsPattern(f)).(T)
	case SpecsPattern:
		return any(m.mkSpecsPattern(f)).(T)
	case IdentsPattern:
		return any(m.mkIdentsPattern(f)).(T)
	case FieldsPattern:
		return any(m.mkFieldsPattern(f)).(T)
	default:
		panic("unreachable")
	}
}

// TryGetMatchFun get MatchFun from pattern
// if T is NodePattern, n must be ast.Node
// if T is StmtPattern, n must be ast.Stmt
// if T is ExprPattern, n must be ast.Expr
// if T is DeclPattern, n must be ast.Decl
// if T is SpecPattern, n must be ast.Spec
// if T is IdentPattern, n must be *ast.Ident
// if T is FieldPattern, n must be *ast.Field
// if T is FieldListPattern, n must be *ast.FieldList
// if T is CallExprPattern, n must be *ast.CallExpr
// if T is FuncTypePattern, n must be *ast.FuncType
// if T is BlockStmtPattern, n must be *ast.BlockStmt
// if T is TokenPattern, n must be token.Token
// if T is BasicLitPattern, n must be *ast.BasicLit
// if T is StmtsPattern, n must be []ast.Stmt
// if T is ExprsPattern, n must be []ast.Expr
// if T is SpecsPattern, n must be []ast.Spec
// if T is IdentsPattern, n must be []*ast.Ident
// if T is FieldsPattern, n must be []*ast.Field
func TryGetMatchFun[T Pattern](m *Matcher, n any) MatchFun {
	var zero T
	switch any(zero).(type) {
	case NodePattern:
		return m.tryGetNodeMatchFun(n.(ast.Node))
	case StmtPattern:
		return m.tryGetStmtMatchFun(n.(ast.Stmt))
	case RestStmtPattern:
		return m.tryGetRestStmtMatchFun(n.(ast.Stmt))
	case ExprPattern:
		return m.tryGetExprMatchFun(n.(ast.Expr))
	case RestExprPattern:
		return m.tryGetRestExprMatchFun(n.(ast.Expr))
	case DeclPattern:
		return m.tryGetDeclMatchFun(n.(ast.Decl))
	case SpecPattern:
		return m.tryGetSpecMatchFun(n.(ast.Spec))
	case IdentPattern:
		return m.tryGetIdentMatchFun(n.(*ast.Ident))
	case FieldPattern:
		return m.tryGetFieldMatchFun(n.(*ast.Field))
	case FieldListPattern:
		return m.tryGetFieldListMatchFun(n.(*ast.FieldList))
	case CallExprPattern:
		return m.tryGetCallExprMatchFun(n.(*ast.CallExpr))
	case FuncTypePattern:
		return m.tryGetFuncTypeMatchFun(n.(*ast.FuncType))
	case BlockStmtPattern:
		return m.tryGetBlockStmtMatchFun(n.(*ast.BlockStmt))
	case TokenPattern:
		return m.tryGetTokenMatchFun(n.(token.Token))
	case BasicLitPattern:
		return m.tryGetBasicLitMatchFun(n.(*ast.BasicLit))
	case StmtsPattern:
		return m.tryGetStmtsMatchFun(n.([]ast.Stmt))
	case ExprsPattern:
		return m.tryGetExprsMatchFun(n.([]ast.Expr))
	case SpecsPattern:
		return m.tryGetSpecsMatchFun(n.([]ast.Spec))
	case IdentsPattern:
		return m.tryGetIdentsMatchFun(n.([]*ast.Ident))
	case FieldsPattern:
		return m.tryGetFieldsMatchFun(n.([]*ast.Field))
	default:
		// panic("unreachable")
		return nil
	}
}

func MustGetMatchFun[T Pattern](m *Matcher, n any) MatchFun {
	fun := TryGetMatchFun[T](m, n)
	assert(fun != nil, "invalid pattern")
	return fun
}

func TryGetOrMkMatchFun[T Pattern](m *Matcher, nodeOrPtn ast.Node) MatchFun {
	fun := TryGetMatchFun[T](m, nodeOrPtn)
	if fun != nil {
		return fun
	}
	return func(n ast.Node, ctx *MatchCtx) bool {
		return ctx.match(nodeOrPtn, n)
	}
}
