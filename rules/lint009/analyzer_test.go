package lint009_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/daiteo/relint/rules/lint009"
)

func TestAnalyzer(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "..", "..", "example")

	t.Run("default_exceptions", func(t *testing.T) {
		lint009.Analyzer.Flags.Set("exceptions", "types,handlertypes,params")
		analysistest.Run(t, testdata, lint009.Analyzer, "lint009", "lint009models", "lint009typesok", "lint009statusok", "lint009handlertypesok")
	})

	t.Run("types_not_excepted", func(t *testing.T) {
		lint009.Analyzer.Flags.Set("exceptions", "")
		analysistest.Run(t, testdata, lint009.Analyzer, "lint009typesbad")
	})

	t.Cleanup(func() { lint009.Analyzer.Flags.Set("exceptions", "types,handlertypes,params") })
}
