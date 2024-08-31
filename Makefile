APP_BINARY=x-proxy

up_app: build_app
	docker-compose --env-file .env  up -d

## up: stops docker-compose (if running), builds all projects and starts docker compose
up: build_app
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose --env-file .env  up --build -d
	@echo "Docker images built and started!"

## build_app: builds the app binary as a linux executable
build_app:
	@echo "Building broker binary..."
	env GOOS=linux CGO_ENABLED=0 go build -o ${APP_BINARY} ./cmd/
	@echo "Done!"

## clear: remove the app binary
clear:
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Removing binary..."
	rm ${APP_BINARY}
	@echo "Done!"