.PHONY: test test-verbose test-coverage lint format vet check

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

format:
	go fmt ./...
	go tool golangci-lint run --fix

lint:
	make format
	go tool golangci-lint run

vet:
	go vet ./...

# lint・vet・testをまとめて実行する（CI相当のチェック）
check:
	make lint
	make vet
	make test
