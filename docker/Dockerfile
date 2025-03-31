FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/config /app/config

EXPOSE 50051

CMD ["/app/api"] 