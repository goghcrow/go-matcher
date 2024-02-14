package matcher

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

// Used For MatchFun Callback Param
// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ Pseudo Node ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓

type PseudoNode interface {
	FunNode | StmtsNode | ExprsNode | SpecsNode | IdentsNode | FieldsNode | TokenNode
}

type (
	FunNode    = MatchFun
	StmtsNode  []ast.Stmt   // for the callback param of StmtsPattern
	ExprsNode  []ast.Expr   // for the callback param of ExprsPattern
	SpecsNode  []ast.Spec   // for the callback param of SpecsPattern
	IdentsNode []*ast.Ident // for the callback param of IdentsPattern
	FieldsNode []*ast.Field // for the callback param of FieldsPattern
	TokenNode  token.Token  // for the callback param of TokenPattern
)

func (FunNode) Pos() token.Pos    { return token.NoPos }
func (FunNode) End() token.Pos    { return token.NoPos }
func (StmtsNode) Pos() token.Pos  { return token.NoPos }
func (StmtsNode) End() token.Pos  { return token.NoPos }
func (ExprsNode) Pos() token.Pos  { return token.NoPos }
func (ExprsNode) End() token.Pos  { return token.NoPos }
func (SpecsNode) Pos() token.Pos  { return token.NoPos }
func (SpecsNode) End() token.Pos  { return token.NoPos }
func (IdentsNode) Pos() token.Pos { return token.NoPos }
func (IdentsNode) End() token.Pos { return token.NoPos }
func (FieldsNode) Pos() token.Pos { return token.NoPos }
func (FieldsNode) End() token.Pos { return token.NoPos }
func (TokenNode) Pos() token.Pos  { return token.NoPos }
func (TokenNode) End() token.Pos  { return token.NoPos }

func IsPseudoNode(n ast.Node) bool {
	switch n.(type) {
	case FunNode:
		return true
	case StmtsNode:
		return true
	case ExprsNode:
		return true
	case SpecsNode:
		return true
	case IdentsNode:
		return true
	case FieldsNode:
		return true
	case TokenNode:
		return true
	}
	return false
}

func showPseudoNode(fset *token.FileSet, n ast.Node) string {
	switch n := n.(type) {
	case FunNode:
		return "match-fun"
	case StmtsNode:
		xs := make([]string, len(n))
		for i, it := range n {
			xs[i] = ShowNode(fset, it)
		}
		return strings.Join(xs, "\n")
	case ExprsNode:
		xs := make([]string, len(n))
		for i, it := range n {
			xs[i] = ShowNode(fset, it)
		}
		return strings.Join(xs, "\n")
	case SpecsNode:
		xs := make([]string, len(n))
		for i, it := range n {
			xs[i] = ShowNode(fset, it)
		}
		return strings.Join(xs, "\n")
	case IdentsNode:
		xs := make([]string, len(n))
		for i, it := range n {
			xs[i] = ShowNode(fset, it)
		}
		return strings.Join(xs, "\n")
	case FieldsNode:
		xs := make([]string, len(n))
		for i, it := range n {
			xs[i] = ShowNode(fset, it)
		}
		return strings.Join(xs, "\n")
	case TokenNode:
		return token.Token(n).String()
	default:
		panic("unknown pseudo node")
	}
}

func ShowNode(fset *token.FileSet, n ast.Node) string {
	if IsPseudoNode(n) {
		return showPseudoNode(fset, n)
	}
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, n)
	return buf.String()
}
