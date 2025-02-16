.PHONY: run down go test test_k6 cover check

run:
	docker compose up

test:
	@go test -coverprofile=coverage.out ./...

cover: test
	@go tool cover -func=coverage.out
