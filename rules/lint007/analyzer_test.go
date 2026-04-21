package lint007_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint007"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.Run(t, testdata, lint007.Analyzer, "lint007", "environment")
}

func TestAnalyzer_WithConfiguredExceptions(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")

	lint007.Analyzer.Flags.Set("exceptions", "environment.Environment,lint007exceptions.Status")
	t.Cleanup(func() { lint007.Analyzer.Flags.Set("exceptions", "environment.Environment") })

	analysistest.Run(t, testdata, lint007.Analyzer, "lint007exceptions")
}
