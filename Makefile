GOFMT_FILES?=$$(find . -name '*.go')

.PHONY: install
install:
	go build -ldflags "-s -w -X github.com/ucloud/ucloud-cli/base.Version=$(shell git describe --tags --always --dirty)" -o out/ucloud .
	cp out/ucloud /usr/local/bin

.PHONY: fmt
fmt:
	gofmt -w -s $(GOFMT_FILES)

.PHONY: release-snapshot
release-snapshot:
	goreleaser release --snapshot --clean
