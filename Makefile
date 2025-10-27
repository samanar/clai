.PHONY: build install clean test db-reset db-show

# Build binary with FTS5 support
build:
	CGO_ENABLED=1 go build -tags sqlite_fts5 -o clai

# Install to /usr/local/bin
install: build
	install -m 755 clai /usr/local/bin/clai

# Clean build artifacts
clean:
	rm -f clai

# Run tests
test:
	CGO_ENABLED=1 go test -tags sqlite_fts5 ./...

# Reset the man pages database
db-reset: build
	./clai db reset

# Show database information
db-show: build
	./clai db show

# Build for release (includes test subcommand removal in production)
release:
	CGO_ENABLED=1 go build -tags sqlite_fts5 -ldflags="-s -w" -o clai
