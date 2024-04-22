build:
	go build -o booking-service cmd/app/*.go

run:
	go run cmd/app/*.go

test:
	go test ./...

all: test build 