BINARY = inventory-mcp
MODULE = github.com/venkytv/inventory-mgmt

.PHONY: build test clean linux

build:
	go build -o $(BINARY) .

test:
	go test ./... -v

clean:
	rm -f $(BINARY) $(BINARY)-*

linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux .
