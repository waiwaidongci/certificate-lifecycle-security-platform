.PHONY: fmt test vet build run-dev count clean

fmt:
	gofmt -w .

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./...

run-dev:
	./scripts/run-dev.sh

count:
	./scripts/count-go.sh

clean:
	rm -rf bin certpilot.db certpilot.db-shm certpilot.db-wal
