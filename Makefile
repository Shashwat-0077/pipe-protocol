.PHONY: build run test clean client

# Build the server
build:
	go build -o bin/pipe-server cmd/pipe-server/main.go
	go build -o bin/pipe-client cmd/pipe-client/main.go

# Run the server
run: build
	mkdir -p data
	./bin/pipe-server -config config.yaml

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf data/

# Send a test message
client: build
	./bin/pipe-client \
		-host localhost \
		-port 2525 \
		-from "sender@remote.com" \
		-to "user@pipe.local" \
		-subject "Test Message" \
		-body "Hello from PIPE protocol!"

# Setup development environment
setup:
	go mod tidy
	mkdir -p bin data

# Run multiple test messages
test-messages: build
	./bin/pipe-client -from "alice@remote.com" -to "bob@pipe.local" -subject "Hello Bob" -body "This is a test message from Alice"
	./bin/pipe-client -from "charlie@external.com" -to "bob@pipe.local,alice@pipe.local" -subject "Group Message" -body "Hello everyone!"
	./bin/pipe-client -from "system@pipe.local" -to "admin@pipe.local" -subject "System Alert" -body "System is running normally"

# View stored messages
view-inbox:
	@echo "=== INBOX MESSAGES ==="
	@find data/inbox -name "*.json" -exec echo "File: {}" \; -exec cat {} \; -exec echo "" \; 2>/dev/null || echo "No messages in inbox"

view-outbox:
	@echo "=== OUTBOX MESSAGES ==="
	@find data/outbox -name "*.json" -exec echo "File: {}" \; -exec cat {} \; -exec echo "" \; 2>/dev/null || echo "No messages in outbox"

view-sent:
	@echo "=== SENT MESSAGES ==="
	@find data/sent -name "*.json" -exec echo "File: {}" \; -exec cat {} \; -exec echo "" \; 2>/dev/null || echo "No messages in sent folder"

# Development commands
dev-server: build
	@echo "Starting development server..."
	@echo "Server will listen on localhost:2525"
	@echo "Data will be stored in ./data/"
	@echo "Press Ctrl+C to stop"
	./bin/pipe-server -config config.yaml

# Interactive client session
interactive-client:
	@echo "=== PIPE Client Interactive Mode ==="
	@echo "Enter message details:"
	@read -p "From address: " from; \
	read -p "To address(es) [comma separated]: " to; \
	read -p "Subject: " subject; \
	read -p "Message body: " body; \
	./bin/pipe-client -from "$from" -to "$to" -subject "$subject" -body "$body"
