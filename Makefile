APP_NAME = gotcha
VERSION ?= dev
BUILD_DIR = bin

.PHONY: all build install uninstall clean test fmt

all: build

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) .

install:
	go install -ldflags "-X main.version=$(VERSION)"

uninstall:
	@bin_dir=$$(go env GOBIN); \
	if [ -z "$$bin_dir" ]; then \
		bin_dir=$$(go env GOPATH)/bin; \
	fi; \
	rm -f $$bin_dir/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

fmt:
	gofmt -w -l .
