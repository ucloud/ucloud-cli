export VERSION=0.1.40
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

.PHONY : install
install:
	go build -i -v -mod=vendor -o out/ucloud main.go
	cp out/ucloud /usr/local/bin

.PHONY : build_mac
build_mac:
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -o out/ucloud main.go
	tar zcvf out/ucloud-cli-macosx-${VERSION}-amd64.tgz -C out ucloud
	shasum -a 256 out/ucloud-cli-macosx-${VERSION}-amd64.tgz

build_mac_arm:
	GOOS=darwin GOARCH=arm64 go build -mod=vendor -o out/ucloud main.go
	tar zcvf out/ucloud-cli-macosx-${VERSION}-arm64.tgz -C out ucloud
	shasum -a 256 out/ucloud-cli-macosx-${VERSION}-arm64.tgz

.PHONY : build_linux
build_linux:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o out/ucloud main.go
	tar zcvf out/ucloud-cli-linux-${VERSION}-amd64.tgz -C out ucloud
	shasum -a 256 out/ucloud-cli-linux-${VERSION}-amd64.tgz

.PHONY : build_windows
build_windows:
	GOOS=windows GOARCH=amd64 go build -mod=vendor -o out/ucloud.exe main.go
	zip -r out/ucloud-cli-windows-${VERSION}-amd64.zip out/ucloud.exe
	shasum -a 256 out/ucloud-cli-windows-${VERSION}-amd64.zip

.PHONY : build_all
build_all: build_mac build_linux build_windows

.PHONY: fmt
fmt:
	gofmt -w -s $(GOFMT_FILES)
