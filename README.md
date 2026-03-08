# gotcha

**gotcha** is a test automation tool for Go developers. It watches for file changes and automatically runs `go test`, supporting TDD and fast feedback.

## Features

- Automatically detects changes in `.go` files and runs `go test`
- `--fast` mode: only test packages with uncommitted changes
- `--focus` mode: show only failed test output
- `--summary` mode: show pass/fail/skip counts
- `gotcha watch` for continuous test running with keyboard control (`r` to rerun, `q` to quit)
- `.gotcha.yaml` for fine-grained test control

## Install

```sh
go install github.com/mickamy/gotcha@latest
```

or build from source:

```sh
git clone https://github.com/mickamy/gotcha.git
cd gotcha
make install
```

## Usage

### Initialize

```sh
gotcha init
```

Generates `.gotcha.yaml` with sensible defaults:

```yaml
include:
  - "./..."
exclude:
  - "vendor/"
  - "mocks/"
args:
  - "-v"
```

### Run tests once

```sh
# Run all tests
gotcha run

# Only test packages with uncommitted changes
gotcha run --fast

# Show only failed test output
gotcha run --focus

# Show pass/fail/skip summary
gotcha run --summary
```

### Watch for changes

```sh
# Watch and rerun on file changes
gotcha watch

# Combine with flags
gotcha watch --fast --focus
```

In watch mode:
- Press `r` to manually rerun tests
- Press `q` to quit

### Version

```sh
gotcha version
```

## Configuration

`.gotcha.yaml` fields:

| Field     | Description                                        |
|-----------|----------------------------------------------------|
| `include` | Target packages (passed to `go list`)              |
| `exclude` | Directories to ignore (matched per path segment)   |
| `args`    | Arguments passed to `go test` (e.g. `-v`, `-race`) |

## Comparison

| Tool         | How gotcha differs                                              |
|--------------|-----------------------------------------------------------------|
| `watchexec`  | Go-native, built specifically for `go test`                     |
| `gotestsum`  | Supports live watching, `--fast`, and `--focus`                 |
| `entr`       | YAML config, keyboard control, structured output                |
| `ginkgo watch` | No framework dependency, works with standard `go test`       |

## License

[MIT](./LICENSE)
