FROM golang:1.26 AS builder
COPY . /src
WORKDIR /src
RUN make tidy && make build

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/bin /app
WORKDIR /app
EXPOSE 8000 9000
CMD ["./game", "-conf", "/data/conf"]
