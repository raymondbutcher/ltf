.PHONY: build
build: ltf

.PHONY: clean
clean:
	rm -rf ltf

.PHONY: test
test:
	go test ./...

.PHONY: test
testv:
	go test -v ./...

ltf: *.go go.sum
	go build
