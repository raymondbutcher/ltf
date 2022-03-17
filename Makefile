sources = go.sum $(shell find -name '*.go')

.PHONY: build
build: test ltf

.PHONY: clean
clean:
	rm -rf ltf

.PHONY: test
test: $(sources)
	go test ./...

.PHONY: test
testv: $(sources)
	go test -v ./...

ltf: $(sources)
	go build ./cmd/ltf
