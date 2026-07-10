GOFMT_FILES?=$$(find . -name '*.go')
LDFLAGS=-s -w -X github.com/ucloud/ucloud-cli/cmd/internal/version.Version=$(shell git describe --tags --always --dirty)

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o out/ucloud .

.PHONY: install
install: build
	cp out/ucloud /usr/local/bin

.PHONY: fmt
fmt:
	gofmt -w -s $(GOFMT_FILES)

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean
