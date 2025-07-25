.PHONY: build
build:
	go mod vendor
	go build -mod=vendor -o bin/marketplace ./cmd/marketplace

.PHONY: test
test:
	go test -tags=handlers_test -v -coverprofile=cover.out -coverpkg=./... ./internal/handlers
	go tool cover -html=cover.out -o cover.html

.PHONY: lint
lint:
	go mod vendor
	golangci-lint run -c .golangci.yml -v --modules-download-mode=vendor ./...

.PHONY: clean
clean:
	rm -rf bin/* vendor/*