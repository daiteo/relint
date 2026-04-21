package lint018_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint018"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.Run(t, testdata, lint018.Analyzer, "lint018", "lint018modulehandler")
}
