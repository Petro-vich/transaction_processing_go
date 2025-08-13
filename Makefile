.PHONY: docker-build docker-run test test_cover clean

run-local:
	CONFIG_PATH=config/local.yaml go run cmd/main.go

run-prod:
	CONFIG_PATH=config/docker.yaml go run main.go

test:
	go clean -testcache
	go test ./...

test_cover:
	go clean -testcache
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html

docker-build:
	docker build -t transaction-service .

docker-run:
	docker run -p 8080:8080 \
		-e CONFIG_PATH=config/docker.yaml \
		-v $(PWD)/storage/sqlite:/app/storage/sqlite \
		transaction-service

clean:
	rm -f *.out *.html