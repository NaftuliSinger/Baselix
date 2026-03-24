APP_NAME=baselix

run:
	go run ./cmd

build:
	go build -o bin/$(APP_NAME) ./cmd

templ:
	templ generate

dev:
	templ generate && go run ./cmd

tidy:
	go mod tidy