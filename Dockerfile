FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

WORKDIR /app/cmd/api
RUN go build -o api .

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/cmd/api/api .

EXPOSE 8080
CMD ["./api"]

