.PHONY: build run-be run-be-seed run-fe test clean

# Run the back-end
run-be:
	cd be/cmd/server && go run .

# Run the back-end seeder
run-be-seed:
	cd be/cmd/seed && go run .

# Run the front-end
run-fe:
	cd fe && npm run dev

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf ./be/bin