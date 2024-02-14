package example

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfIdCompare(m *Matcher) ast.Node {
	isIdIdent := IdentOf(m, func(_ *MatchCtx, id *ast.Ident) bool {
		// return strings.ToLower(id.Name) == "id"
		return strings.HasSuffix(strings.ToLower(id.Name), "id")
	})
	isIdSel := &ast.SelectorExpr{Sel: isIdIdent}
	isIdPtn := OrEx[ExprPattern](m,
		isIdSel,
		isIdIdent,
	)
	isCmpTokPtn := matcher.MkPattern[TokenPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		tok := token.Token(n.(TokenNode))
		return tok == token.LSS || tok == token.GTR || tok == token.LEQ || tok == token.GEQ
	})
	notIntLit := Not(m, matcher.MkPattern[ExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		basic, _ := n.(*ast.BasicLit)
		if basic == nil {
			return false
		}
		return basic.Kind == token.INT
	}))

	// return &ast.BinaryExpr{
	// 	X:  isIdPtn,
	// 	Op: isCmpTokPtn,
	// 	Y:  isIdPtn,
	// }

	return OrEx[ExprPattern](m,
		&ast.BinaryExpr{
			X:  isIdPtn,
			Op: isCmpTokPtn,
			Y:  notIntLit,
		},
		&ast.BinaryExpr{
			X:  notIntLit,
			Op: isCmpTokPtn,
			Y:  isIdPtn,
		},
	)
}
