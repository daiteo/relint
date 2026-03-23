package fmtfix

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "fmtfix",
	Doc:      "FMTFIX: merge consecutive type/const/var declarations, normalize type-block spacing and return-block spacing, and reorder top-level declarations (type, const, var, func)",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	_ = insp

	for _, f := range pass.Files {
		checkFile(pass, f)
	}
	return nil, nil
}

// declItem is either a group of consecutive single type/const/var decls (to be merged)
// or a single other decl (non-grouped declaration).
type declItem struct {
	order      int
	groupTok   token.Token
	groupDecls []*ast.GenDecl // non-nil: consecutive single type/const/var decls
	decl       ast.Decl       // non-nil: everything else
}

func declOrder(d ast.Decl) int {
	gd, ok := d.(*ast.GenDecl)
	if !ok {
		return 3
	}
	switch gd.Tok {
	case token.TYPE:
		return 0
	case token.CONST:
		return 1
	case token.VAR:
		return 2
	}
	return 4
}

// declStart returns the start position of a declaration including its doc comment.
// Using d.Pos() alone would leave the doc comment orphaned before the TextEdit range.
func declStart(d ast.Decl) token.Pos {
	switch v := d.(type) {
	case *ast.GenDecl:
		if v.Doc != nil {
			return v.Doc.Pos()
		}
	case *ast.FuncDecl:
		if v.Doc != nil {
			return v.Doc.Pos()
		}
	}
	return d.Pos()
}

// buildDeclItems groups consecutive single (non-parenthesised) type/const/var decls together;
// everything else is an individual item. Import decls must be excluded before calling.
func buildDeclItems(decls []ast.Decl) []declItem {
	var items []declItem
	i := 0
	for i < len(decls) {
		gd, ok := decls[i].(*ast.GenDecl)
		if ok && isMergeableGenDecl(gd) {
			j := i + 1
			for j < len(decls) {
				next, ok2 := decls[j].(*ast.GenDecl)
				if !ok2 || !isMergeableGenDecl(next) || next.Tok != gd.Tok {
					break
				}
				j++
			}
			group := make([]*ast.GenDecl, j-i)
			for k := range group {
				group[k] = decls[i+k].(*ast.GenDecl)
			}
			items = append(items, declItem{order: declOrder(gd), groupTok: gd.Tok, groupDecls: group})
			i = j
		} else {
			items = append(items, declItem{order: declOrder(decls[i]), decl: decls[i]})
			i++
		}
	}
	return items
}

// mergeAdjacentGroups coalesces same-token groups that end up adjacent after sorting.
func mergeAdjacentGroups(items []declItem) []declItem {
	merged := make([]declItem, 0, len(items))
	for _, item := range items {
		if item.groupDecls != nil &&
			len(merged) > 0 &&
			merged[len(merged)-1].groupDecls != nil &&
			merged[len(merged)-1].groupTok == item.groupTok {
			merged[len(merged)-1].groupDecls = append(merged[len(merged)-1].groupDecls, item.groupDecls...)
		} else {
			merged = append(merged, item)
		}
	}
	return merged
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	// Exclude import decls — they are always first and must not be reordered.
	var decls []ast.Decl
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if ok && gd.Tok == token.IMPORT {
			continue
		}
		decls = append(decls, d)
	}

	if len(decls) == 0 {
		return
	}

	items := buildDeclItems(decls)
	needsTypeSpacing := hasTypeBlockSpacingIssue(pass, f)
	needsReturnBlockSpacing := hasReturnBlockSpacingIssue(pass, f)

	needsMerge := false
	for _, item := range items {
		if len(item.groupDecls) > 1 {
			needsMerge = true
			break
		}
	}

	ordered := true
	for i := 1; i < len(items); i++ {
		if items[i].order < items[i-1].order {
			ordered = false
			break
		}
	}

	if !needsMerge && ordered && !needsTypeSpacing && !needsReturnBlockSpacing {
		return
	}

	sorted := make([]declItem, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].order < sorted[j].order
	})

	// After reordering, groups that are now adjacent should be merged too.
	sorted = mergeAdjacentGroups(sorted)

	// Re-check: if nothing actually changed, skip.
	needsMerge = false
	for _, item := range sorted {
		if len(item.groupDecls) > 1 {
			needsMerge = true
			break
		}
	}
	if !needsMerge && ordered && !needsTypeSpacing && !needsReturnBlockSpacing {
		return
	}

	filename := pass.Fset.File(f.Pos()).Name()
	fileSrc, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	newSrc := generateText(pass, f, sorted, fileSrc)
	if newSrc == nil {
		return
	}

	// TextEdit.Pos must include the doc comment of the first decl so it is not
	// left orphaned before the edit range (which would misplace directives like //go:embed).
	startPos := declStart(decls[0])

	pass.Report(analysis.Diagnostic{
		Pos:     decls[0].Pos(),
		Message: "FMTFIX: apply format fixes (merge declaration blocks, reorder declarations)",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "Apply format fixes",
			TextEdits: []analysis.TextEdit{{
				Pos:     startPos,
				End:     decls[len(decls)-1].End(),
				NewText: newSrc,
			}},
		}},
	})
}

