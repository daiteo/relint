package lint008_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint008"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.Run(t, testdata, lint008.Analyzer, "lint008", "lint008testsuffix")
}

func TestAnalyzer_WithConfiguredExclusions(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")

	lint008.Analyzer.Flags.Set("excluded-suffixes", "_test,_v2")
	t.Cleanup(func() { lint008.Analyzer.Flags.Set("excluded-suffixes", "_test") })

	analysistest.Run(t, testdata, lint008.Analyzer, "lint008customexcluded", "lint008customnotexcluded")
}
