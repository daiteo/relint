package all

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/alexisvisco/relint/rules/fmt001"
	"github.com/alexisvisco/relint/rules/fmt002"
	"github.com/alexisvisco/relint/rules/fmt003"
	"github.com/alexisvisco/relint/rules/fmt004"
	"github.com/alexisvisco/relint/rules/fmt005"
	"github.com/alexisvisco/relint/rules/fmt006"
	"github.com/alexisvisco/relint/rules/fmtfix"
	"github.com/alexisvisco/relint/rules/lint001"
	"github.com/alexisvisco/relint/rules/lint002"
	"github.com/alexisvisco/relint/rules/lint003"
	"github.com/alexisvisco/relint/rules/lint004"
	"github.com/alexisvisco/relint/rules/lint005"
	"github.com/alexisvisco/relint/rules/lint006"
	"github.com/alexisvisco/relint/rules/lint007"
	"github.com/alexisvisco/relint/rules/lint008"
	"github.com/alexisvisco/relint/rules/lint009"
	"github.com/alexisvisco/relint/rules/lint010"
	"github.com/alexisvisco/relint/rules/lint011"
	"github.com/alexisvisco/relint/rules/lint012"
	"github.com/alexisvisco/relint/rules/lint013"
	"github.com/alexisvisco/relint/rules/lint014"
	"github.com/alexisvisco/relint/rules/lint015"
	"github.com/alexisvisco/relint/rules/lint016"
	"github.com/alexisvisco/relint/rules/lint017"
	"github.com/alexisvisco/relint/rules/lint018"
	"github.com/alexisvisco/relint/rules/lint019"
	"github.com/alexisvisco/relint/rules/lint020"
	"github.com/alexisvisco/relint/rules/lint021"
	"github.com/alexisvisco/relint/rules/lint022"
	"github.com/alexisvisco/relint/rules/lint023"
	"github.com/alexisvisco/relint/rules/lint024"
	"github.com/alexisvisco/relint/rules/lint025"
	"github.com/alexisvisco/relint/rules/lint026"
	"github.com/alexisvisco/relint/rules/lint027"
	"github.com/alexisvisco/relint/rules/lint028"
	"github.com/alexisvisco/relint/rules/lint029"
	"github.com/alexisvisco/relint/rules/lint030"
	"github.com/alexisvisco/relint/rules/lint031"
	"github.com/alexisvisco/relint/rules/lint032"
)

// Analyzers is the list of all relint analyzers.
var Analyzers = []*analysis.Analyzer{
	fmt001.Analyzer,
	fmt002.Analyzer,
	fmt003.Analyzer,
	fmt004.Analyzer,
	fmt005.Analyzer,
	fmt006.Analyzer,
	fmtfix.Analyzer,
	lint001.Analyzer,
	lint002.Analyzer,
	lint003.Analyzer,
	lint004.Analyzer,
	lint005.Analyzer,
	lint006.Analyzer,
	lint007.Analyzer,
	lint008.Analyzer,
	lint009.Analyzer,
	lint010.Analyzer,
	lint011.Analyzer,
	lint012.Analyzer,
	lint013.Analyzer,
	lint014.Analyzer,
	lint015.Analyzer,
	lint016.Analyzer,
	lint017.Analyzer,
	lint018.Analyzer,
	lint019.Analyzer,
	lint020.Analyzer,
	lint021.Analyzer,
	lint022.Analyzer,
	lint023.Analyzer,
	lint024.Analyzer,
	lint025.Analyzer,
	lint026.Analyzer,
	lint027.Analyzer,
	lint028.Analyzer,
	lint029.Analyzer,
	lint030.Analyzer,
	lint031.Analyzer,
	lint032.Analyzer,
}

func init() {
	for i, analyzer := range Analyzers {
		Analyzers[i] = wrapSkipGeneratedFiles(analyzer)
	}
}

// wrapSkipGeneratedFiles ensures all analyzers ignore generated files
// (for example testmain wrappers created by `go test` in GOCACHE).
func wrapSkipGeneratedFiles(analyzer *analysis.Analyzer) *analysis.Analyzer {
	wrapped := *analyzer
	originalRun := analyzer.Run

	wrapped.Run = func(pass *analysis.Pass) (interface{}, error) {
		if len(pass.Files) == 0 {
			return originalRun(pass)
		}

		filteredFiles := make([]*ast.File, 0, len(pass.Files))
		generatedByFilename := make(map[string]bool)
		for _, f := range pass.Files {
			filename := pass.Fset.File(f.Pos()).Name()
			if ast.IsGenerated(f) {
				generatedByFilename[filename] = true
				continue
			}
			filteredFiles = append(filteredFiles, f)
		}

		// Entirely generated package (e.g. synthetic testmain): nothing to lint.
		if len(filteredFiles) == 0 {
			return nil, nil
		}

		originalFiles := pass.Files
		originalReport := pass.Report
		pass.Files = filteredFiles
		pass.Report = func(d analysis.Diagnostic) {
			if d.Pos != token.NoPos {
				if file := pass.Fset.File(d.Pos); file != nil && generatedByFilename[file.Name()] {
					return
				}
			}
			originalReport(d)
		}
		defer func() {
			pass.Files = originalFiles
			pass.Report = originalReport
		}()

		return originalRun(pass)
	}

	return &wrapped
}
