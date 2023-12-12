BINARY_NAME=hispeed2App

build:
	@go mod vendor
	@echo "Building HiSpeed2..."
	@go build -o tmp/${BINARY_NAME} .
	@echo "HiSpeed2 built!"

run: build
	@echo "Starting HiSpeed2..."
	@./tmp/${BINARY_NAME} &
	@echo "HiSpeed2 started!"

clean:
	@echo "Cleaning..."
	@go clean
	@rm tmp/${BINARY_NAME}
	@echo "Cleaned!"

test:
	@echo "Testing..."
	@go test ./...
	@echo "Done!"

start: run

stop:
	@echo "Stopping HiSpeed2..."
	@-pkill -SIGTERM -f "./tmp/${BINARY_NAME}"
	@echo "Stopped HiSpeed2!"

restart: stop start