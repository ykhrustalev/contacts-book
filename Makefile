DEFAULT: build

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build
build:
	go build -o bin/contacts cmd/app/main.go

.PHONY: test*
test:
	rm -f coverage.out
	go test -p 1 -race -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
