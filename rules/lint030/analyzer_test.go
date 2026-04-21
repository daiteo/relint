package lint030_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint030"
)

func TestAnalyzer(t *testing.T) {
	oldRoots := lint030.Analyzer.Flags.Lookup("roots").Value.String()
	if err := lint030.Analyzer.Flags.Set("roots", "lint030core"); err != nil {
		t.Fatalf("failed to set lint030 roots flag: %v", err)
	}
	t.Cleanup(func() {
		_ = lint030.Analyzer.Flags.Set("roots", oldRoots)
	})

	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")
	analysistest.Run(t, testdata, lint030.Analyzer, "lint030core", "lint030smarthubserver/handler")
}
