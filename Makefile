.PHONY: build test lint docker-build

build:
	go build -o trashcommand .

test:
	go test ./...

lint:
	go vet ./...

docker-build:
	docker build -t trashcommand:latest .
