package lint026

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"github.com/daiteo/relint/analysisutil"
)

var Analyzer = &analysis.Analyzer{
	Name:     "lint026",
	Doc:      "LINT-026: body-only helper structs in packages ending with handler must use body prefix and matching Input/Output suffix",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if !analysisutil.IsHandlerPackage(pass.Pkg.Name()) {
		return nil, nil
	}

	usage := analysisutil.AnalyzeBodyTypeUsage(pass)

	for helperName := range usage.BodyOnlyStructs {
		bodyStructs := usage.UsedByBodyStruct[helperName]
		for bodyStructName := range bodyStructs {
			meta, ok := usage.BodyStructs[bodyStructName]
			if !ok {
				continue
			}
			if !strings.HasPrefix(helperName, meta.Prefix) || !strings.HasSuffix(helperName, meta.Suffix) {
				pass.Reportf(
					usage.DeclPos[helperName],
					"LINT-026: body-only struct %q must be prefixed with %q and suffixed with %q",
					helperName,
					meta.Prefix,
					meta.Suffix,
				)
				break
			}
		}
	}

	return nil, nil
}
