# Go test cover wrapper

## Features
- `go test` / `go tool cover` wrapper
- Multiple packages support (`go test` doesn't support coverage profile with multiple packages)
- Merged HTML output

## Install
`go get github.com/pierrre/gotestcover`

## Usage
`gotestcover <options> <packages>`

Run on multiple package with:
- `package1 package2`
- `package/...`

Some `go test / build` flags are available.
