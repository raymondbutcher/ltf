.PHONY: build
build: test ltf

.PHONY: clean
clean:
	rm -rf ltf

.PHONY: test
test: *.go go.sum
	go test ./...

.PHONY: test
testv: *.go go.sum
	go test -v ./...

ltf: *.go go.sum
	go build
