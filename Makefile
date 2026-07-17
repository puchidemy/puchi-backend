.PHONY: build build-all tidy test clean dev-up dev-down

# Local: run all 5 services with unique ports (see scripts/dev/)
dev-up:
	@powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev/run-all.ps1

dev-down:
	@powershell -NoProfile -ExecutionPolicy Bypass -File scripts/dev/stop-all.ps1 -KillPorts

# Build specific service
build-%:
	cd app/$* && go build -o ../../bin/$* ./cmd/*/

# Build all services
build-all:
	cd app/auth && go build -o ../../bin/auth ./cmd/*/
	cd app/core && go build -o ../../bin/core ./cmd/*/
	cd app/learn && go build -o ../../bin/learn ./cmd/*/
	cd app/media && go build -o ../../bin/media ./cmd/*/
	cd app/notification && go build -o ../../bin/notification ./cmd/*/

tidy:
	cd app/auth && go mod tidy
	cd app/core && go mod tidy
	cd app/learn && go mod tidy
	cd app/media && go mod tidy
	cd app/notification && go mod tidy

test:
	cd app/auth && go test ./...
	cd app/core && go test ./...
	cd app/learn && go test ./...
	cd app/media && go test ./...
	cd app/notification && go test ./...

clean:
	rm -rf bin/
