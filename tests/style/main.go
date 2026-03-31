// Package main implements a linter for the backend acceptance test DSL.
// It enforces two rules:
//
//  1. dsl-shape: every exported function in tests/acceptance/backend/dsl/ that
//     returns Step must have the signature func(params ...string) Step.
//
//  2. dsl-import: every _test.go file in tests/acceptance/backend/ (direct
//     children only, not under dsl/) must dot-import the DSL package.
//
// Run from the repo root:
//
//	go run ./tests/style
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type target struct {
	dslImportPath string
	dslDir        string
	testDir       string
}

var targets = []target{
	{
		dslImportPath: "github.com/jmsargent/kanban/tests/acceptance/dsl",
		dslDir:        "tests/acceptance/dsl",
		testDir:       "tests/acceptance",
	},
	{
		dslImportPath: "github.com/jmsargent/kanban/tests/acceptance/backend/dsl",
		dslDir:        "tests/acceptance/backend/dsl",
		testDir:       "tests/acceptance/backend",
	},
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	failures := 0
	for _, t := range targets {
		failures += checkDSLShape(filepath.Join(root, t.dslDir))
		failures += checkDSLImport(filepath.Join(root, t.testDir), t.dslImportPath)
	}

	if failures > 0 {
		fmt.Fprintf(os.Stderr, "\ndsl-check: %d violation(s)\n", failures)
		os.Exit(1)
	}
	fmt.Println("dsl-check: OK")
}

// checkDSLShape reports every exported function in dir that returns Step but
// does not have the signature func(params ...string) Step.
func checkDSLShape(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dsl-shape: read %s: %v\n", dir, err)
		return 1
	}
	fset := token.NewFileSet()
	failures := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dsl-shape: parse %s: %v\n", path, err)
			failures++
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || !ast.IsExported(fn.Name.Name) {
				continue
			}
			if !returnsStep(fn) {
				continue
			}
			if !hasVariadicStringParam(fn) {
				pos := fset.Position(fn.Pos())
				fmt.Fprintf(os.Stderr,
					"dsl-shape: %s: %s must have signature func(params ...string) Step\n",
					pos, fn.Name.Name)
				failures++
			}
		}
	}
	return failures
}

func returnsStep(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	ident, ok := fn.Type.Results.List[0].Type.(*ast.Ident)
	return ok && ident.Name == "Step"
}

func hasVariadicStringParam(fn *ast.FuncDecl) bool {
	params := fn.Type.Params
	if params == nil || len(params.List) != 1 {
		return false
	}
	ellipsis, ok := params.List[0].Type.(*ast.Ellipsis)
	if !ok {
		return false
	}
	elt, ok := ellipsis.Elt.(*ast.Ident)
	return ok && elt.Name == "string"
}

// checkDSLImport reports every _test.go in dir (direct children only) that
// does not dot-import the DSL package at importPath.
func checkDSLImport(dir, importPath string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dsl-import: read %s: %v\n", dir, err)
		return 1
	}
	fset := token.NewFileSet()
	failures := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dsl-import: parse %s: %v\n", path, err)
			failures++
			continue
		}
		if !hasDotImport(f, importPath) {
			fmt.Fprintf(os.Stderr,
				"dsl-import: %s: must dot-import DSL — replace with: import . %q\n",
				path, importPath)
			failures++
		}
	}
	return failures
}

func hasDotImport(f *ast.File, path string) bool {
	for _, imp := range f.Imports {
		impPath := strings.Trim(imp.Path.Value, `"`)
		if impPath != path {
			continue
		}
		return imp.Name != nil && imp.Name.Name == "."
	}
	return false
}

// repoRoot walks up from the working directory until it finds go.mod.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found — run from within the repo")
		}
		dir = parent
	}
}