dev: clean
	mkdir -p ./bin
	cp -r ./config/sample.yaml ./bin/config.yaml
	go build -o ./bin/runner cmd/runner/*.go

build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o runner cmd/runner/*.go

clean:
	@if [ -d "./bin" ]; then \
		rm -rf ./bin; \
	fi

run:
	go run cmd/runner/*.go
