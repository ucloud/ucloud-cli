#!/bin/bash

set -ue

targets=( \
	"darwin_amd64" \
	"darwin_arm64" \
	"linux_amd64" \
	"linux_arm64" \
	"windows_amd64" \
	"windows_arm64" \
)

VERSION=${GITHUB_REF#refs/*/}
echo "VERSION=${VERSION}" >> $GITHUB_ENV

mkdir -p out

for target in "${targets[@]}"; do
	echo "Build target: ${target}"
	IFS='_' read -r -a tmp <<< "$target"
	BUILD_OS="${tmp[0]}"
	BUILD_ARCH="${tmp[1]}"
	GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH}" go build -mod=vendor -o bin/ucloud
	zip -r out/ucloud-${target}.zip ./bin LICENSE
done
