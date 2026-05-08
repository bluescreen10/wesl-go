# wesl-go

<p align="center">
    <img src="logo/logo.png" width="300" alt="logo" />
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/bluescreen10/wesl-go.svg)](https://pkg.go.dev/github.com/bluescreen10/wesl-go)
[![Tests](https://github.com/bluescreen10/wesl-go/actions/workflows/go.yml/badge.svg)](https://github.com/bluescreen10/wesl-go/actions)
[![codecov](https://codecov.io/gh/bluescreen10/wesl-go/branch/main/graph/badge.svg)](https://codecov.io/gh/bluescreen10/wesl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/bluescreen10/wesl-go)](https://goreportcard.com/report/github.com/bluescreen10/wesl-go)
![Go Version](https://img.shields.io/github/go-mod/go-version/bluescreen10/wesl-go)

A Go implementation of the [WESL](https://github.com/wgsl-tooling-wg/wesl-spec) compiler — a superset of WGSL that adds a module system with cross-file imports and compile-time `@if` conditionals. It takes one or more WESL source files and produces a single, flat WGSL output suitable for use with the WebGPU API.

## Usage

```go
import "github.com/bluescreen10/wesl-go"

// Create a compiler instance.
c := wesl.New()

// Register source files. Paths are stored relative to the FS root so that
// package:: imports resolve correctly.
c.ParseFS(os.DirFS("./shaders"), "**/*.wesl")

// Compile an entry-point file, optionally passing compile-time feature flags.
wgsl, err := c.Compile("main.wesl", map[string]bool{"MY_FEATURE": true})
if err != nil {
    log.Fatal(err)
}
fmt.Println(wgsl)
```


## Credits

Test cases are adapted from the [WESL Test Suite](https://github.com/wgsl-tooling-wg/wesl-testsuite/) maintained by the WebGPU Shading Language tooling working group.
