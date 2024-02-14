package example

import (
	"go/ast"
	"reflect"

	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfStructFieldWithJsonTag(m *Matcher) ast.Node {
	return &ast.Field{
		Tag: Bind(m,
			"var",
			TagOf(m, func(_ *MatchCtx, tag *reflect.StructTag) bool {
				if tag == nil {
					return false
				}
				_, ok := tag.Lookup("json")
				return ok
			}),
		),
	}
}
