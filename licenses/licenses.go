package main

import (
	"context"
	"fmt"
	"go/build"
	"go/token"
	"path/filepath"
	"regexp"

	"github.com/liut0/gomultilinter-commonlinters/internal/utils"
	"github.com/liut0/gomultilinter/api"
	"github.com/ryanuber/go-license"
)

const categoryLicense = "license"

var (
	defaultWhitelistedLicenses = []string{
		license.LicenseMIT,
		license.LicenseISC,
		license.LicenseNewBSD,
		license.LicenseFreeBSD,
		license.LicenseApache20,
		license.LicenseCDDL10,
		license.LicenseEPL10,
		license.LicenseUnlicense,
	}

	defaultIncludedPackages = []string{
		`github\.com\/.+$`,
		`gopkg\.in\/.+$`,
	}
)

type licensesLinterFactory struct {
}

type licensesLinterConfig struct {
	IncludedPackages          []string `json:"included_packages"`
	WhitelistedLicenses       []string `json:"whitelisted_licenses"`
	FailWhenNoLicensePresent  bool     `json:"fail_when_no_license_present"`
	FailOnUnrecognizedLicense bool     `json:"fail_when_on_unrecognized_license"`
}

type licensesLinter struct {
	includedPackages          []*regexp.Regexp
	whitelistedLicenses       map[string]bool
	failWhenNoLicensePresent  bool
	failOnUnrecognizedLicense bool

	donePackages map[string]bool
	licenseCache map[string]string
}

var LinterFactory api.LinterFactory = &licensesLinterFactory{}

func (l *licensesLinterFactory) NewLinterConfig() api.LinterConfig {
	return &licensesLinterConfig{
		IncludedPackages:    defaultIncludedPackages,
		WhitelistedLicenses: defaultWhitelistedLicenses,
	}
}

func (cfg *licensesLinterConfig) NewLinter() (api.Linter, error) {
	linter := &licensesLinter{
		whitelistedLicenses:       utils.StringArrayToMap(cfg.WhitelistedLicenses),
		failWhenNoLicensePresent:  cfg.FailWhenNoLicensePresent,
		failOnUnrecognizedLicense: cfg.FailOnUnrecognizedLicense,
		donePackages:              make(map[string]bool),
		licenseCache:              make(map[string]string),
	}

	includedPkgs := make([]*regexp.Regexp, 0, len(cfg.IncludedPackages))
	for _, r := range cfg.IncludedPackages {
		includedPkgs = append(includedPkgs, regexp.MustCompile(r))
	}
	linter.includedPackages = includedPkgs

	return linter, nil
}

func (*licensesLinter) Name() string {
	return "licenses"
}

func (l *licensesLinter) LintPackage(ctx context.Context, pkg *api.Package, reporter api.IssueReporter) error {
	pkgPos := utils.GetPkgPos(pkg)
	for _, pkgImport := range pkg.PkgInfo.Pkg.Imports() {
		if iss, err := l.lintImport(pkgImport.Path(), pkgPos); err != nil {
			return err
		} else if iss != nil {
			reporter.Report(iss)
		}
	}

	return nil
}

func (l *licensesLinter) lintImport(pkgImportPath string, pos *token.Position) (*api.Issue, error) {
	if l.donePackages[pkgImportPath] || !matchesAny(pkgImportPath, l.includedPackages) {
		return nil, nil
	}

	importedPkg, err := build.Import(pkgImportPath, ".", build.FindOnly)
	if err != nil {
		return nil, err
	}

	lic, err := l.getLicense(importedPkg.Dir)
	switch err {
	case license.ErrNoLicenseFile:
		if l.failWhenNoLicensePresent {
			return &api.Issue{
				Position: *pos,
				Message:  fmt.Sprintf("no license found for pkg %v", pkgImportPath),
				Severity: api.SeverityWarning,
				Category: categoryLicense,
			}, nil
		}
		return nil, nil
	case license.ErrMultipleLicenses, license.ErrUnrecognizedLicense:
		if l.failWhenNoLicensePresent {
			return &api.Issue{
				Position: *pos,
				Message:  fmt.Sprintf("multiple or unrecognized license file found for pkg %v", pkgImportPath),
				Severity: api.SeverityWarning,
				Category: categoryLicense,
			}, nil
		}
		return nil, nil
	default:
		if err != nil {
			return nil, err
		}
	}

	l.donePackages[pkgImportPath] = true

	if !l.whitelistedLicenses[lic] {
		return &api.Issue{
			Position: *pos,
			Message:  fmt.Sprintf("%v license found for pkg %v, which is not whitelisted", lic, pkgImportPath),
			Severity: api.SeverityWarning,
			Category: categoryLicense,
		}, nil
	}

	return nil, nil
}

func (l *licensesLinter) getLicense(pkgPath string) (string, error) {
	if license, ok := l.licenseCache[pkgPath]; ok {
		return license, nil
	}

	lic, err := license.NewFromDir(pkgPath)

	if err != nil {
		if err == license.ErrNoLicenseFile {
			parentPath := filepath.Dir(pkgPath)
			if parentPath != pkgPath {
				return l.getLicense(parentPath)
			}
		}

		return "", err
	}

	l.licenseCache[pkgPath] = lic.Type
	return lic.Type, nil
}

func matchesAny(s string, regExps []*regexp.Regexp) bool {
	for _, rgx := range regExps {
		if rgx.MatchString(s) {
			return true
		}
	}
	return false
}
