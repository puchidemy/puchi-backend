FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/grading/go.mod app/grading/go.sum ./app/grading/
RUN cd app/grading && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/grading/ ./app/grading/

# Build
RUN cd app/grading && go build -o /bin/grading ./cmd/grading/

FROM debian:stable-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates netbase \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/grading /app/grading
WORKDIR /app
EXPOSE 8000 9000
CMD ["./grading", "-conf", "/data/conf"]
