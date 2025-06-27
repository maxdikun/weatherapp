gen:
	@go generate ./...

test: gen
	@go test -v ./... -race -coverprofile=coverage.out -covermode=atomic 

run: gen
	go run .