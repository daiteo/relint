package fmtfix_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/fmtfix"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.RunWithSuggestedFixes(t, testdata, fmtfix.Analyzer, "fmtfix", "fmtfixtypespacing", "fmtfixcomments", "fmtfixreturnspacing")
}
