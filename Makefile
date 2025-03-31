.PHONY: build run test clean seed seed-operations seed-territory

# Build the application
build:
	go build -o ./be/bin/mwce-be ./be/cmd/server

# Run the application
run:
	go run ./be/cmd/server/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf ./be/bin

# Seed data
seed:
	go run ./be/cmd/seed/main.go