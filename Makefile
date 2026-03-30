APP_NAME=baselix

dev:
	$(MAKE) templ
	$(MAKE) run

run:
	go run ./cmd

build:
	go build -o bin/$(APP_NAME) ./cmd

templ:
	templ generate

tidy:
	go mod tidy

test:
	go test ./...