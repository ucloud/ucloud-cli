export VERSION=0.1.14

.PHONY : build
build: 
	go install && mv ${GOPATH}/bin/ucloud-cli /usr/local/bin/ucloud

.PHONY : build_mac
build_mac:
	GOOS=darwin GOARCH=amd64 go build -o out/ucloud main.go
	tar zcvf out/ucloud-cli-macosx-${VERSION}-amd64.tgz -C out ucloud
	shasum -a 256 out/ucloud-cli-macosx-${VERSION}-amd64.tgz

.PHONY : build_linux
build_linux:
	GOOS=linux GOARCH=amd64 go build -o out/ucloud main.go
	tar zcvf out/ucloud-cli-linux-${VERSION}-amd64.tgz -C out ucloud
	shasum -a 256 out/ucloud-cli-linux-${VERSION}-amd64.tgz

.PHONY : build_windows
build_windows:
	GOOS=windows GOARCH=amd64 go build -o out/ucloud.exe main.go
	zip -r out/ucloud-cli-windows-${VERSION}-amd64.zip out/ucloud.exe
	shasum -a 256 out/ucloud-cli-windows-${VERSION}-amd64.zip

.PHONY : build_all
build_all: build_mac build_linux build_windows

.PHONY : install
install:
	go build -o out/ucloud main.go
	cp out/ucloud /usr/local/bin
