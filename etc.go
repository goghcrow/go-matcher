package matcher

import (
	"flag"
	"go/ast"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// https://stackoverflow.com/questions/14249217/how-do-i-know-im-running-within-go-test
var runningWithGoTest = flag.Lookup("test.v") != nil ||
	strings.HasSuffix(os.Args[0], ".test")

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func preOrder(root ast.Node, f astutil.ApplyFunc) {
	astutil.Apply(root, f, nil)
}

func postOrder(root ast.Node, f astutil.ApplyFunc) {
	astutil.Apply(root, nil, f)
}

func IsNilNode(n ast.Node) bool {
	if n == nil {
		return true
	}
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() {
		return true
	}
	return false
}
