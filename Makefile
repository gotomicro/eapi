.PHONY: build

build:
	@go build -o ./bin/egogen cmd/egogen/egogen.go
