.PHONY: run test cover

run:
	docker compose up

test:
	@go test -coverprofile=coverage.out ./...

cover: test
	@go tool cover -func=coverage.out
