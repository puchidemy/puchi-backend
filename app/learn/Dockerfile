FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/learn/go.mod app/learn/go.sum ./app/learn/
RUN cd app/learn && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/learn/ ./app/learn/

# Build
RUN cd app/learn && go build -o /bin/learn ./cmd/learn/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/learn /app/learn
WORKDIR /app
EXPOSE 8000 9000
CMD ["./learn", "-conf", "/data/conf"]
