default: build

build:
	go build -o bin/sbanalyzer cmd/main.go

test:
	go test $$(go list ./...) -cover
