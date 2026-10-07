BIN := bin/emulator

.PHONY: build run test lint clean

build:
	go build -o $(BIN) ./src

run: build
	./$(BIN) $(ARGS)

test:
	go test -v ./tests/...

lint:
	gofmt -l src tests
	go vet ./...

clean:
	rm -rf bin
