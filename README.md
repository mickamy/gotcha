# 🐹 gotcha

**gotcha** is a test automation tool for Go developers. It watches for file changes and automatically runs `go test`,
supporting TDD and fast feedback.

---

## ✨ Features

- ✅ Automatically detects changes in `.go` files and runs `go test`
- ✅ Define target/excluded paths and test flags in `.gotcha.yaml`
- ✅ `gotcha run` for one-shot test execution
- ✅ `gotcha watch` for continuous test running on save
- ✅ Colored output for test success/failure
- ✅ Easy initialization with `gotcha init`

---

## 🔍 Comparison to Other Tools

| Tool           | Language | Purpose                        | How gotcha differs                                                         |
|----------------|----------|--------------------------------|----------------------------------------------------------------------------|
| `spring rspec` | Ruby     | Fast RSpec runs via preload    | gotcha is CLI-based and watch-driven, optimized for Go                     |
| `watchexec`    | Any      | Run any command on file change | gotcha is Go-native, built specifically for `go test`                      |
| `gotestsum`    | Go       | Pretty test output formatting  | gotcha supports live watching and reruns with test summary                 |
| `entr`         | Any      | Minimal file-change trigger    | gotcha provides YAML config, input control (`r`, `q`), and output handling |

Unlike generic file watchers, **gotcha is purpose-built for Go developers** who want fast, automatic test feedback
without extra setup or dependencies.

✅ Lightweight CLI  
✅ Integrated `go test` runner  
✅ `.gotcha.yaml` for fine-grained test control  
✅ Supports `--summary` for clean test result overview  
✅ `watch` mode with keyboard control (`r` to rerun, `q` to quit)

---

## 📦 Install

```sh
# go >= v1.24
go get -tool github.com/mickamy/gotcha@latest
# go < v1.24
go install github.com/mickamy/gotcha@latest
```

or

```sh
git clone https://github.com/mickamy/gotcha.git
cd gotcha
make install
```

---

## 🚀 Usage

### 1. Initialize

```sh
gotcha init
```

This generates `.gotcha.yaml`:

```yaml
include:
  - "./..."
exclude:
  - "vendor/"
  - "mocks/"
args:
  - "-v"
```

### 2. Run once

```sh
gotcha run
```

### 3. Watch for changes

```sh
gotcha watch
```

This reruns tests automatically whenever `.go` files change.

---

## ⚙️ Configuration (`.gotcha.yaml`)

| Field     | Description                                        |
|-----------|----------------------------------------------------|
| `include` | Target packages (passed to `go list`)              |
| `exclude` | Directories or paths to ignore                     |
| `args`    | Arguments passed to `go test` (e.g. `-v`, `-race`) |

## 💡 Why gotcha?

Compared to similar tools like `gotestsum`, `richgo`, `ginkgo watch`, and `entr`, gotcha provides:

- 🧠 **Zero-config default behavior**: just run `gotcha watch` in any Go project
- 🧹 **YAML-based filtering**: simple includes/excludes for packages and paths
- 📦 **Lightweight**: single-purpose CLI with no dependencies on external runners or frameworks
- 🎯 **TDD-first design**: optimized for fast, repeated test execution during development
- 🌈 **Clean, colored terminal UX**: highlights pass/fail status clearlyz

## 🛣 Roadmap

- [ ] `--fast` flag to test only changed packages
- [ ] `.gotcha.yaml` watch-specific settings (e.g. debounce)

---

## 📄 License

[MIT](./LICENSE)
