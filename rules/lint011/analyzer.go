package lint011

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint011",
	Doc:      "LINT-011: interfaces in types package must end with Service, Store, or Worker",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "types" {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.GenDecl)(nil)}, func(n ast.Node) {
		gd := n.(*ast.GenDecl)
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, isIface := ts.Type.(*ast.InterfaceType); !isIface {
				continue
			}
			name := ts.Name.Name
			if !strings.HasSuffix(name, "Service") && !strings.HasSuffix(name, "Store") && !strings.HasSuffix(name, "Worker") {
				pass.Reportf(ts.Name.Pos(), "LINT-011: interface %q in types package must end with \"Service\", \"Store\", or \"Worker\"", name)
			}
		}
	})

	return nil, nil
}
