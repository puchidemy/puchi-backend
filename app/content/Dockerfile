FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/content/go.mod app/content/go.sum ./app/content/
RUN cd app/content && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/content/ ./app/content/

# Build
RUN cd app/content && go build -o /bin/content ./cmd/content/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/content /app/content
WORKDIR /app
EXPOSE 8000 9000
CMD ["./content", "-conf", "/data/conf"]
