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

## 📦 Install

```sh
go get -tool github.com/mickamy/gotcha@latest
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

## 🛣 Roadmap

- [ ] `--fast` flag to test only changed packages
- [ ] `.gotcha.yaml` watch-specific settings (e.g. debounce)

---

## 📄 License

[MIT](./LICENSE)
