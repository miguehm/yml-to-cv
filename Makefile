BINARY := cvyml
PKG    := github.com/miguehm/yml-to-cv/cmd/cvyml
WASM   := cvyml.wasm
PDF    := cv.pdf
YAML   := cv.yml

.PHONY: all build install wasm render run test fmt vet check clean

all: build

build:
	go build -o bin/$(BINARY) ./cmd/cvyml

install:
	go install $(PKG)

wasm:
	GOOS=js GOARCH=wasm go build -o bin/$(WASM) ./cmd/cvyml-wasm

render:
	go run ./cmd/cvyml -in $(YAML) -out $(PDF)

run: build
	./bin/$(BINARY) -in $(YAML) -out $(PDF)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test

clean:
	rm -rf bin $(PDF)
