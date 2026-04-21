package lint018

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint018",
	Doc:      "LINT-018: middleware functions outside packages ending with handler must be named Middleware",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if analysisutil.IsHandlerPackage(pass.Pkg.Name()) {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if !fn.Name.IsExported() {
			return
		}
		if fn.Name.Name == "Middleware" {
			return
		}
		if !isMiddlewareFunc(pass, fn) {
			return
		}
		pass.Reportf(fn.Name.Pos(), "LINT-018: middleware function %q outside packages ending with handler must be named \"Middleware\"", fn.Name.Name)
	})

	return nil, nil
}

// isMiddlewareFunc returns true if fn has signature func(http.Handler) http.Handler
func isMiddlewareFunc(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	ft := fn.Type
	if ft.Params == nil || ft.Results == nil {
		return false
	}
	// Must have exactly 1 param and 1 result
	if len(ft.Params.List) != 1 || len(ft.Results.List) != 1 {
		return false
	}
	paramCount := 0
	for _, f := range ft.Params.List {
		if len(f.Names) == 0 {
			paramCount++
		} else {
			paramCount += len(f.Names)
		}
	}
	resultCount := 0
	for _, f := range ft.Results.List {
		if len(f.Names) == 0 {
			resultCount++
		} else {
			resultCount += len(f.Names)
		}
	}
	if paramCount != 1 || resultCount != 1 {
		return false
	}

	paramType := pass.TypesInfo.TypeOf(ft.Params.List[0].Type)
	resultType := pass.TypesInfo.TypeOf(ft.Results.List[0].Type)

	return isHTTPHandler(paramType) && isHTTPHandler(resultType)
}

func isHTTPHandler(t types.Type) bool {
	if t == nil {
		return false
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == "net/http" && obj.Name() == "Handler"
}
