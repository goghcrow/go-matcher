package combinator

import (
	"go/ast"
	"go/types"

	"github.com/goghcrow/go-matcher"
	"golang.org/x/tools/go/types/typeutil"
)

// ref testdata/match/callee.txt

// CalleeOf a builtin / function / method / var call
func CalleeOf(m *Matcher, p Predicate[types.Object]) CallExprPattern {
	return matcher.MkPattern[CallExprPattern](m, func(n ast.Node, ctx *MatchCtx) bool {
		if n == nil /*ast.Node(nil)*/ {
			return false
		}
		call := n.(*ast.CallExpr)
		if call == nil {
			return false
		}

		callee := typeutil.Callee(ctx.TypeInfo(), call)
		if callee == nil {
			return false
		}
		return p(ctx, callee)
	})
}

// BuiltinCalleeOf a builtin function call
func BuiltinCalleeOf(m *Matcher, p Predicate[*types.Builtin]) CallExprPattern {
	return CalleeOf(m, func(ctx *MatchCtx, callee types.Object) bool {
		if f, ok := callee.(*types.Builtin); ok {
			return p(ctx, f)
		}
		return false
	})
}

// BuiltinCallee match fun exactly
func BuiltinCallee(m *Matcher, fun string) CallExprPattern {
	builtIn := types.Universe.Lookup(fun)
	assert(builtIn != nil, fun+" not found")

	return BuiltinCalleeOf(m, func(ctx *MatchCtx, callee *types.Builtin) bool {
		return callee.Name() == fun &&
			callee.Type() == builtIn.Type()
	})
}

// VarCalleeOf a var call
func VarCalleeOf(m *Matcher, p Predicate[*types.Var]) CallExprPattern {
	return CalleeOf(m, func(ctx *MatchCtx, callee types.Object) bool {
		if f, ok := callee.(*types.Var); ok {
			return p(ctx, f)
		}
		return false
	})
}

// FuncOrMethodCalleeOf a function or method call, exclude builtin and var call
func FuncOrMethodCalleeOf(m *Matcher, p Predicate[*types.Func]) CallExprPattern {
	return CalleeOf(m, func(ctx *MatchCtx, callee types.Object) bool {
		if f, ok := callee.(*types.Func); ok {
			return p(ctx, f)
		}
		return false
	})
}

func FuncCalleeOf(m *Matcher, p Predicate[*types.Func]) CallExprPattern {
	return CalleeOf(m, func(ctx *MatchCtx, callee types.Object) bool {
		if f, ok := callee.(*types.Func); ok {
			recv := f.Type().(*types.Signature).Recv()
			return recv == nil && p(ctx, f)
		}
		return false
	})
}

// FuncCallee match pkg.fun exactly
func FuncCallee(m *Matcher, funObj types.Object /*pkg,*/, fun string) CallExprPattern {
	// qualified := pkg + "." + fun
	// funObj := m.Lookup(qualified)
	// assert(funObj != nil, qualified+" not found")

	_, isFunc := funObj.Type().(*types.Signature)
	assert(isFunc, funObj.String()+" not func")

	return FuncCalleeOf(m, func(ctx *MatchCtx, f *types.Func) bool {
		return f.Name() == fun &&
			funObj.Type() == f.Type()
	})
}

func MethodCalleeOf(m *Matcher, p Predicate[*types.Func]) CallExprPattern {
	return CalleeOf(m, func(ctx *MatchCtx, callee types.Object) bool {
		if f, ok := callee.(*types.Func); ok {
			recv := f.Type().(*types.Signature).Recv()
			return recv != nil && p(ctx, f)
		}
		return false
	})
}

// MethodCallee match pkg.typ.method exactly
// addressable means whether the receiver is addressable
func MethodCallee(m *Matcher, tyObj types.Object /*pkg, typ, */, method string, addressable bool) CallExprPattern {
	// qualified := pkg + "." + typ
	// tyObj := m.Lookup(qualified)
	// assert(tyObj != nil, qualified+" not found")

	methodObj, _, _ := types.LookupFieldOrMethod(tyObj.Type(), addressable, tyObj.Pkg(), method)
	assert(methodObj != nil, method+" not found")

	_, isFunc := methodObj.Type().(*types.Signature)
	assert(isFunc, method+" not func")

	return MethodCalleeOf(m, func(ctx *MatchCtx, f *types.Func) bool {
		return f.Name() == method &&
			f.Type() == methodObj.Type()
	})
}

// StaticCalleeOf a static function (or method) call, exclude var / builtin call
func StaticCalleeOf(m *Matcher, p Predicate[*types.Func]) CallExprPattern {
	return FuncOrMethodCalleeOf(m, func(ctx *MatchCtx, f *types.Func) bool {
		recv := f.Type().(*types.Signature).Recv()
		isIfaceRecv := recv != nil && types.IsInterface(recv.Type())
		return !isIfaceRecv && p(ctx, f)
	})
}

func IfaceCalleeOf(m *Matcher, p Predicate[*types.Func]) CallExprPattern {
	return MethodCalleeOf(m, func(ctx *MatchCtx, f *types.Func) bool {
		recv := f.Type().(*types.Signature).Recv()
		return types.IsInterface(recv.Type()) && p(ctx, f)
	})
}

// IfaceCallee match pkg.iface.method exactly
func IfaceCallee(m *Matcher, ifaceObj types.Object /*pkg, iface, */, method string) CallExprPattern {
	// qualified := pkg + "." + iface
	// ifaceObj := m.Lookup(qualified)
	assert(ifaceObj != nil, ifaceObj.String()+" not found")

	// types.Named -> types.Interface
	_, isIface := ifaceObj.Type().Underlying().(*types.Interface)
	assert(isIface, ifaceObj.String()+" not interface")

	methodObj, _, _ := types.LookupFieldOrMethod(ifaceObj.Type(), false, ifaceObj.Pkg(), method)
	assert(methodObj != nil, method+" not found")

	_, isFunc := methodObj.Type().(*types.Signature)
	assert(isFunc, method+" not func")

	return IfaceCalleeOf(m, func(ctx *MatchCtx, f *types.Func) bool {
		return f.Name() == method &&
			f.Type() == methodObj.Type()
	})
}
