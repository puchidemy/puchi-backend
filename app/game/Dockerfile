FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/game/go.mod app/game/go.sum ./app/game/
RUN cd app/game && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/game/ ./app/game/

# Build
RUN cd app/game && go build -o /bin/game ./cmd/game/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/game /app/game
WORKDIR /app
EXPOSE 8000 9000
CMD ["./game", "-conf", "/data/conf"]
