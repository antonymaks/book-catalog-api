FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /book-catalog \
    ./cmd/server

FROM alpine:3.23

WORKDIR /app

COPY --from=builder /book-catalog /app/book-catalog
COPY --from=builder /app/web /app/web

EXPOSE 8080

CMD ["/app/book-catalog"]