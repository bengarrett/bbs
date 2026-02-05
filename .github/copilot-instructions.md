# Copilot Instructions for BBS Package

## Overview
This is a Go module that converts legacy Bulletin Board System (BBS) color codes from the 1980s-1990s into HTML. It supports six different BBS systems: PCBoard, Celerity, Renegade, Telegard, WWIV, and Wildcat, each with their own unique color code syntax.

## Build, Test, and Lint Commands

**Using Task runner** (https://taskfile.dev/):
- `task lint` - Run gofumpt formatter and golangci-lint with custom config
- `task test` - Run test suite (enforces single execution per test with `-count 1`)
- `task testr` - Run tests with race detection enabled
- `task nil` - Run nilaway static analysis for nil dereference detection
- `task pkg-update` / `task update` - Update all dependencies
- `task pkg-patch` / `task patch` - Update patch versions only
- `task doc` - Start pkgsite documentation server on localhost:8090

**Using Go directly**:
- `go test -count 1 ./...` - Run all tests
- `go test -count 1 -run TestBBS ./...` - Run specific test pattern
- `golangci-lint run -c .golangci.yaml` - Lint with configured rules
- `gofumpt -l -w .` - Format code

## Architecture & Key Patterns

### Package Structure
- **Main package** (`bbs.go`, `bbs_test.go`): Core API for BBS code detection and conversion
- **`internal/split`**: Internal utilities for string splitting/parsing (used by main package)
- **Example tests** (`example*.go`): Go testable examples showing package usage

### Core Types & Functions

**BBS Type** (enum-like):
```go
type BBS int
```
Enumerated BBS system types: `Celerity`, `PCBoard`, `Renegade`, `Telegard`, `WWIVHash`, `WWIVHeart`, `Wildcat`, `ANSI`

**Main API Functions**:
- `Find(r io.Reader) BBS` - Detect which BBS system encoded the data
- `HTML(buf *bytes.Buffer, src io.Reader) (BBS, error)` - Convert BBS codes to HTML and detect system type
- `Fields(src io.Reader) ([]string, BBS, error)` - Extract fields and detect BBS type
- `(b BBS) CSS(buf *bytes.Buffer) error` - Generate CSS for detected BBS system
- `(b BBS) HTML(buf *bytes.Buffer, src []byte) error` - Convert known BBS type to HTML
- `(b BBS) Remove(buf *bytes.Buffer, src ...byte) error` - Strip all BBS codes

**System-Specific Functions**:
- `IsCelerity()`, `IsRenegade()`, `IsWildcat()`, `IsPCBoard()`, `IsWWIVHash()`, `IsWWIVHeart()`, `IsTelegard()` - Detection functions
- `CelerityHTML()`, `RenegadeHTML()`, etc. - System-specific converters

### Test Conventions
- Tests use table-driven patterns with named structs
- Example tests demonstrate real usage in `example_test.go`, `examplebbs_test.go`, `examplefields_test.go`
- Tests use constants for escape sequences: `const ansiEsc = "\x1B\x5B"`
- Single test execution: `go test -count 1` prevents caching effects

### Code Quality Standards
- **Linter**: Custom golangci-lint config in `.golangci.yaml`
  - Strict mode (default: all linters enabled)
  - Complexity limit: 15
  - Several linters disabled: depguard, nlreturn, noinlineerr, paralleltest, wsl
  - Test files exempt from varnamelen, lll, exhaustruct
- **Formatting**: gofumpt, goimports, gci, gofmt (applied in order)
- **Static Analysis**: nilaway tool for nil dereference detection
- **Coverage Target**: 93%

## Key Conventions
- **Color Code Syntax**: Each BBS system uses different symbols:
  - PCBoard/Wildcat: `@X` syntax (hex color pairs)
  - Celerity: `|` pipe syntax
  - Renegade: Pipe Bar Color Codes
  - Telegard: Two-digit pipe colors
  - WWIV: `|` pipes (10-color or 2-digit), or `♥` heart codes
- **Byte Handling**: Functions accept `...byte` variadic args for flexibility in input types
- **Error Handling**: Functions return `error` type; check for nil before proceeding
- **Constants**: Validation constants use iota for enumeration (e.g., BBS system types)
