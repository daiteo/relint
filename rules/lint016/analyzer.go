package lint016

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint016",
	Doc:      "LINT-016: Inject* middleware in packages ending with handler must be in inject_{name}.go",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if !analysisutil.IsHandlerPackage(pass.Pkg.Name()) {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		name := fn.Name.Name
		suffix, ok := extractSuffix(name, "Inject", "inject")
		if !ok || suffix == "" {
			return
		}
		expectedFile := fmt.Sprintf("inject_%s.go", toSnake(suffix))
		actualFile := analysisutil.FileBasename(pass, fn.Name.Pos())
		if actualFile != expectedFile {
			pass.Reportf(fn.Name.Pos(), "LINT-016: middleware %q must be in file %q", name, expectedFile)
		}
	})

	return nil, nil
}

func extractSuffix(name string, prefixes ...string) (string, bool) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return name[len(prefix):], true
		}
	}
	return "", false
}

func toSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}
