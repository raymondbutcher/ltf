sources = go.sum $(shell find -name '*.go')

.PHONY: build
build: test ltf

.PHONY: clean
clean:
	rm -rf ltf

.PHONY: test
test: $(sources)
	go test ./...
	# success

.PHONY: test
testv: $(sources)
	go test -v ./...
	# success

ltf: $(sources)
	go build ./cmd/ltf
