// Package exitcheck defines an analyzer that reports direct calls to os.Exit
// in the main function of the main package.
//
// Using os.Exit bypasses deferred functions and prevents proper cleanup.
// Instead, return from main or use a run() pattern that returns an error code.
package exitcheck

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports direct calls to os.Exit in the main function of package main.
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "reports direct calls to os.Exit in the main function of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		pos := pass.Fset.Position(file.Pos())
		if pos.Filename == "" || strings.Contains(pos.Filename, "go-build") {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if sel.Sel.Name != "Exit" {
					return true
				}

				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				obj, ok := pass.TypesInfo.Uses[ident]
				if !ok {
					return true
				}

				pkgName, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				if pkgName.Imported().Path() == "os" {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main function of package main is not allowed")
				}

				return true
			})
		}
	}

	return nil, nil
}
