.PHONY: build build-all tidy test clean

# Build specific service
build-%:
	cd app/$* && go build -o ../../bin/$* ./cmd/*/

# Build all services
build-all:
	cd app/core && go build -o ../../bin/core ./cmd/*/
	cd app/content && go build -o ../../bin/content ./cmd/*/
	cd app/grading && go build -o ../../bin/grading ./cmd/*/
	cd app/user && go build -o ../../bin/user ./cmd/*/
	cd app/game && go build -o ../../bin/game ./cmd/*/
	cd app/media && go build -o ../../bin/media ./cmd/*/
	cd app/notification && go build -o ../../bin/notification ./cmd/*/

tidy:
	cd app/core && go mod tidy
	cd app/content && go mod tidy
	cd app/grading && go mod tidy
	cd app/user && go mod tidy
	cd app/game && go mod tidy
	cd app/media && go mod tidy
	cd app/notification && go mod tidy

test:
	cd app/core && go test ./...
	cd app/content && go test ./...
	cd app/grading && go test ./...
	cd app/user && go test ./...
	cd app/game && go test ./...
	cd app/media && go test ./...
	cd app/notification && go test ./...

clean:
	rm -rf bin/
