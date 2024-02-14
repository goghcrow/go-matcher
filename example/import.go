package example

import (
	"go/ast"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfAllImportSpec(m *Matcher) ast.Node {
	return &ast.ImportSpec{
		// pattern variable, match any and bind to "var"
		Path: matcher.MkVar[BasicLitPattern](m, "var"),
	}
}
