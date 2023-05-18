.PHONY: setup
setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

.PHONY: fmt
fmt:
	goimports -l -w ./

.PHONY: check
check:
	if ! command -v golangci-lint; then \
		echo 'missing golangci-lint; run \\'make setup\\''; \
		exit 1; \
	fi
	golangci-lint --version
	golangci-lint --verbose run

.PHONY: build
build:
	go build -C ./cmd/turret/ -ldflags '-s -w' -o ./build/
