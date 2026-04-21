package lint020

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint020",
	Doc:      "LINT-020: in types packages, error variables (Err...) must be defined in errors.go",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "types" {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	_ = insp

	for _, f := range pass.Files {
		if analysisutil.FileBasename(pass, f.Pos()) == "errors.go" {
			continue
		}
		checkFile(pass, f)
	}

	return nil, nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if strings.HasPrefix(name.Name, "Err") {
					pass.Reportf(name.Pos(), "LINT-020: error variable %q must be defined in errors.go", name.Name)
				}
			}
		}
	}
}
