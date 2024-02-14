package matcher

import (
	"go/ast"
	"go/token"
)

// MatchFun Container
// Index ref MkXXXPattern
// index: (BadExpr|BadStmt|BadDecl).FromPos
// ImportSpec.EndPos
// Ident.NamePos
// Field.Doc.List[0].Slash
// BasicLit.ValuePos
// FieldList.Opening
// CallExpr.Lparen
// FuncType.Func
// BlockStmt.Lbrace
type matchFuns []MatchFun

func (p *matchFuns) append(f MatchFun) token.Pos {
	*p = append(*p, f)
	return token.Pos(-len(*p))
}

func (p *matchFuns) get(pos token.Pos) MatchFun {
	return (*p)[-pos-1]
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ mkXXXPattern ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

// MkNodePattern type of callback param node is ast.Node
func (p *matchFuns) mkNodePattern(f MatchFun) NodePattern {
	return f
}

// MkStmtPattern type of callback param node is ast.Stmt
func (p *matchFuns) mkStmtPattern(f MatchFun) StmtPattern {
	return &ast.BadStmt{From: p.append(f)}
}

// MkRestStmtPattern type of callback param node is StmtsNode
func (p *matchFuns) mkRestStmtPattern(f MatchFun) RestStmtPattern {
	return &ast.EmptyStmt{Semicolon: p.append(f)}
}

// MkExprPattern type of callback param node is ast.Expr
func (p *matchFuns) mkExprPattern(f MatchFun) ExprPattern {
	return &ast.BadExpr{From: p.append(f)}
}

// MkRestExprPattern type of callback param node is ExprsNode
func (p *matchFuns) mkRestExprPattern(f MatchFun) RestExprPattern {
	return &ast.Ellipsis{Ellipsis: p.append(f)}
}

// MkDeclPattern type of callback param node is ast.Decl
func (p *matchFuns) mkDeclPattern(f MatchFun) DeclPattern {
	return &ast.BadDecl{From: p.append(f)}
}

// MkSpecPattern type of callback param node is ast.Spec
func (p *matchFuns) mkSpecPattern(f MatchFun) SpecPattern {
	return &ast.ImportSpec{EndPos: p.append(f)}
}

// MkIdentPattern type of callback param node is *ast.Ident
func (p *matchFuns) mkIdentPattern(f MatchFun) IdentPattern {
	return &ast.Ident{NamePos: p.append(f)}
}

// MkFieldPattern type of callback param node is *ast.Field
func (p *matchFuns) mkFieldPattern(f MatchFun) FieldPattern {
	// Putting pos it in Type/Tag/Name will cause ambiguity
	// e.g. Field{ Type: MkExprPattern() }
	// e.g. Field{ Tag: MkBasicLitPattern() }
	// e.g. Field{ Name: MkIdentsPattern() }
	return &ast.Field{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{Slash: p.append(f)},
				nil,
			},
		},
	}
}

// MkFieldListPattern type of callback param node is *ast.FieldList
func (p *matchFuns) mkFieldListPattern(f MatchFun) FieldListPattern {
	return &ast.FieldList{Opening: p.append(f)}
}

// MkCallExprPattern type of callback param node is *ast.CallExpr
func (p *matchFuns) mkCallExprPattern(f MatchFun) CallExprPattern {
	return &ast.CallExpr{Lparen: p.append(f)}
}

// MkFuncTypePattern type of callback param node is *ast.FuncType
func (p *matchFuns) mkFuncTypePattern(f MatchFun) FuncTypePattern {
	return &ast.FuncType{Func: p.append(f)}
}

// MkBlockStmtPattern type of callback param node is *ast.BlockStmt
func (p *matchFuns) mkBlockStmtPattern(f MatchFun) BlockStmtPattern {
	return &ast.BlockStmt{Lbrace: p.append(f)}
}

// MkTokenPattern type of callback param node is TokenNode
func (p *matchFuns) mkTokenPattern(f MatchFun) TokenPattern {
	return token.Token(p.append(f))
}

// MkBasicLitPattern type of callback param node is *ast.BasicLit
func (p *matchFuns) mkBasicLitPattern(f MatchFun) BasicLitPattern {
	return &ast.BasicLit{ValuePos: p.append(f)}
}

// []Pattern
// one more nil is for avoiding ambiguity
// e.g. []Expr{ XXXExprPattern }
// will be recognized as ExprsPattern, not ExprsPattern with only one element
// Normal grammar shouldn't be []Node {, nil }

// MkStmtsPattern type of callback param node is StmtsNode
func (p *matchFuns) mkStmtsPattern(f MatchFun) StmtsPattern {
	return []ast.Stmt{p.mkStmtPattern(f), nil}
}

// MkExprsPattern type of callback param node is ExprsNode
func (p *matchFuns) mkExprsPattern(f MatchFun) ExprsPattern {
	return []ast.Expr{p.mkExprPattern(f), nil}
}

// MkSpecsPattern type of callback param node is SpecsNode
func (p *matchFuns) mkSpecsPattern(f MatchFun) SpecsPattern {
	return []ast.Spec{p.mkSpecPattern(f), nil}
}

