# sigmafmt
`sigmafmt` is an opinionated formatter/linter for [Sigma rules](https://github.com/Neo23x0/sigma). The aim is to make Sigma rules as readable as possible by keeping them as consistently formatted as possible.

⚠️ The rules enforced by `sigmafmt` are not set in stone (this is primarily a personal project to keep my own threathunting rules tidy). The "style" enforced by `sigmafmt` is likely to change drastically between releases.

🐛 Suggestions for new rules or improvements to existing rules are welcomed!

## Usage
The `sigmafmt` interface is heavily inspired by `gofmt`

If you're using it as an auto-formatter for your code, use `sigmafmt -w <path to file(s)>`
to automatically re-write files (`sigmafmt` format rules recursively if passed a directory).

To use as a CI check, use `sigmafmt -l <path to rules>` to list all files whose formatting needs fixing.

```
usage: sigmafmt [flags] [path ...]
  -d    display diffs instead of rewriting files
  -l    list files whose formatting differs from sigmafmt's
  -v    write extra info (e.g. reasons) to stdout
  -w    write result to (source) file instead of stdout
```

## Installation
Via Homebrew:
```
brew install bradleyjkemp/formulae/sigmafmt
```

From source:
```
go install github.com/bradleyjkemp/sigmafmt
```