build: godeps golint
	go build -o ./bin/spry

godeps:
	go get

golint:
	golangci-lint run --skip-dirs api,space

test: build
	go test -v ./tests

test-pg: build
	go test -v ./postgres

compose:
	cd ./.docker && \
		docker-compose up

decomp:
	cd ./.docker && \
		docker-compose down