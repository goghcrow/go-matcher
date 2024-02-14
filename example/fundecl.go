package example

import (
	"go/ast"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfAllFuncOrMethodDeclName(m *Matcher) ast.Node {
	return &ast.FuncDecl{
		Name: matcher.MkVar[IdentPattern](m, "var"),
	}
}

func PatternOfFuncOrMethodDeclWithSpecName(m *Matcher, name string) ast.Node {
	return &ast.FuncDecl{
		Name: IdentNameOf(m, name),
	}
}

func PatternOfFuncDeclHasAnyParam(m *Matcher, param *ast.Field) *ast.FuncDecl {
	return &ast.FuncDecl{
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: SliceContains[FieldsPattern](m, param),
			},
		},
	}
}

func PatternOfFuncDeclHasAnyParamNode(m *Matcher, param *ast.Field) *ast.FuncDecl {
	return &ast.FuncDecl{
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: SliceContains[FieldsPattern](m, param),
			},
		},
	}
}

func PatternOfMethodHasAnyParam(m *Matcher, param *ast.Field) *ast.FuncDecl {
	return &ast.FuncDecl{
		Recv: IsMethodRecv(m),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: SliceContains[FieldsPattern](m, param),
			},
		},
	}
}

func IsMethodRecv(m *Matcher) FieldListPattern {
	return Not[FieldListPattern](m, IsFuncRecv(m))
}

func IsFuncRecv(m *Matcher) FieldListPattern {
	return matcher.MkPattern[FieldListPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		return matcher.IsNilNode(n)
	})
}

func RecvOf(m *Matcher, f func(recv *ast.Field) bool) FieldListPattern {
	return matcher.MkPattern[FieldListPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		lst := n.(*ast.FieldList)
		if lst == nil {
			return false
		}
		if lst.NumFields() != 1 {
			return false
		}
		return f(lst.List[0])
	})
}
