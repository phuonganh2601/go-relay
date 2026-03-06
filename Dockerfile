FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o relay .

FROM alpine:3.19

COPY --from=builder /app/relay /relay

EXPOSE 8080

CMD ["/relay"]