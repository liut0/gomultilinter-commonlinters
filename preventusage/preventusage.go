package main

import (
	"context"
	"go/ast"
	"go/types"

	"github.com/liut0/gomultilinter/api"
	"golang.org/x/tools/go/loader"
)

type preventUsageLinterFactory struct{}

type preventUsageLinter struct {
	Packages map[string]string `json:"packages"`
	Funcs    map[string]string `json:"funcs"`
}

var LinterFactory api.LinterFactory = &preventUsageLinterFactory{}

func (l *preventUsageLinterFactory) NewLinterConfig() api.LinterConfig {
	return &preventUsageLinter{
		Packages: map[string]string{},
		Funcs:    map[string]string{},
	}
}

func (l *preventUsageLinter) NewLinter() (api.Linter, error) {
	return l, nil
}

func (*preventUsageLinter) Name() string {
	return "preventusage"
}

func (l *preventUsageLinter) LintFile(ctx context.Context, file *api.File, reporter api.IssueReporter) error {
	for _, pkgImport := range file.ASTFile.Imports {
		pkgPath := pkgImport.Path.Value[1 : len(pkgImport.Path.Value)-1]
		if msg, ok := l.Packages[pkgPath]; ok {
			reporter.Report(&api.Issue{
				Position: file.FSet.Position(pkgImport.Pos()),
				Category: "prevent-usage-pkg",
				Message:  msg,
				Severity: api.SeverityWarning,
			})
		}
	}

	ast.Inspect(file.ASTFile, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		fnName, ok := FullFuncName(call, file.PkgInfo)
		if !ok {
			return true
		}

		if msg, ok := l.Funcs[fnName]; ok {
			reporter.Report(&api.Issue{
				Position: file.FSet.Position(call.Pos()),
				Category: "prevent-usage-func",
				Message:  msg,
				Severity: api.SeverityWarning,
			})
		}

		return false
	})

	return nil
}

func FullFuncName(call *ast.CallExpr, pkgInfo *loader.PackageInfo) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	fn, ok := pkgInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return "", false
	}

	return fn.FullName(), true
}
