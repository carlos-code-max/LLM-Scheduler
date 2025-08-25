vendor:
	go mod tidy
	go mod vendor

build:
	go build -mod=vendor -v main.go

run:
	go run main.go

clean:
	find ./ -name '*.log' -delete

lint: clean
	golangci-lint run --modules-download-mode=vendor ./...

test: clean
	go test -v ./...
