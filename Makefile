example: # run example
	@go run cmd/example/main.go

test: # run test
	@go run gotest.tools/gotestsum@latest

bump: # bump dependencies
	@go get -u ./...
	@go mod tidy