func generateText(pass *analysis.Pass, file *ast.File, items []declItem, fileSrc []byte) []byte {
	var result bytes.Buffer
	for i, item := range items {
		if i > 0 {
			result.WriteString("\n\n")
		}
		if item.decl != nil {
			if _, ok := item.decl.(*ast.FuncDecl); ok {
				funcText := formatFuncDecl(pass, file, item.decl, fileSrc)
				if funcText == nil {
					return nil
				}
				result.Write(funcText)
				continue
			}
			if gd, ok := item.decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE && gd.Lparen != token.NoPos && len(gd.Specs) > 1 {
				block := buildTypeBlockWithSpacing(pass, gd)
				if block == nil {
					return nil
				}
				result.Write(block)
				continue
			}
			if err := format.Node(&result, pass.Fset, item.decl); err != nil {
				return nil
			}
		} else if len(item.groupDecls) == 1 {
			if err := format.Node(&result, pass.Fset, item.groupDecls[0]); err != nil {
				return nil
			}
		} else {
			merged := buildMergedGenDeclBlock(pass, item.groupTok, item.groupDecls)
			if merged == nil {
				return nil
			}
			result.Write(merged)
		}
	}
	return result.Bytes()
}

func formatFuncDecl(pass *analysis.Pass, file *ast.File, decl ast.Decl, fileSrc []byte) []byte {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok {
		return nil
	}

	original := originalDeclText(pass, decl, fileSrc)
	if original == nil {
		return nil
	}

	if fn.Body == nil {
		return original
	}

	return normalizeReturnBlockSpacing(pass, file, fn, original)
}

func originalDeclText(pass *analysis.Pass, decl ast.Decl, fileSrc []byte) []byte {
	file := pass.Fset.File(decl.Pos())
	if file == nil {
		return nil
	}

	start := file.Offset(declStart(decl))
	end := file.Offset(decl.End())
	if start < 0 || end < start || end > len(fileSrc) {
		return nil
	}

	out := make([]byte, end-start)
	copy(out, fileSrc[start:end])
	return out
}

type lineEdit struct {
	start int
	end   int
	lines []string
}

func normalizeReturnBlockSpacing(pass *analysis.Pass, file *ast.File, fn *ast.FuncDecl, src []byte) []byte {
	startLine := pass.Fset.Position(declStart(fn)).Line
	lines := strings.Split(string(src), "\n")
	edits := collectReturnBlockSpacingEdits(pass, file, fn.Body, startLine, lines)

	for _, edit := range edits {
		if edit.start < 0 || edit.end < edit.start || edit.end > len(lines) {
			return src
		}
		replaced := make([]string, 0, len(lines)-(edit.end-edit.start)+len(edit.lines))
		replaced = append(replaced, lines[:edit.start]...)
		replaced = append(replaced, edit.lines...)
		replaced = append(replaced, lines[edit.end:]...)
		lines = replaced
	}

	return []byte(strings.Join(lines, "\n"))
}

