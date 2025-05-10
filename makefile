.PHONY: build run test swagger clean docker docker-compose unittest e2e-test

build:
	go build -o app ./cmd/server/main.go

run: build
	./app

swagger:
	swag init -g ./cmd/server/main.go

unittest:
	go test -v ./tests/unit/...

e2e-test:
	chmod +x ./tests/e2e/api_test.sh
	./tests/e2e/api_test.sh

docker:
	docker build -t ecommerce-api .

docker-compose:
	docker-compose up --build

clean:
	rm -f app
	rm -rf ./uploads/*