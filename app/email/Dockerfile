FROM golang:1.26 AS builder
WORKDIR /src

# Copy pkg module (needed for replace directive)
COPY pkg/go.mod pkg/go.sum ./pkg/

# Copy service dependency files
COPY app/email/go.mod app/email/go.sum ./app/email/
RUN cd app/email && go mod download

# Copy source
COPY pkg/ ./pkg/
COPY app/email/ ./app/email/

# Build
RUN cd app/email && go build -o /bin/email ./cmd/email/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/email /usr/local/bin/email
CMD ["email"]
