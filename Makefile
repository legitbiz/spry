build: godeps golint
	go build -o ./bin/spacedora

godeps:
	go get

golint:
	golangci-lint run --skip-dirs api,space

test: build
	go test -v .