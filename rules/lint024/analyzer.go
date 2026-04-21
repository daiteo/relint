package lint024

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint024",
	Doc:      "LINT-024: body helper types in packages ending with handler must be named {Name}BodyInput or {Name}BodyOutput",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

// validBodyPattern matches explicit body helper names like XBodyInput or XBodyOutput
var validBodyPattern = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*Body(Input|Output)$`)

func run(pass *analysis.Pass) (interface{}, error) {
	if !analysisutil.IsHandlerPackage(pass.Pkg.Name()) {
		return nil, nil
	}

	bodyUsage := analysisutil.AnalyzeBodyTypeUsage(pass)
	routeFiles := collectRouteFiles(pass)
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.GenDecl)(nil)}, func(n ast.Node) {
		gd := n.(*ast.GenDecl)
		if gd.Tok != token.TYPE {
			return
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			name := ts.Name.Name
			if !strings.Contains(name, "Body") {
				continue
			}
			if bodyUsage.BodyOnlyStructs[name] {
				// Nested body-only helper structs are validated by LINT-026.
				continue
			}

			// Only check in non-route files (files not ending with _handler.go)
			filename := analysisutil.FileBasename(pass, ts.Name.Pos())
			if strings.HasSuffix(filename, "_handler.go") || routeFiles[filename] {
				continue
			}

			if !validBodyPattern.MatchString(name) {
				pass.Reportf(ts.Name.Pos(), "LINT-024: body type %q must be named \"{Name}BodyInput\" or \"{Name}BodyOutput\"", name)
			}
		}
	})

	return nil, nil
}

func collectRouteFiles(pass *analysis.Pass) map[string]bool {
	routeFiles := make(map[string]bool)
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if !fn.Name.IsExported() {
				continue
			}
			recv := fn.Recv.List[0].Type
			if star, ok := recv.(*ast.StarExpr); ok {
				recv = star.X
			}
			recvIdent, ok := recv.(*ast.Ident)
			if !ok || !strings.HasSuffix(recvIdent.Name, "Handler") {
				continue
			}
			routeFiles[analysisutil.FileBasename(pass, fn.Name.Pos())] = true
		}
	}
	return routeFiles
}
