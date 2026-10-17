BINARY := cvyml
PDF    := cv.pdf
YAML   := cv.yml

.PHONY: all build render run test fmt vet check clean

all: build

build:
	go build -o bin/$(BINARY) ./cmd/cvyml

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
