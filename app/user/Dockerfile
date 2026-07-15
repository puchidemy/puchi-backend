FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy && go build -o /bin/user ./cmd/user/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/user /app/user
WORKDIR /app
EXPOSE 8000 9000
CMD ["./user", "-conf", "/data/conf"]
