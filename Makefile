default: build

SETENV=
ifeq ($(OS),Windows_NT)
	SETENV=set
endif

lefthook:
	@go install github.com/evilmartians/lefthook@latest
	lefthook install

build:
	go build -v ./...

fix:
	go fix -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	go generate ./...

fmt:
	gofmt -s -w -e .

# Run acceptance tests
test:
	go test -cover  -count=1 ./... -v $(TESTARGS) -timeout 10m

.PHONY: default lefthook build fix install lint generate fmt test
