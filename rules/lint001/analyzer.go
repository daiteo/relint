package lint001

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint001",
	Doc:      "LINT-001: slog key args must be lowercase_snake_case or dot-notation",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !analysisutil.IsSlogCall(pass, call) {
			return
		}

		args := call.Args
		// args[0] = message, args[1] = key1, args[2] = val1, ...
		// keys are at odd indices: 1, 3, 5, ...
		for i := 1; i < len(args); i += 2 {
			lit, ok := args[i].(*ast.BasicLit)
			if !ok {
				continue
			}
			// Strip quotes
			key := lit.Value
			if len(key) < 2 {
				continue
			}
			key = key[1 : len(key)-1]

			if !analysisutil.IsSnakeLower(key) {
				pass.Reportf(lit.Pos(), "LINT-001: slog key %q must be lowercase_snake_case", key)
			}
		}
	})

	return nil, nil
}
