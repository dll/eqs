.PHONY: run build clean migrate

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

clean:
	rm -rf bin/

migrate:
	go run cmd/server/main.go

test:
	go test ./...

lint:
	golangci-lint run
