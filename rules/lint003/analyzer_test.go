package lint003_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint003"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")

	lint003.Analyzer.Flags.Set("dot-notation", "error=error.message,userId=user.id,userID=user.id,sessionId=session.id,sessionID=session.id")
	t.Cleanup(func() { lint003.Analyzer.Flags.Set("dot-notation", "") })

	analysistest.Run(t, testdata, lint003.Analyzer, "lint003")
}
