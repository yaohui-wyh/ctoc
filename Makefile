.PHONY: test build

build:
	mkdir -p bin
	GO111MODULE=on go build -o ./bin/gocloc cmd/gocloc/main.go

update-package:
	GO111MODULE=on go get -u github.com/hhatto/gocloc

cleanup-package:
	GO111MODULE=on go mod tidy

run-example:
	GO111MODULE=on go run examples/languages/main.go
	GO111MODULE=on go run examples/files/main.go

test:
	GO111MODULE=on go test -v

test-cover:
	GO111MODULE=on go test -v -coverprofile=coverage.out
