package example

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/goghcrow/go-matcher"
	. "github.com/goghcrow/go-matcher/combinator"
)

func PatternOfGormTabler(m *Matcher, gormTabler *types.Interface) ast.Node {
	// types.Named -> types.Interface
	// gormTabler := l.MustLookupType("gorm.io/gorm/schema.Tabler").Underlying().(*types.Interface)
	return &ast.TypeSpec{
		// must impl gorm.Tabler, and bind node to variable
		Name: Bind(m, "typeId",
			Or(m,
				TypeOf[IdentPattern](m, func(_ *MatchCtx, t types.Type) bool {
					return types.Implements(t, gormTabler)
				}),
				TypeOf[IdentPattern](m, func(_ *MatchCtx, t types.Type) bool {
					return types.Implements(types.NewPointer(t), gormTabler)
				}),
			),
		),
		// wildcard pattern can be ignored
		// TypeParams: Wildcard[FieldListPattern](m), // ignore type params
		// Type: Wildcard[ExprPattern](m), // ignore type
	}
}

func PatternOfGormTablerTableName(m *Matcher, gormTabler *types.Interface) ast.Node {
	return &ast.FuncDecl{
		// recv must impl gorm.Tabler
		// Recv: MethodRecvOf(m, func(recv *ast.Field) bool {
		// 	ty := m.TypeOf(recv.Type)
		// 	return types.Implements(ty, gormTabler) ||
		// 		types.Implements(types.NewPointer(ty), gormTabler)
		// }),
		// method name must be TableName
		// Name: IdentNameOf(m, "TableName"),

		Name: And[IdentPattern](m,
			// IsMethod(m),
			IdentNameOf(m, "TableName"),
			IdentRecvTypeOf(m, func(_ *MatchCtx, ty types.Type) bool {
				return types.Implements(ty, gormTabler) ||
					types.Implements(types.NewPointer(ty), gormTabler)
			}),
		),

		// ignore type signature
		// Type: Wildcard[FuncTypePattern](m),
		// Body match { return "xxxx" } or { return `xxxx` },
		// and binding stringLit to the variable of tableName
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						Bind[ExprPattern](m, "tableName", LitKindOf(m, token.STRING)),

						// // Or
						// Bind[BasicLitPattern](m, "tableName", matcher.MkPattern[BasicLitPattern](m, func(n ast.Node, ctx *matcher.MatchCtx) bool {
						// 	lit, _ := n.(*ast.BasicLit)
						// 	if lit == nil {
						// 		return false
						// 	}
						// 	return lit.Kind == token.STRING
						// })),
					},
				},
			},
		},
	}
}

func PatternOfNonCompositeModelCall1(m *Matcher, gormDB types.Object) ast.Node {
	// db.Model(&Model{})
	// db.Model(Model{})
	// gormDB := .Loader.MustLookup("gorm.io/gorm.DB")
	return And(m,
		MethodCallee(m, gormDB, "Model", true),
		matcher.PatternOf[CallExprPattern](m, &ast.CallExpr{
			Args: []ast.Expr{
				Not(m,
					Or(m,
						matcher.PatternOf[ExprPattern](m, PtrOf(&ast.CompositeLit{})),
						matcher.PatternOf[ExprPattern](m, &ast.CompositeLit{}),
					),
				),
			},
		}),
	)
}

func PatternOfNonCompositeModelCall2(m *Matcher, gormDB types.Object) func() ast.Node {
	// gormDB := .Loader.MustLookupType("gorm.io/gorm.DB")
	gormDBPtr := types.NewPointer(gormDB.Type())
	return func() ast.Node {
		// db.Model(&Model{})
		// db.Model(Model{})
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   TypeAssignableTo[ExprPattern](m, gormDBPtr),
				Sel: IdentNameOf(m, "Model"),
			},
			Args: []ast.Expr{
				Not(m,
					Or(m,
						matcher.PatternOf[ExprPattern](m, PtrOf(&ast.CompositeLit{})),
						matcher.PatternOf[ExprPattern](m, &ast.CompositeLit{}),
					),
				),
			},
		}
	}
}
