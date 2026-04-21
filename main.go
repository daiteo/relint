package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/daiteo/relint/all"
)

func main() {
	showVersion, args, err := stripVersionArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	if showVersion {
		fmt.Println(binaryVersion())
		return
	}

	args, err = preprocessArgs(args, all.Analyzers)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	os.Args = args
	multichecker.Main(all.Analyzers...)
}

func preprocessArgs(args []string, analyzers []*analysis.Analyzer) ([]string, error) {
	if len(args) == 0 {
		return args, nil
	}

	onlyFmtfix := false
	filtered := make([]string, 0, len(args))
	filtered = append(filtered, args[0])

	for _, arg := range args[1:] {
		handled, enabled, err := parseOnlyFmtfixArg(arg)
		if err != nil {
			return nil, err
		}
		if handled {
			onlyFmtfix = enabled
			continue
		}
		filtered = append(filtered, arg)
	}

	if !onlyFmtfix {
		return filtered, nil
	}

	// Keep fmtfix enabled and disable all other analyzers.
	injected := make([]string, 0, len(analyzers)+1)
	injected = append(injected, "-fmtfix=true")
	for _, analyzer := range analyzers {
		if analyzer.Name == "fmtfix" {
			continue
		}
		injected = append(injected, "-"+analyzer.Name+"=false")
	}

	// Analyzer flags must appear before package patterns (e.g. ./...).
	out := make([]string, 0, len(filtered)+len(injected))
	out = append(out, filtered[0])

	inserted := false
	for _, arg := range filtered[1:] {
		if !inserted && (arg == "--" || !strings.HasPrefix(arg, "-")) {
			out = append(out, injected...)
			inserted = true
		}
		out = append(out, arg)
	}
	if !inserted {
		out = append(out, injected...)
	}

	return out, nil
}

func parseOnlyFmtfixArg(arg string) (handled bool, enabled bool, err error) {
	if arg == "-only-fmtfix" || arg == "--only-fmtfix" {
		return true, true, nil
	}

	const shortPrefix = "-only-fmtfix="
	const longPrefix = "--only-fmtfix="
	if strings.HasPrefix(arg, shortPrefix) || strings.HasPrefix(arg, longPrefix) {
		value := strings.TrimPrefix(strings.TrimPrefix(arg, shortPrefix), longPrefix)
		b, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return false, false, fmt.Errorf("invalid value for -only-fmtfix: %q", value)
		}
		return true, b, nil
	}

	return false, false, nil
}

func stripVersionArgs(args []string) (showVersion bool, filtered []string, err error) {
	if len(args) == 0 {
		return false, args, nil
	}

	showVersion = false
	filtered = make([]string, 0, len(args))
	filtered = append(filtered, args[0])

	for _, arg := range args[1:] {
		handled, enabled, parseErr := parseVersionArg(arg)
		if parseErr != nil {
			return false, nil, parseErr
		}
		if handled {
			showVersion = enabled
			continue
		}
		filtered = append(filtered, arg)
	}

	return showVersion, filtered, nil
}

func parseVersionArg(arg string) (handled bool, enabled bool, err error) {
	if arg == "-version" || arg == "--version" {
		return true, true, nil
	}

	const shortPrefix = "-version="
	const longPrefix = "--version="
	if strings.HasPrefix(arg, shortPrefix) || strings.HasPrefix(arg, longPrefix) {
		value := strings.TrimPrefix(strings.TrimPrefix(arg, shortPrefix), longPrefix)
		b, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return false, false, fmt.Errorf("invalid value for -version: %q", value)
		}
		return true, b, nil
	}

	return false, false, nil
}
