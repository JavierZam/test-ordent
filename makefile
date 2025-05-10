.PHONY: build run test swagger clean docker docker-compose unittest e2e-test

# Build the application
build:
	go build -o app ./cmd/server/main.go

# Run the application
run: build
	./app

# Generate Swagger docs
swagger:
	swag init -g ./cmd/server/main.go

# Run unit tests
unittest:
	go test -v ./tests/unit/...

# Run E2E tests (requires the app to be running)
e2e-test:
	chmod +x ./tests/e2e/api_test.sh
	./tests/e2e/api_test.sh

# Build Docker image
docker:
	docker build -t ecommerce-api .

# Run with Docker Compose
docker-compose:
	docker-compose up --build

# Clean up
clean:
	rm -f app
	rm -rf ./uploads/*