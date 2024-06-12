export VERSION=0.2.0
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

.PHONY : install
install:
	go build -v -mod=vendor -o out/ucloud main.go
	cp out/ucloud /usr/local/bin

.PHONY : build-darwin-amd64
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -o out/darwin_amd64/ucloud main.go
	@cp LICENSE out/darwin_amd64

.PHONY : build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -mod=vendor -o out/darwin_arm64/ucloud main.go
	@cp LICENSE out/darwin_arm64

.PHONY : build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o out/linux_amd64/ucloud main.go
	@cp LICENSE out/linux_amd64

.PHONY : build-linux-arm64
build-linux-arm64:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o out/linux_arm64/ucloud main.go
	@cp LICENSE out/linux_arm64

.PHONY : build-windows-amd64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -mod=vendor -o out/windows_amd64/ucloud.exe main.go
	@cp LICENSE out/windows_amd64

.PHONY : build-all
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-arm64 build-linux-amd64 build-windows-amd64

.PHONY: fmt
fmt:
	gofmt -w -s $(GOFMT_FILES)
