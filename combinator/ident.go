package combinator

import (
	"go/ast"
	"go/types"
	"regexp"

	"github.com/goghcrow/go-matcher"
)

func IdentOf(m *Matcher, p Predicate[*ast.Ident]) IdentPattern {
	return matcher.MkPattern[IdentPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if !ok || ident == nil {
			return false
		}
		return p(ctx, ident)
	})
}

func IdentObjectOf(m *Matcher, p Predicate[types.Object]) IdentPattern {
	return IdentOf(m, func(ctx *MatchCtx, id *ast.Ident) bool {
		obj := ctx.ObjectOf(id)
		if obj == nil {
			return false
		}
		return p(ctx, obj)
	})
}

func IdentNameOf(m *Matcher, name string) IdentPattern {
	return IdentOf(m, func(ctx *MatchCtx, id *ast.Ident) bool {
		return name == id.Name
	})
}

func IdentNameMatch(m *Matcher, reg *regexp.Regexp) IdentPattern {
	return IdentOf(m, func(ctx *MatchCtx, id *ast.Ident) bool {
		return reg.Match([]byte(id.Name))
	})
}

func IdentTypeOf(m *Matcher, p Predicate[types.Type]) IdentPattern {
	// return TypeOf[IdentPattern](m, p)
	return IdentObjectOf(m, func(ctx *MatchCtx, obj types.Object) bool {
		return p(ctx, obj.Type())
	})
}

func IdentSigOf(m *Matcher, p Predicate[*types.Signature]) IdentPattern {
	return IdentTypeOf(m, func(ctx *MatchCtx, t types.Type) bool {
		if sig, ok := t.(*types.Signature); ok {
			return p(ctx, sig)
		}
		return false
	})
}

func IdentRecvOf(m *Matcher, p Predicate[*types.Var]) IdentPattern {
	return IdentSigOf(m, func(ctx *MatchCtx, sig *types.Signature) bool {
		if sig.Recv() == nil {
			return false
		}
		return p(ctx, sig.Recv())
	})
}

// IdentRecvTypeOf for ast.FuncDecl { Name }
func IdentRecvTypeOf(m *Matcher, p Predicate[types.Type]) IdentPattern {
	return IdentRecvOf(m, func(ctx *MatchCtx, recv *types.Var) bool {
		return p(ctx, recv.Type())
	})
}

func IdentIsFun(m *Matcher) IdentPattern {
	return IdentSigOf(m, func(ctx *MatchCtx, sig *types.Signature) bool {
		return sig.Recv() == nil
	})
}

func IdentIsMethod(m *Matcher) IdentPattern {
	return IdentSigOf(m, func(ctx *MatchCtx, sig *types.Signature) bool {
		return sig.Recv() != nil
	})
}

func IdentParamsOf(m *Matcher, p Predicate[*types.Tuple]) IdentPattern {
	return IdentSigOf(m, func(ctx *MatchCtx, sig *types.Signature) bool {
		return p(ctx, sig.Params())
	})
}

func IdentAnyParamOf(m *Matcher, p Predicate[*types.Var]) IdentPattern {
	return IdentParamsOf(m, func(ctx *MatchCtx, params *types.Tuple) bool {
		for i, n := 0, params.Len(); i < n; i++ {
			if p(ctx, params.At(i)) {
				return true
			}
		}
		return false
	})
}

func IsBuiltin(m *Matcher) IdentPattern {
	return IdentObjectOf(m, func(ctx *MatchCtx, obj types.Object) bool {
		_, ok := obj.(*types.Builtin)
		return ok
	})
}
