.PHONY: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -mod=vendor -o bin/stecagrid-exporter src/stecagrid-exporter.go 
