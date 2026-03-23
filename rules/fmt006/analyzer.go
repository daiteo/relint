package fmt006

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "fmt006",
	Doc:      "FMT-006: blocks ending with a return must be followed by a blank line in the parent block",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.BlockStmt)(nil)}, func(n ast.Node) {
		block := n.(*ast.BlockStmt)
		file := fileOf(pass, block.Pos())
		if file == nil {
			return
		}
		checkBlock(pass, file, block)
	})

	return nil, nil
}

func fileOf(pass *analysis.Pass, pos token.Pos) *ast.File {
	for _, f := range pass.Files {
		if f.Pos() <= pos && pos <= f.End() {
			return f
		}
	}
	return nil
}

func checkBlock(pass *analysis.Pass, file *ast.File, block *ast.BlockStmt) {
	if len(block.List) < 2 {
		return
	}

	for i := 1; i < len(block.List); i++ {
		prev := block.List[i-1]
		curr := block.List[i]

		if !stmtContainsTrailingReturn(prev) {
			continue
		}

		if blankLinesBetween(pass, file, prev.End(), curr.Pos()) != 1 {
			pass.Reportf(curr.Pos(), "FMT-006: parent block must contain a blank line after a block ending with return")
		}
	}
}

func stmtContainsTrailingReturn(stmt ast.Stmt) bool {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return blockEndsWithReturn(s)
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		if blockEndsWithReturn(s.Body) {
			return true
		}
		return stmtContainsTrailingReturn(s.Else)
	case *ast.ForStmt:
		return blockEndsWithReturn(s.Body)
	case *ast.RangeStmt:
		return blockEndsWithReturn(s.Body)
	case *ast.SwitchStmt:
		return caseClausesContainTrailingReturn(s.Body.List)
	case *ast.TypeSwitchStmt:
		return caseClausesContainTrailingReturn(s.Body.List)
	case *ast.SelectStmt:
		return commClausesContainTrailingReturn(s.Body.List)
	case *ast.LabeledStmt:
		return stmtContainsTrailingReturn(s.Stmt)
	default:
		return false
	}
}

func blockEndsWithReturn(block *ast.BlockStmt) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}
	return stmtContainsTrailingReturn(block.List[len(block.List)-1])
}

func caseClausesContainTrailingReturn(list []ast.Stmt) bool {
	for _, stmt := range list {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok || len(clause.Body) == 0 {
			continue
		}
		if stmtContainsTrailingReturn(clause.Body[len(clause.Body)-1]) {
			return true
		}
	}
	return false
}

func commClausesContainTrailingReturn(list []ast.Stmt) bool {
	for _, stmt := range list {
		clause, ok := stmt.(*ast.CommClause)
		if !ok || len(clause.Body) == 0 {
			continue
		}
		if stmtContainsTrailingReturn(clause.Body[len(clause.Body)-1]) {
			return true
		}
	}
	return false
}

func blankLinesBetween(pass *analysis.Pass, file *ast.File, from, to token.Pos) int {
	fset := pass.Fset
	fromLine := fset.Position(from).Line
	toLine := fset.Position(to).Line

	if toLine <= fromLine+1 {
		return 0
	}

	commentLines := make(map[int]bool)
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			line := fset.Position(c.Slash).Line
			if line > fromLine && line < toLine {
				commentLines[line] = true
			}
		}
	}

	blank := 0
	for line := fromLine + 1; line < toLine; line++ {
		if !commentLines[line] {
			blank++
		}
	}
	return blank
}
