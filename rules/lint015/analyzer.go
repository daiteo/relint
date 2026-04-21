package lint015

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint015",
	Doc:      "LINT-015: store/service/handler files with layer methods must contain exactly one exported method",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var exemptFiles = map[string]bool{
	"store.go":     true,
	"service.go":   true,
	"fx_module.go": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()
	if !strings.Contains(pkgName, "store") &&
		!strings.Contains(pkgName, "service") &&
		!strings.Contains(pkgName, "handler") {
		return nil, nil
	}

	for _, f := range pass.Files {
		basename := analysisutil.FileBasename(pass, f.Pos())
		if exemptFiles[basename] {
			continue
		}

		count := exportedStoreServiceMethodCount(f)
		if count > 1 {
			pass.Reportf(f.Pos(), "LINT-015: file %q in store/service/handler package must contain exactly one exported store/service/handler method, found %d", basename, count)
		}
	}

	return nil, nil
}

func exportedStoreServiceMethodCount(file *ast.File) int {
	count := 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !fn.Name.IsExported() || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recv := fn.Recv.List[0].Type
		if star, ok := recv.(*ast.StarExpr); ok {
			recv = star.X
		}
		recvIdent, ok := recv.(*ast.Ident)
		if !ok {
			continue
		}
		if strings.HasSuffix(recvIdent.Name, "Store") ||
			strings.HasSuffix(recvIdent.Name, "Service") ||
			strings.HasSuffix(recvIdent.Name, "Handler") {
			count++
		}
	}
	return count
}
