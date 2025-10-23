.PHONY: build new-database clean help

BINARY := dist/local/ottomat
DB_PATH := testdata/ottomat.db
ADMIN_PASSWORD := happy.cat.happy.nap

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the ottomat binary
	@echo "Building ottomat..."
	@go build -o $(BINARY) ./cmd/ottomat
	@echo "Build complete: $(BINARY)"

new-database: build ## Initialize a new test database with default admin
	@./tools/init-new-database.sh $(DB_PATH) $(ADMIN_PASSWORD)

clean: ## Remove build artifacts and test database
	@echo "Cleaning up..."
	@rm -f $(BINARY)
	@rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal
	@echo "Clean complete"
