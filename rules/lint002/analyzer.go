package lint002

import (
	"go/ast"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint002",
	Doc:      "LINT-002: slog message must start with a lowercase letter",
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
		if len(args) == 0 {
			return
		}

		// Check if it's slog.Log / slog.LogContext (first arg is context/level, not message)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		name := sel.Sel.Name
		// slog.Log(ctx, level, msg, ...) — msg is at index 2
		// slog.LogAttrs(ctx, level, msg, ...) — msg is at index 2
		msgIdx := 0
		if name == "Log" || name == "LogAttrs" {
			msgIdx = 2
		} else if name == "InfoContext" || name == "ErrorContext" || name == "WarnContext" || name == "DebugContext" {
			msgIdx = 1
		}

		if msgIdx >= len(args) {
			return
		}

		lit, ok := args[msgIdx].(*ast.BasicLit)
		if !ok {
			return
		}
		msg := lit.Value
		if len(msg) < 2 {
			return
		}
		msg = msg[1 : len(msg)-1]
		if msg == "" {
			return
		}

		runes := []rune(msg)
		if unicode.IsUpper(runes[0]) {
			// Reconstruct the literal with the first content rune lowercased.
			// lit.Value is the raw source including quotes, e.g. `"User created"`.
			quote := string(lit.Value[0]) // '"' or '`'
			fixed := quote + string(unicode.ToLower(runes[0])) + string(runes[1:]) + quote
			pass.Report(analysis.Diagnostic{
				Pos:     lit.Pos(),
				Message: "LINT-002: slog message " + lit.Value + " must start with a lowercase letter",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message: "Lowercase the first letter",
					TextEdits: []analysis.TextEdit{{
						Pos:     lit.Pos(),
						End:     lit.End(),
						NewText: []byte(fixed),
					}},
				}},
			})
		}
	})

	return nil, nil
}
