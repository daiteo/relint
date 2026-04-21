package lint019_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint019"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.Run(
		t,
		testdata,
		lint019.Analyzer,
		"lint019",
		"lint019nonlayer",
		"lint019handleroptional",
		"lint019wrongstore",
		"lint019wrongservice",
		"lint019wronghandler",
		"lint019okstore",
		"lint019okservice",
		"lint019okhandler",
	)
}
