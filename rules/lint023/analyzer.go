package lint023

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint023",
	Doc:      "LINT-023: route Input/Output types in module-scoped *handler packages must be in {route}.go",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if !analysisutil.IsHandlerPackage(pass.Pkg.Name()) || pass.Pkg.Name() == "handler" {
		return nil, nil
	}

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
			routeName, ok := routeNameFromType(name)
			if !ok {
				continue
			}

			expectedFiles := findExpectedRouteFiles(pass, routeName)
			if len(expectedFiles) == 0 {
				continue
			}

			actualFile := analysisutil.FileBasename(pass, ts.Name.Pos())
			matched := false
			for _, expected := range expectedFiles {
				if actualFile == expected {
					matched = true
					break
				}
			}
			if matched {
				continue
			}

			pass.Reportf(ts.Name.Pos(), "LINT-023: type %q must be declared in route file %q", name, expectedFiles[0])
		}
	})

	return nil, nil
}

func routeNameFromType(typeName string) (string, bool) {
	if strings.HasSuffix(typeName, "Input") {
		routeName := strings.TrimSuffix(typeName, "Input")
		return routeName, routeName != ""
	}
	if strings.HasSuffix(typeName, "Output") {
		routeName := strings.TrimSuffix(typeName, "Output")
		return routeName, routeName != ""
	}
	return "", false
}

func findExpectedRouteFiles(pass *analysis.Pass, routeName string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if !fn.Name.IsExported() || fn.Name.Name != routeName {
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

			handlerName := strings.TrimSuffix(recvIdent.Name, "Handler")
			routeFile := expectedRouteFile(handlerName, routeName)
			if !seen[routeFile] {
				out = append(out, routeFile)
				seen[routeFile] = true
			}
		}
	}
	return out
}

func expectedRouteFile(handlerName, routeName string) string {
	handlerSnake := analysisutil.ToSnake(handlerName)
	routeSnake := analysisutil.ToSnake(routeName)
	routePart := normalizeRoutePart(handlerSnake, routeSnake)
	if routePart != "" {
		return fmt.Sprintf("%s.go", routePart)
	}
	return fmt.Sprintf("%s.go", routeSnake)
}

func normalizeRoutePart(handlerSnake, routeSnake string) string {
	routePart := routeSnake
	aliases := []string{handlerSnake, pluralize(handlerSnake)}
	for _, alias := range aliases {
		routePart = strings.TrimPrefix(routePart, alias+"_")
		routePart = strings.TrimSuffix(routePart, "_"+alias)
	}
	routePart = strings.Trim(routePart, "_")
	if routePart == handlerSnake || routePart == pluralize(handlerSnake) {
		return ""
	}
	return routePart
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		prev := s[len(s)-2]
		if !strings.ContainsRune("aeiou", rune(prev)) {
			return s[:len(s)-1] + "ies"
		}
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "z") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}
