package combinator

import (
	"go/ast"
	"reflect"

	"github.com/goghcrow/go-matcher"
)

func SliceContains[S SlicePattern](m *Matcher, p NodeOrPtn) S {
	return matcher.MkPattern[S](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		xs := reflect.ValueOf(n)
		for i := 0; i < xs.Len(); i++ {
			node := xs.Index(i).Interface().(ast.Node)
			if ctx.Matched(p, node) {
				return true
			}
		}
		return false
	})
}

func SliceLenOf[T SlicePattern](m *Matcher, p Predicate[int]) T {
	return matcher.MkPattern[T](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		return p(ctx, reflect.ValueOf(n).Len())
	})
}

func SliceLenEQ[T SlicePattern](m *Matcher, n int) T {
	return SliceLenOf[T](m, func(ctx *MatchCtx, len int) bool { return len == n })
}

func SliceLenGT[T SlicePattern](m *Matcher, n int) T {
	return SliceLenOf[T](m, func(ctx *MatchCtx, len int) bool { return len > n })
}

func SliceLenGE[T SlicePattern](m *Matcher, n int) T {
	return SliceLenOf[T](m, func(ctx *MatchCtx, len int) bool { return len >= n })
}

func SliceLenLT[T SlicePattern](m *Matcher, n int) T {
	return SliceLenOf[T](m, func(ctx *MatchCtx, len int) bool { return len < n })
}

func SliceLenLE[T SlicePattern](m *Matcher, n int) T {
	return SliceLenOf[T](m, func(ctx *MatchCtx, len int) bool { return len >= n })
}
