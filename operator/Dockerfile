# Build stage
FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

# Build the Go binary for Linux explicitly
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o operator .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/operator .

CMD ["./operator"]
