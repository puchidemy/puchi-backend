FROM golang:1.26 AS builder
WORKDIR /src

COPY app/auth/go.mod app/auth/go.sum ./app/auth/
RUN cd app/auth && go mod download

COPY app/auth/ ./app/auth/
RUN cd app/auth && CGO_ENABLED=0 go build -o /bin/auth ./cmd/auth/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/auth /app/auth
WORKDIR /app
EXPOSE 8000
CMD ["./auth", "-conf", "/data/conf"]
