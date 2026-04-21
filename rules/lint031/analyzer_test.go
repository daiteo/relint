package lint031_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint031"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.Run(t, testdata, lint031.Analyzer, "lint031", "lint031ok")
}
