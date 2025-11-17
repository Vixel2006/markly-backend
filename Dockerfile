# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy modules first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

WORKDIR /app/cmd/api
RUN go build -o api .

# Run stage
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/cmd/api/api .

EXPOSE 8080
CMD ["./backend"]

