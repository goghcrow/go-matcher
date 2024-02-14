package example

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfWildcardIdent(m *Matcher) ast.Node {
	// bind matched wildcard ident to "var"
	return Bind(m,
		"var",
		Wildcard[IdentPattern](m),
	)
}

func PatternOfLitVal(m *Matcher) ast.Node {
	return &ast.CompositeLit{
		// Type: nil,
		// nil value is wildcard, so Nil Pattern is needed to represent exactly nil type expr
		Type: Nil[ExprPattern](m),
	}
}

func PatternOfVarDecl(m *Matcher) ast.Node {
	return &ast.GenDecl{
		Tok: token.VAR, // IMPORT, CONST, TYPE, or VAR
		// pattern variable, match any and bind to "var"
		Specs: matcher.MkVar[SpecsPattern](m, "var"),
	}
}

func PatternOfConstDecl(m *Matcher) ast.Node {
	return &ast.GenDecl{
		Tok: token.CONST, // IMPORT, CONST, TYPE, or VAR
		// pattern variable, match any and bind to "var"
		Specs: matcher.MkVar[SpecsPattern](m, "var"),
	}
}

func PatternOfValSpec(m *Matcher) ast.Node {
	return &ast.ValueSpec{
		// pattern variable, match any and bind to "var"
		Names: matcher.MkVar[IdentsPattern](m, "var"),
		Type:  TypeIdentical[ExprPattern](m, types.Typ[types.Int]),
	}
}
