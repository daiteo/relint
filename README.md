# relint

`relint` is a custom Go multichecker that runs this repository's analyzers.
All analyzers skip generated Go files (`// Code generated ... DO NOT EDIT.`).

## Requirements

- Go 1.26+

## Build

```bash
go build -o relint .
```

## Run

```bash
go run . ./...
```

Or run the built binary:

```bash
./relint ./...
```

## Rules

### Formatter rules

- `FMT-001` Declaration merging (`type`/`const`/`var`)
- `FMT-002` File declaration order
- `FMT-003` Function body spacing
- `FMT-004` Interface method spacing
- `FMT-005` Type block spec spacing
- `FMT-006` Blank line after blocks ending with `return`
- `FMTFIX` Auto-fix suggestions

### Linter rules

- `LINT-001` Log key casing
- `LINT-002` Log message casing
- `LINT-003` Log key dot notation for grouped keys
- `LINT-004` Context as first parameter
- `LINT-005` Excessive function parameters
- `LINT-006` Excessive return values
- `LINT-007` Enum value prefix
- `LINT-008` Package name underscore
- `LINT-009` Package name plural (generic detection, exceptions supported)
- `LINT-010` Interface location (`core/*` packages excluded)
- `LINT-011` Service/store/worker interface suffix
- `LINT-012` Store function return types
- `LINT-013` Store struct interface assertion
- `LINT-014` Service struct interface assertion
- `LINT-015` One exported layer method per store/service/handler file
- `LINT-016` `Inject*`/`inject*` middleware file naming in `*handler` packages
- `LINT-017` `Require*`/`require*` middleware file naming in `*handler` packages
- `LINT-018` Middleware naming outside `*handler` packages
- `LINT-019` `FxModule` must be in `store.go` / `service.go` / `handler.go`
- `LINT-020` `Err*` location in `types/errors.go`
- `LINT-021` Store `RecordNotFound` sentinel return
- `LINT-022` Handler route method file naming
- `LINT-023` Route `Input`/`Output` type location in module-scoped `*handler` packages
- `LINT-024` Shared body helper type naming in `*handler` packages
- `LINT-025` Handler struct file location
- `LINT-026` Body-only helper struct naming
- `LINT-027` No `json` tags in `model` package structs
- `LINT-028` Exported model fields require `gorm` tag in package `model`
- `LINT-029` Model relation fields must be `*Type` or `[]*Type` in package `model`
- `LINT-030` Protected roots (default `core`) must not import sibling roots
- `LINT-031` `httpapi` path params must be `lowerCamelCase`
- `LINT-032` layer constructors must expose a single `New`

See [spec.md](./spec.md) for full rule definitions.

## Configurable rules

Rule-specific options are exposed as analyzer flags:

- `-lint003.dot-notation` (default: empty)
- `-lint007.exceptions` (default: `environment.Environment`)
- `-lint008.excluded-suffixes` (default: `_test`)
- `-lint009.exceptions` (default: `types,handlertypes,params`)
- `-lint030.roots` (default: `core`)

Examples:

```bash
./relint \
  -lint003.dot-notation="error=error.message,userId=user.id" \
  -lint007.exceptions="environment.Environment,foo.Status" \
  -lint008.excluded-suffixes="_test,_v2" \
  -lint009.exceptions="types,models" \
  -lint030.roots="core,shared" \
  ./...
```

To inspect all flags:

```bash
./relint -help
./relint -flags
```

Run only `fmtfix` with a convenience flag:

```bash
./relint -only-fmtfix ./...
```

## Excluding rules

### By CLI

Disable one or more analyzers using boolean analyzer flags:

```bash
./relint -lint016=false -lint017=false ./...
```

Run only specific analyzers by disabling others as needed (same mechanism).

### By config file

`relint` itself does not currently have a dedicated config file.  
For file-based exclusions, run it via `golangci-lint` and use `.golangci.yml`:

```yaml
linters:
  enable:
    - relint

issues:
  exclude-rules:
    - linters:
        - relint
      text: "LINT-016"
    - linters:
        - relint
      text: "LINT-017"
      path: ".*_test\\.go"
```

## Test

```bash
go test ./...
```
