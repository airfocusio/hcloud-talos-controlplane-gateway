.PHONY: *

test:
	go test -v ./...

build:
	goreleaser release --clean --skip=publish --snapshot

release:
	goreleaser release --clean
