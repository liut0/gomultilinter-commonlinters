# my custom linters for gomultilinter

To be used in combination with [gomultilinter](https://github.com/liut0/gomultilinter)

## dep

Runs `dep status` and checks the exit code. Literally checks whether `Gopkg.lock` is in sync.
No configs available.

### Example

```yaml
- package: 'github.com/liut0/gomultilinter-commonlinters/dep'
```

## licenses

Checks whether imported packages' licenses are whitelisted

### Config

- `included_packages`: List of package regular expressions which are validated (default `github.com`, `gopkg.in`)
- `whitelisted_licenses`: List of whitelisted licenses (default `MIT;ISC;NewBSD;FreeBSD;Apache2.0;CDDL10;EPL10;Unlicense`)
- `fail_when_no_license_present`: Reports a warning when a package without a license is imported (default `false`)
- `fail_when_on_unrecognized_license`: Reports a warning when a package whith an unrecognized license or multiple licesnes is found (default `false`)

### Example

```yaml
- package: 'github.com/liut0/gomultilinter-commonlinters/licenses'
  config:
    fail_when_no_license_present: true
    fail_when_on_unrecognized_license: true
```

## preventusage

Lints several go packages/functions are not used.

### Config

- `packages`: Map of packages which should be prevented and the corresponding messages
- `funcs`: Map of function names which should be prevented and the corresponding messages

#### Example

```yaml
- package: 'github.com/liut0/gomultilinter-commonlinters/preventusage'
  config:
    packages:
      'golang.org/x/net/context': 'use "context" package instead'
    funcs:
      '(*log.Logger).Println': 'use company internal log package'
```