// MkIdentsPattern type of callback param node is IdentsNode
func (p *matchFuns) mkIdentsPattern(f MatchFun) IdentsPattern {
	return []*ast.Ident{p.mkIdentPattern(f), nil}
}

// MkFieldsPattern type of callback param node is FieldsNode
func (p *matchFuns) mkFieldsPattern(f MatchFun) FieldsPattern {
	return []*ast.Field{p.mkFieldPattern(f), nil}
}

// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ tryGetXXXMatchFun ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

func (p *matchFuns) tryGetNodeMatchFun(n ast.Node) MatchFun {
	if x, ok := n.(NodePattern); ok {
		return x
	}
	return nil
}

func (p *matchFuns) tryGetStmtMatchFun(n ast.Stmt) MatchFun {
	if x, _ := n.(StmtPattern); x != nil && x.From < 0 {
		return p.get(x.From)
	}
	return nil
}

func (p *matchFuns) tryGetRestStmtMatchFun(n ast.Stmt) MatchFun {
	if x, _ := n.(RestStmtPattern); x != nil && x.Semicolon < 0 {
		return p.get(x.Semicolon)
	}
	return nil
}

func (p *matchFuns) tryGetExprMatchFun(n ast.Expr) MatchFun {
	if x, _ := n.(ExprPattern); x != nil && x.From < 0 {
		return p.get(x.From)
	}
	return nil
}

func (p *matchFuns) tryGetRestExprMatchFun(n ast.Expr) MatchFun {
	if x, _ := n.(RestExprPattern); x != nil && x.Ellipsis < 0 {
		return p.get(x.Ellipsis)
	}
	return nil
}

func (p *matchFuns) tryGetDeclMatchFun(n ast.Decl) MatchFun {
	if x, _ := n.(DeclPattern); x != nil && x.From < 0 {
		return p.get(x.From)
	}
	return nil
}

func (p *matchFuns) tryGetSpecMatchFun(n ast.Spec) MatchFun {
	if x, _ := n.(*ast.ImportSpec); x != nil && x.EndPos < 0 {
		return p.get(x.EndPos)
	}
	return nil
}

func (p *matchFuns) tryGetIdentMatchFun(x *ast.Ident) MatchFun {
	if x != nil && x.NamePos < 0 {
		return p.get(x.NamePos)
	}
	return nil
}

func (p *matchFuns) tryGetFieldMatchFun(x *ast.Field) MatchFun {
	if x != nil && x.Doc != nil &&
		len(x.Doc.List) == 2 &&
		x.Doc.List[0].Slash < 0 &&
		x.Doc.List[1] == nil {
		return p.get(x.Doc.List[0].Slash)
	}
	return nil
}

func (p *matchFuns) tryGetFieldListMatchFun(x *ast.FieldList) MatchFun {
	if x != nil && x.Opening < 0 {
		return p.get(x.Opening)
	}
	return nil
}

func (p *matchFuns) tryGetCallExprMatchFun(x *ast.CallExpr) MatchFun {
	if x != nil && x.Lparen < 0 {
		return p.get(x.Lparen)
	}
	return nil
}

func (p *matchFuns) tryGetFuncTypeMatchFun(x *ast.FuncType) MatchFun {
	if x != nil && x.Func < 0 {
		return p.get(x.Func)
	}
	return nil
}

func (p *matchFuns) tryGetBlockStmtMatchFun(x *ast.BlockStmt) MatchFun {
	if x != nil && x.Lbrace < 0 {
		return p.get(x.Lbrace)
	}
	return nil
}

func (p *matchFuns) tryGetTokenMatchFun(x token.Token) MatchFun {
	if x < 0 {
		return p.get(token.Pos(x))
	}
	return nil
}

func (p *matchFuns) tryGetBasicLitMatchFun(x *ast.BasicLit) MatchFun {
	if x != nil && x.ValuePos < 0 {
		return p.get(x.ValuePos)
	}
	return nil
}

func (p *matchFuns) tryGetStmtsMatchFun(xs []ast.Stmt) MatchFun {
	if len(xs) != 2 || xs[1] != nil {
		return nil
	}
	return p.tryGetStmtMatchFun(xs[0])
}

func (p *matchFuns) tryGetExprsMatchFun(xs []ast.Expr) MatchFun {
	if len(xs) != 2 || xs[1] != nil {
		return nil
	}
	return p.tryGetExprMatchFun(xs[0])
}

func (p *matchFuns) tryGetSpecsMatchFun(xs []ast.Spec) MatchFun {
	if len(xs) != 2 || xs[1] != nil {
		return nil
	}
	return p.tryGetSpecMatchFun(xs[0])
}

func (p *matchFuns) tryGetIdentsMatchFun(xs []*ast.Ident) MatchFun {
	if len(xs) != 2 || xs[1] != nil {
		return nil
	}
	return p.tryGetIdentMatchFun(xs[0])
}

func (p *matchFuns) tryGetFieldsMatchFun(xs []*ast.Field) MatchFun {
	if len(xs) != 2 || xs[1] != nil {
		return nil
	}
	return p.tryGetFieldMatchFun(xs[0])
}
