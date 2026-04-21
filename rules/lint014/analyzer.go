package lint014

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
	Name:     "lint014",
	Doc:      "LINT-014: service structs must have compile-time interface assertion in service.go",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()
	if !strings.Contains(pkgName, "service") {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Find service.go file
	var serviceFile *ast.File
	for _, f := range pass.Files {
		if analysisutil.FileBasename(pass, f.Pos()) == "service.go" {
			serviceFile = f
			break
		}
	}

	// Collect exported Service structs
	var serviceStructs []string
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
			if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
				continue
			}
			name := ts.Name.Name
			if !ts.Name.IsExported() {
				continue
			}
			if strings.HasSuffix(name, "Service") {
				serviceStructs = append(serviceStructs, name)
			}
		}
	})

	for _, structName := range serviceStructs {
		if serviceFile == nil || !hasInterfaceAssertion(serviceFile, structName) {
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
					if ts.Name.Name == structName {
						pass.Reportf(ts.Name.Pos(), "LINT-014: service struct %q missing compile-time interface assertion in service.go", structName)
					}
				}
			})
		}
	}

	return nil, nil
}

func hasInterfaceAssertion(f *ast.File, structName string) bool {
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
				if name.Name != "_" {
					continue
				}
				// Match the value side: (*StructName)(nil)
				// The type side is the interface (e.g. types.MetricsService) which may
				// differ from the struct name, so we must not compare against it.
				if len(vs.Values) > 0 && assertionValueMatchesStruct(vs.Values[0], structName) {
					return true
				}
			}
		}
	}
	return false
}

// assertionValueMatchesStruct reports whether expr matches the pattern (*StructName)(nil).
func assertionValueMatchesStruct(expr ast.Expr, structName string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	paren, ok := call.Fun.(*ast.ParenExpr)
	if !ok {
		return false
	}
	star, ok := paren.X.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == structName
}
