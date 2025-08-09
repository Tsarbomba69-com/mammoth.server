TEST_DIR ?= ./tests
BINARY_NAME ?= mammoth.exe
IMAGE_NAME ?= mammoth-server:latest
# Command to set up and running the docker compose
up:
	@docker compose -f docker-compose.dev.yml --env-file .env.dev up -d --build

down:
	@docker compose down

docker-build:
	@docker build -t $(IMAGE_NAME) .

build:
	@go build -o .build/$(BINARY_NAME)

run: build
	@echo "Running the project..."
	@./.build/$(BINARY_NAME)

# Command to run the project.
test:
	@go test $(TEST_DIR) -v -cover

swag:
	@swag init -g main.go -o ./docs

lint:
	@echo "Running go fmt..."
	@go fmt ./...
	@echo "Running go vet..."
	@go vet ./...
	@echo "Running golangci-lint..."
	@golangci-lint run

mod-tidy:
	@echo "Running go mod tidy..."
	@go mod tidy