package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/daiteo/relint/all"
)

// New returns the list of analyzers for use as a golangci-lint plugin.
func New(conf any) ([]*analysis.Analyzer, error) {
	return all.Analyzers, nil
}

func main() {}