func collectReturnBlockSpacingEdits(pass *analysis.Pass, file *ast.File, root *ast.BlockStmt, startLine int, lines []string) []lineEdit {
	var edits []lineEdit

	ast.Inspect(root, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok || len(block.List) < 2 {
			return true
		}

		for i := 1; i < len(block.List); i++ {
			prev := block.List[i-1]
			curr := block.List[i]
			if !stmtContainsTrailingReturn(prev) {
				continue
			}

			if blankLinesBetween(pass, file, prev.End(), curr.Pos()) == 1 {
				continue
			}

			prevEndLine := pass.Fset.Position(prev.End()).Line
			currLine := pass.Fset.Position(curr.Pos()).Line
			startIdx := prevEndLine - startLine + 1
			endIdx := currLine - startLine
			if startIdx < 0 || endIdx < startIdx {
				continue
			}

			var comments []string
			for line := prevEndLine + 1; line < currLine; line++ {
				idx := line - startLine
				if idx < 0 || idx >= len(lines) {
					continue
				}
				text := strings.TrimSpace(lines[idx])
				if text == "" {
					continue
				}
				comments = append(comments, lines[idx])
			}

			middle := make([]string, 0, len(comments)+1)
			middle = append(middle, comments...)
			middle = append(middle, "")

			edits = append(edits, lineEdit{
				start: startIdx,
				end:   endIdx,
				lines: middle,
			})
		}

		return true
	})

	sort.SliceStable(edits, func(i, j int) bool {
		return edits[i].start > edits[j].start
	})

	return edits
}

func hasReturnBlockSpacingIssue(pass *analysis.Pass, file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok || len(block.List) < 2 {
			return true
		}

		for i := 1; i < len(block.List); i++ {
			prev := block.List[i-1]
			curr := block.List[i]
			if stmtContainsTrailingReturn(prev) && blankLinesBetween(pass, file, prev.End(), curr.Pos()) != 1 {
				found = true
				return false
			}
		}

		return !found
	})

	return found
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

// buildMergedGenDeclBlock produces a `{type|const|var} ( ... )` block from a slice
// of single-spec GenDecls.
// It preserves each spec's doc comment and uses format.Node for correct indentation.
func buildMergedGenDeclBlock(pass *analysis.Pass, tok token.Token, decls []*ast.GenDecl) []byte {
	var buf bytes.Buffer
	buf.WriteString(tok.String())
	buf.WriteString(" (\n")
	for i, d := range decls {
		// Include the GenDecl doc comment (e.g. // Foo is ...) indented by one tab.
		if d.Doc != nil {
			for _, c := range d.Doc.List {
				buf.WriteString("\t")
				buf.WriteString(c.Text)
				buf.WriteString("\n")
			}
		}
		// Format the single spec and indent every line by one tab.
		spec := d.Specs[0]
		var specBuf bytes.Buffer
		if err := format.Node(&specBuf, pass.Fset, spec); err != nil {
			return nil
		}
		lines := strings.Split(specBuf.String(), "\n")
		for j, line := range lines {
			if j == len(lines)-1 && line == "" {
				continue // trailing newline from format.Node — skip it
			}
			buf.WriteString("\t")
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		if i < len(decls)-1 {
			buf.WriteString("\n")
		}
	}
	buf.WriteString(")")
	return buf.Bytes()
}

func isMergeableGenDecl(gd *ast.GenDecl) bool {
	if gd == nil || gd.Lparen != token.NoPos {
		return false
	}
	return gd.Tok == token.TYPE || gd.Tok == token.CONST || gd.Tok == token.VAR
}

// buildTypeBlockWithSpacing re-renders an existing type (...) block
// with exactly one blank line between each type spec.
func buildTypeBlockWithSpacing(pass *analysis.Pass, gd *ast.GenDecl) []byte {
	var buf bytes.Buffer
	buf.WriteString("type (\n")
	for i, spec := range gd.Specs {
		var specBuf bytes.Buffer
		if err := format.Node(&specBuf, pass.Fset, spec); err != nil {
			return nil
		}
		lines := strings.Split(specBuf.String(), "\n")
		for j, line := range lines {
			if j == len(lines)-1 && line == "" {
				continue
			}
			buf.WriteString("\t")
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		if i < len(gd.Specs)-1 {
			buf.WriteString("\n")
		}
	}
	buf.WriteString(")")
	return buf.Bytes()
}

func hasTypeBlockSpacingIssue(pass *analysis.Pass, f *ast.File) bool {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE || gd.Lparen == token.NoPos || len(gd.Specs) < 2 {
			continue
		}
		for i := 1; i < len(gd.Specs); i++ {
			if blankLinesBetween(pass, f, gd.Specs[i-1].End(), gd.Specs[i].Pos()) != 1 {
				return true
			}
		}
	}
	return false
}

// blankLinesBetween counts lines without code/comments between two positions.
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
