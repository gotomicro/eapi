.PHONY: build

build-js:
	@cd internal/jsvm/lib && pnpm run build

build:
	@go build -o ./bin/eapi cmd/eapi/eapi.go
