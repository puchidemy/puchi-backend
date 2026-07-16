FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/notification/go.mod app/notification/go.sum ./app/notification/
RUN cd app/notification && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/notification/ ./app/notification/

# Build
RUN cd app/notification && go build -o /bin/notification ./cmd/notification/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/notification /app/notification
WORKDIR /app
EXPOSE 8000 9000
CMD ["./notification", "-conf", "/data/conf"]
