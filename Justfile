PROJECT := "vendorito"

test:
    go test ./...

build:
    go build -o bin/{{ PROJECT }} ./cmd/vendorito

run *args: build
    bin/vendorito {{ args }}