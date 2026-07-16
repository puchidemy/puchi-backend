FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/core/go.mod app/core/go.sum ./app/core/
RUN cd app/core && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/core/ ./app/core/

# Build
RUN cd app/core && go build -o /bin/core ./cmd/core/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/core /app/core
WORKDIR /app
EXPOSE 8000 9000
CMD ["./core", "-conf", "/data/conf"]
