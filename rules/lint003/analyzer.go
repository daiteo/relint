package lint003

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var dotNotationFlag string

var Analyzer = &analysis.Analyzer{
	Name:     "lint003",
	Doc:      "LINT-003: slog keys that belong to a group must use dot notation",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func init() {
	Analyzer.Flags.StringVar(
		&dotNotationFlag,
		"dot-notation",
		"",
		`comma-separated key=dotted pairs of slog keys that must use dot notation, e.g. "error=error.message,userId=user.id"`,
	)
}

// parseDotNotation parses the flag value into a map[bad-key]suggested-key.
func parseDotNotation() map[string]string {
	m := make(map[string]string)
	for _, pair := range strings.Split(dotNotationFlag, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return m
}

func run(pass *analysis.Pass) (interface{}, error) {
	dotNotationMap := parseDotNotation()
	if len(dotNotationMap) == 0 {
		return nil, nil
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !analysisutil.IsSlogCall(pass, call) {
			return
		}

		args := call.Args
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		name := sel.Sel.Name
		keyStart := 1
		if name == "Log" || name == "LogAttrs" {
			keyStart = 3
		} else if name == "InfoContext" || name == "ErrorContext" || name == "WarnContext" || name == "DebugContext" {
			keyStart = 2
		}

		for i := keyStart; i < len(args); i += 2 {
			lit, ok := args[i].(*ast.BasicLit)
			if !ok {
				continue
			}
			key := lit.Value
			if len(key) < 2 {
				continue
			}
			key = key[1 : len(key)-1]

			if suggested, bad := dotNotationMap[key]; bad {
				pass.Reportf(lit.Pos(), "LINT-003: slog key %q should use dot notation, e.g. %q", key, suggested)
			}
		}
	})

	return nil, nil
}
