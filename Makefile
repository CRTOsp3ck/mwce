.PHONY: build run test clean seed seed-operations seed-territory

# Build the application
build:
	go build -o ./bin/mwce-be ./cmd/server

# Run the application
run:
	go run ./cmd/server/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf ./bin

# Seed territory data
seed-territory:
	go run ./cmd/seed/main.go

# Seed operations data
seed-operations:
	go run ./cmd/seed/operations/main.go

# Run both seed commands
seed: seed-territory seed-operations