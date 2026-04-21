package lint022

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint022",
	Doc:      "LINT-022: route methods in module-scoped *handler packages must be in {route}.go files",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if !analysisutil.IsHandlerPackage(pass.Pkg.Name()) || pass.Pkg.Name() == "handler" {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}

		recv := fn.Recv.List[0].Type
		if star, ok := recv.(*ast.StarExpr); ok {
			recv = star.X
		}
		recvIdent, ok := recv.(*ast.Ident)
		if !ok {
			return
		}
		recvName := recvIdent.Name
		if !strings.HasSuffix(recvName, "Handler") {
			return
		}

		methodName := fn.Name.Name
		if !fn.Name.IsExported() {
			return
		}

		// Compute expected file names
		// handlerName without "Handler" suffix, lowercased
		baseName := strings.TrimSuffix(recvName, "Handler")
		expectedFiles := expectedRouteFiles(baseName, methodName)
		actualFile := analysisutil.FileBasename(pass, fn.Name.Pos())

		matched := slices.Contains(expectedFiles, actualFile)

		if !matched {
			pass.Reportf(fn.Name.Pos(), "LINT-022: route handler %q on %q must be in file %q", methodName, recvName, expectedFiles[0])
		}
	})

	return nil, nil
}

func expectedRouteFiles(handlerName, routeName string) []string {
	handlerSnake := toSnake(handlerName)
	routeSnake := toSnake(routeName)
	routePart := normalizeRoutePart(handlerSnake, routeSnake)
	if routePart != "" {
		return []string{fmt.Sprintf("%s.go", routePart)}
	}
	return []string{fmt.Sprintf("%s.go", routeSnake)}
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
