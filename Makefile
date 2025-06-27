gen:
	@go generate ./...

test: gen
	@go test -v ./... -race -coverprofile=coverage.out -covermode=atomic 

lint:
	@golangci-lint run

run: gen
	go run .

up:
	@docker-compose up --detach

down:
	@docker-compose down