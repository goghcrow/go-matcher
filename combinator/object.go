package combinator

import "go/types"

// ObjectOf
// notice: can't be used for `f` or `a.b` in `f[T]()` `a.b[T]()`
// unpacking index/indexList is needed firstly
// please use XXX CalleeOf
func ObjectOf(m *Matcher, p Predicate[types.Object]) ExprPattern {
	return OrEx[ExprPattern](m,
		IdentObjectOf(m, p),
		SelectorObjectOf(m, p),
	)
}
