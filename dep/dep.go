package main

import (
	"bytes"
	"context"
	"fmt"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/liut0/gomultilinter-commonlinters/internal/utils"
	"github.com/liut0/gomultilinter/api"
)

var (
	rgxNewLine = regexp.MustCompile(`\r?\n`)
)

type depLinterFactory struct{}

type depLinter struct {
}

var LinterFactory api.LinterFactory = &depLinterFactory{}

func (l *depLinterFactory) NewLinterConfig() api.LinterConfig {
	return &depLinter{}
}

func (l *depLinter) NewLinter() (api.Linter, error) {
	return l, nil
}

func (*depLinter) Name() string {
	return "dep"
}

func (l *depLinter) LintPackage(ctx context.Context, pkg *api.Package, reporter api.IssueReporter) error {
	pos := utils.GetPkgPos(pkg)
	if pos == nil {
		reporter.Debug("could not determine path of package, maybe no go file is present", "pkg", pkg.PkgInfo.Pkg.Name())
		return nil
	}

	if errStr, ok := l.isDepOk(pos); !ok {
		reporter.Report(&api.Issue{
			Category: "dep-lock-file-not-synced",
			Severity: api.SeverityWarning,
			Message:  fmt.Sprintf("dep seems to be out of sync [%v]", errStr),
			Position: *pos,
		})
	}

	return nil
}

// isDepOk checks wehter dep status is ok. returns the error string of
// dep status and wether the check failed
// the check fails only if a Gopkg.toml file is found in the directory of
// the filename of pos
func (l *depLinter) isDepOk(pos *token.Position) (string, bool) {
	pkgPath := filepath.Dir(pos.Filename)

	if pkgPath == "" {
		return "", true
	}

	if _, err := os.Stat(filepath.Join(pkgPath, "Gopkg.toml")); err != nil {
		return "", true
	}

	cmd := exec.Command("dep", "status") // nolint: gas
	cmd.Dir = pkgPath

	errOut := &bytes.Buffer{}
	cmd.Stderr = errOut

	err := cmd.Run()
	if err == nil {
		return "", true
	}

	return rgxNewLine.ReplaceAllString(errOut.String(), " "), false
}
