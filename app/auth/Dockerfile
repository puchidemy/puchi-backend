FROM golang:1.26 AS builder
WORKDIR /src

# Copy service dependency files
COPY app/auth/go.mod app/auth/go.sum ./app/auth/
RUN cd app/auth && go mod download

# Copy source
COPY app/auth/ ./app/auth/

# Build
RUN cd app/auth && go build -o /bin/auth ./cmd/auth/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/auth /app/auth
WORKDIR /app
EXPOSE 8000 9000
CMD ["./auth", "-conf", "/data/conf"]
