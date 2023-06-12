.PHONY: build

build_dir = bin
name = halftone
os = $(shell go env GOOS)
arch = $(shell go env GOARCH)

build:
	GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=0 go build -o $(build_dir)/$(name) .