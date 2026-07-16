FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/user/go.mod app/user/go.sum ./app/user/
RUN cd app/user && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/user/ ./app/user/

# Build
RUN cd app/user && go build -o /bin/user ./cmd/user/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/user /app/user
WORKDIR /app
EXPOSE 8000 9000
CMD ["./user", "-conf", "/data/conf"]
