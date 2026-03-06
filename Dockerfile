FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY main.go .

RUN go build -ldflags="-s -w" -o relay

FROM alpine:3.19

COPY --from=builder /app/relay /relay

EXPOSE 8080

CMD ["/relay"]